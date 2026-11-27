package iam

import (
	"context"
	"errors"
	"testing"

	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	iamv1 "github.com/integrione/erp-retail/backend/services/iam/contract/gen/iam/v1"
	"google.golang.org/grpc"
)

type mandateFixture struct {
	response *iamv1.VerifyRequisitionMandateResponse
}

func (m mandateFixture) VerifyRequisitionMandate(context.Context, *iamv1.VerifyRequisitionMandateRequest, ...grpc.CallOption) (*iamv1.VerifyRequisitionMandateResponse, error) {
	return m.response, nil
}

func TestRequisitionMandateChecksCorrelationBeforeDecision(t *testing.T) {
	t.Parallel()
	const principal = "00000000-0000-4000-8000-000000000001"
	const unit = "00000000-0000-4000-8000-000000000002"
	for _, tc := range []struct {
		name   string
		result *iamv1.VerifyRequisitionMandateResponse
		want   error
	}{
		{name: "allowed", result: &iamv1.VerifyRequisitionMandateResponse{PrincipalId: principal, OrganizationUnitId: unit, Allowed: true}},
		{name: "revoked", result: &iamv1.VerifyRequisitionMandateResponse{PrincipalId: principal, OrganizationUnitId: unit}, want: hrerrors.ErrForbidden},
		{name: "wrong unit", result: &iamv1.VerifyRequisitionMandateResponse{PrincipalId: principal, OrganizationUnitId: principal, Allowed: true}, want: hrerrors.ErrUnavailable},
		{name: "wrong principal", result: &iamv1.VerifyRequisitionMandateResponse{PrincipalId: unit, OrganizationUnitId: unit, Allowed: true}, want: hrerrors.ErrUnavailable},
		{name: "missing result", want: hrerrors.ErrUnavailable},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			client := RequisitionMandateClient{client: mandateFixture{response: tc.result}}
			if err := client.VerifyRequisitionMandate(t.Context(), principal, unit); !errors.Is(err, tc.want) {
				t.Fatalf("mandate error = %v, want %v", err, tc.want)
			}
		})
	}
}
