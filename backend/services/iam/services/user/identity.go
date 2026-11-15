package user

import (
	"context"
	"errors"
	"time"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
)

// ListIdentities mengambil seluruh identitas login yang terhubung ke akun pengguna.
func (s *Service) ListIdentities(ctx context.Context, userID string) ([]*entity.UserIdentity, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("user_id", userID); err != nil {
		return nil, err
	}
	return s.identities.SelectByUserID(ctx, userID)
}

// AddIdentity menambahkan pengenal login baru (email/username) setelah memvalidasi keunikannya lintas pengguna.
func (s *Service) AddIdentity(ctx context.Context, in AddIdentityInput) (*entity.UserIdentity, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("user_id", in.UserID); err != nil {
		return nil, err
	}

	identID, err := s.idGen.NewUUID()
	if err != nil {
		return nil, iamerrors.Wrap(err, iamerrors.KindUnavailable, iamerrors.Code("iam.identity.id_generation_failed"), "failed to generate identity id")
	}

	now := s.clock.Now()
	ident, err := entity.NewUserIdentity(identID, in.UserID, in.Kind, in.Value, in.IsPrimary, now)
	if err != nil {
		return nil, err
	}
	// Username tidak memiliki alur verifikasi terpisah sehingga langsung aktif begitu ditambahkan.
	if ident.Kind == entity.IdentityKindUsername {
		ident.MarkVerified(now)
	}

	err = s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		u, err := s.users.SelectByID(txCtx, in.UserID)
		if err != nil {
			return err
		}
		if u.IsDeleted() {
			return iamerrors.ErrUserDeleted
		}

		// Pastikan maksimal satu identitas email dan satu identitas username per pengguna serta nilai belum dipakai akun lain.
		if err := s.ensureSingleIdentityKindPerUser(txCtx, in.UserID, ident.Kind); err != nil {
			return err
		}
		if ident.Kind == entity.IdentityKindEmail {
			err = s.ensureEmailAndUsernameAvailable(txCtx, ident.Value, "")
		} else {
			err = s.ensureEmailAndUsernameAvailable(txCtx, "", ident.Value)
		}
		if err != nil {
			return err
		}

		if in.IsPrimary {
			if err := s.demotePrimaryIdentities(txCtx, in.UserID, now); err != nil {
				return err
			}
		}
		return s.identities.Insert(txCtx, ident)
	})
	if err != nil {
		return nil, err
	}
	return ident, nil
}

// RequestIdentityVerification menerbitkan token verifikasi email sekali pakai untuk identitas tertentu.
func (s *Service) RequestIdentityVerification(ctx context.Context, identityID string) (*IssuedTokenResult, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("identity_id", identityID); err != nil {
		return nil, err
	}

	ident, err := s.identities.SelectByID(ctx, identityID)
	if err != nil {
		return nil, err
	}
	u, err := s.users.SelectByID(ctx, ident.UserID)
	if err != nil {
		return nil, err
	}
	if u.IsDeleted() {
		return nil, iamerrors.ErrUserDeleted
	}

	var issued *IssuedTokenResult
	err = s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		tokRes, issueErr := s.issueOneTimeTokenForIdentity(txCtx, ident.UserID, &ident.ID, entity.TokenPurposeEmailVerification, s.cfg.IdentityVerificationTTL, s.clock.Now())
		if issueErr != nil {
			return issueErr
		}
		issued = tokRes
		return nil
	})
	if err != nil {
		return nil, err
	}
	return issued, nil
}

// UpdateEmail menyimpan alamat email baru sebagai pending email sekunder terpisah sambil mempertahankan email utama lama hingga terverifikasi, serta menjaga tepat 1 token verifikasi aktif.
func (s *Service) UpdateEmail(ctx context.Context, in UpdateEmailInput) (*UpdateEmailResult, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("user_id", in.UserID); err != nil {
		return nil, err
	}
	normEmail, err := entity.ValidateEmail(in.NewEmail)
	if err != nil {
		return nil, err
	}
	if in.CurrentPassword == "" {
		return nil, iamerrors.WithField(iamerrors.ErrInvalidInput, "current_password", "current password is required")
	}

	u, err := s.users.SelectByID(ctx, in.UserID)
	if err != nil {
		return nil, err
	}
	if err := u.CanAuthenticate(); err != nil {
		return nil, err
	}

	cred, err := s.credentials.SelectByUserID(ctx, in.UserID)
	if err != nil {
		return nil, err
	}
	matched, err := s.passwordHasher.Verify(in.CurrentPassword, cred.PasswordHash())
	if err != nil {
		return nil, err
	}
	if !matched {
		return nil, iamerrors.ErrInvalidCredentials
	}

	now := s.clock.Now()
	var result *UpdateEmailResult
	err = s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		allIdents, err := s.identities.SelectByUserID(txCtx, in.UserID)
		if err != nil && !errors.Is(err, iamerrors.ErrIdentityNotFound) {
			return err
		}

		var pendingIdent *entity.UserIdentity
		for _, item := range allIdents {
			if item.Kind != entity.IdentityKindEmail {
				continue
			}
			if item.IsVerified && item.Value == normEmail {
				return iamerrors.WithField(iamerrors.ErrSameEmail, "email", "new email matches current active email")
			}
			if !item.IsVerified {
				pendingIdent = item
			}
		}

		existing, findErr := s.identities.FindUserIdentityByEmailKind(txCtx, normEmail)
		if findErr == nil {
			if existing.UserID != in.UserID || (pendingIdent == nil || existing.ID != pendingIdent.ID) {
				return iamerrors.WithField(iamerrors.ErrIdentityAlreadyExists, "email", "email is already registered")
			}
		} else if !errors.Is(findErr, iamerrors.ErrIdentityNotFound) {
			return findErr
		}

		if pendingIdent != nil {
			pendingIdent.SetPrimary(false, now)
			if err := pendingIdent.UpdateValue(normEmail, now); err != nil {
				return err
			}
			if err := s.identities.Update(txCtx, pendingIdent); err != nil {
				return err
			}
		} else {
			identID, err := s.idGen.NewUUID()
			if err != nil {
				return iamerrors.Wrap(err, iamerrors.KindUnavailable, iamerrors.Code("iam.identity.id_generation_failed"), "failed to generate identity id")
			}
			pendingIdent, err = entity.NewUserIdentity(identID, in.UserID, entity.IdentityKindEmail, normEmail, false, now)
			if err != nil {
				return err
			}
			if err := s.identities.Insert(txCtx, pendingIdent); err != nil {
				return err
			}
		}

		issuedTok, err := s.issueOneTimeTokenForIdentity(txCtx, in.UserID, &pendingIdent.ID, entity.TokenPurposeEmailVerification, s.cfg.IdentityVerificationTTL, now)
		if err != nil {
			return err
		}
		result = &UpdateEmailResult{
			PendingIdentity:   pendingIdent,
			VerificationToken: issuedTok.PlaintextToken,
			ExpiresAt:         issuedTok.Token.ExpiresAt,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.dispatchVerificationEmail(ctx, u.DisplayName, result.PendingIdentity.Value, result.VerificationToken, s.cfg.IdentityVerificationTTL)
	return result, nil
}

// UpdateUsername mengganti atau menambahkan username terverifikasi secara langsung agar username lama segera bebas digunakan akun lain.
func (s *Service) UpdateUsername(ctx context.Context, in UpdateUsernameInput) (*entity.UserIdentity, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("user_id", in.UserID); err != nil {
		return nil, err
	}
	normUsername, err := entity.ValidateUsername(in.Username)
	if err != nil {
		return nil, err
	}

	now := s.clock.Now()
	var updated *entity.UserIdentity
	err = s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		u, err := s.users.SelectByID(txCtx, in.UserID)
		if err != nil {
			return err
		}
		if err := u.CanAuthenticate(); err != nil {
			return err
		}

		allIdents, err := s.identities.SelectByUserID(txCtx, in.UserID)
		if err != nil && !errors.Is(err, iamerrors.ErrIdentityNotFound) {
			return err
		}

		var currentUsername *entity.UserIdentity
		var hasPrimary bool
		for _, item := range allIdents {
			if item.IsPrimary {
				hasPrimary = true
			}
			if item.Kind == entity.IdentityKindUsername && currentUsername == nil {
				currentUsername = item
			}
		}

		if currentUsername != nil && currentUsername.Value == normUsername {
			return iamerrors.WithField(iamerrors.ErrSameUsername, "username", "new username must differ from current username")
		}

		if err := s.ensureEmailAndUsernameAvailable(txCtx, "", normUsername); err != nil {
			return err
		}

		if currentUsername != nil {
			// Perbarui baris username di tempat agar username lama langsung terhapus dari user_identities dan dapat dipakai pengguna lain.
			if err := currentUsername.UpdateUsernameValue(normUsername, now); err != nil {
				return err
			}
			if err := s.identities.Update(txCtx, currentUsername); err != nil {
				return err
			}
			updated = currentUsername
		} else {
			identID, err := s.idGen.NewUUID()
			if err != nil {
				return iamerrors.Wrap(err, iamerrors.KindUnavailable, iamerrors.Code("iam.identity.id_generation_failed"), "failed to generate identity id")
			}
			newIdent, err := entity.NewUserIdentity(identID, in.UserID, entity.IdentityKindUsername, normUsername, !hasPrimary, now)
			if err != nil {
				return err
			}
			newIdent.MarkVerified(now)
			if err := s.identities.Insert(txCtx, newIdent); err != nil {
				return err
			}
			updated = newIdent
		}

		return s.recordUserSecurityChangedInTx(txCtx, u.ID, false, false)
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}

// ensureEmailAndUsernameAvailable memeriksa ketersediaan username terlebih dahulu lalu email per kind sebelum pembuatan atau perubahan identitas.
func (s *Service) ensureEmailAndUsernameAvailable(ctx context.Context, normEmail, normUsername string) error {
	// Periksa username terlebih dahulu agar ketersediaan username tidak menjadi oracle eksistensi email.
	if normUsername != "" {
		existingUsername, err := s.identities.FindUserIdentityByUsernameKind(ctx, normUsername)
		if err == nil && existingUsername != nil {
			return iamerrors.WithField(iamerrors.ErrIdentityAlreadyExists, "username", "username has been taken")
		}
		if err != nil && !errors.Is(err, iamerrors.ErrIdentityNotFound) {
			return err
		}
	}
	if normEmail != "" {
		existingEmail, err := s.identities.FindUserIdentityByEmailKind(ctx, normEmail)
		if err == nil && existingEmail != nil {
			return iamerrors.WithField(iamerrors.ErrIdentityAlreadyExists, "email", "email is already registered")
		}
		if err != nil && !errors.Is(err, iamerrors.ErrIdentityNotFound) {
			return err
		}
	}
	return nil
}

func (s *Service) ensureSingleIdentityKindPerUser(ctx context.Context, userID string, kind entity.IdentityKind) error {
	existing, err := s.identities.SelectByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, iamerrors.ErrIdentityNotFound) {
			return nil
		}
		return err
	}
	for _, item := range existing {
		if item.Kind != kind {
			continue
		}
		if kind == entity.IdentityKindEmail {
			return iamerrors.WithField(iamerrors.ErrIdentityLimitExceeded, "email", "user may only have one email identity")
		}
		return iamerrors.WithField(iamerrors.ErrIdentityLimitExceeded, "username", "user may only have one username identity")
	}
	return nil
}

// DeletePendingEmail menghapus pending email sekunder milik pengguna terautentikasi dan mencabut token verifikasi aktifnya tanpa mengubah email utama aktif.
func (s *Service) DeletePendingEmail(ctx context.Context, userID string) (bool, error) {
	if err := s.ensureReady(); err != nil {
		return false, err
	}
	if err := entity.ValidateUUID("user_id", userID); err != nil {
		return false, err
	}

	err := s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		// Pastikan hanya pengguna berstatus ACTIVE yang dapat menghapus pending email secara mandiri.
		if _, err := s.loadAuthenticatableUser(txCtx, userID); err != nil {
			return err
		}

		allIdents, err := s.identities.SelectByUserID(txCtx, userID)
		if err != nil {
			return err
		}
		var pendingIdent *entity.UserIdentity
		for _, item := range allIdents {
			if item.Kind == entity.IdentityKindEmail && !item.IsVerified && !item.IsPrimary {
				pendingIdent = item
				break
			}
		}
		if pendingIdent == nil {
			return iamerrors.ErrPendingEmailNotFound
		}

		if _, err := s.tokens.UpdateActiveStatusByUserAndPurpose(txCtx, userID, entity.TokenPurposeEmailVerification, entity.TokenStatusRevoked); err != nil {
			return err
		}
		return s.identities.Delete(txCtx, pendingIdent.ID)
	})
	if err != nil {
		return false, err
	}
	return true, nil
}

// ConfirmIdentityVerification memverifikasi token kepemilikan identitas dan mengaktifkan pengguna pending bila diminta.
func (s *Service) ConfirmIdentityVerification(
	ctx context.Context,
	identityID string,
	rawToken string,
	activatePendingUser bool,
) (*entity.UserIdentity, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("identity_id", identityID); err != nil {
		return nil, err
	}

	digest := s.tokenGen.Digest(rawToken)
	if len(digest) == 0 {
		return nil, iamerrors.WithField(iamerrors.ErrInvalidInput, "token", "verification token is required")
	}

	tok, err := s.tokens.SelectByDigestAndPurpose(ctx, digest, entity.TokenPurposeEmailVerification)
	if err != nil {
		return nil, err
	}
	now := s.clock.Now()
	if err := tok.ValidateForUse(entity.TokenPurposeEmailVerification, now); err != nil {
		return nil, err
	}

	var verified *entity.UserIdentity
	err = s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		ident, err := s.identities.SelectByID(txCtx, identityID)
		if err != nil {
			return err
		}
		if ident.UserID != tok.UserID {
			return iamerrors.ErrForbidden
		}

		consumed, err := s.tokens.UpdateUsedAtIfUnused(txCtx, tok.ID, now)
		if err != nil {
			return err
		}
		if !consumed {
			return iamerrors.ErrTokenAlreadyUsed
		}

		if ident.Kind == entity.IdentityKindEmail {
			allIdents, listErr := s.identities.SelectByUserID(txCtx, ident.UserID)
			if listErr != nil && !errors.Is(listErr, iamerrors.ErrIdentityNotFound) {
				return listErr
			}
			// Hapus email lama terlebih dahulu sebelum verifikasi email baru agar indeks unik parsial tetap terpenuhi.
			for _, item := range allIdents {
				if item.Kind == entity.IdentityKindEmail && item.ID != ident.ID {
					if err := s.identities.Delete(txCtx, item.ID); err != nil {
						return err
					}
				}
			}
			for _, item := range allIdents {
				if item.Kind == entity.IdentityKindUsername && !item.IsVerified {
					item.MarkVerified(now)
					if err := s.identities.Update(txCtx, item); err != nil {
						return err
					}
				}
			}
		}

		ident.MarkVerified(now)
		if ident.Kind == entity.IdentityKindEmail {
			ident.SetPrimary(true, now)
		}
		if err := s.identities.Update(txCtx, ident); err != nil {
			return err
		}

		if activatePendingUser {
			u, err := s.users.SelectByID(txCtx, ident.UserID)
			if err != nil {
				return err
			}
			if u.Status == entity.UserStatusPending {
				if err := u.Activate(now); err != nil {
					return err
				}
				if err := s.users.Update(txCtx, u); err != nil {
					return err
				}
				if err := s.recordUserSecurityChangedInTx(txCtx, u.ID, true, true); err != nil {
					return err
				}
			}
		}

		verified = ident
		return nil
	})
	if err != nil {
		return nil, err
	}
	return verified, nil
}

// SetPrimaryIdentity menetapkan salah satu identitas terverifikasi milik pengguna sebagai identitas utama.
func (s *Service) SetPrimaryIdentity(ctx context.Context, userID, identityID string) (*entity.UserIdentity, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("user_id", userID); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("identity_id", identityID); err != nil {
		return nil, err
	}

	var target *entity.UserIdentity
	err := s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		ident, err := s.identities.SelectByID(txCtx, identityID)
		if err != nil {
			return err
		}
		if ident.UserID != userID {
			return iamerrors.ErrForbidden
		}
		if !ident.IsVerified {
			return iamerrors.ErrIdentityNotVerified
		}

		now := s.clock.Now()
		if err := s.demotePrimaryIdentities(txCtx, userID, now); err != nil {
			return err
		}
		ident.SetPrimary(true, now)
		if err := s.identities.Update(txCtx, ident); err != nil {
			return err
		}
		target = ident
		return nil
	})
	if err != nil {
		return nil, err
	}
	return target, nil
}

// RemoveIdentity menghapus identitas sekunder sambil melindungi identitas terakhir atau identitas utama agar tidak terhapus langsung.
func (s *Service) RemoveIdentity(ctx context.Context, userID, identityID string) error {
	if err := s.ensureReady(); err != nil {
		return err
	}
	if err := entity.ValidateUUID("user_id", userID); err != nil {
		return err
	}
	if err := entity.ValidateUUID("identity_id", identityID); err != nil {
		return err
	}

	return s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		all, err := s.identities.SelectByUserID(txCtx, userID)
		if err != nil {
			return err
		}
		var target *entity.UserIdentity
		for _, item := range all {
			if item.ID == identityID {
				target = item
				break
			}
		}
		if target == nil {
			return iamerrors.ErrIdentityNotFound
		}
		if len(all) <= 1 {
			return iamerrors.New(iamerrors.KindConflict, iamerrors.Code("iam.identity.cannot_remove_last"), "cannot remove user's last remaining login identity")
		}
		if target.IsPrimary {
			return iamerrors.New(iamerrors.KindConflict, iamerrors.Code("iam.identity.cannot_remove_primary"), "assign another primary identity before removing this identity")
		}
		if target.Kind == entity.IdentityKindEmail && !target.IsVerified {
			if _, err := s.tokens.UpdateActiveStatusByUserAndPurpose(txCtx, userID, entity.TokenPurposeEmailVerification, entity.TokenStatusRevoked); err != nil {
				return err
			}
		}
		return s.identities.Delete(txCtx, identityID)
	})
}

func (s *Service) demotePrimaryIdentities(ctx context.Context, userID string, now time.Time) error {
	all, err := s.identities.SelectByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, iamerrors.ErrIdentityNotFound) {
			return nil
		}
		return err
	}
	for _, item := range all {
		if item.IsPrimary {
			item.SetPrimary(false, now)
			if err := s.identities.Update(ctx, item); err != nil {
				return err
			}
		}
	}
	return nil
}
