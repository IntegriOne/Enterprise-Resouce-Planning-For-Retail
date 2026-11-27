package service

import (
	"context"
	"crypto/sha256"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

// RequestPayrollPosting mencatat permintaan jurnal untuk Finance melalui outbox HRM.
func (s *Service) RequestPayrollPosting(ctx context.Context, id string) (*entity.PayrollRun, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.payrollRun.Command(ctx, id, "request_posting", func(txCtx context.Context, run *entity.PayrollRun, actor operation.Actor, now time.Time) (*entity.PayrollRun, error) {
		if actor.Kind != "human" {
			return nil, hrerrors.ErrForbidden
		}
		if run.Status != entity.PayrollRunStatusApproved || run.ApprovedAt == nil || run.ApprovedByPrincipalID == nil ||
			now.Before(*run.ApprovedAt) || len(run.InputDigest) != sha256.Size || len(run.OutputDigest) != sha256.Size ||
			run.PostingRequestedAt != nil || run.PostingRequestedByPrincipalID != nil ||
			run.FinanceJournalReferenceID != nil || run.PaymentBatchReferenceID != nil ||
			run.FinanceSettlementEventReferenceID != nil || run.FinanceSettledAt != nil {
			return nil, hrerrors.ErrWorkflowRequired
		}
		if err := s.validateRunResults(txCtx, run); err != nil {
			return nil, err
		}
		run.PostingRequestedAt, run.PostingRequestedByPrincipalID = &now, &actor.PrincipalID
		return run, nil
	})
}
