package handler_test

import (
	"context"
	"net/url"
	"testing"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"

	iamv1 "github.com/integrione/erp-retail/backend/services/iam/contract/gen/iam/v1"
	"github.com/integrione/erp-retail/backend/services/iam/pkg/totp"
)

func TestPermissionRoleGroupRemainingRPCs_Bufconn(t *testing.T) {
	t.Parallel()

	env := startBufconnServer(t)
	ctx := t.Context()

	permClient := iamv1.NewPermissionServiceClient(env.conn)
	roleClient := iamv1.NewRoleServiceClient(env.conn)
	groupClient := iamv1.NewGroupServiceClient(env.conn)
	userClient := iamv1.NewUserServiceClient(env.conn)

	// Siapkan permission, role, group, dan user untuk menguji seluruh siklus CRUD & pencabutan RBAC.
	permResp, err := permClient.CreatePermission(ctx, &iamv1.CreatePermissionRequest{
		Code:        "catalog.item.edit",
		Resource:    "catalog.item",
		Action:      "edit",
		Description: "Edit catalog items",
	})
	if err != nil {
		t.Fatalf("CreatePermission error = %v", err)
	}
	permID := permResp.GetPermission().GetId()

	roleResp, err := roleClient.CreateRole(ctx, &iamv1.CreateRoleRequest{
		Code:                "catalog_editor",
		Name:                "Catalog Editor",
		Description:         "Initial description",
		ManagementDomain:    "plm",
		DefaultRedirectPath: "/product",
		RedirectPriority:    30,
	})
	if err != nil {
		t.Fatalf("CreateRole error = %v", err)
	}
	roleID := roleResp.GetRole().GetId()

	if _, err := roleClient.AssignRolePermission(ctx, &iamv1.AssignRolePermissionRequest{
		RoleId:       roleID,
		PermissionId: permID,
	}); err != nil {
		t.Fatalf("AssignRolePermission error = %v", err)
	}

	gotRoleCode, err := roleClient.GetRoleByCode(ctx, &iamv1.GetRoleByCodeRequest{Code: "catalog_editor"})
	if err != nil || gotRoleCode.GetRole().GetId() != roleID {
		t.Fatalf("GetRoleByCode failed: resp=%+v err=%v", gotRoleCode, err)
	}

	listRolesResp, err := roleClient.ListRoles(ctx, &iamv1.ListRolesRequest{
		Query:            "catalog",
		ManagementDomain: "plm",
		ActiveOnly:       true,
		Page:             &iamv1.PageRequest{Limit: 10, Offset: 0},
	})
	if err != nil || len(listRolesResp.GetItems()) != 1 {
		t.Fatalf("ListRoles failed: resp=%+v err=%v", listRolesResp, err)
	}

	updRole, err := roleClient.UpdateRoleDetails(ctx, &iamv1.UpdateRoleDetailsRequest{
		Id:                  roleID,
		Name:                "Senior Catalog Editor",
		Description:         "Updated description",
		ManagementDomain:    "plm",
		DefaultRedirectPath: "/product/items",
		RedirectPriority:    25,
	})
	if err != nil || updRole.GetRole().GetName() != "Senior Catalog Editor" {
		t.Fatalf("UpdateRoleDetails failed: resp=%+v err=%v", updRole, err)
	}

	deactRole, err := roleClient.DeactivateRole(ctx, &iamv1.DeactivateRoleRequest{Id: roleID})
	if err != nil || deactRole.GetRole().GetIsActive() {
		t.Fatalf("DeactivateRole failed: resp=%+v err=%v", deactRole, err)
	}
	actRole, err := roleClient.ActivateRole(ctx, &iamv1.ActivateRoleRequest{Id: roleID})
	if err != nil || !actRole.GetRole().GetIsActive() {
		t.Fatalf("ActivateRole failed: resp=%+v err=%v", actRole, err)
	}

	uResp, err := userClient.CreateUser(ctx, &iamv1.CreateUserRequest{
		Kind:                   iamv1.UserKind_USER_KIND_EMPLOYEE,
		Status:                 iamv1.UserStatus_USER_STATUS_ACTIVE,
		DisplayName:            "Sari Editor",
		Email:                  "sari.editor@integrione.local",
		Username:               "sari.editor",
		Password:               "VaultPass#2026!Q",
		MarkIdentitiesVerified: true,
	})
	if err != nil {
		t.Fatalf("CreateUser error = %v", err)
	}
	uid := uResp.GetUser().GetId()

	urResp, err := roleClient.AssignUserRole(ctx, &iamv1.AssignUserRoleRequest{
		UserId:    uid,
		RoleId:    roleID,
		ScopeType: iamv1.ScopeType_SCOPE_TYPE_GLOBAL,
	})
	if err != nil {
		t.Fatalf("AssignUserRole error = %v", err)
	}
	userRolesResp, err := roleClient.ListUserRoles(ctx, &iamv1.ListUserRolesRequest{UserId: uid, IncludeExpired: true})
	if err != nil || len(userRolesResp.GetUserRoles()) != 1 {
		t.Fatalf("ListUserRoles failed: resp=%+v err=%v", userRolesResp, err)
	}
	if _, err := roleClient.RevokeUserRole(ctx, &iamv1.RevokeUserRoleRequest{AssignmentId: urResp.GetUserRole().GetId()}); err != nil {
		t.Fatalf("RevokeUserRole error = %v", err)
	}

	// Uji seluruh RPC GroupService termasuk get/list/update/activate/deactivate/revoke/remove/delete.
	grpResp, err := groupClient.CreateGroup(ctx, &iamv1.CreateGroupRequest{
		Code:        "plm_editors",
		Name:        "PLM Editors",
		Description: "Initial group",
	})
	if err != nil {
		t.Fatalf("CreateGroup error = %v", err)
	}
	gid := grpResp.GetGroup().GetId()

	if gotByID, err := groupClient.GetGroupById(ctx, &iamv1.GetGroupByIdRequest{Id: gid}); err != nil || gotByID.GetGroup().GetCode() != "plm_editors" {
		t.Fatalf("GetGroupById failed: resp=%+v err=%v", gotByID, err)
	}
	if gotByCode, err := groupClient.GetGroupByCode(ctx, &iamv1.GetGroupByCodeRequest{Code: "plm_editors"}); err != nil || gotByCode.GetGroup().GetId() != gid {
		t.Fatalf("GetGroupByCode failed: resp=%+v err=%v", gotByCode, err)
	}
	if listGrps, err := groupClient.ListGroups(ctx, &iamv1.ListGroupsRequest{Query: "plm", ActiveOnly: true}); err != nil || len(listGrps.GetItems()) != 1 {
		t.Fatalf("ListGroups failed: resp=%+v err=%v", listGrps, err)
	}
	if updGrp, err := groupClient.UpdateGroupDetails(ctx, &iamv1.UpdateGroupDetailsRequest{
		Id:          gid,
		Name:        "PLM Senior Editors",
		Description: "Updated group",
	}); err != nil || updGrp.GetGroup().GetName() != "PLM Senior Editors" {
		t.Fatalf("UpdateGroupDetails failed: resp=%+v err=%v", updGrp, err)
	}
	if deactGrp, err := groupClient.DeactivateGroup(ctx, &iamv1.DeactivateGroupRequest{Id: gid}); err != nil || deactGrp.GetGroup().GetIsActive() {
		t.Fatalf("DeactivateGroup failed: resp=%+v err=%v", deactGrp, err)
	}
	if actGrp, err := groupClient.ActivateGroup(ctx, &iamv1.ActivateGroupRequest{Id: gid}); err != nil || !actGrp.GetGroup().GetIsActive() {
		t.Fatalf("ActivateGroup failed: resp=%+v err=%v", actGrp, err)
	}
	if _, err := groupClient.AddGroupMember(ctx, &iamv1.AddGroupMemberRequest{GroupId: gid, UserId: uid}); err != nil {
		t.Fatalf("AddGroupMember error = %v", err)
	}
	if uGrps, err := groupClient.ListUserGroups(ctx, &iamv1.ListUserGroupsRequest{UserId: uid}); err != nil || len(uGrps.GetGroups()) != 1 {
		t.Fatalf("ListUserGroups failed: resp=%+v err=%v", uGrps, err)
	}
	grResp, err := groupClient.AssignGroupRole(ctx, &iamv1.AssignGroupRoleRequest{
		GroupId:   gid,
		RoleId:    roleID,
		ScopeType: iamv1.ScopeType_SCOPE_TYPE_GLOBAL,
	})
	if err != nil {
		t.Fatalf("AssignGroupRole error = %v", err)
	}
	if gRoles, err := groupClient.ListGroupRoles(ctx, &iamv1.ListGroupRolesRequest{GroupId: gid, IncludeExpired: true}); err != nil || len(gRoles.GetGroupRoles()) != 1 {
		t.Fatalf("ListGroupRoles failed: resp=%+v err=%v", gRoles, err)
	}
	if _, err := groupClient.RevokeGroupRole(ctx, &iamv1.RevokeGroupRoleRequest{AssignmentId: grResp.GetGroupRole().GetId()}); err != nil {
		t.Fatalf("RevokeGroupRole error = %v", err)
	}
	if _, err := groupClient.RemoveGroupMember(ctx, &iamv1.RemoveGroupMemberRequest{GroupId: gid, UserId: uid}); err != nil {
		t.Fatalf("RemoveGroupMember error = %v", err)
	}
	if _, err := groupClient.DeleteGroup(ctx, &iamv1.DeleteGroupRequest{Id: gid}); err != nil {
		t.Fatalf("DeleteGroup error = %v", err)
	}

	// Bersihkan relasi izin -> hapus role -> hapus permission.
	if _, err := roleClient.RevokeRolePermission(ctx, &iamv1.RevokeRolePermissionRequest{
		RoleId:       roleID,
		PermissionId: permID,
	}); err != nil {
		t.Fatalf("RevokeRolePermission error = %v", err)
	}
	if _, err := roleClient.DeleteRole(ctx, &iamv1.DeleteRoleRequest{Id: roleID}); err != nil {
		t.Fatalf("DeleteRole error = %v", err)
	}
	if _, err := permClient.DeletePermission(ctx, &iamv1.DeletePermissionRequest{Id: permID}); err != nil {
		t.Fatalf("DeletePermission error = %v", err)
	}
}

func TestUserSessionAuthRemainingRPCs_Bufconn(t *testing.T) {
	t.Parallel()

	env := startBufconnServer(t)
	ctx := t.Context()

	userClient := iamv1.NewUserServiceClient(env.conn)
	sessClient := iamv1.NewSessionServiceClient(env.conn)
	authClient := iamv1.NewAuthServiceClient(env.conn)
	roleClient := iamv1.NewRoleServiceClient(env.conn)

	// Siapkan role agar ResolvePostLoginRedirect dapat dievaluasi deterministik.
	roleResp, err := roleClient.CreateRole(ctx, &iamv1.CreateRoleRequest{
		Code:                "wms_operator",
		Name:                "WMS Operator",
		ManagementDomain:    "wms",
		DefaultRedirectPath: "/warehouse",
		RedirectPriority:    15,
	})
	if err != nil {
		t.Fatalf("CreateRole error = %v", err)
	}

	created, err := userClient.CreateUser(ctx, &iamv1.CreateUserRequest{
		Kind:                   iamv1.UserKind_USER_KIND_EMPLOYEE,
		Status:                 iamv1.UserStatus_USER_STATUS_ACTIVE,
		DisplayName:            "Andi Gudang",
		Email:                  "andi.gudang@integrione.local",
		Username:               "andi.gudang",
		Password:               "SteelKey#2026!W",
		MarkIdentitiesVerified: true,
	})
	if err != nil {
		t.Fatalf("CreateUser error = %v", err)
	}
	uid := created.GetUser().GetId()

	if _, err := roleClient.AssignUserRole(ctx, &iamv1.AssignUserRoleRequest{
		UserId:    uid,
		RoleId:    roleResp.GetRole().GetId(),
		ScopeType: iamv1.ScopeType_SCOPE_TYPE_GLOBAL,
	}); err != nil {
		t.Fatalf("AssignUserRole error = %v", err)
	}

	// Uji UserService query, profil, siklus hidup status, dan manajemen identitas sekunder.
	if gotU, err := userClient.GetUserById(ctx, &iamv1.GetUserByIdRequest{UserId: uid}); err != nil || gotU.GetUser().GetDisplayName() != "Andi Gudang" {
		t.Fatalf("GetUserById failed: resp=%+v err=%v", gotU, err)
	}
	if snap, err := userClient.GetUserAccessSnapshot(ctx, &iamv1.GetUserAccessSnapshotRequest{UserId: uid}); err != nil || snap.GetSnapshot().GetUserId() != uid {
		t.Fatalf("GetUserAccessSnapshot failed: resp=%+v err=%v", snap, err)
	}
	if listed, err := userClient.ListUsers(ctx, &iamv1.ListUsersRequest{Query: "Andi"}); err != nil || len(listed.GetItems()) != 1 {
		t.Fatalf("ListUsers failed: resp=%+v err=%v", listed, err)
	}
	if updU, err := userClient.UpdateUserProfile(ctx, &iamv1.UpdateUserProfileRequest{
		UserId:      uid,
		DisplayName: "Andi Warehouse",
		Locale:      "id-ID",
		Timezone:    "Asia/Jakarta",
	}); err != nil || updU.GetUser().GetDisplayName() != "Andi Warehouse" {
		t.Fatalf("UpdateUserProfile failed: resp=%+v err=%v", updU, err)
	}
	if susp, err := userClient.SuspendUser(ctx, &iamv1.SuspendUserRequest{UserId: uid}); err != nil || susp.GetUser().GetStatus() != iamv1.UserStatus_USER_STATUS_SUSPENDED {
		t.Fatalf("SuspendUser failed: resp=%+v err=%v", susp, err)
	}
	if act, err := userClient.ActivateUser(ctx, &iamv1.ActivateUserRequest{UserId: uid}); err != nil || act.GetUser().GetStatus() != iamv1.UserStatus_USER_STATUS_ACTIVE {
		t.Fatalf("ActivateUser failed: resp=%+v err=%v", act, err)
	}
	if dis, err := userClient.DisableUser(ctx, &iamv1.DisableUserRequest{UserId: uid}); err != nil || dis.GetUser().GetStatus() != iamv1.UserStatus_USER_STATUS_DISABLED {
		t.Fatalf("DisableUser failed: resp=%+v err=%v", dis, err)
	}
	if _, err := userClient.ActivateUser(ctx, &iamv1.ActivateUserRequest{UserId: uid}); err != nil {
		t.Fatalf("re-ActivateUser failed: %v", err)
	}

	// Batas maksimal 1 email dan 1 username per pengguna menolak penambahan duplikat per jenis.
	if _, err := userClient.AddUserIdentity(ctx, &iamv1.AddUserIdentityRequest{
		UserId:    uid,
		Kind:      iamv1.IdentityKind_IDENTITY_KIND_EMAIL,
		Value:     "andi.alt@integrione.local",
		IsPrimary: false,
	}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("expected FailedPrecondition for 2nd email identity, got %s (%v)", status.Code(err), err)
	}
	if _, err := userClient.AddUserIdentity(ctx, &iamv1.AddUserIdentityRequest{
		UserId:    uid,
		Kind:      iamv1.IdentityKind_IDENTITY_KIND_USERNAME,
		Value:     "andi.ops2",
		IsPrimary: false,
	}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("expected FailedPrecondition for 2nd username identity, got %s (%v)", status.Code(err), err)
	}

	// Ganti username uid dari andi.gudang -> andi.gudang2 agar andi.gudang bebas dipakai pengguna lain.
	updUsernameResp, err := userClient.UpdateUserUsername(ctx, &iamv1.UpdateUserUsernameRequest{
		UserId:   uid,
		Username: "andi.gudang2",
	})
	if err != nil || updUsernameResp.GetUsername() != "andi.gudang2" || !updUsernameResp.GetIdentity().GetIsVerified() {
		t.Fatalf("UpdateUserUsername failed: resp=%+v err=%v", updUsernameResp, err)
	}

	// Buat akun email-only pending untuk menguji Request/ConfirmIdentityVerification, AddUserIdentity(username bekas), SetPrimaryUserIdentity, dan RemoveUserIdentity.
	emailOnlyCreated, err := userClient.CreateUser(ctx, &iamv1.CreateUserRequest{
		Kind:                   iamv1.UserKind_USER_KIND_CUSTOMER,
		Status:                 iamv1.UserStatus_USER_STATUS_PENDING,
		DisplayName:            "Bima Pembeli",
		Email:                  "bima.pembeli@integrione.local",
		Password:               "SteelKey#2026!Z",
		MarkIdentitiesVerified: false,
	})
	if err != nil {
		t.Fatalf("CreateUser(email-only) error = %v", err)
	}
	emailOnlyUID := emailOnlyCreated.GetUser().GetId()
	emailOnlyIdents, err := userClient.ListUserIdentities(ctx, &iamv1.ListUserIdentitiesRequest{UserId: emailOnlyUID})
	if err != nil || len(emailOnlyIdents.GetIdentities()) != 1 {
		t.Fatalf("ListUserIdentities(email-only) failed: resp=%+v err=%v", emailOnlyIdents, err)
	}
	emailID := emailOnlyIdents.GetIdentities()[0].GetId()
	verTok, err := userClient.RequestIdentityVerification(ctx, &iamv1.RequestIdentityVerificationRequest{IdentityId: emailID})
	if err != nil || verTok.GetIssuedToken().GetPlaintextToken() == "" {
		t.Fatalf("RequestIdentityVerification failed: resp=%+v err=%v", verTok, err)
	}
	confID, err := userClient.ConfirmIdentityVerification(ctx, &iamv1.ConfirmIdentityVerificationRequest{
		IdentityId:          emailID,
		RawToken:            verTok.GetIssuedToken().GetPlaintextToken(),
		ActivatePendingUser: true,
	})
	if err != nil || !confID.GetIdentity().GetIsVerified() {
		t.Fatalf("ConfirmIdentityVerification failed: resp=%+v err=%v", confID, err)
	}

	// Tambahkan username bekas yang sudah dilepas oleh uid -> wajib berhasil dan bisa dijadikan primary sementara lalu dihapus.
	reusedIdent, err := userClient.AddUserIdentity(ctx, &iamv1.AddUserIdentityRequest{
		UserId:    emailOnlyUID,
		Kind:      iamv1.IdentityKind_IDENTITY_KIND_USERNAME,
		Value:     "andi.gudang",
		IsPrimary: false,
	})
	if err != nil || !reusedIdent.GetIdentity().GetIsVerified() {
		t.Fatalf("AddUserIdentity(reused username) failed: resp=%+v err=%v", reusedIdent, err)
	}
	if prim, err := userClient.SetPrimaryUserIdentity(ctx, &iamv1.SetPrimaryUserIdentityRequest{UserId: emailOnlyUID, IdentityId: reusedIdent.GetIdentity().GetId()}); err != nil || !prim.GetIdentity().GetIsPrimary() {
		t.Fatalf("SetPrimaryUserIdentity(username) failed: resp=%+v err=%v", prim, err)
	}
	if prim, err := userClient.SetPrimaryUserIdentity(ctx, &iamv1.SetPrimaryUserIdentityRequest{UserId: emailOnlyUID, IdentityId: emailID}); err != nil || !prim.GetIdentity().GetIsPrimary() {
		t.Fatalf("SetPrimaryUserIdentity(email) failed: resp=%+v err=%v", prim, err)
	}
	if _, err := userClient.RemoveUserIdentity(ctx, &iamv1.RemoveUserIdentityRequest{
		UserId:     emailOnlyUID,
		IdentityId: reusedIdent.GetIdentity().GetId(),
	}); err != nil {
		t.Fatalf("RemoveUserIdentity error = %v", err)
	}

	// Uji SessionService RPCs: IssueSession, VerifyCsrf, MarkSessionMfaVerified, ResolvePostLoginRedirect.
	issued, err := sessClient.IssueSession(ctx, &iamv1.IssueSessionRequest{
		UserId:      uid,
		MfaVerified: false,
		IpAddress:   "127.0.0.1",
		UserAgent:   "unit-test/1.0",
	})
	if err != nil || issued.GetIssuedSession().GetSessionToken() == "" {
		t.Fatalf("IssueSession failed: resp=%+v err=%v", issued, err)
	}
	sid1 := issued.GetIssuedSession().GetSession().GetId()
	if _, err := sessClient.VerifyCsrf(ctx, &iamv1.VerifyCsrfRequest{
		SessionId: sid1,
		CsrfToken: issued.GetIssuedSession().GetCsrfToken(),
	}); err != nil {
		t.Fatalf("VerifyCsrf error = %v", err)
	}
	marked, err := sessClient.MarkSessionMfaVerified(ctx, &iamv1.MarkSessionMfaVerifiedRequest{SessionId: sid1})
	if err != nil || !marked.GetSession().GetMfaVerified() || marked.GetCsrfToken() == "" {
		t.Fatalf("MarkSessionMfaVerified failed: resp=%+v err=%v", marked, err)
	}
	redir, err := sessClient.ResolvePostLoginRedirect(ctx, &iamv1.ResolvePostLoginRedirectRequest{UserId: uid})
	if err != nil || redir.GetRedirectPath() != "/warehouse" {
		t.Fatalf("ResolvePostLoginRedirect failed: resp=%+v err=%v", redir, err)
	}

	// Enroll + Confirm TOTP pada step -1 agar VerifyMfaChallenge (step 0) dan VerifySessionMfa (step +1) sukses tanpa replay.
	baseNow := time.Now()
	totpSec := enrollAndConfirmTOTPAt(t, ctx, authClient, env.totpProv, uid, "andi.gudang@integrione.local", baseNow.Add(-30*time.Second))

	if cntResp, err := authClient.CountUnusedRecoveryCodes(ctx, &iamv1.CountUnusedRecoveryCodesRequest{UserId: uid}); err != nil || cntResp.GetCount() == 0 {
		t.Fatalf("CountUnusedRecoveryCodes failed: resp=%+v err=%v", cntResp, err)
	}
	if regenResp, err := authClient.RegenerateRecoveryCodes(ctx, &iamv1.RegenerateRecoveryCodesRequest{UserId: uid}); err != nil || len(regenResp.GetRecoveryCodes()) == 0 {
		t.Fatalf("RegenerateRecoveryCodes failed: resp=%+v err=%v", regenResp, err)
	}

	// Username lama "andi.gudang" harus sudah tidak bisa dipakai login oleh uid setelah diubah ke "andi.gudang2".
	if _, err := authClient.AuthenticatePassword(ctx, &iamv1.AuthenticatePasswordRequest{
		Identifier: "andi.gudang",
		Password:   "SteelKey#2026!W",
		IpAddress:  "127.0.0.1",
		UserAgent:  "browser/1.0",
	}); status.Code(err) != codes.Unauthenticated {
		t.Fatalf("expected Unauthenticated for old username 'andi.gudang', got %s (%v)", status.Code(err), err)
	}

	mfaLogin, err := authClient.AuthenticatePassword(ctx, &iamv1.AuthenticatePasswordRequest{
		Identifier: "andi.gudang2",
		Password:   "SteelKey#2026!W",
		IpAddress:  "127.0.0.1",
		UserAgent:  "browser/1.0",
	})
	if err != nil || mfaLogin.GetStatus() != iamv1.AuthStatus_AUTH_STATUS_MFA_REQUIRED {
		t.Fatalf("AuthenticatePassword(MFA) failed: resp=%+v err=%v", mfaLogin, err)
	}
	codeStep0, err := env.totpProv.GenerateCodeAt(totpSec, baseNow)
	if err != nil {
		t.Fatalf("GenerateCodeAt(step0) error = %v", err)
	}
	verifiedChallenge, err := authClient.VerifyMfaChallenge(ctx, &iamv1.VerifyMfaChallengeRequest{
		ChallengeToken: mfaLogin.GetMfaChallengeToken(),
		Code:           codeStep0,
		IpAddress:      "127.0.0.1",
		UserAgent:      "browser/1.0",
	})
	if err != nil || verifiedChallenge.GetStatus() != iamv1.AuthStatus_AUTH_STATUS_AUTHENTICATED {
		t.Fatalf("VerifyMfaChallenge failed: resp=%+v err=%v", verifiedChallenge, err)
	}

	// Buat sesi baru non-MFA dan uji VerifySessionMfa dengan step +1 (masih dalam jendela skew +-1).
	issuedForMFA, err := sessClient.IssueSession(ctx, &iamv1.IssueSessionRequest{
		UserId:    uid,
		IpAddress: "127.0.0.1",
		UserAgent: "stepup/1.0",
	})
	if err != nil {
		t.Fatalf("IssueSession(for MFA) error = %v", err)
	}
	codeStepPlus1, err := env.totpProv.GenerateCodeAt(totpSec, baseNow.Add(30*time.Second))
	if err != nil {
		t.Fatalf("GenerateCodeAt(step+1) error = %v", err)
	}
	sessMFAResp, err := authClient.VerifySessionMfa(ctx, &iamv1.VerifySessionMfaRequest{
		UserId:    uid,
		SessionId: issuedForMFA.GetIssuedSession().GetSession().GetId(),
		Code:      codeStepPlus1,
	})
	if err != nil || !sessMFAResp.GetSession().GetMfaVerified() || sessMFAResp.GetCsrfToken() == "" {
		t.Fatalf("VerifySessionMfa failed: resp=%+v err=%v", sessMFAResp, err)
	}

	if _, err := authClient.DisableTotp(ctx, &iamv1.DisableTotpRequest{UserId: uid}); err != nil {
		t.Fatalf("DisableTotp error = %v", err)
	}

	// Uji ChangePassword, RequestPasswordReset, ConfirmPasswordReset, serta UpdateUserEmail & DeletePendingUserEmail.
	chgResp, err := authClient.ChangePassword(ctx, &iamv1.ChangePasswordRequest{
		UserId:          uid,
		CurrentPassword: "SteelKey#2026!W",
		NewPassword:     "NewSteel#2026!W2",
	})
	if err != nil || !chgResp.GetChanged() || chgResp.GetPasswordChangedAt() == nil {
		t.Fatalf("ChangePassword failed: resp=%+v err=%v", chgResp, err)
	}
	resetReq, err := authClient.RequestPasswordReset(ctx, &iamv1.RequestPasswordResetRequest{
		Email: "andi.gudang@integrione.local",
	})
	if err != nil || resetReq.GetIssuedToken().GetPlaintextToken() == "" {
		t.Fatalf("RequestPasswordReset failed: resp=%+v err=%v", resetReq, err)
	}
	if _, err := authClient.ConfirmPasswordReset(ctx, &iamv1.ConfirmPasswordResetRequest{
		RawToken:    resetReq.GetIssuedToken().GetPlaintextToken(),
		NewPassword: "ResetKey#2026!W3",
	}); err != nil {
		t.Fatalf("ConfirmPasswordReset error = %v", err)
	}

	updEmail1, err := userClient.UpdateUserEmail(ctx, &iamv1.UpdateUserEmailRequest{
		UserId:          uid,
		Email:           "andi.pending@integrione.local",
		CurrentPassword: "ResetKey#2026!W3",
	})
	if err != nil || !updEmail1.GetIsPending() || updEmail1.GetPendingEmail() != "andi.pending@integrione.local" || updEmail1.GetVerificationToken() == "" {
		t.Fatalf("UpdateUserEmail(1) failed: resp=%+v err=%v", updEmail1, err)
	}
	resendAuth, err := userClient.ResendValidationEmailToken(ctx, &iamv1.ResendValidationEmailTokenRequest{UserId: uid})
	if err != nil || !resendAuth.GetIsResend() || resendAuth.GetVerificationToken() == updEmail1.GetVerificationToken() {
		t.Fatalf("ResendValidationEmailToken(user) failed: resp=%+v err=%v", resendAuth, err)
	}
	delPend, err := userClient.DeletePendingUserEmail(ctx, &iamv1.DeletePendingUserEmailRequest{UserId: uid})
	if err != nil || !delPend.GetDeleted() {
		t.Fatalf("DeletePendingUserEmail failed: resp=%+v err=%v", delPend, err)
	}
	updEmail2, err := userClient.UpdateUserEmail(ctx, &iamv1.UpdateUserEmailRequest{
		UserId:          uid,
		Email:           "andi.final@integrione.local",
		CurrentPassword: "ResetKey#2026!W3",
	})
	if err != nil || updEmail2.GetVerificationToken() == "" {
		t.Fatalf("UpdateUserEmail(2) failed: resp=%+v err=%v", updEmail2, err)
	}
	verSwap, err := userClient.ConfirmEmailVerificationByToken(ctx, &iamv1.ConfirmEmailVerificationByTokenRequest{
		RawToken: updEmail2.GetVerificationToken(),
	})
	if err != nil || !verSwap.GetVerified() || verSwap.GetIdentity().GetValue() != "andi.final@integrione.local" {
		t.Fatalf("ConfirmEmailVerificationByToken(swap) failed: resp=%+v err=%v", verSwap, err)
	}

	// Login memakai sandi baru hasil reset dan username terkini ('andi.gudang2') lalu terbitkan sesi kedua untuk menguji ListUserSessions, RevokeSessionById, RevokeAllUserSessions, dan CleanupEndedSessions.
	reloginResp, err := authClient.AuthenticatePassword(ctx, &iamv1.AuthenticatePasswordRequest{
		Identifier: "andi.gudang2",
		Password:   "ResetKey#2026!W3",
		IpAddress:  "127.0.0.1",
		UserAgent:  "browser/2.0",
	})
	if err != nil || reloginResp.GetStatus() != iamv1.AuthStatus_AUTH_STATUS_AUTHENTICATED {
		t.Fatalf("AuthenticatePassword(after reset) failed: resp=%+v err=%v", reloginResp, err)
	}
	extraSess, err := sessClient.IssueSession(ctx, &iamv1.IssueSessionRequest{
		UserId:    uid,
		IpAddress: "127.0.0.1",
		UserAgent: "unit-test/3.0",
	})
	if err != nil {
		t.Fatalf("IssueSession(extra) error = %v", err)
	}
	if sessList, err := sessClient.ListUserSessions(ctx, &iamv1.ListUserSessionsRequest{UserId: uid, ActiveOnly: true}); err != nil || len(sessList.GetItems()) < 2 {
		t.Fatalf("ListUserSessions failed: resp=%+v err=%v", sessList, err)
	}
	if _, err := sessClient.RevokeSessionById(ctx, &iamv1.RevokeSessionByIdRequest{
		SessionId: extraSess.GetIssuedSession().GetSession().GetId(),
		Reason:    iamv1.SessionEndReason_SESSION_END_REASON_ADMIN_REVOKED,
	}); err != nil {
		t.Fatalf("RevokeSessionById error = %v", err)
	}
	if revAll, err := sessClient.RevokeAllUserSessions(ctx, &iamv1.RevokeAllUserSessionsRequest{
		UserId: uid,
		Reason: iamv1.SessionEndReason_SESSION_END_REASON_ADMIN_REVOKED,
	}); err != nil || revAll.GetRevokedCount() < 1 {
		t.Fatalf("RevokeAllUserSessions failed: resp=%+v err=%v", revAll, err)
	}
	if cleanResp, err := sessClient.CleanupEndedSessions(ctx, &iamv1.CleanupEndedSessionsRequest{
		Retention: durationpb.New(time.Second),
	}); err != nil || cleanResp.GetDeletedCount() < 0 {
		t.Fatalf("CleanupEndedSessions failed: resp=%+v err=%v", cleanResp, err)
	}
	if _, err := userClient.SoftDeleteUser(ctx, &iamv1.SoftDeleteUserRequest{UserId: uid}); err != nil {
		t.Fatalf("SoftDeleteUser error = %v", err)
	}
}

func TestEmptyRequestsReturnInvalidArgument_Bufconn(t *testing.T) {
	t.Parallel()

	env := startBufconnServer(t)
	ctx := t.Context()

	if _, err := iamv1.NewPermissionServiceClient(env.conn).GetPermissionById(ctx, &iamv1.GetPermissionByIdRequest{}); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("Permission.GetPermissionById empty: got %s (%v)", status.Code(err), err)
	}
	if _, err := iamv1.NewRoleServiceClient(env.conn).GetRoleById(ctx, &iamv1.GetRoleByIdRequest{}); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("Role.GetRoleById empty: got %s (%v)", status.Code(err), err)
	}
	if _, err := iamv1.NewGroupServiceClient(env.conn).GetGroupById(ctx, &iamv1.GetGroupByIdRequest{}); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("Group.GetGroupById empty: got %s (%v)", status.Code(err), err)
	}
	if _, err := iamv1.NewSessionServiceClient(env.conn).AuthenticateSession(ctx, &iamv1.AuthenticateSessionRequest{}); status.Code(err) != codes.InvalidArgument && status.Code(err) != codes.Unauthenticated {
		t.Fatalf("Session.AuthenticateSession empty: got %s (%v)", status.Code(err), err)
	}
	if _, err := iamv1.NewUserServiceClient(env.conn).GetUserById(ctx, &iamv1.GetUserByIdRequest{}); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("User.GetUserById empty: got %s (%v)", status.Code(err), err)
	}
	if _, err := iamv1.NewAuthServiceClient(env.conn).AuthenticatePassword(ctx, &iamv1.AuthenticatePasswordRequest{}); status.Code(err) != codes.InvalidArgument && status.Code(err) != codes.Unauthenticated {
		t.Fatalf("Auth.AuthenticatePassword empty: got %s (%v)", status.Code(err), err)
	}
}

func enrollAndConfirmTOTPAt(
	t *testing.T,
	ctx context.Context,
	authClient iamv1.AuthServiceClient,
	prov *totp.RFC6238Provider,
	userID, label string,
	confirmTime time.Time,
) []byte {
	t.Helper()

	enrollResp, err := authClient.EnrollTotp(ctx, &iamv1.EnrollTotpRequest{
		UserId:       userID,
		AccountLabel: label,
	})
	if err != nil {
		t.Fatalf("EnrollTotp error = %v", err)
	}
	parsedURI, err := url.Parse(enrollResp.GetOtpAuthUri())
	if err != nil {
		t.Fatalf("url.Parse(otp_auth_uri) error = %v", err)
	}
	secretBytes, err := totpDecodeBase32ForTest(parsedURI.Query().Get("secret"))
	if err != nil {
		t.Fatalf("totpDecodeBase32ForTest error = %v", err)
	}
	code, err := prov.GenerateCodeAt(secretBytes, confirmTime)
	if err != nil {
		t.Fatalf("GenerateCodeAt error = %v", err)
	}
	confirmResp, err := authClient.ConfirmTotp(ctx, &iamv1.ConfirmTotpRequest{
		UserId: userID,
		Code:   code,
	})
	if err != nil || !confirmResp.GetFactor().GetIsEnabled() {
		t.Fatalf("ConfirmTotp failed: resp=%+v err=%v", confirmResp, err)
	}
	return secretBytes
}
