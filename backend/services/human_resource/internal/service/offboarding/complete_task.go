package service

import (
	"context"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

type ConfirmationRequest struct {
	ReferenceID, TaskID, PlanID, OwnerManagement, PrincipalID string
}

// ConfirmationVerifier memeriksa bukti penyelesaian yang terikat pada tugas dan pelakunya.
type ConfirmationVerifier interface {
	VerifyOffboardingConfirmation(context.Context, ConfirmationRequest) error
}

// CompleteOffboardingTask mencatat penyelesaian setelah konfirmasi pemilik diverifikasi.
func (s *Service) CompleteOffboardingTask(ctx context.Context, taskID, confirmationID string) (*entity.OffboardingTask, error) {
	if s == nil || operation.RequireDependency(s.confirmation) != nil {
		return nil, hrerrors.ErrUnavailable
	}
	if err := entity.ValidateUUID("owner_confirmation_reference_id", confirmationID); err != nil {
		return nil, err
	}
	return s.offboardingTask.Command(ctx, taskID, "complete", func(txCtx context.Context, task *entity.OffboardingTask, actor operation.Actor, now time.Time) (*entity.OffboardingTask, error) {
		if actor.Kind != "human" {
			return nil, hrerrors.ErrForbidden
		}
		if task.Status != "pending" || task.OwnerManagement != "human_resource" {
			return nil, hrerrors.ErrWorkflowRequired
		}
		before := *task
		task.Status = "completed"
		task.CompletedAt = &now
		task.OwnerConfirmationReferenceID = &confirmationID
		if err := s.offboardingTask.CheckIntegrity(txCtx, &before, task); err != nil {
			return nil, err
		}
		if err := s.confirmation.VerifyOffboardingConfirmation(txCtx, ConfirmationRequest{
			ReferenceID: confirmationID, TaskID: task.ID, PlanID: task.OffboardingPlanID,
			OwnerManagement: task.OwnerManagement, PrincipalID: actor.PrincipalID,
		}); err != nil {
			return nil, operation.OwnerVerificationError(txCtx, err)
		}
		return task, nil
	})
}
