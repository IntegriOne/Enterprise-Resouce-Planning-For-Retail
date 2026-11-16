// Kerangka tes awal dibuat dengan alat bantu; tes ini dipelihara langsung.
// Fixture dan dependency pada file ini hanya untuk pengujian service.
package service_test

import (
	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/policy"
	mockdb "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/testsupport"
	"testing"
)

func setup(tb testing.TB) (*mockdb.Harness, *service.Service, *mockdb.Store[entity.PolicyAcknowledgement, string]) {
	tb.Helper()
	h := mockdb.New()
	pPolicy := mockdb.NewStore[entity.Policy, string](h, func(v *entity.Policy) string { return v.ID })
	pPolicyAcknowledgement := mockdb.NewStore[entity.PolicyAcknowledgement, string](h, func(v *entity.PolicyAcknowledgement) string { return v.ID })
	svc, err := service.NewService(service.Dependencies{Runtime: h.Runtime(), Policy: pPolicy, PolicyAcknowledgement: pPolicyAcknowledgement})
	if err != nil {
		tb.Fatal(err)
	}
	return h, svc, pPolicyAcknowledgement
}
func input() service.CreatePolicyAcknowledgementInput {
	return service.CreatePolicyAcknowledgementInput{PolicyID: mockdb.ID(20), WorkerID: mockdb.ID(20)}
}
