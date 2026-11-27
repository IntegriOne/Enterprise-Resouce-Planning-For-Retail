// Sumber model adalah baseline SQL HRM; aturan bisnis tetap milik service.
package workforceplanning

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository/pgstore"
)

// WorkforcePlanLineRepository menyediakan kueri bertipe untuk hr_workforce_plan_lines.
type WorkforcePlanLineRepository struct {
	db *gorm.DB
}

var _ domainrepo.WorkforcePlanLineRepository = (*WorkforcePlanLineRepository)(nil)

// Method bernama objek memperjelas tabel detail rencana dan kolom pencarian.
func (r *WorkforcePlanLineRepository) InsertWorkforcePlanLines(ctx context.Context, v *entity.WorkforcePlanLine) error {
	return r.Insert(ctx, v)
}
func (r *WorkforcePlanLineRepository) UpdateWorkforcePlanLines(ctx context.Context, v *entity.WorkforcePlanLine) error {
	return r.Update(ctx, v)
}
func (r *WorkforcePlanLineRepository) DeleteWorkforcePlanLinesByUUID(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
func (r *WorkforcePlanLineRepository) FindWorkforcePlanLinesByUUID(ctx context.Context, id string) (*entity.WorkforcePlanLine, error) {
	return r.FindByID(ctx, id)
}
func (r *WorkforcePlanLineRepository) FindWorkforcePlanLinesByUUIDForUpdate(ctx context.Context, id string) (*entity.WorkforcePlanLine, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *WorkforcePlanLineRepository) GetAllWorkforcePlanLines(ctx context.Context, limit, offset int) ([]entity.WorkforcePlanLine, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *WorkforcePlanLineRepository) CountWorkforcePlanLines(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *WorkforcePlanLineRepository) SelectWorkforcePlanLinesWhere(ctx context.Context, f []domainrepo.Filter, limit, offset int) ([]entity.WorkforcePlanLine, error) {
	return r.SelectWhere(ctx, f, limit, offset)
}
func (r *WorkforcePlanLineRepository) CountWorkforcePlanLinesWhere(ctx context.Context, f []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, f)
}
func (r *WorkforcePlanLineRepository) FindWorkforcePlanLinesByWorkforcePlanID(ctx context.Context, v string, limit, offset int) ([]entity.WorkforcePlanLine, error) {
	return r.FindByWorkforcePlanID(ctx, v, limit, offset)
}
func (r *WorkforcePlanLineRepository) FindWorkforcePlanLinesByLineNumber(ctx context.Context, v int32, limit, offset int) ([]entity.WorkforcePlanLine, error) {
	return r.FindByLineNumber(ctx, v, limit, offset)
}
func (r *WorkforcePlanLineRepository) FindWorkforcePlanLinesByOrganizationUnitReferenceID(ctx context.Context, v string, limit, offset int) ([]entity.WorkforcePlanLine, error) {
	return r.FindByOrganizationUnitReferenceID(ctx, v, limit, offset)
}
func (r *WorkforcePlanLineRepository) FindWorkforcePlanLinesByJobID(ctx context.Context, v string, limit, offset int) ([]entity.WorkforcePlanLine, error) {
	return r.FindByJobID(ctx, v, limit, offset)
}
func (r *WorkforcePlanLineRepository) FindWorkforcePlanLinesByWorkLocationReferenceID(ctx context.Context, v *string, limit, offset int) ([]entity.WorkforcePlanLine, error) {
	return r.FindByWorkLocationReferenceID(ctx, v, limit, offset)
}
func (r *WorkforcePlanLineRepository) FindWorkforcePlanLinesByCurrencyCode(ctx context.Context, v *string, limit, offset int) ([]entity.WorkforcePlanLine, error) {
	return r.FindByCurrencyCode(ctx, v, limit, offset)
}
func (r *WorkforcePlanLineRepository) FindWorkforcePlanLinesByActionCode(ctx context.Context, v string, limit, offset int) ([]entity.WorkforcePlanLine, error) {
	return r.FindByActionCode(ctx, v, limit, offset)
}

// NewWorkforcePlanLineRepository memvalidasi koneksi dan menyiapkan repository.
func NewWorkforcePlanLineRepository(db *gorm.DB) (*WorkforcePlanLineRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &WorkforcePlanLineRepository{db: db}, nil
}

func (r *WorkforcePlanLineRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.WorkforcePlanLine{})
}

// Insert menyimpan baris kebutuhan tenaga kerja dalam transaksi pemanggil.
func (r *WorkforcePlanLineRepository) Insert(ctx context.Context, value *entity.WorkforcePlanLine) error {
	if value == nil {
		return hrerrors.ErrInvalidInput
	}
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Clauses(clause.Returning{}).Create(value).Error)
}
func (r *WorkforcePlanLineRepository) Create(ctx context.Context, value *entity.WorkforcePlanLine) error {
	return r.Insert(ctx, value)
}
func (r *WorkforcePlanLineRepository) FindByID(ctx context.Context, key string) (*entity.WorkforcePlanLine, error) {
	return r.find(ctx, key, false)
}
func (r *WorkforcePlanLineRepository) SelectByID(ctx context.Context, key string) (*entity.WorkforcePlanLine, error) {
	return r.FindByID(ctx, key)
}
func (r *WorkforcePlanLineRepository) FindByIDForUpdate(ctx context.Context, key string) (*entity.WorkforcePlanLine, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}
func (r *WorkforcePlanLineRepository) find(ctx context.Context, key string, lock bool) (*entity.WorkforcePlanLine, error) {
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
	value := new(entity.WorkforcePlanLine)
	if err := db.Where("id = ?", key).First(value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return value, nil
}
func (r *WorkforcePlanLineRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.WorkforcePlanLine, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}
func (r *WorkforcePlanLineRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.WorkforcePlanLine, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *WorkforcePlanLineRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.WorkforcePlanLine, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// planLineReadScope mengambil badan hukum dari rencana induk ketika query memerlukan scope.
func planLineReadScope(db *gorm.DB, filters []domainrepo.Filter) (*gorm.DB, error) {
	for _, filter := range filters {
		if filter.Column == "plan_legal_entity_reference_id" {
			db = db.Table(`(SELECT line.*, plan.legal_entity_reference_id AS plan_legal_entity_reference_id
				FROM hr_workforce_plan_lines AS line
				JOIN hr_workforce_plans AS plan ON plan.id = line.workforce_plan_id) AS hr_workforce_plan_lines`)
			break
		}
	}
	return pgstore.ApplyFilters(db, filters)
}

// SelectWhere menambahkan filter sebelum pagination dengan urutan ID yang deterministik.
func (r *WorkforcePlanLineRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.WorkforcePlanLine, error) {
	limit, offset, err := pgstore.NormalizePage(limit, offset)
	if err != nil {
		return nil, err
	}
	db, err := r.query(ctx)
	if err != nil {
		return nil, err
	}
	db, err = planLineReadScope(db, filters)
	if err != nil {
		return nil, err
	}
	rows := make([]entity.WorkforcePlanLine, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}
func (r *WorkforcePlanLineRepository) Count(ctx context.Context) (int64, error) {
	return r.CountWhere(ctx, nil)
}
func (r *WorkforcePlanLineRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	db, err := r.query(ctx)
	if err != nil {
		return 0, err
	}
	db, err = planLineReadScope(db, filters)
	if err != nil {
		return 0, err
	}
	var total int64
	err = db.Count(&total).Error
	return total, pgstore.MapError(err)
}

// Update mengganti snapshot non-key termasuk nilai nol/NULL tanpa upsert.
func (r *WorkforcePlanLineRepository) Update(ctx context.Context, value *entity.WorkforcePlanLine) error {
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
func (r *WorkforcePlanLineRepository) DeleteByID(ctx context.Context, key string) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.WorkforcePlanLine{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}
func (r *WorkforcePlanLineRepository) Delete(ctx context.Context, key string) error {
	return r.DeleteByID(ctx, key)
}

// FindByWorkforcePlanID mencari kecocokan workforce_plan_id dengan paginasi dan urutan primary key.
func (r *WorkforcePlanLineRepository) FindByWorkforcePlanID(ctx context.Context, value string, limit, offset int) ([]entity.WorkforcePlanLine, error) {
	return r.SelectBy(ctx, "workforce_plan_id", value, limit, offset)
}

// FindByLineNumber mencari kecocokan line_number dengan paginasi dan urutan primary key.
func (r *WorkforcePlanLineRepository) FindByLineNumber(ctx context.Context, value int32, limit, offset int) ([]entity.WorkforcePlanLine, error) {
	return r.SelectBy(ctx, "line_number", value, limit, offset)
}

// FindByOrganizationUnitReferenceID mencari kecocokan organization_unit_reference_id dengan paginasi dan urutan primary key.
func (r *WorkforcePlanLineRepository) FindByOrganizationUnitReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.WorkforcePlanLine, error) {
	return r.SelectBy(ctx, "organization_unit_reference_id", value, limit, offset)
}

// FindByJobID mencari kecocokan job_id dengan paginasi dan urutan primary key.
func (r *WorkforcePlanLineRepository) FindByJobID(ctx context.Context, value string, limit, offset int) ([]entity.WorkforcePlanLine, error) {
	return r.SelectBy(ctx, "job_id", value, limit, offset)
}

// FindByWorkLocationReferenceID mencari kecocokan work_location_reference_id dengan paginasi dan urutan primary key.
func (r *WorkforcePlanLineRepository) FindByWorkLocationReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.WorkforcePlanLine, error) {
	return r.SelectBy(ctx, "work_location_reference_id", value, limit, offset)
}

// FindByCurrencyCode mencari kecocokan currency_code dengan paginasi dan urutan primary key.
func (r *WorkforcePlanLineRepository) FindByCurrencyCode(ctx context.Context, value *string, limit, offset int) ([]entity.WorkforcePlanLine, error) {
	return r.SelectBy(ctx, "currency_code", value, limit, offset)
}

// FindByActionCode mencari kecocokan action_code dengan paginasi dan urutan primary key.
func (r *WorkforcePlanLineRepository) FindByActionCode(ctx context.Context, value string, limit, offset int) ([]entity.WorkforcePlanLine, error) {
	return r.SelectBy(ctx, "action_code", value, limit, offset)
}
