package security

import (
	"bytes"
	"context"
	"reflect"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository"
)

type benefitPlanReader interface {
	FindByIDForUpdate(context.Context, string) (*entity.BenefitPlan, error)
	SelectWhere(context.Context, []domainrepo.Filter, int, int) ([]entity.BenefitPlan, error)
}

type benefitEnrollmentReader interface {
	CountWhere(context.Context, []domainrepo.Filter) (int64, error)
	SelectWhere(context.Context, []domainrepo.Filter, int, int) ([]entity.BenefitEnrollment, error)
}

type benefitPlanValidator struct {
	plans       benefitPlanReader
	enrollments benefitEnrollmentReader
}

type benefitEnrollmentValidator struct {
	workers     lockedWorkerReader
	plans       benefitPlanReader
	enrollments benefitEnrollmentReader
}

// BenefitsValidators memasang validasi katalog dan pengajuan benefit milik HRM.
func BenefitsValidators(repos *repository.Registry) (map[string]MutationValidator, error) {
	if repos == nil || operation.RequireDependency(repos.Worker) != nil ||
		operation.RequireDependency(repos.BenefitPlan) != nil || operation.RequireDependency(repos.BenefitEnrollment) != nil {
		return nil, hrerrors.ErrUnavailable
	}
	return map[string]MutationValidator{
		"hr_benefit_plans": &benefitPlanValidator{plans: repos.BenefitPlan, enrollments: repos.BenefitEnrollment},
		"hr_benefit_enrollments": &benefitEnrollmentValidator{
			workers: repos.Worker, plans: repos.BenefitPlan, enrollments: repos.BenefitEnrollment,
		},
	}, nil
}

// ValidateMutation membatasi plan lokal dan membekukan versi yang telah dipakai.
func (v *benefitPlanValidator) ValidateMutation(ctx context.Context, table string, before, after entity.Validatable) error {
	if v == nil || table != "hr_benefit_plans" || operation.RequireDependency(v.plans) != nil ||
		operation.RequireDependency(v.enrollments) != nil {
		return hrerrors.ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	var previous *entity.BenefitPlan
	if before != nil {
		var ok bool
		previous, ok = before.(*entity.BenefitPlan)
		if !ok || previous == nil {
			return hrerrors.ErrInvalidInput
		}
	}
	if after == nil {
		if previous == nil {
			return hrerrors.ErrInvalidInput
		}
		return v.rejectIfUsed(ctx, previous.ID)
	}
	plan, ok := after.(*entity.BenefitPlan)
	if !ok || plan == nil || (previous != nil && previous.ID != plan.ID) {
		return hrerrors.ErrInvalidInput
	}
	if plan.ProviderPartyReferenceID != nil || plan.EmployeeContributionRuleReferenceID != nil ||
		plan.EmployerContributionRuleReferenceID != nil {
		return hrerrors.ErrUnavailable
	}
	if !bytes.Equal(bytes.TrimSpace(plan.EligibilityRule), []byte("{}")) {
		return hrerrors.ErrWorkflowRequired
	}
	matches, err := v.plans.SelectWhere(ctx, []domainrepo.Filter{
		{Column: "plan_code", Values: []any{plan.PlanCode}},
		{Column: "plan_version", Values: []any{plan.PlanVersion}},
		{Column: "jurisdiction_code", Values: []any{plan.JurisdictionCode}},
	}, 2, 0)
	if err != nil {
		return err
	}
	for _, existing := range matches {
		if existing.ID != plan.ID {
			return hrerrors.ErrAlreadyExists
		}
	}
	if previous != nil && !reflect.DeepEqual(previous, plan) {
		return v.rejectIfUsed(ctx, plan.ID)
	}
	return nil
}

func (v *benefitPlanValidator) rejectIfUsed(ctx context.Context, planID string) error {
	count, err := v.enrollments.CountWhere(ctx, []domainrepo.Filter{{Column: "benefit_plan_id", Values: []any{planID}}})
	if err != nil {
		return err
	}
	if count != 0 {
		return hrerrors.ErrReferenceConflict
	}
	return nil
}

// ValidateMutation hanya mengizinkan pengajuan pending tanpa klaim aktivasi atau kontribusi.
func (v *benefitEnrollmentValidator) ValidateMutation(ctx context.Context, table string, before, after entity.Validatable) error {
	if v == nil || table != "hr_benefit_enrollments" || operation.RequireDependency(v.workers) != nil ||
		operation.RequireDependency(v.plans) != nil || operation.RequireDependency(v.enrollments) != nil {
		return hrerrors.ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	var previous *entity.BenefitEnrollment
	if before != nil {
		var ok bool
		previous, ok = before.(*entity.BenefitEnrollment)
		if !ok || previous == nil || previous.EnrollmentStatus != "pending" {
			return hrerrors.ErrWorkflowRequired
		}
	}
	if after == nil && previous == nil {
		return hrerrors.ErrInvalidInput
	}
	enrollment := previous
	if after != nil {
		var ok bool
		enrollment, ok = after.(*entity.BenefitEnrollment)
		if !ok || enrollment == nil || (previous != nil && (previous.ID != enrollment.ID ||
			previous.WorkerID != enrollment.WorkerID || previous.BenefitPlanID != enrollment.BenefitPlanID)) {
			return hrerrors.ErrInvalidInput
		}
		if enrollment.EnrollmentStatus != "pending" || enrollment.EnrollmentReference != nil ||
			enrollment.ProviderMemberReference != nil || enrollment.EmployeeContributionAmount != nil ||
			enrollment.EmployerContributionAmount != nil || enrollment.CurrencyCode != nil ||
			enrollment.ElectedAt != nil || enrollment.VerifiedAt != nil {
			return hrerrors.ErrWorkflowRequired
		}
	}
	worker, err := v.workers.FindByIDForUpdate(ctx, enrollment.WorkerID)
	if err != nil {
		return err
	}
	if worker == nil {
		return hrerrors.ErrReferenceConflict
	}
	plan, err := v.plans.FindByIDForUpdate(ctx, enrollment.BenefitPlanID)
	if err != nil {
		return err
	}
	if plan == nil {
		return hrerrors.ErrReferenceConflict
	}
	if after == nil {
		return nil
	}
	if worker.Status != entity.WorkerStatusActive || !plan.IsActive ||
		plan.ProviderPartyReferenceID != nil || plan.EmployeeContributionRuleReferenceID != nil ||
		plan.EmployerContributionRuleReferenceID != nil ||
		!bytes.Equal(bytes.TrimSpace(plan.EligibilityRule), []byte("{}")) ||
		(worker.HireDate != nil && entity.CompareDates(enrollment.EffectiveFrom, *worker.HireDate) < 0) ||
		(worker.TerminationDate != nil && (enrollment.EffectiveTo == nil || entity.CompareDates(*enrollment.EffectiveTo, *worker.TerminationDate) > 0)) ||
		entity.CompareDates(enrollment.EffectiveFrom, plan.ValidFrom) < 0 ||
		(plan.ValidTo != nil && (enrollment.EffectiveTo == nil || entity.CompareDates(*enrollment.EffectiveTo, *plan.ValidTo) > 0)) {
		return hrerrors.ErrReferenceConflict
	}
	matches, err := v.enrollments.SelectWhere(ctx, []domainrepo.Filter{
		{Column: "worker_id", Values: []any{enrollment.WorkerID}},
		{Column: "benefit_plan_id", Values: []any{enrollment.BenefitPlanID}},
	}, 2, 0)
	if err != nil {
		return err
	}
	for _, existing := range matches {
		if existing.ID != enrollment.ID {
			return hrerrors.ErrAlreadyExists
		}
	}
	return nil
}
