package iam

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/browserauth"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	"github.com/integrione/erp-retail/backend/gateway/internal/source_connections"
	"github.com/integrione/erp-retail/backend/services/human_resource/contract/provisioning"
	iamv1 "github.com/integrione/erp-retail/backend/services/iam/contract/gen/iam/v1"
	"github.com/integrione/erp-retail/backend/shared/authcontext"
	"github.com/integrione/erp-retail/backend/shared/authsession"
	"github.com/integrione/erp-retail/backend/shared/authz"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type mockPartyPortalProvisioner struct {
	prepareFn func(ctx context.Context, partyID string, contactID *string, requiredRole string) (string, *string, error)
	linkFn    func(ctx context.Context, req PartyUserLinkRequest, iamUserID string) (*model.PartyContact, error)
	prepCalls atomic.Int32
	linkCalls atomic.Int32
}

func (m *mockPartyPortalProvisioner) PreparePortalUserProvisioning(ctx context.Context, partyID string, contactID *string, requiredRole string) (string, *string, error) {
	m.prepCalls.Add(1)
	if m.prepareFn != nil {
		return m.prepareFn(ctx, partyID, contactID, requiredRole)
	}
	return partyID, contactID, nil
}

func (m *mockPartyPortalProvisioner) LinkPortalUser(ctx context.Context, req PartyUserLinkRequest, iamUserID string) (*model.PartyContact, error) {
	m.linkCalls.Add(1)
	if m.linkFn != nil {
		return m.linkFn(ctx, req, iamUserID)
	}
	userGID := EncodeGID(ResourceUser, iamUserID)
	return &model.PartyContact{
		ID:                 "gid://integrione/party-contact/cccccccc-cccc-4ccc-8ccc-cccccccccccc",
		RawID:              "cccccccc-cccc-4ccc-8ccc-cccccccccccc",
		PartyID:            "gid://integrione/party/" + req.PartyID,
		IamUserID:          &userGID,
		FullName:           req.FullName,
		ContactRole:        "portal_pic",
		PortalAccessStatus: model.PortalAccessStatusPendingLink,
		IsActive:           true,
	}, nil
}

type mockWorkerPrincipalProvisioner struct {
	prepareFn func(ctx context.Context, workerID string) (string, error)
	linkFn    func(ctx context.Context, workerID, iamUserID string) error
}

func (m *mockWorkerPrincipalProvisioner) PrepareWorkerPrincipalLink(ctx context.Context, workerID string) (string, error) {
	if m.prepareFn != nil {
		return m.prepareFn(ctx, workerID)
	}
	return workerID, nil
}

func (m *mockWorkerPrincipalProvisioner) CompleteWorkerPrincipalLink(ctx context.Context, workerID, iamUserID string) error {
	if m.linkFn != nil {
		return m.linkFn(ctx, workerID, iamUserID)
	}
	return nil
}

func newProvisioningTestContext(t *testing.T, withUsersWrite bool) (*memSessionStore, *middleware.AuthManager, *middleware.Authorizer, *browserauth.RedirectValidator, context.Context) {
	t.Helper()
	now := time.Now()
	store := newMemStore()
	rawTok := "prov-tok"
	tokHash := authsession.HashOpaqueToken(rawTok)
	var perms []authz.PermissionGrant
	if withUsersWrite {
		perms = append(perms, authz.PermissionGrant{Code: "iam.users.write", IsSensitive: true, Scope: authz.ScopeGrant{Type: authz.ScopeGlobal}})
	}
	store.sessions[tokHash] = &authsession.SessionValidationContext{
		Token: &authsession.SessionTokenRecord{
			SessionID: "sess-prov", UserID: "admin-1", TokenHashHex: tokHash, UserKind: "employee",
			AuthEpoch: 1, IssuedAtUnix: now.Unix(), AbsExpiresAtUnix: now.Add(time.Hour).Unix(), MFAAuthenticated: true,
		},
		Security: &authsession.UserSecurityRecord{UserID: "admin-1", Status: authz.StatusActive, AuthEpoch: 1, PermRevision: 1, RecoveryEpoch: 1},
		Snapshot: &authz.AccessSnapshot{
			PrincipalID: "admin-1", UserKind: "employee", Status: authz.StatusActive, MFAEnabled: true,
			AuthEpoch: 1, PermRevision: 1, RecoveryEpoch: 1, Permissions: perms,
			SyncedAtUnix: now.Unix(), FreshUntilUnix: now.Add(30 * time.Second).Unix(),
		},
	}
	authMgr, err := middleware.NewAuthManager(store, middleware.NewSnapshotLoader(fnRefresher{
		fn: func(_ context.Context, _ string, _ authcontext.CallerContext) (*authz.AccessSnapshot, error) {
			return nil, nil
		},
	}, time.Second, 30*time.Second), middleware.SessionConfig{CookieSecure: true})
	if err != nil {
		t.Fatalf("NewAuthManager failed: %v", err)
	}
	authzSvc := middleware.NewAuthorizer(authMgr, nil, 10*time.Minute)
	rv := browserauth.NewRedirectValidator([]string{"/", "/identity"})
	ctx := middleware.WithRequestMetadata(context.Background(), middleware.RequestMetadata{RawCookie: rawTok, CookieHash: tokHash, ClientIP: "127.0.0.1"})
	return store, authMgr, authzSvc, rv, ctx
}

func TestVendorAndSupplierProvisioning_EFVerificationAndCompensationRollback(t *testing.T) {
	now := time.Now()
	_, authMgr, authzSvc, rv, ctx := newProvisioningTestContext(t, true)
	partyUUID := "11111111-1111-4111-8111-111111111111"

	var createCalls, softDeleteCalls atomic.Int32
	var lastSoftDeletedID string
	usersMock := &mockUserClient{
		createFn: func(_ context.Context, in *iamv1.CreateUserRequest) (*iamv1.CreateUserResponse, error) {
			createCalls.Add(1)
			return &iamv1.CreateUserResponse{
				User: &iamv1.User{
					Id: "99999999-9999-4999-8999-999999999999", Kind: in.GetKind(),
					Status: iamv1.UserStatus_USER_STATUS_PENDING, DisplayName: in.GetDisplayName(),
					CreatedAt: timestamppb.New(now), UpdatedAt: timestamppb.New(now),
				},
			}, nil
		},
		softDeleteFn: func(_ context.Context, in *iamv1.SoftDeleteUserRequest) (*iamv1.SoftDeleteUserResponse, error) {
			softDeleteCalls.Add(1)
			lastSoftDeletedID = in.GetUserId()
			return &iamv1.SoftDeleteUserResponse{}, nil
		},
	}

	// 1. Jika partyProvisioner nil (EF tak tersedia) -> fail-closed UNAVAILABLE tanpa memanggil IAM CreateUser.
	adpNoEF, err := NewAdapter(source_connections.IAMClientSet{Users: usersMock}, authMgr, authzSvc, rv, Config{})
	if err != nil {
		t.Fatalf("NewAdapter failed: %v", err)
	}
	resNoEF, err := adpNoEF.CreateVendorUser(ctx, model.CreateVendorUserInput{
		PartyID: partyUUID, FullName: "Vendor A", Email: "va@integrione.test", Username: "va1", Password: "StrongPassword#2026!",
	})
	if err != nil || resNoEF.WasSuccessful || len(resNoEF.Errors) != 1 || resNoEF.Errors[0].Type != model.ErrorTypeEnumUnavailable {
		t.Fatalf("expected UNAVAILABLE when EF provisioner missing, got payload=%+v err=%v", resNoEF, err)
	}
	if createCalls.Load() != 0 {
		t.Fatalf("expected 0 IAM CreateUser calls when EF provisioner unavailable")
	}

	// 2. Jika PreparePortalUserProvisioning gagal (mis. party bukan SUPPLIER -> CONFLICT), IAM CreateUser tidak boleh dipanggil.
	efConflictProv := &mockPartyPortalProvisioner{
		prepareFn: func(context.Context, string, *string, string) (string, *string, error) {
			return "", nil, middleware.ErrConflict("party missing SUPPLIER role", nil)
		},
	}
	adpPreFail, err := NewAdapter(source_connections.IAMClientSet{Users: usersMock}, authMgr, authzSvc, rv, Config{}, WithPartyPortalProvisioner(efConflictProv))
	if err != nil {
		t.Fatalf("NewAdapter failed: %v", err)
	}
	resPreFail, err := adpPreFail.CreateSupplierUser(ctx, model.CreateSupplierUserInput{
		PartyID: partyUUID, FullName: "Sup A", Email: "sa@integrione.test", Username: "sa1", Password: "StrongPassword#2026!",
	})
	if err != nil || resPreFail.WasSuccessful || len(resPreFail.Errors) != 1 || resPreFail.Errors[0].Type != model.ErrorTypeEnumConflict {
		t.Fatalf("expected CONFLICT from EF pre-check, got payload=%+v err=%v", resPreFail, err)
	}
	if createCalls.Load() != 0 || efConflictProv.linkCalls.Load() != 0 {
		t.Fatalf("expected 0 CreateUser and 0 LinkPortalUser on pre-check failure")
	}

	// 3. Jika LinkPortalUser gagal pasca-CreateUser -> kompensasi SoftDeleteUser wajib berjalan agar tidak ada akun yatim.
	efLinkFailProv := &mockPartyPortalProvisioner{
		linkFn: func(context.Context, PartyUserLinkRequest, string) (*model.PartyContact, error) {
			return nil, middleware.ErrUnavailable("enterprise foundation service unavailable", nil)
		},
	}
	adpComp, err := NewAdapter(source_connections.IAMClientSet{Users: usersMock}, authMgr, authzSvc, rv, Config{}, WithPartyPortalProvisioner(efLinkFailProv))
	if err != nil {
		t.Fatalf("NewAdapter failed: %v", err)
	}
	resComp, err := adpComp.CreateVendorUser(ctx, model.CreateVendorUserInput{
		PartyID: partyUUID, FullName: "Vendor Race", Email: "vr@integrione.test", Username: "vr1", Password: "StrongPassword#2026!",
	})
	if err != nil || resComp.WasSuccessful || len(resComp.Errors) != 1 || resComp.Errors[0].Type != model.ErrorTypeEnumUnavailable {
		t.Fatalf("expected UNAVAILABLE after EF link failure, got payload=%+v err=%v", resComp, err)
	}
	if createCalls.Load() != 1 || softDeleteCalls.Load() != 1 || lastSoftDeletedID != "99999999-9999-4999-8999-999999999999" {
		t.Fatalf("expected SoftDeleteUser compensation after link failure: creates=%d deletes=%d id=%q", createCalls.Load(), softDeleteCalls.Load(), lastSoftDeletedID)
	}

	// 4. Jika kompensasi SoftDeleteUser ikut gagal -> tetap return error aman tanpa panic.
	usersMock.softDeleteFn = func(context.Context, *iamv1.SoftDeleteUserRequest) (*iamv1.SoftDeleteUserResponse, error) {
		return nil, errors.New("iam down during rollback")
	}
	resDoubleFail, err := adpComp.CreateSupplierUser(ctx, model.CreateSupplierUserInput{
		PartyID: partyUUID, FullName: "Sup DoubleFail", Email: "sdf@integrione.test", Username: "sdf1", Password: "StrongPassword#2026!",
	})
	if err != nil || resDoubleFail.WasSuccessful || len(resDoubleFail.Errors) != 1 || resDoubleFail.Errors[0].Type != model.ErrorTypeEnumUnavailable {
		t.Fatalf("expected safe UNAVAILABLE when compensation also fails, got payload=%+v err=%v", resDoubleFail, err)
	}
}

func TestCreateEmployeeUser_RequiredWorkerIDAndHRMFailClosed(t *testing.T) {
	now := time.Now()
	_, authMgr, authzSvc, rv, ctx := newProvisioningTestContext(t, true)
	var createCalls, softDeleteCalls atomic.Int32
	usersMock := &mockUserClient{
		createFn: func(_ context.Context, in *iamv1.CreateUserRequest) (*iamv1.CreateUserResponse, error) {
			createCalls.Add(1)
			return &iamv1.CreateUserResponse{
				User: &iamv1.User{
					Id: "77777777-7777-4777-8777-777777777777", Kind: in.GetKind(),
					Status: iamv1.UserStatus_USER_STATUS_PENDING, DisplayName: in.GetDisplayName(),
					CreatedAt: timestamppb.New(now), UpdatedAt: timestamppb.New(now),
				},
			}, nil
		},
		softDeleteFn: func(context.Context, *iamv1.SoftDeleteUserRequest) (*iamv1.SoftDeleteUserResponse, error) {
			softDeleteCalls.Add(1)
			return &iamv1.SoftDeleteUserResponse{}, nil
		},
	}
	adp, err := NewAdapter(source_connections.IAMClientSet{Users: usersMock, EmployeeProvisioning: mockEmployeeProvisioningClient{}}, authMgr, authzSvc, rv, Config{})
	if err != nil {
		t.Fatalf("NewAdapter failed: %v", err)
	}

	// 1. worker_id format tidak valid -> ARGUMENT.
	badWorkerID := "not-a-worker-uuid"
	badRes, err := adp.CreateEmployeeUser(ctx, model.CreateEmployeeUserInput{
		ProvisioningRequestID: fixtureEmployeeProvisioningID, WorkerID: badWorkerID, FullName: "Emp Bad", Email: "eb@integrione.test", Username: "eb1", Password: "StrongPassword#2026!",
	})
	if err != nil || badRes.WasSuccessful || len(badRes.Errors) != 1 || badRes.Errors[0].Type != model.ErrorTypeEnumArgument {
		t.Fatalf("expected ARGUMENT for invalid worker_id, got payload=%+v err=%v", badRes, err)
	}

	// 2. worker_id valid namun adapter HRM belum terpasang -> UNAVAILABLE fail-closed (0 CreateUser).
	validWorkerID := "55555555-5555-4555-8555-555555555555"
	unavailRes, err := adp.CreateEmployeeUser(ctx, model.CreateEmployeeUserInput{
		ProvisioningRequestID: fixtureEmployeeProvisioningID, WorkerID: validWorkerID, FullName: "Emp HRM Missing", Email: "eh@integrione.test", Username: "eh1", Password: "StrongPassword#2026!",
	})
	if err != nil || unavailRes.WasSuccessful || len(unavailRes.Errors) != 1 || unavailRes.Errors[0].Type != model.ErrorTypeEnumUnavailable {
		t.Fatalf("expected UNAVAILABLE when worker_id given without HRM provisioner, got payload=%+v err=%v", unavailRes, err)
	}
	if createCalls.Load() != 0 {
		t.Fatalf("expected 0 CreateUser calls when HRM provisioner unavailable")
	}

	// 3. worker_id valid dengan mock provisioner HRM sukses, gagal, dan hasil penautan tidak pasti.
	hrmFailAdp, _ := NewAdapter(source_connections.IAMClientSet{Users: usersMock, EmployeeProvisioning: mockEmployeeProvisioningClient{}}, authMgr, authzSvc, rv, Config{},
		WithWorkerPrincipalProvisioner(&mockWorkerPrincipalProvisioner{
			linkFn: func(context.Context, string, string) error {
				return middleware.ErrConflict("worker already linked", nil)
			},
		}))
	compRes, err := hrmFailAdp.CreateEmployeeUser(ctx, model.CreateEmployeeUserInput{
		ProvisioningRequestID: fixtureEmployeeProvisioningID, WorkerID: validWorkerID, FullName: "Emp Race", Email: "er@integrione.test", Username: "er1", Password: "StrongPassword#2026!",
	})
	if err != nil || compRes.WasSuccessful || len(compRes.Errors) != 1 || compRes.Errors[0].Type != model.ErrorTypeEnumConflict {
		t.Fatalf("expected CONFLICT on failed worker link, got payload=%+v err=%v", compRes, err)
	}
	if softDeleteCalls.Load() != 0 {
		t.Fatalf("account must remain available for retry after worker link failure")
	}

	unconfirmed, _ := NewAdapter(source_connections.IAMClientSet{Users: usersMock, EmployeeProvisioning: mockEmployeeProvisioningClient{}}, authMgr, authzSvc, rv, Config{},
		WithWorkerPrincipalProvisioner(&mockWorkerPrincipalProvisioner{
			linkFn: func(context.Context, string, string) error {
				return middleware.ErrUnavailable("worker account link could not be confirmed", provisioning.ErrLinkUnconfirmed)
			},
		}))
	result, err := unconfirmed.CreateEmployeeUser(ctx, model.CreateEmployeeUserInput{
		ProvisioningRequestID: fixtureEmployeeProvisioningID, WorkerID: validWorkerID, FullName: "Emp Unconfirmed", Email: "eu@integrione.test", Username: "eu1", Password: "StrongPassword#2026!",
	})
	if err != nil || result.WasSuccessful || len(result.Errors) != 1 || result.Errors[0].Type != model.ErrorTypeEnumUnavailable || softDeleteCalls.Load() != 0 {
		t.Fatalf("unconfirmed link should retain pending account, got payload=%+v err=%v cleanup=%d", result, err, softDeleteCalls.Load())
	}

	hrmOKAdp, _ := NewAdapter(source_connections.IAMClientSet{Users: usersMock, EmployeeProvisioning: mockEmployeeProvisioningClient{}}, authMgr, authzSvc, rv, Config{},
		WithWorkerPrincipalProvisioner(&mockWorkerPrincipalProvisioner{}))
	okRes, err := hrmOKAdp.CreateEmployeeUser(ctx, model.CreateEmployeeUserInput{
		ProvisioningRequestID: fixtureEmployeeProvisioningID, WorkerID: validWorkerID, FullName: "Emp OK", Email: "eok@integrione.test", Username: "eok1", Password: "StrongPassword#2026!",
	})
	if err != nil || !okRes.WasSuccessful || okRes.User == nil {
		t.Fatalf("expected CreateEmployeeUser success with valid HRM provisioner, got payload=%+v err=%v", okRes, err)
	}
}

const fixtureEmployeeProvisioningID = "66666666-6666-4666-8666-666666666666"

type mockEmployeeProvisioningClient struct {
	iamv1.EmployeeProvisioningServiceClient
}

func (mockEmployeeProvisioningClient) GetEmployeeProvisioningRequest(_ context.Context, request *iamv1.GetEmployeeProvisioningRequestRequest, _ ...grpc.CallOption) (*iamv1.GetEmployeeProvisioningRequestResponse, error) {
	return &iamv1.GetEmployeeProvisioningRequestResponse{Item: &iamv1.EmployeeProvisioningRequest{
		Id: request.GetId(), WorkerId: "55555555-5555-4555-8555-555555555555",
		RequestedAt: timestamppb.Now(), ReceivedAt: timestamppb.Now(), Status: "pending",
	}}, nil
}

func (mockEmployeeProvisioningClient) CompleteEmployeeProvisioningRequest(context.Context, *iamv1.CompleteEmployeeProvisioningRequestRequest, ...grpc.CallOption) (*iamv1.CompleteEmployeeProvisioningRequestResponse, error) {
	return &iamv1.CompleteEmployeeProvisioningRequestResponse{}, nil
}
