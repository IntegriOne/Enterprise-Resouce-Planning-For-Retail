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

type benefitPlanStub struct {
	plan *entity.BenefitPlan
	rows []entity.BenefitPlan
}

func (s *benefitPlanStub) FindByIDForUpdate(context.Context, string) (*entity.BenefitPlan, error) {
	return s.plan, nil
}

func (s *benefitPlanStub) SelectWhere(context.Context, []domainrepo.Filter, int, int) ([]entity.BenefitPlan, error) {
	return s.rows, nil
}

type benefitEnrollmentStub struct {
	count int64
	rows  []entity.BenefitEnrollment
}

func (s *benefitEnrollmentStub) CountWhere(context.Context, []domainrepo.Filter) (int64, error) {
	return s.count, nil
}

func (s *benefitEnrollmentStub) SelectWhere(context.Context, []domainrepo.Filter, int, int) ([]entity.BenefitEnrollment, error) {
	return s.rows, nil
}

type benefitWorkerStub struct {
	worker *entity.Worker
}

func (s *benefitWorkerStub) FindByIDForUpdate(context.Context, string) (*entity.Worker, error) {
	return s.worker, nil
}

func TestBenefitPlanValidatorFreezesUsedVersionAndExternalRules(t *testing.T) {
	t.Parallel()
	plans := &benefitPlanStub{}
	enrollments := &benefitEnrollmentStub{}
	validator := &benefitPlanValidator{plans: plans, enrollments: enrollments}
	plan := &entity.BenefitPlan{ID: "plan-1", PlanCode: "MEAL", PlanVersion: 1, JurisdictionCode: "ID", EligibilityRule: entity.JSON(`{}`)}
	if err := validator.ValidateMutation(t.Context(), "hr_benefit_plans", nil, plan); err != nil {
		t.Fatalf("create local plan: %v", err)
	}
	provider := "external-party"
	external := *plan
	external.ProviderPartyReferenceID = &provider
	if err := validator.ValidateMutation(t.Context(), "hr_benefit_plans", nil, &external); !errors.Is(err, hrerrors.ErrUnavailable) {
		t.Fatalf("unverified provider reference: %v", err)
	}
	customRule := *plan
	customRule.EligibilityRule = entity.JSON(`{"min_days":30}`)
	if err := validator.ValidateMutation(t.Context(), "hr_benefit_plans", nil, &customRule); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("unimplemented eligibility rule: %v", err)
	}
	plans.rows = []entity.BenefitPlan{{ID: "plan-2"}}
	if err := validator.ValidateMutation(t.Context(), "hr_benefit_plans", nil, plan); !errors.Is(err, hrerrors.ErrAlreadyExists) {
		t.Fatalf("duplicate plan version: %v", err)
	}
	plans.rows = nil
	enrollments.count = 1
	changed := *plan
	changed.PlanName = "Meal revised"
	if err := validator.ValidateMutation(t.Context(), "hr_benefit_plans", plan, &changed); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("change used plan: %v", err)
	}
	if err := validator.ValidateMutation(t.Context(), "hr_benefit_plans", plan, nil); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("delete used plan: %v", err)
	}
}

func TestBenefitEnrollmentValidatorKeepsRequestPending(t *testing.T) {
	t.Parallel()
	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	plans := &benefitPlanStub{plan: &entity.BenefitPlan{ID: "plan-1", IsActive: true, ValidFrom: from, EligibilityRule: entity.JSON(`{}`)}}
	workers := &benefitWorkerStub{worker: &entity.Worker{ID: "worker-1", Status: entity.WorkerStatusActive, HireDate: &from}}
	enrollments := &benefitEnrollmentStub{}
	validator := &benefitEnrollmentValidator{workers: workers, plans: plans, enrollments: enrollments}
	request := &entity.BenefitEnrollment{ID: "request-1", WorkerID: "worker-1", BenefitPlanID: "plan-1", EnrollmentStatus: "pending", EffectiveFrom: from}
	if err := validator.ValidateMutation(t.Context(), "hr_benefit_enrollments", nil, request); err != nil {
		t.Fatalf("pending request: %v", err)
	}
	approved := *request
	approved.EnrollmentStatus = "approved"
	if err := validator.ValidateMutation(t.Context(), "hr_benefit_enrollments", nil, &approved); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("direct approval: %v", err)
	}
	amount := entity.Decimal("100.000000")
	priced := *request
	priced.EmployeeContributionAmount = &amount
	if err := validator.ValidateMutation(t.Context(), "hr_benefit_enrollments", nil, &priced); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("unverified contribution: %v", err)
	}
	enrollments.rows = []entity.BenefitEnrollment{{ID: "request-2"}}
	if err := validator.ValidateMutation(t.Context(), "hr_benefit_enrollments", nil, request); !errors.Is(err, hrerrors.ErrAlreadyExists) {
		t.Fatalf("duplicate worker plan: %v", err)
	}
	enrollments.rows = nil
	plans.plan.IsActive = false
	if err := validator.ValidateMutation(t.Context(), "hr_benefit_enrollments", nil, request); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("inactive plan: %v", err)
	}
	if err := validator.ValidateMutation(t.Context(), "hr_benefit_enrollments", request, nil); err != nil {
		t.Fatalf("delete pending request after plan deactivation: %v", err)
	}
}
