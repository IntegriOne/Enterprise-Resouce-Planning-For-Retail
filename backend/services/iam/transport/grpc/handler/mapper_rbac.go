package handler

import (
	iamv1 "github.com/integrione/erp-retail/backend/services/iam/contract/gen/iam/v1"
	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
)

func toProtoPermission(p *entity.Permission) *iamv1.Permission {
	if p == nil {
		return nil
	}
	return &iamv1.Permission{
		Id:          p.ID,
		Code:        p.Code,
		Resource:    p.Resource,
		Action:      p.Action,
		Description: p.Description,
		IsSensitive: p.IsSensitive,
		CreatedAt:   toProtoTimestamp(p.CreatedAt),
	}
}

func toProtoPermissions(items []*entity.Permission) []*iamv1.Permission {
	out := make([]*iamv1.Permission, 0, len(items))
	for _, p := range items {
		out = append(out, toProtoPermission(p))
	}
	return out
}

func toProtoRole(r *entity.Role) *iamv1.Role {
	if r == nil {
		return nil
	}
	return &iamv1.Role{
		Id:                  r.ID,
		Code:                r.Code,
		Name:                r.Name,
		Description:         r.Description,
		ManagementDomain:    r.ManagementDomain,
		DefaultRedirectPath: r.DefaultRedirectPath,
		RedirectPriority:    int32(r.RedirectPriority),
		IsSystem:            r.IsSystem,
		IsActive:            r.IsActive,
		CreatedAt:           toProtoTimestamp(r.CreatedAt),
		UpdatedAt:           toProtoTimestamp(r.UpdatedAt),
	}
}

func toProtoRoles(items []*entity.Role) []*iamv1.Role {
	out := make([]*iamv1.Role, 0, len(items))
	for _, r := range items {
		out = append(out, toProtoRole(r))
	}
	return out
}

func toProtoRoleValues(items []entity.Role) []*iamv1.Role {
	out := make([]*iamv1.Role, 0, len(items))
	for i := range items {
		out = append(out, toProtoRole(&items[i]))
	}
	return out
}

func toProtoRolePermission(rp *entity.RolePermission) *iamv1.RolePermission {
	if rp == nil {
		return nil
	}
	return &iamv1.RolePermission{
		RoleId:       rp.RoleID,
		PermissionId: rp.PermissionID,
		CreatedAt:    toProtoTimestamp(rp.CreatedAt),
	}
}

func toProtoUserRole(ur *entity.UserRole) *iamv1.UserRole {
	if ur == nil {
		return nil
	}
	return &iamv1.UserRole{
		Id:        ur.ID,
		UserId:    ur.UserID,
		RoleId:    ur.RoleID,
		Scope:     toProtoScope(ur.Scope),
		GrantedBy: cloneStringPtr(ur.GrantedBy),
		ExpiresAt: toProtoTimestampPtr(ur.ExpiresAt),
		CreatedAt: toProtoTimestamp(ur.CreatedAt),
	}
}

func toProtoUserRoles(items []*entity.UserRole) []*iamv1.UserRole {
	out := make([]*iamv1.UserRole, 0, len(items))
	for _, ur := range items {
		out = append(out, toProtoUserRole(ur))
	}
	return out
}

func toProtoGroup(g *entity.Group) *iamv1.Group {
	if g == nil {
		return nil
	}
	return &iamv1.Group{
		Id:          g.ID,
		Code:        g.Code,
		Name:        g.Name,
		Description: g.Description,
		IsActive:    g.IsActive,
		CreatedAt:   toProtoTimestamp(g.CreatedAt),
		UpdatedAt:   toProtoTimestamp(g.UpdatedAt),
	}
}

func toProtoGroups(items []*entity.Group) []*iamv1.Group {
	out := make([]*iamv1.Group, 0, len(items))
	for _, g := range items {
		out = append(out, toProtoGroup(g))
	}
	return out
}

func toProtoGroupMember(m *entity.GroupMember) *iamv1.GroupMember {
	if m == nil {
		return nil
	}
	return &iamv1.GroupMember{
		GroupId:   m.GroupID,
		UserId:    m.UserID,
		CreatedAt: toProtoTimestamp(m.CreatedAt),
	}
}

func toProtoGroupRole(gr *entity.GroupRole) *iamv1.GroupRole {
	if gr == nil {
		return nil
	}
	return &iamv1.GroupRole{
		Id:        gr.ID,
		GroupId:   gr.GroupID,
		RoleId:    gr.RoleID,
		Scope:     toProtoScope(gr.Scope),
		GrantedBy: cloneStringPtr(gr.GrantedBy),
		ExpiresAt: toProtoTimestampPtr(gr.ExpiresAt),
		CreatedAt: toProtoTimestamp(gr.CreatedAt),
	}
}

func toProtoGroupRoles(items []*entity.GroupRole) []*iamv1.GroupRole {
	out := make([]*iamv1.GroupRole, 0, len(items))
	for _, gr := range items {
		out = append(out, toProtoGroupRole(gr))
	}
	return out
}

func toProtoAccessSnapshot(s *entity.AccessSnapshot) *iamv1.AccessSnapshot {
	if s == nil {
		return nil
	}
	perms := make([]*iamv1.ScopedPermission, 0, len(s.Permissions))
	for _, p := range s.Permissions {
		perms = append(perms, &iamv1.ScopedPermission{
			PermissionCode: p.PermissionCode,
			IsSensitive:    p.IsSensitive,
			Scope:          toProtoScope(p.Scope),
		})
	}
	return &iamv1.AccessSnapshot{
		UserId:        s.UserID,
		UserKind:      toProtoUserKind(s.UserKind),
		UserStatus:    toProtoUserStatus(s.UserStatus),
		MfaEnabled:    s.MFAEnabled,
		Roles:         toProtoRoleValues(s.Roles),
		Permissions:   perms,
		CheckedAt:     toProtoTimestamp(s.CheckedAt),
		AuthEpoch:     s.AuthEpoch,
		PermRevision:  s.PermRevision,
		RecoveryEpoch: s.RecoveryEpoch,
	}
}

func cloneStringPtr(s *string) *string {
	if s == nil {
		return nil
	}
	v := *s
	return &v
}
