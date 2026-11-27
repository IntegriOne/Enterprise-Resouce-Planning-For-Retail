package security

import (
	"context"
	"math/big"
	"reflect"
	"strings"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	personnelpolicy "github.com/integrione/erp-retail/backend/services/human_resource/domain/service/personnel"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

type assignmentWorkerReader interface {
	FindByIDForUpdate(context.Context, string) (*entity.Worker, error)
	FindByID(context.Context, string) (*entity.Worker, error)
}

type assignmentPositionReader interface {
	FindByIDForUpdate(context.Context, string) (*entity.Position, error)
}

type assignmentListReader interface {
	CountWhere(context.Context, []domainrepo.Filter) (int64, error)
	FindByWorkerID(context.Context, string, int, int) ([]entity.WorkerAssignment, error)
	FindByPositionID(context.Context, string, int, int) ([]entity.WorkerAssignment, error)
}

type workerAssignmentValidator struct {
	workers     assignmentWorkerReader
	positions   assignmentPositionReader
	assignments assignmentListReader
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
	if after == nil {
		return hrerrors.ErrWorkflowRequired
	}
	assignment, ok := after.(*entity.WorkerAssignment)
	if !ok || assignment == nil {
		return hrerrors.ErrInvalidInput
	}
	approval := before != nil
	if approval {
		previous, ok := before.(*entity.WorkerAssignment)
		if !ok || previous == nil {
			return hrerrors.ErrInvalidInput
		}
		if !validAssignmentApproval(previous, assignment) {
			return hrerrors.ErrWorkflowRequired
		}
	} else if assignment.ApprovedByPrincipalID != nil || assignment.ApprovedAt != nil || assignment.EmploymentStatusSnapshot != nil {
		return hrerrors.ErrWorkflowRequired
	}
	if assignment.RequestedByPrincipalID == nil || assignment.RequestedAt == nil || !supportedAssignmentType(assignment.AssignmentType) {
		return hrerrors.ErrWorkflowRequired
	}
	if approval && operation.RequireDependency(v.assignments) != nil {
		return hrerrors.ErrUnavailable
	}
	worker, err := v.workers.FindByIDForUpdate(ctx, assignment.WorkerID)
	if err != nil {
		return workerAssessmentReferenceError(err)
	}
	if worker == nil || (worker.Status != entity.WorkerStatusPreboarding && worker.Status != entity.WorkerStatusActive) ||
		worker.TerminationDate != nil ||
		(worker.HireDate != nil && entity.CompareDates(assignment.ValidFrom, *worker.HireDate) < 0) ||
		(worker.ServiceStartDate != nil && entity.CompareDates(assignment.ValidFrom, *worker.ServiceStartDate) < 0) {
		return hrerrors.ErrReferenceConflict
	}
	if approval && worker.Status != entity.WorkerStatusActive {
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
	if approval && position.PositionStatus != "active" {
		return hrerrors.ErrReferenceConflict
	}
	if err := verifyAssignmentManager(ctx, v.workers, worker, assignment); err != nil {
		return err
	}
	if approval {
		return v.checkAssignmentCapacity(ctx, assignment, position)
	}
	return nil
}

func verifyAssignmentManager(ctx context.Context, reader assignmentWorkerReader, worker *entity.Worker, assignment *entity.WorkerAssignment) error {
	if assignment.ManagerWorkerID == nil {
		return nil
	}
	if operation.RequireDependency(reader) != nil {
		return hrerrors.ErrUnavailable
	}
	manager, err := reader.FindByID(ctx, *assignment.ManagerWorkerID)
	if err != nil {
		return workerAssessmentReferenceError(err)
	}
	return personnelpolicy.ValidateAssignmentManager(worker, manager, *assignment.ManagerWorkerID, assignment.ValidFrom, assignment.ValidTo)
}

func validAssignmentApproval(before, after *entity.WorkerAssignment) bool {
	if before.ApprovedByPrincipalID != nil || before.ApprovedAt != nil || before.EmploymentStatusSnapshot != nil ||
		before.RequestedByPrincipalID == nil || before.RequestedAt == nil ||
		after.ApprovedByPrincipalID == nil || after.ApprovedAt == nil || after.EmploymentStatusSnapshot == nil ||
		*after.EmploymentStatusSnapshot != string(entity.WorkerStatusActive) ||
		strings.EqualFold(*before.RequestedByPrincipalID, *after.ApprovedByPrincipalID) ||
		after.ApprovedAt.Before(*before.RequestedAt) {
		return false
	}
	copy := *after
	copy.ApprovedByPrincipalID = nil
	copy.ApprovedAt = nil
	copy.EmploymentStatusSnapshot = nil
	return reflect.DeepEqual(before, &copy)
}

func (v *workerAssignmentValidator) checkAssignmentCapacity(ctx context.Context, candidate *entity.WorkerAssignment, position *entity.Position) error {
	workerRows, err := v.loadAssignments(ctx, "worker_id", candidate.WorkerID, v.assignments.FindByWorkerID)
	if err != nil {
		return err
	}
	positionRows, err := v.loadAssignments(ctx, "position_id", candidate.PositionID, v.assignments.FindByPositionID)
	if err != nil {
		return err
	}
	workerLimit := big.NewRat(100, 1)
	if err := personnelpolicy.CheckAssignmentCapacity(candidate, workerRows, candidate.ValidFrom, candidate.ValidTo, workerLimit); err != nil {
		return err
	}
	headcount, err := position.HeadcountAuthorized.Rat()
	if err != nil {
		return hrerrors.ErrReferenceConflict
	}
	positionLimit := new(big.Rat).Mul(headcount, workerLimit)
	return personnelpolicy.CheckAssignmentCapacity(candidate, positionRows, candidate.ValidFrom, candidate.ValidTo, positionLimit)
}

func (v *workerAssignmentValidator) loadAssignments(ctx context.Context, column, id string, find func(context.Context, string, int, int) ([]entity.WorkerAssignment, error)) ([]entity.WorkerAssignment, error) {
	count, err := v.assignments.CountWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{id}}})
	if err != nil {
		return nil, err
	}
	return operation.CollectComplete(ctx, count, func(limit, offset int) (operation.Page[entity.WorkerAssignment], error) {
		rows, err := find(ctx, id, limit, offset)
		return operation.Page[entity.WorkerAssignment]{Items: rows}, err
	})
}

func assignmentContainsDate(row *entity.WorkerAssignment, day time.Time) bool {
	return row.CancelledAt == nil && entity.CompareDates(row.ValidFrom, day) <= 0 && (row.ValidTo == nil || entity.CompareDates(*row.ValidTo, day) >= 0)
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
