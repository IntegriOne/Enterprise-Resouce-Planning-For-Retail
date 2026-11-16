// Kerangka tes awal dibuat dengan alat bantu; tes ini dipelihara langsung.
// Fixture dan dependency pada file ini hanya untuk pengujian service.
package service_test

import (
	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/promotion"
	mockdb "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/testsupport"
	"testing"
)

func setup(tb testing.TB) (*mockdb.Harness, *service.Service, *mockdb.Store[entity.PromotionRecommendation, string]) {
	tb.Helper()
	h := mockdb.New()
	pPromotionRecommendation := mockdb.NewStore[entity.PromotionRecommendation, string](h, func(v *entity.PromotionRecommendation) string { return v.ID })
	pPromotionDecision := mockdb.NewStore[entity.PromotionDecision, string](h, func(v *entity.PromotionDecision) string { return v.ID })
	svc, err := service.NewService(service.Dependencies{Runtime: h.Runtime(), PromotionRecommendation: pPromotionRecommendation, PromotionDecision: pPromotionDecision})
	if err != nil {
		tb.Fatal(err)
	}
	return h, svc, pPromotionRecommendation
}
func input() service.CreatePromotionRecommendationInput {
	return service.CreatePromotionRecommendationInput{RecommendationNumber: "TEST", WorkerID: mockdb.ID(20), SourceAssignmentID: mockdb.ID(20), TargetPositionID: mockdb.ID(20), AssessmentID: mockdb.ID(20), RecommendedEffectiveDate: mockdb.Now(), RationaleDocumentVersionReferenceID: mockdb.ID(20), RationaleDigest: []byte{1}, IdempotencyKey: "TEST"}
}
