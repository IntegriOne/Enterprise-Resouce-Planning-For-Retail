//go:build integration

package learning_test

import (
	"context"
	"net"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/learning/v1"
	hrdb "github.com/integrione/erp-retail/backend/services/human_resource/db"
	"github.com/integrione/erp-retail/backend/services/human_resource/infrastructure/security"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/learning"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository"
	handler "github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/learning"
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

type learningPolicy struct{ principalID string }

func (p learningPolicy) Authorize(context.Context, operation.Access) (operation.Actor, error) {
	return operation.Actor{PrincipalID: p.principalID, SessionID: "00000000-0000-4000-8000-000000000002", Kind: "human"}, nil
}

func (learningPolicy) ReadScope(context.Context, operation.Access) (operation.Scope, error) {
	return operation.Scope{All: true}, nil
}

func text(v string) *string { return &v }

func TestLearningCourseCompetencyAndEnrollmentWithPostgreSQL(t *testing.T) {
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
	validators, err := security.LearningValidators(repos)
	if err != nil {
		t.Fatal(err)
	}

	actorID := "00000000-0000-4000-8000-0000000000b1"
	svc, err := service.NewService(service.Dependencies{
		Runtime: operation.Dependencies{
			Transactor: repos.Transactor, Outbox: repos.OutboxEvent,
			Policy: learningPolicy{principalID: actorID}, Integrity: security.NewIntegrity(validators),
		},
		LearningCourse:           repos.LearningCourse,
		LearningEnrollment:       repos.LearningEnrollment,
		LearningCourseCompetency: repos.LearningCourseCompetency,
	})
	if err != nil {
		t.Fatal(err)
	}

	listener := bufconn.Listen(1 << 20)
	server := grpc.NewServer(grpc.UnaryInterceptor(func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, next grpc.UnaryHandler) (any, error) {
		return next(authcontext.WithCaller(ctx, authcontext.CallerContext{PrincipalID: actorID, Authenticated: true}), req)
	}))
	if err := handler.Register(server, svc); err != nil {
		t.Fatal(err)
	}
	served := make(chan error, 1)
	go func() { served <- server.Serve(listener) }()
	t.Cleanup(func() { server.Stop(); _ = listener.Close(); <-served })
	conn, err := grpc.NewClient("passthrough:///hrm-learning-integration",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return listener.DialContext(ctx) }),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	client := pb.NewLearningServiceClient(conn)

	ids := operation.UUIDGenerator{}
	legalID, _ := ids.NewID()
	personID, _ := ids.NewID()
	workerID, _ := ids.NewID()
	compID, _ := ids.NewID()
	var courseID, enrollmentID, historicalEnrollmentID, withdrawID string

	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		if courseID != "" {
			if err := database.WithContext(cleanupCtx).Exec("DELETE FROM hr_learning_course_competencies WHERE course_id = ?", courseID).Error; err != nil {
				t.Errorf("cleanup course competency: %v", err)
			}
		}
		for _, stmt := range []struct {
			q string
			v string
		}{
			{"DELETE FROM hr_learning_enrollments WHERE id = ?", enrollmentID},
			{"DELETE FROM hr_learning_enrollments WHERE id = ?", historicalEnrollmentID},
			{"DELETE FROM hr_learning_enrollments WHERE id = ?", withdrawID},
			{"DELETE FROM hr_learning_courses WHERE id = ?", courseID},
			{"DELETE FROM hr_competencies WHERE id = ?", compID},
			{"DELETE FROM hr_workers WHERE id = ?", workerID},
		} {
			if stmt.v != "" {
				if err := database.WithContext(cleanupCtx).Exec(stmt.q, stmt.v).Error; err != nil {
					t.Errorf("cleanup %s: %v", stmt.q, err)
				}
			}
		}
		for _, agg := range []string{enrollmentID, courseID, withdrawID} {
			if agg != "" {
				if err := database.WithContext(cleanupCtx).Exec("DELETE FROM hr_outbox_events WHERE aggregate_id = ?", agg).Error; err != nil {
					t.Errorf("cleanup outbox %s: %v", agg, err)
				}
			}
		}
	})

	if err := database.WithContext(ctx).Exec(
		`INSERT INTO hr_workers (id, worker_number, person_party_reference_id, worker_category, status,
		 primary_legal_entity_reference_id, hire_date, service_start_date)
		 VALUES (?, ?, ?, 'employee', 'active', ?, '2026-01-01', '2026-01-01')`,
		workerID, "LRN-W-"+workerID, personID, legalID,
	).Error; err != nil {
		t.Fatal(err)
	}
	if err := database.WithContext(ctx).Exec(
		`INSERT INTO hr_competencies (id, competency_code, competency_version, competency_name, category_code, proficiency_scale, is_active)
		 VALUES (?, ?, 1, 'Cloud Architecture', 'technical', '{"1":"basic","5":"master"}'::jsonb, true)`,
		compID, "COMP-LRN-"+compID[:8],
	).Error; err != nil {
		t.Fatal(err)
	}

	active := true
	courseInput := &pb.LearningCourseInput{
		CourseCode: text("CRS-" + legalID[:8]), CourseVersion: text("v1"),
		CourseName: text("Go Microservices"), CourseType: text("technical"),
		DeliveryMode: text("self_paced"), DurationHours: text("4.00"),
		ValidFrom: text("2026-01-01"), IsActive: &active,
	}
	crsRes, err := client.CreateLearningCourse(ctx, &pb.CreateLearningCourseRequest{Input: courseInput})
	if err != nil {
		t.Fatalf("create course: %v", err)
	}
	courseID = crsRes.GetItem().GetId()
	if _, err := client.CreateLearningCourse(ctx, &pb.CreateLearningCourseRequest{Input: courseInput}); status.Code(err) != codes.AlreadyExists {
		t.Fatalf("duplicate course code+version = %v, want AlreadyExists", status.Code(err))
	}

	targetLvl := int32(3)
	linkInput := &pb.LearningCourseCompetencyInput{CourseId: &courseID, CompetencyId: &compID, TargetLevel: &targetLvl}
	if _, err := client.CreateLearningCourseCompetency(ctx, &pb.CreateLearningCourseCompetencyRequest{Input: linkInput}); err != nil {
		t.Fatalf("create course competency: %v", err)
	}
	if _, err := client.CreateLearningCourseCompetency(ctx, &pb.CreateLearningCourseCompetencyRequest{Input: linkInput}); status.Code(err) != codes.AlreadyExists {
		t.Fatalf("duplicate course competency = %v, want AlreadyExists", status.Code(err))
	}

	historicalEnrollmentID, err = ids.NewID()
	if err != nil {
		t.Fatal(err)
	}
	if err := database.WithContext(ctx).Exec(`INSERT INTO hr_learning_enrollments
		(id, worker_id, course_id, enrollment_status, assigned_at, idempotency_key)
		VALUES (?, ?, ?, 'withdrawn', NOW() - INTERVAL '1 day', ?)`,
		historicalEnrollmentID, workerID, courseID, "historical-"+workerID).Error; err != nil {
		t.Fatal(err)
	}

	for _, terminalStatus := range []string{"withdrawn", "failed", "expired"} {
		if err := database.WithContext(ctx).Exec("UPDATE hr_learning_enrollments SET enrollment_status = ? WHERE id = ?",
			terminalStatus, historicalEnrollmentID).Error; err != nil {
			t.Fatal(err)
		}
		if _, err := client.UpdateLearningEnrollment(ctx, &pb.UpdateLearningEnrollmentRequest{
			Id: historicalEnrollmentID, Input: &pb.LearningEnrollmentInput{DueAt: timestamppb.Now()},
			UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"due_at"}},
		}); status.Code(err) != codes.FailedPrecondition {
			t.Fatalf("update %s enrollment = %v, want FailedPrecondition", terminalStatus, err)
		}
	}

	type enrollmentResult struct {
		id  string
		err error
	}
	beginEnrollment := make(chan struct{})
	enrollments := make(chan enrollmentResult, 2)
	for _, keyPrefix := range []string{"idem-enr-1-", "idem-enr-2-"} {
		go func() {
			<-beginEnrollment
			result, err := client.CreateLearningEnrollment(ctx, &pb.CreateLearningEnrollmentRequest{Input: &pb.LearningEnrollmentInput{
				WorkerId: &workerID, CourseId: &courseID, EnrollmentStatus: text("planned"),
				IdempotencyKey: text(keyPrefix + workerID),
			}})
			enrollments <- enrollmentResult{id: result.GetItem().GetId(), err: err}
		}()
	}
	close(beginEnrollment)
	var duplicateEnrollments int
	for range 2 {
		result := <-enrollments
		switch {
		case result.err == nil && result.id != "":
			enrollmentID = result.id
		case status.Code(result.err) == codes.AlreadyExists:
			duplicateEnrollments++
		default:
			t.Errorf("create concurrent enrollment: %v", result.err)
		}
	}
	if enrollmentID == "" || duplicateEnrollments != 1 {
		t.Fatalf("concurrent enrollment: accepted ID %q, duplicates %d", enrollmentID, duplicateEnrollments)
	}

	if _, err := client.CreateLearningEnrollment(ctx, &pb.CreateLearningEnrollmentRequest{Input: &pb.LearningEnrollmentInput{
		WorkerId: &workerID, CourseId: &courseID, EnrollmentStatus: text("planned"),
		IdempotencyKey: text("idem-enr-duplicate-" + workerID),
	}}); status.Code(err) != codes.AlreadyExists {
		t.Fatalf("duplicate active enrollment with different key = %v, want AlreadyExists", err)
	}

	if _, err := client.UpdateLearningEnrollment(ctx, &pb.UpdateLearningEnrollmentRequest{
		Id: enrollmentID, Input: &pb.LearningEnrollmentInput{Score: text("95.000")},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"score"}},
	}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("bypass workflow setting score = %v, want FailedPrecondition", status.Code(err))
	}

	// Penghapusan kursus yang masih terhubung ke kompetensi dan enrollment ditolak.
	if _, err := client.DeleteLearningCourse(ctx, &common.IDRequest{Id: courseID}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("delete referenced course = %v, want FailedPrecondition", status.Code(err))
	}

	if _, err := client.DeleteLearningEnrollment(ctx, &common.IDRequest{Id: enrollmentID}); err != nil {
		t.Fatalf("delete planned enrollment: %v", err)
	}
	workflowEnrollment, err := client.CreateLearningEnrollment(ctx, &pb.CreateLearningEnrollmentRequest{Input: &pb.LearningEnrollmentInput{
		WorkerId: &workerID, CourseId: &courseID, EnrollmentStatus: text("planned"),
		IdempotencyKey: text("idem-enr-workflow-" + workerID),
	}})
	if err != nil {
		t.Fatalf("create learning workflow enrollment: %v", err)
	}
	enrollmentID = workflowEnrollment.GetItem().GetId()
	enrollmentRequest := &common.IDRequest{Id: enrollmentID}
	if _, err := client.CompleteLearningEnrollment(ctx, enrollmentRequest); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("complete before learning starts: %v", err)
	}
	assignedAt := workflowEnrollment.GetItem().GetAssignedAt().AsTime()
	if err := database.WithContext(ctx).Table("hr_learning_enrollments").Where("id = ?", enrollmentID).
		Update("assigned_at", assignedAt.AddDate(0, 0, -2)).Error; err != nil {
		t.Fatal(err)
	}
	courseRow := database.WithContext(ctx).Table("hr_learning_courses").Where("id = ?", courseID)
	if err := courseRow.Update("valid_to", assignedAt.AddDate(0, 0, -1).Format(time.DateOnly)).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := client.StartLearningEnrollment(ctx, enrollmentRequest); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("start learning after course validity ends: %v", err)
	}
	if err := courseRow.Update("valid_to", nil).Error; err != nil {
		t.Fatal(err)
	}
	started, err := client.StartLearningEnrollment(ctx, enrollmentRequest)
	if err != nil || started.GetItem().GetEnrollmentStatus() != "in_progress" || started.GetItem().GetStartedAt() == nil {
		t.Fatalf("start learning enrollment: %+v, %v", started, err)
	}
	if _, err := client.StartLearningEnrollment(ctx, enrollmentRequest); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("start learning again: %v", err)
	}
	inactive := false
	retire := &pb.UpdateLearningCourseRequest{
		Id: courseID, Input: &pb.LearningCourseInput{IsActive: &inactive},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"is_active"}},
	}
	if _, err := client.UpdateLearningCourse(ctx, retire); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("retire course with enrollment in progress: %v", err)
	}
	completed, err := client.CompleteLearningEnrollment(ctx, enrollmentRequest)
	if err != nil || completed.GetItem().GetEnrollmentStatus() != "completed" || completed.GetItem().GetCompletedAt() == nil ||
		completed.GetItem().GetCompletedAt().AsTime().Before(started.GetItem().GetStartedAt().AsTime()) {
		t.Fatalf("complete learning enrollment: %+v, %v", completed, err)
	}
	if _, err := client.CompleteLearningEnrollment(ctx, enrollmentRequest); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("complete learning again: %v", err)
	}
	if _, err := client.WithdrawLearningEnrollment(ctx, enrollmentRequest); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("withdraw completed enrollment: %v", err)
	}
	withdrawDraft, err := client.CreateLearningEnrollment(ctx, &pb.CreateLearningEnrollmentRequest{Input: &pb.LearningEnrollmentInput{
		WorkerId: &workerID, CourseId: &courseID, IdempotencyKey: text("withdraw-" + workerID),
	}})
	if err != nil {
		t.Fatal(err)
	}
	withdrawID = withdrawDraft.GetItem().GetId()
	withdrawRequest := &common.IDRequest{Id: withdrawID}
	if _, err := client.StartLearningEnrollment(ctx, withdrawRequest); err != nil {
		t.Fatal(err)
	}
	withdrawStarted, err := client.GetLearningEnrollmentByID(ctx, withdrawRequest)
	if err != nil {
		t.Fatal(err)
	}
	workerRow := database.WithContext(ctx).Table("hr_workers").Where("id = ?", workerID)
	if err := workerRow.Update("status", "suspended").Error; err != nil {
		t.Fatal(err)
	}
	withdrawn, err := client.WithdrawLearningEnrollment(ctx, withdrawRequest)
	if err != nil || withdrawn.GetItem().GetEnrollmentStatus() != "withdrawn" || withdrawn.GetItem().GetCompletedAt() != nil ||
		!withdrawn.GetItem().GetStartedAt().AsTime().Equal(withdrawStarted.GetItem().GetStartedAt().AsTime()) {
		t.Fatalf("withdraw started enrollment for suspended worker: %+v, %v", withdrawn, err)
	}
	if _, err := client.StartLearningEnrollment(ctx, withdrawRequest); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("start withdrawn enrollment: %v", err)
	}
	if _, err := client.WithdrawLearningEnrollment(ctx, withdrawRequest); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("withdraw enrollment again: %v", err)
	}
	var withdrawalEvents int64
	if err := database.WithContext(ctx).Table("hr_outbox_events").
		Where("aggregate_id = ? AND event_type = ?", withdrawID, "hr_learning_enrollments.withdraw").
		Count(&withdrawalEvents).Error; err != nil || withdrawalEvents != 1 {
		t.Fatalf("withdrawal outbox: count=%d, error=%v", withdrawalEvents, err)
	}
	if err := workerRow.Update("status", "active").Error; err != nil {
		t.Fatal(err)
	}
	retired, err := client.UpdateLearningCourse(ctx, retire)
	if err != nil || retired.GetItem().GetIsActive() {
		t.Fatalf("retire course with completed enrollment: %+v, %v", retired, err)
	}
	if _, err := client.CreateLearningEnrollment(ctx, &pb.CreateLearningEnrollmentRequest{Input: &pb.LearningEnrollmentInput{
		WorkerId: &workerID, CourseId: &courseID, IdempotencyKey: text("retired-" + workerID),
	}}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("assign retired course: %v", err)
	}
	if err := database.WithContext(ctx).Exec("DELETE FROM hr_learning_enrollments WHERE id = ?", withdrawID).Error; err != nil {
		t.Fatal(err)
	}
	if err := database.WithContext(ctx).Exec("DELETE FROM hr_learning_enrollments WHERE id = ?", enrollmentID).Error; err != nil {
		t.Fatal(err)
	}
	enrollmentID = ""
	if err := database.WithContext(ctx).Exec("DELETE FROM hr_learning_enrollments WHERE id = ?", historicalEnrollmentID).Error; err != nil {
		t.Fatal(err)
	}
	historicalEnrollmentID = ""
	if _, err := client.DeleteLearningCourseCompetency(ctx, &pb.LearningCourseCompetencyKey{
		CourseId: courseID, CompetencyId: compID,
	}); err != nil {
		t.Fatalf("delete course competency: %v", err)
	}
	if _, err := client.DeleteLearningCourse(ctx, &common.IDRequest{Id: courseID}); err != nil {
		t.Fatalf("delete unreferenced course: %v", err)
	}
	courseID = ""
}
