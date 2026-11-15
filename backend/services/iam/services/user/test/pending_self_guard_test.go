package user_test

import (
	"context"
	"encoding/base32"
	"errors"
	"net/url"
	"strings"
	"testing"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
	"github.com/integrione/erp-retail/backend/services/iam/services/user"
)

func seedBaselineSelfRolesInHarness(t *testing.T, h *testHarness) {
	t.Helper()
	now := h.clock.Now()
	custID, _ := h.idGen.NewUUID()
	custRole, err := entity.NewRole(custID, "customer", "Customer", "Baseline customer self-service", "commerce", "", 10, true, now)
	if err != nil {
		t.Fatalf("NewRole(customer) error = %v", err)
	}
	selfID, _ := h.idGen.NewUUID()
	selfRole, err := entity.NewRole(selfID, "iam_selfservice", "IAM Self Service", "Baseline internal/partner self-service", "iam", "", 5, true, now)
	if err != nil {
		t.Fatalf("NewRole(iam_selfservice) error = %v", err)
	}
	h.store.roleDefs[custRole.ID] = custRole
	h.store.roleDefs[selfRole.ID] = selfRole
}

func extractSecretFromOTPAuthURI(t *testing.T, rawURI string) []byte {
	t.Helper()
	parsed, err := url.Parse(rawURI)
	if err != nil {
		t.Fatalf("parse otpauth URI: %v", err)
	}
	rawSecret := strings.ToUpper(strings.TrimSpace(parsed.Query().Get("secret")))
	secret, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(rawSecret)
	if err != nil {
		t.Fatalf("decode TOTP base32 secret: %v", err)
	}
	return secret
}

func TestUserService_AutoAssignedSelfServiceBlockedWhilePendingAndUnverified(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	h := newTestHarness(t)
	seedBaselineSelfRolesInHarness(t, h)

	cases := []struct {
		kind         entity.UserKind
		wantRoleCode string
	}{
		{kind: entity.UserKindCustomer, wantRoleCode: "customer"},
		{kind: entity.UserKindEmployee, wantRoleCode: "iam_selfservice"},
		{kind: entity.UserKindVendor, wantRoleCode: "iam_selfservice"},
		{kind: entity.UserKindSupplier, wantRoleCode: "iam_selfservice"},
	}

	for _, tc := range cases {
		t.Run(string(tc.kind), func(t *testing.T) {
			email := string(tc.kind) + ".pending@integrione.local"
			username := string(tc.kind) + "_pending"
			pass := "StrongPass#2026!"

			created, err := h.userSvc.CreateUser(ctx, user.CreateUserInput{
				Kind:        tc.kind,
				Status:      entity.UserStatusPending,
				DisplayName: "Pending " + string(tc.kind),
				Email:       email,
				Username:    username,
				Password:    pass,
			})
			if err != nil {
				t.Fatalf("CreateUser(%s) error = %v", tc.kind, err)
			}
			if created.Status != entity.UserStatusPending {
				t.Fatalf("initial status = %s, want pending", created.Status)
			}

			// 1. Pastikan role dasar self-service SUDAH otomatis terpasang saat user PENDING dibuat.
			h.store.mu.Lock()
			urs := h.store.userRoles[created.ID]
			h.store.mu.Unlock()
			if len(urs) != 1 || urs[0].Scope.Type != entity.ScopeTypeSelf || urs[0].Scope.ID == nil || *urs[0].Scope.ID != created.ID {
				t.Fatalf("expected 1 baseline SELF-scoped role assignment, got %+v", urs)
			}
			if roleDef := h.store.roleDefs[urs[0].RoleID]; roleDef == nil || roleDef.Code != tc.wantRoleCode {
				t.Fatalf("assigned role = %+v, want %q", roleDef, tc.wantRoleCode)
			}

			// 2. Meski role sudah terpasang, login & seluruh layanan mandiri wajib DITOLAK saat status masih PENDING.
			if _, err := h.userSvc.AuthenticatePassword(ctx, user.AuthenticatePasswordInput{Identifier: email, Password: pass}); !errors.Is(err, iamerrors.ErrInvalidCredentials) {
				t.Fatalf("pending login by email error = %v, want ErrInvalidCredentials", err)
			}
			if _, err := h.userSvc.AuthenticatePassword(ctx, user.AuthenticatePasswordInput{Identifier: username, Password: pass}); !errors.Is(err, iamerrors.ErrInvalidCredentials) {
				t.Fatalf("pending login by username error = %v, want ErrInvalidCredentials", err)
			}
			if _, err := h.userSvc.UpdateProfile(ctx, user.UpdateProfileInput{UserID: created.ID, DisplayName: "Blocked"}); !errors.Is(err, iamerrors.ErrUserInactive) {
				t.Fatalf("pending UpdateProfile error = %v, want ErrUserInactive", err)
			}
			if _, err := h.userSvc.UpdateEmail(ctx, user.UpdateEmailInput{UserID: created.ID, NewEmail: "new." + email, CurrentPassword: pass}); !errors.Is(err, iamerrors.ErrUserInactive) {
				t.Fatalf("pending UpdateEmail error = %v, want ErrUserInactive", err)
			}
			if _, err := h.userSvc.UpdateUsername(ctx, user.UpdateUsernameInput{UserID: created.ID, Username: username + "_new"}); !errors.Is(err, iamerrors.ErrUserInactive) {
				t.Fatalf("pending UpdateUsername error = %v, want ErrUserInactive", err)
			}
			if _, err := h.userSvc.DeletePendingEmail(ctx, created.ID); !errors.Is(err, iamerrors.ErrUserInactive) {
				t.Fatalf("pending DeletePendingEmail error = %v, want ErrUserInactive", err)
			}
			if _, err := h.userSvc.ChangePassword(ctx, user.ChangePasswordInput{UserID: created.ID, CurrentPassword: pass, NewPassword: "NewStrongPass#2026!"}); !errors.Is(err, iamerrors.ErrUserInactive) {
				t.Fatalf("pending ChangePassword error = %v, want ErrUserInactive", err)
			}
			if _, err := h.userSvc.EnrollTOTP(ctx, created.ID, email); !errors.Is(err, iamerrors.ErrUserInactive) {
				t.Fatalf("pending EnrollTOTP error = %v, want ErrUserInactive", err)
			}
			if _, err := h.userSvc.ConfirmTOTP(ctx, created.ID, "123456"); !errors.Is(err, iamerrors.ErrUserInactive) {
				t.Fatalf("pending ConfirmTOTP error = %v, want ErrUserInactive", err)
			}
			if err := h.userSvc.DisableTOTP(ctx, created.ID); !errors.Is(err, iamerrors.ErrUserInactive) {
				t.Fatalf("pending DisableTOTP error = %v, want ErrUserInactive", err)
			}
			if _, err := h.userSvc.RegenerateRecoveryCodes(ctx, created.ID); !errors.Is(err, iamerrors.ErrUserInactive) {
				t.Fatalf("pending RegenerateRecoveryCodes error = %v, want ErrUserInactive", err)
			}
			fakeSessID, _ := h.idGen.NewUUID()
			if _, err := h.userSvc.VerifySessionMFA(ctx, created.ID, fakeSessID, "123456"); !errors.Is(err, iamerrors.ErrUserInactive) {
				t.Fatalf("pending VerifySessionMFA error = %v, want ErrUserInactive", err)
			}

			// 3. Jika diaktifkan tanpa verifikasi email (is_verified=false), login tetap ditolak.
			activatedOnly, err := h.userSvc.Activate(ctx, created.ID)
			if err != nil || activatedOnly.Status != entity.UserStatusActive {
				t.Fatalf("Activate() = (%+v, %v), want active", activatedOnly, err)
			}
			if _, err := h.userSvc.AuthenticatePassword(ctx, user.AuthenticatePasswordInput{Identifier: email, Password: pass}); !errors.Is(err, iamerrors.ErrInvalidCredentials) {
				t.Fatalf("active-but-unverified email login error = %v, want ErrInvalidCredentials", err)
			}
			if _, err := h.userSvc.AuthenticatePassword(ctx, user.AuthenticatePasswordInput{Identifier: username, Password: pass}); !errors.Is(err, iamerrors.ErrInvalidCredentials) {
				t.Fatalf("active-but-unverified username login error = %v, want ErrInvalidCredentials", err)
			}

			// 4. Uji alur nyata pada akun PENDING baru: verifikasi email otomatis memverifikasi identitas + mengaktifkan akun -> self-service langsung terbuka.
			verEmail := string(tc.kind) + ".verify@integrione.local"
			verUsername := string(tc.kind) + "_verify"
			pendingForVerify, err := h.userSvc.CreateUser(ctx, user.CreateUserInput{
				Kind:        tc.kind,
				Status:      entity.UserStatusPending,
				DisplayName: "Verify " + string(tc.kind),
				Email:       verEmail,
				Username:    verUsername,
				Password:    pass,
			})
			if err != nil {
				t.Fatalf("CreateUser(verify) error = %v", err)
			}
			resendRes, err := h.userSvc.ResendValidationEmailToken(ctx, user.ResendValidationEmailInput{Email: verEmail})
			if err != nil || resendRes.VerificationToken == "" {
				t.Fatalf("ResendValidationEmailToken() = (%+v, %v)", resendRes, err)
			}
			verRes, err := h.userSvc.ConfirmEmailVerificationByToken(ctx, resendRes.VerificationToken)
			if err != nil || verRes.Status != user.VerifyEmailStatusVerified {
				t.Fatalf("ConfirmEmailVerificationByToken() = (%+v, %v)", verRes, err)
			}
			if verRes.User == nil || verRes.User.Status != entity.UserStatusActive {
				t.Fatalf("verified user status = %+v, want active", verRes.User)
			}

			// Login via email dan username kini langsung berhasil + seluruh self-service bekerja.
			if authEmail, err := h.userSvc.AuthenticatePassword(ctx, user.AuthenticatePasswordInput{Identifier: verEmail, Password: pass}); err != nil || authEmail.Status != user.AuthStatusAuthenticated {
				t.Fatalf("verified login by email = (%+v, %v)", authEmail, err)
			}
			if authUname, err := h.userSvc.AuthenticatePassword(ctx, user.AuthenticatePasswordInput{Identifier: verUsername, Password: pass}); err != nil || authUname.Status != user.AuthStatusAuthenticated {
				t.Fatalf("verified login by username = (%+v, %v)", authUname, err)
			}
			if updated, err := h.userSvc.UpdateProfile(ctx, user.UpdateProfileInput{UserID: pendingForVerify.ID, DisplayName: "Verified User"}); err != nil || updated.DisplayName != "Verified User" {
				t.Fatalf("verified UpdateProfile = (%+v, %v)", updated, err)
			}
			if _, err := h.userSvc.UpdateUsername(ctx, user.UpdateUsernameInput{UserID: pendingForVerify.ID, Username: verUsername + "_ok"}); err != nil {
				t.Fatalf("verified UpdateUsername error = %v", err)
			}
			enrolled, err := h.userSvc.EnrollTOTP(ctx, pendingForVerify.ID, verEmail)
			if err != nil {
				t.Fatalf("verified EnrollTOTP error = %v", err)
			}
			code, err := h.totpProv.GenerateCodeAt(extractSecretFromOTPAuthURI(t, enrolled.OTPAuthURI), h.clock.Now())
			if err != nil {
				t.Fatalf("GenerateCodeAt() error = %v", err)
			}
			if _, err := h.userSvc.ConfirmTOTP(ctx, pendingForVerify.ID, code); err != nil {
				t.Fatalf("verified ConfirmTOTP error = %v", err)
			}
			if codes, err := h.userSvc.RegenerateRecoveryCodes(ctx, pendingForVerify.ID); err != nil || len(codes) == 0 {
				t.Fatalf("verified RegenerateRecoveryCodes = (%d, %v)", len(codes), err)
			}
			if err := h.userSvc.DisableTOTP(ctx, pendingForVerify.ID); err != nil {
				t.Fatalf("verified DisableTOTP error = %v", err)
			}
			newPass := "UpdatedPass#2026!"
			if chg, err := h.userSvc.ChangePassword(ctx, user.ChangePasswordInput{UserID: pendingForVerify.ID, CurrentPassword: pass, NewPassword: newPass}); err != nil || !chg.Changed {
				t.Fatalf("verified ChangePassword = (%+v, %v)", chg, err)
			}
		})
	}
}
