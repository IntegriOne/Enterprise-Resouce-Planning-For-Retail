package service

import (
	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
)

// RequireScopeOwnerVerification menolak grant ke objek eksternal sampai kontrak pembuktian pemilik tersedia.
func RequireScopeOwnerVerification(scope entity.Scope) error {
	if !scope.IsValid() {
		return iamerrors.ErrInvalidScope
	}
	if scope.Type == entity.ScopeTypeGlobal || scope.Type == entity.ScopeTypeSelf {
		return nil
	}
	return iamerrors.New(iamerrors.KindUnavailable, iamerrors.Code("iam.scope.owner_verification_unavailable"), "scope owner verification is unavailable")
}
