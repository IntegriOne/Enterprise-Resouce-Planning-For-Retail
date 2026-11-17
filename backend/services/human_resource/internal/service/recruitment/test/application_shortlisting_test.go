package service_test

import (
	"errors"
	"testing"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/recruitment"
)

func TestShortlistApplicationRequiresScreeningAndAssignedRecruiter(t *testing.T) {
	t.Parallel()
	h, svc, _ := setup(t)
	created, err := svc.CreateApplication(t.Context(), service.CreateApplicationInput{
		CandidateID:       "00000000-0000-4000-8000-000000000601",
		RequisitionID:     "00000000-0000-4000-8000-000000000602",
		ApplicationNumber: "APP-SHORTLIST-1", IdempotencyKey: "APP-SHORTLIST-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.ShortlistApplication(t.Context(), created.ID, "MEETS_REQUIREMENTS"); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("shortlist before screening = %v", err)
	}
	if _, err := svc.StartApplicationScreening(t.Context(), created.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.ShortlistApplication(t.Context(), created.ID, " "); !errors.Is(err, hrerrors.ErrInvalidInput) {
		t.Fatalf("shortlist without reason = %v", err)
	}
	assigned := h.Actor.PrincipalID
	h.Actor.PrincipalID = "00000000-0000-4000-8000-000000000603"
	if _, err := svc.ShortlistApplication(t.Context(), created.ID, "MEETS_REQUIREMENTS"); !errors.Is(err, hrerrors.ErrForbidden) {
		t.Fatalf("other recruiter = %v", err)
	}
	h.Actor.PrincipalID = assigned
	shortlisted, err := svc.ShortlistApplication(t.Context(), created.ID, "MEETS_REQUIREMENTS")
	if err != nil || shortlisted.Status != entity.ApplicationStatusShortlisted ||
		shortlisted.CurrentStageCode == nil || *shortlisted.CurrentStageCode != "shortlisted" ||
		shortlisted.ScreeningScore != nil || shortlisted.ScorecardVersion != nil || len(h.Events) != 3 {
		t.Fatalf("shortlisted application = %+v, events = %d, error = %v", shortlisted, len(h.Events), err)
	}
	history, err := svc.FindApplicationStageHistoryByApplicationID(t.Context(), created.ID, 10, 0)
	if err != nil || len(history.Items) != 2 || history.Items[1].NewStatus != entity.ApplicationStatusShortlisted ||
		history.Items[1].ReasonCode == nil || *history.Items[1].ReasonCode != "MEETS_REQUIREMENTS" ||
		history.Items[1].ChangedByPrincipalID == nil || *history.Items[1].ChangedByPrincipalID != assigned {
		t.Fatalf("shortlist history = %+v, error = %v", history, err)
	}
	if _, err := svc.ShortlistApplication(t.Context(), created.ID, "MEETS_REQUIREMENTS"); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("repeat shortlist = %v", err)
	}
	if len(h.Events) != 3 {
		t.Fatalf("repeat shortlist wrote %d events", len(h.Events))
	}
	rejected, err := svc.RejectApplication(t.Context(), created.ID, "NO_LONGER_AVAILABLE")
	if err != nil || rejected.Status != entity.ApplicationStatusRejected || len(h.Events) != 4 {
		t.Fatalf("reject shortlisted application = %+v, events = %d, error = %v", rejected, len(h.Events), err)
	}
}
