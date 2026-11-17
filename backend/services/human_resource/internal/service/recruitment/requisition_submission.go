package service

import (
	"context"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

// SubmitRequisition mengirim draf pembuatnya ke tahap persetujuan secara atomik.
func (s *Service) SubmitRequisition(ctx context.Context, requisitionID string) (*entity.Requisition, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.requisition.Command(ctx, requisitionID, "submit", func(_ context.Context, requisition *entity.Requisition, actor operation.Actor, now time.Time) (*entity.Requisition, error) {
		if actor.Kind != "human" {
			return nil, hrerrors.ErrForbidden
		}
		if requisition.Status != "draft" || requisition.CreatedByPrincipalID == nil ||
			*requisition.CreatedByPrincipalID != actor.PrincipalID || requisition.SubmittedAt != nil ||
			requisition.SubmittedByPrincipalID != nil || requisition.ApprovedAt != nil ||
			requisition.ApprovedByPrincipalID != nil || requisition.OpenedAt != nil || requisition.ClosedAt != nil {
			return nil, hrerrors.ErrWorkflowRequired
		}
		requisition.Status = "submitted"
		requisition.SubmittedAt = &now
		requisition.SubmittedByPrincipalID = &actor.PrincipalID
		return requisition, nil
	})
}
