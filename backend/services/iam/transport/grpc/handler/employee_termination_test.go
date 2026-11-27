package handler

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net/url"
	"testing"

	iamv1 "github.com/integrione/erp-retail/backend/services/iam/contract/gen/iam/v1"
	"github.com/integrione/erp-retail/backend/services/iam/services/user"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

type employeeTerminationStub struct {
	calls          int
	clearanceCalls int
}

func (s *employeeTerminationStub) ApplyEmployeeTermination(_ context.Context, _ user.TerminationAccessVerifier, _, _, _ string) (user.TerminationOutcome, error) {
	s.calls++
	return user.TerminationAccountDisabled, nil
}

func (s *employeeTerminationStub) VerifyEmployeeTerminationClearance(context.Context, user.TerminationAccessVerifier, string, string, string, string) error {
	s.clearanceCalls++
	return nil
}

func (*employeeTerminationStub) Verify(context.Context, string, string, string) (user.TerminationAccessDecision, error) {
	return user.TerminationAccessDecision{}, nil
}

func TestEmployeeTerminationAuthenticatesWorkloadBeforeDispatch(t *testing.T) {
	t.Parallel()
	request := &iamv1.ApplyEmployeeTerminationRequest{
		WorkerId: "dfe284c3-ab74-40c0-a847-944adf664cab", OffboardingPlanId: "2267e67b-a87b-425a-8ca4-b9e8909d647f",
		EffectiveDate: "2026-11-26",
	}
	clearance := &iamv1.VerifyEmployeeTerminationClearanceRequest{
		WorkerId: request.GetWorkerId(), OffboardingPlanId: request.GetOffboardingPlanId(), EffectiveDate: request.GetEffectiveDate(),
		WorkerPrincipalId: "dc2843c4-222c-41e5-8c2a-5066eeba1131",
	}
	stub := &employeeTerminationStub{}
	h := &employeeTerminationHandler{service: stub, verifier: stub, errMapper: NewErrorMapper(nil)}
	if _, err := h.ApplyEmployeeTermination(t.Context(), request); status.Code(err) != codes.Unauthenticated {
		t.Fatalf("missing certificate: %v", err)
	}
	if _, err := h.VerifyEmployeeTerminationClearance(t.Context(), clearance); status.Code(err) != codes.Unauthenticated {
		t.Fatalf("clearance without certificate: %v", err)
	}
	for _, workload := range []string{"spiffe://integrione/gateway", hrmWorkloadURI} {
		uri, err := url.Parse(workload)
		if err != nil {
			t.Fatal(err)
		}
		leaf := &x509.Certificate{URIs: []*url.URL{uri}}
		ctx := peer.NewContext(t.Context(), &peer.Peer{AuthInfo: credentials.TLSInfo{
			State: tls.ConnectionState{VerifiedChains: [][]*x509.Certificate{{leaf}}},
		}})
		reply, err := h.ApplyEmployeeTermination(ctx, request)
		proof, proofErr := h.VerifyEmployeeTerminationClearance(ctx, clearance)
		if workload != hrmWorkloadURI {
			if status.Code(err) != codes.PermissionDenied || status.Code(proofErr) != codes.PermissionDenied || stub.calls != 0 || stub.clearanceCalls != 0 {
				t.Fatalf("wrong workload reached termination use case: %v", err)
			}
			continue
		}
		if proofErr != nil || stub.clearanceCalls != 1 ||
			proof.GetWorkerId() != clearance.GetWorkerId() || proof.GetOffboardingPlanId() != clearance.GetOffboardingPlanId() ||
			proof.GetEffectiveDate() != clearance.GetEffectiveDate() || proof.GetWorkerPrincipalId() != clearance.GetWorkerPrincipalId() {
			t.Fatalf("verified workload did not receive correlated clearance: %v, %v", proof, proofErr)
		}
		if err != nil || stub.calls != 1 || reply.GetWorkerId() != request.GetWorkerId() ||
			reply.GetOffboardingPlanId() != request.GetOffboardingPlanId() || reply.GetEffectiveDate() != request.GetEffectiveDate() ||
			reply.GetOutcome() != iamv1.EmployeeTerminationOutcome_EMPLOYEE_TERMINATION_OUTCOME_ACCOUNT_DISABLED {
			t.Fatalf("verified workload did not receive correlated outcome: %v, %v", reply, err)
		}
	}
}
