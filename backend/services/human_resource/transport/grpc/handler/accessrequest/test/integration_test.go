//go:build integration

package accessrequest_test

import (
	"context"
	"net"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/accessrequest/v1"
	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	hrdb "github.com/integrione/erp-retail/backend/services/human_resource/db"
	"github.com/integrione/erp-retail/backend/services/human_resource/infrastructure/security"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/accessrequest"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository"
	handler "github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/accessrequest"
	"github.com/integrione/erp-retail/backend/shared/authcontext"
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

type accessRequestPolicy struct {
	mu          sync.Mutex
	principalID string
}

func (p *accessRequestPolicy) currentPrincipal() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.principalID
}

func (p *accessRequestPolicy) Authorize(context.Context, operation.Access) (operation.Actor, error) {
	return operation.Actor{PrincipalID: p.currentPrincipal(), SessionID: "00000000-0000-4000-8000-000000000002", Kind: "human"}, nil
}

func (*accessRequestPolicy) ReadScope(context.Context, operation.Access) (operation.Scope, error) {
	return operation.Scope{All: true}, nil
}

func text(v string) *string { return &v }

func TestAccessRequestProposalLifecycleWithPostgreSQL(t *testing.T) {
	dsn := os.Getenv("HRM_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("HRM_TEST_DATABASE_URL is not configured")
	}
	parsed, err := url.Parse(dsn)
	if err != nil || parsed.Scheme != "postgres" || !strings.HasSuffix(strings.TrimPrefix(parsed.Path, "/"), "_test") {
		t.Fatal("integration test requires a dedicated *_test database")
	}
	ctx, cancel := context.WithTimeout(t.Context(), 45*time.Second)
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
	validators, err := security.AccessRequestValidators(repos)
	if err != nil {
		t.Fatal(err)
	}

	requesterID := "00000000-0000-4000-8000-0000000000a1"
	policy := &accessRequestPolicy{principalID: requesterID}
	svc, err := service.NewService(service.Dependencies{
		Runtime: operation.Dependencies{
			Transactor: repos.Transactor, Outbox: repos.OutboxEvent,
			Policy: policy, Integrity: security.NewIntegrity(validators),
		},
		AccessRequest:        repos.AccessRequest,
		WorkerLifecycleEvent: repos.WorkerLifecycleEvent,
	})
	if err != nil {
		t.Fatal(err)
	}

	listener := bufconn.Listen(1 << 20)
	server := grpc.NewServer(grpc.UnaryInterceptor(func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, next grpc.UnaryHandler) (any, error) {
		return next(authcontext.WithCaller(ctx, authcontext.CallerContext{PrincipalID: policy.currentPrincipal(), Authenticated: true}), req)
	}))
	if err := handler.Register(server, svc); err != nil {
		t.Fatal(err)
	}
	served := make(chan error, 1)
	go func() { served <- server.Serve(listener) }()
	t.Cleanup(func() { server.Stop(); _ = listener.Close(); <-served })
	conn, err := grpc.NewClient("passthrough:///hrm-accessrequest-integration",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return listener.DialContext(ctx) }),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	client := pb.NewAccessRequestServiceClient(conn)

	ids := operation.UUIDGenerator{}
	legalID, _ := ids.NewID()
	personID, _ := ids.NewID()
	workerID, _ := ids.NewID()
	unitID, _ := ids.NewID()
	catalogID, _ := ids.NewID()
	var eventID int64
	var reqID string

	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		if reqID != "" {
			if err := database.WithContext(cleanupCtx).Exec("DELETE FROM hr_access_requests WHERE id = ?", reqID).Error; err != nil {
				t.Errorf("cleanup access request: %v", err)
			}
			if err := database.WithContext(cleanupCtx).Exec("DELETE FROM hr_outbox_events WHERE aggregate_id = ?", reqID).Error; err != nil {
				t.Errorf("cleanup outbox: %v", err)
			}
		}
		if eventID > 0 {
			if err := database.WithContext(cleanupCtx).Exec("DELETE FROM hr_worker_lifecycle_events WHERE id = ?", eventID).Error; err != nil {
				t.Errorf("cleanup lifecycle event: %v", err)
			}
		}
		if err := database.WithContext(cleanupCtx).Exec("DELETE FROM hr_workers WHERE id = ?", workerID).Error; err != nil {
			t.Errorf("cleanup worker: %v", err)
		}
	})

	if err := database.WithContext(ctx).Exec(
		`INSERT INTO hr_workers (id, worker_number, person_party_reference_id, worker_category, status,
		 primary_legal_entity_reference_id, hire_date, service_start_date)
		 VALUES (?, ?, ?, 'employee', 'active', ?, '2026-01-01', '2026-01-01')`,
		workerID, "AR-W-"+workerID, personID, legalID,
	).Error; err != nil {
		t.Fatal(err)
	}
	if err := database.WithContext(ctx).Raw(
		`INSERT INTO hr_worker_lifecycle_events (worker_id, event_type, occurred_at, effective_date,
		 previous_status, new_status, actor_principal_id, idempotency_key)
		 VALUES (?, 'hire', NOW(), '2026-01-01', 'preboarding', 'active', ?, ?) RETURNING id`,
		workerID, requesterID, "worker-activate:"+workerID,
	).Scan(&eventID).Error; err != nil || eventID <= 0 {
		t.Fatalf("insert lifecycle event: id=%d err=%v", eventID, err)
	}

	catalog2ID, _ := ids.NewID()
	if _, err := client.CreateAccessRequest(ctx, &pb.CreateAccessRequestRequest{Input: &pb.AccessRequestInput{
		WorkerId: &workerID, LifecycleEventId: &eventID, RequestKind: text("provision"),
		RequestedScopeType: text("organization_unit"), RequestedScopeReferenceId: &unitID,
		ApprovedCatalogReferenceId: &catalogID,
		Status:                     text("failed"), IdempotencyKey: text("idem-ar-bad-" + workerID),
	}}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("direct failed access request = %v, want FailedPrecondition", status.Code(err))
	}

	arRes, err := client.CreateAccessRequest(ctx, &pb.CreateAccessRequestRequest{Input: &pb.AccessRequestInput{
		WorkerId: &workerID, RequestKind: text("provision"),
		RequestedScopeType: text("organization_unit"), RequestedScopeReferenceId: &unitID,
		ApprovedCatalogReferenceId: &catalogID,
		Status:                     text("pending"), IdempotencyKey: text("idem-ar-1-" + workerID),
	}})
	if err != nil || arRes.GetValue().GetStatus() != "pending" {
		t.Fatalf("create access request: res=%+v err=%v", arRes, err)
	}
	reqID = arRes.GetValue().GetId()
	if arRes.GetValue().GetLifecycleEventId() != eventID {
		t.Fatalf("generated lifecycle event reference = %d, want %d", arRes.GetValue().GetLifecycleEventId(), eventID)
	}

	if _, err := client.UpdateAccessRequest(ctx, &pb.UpdateAccessRequestRequest{
		Id: reqID, Input: &pb.AccessRequestInput{ApprovedCatalogReferenceId: &catalog2ID},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"approved_catalog_reference_id"}},
	}); err != nil {
		t.Fatalf("update pending access request: %v", err)
	}

	if _, err := client.UpdateAccessRequest(ctx, &pb.UpdateAccessRequestRequest{
		Id: reqID, Input: &pb.AccessRequestInput{IdempotencyKey: text("replacement-" + reqID)},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"idempotency_key"}},
	}); status.Code(err) != codes.FailedPrecondition {
		t.Errorf("change access request idempotency key = %v, want FailedPrecondition", err)
	}

	found, err := client.FindAccessRequestByWorkerID(ctx, &common.StringFinderRequest{Value: &workerID, Limit: 10})
	if err != nil || found.GetPage().GetTotal() != 1 || found.GetItems()[0].GetApprovedCatalogReferenceId() != catalog2ID {
		t.Fatalf("find access requests: res=%+v err=%v", found, err)
	}

	// Fixture mewakili tanda terima IAM yang sudah tiba sebelum status lokal tersinkron.
	if err := database.WithContext(ctx).Exec("UPDATE hr_access_requests SET iam_request_reference_id = ? WHERE id = ?",
		catalog2ID, reqID).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := client.DeleteAccessRequest(ctx, &common.IDRequest{Id: reqID}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("delete access request with IAM receipt = %v, want FailedPrecondition", err)
	}
	if err := database.WithContext(ctx).Exec("UPDATE hr_access_requests SET iam_request_reference_id = NULL WHERE id = ?",
		reqID).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := client.DeleteAccessRequest(ctx, &common.IDRequest{Id: reqID}); err != nil {
		t.Fatalf("delete pending access request: %v", err)
	}
	reqID = ""
}
