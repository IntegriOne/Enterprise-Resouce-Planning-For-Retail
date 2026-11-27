package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

type InputSnapshotStore interface {
	Capture(ctx context.Context, id, periodID, principalID string, at time.Time) (entity.PayrollInputSnapshot, error)
	SelectMetadata(ctx context.Context, id string) (entity.PayrollInputSnapshot, error)
}

// FreezePayrollInput membekukan input dari periode yang diizinkan dan mengembalikan metadata.
func (s *Service) FreezePayrollInput(ctx context.Context, periodID, idempotencyKey string) (entity.PayrollInputSnapshot, error) {
	var result entity.PayrollInputSnapshot
	if s == nil || operation.RequireDependency(s.dependencies.InputSnapshots) != nil {
		return result, hrerrors.ErrUnavailable
	}
	if err := entity.ValidateUUID("payroll_period_id", periodID); err != nil {
		return result, err
	}
	if err := entity.ValidateUUID("idempotency_key", idempotencyKey); err != nil {
		return result, err
	}
	deps := s.dependencies
	access := operation.Access{Domain: "payroll", Resource: "hr_payroll_runs", Action: "create"}
	actor, err := deps.Runtime.Policy.Authorize(ctx, access)
	if err != nil {
		return result, err
	}
	if actor.Kind != "human" || entity.ValidateUUID("principal_id", actor.PrincipalID) != nil {
		return result, hrerrors.ErrForbidden
	}
	periodID = strings.ToLower(periodID)
	snapshotID := uuid.NewSHA1(uuid.NameSpaceOID, []byte("hrm.payroll.snapshot/v1:"+strings.ToLower(actor.PrincipalID)+":"+periodID+":"+strings.ToLower(idempotencyKey))).String()
	err = deps.Runtime.Transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		snapshot, err := deps.InputSnapshots.SelectMetadata(txCtx, snapshotID)
		if errors.Is(err, hrerrors.ErrNotFound) {
			snapshot, err = deps.InputSnapshots.Capture(txCtx, snapshotID, periodID, actor.PrincipalID, deps.Runtime.Clock.Now())
			if errors.Is(err, hrerrors.ErrWorkflowRequired) {
				existing, readErr := deps.InputSnapshots.SelectMetadata(txCtx, snapshotID)
				if !errors.Is(readErr, hrerrors.ErrNotFound) {
					snapshot, err = existing, readErr
				}
			}
		}
		if err != nil {
			return err
		}
		if !strings.EqualFold(snapshot.ID, snapshotID) || !strings.EqualFold(snapshot.PeriodID, periodID) ||
			!strings.EqualFold(snapshot.CreatedBy, actor.PrincipalID) {
			return hrerrors.ErrReferenceConflict
		}
		checked, err := deps.Runtime.Policy.Authorize(txCtx, access)
		if err != nil {
			return err
		}
		if checked != actor {
			return hrerrors.ErrForbidden
		}
		snapshot.Content = nil
		result = snapshot
		return txCtx.Err()
	})
	if err != nil {
		return entity.PayrollInputSnapshot{}, err
	}
	return result, nil
}
