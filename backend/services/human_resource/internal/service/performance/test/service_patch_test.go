package service_test

import (
	"context"
	"testing"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/performance"
	mockdb "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/testsupport"
)

func TestPatchPreservesSQLNullAndJSONNull(t *testing.T) {
	h := mockdb.New()
	cycles := mockdb.NewStore[entity.PerformanceCycle, string](h, func(v *entity.PerformanceCycle) string { return v.ID })
	goals := mockdb.NewStore[entity.PerformanceGoal, string](h, func(v *entity.PerformanceGoal) string { return v.ID })
	reviews := mockdb.NewStore[entity.PerformanceReview, string](h, func(v *entity.PerformanceReview) string { return v.ID })
	updates := mockdb.NewStore[entity.PerformanceGoalUpdate, string](h, func(v *entity.PerformanceGoalUpdate) string { return v.ID })
	scores := mockdb.NewStore[entity.PerformanceReviewGoalScore, string](h, func(v *entity.PerformanceReviewGoalScore) string { return v.ID })
	svc, err := service.NewService(service.Dependencies{
		Runtime: h.Runtime(), PerformanceCycle: cycles, PerformanceGoal: goals,
		PerformanceReview: reviews, PerformanceGoalUpdate: updates, PerformanceReviewGoalScore: scores,
	})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	goal, err := svc.CreatePerformanceGoal(ctx, service.CreatePerformanceGoalInput{
		WorkerID: mockdb.ID(1), PerformanceCycleID: mockdb.ID(2), GoalTitle: "Sales", GoalType: "kpi",
	})
	if err != nil {
		t.Fatal(err)
	}
	if goal.MeasureDefinition != nil || goals.Rows[goal.ID].MeasureDefinition != nil {
		t.Fatal("SQL NULL was changed into JSON null during create")
	}
	_, err = svc.UpdatePerformanceGoal(ctx, goal.ID, service.UpdatePerformanceGoalInput{})
	if err != nil || goals.Rows[goal.ID].MeasureDefinition != nil {
		t.Fatalf("omitted patch changed SQL NULL: %v", err)
	}
	literal := entity.JSON("null")
	_, err = svc.UpdatePerformanceGoal(ctx, goal.ID, service.UpdatePerformanceGoalInput{
		MeasureDefinition: operation.Field[entity.JSON]{Set: true, Value: &literal},
	})
	if err != nil || string(goals.Rows[goal.ID].MeasureDefinition) != "null" {
		t.Fatalf("literal JSON null was lost: %v", err)
	}
	_, err = svc.UpdatePerformanceGoal(ctx, goal.ID, service.UpdatePerformanceGoalInput{
		MeasureDefinition: operation.Field[entity.JSON]{Set: true},
	})
	if err != nil || goals.Rows[goal.ID].MeasureDefinition != nil {
		t.Fatalf("explicit SQL NULL was lost: %v", err)
	}
}
