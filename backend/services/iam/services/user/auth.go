package user

import (
	"context"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
	"github.com/integrione/erp-retail/backend/services/iam/services/session"
)

const oneTimeTokenBytes = 32

// AuthenticatePassword memverifikasi login berbasis email/username, menegakkan lockout, serta menangani transisi ke MFA bila aktif.
func (s *Service) AuthenticatePassword(ctx context.Context, in AuthenticatePasswordInput) (*AuthenticateResult, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	if strings.TrimSpace(in.Identifier) == "" {
		return nil, iamerrors.WithField(iamerrors.ErrInvalidInput, "identifier", "login identifier is required")
	}
	if in.Password == "" {
		return nil, iamerrors.WithField(iamerrors.ErrInvalidInput, "password", "password is required")
	}

	kind, normalizedValue, err := parseLoginIdentifier(in.Identifier)
	if err != nil {
		s.verifyDummyPassword(in.Password)
		return nil, iamerrors.ErrInvalidCredentials
	}

	var ident *entity.UserIdentity
	if kind == entity.IdentityKindEmail {
		ident, err = s.identities.FindUserIdentityByEmailKind(ctx, normalizedValue)
	} else {
		ident, err = s.identities.FindUserIdentityByUsernameKind(ctx, normalizedValue)
	}
	if err != nil {
		if errors.Is(err, iamerrors.ErrIdentityNotFound) {
			s.verifyDummyPassword(in.Password)
			return nil, iamerrors.ErrInvalidCredentials
		}
		return nil, err
	}

	u, err := s.users.SelectByID(ctx, ident.UserID)
	if err != nil {
		if errors.Is(err, iamerrors.ErrUserNotFound) {
			s.verifyDummyPassword(in.Password)
			return nil, iamerrors.ErrInvalidCredentials
		}
		return nil, err
	}
	if u.IsDeleted() {
		s.verifyDummyPassword(in.Password)
		if s.logger != nil {
			s.logger.Warn("authentication rejected", slog.String("reason", "user_deleted"), slog.String("user_id", u.ID))
		}
		return nil, iamerrors.ErrInvalidCredentials
	}

	cred, err := s.credentials.SelectByUserID(ctx, u.ID)
	if err != nil {
		if errors.Is(err, iamerrors.ErrCredentialNotFound) {
			s.verifyDummyPassword(in.Password)
			return nil, iamerrors.ErrInvalidCredentials
		}
		return nil, err
	}

	now := s.clock.Now()
	if cred.IsLocked(now) {
		// Tetap jalankan verifikasi dummy constant-time agar penyerang tidak dapat mendeteksi status lock lewat timing.
		s.verifyDummyPassword(in.Password)
		if s.logger != nil {
			s.logger.Warn("authentication rejected", slog.String("reason", "account_locked"), slog.String("user_id", u.ID))
		}
		return nil, iamerrors.ErrInvalidCredentials
	}

	matched, err := s.passwordHasher.Verify(in.Password, cred.PasswordHash())
	if err != nil {
		return nil, err
	}
	if !matched {
		attempts, lockedUntil, lockErr := s.credentials.RecordFailedAttemptAtomic(ctx, u.ID, s.cfg.MaxFailedAttempts, s.cfg.LockDuration, now)
		if lockErr != nil {
			return nil, lockErr
		}
		if s.logger != nil {
			s.logger.Warn("authentication rejected",
				slog.String("reason", "invalid_password"),
				slog.String("user_id", u.ID),
				slog.Int("failed_attempts", attempts),
				slog.Bool("locked", lockedUntil != nil && now.Before(*lockedUntil)),
			)
		}
		return nil, iamerrors.ErrInvalidCredentials
	}

	if err := u.CanAuthenticate(); err != nil {
		if s.logger != nil {
			s.logger.Warn("authentication rejected", slog.String("reason", "user_status"), slog.String("user_id", u.ID), slog.String("status", u.Status.String()))
		}
		return nil, iamerrors.ErrInvalidCredentials
	}
	if !ident.IsVerified {
		if s.logger != nil {
			s.logger.Warn("authentication rejected", slog.String("reason", "identity_not_verified"), slog.String("user_id", u.ID))
		}
		return nil, iamerrors.ErrInvalidCredentials
	}

	hasMFA, err := s.hasConfirmedTOTP(ctx, u.ID)
	if err != nil {
		return nil, err
	}

	var result *AuthenticateResult
	err = s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		if cred.FailedLoginAttempts > 0 || cred.LockedUntil != nil {
			if err := s.credentials.ResetFailedAttemptsAtomic(txCtx, u.ID, now); err != nil {
				return err
			}
		}

		if hasMFA {
			issuedTok, err := s.issueOneTimeToken(txCtx, u.ID, entity.TokenPurposeMFAChallenge, s.cfg.MFAChallengeTTL, now)
			if err != nil {
				return err
			}
			exp := issuedTok.Token.ExpiresAt
			result = &AuthenticateResult{
				Status:                AuthStatusMFARequired,
				User:                  u,
				MustChangePassword:    cred.MustChangePassword,
				MFAChallengeToken:     issuedTok.PlaintextToken,
				MFAChallengeExpiresAt: &exp,
			}
			return nil
		}

		issuedSess, redirectPath, err := s.finalizeAuthenticatedLogin(txCtx, u, false, in.IPAddress, in.UserAgent, now)
		if err != nil {
			return err
		}
		result = &AuthenticateResult{
			Status:             AuthStatusAuthenticated,
			User:               u,
			IssuedSession:      issuedSess,
			RedirectPath:       redirectPath,
			MustChangePassword: cred.MustChangePassword,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// VerifyMFAChallenge menuntaskan tantangan MFA memakai kode TOTP sekali pakai anti-replay lalu menerbitkan sesi penuh.
func (s *Service) VerifyMFAChallenge(ctx context.Context, in VerifyMFAChallengeInput) (*AuthenticateResult, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}

	tok, u, now, err := s.loadValidMFAChallengeUser(ctx, in.ChallengeToken)
	if err != nil {
		if isAmbiguousMFAChallengeError(err) {
			return nil, iamerrors.ErrInvalidTOTPCode
		}
		return nil, err
	}

	factor, err := s.totpFactors.SelectByUserID(ctx, u.ID)
	if err != nil {
		if errors.Is(err, iamerrors.ErrTOTPNotFound) {
			return nil, iamerrors.ErrInvalidTOTPCode
		}
		return nil, err
	}
	if !factor.IsEnabled || factor.ConfirmedAt == nil {
		return nil, iamerrors.ErrInvalidTOTPCode
	}

	secret, err := s.secretCipher.Decrypt(factor.EncryptedSecret(), factor.KeyVersion)
	if err != nil {
		return nil, err
	}
	defer clear(secret)

	matched, step, err := s.totpProvider.VerifyCode(secret, in.Code, now)
	if err != nil {
		return nil, err
	}
	if !matched {
		return nil, iamerrors.ErrInvalidTOTPCode
	}
	if err := factor.CanUseStep(step); err != nil {
		if errors.Is(err, iamerrors.ErrTOTPReplayDetected) {
			if s.logger != nil {
				s.logger.Warn("mfa challenge rejected", slog.String("reason", "totp_replay"), slog.String("user_id", u.ID))
			}
			return nil, iamerrors.ErrInvalidTOTPCode
		}
		return nil, err
	}

	var result *AuthenticateResult
	err = s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		stepAccepted, err := s.totpFactors.UpdateLastUsedStepIfGreater(txCtx, factor.ID, step, now)
		if err != nil {
			return err
		}
		if !stepAccepted {
			if s.logger != nil {
				s.logger.Warn("mfa challenge rejected", slog.String("reason", "totp_replay_race"), slog.String("user_id", u.ID))
			}
			return iamerrors.ErrInvalidTOTPCode
		}

		consumed, err := s.tokens.UpdateUsedAtIfUnused(txCtx, tok.ID, now)
		if err != nil {
			return err
		}
		if !consumed {
			if s.logger != nil {
				s.logger.Warn("mfa challenge rejected", slog.String("reason", "challenge_consumed"), slog.String("user_id", u.ID))
			}
			return iamerrors.ErrInvalidTOTPCode
		}

		cred, err := s.credentials.SelectByUserID(txCtx, u.ID)
		if err != nil {
			return err
		}
		issuedSess, redirectPath, err := s.finalizeAuthenticatedLogin(txCtx, u, true, in.IPAddress, in.UserAgent, now)
		if err != nil {
			return err
		}
		result = &AuthenticateResult{
			Status:             AuthStatusAuthenticated,
			User:               u,
			IssuedSession:      issuedSess,
			RedirectPath:       redirectPath,
			MustChangePassword: cred.MustChangePassword,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// VerifyRecoveryCodeChallenge menuntaskan tantangan MFA memakai kode cadangan pemulihan sekali pakai lalu menerbitkan sesi penuh.
func (s *Service) VerifyRecoveryCodeChallenge(ctx context.Context, in VerifyRecoveryCodeInput) (*AuthenticateResult, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	normalizedCode := strings.ToUpper(strings.TrimSpace(in.RecoveryCode))
	if normalizedCode == "" {
		return nil, iamerrors.ErrInvalidRecoveryCode
	}

	tok, u, now, err := s.loadValidMFAChallengeUser(ctx, in.ChallengeToken)
	if err != nil {
		if isAmbiguousMFAChallengeError(err) {
			return nil, iamerrors.ErrInvalidRecoveryCode
		}
		return nil, err
	}

	candidateDigest := s.tokenGen.Digest(normalizedCode)
	if len(candidateDigest) == 0 {
		return nil, iamerrors.ErrInvalidRecoveryCode
	}
	candidateHex := hex.EncodeToString(candidateDigest)

	codes, err := s.recoveryCodes.SelectUnusedByUserID(ctx, u.ID)
	if err != nil {
		return nil, err
	}

	var matchedCode *entity.UserRecoveryCode
	for _, rc := range codes {
		if subtle.ConstantTimeCompare([]byte(rc.CodeHash()), []byte(candidateHex)) == 1 && matchedCode == nil {
			matchedCode = rc
		}
	}
	if matchedCode == nil {
		return nil, iamerrors.ErrInvalidRecoveryCode
	}

	var result *AuthenticateResult
	err = s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		usedCode, err := s.recoveryCodes.UpdateUsedAtIfUnused(txCtx, matchedCode.ID, now)
		if err != nil {
			return err
		}
		if !usedCode {
			return iamerrors.ErrInvalidRecoveryCode
		}

		consumedTok, err := s.tokens.UpdateUsedAtIfUnused(txCtx, tok.ID, now)
		if err != nil {
			return err
		}
		if !consumedTok {
			return iamerrors.ErrInvalidRecoveryCode
		}

		cred, err := s.credentials.SelectByUserID(txCtx, u.ID)
		if err != nil {
			return err
		}
		issuedSess, redirectPath, err := s.finalizeAuthenticatedLogin(txCtx, u, true, in.IPAddress, in.UserAgent, now)
		if err != nil {
			return err
		}
		result = &AuthenticateResult{
			Status:             AuthStatusAuthenticated,
			User:               u,
			IssuedSession:      issuedSess,
			RedirectPath:       redirectPath,
			MustChangePassword: cred.MustChangePassword,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// VerifySessionMFA memverifikasi TOTP untuk peningkatan status MFA (step-up) pada sesi yang sudah aktif.
func (s *Service) VerifySessionMFA(ctx context.Context, userID, sessionID, code string) (*session.CSRFTokenResult, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("user_id", userID); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("session_id", sessionID); err != nil {
		return nil, err
	}

	// Pastikan hanya pengguna berstatus ACTIVE yang dapat memverifikasi step-up MFA sesi.
	if _, err := s.loadAuthenticatableUser(ctx, userID); err != nil {
		return nil, err
	}

	factor, err := s.totpFactors.SelectByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !factor.IsEnabled || factor.ConfirmedAt == nil {
		return nil, iamerrors.ErrTOTPNotConfirmed
	}

	secret, err := s.secretCipher.Decrypt(factor.EncryptedSecret(), factor.KeyVersion)
	if err != nil {
		return nil, err
	}
	defer clear(secret)

	now := s.clock.Now()
	matched, step, err := s.totpProvider.VerifyCode(secret, code, now)
	if err != nil {
		return nil, err
	}
	if !matched {
		return nil, iamerrors.ErrInvalidTOTPCode
	}
	if err := factor.CanUseStep(step); err != nil {
		return nil, err
	}

	var rotated *session.CSRFTokenResult
	err = s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		accepted, err := s.totpFactors.UpdateLastUsedStepIfGreater(txCtx, factor.ID, step, now)
		if err != nil {
			return err
		}
		if !accepted {
			return iamerrors.ErrTOTPReplayDetected
		}
		res, err := s.sessions.MarkMFAVerified(txCtx, sessionID)
		if err != nil {
			return err
		}
		rotated = res
		return nil
	})
	if err != nil {
		return nil, err
	}
	return rotated, nil
}

// ChangePassword memverifikasi sandi lama, menegakkan kebijakan sandi baru, memperbarui hash, dan mencabut seluruh sesi lama.
func (s *Service) ChangePassword(ctx context.Context, in ChangePasswordInput) (*ChangePasswordResult, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("user_id", in.UserID); err != nil {
		return nil, err
	}
	if in.CurrentPassword == "" {
		return nil, iamerrors.WithField(iamerrors.ErrInvalidInput, "current_password", "current password is required")
	}

	// Pastikan hanya pengguna berstatus ACTIVE yang dapat mengganti kata sandi secara mandiri.
	u, err := s.loadAuthenticatableUser(ctx, in.UserID)
	if err != nil {
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

	revoked, err := s.applyNewPassword(ctx, u, cred, in.NewPassword, nil)
	if err != nil {
		return nil, err
	}
	return &ChangePasswordResult{
		Changed:           true,
		SessionsRevoked:   revoked,
		PasswordChangedAt: cred.PasswordChangedAt,
	}, nil
}

// RequestPasswordReset menerbitkan tepat 1 token pengaturan ulang kata sandi aktif untuk alamat email terverifikasi milik akun aktif.
func (s *Service) RequestPasswordReset(ctx context.Context, email string) (*IssuedTokenResult, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	normEmail, err := entity.ValidateEmail(email)
	if err != nil {
		return nil, err
	}

	ident, err := s.identities.FindUserIdentityByEmailKind(ctx, normEmail)
	if err != nil {
		return nil, err
	}
	if !ident.IsVerified {
		return nil, iamerrors.ErrIdentityNotVerified
	}
	u, err := s.users.SelectByID(ctx, ident.UserID)
	if err != nil {
		return nil, err
	}
	if err := u.CanAuthenticate(); err != nil {
		return nil, err
	}

	var issued *IssuedTokenResult
	err = s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		tokRes, issueErr := s.issueOneTimeTokenForIdentity(txCtx, u.ID, &ident.ID, entity.TokenPurposePasswordReset, s.cfg.PasswordResetTTL, s.clock.Now())
		if issueErr != nil {
			return issueErr
		}
		issued = tokRes
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.dispatchPasswordResetEmail(ctx, u.DisplayName, ident.Value, issued)
	return issued, nil
}

// ConfirmPasswordReset memvalidasi token sekali pakai, memperbarui kata sandi sesuai kebijakan, dan mencabut seluruh sesi aktif.
func (s *Service) ConfirmPasswordReset(ctx context.Context, in ConfirmPasswordResetInput) error {
	if err := s.ensureReady(); err != nil {
		return err
	}
	digest := s.tokenGen.Digest(in.RawToken)
	if len(digest) == 0 {
		return iamerrors.WithField(iamerrors.ErrInvalidInput, "token", "reset token is required")
	}

	tok, err := s.tokens.SelectByDigestAndPurpose(ctx, digest, entity.TokenPurposePasswordReset)
	if err != nil {
		return err
	}
	now := s.clock.Now()
	if err := tok.ValidateForUse(entity.TokenPurposePasswordReset, now); err != nil {
		return err
	}

	u, err := s.users.SelectByID(ctx, tok.UserID)
	if err != nil {
		if errors.Is(err, iamerrors.ErrUserNotFound) {
			return iamerrors.ErrTokenNotFound
		}
		return err
	}
	if err := u.CanAuthenticate(); err != nil {
		if s.logger != nil {
			s.logger.Warn("password reset confirmation rejected", slog.String("reason", "user_status"), slog.String("user_id", u.ID), slog.String("status", u.Status.String()))
		}
		return iamerrors.ErrTokenNotFound
	}
	cred, err := s.credentials.SelectByUserID(ctx, u.ID)
	if err != nil {
		return err
	}

	_, err = s.applyNewPassword(ctx, u, cred, in.NewPassword, &tok.ID)
	return err
}

func (s *Service) applyNewPassword(
	ctx context.Context,
	u *entity.User,
	cred *entity.UserCredential,
	newPassword string,
	consumeTokenID *string,
) (int64, error) {
	idents, err := s.identities.SelectByUserID(ctx, u.ID)
	if err != nil {
		return 0, err
	}
	forbidden := make([]string, 0, len(idents)+1)
	forbidden = append(forbidden, u.DisplayName)
	for _, ident := range idents {
		forbidden = append(forbidden, ident.Value)
	}

	if err := s.passwordPolicy.Validate(newPassword, forbidden...); err != nil {
		return 0, err
	}
	newHash, err := s.passwordHasher.Hash(newPassword)
	if err != nil {
		return 0, err
	}

	now := s.clock.Now()
	if err := cred.UpdatePasswordHash(newHash, now); err != nil {
		return 0, err
	}

	var revokedCount int64
	err = s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		if consumeTokenID != nil {
			consumed, err := s.tokens.UpdateUsedAtIfUnused(txCtx, *consumeTokenID, now)
			if err != nil {
				return err
			}
			if !consumed {
				return iamerrors.ErrTokenAlreadyUsed
			}
		}
		if err := s.credentials.Update(txCtx, cred); err != nil {
			return err
		}
		revoked, err := s.sessions.RevokeAllUserSessions(txCtx, u.ID, entity.SessionEndPasswordChanged)
		if err != nil {
			return err
		}
		revokedCount = revoked
		return nil
	})
	if err != nil {
		return 0, err
	}
	return revokedCount, nil
}

func (s *Service) loadValidMFAChallengeUser(ctx context.Context, rawChallenge string) (*entity.UserToken, *entity.User, time.Time, error) {
	digest := s.tokenGen.Digest(rawChallenge)
	if len(digest) == 0 {
		return nil, nil, time.Time{}, iamerrors.WithField(iamerrors.ErrInvalidInput, "challenge_token", "challenge token is required")
	}

	tok, err := s.tokens.SelectByDigestAndPurpose(ctx, digest, entity.TokenPurposeMFAChallenge)
	if err != nil {
		return nil, nil, time.Time{}, err
	}
	now := s.clock.Now()
	if err := tok.ValidateForUse(entity.TokenPurposeMFAChallenge, now); err != nil {
		return nil, nil, time.Time{}, err
	}

	u, err := s.users.SelectByID(ctx, tok.UserID)
	if err != nil {
		return nil, nil, time.Time{}, err
	}
	if err := u.CanAuthenticate(); err != nil {
		return nil, nil, time.Time{}, err
	}
	return tok, u, now, nil
}

const fallbackDummyArgon2Hash = "$argon2id$v=19$m=65536,t=3,p=2$AAAAAAAAAAAAAAAAAAAAAA$AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"

func (s *Service) verifyDummyPassword(password string) {
	if s == nil || s.passwordHasher == nil {
		return
	}
	s.dummyHashOnce.Do(func() {
		hashed, err := s.passwordHasher.Hash("Integrione!Dummy#2026$ConstantTime")
		if err == nil && strings.TrimSpace(hashed) != "" {
			s.dummyHash = hashed
			return
		}
		s.dummyHash = fallbackDummyArgon2Hash
	})
	_, _ = s.passwordHasher.Verify(password, s.dummyHash)
}

func isAmbiguousMFAChallengeError(err error) bool {
	return errors.Is(err, iamerrors.ErrTokenNotFound) ||
		errors.Is(err, iamerrors.ErrTokenExpired) ||
		errors.Is(err, iamerrors.ErrTokenAlreadyUsed) ||
		errors.Is(err, iamerrors.ErrTokenRevoked) ||
		errors.Is(err, iamerrors.ErrTokenInvalidPurpose) ||
		errors.Is(err, iamerrors.ErrInvalidTokenStatus) ||
		errors.Is(err, iamerrors.ErrUserNotFound) ||
		errors.Is(err, iamerrors.ErrUserDeleted) ||
		errors.Is(err, iamerrors.ErrUserSuspended) ||
		errors.Is(err, iamerrors.ErrUserDisabled) ||
		errors.Is(err, iamerrors.ErrUserInactive)
}

func (s *Service) finalizeAuthenticatedLogin(
	ctx context.Context,
	u *entity.User,
	mfaVerified bool,
	ipAddress string,
	userAgent string,
	now time.Time,
) (*session.IssuedSession, string, error) {
	u.RecordLogin(now)
	if err := s.users.UpdateLastLogin(ctx, u.ID, now); err != nil {
		return nil, "", err
	}

	issuedSess, err := s.sessions.Issue(ctx, session.IssueInput{
		UserID:      u.ID,
		MFAVerified: mfaVerified,
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
	})
	if err != nil {
		return nil, "", err
	}

	effectiveRoles, err := s.roles.SelectEffectiveByUserID(ctx, u.ID, now)
	if err != nil {
		return nil, "", err
	}
	redirectPath, err := s.redirect.ResolvePostLoginPath(u.Kind, effectiveRoles)
	if err != nil {
		return nil, "", err
	}
	return issuedSess, redirectPath, nil
}

func (s *Service) hasConfirmedTOTP(ctx context.Context, userID string) (bool, error) {
	factor, err := s.totpFactors.SelectByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, iamerrors.ErrTOTPNotFound) {
			return false, nil
		}
		return false, err
	}
	return factor.IsEnabled && factor.ConfirmedAt != nil, nil
}

func (s *Service) issueOneTimeToken(
	ctx context.Context,
	userID string,
	purpose entity.TokenPurpose,
	ttl time.Duration,
	now time.Time,
) (*IssuedTokenResult, error) {
	return s.issueOneTimeTokenForIdentity(ctx, userID, nil, purpose, ttl, now)
}

// issueOneTimeTokenForIdentity memastikan maksimal 1 token active per (user_id, purpose) dengan menandai token aktif sebelumnya menjadi expired sebelum menyisipkan token baru.
func (s *Service) issueOneTimeTokenForIdentity(
	ctx context.Context,
	userID string,
	identityID *string,
	purpose entity.TokenPurpose,
	ttl time.Duration,
	now time.Time,
) (*IssuedTokenResult, error) {
	if _, err := s.tokens.UpdateActiveStatusByUserAndPurpose(ctx, userID, purpose, entity.TokenStatusExpired); err != nil {
		return nil, err
	}
	id, err := s.idGen.NewUUID()
	if err != nil {
		return nil, iamerrors.Wrap(err, iamerrors.KindUnavailable, iamerrors.Code("iam.token.id_generation_failed"), "failed to generate token id")
	}
	plaintext, digest, err := s.tokenGen.GenerateOpaqueToken(oneTimeTokenBytes)
	if err != nil {
		return nil, err
	}
	var tok *entity.UserToken
	if identityID != nil && strings.TrimSpace(*identityID) != "" {
		tok, err = entity.NewUserTokenForIdentity(id, userID, *identityID, purpose, digest, now, now.Add(ttl))
	} else {
		tok, err = entity.NewUserToken(id, userID, purpose, digest, now, now.Add(ttl))
	}
	if err != nil {
		return nil, err
	}
	if err := s.tokens.Insert(ctx, tok); err != nil {
		return nil, err
	}
	return &IssuedTokenResult{
		Token:          tok,
		PlaintextToken: plaintext,
	}, nil
}

func parseLoginIdentifier(raw string) (entity.IdentityKind, string, error) {
	trimmed := strings.TrimSpace(raw)
	if strings.Contains(trimmed, "@") {
		val, err := entity.ValidateEmail(trimmed)
		return entity.IdentityKindEmail, val, err
	}
	val, err := entity.ValidateUsername(trimmed)
	return entity.IdentityKindUsername, val, err
}
