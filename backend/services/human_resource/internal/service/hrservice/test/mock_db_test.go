// Kerangka tes awal dibuat dengan alat bantu; tes ini dipelihara langsung.
// Fixture dan dependency pada file ini hanya untuk pengujian service.
package service_test

import (
	"testing"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/hrservice"
	mockdb "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/testsupport"
)

func setup(tb testing.TB) (*mockdb.Harness, *service.Service, *mockdb.Store[entity.ServiceRequest, string]) {
	tb.Helper()
	h := mockdb.New()
	pServiceRequest := mockdb.NewStore[entity.ServiceRequest, string](h, func(v *entity.ServiceRequest) string { return v.ID })
	pServiceRequestEvent := mockdb.NewStore[entity.ServiceRequestEvent, string](h, func(v *entity.ServiceRequestEvent) string { return v.ID })
	svc, err := service.NewService(service.Dependencies{Runtime: h.Runtime(), ServiceRequest: pServiceRequest, ServiceRequestEvent: pServiceRequestEvent})
	if err != nil {
		tb.Fatal(err)
	}
	return h, svc, pServiceRequest
}
func input() service.CreateServiceRequestInput {
	return service.CreateServiceRequestInput{RequestNumber: "TEST", WorkerID: mockdb.ID(20), RequestType: "TEST", RestrictedDetailsDocumentVersionReferenceID: mockdb.ID(20), IdempotencyKey: "TEST"}
}
