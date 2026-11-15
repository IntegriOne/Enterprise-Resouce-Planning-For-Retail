package entity_test

import (
	"bytes"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
)

const (
	testTOTPFactorID = "33333333-3333-4333-8333-333333333333"
	testRecoveryID   = "44444444-4444-4444-8444-444444444444"
	testTokenID      = "55555555-5555-4555-8555-555555555555"
)

func TestUserCredentialLockoutAndRedaction(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 11, 14, 10, 0, 0, 0, time.Local)
	secretHash := "$argon2id$v=19$m=65536,t=3,p=2$supersecret"

	cred, err := entity.NewUserCredential(testUserID, secretHash, true, now)
	if err != nil {
		t.Fatalf("NewUserCredential() error = %v", err)
	}
	if cred.PasswordHash() != secretHash {
		t.Fatal("PasswordHash() mismatch")
	}

	if strings.Contains(cred.String(), secretHash) {
		t.Fatal("Credential.String() must redact PasswordHash")
	}
	var logBuf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logBuf, nil))
	logger.Info("cred", "credential", cred)
	if strings.Contains(logBuf.String(), secretHash) {
		t.Fatal("slog output must redact PasswordHash")
	}

	// 4 kali gagal belum terkunci; gagal ke-5 mengunci selama 15 menit.
	for i := range 4 {
		cred.RecordFailedAttempt(5, 15*time.Minute, now.Add(time.Duration(i)*time.Second))
		if cred.IsLocked(now.Add(5 * time.Second)) {
			t.Fatalf("should not lock on attempt %d", i+1)
		}
	}
	cred.RecordFailedAttempt(5, 15*time.Minute, now.Add(10*time.Second))
	if !cred.IsLocked(now.Add(11 * time.Second)) {
		t.Fatal("must be locked after 5th failed attempt")
	}
	cred.ResetFailedAttempts(now.Add(12 * time.Second))
	if cred.FailedLoginAttempts != 0 || cred.IsLocked(now.Add(12*time.Second)) {
		t.Fatal("ResetFailedAttempts() must clear count and lock")
	}
	cred.RecordFailedAttempt(1, 15*time.Minute, now.Add(13*time.Second))
	if cred.IsLocked(now.Add(29 * time.Minute)) {
		t.Fatal("lockout must expire after lock duration")
	}

	if err := cred.UpdatePasswordHash("$argon2id$v=19$m=65536,t=3,p=2$newhash", now.Add(20*time.Minute)); err != nil {
		t.Fatalf("UpdatePasswordHash() error = %v", err)
	}
	if cred.FailedLoginAttempts != 0 || cred.LockedUntil != nil || cred.MustChangePassword {
		t.Fatal("UpdatePasswordHash() must reset attempts, lockout, and MustChangePassword")
	}
}

func TestUserTOTPFactorDefensiveCopyReplayAndRedaction(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 11, 14, 10, 0, 0, 0, time.Local)
	rawEncrypted := []byte("top-secret-encrypted-totp-seed")
	originalCopy := bytes.Clone(rawEncrypted)

	factor, err := entity.NewUserTOTPFactor(testTOTPFactorID, testUserID, rawEncrypted, 1, now)
	if err != nil {
		t.Fatalf("NewUserTOTPFactor() error = %v", err)
	}

	// Mutasi slice caller dan slice getter tidak boleh memengaruhi entitas.
	rawEncrypted[0] = 'X'
	gotSecret := factor.EncryptedSecret()
	if !bytes.Equal(gotSecret, originalCopy) {
		t.Fatal("EncryptedSecret() was mutated through input slice aliasing")
	}
	gotSecret[0] = 'Y'
	if !bytes.Equal(factor.EncryptedSecret(), originalCopy) {
		t.Fatal("EncryptedSecret() getter must return defensive clone")
	}

	if strings.Contains(factor.String(), string(originalCopy)) {
		t.Fatal("UserTOTPFactor.String() must redact encrypted secret")
	}
	if err := factor.CanUseStep(100); !errors.Is(err, iamerrors.ErrTOTPNotConfirmed) {
		t.Fatalf("unconfirmed TOTP CanUseStep() = %v, want ErrTOTPNotConfirmed", err)
	}

	if err := factor.Confirm(now.Add(time.Minute)); err != nil {
		t.Fatalf("Confirm() error = %v", err)
	}
	if err := factor.Confirm(now.Add(2 * time.Minute)); !errors.Is(err, iamerrors.ErrTOTPAlreadyConfirmed) {
		t.Fatalf("second Confirm() = %v, want ErrTOTPAlreadyConfirmed", err)
	}

	if err := factor.MarkStepUsed(500, now.Add(3*time.Minute)); err != nil {
		t.Fatalf("MarkStepUsed(500) error = %v", err)
	}
	if err := factor.CanUseStep(500); !errors.Is(err, iamerrors.ErrTOTPReplayDetected) {
		t.Fatalf("CanUseStep(500) = %v, want ErrTOTPReplayDetected", err)
	}
	if err := factor.CanUseStep(499); !errors.Is(err, iamerrors.ErrTOTPReplayDetected) {
		t.Fatalf("CanUseStep(499) = %v, want ErrTOTPReplayDetected", err)
	}
	if err := factor.CanUseStep(501); err != nil {
		t.Fatalf("CanUseStep(501) error = %v", err)
	}

	var totpBuf bytes.Buffer
	slog.New(slog.NewTextHandler(&totpBuf, nil)).Info("totp", "factor", factor)
	if strings.Contains(totpBuf.String(), string(originalCopy)) {
		t.Fatal("UserTOTPFactor.LogValue() must redact secret")
	}
	factor.Disable(now.Add(4 * time.Minute))
	if err := factor.CanUseStep(502); !errors.Is(err, iamerrors.ErrTOTPNotConfirmed) {
		t.Fatalf("disabled TOTP CanUseStep() = %v, want ErrTOTPNotConfirmed", err)
	}
}

func TestUserRecoveryCodeAndUserToken(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 11, 14, 10, 0, 0, 0, time.Local)

	rc, err := entity.NewUserRecoveryCode(testRecoveryID, testUserID, "$argon2id$recovery-hash", now)
	if err != nil {
		t.Fatalf("NewUserRecoveryCode() error = %v", err)
	}
	if rc.CodeHash() != "$argon2id$recovery-hash" || rc.IsUsed() {
		t.Fatal("new recovery code mismatch")
	}
	if strings.Contains(rc.String(), "$argon2id$recovery-hash") {
		t.Fatal("UserRecoveryCode.String() must redact CodeHash")
	}
	var rcBuf bytes.Buffer
	slog.New(slog.NewTextHandler(&rcBuf, nil)).Info("rc", "recovery", rc)
	if strings.Contains(rcBuf.String(), "$argon2id$recovery-hash") {
		t.Fatal("UserRecoveryCode.LogValue() must redact CodeHash")
	}
	if err := rc.Consume(now.Add(time.Minute)); err != nil {
		t.Fatalf("Consume() error = %v", err)
	}
	if !rc.IsUsed() {
		t.Fatal("IsUsed() must be true after Consume")
	}
	rehydratedRC, err := entity.RehydrateUserRecoveryCode(testRecoveryID, testUserID, "$argon2id$recovery-hash", rc.UsedAt, now)
	if err != nil || !rehydratedRC.IsUsed() {
		t.Fatalf("RehydrateUserRecoveryCode() mismatch: rc=%v err=%v", rehydratedRC, err)
	}
	if err := rc.Consume(now.Add(2 * time.Minute)); !errors.Is(err, iamerrors.ErrRecoveryCodeAlreadyUsed) {
		t.Fatalf("second Consume() = %v, want ErrRecoveryCodeAlreadyUsed", err)
	}

	digest := []byte("0123456789abcdef0123456789abcdef")
	if _, err := entity.NewUserToken(testTokenID, testUserID, entity.TokenPurposePasswordReset, digest, now, now); !errors.Is(err, iamerrors.ErrInvalidInput) {
		t.Fatalf("token with expires_at == created_at should fail ErrInvalidInput, got %v", err)
	}

	tok, err := entity.NewUserToken(testTokenID, testUserID, entity.TokenPurposePasswordReset, digest, now, now.Add(15*time.Minute))
	if err != nil {
		t.Fatalf("NewUserToken() error = %v", err)
	}
	digest[0] = 'Z'
	if bytes.Equal(tok.TokenDigest(), digest) {
		t.Fatal("UserToken must defensively clone TokenDigest")
	}
	if strings.Contains(tok.String(), "123456789abcdef") {
		t.Fatal("UserToken.String() must redact TokenDigest")
	}
	var tokBuf bytes.Buffer
	slog.New(slog.NewTextHandler(&tokBuf, nil)).Info("tok", "token", tok)
	if strings.Contains(tokBuf.String(), "123456789abcdef") {
		t.Fatal("UserToken.LogValue() must redact TokenDigest")
	}
	if err := tok.ValidateForUse(entity.TokenPurposeEmailVerification, now.Add(time.Minute)); !errors.Is(err, iamerrors.ErrTokenInvalidPurpose) {
		t.Fatalf("wrong purpose error = %v, want ErrTokenInvalidPurpose", err)
	}
	if err := tok.ValidateForUse(entity.TokenPurposePasswordReset, now.Add(16*time.Minute)); !errors.Is(err, iamerrors.ErrTokenExpired) {
		t.Fatalf("expired token error = %v, want ErrTokenExpired", err)
	}
	if err := tok.Consume(entity.TokenPurposePasswordReset, now.Add(5*time.Minute)); err != nil {
		t.Fatalf("Consume() error = %v", err)
	}
	rehydratedTok, err := entity.RehydrateUserToken(testTokenID, testUserID, nil, entity.TokenPurposePasswordReset, entity.TokenStatusConsumed, tok.TokenDigest(), now.Add(15*time.Minute), tok.UsedAt, now)
	if err != nil {
		t.Fatalf("RehydrateUserToken() error = %v", err)
	}
	if err := rehydratedTok.Consume(entity.TokenPurposePasswordReset, now.Add(6*time.Minute)); !errors.Is(err, iamerrors.ErrTokenAlreadyUsed) {
		t.Fatalf("reused token error = %v, want ErrTokenAlreadyUsed", err)
	}

	expiredTok, err := entity.NewUserTokenForIdentity(testTokenID, testUserID, testIdentityID, entity.TokenPurposeEmailVerification, digest, now, now.Add(15*time.Minute))
	if err != nil {
		t.Fatalf("NewUserTokenForIdentity() error = %v", err)
	}
	if expiredTok.IdentityID == nil || *expiredTok.IdentityID != testIdentityID {
		t.Fatalf("expected IdentityID = %s, got %v", testIdentityID, expiredTok.IdentityID)
	}
	expiredTok.Expire()
	if err := expiredTok.ValidateForUse(entity.TokenPurposeEmailVerification, now.Add(time.Minute)); !errors.Is(err, iamerrors.ErrTokenExpired) {
		t.Fatalf("status=expired error = %v, want ErrTokenExpired", err)
	}

	revokedTok, err := entity.NewUserToken(testTokenID, testUserID, entity.TokenPurposeEmailVerification, digest, now, now.Add(15*time.Minute))
	if err != nil {
		t.Fatalf("NewUserToken() error = %v", err)
	}
	revokedTok.Revoke()
	if err := revokedTok.ValidateForUse(entity.TokenPurposeEmailVerification, now.Add(time.Minute)); !errors.Is(err, iamerrors.ErrTokenRevoked) {
		t.Fatalf("status=revoked error = %v, want ErrTokenRevoked", err)
	}
}
