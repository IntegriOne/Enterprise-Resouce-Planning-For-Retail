// Kerangka tes awal dibuat dengan alat bantu; tes ini dipelihara langsung.
// Fixture dan dependency pada file ini hanya untuk pengujian service.
package service_test

import (
	"testing"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/competency"
	mockdb "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/testsupport"
)

func setup(tb testing.TB) (*mockdb.Harness, *service.Service, *mockdb.Store[entity.WorkerQualification, string]) {
	tb.Helper()
	h := mockdb.New()
	pWorkerQualification := mockdb.NewStore[entity.WorkerQualification, string](h, func(v *entity.WorkerQualification) string { return v.ID })
	pCompetency := mockdb.NewStore[entity.Competency, string](h, func(v *entity.Competency) string { return v.ID })
	pJobCompetency := mockdb.NewStore[entity.JobCompetency, string](h, func(v *entity.JobCompetency) string { return v.ID })
	pWorkerCompetencyAssessment := mockdb.NewStore[entity.WorkerCompetencyAssessment, string](h, func(v *entity.WorkerCompetencyAssessment) string { return v.ID })
	svc, err := service.NewService(service.Dependencies{Runtime: h.Runtime(), WorkerQualification: pWorkerQualification, Competency: pCompetency, JobCompetency: pJobCompetency, WorkerCompetencyAssessment: pWorkerCompetencyAssessment})
	if err != nil {
		tb.Fatal(err)
	}
	return h, svc, pWorkerQualification
}
func input() service.CreateWorkerQualificationInput {
	return service.CreateWorkerQualificationInput{WorkerID: mockdb.ID(20), QualificationType: "TEST", Title: "TEST"}
}
