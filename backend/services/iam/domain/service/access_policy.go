package service

import (
	"strings"
	"time"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
)

// AuthorizeInput merangkum seluruh konteks yang diperlukan untuk mengambil keputusan otorisasi secara fail-closed.
type AuthorizeInput struct {
	User                   *entity.User
	Session                *entity.Session
	Snapshot               *entity.AccessSnapshot
	RequiredPermission     string
	RequiredScope          entity.Scope
	RequireSensitiveReauth bool
	Now                    time.Time
}

// AccessEvaluator mengevaluasi kelayakan akun, keaktifan sesi, kesegaran snapshot, izin berscope, serta aksi sensitif.
type AccessEvaluator struct {
	sessionPolicy SessionPolicy
}

// NewAccessEvaluator membentuk AccessEvaluator dengan SessionPolicy yang ditetapkan.
func NewAccessEvaluator(sessionPolicy SessionPolicy) AccessEvaluator {
	return AccessEvaluator{sessionPolicy: sessionPolicy}
}

// Authorize menegakkan evaluasi otorisasi fail-closed lintas user, session, snapshot, scope, dan re-autentikasi sensitif.
func (e AccessEvaluator) Authorize(input AuthorizeInput) error {
	if err := input.User.CanAuthenticate(); err != nil {
		return err
	}
	if err := input.Session.CheckStatus(input.Now); err != nil {
		return err
	}
	if input.Session.UserID != input.User.ID {
		return iamerrors.ErrUnauthorized
	}
	if err := e.sessionPolicy.ValidateSnapshotFreshness(input.Snapshot, input.Now); err != nil {
		return err
	}
	if input.Snapshot.UserID != input.User.ID || input.Snapshot.UserStatus != entity.UserStatusActive {
		return iamerrors.ErrForbidden
	}

	normalizedPerm := entity.NormalizeCode(input.RequiredPermission)
	if normalizedPerm == "" {
		// Untuk verifikasi re-autentikasi sensitif mandiri (tanpa kode izin RBAC), tegakkan kesegaran reauth & MFA sesi.
		if input.RequireSensitiveReauth {
			return e.sessionPolicy.ValidateSensitiveReauth(input.Session, input.Snapshot.MFAEnabled, input.Now)
		}
		return iamerrors.ErrInvalidPermissionCode
	}
	if !input.RequiredScope.IsValid() {
		return iamerrors.ErrInvalidScope
	}
	if entity.IsSelfServicePermissionCode(normalizedPerm) {
		// Tegakkan pembatasan bahwa izin *self tidak pernah bisa dievaluasi pada scope global atau user_id pihak lain.
		if input.RequiredScope.Type != entity.ScopeTypeSelf ||
			input.RequiredScope.ID == nil ||
			!strings.EqualFold(strings.TrimSpace(*input.RequiredScope.ID), strings.TrimSpace(input.User.ID)) {
			return iamerrors.ErrPermissionDenied
		}
	}

	matchedPerm, ok := input.Snapshot.HasPermission(normalizedPerm, input.RequiredScope)
	if !ok {
		return iamerrors.ErrPermissionDenied
	}

	if matchedPerm.IsSensitive || input.RequireSensitiveReauth {
		if err := e.sessionPolicy.ValidateSensitiveReauth(input.Session, input.Snapshot.MFAEnabled, input.Now); err != nil {
			if matchedPerm.IsSensitive {
				return iamerrors.ErrSensitiveReauthRequired
			}
			return err
		}
	}

	return nil
}
