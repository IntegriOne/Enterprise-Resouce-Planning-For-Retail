package security

import (
	"context"
	"strings"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

type assignmentWorkerReader interface {
	FindByIDForUpdate(context.Context, string) (*entity.Worker, error)
	FindByID(context.Context, string) (*entity.Worker, error)
}

type assignmentPositionReader interface {
	FindByIDForUpdate(context.Context, string) (*entity.Position, error)
}

type workerAssignmentValidator struct {
	workers   assignmentWorkerReader
	positions assignmentPositionReader
}

// ValidateMutation menerima permintaan penempatan draf dengan referensi HRM yang konsisten.
func (v *workerAssignmentValidator) ValidateMutation(ctx context.Context, table string, before, after entity.Validatable) error {
	if v == nil || table != "hr_worker_assignments" ||
		operation.RequireDependency(v.workers) != nil || operation.RequireDependency(v.positions) != nil {
		return hrerrors.ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if before != nil || after == nil {
		return hrerrors.ErrWorkflowRequired
	}
	assignment, ok := after.(*entity.WorkerAssignment)
	if !ok || assignment == nil {
		return hrerrors.ErrInvalidInput
	}
	if assignment.RequestedByPrincipalID == nil || assignment.RequestedAt == nil ||
		assignment.ApprovedByPrincipalID != nil || assignment.ApprovedAt != nil ||
		assignment.EmploymentStatusSnapshot != nil || !supportedAssignmentType(assignment.AssignmentType) {
		return hrerrors.ErrWorkflowRequired
	}
	worker, err := v.workers.FindByIDForUpdate(ctx, assignment.WorkerID)
	if err != nil {
		return workerAssessmentReferenceError(err)
	}
	if worker == nil || (worker.Status != entity.WorkerStatusPreboarding && worker.Status != entity.WorkerStatusActive) ||
		worker.TerminationDate != nil ||
		(worker.HireDate != nil && entity.CompareDates(assignment.ValidFrom, *worker.HireDate) < 0) {
		return hrerrors.ErrReferenceConflict
	}
	position, err := v.positions.FindByIDForUpdate(ctx, assignment.PositionID)
	if err != nil {
		return workerAssessmentReferenceError(err)
	}
	if position == nil || (position.PositionStatus != "planned" && position.PositionStatus != "active") ||
		!strings.EqualFold(position.LegalEntityReferenceID, worker.PrimaryLegalEntityReferenceID) ||
		!assignmentPeriodWithin(assignment.ValidFrom, assignment.ValidTo, position.ValidFrom, position.ValidTo) {
		return hrerrors.ErrReferenceConflict
	}
	if assignment.ManagerWorkerID == nil {
		return nil
	}
	manager, err := v.workers.FindByID(ctx, *assignment.ManagerWorkerID)
	if err != nil {
		return workerAssessmentReferenceError(err)
	}
	if manager == nil || manager.Status != entity.WorkerStatusActive ||
		!strings.EqualFold(manager.PrimaryLegalEntityReferenceID, worker.PrimaryLegalEntityReferenceID) {
		return hrerrors.ErrReferenceConflict
	}
	return nil
}

func supportedAssignmentType(value string) bool {
	switch value {
	case "primary", "secondary", "acting", "temporary", "project", "concurrent":
		return true
	default:
		return false
	}
}

func assignmentPeriodWithin(from time.Time, to *time.Time, parentFrom time.Time, parentTo *time.Time) bool {
	return entity.CompareDates(from, parentFrom) >= 0 &&
		(parentTo == nil || (to != nil && entity.CompareDates(*to, *parentTo) <= 0))
}
