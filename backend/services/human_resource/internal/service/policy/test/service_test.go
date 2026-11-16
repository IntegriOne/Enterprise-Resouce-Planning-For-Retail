// Kerangka tes awal dibuat dengan alat bantu; tes ini dipelihara langsung.
// Fixture dan dependency pada file ini hanya untuk pengujian service.
package service_test

import (
	"context"
	"errors"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/policy"
	mockdb "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/testsupport"
	"testing"
)

func TestServiceCRUDAndPagination(t *testing.T) {
	t.Parallel()
	h, svc, store := setup(t)
	ctx := context.Background()
	created, err := svc.CreatePolicyAcknowledgement(ctx, input())
	if err != nil {
		t.Fatal(err)
	}
	if len(store.Rows) != 1 || len(h.Events) != 1 || h.IntegrityCalls != 1 {
		t.Fatal("create must atomically persist record and outbox after integrity validation")
	}
	key := created.ID
	loaded, err := svc.GetPolicyAcknowledgementByID(ctx, key)
	if err != nil || loaded == nil {
		t.Fatalf("get: %v", err)
	}
	replacement := mockdb.ID(555)
	_, err = svc.UpdatePolicyAcknowledgement(ctx, key, servicePatch(replacement))
	if err != nil {
		t.Fatal(err)
	}
	loaded, err = svc.GetPolicyAcknowledgementByID(ctx, key)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.PolicyID != replacement {
		t.Fatal("patch value was not persisted")
	}
	page, err := svc.GetAllPolicyAcknowledgements(ctx, 1, 0)
	if err != nil || len(page.Items) != 1 || page.Total != 1 || page.Limit != 1 {
		t.Fatalf("page: %+v %v", page, err)
	}
	empty, err := svc.GetAllPolicyAcknowledgements(ctx, 1, 1)
	if err != nil || empty.Items == nil || len(empty.Items) != 0 {
		t.Fatalf("empty page: %+v %v", empty, err)
	}
	_, err = svc.GetAllPolicyAcknowledgements(ctx, -1, 0)
	if !errors.Is(err, hrerrors.ErrInvalidInput) {
		t.Fatal("negative pagination accepted")
	}
	if err := svc.DeletePolicyAcknowledgement(ctx, key); err != nil {
		t.Fatal(err)
	}
	if len(store.Rows) != 0 || len(h.Events) != 3 {
		t.Fatal("delete must persist alongside its event")
	}
}
func servicePatch(value string) service.UpdatePolicyAcknowledgementInput {
	return service.UpdatePolicyAcknowledgementInput{PolicyID: operation.Field[string]{Set: true, Value: &value}}
}
func TestServiceFailureAndRollback(t *testing.T) {
	t.Parallel()
	cases := []string{"authorize", "integrity", "insert", "outbox", "outbox_read", "commit"}
	for _, stage := range cases {
		t.Run(stage, func(t *testing.T) {
			h, svc, store := setup(t)
			cause := errors.New("injected failure")
			h.Failures[stage] = cause
			result, err := svc.CreatePolicyAcknowledgement(context.Background(), input())
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
	created, err := svc.CreatePolicyAcknowledgement(ctx, input())
	if err != nil {
		t.Fatal(err)
	}
	key := created.ID
	h.DeniedObject = created.ID
	if _, err := svc.GetPolicyAcknowledgementByID(ctx, key); !errors.Is(err, hrerrors.ErrForbidden) {
		t.Fatalf("wrong object scope: %v", err)
	}
	h.DeniedObject = ""
	h.Scope = operation.Scope{}
	if _, err := svc.GetAllPolicyAcknowledgements(ctx, 20, 0); !errors.Is(err, hrerrors.ErrForbidden) {
		t.Fatalf("missing scope: %v", err)
	}
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := svc.CreatePolicyAcknowledgement(canceled, input()); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation: %v", err)
	}
	if len(store.Rows) != 1 || len(h.Events) != 1 {
		t.Fatal("denied calls changed committed data")
	}
}
