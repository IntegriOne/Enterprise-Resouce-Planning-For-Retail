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

type lockedSuccessionPlanStub struct{ plan *entity.SuccessionPlan }

func (s *lockedSuccessionPlanStub) FindByIDForUpdate(context.Context, string) (*entity.SuccessionPlan, error) {
	return s.plan, nil
}

type successionWorkerStub struct{ worker *entity.Worker }

func (s *successionWorkerStub) FindByIDForUpdate(context.Context, string) (*entity.Worker, error) {
	return s.worker, nil
}

type successionAssessmentStub struct{ assessment *entity.TalentAssessment }

func (s *successionAssessmentStub) FindByIDForUpdate(context.Context, string) (*entity.TalentAssessment, error) {
	return s.assessment, nil
}

type successionDevelopmentStub struct{ plan *entity.DevelopmentPlan }

func (s *successionDevelopmentStub) FindByIDForUpdate(context.Context, string) (*entity.DevelopmentPlan, error) {
	return s.plan, nil
}

type successionCandidateRowsStub struct{ rows []entity.SuccessionCandidate }

func (s *successionCandidateRowsStub) SelectWhere(context.Context, []domainrepo.Filter, int, int) ([]entity.SuccessionCandidate, error) {
	return s.rows, nil
}

func TestSuccessionCandidateValidatorKeepsNominationInDraftPlan(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()
	reviewDate := now.Truncate(24*time.Hour).AddDate(0, 0, -1)
	wib := time.FixedZone("WIB", 7*60*60)
	evidenceTime := reviewDate.Add(23 * time.Hour).In(wib)
	legalID := "00000000-0000-4000-8000-000000000001"
	workerID := "00000000-0000-4000-8000-000000000002"
	plan := &lockedSuccessionPlanStub{plan: &entity.SuccessionPlan{
		ID:                   "00000000-0000-4000-8000-000000000003",
		PositionID:           "00000000-0000-4000-8000-000000000004",
		ConfidentialityClass: "highly_restricted",
		ReviewDate:           now.AddDate(0, 0, 1),
	}}
	position := &successionPositionStub{position: &entity.Position{
		ID: plan.plan.PositionID, PositionStatus: "active", LegalEntityReferenceID: legalID,
		ValidFrom: now.AddDate(-1, 0, 0),
	}}
	worker := &successionWorkerStub{worker: &entity.Worker{
		ID: workerID, Status: entity.WorkerStatusActive, PrimaryLegalEntityReferenceID: legalID,
	}}
	assessment := &successionAssessmentStub{}
	development := &successionDevelopmentStub{}
	rows := &successionCandidateRowsStub{}
	v := &successionCandidateValidator{plans: plan, positions: position, workers: worker,
		assessments: assessment, development: development, candidates: rows}
	rationale := "Needs development for target position"
	candidate := &entity.SuccessionCandidate{
		ID: "00000000-0000-4000-8000-000000000005", SuccessionPlanID: plan.plan.ID,
		CandidateWorkerID: workerID, ReadinessCode: "development_needed",
		ReadinessAssessedAt: reviewDate, ReadinessRationale: &rationale,
	}
	if err := v.ValidateMutation(t.Context(), "hr_succession_candidates", nil, candidate); err != nil {
		t.Fatalf("create draft nomination: %v", err)
	}
	earlierReview := reviewDate.AddDate(0, 0, -1)
	originalReview := plan.plan.ReviewDate
	plan.plan.ReviewDate = earlierReview
	if err := v.ValidateMutation(t.Context(), "hr_succession_candidates", nil, candidate); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("readiness assessment after succession review: %v", err)
	}
	if err := v.ValidateMutation(t.Context(), "hr_succession_candidates", candidate, nil); err != nil {
		t.Fatalf("delete draft nomination assessed after review: %v", err)
	}
	plan.plan.ReviewDate = originalReview
	position.position.ValidTo = &reviewDate
	if err := v.ValidateMutation(t.Context(), "hr_succession_candidates", nil, candidate); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("position expires before succession plan review: %v", err)
	}
	if err := v.ValidateMutation(t.Context(), "hr_succession_candidates", candidate, nil); err != nil {
		t.Fatalf("delete draft nomination for position ending before review: %v", err)
	}
	position.position.ValidTo = nil
	hireDate := now
	worker.worker.HireDate = &hireDate
	if err := v.ValidateMutation(t.Context(), "hr_succession_candidates", nil, candidate); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("readiness assessment before hire: %v", err)
	}
	if err := v.ValidateMutation(t.Context(), "hr_succession_candidates", candidate, nil); err != nil {
		t.Fatalf("delete outdated nomination: %v", err)
	}
	worker.worker.HireDate = nil
	terminationDate := reviewDate.AddDate(0, 0, -1)
	worker.worker.TerminationDate = &terminationDate
	if err := v.ValidateMutation(t.Context(), "hr_succession_candidates", nil, candidate); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("readiness assessment after termination: %v", err)
	}
	worker.worker.TerminationDate = &now
	if err := v.ValidateMutation(t.Context(), "hr_succession_candidates", nil, candidate); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("employment ends before succession review: %v", err)
	}
	worker.worker.TerminationDate = &plan.plan.ReviewDate
	if err := v.ValidateMutation(t.Context(), "hr_succession_candidates", nil, candidate); err != nil {
		t.Fatalf("employment ends on succession review: %v", err)
	}
	worker.worker.TerminationDate = nil
	updated := *candidate
	updated.ReadinessAssessedAt = now
	if err := v.ValidateMutation(t.Context(), "hr_succession_candidates", candidate, &updated); err != nil {
		t.Fatalf("update draft nomination: %v", err)
	}
	if err := v.ValidateMutation(t.Context(), "hr_succession_candidates", candidate, nil); err != nil {
		t.Fatalf("delete draft nomination: %v", err)
	}
	rows.rows = []entity.SuccessionCandidate{{ID: "00000000-0000-4000-8000-000000000006"}}
	if err := v.ValidateMutation(t.Context(), "hr_succession_candidates", nil, candidate); !errors.Is(err, hrerrors.ErrAlreadyExists) {
		t.Fatalf("duplicate nomination: %v", err)
	}
	rows.rows = nil
	ready := *candidate
	ready.ReadinessCode = "ready_now"
	if err := v.ValidateMutation(t.Context(), "hr_succession_candidates", nil, &ready); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("readiness without assessment: %v", err)
	}
	assessmentID := "00000000-0000-4000-8000-000000000007"
	actorID := "00000000-0000-4000-8000-000000000008"
	ready.TalentAssessmentID = &assessmentID
	assessment.assessment = &entity.TalentAssessment{ID: assessmentID, WorkerID: workerID,
		Status: "finalized", FinalizedAt: &evidenceTime, FinalizedByPrincipalID: &actorID}
	if err := v.ValidateMutation(t.Context(), "hr_succession_candidates", nil, &ready); err != nil {
		t.Fatalf("readiness with finalized assessment: %v", err)
	}
	assessment.assessment.FinalizedAt = &now
	if err := v.ValidateMutation(t.Context(), "hr_succession_candidates", nil, &ready); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("assessment finalized after readiness date: %v", err)
	}
	assessment.assessment.FinalizedAt = &reviewDate
	developmentID := "00000000-0000-4000-8000-000000000009"
	ready.DevelopmentPlanID = &developmentID
	development.plan = &entity.DevelopmentPlan{ID: developmentID, WorkerID: workerID,
		Status: "active", ApprovedAt: &evidenceTime, ApprovedByPrincipalID: &actorID}
	if err := v.ValidateMutation(t.Context(), "hr_succession_candidates", nil, &ready); err != nil {
		t.Fatalf("nomination with approved development plan: %v", err)
	}
	plan.plan.ApprovedAt = &now
	if err := v.ValidateMutation(t.Context(), "hr_succession_candidates", nil, candidate); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("nomination after plan approval: %v", err)
	}
	plan.plan.ApprovedAt = nil
	plan.plan.SuccessionRisk = &rationale
	if err := v.ValidateMutation(t.Context(), "hr_succession_candidates", candidate, nil); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("delete nomination from plan with recorded succession risk: %v", err)
	}
	plan.plan.SuccessionRisk = nil
	worker.worker.Status = entity.WorkerStatusTerminated
	if err := v.ValidateMutation(t.Context(), "hr_succession_candidates", nil, candidate); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("inactive worker nomination: %v", err)
	}
	worker.worker.Status = entity.WorkerStatusActive
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if err := v.ValidateMutation(ctx, "hr_succession_candidates", nil, candidate); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled request: %v", err)
	}
}
