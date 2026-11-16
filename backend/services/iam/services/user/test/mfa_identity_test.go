package user_test

import (
	"context"
	"encoding/base32"
	"errors"
	"net/url"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
	"github.com/integrione/erp-retail/backend/services/iam/services/user"
)

func extractSecretFromURI(t *testing.T, otpAuthURI string) []byte {
	t.Helper()
	u, err := url.Parse(otpAuthURI)
	if err != nil {
		t.Fatalf("url.Parse(otpAuthURI) error = %v", err)
	}
	secretB32 := u.Query().Get("secret")
	raw, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(secretB32)
	if err != nil {
		t.Fatalf("decode base32 totp secret error = %v", err)
	}
	return raw
}

func TestUserService_TOTPEnrollConfirmChallengeReplayAndRecoveryCode(t *testing.T) {
	t.Parallel()

	h := newTestHarness(t)
	u, err := h.userSvc.CreateUser(context.Background(), user.CreateUserInput{
		Kind:                   entity.UserKindEmployee,
		Status:                 entity.UserStatusActive,
		DisplayName:            "Rina Security",
		Email:                  "rina@integrione.local",
		Username:               "rina.sec",
		Password:               "Bv3#AmberFort!2026",
		MarkIdentitiesVerified: true,
	})
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}

	// Enroll & Confirm TOTP.
	enroll, err := h.userSvc.EnrollTOTP(context.Background(), u.ID, "rina@integrione.local")
	if err != nil {
		t.Fatalf("EnrollTOTP() error = %v", err)
	}
	secret := extractSecretFromURI(t, enroll.OTPAuthURI)

	firstCode, err := h.totpProv.GenerateCodeAt(secret, h.clock.Now())
	if err != nil {
		t.Fatalf("GenerateCodeAt() error = %v", err)
	}
	confirmed, err := h.userSvc.ConfirmTOTP(context.Background(), u.ID, firstCode)
	if err != nil {
		t.Fatalf("ConfirmTOTP() error = %v", err)
	}
	if len(confirmed.RecoveryCodes) != 8 {
		t.Fatalf("expected 8 recovery codes, got %d", len(confirmed.RecoveryCodes))
	}

	// Login dengan sandi saat MFA aktif tidak langsung menerbitkan sesi penuh melainkan MFA challenge token.
	loginChallenge, err := h.userSvc.AuthenticatePassword(context.Background(), user.AuthenticatePasswordInput{
		Identifier: "rina@integrione.local",
		Password:   "Bv3#AmberFort!2026",
	})
	if err != nil {
		t.Fatalf("AuthenticatePassword(MFA) error = %v", err)
	}
	if loginChallenge.Status != user.AuthStatusMFARequired || loginChallenge.IssuedSession != nil || loginChallenge.MFAChallengeToken == "" {
		t.Fatalf("expected MFA challenge without full session, got %+v", loginChallenge)
	}

	// Replay kode pada step yang sama (digunakan saat ConfirmTOTP) wajib ditolak secara ambigu tanpa membocorkan clue replay.
	if _, err := h.userSvc.VerifyMFAChallenge(context.Background(), user.VerifyMFAChallengeInput{
		ChallengeToken: loginChallenge.MFAChallengeToken,
		Code:           firstCode,
	}); !errors.Is(err, iamerrors.ErrInvalidTOTPCode) || errors.Is(err, iamerrors.ErrTOTPReplayDetected) {
		t.Fatalf("expected ambiguous ErrInvalidTOTPCode on step replay without replay clue, got %v", err)
	}

	// Maju 35 detik -> step TOTP berikutnya berhasil menerbitkan sesi dengan MFAVerified = true.
	h.clock.Advance(35 * time.Second)
	nextCode, _ := h.totpProv.GenerateCodeAt(secret, h.clock.Now())
	mfaSuccess, err := h.userSvc.VerifyMFAChallenge(context.Background(), user.VerifyMFAChallengeInput{
		ChallengeToken: loginChallenge.MFAChallengeToken,
		Code:           nextCode,
		IPAddress:      "127.0.0.1",
	})
	if err != nil {
		t.Fatalf("VerifyMFAChallenge() error = %v", err)
	}
	if !mfaSuccess.IssuedSession.Session.MFAVerified {
		t.Fatal("expected session MFAVerified=true after TOTP verification")
	}

	// Uji login kedua diselesaikan menggunakan recovery code sekali pakai.
	challenge2, err := h.userSvc.AuthenticatePassword(context.Background(), user.AuthenticatePasswordInput{
		Identifier: "rina.sec",
		Password:   "Bv3#AmberFort!2026",
	})
	if err != nil {
		t.Fatalf("AuthenticatePassword(2) error = %v", err)
	}
	recRes, err := h.userSvc.VerifyRecoveryCodeChallenge(context.Background(), user.VerifyRecoveryCodeInput{
		ChallengeToken: challenge2.MFAChallengeToken,
		RecoveryCode:   confirmed.RecoveryCodes[0],
	})
	if err != nil || !recRes.IssuedSession.Session.MFAVerified {
		t.Fatalf("VerifyRecoveryCodeChallenge() = (%+v, %v)", recRes, err)
	}
	unusedCount, err := h.userSvc.CountUnusedRecoveryCodes(context.Background(), u.ID)
	if err != nil || unusedCount != 7 {
		t.Fatalf("CountUnusedRecoveryCodes() = (%d, %v), want 7", unusedCount, err)
	}

	// Recovery code yang sudah dipakai tidak bisa dipakai ulang.
	challenge3, _ := h.userSvc.AuthenticatePassword(context.Background(), user.AuthenticatePasswordInput{
		Identifier: "rina.sec",
		Password:   "Bv3#AmberFort!2026",
	})
	if _, err := h.userSvc.VerifyRecoveryCodeChallenge(context.Background(), user.VerifyRecoveryCodeInput{
		ChallengeToken: challenge3.MFAChallengeToken,
		RecoveryCode:   confirmed.RecoveryCodes[0],
	}); !errors.Is(err, iamerrors.ErrInvalidRecoveryCode) {
		t.Fatalf("expected ErrInvalidRecoveryCode on used recovery code, got %v", err)
	}
}

func TestUserService_IdentityLifecycleAndVerification(t *testing.T) {
	t.Parallel()

	h := newTestHarness(t)
	u, err := h.userSvc.CreateUser(context.Background(), user.CreateUserInput{
		Kind:        entity.UserKindCustomer,
		Status:      entity.UserStatusPending,
		DisplayName: "Pembeli Baru",
		Email:       "buyer@integrione.local",
		Password:    "Zx5#OceanHarbor!2026",
	})
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}

	idents, err := h.userSvc.ListIdentities(context.Background(), u.ID)
	if err != nil || len(idents) != 1 {
		t.Fatalf("ListIdentities() = (%+v, %v)", idents, err)
	}

	// Sebelum verifikasi & aktivasi, login ditolak secara ambigu ErrInvalidCredentials tanpa membocorkan status pending/unverified.
	if _, err := h.userSvc.AuthenticatePassword(context.Background(), user.AuthenticatePasswordInput{
		Identifier: "buyer@integrione.local",
		Password:   "Zx5#OceanHarbor!2026",
	}); !errors.Is(err, iamerrors.ErrInvalidCredentials) || errors.Is(err, iamerrors.ErrUserInactive) {
		t.Fatalf("expected ambiguous ErrInvalidCredentials for pending user, got %v", err)
	}

	verifTok, err := h.userSvc.RequestIdentityVerification(context.Background(), idents[0].ID)
	if err != nil {
		t.Fatalf("RequestIdentityVerification() error = %v", err)
	}
	verifiedIdent, err := h.userSvc.ConfirmIdentityVerification(context.Background(), idents[0].ID, verifTok.PlaintextToken, true)
	if err != nil || !verifiedIdent.IsVerified {
		t.Fatalf("ConfirmIdentityVerification() = (%+v, %v)", verifiedIdent, err)
	}

	// Sekarang akun aktif dan terverifikasi -> login berhasil.
	if _, err := h.userSvc.AuthenticatePassword(context.Background(), user.AuthenticatePasswordInput{
		Identifier: "buyer@integrione.local",
		Password:   "Zx5#OceanHarbor!2026",
	}); err != nil {
		t.Fatalf("expected login success after identity verification, got %v", err)
	}

	// Identitas terakhir tidak boleh dihapus (fail-closed conflict).
	if err := h.userSvc.RemoveIdentity(context.Background(), u.ID, idents[0].ID); err == nil {
		t.Fatal("expected error when deleting user's sole remaining identity")
	}
}

func TestUserService_EmailVerificationTokenIsBoundToIdentity(t *testing.T) {
	t.Parallel()

	h := newTestHarness(t)
	ctx := context.Background()
	u, err := h.userSvc.CreateUser(ctx, user.CreateUserInput{
		Kind:        entity.UserKindCustomer,
		Status:      entity.UserStatusPending,
		DisplayName: "Buyer Bound Token",
		Email:       "bound-token@integrione.local",
		Username:    "bound_token_buyer",
		Password:    "Zx5#OceanHarbor!2026",
	})
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}
	identities, err := h.userSvc.ListIdentities(ctx, u.ID)
	if err != nil || len(identities) != 2 {
		t.Fatalf("ListIdentities() = (%+v, %v)", identities, err)
	}
	var emailID, usernameID string
	for _, ident := range identities {
		if ident.Kind == entity.IdentityKindEmail {
			emailID = ident.ID
		} else if ident.Kind == entity.IdentityKindUsername {
			usernameID = ident.ID
		}
	}
	if _, err := h.userSvc.RequestIdentityVerification(ctx, usernameID); !errors.Is(err, iamerrors.ErrInvalidInput) {
		t.Fatalf("expected email-only verification, got %v", err)
	}
	legacyRaw := "legacy-email-token-without-identity"
	legacyID, err := h.idGen.NewUUID()
	if err != nil {
		t.Fatalf("NewUUID() error = %v", err)
	}
	legacyToken, err := entity.NewUserToken(legacyID, u.ID, entity.TokenPurposeEmailVerification, h.tokGen.Digest(legacyRaw), h.clock.Now(), h.clock.Now().Add(15*time.Minute))
	if err != nil {
		t.Fatalf("NewUserToken() error = %v", err)
	}
	h.store.mu.Lock()
	h.store.tokens[legacyID] = legacyToken
	h.store.mu.Unlock()
	legacyResult, err := h.userSvc.ConfirmEmailVerificationByToken(ctx, legacyRaw)
	if err != nil || legacyResult.Status != user.VerifyEmailStatusInvalid {
		t.Fatalf("expected unbound token to be invalid, got (%+v, %v)", legacyResult, err)
	}
	token, err := h.userSvc.RequestIdentityVerification(ctx, emailID)
	if err != nil {
		t.Fatalf("RequestIdentityVerification() error = %v", err)
	}
	if _, err := h.userSvc.ConfirmIdentityVerification(ctx, usernameID, token.PlaintextToken, true); !errors.Is(err, iamerrors.ErrTokenNotFound) {
		t.Fatalf("expected token bound to email identity, got %v", err)
	}
	verified, err := h.userSvc.ConfirmIdentityVerification(ctx, emailID, token.PlaintextToken, true)
	if err != nil || !verified.IsVerified {
		t.Fatalf("expected bound token to remain usable for email, got (%+v, %v)", verified, err)
	}
	if _, err := h.userSvc.RequestIdentityVerification(ctx, emailID); !errors.Is(err, iamerrors.ErrEmailAlreadyVerified) {
		t.Fatalf("expected verified email to reject a new token, got %v", err)
	}
}

func TestUserService_IdentityCardinalityAndUpdateUsername(t *testing.T) {
	t.Parallel()

	h := newTestHarness(t)
	ctx := context.Background()

	// 1. User1 dibuat dengan 1 email + 1 username ("sancyty").
	u1, err := h.userSvc.CreateUser(ctx, user.CreateUserInput{
		Kind:                   entity.UserKindEmployee,
		Status:                 entity.UserStatusActive,
		DisplayName:            "User Satu",
		Email:                  "user1@integrione.local",
		Username:               "sancyty",
		Password:               "K9#StrongPass!2026",
		MarkIdentitiesVerified: true,
	})
	if err != nil {
		t.Fatalf("CreateUser(u1) error = %v", err)
	}

	// 2. Duplikat email atau username saat CreateUser wajib ditolak eksplisit.
	if _, err := h.userSvc.CreateUser(ctx, user.CreateUserInput{
		Kind:        entity.UserKindCustomer,
		Status:      entity.UserStatusActive,
		DisplayName: "Dup Email",
		Email:       "USER1@integrione.local",
		Password:    "L8#UniquePass!2026",
	}); !errors.Is(err, iamerrors.ErrIdentityAlreadyExists) {
		t.Fatalf("expected ErrIdentityAlreadyExists for duplicate email, got %v", err)
	}
	if _, err := h.userSvc.CreateUser(ctx, user.CreateUserInput{
		Kind:        entity.UserKindCustomer,
		Status:      entity.UserStatusActive,
		DisplayName: "Dup Username",
		Email:       "user-other@integrione.local",
		Username:    "SANCYTY",
		Password:    "L8#UniquePass!2026",
	}); !errors.Is(err, iamerrors.ErrIdentityAlreadyExists) {
		t.Fatalf("expected ErrIdentityAlreadyExists for duplicate username, got %v", err)
	}

	// 3. Menambah email ke-2 atau username ke-2 via AddIdentity wajib ditolak ErrIdentityLimitExceeded.
	if _, err := h.userSvc.AddIdentity(ctx, user.AddIdentityInput{
		UserID: u1.ID,
		Kind:   entity.IdentityKindEmail,
		Value:  "user1.second@integrione.local",
	}); !errors.Is(err, iamerrors.ErrIdentityLimitExceeded) {
		t.Fatalf("expected ErrIdentityLimitExceeded for 2nd email, got %v", err)
	}
	if _, err := h.userSvc.AddIdentity(ctx, user.AddIdentityInput{
		UserID: u1.ID,
		Kind:   entity.IdentityKindUsername,
		Value:  "user1_second",
	}); !errors.Is(err, iamerrors.ErrIdentityLimitExceeded) {
		t.Fatalf("expected ErrIdentityLimitExceeded for 2nd username, got %v", err)
	}

	// 4. Mengganti username dengan nilai sama atau format tidak valid wajib ditolak.
	if _, err := h.userSvc.UpdateUsername(ctx, user.UpdateUsernameInput{UserID: u1.ID, Username: " SancyTy "}); !errors.Is(err, iamerrors.ErrSameUsername) {
		t.Fatalf("expected ErrSameUsername, got %v", err)
	}
	if _, err := h.userSvc.UpdateUsername(ctx, user.UpdateUsernameInput{UserID: u1.ID, Username: "invalid space"}); !errors.Is(err, iamerrors.ErrInvalidUsername) {
		t.Fatalf("expected ErrInvalidUsername, got %v", err)
	}

	// 5. Ganti username User1: sancyty -> kibo langsung terverifikasi tanpa email verification dan langsung aktif untuk login.
	updatedU1, err := h.userSvc.UpdateUsername(ctx, user.UpdateUsernameInput{UserID: u1.ID, Username: "KiBo"})
	if err != nil {
		t.Fatalf("UpdateUsername(sancyty -> kibo) error = %v", err)
	}
	if updatedU1.Value != "kibo" || !updatedU1.IsVerified || updatedU1.VerifiedAt == nil {
		t.Fatalf("expected verified username 'kibo', got %+v", updatedU1)
	}
	if _, err := h.userSvc.AuthenticatePassword(ctx, user.AuthenticatePasswordInput{
		Identifier: "kibo",
		Password:   "K9#StrongPass!2026",
	}); err != nil {
		t.Fatalf("expected login with new username 'kibo' to succeed, got %v", err)
	}
	if _, err := h.userSvc.AuthenticatePassword(ctx, user.AuthenticatePasswordInput{
		Identifier: "sancyty",
		Password:   "K9#StrongPass!2026",
	}); !errors.Is(err, iamerrors.ErrInvalidCredentials) {
		t.Fatalf("expected old username 'sancyty' to be rejected for u1, got %v", err)
	}

	// 6. Username lama "sancyty" sekarang valid dan bebas dipakai user baru, sementara "kibo" terkunci untuk User1.
	u2, err := h.userSvc.CreateUser(ctx, user.CreateUserInput{
		Kind:                   entity.UserKindCustomer,
		Status:                 entity.UserStatusActive,
		DisplayName:            "User Dua",
		Email:                  "user2@integrione.local",
		Username:               "sancyty",
		Password:               "M4#SilverMoon!2026",
		MarkIdentitiesVerified: true,
	})
	if err != nil {
		t.Fatalf("CreateUser(u2 with freed 'sancyty') error = %v", err)
	}
	if _, err := h.userSvc.AuthenticatePassword(ctx, user.AuthenticatePasswordInput{
		Identifier: "sancyty",
		Password:   "M4#SilverMoon!2026",
	}); err != nil {
		t.Fatalf("expected u2 login with reused 'sancyty' to succeed, got %v", err)
	}
	if _, err := h.userSvc.UpdateUsername(ctx, user.UpdateUsernameInput{UserID: u2.ID, Username: "kibo"}); !errors.Is(err, iamerrors.ErrIdentityAlreadyExists) {
		t.Fatalf("expected ErrIdentityAlreadyExists when u2 claims active 'kibo', got %v", err)
	}

	// 7. Customer tanpa username awal dapat menambahkan username pertamanya via UpdateUsername.
	u3, err := h.userSvc.CreateUser(ctx, user.CreateUserInput{
		Kind:        entity.UserKindCustomer,
		Status:      entity.UserStatusPending,
		Email:       "buyer.nouname@integrione.local",
		DisplayName: "Pembeli Tanpa Username",
		Password:    "N5#HarborKey!2026",
	})
	if err != nil {
		t.Fatalf("CreateUser(no username) error = %v", err)
	}
	if _, err := h.userSvc.UpdateUsername(ctx, user.UpdateUsernameInput{UserID: u3.ID, Username: "first.uname"}); !errors.Is(err, iamerrors.ErrUserInactive) {
		t.Fatalf("expected ErrUserInactive for pending customer UpdateUsername, got %v", err)
	}
	u3Resend, err := h.userSvc.ResendValidationEmailToken(ctx, user.ResendValidationEmailInput{Email: "buyer.nouname@integrione.local"})
	if err != nil || u3Resend.VerificationToken == "" {
		t.Fatalf("ResendValidationEmailToken(u3) error = (%+v, %v)", u3Resend, err)
	}
	if _, err := h.userSvc.ConfirmEmailVerificationByToken(ctx, u3Resend.VerificationToken); err != nil {
		t.Fatalf("ConfirmEmailVerificationByToken() error = %v", err)
	}
	firstUname, err := h.userSvc.UpdateUsername(ctx, user.UpdateUsernameInput{UserID: u3.ID, Username: "first.uname"})
	if err != nil || firstUname.Value != "first.uname" || !firstUname.IsVerified {
		t.Fatalf("expected first username created and verified, got (%+v, %v)", firstUname, err)
	}
	if _, err := h.userSvc.AuthenticatePassword(ctx, user.AuthenticatePasswordInput{
		Identifier: "first.uname",
		Password:   "N5#HarborKey!2026",
	}); err != nil {
		t.Fatalf("expected login with newly added username to succeed, got %v", err)
	}

	// 8. Akun suspended/disabled ditolak saat UpdateUsername.
	if _, err := h.userSvc.Suspend(ctx, u2.ID); err != nil {
		t.Fatalf("Suspend(u2) error = %v", err)
	}
	if _, err := h.userSvc.UpdateUsername(ctx, user.UpdateUsernameInput{UserID: u2.ID, Username: "u2.new"}); !errors.Is(err, iamerrors.ErrUserSuspended) {
		t.Fatalf("expected ErrUserSuspended on UpdateUsername, got %v", err)
	}
}

func TestUserService_ConfirmEmailVerificationByTokenOnDeletedUser(t *testing.T) {
	t.Parallel()

	h := newTestHarness(t)
	ctx := context.Background()

	u, err := h.userSvc.CreateUser(ctx, user.CreateUserInput{
		Kind:        entity.UserKindCustomer,
		Status:      entity.UserStatusPending,
		Email:       "deleted.verify@integrione.local",
		DisplayName: "Deleted Verify Buyer",
		Password:    "P7#HarborGate!2026",
	})
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}
	resendTok, err := h.userSvc.ResendValidationEmailToken(ctx, user.ResendValidationEmailInput{Email: "deleted.verify@integrione.local"})
	if err != nil || resendTok.VerificationToken == "" {
		t.Fatalf("ResendValidationEmailToken() error = (%+v, %v)", resendTok, err)
	}
	if err := h.userSvc.SoftDelete(ctx, u.ID); err != nil {
		t.Fatalf("SoftDelete() error = %v", err)
	}

	res, err := h.userSvc.ConfirmEmailVerificationByToken(ctx, resendTok.VerificationToken)
	if err != nil {
		t.Fatalf("ConfirmEmailVerificationByToken() unexpected error = %v", err)
	}
	if res.Status != user.VerifyEmailStatusInvalid {
		t.Fatalf("expected VerifyEmailStatusInvalid for deleted user, got %q", res.Status)
	}
}
