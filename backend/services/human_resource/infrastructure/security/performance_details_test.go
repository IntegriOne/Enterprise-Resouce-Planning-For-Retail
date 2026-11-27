package security

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	"github.com/integrione/erp-retail/backend/shared/authcontext"
)

type goalUpdateRepoStub struct {
	byID map[string]entity.PerformanceGoalUpdate
	rows []entity.PerformanceGoalUpdate
}

func (s *goalUpdateRepoStub) FindByIDForUpdate(_ context.Context, id string) (*entity.PerformanceGoalUpdate, error) {
	if s.byID == nil {
		return nil, hrerrors.ErrNotFound
	}
	v, ok := s.byID[id]
	if !ok {
		return nil, hrerrors.ErrNotFound
	}
	copyVal := v
	return &copyVal, nil
}

func (s *goalUpdateRepoStub) SelectWhere(_ context.Context, filters []domainrepo.Filter, _, _ int) ([]entity.PerformanceGoalUpdate, error) {
	var out []entity.PerformanceGoalUpdate
	for _, row := range s.rows {
		match := true
		for _, f := range filters {
			switch f.Column {
			case "idempotency_key":
				if row.IdempotencyKey != f.Values[0] {
					match = false
				}
			case "supersedes_update_id":
				if row.SupersedesUpdateID == nil || *row.SupersedesUpdateID != f.Values[0] {
					match = false
				}
			}
		}
		if match {
			out = append(out, row)
		}
	}
	return out, nil
}

type performanceGoalLookupStub struct {
	byID map[string]entity.PerformanceGoal
}

func (s *performanceGoalLookupStub) FindByIDForUpdate(_ context.Context, id string) (*entity.PerformanceGoal, error) {
	v, ok := s.byID[id]
	if !ok {
		return nil, hrerrors.ErrNotFound
	}
	copyVal := v
	return &copyVal, nil
}

func (s *performanceGoalLookupStub) SelectWhere(context.Context, []domainrepo.Filter, int, int) ([]entity.PerformanceGoal, error) {
	return nil, nil
}

func (s *performanceGoalLookupStub) CountWhere(context.Context, []domainrepo.Filter) (int64, error) {
	return 0, nil
}

type performanceReviewLookupStub struct {
	byID map[string]entity.PerformanceReview
}

func (s *performanceReviewLookupStub) FindByIDForUpdate(_ context.Context, id string) (*entity.PerformanceReview, error) {
	v, ok := s.byID[id]
	if !ok {
		return nil, hrerrors.ErrNotFound
	}
	copyVal := v
	return &copyVal, nil
}

type performanceReviewGoalScoreStub struct {
	rows []entity.PerformanceReviewGoalScore
}

func (s *performanceReviewGoalScoreStub) SelectWhere(_ context.Context, filters []domainrepo.Filter, _, _ int) ([]entity.PerformanceReviewGoalScore, error) {
	var out []entity.PerformanceReviewGoalScore
	for _, row := range s.rows {
		match := true
		for _, f := range filters {
			switch f.Column {
			case "performance_review_id":
				if row.PerformanceReviewID != f.Values[0] {
					match = false
				}
			case "goal_id":
				if row.GoalID != f.Values[0] {
					match = false
				}
			}
		}
		if match {
			out = append(out, row)
		}
	}
	return out, nil
}

func TestPerformanceGoalUpdateValidator(t *testing.T) {
	t.Parallel()
	principal := "00000000-0000-4000-8000-000000000001"
	ctx := authcontext.WithCaller(t.Context(), authcontext.CallerContext{PrincipalID: principal, Authenticated: true})
	goal := entity.PerformanceGoal{ID: "goal-1", WorkerID: "worker-1", PerformanceCycleID: "cycle-1", GoalStatus: "approved"}
	goals := &performanceGoalLookupStub{byID: map[string]entity.PerformanceGoal{"goal-1": goal}}
	cycle := &performanceCycleStub{cycle: &entity.PerformanceCycle{ID: "cycle-1", Status: entity.PerformanceCycleStatusOpen}}
	workers := &lockedWorkerStub{worker: &entity.Worker{ID: "worker-1", Status: entity.WorkerStatusActive}, order: &[]string{}}
	updates := &goalUpdateRepoStub{byID: map[string]entity.PerformanceGoalUpdate{
		"upd-prev":  {ID: "upd-prev", GoalID: "goal-1"},
		"upd-other": {ID: "upd-other", GoalID: "goal-other"},
	}}
	validator := &performanceGoalUpdateValidator{goals: goals, cycles: cycle, workers: workers, updates: updates}

	valid := &entity.PerformanceGoalUpdate{
		ID:                                 "upd-1",
		GoalID:                             "goal-1",
		ObservedAt:                         time.Date(2027, 2, 1, 0, 0, 0, 0, time.UTC),
		ActualValue:                        entity.Decimal("50"),
		EvidenceDocumentVersionReferenceID: "doc-1",
		EvidenceDigest:                     []byte("sha256"),
		RecordedByPrincipalID:              principal,
		RecordedAt:                         time.Date(2027, 2, 1, 1, 0, 0, 0, time.UTC),
		IdempotencyKey:                     "idem-1",
	}
	if err := validator.ValidateMutation(ctx, "hr_performance_goal_updates", nil, valid); err != nil {
		t.Fatalf("create valid goal update: %v", err)
	}
	if err := validator.ValidateMutation(ctx, "hr_performance_goal_updates", valid, valid); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("update must be blocked as immutable: %v", err)
	}
	if err := validator.ValidateMutation(ctx, "hr_performance_goal_updates", valid, nil); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("delete must be blocked as immutable: %v", err)
	}
	if err := validator.ValidateMutation(t.Context(), "hr_performance_goal_updates", nil, valid); !errors.Is(err, hrerrors.ErrUnauthenticated) {
		t.Fatalf("unauthenticated caller must fail: %v", err)
	}
	wrongCaller := authcontext.WithCaller(t.Context(), authcontext.CallerContext{PrincipalID: "other-principal", Authenticated: true})
	if err := validator.ValidateMutation(wrongCaller, "hr_performance_goal_updates", nil, valid); !errors.Is(err, hrerrors.ErrForbidden) {
		t.Fatalf("wrong recorded_by principal must fail: %v", err)
	}

	// Goal belum disetujui harus ditolak.
	unapproved := goal
	unapproved.GoalStatus = "draft"
	goals.byID["goal-1"] = unapproved
	if err := validator.ValidateMutation(ctx, "hr_performance_goal_updates", nil, valid); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("draft goal update must fail: %v", err)
	}
	goals.byID["goal-1"] = goal

	// Revisi lintas goal harus ditolak, revisi pada goal yang sama diterima.
	crossGoal := "upd-other"
	valid.SupersedesUpdateID = &crossGoal
	if err := validator.ValidateMutation(ctx, "hr_performance_goal_updates", nil, valid); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("cross-goal supersede must fail: %v", err)
	}
	sameGoal := "upd-prev"
	valid.SupersedesUpdateID = &sameGoal
	if err := validator.ValidateMutation(ctx, "hr_performance_goal_updates", nil, valid); err != nil {
		t.Fatalf("same-goal supersede failed: %v", err)
	}

	// Duplikat supersedes_update_id maupun idempotency_key ditolak.
	updates.rows = []entity.PerformanceGoalUpdate{{ID: "upd-existing", GoalID: "goal-1", SupersedesUpdateID: &sameGoal, IdempotencyKey: "idem-2"}}
	if err := validator.ValidateMutation(ctx, "hr_performance_goal_updates", nil, valid); !errors.Is(err, hrerrors.ErrAlreadyExists) {
		t.Fatalf("duplicate supersedes must fail: %v", err)
	}
	valid.SupersedesUpdateID = nil
	valid.IdempotencyKey = "idem-2"
	if err := validator.ValidateMutation(ctx, "hr_performance_goal_updates", nil, valid); !errors.Is(err, hrerrors.ErrAlreadyExists) {
		t.Fatalf("duplicate idempotency_key must fail: %v", err)
	}
}

func TestPerformanceReviewGoalScoreValidator(t *testing.T) {
	t.Parallel()
	weight := entity.Decimal("50.000")
	reviews := &performanceReviewLookupStub{byID: map[string]entity.PerformanceReview{
		"rev-1": {ID: "rev-1", WorkerID: "worker-1", PerformanceCycleID: "cycle-1"},
	}}
	goals := &performanceGoalLookupStub{byID: map[string]entity.PerformanceGoal{
		"goal-1": {ID: "goal-1", WorkerID: "worker-1", PerformanceCycleID: "cycle-1", GoalStatus: "approved", WeightPercent: &weight},
	}}
	cycles := &performanceCycleStub{cycle: &entity.PerformanceCycle{ID: "cycle-1", Status: entity.PerformanceCycleStatusOpen}}
	scores := &performanceReviewGoalScoreStub{}
	validator := &performanceReviewGoalScoreValidator{reviews: reviews, goals: goals, cycles: cycles, scores: scores}

	score := &entity.PerformanceReviewGoalScore{
		ID:                                 "score-1",
		PerformanceReviewID:                "rev-1",
		GoalID:                             "goal-1",
		WorkerID:                           "worker-1",
		PerformanceCycleID:                 "cycle-1",
		Rating:                             4,
		WeightPercentSnapshot:              entity.Decimal("50"),
		RubricDocumentVersionReferenceID:   "rubric-1",
		EvidenceDocumentVersionReferenceID: "ev-1",
	}
	if err := validator.ValidateMutation(t.Context(), "hr_performance_review_goal_scores", nil, score); err != nil {
		t.Fatalf("create valid score: %v", err)
	}

	// Ketidakcocokan bobot ditolak.
	wrongWeight := *score
	wrongWeight.WeightPercentSnapshot = entity.Decimal("40")
	if err := validator.ValidateMutation(t.Context(), "hr_performance_review_goal_scores", nil, &wrongWeight); !errors.Is(err, hrerrors.ErrInvalidInput) {
		t.Fatalf("weight mismatch must fail: %v", err)
	}

	// Duplikat pasangan (review, goal) ditolak.
	scores.rows = []entity.PerformanceReviewGoalScore{{ID: "score-existing", PerformanceReviewID: "rev-1", GoalID: "goal-1"}}
	if err := validator.ValidateMutation(t.Context(), "hr_performance_review_goal_scores", nil, score); !errors.Is(err, hrerrors.ErrAlreadyExists) {
		t.Fatalf("duplicate review+goal score must fail: %v", err)
	}
	scores.rows = nil

	// Update rating pada review belum final diterima; mengubah FK review/goal ditolak.
	updated := *score
	updated.Rating = 5
	if err := validator.ValidateMutation(t.Context(), "hr_performance_review_goal_scores", score, &updated); err != nil {
		t.Fatalf("update rating before finalize: %v", err)
	}
	rebound := *score
	rebound.GoalID = "goal-2"
	if err := validator.ValidateMutation(t.Context(), "hr_performance_review_goal_scores", score, &rebound); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("rebinding goal_id must fail: %v", err)
	}

	// Review final membekukan update dan delete skor.
	finalized := time.Date(2027, 3, 1, 0, 0, 0, 0, time.UTC)
	reviews.byID["rev-1"] = entity.PerformanceReview{ID: "rev-1", WorkerID: "worker-1", PerformanceCycleID: "cycle-1", FinalizedAt: &finalized}
	if err := validator.ValidateMutation(t.Context(), "hr_performance_review_goal_scores", score, &updated); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("update score after review finalized must fail: %v", err)
	}
	if err := validator.ValidateMutation(t.Context(), "hr_performance_review_goal_scores", score, nil); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("delete score after review finalized must fail: %v", err)
	}
}
