package handler

import (
	"context"
	"time"

	iamv1 "github.com/integrione/erp-retail/backend/services/iam/contract/gen/iam/v1"
	"github.com/integrione/erp-retail/backend/services/iam/services/user"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type employeeTerminationApplier interface {
	ApplyEmployeeTermination(context.Context, user.TerminationAccessVerifier, string, string, string) (user.TerminationOutcome, error)
	VerifyEmployeeTerminationClearance(context.Context, user.TerminationAccessVerifier, string, string, string, string) error
}

type employeeTerminationHandler struct {
	iamv1.UnimplementedEmployeeTerminationServiceServer
	service   employeeTerminationApplier
	verifier  user.TerminationAccessVerifier
	errMapper *ErrorMapper
}

// RegisterEmployeeTermination mendaftarkan penerima terminasi pada listener mTLS privat.
func RegisterEmployeeTermination(server grpc.ServiceRegistrar, service employeeTerminationApplier, verifier user.TerminationAccessVerifier, mapper *ErrorMapper) {
	iamv1.RegisterEmployeeTerminationServiceServer(server, &employeeTerminationHandler{
		service: service, verifier: verifier, errMapper: mapper,
	})
}

func (h *employeeTerminationHandler) ApplyEmployeeTermination(ctx context.Context, req *iamv1.ApplyEmployeeTerminationRequest) (*iamv1.ApplyEmployeeTerminationResponse, error) {
	if err := requireHRMWorkload(ctx); err != nil {
		return nil, err
	}
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	date, err := time.Parse(time.DateOnly, req.GetEffectiveDate())
	if err != nil || date.IsZero() {
		return nil, status.Error(codes.InvalidArgument, "valid termination effective date required")
	}
	if h.service == nil || h.verifier == nil {
		return nil, status.Error(codes.Unavailable, "termination workflow is unavailable")
	}
	outcome, err := h.service.ApplyEmployeeTermination(ctx, h.verifier, req.GetWorkerId(), req.GetOffboardingPlanId(), req.GetEffectiveDate())
	if err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	var result iamv1.EmployeeTerminationOutcome
	switch outcome {
	case user.TerminationAccountDisabled:
		result = iamv1.EmployeeTerminationOutcome_EMPLOYEE_TERMINATION_OUTCOME_ACCOUNT_DISABLED
	case user.TerminationAccountRetained:
		result = iamv1.EmployeeTerminationOutcome_EMPLOYEE_TERMINATION_OUTCOME_ACCOUNT_RETAINED
	case user.TerminationNoLinkedAccount:
		result = iamv1.EmployeeTerminationOutcome_EMPLOYEE_TERMINATION_OUTCOME_NO_LINKED_ACCOUNT
	default:
		return nil, status.Error(codes.Internal, "unexpected termination outcome")
	}
	return &iamv1.ApplyEmployeeTerminationResponse{
		WorkerId: req.GetWorkerId(), OffboardingPlanId: req.GetOffboardingPlanId(),
		EffectiveDate: req.GetEffectiveDate(), Outcome: result,
	}, nil
}

func (h *employeeTerminationHandler) VerifyEmployeeTerminationClearance(ctx context.Context, req *iamv1.VerifyEmployeeTerminationClearanceRequest) (*iamv1.VerifyEmployeeTerminationClearanceResponse, error) {
	if err := requireHRMWorkload(ctx); err != nil {
		return nil, err
	}
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	if h.service == nil || h.verifier == nil {
		return nil, status.Error(codes.Unavailable, "termination workflow is unavailable")
	}
	if err := h.service.VerifyEmployeeTerminationClearance(ctx, h.verifier, req.GetWorkerId(), req.GetOffboardingPlanId(), req.GetEffectiveDate(), req.GetWorkerPrincipalId()); err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	return &iamv1.VerifyEmployeeTerminationClearanceResponse{
		WorkerId: req.GetWorkerId(), OffboardingPlanId: req.GetOffboardingPlanId(),
		EffectiveDate: req.GetEffectiveDate(), WorkerPrincipalId: req.GetWorkerPrincipalId(),
	}, nil
}
