// Kerangka tes awal dibuat dengan alat bantu; tes ini dipelihara langsung.
// Fixture dan dependency pada file ini hanya untuk pengujian service.
package service_test

import (
	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	mockdb "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/testsupport"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/workforceplanning"
	"testing"
)

func setup(tb testing.TB) (*mockdb.Harness, *service.Service, *mockdb.Store[entity.JobFamily, string]) {
	tb.Helper()
	h := mockdb.New()
	pJobFamily := mockdb.NewStore[entity.JobFamily, string](h, func(v *entity.JobFamily) string { return v.ID })
	pJob := mockdb.NewStore[entity.Job, string](h, func(v *entity.Job) string { return v.ID })
	pPosition := mockdb.NewStore[entity.Position, string](h, func(v *entity.Position) string { return v.ID })
	pWorkforcePlan := mockdb.NewStore[entity.WorkforcePlan, string](h, func(v *entity.WorkforcePlan) string { return v.ID })
	pWorkforcePlanLine := mockdb.NewStore[entity.WorkforcePlanLine, string](h, func(v *entity.WorkforcePlanLine) string { return v.ID })
	svc, err := service.NewService(service.Dependencies{Runtime: h.Runtime(), JobFamily: pJobFamily, Job: pJob, Position: pPosition, WorkforcePlan: pWorkforcePlan, WorkforcePlanLine: pWorkforcePlanLine})
	if err != nil {
		tb.Fatal(err)
	}
	return h, svc, pJobFamily
}
func input() service.CreateJobFamilyInput {
	return service.CreateJobFamilyInput{FamilyCode: "TEST", FamilyName: "TEST", JobDomain: "TEST"}
}
