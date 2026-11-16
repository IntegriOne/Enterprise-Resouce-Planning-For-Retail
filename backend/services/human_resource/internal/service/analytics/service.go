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
	Runtime                      operation.Dependencies
	HumanCapitalMetricDefinition operation.Repository[entity.HumanCapitalMetricDefinition, string]
	HumanCapitalMetricSnapshot   operation.Repository[entity.HumanCapitalMetricSnapshot, string]
}

// Service mengoordinasikan operasi domain analytics.
type Service struct {
	humanCapitalMetricDefinition *operation.Resource[entity.HumanCapitalMetricDefinition, string, *entity.HumanCapitalMetricDefinition]
	humanCapitalMetricSnapshot   *operation.Resource[entity.HumanCapitalMetricSnapshot, string, *entity.HumanCapitalMetricSnapshot]
}

// NewService memeriksa seluruh port dan dependency sebelum service dapat dipakai.
func NewService(deps Dependencies) (*Service, error) {
	s := &Service{}
	var err error
	s.humanCapitalMetricDefinition, err = operation.NewResource[entity.HumanCapitalMetricDefinition, string, *entity.HumanCapitalMetricDefinition](deps.HumanCapitalMetricDefinition, deps.Runtime, specHumanCapitalMetricDefinition())
	if err != nil {
		return nil, err
	}
	s.humanCapitalMetricSnapshot, err = operation.NewResource[entity.HumanCapitalMetricSnapshot, string, *entity.HumanCapitalMetricSnapshot](deps.HumanCapitalMetricSnapshot, deps.Runtime, specHumanCapitalMetricSnapshot())
	if err != nil {
		return nil, err
	}
	return s, nil
}

// CreateHumanCapitalMetricDefinition menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreateHumanCapitalMetricDefinition(ctx context.Context, in CreateHumanCapitalMetricDefinitionInput) (*entity.HumanCapitalMetricDefinition, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.humanCapitalMetricDefinition.Create(ctx, func() (*entity.HumanCapitalMetricDefinition, error) { return createHumanCapitalMetricDefinition(in) })
}

// UpdateHumanCapitalMetricDefinition menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdateHumanCapitalMetricDefinition(ctx context.Context, key string, in UpdateHumanCapitalMetricDefinitionInput) (*entity.HumanCapitalMetricDefinition, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.humanCapitalMetricDefinition.Update(ctx, key, func(value *entity.HumanCapitalMetricDefinition) error {
		return patchHumanCapitalMetricDefinition(value, in)
	})
}

// DeleteHumanCapitalMetricDefinition menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeleteHumanCapitalMetricDefinition(ctx context.Context, key string) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.humanCapitalMetricDefinition.Delete(ctx, key)
}

// GetHumanCapitalMetricDefinitionByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetHumanCapitalMetricDefinitionByID(ctx context.Context, key string) (*entity.HumanCapitalMetricDefinition, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.humanCapitalMetricDefinition.Get(ctx, key)
}

// GetAllHumanCapitalMetricDefinitions menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllHumanCapitalMetricDefinitions(ctx context.Context, limit, offset int) (operation.Page[entity.HumanCapitalMetricDefinition], error) {
	if s == nil {
		return operation.Page[entity.HumanCapitalMetricDefinition]{}, hrerrors.ErrUnavailable
	}
	return s.humanCapitalMetricDefinition.List(ctx, nil, limit, offset)
}

// FindHumanCapitalMetricDefinitionByMetricCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindHumanCapitalMetricDefinitionByMetricCode(ctx context.Context, value string, limit, offset int) (operation.Page[entity.HumanCapitalMetricDefinition], error) {
	if s == nil {
		return operation.Page[entity.HumanCapitalMetricDefinition]{}, hrerrors.ErrUnavailable
	}
	return s.humanCapitalMetricDefinition.List(ctx, []domainrepo.Filter{{Column: "metric_code", Values: []any{value}}}, limit, offset)
}

// FindHumanCapitalMetricDefinitionByUnitCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindHumanCapitalMetricDefinitionByUnitCode(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.HumanCapitalMetricDefinition], error) {
	if s == nil {
		return operation.Page[entity.HumanCapitalMetricDefinition]{}, hrerrors.ErrUnavailable
	}
	return s.humanCapitalMetricDefinition.List(ctx, []domainrepo.Filter{{Column: "unit_code", Values: []any{value}}}, limit, offset)
}

// CreateHumanCapitalMetricSnapshot menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreateHumanCapitalMetricSnapshot(ctx context.Context, in CreateHumanCapitalMetricSnapshotInput) (*entity.HumanCapitalMetricSnapshot, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.humanCapitalMetricSnapshot.Create(ctx, func() (*entity.HumanCapitalMetricSnapshot, error) { return createHumanCapitalMetricSnapshot(in) })
}

// UpdateHumanCapitalMetricSnapshot menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdateHumanCapitalMetricSnapshot(ctx context.Context, key string, in UpdateHumanCapitalMetricSnapshotInput) (*entity.HumanCapitalMetricSnapshot, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.humanCapitalMetricSnapshot.Update(ctx, key, func(value *entity.HumanCapitalMetricSnapshot) error {
		return patchHumanCapitalMetricSnapshot(value, in)
	})
}

// DeleteHumanCapitalMetricSnapshot menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeleteHumanCapitalMetricSnapshot(ctx context.Context, key string) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.humanCapitalMetricSnapshot.Delete(ctx, key)
}

// GetHumanCapitalMetricSnapshotByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetHumanCapitalMetricSnapshotByID(ctx context.Context, key string) (*entity.HumanCapitalMetricSnapshot, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.humanCapitalMetricSnapshot.Get(ctx, key)
}

// GetAllHumanCapitalMetricSnapshots menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllHumanCapitalMetricSnapshots(ctx context.Context, limit, offset int) (operation.Page[entity.HumanCapitalMetricSnapshot], error) {
	if s == nil {
		return operation.Page[entity.HumanCapitalMetricSnapshot]{}, hrerrors.ErrUnavailable
	}
	return s.humanCapitalMetricSnapshot.List(ctx, nil, limit, offset)
}

// FindHumanCapitalMetricSnapshotByMetricDefinitionID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindHumanCapitalMetricSnapshotByMetricDefinitionID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.HumanCapitalMetricSnapshot], error) {
	if s == nil {
		return operation.Page[entity.HumanCapitalMetricSnapshot]{}, hrerrors.ErrUnavailable
	}
	return s.humanCapitalMetricSnapshot.List(ctx, []domainrepo.Filter{{Column: "metric_definition_id", Values: []any{value}}}, limit, offset)
}

// FindHumanCapitalMetricSnapshotByOrganizationUnitReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindHumanCapitalMetricSnapshotByOrganizationUnitReferenceID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.HumanCapitalMetricSnapshot], error) {
	if s == nil {
		return operation.Page[entity.HumanCapitalMetricSnapshot]{}, hrerrors.ErrUnavailable
	}
	return s.humanCapitalMetricSnapshot.List(ctx, []domainrepo.Filter{{Column: "organization_unit_reference_id", Values: []any{value}}}, limit, offset)
}

// FindHumanCapitalMetricSnapshotByJurisdictionCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindHumanCapitalMetricSnapshotByJurisdictionCode(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.HumanCapitalMetricSnapshot], error) {
	if s == nil {
		return operation.Page[entity.HumanCapitalMetricSnapshot]{}, hrerrors.ErrUnavailable
	}
	return s.humanCapitalMetricSnapshot.List(ctx, []domainrepo.Filter{{Column: "jurisdiction_code", Values: []any{value}}}, limit, offset)
}

// FindHumanCapitalMetricSnapshotByDataQualityStatus menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindHumanCapitalMetricSnapshotByDataQualityStatus(ctx context.Context, value string, limit, offset int) (operation.Page[entity.HumanCapitalMetricSnapshot], error) {
	if s == nil {
		return operation.Page[entity.HumanCapitalMetricSnapshot]{}, hrerrors.ErrUnavailable
	}
	return s.humanCapitalMetricSnapshot.List(ctx, []domainrepo.Filter{{Column: "data_quality_status", Values: []any{value}}}, limit, offset)
}
