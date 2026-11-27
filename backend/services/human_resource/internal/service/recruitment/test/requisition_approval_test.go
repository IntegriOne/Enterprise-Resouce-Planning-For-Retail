package service_test

import (
	"context"
	"errors"
	"testing"

	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/recruitment"
)

type approvalVerifierStub struct {
	check            service.RequisitionApprovalCheck
	budgetID         string
	err              error
	calls            int
	mandateErr       error
	mandatePrincipal string
	mandateUnit      string
}

func (v *approvalVerifierStub) VerifyRequisitionApproval(_ context.Context, check service.RequisitionApprovalCheck) (string, error) {
	v.calls++
	v.check = check
	return v.budgetID, v.err
}

func (v *approvalVerifierStub) VerifyRequisitionMandate(_ context.Context, principal, unit string) error {
	v.mandatePrincipal, v.mandateUnit = principal, unit
	return v.mandateErr
}

func TestApproveRequisitionRequiresIndependentActorAndOwnerProof(t *testing.T) {
	t.Parallel()
	verifier := &approvalVerifierStub{budgetID: "00000000-0000-4000-8000-000000000777"}
	h, svc, store := setup(t, verifier)
	created, err := svc.CreateRequisition(t.Context(), input())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.SubmitRequisition(t.Context(), created.ID); err != nil {
		t.Fatal(err)
	}
	creatorID := h.Actor.PrincipalID
	if _, err := svc.ApproveRequisition(t.Context(), created.ID); !errors.Is(err, hrerrors.ErrForbidden) || verifier.calls != 0 {
		t.Fatalf("self approval: %v, verifier calls %d", err, verifier.calls)
	}
	h.Actor.PrincipalID = "00000000-0000-4000-8000-000000000998"
	verifier.err = errors.New("owner unavailable")
	if _, err := svc.ApproveRequisition(t.Context(), created.ID); !errors.Is(err, hrerrors.ErrUnavailable) {
		t.Fatalf("owner outage: %v", err)
	}
	if store.Rows[created.ID].Status != "submitted" || len(h.Events) != 2 {
		t.Fatal("owner outage committed approval or outbox")
	}
	verifier.err = nil
	verifier.budgetID = "invalid"
	if _, err := svc.ApproveRequisition(t.Context(), created.ID); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("invalid owner proof: %v", err)
	}
	verifier.budgetID = "00000000-0000-4000-8000-000000000777"
	approved, err := svc.ApproveRequisition(t.Context(), created.ID)
	if err != nil || approved.Status != "approved" || approved.ApprovedAt == nil ||
		approved.ApprovedByPrincipalID == nil || *approved.ApprovedByPrincipalID != h.Actor.PrincipalID ||
		approved.BudgetReferenceID == nil || *approved.BudgetReferenceID != verifier.budgetID || len(h.Events) != 3 {
		t.Fatalf("approval = %+v, error %v", approved, err)
	}
	if verifier.check.RequisitionID != created.ID || verifier.check.CreatedByPrincipalID != creatorID ||
		verifier.check.SubmittedByPrincipalID != creatorID || verifier.check.ApproverPrincipalID != h.Actor.PrincipalID ||
		verifier.check.Headcount != 1 || verifier.check.ExpectedBudgetReferenceID != "" {
		t.Fatalf("owner check = %+v", verifier.check)
	}
	if _, err := svc.ApproveRequisition(t.Context(), created.ID); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("repeated approval: %v", err)
	}
}

func TestApproveRequisitionWithoutOwnerVerifierFailsClosed(t *testing.T) {
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
	if _, err := svc.ApproveRequisition(t.Context(), created.ID); !errors.Is(err, hrerrors.ErrUnavailable) {
		t.Fatalf("missing owner verifier: %v", err)
	}
	if store.Rows[created.ID].Status != "submitted" || len(h.Events) != 2 {
		t.Fatal("missing owner verifier changed requisition")
	}
}
