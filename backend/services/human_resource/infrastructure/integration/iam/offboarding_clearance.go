package iam

import (
	"context"
	"strings"
	"time"

	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	offboarding "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/offboarding"
	iamv1 "github.com/integrione/erp-retail/backend/services/iam/contract/gen/iam/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ offboarding.IAMClearanceVerifier = (*RequisitionMandateClient)(nil)

// VerifyIAMClearance meminta bukti status akun tanpa mengirimkan data payroll.
func (c *RequisitionMandateClient) VerifyIAMClearance(ctx context.Context, request offboarding.IAMClearanceRequest) error {
	if c == nil || c.termination == nil {
		return hrerrors.ErrUnavailable
	}
	date := request.TerminationEffectiveDate.Format(time.DateOnly)
	call, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	response, err := c.termination.VerifyEmployeeTerminationClearance(call, &iamv1.VerifyEmployeeTerminationClearanceRequest{
		WorkerId: request.WorkerID, OffboardingPlanId: request.PlanID,
		EffectiveDate: date, WorkerPrincipalId: request.WorkerPrincipalID,
	})
	if err != nil {
		switch status.Code(err) {
		case codes.PermissionDenied:
			return hrerrors.ErrForbidden
		case codes.NotFound, codes.AlreadyExists, codes.FailedPrecondition, codes.Aborted:
			return hrerrors.ErrReferenceConflict
		default:
			return hrerrors.ErrUnavailable
		}
	}
	if !strings.EqualFold(response.GetWorkerId(), request.WorkerID) ||
		!strings.EqualFold(response.GetOffboardingPlanId(), request.PlanID) ||
		!strings.EqualFold(response.GetWorkerPrincipalId(), request.WorkerPrincipalID) || response.GetEffectiveDate() != date {
		return hrerrors.ErrUnavailable
	}
	return nil
}
