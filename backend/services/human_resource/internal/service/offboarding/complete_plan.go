package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

type ClearanceRequest struct {
	PlanID, WorkerID, FinalPayrollRunID, FinanceClearanceReferenceID string
	ActorPrincipalID, WorkerPrincipalID                              string
	FinalPayrollResultID, EmployingLegalEntityID, CurrencyCode       string
	PaymentReferenceID                                               string
	RunOutputDigest, ResultCalculationDigest                         []byte
	TerminationEffectiveDate                                         time.Time
}

// ClearanceVerifier memeriksa keputusan clearance dari Finance dan IAM.
type ClearanceVerifier interface {
	VerifyOffboardingClearance(context.Context, ClearanceRequest) error
}

// CompleteOffboardingPlan menutup rencana setelah tugas dan bukti clearance terpenuhi.
func (s *Service) CompleteOffboardingPlan(ctx context.Context, planID, payrollRunID, clearanceID string) (*entity.OffboardingPlan, error) {
	if s == nil || operation.RequireDependency(s.clearance) != nil || operation.RequireDependency(s.payrollResults) != nil ||
		operation.RequireDependency(s.payrollRuns) != nil || operation.RequireDependency(s.workers) != nil {
		return nil, hrerrors.ErrUnavailable
	}
	if err := entity.ValidateUUID("final_payroll_run_id", payrollRunID); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("finance_clearance_reference_id", clearanceID); err != nil {
		return nil, err
	}
	return s.offboardingPlan.Command(ctx, planID, "complete", func(txCtx context.Context, plan *entity.OffboardingPlan, actor operation.Actor, now time.Time) (*entity.OffboardingPlan, error) {
		if actor.Kind != "human" {
			return nil, hrerrors.ErrForbidden
		}
		if plan.Status != "in_progress" || plan.CompletedAt != nil {
			return nil, hrerrors.ErrWorkflowRequired
		}
		if entity.CompareDates(now.UTC(), plan.TerminationEffectiveDate) < 0 {
			return nil, hrerrors.ErrReferenceConflict
		}
		run, err := s.payrollRuns.FindByIDForUpdate(txCtx, payrollRunID)
		if errors.Is(err, hrerrors.ErrNotFound) {
			return nil, hrerrors.ErrReferenceConflict
		}
		if err != nil {
			return nil, err
		}
		if err := validateFinalPayrollRun(run, now); err != nil {
			return nil, err
		}
		results, err := s.payrollResults.SelectWhere(txCtx, []domainrepo.Filter{
			{Column: "payroll_run_id", Values: []any{payrollRunID}},
			{Column: "worker_id", Values: []any{plan.WorkerID}},
		}, 2, 0)
		if err != nil {
			return nil, err
		}
		if err := validateFinalPayrollResults(run.Status, results); err != nil {
			return nil, err
		}
		result := results[0]
		if len(run.OutputDigest) != sha256.Size || len(result.CalculationDigest) != sha256.Size {
			return nil, hrerrors.ErrReferenceConflict
		}
		before := *plan
		plan.Status = "completed"
		plan.CompletedAt = &now
		plan.FinalPayrollRunID = &payrollRunID
		plan.FinanceClearanceReferenceID = &clearanceID
		if err := s.offboardingPlan.CheckIntegrity(txCtx, &before, plan); err != nil {
			return nil, err
		}
		worker, err := s.workers.FindByIDForUpdate(txCtx, plan.WorkerID)
		if err != nil {
			return nil, err
		}
		if worker == nil {
			return nil, hrerrors.ErrReferenceConflict
		}
		var workerPrincipalID string
		if worker.IAMPrincipalReferenceID != nil {
			workerPrincipalID = *worker.IAMPrincipalReferenceID
			if err := entity.ValidateUUID("worker_principal_id", workerPrincipalID); err != nil {
				return nil, hrerrors.ErrReferenceConflict
			}
		}
		var paymentReference string
		if result.PaymentReferenceID != nil {
			paymentReference = *result.PaymentReferenceID
		}
		if err := s.clearance.VerifyOffboardingClearance(txCtx, ClearanceRequest{
			PlanID: plan.ID, WorkerID: plan.WorkerID, FinalPayrollRunID: payrollRunID,
			FinanceClearanceReferenceID: clearanceID, ActorPrincipalID: actor.PrincipalID, WorkerPrincipalID: workerPrincipalID,
			FinalPayrollResultID: result.ID, EmployingLegalEntityID: run.EmployingLegalEntityReferenceID,
			CurrencyCode: run.CurrencyCode, PaymentReferenceID: paymentReference, RunOutputDigest: bytes.Clone(run.OutputDigest),
			ResultCalculationDigest:  bytes.Clone(result.CalculationDigest),
			TerminationEffectiveDate: plan.TerminationEffectiveDate,
		}); err != nil {
			return nil, operation.OwnerVerificationError(txCtx, err)
		}
		return plan, nil
	})
}

func validateFinalPayrollResults(runStatus entity.PayrollRunStatus, results []entity.PayrollResult) error {
	if len(results) != 1 {
		return hrerrors.ErrReferenceConflict
	}
	// Persetujuan run dapat mempertahankan status calculated pada hasil individual.
	switch results[0].ResultStatus {
	case "calculated", "approved":
		return nil
	case "posted":
		if runStatus == entity.PayrollRunStatusPosted || runStatus == entity.PayrollRunStatusPaid {
			return nil
		}
	case "paid":
		if (runStatus == entity.PayrollRunStatusPosted || runStatus == entity.PayrollRunStatusPaid) &&
			results[0].PaymentReferenceID != nil && entity.ValidateUUID("payment_reference_id", *results[0].PaymentReferenceID) == nil {
			return nil
		}
	}
	return hrerrors.ErrReferenceConflict
}

// validateFinalPayrollRun memastikan HRM sudah menyetujui run terminasi.
func validateFinalPayrollRun(run *entity.PayrollRun, now time.Time) error {
	if run == nil || run.RunType != "termination" || run.ApprovedAt == nil || run.ApprovedByPrincipalID == nil || run.ApprovedAt.After(now) {
		return hrerrors.ErrReferenceConflict
	}
	switch run.Status {
	case entity.PayrollRunStatusApproved, entity.PayrollRunStatusPosted, entity.PayrollRunStatusPaid:
		return nil
	default:
		return hrerrors.ErrReferenceConflict
	}
}
