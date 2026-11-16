package engagement

import (
	"context"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
)

// Operasi berikut menyebut objek dan kolom secara eksplisit untuk memudahkan penelusuran.
func (r *EngagementSurveyRepository) InsertEngagementSurveys(ctx context.Context, value *entity.EngagementSurvey) error {
	return r.Insert(ctx, value)
}
func (r *EngagementSurveyRepository) UpdateEngagementSurveys(ctx context.Context, value *entity.EngagementSurvey) error {
	return r.Update(ctx, value)
}
func (r *EngagementSurveyRepository) DeleteEngagementSurveysByUUID(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
func (r *EngagementSurveyRepository) FindEngagementSurveysByUUID(ctx context.Context, id string) (*entity.EngagementSurvey, error) {
	return r.FindByID(ctx, id)
}
func (r *EngagementSurveyRepository) FindEngagementSurveysByUUIDForUpdate(ctx context.Context, id string) (*entity.EngagementSurvey, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *EngagementSurveyRepository) GetAllEngagementSurveys(ctx context.Context, limit, offset int) ([]entity.EngagementSurvey, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *EngagementSurveyRepository) CountEngagementSurveys(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *EngagementSurveyRepository) SelectEngagementSurveysWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.EngagementSurvey, error) {
	return r.SelectWhere(ctx, filters, limit, offset)
}
func (r *EngagementSurveyRepository) CountEngagementSurveysWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, filters)
}
func (r *EngagementSurveyRepository) FindEngagementSurveysBySurveyCode(ctx context.Context, value string, limit, offset int) ([]entity.EngagementSurvey, error) {
	return r.FindBySurveyCode(ctx, value, limit, offset)
}
func (r *EngagementSurveyRepository) FindEngagementSurveysByLegalEntityReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.EngagementSurvey, error) {
	return r.FindByLegalEntityReferenceID(ctx, value, limit, offset)
}
func (r *EngagementSurveyRepository) FindEngagementSurveysByTopicCode(ctx context.Context, value string, limit, offset int) ([]entity.EngagementSurvey, error) {
	return r.FindByTopicCode(ctx, value, limit, offset)
}
func (r *EngagementSurveyRepository) FindEngagementSurveysByQuestionnaireDocumentVersionReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.EngagementSurvey, error) {
	return r.FindByQuestionnaireDocumentVersionReferenceID(ctx, value, limit, offset)
}

func (r *EngagementSurveyResultRepository) InsertEngagementSurveyResults(ctx context.Context, value *entity.EngagementSurveyResult) error {
	return r.Insert(ctx, value)
}
func (r *EngagementSurveyResultRepository) UpdateEngagementSurveyResults(ctx context.Context, value *entity.EngagementSurveyResult) error {
	return r.Update(ctx, value)
}
func (r *EngagementSurveyResultRepository) DeleteEngagementSurveyResultsByUUID(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
func (r *EngagementSurveyResultRepository) FindEngagementSurveyResultsByUUID(ctx context.Context, id string) (*entity.EngagementSurveyResult, error) {
	return r.FindByID(ctx, id)
}
func (r *EngagementSurveyResultRepository) FindEngagementSurveyResultsByUUIDForUpdate(ctx context.Context, id string) (*entity.EngagementSurveyResult, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *EngagementSurveyResultRepository) GetAllEngagementSurveyResults(ctx context.Context, limit, offset int) ([]entity.EngagementSurveyResult, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *EngagementSurveyResultRepository) CountEngagementSurveyResults(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *EngagementSurveyResultRepository) SelectEngagementSurveyResultsWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.EngagementSurveyResult, error) {
	return r.SelectWhere(ctx, filters, limit, offset)
}
func (r *EngagementSurveyResultRepository) CountEngagementSurveyResultsWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, filters)
}
func (r *EngagementSurveyResultRepository) FindEngagementSurveyResultsBySurveyID(ctx context.Context, value string, limit, offset int) ([]entity.EngagementSurveyResult, error) {
	return r.FindBySurveyID(ctx, value, limit, offset)
}
func (r *EngagementSurveyResultRepository) FindEngagementSurveyResultsByOrganizationUnitReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.EngagementSurveyResult, error) {
	return r.FindByOrganizationUnitReferenceID(ctx, value, limit, offset)
}
func (r *EngagementSurveyResultRepository) FindEngagementSurveyResultsByQuestionCode(ctx context.Context, value string, limit, offset int) ([]entity.EngagementSurveyResult, error) {
	return r.FindByQuestionCode(ctx, value, limit, offset)
}
