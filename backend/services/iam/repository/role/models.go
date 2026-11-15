// Package role mengimplementasikan repositori PostgreSQL GORM untuk peran RBAC, ikatan izin peran, dan penugasan peran pengguna.
package role

import "time"

type roleModel struct {
	ID                  string    `gorm:"column:id;type:uuid;primaryKey"`
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
}

func (roleModel) TableName() string {
	return "roles"
}

type rolePermissionModel struct {
	RoleID       string    `gorm:"column:role_id;type:uuid;primaryKey"`
	PermissionID string    `gorm:"column:permission_id;type:uuid;primaryKey"`
	CreatedAt    time.Time `gorm:"column:created_at"`
}

func (rolePermissionModel) TableName() string {
	return "role_permissions"
}

type userRoleModel struct {
	ID         string     `gorm:"column:id;type:uuid;primaryKey"`
	UserID     string     `gorm:"column:user_id;type:uuid"`
	RoleID     string     `gorm:"column:role_id;type:uuid"`
	ScopeType  string     `gorm:"column:scope_type"`
	ScopeID    *string    `gorm:"column:scope_id;type:uuid"`
	AssignedBy *string    `gorm:"column:assigned_by;type:uuid"`
	ExpiresAt  *time.Time `gorm:"column:expires_at"`
	CreatedAt  time.Time  `gorm:"column:created_at"`
}

func (userRoleModel) TableName() string {
	return "user_roles"
}
