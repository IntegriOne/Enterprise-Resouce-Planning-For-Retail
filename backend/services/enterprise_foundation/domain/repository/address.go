package repository

import (
	"context"

	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/entity"
)

// AddressRepository mendefinisikan operasi persistensi murni untuk tabel addresses.
type AddressRepository interface {
	SelectByID(ctx context.Context, id string) (*entity.Address, error)
	SelectByIDs(ctx context.Context, ids []string) ([]*entity.Address, error)
	SelectList(ctx context.Context, filter AddressFilter) (PageResult[*entity.Address], error)
	CheckReferenced(ctx context.Context, id string) (bool, error)
	Insert(ctx context.Context, address *entity.Address) error
	Update(ctx context.Context, address *entity.Address) error
	DeleteByID(ctx context.Context, id string) error
}
