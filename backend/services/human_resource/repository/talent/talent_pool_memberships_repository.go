// Sumber model adalah baseline SQL HRM; aturan bisnis tetap milik service.
package talent

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository/pgstore"
)

// TalentPoolMembershipRepository menyediakan kueri bertipe untuk hr_talent_pool_memberships.
type TalentPoolMembershipRepository struct {
	db *gorm.DB
}

var _ domainrepo.TalentPoolMembershipRepository = (*TalentPoolMembershipRepository)(nil)

// NewTalentPoolMembershipRepository memvalidasi koneksi dan menyiapkan repository.
func NewTalentPoolMembershipRepository(db *gorm.DB) (*TalentPoolMembershipRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &TalentPoolMembershipRepository{db: db}, nil
}

// Operasi bernama objek menyediakan kontrak keanggotaan pool yang eksplisit.
func (r *TalentPoolMembershipRepository) InsertTalentPoolMemberships(ctx context.Context, value *entity.TalentPoolMembership) error {
	return r.Insert(ctx, value)
}
func (r *TalentPoolMembershipRepository) UpdateTalentPoolMemberships(ctx context.Context, value *entity.TalentPoolMembership) error {
	return r.Update(ctx, value)
}
func (r *TalentPoolMembershipRepository) DeleteTalentPoolMembershipsByUUID(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
func (r *TalentPoolMembershipRepository) FindTalentPoolMembershipsByUUID(ctx context.Context, id string) (*entity.TalentPoolMembership, error) {
	return r.FindByID(ctx, id)
}
func (r *TalentPoolMembershipRepository) FindTalentPoolMembershipsByUUIDForUpdate(ctx context.Context, id string) (*entity.TalentPoolMembership, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *TalentPoolMembershipRepository) GetAllTalentPoolMemberships(ctx context.Context, limit, offset int) ([]entity.TalentPoolMembership, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *TalentPoolMembershipRepository) CountTalentPoolMemberships(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *TalentPoolMembershipRepository) SelectTalentPoolMembershipsWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.TalentPoolMembership, error) {
	return r.SelectWhere(ctx, filters, limit, offset)
}
func (r *TalentPoolMembershipRepository) CountTalentPoolMembershipsWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, filters)
}
func (r *TalentPoolMembershipRepository) FindTalentPoolMembershipsByTalentPoolID(ctx context.Context, value string, limit, offset int) ([]entity.TalentPoolMembership, error) {
	return r.FindByTalentPoolID(ctx, value, limit, offset)
}
func (r *TalentPoolMembershipRepository) FindTalentPoolMembershipsByWorkerID(ctx context.Context, value string, limit, offset int) ([]entity.TalentPoolMembership, error) {
	return r.FindByWorkerID(ctx, value, limit, offset)
}
func (r *TalentPoolMembershipRepository) FindTalentPoolMembershipsByAssessmentID(ctx context.Context, value string, limit, offset int) ([]entity.TalentPoolMembership, error) {
	return r.FindByAssessmentID(ctx, value, limit, offset)
}
func (r *TalentPoolMembershipRepository) FindTalentPoolMembershipsByNominatedByPrincipalID(ctx context.Context, value string, limit, offset int) ([]entity.TalentPoolMembership, error) {
	return r.FindByNominatedByPrincipalID(ctx, value, limit, offset)
}
func (r *TalentPoolMembershipRepository) FindTalentPoolMembershipsByStatus(ctx context.Context, value string, limit, offset int) ([]entity.TalentPoolMembership, error) {
	return r.FindByStatus(ctx, value, limit, offset)
}
func (r *TalentPoolMembershipRepository) FindTalentPoolMembershipsByApprovedByPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.TalentPoolMembership, error) {
	return r.FindByApprovedByPrincipalID(ctx, value, limit, offset)
}
func (r *TalentPoolMembershipRepository) FindTalentPoolMembershipsByExitReasonCode(ctx context.Context, value *string, limit, offset int) ([]entity.TalentPoolMembership, error) {
	return r.FindByExitReasonCode(ctx, value, limit, offset)
}

func (r *TalentPoolMembershipRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.TalentPoolMembership{})
}

// Insert menyimpan keanggotaan sesuai constraint database.
func (r *TalentPoolMembershipRepository) Insert(ctx context.Context, value *entity.TalentPoolMembership) error {
	if value == nil {
		return hrerrors.ErrInvalidInput
	}
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Clauses(clause.Returning{}).Create(value).Error)
}
func (r *TalentPoolMembershipRepository) Create(ctx context.Context, value *entity.TalentPoolMembership) error {
	return r.Insert(ctx, value)
}
func (r *TalentPoolMembershipRepository) FindByID(ctx context.Context, key string) (*entity.TalentPoolMembership, error) {
	return r.find(ctx, key, false)
}
func (r *TalentPoolMembershipRepository) SelectByID(ctx context.Context, key string) (*entity.TalentPoolMembership, error) {
	return r.FindByID(ctx, key)
}
func (r *TalentPoolMembershipRepository) FindByIDForUpdate(ctx context.Context, key string) (*entity.TalentPoolMembership, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}
func (r *TalentPoolMembershipRepository) find(ctx context.Context, key string, lock bool) (*entity.TalentPoolMembership, error) {
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
	var value entity.TalentPoolMembership
	if err := db.Where("id = ?", key).First(&value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return &value, nil
}
func (r *TalentPoolMembershipRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.TalentPoolMembership, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}
func (r *TalentPoolMembershipRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.TalentPoolMembership, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *TalentPoolMembershipRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.TalentPoolMembership, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere memfilter sebelum membaca halaman berurutan.
func (r *TalentPoolMembershipRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.TalentPoolMembership, error) {
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
	rows := make([]entity.TalentPoolMembership, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}
func (r *TalentPoolMembershipRepository) Count(ctx context.Context) (int64, error) {
	return r.CountWhere(ctx, nil)
}
func (r *TalentPoolMembershipRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
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
func (r *TalentPoolMembershipRepository) Update(ctx context.Context, value *entity.TalentPoolMembership) error {
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
func (r *TalentPoolMembershipRepository) DeleteByID(ctx context.Context, key string) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.TalentPoolMembership{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}
func (r *TalentPoolMembershipRepository) Delete(ctx context.Context, key string) error {
	return r.DeleteByID(ctx, key)
}

// FindByTalentPoolID mencari kecocokan talent_pool_id dengan paginasi dan urutan primary key.
func (r *TalentPoolMembershipRepository) FindByTalentPoolID(ctx context.Context, value string, limit, offset int) ([]entity.TalentPoolMembership, error) {
	return r.SelectBy(ctx, "talent_pool_id", value, limit, offset)
}

// FindByWorkerID mencari kecocokan worker_id dengan paginasi dan urutan primary key.
func (r *TalentPoolMembershipRepository) FindByWorkerID(ctx context.Context, value string, limit, offset int) ([]entity.TalentPoolMembership, error) {
	return r.SelectBy(ctx, "worker_id", value, limit, offset)
}

// FindByAssessmentID mencari kecocokan assessment_id dengan paginasi dan urutan primary key.
func (r *TalentPoolMembershipRepository) FindByAssessmentID(ctx context.Context, value string, limit, offset int) ([]entity.TalentPoolMembership, error) {
	return r.SelectBy(ctx, "assessment_id", value, limit, offset)
}

// FindByNominatedByPrincipalID mencari kecocokan nominated_by_principal_id dengan paginasi dan urutan primary key.
func (r *TalentPoolMembershipRepository) FindByNominatedByPrincipalID(ctx context.Context, value string, limit, offset int) ([]entity.TalentPoolMembership, error) {
	return r.SelectBy(ctx, "nominated_by_principal_id", value, limit, offset)
}

// FindByStatus mencari kecocokan status dengan paginasi dan urutan primary key.
func (r *TalentPoolMembershipRepository) FindByStatus(ctx context.Context, value string, limit, offset int) ([]entity.TalentPoolMembership, error) {
	return r.SelectBy(ctx, "status", value, limit, offset)
}

// FindByApprovedByPrincipalID mencari kecocokan approved_by_principal_id dengan paginasi dan urutan primary key.
func (r *TalentPoolMembershipRepository) FindByApprovedByPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.TalentPoolMembership, error) {
	return r.SelectBy(ctx, "approved_by_principal_id", value, limit, offset)
}

// FindByExitReasonCode mencari kecocokan exit_reason_code dengan paginasi dan urutan primary key.
func (r *TalentPoolMembershipRepository) FindByExitReasonCode(ctx context.Context, value *string, limit, offset int) ([]entity.TalentPoolMembership, error) {
	return r.SelectBy(ctx, "exit_reason_code", value, limit, offset)
}
