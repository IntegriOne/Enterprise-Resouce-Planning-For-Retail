package performance

import (
	"context"
	"time"

	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/performance/v1"
	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/performance"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/response"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func cycle(value *entity.PerformanceCycle) (*pb.PerformanceCycle, error) {
	configuration, err := response.JSONValue(value.Configuration)
	if err != nil {
		return nil, err
	}
	return &pb.PerformanceCycle{Id: value.ID, CycleCode: value.CycleCode, CycleName: value.CycleName, CycleType: value.CycleType, Status: string(value.Status), PeriodStart: value.PeriodStart.Format("2006-01-02"), PeriodEnd: value.PeriodEnd.Format("2006-01-02"), SelfReviewDueAt: timestamp(value.SelfReviewDueAt), ManagerReviewDueAt: timestamp(value.ManagerReviewDueAt), CalibrationDueAt: timestamp(value.CalibrationDueAt), LaunchedAt: timestamp(value.LaunchedAt), LaunchedByPrincipalId: value.LaunchedByPrincipalID, Configuration: configuration}, nil
}
func timestamp(value *time.Time) *timestamppb.Timestamp {
	if value == nil {
		return nil
	}
	return timestamppb.New(*value)
}
func cycleResponse(value *entity.PerformanceCycle, err error) (*pb.PerformanceCycleResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	item, err := cycle(value)
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.PerformanceCycleResponse{Item: item}, nil
}
func cycles(page operation.Page[entity.PerformanceCycle], err error) (*pb.PerformanceCyclesResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	items := make([]*pb.PerformanceCycle, 0, len(page.Items))
	for i := range page.Items {
		item, err := cycle(&page.Items[i])
		if err != nil {
			return nil, response.Error(err)
		}
		items = append(items, item)
	}
	return &pb.PerformanceCyclesResponse{Items: items, Page: pageInfo(page)}, nil
}
func (h *Handler) CreatePerformanceCycle(ctx context.Context, req *pb.CreatePerformanceCycleRequest) (*pb.PerformanceCycleResponse, error) {
	if req == nil || req.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	in := req.Input
	start, err := response.Date(in.GetPeriodStart())
	if err != nil {
		return nil, response.Error(err)
	}
	end, err := response.Date(in.GetPeriodEnd())
	if err != nil {
		return nil, response.Error(err)
	}
	self, err := response.OptionalTimestamp(in.SelfReviewDueAt)
	if err != nil {
		return nil, response.Error(err)
	}
	manager, err := response.OptionalTimestamp(in.ManagerReviewDueAt)
	if err != nil {
		return nil, response.Error(err)
	}
	calibration, err := response.OptionalTimestamp(in.CalibrationDueAt)
	if err != nil {
		return nil, response.Error(err)
	}
	launched, err := response.OptionalTimestamp(in.LaunchedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	configuration, err := response.JSON(in.Configuration)
	if err != nil {
		return nil, response.Error(err)
	}
	return cycleResponse(h.service.CreatePerformanceCycle(ctx, service.CreatePerformanceCycleInput{CycleCode: in.GetCycleCode(), CycleName: in.GetCycleName(), CycleType: in.GetCycleType(), Status: enumPtr[entity.PerformanceCycleStatus](in.Status), PeriodStart: start, PeriodEnd: end, SelfReviewDueAt: self, ManagerReviewDueAt: manager, CalibrationDueAt: calibration, LaunchedAt: launched, LaunchedByPrincipalID: in.LaunchedByPrincipalId, Configuration: configuration}))
}
func (h *Handler) LaunchPerformanceCycle(ctx context.Context, req *common.IDRequest) (*pb.PerformanceCycleResponse, error) {
	if req == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return cycleResponse(h.service.LaunchPerformanceCycle(ctx, req.Id))
}
func (h *Handler) UpdatePerformanceCycle(ctx context.Context, req *pb.UpdatePerformanceCycleRequest) (*pb.PerformanceCycleResponse, error) {
	if req == nil || req.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	m, err := response.ValidateMask(req.UpdateMask, "cycle_code", "cycle_name", "cycle_type", "status", "period_start", "period_end", "self_review_due_at", "manager_review_due_at", "calibration_due_at", "launched_at", "launched_by_principal_id", "configuration")
	if err != nil {
		return nil, response.Error(err)
	}
	in := req.Input
	start, err := dateField(m["period_start"], in.PeriodStart)
	if err != nil {
		return nil, response.Error(err)
	}
	end, err := dateField(m["period_end"], in.PeriodEnd)
	if err != nil {
		return nil, response.Error(err)
	}
	self, err := timeField(m["self_review_due_at"], in.SelfReviewDueAt)
	if err != nil {
		return nil, response.Error(err)
	}
	manager, err := timeField(m["manager_review_due_at"], in.ManagerReviewDueAt)
	if err != nil {
		return nil, response.Error(err)
	}
	calibration, err := timeField(m["calibration_due_at"], in.CalibrationDueAt)
	if err != nil {
		return nil, response.Error(err)
	}
	launched, err := timeField(m["launched_at"], in.LaunchedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	configuration, err := jsonField(m["configuration"], in.Configuration)
	if err != nil {
		return nil, response.Error(err)
	}
	return cycleResponse(h.service.UpdatePerformanceCycle(ctx, req.Id, service.UpdatePerformanceCycleInput{CycleCode: field(m["cycle_code"], in.CycleCode), CycleName: field(m["cycle_name"], in.CycleName), CycleType: field(m["cycle_type"], in.CycleType), Status: field(m["status"], enumPtr[entity.PerformanceCycleStatus](in.Status)), PeriodStart: start, PeriodEnd: end, SelfReviewDueAt: self, ManagerReviewDueAt: manager, CalibrationDueAt: calibration, LaunchedAt: launched, LaunchedByPrincipalID: field(m["launched_by_principal_id"], in.LaunchedByPrincipalId), Configuration: configuration}))
}
func (h *Handler) DeletePerformanceCycle(ctx context.Context, req *common.IDRequest) (*common.EmptyResponse, error) {
	return remove(ctx, req, h.service.DeletePerformanceCycle)
}
func (h *Handler) GetPerformanceCycleByID(ctx context.Context, req *common.IDRequest) (*pb.PerformanceCycleResponse, error) {
	if req == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return cycleResponse(h.service.GetPerformanceCycleByID(ctx, req.Id))
}
func (h *Handler) GetAllPerformanceCycles(ctx context.Context, req *common.PageRequest) (*pb.PerformanceCyclesResponse, error) {
	if req == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return cycles(h.service.GetAllPerformanceCycles(ctx, int(req.Limit), int(req.Offset)))
}
func (h *Handler) FindPerformanceCycleByCycleCode(ctx context.Context, req *common.StringFinderRequest) (*pb.PerformanceCyclesResponse, error) {
	if req == nil || req.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return cycles(h.service.FindPerformanceCycleByCycleCode(ctx, *req.Value, int(req.Limit), int(req.Offset)))
}
func (h *Handler) FindPerformanceCycleByCycleType(ctx context.Context, req *common.StringFinderRequest) (*pb.PerformanceCyclesResponse, error) {
	if req == nil || req.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return cycles(h.service.FindPerformanceCycleByCycleType(ctx, *req.Value, int(req.Limit), int(req.Offset)))
}
func (h *Handler) FindPerformanceCycleByStatus(ctx context.Context, req *common.StringFinderRequest) (*pb.PerformanceCyclesResponse, error) {
	if req == nil || req.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return cycles(h.service.FindPerformanceCycleByStatus(ctx, entity.PerformanceCycleStatus(*req.Value), int(req.Limit), int(req.Offset)))
}
func (h *Handler) FindPerformanceCycleByLaunchedByPrincipalID(ctx context.Context, req *common.StringFinderRequest) (*pb.PerformanceCyclesResponse, error) {
	if req == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return cycles(h.service.FindPerformanceCycleByLaunchedByPrincipalID(ctx, req.Value, int(req.Limit), int(req.Offset)))
}
