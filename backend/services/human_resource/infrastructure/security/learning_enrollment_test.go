package security

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
)

type lockedWorkerStub struct {
	worker *entity.Worker
	order  *[]string
}

func (s *lockedWorkerStub) FindByIDForUpdate(context.Context, string) (*entity.Worker, error) {
	*s.order = append(*s.order, "worker")
	return s.worker, nil
}

type learningEnrollmentStub struct {
	rows []entity.LearningEnrollment
}

func (s *learningEnrollmentStub) SelectWhere(_ context.Context, filters []domainrepo.Filter, _, _ int) ([]entity.LearningEnrollment, error) {
	if len(filters) == 3 && filters[0].Column == "worker_id" &&
		filters[1].Column == "course_id" && filters[2].Column == "enrollment_status" {
		return nil, nil
	}
	if len(filters) != 1 || filters[0].Column != "idempotency_key" {
		return nil, hrerrors.ErrInvalidInput
	}
	return s.rows, nil
}

func TestLearningEnrollmentValidatorChecksParentsDatesAndIdempotency(t *testing.T) {
	t.Parallel()
	order := []string{}
	assigned := time.Date(2026, 1, 10, 13, 0, 0, 0, time.UTC)
	validTo := time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)
	workers := &lockedWorkerStub{worker: &entity.Worker{Status: entity.WorkerStatusActive}, order: &order}
	courses := &learningCourseStub{course: &entity.LearningCourse{IsActive: true, ValidTo: &validTo}, order: &order}
	enrollments := &learningEnrollmentStub{}
	validator := &learningEnrollmentValidator{workers: workers, courses: courses, enrollments: enrollments}
	value := &entity.LearningEnrollment{
		ID: "enrollment", WorkerID: "worker", CourseID: "course", EnrollmentStatus: entity.LearningStatusPlanned,
		AssignedAt: assigned, IdempotencyKey: "assignment-1",
	}
	if err := validator.ValidateMutation(t.Context(), "hr_learning_enrollments", nil, value); err != nil {
		t.Fatalf("course must remain valid through last calendar day: %v", err)
	}
	if len(order) != 2 || order[0] != "worker" || order[1] != "course" {
		t.Fatalf("parent lock order = %v", order)
	}
	futureStart := assigned.AddDate(0, 0, 1)
	courses.course.ValidFrom = &futureStart
	courses.course.ValidTo = nil
	if err := validator.ValidateMutation(t.Context(), "hr_learning_enrollments", nil, value); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("assignment before course validity: %v", err)
	}
	courses.course.ValidFrom = &validTo
	if err := validator.ValidateMutation(t.Context(), "hr_learning_enrollments", nil, value); err != nil {
		t.Fatalf("assignment on first course validity date: %v", err)
	}
	courses.course.ValidTo = &validTo
	enrollments.rows = []entity.LearningEnrollment{{ID: "other"}}
	if err := validator.ValidateMutation(t.Context(), "hr_learning_enrollments", nil, value); !errors.Is(err, hrerrors.ErrAlreadyExists) {
		t.Fatalf("duplicate idempotency key: %v", err)
	}
	enrollments.rows = nil
	workers.worker.Status = entity.WorkerStatusTerminated
	if err := validator.ValidateMutation(t.Context(), "hr_learning_enrollments", nil, value); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("terminated worker assigned: %v", err)
	}
	workers.worker.Status = entity.WorkerStatusPreboarding
	courses.course.IsActive = false
	if err := validator.ValidateMutation(t.Context(), "hr_learning_enrollments", nil, value); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("inactive course assigned: %v", err)
	}
	courses.course.IsActive = true
	due := assigned.Add(-time.Minute)
	value.DueAt = &due
	if err := validator.ValidateMutation(t.Context(), "hr_learning_enrollments", nil, value); !errors.Is(err, hrerrors.ErrInvalidInput) {
		t.Fatalf("due date before assignment: %v", err)
	}
	value.DueAt = nil
	funding := "external-funding"
	value.FundingReferenceID = &funding
	if err := validator.ValidateMutation(t.Context(), "hr_learning_enrollments", nil, value); !errors.Is(err, hrerrors.ErrUnavailable) {
		t.Fatalf("unverified funding reference: %v", err)
	}
	value.FundingReferenceID = nil
	if err := validator.ValidateMutation(t.Context(), "hr_learning_enrollments", value, nil); err != nil {
		t.Fatalf("planned enrollment deletion: %v", err)
	}
	value.EnrollmentStatus = entity.LearningStatusEnrolled
	if err := validator.ValidateMutation(t.Context(), "hr_learning_enrollments", value, nil); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("enrolled record deletion: %v", err)
	}
}
