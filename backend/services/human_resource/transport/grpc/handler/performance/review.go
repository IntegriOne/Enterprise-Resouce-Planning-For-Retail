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

func review(value *entity.PerformanceReview) (*pb.PerformanceReview, error) {
	assessment, err := response.JSONValue(value.CompetencyAssessment)
	if err != nil {
		return nil, err
	}
	return &pb.PerformanceReview{Id: value.ID, WorkerId: value.WorkerID, PerformanceCycleId: value.PerformanceCycleID, ReviewerPrincipalId: value.ReviewerPrincipalID, ReviewerRole: value.ReviewerRole, ReviewVersion: value.ReviewVersion, OverallRatingCode: value.OverallRatingCode, CompetencyAssessment: assessment, Narrative: value.Narrative, DevelopmentPlanId: value.DevelopmentPlanID, SubmittedAt: timestamp(value.SubmittedAt), FinalizedAt: timestamp(value.FinalizedAt), VisibilityClass: value.VisibilityClass, ContentDigest: value.ContentDigest}, nil
}
func reviewResponse(value *entity.PerformanceReview, err error) (*pb.PerformanceReviewResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	item, err := review(value)
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.PerformanceReviewResponse{Item: item}, nil
}
func reviews(page operation.Page[entity.PerformanceReview], err error) (*pb.PerformanceReviewsResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	items := make([]*pb.PerformanceReview, 0, len(page.Items))
	for i := range page.Items {
		item, err := review(&page.Items[i])
		if err != nil {
			return nil, response.Error(err)
		}
		items = append(items, item)
	}
	return &pb.PerformanceReviewsResponse{Items: items, Page: pageInfo(page)}, nil
}
func (h *Handler) CreatePerformanceReview(ctx context.Context, req *pb.CreatePerformanceReviewRequest) (*pb.PerformanceReviewResponse, error) {
	if req == nil || req.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	in := req.Input
	assessment, err := response.JSON(in.CompetencyAssessment)
	if err != nil {
		return nil, response.Error(err)
	}
	submitted, err := response.OptionalTimestamp(in.SubmittedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	finalized, err := response.OptionalTimestamp(in.FinalizedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	return reviewResponse(h.service.CreatePerformanceReview(ctx, service.CreatePerformanceReviewInput{WorkerID: in.GetWorkerId(), PerformanceCycleID: in.GetPerformanceCycleId(), ReviewerPrincipalID: in.GetReviewerPrincipalId(), ReviewerRole: in.GetReviewerRole(), ReviewVersion: in.ReviewVersion, OverallRatingCode: in.OverallRatingCode, CompetencyAssessment: assessment, Narrative: in.Narrative, DevelopmentPlanID: in.DevelopmentPlanId, SubmittedAt: submitted, FinalizedAt: finalized, VisibilityClass: in.VisibilityClass, ContentDigest: in.ContentDigest}))
}
func (h *Handler) UpdatePerformanceReview(ctx context.Context, req *pb.UpdatePerformanceReviewRequest) (*pb.PerformanceReviewResponse, error) {
	if req == nil || req.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	m, err := response.ValidateMask(req.UpdateMask, "worker_id", "performance_cycle_id", "reviewer_principal_id", "reviewer_role", "review_version", "overall_rating_code", "competency_assessment", "narrative", "development_plan_id", "submitted_at", "finalized_at", "visibility_class", "content_digest")
	if err != nil {
		return nil, response.Error(err)
	}
	in := req.Input
	assessment, err := jsonField(m["competency_assessment"], in.CompetencyAssessment)
	if err != nil {
		return nil, response.Error(err)
	}
	submitted, err := timeField(m["submitted_at"], in.SubmittedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	finalized, err := timeField(m["finalized_at"], in.FinalizedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	return reviewResponse(h.service.UpdatePerformanceReview(ctx, req.Id, service.UpdatePerformanceReviewInput{WorkerID: field(m["worker_id"], in.WorkerId), PerformanceCycleID: field(m["performance_cycle_id"], in.PerformanceCycleId), ReviewerPrincipalID: field(m["reviewer_principal_id"], in.ReviewerPrincipalId), ReviewerRole: field(m["reviewer_role"], in.ReviewerRole), ReviewVersion: field(m["review_version"], in.ReviewVersion), OverallRatingCode: field(m["overall_rating_code"], in.OverallRatingCode), CompetencyAssessment: assessment, Narrative: field(m["narrative"], in.Narrative), DevelopmentPlanID: field(m["development_plan_id"], in.DevelopmentPlanId), SubmittedAt: submitted, FinalizedAt: finalized, VisibilityClass: field(m["visibility_class"], in.VisibilityClass), ContentDigest: bytesField(m["content_digest"], in.ContentDigest)}))
}
func (h *Handler) DeletePerformanceReview(ctx context.Context, req *common.IDRequest) (*common.EmptyResponse, error) {
	return remove(ctx, req, h.service.DeletePerformanceReview)
}
func (h *Handler) GetPerformanceReviewByID(ctx context.Context, req *common.IDRequest) (*pb.PerformanceReviewResponse, error) {
	if req == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return reviewResponse(h.service.GetPerformanceReviewByID(ctx, req.Id))
}
func (h *Handler) GetAllPerformanceReviews(ctx context.Context, req *common.PageRequest) (*pb.PerformanceReviewsResponse, error) {
	if req == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return reviews(h.service.GetAllPerformanceReviews(ctx, int(req.Limit), int(req.Offset)))
}
func (h *Handler) FindPerformanceReviewByWorkerID(ctx context.Context, req *common.StringFinderRequest) (*pb.PerformanceReviewsResponse, error) {
	if req == nil || req.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return reviews(h.service.FindPerformanceReviewByWorkerID(ctx, *req.Value, int(req.Limit), int(req.Offset)))
}
func (h *Handler) FindPerformanceReviewByPerformanceCycleID(ctx context.Context, req *common.StringFinderRequest) (*pb.PerformanceReviewsResponse, error) {
	if req == nil || req.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return reviews(h.service.FindPerformanceReviewByPerformanceCycleID(ctx, *req.Value, int(req.Limit), int(req.Offset)))
}
func (h *Handler) FindPerformanceReviewByReviewerPrincipalID(ctx context.Context, req *common.StringFinderRequest) (*pb.PerformanceReviewsResponse, error) {
	if req == nil || req.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return reviews(h.service.FindPerformanceReviewByReviewerPrincipalID(ctx, *req.Value, int(req.Limit), int(req.Offset)))
}
func (h *Handler) FindPerformanceReviewByOverallRatingCode(ctx context.Context, req *common.StringFinderRequest) (*pb.PerformanceReviewsResponse, error) {
	if req == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return reviews(h.service.FindPerformanceReviewByOverallRatingCode(ctx, req.Value, int(req.Limit), int(req.Offset)))
}
func (h *Handler) FindPerformanceReviewByDevelopmentPlanID(ctx context.Context, req *common.StringFinderRequest) (*pb.PerformanceReviewsResponse, error) {
	if req == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return reviews(h.service.FindPerformanceReviewByDevelopmentPlanID(ctx, req.Value, int(req.Limit), int(req.Offset)))
}
