package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

// RequisitionApprovalCheck membawa snapshot yang harus dibuktikan oleh owner terkait.
type RequisitionApprovalCheck struct {
	RequisitionID               string
	RequisitionType             string
	OrganizationUnitReferenceID string
	HiringManagerPrincipalID    string
	JobID                       string
	PositionID                  *string
	WorkforcePlanLineID         *string
	Headcount                   int32
	TargetStartDate             *time.Time
	EmploymentCategory          entity.WorkerCategory
	CreatedByPrincipalID        string
	SubmittedByPrincipalID      string
	ApproverPrincipalID         string
}

// RequisitionApprovalVerifier membuktikan mandat unit, headcount dan anggaran.
// Nilai kembali adalah referensi otorisasi anggaran milik Finance yang stabil.
type RequisitionApprovalVerifier interface {
	VerifyRequisitionApproval(context.Context, RequisitionApprovalCheck) (string, error)
}

// ApproveRequisition memutuskan requisition submitted setelah bukti owner tersedia.
func (s *Service) ApproveRequisition(ctx context.Context, requisitionID string) (*entity.Requisition, error) {
	if s == nil || operation.RequireDependency(s.approvalVerifier) != nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.requisition.Command(ctx, requisitionID, "approve", func(txCtx context.Context, requisition *entity.Requisition, actor operation.Actor, now time.Time) (*entity.Requisition, error) {
		if actor.Kind != "human" {
			return nil, hrerrors.ErrForbidden
		}
		if requisition.Status != "submitted" || requisition.CreatedByPrincipalID == nil ||
			requisition.SubmittedAt == nil || requisition.SubmittedByPrincipalID == nil ||
			requisition.ApprovedAt != nil || requisition.ApprovedByPrincipalID != nil ||
			requisition.OpenedAt != nil || requisition.ClosedAt != nil || requisition.BudgetReferenceID != nil {
			return nil, hrerrors.ErrWorkflowRequired
		}
		if strings.EqualFold(actor.PrincipalID, *requisition.CreatedByPrincipalID) ||
			strings.EqualFold(actor.PrincipalID, *requisition.SubmittedByPrincipalID) {
			return nil, hrerrors.ErrForbidden
		}
		budgetID, err := s.approvalVerifier.VerifyRequisitionApproval(txCtx, RequisitionApprovalCheck{
			RequisitionID: requisition.ID, RequisitionType: requisition.RequisitionType,
			OrganizationUnitReferenceID: requisition.OrganizationUnitReferenceID,
			HiringManagerPrincipalID:    requisition.HiringManagerPrincipalID,
			JobID:                       requisition.JobID, PositionID: requisition.PositionID,
			WorkforcePlanLineID: requisition.WorkforcePlanLineID, Headcount: requisition.Headcount,
			TargetStartDate: requisition.TargetStartDate, EmploymentCategory: requisition.EmploymentCategory,
			CreatedByPrincipalID:   *requisition.CreatedByPrincipalID,
			SubmittedByPrincipalID: *requisition.SubmittedByPrincipalID,
			ApproverPrincipalID:    actor.PrincipalID,
		})
		if err != nil {
			if txCtx.Err() != nil {
				return nil, txCtx.Err()
			}
			if errors.Is(err, hrerrors.ErrReferenceConflict) {
				return nil, hrerrors.ErrReferenceConflict
			}
			return nil, hrerrors.ErrUnavailable
		}
		if err := entity.ValidateUUID("budget_reference_id", budgetID); err != nil {
			return nil, hrerrors.ErrReferenceConflict
		}
		requisition.Status = "approved"
		requisition.BudgetReferenceID = &budgetID
		requisition.ApprovedAt = &now
		requisition.ApprovedByPrincipalID = &actor.PrincipalID
		return requisition, nil
	})
}
