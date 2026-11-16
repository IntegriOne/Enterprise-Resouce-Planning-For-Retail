package recruitment

import (
	"context"

	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/recruitment/v1"
	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/recruitment"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/response"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func applicationStageHistory(v *entity.ApplicationStageHistory) *pb.ApplicationStageHistory {
	return &pb.ApplicationStageHistory{
		Id: v.ID, ApplicationId: v.ApplicationID, PreviousStatus: stringPointer(v.PreviousStatus), NewStatus: string(v.NewStatus),
		StageCode: v.StageCode, ChangedAt: timestamppb.New(v.ChangedAt), ChangedByPrincipalId: v.ChangedByPrincipalID, ReasonCode: v.ReasonCode,
	}
}
func applicationStageHistoryResponse(v *entity.ApplicationStageHistory, err error) (*pb.ApplicationStageHistoryResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.ApplicationStageHistoryResponse{Item: applicationStageHistory(v)}, nil
}
func applicationStageHistories(page operation.Page[entity.ApplicationStageHistory], err error) (*pb.ApplicationStageHistoriesResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	items := make([]*pb.ApplicationStageHistory, 0, len(page.Items))
	for i := range page.Items {
		items = append(items, applicationStageHistory(&page.Items[i]))
	}
	return &pb.ApplicationStageHistoriesResponse{Items: items, Page: pageInfo(page)}, nil
}
func (h *Handler) CreateApplicationStageHistory(ctx context.Context, r *pb.CreateApplicationStageHistoryRequest) (*pb.ApplicationStageHistoryResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	in := r.Input
	return applicationStageHistoryResponse(h.service.CreateApplicationStageHistory(ctx, service.CreateApplicationStageHistoryInput{
		ApplicationID: in.GetApplicationId(), PreviousStatus: enumPointer[entity.ApplicationStatus](in.PreviousStatus), NewStatus: entity.ApplicationStatus(in.GetNewStatus()),
		StageCode: in.StageCode, ChangedByPrincipalID: in.ChangedByPrincipalId, ReasonCode: in.ReasonCode,
	}))
}
func (h *Handler) UpdateApplicationStageHistory(ctx context.Context, r *pb.UpdateApplicationStageHistoryRequest) (*pb.ApplicationStageHistoryResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	m, err := response.ValidateMask(r.UpdateMask, "application_id", "previous_status", "new_status", "stage_code", "changed_by_principal_id", "reason_code")
	if err != nil {
		return nil, response.Error(err)
	}
	in := r.Input
	return applicationStageHistoryResponse(h.service.UpdateApplicationStageHistory(ctx, r.Id, service.UpdateApplicationStageHistoryInput{
		ApplicationID: field(m["application_id"], in.ApplicationId), PreviousStatus: field(m["previous_status"], enumPointer[entity.ApplicationStatus](in.PreviousStatus)),
		NewStatus: field(m["new_status"], enumPointer[entity.ApplicationStatus](in.NewStatus)), StageCode: field(m["stage_code"], in.StageCode),
		ChangedByPrincipalID: field(m["changed_by_principal_id"], in.ChangedByPrincipalId), ReasonCode: field(m["reason_code"], in.ReasonCode),
	}))
}
func (h *Handler) DeleteApplicationStageHistory(ctx context.Context, r *common.IdentityIDRequest) (*common.EmptyResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	if err := h.service.DeleteApplicationStageHistory(ctx, r.Id); err != nil {
		return nil, response.Error(err)
	}
	return &common.EmptyResponse{}, nil
}
func (h *Handler) GetApplicationStageHistoryByID(ctx context.Context, r *common.IdentityIDRequest) (*pb.ApplicationStageHistoryResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return applicationStageHistoryResponse(h.service.GetApplicationStageHistoryByID(ctx, r.Id))
}
func (h *Handler) GetAllApplicationStageHistories(ctx context.Context, r *common.PageRequest) (*pb.ApplicationStageHistoriesResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return applicationStageHistories(h.service.GetAllApplicationStageHistory(ctx, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindApplicationStageHistoryByApplicationID(ctx context.Context, r *common.StringFinderRequest) (*pb.ApplicationStageHistoriesResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return applicationStageHistories(h.service.FindApplicationStageHistoryByApplicationID(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindApplicationStageHistoryByPreviousStatus(ctx context.Context, r *common.StringFinderRequest) (*pb.ApplicationStageHistoriesResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return applicationStageHistories(h.service.FindApplicationStageHistoryByPreviousStatus(ctx, enumPointer[entity.ApplicationStatus](r.Value), int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindApplicationStageHistoryByNewStatus(ctx context.Context, r *common.StringFinderRequest) (*pb.ApplicationStageHistoriesResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return applicationStageHistories(h.service.FindApplicationStageHistoryByNewStatus(ctx, entity.ApplicationStatus(*r.Value), int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindApplicationStageHistoryByStageCode(ctx context.Context, r *common.StringFinderRequest) (*pb.ApplicationStageHistoriesResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return applicationStageHistories(h.service.FindApplicationStageHistoryByStageCode(ctx, r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindApplicationStageHistoryByChangedByPrincipalID(ctx context.Context, r *common.StringFinderRequest) (*pb.ApplicationStageHistoriesResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return applicationStageHistories(h.service.FindApplicationStageHistoryByChangedByPrincipalID(ctx, r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindApplicationStageHistoryByReasonCode(ctx context.Context, r *common.StringFinderRequest) (*pb.ApplicationStageHistoriesResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return applicationStageHistories(h.service.FindApplicationStageHistoryByReasonCode(ctx, r.Value, int(r.Limit), int(r.Offset)))
}
