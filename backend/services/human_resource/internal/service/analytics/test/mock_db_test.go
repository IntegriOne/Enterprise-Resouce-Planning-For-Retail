// Kerangka tes awal dibuat dengan alat bantu; tes ini dipelihara langsung.
// Fixture dan dependency pada file ini hanya untuk pengujian service.
package service_test

import (
	"testing"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/analytics"
	mockdb "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/testsupport"
)

func setup(tb testing.TB) (*mockdb.Harness, *service.Service, *mockdb.Store[entity.HumanCapitalMetricDefinition, string]) {
	tb.Helper()
	h := mockdb.New()
	pHumanCapitalMetricDefinition := mockdb.NewStore[entity.HumanCapitalMetricDefinition, string](h, func(v *entity.HumanCapitalMetricDefinition) string { return v.ID })
	pHumanCapitalMetricSnapshot := mockdb.NewStore[entity.HumanCapitalMetricSnapshot, string](h, func(v *entity.HumanCapitalMetricSnapshot) string { return v.ID })
	svc, err := service.NewService(service.Dependencies{Runtime: h.Runtime(), HumanCapitalMetricDefinition: pHumanCapitalMetricDefinition, HumanCapitalMetricSnapshot: pHumanCapitalMetricSnapshot})
	if err != nil {
		tb.Fatal(err)
	}
	return h, svc, pHumanCapitalMetricDefinition
}
func input() service.CreateHumanCapitalMetricDefinitionInput {
	return service.CreateHumanCapitalMetricDefinitionInput{MetricCode: "TEST", MetricVersion: "TEST", MetricName: "TEST", MetricDomain: "TEST", InclusionExclusionRules: entity.JSON(`{}`)}
}
