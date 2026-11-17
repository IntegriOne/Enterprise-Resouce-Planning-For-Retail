package security

import (
	"context"
	"errors"
	"reflect"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository"
)

// WorkerReferenceCheck memuat referensi identitas dan badan hukum yang dimiliki Foundation.
type WorkerReferenceCheck struct {
	PersonPartyID string
	LegalEntityID string
}

// WorkerReferenceVerifier membuktikan bahwa party adalah orang dan badan hukum aktif.
type WorkerReferenceVerifier interface {
	VerifyWorkerReferences(context.Context, WorkerReferenceCheck) error
}

type workerNumberReader interface {
	FindByWorkerNumber(context.Context, string, int, int) ([]entity.Worker, error)
}

type workerReferenceCounter interface {
	CountWhere(context.Context, []domainrepo.Filter) (int64, error)
}

type workerContractLookup interface {
	FindByWorkerID(context.Context, string, int, int) ([]entity.EmploymentContract, error)
}

type workerAssignmentLookup interface {
	FindByID(context.Context, string) (*entity.WorkerAssignment, error)
}

type workerPositionLookup interface {
	FindByID(context.Context, string) (*entity.Position, error)
}

type workerValidator struct {
	workers          workerNumberReader
	contracts        workerReferenceCounter
	contractLookup   workerContractLookup
	assignments      workerReferenceCounter
	assignmentLookup workerAssignmentLookup
	positions        workerPositionLookup
	onboarding       workerReferenceCounter
	owner            WorkerReferenceVerifier
}

// WorkerValidators mendaftarkan aturan worker preboarding tanpa mengambil alih identitas IAM.
func WorkerValidators(repos *repository.Registry, owner WorkerReferenceVerifier) (map[string]MutationValidator, error) {
	if repos == nil || operation.RequireDependency(repos.Worker) != nil ||
		operation.RequireDependency(repos.EmploymentContract) != nil ||
		operation.RequireDependency(repos.WorkerAssignment) != nil ||
		operation.RequireDependency(repos.Position) != nil ||
		operation.RequireDependency(repos.OnboardingPlan) != nil {
		return nil, hrerrors.ErrUnavailable
	}
	return map[string]MutationValidator{
		"hr_workers": &workerValidator{
			workers: repos.Worker, contracts: repos.EmploymentContract, contractLookup: repos.EmploymentContract,
			assignments: repos.WorkerAssignment, assignmentLookup: repos.WorkerAssignment,
			positions: repos.Position, onboarding: repos.OnboardingPlan, owner: owner,
		},
	}, nil
}

// ValidateMutation membatasi CRUD generik pada pembuatan dan koreksi tanggal preboarding.
func (v *workerValidator) ValidateMutation(ctx context.Context, table string, before, after entity.Validatable) error {
	if v == nil || table != "hr_workers" || operation.RequireDependency(v.workers) != nil ||
		operation.RequireDependency(v.contracts) != nil || operation.RequireDependency(v.assignments) != nil ||
		operation.RequireDependency(v.onboarding) != nil {
		return hrerrors.ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	var previous *entity.Worker
	if before != nil {
		var ok bool
		previous, ok = before.(*entity.Worker)
		if !ok || previous == nil {
			return hrerrors.ErrInvalidInput
		}
	}
	if after == nil {
		if previous == nil {
			return hrerrors.ErrInvalidInput
		}
		return hrerrors.ErrWorkflowRequired
	}
	worker, ok := after.(*entity.Worker)
	if !ok || worker == nil {
		return hrerrors.ErrInvalidInput
	}
	if previous != nil && previous.Status == entity.WorkerStatusActive && worker.Status == entity.WorkerStatusActive &&
		!reflect.DeepEqual(previous.PrimaryAssignmentID, worker.PrimaryAssignmentID) {
		return v.validatePrimaryAssignment(ctx, previous, worker)
	}
	if previous != nil && previous.Status == entity.WorkerStatusPreboarding && worker.Status == entity.WorkerStatusActive {
		return v.validateActivation(ctx, previous, worker)
	}
	if previous != nil && previous.Status != entity.WorkerStatusPreboarding {
		return hrerrors.ErrWorkflowRequired
	}
	if worker.Status != entity.WorkerStatusPreboarding || worker.PrimaryAssignmentID != nil ||
		worker.IAMPrincipalReferenceID != nil || worker.TerminationDate != nil ||
		worker.TerminationReasonCode != nil {
		return hrerrors.ErrWorkflowRequired
	}
	if previous != nil && (previous.ID != worker.ID || previous.WorkerNumber != worker.WorkerNumber ||
		previous.PersonPartyReferenceID != worker.PersonPartyReferenceID ||
		previous.WorkerCategory != worker.WorkerCategory ||
		previous.PrimaryLegalEntityReferenceID != worker.PrimaryLegalEntityReferenceID ||
		!reflect.DeepEqual(previous.CreatedByPrincipalID, worker.CreatedByPrincipalID) ||
		!previous.CreatedAt.Equal(worker.CreatedAt) || previous.PrimaryAssignmentID != nil ||
		previous.IAMPrincipalReferenceID != nil) {
		return hrerrors.ErrWorkflowRequired
	}
	if previous != nil && (!reflect.DeepEqual(previous.HireDate, worker.HireDate) ||
		!reflect.DeepEqual(previous.ServiceStartDate, worker.ServiceStartDate)) {
		for _, counter := range []workerReferenceCounter{v.contracts, v.assignments, v.onboarding} {
			count, err := counter.CountWhere(ctx, []domainrepo.Filter{{Column: "worker_id", Values: []any{worker.ID}}})
			if err != nil {
				return err
			}
			if count != 0 {
				return hrerrors.ErrReferenceConflict
			}
		}
	}
	if worker.HireDate == nil && worker.ServiceStartDate != nil ||
		worker.HireDate != nil && worker.ServiceStartDate != nil &&
			entity.CompareDates(*worker.ServiceStartDate, *worker.HireDate) < 0 {
		return hrerrors.ErrReferenceConflict
	}
	if err := v.verifyWorkerReferences(ctx, worker); err != nil {
		return err
	}
	matches, err := v.workers.FindByWorkerNumber(ctx, worker.WorkerNumber, 2, 0)
	if err != nil {
		return err
	}
	for _, match := range matches {
		if match.ID != worker.ID {
			return hrerrors.ErrAlreadyExists
		}
	}
	return nil
}

func (v *workerValidator) validateActivation(ctx context.Context, previous, worker *entity.Worker) error {
	if operation.RequireDependency(v.contractLookup) != nil {
		return hrerrors.ErrUnavailable
	}
	if worker.HireDate == nil || worker.ServiceStartDate == nil || worker.PrimaryAssignmentID != nil ||
		worker.IAMPrincipalReferenceID != nil || worker.TerminationDate != nil ||
		worker.TerminationReasonCode != nil || worker.UpdatedAt.Before(previous.UpdatedAt) {
		return hrerrors.ErrReferenceConflict
	}
	unchanged := *worker
	unchanged.Status = previous.Status
	unchanged.UpdatedAt = previous.UpdatedAt
	if !reflect.DeepEqual(previous, &unchanged) {
		return hrerrors.ErrWorkflowRequired
	}
	if entity.CompareDates(*worker.HireDate, worker.UpdatedAt) > 0 ||
		entity.CompareDates(*worker.ServiceStartDate, worker.UpdatedAt) > 0 {
		return hrerrors.ErrReferenceConflict
	}
	if err := v.verifyWorkerReferences(ctx, worker); err != nil {
		return err
	}
	count, err := v.contracts.CountWhere(ctx, []domainrepo.Filter{{Column: "worker_id", Values: []any{worker.ID}}})
	if err != nil {
		return err
	}
	if count <= 0 {
		return hrerrors.ErrReferenceConflict
	}
	if count > 10000 {
		return hrerrors.ErrConflict
	}
	seen := int64(0)
	for offset := 0; seen < count; offset += domainrepo.MaxLimit {
		contracts, err := v.contractLookup.FindByWorkerID(ctx, worker.ID, domainrepo.MaxLimit, offset)
		if err != nil {
			return err
		}
		if len(contracts) == 0 || int64(len(contracts))+seen > count {
			return hrerrors.ErrConflict
		}
		for i := range contracts {
			contract := &contracts[i]
			if contract.WorkerID != worker.ID || contract.ContractStatus != "signed" ||
				contract.EmployingLegalEntityReferenceID != worker.PrimaryLegalEntityReferenceID ||
				contract.SignedAt == nil || contract.SignedAt.After(worker.UpdatedAt) ||
				contract.SignedDocumentReferenceID == nil || contract.SignedDocumentVersionReferenceID == nil ||
				len(contract.SignedContentDigest) == 0 ||
				entity.CompareDates(contract.EffectiveFrom, worker.UpdatedAt) > 0 ||
				(contract.EffectiveTo != nil && entity.CompareDates(*contract.EffectiveTo, worker.UpdatedAt) < 0) {
				continue
			}
			return nil
		}
		seen += int64(len(contracts))
	}
	return hrerrors.ErrReferenceConflict
}

func (v *workerValidator) validatePrimaryAssignment(ctx context.Context, previous, worker *entity.Worker) error {
	if operation.RequireDependency(v.assignmentLookup) != nil || operation.RequireDependency(v.positions) != nil {
		return hrerrors.ErrUnavailable
	}
	if worker.PrimaryAssignmentID == nil || worker.TerminationDate != nil ||
		worker.UpdatedAt.Before(previous.UpdatedAt) ||
		(worker.HireDate != nil && entity.CompareDates(worker.UpdatedAt, *worker.HireDate) < 0) {
		return hrerrors.ErrReferenceConflict
	}
	unchanged := *worker
	unchanged.PrimaryAssignmentID = previous.PrimaryAssignmentID
	unchanged.UpdatedAt = previous.UpdatedAt
	if !reflect.DeepEqual(previous, &unchanged) {
		return hrerrors.ErrWorkflowRequired
	}
	if previous.PrimaryAssignmentID != nil {
		prior, err := v.assignmentLookup.FindByID(ctx, *previous.PrimaryAssignmentID)
		if err != nil {
			return workerAssessmentReferenceError(err)
		}
		if prior == nil || prior.WorkerID != worker.ID || prior.AssignmentType != "primary" ||
			prior.ApprovedAt == nil || prior.ValidTo == nil || entity.CompareDates(*prior.ValidTo, worker.UpdatedAt) >= 0 {
			return hrerrors.ErrReferenceConflict
		}
	}
	assignment, err := v.assignmentLookup.FindByID(ctx, *worker.PrimaryAssignmentID)
	if err != nil {
		return workerAssessmentReferenceError(err)
	}
	if assignment == nil || assignment.WorkerID != worker.ID || assignment.AssignmentType != "primary" ||
		assignment.ApprovedByPrincipalID == nil || assignment.ApprovedAt == nil ||
		assignment.EmploymentStatusSnapshot == nil || *assignment.EmploymentStatusSnapshot != string(entity.WorkerStatusActive) ||
		worker.UpdatedAt.Before(*assignment.ApprovedAt) ||
		!assignmentContainsDate(assignment, worker.UpdatedAt) {
		return hrerrors.ErrReferenceConflict
	}
	position, err := v.positions.FindByID(ctx, assignment.PositionID)
	if err != nil {
		return workerAssessmentReferenceError(err)
	}
	if position == nil || position.PositionStatus != "active" ||
		position.LegalEntityReferenceID != worker.PrimaryLegalEntityReferenceID ||
		entity.CompareDates(worker.UpdatedAt, position.ValidFrom) < 0 ||
		(position.ValidTo != nil && entity.CompareDates(worker.UpdatedAt, *position.ValidTo) > 0) {
		return hrerrors.ErrReferenceConflict
	}
	return v.verifyWorkerReferences(ctx, worker)
}

func (v *workerValidator) verifyWorkerReferences(ctx context.Context, worker *entity.Worker) error {
	if operation.RequireDependency(v.owner) != nil {
		return hrerrors.ErrUnavailable
	}
	if err := v.owner.VerifyWorkerReferences(ctx, WorkerReferenceCheck{
		PersonPartyID: worker.PersonPartyReferenceID,
		LegalEntityID: worker.PrimaryLegalEntityReferenceID,
	}); err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if errors.Is(err, hrerrors.ErrReferenceConflict) {
			return hrerrors.ErrReferenceConflict
		}
		return hrerrors.ErrUnavailable
	}
	return nil
}
