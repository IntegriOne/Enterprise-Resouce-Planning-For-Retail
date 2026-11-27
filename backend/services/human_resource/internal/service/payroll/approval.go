package service

import (
	"context"
	"crypto/sha256"
	"slices"
	"strings"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	payrollpolicy "github.com/integrione/erp-retail/backend/services/human_resource/domain/service/payroll"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

// ApprovePayrollRun merekonsiliasi seluruh hasil dan menyimpan approval bersama outbox.
// Perhitungan tarif, jurnal, pembayaran dan mandat approval tetap mengikuti kontrak owner.
func (s *Service) ApprovePayrollRun(ctx context.Context, runID string) (*entity.PayrollRun, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.payrollRun.Command(ctx, runID, "approve", func(txCtx context.Context, run *entity.PayrollRun, actor operation.Actor, now time.Time) (*entity.PayrollRun, error) {
		if run.Status != entity.PayrollRunStatusReview {
			return nil, hrerrors.ErrWorkflowRequired
		}
		if actor.Kind != "human" || strings.EqualFold(run.StartedByPrincipalID, actor.PrincipalID) {
			return nil, hrerrors.ErrForbidden
		}
		if run.ApprovedAt != nil || run.ApprovedByPrincipalID != nil {
			return nil, hrerrors.ErrConflict
		}
		if err := s.validateRunResults(txCtx, run); err != nil {
			return nil, err
		}
		run.Status = entity.PayrollRunStatusApproved
		run.ApprovedByPrincipalID = &actor.PrincipalID
		run.ApprovedAt = &now
		return run, nil
	})
}

// validateRunResults merekonsiliasi baris dari snapshot yang dapat dilihat secara lengkap.
func (s *Service) validateRunResults(txCtx context.Context, run *entity.PayrollRun) error {
	if len(run.InputDigest) != sha256.Size || len(run.OutputDigest) != sha256.Size {
		return hrerrors.ErrReferenceConflict
	}
	group, err := s.dependencies.PayrollGroup.SelectByID(txCtx, run.PayrollGroupID)
	if err != nil {
		return err
	}
	if group == nil || !strings.EqualFold(group.ID, run.PayrollGroupID) ||
		group.CurrencyCode != run.CurrencyCode {
		return hrerrors.ErrReferenceConflict
	}
	resultFilter := []domainrepo.Filter{{Column: "payroll_run_id", Values: []any{run.ID}}}
	expected, err := s.dependencies.PayrollResult.CountWhere(txCtx, resultFilter)
	if err != nil {
		return err
	}
	results, err := operation.CollectComplete(txCtx, expected, func(limit, offset int) (operation.Page[entity.PayrollResult], error) {
		return s.payrollResult.List(txCtx, resultFilter, limit, offset)
	})
	if err != nil {
		return err
	}
	lines := make([]entity.PayrollResultLine, 0)
	for batch := range slices.Chunk(results, domainrepo.MaxLimit) {
		resultIDs := make([]any, 0, len(batch))
		for _, result := range batch {
			if len(result.CalculationDigest) != sha256.Size ||
				(result.ResultStatus != "calculated" && result.ResultStatus != "approved") ||
				result.PaymentReferenceID != nil || result.PayslipDocumentReferenceID != nil ||
				result.PayslipPublicationReferenceID != nil || !validSettlementPrecision(result.NetPayAmount, group.CurrencyScale) {
				return hrerrors.ErrReferenceConflict
			}
			resultIDs = append(resultIDs, result.ID)
		}
		filter := []domainrepo.Filter{{Column: "payroll_result_id", Values: resultIDs}}
		count, err := s.dependencies.PayrollResultLine.CountWhere(txCtx, filter)
		if err != nil {
			return err
		}
		if int64(len(lines))+count > 10000 {
			return hrerrors.ErrConflict
		}
		components, err := operation.CollectComplete(txCtx, count, func(limit, offset int) (operation.Page[entity.PayrollResultLine], error) {
			return s.payrollResultLine.List(txCtx, filter, limit, offset)
		})
		if err != nil {
			return err
		}
		lines = append(lines, components...)
	}
	return payrollpolicy.ValidateRunResults(run, results, lines)
}

func validSettlementPrecision(value entity.Decimal, scale int16) bool {
	return scale >= 0 && scale <= 6 && value.ValidatePrecision(14+int(scale), int(scale)) == nil
}
