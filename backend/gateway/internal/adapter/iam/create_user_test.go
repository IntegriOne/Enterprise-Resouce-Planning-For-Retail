package iam

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/browserauth"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	"github.com/integrione/erp-retail/backend/gateway/internal/source_connections"
	iamv1 "github.com/integrione/erp-retail/backend/services/iam/contract/gen/iam/v1"
	"github.com/integrione/erp-retail/backend/shared/authsession"
	"github.com/integrione/erp-retail/backend/shared/authz"
)

func buildCreateUserConflictError(field, message string) error {
	st, _ := status.New(codes.AlreadyExists, message).WithDetails(
		&errdetails.ErrorInfo{
			Reason: "iam.identity.already_exists",
			Domain: "iam",
		},
		&errdetails.BadRequest{
			FieldViolations: []*errdetails.BadRequest_FieldViolation{
				{
					Field:       field,
					Description: message,
				},
			},
		},
	)
	return st.Err()
}

// TestCreateUser_PublicAndAuthenticatedBehavior menguji akses publik createUser, proteksi email duplikat via redirect /login,
// pengungkapan username terpakai, serta perbedaan visibilitas CreateUserPayload.user bagi pemanggil anonim dan terautentikasi.
func TestCreateUser_PublicAndAuthenticatedBehavior(t *testing.T) {
	now := time.Now()
	store := newMemStore()

	custUUID := "55555555-5555-4555-8555-555555555555"
	custTok := "cust-session-tok"
	custHash := authsession.HashOpaqueToken(custTok)
	store.sessions[custHash] = &authsession.SessionValidationContext{
		Token: &authsession.SessionTokenRecord{
			SessionID:    "sess-cust",
			UserID:       custUUID,
			TokenHashHex: custHash,
			UserKind:     "customer",
			AuthEpoch:    1,
			IssuedAtUnix: now.Unix(),
		},
		Security: &authsession.UserSecurityRecord{UserID: custUUID, Status: authz.StatusActive, AuthEpoch: 1, PermRevision: 1, RecoveryEpoch: 1},
		Snapshot: &authz.AccessSnapshot{
			PrincipalID:    custUUID,
			UserKind:       "customer",
			Status:         authz.StatusActive,
			AuthEpoch:      1,
			PermRevision:   1,
			RecoveryEpoch:  1,
			SyncedAtUnix:   now.Unix(),
			FreshUntilUnix: now.Add(30 * time.Second).Unix(),
		},
	}

	adminUUID := "66666666-6666-4666-8666-666666666666"
	adminTok := "admin-session-tok"
	adminHash := authsession.HashOpaqueToken(adminTok)
	store.sessions[adminHash] = &authsession.SessionValidationContext{
		Token: &authsession.SessionTokenRecord{
			SessionID:    "sess-admin",
			UserID:       adminUUID,
			TokenHashHex: adminHash,
			UserKind:     "employee",
			AuthEpoch:    1,
			IssuedAtUnix: now.Unix(),
		},
		Security: &authsession.UserSecurityRecord{UserID: adminUUID, Status: authz.StatusActive, AuthEpoch: 1, PermRevision: 1, RecoveryEpoch: 1},
		Snapshot: &authz.AccessSnapshot{
			PrincipalID:    adminUUID,
			UserKind:       "employee",
			Status:         authz.StatusActive,
			AuthEpoch:      1,
			PermRevision:   1,
			RecoveryEpoch:  1,
			Permissions:    []authz.PermissionGrant{{Code: "iam.users.write", Scope: authz.ScopeGrant{Type: authz.ScopeGlobal}}},
			SyncedAtUnix:   now.Unix(),
			FreshUntilUnix: now.Add(30 * time.Second).Unix(),
		},
	}

	usersMock := &mockUserClient{
		createFn: func(_ context.Context, in *iamv1.CreateUserRequest) (*iamv1.CreateUserResponse, error) {
			if strings.EqualFold(in.GetUsername(), "taken_user") {
				return nil, buildCreateUserConflictError("username", "username has been taken")
			}
			if strings.EqualFold(in.GetEmail(), "taken@integrione.id") {
				return nil, buildCreateUserConflictError("email", "email is already registered")
			}
			return &iamv1.CreateUserResponse{
				User: &iamv1.User{
					Id:          "77777777-7777-4777-8777-777777777777",
					Kind:        in.GetKind(),
					Status:      iamv1.UserStatus_USER_STATUS_PENDING,
					DisplayName: in.GetDisplayName(),
					CreatedAt:   timestamppb.New(now),
					UpdatedAt:   timestamppb.New(now),
				},
			}, nil
		},
	}

	authMgr, err := middleware.NewAuthManager(store, middleware.NewSnapshotLoader(dummyRefresher{}, time.Second, 30*time.Second), middleware.SessionConfig{CookieSecure: true})
	if err != nil {
		t.Fatalf("NewAuthManager error = %v", err)
	}
	authzSvc := middleware.NewAuthorizer(authMgr, nil, 10*time.Minute)
	rv := browserauth.NewRedirectValidator([]string{"/", "/login", "/identity"})
	adp, err := NewAdapter(source_connections.IAMClientSet{Users: usersMock}, authMgr, authzSvc, rv, Config{LoginRateMax: 10, LoginRateWindow: time.Minute})
	if err != nil {
		t.Fatalf("NewAdapter error = %v", err)
	}

	anonCtx := middleware.WithRequestMetadata(context.Background(), middleware.RequestMetadata{ClientIP: "127.0.0.1", UserAgent: "ua"})

	// 1. Pemanggil anonim berhasil mendaftar tanpa izin iam.users.write, tetapi User wajib nil.
	pubOK, err := adp.CreateUser(anonCtx, model.CreateUserInput{
		FullName: "Public Buyer",
		Email:    "fresh@integrione.id",
		Username: "freshbuyer",
		Password: "StrongPassword#2026!",
	})
	if err != nil || !pubOK.WasSuccessful || len(pubOK.Errors) != 0 || pubOK.User != nil || pubOK.RedirectLocation != nil {
		t.Fatalf("expected public CreateUser success with nil User, got payload=%+v err=%v", pubOK, err)
	}

	// 2. Email sudah terdaftar pada jalur publik tidak boleh membocorkan error email; wajib mengarah ke /login + header Location.
	eff := browserauth.NewResponseEffects()
	dupEmailCtx := browserauth.WithResponseEffects(anonCtx, eff)
	dupEmailRes, err := adp.CreateUser(dupEmailCtx, model.CreateUserInput{
		FullName: "Duplicate Email Buyer",
		Email:    "taken@integrione.id",
		Username: "freeuser99",
		Password: "StrongPassword#2026!",
	})
	rec := httptest.NewRecorder()
	eff.ApplyTo(rec)
	if err != nil || !dupEmailRes.WasSuccessful || len(dupEmailRes.Errors) != 0 || dupEmailRes.User != nil || dupEmailRes.RedirectLocation == nil || *dupEmailRes.RedirectLocation != "/login" {
		t.Fatalf("expected safe redirect /login payload for duplicate email, got payload=%+v err=%v", dupEmailRes, err)
	}
	if gotLoc := rec.Header().Get(browserauth.LocationHeaderName); gotLoc != "/login" {
		t.Fatalf("expected HTTP Location header '/login', got %q", gotLoc)
	}

	// 3. Username sudah dipakai pada jalur publik wajib mengembalikan CONFLICT "username has been taken" dengan user nil.
	dupUnameRes, err := adp.CreateUser(anonCtx, model.CreateUserInput{
		FullName: "Duplicate Username Buyer",
		Email:    "fresh2@integrione.id",
		Username: "taken_user",
		Password: "StrongPassword#2026!",
	})
	if err != nil || dupUnameRes.WasSuccessful || dupUnameRes.User != nil || len(dupUnameRes.Errors) != 1 || dupUnameRes.Errors[0].Type != model.ErrorTypeEnumConflict || dupUnameRes.Errors[0].Message != "username has been taken" {
		t.Fatalf("expected CONFLICT 'username has been taken', got payload=%+v err=%v", dupUnameRes, err)
	}

	// 4. Input kosong wajib ditolak ARGUMENT dengan user nil.
	emptyRes, err := adp.CreateUser(anonCtx, model.CreateUserInput{FullName: "   ", Email: "a@b.c", Username: "u1", Password: "p"})
	if err != nil || emptyRes.WasSuccessful || emptyRes.User != nil || len(emptyRes.Errors) != 1 || emptyRes.Errors[0].Type != model.ErrorTypeEnumArgument {
		t.Fatalf("expected ARGUMENT on empty fields, got payload=%+v err=%v", emptyRes, err)
	}

	// 5. Pengguna customer yang sudah login menerima User non-nil saat sukses, dan tetap dialihkan aman ke /login bila email duplikat.
	custCtx := middleware.WithRequestMetadata(context.Background(), middleware.RequestMetadata{RawCookie: custTok, CookieHash: custHash, ClientIP: "127.0.0.2"})
	custOK, err := adp.CreateUser(custCtx, model.CreateUserInput{
		FullName: "Customer Created",
		Email:    "custnew@integrione.id",
		Username: "custnew01",
		Password: "StrongPassword#2026!",
	})
	if err != nil || !custOK.WasSuccessful || custOK.User == nil || custOK.User.Kind != model.UserKindCustomer {
		t.Fatalf("expected non-nil User for authenticated customer CreateUser, got payload=%+v err=%v", custOK, err)
	}
	custDupEmail, err := adp.CreateUser(custCtx, model.CreateUserInput{
		FullName: "Customer Dup Email",
		Email:    "taken@integrione.id",
		Username: "custnew02",
		Password: "StrongPassword#2026!",
	})
	if err != nil || !custDupEmail.WasSuccessful || custDupEmail.User != nil || custDupEmail.RedirectLocation == nil || *custDupEmail.RedirectLocation != "/login" {
		t.Fatalf("expected non-admin authenticated caller to still receive /login redirect on duplicate email, got payload=%+v err=%v", custDupEmail, err)
	}

	// 6. Admin berizin iam.users.write menerima User saat sukses serta error CONFLICT saat email sudah terdaftar.
	adminCtx := middleware.WithRequestMetadata(context.Background(), middleware.RequestMetadata{RawCookie: adminTok, CookieHash: adminHash, ClientIP: "127.0.0.3"})
	adminDupEmail, err := adp.CreateUser(adminCtx, model.CreateUserInput{
		FullName: "Admin Dup Email",
		Email:    "taken@integrione.id",
		Username: "adminnew01",
		Password: "StrongPassword#2026!",
	})
	if err != nil || adminDupEmail.WasSuccessful || adminDupEmail.User != nil || len(adminDupEmail.Errors) != 1 || adminDupEmail.Errors[0].Type != model.ErrorTypeEnumConflict {
		t.Fatalf("expected CONFLICT for admin duplicate email, got payload=%+v err=%v", adminDupEmail, err)
	}

	// 7. Pembatasan laju melindungi registrasi publik terhadap spam berulang.
	rlCtx := middleware.WithRequestMetadata(context.Background(), middleware.RequestMetadata{ClientIP: "198.51.100.77"})
	for i := 0; i < 10; i++ {
		_, _ = adp.CreateUser(rlCtx, model.CreateUserInput{
			FullName: "Rate Test",
			Email:    "rl-user@integrione.id",
			Username: "rluser01",
			Password: "StrongPassword#2026!",
		})
	}
	throttled, err := adp.CreateUser(rlCtx, model.CreateUserInput{
		FullName: "Rate Test",
		Email:    "rl-user@integrione.id",
		Username: "rluser02",
		Password: "StrongPassword#2026!",
	})
	if err != nil || throttled.WasSuccessful || throttled.User != nil || len(throttled.Errors) != 1 || throttled.Errors[0].Type != model.ErrorTypeEnumThrottle {
		t.Fatalf("expected THROTTLE after exceeding public CreateUser rate limit, got payload=%+v err=%v", throttled, err)
	}
}
