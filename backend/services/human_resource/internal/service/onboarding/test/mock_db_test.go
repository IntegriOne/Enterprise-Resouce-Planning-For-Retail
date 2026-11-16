// Kerangka tes awal dibuat dengan alat bantu; tes ini dipelihara langsung.
// Fixture dan dependency pada file ini hanya untuk pengujian service.
package service_test

import (
	"testing"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/onboarding"
	mockdb "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/testsupport"
)

func setup(tb testing.TB) (*mockdb.Harness, *service.Service, *mockdb.Store[entity.OnboardingPlan, string]) {
	tb.Helper()
	h := mockdb.New()
	pOnboardingPlan := mockdb.NewStore[entity.OnboardingPlan, string](h, func(v *entity.OnboardingPlan) string { return v.ID })
	pOnboardingTask := mockdb.NewStore[entity.OnboardingTask, string](h, func(v *entity.OnboardingTask) string { return v.ID })
	svc, err := service.NewService(service.Dependencies{Runtime: h.Runtime(), OnboardingPlan: pOnboardingPlan, OnboardingTask: pOnboardingTask})
	if err != nil {
		tb.Fatal(err)
	}
	return h, svc, pOnboardingPlan
}
func input() service.CreateOnboardingPlanInput {
	return service.CreateOnboardingPlanInput{WorkerID: mockdb.ID(20), PlanCode: "TEST", StartDate: mockdb.Now()}
}
