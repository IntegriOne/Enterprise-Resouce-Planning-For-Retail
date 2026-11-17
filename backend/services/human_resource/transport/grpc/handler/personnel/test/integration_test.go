//go:build integration

package test

import (
	"context"
	"net"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/personnel/v1"
	hrdb "github.com/integrione/erp-retail/backend/services/human_resource/db"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	"github.com/integrione/erp-retail/backend/services/human_resource/infrastructure/security"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/personnel"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository"
	handler "github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/personnel"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type workerScopePolicy struct {
	legalEntityID string
	principalID   *string
}

type workerFixedClock struct{ now time.Time }

func (c workerFixedClock) Now() time.Time { return c.now }

type workerReferenceVerifier struct{}

func (workerReferenceVerifier) VerifyWorkerReferences(_ context.Context, check security.WorkerReferenceCheck) error {
	if check.PersonPartyID == "" || check.LegalEntityID == "" {
		return hrerrors.ErrReferenceConflict
	}
	return nil
}

func (p workerScopePolicy) Authorize(context.Context, operation.Access) (operation.Actor, error) {
	principalID := "00000000-0000-4000-8000-000000000001"
	if p.principalID != nil {
		principalID = *p.principalID
	}
	return operation.Actor{
		PrincipalID: principalID,
		SessionID:   "00000000-0000-4000-8000-000000000002",
		Kind:        "human",
	}, nil
}

func (p workerScopePolicy) ReadScope(_ context.Context, access operation.Access) (operation.Scope, error) {
	if access.Resource != "hr_workers" {
		return operation.Scope{All: true}, nil
	}
	return operation.Scope{Filters: []domainrepo.Filter{{Column: "primary_legal_entity_reference_id", Values: []any{p.legalEntityID}}}}, nil
}

func TestWorkerScopedQueryRPCWithPostgreSQL(t *testing.T) {
	dsn := os.Getenv("HRM_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("HRM_TEST_DATABASE_URL is not configured")
	}
	parsed, err := url.Parse(dsn)
	if err != nil || parsed.Scheme != "postgres" || !strings.HasSuffix(strings.TrimPrefix(parsed.Path, "/"), "_test") {
		t.Fatal("integration test requires a dedicated *_test database")
	}
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal("open HRM test database failed")
	}
	pool, err := database.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := pool.Close(); err != nil {
			t.Errorf("close HRM test database pool: %v", err)
		}
	})
	if err := hrdb.RunMigrationsUp(ctx, pool); err != nil {
		t.Fatal(err)
	}
	repos, err := repository.New(database)
	if err != nil {
		t.Fatal(err)
	}
	validators, err := security.PersonnelValidators(repos)
	if err != nil {
		t.Fatal(err)
	}
	workerValidators, err := security.WorkerValidators(repos, workerReferenceVerifier{})
	if err != nil {
		t.Fatal(err)
	}
	for resource, validator := range workerValidators {
		validators[resource] = validator
	}
	ids := operation.UUIDGenerator{}
	insideID, err := ids.NewID()
	if err != nil {
		t.Fatal(err)
	}
	outsideID, err := ids.NewID()
	if err != nil {
		t.Fatal(err)
	}
	legalA, err := ids.NewID()
	if err != nil {
		t.Fatal(err)
	}
	legalB, err := ids.NewID()
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range []struct{ id, legal string }{{insideID, legalA}, {outsideID, legalB}} {
		if err := database.WithContext(ctx).Exec(
			"INSERT INTO hr_workers (id, worker_number, person_party_reference_id, worker_category, status, primary_legal_entity_reference_id) VALUES (?, ?, ?, ?, ?, ?)",
			row.id, "SCOPE-"+row.id, row.id, "employee", "active", row.legal,
		).Error; err != nil {
			t.Fatal(err)
		}
	}
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		if err := database.WithContext(cleanupCtx).Exec("DELETE FROM hr_workers WHERE id IN (?, ?)", insideID, outsideID).Error; err != nil {
			t.Errorf("clean HRM worker test rows: %v", err)
		}
	})
	principalID := "00000000-0000-4000-8000-000000000001"
	testNow := time.Date(2026, 11, 22, 12, 0, 0, 0, time.UTC)
	svc, err := service.NewService(service.Dependencies{
		Runtime: operation.Dependencies{
			Transactor: repos.Transactor, Outbox: repos.OutboxEvent,
			Policy:    workerScopePolicy{legalEntityID: legalA, principalID: &principalID},
			Integrity: security.NewIntegrity(validators), Clock: workerFixedClock{now: testNow},
		},
		Worker: repos.Worker, EmploymentContract: repos.EmploymentContract,
		WorkerAssignment: repos.WorkerAssignment, WorkerLifecycleEvent: repos.WorkerLifecycleEvent,
		WorkerDocument: repos.WorkerDocument, EmergencyContact: repos.EmergencyContact,
	})
	if err != nil {
		t.Fatal(err)
	}
	listener := bufconn.Listen(1 << 20)
	server := grpc.NewServer()
	if err := handler.Register(server, svc); err != nil {
		t.Fatal(err)
	}
	served := make(chan error, 1)
	go func() { served <- server.Serve(listener) }()
	t.Cleanup(func() {
		server.Stop()
		if err := <-served; err != nil {
			t.Errorf("serve HRM test gRPC: %v", err)
		}
	})
	connection, err := grpc.NewClient("passthrough:///hrm-personnel-scope-integration",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return listener.DialContext(ctx) }),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := connection.Close(); err != nil {
			t.Errorf("close HRM test gRPC connection: %v", err)
		}
	})
	client := pb.NewPersonnelServiceClient(connection)
	if _, err := client.GetWorkerByID(ctx, &common.IDRequest{Id: outsideID}); status.Code(err) != codes.NotFound {
		t.Fatalf("outside-scope worker query = %v, want not found", err)
	}
	inside, err := client.GetWorkerByID(ctx, &common.IDRequest{Id: insideID})
	if err != nil || inside.GetItem().GetId() != insideID {
		t.Fatalf("inside-scope worker query = %v", err)
	}
	page, err := client.GetAllWorkers(ctx, &common.PageRequest{Limit: 1, Offset: 0})
	if err != nil || page.GetPage().GetTotal() != 1 || len(page.GetItems()) != 1 || page.GetItems()[0].GetId() != insideID {
		t.Fatalf("scoped worker page = %+v, %v", page, err)
	}
	empty, err := client.GetAllWorkers(ctx, &common.PageRequest{Limit: 1, Offset: 1})
	if err != nil || empty.GetPage().GetTotal() != 1 || len(empty.GetItems()) != 0 {
		t.Fatalf("scoped worker next page = %+v, %v", empty, err)
	}
	missing, err := client.FindWorkerByPrimaryLegalEntityReferenceID(ctx, &common.StringFinderRequest{Value: &legalB, Limit: 10})
	if err != nil || missing.GetPage().GetTotal() != 0 || len(missing.GetItems()) != 0 {
		t.Fatalf("outside-scope worker finder = %+v, %v", missing, err)
	}
	contractNumber := "DRAFT-" + insideID
	contractInput := &pb.EmploymentContractInput{
		WorkerId: &insideID, ContractNumber: &contractNumber,
		ContractType: stringPointer("fixed_term"), EmployingLegalEntityReferenceId: &legalA,
		EffectiveFrom: stringPointer("2026-11-17"), WorkCountryCode: stringPointer("ID"),
		GoverningJurisdictionCode: stringPointer("ID"),
	}
	signedInput := *contractInput
	signedInput.SignedDocumentReferenceId = stringPointer(insideID)
	if _, err := client.CreateEmploymentContract(ctx, &pb.CreateEmploymentContractRequest{Input: &signedInput}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("signed evidence through draft CRUD = %v", err)
	}
	created, err := client.CreateEmploymentContract(ctx, &pb.CreateEmploymentContractRequest{Input: contractInput})
	if err != nil || created.GetItem().GetContractStatus() != "draft" {
		t.Fatalf("create draft contract = %+v, %v", created, err)
	}
	contractID := created.GetItem().GetId()
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		if err := database.WithContext(cleanupCtx).Exec("DELETE FROM hr_employment_contracts WHERE id = ?", contractID).Error; err != nil {
			t.Errorf("clean HRM contract test row: %v", err)
		}
	})
	if _, err := client.CreateEmploymentContract(ctx, &pb.CreateEmploymentContractRequest{Input: contractInput}); status.Code(err) != codes.AlreadyExists {
		t.Fatalf("duplicate contract number = %v", err)
	}
	updated, err := client.UpdateEmploymentContract(ctx, &pb.UpdateEmploymentContractRequest{
		Id: contractID, Input: &pb.EmploymentContractInput{ContractType: stringPointer("permanent")},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"contract_type"}},
	})
	if err != nil || updated.GetItem().GetContractType() != "permanent" {
		t.Fatalf("update draft contract = %+v, %v", updated, err)
	}
	got, err := client.GetEmploymentContractByID(ctx, &common.IDRequest{Id: contractID})
	if err != nil || got.GetItem().GetContractNumber() != contractNumber {
		t.Fatalf("read draft contract = %+v, %v", got, err)
	}
	pageContracts, err := client.GetAllEmploymentContracts(ctx, &common.PageRequest{Limit: 1})
	if err != nil || pageContracts.GetPage().GetTotal() != 1 || len(pageContracts.GetItems()) != 1 {
		t.Fatalf("contract pagination = %+v, %v", pageContracts, err)
	}
	if _, err := client.DeleteEmploymentContract(ctx, &common.IDRequest{Id: contractID}); err != nil {
		t.Fatalf("delete draft contract: %v", err)
	}
	var events int64
	if err := database.WithContext(ctx).Table("hr_outbox_events").Where("aggregate_id = ?", contractID).Count(&events).Error; err != nil || events != 3 {
		t.Fatalf("contract outbox events = %d, %v", events, err)
	}
	workerNumber := "WORKER-DRAFT-" + insideID
	personPartyID, err := ids.NewID()
	if err != nil {
		t.Fatal(err)
	}
	workerInput := &pb.WorkerInput{
		WorkerNumber: &workerNumber, PersonPartyReferenceId: &personPartyID,
		WorkerCategory: stringPointer("employee"), PrimaryLegalEntityReferenceId: &legalA,
		HireDate: stringPointer("2026-11-20"), ServiceStartDate: stringPointer("2026-11-21"),
	}
	createdWorker, err := client.CreateWorker(ctx, &pb.CreateWorkerRequest{Input: workerInput})
	if err != nil || createdWorker.GetItem().GetStatus() != "preboarding" {
		t.Fatalf("create preboarding worker = %+v, %v", createdWorker, err)
	}
	newWorkerID := createdWorker.GetItem().GetId()
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		if err := database.WithContext(cleanupCtx).Exec("DELETE FROM hr_workers WHERE id = ?", newWorkerID).Error; err != nil {
			t.Errorf("clean preboarding worker: %v", err)
		}
	})
	if _, err := client.CreateWorker(ctx, &pb.CreateWorkerRequest{Input: workerInput}); status.Code(err) != codes.AlreadyExists {
		t.Fatalf("duplicate worker number = %v", err)
	}
	linkedInput := *workerInput
	linkedInput.IamPrincipalReferenceId = &insideID
	if _, err := client.CreateWorker(ctx, &pb.CreateWorkerRequest{Input: &linkedInput}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("IAM link through worker CRUD = %v", err)
	}
	updatedWorker, err := client.UpdateWorker(ctx, &pb.UpdateWorkerRequest{Id: newWorkerID,
		Input:      &pb.WorkerInput{ServiceStartDate: stringPointer("2026-11-22")},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"service_start_date"}}})
	if err != nil || updatedWorker.GetItem().GetServiceStartDate() != "2026-11-22" {
		t.Fatalf("correct preboarding date = %+v, %v", updatedWorker, err)
	}
	assignmentFamilyID, err := ids.NewID()
	if err != nil {
		t.Fatal(err)
	}
	assignmentJobID, err := ids.NewID()
	if err != nil {
		t.Fatal(err)
	}
	assignmentPositionID, err := ids.NewID()
	if err != nil {
		t.Fatal(err)
	}
	assignmentStart := "2026-11-22"
	for _, fixture := range []struct {
		query string
		args  []any
	}{
		{"INSERT INTO hr_job_families (id, family_code, family_name, job_domain) VALUES (?, ?, 'Store', 'operations')", []any{assignmentFamilyID, "FAMILY-" + assignmentFamilyID}},
		{"INSERT INTO hr_jobs (id, job_family_id, job_code, job_title, valid_from) VALUES (?, ?, ?, 'Associate', ?)", []any{assignmentJobID, assignmentFamilyID, "JOB-" + assignmentJobID, assignmentStart}},
		{"INSERT INTO hr_positions (id, position_code, position_title, job_id, organization_unit_reference_id, legal_entity_reference_id, valid_from) VALUES (?, ?, 'Associate', ?, ?, ?, ?)", []any{assignmentPositionID, "POSITION-" + assignmentPositionID, assignmentJobID, insideID, legalA, assignmentStart}},
	} {
		if err := database.WithContext(ctx).Exec(fixture.query, fixture.args...).Error; err != nil {
			t.Fatalf("insert assignment fixture: %v", err)
		}
	}
	assignmentID := ""
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		if assignmentID != "" {
			if err := database.WithContext(cleanupCtx).Exec("UPDATE hr_workers SET primary_assignment_id = NULL WHERE id = ?", newWorkerID).Error; err != nil {
				t.Errorf("clear primary assignment fixture link: %v", err)
			}
			if err := database.WithContext(cleanupCtx).Exec("DELETE FROM hr_worker_assignments WHERE id = ?", assignmentID).Error; err != nil {
				t.Errorf("clean assignment: %v", err)
			}
		}
		for _, cleanup := range []struct {
			query string
			id    string
		}{
			{"DELETE FROM hr_positions WHERE id = ?", assignmentPositionID},
			{"DELETE FROM hr_jobs WHERE id = ?", assignmentJobID},
			{"DELETE FROM hr_job_families WHERE id = ?", assignmentFamilyID},
		} {
			if err := database.WithContext(cleanupCtx).Exec(cleanup.query, cleanup.id).Error; err != nil {
				t.Errorf("clean assignment fixture: %v", err)
			}
		}
	})
	assignmentType := "primary"
	assignmentInput := &pb.WorkerAssignmentInput{WorkerId: &newWorkerID,
		PositionId: &assignmentPositionID, AssignmentType: &assignmentType, ValidFrom: &assignmentStart}
	forgedApproval := *assignmentInput
	forgedApproval.ApprovedByPrincipalId = &insideID
	if _, err := client.CreateWorkerAssignment(ctx, &pb.CreateWorkerAssignmentRequest{Input: &forgedApproval}); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("incomplete assignment approval = %v", err)
	}
	forgedApproval.ApprovedAt = timestamppb.New(testNow.Add(time.Hour))
	if _, err := client.CreateWorkerAssignment(ctx, &pb.CreateWorkerAssignmentRequest{Input: &forgedApproval}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("forged assignment approval = %v", err)
	}
	assignmentResponse, err := client.CreateWorkerAssignment(ctx, &pb.CreateWorkerAssignmentRequest{Input: assignmentInput})
	if err != nil || assignmentResponse.GetItem().GetId() == "" ||
		assignmentResponse.GetItem().GetRequestedByPrincipalId() != "00000000-0000-4000-8000-000000000001" ||
		assignmentResponse.GetItem().GetRequestedAt() == nil || assignmentResponse.GetItem().GetApprovedAt() != nil {
		t.Fatalf("create draft assignment = %+v, %v", assignmentResponse, err)
	}
	assignmentID = assignmentResponse.GetItem().GetId()
	requesterID := "00000000-0000-4000-8000-000000000001"
	assignmentPage, err := client.FindWorkerAssignmentByRequestedByPrincipalID(ctx, &common.StringFinderRequest{Value: &requesterID, Limit: 1})
	if err != nil || assignmentPage.GetPage().GetTotal() != 1 || len(assignmentPage.GetItems()) != 1 {
		t.Fatalf("assignment requester finder = %+v, %v", assignmentPage, err)
	}
	if _, err := client.UpdateWorkerAssignment(ctx, &pb.UpdateWorkerAssignmentRequest{Id: assignmentID,
		Input:      &pb.WorkerAssignmentInput{AssignmentType: stringPointer("secondary")},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"assignment_type"}}}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("generic assignment update = %v", err)
	}
	if _, err := client.DeleteWorkerAssignment(ctx, &common.IDRequest{Id: assignmentID}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("generic assignment delete = %v", err)
	}
	if err := database.WithContext(ctx).Exec("UPDATE hr_worker_assignments SET approved_by_principal_id = ?, approved_at = now() WHERE id = ?", requesterID, assignmentID).Error; err == nil {
		t.Fatal("self approval bypassed database provenance constraint")
	}
	if _, err := client.ApproveWorkerAssignment(ctx, &common.IDRequest{Id: assignmentID}); status.Code(err) != codes.PermissionDenied {
		t.Fatalf("self approval = %v", err)
	}
	principalID = "00000000-0000-4000-8000-000000000009"
	if _, err := client.ApproveWorkerAssignment(ctx, &common.IDRequest{Id: assignmentID}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("approve planned position or preboarding worker = %v", err)
	}
	if err := database.WithContext(ctx).Exec("UPDATE hr_workers SET status = 'active' WHERE id = ?", newWorkerID).Error; err != nil {
		t.Fatalf("activate worker test fixture: %v", err)
	}
	if err := database.WithContext(ctx).Exec("UPDATE hr_positions SET position_status = 'active', headcount_authorized = 1 WHERE id = ?", assignmentPositionID).Error; err != nil {
		t.Fatalf("activate position test fixture: %v", err)
	}
	approvedAssignment, err := client.ApproveWorkerAssignment(ctx, &common.IDRequest{Id: assignmentID})
	if err != nil || approvedAssignment.GetItem().GetApprovedByPrincipalId() != principalID || approvedAssignment.GetItem().GetApprovedAt() == nil {
		t.Fatalf("approve assignment = %+v, %v", approvedAssignment, err)
	}
	if _, err := client.ApproveWorkerAssignment(ctx, &common.IDRequest{Id: assignmentID}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("repeat approval = %v", err)
	}
	if err := database.WithContext(ctx).Table("hr_outbox_events").Where("aggregate_id = ?", assignmentID).Count(&events).Error; err != nil || events != 2 {
		t.Fatalf("assignment outbox events = %d, %v", events, err)
	}
	capacityPositionID, err := ids.NewID()
	if err != nil {
		t.Fatal(err)
	}
	if err := database.WithContext(ctx).Exec("INSERT INTO hr_positions (id, position_code, position_title, job_id, organization_unit_reference_id, legal_entity_reference_id, position_status, headcount_authorized, valid_from) VALUES (?, ?, 'Associate', ?, ?, ?, 'active', 1, ?)",
		capacityPositionID, "POSITION-"+capacityPositionID, assignmentJobID, insideID, legalA, assignmentStart).Error; err != nil {
		t.Fatalf("insert capacity position fixture: %v", err)
	}
	capacityWorkerID, err := ids.NewID()
	if err != nil {
		t.Fatal(err)
	}
	if err := database.WithContext(ctx).Exec("INSERT INTO hr_workers (id, worker_number, person_party_reference_id, worker_category, status, primary_legal_entity_reference_id) VALUES (?, ?, ?, 'employee', 'active', ?)",
		capacityWorkerID, "CAPACITY-"+capacityWorkerID, capacityWorkerID, legalA).Error; err != nil {
		t.Fatalf("insert capacity worker fixture: %v", err)
	}
	capacityAssignmentIDs := make([]string, 0, 2)
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		for _, id := range capacityAssignmentIDs {
			if err := database.WithContext(cleanupCtx).Exec("DELETE FROM hr_worker_assignments WHERE id = ?", id).Error; err != nil {
				t.Errorf("clean capacity assignment: %v", err)
			}
		}
		if err := database.WithContext(cleanupCtx).Exec("DELETE FROM hr_positions WHERE id = ?", capacityPositionID).Error; err != nil {
			t.Errorf("clean capacity position: %v", err)
		}
		if err := database.WithContext(cleanupCtx).Exec("DELETE FROM hr_workers WHERE id = ?", capacityWorkerID).Error; err != nil {
			t.Errorf("clean capacity worker: %v", err)
		}
	})
	principalID = requesterID
	secondary := "secondary"
	fti := "60"
	for _, workerID := range []string{insideID, capacityWorkerID} {
		created, err := client.CreateWorkerAssignment(ctx, &pb.CreateWorkerAssignmentRequest{Input: &pb.WorkerAssignmentInput{
			WorkerId: &workerID, PositionId: &capacityPositionID, AssignmentType: &secondary,
			FtePercent: &fti, ValidFrom: &assignmentStart,
		}})
		if err != nil {
			t.Fatalf("create competing assignment: %v", err)
		}
		capacityAssignmentIDs = append(capacityAssignmentIDs, created.GetItem().GetId())
	}
	principalID = "00000000-0000-4000-8000-000000000009"
	approvalResults := make(chan error, 2)
	for _, id := range capacityAssignmentIDs {
		go func(id string) {
			_, err := client.ApproveWorkerAssignment(ctx, &common.IDRequest{Id: id})
			approvalResults <- err
		}(id)
	}
	successes, capacityRejections := 0, 0
	for range 2 {
		err := <-approvalResults
		switch status.Code(err) {
		case codes.OK:
			successes++
		case codes.FailedPrecondition:
			capacityRejections++
		default:
			t.Fatalf("unexpected concurrent approval = %v", err)
		}
	}
	if successes != 1 || capacityRejections != 1 {
		t.Fatalf("concurrent capacity approvals: success=%d rejected=%d", successes, capacityRejections)
	}
	if _, err := client.ApplyPrimaryAssignment(ctx, &pb.ApplyPrimaryAssignmentRequest{WorkerId: newWorkerID,
		AssignmentId: capacityPositionID}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("apply non-assignment reference = %v", err)
	}
	applyResults := make(chan error, 2)
	for range 2 {
		go func() {
			_, err := client.ApplyPrimaryAssignment(ctx, &pb.ApplyPrimaryAssignmentRequest{
				WorkerId: newWorkerID, AssignmentId: assignmentID,
			})
			applyResults <- err
		}()
	}
	applied, repeated := 0, 0
	for range 2 {
		switch code := status.Code(<-applyResults); code {
		case codes.OK:
			applied++
		case codes.FailedPrecondition:
			repeated++
		default:
			t.Fatalf("unexpected primary application status = %v", code)
		}
	}
	if applied != 1 || repeated != 1 {
		t.Fatalf("parallel primary application: applied=%d repeated=%d", applied, repeated)
	}
	appliedWorker, err := client.GetWorkerByID(ctx, &common.IDRequest{Id: newWorkerID})
	if err != nil || appliedWorker.GetItem().GetPrimaryAssignmentId() != assignmentID {
		t.Fatalf("persisted primary assignment = %+v, %v", appliedWorker, err)
	}
	if _, err := client.UpdateWorker(ctx, &pb.UpdateWorkerRequest{Id: newWorkerID,
		Input:      &pb.WorkerInput{PrimaryAssignmentId: &capacityPositionID},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"primary_assignment_id"}}}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("generic primary assignment update = %v", err)
	}
	if _, err := client.DeleteWorker(ctx, &common.IDRequest{Id: newWorkerID}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("delete employment history = %v", err)
	}
	if err := database.WithContext(ctx).Table("hr_outbox_events").Where("aggregate_id = ?", newWorkerID).Count(&events).Error; err != nil || events != 3 {
		t.Fatalf("worker outbox events = %d, %v", events, err)
	}
}

func stringPointer(value string) *string { return &value }
