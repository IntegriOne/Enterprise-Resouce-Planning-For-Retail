package service_test

import (
	"testing"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	service "github.com/integrione/erp-retail/backend/services/human_resource/domain/service/leave"
	f "github.com/integrione/erp-retail/backend/services/human_resource/domain/service/test/fixture"
)

func TestLedgerReplay(t *testing.T) {
	t.Parallel()
	input := service.BalanceInput{Entries: []entity.LeaveLedgerEntry{f.Entry(1, "opening", "12"), f.Entry(2, "usage", "-2"), f.Entry(3, "reversal", "2")}, CalculatedAt: f.Date()}
	input.Entries[2].ReversesEntryID = &input.Entries[1].ID
	tests := []struct {
		name   string
		mutate func(*service.BalanceInput)
		valid  bool
	}{
		{"exact reversal", func(_ *service.BalanceInput) {}, true},
		{"partial reversal", func(i *service.BalanceInput) { i.Entries[2].QuantityDelta = "1" }, false},
		{"gap", func(i *service.BalanceInput) { i.Entries[2].EntrySequence = 4 }, false},
		{"foreign entitlement", func(i *service.BalanceInput) { i.Entries[1].EntitlementID = f.ID(999) }, false},
		{"negative balance", func(i *service.BalanceInput) { i.Entries = i.Entries[:2]; i.Entries[1].QuantityDelta = "-13" }, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			current := f.Entitlement()
			i := input
			i.Entitlement = &current
			i.Entries = append([]entity.LeaveLedgerEntry{}, input.Entries...)
			tt.mutate(&i)
			result, err := service.RebuildBalance(i)
			if (err == nil) != tt.valid {
				t.Fatalf("rebuild: %v", err)
			}
			if err == nil && result.ClosingBalanceSnapshot != "12.0000" {
				t.Fatalf("balance=%s", result.ClosingBalanceSnapshot)
			}
			if current.ClosingBalanceSnapshot != "0" || current.Revision != 1 {
				t.Fatal("input snapshot mutated")
			}
		})
	}
}

func TestApprovedAdjustmentIsPostedOnce(t *testing.T) {
	t.Parallel()
	current := f.Entitlement()
	at := f.Date()
	approver := f.ID(500)
	adjustment := entity.LeaveAdjustment{ID: f.ID(501), EntitlementID: current.ID, QuantityDelta: "1", ReasonCode: "CORRECTION", Reason: "Approved correction", RequestedByPrincipalID: f.ID(502), ApprovedByPrincipalID: &approver, ApprovedAt: &at, CreatedAt: at, IdempotencyKey: "ADJUSTMENT"}
	entry := f.Entry(1, "adjustment", "1")
	entry.AdjustmentID = &adjustment.ID
	input := service.BalanceInput{Entitlement: &current, Entries: []entity.LeaveLedgerEntry{entry}, Adjustments: []entity.LeaveAdjustment{adjustment}, CalculatedAt: at}
	if _, err := service.RebuildBalance(input); err != nil {
		t.Fatal(err)
	}
	duplicate := f.Entry(2, "adjustment", "1")
	duplicate.AdjustmentID = &adjustment.ID
	input.Entries = append(input.Entries, duplicate)
	if _, err := service.RebuildBalance(input); err == nil {
		t.Fatal("same approved adjustment posted twice")
	}
	input.Entries = input.Entries[:1]
	input.Adjustments[0].ApprovedByPrincipalID = &input.Adjustments[0].RequestedByPrincipalID
	if _, err := service.RebuildBalance(input); err == nil {
		t.Fatal("self-approved adjustment posted")
	}
}
