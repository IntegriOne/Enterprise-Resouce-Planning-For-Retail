package security

import (
	"context"
	"errors"
	"testing"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
)

type competencyReaderStub struct {
	matches []entity.Competency
	filters []domainrepo.Filter
}

func (s *competencyReaderStub) SelectWhere(_ context.Context, filters []domainrepo.Filter, _, _ int) ([]entity.Competency, error) {
	s.filters = filters
	return s.matches, nil
}

type competencyCounterStub struct {
	count int64
	err   error
	calls int
}

func (s *competencyCounterStub) CountWhere(_ context.Context, filters []domainrepo.Filter) (int64, error) {
	s.calls++
	if len(filters) != 1 || filters[0].Column != "competency_id" || len(filters[0].Values) != 1 {
		return 0, hrerrors.ErrInvalidInput
	}
	return s.count, s.err
}

func TestCompetencyValidatorRejectsDuplicateCodeVersion(t *testing.T) {
	t.Parallel()
	reader := &competencyReaderStub{matches: []entity.Competency{{ID: "other"}}}
	validator := &competencyValidator{competencies: reader, references: competencyCounters()}
	value := &entity.Competency{ID: "competency", CompetencyCode: "LEAD", CompetencyVersion: 2}
	if err := validator.ValidateMutation(t.Context(), "hr_competencies", nil, value); !errors.Is(err, hrerrors.ErrAlreadyExists) {
		t.Fatalf("duplicate code and version was not rejected: %v", err)
	}
	if len(reader.filters) != 2 || reader.filters[0].Column != "competency_code" ||
		reader.filters[1].Column != "competency_version" || reader.filters[1].Values[0] != int32(2) {
		t.Fatalf("uniqueness filters do not target code and version: %#v", reader.filters)
	}
	reader.matches = []entity.Competency{{ID: value.ID}}
	if err := validator.ValidateMutation(t.Context(), "hr_competencies", value, value); err != nil {
		t.Fatalf("same record was rejected: %v", err)
	}
}

func TestCompetencyValidatorProtectsReferencedRubricAndDeletion(t *testing.T) {
	t.Parallel()
	for index := 0; index < 4; index++ {
		t.Run([]string{"job_requirement", "worker_assessment", "development_action", "course_requirement"}[index], func(t *testing.T) {
			reader := &competencyReaderStub{}
			counters := competencyCounters()
			counters[index].(*competencyCounterStub).count = 1
			validator := &competencyValidator{competencies: reader, references: counters}
			previous := &entity.Competency{ID: "competency", CompetencyCode: "LEAD", CompetencyVersion: 1, ProficiencyScale: entity.JSON(`{"1":"basic"}`), IsActive: true}
			next := *previous
			next.ProficiencyScale = entity.JSON(`{"1":"advanced"}`)
			if err := validator.ValidateMutation(t.Context(), "hr_competencies", previous, &next); !errors.Is(err, hrerrors.ErrReferenceConflict) {
				t.Fatalf("referenced rubric changed: %v", err)
			}
			next = *previous
			next.IsActive = false
			if err := validator.ValidateMutation(t.Context(), "hr_competencies", previous, &next); !errors.Is(err, hrerrors.ErrReferenceConflict) {
				t.Fatalf("referenced competency deactivated: %v", err)
			}
			if err := validator.ValidateMutation(t.Context(), "hr_competencies", previous, nil); !errors.Is(err, hrerrors.ErrReferenceConflict) {
				t.Fatalf("referenced competency deleted: %v", err)
			}
		})
	}
}

func TestCompetencyValidatorAllowsUnreferencedChangesAndPropagatesDependencyFailure(t *testing.T) {
	t.Parallel()
	reader := &competencyReaderStub{}
	counters := competencyCounters()
	validator := &competencyValidator{competencies: reader, references: counters}
	previous := &entity.Competency{ID: "competency", CompetencyCode: "LEAD", CompetencyVersion: 1, ProficiencyScale: entity.JSON(`{}`), IsActive: true}
	next := *previous
	next.CompetencyName = "Leadership"
	if err := validator.ValidateMutation(t.Context(), "hr_competencies", previous, &next); err != nil {
		t.Fatalf("metadata change was rejected: %v", err)
	}
	for _, counter := range counters {
		if counter.(*competencyCounterStub).calls != 0 {
			t.Fatal("metadata update queried references unnecessarily")
		}
	}
	next.ProficiencyScale = entity.JSON(`{"1":"new"}`)
	if err := validator.ValidateMutation(t.Context(), "hr_competencies", previous, &next); err != nil {
		t.Fatalf("unreferenced rubric change was rejected: %v", err)
	}
	counters[2].(*competencyCounterStub).err = hrerrors.ErrUnavailable
	if err := validator.ValidateMutation(t.Context(), "hr_competencies", previous, nil); !errors.Is(err, hrerrors.ErrUnavailable) {
		t.Fatalf("dependency error was not propagated: %v", err)
	}
}

func competencyCounters() []competencyReferenceCounter {
	return []competencyReferenceCounter{
		&competencyCounterStub{}, &competencyCounterStub{},
		&competencyCounterStub{}, &competencyCounterStub{},
	}
}
