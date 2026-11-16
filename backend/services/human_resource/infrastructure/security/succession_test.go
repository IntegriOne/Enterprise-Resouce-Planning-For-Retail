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

type successionPositionStub struct {
	position *entity.Position
	err      error
}

func (s *successionPositionStub) FindByIDForUpdate(context.Context, string) (*entity.Position, error) {
	return s.position, s.err
}

type successionPlanStub struct {
	items []entity.SuccessionPlan
	err   error
}

func (s *successionPlanStub) SelectWhere(context.Context, []domainrepo.Filter, int, int) ([]entity.SuccessionPlan, error) {
	return s.items, s.err
}

type successionCandidateStub struct {
	count int64
	err   error
}

func (s *successionCandidateStub) CountWhere(context.Context, []domainrepo.Filter) (int64, error) {
	return s.count, s.err
}

func TestSuccessionPlanValidatorDraftLifecycle(t *testing.T) {
	t.Parallel()
	date := time.Date(2026, time.November, 17, 0, 0, 0, 0, time.UTC)
	positions := &successionPositionStub{position: &entity.Position{
		ID: "position-1", PositionStatus: "active", ValidFrom: date.AddDate(-1, 0, 0),
	}}
	plans := &successionPlanStub{}
	candidates := &successionCandidateStub{}
	validator := &successionPlanValidator{positions: positions, plans: plans, candidates: candidates}
	plan := &entity.SuccessionPlan{
		ID: "plan-1", PositionID: "position-1", PlanCode: "STORE-MANAGER", PlanVersion: 1,
		ReviewDate: date, ConfidentialityClass: "highly_restricted",
	}
	if err := validator.ValidateMutation(t.Context(), "hr_succession_plans", nil, plan); err != nil {
		t.Fatalf("create draft: %v", err)
	}
	updated := *plan
	updated.ReviewDate = date.AddDate(0, 1, 0)
	if err := validator.ValidateMutation(t.Context(), "hr_succession_plans", plan, &updated); err != nil {
		t.Fatalf("update draft review date: %v", err)
	}
	if err := validator.ValidateMutation(t.Context(), "hr_succession_plans", plan, nil); err != nil {
		t.Fatalf("delete unused draft: %v", err)
	}

	plans.items = []entity.SuccessionPlan{{ID: "plan-2"}}
	if err := validator.ValidateMutation(t.Context(), "hr_succession_plans", nil, plan); !errors.Is(err, hrerrors.ErrAlreadyExists) {
		t.Fatalf("duplicate version: %v", err)
	}
	plans.items = nil
	candidates.count = 1
	if err := validator.ValidateMutation(t.Context(), "hr_succession_plans", plan, &updated); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("update with candidate: %v", err)
	}
	if err := validator.ValidateMutation(t.Context(), "hr_succession_plans", plan, nil); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("delete with candidate: %v", err)
	}
}

func TestSuccessionPlanValidatorRejectsUnverifiedTransitions(t *testing.T) {
	t.Parallel()
	date := time.Date(2026, time.November, 17, 0, 0, 0, 0, time.UTC)
	positions := &successionPositionStub{position: &entity.Position{PositionStatus: "active", ValidFrom: date}}
	validator := &successionPlanValidator{positions: positions, plans: &successionPlanStub{}, candidates: &successionCandidateStub{}}
	plan := &entity.SuccessionPlan{ID: "plan-1", PositionID: "position-1", PlanCode: "MANAGER", PlanVersion: 1,
		ReviewDate: date, ConfidentialityClass: "highly_restricted"}
	actor := "actor-1"
	approved := *plan
	approved.ApprovedByPrincipalID = &actor
	if err := validator.ValidateMutation(t.Context(), "hr_succession_plans", nil, &approved); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("client approval: %v", err)
	}
	changedPosition := *plan
	changedPosition.PositionID = "position-2"
	if err := validator.ValidateMutation(t.Context(), "hr_succession_plans", plan, &changedPosition); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("move plan to another position: %v", err)
	}
	weakerClass := *plan
	weakerClass.ConfidentialityClass = "restricted"
	if err := validator.ValidateMutation(t.Context(), "hr_succession_plans", plan, &weakerClass); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("weaker confidentiality: %v", err)
	}
	positions.position.PositionStatus = "closed"
	if err := validator.ValidateMutation(t.Context(), "hr_succession_plans", nil, plan); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("inactive position: %v", err)
	}
	if err := validator.ValidateMutation(t.Context(), "hr_succession_plans", plan, nil); err != nil {
		t.Fatalf("delete unused draft after position closed: %v", err)
	}
	positions.position.PositionStatus = "active"
	positions.err = hrerrors.ErrNotFound
	if err := validator.ValidateMutation(t.Context(), "hr_succession_plans", nil, plan); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("missing position: %v", err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if err := validator.ValidateMutation(ctx, "hr_succession_plans", nil, plan); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled request: %v", err)
	}
}

func TestSuccessionOperationsRequireStepUp(t *testing.T) {
	t.Parallel()
	for _, resource := range []string{"hr_succession_plans", "hr_succession_candidates"} {
		for _, action := range []string{"read", "list", "create", "update", "delete"} {
			if !IsSensitiveOperation(resource, action) {
				t.Errorf("%s.%s must require step-up", resource, action)
			}
		}
	}
}
