package provisioning

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/iam/domain/repository"
	"github.com/integrione/erp-retail/backend/services/iam/repository/pgstore"
)

const employeeRequestsTable = "employee_provisioning_requests"

var _ domainrepo.EmployeeProvisioningRepository = (*EmployeeRepository)(nil)

type EmployeeRepository struct {
	db *gorm.DB
}

func NewEmployeeRepository(db *gorm.DB) (*EmployeeRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &EmployeeRepository{db: db}, nil
}

// InsertReceipt mengembalikan penerimaan yang sama untuk pengiriman ulang dengan isi identik.
func (r *EmployeeRepository) InsertReceipt(ctx context.Context, request entity.EmployeeProvisioningRequest) (*entity.EmployeeProvisioningRequest, error) {
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return nil, err
	}
	values := map[string]any{
		"hrm_request_id": request.HRMRequestID, "worker_id": request.WorkerID, "worker_number": request.WorkerNumber,
		"legal_entity_id": request.LegalEntityID, "requested_by_principal_id": request.RequestedByPrincipalID,
		"requested_at": request.RequestedAt, "catalog_role_id": request.CatalogRoleID,
	}
	err = db.Table(employeeRequestsTable).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "hrm_request_id"}}, DoNothing: true,
	}).Create(values).Error
	if err != nil {
		return nil, pgstore.MapError(err, nil)
	}
	var receipt entity.EmployeeProvisioningRequest
	if err := db.Table(employeeRequestsTable).Where("hrm_request_id = ?", request.HRMRequestID).Take(&receipt).Error; err != nil {
		return nil, pgstore.MapError(err, nil)
	}
	if receipt.WorkerID != request.WorkerID || receipt.WorkerNumber != request.WorkerNumber || receipt.LegalEntityID != request.LegalEntityID ||
		receipt.CatalogRoleID != request.CatalogRoleID || receipt.RequestedByPrincipalID != request.RequestedByPrincipalID || !receipt.RequestedAt.Equal(request.RequestedAt.Truncate(time.Microsecond)) {
		return nil, iamerrors.ErrConflict
	}
	return &receipt, nil
}

// SelectOpen membaca pengajuan yang belum selesai dalam urutan penerimaan yang stabil.
func (r *EmployeeRepository) SelectOpen(ctx context.Context, limit, offset int) ([]entity.EmployeeProvisioningRequest, int64, error) {
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return nil, 0, err
	}
	query := db.Table(employeeRequestsTable).
		Where("status IN ?", []string{entity.EmployeeProvisioningPending, entity.EmployeeProvisioningProcessing})
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, pgstore.MapError(err, nil)
	}
	var items []entity.EmployeeProvisioningRequest
	if err := query.Order("received_at ASC, id ASC").Limit(limit).Offset(offset).Find(&items).Error; err != nil {
		return nil, 0, pgstore.MapError(err, nil)
	}
	return items, total, nil
}

// SelectByHRMRequest membaca penerimaan pengajuan pada badan hukum yang diotorisasi.
func (r *EmployeeRepository) SelectByHRMRequest(ctx context.Context, hrmRequestID, legalEntityID string) (*entity.EmployeeProvisioningRequest, error) {
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return nil, err
	}
	var request entity.EmployeeProvisioningRequest
	err = db.Table(employeeRequestsTable).Where("hrm_request_id = ? AND legal_entity_id = ?", hrmRequestID, legalEntityID).Take(&request).Error
	if err != nil {
		return nil, pgstore.MapError(err, nil)
	}
	return &request, nil
}

func (r *EmployeeRepository) SelectByID(ctx context.Context, id string) (*entity.EmployeeProvisioningRequest, error) {
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return nil, err
	}
	var request entity.EmployeeProvisioningRequest
	if err := db.Table(employeeRequestsTable).Where("id = ?", id).Take(&request).Error; err != nil {
		return nil, pgstore.MapError(err, nil)
	}
	return &request, nil
}

// UpdateClaim hanya mengizinkan satu operator memulai pemrosesan pengajuan pending.
func (r *EmployeeRepository) UpdateClaim(ctx context.Context, id, operatorID string, at time.Time) (*entity.EmployeeProvisioningRequest, error) {
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return nil, err
	}
	var request entity.EmployeeProvisioningRequest
	result := db.Table(employeeRequestsTable).Model(&request).Clauses(clause.Returning{}).
		Where("id = ? AND status = ?", id, entity.EmployeeProvisioningPending).
		Updates(map[string]any{
			"status": entity.EmployeeProvisioningProcessing, "processed_by_user_id": operatorID, "claimed_at": at,
		})
	if result.Error != nil {
		return nil, pgstore.MapError(result.Error, nil)
	}
	if result.RowsAffected != 1 {
		return nil, iamerrors.ErrConflict
	}
	return &request, nil
}

// SelectByIDForUpdate mengunci pengajuan selama akun dan asosiasinya disimpan.
func (r *EmployeeRepository) SelectByIDForUpdate(ctx context.Context, id string) (*entity.EmployeeProvisioningRequest, error) {
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return nil, err
	}
	var request entity.EmployeeProvisioningRequest
	err = db.Table(employeeRequestsTable).Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", id).Take(&request).Error
	if err != nil {
		return nil, pgstore.MapError(err, nil)
	}
	return &request, nil
}

// HasUnlinkedAccount memeriksa penerimaan provisioning yang belum menyelesaikan penautan.
func (r *EmployeeRepository) HasUnlinkedAccount(ctx context.Context, accountID string) (bool, error) {
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return false, err
	}
	var count int64
	err = db.Table(employeeRequestsTable).
		Where("account_user_id = ? AND status <> ?", accountID, entity.EmployeeProvisioningLinked).
		Count(&count).Error
	return count > 0, pgstore.MapError(err, nil)
}

func (r *EmployeeRepository) UpdateAccount(ctx context.Context, id, operatorID, accountID string) error {
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return err
	}
	result := db.Table(employeeRequestsTable).
		Where("id = ? AND status = ? AND processed_by_user_id = ? AND account_user_id IS NULL",
			id, entity.EmployeeProvisioningProcessing, operatorID).
		Update("account_user_id", accountID)
	if result.Error != nil {
		return pgstore.MapError(result.Error, nil)
	}
	if result.RowsAffected != 1 {
		return iamerrors.ErrConflict
	}
	return nil
}

// UpdateLinked menyelesaikan pengajuan setelah Gateway mengonfirmasi tautan ke pekerja.
func (r *EmployeeRepository) UpdateLinked(ctx context.Context, id, operatorID, accountID string, at time.Time) error {
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return err
	}
	result := db.Table(employeeRequestsTable).
		Where("id = ? AND processed_by_user_id = ? AND account_user_id = ? AND status IN ?", id, operatorID, accountID,
			[]string{entity.EmployeeProvisioningProcessing, entity.EmployeeProvisioningLinked}).
		Updates(map[string]any{"status": entity.EmployeeProvisioningLinked, "resolved_at": gorm.Expr("COALESCE(resolved_at, ?)", at)})
	if result.Error != nil {
		return pgstore.MapError(result.Error, nil)
	}
	if result.RowsAffected != 1 {
		return iamerrors.ErrConflict
	}
	return nil
}
