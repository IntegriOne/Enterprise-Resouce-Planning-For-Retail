package human_resource

import (
	"context"

	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	hrservice "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/hrservice/v1"
)

// ResolveServiceRequest menyelesaikan permintaan layanan HR dengan bukti dokumen resolusi.
func (a *Adapter) ResolveServiceRequest(ctx context.Context, rawID, resolutionDocumentVersionRefID string) (*model.ServiceRequest, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_service_requests", "resolve", "human_resource.service_requests.resolve")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "service-request", "service-request")
	if err != nil {
		return nil, err
	}
	docVerID, err := normalizeRequiredUUID(resolutionDocumentVersionRefID, "document-version", "resolution_document_version_reference_id")
	if err != nil {
		return nil, err
	}
	if a.hrService == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.hrService.ResolveServiceRequest(rpcCtx, &hrservice.ResolveServiceRequestRequest{
		Id:                                   id,
		ResolutionDocumentVersionReferenceId: docVerID,
	})
	if err != nil {
		return nil, mapHRMWriteError(err, "service-request")
	}
	return toModelServiceRequest(resp.GetValue())
}

// CloseServiceRequest menutup permintaan layanan HR yang telah selesai.
func (a *Adapter) CloseServiceRequest(ctx context.Context, rawID string) (*model.ServiceRequest, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_service_requests", "close", "human_resource.service_requests.close")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "service-request", "service-request")
	if err != nil {
		return nil, err
	}
	if a.hrService == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.hrService.CloseServiceRequest(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapHRMWriteError(err, "service-request")
	}
	return toModelServiceRequest(resp.GetValue())
}

// CancelServiceRequest membatalkan permintaan layanan HR.
func (a *Adapter) CancelServiceRequest(ctx context.Context, rawID string) (*model.ServiceRequest, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_service_requests", "cancel", "human_resource.service_requests.cancel")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "service-request", "service-request")
	if err != nil {
		return nil, err
	}
	if a.hrService == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.hrService.CancelServiceRequest(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapHRMWriteError(err, "service-request")
	}
	return toModelServiceRequest(resp.GetValue())
}

// AssignServiceRequest menugaskan permintaan layanan HR kepada petugas.
func (a *Adapter) AssignServiceRequest(ctx context.Context, rawID string) (*model.ServiceRequest, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_service_requests", "assign", "human_resource.service_requests.assign")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "service-request", "service-request")
	if err != nil {
		return nil, err
	}
	if a.hrService == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.hrService.AssignServiceRequest(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapHRMWriteError(err, "service-request")
	}
	return toModelServiceRequest(resp.GetValue())
}
