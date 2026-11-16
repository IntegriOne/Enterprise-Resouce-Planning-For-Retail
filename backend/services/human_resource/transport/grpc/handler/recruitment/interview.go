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

func interview(v *entity.Interview) *pb.Interview {
	return &pb.Interview{
		Id: v.ID, ApplicationId: v.ApplicationID, InterviewRound: v.InterviewRound, InterviewType: v.InterviewType,
		ScheduledStart: timestamp(v.ScheduledStart), ScheduledEnd: timestamp(v.ScheduledEnd), LocationReferenceId: v.LocationReferenceID,
		MeetingReferenceId: v.MeetingReferenceID, Status: v.Status, CandidateTimezone: v.CandidateTimezone, CreatedByPrincipalId: v.CreatedByPrincipalID,
	}
}
func interviewResponse(v *entity.Interview, err error) (*pb.InterviewResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.InterviewResponse{Item: interview(v)}, nil
}
func interviews(page operation.Page[entity.Interview], err error) (*pb.InterviewsResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	items := make([]*pb.Interview, 0, len(page.Items))
	for i := range page.Items {
		items = append(items, interview(&page.Items[i]))
	}
	return &pb.InterviewsResponse{Items: items, Page: pageInfo(page)}, nil
}
func (h *Handler) CreateInterview(ctx context.Context, r *pb.CreateInterviewRequest) (*pb.InterviewResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	in := r.Input
	start, err := response.OptionalTimestamp(in.ScheduledStart)
	if err != nil {
		return nil, response.Error(err)
	}
	end, err := response.OptionalTimestamp(in.ScheduledEnd)
	if err != nil {
		return nil, response.Error(err)
	}
	return interviewResponse(h.service.CreateInterview(ctx, service.CreateInterviewInput{
		ApplicationID: in.GetApplicationId(), InterviewRound: in.GetInterviewRound(), InterviewType: in.GetInterviewType(), ScheduledStart: start, ScheduledEnd: end,
		LocationReferenceID: in.LocationReferenceId, MeetingReferenceID: in.MeetingReferenceId, Status: in.Status, CandidateTimezone: in.CandidateTimezone,
	}))
}
func (h *Handler) UpdateInterview(ctx context.Context, r *pb.UpdateInterviewRequest) (*pb.InterviewResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	m, err := response.ValidateMask(r.UpdateMask, "application_id", "interview_round", "interview_type", "scheduled_start", "scheduled_end", "location_reference_id", "meeting_reference_id", "status", "candidate_timezone")
	if err != nil {
		return nil, response.Error(err)
	}
	in := r.Input
	start, err := timeField(m["scheduled_start"], in.ScheduledStart)
	if err != nil {
		return nil, response.Error(err)
	}
	end, err := timeField(m["scheduled_end"], in.ScheduledEnd)
	if err != nil {
		return nil, response.Error(err)
	}
	return interviewResponse(h.service.UpdateInterview(ctx, r.Id, service.UpdateInterviewInput{
		ApplicationID: field(m["application_id"], in.ApplicationId), InterviewRound: field(m["interview_round"], in.InterviewRound), InterviewType: field(m["interview_type"], in.InterviewType),
		ScheduledStart: start, ScheduledEnd: end, LocationReferenceID: field(m["location_reference_id"], in.LocationReferenceId), MeetingReferenceID: field(m["meeting_reference_id"], in.MeetingReferenceId),
		Status: field(m["status"], in.Status), CandidateTimezone: field(m["candidate_timezone"], in.CandidateTimezone),
	}))
}
func (h *Handler) DeleteInterview(ctx context.Context, r *common.IDRequest) (*common.EmptyResponse, error) {
	return deleteRecord(ctx, r, h.service.DeleteInterview)
}
func (h *Handler) GetInterviewByID(ctx context.Context, r *common.IDRequest) (*pb.InterviewResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return interviewResponse(h.service.GetInterviewByID(ctx, r.Id))
}
func (h *Handler) GetAllInterviews(ctx context.Context, r *common.PageRequest) (*pb.InterviewsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return interviews(h.service.GetAllInterviews(ctx, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindInterviewByApplicationID(ctx context.Context, r *common.StringFinderRequest) (*pb.InterviewsResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return interviews(h.service.FindInterviewByApplicationID(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindInterviewByInterviewType(ctx context.Context, r *common.StringFinderRequest) (*pb.InterviewsResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return interviews(h.service.FindInterviewByInterviewType(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindInterviewByLocationReferenceID(ctx context.Context, r *common.StringFinderRequest) (*pb.InterviewsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return interviews(h.service.FindInterviewByLocationReferenceID(ctx, r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindInterviewByMeetingReferenceID(ctx context.Context, r *common.StringFinderRequest) (*pb.InterviewsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return interviews(h.service.FindInterviewByMeetingReferenceID(ctx, r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindInterviewByStatus(ctx context.Context, r *common.StringFinderRequest) (*pb.InterviewsResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return interviews(h.service.FindInterviewByStatus(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindInterviewByCreatedByPrincipalID(ctx context.Context, r *common.StringFinderRequest) (*pb.InterviewsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return interviews(h.service.FindInterviewByCreatedByPrincipalID(ctx, r.Value, int(r.Limit), int(r.Offset)))
}
