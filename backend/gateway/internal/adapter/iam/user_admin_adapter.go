package iam

import (
	"context"
	"strings"

	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/browserauth"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	"github.com/integrione/erp-retail/backend/gateway/internal/source_connections"
	iamv1 "github.com/integrione/erp-retail/backend/services/iam/contract/gen/iam/v1"
	"github.com/integrione/erp-retail/backend/shared/authz"
)

type createUserParams struct {
	fullName string
	email    string
	username string
	password string
}

// User mengambil detail satu akun pengguna berdasarkan GID atau UUID mentah (wajib izin iam.users.read).
func (a *Adapter) User(ctx context.Context, id string) (*model.User, error) {
	sess, err := a.requirePerm(ctx, "iam.users.read", false)
	if err != nil {
		return nil, err
	}
	cleanID, err := ParseTypedID(id, ResourceUser)
	if err != nil {
		return nil, err
	}
	rpcCtx, cancel := a.rpcCtx(ctx, sess.Caller)
	defer cancel()

	resp, rpcErr := a.clients.Users.GetUserById(rpcCtx, &iamv1.GetUserByIdRequest{
		UserId: cleanID,
	})
	if rpcErr != nil {
		return nil, mapRPCError(rpcErr)
	}
	return toModelUser(resp.GetUser()), nil
}

// authorizeUserRef menormalisasi referensi user ID (baik GID maupun UUID mentah) dan memvalidasi akses self-or-permission.
func (a *Adapter) authorizeUserRef(
	ctx context.Context,
	userRefID string,
	selfPermCode string,
	adminPermCode string,
) (*middleware.ResolvedSession, string, error) {
	cleanID, err := ParseTypedID(userRefID, ResourceUser)
	if err != nil {
		return nil, "", err
	}
	sess, err := a.requireSelfOrPerm(ctx, cleanID, selfPermCode, adminPermCode)
	if err != nil {
		return nil, "", err
	}
	return sess, cleanID, nil
}

// UserForNested mengambil data User induk bagi field bersarang UserIdentity.user atau Session.user dengan mengizinkan pemilik sendiri (*self) maupun iam.users.read.
func (a *Adapter) UserForNested(ctx context.Context, userRefID string) (*model.User, error) {
	sess, cleanID, err := a.authorizeUserRef(ctx, userRefID, authz.PermUsersReadSelf, authz.PermUsersRead)
	if err != nil {
		return nil, err
	}
	rpcCtx, cancel := a.rpcCtx(ctx, sess.Caller)
	defer cancel()

	resp, rpcErr := a.clients.Users.GetUserById(rpcCtx, &iamv1.GetUserByIdRequest{
		UserId: cleanID,
	})
	if rpcErr != nil {
		return nil, mapRPCError(rpcErr)
	}
	return toModelUser(resp.GetUser()), nil
}

// Users mencari dan mempaginasi daftar akun pengguna lintas kategori dalam format Relay Connection (wajib izin iam.users.read).
func (a *Adapter) Users(
	ctx context.Context,
	after *string,
	before *string,
	first *int,
	last *int,
	query *string,
	kind *model.UserKind,
	status *model.UserStatus,
	includeDeleted *bool,
) (*model.UserConnection, error) {
	sess, err := a.requirePerm(ctx, "iam.users.read", false)
	if err != nil {
		return nil, err
	}
	win, err := ParseRelayPage(after, before, first, last)
	if err != nil {
		return nil, err
	}

	var protoKind *iamv1.UserKind
	if kind != nil {
		k := toProtoUserKind(*kind)
		protoKind = &k
	}
	var protoStatus *iamv1.UserStatus
	if status != nil {
		s := toProtoUserStatus(*status)
		protoStatus = &s
	}

	items, startOffset, pageInfo, total, err := FetchRelayPage(win, func(limit, offset int32) ([]*iamv1.User, *iamv1.PageInfo, error) {
		rpcCtx, cancel := a.rpcCtx(ctx, sess.Caller)
		defer cancel()
		resp, rpcErr := a.clients.Users.ListUsers(rpcCtx, &iamv1.ListUsersRequest{
			Query:          derefStr(query),
			Kind:           protoKind,
			Status:         protoStatus,
			IncludeDeleted: derefBool(includeDeleted, false),
			Page:           &iamv1.PageRequest{Limit: limit, Offset: offset},
		})
		if rpcErr != nil {
			return nil, nil, mapRPCError(rpcErr)
		}
		return resp.GetItems(), resp.GetPage(), nil
	})
	if err != nil {
		return nil, err
	}
	return buildUserConnection(items, startOffset, pageInfo, total), nil
}

// UserIdentities mengambil daftar identitas login pengguna dalam format Relay Connection (wajib izin iam.users.read).
func (a *Adapter) UserIdentities(
	ctx context.Context,
	userID string,
	after *string,
	before *string,
	first *int,
	last *int,
) (*model.UserIdentityConnection, error) {
	sess, err := a.requirePerm(ctx, "iam.users.read", false)
	if err != nil {
		return nil, err
	}
	cleanUserID, err := ParseTypedID(userID, ResourceUser)
	if err != nil {
		return nil, err
	}
	return a.fetchUserIdentitiesConnection(ctx, sess, cleanUserID, after, before, first, last)
}

// UserIdentitiesForUser meresolusi field User.identities dengan mengizinkan principal pemilik maupun pemegang iam.users.read.
func (a *Adapter) UserIdentitiesForUser(
	ctx context.Context,
	rawUserID string,
	after *string,
	before *string,
	first *int,
	last *int,
) (*model.UserIdentityConnection, error) {
	sess, cleanUserID, err := a.authorizeUserRef(ctx, rawUserID, authz.PermUsersReadSelf, authz.PermUsersRead)
	if err != nil {
		return nil, err
	}
	return a.fetchUserIdentitiesConnection(ctx, sess, cleanUserID, after, before, first, last)
}

func (a *Adapter) fetchUserIdentitiesConnection(
	ctx context.Context,
	sess *middleware.ResolvedSession,
	cleanUserID string,
	after *string,
	before *string,
	first *int,
	last *int,
) (*model.UserIdentityConnection, error) {
	rpcCtx, cancel := a.rpcCtx(ctx, sess.Caller)
	defer cancel()

	resp, rpcErr := a.clients.Users.ListUserIdentities(rpcCtx, &iamv1.ListUserIdentitiesRequest{
		UserId: cleanUserID,
	})
	if rpcErr != nil {
		return nil, mapRPCError(rpcErr)
	}
	sliced, startOffset, pageInfo, total, err := SliceRelayWindow(resp.GetIdentities(), after, before, first, last)
	if err != nil {
		return nil, err
	}
	return buildUserIdentityConnection(sliced, startOffset, pageInfo, total), nil
}

// UserRoles mengambil daftar penugasan peran langsung pengguna dalam format Relay Connection (wajib izin iam.roles.read).
func (a *Adapter) UserRoles(
	ctx context.Context,
	userID string,
	includeExpired *bool,
	after *string,
	before *string,
	first *int,
	last *int,
) (*model.UserRoleConnection, error) {
	sess, err := a.requirePerm(ctx, "iam.roles.read", false)
	if err != nil {
		return nil, err
	}
	cleanUserID, err := ParseTypedID(userID, ResourceUser)
	if err != nil {
		return nil, err
	}
	return a.fetchUserRolesConnection(ctx, sess, cleanUserID, includeExpired, after, before, first, last)
}

// UserRolesForUser meresolusi field User.roles dengan mengizinkan principal pemilik maupun pemegang iam.roles.read.
func (a *Adapter) UserRolesForUser(
	ctx context.Context,
	rawUserID string,
	includeExpired *bool,
	after *string,
	before *string,
	first *int,
	last *int,
) (*model.UserRoleConnection, error) {
	sess, cleanUserID, err := a.authorizeUserRef(ctx, rawUserID, authz.PermUsersReadSelf, authz.PermRolesRead)
	if err != nil {
		return nil, err
	}
	return a.fetchUserRolesConnection(ctx, sess, cleanUserID, includeExpired, after, before, first, last)
}

func (a *Adapter) fetchUserRolesConnection(
	ctx context.Context,
	sess *middleware.ResolvedSession,
	cleanUserID string,
	includeExpired *bool,
	after *string,
	before *string,
	first *int,
	last *int,
) (*model.UserRoleConnection, error) {
	rpcCtx, cancel := a.rpcCtx(ctx, sess.Caller)
	defer cancel()

	resp, rpcErr := a.clients.Roles.ListUserRoles(rpcCtx, &iamv1.ListUserRolesRequest{
		UserId:         cleanUserID,
		IncludeExpired: derefBool(includeExpired, false),
	})
	if rpcErr != nil {
		return nil, mapRPCError(rpcErr)
	}
	sliced, startOffset, pageInfo, total, err := SliceRelayWindow(resp.GetUserRoles(), after, before, first, last)
	if err != nil {
		return nil, err
	}

	roleIDs := make([]string, 0, len(sliced))
	for _, ur := range sliced {
		roleIDs = append(roleIDs, ur.GetRoleId())
	}
	roleMap, err := a.fetchRoleMap(rpcCtx, roleIDs)
	if err != nil {
		return nil, err
	}

	edges := make([]*model.UserRoleEdge, 0, len(sliced))
	nodes := make([]*model.Role, 0, len(sliced))
	for i, ur := range sliced {
		rNode := roleMap[ur.GetRoleId()]
		edges = append(edges, toModelUserRoleEdge(ur, rNode, EncodeCursor(startOffset+i)))
		if rNode != nil {
			nodes = append(nodes, rNode)
		}
	}
	return &model.UserRoleConnection{
		Edges:      edges,
		Nodes:      nodes,
		PageInfo:   pageInfo,
		TotalCount: total,
	}, nil
}

// UserEffectiveRoles mengambil seluruh peran efektif gabungan langsung dan pewarisan grup (wajib izin iam.roles.read).
func (a *Adapter) UserEffectiveRoles(
	ctx context.Context,
	userID string,
	after *string,
	before *string,
	first *int,
	last *int,
) (*model.RoleConnection, error) {
	sess, err := a.requirePerm(ctx, "iam.roles.read", false)
	if err != nil {
		return nil, err
	}
	cleanUserID, err := ParseTypedID(userID, ResourceUser)
	if err != nil {
		return nil, err
	}
	return a.fetchUserEffectiveRolesConnection(ctx, sess, cleanUserID, after, before, first, last)
}

// UserEffectiveRolesForUser meresolusi field User.effective_roles dengan mengizinkan pemilik maupun iam.roles.read.
func (a *Adapter) UserEffectiveRolesForUser(
	ctx context.Context,
	rawUserID string,
	after *string,
	before *string,
	first *int,
	last *int,
) (*model.RoleConnection, error) {
	sess, cleanUserID, err := a.authorizeUserRef(ctx, rawUserID, authz.PermUsersReadSelf, authz.PermRolesRead)
	if err != nil {
		return nil, err
	}
	return a.fetchUserEffectiveRolesConnection(ctx, sess, cleanUserID, after, before, first, last)
}

func (a *Adapter) fetchUserEffectiveRolesConnection(
	ctx context.Context,
	sess *middleware.ResolvedSession,
	cleanUserID string,
	after *string,
	before *string,
	first *int,
	last *int,
) (*model.RoleConnection, error) {
	rpcCtx, cancel := a.rpcCtx(ctx, sess.Caller)
	defer cancel()

	resp, rpcErr := a.clients.Roles.ListEffectiveUserRoles(rpcCtx, &iamv1.ListEffectiveUserRolesRequest{
		UserId: cleanUserID,
	})
	if rpcErr != nil {
		return nil, mapRPCError(rpcErr)
	}
	sliced, startOffset, pageInfo, total, err := SliceRelayWindow(resp.GetRoles(), after, before, first, last)
	if err != nil {
		return nil, err
	}
	return buildRoleConnection(sliced, startOffset, pageInfo, total), nil
}

// UserGroups mengambil daftar grup tempat pengguna tergabung dalam format Relay Connection (wajib izin iam.groups.read).
func (a *Adapter) UserGroups(
	ctx context.Context,
	userID string,
	after *string,
	before *string,
	first *int,
	last *int,
) (*model.GroupConnection, error) {
	sess, err := a.requirePerm(ctx, "iam.groups.read", false)
	if err != nil {
		return nil, err
	}
	cleanUserID, err := ParseTypedID(userID, ResourceUser)
	if err != nil {
		return nil, err
	}
	return a.fetchUserGroupsConnection(ctx, sess, cleanUserID, after, before, first, last)
}

// UserGroupsForUser meresolusi field User.groups dengan mengizinkan pemilik maupun iam.groups.read.
func (a *Adapter) UserGroupsForUser(
	ctx context.Context,
	rawUserID string,
	after *string,
	before *string,
	first *int,
	last *int,
) (*model.GroupConnection, error) {
	sess, cleanUserID, err := a.authorizeUserRef(ctx, rawUserID, authz.PermUsersReadSelf, authz.PermGroupsRead)
	if err != nil {
		return nil, err
	}
	return a.fetchUserGroupsConnection(ctx, sess, cleanUserID, after, before, first, last)
}

func (a *Adapter) fetchUserGroupsConnection(
	ctx context.Context,
	sess *middleware.ResolvedSession,
	cleanUserID string,
	after *string,
	before *string,
	first *int,
	last *int,
) (*model.GroupConnection, error) {
	rpcCtx, cancel := a.rpcCtx(ctx, sess.Caller)
	defer cancel()

	resp, rpcErr := a.clients.Groups.ListUserGroups(rpcCtx, &iamv1.ListUserGroupsRequest{
		UserId: cleanUserID,
	})
	if rpcErr != nil {
		return nil, mapRPCError(rpcErr)
	}
	sliced, startOffset, pageInfo, total, err := SliceRelayWindow(resp.GetGroups(), after, before, first, last)
	if err != nil {
		return nil, err
	}
	return buildGroupConnection(sliced, startOffset, pageInfo, total), nil
}

// UserSessions mengambil daftar sesi milik pengguna tertentu dalam format Relay Connection (wajib izin iam.sessions.read).
func (a *Adapter) UserSessions(
	ctx context.Context,
	userID string,
	activeOnly *bool,
	after *string,
	before *string,
	first *int,
	last *int,
) (*model.SessionConnection, error) {
	sess, err := a.requirePerm(ctx, "iam.sessions.read", false)
	if err != nil {
		return nil, err
	}
	cleanUserID, err := ParseTypedID(userID, ResourceUser)
	if err != nil {
		return nil, err
	}
	return a.fetchUserSessionsConnection(ctx, sess, cleanUserID, activeOnly, after, before, first, last)
}

// UserSessionsForUser meresolusi field User.sessions dengan mengizinkan pemilik maupun iam.sessions.read.
func (a *Adapter) UserSessionsForUser(
	ctx context.Context,
	rawUserID string,
	activeOnly *bool,
	after *string,
	before *string,
	first *int,
	last *int,
) (*model.SessionConnection, error) {
	sess, cleanUserID, err := a.authorizeUserRef(ctx, rawUserID, authz.PermSessionsReadSelf, authz.PermSessionsRead)
	if err != nil {
		return nil, err
	}
	return a.fetchUserSessionsConnection(ctx, sess, cleanUserID, activeOnly, after, before, first, last)
}

func (a *Adapter) fetchUserSessionsConnection(
	ctx context.Context,
	sess *middleware.ResolvedSession,
	cleanUserID string,
	activeOnly *bool,
	after *string,
	before *string,
	first *int,
	last *int,
) (*model.SessionConnection, error) {
	win, err := ParseRelayPage(after, before, first, last)
	if err != nil {
		return nil, err
	}
	items, startOffset, pageInfo, total, err := FetchRelayPage(win, func(limit, offset int32) ([]*iamv1.Session, *iamv1.PageInfo, error) {
		rpcCtx, cancel := a.rpcCtx(ctx, sess.Caller)
		defer cancel()
		resp, rpcErr := a.clients.Sessions.ListUserSessions(rpcCtx, &iamv1.ListUserSessionsRequest{
			UserId:     cleanUserID,
			ActiveOnly: derefBool(activeOnly, true),
			Page:       &iamv1.PageRequest{Limit: limit, Offset: offset},
		})
		if rpcErr != nil {
			return nil, nil, mapRPCError(rpcErr)
		}
		return resp.GetItems(), resp.GetPage(), nil
	})
	if err != nil {
		return nil, err
	}
	return buildSessionConnection(items, startOffset, pageInfo, total), nil
}

// CreateUser membuat akun kategori pelanggan (CUSTOMER) untuk registrasi publik maupun admin dengan status awal PENDING hingga verifikasi email.
func (a *Adapter) CreateUser(ctx context.Context, input model.CreateUserInput) (*model.CreateUserPayload, error) {
	fail := func(err error) (*model.CreateUserPayload, error) {
		return &model.CreateUserPayload{
			User:          nil,
			Errors:        ToMutationErrors(err),
			WasSuccessful: false,
		}, nil
	}

	sess, err := a.authManager.AuthenticateOptional(ctx)
	if err != nil {
		return fail(err)
	}
	isAuthenticated := sess != nil && sess.Token != nil
	isAdmin := false
	if isAuthenticated && sess.Snapshot != nil {
		_, isAdmin = sess.Snapshot.HasPermission("iam.users.write")
	}

	displayName := strings.TrimSpace(input.FullName)
	email := strings.TrimSpace(input.Email)
	username := strings.TrimSpace(input.Username)

	// Lindungi endpoint registrasi publik dengan pembatasan laju per IP dan subjek email sebelum menyentuh gRPC IAM.
	if !isAdmin {
		if rateErr := a.checkAuthRateLimit(ctx, email); rateErr != nil {
			return fail(rateErr)
		}
	}

	if displayName == "" || email == "" || username == "" || input.Password == "" {
		return fail(middleware.ErrInvalidArgument("full_name, email, username, and password are required", nil))
	}

	caller := a.authManager.BuildCallerContext(ctx)
	if isAuthenticated {
		caller = sess.Caller
	}
	rpcCtx, cancel := a.rpcCtx(ctx, caller)
	defer cancel()

	resp, rpcErr := a.clients.Users.CreateUser(rpcCtx, &iamv1.CreateUserRequest{
		Kind:                   iamv1.UserKind_USER_KIND_CUSTOMER,
		Status:                 iamv1.UserStatus_USER_STATUS_PENDING,
		DisplayName:            displayName,
		Email:                  email,
		Username:               username,
		Password:               input.Password,
		RequirePasswordChange:  false,
		MarkIdentitiesVerified: false,
	})
	if rpcErr != nil {
		reason := source_connections.ExtractIAMErrorReason(rpcErr)
		field := source_connections.ExtractIAMErrorField(rpcErr)
		if reason == "iam.identity.already_exists" {
			if field == "username" {
				return fail(middleware.ErrConflict("username has been taken", rpcErr))
			}
			if !isAdmin && field == "email" {
				// Alamat email bersifat sensitif pada jalur publik: arahkan ke /login tanpa pesan bahwa email sudah terdaftar.
				if eff := browserauth.EffectsFromContext(ctx); eff != nil {
					eff.SetHeader(browserauth.LocationHeaderName, "/login")
				}
				return &model.CreateUserPayload{
					User:             nil,
					RedirectLocation: ptrStr("/login"),
					Errors:           noMutationErrors(),
					WasSuccessful:    true,
				}, nil
			}
		}
		return fail(mapRPCError(rpcErr))
	}

	// Hanya kembalikan objek User apabila pemanggil memiliki sesi login yang sah.
	var outUser *model.User
	if isAuthenticated {
		outUser = toModelUser(resp.GetUser())
	}
	return &model.CreateUserPayload{
		User:          outUser,
		Errors:        noMutationErrors(),
		WasSuccessful: true,
	}, nil
}

// CreateEmployeeUser membuat akun internal karyawan (EMPLOYEE) dengan status awal PENDING hingga verifikasi email.
func (a *Adapter) CreateEmployeeUser(ctx context.Context, input model.CreateEmployeeUserInput) (*model.CreateEmployeeUserPayload, error) {
	user, err := a.createUserByKind(ctx, model.UserKindEmployee, createUserParams{
		fullName: input.FullName,
		email:    input.Email,
		username: input.Username,
		password: input.Password,
	}, true)
	if err != nil {
		return &model.CreateEmployeeUserPayload{Errors: ToMutationErrors(err), WasSuccessful: false}, nil
	}
	return &model.CreateEmployeeUserPayload{User: user, Errors: noMutationErrors(), WasSuccessful: true}, nil
}

// CreateVendorUser membuat akun mitra vendor (VENDOR) dengan status awal PENDING hingga verifikasi email.
func (a *Adapter) CreateVendorUser(ctx context.Context, input model.CreateVendorUserInput) (*model.CreateVendorUserPayload, error) {
	user, err := a.createUserByKind(ctx, model.UserKindVendor, createUserParams{
		fullName: input.FullName,
		email:    input.Email,
		username: input.Username,
		password: input.Password,
	}, true)
	if err != nil {
		return &model.CreateVendorUserPayload{Errors: ToMutationErrors(err), WasSuccessful: false}, nil
	}
	return &model.CreateVendorUserPayload{User: user, Errors: noMutationErrors(), WasSuccessful: true}, nil
}

// CreateSupplierUser membuat akun mitra pemasok (SUPPLIER) dengan status awal PENDING hingga verifikasi email.
func (a *Adapter) CreateSupplierUser(ctx context.Context, input model.CreateSupplierUserInput) (*model.CreateSupplierUserPayload, error) {
	user, err := a.createUserByKind(ctx, model.UserKindSupplier, createUserParams{
		fullName: input.FullName,
		email:    input.Email,
		username: input.Username,
		password: input.Password,
	}, true)
	if err != nil {
		return &model.CreateSupplierUserPayload{Errors: ToMutationErrors(err), WasSuccessful: false}, nil
	}
	return &model.CreateSupplierUserPayload{User: user, Errors: noMutationErrors(), WasSuccessful: true}, nil
}

func (a *Adapter) createUserByKind(ctx context.Context, kind model.UserKind, p createUserParams, defaultReqPwdChange bool) (*model.User, error) {
	sess, err := a.requirePerm(ctx, "iam.users.write", true)
	if err != nil {
		return nil, err
	}
	displayName := strings.TrimSpace(p.fullName)
	email := strings.TrimSpace(p.email)
	username := strings.TrimSpace(p.username)
	if displayName == "" || email == "" || username == "" || p.password == "" {
		return nil, middleware.ErrInvalidArgument("full_name, email, username, and password are required", nil)
	}

	rpcCtx, cancel := a.rpcCtx(ctx, sess.Caller)
	defer cancel()

	resp, rpcErr := a.clients.Users.CreateUser(rpcCtx, &iamv1.CreateUserRequest{
		Kind:                   toProtoUserKind(kind),
		Status:                 iamv1.UserStatus_USER_STATUS_PENDING,
		DisplayName:            displayName,
		Email:                  email,
		Username:               username,
		Password:               p.password,
		RequirePasswordChange:  defaultReqPwdChange,
		MarkIdentitiesVerified: false,
	})
	if rpcErr != nil {
		return nil, mapRPCError(rpcErr)
	}
	return toModelUser(resp.GetUser()), nil
}

func (a *Adapter) mutateUserStatusByID(ctx context.Context, userID string, allowSelf bool, call func(context.Context, string) (*iamv1.User, error)) (*model.User, error) {
	sess, err := a.requirePerm(ctx, "iam.users.write", true)
	if err != nil {
		return nil, err
	}
	cleanUserID, err := ParseTypedID(userID, ResourceUser)
	if err != nil {
		return nil, err
	}
	rpcCtx, cancel := a.rpcCtx(ctx, sess.Caller)
	defer cancel()

	if err := a.ensureCanTouchTargetUser(rpcCtx, sess, cleanUserID, allowSelf); err != nil {
		return nil, err
	}
	user, rpcErr := call(rpcCtx, cleanUserID)
	if rpcErr != nil {
		return nil, mapRPCError(rpcErr)
	}
	return toModelUser(user), nil
}

// ActivateUser mengaktifkan status akun pengguna tertentu (wajib iam.users.write + reauth sensitif).
func (a *Adapter) ActivateUser(ctx context.Context, userID string) (*model.ActivateUserPayload, error) {
	user, err := a.mutateUserStatusByID(ctx, userID, true, func(rpcCtx context.Context, id string) (*iamv1.User, error) {
		resp, rpcErr := a.clients.Users.ActivateUser(rpcCtx, &iamv1.ActivateUserRequest{UserId: id})
		if rpcErr != nil {
			return nil, rpcErr
		}
		return resp.GetUser(), nil
	})
	if err != nil {
		return &model.ActivateUserPayload{Errors: ToMutationErrors(err), WasSuccessful: false}, nil
	}
	return &model.ActivateUserPayload{User: user, Errors: noMutationErrors(), WasSuccessful: true}, nil
}

// SuspendUser membekukan sementara akun pengguna (wajib iam.users.write + reauth sensitif).
func (a *Adapter) SuspendUser(ctx context.Context, userID string) (*model.SuspendUserPayload, error) {
	user, err := a.mutateUserStatusByID(ctx, userID, false, func(rpcCtx context.Context, id string) (*iamv1.User, error) {
		resp, rpcErr := a.clients.Users.SuspendUser(rpcCtx, &iamv1.SuspendUserRequest{UserId: id})
		if rpcErr != nil {
			return nil, rpcErr
		}
		return resp.GetUser(), nil
	})
	if err != nil {
		return &model.SuspendUserPayload{Errors: ToMutationErrors(err), WasSuccessful: false}, nil
	}
	return &model.SuspendUserPayload{User: user, Errors: noMutationErrors(), WasSuccessful: true}, nil
}

// DisableUser menonaktifkan permanen akun pengguna dan mencabut sesi aktifnya (wajib iam.users.write + reauth sensitif).
func (a *Adapter) DisableUser(ctx context.Context, userID string) (*model.DisableUserPayload, error) {
	user, err := a.mutateUserStatusByID(ctx, userID, false, func(rpcCtx context.Context, id string) (*iamv1.User, error) {
		resp, rpcErr := a.clients.Users.DisableUser(rpcCtx, &iamv1.DisableUserRequest{UserId: id})
		if rpcErr != nil {
			return nil, rpcErr
		}
		return resp.GetUser(), nil
	})
	if err != nil {
		return &model.DisableUserPayload{Errors: ToMutationErrors(err), WasSuccessful: false}, nil
	}
	return &model.DisableUserPayload{User: user, Errors: noMutationErrors(), WasSuccessful: true}, nil
}

// SoftDeleteUser menandai penghapusan lunak akun pengguna (wajib iam.users.write + reauth sensitif).
func (a *Adapter) SoftDeleteUser(ctx context.Context, userID string) (*model.SoftDeleteUserPayload, error) {
	fail := func(err error) (*model.SoftDeleteUserPayload, error) {
		return &model.SoftDeleteUserPayload{Errors: ToMutationErrors(err), WasSuccessful: false}, nil
	}
	sess, err := a.requirePerm(ctx, "iam.users.write", true)
	if err != nil {
		return fail(err)
	}
	cleanUserID, err := ParseTypedID(userID, ResourceUser)
	if err != nil {
		return fail(err)
	}
	rpcCtx, cancel := a.rpcCtx(ctx, sess.Caller)
	defer cancel()

	if err := a.ensureCanTouchTargetUser(rpcCtx, sess, cleanUserID, false); err != nil {
		return fail(err)
	}
	_, rpcErr := a.clients.Users.SoftDeleteUser(rpcCtx, &iamv1.SoftDeleteUserRequest{UserId: cleanUserID})
	if rpcErr != nil {
		return fail(mapRPCError(rpcErr))
	}
	return &model.SoftDeleteUserPayload{Deleted: ptrBool(true), Errors: noMutationErrors(), WasSuccessful: true}, nil
}

// SetPrimaryUserIdentity menetapkan identitas utama pengguna untuk jenis email/username terkait.
func (a *Adapter) SetPrimaryUserIdentity(ctx context.Context, userID, identityID string) (*model.SetPrimaryUserIdentityPayload, error) {
	fail := func(err error) (*model.SetPrimaryUserIdentityPayload, error) {
		return &model.SetPrimaryUserIdentityPayload{Errors: ToMutationErrors(err), WasSuccessful: false}, nil
	}
	sess, err := a.requirePerm(ctx, "iam.users.write", true)
	if err != nil {
		return fail(err)
	}
	cleanUserID, err := ParseTypedID(userID, ResourceUser)
	if err != nil {
		return fail(err)
	}
	cleanIdentityID, err := ParseTypedID(identityID, ResourceUserIdentity)
	if err != nil {
		return fail(err)
	}
	rpcCtx, cancel := a.rpcCtx(ctx, sess.Caller)
	defer cancel()

	if err := a.ensureCanTouchTargetUser(rpcCtx, sess, cleanUserID, true); err != nil {
		return fail(err)
	}
	resp, rpcErr := a.clients.Users.SetPrimaryUserIdentity(rpcCtx, &iamv1.SetPrimaryUserIdentityRequest{
		UserId:     cleanUserID,
		IdentityId: cleanIdentityID,
	})
	if rpcErr != nil {
		return fail(mapRPCError(rpcErr))
	}
	return &model.SetPrimaryUserIdentityPayload{
		UserIdentity:  toModelUserIdentity(resp.GetIdentity()),
		Errors:        noMutationErrors(),
		WasSuccessful: true,
	}, nil
}

// RemoveUserIdentity mencabut identitas login milik pengguna (wajib iam.users.write + reauth sensitif).
func (a *Adapter) RemoveUserIdentity(ctx context.Context, userID, identityID string) (*model.RemoveUserIdentityPayload, error) {
	fail := func(err error) (*model.RemoveUserIdentityPayload, error) {
		return &model.RemoveUserIdentityPayload{Errors: ToMutationErrors(err), WasSuccessful: false}, nil
	}
	sess, err := a.requirePerm(ctx, "iam.users.write", true)
	if err != nil {
		return fail(err)
	}
	cleanUserID, err := ParseTypedID(userID, ResourceUser)
	if err != nil {
		return fail(err)
	}
	cleanIdentityID, err := ParseTypedID(identityID, ResourceUserIdentity)
	if err != nil {
		return fail(err)
	}
	rpcCtx, cancel := a.rpcCtx(ctx, sess.Caller)
	defer cancel()

	if err := a.ensureCanTouchTargetUser(rpcCtx, sess, cleanUserID, true); err != nil {
		return fail(err)
	}
	_, rpcErr := a.clients.Users.RemoveUserIdentity(rpcCtx, &iamv1.RemoveUserIdentityRequest{
		UserId:     cleanUserID,
		IdentityId: cleanIdentityID,
	})
	if rpcErr != nil {
		return fail(mapRPCError(rpcErr))
	}
	return &model.RemoveUserIdentityPayload{Removed: ptrBool(true), Errors: noMutationErrors(), WasSuccessful: true}, nil
}
