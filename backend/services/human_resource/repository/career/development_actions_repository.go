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

// DevelopmentActionRepository menyediakan kueri bertipe untuk hr_development_actions.
type DevelopmentActionRepository struct {
	db *gorm.DB
}

var _ domainrepo.DevelopmentActionRepository = (*DevelopmentActionRepository)(nil)

// NewDevelopmentActionRepository memvalidasi koneksi dan menyiapkan repository.
func NewDevelopmentActionRepository(db *gorm.DB) (*DevelopmentActionRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &DevelopmentActionRepository{db: db}, nil
}

// Operasi bernama objek menyediakan kontrak tindakan pengembangan yang eksplisit.
func (r *DevelopmentActionRepository) InsertDevelopmentActions(ctx context.Context, value *entity.DevelopmentAction) error {
	return r.Insert(ctx, value)
}
func (r *DevelopmentActionRepository) UpdateDevelopmentActions(ctx context.Context, value *entity.DevelopmentAction) error {
	return r.Update(ctx, value)
}
func (r *DevelopmentActionRepository) DeleteDevelopmentActionsByUUID(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
func (r *DevelopmentActionRepository) FindDevelopmentActionsByUUID(ctx context.Context, id string) (*entity.DevelopmentAction, error) {
	return r.FindByID(ctx, id)
}
func (r *DevelopmentActionRepository) FindDevelopmentActionsByUUIDForUpdate(ctx context.Context, id string) (*entity.DevelopmentAction, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *DevelopmentActionRepository) GetAllDevelopmentActions(ctx context.Context, limit, offset int) ([]entity.DevelopmentAction, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *DevelopmentActionRepository) CountDevelopmentActions(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *DevelopmentActionRepository) SelectDevelopmentActionsWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.DevelopmentAction, error) {
	return r.SelectWhere(ctx, filters, limit, offset)
}
func (r *DevelopmentActionRepository) CountDevelopmentActionsWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, filters)
}
func (r *DevelopmentActionRepository) FindDevelopmentActionsByDevelopmentPlanID(ctx context.Context, value string, limit, offset int) ([]entity.DevelopmentAction, error) {
	return r.FindByDevelopmentPlanID(ctx, value, limit, offset)
}
func (r *DevelopmentActionRepository) FindDevelopmentActionsByActionNumber(ctx context.Context, value int32, limit, offset int) ([]entity.DevelopmentAction, error) {
	return r.FindByActionNumber(ctx, value, limit, offset)
}
func (r *DevelopmentActionRepository) FindDevelopmentActionsByCompetencyID(ctx context.Context, value *string, limit, offset int) ([]entity.DevelopmentAction, error) {
	return r.FindByCompetencyID(ctx, value, limit, offset)
}
func (r *DevelopmentActionRepository) FindDevelopmentActionsByCourseID(ctx context.Context, value *string, limit, offset int) ([]entity.DevelopmentAction, error) {
	return r.FindByCourseID(ctx, value, limit, offset)
}
func (r *DevelopmentActionRepository) FindDevelopmentActionsByActionType(ctx context.Context, value string, limit, offset int) ([]entity.DevelopmentAction, error) {
	return r.FindByActionType(ctx, value, limit, offset)
}
func (r *DevelopmentActionRepository) FindDevelopmentActionsByEvidenceDocumentVersionReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.DevelopmentAction, error) {
	return r.FindByEvidenceDocumentVersionReferenceID(ctx, value, limit, offset)
}

func (r *DevelopmentActionRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.DevelopmentAction{})
}

// Insert menyimpan tindakan sesuai constraint database.
func (r *DevelopmentActionRepository) Insert(ctx context.Context, value *entity.DevelopmentAction) error {
	if value == nil {
		return hrerrors.ErrInvalidInput
	}
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Clauses(clause.Returning{}).Create(value).Error)
}
func (r *DevelopmentActionRepository) Create(ctx context.Context, value *entity.DevelopmentAction) error {
	return r.Insert(ctx, value)
}
func (r *DevelopmentActionRepository) FindByID(ctx context.Context, key string) (*entity.DevelopmentAction, error) {
	return r.find(ctx, key, false)
}
func (r *DevelopmentActionRepository) SelectByID(ctx context.Context, key string) (*entity.DevelopmentAction, error) {
	return r.FindByID(ctx, key)
}
func (r *DevelopmentActionRepository) FindByIDForUpdate(ctx context.Context, key string) (*entity.DevelopmentAction, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}
func (r *DevelopmentActionRepository) find(ctx context.Context, key string, lock bool) (*entity.DevelopmentAction, error) {
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
	var value entity.DevelopmentAction
	if err := db.Where("id = ?", key).First(&value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return &value, nil
}
func (r *DevelopmentActionRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.DevelopmentAction, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}
func (r *DevelopmentActionRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.DevelopmentAction, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *DevelopmentActionRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.DevelopmentAction, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere memfilter sebelum membaca halaman berurutan.
func (r *DevelopmentActionRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.DevelopmentAction, error) {
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
	rows := make([]entity.DevelopmentAction, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}
func (r *DevelopmentActionRepository) Count(ctx context.Context) (int64, error) {
	return r.CountWhere(ctx, nil)
}
func (r *DevelopmentActionRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
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
func (r *DevelopmentActionRepository) Update(ctx context.Context, value *entity.DevelopmentAction) error {
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
func (r *DevelopmentActionRepository) DeleteByID(ctx context.Context, key string) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.DevelopmentAction{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}
func (r *DevelopmentActionRepository) Delete(ctx context.Context, key string) error {
	return r.DeleteByID(ctx, key)
}

// FindByDevelopmentPlanID mencari kecocokan development_plan_id dengan paginasi dan urutan primary key.
func (r *DevelopmentActionRepository) FindByDevelopmentPlanID(ctx context.Context, value string, limit, offset int) ([]entity.DevelopmentAction, error) {
	return r.SelectBy(ctx, "development_plan_id", value, limit, offset)
}

// FindByActionNumber mencari kecocokan action_number dengan paginasi dan urutan primary key.
func (r *DevelopmentActionRepository) FindByActionNumber(ctx context.Context, value int32, limit, offset int) ([]entity.DevelopmentAction, error) {
	return r.SelectBy(ctx, "action_number", value, limit, offset)
}

// FindByCompetencyID mencari kecocokan competency_id dengan paginasi dan urutan primary key.
func (r *DevelopmentActionRepository) FindByCompetencyID(ctx context.Context, value *string, limit, offset int) ([]entity.DevelopmentAction, error) {
	return r.SelectBy(ctx, "competency_id", value, limit, offset)
}

// FindByCourseID mencari kecocokan course_id dengan paginasi dan urutan primary key.
func (r *DevelopmentActionRepository) FindByCourseID(ctx context.Context, value *string, limit, offset int) ([]entity.DevelopmentAction, error) {
	return r.SelectBy(ctx, "course_id", value, limit, offset)
}

// FindByActionType mencari kecocokan action_type dengan paginasi dan urutan primary key.
func (r *DevelopmentActionRepository) FindByActionType(ctx context.Context, value string, limit, offset int) ([]entity.DevelopmentAction, error) {
	return r.SelectBy(ctx, "action_type", value, limit, offset)
}

// FindByEvidenceDocumentVersionReferenceID mencari kecocokan evidence_document_version_reference_id dengan paginasi dan urutan primary key.
func (r *DevelopmentActionRepository) FindByEvidenceDocumentVersionReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.DevelopmentAction, error) {
	return r.SelectBy(ctx, "evidence_document_version_reference_id", value, limit, offset)
}
