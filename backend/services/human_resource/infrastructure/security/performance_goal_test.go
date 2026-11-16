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

type performanceGoalStub struct {
	matches []entity.PerformanceGoal
	parent  *entity.PerformanceGoal
	count   int64
}

func (s *performanceGoalStub) FindByIDForUpdate(_ context.Context, _ string) (*entity.PerformanceGoal, error) {
	return s.parent, nil
}

func (s *performanceGoalStub) SelectWhere(_ context.Context, filters []domainrepo.Filter, _, _ int) ([]entity.PerformanceGoal, error) {
	if len(filters) != 3 || filters[0].Column != "worker_id" || filters[1].Column != "performance_cycle_id" || filters[2].Column != "goal_code" {
		return nil, hrerrors.ErrInvalidInput
	}
	return s.matches, nil
}

func (s *performanceGoalStub) CountWhere(_ context.Context, filters []domainrepo.Filter) (int64, error) {
	if len(filters) != 1 || filters[0].Column != "parent_goal_id" {
		return 0, hrerrors.ErrInvalidInput
	}
	return s.count, nil
}

type goalReferenceStub struct {
	count int64
}

func (s *goalReferenceStub) CountWhere(_ context.Context, filters []domainrepo.Filter) (int64, error) {
	if len(filters) != 1 || filters[0].Column != "goal_id" {
		return 0, hrerrors.ErrInvalidInput
	}
	return s.count, nil
}

func TestPerformanceGoalValidatorLocksParentsAndProtectsReferences(t *testing.T) {
	t.Parallel()
	order := []string{}
	worker := &lockedWorkerStub{worker: &entity.Worker{Status: entity.WorkerStatusActive}, order: &order}
	start := time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2027, 12, 31, 0, 0, 0, 0, time.UTC)
	cycle := &performanceCycleStub{cycle: &entity.PerformanceCycle{Status: entity.PerformanceCycleStatusOpen, PeriodStart: start, PeriodEnd: end}}
	goals := &performanceGoalStub{}
	updates := &goalReferenceStub{}
	scores := &goalReferenceStub{}
	validator := &performanceGoalValidator{workers: worker, cycles: cycle, goals: goals, updates: updates, scores: scores}
	code := "G-1"
	goal := &entity.PerformanceGoal{ID: "goal-1", WorkerID: "worker-1", PerformanceCycleID: "cycle-1", GoalCode: &code, GoalStatus: "draft"}
	if err := validator.ValidateMutation(t.Context(), "hr_performance_goals", nil, goal); err != nil || len(order) != 1 {
		t.Fatalf("create goal with active worker/open cycle: %v", err)
	}
	goals.matches = []entity.PerformanceGoal{{ID: "goal-2"}}
	if err := validator.ValidateMutation(t.Context(), "hr_performance_goals", nil, goal); !errors.Is(err, hrerrors.ErrAlreadyExists) {
		t.Fatalf("duplicate goal code: %v", err)
	}
	goals.matches = nil
	cycle.cycle.Status = entity.PerformanceCycleStatusDraft
	if err := validator.ValidateMutation(t.Context(), "hr_performance_goals", nil, goal); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("goal in draft cycle: %v", err)
	}
	cycle.cycle.Status = entity.PerformanceCycleStatusOpen
	parentID := "goal-parent"
	goal.ParentGoalID = &parentID
	goals.parent = &entity.PerformanceGoal{WorkerID: "other-worker", PerformanceCycleID: "cycle-1", GoalStatus: "draft"}
	if err := validator.ValidateMutation(t.Context(), "hr_performance_goals", nil, goal); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("parent from another worker: %v", err)
	}
	goals.parent.WorkerID = goal.WorkerID
	if err := validator.ValidateMutation(t.Context(), "hr_performance_goals", nil, goal); err != nil {
		t.Fatalf("matching parent: %v", err)
	}
	goal.ParentGoalID = nil
	changed := *goal
	changed.GoalTitle = "Revised"
	updates.count = 1
	if err := validator.ValidateMutation(t.Context(), "hr_performance_goals", goal, &changed); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("update measured goal: %v", err)
	}
	updates.count = 0
	scores.count = 1
	if err := validator.ValidateMutation(t.Context(), "hr_performance_goals", goal, nil); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("delete scored goal: %v", err)
	}
	scores.count = 0
	if err := validator.ValidateMutation(t.Context(), "hr_performance_goals", goal, nil); err != nil {
		t.Fatalf("delete unused draft goal: %v", err)
	}
}
