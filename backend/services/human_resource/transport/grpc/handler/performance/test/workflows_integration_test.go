//go:build integration

package test

import (
	"bytes"
	"context"
	"net"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/performance/v1"
	hrdb "github.com/integrione/erp-retail/backend/services/human_resource/db"
	"github.com/integrione/erp-retail/backend/services/human_resource/infrastructure/security"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/performance"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository"
	handler "github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/performance"
	"github.com/integrione/erp-retail/backend/shared/authcontext"
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

type perfWfPolicy struct {
	mu          sync.Mutex
	principalID string
}

func (p *perfWfPolicy) setPrincipal(id string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.principalID = id
}

func (p *perfWfPolicy) currentPrincipal() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.principalID
}

func (p *perfWfPolicy) Authorize(context.Context, operation.Access) (operation.Actor, error) {
	return operation.Actor{PrincipalID: p.currentPrincipal(), SessionID: "00000000-0000-4000-8000-000000000002", Kind: "human"}, nil
}

func (*perfWfPolicy) ReadScope(context.Context, operation.Access) (operation.Scope, error) {
	return operation.Scope{All: true}, nil
}

func strPtr(v string) *string { return &v }

func TestPerformanceWorkflowsAndDetailsWithPostgreSQL(t *testing.T) {
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
	validators, err := security.PerformanceValidators(repos)
	if err != nil {
		t.Fatal(err)
	}

	workerPrincipalID := "00000000-0000-4000-8000-000000000041"
	managerPrincipalID := "00000000-0000-4000-8000-000000000042"
	policy := &perfWfPolicy{principalID: managerPrincipalID}

	svc, err := service.NewService(service.Dependencies{
		Runtime: operation.Dependencies{
			Transactor: repos.Transactor, Outbox: repos.OutboxEvent,
			Policy: policy, Integrity: security.NewIntegrity(validators),
		},
		Worker:                     repos.Worker,
		PerformanceCycle:           repos.PerformanceCycle,
		PerformanceGoal:            repos.PerformanceGoal,
		PerformanceReview:          repos.PerformanceReview,
		PerformanceGoalUpdate:      repos.PerformanceGoalUpdate,
		PerformanceReviewGoalScore: repos.PerformanceReviewGoalScore,
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
	conn, err := grpc.NewClient("passthrough:///hrm-performance-workflows-integration",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return listener.DialContext(ctx) }),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	client := pb.NewPerformanceServiceClient(conn)

	ids := operation.UUIDGenerator{}
	legalID, _ := ids.NewID()
	workerID, _ := ids.NewID()
	mgrWorkerID, _ := ids.NewID()
	person1ID, _ := ids.NewID()
	person2ID, _ := ids.NewID()
	doc1ID, _ := ids.NewID()
	doc2ID, _ := ids.NewID()
	var cycleID, goalID, goalRejectID, update1ID, update2ID, reviewID, scoreID string

	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		err := database.WithContext(cleanupCtx).Transaction(func(tx *gorm.DB) error {
			if err := tx.Exec("ALTER TABLE hr_performance_goal_updates DISABLE TRIGGER tr_hr_goal_update_immutable").Error; err != nil {
				return err
			}
			for _, stmt := range []struct {
				q string
				v string
			}{
				{"DELETE FROM hr_performance_review_goal_scores WHERE id = ?", scoreID},
				{"DELETE FROM hr_performance_reviews WHERE id = ?", reviewID},
				{"DELETE FROM hr_performance_goal_updates WHERE id = ?", update2ID},
				{"DELETE FROM hr_performance_goal_updates WHERE id = ?", update1ID},
				{"DELETE FROM hr_performance_goals WHERE id = ?", goalRejectID},
				{"DELETE FROM hr_performance_goals WHERE id = ?", goalID},
				{"DELETE FROM hr_performance_cycles WHERE id = ?", cycleID},
				{"DELETE FROM hr_workers WHERE id = ?", workerID},
				{"DELETE FROM hr_workers WHERE id = ?", mgrWorkerID},
			} {
				if stmt.v != "" {
					if err := tx.Exec(stmt.q, stmt.v).Error; err != nil {
						return err
					}
				}
			}
			for _, agg := range []string{scoreID, reviewID, update2ID, update1ID, goalRejectID, goalID, cycleID} {
				if agg != "" {
					if err := tx.Exec("DELETE FROM hr_outbox_events WHERE aggregate_id = ?", agg).Error; err != nil {
						return err
					}
				}
			}
			return tx.Exec("ALTER TABLE hr_performance_goal_updates ENABLE TRIGGER tr_hr_goal_update_immutable").Error
		})
		if err != nil {
			t.Errorf("performance workflows cleanup failed: %v", err)
		}
	})

	for _, w := range []struct {
		id, person, principal, code string
	}{
		{workerID, person1ID, workerPrincipalID, "PFW"},
		{mgrWorkerID, person2ID, managerPrincipalID, "PFM"},
	} {
		if err := database.WithContext(ctx).Exec(
			`INSERT INTO hr_workers (id, worker_number, person_party_reference_id, worker_category, status,
			 primary_legal_entity_reference_id, iam_principal_reference_id, hire_date, service_start_date)
			 VALUES (?, ?, ?, 'employee', 'active', ?, ?, '2026-01-01', '2026-01-01')`,
			w.id, w.code+"-"+w.id, w.person, legalID, w.principal,
		).Error; err != nil {
			t.Fatal(err)
		}
	}

	// Buat dan luncurkan siklus performa.
	cycleRes, err := client.CreatePerformanceCycle(ctx, &pb.CreatePerformanceCycleRequest{Input: &pb.PerformanceCycleInput{
		CycleCode: strPtr("PERF-WF-" + cycleID + workerID[:8]), CycleName: strPtr("FY2026 Review"),
		CycleType: strPtr("annual"), PeriodStart: strPtr("2026-01-01"), PeriodEnd: strPtr("2026-12-31"),
		Configuration: &common.JsonDocument{Utf8: []byte(`{"scale":"1-5"}`)},
	}})
	if err != nil {
		t.Fatalf("create cycle: %v", err)
	}
	cycleID = cycleRes.GetItem().GetId()
	if launched, err := client.LaunchPerformanceCycle(ctx, &common.IDRequest{Id: cycleID}); err != nil || launched.GetItem().GetStatus() != "open" {
		t.Fatalf("launch cycle: res=%+v err=%v", launched, err)
	}

	// Buat dua goal: satu disetujui (bobot 100.000), satu ditolak.
	policy.setPrincipal(workerPrincipalID)
	goalRes, err := client.CreatePerformanceGoal(ctx, &pb.CreatePerformanceGoalRequest{Input: &pb.PerformanceGoalInput{
		WorkerId: &workerID, PerformanceCycleId: &cycleID, GoalCode: strPtr("KPI-1-" + workerID[:8]),
		GoalTitle: strPtr("Service Reliability"), GoalType: strPtr("kpi"),
		MeasureDefinition: &common.JsonDocument{Utf8: []byte(`{"metric":"slo"}`)},
		TargetValue:       strPtr("99.9000"), UnitCode: strPtr("percent"),
		WeightPercent: strPtr("100.000"), DueDate: strPtr("2026-12-15"),
	}})
	if err != nil {
		t.Fatalf("create goal 1: %v", err)
	}
	goalID = goalRes.GetItem().GetId()
	if _, err := client.SubmitPerformanceGoal(ctx, &common.IDRequest{Id: goalID}); err != nil {
		t.Fatalf("submit goal 1: %v", err)
	}
	if _, err := client.ApprovePerformanceGoal(ctx, &common.IDRequest{Id: goalID}); status.Code(err) != codes.PermissionDenied {
		t.Fatalf("self approve goal = %v, want PermissionDenied", status.Code(err))
	}

	policy.setPrincipal(managerPrincipalID)
	if appGoal, err := client.ApprovePerformanceGoal(ctx, &common.IDRequest{Id: goalID}); err != nil || appGoal.GetItem().GetGoalStatus() != "approved" {
		t.Fatalf("approve goal 1: res=%+v err=%v", appGoal, err)
	}

	policy.setPrincipal(workerPrincipalID)
	goal2Res, err := client.CreatePerformanceGoal(ctx, &pb.CreatePerformanceGoalRequest{Input: &pb.PerformanceGoalInput{
		WorkerId: &workerID, PerformanceCycleId: &cycleID, GoalCode: strPtr("KPI-2-" + workerID[:8]),
		GoalTitle: strPtr("Rejected Draft"), GoalType: strPtr("kpi"),
		TargetValue: strPtr("10.0000"), WeightPercent: strPtr("40.000"), DueDate: strPtr("2026-12-15"),
	}})
	if err != nil {
		t.Fatalf("create goal 2: %v", err)
	}
	goalRejectID = goal2Res.GetItem().GetId()
	if _, err := client.SubmitPerformanceGoal(ctx, &common.IDRequest{Id: goalRejectID}); err != nil {
		t.Fatalf("submit goal 2: %v", err)
	}
	policy.setPrincipal(managerPrincipalID)
	if rejGoal, err := client.RejectPerformanceGoal(ctx, &common.IDRequest{Id: goalRejectID}); err != nil || rejGoal.GetItem().GetGoalStatus() != "rejected" {
		t.Fatalf("reject goal 2: res=%+v err=%v", rejGoal, err)
	}

	// Catat pengukuran append-only + superseding update + tolak hapus immutable.
	policy.setPrincipal(workerPrincipalID)
	up1, err := client.CreatePerformanceGoalUpdate(ctx, &pb.CreatePerformanceGoalUpdateRequest{Input: &pb.PerformanceGoalUpdateInput{
		GoalId: &goalID, ObservedAt: timestamppb.Now(), ActualValue: strPtr("99.5000"),
		EvidenceDocumentVersionReferenceId: &doc1ID, EvidenceDigest: bytes.Repeat([]byte{0x61}, 32),
		IdempotencyKey: strPtr("idem-up1-" + goalID),
	}})
	if err != nil {
		t.Fatalf("create goal update 1: %v", err)
	}
	update1ID = up1.GetItem().GetId()

	up2, err := client.CreatePerformanceGoalUpdate(ctx, &pb.CreatePerformanceGoalUpdateRequest{Input: &pb.PerformanceGoalUpdateInput{
		GoalId: &goalID, ObservedAt: timestamppb.Now(), ActualValue: strPtr("99.9500"),
		EvidenceDocumentVersionReferenceId: &doc2ID, EvidenceDigest: bytes.Repeat([]byte{0x62}, 32),
		SupersedesUpdateId: &update1ID, IdempotencyKey: strPtr("idem-up2-" + goalID),
	}})
	if err != nil {
		t.Fatalf("create superseding goal update: %v", err)
	}
	update2ID = up2.GetItem().GetId()
	if _, err := client.DeletePerformanceGoalUpdate(ctx, &common.IDRequest{Id: update1ID}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("delete goal update = %v, want FailedPrecondition", status.Code(err))
	}

	// Buat self review, beri skor goal, submit, lalu finalisasi oleh manajer dan verifikasi skor membeku.
	revVer := int32(1)
	revRes, err := client.CreatePerformanceReview(ctx, &pb.CreatePerformanceReviewRequest{Input: &pb.PerformanceReviewInput{
		WorkerId: &workerID, PerformanceCycleId: &cycleID,
		ReviewerRole: strPtr("self"), ReviewVersion: &revVer, VisibilityClass: strPtr("restricted"),
	}})
	if err != nil {
		t.Fatalf("create self review: %v", err)
	}
	reviewID = revRes.GetItem().GetId()

	rating := int32(4)
	if _, err := client.CreatePerformanceReviewGoalScore(ctx, &pb.CreatePerformanceReviewGoalScoreRequest{Input: &pb.PerformanceReviewGoalScoreInput{
		PerformanceReviewId: &reviewID, GoalId: &goalID, WorkerId: &workerID, PerformanceCycleId: &cycleID,
		Rating: &rating, WeightPercentSnapshot: strPtr("50.000"),
		RubricDocumentVersionReferenceId: &doc1ID, EvidenceDocumentVersionReferenceId: &doc2ID,
	}}); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("mismatched weight snapshot = %v, want InvalidArgument", status.Code(err))
	}

	scoreRes, err := client.CreatePerformanceReviewGoalScore(ctx, &pb.CreatePerformanceReviewGoalScoreRequest{Input: &pb.PerformanceReviewGoalScoreInput{
		PerformanceReviewId: &reviewID, GoalId: &goalID, WorkerId: &workerID, PerformanceCycleId: &cycleID,
		Rating: &rating, WeightPercentSnapshot: strPtr("100.000"),
		RubricDocumentVersionReferenceId: &doc1ID, EvidenceDocumentVersionReferenceId: &doc2ID,
	}})
	if err != nil {
		t.Fatalf("create review goal score: %v", err)
	}
	scoreID = scoreRes.GetItem().GetId()

	if subRev, err := client.SubmitPerformanceReview(ctx, &common.IDRequest{Id: reviewID}); err != nil || subRev.GetItem().GetSubmittedAt() == nil {
		t.Fatalf("submit review: res=%+v err=%v", subRev, err)
	}

	policy.setPrincipal(managerPrincipalID)
	if finRev, err := client.FinalizePerformanceReview(ctx, &common.IDRequest{Id: reviewID}); err != nil || finRev.GetItem().GetFinalizedAt() == nil {
		t.Fatalf("finalize review: res=%+v err=%v", finRev, err)
	}

	newRating := int32(5)
	if _, err := client.UpdatePerformanceReviewGoalScore(ctx, &pb.UpdatePerformanceReviewGoalScoreRequest{
		Id: scoreID, Input: &pb.PerformanceReviewGoalScoreInput{Rating: &newRating},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"rating"}},
	}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("update frozen goal score = %v, want FailedPrecondition", status.Code(err))
	}

	if closedCycle, err := client.ClosePerformanceCycle(ctx, &common.IDRequest{Id: cycleID}); err != nil || closedCycle.GetItem().GetStatus() != "finalized" {
		t.Fatalf("close cycle: res=%+v err=%v", closedCycle, err)
	}
}
