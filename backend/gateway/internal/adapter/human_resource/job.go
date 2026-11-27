package human_resource

import (
	"context"

	"github.com/google/uuid"
	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	workforce "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/workforceplanning/v1"
)

const jobReadPermission = "human_resource.jobs.read"

// Job membaca detail jabatan berdasarkan ID.
func (a *Adapter) Job(ctx context.Context, rawID string) (*model.Job, error) {
	rpcCtx, err := a.authorizeGlobalEmployee(ctx, jobReadPermission)
	if err != nil {
		return nil, err
	}
	id, err := parseNodeID(rawID, "job")
	if err != nil || id == uuid.Nil {
		return nil, middleware.ErrInvalidArgument("invalid job ID", nil)
	}
	if a.workforce == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	response, err := a.workforce.GetJobByID(rpcCtx, &common.IDRequest{Id: id.String()})
	if err != nil {
		return nil, mapReadError(err, "job")
	}
	job := response.GetItem()
	if job.GetId() != id.String() {
		return nil, middleware.ErrUnavailable("invalid human resource job response", nil)
	}
	return toModelJob(job)
}

// Jobs mendaftar jabatan dengan izin baca tersendiri.
func (a *Adapter) Jobs(ctx context.Context, limit, offset int, rawFamilyID *string) (*Page[model.Job], error) {
	rpcCtx, err := a.authorizeGlobalEmployee(ctx, jobReadPermission)
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid job pagination", nil)
	}
	if a.workforce == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	var response *workforce.JobsResponse
	var familyID string
	if rawFamilyID == nil {
		response, err = a.workforce.GetAllJobs(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	} else {
		id, parseErr := parseNodeID(*rawFamilyID, "job-family")
		if parseErr != nil || id == uuid.Nil {
			return nil, middleware.ErrInvalidArgument("invalid job family ID", nil)
		}
		familyID = id.String()
		response, err = a.workforce.FindJobByJobFamilyID(rpcCtx, &common.StringFinderRequest{
			Value: &familyID, Limit: int32(limit), Offset: int32(offset),
		})
	}
	if err != nil {
		return nil, mapReadError(err, "job")
	}
	page, items := response.GetPage(), response.GetItems()
	if !validPage(page, limit, offset, len(items)) {
		return nil, middleware.ErrUnavailable("invalid human resource job page", nil)
	}
	result := &Page[model.Job]{
		Items: make([]*model.Job, 0, len(items)), TotalCount: int(page.GetTotal()), Limit: limit, Offset: offset,
	}
	for _, job := range items {
		if rawFamilyID != nil && job.GetJobFamilyId() != familyID {
			return nil, middleware.ErrUnavailable("invalid human resource job family match", nil)
		}
		item, err := toModelJob(job)
		if err != nil {
			return nil, err
		}
		result.Items = append(result.Items, item)
	}
	return result, nil
}

func toModelJob(job *workforce.Job) (*model.Job, error) {
	for _, rawID := range []string{job.GetId(), job.GetJobFamilyId()} {
		if id, err := parseNodeID(rawID, "job"); err != nil || id == uuid.Nil {
			return nil, middleware.ErrUnavailable("invalid human resource job reference", nil)
		}
	}
	return &model.Job{
		ID: nodeID("job", job.GetId()), RawID: job.GetId(), JobFamilyID: job.GetJobFamilyId(), JobCode: job.GetJobCode(), JobTitle: job.GetJobTitle(),
		JobLevelCode: job.JobLevelCode, JobDescription: job.JobDescription, ValidFrom: job.GetValidFrom(),
		ValidTo: job.ValidTo, IsActive: job.GetIsActive(),
	}, nil
}

// JobsByIDs membaca label jabatan dengan satu panggilan RPC dan urutan input.
func (a *Adapter) JobsByIDs(ctx context.Context, rawIDs []string) ([]*model.Job, error) {
	rpcCtx, err := a.authorizeGlobalEmployee(ctx, jobReadPermission)
	if err != nil {
		return nil, err
	}
	if len(rawIDs) > 100 {
		return nil, middleware.ErrInvalidArgument("at most 100 job IDs are allowed", nil)
	}
	ids := make([]string, 0, len(rawIDs))
	requested := make(map[string]struct{}, len(rawIDs))
	for _, raw := range rawIDs {
		id, err := parseNodeID(raw, "job")
		if err != nil || id == uuid.Nil {
			return nil, middleware.ErrInvalidArgument("invalid job ID", nil)
		}
		canonical := id.String()
		if _, found := requested[canonical]; !found {
			requested[canonical] = struct{}{}
			ids = append(ids, canonical)
		}
	}
	nodes := make([]*model.Job, 0, len(ids))
	if len(ids) == 0 {
		return nodes, nil
	}
	if a.workforce == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	response, err := a.workforce.GetJobsByIds(rpcCtx, &workforce.GetJobsByIdsRequest{Ids: ids})
	if err != nil {
		return nil, mapReadError(err, "job")
	}
	page, items := response.GetPage(), response.GetItems()
	if !validPage(page, len(ids), 0, len(items)) || page.GetTotal() != int64(len(items)) {
		return nil, middleware.ErrUnavailable("invalid human resource job batch response", nil)
	}
	byID := make(map[string]*model.Job, len(items))
	for _, item := range items {
		id := item.GetId()
		if _, present := requested[id]; !present || byID[id] != nil {
			return nil, middleware.ErrUnavailable("unexpected job in human resource batch response", nil)
		}
		node, err := toModelJob(item)
		if err != nil {
			return nil, err
		}
		byID[id] = node
	}
	for _, id := range ids {
		if node := byID[id]; node != nil {
			nodes = append(nodes, node)
		}
	}
	return nodes, nil
}
