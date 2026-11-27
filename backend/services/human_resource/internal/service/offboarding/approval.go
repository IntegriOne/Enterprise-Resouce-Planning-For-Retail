package service

import (
	"context"
	"strings"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

// ApproveOffboardingPlan mencatat persetujuan independen atas rencana draf.
func (s *Service) ApproveOffboardingPlan(ctx context.Context, planID string) (*entity.OffboardingPlan, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.offboardingPlan.Command(ctx, planID, "approve", func(_ context.Context, plan *entity.OffboardingPlan, actor operation.Actor, now time.Time) (*entity.OffboardingPlan, error) {
		if actor.Kind != "human" || strings.EqualFold(plan.CreatedByPrincipalID, actor.PrincipalID) {
			return nil, hrerrors.ErrForbidden
		}
		if plan.Status != "draft" || plan.ApprovedAt != nil || plan.ApprovedByPrincipalID != nil {
			return nil, hrerrors.ErrWorkflowRequired
		}
		plan.Status = "approved"
		plan.ApprovedAt = &now
		plan.ApprovedByPrincipalID = &actor.PrincipalID
		return plan, nil
	})
}

// StartOffboardingPlan membuka rencana yang telah disetujui dalam periode offboarding.
func (s *Service) StartOffboardingPlan(ctx context.Context, planID string) (*entity.OffboardingPlan, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.offboardingPlan.Command(ctx, planID, "start", func(_ context.Context, plan *entity.OffboardingPlan, actor operation.Actor, now time.Time) (*entity.OffboardingPlan, error) {
		if actor.Kind != "human" {
			return nil, hrerrors.ErrForbidden
		}
		if plan.Status != "approved" || plan.ApprovedAt == nil || plan.ApprovedAt.After(now) {
			return nil, hrerrors.ErrWorkflowRequired
		}
		if entity.CompareDates(now.UTC(), plan.NoticeDate) < 0 ||
			entity.CompareDates(now.UTC(), plan.TerminationEffectiveDate) > 0 {
			return nil, hrerrors.ErrReferenceConflict
		}
		plan.Status = "in_progress"
		return plan, nil
	})
}

// WaiveOffboardingTask membebaskan tugas lokal dengan alasan dan pelaku yang tercatat.
func (s *Service) WaiveOffboardingTask(ctx context.Context, taskID, reason string) (*entity.OffboardingTask, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return nil, hrerrors.Invalid("reason", "must not be blank")
	}
	return s.offboardingTask.Command(ctx, taskID, "waive", func(_ context.Context, task *entity.OffboardingTask, actor operation.Actor, _ time.Time) (*entity.OffboardingTask, error) {
		if actor.Kind != "human" {
			return nil, hrerrors.ErrForbidden
		}
		if task.Status != "pending" || task.OwnerManagement != "human_resource" {
			return nil, hrerrors.ErrWorkflowRequired
		}
		task.Status = "waived"
		task.WaiverReason = &reason
		task.WaivedByPrincipalID = &actor.PrincipalID
		return task, nil
	})
}
