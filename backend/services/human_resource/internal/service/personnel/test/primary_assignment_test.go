package service_test

import (
	"errors"
	"testing"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/personnel"
	mockdb "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/testsupport"
)

func TestApplyPrimaryAssignmentRequiresDedicatedHumanCommand(t *testing.T) {
	t.Parallel()
	harness, svc, workers := setup(t)
	created, err := svc.CreateWorker(t.Context(), input())
	if err != nil {
		t.Fatalf("create preboarding worker: %v", err)
	}
	assignmentID := mockdb.ID(90)
	if _, err := svc.ApplyPrimaryAssignment(t.Context(), created.ID, assignmentID); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("preboarding worker assignment: %v", err)
	}
	row := workers.Rows[created.ID]
	row.Status = entity.WorkerStatusActive
	workers.Rows[created.ID] = row
	if _, err := svc.UpdateWorker(t.Context(), created.ID, service.UpdateWorkerInput{
		PrimaryAssignmentID: operation.Field[string]{Set: true, Value: &assignmentID},
	}); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("generic worker update: %v", err)
	}
	harness.Actor.Kind = "workload"
	if _, err := svc.ApplyPrimaryAssignment(t.Context(), created.ID, assignmentID); !errors.Is(err, hrerrors.ErrForbidden) {
		t.Fatalf("workload assignment application: %v", err)
	}
	harness.Actor.Kind = "human"
	applied, err := svc.ApplyPrimaryAssignment(t.Context(), created.ID, assignmentID)
	if err != nil || applied.PrimaryAssignmentID == nil || *applied.PrimaryAssignmentID != assignmentID || len(harness.Events) != 2 {
		t.Fatalf("apply primary assignment = %+v, %v, events=%d", applied, err, len(harness.Events))
	}
	if _, err := svc.ApplyPrimaryAssignment(t.Context(), created.ID, assignmentID); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("repeat assignment application: %v", err)
	}
	if len(harness.Events) != 2 {
		t.Fatalf("repeat command wrote event: %d", len(harness.Events))
	}
	history, err := svc.FindWorkerLifecycleEventByWorkerID(t.Context(), created.ID, 10, 0)
	if err != nil || len(history.Items) != 0 {
		t.Fatalf("initial assignment lifecycle = %+v, %v", history, err)
	}
	replacementID := mockdb.ID(91)
	if _, err := svc.ApplyPrimaryAssignment(t.Context(), created.ID, replacementID); err != nil {
		t.Fatalf("apply replacement assignment: %v", err)
	}
	history, err = svc.FindWorkerLifecycleEventByWorkerID(t.Context(), created.ID, 10, 0)
	if err != nil || len(history.Items) != 1 {
		t.Fatalf("replacement assignment lifecycle = %+v, %v", history, err)
	}
	transfer := history.Items[0]
	if transfer.EventType != "transfer" || transfer.SourceReference == nil || *transfer.SourceReference != replacementID {
		t.Fatalf("unexpected transfer event: %+v", transfer)
	}
}
