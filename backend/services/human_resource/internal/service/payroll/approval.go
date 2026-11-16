package service

import (
	"context"
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
		if run.Status != entity.PayrollRunStatusCalculated && run.Status != entity.PayrollRunStatusReview {
			return nil, hrerrors.ErrWorkflowRequired
		}
		if strings.EqualFold(run.StartedByPrincipalID, actor.PrincipalID) {
			return nil, hrerrors.ErrForbidden
		}
		if run.ApprovedAt != nil || run.ApprovedByPrincipalID != nil {
			return nil, hrerrors.ErrConflict
		}
		resultFilter := []domainrepo.Filter{{Column: "payroll_run_id", Values: []any{runID}}}
		expected, err := s.dependencies.PayrollResult.CountWhere(txCtx, resultFilter)
		if err != nil {
			return nil, err
		}
		results, err := operation.CollectComplete(txCtx, expected, func(limit, offset int) (operation.Page[entity.PayrollResult], error) {
			return s.payrollResult.List(txCtx, resultFilter, limit, offset)
		})
		if err != nil {
			return nil, err
		}
		lines := make([]entity.PayrollResultLine, 0)
		for _, result := range results {
			filter := []domainrepo.Filter{{Column: "payroll_result_id", Values: []any{result.ID}}}
			count, err := s.dependencies.PayrollResultLine.CountWhere(txCtx, filter)
			if err != nil {
				return nil, err
			}
			if int64(len(lines))+count > 10000 {
				return nil, hrerrors.ErrConflict
			}
			components, err := operation.CollectComplete(txCtx, count, func(limit, offset int) (operation.Page[entity.PayrollResultLine], error) {
				return s.payrollResultLine.List(txCtx, filter, limit, offset)
			})
			if err != nil {
				return nil, err
			}
			lines = append(lines, components...)
		}
		if err := payrollpolicy.ValidateRunResults(run, results, lines); err != nil {
			return nil, err
		}
		run.Status = entity.PayrollRunStatusApproved
		run.ApprovedByPrincipalID = &actor.PrincipalID
		run.ApprovedAt = &now
		return run, nil
	})
}
