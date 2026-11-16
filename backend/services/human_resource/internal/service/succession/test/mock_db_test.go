// Kerangka tes awal dibuat dengan alat bantu; tes ini dipelihara langsung.
// Fixture dan dependency pada file ini hanya untuk pengujian service.
package service_test

import (
	"testing"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/succession"
	mockdb "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/testsupport"
)

func setup(tb testing.TB) (*mockdb.Harness, *service.Service, *mockdb.Store[entity.SuccessionPlan, string]) {
	tb.Helper()
	h := mockdb.New()
	pSuccessionPlan := mockdb.NewStore[entity.SuccessionPlan, string](h, func(v *entity.SuccessionPlan) string { return v.ID })
	pSuccessionCandidate := mockdb.NewStore[entity.SuccessionCandidate, string](h, func(v *entity.SuccessionCandidate) string { return v.ID })
	svc, err := service.NewService(service.Dependencies{Runtime: h.Runtime(), SuccessionPlan: pSuccessionPlan, SuccessionCandidate: pSuccessionCandidate})
	if err != nil {
		tb.Fatal(err)
	}
	return h, svc, pSuccessionPlan
}
func input() service.CreateSuccessionPlanInput {
	return service.CreateSuccessionPlanInput{PositionID: mockdb.ID(20), PlanCode: "TEST", PlanVersion: 1, ReviewDate: mockdb.Now()}
}
