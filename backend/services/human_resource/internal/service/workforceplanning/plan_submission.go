package service

import (
	"context"
	"strings"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

// SubmitWorkforcePlan mengirim draf milik pengusul untuk pemeriksaan.
func (s *Service) SubmitWorkforcePlan(ctx context.Context, id string) (*entity.WorkforcePlan, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.workforcePlan.Command(ctx, id, "submit", func(_ context.Context, plan *entity.WorkforcePlan, actor operation.Actor, now time.Time) (*entity.WorkforcePlan, error) {
		if actor.Kind != "human" || !strings.EqualFold(actor.PrincipalID, plan.CreatedByPrincipalID) {
			return nil, hrerrors.ErrForbidden
		}
		if plan.Status != "draft" {
			return nil, hrerrors.ErrWorkflowRequired
		}
		if entity.CompareDates(now.UTC(), plan.PeriodEnd) > 0 {
			return nil, hrerrors.ErrReferenceConflict
		}
		plan.Status = "review"
		return plan, nil
	})
}

// CancelWorkforcePlan menarik rencana yang belum disetujui dan menyimpan riwayatnya.
func (s *Service) CancelWorkforcePlan(ctx context.Context, id string) (*entity.WorkforcePlan, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.workforcePlan.Command(ctx, id, "cancel", func(_ context.Context, plan *entity.WorkforcePlan, actor operation.Actor, _ time.Time) (*entity.WorkforcePlan, error) {
		if actor.Kind != "human" || !strings.EqualFold(actor.PrincipalID, plan.CreatedByPrincipalID) {
			return nil, hrerrors.ErrForbidden
		}
		if plan.Status != "draft" && plan.Status != "review" {
			return nil, hrerrors.ErrWorkflowRequired
		}
		plan.Status = "cancelled"
		return plan, nil
	})
}

// SupersedeWorkforcePlan menautkan rencana disetujui ke versi pengganti yang sah.
func (s *Service) SupersedeWorkforcePlan(ctx context.Context, id, replacementID, reasonCode string) (*entity.WorkforcePlan, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	if err := entity.ValidateUUID("replacement_plan_id", replacementID); err != nil {
		return nil, err
	}
	reasonCode = strings.TrimSpace(reasonCode)
	if reasonCode == "" || len(reasonCode) > 80 || strings.EqualFold(id, replacementID) {
		return nil, hrerrors.ErrInvalidInput
	}
	return s.workforcePlan.Command(ctx, id, "supersede", func(_ context.Context, plan *entity.WorkforcePlan, actor operation.Actor, _ time.Time) (*entity.WorkforcePlan, error) {
		if actor.Kind != "human" {
			return nil, hrerrors.ErrForbidden
		}
		if plan.Status != "approved" {
			return nil, hrerrors.ErrWorkflowRequired
		}
		plan.Status = "superseded"
		plan.SupersededByPlanID = &replacementID
		plan.SupersessionReasonCode = &reasonCode
		return plan, nil
	})
}
