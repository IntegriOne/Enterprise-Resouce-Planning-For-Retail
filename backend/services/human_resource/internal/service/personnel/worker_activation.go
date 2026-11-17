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
	return s.worker.Command(ctx, workerID, "activate", func(_ context.Context, worker *entity.Worker, actor operation.Actor, now time.Time) (*entity.Worker, error) {
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
		worker.Status = entity.WorkerStatusActive
		worker.UpdatedAt = now
		return worker, nil
	})
}
