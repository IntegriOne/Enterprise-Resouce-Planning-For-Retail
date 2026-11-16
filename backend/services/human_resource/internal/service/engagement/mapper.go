// Kerangka awal dibuat dengan alat bantu; kode ini dipelihara langsung per domain.
// Aturan bisnis lintas record tetap dijalankan melalui policy/integrity dan command domain.
package service

import (
	"fmt"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

func createEngagementSurvey(in CreateEngagementSurveyInput) (*entity.EngagementSurvey, error) {
	v := &entity.EngagementSurvey{}
	v.SurveyCode = in.SurveyCode
	v.SurveyVersion = in.SurveyVersion
	v.LegalEntityReferenceID = in.LegalEntityReferenceID
	v.TopicCode = in.TopicCode
	v.OpensAt = in.OpensAt
	v.ClosesAt = in.ClosesAt
	v.QuestionnaireDocumentVersionReferenceID = in.QuestionnaireDocumentVersionReferenceID
	v.QuestionnaireDigest = in.QuestionnaireDigest
	v.MinimumReportingPopulation = in.MinimumReportingPopulation
	return v, nil
}
func patchEngagementSurvey(v *entity.EngagementSurvey, in UpdateEngagementSurveyInput) error {
	if in.SurveyCode.Set {
		if in.SurveyCode.Value == nil {
			return hrerrors.Invalid("survey_code", "must not be null")
		}
		v.SurveyCode = *in.SurveyCode.Value
	}
	if in.SurveyVersion.Set {
		if in.SurveyVersion.Value == nil {
			return hrerrors.Invalid("survey_version", "must not be null")
		}
		v.SurveyVersion = *in.SurveyVersion.Value
	}
	if in.LegalEntityReferenceID.Set {
		if in.LegalEntityReferenceID.Value == nil {
			return hrerrors.Invalid("legal_entity_reference_id", "must not be null")
		}
		v.LegalEntityReferenceID = *in.LegalEntityReferenceID.Value
	}
	if in.TopicCode.Set {
		if in.TopicCode.Value == nil {
			return hrerrors.Invalid("topic_code", "must not be null")
		}
		v.TopicCode = *in.TopicCode.Value
	}
	if in.OpensAt.Set {
		if in.OpensAt.Value == nil {
			return hrerrors.Invalid("opens_at", "must not be null")
		}
		v.OpensAt = *in.OpensAt.Value
	}
	if in.ClosesAt.Set {
		if in.ClosesAt.Value == nil {
			return hrerrors.Invalid("closes_at", "must not be null")
		}
		v.ClosesAt = *in.ClosesAt.Value
	}
	if in.QuestionnaireDocumentVersionReferenceID.Set {
		if in.QuestionnaireDocumentVersionReferenceID.Value == nil {
			return hrerrors.Invalid("questionnaire_document_version_reference_id", "must not be null")
		}
		v.QuestionnaireDocumentVersionReferenceID = *in.QuestionnaireDocumentVersionReferenceID.Value
	}
	if in.QuestionnaireDigest.Set {
		if in.QuestionnaireDigest.Value == nil {
			v.QuestionnaireDigest = nil
		} else {
			v.QuestionnaireDigest = *in.QuestionnaireDigest.Value
		}
	}
	if in.MinimumReportingPopulation.Set {
		if in.MinimumReportingPopulation.Value == nil {
			return hrerrors.Invalid("minimum_reporting_population", "must not be null")
		}
		v.MinimumReportingPopulation = *in.MinimumReportingPopulation.Value
	}
	return nil
}
func specEngagementSurvey() operation.Specification[entity.EngagementSurvey, string] {
	return operation.Specification[entity.EngagementSurvey, string]{
		Domain: "engagement", Table: "hr_engagement_surveys",
		Key:         func(v *entity.EngagementSurvey) string { return v.ID },
		KeyString:   func(key string) string { return fmt.Sprint(key) },
		ValidateKey: func(key string) error { return entity.ValidateUUID("id", key) },
		AggregateID: func(v *entity.EngagementSurvey) string { return v.ID },
		Initialize: func(v *entity.EngagementSurvey, actor operation.Actor, now time.Time, id string) {
			v.ID = id
		},
		Touch: func(v *entity.EngagementSurvey, actor operation.Actor, now time.Time) {},
		Attributes: func(v *entity.EngagementSurvey) map[string]string {
			a := map[string]string{}
			a["id"] = fmt.Sprint(v.ID)
			a["legal_entity_reference_id"] = fmt.Sprint(v.LegalEntityReferenceID)
			a["questionnaire_document_version_reference_id"] = fmt.Sprint(v.QuestionnaireDocumentVersionReferenceID)
			return a
		},
		Guard: func(action string, old, next *entity.EngagementSurvey) error {
			return nil
		},
	}
}
func createEngagementSurveyResult(in CreateEngagementSurveyResultInput) (*entity.EngagementSurveyResult, error) {
	v := &entity.EngagementSurveyResult{}
	v.SurveyID = in.SurveyID
	v.OrganizationUnitReferenceID = in.OrganizationUnitReferenceID
	v.QuestionCode = in.QuestionCode
	v.ResultVersion = in.ResultVersion
	v.RespondentCount = in.RespondentCount
	v.ScoreValue = in.ScoreValue
	v.ScaleMinimum = in.ScaleMinimum
	v.ScaleMaximum = in.ScaleMaximum
	v.SourceDigest = in.SourceDigest
	return v, nil
}
func patchEngagementSurveyResult(v *entity.EngagementSurveyResult, in UpdateEngagementSurveyResultInput) error {
	if in.SurveyID.Set {
		if in.SurveyID.Value == nil {
			return hrerrors.Invalid("survey_id", "must not be null")
		}
		v.SurveyID = *in.SurveyID.Value
	}
	if in.OrganizationUnitReferenceID.Set {
		v.OrganizationUnitReferenceID = in.OrganizationUnitReferenceID.Value
	}
	if in.QuestionCode.Set {
		if in.QuestionCode.Value == nil {
			return hrerrors.Invalid("question_code", "must not be null")
		}
		v.QuestionCode = *in.QuestionCode.Value
	}
	if in.ResultVersion.Set {
		if in.ResultVersion.Value == nil {
			return hrerrors.Invalid("result_version", "must not be null")
		}
		v.ResultVersion = *in.ResultVersion.Value
	}
	if in.RespondentCount.Set {
		if in.RespondentCount.Value == nil {
			return hrerrors.Invalid("respondent_count", "must not be null")
		}
		v.RespondentCount = *in.RespondentCount.Value
	}
	if in.ScoreValue.Set {
		if in.ScoreValue.Value == nil {
			return hrerrors.Invalid("score_value", "must not be null")
		}
		v.ScoreValue = *in.ScoreValue.Value
	}
	if in.ScaleMinimum.Set {
		if in.ScaleMinimum.Value == nil {
			return hrerrors.Invalid("scale_minimum", "must not be null")
		}
		v.ScaleMinimum = *in.ScaleMinimum.Value
	}
	if in.ScaleMaximum.Set {
		if in.ScaleMaximum.Value == nil {
			return hrerrors.Invalid("scale_maximum", "must not be null")
		}
		v.ScaleMaximum = *in.ScaleMaximum.Value
	}
	if in.SourceDigest.Set {
		if in.SourceDigest.Value == nil {
			v.SourceDigest = nil
		} else {
			v.SourceDigest = *in.SourceDigest.Value
		}
	}
	return nil
}
func specEngagementSurveyResult() operation.Specification[entity.EngagementSurveyResult, string] {
	return operation.Specification[entity.EngagementSurveyResult, string]{
		Domain: "engagement", Table: "hr_engagement_survey_results",
		Key:         func(v *entity.EngagementSurveyResult) string { return v.ID },
		KeyString:   func(key string) string { return fmt.Sprint(key) },
		ValidateKey: func(key string) error { return entity.ValidateUUID("id", key) },
		AggregateID: func(v *entity.EngagementSurveyResult) string { return v.ID },
		Initialize: func(v *entity.EngagementSurveyResult, actor operation.Actor, now time.Time, id string) {
			v.ID = id
			v.CalculatedAt = now
		},
		Touch: func(v *entity.EngagementSurveyResult, actor operation.Actor, now time.Time) {},
		Attributes: func(v *entity.EngagementSurveyResult) map[string]string {
			a := map[string]string{}
			a["id"] = fmt.Sprint(v.ID)
			a["survey_id"] = fmt.Sprint(v.SurveyID)
			if v.OrganizationUnitReferenceID != nil {
				a["organization_unit_reference_id"] = fmt.Sprint(*v.OrganizationUnitReferenceID)
			}
			return a
		},
		Guard: func(action string, old, next *entity.EngagementSurveyResult) error {
			return nil
		},
	}
}
