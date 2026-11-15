package group

import (
	"strings"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
	"github.com/integrione/erp-retail/backend/services/iam/repository/pgstore"
)

func groupToDomain(m *groupModel) (*entity.Group, error) {
	if m == nil {
		return nil, iamerrors.ErrGroupNotFound
	}
	return &entity.Group{
		ID:          strings.ToLower(strings.TrimSpace(m.ID)),
		Code:        m.Code,
		Name:        m.Name,
		Description: pgstore.DerefString(m.Description),
		IsActive:    m.IsActive,
		CreatedAt:   pgstore.NormalizeLocalTime(m.CreatedAt),
		UpdatedAt:   pgstore.NormalizeLocalTime(m.UpdatedAt),
	}, nil
}

func groupToModel(g *entity.Group) (groupModel, error) {
	if g == nil {
		return groupModel{}, iamerrors.NewValidation("group", iamerrors.Code("iam.group.nil"), "group must not be nil")
	}
	id, err := pgstore.ValidateUUIDParam("id", g.ID)
	if err != nil {
		return groupModel{}, err
	}
	code, err := entity.ValidateEntityCode(g.Code, entity.MaxGroupCodeLength, iamerrors.ErrInvalidGroupCode)
	if err != nil {
		return groupModel{}, err
	}
	name := strings.TrimSpace(g.Name)
	if name == "" {
		return groupModel{}, iamerrors.NewValidation("name", iamerrors.Code("iam.group.invalid_name"), "group name must not be empty")
	}
	return groupModel{
		ID:          id,
		Code:        code,
		Name:        name,
		Description: pgstore.NullableStringPtr(g.Description),
		IsActive:    g.IsActive,
		CreatedAt:   pgstore.MustLocalNow(g.CreatedAt),
		UpdatedAt:   pgstore.MustLocalNow(g.UpdatedAt),
	}, nil
}

func groupMemberToModel(gm *entity.GroupMember) (groupMemberModel, error) {
	if gm == nil {
		return groupMemberModel{}, iamerrors.NewValidation("group_member", iamerrors.Code("iam.group_member.nil"), "group member must not be nil")
	}
	groupID, err := pgstore.ValidateUUIDParam("group_id", gm.GroupID)
	if err != nil {
		return groupMemberModel{}, err
	}
	userID, err := pgstore.ValidateUUIDParam("user_id", gm.UserID)
	if err != nil {
		return groupMemberModel{}, err
	}
	return groupMemberModel{
		GroupID:   groupID,
		UserID:    userID,
		CreatedAt: pgstore.MustLocalNow(gm.CreatedAt),
	}, nil
}

func groupRoleToDomain(m *groupRoleModel) (*entity.GroupRole, error) {
	if m == nil {
		return nil, iamerrors.ErrGroupRoleNotFound
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
	return &entity.GroupRole{
		ID:        strings.ToLower(strings.TrimSpace(m.ID)),
		GroupID:   strings.ToLower(strings.TrimSpace(m.GroupID)),
		RoleID:    strings.ToLower(strings.TrimSpace(m.RoleID)),
		Scope:     scope,
		GrantedBy: grantedBy,
		ExpiresAt: pgstore.NormalizeLocalTimePtr(m.ExpiresAt),
		CreatedAt: pgstore.NormalizeLocalTime(m.CreatedAt),
	}, nil
}

func groupRoleToModel(gr *entity.GroupRole) (groupRoleModel, error) {
	if gr == nil {
		return groupRoleModel{}, iamerrors.NewValidation("group_role", iamerrors.Code("iam.group_role.nil"), "group role must not be nil")
	}
	id, err := pgstore.ValidateUUIDParam("id", gr.ID)
	if err != nil {
		return groupRoleModel{}, err
	}
	groupID, err := pgstore.ValidateUUIDParam("group_id", gr.GroupID)
	if err != nil {
		return groupRoleModel{}, err
	}
	roleID, err := pgstore.ValidateUUIDParam("role_id", gr.RoleID)
	if err != nil {
		return groupRoleModel{}, err
	}
	if !gr.Scope.IsValid() {
		return groupRoleModel{}, iamerrors.ErrInvalidScope
	}
	var scopeID *string
	if gr.Scope.ID != nil {
		s, err := pgstore.ValidateUUIDParam("scope_id", *gr.Scope.ID)
		if err != nil {
			return groupRoleModel{}, iamerrors.ErrInvalidScope
		}
		scopeID = &s
	}
	var assignedBy *string
	if gr.GrantedBy != nil && strings.TrimSpace(*gr.GrantedBy) != "" {
		g, err := pgstore.ValidateUUIDParam("granted_by", *gr.GrantedBy)
		if err != nil {
			return groupRoleModel{}, err
		}
		assignedBy = &g
	}
	return groupRoleModel{
		ID:         id,
		GroupID:    groupID,
		RoleID:     roleID,
		ScopeType:  gr.Scope.Type.String(),
		ScopeID:    scopeID,
		AssignedBy: assignedBy,
		ExpiresAt:  pgstore.NormalizeLocalTimePtr(gr.ExpiresAt),
		CreatedAt:  pgstore.MustLocalNow(gr.CreatedAt),
	}, nil
}

func groupUserToDomain(m *groupUserModel) (*entity.User, error) {
	if m == nil {
		return nil, iamerrors.ErrUserNotFound
	}
	kind, err := entity.ParseUserKind(m.UserKind)
	if err != nil {
		return nil, err
	}
	status, err := entity.ParseUserStatus(m.Status)
	if err != nil {
		return nil, err
	}
	return &entity.User{
		ID:          strings.ToLower(strings.TrimSpace(m.ID)),
		Kind:        kind,
		Status:      status,
		DisplayName: m.DisplayName,
		AvatarURL:   pgstore.DerefString(m.AvatarURL),
		Locale:      m.Locale,
		Timezone:    m.Timezone,
		LastLoginAt: pgstore.NormalizeLocalTimePtr(m.LastLoginAt),
		CreatedAt:   pgstore.NormalizeLocalTime(m.CreatedAt),
		UpdatedAt:   pgstore.NormalizeLocalTime(m.UpdatedAt),
		DeletedAt:   pgstore.NormalizeLocalTimePtr(m.DeletedAt),
	}, nil
}
