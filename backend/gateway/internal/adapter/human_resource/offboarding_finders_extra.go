package human_resource

import (
	"context"

	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	offboarding "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/offboarding/v1"
)

// OffboardingPlansFiltered mendaftar rencana offboarding dengan filter konsolidasi.
func (a *Adapter) OffboardingPlansFiltered(ctx context.Context, limit, offset int, workerID, planNumber, reasonCode, status, exitInterviewDocVerRefID, finalPayrollRunID, financeClearanceRefID, createdByPrincipalID, approvedByPrincipalID *string) (*Page[model.OffboardingPlan], error) {
	if workerID != nil && planNumber == nil && reasonCode == nil && status == nil && exitInterviewDocVerRefID == nil && finalPayrollRunID == nil && financeClearanceRefID == nil && createdByPrincipalID == nil && approvedByPrincipalID == nil {
		return a.OffboardingPlans(ctx, *workerID, limit, offset)
	}
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_offboarding_plans", "human_resource.offboarding_plans.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid offboarding-plan pagination", nil)
	}
	if a.offboarding == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	idx, err := pickSingleFilter(
		workerID != nil,
		planNumber != nil,
		reasonCode != nil,
		status != nil,
		exitInterviewDocVerRefID != nil,
		finalPayrollRunID != nil,
		financeClearanceRefID != nil,
		createdByPrincipalID != nil,
		approvedByPrincipalID != nil,
	)
	if err != nil {
		return nil, err
	}
	var resp *offboarding.OffboardingPlansResponse
	switch idx {
	case 0:
		v, fErr := validateUUIDFilter(workerID, "worker", "worker_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.offboarding.FindOffboardingPlanByWorkerID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 1:
		v, fErr := validateFilterString(planNumber, "plan_number")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.offboarding.FindOffboardingPlanByPlanNumber(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 2:
		v, fErr := validateFilterString(reasonCode, "reason_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.offboarding.FindOffboardingPlanByReasonCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 3:
		v, fErr := validateFilterString(status, "status")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.offboarding.FindOffboardingPlanByStatus(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 4:
		v, fErr := validateUUIDFilter(exitInterviewDocVerRefID, "document-version", "exit_interview_document_version_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.offboarding.FindOffboardingPlanByExitInterviewDocumentVersionReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 5:
		v, fErr := validateUUIDFilter(finalPayrollRunID, "payroll-run", "final_payroll_run_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.offboarding.FindOffboardingPlanByFinalPayrollRunID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 6:
		v, fErr := validateUUIDFilter(financeClearanceRefID, "finance-clearance", "finance_clearance_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.offboarding.FindOffboardingPlanByFinanceClearanceReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 7:
		v, fErr := validateUUIDFilter(createdByPrincipalID, "principal", "created_by_principal_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.offboarding.FindOffboardingPlanByCreatedByPrincipalID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 8:
		v, fErr := validateUUIDFilter(approvedByPrincipalID, "principal", "approved_by_principal_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.offboarding.FindOffboardingPlanByApprovedByPrincipalID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	default:
		resp, err = a.offboarding.GetAllOffboardingPlans(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "offboarding-plan")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "offboarding-plan", toModelOffboardingPlan)
}

// OffboardingTasksFiltered mendaftar tugas offboarding dengan filter konsolidasi.
func (a *Adapter) OffboardingTasksFiltered(ctx context.Context, limit, offset int, offboardingPlanID, taskCode, assignedPrincipalID, status, ownerConfirmationRefID, waivedByPrincipalID *string) (*Page[model.OffboardingTask], error) {
	if offboardingPlanID != nil && taskCode == nil && assignedPrincipalID == nil && status == nil && ownerConfirmationRefID == nil && waivedByPrincipalID == nil {
		return a.OffboardingTasks(ctx, *offboardingPlanID, limit, offset)
	}
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_offboarding_tasks", "human_resource.offboarding_tasks.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid offboarding-task pagination", nil)
	}
	if a.offboarding == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	idx, err := pickSingleFilter(
		offboardingPlanID != nil,
		taskCode != nil,
		assignedPrincipalID != nil,
		status != nil,
		ownerConfirmationRefID != nil,
		waivedByPrincipalID != nil,
	)
	if err != nil {
		return nil, err
	}
	var resp *offboarding.OffboardingTasksResponse
	switch idx {
	case 0:
		v, fErr := validateUUIDFilter(offboardingPlanID, "offboarding-plan", "offboarding_plan_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.offboarding.FindOffboardingTaskByOffboardingPlanID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 1:
		v, fErr := validateFilterString(taskCode, "task_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.offboarding.FindOffboardingTaskByTaskCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 2:
		v, fErr := validateUUIDFilter(assignedPrincipalID, "principal", "assigned_principal_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.offboarding.FindOffboardingTaskByAssignedPrincipalID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 3:
		v, fErr := validateFilterString(status, "status")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.offboarding.FindOffboardingTaskByStatus(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 4:
		v, fErr := validateUUIDFilter(ownerConfirmationRefID, "confirmation", "owner_confirmation_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.offboarding.FindOffboardingTaskByOwnerConfirmationReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 5:
		v, fErr := validateUUIDFilter(waivedByPrincipalID, "principal", "waived_by_principal_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.offboarding.FindOffboardingTaskByWaivedByPrincipalID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	default:
		resp, err = a.offboarding.GetAllOffboardingTasks(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "offboarding-task")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "offboarding-task", toModelOffboardingTask)
}
