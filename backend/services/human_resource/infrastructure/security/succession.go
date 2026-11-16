package security

import (
	"context"
	"reflect"
	"strings"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository"
)

type successionPositionReader interface {
	FindByIDForUpdate(context.Context, string) (*entity.Position, error)
}

type successionPlanReader interface {
	SelectWhere(context.Context, []domainrepo.Filter, int, int) ([]entity.SuccessionPlan, error)
}

type successionCandidateCounter interface {
	CountWhere(context.Context, []domainrepo.Filter) (int64, error)
}

type successionLockedPlanReader interface {
	FindByIDForUpdate(context.Context, string) (*entity.SuccessionPlan, error)
}

type successionCandidateReader interface {
	SelectWhere(context.Context, []domainrepo.Filter, int, int) ([]entity.SuccessionCandidate, error)
}

type successionWorkerReader interface {
	FindByIDForUpdate(context.Context, string) (*entity.Worker, error)
}

type successionAssessmentReader interface {
	FindByIDForUpdate(context.Context, string) (*entity.TalentAssessment, error)
}

type successionDevelopmentReader interface {
	FindByIDForUpdate(context.Context, string) (*entity.DevelopmentPlan, error)
}

type successionPlanValidator struct {
	positions  successionPositionReader
	plans      successionPlanReader
	candidates successionCandidateCounter
}

type successionCandidateValidator struct {
	plans       successionLockedPlanReader
	positions   successionPositionReader
	workers     successionWorkerReader
	assessments successionAssessmentReader
	development successionDevelopmentReader
	candidates  successionCandidateReader
}

// SuccessionValidators memasang aturan draf rencana dan nominasi suksesi lokal HRM.
func SuccessionValidators(repos *repository.Registry) (map[string]MutationValidator, error) {
	if repos == nil || operation.RequireDependency(repos.Position) != nil ||
		operation.RequireDependency(repos.SuccessionPlan) != nil ||
		operation.RequireDependency(repos.SuccessionCandidate) != nil ||
		operation.RequireDependency(repos.Worker) != nil ||
		operation.RequireDependency(repos.TalentAssessment) != nil ||
		operation.RequireDependency(repos.DevelopmentPlan) != nil {
		return nil, hrerrors.ErrUnavailable
	}
	return map[string]MutationValidator{
		"hr_succession_plans": &successionPlanValidator{
			positions: repos.Position, plans: repos.SuccessionPlan, candidates: repos.SuccessionCandidate,
		},
		"hr_succession_candidates": &successionCandidateValidator{
			plans: repos.SuccessionPlan, positions: repos.Position, workers: repos.Worker,
			assessments: repos.TalentAssessment, development: repos.DevelopmentPlan,
			candidates: repos.SuccessionCandidate,
		},
	}, nil
}

// ValidateMutation menjaga nominasi hanya selama rencana suksesi belum disetujui.
func (v *successionCandidateValidator) ValidateMutation(ctx context.Context, table string, before, after entity.Validatable) error {
	if v == nil || table != "hr_succession_candidates" || operation.RequireDependency(v.plans) != nil ||
		operation.RequireDependency(v.positions) != nil || operation.RequireDependency(v.workers) != nil ||
		operation.RequireDependency(v.assessments) != nil || operation.RequireDependency(v.development) != nil ||
		operation.RequireDependency(v.candidates) != nil {
		return hrerrors.ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	var previous *entity.SuccessionCandidate
	if before != nil {
		var ok bool
		previous, ok = before.(*entity.SuccessionCandidate)
		if !ok || previous == nil {
			return hrerrors.ErrInvalidInput
		}
	}
	if previous == nil && after == nil {
		return hrerrors.ErrInvalidInput
	}
	candidate := previous
	if after != nil {
		var ok bool
		candidate, ok = after.(*entity.SuccessionCandidate)
		if !ok || candidate == nil {
			return hrerrors.ErrInvalidInput
		}
		if previous != nil && (previous.ID != candidate.ID ||
			previous.SuccessionPlanID != candidate.SuccessionPlanID ||
			previous.CandidateWorkerID != candidate.CandidateWorkerID) {
			return hrerrors.ErrWorkflowRequired
		}
		if candidate.RiskOfLoss != nil || candidate.ReadinessRationale == nil ||
			strings.TrimSpace(*candidate.ReadinessRationale) == "" {
			return hrerrors.ErrWorkflowRequired
		}
		switch candidate.ReadinessCode {
		case "ready_now", "1_year", "2_3_years", "development_needed":
		default:
			return hrerrors.ErrWorkflowRequired
		}
		if entity.CompareDates(candidate.ReadinessAssessedAt, time.Now()) > 0 ||
			(candidate.ReadinessCode != "development_needed" && candidate.TalentAssessmentID == nil) {
			return hrerrors.ErrReferenceConflict
		}
	}
	plan, err := v.plans.FindByIDForUpdate(ctx, candidate.SuccessionPlanID)
	if err != nil {
		return workerAssessmentReferenceError(err)
	}
	if plan == nil {
		return hrerrors.ErrReferenceConflict
	}
	if plan.ApprovedAt != nil || plan.ApprovedByPrincipalID != nil ||
		plan.ConfidentialityClass != "highly_restricted" {
		return hrerrors.ErrWorkflowRequired
	}
	position, err := v.positions.FindByIDForUpdate(ctx, plan.PositionID)
	if err != nil {
		return workerAssessmentReferenceError(err)
	}
	if position == nil || (after != nil && (position.PositionStatus != "active" ||
		entity.CompareDates(candidate.ReadinessAssessedAt, position.ValidFrom) < 0 ||
		(position.ValidTo != nil && entity.CompareDates(candidate.ReadinessAssessedAt, *position.ValidTo) > 0))) {
		return hrerrors.ErrReferenceConflict
	}
	worker, err := v.workers.FindByIDForUpdate(ctx, candidate.CandidateWorkerID)
	if err != nil {
		return workerAssessmentReferenceError(err)
	}
	if worker == nil || worker.PrimaryLegalEntityReferenceID != position.LegalEntityReferenceID ||
		(after != nil && worker.Status != entity.WorkerStatusActive) {
		return hrerrors.ErrReferenceConflict
	}
	if after == nil {
		return nil
	}
	if candidate.TalentAssessmentID != nil {
		assessment, err := v.assessments.FindByIDForUpdate(ctx, *candidate.TalentAssessmentID)
		if err != nil {
			return workerAssessmentReferenceError(err)
		}
		if assessment == nil || assessment.WorkerID != worker.ID || assessment.Status != "finalized" ||
			assessment.FinalizedAt == nil || assessment.FinalizedByPrincipalID == nil ||
			entity.CompareDates(*assessment.FinalizedAt, candidate.ReadinessAssessedAt) > 0 {
			return hrerrors.ErrReferenceConflict
		}
	}
	if candidate.DevelopmentPlanID != nil {
		development, err := v.development.FindByIDForUpdate(ctx, *candidate.DevelopmentPlanID)
		if err != nil {
			return workerAssessmentReferenceError(err)
		}
		if development == nil || development.WorkerID != worker.ID ||
			(development.Status != "active" && development.Status != "completed") ||
			development.ApprovedAt == nil || development.ApprovedByPrincipalID == nil ||
			entity.CompareDates(*development.ApprovedAt, candidate.ReadinessAssessedAt) > 0 {
			return hrerrors.ErrReferenceConflict
		}
	}
	matches, err := v.candidates.SelectWhere(ctx, []domainrepo.Filter{
		{Column: "succession_plan_id", Values: []any{candidate.SuccessionPlanID}},
		{Column: "candidate_worker_id", Values: []any{candidate.CandidateWorkerID}},
	}, 2, 0)
	if err != nil {
		return err
	}
	for _, existing := range matches {
		if existing.ID != candidate.ID {
			return hrerrors.ErrAlreadyExists
		}
	}
	return nil
}

// ValidateMutation membatasi CRUD pada draf yang posisinya aktif dan belum memiliki kandidat.
func (v *successionPlanValidator) ValidateMutation(ctx context.Context, table string, before, after entity.Validatable) error {
	if v == nil || table != "hr_succession_plans" || operation.RequireDependency(v.positions) != nil ||
		operation.RequireDependency(v.plans) != nil || operation.RequireDependency(v.candidates) != nil {
		return hrerrors.ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	var previous *entity.SuccessionPlan
	if before != nil {
		var ok bool
		previous, ok = before.(*entity.SuccessionPlan)
		if !ok || previous == nil {
			return hrerrors.ErrInvalidInput
		}
		if previous.ApprovedByPrincipalID != nil || previous.ApprovedAt != nil || previous.SuccessionRisk != nil {
			return hrerrors.ErrWorkflowRequired
		}
	}
	plan := previous
	if after != nil {
		var ok bool
		plan, ok = after.(*entity.SuccessionPlan)
		if !ok || plan == nil || (previous != nil && (previous.ID != plan.ID ||
			previous.PositionID != plan.PositionID || previous.PlanCode != plan.PlanCode ||
			previous.PlanVersion != plan.PlanVersion)) {
			return hrerrors.ErrWorkflowRequired
		}
		if plan.ApprovedByPrincipalID != nil || plan.ApprovedAt != nil || plan.SuccessionRisk != nil ||
			plan.ConfidentialityClass != "highly_restricted" {
			return hrerrors.ErrWorkflowRequired
		}
	} else if previous == nil {
		return hrerrors.ErrInvalidInput
	}

	position, err := v.positions.FindByIDForUpdate(ctx, plan.PositionID)
	if err != nil {
		return workerAssessmentReferenceError(err)
	}
	if position == nil || (after != nil && (position.PositionStatus != "active" ||
		entity.CompareDates(plan.ReviewDate, position.ValidFrom) < 0 ||
		(position.ValidTo != nil && entity.CompareDates(plan.ReviewDate, *position.ValidTo) > 0))) {
		return hrerrors.ErrReferenceConflict
	}
	count, err := v.candidates.CountWhere(ctx, []domainrepo.Filter{{Column: "succession_plan_id", Values: []any{plan.ID}}})
	if err != nil {
		return err
	}
	if count > 0 && (after == nil || !reflect.DeepEqual(previous, plan)) {
		return hrerrors.ErrReferenceConflict
	}
	if after == nil {
		return nil
	}
	matches, err := v.plans.SelectWhere(ctx, []domainrepo.Filter{
		{Column: "position_id", Values: []any{plan.PositionID}},
		{Column: "plan_code", Values: []any{plan.PlanCode}},
		{Column: "plan_version", Values: []any{plan.PlanVersion}},
	}, 2, 0)
	if err != nil {
		return err
	}
	for _, existing := range matches {
		if existing.ID != plan.ID {
			return hrerrors.ErrAlreadyExists
		}
	}
	return nil
}
