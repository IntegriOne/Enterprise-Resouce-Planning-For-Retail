// Sumber model adalah baseline SQL HRM; aturan bisnis tetap milik service.
package recruitment

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository/pgstore"
)

// ApplicationStageHistoryRepository menyediakan kueri bertipe untuk hr_application_stage_history.
type ApplicationStageHistoryRepository struct {
	db *gorm.DB
}

var _ domainrepo.ApplicationStageHistoryRepository = (*ApplicationStageHistoryRepository)(nil)

// Method bernama objek mempertahankan nama tabel history dan primary key identity ByID.
func (r *ApplicationStageHistoryRepository) InsertApplicationStageHistory(ctx context.Context, v *entity.ApplicationStageHistory) error {
	return r.Insert(ctx, v)
}
func (r *ApplicationStageHistoryRepository) UpdateApplicationStageHistory(ctx context.Context, v *entity.ApplicationStageHistory) error {
	return r.Update(ctx, v)
}
func (r *ApplicationStageHistoryRepository) DeleteApplicationStageHistoryByID(ctx context.Context, id int64) error {
	return r.DeleteByID(ctx, id)
}
func (r *ApplicationStageHistoryRepository) FindApplicationStageHistoryByID(ctx context.Context, id int64) (*entity.ApplicationStageHistory, error) {
	return r.FindByID(ctx, id)
}
func (r *ApplicationStageHistoryRepository) FindApplicationStageHistoryByIDForUpdate(ctx context.Context, id int64) (*entity.ApplicationStageHistory, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *ApplicationStageHistoryRepository) GetAllApplicationStageHistory(ctx context.Context, limit, offset int) ([]entity.ApplicationStageHistory, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *ApplicationStageHistoryRepository) CountApplicationStageHistory(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *ApplicationStageHistoryRepository) SelectApplicationStageHistoryWhere(ctx context.Context, f []domainrepo.Filter, limit, offset int) ([]entity.ApplicationStageHistory, error) {
	return r.SelectWhere(ctx, f, limit, offset)
}
func (r *ApplicationStageHistoryRepository) CountApplicationStageHistoryWhere(ctx context.Context, f []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, f)
}
func (r *ApplicationStageHistoryRepository) FindApplicationStageHistoryByApplicationID(ctx context.Context, v string, limit, offset int) ([]entity.ApplicationStageHistory, error) {
	return r.FindByApplicationID(ctx, v, limit, offset)
}
func (r *ApplicationStageHistoryRepository) FindApplicationStageHistoryByPreviousStatus(ctx context.Context, v *entity.ApplicationStatus, limit, offset int) ([]entity.ApplicationStageHistory, error) {
	return r.FindByPreviousStatus(ctx, v, limit, offset)
}
func (r *ApplicationStageHistoryRepository) FindApplicationStageHistoryByNewStatus(ctx context.Context, v entity.ApplicationStatus, limit, offset int) ([]entity.ApplicationStageHistory, error) {
	return r.FindByNewStatus(ctx, v, limit, offset)
}
func (r *ApplicationStageHistoryRepository) FindApplicationStageHistoryByStageCode(ctx context.Context, v *string, limit, offset int) ([]entity.ApplicationStageHistory, error) {
	return r.FindByStageCode(ctx, v, limit, offset)
}
func (r *ApplicationStageHistoryRepository) FindApplicationStageHistoryByChangedByPrincipalID(ctx context.Context, v *string, limit, offset int) ([]entity.ApplicationStageHistory, error) {
	return r.FindByChangedByPrincipalID(ctx, v, limit, offset)
}
func (r *ApplicationStageHistoryRepository) FindApplicationStageHistoryByReasonCode(ctx context.Context, v *string, limit, offset int) ([]entity.ApplicationStageHistory, error) {
	return r.FindByReasonCode(ctx, v, limit, offset)
}

// NewApplicationStageHistoryRepository memvalidasi koneksi dan menyiapkan repository.
func NewApplicationStageHistoryRepository(db *gorm.DB) (*ApplicationStageHistoryRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &ApplicationStageHistoryRepository{db: db}, nil
}

func (r *ApplicationStageHistoryRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.ApplicationStageHistory{})
}

// Insert menyimpan riwayat tahap dan mengembalikan ID identity dari PostgreSQL.
func (r *ApplicationStageHistoryRepository) Insert(ctx context.Context, value *entity.ApplicationStageHistory) error {
	if value == nil {
		return hrerrors.ErrInvalidInput
	}
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Clauses(clause.Returning{}).Create(value).Error)
}
func (r *ApplicationStageHistoryRepository) Create(ctx context.Context, value *entity.ApplicationStageHistory) error {
	return r.Insert(ctx, value)
}
func (r *ApplicationStageHistoryRepository) FindByID(ctx context.Context, key int64) (*entity.ApplicationStageHistory, error) {
	return r.find(ctx, key, false)
}
func (r *ApplicationStageHistoryRepository) SelectByID(ctx context.Context, key int64) (*entity.ApplicationStageHistory, error) {
	return r.FindByID(ctx, key)
}
func (r *ApplicationStageHistoryRepository) FindByIDForUpdate(ctx context.Context, key int64) (*entity.ApplicationStageHistory, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}
func (r *ApplicationStageHistoryRepository) find(ctx context.Context, key int64, lock bool) (*entity.ApplicationStageHistory, error) {
	db, err := r.query(ctx)
	if err != nil {
		return nil, err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return nil, err
	}
	if lock {
		db = db.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	value := new(entity.ApplicationStageHistory)
	if err := db.Where("id = ?", key).First(value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return value, nil
}
func (r *ApplicationStageHistoryRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.ApplicationStageHistory, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}
func (r *ApplicationStageHistoryRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.ApplicationStageHistory, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *ApplicationStageHistoryRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.ApplicationStageHistory, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere memfilter riwayat sebelum pagination dengan urutan identity ascending.
func (r *ApplicationStageHistoryRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.ApplicationStageHistory, error) {
	limit, offset, err := pgstore.NormalizePage(limit, offset)
	if err != nil {
		return nil, err
	}
	db, err := r.query(ctx)
	if err != nil {
		return nil, err
	}
	db, err = pgstore.ApplyFilters(db, filters)
	if err != nil {
		return nil, err
	}
	rows := make([]entity.ApplicationStageHistory, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}
func (r *ApplicationStageHistoryRepository) Count(ctx context.Context) (int64, error) {
	return r.CountWhere(ctx, nil)
}
func (r *ApplicationStageHistoryRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	db, err := r.query(ctx)
	if err != nil {
		return 0, err
	}
	db, err = pgstore.ApplyFilters(db, filters)
	if err != nil {
		return 0, err
	}
	var total int64
	err = db.Count(&total).Error
	return total, pgstore.MapError(err)
}

// Update tetap tunduk pada constraint riwayat dan tidak mengganti ID identity.
func (r *ApplicationStageHistoryRepository) Update(ctx context.Context, value *entity.ApplicationStageHistory) error {
	if value == nil {
		return hrerrors.ErrInvalidInput
	}
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(value.ID); err != nil {
		return err
	}
	result := db.Where("id = ?", value.ID).Select("*").Omit("id").Updates(value)
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}
func (r *ApplicationStageHistoryRepository) DeleteByID(ctx context.Context, key int64) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.ApplicationStageHistory{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}
func (r *ApplicationStageHistoryRepository) Delete(ctx context.Context, key int64) error {
	return r.DeleteByID(ctx, key)
}

// FindByApplicationID mencari kecocokan application_id dengan paginasi dan urutan primary key.
func (r *ApplicationStageHistoryRepository) FindByApplicationID(ctx context.Context, value string, limit, offset int) ([]entity.ApplicationStageHistory, error) {
	return r.SelectBy(ctx, "application_id", value, limit, offset)
}

// FindByPreviousStatus mencari kecocokan previous_status dengan paginasi dan urutan primary key.
func (r *ApplicationStageHistoryRepository) FindByPreviousStatus(ctx context.Context, value *entity.ApplicationStatus, limit, offset int) ([]entity.ApplicationStageHistory, error) {
	return r.SelectBy(ctx, "previous_status", value, limit, offset)
}

// FindByNewStatus mencari kecocokan new_status dengan paginasi dan urutan primary key.
func (r *ApplicationStageHistoryRepository) FindByNewStatus(ctx context.Context, value entity.ApplicationStatus, limit, offset int) ([]entity.ApplicationStageHistory, error) {
	return r.SelectBy(ctx, "new_status", value, limit, offset)
}

// FindByStageCode mencari kecocokan stage_code dengan paginasi dan urutan primary key.
func (r *ApplicationStageHistoryRepository) FindByStageCode(ctx context.Context, value *string, limit, offset int) ([]entity.ApplicationStageHistory, error) {
	return r.SelectBy(ctx, "stage_code", value, limit, offset)
}

// FindByChangedByPrincipalID mencari kecocokan changed_by_principal_id dengan paginasi dan urutan primary key.
func (r *ApplicationStageHistoryRepository) FindByChangedByPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.ApplicationStageHistory, error) {
	return r.SelectBy(ctx, "changed_by_principal_id", value, limit, offset)
}

// FindByReasonCode mencari kecocokan reason_code dengan paginasi dan urutan primary key.
func (r *ApplicationStageHistoryRepository) FindByReasonCode(ctx context.Context, value *string, limit, offset int) ([]entity.ApplicationStageHistory, error) {
	return r.SelectBy(ctx, "reason_code", value, limit, offset)
}
