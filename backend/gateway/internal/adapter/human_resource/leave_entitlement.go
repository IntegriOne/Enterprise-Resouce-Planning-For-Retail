package human_resource

import (
	"context"

	"github.com/google/uuid"
	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	leave "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/leave/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const leaveEntitlementReadPermission = "human_resource.leave_entitlements.read"

// LeaveEntitlement membaca satu hak cuti beserta saldo snapshot.
func (a *Adapter) LeaveEntitlement(ctx context.Context, rawID string) (*model.LeaveEntitlement, error) {
	rpcCtx, err := a.authorizeGlobalEmployee(ctx, leaveEntitlementReadPermission)
	if err != nil {
		return nil, err
	}
	id, err := parseNodeID(rawID, "leave-entitlement")
	if err != nil || id == uuid.Nil {
		return nil, middleware.ErrInvalidArgument("invalid leave entitlement ID", nil)
	}
	if a.leave == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	response, err := a.leave.GetLeaveEntitlementByID(rpcCtx, &common.IDRequest{Id: id.String()})
	if err != nil {
		return nil, mapReadError(err, "leave entitlement")
	}
	item := response.GetItem()
	if item.GetId() != id.String() {
		return nil, middleware.ErrUnavailable("invalid human resource leave entitlement response", nil)
	}
	return toModelLeaveEntitlement(item)
}

// LeaveEntitlements membaca hak cuti untuk satu pekerja.
func (a *Adapter) LeaveEntitlements(ctx context.Context, rawWorkerID string, limit, offset int) (*Page[model.LeaveEntitlement], error) {
	rpcCtx, err := a.authorizeGlobalEmployee(ctx, leaveEntitlementReadPermission)
	if err != nil {
		return nil, err
	}
	id, err := parseNodeID(rawWorkerID, "hrm-worker")
	if err != nil || id == uuid.Nil {
		return nil, middleware.ErrInvalidArgument("invalid worker ID", nil)
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid leave entitlement pagination", nil)
	}
	if a.leave == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	workerID := id.String()
	response, err := a.leave.FindLeaveEntitlementByWorkerID(rpcCtx, &common.StringFinderRequest{
		Value: &workerID, Limit: int32(limit), Offset: int32(offset),
	})
	if err != nil {
		return nil, mapReadError(err, "leave entitlement")
	}
	page, rows := response.GetPage(), response.GetItems()
	if !validPage(page, limit, offset, len(rows)) {
		return nil, middleware.ErrUnavailable("invalid human resource leave entitlement page", nil)
	}
	result := &Page[model.LeaveEntitlement]{
		Items: make([]*model.LeaveEntitlement, 0, len(rows)), TotalCount: int(page.GetTotal()), Limit: limit, Offset: offset,
	}
	for _, row := range rows {
		if row.GetWorkerId() != workerID {
			return nil, middleware.ErrUnavailable("invalid human resource leave entitlement worker", nil)
		}
		item, err := toModelLeaveEntitlement(row)
		if err != nil {
			return nil, err
		}
		result.Items = append(result.Items, item)
	}
	return result, nil
}

// CloseLeaveEntitlement meminta penutupan hak cuti setelah prasyarat HRM terpenuhi.
func (a *Adapter) CloseLeaveEntitlement(ctx context.Context, rawID string) (*model.LeaveEntitlement, error) {
	rpcCtx, err := a.authorizeSensitiveEmployee(ctx, "human_resource.leave_entitlements.close")
	if err != nil {
		return nil, err
	}
	id, err := parseNodeID(rawID, "leave-entitlement")
	if err != nil || id == uuid.Nil {
		return nil, middleware.ErrInvalidArgument("invalid leave entitlement ID", nil)
	}
	if a.leave == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	response, err := a.leave.CloseLeaveEntitlement(rpcCtx, &common.IDRequest{Id: id.String()})
	if status.Code(err) == codes.FailedPrecondition {
		return nil, middleware.ErrConflict("leave entitlement does not meet closing requirements", err)
	}
	if err != nil {
		return nil, mapReadError(err, "leave entitlement")
	}
	row := response.GetItem()
	if row.GetId() != id.String() || row.GetClosedAt() == nil {
		return nil, middleware.ErrUnavailable("invalid human resource closed leave entitlement", nil)
	}
	return toModelLeaveEntitlement(row)
}

func toModelLeaveEntitlement(row *leave.LeaveEntitlement) (*model.LeaveEntitlement, error) {
	for _, raw := range []string{row.GetId(), row.GetWorkerId(), row.GetLeavePolicyId()} {
		if id, err := uuid.Parse(raw); err != nil || id == uuid.Nil {
			return nil, middleware.ErrUnavailable("invalid human resource leave entitlement reference", nil)
		}
	}
	item := &model.LeaveEntitlement{
		ID: nodeID("leave-entitlement", row.GetId()), RawID: row.GetId(), WorkerID: row.GetWorkerId(), LeavePolicyID: row.GetLeavePolicyId(),
		LeaveTypeCode: row.GetLeaveTypeCodeSnapshot(), PeriodStart: row.GetEntitlementPeriodStart(),
		PeriodEnd: row.GetEntitlementPeriodEnd(), OpeningBalance: row.GetOpeningBalance(),
		AccruedQuantity: row.GetAccruedQuantity(), UsedQuantity: row.GetUsedQuantitySnapshot(),
		AdjustedQuantity: row.GetAdjustedQuantity(), ClosingBalance: row.GetClosingBalanceSnapshot(), UomCode: row.GetUomCode(),
	}
	if closed := row.GetClosedAt(); closed != nil {
		if err := closed.CheckValid(); err != nil {
			return nil, middleware.ErrUnavailable("invalid human resource leave entitlement closing time", nil)
		}
		closedAt := closed.AsTime()
		item.ClosedAt = &closedAt
	}
	return item, nil
}
