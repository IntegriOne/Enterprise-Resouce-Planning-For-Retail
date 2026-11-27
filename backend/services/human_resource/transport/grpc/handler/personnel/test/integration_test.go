//go:build integration

package test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/url"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/personnel/v1"
	hrdb "github.com/integrione/erp-retail/backend/services/human_resource/db"
	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	"github.com/integrione/erp-retail/backend/services/human_resource/infrastructure/security"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/personnel"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository"
	handler "github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/personnel"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/interceptor"
	"github.com/integrione/erp-retail/backend/shared/authcontext"
	"github.com/integrione/erp-retail/backend/shared/authsession"
	"github.com/integrione/erp-retail/backend/shared/authz"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type workerScopePolicy struct {
	workerReads *security.Policy
	principalID *string
}

const scopedWorkerPrincipal = "00000000-0000-4000-8000-000000000001"
const scopedWorkerSession = "00000000-0000-4000-8000-000000000002"

var scopedWorkerTokenHash = strings.Repeat("a", 64)

type workerSessionProjection struct{ legalEntityID string }

func (p workerSessionProjection) ReadSessionContext(_ context.Context, tokenHash string, now time.Time) (*authsession.SessionValidationContext, error) {
	if tokenHash != scopedWorkerTokenHash {
		return nil, authsession.ErrSessionNotFound
	}
	return &authsession.SessionValidationContext{
		Gate:     &authsession.SecurityGateRecord{Open: true, RecoveryEpoch: 1},
		Token:    &authsession.SessionTokenRecord{UserID: scopedWorkerPrincipal, SessionID: scopedWorkerSession, UserKind: "employee"},
		Security: &authsession.UserSecurityRecord{UserID: scopedWorkerPrincipal, AuthEpoch: 1, PermRevision: 1, RecoveryEpoch: 1},
		Snapshot: &authz.AccessSnapshot{
			PrincipalID: scopedWorkerPrincipal, UserKind: "employee", Status: authz.StatusActive,
			AuthEpoch: 1, PermRevision: 1, RecoveryEpoch: 1,
			SyncedAtUnix: now.Unix(), FreshUntilUnix: now.Add(authz.DefaultSnapshotMaxAge).Unix(),
			Permissions: []authz.PermissionGrant{{
				Code: "human_resource.workers.read", Scope: authz.ScopeGrant{Type: "legal_entity", ID: p.legalEntityID},
			}},
		},
	}, nil
}

type workerFixedClock struct{ now time.Time }

func (c workerFixedClock) Now() time.Time { return c.now }

type principalLinkVerifier struct{ expected service.PrincipalLinkCheck }

func (v *principalLinkVerifier) VerifyPrincipalLink(_ context.Context, check service.PrincipalLinkCheck) error {
	if check != v.expected {
		return hrerrors.ErrReferenceConflict
	}
	return nil
}

type contractSignatureVerifier struct {
	contractID, workerID, documentID, versionID string
	signedAt                                    time.Time
}

func (v *contractSignatureVerifier) VerifyContractSignature(_ context.Context, check service.ContractSignatureCheck) (service.ContractSignatureProof, error) {
	if check.Contract.ID != v.contractID || check.Contract.WorkerID != v.workerID ||
		check.DocumentVersionID != v.versionID || check.PrincipalID == "" {
		return service.ContractSignatureProof{}, hrerrors.ErrReferenceConflict
	}
	return service.ContractSignatureProof{
		DocumentID: v.documentID, DocumentVersionID: v.versionID, SignedAt: v.signedAt, ContentDigest: make([]byte, 32),
	}, nil
}

type workerReferenceVerifier struct{}

func (workerReferenceVerifier) VerifyPositionReferences(_ context.Context, check security.PositionReferenceCheck) error {
	if check.OrganizationUnitID == "" || check.LegalEntityID == "" {
		return hrerrors.ErrReferenceConflict
	}
	return nil
}

func (workerReferenceVerifier) VerifyWorkerReferences(_ context.Context, check security.WorkerReferenceCheck) error {
	if check.PersonPartyID == "" || check.LegalEntityID == "" {
		return hrerrors.ErrReferenceConflict
	}
	return nil
}

func (p workerScopePolicy) Authorize(ctx context.Context, access operation.Access) (operation.Actor, error) {
	if p.workerReads != nil && access.Resource == "hr_workers" && (access.Action == "read" || access.Action == "list") {
		return p.workerReads.Authorize(ctx, access)
	}
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

func (p workerScopePolicy) ReadScope(ctx context.Context, access operation.Access) (operation.Scope, error) {
	if p.workerReads == nil || access.Resource != "hr_workers" {
		return operation.Scope{All: true}, nil
	}
	return p.workerReads.ReadScope(ctx, access)
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
	sessions := workerSessionReader(t, ctx, workerSessionProjection{legalEntityID: legalA})
	workerReads, err := security.NewPolicy(sessions, security.PermissionCatalog())
	if err != nil {
		t.Fatal(err)
	}
	principalID := "00000000-0000-4000-8000-000000000001"
	testNow := time.Date(2026, 11, 22, 12, 0, 0, 0, time.UTC)
	clock := &workerFixedClock{now: testNow}
	linkVerifier := &principalLinkVerifier{}
	signatureVerifier := &contractSignatureVerifier{}
	svc, err := service.NewService(service.Dependencies{
		PrincipalLinkVerifier:     linkVerifier,
		ContractSignatureVerifier: signatureVerifier,
		Runtime: operation.Dependencies{
			Transactor: repos.Transactor, Outbox: repos.OutboxEvent,
			Policy:    workerScopePolicy{workerReads: workerReads, principalID: &principalID},
			Integrity: security.NewIntegrity(validators), Clock: clock,
		},
		Worker: repos.Worker, EmploymentContract: repos.EmploymentContract,
		WorkerAssignment: repos.WorkerAssignment, WorkerLifecycleEvent: repos.WorkerLifecycleEvent,
		WorkerAssignmentTransfer: repos.WorkerAssignmentTransfer, TransferPosition: repos.Position,
		WorkerDocument: repos.WorkerDocument, EmergencyContact: repos.EmergencyContact,
	})
	if err != nil {
		t.Fatal(err)
	}
	listener := bufconn.Listen(1 << 20)
	gatewayKey := []byte(strings.Repeat("k", 32))
	guard := interceptor.Unary(gatewayKey, slog.New(slog.NewTextHandler(io.Discard, nil)))
	server := grpc.NewServer(grpc.UnaryInterceptor(guard))
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
	if _, err := client.GetWorkerByID(ctx, &common.IDRequest{Id: insideID}); status.Code(err) != codes.Unauthenticated {
		t.Fatalf("worker query without gateway proof = %v, want Unauthenticated", err)
	}
	ctx = authcontext.AppendToOutgoingGRPCContext(ctx, authcontext.CallerContext{
		Authenticated: true, PrincipalID: scopedWorkerPrincipal, SessionID: scopedWorkerSession,
		UserKind: "employee", AuthEpoch: 1, PermRevision: 1,
	}, gatewayKey, time.Now())
	if _, err := client.GetWorkerByID(ctx, &common.IDRequest{Id: insideID}); status.Code(err) != codes.Unauthenticated {
		t.Fatalf("worker query without session hash = %v, want Unauthenticated", err)
	}
	unknownSession := metadata.AppendToOutgoingContext(ctx, "x-ig-session-token-hash", strings.Repeat("b", 64))
	if _, err := client.GetWorkerByID(unknownSession, &common.IDRequest{Id: insideID}); status.Code(err) != codes.Unauthenticated {
		t.Fatalf("worker query with unknown session = %v, want Unauthenticated", err)
	}
	ctx = metadata.AppendToOutgoingContext(ctx, "x-ig-session-token-hash", scopedWorkerTokenHash)
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
	foundByNumber, err := client.FindWorkerByWorkerNumber(ctx, &common.StringFinderRequest{
		Value: stringPointer("SCOPE-" + insideID), Limit: 1,
	})
	if err != nil || foundByNumber.GetPage().GetTotal() != 1 ||
		len(foundByNumber.GetItems()) != 1 || foundByNumber.GetItems()[0].GetId() != insideID {
		t.Fatalf("inside-scope worker number finder = %+v, %v", foundByNumber, err)
	}
	hiddenByNumber, err := client.FindWorkerByWorkerNumber(ctx, &common.StringFinderRequest{
		Value: stringPointer("SCOPE-" + outsideID), Limit: 1,
	})
	if err != nil || hiddenByNumber.GetPage().GetTotal() != 0 || len(hiddenByNumber.GetItems()) != 0 {
		t.Fatalf("outside-scope worker number finder = %+v, %v", hiddenByNumber, err)
	}
	contractNumber := "DRAFT-" + insideID
	contractInput := &pb.EmploymentContractInput{
		WorkerId: &insideID, ContractNumber: &contractNumber,
		ContractType: stringPointer("fixed_term"), EmployingLegalEntityReferenceId: &legalA,
		EffectiveFrom: stringPointer("2026-11-17"), WorkCountryCode: stringPointer("ID"),
		GoverningJurisdictionCode: stringPointer("ID"),
	}
	signedInput := proto.CloneOf(contractInput)
	signedInput.SignedDocumentReferenceId = stringPointer(insideID)
	if _, err := client.CreateEmploymentContract(ctx, &pb.CreateEmploymentContractRequest{Input: signedInput}); status.Code(err) != codes.FailedPrecondition {
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
		if err := database.WithContext(cleanupCtx).Exec("DELETE FROM hr_worker_lifecycle_events WHERE worker_id = ?", newWorkerID).Error; err != nil {
			t.Errorf("clean worker lifecycle events: %v", err)
		}
		if err := database.WithContext(cleanupCtx).Exec("DELETE FROM hr_workers WHERE id = ?", newWorkerID).Error; err != nil {
			t.Errorf("clean preboarding worker: %v", err)
		}
	})
	if _, err := client.CreateWorker(ctx, &pb.CreateWorkerRequest{Input: workerInput}); status.Code(err) != codes.AlreadyExists {
		t.Fatalf("duplicate worker number = %v", err)
	}
	linkedInput := proto.CloneOf(workerInput)
	linkedInput.IamPrincipalReferenceId = &insideID
	if _, err := client.CreateWorker(ctx, &pb.CreateWorkerRequest{Input: linkedInput}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("IAM link through worker CRUD = %v", err)
	}
	if _, err := client.UpdateWorker(ctx, &pb.UpdateWorkerRequest{
		Id: newWorkerID,
		Input: &pb.WorkerInput{
			IamPrincipalReferenceId: &insideID,
			ServiceStartDate:        stringPointer("2026-11-22"),
		},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"iam_principal_reference_id", "service_start_date"}},
	}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("IAM link through worker update = %v", err)
	}
	afterRejectedLink, err := client.GetWorkerByID(ctx, &common.IDRequest{Id: newWorkerID})
	if err != nil || afterRejectedLink.GetItem().GetIamPrincipalReferenceId() != "" ||
		afterRejectedLink.GetItem().GetServiceStartDate() != "2026-11-21" {
		t.Fatalf("rejected worker update changed stored data: %+v, %v", afterRejectedLink, err)
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
		PositionId: &assignmentPositionID, AssignmentType: &assignmentType, ValidFrom: &assignmentStart,
		ManagerWorkerId: &insideID}
	if err := database.WithContext(ctx).Exec("UPDATE hr_workers SET hire_date = ? WHERE id = ?", "2026-11-23", insideID).Error; err != nil {
		t.Fatalf("set manager start date: %v", err)
	}
	if _, err := client.CreateWorkerAssignment(ctx, &pb.CreateWorkerAssignmentRequest{Input: assignmentInput}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("create assignment before manager employment = %v", err)
	}
	if err := database.WithContext(ctx).Exec("UPDATE hr_workers SET hire_date = ? WHERE id = ?", assignmentStart, insideID).Error; err != nil {
		t.Fatalf("align manager start date: %v", err)
	}
	forgedApproval := proto.CloneOf(assignmentInput)
	forgedApproval.ApprovedByPrincipalId = &insideID
	if _, err := client.CreateWorkerAssignment(ctx, &pb.CreateWorkerAssignmentRequest{Input: forgedApproval}); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("incomplete assignment approval = %v", err)
	}
	forgedApproval.ApprovedAt = timestamppb.New(testNow.Add(time.Hour))
	if _, err := client.CreateWorkerAssignment(ctx, &pb.CreateWorkerAssignmentRequest{Input: forgedApproval}); status.Code(err) != codes.FailedPrecondition {
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
	if _, err := client.ActivateWorker(ctx, &common.IDRequest{Id: newWorkerID}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("activate worker without signed contract = %v", err)
	}
	contractNumberForActivation := "SIGNED-" + newWorkerID
	draftContract, err := client.CreateEmploymentContract(ctx, &pb.CreateEmploymentContractRequest{Input: &pb.EmploymentContractInput{
		WorkerId: &newWorkerID, ContractNumber: &contractNumberForActivation, ContractType: stringPointer("permanent"),
		EmployingLegalEntityReferenceId: &legalA, EffectiveFrom: stringPointer("2026-11-20"),
		WorkCountryCode: stringPointer("ID"), GoverningJurisdictionCode: stringPointer("ID"),
	}})
	if err != nil {
		t.Fatalf("create contract for worker activation: %v", err)
	}
	signedContractID := draftContract.GetItem().GetId()
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		for _, query := range []string{
			"DELETE FROM hr_outbox_events WHERE aggregate_id = ?",
			"DELETE FROM hr_employment_contracts WHERE id = ?",
		} {
			if err := database.WithContext(cleanupCtx).Exec(query, signedContractID).Error; err != nil {
				t.Errorf("clean signed contract fixture: %v", err)
			}
		}
	})
	*signatureVerifier = contractSignatureVerifier{
		contractID: signedContractID, workerID: newWorkerID, documentID: insideID, versionID: outsideID, signedAt: testNow,
	}
	signature := &pb.RecordSignedEmploymentContractRequest{ContractId: signedContractID, DocumentVersionId: insideID}
	if _, err := client.RecordSignedEmploymentContract(ctx, signature); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("record signature for wrong document version: %v", err)
	}
	signature.DocumentVersionId = outsideID
	signed, err := client.RecordSignedEmploymentContract(ctx, signature)
	if err != nil || signed.GetItem().GetContractStatus() != "signed" ||
		signed.GetItem().GetSignedDocumentVersionReferenceId() != outsideID ||
		len(signed.GetItem().GetSignedContentDigest()) != 32 {
		t.Fatalf("record verified contract signature: %+v, %v", signed, err)
	}
	if _, err := client.RecordSignedEmploymentContract(ctx, signature); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("repeat contract signature: %v", err)
	}
	activation := &common.IDRequest{Id: newWorkerID}
	lifecycleQuery := &common.StringFinderRequest{Value: &newWorkerID, Limit: 10}
	var outboxBeforeActivation int64
	if err := database.WithContext(ctx).Table("hr_outbox_events").Where("aggregate_id = ?", newWorkerID).
		Count(&outboxBeforeActivation).Error; err != nil {
		t.Fatalf("count worker outbox before activation: %v", err)
	}
	if err := database.WithContext(ctx).Exec("UPDATE hr_employment_contracts SET effective_from = ? WHERE id = ?",
		testNow.Format("2006-01-02"), signedContractID).Error; err != nil {
		t.Fatalf("move contract start after hire date: %v", err)
	}
	if _, err := client.ActivateWorker(ctx, activation); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("activate worker with contract starting after hire date = %v", err)
	}
	if err := database.WithContext(ctx).Exec("UPDATE hr_employment_contracts SET effective_from = ?, effective_to = ? WHERE id = ?",
		"2026-11-20", testNow.AddDate(0, 0, -1).Format("2006-01-02"), signedContractID).Error; err != nil {
		t.Fatalf("expire signed contract fixture: %v", err)
	}
	if _, err := client.ActivateWorker(ctx, activation); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("activate worker with expired signed contract = %v", err)
	}
	if history, err := client.FindWorkerLifecycleEventByWorkerID(ctx, lifecycleQuery); err != nil || len(history.GetItems()) != 0 {
		t.Fatalf("failed activation lifecycle = %+v, %v", history, err)
	}
	var outboxAfterFailure int64
	if err := database.WithContext(ctx).Table("hr_outbox_events").Where("aggregate_id = ?", newWorkerID).
		Count(&outboxAfterFailure).Error; err != nil || outboxAfterFailure != outboxBeforeActivation {
		t.Fatalf("failed activation outbox count changed from %d to %d: %v", outboxBeforeActivation, outboxAfterFailure, err)
	}
	if err := database.WithContext(ctx).Exec("UPDATE hr_employment_contracts SET effective_to = ? WHERE id = ?",
		testNow.Format("2006-01-02"), signedContractID).Error; err != nil {
		t.Fatalf("set signed contract end to activation date: %v", err)
	}
	activatedWorker, err := client.ActivateWorker(ctx, activation)
	if err != nil || activatedWorker.GetItem().GetStatus() != "active" {
		t.Fatalf("activate worker = %+v, %v", activatedWorker, err)
	}
	if _, err := client.ActivateWorker(ctx, &common.IDRequest{Id: newWorkerID}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("repeat worker activation = %v", err)
	}
	history, err := client.FindWorkerLifecycleEventByWorkerID(ctx, lifecycleQuery)
	if err != nil || len(history.GetItems()) != 1 || history.GetPage().GetTotal() != 1 {
		t.Fatalf("worker lifecycle = %+v, %v", history, err)
	}
	hire := history.GetItems()[0]
	if hire.GetEventType() != "hire" || hire.GetPreviousStatus() != "preboarding" || hire.GetNewStatus() != "active" ||
		hire.GetActorPrincipalId() != principalID || hire.GetEffectiveDate() != "2026-11-20" ||
		!hire.GetOccurredAt().AsTime().Equal(testNow) {
		t.Fatalf("unexpected hire lifecycle event: %+v", hire)
	}
	if err := database.WithContext(ctx).Exec("UPDATE hr_positions SET position_status = 'active', headcount_authorized = 1 WHERE id = ?", assignmentPositionID).Error; err != nil {
		t.Fatalf("activate position test fixture: %v", err)
	}
	if err := database.WithContext(ctx).Exec("UPDATE hr_workers SET hire_date = ? WHERE id = ?", "2026-11-23", insideID).Error; err != nil {
		t.Fatalf("correct manager start date: %v", err)
	}
	if _, err := client.ApproveWorkerAssignment(ctx, &common.IDRequest{Id: assignmentID}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("approve assignment before manager employment = %v", err)
	}
	if err := database.WithContext(ctx).Exec("UPDATE hr_workers SET hire_date = ? WHERE id = ?", assignmentStart, insideID).Error; err != nil {
		t.Fatalf("restore manager start date: %v", err)
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
	positionEnd := testNow.AddDate(0, 0, 1).Format("2006-01-02")
	positionChange := database.WithContext(ctx).Begin()
	if positionChange.Error != nil {
		t.Fatal(positionChange.Error)
	}
	defer positionChange.Rollback()
	if err := positionChange.Exec("UPDATE hr_positions SET valid_to = ? WHERE id = ?", positionEnd, assignmentPositionID).Error; err != nil {
		t.Fatalf("shorten approved assignment position validity: %v", err)
	}
	positionRace := make(chan error, 1)
	go func() {
		_, err := client.ApplyPrimaryAssignment(ctx, &pb.ApplyPrimaryAssignmentRequest{
			WorkerId: newWorkerID, AssignmentId: assignmentID,
		})
		positionRace <- err
	}()
	select {
	case err := <-positionRace:
		t.Fatalf("assignment application finished before position change committed: %v", err)
	case <-time.After(100 * time.Millisecond):
	}
	if err := positionChange.Commit().Error; err != nil {
		t.Fatalf("commit position validity change: %v", err)
	}
	if err := <-positionRace; status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("apply open-ended assignment to position with an end date = %v", err)
	}
	unchangedWorker, err := client.GetWorkerByID(ctx, &common.IDRequest{Id: newWorkerID})
	if err != nil || unchangedWorker.GetItem().GetPrimaryAssignmentId() != "" {
		t.Fatalf("primary assignment changed after failed application: %+v, %v", unchangedWorker, err)
	}
	if err := database.WithContext(ctx).Exec("UPDATE hr_worker_assignments SET valid_to = ? WHERE id = ?", positionEnd, assignmentID).Error; err != nil {
		t.Fatalf("align assignment end date with position: %v", err)
	}
	if err := database.WithContext(ctx).Exec("UPDATE hr_workers SET hire_date = ? WHERE id = ?", positionEnd, insideID).Error; err != nil {
		t.Fatalf("correct manager hire date after approval: %v", err)
	}
	if _, err := client.ApplyPrimaryAssignment(ctx, &pb.ApplyPrimaryAssignmentRequest{
		WorkerId: newWorkerID, AssignmentId: assignmentID,
	}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("apply primary assignment before manager hire date = %v", err)
	}
	if err := database.WithContext(ctx).Exec("UPDATE hr_workers SET hire_date = ? WHERE id = ?", assignmentStart, insideID).Error; err != nil {
		t.Fatalf("restore manager hire date: %v", err)
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
	if err := database.WithContext(ctx).Table("hr_outbox_events").Where("aggregate_id = ?", newWorkerID).Count(&events).Error; err != nil || events != 4 {
		t.Fatalf("worker outbox events = %d, %v", events, err)
	}
	transferKey, err := ids.NewID()
	if err != nil {
		t.Fatal(err)
	}
	submitTransfer := &pb.SubmitWorkerAssignmentTransferRequest{
		WorkerId: newWorkerID, SourceAssignmentId: assignmentID, TargetPositionId: capacityPositionID,
		EffectiveDate: testNow.AddDate(0, 0, 1).Format(time.DateOnly), ReasonCode: "rotation", IdempotencyKey: transferKey,
	}
	sourceBeforeTransfer, err := repos.WorkerAssignment.SelectByID(ctx, assignmentID)
	if err != nil {
		t.Fatal(err)
	}
	submittedTransfer, err := client.SubmitWorkerAssignmentTransfer(ctx, submitTransfer)
	if err != nil || submittedTransfer.GetItem().GetStatus() != "submitted" {
		t.Fatalf("submit transfer with active source: %+v, %v", submittedTransfer, err)
	}
	t.Cleanup(func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := database.WithContext(cleanupCtx).Exec("DELETE FROM hr_worker_assignment_transfers WHERE idempotency_key = ?", transferKey).Error; err != nil {
			t.Errorf("delete submitted transfer fixture: %v", err)
		}
	})
	readTransfer, err := client.GetWorkerAssignmentTransfer(ctx, &common.IDRequest{Id: submittedTransfer.GetItem().GetId()})
	if err != nil || !proto.Equal(readTransfer.GetItem(), submittedTransfer.GetItem()) {
		t.Fatalf("read submitted transfer: %+v, %v", readTransfer, err)
	}
	if _, err := client.GetWorkerAssignmentTransfer(ctx, &common.IDRequest{Id: "invalid"}); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("malformed transfer ID: %v", err)
	}

	transferPage, err := client.ListWorkerAssignmentTransfers(ctx, &pb.ListWorkerAssignmentTransfersRequest{WorkerId: newWorkerID, Limit: 1})
	if err != nil || len(transferPage.GetItems()) != 1 || transferPage.GetPage().GetTotal() != 1 ||
		!proto.Equal(transferPage.GetItems()[0], submittedTransfer.GetItem()) {
		t.Fatalf("list submitted transfer: %+v, %v", transferPage, err)
	}
	afterTransfer, err := client.ListWorkerAssignmentTransfers(ctx, &pb.ListWorkerAssignmentTransfersRequest{WorkerId: newWorkerID, Limit: 1, Offset: 1})
	if err != nil || len(afterTransfer.GetItems()) != 0 || afterTransfer.GetPage().GetTotal() != 1 {
		t.Fatalf("list after submitted transfer: %+v, %v", afterTransfer, err)
	}

	replayedTransfer, err := client.SubmitWorkerAssignmentTransfer(ctx, submitTransfer)
	if err != nil || replayedTransfer.GetItem().GetId() != submittedTransfer.GetItem().GetId() {
		t.Fatalf("retry transfer submission: %+v, %v", replayedTransfer, err)
	}
	changedTransfer := proto.CloneOf(submitTransfer)
	changedTransfer.ReasonCode = "changed"
	if _, err := client.SubmitWorkerAssignmentTransfer(ctx, changedTransfer); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("retry transfer with changed payload: %v", err)
	}
	sourceAfterTransfer, err := repos.WorkerAssignment.SelectByID(ctx, assignmentID)
	if err != nil || !reflect.DeepEqual(sourceBeforeTransfer, sourceAfterTransfer) {
		t.Fatalf("submission altered source assignment: %+v, %v", sourceAfterTransfer, err)
	}
	previousDay := testNow.AddDate(0, 0, -1).Format("2006-01-02")
	if err := database.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("UPDATE hr_positions SET valid_from = ? WHERE id = ?", previousDay, assignmentPositionID).Error; err != nil {
			return err
		}
		return tx.Exec("UPDATE hr_worker_assignments SET valid_from = ?, valid_to = ? WHERE id = ?",
			previousDay, previousDay, assignmentID).Error
	}); err != nil {
		t.Fatalf("prepare ended primary assignment: %v", err)
	}
	approverID := principalID
	principalID = requesterID
	replacementFTE := "40"
	replacement, err := client.CreateWorkerAssignment(ctx, &pb.CreateWorkerAssignmentRequest{Input: &pb.WorkerAssignmentInput{
		WorkerId: &newWorkerID, PositionId: &capacityPositionID, AssignmentType: &assignmentType,
		ValidFrom: &assignmentStart, FtePercent: &replacementFTE,
	}})
	if err != nil {
		t.Fatalf("create replacement primary assignment: %v", err)
	}
	replacementID := replacement.GetItem().GetId()
	capacityAssignmentIDs = append(capacityAssignmentIDs, replacementID)
	t.Cleanup(func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := database.WithContext(cleanupCtx).Exec("UPDATE hr_workers SET primary_assignment_id = NULL WHERE id = ?", newWorkerID).Error; err != nil {
			t.Errorf("clear replacement primary assignment: %v", err)
		}
	})
	principalID = approverID
	if _, err := client.ApproveWorkerAssignment(ctx, &common.IDRequest{Id: replacementID}); err != nil {
		t.Fatalf("approve replacement primary assignment: %v", err)
	}
	transferID, err := ids.NewID()
	if err != nil {
		t.Fatal(err)
	}
	transferRequest := &entity.WorkerAssignmentTransfer{
		ID: transferID, WorkerID: newWorkerID, SourceAssignmentID: assignmentID,
		TargetPositionID: capacityPositionID, EffectiveDate: testNow, ReasonCode: "rotation",
		IdempotencyKey: transferID, Status: "submitted", RequestedByPrincipalID: requesterID, RequestedAt: testNow,
	}
	rollback := errors.New("rollback transfer repository fixture")
	err = repos.Transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := repos.WorkerAssignmentTransfer.Insert(txCtx, transferRequest); err != nil {
			return err
		}
		replayed, err := repos.WorkerAssignmentTransfer.FindByIdempotencyKey(txCtx, transferID)
		if err != nil {
			return err
		}
		if replayed.ID != transferID || replayed.Status != "submitted" {
			return errors.New("unexpected stored transfer request")
		}
		locked, err := repos.WorkerAssignmentTransfer.FindByIDForUpdate(txCtx, transferID)
		if err != nil {
			return err
		}
		locked.Status = "rejected"
		locked.DecidedByPrincipalID, locked.DecidedAt = &approverID, &testNow
		if err := repos.WorkerAssignmentTransfer.UpdateDecision(txCtx, locked); err != nil {
			return err
		}
		if err := repos.WorkerAssignmentTransfer.UpdateDecision(txCtx, locked); !errors.Is(err, hrerrors.ErrConflict) {
			return errors.New("terminal transfer decision accepted a second update")
		}
		return rollback
	})
	if !errors.Is(err, rollback) {
		t.Fatalf("transfer repository transaction: %v", err)
	}
	if _, err := repos.WorkerAssignmentTransfer.FindByIdempotencyKey(ctx, transferID); !errors.Is(err, hrerrors.ErrNotFound) {
		t.Fatalf("transfer repository rollback: %v", err)
	}
	replaceRequest := &pb.ApplyPrimaryAssignmentRequest{WorkerId: newWorkerID, AssignmentId: replacementID}
	replaced, err := client.ApplyPrimaryAssignment(ctx, replaceRequest)
	if err != nil || replaced.GetItem().GetPrimaryAssignmentId() != replacementID {
		t.Fatalf("apply replacement primary assignment = %+v, %v", replaced, err)
	}
	if _, err := client.ApplyPrimaryAssignment(ctx, replaceRequest); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("repeat replacement application: %v", err)
	}
	history, err = client.FindWorkerLifecycleEventByWorkerID(ctx, lifecycleQuery)
	if err != nil || len(history.GetItems()) != 2 || history.GetPage().GetTotal() != 2 {
		t.Fatalf("replacement worker lifecycle = %+v, %v", history, err)
	}
	transfers := 0
	for _, entry := range history.GetItems() {
		if entry.GetEventType() == "hire" {
			continue
		}
		if entry.GetEventType() != "transfer" || entry.GetSourceReference() != replacementID ||
			entry.GetActorPrincipalId() != approverID || entry.GetEffectiveDate() != assignmentStart {
			t.Fatalf("unexpected lifecycle after replacement: %+v", entry)
		}
		transfers++
	}
	if transfers != 1 {
		t.Fatalf("transfer event count = %d, want 1", transfers)
	}
	if err := database.WithContext(ctx).Table("hr_outbox_events").Where("aggregate_id = ?", newWorkerID).Count(&events).Error; err != nil || events != 5 {
		t.Fatalf("worker outbox events after replacement = %d, %v", events, err)
	}
	decisionKey, err := ids.NewID()
	if err != nil {
		t.Fatal(err)
	}
	principalID = requesterID
	decisionFixture, err := client.SubmitWorkerAssignmentTransfer(ctx, &pb.SubmitWorkerAssignmentTransferRequest{
		WorkerId: newWorkerID, SourceAssignmentId: replacementID, TargetPositionId: assignmentPositionID,
		EffectiveDate: positionEnd, ReasonCode: "rotation", IdempotencyKey: decisionKey,
	})
	if err != nil {
		t.Fatalf("submit transfer for decision: %v", err)
	}
	decisionID := decisionFixture.GetItem().GetId()
	var appliedID string
	t.Cleanup(func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := database.WithContext(cleanupCtx).Transaction(func(tx *gorm.DB) error {
			if err := tx.Exec("DELETE FROM hr_worker_assignment_transfers WHERE id = ?", decisionID).Error; err != nil {
				return err
			}
			if appliedID == "" {
				return nil
			}
			if err := tx.Exec("UPDATE hr_workers SET primary_assignment_id = NULL WHERE id = ?", newWorkerID).Error; err != nil {
				return err
			}
			return tx.Exec("DELETE FROM hr_worker_assignments WHERE id = ?", appliedID).Error
		}); err != nil {
			t.Errorf("clean applied transfer fixture: %v", err)
		}
	})
	decision := &pb.DecideWorkerAssignmentTransferRequest{
		TransferId: decisionID, Decision: pb.TransferDecision_TRANSFER_DECISION_APPROVE,
	}
	if _, err := client.DecideWorkerAssignmentTransfer(ctx, decision); status.Code(err) != codes.PermissionDenied {
		t.Fatalf("approve own transfer: %v", err)
	}
	principalID = approverID
	if _, err := client.DecideWorkerAssignmentTransfer(ctx, decision); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("apply transfer before effective date: %v", err)
	}
	clock.now = testNow.AddDate(0, 0, 1)
	if err := database.WithContext(ctx).Exec("UPDATE hr_positions SET headcount_authorized = 0.25 WHERE id = ?", assignmentPositionID).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := client.DecideWorkerAssignmentTransfer(ctx, decision); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("apply transfer above target capacity: %v", err)
	}
	stillOpen, err := repos.WorkerAssignment.SelectByID(ctx, replacementID)
	if err != nil || stillOpen.ValidTo != nil {
		t.Fatalf("failed transfer did not roll back source: %+v, %v", stillOpen, err)
	}
	if err := database.WithContext(ctx).Exec("UPDATE hr_positions SET headcount_authorized = 1 WHERE id = ?", assignmentPositionID).Error; err != nil {
		t.Fatal(err)
	}
	if err := database.WithContext(ctx).Exec("UPDATE hr_positions SET position_status = 'inactive' WHERE id = ?", assignmentPositionID).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := client.DecideWorkerAssignmentTransfer(ctx, decision); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("apply transfer to position deactivated after submission: %v", err)
	}
	if err := database.WithContext(ctx).Exec("UPDATE hr_positions SET position_status = 'active' WHERE id = ?", assignmentPositionID).Error; err != nil {
		t.Fatal(err)
	}

	effectiveDay := clock.now
	if err := database.WithContext(ctx).Exec("UPDATE hr_positions SET valid_to = ? WHERE id = ?", effectiveDay, assignmentPositionID).Error; err != nil {
		t.Fatal(err)
	}
	clock.now = effectiveDay.AddDate(0, 0, 1)
	if _, err := client.DecideWorkerAssignmentTransfer(ctx, decision); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("apply transfer after target expired despite valid effective date: %v", err)
	}
	if err := database.WithContext(ctx).Exec("UPDATE hr_positions SET valid_to = NULL WHERE id = ?", assignmentPositionID).Error; err != nil {
		t.Fatal(err)
	}
	clock.now = effectiveDay

	decided, err := client.DecideWorkerAssignmentTransfer(ctx, decision)
	if err != nil || decided.GetItem().GetStatus() != "applied" || decided.GetItem().GetAppliedAssignmentId() == "" {
		t.Fatalf("apply transfer: %+v, %v", decided, err)
	}
	appliedID = decided.GetItem().GetAppliedAssignmentId()
	replayedDecision, err := client.DecideWorkerAssignmentTransfer(ctx, decision)
	if err != nil || replayedDecision.GetItem().GetAppliedAssignmentId() != appliedID {
		t.Fatalf("replay transfer decision: %+v, %v", replayedDecision, err)
	}
	storedWorker, err := repos.Worker.SelectByID(ctx, newWorkerID)
	if err != nil || storedWorker.PrimaryAssignmentID == nil || *storedWorker.PrimaryAssignmentID != appliedID {
		t.Fatalf("transfer did not replace worker primary assignment: %+v, %v", storedWorker, err)
	}
	closedSource, err := repos.WorkerAssignment.SelectByID(ctx, replacementID)
	if err != nil || closedSource.ValidTo == nil || entity.CompareDates(*closedSource.ValidTo, testNow) != 0 {
		t.Fatalf("transfer did not close source at H-1: %+v, %v", closedSource, err)
	}
	if err := database.WithContext(ctx).Table("hr_outbox_events").Where("aggregate_id = ?", decisionID).Count(&events).Error; err != nil || events != 1 {
		t.Fatalf("transfer decision outbox count = %d, %v", events, err)
	}
	if err := database.WithContext(ctx).Table("hr_worker_lifecycle_events").
		Where("worker_id = ? AND source_reference = ? AND event_type = ?", newWorkerID, decisionID, "transfer").
		Count(&events).Error; err != nil || events != 1 {
		t.Fatalf("applied transfer lifecycle count = %d, %v", events, err)
	}
	targetPrincipalID, err := ids.NewID()
	if err != nil {
		t.Fatal(err)
	}
	linkVerifier.expected = service.PrincipalLinkCheck{
		EvidenceID: insideID, WorkerID: newWorkerID, PersonPartyID: personPartyID,
		PrincipalID: targetPrincipalID, RequestedByPrincipalID: principalID,
	}
	link := &pb.LinkWorkerPrincipalRequest{
		WorkerId: newWorkerID, PrincipalId: targetPrincipalID, EvidenceId: outsideID,
	}
	if _, err := client.LinkWorkerPrincipal(ctx, link); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("link with mismatched provisioning evidence: %v", err)
	}
	link.EvidenceId = insideID
	historicalWorkerID, err := ids.NewID()
	if err != nil {
		t.Fatal(err)
	}
	if err := database.WithContext(ctx).Exec(`
		INSERT INTO hr_workers (id, worker_number, person_party_reference_id, worker_category, status,
		 primary_legal_entity_reference_id, iam_principal_reference_id, termination_date)
		VALUES (?, ?, ?, 'employee', 'terminated', ?, ?, CURRENT_DATE)`,
		historicalWorkerID, "history-"+historicalWorkerID, outsideID, legalA, targetPrincipalID).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := client.LinkWorkerPrincipal(ctx, link); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("link principal with history belonging to a different person: %v", err)
	}
	if err := database.WithContext(ctx).Table("hr_workers").Where("id = ?", historicalWorkerID).
		Update("person_party_reference_id", personPartyID).Error; err != nil {
		t.Fatal(err)
	}
	linked, err := client.LinkWorkerPrincipal(ctx, link)
	if err != nil || linked.GetItem().GetIamPrincipalReferenceId() != targetPrincipalID {
		t.Fatalf("link employee with verified evidence: %+v, %v", linked, err)
	}
	if err := database.WithContext(ctx).Exec("DELETE FROM hr_workers WHERE id = ?", historicalWorkerID).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := client.LinkWorkerPrincipal(ctx, link); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("repeat principal link: %v", err)
	}
	if err := database.WithContext(ctx).Table("hr_worker_lifecycle_events").
		Where("worker_id = ? AND source_reference = ? AND event_type = ?", newWorkerID, insideID, "principal_linked").
		Count(&events).Error; err != nil || events != 1 {
		t.Fatalf("principal link lifecycle count = %d, %v", events, err)
	}
	if err := database.WithContext(ctx).Table("hr_workers").Where("id = ?", insideID).
		Update("updated_at", clock.now).Error; err != nil {
		t.Fatalf("synchronize second worker fixture with test clock: %v", err)
	}
	linkVerifier.expected = service.PrincipalLinkCheck{
		EvidenceID: outsideID, WorkerID: insideID, PersonPartyID: insideID,
		PrincipalID: targetPrincipalID, RequestedByPrincipalID: principalID,
	}
	if _, err := client.LinkWorkerPrincipal(ctx, &pb.LinkWorkerPrincipalRequest{
		WorkerId: insideID, PrincipalId: targetPrincipalID, EvidenceId: outsideID,
	}); status.Code(err) != codes.AlreadyExists {
		t.Fatalf("link one principal to two current workers: %v", err)
	}
	afterDuplicateLink, err := client.GetWorkerByID(ctx, &common.IDRequest{Id: insideID})
	if err != nil || afterDuplicateLink.GetItem().GetIamPrincipalReferenceId() != "" {
		t.Fatalf("duplicate link changed worker: %+v, %v", afterDuplicateLink, err)
	}
	if err := database.WithContext(ctx).Table("hr_worker_lifecycle_events").
		Where("worker_id = ? AND event_type = ?", insideID, "principal_linked").
		Count(&events).Error; err != nil || events != 0 {
		t.Fatalf("duplicate link persisted lifecycle: count=%d, err=%v", events, err)
	}
	if err := database.WithContext(ctx).Table("hr_outbox_events").
		Where("aggregate_id = ? AND event_type = ?", insideID, "hr_workers.link_principal").
		Count(&events).Error; err != nil || events != 0 {
		t.Fatalf("duplicate link persisted outbox: count=%d, err=%v", events, err)
	}
	principalID = requesterID
	if principalID == submittedTransfer.GetItem().GetRequestedByPrincipalId() {
		principalID = approverID
	}
	if _, err := client.DecideWorkerAssignmentTransfer(ctx, &pb.DecideWorkerAssignmentTransferRequest{
		TransferId: submittedTransfer.GetItem().GetId(), Decision: pb.TransferDecision_TRANSFER_DECISION_APPROVE,
	}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("approve stale transfer after primary assignment changed: %v", err)
	}
	rejected, err := client.DecideWorkerAssignmentTransfer(ctx, &pb.DecideWorkerAssignmentTransferRequest{
		TransferId: submittedTransfer.GetItem().GetId(), Decision: pb.TransferDecision_TRANSFER_DECISION_REJECT,
	})
	if err != nil || rejected.GetItem().GetStatus() != "rejected" || rejected.GetItem().GetAppliedAssignmentId() != "" {
		t.Fatalf("reject an obsolete transfer request: %+v, %v", rejected, err)
	}
	storedWorker, err = repos.Worker.SelectByID(ctx, newWorkerID)
	if err != nil || storedWorker.PrimaryAssignmentID == nil || *storedWorker.PrimaryAssignmentID != appliedID {
		t.Fatalf("rejected transfer changed worker primary assignment: %+v, %v", storedWorker, err)
	}
	if store, ok := sessions.(*authsession.Store); ok {
		if _, err := store.SetGate(ctx, false, 1, time.Now(), "hrm_integration_test"); err != nil {
			t.Fatalf("close disposable security gate: %v", err)
		}
		if _, err := client.GetAllWorkers(ctx, &common.PageRequest{Limit: 1}); status.Code(err) != codes.Unavailable {
			t.Fatalf("list workers while security gate is closed = %v, want Unavailable", err)
		}
		if _, err := store.SetGate(ctx, true, 1, time.Now(), "hrm_integration_test"); err != nil {
			t.Fatalf("reopen disposable security gate: %v", err)
		}
		recovered, err := client.GetWorkerByID(ctx, &common.IDRequest{Id: insideID})
		if err != nil || recovered.GetItem().GetId() != insideID {
			t.Fatalf("read worker after the security gate reopens: %+v, %v", recovered, err)
		}
		const recoveredEpoch = 2
		if _, err := store.SetGate(ctx, true, recoveredEpoch, time.Now(), "hrm_integration_test"); err != nil {
			t.Fatalf("advance disposable recovery epoch: %v", err)
		}
		if _, err := client.GetWorkerByID(ctx, &common.IDRequest{Id: insideID}); status.Code(err) != codes.PermissionDenied {
			t.Fatalf("worker query with previous recovery epoch = %v, want PermissionDenied", err)
		}
		now := time.Now()
		projection := workerSessionProjection{legalEntityID: legalA}
		state, err := projection.ReadSessionContext(ctx, scopedWorkerTokenHash, now)
		if err != nil {
			t.Fatal(err)
		}
		state.Snapshot.RecoveryEpoch = recoveredEpoch
		if err := store.ProjectAccessSnapshot(ctx, state.Snapshot, 0, now); err != nil {
			t.Fatalf("refresh disposable recovery projection: %v", err)
		}
		recovered, err = client.GetWorkerByID(ctx, &common.IDRequest{Id: insideID})
		if err != nil || recovered.GetItem().GetId() != insideID {
			t.Fatalf("read worker after recovery projection refresh: %+v, %v", recovered, err)
		}
		state.Snapshot.PermRevision++
		state.Snapshot.Permissions[0].Scope.ID = legalB
		now = time.Now()
		if err := store.ProjectAccessSnapshot(ctx, state.Snapshot, 0, now); err != nil {
			t.Fatalf("publish new permission revision: %v", err)
		}
		if _, err := client.GetWorkerByID(ctx, &common.IDRequest{Id: insideID}); status.Code(err) != codes.PermissionDenied {
			t.Fatalf("worker query with stale permission revision = %v, want PermissionDenied", err)
		}
		if _, err := client.GetAllWorkers(ctx, &common.PageRequest{Limit: 1}); status.Code(err) != codes.PermissionDenied {
			t.Fatalf("list workers with stale permission revision = %v, want PermissionDenied", err)
		}
		freshCtx := metadata.NewOutgoingContext(ctx, metadata.Pairs("x-ig-session-token-hash", scopedWorkerTokenHash))
		freshCtx = authcontext.AppendToOutgoingGRPCContext(freshCtx, authcontext.CallerContext{
			Authenticated: true, PrincipalID: scopedWorkerPrincipal, SessionID: scopedWorkerSession,
			UserKind: "employee", AuthEpoch: 1, PermRevision: state.Snapshot.PermRevision,
		}, gatewayKey, now)
		if _, err := client.GetWorkerByID(freshCtx, &common.IDRequest{Id: insideID}); status.Code(err) != codes.NotFound {
			t.Fatalf("worker from previous legal-entity scope = %v, want NotFound", err)
		}
		recovered, err = client.GetWorkerByID(freshCtx, &common.IDRequest{Id: outsideID})
		if err != nil || recovered.GetItem().GetId() != outsideID {
			t.Fatalf("read worker with refreshed legal-entity scope: %v", err)
		}
		page, err := client.GetAllWorkers(freshCtx, &common.PageRequest{Limit: 1})
		if err != nil || page.GetPage().GetTotal() != 1 || len(page.GetItems()) != 1 || page.GetItems()[0].GetId() != outsideID {
			t.Fatalf("scoped worker page after scope change: %v", err)
		}
		ctx = freshCtx
		t.Log("get and list denied for stale revision and restricted to the new scope with refreshed proof")
		if err := store.RevokeSessionTokens(ctx, []string{scopedWorkerTokenHash}, nil, false, time.Now()); err != nil {
			t.Fatalf("revoke session in disposable Redis: %v", err)
		}
		if _, err := client.GetWorkerByID(ctx, &common.IDRequest{Id: insideID}); status.Code(err) != codes.Unauthenticated {
			t.Fatalf("worker query after session revocation = %v, want Unauthenticated", err)
		}
		if _, err := client.GetAllWorkers(ctx, &common.PageRequest{Limit: 1}); status.Code(err) != codes.Unauthenticated {
			t.Fatalf("list workers after session revocation = %v, want Unauthenticated", err)
		}
		t.Log("get and list denied after Redis session revocation")
	}
}

func stringPointer(value string) *string { return &value }
