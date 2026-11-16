package handler

import (
	"context"
	"fmt"

	iamv1 "github.com/integrione/erp-retail/backend/services/iam/contract/gen/iam/v1"
	domainrepo "github.com/integrione/erp-retail/backend/services/iam/domain/repository"
	"github.com/integrione/erp-retail/backend/services/iam/services/group"
)

// GroupHandler menangani permintaan gRPC untuk pengelolaan grup, anggota, dan peran grup.
type GroupHandler struct {
	iamv1.UnimplementedGroupServiceServer
	svc       *group.Service
	errMapper *ErrorMapper
}

// NewGroupHandler membentuk handler gRPC untuk GroupService secara fail-closed.
func NewGroupHandler(svc *group.Service, errMapper *ErrorMapper) (*GroupHandler, error) {
	if svc == nil {
		return nil, fmt.Errorf("group service cannot be nil")
	}
	if errMapper == nil {
		errMapper = NewErrorMapper(nil)
	}
	return &GroupHandler{
		svc:       svc,
		errMapper: errMapper,
	}, nil
}

// GetGroupById mengambil satu grup berdasarkan UUID.
func (h *GroupHandler) GetGroupById(ctx context.Context, req *iamv1.GetGroupByIdRequest) (*iamv1.GetGroupByIdResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	g, err := h.svc.GetByID(ctx, req.GetId())
	if err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	return &iamv1.GetGroupByIdResponse{Group: toProtoGroup(g)}, nil
}

// GetGroupByCode mengambil satu grup berdasarkan kode uniknya.
func (h *GroupHandler) GetGroupByCode(ctx context.Context, req *iamv1.GetGroupByCodeRequest) (*iamv1.GetGroupByCodeResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	g, err := h.svc.GetByCode(ctx, req.GetCode())
	if err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	return &iamv1.GetGroupByCodeResponse{Group: toProtoGroup(g)}, nil
}

// ListGroups mengembalikan daftar grup terpaginasi berdasarkan filter.
func (h *GroupHandler) ListGroups(ctx context.Context, req *iamv1.ListGroupsRequest) (*iamv1.ListGroupsResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	res, err := h.svc.List(ctx, domainrepo.GroupFilter{
		Query:      req.GetQuery(),
		ActiveOnly: req.GetActiveOnly(),
		Page:       fromProtoPageRequest(req.GetPage()),
	})
	if err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	return &iamv1.ListGroupsResponse{
		Items: toProtoGroups(res.Items),
		Page:  toProtoPageInfo(res),
	}, nil
}

// CreateGroup membuat grup pengguna baru.
func (h *GroupHandler) CreateGroup(ctx context.Context, req *iamv1.CreateGroupRequest) (*iamv1.CreateGroupResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	g, err := h.svc.Create(ctx, group.CreateInput{
		Code:        req.GetCode(),
		Name:        req.GetName(),
		Description: req.GetDescription(),
	})
	if err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	return &iamv1.CreateGroupResponse{Group: toProtoGroup(g)}, nil
}

// UpdateGroupDetails memperbarui nama dan deskripsi grup.
func (h *GroupHandler) UpdateGroupDetails(ctx context.Context, req *iamv1.UpdateGroupDetailsRequest) (*iamv1.UpdateGroupDetailsResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	g, err := h.svc.UpdateDetails(ctx, group.UpdateInput{
		ID:          req.GetId(),
		Name:        req.GetName(),
		Description: req.GetDescription(),
	})
	if err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	return &iamv1.UpdateGroupDetailsResponse{Group: toProtoGroup(g)}, nil
}

// ActivateGroup mengaktifkan kembali grup agar peran warisannya berlaku.
func (h *GroupHandler) ActivateGroup(ctx context.Context, req *iamv1.ActivateGroupRequest) (*iamv1.ActivateGroupResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	g, err := h.svc.Activate(ctx, req.GetId())
	if err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	return &iamv1.ActivateGroupResponse{Group: toProtoGroup(g)}, nil
}

// DeactivateGroup menonaktifkan grup.
func (h *GroupHandler) DeactivateGroup(ctx context.Context, req *iamv1.DeactivateGroupRequest) (*iamv1.DeactivateGroupResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	g, err := h.svc.Deactivate(ctx, req.GetId())
	if err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	return &iamv1.DeactivateGroupResponse{Group: toProtoGroup(g)}, nil
}

// DeleteGroup menghapus grup berdasarkan UUID.
func (h *GroupHandler) DeleteGroup(ctx context.Context, req *iamv1.DeleteGroupRequest) (*iamv1.DeleteGroupResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	if err := h.svc.Delete(ctx, req.GetId()); err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	return &iamv1.DeleteGroupResponse{}, nil
}

// AddGroupMember menambahkan pengguna aktif ke dalam grup.
func (h *GroupHandler) AddGroupMember(ctx context.Context, req *iamv1.AddGroupMemberRequest) (*iamv1.AddGroupMemberResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	m, err := h.svc.AddMember(ctx, req.GetGroupId(), req.GetUserId())
	if err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	return &iamv1.AddGroupMemberResponse{Member: toProtoGroupMember(m)}, nil
}

// RemoveGroupMember menghapus pengguna dari keanggotaan grup.
func (h *GroupHandler) RemoveGroupMember(ctx context.Context, req *iamv1.RemoveGroupMemberRequest) (*iamv1.RemoveGroupMemberResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	if err := h.svc.RemoveMember(ctx, req.GetGroupId(), req.GetUserId()); err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	return &iamv1.RemoveGroupMemberResponse{}, nil
}

// ListGroupMembers mengambil daftar pengguna anggota grup secara terpaginasi.
func (h *GroupHandler) ListGroupMembers(ctx context.Context, req *iamv1.ListGroupMembersRequest) (*iamv1.ListGroupMembersResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	res, err := h.svc.ListMembers(ctx, req.GetGroupId(), fromProtoPageRequest(req.GetPage()))
	if err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	return &iamv1.ListGroupMembersResponse{
		Items: toProtoUsers(res.Items),
		Page:  toProtoPageInfo(res),
	}, nil
}

// ListUserGroups mengambil seluruh grup yang diikuti pengguna tertentu.
func (h *GroupHandler) ListUserGroups(ctx context.Context, req *iamv1.ListUserGroupsRequest) (*iamv1.ListUserGroupsResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	groups, err := h.svc.ListUserGroups(ctx, req.GetUserId())
	if err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	return &iamv1.ListUserGroupsResponse{Groups: toProtoGroups(groups)}, nil
}

// AssignGroupRole menugaskan peran berscope ke sebuah grup.
func (h *GroupHandler) AssignGroupRole(ctx context.Context, req *iamv1.AssignGroupRoleRequest) (*iamv1.AssignGroupRoleResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	expiresAt, err := fromProtoTimestampPtr("expires_at", req.GetExpiresAt())
	if err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	gr, err := h.svc.AssignRole(ctx, group.AssignGroupRoleInput{
		GroupID:   req.GetGroupId(),
		RoleID:    req.GetRoleId(),
		ScopeType: fromProtoScopeType(req.GetScopeType()),
		ScopeID:   cloneStringPtr(req.ScopeId),
		ExpiresAt: expiresAt,
	})
	if err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	return &iamv1.AssignGroupRoleResponse{GroupRole: toProtoGroupRole(gr)}, nil
}

// RevokeGroupRole mencabut penugasan peran grup berdasarkan ID assignment.
func (h *GroupHandler) RevokeGroupRole(ctx context.Context, req *iamv1.RevokeGroupRoleRequest) (*iamv1.RevokeGroupRoleResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	if err := h.svc.RevokeRole(ctx, req.GetAssignmentId()); err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	return &iamv1.RevokeGroupRoleResponse{}, nil
}

// ListGroupRoles mengambil daftar penugasan peran milik suatu grup.
func (h *GroupHandler) ListGroupRoles(ctx context.Context, req *iamv1.ListGroupRolesRequest) (*iamv1.ListGroupRolesResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	items, err := h.svc.ListGroupRoles(ctx, req.GetGroupId(), req.GetIncludeExpired())
	if err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	return &iamv1.ListGroupRolesResponse{GroupRoles: toProtoGroupRoles(items)}, nil
}
