package service

import (
	"context"
	"strings"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

// RejectRequisition mencatat keputusan penolakan atas pengajuan yang sudah diserahkan.
func (s *Service) RejectRequisition(ctx context.Context, requisitionID, reasonCode string) (*entity.Requisition, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	reasonCode = strings.TrimSpace(reasonCode)
	if reasonCode == "" || len(reasonCode) > 80 {
		return nil, hrerrors.Invalid("reason_code", "has invalid value")
	}
	return s.requisition.Command(ctx, requisitionID, "reject", func(_ context.Context, requisition *entity.Requisition, actor operation.Actor, now time.Time) (*entity.Requisition, error) {
		if actor.Kind != "human" {
			return nil, hrerrors.ErrForbidden
		}
		if requisition.Status != "submitted" || requisition.CreatedByPrincipalID == nil ||
			requisition.SubmittedAt == nil || requisition.SubmittedByPrincipalID == nil ||
			requisition.RejectedAt != nil || requisition.RejectionReasonCode != nil ||
			requisition.RejectedByPrincipalID != nil || requisition.ApprovedAt != nil ||
			requisition.ApprovedByPrincipalID != nil || requisition.BudgetReferenceID != nil ||
			requisition.OpenedAt != nil || requisition.ClosedAt != nil ||
			now.Before(*requisition.SubmittedAt) {
			return nil, hrerrors.ErrWorkflowRequired
		}
		if strings.EqualFold(actor.PrincipalID, *requisition.CreatedByPrincipalID) ||
			strings.EqualFold(actor.PrincipalID, *requisition.SubmittedByPrincipalID) {
			return nil, hrerrors.ErrForbidden
		}
		requisition.Status = "rejected"
		requisition.RejectedAt = &now
		requisition.RejectionReasonCode = &reasonCode
		requisition.RejectedByPrincipalID = &actor.PrincipalID
		return requisition, nil
	})
}
