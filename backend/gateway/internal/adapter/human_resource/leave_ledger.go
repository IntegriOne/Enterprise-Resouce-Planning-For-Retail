package human_resource

import (
	"context"
	"strconv"

	"github.com/google/uuid"
	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	leave "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/leave/v1"
)

const leaveLedgerEntryReadPermission = "human_resource.leave_ledger_entries.read"

// LeaveLedgerEntry membaca detail satu entri pembukuan saldo cuti.
func (a *Adapter) LeaveLedgerEntry(ctx context.Context, rawID string) (*model.LeaveLedgerEntry, error) {
	rpcCtx, err := a.authorizeGlobalEmployee(ctx, leaveLedgerEntryReadPermission)
	if err != nil {
		return nil, err
	}
	id, err := parseNodeID(rawID, "leave-ledger-entry")
	if err != nil || id == uuid.Nil {
		return nil, middleware.ErrInvalidArgument("invalid leave ledger entry ID", nil)
	}
	if a.leave == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	response, err := a.leave.GetLeaveLedgerEntryByID(rpcCtx, &common.IDRequest{Id: id.String()})
	if err != nil {
		return nil, mapReadError(err, "leave ledger entry")
	}
	row := response.GetItem()
	if row.GetId() != id.String() {
		return nil, middleware.ErrUnavailable("invalid human resource leave ledger entry response", nil)
	}
	return toModelLeaveLedgerEntry(row)
}

// LeaveLedgerEntries membaca riwayat perubahan untuk satu hak cuti.
func (a *Adapter) LeaveLedgerEntries(ctx context.Context, rawEntitlementID, rawAdjustmentID *string, limit, offset int) (*Page[model.LeaveLedgerEntry], error) {
	rpcCtx, err := a.authorizeGlobalEmployee(ctx, leaveLedgerEntryReadPermission)
	if err != nil {
		return nil, err
	}
	if (rawEntitlementID == nil) == (rawAdjustmentID == nil) {
		return nil, middleware.ErrInvalidArgument("use exactly one leave ledger filter", nil)
	}
	rawID, resource := rawEntitlementID, "leave-entitlement"
	if rawAdjustmentID != nil {
		rawID, resource = rawAdjustmentID, "leave-adjustment"
	}
	id, err := parseNodeID(*rawID, resource)
	if err != nil || id == uuid.Nil {
		return nil, middleware.ErrInvalidArgument("invalid leave ledger filter ID", nil)
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid leave ledger entry pagination", nil)
	}
	if a.leave == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	filterID := id.String()
	request := &common.StringFinderRequest{Value: &filterID, Limit: int32(limit), Offset: int32(offset)}
	var response *leave.LeaveLedgerEntriesResponse
	if rawAdjustmentID == nil {
		response, err = a.leave.FindLeaveLedgerEntryByEntitlementID(rpcCtx, request)
	} else {
		response, err = a.leave.FindLeaveLedgerEntryByAdjustmentID(rpcCtx, request)
	}
	if err != nil {
		return nil, mapReadError(err, "leave ledger entry")
	}
	page, rows := response.GetPage(), response.GetItems()
	if !validPage(page, limit, offset, len(rows)) {
		return nil, middleware.ErrUnavailable("invalid human resource leave ledger entry page", nil)
	}
	result := &Page[model.LeaveLedgerEntry]{
		Items: make([]*model.LeaveLedgerEntry, 0, len(rows)), TotalCount: int(page.GetTotal()), Limit: limit, Offset: offset,
	}
	for _, row := range rows {
		if (rawEntitlementID != nil && row.GetEntitlementId() != filterID) ||
			(rawAdjustmentID != nil && row.GetAdjustmentId() != filterID) {
			return nil, middleware.ErrUnavailable("invalid human resource leave ledger entry match", nil)
		}
		item, err := toModelLeaveLedgerEntry(row)
		if err != nil {
			return nil, err
		}
		result.Items = append(result.Items, item)
	}
	return result, nil
}

func toModelLeaveLedgerEntry(row *leave.LeaveLedgerEntry) (*model.LeaveLedgerEntry, error) {
	for _, raw := range []string{row.GetId(), row.GetEntitlementId()} {
		if id, err := uuid.Parse(raw); err != nil || id == uuid.Nil {
			return nil, middleware.ErrUnavailable("invalid human resource leave ledger entry reference", nil)
		}
	}
	return &model.LeaveLedgerEntry{
		ID: nodeID("leave-ledger-entry", row.GetId()), RawID: row.GetId(), EntitlementID: row.GetEntitlementId(), EntrySequence: strconv.FormatInt(row.GetEntrySequence(), 10),
		EntryType: row.GetEntryType(), QuantityDelta: row.GetQuantityDelta(), EffectiveDate: row.GetEffectiveDate(),
		AdjustmentID: row.AdjustmentId, ReversesEntryID: row.ReversesEntryId, ReasonCode: row.GetReasonCode(),
	}, nil
}
