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
	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/workforceplanning/v1"
	hrdb "github.com/integrione/erp-retail/backend/services/human_resource/db"
	"github.com/integrione/erp-retail/backend/services/human_resource/infrastructure/security"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/workforceplanning"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository"
	handler "github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/workforceplanning"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type positionPolicy struct{}

func (positionPolicy) Authorize(context.Context, operation.Access) (operation.Actor, error) {
	return operation.Actor{
		PrincipalID: "00000000-0000-4000-8000-000000000001",
		SessionID:   "00000000-0000-4000-8000-000000000002", Kind: "human",
	}, nil
}

func (positionPolicy) ReadScope(context.Context, operation.Access) (operation.Scope, error) {
	return operation.Scope{All: true}, nil
}

type foundationPositionStub struct {
	checked security.PositionReferenceCheck
	err     error
}

func (s *foundationPositionStub) VerifyPositionReferences(_ context.Context, check security.PositionReferenceCheck) error {
	s.checked = check
	return s.err
}

func TestPlannedPositionRPCWithPostgreSQL(t *testing.T) {
	dsn := os.Getenv("HRM_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("HRM_TEST_DATABASE_URL is not configured")
	}
	parsed, err := url.Parse(dsn)
	if err != nil {
		t.Fatal("integration test requires a valid disposable HRM URL")
	}
	dbName := strings.TrimPrefix(parsed.Path, "/")
	if parsed.Scheme != "postgres" ||
		!strings.HasPrefix(dbName, "hrm_disposable_") || !strings.HasSuffix(dbName, "_test") ||
		os.Getenv("HRM_TEST_DATABASE_DISPOSABLE") != "1" {
		t.Fatal("integration test requires a disposable HRM test database")
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
	ids := operation.UUIDGenerator{}
	familyID, err := ids.NewID()
	if err != nil {
		t.Fatal(err)
	}
	jobID, err := ids.NewID()
	if err != nil {
		t.Fatal(err)
	}
	legalID, err := ids.NewID()
	if err != nil {
		t.Fatal(err)
	}
	unitID, err := ids.NewID()
	if err != nil {
		t.Fatal(err)
	}
	if err := database.WithContext(ctx).Exec(
		"INSERT INTO hr_job_families (id, family_code, family_name, job_domain) VALUES (?, ?, ?, ?)",
		familyID, "POSITION-"+familyID, "Retail", "operations",
	).Error; err != nil {
		t.Fatal(err)
	}
	if err := database.WithContext(ctx).Exec(
		"INSERT INTO hr_jobs (id, job_family_id, job_code, job_title, valid_from) VALUES (?, ?, ?, ?, ?)",
		jobID, familyID, "POSITION-"+jobID, "Store associate", "2026-01-01",
	).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		if err := database.WithContext(cleanupCtx).Exec("DELETE FROM hr_jobs WHERE id = ?", jobID).Error; err != nil {
			t.Errorf("clean HRM job fixture: %v", err)
		}
		if err := database.WithContext(cleanupCtx).Exec("DELETE FROM hr_job_families WHERE id = ?", familyID).Error; err != nil {
			t.Errorf("clean HRM family fixture: %v", err)
		}
	})
	owner := &foundationPositionStub{}
	validators, err := security.PositionValidators(repos, owner)
	if err != nil {
		t.Fatal(err)
	}
	svc, err := service.NewService(service.Dependencies{
		Runtime: operation.Dependencies{
			Transactor: repos.Transactor, Outbox: repos.OutboxEvent,
			Policy: positionPolicy{}, Integrity: security.NewIntegrity(validators),
		},
		JobFamily: repos.JobFamily, Job: repos.Job, Position: repos.Position,
		WorkforcePlan: repos.WorkforcePlan, WorkforcePlanLine: repos.WorkforcePlanLine,
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
	connection, err := grpc.NewClient("passthrough:///hrm-position-integration",
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
	client := pb.NewWorkforcePlanningServiceClient(connection)
	code := "SEAT-" + jobID
	input := &pb.PositionInput{
		PositionCode: &code, PositionTitle: stringPointer("Store associate"), JobId: &jobID,
		OrganizationUnitReferenceId: &unitID, LegalEntityReferenceId: &legalID,
		ValidFrom: stringPointer("2026-11-17"),
	}
	active := proto.CloneOf(input)
	active.PositionStatus = stringPointer("active")
	if _, err := client.CreatePosition(ctx, &pb.CreatePositionRequest{Input: active}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("active position through CRUD = %v", err)
	}
	created, err := client.CreatePosition(ctx, &pb.CreatePositionRequest{Input: input})
	if err != nil || created.GetItem().GetPositionStatus() != "planned" {
		t.Fatalf("create planned position = %+v, %v", created, err)
	}
	positionID := created.GetItem().GetId()
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		if err := database.WithContext(cleanupCtx).Exec("DELETE FROM hr_positions WHERE id = ?", positionID).Error; err != nil {
			t.Errorf("clean HRM position fixture: %v", err)
		}
	})
	if owner.checked.OrganizationUnitID != unitID || owner.checked.LegalEntityID != legalID {
		t.Fatalf("Foundation verifier received wrong scope: %+v", owner.checked)
	}
	if _, err := client.CreatePosition(ctx, &pb.CreatePositionRequest{Input: input}); status.Code(err) != codes.AlreadyExists {
		t.Fatalf("duplicate planned position = %v", err)
	}
	updated, err := client.UpdatePosition(ctx, &pb.UpdatePositionRequest{
		Id: positionID, Input: &pb.PositionInput{PositionTitle: stringPointer("Senior store associate")},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"position_title"}},
	})
	if err != nil || updated.GetItem().GetPositionTitle() != "Senior store associate" {
		t.Fatalf("update planned position = %+v, %v", updated, err)
	}
	page, err := client.GetAllPositions(ctx, &common.PageRequest{Limit: 1})
	if err != nil || page.GetPage().GetTotal() != 1 || len(page.GetItems()) != 1 {
		t.Fatalf("position page = %+v, %v", page, err)
	}
	if _, err := client.DeletePosition(ctx, &common.IDRequest{Id: positionID}); err != nil {
		t.Fatalf("delete planned position: %v", err)
	}
	var events int64
	if err := database.WithContext(ctx).Table("hr_outbox_events").Where("aggregate_id = ?", positionID).Count(&events).Error; err != nil || events != 3 {
		t.Fatalf("position outbox events = %d, %v", events, err)
	}
	activeCode := "ACTIVE-SEAT-" + jobID
	activeInput := proto.CloneOf(input)
	activeInput.PositionCode = &activeCode
	planned, err := client.CreatePosition(ctx, &pb.CreatePositionRequest{Input: activeInput})
	if err != nil {
		t.Fatalf("create position for activation: %v", err)
	}
	activeID := planned.GetItem().GetId()
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		if err := database.WithContext(cleanupCtx).Exec("DELETE FROM hr_positions WHERE id = ?", activeID).Error; err != nil {
			t.Errorf("clean active position fixture: %v", err)
		}
	})
	if _, err := client.UpdatePosition(ctx, &pb.UpdatePositionRequest{Id: activeID,
		Input:      &pb.PositionInput{PositionStatus: stringPointer("active")},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"position_status"}}}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("activation through generic update = %v", err)
	}
	owner.err = context.DeadlineExceeded
	if _, err := client.ActivatePosition(ctx, &common.IDRequest{Id: activeID}); status.Code(err) != codes.Unavailable {
		t.Fatalf("activation with owner unavailable = %v", err)
	}
	owner.err = nil
	activated, err := client.ActivatePosition(ctx, &common.IDRequest{Id: activeID})
	if err != nil || activated.GetItem().GetPositionStatus() != "active" {
		t.Fatalf("activate position = %+v, %v", activated, err)
	}
	if _, err := client.ActivatePosition(ctx, &common.IDRequest{Id: activeID}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("repeat activation = %v", err)
	}
	if _, err := client.DeletePosition(ctx, &common.IDRequest{Id: activeID}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("delete active position = %v", err)
	}
	if err := database.WithContext(ctx).Table("hr_outbox_events").Where("aggregate_id = ?", activeID).Count(&events).Error; err != nil || events != 2 {
		t.Fatalf("active position outbox events = %d, %v", events, err)
	}
}

func stringPointer(value string) *string { return &value }
