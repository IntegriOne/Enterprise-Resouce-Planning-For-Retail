package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/performance"
	mockdb "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/testsupport"
)

func TestPerformanceWorkflowsLifecycleAndGuards(t *testing.T) {
	h := mockdb.New()
	workerPrincipal := mockdb.ID(10)
	managerPrincipal := mockdb.ID(11)
	h.Actor = operation.Actor{PrincipalID: managerPrincipal, SessionID: mockdb.ID(99), Kind: "human"}
	now := mockdb.Now()
	hireDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	workers := mockdb.NewStore[entity.Worker, string](h, func(v *entity.Worker) string { return v.ID })
	cycles := mockdb.NewStore[entity.PerformanceCycle, string](h, func(v *entity.PerformanceCycle) string { return v.ID })
	goals := mockdb.NewStore[entity.PerformanceGoal, string](h, func(v *entity.PerformanceGoal) string { return v.ID })
	reviews := mockdb.NewStore[entity.PerformanceReview, string](h, func(v *entity.PerformanceReview) string { return v.ID })
	updates := mockdb.NewStore[entity.PerformanceGoalUpdate, string](h, func(v *entity.PerformanceGoalUpdate) string { return v.ID })
	scores := mockdb.NewStore[entity.PerformanceReviewGoalScore, string](h, func(v *entity.PerformanceReviewGoalScore) string { return v.ID })

	svc, err := service.NewService(service.Dependencies{
		Runtime:                    h.Runtime(),
		Worker:                     workers,
		PerformanceCycle:           cycles,
		PerformanceGoal:            goals,
		PerformanceReview:          reviews,
		PerformanceGoalUpdate:      updates,
		PerformanceReviewGoalScore: scores,
	})
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	workerID := mockdb.ID(1)
	workers.Rows[workerID] = entity.Worker{
		ID:                            workerID,
		WorkerNumber:                  "W-001",
		WorkerCategory:                entity.WorkerCategoryEmployee,
		Status:                        entity.WorkerStatusActive,
		HireDate:                      &hireDate,
		PrimaryLegalEntityReferenceID: mockdb.ID(20),
		PersonPartyReferenceID:        mockdb.ID(21),
		IAMPrincipalReferenceID:       &workerPrincipal,
		CreatedAt:                     now,
		UpdatedAt:                     now,
	}

	cycle, err := svc.CreatePerformanceCycle(ctx, service.CreatePerformanceCycleInput{
		CycleCode:   "FY27-PERF",
		CycleName:   "FY27 Performance",
		CycleType:   "annual",
		PeriodStart: time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC),
		PeriodEnd:   time.Date(2027, 12, 31, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatal(err)
	}
	cycle, err = svc.LaunchPerformanceCycle(ctx, cycle.ID)
	if err != nil || cycle.Status != entity.PerformanceCycleStatusOpen {
		t.Fatalf("launch cycle: cycle=%+v err=%v", cycle, err)
	}

	w100 := entity.Decimal("100.000")
	goalNoWeight, err := svc.CreatePerformanceGoal(ctx, service.CreatePerformanceGoalInput{
		WorkerID:           workerID,
		PerformanceCycleID: cycle.ID,
		GoalTitle:          "Unweighted Goal",
		GoalType:           "kpi",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.SubmitPerformanceGoal(ctx, goalNoWeight.ID); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("submit goal without weight must fail: %v", err)
	}

	goalRejected, err := svc.CreatePerformanceGoal(ctx, service.CreatePerformanceGoalInput{
		WorkerID:           workerID,
		PerformanceCycleID: cycle.ID,
		GoalTitle:          "Secondary Goal",
		GoalType:           "kpi",
		WeightPercent:      &w100,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.SubmitPerformanceGoal(ctx, goalRejected.ID); err != nil {
		t.Fatalf("submit secondary goal: %v", err)
	}
	rejectedGoal, err := svc.RejectPerformanceGoal(ctx, goalRejected.ID)
	if err != nil || rejectedGoal.GoalStatus != "rejected" {
		t.Fatalf("reject goal: goal=%+v err=%v", rejectedGoal, err)
	}

	// Tolak penolakan/persetujuan mandiri oleh pekerja bersangkutan (SoD).
	h.Actor.PrincipalID = workerPrincipal
	goalApproved, err := svc.CreatePerformanceGoal(ctx, service.CreatePerformanceGoalInput{
		WorkerID:           workerID,
		PerformanceCycleID: cycle.ID,
		GoalTitle:          "Revenue Growth",
		GoalType:           "kpi",
		WeightPercent:      &w100,
	})
	if err != nil {
		t.Fatal(err)
	}
	goalApproved, err = svc.SubmitPerformanceGoal(ctx, goalApproved.ID)
	if err != nil || goalApproved.GoalStatus != "submitted" {
		t.Fatalf("submit goal: goal=%+v err=%v", goalApproved, err)
	}
	if _, err := svc.ApprovePerformanceGoal(ctx, goalApproved.ID); !errors.Is(err, hrerrors.ErrForbidden) {
		t.Fatalf("self-approve goal must be forbidden: %v", err)
	}

	// Review mandiri oleh pekerja -> SubmitPerformanceReview.
	review, err := svc.CreatePerformanceReview(ctx, service.CreatePerformanceReviewInput{
		WorkerID:            workerID,
		PerformanceCycleID:  cycle.ID,
		ReviewerPrincipalID: workerPrincipal,
		ReviewerRole:        "self",
	})
	if err != nil {
		t.Fatal(err)
	}
	review, err = svc.SubmitPerformanceReview(ctx, review.ID)
	if err != nil || review.SubmittedAt == nil {
		t.Fatalf("submit review: review=%+v err=%v", review, err)
	}
	if _, err := svc.FinalizePerformanceReview(ctx, review.ID); !errors.Is(err, hrerrors.ErrForbidden) {
		t.Fatalf("self-finalize review must be forbidden: %v", err)
	}

	// Manager menyetujui goal dan memfinalisasi review setelah skor lengkap.
	h.Actor.PrincipalID = managerPrincipal
	goalApproved, err = svc.ApprovePerformanceGoal(ctx, goalApproved.ID)
	if err != nil || goalApproved.GoalStatus != "approved" || goalApproved.ApprovedByPrincipalID == nil {
		t.Fatalf("approve goal: goal=%+v err=%v", goalApproved, err)
	}

	// Tanpa skor, finalisasi review ditolak.
	if _, err := svc.FinalizePerformanceReview(ctx, review.ID); err == nil {
		t.Fatal("finalize review without scores must fail")
	}

	if _, err := svc.CreatePerformanceReviewGoalScore(ctx, service.CreatePerformanceReviewGoalScoreInput{
		PerformanceReviewID:                review.ID,
		GoalID:                             goalApproved.ID,
		WorkerID:                           workerID,
		PerformanceCycleID:                 cycle.ID,
		Rating:                             5,
		WeightPercentSnapshot:              entity.Decimal("100.000"),
		RubricDocumentVersionReferenceID:   mockdb.ID(30),
		EvidenceDocumentVersionReferenceID: mockdb.ID(31),
	}); err != nil {
		t.Fatalf("create goal score: %v", err)
	}

	finalizedReview, err := svc.FinalizePerformanceReview(ctx, review.ID)
	if err != nil || finalizedReview.FinalizedAt == nil {
		t.Fatalf("finalize review: review=%+v err=%v", finalizedReview, err)
	}

	// Siklus belum bisa ditutup selama masih ada goal berstatus draft.
	if _, err := svc.ClosePerformanceCycle(ctx, cycle.ID); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("close cycle with draft goal must fail: %v", err)
	}
	if err := svc.DeletePerformanceGoal(ctx, goalNoWeight.ID); err != nil {
		t.Fatalf("delete draft goal: %v", err)
	}

	closedCycle, err := svc.ClosePerformanceCycle(ctx, cycle.ID)
	if err != nil || closedCycle.Status != entity.PerformanceCycleStatusFinalized {
		t.Fatalf("close cycle: cycle=%+v err=%v", closedCycle, err)
	}
}
