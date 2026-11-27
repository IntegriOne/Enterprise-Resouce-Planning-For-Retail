package human_resource

import (
	"context"

	"github.com/google/uuid"
	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	workforce "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/workforceplanning/v1"
)

const jobFamilyReadPermission = "human_resource.job_families.read"

// JobFamily membaca detail kelompok jabatan berdasarkan ID.
func (a *Adapter) JobFamily(ctx context.Context, rawID string) (*model.JobFamily, error) {
	rpcCtx, err := a.authorizeGlobalEmployee(ctx, jobFamilyReadPermission)
	if err != nil {
		return nil, err
	}
	id, err := parseNodeID(rawID, "job-family")
	if err != nil || id == uuid.Nil {
		return nil, middleware.ErrInvalidArgument("invalid job family ID", nil)
	}
	if a.workforce == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	response, err := a.workforce.GetJobFamilyByID(rpcCtx, &common.IDRequest{Id: id.String()})
	if err != nil {
		return nil, mapReadError(err, "job family")
	}
	family := response.GetItem()
	if family.GetId() != id.String() {
		return nil, middleware.ErrUnavailable("invalid human resource job family response", nil)
	}
	return toModelJobFamily(family)
}

// JobFamilies mendaftar kelompok jabatan dengan izin baca tersendiri.
func (a *Adapter) JobFamilies(ctx context.Context, limit, offset int) (*Page[model.JobFamily], error) {
	rpcCtx, err := a.authorizeGlobalEmployee(ctx, jobFamilyReadPermission)
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid job family pagination", nil)
	}
	if a.workforce == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	response, err := a.workforce.GetAllJobFamilies(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	if err != nil {
		return nil, mapReadError(err, "job family")
	}
	page, items := response.GetPage(), response.GetItems()
	if !validPage(page, limit, offset, len(items)) {
		return nil, middleware.ErrUnavailable("invalid human resource job family page", nil)
	}
	result := &Page[model.JobFamily]{
		Items: make([]*model.JobFamily, 0, len(items)), TotalCount: int(page.GetTotal()), Limit: limit, Offset: offset,
	}
	for _, family := range items {
		item, err := toModelJobFamily(family)
		if err != nil {
			return nil, err
		}
		result.Items = append(result.Items, item)
	}
	return result, nil
}

func toModelJobFamily(row *workforce.JobFamily) (*model.JobFamily, error) {
	if id, err := uuid.Parse(row.GetId()); err != nil || id == uuid.Nil {
		return nil, middleware.ErrUnavailable("invalid human resource job family reference", nil)
	}
	return &model.JobFamily{
		ID: nodeID("job-family", row.GetId()), RawID: row.GetId(), FamilyCode: row.GetFamilyCode(), FamilyName: row.GetFamilyName(),
		Description: row.Description, JobDomain: row.GetJobDomain(), IsActive: row.GetIsActive(),
	}, nil
}
