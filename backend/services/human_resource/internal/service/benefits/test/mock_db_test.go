// Kerangka tes awal dibuat dengan alat bantu; tes ini dipelihara langsung.
// Fixture dan dependency pada file ini hanya untuk pengujian service.
package service_test

import (
	"testing"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/benefits"
	mockdb "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/testsupport"
)

func setup(tb testing.TB) (*mockdb.Harness, *service.Service, *mockdb.Store[entity.BenefitPlan, string]) {
	tb.Helper()
	h := mockdb.New()
	pBenefitPlan := mockdb.NewStore[entity.BenefitPlan, string](h, func(v *entity.BenefitPlan) string { return v.ID })
	pBenefitEnrollment := mockdb.NewStore[entity.BenefitEnrollment, string](h, func(v *entity.BenefitEnrollment) string { return v.ID })
	svc, err := service.NewService(service.Dependencies{Runtime: h.Runtime(), BenefitPlan: pBenefitPlan, BenefitEnrollment: pBenefitEnrollment})
	if err != nil {
		tb.Fatal(err)
	}
	return h, svc, pBenefitPlan
}
func input() service.CreateBenefitPlanInput {
	return service.CreateBenefitPlanInput{PlanCode: "TEST", PlanVersion: 1, PlanName: "TEST", BenefitType: "TEST", JurisdictionCode: "TEST", ValidFrom: mockdb.Now()}
}
