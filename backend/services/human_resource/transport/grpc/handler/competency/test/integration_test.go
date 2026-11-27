//go:build integration

package test

import (
	"context"
	"net"
	"net/url"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/competency/v1"
	learningpb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/learning/v1"
	hrdb "github.com/integrione/erp-retail/backend/services/human_resource/db"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	"github.com/integrione/erp-retail/backend/services/human_resource/infrastructure/security"
	competencyservice "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/competency"
	learningservice "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/learning"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository"
	competencyhandler "github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/competency"
	learninghandler "github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/learning"
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

type integrationPolicy struct{ denyDelete atomic.Bool }

func (p *integrationPolicy) Authorize(_ context.Context, access operation.Access) (operation.Actor, error) {
	if p.denyDelete.Load() && access.Action == "delete" {
		return operation.Actor{}, hrerrors.ErrForbidden
	}
	return operation.Actor{
		PrincipalID: "00000000-0000-4000-8000-000000000001",
		SessionID:   "00000000-0000-4000-8000-000000000002",
		Kind:        "human",
	}, nil
}

func (*integrationPolicy) ReadScope(context.Context, operation.Access) (operation.Scope, error) {
	return operation.Scope{All: true}, nil
}

func TestCompetencyAndLearningRPCWithPostgreSQL(t *testing.T) {
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
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal("open HRM test database failed")
	}
	pool, err := db.DB()
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
	repos, err := repository.New(db)
	if err != nil {
		t.Fatal(err)
	}
	validators, err := security.CompetencyValidators(repos)
	if err != nil {
		t.Fatal(err)
	}
	learningValidators, err := security.LearningValidators(repos)
	if err != nil {
		t.Fatal(err)
	}
	for resource, validator := range learningValidators {
		validators[resource] = validator
	}
	policy := &integrationPolicy{}
	runtime := operation.Dependencies{
		Transactor: repos.Transactor, Outbox: repos.OutboxEvent,
		Policy: policy, Integrity: security.NewIntegrity(validators),
	}
	svc, err := competencyservice.NewService(competencyservice.Dependencies{
		Runtime:             runtime,
		WorkerQualification: repos.WorkerQualification, Competency: repos.Competency,
		JobCompetency: repos.JobCompetency, WorkerCompetencyAssessment: repos.WorkerCompetencyAssessment,
	})
	if err != nil {
		t.Fatal(err)
	}
	learningSvc, err := learningservice.NewService(learningservice.Dependencies{
		Runtime: runtime, LearningCourse: repos.LearningCourse,
		LearningEnrollment: repos.LearningEnrollment, LearningCourseCompetency: repos.LearningCourseCompetency,
	})
	if err != nil {
		t.Fatal(err)
	}
	listener := bufconn.Listen(1 << 20)
	server := grpc.NewServer()
	if err := competencyhandler.Register(server, svc); err != nil {
		t.Fatal(err)
	}
	if err := learninghandler.Register(server, learningSvc); err != nil {
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
	connection, err := grpc.NewClient("passthrough:///hrm-competency-integration",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return listener.DialContext(ctx)
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := connection.Close(); err != nil {
			t.Errorf("close HRM test gRPC connection: %v", err)
		}
	})
	client := pb.NewCompetencyServiceClient(connection)
	learningClient := learningpb.NewLearningServiceClient(connection)
	ids := operation.UUIDGenerator{}
	familyID, err := ids.NewID()
	if err != nil {
		t.Fatal(err)
	}
	jobID, err := ids.NewID()
	if err != nil {
		t.Fatal(err)
	}
	workerID, err := ids.NewID()
	if err != nil {
		t.Fatal(err)
	}
	var requirementID string
	var competencyID string
	var courseID string
	var enrollmentID string
	var assessmentID string
	var qualificationID string
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		for _, statement := range []struct {
			query string
			value string
		}{
			{"DELETE FROM hr_learning_enrollments WHERE id = ?", enrollmentID},
			{"DELETE FROM hr_outbox_events WHERE aggregate_id = ?", enrollmentID},
			{"DELETE FROM hr_worker_competency_assessments WHERE id = ?", assessmentID},
			{"DELETE FROM hr_outbox_events WHERE aggregate_id = ?", assessmentID},
			{"DELETE FROM hr_worker_qualifications WHERE id = ?", qualificationID},
			{"DELETE FROM hr_outbox_events WHERE aggregate_id = ?", qualificationID},
			{"DELETE FROM hr_workers WHERE id = ?", workerID},
			{"DELETE FROM hr_learning_course_competencies WHERE course_id = ?", courseID},
			{"DELETE FROM hr_outbox_events WHERE aggregate_id = ?", courseID},
			{"DELETE FROM hr_learning_courses WHERE id = ?", courseID},
			{"DELETE FROM hr_job_competencies WHERE id = ?", requirementID},
			{"DELETE FROM hr_outbox_events WHERE aggregate_id = ?", requirementID},
			{"DELETE FROM hr_jobs WHERE id = ?", jobID},
			{"DELETE FROM hr_job_families WHERE id = ?", familyID},
			{"DELETE FROM hr_outbox_events WHERE aggregate_id = ?", competencyID},
			{"DELETE FROM hr_competencies WHERE id = ?", competencyID},
		} {
			if err := db.WithContext(cleanupCtx).Exec(statement.query, statement.value).Error; err != nil {
				t.Errorf("clean HRM test data: %v", err)
			}
		}
	})
	code := "TEST-" + familyID
	name := "Leadership"
	category := "LEADERSHIP"
	version := int32(1)
	input := &pb.CompetencyInput{
		CompetencyCode: &code, CompetencyVersion: &version,
		CompetencyName: &name, CategoryCode: &category,
		ProficiencyScale: &common.JsonDocument{Utf8: []byte(`{"1":"basic"}`)},
	}
	created, err := client.CreateCompetency(ctx, &pb.CreateCompetencyRequest{Input: input})
	if err != nil || created.GetItem() == nil {
		t.Fatalf("create competency: %v", err)
	}
	competencyID = created.Item.Id
	if competencyID == "" || !created.Item.IsActive {
		t.Fatal("created competency has no identity or active default")
	}
	read, err := client.GetCompetencyByID(ctx, &common.IDRequest{Id: competencyID})
	if err != nil || read.GetItem().GetCompetencyCode() != code {
		t.Fatalf("read competency: %v", err)
	}
	found, err := client.FindCompetencyByCompetencyCode(ctx, &common.StringFinderRequest{Value: &code, Limit: 1})
	if err != nil || len(found.GetItems()) != 1 || found.GetPage().GetTotal() != 1 {
		t.Fatalf("find competency by code: %v", err)
	}
	nextPage, err := client.FindCompetencyByCompetencyCode(ctx, &common.StringFinderRequest{Value: &code, Limit: 1, Offset: 1})
	if err != nil || len(nextPage.GetItems()) != 0 || nextPage.GetPage().GetTotal() != 1 {
		t.Fatalf("find competency beyond first page: %v", err)
	}
	listed, err := client.GetAllCompetencies(ctx, &common.PageRequest{Limit: 1})
	if err != nil || len(listed.GetItems()) != 1 || listed.GetPage().GetLimit() != 1 || listed.GetPage().GetTotal() < 1 {
		t.Fatalf("list competencies with bounded page: %v", err)
	}
	if _, err := client.CreateCompetency(ctx, &pb.CreateCompetencyRequest{Input: input}); status.Code(err) != codes.AlreadyExists {
		t.Fatalf("duplicate code and version: %v", err)
	}
	updatedName := "Leadership and coaching"
	updated, err := client.UpdateCompetency(ctx, &pb.UpdateCompetencyRequest{
		Id: competencyID, Input: &pb.CompetencyInput{CompetencyName: &updatedName},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"competency_name"}},
	})
	if err != nil || updated.GetItem().GetCompetencyName() != updatedName {
		t.Fatalf("update competency metadata: %v", err)
	}
	if err := db.WithContext(ctx).Exec(
		"INSERT INTO hr_job_families (id, family_code, family_name, job_domain) VALUES (?, ?, ?, ?)",
		familyID, code, "Test family", "professional",
	).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.WithContext(ctx).Exec(
		"INSERT INTO hr_jobs (id, job_family_id, job_code, job_title, valid_from) VALUES (?, ?, ?, ?, ?)",
		jobID, familyID, code, "Test job", "2026-01-01",
	).Error; err != nil {
		t.Fatal(err)
	}
	level := int32(2)
	effectiveFrom := "2026-01-01"
	requirementInput := &pb.JobCompetencyInput{
		JobId: &jobID, CompetencyId: &competencyID,
		RequiredLevel: &level, EffectiveFrom: &effectiveFrom,
	}
	requirement, err := client.CreateJobCompetency(ctx, &pb.CreateJobCompetencyRequest{Input: requirementInput})
	if err != nil || requirement.GetItem() == nil {
		t.Fatalf("create job competency through gRPC: %v", err)
	}
	requirementID = requirement.Item.Id
	if requirementID == "" {
		t.Fatal("created job competency has no identity")
	}
	overlapFrom := "2026-03-01"
	overlapping := &pb.JobCompetencyInput{
		JobId: &jobID, CompetencyId: &competencyID,
		RequiredLevel: &level, EffectiveFrom: &overlapFrom,
	}
	if _, err := client.CreateJobCompetency(ctx, &pb.CreateJobCompetencyRequest{Input: overlapping}); status.Code(err) != codes.AlreadyExists {
		t.Fatalf("overlapping job competency requirement: %v", err)
	}
	courseCode := "COURSE-" + familyID
	courseVersion := "1"
	courseName := "Safety basics"
	courseType := "safety"
	deliveryMode := "classroom"
	course, err := learningClient.CreateLearningCourse(ctx, &learningpb.CreateLearningCourseRequest{Input: &learningpb.LearningCourseInput{
		CourseCode: &courseCode, CourseVersion: &courseVersion, CourseName: &courseName,
		CourseType: &courseType, DeliveryMode: &deliveryMode,
	}})
	if err != nil || course.GetItem() == nil {
		t.Fatalf("create learning course through gRPC: %v", err)
	}
	courseID = course.Item.Id
	if courseID == "" {
		t.Fatal("created course has no identity")
	}
	externalCourseCode := "PROVIDER-" + familyID
	if _, err := learningClient.CreateLearningCourse(ctx, &learningpb.CreateLearningCourseRequest{Input: &learningpb.LearningCourseInput{
		CourseCode: &externalCourseCode, CourseVersion: &courseVersion, CourseName: &courseName,
		CourseType: &courseType, DeliveryMode: &deliveryMode, ProviderPartyReferenceId: &familyID,
	}}); status.Code(err) != codes.Unavailable {
		t.Fatalf("unverified external course provider: %v", err)
	}
	targetLevel := int32(1)
	link, err := learningClient.CreateLearningCourseCompetency(ctx, &learningpb.CreateLearningCourseCompetencyRequest{Input: &learningpb.LearningCourseCompetencyInput{
		CourseId: &courseID, CompetencyId: &competencyID, TargetLevel: &targetLevel,
	}})
	if err != nil || link.GetItem() == nil {
		t.Fatalf("create course competency through gRPC: %v", err)
	}
	if _, err := learningClient.CreateLearningCourseCompetency(ctx, &learningpb.CreateLearningCourseCompetencyRequest{Input: &learningpb.LearningCourseCompetencyInput{
		CourseId: &courseID, CompetencyId: &competencyID, TargetLevel: &targetLevel,
	}}); status.Code(err) != codes.AlreadyExists {
		t.Fatalf("duplicate course competency: %v", err)
	}
	nextVersion := "2"
	if _, err := learningClient.UpdateLearningCourse(ctx, &learningpb.UpdateLearningCourseRequest{
		Id: courseID, Input: &learningpb.LearningCourseInput{CourseVersion: &nextVersion},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"course_version"}},
	}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("referenced course version changed: %v", err)
	}
	linkKey := &learningpb.LearningCourseCompetencyKey{CourseId: courseID, CompetencyId: competencyID}
	readLink, err := learningClient.GetLearningCourseCompetencyByID(ctx, linkKey)
	if err != nil || readLink.GetItem().GetTargetLevel() != targetLevel {
		t.Fatalf("read course competency: %v", err)
	}
	listedLinks, err := learningClient.GetAllLearningCourseCompetencies(ctx, &common.PageRequest{Limit: 1})
	if err != nil || listedLinks.GetPage().GetTotal() < 1 || len(listedLinks.GetItems()) != 1 {
		t.Fatalf("list course competencies: %v", err)
	}
	nextLevel := int32(2)
	updatedLink, err := learningClient.UpdateLearningCourseCompetency(ctx, &learningpb.UpdateLearningCourseCompetencyRequest{
		Key: linkKey, TargetLevel: &nextLevel,
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"target_level"}},
	})
	if err != nil || updatedLink.GetItem().GetTargetLevel() != nextLevel {
		t.Fatalf("update course competency: %v", err)
	}
	if _, err := learningClient.DeleteLearningCourseCompetency(ctx, linkKey); err != nil {
		t.Fatalf("delete course competency: %v", err)
	}
	updatedCourse, err := learningClient.UpdateLearningCourse(ctx, &learningpb.UpdateLearningCourseRequest{
		Id: courseID, Input: &learningpb.LearningCourseInput{CourseVersion: &nextVersion},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"course_version"}},
	})
	if err != nil || updatedCourse.GetItem().GetCourseVersion() != nextVersion {
		t.Fatalf("update unreferenced course version: %v", err)
	}
	if err := db.WithContext(ctx).Exec(
		"INSERT INTO hr_workers (id, worker_number, person_party_reference_id, worker_category, status, primary_legal_entity_reference_id) VALUES (?, ?, ?, ?, ?, ?)",
		workerID, "WORKER-"+workerID, familyID, "employee", "active", familyID,
	).Error; err != nil {
		t.Fatal(err)
	}
	qualificationType := "certification"
	qualificationTitle := "Safety certificate"
	qualificationInput := &pb.WorkerQualificationInput{
		WorkerId: &workerID, QualificationType: &qualificationType, Title: &qualificationTitle,
	}
	qualificationInput.IssuerPartyReferenceId = &familyID
	if _, err := client.CreateWorkerQualification(ctx, &pb.CreateWorkerQualificationRequest{Input: qualificationInput}); status.Code(err) != codes.Unavailable {
		t.Fatalf("unverified qualification issuer: %v", err)
	}
	qualificationInput.IssuerPartyReferenceId = nil
	createdQualification, err := client.CreateWorkerQualification(ctx, &pb.CreateWorkerQualificationRequest{Input: qualificationInput})
	if err != nil || createdQualification.GetItem().GetId() == "" || createdQualification.GetItem().GetVerificationStatus() != "unverified" {
		t.Fatalf("create qualification draft: %+v, %v", createdQualification, err)
	}
	qualificationID = createdQualification.Item.Id
	if got, err := client.GetWorkerQualificationByID(ctx, &common.IDRequest{Id: qualificationID}); err != nil || got.GetItem().GetWorkerId() != workerID {
		t.Fatalf("read qualification draft: %+v, %v", got, err)
	}
	if found, err := client.FindWorkerQualificationByWorkerID(ctx, &common.StringFinderRequest{Value: &workerID, Limit: 1}); err != nil ||
		found.GetPage().GetTotal() != 1 || len(found.GetItems()) != 1 {
		t.Fatalf("find qualification by worker: %+v, %v", found, err)
	}
	verifiedStatus := "verified"
	if _, err := client.UpdateWorkerQualification(ctx, &pb.UpdateWorkerQualificationRequest{
		Id: qualificationID, Input: &pb.WorkerQualificationInput{VerificationStatus: &verifiedStatus},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"verification_status"}},
	}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("verification through generic update: %v", err)
	}
	updatedQualificationTitle := "Updated safety certificate"
	if updated, err := client.UpdateWorkerQualification(ctx, &pb.UpdateWorkerQualificationRequest{
		Id: qualificationID, Input: &pb.WorkerQualificationInput{Title: &updatedQualificationTitle},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"title"}},
	}); err != nil || updated.GetItem().GetTitle() != updatedQualificationTitle {
		t.Fatalf("update qualification draft: %+v, %v", updated, err)
	}
	if _, err := client.DeleteWorkerQualification(ctx, &common.IDRequest{Id: qualificationID}); err != nil {
		t.Fatalf("delete qualification draft: %v", err)
	}
	assessmentMethod := "manual_observation"
	assessmentKey := "competency-assessment-" + workerID
	assessmentAt := timestamppb.New(time.Now().Add(-time.Hour))
	assessmentInput := &pb.WorkerCompetencyAssessmentInput{
		WorkerId: &workerID, CompetencyId: &competencyID, AssessedLevel: &level,
		AssessmentMethod: &assessmentMethod, AssessedAt: assessmentAt, IdempotencyKey: &assessmentKey,
	}
	assessmentInput.EvidenceDocumentVersionReferenceId = &familyID
	assessmentInput.EvidenceDigest = []byte("unverified evidence")
	if _, err := client.CreateWorkerCompetencyAssessment(ctx, &pb.CreateWorkerCompetencyAssessmentRequest{Input: assessmentInput}); status.Code(err) != codes.Unavailable {
		t.Fatalf("unverified assessment evidence status = %v, want Unavailable", status.Code(err))
	}
	assessmentInput.EvidenceDocumentVersionReferenceId = nil
	assessmentInput.EvidenceDigest = nil
	assessment, err := client.CreateWorkerCompetencyAssessment(ctx, &pb.CreateWorkerCompetencyAssessmentRequest{Input: assessmentInput})
	if err != nil || assessment.GetItem().GetId() == "" || assessment.GetItem().GetAssessedByPrincipalId() != "00000000-0000-4000-8000-000000000001" {
		t.Fatalf("create manual assessment and bind actor: %+v, %v", assessment, err)
	}
	assessmentID = assessment.GetItem().GetId()
	if _, err := client.CreateWorkerCompetencyAssessment(ctx, &pb.CreateWorkerCompetencyAssessmentRequest{Input: assessmentInput}); status.Code(err) != codes.AlreadyExists {
		t.Fatalf("duplicate assessment key status = %v, want AlreadyExists", status.Code(err))
	}
	if got, err := client.GetWorkerCompetencyAssessmentByID(ctx, &common.IDRequest{Id: assessmentID}); err != nil ||
		got.GetItem().GetAssessedLevel() != level {
		t.Fatalf("read assessment: %+v, %v", got, err)
	}
	if page, err := client.FindWorkerCompetencyAssessmentByWorkerID(ctx, &common.StringFinderRequest{Value: &workerID, Limit: 1}); err != nil ||
		page.GetPage().GetTotal() != 1 || len(page.GetItems()) != 1 {
		t.Fatalf("find assessment: %+v, %v", page, err)
	}
	if page, err := client.FindWorkerCompetencyAssessmentByWorkerID(ctx, &common.StringFinderRequest{Value: &workerID, Limit: 1, Offset: 1}); err != nil ||
		page.GetPage().GetTotal() != 1 || len(page.GetItems()) != 0 {
		t.Fatalf("assessment offset: %+v, %v", page, err)
	}
	if _, err := client.UpdateWorkerCompetencyAssessment(ctx, &pb.UpdateWorkerCompetencyAssessmentRequest{Id: assessmentID,
		Input: &pb.WorkerCompetencyAssessmentInput{AssessedLevel: &level}, UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"assessed_level"}}}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("overwrite assessment status = %v, want FailedPrecondition", status.Code(err))
	}
	if _, err := client.DeleteWorkerCompetencyAssessment(ctx, &common.IDRequest{Id: assessmentID}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("delete assessment status = %v, want FailedPrecondition", status.Code(err))
	}
	var assessmentEvents int64
	if err := db.WithContext(ctx).Table("hr_outbox_events").Where("aggregate_id = ?", assessmentID).Count(&assessmentEvents).Error; err != nil || assessmentEvents != 1 {
		t.Fatalf("assessment outbox count = %d, %v", assessmentEvents, err)
	}
	idempotencyKey := "enrollment-" + workerID
	spoofedAssigner := familyID
	enrollmentInput := &learningpb.LearningEnrollmentInput{
		WorkerId: &workerID, CourseId: &courseID, IdempotencyKey: &idempotencyKey,
		AssignedByPrincipalId: &spoofedAssigner,
	}
	enrollment, err := learningClient.CreateLearningEnrollment(ctx, &learningpb.CreateLearningEnrollmentRequest{Input: enrollmentInput})
	if err != nil || enrollment.GetItem() == nil {
		t.Fatalf("create learning enrollment through gRPC: %v", err)
	}
	enrollmentID = enrollment.Item.Id
	if enrollmentID == "" || enrollment.Item.GetAssignedByPrincipalId() != "00000000-0000-4000-8000-000000000001" {
		t.Fatal("server did not bind assignment identity")
	}
	if _, err := learningClient.CreateLearningEnrollment(ctx, &learningpb.CreateLearningEnrollmentRequest{Input: enrollmentInput}); status.Code(err) != codes.AlreadyExists {
		t.Fatalf("duplicate enrollment idempotency key: %v", err)
	}
	score := "98.000"
	if _, err := learningClient.CreateLearningEnrollment(ctx, &learningpb.CreateLearningEnrollmentRequest{Input: &learningpb.LearningEnrollmentInput{
		WorkerId: &workerID, CourseId: &courseID, IdempotencyKey: &idempotencyKey, Score: &score,
	}}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("score through enrollment CRUD: %v", err)
	}
	readEnrollment, err := learningClient.GetLearningEnrollmentByID(ctx, &common.IDRequest{Id: enrollmentID})
	if err != nil || readEnrollment.GetItem().GetEnrollmentStatus() != "planned" {
		t.Fatalf("read learning enrollment: %v", err)
	}
	dueAt := timestamppb.New(time.Now().Add(48 * time.Hour))
	updatedEnrollment, err := learningClient.UpdateLearningEnrollment(ctx, &learningpb.UpdateLearningEnrollmentRequest{
		Id: enrollmentID, Input: &learningpb.LearningEnrollmentInput{DueAt: dueAt},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"due_at"}},
	})
	if err != nil || updatedEnrollment.GetItem().GetDueAt() == nil {
		t.Fatalf("update planned enrollment due date: %v", err)
	}
	if _, err := learningClient.DeleteLearningEnrollment(ctx, &common.IDRequest{Id: enrollmentID}); err != nil {
		t.Fatalf("delete planned learning enrollment: %v", err)
	}
	if _, err := client.UpdateCompetency(ctx, &pb.UpdateCompetencyRequest{
		Id:         competencyID,
		Input:      &pb.CompetencyInput{ProficiencyScale: &common.JsonDocument{Utf8: []byte(`{"1":"changed"}`)}},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"proficiency_scale"}},
	}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("referenced rubric change: %v", err)
	}
	if _, err := client.DeleteCompetency(ctx, &common.IDRequest{Id: competencyID}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("referenced competency deletion: %v", err)
	}
	policy.denyDelete.Store(true)
	if _, err := client.DeleteCompetency(ctx, &common.IDRequest{Id: competencyID}); status.Code(err) != codes.PermissionDenied {
		t.Fatalf("permission denial: %v", err)
	}
	read, err = client.GetCompetencyByID(ctx, &common.IDRequest{Id: competencyID})
	if err != nil || read.GetItem().GetCompetencyName() != updatedName {
		t.Fatalf("denied operations changed competency: %v", err)
	}
	var events int64
	if err := db.WithContext(ctx).Raw("SELECT count(*) FROM hr_outbox_events WHERE aggregate_id = ?", competencyID).Scan(&events).Error; err != nil {
		t.Fatal(err)
	}
	if events != 2 {
		t.Fatalf("committed event count = %d, want 2", events)
	}
	if err := db.WithContext(ctx).Raw("SELECT count(*) FROM hr_outbox_events WHERE aggregate_id = ?", requirementID).Scan(&events).Error; err != nil {
		t.Fatal(err)
	}
	if events != 1 {
		t.Fatalf("job competency event count = %d, want 1", events)
	}
	if err := db.WithContext(ctx).Raw("SELECT count(*) FROM hr_outbox_events WHERE aggregate_id = ?", courseID).Scan(&events).Error; err != nil {
		t.Fatal(err)
	}
	if events != 5 {
		t.Fatalf("learning course event count = %d, want 5", events)
	}
	if err := db.WithContext(ctx).Raw("SELECT count(*) FROM hr_outbox_events WHERE aggregate_id = ?", enrollmentID).Scan(&events).Error; err != nil {
		t.Fatal(err)
	}
	if events != 3 {
		t.Fatalf("learning enrollment event count = %d, want 3", events)
	}
}
