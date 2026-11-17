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

func createRequisition(in CreateRequisitionInput) (*entity.Requisition, error) {
	v := &entity.Requisition{}
	v.RequisitionNumber = in.RequisitionNumber
	v.WorkforcePlanLineID = in.WorkforcePlanLineID
	v.PositionID = in.PositionID
	v.JobID = in.JobID
	v.OrganizationUnitReferenceID = in.OrganizationUnitReferenceID
	v.HiringManagerPrincipalID = in.HiringManagerPrincipalID
	v.RecruiterPrincipalID = in.RecruiterPrincipalID
	v.RequisitionType = in.RequisitionType
	if in.Status != nil {
		v.Status = *in.Status
	} else {
		v.Status = "draft"
	}
	if in.Headcount != nil {
		v.Headcount = *in.Headcount
	} else {
		v.Headcount = 1
	}
	v.TargetStartDate = in.TargetStartDate
	v.BudgetReferenceID = in.BudgetReferenceID
	v.WorkLocationReferenceID = in.WorkLocationReferenceID
	v.EmploymentCategory = in.EmploymentCategory
	if in.MinimumQualifications != nil {
		v.MinimumQualifications = in.MinimumQualifications
	} else {
		v.MinimumQualifications = entity.JSON(`{}`)
	}
	if in.DesiredQualifications != nil {
		v.DesiredQualifications = in.DesiredQualifications
	} else {
		v.DesiredQualifications = entity.JSON(`{}`)
	}
	v.ApprovedAt = in.ApprovedAt
	v.ApprovedByPrincipalID = in.ApprovedByPrincipalID
	v.OpenedAt = in.OpenedAt
	v.ClosedAt = in.ClosedAt
	v.IdempotencyKey = in.IdempotencyKey
	return v, nil
}
func patchRequisition(v *entity.Requisition, in UpdateRequisitionInput) error {
	if in.RequisitionNumber.Set {
		if in.RequisitionNumber.Value == nil {
			return hrerrors.Invalid("requisition_number", "must not be null")
		}
		v.RequisitionNumber = *in.RequisitionNumber.Value
	}
	if in.WorkforcePlanLineID.Set {
		v.WorkforcePlanLineID = in.WorkforcePlanLineID.Value
	}
	if in.PositionID.Set {
		v.PositionID = in.PositionID.Value
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
	if in.HiringManagerPrincipalID.Set {
		if in.HiringManagerPrincipalID.Value == nil {
			return hrerrors.Invalid("hiring_manager_principal_id", "must not be null")
		}
		v.HiringManagerPrincipalID = *in.HiringManagerPrincipalID.Value
	}
	if in.RecruiterPrincipalID.Set {
		v.RecruiterPrincipalID = in.RecruiterPrincipalID.Value
	}
	if in.RequisitionType.Set {
		if in.RequisitionType.Value == nil {
			return hrerrors.Invalid("requisition_type", "must not be null")
		}
		v.RequisitionType = *in.RequisitionType.Value
	}
	if in.Status.Set {
		if in.Status.Value == nil {
			return hrerrors.Invalid("status", "must not be null")
		}
		v.Status = *in.Status.Value
	}
	if in.Headcount.Set {
		if in.Headcount.Value == nil {
			return hrerrors.Invalid("headcount", "must not be null")
		}
		v.Headcount = *in.Headcount.Value
	}
	if in.TargetStartDate.Set {
		v.TargetStartDate = in.TargetStartDate.Value
	}
	if in.BudgetReferenceID.Set {
		v.BudgetReferenceID = in.BudgetReferenceID.Value
	}
	if in.WorkLocationReferenceID.Set {
		v.WorkLocationReferenceID = in.WorkLocationReferenceID.Value
	}
	if in.EmploymentCategory.Set {
		if in.EmploymentCategory.Value == nil {
			return hrerrors.Invalid("employment_category", "must not be null")
		}
		v.EmploymentCategory = *in.EmploymentCategory.Value
	}
	if in.MinimumQualifications.Set {
		if in.MinimumQualifications.Value == nil {
			v.MinimumQualifications = nil
		} else {
			v.MinimumQualifications = *in.MinimumQualifications.Value
		}
	}
	if in.DesiredQualifications.Set {
		if in.DesiredQualifications.Value == nil {
			v.DesiredQualifications = nil
		} else {
			v.DesiredQualifications = *in.DesiredQualifications.Value
		}
	}
	if in.ApprovedAt.Set {
		v.ApprovedAt = in.ApprovedAt.Value
	}
	if in.ApprovedByPrincipalID.Set {
		v.ApprovedByPrincipalID = in.ApprovedByPrincipalID.Value
	}
	if in.OpenedAt.Set {
		v.OpenedAt = in.OpenedAt.Value
	}
	if in.ClosedAt.Set {
		v.ClosedAt = in.ClosedAt.Value
	}
	if in.IdempotencyKey.Set {
		if in.IdempotencyKey.Value == nil {
			return hrerrors.Invalid("idempotency_key", "must not be null")
		}
		v.IdempotencyKey = *in.IdempotencyKey.Value
	}
	return nil
}
func specRequisition() operation.Specification[entity.Requisition, string] {
	return operation.Specification[entity.Requisition, string]{
		Domain: "recruitment", Table: "hr_requisitions",
		Key:         func(v *entity.Requisition) string { return v.ID },
		KeyString:   func(key string) string { return fmt.Sprint(key) },
		ValidateKey: func(key string) error { return entity.ValidateUUID("id", key) },
		AggregateID: func(v *entity.Requisition) string { return v.ID },
		Initialize: func(v *entity.Requisition, actor operation.Actor, now time.Time, id string) {
			v.ID = id
			v.CreatedByPrincipalID = &actor.PrincipalID
		},
		Touch: func(v *entity.Requisition, actor operation.Actor, now time.Time) {},
		Attributes: func(v *entity.Requisition) map[string]string {
			a := map[string]string{}
			a["id"] = fmt.Sprint(v.ID)
			if v.WorkforcePlanLineID != nil {
				a["workforce_plan_line_id"] = fmt.Sprint(*v.WorkforcePlanLineID)
			}
			if v.PositionID != nil {
				a["position_id"] = fmt.Sprint(*v.PositionID)
			}
			a["job_id"] = fmt.Sprint(v.JobID)
			a["organization_unit_reference_id"] = fmt.Sprint(v.OrganizationUnitReferenceID)
			a["hiring_manager_principal_id"] = fmt.Sprint(v.HiringManagerPrincipalID)
			if v.CreatedByPrincipalID != nil {
				a["created_by_principal_id"] = *v.CreatedByPrincipalID
			}
			if v.RecruiterPrincipalID != nil {
				a["recruiter_principal_id"] = fmt.Sprint(*v.RecruiterPrincipalID)
			}
			a["status"] = fmt.Sprint(v.Status)
			if v.BudgetReferenceID != nil {
				a["budget_reference_id"] = fmt.Sprint(*v.BudgetReferenceID)
			}
			if v.WorkLocationReferenceID != nil {
				a["work_location_reference_id"] = fmt.Sprint(*v.WorkLocationReferenceID)
			}
			if v.ApprovedByPrincipalID != nil {
				a["approved_by_principal_id"] = fmt.Sprint(*v.ApprovedByPrincipalID)
			}
			return a
		},
		Guard: func(action string, old, next *entity.Requisition) error {
			if action == "create" && (next.CreatedByPrincipalID == nil || next.SubmittedAt != nil || next.SubmittedByPrincipalID != nil) {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "update" && (!reflect.DeepEqual(old.CreatedByPrincipalID, next.CreatedByPrincipalID) ||
				!reflect.DeepEqual(old.SubmittedAt, next.SubmittedAt) ||
				!reflect.DeepEqual(old.SubmittedByPrincipalID, next.SubmittedByPrincipalID)) {
				return hrerrors.ErrWorkflowRequired
			}
			if action != "create" && old.Status != "draft" {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "update" && !reflect.DeepEqual(old.Status, next.Status) {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "create" && string(next.Status) != "draft" {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "create" && next.ApprovedAt != nil {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "update" && !reflect.DeepEqual(old.ApprovedAt, next.ApprovedAt) {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "create" && next.ApprovedByPrincipalID != nil {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "update" && !reflect.DeepEqual(old.ApprovedByPrincipalID, next.ApprovedByPrincipalID) {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "create" && next.OpenedAt != nil {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "update" && !reflect.DeepEqual(old.OpenedAt, next.OpenedAt) {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "create" && next.ClosedAt != nil {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "update" && !reflect.DeepEqual(old.ClosedAt, next.ClosedAt) {
				return hrerrors.ErrWorkflowRequired
			}
			return nil
		},
	}
}
func createCandidate(in CreateCandidateInput) (*entity.Candidate, error) {
	v := &entity.Candidate{}
	v.CandidateNumber = in.CandidateNumber
	v.PersonPartyReferenceID = in.PersonPartyReferenceID
	v.PreferredName = in.PreferredName
	v.ContactReferenceID = in.ContactReferenceID
	v.SourceType = in.SourceType
	v.SourceReference = in.SourceReference
	v.PrivacyNoticeVersion = in.PrivacyNoticeVersion
	v.NoticeAcknowledgedAt = in.NoticeAcknowledgedAt
	v.DataRetentionUntil = in.DataRetentionUntil
	if in.CandidateStatus != nil {
		v.CandidateStatus = *in.CandidateStatus
	} else {
		v.CandidateStatus = "active"
	}
	return v, nil
}
func patchCandidate(v *entity.Candidate, in UpdateCandidateInput) error {
	if in.CandidateNumber.Set {
		if in.CandidateNumber.Value == nil {
			return hrerrors.Invalid("candidate_number", "must not be null")
		}
		v.CandidateNumber = *in.CandidateNumber.Value
	}
	if in.PersonPartyReferenceID.Set {
		v.PersonPartyReferenceID = in.PersonPartyReferenceID.Value
	}
	if in.PreferredName.Set {
		v.PreferredName = in.PreferredName.Value
	}
	if in.ContactReferenceID.Set {
		v.ContactReferenceID = in.ContactReferenceID.Value
	}
	if in.SourceType.Set {
		v.SourceType = in.SourceType.Value
	}
	if in.SourceReference.Set {
		v.SourceReference = in.SourceReference.Value
	}
	if in.PrivacyNoticeVersion.Set {
		if in.PrivacyNoticeVersion.Value == nil {
			return hrerrors.Invalid("privacy_notice_version", "must not be null")
		}
		v.PrivacyNoticeVersion = *in.PrivacyNoticeVersion.Value
	}
	if in.NoticeAcknowledgedAt.Set {
		v.NoticeAcknowledgedAt = in.NoticeAcknowledgedAt.Value
	}
	if in.DataRetentionUntil.Set {
		v.DataRetentionUntil = in.DataRetentionUntil.Value
	}
	if in.CandidateStatus.Set {
		if in.CandidateStatus.Value == nil {
			return hrerrors.Invalid("candidate_status", "must not be null")
		}
		v.CandidateStatus = *in.CandidateStatus.Value
	}
	return nil
}
func specCandidate() operation.Specification[entity.Candidate, string] {
	return operation.Specification[entity.Candidate, string]{
		Domain: "recruitment", Table: "hr_candidates",
		Key:         func(v *entity.Candidate) string { return v.ID },
		KeyString:   func(key string) string { return fmt.Sprint(key) },
		ValidateKey: func(key string) error { return entity.ValidateUUID("id", key) },
		AggregateID: func(v *entity.Candidate) string { return v.ID },
		Initialize: func(v *entity.Candidate, actor operation.Actor, now time.Time, id string) {
			v.ID = id
			v.CreatedAt = now
			v.CreatedByPrincipalID = &actor.PrincipalID
		},
		Touch: func(v *entity.Candidate, actor operation.Actor, now time.Time) {},
		Attributes: func(v *entity.Candidate) map[string]string {
			a := map[string]string{}
			a["id"] = fmt.Sprint(v.ID)
			if v.PersonPartyReferenceID != nil {
				a["person_party_reference_id"] = fmt.Sprint(*v.PersonPartyReferenceID)
			}
			if v.ContactReferenceID != nil {
				a["contact_reference_id"] = fmt.Sprint(*v.ContactReferenceID)
			}
			a["candidate_status"] = fmt.Sprint(v.CandidateStatus)
			if v.CreatedByPrincipalID != nil {
				a["created_by_principal_id"] = fmt.Sprint(*v.CreatedByPrincipalID)
			}
			return a
		},
		Guard: func(action string, old, next *entity.Candidate) error {
			if action == "update" && !reflect.DeepEqual(old.CandidateStatus, next.CandidateStatus) {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "create" && string(next.CandidateStatus) != "active" {
				return hrerrors.ErrWorkflowRequired
			}
			if action != "create" {
				switch string(old.CandidateStatus) {
				case "approved", "posted", "paid", "finalized", "applied", "closed", "archived", "terminated", "deceased", "completed", "accepted", "issued", "signed", "hired", "processed", "rejected", "cancelled", "withdrawn", "superseded":
					return hrerrors.ErrWorkflowRequired
				}
			}
			return nil
		},
	}
}
func createApplication(in CreateApplicationInput) (*entity.Application, error) {
	v := &entity.Application{}
	v.CandidateID = in.CandidateID
	v.RequisitionID = in.RequisitionID
	v.ApplicationNumber = in.ApplicationNumber
	if in.Status != nil {
		v.Status = *in.Status
	} else {
		v.Status = "received"
	}
	v.CurrentStageCode = in.CurrentStageCode
	v.AssignedRecruiterPrincipalID = in.AssignedRecruiterPrincipalID
	v.ScreeningScore = in.ScreeningScore
	v.ScorecardVersion = in.ScorecardVersion
	v.DispositionReasonCode = in.DispositionReasonCode
	v.DecisionAt = in.DecisionAt
	v.DecisionByPrincipalID = in.DecisionByPrincipalID
	v.HiredWorkerID = in.HiredWorkerID
	v.SourcePostingID = in.SourcePostingID
	v.IdempotencyKey = in.IdempotencyKey
	return v, nil
}
func patchApplication(v *entity.Application, in UpdateApplicationInput) error {
	if in.CandidateID.Set {
		if in.CandidateID.Value == nil {
			return hrerrors.Invalid("candidate_id", "must not be null")
		}
		v.CandidateID = *in.CandidateID.Value
	}
	if in.RequisitionID.Set {
		if in.RequisitionID.Value == nil {
			return hrerrors.Invalid("requisition_id", "must not be null")
		}
		v.RequisitionID = *in.RequisitionID.Value
	}
	if in.ApplicationNumber.Set {
		if in.ApplicationNumber.Value == nil {
			return hrerrors.Invalid("application_number", "must not be null")
		}
		v.ApplicationNumber = *in.ApplicationNumber.Value
	}
	if in.Status.Set {
		if in.Status.Value == nil {
			return hrerrors.Invalid("status", "must not be null")
		}
		v.Status = *in.Status.Value
	}
	if in.CurrentStageCode.Set {
		v.CurrentStageCode = in.CurrentStageCode.Value
	}
	if in.AssignedRecruiterPrincipalID.Set {
		v.AssignedRecruiterPrincipalID = in.AssignedRecruiterPrincipalID.Value
	}
	if in.ScreeningScore.Set {
		v.ScreeningScore = in.ScreeningScore.Value
	}
	if in.ScorecardVersion.Set {
		v.ScorecardVersion = in.ScorecardVersion.Value
	}
	if in.DispositionReasonCode.Set {
		v.DispositionReasonCode = in.DispositionReasonCode.Value
	}
	if in.DecisionAt.Set {
		v.DecisionAt = in.DecisionAt.Value
	}
	if in.DecisionByPrincipalID.Set {
		v.DecisionByPrincipalID = in.DecisionByPrincipalID.Value
	}
	if in.HiredWorkerID.Set {
		v.HiredWorkerID = in.HiredWorkerID.Value
	}
	if in.SourcePostingID.Set {
		v.SourcePostingID = in.SourcePostingID.Value
	}
	if in.IdempotencyKey.Set {
		if in.IdempotencyKey.Value == nil {
			return hrerrors.Invalid("idempotency_key", "must not be null")
		}
		v.IdempotencyKey = *in.IdempotencyKey.Value
	}
	return nil
}
func specApplication() operation.Specification[entity.Application, string] {
	return operation.Specification[entity.Application, string]{
		Domain: "recruitment", Table: "hr_applications",
		Key:         func(v *entity.Application) string { return v.ID },
		KeyString:   func(key string) string { return fmt.Sprint(key) },
		ValidateKey: func(key string) error { return entity.ValidateUUID("id", key) },
		AggregateID: func(v *entity.Application) string { return v.ID },
		Initialize: func(v *entity.Application, actor operation.Actor, now time.Time, id string) {
			v.ID = id
			v.AppliedAt = now
		},
		Touch: func(v *entity.Application, actor operation.Actor, now time.Time) {},
		Attributes: func(v *entity.Application) map[string]string {
			a := map[string]string{}
			a["id"] = fmt.Sprint(v.ID)
			a["candidate_id"] = fmt.Sprint(v.CandidateID)
			a["requisition_id"] = fmt.Sprint(v.RequisitionID)
			a["status"] = fmt.Sprint(v.Status)
			if v.AssignedRecruiterPrincipalID != nil {
				a["assigned_recruiter_principal_id"] = fmt.Sprint(*v.AssignedRecruiterPrincipalID)
			}
			if v.DecisionByPrincipalID != nil {
				a["decision_by_principal_id"] = fmt.Sprint(*v.DecisionByPrincipalID)
			}
			if v.HiredWorkerID != nil {
				a["hired_worker_id"] = fmt.Sprint(*v.HiredWorkerID)
			}
			if v.SourcePostingID != nil {
				a["source_posting_id"] = fmt.Sprint(*v.SourcePostingID)
			}
			return a
		},
		Guard: func(action string, old, next *entity.Application) error {
			if action == "update" && !reflect.DeepEqual(old.Status, next.Status) {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "create" && string(next.Status) != "received" {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "update" && !reflect.DeepEqual(old.AppliedAt, next.AppliedAt) {
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
func createApplicationStageHistory(in CreateApplicationStageHistoryInput) (*entity.ApplicationStageHistory, error) {
	v := &entity.ApplicationStageHistory{}
	v.ApplicationID = in.ApplicationID
	v.PreviousStatus = in.PreviousStatus
	v.NewStatus = in.NewStatus
	v.StageCode = in.StageCode
	v.ChangedByPrincipalID = in.ChangedByPrincipalID
	v.ReasonCode = in.ReasonCode
	return v, nil
}
func patchApplicationStageHistory(v *entity.ApplicationStageHistory, in UpdateApplicationStageHistoryInput) error {
	if in.ApplicationID.Set {
		if in.ApplicationID.Value == nil {
			return hrerrors.Invalid("application_id", "must not be null")
		}
		v.ApplicationID = *in.ApplicationID.Value
	}
	if in.PreviousStatus.Set {
		v.PreviousStatus = in.PreviousStatus.Value
	}
	if in.NewStatus.Set {
		if in.NewStatus.Value == nil {
			return hrerrors.Invalid("new_status", "must not be null")
		}
		v.NewStatus = *in.NewStatus.Value
	}
	if in.StageCode.Set {
		v.StageCode = in.StageCode.Value
	}
	if in.ChangedByPrincipalID.Set {
		v.ChangedByPrincipalID = in.ChangedByPrincipalID.Value
	}
	if in.ReasonCode.Set {
		v.ReasonCode = in.ReasonCode.Value
	}
	return nil
}
func specApplicationStageHistory() operation.Specification[entity.ApplicationStageHistory, int64] {
	return operation.Specification[entity.ApplicationStageHistory, int64]{
		Domain: "recruitment", Table: "hr_application_stage_history",
		Key:       func(v *entity.ApplicationStageHistory) int64 { return v.ID },
		KeyString: func(key int64) string { return fmt.Sprint(key) },
		ValidateKey: func(key int64) error {
			return func() error {
				if key <= 0 {
					return hrerrors.ErrInvalidInput
				}
				return nil
			}()
		},
		AggregateID: func(v *entity.ApplicationStageHistory) string { return v.ApplicationID },
		Initialize: func(v *entity.ApplicationStageHistory, actor operation.Actor, now time.Time, id string) {
			v.ChangedAt = now
		},
		Touch: func(v *entity.ApplicationStageHistory, actor operation.Actor, now time.Time) {},
		Attributes: func(v *entity.ApplicationStageHistory) map[string]string {
			a := map[string]string{}
			a["id"] = fmt.Sprint(v.ID)
			a["application_id"] = fmt.Sprint(v.ApplicationID)
			if v.PreviousStatus != nil {
				a["previous_status"] = fmt.Sprint(*v.PreviousStatus)
			}
			a["new_status"] = fmt.Sprint(v.NewStatus)
			if v.ChangedByPrincipalID != nil {
				a["changed_by_principal_id"] = fmt.Sprint(*v.ChangedByPrincipalID)
			}
			return a
		},
		Guard: func(action string, old, next *entity.ApplicationStageHistory) error {
			if action != "create" {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "update" && !reflect.DeepEqual(old.PreviousStatus, next.PreviousStatus) {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "update" && !reflect.DeepEqual(old.NewStatus, next.NewStatus) {
				return hrerrors.ErrWorkflowRequired
			}
			if action != "create" {
				switch string(old.NewStatus) {
				case "approved", "posted", "paid", "finalized", "applied", "closed", "archived", "terminated", "deceased", "completed", "accepted", "issued", "signed", "hired", "processed", "rejected", "cancelled", "withdrawn", "superseded":
					return hrerrors.ErrWorkflowRequired
				}
			}
			return nil
		},
	}
}
func createInterview(in CreateInterviewInput) (*entity.Interview, error) {
	v := &entity.Interview{}
	v.ApplicationID = in.ApplicationID
	v.InterviewRound = in.InterviewRound
	v.InterviewType = in.InterviewType
	v.ScheduledStart = in.ScheduledStart
	v.ScheduledEnd = in.ScheduledEnd
	v.LocationReferenceID = in.LocationReferenceID
	v.MeetingReferenceID = in.MeetingReferenceID
	if in.Status != nil {
		v.Status = *in.Status
	} else {
		v.Status = "planned"
	}
	v.CandidateTimezone = in.CandidateTimezone
	return v, nil
}
func patchInterview(v *entity.Interview, in UpdateInterviewInput) error {
	if in.ApplicationID.Set {
		if in.ApplicationID.Value == nil {
			return hrerrors.Invalid("application_id", "must not be null")
		}
		v.ApplicationID = *in.ApplicationID.Value
	}
	if in.InterviewRound.Set {
		if in.InterviewRound.Value == nil {
			return hrerrors.Invalid("interview_round", "must not be null")
		}
		v.InterviewRound = *in.InterviewRound.Value
	}
	if in.InterviewType.Set {
		if in.InterviewType.Value == nil {
			return hrerrors.Invalid("interview_type", "must not be null")
		}
		v.InterviewType = *in.InterviewType.Value
	}
	if in.ScheduledStart.Set {
		v.ScheduledStart = in.ScheduledStart.Value
	}
	if in.ScheduledEnd.Set {
		v.ScheduledEnd = in.ScheduledEnd.Value
	}
	if in.LocationReferenceID.Set {
		v.LocationReferenceID = in.LocationReferenceID.Value
	}
	if in.MeetingReferenceID.Set {
		v.MeetingReferenceID = in.MeetingReferenceID.Value
	}
	if in.Status.Set {
		if in.Status.Value == nil {
			return hrerrors.Invalid("status", "must not be null")
		}
		v.Status = *in.Status.Value
	}
	if in.CandidateTimezone.Set {
		v.CandidateTimezone = in.CandidateTimezone.Value
	}
	return nil
}
func specInterview() operation.Specification[entity.Interview, string] {
	return operation.Specification[entity.Interview, string]{
		Domain: "recruitment", Table: "hr_interviews",
		Key:         func(v *entity.Interview) string { return v.ID },
		KeyString:   func(key string) string { return fmt.Sprint(key) },
		ValidateKey: func(key string) error { return entity.ValidateUUID("id", key) },
		AggregateID: func(v *entity.Interview) string { return v.ID },
		Initialize: func(v *entity.Interview, actor operation.Actor, now time.Time, id string) {
			v.ID = id
			v.CreatedByPrincipalID = &actor.PrincipalID
		},
		Touch: func(v *entity.Interview, actor operation.Actor, now time.Time) {},
		Attributes: func(v *entity.Interview) map[string]string {
			a := map[string]string{}
			a["id"] = fmt.Sprint(v.ID)
			a["application_id"] = fmt.Sprint(v.ApplicationID)
			if v.LocationReferenceID != nil {
				a["location_reference_id"] = fmt.Sprint(*v.LocationReferenceID)
			}
			if v.MeetingReferenceID != nil {
				a["meeting_reference_id"] = fmt.Sprint(*v.MeetingReferenceID)
			}
			a["status"] = fmt.Sprint(v.Status)
			if v.CreatedByPrincipalID != nil {
				a["created_by_principal_id"] = fmt.Sprint(*v.CreatedByPrincipalID)
			}
			return a
		},
		Guard: func(action string, old, next *entity.Interview) error {
			if action == "update" && !reflect.DeepEqual(old.Status, next.Status) {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "create" && string(next.Status) != "planned" {
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
func createInterviewPanelist(in CreateInterviewPanelistInput) (*entity.InterviewPanelist, error) {
	v := &entity.InterviewPanelist{}
	v.InterviewID = in.InterviewID
	v.InterviewerPrincipalID = in.InterviewerPrincipalID
	v.InterviewerRole = in.InterviewerRole
	v.ScorecardTemplateReferenceID = in.ScorecardTemplateReferenceID
	v.FeedbackDueAt = in.FeedbackDueAt
	return v, nil
}
func patchInterviewPanelist(v *entity.InterviewPanelist, in UpdateInterviewPanelistInput) error {
	if in.InterviewID.Set {
		if in.InterviewID.Value == nil {
			return hrerrors.Invalid("interview_id", "must not be null")
		}
		v.InterviewID = *in.InterviewID.Value
	}
	if in.InterviewerPrincipalID.Set {
		if in.InterviewerPrincipalID.Value == nil {
			return hrerrors.Invalid("interviewer_principal_id", "must not be null")
		}
		v.InterviewerPrincipalID = *in.InterviewerPrincipalID.Value
	}
	if in.InterviewerRole.Set {
		v.InterviewerRole = in.InterviewerRole.Value
	}
	if in.ScorecardTemplateReferenceID.Set {
		v.ScorecardTemplateReferenceID = in.ScorecardTemplateReferenceID.Value
	}
	if in.FeedbackDueAt.Set {
		v.FeedbackDueAt = in.FeedbackDueAt.Value
	}
	return nil
}
func specInterviewPanelist() operation.Specification[entity.InterviewPanelist, string] {
	return operation.Specification[entity.InterviewPanelist, string]{
		Domain: "recruitment", Table: "hr_interview_panelists",
		Key:         func(v *entity.InterviewPanelist) string { return v.ID },
		KeyString:   func(key string) string { return fmt.Sprint(key) },
		ValidateKey: func(key string) error { return entity.ValidateUUID("id", key) },
		AggregateID: func(v *entity.InterviewPanelist) string { return v.ID },
		Initialize: func(v *entity.InterviewPanelist, actor operation.Actor, now time.Time, id string) {
			v.ID = id
		},
		Touch: func(v *entity.InterviewPanelist, actor operation.Actor, now time.Time) {},
		Attributes: func(v *entity.InterviewPanelist) map[string]string {
			a := map[string]string{}
			a["id"] = fmt.Sprint(v.ID)
			a["interview_id"] = fmt.Sprint(v.InterviewID)
			a["interviewer_principal_id"] = fmt.Sprint(v.InterviewerPrincipalID)
			if v.ScorecardTemplateReferenceID != nil {
				a["scorecard_template_reference_id"] = fmt.Sprint(*v.ScorecardTemplateReferenceID)
			}
			return a
		},
		Guard: func(action string, old, next *entity.InterviewPanelist) error {
			return nil
		},
	}
}
func createInterviewFeedback(in CreateInterviewFeedbackInput) (*entity.InterviewFeedback, error) {
	v := &entity.InterviewFeedback{}
	v.InterviewID = in.InterviewID
	v.InterviewerPrincipalID = in.InterviewerPrincipalID
	if in.FeedbackVersion != nil {
		v.FeedbackVersion = *in.FeedbackVersion
	} else {
		v.FeedbackVersion = 1
	}
	v.OverallRecommendation = in.OverallRecommendation
	v.CompetencyScores = in.CompetencyScores
	v.EvidenceNotes = in.EvidenceNotes
	v.ContentDigest = in.ContentDigest
	if in.VisibilityClass != nil {
		v.VisibilityClass = *in.VisibilityClass
	} else {
		v.VisibilityClass = "restricted"
	}
	return v, nil
}
func patchInterviewFeedback(v *entity.InterviewFeedback, in UpdateInterviewFeedbackInput) error {
	if in.InterviewID.Set {
		if in.InterviewID.Value == nil {
			return hrerrors.Invalid("interview_id", "must not be null")
		}
		v.InterviewID = *in.InterviewID.Value
	}
	if in.InterviewerPrincipalID.Set {
		if in.InterviewerPrincipalID.Value == nil {
			return hrerrors.Invalid("interviewer_principal_id", "must not be null")
		}
		v.InterviewerPrincipalID = *in.InterviewerPrincipalID.Value
	}
	if in.FeedbackVersion.Set {
		if in.FeedbackVersion.Value == nil {
			return hrerrors.Invalid("feedback_version", "must not be null")
		}
		v.FeedbackVersion = *in.FeedbackVersion.Value
	}
	if in.OverallRecommendation.Set {
		if in.OverallRecommendation.Value == nil {
			return hrerrors.Invalid("overall_recommendation", "must not be null")
		}
		v.OverallRecommendation = *in.OverallRecommendation.Value
	}
	if in.CompetencyScores.Set {
		if in.CompetencyScores.Value == nil {
			v.CompetencyScores = nil
		} else {
			v.CompetencyScores = *in.CompetencyScores.Value
		}
	}
	if in.EvidenceNotes.Set {
		v.EvidenceNotes = in.EvidenceNotes.Value
	}
	if in.ContentDigest.Set {
		if in.ContentDigest.Value == nil {
			v.ContentDigest = nil
		} else {
			v.ContentDigest = *in.ContentDigest.Value
		}
	}
	if in.VisibilityClass.Set {
		if in.VisibilityClass.Value == nil {
			return hrerrors.Invalid("visibility_class", "must not be null")
		}
		v.VisibilityClass = *in.VisibilityClass.Value
	}
	return nil
}
func specInterviewFeedback() operation.Specification[entity.InterviewFeedback, string] {
	return operation.Specification[entity.InterviewFeedback, string]{
		Domain: "recruitment", Table: "hr_interview_feedback",
		Key:         func(v *entity.InterviewFeedback) string { return v.ID },
		KeyString:   func(key string) string { return fmt.Sprint(key) },
		ValidateKey: func(key string) error { return entity.ValidateUUID("id", key) },
		AggregateID: func(v *entity.InterviewFeedback) string { return v.ID },
		Initialize: func(v *entity.InterviewFeedback, actor operation.Actor, now time.Time, id string) {
			v.ID = id
			v.SubmittedAt = now
		},
		Touch: func(v *entity.InterviewFeedback, actor operation.Actor, now time.Time) {},
		Attributes: func(v *entity.InterviewFeedback) map[string]string {
			a := map[string]string{}
			a["id"] = fmt.Sprint(v.ID)
			a["interview_id"] = fmt.Sprint(v.InterviewID)
			a["interviewer_principal_id"] = fmt.Sprint(v.InterviewerPrincipalID)
			return a
		},
		Guard: func(action string, old, next *entity.InterviewFeedback) error {
			return nil
		},
	}
}
func createOffer(in CreateOfferInput) (*entity.Offer, error) {
	v := &entity.Offer{}
	v.ApplicationID = in.ApplicationID
	v.OfferNumber = in.OfferNumber
	if in.OfferVersion != nil {
		v.OfferVersion = *in.OfferVersion
	} else {
		v.OfferVersion = 1
	}
	if in.OfferStatus != nil {
		v.OfferStatus = *in.OfferStatus
	} else {
		v.OfferStatus = "draft"
	}
	v.ProposedJobID = in.ProposedJobID
	v.ProposedLegalEntityReferenceID = in.ProposedLegalEntityReferenceID
	v.ProposedStartDate = in.ProposedStartDate
	v.ProposedContractType = in.ProposedContractType
	v.ProposedCompensationDocumentVersionReferenceID = in.ProposedCompensationDocumentVersionReferenceID
	v.CurrencyCode = in.CurrencyCode
	v.OfferDocumentReferenceID = in.OfferDocumentReferenceID
	v.OfferDocumentVersionReferenceID = in.OfferDocumentVersionReferenceID
	v.OfferContentDigest = in.OfferContentDigest
	v.IssuedAt = in.IssuedAt
	v.ExpiresAt = in.ExpiresAt
	v.CandidateResponse = in.CandidateResponse
	v.RespondedAt = in.RespondedAt
	v.ApprovedByPrincipalID = in.ApprovedByPrincipalID
	v.ApprovedAt = in.ApprovedAt
	return v, nil
}
func patchOffer(v *entity.Offer, in UpdateOfferInput) error {
	if in.ApplicationID.Set {
		if in.ApplicationID.Value == nil {
			return hrerrors.Invalid("application_id", "must not be null")
		}
		v.ApplicationID = *in.ApplicationID.Value
	}
	if in.OfferNumber.Set {
		if in.OfferNumber.Value == nil {
			return hrerrors.Invalid("offer_number", "must not be null")
		}
		v.OfferNumber = *in.OfferNumber.Value
	}
	if in.OfferVersion.Set {
		if in.OfferVersion.Value == nil {
			return hrerrors.Invalid("offer_version", "must not be null")
		}
		v.OfferVersion = *in.OfferVersion.Value
	}
	if in.OfferStatus.Set {
		if in.OfferStatus.Value == nil {
			return hrerrors.Invalid("offer_status", "must not be null")
		}
		v.OfferStatus = *in.OfferStatus.Value
	}
	if in.ProposedJobID.Set {
		if in.ProposedJobID.Value == nil {
			return hrerrors.Invalid("proposed_job_id", "must not be null")
		}
		v.ProposedJobID = *in.ProposedJobID.Value
	}
	if in.ProposedLegalEntityReferenceID.Set {
		if in.ProposedLegalEntityReferenceID.Value == nil {
			return hrerrors.Invalid("proposed_legal_entity_reference_id", "must not be null")
		}
		v.ProposedLegalEntityReferenceID = *in.ProposedLegalEntityReferenceID.Value
	}
	if in.ProposedStartDate.Set {
		v.ProposedStartDate = in.ProposedStartDate.Value
	}
	if in.ProposedContractType.Set {
		if in.ProposedContractType.Value == nil {
			return hrerrors.Invalid("proposed_contract_type", "must not be null")
		}
		v.ProposedContractType = *in.ProposedContractType.Value
	}
	if in.ProposedCompensationDocumentVersionReferenceID.Set {
		v.ProposedCompensationDocumentVersionReferenceID = in.ProposedCompensationDocumentVersionReferenceID.Value
	}
	if in.CurrencyCode.Set {
		v.CurrencyCode = in.CurrencyCode.Value
	}
	if in.OfferDocumentReferenceID.Set {
		v.OfferDocumentReferenceID = in.OfferDocumentReferenceID.Value
	}
	if in.OfferDocumentVersionReferenceID.Set {
		v.OfferDocumentVersionReferenceID = in.OfferDocumentVersionReferenceID.Value
	}
	if in.OfferContentDigest.Set {
		if in.OfferContentDigest.Value == nil {
			v.OfferContentDigest = nil
		} else {
			v.OfferContentDigest = *in.OfferContentDigest.Value
		}
	}
	if in.IssuedAt.Set {
		v.IssuedAt = in.IssuedAt.Value
	}
	if in.ExpiresAt.Set {
		v.ExpiresAt = in.ExpiresAt.Value
	}
	if in.CandidateResponse.Set {
		v.CandidateResponse = in.CandidateResponse.Value
	}
	if in.RespondedAt.Set {
		v.RespondedAt = in.RespondedAt.Value
	}
	if in.ApprovedByPrincipalID.Set {
		v.ApprovedByPrincipalID = in.ApprovedByPrincipalID.Value
	}
	if in.ApprovedAt.Set {
		v.ApprovedAt = in.ApprovedAt.Value
	}
	return nil
}
func specOffer() operation.Specification[entity.Offer, string] {
	return operation.Specification[entity.Offer, string]{
		Domain: "recruitment", Table: "hr_offers",
		Key:         func(v *entity.Offer) string { return v.ID },
		KeyString:   func(key string) string { return fmt.Sprint(key) },
		ValidateKey: func(key string) error { return entity.ValidateUUID("id", key) },
		AggregateID: func(v *entity.Offer) string { return v.ID },
		Initialize: func(v *entity.Offer, actor operation.Actor, now time.Time, id string) {
			v.ID = id
		},
		Touch: func(v *entity.Offer, actor operation.Actor, now time.Time) {},
		Attributes: func(v *entity.Offer) map[string]string {
			a := map[string]string{}
			a["id"] = fmt.Sprint(v.ID)
			a["application_id"] = fmt.Sprint(v.ApplicationID)
			a["offer_status"] = fmt.Sprint(v.OfferStatus)
			a["proposed_job_id"] = fmt.Sprint(v.ProposedJobID)
			a["proposed_legal_entity_reference_id"] = fmt.Sprint(v.ProposedLegalEntityReferenceID)
			if v.ProposedCompensationDocumentVersionReferenceID != nil {
				a["proposed_compensation_document_version_reference_id"] = fmt.Sprint(*v.ProposedCompensationDocumentVersionReferenceID)
			}
			if v.OfferDocumentReferenceID != nil {
				a["offer_document_reference_id"] = fmt.Sprint(*v.OfferDocumentReferenceID)
			}
			if v.OfferDocumentVersionReferenceID != nil {
				a["offer_document_version_reference_id"] = fmt.Sprint(*v.OfferDocumentVersionReferenceID)
			}
			if v.ApprovedByPrincipalID != nil {
				a["approved_by_principal_id"] = fmt.Sprint(*v.ApprovedByPrincipalID)
			}
			return a
		},
		Guard: func(action string, old, next *entity.Offer) error {
			if action == "update" && !reflect.DeepEqual(old.OfferStatus, next.OfferStatus) {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "create" && string(next.OfferStatus) != "draft" {
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
				switch string(old.OfferStatus) {
				case "approved", "posted", "paid", "finalized", "applied", "closed", "archived", "terminated", "deceased", "completed", "accepted", "issued", "signed", "hired", "processed", "rejected", "cancelled", "withdrawn", "superseded":
					return hrerrors.ErrWorkflowRequired
				}
			}
			return nil
		},
	}
}
func createJobPosting(in CreateJobPostingInput) (*entity.JobPosting, error) {
	v := &entity.JobPosting{}
	v.RequisitionID = in.RequisitionID
	v.ChannelCode = in.ChannelCode
	v.PostingVersion = in.PostingVersion
	v.ExternalPostingReference = in.ExternalPostingReference
	v.ContentDocumentVersionReferenceID = in.ContentDocumentVersionReferenceID
	v.ContentDigest = in.ContentDigest
	if in.Status != nil {
		v.Status = *in.Status
	} else {
		v.Status = "draft"
	}
	v.PublishedAt = in.PublishedAt
	v.ClosesAt = in.ClosesAt
	return v, nil
}
func patchJobPosting(v *entity.JobPosting, in UpdateJobPostingInput) error {
	if in.RequisitionID.Set {
		if in.RequisitionID.Value == nil {
			return hrerrors.Invalid("requisition_id", "must not be null")
		}
		v.RequisitionID = *in.RequisitionID.Value
	}
	if in.ChannelCode.Set {
		if in.ChannelCode.Value == nil {
			return hrerrors.Invalid("channel_code", "must not be null")
		}
		v.ChannelCode = *in.ChannelCode.Value
	}
	if in.PostingVersion.Set {
		if in.PostingVersion.Value == nil {
			return hrerrors.Invalid("posting_version", "must not be null")
		}
		v.PostingVersion = *in.PostingVersion.Value
	}
	if in.ExternalPostingReference.Set {
		v.ExternalPostingReference = in.ExternalPostingReference.Value
	}
	if in.ContentDocumentVersionReferenceID.Set {
		if in.ContentDocumentVersionReferenceID.Value == nil {
			return hrerrors.Invalid("content_document_version_reference_id", "must not be null")
		}
		v.ContentDocumentVersionReferenceID = *in.ContentDocumentVersionReferenceID.Value
	}
	if in.ContentDigest.Set {
		if in.ContentDigest.Value == nil {
			v.ContentDigest = nil
		} else {
			v.ContentDigest = *in.ContentDigest.Value
		}
	}
	if in.Status.Set {
		if in.Status.Value == nil {
			return hrerrors.Invalid("status", "must not be null")
		}
		v.Status = *in.Status.Value
	}
	if in.PublishedAt.Set {
		v.PublishedAt = in.PublishedAt.Value
	}
	if in.ClosesAt.Set {
		v.ClosesAt = in.ClosesAt.Value
	}
	return nil
}
func specJobPosting() operation.Specification[entity.JobPosting, string] {
	return operation.Specification[entity.JobPosting, string]{
		Domain: "recruitment", Table: "hr_job_postings",
		Key:         func(v *entity.JobPosting) string { return v.ID },
		KeyString:   func(key string) string { return fmt.Sprint(key) },
		ValidateKey: func(key string) error { return entity.ValidateUUID("id", key) },
		AggregateID: func(v *entity.JobPosting) string { return v.ID },
		Initialize: func(v *entity.JobPosting, actor operation.Actor, now time.Time, id string) {
			v.ID = id
			v.CreatedByPrincipalID = actor.PrincipalID
		},
		Touch: func(v *entity.JobPosting, actor operation.Actor, now time.Time) {},
		Attributes: func(v *entity.JobPosting) map[string]string {
			a := map[string]string{}
			a["id"] = fmt.Sprint(v.ID)
			a["requisition_id"] = fmt.Sprint(v.RequisitionID)
			a["content_document_version_reference_id"] = fmt.Sprint(v.ContentDocumentVersionReferenceID)
			a["status"] = fmt.Sprint(v.Status)
			a["created_by_principal_id"] = fmt.Sprint(v.CreatedByPrincipalID)
			return a
		},
		Guard: func(action string, old, next *entity.JobPosting) error {
			if action == "update" && !reflect.DeepEqual(old.Status, next.Status) {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "create" && string(next.Status) != "draft" {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "create" && next.PublishedAt != nil {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "update" && !reflect.DeepEqual(old.PublishedAt, next.PublishedAt) {
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
func createApplicationAssessment(in CreateApplicationAssessmentInput) (*entity.ApplicationAssessment, error) {
	v := &entity.ApplicationAssessment{}
	v.ApplicationID = in.ApplicationID
	v.AssessmentCode = in.AssessmentCode
	if in.AttemptNumber != nil {
		v.AttemptNumber = *in.AttemptNumber
	} else {
		v.AttemptNumber = 1
	}
	v.RubricDocumentVersionReferenceID = in.RubricDocumentVersionReferenceID
	v.RubricDigest = in.RubricDigest
	if in.Status != nil {
		v.Status = *in.Status
	} else {
		v.Status = "planned"
	}
	v.AssessorPrincipalID = in.AssessorPrincipalID
	v.ScheduledAt = in.ScheduledAt
	v.CompletedAt = in.CompletedAt
	v.NormalizedScore = in.NormalizedScore
	v.RecommendationCode = in.RecommendationCode
	v.EvidenceDocumentVersionReferenceID = in.EvidenceDocumentVersionReferenceID
	v.EvidenceDigest = in.EvidenceDigest
	return v, nil
}
func patchApplicationAssessment(v *entity.ApplicationAssessment, in UpdateApplicationAssessmentInput) error {
	if in.ApplicationID.Set {
		if in.ApplicationID.Value == nil {
			return hrerrors.Invalid("application_id", "must not be null")
		}
		v.ApplicationID = *in.ApplicationID.Value
	}
	if in.AssessmentCode.Set {
		if in.AssessmentCode.Value == nil {
			return hrerrors.Invalid("assessment_code", "must not be null")
		}
		v.AssessmentCode = *in.AssessmentCode.Value
	}
	if in.AttemptNumber.Set {
		if in.AttemptNumber.Value == nil {
			return hrerrors.Invalid("attempt_number", "must not be null")
		}
		v.AttemptNumber = *in.AttemptNumber.Value
	}
	if in.RubricDocumentVersionReferenceID.Set {
		if in.RubricDocumentVersionReferenceID.Value == nil {
			return hrerrors.Invalid("rubric_document_version_reference_id", "must not be null")
		}
		v.RubricDocumentVersionReferenceID = *in.RubricDocumentVersionReferenceID.Value
	}
	if in.RubricDigest.Set {
		if in.RubricDigest.Value == nil {
			v.RubricDigest = nil
		} else {
			v.RubricDigest = *in.RubricDigest.Value
		}
	}
	if in.Status.Set {
		if in.Status.Value == nil {
			return hrerrors.Invalid("status", "must not be null")
		}
		v.Status = *in.Status.Value
	}
	if in.AssessorPrincipalID.Set {
		v.AssessorPrincipalID = in.AssessorPrincipalID.Value
	}
	if in.ScheduledAt.Set {
		v.ScheduledAt = in.ScheduledAt.Value
	}
	if in.CompletedAt.Set {
		v.CompletedAt = in.CompletedAt.Value
	}
	if in.NormalizedScore.Set {
		v.NormalizedScore = in.NormalizedScore.Value
	}
	if in.RecommendationCode.Set {
		v.RecommendationCode = in.RecommendationCode.Value
	}
	if in.EvidenceDocumentVersionReferenceID.Set {
		v.EvidenceDocumentVersionReferenceID = in.EvidenceDocumentVersionReferenceID.Value
	}
	if in.EvidenceDigest.Set {
		if in.EvidenceDigest.Value == nil {
			v.EvidenceDigest = nil
		} else {
			v.EvidenceDigest = *in.EvidenceDigest.Value
		}
	}
	return nil
}
func specApplicationAssessment() operation.Specification[entity.ApplicationAssessment, string] {
	return operation.Specification[entity.ApplicationAssessment, string]{
		Domain: "recruitment", Table: "hr_application_assessments",
		Key:         func(v *entity.ApplicationAssessment) string { return v.ID },
		KeyString:   func(key string) string { return fmt.Sprint(key) },
		ValidateKey: func(key string) error { return entity.ValidateUUID("id", key) },
		AggregateID: func(v *entity.ApplicationAssessment) string { return v.ID },
		Initialize: func(v *entity.ApplicationAssessment, actor operation.Actor, now time.Time, id string) {
			v.ID = id
		},
		Touch: func(v *entity.ApplicationAssessment, actor operation.Actor, now time.Time) {},
		Attributes: func(v *entity.ApplicationAssessment) map[string]string {
			a := map[string]string{}
			a["id"] = fmt.Sprint(v.ID)
			a["application_id"] = fmt.Sprint(v.ApplicationID)
			a["rubric_document_version_reference_id"] = fmt.Sprint(v.RubricDocumentVersionReferenceID)
			a["status"] = fmt.Sprint(v.Status)
			if v.AssessorPrincipalID != nil {
				a["assessor_principal_id"] = fmt.Sprint(*v.AssessorPrincipalID)
			}
			if v.EvidenceDocumentVersionReferenceID != nil {
				a["evidence_document_version_reference_id"] = fmt.Sprint(*v.EvidenceDocumentVersionReferenceID)
			}
			return a
		},
		Guard: func(action string, old, next *entity.ApplicationAssessment) error {
			if action == "update" && !reflect.DeepEqual(old.Status, next.Status) {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "create" && string(next.Status) != "planned" {
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
