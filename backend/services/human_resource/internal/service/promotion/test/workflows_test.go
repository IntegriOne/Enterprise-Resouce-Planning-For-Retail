package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/promotion"
	mockdb "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/testsupport"
)

type identityLifecycleStore struct {
	*mockdb.Store[entity.WorkerLifecycleEvent, int64]
	seq int64
}

type assignmentStore struct {
	*mockdb.Store[entity.WorkerAssignment, string]
	reportedCount *int64
}

func (s *assignmentStore) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	if s.reportedCount != nil {
		return *s.reportedCount, nil
	}
	return s.Store.CountWhere(ctx, filters)
}

func (s *identityLifecycleStore) Insert(ctx context.Context, value *entity.WorkerLifecycleEvent) error {
	if value != nil && value.ID == 0 {
		s.seq++
		value.ID = s.seq
	}
	return s.Store.Insert(ctx, value)
}

func TestPromotionWorkflowsSubmitDecideAndApply(t *testing.T) {
	h := mockdb.New()
	proposer := mockdb.ID(10)
	workerPrincipal := mockdb.ID(11)
	approver := mockdb.ID(12)
	h.Actor = operation.Actor{PrincipalID: proposer, SessionID: mockdb.ID(99), Kind: "human"}
	now := mockdb.Now()
	legalEntity := mockdb.ID(20)
	hireDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	effectiveDate := time.Date(2026, 11, 16, 0, 0, 0, 0, time.UTC)

	recommendations := mockdb.NewStore[entity.PromotionRecommendation, string](h, func(v *entity.PromotionRecommendation) string { return v.ID })
	decisions := mockdb.NewStore[entity.PromotionDecision, string](h, func(v *entity.PromotionDecision) string { return v.ID })
	workers := mockdb.NewStore[entity.Worker, string](h, func(v *entity.Worker) string { return v.ID })
	assignments := &assignmentStore{Store: mockdb.NewStore[entity.WorkerAssignment, string](h, func(v *entity.WorkerAssignment) string { return v.ID })}
	lifecycleBase := mockdb.NewStore[entity.WorkerLifecycleEvent, int64](h, func(v *entity.WorkerLifecycleEvent) int64 { return v.ID })
	lifecycleEvents := &identityLifecycleStore{Store: lifecycleBase}
	positions := mockdb.NewStore[entity.Position, string](h, func(v *entity.Position) string { return v.ID })
	assessments := mockdb.NewStore[entity.TalentAssessment, string](h, func(v *entity.TalentAssessment) string { return v.ID })

	svc, err := service.NewService(service.Dependencies{
		Runtime:                 h.Runtime(),
		PromotionRecommendation: recommendations,
		PromotionDecision:       decisions,
		Worker:                  workers,
		WorkerAssignment:        assignments,
		WorkerLifecycleEvent:    lifecycleEvents,
		Position:                positions,
		TalentAssessment:        assessments,
	})
	if err != nil {
		t.Fatal(err)
	}

	workerID := mockdb.ID(1)
	oldAsgID := mockdb.ID(2)
	oldPosID := mockdb.ID(3)
	newPosID := mockdb.ID(4)
	assessID := mockdb.ID(5)

	workers.Rows[workerID] = entity.Worker{
		ID:                            workerID,
		WorkerNumber:                  "W-PROM",
		PersonPartyReferenceID:        mockdb.ID(21),
		WorkerCategory:                entity.WorkerCategoryEmployee,
		Status:                        entity.WorkerStatusActive,
		HireDate:                      &hireDate,
		PrimaryLegalEntityReferenceID: legalEntity,
		PrimaryAssignmentID:           &oldAsgID,
		IAMPrincipalReferenceID:       &workerPrincipal,
		CreatedAt:                     hireDate,
		UpdatedAt:                     hireDate,
	}
	reqAt := hireDate.Add(-2 * time.Hour)
	appAt := hireDate.Add(-1 * time.Hour)
	assignments.Rows[oldAsgID] = entity.WorkerAssignment{
		ID:                     oldAsgID,
		WorkerID:               workerID,
		PositionID:             oldPosID,
		AssignmentType:         "primary",
		FTEPercent:             entity.Decimal("100.000"),
		ValidFrom:              hireDate,
		RequestedByPrincipalID: &proposer,
		RequestedAt:            &reqAt,
		ApprovedByPrincipalID:  &approver,
		ApprovedAt:             &appAt,
	}
	positions.Rows[newPosID] = entity.Position{
		ID:                          newPosID,
		PositionCode:                "POS-SR",
		PositionTitle:               "Senior Specialist",
		JobID:                       mockdb.ID(30),
		OrganizationUnitReferenceID: mockdb.ID(31),
		LegalEntityReferenceID:      legalEntity,
		PositionStatus:              "active",
		HeadcountAuthorized:         entity.Decimal("2.00"),
		ValidFrom:                   hireDate,
	}
	assessments.Rows[assessID] = entity.TalentAssessment{
		ID:                                 assessID,
		WorkerID:                           workerID,
		TalentReviewCycleID:                mockdb.ID(40),
		FrameworkID:                        mockdb.ID(41),
		AssessmentVersion:                  1,
		Status:                             "draft",
		AssessedByPrincipalID:              proposer,
		AssessedAt:                         hireDate,
		EvidenceDocumentVersionReferenceID: mockdb.ID(42),
		EvidenceDigest:                     []byte("hash"),
	}

	ctx := context.Background()
	rec, err := svc.CreatePromotionRecommendation(ctx, service.CreatePromotionRecommendationInput{
		RecommendationNumber:                "PROM-001",
		WorkerID:                            workerID,
		SourceAssignmentID:                  oldAsgID,
		TargetPositionID:                    newPosID,
		AssessmentID:                        assessID,
		RecommendedEffectiveDate:            effectiveDate,
		RationaleDocumentVersionReferenceID: mockdb.ID(50),
		RationaleDigest:                     []byte("rationale"),
		IdempotencyKey:                      "idem-prom-1",
	})
	if err != nil {
		t.Fatal(err)
	}

	// Pengajuan sebelum asesmen berstatus finalized ditolak.
	if _, err := svc.SubmitPromotionRecommendation(ctx, rec.ID); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("submit before finalized assessment must fail: %v", err)
	}
	assessment := assessments.Rows[assessID]
	assessment.Status = "finalized"
	assessment.FinalizedByPrincipalID = &approver
	assessment.FinalizedAt = &now
	assessments.Rows[assessID] = assessment

	rec, err = svc.SubmitPromotionRecommendation(ctx, rec.ID)
	if err != nil || rec.Status != "submitted" {
		t.Fatalf("submit recommendation: rec=%+v err=%v", rec, err)
	}

	// Self-approval oleh pengusul maupun pekerja itu sendiri harus ditolak (SoD).
	if _, err := svc.ApprovePromotionRecommendation(ctx, service.DecidePromotionRecommendationInput{
		RecommendationID:                    rec.ID,
		RationaleDocumentVersionReferenceID: mockdb.ID(51),
		RationaleDigest:                     []byte("approve"),
		IdempotencyKey:                      "idem-dec-self",
	}); !errors.Is(err, hrerrors.ErrForbidden) {
		t.Fatalf("proposer self-approve must be forbidden: %v", err)
	}

	// Return -> revisi rekomendasi baru -> Reject menguji cabang keputusan non-approve.
	h.Actor.PrincipalID = approver
	returned, err := svc.ReturnPromotionRecommendation(ctx, service.DecidePromotionRecommendationInput{
		RecommendationID:                    rec.ID,
		RationaleDocumentVersionReferenceID: mockdb.ID(52),
		RationaleDigest:                     []byte("need-more-data"),
		IdempotencyKey:                      "idem-dec-return",
	})
	if err != nil || returned.Status != "returned" || returned.ApprovalDecisionID != nil {
		t.Fatalf("return recommendation: rec=%+v err=%v", returned, err)
	}

	// Buat revisi baru lalu setujui dan terapkan ke assignment/lifecycle/worker.
	h.Actor.PrincipalID = proposer
	supID := rec.ID
	rec2, err := svc.CreatePromotionRecommendation(ctx, service.CreatePromotionRecommendationInput{
		RecommendationNumber:                "PROM-002",
		WorkerID:                            workerID,
		SourceAssignmentID:                  oldAsgID,
		TargetPositionID:                    newPosID,
		AssessmentID:                        assessID,
		SupersedesRecommendationID:          &supID,
		RecommendedEffectiveDate:            effectiveDate,
		RationaleDocumentVersionReferenceID: mockdb.ID(53),
		RationaleDigest:                     []byte("rationale-2"),
		IdempotencyKey:                      "idem-prom-2",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.SubmitPromotionRecommendation(ctx, rec2.ID); err != nil {
		t.Fatalf("submit rec2: %v", err)
	}
	h.Actor.PrincipalID = approver
	approvedRec, err := svc.ApprovePromotionRecommendation(ctx, service.DecidePromotionRecommendationInput{
		RecommendationID:                    rec2.ID,
		RationaleDocumentVersionReferenceID: mockdb.ID(54),
		RationaleDigest:                     []byte("approved-2"),
		IdempotencyKey:                      "idem-dec-approve-2",
	})
	if err != nil || approvedRec.Status != "approved" || approvedRec.ApprovalDecisionID == nil {
		t.Fatalf("approve rec2: rec=%+v err=%v", approvedRec, err)
	}

	reportedCount := int64(1)
	assignments.reportedCount = &reportedCount
	if _, err := svc.ApplyPromotionRecommendation(ctx, rec2.ID); !errors.Is(err, hrerrors.ErrForbidden) {
		t.Fatalf("apply with incomplete position assignments: %v", err)
	}
	reportedCount = 10001
	if _, err := svc.ApplyPromotionRecommendation(ctx, rec2.ID); !errors.Is(err, hrerrors.ErrConflict) {
		t.Fatalf("apply with oversized position assignments: %v", err)
	}
	assignments.reportedCount = nil
	eligibleWorker := workers.Rows[workerID]
	workerWithTermination := eligibleWorker
	terminationDate := effectiveDate.AddDate(0, 0, -1)
	workerWithTermination.TerminationDate = &terminationDate
	workers.Rows[workerID] = workerWithTermination
	if _, err := svc.ApplyPromotionRecommendation(ctx, rec2.ID); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("apply promotion after recorded termination: %v", err)
	}
	if assignments.Rows[oldAsgID].ValidTo != nil || len(lifecycleEvents.Rows) != 0 {
		t.Fatal("rejected promotion changed the prior assignment or lifecycle history")
	}
	workerWithLaterHire := eligibleWorker
	laterHireDate := effectiveDate.AddDate(0, 0, 1)
	workerWithLaterHire.HireDate = &laterHireDate
	workers.Rows[workerID] = workerWithLaterHire
	if _, err := svc.ApplyPromotionRecommendation(ctx, rec2.ID); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("apply promotion effective before hire: %v", err)
	}
	workers.Rows[workerID] = eligibleWorker
	sourceAssignment := assignments.Rows[oldAsgID]
	gapEnd := effectiveDate.AddDate(0, 0, -2)
	sourceAssignment.ValidTo = &gapEnd
	assignments.Rows[oldAsgID] = sourceAssignment
	if _, err := svc.ApplyPromotionRecommendation(ctx, rec2.ID); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("apply across a gap after the source assignment ended: %v", err)
	}
	scheduledEnd := effectiveDate.AddDate(0, 0, -1)
	sourceAssignment.ValidTo = &scheduledEnd
	assignments.Rows[oldAsgID] = sourceAssignment
	applied, err := svc.ApplyPromotionRecommendation(ctx, rec2.ID)
	if err != nil {
		t.Fatalf("apply promotion: %v", err)
	}
	if applied.Status != "applied" || applied.AppliedAssignmentID == nil || applied.AppliedLifecycleEventID == nil || !applied.AppliedAt.Equal(now) {
		t.Fatalf("unexpected applied recommendation: %+v", applied)
	}
	oldAsg := assignments.Rows[oldAsgID]
	expectedPrev := effectiveDate.AddDate(0, 0, -1)
	if oldAsg.ValidTo == nil || entity.CompareDates(*oldAsg.ValidTo, expectedPrev) != 0 {
		t.Fatalf("source assignment must close on day before effective date: %+v", oldAsg)
	}
	updatedWorker := workers.Rows[workerID]
	if updatedWorker.PrimaryAssignmentID == nil || *updatedWorker.PrimaryAssignmentID != *applied.AppliedAssignmentID {
		t.Fatalf("worker primary assignment not updated: %+v", updatedWorker)
	}
	event := lifecycleEvents.Rows[*applied.AppliedLifecycleEventID]
	if event.EventType != "promotion" || event.SourceReference == nil || *event.SourceReference != rec2.ID {
		t.Fatalf("unexpected promotion lifecycle event: %+v", event)
	}
}
