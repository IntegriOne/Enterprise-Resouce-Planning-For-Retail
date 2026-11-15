package repository

import (
	"context"
	"time"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
)

// RoleRepository mendefinisikan operasi persistensi untuk tabel roles.
type RoleRepository interface {
	SelectByID(ctx context.Context, id string) (*entity.Role, error)
	SelectByCode(ctx context.Context, code string) (*entity.Role, error)
	SelectList(ctx context.Context, filter RoleFilter) (PageResult[*entity.Role], error)
	SelectEffectiveByUserID(ctx context.Context, userID string, now time.Time) ([]entity.Role, error)
	Insert(ctx context.Context, role *entity.Role) error
	Update(ctx context.Context, role *entity.Role) error
	Delete(ctx context.Context, id string) error
}

// RolePermissionRepository mendefinisikan operasi ikatan izin pada tabel role_permissions.
type RolePermissionRepository interface {
	CheckExists(ctx context.Context, roleID, permissionID string) (bool, error)
	Insert(ctx context.Context, rolePermission *entity.RolePermission) error
	Delete(ctx context.Context, roleID, permissionID string) error
}

// UserRoleRepository mendefinisikan operasi penugasan peran langsung pengguna pada tabel user_roles.
type UserRoleRepository interface {
	SelectByUserID(ctx context.Context, userID string, includeExpired bool, now time.Time) ([]*entity.UserRole, error)
	CheckExists(ctx context.Context, userID, roleID string, scope entity.Scope) (bool, error)
	Insert(ctx context.Context, assignment *entity.UserRole) error
	DeleteByID(ctx context.Context, id string) error
}
