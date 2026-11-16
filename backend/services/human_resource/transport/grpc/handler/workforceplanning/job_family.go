package workforceplanning

import (
	"context"

	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/workforceplanning/v1"
	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/workforceplanning"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/response"
)

func jobFamily(value *entity.JobFamily) *pb.JobFamily {
	return &pb.JobFamily{Id: value.ID, FamilyCode: value.FamilyCode, FamilyName: value.FamilyName,
		Description: value.Description, JobDomain: value.JobDomain, IsActive: value.IsActive}
}

func jobFamilies(page operation.Page[entity.JobFamily], err error) (*pb.JobFamiliesResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	items := make([]*pb.JobFamily, 0, len(page.Items))
	for index := range page.Items {
		items = append(items, jobFamily(&page.Items[index]))
	}
	return &pb.JobFamiliesResponse{Items: items, Page: &common.PageInfo{Total: page.Total, Limit: int32(page.Limit), Offset: int32(page.Offset)}}, nil
}

func (h *Handler) CreateJobFamily(ctx context.Context, request *pb.CreateJobFamilyRequest) (*pb.JobFamilyResponse, error) {
	if request == nil || request.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	in := request.Input
	value, err := h.service.CreateJobFamily(ctx, service.CreateJobFamilyInput{FamilyCode: in.GetFamilyCode(), FamilyName: in.GetFamilyName(), Description: in.Description, JobDomain: in.GetJobDomain(), IsActive: in.IsActive})
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.JobFamilyResponse{Item: jobFamily(value)}, nil
}

func (h *Handler) UpdateJobFamily(ctx context.Context, request *pb.UpdateJobFamilyRequest) (*pb.JobFamilyResponse, error) {
	if request == nil || request.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	mask, err := response.ValidateMask(request.UpdateMask, "family_code", "family_name", "description", "job_domain", "is_active")
	if err != nil {
		return nil, response.Error(err)
	}
	in := request.Input
	value, err := h.service.UpdateJobFamily(ctx, request.Id, service.UpdateJobFamilyInput{
		FamilyCode: field(mask["family_code"], in.FamilyCode), FamilyName: field(mask["family_name"], in.FamilyName),
		Description: field(mask["description"], in.Description), JobDomain: field(mask["job_domain"], in.JobDomain), IsActive: field(mask["is_active"], in.IsActive),
	})
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.JobFamilyResponse{Item: jobFamily(value)}, nil
}

func (h *Handler) DeleteJobFamily(ctx context.Context, request *common.IDRequest) (*common.EmptyResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	if err := h.service.DeleteJobFamily(ctx, request.Id); err != nil {
		return nil, response.Error(err)
	}
	return &common.EmptyResponse{}, nil
}

func (h *Handler) GetJobFamilyByID(ctx context.Context, request *common.IDRequest) (*pb.JobFamilyResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	value, err := h.service.GetJobFamilyByID(ctx, request.Id)
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.JobFamilyResponse{Item: jobFamily(value)}, nil
}

func (h *Handler) GetAllJobFamilies(ctx context.Context, request *common.PageRequest) (*pb.JobFamiliesResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return jobFamilies(h.service.GetAllJobFamilies(ctx, int(request.Limit), int(request.Offset)))
}

func (h *Handler) FindJobFamilyByFamilyCode(ctx context.Context, request *common.StringFinderRequest) (*pb.JobFamiliesResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return jobFamilies(h.service.FindJobFamilyByFamilyCode(ctx, *request.Value, int(request.Limit), int(request.Offset)))
}
