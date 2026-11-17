package service_test

import (
	"errors"
	"testing"

	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
)

func TestRejectRequisitionRequiresIndependentDecisionAndPersistsReason(t *testing.T) {
	t.Parallel()
	h, svc, store := setup(t, &approvalVerifierStub{budgetID: "00000000-0000-4000-8000-000000000777"})
	created, err := svc.CreateRequisition(t.Context(), input())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.RejectRequisition(t.Context(), created.ID, "NOT_APPROVED"); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("draft rejection = %v", err)
	}
	if _, err := svc.SubmitRequisition(t.Context(), created.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.RejectRequisition(t.Context(), created.ID, " "); !errors.Is(err, hrerrors.ErrInvalidInput) {
		t.Fatalf("blank reason = %v", err)
	}
	if _, err := svc.RejectRequisition(t.Context(), created.ID, "NOT_APPROVED"); !errors.Is(err, hrerrors.ErrForbidden) {
		t.Fatalf("self rejection = %v", err)
	}
	h.Actor.PrincipalID = "00000000-0000-4000-8000-000000000998"
	rejected, err := svc.RejectRequisition(t.Context(), created.ID, " NOT_APPROVED ")
	if err != nil || rejected.Status != "rejected" || rejected.RejectedAt == nil ||
		rejected.RejectionReasonCode == nil || *rejected.RejectionReasonCode != "NOT_APPROVED" ||
		rejected.RejectedByPrincipalID == nil || *rejected.RejectedByPrincipalID != h.Actor.PrincipalID ||
		store.Rows[created.ID].Status != "rejected" || len(h.Events) != 3 {
		t.Fatalf("rejected = %+v, events = %d, error = %v", rejected, len(h.Events), err)
	}
	if _, err := svc.RejectRequisition(t.Context(), created.ID, "NOT_APPROVED"); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("repeat rejection = %v", err)
	}
	if _, err := svc.ApproveRequisition(t.Context(), created.ID); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("approval after rejection = %v", err)
	}
	if len(h.Events) != 3 {
		t.Fatalf("rejected retry wrote %d events", len(h.Events))
	}
}

func TestRejectRequisitionRollsBackWhenOutboxFails(t *testing.T) {
	t.Parallel()
	h, svc, store := setup(t)
	created, err := svc.CreateRequisition(t.Context(), input())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.SubmitRequisition(t.Context(), created.ID); err != nil {
		t.Fatal(err)
	}
	h.Actor.PrincipalID = "00000000-0000-4000-8000-000000000998"
	failure := errors.New("outbox unavailable")
	h.Failures["outbox"] = failure
	if _, err := svc.RejectRequisition(t.Context(), created.ID, "BUDGET_DENIED"); !errors.Is(err, failure) {
		t.Fatalf("outbox failure = %v", err)
	}
	if store.Rows[created.ID].Status != "submitted" || len(h.Events) != 2 {
		t.Fatal("failed outbox wrote rejection or event")
	}
}
