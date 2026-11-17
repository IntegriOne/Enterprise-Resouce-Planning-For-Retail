package service_test

import (
	"errors"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/personnel"
	mockdb "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/testsupport"
)

func TestActivateWorkerRequiresDatesAndHumanActor(t *testing.T) {
	t.Parallel()
	harness, svc, workers := setup(t)
	start := mockdb.Now().AddDate(0, 0, -1)
	in := input()
	in.HireDate = &start
	in.ServiceStartDate = &start
	created, err := svc.CreateWorker(t.Context(), in)
	if err != nil {
		t.Fatal(err)
	}
	harness.Actor.Kind = "workload"
	if _, err := svc.ActivateWorker(t.Context(), created.ID); !errors.Is(err, hrerrors.ErrForbidden) {
		t.Fatalf("workload activation: %v", err)
	}
	harness.Actor.Kind = "human"
	active, err := svc.ActivateWorker(t.Context(), created.ID)
	if err != nil || active.Status != entity.WorkerStatusActive || len(harness.Events) != 2 {
		t.Fatalf("activate worker = %+v, %v, events=%d", active, err, len(harness.Events))
	}
	if _, err := svc.ActivateWorker(t.Context(), created.ID); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("repeat activation: %v", err)
	}
	if len(harness.Events) != 2 {
		t.Fatalf("repeat activation wrote event: %d", len(harness.Events))
	}
	row := workers.Rows[created.ID]
	row.Status = entity.WorkerStatusPreboarding
	future := mockdb.Now().Add(24 * time.Hour)
	row.ServiceStartDate = &future
	workers.Rows[created.ID] = row
	if _, err := svc.ActivateWorker(t.Context(), created.ID); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("future service start: %v", err)
	}
	if _, err := svc.UpdateWorker(t.Context(), created.ID, service.UpdateWorkerInput{}); err != nil {
		t.Fatalf("draft correction remains available: %v", err)
	}
}
