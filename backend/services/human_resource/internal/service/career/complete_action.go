package service

import (
	"context"
	"strings"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

// CompleteDevelopmentAction mencatat coaching atau mentoring yang dikonfirmasi mentor.
func (s *Service) CompleteDevelopmentAction(ctx context.Context, id string) (*entity.DevelopmentAction, error) {
	if s == nil || operation.RequireDependency(s.workers) != nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.developmentAction.Command(ctx, id, "complete", func(txCtx context.Context, action *entity.DevelopmentAction, actor operation.Actor, now time.Time) (*entity.DevelopmentAction, error) {
		if actor.Kind != "human" {
			return nil, hrerrors.ErrForbidden
		}
		if action.CompletedAt != nil {
			return nil, hrerrors.ErrWorkflowRequired
		}
		plan, err := s.plans.FindByIDForUpdate(txCtx, action.DevelopmentPlanID)
		if err != nil {
			return nil, err
		}
		if plan == nil || plan.Status != "active" || plan.MentorWorkerID == nil || plan.ApprovedByPrincipalID == nil {
			return nil, hrerrors.ErrWorkflowRequired
		}
		if err := s.verifyActiveMentor(txCtx, plan, actor); err != nil {
			return nil, err
		}
		action.CompletedAt = &now
		return action, nil
	})
}

func (s *Service) verifyActiveMentor(ctx context.Context, plan *entity.DevelopmentPlan, actor operation.Actor) error {
	if !strings.EqualFold(actor.PrincipalID, *plan.ApprovedByPrincipalID) {
		return hrerrors.ErrForbidden
	}
	mentor, err := s.workers.FindByIDForUpdate(ctx, *plan.MentorWorkerID)
	if err != nil {
		return err
	}
	if mentor == nil || mentor.Status != entity.WorkerStatusActive || mentor.IAMPrincipalReferenceID == nil ||
		!strings.EqualFold(actor.PrincipalID, *mentor.IAMPrincipalReferenceID) {
		return hrerrors.ErrForbidden
	}
	return nil
}

// CompleteDevelopmentPlan menutup rencana yang seluruh tindakannya telah selesai.
func (s *Service) CompleteDevelopmentPlan(ctx context.Context, id string) (*entity.DevelopmentPlan, error) {
	if s == nil || operation.RequireDependency(s.workers) != nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.developmentPlan.Command(ctx, id, "complete", func(txCtx context.Context, plan *entity.DevelopmentPlan, actor operation.Actor, now time.Time) (*entity.DevelopmentPlan, error) {
		if actor.Kind != "human" {
			return nil, hrerrors.ErrForbidden
		}
		if plan.Status != "active" || plan.MentorWorkerID == nil || plan.ApprovedByPrincipalID == nil ||
			plan.ApprovedAt == nil || plan.ApprovedAt.After(now) || entity.CompareDates(now.UTC(), plan.StartDate) < 0 {
			return nil, hrerrors.ErrWorkflowRequired
		}
		if err := s.verifyActiveMentor(txCtx, plan, actor); err != nil {
			return nil, err
		}
		filters := []domainrepo.Filter{{Column: "development_plan_id", Values: []any{plan.ID}}}
		for offset := 0; ; offset += domainrepo.MaxLimit {
			actions, err := s.actions.SelectWhere(txCtx, filters, domainrepo.MaxLimit, offset)
			if err != nil {
				return nil, err
			}
			for _, action := range actions {
				if action.CompletedAt == nil || action.CompletedAt.Before(*plan.ApprovedAt) || action.CompletedAt.After(now) ||
					entity.CompareDates(action.CompletedAt.UTC(), plan.StartDate) < 0 {
					return nil, hrerrors.ErrReferenceConflict
				}
			}
			if len(actions) < domainrepo.MaxLimit {
				break
			}
		}
		plan.Status = "completed"
		return plan, nil
	})
}
