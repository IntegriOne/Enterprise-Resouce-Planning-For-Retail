package service

import (
	"context"
	"strings"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

// ApproveLeaveAdjustment memisahkan keputusan koreksi dari pengajuan dan posting saldo.
func (s *Service) ApproveLeaveAdjustment(ctx context.Context, adjustmentID string) (*entity.LeaveAdjustment, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.leaveAdjustment.Command(ctx, adjustmentID, "approve", func(_ context.Context, adjustment *entity.LeaveAdjustment, actor operation.Actor, now time.Time) (*entity.LeaveAdjustment, error) {
		if actor.Kind != "human" || strings.EqualFold(adjustment.RequestedByPrincipalID, actor.PrincipalID) {
			return nil, hrerrors.ErrForbidden
		}
		if adjustment.ApprovedAt != nil || adjustment.ApprovedByPrincipalID != nil || adjustment.RejectedAt != nil {
			return nil, hrerrors.ErrConflict
		}
		adjustment.ApprovedByPrincipalID = &actor.PrincipalID
		adjustment.ApprovedAt = &now
		return adjustment, nil
	})
}

// RejectLeaveAdjustment mencatat penolakan independen tanpa memposting saldo.
func (s *Service) RejectLeaveAdjustment(ctx context.Context, adjustmentID, reasonCode string) (*entity.LeaveAdjustment, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	reasonCode = strings.TrimSpace(reasonCode)
	if reasonCode == "" || len(reasonCode) > 80 {
		return nil, hrerrors.Invalid("reason_code", "has invalid value")
	}
	return s.leaveAdjustment.Command(ctx, adjustmentID, "reject", func(_ context.Context, adjustment *entity.LeaveAdjustment, actor operation.Actor, now time.Time) (*entity.LeaveAdjustment, error) {
		if actor.Kind != "human" || strings.EqualFold(adjustment.RequestedByPrincipalID, actor.PrincipalID) {
			return nil, hrerrors.ErrForbidden
		}
		if adjustment.ApprovedAt != nil || adjustment.ApprovedByPrincipalID != nil || adjustment.RejectedAt != nil {
			return nil, hrerrors.ErrConflict
		}
		adjustment.RejectedAt = &now
		adjustment.RejectedByPrincipalID = &actor.PrincipalID
		adjustment.RejectionReasonCode = &reasonCode
		return adjustment, nil
	})
}
