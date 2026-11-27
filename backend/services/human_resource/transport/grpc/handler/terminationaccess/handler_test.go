package terminationaccess

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net/url"
	"testing"
	"time"

	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/workload/termination/v1"
	personnel "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/personnel"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

type verificationStub struct {
	calls int
}

func (v *verificationStub) VerifyTerminationAccess(_ context.Context, _, _ string, date *time.Time) (personnel.TerminationAccessDecision, error) {
	v.calls++
	return personnel.TerminationAccessDecision{
		EffectiveDate: *date,
		PrincipalID:   "00000000-0000-4000-8000-000000000001",
		MayDisable:    true,
	}, nil
}

func TestVerifyTerminationAccessRequiresIAMWorkload(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, workload, date string
		want                 codes.Code
	}{
		{"no verified certificate", "", "2026-11-27", codes.Unauthenticated},
		{"other workload", "spiffe://integrione/gateway", "2026-11-27", codes.PermissionDenied},
		{"invalid date", iamWorkloadURI, "invalid", codes.InvalidArgument},
		{"verified IAM", iamWorkloadURI, "2026-11-27", codes.OK},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := t.Context()
			if tc.workload != "" {
				uri, err := url.Parse(tc.workload)
				if err != nil {
					t.Fatal(err)
				}
				ctx = peer.NewContext(ctx, &peer.Peer{AuthInfo: credentials.TLSInfo{State: tls.ConnectionState{
					VerifiedChains: [][]*x509.Certificate{{{URIs: []*url.URL{uri}}}},
				}}})
			}
			service := &verificationStub{}
			handler := &handler{verifier: service}
			req := &pb.VerifyTerminationAccessRequest{WorkerId: "worker", OffboardingPlanId: "plan", TerminationEffectiveDate: &tc.date}
			result, err := handler.VerifyTerminationAccess(ctx, req)
			if status.Code(err) != tc.want {
				t.Fatalf("status = %v, want %v", err, tc.want)
			}
			if tc.want == codes.OK {
				if result.GetWorkerId() != req.GetWorkerId() || result.GetOffboardingPlanId() != req.GetOffboardingPlanId() ||
					result.GetTerminationEffectiveDate() != tc.date || !result.GetMayDisable() || service.calls != 1 {
					t.Fatalf("invalid verification response: %+v", result)
				}
			} else if service.calls != 0 {
				t.Fatal("invalid request reached personnel service")
			}
		})
	}
}

func (v *verificationStub) VerifyWorkerAccountLink(context.Context, string, string, string) error {
	v.calls++
	return nil
}

func TestVerifyWorkerAccountLinkRequiresIAMWorkload(t *testing.T) {
	service := &verificationStub{}
	h := &handler{verifier: service}
	req := &pb.VerifyWorkerAccountLinkRequest{WorkerId: "worker", PrincipalId: "account", LegalEntityId: "entity"}
	if _, err := h.VerifyWorkerAccountLink(t.Context(), req); status.Code(err) != codes.Unauthenticated {
		t.Fatalf("without certificate: %v", err)
	}
	for _, workload := range []string{"spiffe://integrione/gateway", iamWorkloadURI} {
		uri, err := url.Parse(workload)
		if err != nil {
			t.Fatal(err)
		}
		ctx := peer.NewContext(t.Context(), &peer.Peer{AuthInfo: credentials.TLSInfo{State: tls.ConnectionState{
			VerifiedChains: [][]*x509.Certificate{{{URIs: []*url.URL{uri}}}},
		}}})
		result, err := h.VerifyWorkerAccountLink(ctx, req)
		if workload != iamWorkloadURI {
			if status.Code(err) != codes.PermissionDenied || service.calls != 0 {
				t.Fatalf("other workload reached verification: %v", err)
			}
			continue
		}
		if err != nil || result.GetWorkerId() != req.GetWorkerId() || result.GetPrincipalId() != req.GetPrincipalId() ||
			result.GetLegalEntityId() != req.GetLegalEntityId() || service.calls != 1 {
			t.Fatalf("IAM verification: %v, %v", result, err)
		}
	}
}
