package service_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	fixture "github.com/integrione/erp-retail/backend/services/human_resource/domain/service/test/fixture"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/leave"
	mockdb "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/testsupport"
)

func ledgerSetup(tb testing.TB) (*mockdb.Harness, *service.Service, *mockdb.Store[entity.LeaveEntitlement, string], *mockdb.Store[entity.LeaveLedgerEntry, string]) {
	tb.Helper()
	h := mockdb.New()
	policies := mockdb.NewStore[entity.LeavePolicy, string](h, func(v *entity.LeavePolicy) string { return v.ID })
	entitlements := mockdb.NewStore[entity.LeaveEntitlement, string](h, func(v *entity.LeaveEntitlement) string { return v.ID })
	adjustments := mockdb.NewStore[entity.LeaveAdjustment, string](h, func(v *entity.LeaveAdjustment) string { return v.ID })
	entries := mockdb.NewStore[entity.LeaveLedgerEntry, string](h, func(v *entity.LeaveLedgerEntry) string { return v.ID })
	svc, err := service.NewService(service.Dependencies{
		Runtime: h.Runtime(), LeavePolicy: policies, LeaveEntitlement: entitlements,
		LeaveAdjustment: adjustments, LeaveLedgerEntry: entries,
	})
	if err != nil {
		tb.Fatal(err)
	}
	initial := fixture.Entitlement()
	entitlements.Rows[initial.ID] = initial
	return h, svc, entitlements, entries
}

func openingEntry() service.PostLedgerEntryInput {
	return service.PostLedgerEntryInput{
		EntryType: "opening", QuantityDelta: "12", EffectiveDate: mockdb.Now(),
		SourceManagement: "hrm", SourceEventID: "opening-event", SourceLineKey: "balance",
		SourceDigest: []byte{1}, ReasonCode: "OPENING",
	}
}

func TestPostLedgerEntryCommitsProjectionAndEvent(t *testing.T) {
	h, svc, entitlements, entries := ledgerSetup(t)
	id := fixture.Entitlement().ID
	result, err := svc.PostLedgerEntry(context.Background(), id, openingEntry())
	if err != nil {
		t.Fatal(err)
	}
	if result.ClosingBalanceSnapshot != "12.0000" || result.Revision != 2 || len(entries.Rows) != 1 || len(h.Events) != 1 {
		t.Fatalf("posting did not commit one ledger/projection/event: %+v", result)
	}
	_, err = svc.PostLedgerEntry(context.Background(), id, openingEntry())
	if !errors.Is(err, hrerrors.ErrAlreadyExists) || len(entries.Rows) != 1 || entitlements.Rows[id].Revision != 2 {
		t.Fatalf("duplicate source changed balance: %v", err)
	}
	changedDigest := openingEntry()
	changedDigest.SourceDigest = []byte{2}
	if _, err := svc.PostLedgerEntry(context.Background(), id, changedDigest); !errors.Is(err, hrerrors.ErrConflict) {
		t.Fatalf("changed digest was not rejected: %v", err)
	}
}

func TestPostLedgerEntryRollsBackEveryWrite(t *testing.T) {
	h, svc, entitlements, entries := ledgerSetup(t)
	cause := errors.New("outbox failed")
	h.Failures["outbox"] = cause
	id := fixture.Entitlement().ID
	if _, err := svc.PostLedgerEntry(context.Background(), id, openingEntry()); err != cause {
		t.Fatalf("cause lost: %v", err)
	}
	if len(entries.Rows) != 0 || len(h.Events) != 0 || entitlements.Rows[id].ClosingBalanceSnapshot != "0" {
		t.Fatal("ledger/projection survived a failed outbox")
	}
}

func TestConcurrentLeaveUsageDoesNotLoseBalance(t *testing.T) {
	h, svc, entitlements, entries := ledgerSetup(t)
	ctx := context.Background()
	id := fixture.Entitlement().ID
	if _, err := svc.PostLedgerEntry(ctx, id, openingEntry()); err != nil {
		t.Fatal(err)
	}
	start := make(chan struct{})
	failures := make(chan error, 16)
	var group sync.WaitGroup
	for i := 0; i < 16; i++ {
		group.Add(1)
		go func(index int) {
			defer group.Done()
			<-start
			entry := openingEntry()
			entry.EntryType = "usage"
			entry.QuantityDelta = "-0.25"
			entry.SourceEventID = fmt.Sprintf("usage-%d", index)
			_, err := svc.PostLedgerEntry(ctx, id, entry)
			failures <- err
		}(i)
	}
	close(start)
	group.Wait()
	close(failures)
	for err := range failures {
		if err != nil {
			t.Fatal(err)
		}
	}
	if entitlements.Rows[id].ClosingBalanceSnapshot != "8.0000" || len(entries.Rows) != 17 || len(h.Events) != 17 {
		t.Fatal("concurrent posting lost a balance change or event")
	}
}
