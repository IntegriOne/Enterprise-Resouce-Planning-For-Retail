package service

import (
	"context"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

// StartLearningEnrollment mencatat waktu mulai untuk penugasan yang belum dijalankan.
func (s *Service) StartLearningEnrollment(ctx context.Context, id string) (*entity.LearningEnrollment, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.learningEnrollment.Command(ctx, id, "start", func(_ context.Context, enrollment *entity.LearningEnrollment, actor operation.Actor, now time.Time) (*entity.LearningEnrollment, error) {
		if actor.Kind != "human" {
			return nil, hrerrors.ErrForbidden
		}
		if (enrollment.EnrollmentStatus != entity.LearningStatusPlanned && enrollment.EnrollmentStatus != entity.LearningStatusEnrolled) ||
			enrollment.StartedAt != nil || enrollment.CompletedAt != nil || enrollment.Score != nil ||
			enrollment.CredentialReferenceID != nil || now.Before(enrollment.AssignedAt) {
			return nil, hrerrors.ErrWorkflowRequired
		}
		enrollment.EnrollmentStatus = entity.LearningStatusInProgress
		enrollment.StartedAt = &now
		return enrollment, nil
	})
}

// CompleteLearningEnrollment mencatat penyelesaian administratif oleh petugas berizin.
func (s *Service) CompleteLearningEnrollment(ctx context.Context, id string) (*entity.LearningEnrollment, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.learningEnrollment.Command(ctx, id, "complete", func(_ context.Context, enrollment *entity.LearningEnrollment, actor operation.Actor, now time.Time) (*entity.LearningEnrollment, error) {
		if actor.Kind != "human" {
			return nil, hrerrors.ErrForbidden
		}
		if enrollment.EnrollmentStatus != entity.LearningStatusInProgress || enrollment.StartedAt == nil ||
			enrollment.StartedAt.Before(enrollment.AssignedAt) || now.Before(*enrollment.StartedAt) || enrollment.CompletedAt != nil {
			return nil, hrerrors.ErrWorkflowRequired
		}
		enrollment.EnrollmentStatus = entity.LearningStatusCompleted
		enrollment.CompletedAt = &now
		return enrollment, nil
	})
}

// WithdrawLearningEnrollment mengakhiri penugasan tanpa mencatat kelulusan.
func (s *Service) WithdrawLearningEnrollment(ctx context.Context, id string) (*entity.LearningEnrollment, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.learningEnrollment.Command(ctx, id, "withdraw", func(_ context.Context, enrollment *entity.LearningEnrollment, actor operation.Actor, _ time.Time) (*entity.LearningEnrollment, error) {
		if actor.Kind != "human" {
			return nil, hrerrors.ErrForbidden
		}
		switch enrollment.EnrollmentStatus {
		case entity.LearningStatusPlanned, entity.LearningStatusEnrolled, entity.LearningStatusInProgress:
		default:
			return nil, hrerrors.ErrWorkflowRequired
		}
		if enrollment.CompletedAt != nil || enrollment.CredentialReferenceID != nil {
			return nil, hrerrors.ErrWorkflowRequired
		}
		enrollment.EnrollmentStatus = entity.LearningStatusWithdrawn
		return enrollment, nil
	})
}
