// Kerangka tes awal dibuat dengan alat bantu; tes ini dipelihara langsung.
// Fixture dan dependency pada file ini hanya untuk pengujian service.
package service_test

import (
	"context"
	"errors"
	"testing"

	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/recruitment"
)

func TestServiceCRUDAndPagination(t *testing.T) {
	t.Parallel()
	h, svc, store := setup(t)
	ctx := context.Background()
	created, err := svc.CreateRequisition(ctx, input())
	if err != nil {
		t.Fatal(err)
	}
	if len(store.Rows) != 1 || len(h.Events) != 1 || h.IntegrityCalls != 1 {
		t.Fatal("create must atomically persist record and outbox after integrity validation")
	}
	key := created.ID
	loaded, err := svc.GetRequisitionByID(ctx, key)
	if err != nil || loaded == nil {
		t.Fatalf("get: %v", err)
	}
	replacement := "UPDATED"
	_, err = svc.UpdateRequisition(ctx, key, servicePatch(replacement))
	if err != nil {
		t.Fatal(err)
	}
	loaded, err = svc.GetRequisitionByID(ctx, key)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.RequisitionNumber != replacement {
		t.Fatal("patch value was not persisted")
	}
	page, err := svc.GetAllRequisitions(ctx, 1, 0)
	if err != nil || len(page.Items) != 1 || page.Total != 1 || page.Limit != 1 {
		t.Fatalf("page: %+v %v", page, err)
	}
	empty, err := svc.GetAllRequisitions(ctx, 1, 1)
	if err != nil || empty.Items == nil || len(empty.Items) != 0 {
		t.Fatalf("empty page: %+v %v", empty, err)
	}
	_, err = svc.GetAllRequisitions(ctx, -1, 0)
	if !errors.Is(err, hrerrors.ErrInvalidInput) {
		t.Fatal("negative pagination accepted")
	}
	if err := svc.DeleteRequisition(ctx, key); err != nil {
		t.Fatal(err)
	}
	if len(store.Rows) != 0 || len(h.Events) != 3 {
		t.Fatal("delete must persist alongside its event")
	}
}
func servicePatch(value string) service.UpdateRequisitionInput {
	return service.UpdateRequisitionInput{RequisitionNumber: operation.Field[string]{Set: true, Value: &value}}
}
func TestServiceFailureAndRollback(t *testing.T) {
	t.Parallel()
	cases := []string{"authorize", "integrity", "insert", "outbox", "outbox_read", "commit"}
	for _, stage := range cases {
		t.Run(stage, func(t *testing.T) {
			h, svc, store := setup(t)
			cause := errors.New("injected failure")
			h.Failures[stage] = cause
			result, err := svc.CreateRequisition(context.Background(), input())
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
	created, err := svc.CreateRequisition(ctx, input())
	if err != nil {
		t.Fatal(err)
	}
	key := created.ID
	h.DeniedObject = created.ID
	if _, err := svc.GetRequisitionByID(ctx, key); !errors.Is(err, hrerrors.ErrForbidden) {
		t.Fatalf("wrong object scope: %v", err)
	}
	h.DeniedObject = ""
	h.Scope = operation.Scope{}
	if _, err := svc.GetAllRequisitions(ctx, 20, 0); !errors.Is(err, hrerrors.ErrForbidden) {
		t.Fatalf("missing scope: %v", err)
	}
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := svc.CreateRequisition(canceled, input()); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation: %v", err)
	}
	if len(store.Rows) != 1 || len(h.Events) != 1 {
		t.Fatal("denied calls changed committed data")
	}
}
