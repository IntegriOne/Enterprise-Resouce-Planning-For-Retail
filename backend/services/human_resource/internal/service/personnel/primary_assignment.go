package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

// ApplyPrimaryAssignment menetapkan assignment primer yang telah berlaku pada pekerja aktif.
func (s *Service) ApplyPrimaryAssignment(ctx context.Context, workerID, assignmentID string) (*entity.Worker, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	if err := entity.ValidateUUID("assignment_id", assignmentID); err != nil {
		return nil, err
	}
	return s.worker.Command(ctx, workerID, "apply_primary_assignment", func(txCtx context.Context, worker *entity.Worker, actor operation.Actor, now time.Time) (*entity.Worker, error) {
		if actor.Kind != "human" {
			return nil, hrerrors.ErrForbidden
		}
		if worker.Status != entity.WorkerStatusActive || (worker.PrimaryAssignmentID != nil && strings.EqualFold(*worker.PrimaryAssignmentID, assignmentID)) {
			return nil, hrerrors.ErrWorkflowRequired
		}
		if now.Before(worker.UpdatedAt) {
			return nil, hrerrors.ErrReferenceConflict
		}
		if worker.PrimaryAssignmentID != nil {
			status := worker.Status
			event, err := entity.NewWorkerLifecycleEvent(entity.WorkerLifecycleEvent{
				WorkerID: worker.ID, EventType: "transfer", OccurredAt: now, EffectiveDate: now,
				PreviousStatus: &status, NewStatus: &status, SourceReference: &assignmentID,
				ActorPrincipalID: &actor.PrincipalID, IdempotencyKey: "primary-assignment:" + strings.ToLower(assignmentID),
				EventPayload: entity.JSON(fmt.Sprintf(
					`{"previous_assignment_id":%q,"assignment_id":%q}`, *worker.PrimaryAssignmentID, assignmentID,
				)),
			})
			if err != nil {
				return nil, err
			}
			if err := s.lifecycleEvents.Insert(txCtx, event); err != nil {
				return nil, err
			}
		}
		worker.PrimaryAssignmentID = &assignmentID
		worker.UpdatedAt = now
		return worker, nil
	})
}
