// Kerangka tes awal dibuat dengan alat bantu; tes ini dipelihara langsung.
// Fixture dan dependency pada file ini hanya untuk pengujian service.
package service_test

import (
	"testing"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/performance"
	mockdb "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/testsupport"
)

func setup(tb testing.TB) (*mockdb.Harness, *service.Service, *mockdb.Store[entity.PerformanceCycle, string]) {
	tb.Helper()
	h := mockdb.New()
	pPerformanceCycle := mockdb.NewStore[entity.PerformanceCycle, string](h, func(v *entity.PerformanceCycle) string { return v.ID })
	pPerformanceGoal := mockdb.NewStore[entity.PerformanceGoal, string](h, func(v *entity.PerformanceGoal) string { return v.ID })
	pPerformanceReview := mockdb.NewStore[entity.PerformanceReview, string](h, func(v *entity.PerformanceReview) string { return v.ID })
	pPerformanceGoalUpdate := mockdb.NewStore[entity.PerformanceGoalUpdate, string](h, func(v *entity.PerformanceGoalUpdate) string { return v.ID })
	pPerformanceReviewGoalScore := mockdb.NewStore[entity.PerformanceReviewGoalScore, string](h, func(v *entity.PerformanceReviewGoalScore) string { return v.ID })
	svc, err := service.NewService(service.Dependencies{Runtime: h.Runtime(), PerformanceCycle: pPerformanceCycle, PerformanceGoal: pPerformanceGoal, PerformanceReview: pPerformanceReview, PerformanceGoalUpdate: pPerformanceGoalUpdate, PerformanceReviewGoalScore: pPerformanceReviewGoalScore})
	if err != nil {
		tb.Fatal(err)
	}
	return h, svc, pPerformanceCycle
}
func input() service.CreatePerformanceCycleInput {
	return service.CreatePerformanceCycleInput{CycleCode: "TEST", CycleName: "TEST", CycleType: "TEST", PeriodStart: mockdb.Now(), PeriodEnd: mockdb.Now().AddDate(0, 0, 1)}
}
