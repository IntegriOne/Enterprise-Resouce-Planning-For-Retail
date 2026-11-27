//go:build integration

package recruitment_test

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
	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/recruitment/v1"
	workforcepb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/workforceplanning/v1"
	hrdb "github.com/integrione/erp-retail/backend/services/human_resource/db"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	"github.com/integrione/erp-retail/backend/services/human_resource/infrastructure/security"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/recruitment"
	workforceservice "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/workforceplanning"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository"
	handler "github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/recruitment"
	workforcehandler "github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/workforceplanning"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type candidateIntegrationPolicy struct{ principalID string }

type requisitionIntegrationVerifier struct{}

type requisitionApprovalIntegrationVerifier struct {
	mu       sync.Mutex
	budgetID string
	check    service.RequisitionApprovalCheck
}

type postingDocumentIntegrationVerifier struct{}

func (postingDocumentIntegrationVerifier) VerifyJobPostingDocument(_ context.Context, request service.PostingDocumentRequest) (service.PostingDocumentProof, error) {
	if request.PostingID == "" || request.RequisitionID == "" || request.PrincipalID == "" || len(request.Digest) != 32 {
		return service.PostingDocumentProof{}, hrerrors.ErrReferenceConflict
	}
	return service.PostingDocumentProof{
		VersionID: request.VersionID, Digest: append([]byte(nil), request.Digest...),
		Verified: true, Immutable: true, OwnerManagement: "human_resource", Purpose: "recruitment_job_posting",
		PostingID: request.PostingID, RequisitionID: request.RequisitionID,
	}, nil
}

func (v *requisitionApprovalIntegrationVerifier) VerifyRequisitionMandate(context.Context, string, string) error {
	return nil
}

func (v *requisitionApprovalIntegrationVerifier) VerifyRequisitionApproval(_ context.Context, check service.RequisitionApprovalCheck) (string, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.check = check
	if check.RequisitionID == "" || check.Headcount <= 0 || check.CreatedByPrincipalID == check.ApproverPrincipalID {
		return "", hrerrors.ErrReferenceConflict
	}
	return v.budgetID, nil
}

func (requisitionIntegrationVerifier) VerifyPositionReferences(_ context.Context, check security.PositionReferenceCheck) error {
	if check.OrganizationUnitID == "" || check.LegalEntityID == "" || check.ValidFrom.IsZero() {
		return hrerrors.ErrReferenceConflict
	}
	return nil
}

func (p *candidateIntegrationPolicy) Authorize(context.Context, operation.Access) (operation.Actor, error) {
	return operation.Actor{PrincipalID: p.principalID, SessionID: "00000000-0000-4000-8000-000000000002", Kind: "human"}, nil
}

func (*candidateIntegrationPolicy) ReadScope(context.Context, operation.Access) (operation.Scope, error) {
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
	validators, err := security.RecruitmentCandidateValidators(repos, recDetailFoundationStub{})
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
	postingValidators, err := security.RecruitmentJobPostingValidators(repos)
	if err != nil {
		t.Fatal(err)
	}
	for resource, validator := range postingValidators {
		validators[resource] = validator
	}
	interviewValidators, err := security.RecruitmentInterviewValidators(repos)
	if err != nil {
		t.Fatal(err)
	}
	for resource, validator := range interviewValidators {
		validators[resource] = validator
	}
	policy := &candidateIntegrationPolicy{principalID: "00000000-0000-4000-8000-000000000001"}
	approvalVerifier := &requisitionApprovalIntegrationVerifier{budgetID: "00000000-0000-4000-8000-000000000099"}
	svc, err := service.NewService(service.Dependencies{
		Runtime: operation.Dependencies{Transactor: repos.Transactor, Outbox: repos.OutboxEvent,
			Policy: policy, Integrity: security.NewIntegrity(validators)},
		ApprovalVerifier: approvalVerifier,
		MandateVerifier:  approvalVerifier,
		PostingVerifier:  postingDocumentIntegrationVerifier{},
		Requisition:      repos.Requisition, Candidate: repos.Candidate, Application: repos.Application, Workers: repos.Worker,
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
	workforceValidators, err := security.WorkforcePlanValidators(repos, nil)
	if err != nil {
		t.Fatal(err)
	}
	workforce, err := workforceservice.NewService(workforceservice.Dependencies{
		Runtime: operation.Dependencies{Transactor: repos.Transactor, Outbox: repos.OutboxEvent,
			Policy: policy, Integrity: security.NewIntegrity(workforceValidators)},
		JobFamily: repos.JobFamily, Job: repos.Job, Position: repos.Position,
		WorkforcePlan: repos.WorkforcePlan, WorkforcePlanLine: repos.WorkforcePlanLine,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := workforcehandler.Register(server, workforce); err != nil {
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
	partyID := "00000000-0000-4000-8000-000000000003"
	contactID := "00000000-0000-4000-8000-000000000004"
	create := func(number string) (*pb.CandidateResponse, error) {
		return client.CreateCandidate(ctx, &pb.CreateCandidateRequest{Input: &pb.CandidateInput{
			CandidateNumber: &number, PreferredName: &name, PrivacyNoticeVersion: text("v1"),
			PersonPartyReferenceId: &partyID, ContactReferenceId: &contactID,
			NoticeAcknowledgedAt: ack, DataRetentionUntil: &retention,
		}})
	}
	created, err := create("CAND-DRAFT-1")
	if err != nil || created.GetItem().GetId() == "" || created.GetItem().GetPersonPartyReferenceId() != partyID ||
		created.GetItem().GetContactReferenceId() != contactID {
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
	if status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("Foundation contact without a party: %v", err)
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
	if _, err := client.CreateRequisition(ctx, &pb.CreateRequisitionRequest{Input: draftInput, HiringManagerWorkerId: text(managerWorkerID)}); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("create requisition with both manager references: %v", err)
	}
	workerInput := proto.CloneOf(draftInput)
	workerInput.HiringManagerPrincipalId = nil
	draft, err := client.CreateRequisition(ctx, &pb.CreateRequisitionRequest{Input: workerInput, HiringManagerWorkerId: text(managerWorkerID)})
	if err != nil || draft.GetItem().GetStatus() != "draft" || draft.GetItem().GetId() == "" ||
		draft.GetItem().GetHiringManagerPrincipalId() != managerPrincipalID {
		t.Fatalf("create requisition draft: %+v, %v", draft, err)
	}
	draftID := draft.GetItem().GetId()
	managerMask := &fieldmaskpb.FieldMask{Paths: []string{"hiring_manager_worker_id"}}
	managerUpdate := &pb.UpdateRequisitionRequest{Id: draftID, Input: &pb.RequisitionInput{}, UpdateMask: managerMask, HiringManagerWorkerId: text(managerWorkerID)}
	changedManager, err := client.UpdateRequisition(ctx, managerUpdate)
	if err != nil || changedManager.GetItem().GetHiringManagerPrincipalId() != managerPrincipalID {
		t.Fatalf("update hiring manager with worker ID: %+v, %v", changedManager, err)
	}
	managerDetail, err := client.GetRequisitionByID(ctx, &common.IDRequest{Id: draftID})
	if err != nil || managerDetail.GetItem().GetHiringManagerWorkerId() != managerWorkerID {
		t.Fatalf("read hiring manager worker reference: %+v, %v", managerDetail, err)
	}
	conflictingManager := proto.CloneOf(managerUpdate)
	conflictingManager.Input.HiringManagerPrincipalId = text(managerPrincipalID)
	if _, err := client.UpdateRequisition(ctx, conflictingManager); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("update with conflicting hiring manager references: %v", err)
	}
	if _, err := client.UpdateRequisition(ctx, &pb.UpdateRequisitionRequest{Id: draftID,
		Input:      &pb.RequisitionInput{HiringManagerPrincipalId: text("00000000-0000-4000-8000-000000000099")},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"hiring_manager_principal_id"}}}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("replace hiring manager by direct principal: %v", err)
	}
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
	if err := database.WithContext(ctx).Table("hr_outbox_events").Where("aggregate_id = ?", draftID).Count(&requisitionOutboxCount).Error; err != nil || requisitionOutboxCount != 4 {
		t.Fatalf("requisition outbox count = %d, err = %v", requisitionOutboxCount, err)
	}
	const workforcePlanID = "00000000-0000-4000-8000-000000000211"
	const workforceLineID = "00000000-0000-4000-8000-000000000212"
	if err := database.WithContext(ctx).Exec(`INSERT INTO hr_workforce_plans
		(id, plan_code, plan_version, legal_entity_reference_id, period_start, period_end, status, created_by_principal_id,
		 approved_at, approved_by_principal_id, finance_budget_reference_id)
		VALUES (?, 'RECRUIT-PLAN', 1, ?, CURRENT_DATE - 1, CURRENT_DATE + 365, 'approved', ?, NOW(), ?, ?)`,
		workforcePlanID, familyID, managerPrincipalID, familyID, familyID).Error; err != nil {
		t.Fatal(err)
	}
	if err := database.WithContext(ctx).Exec(`INSERT INTO hr_workforce_plan_lines
		(id, workforce_plan_id, line_number, organization_unit_reference_id, job_id, worker_category,
		 current_fte, required_fte, needed_from, action_code, rationale)
		VALUES (?, ?, 1, ?, ?, 'employee', 0, 1, CURRENT_DATE - 1, 'recruit', 'Recruitment need')`,
		workforceLineID, workforcePlanID, familyID, jobID).Error; err != nil {
		t.Fatal(err)
	}
	submittedNumber, submittedKey := "REQ-SUBMITTED-1", "REQ-SUBMITTED-1"
	submittedInput := &pb.RequisitionInput{RequisitionNumber: &submittedNumber, JobId: text(jobID),
		OrganizationUnitReferenceId: text(familyID), HiringManagerPrincipalId: text(managerPrincipalID),
		RequisitionType: &draftType, EmploymentCategory: &category, IdempotencyKey: &submittedKey, WorkforcePlanLineId: text(workforceLineID)}
	submissionDraft, err := client.CreateRequisition(ctx, &pb.CreateRequisitionRequest{Input: submittedInput})
	if err != nil {
		t.Fatalf("create requisition for submission: %v", err)
	}
	submittedID := submissionDraft.GetItem().GetId()
	submitted, err := client.SubmitRequisition(ctx, &common.IDRequest{Id: submittedID})
	if err != nil || submitted.GetItem().GetStatus() != "submitted" || submitted.GetItem().GetSubmittedAt() == nil ||
		submitted.GetItem().GetSubmittedByPrincipalId() != managerPrincipalID || submitted.GetItem().GetCreatedByPrincipalId() != managerPrincipalID {
		t.Fatalf("submit requisition: %+v, %v", submitted, err)
	}
	if _, err := client.SubmitRequisition(ctx, &common.IDRequest{Id: submittedID}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("repeat requisition submission = %v", status.Code(err))
	}
	if _, err := client.UpdateRequisition(ctx, &pb.UpdateRequisitionRequest{Id: submittedID,
		Input:      &pb.RequisitionInput{Headcount: &updatedHeadcount},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"headcount"}}}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("edit submitted requisition through CRUD = %v", status.Code(err))
	}
	if _, err := client.DeleteRequisition(ctx, &common.IDRequest{Id: submittedID}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("delete submitted requisition through CRUD = %v", status.Code(err))
	}
	if err := database.WithContext(ctx).Table("hr_outbox_events").Where("aggregate_id = ?", submittedID).Count(&requisitionOutboxCount).Error; err != nil || requisitionOutboxCount != 2 {
		t.Fatalf("submitted requisition outbox count = %d, err = %v", requisitionOutboxCount, err)
	}
	rejectedNumber := "REQ-REJECT-1"
	rejectedKey := "REQ-REJECT-1"
	rejectionDraft, err := client.CreateRequisition(ctx, &pb.CreateRequisitionRequest{Input: &pb.RequisitionInput{
		RequisitionNumber: &rejectedNumber, JobId: text(jobID), OrganizationUnitReferenceId: text(familyID),
		HiringManagerPrincipalId: text(managerPrincipalID), RequisitionType: &draftType,
		EmploymentCategory: &category, IdempotencyKey: &rejectedKey,
	}})
	if err != nil {
		t.Fatalf("create requisition for rejection: %v", err)
	}
	rejectedID := rejectionDraft.GetItem().GetId()
	if _, err := client.SubmitRequisition(ctx, &common.IDRequest{Id: rejectedID}); err != nil {
		t.Fatalf("submit requisition for rejection: %v", err)
	}
	if _, err := client.RejectRequisition(ctx, &pb.RejectRequisitionRequest{Id: rejectedID, ReasonCode: "BUDGET_DENIED"}); status.Code(err) != codes.PermissionDenied {
		t.Fatalf("self rejection status = %v", status.Code(err))
	}
	if _, err := client.ApproveRequisition(ctx, &common.IDRequest{Id: submittedID}); status.Code(err) != codes.PermissionDenied {
		t.Fatalf("self approval status = %v", status.Code(err))
	}
	policy.principalID = "00000000-0000-4000-8000-000000000098"
	if _, err := client.RejectRequisition(ctx, &pb.RejectRequisitionRequest{Id: rejectedID, ReasonCode: " "}); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("blank rejection reason status = %v", status.Code(err))
	}
	rejected, err := client.RejectRequisition(ctx, &pb.RejectRequisitionRequest{Id: rejectedID, ReasonCode: "BUDGET_DENIED"})
	if err != nil || rejected.GetItem().GetStatus() != "rejected" || rejected.GetItem().GetRejectedAt() == nil ||
		rejected.GetItem().GetRejectionReasonCode() != "BUDGET_DENIED" ||
		rejected.GetItem().GetRejectedByPrincipalId() != policy.principalID {
		t.Fatalf("reject requisition: %+v, %v", rejected, err)
	}
	rejectedByReason, err := client.FindRequisitionByRejectionReasonCode(ctx, &common.StringFinderRequest{
		Value: text("BUDGET_DENIED"), Limit: 1,
	})
	if err != nil || rejectedByReason.GetPage().GetTotal() != 1 || len(rejectedByReason.GetItems()) != 1 ||
		rejectedByReason.GetItems()[0].GetId() != rejectedID {
		t.Fatalf("find rejected requisition by reason: %+v, %v", rejectedByReason, err)
	}
	rejectedByActor, err := client.FindRequisitionByRejectedByPrincipalID(ctx, &common.StringFinderRequest{
		Value: text(policy.principalID), Limit: 1,
	})
	if err != nil || rejectedByActor.GetPage().GetTotal() != 1 || len(rejectedByActor.GetItems()) != 1 ||
		rejectedByActor.GetItems()[0].GetId() != rejectedID {
		t.Fatalf("find rejected requisition by actor: %+v, %v", rejectedByActor, err)
	}
	if _, err := client.RejectRequisition(ctx, &pb.RejectRequisitionRequest{Id: rejectedID, ReasonCode: "BUDGET_DENIED"}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("repeat rejection status = %v", status.Code(err))
	}
	if _, err := client.ApproveRequisition(ctx, &common.IDRequest{Id: rejectedID}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("approval after rejection status = %v", status.Code(err))
	}
	var rejectedOutboxCount int64
	if err := database.WithContext(ctx).Table("hr_outbox_events").Where("aggregate_id = ?", rejectedID).Count(&rejectedOutboxCount).Error; err != nil || rejectedOutboxCount != 3 {
		t.Fatalf("rejected requisition outbox count = %d, err = %v", rejectedOutboxCount, err)
	}
	if err := database.WithContext(ctx).Exec("UPDATE hr_requisitions SET rejected_by_principal_id = NULL WHERE id = ?", rejectedID).Error; err == nil {
		t.Fatal("database accepted rejected requisition without decision actor")
	}
	if err := database.WithContext(ctx).Exec("UPDATE hr_requisitions SET rejection_reason_code = NULL WHERE id = ?", rejectedID).Error; err == nil {
		t.Fatal("database accepted rejected requisition without reason")
	}
	approved, err := client.ApproveRequisition(ctx, &common.IDRequest{Id: submittedID})
	if err != nil || approved.GetItem().GetStatus() != "approved" || approved.GetItem().GetApprovedAt() == nil ||
		approved.GetItem().GetApprovedByPrincipalId() != policy.principalID ||
		approved.GetItem().GetBudgetReferenceId() != approvalVerifier.budgetID ||
		approvalVerifier.check.RequisitionID != submittedID || approvalVerifier.check.Headcount != 1 ||
		approvalVerifier.check.WorkforcePlanLineID == nil || *approvalVerifier.check.WorkforcePlanLineID != workforceLineID {
		t.Fatalf("approve requisition: %+v, %v, check %+v", approved, err, approvalVerifier.check)
	}
	if _, err := client.ApproveRequisition(ctx, &common.IDRequest{Id: submittedID}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("repeat approval status = %v", status.Code(err))
	}
	policy.principalID = managerPrincipalID
	if err := database.WithContext(ctx).Table("hr_outbox_events").Where("aggregate_id = ?", submittedID).Count(&requisitionOutboxCount).Error; err != nil || requisitionOutboxCount != 3 {
		t.Fatalf("approved requisition outbox count = %d, err = %v", requisitionOutboxCount, err)
	}
	opened, err := client.OpenRequisition(ctx, &common.IDRequest{Id: submittedID})
	if err != nil || opened.GetItem().GetStatus() != "open" || opened.GetItem().GetOpenedAt() == nil ||
		opened.GetItem().GetBudgetReferenceId() != approvalVerifier.budgetID {
		t.Fatalf("open requisition: %+v, %v", opened, err)
	}
	if _, err := client.OpenRequisition(ctx, &common.IDRequest{Id: submittedID}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("repeat opening status = %v", status.Code(err))
	}
	if err := database.WithContext(ctx).Table("hr_outbox_events").Where("aggregate_id = ?", submittedID).Count(&requisitionOutboxCount).Error; err != nil || requisitionOutboxCount != 4 {
		t.Fatalf("opened requisition outbox count = %d, err = %v", requisitionOutboxCount, err)
	}
	staleNumber, staleKey := "REQ-STALE-WORKFORCE", "REQ-STALE-WORKFORCE"
	staleDraft, err := client.CreateRequisition(ctx, &pb.CreateRequisitionRequest{Input: &pb.RequisitionInput{
		RequisitionNumber: &staleNumber, IdempotencyKey: &staleKey, JobId: text(jobID),
		OrganizationUnitReferenceId: text(familyID), HiringManagerPrincipalId: text(managerPrincipalID),
		RequisitionType: &draftType, EmploymentCategory: &category, WorkforcePlanLineId: text(workforceLineID),
	}})
	if err != nil {
		t.Fatalf("create draft before workforce replacement: %v", err)
	}
	const replacementPlanID = "00000000-0000-4000-8000-000000000213"
	if err := database.WithContext(ctx).Exec(`INSERT INTO hr_workforce_plans
		(id, plan_code, plan_version, legal_entity_reference_id, period_start, period_end, scenario_code, status,
		 created_by_principal_id, approved_at, approved_by_principal_id, finance_budget_reference_id)
		SELECT ?, plan_code, 2, legal_entity_reference_id, period_start, period_end, scenario_code, 'approved',
		 created_by_principal_id, approved_at, approved_by_principal_id, finance_budget_reference_id
		FROM hr_workforce_plans WHERE id = ?`, replacementPlanID, workforcePlanID).Error; err != nil {
		t.Fatal(err)
	}
	workforceClient := workforcepb.NewWorkforcePlanningServiceClient(connection)
	if _, err := workforceClient.SupersedeWorkforcePlan(ctx, &workforcepb.SupersedeWorkforcePlanRequest{
		Id: workforcePlanID, ReplacementPlanId: replacementPlanID, ReasonCode: "NEW_VERSION",
	}); err != nil {
		t.Fatalf("supersede originating workforce plan: %v", err)
	}
	staleID := &common.IDRequest{Id: staleDraft.GetItem().GetId()}
	if _, err := client.SubmitRequisition(ctx, staleID); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("submit against superseded plan = %v", err)
	}
	stillDraft, err := client.GetRequisitionByID(ctx, staleID)
	if err != nil || stillDraft.GetItem().GetStatus() != "draft" || stillDraft.GetItem().GetSubmittedAt() != nil {
		t.Fatalf("read requisition after rejected submit: %+v, %v", stillDraft, err)
	}
	if err := database.WithContext(ctx).Table("hr_outbox_events").Where("aggregate_id = ?", staleID.Id).Count(&requisitionOutboxCount).Error; err != nil || requisitionOutboxCount != 1 {
		t.Fatalf("stale requisition outbox count = %d, err = %v", requisitionOutboxCount, err)
	}
	postingVersion := int32(1)
	postingInput := &pb.JobPostingInput{RequisitionId: text(submittedID), ChannelCode: text("career_site"),
		PostingVersion: &postingVersion, ContentDocumentVersionReferenceId: text("00000000-0000-4000-8000-000000000200"),
		ContentDigest: bytes.Repeat([]byte{0x42}, 32)}
	posting, err := client.CreateJobPosting(ctx, &pb.CreateJobPostingRequest{Input: postingInput})
	if err != nil || posting.GetItem().GetStatus() != "draft" || posting.GetItem().GetCreatedByPrincipalId() != policy.principalID {
		t.Fatalf("create draft posting: %+v, %v", posting, err)
	}
	if _, err := client.UpdateJobPosting(ctx, &pb.UpdateJobPostingRequest{Id: posting.GetItem().GetId(),
		Input:      &pb.JobPostingInput{ContentDocumentVersionReferenceId: text("00000000-0000-4000-8000-000000000206")},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"content_document_version_reference_id"}}}); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("change posting content version without digest = %v, want InvalidArgument", err)
	}
	published, err := client.PublishJobPosting(ctx, &common.IDRequest{Id: posting.GetItem().GetId()})
	if err != nil || published.GetItem().GetStatus() != "published" || published.GetItem().GetPublishedAt() == nil {
		t.Fatalf("publish posting: %+v, %v", published, err)
	}
	if _, err := client.PublishJobPosting(ctx, &common.IDRequest{Id: posting.GetItem().GetId()}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("repeat publication status = %v", status.Code(err))
	}
	var postingOutboxCount int64
	if err := database.WithContext(ctx).Table("hr_outbox_events").Where("aggregate_id = ?", posting.GetItem().GetId()).Count(&postingOutboxCount).Error; err != nil || postingOutboxCount != 2 {
		t.Fatalf("published posting outbox count = %d, err = %v", postingOutboxCount, err)
	}
	if _, err := client.UpdateJobPosting(ctx, &pb.UpdateJobPostingRequest{Id: posting.GetItem().GetId(),
		Input:      &pb.JobPostingInput{Status: text("draft")},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"status"}}}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("edit published posting status = %v", status.Code(err))
	}
	if _, err := client.CloseRequisition(ctx, &common.IDRequest{Id: submittedID}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("close with published posting status = %v", status.Code(err))
	}
	closedPosting, err := client.CloseJobPosting(ctx, &common.IDRequest{Id: posting.GetItem().GetId()})
	if err != nil || closedPosting.GetItem().GetStatus() != "closed" || closedPosting.GetItem().GetClosesAt() == nil ||
		!closedPosting.GetItem().GetClosesAt().AsTime().After(published.GetItem().GetPublishedAt().AsTime()) {
		t.Fatalf("close posting: %+v, %v", closedPosting, err)
	}
	if _, err := client.CloseJobPosting(ctx, &common.IDRequest{Id: posting.GetItem().GetId()}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("repeat posting close status = %v", status.Code(err))
	}
	if err := database.WithContext(ctx).Table("hr_outbox_events").Where("aggregate_id = ?", posting.GetItem().GetId()).Count(&postingOutboxCount).Error; err != nil || postingOutboxCount != 3 {
		t.Fatalf("closed posting outbox count = %d, err = %v", postingOutboxCount, err)
	}
	postingInput.ChannelCode = text("external_board")
	withdrawalPosting, err := client.CreateJobPosting(ctx, &pb.CreateJobPostingRequest{Input: postingInput})
	if err != nil {
		t.Fatalf("create withdrawal posting: %v", err)
	}
	if _, err := client.PublishJobPosting(ctx, &common.IDRequest{Id: withdrawalPosting.GetItem().GetId()}); err != nil {
		t.Fatalf("publish withdrawal posting: %v", err)
	}
	if _, err := client.CloseRequisition(ctx, &common.IDRequest{Id: submittedID}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("close with second published posting status = %v", status.Code(err))
	}
	if _, err := client.WithdrawJobPosting(ctx, &pb.WithdrawJobPostingRequest{Id: withdrawalPosting.GetItem().GetId()}); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("withdraw without reason status = %v", status.Code(err))
	}
	withdrawnPosting, err := client.WithdrawJobPosting(ctx, &pb.WithdrawJobPostingRequest{
		Id: withdrawalPosting.GetItem().GetId(), ReasonCode: "DOCUMENT_REVOKED",
	})
	if err != nil || withdrawnPosting.GetItem().GetStatus() != "withdrawn" || withdrawnPosting.GetItem().GetWithdrawnAt() == nil ||
		withdrawnPosting.GetItem().GetWithdrawalReasonCode() != "DOCUMENT_REVOKED" ||
		withdrawnPosting.GetItem().GetWithdrawnByPrincipalId() != policy.principalID {
		t.Fatalf("withdraw posting: %+v, %v", withdrawnPosting, err)
	}
	reasonCode := "DOCUMENT_REVOKED"
	byReason, err := client.FindJobPostingByWithdrawalReasonCode(ctx, &common.StringFinderRequest{Value: &reasonCode, Limit: 1})
	if err != nil || len(byReason.GetItems()) != 1 || byReason.GetItems()[0].GetId() != withdrawalPosting.GetItem().GetId() {
		t.Fatalf("find withdrawn posting by reason: %+v, %v", byReason, err)
	}
	byActor, err := client.FindJobPostingByWithdrawnByPrincipalID(ctx, &common.StringFinderRequest{Value: &policy.principalID, Limit: 1})
	if err != nil || len(byActor.GetItems()) != 1 || byActor.GetItems()[0].GetId() != withdrawalPosting.GetItem().GetId() {
		t.Fatalf("find withdrawn posting by actor: %+v, %v", byActor, err)
	}
	if _, err := client.WithdrawJobPosting(ctx, &pb.WithdrawJobPostingRequest{
		Id: withdrawalPosting.GetItem().GetId(), ReasonCode: "DOCUMENT_REVOKED",
	}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("repeat posting withdrawal status = %v", status.Code(err))
	}
	if err := database.WithContext(ctx).Table("hr_outbox_events").Where("aggregate_id = ?", withdrawalPosting.GetItem().GetId()).Count(&postingOutboxCount).Error; err != nil || postingOutboxCount != 3 {
		t.Fatalf("withdrawn posting outbox count = %d, err = %v", postingOutboxCount, err)
	}
	if err := database.WithContext(ctx).Table("hr_job_postings").Where("id = ?", withdrawalPosting.GetItem().GetId()).Update("withdrawal_reason_code", nil).Error; err == nil {
		t.Fatal("database accepted withdrawn posting without reason")
	}
	closed, err := client.CloseRequisition(ctx, &common.IDRequest{Id: submittedID})
	if err != nil || closed.GetItem().GetStatus() != "closed" || closed.GetItem().GetClosedAt() == nil {
		t.Fatalf("close requisition: %+v, %v", closed, err)
	}
	if _, err := client.CloseRequisition(ctx, &common.IDRequest{Id: submittedID}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("repeat closing status = %v", status.Code(err))
	}
	if err := database.WithContext(ctx).Table("hr_outbox_events").Where("aggregate_id = ?", submittedID).Count(&requisitionOutboxCount).Error; err != nil || requisitionOutboxCount != 5 {
		t.Fatalf("closed requisition outbox count = %d, err = %v", requisitionOutboxCount, err)
	}
	postingInput.ChannelCode = text("late_board")
	if _, err := client.CreateJobPosting(ctx, &pb.CreateJobPostingRequest{Input: postingInput}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("create posting after close status = %v", status.Code(err))
	}
	if _, err := client.DeleteJobPosting(ctx, &common.IDRequest{Id: posting.GetItem().GetId()}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("delete closed posting status = %v", status.Code(err))
	}
	concurrentNumber, concurrentKey := "REQ-CONCURRENT-1", "REQ-CONCURRENT-1"
	concurrentInput := &pb.RequisitionInput{RequisitionNumber: &concurrentNumber, JobId: text(jobID),
		OrganizationUnitReferenceId: text(familyID), HiringManagerPrincipalId: text(managerPrincipalID),
		RequisitionType: &draftType, EmploymentCategory: &category, IdempotencyKey: &concurrentKey}
	concurrentDraft, err := client.CreateRequisition(ctx, &pb.CreateRequisitionRequest{Input: concurrentInput})
	if err != nil {
		t.Fatal(err)
	}
	concurrentID := concurrentDraft.GetItem().GetId()
	if _, err := client.SubmitRequisition(ctx, &common.IDRequest{Id: concurrentID}); err != nil {
		t.Fatal(err)
	}
	policy.principalID = "00000000-0000-4000-8000-000000000098"
	approvals := make(chan error, 2)
	for range 2 {
		go func() {
			_, approvalErr := client.ApproveRequisition(ctx, &common.IDRequest{Id: concurrentID})
			approvals <- approvalErr
		}()
	}
	successes, conflicts := 0, 0
	for range 2 {
		approvalErr := <-approvals
		switch status.Code(approvalErr) {
		case codes.OK:
			successes++
		case codes.FailedPrecondition:
			conflicts++
		default:
			t.Fatalf("concurrent approval status = %v", status.Code(approvalErr))
		}
	}
	policy.principalID = managerPrincipalID
	if successes != 1 || conflicts != 1 {
		t.Fatalf("concurrent approvals: success %d, conflict %d", successes, conflicts)
	}
	if err := database.WithContext(ctx).Table("hr_outbox_events").Where("aggregate_id = ?", concurrentID).Count(&requisitionOutboxCount).Error; err != nil || requisitionOutboxCount != 3 {
		t.Fatalf("concurrent approval outbox count = %d, err = %v", requisitionOutboxCount, err)
	}
	if _, err := client.OpenRequisition(ctx, &common.IDRequest{Id: concurrentID}); err != nil {
		t.Fatal(err)
	}
	activeApplicationID := "00000000-0000-4000-8000-000000000400"
	if err := database.WithContext(ctx).Exec(`INSERT INTO hr_applications
		(id, candidate_id, requisition_id, application_number, status, idempotency_key)
		VALUES (?, ?, ?, 'APP-CLOSE-1', 'received', 'APP-CLOSE-1')`,
		activeApplicationID, linkedCandidate.GetItem().GetId(), concurrentID).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := client.CloseRequisition(ctx, &common.IDRequest{Id: concurrentID}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("close with active application status = %v", status.Code(err))
	}
	rejectedApplication, err := client.RejectApplication(ctx, &pb.RejectApplicationRequest{
		Id: activeApplicationID, ReasonCode: "POSITION_CANCELLED",
	})
	if err != nil || rejectedApplication.GetItem().GetStatus() != "rejected" ||
		rejectedApplication.GetItem().GetDispositionReasonCode() != "POSITION_CANCELLED" ||
		rejectedApplication.GetItem().GetDecisionAt() == nil ||
		rejectedApplication.GetItem().GetDecisionByPrincipalId() != policy.principalID {
		t.Fatalf("reject active application: %+v, %v", rejectedApplication, err)
	}
	if _, err := client.RejectApplication(ctx, &pb.RejectApplicationRequest{
		Id: activeApplicationID, ReasonCode: "POSITION_CANCELLED",
	}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("repeat application rejection status = %v", status.Code(err))
	}
	var rejectionHistoryCount, rejectionOutboxCount int64
	if err := database.WithContext(ctx).Table("hr_application_stage_history").Where("application_id = ?", activeApplicationID).Count(&rejectionHistoryCount).Error; err != nil || rejectionHistoryCount != 1 {
		t.Fatalf("rejection history count = %d, err = %v", rejectionHistoryCount, err)
	}
	if err := database.WithContext(ctx).Table("hr_outbox_events").Where("aggregate_id = ?", activeApplicationID).Count(&rejectionOutboxCount).Error; err != nil || rejectionOutboxCount != 1 {
		t.Fatalf("rejection outbox count = %d, err = %v", rejectionOutboxCount, err)
	}
	if err := database.WithContext(ctx).Table("hr_applications").Where("id = ?", activeApplicationID).Update("disposition_reason_code", nil).Error; err == nil {
		t.Fatal("database accepted rejected application without reason")
	}
	if err := database.WithContext(ctx).Table("hr_applications").Where("id = ?", activeApplicationID).Update("current_stage_code", nil).Error; err == nil {
		t.Fatal("database accepted rejected application without stage")
	}
	if _, err := client.CloseRequisition(ctx, &common.IDRequest{Id: concurrentID}); err != nil {
		t.Fatalf("close with terminal application: %v", err)
	}
	if err := database.WithContext(ctx).Table("hr_outbox_events").Where("aggregate_id = ?", concurrentID).Count(&requisitionOutboxCount).Error; err != nil || requisitionOutboxCount != 5 {
		t.Fatalf("closed requisition outbox count = %d, err = %v", requisitionOutboxCount, err)
	}
	if err := database.WithContext(ctx).Exec("DELETE FROM hr_application_stage_history WHERE application_id = ?", activeApplicationID).Error; err != nil {
		t.Fatal(err)
	}
	if err := database.WithContext(ctx).Exec("DELETE FROM hr_applications WHERE id = ?", activeApplicationID).Error; err != nil {
		t.Fatal(err)
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
	sourcePostingID := "00000000-0000-4000-8000-000000000201"
	if err := database.WithContext(ctx).Exec(`INSERT INTO hr_job_postings
		(id, requisition_id, channel_code, posting_version, content_document_version_reference_id,
		content_digest, status, published_at, created_by_principal_id)
		VALUES (?, ?, 'career_site', 1, ?, ?, 'published', NOW() - INTERVAL '30 minutes', ?)`,
		sourcePostingID, requisitionID, familyID, bytes.Repeat([]byte{0x42}, 32), familyID).Error; err != nil {
		t.Fatal(err)
	}
	applicationNumber := "APP-DRAFT-1"
	idempotencyKey := "APP-DRAFT-1"
	applicationInput := &pb.ApplicationInput{CandidateId: text(linkedCandidate.GetItem().GetId()),
		RequisitionId: text(requisitionID), ApplicationNumber: &applicationNumber, IdempotencyKey: &idempotencyKey,
		SourcePostingId: &sourcePostingID}
	application, err := client.CreateApplication(ctx, &pb.CreateApplicationRequest{Input: applicationInput})
	if err != nil || application.GetItem().GetStatus() != "received" || application.GetItem().GetSourcePostingId() != sourcePostingID {
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
	start := timestamppb.New(time.Now().Add(48 * time.Hour).Truncate(time.Microsecond))
	end := timestamppb.New(start.AsTime().Add(time.Hour))
	createInterview := &pb.InterviewInput{ApplicationId: &applicationID, InterviewRound: &round,
		InterviewType: &interviewType, ScheduledStart: timestamppb.New(start.AsTime().Add(123 * time.Nanosecond)), ScheduledEnd: end, CandidateTimezone: &timezone}
	var originalOpening time.Time
	if err := database.WithContext(ctx).Table("hr_requisitions").Where("id = ?", requisitionID).
		Select("opened_at").Scan(&originalOpening).Error; err != nil {
		t.Fatalf("read requisition opening: %v", err)
	}
	if err := database.WithContext(ctx).Table("hr_requisitions").Where("id = ?", requisitionID).
		Update("opened_at", time.Now().Add(time.Hour)).Error; err != nil {
		t.Fatalf("set future opening on requisition fixture: %v", err)
	}
	if _, err := client.CreateInterview(ctx, &pb.CreateInterviewRequest{Input: createInterview}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("create interview before requisition opens: %v", err)
	}
	if err := database.WithContext(ctx).Table("hr_requisitions").Where("id = ?", requisitionID).
		Update("opened_at", originalOpening).Error; err != nil {
		t.Fatalf("restore requisition opening: %v", err)
	}
	createdInterview, err := client.CreateInterview(ctx, &pb.CreateInterviewRequest{Input: createInterview})
	if err != nil || createdInterview.GetItem().GetId() == "" || createdInterview.GetItem().GetStatus() != "planned" {
		t.Fatalf("create planned interview: %+v, %v", createdInterview, err)
	}
	if !createdInterview.GetItem().GetScheduledStart().AsTime().Equal(start.AsTime()) {
		t.Fatalf("created interview must report persisted timestamp precision: %v", createdInterview.GetItem().GetScheduledStart())
	}
	interviewID := createdInterview.GetItem().GetId()
	if createdInterview.GetItem().GetCreatedByPrincipalId() != "00000000-0000-4000-8000-000000000001" {
		t.Fatal("interview creator did not come from server actor")
	}
	if _, err := client.CreateInterview(ctx, &pb.CreateInterviewRequest{Input: createInterview}); status.Code(err) != codes.AlreadyExists {
		t.Fatalf("duplicate interview round status = %v, want AlreadyExists", status.Code(err))
	}
	otherRound := int32(2)
	if _, err := client.CreateInterview(ctx, &pb.CreateInterviewRequest{Input: &pb.InterviewInput{
		ApplicationId: &applicationID, InterviewRound: &otherRound, InterviewType: &interviewType,
		ScheduledStart: end, ScheduledEnd: timestamppb.New(end.AsTime().Add(time.Nanosecond)), CandidateTimezone: &timezone,
	}}); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("interview interval below database precision = %v, want InvalidArgument", err)
	}
	if _, err := client.CreateInterview(ctx, &pb.CreateInterviewRequest{Input: &pb.InterviewInput{
		ApplicationId: &applicationID, InterviewRound: &otherRound, InterviewType: &interviewType,
		ScheduledStart: timestamppb.New(start.AsTime().Add(30 * time.Minute)),
		ScheduledEnd:   timestamppb.New(end.AsTime().Add(30 * time.Minute)), CandidateTimezone: &timezone,
	}}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("overlapping interview status = %v, want FailedPrecondition", status.Code(err))
	}
	if err := database.WithContext(ctx).Table("hr_interviews").Where("id = ?", interviewID).
		Update("status", "in_progress").Error; err != nil {
		t.Fatalf("set current interview in progress: %v", err)
	}
	if _, err := client.CreateInterview(ctx, &pb.CreateInterviewRequest{Input: &pb.InterviewInput{
		ApplicationId: &applicationID, InterviewRound: &otherRound, InterviewType: &interviewType,
		ScheduledStart: start, ScheduledEnd: end, CandidateTimezone: &timezone,
	}}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("overlap with in-progress interview = %v, want FailedPrecondition", err)
	}
	if err := database.WithContext(ctx).Table("hr_interviews").Where("id = ?", interviewID).
		Update("status", "planned").Error; err != nil {
		t.Fatalf("restore planned interview fixture: %v", err)
	}
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
	const alternativeRequisition = "00000000-0000-4000-8000-000000000490"
	const alternativeApplication = "00000000-0000-4000-8000-000000000491"
	if err := database.WithContext(ctx).Exec(`INSERT INTO hr_requisitions
		(id, requisition_number, job_id, organization_unit_reference_id, hiring_manager_principal_id,
		requisition_type, status, employment_category, approved_at, approved_by_principal_id, opened_at, idempotency_key)
		SELECT ?, 'REQ-INTERVIEW-ALT', job_id, organization_unit_reference_id, hiring_manager_principal_id,
		requisition_type, status, employment_category, approved_at, approved_by_principal_id, opened_at, 'REQ-INTERVIEW-ALT'
		FROM hr_requisitions WHERE id = ?`, alternativeRequisition, requisitionID).Error; err != nil {
		t.Fatal(err)
	}
	if err := database.WithContext(ctx).Exec(`INSERT INTO hr_applications
		(id, candidate_id, requisition_id, application_number, idempotency_key)
		VALUES (?, ?, ?, 'APP-INTERVIEW-ALT', 'APP-INTERVIEW-ALT')`,
		alternativeApplication, linkedCandidate.GetItem().GetId(), alternativeRequisition).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := client.CreateInterview(ctx, &pb.CreateInterviewRequest{Input: &pb.InterviewInput{
		ApplicationId: text(alternativeApplication), InterviewRound: &round, InterviewType: &interviewType,
		ScheduledStart: start, ScheduledEnd: end, CandidateTimezone: &timezone,
	}}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("schedule overlap for candidate across applications = %v, want FailedPrecondition", status.Code(err))
	}
	type scheduleResult struct {
		id  string
		err error
	}
	beginScheduling := make(chan struct{})
	scheduled := make(chan scheduleResult, 2)
	for _, application := range []string{applicationID, alternativeApplication} {
		go func() {
			<-beginScheduling
			result, err := client.CreateInterview(ctx, &pb.CreateInterviewRequest{Input: &pb.InterviewInput{
				ApplicationId: &application, InterviewRound: &otherRound, InterviewType: &interviewType,
				ScheduledStart: end, ScheduledEnd: timestamppb.New(end.AsTime().Add(time.Hour)), CandidateTimezone: &timezone,
			}})
			scheduled <- scheduleResult{id: result.GetItem().GetId(), err: err}
		}()
	}
	close(beginScheduling)
	var adjacentID string
	var rejectedSchedules int
	for range 2 {
		result := <-scheduled
		switch {
		case result.err == nil && result.id != "":
			adjacentID = result.id
		case status.Code(result.err) == codes.FailedPrecondition:
			rejectedSchedules++
		default:
			t.Errorf("schedule adjacent interview: %+v", result)
		}
	}
	if adjacentID == "" || rejectedSchedules != 1 {
		t.Fatalf("concurrent scheduling: accepted interview %q, rejected %d requests", adjacentID, rejectedSchedules)
	}
	if _, err := client.UpdateInterview(ctx, &pb.UpdateInterviewRequest{Id: interviewID,
		Input:      &pb.InterviewInput{ScheduledEnd: timestamppb.New(end.AsTime().Add(time.Minute))},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"scheduled_end"}}}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("reschedule over adjacent interview = %v, want FailedPrecondition", err)
	}
	if _, err := client.DeleteInterview(ctx, &common.IDRequest{Id: adjacentID}); err != nil {
		t.Fatalf("delete adjacent interview fixture: %v", err)
	}
	const withdrawnInterview = "00000000-0000-4000-8000-000000000492"
	if err := database.WithContext(ctx).Exec("UPDATE hr_applications SET status = 'withdrawn' WHERE id = ?", alternativeApplication).Error; err != nil {
		t.Fatal(err)
	}
	if err := database.WithContext(ctx).Exec(`INSERT INTO hr_interviews
		(id, application_id, interview_round, interview_type, scheduled_start, scheduled_end)
		VALUES (?, ?, 1, 'screening', ?, ?)`, withdrawnInterview, alternativeApplication,
		end.AsTime(), end.AsTime().Add(time.Hour)).Error; err != nil {
		t.Fatal(err)
	}
	released, err := client.CreateInterview(ctx, &pb.CreateInterviewRequest{Input: &pb.InterviewInput{
		ApplicationId: &applicationID, InterviewRound: &otherRound, InterviewType: &interviewType,
		ScheduledStart: end, ScheduledEnd: timestamppb.New(end.AsTime().Add(time.Hour)), CandidateTimezone: &timezone,
	}})
	if err != nil {
		t.Fatalf("reuse planned slot from withdrawn application: %v", err)
	}
	for _, fixtureID := range []string{released.GetItem().GetId(), withdrawnInterview} {
		if _, err := client.DeleteInterview(ctx, &common.IDRequest{Id: fixtureID}); err != nil {
			t.Fatalf("delete interview fixture %s: %v", fixtureID, err)
		}
	}
	if err := database.WithContext(ctx).Exec("DELETE FROM hr_applications WHERE id = ?", alternativeApplication).Error; err != nil {
		t.Fatal(err)
	}
	if err := database.WithContext(ctx).Exec("DELETE FROM hr_requisitions WHERE id = ?", alternativeRequisition).Error; err != nil {
		t.Fatal(err)
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
	cancelRound := int32(20)
	cancelInput := &pb.InterviewInput{
		ApplicationId: &applicationID, InterviewRound: &cancelRound, InterviewType: &interviewType,
		ScheduledStart: newStart, ScheduledEnd: end, CandidateTimezone: &timezone,
	}
	toCancel, err := client.CreateInterview(ctx, &pb.CreateInterviewRequest{Input: cancelInput})
	if err != nil {
		t.Fatalf("create interview for cancellation: %v", err)
	}
	cancelID := toCancel.GetItem().GetId()
	if err := database.WithContext(ctx).Exec(`INSERT INTO hr_interview_panelists
		(id, interview_id, interviewer_principal_id) VALUES (?, ?, ?)`,
		panelistID, cancelID, "00000000-0000-4000-8000-000000000001").Error; err != nil {
		t.Fatal(err)
	}
	cancelRequest := &common.IDRequest{Id: cancelID}
	policy.principalID = "00000000-0000-4000-8000-000000000098"
	if _, err := client.CancelInterview(ctx, cancelRequest); status.Code(err) != codes.PermissionDenied {
		t.Fatalf("unassigned recruiter cancellation: %v", err)
	}
	policy.principalID = "00000000-0000-4000-8000-000000000001"
	cancelled, err := client.CancelInterview(ctx, cancelRequest)
	if err != nil || cancelled.GetItem().GetStatus() != "cancelled" {
		t.Fatalf("cancel interview: %+v, %v", cancelled, err)
	}
	if _, err := client.CancelInterview(ctx, cancelRequest); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("cancel already cancelled interview: %v", err)
	}
	cancelRound++
	if _, err := client.CreateInterview(ctx, &pb.CreateInterviewRequest{Input: cancelInput}); err != nil {
		t.Fatalf("reuse cancelled interview slot: %v", err)
	}
	var retainedPanelists int64
	if err := database.WithContext(ctx).Table("hr_interview_panelists").Where("interview_id = ?", cancelID).Count(&retainedPanelists).Error; err != nil || retainedPanelists != 1 {
		t.Fatalf("preserve panelist after cancellation: count=%d err=%v", retainedPanelists, err)
	}
	if err := database.WithContext(ctx).Table("hr_outbox_events").Where("aggregate_id = ? AND event_type = ?", cancelID, "hr_interviews.cancel").Count(&outboxCount).Error; err != nil || outboxCount != 1 {
		t.Fatalf("interview cancellation outbox count = %d, err = %v", outboxCount, err)
	}
	policy.principalID = "00000000-0000-4000-8000-000000000098"
	if _, err := client.RejectApplication(ctx, &pb.RejectApplicationRequest{Id: applicationID, ReasonCode: "NOT_ELIGIBLE"}); status.Code(err) != codes.PermissionDenied {
		t.Fatalf("other recruiter rejection status = %v", status.Code(err))
	}
	policy.principalID = "00000000-0000-4000-8000-000000000001"
	if _, err := client.RejectApplication(ctx, &pb.RejectApplicationRequest{Id: applicationID, ReasonCode: "NOT_ELIGIBLE"}); err != nil {
		t.Fatalf("assigned recruiter rejection: %v", err)
	}
	if err := database.WithContext(ctx).Table("hr_application_stage_history").Where("application_id = ?", applicationID).Count(&stageCount).Error; err != nil || stageCount != 2 {
		t.Fatalf("screening rejection history count = %d, err = %v", stageCount, err)
	}
	if err := database.WithContext(ctx).Table("hr_outbox_events").Where("aggregate_id = ?", applicationID).Count(&outboxCount).Error; err != nil || outboxCount != 3 {
		t.Fatalf("screening rejection outbox count = %d, err = %v", outboxCount, err)
	}
	shortlistCandidate, err := create("CAND-SHORTLIST-1")
	if err != nil {
		t.Fatalf("create shortlist candidate: %v", err)
	}
	shortlistNumber := "APP-SHORTLIST-1"
	shortlistApplication, err := client.CreateApplication(ctx, &pb.CreateApplicationRequest{Input: &pb.ApplicationInput{
		CandidateId: text(shortlistCandidate.GetItem().GetId()), RequisitionId: text(requisitionID),
		ApplicationNumber: &shortlistNumber, IdempotencyKey: &shortlistNumber,
	}})
	if err != nil {
		t.Fatalf("create shortlist application: %v", err)
	}
	shortlistID := shortlistApplication.GetItem().GetId()
	if _, err := client.ShortlistApplication(ctx, &pb.ShortlistApplicationRequest{
		Id: shortlistID, ReasonCode: "MEETS_REQUIREMENTS",
	}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("shortlist before screening status = %v", status.Code(err))
	}
	if _, err := client.StartApplicationScreening(ctx, &common.IDRequest{Id: shortlistID}); err != nil {
		t.Fatalf("start shortlist screening: %v", err)
	}
	policy.principalID = "00000000-0000-4000-8000-000000000098"
	if _, err := client.ShortlistApplication(ctx, &pb.ShortlistApplicationRequest{
		Id: shortlistID, ReasonCode: "MEETS_REQUIREMENTS",
	}); status.Code(err) != codes.PermissionDenied {
		t.Fatalf("other recruiter shortlist status = %v", status.Code(err))
	}
	policy.principalID = "00000000-0000-4000-8000-000000000001"
	if _, err := client.ShortlistApplication(ctx, &pb.ShortlistApplicationRequest{Id: shortlistID}); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("shortlist without reason status = %v", status.Code(err))
	}
	shortlisted, err := client.ShortlistApplication(ctx, &pb.ShortlistApplicationRequest{
		Id: shortlistID, ReasonCode: "MEETS_REQUIREMENTS",
	})
	if err != nil || shortlisted.GetItem().GetStatus() != "shortlisted" ||
		shortlisted.GetItem().GetCurrentStageCode() != "shortlisted" ||
		shortlisted.GetItem().GetScreeningScore() != "" || shortlisted.GetItem().GetScorecardVersion() != "" {
		t.Fatalf("shortlist application: %+v, %v", shortlisted, err)
	}
	if _, err := client.ShortlistApplication(ctx, &pb.ShortlistApplicationRequest{
		Id: shortlistID, ReasonCode: "MEETS_REQUIREMENTS",
	}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("repeat shortlist status = %v", status.Code(err))
	}
	shortlistHistory, err := client.FindApplicationStageHistoryByApplicationID(ctx, &common.StringFinderRequest{Value: &shortlistID, Limit: 10})
	if err != nil || len(shortlistHistory.GetItems()) != 2 ||
		shortlistHistory.GetItems()[1].GetNewStatus() != "shortlisted" ||
		shortlistHistory.GetItems()[1].GetReasonCode() != "MEETS_REQUIREMENTS" ||
		shortlistHistory.GetItems()[1].GetChangedByPrincipalId() != policy.principalID {
		t.Fatalf("shortlist stage history: %+v, %v", shortlistHistory, err)
	}
	if err := database.WithContext(ctx).Table("hr_outbox_events").Where("aggregate_id = ?", shortlistID).Count(&outboxCount).Error; err != nil || outboxCount != 3 {
		t.Fatalf("shortlist outbox count = %d, err = %v", outboxCount, err)
	}
	if err := database.WithContext(ctx).Table("hr_applications").Where("id = ?", shortlistID).Update("current_stage_code", nil).Error; err == nil {
		t.Fatal("database accepted shortlisted application without stage")
	}
	shortlistRejection, err := client.RejectApplication(ctx, &pb.RejectApplicationRequest{
		Id: shortlistID, ReasonCode: "NO_LONGER_AVAILABLE",
	})
	if err != nil || shortlistRejection.GetItem().GetStatus() != "rejected" {
		t.Fatalf("reject shortlisted application: %+v, %v", shortlistRejection, err)
	}
	if err := database.WithContext(ctx).Table("hr_application_stage_history").Where("application_id = ?", shortlistID).Count(&stageCount).Error; err != nil || stageCount != 3 {
		t.Fatalf("shortlist rejection history count = %d, err = %v", stageCount, err)
	}
	if err := database.WithContext(ctx).Table("hr_outbox_events").Where("aggregate_id = ?", shortlistID).Count(&outboxCount).Error; err != nil || outboxCount != 4 {
		t.Fatalf("shortlist rejection outbox count = %d, err = %v", outboxCount, err)
	}
	interviewCandidate, err := create("CAND-INTERVIEW-START")
	if err != nil {
		t.Fatal(err)
	}
	interviewNumber := "APP-INTERVIEW-START"
	interviewApplication, err := client.CreateApplication(ctx, &pb.CreateApplicationRequest{Input: &pb.ApplicationInput{
		CandidateId: text(interviewCandidate.GetItem().GetId()), RequisitionId: text(requisitionID),
		ApplicationNumber: &interviewNumber, IdempotencyKey: &interviewNumber,
	}})
	if err != nil {
		t.Fatal(err)
	}
	interviewApplicationID := interviewApplication.GetItem().GetId()
	if _, err := client.StartApplicationScreening(ctx, &common.IDRequest{Id: interviewApplicationID}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.ShortlistApplication(ctx, &pb.ShortlistApplicationRequest{Id: interviewApplicationID, ReasonCode: "MEETS_REQUIREMENTS"}); err != nil {
		t.Fatal(err)
	}
	toStart, err := client.CreateInterview(ctx, &pb.CreateInterviewRequest{Input: &pb.InterviewInput{
		ApplicationId: &interviewApplicationID, InterviewRound: &cancelRound, InterviewType: &interviewType,
		ScheduledStart: newStart, ScheduledEnd: end, CandidateTimezone: &timezone,
	}})
	if err != nil {
		t.Fatal(err)
	}
	startRequest := &common.IDRequest{Id: toStart.GetItem().GetId()}
	if _, err := client.StartInterview(ctx, startRequest); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("start interview before schedule: %v", err)
	}
	if err := database.WithContext(ctx).Exec("UPDATE hr_interviews SET scheduled_start = NOW() - INTERVAL '1 minute' WHERE id = ?", startRequest.Id).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := client.StartInterview(ctx, startRequest); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("start interview without panelists: %v", err)
	}
	stillShortlisted, err := client.GetApplicationByID(ctx, &common.IDRequest{Id: interviewApplicationID})
	if err != nil || stillShortlisted.GetItem().GetStatus() != "shortlisted" {
		t.Fatalf("failed interview start did not roll back application: %+v, %v", stillShortlisted, err)
	}
	if err := database.WithContext(ctx).Exec("INSERT INTO hr_interview_panelists (id, interview_id, interviewer_principal_id) VALUES (?, ?, ?)",
		"00000000-0000-4000-8000-000000000396", startRequest.Id, policy.principalID).Error; err != nil {
		t.Fatal(err)
	}
	startedInterview, err := client.StartInterview(ctx, startRequest)
	if err != nil || startedInterview.GetItem().GetStatus() != "in_progress" {
		t.Fatalf("start interview: %+v, %v", startedInterview, err)
	}
	startedApplication, err := client.GetApplicationByID(ctx, &common.IDRequest{Id: interviewApplicationID})
	if err != nil || startedApplication.GetItem().GetStatus() != "interview" {
		t.Fatalf("start interview did not advance the application: %+v, %v", startedApplication, err)
	}
	if _, err := client.StartInterview(ctx, startRequest); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("repeat interview start: %v", err)
	}
	if err := database.WithContext(ctx).Table("hr_application_stage_history").
		Where("application_id = ? AND new_status = ?", interviewApplicationID, "interview").Count(&stageCount).Error; err != nil || stageCount != 1 {
		t.Fatalf("interview stage history count = %d, %v", stageCount, err)
	}
	if _, err := client.CompleteInterview(ctx, startRequest); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("complete without panel feedback: %v", err)
	}
	if err := database.WithContext(ctx).Exec(`INSERT INTO hr_interview_feedback
		(id, interview_id, interviewer_principal_id, feedback_version, submitted_at, overall_recommendation, competency_scores)
		VALUES (?, ?, ?, 1, NOW() - INTERVAL '2 minutes', 'yes', '{}'::jsonb)`,
		"00000000-0000-4000-8000-000000000397", startRequest.Id, policy.principalID).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := client.CompleteInterview(ctx, startRequest); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("complete with feedback predating the slot: %v", err)
	}
	if err := database.WithContext(ctx).Exec(`INSERT INTO hr_interview_feedback
		(id, interview_id, interviewer_principal_id, feedback_version, overall_recommendation, competency_scores)
		VALUES (?, ?, ?, 2, 'yes', '{}'::jsonb)`,
		"00000000-0000-4000-8000-000000000398", startRequest.Id, policy.principalID).Error; err != nil {
		t.Fatal(err)
	}
	completed, err := client.CompleteInterview(ctx, startRequest)
	if err != nil || completed.GetItem().GetStatus() != "completed" {
		t.Fatalf("complete interview with current panel feedback: %+v, %v", completed, err)
	}
	if _, err := client.CompleteInterview(ctx, startRequest); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("repeat interview completion: %v", err)
	}
	if err := database.WithContext(ctx).Table("hr_outbox_events").Where("aggregate_id = ? AND event_type = ?", startRequest.Id, "hr_interviews.complete").Count(&outboxCount).Error; err != nil || outboxCount != 1 {
		t.Fatalf("interview completion outbox count = %d, err = %v", outboxCount, err)
	}

}
