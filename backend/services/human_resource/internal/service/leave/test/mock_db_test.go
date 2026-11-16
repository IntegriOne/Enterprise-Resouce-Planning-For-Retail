// Kerangka tes awal dibuat dengan alat bantu; tes ini dipelihara langsung.
// Fixture dan dependency pada file ini hanya untuk pengujian service.
package service_test

import (
	"testing"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/leave"
	mockdb "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/testsupport"
)

func setup(tb testing.TB) (*mockdb.Harness, *service.Service, *mockdb.Store[entity.LeavePolicy, string]) {
	tb.Helper()
	h := mockdb.New()
	pLeavePolicy := mockdb.NewStore[entity.LeavePolicy, string](h, func(v *entity.LeavePolicy) string { return v.ID })
	pLeaveEntitlement := mockdb.NewStore[entity.LeaveEntitlement, string](h, func(v *entity.LeaveEntitlement) string { return v.ID })
	pLeaveAdjustment := mockdb.NewStore[entity.LeaveAdjustment, string](h, func(v *entity.LeaveAdjustment) string { return v.ID })
	pLeaveLedgerEntry := mockdb.NewStore[entity.LeaveLedgerEntry, string](h, func(v *entity.LeaveLedgerEntry) string { return v.ID })
	svc, err := service.NewService(service.Dependencies{Runtime: h.Runtime(), LeavePolicy: pLeavePolicy, LeaveEntitlement: pLeaveEntitlement, LeaveAdjustment: pLeaveAdjustment, LeaveLedgerEntry: pLeaveLedgerEntry})
	if err != nil {
		tb.Fatal(err)
	}
	return h, svc, pLeavePolicy
}
func input() service.CreateLeavePolicyInput {
	return service.CreateLeavePolicyInput{PolicyCode: "TEST", PolicyVersion: 1, PolicyName: "TEST", LeaveTypeCode: "TEST", JurisdictionCode: "TEST", AccrualMethod: "TEST", EffectiveFrom: mockdb.Now()}
}
