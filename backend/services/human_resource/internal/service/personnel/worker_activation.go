package service

import (
	"context"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

// ActivateWorker mengesahkan hubungan kerja setelah prasyarat tanggal dan kontrak diverifikasi.
func (s *Service) ActivateWorker(ctx context.Context, workerID string) (*entity.Worker, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.worker.Command(ctx, workerID, "activate", func(txCtx context.Context, worker *entity.Worker, actor operation.Actor, now time.Time) (*entity.Worker, error) {
		if actor.Kind != "human" {
			return nil, hrerrors.ErrForbidden
		}
		if worker.Status != entity.WorkerStatusPreboarding {
			return nil, hrerrors.ErrWorkflowRequired
		}
		if worker.HireDate == nil || worker.ServiceStartDate == nil ||
			entity.CompareDates(*worker.HireDate, now) > 0 ||
			entity.CompareDates(*worker.ServiceStartDate, now) > 0 || now.Before(worker.UpdatedAt) {
			return nil, hrerrors.ErrReferenceConflict
		}
		previousStatus, nextStatus := worker.Status, entity.WorkerStatusActive
		event, err := entity.NewWorkerLifecycleEvent(entity.WorkerLifecycleEvent{
			WorkerID: worker.ID, EventType: "hire", OccurredAt: now, EffectiveDate: *worker.HireDate,
			PreviousStatus: &previousStatus, NewStatus: &nextStatus, ActorPrincipalID: &actor.PrincipalID,
			EventPayload: entity.JSON("{}"), IdempotencyKey: "worker-activate:" + worker.ID,
		})
		if err != nil {
			return nil, err
		}
		if err := s.lifecycleEvents.Insert(txCtx, event); err != nil {
			return nil, err
		}
		worker.Status = nextStatus
		worker.UpdatedAt = now
		return worker, nil
	})
}
