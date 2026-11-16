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
)

func reviewGoalScore(v *entity.PerformanceReviewGoalScore) *pb.PerformanceReviewGoalScore {
	return &pb.PerformanceReviewGoalScore{
		Id: v.ID, PerformanceReviewId: v.PerformanceReviewID, GoalId: v.GoalID, WorkerId: v.WorkerID, PerformanceCycleId: v.PerformanceCycleID,
		Rating: int32(v.Rating), WeightPercentSnapshot: string(v.WeightPercentSnapshot), RubricDocumentVersionReferenceId: v.RubricDocumentVersionReferenceID, EvidenceDocumentVersionReferenceId: v.EvidenceDocumentVersionReferenceID,
	}
}
func reviewGoalScoreResponse(v *entity.PerformanceReviewGoalScore, err error) (*pb.PerformanceReviewGoalScoreResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.PerformanceReviewGoalScoreResponse{Item: reviewGoalScore(v)}, nil
}
func reviewGoalScores(page operation.Page[entity.PerformanceReviewGoalScore], err error) (*pb.PerformanceReviewGoalScoresResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	items := make([]*pb.PerformanceReviewGoalScore, 0, len(page.Items))
	for i := range page.Items {
		items = append(items, reviewGoalScore(&page.Items[i]))
	}
	return &pb.PerformanceReviewGoalScoresResponse{Items: items, Page: pageInfo(page)}, nil
}
func (h *Handler) CreatePerformanceReviewGoalScore(ctx context.Context, r *pb.CreatePerformanceReviewGoalScoreRequest) (*pb.PerformanceReviewGoalScoreResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	in := r.Input
	if in.GetRating() < -32768 || in.GetRating() > 32767 {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return reviewGoalScoreResponse(h.service.CreatePerformanceReviewGoalScore(ctx, service.CreatePerformanceReviewGoalScoreInput{
		PerformanceReviewID: in.GetPerformanceReviewId(), GoalID: in.GetGoalId(), WorkerID: in.GetWorkerId(), PerformanceCycleID: in.GetPerformanceCycleId(),
		Rating: int16(in.GetRating()), WeightPercentSnapshot: entity.Decimal(in.GetWeightPercentSnapshot()),
		RubricDocumentVersionReferenceID: in.GetRubricDocumentVersionReferenceId(), EvidenceDocumentVersionReferenceID: in.GetEvidenceDocumentVersionReferenceId(),
	}))
}
func (h *Handler) UpdatePerformanceReviewGoalScore(ctx context.Context, r *pb.UpdatePerformanceReviewGoalScoreRequest) (*pb.PerformanceReviewGoalScoreResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	m, err := response.ValidateMask(r.UpdateMask, "performance_review_id", "goal_id", "worker_id", "performance_cycle_id", "rating", "weight_percent_snapshot", "rubric_document_version_reference_id", "evidence_document_version_reference_id")
	if err != nil {
		return nil, response.Error(err)
	}
	in := r.Input
	var rating *int16
	if m["rating"] && in.Rating != nil {
		if *in.Rating < -32768 || *in.Rating > 32767 {
			return nil, response.Error(hrerrors.ErrInvalidInput)
		}
		converted := int16(*in.Rating)
		rating = &converted
	}
	return reviewGoalScoreResponse(h.service.UpdatePerformanceReviewGoalScore(ctx, r.Id, service.UpdatePerformanceReviewGoalScoreInput{
		PerformanceReviewID: field(m["performance_review_id"], in.PerformanceReviewId), GoalID: field(m["goal_id"], in.GoalId), WorkerID: field(m["worker_id"], in.WorkerId), PerformanceCycleID: field(m["performance_cycle_id"], in.PerformanceCycleId),
		Rating: field(m["rating"], rating), WeightPercentSnapshot: decimalField(m["weight_percent_snapshot"], in.WeightPercentSnapshot),
		RubricDocumentVersionReferenceID: field(m["rubric_document_version_reference_id"], in.RubricDocumentVersionReferenceId), EvidenceDocumentVersionReferenceID: field(m["evidence_document_version_reference_id"], in.EvidenceDocumentVersionReferenceId),
	}))
}
func (h *Handler) DeletePerformanceReviewGoalScore(ctx context.Context, r *common.IDRequest) (*common.EmptyResponse, error) {
	return remove(ctx, r, h.service.DeletePerformanceReviewGoalScore)
}
func (h *Handler) GetPerformanceReviewGoalScoreByID(ctx context.Context, r *common.IDRequest) (*pb.PerformanceReviewGoalScoreResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return reviewGoalScoreResponse(h.service.GetPerformanceReviewGoalScoreByID(ctx, r.Id))
}
func (h *Handler) GetAllPerformanceReviewGoalScores(ctx context.Context, r *common.PageRequest) (*pb.PerformanceReviewGoalScoresResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return reviewGoalScores(h.service.GetAllPerformanceReviewGoalScores(ctx, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindPerformanceReviewGoalScoreByPerformanceReviewID(ctx context.Context, r *common.StringFinderRequest) (*pb.PerformanceReviewGoalScoresResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return reviewGoalScores(h.service.FindPerformanceReviewGoalScoreByPerformanceReviewID(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindPerformanceReviewGoalScoreByGoalID(ctx context.Context, r *common.StringFinderRequest) (*pb.PerformanceReviewGoalScoresResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return reviewGoalScores(h.service.FindPerformanceReviewGoalScoreByGoalID(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindPerformanceReviewGoalScoreByWorkerID(ctx context.Context, r *common.StringFinderRequest) (*pb.PerformanceReviewGoalScoresResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return reviewGoalScores(h.service.FindPerformanceReviewGoalScoreByWorkerID(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindPerformanceReviewGoalScoreByPerformanceCycleID(ctx context.Context, r *common.StringFinderRequest) (*pb.PerformanceReviewGoalScoresResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return reviewGoalScores(h.service.FindPerformanceReviewGoalScoreByPerformanceCycleID(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindPerformanceReviewGoalScoreByRubricDocumentVersionReferenceID(ctx context.Context, r *common.StringFinderRequest) (*pb.PerformanceReviewGoalScoresResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return reviewGoalScores(h.service.FindPerformanceReviewGoalScoreByRubricDocumentVersionReferenceID(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindPerformanceReviewGoalScoreByEvidenceDocumentVersionReferenceID(ctx context.Context, r *common.StringFinderRequest) (*pb.PerformanceReviewGoalScoresResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return reviewGoalScores(h.service.FindPerformanceReviewGoalScoreByEvidenceDocumentVersionReferenceID(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
