package service

import (
	"context"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

// CloseRequisition menutup requisition terbuka setelah seluruh proses aktif selesai.
func (s *Service) CloseRequisition(ctx context.Context, requisitionID string) (*entity.Requisition, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.requisition.Command(ctx, requisitionID, "close", func(_ context.Context, requisition *entity.Requisition, actor operation.Actor, now time.Time) (*entity.Requisition, error) {
		if actor.Kind != "human" {
			return nil, hrerrors.ErrForbidden
		}
		if requisition.Status != "open" || requisition.CreatedByPrincipalID == nil ||
			requisition.SubmittedAt == nil || requisition.SubmittedByPrincipalID == nil ||
			requisition.ApprovedAt == nil || requisition.ApprovedByPrincipalID == nil ||
			requisition.BudgetReferenceID == nil || requisition.OpenedAt == nil || requisition.ClosedAt != nil ||
			now.Before(*requisition.OpenedAt) {
			return nil, hrerrors.ErrWorkflowRequired
		}
		requisition.Status = "closed"
		requisition.ClosedAt = &now
		return requisition, nil
	})
}
