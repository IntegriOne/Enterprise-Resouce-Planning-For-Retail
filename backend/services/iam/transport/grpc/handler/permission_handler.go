package handler

import (
	"context"
	"fmt"

	iamv1 "github.com/integrione/erp-retail/backend/services/iam/contract/gen/iam/v1"
	domainrepo "github.com/integrione/erp-retail/backend/services/iam/domain/repository"
	"github.com/integrione/erp-retail/backend/services/iam/services/permission"
)

// PermissionHandler menangani permintaan gRPC untuk katalog izin atomik RBAC IAM.
type PermissionHandler struct {
	iamv1.UnimplementedPermissionServiceServer
	svc       *permission.Service
	errMapper *ErrorMapper
}

// NewPermissionHandler membentuk handler gRPC untuk PermissionService secara fail-closed.
func NewPermissionHandler(svc *permission.Service, errMapper *ErrorMapper) (*PermissionHandler, error) {
	if svc == nil {
		return nil, fmt.Errorf("permission service cannot be nil")
	}
	if errMapper == nil {
		errMapper = NewErrorMapper(nil)
	}
	return &PermissionHandler{
		svc:       svc,
		errMapper: errMapper,
	}, nil
}

// GetPermissionById mengambil satu entitas izin berdasarkan UUID-nya.
func (h *PermissionHandler) GetPermissionById(ctx context.Context, req *iamv1.GetPermissionByIdRequest) (*iamv1.GetPermissionByIdResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	p, err := h.svc.GetByID(ctx, req.GetId())
	if err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	return &iamv1.GetPermissionByIdResponse{Permission: toProtoPermission(p)}, nil
}

// GetPermissionByCode mengambil satu entitas izin berdasarkan kode uniknya.
func (h *PermissionHandler) GetPermissionByCode(ctx context.Context, req *iamv1.GetPermissionByCodeRequest) (*iamv1.GetPermissionByCodeResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	p, err := h.svc.GetByCode(ctx, req.GetCode())
	if err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	return &iamv1.GetPermissionByCodeResponse{Permission: toProtoPermission(p)}, nil
}

// ListPermissions mengembalikan daftar izin terpaginasi berdasarkan filter.
func (h *PermissionHandler) ListPermissions(ctx context.Context, req *iamv1.ListPermissionsRequest) (*iamv1.ListPermissionsResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	var isSens *bool
	if req.IsSensitive != nil {
		v := req.GetIsSensitive()
		isSens = &v
	}
	res, err := h.svc.List(ctx, domainrepo.PermissionFilter{
		Query:       req.GetQuery(),
		Resource:    req.GetResource(),
		IsSensitive: isSens,
		Page:        fromProtoPageRequest(req.GetPage()),
	})
	if err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	return &iamv1.ListPermissionsResponse{
		Items: toProtoPermissions(res.Items),
		Page:  toProtoPageInfo(res),
	}, nil
}

// ListPermissionsByRole mengambil seluruh izin yang terikat ke peran tertentu.
func (h *PermissionHandler) ListPermissionsByRole(ctx context.Context, req *iamv1.ListPermissionsByRoleRequest) (*iamv1.ListPermissionsByRoleResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	perms, err := h.svc.ListByRoleID(ctx, req.GetRoleId())
	if err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	return &iamv1.ListPermissionsByRoleResponse{
		Permissions: toProtoPermissions(perms),
	}, nil
}

// CreatePermission menambahkan entri izin baru ke katalog RBAC IAM.
func (h *PermissionHandler) CreatePermission(ctx context.Context, req *iamv1.CreatePermissionRequest) (*iamv1.CreatePermissionResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	p, err := h.svc.Create(ctx, permission.CreateInput{
		Code:        req.GetCode(),
		Resource:    req.GetResource(),
		Action:      req.GetAction(),
		Description: req.GetDescription(),
		IsSensitive: req.GetIsSensitive(),
	})
	if err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	return &iamv1.CreatePermissionResponse{Permission: toProtoPermission(p)}, nil
}

// DeletePermission menghapus izin berdasarkan UUID-nya.
func (h *PermissionHandler) DeletePermission(ctx context.Context, req *iamv1.DeletePermissionRequest) (*iamv1.DeletePermissionResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	if err := h.svc.Delete(ctx, req.GetId()); err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	return &iamv1.DeletePermissionResponse{}, nil
}
