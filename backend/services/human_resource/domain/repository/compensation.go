// Sumber model adalah baseline SQL HRM; aturan bisnis tetap milik service.
package repository

import (
	"context"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
)

// CompensationBandRepository adalah port persistensi hr_compensation_bands.
type CompensationBandRepository interface {
	Repository[entity.CompensationBand, string]
	InsertCompensationBands(context.Context, *entity.CompensationBand) error
	UpdateCompensationBands(context.Context, *entity.CompensationBand) error
	DeleteCompensationBandsByUUID(context.Context, string) error
	FindCompensationBandsByUUID(context.Context, string) (*entity.CompensationBand, error)
	FindCompensationBandsByUUIDForUpdate(context.Context, string) (*entity.CompensationBand, error)
	GetAllCompensationBands(context.Context, int, int) ([]entity.CompensationBand, error)
	CountCompensationBands(context.Context) (int64, error)
	SelectCompensationBandsWhere(context.Context, []Filter, int, int) ([]entity.CompensationBand, error)
	CountCompensationBandsWhere(context.Context, []Filter) (int64, error)
	FindCompensationBandsByBandCode(context.Context, string, int, int) ([]entity.CompensationBand, error)
	FindCompensationBandsByJobLevelCode(context.Context, string, int, int) ([]entity.CompensationBand, error)
	FindCompensationBandsByJurisdictionCode(context.Context, string, int, int) ([]entity.CompensationBand, error)
	FindCompensationBandsByCurrencyCode(context.Context, string, int, int) ([]entity.CompensationBand, error)
	FindCompensationBandsByApprovalReferenceID(context.Context, *string, int, int) ([]entity.CompensationBand, error)
	FindByBandCode(ctx context.Context, value string, limit, offset int) ([]entity.CompensationBand, error)
	FindByJobLevelCode(ctx context.Context, value string, limit, offset int) ([]entity.CompensationBand, error)
	FindByJurisdictionCode(ctx context.Context, value string, limit, offset int) ([]entity.CompensationBand, error)
	FindByCurrencyCode(ctx context.Context, value string, limit, offset int) ([]entity.CompensationBand, error)
	FindByApprovalReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.CompensationBand, error)
}

// CompensationRecordRepository adalah port persistensi hr_compensation_records.
type CompensationRecordRepository interface {
	Repository[entity.CompensationRecord, string]
	InsertCompensationRecords(context.Context, *entity.CompensationRecord) error
	UpdateCompensationRecords(context.Context, *entity.CompensationRecord) error
	DeleteCompensationRecordsByUUID(context.Context, string) error
	FindCompensationRecordsByUUID(context.Context, string) (*entity.CompensationRecord, error)
	FindCompensationRecordsByUUIDForUpdate(context.Context, string) (*entity.CompensationRecord, error)
	GetAllCompensationRecords(context.Context, int, int) ([]entity.CompensationRecord, error)
	CountCompensationRecords(context.Context) (int64, error)
	SelectCompensationRecordsWhere(context.Context, []Filter, int, int) ([]entity.CompensationRecord, error)
	CountCompensationRecordsWhere(context.Context, []Filter) (int64, error)
	FindCompensationRecordsByWorkerID(context.Context, string, int, int) ([]entity.CompensationRecord, error)
	FindCompensationRecordsByAssignmentID(context.Context, string, int, int) ([]entity.CompensationRecord, error)
	FindCompensationRecordsByComponentID(context.Context, string, int, int) ([]entity.CompensationRecord, error)
	FindCompensationRecordsByCompensationBandID(context.Context, *string, int, int) ([]entity.CompensationRecord, error)
	FindCompensationRecordsByCompensationType(context.Context, string, int, int) ([]entity.CompensationRecord, error)
	FindCompensationRecordsByCurrencyCode(context.Context, string, int, int) ([]entity.CompensationRecord, error)
	FindCompensationRecordsByStatus(context.Context, entity.CompensationStatus, int, int) ([]entity.CompensationRecord, error)
	FindCompensationRecordsByReasonCode(context.Context, *string, int, int) ([]entity.CompensationRecord, error)
	FindCompensationRecordsByApprovedByPrincipalID(context.Context, *string, int, int) ([]entity.CompensationRecord, error)
	FindCompensationRecordsBySourceChangeReferenceID(context.Context, *string, int, int) ([]entity.CompensationRecord, error)
	FindByWorkerID(ctx context.Context, value string, limit, offset int) ([]entity.CompensationRecord, error)
	FindByAssignmentID(ctx context.Context, value string, limit, offset int) ([]entity.CompensationRecord, error)
	FindByComponentID(ctx context.Context, value string, limit, offset int) ([]entity.CompensationRecord, error)
	FindByCompensationBandID(ctx context.Context, value *string, limit, offset int) ([]entity.CompensationRecord, error)
	FindByCompensationType(ctx context.Context, value string, limit, offset int) ([]entity.CompensationRecord, error)
	FindByCurrencyCode(ctx context.Context, value string, limit, offset int) ([]entity.CompensationRecord, error)
	FindByStatus(ctx context.Context, value entity.CompensationStatus, limit, offset int) ([]entity.CompensationRecord, error)
	FindByReasonCode(ctx context.Context, value *string, limit, offset int) ([]entity.CompensationRecord, error)
	FindByApprovedByPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.CompensationRecord, error)
	FindBySourceChangeReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.CompensationRecord, error)
}
