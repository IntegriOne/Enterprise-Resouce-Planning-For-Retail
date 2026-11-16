package service

import (
	"context"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

// CompleteOnboardingPlan menyelesaikan rencana setelah seluruh tugas wajib memiliki bukti selesai.
func (s *Service) CompleteOnboardingPlan(ctx context.Context, planID string) (*entity.OnboardingPlan, error) {
	if s == nil || s.tasks == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.onboardingPlan.Command(ctx, planID, "complete", func(txCtx context.Context, plan *entity.OnboardingPlan, actor operation.Actor, now time.Time) (*entity.OnboardingPlan, error) {
		if actor.Kind != "human" {
			return nil, hrerrors.ErrForbidden
		}
		if plan.Status != "in_progress" || plan.StartedAt == nil || plan.StartedByPrincipalID == nil ||
			plan.CompletedAt != nil || plan.CompletedByPrincipalID != nil {
			return nil, hrerrors.ErrWorkflowRequired
		}
		total, err := s.tasks.CountWhere(txCtx, []domainrepo.Filter{{Column: "onboarding_plan_id", Values: []any{plan.ID}}})
		if err != nil {
			return nil, err
		}
		if total == 0 {
			return nil, hrerrors.ErrReferenceConflict
		}
		pending, err := s.tasks.CountWhere(txCtx, []domainrepo.Filter{
			{Column: "onboarding_plan_id", Values: []any{plan.ID}},
			{Column: "is_mandatory", Values: []any{true}},
			{Column: "completed_at", Values: []any{nil}},
		})
		if err != nil {
			return nil, err
		}
		if pending > 0 {
			return nil, hrerrors.ErrReferenceConflict
		}
		plan.Status = "completed"
		plan.CompletedAt = &now
		plan.CompletedByPrincipalID = &actor.PrincipalID
		return plan, nil
	})
}
