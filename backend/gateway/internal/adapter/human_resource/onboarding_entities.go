package human_resource

import (
	"context"

	"github.com/google/uuid"
	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	onboarding "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/onboarding/v1"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func toModelOnboardingPlan(row *onboarding.OnboardingPlan) (*model.OnboardingPlan, error) {
	if row == nil {
		return nil, middleware.ErrUnavailable("invalid human resource onboarding-plan response", nil)
	}
	if id, err := uuid.Parse(row.GetId()); err != nil || id == uuid.Nil {
		return nil, middleware.ErrUnavailable("invalid human resource onboarding-plan ID", nil)
	}
	out := &model.OnboardingPlan{ID: nodeID("onboarding-plan", row.GetId()), RawID: row.GetId()}
	out.WorkerID = row.GetWorkerId()
	out.PlanCode = row.GetPlanCode()
	out.PlanVersion = int(row.GetPlanVersion())
	out.StartDate = row.GetStartDate()
	out.TargetCompletionDate = row.TargetCompletionDate
	out.Status = row.GetStatus()
	out.TemplateReferenceID = row.TemplateReferenceId
	out.CreatedByPrincipalID = row.CreatedByPrincipalId
	out.StartedAt = pbTimePtr(row.GetStartedAt())
	out.StartedByPrincipalID = row.StartedByPrincipalId
	out.CompletedAt = pbTimePtr(row.GetCompletedAt())
	out.CompletedByPrincipalID = row.CompletedByPrincipalId
	return out, nil
}

// OnboardingPlan membaca detail onboarding-plan berdasarkan ID.
func (a *Adapter) OnboardingPlan(ctx context.Context, rawID string) (*model.OnboardingPlan, error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_onboarding_plans", "human_resource.onboarding_plans.read")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "onboarding-plan", "onboarding-plan")
	if err != nil {
		return nil, err
	}
	if a.onboarding == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.onboarding.GetOnboardingPlanByID(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapReadError(err, "onboarding-plan")
	}
	return toModelOnboardingPlan(resp.GetItem())
}

// OnboardingPlans mendaftar dan memfilter onboarding-plan dengan paginasi.
func (a *Adapter) OnboardingPlans(ctx context.Context, limit, offset int, f_worker_id *string, f_plan_code *string, f_status *string, f_template_reference_id *string, f_created_by_principal_id *string) (*Page[model.OnboardingPlan], error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_onboarding_plans", "human_resource.onboarding_plans.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid onboarding-plan pagination", nil)
	}
	if a.onboarding == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	idx, err := pickSingleFilter(f_worker_id != nil, f_plan_code != nil, f_status != nil, f_template_reference_id != nil, f_created_by_principal_id != nil)
	if err != nil {
		return nil, err
	}
	var resp *onboarding.OnboardingPlansResponse
	switch idx {
	case 0:
		v, fErr := validateFilterString(f_worker_id, "worker_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.onboarding.FindOnboardingPlanByWorkerID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 1:
		v, fErr := validateFilterString(f_plan_code, "plan_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.onboarding.FindOnboardingPlanByPlanCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 2:
		v, fErr := validateFilterString(f_status, "status")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.onboarding.FindOnboardingPlanByStatus(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 3:
		v, fErr := validateFilterString(f_template_reference_id, "template_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.onboarding.FindOnboardingPlanByTemplateReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 4:
		v, fErr := validateFilterString(f_created_by_principal_id, "created_by_principal_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.onboarding.FindOnboardingPlanByCreatedByPrincipalID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	default:
		resp, err = a.onboarding.GetAllOnboardingPlans(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "onboarding-plan")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "onboarding-plan", toModelOnboardingPlan)
}

// CreateOnboardingPlan membuat onboarding-plan baru melalui HRM.
func (a *Adapter) CreateOnboardingPlan(ctx context.Context, input model.CreateOnboardingPlanInput) (*model.OnboardingPlan, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_onboarding_plans", "create", "human_resource.onboarding_plans.create")
	if err != nil {
		return nil, err
	}
	if a.onboarding == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &onboarding.OnboardingPlanInput{}
	patch.WorkerId = input.WorkerID
	patch.PlanCode = input.PlanCode
	patch.PlanVersion = optionalInt32Ptr(input.PlanVersion)
	patch.StartDate = input.StartDate
	patch.TargetCompletionDate = input.TargetCompletionDate
	patch.Status = input.Status
	patch.TemplateReferenceId = input.TemplateReferenceID
	resp, err := a.onboarding.CreateOnboardingPlan(rpcCtx, &onboarding.CreateOnboardingPlanRequest{Input: patch})
	if err != nil {
		return nil, mapHRMWriteError(err, "onboarding-plan")
	}
	return toModelOnboardingPlan(resp.GetItem())
}

// UpdateOnboardingPlan memperbarui onboarding-plan berdasarkan FieldMask.
func (a *Adapter) UpdateOnboardingPlan(ctx context.Context, rawID string, input model.UpdateOnboardingPlanInput) (*model.OnboardingPlan, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_onboarding_plans", "update", "human_resource.onboarding_plans.update")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "onboarding-plan", "onboarding-plan")
	if err != nil {
		return nil, err
	}
	if a.onboarding == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &onboarding.OnboardingPlanInput{}
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
	if err := applyStringMask(mask, "start_date", input.StartDate, &patch.StartDate, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "target_completion_date", input.TargetCompletionDate, &patch.TargetCompletionDate, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "status", input.Status, &patch.Status, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "template_reference_id", input.TemplateReferenceID, &patch.TemplateReferenceId, true); err != nil {
		return nil, err
	}
	if len(mask.Paths) == 0 {
		return nil, middleware.ErrInvalidArgument("onboarding-plan update is empty", nil)
	}
	resp, err := a.onboarding.UpdateOnboardingPlan(rpcCtx, &onboarding.UpdateOnboardingPlanRequest{Id: id, Input: patch, UpdateMask: mask})
	if err != nil {
		return nil, mapHRMWriteError(err, "onboarding-plan")
	}
	return toModelOnboardingPlan(resp.GetItem())
}

// DeleteOnboardingPlan menghapus onboarding-plan berdasarkan ID.
func (a *Adapter) DeleteOnboardingPlan(ctx context.Context, rawID string) (*bool, error) {
	if a == nil || a.onboarding == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	return deleteByUUID(ctx, a, rawID, "hr_onboarding_plans", "human_resource.onboarding_plans.delete", "onboarding-plan", func(c context.Context, req *common.IDRequest) (*common.EmptyResponse, error) {
		return a.onboarding.DeleteOnboardingPlan(c, req)
	})
}

func toModelOnboardingTask(row *onboarding.OnboardingTask) (*model.OnboardingTask, error) {
	if row == nil {
		return nil, middleware.ErrUnavailable("invalid human resource onboarding-task response", nil)
	}
	if id, err := uuid.Parse(row.GetId()); err != nil || id == uuid.Nil {
		return nil, middleware.ErrUnavailable("invalid human resource onboarding-task ID", nil)
	}
	out := &model.OnboardingTask{ID: nodeID("onboarding-task", row.GetId()), RawID: row.GetId()}
	out.OnboardingPlanID = row.GetOnboardingPlanId()
	out.TaskCode = row.GetTaskCode()
	out.TaskName = row.GetTaskName()
	out.TaskOwnerType = row.GetTaskOwnerType()
	out.AssignedPrincipalID = row.AssignedPrincipalId
	out.DueAt = pbTimePtr(row.GetDueAt())
	out.CompletedAt = pbTimePtr(row.GetCompletedAt())
	out.CompletionEvidenceReferenceID = row.CompletionEvidenceReferenceId
	out.IsMandatory = row.GetIsMandatory()
	out.CompletedByPrincipalID = row.CompletedByPrincipalId
	return out, nil
}

// OnboardingTask membaca detail onboarding-task berdasarkan ID.
func (a *Adapter) OnboardingTask(ctx context.Context, rawID string) (*model.OnboardingTask, error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_onboarding_tasks", "human_resource.onboarding_tasks.read")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "onboarding-task", "onboarding-task")
	if err != nil {
		return nil, err
	}
	if a.onboarding == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.onboarding.GetOnboardingTaskByID(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapReadError(err, "onboarding-task")
	}
	return toModelOnboardingTask(resp.GetItem())
}

// OnboardingTasks mendaftar dan memfilter onboarding-task dengan paginasi.
func (a *Adapter) OnboardingTasks(ctx context.Context, limit, offset int, f_onboarding_plan_id *string, f_task_code *string, f_task_owner_type *string, f_assigned_principal_id *string, f_completion_evidence_reference_id *string) (*Page[model.OnboardingTask], error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_onboarding_tasks", "human_resource.onboarding_tasks.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid onboarding-task pagination", nil)
	}
	if a.onboarding == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	idx, err := pickSingleFilter(f_onboarding_plan_id != nil, f_task_code != nil, f_task_owner_type != nil, f_assigned_principal_id != nil, f_completion_evidence_reference_id != nil)
	if err != nil {
		return nil, err
	}
	var resp *onboarding.OnboardingTasksResponse
	switch idx {
	case 0:
		v, fErr := validateFilterString(f_onboarding_plan_id, "onboarding_plan_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.onboarding.FindOnboardingTaskByOnboardingPlanID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 1:
		v, fErr := validateFilterString(f_task_code, "task_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.onboarding.FindOnboardingTaskByTaskCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 2:
		v, fErr := validateFilterString(f_task_owner_type, "task_owner_type")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.onboarding.FindOnboardingTaskByTaskOwnerType(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 3:
		v, fErr := validateFilterString(f_assigned_principal_id, "assigned_principal_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.onboarding.FindOnboardingTaskByAssignedPrincipalID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 4:
		v, fErr := validateFilterString(f_completion_evidence_reference_id, "completion_evidence_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.onboarding.FindOnboardingTaskByCompletionEvidenceReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	default:
		resp, err = a.onboarding.GetAllOnboardingTasks(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "onboarding-task")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "onboarding-task", toModelOnboardingTask)
}

// CreateOnboardingTask membuat onboarding-task baru melalui HRM.
func (a *Adapter) CreateOnboardingTask(ctx context.Context, input model.CreateOnboardingTaskInput) (*model.OnboardingTask, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_onboarding_tasks", "create", "human_resource.onboarding_tasks.create")
	if err != nil {
		return nil, err
	}
	if a.onboarding == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &onboarding.OnboardingTaskInput{}
	patch.OnboardingPlanId = input.OnboardingPlanID
	patch.TaskCode = input.TaskCode
	patch.TaskName = input.TaskName
	patch.TaskOwnerType = input.TaskOwnerType
	patch.AssignedPrincipalId = input.AssignedPrincipalID
	patch.DueAt = optionalTimestamp(input.DueAt)
	patch.CompletedAt = optionalTimestamp(input.CompletedAt)
	patch.CompletionEvidenceReferenceId = input.CompletionEvidenceReferenceID
	patch.IsMandatory = input.IsMandatory
	resp, err := a.onboarding.CreateOnboardingTask(rpcCtx, &onboarding.CreateOnboardingTaskRequest{Input: patch})
	if err != nil {
		return nil, mapHRMWriteError(err, "onboarding-task")
	}
	return toModelOnboardingTask(resp.GetItem())
}

// UpdateOnboardingTask memperbarui onboarding-task berdasarkan FieldMask.
func (a *Adapter) UpdateOnboardingTask(ctx context.Context, rawID string, input model.UpdateOnboardingTaskInput) (*model.OnboardingTask, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_onboarding_tasks", "update", "human_resource.onboarding_tasks.update")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "onboarding-task", "onboarding-task")
	if err != nil {
		return nil, err
	}
	if a.onboarding == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &onboarding.OnboardingTaskInput{}
	mask := &fieldmaskpb.FieldMask{}
	if err := applyStringMask(mask, "onboarding_plan_id", input.OnboardingPlanID, &patch.OnboardingPlanId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "task_code", input.TaskCode, &patch.TaskCode, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "task_name", input.TaskName, &patch.TaskName, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "task_owner_type", input.TaskOwnerType, &patch.TaskOwnerType, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "assigned_principal_id", input.AssignedPrincipalID, &patch.AssignedPrincipalId, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "due_at", input.DueAt, &patch.DueAt, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "completed_at", input.CompletedAt, &patch.CompletedAt, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "completion_evidence_reference_id", input.CompletionEvidenceReferenceID, &patch.CompletionEvidenceReferenceId, true); err != nil {
		return nil, err
	}
	if err := applyBoolMask(mask, "is_mandatory", input.IsMandatory, &patch.IsMandatory, true); err != nil {
		return nil, err
	}
	if len(mask.Paths) == 0 {
		return nil, middleware.ErrInvalidArgument("onboarding-task update is empty", nil)
	}
	resp, err := a.onboarding.UpdateOnboardingTask(rpcCtx, &onboarding.UpdateOnboardingTaskRequest{Id: id, Input: patch, UpdateMask: mask})
	if err != nil {
		return nil, mapHRMWriteError(err, "onboarding-task")
	}
	return toModelOnboardingTask(resp.GetItem())
}

// DeleteOnboardingTask menghapus onboarding-task berdasarkan ID.
func (a *Adapter) DeleteOnboardingTask(ctx context.Context, rawID string) (*bool, error) {
	if a == nil || a.onboarding == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	return deleteByUUID(ctx, a, rawID, "hr_onboarding_tasks", "human_resource.onboarding_tasks.delete", "onboarding-task", func(c context.Context, req *common.IDRequest) (*common.EmptyResponse, error) {
		return a.onboarding.DeleteOnboardingTask(c, req)
	})
}
