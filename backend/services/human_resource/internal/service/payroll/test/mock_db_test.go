// Kerangka awal dibuat dengan alat bantu; fixture ini dipelihara langsung.
// Fixture dan dependency pada file ini hanya untuk pengujian service.
package service_test

import (
	"testing"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/payroll"
	mockdb "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/testsupport"
)

func setup(tb testing.TB) (*mockdb.Harness, *service.Service, *mockdb.Store[entity.PayrollProfile, string]) {
	h, svc, profile, _, _, _ := setupPayroll(tb)
	return h, svc, profile
}

func setupPayroll(tb testing.TB) (
	*mockdb.Harness, *service.Service, *mockdb.Store[entity.PayrollProfile, string],
	*mockdb.Store[entity.PayrollRun, string], *mockdb.Store[entity.PayrollResult, string],
	*mockdb.Store[entity.PayrollResultLine, string],
) {
	tb.Helper()
	h := mockdb.New()
	pPayrollProfile := mockdb.NewStore[entity.PayrollProfile, string](h, func(v *entity.PayrollProfile) string { return v.ID })
	pPayComponent := mockdb.NewStore[entity.PayComponent, string](h, func(v *entity.PayComponent) string { return v.ID })
	pPayrollPeriod := mockdb.NewStore[entity.PayrollPeriod, string](h, func(v *entity.PayrollPeriod) string { return v.ID })
	pPayrollRun := mockdb.NewStore[entity.PayrollRun, string](h, func(v *entity.PayrollRun) string { return v.ID })
	pPayrollResult := mockdb.NewStore[entity.PayrollResult, string](h, func(v *entity.PayrollResult) string { return v.ID })
	pPayrollResultLine := mockdb.NewStore[entity.PayrollResultLine, string](h, func(v *entity.PayrollResultLine) string { return v.ID })
	pPayrollAdjustment := mockdb.NewStore[entity.PayrollAdjustment, string](h, func(v *entity.PayrollAdjustment) string { return v.ID })
	pPayrollGroup := mockdb.NewStore[entity.PayrollGroup, string](h, func(v *entity.PayrollGroup) string { return v.ID })
	pPayrollRuleSet := mockdb.NewStore[entity.PayrollRuleSet, string](h, func(v *entity.PayrollRuleSet) string { return v.ID })
	svc, err := service.NewService(service.Dependencies{Runtime: h.Runtime(), PayrollProfile: pPayrollProfile, PayComponent: pPayComponent, PayrollPeriod: pPayrollPeriod, PayrollRun: pPayrollRun, PayrollResult: pPayrollResult, PayrollResultLine: pPayrollResultLine, PayrollAdjustment: pPayrollAdjustment, PayrollGroup: pPayrollGroup, PayrollRuleSet: pPayrollRuleSet})
	if err != nil {
		tb.Fatal(err)
	}
	return h, svc, pPayrollProfile, pPayrollRun, pPayrollResult, pPayrollResultLine
}
func input() service.CreatePayrollProfileInput {
	return service.CreatePayrollProfileInput{WorkerID: mockdb.ID(20), PayrollCountryCode: "ID", PayrollJurisdictionCode: "TEST", PayrollGroupID: mockdb.ID(20), EmployingLegalEntityReferenceID: mockdb.ID(20), CurrencyCode: "IDR", EffectiveFrom: mockdb.Now()}
}
