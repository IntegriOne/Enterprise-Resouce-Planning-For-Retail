package service

import (
	"bytes"
	"context"
	"strings"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	leavepolicy "github.com/integrione/erp-retail/backend/services/human_resource/domain/service/leave"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

// PostLedgerEntry mengunci entitlement dan menyimpan entry, projection dan outbox secara atomik.
// Sumber lintas management tetap wajib diverifikasi adapter Integrity melalui kontrak owner.
func (s *Service) PostLedgerEntry(ctx context.Context, entitlementID string, in PostLedgerEntryInput) (*entity.LeaveEntitlement, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.leaveEntitlement.Command(ctx, entitlementID, "post_leave_entry", func(txCtx context.Context, current *entity.LeaveEntitlement, actor operation.Actor, now time.Time) (*entity.LeaveEntitlement, error) {
		if actor.Kind != "workload" && in.SourceManagement != "hrm" {
			return nil, hrerrors.ErrForbidden
		}
		if entity.CompareDates(in.EffectiveDate, now) > 0 {
			return nil, hrerrors.Invalid("effective_date", "future ledger posting is not allowed")
		}
		filter := []domainrepo.Filter{{Column: "entitlement_id", Values: []any{entitlementID}}}
		expected, err := s.dependencies.LeaveLedgerEntry.CountWhere(txCtx, filter)
		if err != nil {
			return nil, err
		}
		entries, err := operation.CollectComplete(txCtx, expected, func(limit, offset int) (operation.Page[entity.LeaveLedgerEntry], error) {
			return s.leaveLedgerEntry.List(txCtx, filter, limit, offset)
		})
		if err != nil {
			return nil, err
		}
		for _, existing := range entries {
			if existing.SourceManagement == in.SourceManagement && existing.SourceEventID == in.SourceEventID && existing.SourceLineKey == in.SourceLineKey {
				if !bytes.Equal(existing.SourceDigest, in.SourceDigest) {
					return nil, hrerrors.ErrConflict
				}
				return nil, hrerrors.ErrAlreadyExists
			}
		}
		id, err := s.leaveLedgerEntry.NewID()
		if err != nil {
			return nil, err
		}
		entry := entity.LeaveLedgerEntry{
			ID: id, EntitlementID: entitlementID, EntrySequence: expected + 1,
			EntryType: in.EntryType, QuantityDelta: in.QuantityDelta, EffectiveDate: in.EffectiveDate,
			AdjustmentID: in.AdjustmentID, ReversesEntryID: in.ReversesEntryID,
			SourceManagement: in.SourceManagement, SourceEventID: in.SourceEventID,
			SourceLineKey: in.SourceLineKey, SourceDigest: append([]byte(nil), in.SourceDigest...),
			ReasonCode: in.ReasonCode, ActorPrincipalID: actor.PrincipalID, RecordedAt: now,
		}
		if err := entry.Validate(); err != nil {
			return nil, err
		}
		entries = append(entries, entry)
		adjustments := make([]entity.LeaveAdjustment, 0)
		seen := make(map[string]bool)
		for _, item := range entries {
			if item.AdjustmentID == nil {
				continue
			}
			key := strings.ToLower(*item.AdjustmentID)
			if seen[key] {
				continue
			}
			adjustment, err := s.leaveAdjustment.Get(txCtx, key)
			if err != nil {
				return nil, err
			}
			adjustments = append(adjustments, *adjustment)
			seen[key] = true
		}
		projection, err := leavepolicy.RebuildBalance(leavepolicy.BalanceInput{
			Entitlement: current, Entries: entries, Adjustments: adjustments, CalculatedAt: now,
		})
		if err != nil {
			return nil, err
		}
		if err := s.leaveLedgerEntry.CheckIntegrity(txCtx, nil, &entry); err != nil {
			return nil, err
		}
		if err := s.dependencies.LeaveLedgerEntry.Insert(txCtx, &entry); err != nil {
			return nil, err
		}
		return projection, nil
	})
}
