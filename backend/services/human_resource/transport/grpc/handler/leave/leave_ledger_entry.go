package leave

import (
	"context"

	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/leave/v1"
	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/leave"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/response"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func leaveLedgerEntry(v *entity.LeaveLedgerEntry) *pb.LeaveLedgerEntry {
	return &pb.LeaveLedgerEntry{
		Id: v.ID, EntitlementId: v.EntitlementID, EntrySequence: v.EntrySequence, EntryType: v.EntryType, QuantityDelta: string(v.QuantityDelta),
		EffectiveDate: v.EffectiveDate.Format("2006-01-02"), AdjustmentId: v.AdjustmentID, ReversesEntryId: v.ReversesEntryID,
		SourceManagement: v.SourceManagement, SourceEventId: v.SourceEventID, SourceLineKey: v.SourceLineKey, SourceDigest: append([]byte(nil), v.SourceDigest...),
		ReasonCode: v.ReasonCode, ActorPrincipalId: v.ActorPrincipalID, RecordedAt: timestamppb.New(v.RecordedAt),
	}
}
func leaveLedgerEntries(page operation.Page[entity.LeaveLedgerEntry], err error) (*pb.LeaveLedgerEntriesResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	items := make([]*pb.LeaveLedgerEntry, 0, len(page.Items))
	for i := range page.Items {
		items = append(items, leaveLedgerEntry(&page.Items[i]))
	}
	return &pb.LeaveLedgerEntriesResponse{Items: items, Page: &common.PageInfo{Total: page.Total, Limit: int32(page.Limit), Offset: int32(page.Offset)}}, nil
}
func (h *Handler) CreateLeaveLedgerEntry(ctx context.Context, r *pb.CreateLeaveLedgerEntryRequest) (*pb.LeaveLedgerEntryResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	in := r.Input
	date, err := response.Date(in.GetEffectiveDate())
	if err != nil {
		return nil, response.Error(err)
	}
	v, err := h.service.CreateLeaveLedgerEntry(ctx, service.CreateLeaveLedgerEntryInput{
		EntitlementID: in.GetEntitlementId(), EntrySequence: in.GetEntrySequence(), EntryType: in.GetEntryType(), QuantityDelta: entity.Decimal(in.GetQuantityDelta()),
		EffectiveDate: date, AdjustmentID: in.AdjustmentId, ReversesEntryID: in.ReversesEntryId, SourceManagement: in.GetSourceManagement(),
		SourceEventID: in.GetSourceEventId(), SourceLineKey: in.GetSourceLineKey(), SourceDigest: append([]byte(nil), in.SourceDigest...), ReasonCode: in.GetReasonCode(),
	})
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.LeaveLedgerEntryResponse{Item: leaveLedgerEntry(v)}, nil
}
func (h *Handler) UpdateLeaveLedgerEntry(ctx context.Context, r *pb.UpdateLeaveLedgerEntryRequest) (*pb.LeaveLedgerEntryResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	m, err := response.ValidateMask(r.UpdateMask, "entitlement_id", "entry_sequence", "entry_type", "quantity_delta", "effective_date", "adjustment_id", "reverses_entry_id", "source_management", "source_event_id", "source_line_key", "source_digest", "reason_code")
	if err != nil {
		return nil, response.Error(err)
	}
	in := r.Input
	date, err := dateField(m["effective_date"], in.EffectiveDate)
	if err != nil {
		return nil, response.Error(err)
	}
	var digest *[]byte
	if in.SourceDigest != nil {
		copied := append([]byte{}, in.SourceDigest...)
		digest = &copied
	}
	v, err := h.service.UpdateLeaveLedgerEntry(ctx, r.Id, service.UpdateLeaveLedgerEntryInput{
		EntitlementID: field(m["entitlement_id"], in.EntitlementId), EntrySequence: field(m["entry_sequence"], in.EntrySequence),
		EntryType: field(m["entry_type"], in.EntryType), QuantityDelta: field(m["quantity_delta"], decimal(in.QuantityDelta)), EffectiveDate: date,
		AdjustmentID: field(m["adjustment_id"], in.AdjustmentId), ReversesEntryID: field(m["reverses_entry_id"], in.ReversesEntryId),
		SourceManagement: field(m["source_management"], in.SourceManagement), SourceEventID: field(m["source_event_id"], in.SourceEventId),
		SourceLineKey: field(m["source_line_key"], in.SourceLineKey), SourceDigest: field(m["source_digest"], digest), ReasonCode: field(m["reason_code"], in.ReasonCode),
	})
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.LeaveLedgerEntryResponse{Item: leaveLedgerEntry(v)}, nil
}
func (h *Handler) DeleteLeaveLedgerEntry(ctx context.Context, r *common.IDRequest) (*common.EmptyResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	if err := h.service.DeleteLeaveLedgerEntry(ctx, r.Id); err != nil {
		return nil, response.Error(err)
	}
	return &common.EmptyResponse{}, nil
}
func (h *Handler) GetLeaveLedgerEntryByID(ctx context.Context, r *common.IDRequest) (*pb.LeaveLedgerEntryResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	v, err := h.service.GetLeaveLedgerEntryByID(ctx, r.Id)
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.LeaveLedgerEntryResponse{Item: leaveLedgerEntry(v)}, nil
}
func (h *Handler) GetAllLeaveLedgerEntries(ctx context.Context, r *common.PageRequest) (*pb.LeaveLedgerEntriesResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return leaveLedgerEntries(h.service.GetAllLeaveLedgerEntries(ctx, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindLeaveLedgerEntryByEntitlementID(ctx context.Context, r *common.StringFinderRequest) (*pb.LeaveLedgerEntriesResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return leaveLedgerEntries(h.service.FindLeaveLedgerEntryByEntitlementID(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindLeaveLedgerEntryByEntryType(ctx context.Context, r *common.StringFinderRequest) (*pb.LeaveLedgerEntriesResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return leaveLedgerEntries(h.service.FindLeaveLedgerEntryByEntryType(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindLeaveLedgerEntryByAdjustmentID(ctx context.Context, r *common.StringFinderRequest) (*pb.LeaveLedgerEntriesResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return leaveLedgerEntries(h.service.FindLeaveLedgerEntryByAdjustmentID(ctx, r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindLeaveLedgerEntryByReversesEntryID(ctx context.Context, r *common.StringFinderRequest) (*pb.LeaveLedgerEntriesResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return leaveLedgerEntries(h.service.FindLeaveLedgerEntryByReversesEntryID(ctx, r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindLeaveLedgerEntryBySourceManagement(ctx context.Context, r *common.StringFinderRequest) (*pb.LeaveLedgerEntriesResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return leaveLedgerEntries(h.service.FindLeaveLedgerEntryBySourceManagement(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindLeaveLedgerEntryBySourceEventID(ctx context.Context, r *common.StringFinderRequest) (*pb.LeaveLedgerEntriesResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return leaveLedgerEntries(h.service.FindLeaveLedgerEntryBySourceEventID(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindLeaveLedgerEntryByReasonCode(ctx context.Context, r *common.StringFinderRequest) (*pb.LeaveLedgerEntriesResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return leaveLedgerEntries(h.service.FindLeaveLedgerEntryByReasonCode(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindLeaveLedgerEntryByActorPrincipalID(ctx context.Context, r *common.StringFinderRequest) (*pb.LeaveLedgerEntriesResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return leaveLedgerEntries(h.service.FindLeaveLedgerEntryByActorPrincipalID(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) PostLedgerEntry(ctx context.Context, r *pb.PostLedgerEntryRequest) (*pb.LeaveEntitlementResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	in := r.Input
	date, err := response.Date(in.GetEffectiveDate())
	if err != nil {
		return nil, response.Error(err)
	}
	v, err := h.service.PostLedgerEntry(ctx, r.EntitlementId, service.PostLedgerEntryInput{
		EntryType: in.GetEntryType(), QuantityDelta: entity.Decimal(in.GetQuantityDelta()), EffectiveDate: date, AdjustmentID: in.AdjustmentId,
		ReversesEntryID: in.ReversesEntryId, SourceManagement: in.GetSourceManagement(), SourceEventID: in.GetSourceEventId(),
		SourceLineKey: in.GetSourceLineKey(), SourceDigest: append([]byte(nil), in.SourceDigest...), ReasonCode: in.GetReasonCode(),
	})
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.LeaveEntitlementResponse{Item: leaveEntitlement(v)}, nil
}
