package service_test

import (
	"errors"
	"testing"

	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
)

func TestCloseRequisitionRequiresOpenAndWritesOneEvent(t *testing.T) {
	t.Parallel()
	verifier := &approvalVerifierStub{budgetID: "00000000-0000-4000-8000-000000000777"}
	h, svc, store := setup(t, verifier)
	created, err := svc.CreateRequisition(t.Context(), input())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CloseRequisition(t.Context(), created.ID); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("close draft: %v", err)
	}
	if _, err := svc.SubmitRequisition(t.Context(), created.ID); err != nil {
		t.Fatal(err)
	}
	h.Actor.PrincipalID = "00000000-0000-4000-8000-000000000998"
	if _, err := svc.ApproveRequisition(t.Context(), created.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CloseRequisition(t.Context(), created.ID); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("close before open: %v", err)
	}
	if _, err := svc.OpenRequisition(t.Context(), created.ID); err != nil {
		t.Fatal(err)
	}
	closed, err := svc.CloseRequisition(t.Context(), created.ID)
	if err != nil || closed.Status != "closed" || closed.ClosedAt == nil ||
		closed.OpenedAt == nil || closed.ClosedAt.Before(*closed.OpenedAt) || len(h.Events) != 5 {
		t.Fatalf("closed requisition = %+v, error = %v", closed, err)
	}
	if _, err := svc.CloseRequisition(t.Context(), created.ID); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("repeat close: %v", err)
	}
	if store.Rows[created.ID].Status != "closed" || len(h.Events) != 5 {
		t.Fatal("repeat close changed requisition or outbox")
	}
}
