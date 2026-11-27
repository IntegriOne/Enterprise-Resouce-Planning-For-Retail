//go:build integration

package recruitment_test

import (
	"bytes"
	"context"
	"errors"
	"net"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/recruitment/v1"
	hrdb "github.com/integrione/erp-retail/backend/services/human_resource/db"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	"github.com/integrione/erp-retail/backend/services/human_resource/infrastructure/security"
	appservice "github.com/integrione/erp-retail/backend/services/human_resource/internal/service"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/recruitment"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository"
	handler "github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/recruitment"
	"github.com/integrione/erp-retail/backend/shared/authcontext"
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

type recDetailPolicy struct {
	mu          sync.Mutex
	principalID string
}

func (p *recDetailPolicy) setPrincipal(id string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.principalID = id
}

func (p *recDetailPolicy) currentPrincipal() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.principalID
}

func (p *recDetailPolicy) Authorize(context.Context, operation.Access) (operation.Actor, error) {
	return operation.Actor{PrincipalID: p.currentPrincipal(), SessionID: "00000000-0000-4000-8000-000000000002", Kind: "human"}, nil
}

func (*recDetailPolicy) ReadScope(context.Context, operation.Access) (operation.Scope, error) {
	return operation.Scope{All: true}, nil
}

type recDetailFoundationStub struct {
	requisitionIntegrationVerifier
}

func (recDetailFoundationStub) VerifyLegalEntity(context.Context, string) error { return nil }
func (recDetailFoundationStub) VerifyWorkerReferences(context.Context, security.WorkerReferenceCheck) error {
	return nil
}
func (recDetailFoundationStub) VerifyOrganizationUnit(context.Context, string, string) error {
	return nil
}
func (recDetailFoundationStub) VerifyLocation(context.Context, string, string) error { return nil }
func (recDetailFoundationStub) VerifyParty(context.Context, string, bool) error      { return nil }

func (recDetailFoundationStub) VerifyPartyContact(context.Context, string, string) error { return nil }

type recAssessmentEvidenceStub struct{}

func (recAssessmentEvidenceStub) VerifyAssessmentEvidence(_ context.Context, req service.AssessmentEvidenceRequest) (service.AssessmentEvidenceProof, error) {
	return service.AssessmentEvidenceProof{
		VersionID:       req.VersionID,
		Digest:          append([]byte(nil), req.Digest...),
		Verified:        true,
		Immutable:       true,
		OwnerManagement: "human_resource",
		Purpose:         "recruitment_assessment_evidence",
		AssessmentID:    req.AssessmentID,
		ApplicationID:   req.ApplicationID,
	}, nil
}

type offerCompensationIntegrationVerifier struct{ mismatched bool }

func (v *offerCompensationIntegrationVerifier) VerifyOfferCompensation(_ context.Context, check service.OfferCompensationCheck) (service.OfferCompensationProof, error) {
	proof := service.OfferCompensationProof{
		VersionID: check.VersionID, OfferID: check.OfferID, ApplicationID: check.ApplicationID, JobID: check.JobID,
		LegalEntityID: check.LegalEntityID, CurrencyCode: check.CurrencyCode, OwnerManagement: "human_resource",
		Purpose: "recruitment_offer_compensation", Verified: true, Immutable: true, Approved: true,
	}
	if v.mismatched {
		proof.CurrencyCode = "USD"
	}
	return proof, nil
}

func (v *offerCompensationIntegrationVerifier) VerifyOfferDocument(_ context.Context, check service.OfferDocumentCheck) (service.OfferDocumentProof, error) {
	digest := bytes.Clone(check.Digest)
	if v.mismatched {
		digest[0] ^= 1
	}
	return service.OfferDocumentProof{
		DocumentID: check.DocumentID, VersionID: check.VersionID, OfferID: check.OfferID,
		ApplicationID: check.ApplicationID, CompensationVersionID: check.CompensationVersionID,
		OwnerManagement: "human_resource", Purpose: "recruitment_offer", Digest: digest,
		Verified: true, Immutable: true,
	}, nil
}

type offerResponseIntegrationVerifier struct {
	decision    string
	respondedAt time.Time
	mismatched  bool
}

func (v *offerResponseIntegrationVerifier) VerifyOfferResponse(_ context.Context, check service.OfferResponseCheck) (service.OfferResponseProof, error) {
	proof := service.OfferResponseProof{
		EvidenceID: check.EvidenceID, OfferID: check.OfferID, CandidateID: check.CandidateID,
		VersionID: check.VersionID, Digest: check.Digest, Decision: v.decision,
		RespondedAt: v.respondedAt, CandidateVerified: !v.mismatched,
	}
	return proof, nil
}

func TestRecruitmentDetailAndAssessmentRPCWithPostgreSQL(t *testing.T) {
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
	validators, err := security.RecruitmentCandidateValidators(repos, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, fn := range []func(*repository.Registry) (map[string]security.MutationValidator, error){
		security.RecruitmentApplicationValidators,
		security.RecruitmentJobPostingValidators,
		security.RecruitmentInterviewValidators,
	} {
		m, err := fn(repos)
		if err != nil {
			t.Fatal(err)
		}
		for k, v := range m {
			validators[k] = v
		}
	}
	foundation := recDetailFoundationStub{}
	for _, mfn := range []func() (map[string]security.MutationValidator, error){
		func() (map[string]security.MutationValidator, error) {
			return security.RecruitmentRequisitionValidators(repos, foundation)
		},
		func() (map[string]security.MutationValidator, error) {
			return security.RecruitmentDetailValidators(repos, foundation)
		},
	} {
		m, err := mfn()
		if err != nil {
			t.Fatal(err)
		}
		for k, v := range m {
			validators[k] = v
		}
	}

	assessorID := "00000000-0000-4000-8000-000000000010"
	otherID := "00000000-0000-4000-8000-000000000011"
	policy := &recDetailPolicy{principalID: assessorID}
	offerCompensation := &offerCompensationIntegrationVerifier{}
	offerResponse := &offerResponseIntegrationVerifier{}
	services, err := appservice.NewRegistry(repos, operation.Dependencies{
		Policy: policy, Integrity: security.NewIntegrity(validators),
	}, appservice.WithRecruitmentVerifiers(appservice.RecruitmentVerifiers{
		RequisitionMandate:  &requisitionApprovalIntegrationVerifier{},
		RequisitionApproval: &requisitionApprovalIntegrationVerifier{budgetID: "00000000-0000-4000-8000-000000000099"},
		PostingDocument:     postingDocumentIntegrationVerifier{},
		OfferCompensation:   offerCompensation,
		OfferDocument:       offerCompensation,
		OfferResponse:       offerResponse,
		AssessmentEvidence:  recAssessmentEvidenceStub{},
	}))
	if err != nil {
		t.Fatal(err)
	}
	listener := bufconn.Listen(1 << 20)
	server := grpc.NewServer(grpc.UnaryInterceptor(func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, next grpc.UnaryHandler) (any, error) {
		return next(authcontext.WithCaller(ctx, authcontext.CallerContext{PrincipalID: policy.currentPrincipal(), Authenticated: true}), req)
	}))
	if err := handler.Register(server, services.Recruitment); err != nil {
		t.Fatal(err)
	}
	served := make(chan error, 1)
	go func() { served <- server.Serve(listener) }()
	t.Cleanup(func() { server.Stop(); _ = listener.Close(); <-served })
	conn, err := grpc.NewClient("passthrough:///hrm-recruitment-detail-integration",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return listener.DialContext(ctx) }),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	client := pb.NewRecruitmentServiceClient(conn)

	ids := operation.UUIDGenerator{}
	familyID, _ := ids.NewID()
	jobID, _ := ids.NewID()
	unitID, _ := ids.NewID()
	legalID, _ := ids.NewID()
	workerID, _ := ids.NewID()
	rubricID, _ := ids.NewID()
	evDocID, _ := ids.NewID()
	reqID, _ := ids.NewID()
	var candidateID, appID, interviewID, overlapInterviewID, panelistID, feedbackID, offerID, assessID string
	var adjacentPanelistID string
	var secondCandidateID, secondApplicationID string

	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		for _, stmt := range []struct {
			q string
			v string
		}{
			{"DELETE FROM hr_application_assessments WHERE id = ?", assessID},
			{"DELETE FROM hr_offers WHERE id = ?", offerID},
			{"DELETE FROM hr_interview_feedback WHERE id = ?", feedbackID},
			{"DELETE FROM hr_interview_panelists WHERE id = ?", panelistID},
			{"DELETE FROM hr_interview_panelists WHERE id = ?", adjacentPanelistID},
			{"DELETE FROM hr_interviews WHERE id = ?", interviewID},
			{"DELETE FROM hr_interviews WHERE id = ?", overlapInterviewID},
			{"DELETE FROM hr_application_stage_history WHERE application_id = ?", appID},
			{"DELETE FROM hr_applications WHERE id = ?", secondApplicationID},
			{"DELETE FROM hr_applications WHERE id = ?", appID},
			{"DELETE FROM hr_requisitions WHERE id = ?", reqID},
			{"DELETE FROM hr_candidates WHERE id = ?", candidateID},
			{"DELETE FROM hr_candidates WHERE id = ?", secondCandidateID},
			{"DELETE FROM hr_workers WHERE id = ?", workerID},
			{"DELETE FROM hr_jobs WHERE id = ?", jobID},
			{"DELETE FROM hr_job_families WHERE id = ?", familyID},
		} {
			if stmt.v != "" {
				if err := database.WithContext(cleanupCtx).Exec(stmt.q, stmt.v).Error; err != nil {
					t.Errorf("cleanup %s: %v", stmt.q, err)
				}
			}
		}
		for _, agg := range []string{assessID, offerID, feedbackID, panelistID, adjacentPanelistID, interviewID, appID, reqID, candidateID} {
			if agg != "" {
				if err := database.WithContext(cleanupCtx).Exec("DELETE FROM hr_outbox_events WHERE aggregate_id = ?", agg).Error; err != nil {
					t.Errorf("cleanup outbox %s: %v", agg, err)
				}
			}
		}
	})

	if err := database.WithContext(ctx).Exec(
		"INSERT INTO hr_job_families (id, family_code, family_name, job_domain) VALUES (?, ?, ?, ?)",
		familyID, "RDT-FAM-"+familyID, "Tech", "engineering",
	).Error; err != nil {
		t.Fatal(err)
	}
	if err := database.WithContext(ctx).Exec(
		"INSERT INTO hr_jobs (id, job_family_id, job_code, job_title, valid_from) VALUES (?, ?, ?, ?, ?)",
		jobID, familyID, "RDT-JOB-"+jobID, "Backend Engineer", "2026-01-01",
	).Error; err != nil {
		t.Fatal(err)
	}
	if err := database.WithContext(ctx).Exec(
		`INSERT INTO hr_workers (id, worker_number, person_party_reference_id, worker_category, status,
		 primary_legal_entity_reference_id, iam_principal_reference_id, hire_date, service_start_date)
		 VALUES (?, ?, ?, 'employee', 'active', ?, ?, '2026-01-01', '2026-01-01')`,
		workerID, "RDT-W-"+workerID, familyID, legalID, assessorID,
	).Error; err != nil {
		t.Fatal(err)
	}
	if err := database.WithContext(ctx).Exec(`INSERT INTO hr_requisitions
		(id, requisition_number, job_id, organization_unit_reference_id, hiring_manager_principal_id,
		requisition_type, status, employment_category, approved_at, approved_by_principal_id,
		opened_at, idempotency_key)
		VALUES (?, ?, ?, ?, ?, 'new_headcount', 'open', 'employee',
		NOW() - INTERVAL '2 hours', ?, NOW() - INTERVAL '1 hour', ?)`,
		reqID, "REQ-DTL-"+familyID, jobID, unitID, assessorID, otherID, "idem-req-dtl-"+familyID,
	).Error; err != nil {
		t.Fatal(err)
	}

	retention := time.Now().AddDate(0, 2, 0).Format("2006-01-02")
	candRes, err := client.CreateCandidate(ctx, &pb.CreateCandidateRequest{Input: &pb.CandidateInput{
		CandidateNumber: text("CAND-DTL-" + familyID), PreferredName: text("Detail Cand"),
		PrivacyNoticeVersion: text("v1"), NoticeAcknowledgedAt: timestamppb.Now(),
		DataRetentionUntil: &retention, CandidateStatus: text("active"),
	}})
	if err != nil {
		t.Fatalf("create candidate: %v", err)
	}
	candidateID = candRes.GetItem().GetId()

	appRes, err := client.CreateApplication(ctx, &pb.CreateApplicationRequest{Input: &pb.ApplicationInput{
		CandidateId: &candidateID, RequisitionId: &reqID, ApplicationNumber: text("APP-DTL-" + familyID),
		IdempotencyKey: text("idem-app-dtl-" + familyID),
	}})
	if err != nil {
		t.Fatalf("create application: %v", err)
	}
	appID = appRes.GetItem().GetId()
	if _, err := client.StartApplicationScreening(ctx, &common.IDRequest{Id: appID}); err != nil {
		t.Fatalf("start screening: %v", err)
	}
	if _, err := client.ShortlistApplication(ctx, &pb.ShortlistApplicationRequest{Id: appID, ReasonCode: "QUALIFIED"}); err != nil {
		t.Fatalf("shortlist application: %v", err)
	}

	round := int32(1)
	intType, tz := "technical", "Asia/Jakarta"
	startAt := timestamppb.New(time.Now().Add(48 * time.Hour))
	endAt := timestamppb.New(startAt.AsTime().Add(time.Hour))
	intRes, err := client.CreateInterview(ctx, &pb.CreateInterviewRequest{Input: &pb.InterviewInput{
		ApplicationId: &appID, InterviewRound: &round, InterviewType: &intType,
		ScheduledStart: startAt, ScheduledEnd: endAt, CandidateTimezone: &tz,
	}})
	if err != nil {
		t.Fatalf("create interview: %v", err)
	}
	interviewID = intRes.GetItem().GetId()

	panelRequest := &pb.CreateInterviewPanelistRequest{Input: &pb.InterviewPanelistInput{
		InterviewId: &interviewID, InterviewerPrincipalId: &assessorID,
	}}
	lastDay := startAt.AsTime().UTC().AddDate(0, 0, -1).Format("2006-01-02")
	if err := database.WithContext(ctx).Table("hr_workers").Where("id = ?", workerID).
		Update("termination_date", lastDay).Error; err != nil {
		t.Fatalf("set panelist last working date: %v", err)
	}
	if _, err := client.CreateInterviewPanelist(ctx, panelRequest); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("assign panelist whose employment ends before interview = %v", err)
	}
	if err := database.WithContext(ctx).Table("hr_workers").Where("id = ?", workerID).
		Update("termination_date", endAt.AsTime().UTC().Format("2006-01-02")).Error; err != nil {
		t.Fatalf("set panelist last working date to interview date: %v", err)
	}
	panelRequest.Input.FeedbackDueAt = timestamppb.New(startAt.AsTime().Add(30 * time.Minute))
	if _, err := client.CreateInterviewPanelist(ctx, panelRequest); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("assign panelist with deadline before interview ends = %v", err)
	}
	panelRequest.Input.FeedbackDueAt = intRes.GetItem().GetScheduledEnd()
	panelRes, err := client.CreateInterviewPanelist(ctx, panelRequest)
	if err != nil {
		t.Fatalf("create panelist: %v", err)
	}
	panelistID = panelRes.GetItem().GetId()
	overlapInterviewID, err = ids.NewID()
	if err != nil {
		t.Fatal(err)
	}
	secondCandidateID, err = ids.NewID()
	if err != nil {
		t.Fatal(err)
	}
	secondApplicationID, err = ids.NewID()
	if err != nil {
		t.Fatal(err)
	}
	if err := database.WithContext(ctx).Exec(
		`INSERT INTO hr_candidates (id, candidate_number, privacy_notice_version, notice_acknowledged_at, data_retention_until)
		 SELECT ?, ?, privacy_notice_version, notice_acknowledged_at, data_retention_until FROM hr_candidates WHERE id = ?`,
		secondCandidateID, "CAND-PANEL-"+secondCandidateID, candidateID,
	).Error; err != nil {
		t.Fatalf("insert second candidate fixture: %v", err)
	}
	if err := database.WithContext(ctx).Exec(
		`INSERT INTO hr_applications (id, candidate_id, requisition_id, application_number, idempotency_key)
		 SELECT ?, ?, requisition_id, ?, ? FROM hr_applications WHERE id = ?`,
		secondApplicationID, secondCandidateID, "APP-PANEL-"+secondApplicationID, secondApplicationID, appID,
	).Error; err != nil {
		t.Fatalf("insert second application fixture: %v", err)
	}
	if err := database.WithContext(ctx).Exec(
		`INSERT INTO hr_interviews (id, application_id, interview_round, interview_type, scheduled_start, scheduled_end, status)
		 SELECT ?, ?, interview_round, interview_type, scheduled_start, scheduled_end, status
		 FROM hr_interviews WHERE id = ?`, overlapInterviewID, secondApplicationID, interviewID,
	).Error; err != nil {
		t.Fatalf("insert overlapping interview fixture: %v", err)
	}
	if _, err := client.CreateInterviewPanelist(ctx, &pb.CreateInterviewPanelistRequest{Input: &pb.InterviewPanelistInput{
		InterviewId: &overlapInterviewID, InterviewerPrincipalId: &assessorID,
	}}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("assign panelist to overlapping interview = %v", err)
	}
	if err := database.WithContext(ctx).Table("hr_workers").Where("id = ?", workerID).
		Update("termination_date", nil).Error; err != nil {
		t.Fatalf("restore panelist employment fixture: %v", err)
	}
	if err := database.WithContext(ctx).Exec(
		`UPDATE hr_interviews AS next
		 SET scheduled_start = previous.scheduled_end, scheduled_end = previous.scheduled_end + INTERVAL '1 hour'
		 FROM hr_interviews AS previous WHERE next.id = ? AND previous.id = ?`, overlapInterviewID, interviewID,
	).Error; err != nil {
		t.Fatalf("move second interview to adjacent slot: %v", err)
	}
	adjacentPanelist, err := client.CreateInterviewPanelist(ctx, &pb.CreateInterviewPanelistRequest{Input: &pb.InterviewPanelistInput{
		InterviewId: &overlapInterviewID, InterviewerPrincipalId: &assessorID,
	}})
	if err != nil || adjacentPanelist.GetItem().GetId() == "" {
		t.Fatalf("assign panelist to adjacent interview: %+v, %v", adjacentPanelist, err)
	}
	adjacentPanelistID = adjacentPanelist.GetItem().GetId()

	var parallelInterviewIDs []string
	var parallelPanelistID string
	t.Cleanup(func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		tx := database.WithContext(cleanupCtx)
		if err := tx.Exec("DELETE FROM hr_interviews WHERE id IN ?", parallelInterviewIDs).Error; err != nil {
			t.Errorf("cleanup parallel interview fixtures: %v", err)
		}
		if parallelPanelistID != "" {
			if err := tx.Exec("DELETE FROM hr_outbox_events WHERE aggregate_id = ?", parallelPanelistID).Error; err != nil {
				t.Errorf("cleanup parallel panelist outbox: %v", err)
			}
		}
	})
	for _, applicationID := range []string{appID, secondApplicationID} {
		id, err := ids.NewID()
		if err != nil {
			t.Fatal(err)
		}
		parallelInterviewIDs = append(parallelInterviewIDs, id)
		if err := database.WithContext(ctx).Exec(
			`INSERT INTO hr_interviews (id, application_id, interview_round, interview_type, scheduled_start, scheduled_end)
			 VALUES (?, ?, 3, 'technical', ?, ?)`,
			id, applicationID, endAt.AsTime().Add(2*time.Hour), endAt.AsTime().Add(3*time.Hour),
		).Error; err != nil {
			t.Fatalf("insert parallel interview fixture: %v", err)
		}
	}
	type panelistResult struct {
		id  string
		err error
	}
	results := make(chan panelistResult, 2)
	startAssignments := make(chan struct{})
	for _, id := range parallelInterviewIDs {
		go func() {
			<-startAssignments
			res, err := client.CreateInterviewPanelist(ctx, &pb.CreateInterviewPanelistRequest{Input: &pb.InterviewPanelistInput{
				InterviewId: &id, InterviewerPrincipalId: &assessorID,
			}})
			results <- panelistResult{id: res.GetItem().GetId(), err: err}
		}()
	}
	close(startAssignments)
	var conflicts int
	for range parallelInterviewIDs {
		result := <-results
		switch status.Code(result.err) {
		case codes.OK:
			parallelPanelistID = result.id
		case codes.FailedPrecondition:
			conflicts++
		default:
			t.Fatalf("unexpected concurrent panelist assignment: %v", result.err)
		}
	}
	if parallelPanelistID == "" || conflicts != 1 {
		t.Fatalf("parallel panelist requests: accepted ID=%q, conflicts=%d", parallelPanelistID, conflicts)
	}
	var reservations int64
	if err := database.WithContext(ctx).Table("hr_interview_panelists").
		Where("interview_id IN ? AND interviewer_principal_id = ?", parallelInterviewIDs, assessorID).
		Count(&reservations).Error; err != nil || reservations != 1 {
		t.Fatalf("parallel panelist reservations = %d, %v", reservations, err)
	}

	fbVer := int32(1)
	feedbackRequest := &pb.CreateInterviewFeedbackRequest{Input: &pb.InterviewFeedbackInput{
		InterviewId: &interviewID, InterviewerPrincipalId: &assessorID, FeedbackVersion: &fbVer,
		OverallRecommendation: text("strong_yes"), CompetencyScores: &common.JsonDocument{Utf8: []byte(`{"system_design":5}`)},
		ContentDigest: bytes.Repeat([]byte{0xab}, 32), VisibilityClass: text("restricted"),
	}}
	if _, err := client.CreateInterviewFeedback(ctx, feedbackRequest); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("feedback before interview starts: %v", err)
	}
	if err := database.WithContext(ctx).Exec("UPDATE hr_interviews SET scheduled_start = NOW() - INTERVAL '1 minute' WHERE id = ?", interviewID).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := client.StartInterview(ctx, &common.IDRequest{Id: interviewID}); err != nil {
		t.Fatalf("start interview before submitting feedback: %v", err)
	}
	policy.setPrincipal(otherID)
	if _, err := client.CreateInterviewFeedback(ctx, feedbackRequest); status.Code(err) != codes.PermissionDenied {
		t.Fatalf("feedback impersonating an assigned panelist: %v", err)
	}
	policy.setPrincipal(assessorID)
	fbRes, err := client.CreateInterviewFeedback(ctx, feedbackRequest)
	if err != nil {
		t.Fatalf("create feedback: %v", err)
	}
	feedbackID = fbRes.GetItem().GetId()

	editFeedback := &pb.UpdateInterviewFeedbackRequest{
		Id: feedbackID, Input: &pb.InterviewFeedbackInput{OverallRecommendation: text("yes")},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"overall_recommendation"}},
	}
	policy.setPrincipal(otherID)
	if _, err := client.UpdateInterviewFeedback(ctx, editFeedback); status.Code(err) != codes.PermissionDenied {
		t.Fatalf("edit someone else's feedback: %v", err)
	}
	if _, err := client.DeleteInterviewFeedback(ctx, &common.IDRequest{Id: feedbackID}); status.Code(err) != codes.PermissionDenied {
		t.Fatalf("delete someone else's feedback: %v", err)
	}
	policy.setPrincipal(assessorID)

	// Panelis tidak boleh dihapus jika sudah memberi umpan balik.
	if _, err := client.DeleteInterviewPanelist(ctx, &common.IDRequest{Id: panelistID}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("delete panelist with feedback = %v, want FailedPrecondition", status.Code(err))
	}

	completedInterview, err := client.CompleteInterview(ctx, &common.IDRequest{Id: interviewID})
	if err != nil || completedInterview.GetItem().GetStatus() != "completed" {
		t.Fatalf("complete interview with panel feedback RPC: %+v, %v", completedInterview, err)
	}
	if _, err := client.UpdateInterviewFeedback(ctx, editFeedback); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("edit feedback after interview completion: %v", err)
	}

	attempt := int32(1)
	assessRes, err := client.CreateApplicationAssessment(ctx, &pb.CreateApplicationAssessmentRequest{Input: &pb.ApplicationAssessmentInput{
		ApplicationId: &appID, AssessmentCode: text("CODING"), AttemptNumber: &attempt,
		RubricDocumentVersionReferenceId: &rubricID, RubricDigest: bytes.Repeat([]byte{0xef}, 32),
		Status: text("planned"), AssessorPrincipalId: &assessorID, ScheduledAt: startAt,
	}})
	if err != nil {
		t.Fatalf("create assessment: %v", err)
	}
	assessID = assessRes.GetItem().GetId()

	// Pihak bukan penilai ditolak.
	policy.setPrincipal(otherID)
	evDigest := bytes.Repeat([]byte{0x12}, 32)
	if _, err := client.CompleteApplicationAssessment(ctx, &pb.CompleteApplicationAssessmentRequest{
		Id: assessID, NormalizedScore: text("88.500"), RecommendationCode: text("proceed"),
		EvidenceDocumentVersionReferenceId: &evDocID, EvidenceDigest: evDigest,
	}); status.Code(err) != codes.PermissionDenied {
		t.Fatalf("non-assessor completion = %v, want PermissionDenied", status.Code(err))
	}

	// Penilai berhak menyelesaikan asesmen, lalu pengulangan kedua ditolak.
	policy.setPrincipal(assessorID)
	completeRequest := &pb.CompleteApplicationAssessmentRequest{
		Id: assessID, NormalizedScore: text("88.500"), RecommendationCode: text("proceed"),
		EvidenceDocumentVersionReferenceId: &evDocID, EvidenceDigest: evDigest,
	}
	if _, err := client.CompleteApplicationAssessment(ctx, completeRequest); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("complete assessment before scheduled time = %v", err)
	}
	if _, err := client.UpdateApplicationAssessment(ctx, &pb.UpdateApplicationAssessmentRequest{
		Id: assessID, Input: &pb.ApplicationAssessmentInput{},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"scheduled_at"}},
	}); err != nil {
		t.Fatalf("clear assessment schedule: %v", err)
	}
	var originalOpening time.Time
	if err := database.WithContext(ctx).Table("hr_requisitions").Where("id = ?", reqID).
		Select("opened_at").Scan(&originalOpening).Error; err != nil {
		t.Fatalf("read assessment requisition opening: %v", err)
	}
	if err := database.WithContext(ctx).Table("hr_requisitions").Where("id = ?", reqID).
		Update("opened_at", time.Now().Add(time.Hour)).Error; err != nil {
		t.Fatalf("set future assessment requisition opening: %v", err)
	}
	if _, err := client.CompleteApplicationAssessment(ctx, completeRequest); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("complete assessment before requisition opens: %v", err)
	}
	if err := database.WithContext(ctx).Table("hr_requisitions").Where("id = ?", reqID).
		Update("opened_at", originalOpening).Error; err != nil {
		t.Fatalf("restore assessment requisition opening: %v", err)
	}
	doneAssess, err := client.CompleteApplicationAssessment(ctx, completeRequest)
	if err != nil || doneAssess.GetItem().GetStatus() != "completed" {
		t.Fatalf("complete assessment: res=%+v err=%v", doneAssess, err)
	}
	if _, err := client.CompleteApplicationAssessment(ctx, &pb.CompleteApplicationAssessmentRequest{
		Id: assessID, NormalizedScore: text("90.000"), RecommendationCode: text("proceed"),
		EvidenceDocumentVersionReferenceId: &evDocID, EvidenceDigest: evDigest,
	}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("repeat completion = %v, want FailedPrecondition", status.Code(err))
	}
	offerVer := int32(1)
	offerRes, err := client.CreateOffer(ctx, &pb.CreateOfferRequest{Input: &pb.OfferInput{
		ApplicationId: &appID, OfferNumber: text("OFR-DTL-" + familyID), OfferVersion: &offerVer,
		OfferStatus: text("draft"), ProposedJobId: &jobID, ProposedLegalEntityReferenceId: &legalID,
		ProposedContractType: text("permanent"), CurrencyCode: text("IDR"),
		ExpiresAt: timestamppb.New(time.Now().Add(48 * time.Hour)),
	}})
	if err != nil {
		t.Fatalf("create draft offer: %v", err)
	}
	offerID = offerRes.GetItem().GetId()
	if offerRes.GetItem().GetCreatedByPrincipalId() != assessorID || offerRes.GetItem().GetCreatedAt() == nil {
		t.Fatalf("offer draft missing its author or creation time: %+v", offerRes.GetItem())
	}
	if err := database.WithContext(ctx).Exec("UPDATE hr_offers SET approved_by_principal_id = ?, approved_at = NOW() WHERE id = ?", assessorID, offerID).Error; err == nil {
		t.Fatal("database accepted offer self-approval")
	}

	withdraw := &pb.WithdrawOfferRequest{Id: offerID, ReasonCode: "ROLE_CLOSED"}
	if _, err := client.WithdrawOffer(ctx, withdraw); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("withdraw unissued offer: %v", err)
	}
	approval := &pb.ApproveOfferRequest{Id: offerID, CompensationDocumentVersionId: rubricID}
	if _, err := client.ApproveOffer(ctx, approval); status.Code(err) != codes.PermissionDenied {
		t.Fatalf("approve own offer draft: %v", err)
	}
	policy.setPrincipal(otherID)
	withoutVerifiers, err := appservice.NewRegistry(repos, operation.Dependencies{
		Policy: policy, Integrity: security.NewIntegrity(validators),
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := withoutVerifiers.Recruitment.ApproveOffer(ctx, offerID, rubricID); !errors.Is(err, hrerrors.ErrUnavailable) {
		t.Fatalf("approve offer without a verified owner adapter: %v", err)
	}
	offerCompensation.mismatched = true
	if _, err := client.ApproveOffer(ctx, approval); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("approve offer with mismatched compensation evidence: %v", err)
	}
	offerCompensation.mismatched = false
	approvedOffer, err := client.ApproveOffer(ctx, approval)
	if err != nil || approvedOffer.GetItem().GetOfferStatus() != "approved" ||
		approvedOffer.GetItem().GetApprovedByPrincipalId() != otherID ||
		approvedOffer.GetItem().GetProposedCompensationDocumentVersionReferenceId() != rubricID {
		t.Fatalf("approve offer with verified compensation evidence: %+v, %v", approvedOffer, err)
	}
	issue := &pb.IssueOfferRequest{
		Id: offerID, DocumentId: familyID, DocumentVersionId: rubricID, ContentDigest: bytes.Repeat([]byte{1}, 32),
	}
	if _, err := client.IssueOffer(ctx, issue); status.Code(err) != codes.PermissionDenied {
		t.Fatalf("unassigned recruiter issued offer: %v", err)
	}
	policy.setPrincipal(assessorID)
	offerCompensation.mismatched = true
	if _, err := client.IssueOffer(ctx, issue); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("issue offer with mismatched document digest: %v", err)
	}
	offerCompensation.mismatched = false
	issued, err := client.IssueOffer(ctx, issue)
	if err != nil || issued.GetItem().GetOfferStatus() != "issued" ||
		issued.GetItem().GetOfferDocumentVersionReferenceId() != rubricID ||
		!bytes.Equal(issued.GetItem().GetOfferContentDigest(), issue.ContentDigest) {
		t.Fatalf("issue offer with verified document: %+v, %v", issued, err)
	}
	applicationAfterIssue, err := client.GetApplicationByID(ctx, &common.IDRequest{Id: appID})
	if err != nil || applicationAfterIssue.GetItem().GetStatus() != "offer" {
		t.Fatalf("issue offer did not advance application: %+v, %v", applicationAfterIssue, err)
	}
	conflictingOfferID, err := ids.NewID()
	if err != nil {
		t.Fatal(err)
	}
	insertActiveOffer := func(tx *gorm.DB) error {
		return tx.Exec(`INSERT INTO hr_offers
			(id, application_id, offer_number, offer_version, offer_status, proposed_job_id,
			 proposed_legal_entity_reference_id, proposed_contract_type,
			 approved_by_principal_id, approved_at, issued_at, offer_document_version_reference_id, offer_content_digest)
			SELECT ?, application_id, offer_number || '-R2', offer_version + 1, 'issued', proposed_job_id,
			 proposed_legal_entity_reference_id, proposed_contract_type,
			 approved_by_principal_id, approved_at, issued_at, offer_document_version_reference_id, offer_content_digest
			FROM hr_offers WHERE id = ?`, conflictingOfferID, offerID).Error
	}
	if err := insertActiveOffer(database.WithContext(ctx)); err == nil || !strings.Contains(err.Error(), "uq_hr_offer_active_application") {
		t.Fatalf("database accepted another active offer on the same application: %v", err)
	}

	policy.setPrincipal(otherID)
	if _, err := client.WithdrawOffer(ctx, withdraw); status.Code(err) != codes.PermissionDenied {
		t.Fatalf("unassigned recruiter withdrew offer: %v", err)
	}
	policy.setPrincipal(assessorID)
	responseEvidenceID, err := ids.NewID()
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name, statement string
		args            []any
	}{
		{"missing evidence", "UPDATE hr_offers SET offer_status = 'accepted', candidate_response = 'accepted', responded_at = NOW() WHERE id = ?", []any{offerID}},
		{"expired offer", "UPDATE hr_offers SET offer_status = 'accepted', candidate_response = 'accepted', responded_at = expires_at, candidate_response_evidence_id = ? WHERE id = ?", []any{responseEvidenceID, offerID}},
	} {
		if err := database.WithContext(ctx).Exec(test.statement, test.args...).Error; err == nil ||
			!strings.Contains(err.Error(), "ck_hr_offer_candidate_response_evidence") {
			t.Fatalf("database accepted candidate response with %s: %v", test.name, err)
		}
	}
	offerResponse.decision = "accepted"
	offerResponse.respondedAt = issued.GetItem().GetIssuedAt().AsTime()
	accept := &pb.AcceptOfferRequest{Id: offerID, ResponseEvidenceId: responseEvidenceID}
	offerResponse.mismatched = true
	if _, err := client.AcceptOffer(ctx, accept); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("accept offer with unverified candidate: %v", err)
	}
	offerResponse.mismatched = false
	issuedAt := issued.GetItem().GetIssuedAt().AsTime()
	for _, respondedAt := range []time.Time{
		issuedAt.Add(-time.Second),
		time.Now().Add(time.Minute),
		issued.GetItem().GetExpiresAt().AsTime(),
	} {
		offerResponse.respondedAt = respondedAt
		if _, err := client.AcceptOffer(ctx, accept); status.Code(err) != codes.FailedPrecondition {
			t.Fatalf("accept offer with invalid response time %v: %v", respondedAt, err)
		}
	}
	offerResponse.respondedAt = issuedAt
	decline := &pb.DeclineOfferRequest{Id: offerID, ResponseEvidenceId: responseEvidenceID}
	if _, err := client.DeclineOffer(ctx, decline); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("decline offer with acceptance evidence: %v", err)
	}
	accepted, err := client.AcceptOffer(ctx, accept)
	if err != nil || accepted.GetItem().GetOfferStatus() != "accepted" || accepted.GetItem().GetCandidateResponse() != "accepted" {
		t.Fatalf("record verified acceptance: %+v, %v", accepted, err)
	}
	if _, err := client.AcceptOffer(ctx, accept); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("repeat acceptance: %v", err)
	}

	if _, err := client.WithdrawOffer(ctx, withdraw); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("withdraw after candidate response: %v", err)
	}
	if err := database.WithContext(ctx).Exec("UPDATE hr_offers SET offer_status = 'issued', candidate_response = NULL, responded_at = NULL, candidate_response_evidence_id = NULL WHERE id = ?", offerID).Error; err != nil {
		t.Fatal(err)
	}
	withdrawn, err := client.WithdrawOffer(ctx, withdraw)
	if err != nil || withdrawn.GetItem().GetOfferStatus() != "withdrawn" ||
		withdrawn.GetItem().GetWithdrawnByPrincipalId() != assessorID ||
		withdrawn.GetItem().GetWithdrawalReasonCode() != "ROLE_CLOSED" ||
		withdrawn.GetItem().GetOfferDocumentVersionReferenceId() != rubricID {
		t.Fatalf("withdraw offer preserving issued evidence: %+v, %v", withdrawn, err)
	}
	fixtureRollback := errors.New("rollback replacement offer fixture")
	if err := database.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := insertActiveOffer(tx); err != nil {
			return err
		}
		return fixtureRollback
	}); !errors.Is(err, fixtureRollback) {
		t.Fatalf("withdrawal did not release the active offer slot: %v", err)
	}

	if err := insertActiveOffer(database.WithContext(ctx)); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := database.WithContext(ctx).Exec("DELETE FROM hr_offers WHERE id = ?", conflictingOfferID).Error; err != nil {
			t.Errorf("clean decline fixture: %v", err)
		}
	}()
	if err := database.WithContext(ctx).Exec("UPDATE hr_offers SET expires_at = NOW() + INTERVAL '1 hour' WHERE id = ?", conflictingOfferID).Error; err != nil {
		t.Fatal(err)
	}
	offerResponse.decision = "declined"
	declined, err := client.DeclineOffer(ctx, &pb.DeclineOfferRequest{Id: conflictingOfferID, ResponseEvidenceId: responseEvidenceID})
	if err != nil || declined.GetItem().GetOfferStatus() != "declined" {
		t.Fatalf("record verified decline: %+v, %v", declined, err)
	}

	if _, err := client.WithdrawOffer(ctx, withdraw); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("repeat offer withdrawal: %v", err)
	}

	revision, err := client.CreateOffer(ctx, &pb.CreateOfferRequest{Input: &pb.OfferInput{
		ApplicationId: &appID, OfferNumber: text("OFR-R3-" + familyID), OfferVersion: proto.Int32(3),
		ProposedJobId: &jobID, ProposedLegalEntityReferenceId: &legalID, ProposedContractType: text("permanent"),
		CurrencyCode: text("IDR"), ExpiresAt: timestamppb.New(time.Now().Add(time.Hour)),
	}})
	if err != nil {
		t.Fatal(err)
	}
	revisionID := revision.GetItem().GetId()
	defer func() {
		if err := database.WithContext(ctx).Exec("DELETE FROM hr_offers WHERE id = ?", revisionID).Error; err != nil {
			t.Errorf("clean offer revision fixture: %v", err)
		}
	}()
	policy.setPrincipal(otherID)
	if _, err := client.ApproveOffer(ctx, &pb.ApproveOfferRequest{Id: revisionID, CompensationDocumentVersionId: rubricID}); err != nil {
		t.Fatal(err)
	}
	policy.setPrincipal(assessorID)
	if err := database.WithContext(ctx).Exec("UPDATE hr_jobs SET is_active = false WHERE id = ?", jobID).Error; err != nil {
		t.Fatal(err)
	}
	revisionIssue := proto.CloneOf(issue)
	revisionIssue.Id = revisionID
	if _, err := client.IssueOffer(ctx, revisionIssue); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("issue approved revision after job became inactive: %v", err)
	}
	storedRevision, err := client.GetOfferByID(ctx, &common.IDRequest{Id: revisionID})
	if err != nil || storedRevision.GetItem().GetOfferStatus() != "approved" {
		t.Fatalf("failed issue changed approved revision: %+v, %v", storedRevision, err)
	}

}
