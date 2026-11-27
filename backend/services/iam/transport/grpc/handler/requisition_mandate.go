package handler

import (
	"context"
	"strings"

	iamv1 "github.com/integrione/erp-retail/backend/services/iam/contract/gen/iam/v1"
	"google.golang.org/grpc"
)

type requisitionMandateReader interface {
	HasRequisitionApprovalMandate(context.Context, string, string) (bool, error)
}

type requisitionMandateHandler struct {
	iamv1.UnimplementedRequisitionMandateServiceServer
	service   requisitionMandateReader
	errMapper *ErrorMapper
}

// RegisterRequisitionMandate mendaftarkan verifier pada server privat dengan mTLS.
func RegisterRequisitionMandate(server grpc.ServiceRegistrar, service requisitionMandateReader, mapper *ErrorMapper) {
	iamv1.RegisterRequisitionMandateServiceServer(server, &requisitionMandateHandler{service: service, errMapper: mapper})
}

func (h *requisitionMandateHandler) VerifyRequisitionMandate(ctx context.Context, req *iamv1.VerifyRequisitionMandateRequest) (*iamv1.VerifyRequisitionMandateResponse, error) {
	if err := requireHRMWorkload(ctx); err != nil {
		return nil, err
	}
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}

	principal := strings.TrimSpace(req.GetPrincipalId())
	unit := strings.TrimSpace(req.GetOrganizationUnitId())
	allowed, err := h.service.HasRequisitionApprovalMandate(ctx, principal, unit)
	if err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	return &iamv1.VerifyRequisitionMandateResponse{
		PrincipalId: principal, OrganizationUnitId: unit, Allowed: allowed,
	}, nil
}
