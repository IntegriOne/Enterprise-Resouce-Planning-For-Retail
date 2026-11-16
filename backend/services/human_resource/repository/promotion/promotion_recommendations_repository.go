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

// PromotionRecommendationRepository menyediakan kueri bertipe untuk hr_promotion_recommendations.
type PromotionRecommendationRepository struct {
	db *gorm.DB
}

var _ domainrepo.PromotionRecommendationRepository = (*PromotionRecommendationRepository)(nil)

// NewPromotionRecommendationRepository memvalidasi koneksi dan menyiapkan repository.
func NewPromotionRecommendationRepository(db *gorm.DB) (*PromotionRecommendationRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &PromotionRecommendationRepository{db: db}, nil
}

// Operasi bernama objek menyediakan kontrak rekomendasi promosi yang eksplisit.
func (r *PromotionRecommendationRepository) InsertPromotionRecommendations(ctx context.Context, value *entity.PromotionRecommendation) error {
	return r.Insert(ctx, value)
}
func (r *PromotionRecommendationRepository) UpdatePromotionRecommendations(ctx context.Context, value *entity.PromotionRecommendation) error {
	return r.Update(ctx, value)
}
func (r *PromotionRecommendationRepository) DeletePromotionRecommendationsByUUID(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
func (r *PromotionRecommendationRepository) FindPromotionRecommendationsByUUID(ctx context.Context, id string) (*entity.PromotionRecommendation, error) {
	return r.FindByID(ctx, id)
}
func (r *PromotionRecommendationRepository) FindPromotionRecommendationsByUUIDForUpdate(ctx context.Context, id string) (*entity.PromotionRecommendation, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *PromotionRecommendationRepository) GetAllPromotionRecommendations(ctx context.Context, limit, offset int) ([]entity.PromotionRecommendation, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *PromotionRecommendationRepository) CountPromotionRecommendations(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *PromotionRecommendationRepository) SelectPromotionRecommendationsWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.PromotionRecommendation, error) {
	return r.SelectWhere(ctx, filters, limit, offset)
}
func (r *PromotionRecommendationRepository) CountPromotionRecommendationsWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, filters)
}
func (r *PromotionRecommendationRepository) FindPromotionRecommendationsByRecommendationNumber(ctx context.Context, value string, limit, offset int) ([]entity.PromotionRecommendation, error) {
	return r.FindByRecommendationNumber(ctx, value, limit, offset)
}
func (r *PromotionRecommendationRepository) FindPromotionRecommendationsByWorkerID(ctx context.Context, value string, limit, offset int) ([]entity.PromotionRecommendation, error) {
	return r.FindByWorkerID(ctx, value, limit, offset)
}
func (r *PromotionRecommendationRepository) FindPromotionRecommendationsBySourceAssignmentID(ctx context.Context, value string, limit, offset int) ([]entity.PromotionRecommendation, error) {
	return r.FindBySourceAssignmentID(ctx, value, limit, offset)
}
func (r *PromotionRecommendationRepository) FindPromotionRecommendationsByTargetPositionID(ctx context.Context, value string, limit, offset int) ([]entity.PromotionRecommendation, error) {
	return r.FindByTargetPositionID(ctx, value, limit, offset)
}
func (r *PromotionRecommendationRepository) FindPromotionRecommendationsByAssessmentID(ctx context.Context, value string, limit, offset int) ([]entity.PromotionRecommendation, error) {
	return r.FindByAssessmentID(ctx, value, limit, offset)
}
func (r *PromotionRecommendationRepository) FindPromotionRecommendationsBySupersedesRecommendationID(ctx context.Context, value *string, limit, offset int) ([]entity.PromotionRecommendation, error) {
	return r.FindBySupersedesRecommendationID(ctx, value, limit, offset)
}
func (r *PromotionRecommendationRepository) FindPromotionRecommendationsByRationaleDocumentVersionReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.PromotionRecommendation, error) {
	return r.FindByRationaleDocumentVersionReferenceID(ctx, value, limit, offset)
}
func (r *PromotionRecommendationRepository) FindPromotionRecommendationsByRecommendedByPrincipalID(ctx context.Context, value string, limit, offset int) ([]entity.PromotionRecommendation, error) {
	return r.FindByRecommendedByPrincipalID(ctx, value, limit, offset)
}
func (r *PromotionRecommendationRepository) FindPromotionRecommendationsByStatus(ctx context.Context, value string, limit, offset int) ([]entity.PromotionRecommendation, error) {
	return r.FindByStatus(ctx, value, limit, offset)
}
func (r *PromotionRecommendationRepository) FindPromotionRecommendationsByApprovalDecisionID(ctx context.Context, value *string, limit, offset int) ([]entity.PromotionRecommendation, error) {
	return r.FindByApprovalDecisionID(ctx, value, limit, offset)
}
func (r *PromotionRecommendationRepository) FindPromotionRecommendationsByAppliedAssignmentID(ctx context.Context, value *string, limit, offset int) ([]entity.PromotionRecommendation, error) {
	return r.FindByAppliedAssignmentID(ctx, value, limit, offset)
}
func (r *PromotionRecommendationRepository) FindPromotionRecommendationsByAppliedLifecycleEventID(ctx context.Context, value *int64, limit, offset int) ([]entity.PromotionRecommendation, error) {
	return r.FindByAppliedLifecycleEventID(ctx, value, limit, offset)
}
func (r *PromotionRecommendationRepository) FindPromotionRecommendationsByIdempotencyKey(ctx context.Context, value string, limit, offset int) ([]entity.PromotionRecommendation, error) {
	return r.FindByIdempotencyKey(ctx, value, limit, offset)
}

func (r *PromotionRecommendationRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.PromotionRecommendation{})
}

// Insert menyimpan rekomendasi sesuai constraint database.
func (r *PromotionRecommendationRepository) Insert(ctx context.Context, value *entity.PromotionRecommendation) error {
	if value == nil {
		return hrerrors.ErrInvalidInput
	}
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Clauses(clause.Returning{}).Create(value).Error)
}
func (r *PromotionRecommendationRepository) Create(ctx context.Context, value *entity.PromotionRecommendation) error {
	return r.Insert(ctx, value)
}
func (r *PromotionRecommendationRepository) FindByID(ctx context.Context, key string) (*entity.PromotionRecommendation, error) {
	return r.find(ctx, key, false)
}
func (r *PromotionRecommendationRepository) SelectByID(ctx context.Context, key string) (*entity.PromotionRecommendation, error) {
	return r.FindByID(ctx, key)
}
func (r *PromotionRecommendationRepository) FindByIDForUpdate(ctx context.Context, key string) (*entity.PromotionRecommendation, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}
func (r *PromotionRecommendationRepository) find(ctx context.Context, key string, lock bool) (*entity.PromotionRecommendation, error) {
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
	var value entity.PromotionRecommendation
	if err := db.Where("id = ?", key).First(&value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return &value, nil
}
func (r *PromotionRecommendationRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.PromotionRecommendation, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}
func (r *PromotionRecommendationRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.PromotionRecommendation, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *PromotionRecommendationRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.PromotionRecommendation, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere memfilter sebelum membaca halaman berurutan.
func (r *PromotionRecommendationRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.PromotionRecommendation, error) {
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
	rows := make([]entity.PromotionRecommendation, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}
func (r *PromotionRecommendationRepository) Count(ctx context.Context) (int64, error) {
	return r.CountWhere(ctx, nil)
}
func (r *PromotionRecommendationRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
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
func (r *PromotionRecommendationRepository) Update(ctx context.Context, value *entity.PromotionRecommendation) error {
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
func (r *PromotionRecommendationRepository) DeleteByID(ctx context.Context, key string) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.PromotionRecommendation{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}
func (r *PromotionRecommendationRepository) Delete(ctx context.Context, key string) error {
	return r.DeleteByID(ctx, key)
}

// FindByRecommendationNumber mencari kecocokan recommendation_number dengan paginasi dan urutan primary key.
func (r *PromotionRecommendationRepository) FindByRecommendationNumber(ctx context.Context, value string, limit, offset int) ([]entity.PromotionRecommendation, error) {
	return r.SelectBy(ctx, "recommendation_number", value, limit, offset)
}

// FindByWorkerID mencari kecocokan worker_id dengan paginasi dan urutan primary key.
func (r *PromotionRecommendationRepository) FindByWorkerID(ctx context.Context, value string, limit, offset int) ([]entity.PromotionRecommendation, error) {
	return r.SelectBy(ctx, "worker_id", value, limit, offset)
}

// FindBySourceAssignmentID mencari kecocokan source_assignment_id dengan paginasi dan urutan primary key.
func (r *PromotionRecommendationRepository) FindBySourceAssignmentID(ctx context.Context, value string, limit, offset int) ([]entity.PromotionRecommendation, error) {
	return r.SelectBy(ctx, "source_assignment_id", value, limit, offset)
}

// FindByTargetPositionID mencari kecocokan target_position_id dengan paginasi dan urutan primary key.
func (r *PromotionRecommendationRepository) FindByTargetPositionID(ctx context.Context, value string, limit, offset int) ([]entity.PromotionRecommendation, error) {
	return r.SelectBy(ctx, "target_position_id", value, limit, offset)
}

// FindByAssessmentID mencari kecocokan assessment_id dengan paginasi dan urutan primary key.
func (r *PromotionRecommendationRepository) FindByAssessmentID(ctx context.Context, value string, limit, offset int) ([]entity.PromotionRecommendation, error) {
	return r.SelectBy(ctx, "assessment_id", value, limit, offset)
}

// FindBySupersedesRecommendationID mencari kecocokan supersedes_recommendation_id dengan paginasi dan urutan primary key.
func (r *PromotionRecommendationRepository) FindBySupersedesRecommendationID(ctx context.Context, value *string, limit, offset int) ([]entity.PromotionRecommendation, error) {
	return r.SelectBy(ctx, "supersedes_recommendation_id", value, limit, offset)
}

// FindByRationaleDocumentVersionReferenceID mencari kecocokan rationale_document_version_reference_id dengan paginasi dan urutan primary key.
func (r *PromotionRecommendationRepository) FindByRationaleDocumentVersionReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.PromotionRecommendation, error) {
	return r.SelectBy(ctx, "rationale_document_version_reference_id", value, limit, offset)
}

// FindByRecommendedByPrincipalID mencari kecocokan recommended_by_principal_id dengan paginasi dan urutan primary key.
func (r *PromotionRecommendationRepository) FindByRecommendedByPrincipalID(ctx context.Context, value string, limit, offset int) ([]entity.PromotionRecommendation, error) {
	return r.SelectBy(ctx, "recommended_by_principal_id", value, limit, offset)
}

// FindByStatus mencari kecocokan status dengan paginasi dan urutan primary key.
func (r *PromotionRecommendationRepository) FindByStatus(ctx context.Context, value string, limit, offset int) ([]entity.PromotionRecommendation, error) {
	return r.SelectBy(ctx, "status", value, limit, offset)
}

// FindByApprovalDecisionID mencari kecocokan approval_decision_id dengan paginasi dan urutan primary key.
func (r *PromotionRecommendationRepository) FindByApprovalDecisionID(ctx context.Context, value *string, limit, offset int) ([]entity.PromotionRecommendation, error) {
	return r.SelectBy(ctx, "approval_decision_id", value, limit, offset)
}

// FindByAppliedAssignmentID mencari kecocokan applied_assignment_id dengan paginasi dan urutan primary key.
func (r *PromotionRecommendationRepository) FindByAppliedAssignmentID(ctx context.Context, value *string, limit, offset int) ([]entity.PromotionRecommendation, error) {
	return r.SelectBy(ctx, "applied_assignment_id", value, limit, offset)
}

// FindByAppliedLifecycleEventID mencari kecocokan applied_lifecycle_event_id dengan paginasi dan urutan primary key.
func (r *PromotionRecommendationRepository) FindByAppliedLifecycleEventID(ctx context.Context, value *int64, limit, offset int) ([]entity.PromotionRecommendation, error) {
	return r.SelectBy(ctx, "applied_lifecycle_event_id", value, limit, offset)
}

// FindByIdempotencyKey mencari kecocokan idempotency_key dengan paginasi dan urutan primary key.
func (r *PromotionRecommendationRepository) FindByIdempotencyKey(ctx context.Context, value string, limit, offset int) ([]entity.PromotionRecommendation, error) {
	return r.SelectBy(ctx, "idempotency_key", value, limit, offset)
}
