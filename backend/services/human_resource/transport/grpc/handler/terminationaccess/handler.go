package terminationaccess

import (
	"context"
	"time"

	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/workload/termination/v1"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	personnel "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/personnel"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/response"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

const iamWorkloadURI = "spiffe://integrione/iam"

type verifier interface {
	VerifyWorkerAccountLink(context.Context, string, string, string) error
	VerifyTerminationAccess(context.Context, string, string, *time.Time) (personnel.TerminationAccessDecision, error)
}

type handler struct {
	pb.UnimplementedTerminationAccessServiceServer
	verifier verifier
}

// Register mendaftarkan verifier pada listener mTLS yang terpisah dari listener bisnis HRM.
func Register(registrar grpc.ServiceRegistrar, service verifier) error {
	if operation.RequireDependency(registrar) != nil || operation.RequireDependency(service) != nil {
		return hrerrors.ErrUnavailable
	}
	pb.RegisterTerminationAccessServiceServer(registrar, &handler{verifier: service})
	return nil
}

func (h *handler) VerifyTerminationAccess(ctx context.Context, req *pb.VerifyTerminationAccessRequest) (*pb.VerifyTerminationAccessResponse, error) {
	if err := requireIAMWorkload(ctx); err != nil {
		return nil, err
	}
	var effectiveDate *time.Time
	if req != nil && req.TerminationEffectiveDate != nil {
		parsed, err := time.Parse(time.DateOnly, req.GetTerminationEffectiveDate())
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid termination effective date")
		}
		effectiveDate = &parsed
	}
	decision, err := h.verifier.VerifyTerminationAccess(ctx, req.GetWorkerId(), req.GetOffboardingPlanId(), effectiveDate)
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.VerifyTerminationAccessResponse{
		WorkerId: req.GetWorkerId(), OffboardingPlanId: req.GetOffboardingPlanId(),
		TerminationEffectiveDate: decision.EffectiveDate.Format(time.DateOnly),
		PrincipalId:              decision.PrincipalID, MayDisable: decision.MayDisable,
	}, nil
}

func (h *handler) VerifyWorkerAccountLink(ctx context.Context, req *pb.VerifyWorkerAccountLinkRequest) (*pb.VerifyWorkerAccountLinkResponse, error) {
	if err := requireIAMWorkload(ctx); err != nil {
		return nil, err
	}
	workerID, principalID, legalEntityID := req.GetWorkerId(), req.GetPrincipalId(), req.GetLegalEntityId()
	if err := h.verifier.VerifyWorkerAccountLink(ctx, workerID, principalID, legalEntityID); err != nil {
		return nil, response.Error(err)
	}
	return &pb.VerifyWorkerAccountLinkResponse{WorkerId: workerID, PrincipalId: principalID, LegalEntityId: legalEntityID}, nil
}

func requireIAMWorkload(ctx context.Context) error {
	caller, ok := peer.FromContext(ctx)
	if !ok || caller == nil {
		return status.Error(codes.Unauthenticated, "verified workload certificate required")
	}
	info, ok := caller.AuthInfo.(credentials.TLSInfo)
	if !ok || len(info.State.VerifiedChains) == 0 || len(info.State.VerifiedChains[0]) == 0 {
		return status.Error(codes.Unauthenticated, "verified workload certificate required")
	}
	workload := info.State.VerifiedChains[0][0]
	if workload == nil || len(workload.URIs) != 1 || workload.URIs[0] == nil || workload.URIs[0].String() != iamWorkloadURI {
		return status.Error(codes.PermissionDenied, "workload is not authorized to access HRM verification")
	}
	return nil
}
