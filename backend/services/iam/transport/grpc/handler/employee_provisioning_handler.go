package handler

import (
	"context"

	"google.golang.org/protobuf/types/known/timestamppb"

	iamv1 "github.com/integrione/erp-retail/backend/services/iam/contract/gen/iam/v1"
	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
	"github.com/integrione/erp-retail/backend/services/iam/services/provisioning"
)

type EmployeeProvisioningHandler struct {
	iamv1.UnimplementedEmployeeProvisioningServiceServer
	service   *provisioning.Service
	errMapper *ErrorMapper
}

func (h *EmployeeProvisioningHandler) ReceiveEmployeeProvisioningRequest(ctx context.Context, request *iamv1.ReceiveEmployeeProvisioningRequestRequest) (*iamv1.ReceiveEmployeeProvisioningRequestResponse, error) {
	if request == nil || request.GetRequestedAt() == nil || !request.GetRequestedAt().IsValid() {
		return nil, h.errMapper.ToStatus(iamerrors.ErrInvalidInput)
	}
	receipt, err := h.service.Receive(ctx, request.GetHrmRequestId(), request.GetWorkerId(), request.GetLegalEntityId(), request.GetWorkerNumber(), request.GetCatalogRoleId(), request.GetRequestedAt().AsTime())
	if err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	return &iamv1.ReceiveEmployeeProvisioningRequestResponse{
		Id: receipt.ID, Status: receipt.Status, HrmRequestId: receipt.HRMRequestID, WorkerId: receipt.WorkerID, LegalEntityId: receipt.LegalEntityID,
	}, nil
}

func (h *EmployeeProvisioningHandler) ListEmployeeProvisioningRequests(ctx context.Context, request *iamv1.ListEmployeeProvisioningRequestsRequest) (*iamv1.ListEmployeeProvisioningRequestsResponse, error) {
	if request == nil || request.GetPage() == nil {
		return nil, nilRequestError(h.errMapper)
	}
	page := request.GetPage()
	limit, offset := page.GetLimit(), page.GetOffset()
	items, total, err := h.service.ListOpen(ctx, int(limit), int(offset))
	if err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	wireItems := make([]*iamv1.EmployeeProvisioningRequest, 0, len(items))
	for _, item := range items {
		wireItems = append(wireItems, toProtoEmployeeProvisioningRequest(item))
	}
	return &iamv1.ListEmployeeProvisioningRequestsResponse{
		Items: wireItems,
		Page:  &iamv1.PageInfo{Total: total, Limit: limit, Offset: offset, HasMore: int64(offset)+int64(len(items)) < total},
	}, nil
}

func (h *EmployeeProvisioningHandler) GetEmployeeProvisioningRequest(ctx context.Context, request *iamv1.GetEmployeeProvisioningRequestRequest) (*iamv1.GetEmployeeProvisioningRequestResponse, error) {
	if request == nil {
		return nil, nilRequestError(h.errMapper)
	}
	item, err := h.service.Get(ctx, request.GetId())
	if err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	return &iamv1.GetEmployeeProvisioningRequestResponse{Item: toProtoEmployeeProvisioningRequest(*item)}, nil
}

func toProtoEmployeeProvisioningRequest(item entity.EmployeeProvisioningRequest) *iamv1.EmployeeProvisioningRequest {
	return &iamv1.EmployeeProvisioningRequest{
		Id: item.ID, WorkerId: item.WorkerID, WorkerNumber: item.WorkerNumber, LegalEntityId: item.LegalEntityID,
		RequestedByPrincipalId: item.RequestedByPrincipalID, RequestedAt: timestamppb.New(item.RequestedAt),
		ReceivedAt: timestamppb.New(item.ReceivedAt), Status: item.Status,
	}
}

func (h *EmployeeProvisioningHandler) CompleteEmployeeProvisioningRequest(ctx context.Context, request *iamv1.CompleteEmployeeProvisioningRequestRequest) (*iamv1.CompleteEmployeeProvisioningRequestResponse, error) {
	if request == nil {
		return nil, nilRequestError(h.errMapper)
	}
	if err := h.service.Complete(ctx, request.GetId(), request.GetAccountId()); err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	return &iamv1.CompleteEmployeeProvisioningRequestResponse{}, nil
}

func (h *EmployeeProvisioningHandler) GetEmployeeProvisioningStatus(ctx context.Context, request *iamv1.GetEmployeeProvisioningStatusRequest) (*iamv1.GetEmployeeProvisioningStatusResponse, error) {
	if request == nil {
		return nil, nilRequestError(h.errMapper)
	}
	item, err := h.service.GetStatusForHRM(ctx, request.GetHrmRequestId(), request.GetLegalEntityId())
	if err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	return &iamv1.GetEmployeeProvisioningStatusResponse{
		Id: item.ID, Status: item.Status, HrmRequestId: item.HRMRequestID, WorkerId: item.WorkerID, LegalEntityId: item.LegalEntityID,
	}, nil
}

func (h *EmployeeProvisioningHandler) ResolveEmployeeProvisioningCatalog(ctx context.Context, request *iamv1.ResolveEmployeeProvisioningCatalogRequest) (*iamv1.ResolveEmployeeProvisioningCatalogResponse, error) {
	if request == nil {
		return nil, nilRequestError(h.errMapper)
	}
	id, err := h.service.ResolveCatalog(ctx, request.GetLegalEntityId())
	if err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	return &iamv1.ResolveEmployeeProvisioningCatalogResponse{CatalogRoleId: id}, nil
}
