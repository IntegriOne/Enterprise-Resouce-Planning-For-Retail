package repository

import (
	"context"
	"time"

	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/entity"
)

// LegalEntityRepository mendefinisikan operasi persistensi murni untuk tabel legal_entities.
type LegalEntityRepository interface {
	SelectByID(ctx context.Context, id string) (*entity.LegalEntity, error)
	SelectByIDs(ctx context.Context, ids []string) ([]*entity.LegalEntity, error)
	SelectByCode(ctx context.Context, entityCode string) (*entity.LegalEntity, error)
	SelectByTaxIDNumber(ctx context.Context, taxIDNumber string) (*entity.LegalEntity, error)
	SelectByNIB(ctx context.Context, nib string, page PageRequest) (PageResult[*entity.LegalEntity], error)
	SelectPrimaryOperating(ctx context.Context) (*entity.LegalEntity, error)
	SelectList(ctx context.Context, filter LegalEntityFilter) (PageResult[*entity.LegalEntity], error)
	SelectParentID(ctx context.Context, id string) (*string, error)
	SelectChildrenByParentID(ctx context.Context, parentID string, page PageRequest) (PageResult[*entity.LegalEntity], error)
	CountActiveDependents(ctx context.Context, id string) (int64, error)
	CheckCodeExists(ctx context.Context, entityCode, excludeID string) (bool, error)
	CheckTaxIDExists(ctx context.Context, taxIDNumber, excludeID string) (bool, error)
	Insert(ctx context.Context, legalEntity *entity.LegalEntity) error
	Update(ctx context.Context, legalEntity *entity.LegalEntity) error
	UpdateClearPrimaryOperatingExcept(ctx context.Context, keepID string, updatedAt time.Time) error
	SoftDeleteByID(ctx context.Context, id string, deletedAt time.Time) error
}
