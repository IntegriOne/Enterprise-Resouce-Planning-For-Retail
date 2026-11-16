// Kerangka tes awal dibuat dengan alat bantu; tes ini dipelihara langsung.
// Fixture dan dependency pada file ini hanya untuk pengujian service.
package service_test

import (
	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/accessrequest"
	mockdb "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/testsupport"
	"testing"
)

func setup(tb testing.TB) (*mockdb.Harness, *service.Service, *mockdb.Store[entity.AccessRequest, string]) {
	tb.Helper()
	h := mockdb.New()
	pAccessRequest := mockdb.NewStore[entity.AccessRequest, string](h, func(v *entity.AccessRequest) string { return v.ID })
	svc, err := service.NewService(service.Dependencies{Runtime: h.Runtime(), AccessRequest: pAccessRequest})
	if err != nil {
		tb.Fatal(err)
	}
	return h, svc, pAccessRequest
}
func input() service.CreateAccessRequestInput {
	return service.CreateAccessRequestInput{WorkerID: mockdb.ID(20), LifecycleEventID: 1, RequestKind: "revoke", RequestedScopeType: "legal_entity", RequestedScopeReferenceID: mockdb.ID(20), IdempotencyKey: "TEST"}
}
