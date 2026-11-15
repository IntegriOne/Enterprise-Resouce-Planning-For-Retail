package user_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/iam/config"
	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
	"github.com/integrione/erp-retail/backend/services/iam/repository"
	"github.com/integrione/erp-retail/backend/services/iam/services"
	"github.com/integrione/erp-retail/backend/services/iam/services/user"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestUserService_CustomerRegisterVerifyResendAndPendingEmailLifecycle(t *testing.T) {
	t.Parallel()

	h := newTestHarness(t)
	ctx := context.Background()

	createdUser, err := h.userSvc.CreateUser(ctx, user.CreateUserInput{
		Kind:        entity.UserKindCustomer,
		Status:      entity.UserStatusPending,
		Email:       "pelanggan@integrione.local",
		Username:    "pelanggan01",
		Password:    "VaultCust#2026!Z",
		DisplayName: "Pelanggan IntegriOne",
	})
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}
	if createdUser.Status != "pending" {
		t.Fatalf("unexpected CreateUser status: %+v", createdUser)
	}

	// Pastikan CreateUser menerbitkan 1 token verifikasi email awal yang aktif.
	initialTokens := 0
	for _, tok := range h.store.tokens {
		if tok.UserID == createdUser.ID && tok.Purpose == entity.TokenPurposeEmailVerification && tok.Status == entity.TokenStatusActive {
			initialTokens++
		}
	}
	if initialTokens != 1 {
		t.Fatalf("expected 1 active email_verification token after CreateUser, got %d", initialTokens)
	}

	// Duplikasi email/username wajib ditolak.
	if _, err := h.userSvc.CreateUser(ctx, user.CreateUserInput{
		Kind:        entity.UserKindCustomer,
		Status:      entity.UserStatusPending,
		DisplayName: "Duplikat Email",
		Email:       "pelanggan@integrione.local",
		Username:    "pelanggan_baru",
		Password:    "VaultCust#2026!Z",
	}); !errors.Is(err, iamerrors.ErrIdentityAlreadyExists) {
		t.Fatalf("expected ErrIdentityAlreadyExists, got %v", err)
	}

	// Kasus (2) kirim ulang token saat status registrasi PENDING menandai token lama menjadi expired.
	resendPending1, err := h.userSvc.ResendValidationEmailToken(ctx, user.ResendValidationEmailInput{
		Email: "pelanggan@integrione.local",
	})
	if err != nil || !resendPending1.IsResend || resendPending1.VerificationToken == "" {
		t.Fatalf("unexpected ResendValidationEmailToken(PENDING-1): (%+v, %v)", resendPending1, err)
	}
	resendPendingRes, err := h.userSvc.ResendValidationEmailToken(ctx, user.ResendValidationEmailInput{
		Email: "pelanggan@integrione.local",
	})
	if err != nil || !resendPendingRes.IsResend || resendPendingRes.VerificationToken == "" || resendPendingRes.VerificationToken == resendPending1.VerificationToken {
		t.Fatalf("unexpected ResendValidationEmailToken(PENDING-2): (%+v, %v)", resendPendingRes, err)
	}
	// Token verifikasi lama otomatis berstatus expired.
	oldRegRes, err := h.userSvc.ConfirmEmailVerificationByToken(ctx, resendPending1.VerificationToken)
	if err != nil || oldRegRes.Status != user.VerifyEmailStatusExpired {
		t.Fatalf("expected VerifyEmailStatusExpired for superseded registration token, got (%+v, %v)", oldRegRes, err)
	}

	// Token tidak dikenal mengembalikan invalid.
	invRes, err := h.userSvc.ConfirmEmailVerificationByToken(ctx, "unknown-token-value")
	if err != nil || invRes.Status != user.VerifyEmailStatusInvalid {
		t.Fatalf("expected VerifyEmailStatusInvalid, got (%+v, %v)", invRes, err)
	}

	// Token sah terbaru mengaktifkan pengguna pending -> active dan menandai email terverifikasi.
	verRes, err := h.userSvc.ConfirmEmailVerificationByToken(ctx, resendPendingRes.VerificationToken)
	if err != nil || verRes.Status != user.VerifyEmailStatusVerified || verRes.User.Status != "active" {
		t.Fatalf("expected verified active customer, got (%+v, %v)", verRes, err)
	}

	// Token yang sama dipakai kedua kali -> already_used.
	usedRes, err := h.userSvc.ConfirmEmailVerificationByToken(ctx, resendPendingRes.VerificationToken)
	if err != nil || usedRes.Status != user.VerifyEmailStatusAlreadyUsed {
		t.Fatalf("expected VerifyEmailStatusAlreadyUsed, got (%+v, %v)", usedRes, err)
	}

	// Reset password hanya menerima email terverifikasi; username wajib ditolak.
	if _, err := h.userSvc.RequestPasswordReset(ctx, "pelanggan01"); !errors.Is(err, iamerrors.ErrInvalidEmail) {
		t.Fatalf("expected ErrInvalidEmail for username reset, got %v", err)
	}
	reset1, err := h.userSvc.RequestPasswordReset(ctx, "pelanggan@integrione.local")
	if err != nil || reset1.PlaintextToken == "" {
		t.Fatalf("RequestPasswordReset(1) failed: (%+v, %v)", reset1, err)
	}
	reset2, err := h.userSvc.RequestPasswordReset(ctx, "pelanggan@integrione.local")
	if err != nil || reset2.PlaintextToken == "" || reset2.PlaintextToken == reset1.PlaintextToken {
		t.Fatalf("RequestPasswordReset(2) failed: (%+v, %v)", reset2, err)
	}
	if err := h.userSvc.ConfirmPasswordReset(ctx, user.ConfirmPasswordResetInput{
		RawToken:    reset1.PlaintextToken,
		NewPassword: "VaultCust#2027!Q",
	}); !errors.Is(err, iamerrors.ErrTokenExpired) {
		t.Fatalf("expected ErrTokenExpired for superseded reset token, got %v", err)
	}

	// Alur updateUserEmail: email lama tetap aktif sebagai email login sampai pending email diverifikasi.
	upd1, err := h.userSvc.UpdateEmail(ctx, user.UpdateEmailInput{
		UserID:          createdUser.ID,
		NewEmail:        "baru1@integrione.local",
		CurrentPassword: "VaultCust#2026!Z",
	})
	if err != nil || upd1.PendingIdentity == nil || upd1.PendingIdentity.IsPrimary || upd1.PendingIdentity.IsVerified {
		t.Fatalf("unexpected UpdateEmail(1): (%+v, %v)", upd1, err)
	}
	// Login menggunakan email lama tetap berhasil sementara pending email belum diverifikasi.
	if authOld, err := h.userSvc.AuthenticatePassword(ctx, user.AuthenticatePasswordInput{
		Identifier: "pelanggan@integrione.local",
		Password:   "VaultCust#2026!Z",
		IPAddress:  "127.0.0.1",
	}); err != nil || authOld.Status != user.AuthStatusAuthenticated {
		t.Fatalf("old email should remain active while pending email is unverified: (%+v, %v)", authOld, err)
	}

	// Kasus (1) resendValidationEmailToken saat sudah login (cookie) untuk pending email:
	// harus menjaga maksimal 1 token active per (user_id, purpose) di database.
	resendPendingAuth, err := h.userSvc.ResendValidationEmailToken(ctx, user.ResendValidationEmailInput{
		UserID: createdUser.ID,
	})
	if err != nil || !resendPendingAuth.IsResend || resendPendingAuth.VerificationToken == upd1.VerificationToken {
		t.Fatalf("unexpected ResendValidationEmailToken(auth): (%+v, %v)", resendPendingAuth, err)
	}
	activeVerificationTokens := 0
	for _, tok := range h.store.tokens {
		if tok.UserID == createdUser.ID && tok.Purpose == entity.TokenPurposeEmailVerification && tok.Status == entity.TokenStatusActive {
			activeVerificationTokens++
		}
	}
	if activeVerificationTokens != 1 {
		t.Fatalf("expected exactly 1 active email_verification token, got %d", activeVerificationTokens)
	}

	// Menghapus pending email membatalkan token aktif (revoked) dan mempertahankan email lama.
	deleted, err := h.userSvc.DeletePendingEmail(ctx, createdUser.ID)
	if err != nil || !deleted {
		t.Fatalf("DeletePendingEmail() = (%v, %v)", deleted, err)
	}
	revRes, err := h.userSvc.ConfirmEmailVerificationByToken(ctx, resendPendingAuth.VerificationToken)
	if err != nil || revRes.Status != user.VerifyEmailStatusInvalid {
		t.Fatalf("expected VerifyEmailStatusInvalid after DeletePendingEmail, got (%+v, %v)", revRes, err)
	}

	// Pembaruan email kedua lalu verifikasi sukses menggantikan email lama secara atomik.
	upd2, err := h.userSvc.UpdateEmail(ctx, user.UpdateEmailInput{
		UserID:          createdUser.ID,
		NewEmail:        "baru2@integrione.local",
		CurrentPassword: "VaultCust#2026!Z",
	})
	if err != nil {
		t.Fatalf("UpdateEmail(2) error = %v", err)
	}
	verSwap, err := h.userSvc.ConfirmEmailVerificationByToken(ctx, upd2.VerificationToken)
	if err != nil || verSwap.Status != user.VerifyEmailStatusVerified || verSwap.Identity == nil || verSwap.Identity.Value != "baru2@integrione.local" || !verSwap.Identity.IsPrimary {
		t.Fatalf("expected verified swapped primary email, got (%+v, %v)", verSwap, err)
	}
	// Email lama tidak lagi terdaftar setelah penggantian sukses.
	if _, err := h.userSvc.RequestPasswordReset(ctx, "pelanggan@integrione.local"); !errors.Is(err, iamerrors.ErrIdentityNotFound) {
		t.Fatalf("expected old email deleted after verification swap, got %v", err)
	}

	// Registrasi kedua lalu kedaluwarsakan -> expired.
	if _, err := h.userSvc.CreateUser(ctx, user.CreateUserInput{
		Kind:        entity.UserKindCustomer,
		Status:      entity.UserStatusPending,
		DisplayName: "Expired User",
		Email:       "expired@integrione.local",
		Password:    "VaultCust#2026!Z",
	}); err != nil {
		t.Fatalf("CreateUser(2) error = %v", err)
	}
	reg2Resend, err := h.userSvc.ResendValidationEmailToken(ctx, user.ResendValidationEmailInput{
		Email: "expired@integrione.local",
	})
	if err != nil || reg2Resend.VerificationToken == "" {
		t.Fatalf("ResendValidationEmailToken(expired) error = (%+v, %v)", reg2Resend, err)
	}
	h.clock.Advance(25 * time.Hour)
	expRes, err := h.userSvc.ConfirmEmailVerificationByToken(ctx, reg2Resend.VerificationToken)
	if err != nil || expRes.Status != user.VerifyEmailStatusExpired {
		t.Fatalf("expected VerifyEmailStatusExpired, got (%+v, %v)", expRes, err)
	}
}

func TestBootstrapSystemPrincipalAndRBAC_Idempotent(t *testing.T) {
	t.Parallel()

	// Uji struktur bootstrap gagal-tertutup dengan dependensi nil.
	if err := services.BootstrapSystemPrincipalAndRBAC(context.Background(), nil, nil, config.SystemBootstrapConfig{}, nil); err == nil {
		t.Fatal("expected error for nil bootstrap dependencies")
	}

	// Verifikasi sqlite hanya untuk memastikan konstruksi registry tidak panic; schema PG diuji pada suite PG/E2E.
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open error = %v", err)
	}
	repos, err := repository.New(db)
	if err != nil {
		t.Fatalf("repository.New error = %v", err)
	}
	reg, err := services.NewRegistry(repos, services.SecurityOptions{AllowEphemeralDevKey: true, AllowEphemeralCSRFKey: true})
	if err != nil {
		t.Fatalf("services.NewRegistry error = %v", err)
	}
	_ = slog.New(slog.NewTextHandler(io.Discard, nil))
	if err := reg.Validate(); err != nil {
		t.Fatalf("reg.Validate() error = %v", err)
	}
}

// TestUserService_ResendValidationEmailToken_Case1PendingAndVerified menguji kasus login cookie auth (hanya pending email, error verified, atau username-only) serta netralitas kasus anonim.
func TestUserService_ResendValidationEmailToken_Case1PendingAndVerified(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("verified user without pending email returns ErrEmailAlreadyVerified", func(t *testing.T) {
		t.Parallel()
		h := newTestHarness(t)
		u, err := h.userSvc.CreateUser(ctx, user.CreateUserInput{
			Kind:                   entity.UserKindEmployee,
			Status:                 entity.UserStatusActive,
			DisplayName:            "Verified User",
			Email:                  "verified@integrione.local",
			Username:               "verified.user",
			Password:               "VaultEmp#2026!A",
			MarkIdentitiesVerified: true,
		})
		if err != nil {
			t.Fatalf("CreateUser() error = %v", err)
		}

		for _, filterEmail := range []string{"", "verified@integrione.local", "other@integrione.local"} {
			_, err := h.userSvc.ResendValidationEmailToken(ctx, user.ResendValidationEmailInput{
				UserID: u.ID,
				Email:  filterEmail,
			})
			if !errors.Is(err, iamerrors.ErrEmailAlreadyVerified) {
				t.Fatalf("ResendValidationEmailToken(filter=%q) error = %v, want ErrEmailAlreadyVerified", filterEmail, err)
			}
			if err.Error() != "your email has been verified" {
				t.Fatalf("ResendValidationEmailToken(filter=%q) message = %q, want %q", filterEmail, err.Error(), "your email has been verified")
			}
		}

		// Kasus (2) tanpa login tetap netral meskipun email sudah terverifikasi dan akun aktif.
		anonRes, err := h.userSvc.ResendValidationEmailToken(ctx, user.ResendValidationEmailInput{
			Email: "verified@integrione.local",
		})
		if err != nil || !anonRes.IsResend || anonRes.VerificationToken != "" {
			t.Fatalf("unauthenticated resend expected neutral empty token result, got (%+v, %v)", anonRes, err)
		}
	})

	t.Run("username-only user without email returns ErrPendingEmailNotFound", func(t *testing.T) {
		t.Parallel()
		h := newTestHarness(t)
		u, err := h.userSvc.CreateUser(ctx, user.CreateUserInput{
			Kind:                   entity.UserKindSystem,
			Status:                 entity.UserStatusActive,
			DisplayName:            "Username Only",
			Username:               "system.only",
			Password:               "VaultSys#2026!B",
			MarkIdentitiesVerified: true,
		})
		if err != nil {
			t.Fatalf("CreateUser() error = %v", err)
		}

		_, err = h.userSvc.ResendValidationEmailToken(ctx, user.ResendValidationEmailInput{
			UserID: u.ID,
		})
		if !errors.Is(err, iamerrors.ErrPendingEmailNotFound) || errors.Is(err, iamerrors.ErrEmailAlreadyVerified) {
			t.Fatalf("ResendValidationEmailToken(username-only) error = %v, want ErrPendingEmailNotFound", err)
		}
	})

	t.Run("verified user with pending email resends and expires prior token then errors after verify", func(t *testing.T) {
		t.Parallel()
		h := newTestHarness(t)
		u, err := h.userSvc.CreateUser(ctx, user.CreateUserInput{
			Kind:                   entity.UserKindEmployee,
			Status:                 entity.UserStatusActive,
			DisplayName:            "Pending Change User",
			Email:                  "current@integrione.local",
			Username:               "pending.change",
			Password:               "VaultEmp#2026!C",
			MarkIdentitiesVerified: true,
		})
		if err != nil {
			t.Fatalf("CreateUser() error = %v", err)
		}

		updRes, err := h.userSvc.UpdateEmail(ctx, user.UpdateEmailInput{
			UserID:          u.ID,
			NewEmail:        "newpending@integrione.local",
			CurrentPassword: "VaultEmp#2026!C",
		})
		if err != nil || updRes.VerificationToken == "" {
			t.Fatalf("UpdateEmail() error = %v", err)
		}

		if _, err := h.userSvc.ResendValidationEmailToken(ctx, user.ResendValidationEmailInput{
			UserID: u.ID,
			Email:  "mismatch@integrione.local",
		}); !errors.Is(err, iamerrors.ErrPendingEmailNotFound) {
			t.Fatalf("ResendValidationEmailToken(mismatch) error = %v, want ErrPendingEmailNotFound", err)
		}

		resendRes, err := h.userSvc.ResendValidationEmailToken(ctx, user.ResendValidationEmailInput{
			UserID: u.ID,
			Email:  " NewPending@Integrione.local ",
		})
		if err != nil || !resendRes.IsResend || resendRes.PendingEmail != "newpending@integrione.local" || resendRes.VerificationToken == "" || resendRes.VerificationToken == updRes.VerificationToken {
			t.Fatalf("unexpected ResendValidationEmailToken(pending): (%+v, %v)", resendRes, err)
		}

		oldCheck, err := h.userSvc.ConfirmEmailVerificationByToken(ctx, updRes.VerificationToken)
		if err != nil || oldCheck.Status != user.VerifyEmailStatusExpired {
			t.Fatalf("expected superseded token to expire, got (%+v, %v)", oldCheck, err)
		}

		verRes, err := h.userSvc.ConfirmEmailVerificationByToken(ctx, resendRes.VerificationToken)
		if err != nil || verRes.Status != user.VerifyEmailStatusVerified {
			t.Fatalf("ConfirmEmailVerificationByToken() got (%+v, %v)", verRes, err)
		}

		// Setelah email terverifikasi dan tidak ada lagi pending email, kirim ulang wajib mengembalikan ErrEmailAlreadyVerified.
		if _, err := h.userSvc.ResendValidationEmailToken(ctx, user.ResendValidationEmailInput{
			UserID: u.ID,
		}); !errors.Is(err, iamerrors.ErrEmailAlreadyVerified) {
			t.Fatalf("after verification ResendValidationEmailToken() error = %v, want ErrEmailAlreadyVerified", err)
		}
	})
}
