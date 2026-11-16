// Package service merekonstruksi proyeksi saldo dari ledger cuti yang immutable.
package service

import (
	"math"
	"sort"
	"strings"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
)

// RebuildBalance menghitung salinan saldo dari seluruh ledger, tanpa mengubah input atau repository.
// Pemanggil memuat seluruh halaman ledger di dalam transaksi dan memverifikasi scope lebih dahulu.
func RebuildBalance(input BalanceInput) (*entity.LeaveEntitlement, error) {
	current := input.Entitlement
	if err := current.Validate(); err != nil {
		return nil, err
	}
	if input.CalculatedAt.IsZero() || current.Revision == math.MaxInt64 {
		return nil, hrerrors.Invalid("revision", "valid calculation time and revision required")
	}
	entries := append([]entity.LeaveLedgerEntry(nil), input.Entries...)
	sort.Slice(entries, func(i, j int) bool { return entries[i].EntrySequence < entries[j].EntrySequence })
	adjustments := map[string]entity.LeaveAdjustment{}
	for _, adjustment := range input.Adjustments {
		if err := adjustment.Validate(); err != nil {
			return nil, err
		}
		key := strings.ToLower(adjustment.ID)
		if _, exists := adjustments[key]; exists {
			return nil, hrerrors.Invalid("adjustments", "duplicate adjustment")
		}
		adjustments[key] = adjustment
	}
	seen := map[string]entity.LeaveLedgerEntry{}
	reversed := map[string]bool{}
	sources := map[[3]string]bool{}
	buckets := map[string][]entity.Decimal{}
	openingSeen := false
	postedAdjustments := map[string]bool{}
	for index, entry := range entries {
		if err := entry.Validate(); err != nil {
			return nil, err
		}
		if entity.CompareDates(entry.EffectiveDate, current.EntitlementPeriodStart) < 0 ||
			entity.CompareDates(entry.EffectiveDate, current.EntitlementPeriodEnd) > 0 {
			return nil, hrerrors.Invalid("effective_date", "ledger entry must fall within the entitlement period")
		}
		key := strings.ToLower(entry.ID)
		_, duplicate := seen[key]
		source := [3]string{entry.SourceManagement, entry.SourceEventID, entry.SourceLineKey}
		if duplicate || sources[source] || entry.EntrySequence != int64(index)+1 || !strings.EqualFold(entry.EntitlementID, current.ID) {
			return nil, hrerrors.Invalid("entries", "ledger must be complete, contiguous and unique for one entitlement")
		}
		sources[source] = true
		kind := entry.EntryType
		switch kind {
		case "reversal":
			originalKey := strings.ToLower(*entry.ReversesEntryID)
			original, ok := seen[originalKey]
			if !ok || reversed[originalKey] || original.EntryType == "reversal" {
				return nil, hrerrors.Invalid("reverses_entry_id", "must reference an unreversed earlier entry")
			}
			total, err := entity.DecimalSum(4, original.QuantityDelta, entry.QuantityDelta)
			if err != nil {
				return nil, err
			}
			number, err := total.Rat()
			if err != nil {
				return nil, err
			}
			if number.Sign() != 0 {
				return nil, hrerrors.Invalid("quantity_delta", "reversal must exactly negate the original entry")
			}
			reversed[originalKey] = true
			kind = original.EntryType
		case "opening":
			if openingSeen || index != 0 {
				return nil, hrerrors.Invalid("entry_type", "opening entry must be the first and unique entry")
			}
			openingSeen = true
		case "adjustment":
			adjustmentKey := strings.ToLower(*entry.AdjustmentID)
			adjustment, ok := adjustments[adjustmentKey]
			if !ok || postedAdjustments[adjustmentKey] || !strings.EqualFold(adjustment.EntitlementID, current.ID) || adjustment.ApprovedAt == nil || adjustment.ApprovedByPrincipalID == nil || strings.EqualFold(*adjustment.ApprovedByPrincipalID, adjustment.RequestedByPrincipalID) {
				return nil, hrerrors.Invalid("adjustment_id", "adjustment requires matching independent approval")
			}
			postedAdjustments[adjustmentKey] = true
			total, err := entity.DecimalNegate(adjustment.QuantityDelta)
			if err != nil {
				return nil, err
			}
			difference, err := entity.DecimalSum(4, total, entry.QuantityDelta)
			if err != nil {
				return nil, err
			}
			number, err := difference.Rat()
			if err != nil {
				return nil, err
			}
			if number.Sign() != 0 {
				return nil, hrerrors.Invalid("quantity_delta", "adjustment entry must equal its approved amount")
			}
		}
		buckets[kind] = append(buckets[kind], entry.QuantityDelta)
		seen[key] = entry
	}
	result := *current
	var err error
	result.OpeningBalance, err = entity.DecimalSum(4, buckets["opening"]...)
	if err != nil {
		return nil, err
	}
	result.AccruedQuantity, err = entity.DecimalSum(4, buckets["accrual"]...)
	if err != nil {
		return nil, err
	}
	result.AdjustedQuantity, err = entity.DecimalSum(4, buckets["adjustment"]...)
	if err != nil {
		return nil, err
	}
	debits := append(append([]entity.Decimal{}, buckets["usage"]...), buckets["expiry"]...)
	used, err := entity.DecimalSum(4, debits...)
	if err != nil {
		return nil, err
	}
	result.UsedQuantitySnapshot, err = entity.DecimalNegate(used)
	if err != nil {
		return nil, err
	}
	result.ClosingBalanceSnapshot, err = entity.DecimalSum(4, result.OpeningBalance, result.AccruedQuantity, result.AdjustedQuantity, used)
	if err != nil {
		return nil, err
	}
	result.Revision++
	result.CalculatedAt = input.CalculatedAt
	if err := result.Validate(); err != nil {
		return nil, err
	}
	return &result, nil
}
