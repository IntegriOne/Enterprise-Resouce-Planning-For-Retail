package handler

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

const hrmWorkloadURI = "spiffe://integrione/human-resource"

func requireHRMWorkload(ctx context.Context) error {
	caller, ok := peer.FromContext(ctx)
	if !ok || caller == nil {
		return status.Error(codes.Unauthenticated, "verified workload certificate required")
	}
	info, ok := caller.AuthInfo.(credentials.TLSInfo)
	if !ok || len(info.State.VerifiedChains) == 0 || len(info.State.VerifiedChains[0]) == 0 {
		return status.Error(codes.Unauthenticated, "verified workload certificate required")
	}
	workload := info.State.VerifiedChains[0][0]
	if workload == nil || len(workload.URIs) != 1 || workload.URIs[0] == nil || workload.URIs[0].String() != hrmWorkloadURI {
		return status.Error(codes.PermissionDenied, "workload is not authorized")
	}
	return nil
}
