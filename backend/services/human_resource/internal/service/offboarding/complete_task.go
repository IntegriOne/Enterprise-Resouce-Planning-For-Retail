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

// CompleteOffboardingTask mencatat konfirmasi petugas HR atau memverifikasi bukti pemilik yang diberikan.
func (s *Service) CompleteOffboardingTask(ctx context.Context, taskID, confirmationID string) (*entity.OffboardingTask, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	if confirmationID != "" {
		if err := entity.ValidateUUID("owner_confirmation_reference_id", confirmationID); err != nil {
			return nil, err
		}
		if operation.RequireDependency(s.confirmation) != nil {
			return nil, hrerrors.ErrUnavailable
		}
	}
	return s.offboardingTask.Command(ctx, taskID, "complete", func(txCtx context.Context, task *entity.OffboardingTask, actor operation.Actor, now time.Time) (*entity.OffboardingTask, error) {
		if actor.Kind != "human" {
			return nil, hrerrors.ErrForbidden
		}
		if task.Status != "pending" || task.OwnerManagement != "human_resource" {
			return nil, hrerrors.ErrWorkflowRequired
		}
		referenceID := confirmationID
		if referenceID == "" {
			var err error
			referenceID, err = s.ids.NewID()
			if err != nil {
				return nil, err
			}
			if err := entity.ValidateUUID("owner_confirmation_reference_id", referenceID); err != nil {
				return nil, err
			}
		}
		before := *task
		task.Status = "completed"
		task.CompletedAt = &now
		task.OwnerConfirmationReferenceID = &referenceID
		if err := s.offboardingTask.CheckIntegrity(txCtx, &before, task); err != nil {
			return nil, err
		}
		if confirmationID != "" {
			if err := s.confirmation.VerifyOffboardingConfirmation(txCtx, ConfirmationRequest{
				ReferenceID: referenceID, TaskID: task.ID, PlanID: task.OffboardingPlanID,
				OwnerManagement: task.OwnerManagement, PrincipalID: actor.PrincipalID,
			}); err != nil {
				return nil, operation.OwnerVerificationError(txCtx, err)
			}
		}
		return task, nil
	})
}
