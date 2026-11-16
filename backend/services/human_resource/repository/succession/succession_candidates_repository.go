// Sumber model adalah baseline SQL HRM; aturan bisnis tetap milik service.
package succession

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository/pgstore"
)

// SuccessionCandidateRepository menyediakan kueri bertipe untuk hr_succession_candidates.
type SuccessionCandidateRepository struct {
	db *gorm.DB
}

var _ domainrepo.SuccessionCandidateRepository = (*SuccessionCandidateRepository)(nil)

// NewSuccessionCandidateRepository memvalidasi koneksi dan menyiapkan repository.
func NewSuccessionCandidateRepository(db *gorm.DB) (*SuccessionCandidateRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &SuccessionCandidateRepository{db: db}, nil
}

// Operasi bernama objek menyediakan kontrak kandidat suksesi yang eksplisit.
func (r *SuccessionCandidateRepository) InsertSuccessionCandidates(ctx context.Context, value *entity.SuccessionCandidate) error {
	return r.Insert(ctx, value)
}
func (r *SuccessionCandidateRepository) UpdateSuccessionCandidates(ctx context.Context, value *entity.SuccessionCandidate) error {
	return r.Update(ctx, value)
}
func (r *SuccessionCandidateRepository) DeleteSuccessionCandidatesByUUID(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
func (r *SuccessionCandidateRepository) FindSuccessionCandidatesByUUID(ctx context.Context, id string) (*entity.SuccessionCandidate, error) {
	return r.FindByID(ctx, id)
}
func (r *SuccessionCandidateRepository) FindSuccessionCandidatesByUUIDForUpdate(ctx context.Context, id string) (*entity.SuccessionCandidate, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *SuccessionCandidateRepository) GetAllSuccessionCandidates(ctx context.Context, limit, offset int) ([]entity.SuccessionCandidate, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *SuccessionCandidateRepository) CountSuccessionCandidates(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *SuccessionCandidateRepository) SelectSuccessionCandidatesWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.SuccessionCandidate, error) {
	return r.SelectWhere(ctx, filters, limit, offset)
}
func (r *SuccessionCandidateRepository) CountSuccessionCandidatesWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, filters)
}
func (r *SuccessionCandidateRepository) FindSuccessionCandidatesBySuccessionPlanID(ctx context.Context, value string, limit, offset int) ([]entity.SuccessionCandidate, error) {
	return r.FindBySuccessionPlanID(ctx, value, limit, offset)
}
func (r *SuccessionCandidateRepository) FindSuccessionCandidatesByCandidateWorkerID(ctx context.Context, value string, limit, offset int) ([]entity.SuccessionCandidate, error) {
	return r.FindByCandidateWorkerID(ctx, value, limit, offset)
}
func (r *SuccessionCandidateRepository) FindSuccessionCandidatesByReadinessCode(ctx context.Context, value string, limit, offset int) ([]entity.SuccessionCandidate, error) {
	return r.FindByReadinessCode(ctx, value, limit, offset)
}
func (r *SuccessionCandidateRepository) FindSuccessionCandidatesByTalentAssessmentID(ctx context.Context, value *string, limit, offset int) ([]entity.SuccessionCandidate, error) {
	return r.FindByTalentAssessmentID(ctx, value, limit, offset)
}
func (r *SuccessionCandidateRepository) FindSuccessionCandidatesByDevelopmentPlanID(ctx context.Context, value *string, limit, offset int) ([]entity.SuccessionCandidate, error) {
	return r.FindByDevelopmentPlanID(ctx, value, limit, offset)
}

func (r *SuccessionCandidateRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.SuccessionCandidate{})
}

// Insert menyimpan kandidat suksesi sesuai constraint database.
func (r *SuccessionCandidateRepository) Insert(ctx context.Context, value *entity.SuccessionCandidate) error {
	if value == nil {
		return hrerrors.ErrInvalidInput
	}
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Clauses(clause.Returning{}).Create(value).Error)
}
func (r *SuccessionCandidateRepository) Create(ctx context.Context, value *entity.SuccessionCandidate) error {
	return r.Insert(ctx, value)
}
func (r *SuccessionCandidateRepository) FindByID(ctx context.Context, key string) (*entity.SuccessionCandidate, error) {
	return r.find(ctx, key, false)
}
func (r *SuccessionCandidateRepository) SelectByID(ctx context.Context, key string) (*entity.SuccessionCandidate, error) {
	return r.FindByID(ctx, key)
}
func (r *SuccessionCandidateRepository) FindByIDForUpdate(ctx context.Context, key string) (*entity.SuccessionCandidate, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}
func (r *SuccessionCandidateRepository) find(ctx context.Context, key string, lock bool) (*entity.SuccessionCandidate, error) {
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
	var value entity.SuccessionCandidate
	if err := db.Where("id = ?", key).First(&value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return &value, nil
}
func (r *SuccessionCandidateRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.SuccessionCandidate, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}
func (r *SuccessionCandidateRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.SuccessionCandidate, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *SuccessionCandidateRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.SuccessionCandidate, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere memfilter sebelum membaca halaman berurutan.
func (r *SuccessionCandidateRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.SuccessionCandidate, error) {
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
	rows := make([]entity.SuccessionCandidate, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}
func (r *SuccessionCandidateRepository) Count(ctx context.Context) (int64, error) {
	return r.CountWhere(ctx, nil)
}
func (r *SuccessionCandidateRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
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
func (r *SuccessionCandidateRepository) Update(ctx context.Context, value *entity.SuccessionCandidate) error {
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
func (r *SuccessionCandidateRepository) DeleteByID(ctx context.Context, key string) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.SuccessionCandidate{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}
func (r *SuccessionCandidateRepository) Delete(ctx context.Context, key string) error {
	return r.DeleteByID(ctx, key)
}

// FindBySuccessionPlanID mencari kecocokan succession_plan_id dengan paginasi dan urutan primary key.
func (r *SuccessionCandidateRepository) FindBySuccessionPlanID(ctx context.Context, value string, limit, offset int) ([]entity.SuccessionCandidate, error) {
	return r.SelectBy(ctx, "succession_plan_id", value, limit, offset)
}

// FindByCandidateWorkerID mencari kecocokan candidate_worker_id dengan paginasi dan urutan primary key.
func (r *SuccessionCandidateRepository) FindByCandidateWorkerID(ctx context.Context, value string, limit, offset int) ([]entity.SuccessionCandidate, error) {
	return r.SelectBy(ctx, "candidate_worker_id", value, limit, offset)
}

// FindByReadinessCode mencari kecocokan readiness_code dengan paginasi dan urutan primary key.
func (r *SuccessionCandidateRepository) FindByReadinessCode(ctx context.Context, value string, limit, offset int) ([]entity.SuccessionCandidate, error) {
	return r.SelectBy(ctx, "readiness_code", value, limit, offset)
}

// FindByTalentAssessmentID mencari kecocokan talent_assessment_id dengan paginasi dan urutan primary key.
func (r *SuccessionCandidateRepository) FindByTalentAssessmentID(ctx context.Context, value *string, limit, offset int) ([]entity.SuccessionCandidate, error) {
	return r.SelectBy(ctx, "talent_assessment_id", value, limit, offset)
}

// FindByDevelopmentPlanID mencari kecocokan development_plan_id dengan paginasi dan urutan primary key.
func (r *SuccessionCandidateRepository) FindByDevelopmentPlanID(ctx context.Context, value *string, limit, offset int) ([]entity.SuccessionCandidate, error) {
	return r.SelectBy(ctx, "development_plan_id", value, limit, offset)
}
