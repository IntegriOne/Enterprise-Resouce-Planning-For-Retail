package security

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	"github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
)

type performanceCycleStub struct {
	matches []entity.PerformanceCycle
	cycle   *entity.PerformanceCycle
}

func (s *performanceCycleStub) FindByCycleCode(_ context.Context, _ string, _, _ int) ([]entity.PerformanceCycle, error) {
	return s.matches, nil
}

func (s *performanceCycleStub) FindByIDForUpdate(_ context.Context, _ string) (*entity.PerformanceCycle, error) {
	return s.cycle, nil
}

type cycleReferenceStub struct {
	count int64
}

func (s *cycleReferenceStub) CountWhere(_ context.Context, filters []repository.Filter) (int64, error) {
	if len(filters) != 1 || filters[0].Column != "performance_cycle_id" {
		return 0, hrerrors.ErrInvalidInput
	}
	return s.count, nil
}

func TestPerformanceCycleValidatorProtectsDraftAndReferences(t *testing.T) {
	t.Parallel()
	cycles := &performanceCycleStub{}
	goals := &cycleReferenceStub{}
	reviews := &cycleReferenceStub{}
	validator := &performanceCycleValidator{cycles: cycles, goals: goals, reviews: reviews}
	cycle := &entity.PerformanceCycle{ID: "cycle-1", CycleCode: "FY27", Status: entity.PerformanceCycleStatusDraft}
	if err := validator.ValidateMutation(t.Context(), "hr_performance_cycles", nil, cycle); err != nil {
		t.Fatalf("create draft: %v", err)
	}
	cycles.matches = []entity.PerformanceCycle{{ID: "cycle-2"}}
	if err := validator.ValidateMutation(t.Context(), "hr_performance_cycles", nil, cycle); !errors.Is(err, hrerrors.ErrAlreadyExists) {
		t.Fatalf("duplicate cycle code: %v", err)
	}
	cycles.matches = nil
	changed := *cycle
	changed.CycleName = "Revised"
	if err := validator.ValidateMutation(t.Context(), "hr_performance_cycles", cycle, &changed); err != nil {
		t.Fatalf("update unreferenced draft: %v", err)
	}
	goals.count = 1
	if err := validator.ValidateMutation(t.Context(), "hr_performance_cycles", cycle, &changed); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("update with goals: %v", err)
	}
	if err := validator.ValidateMutation(t.Context(), "hr_performance_cycles", cycle, nil); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("delete with goals: %v", err)
	}
	goals.count = 0
	reviews.count = 1
	if err := validator.ValidateMutation(t.Context(), "hr_performance_cycles", cycle, nil); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("delete with reviews: %v", err)
	}
	reviews.count = 0
	if err := validator.ValidateMutation(t.Context(), "hr_performance_cycles", cycle, nil); err != nil {
		t.Fatalf("delete unreferenced draft: %v", err)
	}
	open := *cycle
	open.Status = entity.PerformanceCycleStatusOpen
	if err := validator.ValidateMutation(t.Context(), "hr_performance_cycles", cycle, &open); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("launch without server audit: %v", err)
	}
	launchedAt := time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)
	actor := "00000000-0000-4000-8000-000000000001"
	open.LaunchedAt = &launchedAt
	open.LaunchedByPrincipalID = &actor
	if err := validator.ValidateMutation(t.Context(), "hr_performance_cycles", cycle, &open); err != nil {
		t.Fatalf("launch draft: %v", err)
	}
	open.CycleName = "Changed during launch"
	if err := validator.ValidateMutation(t.Context(), "hr_performance_cycles", cycle, &open); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("launch with business edit: %v", err)
	}
	if err := validator.ValidateMutation(t.Context(), "hr_performance_cycles", &open, nil); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("delete open cycle: %v", err)
	}
}
