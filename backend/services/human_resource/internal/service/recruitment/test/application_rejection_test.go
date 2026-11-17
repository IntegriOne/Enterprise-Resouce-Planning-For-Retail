package service_test

import (
	"errors"
	"testing"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/recruitment"
)

func TestRejectApplicationWritesDecisionHistoryAndOutbox(t *testing.T) {
	t.Parallel()
	h, svc, _ := setup(t)
	created, err := svc.CreateApplication(t.Context(), service.CreateApplicationInput{
		CandidateID:       "00000000-0000-4000-8000-000000000501",
		RequisitionID:     "00000000-0000-4000-8000-000000000502",
		ApplicationNumber: "APP-REJECT-1", IdempotencyKey: "APP-REJECT-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.RejectApplication(t.Context(), created.ID, " "); !errors.Is(err, hrerrors.ErrInvalidInput) {
		t.Fatalf("blank reason = %v", err)
	}
	rejected, err := svc.RejectApplication(t.Context(), created.ID, "POSITION_CANCELLED")
	if err != nil || rejected.Status != entity.ApplicationStatusRejected || rejected.DecisionAt == nil ||
		rejected.DecisionByPrincipalID == nil || *rejected.DecisionByPrincipalID != h.Actor.PrincipalID ||
		rejected.DispositionReasonCode == nil || *rejected.DispositionReasonCode != "POSITION_CANCELLED" || len(h.Events) != 2 {
		t.Fatalf("rejected application = %+v, events = %d, error = %v", rejected, len(h.Events), err)
	}
	history, err := svc.FindApplicationStageHistoryByApplicationID(t.Context(), created.ID, 10, 0)
	if err != nil || len(history.Items) != 1 || history.Items[0].NewStatus != entity.ApplicationStatusRejected ||
		history.Items[0].ReasonCode == nil || *history.Items[0].ReasonCode != "POSITION_CANCELLED" {
		t.Fatalf("rejection history = %+v, error = %v", history, err)
	}
	if _, err := svc.RejectApplication(t.Context(), created.ID, "POSITION_CANCELLED"); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("repeat rejection = %v", err)
	}
	if len(h.Events) != 2 {
		t.Fatalf("repeat rejection wrote %d events", len(h.Events))
	}
}

func TestRejectScreeningApplicationRequiresAssignedRecruiter(t *testing.T) {
	t.Parallel()
	h, svc, _ := setup(t)
	created, err := svc.CreateApplication(t.Context(), service.CreateApplicationInput{
		CandidateID:       "00000000-0000-4000-8000-000000000503",
		RequisitionID:     "00000000-0000-4000-8000-000000000504",
		ApplicationNumber: "APP-REJECT-2", IdempotencyKey: "APP-REJECT-2",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.StartApplicationScreening(t.Context(), created.ID); err != nil {
		t.Fatal(err)
	}
	assigned := h.Actor.PrincipalID
	h.Actor.PrincipalID = "00000000-0000-4000-8000-000000000505"
	if _, err := svc.RejectApplication(t.Context(), created.ID, "NOT_ELIGIBLE"); !errors.Is(err, hrerrors.ErrForbidden) {
		t.Fatalf("other recruiter = %v", err)
	}
	h.Actor.PrincipalID = assigned
	if _, err := svc.RejectApplication(t.Context(), created.ID, "NOT_ELIGIBLE"); err != nil {
		t.Fatalf("assigned recruiter rejection = %v", err)
	}
}
