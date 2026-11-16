// Kerangka tes awal dibuat dengan alat bantu; tes ini dipelihara langsung.
// Fixture dan dependency pada file ini hanya untuk pengujian service.
package service_test

import (
	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/personnel"
	mockdb "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/testsupport"
	"testing"
)

func setup(tb testing.TB) (*mockdb.Harness, *service.Service, *mockdb.Store[entity.Worker, string]) {
	tb.Helper()
	h := mockdb.New()
	pWorker := mockdb.NewStore[entity.Worker, string](h, func(v *entity.Worker) string { return v.ID })
	pEmploymentContract := mockdb.NewStore[entity.EmploymentContract, string](h, func(v *entity.EmploymentContract) string { return v.ID })
	pWorkerAssignment := mockdb.NewStore[entity.WorkerAssignment, string](h, func(v *entity.WorkerAssignment) string { return v.ID })
	pWorkerLifecycleEvent := mockdb.NewStore[entity.WorkerLifecycleEvent, int64](h, func(v *entity.WorkerLifecycleEvent) int64 { return v.ID })
	pWorkerDocument := mockdb.NewStore[entity.WorkerDocument, string](h, func(v *entity.WorkerDocument) string { return v.ID })
	pEmergencyContact := mockdb.NewStore[entity.EmergencyContact, string](h, func(v *entity.EmergencyContact) string { return v.ID })
	svc, err := service.NewService(service.Dependencies{Runtime: h.Runtime(), Worker: pWorker, EmploymentContract: pEmploymentContract, WorkerAssignment: pWorkerAssignment, WorkerLifecycleEvent: pWorkerLifecycleEvent, WorkerDocument: pWorkerDocument, EmergencyContact: pEmergencyContact})
	if err != nil {
		tb.Fatal(err)
	}
	return h, svc, pWorker
}
func input() service.CreateWorkerInput {
	return service.CreateWorkerInput{WorkerNumber: "TEST", PersonPartyReferenceID: mockdb.ID(20), WorkerCategory: "employee", PrimaryLegalEntityReferenceID: mockdb.ID(20)}
}
