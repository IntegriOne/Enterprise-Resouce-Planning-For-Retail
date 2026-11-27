package iam

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/browserauth"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	"github.com/integrione/erp-retail/backend/gateway/internal/source_connections"
	iamv1 "github.com/integrione/erp-retail/backend/services/iam/contract/gen/iam/v1"
	"github.com/integrione/erp-retail/backend/shared/authcontext"
	"github.com/integrione/erp-retail/backend/shared/authz"
)

type createUserParams struct {
	provisioningRequestID string
	workerID              string
	fullName              string
	email                 string
	username              string
	password              string
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
	resp, rpcErr := callIAMUsers(ctx, a, &sess.Caller, func(rpcCtx context.Context, users iamv1.UserServiceClient) (*iamv1.GetUserByIdResponse, error) {
		return users.GetUserById(rpcCtx, &iamv1.GetUserByIdRequest{
			UserId: cleanID,
		})
	})
	if rpcErr != nil {
		return nil, rpcErr
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
	resp, rpcErr := callIAMUsers(ctx, a, &sess.Caller, func(rpcCtx context.Context, users iamv1.UserServiceClient) (*iamv1.GetUserByIdResponse, error) {
		return users.GetUserById(rpcCtx, &iamv1.GetUserByIdRequest{
			UserId: cleanID,
		})
	})
	if rpcErr != nil {
		return nil, rpcErr
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
		resp, rpcErr := callIAMUsers(ctx, a, &sess.Caller, func(rpcCtx context.Context, users iamv1.UserServiceClient) (*iamv1.ListUsersResponse, error) {
			return users.ListUsers(rpcCtx, &iamv1.ListUsersRequest{
				Query:          derefStr(query),
				Kind:           protoKind,
				Status:         protoStatus,
				IncludeDeleted: derefBool(includeDeleted, false),
				Page:           &iamv1.PageRequest{Limit: limit, Offset: offset},
			})
		})
		if rpcErr != nil {
			return nil, nil, rpcErr
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
	resp, rpcErr := callIAMUsers(ctx, a, &sess.Caller, func(rpcCtx context.Context, users iamv1.UserServiceClient) (*iamv1.ListUserIdentitiesResponse, error) {
		return users.ListUserIdentities(rpcCtx, &iamv1.ListUserIdentitiesRequest{
			UserId: cleanUserID,
		})
	})
	if rpcErr != nil {
		return nil, rpcErr
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
	resp, rpcErr := callIAMRoles(ctx, a, &sess.Caller, func(rpcCtx context.Context, client iamv1.RoleServiceClient) (*iamv1.ListUserRolesResponse, error) {
		return client.ListUserRoles(rpcCtx, &iamv1.ListUserRolesRequest{
			UserId:         cleanUserID,
			IncludeExpired: derefBool(includeExpired, false),
		})
	})
	if rpcErr != nil {
		return nil, rpcErr
	}
	sliced, startOffset, pageInfo, total, err := SliceRelayWindow(resp.GetUserRoles(), after, before, first, last)
	if err != nil {
		return nil, err
	}

	roleIDs := make([]string, 0, len(sliced))
	for _, ur := range sliced {
		roleIDs = append(roleIDs, ur.GetRoleId())
	}
	roleMap, err := a.fetchRoleMap(authcontext.WithCaller(ctx, sess.Caller), roleIDs)
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
	resp, rpcErr := callIAMRoles(ctx, a, &sess.Caller, func(rpcCtx context.Context, client iamv1.RoleServiceClient) (*iamv1.ListEffectiveUserRolesResponse, error) {
		return client.ListEffectiveUserRoles(rpcCtx, &iamv1.ListEffectiveUserRolesRequest{
			UserId: cleanUserID,
		})
	})
	if rpcErr != nil {
		return nil, rpcErr
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
	resp, rpcErr := callIAMGroups(ctx, a, &sess.Caller, func(rpcCtx context.Context, client iamv1.GroupServiceClient) (*iamv1.ListUserGroupsResponse, error) {
		return client.ListUserGroups(rpcCtx, &iamv1.ListUserGroupsRequest{
			UserId: cleanUserID,
		})
	})
	if rpcErr != nil {
		return nil, rpcErr
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
		resp, rpcErr := callIAMSessions(ctx, a, &sess.Caller, func(rpcCtx context.Context, client iamv1.SessionServiceClient) (*iamv1.ListUserSessionsResponse, error) {
			return client.ListUserSessions(rpcCtx, &iamv1.ListUserSessionsRequest{
				UserId:     cleanUserID,
				ActiveOnly: derefBool(activeOnly, true),
				Page:       &iamv1.PageRequest{Limit: limit, Offset: offset},
			})
		})
		if rpcErr != nil {
			return nil, nil, rpcErr
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
	if err := a.ensureAuthDeps(); err != nil {
		return fail(err)
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

	resp, rpcErr := callIAMUsers(ctx, a, &caller, func(rpcCtx context.Context, users iamv1.UserServiceClient) (*iamv1.CreateUserResponse, error) {
		return users.CreateUser(rpcCtx, &iamv1.CreateUserRequest{
			Kind:                   iamv1.UserKind_USER_KIND_CUSTOMER,
			Status:                 iamv1.UserStatus_USER_STATUS_PENDING,
			DisplayName:            displayName,
			Email:                  email,
			Username:               username,
			Password:               input.Password,
			RequirePasswordChange:  false,
			MarkIdentitiesVerified: false,
		})
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
// Gateway memverifikasi pekerja dan mewajibkan penautan ke layanan Human Resource.
func (a *Adapter) CreateEmployeeUser(ctx context.Context, input model.CreateEmployeeUserInput) (*model.CreateEmployeeUserPayload, error) {
	sess, err := a.requireEmployeeProvisioning(ctx)
	if err != nil {
		return &model.CreateEmployeeUserPayload{Errors: ToMutationErrors(err), WasSuccessful: false}, nil
	}
	params := createUserParams{
		fullName: input.FullName,
		email:    input.Email,
		username: input.Username,
		password: input.Password,
	}
	if err := validateCreateUserParams(params); err != nil {
		return &model.CreateEmployeeUserPayload{Errors: ToMutationErrors(err), WasSuccessful: false}, nil
	}

	cleanWorkerID, idErr := parseStrictUUIDOrTypedID(strings.TrimSpace(input.WorkerID), ResourceWorker)
	if idErr != nil {
		return &model.CreateEmployeeUserPayload{Errors: ToMutationErrors(idErr), WasSuccessful: false}, nil
	}
	if a.workerProvisioner == nil {
		return &model.CreateEmployeeUserPayload{
			Errors:        ToMutationErrors(middleware.ErrUnavailable("human resource service unavailable", nil)),
			WasSuccessful: false,
		}, nil
	}
	normWorkerID := cleanWorkerID
	requestID, err := parseStrictUUIDOrTypedID(input.ProvisioningRequestID, ResourceEmployeeProvisioningRequest)
	if err != nil {
		return &model.CreateEmployeeUserPayload{Errors: ToMutationErrors(err)}, nil
	}
	receipt, err := a.employeeProvisioningRequest(ctx, sess, requestID)
	if err != nil {
		return &model.CreateEmployeeUserPayload{Errors: ToMutationErrors(err)}, nil
	}
	if receipt.WorkerID != EncodeGID(ResourceWorker, cleanWorkerID) {
		return &model.CreateEmployeeUserPayload{Errors: ToMutationErrors(middleware.ErrConflict("worker does not match provisioning request", nil))}, nil
	}
	params.provisioningRequestID, params.workerID = requestID, cleanWorkerID
	if receipt.Status == "pending" {
		normWorkerID, err = a.workerProvisioner.PrepareWorkerPrincipalLink(ctx, cleanWorkerID)
		if err != nil {
			return &model.CreateEmployeeUserPayload{Errors: ToMutationErrors(err)}, nil
		}
	}

	user, createErr := a.createUserByKindWithSession(ctx, sess, model.UserKindEmployee, params, true)
	if createErr != nil {
		return &model.CreateEmployeeUserPayload{Errors: ToMutationErrors(createErr), WasSuccessful: false}, nil
	}
	if user == nil || user.RawID == "" {
		return &model.CreateEmployeeUserPayload{Errors: ToMutationErrors(middleware.ErrUnavailable("invalid IAM user provisioning response", nil))}, nil
	}
	if linkErr := a.workerProvisioner.CompleteWorkerPrincipalLink(ctx, normWorkerID, user.RawID); linkErr != nil {
		return &model.CreateEmployeeUserPayload{Errors: ToMutationErrors(linkErr)}, nil
	}

	_, err = callIAM(ctx, a, &sess.Caller, hasEmployeeProvisioningClient,
		func(rpcCtx context.Context, clients source_connections.IAMClientSet) (*iamv1.CompleteEmployeeProvisioningRequestResponse, error) {
			return clients.EmployeeProvisioning.CompleteEmployeeProvisioningRequest(rpcCtx,
				&iamv1.CompleteEmployeeProvisioningRequestRequest{Id: params.provisioningRequestID, AccountId: user.RawID})
		})
	if err != nil {
		return &model.CreateEmployeeUserPayload{Errors: ToMutationErrors(err)}, nil
	}

	return &model.CreateEmployeeUserPayload{User: user, Errors: noMutationErrors(), WasSuccessful: true}, nil
}

// CreateVendorUser membuat akun mitra vendor (VENDOR) berstatus awal PENDING dan langsung menetapkan FK `party_contacts.iam_user_id` di Enterprise Foundation.
func (a *Adapter) CreateVendorUser(ctx context.Context, input model.CreateVendorUserInput) (*model.CreateVendorUserPayload, error) {
	user, contact, err := a.createPartyPortalUser(ctx, model.UserKindVendor, "VENDOR", partyUserProvisioningParams{
		partyID:          input.PartyID,
		contactID:        input.ContactID,
		contactRole:      input.ContactRole,
		jobTitle:         input.JobTitle,
		phone:            input.Phone,
		isPrimaryContact: input.IsPrimaryContact,
		user: createUserParams{
			fullName: input.FullName,
			email:    input.Email,
			username: input.Username,
			password: input.Password,
		},
	})
	if err != nil {
		return &model.CreateVendorUserPayload{Errors: ToMutationErrors(err), WasSuccessful: false}, nil
	}
	return &model.CreateVendorUserPayload{
		User:          user,
		PartyContact:  contact,
		Errors:        noMutationErrors(),
		WasSuccessful: true,
	}, nil
}

// CreateSupplierUser membuat akun mitra pemasok (SUPPLIER) berstatus awal PENDING dan langsung menetapkan FK `party_contacts.iam_user_id` di Enterprise Foundation.
func (a *Adapter) CreateSupplierUser(ctx context.Context, input model.CreateSupplierUserInput) (*model.CreateSupplierUserPayload, error) {
	user, contact, err := a.createPartyPortalUser(ctx, model.UserKindSupplier, "SUPPLIER", partyUserProvisioningParams{
		partyID:          input.PartyID,
		contactID:        input.ContactID,
		contactRole:      input.ContactRole,
		jobTitle:         input.JobTitle,
		phone:            input.Phone,
		isPrimaryContact: input.IsPrimaryContact,
		user: createUserParams{
			fullName: input.FullName,
			email:    input.Email,
			username: input.Username,
			password: input.Password,
		},
	})
	if err != nil {
		return &model.CreateSupplierUserPayload{Errors: ToMutationErrors(err), WasSuccessful: false}, nil
	}
	return &model.CreateSupplierUserPayload{
		User:          user,
		PartyContact:  contact,
		Errors:        noMutationErrors(),
		WasSuccessful: true,
	}, nil
}

type partyUserProvisioningParams struct {
	partyID          string
	contactID        *string
	contactRole      *string
	jobTitle         *string
	phone            *string
	isPrimaryContact *bool
	user             createUserParams
}

func (a *Adapter) createPartyPortalUser(
	ctx context.Context,
	kind model.UserKind,
	requiredRole string,
	in partyUserProvisioningParams,
) (*model.User, *model.PartyContact, error) {
	sess, err := a.requirePerm(ctx, "iam.users.write", true)
	if err != nil {
		return nil, nil, err
	}
	if err := validateCreateUserParams(in.user); err != nil {
		return nil, nil, err
	}
	if a.partyProvisioner == nil {
		return nil, nil, middleware.ErrUnavailable("enterprise foundation service unavailable", nil)
	}

	normPartyID, normContactID, prepErr := a.partyProvisioner.PreparePortalUserProvisioning(ctx, in.partyID, in.contactID, requiredRole)
	if prepErr != nil {
		return nil, nil, prepErr
	}

	user, createErr := a.createUserByKindWithSession(ctx, sess, kind, in.user, true)
	if createErr != nil {
		return nil, nil, createErr
	}
	if user == nil || strings.TrimSpace(user.RawID) == "" {
		return nil, nil, middleware.NewGatewayError(middleware.CodeInternalError, "internal server error", errors.New("created IAM user missing identifier"))
	}

	linkReq := PartyUserLinkRequest{
		PartyID:          normPartyID,
		ContactID:        normContactID,
		RequiredRole:     requiredRole,
		FullName:         strings.TrimSpace(in.user.fullName),
		Email:            strings.TrimSpace(in.user.email),
		ContactRole:      in.contactRole,
		JobTitle:         in.jobTitle,
		Phone:            in.phone,
		IsPrimaryContact: in.isPrimaryContact,
	}
	contact, linkErr := a.partyProvisioner.LinkPortalUser(ctx, linkReq, user.RawID)
	if linkErr != nil {
		return nil, nil, a.rollbackUnlinkedUser(ctx, sess, user.RawID, linkErr)
	}
	return user, contact, nil
}

func validateCreateUserParams(p createUserParams) error {
	displayName := strings.TrimSpace(p.fullName)
	email := strings.TrimSpace(p.email)
	username := strings.TrimSpace(p.username)
	if displayName == "" || email == "" || username == "" || p.password == "" {
		return middleware.ErrInvalidArgument("full_name, email, username, and password are required", nil)
	}
	return nil
}

func (a *Adapter) createUserByKindWithSession(
	ctx context.Context,
	sess *middleware.ResolvedSession,
	kind model.UserKind,
	p createUserParams,
	defaultReqPwdChange bool,
) (*model.User, error) {
	if err := validateCreateUserParams(p); err != nil {
		return nil, err
	}
	displayName := strings.TrimSpace(p.fullName)
	email := strings.TrimSpace(p.email)
	username := strings.TrimSpace(p.username)

	resp, rpcErr := callIAMUsers(ctx, a, &sess.Caller, func(rpcCtx context.Context, users iamv1.UserServiceClient) (*iamv1.CreateUserResponse, error) {
		return users.CreateUser(rpcCtx, &iamv1.CreateUserRequest{
			Kind:                   toProtoUserKind(kind),
			Status:                 iamv1.UserStatus_USER_STATUS_PENDING,
			DisplayName:            displayName,
			Email:                  email,
			Username:               username,
			Password:               p.password,
			RequirePasswordChange:  defaultReqPwdChange,
			MarkIdentitiesVerified: false,
			ProvisioningRequestId:  p.provisioningRequestID,
			WorkerId:               p.workerID,
		})
	})
	if rpcErr != nil {
		return nil, rpcErr
	}
	return toModelUser(resp.GetUser()), nil
}

func (a *Adapter) rollbackUnlinkedUser(
	ctx context.Context,
	sess *middleware.ResolvedSession,
	userID string,
	linkErr error,
) error {
	if a == nil || sess == nil || strings.TrimSpace(userID) == "" {
		return linkErr
	}
	compCtx := context.Background()
	if ctx != nil {
		compCtx = context.WithoutCancel(ctx)
	}
	_, compErr := callIAMUsers(compCtx, a, &sess.Caller, func(rpcCtx context.Context, users iamv1.UserServiceClient) (*iamv1.SoftDeleteUserResponse, error) {
		return users.SoftDeleteUser(rpcCtx, &iamv1.SoftDeleteUserRequest{UserId: userID})
	})
	if compErr == nil {
		return linkErr
	}
	joined := errors.Join(linkErr, compErr)
	if gwErr, ok := middleware.AsGatewayError(linkErr); ok {
		return middleware.NewGatewayError(gwErr.Code, gwErr.Message, joined)
	}
	return middleware.ErrUnavailable("failed to link partner user cleanly", joined)
}

func parseStrictUUIDOrTypedID(raw string, expected GIDResource) (string, error) {
	cleanID, err := ParseTypedID(raw, expected)
	if err != nil {
		return "", err
	}
	parsed, parseErr := uuid.Parse(cleanID)
	if parseErr != nil || parsed == uuid.Nil {
		return "", middleware.ErrInvalidArgument("identifier must be a valid UUID or Global ID", parseErr)
	}
	return parsed.String(), nil
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
	callCtx := authcontext.WithCaller(ctx, sess.Caller)

	if err := a.ensureCanTouchTargetUser(callCtx, sess, cleanUserID, allowSelf); err != nil {
		return nil, err
	}
	user, rpcErr := call(callCtx, cleanUserID)
	if rpcErr != nil {
		return nil, rpcErr
	}
	return toModelUser(user), nil
}

// ActivateUser mengaktifkan status akun pengguna tertentu (wajib iam.users.write + reauth sensitif).
func (a *Adapter) ActivateUser(ctx context.Context, userID string) (*model.ActivateUserPayload, error) {
	user, err := a.mutateUserStatusByID(ctx, userID, true, func(callCtx context.Context, id string) (*iamv1.User, error) {
		resp, rpcErr := callIAMUsers(callCtx, a, nil, func(rpcCtx context.Context, users iamv1.UserServiceClient) (*iamv1.ActivateUserResponse, error) {
			return users.ActivateUser(rpcCtx, &iamv1.ActivateUserRequest{UserId: id})
		})
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
	user, err := a.mutateUserStatusByID(ctx, userID, false, func(callCtx context.Context, id string) (*iamv1.User, error) {
		resp, rpcErr := callIAMUsers(callCtx, a, nil, func(rpcCtx context.Context, users iamv1.UserServiceClient) (*iamv1.SuspendUserResponse, error) {
			return users.SuspendUser(rpcCtx, &iamv1.SuspendUserRequest{UserId: id})
		})
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
	user, err := a.mutateUserStatusByID(ctx, userID, false, func(callCtx context.Context, id string) (*iamv1.User, error) {
		resp, rpcErr := callIAMUsers(callCtx, a, nil, func(rpcCtx context.Context, users iamv1.UserServiceClient) (*iamv1.DisableUserResponse, error) {
			return users.DisableUser(rpcCtx, &iamv1.DisableUserRequest{UserId: id})
		})
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
	callCtx := authcontext.WithCaller(ctx, sess.Caller)

	if err := a.ensureCanTouchTargetUser(callCtx, sess, cleanUserID, false); err != nil {
		return fail(err)
	}
	_, rpcErr := callIAMUsers(callCtx, a, nil, func(rpcCtx context.Context, users iamv1.UserServiceClient) (*iamv1.SoftDeleteUserResponse, error) {
		return users.SoftDeleteUser(rpcCtx, &iamv1.SoftDeleteUserRequest{UserId: cleanUserID})
	})
	if rpcErr != nil {
		return fail(rpcErr)
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
	callCtx := authcontext.WithCaller(ctx, sess.Caller)

	if err := a.ensureCanTouchTargetUser(callCtx, sess, cleanUserID, true); err != nil {
		return fail(err)
	}
	resp, rpcErr := callIAMUsers(callCtx, a, nil, func(rpcCtx context.Context, users iamv1.UserServiceClient) (*iamv1.SetPrimaryUserIdentityResponse, error) {
		return users.SetPrimaryUserIdentity(rpcCtx, &iamv1.SetPrimaryUserIdentityRequest{
			UserId:     cleanUserID,
			IdentityId: cleanIdentityID,
		})
	})
	if rpcErr != nil {
		return fail(rpcErr)
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
	callCtx := authcontext.WithCaller(ctx, sess.Caller)

	if err := a.ensureCanTouchTargetUser(callCtx, sess, cleanUserID, true); err != nil {
		return fail(err)
	}
	_, rpcErr := callIAMUsers(callCtx, a, nil, func(rpcCtx context.Context, users iamv1.UserServiceClient) (*iamv1.RemoveUserIdentityResponse, error) {
		return users.RemoveUserIdentity(rpcCtx, &iamv1.RemoveUserIdentityRequest{
			UserId:     cleanUserID,
			IdentityId: cleanIdentityID,
		})
	})
	if rpcErr != nil {
		return fail(rpcErr)
	}
	return &model.RemoveUserIdentityPayload{Removed: ptrBool(true), Errors: noMutationErrors(), WasSuccessful: true}, nil
}

// CheckUsername memeriksa ketersediaan username dengan pembatasan laju per IP dan tanpa mengekspos PII atau jenis akun.
func (a *Adapter) CheckUsername(ctx context.Context, username string) (*model.UsernameAvailabilityPayload, error) {
	if err := a.ensureAuthDeps(); err != nil {
		return nil, err
	}
	if err := a.checkUsernameRateLimit(ctx); err != nil {
		return nil, err
	}
	caller := a.authManager.BuildCallerContext(ctx)
	resp, rpcErr := callIAMUsers(ctx, a, &caller, func(rpcCtx context.Context, users iamv1.UserServiceClient) (*iamv1.CheckUsernameAvailabilityResponse, error) {
		return users.CheckUsernameAvailability(rpcCtx, &iamv1.CheckUsernameAvailabilityRequest{
			Username: username,
		})
	})
	if rpcErr != nil {
		return nil, rpcErr
	}
	return &model.UsernameAvailabilityPayload{
		Available:          resp.GetAvailable(),
		NormalizedUsername: resp.GetNormalizedUsername(),
		Message:            resp.GetReason(),
	}, nil
}
