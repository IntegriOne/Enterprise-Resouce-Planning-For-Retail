package security

import (
	"context"
	"reflect"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository"
)

type performanceCycleReader interface {
	FindByCycleCode(context.Context, string, int, int) ([]entity.PerformanceCycle, error)
	FindByIDForUpdate(context.Context, string) (*entity.PerformanceCycle, error)
}

type performanceReferenceCounter interface {
	CountWhere(context.Context, []domainrepo.Filter) (int64, error)
}

type performanceCycleValidator struct {
	cycles  performanceCycleReader
	goals   performanceReferenceCounter
	reviews performanceReferenceCounter
}

// PerformanceValidators memasang pemeriksaan draft siklus kinerja.
func PerformanceValidators(repos *repository.Registry) (map[string]MutationValidator, error) {
	if repos == nil || operation.RequireDependency(repos.PerformanceCycle) != nil ||
		operation.RequireDependency(repos.PerformanceGoal) != nil ||
		operation.RequireDependency(repos.PerformanceReview) != nil ||
		operation.RequireDependency(repos.Worker) != nil ||
		operation.RequireDependency(repos.PerformanceGoalUpdate) != nil ||
		operation.RequireDependency(repos.PerformanceReviewGoalScore) != nil ||
		operation.RequireDependency(repos.TalentAssessment) != nil {
		return nil, hrerrors.ErrUnavailable
	}
	return map[string]MutationValidator{
		"hr_performance_cycles": &performanceCycleValidator{
			cycles: repos.PerformanceCycle, goals: repos.PerformanceGoal, reviews: repos.PerformanceReview,
		},
		"hr_performance_goals": &performanceGoalValidator{
			workers: repos.Worker, cycles: repos.PerformanceCycle, goals: repos.PerformanceGoal,
			updates: repos.PerformanceGoalUpdate, scores: repos.PerformanceReviewGoalScore,
		},
		"hr_performance_reviews": &performanceReviewValidator{
			workers: repos.Worker, cycles: repos.PerformanceCycle, reviews: repos.PerformanceReview,
			scores: repos.PerformanceReviewGoalScore, assessments: repos.TalentAssessment,
		},
		"hr_performance_goal_updates": &performanceGoalUpdateValidator{
			goals: repos.PerformanceGoal, cycles: repos.PerformanceCycle,
			workers: repos.Worker, updates: repos.PerformanceGoalUpdate,
		},
		"hr_performance_review_goal_scores": &performanceReviewGoalScoreValidator{
			reviews: repos.PerformanceReview, goals: repos.PerformanceGoal,
			cycles: repos.PerformanceCycle, scores: repos.PerformanceReviewGoalScore,
		},
	}, nil
}

// ValidateMutation menjaga kode unik dan membekukan siklus yang telah dipakai.
func (v *performanceCycleValidator) ValidateMutation(ctx context.Context, table string, before, after entity.Validatable) error {
	if v == nil || table != "hr_performance_cycles" || operation.RequireDependency(v.cycles) != nil ||
		operation.RequireDependency(v.goals) != nil || operation.RequireDependency(v.reviews) != nil {
		return hrerrors.ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	var previous *entity.PerformanceCycle
	if before != nil {
		var ok bool
		previous, ok = before.(*entity.PerformanceCycle)
		if !ok || previous == nil {
			return hrerrors.ErrInvalidInput
		}
		if previous.Status != entity.PerformanceCycleStatusDraft && previous.Status != entity.PerformanceCycleStatusOpen {
			return hrerrors.ErrWorkflowRequired
		}
	}
	if after == nil {
		if previous == nil {
			return hrerrors.ErrInvalidInput
		}
		if previous.Status != entity.PerformanceCycleStatusDraft {
			return hrerrors.ErrWorkflowRequired
		}
		return v.rejectIfReferenced(ctx, previous.ID)
	}
	cycle, ok := after.(*entity.PerformanceCycle)
	if !ok || cycle == nil || (previous != nil && previous.ID != cycle.ID) {
		return hrerrors.ErrInvalidInput
	}
	launching := previous != nil && previous.Status == entity.PerformanceCycleStatusDraft && cycle.Status == entity.PerformanceCycleStatusOpen
	closing := previous != nil && previous.Status == entity.PerformanceCycleStatusOpen && cycle.Status == entity.PerformanceCycleStatusFinalized
	if launching {
		if previous.LaunchedAt != nil || previous.LaunchedByPrincipalID != nil ||
			cycle.LaunchedAt == nil || cycle.LaunchedByPrincipalID == nil {
			return hrerrors.ErrWorkflowRequired
		}
		unchanged := *cycle
		unchanged.Status = previous.Status
		unchanged.LaunchedAt = previous.LaunchedAt
		unchanged.LaunchedByPrincipalID = previous.LaunchedByPrincipalID
		if !reflect.DeepEqual(previous, &unchanged) {
			return hrerrors.ErrWorkflowRequired
		}
	} else if closing {
		unchanged := *cycle
		unchanged.Status = previous.Status
		if !reflect.DeepEqual(previous, &unchanged) {
			return hrerrors.ErrWorkflowRequired
		}
	} else if (previous != nil && previous.Status != entity.PerformanceCycleStatusDraft) ||
		cycle.Status != entity.PerformanceCycleStatusDraft || cycle.LaunchedAt != nil || cycle.LaunchedByPrincipalID != nil {
		return hrerrors.ErrWorkflowRequired
	}
	matches, err := v.cycles.FindByCycleCode(ctx, cycle.CycleCode, 2, 0)
	if err != nil {
		return err
	}
	for _, existing := range matches {
		if existing.ID != cycle.ID {
			return hrerrors.ErrAlreadyExists
		}
	}
	if previous != nil && !launching && !closing && !reflect.DeepEqual(previous, cycle) {
		return v.rejectIfReferenced(ctx, cycle.ID)
	}
	return nil
}

func (v *performanceCycleValidator) rejectIfReferenced(ctx context.Context, id string) error {
	filters := []domainrepo.Filter{{Column: "performance_cycle_id", Values: []any{id}}}
	count, err := v.goals.CountWhere(ctx, filters)
	if err != nil {
		return err
	}
	if count > 0 {
		return hrerrors.ErrReferenceConflict
	}
	count, err = v.reviews.CountWhere(ctx, filters)
	if err != nil {
		return err
	}
	if count > 0 {
		return hrerrors.ErrReferenceConflict
	}
	return nil
}
