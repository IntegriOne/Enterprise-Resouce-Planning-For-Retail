package security

import (
	"context"
	"errors"
	"testing"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
)

type learningCourseStub struct {
	course  *entity.LearningCourse
	matches []entity.LearningCourse
	order   *[]string
}

func (s *learningCourseStub) SelectWhere(_ context.Context, filters []domainrepo.Filter, _, _ int) ([]entity.LearningCourse, error) {
	if len(filters) != 2 || filters[0].Column != "course_code" || filters[1].Column != "course_version" {
		return nil, hrerrors.ErrInvalidInput
	}
	return s.matches, nil
}

func (s *learningCourseStub) FindByIDForUpdate(context.Context, string) (*entity.LearningCourse, error) {
	if s.order != nil {
		*s.order = append(*s.order, "course")
	}
	return s.course, nil
}

type learningCounterStub struct {
	count int64
	err   error
}

func (s *learningCounterStub) CountWhere(_ context.Context, filters []domainrepo.Filter) (int64, error) {
	if len(filters) != 1 || filters[0].Column != "course_id" {
		return 0, hrerrors.ErrInvalidInput
	}
	return s.count, s.err
}

type learningLinkStub struct {
	rows []entity.LearningCourseCompetency
}

func (s *learningLinkStub) SelectWhere(_ context.Context, filters []domainrepo.Filter, _, _ int) ([]entity.LearningCourseCompetency, error) {
	if len(filters) != 2 || filters[0].Column != "course_id" || filters[1].Column != "competency_id" {
		return nil, hrerrors.ErrInvalidInput
	}
	return s.rows, nil
}

func TestLearningCourseValidatorKeepsVersionAndReferencesSafe(t *testing.T) {
	t.Parallel()
	courses := &learningCourseStub{}
	links := &learningCounterStub{}
	enrollments := &learningCounterStub{}
	actions := &learningCounterStub{}
	validator := &learningCourseValidator{courses: courses, competencies: links, enrollments: enrollments, developmentActions: actions}
	value := &entity.LearningCourse{ID: "course", CourseCode: "SAFETY", CourseVersion: "1", IsActive: true}
	courses.matches = []entity.LearningCourse{{ID: "other"}}
	if err := validator.ValidateMutation(t.Context(), "hr_learning_courses", nil, value); !errors.Is(err, hrerrors.ErrAlreadyExists) {
		t.Fatalf("duplicate course version: %v", err)
	}
	courses.matches = nil
	if err := validator.ValidateMutation(t.Context(), "hr_learning_courses", nil, value); err != nil {
		t.Fatalf("new internal course: %v", err)
	}
	provider := "external-party"
	withProvider := *value
	withProvider.ProviderPartyReferenceID = &provider
	if err := validator.ValidateMutation(t.Context(), "hr_learning_courses", nil, &withProvider); !errors.Is(err, hrerrors.ErrUnavailable) {
		t.Fatalf("unverified external provider: %v", err)
	}
	links.count = 1
	changed := *value
	changed.CourseVersion = "2"
	if err := validator.ValidateMutation(t.Context(), "hr_learning_courses", value, &changed); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("referenced course definition changed: %v", err)
	}
	if err := validator.ValidateMutation(t.Context(), "hr_learning_courses", value, nil); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("referenced course deleted: %v", err)
	}
	links.count = 0
	enrollments.count = 1
	if err := validator.ValidateMutation(t.Context(), "hr_learning_courses", value, nil); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("enrolled course deleted: %v", err)
	}
	enrollments.count = 0
	actions.count = 1
	if err := validator.ValidateMutation(t.Context(), "hr_learning_courses", value, nil); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("course used by development action deleted: %v", err)
	}
}

func TestLearningCourseCompetencyValidatorLocksParentsAndRejectsDuplicates(t *testing.T) {
	t.Parallel()
	order := []string{}
	courses := &learningCourseStub{course: &entity.LearningCourse{IsActive: true}, order: &order}
	competencies := &lockedCompetencyStub{competency: &entity.Competency{IsActive: true}, order: &order}
	links := &learningLinkStub{}
	validator := &learningCourseCompetencyValidator{courses: courses, competencies: competencies, links: links}
	value := &entity.LearningCourseCompetency{CourseID: "course", CompetencyID: "competency", TargetLevel: 2}
	if err := validator.ValidateMutation(t.Context(), "hr_learning_course_competencies", nil, value); err != nil {
		t.Fatalf("new course competency: %v", err)
	}
	if len(order) != 2 || order[0] != "course" || order[1] != "competency" {
		t.Fatalf("parent lock order = %v", order)
	}
	links.rows = []entity.LearningCourseCompetency{*value}
	if err := validator.ValidateMutation(t.Context(), "hr_learning_course_competencies", nil, value); !errors.Is(err, hrerrors.ErrAlreadyExists) {
		t.Fatalf("duplicate course competency: %v", err)
	}
	courses.course.IsActive = false
	if err := validator.ValidateMutation(t.Context(), "hr_learning_course_competencies", value, value); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("inactive course mutation: %v", err)
	}
	if err := validator.ValidateMutation(t.Context(), "hr_learning_course_competencies", value, nil); err != nil {
		t.Fatalf("child removal must remain possible: %v", err)
	}
}
