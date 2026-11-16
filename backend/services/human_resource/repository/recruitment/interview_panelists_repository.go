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

// InterviewPanelistRepository menyediakan kueri bertipe untuk hr_interview_panelists.
type InterviewPanelistRepository struct {
	db *gorm.DB
}

var _ domainrepo.InterviewPanelistRepository = (*InterviewPanelistRepository)(nil)

// Method bernama objek memperjelas tabel panelis dan kolom pencarian.
func (r *InterviewPanelistRepository) InsertInterviewPanelists(ctx context.Context, v *entity.InterviewPanelist) error {
	return r.Insert(ctx, v)
}
func (r *InterviewPanelistRepository) UpdateInterviewPanelists(ctx context.Context, v *entity.InterviewPanelist) error {
	return r.Update(ctx, v)
}
func (r *InterviewPanelistRepository) DeleteInterviewPanelistsByUUID(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
func (r *InterviewPanelistRepository) FindInterviewPanelistsByUUID(ctx context.Context, id string) (*entity.InterviewPanelist, error) {
	return r.FindByID(ctx, id)
}
func (r *InterviewPanelistRepository) FindInterviewPanelistsByUUIDForUpdate(ctx context.Context, id string) (*entity.InterviewPanelist, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *InterviewPanelistRepository) GetAllInterviewPanelists(ctx context.Context, limit, offset int) ([]entity.InterviewPanelist, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *InterviewPanelistRepository) CountInterviewPanelists(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *InterviewPanelistRepository) SelectInterviewPanelistsWhere(ctx context.Context, f []domainrepo.Filter, limit, offset int) ([]entity.InterviewPanelist, error) {
	return r.SelectWhere(ctx, f, limit, offset)
}
func (r *InterviewPanelistRepository) CountInterviewPanelistsWhere(ctx context.Context, f []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, f)
}
func (r *InterviewPanelistRepository) FindInterviewPanelistsByInterviewID(ctx context.Context, v string, limit, offset int) ([]entity.InterviewPanelist, error) {
	return r.FindByInterviewID(ctx, v, limit, offset)
}
func (r *InterviewPanelistRepository) FindInterviewPanelistsByInterviewerPrincipalID(ctx context.Context, v string, limit, offset int) ([]entity.InterviewPanelist, error) {
	return r.FindByInterviewerPrincipalID(ctx, v, limit, offset)
}
func (r *InterviewPanelistRepository) FindInterviewPanelistsByScorecardTemplateReferenceID(ctx context.Context, v *string, limit, offset int) ([]entity.InterviewPanelist, error) {
	return r.FindByScorecardTemplateReferenceID(ctx, v, limit, offset)
}

// NewInterviewPanelistRepository memvalidasi koneksi dan menyiapkan repository.
func NewInterviewPanelistRepository(db *gorm.DB) (*InterviewPanelistRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &InterviewPanelistRepository{db: db}, nil
}

func (r *InterviewPanelistRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.InterviewPanelist{})
}

// Insert menyimpan keanggotaan panel dengan primary key yang diberikan service.
func (r *InterviewPanelistRepository) Insert(ctx context.Context, value *entity.InterviewPanelist) error {
	if value == nil {
		return hrerrors.ErrInvalidInput
	}
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Clauses(clause.Returning{}).Create(value).Error)
}
func (r *InterviewPanelistRepository) Create(ctx context.Context, value *entity.InterviewPanelist) error {
	return r.Insert(ctx, value)
}
func (r *InterviewPanelistRepository) FindByID(ctx context.Context, key string) (*entity.InterviewPanelist, error) {
	return r.find(ctx, key, false)
}
func (r *InterviewPanelistRepository) SelectByID(ctx context.Context, key string) (*entity.InterviewPanelist, error) {
	return r.FindByID(ctx, key)
}
func (r *InterviewPanelistRepository) FindByIDForUpdate(ctx context.Context, key string) (*entity.InterviewPanelist, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}
func (r *InterviewPanelistRepository) find(ctx context.Context, key string, lock bool) (*entity.InterviewPanelist, error) {
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
	value := new(entity.InterviewPanelist)
	if err := db.Where("id = ?", key).First(value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return value, nil
}
func (r *InterviewPanelistRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.InterviewPanelist, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}
func (r *InterviewPanelistRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.InterviewPanelist, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *InterviewPanelistRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.InterviewPanelist, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere membatasi panel sesuai filter sebelum pembatasan halaman.
func (r *InterviewPanelistRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.InterviewPanelist, error) {
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
	rows := make([]entity.InterviewPanelist, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}
func (r *InterviewPanelistRepository) Count(ctx context.Context) (int64, error) {
	return r.CountWhere(ctx, nil)
}
func (r *InterviewPanelistRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
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

// Update mengganti snapshot non-key termasuk nilai nol dan NULL tanpa upsert.
func (r *InterviewPanelistRepository) Update(ctx context.Context, value *entity.InterviewPanelist) error {
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
func (r *InterviewPanelistRepository) DeleteByID(ctx context.Context, key string) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.InterviewPanelist{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}
func (r *InterviewPanelistRepository) Delete(ctx context.Context, key string) error {
	return r.DeleteByID(ctx, key)
}

// FindByInterviewID mencari kecocokan interview_id dengan paginasi dan urutan primary key.
func (r *InterviewPanelistRepository) FindByInterviewID(ctx context.Context, value string, limit, offset int) ([]entity.InterviewPanelist, error) {
	return r.SelectBy(ctx, "interview_id", value, limit, offset)
}

// FindByInterviewerPrincipalID mencari kecocokan interviewer_principal_id dengan paginasi dan urutan primary key.
func (r *InterviewPanelistRepository) FindByInterviewerPrincipalID(ctx context.Context, value string, limit, offset int) ([]entity.InterviewPanelist, error) {
	return r.SelectBy(ctx, "interviewer_principal_id", value, limit, offset)
}

// FindByScorecardTemplateReferenceID mencari kecocokan scorecard_template_reference_id dengan paginasi dan urutan primary key.
func (r *InterviewPanelistRepository) FindByScorecardTemplateReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.InterviewPanelist, error) {
	return r.SelectBy(ctx, "scorecard_template_reference_id", value, limit, offset)
}
