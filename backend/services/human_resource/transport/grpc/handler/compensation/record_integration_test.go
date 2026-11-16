//go:build integration

package compensation_test

import (
	"context"
	"net"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	compensationpb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/compensation/v1"
	payrollpb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/payroll/v1"
	hrdb "github.com/integrione/erp-retail/backend/services/human_resource/db"
	"github.com/integrione/erp-retail/backend/services/human_resource/infrastructure/security"
	compensationservice "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/compensation"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	payrollservice "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/payroll"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository"
	compensationhandler "github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/compensation"
	payrollhandler "github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/payroll"
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

func TestCompensationProposalRPCWithPostgreSQL(t *testing.T) {
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
	t.Cleanup(func() { _ = pool.Close() })
	if err := hrdb.RunMigrationsUp(ctx, pool); err != nil {
		t.Fatal(err)
	}
	repos, err := repository.New(database)
	if err != nil {
		t.Fatal(err)
	}
	validators, err := security.CompensationValidators(repos)
	if err != nil {
		t.Fatal(err)
	}
	payValidators, err := security.PayComponentValidators(repos)
	if err != nil {
		t.Fatal(err)
	}
	for name, validator := range payValidators {
		validators[name] = validator
	}
	runtime := operation.Dependencies{Transactor: repos.Transactor, Outbox: repos.OutboxEvent,
		Policy: compensationIntegrationPolicy{}, Integrity: security.NewIntegrity(validators)}
	compensationService, err := compensationservice.NewService(compensationservice.Dependencies{
		Runtime: runtime, CompensationBand: repos.CompensationBand, CompensationRecord: repos.CompensationRecord,
	})
	if err != nil {
		t.Fatal(err)
	}
	payrollService, err := payrollservice.NewService(payrollservice.Dependencies{
		Runtime: runtime, PayrollProfile: repos.PayrollProfile, PayComponent: repos.PayComponent,
		PayrollPeriod: repos.PayrollPeriod, PayrollRun: repos.PayrollRun, PayrollResult: repos.PayrollResult,
		PayrollResultLine: repos.PayrollResultLine, PayrollAdjustment: repos.PayrollAdjustment,
		PayrollGroup: repos.PayrollGroup, PayrollRuleSet: repos.PayrollRuleSet,
	})
	if err != nil {
		t.Fatal(err)
	}
	listener := bufconn.Listen(1 << 20)
	server := grpc.NewServer()
	if err := compensationhandler.Register(server, compensationService); err != nil {
		t.Fatal(err)
	}
	if err := payrollhandler.Register(server, payrollService); err != nil {
		t.Fatal(err)
	}
	served := make(chan error, 1)
	go func() { served <- server.Serve(listener) }()
	t.Cleanup(func() {
		server.Stop()
		if err := <-served; err != nil {
			t.Errorf("serve HRM compensation gRPC: %v", err)
		}
	})
	connection, err := grpc.NewClient("passthrough:///hrm-compensation-proposal-integration",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return listener.DialContext(ctx) }),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = connection.Close() })
	compensationClient := compensationpb.NewCompensationServiceClient(connection)
	payrollClient := payrollpb.NewPayrollServiceClient(connection)
	ids := operation.UUIDGenerator{}
	newID := func() string {
		id, err := ids.NewID()
		if err != nil {
			t.Fatal(err)
		}
		return id
	}
	workerID, familyID, jobID, positionID, assignmentID := newID(), newID(), newID(), newID(), newID()
	code := strings.ReplaceAll(workerID, "-", "")
	legalID, organizationID, personID := newID(), newID(), newID()
	from, to := "2026-01-01", "2026-12-31"
	fixtures := []struct {
		query string
		args  []any
	}{
		{"INSERT INTO hr_job_families (id, family_code, family_name, job_domain) VALUES (?, ?, 'Store', 'operations')", []any{familyID, "STORE-" + code}},
		{"INSERT INTO hr_jobs (id, job_family_id, job_code, job_title, job_level_code, valid_from) VALUES (?, ?, ?, 'Supervisor', 'L2', ?)", []any{jobID, familyID, "JOB-" + code, from}},
		{"INSERT INTO hr_positions (id, position_code, position_title, job_id, organization_unit_reference_id, legal_entity_reference_id, valid_from) VALUES (?, ?, 'Supervisor', ?, ?, ?, ?)", []any{positionID, "POS-" + code, jobID, organizationID, legalID, from}},
		{"INSERT INTO hr_workers (id, worker_number, person_party_reference_id, worker_category, status, hire_date, primary_legal_entity_reference_id) VALUES (?, ?, ?, 'employee', 'active', ?, ?)", []any{workerID, "WORKER-" + code, personID, from, legalID}},
		{"INSERT INTO hr_worker_assignments (id, worker_id, position_id, assignment_type, valid_from, valid_to) VALUES (?, ?, ?, 'primary', ?, ?)", []any{assignmentID, workerID, positionID, from, to}},
	}
	for _, fixture := range fixtures {
		if err := database.WithContext(ctx).Exec(fixture.query, fixture.args...).Error; err != nil {
			t.Fatalf("insert HRM fixture: %v", err)
		}
	}
	componentCode, name, componentType, method, jurisdiction, currency := "BASE-"+code, "Base salary", "earning", "fixed", "ID", "IDR"
	version := int32(1)
	componentInput := &payrollpb.PayComponentInput{ComponentCode: &componentCode, ComponentVersion: &version,
		ComponentName: &name, ComponentType: &componentType, CalculationMethod: &method,
		JurisdictionCode: &jurisdiction, CurrencyCode: &currency, EffectiveFrom: &from}
	formulaID := newID()
	componentInput.FormulaReferenceId = &formulaID
	if _, err := payrollClient.CreatePayComponent(ctx, &payrollpb.CreatePayComponentRequest{Input: componentInput}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("unverified formula status = %v, want FailedPrecondition", status.Code(err))
	}
	componentInput.FormulaReferenceId = nil
	component, err := payrollClient.CreatePayComponent(ctx, &payrollpb.CreatePayComponentRequest{Input: componentInput})
	if err != nil || component.GetItem().GetId() == "" {
		t.Fatalf("create component: %+v, %v", component, err)
	}
	componentID := component.GetItem().GetId()
	if _, err := payrollClient.CreatePayComponent(ctx, &payrollpb.CreatePayComponentRequest{Input: componentInput}); status.Code(err) != codes.AlreadyExists {
		t.Fatalf("duplicate component status = %v, want AlreadyExists", status.Code(err))
	}
	bandCode, level, frequency := "BAND-"+code, "L2", "monthly"
	minimum, midpoint, maximum := "5000000", "6000000", "7000000"
	bandInput := &compensationpb.CompensationBandInput{BandCode: &bandCode, BandVersion: &version,
		JobLevelCode: &level, JurisdictionCode: &jurisdiction, CurrencyCode: &currency,
		PayFrequency: &frequency, MinimumAmount: &minimum, MidpointAmount: &midpoint,
		MaximumAmount: &maximum, ValidFrom: &from}
	band, err := compensationClient.CreateCompensationBand(ctx, &compensationpb.CreateCompensationBandRequest{Input: bandInput})
	if err != nil || band.GetItem().GetId() == "" {
		t.Fatalf("create band: %+v, %v", band, err)
	}
	bandID := band.GetItem().GetId()
	compensationType, amount := "base_salary", "6000000"
	recordInput := &compensationpb.CompensationRecordInput{WorkerId: &workerID, AssignmentId: &assignmentID,
		ComponentId: &componentID, ComponentCodeSnapshot: &componentCode, CompensationBandId: &bandID,
		CompensationType: &compensationType, Amount: &amount, CurrencyCode: &currency,
		PayFrequency: &frequency, EffectiveFrom: &from, EffectiveTo: &to}
	approved := "approved"
	recordInput.Status = &approved
	approverID := newID()
	recordInput.ApprovedByPrincipalId = &approverID
	recordInput.ApprovedAt = timestamppb.New(time.Now().UTC())
	if _, err := compensationClient.CreateCompensationRecord(ctx, &compensationpb.CreateCompensationRecordRequest{Input: recordInput}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("direct approval status = %v, want FailedPrecondition", status.Code(err))
	}
	recordInput.Status = nil
	recordInput.ApprovedByPrincipalId = nil
	recordInput.ApprovedAt = nil
	wrongAmount := "8000000"
	recordInput.Amount = &wrongAmount
	if _, err := compensationClient.CreateCompensationRecord(ctx, &compensationpb.CreateCompensationRecordRequest{Input: recordInput}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("amount outside band status = %v, want FailedPrecondition", status.Code(err))
	}
	recordInput.Amount = &amount
	record, err := compensationClient.CreateCompensationRecord(ctx, &compensationpb.CreateCompensationRecordRequest{Input: recordInput})
	if err != nil || record.GetItem().GetId() == "" || record.GetItem().GetStatus() != "proposed" {
		t.Fatalf("create proposal: %+v, %v", record, err)
	}
	recordID := record.GetItem().GetId()
	if page, err := compensationClient.FindCompensationRecordByWorkerID(ctx, &common.StringFinderRequest{Value: &workerID, Limit: 1}); err != nil ||
		page.GetPage().GetTotal() != 1 || len(page.GetItems()) != 1 {
		t.Fatalf("find proposal: %+v, %v", page, err)
	}
	if page, err := compensationClient.FindCompensationRecordByWorkerID(ctx, &common.StringFinderRequest{Value: &workerID, Limit: 1, Offset: 1}); err != nil ||
		page.GetPage().GetTotal() != 1 || len(page.GetItems()) != 0 {
		t.Fatalf("proposal offset: %+v, %v", page, err)
	}
	changedName := "Salary revised"
	if _, err := payrollClient.UpdatePayComponent(ctx, &payrollpb.UpdatePayComponentRequest{Id: componentID,
		Input: &payrollpb.PayComponentInput{ComponentName: &changedName}, UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"component_name"}}}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("referenced component update status = %v, want FailedPrecondition", status.Code(err))
	}
	changedMaximum := "8000000"
	if _, err := compensationClient.UpdateCompensationBand(ctx, &compensationpb.UpdateCompensationBandRequest{Id: bandID,
		Input: &compensationpb.CompensationBandInput{MaximumAmount: &changedMaximum}, UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"maximum_amount"}}}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("referenced band update status = %v, want FailedPrecondition", status.Code(err))
	}
	changedAmount := "6500000"
	if updated, err := compensationClient.UpdateCompensationRecord(ctx, &compensationpb.UpdateCompensationRecordRequest{Id: recordID,
		Input: &compensationpb.CompensationRecordInput{Amount: &changedAmount}, UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"amount"}}}); err != nil ||
		updated.GetItem().GetAmount() != changedAmount {
		t.Fatalf("update proposal: %+v, %v", updated, err)
	}
	if _, err := compensationClient.DeleteCompensationRecord(ctx, &common.IDRequest{Id: recordID}); err != nil {
		t.Fatalf("delete proposal: %v", err)
	}
	if _, err := payrollClient.DeletePayComponent(ctx, &common.IDRequest{Id: componentID}); err != nil {
		t.Fatalf("delete unused component: %v", err)
	}
	if _, err := compensationClient.DeleteCompensationBand(ctx, &common.IDRequest{Id: bandID}); err != nil {
		t.Fatalf("delete unused band: %v", err)
	}
	for _, aggregateID := range []string{componentID, bandID, recordID} {
		var count int64
		if err := database.WithContext(ctx).Table("hr_outbox_events").Where("aggregate_id = ?", aggregateID).Count(&count).Error; err != nil ||
			(count != 2 && count != 3) {
			t.Fatalf("outbox for %s = %d, %v", aggregateID, count, err)
		}
	}
}
