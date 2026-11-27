package human_resource

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	"github.com/integrione/erp-retail/backend/gateway/internal/source_connections"
	accessrequest "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/accessrequest/v1"
	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	iamv1 "github.com/integrione/erp-retail/backend/services/iam/contract/gen/iam/v1"
	"github.com/integrione/erp-retail/backend/shared/authcontext"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RequestEmployeeAccount mengajukan akun untuk pekerja dan memakai badan hukum dari record HRM.
func (a *Adapter) RequestEmployeeAccount(ctx context.Context, input model.RequestEmployeeAccountInput) (*model.EmployeeAccountRequest, error) {
	writeCtx, err := a.authorizeGlobalEmployee(ctx, "human_resource.access_requests.create")
	if err != nil {
		return nil, err
	}
	readCtx, err := a.authorizeRead(ctx, workerReadPermission)
	if err != nil {
		return nil, err
	}
	workerID, workerErr := uuid.Parse(strings.TrimPrefix(input.WorkerID, workerGIDPrefix))
	if workerErr != nil || workerID == uuid.Nil {
		return nil, middleware.ErrInvalidArgument("invalid employee account request reference", nil)
	}
	if a.personnel == nil || a.accessRequests == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	if a.iam.EmployeeProvisioning == nil {
		return nil, middleware.ErrUnavailable("identity provisioning service unavailable", nil)
	}
	detail, err := a.personnel.GetWorkerByID(readCtx, &common.IDRequest{Id: workerID.String()})
	if err != nil {
		return nil, mapReadError(err, "worker")
	}
	worker := detail.GetItem()
	if worker.GetId() != workerID.String() || worker.GetPrimaryLegalEntityReferenceId() == "" {
		return nil, middleware.ErrUnavailable("invalid human resource worker response", nil)
	}
	workerReference, legalEntity := worker.GetId(), worker.GetPrimaryLegalEntityReferenceId()
	session, ok := middleware.ResolvedSessionFromContext(writeCtx)
	if !ok {
		return nil, middleware.ErrUnavailable("verified employee session unavailable", nil)
	}
	iamCtx := authcontext.WithCaller(writeCtx, session.Caller)
	resolvedCatalog, err := source_connections.InvokeIAM(iamCtx, a.iam, true, 0,
		func(ctx context.Context, clients source_connections.IAMClientSet) (*iamv1.ResolveEmployeeProvisioningCatalogResponse, error) {
			return clients.EmployeeProvisioning.ResolveEmployeeProvisioningCatalog(ctx, &iamv1.ResolveEmployeeProvisioningCatalogRequest{LegalEntityId: legalEntity})
		})
	if err != nil {
		return nil, err
	}
	catalog := resolvedCatalog.GetCatalogRoleId()
	if id, err := uuid.Parse(catalog); err != nil || id == uuid.Nil {
		return nil, middleware.ErrUnavailable("invalid identity provisioning catalog response", nil)
	}
	kind, scope := "provision", "legal_entity"
	key := "employee-account:" + workerReference
	result, err := a.accessRequests.CreateAccessRequest(writeCtx, &accessrequest.CreateAccessRequestRequest{
		Input: &accessrequest.AccessRequestInput{
			WorkerId: &workerReference, RequestKind: &kind,
			RequestedScopeType: &scope, RequestedScopeReferenceId: &legalEntity,
			ApprovedCatalogReferenceId: &catalog, IdempotencyKey: &key,
		},
	})
	request := result.GetValue()
	if status.Code(err) == codes.AlreadyExists {
		// Pengiriman ulang hanya membaca permintaan yang sebelumnya dibuat dengan key yang sama.
		retryCtx, authErr := a.authorizeGlobalEmployee(ctx, "human_resource.access_requests.read")
		if authErr != nil {
			return nil, authErr
		}
		existing, findErr := a.accessRequests.FindAccessRequestByIdempotencyKey(retryCtx, &common.StringFinderRequest{Value: &key, Limit: 1})
		if findErr != nil {
			return nil, mapReadError(findErr, "employee account request")
		}
		items := existing.GetItems()
		if len(items) != 1 {
			return nil, middleware.ErrConflict("employee account request conflicts with an existing request", nil)
		}
		request, err = items[0], nil
	}
	if err != nil {
		if status.Code(err) == codes.FailedPrecondition {
			return nil, middleware.ErrConflict("employee account request conflicts with worker state", err)
		}
		return nil, mapReadError(err, "employee account request")
	}
	if request.GetId() == "" || request.GetWorkerId() != workerReference || request.GetStatus() != "pending" ||
		request.GetRequestedScopeReferenceId() != legalEntity || request.GetRequestedAt() == nil {
		return nil, middleware.ErrUnavailable("invalid human resource account request response", nil)
	}
	if request.GetRequestedByPrincipalId() != session.Caller.PrincipalID ||
		request.GetApprovedCatalogReferenceId() != catalog || request.GetLifecycleEventId() <= 0 ||
		request.GetRequestKind() != kind || request.GetIdempotencyKey() != key {
		return nil, middleware.ErrConflict("employee account request does not match the original submission", nil)
	}
	receipt, err := source_connections.InvokeIAM(iamCtx, a.iam, true, 0,
		func(rpcCtx context.Context, clients source_connections.IAMClientSet) (*iamv1.ReceiveEmployeeProvisioningRequestResponse, error) {
			return clients.EmployeeProvisioning.ReceiveEmployeeProvisioningRequest(rpcCtx, &iamv1.ReceiveEmployeeProvisioningRequestRequest{
				HrmRequestId: request.GetId(), WorkerId: workerReference, WorkerNumber: worker.GetWorkerNumber(), LegalEntityId: legalEntity, CatalogRoleId: catalog, RequestedAt: request.GetRequestedAt(),
			})
		})
	if err != nil {
		// Kembalikan ID HRM agar pengaju masih bisa melihat dan mengulang pengiriman.
		return toEmployeeAccountRequest(request), err
	}
	modelRequest := toEmployeeAccountRequest(request)
	if err := applyIAMProvisioningReceipt(modelRequest, request, receipt); err != nil {
		return modelRequest, err
	}
	return modelRequest, nil
}

// EmployeeAccountRequest membaca status pengajuan yang pekerjanya dapat dilihat operator HRM.
func (a *Adapter) EmployeeAccountRequest(ctx context.Context, rawID string) (*model.EmployeeAccountRequest, error) {
	rpcCtx, err := a.authorizeGlobalEmployee(ctx, "human_resource.access_requests.read")
	if err != nil {
		return nil, err
	}
	id, err := parseNodeID(rawID, "employee-account-request")
	if err != nil || id == uuid.Nil {
		return nil, middleware.ErrInvalidArgument("invalid employee account request ID", nil)
	}
	if a.accessRequests == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	result, err := a.accessRequests.GetAccessRequestByID(rpcCtx, &common.IDRequest{Id: id.String()})
	if err != nil {
		return nil, mapReadError(err, "employee account request")
	}
	request := result.GetValue()
	if request.GetId() != id.String() || request.GetRequestKind() != "provision" {
		return nil, middleware.ErrNotFound("employee account request not found", nil)
	}
	if _, err := a.Worker(ctx, request.GetWorkerId()); err != nil {
		return nil, err
	}
	return a.readIAMEmployeeAccountStatus(rpcCtx, request)
}

// WorkerEmployeeAccountRequest membaca pengajuan provisioning dari halaman pekerja.
func (a *Adapter) WorkerEmployeeAccountRequest(ctx context.Context, workerID string) (*model.EmployeeAccountRequest, error) {
	rpcCtx, err := a.authorizeGlobalEmployee(ctx, "human_resource.access_requests.read")
	if err != nil {
		return nil, err
	}
	if a.accessRequests == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	worker, err := a.Worker(ctx, workerID)
	if err != nil {
		return nil, err
	}
	key := "employee-account:" + worker.RawID
	result, err := a.accessRequests.FindAccessRequestByIdempotencyKey(rpcCtx, &common.StringFinderRequest{Value: &key, Limit: 1})
	if err != nil {
		return nil, mapReadError(err, "employee account request")
	}
	items := result.GetItems()
	if len(items) == 0 {
		return nil, nil
	}
	request := items[0]
	if len(items) != 1 || request.GetId() == "" ||
		request.GetWorkerId() != worker.RawID || request.GetRequestKind() != "provision" {
		return nil, middleware.ErrUnavailable("invalid human resource account request response", nil)
	}
	return a.readIAMEmployeeAccountStatus(rpcCtx, request)
}

func (a *Adapter) readIAMEmployeeAccountStatus(rpcCtx context.Context, request *accessrequest.AccessRequest) (*model.EmployeeAccountRequest, error) {
	modelRequest := toEmployeeAccountRequest(request)
	session, ok := middleware.ResolvedSessionFromContext(rpcCtx)
	if !ok {
		return nil, middleware.ErrUnavailable("verified employee session unavailable", nil)
	}
	receipt, err := source_connections.InvokeIAM(authcontext.WithCaller(rpcCtx, session.Caller), a.iam, a.iam.EmployeeProvisioning != nil, 0,
		func(ctx context.Context, clients source_connections.IAMClientSet) (*iamv1.GetEmployeeProvisioningStatusResponse, error) {
			return clients.EmployeeProvisioning.GetEmployeeProvisioningStatus(ctx, &iamv1.GetEmployeeProvisioningStatusRequest{
				HrmRequestId: request.GetId(), LegalEntityId: request.GetRequestedScopeReferenceId(),
			})
		})
	if err != nil {
		if gatewayError, ok := middleware.AsGatewayError(err); ok {
			switch gatewayError.Code {
			case middleware.CodeNotFound:
				return modelRequest, nil
			case middleware.CodeUnavailable:
				modelRequest.Status = "verification_unavailable"
				return modelRequest, nil
			}
		}
		return nil, err
	}
	if err := applyIAMProvisioningReceipt(modelRequest, request, receipt); err != nil {
		return nil, err
	}
	return modelRequest, nil
}

func toEmployeeAccountRequest(request *accessrequest.AccessRequest) *model.EmployeeAccountRequest {
	return &model.EmployeeAccountRequest{
		ID: nodeID("employee-account-request", request.GetId()), RawID: request.GetId(), WorkerID: workerGIDPrefix + request.GetWorkerId(),
		LegalEntityID: request.GetRequestedScopeReferenceId(), Status: request.GetStatus(),
		RequestedAt: request.GetRequestedAt().AsTime(),
	}
}

type iamProvisioningReceipt interface {
	GetId() string
	GetStatus() string
	GetHrmRequestId() string
	GetWorkerId() string
	GetLegalEntityId() string
}

func applyIAMProvisioningReceipt(modelRequest *model.EmployeeAccountRequest, request *accessrequest.AccessRequest, receipt iamProvisioningReceipt) error {
	requestID, workerID, entityID := receipt.GetHrmRequestId(), receipt.GetWorkerId(), receipt.GetLegalEntityId()
	if requestID == "" || workerID == "" || entityID == "" || requestID != request.GetId() ||
		workerID != request.GetWorkerId() || entityID != request.GetRequestedScopeReferenceId() {
		return middleware.ErrUnavailable("identity provisioning receipt does not match human resource request", nil)
	}
	return applyIAMProvisioningStatus(modelRequest, receipt.GetId(), receipt.GetStatus())
}

func applyIAMProvisioningStatus(request *model.EmployeeAccountRequest, receiptID, status string) error {
	id, err := uuid.Parse(receiptID)
	if err != nil || id == uuid.Nil {
		return middleware.ErrUnavailable("invalid identity provisioning receipt", nil)
	}
	switch status {
	case "pending", "processing":
		request.Status = "submitted"
	case "linked":
		request.Status = "applied"
	case "rejected":
		request.Status = "rejected"
	default:
		return middleware.ErrUnavailable("invalid identity provisioning status", nil)
	}
	request.IamRequestID = &receiptID
	return nil
}
