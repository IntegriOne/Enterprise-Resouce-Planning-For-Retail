package recruitment

import (
	"context"

	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/recruitment/v1"
	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/recruitment"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/response"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func interviewFeedback(v *entity.InterviewFeedback) (*pb.InterviewFeedback, error) {
	scores, err := response.JSONValue(v.CompetencyScores)
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.InterviewFeedback{
		Id: v.ID, InterviewId: v.InterviewID, InterviewerPrincipalId: v.InterviewerPrincipalID, FeedbackVersion: v.FeedbackVersion,
		SubmittedAt: timestamppb.New(v.SubmittedAt), OverallRecommendation: v.OverallRecommendation, CompetencyScores: scores,
		EvidenceNotes: v.EvidenceNotes, ContentDigest: append([]byte(nil), v.ContentDigest...), VisibilityClass: v.VisibilityClass,
	}, nil
}
func interviewFeedbackResponse(v *entity.InterviewFeedback, err error) (*pb.InterviewFeedbackResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	item, err := interviewFeedback(v)
	if err != nil {
		return nil, err
	}
	return &pb.InterviewFeedbackResponse{Item: item}, nil
}
func interviewFeedbacks(page operation.Page[entity.InterviewFeedback], err error) (*pb.InterviewFeedbacksResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	items := make([]*pb.InterviewFeedback, 0, len(page.Items))
	for i := range page.Items {
		item, err := interviewFeedback(&page.Items[i])
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return &pb.InterviewFeedbacksResponse{Items: items, Page: pageInfo(page)}, nil
}
func (h *Handler) CreateInterviewFeedback(ctx context.Context, r *pb.CreateInterviewFeedbackRequest) (*pb.InterviewFeedbackResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	in := r.Input
	scores, err := response.JSON(in.CompetencyScores)
	if err != nil {
		return nil, response.Error(err)
	}
	return interviewFeedbackResponse(h.service.CreateInterviewFeedback(ctx, service.CreateInterviewFeedbackInput{
		InterviewID: in.GetInterviewId(), InterviewerPrincipalID: in.GetInterviewerPrincipalId(), FeedbackVersion: in.FeedbackVersion,
		OverallRecommendation: in.GetOverallRecommendation(), CompetencyScores: scores, EvidenceNotes: in.EvidenceNotes,
		ContentDigest: append([]byte(nil), in.ContentDigest...), VisibilityClass: in.VisibilityClass,
	}))
}
func (h *Handler) UpdateInterviewFeedback(ctx context.Context, r *pb.UpdateInterviewFeedbackRequest) (*pb.InterviewFeedbackResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	m, err := response.ValidateMask(r.UpdateMask, "interview_id", "interviewer_principal_id", "feedback_version", "overall_recommendation", "competency_scores", "evidence_notes", "content_digest", "visibility_class")
	if err != nil {
		return nil, response.Error(err)
	}
	in := r.Input
	scores, err := jsonField(m["competency_scores"], in.CompetencyScores)
	if err != nil {
		return nil, response.Error(err)
	}
	return interviewFeedbackResponse(h.service.UpdateInterviewFeedback(ctx, r.Id, service.UpdateInterviewFeedbackInput{
		InterviewID: field(m["interview_id"], in.InterviewId), InterviewerPrincipalID: field(m["interviewer_principal_id"], in.InterviewerPrincipalId),
		FeedbackVersion: field(m["feedback_version"], in.FeedbackVersion), OverallRecommendation: field(m["overall_recommendation"], in.OverallRecommendation),
		CompetencyScores: scores, EvidenceNotes: field(m["evidence_notes"], in.EvidenceNotes), ContentDigest: bytesField(m["content_digest"], in.ContentDigest), VisibilityClass: field(m["visibility_class"], in.VisibilityClass),
	}))
}
func (h *Handler) DeleteInterviewFeedback(ctx context.Context, r *common.IDRequest) (*common.EmptyResponse, error) {
	return deleteRecord(ctx, r, h.service.DeleteInterviewFeedback)
}
func (h *Handler) GetInterviewFeedbackByID(ctx context.Context, r *common.IDRequest) (*pb.InterviewFeedbackResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return interviewFeedbackResponse(h.service.GetInterviewFeedbackByID(ctx, r.Id))
}
func (h *Handler) GetAllInterviewFeedbacks(ctx context.Context, r *common.PageRequest) (*pb.InterviewFeedbacksResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return interviewFeedbacks(h.service.GetAllInterviewFeedback(ctx, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindInterviewFeedbackByInterviewID(ctx context.Context, r *common.StringFinderRequest) (*pb.InterviewFeedbacksResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return interviewFeedbacks(h.service.FindInterviewFeedbackByInterviewID(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindInterviewFeedbackByInterviewerPrincipalID(ctx context.Context, r *common.StringFinderRequest) (*pb.InterviewFeedbacksResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return interviewFeedbacks(h.service.FindInterviewFeedbackByInterviewerPrincipalID(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
