package handler

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net/url"
	"testing"

	iamv1 "github.com/integrione/erp-retail/backend/services/iam/contract/gen/iam/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

type requisitionMandateStub struct {
	calls int
}

func (s *requisitionMandateStub) HasRequisitionApprovalMandate(_ context.Context, principal, unit string) (bool, error) {
	s.calls++
	return principal == "reviewer" && unit == "unit", nil
}

func TestRequisitionMandateRequiresVerifiedHRMWorkload(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name     string
		uri      string
		verified bool
		code     codes.Code
	}{
		{name: "missing certificate", code: codes.Unauthenticated},
		{name: "unverified workload", uri: hrmWorkloadURI, code: codes.Unauthenticated},
		{name: "other verified workload", uri: "spiffe://integrione/gateway", verified: true, code: codes.PermissionDenied},
		{name: "verified HRM", uri: hrmWorkloadURI, verified: true, code: codes.OK},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()
			if tc.uri != "" {
				uri, err := url.Parse(tc.uri)
				if err != nil {
					t.Fatal(err)
				}
				leaf := &x509.Certificate{URIs: []*url.URL{uri}}
				state := tls.ConnectionState{PeerCertificates: []*x509.Certificate{leaf}}
				if tc.verified {
					state.VerifiedChains = [][]*x509.Certificate{{leaf}}
				}
				ctx = peer.NewContext(ctx, &peer.Peer{AuthInfo: credentials.TLSInfo{State: state}})
			}
			service := &requisitionMandateStub{}
			h := &requisitionMandateHandler{service: service, errMapper: NewErrorMapper(nil)}
			res, err := h.VerifyRequisitionMandate(ctx, &iamv1.VerifyRequisitionMandateRequest{
				PrincipalId: "reviewer", OrganizationUnitId: "unit",
			})
			if status.Code(err) != tc.code {
				t.Fatalf("mandate code = %s, want %s", status.Code(err), tc.code)
			}
			if tc.code == codes.OK && (service.calls != 1 || !res.GetAllowed() || res.GetPrincipalId() != "reviewer" || res.GetOrganizationUnitId() != "unit") {
				t.Fatal("verified workload did not receive the correlated mandate")
			}
			if tc.code != codes.OK && service.calls != 0 {
				t.Fatal("unauthorized workload reached the mandate service")
			}
		})
	}
}
