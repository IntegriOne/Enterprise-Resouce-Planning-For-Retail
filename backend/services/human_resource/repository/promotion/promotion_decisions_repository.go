// Sumber model adalah baseline SQL HRM; aturan bisnis tetap milik service.
package promotion

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository/pgstore"
)

// PromotionDecisionRepository menyediakan kueri bertipe untuk hr_promotion_decisions.
type PromotionDecisionRepository struct {
	db *gorm.DB
}

var _ domainrepo.PromotionDecisionRepository = (*PromotionDecisionRepository)(nil)

// NewPromotionDecisionRepository memvalidasi koneksi dan menyiapkan repository.
func NewPromotionDecisionRepository(db *gorm.DB) (*PromotionDecisionRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &PromotionDecisionRepository{db: db}, nil
}

// Operasi bernama objek menyediakan kontrak keputusan promosi yang eksplisit.
func (r *PromotionDecisionRepository) InsertPromotionDecisions(ctx context.Context, value *entity.PromotionDecision) error {
	return r.Insert(ctx, value)
}
func (r *PromotionDecisionRepository) UpdatePromotionDecisions(ctx context.Context, value *entity.PromotionDecision) error {
	return r.Update(ctx, value)
}
func (r *PromotionDecisionRepository) DeletePromotionDecisionsByUUID(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
func (r *PromotionDecisionRepository) FindPromotionDecisionsByUUID(ctx context.Context, id string) (*entity.PromotionDecision, error) {
	return r.FindByID(ctx, id)
}
func (r *PromotionDecisionRepository) FindPromotionDecisionsByUUIDForUpdate(ctx context.Context, id string) (*entity.PromotionDecision, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *PromotionDecisionRepository) GetAllPromotionDecisions(ctx context.Context, limit, offset int) ([]entity.PromotionDecision, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *PromotionDecisionRepository) CountPromotionDecisions(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *PromotionDecisionRepository) SelectPromotionDecisionsWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.PromotionDecision, error) {
	return r.SelectWhere(ctx, filters, limit, offset)
}
func (r *PromotionDecisionRepository) CountPromotionDecisionsWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, filters)
}
func (r *PromotionDecisionRepository) FindPromotionDecisionsByRecommendationID(ctx context.Context, value string, limit, offset int) ([]entity.PromotionDecision, error) {
	return r.FindByRecommendationID(ctx, value, limit, offset)
}
func (r *PromotionDecisionRepository) FindPromotionDecisionsByStepNumber(ctx context.Context, value int32, limit, offset int) ([]entity.PromotionDecision, error) {
	return r.FindByStepNumber(ctx, value, limit, offset)
}
func (r *PromotionDecisionRepository) FindPromotionDecisionsByDecidedByPrincipalID(ctx context.Context, value string, limit, offset int) ([]entity.PromotionDecision, error) {
	return r.FindByDecidedByPrincipalID(ctx, value, limit, offset)
}
func (r *PromotionDecisionRepository) FindPromotionDecisionsByRationaleDocumentVersionReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.PromotionDecision, error) {
	return r.FindByRationaleDocumentVersionReferenceID(ctx, value, limit, offset)
}
func (r *PromotionDecisionRepository) FindPromotionDecisionsByIdempotencyKey(ctx context.Context, value string, limit, offset int) ([]entity.PromotionDecision, error) {
	return r.FindByIdempotencyKey(ctx, value, limit, offset)
}

func (r *PromotionDecisionRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.PromotionDecision{})
}

// Insert menyimpan keputusan sesuai constraint database.
func (r *PromotionDecisionRepository) Insert(ctx context.Context, value *entity.PromotionDecision) error {
	if value == nil {
		return hrerrors.ErrInvalidInput
	}
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Clauses(clause.Returning{}).Create(value).Error)
}
func (r *PromotionDecisionRepository) Create(ctx context.Context, value *entity.PromotionDecision) error {
	return r.Insert(ctx, value)
}
func (r *PromotionDecisionRepository) FindByID(ctx context.Context, key string) (*entity.PromotionDecision, error) {
	return r.find(ctx, key, false)
}
func (r *PromotionDecisionRepository) SelectByID(ctx context.Context, key string) (*entity.PromotionDecision, error) {
	return r.FindByID(ctx, key)
}
func (r *PromotionDecisionRepository) FindByIDForUpdate(ctx context.Context, key string) (*entity.PromotionDecision, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}
func (r *PromotionDecisionRepository) find(ctx context.Context, key string, lock bool) (*entity.PromotionDecision, error) {
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
	var value entity.PromotionDecision
	if err := db.Where("id = ?", key).First(&value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return &value, nil
}
func (r *PromotionDecisionRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.PromotionDecision, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}
func (r *PromotionDecisionRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.PromotionDecision, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *PromotionDecisionRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.PromotionDecision, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere memfilter sebelum membaca halaman berurutan.
func (r *PromotionDecisionRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.PromotionDecision, error) {
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
	rows := make([]entity.PromotionDecision, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}
func (r *PromotionDecisionRepository) Count(ctx context.Context) (int64, error) {
	return r.CountWhere(ctx, nil)
}
func (r *PromotionDecisionRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
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
func (r *PromotionDecisionRepository) Update(ctx context.Context, value *entity.PromotionDecision) error {
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
func (r *PromotionDecisionRepository) DeleteByID(ctx context.Context, key string) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.PromotionDecision{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}
func (r *PromotionDecisionRepository) Delete(ctx context.Context, key string) error {
	return r.DeleteByID(ctx, key)
}

// FindByRecommendationID mencari kecocokan recommendation_id dengan paginasi dan urutan primary key.
func (r *PromotionDecisionRepository) FindByRecommendationID(ctx context.Context, value string, limit, offset int) ([]entity.PromotionDecision, error) {
	return r.SelectBy(ctx, "recommendation_id", value, limit, offset)
}

// FindByStepNumber mencari kecocokan step_number dengan paginasi dan urutan primary key.
func (r *PromotionDecisionRepository) FindByStepNumber(ctx context.Context, value int32, limit, offset int) ([]entity.PromotionDecision, error) {
	return r.SelectBy(ctx, "step_number", value, limit, offset)
}

// FindByDecidedByPrincipalID mencari kecocokan decided_by_principal_id dengan paginasi dan urutan primary key.
func (r *PromotionDecisionRepository) FindByDecidedByPrincipalID(ctx context.Context, value string, limit, offset int) ([]entity.PromotionDecision, error) {
	return r.SelectBy(ctx, "decided_by_principal_id", value, limit, offset)
}

// FindByRationaleDocumentVersionReferenceID mencari kecocokan rationale_document_version_reference_id dengan paginasi dan urutan primary key.
func (r *PromotionDecisionRepository) FindByRationaleDocumentVersionReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.PromotionDecision, error) {
	return r.SelectBy(ctx, "rationale_document_version_reference_id", value, limit, offset)
}

// FindByIdempotencyKey mencari kecocokan idempotency_key dengan paginasi dan urutan primary key.
func (r *PromotionDecisionRepository) FindByIdempotencyKey(ctx context.Context, value string, limit, offset int) ([]entity.PromotionDecision, error) {
	return r.SelectBy(ctx, "idempotency_key", value, limit, offset)
}
