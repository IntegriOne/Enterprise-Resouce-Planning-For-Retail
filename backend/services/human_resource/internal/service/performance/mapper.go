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

func createPerformanceCycle(in CreatePerformanceCycleInput) (*entity.PerformanceCycle, error) {
	v := &entity.PerformanceCycle{}
	v.CycleCode = in.CycleCode
	v.CycleName = in.CycleName
	v.CycleType = in.CycleType
	if in.Status != nil {
		v.Status = *in.Status
	} else {
		v.Status = "draft"
	}
	v.PeriodStart = in.PeriodStart
	v.PeriodEnd = in.PeriodEnd
	v.SelfReviewDueAt = in.SelfReviewDueAt
	v.ManagerReviewDueAt = in.ManagerReviewDueAt
	v.CalibrationDueAt = in.CalibrationDueAt
	v.LaunchedAt = in.LaunchedAt
	v.LaunchedByPrincipalID = in.LaunchedByPrincipalID
	if in.Configuration != nil {
		v.Configuration = in.Configuration
	} else {
		v.Configuration = entity.JSON(`{}`)
	}
	return v, nil
}
func patchPerformanceCycle(v *entity.PerformanceCycle, in UpdatePerformanceCycleInput) error {
	if in.CycleCode.Set {
		if in.CycleCode.Value == nil {
			return hrerrors.Invalid("cycle_code", "must not be null")
		}
		v.CycleCode = *in.CycleCode.Value
	}
	if in.CycleName.Set {
		if in.CycleName.Value == nil {
			return hrerrors.Invalid("cycle_name", "must not be null")
		}
		v.CycleName = *in.CycleName.Value
	}
	if in.CycleType.Set {
		if in.CycleType.Value == nil {
			return hrerrors.Invalid("cycle_type", "must not be null")
		}
		v.CycleType = *in.CycleType.Value
	}
	if in.Status.Set {
		if in.Status.Value == nil {
			return hrerrors.Invalid("status", "must not be null")
		}
		v.Status = *in.Status.Value
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
	if in.SelfReviewDueAt.Set {
		v.SelfReviewDueAt = in.SelfReviewDueAt.Value
	}
	if in.ManagerReviewDueAt.Set {
		v.ManagerReviewDueAt = in.ManagerReviewDueAt.Value
	}
	if in.CalibrationDueAt.Set {
		v.CalibrationDueAt = in.CalibrationDueAt.Value
	}
	if in.LaunchedAt.Set {
		v.LaunchedAt = in.LaunchedAt.Value
	}
	if in.LaunchedByPrincipalID.Set {
		v.LaunchedByPrincipalID = in.LaunchedByPrincipalID.Value
	}
	if in.Configuration.Set {
		if in.Configuration.Value == nil {
			v.Configuration = nil
		} else {
			v.Configuration = *in.Configuration.Value
		}
	}
	return nil
}
func specPerformanceCycle() operation.Specification[entity.PerformanceCycle, string] {
	return operation.Specification[entity.PerformanceCycle, string]{
		Domain: "performance", Table: "hr_performance_cycles",
		Key:         func(v *entity.PerformanceCycle) string { return v.ID },
		KeyString:   func(key string) string { return fmt.Sprint(key) },
		ValidateKey: func(key string) error { return entity.ValidateUUID("id", key) },
		AggregateID: func(v *entity.PerformanceCycle) string { return v.ID },
		Initialize: func(v *entity.PerformanceCycle, actor operation.Actor, now time.Time, id string) {
			v.ID = id
		},
		Touch: func(v *entity.PerformanceCycle, actor operation.Actor, now time.Time) {},
		Attributes: func(v *entity.PerformanceCycle) map[string]string {
			a := map[string]string{}
			a["id"] = fmt.Sprint(v.ID)
			a["status"] = fmt.Sprint(v.Status)
			if v.LaunchedByPrincipalID != nil {
				a["launched_by_principal_id"] = fmt.Sprint(*v.LaunchedByPrincipalID)
			}
			return a
		},
		Guard: func(action string, old, next *entity.PerformanceCycle) error {
			if action == "update" && !reflect.DeepEqual(old.Status, next.Status) {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "create" && string(next.Status) != "draft" {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "create" && (next.LaunchedAt != nil || next.LaunchedByPrincipalID != nil) {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "update" && (!reflect.DeepEqual(old.LaunchedAt, next.LaunchedAt) ||
				!reflect.DeepEqual(old.LaunchedByPrincipalID, next.LaunchedByPrincipalID)) {
				return hrerrors.ErrWorkflowRequired
			}
			if action != "create" && old.Status != entity.PerformanceCycleStatusDraft {
				return hrerrors.ErrWorkflowRequired
			}
			return nil
		},
	}
}
func createPerformanceGoal(in CreatePerformanceGoalInput) (*entity.PerformanceGoal, error) {
	v := &entity.PerformanceGoal{}
	v.WorkerID = in.WorkerID
	v.PerformanceCycleID = in.PerformanceCycleID
	v.ParentGoalID = in.ParentGoalID
	v.GoalCode = in.GoalCode
	v.GoalTitle = in.GoalTitle
	v.GoalDescription = in.GoalDescription
	v.GoalType = in.GoalType
	v.MeasureDefinition = in.MeasureDefinition
	v.TargetValue = in.TargetValue
	v.ActualValue = in.ActualValue
	v.UnitCode = in.UnitCode
	v.WeightPercent = in.WeightPercent
	v.DueDate = in.DueDate
	if in.GoalStatus != nil {
		v.GoalStatus = *in.GoalStatus
	} else {
		v.GoalStatus = "draft"
	}
	v.ApprovedByPrincipalID = in.ApprovedByPrincipalID
	v.ApprovedAt = in.ApprovedAt
	return v, nil
}
func patchPerformanceGoal(v *entity.PerformanceGoal, in UpdatePerformanceGoalInput) error {
	if in.WorkerID.Set {
		if in.WorkerID.Value == nil {
			return hrerrors.Invalid("worker_id", "must not be null")
		}
		v.WorkerID = *in.WorkerID.Value
	}
	if in.PerformanceCycleID.Set {
		if in.PerformanceCycleID.Value == nil {
			return hrerrors.Invalid("performance_cycle_id", "must not be null")
		}
		v.PerformanceCycleID = *in.PerformanceCycleID.Value
	}
	if in.ParentGoalID.Set {
		v.ParentGoalID = in.ParentGoalID.Value
	}
	if in.GoalCode.Set {
		v.GoalCode = in.GoalCode.Value
	}
	if in.GoalTitle.Set {
		if in.GoalTitle.Value == nil {
			return hrerrors.Invalid("goal_title", "must not be null")
		}
		v.GoalTitle = *in.GoalTitle.Value
	}
	if in.GoalDescription.Set {
		v.GoalDescription = in.GoalDescription.Value
	}
	if in.GoalType.Set {
		if in.GoalType.Value == nil {
			return hrerrors.Invalid("goal_type", "must not be null")
		}
		v.GoalType = *in.GoalType.Value
	}
	if in.MeasureDefinition.Set {
		if in.MeasureDefinition.Value == nil {
			v.MeasureDefinition = nil
		} else {
			v.MeasureDefinition = *in.MeasureDefinition.Value
		}
	}
	if in.TargetValue.Set {
		v.TargetValue = in.TargetValue.Value
	}
	if in.ActualValue.Set {
		v.ActualValue = in.ActualValue.Value
	}
	if in.UnitCode.Set {
		v.UnitCode = in.UnitCode.Value
	}
	if in.WeightPercent.Set {
		v.WeightPercent = in.WeightPercent.Value
	}
	if in.DueDate.Set {
		v.DueDate = in.DueDate.Value
	}
	if in.GoalStatus.Set {
		if in.GoalStatus.Value == nil {
			return hrerrors.Invalid("goal_status", "must not be null")
		}
		v.GoalStatus = *in.GoalStatus.Value
	}
	if in.ApprovedByPrincipalID.Set {
		v.ApprovedByPrincipalID = in.ApprovedByPrincipalID.Value
	}
	if in.ApprovedAt.Set {
		v.ApprovedAt = in.ApprovedAt.Value
	}
	return nil
}
func specPerformanceGoal() operation.Specification[entity.PerformanceGoal, string] {
	return operation.Specification[entity.PerformanceGoal, string]{
		Domain: "performance", Table: "hr_performance_goals",
		Key:         func(v *entity.PerformanceGoal) string { return v.ID },
		KeyString:   func(key string) string { return fmt.Sprint(key) },
		ValidateKey: func(key string) error { return entity.ValidateUUID("id", key) },
		AggregateID: func(v *entity.PerformanceGoal) string { return v.ID },
		Initialize: func(v *entity.PerformanceGoal, actor operation.Actor, now time.Time, id string) {
			v.ID = id
		},
		Touch: func(v *entity.PerformanceGoal, actor operation.Actor, now time.Time) {},
		Attributes: func(v *entity.PerformanceGoal) map[string]string {
			a := map[string]string{}
			a["id"] = fmt.Sprint(v.ID)
			a["worker_id"] = fmt.Sprint(v.WorkerID)
			a["performance_cycle_id"] = fmt.Sprint(v.PerformanceCycleID)
			if v.ParentGoalID != nil {
				a["parent_goal_id"] = fmt.Sprint(*v.ParentGoalID)
			}
			a["goal_status"] = fmt.Sprint(v.GoalStatus)
			if v.ApprovedByPrincipalID != nil {
				a["approved_by_principal_id"] = fmt.Sprint(*v.ApprovedByPrincipalID)
			}
			return a
		},
		Guard: func(action string, old, next *entity.PerformanceGoal) error {
			if action == "update" && !reflect.DeepEqual(old.GoalStatus, next.GoalStatus) {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "create" && string(next.GoalStatus) != "draft" {
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
				switch string(old.GoalStatus) {
				case "approved", "posted", "paid", "finalized", "applied", "closed", "archived", "terminated", "deceased", "completed", "accepted", "issued", "signed", "hired", "processed", "rejected", "cancelled", "withdrawn", "superseded":
					return hrerrors.ErrWorkflowRequired
				}
			}
			return nil
		},
	}
}
func createPerformanceReview(in CreatePerformanceReviewInput) (*entity.PerformanceReview, error) {
	v := &entity.PerformanceReview{}
	v.WorkerID = in.WorkerID
	v.PerformanceCycleID = in.PerformanceCycleID
	v.ReviewerPrincipalID = in.ReviewerPrincipalID
	v.ReviewerRole = in.ReviewerRole
	if in.ReviewVersion != nil {
		v.ReviewVersion = *in.ReviewVersion
	} else {
		v.ReviewVersion = 1
	}
	v.OverallRatingCode = in.OverallRatingCode
	if in.CompetencyAssessment != nil {
		v.CompetencyAssessment = in.CompetencyAssessment
	} else {
		v.CompetencyAssessment = entity.JSON(`{}`)
	}
	v.Narrative = in.Narrative
	v.DevelopmentPlanID = in.DevelopmentPlanID
	v.SubmittedAt = in.SubmittedAt
	v.FinalizedAt = in.FinalizedAt
	if in.VisibilityClass != nil {
		v.VisibilityClass = *in.VisibilityClass
	} else {
		v.VisibilityClass = "restricted"
	}
	v.ContentDigest = in.ContentDigest
	return v, nil
}
func patchPerformanceReview(v *entity.PerformanceReview, in UpdatePerformanceReviewInput) error {
	if in.WorkerID.Set {
		if in.WorkerID.Value == nil {
			return hrerrors.Invalid("worker_id", "must not be null")
		}
		v.WorkerID = *in.WorkerID.Value
	}
	if in.PerformanceCycleID.Set {
		if in.PerformanceCycleID.Value == nil {
			return hrerrors.Invalid("performance_cycle_id", "must not be null")
		}
		v.PerformanceCycleID = *in.PerformanceCycleID.Value
	}
	if in.ReviewerPrincipalID.Set {
		if in.ReviewerPrincipalID.Value == nil {
			return hrerrors.Invalid("reviewer_principal_id", "must not be null")
		}
		v.ReviewerPrincipalID = *in.ReviewerPrincipalID.Value
	}
	if in.ReviewerRole.Set {
		if in.ReviewerRole.Value == nil {
			return hrerrors.Invalid("reviewer_role", "must not be null")
		}
		v.ReviewerRole = *in.ReviewerRole.Value
	}
	if in.ReviewVersion.Set {
		if in.ReviewVersion.Value == nil {
			return hrerrors.Invalid("review_version", "must not be null")
		}
		v.ReviewVersion = *in.ReviewVersion.Value
	}
	if in.OverallRatingCode.Set {
		v.OverallRatingCode = in.OverallRatingCode.Value
	}
	if in.CompetencyAssessment.Set {
		if in.CompetencyAssessment.Value == nil {
			v.CompetencyAssessment = nil
		} else {
			v.CompetencyAssessment = *in.CompetencyAssessment.Value
		}
	}
	if in.Narrative.Set {
		v.Narrative = in.Narrative.Value
	}
	if in.DevelopmentPlanID.Set {
		v.DevelopmentPlanID = in.DevelopmentPlanID.Value
	}
	if in.SubmittedAt.Set {
		v.SubmittedAt = in.SubmittedAt.Value
	}
	if in.FinalizedAt.Set {
		v.FinalizedAt = in.FinalizedAt.Value
	}
	if in.VisibilityClass.Set {
		if in.VisibilityClass.Value == nil {
			return hrerrors.Invalid("visibility_class", "must not be null")
		}
		v.VisibilityClass = *in.VisibilityClass.Value
	}
	if in.ContentDigest.Set {
		if in.ContentDigest.Value == nil {
			v.ContentDigest = nil
		} else {
			v.ContentDigest = *in.ContentDigest.Value
		}
	}
	return nil
}
func specPerformanceReview() operation.Specification[entity.PerformanceReview, string] {
	return operation.Specification[entity.PerformanceReview, string]{
		Domain: "performance", Table: "hr_performance_reviews",
		Key:         func(v *entity.PerformanceReview) string { return v.ID },
		KeyString:   func(key string) string { return fmt.Sprint(key) },
		ValidateKey: func(key string) error { return entity.ValidateUUID("id", key) },
		AggregateID: func(v *entity.PerformanceReview) string { return v.ID },
		Initialize: func(v *entity.PerformanceReview, actor operation.Actor, now time.Time, id string) {
			v.ID = id
		},
		Touch: func(v *entity.PerformanceReview, actor operation.Actor, now time.Time) {},
		Attributes: func(v *entity.PerformanceReview) map[string]string {
			a := map[string]string{}
			a["id"] = fmt.Sprint(v.ID)
			a["worker_id"] = fmt.Sprint(v.WorkerID)
			a["performance_cycle_id"] = fmt.Sprint(v.PerformanceCycleID)
			a["reviewer_principal_id"] = fmt.Sprint(v.ReviewerPrincipalID)
			if v.DevelopmentPlanID != nil {
				a["development_plan_id"] = fmt.Sprint(*v.DevelopmentPlanID)
			}
			return a
		},
		Guard: func(action string, old, next *entity.PerformanceReview) error {
			if action == "create" && next.FinalizedAt != nil {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "update" && !reflect.DeepEqual(old.FinalizedAt, next.FinalizedAt) {
				return hrerrors.ErrWorkflowRequired
			}
			if action != "create" && old.FinalizedAt != nil {
				return hrerrors.ErrWorkflowRequired
			}
			return nil
		},
	}
}
func createPerformanceGoalUpdate(in CreatePerformanceGoalUpdateInput) (*entity.PerformanceGoalUpdate, error) {
	v := &entity.PerformanceGoalUpdate{}
	v.GoalID = in.GoalID
	v.ObservedAt = in.ObservedAt
	v.ActualValue = in.ActualValue
	v.EvidenceDocumentVersionReferenceID = in.EvidenceDocumentVersionReferenceID
	v.EvidenceDigest = in.EvidenceDigest
	v.SupersedesUpdateID = in.SupersedesUpdateID
	v.IdempotencyKey = in.IdempotencyKey
	return v, nil
}
func patchPerformanceGoalUpdate(v *entity.PerformanceGoalUpdate, in UpdatePerformanceGoalUpdateInput) error {
	if in.GoalID.Set {
		if in.GoalID.Value == nil {
			return hrerrors.Invalid("goal_id", "must not be null")
		}
		v.GoalID = *in.GoalID.Value
	}
	if in.ObservedAt.Set {
		if in.ObservedAt.Value == nil {
			return hrerrors.Invalid("observed_at", "must not be null")
		}
		v.ObservedAt = *in.ObservedAt.Value
	}
	if in.ActualValue.Set {
		if in.ActualValue.Value == nil {
			return hrerrors.Invalid("actual_value", "must not be null")
		}
		v.ActualValue = *in.ActualValue.Value
	}
	if in.EvidenceDocumentVersionReferenceID.Set {
		if in.EvidenceDocumentVersionReferenceID.Value == nil {
			return hrerrors.Invalid("evidence_document_version_reference_id", "must not be null")
		}
		v.EvidenceDocumentVersionReferenceID = *in.EvidenceDocumentVersionReferenceID.Value
	}
	if in.EvidenceDigest.Set {
		if in.EvidenceDigest.Value == nil {
			v.EvidenceDigest = nil
		} else {
			v.EvidenceDigest = *in.EvidenceDigest.Value
		}
	}
	if in.SupersedesUpdateID.Set {
		v.SupersedesUpdateID = in.SupersedesUpdateID.Value
	}
	if in.IdempotencyKey.Set {
		if in.IdempotencyKey.Value == nil {
			return hrerrors.Invalid("idempotency_key", "must not be null")
		}
		v.IdempotencyKey = *in.IdempotencyKey.Value
	}
	return nil
}
func specPerformanceGoalUpdate() operation.Specification[entity.PerformanceGoalUpdate, string] {
	return operation.Specification[entity.PerformanceGoalUpdate, string]{
		Domain: "performance", Table: "hr_performance_goal_updates",
		Key:         func(v *entity.PerformanceGoalUpdate) string { return v.ID },
		KeyString:   func(key string) string { return fmt.Sprint(key) },
		ValidateKey: func(key string) error { return entity.ValidateUUID("id", key) },
		AggregateID: func(v *entity.PerformanceGoalUpdate) string { return v.ID },
		Initialize: func(v *entity.PerformanceGoalUpdate, actor operation.Actor, now time.Time, id string) {
			v.ID = id
			v.RecordedByPrincipalID = actor.PrincipalID
			v.RecordedAt = now
		},
		Touch: func(v *entity.PerformanceGoalUpdate, actor operation.Actor, now time.Time) {},
		Attributes: func(v *entity.PerformanceGoalUpdate) map[string]string {
			a := map[string]string{}
			a["id"] = fmt.Sprint(v.ID)
			a["goal_id"] = fmt.Sprint(v.GoalID)
			a["evidence_document_version_reference_id"] = fmt.Sprint(v.EvidenceDocumentVersionReferenceID)
			a["recorded_by_principal_id"] = fmt.Sprint(v.RecordedByPrincipalID)
			if v.SupersedesUpdateID != nil {
				a["supersedes_update_id"] = fmt.Sprint(*v.SupersedesUpdateID)
			}
			return a
		},
		Guard: func(action string, old, next *entity.PerformanceGoalUpdate) error {
			if action != "create" {
				return hrerrors.ErrWorkflowRequired
			}
			return nil
		},
	}
}
func createPerformanceReviewGoalScore(in CreatePerformanceReviewGoalScoreInput) (*entity.PerformanceReviewGoalScore, error) {
	v := &entity.PerformanceReviewGoalScore{}
	v.PerformanceReviewID = in.PerformanceReviewID
	v.GoalID = in.GoalID
	v.WorkerID = in.WorkerID
	v.PerformanceCycleID = in.PerformanceCycleID
	v.Rating = in.Rating
	v.WeightPercentSnapshot = in.WeightPercentSnapshot
	v.RubricDocumentVersionReferenceID = in.RubricDocumentVersionReferenceID
	v.EvidenceDocumentVersionReferenceID = in.EvidenceDocumentVersionReferenceID
	return v, nil
}
func patchPerformanceReviewGoalScore(v *entity.PerformanceReviewGoalScore, in UpdatePerformanceReviewGoalScoreInput) error {
	if in.PerformanceReviewID.Set {
		if in.PerformanceReviewID.Value == nil {
			return hrerrors.Invalid("performance_review_id", "must not be null")
		}
		v.PerformanceReviewID = *in.PerformanceReviewID.Value
	}
	if in.GoalID.Set {
		if in.GoalID.Value == nil {
			return hrerrors.Invalid("goal_id", "must not be null")
		}
		v.GoalID = *in.GoalID.Value
	}
	if in.WorkerID.Set {
		if in.WorkerID.Value == nil {
			return hrerrors.Invalid("worker_id", "must not be null")
		}
		v.WorkerID = *in.WorkerID.Value
	}
	if in.PerformanceCycleID.Set {
		if in.PerformanceCycleID.Value == nil {
			return hrerrors.Invalid("performance_cycle_id", "must not be null")
		}
		v.PerformanceCycleID = *in.PerformanceCycleID.Value
	}
	if in.Rating.Set {
		if in.Rating.Value == nil {
			return hrerrors.Invalid("rating", "must not be null")
		}
		v.Rating = *in.Rating.Value
	}
	if in.WeightPercentSnapshot.Set {
		if in.WeightPercentSnapshot.Value == nil {
			return hrerrors.Invalid("weight_percent_snapshot", "must not be null")
		}
		v.WeightPercentSnapshot = *in.WeightPercentSnapshot.Value
	}
	if in.RubricDocumentVersionReferenceID.Set {
		if in.RubricDocumentVersionReferenceID.Value == nil {
			return hrerrors.Invalid("rubric_document_version_reference_id", "must not be null")
		}
		v.RubricDocumentVersionReferenceID = *in.RubricDocumentVersionReferenceID.Value
	}
	if in.EvidenceDocumentVersionReferenceID.Set {
		if in.EvidenceDocumentVersionReferenceID.Value == nil {
			return hrerrors.Invalid("evidence_document_version_reference_id", "must not be null")
		}
		v.EvidenceDocumentVersionReferenceID = *in.EvidenceDocumentVersionReferenceID.Value
	}
	return nil
}
func specPerformanceReviewGoalScore() operation.Specification[entity.PerformanceReviewGoalScore, string] {
	return operation.Specification[entity.PerformanceReviewGoalScore, string]{
		Domain: "performance", Table: "hr_performance_review_goal_scores",
		Key:         func(v *entity.PerformanceReviewGoalScore) string { return v.ID },
		KeyString:   func(key string) string { return fmt.Sprint(key) },
		ValidateKey: func(key string) error { return entity.ValidateUUID("id", key) },
		AggregateID: func(v *entity.PerformanceReviewGoalScore) string { return v.ID },
		Initialize: func(v *entity.PerformanceReviewGoalScore, actor operation.Actor, now time.Time, id string) {
			v.ID = id
		},
		Touch: func(v *entity.PerformanceReviewGoalScore, actor operation.Actor, now time.Time) {},
		Attributes: func(v *entity.PerformanceReviewGoalScore) map[string]string {
			a := map[string]string{}
			a["id"] = fmt.Sprint(v.ID)
			a["performance_review_id"] = fmt.Sprint(v.PerformanceReviewID)
			a["goal_id"] = fmt.Sprint(v.GoalID)
			a["worker_id"] = fmt.Sprint(v.WorkerID)
			a["performance_cycle_id"] = fmt.Sprint(v.PerformanceCycleID)
			a["rubric_document_version_reference_id"] = fmt.Sprint(v.RubricDocumentVersionReferenceID)
			a["evidence_document_version_reference_id"] = fmt.Sprint(v.EvidenceDocumentVersionReferenceID)
			return a
		},
		Guard: func(action string, old, next *entity.PerformanceReviewGoalScore) error {
			return nil
		},
	}
}
