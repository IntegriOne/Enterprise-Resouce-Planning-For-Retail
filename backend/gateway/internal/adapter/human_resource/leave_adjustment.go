package human_resource

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	leave "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/leave/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const leaveAdjustmentReadPermission = "human_resource.leave_adjustments.read"

// LeaveAdjustment membaca detail dan status keputusan satu koreksi saldo.
func (a *Adapter) LeaveAdjustment(ctx context.Context, rawID string) (*model.LeaveAdjustment, error) {
	rpcCtx, err := a.authorizeGlobalEmployee(ctx, leaveAdjustmentReadPermission)
	if err != nil {
		return nil, err
	}
	id, err := parseNodeID(rawID, "leave-adjustment")
	if err != nil || id == uuid.Nil {
		return nil, middleware.ErrInvalidArgument("invalid leave adjustment ID", nil)
	}
	if a.leave == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	response, err := a.leave.GetLeaveAdjustmentByID(rpcCtx, &common.IDRequest{Id: id.String()})
	if err != nil {
		return nil, mapReadError(err, "leave adjustment")
	}
	row := response.GetItem()
	if row.GetId() != id.String() {
		return nil, middleware.ErrUnavailable("invalid human resource leave adjustment response", nil)
	}
	return toModelLeaveAdjustment(row)
}

// LeaveAdjustments membaca pengajuan koreksi untuk satu hak cuti.
func (a *Adapter) LeaveAdjustments(ctx context.Context, rawEntitlementID, idempotencyKey *string, limit, offset int) (*Page[model.LeaveAdjustment], error) {
	rpcCtx, err := a.authorizeGlobalEmployee(ctx, leaveAdjustmentReadPermission)
	if err != nil {
		return nil, err
	}
	if (rawEntitlementID == nil) == (idempotencyKey == nil) {
		return nil, middleware.ErrInvalidArgument("use exactly one leave adjustment filter", nil)
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid leave adjustment pagination", nil)
	}
	if a.leave == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	request := &common.StringFinderRequest{Limit: int32(limit), Offset: int32(offset)}
	var response *leave.LeaveAdjustmentsResponse
	if rawEntitlementID != nil {
		id, parseErr := parseNodeID(*rawEntitlementID, "leave-entitlement")
		if parseErr != nil || id == uuid.Nil {
			return nil, middleware.ErrInvalidArgument("invalid leave entitlement ID", nil)
		}
		entitlementID := id.String()
		request.Value = &entitlementID
		response, err = a.leave.FindLeaveAdjustmentByEntitlementID(rpcCtx, request)
	} else {
		if !validIdempotencyKey(*idempotencyKey) {
			return nil, middleware.ErrInvalidArgument("invalid leave adjustment idempotency key", nil)
		}
		request.Value = idempotencyKey
		response, err = a.leave.FindLeaveAdjustmentByIdempotencyKey(rpcCtx, request)
	}
	if err != nil {
		return nil, mapReadError(err, "leave adjustment")
	}
	page, rows := response.GetPage(), response.GetItems()
	if !validPage(page, limit, offset, len(rows)) {
		return nil, middleware.ErrUnavailable("invalid human resource leave adjustment page", nil)
	}
	result := &Page[model.LeaveAdjustment]{
		Items: make([]*model.LeaveAdjustment, 0, len(rows)), TotalCount: int(page.GetTotal()), Limit: limit, Offset: offset,
	}
	for _, row := range rows {
		if (rawEntitlementID != nil && row.GetEntitlementId() != request.GetValue()) ||
			(idempotencyKey != nil && row.GetIdempotencyKey() != request.GetValue()) {
			return nil, middleware.ErrUnavailable("invalid human resource leave adjustment match", nil)
		}
		item, err := toModelLeaveAdjustment(row)
		if err != nil {
			return nil, err
		}
		result.Items = append(result.Items, item)
	}
	return result, nil
}

func toModelLeaveAdjustment(row *leave.LeaveAdjustment) (*model.LeaveAdjustment, error) {
	for _, raw := range []string{row.GetId(), row.GetEntitlementId()} {
		if id, err := uuid.Parse(raw); err != nil || id == uuid.Nil {
			return nil, middleware.ErrUnavailable("invalid human resource leave adjustment reference", nil)
		}
	}
	item := &model.LeaveAdjustment{
		ID: nodeID("leave-adjustment", row.GetId()), RawID: row.GetId(), EntitlementID: row.GetEntitlementId(), QuantityDelta: row.GetQuantityDelta(),
		ReasonCode: row.GetReasonCode(), RejectionReasonCode: row.RejectionReasonCode,
	}
	for _, field := range []struct {
		source *timestamppb.Timestamp
		target **time.Time
	}{
		{row.GetApprovedAt(), &item.ApprovedAt},
		{row.GetRejectedAt(), &item.RejectedAt},
	} {
		if field.source == nil {
			continue
		}
		if err := field.source.CheckValid(); err != nil {
			return nil, middleware.ErrUnavailable("invalid human resource leave adjustment timestamp", nil)
		}
		at := field.source.AsTime()
		*field.target = &at
	}
	return item, nil
}

// ApproveLeaveAdjustment meminta persetujuan koreksi saldo dari HRM.
func (a *Adapter) ApproveLeaveAdjustment(ctx context.Context, id string) (*model.LeaveAdjustment, error) {
	return a.decideLeaveAdjustment(ctx, id, nil)
}

// RejectLeaveAdjustment meminta penolakan koreksi saldo beserta kode alasan.
func (a *Adapter) RejectLeaveAdjustment(ctx context.Context, id, reason string) (*model.LeaveAdjustment, error) {
	return a.decideLeaveAdjustment(ctx, id, &reason)
}

func (a *Adapter) decideLeaveAdjustment(ctx context.Context, rawID string, rejection *string) (*model.LeaveAdjustment, error) {
	permission := "human_resource.leave_adjustments.approve"
	if rejection != nil {
		permission = "human_resource.leave_adjustments.reject"
	}
	rpcCtx, err := a.authorizeSensitiveEmployee(ctx, permission)
	if err != nil {
		return nil, err
	}
	id, err := parseNodeID(rawID, "leave-adjustment")
	if err != nil || id == uuid.Nil {
		return nil, middleware.ErrInvalidArgument("invalid leave adjustment ID", nil)
	}
	if a.leave == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	var response *leave.LeaveAdjustmentResponse
	if rejection == nil {
		response, err = a.leave.ApproveLeaveAdjustment(rpcCtx, &common.IDRequest{Id: id.String()})
	} else {
		reasonCode := strings.TrimSpace(*rejection)
		if reasonCode == "" || len(reasonCode) > 80 {
			return nil, middleware.ErrInvalidArgument("invalid rejection reason code", nil)
		}
		response, err = a.leave.RejectLeaveAdjustment(rpcCtx, &leave.RejectLeaveAdjustmentRequest{Id: id.String(), ReasonCode: reasonCode})
	}
	if status.Code(err) == codes.FailedPrecondition {
		return nil, middleware.ErrConflict("leave adjustment does not meet workflow requirements", err)
	}
	if err != nil {
		return nil, mapReadError(err, "leave adjustment")
	}
	row := response.GetItem()
	if row.GetId() != id.String() ||
		(rejection == nil && (row.GetApprovedAt() == nil || row.GetRejectedAt() != nil)) ||
		(rejection != nil && (row.GetRejectedAt() == nil || row.GetApprovedAt() != nil || row.GetRejectionReasonCode() != strings.TrimSpace(*rejection))) {
		return nil, middleware.ErrUnavailable("invalid human resource leave adjustment decision", nil)
	}
	return toModelLeaveAdjustment(row)
}

// CreateLeaveAdjustment mencatat draf koreksi saldo di HRM.
func (a *Adapter) CreateLeaveAdjustment(ctx context.Context, input model.CreateLeaveAdjustmentInput) (*model.LeaveAdjustment, error) {
	rpcCtx, err := a.authorizeGlobalEmployee(ctx, "human_resource.leave_adjustments.create")
	if err != nil {
		return nil, err
	}
	entitlement, err := parseNodeID(input.EntitlementID, "leave-entitlement")
	if err != nil || entitlement == uuid.Nil {
		return nil, middleware.ErrInvalidArgument("invalid leave entitlement ID", nil)
	}
	if !validIdempotencyKey(input.IdempotencyKey) {
		return nil, middleware.ErrInvalidArgument("invalid leave adjustment idempotency key", nil)
	}
	if a.leave == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	entitlementID := entitlement.String()
	response, err := a.leave.CreateLeaveAdjustment(rpcCtx, &leave.CreateLeaveAdjustmentRequest{
		Input: &leave.LeaveAdjustmentInput{
			EntitlementId: &entitlementID, QuantityDelta: &input.QuantityDelta,
			ReasonCode: &input.ReasonCode, Reason: &input.Reason, IdempotencyKey: &input.IdempotencyKey,
		},
	})
	switch status.Code(err) {
	case codes.AlreadyExists:
		return nil, middleware.ErrConflict("leave adjustment already exists", err)
	case codes.FailedPrecondition:
		return nil, middleware.ErrConflict("leave adjustment does not meet workflow requirements", err)
	}
	if err != nil {
		return nil, mapReadError(err, "leave adjustment")
	}
	row := response.GetItem()
	if row.GetEntitlementId() != entitlementID || row.GetApprovedAt() != nil || row.GetRejectedAt() != nil {
		return nil, middleware.ErrUnavailable("invalid human resource leave adjustment draft", nil)
	}
	return toModelLeaveAdjustment(row)
}
