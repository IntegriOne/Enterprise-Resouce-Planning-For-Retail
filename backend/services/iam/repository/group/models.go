// Package group mengimplementasikan repositori PostgreSQL GORM untuk grup pengguna, keanggotaan grup, dan penugasan peran grup.
package group

import "time"

type groupModel struct {
	ID          string    `gorm:"column:id;type:uuid;primaryKey"`
	Code        string    `gorm:"column:code"`
	Name        string    `gorm:"column:name"`
	Description *string   `gorm:"column:description"`
	IsActive    bool      `gorm:"column:is_active"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

func (groupModel) TableName() string {
	return "groups"
}

type groupMemberModel struct {
	ID        string    `gorm:"column:id;type:uuid;primaryKey;->"`
	GroupID   string    `gorm:"column:group_id;type:uuid"`
	UserID    string    `gorm:"column:user_id;type:uuid"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (groupMemberModel) TableName() string {
	return "group_members"
}

type groupRoleModel struct {
	ID         string     `gorm:"column:id;type:uuid;primaryKey"`
	GroupID    string     `gorm:"column:group_id;type:uuid"`
	RoleID     string     `gorm:"column:role_id;type:uuid"`
	ScopeType  string     `gorm:"column:scope_type"`
	ScopeID    *string    `gorm:"column:scope_id;type:uuid"`
	AssignedBy *string    `gorm:"column:assigned_by;type:uuid"`
	ExpiresAt  *time.Time `gorm:"column:expires_at"`
	CreatedAt  time.Time  `gorm:"column:created_at"`
}

func (groupRoleModel) TableName() string {
	return "group_roles"
}

type groupUserModel struct {
	ID          string     `gorm:"column:id"`
	UserKind    string     `gorm:"column:user_kind"`
	Status      string     `gorm:"column:status"`
	DisplayName string     `gorm:"column:display_name"`
	AvatarURL   *string    `gorm:"column:avatar_url"`
	Locale      string     `gorm:"column:locale"`
	Timezone    string     `gorm:"column:timezone"`
	LastLoginAt *time.Time `gorm:"column:last_login_at"`
	CreatedAt   time.Time  `gorm:"column:created_at"`
	UpdatedAt   time.Time  `gorm:"column:updated_at"`
	DeletedAt   *time.Time `gorm:"column:deleted_at"`
}
