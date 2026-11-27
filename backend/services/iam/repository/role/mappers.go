package role

import (
	"strings"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
	"github.com/integrione/erp-retail/backend/services/iam/repository/pgstore"
)

const defaultManagementDomain = "iam"

func roleToDomain(m *roleModel) (*entity.Role, error) {
	if m == nil {
		return nil, iamerrors.ErrRoleNotFound
	}
	return &entity.Role{
		ID:                  strings.ToLower(strings.TrimSpace(m.ID)),
		Code:                m.Code,
		Name:                m.Name,
		Description:         pgstore.DerefString(m.Description),
		ManagementDomain:    m.ManagementDomain,
		DefaultRedirectPath: m.DefaultRedirectPath,
		RedirectPriority:    m.RedirectPriority,
		IsSystem:            m.IsSystem,
		IsActive:            m.IsActive,
		CreatedAt:           pgstore.NormalizeLocalTime(m.CreatedAt),
		UpdatedAt:           pgstore.NormalizeLocalTime(m.UpdatedAt),
	}, nil
}

func roleToModel(r *entity.Role) (roleModel, error) {
	if r == nil {
		return roleModel{}, iamerrors.NewValidation("role", iamerrors.Code("iam.role.nil"), "role must not be nil")
	}
	id, err := pgstore.ValidateUUIDParam("id", r.ID)
	if err != nil {
		return roleModel{}, err
	}
	code, err := entity.ValidateEntityCode(r.Code, entity.MaxRoleCodeLength, iamerrors.ErrInvalidRoleCode)
	if err != nil {
		return roleModel{}, err
	}
	name := strings.TrimSpace(r.Name)
	if name == "" {
		return roleModel{}, iamerrors.NewValidation("name", iamerrors.Code("iam.role.invalid_name"), "role name must not be empty")
	}
	domain := entity.NormalizeCode(r.ManagementDomain)
	if domain == "" {
		domain = defaultManagementDomain
	}
	redirectPath := strings.TrimSpace(r.DefaultRedirectPath)
	// Biarkan default_redirect_path kosong untuk peran netral (seperti iam_selfservice) agar fallback per UserKind bekerja.
	if redirectPath != "" {
		if err := entity.ValidateInternalRedirectPath(redirectPath); err != nil {
			return roleModel{}, err
		}
	}
	if r.RedirectPriority < 0 || r.RedirectPriority > entity.MaxRedirectPriority {
		return roleModel{}, iamerrors.ErrInvalidRedirectPriority
	}

	return roleModel{
		ID:                  id,
		Code:                code,
		Name:                name,
		Description:         pgstore.NullableStringPtr(r.Description),
		ManagementDomain:    domain,
		DefaultRedirectPath: redirectPath,
		RedirectPriority:    r.RedirectPriority,
		IsSystem:            r.IsSystem,
		IsActive:            r.IsActive,
		CreatedAt:           pgstore.MustLocalNow(r.CreatedAt),
		UpdatedAt:           pgstore.MustLocalNow(r.UpdatedAt),
	}, nil
}

func rolePermissionToModel(rp *entity.RolePermission) (rolePermissionModel, error) {
	if rp == nil {
		return rolePermissionModel{}, iamerrors.NewValidation("role_permission", iamerrors.Code("iam.role_permission.nil"), "role permission must not be nil")
	}
	roleID, err := pgstore.ValidateUUIDParam("role_id", rp.RoleID)
	if err != nil {
		return rolePermissionModel{}, err
	}
	permID, err := pgstore.ValidateUUIDParam("permission_id", rp.PermissionID)
	if err != nil {
		return rolePermissionModel{}, err
	}
	return rolePermissionModel{
		RoleID:       roleID,
		PermissionID: permID,
		CreatedAt:    pgstore.MustLocalNow(rp.CreatedAt),
	}, nil
}

func userRoleToDomain(m *userRoleModel) (*entity.UserRole, error) {
	if m == nil {
		return nil, iamerrors.ErrUserRoleNotFound
	}
	scope, err := pgstore.ToScope(m.ScopeType, m.ScopeID)
	if err != nil {
		return nil, err
	}
	var grantedBy *string
	if m.AssignedBy != nil {
		trimmed := strings.ToLower(strings.TrimSpace(*m.AssignedBy))
		if trimmed != "" {
			grantedBy = &trimmed
		}
	}
	return &entity.UserRole{
		ID:        strings.ToLower(strings.TrimSpace(m.ID)),
		UserID:    strings.ToLower(strings.TrimSpace(m.UserID)),
		RoleID:    strings.ToLower(strings.TrimSpace(m.RoleID)),
		Scope:     scope,
		GrantedBy: grantedBy,
		ExpiresAt: pgstore.NormalizeLocalTimePtr(m.ExpiresAt),
		CreatedAt: pgstore.NormalizeLocalTime(m.CreatedAt),
	}, nil
}

func userRoleToModel(ur *entity.UserRole) (userRoleModel, error) {
	if ur == nil {
		return userRoleModel{}, iamerrors.NewValidation("user_role", iamerrors.Code("iam.user_role.nil"), "user role must not be nil")
	}
	id, err := pgstore.ValidateUUIDParam("id", ur.ID)
	if err != nil {
		return userRoleModel{}, err
	}
	userID, err := pgstore.ValidateUUIDParam("user_id", ur.UserID)
	if err != nil {
		return userRoleModel{}, err
	}
	roleID, err := pgstore.ValidateUUIDParam("role_id", ur.RoleID)
	if err != nil {
		return userRoleModel{}, err
	}
	if !ur.Scope.IsValid() {
		return userRoleModel{}, iamerrors.ErrInvalidScope
	}
	var scopeID *string
	if ur.Scope.ID != nil {
		s, err := pgstore.ValidateUUIDParam("scope_id", *ur.Scope.ID)
		if err != nil {
			return userRoleModel{}, iamerrors.ErrInvalidScope
		}
		scopeID = &s
	}
	var assignedBy *string
	if ur.GrantedBy != nil && strings.TrimSpace(*ur.GrantedBy) != "" {
		g, err := pgstore.ValidateUUIDParam("granted_by", *ur.GrantedBy)
		if err != nil {
			return userRoleModel{}, err
		}
		assignedBy = &g
	}
	return userRoleModel{
		ID:         id,
		UserID:     userID,
		RoleID:     roleID,
		ScopeType:  ur.Scope.Type.String(),
		ScopeID:    scopeID,
		AssignedBy: assignedBy,
		ExpiresAt:  pgstore.NormalizeLocalTimePtr(ur.ExpiresAt),
		CreatedAt:  pgstore.MustLocalNow(ur.CreatedAt),
	}, nil
}
