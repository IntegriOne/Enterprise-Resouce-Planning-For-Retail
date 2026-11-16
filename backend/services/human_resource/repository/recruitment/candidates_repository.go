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

// CandidateRepository menyediakan kueri bertipe untuk hr_candidates.
type CandidateRepository struct {
	db *gorm.DB
}

var _ domainrepo.CandidateRepository = (*CandidateRepository)(nil)

// Method bernama objek menjadi API eksplisit data kandidat HRM.
func (r *CandidateRepository) InsertCandidates(ctx context.Context, v *entity.Candidate) error {
	return r.Insert(ctx, v)
}
func (r *CandidateRepository) UpdateCandidates(ctx context.Context, v *entity.Candidate) error {
	return r.Update(ctx, v)
}
func (r *CandidateRepository) DeleteCandidatesByUUID(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
func (r *CandidateRepository) FindCandidatesByUUID(ctx context.Context, id string) (*entity.Candidate, error) {
	return r.FindByID(ctx, id)
}
func (r *CandidateRepository) FindCandidatesByUUIDForUpdate(ctx context.Context, id string) (*entity.Candidate, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *CandidateRepository) GetAllCandidates(ctx context.Context, limit, offset int) ([]entity.Candidate, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *CandidateRepository) CountCandidates(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *CandidateRepository) SelectCandidatesWhere(ctx context.Context, f []domainrepo.Filter, limit, offset int) ([]entity.Candidate, error) {
	return r.SelectWhere(ctx, f, limit, offset)
}
func (r *CandidateRepository) CountCandidatesWhere(ctx context.Context, f []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, f)
}
func (r *CandidateRepository) FindCandidatesByCandidateNumber(ctx context.Context, v string, limit, offset int) ([]entity.Candidate, error) {
	return r.FindByCandidateNumber(ctx, v, limit, offset)
}
func (r *CandidateRepository) FindCandidatesByPersonPartyReferenceID(ctx context.Context, v *string, limit, offset int) ([]entity.Candidate, error) {
	return r.FindByPersonPartyReferenceID(ctx, v, limit, offset)
}
func (r *CandidateRepository) FindCandidatesByContactReferenceID(ctx context.Context, v *string, limit, offset int) ([]entity.Candidate, error) {
	return r.FindByContactReferenceID(ctx, v, limit, offset)
}
func (r *CandidateRepository) FindCandidatesBySourceType(ctx context.Context, v *string, limit, offset int) ([]entity.Candidate, error) {
	return r.FindBySourceType(ctx, v, limit, offset)
}
func (r *CandidateRepository) FindCandidatesByCandidateStatus(ctx context.Context, v string, limit, offset int) ([]entity.Candidate, error) {
	return r.FindByCandidateStatus(ctx, v, limit, offset)
}
func (r *CandidateRepository) FindCandidatesByCreatedByPrincipalID(ctx context.Context, v *string, limit, offset int) ([]entity.Candidate, error) {
	return r.FindByCreatedByPrincipalID(ctx, v, limit, offset)
}

// NewCandidateRepository memvalidasi koneksi dan menyiapkan repository.
func NewCandidateRepository(db *gorm.DB) (*CandidateRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &CandidateRepository{db: db}, nil
}

func (r *CandidateRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.Candidate{})
}

// Insert menyimpan kandidat HRM dengan referensi identitas yang disiapkan service.
func (r *CandidateRepository) Insert(ctx context.Context, value *entity.Candidate) error {
	if value == nil {
		return hrerrors.ErrInvalidInput
	}
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Clauses(clause.Returning{}).Create(value).Error)
}
func (r *CandidateRepository) Create(ctx context.Context, value *entity.Candidate) error {
	return r.Insert(ctx, value)
}
func (r *CandidateRepository) FindByID(ctx context.Context, key string) (*entity.Candidate, error) {
	return r.find(ctx, key, false)
}
func (r *CandidateRepository) SelectByID(ctx context.Context, key string) (*entity.Candidate, error) {
	return r.FindByID(ctx, key)
}
func (r *CandidateRepository) FindByIDForUpdate(ctx context.Context, key string) (*entity.Candidate, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}
func (r *CandidateRepository) find(ctx context.Context, key string, lock bool) (*entity.Candidate, error) {
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
	value := new(entity.Candidate)
	if err := db.Where("id = ?", key).First(value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return value, nil
}
func (r *CandidateRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.Candidate, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}
func (r *CandidateRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.Candidate, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *CandidateRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.Candidate, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere memfilter kandidat sebelum pagination dan mengikat nilai sebagai parameter SQL.
func (r *CandidateRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.Candidate, error) {
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
	rows := make([]entity.Candidate, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}
func (r *CandidateRepository) Count(ctx context.Context) (int64, error) {
	return r.CountWhere(ctx, nil)
}
func (r *CandidateRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
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

// Update mengganti snapshot kandidat non-key, termasuk field nullable, tanpa upsert.
func (r *CandidateRepository) Update(ctx context.Context, value *entity.Candidate) error {
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
func (r *CandidateRepository) DeleteByID(ctx context.Context, key string) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.Candidate{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}
func (r *CandidateRepository) Delete(ctx context.Context, key string) error {
	return r.DeleteByID(ctx, key)
}

// FindByCandidateNumber mencari kecocokan candidate_number dengan paginasi dan urutan primary key.
func (r *CandidateRepository) FindByCandidateNumber(ctx context.Context, value string, limit, offset int) ([]entity.Candidate, error) {
	return r.SelectBy(ctx, "candidate_number", value, limit, offset)
}

// FindByPersonPartyReferenceID mencari kecocokan person_party_reference_id dengan paginasi dan urutan primary key.
func (r *CandidateRepository) FindByPersonPartyReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.Candidate, error) {
	return r.SelectBy(ctx, "person_party_reference_id", value, limit, offset)
}

// FindByContactReferenceID mencari kecocokan contact_reference_id dengan paginasi dan urutan primary key.
func (r *CandidateRepository) FindByContactReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.Candidate, error) {
	return r.SelectBy(ctx, "contact_reference_id", value, limit, offset)
}

// FindBySourceType mencari kecocokan source_type dengan paginasi dan urutan primary key.
func (r *CandidateRepository) FindBySourceType(ctx context.Context, value *string, limit, offset int) ([]entity.Candidate, error) {
	return r.SelectBy(ctx, "source_type", value, limit, offset)
}

// FindByCandidateStatus mencari kecocokan candidate_status dengan paginasi dan urutan primary key.
func (r *CandidateRepository) FindByCandidateStatus(ctx context.Context, value string, limit, offset int) ([]entity.Candidate, error) {
	return r.SelectBy(ctx, "candidate_status", value, limit, offset)
}

// FindByCreatedByPrincipalID mencari kecocokan created_by_principal_id dengan paginasi dan urutan primary key.
func (r *CandidateRepository) FindByCreatedByPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.Candidate, error) {
	return r.SelectBy(ctx, "created_by_principal_id", value, limit, offset)
}
