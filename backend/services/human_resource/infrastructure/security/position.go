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

// PositionReferenceCheck adalah fakta posisi yang harus dibuktikan oleh pemilik organisasi.
type PositionReferenceCheck struct {
	OrganizationUnitID string
	LegalEntityID      string
	WorkLocationID     *string
	ValidFrom          time.Time
	ValidTo            *time.Time
}

// PositionReferenceVerifier memeriksa status dan hubungan unit, badan hukum, dan lokasi di Foundation.
type PositionReferenceVerifier interface {
	VerifyPositionReferences(context.Context, PositionReferenceCheck) error
}

type positionJobReader interface {
	FindByIDForUpdate(context.Context, string) (*entity.Job, error)
}

type positionReader interface {
	FindByIDForUpdate(context.Context, string) (*entity.Position, error)
	FindByPositionCode(context.Context, string, int, int) ([]entity.Position, error)
	CountWhere(context.Context, []domainrepo.Filter) (int64, error)
}

type positionReferenceCounter interface {
	CountWhere(context.Context, []domainrepo.Filter) (int64, error)
}

type positionValidator struct {
	jobs         positionJobReader
	positions    positionReader
	assignments  positionReferenceCounter
	requisitions positionReferenceCounter
	succession   positionReferenceCounter
	owner        PositionReferenceVerifier
}

// PositionValidators mendaftarkan pemeriksaan posisi terencana; owner eksternal wajib diinjeksi sebelum create/update.
func PositionValidators(repos *repository.Registry, owner PositionReferenceVerifier) (map[string]MutationValidator, error) {
	if repos == nil || operation.RequireDependency(repos.Job) != nil ||
		operation.RequireDependency(repos.Position) != nil ||
		operation.RequireDependency(repos.WorkerAssignment) != nil ||
		operation.RequireDependency(repos.Requisition) != nil ||
		operation.RequireDependency(repos.SuccessionPlan) != nil {
		return nil, hrerrors.ErrUnavailable
	}
	return map[string]MutationValidator{
		"hr_positions": &positionValidator{
			jobs: repos.Job, positions: repos.Position, assignments: repos.WorkerAssignment,
			requisitions: repos.Requisition, succession: repos.SuccessionPlan, owner: owner,
		},
	}, nil
}

// ValidateMutation menjaga posisi tetap draf dan semua referensinya dapat dibuktikan.
func (v *positionValidator) ValidateMutation(ctx context.Context, table string, before, after entity.Validatable) error {
	if v == nil || table != "hr_positions" || operation.RequireDependency(v.jobs) != nil ||
		operation.RequireDependency(v.positions) != nil || operation.RequireDependency(v.assignments) != nil ||
		operation.RequireDependency(v.requisitions) != nil || operation.RequireDependency(v.succession) != nil {
		return hrerrors.ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	var previous *entity.Position
	if before != nil {
		var ok bool
		previous, ok = before.(*entity.Position)
		if !ok || previous == nil {
			return hrerrors.ErrInvalidInput
		}
		if previous.PositionStatus != "planned" {
			return hrerrors.ErrWorkflowRequired
		}
	}
	if after == nil {
		if previous == nil {
			return hrerrors.ErrInvalidInput
		}
		return v.rejectIfReferenced(ctx, previous.ID)
	}
	position, ok := after.(*entity.Position)
	if !ok || position == nil {
		return hrerrors.ErrInvalidInput
	}
	if position.PositionStatus != "planned" || (previous != nil &&
		(previous.ID != position.ID || previous.PositionCode != position.PositionCode ||
			previous.JobID != position.JobID ||
			previous.OrganizationUnitReferenceID != position.OrganizationUnitReferenceID ||
			previous.LegalEntityReferenceID != position.LegalEntityReferenceID ||
			!reflect.DeepEqual(previous.ReportsToPositionID, position.ReportsToPositionID))) {
		return hrerrors.ErrWorkflowRequired
	}
	if position.CostCenterReferenceID != nil || operation.RequireDependency(v.owner) != nil {
		return hrerrors.ErrUnavailable
	}
	job, err := v.jobs.FindByIDForUpdate(ctx, position.JobID)
	if err != nil {
		return workerAssessmentReferenceError(err)
	}
	if job == nil || !job.IsActive || entity.CompareDates(position.ValidFrom, job.ValidFrom) < 0 ||
		(job.ValidTo != nil && (position.ValidTo == nil || entity.CompareDates(*position.ValidTo, *job.ValidTo) > 0)) {
		return hrerrors.ErrReferenceConflict
	}
	if position.ReportsToPositionID != nil {
		if *position.ReportsToPositionID == position.ID {
			return hrerrors.ErrReferenceConflict
		}
		parent, err := v.positions.FindByIDForUpdate(ctx, *position.ReportsToPositionID)
		if err != nil {
			return workerAssessmentReferenceError(err)
		}
		if parent == nil || parent.LegalEntityReferenceID != position.LegalEntityReferenceID ||
			(parent.PositionStatus != "active" && parent.PositionStatus != "planned") ||
			entity.CompareDates(position.ValidFrom, parent.ValidFrom) < 0 ||
			(parent.ValidTo != nil && (position.ValidTo == nil || entity.CompareDates(*position.ValidTo, *parent.ValidTo) > 0)) {
			return hrerrors.ErrReferenceConflict
		}
	}
	if err := v.owner.VerifyPositionReferences(ctx, PositionReferenceCheck{
		OrganizationUnitID: position.OrganizationUnitReferenceID,
		LegalEntityID:      position.LegalEntityReferenceID,
		WorkLocationID:     position.WorkLocationReferenceID,
		ValidFrom:          position.ValidFrom, ValidTo: position.ValidTo,
	}); err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if errors.Is(err, hrerrors.ErrReferenceConflict) {
			return hrerrors.ErrReferenceConflict
		}
		return hrerrors.ErrUnavailable
	}
	matches, err := v.positions.FindByPositionCode(ctx, position.PositionCode, 2, 0)
	if err != nil {
		return err
	}
	for _, existing := range matches {
		if existing.ID != position.ID {
			return hrerrors.ErrAlreadyExists
		}
	}
	if previous != nil && !reflect.DeepEqual(previous, position) {
		return v.rejectIfReferenced(ctx, position.ID)
	}
	return nil
}

func (v *positionValidator) rejectIfReferenced(ctx context.Context, id string) error {
	for _, reference := range []struct {
		counter positionReferenceCounter
		column  string
	}{
		{v.positions, "reports_to_position_id"},
		{v.assignments, "position_id"},
		{v.requisitions, "position_id"},
		{v.succession, "position_id"},
	} {
		count, err := reference.counter.CountWhere(ctx, []domainrepo.Filter{{Column: reference.column, Values: []any{id}}})
		if err != nil {
			return err
		}
		if count > 0 {
			return hrerrors.ErrReferenceConflict
		}
	}
	return nil
}
