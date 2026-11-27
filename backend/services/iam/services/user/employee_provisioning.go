package user

import (
	"context"
	"time"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
	"github.com/integrione/erp-retail/backend/shared/authcontext"
)

func (s *Service) claimEmployeeProvisioning(ctx context.Context, input CreateUserInput, now time.Time) (*entity.User, error) {
	if s.employeeProvisioning == nil {
		return nil, iamerrors.ErrUnavailable
	}
	if !entity.IsValidUUID(input.ProvisioningRequestID) || !entity.IsValidUUID(input.WorkerID) {
		return nil, iamerrors.ErrInvalidInput
	}
	caller, ok := authcontext.FromContext(ctx)
	if !ok || !caller.Authenticated || caller.SessionID == "" {
		return nil, iamerrors.ErrUnauthorized
	}
	if caller.UserKind != string(entity.UserKindEmployee) {
		return nil, iamerrors.ErrForbidden
	}
	snapshot, err := s.users.SelectAccessSnapshotByUserID(ctx, caller.PrincipalID, now)
	if err != nil {
		return nil, err
	}
	if snapshot == nil || snapshot.AuthEpoch != caller.AuthEpoch || snapshot.PermRevision != caller.PermRevision {
		return nil, iamerrors.ErrForbidden
	}
	if _, allowed := snapshot.HasPermission("iam.users.write", entity.GlobalScope()); !allowed {
		return nil, iamerrors.ErrForbidden
	}
	request, err := s.employeeProvisioning.SelectByIDForUpdate(ctx, input.ProvisioningRequestID)
	if err != nil {
		return nil, err
	}
	if request.WorkerID != input.WorkerID {
		return nil, iamerrors.ErrConflict
	}
	if request.Status == entity.EmployeeProvisioningPending {
		_, err := s.employeeProvisioning.UpdateClaim(ctx, request.ID, caller.PrincipalID, now)
		return nil, err
	}
	if request.ProcessedByUserID == nil || *request.ProcessedByUserID != caller.PrincipalID || request.AccountUserID == nil ||
		(request.Status != entity.EmployeeProvisioningProcessing && request.Status != entity.EmployeeProvisioningLinked) {
		return nil, iamerrors.ErrConflict
	}
	return s.users.SelectByID(ctx, *request.AccountUserID)
}

func (s *Service) associateProvisionedAccount(ctx context.Context, requestID, accountID string) error {
	caller, _ := authcontext.FromContext(ctx)
	return s.employeeProvisioning.UpdateAccount(ctx, requestID, caller.PrincipalID, accountID)
}

func (s *Service) requireEmployeeLinkForActivation(ctx context.Context, u *entity.User) error {
	if u.Kind != entity.UserKindEmployee || s.employeeProvisioning == nil {
		return nil
	}
	unlinked, err := s.employeeProvisioning.HasUnlinkedAccount(ctx, u.ID)
	if err != nil {
		return err
	}
	if unlinked {
		return iamerrors.ErrConflict
	}
	return nil
}
