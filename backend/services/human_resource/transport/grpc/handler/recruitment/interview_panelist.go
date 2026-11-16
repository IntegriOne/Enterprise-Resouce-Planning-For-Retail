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
)

func interviewPanelist(v *entity.InterviewPanelist) *pb.InterviewPanelist {
	return &pb.InterviewPanelist{
		Id: v.ID, InterviewId: v.InterviewID, InterviewerPrincipalId: v.InterviewerPrincipalID, InterviewerRole: v.InterviewerRole,
		ScorecardTemplateReferenceId: v.ScorecardTemplateReferenceID, FeedbackDueAt: timestamp(v.FeedbackDueAt),
	}
}
func interviewPanelistResponse(v *entity.InterviewPanelist, err error) (*pb.InterviewPanelistResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.InterviewPanelistResponse{Item: interviewPanelist(v)}, nil
}
func interviewPanelists(page operation.Page[entity.InterviewPanelist], err error) (*pb.InterviewPanelistsResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	items := make([]*pb.InterviewPanelist, 0, len(page.Items))
	for i := range page.Items {
		items = append(items, interviewPanelist(&page.Items[i]))
	}
	return &pb.InterviewPanelistsResponse{Items: items, Page: pageInfo(page)}, nil
}
func (h *Handler) CreateInterviewPanelist(ctx context.Context, r *pb.CreateInterviewPanelistRequest) (*pb.InterviewPanelistResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	in := r.Input
	due, err := response.OptionalTimestamp(in.FeedbackDueAt)
	if err != nil {
		return nil, response.Error(err)
	}
	return interviewPanelistResponse(h.service.CreateInterviewPanelist(ctx, service.CreateInterviewPanelistInput{
		InterviewID: in.GetInterviewId(), InterviewerPrincipalID: in.GetInterviewerPrincipalId(), InterviewerRole: in.InterviewerRole,
		ScorecardTemplateReferenceID: in.ScorecardTemplateReferenceId, FeedbackDueAt: due,
	}))
}
func (h *Handler) UpdateInterviewPanelist(ctx context.Context, r *pb.UpdateInterviewPanelistRequest) (*pb.InterviewPanelistResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	m, err := response.ValidateMask(r.UpdateMask, "interview_id", "interviewer_principal_id", "interviewer_role", "scorecard_template_reference_id", "feedback_due_at")
	if err != nil {
		return nil, response.Error(err)
	}
	in := r.Input
	due, err := timeField(m["feedback_due_at"], in.FeedbackDueAt)
	if err != nil {
		return nil, response.Error(err)
	}
	return interviewPanelistResponse(h.service.UpdateInterviewPanelist(ctx, r.Id, service.UpdateInterviewPanelistInput{
		InterviewID: field(m["interview_id"], in.InterviewId), InterviewerPrincipalID: field(m["interviewer_principal_id"], in.InterviewerPrincipalId),
		InterviewerRole: field(m["interviewer_role"], in.InterviewerRole), ScorecardTemplateReferenceID: field(m["scorecard_template_reference_id"], in.ScorecardTemplateReferenceId), FeedbackDueAt: due,
	}))
}
func (h *Handler) DeleteInterviewPanelist(ctx context.Context, r *common.IDRequest) (*common.EmptyResponse, error) {
	return deleteRecord(ctx, r, h.service.DeleteInterviewPanelist)
}
func (h *Handler) GetInterviewPanelistByID(ctx context.Context, r *common.IDRequest) (*pb.InterviewPanelistResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return interviewPanelistResponse(h.service.GetInterviewPanelistByID(ctx, r.Id))
}
func (h *Handler) GetAllInterviewPanelists(ctx context.Context, r *common.PageRequest) (*pb.InterviewPanelistsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return interviewPanelists(h.service.GetAllInterviewPanelists(ctx, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindInterviewPanelistByInterviewID(ctx context.Context, r *common.StringFinderRequest) (*pb.InterviewPanelistsResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return interviewPanelists(h.service.FindInterviewPanelistByInterviewID(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindInterviewPanelistByInterviewerPrincipalID(ctx context.Context, r *common.StringFinderRequest) (*pb.InterviewPanelistsResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return interviewPanelists(h.service.FindInterviewPanelistByInterviewerPrincipalID(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindInterviewPanelistByScorecardTemplateReferenceID(ctx context.Context, r *common.StringFinderRequest) (*pb.InterviewPanelistsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return interviewPanelists(h.service.FindInterviewPanelistByScorecardTemplateReferenceID(ctx, r.Value, int(r.Limit), int(r.Offset)))
}
