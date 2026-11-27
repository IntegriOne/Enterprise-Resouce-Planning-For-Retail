package human_resource

import (
	"context"
	"google.golang.org/protobuf/proto"
	"testing"

	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	"github.com/integrione/erp-retail/backend/gateway/internal/source_connections"
	accessrequest "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/accessrequest/v1"
	iamv1 "github.com/integrione/erp-retail/backend/services/iam/contract/gen/iam/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type employeeProvisioningStatusStub struct {
	iamv1.EmployeeProvisioningServiceClient
	code codes.Code
}

func (s employeeProvisioningStatusStub) GetEmployeeProvisioningStatus(
	context.Context, *iamv1.GetEmployeeProvisioningStatusRequest, ...grpc.CallOption,
) (*iamv1.GetEmployeeProvisioningStatusResponse, error) {
	return nil, status.Error(s.code, "fixture failure")
}

func TestReadEmployeeAccountStatusDuringIAMFailure(t *testing.T) {
	t.Parallel()
	ctx := middleware.WithResolvedSession(t.Context(), &middleware.ResolvedSession{})
	request := &accessrequest.AccessRequest{Id: "request", Status: "pending", RequestedAt: timestamppb.Now()}
	for _, tc := range []struct {
		code       codes.Code
		wantStatus string
	}{
		{codes.Unavailable, "verification_unavailable"},
		{codes.NotFound, "pending"},
		{codes.PermissionDenied, ""},
	} {
		t.Run(tc.code.String(), func(t *testing.T) {
			adapter := Adapter{iam: source_connections.IAMClientSet{
				EmployeeProvisioning: employeeProvisioningStatusStub{code: tc.code},
			}}
			got, err := adapter.readIAMEmployeeAccountStatus(ctx, request)
			if tc.wantStatus == "" {
				if err == nil || got != nil {
					t.Fatal("expected denial without returning the request")
				}
				return
			}
			if err != nil || got == nil || got.Status != tc.wantStatus || got.IamRequestID != nil {
				t.Fatalf("unexpected provisioning status: %+v, %v", got, err)
			}
		})
	}
}

func TestIAMEmployeeAccountReceiptMustMatchRequest(t *testing.T) {
	t.Parallel()
	request := &accessrequest.AccessRequest{Id: "hrm-request", WorkerId: "worker", RequestedScopeReferenceId: "legal-entity", Status: "pending", RequestedAt: timestamppb.Now()}
	receipt := &iamv1.GetEmployeeProvisioningStatusResponse{Id: "00000000-0000-4000-8000-000000000001", Status: "linked", HrmRequestId: request.GetId(), WorkerId: request.GetWorkerId(), LegalEntityId: request.GetRequestedScopeReferenceId()}
	for _, tc := range []struct {
		name  string
		alter func(*iamv1.GetEmployeeProvisioningStatusResponse)
	}{
		{"matching", nil},
		{"wrong HRM request", func(r *iamv1.GetEmployeeProvisioningStatusResponse) { r.HrmRequestId = "another" }},
		{"wrong worker", func(r *iamv1.GetEmployeeProvisioningStatusResponse) { r.WorkerId = "another" }},
		{"wrong legal entity", func(r *iamv1.GetEmployeeProvisioningStatusResponse) { r.LegalEntityId = "another" }},
		{"missing identity", func(r *iamv1.GetEmployeeProvisioningStatusResponse) { r.WorkerId = "" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			response := proto.CloneOf(receipt)
			modelRequest := toEmployeeAccountRequest(request)
			if tc.alter != nil {
				tc.alter(response)
			}
			err := applyIAMProvisioningReceipt(modelRequest, request, response)
			if tc.alter == nil {
				if err != nil || modelRequest.Status != "applied" || modelRequest.IamRequestID == nil {
					t.Fatalf("matching receipt was not applied: %+v, %v", modelRequest, err)
				}
			} else if err == nil || modelRequest.Status != "pending" || modelRequest.IamRequestID != nil {
				t.Fatalf("mismatched receipt changed local state: %+v, %v", modelRequest, err)
			}
		})
	}
}
