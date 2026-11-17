//go:build integration

package recruitment_test

import (
	"context"
	"net"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/recruitment/v1"
	hrdb "github.com/integrione/erp-retail/backend/services/human_resource/db"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	"github.com/integrione/erp-retail/backend/services/human_resource/infrastructure/security"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/recruitment"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository"
	handler "github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/recruitment"
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

type candidateIntegrationPolicy struct{}

type requisitionIntegrationVerifier struct{}

func (requisitionIntegrationVerifier) VerifyPositionReferences(_ context.Context, check security.PositionReferenceCheck) error {
	if check.OrganizationUnitID == "" || check.LegalEntityID == "" || check.ValidFrom.IsZero() {
		return hrerrors.ErrReferenceConflict
	}
	return nil
}

func (candidateIntegrationPolicy) Authorize(context.Context, operation.Access) (operation.Actor, error) {
	return operation.Actor{PrincipalID: "00000000-0000-4000-8000-000000000001", SessionID: "00000000-0000-4000-8000-000000000002", Kind: "human"}, nil
}

func (candidateIntegrationPolicy) ReadScope(context.Context, operation.Access) (operation.Scope, error) {
	return operation.Scope{All: true}, nil
}

func TestRecruitmentDraftRPCWithPostgreSQL(t *testing.T) {
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
	validators, err := security.RecruitmentCandidateValidators(repos)
	if err != nil {
		t.Fatal(err)
	}
	applicationValidators, err := security.RecruitmentApplicationValidators(repos)
	if err != nil {
		t.Fatal(err)
	}
	for resource, validator := range applicationValidators {
		validators[resource] = validator
	}
	requisitionValidators, err := security.RecruitmentRequisitionValidators(repos, requisitionIntegrationVerifier{})
	if err != nil {
		t.Fatal(err)
	}
	for resource, validator := range requisitionValidators {
		validators[resource] = validator
	}
	interviewValidators, err := security.RecruitmentInterviewValidators(repos)
	if err != nil {
		t.Fatal(err)
	}
	for resource, validator := range interviewValidators {
		validators[resource] = validator
	}
	svc, err := service.NewService(service.Dependencies{
		Runtime: operation.Dependencies{Transactor: repos.Transactor, Outbox: repos.OutboxEvent,
			Policy: candidateIntegrationPolicy{}, Integrity: security.NewIntegrity(validators)},
		Requisition: repos.Requisition, Candidate: repos.Candidate, Application: repos.Application,
		ApplicationStageHistory: repos.ApplicationStageHistory, Interview: repos.Interview,
		InterviewPanelist: repos.InterviewPanelist, InterviewFeedback: repos.InterviewFeedback,
		Offer: repos.Offer, JobPosting: repos.JobPosting, ApplicationAssessment: repos.ApplicationAssessment,
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
	t.Cleanup(func() { server.Stop(); _ = listener.Close(); <-served })
	connection, err := grpc.NewClient("passthrough:///hrm-candidate-integration",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return listener.DialContext(ctx) }))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = connection.Close() })
	client := pb.NewRecruitmentServiceClient(connection)
	ack := timestamppb.New(time.Now().Add(-time.Minute))
	retention := time.Now().AddDate(0, 1, 0).Format("2006-01-02")
	name := "Candidate A"
	create := func(number string) (*pb.CandidateResponse, error) {
		return client.CreateCandidate(ctx, &pb.CreateCandidateRequest{Input: &pb.CandidateInput{
			CandidateNumber: &number, PreferredName: &name, PrivacyNoticeVersion: text("v1"),
			NoticeAcknowledgedAt: ack, DataRetentionUntil: &retention,
		}})
	}
	created, err := create("CAND-DRAFT-1")
	if err != nil || created.GetItem().GetId() == "" {
		t.Fatalf("create candidate: %v", err)
	}
	id := created.GetItem().GetId()
	if created.GetItem().GetCreatedByPrincipalId() != "00000000-0000-4000-8000-000000000001" {
		t.Fatal("candidate creator did not come from server actor")
	}
	if _, err := create("CAND-DRAFT-1"); status.Code(err) != codes.AlreadyExists {
		t.Fatalf("duplicate candidate number: %v", err)
	}
	badNumber := "CAND-DRAFT-BAD"
	_, err = client.CreateCandidate(ctx, &pb.CreateCandidateRequest{Input: &pb.CandidateInput{
		CandidateNumber: &badNumber, PrivacyNoticeVersion: text("v1"), DataRetentionUntil: &retention,
	}})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("candidate without notice acknowledgement: %v", err)
	}
	_, err = client.CreateCandidate(ctx, &pb.CreateCandidateRequest{Input: &pb.CandidateInput{
		CandidateNumber: &badNumber, PrivacyNoticeVersion: text("v1"),
		NoticeAcknowledgedAt: ack, DataRetentionUntil: &retention,
		ContactReferenceId: text("00000000-0000-4000-8000-000000000003"),
	}})
	if status.Code(err) != codes.Unavailable {
		t.Fatalf("unverified Foundation contact reference: %v", err)
	}
	page, err := client.GetAllCandidates(ctx, &common.PageRequest{Limit: 1})
	if err != nil || page.GetPage().GetTotal() != 1 || len(page.GetItems()) != 1 {
		t.Fatalf("candidate pagination: %v", err)
	}
	corrected := "Corrected candidate name"
	updated, err := client.UpdateCandidate(ctx, &pb.UpdateCandidateRequest{Id: id,
		Input:      &pb.CandidateInput{PreferredName: &corrected},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"preferred_name"}}})
	if err != nil || updated.GetItem().GetPreferredName() != corrected {
		t.Fatalf("candidate name correction: %v", err)
	}
	_, err = client.UpdateCandidate(ctx, &pb.UpdateCandidateRequest{Id: id,
		Input:      &pb.CandidateInput{PrivacyNoticeVersion: text("v2")},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"privacy_notice_version"}}})
	if status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("notice version changed through CRUD: %v", err)
	}
	var outboxCount int64
	if err := database.WithContext(ctx).Table("hr_outbox_events").Where("aggregate_id = ?", id).Count(&outboxCount).Error; err != nil || outboxCount != 2 {
		t.Fatalf("candidate outbox count = %d, err = %v", outboxCount, err)
	}
	if _, err := client.DeleteCandidate(ctx, &common.IDRequest{Id: id}); err != nil {
		t.Fatalf("delete unlinked candidate: %v", err)
	}
	linkedCandidate, err := create("CAND-DRAFT-2")
	if err != nil {
		t.Fatalf("create linked candidate: %v", err)
	}
	const familyID = "00000000-0000-4000-8000-000000000010"
	const jobID = "00000000-0000-4000-8000-000000000011"
	const requisitionID = "00000000-0000-4000-8000-000000000012"
	if err := database.WithContext(ctx).Exec(`INSERT INTO hr_job_families
		(id, family_code, family_name, job_domain) VALUES (?, 'RECRUIT-DRAFT', 'Recruitment', 'HRM')`, familyID).Error; err != nil {
		t.Fatal(err)
	}
	if err := database.WithContext(ctx).Exec(`INSERT INTO hr_jobs
		(id, job_family_id, job_code, job_title, valid_from)
		VALUES (?, ?, 'RECRUIT-DRAFT', 'Recruiter', CURRENT_DATE - 1)`, jobID, familyID).Error; err != nil {
		t.Fatal(err)
	}
	const managerWorkerID = "00000000-0000-4000-8000-000000000013"
	const managerPrincipalID = "00000000-0000-4000-8000-000000000001"
	if err := database.WithContext(ctx).Exec(`INSERT INTO hr_workers
		(id, worker_number, person_party_reference_id, worker_category, status,
		 primary_legal_entity_reference_id, iam_principal_reference_id)
		VALUES (?, 'WORKER-RECRUIT-MANAGER', ?, 'employee', 'active', ?, ?)`,
		managerWorkerID, familyID, familyID, managerPrincipalID).Error; err != nil {
		t.Fatal(err)
	}
	draftNumber, draftKey, draftType, category := "REQ-DRAFT-2", "REQ-DRAFT-2", "new_headcount", "employee"
	draftInput := &pb.RequisitionInput{RequisitionNumber: &draftNumber, JobId: text(jobID),
		OrganizationUnitReferenceId: text(familyID), HiringManagerPrincipalId: text(managerPrincipalID),
		RequisitionType: &draftType, EmploymentCategory: &category, IdempotencyKey: &draftKey}
	draft, err := client.CreateRequisition(ctx, &pb.CreateRequisitionRequest{Input: draftInput})
	if err != nil || draft.GetItem().GetStatus() != "draft" || draft.GetItem().GetId() == "" {
		t.Fatalf("create requisition draft: %+v, %v", draft, err)
	}
	draftID := draft.GetItem().GetId()
	var creatorID string
	if err := database.WithContext(ctx).Raw("SELECT created_by_principal_id FROM hr_requisitions WHERE id = ?", draftID).Scan(&creatorID).Error; err != nil || creatorID != "00000000-0000-4000-8000-000000000001" {
		t.Fatalf("requisition creator = %q, error %v", creatorID, err)
	}
	if err := database.WithContext(ctx).Exec("UPDATE hr_requisitions SET submitted_at = NOW() WHERE id = ?", draftID).Error; err == nil {
		t.Fatal("submission timestamp without principal was accepted")
	}
	if err := database.WithContext(ctx).Exec("UPDATE hr_requisitions SET approved_by_principal_id = ? WHERE id = ?", creatorID, draftID).Error; err == nil {
		t.Fatal("self approval was accepted")
	}
	if _, err := client.CreateRequisition(ctx, &pb.CreateRequisitionRequest{Input: draftInput}); status.Code(err) != codes.AlreadyExists {
		t.Fatalf("duplicate requisition status = %v, want AlreadyExists", status.Code(err))
	}
	if _, err := client.UpdateRequisition(ctx, &pb.UpdateRequisitionRequest{Id: draftID,
		Input:      &pb.RequisitionInput{Status: text("open")},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"status"}}}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("open requisition through CRUD status = %v, want FailedPrecondition", status.Code(err))
	}
	openedAt := timestamppb.New(time.Now())
	if _, err := client.UpdateRequisition(ctx, &pb.UpdateRequisitionRequest{Id: draftID,
		Input:      &pb.RequisitionInput{OpenedAt: openedAt},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"opened_at"}}}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("open requisition through CRUD timestamp = %v, want FailedPrecondition", status.Code(err))
	}
	updatedHeadcount := int32(2)
	updatedDraft, err := client.UpdateRequisition(ctx, &pb.UpdateRequisitionRequest{Id: draftID,
		Input:      &pb.RequisitionInput{Headcount: &updatedHeadcount},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"headcount"}}})
	if err != nil || updatedDraft.GetItem().GetHeadcount() != updatedHeadcount {
		t.Fatalf("update requisition draft: %+v, %v", updatedDraft, err)
	}
	draftPage, err := client.GetAllRequisitions(ctx, &common.PageRequest{Limit: 1})
	if err != nil || draftPage.GetPage().GetTotal() != 1 || len(draftPage.GetItems()) != 1 {
		t.Fatalf("requisition draft pagination: %+v, %v", draftPage, err)
	}
	if _, err := client.DeleteRequisition(ctx, &common.IDRequest{Id: draftID}); err != nil {
		t.Fatalf("delete requisition draft: %v", err)
	}
	var requisitionOutboxCount int64
	if err := database.WithContext(ctx).Table("hr_outbox_events").Where("aggregate_id = ?", draftID).Count(&requisitionOutboxCount).Error; err != nil || requisitionOutboxCount != 3 {
		t.Fatalf("requisition outbox count = %d, err = %v", requisitionOutboxCount, err)
	}
	if err := database.WithContext(ctx).Exec(`INSERT INTO hr_requisitions
		(id, requisition_number, job_id, organization_unit_reference_id, hiring_manager_principal_id,
		requisition_type, status, employment_category, approved_at, approved_by_principal_id,
		opened_at, idempotency_key)
		VALUES (?, 'REQ-DRAFT-1', ?, ?, ?, 'replacement', 'open', 'employee',
		NOW() - INTERVAL '2 hours', ?, NOW() - INTERVAL '1 hour', 'REQ-DRAFT-1')`,
		requisitionID, jobID, familyID, familyID, familyID).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := client.UpdateRequisition(ctx, &pb.UpdateRequisitionRequest{Id: requisitionID,
		Input:      &pb.RequisitionInput{Headcount: &updatedHeadcount},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"headcount"}}}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("edit open requisition through CRUD = %v", err)
	}
	if _, err := client.DeleteRequisition(ctx, &common.IDRequest{Id: requisitionID}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("delete open requisition through CRUD = %v", err)
	}
	applicationNumber := "APP-DRAFT-1"
	idempotencyKey := "APP-DRAFT-1"
	applicationInput := &pb.ApplicationInput{CandidateId: text(linkedCandidate.GetItem().GetId()),
		RequisitionId: text(requisitionID), ApplicationNumber: &applicationNumber, IdempotencyKey: &idempotencyKey}
	application, err := client.CreateApplication(ctx, &pb.CreateApplicationRequest{Input: applicationInput})
	if err != nil || application.GetItem().GetStatus() != "received" {
		t.Fatalf("create received application: %v", err)
	}
	applicationID := application.GetItem().GetId()
	duplicateNumber := "APP-DRAFT-2"
	duplicateKey := "APP-DRAFT-2"
	_, err = client.CreateApplication(ctx, &pb.CreateApplicationRequest{Input: &pb.ApplicationInput{
		CandidateId: applicationInput.CandidateId, RequisitionId: applicationInput.RequisitionId,
		ApplicationNumber: &duplicateNumber, IdempotencyKey: &duplicateKey}})
	if status.Code(err) != codes.AlreadyExists {
		t.Fatalf("duplicate candidate and requisition: %v", err)
	}
	applicationPage, err := client.GetAllApplications(ctx, &common.PageRequest{Limit: 1})
	if err != nil || applicationPage.GetPage().GetTotal() != 1 || len(applicationPage.GetItems()) != 1 {
		t.Fatalf("application pagination: %v", err)
	}
	_, err = client.UpdateApplication(ctx, &pb.UpdateApplicationRequest{Id: applicationID,
		Input:      &pb.ApplicationInput{Status: text("screening")},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"status"}}})
	if status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("screening via CRUD: %v", err)
	}
	if err := database.WithContext(ctx).Table("hr_candidates").Where("id = ?", linkedCandidate.GetItem().GetId()).Update("candidate_status", "withdrawn").Error; err != nil {
		t.Fatal(err)
	}
	if _, err := client.StartApplicationScreening(ctx, &common.IDRequest{Id: applicationID}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("screening of withdrawn candidate status = %v, want FailedPrecondition", status.Code(err))
	}
	var stageCount int64
	if err := database.WithContext(ctx).Table("hr_application_stage_history").Where("application_id = ?", applicationID).Count(&stageCount).Error; err != nil || stageCount != 0 {
		t.Fatalf("rejected screening left %d history rows: %v", stageCount, err)
	}
	if err := database.WithContext(ctx).Table("hr_candidates").Where("id = ?", linkedCandidate.GetItem().GetId()).Update("candidate_status", "active").Error; err != nil {
		t.Fatal(err)
	}
	screening, err := client.StartApplicationScreening(ctx, &common.IDRequest{Id: applicationID})
	if err != nil || screening.GetItem().GetStatus() != "screening" ||
		screening.GetItem().GetCurrentStageCode() != "screening" ||
		screening.GetItem().GetAssignedRecruiterPrincipalId() != "00000000-0000-4000-8000-000000000001" {
		t.Fatalf("start application screening: %+v, %v", screening, err)
	}
	history, err := client.FindApplicationStageHistoryByApplicationID(ctx, &common.StringFinderRequest{Value: &applicationID, Limit: 10})
	if err != nil || len(history.GetItems()) != 1 || history.GetItems()[0].GetPreviousStatus() != "received" ||
		history.GetItems()[0].GetNewStatus() != "screening" || history.GetItems()[0].GetStageCode() != "screening" ||
		history.GetItems()[0].GetChangedByPrincipalId() != "00000000-0000-4000-8000-000000000001" {
		t.Fatalf("application stage history: %+v, %v", history, err)
	}
	if _, err := client.StartApplicationScreening(ctx, &common.IDRequest{Id: applicationID}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("repeated screening status = %v, want FailedPrecondition", status.Code(err))
	}
	if _, err := client.DeleteApplication(ctx, &common.IDRequest{Id: applicationID}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("delete application without retention workflow: %v", err)
	}
	if _, err := client.DeleteCandidate(ctx, &common.IDRequest{Id: linkedCandidate.GetItem().GetId()}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("delete linked candidate: %v", err)
	}
	if err := database.WithContext(ctx).Table("hr_outbox_events").Where("aggregate_id = ?", applicationID).Count(&outboxCount).Error; err != nil || outboxCount != 2 {
		t.Fatalf("application outbox count = %d, err = %v", outboxCount, err)
	}
	interviewType, timezone := "screening", "Asia/Jakarta"
	round := int32(1)
	start := timestamppb.New(time.Now().Add(48 * time.Hour))
	end := timestamppb.New(start.AsTime().Add(time.Hour))
	createInterview := &pb.InterviewInput{ApplicationId: &applicationID, InterviewRound: &round,
		InterviewType: &interviewType, ScheduledStart: start, ScheduledEnd: end, CandidateTimezone: &timezone}
	createdInterview, err := client.CreateInterview(ctx, &pb.CreateInterviewRequest{Input: createInterview})
	if err != nil || createdInterview.GetItem().GetId() == "" || createdInterview.GetItem().GetStatus() != "planned" {
		t.Fatalf("create planned interview: %+v, %v", createdInterview, err)
	}
	interviewID := createdInterview.GetItem().GetId()
	if createdInterview.GetItem().GetCreatedByPrincipalId() != "00000000-0000-4000-8000-000000000001" {
		t.Fatal("interview creator did not come from server actor")
	}
	if _, err := client.CreateInterview(ctx, &pb.CreateInterviewRequest{Input: createInterview}); status.Code(err) != codes.AlreadyExists {
		t.Fatalf("duplicate interview round status = %v, want AlreadyExists", status.Code(err))
	}
	otherRound := int32(2)
	invalidTimezone := "Invalid/Timezone"
	if _, err := client.CreateInterview(ctx, &pb.CreateInterviewRequest{Input: &pb.InterviewInput{
		ApplicationId: &applicationID, InterviewRound: &otherRound, InterviewType: &interviewType,
		ScheduledStart: start, ScheduledEnd: end, CandidateTimezone: &invalidTimezone,
	}}); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("invalid timezone status = %v, want InvalidArgument", status.Code(err))
	}
	shortRetention := time.Now().Add(24 * time.Hour).Format("2006-01-02")
	if err := database.WithContext(ctx).Exec("UPDATE hr_candidates SET data_retention_until = ? WHERE id = ?",
		shortRetention, linkedCandidate.GetItem().GetId()).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := client.CreateInterview(ctx, &pb.CreateInterviewRequest{Input: &pb.InterviewInput{
		ApplicationId: &applicationID, InterviewRound: &otherRound, InterviewType: &interviewType,
		ScheduledStart: start, ScheduledEnd: end, CandidateTimezone: &timezone,
	}}); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("interview beyond candidate retention status = %v, want InvalidArgument", status.Code(err))
	}
	if err := database.WithContext(ctx).Exec("UPDATE hr_candidates SET data_retention_until = ? WHERE id = ?",
		retention, linkedCandidate.GetItem().GetId()).Error; err != nil {
		t.Fatal(err)
	}
	locationID := "00000000-0000-4000-8000-000000000015"
	if _, err := client.CreateInterview(ctx, &pb.CreateInterviewRequest{Input: &pb.InterviewInput{
		ApplicationId: &applicationID, InterviewRound: &otherRound, InterviewType: &interviewType,
		LocationReferenceId: &locationID,
	}}); status.Code(err) != codes.Unavailable {
		t.Fatalf("unverified location status = %v, want Unavailable", status.Code(err))
	}
	interviewPage, err := client.FindInterviewByApplicationID(ctx, &common.StringFinderRequest{Value: &applicationID, Limit: 1})
	if err != nil || interviewPage.GetPage().GetTotal() != 1 || len(interviewPage.GetItems()) != 1 {
		t.Fatalf("interview finder: %+v, %v", interviewPage, err)
	}
	interviewPage, err = client.GetAllInterviews(ctx, &common.PageRequest{Limit: 1, Offset: 1})
	if err != nil || interviewPage.GetPage().GetTotal() != 1 || len(interviewPage.GetItems()) != 0 {
		t.Fatalf("interview offset: %+v, %v", interviewPage, err)
	}
	newStart := timestamppb.New(start.AsTime().Add(30 * time.Minute))
	updatedInterview, err := client.UpdateInterview(ctx, &pb.UpdateInterviewRequest{Id: interviewID,
		Input:      &pb.InterviewInput{ScheduledStart: newStart},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"scheduled_start"}}})
	if err != nil || !updatedInterview.GetItem().GetScheduledStart().AsTime().Equal(newStart.AsTime()) {
		t.Fatalf("reschedule planned interview: %+v, %v", updatedInterview, err)
	}
	if _, err := client.UpdateInterview(ctx, &pb.UpdateInterviewRequest{Id: interviewID,
		Input:      &pb.InterviewInput{Status: text("scheduled")},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"status"}}}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("direct interview status change = %v, want FailedPrecondition", status.Code(err))
	}
	if _, err := client.CreateInterviewPanelist(ctx, &pb.CreateInterviewPanelistRequest{Input: &pb.InterviewPanelistInput{
		InterviewId: &interviewID, InterviewerPrincipalId: text("00000000-0000-4000-8000-000000000001"),
	}}); status.Code(err) != codes.Unavailable {
		t.Fatalf("panelist without verifier status = %v, want Unavailable", status.Code(err))
	}
	if err := database.WithContext(ctx).Table("hr_outbox_events").Where("aggregate_id = ?", interviewID).Count(&outboxCount).Error; err != nil || outboxCount != 2 {
		t.Fatalf("interview outbox count before delete = %d, err = %v", outboxCount, err)
	}
	panelistID := "00000000-0000-4000-8000-000000000016"
	if err := database.WithContext(ctx).Exec(`INSERT INTO hr_interview_panelists
		(id, interview_id, interviewer_principal_id) VALUES (?, ?, ?)`,
		panelistID, interviewID, "00000000-0000-4000-8000-000000000001").Error; err != nil {
		t.Fatal(err)
	}
	if _, err := client.UpdateInterview(ctx, &pb.UpdateInterviewRequest{Id: interviewID,
		Input:      &pb.InterviewInput{ScheduledStart: start},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"scheduled_start"}}}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("update interview with panelist status = %v, want FailedPrecondition", status.Code(err))
	}
	if _, err := client.DeleteInterview(ctx, &common.IDRequest{Id: interviewID}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("delete interview with panelist status = %v, want FailedPrecondition", status.Code(err))
	}
	if err := database.WithContext(ctx).Exec("DELETE FROM hr_interview_panelists WHERE id = ?", panelistID).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := client.DeleteInterview(ctx, &common.IDRequest{Id: interviewID}); err != nil {
		t.Fatalf("delete unreferenced planned interview: %v", err)
	}
	if err := database.WithContext(ctx).Table("hr_outbox_events").Where("aggregate_id = ?", interviewID).Count(&outboxCount).Error; err != nil || outboxCount != 3 {
		t.Fatalf("interview outbox count after delete = %d, err = %v", outboxCount, err)
	}
}
