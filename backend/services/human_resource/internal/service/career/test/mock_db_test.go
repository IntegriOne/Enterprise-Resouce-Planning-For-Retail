// Kerangka tes awal dibuat dengan alat bantu; tes ini dipelihara langsung.
// Fixture dan dependency pada file ini hanya untuk pengujian service.
package service_test

import (
	"testing"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/career"
	mockdb "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/testsupport"
)

func setup(tb testing.TB) (*mockdb.Harness, *service.Service, *mockdb.Store[entity.DevelopmentPlan, string]) {
	tb.Helper()
	h := mockdb.New()
	pDevelopmentPlan := mockdb.NewStore[entity.DevelopmentPlan, string](h, func(v *entity.DevelopmentPlan) string { return v.ID })
	pDevelopmentAction := mockdb.NewStore[entity.DevelopmentAction, string](h, func(v *entity.DevelopmentAction) string { return v.ID })
	pCareerPath := mockdb.NewStore[entity.CareerPath, string](h, func(v *entity.CareerPath) string { return v.ID })
	pCareerPathTransition := mockdb.NewStore[entity.CareerPathTransition, string](h, func(v *entity.CareerPathTransition) string { return v.ID })
	pWorkerCareerAspiration := mockdb.NewStore[entity.WorkerCareerAspiration, string](h, func(v *entity.WorkerCareerAspiration) string { return v.ID })
	svc, err := service.NewService(service.Dependencies{Runtime: h.Runtime(), DevelopmentPlan: pDevelopmentPlan, DevelopmentAction: pDevelopmentAction, CareerPath: pCareerPath, CareerPathTransition: pCareerPathTransition, WorkerCareerAspiration: pWorkerCareerAspiration})
	if err != nil {
		tb.Fatal(err)
	}
	return h, svc, pDevelopmentPlan
}
func input() service.CreateDevelopmentPlanInput {
	return service.CreateDevelopmentPlanInput{WorkerID: mockdb.ID(20), PlanCode: "TEST", PlanVersion: 1, StartDate: mockdb.Now(), TargetDate: mockdb.Now().AddDate(0, 0, 1)}
}
