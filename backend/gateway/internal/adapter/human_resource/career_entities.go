package human_resource

import (
	"context"

	"github.com/google/uuid"
	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	career "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/career/v1"
	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func toModelDevelopmentPlan(row *career.DevelopmentPlan) (*model.DevelopmentPlan, error) {
	if row == nil {
		return nil, middleware.ErrUnavailable("invalid human resource development-plan response", nil)
	}
	if id, err := uuid.Parse(row.GetId()); err != nil || id == uuid.Nil {
		return nil, middleware.ErrUnavailable("invalid human resource development-plan ID", nil)
	}
	out := &model.DevelopmentPlan{ID: nodeID("development-plan", row.GetId()), RawID: row.GetId()}
	out.WorkerID = row.GetWorkerId()
	out.PlanCode = row.GetPlanCode()
	out.PlanVersion = int(row.GetPlanVersion())
	out.TargetJobID = row.TargetJobId
	out.StartDate = pbTimePtr(row.GetStartDate())
	out.TargetDate = pbTimePtr(row.GetTargetDate())
	out.Status = row.GetStatus()
	out.MentorWorkerID = row.MentorWorkerId
	out.ApprovedByPrincipalID = row.ApprovedByPrincipalId
	out.ApprovedAt = pbTimePtr(row.GetApprovedAt())
	return out, nil
}

// DevelopmentPlan membaca detail development-plan berdasarkan ID.
func (a *Adapter) DevelopmentPlan(ctx context.Context, rawID string) (*model.DevelopmentPlan, error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_development_plans", "human_resource.development_plans.read")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "development-plan", "development-plan")
	if err != nil {
		return nil, err
	}
	if a.career == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.career.GetDevelopmentPlanByID(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapReadError(err, "development-plan")
	}
	return toModelDevelopmentPlan(resp.GetValue())
}

// DevelopmentPlans mendaftar dan memfilter development-plan dengan paginasi.
func (a *Adapter) DevelopmentPlans(ctx context.Context, limit, offset int, f_worker_id *string, f_plan_code *string, f_target_job_id *string, f_status *string, f_mentor_worker_id *string, f_approved_by_principal_id *string) (*Page[model.DevelopmentPlan], error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_development_plans", "human_resource.development_plans.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid development-plan pagination", nil)
	}
	if a.career == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	idx, err := pickSingleFilter(f_worker_id != nil, f_plan_code != nil, f_target_job_id != nil, f_status != nil, f_mentor_worker_id != nil, f_approved_by_principal_id != nil)
	if err != nil {
		return nil, err
	}
	var resp *career.DevelopmentPlansResponse
	switch idx {
	case 0:
		v, fErr := validateFilterString(f_worker_id, "worker_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.career.FindDevelopmentPlanByWorkerID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 1:
		v, fErr := validateFilterString(f_plan_code, "plan_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.career.FindDevelopmentPlanByPlanCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 2:
		v, fErr := validateFilterString(f_target_job_id, "target_job_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.career.FindDevelopmentPlanByTargetJobID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 3:
		v, fErr := validateFilterString(f_status, "status")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.career.FindDevelopmentPlanByStatus(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 4:
		v, fErr := validateFilterString(f_mentor_worker_id, "mentor_worker_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.career.FindDevelopmentPlanByMentorWorkerID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 5:
		v, fErr := validateFilterString(f_approved_by_principal_id, "approved_by_principal_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.career.FindDevelopmentPlanByApprovedByPrincipalID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	default:
		resp, err = a.career.GetAllDevelopmentPlans(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "development-plan")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "development-plan", toModelDevelopmentPlan)
}

// CreateDevelopmentPlan membuat development-plan baru melalui HRM.
func (a *Adapter) CreateDevelopmentPlan(ctx context.Context, input model.CreateDevelopmentPlanInput) (*model.DevelopmentPlan, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_development_plans", "create", "human_resource.development_plans.create")
	if err != nil {
		return nil, err
	}
	if a.career == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &career.DevelopmentPlanInput{}
	patch.WorkerId = input.WorkerID
	patch.PlanCode = input.PlanCode
	patch.PlanVersion = optionalInt32Ptr(input.PlanVersion)
	patch.TargetJobId = input.TargetJobID
	patch.StartDate = optionalTimestamp(input.StartDate)
	patch.TargetDate = optionalTimestamp(input.TargetDate)
	patch.Status = input.Status
	patch.MentorWorkerId = input.MentorWorkerID
	patch.ApprovedByPrincipalId = input.ApprovedByPrincipalID
	patch.ApprovedAt = optionalTimestamp(input.ApprovedAt)
	resp, err := a.career.CreateDevelopmentPlan(rpcCtx, &career.CreateDevelopmentPlanRequest{Input: patch})
	if err != nil {
		return nil, mapHRMWriteError(err, "development-plan")
	}
	return toModelDevelopmentPlan(resp.GetValue())
}

// UpdateDevelopmentPlan memperbarui development-plan berdasarkan FieldMask.
func (a *Adapter) UpdateDevelopmentPlan(ctx context.Context, rawID string, input model.UpdateDevelopmentPlanInput) (*model.DevelopmentPlan, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_development_plans", "update", "human_resource.development_plans.update")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "development-plan", "development-plan")
	if err != nil {
		return nil, err
	}
	if a.career == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &career.DevelopmentPlanInput{}
	mask := &fieldmaskpb.FieldMask{}
	if err := applyStringMask(mask, "worker_id", input.WorkerID, &patch.WorkerId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "plan_code", input.PlanCode, &patch.PlanCode, true); err != nil {
		return nil, err
	}
	if err := applyInt32Mask(mask, "plan_version", input.PlanVersion, &patch.PlanVersion, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "target_job_id", input.TargetJobID, &patch.TargetJobId, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "start_date", input.StartDate, &patch.StartDate, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "target_date", input.TargetDate, &patch.TargetDate, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "status", input.Status, &patch.Status, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "mentor_worker_id", input.MentorWorkerID, &patch.MentorWorkerId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "approved_by_principal_id", input.ApprovedByPrincipalID, &patch.ApprovedByPrincipalId, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "approved_at", input.ApprovedAt, &patch.ApprovedAt, true); err != nil {
		return nil, err
	}
	if len(mask.Paths) == 0 {
		return nil, middleware.ErrInvalidArgument("development-plan update is empty", nil)
	}
	resp, err := a.career.UpdateDevelopmentPlan(rpcCtx, &career.UpdateDevelopmentPlanRequest{Id: id, Input: patch, UpdateMask: mask})
	if err != nil {
		return nil, mapHRMWriteError(err, "development-plan")
	}
	return toModelDevelopmentPlan(resp.GetValue())
}

// DeleteDevelopmentPlan menghapus development-plan berdasarkan ID.
func (a *Adapter) DeleteDevelopmentPlan(ctx context.Context, rawID string) (*bool, error) {
	if a == nil || a.career == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	return deleteByUUID(ctx, a, rawID, "hr_development_plans", "human_resource.development_plans.delete", "development-plan", func(c context.Context, req *common.IDRequest) (*common.EmptyResponse, error) {
		return a.career.DeleteDevelopmentPlan(c, req)
	})
}

func toModelDevelopmentAction(row *career.DevelopmentAction) (*model.DevelopmentAction, error) {
	if row == nil {
		return nil, middleware.ErrUnavailable("invalid human resource development-action response", nil)
	}
	if id, err := uuid.Parse(row.GetId()); err != nil || id == uuid.Nil {
		return nil, middleware.ErrUnavailable("invalid human resource development-action ID", nil)
	}
	out := &model.DevelopmentAction{ID: nodeID("development-action", row.GetId()), RawID: row.GetId()}
	out.DevelopmentPlanID = row.GetDevelopmentPlanId()
	out.ActionNumber = int(row.GetActionNumber())
	out.CompetencyID = row.CompetencyId
	out.CourseID = row.CourseId
	out.ActionType = row.GetActionType()
	out.Objective = row.GetObjective()
	out.TargetLevel = pbIntPtr(row.TargetLevel)
	out.DueDate = pbTimePtr(row.GetDueDate())
	out.CompletedAt = pbTimePtr(row.GetCompletedAt())
	out.EvidenceDocumentVersionReferenceID = row.EvidenceDocumentVersionReferenceId
	return out, nil
}

// DevelopmentAction membaca detail development-action berdasarkan ID.
func (a *Adapter) DevelopmentAction(ctx context.Context, rawID string) (*model.DevelopmentAction, error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_development_actions", "human_resource.development_actions.read")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "development-action", "development-action")
	if err != nil {
		return nil, err
	}
	if a.career == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.career.GetDevelopmentActionByID(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapReadError(err, "development-action")
	}
	return toModelDevelopmentAction(resp.GetValue())
}

// DevelopmentActions mendaftar dan memfilter development-action dengan paginasi.
func (a *Adapter) DevelopmentActions(ctx context.Context, limit, offset int, f_development_plan_id *string, f_action_number *int, f_competency_id *string, f_course_id *string, f_action_type *string, f_evidence_document_version_reference_id *string) (*Page[model.DevelopmentAction], error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_development_actions", "human_resource.development_actions.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid development-action pagination", nil)
	}
	if a.career == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	idx, err := pickSingleFilter(f_development_plan_id != nil, f_action_number != nil, f_competency_id != nil, f_course_id != nil, f_action_type != nil, f_evidence_document_version_reference_id != nil)
	if err != nil {
		return nil, err
	}
	var resp *career.DevelopmentActionsResponse
	switch idx {
	case 0:
		v, fErr := validateFilterString(f_development_plan_id, "development_plan_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.career.FindDevelopmentActionByDevelopmentPlanID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 1:
		v := int32(*f_action_number)
		resp, err = a.career.FindDevelopmentActionByActionNumber(rpcCtx, &common.Int32FinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 2:
		v, fErr := validateFilterString(f_competency_id, "competency_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.career.FindDevelopmentActionByCompetencyID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 3:
		v, fErr := validateFilterString(f_course_id, "course_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.career.FindDevelopmentActionByCourseID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 4:
		v, fErr := validateFilterString(f_action_type, "action_type")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.career.FindDevelopmentActionByActionType(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 5:
		v, fErr := validateFilterString(f_evidence_document_version_reference_id, "evidence_document_version_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.career.FindDevelopmentActionByEvidenceDocumentVersionReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	default:
		resp, err = a.career.GetAllDevelopmentActions(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "development-action")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "development-action", toModelDevelopmentAction)
}

// CreateDevelopmentAction membuat development-action baru melalui HRM.
func (a *Adapter) CreateDevelopmentAction(ctx context.Context, input model.CreateDevelopmentActionInput) (*model.DevelopmentAction, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_development_actions", "create", "human_resource.development_actions.create")
	if err != nil {
		return nil, err
	}
	if a.career == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &career.DevelopmentActionInput{}
	patch.DevelopmentPlanId = input.DevelopmentPlanID
	patch.ActionNumber = optionalInt32Ptr(input.ActionNumber)
	patch.CompetencyId = input.CompetencyID
	patch.CourseId = input.CourseID
	patch.ActionType = input.ActionType
	patch.Objective = input.Objective
	patch.TargetLevel = optionalInt32Ptr(input.TargetLevel)
	patch.DueDate = optionalTimestamp(input.DueDate)
	patch.CompletedAt = optionalTimestamp(input.CompletedAt)
	patch.EvidenceDocumentVersionReferenceId = input.EvidenceDocumentVersionReferenceID
	resp, err := a.career.CreateDevelopmentAction(rpcCtx, &career.CreateDevelopmentActionRequest{Input: patch})
	if err != nil {
		return nil, mapHRMWriteError(err, "development-action")
	}
	return toModelDevelopmentAction(resp.GetValue())
}

// UpdateDevelopmentAction memperbarui development-action berdasarkan FieldMask.
func (a *Adapter) UpdateDevelopmentAction(ctx context.Context, rawID string, input model.UpdateDevelopmentActionInput) (*model.DevelopmentAction, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_development_actions", "update", "human_resource.development_actions.update")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "development-action", "development-action")
	if err != nil {
		return nil, err
	}
	if a.career == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &career.DevelopmentActionInput{}
	mask := &fieldmaskpb.FieldMask{}
	if err := applyStringMask(mask, "development_plan_id", input.DevelopmentPlanID, &patch.DevelopmentPlanId, true); err != nil {
		return nil, err
	}
	if err := applyInt32Mask(mask, "action_number", input.ActionNumber, &patch.ActionNumber, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "competency_id", input.CompetencyID, &patch.CompetencyId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "course_id", input.CourseID, &patch.CourseId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "action_type", input.ActionType, &patch.ActionType, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "objective", input.Objective, &patch.Objective, true); err != nil {
		return nil, err
	}
	if err := applyInt32Mask(mask, "target_level", input.TargetLevel, &patch.TargetLevel, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "due_date", input.DueDate, &patch.DueDate, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "completed_at", input.CompletedAt, &patch.CompletedAt, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "evidence_document_version_reference_id", input.EvidenceDocumentVersionReferenceID, &patch.EvidenceDocumentVersionReferenceId, true); err != nil {
		return nil, err
	}
	if len(mask.Paths) == 0 {
		return nil, middleware.ErrInvalidArgument("development-action update is empty", nil)
	}
	resp, err := a.career.UpdateDevelopmentAction(rpcCtx, &career.UpdateDevelopmentActionRequest{Id: id, Input: patch, UpdateMask: mask})
	if err != nil {
		return nil, mapHRMWriteError(err, "development-action")
	}
	return toModelDevelopmentAction(resp.GetValue())
}

// DeleteDevelopmentAction menghapus development-action berdasarkan ID.
func (a *Adapter) DeleteDevelopmentAction(ctx context.Context, rawID string) (*bool, error) {
	if a == nil || a.career == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	return deleteByUUID(ctx, a, rawID, "hr_development_actions", "human_resource.development_actions.delete", "development-action", func(c context.Context, req *common.IDRequest) (*common.EmptyResponse, error) {
		return a.career.DeleteDevelopmentAction(c, req)
	})
}

func toModelCareerPath(row *career.CareerPath) (*model.CareerPath, error) {
	if row == nil {
		return nil, middleware.ErrUnavailable("invalid human resource career-path response", nil)
	}
	if id, err := uuid.Parse(row.GetId()); err != nil || id == uuid.Nil {
		return nil, middleware.ErrUnavailable("invalid human resource career-path ID", nil)
	}
	out := &model.CareerPath{ID: nodeID("career-path", row.GetId()), RawID: row.GetId()}
	out.PathCode = row.GetPathCode()
	out.PathVersion = int(row.GetPathVersion())
	out.PathName = row.GetPathName()
	out.JobFamilyID = row.GetJobFamilyId()
	out.ValidFrom = pbTimePtr(row.GetValidFrom())
	out.ValidTo = pbTimePtr(row.GetValidTo())
	out.ApprovedByPrincipalID = row.ApprovedByPrincipalId
	out.ApprovedAt = pbTimePtr(row.GetApprovedAt())
	out.Status = row.GetStatus()
	out.CreatedByPrincipalID = row.GetCreatedByPrincipalId()
	out.CreatedAt = pbTimePtr(row.GetCreatedAt())
	return out, nil
}

// CareerPath membaca detail career-path berdasarkan ID.
func (a *Adapter) CareerPath(ctx context.Context, rawID string) (*model.CareerPath, error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_career_paths", "human_resource.career_paths.read")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "career-path", "career-path")
	if err != nil {
		return nil, err
	}
	if a.career == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.career.GetCareerPathByID(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapReadError(err, "career-path")
	}
	return toModelCareerPath(resp.GetValue())
}

// CareerPaths mendaftar dan memfilter career-path dengan paginasi.
func (a *Adapter) CareerPaths(ctx context.Context, limit, offset int, f_path_code *string, f_job_family_id *string, f_status *string, f_created_by_principal_id *string, f_approved_by_principal_id *string) (*Page[model.CareerPath], error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_career_paths", "human_resource.career_paths.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid career-path pagination", nil)
	}
	if a.career == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	idx, err := pickSingleFilter(f_path_code != nil, f_job_family_id != nil, f_status != nil, f_created_by_principal_id != nil, f_approved_by_principal_id != nil)
	if err != nil {
		return nil, err
	}
	var resp *career.CareerPathsResponse
	switch idx {
	case 0:
		v, fErr := validateFilterString(f_path_code, "path_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.career.FindCareerPathByPathCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 1:
		v, fErr := validateFilterString(f_job_family_id, "job_family_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.career.FindCareerPathByJobFamilyID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 2:
		v, fErr := validateFilterString(f_status, "status")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.career.FindCareerPathByStatus(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 3:
		v, fErr := validateFilterString(f_created_by_principal_id, "created_by_principal_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.career.FindCareerPathByCreatedByPrincipalID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 4:
		v, fErr := validateFilterString(f_approved_by_principal_id, "approved_by_principal_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.career.FindCareerPathByApprovedByPrincipalID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	default:
		resp, err = a.career.GetAllCareerPaths(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "career-path")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "career-path", toModelCareerPath)
}

// CreateCareerPath membuat career-path baru melalui HRM.
func (a *Adapter) CreateCareerPath(ctx context.Context, input model.CreateCareerPathInput) (*model.CareerPath, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_career_paths", "create", "human_resource.career_paths.create")
	if err != nil {
		return nil, err
	}
	if a.career == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &career.CareerPathInput{}
	patch.PathCode = input.PathCode
	patch.PathVersion = optionalInt32Ptr(input.PathVersion)
	patch.PathName = input.PathName
	patch.JobFamilyId = input.JobFamilyID
	patch.ValidFrom = optionalTimestamp(input.ValidFrom)
	patch.ValidTo = optionalTimestamp(input.ValidTo)
	resp, err := a.career.CreateCareerPath(rpcCtx, &career.CreateCareerPathRequest{Input: patch})
	if err != nil {
		return nil, mapHRMWriteError(err, "career-path")
	}
	return toModelCareerPath(resp.GetValue())
}

// UpdateCareerPath memperbarui career-path berdasarkan FieldMask.
func (a *Adapter) UpdateCareerPath(ctx context.Context, rawID string, input model.UpdateCareerPathInput) (*model.CareerPath, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_career_paths", "update", "human_resource.career_paths.update")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "career-path", "career-path")
	if err != nil {
		return nil, err
	}
	if a.career == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &career.CareerPathInput{}
	mask := &fieldmaskpb.FieldMask{}
	if err := applyStringMask(mask, "path_code", input.PathCode, &patch.PathCode, true); err != nil {
		return nil, err
	}
	if err := applyInt32Mask(mask, "path_version", input.PathVersion, &patch.PathVersion, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "path_name", input.PathName, &patch.PathName, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "job_family_id", input.JobFamilyID, &patch.JobFamilyId, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "valid_from", input.ValidFrom, &patch.ValidFrom, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "valid_to", input.ValidTo, &patch.ValidTo, true); err != nil {
		return nil, err
	}
	if len(mask.Paths) == 0 {
		return nil, middleware.ErrInvalidArgument("career-path update is empty", nil)
	}
	resp, err := a.career.UpdateCareerPath(rpcCtx, &career.UpdateCareerPathRequest{Id: id, Input: patch, UpdateMask: mask})
	if err != nil {
		return nil, mapHRMWriteError(err, "career-path")
	}
	return toModelCareerPath(resp.GetValue())
}

// DeleteCareerPath menghapus career-path berdasarkan ID.
func (a *Adapter) DeleteCareerPath(ctx context.Context, rawID string) (*bool, error) {
	if a == nil || a.career == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	return deleteByUUID(ctx, a, rawID, "hr_career_paths", "human_resource.career_paths.delete", "career-path", func(c context.Context, req *common.IDRequest) (*common.EmptyResponse, error) {
		return a.career.DeleteCareerPath(c, req)
	})
}

func toModelCareerPathTransition(row *career.CareerPathTransition) (*model.CareerPathTransition, error) {
	if row == nil {
		return nil, middleware.ErrUnavailable("invalid human resource career-path-transition response", nil)
	}
	if id, err := uuid.Parse(row.GetId()); err != nil || id == uuid.Nil {
		return nil, middleware.ErrUnavailable("invalid human resource career-path-transition ID", nil)
	}
	out := &model.CareerPathTransition{ID: nodeID("career-path-transition", row.GetId()), RawID: row.GetId()}
	out.CareerPathID = row.GetCareerPathId()
	out.FromJobID = row.GetFromJobId()
	out.ToJobID = row.GetToJobId()
	out.MovementType = row.GetMovementType()
	out.MinimumTenureMonths = int(row.GetMinimumTenureMonths())
	out.RequirementsDocumentVersionReferenceID = row.GetRequirementsDocumentVersionReferenceId()
	out.RequirementsDigest = pbBase64StringValue(row.GetRequirementsDigest())
	return out, nil
}

// CareerPathTransition membaca detail career-path-transition berdasarkan ID.
func (a *Adapter) CareerPathTransition(ctx context.Context, rawID string) (*model.CareerPathTransition, error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_career_path_transitions", "human_resource.career_path_transitions.read")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "career-path-transition", "career-path-transition")
	if err != nil {
		return nil, err
	}
	if a.career == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.career.GetCareerPathTransitionByID(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapReadError(err, "career-path-transition")
	}
	return toModelCareerPathTransition(resp.GetValue())
}

// CareerPathTransitions mendaftar dan memfilter career-path-transition dengan paginasi.
func (a *Adapter) CareerPathTransitions(ctx context.Context, limit, offset int, f_career_path_id *string, f_from_job_id *string, f_to_job_id *string, f_movement_type *string, f_requirements_document_version_reference_id *string) (*Page[model.CareerPathTransition], error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_career_path_transitions", "human_resource.career_path_transitions.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid career-path-transition pagination", nil)
	}
	if a.career == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	idx, err := pickSingleFilter(f_career_path_id != nil, f_from_job_id != nil, f_to_job_id != nil, f_movement_type != nil, f_requirements_document_version_reference_id != nil)
	if err != nil {
		return nil, err
	}
	var resp *career.CareerPathTransitionsResponse
	switch idx {
	case 0:
		v, fErr := validateFilterString(f_career_path_id, "career_path_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.career.FindCareerPathTransitionByCareerPathID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 1:
		v, fErr := validateFilterString(f_from_job_id, "from_job_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.career.FindCareerPathTransitionByFromJobID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 2:
		v, fErr := validateFilterString(f_to_job_id, "to_job_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.career.FindCareerPathTransitionByToJobID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 3:
		v, fErr := validateFilterString(f_movement_type, "movement_type")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.career.FindCareerPathTransitionByMovementType(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 4:
		v, fErr := validateFilterString(f_requirements_document_version_reference_id, "requirements_document_version_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.career.FindCareerPathTransitionByRequirementsDocumentVersionReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	default:
		resp, err = a.career.GetAllCareerPathTransitions(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "career-path-transition")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "career-path-transition", toModelCareerPathTransition)
}

// CreateCareerPathTransition membuat career-path-transition baru melalui HRM.
func (a *Adapter) CreateCareerPathTransition(ctx context.Context, input model.CreateCareerPathTransitionInput) (*model.CareerPathTransition, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_career_path_transitions", "create", "human_resource.career_path_transitions.create")
	if err != nil {
		return nil, err
	}
	if a.career == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &career.CareerPathTransitionInput{}
	patch.CareerPathId = input.CareerPathID
	patch.FromJobId = input.FromJobID
	patch.ToJobId = input.ToJobID
	patch.MovementType = input.MovementType
	patch.MinimumTenureMonths = optionalInt32Ptr(input.MinimumTenureMonths)
	patch.RequirementsDocumentVersionReferenceId = input.RequirementsDocumentVersionReferenceID
	if b, bErr := optionalBase64Bytes(input.RequirementsDigest, "requirements_digest"); bErr != nil {
		return nil, bErr
	} else if b != nil {
		patch.RequirementsDigest = b
	}
	resp, err := a.career.CreateCareerPathTransition(rpcCtx, &career.CreateCareerPathTransitionRequest{Input: patch})
	if err != nil {
		return nil, mapHRMWriteError(err, "career-path-transition")
	}
	return toModelCareerPathTransition(resp.GetValue())
}

// UpdateCareerPathTransition memperbarui career-path-transition berdasarkan FieldMask.
func (a *Adapter) UpdateCareerPathTransition(ctx context.Context, rawID string, input model.UpdateCareerPathTransitionInput) (*model.CareerPathTransition, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_career_path_transitions", "update", "human_resource.career_path_transitions.update")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "career-path-transition", "career-path-transition")
	if err != nil {
		return nil, err
	}
	if a.career == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &career.CareerPathTransitionInput{}
	mask := &fieldmaskpb.FieldMask{}
	if err := applyStringMask(mask, "career_path_id", input.CareerPathID, &patch.CareerPathId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "from_job_id", input.FromJobID, &patch.FromJobId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "to_job_id", input.ToJobID, &patch.ToJobId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "movement_type", input.MovementType, &patch.MovementType, true); err != nil {
		return nil, err
	}
	if err := applyInt32Mask(mask, "minimum_tenure_months", input.MinimumTenureMonths, &patch.MinimumTenureMonths, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "requirements_document_version_reference_id", input.RequirementsDocumentVersionReferenceID, &patch.RequirementsDocumentVersionReferenceId, true); err != nil {
		return nil, err
	}
	if err := applyBytesMask(mask, "requirements_digest", input.RequirementsDigest, &patch.RequirementsDigest, true); err != nil {
		return nil, err
	}
	if len(mask.Paths) == 0 {
		return nil, middleware.ErrInvalidArgument("career-path-transition update is empty", nil)
	}
	resp, err := a.career.UpdateCareerPathTransition(rpcCtx, &career.UpdateCareerPathTransitionRequest{Id: id, Input: patch, UpdateMask: mask})
	if err != nil {
		return nil, mapHRMWriteError(err, "career-path-transition")
	}
	return toModelCareerPathTransition(resp.GetValue())
}

// DeleteCareerPathTransition menghapus career-path-transition berdasarkan ID.
func (a *Adapter) DeleteCareerPathTransition(ctx context.Context, rawID string) (*bool, error) {
	if a == nil || a.career == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	return deleteByUUID(ctx, a, rawID, "hr_career_path_transitions", "human_resource.career_path_transitions.delete", "career-path-transition", func(c context.Context, req *common.IDRequest) (*common.EmptyResponse, error) {
		return a.career.DeleteCareerPathTransition(c, req)
	})
}

func toModelWorkerCareerAspiration(row *career.WorkerCareerAspiration) (*model.WorkerCareerAspiration, error) {
	if row == nil {
		return nil, middleware.ErrUnavailable("invalid human resource worker-career-aspiration response", nil)
	}
	if id, err := uuid.Parse(row.GetId()); err != nil || id == uuid.Nil {
		return nil, middleware.ErrUnavailable("invalid human resource worker-career-aspiration ID", nil)
	}
	out := &model.WorkerCareerAspiration{ID: nodeID("worker-career-aspiration", row.GetId()), RawID: row.GetId()}
	out.WorkerID = row.GetWorkerId()
	out.TargetJobID = row.GetTargetJobId()
	out.CareerPathTransitionID = row.CareerPathTransitionId
	out.PreferenceRank = int(row.GetPreferenceRank())
	out.RecordedAt = pbTimePtr(row.GetRecordedAt())
	out.RecordedByPrincipalID = row.GetRecordedByPrincipalId()
	out.TargetDate = pbTimePtr(row.GetTargetDate())
	out.MobilityPreference = row.GetMobilityPreference()
	out.PreferredLocationReferenceID = row.PreferredLocationReferenceId
	out.WithdrawnAt = pbTimePtr(row.GetWithdrawnAt())
	return out, nil
}

// WorkerCareerAspiration membaca detail worker-career-aspiration berdasarkan ID.
func (a *Adapter) WorkerCareerAspiration(ctx context.Context, rawID string) (*model.WorkerCareerAspiration, error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_worker_career_aspirations", "human_resource.worker_career_aspirations.read")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "worker-career-aspiration", "worker-career-aspiration")
	if err != nil {
		return nil, err
	}
	if a.career == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.career.GetWorkerCareerAspirationByID(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapReadError(err, "worker-career-aspiration")
	}
	return toModelWorkerCareerAspiration(resp.GetValue())
}

// WorkerCareerAspirations mendaftar dan memfilter worker-career-aspiration dengan paginasi.
func (a *Adapter) WorkerCareerAspirations(ctx context.Context, limit, offset int, f_worker_id *string, f_target_job_id *string, f_career_path_transition_id *string, f_recorded_by_principal_id *string, f_preferred_location_reference_id *string) (*Page[model.WorkerCareerAspiration], error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_worker_career_aspirations", "human_resource.worker_career_aspirations.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid worker-career-aspiration pagination", nil)
	}
	if a.career == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	idx, err := pickSingleFilter(f_worker_id != nil, f_target_job_id != nil, f_career_path_transition_id != nil, f_recorded_by_principal_id != nil, f_preferred_location_reference_id != nil)
	if err != nil {
		return nil, err
	}
	var resp *career.WorkerCareerAspirationsResponse
	switch idx {
	case 0:
		v, fErr := validateFilterString(f_worker_id, "worker_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.career.FindWorkerCareerAspirationByWorkerID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 1:
		v, fErr := validateFilterString(f_target_job_id, "target_job_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.career.FindWorkerCareerAspirationByTargetJobID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 2:
		v, fErr := validateFilterString(f_career_path_transition_id, "career_path_transition_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.career.FindWorkerCareerAspirationByCareerPathTransitionID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 3:
		v, fErr := validateFilterString(f_recorded_by_principal_id, "recorded_by_principal_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.career.FindWorkerCareerAspirationByRecordedByPrincipalID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 4:
		v, fErr := validateFilterString(f_preferred_location_reference_id, "preferred_location_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.career.FindWorkerCareerAspirationByPreferredLocationReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	default:
		resp, err = a.career.GetAllWorkerCareerAspirations(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "worker-career-aspiration")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "worker-career-aspiration", toModelWorkerCareerAspiration)
}

// CreateWorkerCareerAspiration membuat worker-career-aspiration baru melalui HRM.
func (a *Adapter) CreateWorkerCareerAspiration(ctx context.Context, input model.CreateWorkerCareerAspirationInput) (*model.WorkerCareerAspiration, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_worker_career_aspirations", "create", "human_resource.worker_career_aspirations.create")
	if err != nil {
		return nil, err
	}
	if a.career == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &career.WorkerCareerAspirationInput{}
	patch.WorkerId = input.WorkerID
	patch.TargetJobId = input.TargetJobID
	patch.CareerPathTransitionId = input.CareerPathTransitionID
	patch.PreferenceRank = optionalInt32Ptr(input.PreferenceRank)
	patch.TargetDate = optionalTimestamp(input.TargetDate)
	patch.MobilityPreference = input.MobilityPreference
	patch.PreferredLocationReferenceId = input.PreferredLocationReferenceID
	patch.WithdrawnAt = optionalTimestamp(input.WithdrawnAt)
	resp, err := a.career.CreateWorkerCareerAspiration(rpcCtx, &career.CreateWorkerCareerAspirationRequest{Input: patch})
	if err != nil {
		return nil, mapHRMWriteError(err, "worker-career-aspiration")
	}
	return toModelWorkerCareerAspiration(resp.GetValue())
}

// UpdateWorkerCareerAspiration memperbarui worker-career-aspiration berdasarkan FieldMask.
func (a *Adapter) UpdateWorkerCareerAspiration(ctx context.Context, rawID string, input model.UpdateWorkerCareerAspirationInput) (*model.WorkerCareerAspiration, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_worker_career_aspirations", "update", "human_resource.worker_career_aspirations.update")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "worker-career-aspiration", "worker-career-aspiration")
	if err != nil {
		return nil, err
	}
	if a.career == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &career.WorkerCareerAspirationInput{}
	mask := &fieldmaskpb.FieldMask{}
	if err := applyStringMask(mask, "worker_id", input.WorkerID, &patch.WorkerId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "target_job_id", input.TargetJobID, &patch.TargetJobId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "career_path_transition_id", input.CareerPathTransitionID, &patch.CareerPathTransitionId, true); err != nil {
		return nil, err
	}
	if err := applyInt32Mask(mask, "preference_rank", input.PreferenceRank, &patch.PreferenceRank, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "target_date", input.TargetDate, &patch.TargetDate, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "mobility_preference", input.MobilityPreference, &patch.MobilityPreference, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "preferred_location_reference_id", input.PreferredLocationReferenceID, &patch.PreferredLocationReferenceId, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "withdrawn_at", input.WithdrawnAt, &patch.WithdrawnAt, true); err != nil {
		return nil, err
	}
	if len(mask.Paths) == 0 {
		return nil, middleware.ErrInvalidArgument("worker-career-aspiration update is empty", nil)
	}
	resp, err := a.career.UpdateWorkerCareerAspiration(rpcCtx, &career.UpdateWorkerCareerAspirationRequest{Id: id, Input: patch, UpdateMask: mask})
	if err != nil {
		return nil, mapHRMWriteError(err, "worker-career-aspiration")
	}
	return toModelWorkerCareerAspiration(resp.GetValue())
}

// DeleteWorkerCareerAspiration menghapus worker-career-aspiration berdasarkan ID.
func (a *Adapter) DeleteWorkerCareerAspiration(ctx context.Context, rawID string) (*bool, error) {
	if a == nil || a.career == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	return deleteByUUID(ctx, a, rawID, "hr_worker_career_aspirations", "human_resource.worker_career_aspirations.delete", "worker-career-aspiration", func(c context.Context, req *common.IDRequest) (*common.EmptyResponse, error) {
		return a.career.DeleteWorkerCareerAspiration(c, req)
	})
}
