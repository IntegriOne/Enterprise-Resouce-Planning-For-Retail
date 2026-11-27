package service

import (
	"context"
	"slices"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

// CloseLeaveEntitlement menutup periode bersaldo nol tanpa membuat entri ledger.
func (s *Service) CloseLeaveEntitlement(ctx context.Context, id string) (*entity.LeaveEntitlement, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.leaveEntitlement.Command(ctx, id, "close", func(txCtx context.Context, entitlement *entity.LeaveEntitlement, actor operation.Actor, now time.Time) (*entity.LeaveEntitlement, error) {
		if actor.Kind != "human" {
			return nil, hrerrors.ErrForbidden
		}
		if entitlement.ClosedAt != nil || entitlement.ClosedByPrincipalID != nil {
			return nil, hrerrors.ErrWorkflowRequired
		}
		if entity.CompareDates(now.UTC(), entitlement.EntitlementPeriodEnd) <= 0 {
			return nil, hrerrors.ErrReferenceConflict
		}
		balance, err := entitlement.ClosingBalanceSnapshot.Rat()
		if err != nil {
			return nil, err
		}
		if balance.Sign() != 0 {
			return nil, hrerrors.ErrReferenceConflict
		}
		if err := s.requirePostedAdjustments(txCtx, entitlement.ID); err != nil {
			return nil, err
		}
		entitlement.ClosedAt = &now
		entitlement.ClosedByPrincipalID = &actor.PrincipalID
		return entitlement, nil
	})
}

// requirePostedAdjustments dijalankan saat baris entitlement terkunci.
func (s *Service) requirePostedAdjustments(txCtx context.Context, entitlementID string) error {
	filter := []domainrepo.Filter{{Column: "entitlement_id", Values: []any{entitlementID}}}
	count, err := s.dependencies.LeaveAdjustment.CountWhere(txCtx, filter)
	if err != nil {
		return err
	}
	adjustments, err := operation.CollectComplete(txCtx, count, func(limit, offset int) (operation.Page[entity.LeaveAdjustment], error) {
		items, err := s.dependencies.LeaveAdjustment.SelectWhere(txCtx, filter, limit, offset)
		return operation.Page[entity.LeaveAdjustment]{Items: items}, err
	})
	if err != nil {
		return err
	}
	approvedIDs := make([]any, 0, len(adjustments))
	for _, adjustment := range adjustments {
		if adjustment.ApprovedAt != nil || adjustment.ApprovedByPrincipalID != nil {
			approvedIDs = append(approvedIDs, adjustment.ID)
		}
	}
	// Indeks unik adjustment_id memastikan tiap koreksi hanya menyumbang satu entri.
	for ids := range slices.Chunk(approvedIDs, domainrepo.MaxLimit) {
		posted, err := s.dependencies.LeaveLedgerEntry.CountWhere(txCtx, []domainrepo.Filter{
			{Column: "entitlement_id", Values: []any{entitlementID}},
			{Column: "adjustment_id", Values: ids},
			{Column: "entry_type", Values: []any{"adjustment"}},
		})
		if err != nil {
			return err
		}
		if posted != int64(len(ids)) {
			return hrerrors.ErrReferenceConflict
		}
	}
	return nil
}
