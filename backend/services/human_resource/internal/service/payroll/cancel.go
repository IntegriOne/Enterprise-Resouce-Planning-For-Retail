package service

import (
	"context"
	"strings"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

// CancelPayrollRun membatalkan review tanpa mengubah snapshot hasil payroll.
func (s *Service) CancelPayrollRun(ctx context.Context, id, reason string) (*entity.PayrollRun, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	reason = strings.TrimSpace(reason)
	return s.payrollRun.Command(ctx, id, "cancel", func(_ context.Context, run *entity.PayrollRun, actor operation.Actor, now time.Time) (*entity.PayrollRun, error) {
		if run.Status != entity.PayrollRunStatusReview {
			return nil, hrerrors.ErrWorkflowRequired
		}
		if actor.Kind != "human" || strings.EqualFold(run.StartedByPrincipalID, actor.PrincipalID) {
			return nil, hrerrors.ErrForbidden
		}
		run.Status = entity.PayrollRunStatusCancelled
		run.CancelledAt, run.CancelledByPrincipalID, run.CancellationReason = &now, &actor.PrincipalID, &reason
		return run, nil
	})
}
