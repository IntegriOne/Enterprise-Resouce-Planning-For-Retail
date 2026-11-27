package service

import (
	"context"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

// OpenRequisition membuka requisition yang sudah disetujui setelah bukti owner diperiksa ulang.
func (s *Service) OpenRequisition(ctx context.Context, requisitionID string) (*entity.Requisition, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.requisition.Command(ctx, requisitionID, "open", func(txCtx context.Context, requisition *entity.Requisition, actor operation.Actor, now time.Time) (*entity.Requisition, error) {
		if actor.Kind != "human" {
			return nil, hrerrors.ErrForbidden
		}
		if requisition.Status != "approved" || requisition.CreatedByPrincipalID == nil ||
			requisition.SubmittedAt == nil || requisition.SubmittedByPrincipalID == nil ||
			requisition.ApprovedAt == nil || requisition.ApprovedByPrincipalID == nil ||
			requisition.BudgetReferenceID == nil || requisition.OpenedAt != nil || requisition.ClosedAt != nil {
			return nil, hrerrors.ErrWorkflowRequired
		}
		if now.Before(*requisition.ApprovedAt) {
			return nil, hrerrors.ErrWorkflowRequired
		}
		if operation.RequireDependency(s.mandateVerifier) != nil {
			return nil, hrerrors.ErrUnavailable
		}
		if err := s.mandateVerifier.VerifyRequisitionMandate(txCtx, *requisition.ApprovedByPrincipalID, requisition.OrganizationUnitReferenceID); err != nil {
			return nil, operation.OwnerVerificationError(txCtx, err)
		}
		if operation.RequireDependency(s.approvalVerifier) != nil {
			return nil, hrerrors.ErrUnavailable
		}
		check := requisitionApprovalCheck(requisition, *requisition.ApprovedByPrincipalID)
		check.ExpectedBudgetReferenceID = *requisition.BudgetReferenceID
		budgetID, err := s.approvalVerifier.VerifyRequisitionApproval(txCtx, check)
		if err != nil {
			return nil, operation.OwnerVerificationError(txCtx, err)
		}
		if budgetID != *requisition.BudgetReferenceID {
			return nil, hrerrors.ErrReferenceConflict
		}
		requisition.Status = "open"
		requisition.OpenedAt = &now
		return requisition, nil
	})
}
