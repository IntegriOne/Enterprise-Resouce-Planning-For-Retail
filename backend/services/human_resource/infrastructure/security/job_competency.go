package security

import (
	"context"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

type lockedJobReader interface {
	FindByIDForUpdate(context.Context, string) (*entity.Job, error)
}

type lockedCompetencyReader interface {
	FindByIDForUpdate(context.Context, string) (*entity.Competency, error)
}

type jobRequirementReader interface {
	SelectWhere(context.Context, []domainrepo.Filter, int, int) ([]entity.JobCompetency, error)
}

type jobCompetencyValidator struct {
	jobs         lockedJobReader
	competencies lockedCompetencyReader
	requirements jobRequirementReader
}

// ValidateMutation mengunci kedua katalog induk dan menolak kebutuhan kompetensi yang bertabrakan.
func (v *jobCompetencyValidator) ValidateMutation(ctx context.Context, table string, before, after entity.Validatable) error {
	if v == nil || table != "hr_job_competencies" ||
		operation.RequireDependency(v.jobs) != nil ||
		operation.RequireDependency(v.competencies) != nil ||
		operation.RequireDependency(v.requirements) != nil {
		return hrerrors.ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	var current *entity.JobCompetency
	if before != nil {
		var ok bool
		current, ok = before.(*entity.JobCompetency)
		if !ok || current == nil {
			return hrerrors.ErrInvalidInput
		}
	}
	if after == nil && current == nil {
		return hrerrors.ErrInvalidInput
	}
	if after == nil {
		return v.lockParents(ctx, current, false)
	}
	requirement, ok := after.(*entity.JobCompetency)
	if !ok || requirement == nil || (current != nil && current.ID != requirement.ID) {
		return hrerrors.ErrInvalidInput
	}
	if err := v.lockParents(ctx, requirement, true); err != nil {
		return err
	}
	filters := []domainrepo.Filter{
		{Column: "job_id", Values: []any{requirement.JobID}},
		{Column: "competency_id", Values: []any{requirement.CompetencyID}},
	}
	for offset := 0; ; offset += domainrepo.MaxLimit {
		matches, err := v.requirements.SelectWhere(ctx, filters, domainrepo.MaxLimit, offset)
		if err != nil {
			return err
		}
		for _, existing := range matches {
			if existing.ID == requirement.ID {
				continue
			}
			if intervalsOverlap(existing.EffectiveFrom, existing.EffectiveTo, requirement.EffectiveFrom, requirement.EffectiveTo) {
				return hrerrors.ErrAlreadyExists
			}
		}
		if len(matches) < domainrepo.MaxLimit {
			return nil
		}
	}
}

// lockParents memakai urutan job lalu kompetensi pada seluruh mutasi child.
func (v *jobCompetencyValidator) lockParents(ctx context.Context, requirement *entity.JobCompetency, requireActive bool) error {
	job, err := v.jobs.FindByIDForUpdate(ctx, requirement.JobID)
	if err != nil {
		return err
	}
	if job == nil {
		return hrerrors.ErrReferenceConflict
	}
	competency, err := v.competencies.FindByIDForUpdate(ctx, requirement.CompetencyID)
	if err != nil {
		return err
	}
	if competency == nil {
		return hrerrors.ErrReferenceConflict
	}
	if !requireActive {
		return nil
	}
	if !job.IsActive || !competency.IsActive || requirement.EffectiveFrom.Before(job.ValidFrom) ||
		(job.ValidTo != nil && (requirement.EffectiveTo == nil || requirement.EffectiveTo.After(*job.ValidTo))) {
		return hrerrors.ErrReferenceConflict
	}
	return nil
}

func intervalsOverlap(firstStart time.Time, firstEnd *time.Time, secondStart time.Time, secondEnd *time.Time) bool {
	return (firstEnd == nil || !firstEnd.Before(secondStart)) &&
		(secondEnd == nil || !secondEnd.Before(firstStart))
}
