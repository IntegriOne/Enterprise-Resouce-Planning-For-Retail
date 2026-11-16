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

func job(value *entity.Job) (*pb.Job, error) {
	responsibilities, err := response.JSONValue(value.Responsibilities)
	if err != nil {
		return nil, err
	}
	qualifications, err := response.JSONValue(value.QualificationRequirements)
	if err != nil {
		return nil, err
	}
	return &pb.Job{Id: value.ID, JobFamilyId: value.JobFamilyID, JobCode: value.JobCode, JobTitle: value.JobTitle,
		JobLevelCode: value.JobLevelCode, JobDescription: value.JobDescription, Responsibilities: responsibilities,
		QualificationRequirements: qualifications, ValidFrom: value.ValidFrom.Format("2006-01-02"), ValidTo: dateString(value.ValidTo), IsActive: value.IsActive}, nil
}

func jobResponse(value *entity.Job, err error) (*pb.JobResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	item, err := job(value)
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.JobResponse{Item: item}, nil
}

func jobs(page operation.Page[entity.Job], err error) (*pb.JobsResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	items := make([]*pb.Job, 0, len(page.Items))
	for index := range page.Items {
		item, err := job(&page.Items[index])
		if err != nil {
			return nil, response.Error(err)
		}
		items = append(items, item)
	}
	return &pb.JobsResponse{Items: items, Page: &common.PageInfo{Total: page.Total, Limit: int32(page.Limit), Offset: int32(page.Offset)}}, nil
}

func (h *Handler) CreateJob(ctx context.Context, request *pb.CreateJobRequest) (*pb.JobResponse, error) {
	if request == nil || request.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	in := request.Input
	from, err := response.Date(in.GetValidFrom())
	if err != nil {
		return nil, response.Error(err)
	}
	to, err := optionalDate(in.ValidTo)
	if err != nil {
		return nil, response.Error(err)
	}
	responsibilities, err := response.JSON(in.Responsibilities)
	if err != nil {
		return nil, response.Error(err)
	}
	qualifications, err := response.JSON(in.QualificationRequirements)
	if err != nil {
		return nil, response.Error(err)
	}
	return jobResponse(h.service.CreateJob(ctx, service.CreateJobInput{
		JobFamilyID: in.GetJobFamilyId(), JobCode: in.GetJobCode(), JobTitle: in.GetJobTitle(), JobLevelCode: in.JobLevelCode,
		JobDescription: in.JobDescription, Responsibilities: responsibilities, QualificationRequirements: qualifications,
		ValidFrom: from, ValidTo: to, IsActive: in.IsActive,
	}))
}

func (h *Handler) UpdateJob(ctx context.Context, request *pb.UpdateJobRequest) (*pb.JobResponse, error) {
	if request == nil || request.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	mask, err := response.ValidateMask(request.UpdateMask, "job_family_id", "job_code", "job_title", "job_level_code", "job_description", "responsibilities", "qualification_requirements", "valid_from", "valid_to", "is_active")
	if err != nil {
		return nil, response.Error(err)
	}
	in := request.Input
	from, err := dateField(mask["valid_from"], in.ValidFrom)
	if err != nil {
		return nil, response.Error(err)
	}
	to, err := dateField(mask["valid_to"], in.ValidTo)
	if err != nil {
		return nil, response.Error(err)
	}
	patch := service.UpdateJobInput{JobFamilyID: field(mask["job_family_id"], in.JobFamilyId), JobCode: field(mask["job_code"], in.JobCode),
		JobTitle: field(mask["job_title"], in.JobTitle), JobLevelCode: field(mask["job_level_code"], in.JobLevelCode),
		JobDescription: field(mask["job_description"], in.JobDescription), ValidFrom: from, ValidTo: to, IsActive: field(mask["is_active"], in.IsActive)}
	if mask["responsibilities"] {
		patch.Responsibilities.Set = true
		if in.Responsibilities != nil {
			value, err := response.JSON(in.Responsibilities)
			if err != nil {
				return nil, response.Error(err)
			}
			patch.Responsibilities.Value = &value
		}
	}
	if mask["qualification_requirements"] {
		patch.QualificationRequirements.Set = true
		if in.QualificationRequirements != nil {
			value, err := response.JSON(in.QualificationRequirements)
			if err != nil {
				return nil, response.Error(err)
			}
			patch.QualificationRequirements.Value = &value
		}
	}
	return jobResponse(h.service.UpdateJob(ctx, request.Id, patch))
}

func (h *Handler) DeleteJob(ctx context.Context, request *common.IDRequest) (*common.EmptyResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	if err := h.service.DeleteJob(ctx, request.Id); err != nil {
		return nil, response.Error(err)
	}
	return &common.EmptyResponse{}, nil
}
func (h *Handler) GetJobByID(ctx context.Context, request *common.IDRequest) (*pb.JobResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return jobResponse(h.service.GetJobByID(ctx, request.Id))
}
func (h *Handler) GetAllJobs(ctx context.Context, request *common.PageRequest) (*pb.JobsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return jobs(h.service.GetAllJobs(ctx, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindJobByJobFamilyID(ctx context.Context, request *common.StringFinderRequest) (*pb.JobsResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return jobs(h.service.FindJobByJobFamilyID(ctx, *request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindJobByJobCode(ctx context.Context, request *common.StringFinderRequest) (*pb.JobsResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return jobs(h.service.FindJobByJobCode(ctx, *request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindJobByJobLevelCode(ctx context.Context, request *common.StringFinderRequest) (*pb.JobsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return jobs(h.service.FindJobByJobLevelCode(ctx, request.Value, int(request.Limit), int(request.Offset)))
}
