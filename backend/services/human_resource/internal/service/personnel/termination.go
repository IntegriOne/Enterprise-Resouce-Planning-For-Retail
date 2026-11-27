package service

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"strings"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	personnelpolicy "github.com/integrione/erp-retail/backend/services/human_resource/domain/service/personnel"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

// TerminateWorker menerapkan keputusan offboarding dan mencatat seluruh perubahan dalam satu transaksi.
func (s *Service) TerminateWorker(ctx context.Context, workerID, planID string) (*entity.Worker, error) {
	if s == nil || operation.RequireDependency(s.transferDependencies.OffboardingPlan) != nil {
		return nil, hrerrors.ErrUnavailable
	}
	if err := entity.ValidateUUID("worker_id", workerID); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("offboarding_plan_id", planID); err != nil {
		return nil, err
	}
	deps := s.transferDependencies
	access := operation.Access{Domain: "personnel", Resource: "hr_workers", Action: "terminate"}
	actor, err := authorizeTransfer(ctx, deps.Runtime.Policy, access)
	if err != nil {
		return nil, err
	}
	var result *entity.Worker
	err = deps.Runtime.Transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		plan, err := deps.OffboardingPlan.FindByIDForUpdate(txCtx, planID)
		if err != nil {
			return err
		}
		if plan == nil || !strings.EqualFold(plan.WorkerID, workerID) {
			return hrerrors.ErrReferenceConflict
		}
		worker, err := deps.Worker.FindByIDForUpdate(txCtx, workerID)
		if err != nil {
			return err
		}
		if worker == nil {
			return hrerrors.ErrReferenceConflict
		}
		access.Attributes = specWorker().Attributes(worker)
		verifyActor := func() error {
			checked, err := authorizeTransfer(txCtx, deps.Runtime.Policy, access)
			if err != nil {
				return err
			}
			if checked != actor {
				return hrerrors.ErrForbidden
			}
			return txCtx.Err()
		}
		if err := verifyActor(); err != nil {
			return err
		}
		if worker.IAMPrincipalReferenceID != nil && strings.EqualFold(*worker.IAMPrincipalReferenceID, actor.PrincipalID) {
			return hrerrors.ErrForbidden
		}
		now := deps.Runtime.Clock.Now()
		next, err := personnelpolicy.BuildTerminatedWorker(worker, plan, now)
		if err != nil {
			return err
		}
		access.Attributes = specWorker().Attributes(next)
		if err := verifyActor(); err != nil {
			return err
		}
		if err := closeTerminationAssignments(txCtx, deps.WorkerAssignment, worker.ID, plan.LastWorkingDate, now); err != nil {
			return err
		}
		if err := closeTerminationContracts(txCtx, deps.EmploymentContract, worker.ID, plan.LastWorkingDate); err != nil {
			return err
		}
		if err := deps.Worker.Update(txCtx, next); err != nil {
			return err
		}
		if err := s.recordTermination(txCtx, worker, next, plan.ID, actor, now); err != nil {
			return err
		}
		result = next
		return verifyActor()
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func closeTerminationAssignments(ctx context.Context, repo domainrepo.RecordRepository[entity.WorkerAssignment, string], workerID string, lastDay, now time.Time) error {
	rows, err := terminationRows(ctx, repo, workerID)
	if err != nil {
		return err
	}
	for _, row := range rows {
		if row.CancelledAt != nil || row.ApprovedAt == nil || (row.ValidTo != nil && entity.CompareDates(*row.ValidTo, lastDay) <= 0) {
			continue
		}
		if entity.CompareDates(row.ValidFrom, lastDay) > 0 {
			if now.Before(*row.ApprovedAt) {
				return hrerrors.ErrReferenceConflict
			}
			row.CancelledAt = &now
		} else {
			row.ValidTo = &lastDay
		}
		if err := row.Validate(); err != nil {
			return err
		}
		if err := repo.Update(ctx, &row); err != nil {
			return err
		}
	}
	return nil
}

func closeTerminationContracts(ctx context.Context, repo domainrepo.RecordRepository[entity.EmploymentContract, string], workerID string, lastDay time.Time) error {
	rows, err := terminationRows(ctx, repo, workerID)
	if err != nil {
		return err
	}
	for _, row := range rows {
		if row.ContractStatus != "signed" || (row.EffectiveTo != nil && entity.CompareDates(*row.EffectiveTo, lastDay) <= 0) {
			continue
		}
		if entity.CompareDates(row.EffectiveFrom, lastDay) > 0 {
			row.ContractStatus = "cancelled"
		} else {
			row.EffectiveTo = &lastDay
		}
		if err := row.Validate(); err != nil {
			return err
		}
		if err := repo.Update(ctx, &row); err != nil {
			return err
		}
	}
	return nil
}

func terminationRows[T any](ctx context.Context, repo domainrepo.RecordRepository[T, string], workerID string) ([]T, error) {
	filters := []domainrepo.Filter{{Column: "worker_id", Values: []any{workerID}}}
	count, err := repo.CountWhere(ctx, filters)
	if err != nil {
		return nil, err
	}
	return operation.CollectComplete(ctx, count, func(limit, offset int) (operation.Page[T], error) {
		rows, err := repo.SelectWhere(ctx, filters, limit, offset)
		return operation.Page[T]{Items: rows}, err
	})
}

func (s *Service) recordTermination(ctx context.Context, worker, next *entity.Worker, planID string, actor operation.Actor, now time.Time) error {
	payload, err := json.Marshal(map[string]string{
		"worker_id": worker.ID, "offboarding_plan_id": planID, "actor_principal_id": actor.PrincipalID,
		"termination_effective_date": next.TerminationDate.Format(time.DateOnly),
	})
	if err != nil {
		return err
	}
	event, err := entity.NewWorkerLifecycleEvent(entity.WorkerLifecycleEvent{
		WorkerID: worker.ID, EventType: "termination", OccurredAt: now, EffectiveDate: *next.TerminationDate,
		PreviousStatus: &worker.Status, NewStatus: &next.Status, SourceReference: &planID,
		ActorPrincipalID: &actor.PrincipalID, EventPayload: entity.JSON(payload), IdempotencyKey: "worker-termination:" + planID,
	})
	if err != nil {
		return err
	}
	if err := s.lifecycleEvents.Insert(ctx, event); err != nil {
		return err
	}
	deps := s.transferDependencies.Runtime
	id, err := deps.IDs.NewID()
	if err != nil {
		return err
	}
	digest := sha256.Sum256(payload)
	outbox, err := entity.NewOutboxEvent(entity.OutboxEvent{
		ID: id, AggregateType: "hr_worker_terminations", AggregateID: planID, AggregateVersion: 1,
		EventType: entity.WorkerTerminationRecorded, SchemaVersion: 1, Payload: entity.JSON(payload),
		PayloadDigest: digest[:], OccurredAt: now, AvailableAt: now,
	})
	if err != nil {
		return err
	}
	return deps.Outbox.Insert(ctx, outbox)
}
