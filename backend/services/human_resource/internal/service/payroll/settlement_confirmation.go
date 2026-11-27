package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"strings"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

type PayrollSettlementResult struct {
	ResultID, PaymentID string
}

type PayrollSettlementCheck struct {
	RunID, JournalID, BatchID, SettlementID          string
	LegalEntityID, CurrencyCode, SettlementDirection string
	OutputDigest                                     []byte
	Results                                          []PayrollSettlementResult
}

type PayrollSettlementVerifier interface {
	VerifyPayrollSettlement(context.Context, PayrollSettlementCheck) (time.Time, error)
}

// ConfirmPayrollSettlement memproyeksikan bukti settlement akhir Finance untuk seluruh run.
func (s *Service) ConfirmPayrollSettlement(ctx context.Context, runID, batchID, settlementID string) error {
	if s == nil || operation.RequireDependency(s.dependencies.PayrollSettlement) != nil {
		return hrerrors.ErrUnavailable
	}
	if entity.ValidateUUID("run_id", runID) != nil || entity.ValidateUUID("batch_id", batchID) != nil ||
		entity.ValidateUUID("settlement_id", settlementID) != nil {
		return hrerrors.ErrInvalidInput
	}
	deps := s.dependencies
	return deps.Runtime.Transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		run, err := deps.PayrollRun.FindByIDForUpdate(txCtx, runID)
		if err != nil {
			return err
		}
		if run == nil || !strings.EqualFold(run.ID, runID) {
			return hrerrors.ErrReferenceConflict
		}
		if run.Status == entity.PayrollRunStatusPaid {
			if run.PaymentBatchReferenceID == nil || run.FinanceSettlementEventReferenceID == nil || run.FinanceSettledAt == nil ||
				!strings.EqualFold(*run.PaymentBatchReferenceID, batchID) ||
				!strings.EqualFold(*run.FinanceSettlementEventReferenceID, settlementID) {
				return hrerrors.ErrReferenceConflict
			}
			return nil
		}
		if run.Status != entity.PayrollRunStatusPosted || run.FinanceJournalReferenceID == nil ||
			run.PostingRequestedAt == nil || len(run.OutputDigest) != sha256.Size ||
			run.PaymentBatchReferenceID != nil || run.FinanceSettlementEventReferenceID != nil || run.FinanceSettledAt != nil {
			return hrerrors.ErrWorkflowRequired
		}
		results, err := s.settlementResults(txCtx, run)
		if err != nil {
			return err
		}
		settledAt, err := deps.PayrollSettlement.VerifyPayrollSettlement(txCtx, PayrollSettlementCheck{
			RunID: run.ID, JournalID: *run.FinanceJournalReferenceID, BatchID: batchID, SettlementID: settlementID,
			LegalEntityID: run.EmployingLegalEntityReferenceID, CurrencyCode: run.CurrencyCode,
			SettlementDirection: run.SettlementDirection, OutputDigest: bytes.Clone(run.OutputDigest), Results: results,
		})
		if err != nil {
			return operation.OwnerVerificationError(txCtx, err)
		}
		if settledAt.IsZero() || settledAt.Before(*run.PostingRequestedAt) || settledAt.After(deps.Runtime.Clock.Now()) {
			return hrerrors.ErrReferenceConflict
		}
		if err := txCtx.Err(); err != nil {
			return err
		}
		run.Status = entity.PayrollRunStatusPaid
		run.PaymentBatchReferenceID, run.FinanceSettlementEventReferenceID, run.FinanceSettledAt = &batchID, &settlementID, &settledAt
		if err := run.Validate(); err != nil {
			return err
		}
		return deps.PayrollRun.Update(txCtx, run)
	})
}

func (s *Service) settlementResults(ctx context.Context, run *entity.PayrollRun) ([]PayrollSettlementResult, error) {
	repo := s.dependencies.PayrollResult
	filters := []domainrepo.Filter{{Column: "payroll_run_id", Values: []any{run.ID}}}
	count, err := repo.CountWhere(ctx, filters)
	if err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, hrerrors.ErrWorkflowRequired
	}
	rows, err := operation.CollectComplete(ctx, count, func(limit, offset int) (operation.Page[entity.PayrollResult], error) {
		items, err := repo.SelectWhere(ctx, filters, limit, offset)
		return operation.Page[entity.PayrollResult]{Items: items}, err
	})
	if err != nil {
		return nil, err
	}
	results := make([]PayrollSettlementResult, 0, len(rows))
	for _, row := range rows {
		if !strings.EqualFold(row.PayrollRunID, run.ID) || !strings.EqualFold(row.PayrollGroupID, run.PayrollGroupID) ||
			row.CurrencyCode != run.CurrencyCode {
			return nil, hrerrors.ErrReferenceConflict
		}
		if row.ResultStatus != "paid" || row.PaymentReferenceID == nil {
			return nil, hrerrors.ErrWorkflowRequired
		}
		if entity.ValidateUUID("payment_reference_id", *row.PaymentReferenceID) != nil {
			return nil, hrerrors.ErrReferenceConflict
		}
		results = append(results, PayrollSettlementResult{ResultID: row.ID, PaymentID: *row.PaymentReferenceID})
	}
	return results, nil
}
