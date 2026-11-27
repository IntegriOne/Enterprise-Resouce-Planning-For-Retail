package service

import (
	"context"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
)

// ScopeOwnerVerifier memverifikasi keberadaan dan kelayakan objek target scope pada layanan pemilik secara fail-closed.
type ScopeOwnerVerifier interface {
	VerifyScope(ctx context.Context, scope entity.Scope) error
}

// VerifyScope memvalidasi invarian scope dan memverifikasi target eksternal melalui verifier yang tersedia.
func VerifyScope(ctx context.Context, verifier ScopeOwnerVerifier, scope entity.Scope) error {
	if !scope.IsValid() {
		return iamerrors.ErrInvalidScope
	}
	if scope.Type == entity.ScopeTypeGlobal || scope.Type == entity.ScopeTypeSelf {
		return nil
	}
	if verifier == nil {
		return iamerrors.ErrScopeOwnerVerificationUnavailable
	}
	return verifier.VerifyScope(ctx, scope)
}

// RequireScopeOwnerVerification menolak grant ke objek eksternal bila belum ada kontrak pembuktian pemilik yang terhubung.
func RequireScopeOwnerVerification(scope entity.Scope) error {
	return VerifyScope(context.Background(), nil, scope)
}
