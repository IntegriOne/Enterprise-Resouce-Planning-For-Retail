package security

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
)

type lockedJobStub struct {
	job   *entity.Job
	order *[]string
}

func (s *lockedJobStub) FindByIDForUpdate(_ context.Context, _ string) (*entity.Job, error) {
	*s.order = append(*s.order, "job")
	return s.job, nil
}

type lockedCompetencyStub struct {
	competency *entity.Competency
	order      *[]string
}

func (s *lockedCompetencyStub) FindByIDForUpdate(_ context.Context, _ string) (*entity.Competency, error) {
	*s.order = append(*s.order, "competency")
	return s.competency, nil
}

type jobRequirementStub struct {
	rows []entity.JobCompetency
	err  error
}

func (s *jobRequirementStub) SelectWhere(_ context.Context, filters []domainrepo.Filter, _, _ int) ([]entity.JobCompetency, error) {
	if len(filters) != 2 || filters[0].Column != "job_id" || filters[1].Column != "competency_id" {
		return nil, hrerrors.ErrInvalidInput
	}
	return s.rows, s.err
}

func TestJobCompetencyValidatorChecksParentsAndOverlap(t *testing.T) {
	t.Parallel()
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 6, 0)
	order := []string{}
	jobs := &lockedJobStub{job: &entity.Job{ID: "job", IsActive: true, ValidFrom: start, ValidTo: &end}, order: &order}
	competencies := &lockedCompetencyStub{competency: &entity.Competency{ID: "competency", IsActive: true}, order: &order}
	requirements := &jobRequirementStub{}
	validator := &jobCompetencyValidator{jobs: jobs, competencies: competencies, requirements: requirements}
	value := &entity.JobCompetency{ID: "new", JobID: "job", CompetencyID: "competency", EffectiveFrom: start, EffectiveTo: &end}
	if err := validator.ValidateMutation(t.Context(), "hr_job_competencies", nil, value); err != nil {
		t.Fatalf("valid requirement: %v", err)
	}
	if len(order) != 2 || order[0] != "job" || order[1] != "competency" {
		t.Fatalf("parent lock order = %v", order)
	}
	requirements.rows = []entity.JobCompetency{{ID: "existing", EffectiveFrom: start.AddDate(0, 2, 0), EffectiveTo: &end}}
	if err := validator.ValidateMutation(t.Context(), "hr_job_competencies", nil, value); !errors.Is(err, hrerrors.ErrAlreadyExists) {
		t.Fatalf("overlap must fail: %v", err)
	}
	requirements.rows[0].ID = value.ID
	if err := validator.ValidateMutation(t.Context(), "hr_job_competencies", value, value); err != nil {
		t.Fatalf("same requirement must be ignored: %v", err)
	}
	value.EffectiveTo = nil
	if err := validator.ValidateMutation(t.Context(), "hr_job_competencies", nil, value); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("open interval outside job validity: %v", err)
	}
	value.EffectiveTo = &end
	competencies.competency.IsActive = false
	if err := validator.ValidateMutation(t.Context(), "hr_job_competencies", nil, value); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("inactive competency must fail: %v", err)
	}
	if err := validator.ValidateMutation(t.Context(), "hr_job_competencies", value, nil); err != nil {
		t.Fatalf("deletion of child with inactive parent must remain possible: %v", err)
	}
}
