package repository

import (
	"context"
	"time"

	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/entity"
)

// OrganizationUnitRepository mendefinisikan operasi persistensi murni untuk tabel organization_units.
type OrganizationUnitRepository interface {
	SelectByID(ctx context.Context, id string) (*entity.OrganizationUnit, error)
	SelectByIDForUpdate(ctx context.Context, id string) (*entity.OrganizationUnit, error)
	SelectByIDs(ctx context.Context, ids []string) ([]*entity.OrganizationUnit, error)
	SelectByEntityAndCode(ctx context.Context, legalEntityID, unitCode string) (*entity.OrganizationUnit, error)
	SelectByHeadPersonID(ctx context.Context, headPersonID string, page PageRequest) (PageResult[*entity.OrganizationUnit], error)
	SelectByPrimaryLocationID(ctx context.Context, primaryLocationID string, page PageRequest) (PageResult[*entity.OrganizationUnit], error)
	SelectList(ctx context.Context, filter OrganizationUnitFilter) (PageResult[*entity.OrganizationUnit], error)
	SelectParentID(ctx context.Context, id string) (*string, error)
	SelectChildrenByParentID(ctx context.Context, parentID string, page PageRequest) (PageResult[*entity.OrganizationUnit], error)
	CountActiveDependents(ctx context.Context, id string) (int64, error)
	CheckCodeExistsInEntity(ctx context.Context, legalEntityID, unitCode, excludeID string) (bool, error)
	Insert(ctx context.Context, unit *entity.OrganizationUnit) error
	Update(ctx context.Context, unit *entity.OrganizationUnit) error
	SoftDeleteByID(ctx context.Context, id string, deletedAt time.Time) error
}
