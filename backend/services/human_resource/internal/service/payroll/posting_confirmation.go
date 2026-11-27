package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"strings"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

type PayrollPostingCheck struct {
	RunID, JournalID, LegalEntityID, PeriodID, GroupID, InputSnapshotID string
	CurrencyCode, SettlementDirection                                   string
	InputDigest, OutputDigest                                           []byte
	RequestedAt                                                         time.Time
}

type PayrollPostingVerifier interface {
	VerifyPayrollPosting(context.Context, PayrollPostingCheck) (time.Time, error)
}

// ConfirmPayrollPosting memproyeksikan jurnal Finance yang terikat pada snapshot payroll HRM.
func (s *Service) ConfirmPayrollPosting(ctx context.Context, runID, journalID string) error {
	if s == nil || operation.RequireDependency(s.dependencies.PayrollPosting) != nil {
		return hrerrors.ErrUnavailable
	}
	if entity.ValidateUUID("payroll_run_id", runID) != nil || entity.ValidateUUID("journal_id", journalID) != nil {
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
		if run.Status == entity.PayrollRunStatusPosted || run.Status == entity.PayrollRunStatusPaid {
			if run.FinanceJournalReferenceID == nil || !strings.EqualFold(*run.FinanceJournalReferenceID, journalID) {
				return hrerrors.ErrReferenceConflict
			}
			return nil
		}
		if run.Status != entity.PayrollRunStatusApproved || run.PostingRequestedAt == nil ||
			run.PostingRequestedByPrincipalID == nil || run.ApprovedAt == nil ||
			run.PostingRequestedAt.Before(*run.ApprovedAt) || run.FinanceJournalReferenceID != nil ||
			len(run.InputDigest) != sha256.Size || len(run.OutputDigest) != sha256.Size {
			return hrerrors.ErrWorkflowRequired
		}
		postedAt, err := deps.PayrollPosting.VerifyPayrollPosting(txCtx, PayrollPostingCheck{
			RunID: run.ID, JournalID: journalID, LegalEntityID: run.EmployingLegalEntityReferenceID,
			PeriodID: run.PayrollPeriodID, GroupID: run.PayrollGroupID, InputSnapshotID: run.InputSnapshotReferenceID,
			CurrencyCode: run.CurrencyCode, SettlementDirection: run.SettlementDirection,
			InputDigest: bytes.Clone(run.InputDigest), OutputDigest: bytes.Clone(run.OutputDigest),
			RequestedAt: *run.PostingRequestedAt,
		})
		if err != nil {
			return operation.OwnerVerificationError(txCtx, err)
		}
		if postedAt.IsZero() || postedAt.Before(*run.PostingRequestedAt) || postedAt.After(deps.Runtime.Clock.Now()) {
			return hrerrors.ErrReferenceConflict
		}
		if err := txCtx.Err(); err != nil {
			return err
		}
		run.Status = entity.PayrollRunStatusPosted
		run.FinanceJournalReferenceID = &journalID
		if err := run.Validate(); err != nil {
			return err
		}
		return deps.PayrollRun.Update(txCtx, run)
	})
}
