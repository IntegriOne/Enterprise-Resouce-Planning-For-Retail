package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"strings"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

type PayrollPaymentCheck struct {
	RunID, JournalID, ResultID, WorkerID, PaymentID  string
	LegalEntityID, CurrencyCode, SettlementDirection string
	NetPayAmount                                     entity.Decimal
	RunOutputDigest, CalculationDigest               []byte
}

type PayrollPaymentVerifier interface {
	VerifyPayrollPayment(context.Context, PayrollPaymentCheck) error
}

// ConfirmPayrollPayment mencatat hasil pembayaran individual yang telah diverifikasi oleh Finance.
func (s *Service) ConfirmPayrollPayment(ctx context.Context, runID, resultID, paymentID string) error {
	if s == nil || operation.RequireDependency(s.dependencies.PayrollPayment) != nil {
		return hrerrors.ErrUnavailable
	}
	if entity.ValidateUUID("run_id", runID) != nil || entity.ValidateUUID("result_id", resultID) != nil ||
		entity.ValidateUUID("payment_id", paymentID) != nil {
		return hrerrors.ErrInvalidInput
	}
	deps := s.dependencies
	return deps.Runtime.Transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		run, err := deps.PayrollRun.FindByIDForUpdate(txCtx, runID)
		if err != nil {
			return err
		}
		if run == nil || !strings.EqualFold(run.ID, runID) ||
			(run.Status != entity.PayrollRunStatusPosted && run.Status != entity.PayrollRunStatusPaid) ||
			run.FinanceJournalReferenceID == nil || len(run.OutputDigest) != sha256.Size {
			return hrerrors.ErrWorkflowRequired
		}
		result, err := deps.PayrollResult.FindByIDForUpdate(txCtx, resultID)
		if err != nil {
			return err
		}
		if result == nil || !strings.EqualFold(result.ID, resultID) || !strings.EqualFold(result.PayrollRunID, runID) ||
			!strings.EqualFold(result.PayrollGroupID, run.PayrollGroupID) || result.CurrencyCode != run.CurrencyCode ||
			len(result.CalculationDigest) != sha256.Size {
			return hrerrors.ErrReferenceConflict
		}
		if result.ResultStatus == "paid" && result.PaymentReferenceID != nil &&
			strings.EqualFold(*result.PaymentReferenceID, paymentID) {
			return nil
		}
		if result.PaymentReferenceID != nil {
			return hrerrors.ErrReferenceConflict
		}
		switch result.ResultStatus {
		case "calculated", "approved", "posted":
		default:
			return hrerrors.ErrWorkflowRequired
		}
		if err := deps.PayrollPayment.VerifyPayrollPayment(txCtx, PayrollPaymentCheck{
			RunID: run.ID, JournalID: *run.FinanceJournalReferenceID, ResultID: result.ID, WorkerID: result.WorkerID,
			PaymentID: paymentID, LegalEntityID: run.EmployingLegalEntityReferenceID,
			CurrencyCode: run.CurrencyCode, SettlementDirection: run.SettlementDirection, NetPayAmount: result.NetPayAmount,
			RunOutputDigest: bytes.Clone(run.OutputDigest), CalculationDigest: bytes.Clone(result.CalculationDigest),
		}); err != nil {
			return operation.OwnerVerificationError(txCtx, err)
		}
		if err := txCtx.Err(); err != nil {
			return err
		}
		result.ResultStatus = "paid"
		result.PaymentReferenceID = &paymentID
		if err := result.Validate(); err != nil {
			return err
		}
		return deps.PayrollResult.Update(txCtx, result)
	})
}
