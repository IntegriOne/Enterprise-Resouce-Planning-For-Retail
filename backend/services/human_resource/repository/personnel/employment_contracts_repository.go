// Sumber model adalah baseline SQL HRM; aturan bisnis tetap milik service.
package personnel

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository/pgstore"
)

// EmploymentContractRepository menyediakan kueri bertipe untuk hr_employment_contracts.
type EmploymentContractRepository struct {
	db *gorm.DB
}

var _ domainrepo.EmploymentContractRepository = (*EmploymentContractRepository)(nil)

// Method bernama objek memperjelas operasi kontrak dan filter kolomnya.
func (r *EmploymentContractRepository) InsertEmploymentContracts(ctx context.Context, v *entity.EmploymentContract) error {
	return r.Insert(ctx, v)
}
func (r *EmploymentContractRepository) UpdateEmploymentContracts(ctx context.Context, v *entity.EmploymentContract) error {
	return r.Update(ctx, v)
}
func (r *EmploymentContractRepository) DeleteEmploymentContractsByUUID(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
func (r *EmploymentContractRepository) FindEmploymentContractsByUUID(ctx context.Context, id string) (*entity.EmploymentContract, error) {
	return r.FindByID(ctx, id)
}
func (r *EmploymentContractRepository) FindEmploymentContractsByUUIDForUpdate(ctx context.Context, id string) (*entity.EmploymentContract, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *EmploymentContractRepository) GetAllEmploymentContracts(ctx context.Context, limit, offset int) ([]entity.EmploymentContract, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *EmploymentContractRepository) CountEmploymentContracts(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *EmploymentContractRepository) SelectEmploymentContractsWhere(ctx context.Context, f []domainrepo.Filter, limit, offset int) ([]entity.EmploymentContract, error) {
	return r.SelectWhere(ctx, f, limit, offset)
}
func (r *EmploymentContractRepository) CountEmploymentContractsWhere(ctx context.Context, f []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, f)
}
func (r *EmploymentContractRepository) FindEmploymentContractsByWorkerID(ctx context.Context, v string, limit, offset int) ([]entity.EmploymentContract, error) {
	return r.FindByWorkerID(ctx, v, limit, offset)
}
func (r *EmploymentContractRepository) FindEmploymentContractsByContractNumber(ctx context.Context, v string, limit, offset int) ([]entity.EmploymentContract, error) {
	return r.FindByContractNumber(ctx, v, limit, offset)
}
func (r *EmploymentContractRepository) FindEmploymentContractsByContractType(ctx context.Context, v string, limit, offset int) ([]entity.EmploymentContract, error) {
	return r.FindByContractType(ctx, v, limit, offset)
}
func (r *EmploymentContractRepository) FindEmploymentContractsByContractStatus(ctx context.Context, v string, limit, offset int) ([]entity.EmploymentContract, error) {
	return r.FindByContractStatus(ctx, v, limit, offset)
}
func (r *EmploymentContractRepository) FindEmploymentContractsByEmployingLegalEntityReferenceID(ctx context.Context, v string, limit, offset int) ([]entity.EmploymentContract, error) {
	return r.FindByEmployingLegalEntityReferenceID(ctx, v, limit, offset)
}
func (r *EmploymentContractRepository) FindEmploymentContractsBySignedDocumentReferenceID(ctx context.Context, v *string, limit, offset int) ([]entity.EmploymentContract, error) {
	return r.FindBySignedDocumentReferenceID(ctx, v, limit, offset)
}
func (r *EmploymentContractRepository) FindEmploymentContractsBySignedDocumentVersionReferenceID(ctx context.Context, v *string, limit, offset int) ([]entity.EmploymentContract, error) {
	return r.FindBySignedDocumentVersionReferenceID(ctx, v, limit, offset)
}
func (r *EmploymentContractRepository) FindEmploymentContractsByWorkCountryCode(ctx context.Context, v string, limit, offset int) ([]entity.EmploymentContract, error) {
	return r.FindByWorkCountryCode(ctx, v, limit, offset)
}
func (r *EmploymentContractRepository) FindEmploymentContractsByGoverningJurisdictionCode(ctx context.Context, v string, limit, offset int) ([]entity.EmploymentContract, error) {
	return r.FindByGoverningJurisdictionCode(ctx, v, limit, offset)
}
func (r *EmploymentContractRepository) FindEmploymentContractsByCollectiveAgreementID(ctx context.Context, v *string, limit, offset int) ([]entity.EmploymentContract, error) {
	return r.FindByCollectiveAgreementID(ctx, v, limit, offset)
}

// NewEmploymentContractRepository memvalidasi koneksi dan menyiapkan repository.
func NewEmploymentContractRepository(db *gorm.DB) (*EmploymentContractRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &EmploymentContractRepository{db: db}, nil
}

func (r *EmploymentContractRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.EmploymentContract{})
}

// Insert menyimpan kontrak dengan snapshot yang sudah disiapkan service.
func (r *EmploymentContractRepository) Insert(ctx context.Context, value *entity.EmploymentContract) error {
	if value == nil {
		return hrerrors.ErrInvalidInput
	}
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Clauses(clause.Returning{}).Create(value).Error)
}
func (r *EmploymentContractRepository) Create(ctx context.Context, value *entity.EmploymentContract) error {
	return r.Insert(ctx, value)
}
func (r *EmploymentContractRepository) FindByID(ctx context.Context, key string) (*entity.EmploymentContract, error) {
	return r.find(ctx, key, false)
}
func (r *EmploymentContractRepository) SelectByID(ctx context.Context, key string) (*entity.EmploymentContract, error) {
	return r.FindByID(ctx, key)
}
func (r *EmploymentContractRepository) FindByIDForUpdate(ctx context.Context, key string) (*entity.EmploymentContract, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}
func (r *EmploymentContractRepository) find(ctx context.Context, key string, lock bool) (*entity.EmploymentContract, error) {
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
	value := new(entity.EmploymentContract)
	if err := db.Where("id = ?", key).First(value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return value, nil
}
func (r *EmploymentContractRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.EmploymentContract, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}
func (r *EmploymentContractRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.EmploymentContract, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *EmploymentContractRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.EmploymentContract, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere menerapkan filter kontrak sebelum pembatasan halaman.
func (r *EmploymentContractRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.EmploymentContract, error) {
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
	rows := make([]entity.EmploymentContract, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}
func (r *EmploymentContractRepository) Count(ctx context.Context) (int64, error) {
	return r.CountWhere(ctx, nil)
}
func (r *EmploymentContractRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
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

// Update mengganti seluruh field non-key termasuk false, nol dan NULL tanpa upsert.
func (r *EmploymentContractRepository) Update(ctx context.Context, value *entity.EmploymentContract) error {
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
func (r *EmploymentContractRepository) DeleteByID(ctx context.Context, key string) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.EmploymentContract{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}
func (r *EmploymentContractRepository) Delete(ctx context.Context, key string) error {
	return r.DeleteByID(ctx, key)
}

// FindByWorkerID mencari kecocokan worker_id dengan paginasi dan urutan primary key.
func (r *EmploymentContractRepository) FindByWorkerID(ctx context.Context, value string, limit, offset int) ([]entity.EmploymentContract, error) {
	return r.SelectBy(ctx, "worker_id", value, limit, offset)
}

// FindByContractNumber mencari kecocokan contract_number dengan paginasi dan urutan primary key.
func (r *EmploymentContractRepository) FindByContractNumber(ctx context.Context, value string, limit, offset int) ([]entity.EmploymentContract, error) {
	return r.SelectBy(ctx, "contract_number", value, limit, offset)
}

// FindByContractType mencari kecocokan contract_type dengan paginasi dan urutan primary key.
func (r *EmploymentContractRepository) FindByContractType(ctx context.Context, value string, limit, offset int) ([]entity.EmploymentContract, error) {
	return r.SelectBy(ctx, "contract_type", value, limit, offset)
}

// FindByContractStatus mencari kecocokan contract_status dengan paginasi dan urutan primary key.
func (r *EmploymentContractRepository) FindByContractStatus(ctx context.Context, value string, limit, offset int) ([]entity.EmploymentContract, error) {
	return r.SelectBy(ctx, "contract_status", value, limit, offset)
}

// FindByEmployingLegalEntityReferenceID mencari kecocokan employing_legal_entity_reference_id dengan paginasi dan urutan primary key.
func (r *EmploymentContractRepository) FindByEmployingLegalEntityReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.EmploymentContract, error) {
	return r.SelectBy(ctx, "employing_legal_entity_reference_id", value, limit, offset)
}

// FindBySignedDocumentReferenceID mencari kecocokan signed_document_reference_id dengan paginasi dan urutan primary key.
func (r *EmploymentContractRepository) FindBySignedDocumentReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.EmploymentContract, error) {
	return r.SelectBy(ctx, "signed_document_reference_id", value, limit, offset)
}

// FindBySignedDocumentVersionReferenceID mencari kecocokan signed_document_version_reference_id dengan paginasi dan urutan primary key.
func (r *EmploymentContractRepository) FindBySignedDocumentVersionReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.EmploymentContract, error) {
	return r.SelectBy(ctx, "signed_document_version_reference_id", value, limit, offset)
}

// FindByWorkCountryCode mencari kecocokan work_country_code dengan paginasi dan urutan primary key.
func (r *EmploymentContractRepository) FindByWorkCountryCode(ctx context.Context, value string, limit, offset int) ([]entity.EmploymentContract, error) {
	return r.SelectBy(ctx, "work_country_code", value, limit, offset)
}

// FindByGoverningJurisdictionCode mencari kecocokan governing_jurisdiction_code dengan paginasi dan urutan primary key.
func (r *EmploymentContractRepository) FindByGoverningJurisdictionCode(ctx context.Context, value string, limit, offset int) ([]entity.EmploymentContract, error) {
	return r.SelectBy(ctx, "governing_jurisdiction_code", value, limit, offset)
}

// FindByCollectiveAgreementID mencari kecocokan collective_agreement_id dengan paginasi dan urutan primary key.
func (r *EmploymentContractRepository) FindByCollectiveAgreementID(ctx context.Context, value *string, limit, offset int) ([]entity.EmploymentContract, error) {
	return r.SelectBy(ctx, "collective_agreement_id", value, limit, offset)
}
