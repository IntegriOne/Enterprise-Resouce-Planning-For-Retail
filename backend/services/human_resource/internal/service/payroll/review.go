package service

import (
	"context"
	"strings"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

// SubmitPayrollRunForReview menyerahkan snapshot hasil dari penginisiasi run kepada penyetuju.
func (s *Service) SubmitPayrollRunForReview(ctx context.Context, id string) (*entity.PayrollRun, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.payrollRun.Command(ctx, id, "submit_review", func(txCtx context.Context, run *entity.PayrollRun, actor operation.Actor, now time.Time) (*entity.PayrollRun, error) {
		if actor.Kind != "human" || !strings.EqualFold(run.StartedByPrincipalID, actor.PrincipalID) {
			return nil, hrerrors.ErrForbidden
		}
		if run.Status != entity.PayrollRunStatusCalculated || run.CalculatedAt == nil ||
			run.CalculatedAt.Before(run.CreatedAt) || now.Before(*run.CalculatedAt) || len(run.OutputDigest) == 0 ||
			run.ApprovedAt != nil || run.ApprovedByPrincipalID != nil {
			return nil, hrerrors.ErrWorkflowRequired
		}
		if err := s.validateRunResults(txCtx, run); err != nil {
			return nil, err
		}
		run.Status = entity.PayrollRunStatusReview
		return run, nil
	})
}
