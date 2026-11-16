package security

import (
	"context"
	"reflect"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository"
)

// WorkforcePlanningValidators memasang pemeriksaan awal dua resource katalog HRM.
func WorkforcePlanningValidators(repos *repository.Registry) (map[string]MutationValidator, error) {
	if repos == nil || repos.JobFamily == nil || repos.Job == nil || repos.Position == nil || repos.JobCompetency == nil ||
		repos.DevelopmentPlan == nil || repos.WorkerCareerAspiration == nil || repos.CareerPathTransition == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return map[string]MutationValidator{
		"hr_job_families": &jobFamilyValidator{families: repos.JobFamily, jobs: repos.Job},
		"hr_jobs": &jobValidator{
			families: repos.JobFamily, jobs: repos.Job, positions: repos.Position,
			requirements: repos.JobCompetency, developmentPlans: repos.DevelopmentPlan,
			aspirations: repos.WorkerCareerAspiration, transitions: repos.CareerPathTransition,
		},
	}, nil
}

type jobFamilyReader interface {
	FindByFamilyCode(context.Context, string, int, int) ([]entity.JobFamily, error)
	FindByIDForUpdate(context.Context, string) (*entity.JobFamily, error)
}

type jobReader interface {
	FindByJobCode(context.Context, string, int, int) ([]entity.Job, error)
	CountWhere(context.Context, []domainrepo.Filter) (int64, error)
}

type positionCounter interface {
	CountWhere(context.Context, []domainrepo.Filter) (int64, error)
}

type jobFamilyValidator struct {
	families jobFamilyReader
	jobs     jobReader
}

func (v *jobFamilyValidator) ValidateMutation(ctx context.Context, table string, before, after entity.Validatable) error {
	if v == nil || v.families == nil || v.jobs == nil || table != "hr_job_families" {
		return hrerrors.ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if after != nil {
		family, ok := after.(*entity.JobFamily)
		if !ok || family == nil {
			return hrerrors.ErrInvalidInput
		}
		matches, err := v.families.FindByFamilyCode(ctx, family.FamilyCode, 2, 0)
		if err != nil {
			return err
		}
		for _, existing := range matches {
			if existing.ID != family.ID {
				return hrerrors.ErrAlreadyExists
			}
		}
		if previous, ok := before.(*entity.JobFamily); ok && previous != nil && previous.IsActive && !family.IsActive {
			count, err := v.jobs.CountWhere(ctx, []domainrepo.Filter{{Column: "job_family_id", Values: []any{family.ID}}})
			if err != nil {
				return err
			}
			if count > 0 {
				return hrerrors.ErrReferenceConflict
			}
		}
		return nil
	}
	family, ok := before.(*entity.JobFamily)
	if !ok || family == nil {
		return hrerrors.ErrInvalidInput
	}
	count, err := v.jobs.CountWhere(ctx, []domainrepo.Filter{{Column: "job_family_id", Values: []any{family.ID}}})
	if err != nil {
		return err
	}
	if count > 0 {
		return hrerrors.ErrReferenceConflict
	}
	return nil
}

type jobValidator struct {
	families         jobFamilyReader
	jobs             jobReader
	positions        positionCounter
	requirements     positionCounter
	developmentPlans positionCounter
	aspirations      positionCounter
	transitions      positionCounter
}

func (v *jobValidator) ValidateMutation(ctx context.Context, table string, before, after entity.Validatable) error {
	if v == nil || v.families == nil || v.jobs == nil || v.positions == nil || v.requirements == nil ||
		v.developmentPlans == nil || v.aspirations == nil || v.transitions == nil || table != "hr_jobs" {
		return hrerrors.ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if after != nil {
		job, ok := after.(*entity.Job)
		if !ok || job == nil {
			return hrerrors.ErrInvalidInput
		}
		family, err := v.families.FindByIDForUpdate(ctx, job.JobFamilyID)
		if err != nil {
			return err
		}
		if family == nil || !family.IsActive {
			return hrerrors.ErrReferenceConflict
		}
		matches, err := v.jobs.FindByJobCode(ctx, job.JobCode, 2, 0)
		if err != nil {
			return err
		}
		for _, existing := range matches {
			if existing.ID != job.ID {
				return hrerrors.ErrAlreadyExists
			}
		}
		if previous, ok := before.(*entity.Job); ok && previous != nil &&
			((previous.IsActive && !job.IsActive) || !previous.ValidFrom.Equal(job.ValidFrom) || !reflect.DeepEqual(previous.ValidTo, job.ValidTo)) {
			count, err := v.positions.CountWhere(ctx, []domainrepo.Filter{{Column: "job_id", Values: []any{job.ID}}})
			if err != nil {
				return err
			}
			if count > 0 {
				return hrerrors.ErrReferenceConflict
			}
			if err := v.checkCareerReferences(ctx, job.ID); err != nil {
				return err
			}
			count, err = v.requirements.CountWhere(ctx, []domainrepo.Filter{{Column: "job_id", Values: []any{job.ID}}})
			if err != nil {
				return err
			}
			if count > 0 {
				return hrerrors.ErrReferenceConflict
			}
			count, err = v.developmentPlans.CountWhere(ctx, []domainrepo.Filter{{Column: "target_job_id", Values: []any{job.ID}}})
			if err != nil {
				return err
			}
			if count > 0 {
				return hrerrors.ErrReferenceConflict
			}
		}
		return nil
	}
	job, ok := before.(*entity.Job)
	if !ok || job == nil {
		return hrerrors.ErrInvalidInput
	}
	count, err := v.positions.CountWhere(ctx, []domainrepo.Filter{{Column: "job_id", Values: []any{job.ID}}})
	if err != nil {
		return err
	}
	if count > 0 {
		return hrerrors.ErrReferenceConflict
	}
	count, err = v.requirements.CountWhere(ctx, []domainrepo.Filter{{Column: "job_id", Values: []any{job.ID}}})
	if err != nil {
		return err
	}
	if count > 0 {
		return hrerrors.ErrReferenceConflict
	}
	count, err = v.developmentPlans.CountWhere(ctx, []domainrepo.Filter{{Column: "target_job_id", Values: []any{job.ID}}})
	if err != nil {
		return err
	}
	if count > 0 {
		return hrerrors.ErrReferenceConflict
	}
	return v.checkCareerReferences(ctx, job.ID)
}

func (v *jobValidator) checkCareerReferences(ctx context.Context, jobID string) error {
	for _, reference := range []struct {
		reader positionCounter
		column string
	}{
		{v.aspirations, "target_job_id"},
		{v.transitions, "from_job_id"},
		{v.transitions, "to_job_id"},
	} {
		count, err := reference.reader.CountWhere(ctx, []domainrepo.Filter{{Column: reference.column, Values: []any{jobID}}})
		if err != nil {
			return err
		}
		if count > 0 {
			return hrerrors.ErrReferenceConflict
		}
	}
	return nil
}
