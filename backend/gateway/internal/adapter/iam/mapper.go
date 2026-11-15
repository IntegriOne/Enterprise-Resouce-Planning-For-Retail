package iam

import (
	"strings"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/source_connections"
	iamv1 "github.com/integrione/erp-retail/backend/services/iam/contract/gen/iam/v1"
	"github.com/integrione/erp-retail/backend/shared/authz"
)

func toLocalTime(ts *timestamppb.Timestamp) time.Time {
	if ts == nil || !ts.IsValid() {
		return time.Time{}
	}
	return ts.AsTime().Local()
}

func toLocalTimePtr(ts *timestamppb.Timestamp) *time.Time {
	if ts == nil || !ts.IsValid() {
		return nil
	}
	v := ts.AsTime().Local()
	return &v
}

func toProtoTimePtr(t *time.Time) *timestamppb.Timestamp {
	if t == nil || t.IsZero() {
		return nil
	}
	return timestamppb.New(*t)
}

func toModelUser(u *iamv1.User) *model.User {
	if u == nil {
		return nil
	}
	return &model.User{
		ID:           EncodeGID(ResourceUser, u.GetId()),
		RawID:        u.GetId(),
		Kind:         fromProtoUserKind(u.GetKind()),
		Status:       fromProtoUserStatus(u.GetStatus()),
		DisplayName:  u.GetDisplayName(),
		AvatarURL:    u.GetAvatarUrl(),
		Locale:       u.GetLocale(),
		Timezone:     u.GetTimezone(),
		LastLoginAt:  toLocalTimePtr(u.GetLastLoginAt()),
		CreatedAt:    toLocalTime(u.GetCreatedAt()),
		UpdatedAt:    toLocalTime(u.GetUpdatedAt()),
		DeletedAt:    toLocalTimePtr(u.GetDeletedAt()),
		AuthEpoch:    int(u.GetAuthEpoch()),
		PermRevision: int(u.GetPermRevision()),
	}
}

func toModelUsers(items []*iamv1.User) []*model.User {
	out := make([]*model.User, 0, len(items))
	for _, u := range items {
		out = append(out, toModelUser(u))
	}
	return out
}

func buildUserConnection(items []*iamv1.User, startOffset int, pageInfo *model.PageInfo, total int) *model.UserConnection {
	nodes := toModelUsers(items)
	edges := make([]*model.UserEdge, 0, len(nodes))
	for i, n := range nodes {
		edges = append(edges, &model.UserEdge{
			Cursor: EncodeCursor(startOffset + i),
			Node:   n,
		})
	}
	return &model.UserConnection{
		Edges:      edges,
		Nodes:      nodes,
		PageInfo:   pageInfo,
		TotalCount: total,
	}
}

func toModelUserIdentity(i *iamv1.UserIdentity) *model.UserIdentity {
	if i == nil {
		return nil
	}
	return &model.UserIdentity{
		ID:         EncodeGID(ResourceUserIdentity, i.GetId()),
		RawID:      i.GetId(),
		UserID:     EncodeGID(ResourceUser, i.GetUserId()),
		Kind:       fromProtoIdentityKind(i.GetKind()),
		Value:      i.GetValue(),
		IsVerified: i.GetIsVerified(),
		IsPrimary:  i.GetIsPrimary(),
		VerifiedAt: toLocalTimePtr(i.GetVerifiedAt()),
		CreatedAt:  toLocalTime(i.GetCreatedAt()),
		UpdatedAt:  toLocalTime(i.GetUpdatedAt()),
	}
}

func buildUserIdentityConnection(items []*iamv1.UserIdentity, startOffset int, pageInfo *model.PageInfo, total int) *model.UserIdentityConnection {
	nodes := make([]*model.UserIdentity, 0, len(items))
	edges := make([]*model.UserIdentityEdge, 0, len(items))
	for i, item := range items {
		m := toModelUserIdentity(item)
		nodes = append(nodes, m)
		edges = append(edges, &model.UserIdentityEdge{
			Cursor: EncodeCursor(startOffset + i),
			Node:   m,
		})
	}
	return &model.UserIdentityConnection{
		Edges:      edges,
		Nodes:      nodes,
		PageInfo:   pageInfo,
		TotalCount: total,
	}
}

func toModelSession(s *iamv1.Session) *model.Session {
	if s == nil {
		return nil
	}
	return &model.Session{
		ID:                EncodeGID(ResourceSession, s.GetId()),
		RawID:             s.GetId(),
		UserID:            EncodeGID(ResourceUser, s.GetUserId()),
		MfaVerified:       s.GetMfaVerified(),
		MfaVerifiedAt:     toLocalTimePtr(s.GetMfaVerifiedAt()),
		IPAddress:         s.GetIpAddress(),
		UserAgent:         s.GetUserAgent(),
		AuthenticatedAt:   toLocalTime(s.GetAuthenticatedAt()),
		LastActiveAt:      toLocalTime(s.GetLastActiveAt()),
		IdleExpiresAt:     toLocalTime(s.GetIdleExpiresAt()),
		AbsoluteExpiresAt: toLocalTime(s.GetAbsoluteExpiresAt()),
		EndedAt:           toLocalTimePtr(s.GetEndedAt()),
		EndReason:         fromProtoSessionEndReasonPtr(s.GetEndReason()),
		CreatedAt:         toLocalTime(s.GetCreatedAt()),
		Generation:        int(s.GetGeneration()),
	}
}

func toModelSessions(items []*iamv1.Session) []*model.Session {
	out := make([]*model.Session, 0, len(items))
	for _, s := range items {
		out = append(out, toModelSession(s))
	}
	return out
}

func buildSessionConnection(items []*iamv1.Session, startOffset int, pageInfo *model.PageInfo, total int) *model.SessionConnection {
	nodes := toModelSessions(items)
	edges := make([]*model.SessionEdge, 0, len(nodes))
	for i, n := range nodes {
		edges = append(edges, &model.SessionEdge{
			Cursor: EncodeCursor(startOffset + i),
			Node:   n,
		})
	}
	return &model.SessionConnection{
		Edges:      edges,
		Nodes:      nodes,
		PageInfo:   pageInfo,
		TotalCount: total,
	}
}

func toModelRole(r *iamv1.Role) *model.Role {
	if r == nil {
		return nil
	}
	return &model.Role{
		ID:                  EncodeGID(ResourceRole, r.GetId()),
		RawID:               r.GetId(),
		Code:                r.GetCode(),
		Name:                r.GetName(),
		Description:         r.GetDescription(),
		ManagementDomain:    r.GetManagementDomain(),
		DefaultRedirectPath: r.GetDefaultRedirectPath(),
		RedirectPriority:    int(r.GetRedirectPriority()),
		IsSystem:            r.GetIsSystem(),
		IsActive:            r.GetIsActive(),
		CreatedAt:           toLocalTime(r.GetCreatedAt()),
		UpdatedAt:           toLocalTime(r.GetUpdatedAt()),
	}
}

func buildRoleConnection(items []*iamv1.Role, startOffset int, pageInfo *model.PageInfo, total int) *model.RoleConnection {
	nodes := make([]*model.Role, 0, len(items))
	edges := make([]*model.RoleEdge, 0, len(items))
	for i, r := range items {
		m := toModelRole(r)
		nodes = append(nodes, m)
		edges = append(edges, &model.RoleEdge{
			Cursor: EncodeCursor(startOffset + i),
			Node:   m,
		})
	}
	return &model.RoleConnection{
		Edges:      edges,
		Nodes:      nodes,
		PageInfo:   pageInfo,
		TotalCount: total,
	}
}

func toModelPermission(p *iamv1.Permission) *model.Permission {
	if p == nil {
		return nil
	}
	return &model.Permission{
		ID:          EncodeGID(ResourcePermission, p.GetId()),
		RawID:       p.GetId(),
		Code:        p.GetCode(),
		Resource:    p.GetResource(),
		Action:      p.GetAction(),
		Description: p.GetDescription(),
		IsSensitive: p.GetIsSensitive(),
		CreatedAt:   toLocalTime(p.GetCreatedAt()),
	}
}

func buildPermissionConnection(items []*iamv1.Permission, startOffset int, pageInfo *model.PageInfo, total int) *model.PermissionConnection {
	nodes := make([]*model.Permission, 0, len(items))
	edges := make([]*model.PermissionEdge, 0, len(items))
	for i, p := range items {
		m := toModelPermission(p)
		nodes = append(nodes, m)
		edges = append(edges, &model.PermissionEdge{
			Cursor: EncodeCursor(startOffset + i),
			Node:   m,
		})
	}
	return &model.PermissionConnection{
		Edges:      edges,
		Nodes:      nodes,
		PageInfo:   pageInfo,
		TotalCount: total,
	}
}

func toModelGroup(g *iamv1.Group) *model.Group {
	if g == nil {
		return nil
	}
	return &model.Group{
		ID:          EncodeGID(ResourceGroup, g.GetId()),
		RawID:       g.GetId(),
		Code:        g.GetCode(),
		Name:        g.GetName(),
		Description: g.GetDescription(),
		IsActive:    g.GetIsActive(),
		CreatedAt:   toLocalTime(g.GetCreatedAt()),
		UpdatedAt:   toLocalTime(g.GetUpdatedAt()),
	}
}

func buildGroupConnection(items []*iamv1.Group, startOffset int, pageInfo *model.PageInfo, total int) *model.GroupConnection {
	nodes := make([]*model.Group, 0, len(items))
	edges := make([]*model.GroupEdge, 0, len(items))
	for i, g := range items {
		m := toModelGroup(g)
		nodes = append(nodes, m)
		edges = append(edges, &model.GroupEdge{
			Cursor: EncodeCursor(startOffset + i),
			Node:   m,
		})
	}
	return &model.GroupConnection{
		Edges:      edges,
		Nodes:      nodes,
		PageInfo:   pageInfo,
		TotalCount: total,
	}
}

func buildRolePermissionConnection(roleID string, items []*iamv1.Permission, startOffset int, pageInfo *model.PageInfo, total int) *model.RolePermissionConnection {
	roleGID := EncodeGID(ResourceRole, roleID)
	nodes := make([]*model.Permission, 0, len(items))
	edges := make([]*model.RolePermissionEdge, 0, len(items))
	for i, p := range items {
		m := toModelPermission(p)
		nodes = append(nodes, m)
		edges = append(edges, &model.RolePermissionEdge{
			Cursor:    EncodeCursor(startOffset + i),
			Node:      m,
			RoleID:    roleGID,
			CreatedAt: toLocalTime(p.GetCreatedAt()),
		})
	}
	return &model.RolePermissionConnection{
		Edges:      edges,
		Nodes:      nodes,
		PageInfo:   pageInfo,
		TotalCount: total,
	}
}

func buildGroupMemberConnection(groupID string, items []*iamv1.User, startOffset int, pageInfo *model.PageInfo, total int) *model.GroupMemberConnection {
	groupGID := EncodeGID(ResourceGroup, groupID)
	nodes := toModelUsers(items)
	edges := make([]*model.GroupMemberEdge, 0, len(nodes))
	for i, u := range nodes {
		edges = append(edges, &model.GroupMemberEdge{
			Cursor:    EncodeCursor(startOffset + i),
			Node:      u,
			GroupID:   groupGID,
			CreatedAt: u.CreatedAt,
		})
	}
	return &model.GroupMemberConnection{
		Edges:      edges,
		Nodes:      nodes,
		PageInfo:   pageInfo,
		TotalCount: total,
	}
}

func toModelRolePermissionEdge(rp *iamv1.RolePermission, perm *model.Permission, cursor string) *model.RolePermissionEdge {
	if rp == nil {
		return nil
	}
	return &model.RolePermissionEdge{
		Cursor:    cursor,
		Node:      perm,
		RoleID:    EncodeGID(ResourceRole, rp.GetRoleId()),
		CreatedAt: toLocalTime(rp.GetCreatedAt()),
	}
}

func toModelUserRoleEdge(ur *iamv1.UserRole, role *model.Role, cursor string) *model.UserRoleEdge {
	if ur == nil {
		return nil
	}
	return &model.UserRoleEdge{
		Cursor:       cursor,
		Node:         role,
		AssignmentID: ur.GetId(),
		UserID:       EncodeGID(ResourceUser, ur.GetUserId()),
		Scope:        toModelScope(ur.GetScope()),
		GrantedBy:    EncodeOptionalGID(ResourceUser, ur.GrantedBy),
		ExpiresAt:    toLocalTimePtr(ur.GetExpiresAt()),
		CreatedAt:    toLocalTime(ur.GetCreatedAt()),
	}
}

func toModelGroupMemberEdge(gm *iamv1.GroupMember, user *model.User, cursor string) *model.GroupMemberEdge {
	if gm == nil {
		return nil
	}
	return &model.GroupMemberEdge{
		Cursor:    cursor,
		Node:      user,
		GroupID:   EncodeGID(ResourceGroup, gm.GetGroupId()),
		CreatedAt: toLocalTime(gm.GetCreatedAt()),
	}
}

func toModelGroupRoleEdge(gr *iamv1.GroupRole, role *model.Role, cursor string) *model.GroupRoleEdge {
	if gr == nil {
		return nil
	}
	return &model.GroupRoleEdge{
		Cursor:       cursor,
		Node:         role,
		AssignmentID: gr.GetId(),
		GroupID:      EncodeGID(ResourceGroup, gr.GetGroupId()),
		Scope:        toModelScope(gr.GetScope()),
		GrantedBy:    EncodeOptionalGID(ResourceUser, gr.GrantedBy),
		ExpiresAt:    toLocalTimePtr(gr.GetExpiresAt()),
		CreatedAt:    toLocalTime(gr.GetCreatedAt()),
	}
}

func toModelAccessSnapshot(snap *authz.AccessSnapshot) *model.AccessSnapshot {
	if snap == nil {
		return nil
	}
	roles := make([]*model.RoleGrant, 0, len(snap.Roles))
	for _, r := range snap.Roles {
		roles = append(roles, &model.RoleGrant{
			ID:                  r.ID,
			Code:                r.Code,
			Name:                r.Name,
			Priority:            int(r.Priority),
			DefaultRedirectPath: r.DefaultRedirectPath,
		})
	}
	perms := make([]*model.ScopedPermission, 0, len(snap.Permissions))
	for _, p := range snap.Permissions {
		var scopeID *string
		if strings.TrimSpace(p.Scope.ID) != "" {
			v := strings.TrimSpace(p.Scope.ID)
			scopeID = &v
		}
		perms = append(perms, &model.ScopedPermission{
			PermissionCode: p.Code,
			IsSensitive:    p.IsSensitive,
			Scope: &model.Scope{
				Type: toModelScopeTypeFromString(p.Scope.Type),
				ID:   scopeID,
			},
		})
	}
	return &model.AccessSnapshot{
		PrincipalID:          snap.PrincipalID,
		UserKind:             snap.UserKind,
		Status:               snap.Status,
		MfaEnabled:           snap.MFAEnabled,
		AuthEpoch:            int(snap.AuthEpoch),
		PermRevision:         int(snap.PermRevision),
		RecoveryEpoch:        int(snap.RecoveryEpoch),
		CanAccessIdentity:    snap.CanAccessIdentity,
		DefaultRedirectPath:  snap.DefaultRedirectPath,
		AllowedRedirectPaths: append([]string(nil), snap.AllowedRedirectPaths...),
		Roles:                roles,
		Permissions:          perms,
		SyncedAtUnix:         int(snap.SyncedAtUnix),
		FreshUntilUnix:       int(snap.FreshUntilUnix),
	}
}

func toModelScope(s *iamv1.Scope) *model.Scope {
	if s == nil {
		return &model.Scope{Type: model.ScopeTypeGlobal}
	}
	var id *string
	if s.Id != nil && strings.TrimSpace(*s.Id) != "" {
		v := strings.TrimSpace(*s.Id)
		id = &v
	}
	return &model.Scope{
		Type: toModelScopeTypeFromString(source_connections.FromProtoScopeTypeString(s.GetType())),
		ID:   id,
	}
}

func toModelScopeTypeFromString(raw string) model.ScopeType {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case authz.ScopeGlobal:
		return model.ScopeTypeGlobal
	case "legal_entity":
		return model.ScopeTypeLegalEntity
	case "organization_unit":
		return model.ScopeTypeOrganizationUnit
	case "facility":
		return model.ScopeTypeFacility
	case "supplier_party":
		return model.ScopeTypeSupplierParty
	case "vendor_party":
		return model.ScopeTypeVendorParty
	case "self":
		return model.ScopeTypeSelf
	default:
		return model.ScopeTypeGlobal
	}
}

func toProtoScopeType(t model.ScopeType) iamv1.ScopeType {
	switch t {
	case model.ScopeTypeGlobal:
		return iamv1.ScopeType_SCOPE_TYPE_GLOBAL
	case model.ScopeTypeLegalEntity:
		return iamv1.ScopeType_SCOPE_TYPE_LEGAL_ENTITY
	case model.ScopeTypeOrganizationUnit:
		return iamv1.ScopeType_SCOPE_TYPE_ORGANIZATION_UNIT
	case model.ScopeTypeFacility:
		return iamv1.ScopeType_SCOPE_TYPE_FACILITY
	case model.ScopeTypeSupplierParty:
		return iamv1.ScopeType_SCOPE_TYPE_SUPPLIER_PARTY
	case model.ScopeTypeVendorParty:
		return iamv1.ScopeType_SCOPE_TYPE_VENDOR_PARTY
	case model.ScopeTypeSelf:
		return iamv1.ScopeType_SCOPE_TYPE_SELF
	default:
		return iamv1.ScopeType_SCOPE_TYPE_UNSPECIFIED
	}
}

func fromProtoUserKind(k iamv1.UserKind) model.UserKind {
	switch k {
	case iamv1.UserKind_USER_KIND_CUSTOMER:
		return model.UserKindCustomer
	case iamv1.UserKind_USER_KIND_EMPLOYEE:
		return model.UserKindEmployee
	case iamv1.UserKind_USER_KIND_VENDOR:
		return model.UserKindVendor
	case iamv1.UserKind_USER_KIND_SUPPLIER:
		return model.UserKindSupplier
	case iamv1.UserKind_USER_KIND_SYSTEM:
		return model.UserKindSystem
	default:
		return model.UserKindCustomer
	}
}

func toProtoUserKind(k model.UserKind) iamv1.UserKind {
	switch k {
	case model.UserKindCustomer:
		return iamv1.UserKind_USER_KIND_CUSTOMER
	case model.UserKindEmployee:
		return iamv1.UserKind_USER_KIND_EMPLOYEE
	case model.UserKindVendor:
		return iamv1.UserKind_USER_KIND_VENDOR
	case model.UserKindSupplier:
		return iamv1.UserKind_USER_KIND_SUPPLIER
	case model.UserKindSystem:
		return iamv1.UserKind_USER_KIND_SYSTEM
	default:
		return iamv1.UserKind_USER_KIND_UNSPECIFIED
	}
}

func fromProtoUserStatus(s iamv1.UserStatus) model.UserStatus {
	switch s {
	case iamv1.UserStatus_USER_STATUS_PENDING:
		return model.UserStatusPending
	case iamv1.UserStatus_USER_STATUS_ACTIVE:
		return model.UserStatusActive
	case iamv1.UserStatus_USER_STATUS_SUSPENDED:
		return model.UserStatusSuspended
	case iamv1.UserStatus_USER_STATUS_DISABLED:
		return model.UserStatusDisabled
	default:
		return model.UserStatusPending
	}
}

func toProtoUserStatus(s model.UserStatus) iamv1.UserStatus {
	switch s {
	case model.UserStatusPending:
		return iamv1.UserStatus_USER_STATUS_PENDING
	case model.UserStatusActive:
		return iamv1.UserStatus_USER_STATUS_ACTIVE
	case model.UserStatusSuspended:
		return iamv1.UserStatus_USER_STATUS_SUSPENDED
	case model.UserStatusDisabled:
		return iamv1.UserStatus_USER_STATUS_DISABLED
	default:
		return iamv1.UserStatus_USER_STATUS_UNSPECIFIED
	}
}

func fromProtoIdentityKind(k iamv1.IdentityKind) model.IdentityKind {
	switch k {
	case iamv1.IdentityKind_IDENTITY_KIND_EMAIL:
		return model.IdentityKindEmail
	case iamv1.IdentityKind_IDENTITY_KIND_USERNAME:
		return model.IdentityKindUsername
	default:
		return model.IdentityKindEmail
	}
}

func toProtoIdentityKind(k model.IdentityKind) iamv1.IdentityKind {
	switch k {
	case model.IdentityKindEmail:
		return iamv1.IdentityKind_IDENTITY_KIND_EMAIL
	case model.IdentityKindUsername:
		return iamv1.IdentityKind_IDENTITY_KIND_USERNAME
	default:
		return iamv1.IdentityKind_IDENTITY_KIND_UNSPECIFIED
	}
}

func fromProtoSessionEndReasonPtr(r iamv1.SessionEndReason) *model.SessionEndReason {
	var v model.SessionEndReason
	switch r {
	case iamv1.SessionEndReason_SESSION_END_REASON_LOGOUT:
		v = model.SessionEndReasonLogout
	case iamv1.SessionEndReason_SESSION_END_REASON_IDLE_TIMEOUT:
		v = model.SessionEndReasonIdleTimeout
	case iamv1.SessionEndReason_SESSION_END_REASON_ABSOLUTE_TIMEOUT:
		v = model.SessionEndReasonAbsoluteTimeout
	case iamv1.SessionEndReason_SESSION_END_REASON_PASSWORD_CHANGED:
		v = model.SessionEndReasonPasswordChanged
	case iamv1.SessionEndReason_SESSION_END_REASON_ADMIN_REVOKED:
		v = model.SessionEndReasonAdminRevoked
	case iamv1.SessionEndReason_SESSION_END_REASON_USER_DISABLED:
		v = model.SessionEndReasonUserDisabled
	default:
		return nil
	}
	return &v
}

func toProtoSessionEndReason(r *model.SessionEndReason) iamv1.SessionEndReason {
	if r == nil {
		return iamv1.SessionEndReason_SESSION_END_REASON_ADMIN_REVOKED
	}
	switch *r {
	case model.SessionEndReasonLogout:
		return iamv1.SessionEndReason_SESSION_END_REASON_LOGOUT
	case model.SessionEndReasonIdleTimeout:
		return iamv1.SessionEndReason_SESSION_END_REASON_IDLE_TIMEOUT
	case model.SessionEndReasonAbsoluteTimeout:
		return iamv1.SessionEndReason_SESSION_END_REASON_ABSOLUTE_TIMEOUT
	case model.SessionEndReasonPasswordChanged:
		return iamv1.SessionEndReason_SESSION_END_REASON_PASSWORD_CHANGED
	case model.SessionEndReasonAdminRevoked:
		return iamv1.SessionEndReason_SESSION_END_REASON_ADMIN_REVOKED
	case model.SessionEndReasonUserDisabled:
		return iamv1.SessionEndReason_SESSION_END_REASON_USER_DISABLED
	default:
		return iamv1.SessionEndReason_SESSION_END_REASON_ADMIN_REVOKED
	}
}

func fromProtoVerifyEmailStatus(s iamv1.VerifyEmailStatus) model.VerifyEmailStatus {
	switch s {
	case iamv1.VerifyEmailStatus_VERIFY_EMAIL_STATUS_VERIFIED:
		return model.VerifyEmailStatusVerified
	case iamv1.VerifyEmailStatus_VERIFY_EMAIL_STATUS_EXPIRED:
		return model.VerifyEmailStatusExpired
	case iamv1.VerifyEmailStatus_VERIFY_EMAIL_STATUS_ALREADY_USED:
		return model.VerifyEmailStatusAlreadyUsed
	default:
		return model.VerifyEmailStatusInvalid
	}
}

func derefStr(v *string) string {
	if v == nil {
		return ""
	}
	return strings.TrimSpace(*v)
}

func derefBool(v *bool, fallback bool) bool {
	if v == nil {
		return fallback
	}
	return *v
}

func derefInt32(v *int, fallback int32) int32 {
	if v == nil {
		return fallback
	}
	return int32(*v)
}

func ptrStr(v string) *string {
	return &v
}
