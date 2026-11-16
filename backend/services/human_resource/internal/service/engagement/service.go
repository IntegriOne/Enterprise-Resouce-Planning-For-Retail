// Kerangka awal dibuat dengan alat bantu; kode ini dipelihara langsung per domain.
// Aturan bisnis lintas record tetap dijalankan melalui policy/integrity dan command domain.
package service

import (
	"context"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

type Dependencies struct {
	Runtime                operation.Dependencies
	EngagementSurvey       operation.Repository[entity.EngagementSurvey, string]
	EngagementSurveyResult operation.Repository[entity.EngagementSurveyResult, string]
}

// Service mengoordinasikan operasi domain engagement.
type Service struct {
	engagementSurvey       *operation.Resource[entity.EngagementSurvey, string, *entity.EngagementSurvey]
	engagementSurveyResult *operation.Resource[entity.EngagementSurveyResult, string, *entity.EngagementSurveyResult]
}

// NewService memeriksa seluruh port dan dependency sebelum service dapat dipakai.
func NewService(deps Dependencies) (*Service, error) {
	s := &Service{}
	var err error
	s.engagementSurvey, err = operation.NewResource[entity.EngagementSurvey, string, *entity.EngagementSurvey](deps.EngagementSurvey, deps.Runtime, specEngagementSurvey())
	if err != nil {
		return nil, err
	}
	s.engagementSurveyResult, err = operation.NewResource[entity.EngagementSurveyResult, string, *entity.EngagementSurveyResult](deps.EngagementSurveyResult, deps.Runtime, specEngagementSurveyResult())
	if err != nil {
		return nil, err
	}
	return s, nil
}

// CreateEngagementSurvey menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreateEngagementSurvey(ctx context.Context, in CreateEngagementSurveyInput) (*entity.EngagementSurvey, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.engagementSurvey.Create(ctx, func() (*entity.EngagementSurvey, error) { return createEngagementSurvey(in) })
}

// UpdateEngagementSurvey menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdateEngagementSurvey(ctx context.Context, key string, in UpdateEngagementSurveyInput) (*entity.EngagementSurvey, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.engagementSurvey.Update(ctx, key, func(value *entity.EngagementSurvey) error { return patchEngagementSurvey(value, in) })
}

// DeleteEngagementSurvey menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeleteEngagementSurvey(ctx context.Context, key string) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.engagementSurvey.Delete(ctx, key)
}

// GetEngagementSurveyByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetEngagementSurveyByID(ctx context.Context, key string) (*entity.EngagementSurvey, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.engagementSurvey.Get(ctx, key)
}

// GetAllEngagementSurveys menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllEngagementSurveys(ctx context.Context, limit, offset int) (operation.Page[entity.EngagementSurvey], error) {
	if s == nil {
		return operation.Page[entity.EngagementSurvey]{}, hrerrors.ErrUnavailable
	}
	return s.engagementSurvey.List(ctx, nil, limit, offset)
}

// FindEngagementSurveyBySurveyCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindEngagementSurveyBySurveyCode(ctx context.Context, value string, limit, offset int) (operation.Page[entity.EngagementSurvey], error) {
	if s == nil {
		return operation.Page[entity.EngagementSurvey]{}, hrerrors.ErrUnavailable
	}
	return s.engagementSurvey.List(ctx, []domainrepo.Filter{{Column: "survey_code", Values: []any{value}}}, limit, offset)
}

// FindEngagementSurveyByLegalEntityReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindEngagementSurveyByLegalEntityReferenceID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.EngagementSurvey], error) {
	if s == nil {
		return operation.Page[entity.EngagementSurvey]{}, hrerrors.ErrUnavailable
	}
	return s.engagementSurvey.List(ctx, []domainrepo.Filter{{Column: "legal_entity_reference_id", Values: []any{value}}}, limit, offset)
}

// FindEngagementSurveyByTopicCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindEngagementSurveyByTopicCode(ctx context.Context, value string, limit, offset int) (operation.Page[entity.EngagementSurvey], error) {
	if s == nil {
		return operation.Page[entity.EngagementSurvey]{}, hrerrors.ErrUnavailable
	}
	return s.engagementSurvey.List(ctx, []domainrepo.Filter{{Column: "topic_code", Values: []any{value}}}, limit, offset)
}

// FindEngagementSurveyByQuestionnaireDocumentVersionReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindEngagementSurveyByQuestionnaireDocumentVersionReferenceID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.EngagementSurvey], error) {
	if s == nil {
		return operation.Page[entity.EngagementSurvey]{}, hrerrors.ErrUnavailable
	}
	return s.engagementSurvey.List(ctx, []domainrepo.Filter{{Column: "questionnaire_document_version_reference_id", Values: []any{value}}}, limit, offset)
}

// CreateEngagementSurveyResult menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreateEngagementSurveyResult(ctx context.Context, in CreateEngagementSurveyResultInput) (*entity.EngagementSurveyResult, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.engagementSurveyResult.Create(ctx, func() (*entity.EngagementSurveyResult, error) { return createEngagementSurveyResult(in) })
}

// UpdateEngagementSurveyResult menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdateEngagementSurveyResult(ctx context.Context, key string, in UpdateEngagementSurveyResultInput) (*entity.EngagementSurveyResult, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.engagementSurveyResult.Update(ctx, key, func(value *entity.EngagementSurveyResult) error { return patchEngagementSurveyResult(value, in) })
}

// DeleteEngagementSurveyResult menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeleteEngagementSurveyResult(ctx context.Context, key string) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.engagementSurveyResult.Delete(ctx, key)
}

// GetEngagementSurveyResultByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetEngagementSurveyResultByID(ctx context.Context, key string) (*entity.EngagementSurveyResult, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.engagementSurveyResult.Get(ctx, key)
}

// GetAllEngagementSurveyResults menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllEngagementSurveyResults(ctx context.Context, limit, offset int) (operation.Page[entity.EngagementSurveyResult], error) {
	if s == nil {
		return operation.Page[entity.EngagementSurveyResult]{}, hrerrors.ErrUnavailable
	}
	return s.engagementSurveyResult.List(ctx, nil, limit, offset)
}

// FindEngagementSurveyResultBySurveyID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindEngagementSurveyResultBySurveyID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.EngagementSurveyResult], error) {
	if s == nil {
		return operation.Page[entity.EngagementSurveyResult]{}, hrerrors.ErrUnavailable
	}
	return s.engagementSurveyResult.List(ctx, []domainrepo.Filter{{Column: "survey_id", Values: []any{value}}}, limit, offset)
}

// FindEngagementSurveyResultByOrganizationUnitReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindEngagementSurveyResultByOrganizationUnitReferenceID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.EngagementSurveyResult], error) {
	if s == nil {
		return operation.Page[entity.EngagementSurveyResult]{}, hrerrors.ErrUnavailable
	}
	return s.engagementSurveyResult.List(ctx, []domainrepo.Filter{{Column: "organization_unit_reference_id", Values: []any{value}}}, limit, offset)
}

// FindEngagementSurveyResultByQuestionCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindEngagementSurveyResultByQuestionCode(ctx context.Context, value string, limit, offset int) (operation.Page[entity.EngagementSurveyResult], error) {
	if s == nil {
		return operation.Page[entity.EngagementSurveyResult]{}, hrerrors.ErrUnavailable
	}
	return s.engagementSurveyResult.List(ctx, []domainrepo.Filter{{Column: "question_code", Values: []any{value}}}, limit, offset)
}
