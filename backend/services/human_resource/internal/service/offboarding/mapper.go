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

func createOffboardingPlan(in CreateOffboardingPlanInput) (*entity.OffboardingPlan, error) {
	v := &entity.OffboardingPlan{}
	v.WorkerID = in.WorkerID
	v.PlanNumber = in.PlanNumber
	v.ReasonCode = in.ReasonCode
	v.NoticeDate = in.NoticeDate
	v.LastWorkingDate = in.LastWorkingDate
	v.TerminationEffectiveDate = in.TerminationEffectiveDate
	if in.Status != nil {
		v.Status = *in.Status
	} else {
		v.Status = "draft"
	}
	v.ExitInterviewDocumentVersionReferenceID = in.ExitInterviewDocumentVersionReferenceID
	v.FinalPayrollRunID = in.FinalPayrollRunID
	v.FinanceClearanceReferenceID = in.FinanceClearanceReferenceID
	v.ApprovedByPrincipalID = in.ApprovedByPrincipalID
	v.ApprovedAt = in.ApprovedAt
	v.CompletedAt = in.CompletedAt
	return v, nil
}
func patchOffboardingPlan(v *entity.OffboardingPlan, in UpdateOffboardingPlanInput) error {
	if in.WorkerID.Set {
		if in.WorkerID.Value == nil {
			return hrerrors.Invalid("worker_id", "must not be null")
		}
		v.WorkerID = *in.WorkerID.Value
	}
	if in.PlanNumber.Set {
		if in.PlanNumber.Value == nil {
			return hrerrors.Invalid("plan_number", "must not be null")
		}
		v.PlanNumber = *in.PlanNumber.Value
	}
	if in.ReasonCode.Set {
		if in.ReasonCode.Value == nil {
			return hrerrors.Invalid("reason_code", "must not be null")
		}
		v.ReasonCode = *in.ReasonCode.Value
	}
	if in.NoticeDate.Set {
		if in.NoticeDate.Value == nil {
			return hrerrors.Invalid("notice_date", "must not be null")
		}
		v.NoticeDate = *in.NoticeDate.Value
	}
	if in.LastWorkingDate.Set {
		if in.LastWorkingDate.Value == nil {
			return hrerrors.Invalid("last_working_date", "must not be null")
		}
		v.LastWorkingDate = *in.LastWorkingDate.Value
	}
	if in.TerminationEffectiveDate.Set {
		if in.TerminationEffectiveDate.Value == nil {
			return hrerrors.Invalid("termination_effective_date", "must not be null")
		}
		v.TerminationEffectiveDate = *in.TerminationEffectiveDate.Value
	}
	if in.Status.Set {
		if in.Status.Value == nil {
			return hrerrors.Invalid("status", "must not be null")
		}
		v.Status = *in.Status.Value
	}
	if in.ExitInterviewDocumentVersionReferenceID.Set {
		v.ExitInterviewDocumentVersionReferenceID = in.ExitInterviewDocumentVersionReferenceID.Value
	}
	if in.FinalPayrollRunID.Set {
		v.FinalPayrollRunID = in.FinalPayrollRunID.Value
	}
	if in.FinanceClearanceReferenceID.Set {
		v.FinanceClearanceReferenceID = in.FinanceClearanceReferenceID.Value
	}
	if in.ApprovedByPrincipalID.Set {
		v.ApprovedByPrincipalID = in.ApprovedByPrincipalID.Value
	}
	if in.ApprovedAt.Set {
		v.ApprovedAt = in.ApprovedAt.Value
	}
	if in.CompletedAt.Set {
		v.CompletedAt = in.CompletedAt.Value
	}
	return nil
}
func specOffboardingPlan() operation.Specification[entity.OffboardingPlan, string] {
	return operation.Specification[entity.OffboardingPlan, string]{
		Domain: "offboarding", Table: "hr_offboarding_plans",
		Key:         func(v *entity.OffboardingPlan) string { return v.ID },
		KeyString:   func(key string) string { return fmt.Sprint(key) },
		ValidateKey: func(key string) error { return entity.ValidateUUID("id", key) },
		AggregateID: func(v *entity.OffboardingPlan) string { return v.ID },
		Initialize: func(v *entity.OffboardingPlan, actor operation.Actor, now time.Time, id string) {
			v.ID = id
			v.CreatedByPrincipalID = actor.PrincipalID
			v.CreatedAt = now
		},
		Touch: func(v *entity.OffboardingPlan, actor operation.Actor, now time.Time) {},
		Attributes: func(v *entity.OffboardingPlan) map[string]string {
			a := map[string]string{}
			a["id"] = fmt.Sprint(v.ID)
			a["worker_id"] = fmt.Sprint(v.WorkerID)
			a["status"] = fmt.Sprint(v.Status)
			if v.ExitInterviewDocumentVersionReferenceID != nil {
				a["exit_interview_document_version_reference_id"] = fmt.Sprint(*v.ExitInterviewDocumentVersionReferenceID)
			}
			if v.FinalPayrollRunID != nil {
				a["final_payroll_run_id"] = fmt.Sprint(*v.FinalPayrollRunID)
			}
			if v.FinanceClearanceReferenceID != nil {
				a["finance_clearance_reference_id"] = fmt.Sprint(*v.FinanceClearanceReferenceID)
			}
			if v.ApprovedByPrincipalID != nil {
				a["approved_by_principal_id"] = fmt.Sprint(*v.ApprovedByPrincipalID)
			}
			a["created_by_principal_id"] = fmt.Sprint(v.CreatedByPrincipalID)
			return a
		},
		Guard: func(action string, old, next *entity.OffboardingPlan) error {
			if action == "update" && !reflect.DeepEqual(old.Status, next.Status) {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "create" && string(next.Status) != "draft" {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "create" && next.ApprovedByPrincipalID != nil {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "update" && !reflect.DeepEqual(old.ApprovedByPrincipalID, next.ApprovedByPrincipalID) {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "create" && next.ApprovedAt != nil {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "update" && !reflect.DeepEqual(old.ApprovedAt, next.ApprovedAt) {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "create" && next.CompletedAt != nil {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "update" && !reflect.DeepEqual(old.CompletedAt, next.CompletedAt) {
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
func createOffboardingTask(in CreateOffboardingTaskInput) (*entity.OffboardingTask, error) {
	v := &entity.OffboardingTask{}
	v.OffboardingPlanID = in.OffboardingPlanID
	v.TaskCode = in.TaskCode
	v.TaskName = in.TaskName
	v.OwnerManagement = in.OwnerManagement
	v.AssignedPrincipalID = in.AssignedPrincipalID
	if in.Status != nil {
		v.Status = *in.Status
	} else {
		v.Status = "pending"
	}
	if in.IsMandatory != nil {
		v.IsMandatory = *in.IsMandatory
	} else {
		v.IsMandatory = true
	}
	v.DueAt = in.DueAt
	v.CompletedAt = in.CompletedAt
	v.OwnerConfirmationReferenceID = in.OwnerConfirmationReferenceID
	v.WaiverReason = in.WaiverReason
	v.WaivedByPrincipalID = in.WaivedByPrincipalID
	return v, nil
}
func patchOffboardingTask(v *entity.OffboardingTask, in UpdateOffboardingTaskInput) error {
	if in.OffboardingPlanID.Set {
		if in.OffboardingPlanID.Value == nil {
			return hrerrors.Invalid("offboarding_plan_id", "must not be null")
		}
		v.OffboardingPlanID = *in.OffboardingPlanID.Value
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
	if in.OwnerManagement.Set {
		if in.OwnerManagement.Value == nil {
			return hrerrors.Invalid("owner_management", "must not be null")
		}
		v.OwnerManagement = *in.OwnerManagement.Value
	}
	if in.AssignedPrincipalID.Set {
		v.AssignedPrincipalID = in.AssignedPrincipalID.Value
	}
	if in.Status.Set {
		if in.Status.Value == nil {
			return hrerrors.Invalid("status", "must not be null")
		}
		v.Status = *in.Status.Value
	}
	if in.IsMandatory.Set {
		if in.IsMandatory.Value == nil {
			return hrerrors.Invalid("is_mandatory", "must not be null")
		}
		v.IsMandatory = *in.IsMandatory.Value
	}
	if in.DueAt.Set {
		v.DueAt = in.DueAt.Value
	}
	if in.CompletedAt.Set {
		v.CompletedAt = in.CompletedAt.Value
	}
	if in.OwnerConfirmationReferenceID.Set {
		v.OwnerConfirmationReferenceID = in.OwnerConfirmationReferenceID.Value
	}
	if in.WaiverReason.Set {
		v.WaiverReason = in.WaiverReason.Value
	}
	if in.WaivedByPrincipalID.Set {
		v.WaivedByPrincipalID = in.WaivedByPrincipalID.Value
	}
	return nil
}
func specOffboardingTask() operation.Specification[entity.OffboardingTask, string] {
	return operation.Specification[entity.OffboardingTask, string]{
		Domain: "offboarding", Table: "hr_offboarding_tasks",
		Key:         func(v *entity.OffboardingTask) string { return v.ID },
		KeyString:   func(key string) string { return fmt.Sprint(key) },
		ValidateKey: func(key string) error { return entity.ValidateUUID("id", key) },
		AggregateID: func(v *entity.OffboardingTask) string { return v.ID },
		Initialize: func(v *entity.OffboardingTask, actor operation.Actor, now time.Time, id string) {
			v.ID = id
		},
		Touch: func(v *entity.OffboardingTask, actor operation.Actor, now time.Time) {},
		Attributes: func(v *entity.OffboardingTask) map[string]string {
			a := map[string]string{}
			a["id"] = fmt.Sprint(v.ID)
			a["offboarding_plan_id"] = fmt.Sprint(v.OffboardingPlanID)
			if v.AssignedPrincipalID != nil {
				a["assigned_principal_id"] = fmt.Sprint(*v.AssignedPrincipalID)
			}
			a["status"] = fmt.Sprint(v.Status)
			if v.OwnerConfirmationReferenceID != nil {
				a["owner_confirmation_reference_id"] = fmt.Sprint(*v.OwnerConfirmationReferenceID)
			}
			if v.WaivedByPrincipalID != nil {
				a["waived_by_principal_id"] = fmt.Sprint(*v.WaivedByPrincipalID)
			}
			return a
		},
		EventDetails: func(action string, task *entity.OffboardingTask) map[string]string {
			if action != "complete" || task.OwnerConfirmationReferenceID == nil {
				return nil
			}
			return map[string]string{
				"offboarding_plan_id":             task.OffboardingPlanID,
				"owner_management":                task.OwnerManagement,
				"owner_confirmation_reference_id": *task.OwnerConfirmationReferenceID,
			}
		},
		Guard: func(action string, old, next *entity.OffboardingTask) error {
			if action == "update" && !reflect.DeepEqual(old.Status, next.Status) {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "create" && string(next.Status) != "pending" {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "create" && next.CompletedAt != nil {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "update" && !reflect.DeepEqual(old.CompletedAt, next.CompletedAt) {
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
