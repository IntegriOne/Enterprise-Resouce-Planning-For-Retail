package service

import (
	"context"
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
	WorkLocationReferenceID     *string
	WorkforcePlanLineID         *string
	Headcount                   int32
	TargetStartDate             *time.Time
	EmploymentCategory          entity.WorkerCategory
	CreatedByPrincipalID        string
	SubmittedByPrincipalID      string
	ApproverPrincipalID         string
	ExpectedBudgetReferenceID   string
}

// RequisitionMandateVerifier memeriksa grant peninjau melalui IAM.
type RequisitionMandateVerifier interface {
	VerifyRequisitionMandate(context.Context, string, string) error
}

// RequisitionApprovalVerifier membuktikan headcount dan anggaran.
// Nilai kembali adalah referensi otorisasi anggaran milik Finance yang stabil.
type RequisitionApprovalVerifier interface {
	VerifyRequisitionApproval(context.Context, RequisitionApprovalCheck) (string, error)
}

// ApproveRequisition memutuskan requisition submitted setelah bukti owner tersedia.
func (s *Service) ApproveRequisition(ctx context.Context, requisitionID string) (*entity.Requisition, error) {
	if s == nil {
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
		if operation.RequireDependency(s.mandateVerifier) != nil {
			return nil, hrerrors.ErrUnavailable
		}
		if err := s.mandateVerifier.VerifyRequisitionMandate(txCtx, actor.PrincipalID, requisition.OrganizationUnitReferenceID); err != nil {
			return nil, operation.OwnerVerificationError(txCtx, err)
		}
		if operation.RequireDependency(s.approvalVerifier) != nil {
			return nil, hrerrors.ErrUnavailable
		}
		budgetID, err := s.approvalVerifier.VerifyRequisitionApproval(txCtx, requisitionApprovalCheck(requisition, actor.PrincipalID))
		if err != nil {
			return nil, operation.OwnerVerificationError(txCtx, err)
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

func requisitionApprovalCheck(requisition *entity.Requisition, approverID string) RequisitionApprovalCheck {
	return RequisitionApprovalCheck{
		RequisitionID: requisition.ID, RequisitionType: requisition.RequisitionType,
		OrganizationUnitReferenceID: requisition.OrganizationUnitReferenceID,
		HiringManagerPrincipalID:    requisition.HiringManagerPrincipalID,
		JobID:                       requisition.JobID, PositionID: requisition.PositionID,
		WorkLocationReferenceID: requisition.WorkLocationReferenceID,
		WorkforcePlanLineID:     requisition.WorkforcePlanLineID, Headcount: requisition.Headcount,
		TargetStartDate: requisition.TargetStartDate, EmploymentCategory: requisition.EmploymentCategory,
		CreatedByPrincipalID:   *requisition.CreatedByPrincipalID,
		SubmittedByPrincipalID: *requisition.SubmittedByPrincipalID,
		ApproverPrincipalID:    approverID,
	}
}
