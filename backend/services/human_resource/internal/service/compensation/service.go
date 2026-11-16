// Kerangka awal dibuat dengan alat bantu; kode ini dipelihara langsung per domain.
// Aturan bisnis lintas record tetap dijalankan melalui policy/integrity dan command domain.
package service

import (
	"context"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

type Dependencies struct {
	Runtime            operation.Dependencies
	CompensationBand   domainrepo.RecordRepository[entity.CompensationBand, string]
	CompensationRecord domainrepo.RecordRepository[entity.CompensationRecord, string]
}

// Service mengoordinasikan operasi domain compensation.
type Service struct {
	compensationBand   *operation.Resource[entity.CompensationBand, string, *entity.CompensationBand]
	compensationRecord *operation.Resource[entity.CompensationRecord, string, *entity.CompensationRecord]
}

// NewService memeriksa seluruh port dan dependency sebelum service dapat dipakai.
func NewService(deps Dependencies) (*Service, error) {
	s := &Service{}
	var err error
	s.compensationBand, err = operation.NewResource[entity.CompensationBand, string, *entity.CompensationBand](
		deps.CompensationBand,
		deps.Runtime,
		specCompensationBand(),
	)
	if err != nil {
		return nil, err
	}
	s.compensationRecord, err = operation.NewResource[entity.CompensationRecord, string, *entity.CompensationRecord](
		deps.CompensationRecord,
		deps.Runtime,
		specCompensationRecord(),
	)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// CreateCompensationBand menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreateCompensationBand(ctx context.Context, in CreateCompensationBandInput) (*entity.CompensationBand, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.compensationBand.Create(ctx, func() (*entity.CompensationBand, error) { return createCompensationBand(in) })
}

// UpdateCompensationBand menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdateCompensationBand(ctx context.Context, key string, in UpdateCompensationBandInput) (*entity.CompensationBand, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.compensationBand.Update(ctx, key, func(value *entity.CompensationBand) error { return patchCompensationBand(value, in) })
}

// DeleteCompensationBand menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeleteCompensationBand(ctx context.Context, key string) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.compensationBand.Delete(ctx, key)
}

// GetCompensationBandByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetCompensationBandByID(ctx context.Context, key string) (*entity.CompensationBand, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.compensationBand.Get(ctx, key)
}

// GetAllCompensationBands menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllCompensationBands(ctx context.Context, limit, offset int) (operation.Page[entity.CompensationBand], error) {
	if s == nil {
		return operation.Page[entity.CompensationBand]{}, hrerrors.ErrUnavailable
	}
	return s.compensationBand.List(ctx, nil, limit, offset)
}

// FindCompensationBandByBandCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindCompensationBandByBandCode(ctx context.Context, value string, limit, offset int) (operation.Page[entity.CompensationBand], error) {
	if s == nil {
		return operation.Page[entity.CompensationBand]{}, hrerrors.ErrUnavailable
	}
	return s.compensationBand.List(ctx, []domainrepo.Filter{{Column: "band_code", Values: []any{value}}}, limit, offset)
}

// FindCompensationBandByJobLevelCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindCompensationBandByJobLevelCode(ctx context.Context, value string, limit, offset int) (operation.Page[entity.CompensationBand], error) {
	if s == nil {
		return operation.Page[entity.CompensationBand]{}, hrerrors.ErrUnavailable
	}
	return s.compensationBand.List(ctx, []domainrepo.Filter{{Column: "job_level_code", Values: []any{value}}}, limit, offset)
}

// FindCompensationBandByJurisdictionCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindCompensationBandByJurisdictionCode(ctx context.Context, value string, limit, offset int) (operation.Page[entity.CompensationBand], error) {
	if s == nil {
		return operation.Page[entity.CompensationBand]{}, hrerrors.ErrUnavailable
	}
	return s.compensationBand.List(ctx, []domainrepo.Filter{{Column: "jurisdiction_code", Values: []any{value}}}, limit, offset)
}

// FindCompensationBandByCurrencyCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindCompensationBandByCurrencyCode(ctx context.Context, value string, limit, offset int) (operation.Page[entity.CompensationBand], error) {
	if s == nil {
		return operation.Page[entity.CompensationBand]{}, hrerrors.ErrUnavailable
	}
	return s.compensationBand.List(ctx, []domainrepo.Filter{{Column: "currency_code", Values: []any{value}}}, limit, offset)
}

// FindCompensationBandByApprovalReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindCompensationBandByApprovalReferenceID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.CompensationBand], error) {
	if s == nil {
		return operation.Page[entity.CompensationBand]{}, hrerrors.ErrUnavailable
	}
	return s.compensationBand.List(ctx, []domainrepo.Filter{{Column: "approval_reference_id", Values: []any{value}}}, limit, offset)
}

// CreateCompensationRecord menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreateCompensationRecord(ctx context.Context, in CreateCompensationRecordInput) (*entity.CompensationRecord, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.compensationRecord.Create(ctx, func() (*entity.CompensationRecord, error) { return createCompensationRecord(in) })
}

// UpdateCompensationRecord menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdateCompensationRecord(ctx context.Context, key string, in UpdateCompensationRecordInput) (*entity.CompensationRecord, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.compensationRecord.Update(ctx, key, func(value *entity.CompensationRecord) error { return patchCompensationRecord(value, in) })
}

// DeleteCompensationRecord menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeleteCompensationRecord(ctx context.Context, key string) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.compensationRecord.Delete(ctx, key)
}

// GetCompensationRecordByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetCompensationRecordByID(ctx context.Context, key string) (*entity.CompensationRecord, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.compensationRecord.Get(ctx, key)
}

// GetAllCompensationRecords menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllCompensationRecords(ctx context.Context, limit, offset int) (operation.Page[entity.CompensationRecord], error) {
	if s == nil {
		return operation.Page[entity.CompensationRecord]{}, hrerrors.ErrUnavailable
	}
	return s.compensationRecord.List(ctx, nil, limit, offset)
}

// FindCompensationRecordByWorkerID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindCompensationRecordByWorkerID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.CompensationRecord], error) {
	if s == nil {
		return operation.Page[entity.CompensationRecord]{}, hrerrors.ErrUnavailable
	}
	return s.compensationRecord.List(ctx, []domainrepo.Filter{{Column: "worker_id", Values: []any{value}}}, limit, offset)
}

// FindCompensationRecordByAssignmentID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindCompensationRecordByAssignmentID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.CompensationRecord], error) {
	if s == nil {
		return operation.Page[entity.CompensationRecord]{}, hrerrors.ErrUnavailable
	}
	return s.compensationRecord.List(ctx, []domainrepo.Filter{{Column: "assignment_id", Values: []any{value}}}, limit, offset)
}

// FindCompensationRecordByComponentID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindCompensationRecordByComponentID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.CompensationRecord], error) {
	if s == nil {
		return operation.Page[entity.CompensationRecord]{}, hrerrors.ErrUnavailable
	}
	return s.compensationRecord.List(ctx, []domainrepo.Filter{{Column: "component_id", Values: []any{value}}}, limit, offset)
}

// FindCompensationRecordByCompensationBandID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindCompensationRecordByCompensationBandID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.CompensationRecord], error) {
	if s == nil {
		return operation.Page[entity.CompensationRecord]{}, hrerrors.ErrUnavailable
	}
	return s.compensationRecord.List(ctx, []domainrepo.Filter{{Column: "compensation_band_id", Values: []any{value}}}, limit, offset)
}

// FindCompensationRecordByCompensationType menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindCompensationRecordByCompensationType(ctx context.Context, value string, limit, offset int) (operation.Page[entity.CompensationRecord], error) {
	if s == nil {
		return operation.Page[entity.CompensationRecord]{}, hrerrors.ErrUnavailable
	}
	return s.compensationRecord.List(ctx, []domainrepo.Filter{{Column: "compensation_type", Values: []any{value}}}, limit, offset)
}

// FindCompensationRecordByCurrencyCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindCompensationRecordByCurrencyCode(ctx context.Context, value string, limit, offset int) (operation.Page[entity.CompensationRecord], error) {
	if s == nil {
		return operation.Page[entity.CompensationRecord]{}, hrerrors.ErrUnavailable
	}
	return s.compensationRecord.List(ctx, []domainrepo.Filter{{Column: "currency_code", Values: []any{value}}}, limit, offset)
}

// FindCompensationRecordByStatus menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindCompensationRecordByStatus(ctx context.Context, value entity.CompensationStatus, limit, offset int) (operation.Page[entity.CompensationRecord], error) {
	if s == nil {
		return operation.Page[entity.CompensationRecord]{}, hrerrors.ErrUnavailable
	}
	return s.compensationRecord.List(ctx, []domainrepo.Filter{{Column: "status", Values: []any{value}}}, limit, offset)
}

// FindCompensationRecordByReasonCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindCompensationRecordByReasonCode(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.CompensationRecord], error) {
	if s == nil {
		return operation.Page[entity.CompensationRecord]{}, hrerrors.ErrUnavailable
	}
	return s.compensationRecord.List(ctx, []domainrepo.Filter{{Column: "reason_code", Values: []any{value}}}, limit, offset)
}

// FindCompensationRecordByApprovedByPrincipalID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindCompensationRecordByApprovedByPrincipalID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.CompensationRecord], error) {
	if s == nil {
		return operation.Page[entity.CompensationRecord]{}, hrerrors.ErrUnavailable
	}
	return s.compensationRecord.List(ctx, []domainrepo.Filter{{Column: "approved_by_principal_id", Values: []any{value}}}, limit, offset)
}

// FindCompensationRecordBySourceChangeReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindCompensationRecordBySourceChangeReferenceID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.CompensationRecord], error) {
	if s == nil {
		return operation.Page[entity.CompensationRecord]{}, hrerrors.ErrUnavailable
	}
	return s.compensationRecord.List(ctx, []domainrepo.Filter{{Column: "source_change_reference_id", Values: []any{value}}}, limit, offset)
}
