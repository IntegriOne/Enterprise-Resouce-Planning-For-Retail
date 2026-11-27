package human_resource

import (
	"context"

	"github.com/google/uuid"
	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	workforce "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/workforceplanning/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CreateJobFamily menambahkan kelompok jabatan pada katalog HRM.
func (a *Adapter) CreateJobFamily(ctx context.Context, input model.CreateJobFamilyInput) (*model.JobFamily, error) {
	rpcCtx, err := a.authorizeGlobalEmployee(ctx, "human_resource.job_families.create")
	if err != nil {
		return nil, err
	}
	if a.workforce == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	response, err := a.workforce.CreateJobFamily(rpcCtx, &workforce.CreateJobFamilyRequest{Input: &workforce.JobFamilyInput{
		FamilyCode: &input.FamilyCode, FamilyName: &input.FamilyName, Description: input.Description,
		JobDomain: &input.JobDomain, IsActive: input.IsActive,
	}})
	if err != nil {
		return nil, mapJobCreationError(err, "job family")
	}
	return toModelJobFamily(response.GetItem())
}

// CreateJob menambahkan jabatan ke kelompok yang diperiksa HRM.
func (a *Adapter) CreateJob(ctx context.Context, input model.CreateJobInput) (*model.Job, error) {
	rpcCtx, err := a.authorizeGlobalEmployee(ctx, "human_resource.jobs.create")
	if err != nil {
		return nil, err
	}
	familyID, err := parseNodeID(input.JobFamilyID, "job-family")
	if err != nil || familyID == uuid.Nil {
		return nil, middleware.ErrInvalidArgument("invalid job family ID", nil)
	}
	if a.workforce == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	canonicalFamilyID := familyID.String()
	response, err := a.workforce.CreateJob(rpcCtx, &workforce.CreateJobRequest{Input: &workforce.JobInput{
		JobFamilyId: &canonicalFamilyID, JobCode: &input.JobCode, JobTitle: &input.JobTitle,
		JobLevelCode: input.JobLevelCode, JobDescription: input.JobDescription,
		ValidFrom: &input.ValidFrom, ValidTo: input.ValidTo, IsActive: input.IsActive,
	}})
	if err != nil {
		return nil, mapJobCreationError(err, "job")
	}
	item := response.GetItem()
	if item.GetJobFamilyId() != canonicalFamilyID {
		return nil, middleware.ErrUnavailable("invalid human resource job family response", nil)
	}
	return toModelJob(item)
}

func mapJobCreationError(err error, resource string) error {
	if code := status.Code(err); code == codes.AlreadyExists || code == codes.FailedPrecondition {
		return middleware.ErrConflict(resource+" conflicts with existing state or references", err)
	}
	return mapReadError(err, resource)
}
