// Kerangka tes awal dibuat dengan alat bantu; tes ini dipelihara langsung.
// Fixture dan dependency pada file ini hanya untuk pengujian service.
package service_test

import (
	"testing"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/employeerelations"
	mockdb "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/testsupport"
)

func setup(tb testing.TB) (*mockdb.Harness, *service.Service, *mockdb.Store[entity.EmployeeRelationsCase, string]) {
	tb.Helper()
	h := mockdb.New()
	pEmployeeRelationsCase := mockdb.NewStore[entity.EmployeeRelationsCase, string](h, func(v *entity.EmployeeRelationsCase) string { return v.ID })
	pEmployeeRelationsEvent := mockdb.NewStore[entity.EmployeeRelationsEvent, int64](h, func(v *entity.EmployeeRelationsEvent) int64 { return v.ID })
	pCollectiveAgreement := mockdb.NewStore[entity.CollectiveAgreement, string](h, func(v *entity.CollectiveAgreement) string { return v.ID })
	svc, err := service.NewService(service.Dependencies{Runtime: h.Runtime(), EmployeeRelationsCase: pEmployeeRelationsCase, EmployeeRelationsEvent: pEmployeeRelationsEvent, CollectiveAgreement: pCollectiveAgreement})
	if err != nil {
		tb.Fatal(err)
	}
	return h, svc, pEmployeeRelationsCase
}
func input() service.CreateEmployeeRelationsCaseInput {
	return service.CreateEmployeeRelationsCaseInput{CaseNumber: "TEST", CaseType: "TEST", RestrictedSummaryReferenceID: mockdb.ID(20), IdempotencyKey: "TEST"}
}
