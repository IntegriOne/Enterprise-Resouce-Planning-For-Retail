package service_test

import (
	"errors"
	"testing"

	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	mockdb "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/testsupport"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/workforceplanning"
)

func TestActivatePositionUsesDedicatedCommand(t *testing.T) {
	t.Parallel()
	harness, svc, _ := setup(t)
	created, err := svc.CreatePosition(t.Context(), service.CreatePositionInput{
		PositionCode: "SEAT", PositionTitle: "Store associate", JobID: mockdb.ID(10),
		OrganizationUnitReferenceID: mockdb.ID(11), LegalEntityReferenceID: mockdb.ID(12),
		ValidFrom: mockdb.Now(),
	})
	if err != nil {
		t.Fatalf("create planned position: %v", err)
	}
	active := "active"
	if _, err := svc.UpdatePosition(t.Context(), created.ID, service.UpdatePositionInput{
		PositionStatus: operation.Field[string]{Set: true, Value: &active},
	}); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("activation through CRUD: %v", err)
	}
	if len(harness.Events) != 1 {
		t.Fatalf("denied CRUD wrote event: %d", len(harness.Events))
	}
	harness.Actor.Kind = "workload"
	if _, err := svc.ActivatePosition(t.Context(), created.ID); !errors.Is(err, hrerrors.ErrForbidden) {
		t.Fatalf("workload activation: %v", err)
	}
	harness.Actor.Kind = "human"
	activated, err := svc.ActivatePosition(t.Context(), created.ID)
	if err != nil || activated.PositionStatus != "active" || len(harness.Events) != 2 {
		t.Fatalf("activation result = %+v, %v, events=%d", activated, err, len(harness.Events))
	}
	if _, err := svc.ActivatePosition(t.Context(), created.ID); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("repeat activation: %v", err)
	}
	if len(harness.Events) != 2 {
		t.Fatalf("repeat activation wrote event: %d", len(harness.Events))
	}
}
