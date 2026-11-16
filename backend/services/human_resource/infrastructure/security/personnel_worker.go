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

type workerValidator struct {
	workers     workerNumberReader
	contracts   workerReferenceCounter
	assignments workerReferenceCounter
	onboarding  workerReferenceCounter
	owner       WorkerReferenceVerifier
}

// WorkerValidators mendaftarkan aturan worker preboarding tanpa mengambil alih identitas IAM.
func WorkerValidators(repos *repository.Registry, owner WorkerReferenceVerifier) (map[string]MutationValidator, error) {
	if repos == nil || operation.RequireDependency(repos.Worker) != nil ||
		operation.RequireDependency(repos.EmploymentContract) != nil ||
		operation.RequireDependency(repos.WorkerAssignment) != nil ||
		operation.RequireDependency(repos.OnboardingPlan) != nil {
		return nil, hrerrors.ErrUnavailable
	}
	return map[string]MutationValidator{
		"hr_workers": &workerValidator{
			workers: repos.Worker, contracts: repos.EmploymentContract,
			assignments: repos.WorkerAssignment, onboarding: repos.OnboardingPlan, owner: owner,
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
		if previous.Status != entity.WorkerStatusPreboarding {
			return hrerrors.ErrWorkflowRequired
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
