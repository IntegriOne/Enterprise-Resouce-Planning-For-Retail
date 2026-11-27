package iam

import (
	"context"
	"strings"

	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	iamv1 "github.com/integrione/erp-retail/backend/services/iam/contract/gen/iam/v1"
	"github.com/integrione/erp-retail/backend/shared/authcontext"
	"github.com/integrione/erp-retail/backend/shared/authz"
)

// Node mengambil entitas polimorfik Relay lintas domain IAM berdasarkan Global Object ID sambil menegakkan RBAC per sumber daya.
func (a *Adapter) Node(ctx context.Context, id string) (model.Node, error) {
	if err := a.ensureAuthorizerDeps(); err != nil {
		return nil, err
	}
	sess, err := a.authManager.RequireAuthenticated(ctx)
	if err != nil {
		return nil, err
	}
	res, rawID, err := DecodeNodeGID(id)
	if err != nil {
		return nil, err
	}

	switch res {
	case ResourceUser:
		return a.UserForNested(ctx, rawID)
	case ResourceRole:
		return a.Role(ctx, rawID)
	case ResourcePermission:
		return a.Permission(ctx, rawID)
	case ResourceGroup:
		return a.Group(ctx, rawID)
	case ResourceUserIdentity:
		if sess.Token == nil {
			return nil, middleware.ErrUnauthenticated("authentication required", nil)
		}
		if _, pErr := a.requireSelfOrPerm(ctx, sess.Token.UserID, authz.PermUsersReadSelf, authz.PermUsersRead); pErr != nil {
			return nil, pErr
		}
		resp, rpcErr := callIAMUsers(ctx, a, &sess.Caller, func(rpcCtx context.Context, users iamv1.UserServiceClient) (*iamv1.ListUserIdentitiesResponse, error) {
			return users.ListUserIdentities(rpcCtx, &iamv1.ListUserIdentitiesRequest{UserId: sess.Token.UserID})
		})
		if rpcErr != nil {
			return nil, rpcErr
		}
		for _, item := range resp.GetIdentities() {
			if item.GetId() == rawID {
				return toModelUserIdentity(item), nil
			}
		}
		// Pastikan admin tanpa kepemilikan langsung tetap diperiksa izin iam.users.read sebelum mengembalikan NotFound.
		if _, adminErr := a.requirePerm(ctx, authz.PermUsersRead, false); adminErr != nil {
			return nil, adminErr
		}
		return nil, middleware.ErrNotFound("user identity node not found", nil)
	case ResourceSession:
		if sess.Token == nil {
			return nil, middleware.ErrUnauthenticated("authentication required", nil)
		}
		if _, pErr := a.requireSelfOrPerm(ctx, sess.Token.UserID, authz.PermSessionsReadSelf, authz.PermSessionsRead); pErr != nil {
			return nil, pErr
		}
		ownedSess, findErr := a.findOwnedUserSession(authcontext.WithCaller(ctx, sess.Caller), sess.Token.UserID, rawID)
		if findErr != nil {
			return nil, findErr
		}
		if ownedSess != nil {
			return toModelSession(ownedSess), nil
		}
		if _, adminErr := a.requirePerm(ctx, authz.PermSessionsRead, false); adminErr != nil {
			return nil, adminErr
		}
		return nil, middleware.ErrNotFound("session node not found", nil)
	default:
		return nil, middleware.ErrInvalidArgument("unsupported node resource", nil)
	}
}

// Roles mempaginasi daftar peran RBAC dalam format Relay Connection (wajib izin iam.roles.read).
func (a *Adapter) Roles(
	ctx context.Context,
	after *string,
	before *string,
	first *int,
	last *int,
	query *string,
	managementDomain *string,
	activeOnly *bool,
) (*model.RoleConnection, error) {
	sess, err := a.requirePerm(ctx, "iam.roles.read", false)
	if err != nil {
		return nil, err
	}
	win, err := ParseRelayPage(after, before, first, last)
	if err != nil {
		return nil, err
	}

	items, startOffset, pageInfo, total, err := FetchRelayPage(win, func(limit, offset int32) ([]*iamv1.Role, *iamv1.PageInfo, error) {
		resp, rpcErr := callIAMRoles(ctx, a, &sess.Caller, func(rpcCtx context.Context, roles iamv1.RoleServiceClient) (*iamv1.ListRolesResponse, error) {
			return roles.ListRoles(rpcCtx, &iamv1.ListRolesRequest{
				Query:            derefStr(query),
				ManagementDomain: derefStr(managementDomain),
				ActiveOnly:       derefBool(activeOnly, false),
				Page:             &iamv1.PageRequest{Limit: limit, Offset: offset},
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
	return buildRoleConnection(items, startOffset, pageInfo, total), nil
}

// Role mengambil satu peran RBAC berdasarkan GID atau UUID mentah (wajib izin iam.roles.read).
func (a *Adapter) Role(ctx context.Context, id string) (*model.Role, error) {
	sess, err := a.requirePerm(ctx, "iam.roles.read", false)
	if err != nil {
		return nil, err
	}
	cleanID, err := ParseTypedID(id, ResourceRole)
	if err != nil {
		return nil, err
	}
	resp, rpcErr := callIAMRoles(ctx, a, &sess.Caller, func(rpcCtx context.Context, roles iamv1.RoleServiceClient) (*iamv1.GetRoleByIdResponse, error) {
		return roles.GetRoleById(rpcCtx, &iamv1.GetRoleByIdRequest{Id: cleanID})
	})
	if rpcErr != nil {
		return nil, rpcErr
	}
	return toModelRole(resp.GetRole()), nil
}

// RoleByCode mengambil satu peran RBAC berdasarkan kode kanoniknya (wajib izin iam.roles.read).
func (a *Adapter) RoleByCode(ctx context.Context, code string) (*model.Role, error) {
	sess, err := a.requirePerm(ctx, "iam.roles.read", false)
	if err != nil {
		return nil, err
	}
	cleanCode := strings.TrimSpace(code)
	if cleanCode == "" {
		return nil, middleware.ErrInvalidArgument("role code is required", nil)
	}
	resp, rpcErr := callIAMRoles(ctx, a, &sess.Caller, func(rpcCtx context.Context, roles iamv1.RoleServiceClient) (*iamv1.GetRoleByCodeResponse, error) {
		return roles.GetRoleByCode(rpcCtx, &iamv1.GetRoleByCodeRequest{Code: cleanCode})
	})
	if rpcErr != nil {
		return nil, rpcErr
	}
	return toModelRole(resp.GetRole()), nil
}

// RolePermissions mengambil daftar izin atomik pada suatu peran dalam format Relay Connection (wajib izin iam.roles.read).
func (a *Adapter) RolePermissions(
	ctx context.Context,
	roleID string,
	after *string,
	before *string,
	first *int,
	last *int,
) (*model.RolePermissionConnection, error) {
	sess, err := a.requirePerm(ctx, "iam.roles.read", false)
	if err != nil {
		return nil, err
	}
	cleanRoleID, err := ParseTypedID(roleID, ResourceRole)
	if err != nil {
		return nil, err
	}
	resp, rpcErr := callIAMRoles(ctx, a, &sess.Caller, func(rpcCtx context.Context, roles iamv1.RoleServiceClient) (*iamv1.ListRolePermissionsResponse, error) {
		return roles.ListRolePermissions(rpcCtx, &iamv1.ListRolePermissionsRequest{RoleId: cleanRoleID})
	})
	if rpcErr != nil {
		return nil, rpcErr
	}
	sliced, startOffset, pageInfo, total, err := SliceRelayWindow(resp.GetPermissions(), after, before, first, last)
	if err != nil {
		return nil, err
	}
	return buildRolePermissionConnection(cleanRoleID, sliced, startOffset, pageInfo, total), nil
}

// Permissions mencari dan mempaginasi katalog izin RBAC dalam format Relay Connection (wajib izin iam.permissions.read).
func (a *Adapter) Permissions(
	ctx context.Context,
	after *string,
	before *string,
	first *int,
	last *int,
	query *string,
	resource *string,
	isSensitive *bool,
) (*model.PermissionConnection, error) {
	sess, err := a.requirePerm(ctx, "iam.permissions.read", false)
	if err != nil {
		return nil, err
	}
	win, err := ParseRelayPage(after, before, first, last)
	if err != nil {
		return nil, err
	}

	items, startOffset, pageInfo, total, err := FetchRelayPage(win, func(limit, offset int32) ([]*iamv1.Permission, *iamv1.PageInfo, error) {
		resp, rpcErr := callIAMPermission(ctx, a, &sess.Caller, func(rpcCtx context.Context, perm iamv1.PermissionServiceClient) (*iamv1.ListPermissionsResponse, error) {
			return perm.ListPermissions(rpcCtx, &iamv1.ListPermissionsRequest{
				Query:       derefStr(query),
				Resource:    derefStr(resource),
				IsSensitive: isSensitive,
				Page:        &iamv1.PageRequest{Limit: limit, Offset: offset},
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
	return buildPermissionConnection(items, startOffset, pageInfo, total), nil
}

// Permission mengambil satu entri izin katalog berdasarkan GID atau UUID mentah (wajib izin iam.permissions.read).
func (a *Adapter) Permission(ctx context.Context, id string) (*model.Permission, error) {
	sess, err := a.requirePerm(ctx, "iam.permissions.read", false)
	if err != nil {
		return nil, err
	}
	cleanID, err := ParseTypedID(id, ResourcePermission)
	if err != nil {
		return nil, err
	}
	resp, rpcErr := callIAMPermission(ctx, a, &sess.Caller, func(rpcCtx context.Context, perm iamv1.PermissionServiceClient) (*iamv1.GetPermissionByIdResponse, error) {
		return perm.GetPermissionById(rpcCtx, &iamv1.GetPermissionByIdRequest{Id: cleanID})
	})
	if rpcErr != nil {
		return nil, rpcErr
	}
	return toModelPermission(resp.GetPermission()), nil
}

// Groups mempaginasi daftar grup pengguna dalam format Relay Connection (wajib izin iam.groups.read).
func (a *Adapter) Groups(
	ctx context.Context,
	after *string,
	before *string,
	first *int,
	last *int,
	query *string,
	activeOnly *bool,
) (*model.GroupConnection, error) {
	sess, err := a.requirePerm(ctx, "iam.groups.read", false)
	if err != nil {
		return nil, err
	}
	win, err := ParseRelayPage(after, before, first, last)
	if err != nil {
		return nil, err
	}

	items, startOffset, pageInfo, total, err := FetchRelayPage(win, func(limit, offset int32) ([]*iamv1.Group, *iamv1.PageInfo, error) {
		resp, rpcErr := callIAMGroups(ctx, a, &sess.Caller, func(rpcCtx context.Context, groups iamv1.GroupServiceClient) (*iamv1.ListGroupsResponse, error) {
			return groups.ListGroups(rpcCtx, &iamv1.ListGroupsRequest{
				Query:      derefStr(query),
				ActiveOnly: derefBool(activeOnly, false),
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
	return buildGroupConnection(items, startOffset, pageInfo, total), nil
}

// Group mengambil detail satu grup pengguna berdasarkan GID atau UUID mentah (wajib izin iam.groups.read).
func (a *Adapter) Group(ctx context.Context, id string) (*model.Group, error) {
	sess, err := a.requirePerm(ctx, "iam.groups.read", false)
	if err != nil {
		return nil, err
	}
	cleanID, err := ParseTypedID(id, ResourceGroup)
	if err != nil {
		return nil, err
	}
	resp, rpcErr := callIAMGroups(ctx, a, &sess.Caller, func(rpcCtx context.Context, groups iamv1.GroupServiceClient) (*iamv1.GetGroupByIdResponse, error) {
		return groups.GetGroupById(rpcCtx, &iamv1.GetGroupByIdRequest{Id: cleanID})
	})
	if rpcErr != nil {
		return nil, rpcErr
	}
	return toModelGroup(resp.GetGroup()), nil
}

// GroupMembers mengambil daftar anggota pengguna di dalam satu grup dalam format Relay Connection (wajib izin iam.groups.read).
func (a *Adapter) GroupMembers(
	ctx context.Context,
	groupID string,
	after *string,
	before *string,
	first *int,
	last *int,
) (*model.GroupMemberConnection, error) {
	sess, err := a.requirePerm(ctx, "iam.groups.read", false)
	if err != nil {
		return nil, err
	}
	cleanGroupID, err := ParseTypedID(groupID, ResourceGroup)
	if err != nil {
		return nil, err
	}
	win, err := ParseRelayPage(after, before, first, last)
	if err != nil {
		return nil, err
	}

	items, startOffset, pageInfo, total, err := FetchRelayPage(win, func(limit, offset int32) ([]*iamv1.User, *iamv1.PageInfo, error) {
		resp, rpcErr := callIAMGroups(ctx, a, &sess.Caller, func(rpcCtx context.Context, groups iamv1.GroupServiceClient) (*iamv1.ListGroupMembersResponse, error) {
			return groups.ListGroupMembers(rpcCtx, &iamv1.ListGroupMembersRequest{
				GroupId: cleanGroupID,
				Page:    &iamv1.PageRequest{Limit: limit, Offset: offset},
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
	return buildGroupMemberConnection(cleanGroupID, items, startOffset, pageInfo, total), nil
}

// GroupRoles mengambil daftar peran yang terpasang pada suatu grup dalam format Relay Connection (wajib izin iam.groups.read).
func (a *Adapter) GroupRoles(
	ctx context.Context,
	groupID string,
	includeExpired *bool,
	after *string,
	before *string,
	first *int,
	last *int,
) (*model.GroupRoleConnection, error) {
	sess, err := a.requirePerm(ctx, "iam.groups.read", false)
	if err != nil {
		return nil, err
	}
	cleanGroupID, err := ParseTypedID(groupID, ResourceGroup)
	if err != nil {
		return nil, err
	}
	callCtx := authcontext.WithCaller(ctx, sess.Caller)

	resp, rpcErr := callIAMGroups(callCtx, a, nil, func(rpcCtx context.Context, groups iamv1.GroupServiceClient) (*iamv1.ListGroupRolesResponse, error) {
		return groups.ListGroupRoles(rpcCtx, &iamv1.ListGroupRolesRequest{
			GroupId:        cleanGroupID,
			IncludeExpired: derefBool(includeExpired, false),
		})
	})
	if rpcErr != nil {
		return nil, rpcErr
	}
	sliced, startOffset, pageInfo, total, err := SliceRelayWindow(resp.GetGroupRoles(), after, before, first, last)
	if err != nil {
		return nil, err
	}

	roleIDs := make([]string, 0, len(sliced))
	for _, gr := range sliced {
		roleIDs = append(roleIDs, gr.GetRoleId())
	}
	roleMap, err := a.fetchRoleMap(callCtx, roleIDs)
	if err != nil {
		return nil, err
	}

	edges := make([]*model.GroupRoleEdge, 0, len(sliced))
	nodes := make([]*model.Role, 0, len(sliced))
	for i, gr := range sliced {
		rNode := roleMap[gr.GetRoleId()]
		edges = append(edges, toModelGroupRoleEdge(gr, rNode, EncodeCursor(startOffset+i)))
		if rNode != nil {
			nodes = append(nodes, rNode)
		}
	}
	return &model.GroupRoleConnection{
		Edges:      edges,
		Nodes:      nodes,
		PageInfo:   pageInfo,
		TotalCount: total,
	}, nil
}

// CreateRole membuat peran RBAC baru (wajib izin iam.roles.write + reauth sensitif).
func (a *Adapter) mutateRoleByID(ctx context.Context, id string, call func(context.Context, string) (*iamv1.Role, error)) (*model.Role, error) {
	sess, err := a.requirePerm(ctx, "iam.roles.write", true)
	if err != nil {
		return nil, err
	}
	cleanID, err := ParseTypedID(id, ResourceRole)
	if err != nil {
		return nil, err
	}
	callCtx := authcontext.WithCaller(ctx, sess.Caller)

	if err := a.ensureCanManageRole(callCtx, sess, cleanID); err != nil {
		return nil, err
	}
	role, rpcErr := call(callCtx, cleanID)
	if rpcErr != nil {
		return nil, rpcErr
	}
	return toModelRole(role), nil
}

func (a *Adapter) mutateGroupByID(ctx context.Context, id string, call func(context.Context, string) (*iamv1.Group, error)) (*model.Group, error) {
	sess, err := a.requirePerm(ctx, "iam.groups.write", true)
	if err != nil {
		return nil, err
	}
	cleanID, err := ParseTypedID(id, ResourceGroup)
	if err != nil {
		return nil, err
	}
	callCtx := authcontext.WithCaller(ctx, sess.Caller)

	if err := a.ensureCanManageGroup(callCtx, sess, cleanID); err != nil {
		return nil, err
	}
	group, rpcErr := call(callCtx, cleanID)
	if rpcErr != nil {
		return nil, rpcErr
	}
	return toModelGroup(group), nil
}

func (a *Adapter) CreateRole(ctx context.Context, input model.CreateRoleInput) (*model.CreateRolePayload, error) {
	fail := func(err error) (*model.CreateRolePayload, error) {
		return &model.CreateRolePayload{Errors: ToMutationErrors(err), WasSuccessful: false}, nil
	}
	sess, err := a.requirePerm(ctx, "iam.roles.write", true)
	if err != nil {
		return fail(err)
	}
	if strings.TrimSpace(input.Code) == "" || strings.TrimSpace(input.Name) == "" {
		return fail(middleware.ErrInvalidArgument("code and name are required", nil))
	}
	if input.IsSystem != nil && *input.IsSystem {
		return fail(middleware.ErrForbidden("system roles cannot be created via API", nil))
	}

	resp, rpcErr := callIAMRoles(ctx, a, &sess.Caller, func(rpcCtx context.Context, roles iamv1.RoleServiceClient) (*iamv1.CreateRoleResponse, error) {
		return roles.CreateRole(rpcCtx, &iamv1.CreateRoleRequest{
			Code:                strings.TrimSpace(input.Code),
			Name:                strings.TrimSpace(input.Name),
			Description:         derefStr(input.Description),
			ManagementDomain:    strings.TrimSpace(input.ManagementDomain),
			DefaultRedirectPath: strings.TrimSpace(input.DefaultRedirectPath),
			RedirectPriority:    derefInt32(input.RedirectPriority, 50),
			IsSystem:            false,
		})
	})
	if rpcErr != nil {
		return fail(rpcErr)
	}
	return &model.CreateRolePayload{Role: toModelRole(resp.GetRole()), Errors: noMutationErrors(), WasSuccessful: true}, nil
}

// UpdateRoleDetails mengubah atribut nama, domain, dan rute redirect peran RBAC (wajib iam.roles.write + reauth sensitif).
func (a *Adapter) UpdateRoleDetails(ctx context.Context, input model.UpdateRoleDetailsInput) (*model.UpdateRoleDetailsPayload, error) {
	fail := func(err error) (*model.UpdateRoleDetailsPayload, error) {
		return &model.UpdateRoleDetailsPayload{Errors: ToMutationErrors(err), WasSuccessful: false}, nil
	}
	sess, err := a.requirePerm(ctx, "iam.roles.write", true)
	if err != nil {
		return fail(err)
	}
	cleanID, err := ParseTypedID(input.ID, ResourceRole)
	if err != nil {
		return fail(err)
	}
	if strings.TrimSpace(input.Name) == "" {
		return fail(middleware.ErrInvalidArgument("name is required", nil))
	}
	callCtx := authcontext.WithCaller(ctx, sess.Caller)

	if err := a.ensureCanManageRole(callCtx, sess, cleanID); err != nil {
		return fail(err)
	}
	resp, rpcErr := callIAMRoles(callCtx, a, nil, func(rpcCtx context.Context, roles iamv1.RoleServiceClient) (*iamv1.UpdateRoleDetailsResponse, error) {
		return roles.UpdateRoleDetails(rpcCtx, &iamv1.UpdateRoleDetailsRequest{
			Id:                  cleanID,
			Name:                strings.TrimSpace(input.Name),
			Description:         derefStr(input.Description),
			ManagementDomain:    strings.TrimSpace(input.ManagementDomain),
			DefaultRedirectPath: strings.TrimSpace(input.DefaultRedirectPath),
			RedirectPriority:    derefInt32(input.RedirectPriority, 50),
		})
	})
	if rpcErr != nil {
		return fail(rpcErr)
	}
	return &model.UpdateRoleDetailsPayload{Role: toModelRole(resp.GetRole()), Errors: noMutationErrors(), WasSuccessful: true}, nil
}

// ActivateRole mengaktifkan peran RBAC tertentu (wajib iam.roles.write + reauth sensitif).
func (a *Adapter) ActivateRole(ctx context.Context, id string) (*model.ActivateRolePayload, error) {
	role, err := a.mutateRoleByID(ctx, id, func(callCtx context.Context, roleID string) (*iamv1.Role, error) {
		resp, rpcErr := callIAMRoles(callCtx, a, nil, func(rpcCtx context.Context, roles iamv1.RoleServiceClient) (*iamv1.ActivateRoleResponse, error) {
			return roles.ActivateRole(rpcCtx, &iamv1.ActivateRoleRequest{Id: roleID})
		})
		if rpcErr != nil {
			return nil, rpcErr
		}
		return resp.GetRole(), nil
	})
	if err != nil {
		return &model.ActivateRolePayload{Errors: ToMutationErrors(err), WasSuccessful: false}, nil
	}
	return &model.ActivateRolePayload{Role: role, Errors: noMutationErrors(), WasSuccessful: true}, nil
}

// DeactivateRole menonaktifkan peran RBAC tertentu (wajib iam.roles.write + reauth sensitif).
func (a *Adapter) DeactivateRole(ctx context.Context, id string) (*model.DeactivateRolePayload, error) {
	role, err := a.mutateRoleByID(ctx, id, func(callCtx context.Context, roleID string) (*iamv1.Role, error) {
		resp, rpcErr := callIAMRoles(callCtx, a, nil, func(rpcCtx context.Context, roles iamv1.RoleServiceClient) (*iamv1.DeactivateRoleResponse, error) {
			return roles.DeactivateRole(rpcCtx, &iamv1.DeactivateRoleRequest{Id: roleID})
		})
		if rpcErr != nil {
			return nil, rpcErr
		}
		return resp.GetRole(), nil
	})
	if err != nil {
		return &model.DeactivateRolePayload{Errors: ToMutationErrors(err), WasSuccessful: false}, nil
	}
	return &model.DeactivateRolePayload{Role: role, Errors: noMutationErrors(), WasSuccessful: true}, nil
}

// DeleteRole menghapus peran RBAC non-sistem (wajib iam.roles.write + reauth sensitif).
func (a *Adapter) DeleteRole(ctx context.Context, id string) (*model.DeleteRolePayload, error) {
	_, err := a.mutateRoleByID(ctx, id, func(callCtx context.Context, roleID string) (*iamv1.Role, error) {
		_, rpcErr := callIAMRoles(callCtx, a, nil, func(rpcCtx context.Context, roles iamv1.RoleServiceClient) (*iamv1.DeleteRoleResponse, error) {
			return roles.DeleteRole(rpcCtx, &iamv1.DeleteRoleRequest{Id: roleID})
		})
		return nil, rpcErr
	})
	if err != nil {
		return &model.DeleteRolePayload{Errors: ToMutationErrors(err), WasSuccessful: false}, nil
	}
	return &model.DeleteRolePayload{Deleted: ptrBool(true), Errors: noMutationErrors(), WasSuccessful: true}, nil
}

// AssignRolePermission memasang izin atomik ke peran RBAC dan mengembalikan RolePermissionEdge beserta node Permission.
func (a *Adapter) AssignRolePermission(ctx context.Context, roleID, permissionID string) (*model.AssignRolePermissionPayload, error) {
	fail := func(err error) (*model.AssignRolePermissionPayload, error) {
		return &model.AssignRolePermissionPayload{Errors: ToMutationErrors(err), WasSuccessful: false}, nil
	}
	sess, err := a.requirePerm(ctx, "iam.roles.write", true)
	if err != nil {
		return fail(err)
	}
	cleanRoleID, err := ParseTypedID(roleID, ResourceRole)
	if err != nil {
		return fail(err)
	}
	cleanPermID, err := ParseTypedID(permissionID, ResourcePermission)
	if err != nil {
		return fail(err)
	}
	callCtx := authcontext.WithCaller(ctx, sess.Caller)

	if err := a.ensureCanManageRole(callCtx, sess, cleanRoleID); err != nil {
		return fail(err)
	}
	if err := a.ensureCanManagePermissionID(callCtx, sess, cleanPermID); err != nil {
		return fail(err)
	}
	resp, rpcErr := callIAMRoles(callCtx, a, nil, func(rpcCtx context.Context, roles iamv1.RoleServiceClient) (*iamv1.AssignRolePermissionResponse, error) {
		return roles.AssignRolePermission(rpcCtx, &iamv1.AssignRolePermissionRequest{
			RoleId:       cleanRoleID,
			PermissionId: cleanPermID,
		})
	})
	if rpcErr != nil {
		return fail(rpcErr)
	}
	permResp, pErr := callIAMPermission(callCtx, a, nil, func(rpcCtx context.Context, perm iamv1.PermissionServiceClient) (*iamv1.GetPermissionByIdResponse, error) {
		return perm.GetPermissionById(rpcCtx, &iamv1.GetPermissionByIdRequest{Id: cleanPermID})
	})
	if pErr != nil {
		return fail(pErr)
	}
	edge := toModelRolePermissionEdge(resp.GetRolePermission(), toModelPermission(permResp.GetPermission()), EncodeCursor(0))
	return &model.AssignRolePermissionPayload{RolePermissionEdge: edge, Errors: noMutationErrors(), WasSuccessful: true}, nil
}

// RevokeRolePermission mencabut izin dari peran RBAC (wajib iam.roles.write + reauth sensitif).
func (a *Adapter) RevokeRolePermission(ctx context.Context, roleID, permissionID string) (*model.RevokeRolePermissionPayload, error) {
	fail := func(err error) (*model.RevokeRolePermissionPayload, error) {
		return &model.RevokeRolePermissionPayload{Errors: ToMutationErrors(err), WasSuccessful: false}, nil
	}
	sess, err := a.requirePerm(ctx, "iam.roles.write", true)
	if err != nil {
		return fail(err)
	}
	cleanRoleID, err := ParseTypedID(roleID, ResourceRole)
	if err != nil {
		return fail(err)
	}
	cleanPermID, err := ParseTypedID(permissionID, ResourcePermission)
	if err != nil {
		return fail(err)
	}
	callCtx := authcontext.WithCaller(ctx, sess.Caller)

	if err := a.ensureCanManageRole(callCtx, sess, cleanRoleID); err != nil {
		return fail(err)
	}
	if err := a.ensureCanManagePermissionID(callCtx, sess, cleanPermID); err != nil {
		return fail(err)
	}
	_, rpcErr := callIAMRoles(callCtx, a, nil, func(rpcCtx context.Context, roles iamv1.RoleServiceClient) (*iamv1.RevokeRolePermissionResponse, error) {
		return roles.RevokeRolePermission(rpcCtx, &iamv1.RevokeRolePermissionRequest{
			RoleId:       cleanRoleID,
			PermissionId: cleanPermID,
		})
	})
	if rpcErr != nil {
		return fail(rpcErr)
	}
	return &model.RevokeRolePermissionPayload{Revoked: ptrBool(true), Errors: noMutationErrors(), WasSuccessful: true}, nil
}

// AssignUserRole memberikan peran berscope kepada pengguna dan mengembalikan UserRoleEdge lengkap dengan node Role.
func (a *Adapter) AssignUserRole(ctx context.Context, input model.AssignUserRoleInput) (*model.AssignUserRolePayload, error) {
	fail := func(err error) (*model.AssignUserRolePayload, error) {
		return &model.AssignUserRolePayload{Errors: ToMutationErrors(err), WasSuccessful: false}, nil
	}
	sess, err := a.requirePerm(ctx, "iam.roles.write", true)
	if err != nil {
		return fail(err)
	}
	cleanUserID, err := ParseTypedID(input.UserID, ResourceUser)
	if err != nil {
		return fail(err)
	}
	cleanRoleID, err := ParseTypedID(input.RoleID, ResourceRole)
	if err != nil {
		return fail(err)
	}
	scopeID, err := ParseOptionalScopeID(input.ScopeID)
	if err != nil {
		return fail(err)
	}
	if input.ScopeType == model.ScopeTypeSelf && (scopeID == nil || !strings.EqualFold(strings.TrimSpace(*scopeID), cleanUserID)) {
		return fail(middleware.ErrInvalidArgument("self scope_id must match target user_id", nil))
	}
	grantedBy := sess.Token.UserID
	callCtx := authcontext.WithCaller(ctx, sess.Caller)

	if err := a.ensureCanManageRole(callCtx, sess, cleanRoleID); err != nil {
		return fail(err)
	}
	if err := a.ensureCanTouchTargetUser(callCtx, sess, cleanUserID, true); err != nil {
		return fail(err)
	}
	resp, rpcErr := callIAMRoles(callCtx, a, nil, func(rpcCtx context.Context, roles iamv1.RoleServiceClient) (*iamv1.AssignUserRoleResponse, error) {
		return roles.AssignUserRole(rpcCtx, &iamv1.AssignUserRoleRequest{
			UserId:    cleanUserID,
			RoleId:    cleanRoleID,
			ScopeType: toProtoScopeType(input.ScopeType),
			ScopeId:   scopeID,
			GrantedBy: &grantedBy,
			ExpiresAt: toProtoTimePtr(input.ExpiresAt),
		})
	})
	if rpcErr != nil {
		return fail(rpcErr)
	}
	roleResp, rErr := callIAMRoles(callCtx, a, nil, func(rpcCtx context.Context, roles iamv1.RoleServiceClient) (*iamv1.GetRoleByIdResponse, error) {
		return roles.GetRoleById(rpcCtx, &iamv1.GetRoleByIdRequest{Id: cleanRoleID})
	})
	if rErr != nil {
		return fail(rErr)
	}
	edge := toModelUserRoleEdge(resp.GetUserRole(), toModelRole(roleResp.GetRole()), EncodeCursor(0))
	return &model.AssignUserRolePayload{UserRoleEdge: edge, Errors: noMutationErrors(), WasSuccessful: true}, nil
}

// RevokeUserRole mencabut penugasan peran langsung pengguna (wajib iam.roles.write + reauth sensitif).
func (a *Adapter) RevokeUserRole(ctx context.Context, assignmentID string) (*model.RevokeUserRolePayload, error) {
	fail := func(err error) (*model.RevokeUserRolePayload, error) {
		return &model.RevokeUserRolePayload{Errors: ToMutationErrors(err), WasSuccessful: false}, nil
	}
	sess, err := a.requirePerm(ctx, "iam.roles.write", true)
	if err != nil {
		return fail(err)
	}
	cleanID, err := ParseFlexibleAssignmentID(assignmentID, ResourceUserRole)
	if err != nil {
		return fail(err)
	}
	if !callerIsSuperAdmin(sess) {
		return fail(middleware.ErrForbidden("only superadmin may revoke role assignments by assignment id", nil))
	}

	_, rpcErr := callIAMRoles(ctx, a, &sess.Caller, func(rpcCtx context.Context, roles iamv1.RoleServiceClient) (*iamv1.RevokeUserRoleResponse, error) {
		return roles.RevokeUserRole(rpcCtx, &iamv1.RevokeUserRoleRequest{AssignmentId: cleanID})
	})
	if rpcErr != nil {
		return fail(rpcErr)
	}
	return &model.RevokeUserRolePayload{Revoked: ptrBool(true), Errors: noMutationErrors(), WasSuccessful: true}, nil
}

// CreateGroup membuat grup pengguna baru (wajib iam.groups.write + reauth sensitif).
func (a *Adapter) CreateGroup(ctx context.Context, input model.CreateGroupInput) (*model.CreateGroupPayload, error) {
	fail := func(err error) (*model.CreateGroupPayload, error) {
		return &model.CreateGroupPayload{Errors: ToMutationErrors(err), WasSuccessful: false}, nil
	}
	sess, err := a.requirePerm(ctx, "iam.groups.write", true)
	if err != nil {
		return fail(err)
	}
	if strings.TrimSpace(input.Code) == "" || strings.TrimSpace(input.Name) == "" {
		return fail(middleware.ErrInvalidArgument("code and name are required", nil))
	}

	resp, rpcErr := callIAMGroups(ctx, a, &sess.Caller, func(rpcCtx context.Context, groups iamv1.GroupServiceClient) (*iamv1.CreateGroupResponse, error) {
		return groups.CreateGroup(rpcCtx, &iamv1.CreateGroupRequest{
			Code:        strings.TrimSpace(input.Code),
			Name:        strings.TrimSpace(input.Name),
			Description: derefStr(input.Description),
		})
	})
	if rpcErr != nil {
		return fail(rpcErr)
	}
	return &model.CreateGroupPayload{Group: toModelGroup(resp.GetGroup()), Errors: noMutationErrors(), WasSuccessful: true}, nil
}

// UpdateGroupDetails memperbarui nama dan deskripsi grup (wajib iam.groups.write + reauth sensitif).
func (a *Adapter) UpdateGroupDetails(ctx context.Context, input model.UpdateGroupDetailsInput) (*model.UpdateGroupDetailsPayload, error) {
	fail := func(err error) (*model.UpdateGroupDetailsPayload, error) {
		return &model.UpdateGroupDetailsPayload{Errors: ToMutationErrors(err), WasSuccessful: false}, nil
	}
	sess, err := a.requirePerm(ctx, "iam.groups.write", true)
	if err != nil {
		return fail(err)
	}
	cleanID, err := ParseTypedID(input.ID, ResourceGroup)
	if err != nil {
		return fail(err)
	}
	if strings.TrimSpace(input.Name) == "" {
		return fail(middleware.ErrInvalidArgument("name is required", nil))
	}

	resp, rpcErr := callIAMGroups(ctx, a, &sess.Caller, func(rpcCtx context.Context, groups iamv1.GroupServiceClient) (*iamv1.UpdateGroupDetailsResponse, error) {
		return groups.UpdateGroupDetails(rpcCtx, &iamv1.UpdateGroupDetailsRequest{
			Id:          cleanID,
			Name:        strings.TrimSpace(input.Name),
			Description: derefStr(input.Description),
		})
	})
	if rpcErr != nil {
		return fail(rpcErr)
	}
	return &model.UpdateGroupDetailsPayload{Group: toModelGroup(resp.GetGroup()), Errors: noMutationErrors(), WasSuccessful: true}, nil
}

// ActivateGroup mengaktifkan kembali sebuah grup (wajib iam.groups.write + reauth sensitif).
func (a *Adapter) ActivateGroup(ctx context.Context, id string) (*model.ActivateGroupPayload, error) {
	group, err := a.mutateGroupByID(ctx, id, func(callCtx context.Context, groupID string) (*iamv1.Group, error) {
		resp, rpcErr := callIAMGroups(callCtx, a, nil, func(rpcCtx context.Context, groups iamv1.GroupServiceClient) (*iamv1.ActivateGroupResponse, error) {
			return groups.ActivateGroup(rpcCtx, &iamv1.ActivateGroupRequest{Id: groupID})
		})
		if rpcErr != nil {
			return nil, rpcErr
		}
		return resp.GetGroup(), nil
	})
	if err != nil {
		return &model.ActivateGroupPayload{Errors: ToMutationErrors(err), WasSuccessful: false}, nil
	}
	return &model.ActivateGroupPayload{Group: group, Errors: noMutationErrors(), WasSuccessful: true}, nil
}

// DeactivateGroup menonaktifkan sebuah grup beserta pewarisan perannya (wajib iam.groups.write + reauth sensitif).
func (a *Adapter) DeactivateGroup(ctx context.Context, id string) (*model.DeactivateGroupPayload, error) {
	group, err := a.mutateGroupByID(ctx, id, func(callCtx context.Context, groupID string) (*iamv1.Group, error) {
		resp, rpcErr := callIAMGroups(callCtx, a, nil, func(rpcCtx context.Context, groups iamv1.GroupServiceClient) (*iamv1.DeactivateGroupResponse, error) {
			return groups.DeactivateGroup(rpcCtx, &iamv1.DeactivateGroupRequest{Id: groupID})
		})
		if rpcErr != nil {
			return nil, rpcErr
		}
		return resp.GetGroup(), nil
	})
	if err != nil {
		return &model.DeactivateGroupPayload{Errors: ToMutationErrors(err), WasSuccessful: false}, nil
	}
	return &model.DeactivateGroupPayload{Group: group, Errors: noMutationErrors(), WasSuccessful: true}, nil
}

// DeleteGroup menghapus grup pengguna (wajib iam.groups.write + reauth sensitif).
func (a *Adapter) DeleteGroup(ctx context.Context, id string) (*model.DeleteGroupPayload, error) {
	_, err := a.mutateGroupByID(ctx, id, func(callCtx context.Context, groupID string) (*iamv1.Group, error) {
		_, rpcErr := callIAMGroups(callCtx, a, nil, func(rpcCtx context.Context, groups iamv1.GroupServiceClient) (*iamv1.DeleteGroupResponse, error) {
			return groups.DeleteGroup(rpcCtx, &iamv1.DeleteGroupRequest{Id: groupID})
		})
		return nil, rpcErr
	})
	if err != nil {
		return &model.DeleteGroupPayload{Errors: ToMutationErrors(err), WasSuccessful: false}, nil
	}
	return &model.DeleteGroupPayload{Deleted: ptrBool(true), Errors: noMutationErrors(), WasSuccessful: true}, nil
}

// AddGroupMember menambahkan pengguna ke grup dan mengembalikan GroupMemberEdge lengkap dengan node User.
func (a *Adapter) AddGroupMember(ctx context.Context, groupID, userID string) (*model.AddGroupMemberPayload, error) {
	fail := func(err error) (*model.AddGroupMemberPayload, error) {
		return &model.AddGroupMemberPayload{Errors: ToMutationErrors(err), WasSuccessful: false}, nil
	}
	sess, err := a.requirePerm(ctx, "iam.groups.write", true)
	if err != nil {
		return fail(err)
	}
	cleanGroupID, err := ParseTypedID(groupID, ResourceGroup)
	if err != nil {
		return fail(err)
	}
	cleanUserID, err := ParseTypedID(userID, ResourceUser)
	if err != nil {
		return fail(err)
	}
	callCtx := authcontext.WithCaller(ctx, sess.Caller)

	if err := a.ensureCanTouchTargetUser(callCtx, sess, cleanUserID, true); err != nil {
		return fail(err)
	}
	if err := a.ensureCanManageGroup(callCtx, sess, cleanGroupID); err != nil {
		return fail(err)
	}
	resp, rpcErr := callIAMGroups(callCtx, a, nil, func(rpcCtx context.Context, groups iamv1.GroupServiceClient) (*iamv1.AddGroupMemberResponse, error) {
		return groups.AddGroupMember(rpcCtx, &iamv1.AddGroupMemberRequest{
			GroupId: cleanGroupID,
			UserId:  cleanUserID,
		})
	})
	if rpcErr != nil {
		return fail(rpcErr)
	}
	userResp, uErr := callIAMUsers(callCtx, a, nil, func(rpcCtx context.Context, users iamv1.UserServiceClient) (*iamv1.GetUserByIdResponse, error) {
		return users.GetUserById(rpcCtx, &iamv1.GetUserByIdRequest{UserId: cleanUserID})
	})
	if uErr != nil {
		return fail(uErr)
	}
	edge := toModelGroupMemberEdge(resp.GetMember(), toModelUser(userResp.GetUser()), EncodeCursor(0))
	return &model.AddGroupMemberPayload{GroupMemberEdge: edge, Errors: noMutationErrors(), WasSuccessful: true}, nil
}

// RemoveGroupMember mengeluarkan pengguna dari sebuah grup (wajib iam.groups.write + reauth sensitif).
func (a *Adapter) RemoveGroupMember(ctx context.Context, groupID, userID string) (*model.RemoveGroupMemberPayload, error) {
	fail := func(err error) (*model.RemoveGroupMemberPayload, error) {
		return &model.RemoveGroupMemberPayload{Errors: ToMutationErrors(err), WasSuccessful: false}, nil
	}
	sess, err := a.requirePerm(ctx, "iam.groups.write", true)
	if err != nil {
		return fail(err)
	}
	cleanGroupID, err := ParseTypedID(groupID, ResourceGroup)
	if err != nil {
		return fail(err)
	}
	cleanUserID, err := ParseTypedID(userID, ResourceUser)
	if err != nil {
		return fail(err)
	}
	callCtx := authcontext.WithCaller(ctx, sess.Caller)

	if err := a.ensureCanTouchTargetUser(callCtx, sess, cleanUserID, true); err != nil {
		return fail(err)
	}
	if err := a.ensureCanManageGroup(callCtx, sess, cleanGroupID); err != nil {
		return fail(err)
	}
	_, rpcErr := callIAMGroups(callCtx, a, nil, func(rpcCtx context.Context, groups iamv1.GroupServiceClient) (*iamv1.RemoveGroupMemberResponse, error) {
		return groups.RemoveGroupMember(rpcCtx, &iamv1.RemoveGroupMemberRequest{
			GroupId: cleanGroupID,
			UserId:  cleanUserID,
		})
	})
	if rpcErr != nil {
		return fail(rpcErr)
	}
	return &model.RemoveGroupMemberPayload{Removed: ptrBool(true), Errors: noMutationErrors(), WasSuccessful: true}, nil
}

// AssignGroupRole memasang peran berscope pada grup dan mengembalikan GroupRoleEdge lengkap dengan node Role.
func (a *Adapter) AssignGroupRole(ctx context.Context, input model.AssignGroupRoleInput) (*model.AssignGroupRolePayload, error) {
	fail := func(err error) (*model.AssignGroupRolePayload, error) {
		return &model.AssignGroupRolePayload{Errors: ToMutationErrors(err), WasSuccessful: false}, nil
	}
	sess, err := a.requirePerm(ctx, "iam.groups.write", true)
	if err != nil {
		return fail(err)
	}
	cleanGroupID, err := ParseTypedID(input.GroupID, ResourceGroup)
	if err != nil {
		return fail(err)
	}
	cleanRoleID, err := ParseTypedID(input.RoleID, ResourceRole)
	if err != nil {
		return fail(err)
	}
	scopeID, err := ParseOptionalScopeID(input.ScopeID)
	if err != nil {
		return fail(err)
	}
	if input.ScopeType == model.ScopeTypeSelf {
		return fail(middleware.ErrInvalidArgument("group role assignment cannot use self scope", nil))
	}
	grantedBy := sess.Token.UserID
	callCtx := authcontext.WithCaller(ctx, sess.Caller)

	if err := a.ensureCanManageRole(callCtx, sess, cleanRoleID); err != nil {
		return fail(err)
	}
	if err := a.ensureCanManageGroup(callCtx, sess, cleanGroupID); err != nil {
		return fail(err)
	}
	resp, rpcErr := callIAMGroups(callCtx, a, nil, func(rpcCtx context.Context, groups iamv1.GroupServiceClient) (*iamv1.AssignGroupRoleResponse, error) {
		return groups.AssignGroupRole(rpcCtx, &iamv1.AssignGroupRoleRequest{
			GroupId:   cleanGroupID,
			RoleId:    cleanRoleID,
			ScopeType: toProtoScopeType(input.ScopeType),
			ScopeId:   scopeID,
			GrantedBy: &grantedBy,
			ExpiresAt: toProtoTimePtr(input.ExpiresAt),
		})
	})
	if rpcErr != nil {
		return fail(rpcErr)
	}
	roleResp, rErr := callIAMRoles(callCtx, a, nil, func(rpcCtx context.Context, roles iamv1.RoleServiceClient) (*iamv1.GetRoleByIdResponse, error) {
		return roles.GetRoleById(rpcCtx, &iamv1.GetRoleByIdRequest{Id: cleanRoleID})
	})
	if rErr != nil {
		return fail(rErr)
	}
	edge := toModelGroupRoleEdge(resp.GetGroupRole(), toModelRole(roleResp.GetRole()), EncodeCursor(0))
	return &model.AssignGroupRolePayload{GroupRoleEdge: edge, Errors: noMutationErrors(), WasSuccessful: true}, nil
}

// RevokeGroupRole mencabut penugasan peran dari sebuah grup (wajib iam.groups.write + reauth sensitif).
func (a *Adapter) RevokeGroupRole(ctx context.Context, assignmentID string) (*model.RevokeGroupRolePayload, error) {
	fail := func(err error) (*model.RevokeGroupRolePayload, error) {
		return &model.RevokeGroupRolePayload{Errors: ToMutationErrors(err), WasSuccessful: false}, nil
	}
	sess, err := a.requirePerm(ctx, "iam.groups.write", true)
	if err != nil {
		return fail(err)
	}
	cleanID, err := ParseFlexibleAssignmentID(assignmentID, ResourceGroupRole)
	if err != nil {
		return fail(err)
	}
	if !callerIsSuperAdmin(sess) {
		return fail(middleware.ErrForbidden("only superadmin may revoke group role assignments by assignment id", nil))
	}

	_, rpcErr := callIAMGroups(ctx, a, &sess.Caller, func(rpcCtx context.Context, groups iamv1.GroupServiceClient) (*iamv1.RevokeGroupRoleResponse, error) {
		return groups.RevokeGroupRole(rpcCtx, &iamv1.RevokeGroupRoleRequest{AssignmentId: cleanID})
	})
	if rpcErr != nil {
		return fail(rpcErr)
	}
	return &model.RevokeGroupRolePayload{Revoked: ptrBool(true), Errors: noMutationErrors(), WasSuccessful: true}, nil
}

// RevokeSession mencabut satu sesi aktif milik pengguna tertentu (wajib iam.sessions.revoke + reauth sensitif).
func (a *Adapter) RevokeSession(ctx context.Context, sessionID string, reason *model.SessionEndReason) (*model.RevokeSessionPayload, error) {
	fail := func(err error) (*model.RevokeSessionPayload, error) {
		return &model.RevokeSessionPayload{Errors: ToMutationErrors(err), WasSuccessful: false}, nil
	}
	if err := a.ensureAuthorizerDeps(); err != nil {
		return fail(err)
	}
	sess, err := a.authManager.RequireAuthenticated(ctx)
	if err != nil {
		return fail(err)
	}
	cleanID, err := ParseTypedID(sessionID, ResourceSession)
	if err != nil {
		return fail(err)
	}

	var authorizedSess *middleware.ResolvedSession
	if sess.Snapshot != nil {
		if _, hasAdmin := sess.Snapshot.HasScopedPermission(authz.PermSessionsRevoke, authz.ScopeGrant{Type: authz.ScopeGlobal}); hasAdmin {
			authorizedSess, err = a.requirePerm(ctx, authz.PermSessionsRevoke, true)
			if err != nil {
				return fail(err)
			}
		} else if sess.Token != nil {
			if _, hasSelf := sess.Snapshot.HasScopedPermission(authz.PermSessionsDeleteSelf, authz.ScopeGrant{Type: authz.ScopeSelf, ID: sess.Token.UserID}); hasSelf {
				owned := strings.EqualFold(strings.TrimSpace(sess.Token.SessionID), cleanID)
				if !owned && (a.clients.Managed() || a.clients.Sessions != nil) {
					callCtx := authcontext.WithCaller(ctx, sess.Caller)
					ownedSess, findErr := a.findOwnedUserSession(callCtx, sess.Token.UserID, cleanID)
					if findErr != nil {
						return fail(findErr)
					}
					owned = ownedSess != nil
				}
				if owned {
					authorizedSess, err = a.authorizer.RequireSelfPermission(ctx, authz.PermSessionsDeleteSelf, true)
					if err != nil {
						return fail(err)
					}
				}
			}
		}
	}
	if authorizedSess == nil {
		// Jika sesi bukan milik pemanggil atau izin *self tidak dimiliki, wajib lolos iam.sessions.revoke tanpa membocorkan oracle keberadaan sesi.
		authorizedSess, err = a.requirePerm(ctx, authz.PermSessionsRevoke, true)
		if err != nil {
			return fail(err)
		}
	}

	_, rpcErr := callIAMSessions(ctx, a, &authorizedSess.Caller, func(rpcCtx context.Context, sessions iamv1.SessionServiceClient) (*iamv1.RevokeSessionByIdResponse, error) {
		return sessions.RevokeSessionById(rpcCtx, &iamv1.RevokeSessionByIdRequest{
			SessionId: cleanID,
			Reason:    toProtoSessionEndReason(reason),
		})
	})
	if rpcErr != nil {
		return fail(rpcErr)
	}
	return &model.RevokeSessionPayload{Revoked: ptrBool(true), Errors: noMutationErrors(), WasSuccessful: true}, nil
}

// RevokeAllUserSessions mencabut seluruh sesi aktif seorang pengguna sekaligus (wajib iam.sessions.revoke + reauth sensitif).
func (a *Adapter) RevokeAllUserSessions(ctx context.Context, userID string, reason *model.SessionEndReason) (*model.RevokeAllUserSessionsPayload, error) {
	fail := func(err error) (*model.RevokeAllUserSessionsPayload, error) {
		return &model.RevokeAllUserSessionsPayload{Errors: ToMutationErrors(err), WasSuccessful: false}, nil
	}
	sess, err := a.requirePerm(ctx, "iam.sessions.revoke", true)
	if err != nil {
		return fail(err)
	}
	cleanUserID, err := ParseTypedID(userID, ResourceUser)
	if err != nil {
		return fail(err)
	}
	callCtx := authcontext.WithCaller(ctx, sess.Caller)

	if err := a.ensureCanTouchTargetUser(callCtx, sess, cleanUserID, true); err != nil {
		return fail(err)
	}
	resp, rpcErr := callIAMSessions(callCtx, a, nil, func(rpcCtx context.Context, sessions iamv1.SessionServiceClient) (*iamv1.RevokeAllUserSessionsResponse, error) {
		return sessions.RevokeAllUserSessions(rpcCtx, &iamv1.RevokeAllUserSessionsRequest{
			UserId: cleanUserID,
			Reason: toProtoSessionEndReason(reason),
		})
	})
	if rpcErr != nil {
		return fail(rpcErr)
	}
	return &model.RevokeAllUserSessionsPayload{RevokedCount: ptrInt(int(resp.GetRevokedCount())), Errors: noMutationErrors(), WasSuccessful: true}, nil
}

func isSuperAdminRoleCode(code string) bool {
	trimmed := strings.TrimSpace(code)
	return strings.EqualFold(trimmed, "superadmin") || strings.EqualFold(trimmed, "super_admin")
}

func isAutomationRoleCode(code string) bool {
	trimmed := strings.ToLower(strings.TrimSpace(code))
	return strings.HasSuffix(trimmed, "_automation")
}

func callerIsSuperAdmin(sess *middleware.ResolvedSession) bool {
	if sess == nil || sess.Snapshot == nil {
		return false
	}
	return sess.Snapshot.HasRole("superadmin") || sess.Snapshot.HasRole("super_admin")
}

// findOwnedUserSession memindai seluruh halaman sesi milik principal sendiri untuk memverifikasi kepemilikan tanpa oracle lintas pengguna.
func (a *Adapter) findOwnedUserSession(callCtx context.Context, ownerUserID, targetSessionID string) (*iamv1.Session, error) {
	var offset int32
	for {
		resp, rpcErr := callIAMSessions(callCtx, a, nil, func(rpcCtx context.Context, sessions iamv1.SessionServiceClient) (*iamv1.ListUserSessionsResponse, error) {
			return sessions.ListUserSessions(rpcCtx, &iamv1.ListUserSessionsRequest{
				UserId:     ownerUserID,
				ActiveOnly: false,
				Page:       &iamv1.PageRequest{Limit: maxPageLimit, Offset: offset},
			})
		})
		if rpcErr != nil {
			return nil, rpcErr
		}
		items := resp.GetItems()
		for _, item := range items {
			if item != nil && strings.EqualFold(strings.TrimSpace(item.GetId()), strings.TrimSpace(targetSessionID)) {
				return item, nil
			}
		}
		offset += int32(len(items))
		page := resp.GetPage()
		if len(items) == 0 || page == nil || int64(offset) >= page.GetTotal() {
			return nil, nil
		}
	}
}

func ensureCanDelegatePermissionCodes(sess *middleware.ResolvedSession, permCodes []string) error {
	if callerIsSuperAdmin(sess) {
		return nil
	}
	if sess == nil || sess.Snapshot == nil {
		return middleware.ErrForbidden("caller authorization snapshot missing", nil)
	}
	for _, code := range permCodes {
		var ok bool
		if authz.IsSelfServicePermissionCode(code) && sess.Token != nil {
			_, ok = sess.Snapshot.HasScopedPermission(code, authz.ScopeGrant{Type: authz.ScopeSelf, ID: sess.Token.UserID})
		} else {
			_, ok = sess.Snapshot.HasPermission(code)
		}
		if !ok {
			return middleware.ErrForbidden("cannot delegate or modify permission not held by caller", nil)
		}
	}
	return nil
}

func (a *Adapter) ensureCanManagePermissionID(callCtx context.Context, sess *middleware.ResolvedSession, permID string) error {
	if callerIsSuperAdmin(sess) {
		return nil
	}
	// Izinkan unit test mock parsial melewati pemeriksaan sekunder; gateway terkelola selalu fail-closed.
	if !a.clients.Managed() && (a.clients.Permission == nil || a.clients.Sessions == nil) {
		return nil
	}
	resp, err := callIAMPermission(callCtx, a, nil, func(rpcCtx context.Context, perm iamv1.PermissionServiceClient) (*iamv1.GetPermissionByIdResponse, error) {
		return perm.GetPermissionById(rpcCtx, &iamv1.GetPermissionByIdRequest{Id: permID})
	})
	if err != nil {
		return err
	}
	if resp.GetPermission() == nil {
		return middleware.ErrNotFound("permission not found", nil)
	}
	return ensureCanDelegatePermissionCodes(sess, []string{resp.GetPermission().GetCode()})
}

func (a *Adapter) ensureCanManageRole(callCtx context.Context, sess *middleware.ResolvedSession, roleID string) error {
	if callerIsSuperAdmin(sess) {
		return nil
	}
	if !a.clients.Managed() && (a.clients.Roles == nil || a.clients.Sessions == nil) {
		return nil
	}
	roleResp, err := callIAMRoles(callCtx, a, nil, func(rpcCtx context.Context, roles iamv1.RoleServiceClient) (*iamv1.GetRoleByIdResponse, error) {
		return roles.GetRoleById(rpcCtx, &iamv1.GetRoleByIdRequest{Id: roleID})
	})
	if err != nil {
		return err
	}
	if roleResp.GetRole() != nil && isSuperAdminRoleCode(roleResp.GetRole().GetCode()) {
		return middleware.ErrForbidden("only superadmin may manage superadmin role", nil)
	}
	if roleResp.GetRole() != nil && isAutomationRoleCode(roleResp.GetRole().GetCode()) {
		return middleware.ErrForbidden("only superadmin may manage system automation role", nil)
	}
	rpResp, err := callIAMRoles(callCtx, a, nil, func(rpcCtx context.Context, roles iamv1.RoleServiceClient) (*iamv1.ListRolePermissionsResponse, error) {
		return roles.ListRolePermissions(rpcCtx, &iamv1.ListRolePermissionsRequest{RoleId: roleID})
	})
	if err != nil {
		return err
	}
	codes := make([]string, 0, len(rpResp.GetPermissions()))
	for _, p := range rpResp.GetPermissions() {
		if p != nil && strings.TrimSpace(p.GetCode()) != "" {
			codes = append(codes, p.GetCode())
		}
	}
	return ensureCanDelegatePermissionCodes(sess, codes)
}

func (a *Adapter) ensureCanManageGroup(callCtx context.Context, sess *middleware.ResolvedSession, groupID string) error {
	if callerIsSuperAdmin(sess) {
		return nil
	}
	if !a.clients.Managed() && (a.clients.Groups == nil || a.clients.Roles == nil || a.clients.Sessions == nil) {
		return nil
	}
	grResp, err := callIAMGroups(callCtx, a, nil, func(rpcCtx context.Context, groups iamv1.GroupServiceClient) (*iamv1.ListGroupRolesResponse, error) {
		return groups.ListGroupRoles(rpcCtx, &iamv1.ListGroupRolesRequest{
			GroupId:        groupID,
			IncludeExpired: false,
		})
	})
	if err != nil {
		return err
	}
	for _, gr := range grResp.GetGroupRoles() {
		if gr == nil || gr.GetRoleId() == "" {
			continue
		}
		if err := a.ensureCanManageRole(callCtx, sess, gr.GetRoleId()); err != nil {
			return err
		}
	}
	return nil
}

func (a *Adapter) ensureCanTouchTargetUser(callCtx context.Context, sess *middleware.ResolvedSession, targetUserID string, allowSelf bool) error {
	if !allowSelf && sess != nil && sess.Token != nil && strings.EqualFold(strings.TrimSpace(sess.Token.UserID), strings.TrimSpace(targetUserID)) {
		return middleware.ErrForbidden("cannot perform destructive administrative action on own account", nil)
	}
	if callerIsSuperAdmin(sess) {
		return nil
	}
	if !a.clients.Managed() && (a.clients.Roles == nil || a.clients.Sessions == nil) {
		return nil
	}
	effRoles, err := callIAMRoles(callCtx, a, nil, func(rpcCtx context.Context, roles iamv1.RoleServiceClient) (*iamv1.ListEffectiveUserRolesResponse, error) {
		return roles.ListEffectiveUserRoles(rpcCtx, &iamv1.ListEffectiveUserRolesRequest{UserId: targetUserID})
	})
	if err != nil {
		return err
	}
	for _, r := range effRoles.GetRoles() {
		if r == nil {
			continue
		}
		if isSuperAdminRoleCode(r.GetCode()) {
			return middleware.ErrForbidden("only superadmin may modify a superadmin account", nil)
		}
		if isAutomationRoleCode(r.GetCode()) {
			return middleware.ErrForbidden("only superadmin may modify a system automation account", nil)
		}
	}
	return nil
}

func (a *Adapter) fetchRoleMap(callCtx context.Context, roleIDs []string) (map[string]*model.Role, error) {
	out := make(map[string]*model.Role, len(roleIDs))
	for _, id := range roleIDs {
		if id == "" {
			continue
		}
		if _, seen := out[id]; seen {
			continue
		}
		resp, err := callIAMRoles(callCtx, a, nil, func(rpcCtx context.Context, roles iamv1.RoleServiceClient) (*iamv1.GetRoleByIdResponse, error) {
			return roles.GetRoleById(rpcCtx, &iamv1.GetRoleByIdRequest{Id: id})
		})
		if err != nil {
			return nil, err
		}
		out[id] = toModelRole(resp.GetRole())
	}
	return out, nil
}
