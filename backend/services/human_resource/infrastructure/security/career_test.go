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

type careerWorkerStub struct {
	workers map[string]*entity.Worker
	order   *[]string
}

func (s *careerWorkerStub) FindByIDForUpdate(_ context.Context, id string) (*entity.Worker, error) {
	*s.order = append(*s.order, id)
	return s.workers[id], nil
}

type careerJobStub struct{ job *entity.Job }

func (s *careerJobStub) FindByIDForUpdate(context.Context, string) (*entity.Job, error) {
	return s.job, nil
}

type careerPlanStub struct {
	plan *entity.DevelopmentPlan
	rows []entity.DevelopmentPlan
}

func (s *careerPlanStub) FindByIDForUpdate(context.Context, string) (*entity.DevelopmentPlan, error) {
	return s.plan, nil
}

func (s *careerPlanStub) SelectWhere(_ context.Context, filters []domainrepo.Filter, _, _ int) ([]entity.DevelopmentPlan, error) {
	if len(filters) != 3 || filters[0].Column != "worker_id" || filters[1].Column != "plan_code" || filters[2].Column != "plan_version" {
		return nil, hrerrors.ErrInvalidInput
	}
	return s.rows, nil
}

type careerActionStub struct {
	rows  []entity.DevelopmentAction
	count int64
}

func (s *careerActionStub) SelectWhere(_ context.Context, filters []domainrepo.Filter, _, _ int) ([]entity.DevelopmentAction, error) {
	if len(filters) != 2 || filters[0].Column != "development_plan_id" || filters[1].Column != "action_number" {
		return nil, hrerrors.ErrInvalidInput
	}
	return s.rows, nil
}

func (s *careerActionStub) CountWhere(_ context.Context, filters []domainrepo.Filter) (int64, error) {
	if len(filters) != 1 || filters[0].Column != "development_plan_id" {
		return 0, hrerrors.ErrInvalidInput
	}
	return s.count, nil
}

type careerCourseStub struct{ course *entity.LearningCourse }

func (s *careerCourseStub) FindByIDForUpdate(context.Context, string) (*entity.LearningCourse, error) {
	return s.course, nil
}

type careerCompetencyStub struct{ competency *entity.Competency }

func (s *careerCompetencyStub) FindByIDForUpdate(context.Context, string) (*entity.Competency, error) {
	return s.competency, nil
}

func TestDevelopmentPlanValidatorChecksPeopleJobVersionAndChildren(t *testing.T) {
	t.Parallel()
	order := []string{}
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 6, 0)
	mentor, jobID := "a-mentor", "target-job"
	workerID := "z-worker"
	workers := &careerWorkerStub{workers: map[string]*entity.Worker{
		workerID: {Status: entity.WorkerStatusActive}, mentor: {Status: entity.WorkerStatusActive},
	}, order: &order}
	jobs := &careerJobStub{job: &entity.Job{IsActive: true, ValidFrom: start}}
	plans := &careerPlanStub{}
	actions := &careerActionStub{}
	validator := &developmentPlanValidator{workers: workers, jobs: jobs, plans: plans, actions: actions}
	plan := &entity.DevelopmentPlan{
		ID: "plan", WorkerID: workerID, MentorWorkerID: &mentor, TargetJobID: &jobID,
		PlanCode: "growth", PlanVersion: 1, StartDate: start, TargetDate: end, Status: "draft",
	}
	if err := validator.ValidateMutation(t.Context(), "hr_development_plans", nil, plan); err != nil {
		t.Fatalf("create draft plan: %v", err)
	}
	if len(order) != 2 || order[0] != mentor || order[1] != workerID {
		t.Fatalf("worker lock order = %v", order)
	}
	plans.rows = []entity.DevelopmentPlan{{ID: "other"}}
	if err := validator.ValidateMutation(t.Context(), "hr_development_plans", nil, plan); !errors.Is(err, hrerrors.ErrAlreadyExists) {
		t.Fatalf("duplicate worker/code/version: %v", err)
	}
	plans.rows = nil
	jobs.job.IsActive = false
	if err := validator.ValidateMutation(t.Context(), "hr_development_plans", nil, plan); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("inactive target job: %v", err)
	}
	jobs.job.IsActive = true
	actions.count = 1
	if err := validator.ValidateMutation(t.Context(), "hr_development_plans", plan, nil); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("delete referenced plan: %v", err)
	}
	actions.count = 0
	if err := validator.ValidateMutation(t.Context(), "hr_development_plans", plan, nil); err != nil {
		t.Fatalf("delete empty draft plan: %v", err)
	}
}

func TestDevelopmentActionValidatorChecksPlanCatalogAndDates(t *testing.T) {
	t.Parallel()
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 6, 0)
	plan := &entity.DevelopmentPlan{Status: "draft", StartDate: start, TargetDate: end}
	plans := &careerPlanStub{plan: plan}
	courses := &careerCourseStub{course: &entity.LearningCourse{IsActive: true}}
	competencies := &careerCompetencyStub{competency: &entity.Competency{IsActive: true}}
	actions := &careerActionStub{}
	validator := &developmentActionValidator{plans: plans, courses: courses, competencies: competencies, actions: actions}
	courseID, competencyID := "course", "competency"
	action := &entity.DevelopmentAction{
		ID: "action", DevelopmentPlanID: "plan", ActionNumber: 1,
		CourseID: &courseID, CompetencyID: &competencyID, ActionType: "course", DueDate: end,
	}
	if err := validator.ValidateMutation(t.Context(), "hr_development_actions", nil, action); err != nil {
		t.Fatalf("create draft action: %v", err)
	}
	actions.rows = []entity.DevelopmentAction{{ID: "other"}}
	if err := validator.ValidateMutation(t.Context(), "hr_development_actions", nil, action); !errors.Is(err, hrerrors.ErrAlreadyExists) {
		t.Fatalf("duplicate action number: %v", err)
	}
	actions.rows = nil
	action.CourseID = nil
	if err := validator.ValidateMutation(t.Context(), "hr_development_actions", nil, action); !errors.Is(err, hrerrors.ErrInvalidInput) {
		t.Fatalf("course action without course: %v", err)
	}
	action.CourseID = &courseID
	level := int16(3)
	action.TargetLevel = &level
	action.CompetencyID = nil
	if err := validator.ValidateMutation(t.Context(), "hr_development_actions", nil, action); !errors.Is(err, hrerrors.ErrInvalidInput) {
		t.Fatalf("target level without competency: %v", err)
	}
	action.TargetLevel = nil
	action.CompetencyID = &competencyID
	action.DueDate = start.AddDate(0, 0, -1)
	if err := validator.ValidateMutation(t.Context(), "hr_development_actions", nil, action); !errors.Is(err, hrerrors.ErrInvalidInput) {
		t.Fatalf("due date before plan: %v", err)
	}
	action.DueDate = end
	courses.course.IsActive = false
	if err := validator.ValidateMutation(t.Context(), "hr_development_actions", nil, action); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("inactive course: %v", err)
	}
	courses.course.IsActive = true
	evidence := "external-document"
	action.EvidenceDocumentVersionReferenceID = &evidence
	if err := validator.ValidateMutation(t.Context(), "hr_development_actions", nil, action); !errors.Is(err, hrerrors.ErrUnavailable) {
		t.Fatalf("unverified evidence: %v", err)
	}
	action.EvidenceDocumentVersionReferenceID = nil
	plan.Status = "active"
	if err := validator.ValidateMutation(t.Context(), "hr_development_actions", nil, action); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("change active plan: %v", err)
	}
	plan.Status = "draft"
	if err := validator.ValidateMutation(t.Context(), "hr_development_actions", action, nil); err != nil {
		t.Fatalf("delete draft action: %v", err)
	}
}
