package competency

import (
	"context"

	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/competency/v1"
	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/competency"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/response"
)

func jobCompetency(value *entity.JobCompetency) *pb.JobCompetency {
	return &pb.JobCompetency{
		Id: value.ID, JobId: value.JobID, CompetencyId: value.CompetencyID,
		RequiredLevel: int32(value.RequiredLevel), IsMandatory: value.IsMandatory,
		EffectiveFrom: value.EffectiveFrom.Format("2006-01-02"), EffectiveTo: dateString(value.EffectiveTo),
	}
}
func jobCompetencyResponse(value *entity.JobCompetency, err error) (*pb.JobCompetencyResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.JobCompetencyResponse{Item: jobCompetency(value)}, nil
}
func jobCompetencies(page operation.Page[entity.JobCompetency], err error) (*pb.JobCompetenciesResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	items := make([]*pb.JobCompetency, 0, len(page.Items))
	for index := range page.Items {
		items = append(items, jobCompetency(&page.Items[index]))
	}
	return &pb.JobCompetenciesResponse{Items: items, Page: &common.PageInfo{Total: page.Total, Limit: int32(page.Limit), Offset: int32(page.Offset)}}, nil
}
func (h *Handler) CreateJobCompetency(ctx context.Context, request *pb.CreateJobCompetencyRequest) (*pb.JobCompetencyResponse, error) {
	if request == nil || request.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	in := request.Input
	required, err := level(in.GetRequiredLevel())
	if err != nil {
		return nil, response.Error(err)
	}
	from, err := response.Date(in.GetEffectiveFrom())
	if err != nil {
		return nil, response.Error(err)
	}
	to, err := optionalDate(in.EffectiveTo)
	if err != nil {
		return nil, response.Error(err)
	}
	return jobCompetencyResponse(h.service.CreateJobCompetency(ctx, service.CreateJobCompetencyInput{
		JobID: in.GetJobId(), CompetencyID: in.GetCompetencyId(), RequiredLevel: required,
		IsMandatory: in.IsMandatory, EffectiveFrom: from, EffectiveTo: to,
	}))
}
func (h *Handler) UpdateJobCompetency(ctx context.Context, request *pb.UpdateJobCompetencyRequest) (*pb.JobCompetencyResponse, error) {
	if request == nil || request.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	mask, err := response.ValidateMask(request.UpdateMask, "job_id", "competency_id", "required_level", "is_mandatory", "effective_from", "effective_to")
	if err != nil {
		return nil, response.Error(err)
	}
	in := request.Input
	required, err := levelField(mask["required_level"], in.RequiredLevel)
	if err != nil {
		return nil, response.Error(err)
	}
	from, err := dateField(mask["effective_from"], in.EffectiveFrom)
	if err != nil {
		return nil, response.Error(err)
	}
	to, err := dateField(mask["effective_to"], in.EffectiveTo)
	if err != nil {
		return nil, response.Error(err)
	}
	return jobCompetencyResponse(h.service.UpdateJobCompetency(ctx, request.Id, service.UpdateJobCompetencyInput{
		JobID: field(mask["job_id"], in.JobId), CompetencyID: field(mask["competency_id"], in.CompetencyId),
		RequiredLevel: required, IsMandatory: field(mask["is_mandatory"], in.IsMandatory),
		EffectiveFrom: from, EffectiveTo: to,
	}))
}
func (h *Handler) DeleteJobCompetency(ctx context.Context, request *common.IDRequest) (*common.EmptyResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	if err := h.service.DeleteJobCompetency(ctx, request.Id); err != nil {
		return nil, response.Error(err)
	}
	return &common.EmptyResponse{}, nil
}
func (h *Handler) GetJobCompetencyByID(ctx context.Context, request *common.IDRequest) (*pb.JobCompetencyResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return jobCompetencyResponse(h.service.GetJobCompetencyByID(ctx, request.Id))
}
func (h *Handler) GetAllJobCompetencies(ctx context.Context, request *common.PageRequest) (*pb.JobCompetenciesResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return jobCompetencies(h.service.GetAllJobCompetencies(ctx, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindJobCompetencyByJobID(ctx context.Context, request *common.StringFinderRequest) (*pb.JobCompetenciesResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return jobCompetencies(h.service.FindJobCompetencyByJobID(ctx, *request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindJobCompetencyByCompetencyID(ctx context.Context, request *common.StringFinderRequest) (*pb.JobCompetenciesResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return jobCompetencies(h.service.FindJobCompetencyByCompetencyID(ctx, *request.Value, int(request.Limit), int(request.Offset)))
}
