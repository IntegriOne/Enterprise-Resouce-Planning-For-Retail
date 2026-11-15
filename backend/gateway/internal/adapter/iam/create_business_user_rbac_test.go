package iam

import (
	"context"
	"testing"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/browserauth"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	"github.com/integrione/erp-retail/backend/gateway/internal/source_connections"
	iamv1 "github.com/integrione/erp-retail/backend/services/iam/contract/gen/iam/v1"
	"github.com/integrione/erp-retail/backend/shared/authcontext"
	"github.com/integrione/erp-retail/backend/shared/authsession"
	"github.com/integrione/erp-retail/backend/shared/authz"
)

type fnRefresher struct {
	fn func(context.Context, string, authcontext.CallerContext) (*authz.AccessSnapshot, error)
}

func (r fnRefresher) RefreshUserAccessSnapshot(ctx context.Context, userID string, caller authcontext.CallerContext) (*authz.AccessSnapshot, error) {
	if r.fn == nil {
		return nil, nil
	}
	return r.fn(ctx, userID, caller)
}

type businessUserVariant struct {
	name     string
	wantKind model.UserKind
	invoke   func(*Adapter, context.Context, string) (bool, *model.User, []*model.MutationError, error)
}

func businessUserVariants() []businessUserVariant {
	return []businessUserVariant{
		{
			name:     "createEmployeeUser",
			wantKind: model.UserKindEmployee,
			invoke: func(a *Adapter, ctx context.Context, tag string) (bool, *model.User, []*model.MutationError, error) {
				res, err := a.CreateEmployeeUser(ctx, model.CreateEmployeeUserInput{
					FullName: "Employee " + tag,
					Email:    tag + "@integrione.test",
					Username: tag,
					Password: "StrongPassword#2026!",
				})
				if res == nil {
					return false, nil, nil, err
				}
				return res.WasSuccessful, res.User, res.Errors, err
			},
		},
		{
			name:     "createVendorUser",
			wantKind: model.UserKindVendor,
			invoke: func(a *Adapter, ctx context.Context, tag string) (bool, *model.User, []*model.MutationError, error) {
				res, err := a.CreateVendorUser(ctx, model.CreateVendorUserInput{
					FullName: "Vendor " + tag,
					Email:    tag + "@integrione.test",
					Username: tag,
					Password: "StrongPassword#2026!",
				})
				if res == nil {
					return false, nil, nil, err
				}
				return res.WasSuccessful, res.User, res.Errors, err
			},
		},
		{
			name:     "createSupplierUser",
			wantKind: model.UserKindSupplier,
			invoke: func(a *Adapter, ctx context.Context, tag string) (bool, *model.User, []*model.MutationError, error) {
				res, err := a.CreateSupplierUser(ctx, model.CreateSupplierUserInput{
					FullName: "Supplier " + tag,
					Email:    tag + "@integrione.test",
					Username: tag,
					Password: "StrongPassword#2026!",
				})
				if res == nil {
					return false, nil, nil, err
				}
				return res.WasSuccessful, res.User, res.Errors, err
			},
		},
	}
}

func seedAdapterSession(
	store *memSessionStore,
	rawTok, userID, userKind, status string,
	issuedAtUnix int64,
	mfaEnabled, mfaAuthenticated bool,
	roles []authz.RoleGrant,
	perms []authz.PermissionGrant,
	permRev int64,
	now time.Time,
) context.Context {
	tokHash := authsession.HashOpaqueToken(rawTok)
	store.sessions[tokHash] = &authsession.SessionValidationContext{
		Token: &authsession.SessionTokenRecord{
			SessionID:        "sess-" + userID,
			UserID:           userID,
			TokenHashHex:     tokHash,
			UserKind:         userKind,
			AuthEpoch:        1,
			IssuedAtUnix:     issuedAtUnix,
			AbsExpiresAtUnix: now.Add(8 * time.Hour).Unix(),
			MFAAuthenticated: mfaAuthenticated,
		},
		Security: &authsession.UserSecurityRecord{
			UserID:        userID,
			Status:        status,
			AuthEpoch:     1,
			PermRevision:  permRev,
			RecoveryEpoch: 1,
		},
		Snapshot: &authz.AccessSnapshot{
			PrincipalID:    userID,
			UserKind:       userKind,
			Status:         status,
			MFAEnabled:     mfaEnabled,
			AuthEpoch:      1,
			PermRevision:   permRev,
			RecoveryEpoch:  1,
			Roles:          roles,
			Permissions:    perms,
			SyncedAtUnix:   now.Unix(),
			FreshUntilUnix: now.Add(30 * time.Second).Unix(),
		},
	}
	return middleware.WithRequestMetadata(context.Background(), middleware.RequestMetadata{
		RawCookie:  rawTok,
		CookieHash: tokHash,
		ClientIP:   "127.0.0.10",
	})
}

// TestCreateBusinessUser_RBACAndSensitiveReauth memastikan createEmployeeUser, createVendorUser, dan createSupplierUser hanya dapat dipanggil oleh admin atau pemegang peran/izin iam.users.write global dengan sesi sensitif aktif.
func TestCreateBusinessUser_RBACAndSensitiveReauth(t *testing.T) {
	now := time.Now()
	store := newMemStore()

	var createCalls int
	usersMock := &mockUserClient{
		createFn: func(_ context.Context, in *iamv1.CreateUserRequest) (*iamv1.CreateUserResponse, error) {
			createCalls++
			return &iamv1.CreateUserResponse{
				User: &iamv1.User{
					Id:          "new-biz-user-1",
					Kind:        in.GetKind(),
					Status:      in.GetStatus(),
					DisplayName: in.GetDisplayName(),
					CreatedAt:   timestamppb.New(now),
					UpdatedAt:   timestamppb.New(now),
				},
			}, nil
		},
	}

	refreshedByUser := make(map[string]*authz.AccessSnapshot)
	authMgr, err := middleware.NewAuthManager(store, middleware.NewSnapshotLoader(fnRefresher{
		fn: func(_ context.Context, userID string, _ authcontext.CallerContext) (*authz.AccessSnapshot, error) {
			return refreshedByUser[userID], nil
		},
	}, time.Second, 30*time.Second), middleware.SessionConfig{CookieSecure: true})
	if err != nil {
		t.Fatalf("NewAuthManager failed: %v", err)
	}
	authzSvc := middleware.NewAuthorizer(authMgr, nil, 10*time.Minute)
	rv := browserauth.NewRedirectValidator([]string{"/", "/identity"})

	adp, err := NewAdapter(source_connections.IAMClientSet{
		Auth:     &mockAuthClient{},
		Users:    usersMock,
		Sessions: &mockSessionClient{},
	}, authMgr, authzSvc, rv, Config{LoginRateMax: 5, LoginRateWindow: time.Minute})
	if err != nil {
		t.Fatalf("NewAdapter failed: %v", err)
	}

	// Siapkan berbagai konteks pemanggil untuk memverifikasi isolasi autentikasi, RBAC, scope, reauth, dan status user.
	anonCtx := middleware.WithRequestMetadata(context.Background(), middleware.RequestMetadata{ClientIP: "127.0.0.1"})
	custCtx := seedAdapterSession(store, "cust-tok", "cust-1", "customer", authz.StatusActive, now.Unix(), false, false, nil, nil, 1, now)
	noRoleEmpCtx := seedAdapterSession(store, "norole-emp-tok", "emp-norole", "employee", authz.StatusActive, now.Unix(), false, false, nil, nil, 1, now)
	readOnlyCtx := seedAdapterSession(store, "ro-tok", "emp-ro", "employee", authz.StatusActive, now.Unix(), false, false,
		[]authz.RoleGrant{{Code: "iam_readonly"}},
		[]authz.PermissionGrant{
			{Code: "iam.users.read", Scope: authz.ScopeGrant{Type: authz.ScopeGlobal}},
			{Code: "iam.roles.read", Scope: authz.ScopeGrant{Type: authz.ScopeGlobal}},
		}, 1, now)
	scopedWriteCtx := seedAdapterSession(store, "scoped-tok", "emp-scoped", "employee", authz.StatusActive, now.Unix(), false, false,
		[]authz.RoleGrant{{Code: "branch_hr"}},
		[]authz.PermissionGrant{
			{Code: "iam.users.write", IsSensitive: true, Scope: authz.ScopeGrant{Type: "facility", ID: "33333333-3333-4333-8333-333333333333"}},
		}, 1, now)
	staleSessionCtx := seedAdapterSession(store, "stale-tok", "admin-stale", "employee", authz.StatusActive, now.Add(-15*time.Minute).Unix(), false, false,
		[]authz.RoleGrant{{Code: "superadmin"}},
		[]authz.PermissionGrant{
			{Code: "iam.users.write", IsSensitive: true, Scope: authz.ScopeGrant{Type: authz.ScopeGlobal}},
		}, 1, now)
	missingMFACtx := seedAdapterSession(store, "nomfa-tok", "admin-nomfa", "employee", authz.StatusActive, now.Unix(), true, false,
		[]authz.RoleGrant{{Code: "superadmin"}},
		[]authz.PermissionGrant{
			{Code: "iam.users.write", IsSensitive: true, Scope: authz.ScopeGrant{Type: authz.ScopeGlobal}},
		}, 1, now)
	revokedTokenCtx := seedAdapterSession(store, "revoked-tok", "admin-rev", "employee", authz.StatusActive, now.Unix(), false, false,
		[]authz.RoleGrant{{Code: "superadmin"}},
		[]authz.PermissionGrant{
			{Code: "iam.users.write", IsSensitive: true, Scope: authz.ScopeGrant{Type: authz.ScopeGlobal}},
		}, 1, now)
	store.sessions[authsession.HashOpaqueToken("revoked-tok")].Token.Revoked = true

	suspendedSecurityCtx := seedAdapterSession(store, "susp-sec-tok", "emp-susp-sec", "employee", authz.StatusActive, now.Unix(), false, false,
		[]authz.RoleGrant{{Code: "superadmin"}},
		[]authz.PermissionGrant{
			{Code: "iam.users.write", IsSensitive: true, Scope: authz.ScopeGrant{Type: authz.ScopeGlobal}},
		}, 1, now)
	store.sessions[authsession.HashOpaqueToken("susp-sec-tok")].Security.Status = "suspended"

	inactiveSnapshotCtx := seedAdapterSession(store, "susp-snap-tok", "emp-susp-snap", "employee", "suspended", now.Unix(), false, false,
		[]authz.RoleGrant{{Code: "superadmin"}},
		[]authz.PermissionGrant{
			{Code: "iam.users.write", IsSensitive: true, Scope: authz.ScopeGrant{Type: authz.ScopeGlobal}},
		}, 1, now)
	store.sessions[authsession.HashOpaqueToken("susp-snap-tok")].Security.Status = authz.StatusActive
	refreshedByUser["emp-susp-snap"] = &authz.AccessSnapshot{
		PrincipalID:    "emp-susp-snap",
		UserKind:       "employee",
		Status:         "suspended",
		AuthEpoch:      1,
		PermRevision:   1,
		RecoveryEpoch:  1,
		Permissions:    []authz.PermissionGrant{{Code: "iam.users.write", IsSensitive: true, Scope: authz.ScopeGrant{Type: authz.ScopeGlobal}}},
		SyncedAtUnix:   now.Unix(),
		FreshUntilUnix: now.Add(30 * time.Second).Unix(),
	}
	superAdminCtx := seedAdapterSession(store, "sa-tok", "admin-sa", "employee", authz.StatusActive, now.Unix(), true, true,
		[]authz.RoleGrant{{Code: "superadmin"}},
		[]authz.PermissionGrant{
			{Code: "iam.users.write", IsSensitive: true, Scope: authz.ScopeGrant{Type: authz.ScopeGlobal}},
		}, 1, now)
	customRoleCtx := seedAdapterSession(store, "custom-tok", "emp-custom", "employee", authz.StatusActive, now.Unix(), false, false,
		[]authz.RoleGrant{{Code: "hr_provisioner"}},
		[]authz.PermissionGrant{
			{Code: "iam.users.write", IsSensitive: true, Scope: authz.ScopeGrant{Type: authz.ScopeGlobal}},
		}, 1, now)

	deniedCases := []struct {
		name     string
		ctx      context.Context
		wantType model.ErrorTypeEnum
	}{
		{name: "AnonymousUnauthenticated", ctx: anonCtx, wantType: model.ErrorTypeEnumAuthentication},
		{name: "RevokedSessionBlockedAuthentication", ctx: revokedTokenCtx, wantType: model.ErrorTypeEnumAuthentication},
		{name: "AuthenticatedCustomerWithoutRoles", ctx: custCtx, wantType: model.ErrorTypeEnumAuthorization},
		{name: "EmployeeWithoutRoles", ctx: noRoleEmpCtx, wantType: model.ErrorTypeEnumAuthorization},
		{name: "RoleWithoutUsersWritePermission", ctx: readOnlyCtx, wantType: model.ErrorTypeEnumAuthorization},
		{name: "NonGlobalScopedUsersWriteForbidden", ctx: scopedWriteCtx, wantType: model.ErrorTypeEnumAuthorization},
		{name: "StaleSensitiveSessionForbidden", ctx: staleSessionCtx, wantType: model.ErrorTypeEnumAuthorization},
		{name: "MFAEnabledWithoutMFAVerifiedForbidden", ctx: missingMFACtx, wantType: model.ErrorTypeEnumAuthorization},
		{name: "SuspendedSecurityRecordBlockedAuthorization", ctx: suspendedSecurityCtx, wantType: model.ErrorTypeEnumAuthorization},
		{name: "InactiveSnapshotStatusForbidden", ctx: inactiveSnapshotCtx, wantType: model.ErrorTypeEnumAuthorization},
	}

	variants := businessUserVariants()
	for _, tc := range deniedCases {
		t.Run(tc.name, func(t *testing.T) {
			for _, v := range variants {
				before := createCalls
				ok, u, errs, err := v.invoke(adp, tc.ctx, "denied")
				if err != nil {
					t.Fatalf("%s returned unexpected transport error: %v", v.name, err)
				}
				if ok || u != nil || len(errs) != 1 || errs[0].Type != tc.wantType {
					var gotErr model.MutationError
					if len(errs) > 0 && errs[0] != nil {
						gotErr = *errs[0]
					}
					t.Fatalf("%s(%s): got ok=%v user=%+v err0=%+v (len=%d), want error type %s", v.name, tc.name, ok, u, gotErr, len(errs), tc.wantType)
				}
				if createCalls != before {
					t.Fatalf("%s(%s): IAM gRPC CreateUser was invoked %d times unexpectedly", v.name, tc.name, createCalls-before)
				}
			}
		})
	}

	allowedCases := []struct {
		name string
		ctx  context.Context
	}{
		{name: "SuperAdminWithFreshMFASession", ctx: superAdminCtx},
		{name: "CustomRoleWithGlobalUsersWrite", ctx: customRoleCtx},
	}
	for _, tc := range allowedCases {
		t.Run(tc.name, func(t *testing.T) {
			for _, v := range variants {
				before := createCalls
				ok, u, errs, err := v.invoke(adp, tc.ctx, "allowed")
				if err != nil || !ok || len(errs) != 0 || u == nil || u.Kind != v.wantKind {
					t.Fatalf("%s(%s) failed: ok=%v user=%+v errs=%+v err=%v", v.name, tc.name, ok, u, errs, err)
				}
				if createCalls != before+1 {
					t.Fatalf("%s(%s) expected 1 gRPC CreateUser call, got %d", v.name, tc.name, createCalls-before)
				}
			}
		})
	}

	// Uji pencabutan izin real-time: setelah PermRevision naik dan snapshot baru tanpa iam.users.write dimuat, permintaan langsung ditolak.
	t.Run("RevokedPermissionImmediatelyBlockedAfterPermRevisionBump", func(t *testing.T) {
		customHash := authsession.HashOpaqueToken("custom-tok")
		store.sessions[customHash].Security.PermRevision = 2
		refreshedByUser["emp-custom"] = &authz.AccessSnapshot{
			PrincipalID:    "emp-custom",
			UserKind:       "employee",
			Status:         authz.StatusActive,
			AuthEpoch:      1,
			PermRevision:   2,
			RecoveryEpoch:  1,
			Roles:          []authz.RoleGrant{{Code: "hr_provisioner"}},
			Permissions:    []authz.PermissionGrant{},
			SyncedAtUnix:   now.Unix(),
			FreshUntilUnix: now.Add(30 * time.Second).Unix(),
		}
		for _, v := range variants {
			before := createCalls
			ok, u, errs, err := v.invoke(adp, customRoleCtx, "revoked")
			if err != nil {
				t.Fatalf("%s(revoked) unexpected error: %v", v.name, err)
			}
			if ok || u != nil || len(errs) != 1 || errs[0].Type != model.ErrorTypeEnumAuthorization {
				t.Fatalf("%s(revoked): got ok=%v user=%+v errs=%+v, want AUTHORIZATION", v.name, ok, u, errs)
			}
			if createCalls != before {
				t.Fatalf("%s(revoked): IAM gRPC CreateUser was invoked unexpectedly", v.name)
			}
		}
	})
}
