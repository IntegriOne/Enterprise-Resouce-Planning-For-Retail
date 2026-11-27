package human_resource

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	offboarding "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/offboarding/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	offboardingPlanReadPermission = "human_resource.offboarding_plans.read"
	offboardingTaskReadPermission = "human_resource.offboarding_tasks.read"
)

// OffboardingPlan membaca metadata rencana offboarding berdasarkan ID.
func (a *Adapter) OffboardingPlan(ctx context.Context, rawID string) (*model.OffboardingPlan, error) {
	rpcCtx, err := a.authorizeGlobalEmployee(ctx, offboardingPlanReadPermission)
	if err != nil {
		return nil, err
	}
	id, err := parseNodeID(rawID, "offboarding-plan")
	if err != nil || id == uuid.Nil {
		return nil, middleware.ErrInvalidArgument("invalid offboarding plan ID", nil)
	}
	if a.offboarding == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	response, err := a.offboarding.GetOffboardingPlanByID(rpcCtx, &common.IDRequest{Id: id.String()})
	if err != nil {
		return nil, mapReadError(err, "offboarding plan")
	}
	item := response.GetValue()
	if item.GetId() != id.String() {
		return nil, middleware.ErrUnavailable("invalid human resource offboarding plan response", nil)
	}
	return toModelOffboardingPlan(item)
}

// OffboardingPlans membaca halaman rencana offboarding untuk satu pekerja.
func (a *Adapter) OffboardingPlans(ctx context.Context, rawWorkerID string, limit, offset int) (*Page[model.OffboardingPlan], error) {
	rpcCtx, err := a.authorizeGlobalEmployee(ctx, offboardingPlanReadPermission)
	if err != nil {
		return nil, err
	}
	id, err := uuid.Parse(strings.TrimPrefix(rawWorkerID, workerGIDPrefix))
	if err != nil || id == uuid.Nil || !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid offboarding plans request", nil)
	}
	if a.offboarding == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	workerID := id.String()
	response, err := a.offboarding.FindOffboardingPlanByWorkerID(rpcCtx, &common.StringFinderRequest{
		Value: &workerID, Limit: int32(limit), Offset: int32(offset),
	})
	if err != nil {
		return nil, mapReadError(err, "offboarding plan")
	}
	items, page := response.GetItems(), response.GetPage()
	if !validPage(page, limit, offset, len(items)) {
		return nil, middleware.ErrUnavailable("invalid human resource offboarding plan page", nil)
	}
	result := &Page[model.OffboardingPlan]{
		Items: make([]*model.OffboardingPlan, 0, len(items)), TotalCount: int(page.GetTotal()), Limit: limit, Offset: offset,
	}
	for _, item := range items {
		if item.GetWorkerId() != workerID {
			return nil, middleware.ErrUnavailable("invalid human resource offboarding plan worker", nil)
		}
		plan, err := toModelOffboardingPlan(item)
		if err != nil {
			return nil, err
		}
		result.Items = append(result.Items, plan)
	}
	return result, nil
}

func toModelOffboardingPlan(item *offboarding.OffboardingPlan) (*model.OffboardingPlan, error) {
	for _, value := range []string{item.GetId(), item.GetWorkerId()} {
		if id, err := uuid.Parse(value); err != nil || id == uuid.Nil {
			return nil, middleware.ErrUnavailable("invalid human resource offboarding plan reference", nil)
		}
	}
	lastDay, effective := item.GetLastWorkingDate(), item.GetTerminationEffectiveDate()
	if lastDay.CheckValid() != nil || effective.CheckValid() != nil {
		return nil, middleware.ErrUnavailable("invalid human resource offboarding plan date", nil)
	}
	return &model.OffboardingPlan{
		ID: nodeID("offboarding-plan", item.GetId()), RawID: item.GetId(), WorkerID: workerGIDPrefix + item.GetWorkerId(), PlanNumber: item.GetPlanNumber(),
		Status: item.GetStatus(), LastWorkingDate: lastDay.AsTime(), TerminationEffectiveDate: effective.AsTime(),
	}, nil
}

// ApproveOffboardingPlan meneruskan persetujuan independen ke HRM.
func (a *Adapter) ApproveOffboardingPlan(ctx context.Context, id string) (*model.OffboardingPlan, error) {
	return a.runOffboardingPlanCommand(ctx, id, "human_resource.offboarding_plans.approve", "approved", offboarding.OffboardingServiceClient.ApproveOffboardingPlan)
}

// StartOffboardingPlan memulai rencana yang telah disetujui HRM.
func (a *Adapter) StartOffboardingPlan(ctx context.Context, id string) (*model.OffboardingPlan, error) {
	return a.runOffboardingPlanCommand(ctx, id, "human_resource.offboarding_plans.start", "in_progress", offboarding.OffboardingServiceClient.StartOffboardingPlan)
}

func (a *Adapter) runOffboardingPlanCommand(
	ctx context.Context,
	rawID, permission, expectedStatus string,
	command func(offboarding.OffboardingServiceClient, context.Context, *common.IDRequest, ...grpc.CallOption) (*offboarding.OffboardingPlanResponse, error),
) (*model.OffboardingPlan, error) {
	rpcCtx, err := a.authorizeSensitiveEmployee(ctx, permission)
	if err != nil {
		return nil, err
	}
	id, err := parseNodeID(rawID, "offboarding-plan")
	if err != nil || id == uuid.Nil {
		return nil, middleware.ErrInvalidArgument("invalid offboarding plan ID", nil)
	}
	if a.offboarding == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	response, err := command(a.offboarding, rpcCtx, &common.IDRequest{Id: id.String()})
	if status.Code(err) == codes.FailedPrecondition {
		return nil, middleware.ErrConflict("offboarding plan does not meet workflow requirements", err)
	}
	if err != nil {
		return nil, mapReadError(err, "offboarding plan")
	}
	item := response.GetValue()
	if item.GetId() != id.String() || item.GetStatus() != expectedStatus {
		return nil, middleware.ErrUnavailable("invalid human resource offboarding plan response", nil)
	}
	return toModelOffboardingPlan(item)
}

// CreateOffboardingPlan mencatat draf tanpa metadata keputusan atau clearance.
func (a *Adapter) CreateOffboardingPlan(ctx context.Context, input model.CreateOffboardingPlanInput) (*model.OffboardingPlan, error) {
	rpcCtx, err := a.authorizeGlobalEmployee(ctx, "human_resource.offboarding_plans.create")
	if err != nil {
		return nil, err
	}
	id, err := uuid.Parse(strings.TrimPrefix(input.WorkerID, workerGIDPrefix))
	if err != nil || id == uuid.Nil {
		return nil, middleware.ErrInvalidArgument("invalid worker ID", nil)
	}
	if a.offboarding == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	workerID := id.String()
	response, err := a.offboarding.CreateOffboardingPlan(rpcCtx, &offboarding.CreateOffboardingPlanRequest{
		Input: &offboarding.OffboardingPlanInput{
			WorkerId: &workerID, PlanNumber: &input.PlanNumber, ReasonCode: &input.ReasonCode,
			NoticeDate: timestamppb.New(input.NoticeDate), LastWorkingDate: timestamppb.New(input.LastWorkingDate),
			TerminationEffectiveDate: timestamppb.New(input.TerminationEffectiveDate),
		},
	})
	switch status.Code(err) {
	case codes.AlreadyExists, codes.FailedPrecondition:
		return nil, middleware.ErrConflict("offboarding plan conflicts with existing state", err)
	case codes.OK:
	default:
		return nil, mapReadError(err, "offboarding plan")
	}
	item := response.GetValue()
	if item.GetWorkerId() != workerID || item.GetStatus() != "draft" {
		return nil, middleware.ErrUnavailable("invalid human resource offboarding plan response", nil)
	}
	return toModelOffboardingPlan(item)
}

// CreateOffboardingTask menambahkan tugas pending pada rencana draf.
func (a *Adapter) CreateOffboardingTask(ctx context.Context, input model.CreateOffboardingTaskInput) (*model.OffboardingTask, error) {
	rpcCtx, err := a.authorizeGlobalEmployee(ctx, "human_resource.offboarding_tasks.create")
	if err != nil {
		return nil, err
	}
	id, err := parseNodeID(input.OffboardingPlanID, "offboarding-plan")
	if err != nil || id == uuid.Nil {
		return nil, middleware.ErrInvalidArgument("invalid offboarding plan ID", nil)
	}
	if a.offboarding == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	planID := id.String()
	patch := &offboarding.OffboardingTaskInput{
		OffboardingPlanId: &planID, TaskCode: &input.TaskCode, TaskName: &input.TaskName,
		OwnerManagement: &input.OwnerManagement, AssignedPrincipalId: input.AssignedPrincipalID, IsMandatory: input.IsMandatory,
	}
	if input.DueAt != nil {
		patch.DueAt = timestamppb.New(*input.DueAt)
	}
	response, err := a.offboarding.CreateOffboardingTask(rpcCtx, &offboarding.CreateOffboardingTaskRequest{Input: patch})
	switch status.Code(err) {
	case codes.AlreadyExists, codes.FailedPrecondition:
		return nil, middleware.ErrConflict("offboarding task conflicts with existing state", err)
	case codes.OK:
	default:
		return nil, mapReadError(err, "offboarding task")
	}
	item := response.GetValue()
	if item.GetOffboardingPlanId() != planID || item.GetStatus() != "pending" {
		return nil, middleware.ErrUnavailable("invalid human resource offboarding task response", nil)
	}
	return toModelOffboardingTask(item)
}

func toModelOffboardingTask(item *offboarding.OffboardingTask) (*model.OffboardingTask, error) {
	taskID, err := uuid.Parse(item.GetId())
	if err != nil || taskID == uuid.Nil {
		return nil, middleware.ErrUnavailable("invalid human resource offboarding task ID", nil)
	}
	result := &model.OffboardingTask{
		ID: nodeID("offboarding-task", item.GetId()), RawID: item.GetId(), OffboardingPlanID: item.GetOffboardingPlanId(), TaskCode: item.GetTaskCode(), TaskName: item.GetTaskName(),
		OwnerManagement: item.GetOwnerManagement(), Status: item.GetStatus(), IsMandatory: item.GetIsMandatory(),
	}
	if stamp := item.GetDueAt(); stamp != nil {
		if stamp.CheckValid() != nil {
			return nil, middleware.ErrUnavailable("invalid human resource offboarding task date", nil)
		}
		value := stamp.AsTime()
		result.DueAt = &value
	}
	return result, nil
}

// OffboardingTask membaca detail tugas offboarding berdasarkan ID.
func (a *Adapter) OffboardingTask(ctx context.Context, rawID string) (*model.OffboardingTask, error) {
	rpcCtx, err := a.authorizeGlobalEmployee(ctx, offboardingTaskReadPermission)
	if err != nil {
		return nil, err
	}
	id, err := parseNodeID(rawID, "offboarding-task")
	if err != nil || id == uuid.Nil {
		return nil, middleware.ErrInvalidArgument("invalid offboarding task ID", nil)
	}
	if a.offboarding == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	response, err := a.offboarding.GetOffboardingTaskByID(rpcCtx, &common.IDRequest{Id: id.String()})
	if err != nil {
		return nil, mapReadError(err, "offboarding task")
	}
	item := response.GetValue()
	if item.GetId() != id.String() {
		return nil, middleware.ErrUnavailable("invalid human resource offboarding task response", nil)
	}
	planID, err := uuid.Parse(item.GetOffboardingPlanId())
	if err != nil || planID == uuid.Nil {
		return nil, middleware.ErrUnavailable("invalid human resource offboarding task plan", nil)
	}
	return toModelOffboardingTask(item)
}

// OffboardingTasks membaca halaman tugas yang terkait dengan satu rencana.
func (a *Adapter) OffboardingTasks(ctx context.Context, rawPlanID string, limit, offset int) (*Page[model.OffboardingTask], error) {
	rpcCtx, err := a.authorizeGlobalEmployee(ctx, offboardingTaskReadPermission)
	if err != nil {
		return nil, err
	}
	id, err := parseNodeID(rawPlanID, "offboarding-plan")
	if err != nil || id == uuid.Nil || !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid offboarding tasks request", nil)
	}
	if a.offboarding == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	planID := id.String()
	response, err := a.offboarding.FindOffboardingTaskByOffboardingPlanID(rpcCtx, &common.StringFinderRequest{
		Value: &planID, Limit: int32(limit), Offset: int32(offset),
	})
	if err != nil {
		return nil, mapReadError(err, "offboarding task")
	}
	items, page := response.GetItems(), response.GetPage()
	if !validPage(page, limit, offset, len(items)) {
		return nil, middleware.ErrUnavailable("invalid human resource offboarding task page", nil)
	}
	result := &Page[model.OffboardingTask]{
		Items: make([]*model.OffboardingTask, 0, len(items)), TotalCount: int(page.GetTotal()), Limit: limit, Offset: offset,
	}
	for _, item := range items {
		if item.GetOffboardingPlanId() != planID {
			return nil, middleware.ErrUnavailable("invalid human resource offboarding task plan", nil)
		}
		task, err := toModelOffboardingTask(item)
		if err != nil {
			return nil, err
		}
		result.Items = append(result.Items, task)
	}
	return result, nil
}

// CompleteOffboardingTask meneruskan bukti penyelesaian untuk diverifikasi HRM.
func (a *Adapter) CompleteOffboardingTask(ctx context.Context, id, rawConfirmationID string) (*model.OffboardingTask, error) {
	return a.runOffboardingTaskCommand(ctx, id, "human_resource.offboarding_tasks.complete", "completed",
		func(rpcCtx context.Context, taskID string) (*offboarding.OffboardingTaskResponse, error) {
			confirmationID, err := uuid.Parse(rawConfirmationID)
			if err != nil || confirmationID == uuid.Nil {
				return nil, status.Error(codes.InvalidArgument, "invalid offboarding confirmation ID")
			}
			response, err := a.offboarding.CompleteOffboardingTask(rpcCtx, &offboarding.CompleteOffboardingTaskRequest{
				Id: taskID, OwnerConfirmationReferenceId: confirmationID.String(),
			})
			if err == nil && response.GetValue().GetOwnerConfirmationReferenceId() != confirmationID.String() {
				return nil, status.Error(codes.Internal, "invalid offboarding task confirmation response")
			}
			return response, err
		})
}

// WaiveOffboardingTask meneruskan alasan pengecualian tugas ke HRM.
func (a *Adapter) WaiveOffboardingTask(ctx context.Context, id, reason string) (*model.OffboardingTask, error) {
	return a.runOffboardingTaskCommand(ctx, id, "human_resource.offboarding_tasks.waive", "waived",
		func(rpcCtx context.Context, taskID string) (*offboarding.OffboardingTaskResponse, error) {
			return a.offboarding.WaiveOffboardingTask(rpcCtx, &offboarding.WaiveOffboardingTaskRequest{Id: taskID, Reason: reason})
		})
}

func (a *Adapter) runOffboardingTaskCommand(
	ctx context.Context, rawID, permission, expectedStatus string,
	command func(context.Context, string) (*offboarding.OffboardingTaskResponse, error),
) (*model.OffboardingTask, error) {
	rpcCtx, err := a.authorizeSensitiveEmployee(ctx, permission)
	if err != nil {
		return nil, err
	}
	id, err := parseNodeID(rawID, "offboarding-task")
	if err != nil || id == uuid.Nil {
		return nil, middleware.ErrInvalidArgument("invalid offboarding task ID", nil)
	}
	if a.offboarding == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	response, err := command(rpcCtx, id.String())
	if status.Code(err) == codes.FailedPrecondition {
		return nil, middleware.ErrConflict("offboarding task does not meet workflow requirements", err)
	}
	if err != nil {
		return nil, mapReadError(err, "offboarding task")
	}
	item := response.GetValue()
	if item.GetId() != id.String() || item.GetStatus() != expectedStatus {
		return nil, middleware.ErrUnavailable("invalid human resource offboarding task response", nil)
	}
	if planID, err := uuid.Parse(item.GetOffboardingPlanId()); err != nil || planID == uuid.Nil {
		return nil, middleware.ErrUnavailable("invalid human resource offboarding plan ID", nil)
	}
	return toModelOffboardingTask(item)
}

// CompleteOffboardingPlan meneruskan penutupan setelah payroll final dan clearance diverifikasi HRM.
func (a *Adapter) CompleteOffboardingPlan(ctx context.Context, id, rawRunID, rawClearanceID string) (*model.OffboardingPlan, error) {
	return a.runOffboardingPlanCommand(ctx, id, "human_resource.offboarding_plans.complete", "completed",
		func(client offboarding.OffboardingServiceClient, rpcCtx context.Context, request *common.IDRequest, opts ...grpc.CallOption) (*offboarding.OffboardingPlanResponse, error) {
			runID, runErr := parseNodeID(rawRunID, "payroll-run")
			clearanceID, clearanceErr := uuid.Parse(rawClearanceID)
			if runErr != nil || runID == uuid.Nil || clearanceErr != nil || clearanceID == uuid.Nil {
				return nil, status.Error(codes.InvalidArgument, "invalid offboarding clearance references")
			}
			response, err := client.CompleteOffboardingPlan(rpcCtx, &offboarding.CompleteOffboardingPlanRequest{
				Id: request.Id, FinalPayrollRunId: runID.String(), FinanceClearanceReferenceId: clearanceID.String(),
			}, opts...)
			item := response.GetValue()
			if err == nil && (item.GetFinalPayrollRunId() != runID.String() ||
				item.GetFinanceClearanceReferenceId() != clearanceID.String() || item.GetCompletedAt().CheckValid() != nil) {
				return nil, status.Error(codes.Internal, "invalid offboarding clearance response")
			}
			return response, err
		})
}

// CancelOffboardingPlan membatalkan rencana yang disetujui dan belum dimulai.
func (a *Adapter) CancelOffboardingPlan(ctx context.Context, id, reason string) (*model.OffboardingPlan, error) {
	return a.runOffboardingPlanCommand(ctx, id, "human_resource.offboarding_plans.cancel", "cancelled",
		func(client offboarding.OffboardingServiceClient, rpcCtx context.Context, request *common.IDRequest, opts ...grpc.CallOption) (*offboarding.OffboardingPlanResponse, error) {
			response, err := client.CancelOffboardingPlan(rpcCtx, &offboarding.CancelOffboardingPlanRequest{
				Id: request.Id, Reason: reason,
			}, opts...)
			item := response.GetValue()
			if err == nil && (item.GetCancelledAt().CheckValid() != nil || item.GetCancellationReason() != strings.TrimSpace(reason)) {
				return nil, status.Error(codes.Internal, "invalid offboarding cancellation response")
			}
			return response, err
		})
}
