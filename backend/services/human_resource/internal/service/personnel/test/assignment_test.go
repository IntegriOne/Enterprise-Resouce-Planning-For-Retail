package service_test

import (
	"errors"
	"testing"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/personnel"
	mockdb "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/testsupport"
)

func TestCreateWorkerAssignmentRecordsHumanRequester(t *testing.T) {
	t.Parallel()
	harness, svc, _ := setup(t)
	input := service.CreateWorkerAssignmentInput{
		WorkerID: mockdb.ID(10), PositionID: mockdb.ID(11),
		AssignmentType: "primary", ValidFrom: mockdb.Now(),
	}
	created, err := svc.CreateWorkerAssignment(t.Context(), input)
	if err != nil {
		t.Fatalf("create assignment request: %v", err)
	}
	if created.RequestedByPrincipalID == nil || *created.RequestedByPrincipalID != harness.Actor.PrincipalID ||
		created.RequestedAt == nil || created.ApprovedAt != nil || created.ApprovedByPrincipalID != nil ||
		len(harness.Events) != 1 {
		t.Fatalf("assignment provenance or outbox missing: %+v, events=%d", created, len(harness.Events))
	}
	harness.Actor.Kind = "workload"
	if _, err := svc.CreateWorkerAssignment(t.Context(), input); !errors.Is(err, hrerrors.ErrForbidden) {
		t.Fatalf("workload assignment request: %v", err)
	}
	if len(harness.Events) != 1 {
		t.Fatalf("denied request added an outbox event: %d", len(harness.Events))
	}
}

func TestApproveWorkerAssignmentSeparatesRequesterAndApprover(t *testing.T) {
	t.Parallel()
	harness, svc, _ := setup(t)
	fte := entity.Decimal("60")
	created, err := svc.CreateWorkerAssignment(t.Context(), service.CreateWorkerAssignmentInput{
		WorkerID: mockdb.ID(10), PositionID: mockdb.ID(11), AssignmentType: "secondary",
		FTEPercent: &fte, ValidFrom: mockdb.Now(),
	})
	if err != nil {
		t.Fatalf("create assignment request: %v", err)
	}
	if _, err := svc.ApproveWorkerAssignment(t.Context(), created.ID); !errors.Is(err, hrerrors.ErrForbidden) {
		t.Fatalf("self approval: %v", err)
	}
	if len(harness.Events) != 1 {
		t.Fatalf("self approval wrote event: %d", len(harness.Events))
	}
	harness.Actor.PrincipalID = mockdb.ID(99)
	approved, err := svc.ApproveWorkerAssignment(t.Context(), created.ID)
	if err != nil || approved.ApprovedAt == nil || approved.ApprovedByPrincipalID == nil ||
		*approved.ApprovedByPrincipalID != harness.Actor.PrincipalID || len(harness.Events) != 2 {
		t.Fatalf("approval provenance and event = %+v, %v, events=%d", approved, err, len(harness.Events))
	}
	if _, err := svc.ApproveWorkerAssignment(t.Context(), created.ID); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("repeat approval: %v", err)
	}
	if len(harness.Events) != 2 {
		t.Fatalf("repeat approval wrote event: %d", len(harness.Events))
	}
}
