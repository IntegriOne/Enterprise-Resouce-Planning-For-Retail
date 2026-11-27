package service

import (
	"context"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

// CancelOffboardingPlan membatalkan rencana yang telah disetujui dan belum dimulai.
func (s *Service) CancelOffboardingPlan(ctx context.Context, planID, reason string) (*entity.OffboardingPlan, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	reason = strings.TrimSpace(reason)
	if reason == "" || !utf8.ValidString(reason) || utf8.RuneCountInString(reason) > 300 {
		return nil, hrerrors.Invalid("reason", "must contain 1 to 300 characters")
	}
	return s.offboardingPlan.Command(ctx, planID, "cancel", func(_ context.Context, plan *entity.OffboardingPlan, actor operation.Actor, now time.Time) (*entity.OffboardingPlan, error) {
		if actor.Kind != "human" {
			return nil, hrerrors.ErrForbidden
		}
		if plan.Status != "approved" || plan.ApprovedAt == nil || now.Before(*plan.ApprovedAt) {
			return nil, hrerrors.ErrWorkflowRequired
		}
		plan.Status = "cancelled"
		plan.CancelledAt, plan.CancelledByPrincipalID, plan.CancellationReason = &now, &actor.PrincipalID, &reason
		return plan, nil
	})
}
