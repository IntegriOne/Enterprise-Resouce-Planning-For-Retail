package handler

import (
	"context"
	"fmt"

	iamv1 "github.com/integrione/erp-retail/backend/services/iam/contract/gen/iam/v1"
	domainrepo "github.com/integrione/erp-retail/backend/services/iam/domain/repository"
	"github.com/integrione/erp-retail/backend/services/iam/services/role"
)

// RoleHandler menangani permintaan gRPC untuk manajemen peran RBAC dan ikatan izin/pengguna.
type RoleHandler struct {
	iamv1.UnimplementedRoleServiceServer
	svc       *role.Service
	errMapper *ErrorMapper
}

// NewRoleHandler membentuk handler gRPC untuk RoleService secara fail-closed.
func NewRoleHandler(svc *role.Service, errMapper *ErrorMapper) (*RoleHandler, error) {
	if svc == nil {
		return nil, fmt.Errorf("role service cannot be nil")
	}
	if errMapper == nil {
		errMapper = NewErrorMapper(nil)
	}
	return &RoleHandler{
		svc:       svc,
		errMapper: errMapper,
	}, nil
}

// GetRoleById mengambil satu peran berdasarkan UUID.
func (h *RoleHandler) GetRoleById(ctx context.Context, req *iamv1.GetRoleByIdRequest) (*iamv1.GetRoleByIdResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	r, err := h.svc.GetByID(ctx, req.GetId())
	if err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	return &iamv1.GetRoleByIdResponse{Role: toProtoRole(r)}, nil
}

// GetRoleByCode mengambil satu peran berdasarkan kode uniknya.
func (h *RoleHandler) GetRoleByCode(ctx context.Context, req *iamv1.GetRoleByCodeRequest) (*iamv1.GetRoleByCodeResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	r, err := h.svc.GetByCode(ctx, req.GetCode())
	if err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	return &iamv1.GetRoleByCodeResponse{Role: toProtoRole(r)}, nil
}

// ListRoles mengembalikan daftar peran terpaginasi berdasarkan filter.
func (h *RoleHandler) ListRoles(ctx context.Context, req *iamv1.ListRolesRequest) (*iamv1.ListRolesResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	res, err := h.svc.List(ctx, domainrepo.RoleFilter{
		Query:            req.GetQuery(),
		ManagementDomain: req.GetManagementDomain(),
		ActiveOnly:       req.GetActiveOnly(),
		Page:             fromProtoPageRequest(req.GetPage()),
	})
	if err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	return &iamv1.ListRolesResponse{
		Items: toProtoRoles(res.Items),
		Page:  toProtoPageInfo(res),
	}, nil
}

// CreateRole membuat definisi peran baru beserta validasi rute redirect default-nya.
func (h *RoleHandler) CreateRole(ctx context.Context, req *iamv1.CreateRoleRequest) (*iamv1.CreateRoleResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	r, err := h.svc.Create(ctx, role.CreateInput{
		Code:                req.GetCode(),
		Name:                req.GetName(),
		Description:         req.GetDescription(),
		ManagementDomain:    req.GetManagementDomain(),
		DefaultRedirectPath: req.GetDefaultRedirectPath(),
		RedirectPriority:    int(req.GetRedirectPriority()),
		IsSystem:            req.GetIsSystem(),
	})
	if err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	return &iamv1.CreateRoleResponse{Role: toProtoRole(r)}, nil
}

// UpdateRoleDetails memperbarui detail peran RBAC yang dapat diubah.
func (h *RoleHandler) UpdateRoleDetails(ctx context.Context, req *iamv1.UpdateRoleDetailsRequest) (*iamv1.UpdateRoleDetailsResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	r, err := h.svc.UpdateDetails(ctx, role.UpdateInput{
		ID:                  req.GetId(),
		Name:                req.GetName(),
		Description:         req.GetDescription(),
		ManagementDomain:    req.GetManagementDomain(),
		DefaultRedirectPath: req.GetDefaultRedirectPath(),
		RedirectPriority:    int(req.GetRedirectPriority()),
	})
	if err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	return &iamv1.UpdateRoleDetailsResponse{Role: toProtoRole(r)}, nil
}

// ActivateRole mengaktifkan peran agar kembali efektif dalam otorisasi.
func (h *RoleHandler) ActivateRole(ctx context.Context, req *iamv1.ActivateRoleRequest) (*iamv1.ActivateRoleResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	r, err := h.svc.Activate(ctx, req.GetId())
	if err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	return &iamv1.ActivateRoleResponse{Role: toProtoRole(r)}, nil
}

// DeactivateRole menonaktifkan peran non-sistem.
func (h *RoleHandler) DeactivateRole(ctx context.Context, req *iamv1.DeactivateRoleRequest) (*iamv1.DeactivateRoleResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	r, err := h.svc.Deactivate(ctx, req.GetId())
	if err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	return &iamv1.DeactivateRoleResponse{Role: toProtoRole(r)}, nil
}

// DeleteRole menghapus peran non-sistem berdasarkan UUID.
func (h *RoleHandler) DeleteRole(ctx context.Context, req *iamv1.DeleteRoleRequest) (*iamv1.DeleteRoleResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	if err := h.svc.Delete(ctx, req.GetId()); err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	return &iamv1.DeleteRoleResponse{}, nil
}

// AssignRolePermission mengikat satu izin atomik ke sebuah peran.
func (h *RoleHandler) AssignRolePermission(ctx context.Context, req *iamv1.AssignRolePermissionRequest) (*iamv1.AssignRolePermissionResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	rp, err := h.svc.AssignPermission(ctx, req.GetRoleId(), req.GetPermissionId())
	if err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	return &iamv1.AssignRolePermissionResponse{RolePermission: toProtoRolePermission(rp)}, nil
}

// RevokeRolePermission mencabut ikatan izin dari peran.
func (h *RoleHandler) RevokeRolePermission(ctx context.Context, req *iamv1.RevokeRolePermissionRequest) (*iamv1.RevokeRolePermissionResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	if err := h.svc.RevokePermission(ctx, req.GetRoleId(), req.GetPermissionId()); err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	return &iamv1.RevokeRolePermissionResponse{}, nil
}

// ListRolePermissions mengambil seluruh izin yang terikat ke peran.
func (h *RoleHandler) ListRolePermissions(ctx context.Context, req *iamv1.ListRolePermissionsRequest) (*iamv1.ListRolePermissionsResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	perms, err := h.svc.ListPermissions(ctx, req.GetRoleId())
	if err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	return &iamv1.ListRolePermissionsResponse{Permissions: toProtoPermissions(perms)}, nil
}

// AssignUserRole memberikan peran berscope langsung ke pengguna.
func (h *RoleHandler) AssignUserRole(ctx context.Context, req *iamv1.AssignUserRoleRequest) (*iamv1.AssignUserRoleResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	expiresAt, err := fromProtoTimestampPtr("expires_at", req.GetExpiresAt())
	if err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	ur, err := h.svc.AssignToUser(ctx, role.AssignUserRoleInput{
		UserID:    req.GetUserId(),
		RoleID:    req.GetRoleId(),
		ScopeType: fromProtoScopeType(req.GetScopeType()),
		ScopeID:   cloneStringPtr(req.ScopeId),
		ExpiresAt: expiresAt,
	})
	if err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	return &iamv1.AssignUserRoleResponse{UserRole: toProtoUserRole(ur)}, nil
}

// RevokeUserRole mencabut penugasan peran langsung berdasarkan ID assignment.
func (h *RoleHandler) RevokeUserRole(ctx context.Context, req *iamv1.RevokeUserRoleRequest) (*iamv1.RevokeUserRoleResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	if err := h.svc.RevokeUserRole(ctx, req.GetAssignmentId()); err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	return &iamv1.RevokeUserRoleResponse{}, nil
}

// ListUserRoles mengambil daftar penugasan peran langsung milik pengguna.
func (h *RoleHandler) ListUserRoles(ctx context.Context, req *iamv1.ListUserRolesRequest) (*iamv1.ListUserRolesResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	items, err := h.svc.ListUserRoles(ctx, req.GetUserId(), req.GetIncludeExpired())
	if err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	return &iamv1.ListUserRolesResponse{UserRoles: toProtoUserRoles(items)}, nil
}

// ListEffectiveUserRoles mengambil seluruh peran efektif pengguna (langsung maupun lewat grup).
func (h *RoleHandler) ListEffectiveUserRoles(ctx context.Context, req *iamv1.ListEffectiveUserRolesRequest) (*iamv1.ListEffectiveUserRolesResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	roles, err := h.svc.ListEffectiveRoles(ctx, req.GetUserId())
	if err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	return &iamv1.ListEffectiveUserRolesResponse{Roles: toProtoRoleValues(roles)}, nil
}
