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

func createPromotionRecommendation(in CreatePromotionRecommendationInput) (*entity.PromotionRecommendation, error) {
	v := &entity.PromotionRecommendation{}
	v.RecommendationNumber = in.RecommendationNumber
	v.WorkerID = in.WorkerID
	v.SourceAssignmentID = in.SourceAssignmentID
	v.TargetPositionID = in.TargetPositionID
	v.AssessmentID = in.AssessmentID
	v.SupersedesRecommendationID = in.SupersedesRecommendationID
	v.RecommendedEffectiveDate = in.RecommendedEffectiveDate
	v.RationaleDocumentVersionReferenceID = in.RationaleDocumentVersionReferenceID
	v.RationaleDigest = in.RationaleDigest
	if in.Status != nil {
		v.Status = *in.Status
	} else {
		v.Status = "draft"
	}
	v.ApprovalDecisionID = in.ApprovalDecisionID
	v.AppliedAssignmentID = in.AppliedAssignmentID
	v.AppliedLifecycleEventID = in.AppliedLifecycleEventID
	v.AppliedAt = in.AppliedAt
	v.IdempotencyKey = in.IdempotencyKey
	return v, nil
}
func patchPromotionRecommendation(v *entity.PromotionRecommendation, in UpdatePromotionRecommendationInput) error {
	if in.RecommendationNumber.Set {
		if in.RecommendationNumber.Value == nil {
			return hrerrors.Invalid("recommendation_number", "must not be null")
		}
		v.RecommendationNumber = *in.RecommendationNumber.Value
	}
	if in.WorkerID.Set {
		if in.WorkerID.Value == nil {
			return hrerrors.Invalid("worker_id", "must not be null")
		}
		v.WorkerID = *in.WorkerID.Value
	}
	if in.SourceAssignmentID.Set {
		if in.SourceAssignmentID.Value == nil {
			return hrerrors.Invalid("source_assignment_id", "must not be null")
		}
		v.SourceAssignmentID = *in.SourceAssignmentID.Value
	}
	if in.TargetPositionID.Set {
		if in.TargetPositionID.Value == nil {
			return hrerrors.Invalid("target_position_id", "must not be null")
		}
		v.TargetPositionID = *in.TargetPositionID.Value
	}
	if in.AssessmentID.Set {
		if in.AssessmentID.Value == nil {
			return hrerrors.Invalid("assessment_id", "must not be null")
		}
		v.AssessmentID = *in.AssessmentID.Value
	}
	if in.SupersedesRecommendationID.Set {
		v.SupersedesRecommendationID = in.SupersedesRecommendationID.Value
	}
	if in.RecommendedEffectiveDate.Set {
		if in.RecommendedEffectiveDate.Value == nil {
			return hrerrors.Invalid("recommended_effective_date", "must not be null")
		}
		v.RecommendedEffectiveDate = *in.RecommendedEffectiveDate.Value
	}
	if in.RationaleDocumentVersionReferenceID.Set {
		if in.RationaleDocumentVersionReferenceID.Value == nil {
			return hrerrors.Invalid("rationale_document_version_reference_id", "must not be null")
		}
		v.RationaleDocumentVersionReferenceID = *in.RationaleDocumentVersionReferenceID.Value
	}
	if in.RationaleDigest.Set {
		if in.RationaleDigest.Value == nil {
			v.RationaleDigest = nil
		} else {
			v.RationaleDigest = *in.RationaleDigest.Value
		}
	}
	if in.Status.Set {
		if in.Status.Value == nil {
			return hrerrors.Invalid("status", "must not be null")
		}
		v.Status = *in.Status.Value
	}
	if in.ApprovalDecisionID.Set {
		v.ApprovalDecisionID = in.ApprovalDecisionID.Value
	}
	if in.AppliedAssignmentID.Set {
		v.AppliedAssignmentID = in.AppliedAssignmentID.Value
	}
	if in.AppliedLifecycleEventID.Set {
		v.AppliedLifecycleEventID = in.AppliedLifecycleEventID.Value
	}
	if in.AppliedAt.Set {
		v.AppliedAt = in.AppliedAt.Value
	}
	if in.IdempotencyKey.Set {
		if in.IdempotencyKey.Value == nil {
			return hrerrors.Invalid("idempotency_key", "must not be null")
		}
		v.IdempotencyKey = *in.IdempotencyKey.Value
	}
	return nil
}
func specPromotionRecommendation() operation.Specification[entity.PromotionRecommendation, string] {
	return operation.Specification[entity.PromotionRecommendation, string]{
		Domain: "promotion", Table: "hr_promotion_recommendations",
		Key:         func(v *entity.PromotionRecommendation) string { return v.ID },
		KeyString:   func(key string) string { return fmt.Sprint(key) },
		ValidateKey: func(key string) error { return entity.ValidateUUID("id", key) },
		AggregateID: func(v *entity.PromotionRecommendation) string { return v.ID },
		Initialize: func(v *entity.PromotionRecommendation, actor operation.Actor, now time.Time, id string) {
			v.ID = id
			v.RecommendedByPrincipalID = actor.PrincipalID
			v.RecommendedAt = now
		},
		Touch: func(v *entity.PromotionRecommendation, actor operation.Actor, now time.Time) {},
		Attributes: func(v *entity.PromotionRecommendation) map[string]string {
			a := map[string]string{}
			a["id"] = fmt.Sprint(v.ID)
			a["worker_id"] = fmt.Sprint(v.WorkerID)
			a["source_assignment_id"] = fmt.Sprint(v.SourceAssignmentID)
			a["target_position_id"] = fmt.Sprint(v.TargetPositionID)
			a["assessment_id"] = fmt.Sprint(v.AssessmentID)
			if v.SupersedesRecommendationID != nil {
				a["supersedes_recommendation_id"] = fmt.Sprint(*v.SupersedesRecommendationID)
			}
			a["rationale_document_version_reference_id"] = fmt.Sprint(v.RationaleDocumentVersionReferenceID)
			a["recommended_by_principal_id"] = fmt.Sprint(v.RecommendedByPrincipalID)
			a["status"] = fmt.Sprint(v.Status)
			if v.ApprovalDecisionID != nil {
				a["approval_decision_id"] = fmt.Sprint(*v.ApprovalDecisionID)
			}
			if v.AppliedAssignmentID != nil {
				a["applied_assignment_id"] = fmt.Sprint(*v.AppliedAssignmentID)
			}
			return a
		},
		Guard: func(action string, old, next *entity.PromotionRecommendation) error {
			if action == "update" && !reflect.DeepEqual(old.Status, next.Status) {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "create" && string(next.Status) != "draft" {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "create" && next.ApprovalDecisionID != nil {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "update" && !reflect.DeepEqual(old.ApprovalDecisionID, next.ApprovalDecisionID) {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "create" && next.AppliedAssignmentID != nil {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "update" && !reflect.DeepEqual(old.AppliedAssignmentID, next.AppliedAssignmentID) {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "create" && next.AppliedLifecycleEventID != nil {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "update" && !reflect.DeepEqual(old.AppliedLifecycleEventID, next.AppliedLifecycleEventID) {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "create" && next.AppliedAt != nil {
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
func createPromotionDecision(in CreatePromotionDecisionInput) (*entity.PromotionDecision, error) {
	v := &entity.PromotionDecision{}
	v.RecommendationID = in.RecommendationID
	v.StepNumber = in.StepNumber
	v.Decision = in.Decision
	v.RationaleDocumentVersionReferenceID = in.RationaleDocumentVersionReferenceID
	v.RationaleDigest = in.RationaleDigest
	v.IdempotencyKey = in.IdempotencyKey
	return v, nil
}
func patchPromotionDecision(v *entity.PromotionDecision, in UpdatePromotionDecisionInput) error {
	if in.RecommendationID.Set {
		if in.RecommendationID.Value == nil {
			return hrerrors.Invalid("recommendation_id", "must not be null")
		}
		v.RecommendationID = *in.RecommendationID.Value
	}
	if in.StepNumber.Set {
		if in.StepNumber.Value == nil {
			return hrerrors.Invalid("step_number", "must not be null")
		}
		v.StepNumber = *in.StepNumber.Value
	}
	if in.Decision.Set {
		if in.Decision.Value == nil {
			return hrerrors.Invalid("decision", "must not be null")
		}
		v.Decision = *in.Decision.Value
	}
	if in.RationaleDocumentVersionReferenceID.Set {
		if in.RationaleDocumentVersionReferenceID.Value == nil {
			return hrerrors.Invalid("rationale_document_version_reference_id", "must not be null")
		}
		v.RationaleDocumentVersionReferenceID = *in.RationaleDocumentVersionReferenceID.Value
	}
	if in.RationaleDigest.Set {
		if in.RationaleDigest.Value == nil {
			v.RationaleDigest = nil
		} else {
			v.RationaleDigest = *in.RationaleDigest.Value
		}
	}
	if in.IdempotencyKey.Set {
		if in.IdempotencyKey.Value == nil {
			return hrerrors.Invalid("idempotency_key", "must not be null")
		}
		v.IdempotencyKey = *in.IdempotencyKey.Value
	}
	return nil
}
func specPromotionDecision() operation.Specification[entity.PromotionDecision, string] {
	return operation.Specification[entity.PromotionDecision, string]{
		Domain: "promotion", Table: "hr_promotion_decisions",
		Key:         func(v *entity.PromotionDecision) string { return v.ID },
		KeyString:   func(key string) string { return fmt.Sprint(key) },
		ValidateKey: func(key string) error { return entity.ValidateUUID("id", key) },
		AggregateID: func(v *entity.PromotionDecision) string { return v.ID },
		Initialize: func(v *entity.PromotionDecision, actor operation.Actor, now time.Time, id string) {
			v.ID = id
			v.DecidedByPrincipalID = actor.PrincipalID
			v.DecidedAt = now
		},
		Touch: func(v *entity.PromotionDecision, actor operation.Actor, now time.Time) {},
		Attributes: func(v *entity.PromotionDecision) map[string]string {
			a := map[string]string{}
			a["id"] = fmt.Sprint(v.ID)
			a["recommendation_id"] = fmt.Sprint(v.RecommendationID)
			a["decided_by_principal_id"] = fmt.Sprint(v.DecidedByPrincipalID)
			a["rationale_document_version_reference_id"] = fmt.Sprint(v.RationaleDocumentVersionReferenceID)
			return a
		},
		Guard: func(action string, old, next *entity.PromotionDecision) error {
			if action != "create" {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "create" {
				return hrerrors.ErrWorkflowRequired
			}
			return nil
		},
	}
}
