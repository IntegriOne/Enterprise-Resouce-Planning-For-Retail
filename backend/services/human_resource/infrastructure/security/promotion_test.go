package security

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	"github.com/integrione/erp-retail/backend/shared/authcontext"
)

type promoAssignmentStub struct {
	byID map[string]entity.WorkerAssignment
}

func (s *promoAssignmentStub) FindByIDForUpdate(_ context.Context, id string) (*entity.WorkerAssignment, error) {
	v, ok := s.byID[id]
	if !ok {
		return nil, hrerrors.ErrNotFound
	}
	copyVal := v
	return &copyVal, nil
}

type promoPositionStub struct {
	byID map[string]entity.Position
}

func (s *promoPositionStub) FindByIDForUpdate(_ context.Context, id string) (*entity.Position, error) {
	v, ok := s.byID[id]
	if !ok {
		return nil, hrerrors.ErrNotFound
	}
	copyVal := v
	return &copyVal, nil
}

type promoAssessmentStub struct {
	byID map[string]entity.TalentAssessment
}

func (s *promoAssessmentStub) FindByIDForUpdate(_ context.Context, id string) (*entity.TalentAssessment, error) {
	v, ok := s.byID[id]
	if !ok {
		return nil, hrerrors.ErrNotFound
	}
	copyVal := v
	return &copyVal, nil
}

type promoRecommendationStub struct {
	byID map[string]entity.PromotionRecommendation
	rows []entity.PromotionRecommendation
}

func (s *promoRecommendationStub) FindByIDForUpdate(_ context.Context, id string) (*entity.PromotionRecommendation, error) {
	v, ok := s.byID[id]
	if !ok {
		return nil, hrerrors.ErrNotFound
	}
	copyVal := v
	return &copyVal, nil
}

func (s *promoRecommendationStub) SelectWhere(_ context.Context, filters []domainrepo.Filter, _, _ int) ([]entity.PromotionRecommendation, error) {
	var out []entity.PromotionRecommendation
	for _, row := range s.rows {
		match := true
		for _, f := range filters {
			switch f.Column {
			case "recommendation_number":
				if row.RecommendationNumber != f.Values[0] {
					match = false
				}
			case "idempotency_key":
				if row.IdempotencyKey != f.Values[0] {
					match = false
				}
			}
		}
		if match {
			out = append(out, row)
		}
	}
	return out, nil
}

func (s *promoRecommendationStub) CountWhere(_ context.Context, filters []domainrepo.Filter) (int64, error) {
	var count int64
	for _, row := range s.rows {
		for _, f := range filters {
			if f.Column == "supersedes_recommendation_id" && row.SupersedesRecommendationID != nil && *row.SupersedesRecommendationID == f.Values[0] {
				count++
			}
		}
	}
	return count, nil
}

type promoDecisionStub struct {
	rows []entity.PromotionDecision
}

func (s *promoDecisionStub) SelectWhere(_ context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.PromotionDecision, error) {
	var out []entity.PromotionDecision
	for _, row := range s.rows {
		match := true
		for _, f := range filters {
			switch f.Column {
			case "recommendation_id":
				if row.RecommendationID != f.Values[0] {
					match = false
				}
			case "idempotency_key":
				if row.IdempotencyKey != f.Values[0] {
					match = false
				}
			}
		}
		if match {
			out = append(out, row)
		}
	}
	offset = min(offset, len(out))
	return out[offset:min(offset+limit, len(out))], nil
}

func (s *promoDecisionStub) CountWhere(_ context.Context, filters []domainrepo.Filter) (int64, error) {
	rows, err := s.SelectWhere(context.Background(), filters, len(s.rows), 0)
	return int64(len(rows)), err
}

type promoEventStub struct {
	byID map[int64]entity.WorkerLifecycleEvent
}

func (s *promoEventStub) FindByIDForUpdate(_ context.Context, id int64) (*entity.WorkerLifecycleEvent, error) {
	v, ok := s.byID[id]
	if !ok {
		return nil, hrerrors.ErrNotFound
	}
	copyVal := v
	return &copyVal, nil
}

func TestPromotionRecommendationAndDecisionValidators(t *testing.T) {
	t.Parallel()
	proposer := "00000000-0000-4000-8000-000000000001"
	workerPrincipal := "00000000-0000-4000-8000-000000000002"
	approver := "00000000-0000-4000-8000-000000000003"
	legalEntity := "le-1"
	effectiveDate := time.Date(2027, 2, 1, 0, 0, 0, 0, time.UTC)
	now := time.Date(2027, 2, 1, 9, 0, 0, 0, time.UTC)
	sourceID := "asg-src"

	workers := &lockedWorkerStub{
		worker: &entity.Worker{
			ID:                            "worker-1",
			Status:                        entity.WorkerStatusActive,
			PrimaryLegalEntityReferenceID: legalEntity,
			IAMPrincipalReferenceID:       &workerPrincipal,
			PrimaryAssignmentID:           &sourceID,
		},
		order: &[]string{},
	}
	assignments := &promoAssignmentStub{byID: map[string]entity.WorkerAssignment{
		"asg-src": {ID: "asg-src", WorkerID: "worker-1", PositionID: "pos-old", AssignmentType: "primary", ValidFrom: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), ApprovedAt: &now, ApprovedByPrincipalID: &approver},
	}}
	positions := &promoPositionStub{byID: map[string]entity.Position{
		"pos-new": {ID: "pos-new", PositionStatus: "active", LegalEntityReferenceID: legalEntity, ValidFrom: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)},
	}}
	assessments := &promoAssessmentStub{byID: map[string]entity.TalentAssessment{
		"assess-1": {ID: "assess-1", WorkerID: "worker-1", Status: "draft"},
	}}
	recommendations := &promoRecommendationStub{byID: map[string]entity.PromotionRecommendation{}}
	decisions := &promoDecisionStub{}
	events := &promoEventStub{byID: map[int64]entity.WorkerLifecycleEvent{}}

	recVal := &promotionRecommendationValidator{
		workers: workers, assignments: assignments, positions: positions,
		assessments: assessments, recommendations: recommendations, decisions: decisions, events: events, positionOwner: &positionOwnerStub{},
	}
	decVal := &promotionDecisionValidator{
		recommendations: recommendations, workers: workers, decisions: decisions,
	}

	proposerCtx := authcontext.WithCaller(t.Context(), authcontext.CallerContext{PrincipalID: proposer, Authenticated: true})
	approverCtx := authcontext.WithCaller(t.Context(), authcontext.CallerContext{PrincipalID: approver, Authenticated: true})

	draft := &entity.PromotionRecommendation{
		ID:                                  "rec-1",
		RecommendationNumber:                "PROM-001",
		WorkerID:                            "worker-1",
		SourceAssignmentID:                  "asg-src",
		TargetPositionID:                    "pos-new",
		AssessmentID:                        "assess-1",
		RecommendedEffectiveDate:            effectiveDate,
		RationaleDocumentVersionReferenceID: "doc-1",
		RationaleDigest:                     []byte("hash"),
		RecommendedByPrincipalID:            proposer,
		RecommendedAt:                       now.Add(-time.Hour),
		Status:                              "draft",
		IdempotencyKey:                      "rec-idem-1",
	}
	if err := recVal.ValidateMutation(proposerCtx, "hr_promotion_recommendations", nil, draft); err != nil {
		t.Fatalf("create valid draft recommendation: %v", err)
	}

	// Pengajuan sebelum asesmen selesai ditolak.
	submitted := *draft
	submitted.Status = "submitted"
	if err := recVal.ValidateMutation(proposerCtx, "hr_promotion_recommendations", draft, &submitted); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("submit before finalized assessment must fail: %v", err)
	}
	assessments.byID["assess-1"] = entity.TalentAssessment{ID: "assess-1", WorkerID: "worker-1", Status: "finalized"}
	if err := recVal.ValidateMutation(proposerCtx, "hr_promotion_recommendations", draft, &submitted); err != nil {
		t.Fatalf("submit after finalized assessment failed: %v", err)
	}
	recommendations.byID["rec-1"] = submitted

	// Self-decision (pengusul atau pekerja itu sendiri) ditolak.
	selfDecision := &entity.PromotionDecision{
		ID:                                  "dec-1",
		RecommendationID:                    "rec-1",
		StepNumber:                          1,
		Decision:                            entity.ApprovalDecisionApprove,
		DecidedByPrincipalID:                proposer,
		DecidedAt:                           now,
		RationaleDocumentVersionReferenceID: "doc-2",
		RationaleDigest:                     []byte("dec-hash"),
		IdempotencyKey:                      "dec-idem-1",
	}
	if err := decVal.ValidateMutation(proposerCtx, "hr_promotion_decisions", nil, selfDecision); !errors.Is(err, hrerrors.ErrForbidden) {
		t.Fatalf("proposer self-decision must be forbidden: %v", err)
	}
	workerSelfDecision := *selfDecision
	workerSelfDecision.DecidedByPrincipalID = workerPrincipal
	workerSelfCtx := authcontext.WithCaller(t.Context(), authcontext.CallerContext{PrincipalID: workerPrincipal, Authenticated: true})
	if err := decVal.ValidateMutation(workerSelfCtx, "hr_promotion_decisions", nil, &workerSelfDecision); !errors.Is(err, hrerrors.ErrForbidden) {
		t.Fatalf("worker self-decision must be forbidden: %v", err)
	}

	// Approver independen berhasil mencatat keputusan dan menyetujui rekomendasi.
	validDecision := *selfDecision
	validDecision.DecidedByPrincipalID = approver
	if err := decVal.ValidateMutation(approverCtx, "hr_promotion_decisions", nil, &validDecision); err != nil {
		t.Fatalf("independent decision failed: %v", err)
	}
	if err := decVal.ValidateMutation(approverCtx, "hr_promotion_decisions", &validDecision, &validDecision); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("decision update must fail: %v", err)
	}
	decisions.rows = make([]entity.PromotionDecision, domainrepo.MaxLimit)
	for i := range decisions.rows {
		decisions.rows[i] = entity.PromotionDecision{
			RecommendationID: submitted.ID,
			StepNumber:       int32(i + 1),
			Decision:         entity.ApprovalDecisionReturn,
		}
	}
	validDecision.StepNumber = int32(domainrepo.MaxLimit + 1)
	decisions.rows = append(decisions.rows, validDecision)

	approved := submitted
	approved.Status = "approved"
	approved.ApprovalDecisionID = &validDecision.ID
	if err := recVal.ValidateMutation(approverCtx, "hr_promotion_recommendations", &submitted, &approved); err != nil {
		t.Fatalf("approve recommendation failed: %v", err)
	}
	if err := recVal.ValidateMutation(approverCtx, "hr_promotion_recommendations", &submitted, nil); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("delete submitted recommendation must fail: %v", err)
	}

	// Penerapan promosi memvalidasi assignment lama tertutup, assignment baru primer, worker menunjuk assignment baru, dan event promosi.
	prevDay := effectiveDate.AddDate(0, 0, -1)
	assignments.byID["asg-src"] = entity.WorkerAssignment{ID: "asg-src", WorkerID: "worker-1", PositionID: "pos-old", AssignmentType: "primary", ValidFrom: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), ValidTo: &prevDay}
	newAsgID := "asg-new"
	assignments.byID[newAsgID] = entity.WorkerAssignment{ID: newAsgID, WorkerID: "worker-1", PositionID: "pos-new", AssignmentType: "primary", ValidFrom: effectiveDate}
	workers.worker.PrimaryAssignmentID = &newAsgID
	srcRef := "rec-1"
	var eventID int64 = 77
	events.byID[eventID] = entity.WorkerLifecycleEvent{ID: eventID, WorkerID: "worker-1", EventType: "promotion", EffectiveDate: effectiveDate, SourceReference: &srcRef}

	applied := approved
	applied.Status = "applied"
	applied.AppliedAssignmentID = &newAsgID
	applied.AppliedLifecycleEventID = &eventID
	applied.AppliedAt = &now
	if err := recVal.ValidateMutation(approverCtx, "hr_promotion_recommendations", &approved, &applied); err != nil {
		t.Fatalf("apply recommendation with valid evidence failed: %v", err)
	}
}
