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
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type workerScopePolicy struct{ legalEntityID string }

func (workerScopePolicy) Authorize(context.Context, operation.Access) (operation.Actor, error) {
	return operation.Actor{
		PrincipalID: "00000000-0000-4000-8000-000000000001",
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
	svc, err := service.NewService(service.Dependencies{
		Runtime: operation.Dependencies{
			Transactor: repos.Transactor, Outbox: repos.OutboxEvent,
			Policy: workerScopePolicy{legalEntityID: legalA}, Integrity: security.NewIntegrity(validators),
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
}

func stringPointer(value string) *string { return &value }
