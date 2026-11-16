package security

import (
	"context"
	"reflect"
	"sort"
	"strings"

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

type careerTransitionReader interface {
	FindByIDForUpdate(context.Context, string) (*entity.CareerPathTransition, error)
}

type careerPathReader interface {
	FindByIDForUpdate(context.Context, string) (*entity.CareerPath, error)
}

type careerPathCatalogReader interface {
	SelectWhere(context.Context, []domainrepo.Filter, int, int) ([]entity.CareerPath, error)
}

type careerPathValidator struct {
	families    jobFamilyReader
	paths       careerPathCatalogReader
	transitions positionCounter
}

type workerCareerAspirationValidator struct {
	workers     lockedWorkerReader
	jobs        lockedJobReader
	transitions careerTransitionReader
	paths       careerPathReader
}

// CareerValidators memasang aturan pengembangan dan aspirasi karier HRM.
func CareerValidators(repos *repository.Registry) (map[string]MutationValidator, error) {
	if repos == nil {
		return nil, hrerrors.ErrUnavailable
	}
	for _, dependency := range []any{repos.Worker, repos.Job, repos.DevelopmentPlan, repos.DevelopmentAction, repos.LearningCourse, repos.Competency, repos.CareerPathTransition, repos.CareerPath, repos.JobFamily} {
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
		"hr_career_paths": &careerPathValidator{
			families: repos.JobFamily, paths: repos.CareerPath, transitions: repos.CareerPathTransition,
		},
		"hr_worker_career_aspirations": &workerCareerAspirationValidator{
			workers: repos.Worker, jobs: repos.Job,
			transitions: repos.CareerPathTransition, paths: repos.CareerPath,
		},
	}, nil
}

// ValidateMutation menjaga versi unik dan perubahan status yang hanya sah melalui approval.
func (v *careerPathValidator) ValidateMutation(ctx context.Context, table string, before, after entity.Validatable) error {
	if v == nil || table != "hr_career_paths" || operation.RequireDependency(v.families) != nil ||
		operation.RequireDependency(v.paths) != nil || operation.RequireDependency(v.transitions) != nil {
		return hrerrors.ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	var previous *entity.CareerPath
	if before != nil {
		var ok bool
		previous, ok = before.(*entity.CareerPath)
		if !ok || previous == nil {
			return hrerrors.ErrInvalidInput
		}
	}
	if after == nil {
		if previous == nil || previous.Status != "draft" {
			return hrerrors.ErrWorkflowRequired
		}
		count, err := v.transitions.CountWhere(ctx, []domainrepo.Filter{{Column: "career_path_id", Values: []any{previous.ID}}})
		if err != nil {
			return err
		}
		if count > 0 {
			return hrerrors.ErrReferenceConflict
		}
		return nil
	}
	path, ok := after.(*entity.CareerPath)
	if !ok || path == nil {
		return hrerrors.ErrInvalidInput
	}
	if previous == nil {
		if path.Status != "draft" || path.ApprovedByPrincipalID != nil || path.ApprovedAt != nil {
			return hrerrors.ErrWorkflowRequired
		}
	} else {
		if previous.ID != path.ID || previous.Status != "draft" ||
			previous.CreatedByPrincipalID != path.CreatedByPrincipalID || !previous.CreatedAt.Equal(path.CreatedAt) {
			return hrerrors.ErrWorkflowRequired
		}
		if path.Status == "approved" {
			if path.PathCode != previous.PathCode || path.PathVersion != previous.PathVersion ||
				path.PathName != previous.PathName || path.JobFamilyID != previous.JobFamilyID ||
				!path.ValidFrom.Equal(previous.ValidFrom) || !reflect.DeepEqual(path.ValidTo, previous.ValidTo) {
				return hrerrors.ErrWorkflowRequired
			}
		} else if path.Status != "draft" || path.ApprovedByPrincipalID != nil || path.ApprovedAt != nil {
			return hrerrors.ErrWorkflowRequired
		}
	}
	family, err := v.families.FindByIDForUpdate(ctx, path.JobFamilyID)
	if err != nil {
		return err
	}
	if family == nil || !family.IsActive {
		return hrerrors.ErrReferenceConflict
	}
	matches, err := v.paths.SelectWhere(ctx, []domainrepo.Filter{
		{Column: "path_code", Values: []any{path.PathCode}},
		{Column: "path_version", Values: []any{path.PathVersion}},
	}, 2, 0)
	if err != nil {
		return err
	}
	for _, existing := range matches {
		if existing.ID != path.ID {
			return hrerrors.ErrAlreadyExists
		}
	}
	return nil
}

// ValidateMutation menjaga pemilik aspirasi, job tujuan, dan transisi jalur yang disetujui.
func (v *workerCareerAspirationValidator) ValidateMutation(ctx context.Context, table string, before, after entity.Validatable) error {
	if v == nil || table != "hr_worker_career_aspirations" ||
		operation.RequireDependency(v.workers) != nil || operation.RequireDependency(v.jobs) != nil ||
		operation.RequireDependency(v.transitions) != nil || operation.RequireDependency(v.paths) != nil {
		return hrerrors.ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	var previous *entity.WorkerCareerAspiration
	if before != nil {
		var ok bool
		previous, ok = before.(*entity.WorkerCareerAspiration)
		if !ok || previous == nil {
			return hrerrors.ErrInvalidInput
		}
	}
	if after == nil {
		if previous == nil {
			return hrerrors.ErrInvalidInput
		}
		return nil
	}
	aspiration, ok := after.(*entity.WorkerCareerAspiration)
	if !ok || aspiration == nil {
		return hrerrors.ErrInvalidInput
	}
	if previous != nil && (previous.ID != aspiration.ID || previous.WorkerID != aspiration.WorkerID ||
		previous.TargetJobID != aspiration.TargetJobID || !previous.RecordedAt.Equal(aspiration.RecordedAt) ||
		previous.RecordedByPrincipalID != aspiration.RecordedByPrincipalID) {
		return hrerrors.ErrWorkflowRequired
	}
	if (previous == nil && aspiration.WithdrawnAt != nil) || (previous != nil && previous.WithdrawnAt != nil) {
		return hrerrors.ErrWorkflowRequired
	}
	if aspiration.PreferredLocationReferenceID != nil {
		if previous == nil || previous.PreferredLocationReferenceID == nil ||
			*previous.PreferredLocationReferenceID != *aspiration.PreferredLocationReferenceID {
			return hrerrors.ErrUnavailable
		}
	}
	worker, err := v.workers.FindByIDForUpdate(ctx, aspiration.WorkerID)
	if err != nil {
		return err
	}
	if worker == nil || worker.Status != entity.WorkerStatusActive {
		return hrerrors.ErrReferenceConflict
	}
	job, err := v.jobs.FindByIDForUpdate(ctx, aspiration.TargetJobID)
	if err != nil {
		return err
	}
	if job == nil || !job.IsActive {
		return hrerrors.ErrReferenceConflict
	}
	targetDate := aspiration.RecordedAt
	if aspiration.TargetDate != nil {
		targetDate = *aspiration.TargetDate
		if entity.CompareDates(targetDate, aspiration.RecordedAt) < 0 {
			return hrerrors.ErrInvalidInput
		}
	}
	if entity.CompareDates(job.ValidFrom, targetDate) > 0 ||
		(job.ValidTo != nil && entity.CompareDates(*job.ValidTo, targetDate) < 0) {
		return hrerrors.ErrReferenceConflict
	}
	if aspiration.CareerPathTransitionID == nil {
		return nil
	}
	transition, err := v.transitions.FindByIDForUpdate(ctx, *aspiration.CareerPathTransitionID)
	if err != nil {
		return err
	}
	if transition == nil || !strings.EqualFold(transition.ToJobID, aspiration.TargetJobID) {
		return hrerrors.ErrReferenceConflict
	}
	path, err := v.paths.FindByIDForUpdate(ctx, transition.CareerPathID)
	if err != nil {
		return err
	}
	if path == nil || path.Status != "approved" || path.ApprovedByPrincipalID == nil || path.ApprovedAt == nil ||
		path.ApprovedAt.After(aspiration.RecordedAt) ||
		entity.CompareDates(path.ValidFrom, aspiration.RecordedAt) > 0 ||
		entity.CompareDates(path.ValidFrom, targetDate) > 0 ||
		(path.ValidTo != nil && (entity.CompareDates(*path.ValidTo, aspiration.RecordedAt) < 0 ||
			entity.CompareDates(*path.ValidTo, targetDate) < 0)) {
		return hrerrors.ErrReferenceConflict
	}
	return nil
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
