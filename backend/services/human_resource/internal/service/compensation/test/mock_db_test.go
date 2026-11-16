// Kerangka tes awal dibuat dengan alat bantu; tes ini dipelihara langsung.
// Fixture dan dependency pada file ini hanya untuk pengujian service.
package service_test

import (
	"testing"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/compensation"
	mockdb "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/testsupport"
)

func setup(tb testing.TB) (*mockdb.Harness, *service.Service, *mockdb.Store[entity.CompensationBand, string]) {
	tb.Helper()
	h := mockdb.New()
	pCompensationBand := mockdb.NewStore[entity.CompensationBand, string](h, func(v *entity.CompensationBand) string { return v.ID })
	pCompensationRecord := mockdb.NewStore[entity.CompensationRecord, string](h, func(v *entity.CompensationRecord) string { return v.ID })
	svc, err := service.NewService(service.Dependencies{Runtime: h.Runtime(), CompensationBand: pCompensationBand, CompensationRecord: pCompensationRecord})
	if err != nil {
		tb.Fatal(err)
	}
	return h, svc, pCompensationBand
}
func input() service.CreateCompensationBandInput {
	return service.CreateCompensationBandInput{BandCode: "TEST", BandVersion: 1, JobLevelCode: "TEST", JurisdictionCode: "TEST", CurrencyCode: "IDR", PayFrequency: "TEST", MinimumAmount: entity.Decimal("1"), MidpointAmount: entity.Decimal("1"), MaximumAmount: entity.Decimal("1"), ValidFrom: mockdb.Now()}
}
