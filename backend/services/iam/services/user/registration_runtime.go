package user

import (
	"context"
	"errors"
	"strings"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
	"github.com/integrione/erp-retail/backend/services/iam/infrastructure/redisstore"
	"github.com/integrione/erp-retail/backend/services/iam/repository/pgstore"
)

// VerifyEmailStatus merepresentasikan status hasil verifikasi tautan email satu kali pakai.
type VerifyEmailStatus string

const (
	VerifyEmailStatusVerified    VerifyEmailStatus = "verified"
	VerifyEmailStatusExpired     VerifyEmailStatus = "expired"
	VerifyEmailStatusAlreadyUsed VerifyEmailStatus = "already_used"
	VerifyEmailStatusInvalid     VerifyEmailStatus = "invalid"
)

// VerifyEmailResult memuat status deterministik hasil konfirmasi verifikasi email melalui token.
type VerifyEmailResult struct {
	Status   VerifyEmailStatus
	User     *entity.User
	Identity *entity.UserIdentity
	Message  string
}

func (s *Service) recordUserSecurityChangedInTx(
	txCtx context.Context,
	userID string,
	bumpAuthEpoch bool,
	bumpPermRevision bool,
) error {
	if s.securityRuntime != nil && (bumpAuthEpoch || bumpPermRevision) {
		if _, _, err := s.securityRuntime.IncrementUserSecurityCounters(txCtx, userID, bumpAuthEpoch, bumpPermRevision); err != nil {
			return err
		}
	}
	if s.outboxCoord == nil && s.projector == nil {
		return nil
	}

	payload := redisstore.OutboxPayload{
		UserID:          userID,
		InvalidateAuthz: bumpPermRevision,
	}
	var evID string
	if s.outboxCoord != nil {
		var err error
		evID, err = s.outboxCoord.RecordInTx(txCtx, entity.OutboxEventUserSecurityChanged, "user", userID, payload)
		if err != nil {
			return err
		}
	}

	pgstore.RegisterAfterCommit(txCtx, func(postCtx context.Context) {
		u, err := s.users.SelectByID(postCtx, userID)
		if err != nil {
			return
		}
		if s.outboxCoord != nil {
			s.outboxCoord.CompleteAfterCommit(postCtx, evID, func(recEpoch int64) error {
				if s.projector == nil {
					return nil
				}
				return s.projector.ProjectUserSecurityChanged(postCtx, u, recEpoch, bumpPermRevision)
			})
			return
		}
		if s.projector != nil && s.projector.IsEnabled() {
			_ = s.projector.ProjectUserSecurityChanged(postCtx, u, 1, bumpPermRevision)
		}
	})
	return nil
}

// ResendValidationEmailToken menerbitkan tepat 1 token verifikasi email aktif terbaru baik untuk pengguna terautentikasi dengan pending email maupun pendaftar berstatus pending via email.
func (s *Service) ResendValidationEmailToken(ctx context.Context, in ResendValidationEmailInput) (*ResendValidationEmailResult, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}

	now := s.clock.Now()
	trimmedUserID := strings.TrimSpace(in.UserID)
	trimmedEmail := strings.TrimSpace(in.Email)

	// Kasus (1): pengguna login dengan cookie autentikasi — periksa pending email pada user_identities berdasarkan user_id.
	if trimmedUserID != "" {
		if err := entity.ValidateUUID("user_id", trimmedUserID); err != nil {
			return nil, err
		}
		var normFilterEmail string
		if trimmedEmail != "" {
			var err error
			normFilterEmail, err = entity.ValidateEmail(trimmedEmail)
			if err != nil {
				return nil, err
			}
		}

		var recipientName string
		var result *ResendValidationEmailResult
		err := s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
			u, err := s.users.SelectByID(txCtx, trimmedUserID)
			if err != nil {
				return err
			}
			recipientName = u.DisplayName
			if u.IsDeleted() {
				return iamerrors.ErrUserDeleted
			}
			if u.Status == entity.UserStatusSuspended {
				return iamerrors.ErrUserSuspended
			}
			if u.Status == entity.UserStatusDisabled {
				return iamerrors.ErrUserDisabled
			}

			idents, err := s.identities.SelectByUserID(txCtx, u.ID)
			if err != nil {
				return err
			}
			var pendingEmail *entity.UserIdentity
			var hasVerifiedEmail bool
			for _, item := range idents {
				if item.Kind != entity.IdentityKindEmail {
					continue
				}
				if !item.IsVerified {
					pendingEmail = item
					break
				}
				hasVerifiedEmail = true
			}
			if pendingEmail == nil {
				if hasVerifiedEmail {
					return iamerrors.ErrEmailAlreadyVerified
				}
				return iamerrors.ErrPendingEmailNotFound
			}
			if normFilterEmail != "" && pendingEmail.Value != normFilterEmail {
				return iamerrors.WithField(iamerrors.ErrPendingEmailNotFound, "email", "email does not match pending email verification")
			}

			issuedTok, err := s.issueOneTimeTokenForIdentity(txCtx, u.ID, &pendingEmail.ID, entity.TokenPurposeEmailVerification, s.cfg.IdentityVerificationTTL, now)
			if err != nil {
				return err
			}
			exp := issuedTok.Token.ExpiresAt
			result = &ResendValidationEmailResult{
				IsResend:          true,
				PendingEmail:      pendingEmail.Value,
				VerificationToken: issuedTok.PlaintextToken,
				ExpiresAt:         &exp,
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
		s.dispatchVerificationEmail(ctx, recipientName, result.PendingEmail, result.VerificationToken, s.cfg.IdentityVerificationTTL)
		return result, nil
	}

	// Kasus (2): pengguna belum login (registrasi PENDING) mengirim email; gunakan respons netral anti-enumerasi jika email tidak ada atau akun sudah aktif.
	if trimmedEmail == "" {
		return nil, iamerrors.WithField(iamerrors.ErrInvalidInput, "email", "email is required when unauthenticated")
	}
	normEmail, err := entity.ValidateEmail(trimmedEmail)
	if err != nil {
		return nil, err
	}

	ident, err := s.identities.FindUserIdentityByEmailKind(ctx, normEmail)
	if err != nil {
		if errors.Is(err, iamerrors.ErrIdentityNotFound) {
			return &ResendValidationEmailResult{IsResend: true}, nil
		}
		return nil, err
	}
	u, err := s.users.SelectByID(ctx, ident.UserID)
	if err != nil {
		if errors.Is(err, iamerrors.ErrUserNotFound) {
			return &ResendValidationEmailResult{IsResend: true}, nil
		}
		return nil, err
	}
	if u.IsDeleted() || u.Status != entity.UserStatusPending || ident.IsVerified {
		return &ResendValidationEmailResult{IsResend: true}, nil
	}

	var result *ResendValidationEmailResult
	err = s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		issuedTok, err := s.issueOneTimeTokenForIdentity(txCtx, u.ID, &ident.ID, entity.TokenPurposeEmailVerification, s.cfg.IdentityVerificationTTL, now)
		if err != nil {
			return err
		}
		exp := issuedTok.Token.ExpiresAt
		result = &ResendValidationEmailResult{
			IsResend:          true,
			PendingEmail:      ident.Value,
			VerificationToken: issuedTok.PlaintextToken,
			ExpiresAt:         &exp,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.dispatchVerificationEmail(ctx, u.DisplayName, result.PendingEmail, result.VerificationToken, s.cfg.IdentityVerificationTTL)
	return result, nil
}

// ConfirmEmailVerificationByToken memvalidasi token verifikasi email, mengaktifkan akun pending, atau mengganti email utama aktif saat pembaruan email terkonfirmasi.
func (s *Service) ConfirmEmailVerificationByToken(ctx context.Context, rawToken string) (*VerifyEmailResult, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}

	digest := s.tokenGen.Digest(strings.TrimSpace(rawToken))
	if len(digest) == 0 {
		return &VerifyEmailResult{
			Status:  VerifyEmailStatusInvalid,
			Message: "verification token is invalid",
		}, nil
	}

	tok, err := s.tokens.SelectByDigestAndPurpose(ctx, digest, entity.TokenPurposeEmailVerification)
	if err != nil {
		if errors.Is(err, iamerrors.ErrTokenNotFound) {
			return &VerifyEmailResult{
				Status:  VerifyEmailStatusInvalid,
				Message: "verification token is invalid",
			}, nil
		}
		return nil, err
	}

	now := s.clock.Now()
	if tok.Status == entity.TokenStatusConsumed || tok.UsedAt != nil {
		return &VerifyEmailResult{
			Status:  VerifyEmailStatusAlreadyUsed,
			Message: "verification link has already been used",
		}, nil
	}
	if tok.Status == entity.TokenStatusRevoked {
		return &VerifyEmailResult{
			Status:  VerifyEmailStatusInvalid,
			Message: "verification token is invalid",
		}, nil
	}
	if tok.Status == entity.TokenStatusExpired || !now.Before(tok.ExpiresAt) {
		return &VerifyEmailResult{
			Status:  VerifyEmailStatusExpired,
			Message: "verification link has expired",
		}, nil
	}

	var result *VerifyEmailResult
	err = s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		u, err := s.users.SelectByID(txCtx, tok.UserID)
		if err != nil {
			if errors.Is(err, iamerrors.ErrUserNotFound) {
				result = &VerifyEmailResult{
					Status:  VerifyEmailStatusInvalid,
					Message: "verification token is invalid",
				}
				return nil
			}
			return err
		}
		if err := u.CanAuthenticate(); err != nil {
			// Jangan membocorkan status deleted/suspended/disabled kepada pemanggil publik verifyEmail.
			if u.IsDeleted() || u.Status == entity.UserStatusSuspended || u.Status == entity.UserStatusDisabled {
				result = &VerifyEmailResult{
					Status:  VerifyEmailStatusInvalid,
					Message: "verification token is invalid",
				}
				return nil
			}
		}

		idents, err := s.identities.SelectByUserID(txCtx, tok.UserID)
		if err != nil {
			return err
		}

		var targetIdent *entity.UserIdentity
		if tok.IdentityID != nil {
			for _, item := range idents {
				if item.ID == *tok.IdentityID && item.Kind == entity.IdentityKindEmail {
					targetIdent = item
					break
				}
			}
		}
		if targetIdent == nil || targetIdent.IsVerified {
			result = &VerifyEmailResult{
				Status:  VerifyEmailStatusInvalid,
				Message: "verification token is invalid",
			}
			return nil
		}

		consumed, err := s.tokens.UpdateUsedAtIfUnused(txCtx, tok.ID, now)
		if err != nil {
			return err
		}
		if !consumed {
			result = &VerifyEmailResult{
				Status:  VerifyEmailStatusAlreadyUsed,
				Message: "verification link has already been used",
			}
			return nil
		}

		var supersededEmailIDs []string
		for _, item := range idents {
			if item.Kind == entity.IdentityKindEmail && item.ID != targetIdent.ID {
				supersededEmailIDs = append(supersededEmailIDs, item.ID)
			}
		}

		// Hapus email lama terlebih dahulu sebelum menandai email baru terverifikasi agar constraint satu email aktif per pengguna selalu terpenuhi.
		for _, oldID := range supersededEmailIDs {
			if err := s.identities.Delete(txCtx, oldID); err != nil {
				return err
			}
		}
		targetIdent.MarkVerified(now)
		targetIdent.SetPrimary(true, now)
		if err := s.identities.Update(txCtx, targetIdent); err != nil {
			return err
		}

		// Identitas username tidak memiliki kanal verifikasi terpisah sehingga ikut diaktifkan saat email terverifikasi.
		for _, item := range idents {
			if item.Kind == entity.IdentityKindUsername && !item.IsVerified {
				item.MarkVerified(now)
				if err := s.identities.Update(txCtx, item); err != nil {
					return err
				}
			}
		}

		if u.Status == entity.UserStatusPending {
			if err := s.requireEmployeeLinkForActivation(txCtx, u); err != nil {
				return err
			}
			if err := u.Activate(now); err != nil {
				return err
			}
			if err := s.users.Update(txCtx, u); err != nil {
				return err
			}
			if err := s.recordUserSecurityChangedInTx(txCtx, u.ID, true, true); err != nil {
				return err
			}
		} else if len(supersededEmailIDs) > 0 {
			if err := s.recordUserSecurityChangedInTx(txCtx, u.ID, false, false); err != nil {
				return err
			}
		}

		result = &VerifyEmailResult{
			Status:   VerifyEmailStatusVerified,
			User:     u,
			Identity: targetIdent,
			Message:  "email verified successfully",
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
