package iam

import (
	"context"

	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	"github.com/integrione/erp-retail/backend/gateway/internal/source_connections"
	iamv1 "github.com/integrione/erp-retail/backend/services/iam/contract/gen/iam/v1"
)

func (a *Adapter) EmployeeProvisioningRequests(ctx context.Context, after, before *string, first, last *int) (*model.EmployeeProvisioningRequestConnection, error) {
	session, err := a.requireEmployeeProvisioning(ctx)
	if err != nil {
		return nil, err
	}
	window, err := ParseRelayPage(after, before, first, last)
	if err != nil {
		return nil, err
	}
	items, offset, pageInfo, total, err := FetchRelayPage(window, func(limit, offset int32) ([]*iamv1.EmployeeProvisioningRequest, *iamv1.PageInfo, error) {
		response, err := callIAM(ctx, a, &session.Caller, hasEmployeeProvisioningClient,
			func(rpcCtx context.Context, clients source_connections.IAMClientSet) (*iamv1.ListEmployeeProvisioningRequestsResponse, error) {
				return clients.EmployeeProvisioning.ListEmployeeProvisioningRequests(rpcCtx, &iamv1.ListEmployeeProvisioningRequestsRequest{
					Page: &iamv1.PageRequest{Limit: limit, Offset: offset},
				})
			})
		if err != nil {
			return nil, nil, err
		}
		return response.GetItems(), response.GetPage(), nil
	})
	if err != nil {
		return nil, err
	}
	result := &model.EmployeeProvisioningRequestConnection{
		Nodes:    make([]*model.EmployeeProvisioningRequest, 0, len(items)),
		Edges:    make([]*model.EmployeeProvisioningRequestEdge, 0, len(items)),
		PageInfo: pageInfo, TotalCount: total,
	}
	for index, item := range items {
		node, err := toModelEmployeeProvisioningRequest(item)
		if err != nil {
			return nil, err
		}
		result.Nodes = append(result.Nodes, node)
		result.Edges = append(result.Edges, &model.EmployeeProvisioningRequestEdge{Node: node, Cursor: EncodeCursor(offset + index)})
	}
	return result, nil
}

func (a *Adapter) EmployeeProvisioningRequest(ctx context.Context, id string) (*model.EmployeeProvisioningRequest, error) {
	session, err := a.requireEmployeeProvisioning(ctx)
	if err != nil {
		return nil, err
	}
	id, err = parseStrictUUIDOrTypedID(id, ResourceEmployeeProvisioningRequest)
	if err != nil {
		return nil, err
	}
	return a.employeeProvisioningRequest(ctx, session, id)
}

func (a *Adapter) employeeProvisioningRequest(ctx context.Context, session *middleware.ResolvedSession, id string) (*model.EmployeeProvisioningRequest, error) {
	response, err := callIAM(ctx, a, &session.Caller, hasEmployeeProvisioningClient,
		func(rpcCtx context.Context, clients source_connections.IAMClientSet) (*iamv1.GetEmployeeProvisioningRequestResponse, error) {
			return clients.EmployeeProvisioning.GetEmployeeProvisioningRequest(rpcCtx, &iamv1.GetEmployeeProvisioningRequestRequest{Id: id})
		})
	if err != nil {
		return nil, err
	}
	if response.GetItem().GetId() != id {
		return nil, middleware.ErrUnavailable("invalid identity provisioning response", nil)
	}
	return toModelEmployeeProvisioningRequest(response.GetItem())
}

func (a *Adapter) requireEmployeeProvisioning(ctx context.Context) (*middleware.ResolvedSession, error) {
	session, err := a.requirePerm(ctx, "iam.users.write", true)
	if err != nil {
		return nil, err
	}
	if session.Caller.UserKind != "employee" || !session.Token.MFAAuthenticated {
		return nil, middleware.ErrForbidden("employee MFA authentication required", nil)
	}
	return session, nil
}

func hasEmployeeProvisioningClient(clients source_connections.IAMClientSet) bool {
	return clients.EmployeeProvisioning != nil
}

func toModelEmployeeProvisioningRequest(item *iamv1.EmployeeProvisioningRequest) (*model.EmployeeProvisioningRequest, error) {
	if item.GetId() == "" || item.GetWorkerId() == "" || !item.GetRequestedAt().IsValid() || !item.GetReceivedAt().IsValid() {
		return nil, middleware.ErrUnavailable("invalid identity provisioning response", nil)
	}
	return &model.EmployeeProvisioningRequest{
		ID: item.GetId(), WorkerID: EncodeGID(ResourceWorker, item.GetWorkerId()), WorkerNumber: item.GetWorkerNumber(), LegalEntityID: item.GetLegalEntityId(),
		RequestedByPrincipalID: item.GetRequestedByPrincipalId(), RequestedAt: item.GetRequestedAt().AsTime(),
		ReceivedAt: item.GetReceivedAt().AsTime(), Status: item.GetStatus(),
	}, nil
}
