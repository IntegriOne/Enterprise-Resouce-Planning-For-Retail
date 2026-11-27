package human_resource

import (
	"context"

	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	offboarding "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/offboarding/v1"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// UpdateOffboardingPlan memperbarui rencana offboarding berdasarkan FieldMask.
func (a *Adapter) UpdateOffboardingPlan(ctx context.Context, rawID string, input model.UpdateOffboardingPlanInput) (*model.OffboardingPlan, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_offboarding_plans", "update", "human_resource.offboarding_plans.update")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "offboarding-plan", "offboarding-plan")
	if err != nil {
		return nil, err
	}
	if a.offboarding == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &offboarding.OffboardingPlanInput{}
	mask := &fieldmaskpb.FieldMask{}
	if err := applyStringMask(mask, "worker_id", input.WorkerID, &patch.WorkerId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "plan_number", input.PlanNumber, &patch.PlanNumber, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "reason_code", input.ReasonCode, &patch.ReasonCode, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "notice_date", input.NoticeDate, &patch.NoticeDate, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "last_working_date", input.LastWorkingDate, &patch.LastWorkingDate, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "termination_effective_date", input.TerminationEffectiveDate, &patch.TerminationEffectiveDate, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "status", input.Status, &patch.Status, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "exit_interview_document_version_reference_id", input.ExitInterviewDocumentVersionReferenceID, &patch.ExitInterviewDocumentVersionReferenceId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "final_payroll_run_id", input.FinalPayrollRunID, &patch.FinalPayrollRunId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "finance_clearance_reference_id", input.FinanceClearanceReferenceID, &patch.FinanceClearanceReferenceId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "approved_by_principal_id", input.ApprovedByPrincipalID, &patch.ApprovedByPrincipalId, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "approved_at", input.ApprovedAt, &patch.ApprovedAt, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "completed_at", input.CompletedAt, &patch.CompletedAt, true); err != nil {
		return nil, err
	}
	if len(mask.Paths) == 0 {
		return nil, middleware.ErrInvalidArgument("offboarding-plan update is empty", nil)
	}
	resp, err := a.offboarding.UpdateOffboardingPlan(rpcCtx, &offboarding.UpdateOffboardingPlanRequest{
		Id:         id,
		Input:      patch,
		UpdateMask: mask,
	})
	if err != nil {
		return nil, mapHRMWriteError(err, "offboarding-plan")
	}
	return toModelOffboardingPlan(resp.GetValue())
}

// DeleteOffboardingPlan menghapus rencana offboarding berdasarkan ID.
func (a *Adapter) DeleteOffboardingPlan(ctx context.Context, rawID string) (*bool, error) {
	if a == nil || a.offboarding == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	return deleteByUUID(ctx, a, rawID, "hr_offboarding_plans", "human_resource.offboarding_plans.delete", "offboarding-plan", func(c context.Context, req *common.IDRequest) (*common.EmptyResponse, error) {
		return a.offboarding.DeleteOffboardingPlan(c, req)
	})
}

// UpdateOffboardingTask memperbarui tugas offboarding berdasarkan FieldMask.
func (a *Adapter) UpdateOffboardingTask(ctx context.Context, rawID string, input model.UpdateOffboardingTaskInput) (*model.OffboardingTask, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_offboarding_tasks", "update", "human_resource.offboarding_tasks.update")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "offboarding-task", "offboarding-task")
	if err != nil {
		return nil, err
	}
	if a.offboarding == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &offboarding.OffboardingTaskInput{}
	mask := &fieldmaskpb.FieldMask{}
	if err := applyStringMask(mask, "offboarding_plan_id", input.OffboardingPlanID, &patch.OffboardingPlanId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "task_code", input.TaskCode, &patch.TaskCode, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "task_name", input.TaskName, &patch.TaskName, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "owner_management", input.OwnerManagement, &patch.OwnerManagement, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "assigned_principal_id", input.AssignedPrincipalID, &patch.AssignedPrincipalId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "status", input.Status, &patch.Status, true); err != nil {
		return nil, err
	}
	if err := applyBoolMask(mask, "is_mandatory", input.IsMandatory, &patch.IsMandatory, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "due_at", input.DueAt, &patch.DueAt, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "completed_at", input.CompletedAt, &patch.CompletedAt, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "owner_confirmation_reference_id", input.OwnerConfirmationReferenceID, &patch.OwnerConfirmationReferenceId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "waiver_reason", input.WaiverReason, &patch.WaiverReason, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "waived_by_principal_id", input.WaivedByPrincipalID, &patch.WaivedByPrincipalId, true); err != nil {
		return nil, err
	}
	if len(mask.Paths) == 0 {
		return nil, middleware.ErrInvalidArgument("offboarding-task update is empty", nil)
	}
	resp, err := a.offboarding.UpdateOffboardingTask(rpcCtx, &offboarding.UpdateOffboardingTaskRequest{
		Id:         id,
		Input:      patch,
		UpdateMask: mask,
	})
	if err != nil {
		return nil, mapHRMWriteError(err, "offboarding-task")
	}
	return toModelOffboardingTask(resp.GetValue())
}

// DeleteOffboardingTask menghapus tugas offboarding berdasarkan ID.
func (a *Adapter) DeleteOffboardingTask(ctx context.Context, rawID string) (*bool, error) {
	if a == nil || a.offboarding == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	return deleteByUUID(ctx, a, rawID, "hr_offboarding_tasks", "human_resource.offboarding_tasks.delete", "offboarding-task", func(c context.Context, req *common.IDRequest) (*common.EmptyResponse, error) {
		return a.offboarding.DeleteOffboardingTask(c, req)
	})
}
