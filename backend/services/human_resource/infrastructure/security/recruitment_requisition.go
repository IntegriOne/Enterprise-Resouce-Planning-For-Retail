package security

import (
	"context"
	"errors"
	"reflect"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository"
)

type requisitionWorkerReader interface {
	SelectWhere(context.Context, []domainrepo.Filter, int, int) ([]entity.Worker, error)
	FindByIDForUpdate(context.Context, string) (*entity.Worker, error)
}

type requisitionJobReader interface {
	FindByIDForUpdate(context.Context, string) (*entity.Job, error)
}

type requisitionPositionReader interface {
	FindByIDForUpdate(context.Context, string) (*entity.Position, error)
}

type requisitionReader interface {
	SelectWhere(context.Context, []domainrepo.Filter, int, int) ([]entity.Requisition, error)
}

type requisitionReferenceCounter interface {
	CountWhere(context.Context, []domainrepo.Filter) (int64, error)
}

type requisitionValidator struct {
	workers      requisitionWorkerReader
	jobs         requisitionJobReader
	positions    requisitionPositionReader
	requisitions requisitionReader
	applications requisitionReferenceCounter
	postings     requisitionReferenceCounter
	owner        PositionReferenceVerifier
}

// RecruitmentRequisitionValidators memasang aturan draf lowongan dengan bukti unit dari Foundation.
func RecruitmentRequisitionValidators(repos *repository.Registry, owner PositionReferenceVerifier) (map[string]MutationValidator, error) {
	if repos == nil || operation.RequireDependency(repos.Worker) != nil ||
		operation.RequireDependency(repos.Job) != nil || operation.RequireDependency(repos.Position) != nil ||
		operation.RequireDependency(repos.Requisition) != nil || operation.RequireDependency(repos.Application) != nil ||
		operation.RequireDependency(repos.JobPosting) != nil {
		return nil, hrerrors.ErrUnavailable
	}
	return map[string]MutationValidator{
		"hr_requisitions": &requisitionValidator{
			workers: repos.Worker, jobs: repos.Job, positions: repos.Position,
			requisitions: repos.Requisition, applications: repos.Application,
			postings: repos.JobPosting, owner: owner,
		},
	}, nil
}

func (v *requisitionValidator) ValidateMutation(ctx context.Context, table string, before, after entity.Validatable) error {
	if v == nil || table != "hr_requisitions" || operation.RequireDependency(v.workers) != nil ||
		operation.RequireDependency(v.jobs) != nil || operation.RequireDependency(v.positions) != nil ||
		operation.RequireDependency(v.requisitions) != nil || operation.RequireDependency(v.applications) != nil ||
		operation.RequireDependency(v.postings) != nil {
		return hrerrors.ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	var previous *entity.Requisition
	if before != nil {
		var ok bool
		previous, ok = before.(*entity.Requisition)
		if !ok || previous == nil || previous.Status != "draft" {
			return hrerrors.ErrWorkflowRequired
		}
	}
	if previous == nil && after == nil {
		return hrerrors.ErrInvalidInput
	}
	if previous != nil {
		if err := v.rejectIfReferenced(ctx, previous.ID); err != nil {
			return err
		}
		if after == nil {
			return nil
		}
	}
	value, ok := after.(*entity.Requisition)
	if !ok || value == nil {
		return hrerrors.ErrInvalidInput
	}
	submitting := previous != nil && value.Status == "submitted"
	if submitting {
		original := *value
		original.Status = "draft"
		original.SubmittedAt = nil
		original.SubmittedByPrincipalID = nil
		if previous.CreatedByPrincipalID == nil || previous.SubmittedAt != nil || previous.SubmittedByPrincipalID != nil ||
			value.SubmittedAt == nil || value.SubmittedByPrincipalID == nil ||
			*value.SubmittedByPrincipalID != *previous.CreatedByPrincipalID || !reflect.DeepEqual(previous, &original) {
			return hrerrors.ErrWorkflowRequired
		}
	}
	if (!submitting && (value.Status != "draft" || value.SubmittedAt != nil || value.SubmittedByPrincipalID != nil)) ||
		value.CreatedByPrincipalID == nil || value.ApprovedAt != nil || value.ApprovedByPrincipalID != nil ||
		value.OpenedAt != nil || value.ClosedAt != nil ||
		(previous != nil && (previous.ID != value.ID || previous.RequisitionNumber != value.RequisitionNumber ||
			previous.HiringManagerPrincipalID != value.HiringManagerPrincipalID || previous.IdempotencyKey != value.IdempotencyKey ||
			!reflect.DeepEqual(previous.CreatedByPrincipalID, value.CreatedByPrincipalID))) {
		return hrerrors.ErrWorkflowRequired
	}
	if value.WorkforcePlanLineID != nil || value.BudgetReferenceID != nil || value.RecruiterPrincipalID != nil {
		return hrerrors.ErrUnavailable
	}
	switch value.RequisitionType {
	case "replacement", "new_headcount", "temporary":
	default:
		return hrerrors.ErrWorkflowRequired
	}
	now := time.Now()
	if value.TargetStartDate != nil && entity.CompareDates(*value.TargetStartDate, now) < 0 {
		return hrerrors.ErrReferenceConflict
	}
	manager, err := v.activeManager(ctx, value.HiringManagerPrincipalID, now, value.TargetStartDate)
	if err != nil {
		return err
	}
	job, err := v.jobs.FindByIDForUpdate(ctx, value.JobID)
	if err != nil {
		return workerAssessmentReferenceError(err)
	}
	if job == nil || !job.IsActive || entity.CompareDates(now, job.ValidFrom) < 0 ||
		(job.ValidTo != nil && (entity.CompareDates(now, *job.ValidTo) > 0 ||
			value.TargetStartDate != nil && entity.CompareDates(*value.TargetStartDate, *job.ValidTo) > 0)) {
		return hrerrors.ErrReferenceConflict
	}
	if value.PositionID != nil {
		position, err := v.positions.FindByIDForUpdate(ctx, *value.PositionID)
		if err != nil {
			return workerAssessmentReferenceError(err)
		}
		if position == nil || position.PositionStatus != "active" || position.JobID != value.JobID ||
			position.OrganizationUnitReferenceID != value.OrganizationUnitReferenceID ||
			position.LegalEntityReferenceID != manager.PrimaryLegalEntityReferenceID ||
			entity.CompareDates(now, position.ValidFrom) < 0 ||
			(position.ValidTo != nil && (entity.CompareDates(now, *position.ValidTo) > 0 ||
				value.TargetStartDate != nil && entity.CompareDates(*value.TargetStartDate, *position.ValidTo) > 0)) ||
			(value.WorkLocationReferenceID != nil && !reflect.DeepEqual(value.WorkLocationReferenceID, position.WorkLocationReferenceID)) {
			return hrerrors.ErrReferenceConflict
		}
	}
	if operation.RequireDependency(v.owner) != nil {
		return hrerrors.ErrUnavailable
	}
	if err := v.owner.VerifyPositionReferences(ctx, PositionReferenceCheck{
		OrganizationUnitID: value.OrganizationUnitReferenceID,
		LegalEntityID:      manager.PrimaryLegalEntityReferenceID,
		WorkLocationID:     value.WorkLocationReferenceID,
		ValidFrom:          now, ValidTo: value.TargetStartDate,
	}); err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if errors.Is(err, hrerrors.ErrReferenceConflict) {
			return hrerrors.ErrReferenceConflict
		}
		return hrerrors.ErrUnavailable
	}
	if previous == nil {
		matches, err := v.requisitions.SelectWhere(ctx, []domainrepo.Filter{{Column: "requisition_number", Values: []any{value.RequisitionNumber}}}, 2, 0)
		if err != nil {
			return err
		}
		if len(matches) != 0 {
			return hrerrors.ErrAlreadyExists
		}
	}
	return nil
}

func (v *requisitionValidator) activeManager(ctx context.Context, principalID string, now time.Time, targetDate *time.Time) (*entity.Worker, error) {
	workers, err := v.workers.SelectWhere(ctx, []domainrepo.Filter{{Column: "iam_principal_reference_id", Values: []any{principalID}}}, 2, 0)
	if err != nil {
		return nil, err
	}
	if len(workers) != 1 {
		return nil, hrerrors.ErrReferenceConflict
	}
	worker, err := v.workers.FindByIDForUpdate(ctx, workers[0].ID)
	if err != nil {
		return nil, workerAssessmentReferenceError(err)
	}
	if worker == nil || worker.Status != entity.WorkerStatusActive || worker.IAMPrincipalReferenceID == nil ||
		*worker.IAMPrincipalReferenceID != principalID || worker.PrimaryLegalEntityReferenceID == "" ||
		(worker.TerminationDate != nil && (entity.CompareDates(*worker.TerminationDate, now) <= 0 ||
			targetDate != nil && entity.CompareDates(*worker.TerminationDate, *targetDate) <= 0)) {
		return nil, hrerrors.ErrReferenceConflict
	}
	return worker, nil
}

func (v *requisitionValidator) rejectIfReferenced(ctx context.Context, id string) error {
	for _, counter := range []requisitionReferenceCounter{v.applications, v.postings} {
		count, err := counter.CountWhere(ctx, []domainrepo.Filter{{Column: "requisition_id", Values: []any{id}}})
		if err != nil {
			return err
		}
		if count != 0 {
			return hrerrors.ErrReferenceConflict
		}
	}
	return nil
}
