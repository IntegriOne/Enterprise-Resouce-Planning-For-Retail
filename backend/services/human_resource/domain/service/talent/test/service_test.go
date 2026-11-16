package service_test

import (
	"testing"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	service "github.com/integrione/erp-retail/backend/services/human_resource/domain/service/talent"
	f "github.com/integrione/erp-retail/backend/services/human_resource/domain/service/test/fixture"
)

func TestTalentFinalization(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		mutate func(*service.FinalizationInput)
		valid  bool
	}{
		{"complete", func(_ *service.FinalizationInput) {}, true},
		{"missing required", func(i *service.FinalizationInput) { i.Scores = nil }, false},
		{"wrong framework", func(i *service.FinalizationInput) { i.Scores[0].FrameworkID = f.ID(999) }, false},
		{"invalid weight", func(i *service.FinalizationInput) { i.Dimensions[0].WeightPercent = "99.999" }, false},
		{"missing anchors", func(i *service.FinalizationInput) { i.Dimensions[0].RatingAnchors = entity.JSON(`{"1":"basic"}`) }, false},
		{"duplicate dimension", func(i *service.FinalizationInput) { i.Dimensions = append(i.Dimensions, i.Dimensions[0]) }, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, c, framework := f.Assessment(), f.Cycle(), f.Framework()
			input := service.FinalizationInput{Assessment: &a, Cycle: &c, Framework: &framework, Dimensions: []entity.TalentDimension{f.Dimension()}, Scores: []entity.TalentAssessmentScore{f.Score()}}
			tt.mutate(&input)
			err := service.ValidateFinalization(input)
			if (err == nil) != tt.valid {
				t.Fatalf("finalization: %v", err)
			}
		})
	}
}

func TestCalibrationPreservesOriginalScore(t *testing.T) {
	t.Parallel()
	at := f.Date()
	actor := f.ID(96)
	assessment := f.Assessment()
	assessment.Status = "finalized"
	assessment.FinalizedAt = &at
	assessment.FinalizedByPrincipalID = &actor
	score := f.Score()
	session := entity.TalentCalibrationSession{ID: f.ID(20), TalentReviewCycleID: assessment.TalentReviewCycleID, SessionCode: "PANEL", ChairPrincipalID: actor, ScheduledAt: at, Status: "open"}
	decision := entity.TalentCalibrationDecision{ID: f.ID(21), SessionID: session.ID, TalentReviewCycleID: session.TalentReviewCycleID, AssessmentID: assessment.ID, DimensionID: score.DimensionID, OriginalScore: score.Score, CalibratedScore: 5, RationaleDocumentVersionReferenceID: f.ID(97), RationaleDigest: []byte{1}, DecidedByPrincipalID: actor, DecidedAt: at}
	if err := service.ValidateCalibration(&session, &assessment, &score, &decision); err != nil {
		t.Fatal(err)
	}
	if score.Score != 4 {
		t.Fatal("original score mutated")
	}
	foreignScore := score
	foreignScore.FrameworkID = f.ID(999)
	if err := service.ValidateCalibration(&session, &assessment, &foreignScore, &decision); err == nil {
		t.Fatal("score from another framework accepted")
	}
	decision.OriginalScore = 3
	if err := service.ValidateCalibration(&session, &assessment, &score, &decision); err == nil {
		t.Fatal("false original score accepted")
	}
}
