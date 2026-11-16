// Kerangka awal dibuat dengan alat bantu; fixture ini dipelihara langsung.
// Fixture dan dependency pada file ini hanya untuk pengujian service.
package service_test

import (
	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/talent"
	mockdb "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/testsupport"
	"testing"
)

func setup(tb testing.TB) (*mockdb.Harness, *service.Service, *mockdb.Store[entity.TalentPool, string]) {
	h, svc, pool, _, _, _, _, _ := setupTalent(tb)
	return h, svc, pool
}

func setupTalent(tb testing.TB) (
	*mockdb.Harness, *service.Service, *mockdb.Store[entity.TalentPool, string],
	*mockdb.Store[entity.TalentFramework, string], *mockdb.Store[entity.TalentDimension, string],
	*mockdb.Store[entity.TalentReviewCycle, string], *mockdb.Store[entity.TalentAssessment, string],
	*mockdb.Store[entity.TalentAssessmentScore, string],
) {
	tb.Helper()
	h := mockdb.New()
	pTalentFramework := mockdb.NewStore[entity.TalentFramework, string](h, func(v *entity.TalentFramework) string { return v.ID })
	pTalentDimension := mockdb.NewStore[entity.TalentDimension, string](h, func(v *entity.TalentDimension) string { return v.ID })
	pTalentReviewCycle := mockdb.NewStore[entity.TalentReviewCycle, string](h, func(v *entity.TalentReviewCycle) string { return v.ID })
	pTalentAssessment := mockdb.NewStore[entity.TalentAssessment, string](h, func(v *entity.TalentAssessment) string { return v.ID })
	pTalentAssessmentScore := mockdb.NewStore[entity.TalentAssessmentScore, string](h, func(v *entity.TalentAssessmentScore) string { return v.ID })
	pTalentCalibrationSession := mockdb.NewStore[entity.TalentCalibrationSession, string](h, func(v *entity.TalentCalibrationSession) string { return v.ID })
	pTalentCalibrationDecision := mockdb.NewStore[entity.TalentCalibrationDecision, string](h, func(v *entity.TalentCalibrationDecision) string { return v.ID })
	pTalentPool := mockdb.NewStore[entity.TalentPool, string](h, func(v *entity.TalentPool) string { return v.ID })
	pTalentPoolMembership := mockdb.NewStore[entity.TalentPoolMembership, string](h, func(v *entity.TalentPoolMembership) string { return v.ID })
	svc, err := service.NewService(service.Dependencies{Runtime: h.Runtime(), TalentFramework: pTalentFramework, TalentDimension: pTalentDimension, TalentReviewCycle: pTalentReviewCycle, TalentAssessment: pTalentAssessment, TalentAssessmentScore: pTalentAssessmentScore, TalentCalibrationSession: pTalentCalibrationSession, TalentCalibrationDecision: pTalentCalibrationDecision, TalentPool: pTalentPool, TalentPoolMembership: pTalentPoolMembership})
	if err != nil {
		tb.Fatal(err)
	}
	return h, svc, pTalentPool, pTalentFramework, pTalentDimension, pTalentReviewCycle, pTalentAssessment, pTalentAssessmentScore
}
func input() service.CreateTalentPoolInput {
	return service.CreateTalentPoolInput{PoolCode: "TEST", PoolVersion: 1, PoolName: "TEST", LegalEntityReferenceID: mockdb.ID(20), EligibilityDocumentVersionReferenceID: mockdb.ID(20), EligibilityDigest: []byte{1}, ValidFrom: mockdb.Now(), OwnerPrincipalID: mockdb.ID(20)}
}
