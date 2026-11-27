package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	payrollpolicy "github.com/integrione/erp-retail/backend/services/human_resource/domain/service/payroll"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

type PayslipPublicationCheck struct {
	PublicationID, RunID, JournalID, LegalEntityID string
	InputSnapshotID                                string
	InputDigest, OutputDigest                      []byte
}

type PayslipDocumentInput struct {
	PublicationID, IdempotencyKey string
	RestoreVersionID              string
	WorkerNumber                  string
	PublishedAt                   time.Time
	Result                        entity.PayrollResult
	Period                        entity.PayrollPeriod
	Lines                         []entity.PayrollResultLine
}

// PayslipPublicationProcessor memverifikasi penerbitan Finance dan menyimpan PDF privat secara idempoten di DRM.
type PayslipPublicationProcessor interface {
	VerifyPayrollPublication(context.Context, PayslipPublicationCheck) (time.Time, error)
	StorePayslipPDF(context.Context, PayslipDocumentInput) (string, error)
}

// PublishPayslipsForRun menangani satu penerbitan Finance dengan transaksi per hasil.
func (s *Service) PublishPayslipsForRun(ctx context.Context, publicationID, runID string) error {
	if s == nil || operation.RequireDependency(s.dependencies.PayslipPublication) != nil {
		return hrerrors.ErrUnavailable
	}
	if err := entity.ValidateUUID("publication_id", publicationID); err != nil {
		return err
	}
	if err := entity.ValidateUUID("payroll_run_id", runID); err != nil {
		return err
	}
	repo := s.dependencies.PayrollResult
	filters := []domainrepo.Filter{{Column: "payroll_run_id", Values: []any{runID}}}
	count, err := repo.CountWhere(ctx, filters)
	if err != nil {
		return err
	}
	results, err := operation.CollectComplete(ctx, count, func(limit, offset int) (operation.Page[entity.PayrollResult], error) {
		items, err := repo.SelectWhere(ctx, filters, limit, offset)
		return operation.Page[entity.PayrollResult]{Items: items}, err
	})
	if err != nil {
		return err
	}
	if len(results) == 0 {
		return hrerrors.ErrWorkflowRequired
	}
	var failures []error
	for _, result := range results {
		if err := ctx.Err(); err != nil {
			return errors.Join(append(failures, err)...)
		}
		err := hrerrors.ErrReferenceConflict
		if strings.EqualFold(result.PayrollRunID, runID) {
			_, _, err = s.publishPayslipForResult(ctx, publicationID, result.ID, runID, "", nil)
		}
		if err != nil {
			failures = append(failures, fmt.Errorf("publish payslip for payroll result %s: %w", result.ID, err))
		}
	}
	return errors.Join(failures...)
}

// PublishPayslipForResult memproses satu hasil dari konfirmasi penerbitan privat Finance.
func (s *Service) PublishPayslipForResult(ctx context.Context, publicationID, resultID string) (string, error) {
	documentID, _, err := s.publishPayslipForResult(ctx, publicationID, resultID, "", "", nil)
	return documentID, err
}

func (s *Service) publishPayslipForResult(ctx context.Context, publicationID, resultID, expectedRunID, restoreVersionID string, reauthorize func(context.Context) error) (string, string, error) {
	if s == nil || operation.RequireDependency(s.dependencies.PayslipPublication) != nil {
		return "", "", hrerrors.ErrUnavailable
	}
	if err := entity.ValidateUUID("publication_id", publicationID); err != nil {
		return "", "", err
	}
	if err := entity.ValidateUUID("payroll_result_id", resultID); err != nil {
		return "", "", err
	}
	deps := s.dependencies
	var documentID, outcome string
	err := deps.Runtime.Transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		record, err := deps.PayrollResult.SelectByID(txCtx, resultID)
		if err != nil {
			return err
		}
		if record == nil {
			return hrerrors.ErrNotFound
		}
		runID := record.PayrollRunID
		if expectedRunID != "" {
			if !strings.EqualFold(runID, expectedRunID) {
				return hrerrors.ErrReferenceConflict
			}
			runID = expectedRunID
		}
		run, err := deps.PayrollRun.FindByIDForUpdate(txCtx, runID)
		if err != nil {
			return err
		}
		if run == nil || (run.Status != entity.PayrollRunStatusPosted && run.Status != entity.PayrollRunStatusPaid) ||
			run.FinanceJournalReferenceID == nil || len(run.OutputDigest) != sha256.Size ||
			len(run.InputDigest) != sha256.Size || entity.ValidateUUID("input_snapshot_reference_id", run.InputSnapshotReferenceID) != nil {
			return hrerrors.ErrWorkflowRequired
		}
		result, err := deps.PayrollResult.FindByIDForUpdate(txCtx, resultID)
		if err != nil {
			return err
		}
		if result == nil || !strings.EqualFold(run.ID, runID) || !strings.EqualFold(result.PayrollRunID, run.ID) ||
			!strings.EqualFold(result.PayrollGroupID, run.PayrollGroupID) || result.CurrencyCode != run.CurrencyCode ||
			len(result.CalculationDigest) != sha256.Size {
			return hrerrors.ErrReferenceConflict
		}
		switch result.ResultStatus {
		case "calculated", "approved", "posted":
		case "paid":
			if result.PaymentReferenceID == nil || entity.ValidateUUID("payment_reference_id", *result.PaymentReferenceID) != nil {
				return hrerrors.ErrReferenceConflict
			}
		default:
			return hrerrors.ErrReferenceConflict
		}
		publishedAt, err := deps.PayslipPublication.VerifyPayrollPublication(txCtx, PayslipPublicationCheck{
			PublicationID: publicationID, RunID: run.ID, JournalID: *run.FinanceJournalReferenceID,
			LegalEntityID: run.EmployingLegalEntityReferenceID, InputSnapshotID: run.InputSnapshotReferenceID,
			InputDigest: bytes.Clone(run.InputDigest), OutputDigest: bytes.Clone(run.OutputDigest),
		})
		if err != nil {
			return operation.OwnerVerificationError(txCtx, err)
		}
		now := deps.Runtime.Clock.Now()
		if now.IsZero() || publishedAt.IsZero() || publishedAt.After(now) ||
			run.ApprovedAt == nil || publishedAt.Before(*run.ApprovedAt) ||
			(run.PostingRequestedAt != nil && publishedAt.Before(*run.PostingRequestedAt)) {
			return hrerrors.ErrReferenceConflict
		}
		if result.PayslipDocumentReferenceID != nil {
			if entity.ValidateUUID("document_version_id", *result.PayslipDocumentReferenceID) != nil ||
				result.PayslipPublicationReferenceID == nil || !strings.EqualFold(*result.PayslipPublicationReferenceID, publicationID) {
				return hrerrors.ErrReferenceConflict
			}
			documentID = *result.PayslipDocumentReferenceID
			if restoreVersionID == "" {
				outcome = "reused"
				return nil
			}
			if !strings.EqualFold(restoreVersionID, documentID) {
				return hrerrors.ErrReferenceConflict
			}
		} else if restoreVersionID != "" {
			return hrerrors.ErrReferenceConflict
		}
		if operation.RequireDependency(deps.Workers) != nil {
			return hrerrors.ErrUnavailable
		}
		worker, err := deps.Workers.SelectByID(txCtx, result.WorkerID)
		if err != nil {
			return err
		}
		if worker == nil || !strings.EqualFold(worker.ID, result.WorkerID) || strings.TrimSpace(worker.WorkerNumber) == "" {
			return hrerrors.ErrReferenceConflict
		}
		period, err := deps.PayrollPeriod.SelectByID(txCtx, run.PayrollPeriodID)
		if err != nil {
			return err
		}
		if period == nil || !strings.EqualFold(period.ID, run.PayrollPeriodID) ||
			!strings.EqualFold(period.PayrollGroupID, run.PayrollGroupID) ||
			!strings.EqualFold(period.EmployingLegalEntityReferenceID, run.EmployingLegalEntityReferenceID) ||
			period.CurrencyCode != run.CurrencyCode || period.JurisdictionCode != run.JurisdictionCode {
			return hrerrors.ErrReferenceConflict
		}
		filters := []domainrepo.Filter{{Column: "payroll_result_id", Values: []any{result.ID}}}
		count, err := deps.PayrollResultLine.CountWhere(txCtx, filters)
		if err != nil {
			return err
		}
		lines, err := operation.CollectComplete(txCtx, count, func(limit, offset int) (operation.Page[entity.PayrollResultLine], error) {
			items, err := deps.PayrollResultLine.SelectWhere(txCtx, filters, limit, offset)
			return operation.Page[entity.PayrollResultLine]{Items: items}, err
		})
		if err != nil {
			return err
		}
		if err := payrollpolicy.ValidateResultTotals(result, lines); err != nil {
			return err
		}
		if reauthorize != nil {
			if err := reauthorize(txCtx); err != nil {
				return err
			}
		}
		documentID, err = deps.PayslipPublication.StorePayslipPDF(txCtx, PayslipDocumentInput{
			PublicationID:    publicationID,
			RestoreVersionID: restoreVersionID,
			WorkerNumber:     worker.WorkerNumber,
			PublishedAt:      publishedAt,
			IdempotencyKey:   "payslip:" + strings.ToLower(publicationID) + ":" + strings.ToLower(result.ID),
			Result:           *result, Period: *period, Lines: lines,
		})
		if err != nil {
			return operation.OwnerVerificationError(txCtx, err)
		}
		if reauthorize != nil {
			if err := reauthorize(txCtx); err != nil {
				return err
			}
		}
		if entity.ValidateUUID("document_version_id", documentID) != nil {
			return hrerrors.ErrReferenceConflict
		}
		if restoreVersionID != "" {
			if !strings.EqualFold(restoreVersionID, documentID) {
				return hrerrors.ErrReferenceConflict
			}
			outcome = "restored"
			return nil
		}
		outcome = "created"
		result.PayslipDocumentReferenceID = &documentID
		result.PayslipPublicationReferenceID = &publicationID
		return deps.PayrollResult.Update(txCtx, result)
	})
	if err != nil {
		return "", "", err
	}
	return documentID, outcome, nil
}
