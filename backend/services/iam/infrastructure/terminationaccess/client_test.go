package terminationaccess

import (
	"context"
	"errors"
	"testing"

	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/workload/termination/v1"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type fixtureClient struct {
	response *pb.VerifyTerminationAccessResponse
	link     *pb.VerifyWorkerAccountLinkResponse
	linkErr  error
}

func (f fixtureClient) VerifyTerminationAccess(context.Context, *pb.VerifyTerminationAccessRequest, ...grpc.CallOption) (*pb.VerifyTerminationAccessResponse, error) {
	return f.response, nil
}

func TestVerifyChecksTerminationCorrelation(t *testing.T) {
	t.Parallel()
	const worker = "00000000-0000-4000-8000-000000000001"
	const plan = "00000000-0000-4000-8000-000000000002"
	const principal = "00000000-0000-4000-8000-000000000003"
	const date = "2026-11-27"
	for _, tc := range []struct {
		name     string
		response *pb.VerifyTerminationAccessResponse
		want     error
	}{
		{"matched", &pb.VerifyTerminationAccessResponse{WorkerId: worker, OffboardingPlanId: plan, TerminationEffectiveDate: date, PrincipalId: principal, MayDisable: true}, nil},
		{"other worker", &pb.VerifyTerminationAccessResponse{WorkerId: plan, OffboardingPlanId: plan, TerminationEffectiveDate: date, PrincipalId: principal, MayDisable: true}, iamerrors.ErrUnavailable},
		{"other plan", &pb.VerifyTerminationAccessResponse{WorkerId: worker, OffboardingPlanId: worker, TerminationEffectiveDate: date, PrincipalId: principal, MayDisable: true}, iamerrors.ErrUnavailable},
		{"other date", &pb.VerifyTerminationAccessResponse{WorkerId: worker, OffboardingPlanId: plan, TerminationEffectiveDate: "2026-11-26", PrincipalId: principal, MayDisable: true}, iamerrors.ErrUnavailable},
		{"missing principal", &pb.VerifyTerminationAccessResponse{WorkerId: worker, OffboardingPlanId: plan, TerminationEffectiveDate: date, MayDisable: true}, iamerrors.ErrUnavailable},
		{"no account", &pb.VerifyTerminationAccessResponse{WorkerId: worker, OffboardingPlanId: plan, TerminationEffectiveDate: date}, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client := &Client{rpc: fixtureClient{response: tc.response}}
			decision, err := client.Verify(t.Context(), worker, plan, date)
			if !errors.Is(err, tc.want) {
				t.Fatalf("verify: got %v, want %v", err, tc.want)
			}
			if err == nil && (decision.MayDisable != tc.response.GetMayDisable() || decision.PrincipalID != tc.response.GetPrincipalId() || decision.EffectiveDate != date) {
				t.Fatalf("unexpected decision: %+v", decision)
			}
		})
	}
}

func (f fixtureClient) VerifyWorkerAccountLink(context.Context, *pb.VerifyWorkerAccountLinkRequest, ...grpc.CallOption) (*pb.VerifyWorkerAccountLinkResponse, error) {
	return f.link, f.linkErr
}

func TestVerifyWorkerAccountLinkChecksCorrelatedIdentity(t *testing.T) {
	const worker = "00000000-0000-4000-8000-000000000001"
	const account = "00000000-0000-4000-8000-000000000002"
	const legalEntity = "00000000-0000-4000-8000-000000000003"
	for _, tc := range []struct {
		name string
		stub fixtureClient
		want error
	}{
		{"match", fixtureClient{link: &pb.VerifyWorkerAccountLinkResponse{WorkerId: worker, PrincipalId: account, LegalEntityId: legalEntity}}, nil},
		{"other scope", fixtureClient{link: &pb.VerifyWorkerAccountLinkResponse{WorkerId: worker, PrincipalId: account, LegalEntityId: worker}}, iamerrors.ErrUnavailable},
		{"not linked", fixtureClient{linkErr: status.Error(codes.FailedPrecondition, "mismatch")}, iamerrors.ErrConflict},
		{"unavailable", fixtureClient{linkErr: status.Error(codes.Unavailable, "unavailable")}, iamerrors.ErrUnavailable},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client := Client{rpc: tc.stub}
			if err := client.VerifyWorkerAccountLink(t.Context(), worker, account, legalEntity); !errors.Is(err, tc.want) {
				t.Fatalf("verify: got %v, want %v", err, tc.want)
			}
		})
	}
}
