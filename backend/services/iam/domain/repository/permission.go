package repository

import (
	"context"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
)

// PermissionRepository mendefinisikan operasi baca dan tulis untuk katalog izin atomik pada tabel permissions.
type PermissionRepository interface {
	SelectByID(ctx context.Context, id string) (*entity.Permission, error)
	SelectByCode(ctx context.Context, code string) (*entity.Permission, error)
	SelectList(ctx context.Context, filter PermissionFilter) (PageResult[*entity.Permission], error)
	SelectByRoleID(ctx context.Context, roleID string) ([]*entity.Permission, error)
	Insert(ctx context.Context, permission *entity.Permission) error
	Delete(ctx context.Context, id string) error
}
