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

type compensationBandReader interface {
	SelectWhere(context.Context, []domainrepo.Filter, int, int) ([]entity.CompensationBand, error)
}

type compensationRecordCounter interface {
	CountWhere(context.Context, []domainrepo.Filter) (int64, error)
}

type compensationWorkerReader interface {
	FindByIDForUpdate(context.Context, string) (*entity.Worker, error)
}

type compensationAssignmentReader interface {
	FindByIDForUpdate(context.Context, string) (*entity.WorkerAssignment, error)
}

type compensationPositionReader interface {
	FindByIDForUpdate(context.Context, string) (*entity.Position, error)
}

type compensationJobReader interface {
	FindByIDForUpdate(context.Context, string) (*entity.Job, error)
}

type compensationComponentReader interface {
	FindByIDForUpdate(context.Context, string) (*entity.PayComponent, error)
}

type compensationBandLocker interface {
	FindByIDForUpdate(context.Context, string) (*entity.CompensationBand, error)
}

type compensationBandValidator struct {
	bands   compensationBandReader
	records compensationRecordCounter
}

type compensationRecordValidator struct {
	workers     compensationWorkerReader
	assignments compensationAssignmentReader
	positions   compensationPositionReader
	jobs        compensationJobReader
	components  compensationComponentReader
	bands       compensationBandLocker
}

// CompensationValidators memasang aturan katalog band yang dimiliki HRM.
func CompensationValidators(repos *repository.Registry) (map[string]MutationValidator, error) {
	if repos == nil || operation.RequireDependency(repos.CompensationBand) != nil ||
		operation.RequireDependency(repos.CompensationRecord) != nil || operation.RequireDependency(repos.Worker) != nil ||
		operation.RequireDependency(repos.WorkerAssignment) != nil || operation.RequireDependency(repos.Position) != nil ||
		operation.RequireDependency(repos.Job) != nil || operation.RequireDependency(repos.PayComponent) != nil {
		return nil, hrerrors.ErrUnavailable
	}
	return map[string]MutationValidator{
		"hr_compensation_bands": &compensationBandValidator{
			bands: repos.CompensationBand, records: repos.CompensationRecord,
		},
		"hr_compensation_records": &compensationRecordValidator{
			workers: repos.Worker, assignments: repos.WorkerAssignment, positions: repos.Position,
			jobs: repos.Job, components: repos.PayComponent, bands: repos.CompensationBand,
		},
	}, nil
}

// ValidateMutation menjaga versi unik dan membekukan band setelah dipakai.
func (v *compensationBandValidator) ValidateMutation(ctx context.Context, table string, before, after entity.Validatable) error {
	if v == nil || table != "hr_compensation_bands" || operation.RequireDependency(v.bands) != nil ||
		operation.RequireDependency(v.records) != nil {
		return hrerrors.ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	var previous *entity.CompensationBand
	if before != nil {
		var ok bool
		previous, ok = before.(*entity.CompensationBand)
		if !ok || previous == nil {
			return hrerrors.ErrInvalidInput
		}
		if previous.ApprovalReferenceID != nil {
			return hrerrors.ErrWorkflowRequired
		}
	}
	if after == nil {
		if previous == nil {
			return hrerrors.ErrInvalidInput
		}
		return v.rejectIfUsed(ctx, previous.ID)
	}
	band, ok := after.(*entity.CompensationBand)
	if !ok || band == nil || (previous != nil && previous.ID != band.ID) {
		return hrerrors.ErrInvalidInput
	}
	if band.ApprovalReferenceID != nil {
		return hrerrors.ErrWorkflowRequired
	}
	matches, err := v.bands.SelectWhere(ctx, []domainrepo.Filter{
		{Column: "band_code", Values: []any{band.BandCode}},
		{Column: "band_version", Values: []any{band.BandVersion}},
		{Column: "jurisdiction_code", Values: []any{band.JurisdictionCode}},
	}, 2, 0)
	if err != nil {
		return err
	}
	for _, existing := range matches {
		if existing.ID != band.ID {
			return hrerrors.ErrAlreadyExists
		}
	}
	if previous != nil && !reflect.DeepEqual(previous, band) {
		return v.rejectIfUsed(ctx, band.ID)
	}
	return nil
}

func (v *compensationBandValidator) rejectIfUsed(ctx context.Context, bandID string) error {
	count, err := v.records.CountWhere(ctx, []domainrepo.Filter{{Column: "compensation_band_id", Values: []any{bandID}}})
	if err != nil {
		return err
	}
	if count != 0 {
		return hrerrors.ErrReferenceConflict
	}
	return nil
}

// ValidateMutation hanya mengizinkan proposal pada pekerja dan penempatan yang sesuai.
func (v *compensationRecordValidator) ValidateMutation(ctx context.Context, table string, before, after entity.Validatable) error {
	if v == nil || table != "hr_compensation_records" || operation.RequireDependency(v.workers) != nil ||
		operation.RequireDependency(v.assignments) != nil || operation.RequireDependency(v.positions) != nil ||
		operation.RequireDependency(v.jobs) != nil || operation.RequireDependency(v.components) != nil ||
		operation.RequireDependency(v.bands) != nil {
		return hrerrors.ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	var previous *entity.CompensationRecord
	if before != nil {
		var ok bool
		previous, ok = before.(*entity.CompensationRecord)
		if !ok || previous == nil {
			return hrerrors.ErrInvalidInput
		}
		if previous.Status != entity.CompensationStatusProposed {
			return hrerrors.ErrWorkflowRequired
		}
	}
	if after == nil {
		if previous == nil {
			return hrerrors.ErrInvalidInput
		}
		return nil
	}
	record, ok := after.(*entity.CompensationRecord)
	if !ok || record == nil || (previous != nil && (previous.ID != record.ID ||
		previous.WorkerID != record.WorkerID || previous.AssignmentID != record.AssignmentID ||
		previous.ComponentID != record.ComponentID || previous.ComponentCodeSnapshot != record.ComponentCodeSnapshot ||
		previous.CompensationType != record.CompensationType || previous.CurrencyCode != record.CurrencyCode ||
		previous.PayFrequency != record.PayFrequency || !reflect.DeepEqual(previous.CompensationBandID, record.CompensationBandID))) {
		return hrerrors.ErrInvalidInput
	}
	if record.Status != entity.CompensationStatusProposed || record.ApprovedByPrincipalID != nil ||
		record.ApprovedAt != nil || record.SourceChangeReferenceID != nil || record.AccessClass != "restricted" ||
		(record.CompensationType != "base_salary" && record.CompensationType != "allowance") {
		return hrerrors.ErrWorkflowRequired
	}
	worker, err := v.workers.FindByIDForUpdate(ctx, record.WorkerID)
	if err != nil {
		return compensationReferenceError(err)
	}
	if worker == nil || worker.Status != entity.WorkerStatusActive ||
		(worker.HireDate != nil && entity.CompareDates(record.EffectiveFrom, *worker.HireDate) < 0) ||
		(worker.TerminationDate != nil && (record.EffectiveTo == nil || entity.CompareDates(*record.EffectiveTo, *worker.TerminationDate) > 0)) {
		return hrerrors.ErrReferenceConflict
	}
	assignment, err := v.assignments.FindByIDForUpdate(ctx, record.AssignmentID)
	if err != nil {
		return compensationReferenceError(err)
	}
	if assignment == nil || assignment.WorkerID != record.WorkerID ||
		assignment.ApprovedByPrincipalID == nil || assignment.ApprovedAt == nil ||
		!compensationPeriodWithin(record.EffectiveFrom, record.EffectiveTo, assignment.ValidFrom, assignment.ValidTo) {
		return hrerrors.ErrReferenceConflict
	}
	position, err := v.positions.FindByIDForUpdate(ctx, assignment.PositionID)
	if err != nil {
		return compensationReferenceError(err)
	}
	if position == nil || position.LegalEntityReferenceID != worker.PrimaryLegalEntityReferenceID ||
		!compensationPeriodWithin(record.EffectiveFrom, record.EffectiveTo, position.ValidFrom, position.ValidTo) {
		return hrerrors.ErrReferenceConflict
	}
	job, err := v.jobs.FindByIDForUpdate(ctx, position.JobID)
	if err != nil {
		return compensationReferenceError(err)
	}
	if job == nil || !job.IsActive || !compensationPeriodWithin(record.EffectiveFrom, record.EffectiveTo, job.ValidFrom, job.ValidTo) {
		return hrerrors.ErrReferenceConflict
	}
	component, err := v.components.FindByIDForUpdate(ctx, record.ComponentID)
	if err != nil {
		return compensationReferenceError(err)
	}
	if component == nil || !component.IsActive || !isLocalFixedEarning(component) ||
		component.ComponentCode != record.ComponentCodeSnapshot || *component.CurrencyCode != record.CurrencyCode ||
		!compensationPeriodWithin(record.EffectiveFrom, record.EffectiveTo, component.EffectiveFrom, component.EffectiveTo) {
		return hrerrors.ErrReferenceConflict
	}
	if record.CompensationBandID == nil {
		return nil
	}
	band, err := v.bands.FindByIDForUpdate(ctx, *record.CompensationBandID)
	if err != nil {
		return compensationReferenceError(err)
	}
	if band == nil || job.JobLevelCode == nil || band.JobLevelCode != *job.JobLevelCode ||
		band.JurisdictionCode != component.JurisdictionCode || band.CurrencyCode != record.CurrencyCode ||
		band.PayFrequency != record.PayFrequency ||
		!compensationPeriodWithin(record.EffectiveFrom, record.EffectiveTo, band.ValidFrom, band.ValidTo) {
		return hrerrors.ErrReferenceConflict
	}
	amount, err := record.Amount.Rat()
	if err != nil {
		return err
	}
	minimum, err := band.MinimumAmount.Rat()
	if err != nil {
		return err
	}
	maximum, err := band.MaximumAmount.Rat()
	if err != nil {
		return err
	}
	if amount.Cmp(minimum) < 0 || amount.Cmp(maximum) > 0 {
		return hrerrors.ErrReferenceConflict
	}
	return nil
}

func compensationPeriodWithin(from time.Time, to *time.Time, parentFrom time.Time, parentTo *time.Time) bool {
	return entity.CompareDates(from, parentFrom) >= 0 &&
		(parentTo == nil || (to != nil && entity.CompareDates(*to, *parentTo) <= 0))
}

func compensationReferenceError(err error) error {
	if errors.Is(err, hrerrors.ErrNotFound) {
		return hrerrors.ErrReferenceConflict
	}
	return err
}
