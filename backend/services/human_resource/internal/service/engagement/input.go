// Kerangka awal dibuat dengan alat bantu; kode ini dipelihara langsung per domain.
// Aturan bisnis lintas record tetap dijalankan melalui policy/integrity dan command domain.
package service

import (
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

// CreateEngagementSurveyInput adalah object input aplikasi, bukan entity ORM.
type CreateEngagementSurveyInput struct {
	SurveyCode                              string
	SurveyVersion                           int32
	LegalEntityReferenceID                  string
	TopicCode                               string
	OpensAt                                 time.Time
	ClosesAt                                time.Time
	QuestionnaireDocumentVersionReferenceID string
	QuestionnaireDigest                     []byte
	MinimumReportingPopulation              int32
}

// UpdateEngagementSurveyInput membedakan field absen, null dan nilai eksplisit.
type UpdateEngagementSurveyInput struct {
	SurveyCode                              operation.Field[string]
	SurveyVersion                           operation.Field[int32]
	LegalEntityReferenceID                  operation.Field[string]
	TopicCode                               operation.Field[string]
	OpensAt                                 operation.Field[time.Time]
	ClosesAt                                operation.Field[time.Time]
	QuestionnaireDocumentVersionReferenceID operation.Field[string]
	QuestionnaireDigest                     operation.Field[[]byte]
	MinimumReportingPopulation              operation.Field[int32]
}

// CreateEngagementSurveyResultInput adalah object input aplikasi, bukan entity ORM.
type CreateEngagementSurveyResultInput struct {
	SurveyID                    string
	OrganizationUnitReferenceID *string
	QuestionCode                string
	ResultVersion               int32
	RespondentCount             int32
	ScoreValue                  entity.Decimal
	ScaleMinimum                entity.Decimal
	ScaleMaximum                entity.Decimal
	SourceDigest                []byte
}

// UpdateEngagementSurveyResultInput membedakan field absen, null dan nilai eksplisit.
type UpdateEngagementSurveyResultInput struct {
	SurveyID                    operation.Field[string]
	OrganizationUnitReferenceID operation.Field[string]
	QuestionCode                operation.Field[string]
	ResultVersion               operation.Field[int32]
	RespondentCount             operation.Field[int32]
	ScoreValue                  operation.Field[entity.Decimal]
	ScaleMinimum                operation.Field[entity.Decimal]
	ScaleMaximum                operation.Field[entity.Decimal]
	SourceDigest                operation.Field[[]byte]
}
