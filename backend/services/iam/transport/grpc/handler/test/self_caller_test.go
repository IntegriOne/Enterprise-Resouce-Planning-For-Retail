package handler_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/integrione/erp-retail/backend/services/iam/config"
	iamv1 "github.com/integrione/erp-retail/backend/services/iam/contract/gen/iam/v1"
	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	domainrepo "github.com/integrione/erp-retail/backend/services/iam/domain/repository"
	domainservice "github.com/integrione/erp-retail/backend/services/iam/domain/service"
	"github.com/integrione/erp-retail/backend/services/iam/pkg/crypto"
	iamrepo "github.com/integrione/erp-retail/backend/services/iam/repository"
	"github.com/integrione/erp-retail/backend/services/iam/services"
	"github.com/integrione/erp-retail/backend/services/iam/services/user"
	"github.com/integrione/erp-retail/backend/services/iam/transport/grpc/handler"
	"github.com/integrione/erp-retail/backend/shared/authcontext"
)

type failingSnapshotUserRepository struct {
	domainrepo.UserRepository
	err error
}

func (r failingSnapshotUserRepository) SelectAccessSnapshotByUserID(context.Context, string, time.Time) (*entity.AccessSnapshot, error) {
	return nil, r.err
}

type failingUpdateUserRepository struct {
	domainrepo.UserRepository
	err error
}

func (r failingUpdateUserRepository) Update(context.Context, *entity.User) error {
	return r.err
}

type failingSelectUserRepository struct {
	domainrepo.UserRepository
	err error
}

func (r failingSelectUserRepository) SelectByID(context.Context, string) (*entity.User, error) {
	return nil, r.err
}

func buildMemRepoRegistry(store *memStore) *iamrepo.Registry {
	return &iamrepo.Registry{
		Transactor:        &memTransactor{},
		Users:             &memUserRepo{store},
		UserIdentities:    &memIdentityRepo{store},
		UserCredentials:   &memCredentialRepo{store},
		UserTOTPFactors:   &memTOTPRepo{store},
		UserRecoveryCodes: &memRecoveryCodeRepo{store},
		UserTokens:        &memTokenRepo{store},
		Sessions:          &memSessionRepo{store},
		Roles:             &memRoleRepo{store},
		RolePermissions:   &memRolePermRepo{store},
		UserRoles:         &memUserRoleRepo{store},
		Permissions:       &memPermRepo{store},
		Groups:            &memGroupRepo{store},
		GroupMembers:      &memGroupMemberRepo{store},
		GroupRoles:        &memGroupRoleRepo{store},
	}
}

func TestBootstrapBackfillAndCreateUserBaselineSelfRoles(t *testing.T) {
	t.Parallel()

	env := startBufconnServer(t)
	ctx := context.Background()
	repos := buildMemRepoRegistry(env.store)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	// Buat dua user historis sebelum bootstrap berjalan untuk menguji backfill idempoten.
	idGen := crypto.NewUUIDGenerator()
	legacyCustID, _ := idGen.NewUUID()
	legacyEmpID, _ := idGen.NewUUID()
	now := time.Now()
	legacyCust, err := entity.NewUser(legacyCustID, entity.UserKindCustomer, entity.UserStatusActive, "Legacy Customer", now)
	if err != nil {
		t.Fatalf("NewUser(legacyCust) error = %v", err)
	}
	legacyEmp, err := entity.NewUser(legacyEmpID, entity.UserKindEmployee, entity.UserStatusActive, "Legacy Employee", now)
	if err != nil {
		t.Fatalf("NewUser(legacyEmp) error = %v", err)
	}
	env.store.users[legacyCustID] = legacyCust
	env.store.users[legacyEmpID] = legacyEmp

	if err := services.BootstrapSystemPrincipalAndRBAC(ctx, repos, env.registry, config.SystemBootstrapConfig{}, logger); err != nil {
		t.Fatalf("BootstrapSystemPrincipalAndRBAC() first run error = %v", err)
	}
	legacy, err := repos.Permissions.SelectByCode(ctx, "iam.users.read")
	if err != nil {
		t.Fatalf("SelectByCode() error = %v", err)
	}
	env.store.mu.Lock()
	env.store.permissions[legacy.ID].Resource = "users"
	env.store.mu.Unlock()
	// Jalankan ulang bootstrap untuk memastikan backfill tidak menduplikasi penugasan peran.
	if err := services.BootstrapSystemPrincipalAndRBAC(ctx, repos, env.registry, config.SystemBootstrapConfig{}, logger); err != nil {
		t.Fatalf("BootstrapSystemPrincipalAndRBAC() second run error = %v", err)
	}
	repaired, err := repos.Permissions.SelectByCode(ctx, "iam.users.read")
	if err != nil || repaired.Resource != "iam.users" || repaired.Action != "read" {
		t.Fatalf("bootstrap permission metadata = (%+v, %v), want iam.users/read", repaired, err)
	}

	assertUserHasOnlyBaselineSelfRole := func(t *testing.T, userID, wantRoleCode string) {
		t.Helper()
		urs, err := repos.UserRoles.SelectByUserID(ctx, userID, true, time.Now())
		if err != nil {
			t.Fatalf("SelectByUserID(%s) error = %v", userID, err)
		}
		if len(urs) != 1 {
			t.Fatalf("user %s has %d userRoles, want 1", userID, len(urs))
		}
		if urs[0].Scope.Type != entity.ScopeTypeSelf || urs[0].Scope.ID == nil || *urs[0].Scope.ID != userID {
			t.Fatalf("user %s baseline scope = %+v, want self(%s)", userID, urs[0].Scope, userID)
		}
		r, err := repos.Roles.SelectByID(ctx, urs[0].RoleID)
		if err != nil {
			t.Fatalf("Roles.SelectByID(%s) error = %v", urs[0].RoleID, err)
		}
		if r.Code != wantRoleCode {
			t.Fatalf("user %s role code = %q, want %q", userID, r.Code, wantRoleCode)
		}
	}

	assertUserHasOnlyBaselineSelfRole(t, legacyCustID, "customer")
	assertUserHasOnlyBaselineSelfRole(t, legacyEmpID, "iam_selfservice")

	// Uji auto-assignment saat CreateUser untuk ke-4 UserKind utama.
	kinds := []struct {
		kind     entity.UserKind
		wantRole string
		email    string
		uname    string
	}{
		{entity.UserKindCustomer, "customer", "newcust@example.com", "newcust"},
		{entity.UserKindEmployee, "iam_selfservice", "newemp@example.com", "newemp"},
		{entity.UserKindVendor, "iam_selfservice", "newvendor@example.com", "newvendor"},
		{entity.UserKindSupplier, "iam_selfservice", "newsupplier@example.com", "newsupplier"},
	}

	for _, tc := range kinds {
		created, err := env.registry.Users.CreateUser(ctx, user.CreateUserInput{
			Kind:                   tc.kind,
			Status:                 entity.UserStatusActive,
			DisplayName:            "User " + tc.uname,
			Email:                  tc.email,
			Username:               tc.uname,
			Password:               "ValidPass!2026Strong",
			MarkIdentitiesVerified: true,
		})
		if err != nil {
			t.Fatalf("CreateUser(%s) error = %v", tc.kind, err)
		}
		assertUserHasOnlyBaselineSelfRole(t, created.ID, tc.wantRole)
	}
	sensitive, err := repos.Permissions.SelectByCode(ctx, "iam.users.write")
	if err != nil {
		t.Fatal(err)
	}
	env.store.mu.Lock()
	env.store.permissions[sensitive.ID].IsSensitive = false
	env.store.mu.Unlock()
	if err := services.BootstrapSystemPrincipalAndRBAC(ctx, repos, env.registry, config.SystemBootstrapConfig{}, logger); err == nil {
		t.Fatal("bootstrap must reject a sensitivity mismatch in the existing catalog")
	}
}

func TestBootstrapFailsWhenAccessSnapshotRefreshFails(t *testing.T) {
	t.Parallel()

	env := startBufconnServer(t)
	ctx := context.Background()
	repos := buildMemRepoRegistry(env.store)
	cfg := config.SystemBootstrapConfig{
		Email:    "bootstrap-snapshot@example.com",
		Username: "bootstrap.snapshot",
		Password: "ValidPass!2026Strong",
	}
	if err := services.BootstrapSystemPrincipalAndRBAC(ctx, repos, env.registry, cfg, nil); err != nil {
		t.Fatalf("initial bootstrap error = %v", err)
	}

	snapshotErr := errors.New("snapshot unavailable")
	repos.Users = failingSnapshotUserRepository{UserRepository: repos.Users, err: snapshotErr}
	reg, err := services.NewRegistry(repos, services.SecurityOptions{
		AllowEphemeralDevKey:  true,
		AllowEphemeralCSRFKey: true,
	})
	if err != nil {
		t.Fatalf("NewRegistry() error = %v", err)
	}
	if err := services.BootstrapSystemPrincipalAndRBAC(ctx, repos, reg, cfg, nil); !errors.Is(err, snapshotErr) {
		t.Fatalf("bootstrap error = %v, want snapshot error", err)
	}
}

func TestBootstrapFailsWhenExistingUserActivationCannotPersist(t *testing.T) {
	t.Parallel()

	env := startBufconnServer(t)
	ctx := context.Background()
	repos := buildMemRepoRegistry(env.store)
	cfg := config.SystemBootstrapConfig{
		Email:    "bootstrap-activation@example.com",
		Username: "bootstrap.activation",
		Password: "ValidPass!2026Strong",
	}
	if err := services.BootstrapSystemPrincipalAndRBAC(ctx, repos, env.registry, cfg, nil); err != nil {
		t.Fatalf("initial bootstrap error = %v", err)
	}
	identity, err := repos.UserIdentities.FindUserIdentityByEmailKind(ctx, cfg.Email)
	if err != nil {
		t.Fatal(err)
	}
	env.store.mu.Lock()
	env.store.users[identity.UserID].Status = entity.UserStatusPending
	env.store.mu.Unlock()

	updateErr := errors.New("activation update unavailable")
	repos.Users = failingUpdateUserRepository{UserRepository: repos.Users, err: updateErr}
	reg, err := services.NewRegistry(repos, services.SecurityOptions{
		AllowEphemeralDevKey:  true,
		AllowEphemeralCSRFKey: true,
	})
	if err != nil {
		t.Fatalf("NewRegistry() error = %v", err)
	}
	if err := services.BootstrapSystemPrincipalAndRBAC(ctx, repos, reg, cfg, nil); !errors.Is(err, updateErr) {
		t.Fatalf("bootstrap error = %v, want activation update error", err)
	}
}

func TestBootstrapFailsWhenConfiguredUserLookupFails(t *testing.T) {
	t.Parallel()

	env := startBufconnServer(t)
	repos := buildMemRepoRegistry(env.store)
	lookupErr := errors.New("user lookup unavailable")
	repos.Users = failingSelectUserRepository{UserRepository: repos.Users, err: lookupErr}
	reg, err := services.NewRegistry(repos, services.SecurityOptions{
		AllowEphemeralDevKey:  true,
		AllowEphemeralCSRFKey: true,
	})
	if err != nil {
		t.Fatalf("NewRegistry() error = %v", err)
	}
	cfg := config.SystemBootstrapConfig{
		UserUID:  "99999999-9999-4999-8999-999999999999",
		Email:    "bootstrap-lookup@example.com",
		Username: "bootstrap.lookup",
		Password: "ValidPass!2026Strong",
	}
	if err := services.BootstrapSystemPrincipalAndRBAC(context.Background(), repos, reg, cfg, nil); !errors.Is(err, lookupErr) {
		t.Fatalf("bootstrap error = %v, want user lookup error", err)
	}
}

func TestHandlerEnforceSelfCallerAndSessionGuards(t *testing.T) {
	t.Parallel()

	env := startBufconnServer(t)
	ctx := context.Background()
	idGen := crypto.NewUUIDGenerator()
	now := domainservice.SystemClock{}.Now()

	uA, err := env.registry.Users.CreateUser(ctx, user.CreateUserInput{
		Kind:                   entity.UserKindCustomer,
		Status:                 entity.UserStatusActive,
		DisplayName:            "Alice Self",
		Email:                  "aliceself@example.com",
		Username:               "aliceself",
		Password:               "ValidPass!2026Strong",
		MarkIdentitiesVerified: true,
	})
	if err != nil {
		t.Fatalf("CreateUser(uA) error = %v", err)
	}
	uB, err := env.registry.Users.CreateUser(ctx, user.CreateUserInput{
		Kind:                   entity.UserKindCustomer,
		Status:                 entity.UserStatusActive,
		DisplayName:            "Bob Self",
		Email:                  "bobself@example.com",
		Username:               "bobself",
		Password:               "ValidPass!2026Strong",
		MarkIdentitiesVerified: true,
	})
	if err != nil {
		t.Fatalf("CreateUser(uB) error = %v", err)
	}

	sessAID, _ := idGen.NewUUID()
	sessBID, _ := idGen.NewUUID()
	uHandler, err := handler.NewUserHandler(env.registry.Users, nil)
	if err != nil {
		t.Fatalf("NewUserHandler() error = %v", err)
	}
	aHandler, err := handler.NewAuthHandler(env.registry.Users, nil)
	if err != nil {
		t.Fatalf("NewAuthHandler() error = %v", err)
	}
	sHandler, err := handler.NewSessionHandler(env.registry.Sessions, nil)
	if err != nil {
		t.Fatalf("NewSessionHandler() error = %v", err)
	}

	callerA := authcontext.CallerContext{
		PrincipalID:   uA.ID,
		SessionID:     sessAID,
		UserKind:      string(entity.UserKindCustomer),
		Authenticated: true,
	}
	ctxCallerA := authcontext.WithCaller(ctx, callerA)
	if _, err := uHandler.UpdateUserProfile(ctx, &iamv1.UpdateUserProfileRequest{UserId: uA.ID, DisplayName: "Anonymous Update"}); status.Code(err) != codes.Unauthenticated {
		t.Fatalf("UpdateUserProfile(anonymous) error = %v, want Unauthenticated", err)
	}
	if _, err := aHandler.EnrollTotp(ctx, &iamv1.EnrollTotpRequest{UserId: uA.ID}); status.Code(err) != codes.Unauthenticated {
		t.Fatalf("EnrollTotp(anonymous) error = %v, want Unauthenticated", err)
	}
	if _, err := sHandler.IssueSession(ctxCallerA, &iamv1.IssueSessionRequest{UserId: uB.ID}); status.Code(err) != codes.PermissionDenied {
		t.Fatalf("IssueSession(cross-user) error = %v, want PermissionDenied", err)
	}
	if _, err := sHandler.IssueSession(ctxCallerA, &iamv1.IssueSessionRequest{UserId: uA.ID, MfaVerified: true}); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("IssueSession(unverified MFA) error = %v, want InvalidArgument", err)
	}
	if issued, err := sHandler.IssueSession(ctxCallerA, &iamv1.IssueSessionRequest{UserId: uA.ID}); err != nil || issued.GetIssuedSession().GetSessionToken() == "" {
		t.Fatalf("IssueSession(self) = (%+v, %v), want issued token", issued, err)
	}

	assertPermissionDenied := func(t *testing.T, name string, err error) {
		t.Helper()
		if err == nil {
			t.Fatalf("%s unexpectedly succeeded", name)
		}
		if st, ok := status.FromError(err); !ok || st.Code() != codes.PermissionDenied {
			t.Fatalf("%s error code = %v (%v), want PermissionDenied", name, st.Code(), err)
		}
	}

	// Pemanggil A dilarang menyentuh resource self milik B pada seluruh RPC self UserHandler & AuthHandler.
	_, err = uHandler.UpdateUserProfile(ctxCallerA, &iamv1.UpdateUserProfileRequest{UserId: uB.ID, DisplayName: "Hacked"})
	assertPermissionDenied(t, "UpdateUserProfile(cross)", err)

	_, err = uHandler.UpdateUserEmail(ctxCallerA, &iamv1.UpdateUserEmailRequest{UserId: uB.ID, Email: "evil@example.com", CurrentPassword: "ValidPass!2026Strong"})
	assertPermissionDenied(t, "UpdateUserEmail(cross)", err)

	_, err = uHandler.UpdateUserUsername(ctxCallerA, &iamv1.UpdateUserUsernameRequest{UserId: uB.ID, Username: "hackedname"})
	assertPermissionDenied(t, "UpdateUserUsername(cross)", err)

	_, err = uHandler.DeletePendingUserEmail(ctxCallerA, &iamv1.DeletePendingUserEmailRequest{UserId: uB.ID})
	assertPermissionDenied(t, "DeletePendingUserEmail(cross)", err)

	_, err = uHandler.ResendValidationEmailToken(ctxCallerA, &iamv1.ResendValidationEmailTokenRequest{UserId: uB.ID})
	assertPermissionDenied(t, "ResendValidationEmailToken(cross)", err)

	_, err = aHandler.ChangePassword(ctxCallerA, &iamv1.ChangePasswordRequest{UserId: uB.ID, CurrentPassword: "ValidPass!2026Strong", NewPassword: "AnotherValid!2026Strong"})
	assertPermissionDenied(t, "ChangePassword(cross)", err)

	_, err = aHandler.EnrollTotp(ctxCallerA, &iamv1.EnrollTotpRequest{UserId: uB.ID})
	assertPermissionDenied(t, "EnrollTotp(cross)", err)

	_, err = aHandler.ConfirmTotp(ctxCallerA, &iamv1.ConfirmTotpRequest{UserId: uB.ID, Code: "123456"})
	assertPermissionDenied(t, "ConfirmTotp(cross)", err)

	_, err = aHandler.DisableTotp(ctxCallerA, &iamv1.DisableTotpRequest{UserId: uB.ID})
	assertPermissionDenied(t, "DisableTotp(cross)", err)

	_, err = aHandler.RegenerateRecoveryCodes(ctxCallerA, &iamv1.RegenerateRecoveryCodesRequest{UserId: uB.ID})
	assertPermissionDenied(t, "RegenerateRecoveryCodes(cross)", err)

	_, err = aHandler.CountUnusedRecoveryCodes(ctxCallerA, &iamv1.CountUnusedRecoveryCodesRequest{UserId: uB.ID})
	assertPermissionDenied(t, "CountUnusedRecoveryCodes(cross)", err)

	_, err = aHandler.VerifySessionMfa(ctxCallerA, &iamv1.VerifySessionMfaRequest{UserId: uB.ID, SessionId: sessBID, Code: "123456"})
	assertPermissionDenied(t, "VerifySessionMfa(cross-user)", err)

	// Sama user_id tetapi session_id milik sesi lain juga harus ditolak.
	_, err = aHandler.VerifySessionMfa(ctxCallerA, &iamv1.VerifySessionMfaRequest{UserId: uA.ID, SessionId: sessBID, Code: "123456"})
	assertPermissionDenied(t, "VerifySessionMfa(cross-session)", err)

	callerWithoutSession := callerA
	callerWithoutSession.SessionID = ""
	_, err = aHandler.VerifySessionMfa(authcontext.WithCaller(ctx, callerWithoutSession), &iamv1.VerifySessionMfaRequest{UserId: uA.ID, SessionId: sessAID, Code: "123456"})
	assertPermissionDenied(t, "VerifySessionMfa(missing caller session)", err)

	// Pemanggil A mengubah profilnya sendiri harus berhasil.
	_ = now
	upd, err := uHandler.UpdateUserProfile(ctxCallerA, &iamv1.UpdateUserProfileRequest{UserId: uA.ID, DisplayName: "Alice Updated"})
	if err != nil {
		t.Fatalf("UpdateUserProfile(self) error = %v", err)
	}
	if upd.GetUser().GetDisplayName() != "Alice Updated" {
		t.Fatalf("DisplayName = %q, want Alice Updated", upd.GetUser().GetDisplayName())
	}
}
