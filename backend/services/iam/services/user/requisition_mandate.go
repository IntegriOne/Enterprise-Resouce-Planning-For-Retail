package user

import (
	"context"
	"errors"
	"strings"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
)

const requisitionApprovalPermission = "human_resource.requisitions.approve"

// HasRequisitionApprovalMandate memeriksa mandat efektif untuk unit tertentu tanpa mengembalikan snapshot akun.
func (s *Service) HasRequisitionApprovalMandate(ctx context.Context, principalID, unitID string) (bool, error) {
	required, err := entity.NewScope(entity.ScopeTypeOrganizationUnit, &unitID)
	if err != nil {
		return false, err
	}
	snapshot, err := s.GetAccessSnapshot(ctx, principalID)
	if errors.Is(err, iamerrors.ErrUserNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if snapshot == nil || !strings.EqualFold(snapshot.UserID, strings.TrimSpace(principalID)) {
		return false, iamerrors.New(iamerrors.KindUnavailable, "iam.user.invalid_access_snapshot", "user access snapshot could not be verified")
	}
	if snapshot.UserKind != entity.UserKindEmployee || snapshot.UserStatus != entity.UserStatusActive {
		return false, nil
	}
	for _, permission := range snapshot.Permissions {
		if entity.NormalizeCode(permission.PermissionCode) == requisitionApprovalPermission &&
			permission.Scope.Type == entity.ScopeTypeOrganizationUnit && permission.Scope.Matches(required) {
			return true, nil
		}
	}
	return false, nil
}
