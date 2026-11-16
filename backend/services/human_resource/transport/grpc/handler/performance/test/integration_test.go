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
	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/performance/v1"
	hrdb "github.com/integrione/erp-retail/backend/services/human_resource/db"
	"github.com/integrione/erp-retail/backend/services/human_resource/infrastructure/security"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/performance"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository"
	handler "github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/performance"
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

type integrationPolicy struct{}

func (integrationPolicy) Authorize(context.Context, operation.Access) (operation.Actor, error) {
	return operation.Actor{
		PrincipalID: "00000000-0000-4000-8000-000000000001",
		SessionID:   "00000000-0000-4000-8000-000000000002",
		Kind:        "human",
	}, nil
}

func (integrationPolicy) ReadScope(context.Context, operation.Access) (operation.Scope, error) {
	return operation.Scope{All: true}, nil
}

func TestPerformanceCycleRPCWithPostgreSQL(t *testing.T) {
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
	validators, err := security.PerformanceValidators(repos)
	if err != nil {
		t.Fatal(err)
	}
	svc, err := service.NewService(service.Dependencies{
		Runtime: operation.Dependencies{
			Transactor: repos.Transactor, Outbox: repos.OutboxEvent,
			Policy: integrationPolicy{}, Integrity: security.NewIntegrity(validators),
		},
		PerformanceCycle: repos.PerformanceCycle, PerformanceGoal: repos.PerformanceGoal,
		PerformanceReview: repos.PerformanceReview, PerformanceGoalUpdate: repos.PerformanceGoalUpdate,
		PerformanceReviewGoalScore: repos.PerformanceReviewGoalScore,
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
	connection, err := grpc.NewClient("passthrough:///hrm-performance-integration",
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
	client := pb.NewPerformanceServiceClient(connection)
	ids := operation.UUIDGenerator{}
	workerID, err := ids.NewID()
	if err != nil {
		t.Fatal(err)
	}
	var cycleID, goalID string
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		for _, statement := range []struct{ query, value string }{
			{"DELETE FROM hr_performance_goals WHERE id = ?", goalID},
			{"DELETE FROM hr_outbox_events WHERE aggregate_id = ?", goalID},
			{"DELETE FROM hr_performance_cycles WHERE id = ?", cycleID},
			{"DELETE FROM hr_outbox_events WHERE aggregate_id = ?", cycleID},
			{"DELETE FROM hr_workers WHERE id = ?", workerID},
		} {
			if statement.value == "" {
				continue
			}
			if err := database.WithContext(cleanupCtx).Exec(statement.query, statement.value).Error; err != nil {
				t.Errorf("cleanup HRM performance test data: %v", err)
			}
		}
	})
	code, name, kind := "PERF-"+workerID, "Annual review", "annual"
	start, end := "2027-01-01", "2027-12-31"
	input := &pb.PerformanceCycleInput{CycleCode: &code, CycleName: &name, CycleType: &kind, PeriodStart: &start, PeriodEnd: &end}
	created, err := client.CreatePerformanceCycle(ctx, &pb.CreatePerformanceCycleRequest{Input: input})
	if err != nil || created.GetItem().GetId() == "" || created.GetItem().GetStatus() != "draft" {
		t.Fatalf("create draft cycle: %v", err)
	}
	cycleID = created.Item.Id
	if _, err := client.CreatePerformanceCycle(ctx, &pb.CreatePerformanceCycleRequest{Input: input}); status.Code(err) != codes.AlreadyExists {
		t.Fatalf("duplicate cycle code: %v", err)
	}
	found, err := client.FindPerformanceCycleByCycleCode(ctx, &common.StringFinderRequest{Value: &code, Limit: 1})
	if err != nil || len(found.GetItems()) != 1 || found.GetPage().GetTotal() != 1 {
		t.Fatalf("paged finder: %v", err)
	}
	if _, err := client.CreatePerformanceCycle(ctx, &pb.CreatePerformanceCycleRequest{Input: &pb.PerformanceCycleInput{
		CycleCode: &code, CycleName: &name, CycleType: &kind, PeriodStart: &start, PeriodEnd: &end,
		LaunchedByPrincipalId: &workerID,
	}}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("client supplied launch actor: %v", err)
	}
	revised := "Annual review revised"
	if _, err := client.UpdatePerformanceCycle(ctx, &pb.UpdatePerformanceCycleRequest{
		Id: cycleID, Input: &pb.PerformanceCycleInput{CycleName: &revised},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"cycle_name"}},
	}); err != nil {
		t.Fatalf("update unreferenced draft: %v", err)
	}
	if err := database.WithContext(ctx).Exec(
		"INSERT INTO hr_workers (id, worker_number, person_party_reference_id, worker_category, status, primary_legal_entity_reference_id) VALUES (?, ?, ?, ?, ?, ?)",
		workerID, "WORKER-"+workerID, workerID, "employee", "active", workerID,
	).Error; err != nil {
		t.Fatal(err)
	}
	goalTitle, goalType, goalCode := "Service quality", "business", "quality-1"
	goalInput := &pb.PerformanceGoalInput{
		WorkerId: &workerID, PerformanceCycleId: &cycleID, GoalTitle: &goalTitle,
		GoalType: &goalType, GoalCode: &goalCode,
	}
	if _, err := client.CreatePerformanceGoal(ctx, &pb.CreatePerformanceGoalRequest{Input: goalInput}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("goal before cycle launch: %v", err)
	}
	launched, err := client.LaunchPerformanceCycle(ctx, &common.IDRequest{Id: cycleID})
	if err != nil || launched.GetItem().GetStatus() != "open" || launched.GetItem().GetLaunchedAt() == nil ||
		launched.GetItem().GetLaunchedByPrincipalId() == "" {
		t.Fatalf("launch draft cycle: %v", err)
	}
	if _, err := client.LaunchPerformanceCycle(ctx, &common.IDRequest{Id: cycleID}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("repeat launch: %v", err)
	}
	goalCreated, err := client.CreatePerformanceGoal(ctx, &pb.CreatePerformanceGoalRequest{Input: goalInput})
	if err != nil || goalCreated.GetItem().GetId() == "" || goalCreated.GetItem().GetGoalStatus() != "draft" {
		t.Fatalf("create goal in open cycle: %v", err)
	}
	goalID = goalCreated.Item.Id
	readGoal, err := client.GetPerformanceGoalByID(ctx, &common.IDRequest{Id: goalID})
	if err != nil || readGoal.GetItem().GetGoalCode() != goalCode {
		t.Fatalf("read persisted goal: %v", err)
	}
	foundGoals, err := client.FindPerformanceGoalByWorkerID(ctx, &common.StringFinderRequest{Value: &workerID, Limit: 1})
	if err != nil || len(foundGoals.GetItems()) != 1 || foundGoals.GetPage().GetTotal() != 1 {
		t.Fatalf("paged worker goal finder: %v", err)
	}
	updatedGoalTitle := "Service quality revised"
	if _, err := client.UpdatePerformanceGoal(ctx, &pb.UpdatePerformanceGoalRequest{
		Id: goalID, Input: &pb.PerformanceGoalInput{GoalTitle: &updatedGoalTitle},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"goal_title"}},
	}); err != nil {
		t.Fatalf("update unreferenced draft goal: %v", err)
	}
	if _, err := client.CreatePerformanceGoal(ctx, &pb.CreatePerformanceGoalRequest{Input: goalInput}); status.Code(err) != codes.AlreadyExists {
		t.Fatalf("duplicate worker/cycle/goal code: %v", err)
	}
	actualValue := "1.00"
	if _, err := client.CreatePerformanceGoal(ctx, &pb.CreatePerformanceGoalRequest{Input: &pb.PerformanceGoalInput{
		WorkerId: &workerID, PerformanceCycleId: &cycleID, GoalTitle: &goalTitle,
		GoalType: &goalType, ActualValue: &actualValue,
	}}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("client supplied actual value: %v", err)
	}
	if _, err := client.UpdatePerformanceCycle(ctx, &pb.UpdatePerformanceCycleRequest{
		Id: cycleID, Input: &pb.PerformanceCycleInput{CycleName: &name},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"cycle_name"}},
	}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("update referenced cycle: %v", err)
	}
	if _, err := client.DeletePerformanceCycle(ctx, &common.IDRequest{Id: cycleID}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("delete referenced cycle: %v", err)
	}
	var events int64
	if err := database.WithContext(ctx).Table("hr_outbox_events").Where("aggregate_id = ?", cycleID).Count(&events).Error; err != nil || events != 3 {
		t.Fatalf("outbox after rejected mutations: count=%d error=%v", events, err)
	}
	if err := database.WithContext(ctx).Table("hr_outbox_events").Where("aggregate_id = ?", goalID).Count(&events).Error; err != nil || events != 2 {
		t.Fatalf("goal outbox after rejected mutations: count=%d error=%v", events, err)
	}
	if _, err := client.DeletePerformanceGoal(ctx, &common.IDRequest{Id: goalID}); err != nil {
		t.Fatalf("delete unreferenced draft goal: %v", err)
	}
	if err := database.WithContext(ctx).Table("hr_outbox_events").Where("aggregate_id = ?", goalID).Count(&events).Error; err != nil || events != 3 {
		t.Fatalf("goal outbox after deletion: count=%d error=%v", events, err)
	}
}
