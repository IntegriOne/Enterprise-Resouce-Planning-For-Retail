package service_test

import (
	"errors"
	"testing"

	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
)

func TestOpenRequisitionRechecksOwnerProofAndPreservesBudget(t *testing.T) {
	t.Parallel()
	verifier := &approvalVerifierStub{budgetID: "00000000-0000-4000-8000-000000000777"}
	h, svc, store := setup(t, verifier)
	created, err := svc.CreateRequisition(t.Context(), input())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.OpenRequisition(t.Context(), created.ID); !errors.Is(err, hrerrors.ErrWorkflowRequired) || verifier.calls != 0 {
		t.Fatalf("open draft: %v, verifier calls %d", err, verifier.calls)
	}
	if _, err := svc.SubmitRequisition(t.Context(), created.ID); err != nil {
		t.Fatal(err)
	}
	h.Actor.PrincipalID = "00000000-0000-4000-8000-000000000998"
	if _, err := svc.ApproveRequisition(t.Context(), created.ID); err != nil {
		t.Fatal(err)
	}
	h.Actor.PrincipalID = "00000000-0000-4000-8000-000000000999"
	verifier.err = errors.New("owner unavailable")
	if _, err := svc.OpenRequisition(t.Context(), created.ID); !errors.Is(err, hrerrors.ErrUnavailable) {
		t.Fatalf("owner outage: %v", err)
	}
	verifier.err = nil
	verifier.budgetID = "00000000-0000-4000-8000-000000000778"
	if _, err := svc.OpenRequisition(t.Context(), created.ID); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("changed budget proof: %v", err)
	}
	if store.Rows[created.ID].Status != "approved" || len(h.Events) != 3 {
		t.Fatal("failed opening changed requisition or outbox")
	}
	verifier.budgetID = "00000000-0000-4000-8000-000000000777"
	opened, err := svc.OpenRequisition(t.Context(), created.ID)
	if err != nil || opened.Status != "open" || opened.OpenedAt == nil ||
		opened.BudgetReferenceID == nil || *opened.BudgetReferenceID != verifier.budgetID || len(h.Events) != 4 {
		t.Fatalf("opened requisition = %+v, error = %v", opened, err)
	}
	if verifier.check.ApproverPrincipalID != "00000000-0000-4000-8000-000000000998" {
		t.Fatalf("proof approver = %q", verifier.check.ApproverPrincipalID)
	}
	if _, err := svc.OpenRequisition(t.Context(), created.ID); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("repeated opening: %v", err)
	}
}
