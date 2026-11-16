package service_test

import (
	"testing"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	service "github.com/integrione/erp-retail/backend/services/human_resource/domain/service/promotion"
	f "github.com/integrione/erp-retail/backend/services/human_resource/domain/service/test/fixture"
)

func TestPromotionDecision(t *testing.T) {
	t.Parallel()
	r, w, d := f.Recommendation(), f.Worker(), f.Decision()
	if err := service.ValidateDecision(&r, &w, &d, nil); err != nil {
		t.Fatal(err)
	}
	d.DecidedByPrincipalID = r.RecommendedByPrincipalID
	if err := service.ValidateDecision(&r, &w, &d, nil); err == nil {
		t.Fatal("self approval allowed")
	}
	d = f.Decision()
	if err := service.ValidateDecision(&r, &w, &d, []entity.PromotionDecision{d}); err == nil {
		t.Fatal("repeated decision allowed")
	}
}

func TestPromotionEvidence(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		mutate func(*service.ApplicationInput)
		valid  bool
	}{
		{"matching applied evidence", func(_ *service.ApplicationInput) {}, true},
		{"future effective date", func(i *service.ApplicationInput) { i.BusinessDate = i.BusinessDate.AddDate(0, 0, -1) }, false},
		{"wrong worker", func(i *service.ApplicationInput) { i.NewAssignment.WorkerID = f.ID(999) }, false},
		{"wrong source closure", func(i *service.ApplicationInput) { i.SourceAssignment.ValidTo = nil }, false},
		{"event not persisted", func(i *service.ApplicationInput) { i.Event.ID = 0 }, false},
		{"foreign employer", func(i *service.ApplicationInput) { i.TargetPosition.LegalEntityReferenceID = f.ID(999) }, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, w, d, a := f.Recommendation(), f.Worker(), f.Decision(), f.Assessment()
			r.Status = "approved"
			r.ApprovalDecisionID = &d.ID
			a.Status = "finalized"
			at := f.Date()
			a.FinalizedAt = &at
			actor := f.ID(99)
			a.FinalizedByPrincipalID = &actor
			source, next, target := f.Assignment(), f.Assignment(), f.Position()
			source.PositionID = f.ID(50)
			end := at.AddDate(0, 0, -1)
			source.ValidTo = &end
			next.ID = f.ID(16)
			next.ValidFrom = at
			w.PrimaryAssignmentID = &next.ID
			ref := r.ID
			event := entity.WorkerLifecycleEvent{ID: 1, WorkerID: w.ID, EventType: "promotion", OccurredAt: at, EffectiveDate: at, SourceReference: &ref, EventPayload: entity.JSON(`{}`), IdempotencyKey: "PROMO-EVENT"}
			input := service.ApplicationInput{Recommendation: &r, Worker: &w, Decision: &d, Assessment: &a, SourceAssignment: &source, NewAssignment: &next, TargetPosition: &target, Event: &event, BusinessDate: at, AppliedAt: at}
			tt.mutate(&input)
			result, err := service.BuildAppliedRecommendation(input)
			if (err == nil) != tt.valid {
				t.Fatalf("application: %v", err)
			}
			if err == nil && (result.Status != "applied" || r.Status != "approved") {
				t.Fatal("recommendation mutation")
			}
		})
	}
}
