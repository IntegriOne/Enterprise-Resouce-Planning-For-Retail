package security

import (
	"bytes"
	"context"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository"
)

// CompetencyValidators memasang aturan katalog kompetensi yang seluruh referensinya dimiliki HRM.
func CompetencyValidators(repos *repository.Registry) (map[string]MutationValidator, error) {
	if repos == nil {
		return nil, hrerrors.ErrUnavailable
	}
	for _, dependency := range []any{
		repos.Competency, repos.JobCompetency, repos.WorkerCompetencyAssessment,
		repos.DevelopmentAction, repos.LearningCourseCompetency,
	} {
		if operation.RequireDependency(dependency) != nil {
			return nil, hrerrors.ErrUnavailable
		}
	}
	return map[string]MutationValidator{
		"hr_competencies": &competencyValidator{
			competencies: repos.Competency,
			references: []competencyReferenceCounter{
				repos.JobCompetency,
				repos.WorkerCompetencyAssessment,
				repos.DevelopmentAction,
				repos.LearningCourseCompetency,
			},
		},
	}, nil
}

type competencyReader interface {
	SelectWhere(context.Context, []domainrepo.Filter, int, int) ([]entity.Competency, error)
}

type competencyReferenceCounter interface {
	CountWhere(context.Context, []domainrepo.Filter) (int64, error)
}

type competencyValidator struct {
	competencies competencyReader
	references   []competencyReferenceCounter
}

func (v *competencyValidator) ValidateMutation(ctx context.Context, table string, before, after entity.Validatable) error {
	if v == nil || operation.RequireDependency(v.competencies) != nil ||
		len(v.references) != 4 || table != "hr_competencies" {
		return hrerrors.ErrUnavailable
	}
	for _, reference := range v.references {
		if operation.RequireDependency(reference) != nil {
			return hrerrors.ErrUnavailable
		}
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if after != nil {
		competency, ok := after.(*entity.Competency)
		if !ok || competency == nil {
			return hrerrors.ErrInvalidInput
		}
		matches, err := v.competencies.SelectWhere(ctx, []domainrepo.Filter{
			{Column: "competency_code", Values: []any{competency.CompetencyCode}},
			{Column: "competency_version", Values: []any{competency.CompetencyVersion}},
		}, 2, 0)
		if err != nil {
			return err
		}
		for _, existing := range matches {
			if existing.ID != competency.ID {
				return hrerrors.ErrAlreadyExists
			}
		}
		if before == nil {
			return nil
		}
		previous, ok := before.(*entity.Competency)
		if !ok || previous == nil || previous.ID != competency.ID {
			return hrerrors.ErrInvalidInput
		}
		changedRubric := previous.CompetencyCode != competency.CompetencyCode ||
			previous.CompetencyVersion != competency.CompetencyVersion ||
			!bytes.Equal(previous.ProficiencyScale, competency.ProficiencyScale)
		deactivated := previous.IsActive && !competency.IsActive
		if !changedRubric && !deactivated {
			return nil
		}
		return v.rejectIfReferenced(ctx, competency.ID)
	}
	previous, ok := before.(*entity.Competency)
	if !ok || previous == nil {
		return hrerrors.ErrInvalidInput
	}
	return v.rejectIfReferenced(ctx, previous.ID)
}

func (v *competencyValidator) rejectIfReferenced(ctx context.Context, competencyID string) error {
	filters := []domainrepo.Filter{{Column: "competency_id", Values: []any{competencyID}}}
	for _, reference := range v.references {
		count, err := reference.CountWhere(ctx, filters)
		if err != nil {
			return err
		}
		if count > 0 {
			return hrerrors.ErrReferenceConflict
		}
	}
	return nil
}
