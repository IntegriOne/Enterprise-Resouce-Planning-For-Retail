package user_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
	"github.com/integrione/erp-retail/backend/services/iam/services/user"
)

func TestUserService_LoginLockoutPasswordChangeAndReset(t *testing.T) {
	t.Parallel()

	h := newTestHarness(t)
	u, err := h.userSvc.CreateUser(context.Background(), user.CreateUserInput{
		Kind:                   entity.UserKindEmployee,
		Status:                 entity.UserStatusActive,
		DisplayName:            "Andi Pratama",
		Email:                  "andi@integrione.local",
		Username:               "andi.p",
		Password:               "Xr8#SilverLock!2026",
		MarkIdentitiesVerified: true,
	})
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}

	// Tambahkan peran dengan redirect route ke /identity agar ResolvePostLoginPath memilihnya.
	role, _ := entity.NewRole(
		"11111111-1111-4111-8111-111111111111",
		"iam_admin", "IAM Admin", "Manages IAM", "iam", "/identity", 90, true, h.clock.Now(),
	)
	h.store.roles[u.ID] = []entity.Role{*role}

	// 4 kali gagal login menghasilkan ErrInvalidCredentials, ke-5 mengunci akun secara diam-diam tanpa membocorkan status lockout ke penyerang.
	for i := 0; i < 4; i++ {
		if _, err := h.userSvc.AuthenticatePassword(context.Background(), user.AuthenticatePasswordInput{
			Identifier: "andi.p",
			Password:   "Wrong#Pass!1",
		}); !errors.Is(err, iamerrors.ErrInvalidCredentials) {
			t.Fatalf("attempt %d: expected ErrInvalidCredentials, got %v", i+1, err)
		}
	}
	if _, err := h.userSvc.AuthenticatePassword(context.Background(), user.AuthenticatePasswordInput{
		Identifier: "andi.p",
		Password:   "Wrong#Pass!1",
	}); !errors.Is(err, iamerrors.ErrInvalidCredentials) || errors.Is(err, iamerrors.ErrAccountLocked) {
		t.Fatalf("attempt 5: expected ambiguous ErrInvalidCredentials without lockout clue, got %v", err)
	}

	// Meski kata sandi benar, tetap ditolak selama masa kunci 15 menit berlangsung dengan error ambigu yang sama.
	if _, err := h.userSvc.AuthenticatePassword(context.Background(), user.AuthenticatePasswordInput{
		Identifier: "andi@integrione.local",
		Password:   "Xr8#SilverLock!2026",
	}); !errors.Is(err, iamerrors.ErrInvalidCredentials) || errors.Is(err, iamerrors.ErrAccountLocked) {
		t.Fatalf("expected ambiguous ErrInvalidCredentials during silent lockout cooldown, got %v", err)
	}

	// Lewat 16 menit -> kunci berakhir dan login berhasil dengan redirect /identity.
	h.clock.Advance(16 * time.Minute)
	authRes, err := h.userSvc.AuthenticatePassword(context.Background(), user.AuthenticatePasswordInput{
		Identifier: "andi@integrione.local",
		Password:   "Xr8#SilverLock!2026",
		IPAddress:  "127.0.0.1",
	})
	if err != nil {
		t.Fatalf("AuthenticatePassword(after cooldown) error = %v", err)
	}
	if authRes.Status != user.AuthStatusAuthenticated || authRes.RedirectPath != "/identity" {
		t.Fatalf("unexpected auth result: %+v", authRes)
	}

	// Ganti kata sandi mencabut seluruh sesi lama dan mengganti hash.
	changeRes, err := h.userSvc.ChangePassword(context.Background(), user.ChangePasswordInput{
		UserID:          u.ID,
		CurrentPassword: "Xr8#SilverLock!2026",
		NewPassword:     "Tz5#CopperFort!2027",
	})
	if err != nil {
		t.Fatalf("ChangePassword() error = %v", err)
	}
	if !changeRes.Changed || changeRes.SessionsRevoked < 1 || changeRes.PasswordChangedAt.IsZero() {
		t.Fatalf("unexpected ChangePassword result: %+v", changeRes)
	}
	if _, err := h.sessSvc.Authenticate(context.Background(), authRes.IssuedSession.SessionToken); !errors.Is(err, iamerrors.ErrSessionRevoked) {
		t.Fatalf("expected old session revoked after ChangePassword, got %v", err)
	}

	// Minta token reset kata sandi sekali pakai dan verifikasi token tidak dapat dipakai dua kali.
	resetTok, err := h.userSvc.RequestPasswordReset(context.Background(), "andi@integrione.local")
	if err != nil || resetTok.PlaintextToken == "" {
		t.Fatalf("RequestPasswordReset() = (%+v, %v)", resetTok, err)
	}
	if err := h.userSvc.ConfirmPasswordReset(context.Background(), user.ConfirmPasswordResetInput{
		RawToken:    resetTok.PlaintextToken,
		NewPassword: "Wn6#CobaltVault!2028",
	}); err != nil {
		t.Fatalf("ConfirmPasswordReset() error = %v", err)
	}
	if err := h.userSvc.ConfirmPasswordReset(context.Background(), user.ConfirmPasswordResetInput{
		RawToken:    resetTok.PlaintextToken,
		NewPassword: "Ql2#ReplayGuard!2029",
	}); !errors.Is(err, iamerrors.ErrTokenAlreadyUsed) {
		t.Fatalf("expected ErrTokenAlreadyUsed on reset token replay, got %v", err)
	}

	// Kata sandi hasil reset berhasil login.
	authRes2, err := h.userSvc.AuthenticatePassword(context.Background(), user.AuthenticatePasswordInput{
		Identifier: "andi.p",
		Password:   "Wn6#CobaltVault!2028",
	})
	if err != nil || authRes2.IssuedSession == nil {
		t.Fatalf("login with reset password failed: (%+v, %v)", authRes2, err)
	}
}

func TestUserService_UnknownLoginAndDisabledAccountResetMasking(t *testing.T) {
	t.Parallel()

	h := newTestHarness(t)
	ctx := context.Background()

	// Identifier tidak dikenal wajib tetap mengembalikan ErrInvalidCredentials tanpa membocorkan eksistensi akun.
	if _, err := h.userSvc.AuthenticatePassword(ctx, user.AuthenticatePasswordInput{
		Identifier: "unknown.user@integrione.local",
		Password:   "IrrelevantPass!2026",
	}); !errors.Is(err, iamerrors.ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials for unknown identifier, got %v", err)
	}

	uSusp, err := h.userSvc.CreateUser(ctx, user.CreateUserInput{
		Kind:                   entity.UserKindEmployee,
		Status:                 entity.UserStatusActive,
		DisplayName:            "Suspended Reset User",
		Email:                  "susp.reset@integrione.local",
		Username:               "susp.reset",
		Password:               "Wq4#CobaltKey!2026",
		MarkIdentitiesVerified: true,
	})
	if err != nil {
		t.Fatalf("CreateUser(uSusp) error = %v", err)
	}
	tokSusp, err := h.userSvc.RequestPasswordReset(ctx, "susp.reset@integrione.local")
	if err != nil {
		t.Fatalf("RequestPasswordReset(uSusp) error = %v", err)
	}
	if _, err := h.userSvc.Suspend(ctx, uSusp.ID); err != nil {
		t.Fatalf("Suspend(uSusp) error = %v", err)
	}
	if err := h.userSvc.ConfirmPasswordReset(ctx, user.ConfirmPasswordResetInput{
		RawToken:    tokSusp.PlaintextToken,
		NewPassword: "Zx9#FreshVault!2026",
	}); !errors.Is(err, iamerrors.ErrTokenNotFound) {
		t.Fatalf("expected ErrTokenNotFound on suspended user reset confirmation, got %v", err)
	}

	uDel, err := h.userSvc.CreateUser(ctx, user.CreateUserInput{
		Kind:                   entity.UserKindEmployee,
		Status:                 entity.UserStatusActive,
		DisplayName:            "Deleted Reset User",
		Email:                  "del.reset@integrione.local",
		Username:               "del.reset",
		Password:               "Wq4#CobaltKey!2026",
		MarkIdentitiesVerified: true,
	})
	if err != nil {
		t.Fatalf("CreateUser(uDel) error = %v", err)
	}
	tokDel, err := h.userSvc.RequestPasswordReset(ctx, "del.reset@integrione.local")
	if err != nil {
		t.Fatalf("RequestPasswordReset(uDel) error = %v", err)
	}
	if err := h.userSvc.SoftDelete(ctx, uDel.ID); err != nil {
		t.Fatalf("SoftDelete(uDel) error = %v", err)
	}
	if err := h.userSvc.ConfirmPasswordReset(ctx, user.ConfirmPasswordResetInput{
		RawToken:    tokDel.PlaintextToken,
		NewPassword: "Zx9#FreshVault!2026",
	}); !errors.Is(err, iamerrors.ErrTokenNotFound) {
		t.Fatalf("expected ErrTokenNotFound on deleted user reset confirmation, got %v", err)
	}
}
