package performance

import (
	"context"

	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/performance/v1"
	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/performance"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/response"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func goalUpdate(v *entity.PerformanceGoalUpdate) *pb.PerformanceGoalUpdate {
	return &pb.PerformanceGoalUpdate{
		Id: v.ID, GoalId: v.GoalID, ObservedAt: timestamppb.New(v.ObservedAt), ActualValue: string(v.ActualValue),
		EvidenceDocumentVersionReferenceId: v.EvidenceDocumentVersionReferenceID, EvidenceDigest: append([]byte(nil), v.EvidenceDigest...),
		RecordedByPrincipalId: v.RecordedByPrincipalID, RecordedAt: timestamppb.New(v.RecordedAt), SupersedesUpdateId: v.SupersedesUpdateID, IdempotencyKey: v.IdempotencyKey,
	}
}
func goalUpdateResponse(v *entity.PerformanceGoalUpdate, err error) (*pb.PerformanceGoalUpdateResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.PerformanceGoalUpdateResponse{Item: goalUpdate(v)}, nil
}
func goalUpdates(page operation.Page[entity.PerformanceGoalUpdate], err error) (*pb.PerformanceGoalUpdatesResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	items := make([]*pb.PerformanceGoalUpdate, 0, len(page.Items))
	for i := range page.Items {
		items = append(items, goalUpdate(&page.Items[i]))
	}
	return &pb.PerformanceGoalUpdatesResponse{Items: items, Page: pageInfo(page)}, nil
}
func (h *Handler) CreatePerformanceGoalUpdate(ctx context.Context, r *pb.CreatePerformanceGoalUpdateRequest) (*pb.PerformanceGoalUpdateResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	in := r.Input
	observed, err := response.Timestamp(in.ObservedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	return goalUpdateResponse(h.service.CreatePerformanceGoalUpdate(ctx, service.CreatePerformanceGoalUpdateInput{
		GoalID: in.GetGoalId(), ObservedAt: observed, ActualValue: entity.Decimal(in.GetActualValue()),
		EvidenceDocumentVersionReferenceID: in.GetEvidenceDocumentVersionReferenceId(), EvidenceDigest: append([]byte(nil), in.EvidenceDigest...), SupersedesUpdateID: in.SupersedesUpdateId, IdempotencyKey: in.GetIdempotencyKey(),
	}))
}
func (h *Handler) UpdatePerformanceGoalUpdate(ctx context.Context, r *pb.UpdatePerformanceGoalUpdateRequest) (*pb.PerformanceGoalUpdateResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	m, err := response.ValidateMask(r.UpdateMask, "goal_id", "observed_at", "actual_value", "evidence_document_version_reference_id", "evidence_digest", "supersedes_update_id", "idempotency_key")
	if err != nil {
		return nil, response.Error(err)
	}
	in := r.Input
	observed, err := timeField(m["observed_at"], in.ObservedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	return goalUpdateResponse(h.service.UpdatePerformanceGoalUpdate(ctx, r.Id, service.UpdatePerformanceGoalUpdateInput{
		GoalID: field(m["goal_id"], in.GoalId), ObservedAt: observed, ActualValue: decimalField(m["actual_value"], in.ActualValue),
		EvidenceDocumentVersionReferenceID: field(m["evidence_document_version_reference_id"], in.EvidenceDocumentVersionReferenceId), EvidenceDigest: bytesField(m["evidence_digest"], in.EvidenceDigest),
		SupersedesUpdateID: field(m["supersedes_update_id"], in.SupersedesUpdateId), IdempotencyKey: field(m["idempotency_key"], in.IdempotencyKey),
	}))
}
func (h *Handler) DeletePerformanceGoalUpdate(ctx context.Context, r *common.IDRequest) (*common.EmptyResponse, error) {
	return remove(ctx, r, h.service.DeletePerformanceGoalUpdate)
}
func (h *Handler) GetPerformanceGoalUpdateByID(ctx context.Context, r *common.IDRequest) (*pb.PerformanceGoalUpdateResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return goalUpdateResponse(h.service.GetPerformanceGoalUpdateByID(ctx, r.Id))
}
func (h *Handler) GetAllPerformanceGoalUpdates(ctx context.Context, r *common.PageRequest) (*pb.PerformanceGoalUpdatesResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return goalUpdates(h.service.GetAllPerformanceGoalUpdates(ctx, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindPerformanceGoalUpdateByGoalID(ctx context.Context, r *common.StringFinderRequest) (*pb.PerformanceGoalUpdatesResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return goalUpdates(h.service.FindPerformanceGoalUpdateByGoalID(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindPerformanceGoalUpdateByEvidenceDocumentVersionReferenceID(ctx context.Context, r *common.StringFinderRequest) (*pb.PerformanceGoalUpdatesResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return goalUpdates(h.service.FindPerformanceGoalUpdateByEvidenceDocumentVersionReferenceID(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindPerformanceGoalUpdateByRecordedByPrincipalID(ctx context.Context, r *common.StringFinderRequest) (*pb.PerformanceGoalUpdatesResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return goalUpdates(h.service.FindPerformanceGoalUpdateByRecordedByPrincipalID(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindPerformanceGoalUpdateBySupersedesUpdateID(ctx context.Context, r *common.StringFinderRequest) (*pb.PerformanceGoalUpdatesResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return goalUpdates(h.service.FindPerformanceGoalUpdateBySupersedesUpdateID(ctx, r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindPerformanceGoalUpdateByIdempotencyKey(ctx context.Context, r *common.StringFinderRequest) (*pb.PerformanceGoalUpdatesResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return goalUpdates(h.service.FindPerformanceGoalUpdateByIdempotencyKey(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
