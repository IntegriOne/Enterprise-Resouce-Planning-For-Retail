package repository

import (
	"context"
	"time"

	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/entity"
)

// LocationRepository mendefinisikan operasi persistensi murni untuk tabel locations.
type LocationRepository interface {
	SelectByID(ctx context.Context, id string) (*entity.Location, error)
	SelectByIDs(ctx context.Context, ids []string) ([]*entity.Location, error)
	SelectByCode(ctx context.Context, locationCode string) (*entity.Location, error)
	SelectByLegalEntityAndKind(ctx context.Context, legalEntityID string, kind entity.LocationKind, activeOnly bool, page PageRequest) (PageResult[*entity.Location], error)
	SelectByAddressID(ctx context.Context, addressID string, page PageRequest) (PageResult[*entity.Location], error)
	SelectList(ctx context.Context, filter LocationFilter) (PageResult[*entity.Location], error)
	SelectParentID(ctx context.Context, id string) (*string, error)
	SelectChildrenByParentID(ctx context.Context, parentID string, page PageRequest) (PageResult[*entity.Location], error)
	CountActiveDependents(ctx context.Context, id string) (int64, error)
	CheckCodeExists(ctx context.Context, locationCode, excludeID string) (bool, error)
	Insert(ctx context.Context, location *entity.Location) error
	Update(ctx context.Context, location *entity.Location) error
	SoftDeleteByID(ctx context.Context, id string, deletedAt time.Time) error
}
