package service_test

import (
	"testing"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	service "github.com/integrione/erp-retail/backend/services/human_resource/domain/service/performance"
	f "github.com/integrione/erp-retail/backend/services/human_resource/domain/service/test/fixture"
)

func TestReviewGoalScores(t *testing.T) {
	t.Parallel()
	weight := entity.Decimal("100")
	review := entity.PerformanceReview{ID: f.ID(1), WorkerID: f.ID(2), PerformanceCycleID: f.ID(3), ReviewerPrincipalID: f.ID(4), ReviewerRole: "manager", ReviewVersion: 1, CompetencyAssessment: entity.JSON(`{}`), VisibilityClass: "restricted"}
	goal := entity.PerformanceGoal{ID: f.ID(5), WorkerID: f.ID(2), PerformanceCycleID: f.ID(3), GoalTitle: "Sales", GoalType: "kpi", WeightPercent: &weight, GoalStatus: "approved"}
	score := entity.PerformanceReviewGoalScore{ID: f.ID(6), PerformanceReviewID: f.ID(1), GoalID: f.ID(5), WorkerID: f.ID(2), PerformanceCycleID: f.ID(3), Rating: 4, WeightPercentSnapshot: "100", RubricDocumentVersionReferenceID: f.ID(7), EvidenceDocumentVersionReferenceID: f.ID(8)}
	if err := service.ValidateReviewGoals(&review, []entity.PerformanceGoal{goal}, []entity.PerformanceReviewGoalScore{score}); err != nil {
		t.Fatal(err)
	}
	score.WeightPercentSnapshot = "99"
	if err := service.ValidateReviewGoals(&review, []entity.PerformanceGoal{goal}, []entity.PerformanceReviewGoalScore{score}); err == nil {
		t.Fatal("weight mismatch accepted")
	}
	score.WeightPercentSnapshot = "100"
	score.WorkerID = f.ID(999)
	if err := service.ValidateReviewGoals(&review, []entity.PerformanceGoal{goal}, []entity.PerformanceReviewGoalScore{score}); err == nil {
		t.Fatal("other worker score accepted")
	}
}
