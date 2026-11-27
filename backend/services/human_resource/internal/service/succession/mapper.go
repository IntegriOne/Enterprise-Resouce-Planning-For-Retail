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

func createSuccessionPlan(in CreateSuccessionPlanInput) (*entity.SuccessionPlan, error) {
	v := &entity.SuccessionPlan{}
	v.PositionID = in.PositionID
	v.PlanCode = in.PlanCode
	v.PlanVersion = in.PlanVersion
	v.ReviewDate = in.ReviewDate
	if in.ConfidentialityClass != nil {
		v.ConfidentialityClass = *in.ConfidentialityClass
	} else {
		v.ConfidentialityClass = "highly_restricted"
	}
	v.ApprovedByPrincipalID = in.ApprovedByPrincipalID
	v.ApprovedAt = in.ApprovedAt
	v.SuccessionRisk = in.SuccessionRisk
	return v, nil
}
func patchSuccessionPlan(v *entity.SuccessionPlan, in UpdateSuccessionPlanInput) error {
	if in.PositionID.Set {
		if in.PositionID.Value == nil {
			return hrerrors.Invalid("position_id", "must not be null")
		}
		v.PositionID = *in.PositionID.Value
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
	if in.ReviewDate.Set {
		if in.ReviewDate.Value == nil {
			return hrerrors.Invalid("review_date", "must not be null")
		}
		v.ReviewDate = *in.ReviewDate.Value
	}
	if in.ConfidentialityClass.Set {
		if in.ConfidentialityClass.Value == nil {
			return hrerrors.Invalid("confidentiality_class", "must not be null")
		}
		v.ConfidentialityClass = *in.ConfidentialityClass.Value
	}
	if in.ApprovedByPrincipalID.Set {
		v.ApprovedByPrincipalID = in.ApprovedByPrincipalID.Value
	}
	if in.ApprovedAt.Set {
		v.ApprovedAt = in.ApprovedAt.Value
	}
	if in.SuccessionRisk.Set {
		v.SuccessionRisk = in.SuccessionRisk.Value
	}
	return nil
}
func specSuccessionPlan() operation.Specification[entity.SuccessionPlan, string] {
	return operation.Specification[entity.SuccessionPlan, string]{
		Domain: "succession", Table: "hr_succession_plans",
		Key:         func(v *entity.SuccessionPlan) string { return v.ID },
		KeyString:   func(key string) string { return fmt.Sprint(key) },
		ValidateKey: func(key string) error { return entity.ValidateUUID("id", key) },
		AggregateID: func(v *entity.SuccessionPlan) string { return v.ID },
		Initialize: func(v *entity.SuccessionPlan, actor operation.Actor, now time.Time, id string) {
			v.ID = id
			v.CreatedByPrincipalID, v.CreatedAt = &actor.PrincipalID, &now
		},
		Touch: func(v *entity.SuccessionPlan, actor operation.Actor, now time.Time) {},
		Attributes: func(v *entity.SuccessionPlan) map[string]string {
			a := map[string]string{}
			a["id"] = fmt.Sprint(v.ID)
			a["position_id"] = fmt.Sprint(v.PositionID)
			if v.ApprovedByPrincipalID != nil {
				a["approved_by_principal_id"] = fmt.Sprint(*v.ApprovedByPrincipalID)
			}
			return a
		},
		Guard: func(action string, old, next *entity.SuccessionPlan) error {
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
			return nil
		},
	}
}
func createSuccessionCandidate(in CreateSuccessionCandidateInput) (*entity.SuccessionCandidate, error) {
	v := &entity.SuccessionCandidate{}
	v.SuccessionPlanID = in.SuccessionPlanID
	v.CandidateWorkerID = in.CandidateWorkerID
	v.ReadinessCode = in.ReadinessCode
	v.ReadinessAssessedAt = in.ReadinessAssessedAt
	v.ReadinessRationale = in.ReadinessRationale
	v.TalentAssessmentID = in.TalentAssessmentID
	v.DevelopmentPlanID = in.DevelopmentPlanID
	v.RiskOfLoss = in.RiskOfLoss
	return v, nil
}
func patchSuccessionCandidate(v *entity.SuccessionCandidate, in UpdateSuccessionCandidateInput) error {
	if in.SuccessionPlanID.Set {
		if in.SuccessionPlanID.Value == nil {
			return hrerrors.Invalid("succession_plan_id", "must not be null")
		}
		v.SuccessionPlanID = *in.SuccessionPlanID.Value
	}
	if in.CandidateWorkerID.Set {
		if in.CandidateWorkerID.Value == nil {
			return hrerrors.Invalid("candidate_worker_id", "must not be null")
		}
		v.CandidateWorkerID = *in.CandidateWorkerID.Value
	}
	if in.ReadinessCode.Set {
		if in.ReadinessCode.Value == nil {
			return hrerrors.Invalid("readiness_code", "must not be null")
		}
		v.ReadinessCode = *in.ReadinessCode.Value
	}
	if in.ReadinessAssessedAt.Set {
		if in.ReadinessAssessedAt.Value == nil {
			return hrerrors.Invalid("readiness_assessed_at", "must not be null")
		}
		v.ReadinessAssessedAt = *in.ReadinessAssessedAt.Value
	}
	if in.ReadinessRationale.Set {
		v.ReadinessRationale = in.ReadinessRationale.Value
	}
	if in.TalentAssessmentID.Set {
		v.TalentAssessmentID = in.TalentAssessmentID.Value
	}
	if in.DevelopmentPlanID.Set {
		v.DevelopmentPlanID = in.DevelopmentPlanID.Value
	}
	if in.RiskOfLoss.Set {
		v.RiskOfLoss = in.RiskOfLoss.Value
	}
	return nil
}
func specSuccessionCandidate() operation.Specification[entity.SuccessionCandidate, string] {
	return operation.Specification[entity.SuccessionCandidate, string]{
		Domain: "succession", Table: "hr_succession_candidates",
		Key:         func(v *entity.SuccessionCandidate) string { return v.ID },
		KeyString:   func(key string) string { return fmt.Sprint(key) },
		ValidateKey: func(key string) error { return entity.ValidateUUID("id", key) },
		AggregateID: func(v *entity.SuccessionCandidate) string { return v.ID },
		Initialize: func(v *entity.SuccessionCandidate, actor operation.Actor, now time.Time, id string) {
			v.ID = id
		},
		Touch: func(v *entity.SuccessionCandidate, actor operation.Actor, now time.Time) {},
		Attributes: func(v *entity.SuccessionCandidate) map[string]string {
			a := map[string]string{}
			a["id"] = fmt.Sprint(v.ID)
			a["succession_plan_id"] = fmt.Sprint(v.SuccessionPlanID)
			a["candidate_worker_id"] = fmt.Sprint(v.CandidateWorkerID)
			if v.TalentAssessmentID != nil {
				a["talent_assessment_id"] = fmt.Sprint(*v.TalentAssessmentID)
			}
			if v.DevelopmentPlanID != nil {
				a["development_plan_id"] = fmt.Sprint(*v.DevelopmentPlanID)
			}
			return a
		},
		Guard: func(action string, old, next *entity.SuccessionCandidate) error {
			return nil
		},
	}
}
