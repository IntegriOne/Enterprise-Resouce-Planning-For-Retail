// Sumber model adalah baseline SQL HRM; aturan bisnis tetap milik service.
package repository

import (
	"context"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
)

// EngagementSurveyRepository adalah port persistensi hr_engagement_surveys.
type EngagementSurveyRepository interface {
	Repository[entity.EngagementSurvey, string]
	InsertEngagementSurveys(context.Context, *entity.EngagementSurvey) error
	UpdateEngagementSurveys(context.Context, *entity.EngagementSurvey) error
	DeleteEngagementSurveysByUUID(context.Context, string) error
	FindEngagementSurveysByUUID(context.Context, string) (*entity.EngagementSurvey, error)
	FindEngagementSurveysByUUIDForUpdate(context.Context, string) (*entity.EngagementSurvey, error)
	GetAllEngagementSurveys(context.Context, int, int) ([]entity.EngagementSurvey, error)
	CountEngagementSurveys(context.Context) (int64, error)
	SelectEngagementSurveysWhere(context.Context, []Filter, int, int) ([]entity.EngagementSurvey, error)
	CountEngagementSurveysWhere(context.Context, []Filter) (int64, error)
	FindEngagementSurveysBySurveyCode(context.Context, string, int, int) ([]entity.EngagementSurvey, error)
	FindEngagementSurveysByLegalEntityReferenceID(context.Context, string, int, int) ([]entity.EngagementSurvey, error)
	FindEngagementSurveysByTopicCode(context.Context, string, int, int) ([]entity.EngagementSurvey, error)
	FindEngagementSurveysByQuestionnaireDocumentVersionReferenceID(context.Context, string, int, int) ([]entity.EngagementSurvey, error)
	FindBySurveyCode(ctx context.Context, value string, limit, offset int) ([]entity.EngagementSurvey, error)
	FindByLegalEntityReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.EngagementSurvey, error)
	FindByTopicCode(ctx context.Context, value string, limit, offset int) ([]entity.EngagementSurvey, error)
	FindByQuestionnaireDocumentVersionReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.EngagementSurvey, error)
}

// EngagementSurveyResultRepository adalah port persistensi hr_engagement_survey_results.
type EngagementSurveyResultRepository interface {
	CountStoredBySurveyID(context.Context, string) (int64, error)
	Repository[entity.EngagementSurveyResult, string]
	InsertEngagementSurveyResults(context.Context, *entity.EngagementSurveyResult) error
	UpdateEngagementSurveyResults(context.Context, *entity.EngagementSurveyResult) error
	DeleteEngagementSurveyResultsByUUID(context.Context, string) error
	FindEngagementSurveyResultsByUUID(context.Context, string) (*entity.EngagementSurveyResult, error)
	FindEngagementSurveyResultsByUUIDForUpdate(context.Context, string) (*entity.EngagementSurveyResult, error)
	GetAllEngagementSurveyResults(context.Context, int, int) ([]entity.EngagementSurveyResult, error)
	CountEngagementSurveyResults(context.Context) (int64, error)
	SelectEngagementSurveyResultsWhere(context.Context, []Filter, int, int) ([]entity.EngagementSurveyResult, error)
	CountEngagementSurveyResultsWhere(context.Context, []Filter) (int64, error)
	FindEngagementSurveyResultsBySurveyID(context.Context, string, int, int) ([]entity.EngagementSurveyResult, error)
	FindEngagementSurveyResultsByOrganizationUnitReferenceID(context.Context, *string, int, int) ([]entity.EngagementSurveyResult, error)
	FindEngagementSurveyResultsByQuestionCode(context.Context, string, int, int) ([]entity.EngagementSurveyResult, error)
	FindBySurveyID(ctx context.Context, value string, limit, offset int) ([]entity.EngagementSurveyResult, error)
	FindByOrganizationUnitReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.EngagementSurveyResult, error)
	FindByQuestionCode(ctx context.Context, value string, limit, offset int) ([]entity.EngagementSurveyResult, error)
}
