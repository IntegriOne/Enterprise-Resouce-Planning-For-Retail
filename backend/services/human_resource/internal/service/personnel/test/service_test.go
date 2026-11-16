// Kerangka tes awal dibuat dengan alat bantu; tes ini dipelihara langsung.
// Fixture dan dependency pada file ini hanya untuk pengujian service.
package service_test

import (
	"context"
	"errors"
	"testing"

	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/personnel"
)

func TestServiceCRUDAndPagination(t *testing.T) {
	t.Parallel()
	h, svc, store := setup(t)
	ctx := context.Background()
	created, err := svc.CreateWorker(ctx, input())
	if err != nil {
		t.Fatal(err)
	}
	if len(store.Rows) != 1 || len(h.Events) != 1 || h.IntegrityCalls != 1 {
		t.Fatal("create must atomically persist record and outbox after integrity validation")
	}
	key := created.ID
	loaded, err := svc.GetWorkerByID(ctx, key)
	if err != nil || loaded == nil {
		t.Fatalf("get: %v", err)
	}
	replacement := "UPDATED"
	_, err = svc.UpdateWorker(ctx, key, servicePatch(replacement))
	if err != nil {
		t.Fatal(err)
	}
	loaded, err = svc.GetWorkerByID(ctx, key)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.WorkerNumber != replacement {
		t.Fatal("patch value was not persisted")
	}
	page, err := svc.GetAllWorkers(ctx, 1, 0)
	if err != nil || len(page.Items) != 1 || page.Total != 1 || page.Limit != 1 {
		t.Fatalf("page: %+v %v", page, err)
	}
	empty, err := svc.GetAllWorkers(ctx, 1, 1)
	if err != nil || empty.Items == nil || len(empty.Items) != 0 {
		t.Fatalf("empty page: %+v %v", empty, err)
	}
	_, err = svc.GetAllWorkers(ctx, -1, 0)
	if !errors.Is(err, hrerrors.ErrInvalidInput) {
		t.Fatal("negative pagination accepted")
	}
	if err := svc.DeleteWorker(ctx, key); err != nil {
		t.Fatal(err)
	}
	if len(store.Rows) != 0 || len(h.Events) != 3 {
		t.Fatal("delete must persist alongside its event")
	}
}
func servicePatch(value string) service.UpdateWorkerInput {
	return service.UpdateWorkerInput{WorkerNumber: operation.Field[string]{Set: true, Value: &value}}
}
func TestServiceFailureAndRollback(t *testing.T) {
	t.Parallel()
	cases := []string{"authorize", "integrity", "insert", "outbox", "outbox_read", "commit"}
	for _, stage := range cases {
		t.Run(stage, func(t *testing.T) {
			h, svc, store := setup(t)
			cause := errors.New("injected failure")
			h.Failures[stage] = cause
			result, err := svc.CreateWorker(context.Background(), input())
			if err != cause || result != nil {
				t.Fatalf("error propagation: %v", err)
			}
			if len(store.Rows) != 0 || len(h.Events) != 0 {
				t.Fatal("failed operation left committed data")
			}
			if stage == "authorize" && (h.WriteCalls != 0 || h.TransactionCalls != 0) {
				t.Fatal("denied operation reached database transaction")
			}
		})
	}
}
func TestServiceCancellationAndScope(t *testing.T) {
	t.Parallel()
	h, svc, store := setup(t)
	ctx := context.Background()
	created, err := svc.CreateWorker(ctx, input())
	if err != nil {
		t.Fatal(err)
	}
	key := created.ID
	h.DeniedObject = created.ID
	if _, err := svc.GetWorkerByID(ctx, key); !errors.Is(err, hrerrors.ErrForbidden) {
		t.Fatalf("wrong object scope: %v", err)
	}
	h.DeniedObject = ""
	h.Scope = operation.Scope{}
	if _, err := svc.GetAllWorkers(ctx, 20, 0); !errors.Is(err, hrerrors.ErrForbidden) {
		t.Fatalf("missing scope: %v", err)
	}
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := svc.CreateWorker(canceled, input()); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation: %v", err)
	}
	if len(store.Rows) != 1 || len(h.Events) != 1 {
		t.Fatal("denied calls changed committed data")
	}
}

func TestWorkerScopedGetAndListFilterBeforeReading(t *testing.T) {
	t.Parallel()
	h, svc, _ := setup(t)
	first, err := svc.CreateWorker(t.Context(), input())
	if err != nil {
		t.Fatal(err)
	}
	secondInput := input()
	secondInput.WorkerNumber = "OTHER-LEGAL-ENTITY"
	secondInput.PrimaryLegalEntityReferenceID = "00000000-0000-0000-0000-000000000021"
	second, err := svc.CreateWorker(t.Context(), secondInput)
	if err != nil {
		t.Fatal(err)
	}
	h.Scope = operation.Scope{Filters: []domainrepo.Filter{{Column: "primary_legal_entity_reference_id", Values: []any{input().PrimaryLegalEntityReferenceID}}}}
	if _, err := svc.GetWorkerByID(t.Context(), second.ID); !errors.Is(err, hrerrors.ErrNotFound) {
		t.Fatalf("outside-scope worker lookup = %v, want not found", err)
	}
	worker, err := svc.GetWorkerByID(t.Context(), first.ID)
	if err != nil || worker == nil || worker.ID != first.ID {
		t.Fatalf("inside-scope worker lookup = %+v, %v", worker, err)
	}
	page, err := svc.GetAllWorkers(t.Context(), 1, 0)
	if err != nil || page.Total != 1 || len(page.Items) != 1 || page.Items[0].ID != first.ID {
		t.Fatalf("scoped list = %+v, %v", page, err)
	}
}
