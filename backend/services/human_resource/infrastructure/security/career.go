package security

import (
	"context"
	"sort"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository"
)

type careerPlanReader interface {
	FindByIDForUpdate(context.Context, string) (*entity.DevelopmentPlan, error)
	SelectWhere(context.Context, []domainrepo.Filter, int, int) ([]entity.DevelopmentPlan, error)
}

type careerActionReader interface {
	SelectWhere(context.Context, []domainrepo.Filter, int, int) ([]entity.DevelopmentAction, error)
	CountWhere(context.Context, []domainrepo.Filter) (int64, error)
}

type careerCourseReader interface {
	FindByIDForUpdate(context.Context, string) (*entity.LearningCourse, error)
}

type developmentPlanValidator struct {
	workers lockedWorkerReader
	jobs    lockedJobReader
	plans   careerPlanReader
	actions careerActionReader
}

type developmentActionValidator struct {
	plans        careerPlanReader
	courses      careerCourseReader
	competencies lockedCompetencyReader
	actions      careerActionReader
}

// CareerValidators memasang aturan rencana dan tindakan pengembangan HRM.
func CareerValidators(repos *repository.Registry) (map[string]MutationValidator, error) {
	if repos == nil {
		return nil, hrerrors.ErrUnavailable
	}
	for _, dependency := range []any{repos.Worker, repos.Job, repos.DevelopmentPlan, repos.DevelopmentAction, repos.LearningCourse, repos.Competency} {
		if operation.RequireDependency(dependency) != nil {
			return nil, hrerrors.ErrUnavailable
		}
	}
	return map[string]MutationValidator{
		"hr_development_plans": &developmentPlanValidator{
			workers: repos.Worker, jobs: repos.Job, plans: repos.DevelopmentPlan, actions: repos.DevelopmentAction,
		},
		"hr_development_actions": &developmentActionValidator{
			plans: repos.DevelopmentPlan, courses: repos.LearningCourse,
			competencies: repos.Competency, actions: repos.DevelopmentAction,
		},
	}, nil
}

// ValidateMutation menjaga versi rencana dan pekerja yang menjadi pihaknya.
func (v *developmentPlanValidator) ValidateMutation(ctx context.Context, table string, before, after entity.Validatable) error {
	if v == nil || table != "hr_development_plans" ||
		operation.RequireDependency(v.workers) != nil || operation.RequireDependency(v.jobs) != nil ||
		operation.RequireDependency(v.plans) != nil || operation.RequireDependency(v.actions) != nil {
		return hrerrors.ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	var previous *entity.DevelopmentPlan
	if before != nil {
		var ok bool
		previous, ok = before.(*entity.DevelopmentPlan)
		if !ok || previous == nil {
			return hrerrors.ErrInvalidInput
		}
	}
	if after == nil {
		if previous == nil {
			return hrerrors.ErrInvalidInput
		}
		count, err := v.actions.CountWhere(ctx, []domainrepo.Filter{{Column: "development_plan_id", Values: []any{previous.ID}}})
		if err != nil {
			return err
		}
		if count > 0 {
			return hrerrors.ErrReferenceConflict
		}
		return nil
	}
	plan, ok := after.(*entity.DevelopmentPlan)
	if !ok || plan == nil || (previous != nil && (previous.ID != plan.ID || previous.WorkerID != plan.WorkerID)) {
		return hrerrors.ErrInvalidInput
	}
	if err := v.lockPeopleAndJob(ctx, plan); err != nil {
		return err
	}
	matches, err := v.plans.SelectWhere(ctx, []domainrepo.Filter{
		{Column: "worker_id", Values: []any{plan.WorkerID}},
		{Column: "plan_code", Values: []any{plan.PlanCode}},
		{Column: "plan_version", Values: []any{plan.PlanVersion}},
	}, 2, 0)
	if err != nil {
		return err
	}
	for _, existing := range matches {
		if existing.ID != plan.ID {
			return hrerrors.ErrAlreadyExists
		}
	}
	return nil
}

func (v *developmentPlanValidator) lockPeopleAndJob(ctx context.Context, plan *entity.DevelopmentPlan) error {
	ids := []string{plan.WorkerID}
	if plan.MentorWorkerID != nil {
		ids = append(ids, *plan.MentorWorkerID)
	}
	sort.Strings(ids)
	for _, id := range ids {
		worker, err := v.workers.FindByIDForUpdate(ctx, id)
		if err != nil {
			return err
		}
		if worker == nil || (worker.Status != entity.WorkerStatusActive && worker.Status != entity.WorkerStatusPreboarding) {
			return hrerrors.ErrReferenceConflict
		}
	}
	if plan.TargetJobID != nil {
		job, err := v.jobs.FindByIDForUpdate(ctx, *plan.TargetJobID)
		if err != nil {
			return err
		}
		if job == nil || !job.IsActive || entity.CompareDates(job.ValidFrom, plan.TargetDate) > 0 ||
			(job.ValidTo != nil && entity.CompareDates(*job.ValidTo, plan.TargetDate) < 0) {
			return hrerrors.ErrReferenceConflict
		}
	}
	return nil
}

// ValidateMutation mengunci rencana sebelum tindakan dan semua referensi katalog.
func (v *developmentActionValidator) ValidateMutation(ctx context.Context, table string, before, after entity.Validatable) error {
	if v == nil || table != "hr_development_actions" ||
		operation.RequireDependency(v.plans) != nil || operation.RequireDependency(v.courses) != nil ||
		operation.RequireDependency(v.competencies) != nil || operation.RequireDependency(v.actions) != nil {
		return hrerrors.ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	var previous *entity.DevelopmentAction
	if before != nil {
		var ok bool
		previous, ok = before.(*entity.DevelopmentAction)
		if !ok || previous == nil {
			return hrerrors.ErrInvalidInput
		}
	}
	if after == nil && previous == nil {
		return hrerrors.ErrInvalidInput
	}
	action := previous
	if after != nil {
		var ok bool
		action, ok = after.(*entity.DevelopmentAction)
		if !ok || action == nil || (previous != nil && (previous.ID != action.ID || previous.DevelopmentPlanID != action.DevelopmentPlanID)) {
			return hrerrors.ErrInvalidInput
		}
		if action.EvidenceDocumentVersionReferenceID != nil &&
			(previous == nil || previous.EvidenceDocumentVersionReferenceID == nil ||
				*previous.EvidenceDocumentVersionReferenceID != *action.EvidenceDocumentVersionReferenceID) {
			return hrerrors.ErrUnavailable
		}
	}
	plan, err := v.plans.FindByIDForUpdate(ctx, action.DevelopmentPlanID)
	if err != nil {
		return err
	}
	if plan == nil || plan.Status != "draft" {
		return hrerrors.ErrReferenceConflict
	}
	if after == nil {
		return nil
	}
	if (action.ActionType == "course" && action.CourseID == nil) ||
		(action.TargetLevel != nil && action.CompetencyID == nil) {
		return hrerrors.ErrInvalidInput
	}
	if entity.CompareDates(action.DueDate, plan.StartDate) < 0 || entity.CompareDates(action.DueDate, plan.TargetDate) > 0 {
		return hrerrors.ErrInvalidInput
	}
	if action.CourseID != nil {
		course, err := v.courses.FindByIDForUpdate(ctx, *action.CourseID)
		if err != nil {
			return err
		}
		if course == nil || !course.IsActive {
			return hrerrors.ErrReferenceConflict
		}
	}
	if action.CompetencyID != nil {
		competency, err := v.competencies.FindByIDForUpdate(ctx, *action.CompetencyID)
		if err != nil {
			return err
		}
		if competency == nil || !competency.IsActive {
			return hrerrors.ErrReferenceConflict
		}
	}
	matches, err := v.actions.SelectWhere(ctx, []domainrepo.Filter{
		{Column: "development_plan_id", Values: []any{action.DevelopmentPlanID}},
		{Column: "action_number", Values: []any{action.ActionNumber}},
	}, 2, 0)
	if err != nil {
		return err
	}
	for _, existing := range matches {
		if existing.ID != action.ID {
			return hrerrors.ErrAlreadyExists
		}
	}
	return nil
}
