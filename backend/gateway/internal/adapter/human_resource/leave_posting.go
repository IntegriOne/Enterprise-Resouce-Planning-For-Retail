package human_resource

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	leave "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/leave/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// PostLeaveAdjustment membukukan koreksi dengan identitas sumber HRM yang stabil.
func (a *Adapter) PostLeaveAdjustment(ctx context.Context, input model.PostLeaveAdjustmentInput) (*model.LeaveEntitlement, error) {
	rpcCtx, err := a.authorizeSensitiveEmployee(ctx, "human_resource.leave_entitlements.post")
	if err != nil {
		return nil, err
	}
	entitlement, err := parseNodeID(input.EntitlementID, "leave-entitlement")
	if err != nil || entitlement == uuid.Nil {
		return nil, middleware.ErrInvalidArgument("invalid leave entitlement ID", nil)
	}
	adjustment, err := parseNodeID(input.AdjustmentID, "leave-adjustment")
	if err != nil || adjustment == uuid.Nil {
		return nil, middleware.ErrInvalidArgument("invalid leave adjustment ID", nil)
	}
	date, err := time.Parse(time.DateOnly, input.EffectiveDate)
	if err != nil {
		return nil, middleware.ErrInvalidArgument("invalid leave posting effective date", nil)
	}
	if a.leave == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	entitlementID, adjustmentID := entitlement.String(), adjustment.String()
	effectiveDate, reason := date.Format(time.DateOnly), strings.TrimSpace(input.ReasonCode)
	entryType, source, line := "adjustment", "hrm", "adjustment-posting-v1"
	payload, err := json.Marshal([]string{line, entitlementID, adjustmentID, input.QuantityDelta, effectiveDate, reason})
	if err != nil {
		return nil, middleware.ErrUnavailable("cannot encode leave posting digest", nil)
	}
	digest := sha256.Sum256(payload)
	response, err := a.leave.PostLedgerEntry(rpcCtx, &leave.PostLedgerEntryRequest{
		EntitlementId: entitlementID,
		Input: &leave.PostLedgerEntryInput{
			EntryType: &entryType, AdjustmentId: &adjustmentID, QuantityDelta: &input.QuantityDelta,
			EffectiveDate: &effectiveDate, ReasonCode: &reason, SourceManagement: &source,
			SourceEventId: &adjustmentID, SourceLineKey: &line, SourceDigest: digest[:],
		},
	})
	switch status.Code(err) {
	case codes.AlreadyExists, codes.FailedPrecondition:
		return nil, middleware.ErrConflict("leave posting conflicts with current entitlement state", err)
	}
	if err != nil {
		return nil, mapReadError(err, "leave ledger entry")
	}
	row := response.GetItem()
	if row.GetId() != entitlementID {
		return nil, middleware.ErrUnavailable("invalid human resource leave posting response", nil)
	}
	return toModelLeaveEntitlement(row)
}
