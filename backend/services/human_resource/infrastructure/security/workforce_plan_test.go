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

type workforcePlanStub struct {
	plan    *entity.WorkforcePlan
	matches []entity.WorkforcePlan
}

func (s *workforcePlanStub) FindByIDForUpdate(context.Context, string) (*entity.WorkforcePlan, error) {
	if s.plan == nil {
		return nil, hrerrors.ErrNotFound
	}
	return s.plan, nil
}

func (s *workforcePlanStub) SelectWhere(context.Context, []domainrepo.Filter, int, int) ([]entity.WorkforcePlan, error) {
	return s.matches, nil
}

type workforcePlanLineStub struct {
	count   int64
	lines   []entity.WorkforcePlanLine
	matches []entity.WorkforcePlanLine
}

func (s *workforcePlanLineStub) CountWhere(context.Context, []domainrepo.Filter) (int64, error) {
	return s.count, nil
}

func (s *workforcePlanLineStub) SelectWhere(context.Context, []domainrepo.Filter, int, int) ([]entity.WorkforcePlanLine, error) {
	if len(s.lines) > 0 {
		return s.lines, nil
	}
	return s.matches, nil
}

type wfPlanJobStub struct {
	job *entity.Job
}

func (s *wfPlanJobStub) FindByIDForUpdate(context.Context, string) (*entity.Job, error) {
	if s.job == nil {
		return nil, hrerrors.ErrNotFound
	}
	return s.job, nil
}

type foundationVerifierStub struct {
	leErr    error
	posErr   error
	partyErr error
}

func (s *foundationVerifierStub) VerifyPositionReferences(context.Context, PositionReferenceCheck) error {
	return s.posErr
}

func (s *foundationVerifierStub) VerifyWorkerReferences(context.Context, WorkerReferenceCheck) error {
	return nil
}

func (s *foundationVerifierStub) VerifyLegalEntity(context.Context, string) error {
	return s.leErr
}

func (s *foundationVerifierStub) VerifyOrganizationUnit(context.Context, string, string) error {
	return nil
}

func (s *foundationVerifierStub) VerifyLocation(context.Context, string, string) error {
	return nil
}

func (s *foundationVerifierStub) VerifyParty(context.Context, string, bool) error {
	return s.partyErr
}

func TestWorkforcePlanAndLineValidators(t *testing.T) {
	t.Parallel()
	leID := "00000000-0000-4000-8000-000000000101"
	planID := "00000000-0000-4000-8000-000000000102"
	jobID := "00000000-0000-4000-8000-000000000103"
	unitID := "00000000-0000-4000-8000-000000000104"
	actorID := "00000000-0000-4000-8000-000000000105"
	budgetID := "00000000-0000-4000-8000-000000000106"

	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	plan := &entity.WorkforcePlan{
		ID: planID, PlanCode: "WFP-2026", PlanVersion: 1, LegalEntityReferenceID: leID,
		PeriodStart: start, PeriodEnd: end, ScenarioCode: "baseline", Status: "draft",
		CreatedByPrincipalID: actorID, CreatedAt: start,
	}

	plans := &workforcePlanStub{plan: plan}
	lines := &workforcePlanLineStub{}
	jobs := &wfPlanJobStub{job: &entity.Job{ID: jobID, IsActive: true, ValidFrom: start}}
	reqs := &positionCounterStub{}
	foundation := &foundationVerifierStub{}

	pv := &workforcePlanValidator{plans: plans, lines: lines, foundation: foundation}
	lv := &workforcePlanLineValidator{plans: plans, jobs: jobs, lines: lines, requisitions: reqs, foundation: foundation}

	if err := pv.ValidateMutation(t.Context(), "hr_workforce_plans", nil, plan); err != nil {
		t.Fatalf("create draft workforce plan: %v", err)
	}
	withBudget := *plan
	withBudget.FinanceBudgetReferenceID = &budgetID
	if err := pv.ValidateMutation(t.Context(), "hr_workforce_plans", nil, &withBudget); !errors.Is(err, hrerrors.ErrUnavailable) {
		t.Fatalf("budget reference accepted without Finance: %v", err)
	}

	lines.count = 1
	if err := pv.ValidateMutation(t.Context(), "hr_workforce_plans", plan, nil); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("deleted plan with lines: %v", err)
	}
	lines.count = 0

	line := &entity.WorkforcePlanLine{
		ID: "00000000-0000-4000-8000-000000000107", WorkforcePlanID: planID, LineNumber: 1,
		OrganizationUnitReferenceID: unitID, JobID: jobID, WorkerCategory: entity.WorkerCategoryEmployee,
		CurrentFTE: "5.000", RequiredFTE: "7.000", NeededFrom: start.AddDate(0, 1, 0),
		ActionCode: "recruit", Rationale: "Store expansion",
	}
	if err := lv.ValidateMutation(t.Context(), "hr_workforce_plan_lines", nil, line); err != nil {
		t.Fatalf("create valid workforce plan line: %v", err)
	}
	outside := *line
	outside.NeededFrom = start.AddDate(-1, 0, 0)
	if err := lv.ValidateMutation(t.Context(), "hr_workforce_plan_lines", nil, &outside); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("line outside plan period accepted: %v", err)
	}
	plan.Status = "review"
	if err := lv.ValidateMutation(t.Context(), "hr_workforce_plan_lines", nil, line); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("line edited under non-draft plan: %v", err)
	}
}
