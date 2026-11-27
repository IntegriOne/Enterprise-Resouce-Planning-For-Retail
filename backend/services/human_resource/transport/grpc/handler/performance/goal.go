package performance

import (
	"context"

	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/performance/v1"
	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/performance"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/response"
)

func goal(value *entity.PerformanceGoal) (*pb.PerformanceGoal, error) {
	measure, err := response.JSONValue(value.MeasureDefinition)
	if err != nil {
		return nil, err
	}
	return &pb.PerformanceGoal{Id: value.ID, WorkerId: value.WorkerID, PerformanceCycleId: value.PerformanceCycleID, ParentGoalId: value.ParentGoalID, GoalCode: value.GoalCode, GoalTitle: value.GoalTitle, GoalDescription: value.GoalDescription, GoalType: value.GoalType, MeasureDefinition: measure, TargetValue: decimalString(value.TargetValue), ActualValue: decimalString(value.ActualValue), UnitCode: value.UnitCode, WeightPercent: decimalString(value.WeightPercent), DueDate: dateString(value.DueDate), GoalStatus: value.GoalStatus, ApprovedByPrincipalId: value.ApprovedByPrincipalID, ApprovedAt: timestamp(value.ApprovedAt)}, nil
}
func goalResponse(value *entity.PerformanceGoal, err error) (*pb.PerformanceGoalResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	item, err := goal(value)
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.PerformanceGoalResponse{Item: item}, nil
}
func goals(page operation.Page[entity.PerformanceGoal], err error) (*pb.PerformanceGoalsResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	items := make([]*pb.PerformanceGoal, 0, len(page.Items))
	for i := range page.Items {
		item, err := goal(&page.Items[i])
		if err != nil {
			return nil, response.Error(err)
		}
		items = append(items, item)
	}
	return &pb.PerformanceGoalsResponse{Items: items, Page: pageInfo(page)}, nil
}
func (h *Handler) CreatePerformanceGoal(ctx context.Context, req *pb.CreatePerformanceGoalRequest) (*pb.PerformanceGoalResponse, error) {
	if req == nil || req.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	in := req.Input
	measure, err := response.JSON(in.MeasureDefinition)
	if err != nil {
		return nil, response.Error(err)
	}
	due, err := datePtr(in.DueDate)
	if err != nil {
		return nil, response.Error(err)
	}
	approved, err := response.OptionalTimestamp(in.ApprovedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	return goalResponse(h.service.CreatePerformanceGoal(ctx, service.CreatePerformanceGoalInput{WorkerID: in.GetWorkerId(), PerformanceCycleID: in.GetPerformanceCycleId(), ParentGoalID: in.ParentGoalId, GoalCode: in.GoalCode, GoalTitle: in.GetGoalTitle(), GoalDescription: in.GoalDescription, GoalType: in.GetGoalType(), MeasureDefinition: measure, TargetValue: decimalPtr(in.TargetValue), ActualValue: decimalPtr(in.ActualValue), UnitCode: in.UnitCode, WeightPercent: decimalPtr(in.WeightPercent), DueDate: due, GoalStatus: in.GoalStatus, ApprovedByPrincipalID: in.ApprovedByPrincipalId, ApprovedAt: approved}))
}
func (h *Handler) UpdatePerformanceGoal(ctx context.Context, req *pb.UpdatePerformanceGoalRequest) (*pb.PerformanceGoalResponse, error) {
	if req == nil || req.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	m, err := response.ValidateMask(req.UpdateMask, "worker_id", "performance_cycle_id", "parent_goal_id", "goal_code", "goal_title", "goal_description", "goal_type", "measure_definition", "target_value", "actual_value", "unit_code", "weight_percent", "due_date", "goal_status", "approved_by_principal_id", "approved_at")
	if err != nil {
		return nil, response.Error(err)
	}
	in := req.Input
	measure, err := jsonField(m["measure_definition"], in.MeasureDefinition)
	if err != nil {
		return nil, response.Error(err)
	}
	due, err := dateField(m["due_date"], in.DueDate)
	if err != nil {
		return nil, response.Error(err)
	}
	approved, err := timeField(m["approved_at"], in.ApprovedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	return goalResponse(h.service.UpdatePerformanceGoal(ctx, req.Id, service.UpdatePerformanceGoalInput{WorkerID: field(m["worker_id"], in.WorkerId), PerformanceCycleID: field(m["performance_cycle_id"], in.PerformanceCycleId), ParentGoalID: field(m["parent_goal_id"], in.ParentGoalId), GoalCode: field(m["goal_code"], in.GoalCode), GoalTitle: field(m["goal_title"], in.GoalTitle), GoalDescription: field(m["goal_description"], in.GoalDescription), GoalType: field(m["goal_type"], in.GoalType), MeasureDefinition: measure, TargetValue: decimalField(m["target_value"], in.TargetValue), ActualValue: decimalField(m["actual_value"], in.ActualValue), UnitCode: field(m["unit_code"], in.UnitCode), WeightPercent: decimalField(m["weight_percent"], in.WeightPercent), DueDate: due, GoalStatus: field(m["goal_status"], in.GoalStatus), ApprovedByPrincipalID: field(m["approved_by_principal_id"], in.ApprovedByPrincipalId), ApprovedAt: approved}))
}
func (h *Handler) DeletePerformanceGoal(ctx context.Context, req *common.IDRequest) (*common.EmptyResponse, error) {
	return remove(ctx, req, h.service.DeletePerformanceGoal)
}
func (h *Handler) SubmitPerformanceGoal(ctx context.Context, req *common.IDRequest) (*pb.PerformanceGoalResponse, error) {
	if req == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return goalResponse(h.service.SubmitPerformanceGoal(ctx, req.Id))
}
func (h *Handler) ApprovePerformanceGoal(ctx context.Context, req *common.IDRequest) (*pb.PerformanceGoalResponse, error) {
	if req == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return goalResponse(h.service.ApprovePerformanceGoal(ctx, req.Id))
}
func (h *Handler) RejectPerformanceGoal(ctx context.Context, req *common.IDRequest) (*pb.PerformanceGoalResponse, error) {
	if req == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return goalResponse(h.service.RejectPerformanceGoal(ctx, req.Id))
}
func (h *Handler) GetPerformanceGoalByID(ctx context.Context, req *common.IDRequest) (*pb.PerformanceGoalResponse, error) {
	if req == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return goalResponse(h.service.GetPerformanceGoalByID(ctx, req.Id))
}
func (h *Handler) GetAllPerformanceGoals(ctx context.Context, req *common.PageRequest) (*pb.PerformanceGoalsResponse, error) {
	if req == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return goals(h.service.GetAllPerformanceGoals(ctx, int(req.Limit), int(req.Offset)))
}
func (h *Handler) FindPerformanceGoalByWorkerID(ctx context.Context, req *common.StringFinderRequest) (*pb.PerformanceGoalsResponse, error) {
	if req == nil || req.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return goals(h.service.FindPerformanceGoalByWorkerID(ctx, *req.Value, int(req.Limit), int(req.Offset)))
}
func (h *Handler) FindPerformanceGoalByPerformanceCycleID(ctx context.Context, req *common.StringFinderRequest) (*pb.PerformanceGoalsResponse, error) {
	if req == nil || req.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return goals(h.service.FindPerformanceGoalByPerformanceCycleID(ctx, *req.Value, int(req.Limit), int(req.Offset)))
}
func (h *Handler) FindPerformanceGoalByParentGoalID(ctx context.Context, req *common.StringFinderRequest) (*pb.PerformanceGoalsResponse, error) {
	if req == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return goals(h.service.FindPerformanceGoalByParentGoalID(ctx, req.Value, int(req.Limit), int(req.Offset)))
}
func (h *Handler) FindPerformanceGoalByGoalCode(ctx context.Context, req *common.StringFinderRequest) (*pb.PerformanceGoalsResponse, error) {
	if req == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return goals(h.service.FindPerformanceGoalByGoalCode(ctx, req.Value, int(req.Limit), int(req.Offset)))
}
func (h *Handler) FindPerformanceGoalByGoalType(ctx context.Context, req *common.StringFinderRequest) (*pb.PerformanceGoalsResponse, error) {
	if req == nil || req.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return goals(h.service.FindPerformanceGoalByGoalType(ctx, *req.Value, int(req.Limit), int(req.Offset)))
}
func (h *Handler) FindPerformanceGoalByUnitCode(ctx context.Context, req *common.StringFinderRequest) (*pb.PerformanceGoalsResponse, error) {
	if req == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return goals(h.service.FindPerformanceGoalByUnitCode(ctx, req.Value, int(req.Limit), int(req.Offset)))
}
func (h *Handler) FindPerformanceGoalByGoalStatus(ctx context.Context, req *common.StringFinderRequest) (*pb.PerformanceGoalsResponse, error) {
	if req == nil || req.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return goals(h.service.FindPerformanceGoalByGoalStatus(ctx, *req.Value, int(req.Limit), int(req.Offset)))
}
func (h *Handler) FindPerformanceGoalByApprovedByPrincipalID(ctx context.Context, req *common.StringFinderRequest) (*pb.PerformanceGoalsResponse, error) {
	if req == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return goals(h.service.FindPerformanceGoalByApprovedByPrincipalID(ctx, req.Value, int(req.Limit), int(req.Offset)))
}
