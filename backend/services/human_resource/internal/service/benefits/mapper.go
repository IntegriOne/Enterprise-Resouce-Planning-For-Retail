// Kerangka awal dibuat dengan alat bantu; kode ini dipelihara langsung per domain.
// Aturan bisnis lintas record tetap dijalankan melalui policy/integrity dan command domain.
package service

import (
	"fmt"
	"reflect"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

func createBenefitPlan(in CreateBenefitPlanInput) (*entity.BenefitPlan, error) {
	v := &entity.BenefitPlan{}
	v.PlanCode = in.PlanCode
	v.PlanVersion = in.PlanVersion
	v.PlanName = in.PlanName
	v.BenefitType = in.BenefitType
	v.ProviderPartyReferenceID = in.ProviderPartyReferenceID
	v.JurisdictionCode = in.JurisdictionCode
	if in.EligibilityRule != nil {
		v.EligibilityRule = in.EligibilityRule
	} else {
		v.EligibilityRule = entity.JSON(`{}`)
	}
	v.EmployeeContributionRuleReferenceID = in.EmployeeContributionRuleReferenceID
	v.EmployerContributionRuleReferenceID = in.EmployerContributionRuleReferenceID
	v.ValidFrom = in.ValidFrom
	v.ValidTo = in.ValidTo
	if in.IsActive != nil {
		v.IsActive = *in.IsActive
	} else {
		v.IsActive = true
	}
	return v, nil
}
func patchBenefitPlan(v *entity.BenefitPlan, in UpdateBenefitPlanInput) error {
	if in.PlanCode.Set {
		if in.PlanCode.Value == nil {
			return hrerrors.Invalid("plan_code", "must not be null")
		}
		v.PlanCode = *in.PlanCode.Value
	}
	if in.PlanVersion.Set {
		if in.PlanVersion.Value == nil {
			return hrerrors.Invalid("plan_version", "must not be null")
		}
		v.PlanVersion = *in.PlanVersion.Value
	}
	if in.PlanName.Set {
		if in.PlanName.Value == nil {
			return hrerrors.Invalid("plan_name", "must not be null")
		}
		v.PlanName = *in.PlanName.Value
	}
	if in.BenefitType.Set {
		if in.BenefitType.Value == nil {
			return hrerrors.Invalid("benefit_type", "must not be null")
		}
		v.BenefitType = *in.BenefitType.Value
	}
	if in.ProviderPartyReferenceID.Set {
		v.ProviderPartyReferenceID = in.ProviderPartyReferenceID.Value
	}
	if in.JurisdictionCode.Set {
		if in.JurisdictionCode.Value == nil {
			return hrerrors.Invalid("jurisdiction_code", "must not be null")
		}
		v.JurisdictionCode = *in.JurisdictionCode.Value
	}
	if in.EligibilityRule.Set {
		if in.EligibilityRule.Value == nil {
			v.EligibilityRule = nil
		} else {
			v.EligibilityRule = *in.EligibilityRule.Value
		}
	}
	if in.EmployeeContributionRuleReferenceID.Set {
		v.EmployeeContributionRuleReferenceID = in.EmployeeContributionRuleReferenceID.Value
	}
	if in.EmployerContributionRuleReferenceID.Set {
		v.EmployerContributionRuleReferenceID = in.EmployerContributionRuleReferenceID.Value
	}
	if in.ValidFrom.Set {
		if in.ValidFrom.Value == nil {
			return hrerrors.Invalid("valid_from", "must not be null")
		}
		v.ValidFrom = *in.ValidFrom.Value
	}
	if in.ValidTo.Set {
		v.ValidTo = in.ValidTo.Value
	}
	if in.IsActive.Set {
		if in.IsActive.Value == nil {
			return hrerrors.Invalid("is_active", "must not be null")
		}
		v.IsActive = *in.IsActive.Value
	}
	return nil
}
func specBenefitPlan() operation.Specification[entity.BenefitPlan, string] {
	return operation.Specification[entity.BenefitPlan, string]{
		Domain: "benefits", Table: "hr_benefit_plans",
		Key:         func(v *entity.BenefitPlan) string { return v.ID },
		KeyString:   func(key string) string { return fmt.Sprint(key) },
		ValidateKey: func(key string) error { return entity.ValidateUUID("id", key) },
		AggregateID: func(v *entity.BenefitPlan) string { return v.ID },
		Initialize: func(v *entity.BenefitPlan, actor operation.Actor, now time.Time, id string) {
			v.ID = id
		},
		Touch: func(v *entity.BenefitPlan, actor operation.Actor, now time.Time) {},
		Attributes: func(v *entity.BenefitPlan) map[string]string {
			a := map[string]string{}
			a["id"] = fmt.Sprint(v.ID)
			if v.ProviderPartyReferenceID != nil {
				a["provider_party_reference_id"] = fmt.Sprint(*v.ProviderPartyReferenceID)
			}
			if v.EmployeeContributionRuleReferenceID != nil {
				a["employee_contribution_rule_reference_id"] = fmt.Sprint(*v.EmployeeContributionRuleReferenceID)
			}
			if v.EmployerContributionRuleReferenceID != nil {
				a["employer_contribution_rule_reference_id"] = fmt.Sprint(*v.EmployerContributionRuleReferenceID)
			}
			return a
		},
		Guard: func(action string, old, next *entity.BenefitPlan) error {
			return nil
		},
	}
}
func createBenefitEnrollment(in CreateBenefitEnrollmentInput) (*entity.BenefitEnrollment, error) {
	v := &entity.BenefitEnrollment{}
	v.WorkerID = in.WorkerID
	v.BenefitPlanID = in.BenefitPlanID
	if in.EnrollmentStatus != nil {
		v.EnrollmentStatus = *in.EnrollmentStatus
	} else {
		v.EnrollmentStatus = "pending"
	}
	v.EffectiveFrom = in.EffectiveFrom
	v.EffectiveTo = in.EffectiveTo
	v.EnrollmentReference = in.EnrollmentReference
	v.ProviderMemberReference = in.ProviderMemberReference
	v.EmployeeContributionAmount = in.EmployeeContributionAmount
	v.EmployerContributionAmount = in.EmployerContributionAmount
	v.CurrencyCode = in.CurrencyCode
	v.ElectedAt = in.ElectedAt
	v.VerifiedAt = in.VerifiedAt
	return v, nil
}
func patchBenefitEnrollment(v *entity.BenefitEnrollment, in UpdateBenefitEnrollmentInput) error {
	if in.WorkerID.Set {
		if in.WorkerID.Value == nil {
			return hrerrors.Invalid("worker_id", "must not be null")
		}
		v.WorkerID = *in.WorkerID.Value
	}
	if in.BenefitPlanID.Set {
		if in.BenefitPlanID.Value == nil {
			return hrerrors.Invalid("benefit_plan_id", "must not be null")
		}
		v.BenefitPlanID = *in.BenefitPlanID.Value
	}
	if in.EnrollmentStatus.Set {
		if in.EnrollmentStatus.Value == nil {
			return hrerrors.Invalid("enrollment_status", "must not be null")
		}
		v.EnrollmentStatus = *in.EnrollmentStatus.Value
	}
	if in.EffectiveFrom.Set {
		if in.EffectiveFrom.Value == nil {
			return hrerrors.Invalid("effective_from", "must not be null")
		}
		v.EffectiveFrom = *in.EffectiveFrom.Value
	}
	if in.EffectiveTo.Set {
		v.EffectiveTo = in.EffectiveTo.Value
	}
	if in.EnrollmentReference.Set {
		v.EnrollmentReference = in.EnrollmentReference.Value
	}
	if in.ProviderMemberReference.Set {
		v.ProviderMemberReference = in.ProviderMemberReference.Value
	}
	if in.EmployeeContributionAmount.Set {
		v.EmployeeContributionAmount = in.EmployeeContributionAmount.Value
	}
	if in.EmployerContributionAmount.Set {
		v.EmployerContributionAmount = in.EmployerContributionAmount.Value
	}
	if in.CurrencyCode.Set {
		v.CurrencyCode = in.CurrencyCode.Value
	}
	if in.ElectedAt.Set {
		v.ElectedAt = in.ElectedAt.Value
	}
	if in.VerifiedAt.Set {
		v.VerifiedAt = in.VerifiedAt.Value
	}
	return nil
}
func specBenefitEnrollment() operation.Specification[entity.BenefitEnrollment, string] {
	return operation.Specification[entity.BenefitEnrollment, string]{
		Domain: "benefits", Table: "hr_benefit_enrollments",
		Key:         func(v *entity.BenefitEnrollment) string { return v.ID },
		KeyString:   func(key string) string { return fmt.Sprint(key) },
		ValidateKey: func(key string) error { return entity.ValidateUUID("id", key) },
		AggregateID: func(v *entity.BenefitEnrollment) string { return v.ID },
		Initialize: func(v *entity.BenefitEnrollment, actor operation.Actor, now time.Time, id string) {
			v.ID = id
		},
		Touch: func(v *entity.BenefitEnrollment, actor operation.Actor, now time.Time) {},
		Attributes: func(v *entity.BenefitEnrollment) map[string]string {
			a := map[string]string{}
			a["id"] = fmt.Sprint(v.ID)
			a["worker_id"] = fmt.Sprint(v.WorkerID)
			a["benefit_plan_id"] = fmt.Sprint(v.BenefitPlanID)
			a["enrollment_status"] = fmt.Sprint(v.EnrollmentStatus)
			return a
		},
		Guard: func(action string, old, next *entity.BenefitEnrollment) error {
			if action == "update" && !reflect.DeepEqual(old.EnrollmentStatus, next.EnrollmentStatus) {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "create" && string(next.EnrollmentStatus) != "pending" {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "create" && next.VerifiedAt != nil {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "update" && !reflect.DeepEqual(old.VerifiedAt, next.VerifiedAt) {
				return hrerrors.ErrWorkflowRequired
			}
			if action != "create" {
				switch string(old.EnrollmentStatus) {
				case "approved", "posted", "paid", "finalized", "applied", "closed", "archived", "terminated", "deceased", "completed", "accepted", "issued", "signed", "hired", "processed", "rejected", "cancelled", "withdrawn", "superseded":
					return hrerrors.ErrWorkflowRequired
				}
			}
			return nil
		},
	}
}
