package security

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository"
)

type recruitmentDetailFoundationStub struct {
	legalEntities map[string]error
}

func (s *recruitmentDetailFoundationStub) VerifyPositionReferences(context.Context, PositionReferenceCheck) error {
	return nil
}

func (s *recruitmentDetailFoundationStub) VerifyWorkerReferences(context.Context, WorkerReferenceCheck) error {
	return nil
}

func (s *recruitmentDetailFoundationStub) VerifyLegalEntity(_ context.Context, legalEntityID string) error {
	if err, ok := s.legalEntities[legalEntityID]; ok {
		return err
	}
	return nil
}

func (s *recruitmentDetailFoundationStub) VerifyOrganizationUnit(context.Context, string, string) error {
	return nil
}

func (s *recruitmentDetailFoundationStub) VerifyLocation(context.Context, string, string) error {
	return nil
}

func (s *recruitmentDetailFoundationStub) VerifyParty(context.Context, string, bool) error {
	return nil
}

type pagedInterviewFeedbackStub []entity.InterviewFeedback

func (s pagedInterviewFeedbackStub) SelectWhere(_ context.Context, _ []domainrepo.Filter, limit, offset int) ([]entity.InterviewFeedback, error) {
	if offset >= len(s) {
		return nil, nil
	}
	return s[offset:min(offset+min(limit, domainrepo.MaxLimit), len(s))], nil
}

func (s pagedInterviewFeedbackStub) CountWhere(context.Context, []domainrepo.Filter) (int64, error) {
	return int64(len(s)), nil
}

func TestRecruitmentDetailInterviewPanelistAndFeedbackValidation(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()
	interviewID := "33333333-3333-4333-8333-333333333301"
	applicationID := "33333333-3333-4333-8333-333333333302"
	principalID := "33333333-3333-4333-8333-333333333303"
	workerID := "33333333-3333-4333-8333-333333333304"
	panelistID := "33333333-3333-4333-8333-333333333305"
	feedbackID := "33333333-3333-4333-8333-333333333306"
	legalEntityID := "33333333-3333-4333-8333-333333333307"

	interviews := newRecordStoreStub(func(v *entity.Interview) string { return v.ID }, map[string]entity.Interview{
		interviewID: {ID: interviewID, ApplicationID: applicationID, InterviewRound: 1, InterviewType: "technical", Status: "planned"},
	})
	applications := newRecordStoreStub(func(v *entity.Application) string { return v.ID }, map[string]entity.Application{
		applicationID: {ID: applicationID, CandidateID: "33333333-3333-4333-8333-333333333308", RequisitionID: "33333333-3333-4333-8333-333333333309", Status: entity.ApplicationStatusInterview, AppliedAt: now.Add(-time.Hour)},
	})
	workers := newRecordStoreStub(func(v *entity.Worker) string { return v.ID }, map[string]entity.Worker{
		workerID: {ID: workerID, Status: entity.WorkerStatusActive, IAMPrincipalReferenceID: &principalID, PrimaryLegalEntityReferenceID: legalEntityID},
	})
	panelists := newRecordStoreStub(func(v *entity.InterviewPanelist) string { return v.ID }, nil)
	feedback := newRecordStoreStub(func(v *entity.InterviewFeedback) string { return v.ID }, nil)

	panelistValidator := &interviewPanelistValidator{
		interviews: interviews, applications: applications, workers: workers, panelists: panelists, feedback: feedback,
	}
	feedbackValidator := &interviewFeedbackValidator{
		interviews: interviews, applications: applications, workers: workers, panelists: panelists, feedback: feedback,
	}

	// Tolak panelis jika scorecard DRM diisi karena DRM belum tersedia.
	templateID := "33333333-3333-4333-8333-333333333399"
	draftPanelist := &entity.InterviewPanelist{
		ID: panelistID, InterviewID: interviewID, InterviewerPrincipalID: principalID,
		ScorecardTemplateReferenceID: &templateID,
	}
	if err := panelistValidator.ValidateMutation(t.Context(), "hr_interview_panelists", nil, draftPanelist); !errors.Is(err, hrerrors.ErrUnavailable) {
		t.Fatalf("scorecard DRM reference should fail closed: %v", err)
	}
	draftPanelist.ScorecardTemplateReferenceID = nil
	if err := panelistValidator.ValidateMutation(t.Context(), "hr_interview_panelists", nil, draftPanelist); err != nil {
		t.Fatalf("valid panelist failed: %v", err)
	}

	interview := interviews.rows[interviewID]
	started := interview
	started.Status = "in_progress"
	interviews.rows[interviewID] = started
	if err := panelistValidator.ValidateMutation(t.Context(), "hr_interview_panelists", nil, draftPanelist); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("add panelist while interview is in progress: %v", err)
	}
	alreadyStartedAt := now.Add(-time.Minute)
	started.Status = "planned"
	started.ScheduledStart = &alreadyStartedAt
	interviews.rows[interviewID] = started
	if err := panelistValidator.ValidateMutation(t.Context(), "hr_interview_panelists", draftPanelist, nil); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("remove panelist after scheduled start: %v", err)
	}
	interviews.rows[interviewID] = interview

	// Feedback ditolak sebelum panelis terdaftar di hr_interview_panelists.
	fb := &entity.InterviewFeedback{
		ID: feedbackID, InterviewID: interviewID, InterviewerPrincipalID: principalID,
		FeedbackVersion: 1, SubmittedAt: now, OverallRecommendation: "strong_yes",
		CompetencyScores: entity.JSON(`{"go":5}`), VisibilityClass: "restricted",
	}
	if err := feedbackValidator.ValidateMutation(t.Context(), "hr_interview_feedback", nil, fb); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("feedback before interview starts: %v", err)
	}
	started.Status = "in_progress"
	interviews.rows[interviewID] = started
	if err := feedbackValidator.ValidateMutation(t.Context(), "hr_interview_feedback", nil, fb); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("feedback without assigned panelist = %v, want ErrReferenceConflict", err)
	}

	panelists.rows[panelistID] = *draftPanelist
	if err := feedbackValidator.ValidateMutation(t.Context(), "hr_interview_feedback", nil, fb); err != nil {
		t.Fatalf("valid feedback failed: %v", err)
	}
	feedback.rows[feedbackID] = *fb

	interviews.rows[interviewID] = interview
	// Panelis dengan feedback tidak boleh dihapus.
	if err := panelistValidator.ValidateMutation(t.Context(), "hr_interview_panelists", draftPanelist, nil); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("delete panelist with feedback = %v, want ErrReferenceConflict", err)
	}

	versions := make(pagedInterviewFeedbackStub, domainrepo.MaxLimit+1)
	for i := range versions {
		versions[i] = *fb
		versions[i].ID = fmt.Sprintf("33333333-3333-4333-8333-%012d", i+1)
		versions[i].FeedbackVersion = int32(i + 1)
	}
	latest := versions[len(versions)-1]
	feedbackValidator.feedback = versions
	if err := feedbackValidator.ValidateMutation(t.Context(), "hr_interview_feedback", &latest, nil); err != nil {
		t.Fatalf("delete latest feedback with more than one page: %v", err)
	}
	atPageBoundary := versions[len(versions)-2]
	if err := feedbackValidator.ValidateMutation(t.Context(), "hr_interview_feedback", &atPageBoundary, nil); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("delete superseded feedback with newer revision on next page: %v", err)
	}
}

func TestRecruitmentDetailOfferAndAssessmentValidation(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()
	candidateID := "44444444-4444-4444-8444-444444444401"
	requisitionID := "44444444-4444-4444-8444-444444444402"
	applicationID := "44444444-4444-4444-8444-444444444403"
	jobID := "44444444-4444-4444-8444-444444444404"
	legalEntityID := "44444444-4444-4444-8444-444444444405"
	principalID := "44444444-4444-4444-8444-444444444406"
	workerID := "44444444-4444-4444-8444-444444444407"
	retention := now.AddDate(0, 1, 0)
	ackAt := now.Add(-2 * time.Hour)
	openedAt := now.Add(-2 * time.Hour)
	approvedAt := now.Add(-3 * time.Hour)
	appliedAt := now.Add(-time.Hour)

	candidates := newRecordStoreStub(func(v *entity.Candidate) string { return v.ID }, map[string]entity.Candidate{
		candidateID: {ID: candidateID, CandidateStatus: "active", PrivacyNoticeVersion: "v1", NoticeAcknowledgedAt: &ackAt, DataRetentionUntil: &retention},
	})
	requisitions := newRecordStoreStub(func(v *entity.Requisition) string { return v.ID }, map[string]entity.Requisition{
		requisitionID: {ID: requisitionID, Status: "open", ApprovedAt: &approvedAt, ApprovedByPrincipalID: &principalID, OpenedAt: &openedAt},
	})
	applications := newRecordStoreStub(func(v *entity.Application) string { return v.ID }, map[string]entity.Application{
		applicationID: {ID: applicationID, CandidateID: candidateID, RequisitionID: requisitionID, Status: entity.ApplicationStatusShortlisted, AppliedAt: appliedAt},
	})
	jobs := newRecordStoreStub(func(v *entity.Job) string { return v.ID }, map[string]entity.Job{
		jobID: {ID: jobID, IsActive: true, ValidFrom: now.AddDate(0, 0, -1)},
	})
	offers := newRecordStoreStub(func(v *entity.Offer) string { return v.ID }, nil)
	workers := newRecordStoreStub(func(v *entity.Worker) string { return v.ID }, map[string]entity.Worker{
		workerID: {ID: workerID, Status: entity.WorkerStatusActive, IAMPrincipalReferenceID: &principalID, PrimaryLegalEntityReferenceID: legalEntityID},
	})
	assessments := newRecordStoreStub(func(v *entity.ApplicationAssessment) string { return v.ID }, nil)

	foundation := &recruitmentDetailFoundationStub{legalEntities: map[string]error{
		"44444444-4444-4444-8444-444444444499": hrerrors.ErrReferenceConflict,
	}}
	offerVal := &offerValidator{
		applications: applications, candidates: candidates, requisitions: requisitions,
		jobs: jobs, offers: offers, foundation: foundation,
	}
	currency := "IDR"
	startDate := now.AddDate(0, 0, 7)
	draftOffer := &entity.Offer{
		ID: "44444444-4444-4444-8444-444444444410", ApplicationID: applicationID, OfferNumber: "OFF-1",
		OfferVersion: 1, OfferStatus: "draft", ProposedJobID: jobID, CreatedByPrincipalID: &principalID, CreatedAt: &now,
		ProposedLegalEntityReferenceID: "44444444-4444-4444-8444-444444444499",
		ProposedStartDate:              &startDate, ProposedContractType: "permanent", CurrencyCode: &currency,
	}
	if err := offerVal.ValidateMutation(t.Context(), "hr_offers", nil, draftOffer); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("inactive EF legal entity = %v, want ErrReferenceConflict", err)
	}
	draftOffer.ProposedLegalEntityReferenceID = legalEntityID
	if err := offerVal.ValidateMutation(t.Context(), "hr_offers", nil, draftOffer); err != nil {
		t.Fatalf("valid draft offer failed: %v", err)
	}

	// Tolak offer tanpa verifier EF.
	noEFValidator := &offerValidator{
		applications: applications, candidates: candidates, requisitions: requisitions,
		jobs: jobs, offers: offers, foundation: nil,
	}
	if err := noEFValidator.ValidateMutation(t.Context(), "hr_offers", nil, draftOffer); !errors.Is(err, hrerrors.ErrUnavailable) {
		t.Fatalf("nil foundation verifier = %v, want ErrUnavailable", err)
	}

	// Validasi draf asesmen dan transisi penyelesaiannya.
	assessmentVal := &applicationAssessmentValidator{
		applications: applications, candidates: candidates, requisitions: requisitions,
		workers: workers, assessments: assessments,
	}
	planned := &entity.ApplicationAssessment{
		ID: "44444444-4444-4444-8444-444444444411", ApplicationID: applicationID,
		AssessmentCode: "TECH-1", AttemptNumber: 1,
		RubricDocumentVersionReferenceID: "44444444-4444-4444-8444-444444444412",
		RubricDigest:                     bytes.Repeat([]byte{0x11}, 32),
		Status:                           "planned",
	}
	if err := assessmentVal.ValidateMutation(t.Context(), "hr_application_assessments", nil, planned); err != nil {
		t.Fatalf("valid planned assessment failed: %v", err)
	}
	score := entity.Decimal("88.500")
	rec := "RECOMMEND"
	completed := *planned
	completed.Status = "completed"
	completed.AssessorPrincipalID = &principalID
	completed.CompletedAt = &now
	completed.NormalizedScore = &score
	completed.RecommendationCode = &rec
	if err := assessmentVal.ValidateMutation(t.Context(), "hr_application_assessments", planned, &completed); err != nil {
		t.Fatalf("complete assessment transition failed: %v", err)
	}
	if err := assessmentVal.ValidateMutation(t.Context(), "hr_application_assessments", &completed, &completed); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("editing completed assessment = %v, want ErrWorkflowRequired", err)
	}
}

func newRecordStoreStub[T any](idFn func(*T) string, initial map[string]T) *recordStoreStub[T] {
	rows := make(map[string]T, len(initial))
	for k, v := range initial {
		rows[k] = v
	}
	return &recordStoreStub[T]{idFn: idFn, rows: rows}
}

type recordStoreStub[T any] struct {
	repository.Registry
	idFn func(*T) string
	rows map[string]T
}

func (s *recordStoreStub[T]) FindByIDForUpdate(_ context.Context, id string) (*T, error) {
	v, ok := s.rows[id]
	if !ok {
		return nil, hrerrors.ErrNotFound
	}
	copyVal := v
	return &copyVal, nil
}

func (s *recordStoreStub[T]) SelectWhere(_ context.Context, filters []domainrepo.Filter, _ int, _ int) ([]T, error) {
	out := make([]T, 0, len(s.rows))
	for _, row := range s.rows {
		out = append(out, row)
	}
	_ = filters
	return out, nil
}

func (s *recordStoreStub[T]) CountWhere(_ context.Context, filters []domainrepo.Filter) (int64, error) {
	for _, f := range filters {
		if len(f.Values) == 0 && strings.TrimSpace(f.Column) == "" {
			return 0, nil
		}
	}
	return int64(len(s.rows)), nil
}
