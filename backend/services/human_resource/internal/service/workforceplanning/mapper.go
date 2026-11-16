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

func createJobFamily(in CreateJobFamilyInput) (*entity.JobFamily, error) {
	v := &entity.JobFamily{}
	v.FamilyCode = in.FamilyCode
	v.FamilyName = in.FamilyName
	v.Description = in.Description
	v.JobDomain = in.JobDomain
	if in.IsActive != nil {
		v.IsActive = *in.IsActive
	} else {
		v.IsActive = true
	}
	return v, nil
}
func patchJobFamily(v *entity.JobFamily, in UpdateJobFamilyInput) error {
	if in.FamilyCode.Set {
		if in.FamilyCode.Value == nil {
			return hrerrors.Invalid("family_code", "must not be null")
		}
		v.FamilyCode = *in.FamilyCode.Value
	}
	if in.FamilyName.Set {
		if in.FamilyName.Value == nil {
			return hrerrors.Invalid("family_name", "must not be null")
		}
		v.FamilyName = *in.FamilyName.Value
	}
	if in.Description.Set {
		v.Description = in.Description.Value
	}
	if in.JobDomain.Set {
		if in.JobDomain.Value == nil {
			return hrerrors.Invalid("job_domain", "must not be null")
		}
		v.JobDomain = *in.JobDomain.Value
	}
	if in.IsActive.Set {
		if in.IsActive.Value == nil {
			return hrerrors.Invalid("is_active", "must not be null")
		}
		v.IsActive = *in.IsActive.Value
	}
	return nil
}
func specJobFamily() operation.Specification[entity.JobFamily, string] {
	return operation.Specification[entity.JobFamily, string]{
		Domain: "workforceplanning", Table: "hr_job_families",
		Key:         func(v *entity.JobFamily) string { return v.ID },
		KeyString:   func(key string) string { return fmt.Sprint(key) },
		ValidateKey: func(key string) error { return entity.ValidateUUID("id", key) },
		AggregateID: func(v *entity.JobFamily) string { return v.ID },
		Initialize: func(v *entity.JobFamily, actor operation.Actor, now time.Time, id string) {
			v.ID = id
		},
		Touch: func(v *entity.JobFamily, actor operation.Actor, now time.Time) {},
		Attributes: func(v *entity.JobFamily) map[string]string {
			a := map[string]string{}
			a["id"] = fmt.Sprint(v.ID)
			return a
		},
		Guard: func(action string, old, next *entity.JobFamily) error {
			return nil
		},
	}
}
func createJob(in CreateJobInput) (*entity.Job, error) {
	v := &entity.Job{}
	v.JobFamilyID = in.JobFamilyID
	v.JobCode = in.JobCode
	v.JobTitle = in.JobTitle
	v.JobLevelCode = in.JobLevelCode
	v.JobDescription = in.JobDescription
	if in.Responsibilities != nil {
		v.Responsibilities = in.Responsibilities
	} else {
		v.Responsibilities = entity.JSON(`[]`)
	}
	if in.QualificationRequirements != nil {
		v.QualificationRequirements = in.QualificationRequirements
	} else {
		v.QualificationRequirements = entity.JSON(`{}`)
	}
	v.ValidFrom = in.ValidFrom
	v.ValidTo = in.ValidTo
	if in.IsActive != nil {
		v.IsActive = *in.IsActive
	} else {
		v.IsActive = true
	}
	return v, nil
}
func patchJob(v *entity.Job, in UpdateJobInput) error {
	if in.JobFamilyID.Set {
		if in.JobFamilyID.Value == nil {
			return hrerrors.Invalid("job_family_id", "must not be null")
		}
		v.JobFamilyID = *in.JobFamilyID.Value
	}
	if in.JobCode.Set {
		if in.JobCode.Value == nil {
			return hrerrors.Invalid("job_code", "must not be null")
		}
		v.JobCode = *in.JobCode.Value
	}
	if in.JobTitle.Set {
		if in.JobTitle.Value == nil {
			return hrerrors.Invalid("job_title", "must not be null")
		}
		v.JobTitle = *in.JobTitle.Value
	}
	if in.JobLevelCode.Set {
		v.JobLevelCode = in.JobLevelCode.Value
	}
	if in.JobDescription.Set {
		v.JobDescription = in.JobDescription.Value
	}
	if in.Responsibilities.Set {
		if in.Responsibilities.Value == nil {
			v.Responsibilities = nil
		} else {
			v.Responsibilities = *in.Responsibilities.Value
		}
	}
	if in.QualificationRequirements.Set {
		if in.QualificationRequirements.Value == nil {
			v.QualificationRequirements = nil
		} else {
			v.QualificationRequirements = *in.QualificationRequirements.Value
		}
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
func specJob() operation.Specification[entity.Job, string] {
	return operation.Specification[entity.Job, string]{
		Domain: "workforceplanning", Table: "hr_jobs",
		Key:         func(v *entity.Job) string { return v.ID },
		KeyString:   func(key string) string { return fmt.Sprint(key) },
		ValidateKey: func(key string) error { return entity.ValidateUUID("id", key) },
		AggregateID: func(v *entity.Job) string { return v.ID },
		Initialize: func(v *entity.Job, actor operation.Actor, now time.Time, id string) {
			v.ID = id
		},
		Touch: func(v *entity.Job, actor operation.Actor, now time.Time) {},
		Attributes: func(v *entity.Job) map[string]string {
			a := map[string]string{}
			a["id"] = fmt.Sprint(v.ID)
			a["job_family_id"] = fmt.Sprint(v.JobFamilyID)
			return a
		},
		Guard: func(action string, old, next *entity.Job) error {
			return nil
		},
	}
}
func createPosition(in CreatePositionInput) (*entity.Position, error) {
	v := &entity.Position{}
	v.PositionCode = in.PositionCode
	v.PositionTitle = in.PositionTitle
	v.JobID = in.JobID
	v.OrganizationUnitReferenceID = in.OrganizationUnitReferenceID
	v.LegalEntityReferenceID = in.LegalEntityReferenceID
	v.ReportsToPositionID = in.ReportsToPositionID
	if in.PositionStatus != nil {
		v.PositionStatus = *in.PositionStatus
	} else {
		v.PositionStatus = "planned"
	}
	if in.HeadcountAuthorized != nil {
		v.HeadcountAuthorized = *in.HeadcountAuthorized
	} else {
		v.HeadcountAuthorized = entity.Decimal("1")
	}
	v.WorkLocationReferenceID = in.WorkLocationReferenceID
	v.CostCenterReferenceID = in.CostCenterReferenceID
	v.ValidFrom = in.ValidFrom
	v.ValidTo = in.ValidTo
	return v, nil
}
func patchPosition(v *entity.Position, in UpdatePositionInput) error {
	if in.PositionCode.Set {
		if in.PositionCode.Value == nil {
			return hrerrors.Invalid("position_code", "must not be null")
		}
		v.PositionCode = *in.PositionCode.Value
	}
	if in.PositionTitle.Set {
		if in.PositionTitle.Value == nil {
			return hrerrors.Invalid("position_title", "must not be null")
		}
		v.PositionTitle = *in.PositionTitle.Value
	}
	if in.JobID.Set {
		if in.JobID.Value == nil {
			return hrerrors.Invalid("job_id", "must not be null")
		}
		v.JobID = *in.JobID.Value
	}
	if in.OrganizationUnitReferenceID.Set {
		if in.OrganizationUnitReferenceID.Value == nil {
			return hrerrors.Invalid("organization_unit_reference_id", "must not be null")
		}
		v.OrganizationUnitReferenceID = *in.OrganizationUnitReferenceID.Value
	}
	if in.LegalEntityReferenceID.Set {
		if in.LegalEntityReferenceID.Value == nil {
			return hrerrors.Invalid("legal_entity_reference_id", "must not be null")
		}
		v.LegalEntityReferenceID = *in.LegalEntityReferenceID.Value
	}
	if in.ReportsToPositionID.Set {
		v.ReportsToPositionID = in.ReportsToPositionID.Value
	}
	if in.PositionStatus.Set {
		if in.PositionStatus.Value == nil {
			return hrerrors.Invalid("position_status", "must not be null")
		}
		v.PositionStatus = *in.PositionStatus.Value
	}
	if in.HeadcountAuthorized.Set {
		if in.HeadcountAuthorized.Value == nil {
			return hrerrors.Invalid("headcount_authorized", "must not be null")
		}
		v.HeadcountAuthorized = *in.HeadcountAuthorized.Value
	}
	if in.WorkLocationReferenceID.Set {
		v.WorkLocationReferenceID = in.WorkLocationReferenceID.Value
	}
	if in.CostCenterReferenceID.Set {
		v.CostCenterReferenceID = in.CostCenterReferenceID.Value
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
	return nil
}
func specPosition() operation.Specification[entity.Position, string] {
	return operation.Specification[entity.Position, string]{
		Domain: "workforceplanning", Table: "hr_positions",
		Key:         func(v *entity.Position) string { return v.ID },
		KeyString:   func(key string) string { return fmt.Sprint(key) },
		ValidateKey: func(key string) error { return entity.ValidateUUID("id", key) },
		AggregateID: func(v *entity.Position) string { return v.ID },
		Initialize: func(v *entity.Position, actor operation.Actor, now time.Time, id string) {
			v.ID = id
		},
		Touch: func(v *entity.Position, actor operation.Actor, now time.Time) {},
		Attributes: func(v *entity.Position) map[string]string {
			a := map[string]string{}
			a["id"] = fmt.Sprint(v.ID)
			a["job_id"] = fmt.Sprint(v.JobID)
			a["organization_unit_reference_id"] = fmt.Sprint(v.OrganizationUnitReferenceID)
			a["legal_entity_reference_id"] = fmt.Sprint(v.LegalEntityReferenceID)
			if v.ReportsToPositionID != nil {
				a["reports_to_position_id"] = fmt.Sprint(*v.ReportsToPositionID)
			}
			a["position_status"] = fmt.Sprint(v.PositionStatus)
			if v.WorkLocationReferenceID != nil {
				a["work_location_reference_id"] = fmt.Sprint(*v.WorkLocationReferenceID)
			}
			if v.CostCenterReferenceID != nil {
				a["cost_center_reference_id"] = fmt.Sprint(*v.CostCenterReferenceID)
			}
			return a
		},
		Guard: func(action string, old, next *entity.Position) error {
			if action == "update" && !reflect.DeepEqual(old.PositionStatus, next.PositionStatus) {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "create" && string(next.PositionStatus) != "planned" {
				return hrerrors.ErrWorkflowRequired
			}
			if action != "create" {
				switch string(old.PositionStatus) {
				case "approved", "posted", "paid", "finalized", "applied", "closed", "archived", "terminated", "deceased", "completed", "accepted", "issued", "signed", "hired", "processed", "rejected", "cancelled", "withdrawn", "superseded":
					return hrerrors.ErrWorkflowRequired
				}
			}
			return nil
		},
	}
}
func createWorkforcePlan(in CreateWorkforcePlanInput) (*entity.WorkforcePlan, error) {
	v := &entity.WorkforcePlan{}
	v.PlanCode = in.PlanCode
	v.PlanVersion = in.PlanVersion
	v.LegalEntityReferenceID = in.LegalEntityReferenceID
	v.PeriodStart = in.PeriodStart
	v.PeriodEnd = in.PeriodEnd
	if in.ScenarioCode != nil {
		v.ScenarioCode = *in.ScenarioCode
	} else {
		v.ScenarioCode = "baseline"
	}
	if in.Status != nil {
		v.Status = *in.Status
	} else {
		v.Status = "draft"
	}
	v.FinanceBudgetReferenceID = in.FinanceBudgetReferenceID
	v.ApprovedByPrincipalID = in.ApprovedByPrincipalID
	v.ApprovedAt = in.ApprovedAt
	return v, nil
}
func patchWorkforcePlan(v *entity.WorkforcePlan, in UpdateWorkforcePlanInput) error {
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
	if in.LegalEntityReferenceID.Set {
		if in.LegalEntityReferenceID.Value == nil {
			return hrerrors.Invalid("legal_entity_reference_id", "must not be null")
		}
		v.LegalEntityReferenceID = *in.LegalEntityReferenceID.Value
	}
	if in.PeriodStart.Set {
		if in.PeriodStart.Value == nil {
			return hrerrors.Invalid("period_start", "must not be null")
		}
		v.PeriodStart = *in.PeriodStart.Value
	}
	if in.PeriodEnd.Set {
		if in.PeriodEnd.Value == nil {
			return hrerrors.Invalid("period_end", "must not be null")
		}
		v.PeriodEnd = *in.PeriodEnd.Value
	}
	if in.ScenarioCode.Set {
		if in.ScenarioCode.Value == nil {
			return hrerrors.Invalid("scenario_code", "must not be null")
		}
		v.ScenarioCode = *in.ScenarioCode.Value
	}
	if in.Status.Set {
		if in.Status.Value == nil {
			return hrerrors.Invalid("status", "must not be null")
		}
		v.Status = *in.Status.Value
	}
	if in.FinanceBudgetReferenceID.Set {
		v.FinanceBudgetReferenceID = in.FinanceBudgetReferenceID.Value
	}
	if in.ApprovedByPrincipalID.Set {
		v.ApprovedByPrincipalID = in.ApprovedByPrincipalID.Value
	}
	if in.ApprovedAt.Set {
		v.ApprovedAt = in.ApprovedAt.Value
	}
	return nil
}
func specWorkforcePlan() operation.Specification[entity.WorkforcePlan, string] {
	return operation.Specification[entity.WorkforcePlan, string]{
		Domain: "workforceplanning", Table: "hr_workforce_plans",
		Key:         func(v *entity.WorkforcePlan) string { return v.ID },
		KeyString:   func(key string) string { return fmt.Sprint(key) },
		ValidateKey: func(key string) error { return entity.ValidateUUID("id", key) },
		AggregateID: func(v *entity.WorkforcePlan) string { return v.ID },
		Initialize: func(v *entity.WorkforcePlan, actor operation.Actor, now time.Time, id string) {
			v.ID = id
			v.CreatedByPrincipalID = actor.PrincipalID
			v.CreatedAt = now
		},
		Touch: func(v *entity.WorkforcePlan, actor operation.Actor, now time.Time) {},
		Attributes: func(v *entity.WorkforcePlan) map[string]string {
			a := map[string]string{}
			a["id"] = fmt.Sprint(v.ID)
			a["legal_entity_reference_id"] = fmt.Sprint(v.LegalEntityReferenceID)
			a["status"] = fmt.Sprint(v.Status)
			if v.FinanceBudgetReferenceID != nil {
				a["finance_budget_reference_id"] = fmt.Sprint(*v.FinanceBudgetReferenceID)
			}
			a["created_by_principal_id"] = fmt.Sprint(v.CreatedByPrincipalID)
			if v.ApprovedByPrincipalID != nil {
				a["approved_by_principal_id"] = fmt.Sprint(*v.ApprovedByPrincipalID)
			}
			return a
		},
		Guard: func(action string, old, next *entity.WorkforcePlan) error {
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
func createWorkforcePlanLine(in CreateWorkforcePlanLineInput) (*entity.WorkforcePlanLine, error) {
	v := &entity.WorkforcePlanLine{}
	v.WorkforcePlanID = in.WorkforcePlanID
	v.LineNumber = in.LineNumber
	v.OrganizationUnitReferenceID = in.OrganizationUnitReferenceID
	v.JobID = in.JobID
	v.WorkLocationReferenceID = in.WorkLocationReferenceID
	v.WorkerCategory = in.WorkerCategory
	v.CurrentFTE = in.CurrentFTE
	v.RequiredFTE = in.RequiredFTE
	v.EstimatedCost = in.EstimatedCost
	v.CurrencyCode = in.CurrencyCode
	v.NeededFrom = in.NeededFrom
	v.NeededUntil = in.NeededUntil
	v.ActionCode = in.ActionCode
	v.Rationale = in.Rationale
	return v, nil
}
func patchWorkforcePlanLine(v *entity.WorkforcePlanLine, in UpdateWorkforcePlanLineInput) error {
	if in.WorkforcePlanID.Set {
		if in.WorkforcePlanID.Value == nil {
			return hrerrors.Invalid("workforce_plan_id", "must not be null")
		}
		v.WorkforcePlanID = *in.WorkforcePlanID.Value
	}
	if in.LineNumber.Set {
		if in.LineNumber.Value == nil {
			return hrerrors.Invalid("line_number", "must not be null")
		}
		v.LineNumber = *in.LineNumber.Value
	}
	if in.OrganizationUnitReferenceID.Set {
		if in.OrganizationUnitReferenceID.Value == nil {
			return hrerrors.Invalid("organization_unit_reference_id", "must not be null")
		}
		v.OrganizationUnitReferenceID = *in.OrganizationUnitReferenceID.Value
	}
	if in.JobID.Set {
		if in.JobID.Value == nil {
			return hrerrors.Invalid("job_id", "must not be null")
		}
		v.JobID = *in.JobID.Value
	}
	if in.WorkLocationReferenceID.Set {
		v.WorkLocationReferenceID = in.WorkLocationReferenceID.Value
	}
	if in.WorkerCategory.Set {
		if in.WorkerCategory.Value == nil {
			return hrerrors.Invalid("worker_category", "must not be null")
		}
		v.WorkerCategory = *in.WorkerCategory.Value
	}
	if in.CurrentFTE.Set {
		if in.CurrentFTE.Value == nil {
			return hrerrors.Invalid("current_fte", "must not be null")
		}
		v.CurrentFTE = *in.CurrentFTE.Value
	}
	if in.RequiredFTE.Set {
		if in.RequiredFTE.Value == nil {
			return hrerrors.Invalid("required_fte", "must not be null")
		}
		v.RequiredFTE = *in.RequiredFTE.Value
	}
	if in.EstimatedCost.Set {
		v.EstimatedCost = in.EstimatedCost.Value
	}
	if in.CurrencyCode.Set {
		v.CurrencyCode = in.CurrencyCode.Value
	}
	if in.NeededFrom.Set {
		if in.NeededFrom.Value == nil {
			return hrerrors.Invalid("needed_from", "must not be null")
		}
		v.NeededFrom = *in.NeededFrom.Value
	}
	if in.NeededUntil.Set {
		v.NeededUntil = in.NeededUntil.Value
	}
	if in.ActionCode.Set {
		if in.ActionCode.Value == nil {
			return hrerrors.Invalid("action_code", "must not be null")
		}
		v.ActionCode = *in.ActionCode.Value
	}
	if in.Rationale.Set {
		if in.Rationale.Value == nil {
			return hrerrors.Invalid("rationale", "must not be null")
		}
		v.Rationale = *in.Rationale.Value
	}
	return nil
}
func specWorkforcePlanLine() operation.Specification[entity.WorkforcePlanLine, string] {
	return operation.Specification[entity.WorkforcePlanLine, string]{
		Domain: "workforceplanning", Table: "hr_workforce_plan_lines",
		Key:         func(v *entity.WorkforcePlanLine) string { return v.ID },
		KeyString:   func(key string) string { return fmt.Sprint(key) },
		ValidateKey: func(key string) error { return entity.ValidateUUID("id", key) },
		AggregateID: func(v *entity.WorkforcePlanLine) string { return v.ID },
		Initialize: func(v *entity.WorkforcePlanLine, actor operation.Actor, now time.Time, id string) {
			v.ID = id
		},
		Touch: func(v *entity.WorkforcePlanLine, actor operation.Actor, now time.Time) {},
		Attributes: func(v *entity.WorkforcePlanLine) map[string]string {
			a := map[string]string{}
			a["id"] = fmt.Sprint(v.ID)
			a["workforce_plan_id"] = fmt.Sprint(v.WorkforcePlanID)
			a["organization_unit_reference_id"] = fmt.Sprint(v.OrganizationUnitReferenceID)
			a["job_id"] = fmt.Sprint(v.JobID)
			if v.WorkLocationReferenceID != nil {
				a["work_location_reference_id"] = fmt.Sprint(*v.WorkLocationReferenceID)
			}
			a["worker_category"] = fmt.Sprint(v.WorkerCategory)
			return a
		},
		Guard: func(action string, old, next *entity.WorkforcePlanLine) error {
			return nil
		},
	}
}
