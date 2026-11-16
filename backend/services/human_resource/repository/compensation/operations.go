package compensation

import (
	"context"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
)

// Operasi berikut menyebut objek dan kolom secara eksplisit untuk memudahkan penelusuran.
func (r *CompensationBandRepository) InsertCompensationBands(ctx context.Context, value *entity.CompensationBand) error {
	return r.Insert(ctx, value)
}
func (r *CompensationBandRepository) UpdateCompensationBands(ctx context.Context, value *entity.CompensationBand) error {
	return r.Update(ctx, value)
}
func (r *CompensationBandRepository) DeleteCompensationBandsByUUID(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
func (r *CompensationBandRepository) FindCompensationBandsByUUID(ctx context.Context, id string) (*entity.CompensationBand, error) {
	return r.FindByID(ctx, id)
}
func (r *CompensationBandRepository) FindCompensationBandsByUUIDForUpdate(ctx context.Context, id string) (*entity.CompensationBand, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *CompensationBandRepository) GetAllCompensationBands(ctx context.Context, limit, offset int) ([]entity.CompensationBand, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *CompensationBandRepository) CountCompensationBands(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *CompensationBandRepository) SelectCompensationBandsWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.CompensationBand, error) {
	return r.SelectWhere(ctx, filters, limit, offset)
}
func (r *CompensationBandRepository) CountCompensationBandsWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, filters)
}
func (r *CompensationBandRepository) FindCompensationBandsByBandCode(ctx context.Context, value string, limit, offset int) ([]entity.CompensationBand, error) {
	return r.FindByBandCode(ctx, value, limit, offset)
}
func (r *CompensationBandRepository) FindCompensationBandsByJobLevelCode(ctx context.Context, value string, limit, offset int) ([]entity.CompensationBand, error) {
	return r.FindByJobLevelCode(ctx, value, limit, offset)
}
func (r *CompensationBandRepository) FindCompensationBandsByJurisdictionCode(ctx context.Context, value string, limit, offset int) ([]entity.CompensationBand, error) {
	return r.FindByJurisdictionCode(ctx, value, limit, offset)
}
func (r *CompensationBandRepository) FindCompensationBandsByCurrencyCode(ctx context.Context, value string, limit, offset int) ([]entity.CompensationBand, error) {
	return r.FindByCurrencyCode(ctx, value, limit, offset)
}
func (r *CompensationBandRepository) FindCompensationBandsByApprovalReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.CompensationBand, error) {
	return r.FindByApprovalReferenceID(ctx, value, limit, offset)
}

func (r *CompensationRecordRepository) InsertCompensationRecords(ctx context.Context, value *entity.CompensationRecord) error {
	return r.Insert(ctx, value)
}
func (r *CompensationRecordRepository) UpdateCompensationRecords(ctx context.Context, value *entity.CompensationRecord) error {
	return r.Update(ctx, value)
}
func (r *CompensationRecordRepository) DeleteCompensationRecordsByUUID(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
func (r *CompensationRecordRepository) FindCompensationRecordsByUUID(ctx context.Context, id string) (*entity.CompensationRecord, error) {
	return r.FindByID(ctx, id)
}
func (r *CompensationRecordRepository) FindCompensationRecordsByUUIDForUpdate(ctx context.Context, id string) (*entity.CompensationRecord, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *CompensationRecordRepository) GetAllCompensationRecords(ctx context.Context, limit, offset int) ([]entity.CompensationRecord, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *CompensationRecordRepository) CountCompensationRecords(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *CompensationRecordRepository) SelectCompensationRecordsWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.CompensationRecord, error) {
	return r.SelectWhere(ctx, filters, limit, offset)
}
func (r *CompensationRecordRepository) CountCompensationRecordsWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, filters)
}
func (r *CompensationRecordRepository) FindCompensationRecordsByWorkerID(ctx context.Context, value string, limit, offset int) ([]entity.CompensationRecord, error) {
	return r.FindByWorkerID(ctx, value, limit, offset)
}
func (r *CompensationRecordRepository) FindCompensationRecordsByAssignmentID(ctx context.Context, value string, limit, offset int) ([]entity.CompensationRecord, error) {
	return r.FindByAssignmentID(ctx, value, limit, offset)
}
func (r *CompensationRecordRepository) FindCompensationRecordsByComponentID(ctx context.Context, value string, limit, offset int) ([]entity.CompensationRecord, error) {
	return r.FindByComponentID(ctx, value, limit, offset)
}
func (r *CompensationRecordRepository) FindCompensationRecordsByCompensationBandID(ctx context.Context, value *string, limit, offset int) ([]entity.CompensationRecord, error) {
	return r.FindByCompensationBandID(ctx, value, limit, offset)
}
func (r *CompensationRecordRepository) FindCompensationRecordsByCompensationType(ctx context.Context, value string, limit, offset int) ([]entity.CompensationRecord, error) {
	return r.FindByCompensationType(ctx, value, limit, offset)
}
func (r *CompensationRecordRepository) FindCompensationRecordsByCurrencyCode(ctx context.Context, value string, limit, offset int) ([]entity.CompensationRecord, error) {
	return r.FindByCurrencyCode(ctx, value, limit, offset)
}
func (r *CompensationRecordRepository) FindCompensationRecordsByStatus(ctx context.Context, value entity.CompensationStatus, limit, offset int) ([]entity.CompensationRecord, error) {
	return r.FindByStatus(ctx, value, limit, offset)
}
func (r *CompensationRecordRepository) FindCompensationRecordsByReasonCode(ctx context.Context, value *string, limit, offset int) ([]entity.CompensationRecord, error) {
	return r.FindByReasonCode(ctx, value, limit, offset)
}
func (r *CompensationRecordRepository) FindCompensationRecordsByApprovedByPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.CompensationRecord, error) {
	return r.FindByApprovedByPrincipalID(ctx, value, limit, offset)
}
func (r *CompensationRecordRepository) FindCompensationRecordsBySourceChangeReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.CompensationRecord, error) {
	return r.FindBySourceChangeReferenceID(ctx, value, limit, offset)
}
