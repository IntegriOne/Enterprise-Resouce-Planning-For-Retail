package user

import (
	"context"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
	"github.com/integrione/erp-retail/backend/services/iam/repository/pgstore"
)

type snapshotAssignmentRow struct {
	ID                  string    `gorm:"column:id"`
	Code                string    `gorm:"column:code"`
	Name                string    `gorm:"column:name"`
	Description         *string   `gorm:"column:description"`
	ManagementDomain    string    `gorm:"column:management_domain"`
	DefaultRedirectPath string    `gorm:"column:default_redirect_path"`
	RedirectPriority    int       `gorm:"column:redirect_priority"`
	IsSystem            bool      `gorm:"column:is_system"`
	IsActive            bool      `gorm:"column:is_active"`
	CreatedAt           time.Time `gorm:"column:created_at"`
	UpdatedAt           time.Time `gorm:"column:updated_at"`
	ScopeType           string    `gorm:"column:scope_type"`
	ScopeID             *string   `gorm:"column:scope_id"`
}

type snapshotRolePermRow struct {
	RoleID         string `gorm:"column:role_id"`
	PermissionCode string `gorm:"column:permission_code"`
	IsSensitive    bool   `gorm:"column:is_sensitive"`
}

// SelectAccessSnapshotByUserID menyusun snapshot otorisasi terpusat meliputi status user, MFA, peran langsung dan grup aktif, serta izin berlingkup.
func (r *UserRepository) SelectAccessSnapshotByUserID(ctx context.Context, userID string, now time.Time) (*entity.AccessSnapshot, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	uid, err := pgstore.ValidateUUIDParam("user_id", userID)
	if err != nil {
		return nil, err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return nil, err
	}

	var uModel userModel
	if err := db.Where("id = ? AND deleted_at IS NULL", uid).Take(&uModel).Error; err != nil {
		return nil, pgstore.MapError(err, iamerrors.ErrUserNotFound)
	}
	uDomain, err := userToDomain(&uModel)
	if err != nil {
		return nil, err
	}

	var mfaCount int64
	if err := db.Model(&userTOTPFactorModel{}).
		Where("user_id = ? AND is_enabled = ? AND confirmed_at IS NOT NULL", uid, true).
		Limit(1).
		Count(&mfaCount).Error; err != nil {
		return nil, pgstore.MapError(err, nil)
	}

	checkTime := pgstore.MustLocalNow(now)
	assignmentRows, err := queryEffectiveRoleAssignments(db, uid, checkTime)
	if err != nil {
		return nil, pgstore.MapError(err, nil)
	}

	roles, roleIDs, roleScopes, err := extractRolesAndScopes(assignmentRows)
	if err != nil {
		return nil, err
	}

	scopedPerms := []entity.ScopedPermission{}
	if len(roleIDs) > 0 {
		var permRows []snapshotRolePermRow
		if err := db.Table("role_permissions rp").
			Select("rp.role_id::text AS role_id, p.code AS permission_code, p.is_sensitive").
			Joins("JOIN permissions p ON p.id = rp.permission_id").
			Where("rp.role_id IN ?", roleIDs).
			Order("p.code ASC").
			Scan(&permRows).Error; err != nil {
			return nil, pgstore.MapError(err, nil)
		}
		scopedPerms = buildScopedPermissions(permRows, roleScopes)
	}

	var ctrlRow struct {
		RecoveryEpoch int64 `gorm:"column:recovery_epoch"`
	}
	if err := db.Table("iam_security_control").Select("recovery_epoch").Where("id = ?", 1).Take(&ctrlRow).Error; err != nil || ctrlRow.RecoveryEpoch < 1 {
		ctrlRow.RecoveryEpoch = 1
	}

	return &entity.AccessSnapshot{
		UserID:        uDomain.ID,
		UserKind:      uDomain.Kind,
		UserStatus:    uDomain.Status,
		MFAEnabled:    mfaCount > 0,
		AuthEpoch:     uDomain.AuthEpoch,
		PermRevision:  uDomain.PermRevision,
		RecoveryEpoch: ctrlRow.RecoveryEpoch,
		Roles:         roles,
		Permissions:   scopedPerms,
		CheckedAt:     checkTime,
	}, nil
}

func queryEffectiveRoleAssignments(db *gorm.DB, uid string, checkTime time.Time) ([]snapshotAssignmentRow, error) {
	var rows []snapshotAssignmentRow
	err := db.Raw(`
		SELECT DISTINCT
			r.id::text AS id,
			r.code,
			r.name,
			r.description,
			r.management_domain,
			r.default_redirect_path,
			r.redirect_priority,
			r.is_system,
			r.is_active,
			r.created_at,
			r.updated_at,
			src.scope_type::text AS scope_type,
			src.scope_id::text AS scope_id
		FROM (
			SELECT ur.role_id, ur.scope_type, ur.scope_id
			FROM user_roles ur
			WHERE ur.user_id = ? AND (ur.expires_at IS NULL OR ur.expires_at > ?)
			UNION ALL
			SELECT gr.role_id, gr.scope_type, gr.scope_id
			FROM group_members gm
			JOIN groups g ON g.id = gm.group_id
			JOIN group_roles gr ON gr.group_id = g.id
			WHERE gm.user_id = ? AND g.is_active = true AND (gr.expires_at IS NULL OR gr.expires_at > ?)
		) AS src
		JOIN roles r ON r.id = src.role_id
		WHERE r.is_active = true
		ORDER BY r.redirect_priority ASC, r.code ASC
	`, uid, checkTime, uid, checkTime).Scan(&rows).Error
	return rows, err
}

func extractRolesAndScopes(rows []snapshotAssignmentRow) ([]entity.Role, []string, map[string][]entity.Scope, error) {
	roles := make([]entity.Role, 0, len(rows))
	roleIDs := make([]string, 0, len(rows))
	roleScopes := make(map[string][]entity.Scope, len(rows))
	seenRoles := make(map[string]struct{}, len(rows))

	for i := range rows {
		row := &rows[i]
		roleID := strings.ToLower(strings.TrimSpace(row.ID))
		scope, err := pgstore.ToScope(row.ScopeType, row.ScopeID)
		if err != nil {
			return nil, nil, nil, err
		}
		roleScopes[roleID] = append(roleScopes[roleID], scope)

		if _, exists := seenRoles[roleID]; exists {
			continue
		}
		seenRoles[roleID] = struct{}{}
		roleIDs = append(roleIDs, roleID)
		roles = append(roles, entity.Role{
			ID:                  roleID,
			Code:                row.Code,
			Name:                row.Name,
			Description:         pgstore.DerefString(row.Description),
			ManagementDomain:    row.ManagementDomain,
			DefaultRedirectPath: row.DefaultRedirectPath,
			RedirectPriority:    row.RedirectPriority,
			IsSystem:            row.IsSystem,
			IsActive:            row.IsActive,
			CreatedAt:           pgstore.NormalizeLocalTime(row.CreatedAt),
			UpdatedAt:           pgstore.NormalizeLocalTime(row.UpdatedAt),
		})
	}
	return roles, roleIDs, roleScopes, nil
}

func buildScopedPermissions(permRows []snapshotRolePermRow, roleScopes map[string][]entity.Scope) []entity.ScopedPermission {
	result := make([]entity.ScopedPermission, 0, len(permRows))
	seen := make(map[string]struct{}, len(permRows))

	for _, pr := range permRows {
		roleID := strings.ToLower(strings.TrimSpace(pr.RoleID))
		for _, sc := range roleScopes[roleID] {
			scopeIDKey := ""
			if sc.ID != nil {
				scopeIDKey = *sc.ID
			}
			dedupKey := pr.PermissionCode + "|" + sc.Type.String() + "|" + scopeIDKey
			if _, exists := seen[dedupKey]; exists {
				continue
			}
			seen[dedupKey] = struct{}{}
			result = append(result, entity.ScopedPermission{
				PermissionCode: pr.PermissionCode,
				IsSensitive:    pr.IsSensitive,
				Scope:          sc.Clone(),
			})
		}
	}
	return result
}
