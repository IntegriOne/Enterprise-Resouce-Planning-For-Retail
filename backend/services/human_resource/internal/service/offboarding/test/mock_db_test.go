// Kerangka tes awal dibuat dengan alat bantu; tes ini dipelihara langsung.
// Fixture dan dependency pada file ini hanya untuk pengujian service.
package service_test

import (
	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/offboarding"
	mockdb "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/testsupport"
	"testing"
)

func setup(tb testing.TB) (*mockdb.Harness, *service.Service, *mockdb.Store[entity.OffboardingPlan, string]) {
	tb.Helper()
	h := mockdb.New()
	pOffboardingPlan := mockdb.NewStore[entity.OffboardingPlan, string](h, func(v *entity.OffboardingPlan) string { return v.ID })
	pOffboardingTask := mockdb.NewStore[entity.OffboardingTask, string](h, func(v *entity.OffboardingTask) string { return v.ID })
	svc, err := service.NewService(service.Dependencies{Runtime: h.Runtime(), OffboardingPlan: pOffboardingPlan, OffboardingTask: pOffboardingTask})
	if err != nil {
		tb.Fatal(err)
	}
	return h, svc, pOffboardingPlan
}
func input() service.CreateOffboardingPlanInput {
	return service.CreateOffboardingPlanInput{WorkerID: mockdb.ID(20), PlanNumber: "TEST", ReasonCode: "TEST", NoticeDate: mockdb.Now(), LastWorkingDate: mockdb.Now().AddDate(0, 0, 1), TerminationEffectiveDate: mockdb.Now().AddDate(0, 0, 2)}
}
