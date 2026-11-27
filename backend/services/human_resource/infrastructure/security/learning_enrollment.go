package security

import (
	"context"
	"reflect"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

type lockedWorkerReader interface {
	FindByIDForUpdate(context.Context, string) (*entity.Worker, error)
}

type learningEnrollmentReader interface {
	SelectWhere(context.Context, []domainrepo.Filter, int, int) ([]entity.LearningEnrollment, error)
}

type learningEnrollmentValidator struct {
	workers     lockedWorkerReader
	courses     learningCourseReader
	enrollments learningEnrollmentReader
}

// ValidateMutation memastikan penugasan awal memakai worker dan kursus yang sah.
func (v *learningEnrollmentValidator) ValidateMutation(ctx context.Context, table string, before, after entity.Validatable) error {
	if v == nil || table != "hr_learning_enrollments" || operation.RequireDependency(v.workers) != nil ||
		operation.RequireDependency(v.courses) != nil || operation.RequireDependency(v.enrollments) != nil {
		return hrerrors.ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	var previous *entity.LearningEnrollment
	if before != nil {
		var ok bool
		previous, ok = before.(*entity.LearningEnrollment)
		if !ok || previous == nil {
			return hrerrors.ErrInvalidInput
		}
	}
	if after == nil {
		if previous == nil {
			return hrerrors.ErrInvalidInput
		}
		if previous.EnrollmentStatus != entity.LearningStatusPlanned {
			return hrerrors.ErrWorkflowRequired
		}
		return v.lockParents(ctx, previous, false)
	}
	enrollment, ok := after.(*entity.LearningEnrollment)
	if !ok || enrollment == nil || (previous != nil && previous.ID != enrollment.ID) {
		return hrerrors.ErrInvalidInput
	}
	if previous != nil && previous.EnrollmentStatus != enrollment.EnrollmentStatus {
		expected := *previous
		expected.EnrollmentStatus = enrollment.EnrollmentStatus
		switch enrollment.EnrollmentStatus {
		case entity.LearningStatusInProgress:
			if (previous.EnrollmentStatus != entity.LearningStatusPlanned && previous.EnrollmentStatus != entity.LearningStatusEnrolled) ||
				previous.StartedAt != nil || previous.CompletedAt != nil || enrollment.StartedAt == nil ||
				enrollment.StartedAt.Before(enrollment.AssignedAt) {
				return hrerrors.ErrWorkflowRequired
			}
			expected.StartedAt = enrollment.StartedAt
		case entity.LearningStatusCompleted:
			if previous.EnrollmentStatus != entity.LearningStatusInProgress || previous.StartedAt == nil ||
				previous.StartedAt.Before(previous.AssignedAt) || previous.CompletedAt != nil ||
				enrollment.CompletedAt == nil || enrollment.CompletedAt.Before(*previous.StartedAt) {
				return hrerrors.ErrWorkflowRequired
			}
			expected.CompletedAt = enrollment.CompletedAt
		case entity.LearningStatusWithdrawn:
			switch previous.EnrollmentStatus {
			case entity.LearningStatusPlanned, entity.LearningStatusEnrolled, entity.LearningStatusInProgress:
			default:
				return hrerrors.ErrWorkflowRequired
			}
			if previous.CompletedAt != nil || previous.CredentialReferenceID != nil {
				return hrerrors.ErrWorkflowRequired
			}
		default:
			return hrerrors.ErrWorkflowRequired
		}
		if !reflect.DeepEqual(&expected, enrollment) {
			return hrerrors.ErrWorkflowRequired
		}
	}
	if enrollment.DueAt != nil && enrollment.DueAt.Before(enrollment.AssignedAt) {
		return hrerrors.ErrInvalidInput
	}
	if enrollment.FundingReferenceID != nil && (previous == nil || !reflect.DeepEqual(previous.FundingReferenceID, enrollment.FundingReferenceID)) {
		return hrerrors.ErrUnavailable
	}
	if err := v.lockParents(ctx, enrollment, enrollment.EnrollmentStatus != entity.LearningStatusWithdrawn); err != nil {
		return err
	}
	matches, err := v.enrollments.SelectWhere(ctx, []domainrepo.Filter{{Column: "idempotency_key", Values: []any{enrollment.IdempotencyKey}}}, 2, 0)
	if err != nil {
		return err
	}
	for _, existing := range matches {
		if existing.ID != enrollment.ID {
			return hrerrors.ErrAlreadyExists
		}
	}
	if previous == nil {
		active, err := v.enrollments.SelectWhere(ctx, []domainrepo.Filter{
			{Column: "worker_id", Values: []any{enrollment.WorkerID}},
			{Column: "course_id", Values: []any{enrollment.CourseID}},
			{Column: "enrollment_status", Values: []any{entity.LearningStatusPlanned, entity.LearningStatusEnrolled, entity.LearningStatusInProgress}},
		}, 1, 0)
		if err != nil {
			return err
		}
		if len(active) != 0 {
			return hrerrors.ErrAlreadyExists
		}
	}
	return nil
}

// lockParents memakai urutan worker lalu kursus untuk perubahan enrollment.
func (v *learningEnrollmentValidator) lockParents(ctx context.Context, enrollment *entity.LearningEnrollment, requireActive bool) error {
	worker, err := v.workers.FindByIDForUpdate(ctx, enrollment.WorkerID)
	if err != nil {
		return err
	}
	if worker == nil {
		return hrerrors.ErrReferenceConflict
	}
	course, err := v.courses.FindByIDForUpdate(ctx, enrollment.CourseID)
	if err != nil {
		return err
	}
	if course == nil {
		return hrerrors.ErrReferenceConflict
	}
	if !requireActive {
		return nil
	}
	assignedDate := enrollment.AssignedAt.UTC()
	lastCourseDate := assignedDate
	if enrollment.StartedAt != nil && enrollment.StartedAt.After(assignedDate) {
		lastCourseDate = enrollment.StartedAt.UTC()
	}
	if (worker.Status != entity.WorkerStatusActive && worker.Status != entity.WorkerStatusPreboarding) ||
		!course.IsActive ||
		(course.ValidFrom != nil && entity.CompareDates(assignedDate, *course.ValidFrom) < 0) ||
		(course.ValidTo != nil && entity.CompareDates(lastCourseDate, *course.ValidTo) > 0) {
		return hrerrors.ErrReferenceConflict
	}
	return nil
}
