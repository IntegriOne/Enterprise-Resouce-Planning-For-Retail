package security

import (
	"context"
	"reflect"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository"
)

type engagementSurveyReader interface {
	SelectWhere(context.Context, []domainrepo.Filter, int, int) ([]entity.EngagementSurvey, error)
}

type engagementSurveyResultCounter interface {
	CountStoredBySurveyID(context.Context, string) (int64, error)
}

type engagementSurveyValidator struct {
	surveys    engagementSurveyReader
	results    engagementSurveyResultCounter
	foundation FoundationVerifier
}

// EngagementValidators mendaftarkan validator survei engagement HRM beserta ambang populasi minimumnya.
func EngagementValidators(repos *repository.Registry, foundation FoundationVerifier) (map[string]MutationValidator, error) {
	if repos == nil ||
		operation.RequireDependency(repos.EngagementSurvey) != nil ||
		operation.RequireDependency(repos.EngagementSurveyResult) != nil ||
		operation.RequireDependency(foundation) != nil {
		return nil, hrerrors.ErrUnavailable
	}
	return map[string]MutationValidator{
		"hr_engagement_surveys": &engagementSurveyValidator{
			surveys:    repos.EngagementSurvey,
			results:    repos.EngagementSurveyResult,
			foundation: foundation,
		},
	}, nil
}

// ValidateMutation menjaga ambang anonimitas, keunikan kode+versi, serta membekukan survei yang telah berhasil.
func (v *engagementSurveyValidator) ValidateMutation(ctx context.Context, table string, before, after entity.Validatable) error {
	if v == nil || table != "hr_engagement_surveys" ||
		operation.RequireDependency(v.surveys) != nil ||
		operation.RequireDependency(v.results) != nil ||
		operation.RequireDependency(v.foundation) != nil {
		return hrerrors.ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	var previous *entity.EngagementSurvey
	if before != nil {
		var ok bool
		previous, ok = before.(*entity.EngagementSurvey)
		if !ok || previous == nil {
			return hrerrors.ErrInvalidInput
		}
	}
	if after == nil {
		if previous == nil {
			return hrerrors.ErrInvalidInput
		}
		return v.rejectIfHasResults(ctx, previous.ID)
	}
	survey, ok := after.(*entity.EngagementSurvey)
	if !ok || survey == nil || (previous != nil && previous.ID != survey.ID) {
		return hrerrors.ErrInvalidInput
	}
	if !supportedEngagementTopic(survey.TopicCode) || survey.MinimumReportingPopulation < 5 {
		return hrerrors.ErrInvalidInput
	}
	if err := verifyFoundationLegalEntity(ctx, v.foundation, survey.LegalEntityReferenceID); err != nil {
		return err
	}
	matches, err := v.surveys.SelectWhere(ctx, []domainrepo.Filter{
		{Column: "legal_entity_reference_id", Values: []any{survey.LegalEntityReferenceID}},
		{Column: "survey_code", Values: []any{survey.SurveyCode}},
		{Column: "survey_version", Values: []any{survey.SurveyVersion}},
	}, 2, 0)
	if err != nil {
		return err
	}
	for _, existing := range matches {
		if existing.ID != survey.ID {
			return hrerrors.ErrAlreadyExists
		}
	}
	if previous != nil && !reflect.DeepEqual(previous, survey) {
		return v.rejectIfHasResults(ctx, survey.ID)
	}
	return nil
}

func (v *engagementSurveyValidator) rejectIfHasResults(ctx context.Context, surveyID string) error {
	count, err := v.results.CountStoredBySurveyID(ctx, surveyID)
	if err != nil {
		return err
	}
	if count > 0 {
		return hrerrors.ErrReferenceConflict
	}
	return nil
}

func supportedEngagementTopic(value string) bool {
	switch value {
	case "engagement", "wellbeing", "inclusion", "exit_experience":
		return true
	default:
		return false
	}
}
