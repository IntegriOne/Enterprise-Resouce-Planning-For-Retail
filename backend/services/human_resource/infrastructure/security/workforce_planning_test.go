package security

import (
	"context"
	"errors"
	"testing"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
)

type familyReaderStub struct {
	family  *entity.JobFamily
	matches []entity.JobFamily
	locked  bool
}

func (s *familyReaderStub) FindByFamilyCode(context.Context, string, int, int) ([]entity.JobFamily, error) {
	return s.matches, nil
}

func (s *familyReaderStub) FindByIDForUpdate(context.Context, string) (*entity.JobFamily, error) {
	s.locked = true
	return s.family, nil
}

type jobReaderStub struct {
	matches []entity.Job
	count   int64
}

func (s *jobReaderStub) FindByJobCode(context.Context, string, int, int) ([]entity.Job, error) {
	return s.matches, nil
}

func (s *jobReaderStub) CountWhere(context.Context, []domainrepo.Filter) (int64, error) {
	return s.count, nil
}

type positionCounterStub struct{ count int64 }

func (s *positionCounterStub) CountWhere(context.Context, []domainrepo.Filter) (int64, error) {
	return s.count, nil
}

func TestJobFamilyValidatorPreventsDuplicateAndReferencedDeactivation(t *testing.T) {
	t.Parallel()
	families := &familyReaderStub{matches: []entity.JobFamily{{ID: "other"}}}
	jobs := &jobReaderStub{}
	paths := &positionCounterStub{}
	validator := &jobFamilyValidator{families: families, jobs: jobs, paths: paths}
	proposed := &entity.JobFamily{ID: "family", FamilyCode: "RETAIL", IsActive: true}
	if err := validator.ValidateMutation(t.Context(), "hr_job_families", nil, proposed); !errors.Is(err, hrerrors.ErrAlreadyExists) {
		t.Fatalf("duplicate family code: %v", err)
	}
	families.matches = nil
	if err := validator.ValidateMutation(t.Context(), "hr_job_families", nil, proposed); err != nil {
		t.Fatalf("new family: %v", err)
	}
	jobs.count = 1
	proposed.IsActive = false
	if err := validator.ValidateMutation(t.Context(), "hr_job_families", &entity.JobFamily{ID: "family", IsActive: true}, proposed); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("deactivate referenced family: %v", err)
	}
	if err := validator.ValidateMutation(t.Context(), "hr_job_families", proposed, nil); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("delete referenced family: %v", err)
	}
	jobs.count = 0
	paths.count = 1
	if err := validator.ValidateMutation(t.Context(), "hr_job_families", proposed, nil); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("delete family with career path: %v", err)
	}
}

func TestJobValidatorLocksParentAndProtectsReferencedJob(t *testing.T) {
	t.Parallel()
	families := &familyReaderStub{family: &entity.JobFamily{ID: "family", IsActive: false}}
	jobs := &jobReaderStub{}
	positions := &positionCounterStub{}
	requirements := &positionCounterStub{}
	developmentPlans := &positionCounterStub{}
	aspirations := &positionCounterStub{}
	transitions := &positionCounterStub{}
	validator := &jobValidator{families: families, jobs: jobs, positions: positions, requirements: requirements,
		developmentPlans: developmentPlans, aspirations: aspirations, transitions: transitions}
	proposed := &entity.Job{ID: "job", JobFamilyID: "family", JobCode: "STORE"}
	if err := validator.ValidateMutation(t.Context(), "hr_jobs", nil, proposed); !errors.Is(err, hrerrors.ErrReferenceConflict) || !families.locked {
		t.Fatalf("inactive parent was not locked and rejected: %v", err)
	}
	families.family.IsActive = true
	jobs.matches = []entity.Job{{ID: "other", JobCode: "STORE"}}
	if err := validator.ValidateMutation(t.Context(), "hr_jobs", nil, proposed); !errors.Is(err, hrerrors.ErrAlreadyExists) {
		t.Fatalf("duplicate job code: %v", err)
	}
	jobs.matches = nil
	if err := validator.ValidateMutation(t.Context(), "hr_jobs", nil, proposed); err != nil {
		t.Fatalf("valid job: %v", err)
	}
	positions.count = 1
	if err := validator.ValidateMutation(t.Context(), "hr_jobs", proposed, nil); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("delete referenced job: %v", err)
	}
	positions.count = 0
	requirements.count = 1
	if err := validator.ValidateMutation(t.Context(), "hr_jobs", proposed, nil); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("delete job with competency requirement: %v", err)
	}
	inactive := *proposed
	inactive.IsActive = false
	proposed.IsActive = true
	if err := validator.ValidateMutation(t.Context(), "hr_jobs", proposed, &inactive); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("deactivate job with competency requirement: %v", err)
	}
	requirements.count = 0
	developmentPlans.count = 1
	if err := validator.ValidateMutation(t.Context(), "hr_jobs", proposed, &inactive); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("deactivate job used by development plan: %v", err)
	}
	developmentPlans.count = 0
	aspirations.count = 1
	if err := validator.ValidateMutation(t.Context(), "hr_jobs", proposed, &inactive); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("deactivate job used by career aspiration: %v", err)
	}
	aspirations.count = 0
	transitions.count = 1
	if err := validator.ValidateMutation(t.Context(), "hr_jobs", proposed, nil); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("delete job used by career transition: %v", err)
	}
}
