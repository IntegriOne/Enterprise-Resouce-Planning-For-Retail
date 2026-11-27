package security

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
)

type talentFoundationStub struct {
	err error
}

func (s *talentFoundationStub) VerifyPositionReferences(context.Context, PositionReferenceCheck) error {
	return s.err
}

func (s *talentFoundationStub) VerifyWorkerReferences(context.Context, WorkerReferenceCheck) error {
	return s.err
}

func (s *talentFoundationStub) VerifyLegalEntity(context.Context, string) error {
	return s.err
}

func (s *talentFoundationStub) VerifyOrganizationUnit(context.Context, string, string) error {
	return s.err
}

func (s *talentFoundationStub) VerifyLocation(context.Context, string, string) error {
	return s.err
}

func (s *talentFoundationStub) VerifyParty(context.Context, string, bool) error {
	return s.err
}

func TestTalentCycleAssessmentCalibrationAndPoolValidators(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 11, 24, 0, 0, 0, 0, time.UTC)
	frameworkID := "77777777-7777-4777-8777-777777777701"
	dimensionID := "77777777-7777-4777-8777-777777777702"
	legalEntityID := "77777777-7777-4777-8777-777777777703"
	cycleID := "77777777-7777-4777-8777-777777777704"
	workerID := "77777777-7777-4777-8777-777777777705"
	ownerID := "77777777-7777-4777-8777-777777777706"
	assessorID := "77777777-7777-4777-8777-777777777707"
	calibratorID := "77777777-7777-4777-8777-777777777708"
	workerPrincipalID := "77777777-7777-4777-8777-777777777709"
	assessmentID := "77777777-7777-4777-8777-777777777710"
	scoreID := "77777777-7777-4777-8777-777777777711"
	sessionID := "77777777-7777-4777-8777-777777777712"
	poolID := "77777777-7777-4777-8777-777777777713"
	membershipID := "77777777-7777-4777-8777-777777777714"
	refID := "77777777-7777-4777-8777-777777777799"

	frameworks := newTalentStoreStub(map[string]entity.TalentFramework{
		frameworkID: {ID: frameworkID, EffectiveFrom: now.AddDate(-1, 0, 0), ApprovedByPrincipalID: ownerID, ApprovedAt: now.AddDate(-1, 0, 0)},
	})
	dimensions := newTalentStoreStub(map[string]entity.TalentDimension{
		dimensionID: {ID: dimensionID, FrameworkID: frameworkID, DimensionCode: "LEAD", WeightPercent: "100", IsRequired: true},
	})
	workers := &talentWorkerStub{rows: map[string]entity.Worker{
		workerID:                               {ID: workerID, Status: entity.WorkerStatusActive, IAMPrincipalReferenceID: &workerPrincipalID, PrimaryLegalEntityReferenceID: legalEntityID},
		"77777777-7777-4777-8777-777777777721": {ID: "77777777-7777-4777-8777-777777777721", Status: entity.WorkerStatusActive, IAMPrincipalReferenceID: &ownerID, PrimaryLegalEntityReferenceID: legalEntityID},
		"77777777-7777-4777-8777-777777777722": {ID: "77777777-7777-4777-8777-777777777722", Status: entity.WorkerStatusActive, IAMPrincipalReferenceID: &assessorID, PrimaryLegalEntityReferenceID: legalEntityID},
		"77777777-7777-4777-8777-777777777723": {ID: "77777777-7777-4777-8777-777777777723", Status: entity.WorkerStatusActive, IAMPrincipalReferenceID: &calibratorID, PrimaryLegalEntityReferenceID: legalEntityID},
	}}
	cycles := newTalentStoreStub(map[string]entity.TalentReviewCycle{})
	assessments := newTalentStoreStub(map[string]entity.TalentAssessment{})
	scores := newTalentStoreStub(map[string]entity.TalentAssessmentScore{})
	sessions := newTalentStoreStub(map[string]entity.TalentCalibrationSession{})
	decisions := newTalentStoreStub(map[string]entity.TalentCalibrationDecision{})
	pools := newTalentStoreStub(map[string]entity.TalentPool{})
	memberships := newTalentStoreStub(map[string]entity.TalentPoolMembership{})
	reviews := newTalentStoreStub(map[string]entity.PerformanceReview{})
	jobs := newTalentStoreStub(map[string]entity.Job{})
	foundation := &talentFoundationStub{}

	if err := (&talentFrameworkValidator{}).ValidateMutation(t.Context(), "hr_talent_frameworks", nil, &entity.TalentFramework{ID: frameworkID}); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("framework CRUD = %v, want ErrWorkflowRequired", err)
	}

	cycleVal := &talentReviewCycleValidator{
		frameworks: frameworks, dimensions: dimensions, cycles: cycles,
		assessments: assessments, sessions: sessions, workers: workers, foundation: foundation,
	}
	draftCycle := &entity.TalentReviewCycle{
		ID: cycleID, CycleCode: "2026-Q4", FrameworkID: frameworkID, LegalEntityReferenceID: legalEntityID,
		PeriodStart: now.AddDate(0, -1, 0), PeriodEnd: now.AddDate(0, 1, 0), Status: "draft", OwnerPrincipalID: ownerID,
	}
	if err := cycleVal.ValidateMutation(t.Context(), "hr_talent_review_cycles", nil, draftCycle); err != nil {
		t.Fatalf("valid draft cycle failed: %v", err)
	}
	openCycle := *draftCycle
	openCycle.Status = "open"
	if err := cycleVal.ValidateMutation(t.Context(), "hr_talent_review_cycles", draftCycle, &openCycle); err != nil {
		t.Fatalf("launch cycle failed: %v", err)
	}
	cycles.rows[cycleID] = openCycle

	assessVal := &talentAssessmentValidator{
		frameworks: frameworks, dimensions: dimensions, cycles: cycles,
		assessments: assessments, scores: scores, workers: workers, reviews: reviews,
	}
	scoreVal := &talentAssessmentScoreValidator{
		assessments: assessments, dimensions: dimensions, cycles: cycles, scores: scores, decisions: decisions,
	}
	draftAssess := &entity.TalentAssessment{
		ID: assessmentID, WorkerID: workerID, TalentReviewCycleID: cycleID, FrameworkID: frameworkID,
		AssessmentVersion: 1, Status: "draft", AssessedByPrincipalID: assessorID, AssessedAt: now,
		EvidenceDocumentVersionReferenceID: refID, EvidenceDigest: []byte{1},
	}
	if err := assessVal.ValidateMutation(t.Context(), "hr_talent_assessments", nil, draftAssess); err != nil {
		t.Fatalf("valid draft assessment failed: %v", err)
	}
	assessments.rows[assessmentID] = *draftAssess

	// Finalisasi harus gagal sebelum skor dimensi wajib ada.
	finalAssess := *draftAssess
	finalAssess.Status = "finalized"
	finalAssess.FinalizedAt = &now
	finalAssess.FinalizedByPrincipalID = &calibratorID
	if err := assessVal.ValidateMutation(t.Context(), "hr_talent_assessments", draftAssess, &finalAssess); !errors.Is(err, hrerrors.ErrInvalidInput) {
		t.Fatalf("finalize without score = %v, want ErrInvalidInput", err)
	}

	score := &entity.TalentAssessmentScore{
		ID: scoreID, AssessmentID: assessmentID, FrameworkID: frameworkID, DimensionID: dimensionID,
		Score: 4, EvidenceDocumentVersionReferenceID: refID, EvidenceDigest: []byte{1},
	}
	if err := scoreVal.ValidateMutation(t.Context(), "hr_talent_assessment_scores", nil, score); err != nil {
		t.Fatalf("valid score failed: %v", err)
	}
	scores.rows[scoreID] = *score

	if err := assessVal.ValidateMutation(t.Context(), "hr_talent_assessments", draftAssess, &finalAssess); err != nil {
		t.Fatalf("finalize assessment with complete score failed: %v", err)
	}
	assessments.rows[assessmentID] = finalAssess

	// Skor terkunci setelah assessment final.
	if err := scoreVal.ValidateMutation(t.Context(), "hr_talent_assessment_scores", score, nil); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("delete finalized score = %v, want ErrWorkflowRequired", err)
	}

	calCycle := openCycle
	calCycle.Status = "calibration"
	cycles.rows[cycleID] = calCycle
	sessVal := &talentCalibrationSessionValidator{cycles: cycles, sessions: sessions, decisions: decisions, workers: workers}
	openSession := &entity.TalentCalibrationSession{
		ID: sessionID, TalentReviewCycleID: cycleID, SessionCode: "CAL-1", ChairPrincipalID: calibratorID,
		ScheduledAt: now, Status: "open",
	}
	sessions.rows[sessionID] = *openSession
	if err := sessVal.ValidateMutation(t.Context(), "hr_talent_calibration_sessions", openSession, openSession); err != nil {
		t.Fatalf("stayingOpen session failed: %v", err)
	}

	decVal := &talentCalibrationDecisionValidator{
		sessions: sessions, cycles: cycles, assessments: assessments, scores: scores, decisions: decisions, workers: workers,
	}
	dec := &entity.TalentCalibrationDecision{
		ID: "77777777-7777-4777-8777-777777777715", SessionID: sessionID, TalentReviewCycleID: cycleID,
		AssessmentID: assessmentID, DimensionID: dimensionID, OriginalScore: 4, CalibratedScore: 5,
		RationaleDocumentVersionReferenceID: refID, RationaleDigest: []byte{1},
		DecidedByPrincipalID: calibratorID, DecidedAt: now,
	}
	if err := decVal.ValidateMutation(t.Context(), "hr_talent_calibration_decisions", nil, dec); err != nil {
		t.Fatalf("valid calibration decision failed: %v", err)
	}
	selfDec := *dec
	selfDec.DecidedByPrincipalID = assessorID
	if err := decVal.ValidateMutation(t.Context(), "hr_talent_calibration_decisions", nil, &selfDec); !errors.Is(err, hrerrors.ErrForbidden) {
		t.Fatalf("self calibration = %v, want ErrForbidden", err)
	}

	poolVal := &talentPoolValidator{pools: pools, memberships: memberships, jobs: jobs, workers: workers, foundation: foundation}
	pool := &entity.TalentPool{
		ID: poolID, PoolCode: "HIPO", PoolVersion: 1, PoolName: "High Potential",
		LegalEntityReferenceID: legalEntityID, EligibilityDocumentVersionReferenceID: refID,
		EligibilityDigest: []byte{1}, ValidFrom: now.AddDate(0, -1, 0), OwnerPrincipalID: ownerID,
	}
	if err := poolVal.ValidateMutation(t.Context(), "hr_talent_pools", nil, pool); err != nil {
		t.Fatalf("valid pool failed: %v", err)
	}
	pools.rows[poolID] = *pool

	memVal := &talentPoolMembershipValidator{pools: pools, workers: workers, assessments: assessments, memberships: memberships}
	nominated := &entity.TalentPoolMembership{
		ID: membershipID, TalentPoolID: poolID, WorkerID: workerID, AssessmentID: assessmentID,
		NominatedByPrincipalID: assessorID, Status: "nominated", ValidFrom: now,
	}
	if err := memVal.ValidateMutation(t.Context(), "hr_talent_pool_memberships", nil, nominated); err != nil {
		t.Fatalf("valid nominated membership failed: %v", err)
	}
	active := *nominated
	active.Status = "active"
	active.ApprovedByPrincipalID = &calibratorID
	active.ApprovedAt = &now
	if err := memVal.ValidateMutation(t.Context(), "hr_talent_pool_memberships", nominated, &active); err != nil {
		t.Fatalf("approve membership failed: %v", err)
	}
	memberships.rows[membershipID] = active

	overlapNom := *nominated
	overlapNom.ID = "77777777-7777-4777-8777-777777777716"
	overlapNom.ValidFrom = now.AddDate(0, 0, 5)
	overlapActive := overlapNom
	overlapActive.Status = "active"
	overlapActive.ApprovedByPrincipalID = &calibratorID
	overlapActive.ApprovedAt = &now
	if err := memVal.ValidateMutation(t.Context(), "hr_talent_pool_memberships", &overlapNom, &overlapActive); !errors.Is(err, hrerrors.ErrAlreadyExists) {
		t.Fatalf("overlapping active membership = %v, want ErrAlreadyExists", err)
	}
}

func newTalentStoreStub[T any](initial map[string]T) *talentStoreStub[T] {
	rows := make(map[string]T, len(initial))
	for k, v := range initial {
		rows[k] = v
	}
	return &talentStoreStub[T]{rows: rows}
}

type talentStoreStub[T any] struct {
	rows map[string]T
}

func (s *talentStoreStub[T]) FindByIDForUpdate(_ context.Context, id string) (*T, error) {
	v, ok := s.rows[id]
	if !ok {
		return nil, hrerrors.ErrNotFound
	}
	copyVal := v
	return &copyVal, nil
}

func (s *talentStoreStub[T]) SelectWhere(_ context.Context, _ []domainrepo.Filter, _ int, _ int) ([]T, error) {
	out := make([]T, 0, len(s.rows))
	for _, v := range s.rows {
		out = append(out, v)
	}
	return out, nil
}

func (s *talentStoreStub[T]) CountWhere(_ context.Context, filters []domainrepo.Filter) (int64, error) {
	if len(filters) == 0 {
		return int64(len(s.rows)), nil
	}
	for _, f := range filters {
		if f.Column == "status" {
			return 0, nil
		}
	}
	return int64(len(s.rows)), nil
}

type talentWorkerStub struct {
	rows map[string]entity.Worker
}

func (s *talentWorkerStub) FindByIDForUpdate(_ context.Context, id string) (*entity.Worker, error) {
	v, ok := s.rows[id]
	if !ok {
		return nil, hrerrors.ErrNotFound
	}
	copyVal := v
	return &copyVal, nil
}

func (s *talentWorkerStub) SelectWhere(_ context.Context, filters []domainrepo.Filter, _ int, _ int) ([]entity.Worker, error) {
	out := make([]entity.Worker, 0, len(s.rows))
	for _, v := range s.rows {
		if len(filters) == 1 && filters[0].Column == "iam_principal_reference_id" && len(filters[0].Values) == 1 {
			want, _ := filters[0].Values[0].(string)
			if v.IAMPrincipalReferenceID == nil || *v.IAMPrincipalReferenceID != want {
				continue
			}
		}
		out = append(out, v)
	}
	return out, nil
}
