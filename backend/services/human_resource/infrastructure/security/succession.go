package security

import (
	"context"
	"reflect"

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

type successionPlanValidator struct {
	positions  successionPositionReader
	plans      successionPlanReader
	candidates successionCandidateCounter
}

// SuccessionValidators memasang aturan draf rencana suksesi yang seluruh referensinya dimiliki HRM.
func SuccessionValidators(repos *repository.Registry) (map[string]MutationValidator, error) {
	if repos == nil || operation.RequireDependency(repos.Position) != nil ||
		operation.RequireDependency(repos.SuccessionPlan) != nil ||
		operation.RequireDependency(repos.SuccessionCandidate) != nil {
		return nil, hrerrors.ErrUnavailable
	}
	return map[string]MutationValidator{
		"hr_succession_plans": &successionPlanValidator{
			positions: repos.Position, plans: repos.SuccessionPlan, candidates: repos.SuccessionCandidate,
		},
	}, nil
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
