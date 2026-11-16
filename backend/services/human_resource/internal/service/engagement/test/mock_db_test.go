// Kerangka tes awal dibuat dengan alat bantu; tes ini dipelihara langsung.
// Fixture dan dependency pada file ini hanya untuk pengujian service.
package service_test

import (
	"testing"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/engagement"
	mockdb "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/testsupport"
)

func setup(tb testing.TB) (*mockdb.Harness, *service.Service, *mockdb.Store[entity.EngagementSurvey, string]) {
	tb.Helper()
	h := mockdb.New()
	pEngagementSurvey := mockdb.NewStore[entity.EngagementSurvey, string](h, func(v *entity.EngagementSurvey) string { return v.ID })
	pEngagementSurveyResult := mockdb.NewStore[entity.EngagementSurveyResult, string](h, func(v *entity.EngagementSurveyResult) string { return v.ID })
	svc, err := service.NewService(service.Dependencies{Runtime: h.Runtime(), EngagementSurvey: pEngagementSurvey, EngagementSurveyResult: pEngagementSurveyResult})
	if err != nil {
		tb.Fatal(err)
	}
	return h, svc, pEngagementSurvey
}
func input() service.CreateEngagementSurveyInput {
	return service.CreateEngagementSurveyInput{SurveyCode: "TEST", SurveyVersion: 1, LegalEntityReferenceID: mockdb.ID(20), TopicCode: "TEST", OpensAt: mockdb.Now(), ClosesAt: mockdb.Now().AddDate(0, 0, 1), QuestionnaireDocumentVersionReferenceID: mockdb.ID(20), QuestionnaireDigest: []byte{1}, MinimumReportingPopulation: 5}
}
