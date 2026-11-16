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

func createOnboardingPlan(in CreateOnboardingPlanInput) (*entity.OnboardingPlan, error) {
	v := &entity.OnboardingPlan{}
	v.WorkerID = in.WorkerID
	v.PlanCode = in.PlanCode
	if in.PlanVersion != nil {
		v.PlanVersion = *in.PlanVersion
	} else {
		v.PlanVersion = 1
	}
	v.StartDate = in.StartDate
	v.TargetCompletionDate = in.TargetCompletionDate
	if in.Status != nil {
		v.Status = *in.Status
	} else {
		v.Status = "draft"
	}
	v.TemplateReferenceID = in.TemplateReferenceID
	return v, nil
}
func patchOnboardingPlan(v *entity.OnboardingPlan, in UpdateOnboardingPlanInput) error {
	if in.WorkerID.Set {
		if in.WorkerID.Value == nil {
			return hrerrors.Invalid("worker_id", "must not be null")
		}
		v.WorkerID = *in.WorkerID.Value
	}
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
	if in.StartDate.Set {
		if in.StartDate.Value == nil {
			return hrerrors.Invalid("start_date", "must not be null")
		}
		v.StartDate = *in.StartDate.Value
	}
	if in.TargetCompletionDate.Set {
		v.TargetCompletionDate = in.TargetCompletionDate.Value
	}
	if in.Status.Set {
		if in.Status.Value == nil {
			return hrerrors.Invalid("status", "must not be null")
		}
		v.Status = *in.Status.Value
	}
	if in.TemplateReferenceID.Set {
		v.TemplateReferenceID = in.TemplateReferenceID.Value
	}
	return nil
}
func specOnboardingPlan() operation.Specification[entity.OnboardingPlan, string] {
	return operation.Specification[entity.OnboardingPlan, string]{
		Domain: "onboarding", Table: "hr_onboarding_plans",
		Key:         func(v *entity.OnboardingPlan) string { return v.ID },
		KeyString:   func(key string) string { return fmt.Sprint(key) },
		ValidateKey: func(key string) error { return entity.ValidateUUID("id", key) },
		AggregateID: func(v *entity.OnboardingPlan) string { return v.ID },
		Initialize: func(v *entity.OnboardingPlan, actor operation.Actor, now time.Time, id string) {
			v.ID = id
			v.CreatedByPrincipalID = &actor.PrincipalID
		},
		Touch: func(v *entity.OnboardingPlan, actor operation.Actor, now time.Time) {},
		Attributes: func(v *entity.OnboardingPlan) map[string]string {
			a := map[string]string{}
			a["id"] = fmt.Sprint(v.ID)
			a["worker_id"] = fmt.Sprint(v.WorkerID)
			a["status"] = fmt.Sprint(v.Status)
			if v.TemplateReferenceID != nil {
				a["template_reference_id"] = fmt.Sprint(*v.TemplateReferenceID)
			}
			if v.CreatedByPrincipalID != nil {
				a["created_by_principal_id"] = fmt.Sprint(*v.CreatedByPrincipalID)
			}
			return a
		},
		Guard: func(action string, old, next *entity.OnboardingPlan) error {
			if action == "update" && !reflect.DeepEqual(old.Status, next.Status) {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "create" && string(next.Status) != "draft" {
				return hrerrors.ErrWorkflowRequired
			}
			if action != "create" {
				switch string(old.Status) {
				case "approved", "posted", "paid", "finalized", "applied", "closed", "archived", "terminated", "deceased", "completed", "accepted", "issued", "signed", "hired", "processed", "rejected", "cancelled", "withdrawn", "superseded":
					return hrerrors.ErrWorkflowRequired
				}
			}
			return nil
		},
	}
}
func createOnboardingTask(in CreateOnboardingTaskInput) (*entity.OnboardingTask, error) {
	v := &entity.OnboardingTask{}
	v.OnboardingPlanID = in.OnboardingPlanID
	v.TaskCode = in.TaskCode
	v.TaskName = in.TaskName
	v.TaskOwnerType = in.TaskOwnerType
	v.AssignedPrincipalID = in.AssignedPrincipalID
	v.DueAt = in.DueAt
	v.CompletedAt = in.CompletedAt
	v.CompletionEvidenceReferenceID = in.CompletionEvidenceReferenceID
	if in.IsMandatory != nil {
		v.IsMandatory = *in.IsMandatory
	} else {
		v.IsMandatory = true
	}
	return v, nil
}
func patchOnboardingTask(v *entity.OnboardingTask, in UpdateOnboardingTaskInput) error {
	if in.OnboardingPlanID.Set {
		if in.OnboardingPlanID.Value == nil {
			return hrerrors.Invalid("onboarding_plan_id", "must not be null")
		}
		v.OnboardingPlanID = *in.OnboardingPlanID.Value
	}
	if in.TaskCode.Set {
		if in.TaskCode.Value == nil {
			return hrerrors.Invalid("task_code", "must not be null")
		}
		v.TaskCode = *in.TaskCode.Value
	}
	if in.TaskName.Set {
		if in.TaskName.Value == nil {
			return hrerrors.Invalid("task_name", "must not be null")
		}
		v.TaskName = *in.TaskName.Value
	}
	if in.TaskOwnerType.Set {
		if in.TaskOwnerType.Value == nil {
			return hrerrors.Invalid("task_owner_type", "must not be null")
		}
		v.TaskOwnerType = *in.TaskOwnerType.Value
	}
	if in.AssignedPrincipalID.Set {
		v.AssignedPrincipalID = in.AssignedPrincipalID.Value
	}
	if in.DueAt.Set {
		v.DueAt = in.DueAt.Value
	}
	if in.CompletedAt.Set {
		v.CompletedAt = in.CompletedAt.Value
	}
	if in.CompletionEvidenceReferenceID.Set {
		v.CompletionEvidenceReferenceID = in.CompletionEvidenceReferenceID.Value
	}
	if in.IsMandatory.Set {
		if in.IsMandatory.Value == nil {
			return hrerrors.Invalid("is_mandatory", "must not be null")
		}
		v.IsMandatory = *in.IsMandatory.Value
	}
	return nil
}
func specOnboardingTask() operation.Specification[entity.OnboardingTask, string] {
	return operation.Specification[entity.OnboardingTask, string]{
		Domain: "onboarding", Table: "hr_onboarding_tasks",
		Key:         func(v *entity.OnboardingTask) string { return v.ID },
		KeyString:   func(key string) string { return fmt.Sprint(key) },
		ValidateKey: func(key string) error { return entity.ValidateUUID("id", key) },
		AggregateID: func(v *entity.OnboardingTask) string { return v.ID },
		Initialize: func(v *entity.OnboardingTask, actor operation.Actor, now time.Time, id string) {
			v.ID = id
		},
		Touch: func(v *entity.OnboardingTask, actor operation.Actor, now time.Time) {},
		Attributes: func(v *entity.OnboardingTask) map[string]string {
			a := map[string]string{}
			a["id"] = fmt.Sprint(v.ID)
			a["onboarding_plan_id"] = fmt.Sprint(v.OnboardingPlanID)
			if v.AssignedPrincipalID != nil {
				a["assigned_principal_id"] = fmt.Sprint(*v.AssignedPrincipalID)
			}
			if v.CompletionEvidenceReferenceID != nil {
				a["completion_evidence_reference_id"] = fmt.Sprint(*v.CompletionEvidenceReferenceID)
			}
			return a
		},
		Guard: func(action string, old, next *entity.OnboardingTask) error {
			if action == "create" && next.CompletedAt != nil {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "update" && !reflect.DeepEqual(old.CompletedAt, next.CompletedAt) {
				return hrerrors.ErrWorkflowRequired
			}
			return nil
		},
	}
}
