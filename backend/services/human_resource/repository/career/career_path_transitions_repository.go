// Sumber model adalah baseline SQL HRM; aturan bisnis tetap milik service.
package career

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository/pgstore"
)

// CareerPathTransitionRepository menyediakan kueri bertipe untuk hr_career_path_transitions.
type CareerPathTransitionRepository struct {
	db *gorm.DB
}

var _ domainrepo.CareerPathTransitionRepository = (*CareerPathTransitionRepository)(nil)

// NewCareerPathTransitionRepository memvalidasi koneksi dan menyiapkan repository.
func NewCareerPathTransitionRepository(db *gorm.DB) (*CareerPathTransitionRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &CareerPathTransitionRepository{db: db}, nil
}

// Operasi bernama objek menyediakan kontrak transisi karier yang eksplisit.
func (r *CareerPathTransitionRepository) InsertCareerPathTransitions(ctx context.Context, value *entity.CareerPathTransition) error {
	return r.Insert(ctx, value)
}
func (r *CareerPathTransitionRepository) UpdateCareerPathTransitions(ctx context.Context, value *entity.CareerPathTransition) error {
	return r.Update(ctx, value)
}
func (r *CareerPathTransitionRepository) DeleteCareerPathTransitionsByUUID(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
func (r *CareerPathTransitionRepository) FindCareerPathTransitionsByUUID(ctx context.Context, id string) (*entity.CareerPathTransition, error) {
	return r.FindByID(ctx, id)
}
func (r *CareerPathTransitionRepository) FindCareerPathTransitionsByUUIDForUpdate(ctx context.Context, id string) (*entity.CareerPathTransition, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *CareerPathTransitionRepository) GetAllCareerPathTransitions(ctx context.Context, limit, offset int) ([]entity.CareerPathTransition, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *CareerPathTransitionRepository) CountCareerPathTransitions(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *CareerPathTransitionRepository) SelectCareerPathTransitionsWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.CareerPathTransition, error) {
	return r.SelectWhere(ctx, filters, limit, offset)
}
func (r *CareerPathTransitionRepository) CountCareerPathTransitionsWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, filters)
}
func (r *CareerPathTransitionRepository) FindCareerPathTransitionsByCareerPathID(ctx context.Context, value string, limit, offset int) ([]entity.CareerPathTransition, error) {
	return r.FindByCareerPathID(ctx, value, limit, offset)
}
func (r *CareerPathTransitionRepository) FindCareerPathTransitionsByFromJobID(ctx context.Context, value string, limit, offset int) ([]entity.CareerPathTransition, error) {
	return r.FindByFromJobID(ctx, value, limit, offset)
}
func (r *CareerPathTransitionRepository) FindCareerPathTransitionsByToJobID(ctx context.Context, value string, limit, offset int) ([]entity.CareerPathTransition, error) {
	return r.FindByToJobID(ctx, value, limit, offset)
}
func (r *CareerPathTransitionRepository) FindCareerPathTransitionsByMovementType(ctx context.Context, value string, limit, offset int) ([]entity.CareerPathTransition, error) {
	return r.FindByMovementType(ctx, value, limit, offset)
}
func (r *CareerPathTransitionRepository) FindCareerPathTransitionsByRequirementsDocumentVersionReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.CareerPathTransition, error) {
	return r.FindByRequirementsDocumentVersionReferenceID(ctx, value, limit, offset)
}

func (r *CareerPathTransitionRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.CareerPathTransition{})
}

// Insert menyimpan transisi sesuai constraint database.
func (r *CareerPathTransitionRepository) Insert(ctx context.Context, value *entity.CareerPathTransition) error {
	if value == nil {
		return hrerrors.ErrInvalidInput
	}
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Clauses(clause.Returning{}).Create(value).Error)
}
func (r *CareerPathTransitionRepository) Create(ctx context.Context, value *entity.CareerPathTransition) error {
	return r.Insert(ctx, value)
}
func (r *CareerPathTransitionRepository) FindByID(ctx context.Context, key string) (*entity.CareerPathTransition, error) {
	return r.find(ctx, key, false)
}
func (r *CareerPathTransitionRepository) SelectByID(ctx context.Context, key string) (*entity.CareerPathTransition, error) {
	return r.FindByID(ctx, key)
}
func (r *CareerPathTransitionRepository) FindByIDForUpdate(ctx context.Context, key string) (*entity.CareerPathTransition, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}
func (r *CareerPathTransitionRepository) find(ctx context.Context, key string, lock bool) (*entity.CareerPathTransition, error) {
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
	var value entity.CareerPathTransition
	if err := db.Where("id = ?", key).First(&value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return &value, nil
}
func (r *CareerPathTransitionRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.CareerPathTransition, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}
func (r *CareerPathTransitionRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.CareerPathTransition, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *CareerPathTransitionRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.CareerPathTransition, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere memfilter sebelum membaca halaman berurutan.
func (r *CareerPathTransitionRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.CareerPathTransition, error) {
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
	rows := make([]entity.CareerPathTransition, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}
func (r *CareerPathTransitionRepository) Count(ctx context.Context) (int64, error) {
	return r.CountWhere(ctx, nil)
}
func (r *CareerPathTransitionRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
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

// Update mengganti snapshot non-key, termasuk nol/NULL, tanpa upsert.
func (r *CareerPathTransitionRepository) Update(ctx context.Context, value *entity.CareerPathTransition) error {
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
func (r *CareerPathTransitionRepository) DeleteByID(ctx context.Context, key string) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.CareerPathTransition{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}
func (r *CareerPathTransitionRepository) Delete(ctx context.Context, key string) error {
	return r.DeleteByID(ctx, key)
}

// FindByCareerPathID mencari kecocokan career_path_id dengan paginasi dan urutan primary key.
func (r *CareerPathTransitionRepository) FindByCareerPathID(ctx context.Context, value string, limit, offset int) ([]entity.CareerPathTransition, error) {
	return r.SelectBy(ctx, "career_path_id", value, limit, offset)
}

// FindByFromJobID mencari kecocokan from_job_id dengan paginasi dan urutan primary key.
func (r *CareerPathTransitionRepository) FindByFromJobID(ctx context.Context, value string, limit, offset int) ([]entity.CareerPathTransition, error) {
	return r.SelectBy(ctx, "from_job_id", value, limit, offset)
}

// FindByToJobID mencari kecocokan to_job_id dengan paginasi dan urutan primary key.
func (r *CareerPathTransitionRepository) FindByToJobID(ctx context.Context, value string, limit, offset int) ([]entity.CareerPathTransition, error) {
	return r.SelectBy(ctx, "to_job_id", value, limit, offset)
}

// FindByMovementType mencari kecocokan movement_type dengan paginasi dan urutan primary key.
func (r *CareerPathTransitionRepository) FindByMovementType(ctx context.Context, value string, limit, offset int) ([]entity.CareerPathTransition, error) {
	return r.SelectBy(ctx, "movement_type", value, limit, offset)
}

// FindByRequirementsDocumentVersionReferenceID mencari kecocokan requirements_document_version_reference_id dengan paginasi dan urutan primary key.
func (r *CareerPathTransitionRepository) FindByRequirementsDocumentVersionReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.CareerPathTransition, error) {
	return r.SelectBy(ctx, "requirements_document_version_reference_id", value, limit, offset)
}
