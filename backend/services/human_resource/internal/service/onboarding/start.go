package service

import (
	"context"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

// StartOnboardingPlan membuka rencana draf yang sudah memiliki tugas.
func (s *Service) StartOnboardingPlan(ctx context.Context, planID string) (*entity.OnboardingPlan, error) {
	if s == nil || s.tasks == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.onboardingPlan.Command(ctx, planID, "start", func(txCtx context.Context, plan *entity.OnboardingPlan, actor operation.Actor, now time.Time) (*entity.OnboardingPlan, error) {
		if actor.Kind != "human" {
			return nil, hrerrors.ErrForbidden
		}
		if plan.Status != "draft" || plan.StartedAt != nil || plan.StartedByPrincipalID != nil || plan.CompletedAt != nil || plan.CompletedByPrincipalID != nil {
			return nil, hrerrors.ErrWorkflowRequired
		}
		if entity.CompareDates(now, plan.StartDate) < 0 || plan.TargetCompletionDate != nil && entity.CompareDates(now, *plan.TargetCompletionDate) > 0 {
			return nil, hrerrors.ErrReferenceConflict
		}
		total, err := s.tasks.CountWhere(txCtx, []domainrepo.Filter{{Column: "onboarding_plan_id", Values: []any{plan.ID}}})
		if err != nil {
			return nil, err
		}
		if total == 0 {
			return nil, hrerrors.ErrReferenceConflict
		}
		plan.Status = "in_progress"
		plan.StartedAt = &now
		plan.StartedByPrincipalID = &actor.PrincipalID
		return plan, nil
	})
}
