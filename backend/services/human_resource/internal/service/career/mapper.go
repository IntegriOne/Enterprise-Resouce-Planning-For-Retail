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

func createDevelopmentPlan(in CreateDevelopmentPlanInput) (*entity.DevelopmentPlan, error) {
	v := &entity.DevelopmentPlan{}
	v.WorkerID = in.WorkerID
	v.PlanCode = in.PlanCode
	v.PlanVersion = in.PlanVersion
	v.TargetJobID = in.TargetJobID
	v.StartDate = in.StartDate
	v.TargetDate = in.TargetDate
	if in.Status != nil {
		v.Status = *in.Status
	} else {
		v.Status = "draft"
	}
	v.MentorWorkerID = in.MentorWorkerID
	v.ApprovedByPrincipalID = in.ApprovedByPrincipalID
	v.ApprovedAt = in.ApprovedAt
	return v, nil
}
func patchDevelopmentPlan(v *entity.DevelopmentPlan, in UpdateDevelopmentPlanInput) error {
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
	if in.TargetJobID.Set {
		v.TargetJobID = in.TargetJobID.Value
	}
	if in.StartDate.Set {
		if in.StartDate.Value == nil {
			return hrerrors.Invalid("start_date", "must not be null")
		}
		v.StartDate = *in.StartDate.Value
	}
	if in.TargetDate.Set {
		if in.TargetDate.Value == nil {
			return hrerrors.Invalid("target_date", "must not be null")
		}
		v.TargetDate = *in.TargetDate.Value
	}
	if in.Status.Set {
		if in.Status.Value == nil {
			return hrerrors.Invalid("status", "must not be null")
		}
		v.Status = *in.Status.Value
	}
	if in.MentorWorkerID.Set {
		v.MentorWorkerID = in.MentorWorkerID.Value
	}
	if in.ApprovedByPrincipalID.Set {
		v.ApprovedByPrincipalID = in.ApprovedByPrincipalID.Value
	}
	if in.ApprovedAt.Set {
		v.ApprovedAt = in.ApprovedAt.Value
	}
	return nil
}
func specDevelopmentPlan() operation.Specification[entity.DevelopmentPlan, string] {
	return operation.Specification[entity.DevelopmentPlan, string]{
		Domain: "career", Table: "hr_development_plans",
		Key:         func(v *entity.DevelopmentPlan) string { return v.ID },
		KeyString:   func(key string) string { return fmt.Sprint(key) },
		ValidateKey: func(key string) error { return entity.ValidateUUID("id", key) },
		AggregateID: func(v *entity.DevelopmentPlan) string { return v.ID },
		Initialize: func(v *entity.DevelopmentPlan, actor operation.Actor, now time.Time, id string) {
			v.ID = id
		},
		Touch: func(v *entity.DevelopmentPlan, actor operation.Actor, now time.Time) {},
		Attributes: func(v *entity.DevelopmentPlan) map[string]string {
			a := map[string]string{}
			a["id"] = fmt.Sprint(v.ID)
			a["worker_id"] = fmt.Sprint(v.WorkerID)
			if v.TargetJobID != nil {
				a["target_job_id"] = fmt.Sprint(*v.TargetJobID)
			}
			a["status"] = fmt.Sprint(v.Status)
			if v.MentorWorkerID != nil {
				a["mentor_worker_id"] = fmt.Sprint(*v.MentorWorkerID)
			}
			if v.ApprovedByPrincipalID != nil {
				a["approved_by_principal_id"] = fmt.Sprint(*v.ApprovedByPrincipalID)
			}
			return a
		},
		Guard: func(action string, old, next *entity.DevelopmentPlan) error {
			if action != "create" && old.Status != "draft" {
				return hrerrors.ErrWorkflowRequired
			}
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
func createDevelopmentAction(in CreateDevelopmentActionInput) (*entity.DevelopmentAction, error) {
	v := &entity.DevelopmentAction{}
	v.DevelopmentPlanID = in.DevelopmentPlanID
	v.ActionNumber = in.ActionNumber
	v.CompetencyID = in.CompetencyID
	v.CourseID = in.CourseID
	v.ActionType = in.ActionType
	v.Objective = in.Objective
	v.TargetLevel = in.TargetLevel
	v.DueDate = in.DueDate
	v.CompletedAt = in.CompletedAt
	v.EvidenceDocumentVersionReferenceID = in.EvidenceDocumentVersionReferenceID
	return v, nil
}
func patchDevelopmentAction(v *entity.DevelopmentAction, in UpdateDevelopmentActionInput) error {
	if in.DevelopmentPlanID.Set {
		if in.DevelopmentPlanID.Value == nil {
			return hrerrors.Invalid("development_plan_id", "must not be null")
		}
		v.DevelopmentPlanID = *in.DevelopmentPlanID.Value
	}
	if in.ActionNumber.Set {
		if in.ActionNumber.Value == nil {
			return hrerrors.Invalid("action_number", "must not be null")
		}
		v.ActionNumber = *in.ActionNumber.Value
	}
	if in.CompetencyID.Set {
		v.CompetencyID = in.CompetencyID.Value
	}
	if in.CourseID.Set {
		v.CourseID = in.CourseID.Value
	}
	if in.ActionType.Set {
		if in.ActionType.Value == nil {
			return hrerrors.Invalid("action_type", "must not be null")
		}
		v.ActionType = *in.ActionType.Value
	}
	if in.Objective.Set {
		if in.Objective.Value == nil {
			return hrerrors.Invalid("objective", "must not be null")
		}
		v.Objective = *in.Objective.Value
	}
	if in.TargetLevel.Set {
		v.TargetLevel = in.TargetLevel.Value
	}
	if in.DueDate.Set {
		if in.DueDate.Value == nil {
			return hrerrors.Invalid("due_date", "must not be null")
		}
		v.DueDate = *in.DueDate.Value
	}
	if in.CompletedAt.Set {
		v.CompletedAt = in.CompletedAt.Value
	}
	if in.EvidenceDocumentVersionReferenceID.Set {
		v.EvidenceDocumentVersionReferenceID = in.EvidenceDocumentVersionReferenceID.Value
	}
	return nil
}
func specDevelopmentAction() operation.Specification[entity.DevelopmentAction, string] {
	return operation.Specification[entity.DevelopmentAction, string]{
		Domain: "career", Table: "hr_development_actions",
		Key:         func(v *entity.DevelopmentAction) string { return v.ID },
		KeyString:   func(key string) string { return fmt.Sprint(key) },
		ValidateKey: func(key string) error { return entity.ValidateUUID("id", key) },
		AggregateID: func(v *entity.DevelopmentAction) string { return v.ID },
		Initialize: func(v *entity.DevelopmentAction, actor operation.Actor, now time.Time, id string) {
			v.ID = id
		},
		Touch: func(v *entity.DevelopmentAction, actor operation.Actor, now time.Time) {},
		Attributes: func(v *entity.DevelopmentAction) map[string]string {
			a := map[string]string{}
			a["id"] = fmt.Sprint(v.ID)
			a["development_plan_id"] = fmt.Sprint(v.DevelopmentPlanID)
			if v.CompetencyID != nil {
				a["competency_id"] = fmt.Sprint(*v.CompetencyID)
			}
			if v.CourseID != nil {
				a["course_id"] = fmt.Sprint(*v.CourseID)
			}
			if v.EvidenceDocumentVersionReferenceID != nil {
				a["evidence_document_version_reference_id"] = fmt.Sprint(*v.EvidenceDocumentVersionReferenceID)
			}
			return a
		},
		Guard: func(action string, old, next *entity.DevelopmentAction) error {
			if action == "create" && next.CompletedAt != nil {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "delete" && (old.CompletedAt != nil || old.EvidenceDocumentVersionReferenceID != nil) {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "update" && old.CompletedAt != nil {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "update" && !reflect.DeepEqual(old.CompletedAt, next.CompletedAt) {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "update" && !reflect.DeepEqual(old.EvidenceDocumentVersionReferenceID, next.EvidenceDocumentVersionReferenceID) {
				return hrerrors.ErrWorkflowRequired
			}
			return nil
		},
	}
}
func createCareerPath(in CreateCareerPathInput) (*entity.CareerPath, error) {
	v := &entity.CareerPath{}
	v.PathCode = in.PathCode
	v.PathVersion = in.PathVersion
	v.PathName = in.PathName
	v.JobFamilyID = in.JobFamilyID
	v.ValidFrom = in.ValidFrom
	v.ValidTo = in.ValidTo
	return v, nil
}
func patchCareerPath(v *entity.CareerPath, in UpdateCareerPathInput) error {
	if in.PathCode.Set {
		if in.PathCode.Value == nil {
			return hrerrors.Invalid("path_code", "must not be null")
		}
		v.PathCode = *in.PathCode.Value
	}
	if in.PathVersion.Set {
		if in.PathVersion.Value == nil {
			return hrerrors.Invalid("path_version", "must not be null")
		}
		v.PathVersion = *in.PathVersion.Value
	}
	if in.PathName.Set {
		if in.PathName.Value == nil {
			return hrerrors.Invalid("path_name", "must not be null")
		}
		v.PathName = *in.PathName.Value
	}
	if in.JobFamilyID.Set {
		if in.JobFamilyID.Value == nil {
			return hrerrors.Invalid("job_family_id", "must not be null")
		}
		v.JobFamilyID = *in.JobFamilyID.Value
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
func specCareerPath() operation.Specification[entity.CareerPath, string] {
	return operation.Specification[entity.CareerPath, string]{
		Domain: "career", Table: "hr_career_paths",
		Key:         func(v *entity.CareerPath) string { return v.ID },
		KeyString:   func(key string) string { return fmt.Sprint(key) },
		ValidateKey: func(key string) error { return entity.ValidateUUID("id", key) },
		AggregateID: func(v *entity.CareerPath) string { return v.ID },
		Initialize: func(v *entity.CareerPath, actor operation.Actor, now time.Time, id string) {
			v.ID = id
			v.Status = "draft"
			v.CreatedByPrincipalID = actor.PrincipalID
			v.CreatedAt = now
		},
		Touch: func(v *entity.CareerPath, actor operation.Actor, now time.Time) {},
		Attributes: func(v *entity.CareerPath) map[string]string {
			a := map[string]string{}
			a["id"] = fmt.Sprint(v.ID)
			a["job_family_id"] = fmt.Sprint(v.JobFamilyID)
			a["created_by_principal_id"] = v.CreatedByPrincipalID
			if v.ApprovedByPrincipalID != nil {
				a["approved_by_principal_id"] = *v.ApprovedByPrincipalID
			}
			return a
		},
		Guard: func(action string, old, next *entity.CareerPath) error {
			if action == "create" && next != nil && next.Status == "draft" &&
				next.ApprovedByPrincipalID == nil && next.ApprovedAt == nil {
				return nil
			}
			if action == "update" && old != nil && next != nil && old.Status == "draft" && next.Status == "draft" &&
				old.CreatedByPrincipalID == next.CreatedByPrincipalID && old.CreatedAt.Equal(next.CreatedAt) &&
				reflect.DeepEqual(old.ApprovedByPrincipalID, next.ApprovedByPrincipalID) &&
				reflect.DeepEqual(old.ApprovedAt, next.ApprovedAt) {
				return nil
			}
			if action == "delete" && old != nil && old.Status == "draft" {
				return nil
			}
			if action != "create" {
				return hrerrors.ErrWorkflowRequired
			}
			return hrerrors.ErrInvalidInput
		},
	}
}
func createCareerPathTransition(in CreateCareerPathTransitionInput) (*entity.CareerPathTransition, error) {
	v := &entity.CareerPathTransition{}
	v.CareerPathID = in.CareerPathID
	v.FromJobID = in.FromJobID
	v.ToJobID = in.ToJobID
	v.MovementType = in.MovementType
	if in.MinimumTenureMonths != nil {
		v.MinimumTenureMonths = *in.MinimumTenureMonths
	} else {
		v.MinimumTenureMonths = 0
	}
	v.RequirementsDocumentVersionReferenceID = in.RequirementsDocumentVersionReferenceID
	v.RequirementsDigest = in.RequirementsDigest
	return v, nil
}
func patchCareerPathTransition(v *entity.CareerPathTransition, in UpdateCareerPathTransitionInput) error {
	if in.CareerPathID.Set {
		if in.CareerPathID.Value == nil {
			return hrerrors.Invalid("career_path_id", "must not be null")
		}
		v.CareerPathID = *in.CareerPathID.Value
	}
	if in.FromJobID.Set {
		if in.FromJobID.Value == nil {
			return hrerrors.Invalid("from_job_id", "must not be null")
		}
		v.FromJobID = *in.FromJobID.Value
	}
	if in.ToJobID.Set {
		if in.ToJobID.Value == nil {
			return hrerrors.Invalid("to_job_id", "must not be null")
		}
		v.ToJobID = *in.ToJobID.Value
	}
	if in.MovementType.Set {
		if in.MovementType.Value == nil {
			return hrerrors.Invalid("movement_type", "must not be null")
		}
		v.MovementType = *in.MovementType.Value
	}
	if in.MinimumTenureMonths.Set {
		if in.MinimumTenureMonths.Value == nil {
			return hrerrors.Invalid("minimum_tenure_months", "must not be null")
		}
		v.MinimumTenureMonths = *in.MinimumTenureMonths.Value
	}
	if in.RequirementsDocumentVersionReferenceID.Set {
		if in.RequirementsDocumentVersionReferenceID.Value == nil {
			return hrerrors.Invalid("requirements_document_version_reference_id", "must not be null")
		}
		v.RequirementsDocumentVersionReferenceID = *in.RequirementsDocumentVersionReferenceID.Value
	}
	if in.RequirementsDigest.Set {
		if in.RequirementsDigest.Value == nil {
			v.RequirementsDigest = nil
		} else {
			v.RequirementsDigest = *in.RequirementsDigest.Value
		}
	}
	return nil
}
func specCareerPathTransition() operation.Specification[entity.CareerPathTransition, string] {
	return operation.Specification[entity.CareerPathTransition, string]{
		Domain: "career", Table: "hr_career_path_transitions",
		Key:         func(v *entity.CareerPathTransition) string { return v.ID },
		KeyString:   func(key string) string { return fmt.Sprint(key) },
		ValidateKey: func(key string) error { return entity.ValidateUUID("id", key) },
		AggregateID: func(v *entity.CareerPathTransition) string { return v.ID },
		Initialize: func(v *entity.CareerPathTransition, actor operation.Actor, now time.Time, id string) {
			v.ID = id
		},
		Touch: func(v *entity.CareerPathTransition, actor operation.Actor, now time.Time) {},
		Attributes: func(v *entity.CareerPathTransition) map[string]string {
			a := map[string]string{}
			a["id"] = fmt.Sprint(v.ID)
			a["career_path_id"] = fmt.Sprint(v.CareerPathID)
			a["from_job_id"] = fmt.Sprint(v.FromJobID)
			a["to_job_id"] = fmt.Sprint(v.ToJobID)
			a["requirements_document_version_reference_id"] = fmt.Sprint(v.RequirementsDocumentVersionReferenceID)
			return a
		},
		Guard: func(action string, old, next *entity.CareerPathTransition) error {
			return nil
		},
	}
}
func createWorkerCareerAspiration(in CreateWorkerCareerAspirationInput) (*entity.WorkerCareerAspiration, error) {
	v := &entity.WorkerCareerAspiration{}
	v.WorkerID = in.WorkerID
	v.TargetJobID = in.TargetJobID
	v.CareerPathTransitionID = in.CareerPathTransitionID
	v.PreferenceRank = in.PreferenceRank
	v.TargetDate = in.TargetDate
	if in.MobilityPreference != nil {
		v.MobilityPreference = *in.MobilityPreference
	} else {
		v.MobilityPreference = "unspecified"
	}
	v.PreferredLocationReferenceID = in.PreferredLocationReferenceID
	v.WithdrawnAt = in.WithdrawnAt
	return v, nil
}
func patchWorkerCareerAspiration(v *entity.WorkerCareerAspiration, in UpdateWorkerCareerAspirationInput) error {
	if in.WorkerID.Set {
		if in.WorkerID.Value == nil {
			return hrerrors.Invalid("worker_id", "must not be null")
		}
		v.WorkerID = *in.WorkerID.Value
	}
	if in.TargetJobID.Set {
		if in.TargetJobID.Value == nil {
			return hrerrors.Invalid("target_job_id", "must not be null")
		}
		v.TargetJobID = *in.TargetJobID.Value
	}
	if in.CareerPathTransitionID.Set {
		v.CareerPathTransitionID = in.CareerPathTransitionID.Value
	}
	if in.PreferenceRank.Set {
		if in.PreferenceRank.Value == nil {
			return hrerrors.Invalid("preference_rank", "must not be null")
		}
		v.PreferenceRank = *in.PreferenceRank.Value
	}
	if in.TargetDate.Set {
		v.TargetDate = in.TargetDate.Value
	}
	if in.MobilityPreference.Set {
		if in.MobilityPreference.Value == nil {
			return hrerrors.Invalid("mobility_preference", "must not be null")
		}
		v.MobilityPreference = *in.MobilityPreference.Value
	}
	if in.PreferredLocationReferenceID.Set {
		v.PreferredLocationReferenceID = in.PreferredLocationReferenceID.Value
	}
	if in.WithdrawnAt.Set {
		v.WithdrawnAt = in.WithdrawnAt.Value
	}
	return nil
}
func specWorkerCareerAspiration() operation.Specification[entity.WorkerCareerAspiration, string] {
	return operation.Specification[entity.WorkerCareerAspiration, string]{
		Domain: "career", Table: "hr_worker_career_aspirations",
		Key:         func(v *entity.WorkerCareerAspiration) string { return v.ID },
		KeyString:   func(key string) string { return fmt.Sprint(key) },
		ValidateKey: func(key string) error { return entity.ValidateUUID("id", key) },
		AggregateID: func(v *entity.WorkerCareerAspiration) string { return v.ID },
		Initialize: func(v *entity.WorkerCareerAspiration, actor operation.Actor, now time.Time, id string) {
			v.ID = id
			v.RecordedAt = now
			v.RecordedByPrincipalID = actor.PrincipalID
		},
		Touch: func(v *entity.WorkerCareerAspiration, actor operation.Actor, now time.Time) {},
		Attributes: func(v *entity.WorkerCareerAspiration) map[string]string {
			a := map[string]string{}
			a["id"] = fmt.Sprint(v.ID)
			a["worker_id"] = fmt.Sprint(v.WorkerID)
			a["target_job_id"] = fmt.Sprint(v.TargetJobID)
			if v.CareerPathTransitionID != nil {
				a["career_path_transition_id"] = fmt.Sprint(*v.CareerPathTransitionID)
			}
			a["recorded_by_principal_id"] = fmt.Sprint(v.RecordedByPrincipalID)
			if v.PreferredLocationReferenceID != nil {
				a["preferred_location_reference_id"] = fmt.Sprint(*v.PreferredLocationReferenceID)
			}
			return a
		},
		Guard: func(action string, old, next *entity.WorkerCareerAspiration) error {
			return nil
		},
	}
}
