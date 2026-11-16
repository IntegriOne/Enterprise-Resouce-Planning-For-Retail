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

func createTalentFramework(in CreateTalentFrameworkInput) (*entity.TalentFramework, error) {
	v := &entity.TalentFramework{}
	v.FrameworkCode = in.FrameworkCode
	v.FrameworkVersion = in.FrameworkVersion
	v.FrameworkName = in.FrameworkName
	v.RubricDocumentVersionReferenceID = in.RubricDocumentVersionReferenceID
	v.RubricDigest = in.RubricDigest
	v.EffectiveFrom = in.EffectiveFrom
	v.EffectiveTo = in.EffectiveTo
	v.ApprovedByPrincipalID = in.ApprovedByPrincipalID
	v.ApprovedAt = in.ApprovedAt
	return v, nil
}
func patchTalentFramework(v *entity.TalentFramework, in UpdateTalentFrameworkInput) error {
	if in.FrameworkCode.Set {
		if in.FrameworkCode.Value == nil {
			return hrerrors.Invalid("framework_code", "must not be null")
		}
		v.FrameworkCode = *in.FrameworkCode.Value
	}
	if in.FrameworkVersion.Set {
		if in.FrameworkVersion.Value == nil {
			return hrerrors.Invalid("framework_version", "must not be null")
		}
		v.FrameworkVersion = *in.FrameworkVersion.Value
	}
	if in.FrameworkName.Set {
		if in.FrameworkName.Value == nil {
			return hrerrors.Invalid("framework_name", "must not be null")
		}
		v.FrameworkName = *in.FrameworkName.Value
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
	if in.EffectiveFrom.Set {
		if in.EffectiveFrom.Value == nil {
			return hrerrors.Invalid("effective_from", "must not be null")
		}
		v.EffectiveFrom = *in.EffectiveFrom.Value
	}
	if in.EffectiveTo.Set {
		v.EffectiveTo = in.EffectiveTo.Value
	}
	if in.ApprovedByPrincipalID.Set {
		if in.ApprovedByPrincipalID.Value == nil {
			return hrerrors.Invalid("approved_by_principal_id", "must not be null")
		}
		v.ApprovedByPrincipalID = *in.ApprovedByPrincipalID.Value
	}
	if in.ApprovedAt.Set {
		if in.ApprovedAt.Value == nil {
			return hrerrors.Invalid("approved_at", "must not be null")
		}
		v.ApprovedAt = *in.ApprovedAt.Value
	}
	return nil
}
func specTalentFramework() operation.Specification[entity.TalentFramework, string] {
	return operation.Specification[entity.TalentFramework, string]{
		Domain: "talent", Table: "hr_talent_frameworks",
		Key:         func(v *entity.TalentFramework) string { return v.ID },
		KeyString:   func(key string) string { return fmt.Sprint(key) },
		ValidateKey: func(key string) error { return entity.ValidateUUID("id", key) },
		AggregateID: func(v *entity.TalentFramework) string { return v.ID },
		Initialize: func(v *entity.TalentFramework, actor operation.Actor, now time.Time, id string) {
			v.ID = id
		},
		Touch: func(v *entity.TalentFramework, actor operation.Actor, now time.Time) {},
		Attributes: func(v *entity.TalentFramework) map[string]string {
			a := map[string]string{}
			a["id"] = fmt.Sprint(v.ID)
			a["rubric_document_version_reference_id"] = fmt.Sprint(v.RubricDocumentVersionReferenceID)
			a["approved_by_principal_id"] = fmt.Sprint(v.ApprovedByPrincipalID)
			return a
		},
		Guard: func(action string, old, next *entity.TalentFramework) error {
			if action != "create" {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "create" {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "create" && !reflect.ValueOf(next.ApprovedByPrincipalID).IsZero() {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "update" && !reflect.DeepEqual(old.ApprovedByPrincipalID, next.ApprovedByPrincipalID) {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "create" && !reflect.ValueOf(next.ApprovedAt).IsZero() {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "update" && !reflect.DeepEqual(old.ApprovedAt, next.ApprovedAt) {
				return hrerrors.ErrWorkflowRequired
			}
			if action != "create" {
				return hrerrors.ErrWorkflowRequired
			}
			return nil
		},
	}
}
func createTalentDimension(in CreateTalentDimensionInput) (*entity.TalentDimension, error) {
	v := &entity.TalentDimension{}
	v.FrameworkID = in.FrameworkID
	v.DimensionCode = in.DimensionCode
	v.DimensionName = in.DimensionName
	v.Definition = in.Definition
	v.RatingAnchors = in.RatingAnchors
	v.WeightPercent = in.WeightPercent
	if in.IsRequired != nil {
		v.IsRequired = *in.IsRequired
	} else {
		v.IsRequired = true
	}
	return v, nil
}
func patchTalentDimension(v *entity.TalentDimension, in UpdateTalentDimensionInput) error {
	if in.FrameworkID.Set {
		if in.FrameworkID.Value == nil {
			return hrerrors.Invalid("framework_id", "must not be null")
		}
		v.FrameworkID = *in.FrameworkID.Value
	}
	if in.DimensionCode.Set {
		if in.DimensionCode.Value == nil {
			return hrerrors.Invalid("dimension_code", "must not be null")
		}
		v.DimensionCode = *in.DimensionCode.Value
	}
	if in.DimensionName.Set {
		if in.DimensionName.Value == nil {
			return hrerrors.Invalid("dimension_name", "must not be null")
		}
		v.DimensionName = *in.DimensionName.Value
	}
	if in.Definition.Set {
		if in.Definition.Value == nil {
			return hrerrors.Invalid("definition", "must not be null")
		}
		v.Definition = *in.Definition.Value
	}
	if in.RatingAnchors.Set {
		if in.RatingAnchors.Value == nil {
			v.RatingAnchors = nil
		} else {
			v.RatingAnchors = *in.RatingAnchors.Value
		}
	}
	if in.WeightPercent.Set {
		if in.WeightPercent.Value == nil {
			return hrerrors.Invalid("weight_percent", "must not be null")
		}
		v.WeightPercent = *in.WeightPercent.Value
	}
	if in.IsRequired.Set {
		if in.IsRequired.Value == nil {
			return hrerrors.Invalid("is_required", "must not be null")
		}
		v.IsRequired = *in.IsRequired.Value
	}
	return nil
}
func specTalentDimension() operation.Specification[entity.TalentDimension, string] {
	return operation.Specification[entity.TalentDimension, string]{
		Domain: "talent", Table: "hr_talent_dimensions",
		Key:         func(v *entity.TalentDimension) string { return v.ID },
		KeyString:   func(key string) string { return fmt.Sprint(key) },
		ValidateKey: func(key string) error { return entity.ValidateUUID("id", key) },
		AggregateID: func(v *entity.TalentDimension) string { return v.ID },
		Initialize: func(v *entity.TalentDimension, actor operation.Actor, now time.Time, id string) {
			v.ID = id
		},
		Touch: func(v *entity.TalentDimension, actor operation.Actor, now time.Time) {},
		Attributes: func(v *entity.TalentDimension) map[string]string {
			a := map[string]string{}
			a["id"] = fmt.Sprint(v.ID)
			a["framework_id"] = fmt.Sprint(v.FrameworkID)
			return a
		},
		Guard: func(action string, old, next *entity.TalentDimension) error {
			return nil
		},
	}
}
func createTalentReviewCycle(in CreateTalentReviewCycleInput) (*entity.TalentReviewCycle, error) {
	v := &entity.TalentReviewCycle{}
	v.CycleCode = in.CycleCode
	v.FrameworkID = in.FrameworkID
	v.LegalEntityReferenceID = in.LegalEntityReferenceID
	v.OrganizationUnitReferenceID = in.OrganizationUnitReferenceID
	v.PeriodStart = in.PeriodStart
	v.PeriodEnd = in.PeriodEnd
	if in.Status != nil {
		v.Status = *in.Status
	} else {
		v.Status = "draft"
	}
	v.OwnerPrincipalID = in.OwnerPrincipalID
	v.ClosedAt = in.ClosedAt
	return v, nil
}
func patchTalentReviewCycle(v *entity.TalentReviewCycle, in UpdateTalentReviewCycleInput) error {
	if in.CycleCode.Set {
		if in.CycleCode.Value == nil {
			return hrerrors.Invalid("cycle_code", "must not be null")
		}
		v.CycleCode = *in.CycleCode.Value
	}
	if in.FrameworkID.Set {
		if in.FrameworkID.Value == nil {
			return hrerrors.Invalid("framework_id", "must not be null")
		}
		v.FrameworkID = *in.FrameworkID.Value
	}
	if in.LegalEntityReferenceID.Set {
		if in.LegalEntityReferenceID.Value == nil {
			return hrerrors.Invalid("legal_entity_reference_id", "must not be null")
		}
		v.LegalEntityReferenceID = *in.LegalEntityReferenceID.Value
	}
	if in.OrganizationUnitReferenceID.Set {
		v.OrganizationUnitReferenceID = in.OrganizationUnitReferenceID.Value
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
	if in.Status.Set {
		if in.Status.Value == nil {
			return hrerrors.Invalid("status", "must not be null")
		}
		v.Status = *in.Status.Value
	}
	if in.OwnerPrincipalID.Set {
		if in.OwnerPrincipalID.Value == nil {
			return hrerrors.Invalid("owner_principal_id", "must not be null")
		}
		v.OwnerPrincipalID = *in.OwnerPrincipalID.Value
	}
	if in.ClosedAt.Set {
		v.ClosedAt = in.ClosedAt.Value
	}
	return nil
}
func specTalentReviewCycle() operation.Specification[entity.TalentReviewCycle, string] {
	return operation.Specification[entity.TalentReviewCycle, string]{
		Domain: "talent", Table: "hr_talent_review_cycles",
		Key:         func(v *entity.TalentReviewCycle) string { return v.ID },
		KeyString:   func(key string) string { return fmt.Sprint(key) },
		ValidateKey: func(key string) error { return entity.ValidateUUID("id", key) },
		AggregateID: func(v *entity.TalentReviewCycle) string { return v.ID },
		Initialize: func(v *entity.TalentReviewCycle, actor operation.Actor, now time.Time, id string) {
			v.ID = id
		},
		Touch: func(v *entity.TalentReviewCycle, actor operation.Actor, now time.Time) {},
		Attributes: func(v *entity.TalentReviewCycle) map[string]string {
			a := map[string]string{}
			a["id"] = fmt.Sprint(v.ID)
			a["framework_id"] = fmt.Sprint(v.FrameworkID)
			a["legal_entity_reference_id"] = fmt.Sprint(v.LegalEntityReferenceID)
			if v.OrganizationUnitReferenceID != nil {
				a["organization_unit_reference_id"] = fmt.Sprint(*v.OrganizationUnitReferenceID)
			}
			a["status"] = fmt.Sprint(v.Status)
			a["owner_principal_id"] = fmt.Sprint(v.OwnerPrincipalID)
			return a
		},
		Guard: func(action string, old, next *entity.TalentReviewCycle) error {
			if action == "update" && !reflect.DeepEqual(old.Status, next.Status) {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "create" && string(next.Status) != "draft" {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "create" && next.ClosedAt != nil {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "update" && !reflect.DeepEqual(old.ClosedAt, next.ClosedAt) {
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
func createTalentAssessment(in CreateTalentAssessmentInput) (*entity.TalentAssessment, error) {
	v := &entity.TalentAssessment{}
	v.WorkerID = in.WorkerID
	v.TalentReviewCycleID = in.TalentReviewCycleID
	v.FrameworkID = in.FrameworkID
	v.AssessmentVersion = in.AssessmentVersion
	v.PerformanceReviewID = in.PerformanceReviewID
	if in.Status != nil {
		v.Status = *in.Status
	} else {
		v.Status = "draft"
	}
	v.EvidenceDocumentVersionReferenceID = in.EvidenceDocumentVersionReferenceID
	v.EvidenceDigest = in.EvidenceDigest
	v.SegmentCode = in.SegmentCode
	v.FinalizedByPrincipalID = in.FinalizedByPrincipalID
	v.FinalizedAt = in.FinalizedAt
	return v, nil
}
func patchTalentAssessment(v *entity.TalentAssessment, in UpdateTalentAssessmentInput) error {
	if in.WorkerID.Set {
		if in.WorkerID.Value == nil {
			return hrerrors.Invalid("worker_id", "must not be null")
		}
		v.WorkerID = *in.WorkerID.Value
	}
	if in.TalentReviewCycleID.Set {
		if in.TalentReviewCycleID.Value == nil {
			return hrerrors.Invalid("talent_review_cycle_id", "must not be null")
		}
		v.TalentReviewCycleID = *in.TalentReviewCycleID.Value
	}
	if in.FrameworkID.Set {
		if in.FrameworkID.Value == nil {
			return hrerrors.Invalid("framework_id", "must not be null")
		}
		v.FrameworkID = *in.FrameworkID.Value
	}
	if in.AssessmentVersion.Set {
		if in.AssessmentVersion.Value == nil {
			return hrerrors.Invalid("assessment_version", "must not be null")
		}
		v.AssessmentVersion = *in.AssessmentVersion.Value
	}
	if in.PerformanceReviewID.Set {
		v.PerformanceReviewID = in.PerformanceReviewID.Value
	}
	if in.Status.Set {
		if in.Status.Value == nil {
			return hrerrors.Invalid("status", "must not be null")
		}
		v.Status = *in.Status.Value
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
	if in.SegmentCode.Set {
		v.SegmentCode = in.SegmentCode.Value
	}
	if in.FinalizedByPrincipalID.Set {
		v.FinalizedByPrincipalID = in.FinalizedByPrincipalID.Value
	}
	if in.FinalizedAt.Set {
		v.FinalizedAt = in.FinalizedAt.Value
	}
	return nil
}
func specTalentAssessment() operation.Specification[entity.TalentAssessment, string] {
	return operation.Specification[entity.TalentAssessment, string]{
		Domain: "talent", Table: "hr_talent_assessments",
		Key:         func(v *entity.TalentAssessment) string { return v.ID },
		KeyString:   func(key string) string { return fmt.Sprint(key) },
		ValidateKey: func(key string) error { return entity.ValidateUUID("id", key) },
		AggregateID: func(v *entity.TalentAssessment) string { return v.ID },
		Initialize: func(v *entity.TalentAssessment, actor operation.Actor, now time.Time, id string) {
			v.ID = id
			v.AssessedByPrincipalID = actor.PrincipalID
			v.AssessedAt = now
		},
		Touch: func(v *entity.TalentAssessment, actor operation.Actor, now time.Time) {},
		Attributes: func(v *entity.TalentAssessment) map[string]string {
			a := map[string]string{}
			a["id"] = fmt.Sprint(v.ID)
			a["worker_id"] = fmt.Sprint(v.WorkerID)
			a["talent_review_cycle_id"] = fmt.Sprint(v.TalentReviewCycleID)
			a["framework_id"] = fmt.Sprint(v.FrameworkID)
			if v.PerformanceReviewID != nil {
				a["performance_review_id"] = fmt.Sprint(*v.PerformanceReviewID)
			}
			a["status"] = fmt.Sprint(v.Status)
			a["assessed_by_principal_id"] = fmt.Sprint(v.AssessedByPrincipalID)
			a["evidence_document_version_reference_id"] = fmt.Sprint(v.EvidenceDocumentVersionReferenceID)
			if v.FinalizedByPrincipalID != nil {
				a["finalized_by_principal_id"] = fmt.Sprint(*v.FinalizedByPrincipalID)
			}
			return a
		},
		Guard: func(action string, old, next *entity.TalentAssessment) error {
			if action == "update" && !reflect.DeepEqual(old.Status, next.Status) {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "create" && string(next.Status) != "draft" {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "create" && next.FinalizedAt != nil {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "update" && !reflect.DeepEqual(old.FinalizedAt, next.FinalizedAt) {
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
func createTalentAssessmentScore(in CreateTalentAssessmentScoreInput) (*entity.TalentAssessmentScore, error) {
	v := &entity.TalentAssessmentScore{}
	v.AssessmentID = in.AssessmentID
	v.FrameworkID = in.FrameworkID
	v.DimensionID = in.DimensionID
	v.Score = in.Score
	v.EvidenceDocumentVersionReferenceID = in.EvidenceDocumentVersionReferenceID
	v.EvidenceDigest = in.EvidenceDigest
	return v, nil
}
func patchTalentAssessmentScore(v *entity.TalentAssessmentScore, in UpdateTalentAssessmentScoreInput) error {
	if in.AssessmentID.Set {
		if in.AssessmentID.Value == nil {
			return hrerrors.Invalid("assessment_id", "must not be null")
		}
		v.AssessmentID = *in.AssessmentID.Value
	}
	if in.FrameworkID.Set {
		if in.FrameworkID.Value == nil {
			return hrerrors.Invalid("framework_id", "must not be null")
		}
		v.FrameworkID = *in.FrameworkID.Value
	}
	if in.DimensionID.Set {
		if in.DimensionID.Value == nil {
			return hrerrors.Invalid("dimension_id", "must not be null")
		}
		v.DimensionID = *in.DimensionID.Value
	}
	if in.Score.Set {
		if in.Score.Value == nil {
			return hrerrors.Invalid("score", "must not be null")
		}
		v.Score = *in.Score.Value
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
	return nil
}
func specTalentAssessmentScore() operation.Specification[entity.TalentAssessmentScore, string] {
	return operation.Specification[entity.TalentAssessmentScore, string]{
		Domain: "talent", Table: "hr_talent_assessment_scores",
		Key:         func(v *entity.TalentAssessmentScore) string { return v.ID },
		KeyString:   func(key string) string { return fmt.Sprint(key) },
		ValidateKey: func(key string) error { return entity.ValidateUUID("id", key) },
		AggregateID: func(v *entity.TalentAssessmentScore) string { return v.ID },
		Initialize: func(v *entity.TalentAssessmentScore, actor operation.Actor, now time.Time, id string) {
			v.ID = id
		},
		Touch: func(v *entity.TalentAssessmentScore, actor operation.Actor, now time.Time) {},
		Attributes: func(v *entity.TalentAssessmentScore) map[string]string {
			a := map[string]string{}
			a["id"] = fmt.Sprint(v.ID)
			a["assessment_id"] = fmt.Sprint(v.AssessmentID)
			a["framework_id"] = fmt.Sprint(v.FrameworkID)
			a["dimension_id"] = fmt.Sprint(v.DimensionID)
			a["evidence_document_version_reference_id"] = fmt.Sprint(v.EvidenceDocumentVersionReferenceID)
			return a
		},
		Guard: func(action string, old, next *entity.TalentAssessmentScore) error {
			return nil
		},
	}
}
func createTalentCalibrationSession(in CreateTalentCalibrationSessionInput) (*entity.TalentCalibrationSession, error) {
	v := &entity.TalentCalibrationSession{}
	v.TalentReviewCycleID = in.TalentReviewCycleID
	v.SessionCode = in.SessionCode
	v.ChairPrincipalID = in.ChairPrincipalID
	v.ScheduledAt = in.ScheduledAt
	if in.Status != nil {
		v.Status = *in.Status
	} else {
		v.Status = "planned"
	}
	v.ClosedAt = in.ClosedAt
	v.MinutesDocumentVersionReferenceID = in.MinutesDocumentVersionReferenceID
	return v, nil
}
func patchTalentCalibrationSession(v *entity.TalentCalibrationSession, in UpdateTalentCalibrationSessionInput) error {
	if in.TalentReviewCycleID.Set {
		if in.TalentReviewCycleID.Value == nil {
			return hrerrors.Invalid("talent_review_cycle_id", "must not be null")
		}
		v.TalentReviewCycleID = *in.TalentReviewCycleID.Value
	}
	if in.SessionCode.Set {
		if in.SessionCode.Value == nil {
			return hrerrors.Invalid("session_code", "must not be null")
		}
		v.SessionCode = *in.SessionCode.Value
	}
	if in.ChairPrincipalID.Set {
		if in.ChairPrincipalID.Value == nil {
			return hrerrors.Invalid("chair_principal_id", "must not be null")
		}
		v.ChairPrincipalID = *in.ChairPrincipalID.Value
	}
	if in.ScheduledAt.Set {
		if in.ScheduledAt.Value == nil {
			return hrerrors.Invalid("scheduled_at", "must not be null")
		}
		v.ScheduledAt = *in.ScheduledAt.Value
	}
	if in.Status.Set {
		if in.Status.Value == nil {
			return hrerrors.Invalid("status", "must not be null")
		}
		v.Status = *in.Status.Value
	}
	if in.ClosedAt.Set {
		v.ClosedAt = in.ClosedAt.Value
	}
	if in.MinutesDocumentVersionReferenceID.Set {
		v.MinutesDocumentVersionReferenceID = in.MinutesDocumentVersionReferenceID.Value
	}
	return nil
}
func specTalentCalibrationSession() operation.Specification[entity.TalentCalibrationSession, string] {
	return operation.Specification[entity.TalentCalibrationSession, string]{
		Domain: "talent", Table: "hr_talent_calibration_sessions",
		Key:         func(v *entity.TalentCalibrationSession) string { return v.ID },
		KeyString:   func(key string) string { return fmt.Sprint(key) },
		ValidateKey: func(key string) error { return entity.ValidateUUID("id", key) },
		AggregateID: func(v *entity.TalentCalibrationSession) string { return v.ID },
		Initialize: func(v *entity.TalentCalibrationSession, actor operation.Actor, now time.Time, id string) {
			v.ID = id
		},
		Touch: func(v *entity.TalentCalibrationSession, actor operation.Actor, now time.Time) {},
		Attributes: func(v *entity.TalentCalibrationSession) map[string]string {
			a := map[string]string{}
			a["id"] = fmt.Sprint(v.ID)
			a["talent_review_cycle_id"] = fmt.Sprint(v.TalentReviewCycleID)
			a["chair_principal_id"] = fmt.Sprint(v.ChairPrincipalID)
			a["status"] = fmt.Sprint(v.Status)
			if v.MinutesDocumentVersionReferenceID != nil {
				a["minutes_document_version_reference_id"] = fmt.Sprint(*v.MinutesDocumentVersionReferenceID)
			}
			return a
		},
		Guard: func(action string, old, next *entity.TalentCalibrationSession) error {
			if action == "update" && !reflect.DeepEqual(old.Status, next.Status) {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "create" && string(next.Status) != "planned" {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "create" && next.ClosedAt != nil {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "update" && !reflect.DeepEqual(old.ClosedAt, next.ClosedAt) {
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
func createTalentCalibrationDecision(in CreateTalentCalibrationDecisionInput) (*entity.TalentCalibrationDecision, error) {
	v := &entity.TalentCalibrationDecision{}
	v.SessionID = in.SessionID
	v.TalentReviewCycleID = in.TalentReviewCycleID
	v.AssessmentID = in.AssessmentID
	v.DimensionID = in.DimensionID
	v.OriginalScore = in.OriginalScore
	v.CalibratedScore = in.CalibratedScore
	v.RationaleDocumentVersionReferenceID = in.RationaleDocumentVersionReferenceID
	v.RationaleDigest = in.RationaleDigest
	return v, nil
}
func patchTalentCalibrationDecision(v *entity.TalentCalibrationDecision, in UpdateTalentCalibrationDecisionInput) error {
	if in.SessionID.Set {
		if in.SessionID.Value == nil {
			return hrerrors.Invalid("session_id", "must not be null")
		}
		v.SessionID = *in.SessionID.Value
	}
	if in.TalentReviewCycleID.Set {
		if in.TalentReviewCycleID.Value == nil {
			return hrerrors.Invalid("talent_review_cycle_id", "must not be null")
		}
		v.TalentReviewCycleID = *in.TalentReviewCycleID.Value
	}
	if in.AssessmentID.Set {
		if in.AssessmentID.Value == nil {
			return hrerrors.Invalid("assessment_id", "must not be null")
		}
		v.AssessmentID = *in.AssessmentID.Value
	}
	if in.DimensionID.Set {
		if in.DimensionID.Value == nil {
			return hrerrors.Invalid("dimension_id", "must not be null")
		}
		v.DimensionID = *in.DimensionID.Value
	}
	if in.OriginalScore.Set {
		if in.OriginalScore.Value == nil {
			return hrerrors.Invalid("original_score", "must not be null")
		}
		v.OriginalScore = *in.OriginalScore.Value
	}
	if in.CalibratedScore.Set {
		if in.CalibratedScore.Value == nil {
			return hrerrors.Invalid("calibrated_score", "must not be null")
		}
		v.CalibratedScore = *in.CalibratedScore.Value
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
	return nil
}
func specTalentCalibrationDecision() operation.Specification[entity.TalentCalibrationDecision, string] {
	return operation.Specification[entity.TalentCalibrationDecision, string]{
		Domain: "talent", Table: "hr_talent_calibration_decisions",
		Key:         func(v *entity.TalentCalibrationDecision) string { return v.ID },
		KeyString:   func(key string) string { return fmt.Sprint(key) },
		ValidateKey: func(key string) error { return entity.ValidateUUID("id", key) },
		AggregateID: func(v *entity.TalentCalibrationDecision) string { return v.ID },
		Initialize: func(v *entity.TalentCalibrationDecision, actor operation.Actor, now time.Time, id string) {
			v.ID = id
			v.DecidedByPrincipalID = actor.PrincipalID
			v.DecidedAt = now
		},
		Touch: func(v *entity.TalentCalibrationDecision, actor operation.Actor, now time.Time) {},
		Attributes: func(v *entity.TalentCalibrationDecision) map[string]string {
			a := map[string]string{}
			a["id"] = fmt.Sprint(v.ID)
			a["session_id"] = fmt.Sprint(v.SessionID)
			a["talent_review_cycle_id"] = fmt.Sprint(v.TalentReviewCycleID)
			a["assessment_id"] = fmt.Sprint(v.AssessmentID)
			a["dimension_id"] = fmt.Sprint(v.DimensionID)
			a["rationale_document_version_reference_id"] = fmt.Sprint(v.RationaleDocumentVersionReferenceID)
			a["decided_by_principal_id"] = fmt.Sprint(v.DecidedByPrincipalID)
			return a
		},
		Guard: func(action string, old, next *entity.TalentCalibrationDecision) error {
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
func createTalentPool(in CreateTalentPoolInput) (*entity.TalentPool, error) {
	v := &entity.TalentPool{}
	v.PoolCode = in.PoolCode
	v.PoolVersion = in.PoolVersion
	v.PoolName = in.PoolName
	v.LegalEntityReferenceID = in.LegalEntityReferenceID
	v.TargetJobID = in.TargetJobID
	v.EligibilityDocumentVersionReferenceID = in.EligibilityDocumentVersionReferenceID
	v.EligibilityDigest = in.EligibilityDigest
	v.ValidFrom = in.ValidFrom
	v.ValidTo = in.ValidTo
	v.OwnerPrincipalID = in.OwnerPrincipalID
	return v, nil
}
func patchTalentPool(v *entity.TalentPool, in UpdateTalentPoolInput) error {
	if in.PoolCode.Set {
		if in.PoolCode.Value == nil {
			return hrerrors.Invalid("pool_code", "must not be null")
		}
		v.PoolCode = *in.PoolCode.Value
	}
	if in.PoolVersion.Set {
		if in.PoolVersion.Value == nil {
			return hrerrors.Invalid("pool_version", "must not be null")
		}
		v.PoolVersion = *in.PoolVersion.Value
	}
	if in.PoolName.Set {
		if in.PoolName.Value == nil {
			return hrerrors.Invalid("pool_name", "must not be null")
		}
		v.PoolName = *in.PoolName.Value
	}
	if in.LegalEntityReferenceID.Set {
		if in.LegalEntityReferenceID.Value == nil {
			return hrerrors.Invalid("legal_entity_reference_id", "must not be null")
		}
		v.LegalEntityReferenceID = *in.LegalEntityReferenceID.Value
	}
	if in.TargetJobID.Set {
		v.TargetJobID = in.TargetJobID.Value
	}
	if in.EligibilityDocumentVersionReferenceID.Set {
		if in.EligibilityDocumentVersionReferenceID.Value == nil {
			return hrerrors.Invalid("eligibility_document_version_reference_id", "must not be null")
		}
		v.EligibilityDocumentVersionReferenceID = *in.EligibilityDocumentVersionReferenceID.Value
	}
	if in.EligibilityDigest.Set {
		if in.EligibilityDigest.Value == nil {
			v.EligibilityDigest = nil
		} else {
			v.EligibilityDigest = *in.EligibilityDigest.Value
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
	if in.OwnerPrincipalID.Set {
		if in.OwnerPrincipalID.Value == nil {
			return hrerrors.Invalid("owner_principal_id", "must not be null")
		}
		v.OwnerPrincipalID = *in.OwnerPrincipalID.Value
	}
	return nil
}
func specTalentPool() operation.Specification[entity.TalentPool, string] {
	return operation.Specification[entity.TalentPool, string]{
		Domain: "talent", Table: "hr_talent_pools",
		Key:         func(v *entity.TalentPool) string { return v.ID },
		KeyString:   func(key string) string { return fmt.Sprint(key) },
		ValidateKey: func(key string) error { return entity.ValidateUUID("id", key) },
		AggregateID: func(v *entity.TalentPool) string { return v.ID },
		Initialize: func(v *entity.TalentPool, actor operation.Actor, now time.Time, id string) {
			v.ID = id
		},
		Touch: func(v *entity.TalentPool, actor operation.Actor, now time.Time) {},
		Attributes: func(v *entity.TalentPool) map[string]string {
			a := map[string]string{}
			a["id"] = fmt.Sprint(v.ID)
			a["legal_entity_reference_id"] = fmt.Sprint(v.LegalEntityReferenceID)
			if v.TargetJobID != nil {
				a["target_job_id"] = fmt.Sprint(*v.TargetJobID)
			}
			a["eligibility_document_version_reference_id"] = fmt.Sprint(v.EligibilityDocumentVersionReferenceID)
			a["owner_principal_id"] = fmt.Sprint(v.OwnerPrincipalID)
			return a
		},
		Guard: func(action string, old, next *entity.TalentPool) error {
			return nil
		},
	}
}
func createTalentPoolMembership(in CreateTalentPoolMembershipInput) (*entity.TalentPoolMembership, error) {
	v := &entity.TalentPoolMembership{}
	v.TalentPoolID = in.TalentPoolID
	v.WorkerID = in.WorkerID
	v.AssessmentID = in.AssessmentID
	v.NominatedByPrincipalID = in.NominatedByPrincipalID
	if in.Status != nil {
		v.Status = *in.Status
	} else {
		v.Status = "nominated"
	}
	v.ValidFrom = in.ValidFrom
	v.ValidTo = in.ValidTo
	v.ApprovedByPrincipalID = in.ApprovedByPrincipalID
	v.ApprovedAt = in.ApprovedAt
	v.ExitReasonCode = in.ExitReasonCode
	return v, nil
}
func patchTalentPoolMembership(v *entity.TalentPoolMembership, in UpdateTalentPoolMembershipInput) error {
	if in.TalentPoolID.Set {
		if in.TalentPoolID.Value == nil {
			return hrerrors.Invalid("talent_pool_id", "must not be null")
		}
		v.TalentPoolID = *in.TalentPoolID.Value
	}
	if in.WorkerID.Set {
		if in.WorkerID.Value == nil {
			return hrerrors.Invalid("worker_id", "must not be null")
		}
		v.WorkerID = *in.WorkerID.Value
	}
	if in.AssessmentID.Set {
		if in.AssessmentID.Value == nil {
			return hrerrors.Invalid("assessment_id", "must not be null")
		}
		v.AssessmentID = *in.AssessmentID.Value
	}
	if in.NominatedByPrincipalID.Set {
		if in.NominatedByPrincipalID.Value == nil {
			return hrerrors.Invalid("nominated_by_principal_id", "must not be null")
		}
		v.NominatedByPrincipalID = *in.NominatedByPrincipalID.Value
	}
	if in.Status.Set {
		if in.Status.Value == nil {
			return hrerrors.Invalid("status", "must not be null")
		}
		v.Status = *in.Status.Value
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
	if in.ApprovedByPrincipalID.Set {
		v.ApprovedByPrincipalID = in.ApprovedByPrincipalID.Value
	}
	if in.ApprovedAt.Set {
		v.ApprovedAt = in.ApprovedAt.Value
	}
	if in.ExitReasonCode.Set {
		v.ExitReasonCode = in.ExitReasonCode.Value
	}
	return nil
}
func specTalentPoolMembership() operation.Specification[entity.TalentPoolMembership, string] {
	return operation.Specification[entity.TalentPoolMembership, string]{
		Domain: "talent", Table: "hr_talent_pool_memberships",
		Key:         func(v *entity.TalentPoolMembership) string { return v.ID },
		KeyString:   func(key string) string { return fmt.Sprint(key) },
		ValidateKey: func(key string) error { return entity.ValidateUUID("id", key) },
		AggregateID: func(v *entity.TalentPoolMembership) string { return v.ID },
		Initialize: func(v *entity.TalentPoolMembership, actor operation.Actor, now time.Time, id string) {
			v.ID = id
		},
		Touch: func(v *entity.TalentPoolMembership, actor operation.Actor, now time.Time) {},
		Attributes: func(v *entity.TalentPoolMembership) map[string]string {
			a := map[string]string{}
			a["id"] = fmt.Sprint(v.ID)
			a["talent_pool_id"] = fmt.Sprint(v.TalentPoolID)
			a["worker_id"] = fmt.Sprint(v.WorkerID)
			a["assessment_id"] = fmt.Sprint(v.AssessmentID)
			a["nominated_by_principal_id"] = fmt.Sprint(v.NominatedByPrincipalID)
			a["status"] = fmt.Sprint(v.Status)
			if v.ApprovedByPrincipalID != nil {
				a["approved_by_principal_id"] = fmt.Sprint(*v.ApprovedByPrincipalID)
			}
			return a
		},
		Guard: func(action string, old, next *entity.TalentPoolMembership) error {
			if action == "update" && !reflect.DeepEqual(old.Status, next.Status) {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "create" && string(next.Status) != "nominated" {
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
			if action != "create" && old.Status == "active" {
				return hrerrors.ErrWorkflowRequired
			}
			return nil
		},
	}
}
