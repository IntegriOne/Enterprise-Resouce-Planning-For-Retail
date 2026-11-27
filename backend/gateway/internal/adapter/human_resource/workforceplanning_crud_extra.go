package human_resource

import (
	"context"

	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	workforce "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/workforceplanning/v1"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// UpdateJobFamily memperbarui rumpun jabatan berdasarkan FieldMask.
func (a *Adapter) UpdateJobFamily(ctx context.Context, rawID string, input model.UpdateJobFamilyInput) (*model.JobFamily, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_job_families", "update", "human_resource.job_families.update")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "job-family", "job-family")
	if err != nil {
		return nil, err
	}
	if a.workforce == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &workforce.JobFamilyInput{}
	mask := &fieldmaskpb.FieldMask{}
	if err := applyStringMask(mask, "family_code", input.FamilyCode, &patch.FamilyCode, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "family_name", input.FamilyName, &patch.FamilyName, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "description", input.Description, &patch.Description, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "job_domain", input.JobDomain, &patch.JobDomain, true); err != nil {
		return nil, err
	}
	if err := applyBoolMask(mask, "is_active", input.IsActive, &patch.IsActive, true); err != nil {
		return nil, err
	}
	if len(mask.Paths) == 0 {
		return nil, middleware.ErrInvalidArgument("job-family update is empty", nil)
	}
	resp, err := a.workforce.UpdateJobFamily(rpcCtx, &workforce.UpdateJobFamilyRequest{
		Id:         id,
		Input:      patch,
		UpdateMask: mask,
	})
	if err != nil {
		return nil, mapHRMWriteError(err, "job-family")
	}
	return toModelJobFamily(resp.GetItem())
}

// DeleteJobFamily menghapus rumpun jabatan berdasarkan ID.
func (a *Adapter) DeleteJobFamily(ctx context.Context, rawID string) (*bool, error) {
	if a == nil || a.workforce == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	return deleteByUUID(ctx, a, rawID, "hr_job_families", "human_resource.job_families.delete", "job-family", func(c context.Context, req *common.IDRequest) (*common.EmptyResponse, error) {
		return a.workforce.DeleteJobFamily(c, req)
	})
}

// UpdateJob memperbarui data jabatan berdasarkan FieldMask.
func (a *Adapter) UpdateJob(ctx context.Context, rawID string, input model.UpdateJobInput) (*model.Job, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_jobs", "update", "human_resource.jobs.update")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "job", "job")
	if err != nil {
		return nil, err
	}
	if a.workforce == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &workforce.JobInput{}
	mask := &fieldmaskpb.FieldMask{}
	if err := applyStringMask(mask, "job_family_id", input.JobFamilyID, &patch.JobFamilyId, true); err != nil {
		return nil, err
	}
	if patch.JobFamilyId != nil {
		if norm, nErr := normalizeRequiredUUID(*patch.JobFamilyId, "job-family", "job_family_id"); nErr != nil {
			return nil, nErr
		} else {
			patch.JobFamilyId = &norm
		}
	}
	if err := applyStringMask(mask, "job_code", input.JobCode, &patch.JobCode, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "job_title", input.JobTitle, &patch.JobTitle, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "job_level_code", input.JobLevelCode, &patch.JobLevelCode, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "job_description", input.JobDescription, &patch.JobDescription, true); err != nil {
		return nil, err
	}
	if val, set := input.Responsibilities.ValueOK(); set {
		if val == nil {
			patch.Responsibilities = nil
		} else if b, jErr := requiredJSONBytes(*val, "responsibilities"); jErr != nil {
			return nil, jErr
		} else {
			patch.Responsibilities = &common.JsonDocument{Utf8: b}
		}
		mask.Paths = append(mask.Paths, "responsibilities")
	}
	if val, set := input.QualificationRequirements.ValueOK(); set {
		if val == nil {
			patch.QualificationRequirements = nil
		} else if b, jErr := requiredJSONBytes(*val, "qualification_requirements"); jErr != nil {
			return nil, jErr
		} else {
			patch.QualificationRequirements = &common.JsonDocument{Utf8: b}
		}
		mask.Paths = append(mask.Paths, "qualification_requirements")
	}
	if err := applyStringMask(mask, "valid_from", input.ValidFrom, &patch.ValidFrom, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "valid_to", input.ValidTo, &patch.ValidTo, true); err != nil {
		return nil, err
	}
	if err := applyBoolMask(mask, "is_active", input.IsActive, &patch.IsActive, true); err != nil {
		return nil, err
	}
	if len(mask.Paths) == 0 {
		return nil, middleware.ErrInvalidArgument("job update is empty", nil)
	}
	resp, err := a.workforce.UpdateJob(rpcCtx, &workforce.UpdateJobRequest{
		Id:         id,
		Input:      patch,
		UpdateMask: mask,
	})
	if err != nil {
		return nil, mapHRMWriteError(err, "job")
	}
	return toModelJob(resp.GetItem())
}

// DeleteJob menghapus jabatan berdasarkan ID.
func (a *Adapter) DeleteJob(ctx context.Context, rawID string) (*bool, error) {
	if a == nil || a.workforce == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	return deleteByUUID(ctx, a, rawID, "hr_jobs", "human_resource.jobs.delete", "job", func(c context.Context, req *common.IDRequest) (*common.EmptyResponse, error) {
		return a.workforce.DeleteJob(c, req)
	})
}
