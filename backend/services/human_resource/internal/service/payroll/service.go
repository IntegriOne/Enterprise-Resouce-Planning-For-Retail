// Kerangka awal dibuat oleh alat bantu; selanjutnya dipelihara langsung per domain.
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
	Runtime           operation.Dependencies
	PayrollProfile    domainrepo.RecordRepository[entity.PayrollProfile, string]
	PayComponent      domainrepo.RecordRepository[entity.PayComponent, string]
	PayrollPeriod     domainrepo.RecordRepository[entity.PayrollPeriod, string]
	PayrollRun        domainrepo.RecordRepository[entity.PayrollRun, string]
	PayrollResult     domainrepo.RecordRepository[entity.PayrollResult, string]
	PayrollResultLine domainrepo.RecordRepository[entity.PayrollResultLine, string]
	PayrollAdjustment domainrepo.RecordRepository[entity.PayrollAdjustment, string]
	PayrollGroup      domainrepo.RecordRepository[entity.PayrollGroup, string]
	PayrollRuleSet    domainrepo.RecordRepository[entity.PayrollRuleSet, string]
}

// Service mengoordinasikan operasi domain payroll.
type Service struct {
	dependencies      Dependencies
	payrollProfile    *operation.Resource[entity.PayrollProfile, string, *entity.PayrollProfile]
	payComponent      *operation.Resource[entity.PayComponent, string, *entity.PayComponent]
	payrollPeriod     *operation.Resource[entity.PayrollPeriod, string, *entity.PayrollPeriod]
	payrollRun        *operation.Resource[entity.PayrollRun, string, *entity.PayrollRun]
	payrollResult     *operation.Resource[entity.PayrollResult, string, *entity.PayrollResult]
	payrollResultLine *operation.Resource[entity.PayrollResultLine, string, *entity.PayrollResultLine]
	payrollAdjustment *operation.Resource[entity.PayrollAdjustment, string, *entity.PayrollAdjustment]
	payrollGroup      *operation.Resource[entity.PayrollGroup, string, *entity.PayrollGroup]
	payrollRuleSet    *operation.Resource[entity.PayrollRuleSet, string, *entity.PayrollRuleSet]
}

// NewService memeriksa seluruh port dan dependency sebelum service dapat dipakai.
func NewService(deps Dependencies) (*Service, error) {
	s := &Service{dependencies: deps}
	var err error
	s.payrollProfile, err = operation.NewResource[entity.PayrollProfile, string, *entity.PayrollProfile](
		deps.PayrollProfile,
		deps.Runtime,
		specPayrollProfile(),
	)
	if err != nil {
		return nil, err
	}
	s.payComponent, err = operation.NewResource[entity.PayComponent, string, *entity.PayComponent](
		deps.PayComponent,
		deps.Runtime,
		specPayComponent(),
	)
	if err != nil {
		return nil, err
	}
	s.payrollPeriod, err = operation.NewResource[entity.PayrollPeriod, string, *entity.PayrollPeriod](
		deps.PayrollPeriod,
		deps.Runtime,
		specPayrollPeriod(),
	)
	if err != nil {
		return nil, err
	}
	s.payrollRun, err = operation.NewResource[entity.PayrollRun, string, *entity.PayrollRun](
		deps.PayrollRun,
		deps.Runtime,
		specPayrollRun(),
	)
	if err != nil {
		return nil, err
	}
	s.payrollResult, err = operation.NewResource[entity.PayrollResult, string, *entity.PayrollResult](
		deps.PayrollResult,
		deps.Runtime,
		specPayrollResult(),
	)
	if err != nil {
		return nil, err
	}
	s.payrollResultLine, err = operation.NewResource[entity.PayrollResultLine, string, *entity.PayrollResultLine](
		deps.PayrollResultLine,
		deps.Runtime,
		specPayrollResultLine(),
	)
	if err != nil {
		return nil, err
	}
	s.payrollAdjustment, err = operation.NewResource[entity.PayrollAdjustment, string, *entity.PayrollAdjustment](
		deps.PayrollAdjustment,
		deps.Runtime,
		specPayrollAdjustment(),
	)
	if err != nil {
		return nil, err
	}
	s.payrollGroup, err = operation.NewResource[entity.PayrollGroup, string, *entity.PayrollGroup](
		deps.PayrollGroup,
		deps.Runtime,
		specPayrollGroup(),
	)
	if err != nil {
		return nil, err
	}
	s.payrollRuleSet, err = operation.NewResource[entity.PayrollRuleSet, string, *entity.PayrollRuleSet](
		deps.PayrollRuleSet,
		deps.Runtime,
		specPayrollRuleSet(),
	)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// CreatePayrollProfile menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreatePayrollProfile(ctx context.Context, in CreatePayrollProfileInput) (*entity.PayrollProfile, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.payrollProfile.Create(ctx, func() (*entity.PayrollProfile, error) { return createPayrollProfile(in) })
}

// UpdatePayrollProfile menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdatePayrollProfile(ctx context.Context, key string, in UpdatePayrollProfileInput) (*entity.PayrollProfile, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.payrollProfile.Update(ctx, key, func(value *entity.PayrollProfile) error { return patchPayrollProfile(value, in) })
}

// DeletePayrollProfile menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeletePayrollProfile(ctx context.Context, key string) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.payrollProfile.Delete(ctx, key)
}

// GetPayrollProfileByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetPayrollProfileByID(ctx context.Context, key string) (*entity.PayrollProfile, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.payrollProfile.Get(ctx, key)
}

// GetAllPayrollProfiles menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllPayrollProfiles(ctx context.Context, limit, offset int) (operation.Page[entity.PayrollProfile], error) {
	if s == nil {
		return operation.Page[entity.PayrollProfile]{}, hrerrors.ErrUnavailable
	}
	return s.payrollProfile.List(ctx, nil, limit, offset)
}

// FindPayrollProfileByWorkerID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPayrollProfileByWorkerID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.PayrollProfile], error) {
	if s == nil {
		return operation.Page[entity.PayrollProfile]{}, hrerrors.ErrUnavailable
	}
	return s.payrollProfile.List(ctx, []domainrepo.Filter{{Column: "worker_id", Values: []any{value}}}, limit, offset)
}

// FindPayrollProfileByPayrollCountryCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPayrollProfileByPayrollCountryCode(ctx context.Context, value string, limit, offset int) (operation.Page[entity.PayrollProfile], error) {
	if s == nil {
		return operation.Page[entity.PayrollProfile]{}, hrerrors.ErrUnavailable
	}
	return s.payrollProfile.List(ctx, []domainrepo.Filter{{Column: "payroll_country_code", Values: []any{value}}}, limit, offset)
}

// FindPayrollProfileByPayrollJurisdictionCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPayrollProfileByPayrollJurisdictionCode(ctx context.Context, value string, limit, offset int) (operation.Page[entity.PayrollProfile], error) {
	if s == nil {
		return operation.Page[entity.PayrollProfile]{}, hrerrors.ErrUnavailable
	}
	return s.payrollProfile.List(ctx, []domainrepo.Filter{{Column: "payroll_jurisdiction_code", Values: []any{value}}}, limit, offset)
}

// FindPayrollProfileByPayrollGroupID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPayrollProfileByPayrollGroupID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.PayrollProfile], error) {
	if s == nil {
		return operation.Page[entity.PayrollProfile]{}, hrerrors.ErrUnavailable
	}
	return s.payrollProfile.List(ctx, []domainrepo.Filter{{Column: "payroll_group_id", Values: []any{value}}}, limit, offset)
}

// FindPayrollProfileByEmployingLegalEntityReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPayrollProfileByEmployingLegalEntityReferenceID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.PayrollProfile], error) {
	if s == nil {
		return operation.Page[entity.PayrollProfile]{}, hrerrors.ErrUnavailable
	}
	return s.payrollProfile.List(ctx, []domainrepo.Filter{{Column: "employing_legal_entity_reference_id", Values: []any{value}}}, limit, offset)
}

// FindPayrollProfileByCurrencyCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPayrollProfileByCurrencyCode(ctx context.Context, value string, limit, offset int) (operation.Page[entity.PayrollProfile], error) {
	if s == nil {
		return operation.Page[entity.PayrollProfile]{}, hrerrors.ErrUnavailable
	}
	return s.payrollProfile.List(ctx, []domainrepo.Filter{{Column: "currency_code", Values: []any{value}}}, limit, offset)
}

// FindPayrollProfileByTaxProfileReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPayrollProfileByTaxProfileReferenceID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.PayrollProfile], error) {
	if s == nil {
		return operation.Page[entity.PayrollProfile]{}, hrerrors.ErrUnavailable
	}
	return s.payrollProfile.List(ctx, []domainrepo.Filter{{Column: "tax_profile_reference_id", Values: []any{value}}}, limit, offset)
}

// FindPayrollProfileByBankAccountTokenReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPayrollProfileByBankAccountTokenReferenceID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.PayrollProfile], error) {
	if s == nil {
		return operation.Page[entity.PayrollProfile]{}, hrerrors.ErrUnavailable
	}
	return s.payrollProfile.List(ctx, []domainrepo.Filter{{Column: "bank_account_token_reference_id", Values: []any{value}}}, limit, offset)
}

// FindPayrollProfileByTaxStatusCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPayrollProfileByTaxStatusCode(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.PayrollProfile], error) {
	if s == nil {
		return operation.Page[entity.PayrollProfile]{}, hrerrors.ErrUnavailable
	}
	return s.payrollProfile.List(ctx, []domainrepo.Filter{{Column: "tax_status_code", Values: []any{value}}}, limit, offset)
}

// FindPayrollProfileByWithholdingElectionReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPayrollProfileByWithholdingElectionReferenceID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.PayrollProfile], error) {
	if s == nil {
		return operation.Page[entity.PayrollProfile]{}, hrerrors.ErrUnavailable
	}
	return s.payrollProfile.List(ctx, []domainrepo.Filter{{Column: "withholding_election_reference_id", Values: []any{value}}}, limit, offset)
}

// CreatePayComponent menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreatePayComponent(ctx context.Context, in CreatePayComponentInput) (*entity.PayComponent, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.payComponent.Create(ctx, func() (*entity.PayComponent, error) { return createPayComponent(in) })
}

// UpdatePayComponent menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdatePayComponent(ctx context.Context, key string, in UpdatePayComponentInput) (*entity.PayComponent, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.payComponent.Update(ctx, key, func(value *entity.PayComponent) error { return patchPayComponent(value, in) })
}

// DeletePayComponent menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeletePayComponent(ctx context.Context, key string) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.payComponent.Delete(ctx, key)
}

// GetPayComponentByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetPayComponentByID(ctx context.Context, key string) (*entity.PayComponent, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.payComponent.Get(ctx, key)
}

// GetAllPayComponents menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllPayComponents(ctx context.Context, limit, offset int) (operation.Page[entity.PayComponent], error) {
	if s == nil {
		return operation.Page[entity.PayComponent]{}, hrerrors.ErrUnavailable
	}
	return s.payComponent.List(ctx, nil, limit, offset)
}

// FindPayComponentByComponentCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPayComponentByComponentCode(ctx context.Context, value string, limit, offset int) (operation.Page[entity.PayComponent], error) {
	if s == nil {
		return operation.Page[entity.PayComponent]{}, hrerrors.ErrUnavailable
	}
	return s.payComponent.List(ctx, []domainrepo.Filter{{Column: "component_code", Values: []any{value}}}, limit, offset)
}

// FindPayComponentByComponentType menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPayComponentByComponentType(ctx context.Context, value string, limit, offset int) (operation.Page[entity.PayComponent], error) {
	if s == nil {
		return operation.Page[entity.PayComponent]{}, hrerrors.ErrUnavailable
	}
	return s.payComponent.List(ctx, []domainrepo.Filter{{Column: "component_type", Values: []any{value}}}, limit, offset)
}

// FindPayComponentByTaxTreatmentCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPayComponentByTaxTreatmentCode(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.PayComponent], error) {
	if s == nil {
		return operation.Page[entity.PayComponent]{}, hrerrors.ErrUnavailable
	}
	return s.payComponent.List(ctx, []domainrepo.Filter{{Column: "tax_treatment_code", Values: []any{value}}}, limit, offset)
}

// FindPayComponentByJurisdictionCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPayComponentByJurisdictionCode(ctx context.Context, value string, limit, offset int) (operation.Page[entity.PayComponent], error) {
	if s == nil {
		return operation.Page[entity.PayComponent]{}, hrerrors.ErrUnavailable
	}
	return s.payComponent.List(ctx, []domainrepo.Filter{{Column: "jurisdiction_code", Values: []any{value}}}, limit, offset)
}

// FindPayComponentByCurrencyCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPayComponentByCurrencyCode(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.PayComponent], error) {
	if s == nil {
		return operation.Page[entity.PayComponent]{}, hrerrors.ErrUnavailable
	}
	return s.payComponent.List(ctx, []domainrepo.Filter{{Column: "currency_code", Values: []any{value}}}, limit, offset)
}

// FindPayComponentByRuleSetID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPayComponentByRuleSetID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.PayComponent], error) {
	if s == nil {
		return operation.Page[entity.PayComponent]{}, hrerrors.ErrUnavailable
	}
	return s.payComponent.List(ctx, []domainrepo.Filter{{Column: "rule_set_id", Values: []any{value}}}, limit, offset)
}

// FindPayComponentByFormulaReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPayComponentByFormulaReferenceID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.PayComponent], error) {
	if s == nil {
		return operation.Page[entity.PayComponent]{}, hrerrors.ErrUnavailable
	}
	return s.payComponent.List(ctx, []domainrepo.Filter{{Column: "formula_reference_id", Values: []any{value}}}, limit, offset)
}

// CreatePayrollPeriod menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreatePayrollPeriod(ctx context.Context, in CreatePayrollPeriodInput) (*entity.PayrollPeriod, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.payrollPeriod.Create(ctx, func() (*entity.PayrollPeriod, error) { return createPayrollPeriod(in) })
}

// UpdatePayrollPeriod menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdatePayrollPeriod(ctx context.Context, key string, in UpdatePayrollPeriodInput) (*entity.PayrollPeriod, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.payrollPeriod.Update(ctx, key, func(value *entity.PayrollPeriod) error { return patchPayrollPeriod(value, in) })
}

// DeletePayrollPeriod menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeletePayrollPeriod(ctx context.Context, key string) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.payrollPeriod.Delete(ctx, key)
}

// GetPayrollPeriodByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetPayrollPeriodByID(ctx context.Context, key string) (*entity.PayrollPeriod, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.payrollPeriod.Get(ctx, key)
}

// GetAllPayrollPeriods menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllPayrollPeriods(ctx context.Context, limit, offset int) (operation.Page[entity.PayrollPeriod], error) {
	if s == nil {
		return operation.Page[entity.PayrollPeriod]{}, hrerrors.ErrUnavailable
	}
	return s.payrollPeriod.List(ctx, nil, limit, offset)
}

// FindPayrollPeriodByPayrollGroupID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPayrollPeriodByPayrollGroupID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.PayrollPeriod], error) {
	if s == nil {
		return operation.Page[entity.PayrollPeriod]{}, hrerrors.ErrUnavailable
	}
	return s.payrollPeriod.List(ctx, []domainrepo.Filter{{Column: "payroll_group_id", Values: []any{value}}}, limit, offset)
}

// FindPayrollPeriodByEmployingLegalEntityReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPayrollPeriodByEmployingLegalEntityReferenceID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.PayrollPeriod], error) {
	if s == nil {
		return operation.Page[entity.PayrollPeriod]{}, hrerrors.ErrUnavailable
	}
	return s.payrollPeriod.List(ctx, []domainrepo.Filter{{Column: "employing_legal_entity_reference_id", Values: []any{value}}}, limit, offset)
}

// FindPayrollPeriodByCurrencyCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPayrollPeriodByCurrencyCode(ctx context.Context, value string, limit, offset int) (operation.Page[entity.PayrollPeriod], error) {
	if s == nil {
		return operation.Page[entity.PayrollPeriod]{}, hrerrors.ErrUnavailable
	}
	return s.payrollPeriod.List(ctx, []domainrepo.Filter{{Column: "currency_code", Values: []any{value}}}, limit, offset)
}

// FindPayrollPeriodByJurisdictionCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPayrollPeriodByJurisdictionCode(ctx context.Context, value string, limit, offset int) (operation.Page[entity.PayrollPeriod], error) {
	if s == nil {
		return operation.Page[entity.PayrollPeriod]{}, hrerrors.ErrUnavailable
	}
	return s.payrollPeriod.List(ctx, []domainrepo.Filter{{Column: "jurisdiction_code", Values: []any{value}}}, limit, offset)
}

// FindPayrollPeriodByPeriodCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPayrollPeriodByPeriodCode(ctx context.Context, value string, limit, offset int) (operation.Page[entity.PayrollPeriod], error) {
	if s == nil {
		return operation.Page[entity.PayrollPeriod]{}, hrerrors.ErrUnavailable
	}
	return s.payrollPeriod.List(ctx, []domainrepo.Filter{{Column: "period_code", Values: []any{value}}}, limit, offset)
}

// FindPayrollPeriodByPeriodStatus menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPayrollPeriodByPeriodStatus(ctx context.Context, value string, limit, offset int) (operation.Page[entity.PayrollPeriod], error) {
	if s == nil {
		return operation.Page[entity.PayrollPeriod]{}, hrerrors.ErrUnavailable
	}
	return s.payrollPeriod.List(ctx, []domainrepo.Filter{{Column: "period_status", Values: []any{value}}}, limit, offset)
}

// FindPayrollPeriodByCalendarReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPayrollPeriodByCalendarReferenceID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.PayrollPeriod], error) {
	if s == nil {
		return operation.Page[entity.PayrollPeriod]{}, hrerrors.ErrUnavailable
	}
	return s.payrollPeriod.List(ctx, []domainrepo.Filter{{Column: "calendar_reference_id", Values: []any{value}}}, limit, offset)
}

// CreatePayrollRun menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreatePayrollRun(ctx context.Context, in CreatePayrollRunInput) (*entity.PayrollRun, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.payrollRun.Create(ctx, func() (*entity.PayrollRun, error) { return createPayrollRun(in) })
}

// UpdatePayrollRun menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdatePayrollRun(ctx context.Context, key string, in UpdatePayrollRunInput) (*entity.PayrollRun, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.payrollRun.Update(ctx, key, func(value *entity.PayrollRun) error { return patchPayrollRun(value, in) })
}

// DeletePayrollRun menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeletePayrollRun(ctx context.Context, key string) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.payrollRun.Delete(ctx, key)
}

// GetPayrollRunByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetPayrollRunByID(ctx context.Context, key string) (*entity.PayrollRun, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.payrollRun.Get(ctx, key)
}

// GetAllPayrollRuns menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllPayrollRuns(ctx context.Context, limit, offset int) (operation.Page[entity.PayrollRun], error) {
	if s == nil {
		return operation.Page[entity.PayrollRun]{}, hrerrors.ErrUnavailable
	}
	return s.payrollRun.List(ctx, nil, limit, offset)
}

// FindPayrollRunByPayrollPeriodID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPayrollRunByPayrollPeriodID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.PayrollRun], error) {
	if s == nil {
		return operation.Page[entity.PayrollRun]{}, hrerrors.ErrUnavailable
	}
	return s.payrollRun.List(ctx, []domainrepo.Filter{{Column: "payroll_period_id", Values: []any{value}}}, limit, offset)
}

// FindPayrollRunByPayrollGroupID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPayrollRunByPayrollGroupID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.PayrollRun], error) {
	if s == nil {
		return operation.Page[entity.PayrollRun]{}, hrerrors.ErrUnavailable
	}
	return s.payrollRun.List(ctx, []domainrepo.Filter{{Column: "payroll_group_id", Values: []any{value}}}, limit, offset)
}

// FindPayrollRunByEmployingLegalEntityReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPayrollRunByEmployingLegalEntityReferenceID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.PayrollRun], error) {
	if s == nil {
		return operation.Page[entity.PayrollRun]{}, hrerrors.ErrUnavailable
	}
	return s.payrollRun.List(ctx, []domainrepo.Filter{{Column: "employing_legal_entity_reference_id", Values: []any{value}}}, limit, offset)
}

// FindPayrollRunByCurrencyCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPayrollRunByCurrencyCode(ctx context.Context, value string, limit, offset int) (operation.Page[entity.PayrollRun], error) {
	if s == nil {
		return operation.Page[entity.PayrollRun]{}, hrerrors.ErrUnavailable
	}
	return s.payrollRun.List(ctx, []domainrepo.Filter{{Column: "currency_code", Values: []any{value}}}, limit, offset)
}

// FindPayrollRunByRuleSetID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPayrollRunByRuleSetID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.PayrollRun], error) {
	if s == nil {
		return operation.Page[entity.PayrollRun]{}, hrerrors.ErrUnavailable
	}
	return s.payrollRun.List(ctx, []domainrepo.Filter{{Column: "rule_set_id", Values: []any{value}}}, limit, offset)
}

// FindPayrollRunByOriginalRunID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPayrollRunByOriginalRunID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.PayrollRun], error) {
	if s == nil {
		return operation.Page[entity.PayrollRun]{}, hrerrors.ErrUnavailable
	}
	return s.payrollRun.List(ctx, []domainrepo.Filter{{Column: "original_run_id", Values: []any{value}}}, limit, offset)
}

// FindPayrollRunByRunNumber menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPayrollRunByRunNumber(ctx context.Context, value int32, limit, offset int) (operation.Page[entity.PayrollRun], error) {
	if s == nil {
		return operation.Page[entity.PayrollRun]{}, hrerrors.ErrUnavailable
	}
	return s.payrollRun.List(ctx, []domainrepo.Filter{{Column: "run_number", Values: []any{value}}}, limit, offset)
}

// FindPayrollRunByRunType menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPayrollRunByRunType(ctx context.Context, value string, limit, offset int) (operation.Page[entity.PayrollRun], error) {
	if s == nil {
		return operation.Page[entity.PayrollRun]{}, hrerrors.ErrUnavailable
	}
	return s.payrollRun.List(ctx, []domainrepo.Filter{{Column: "run_type", Values: []any{value}}}, limit, offset)
}

// FindPayrollRunByStatus menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPayrollRunByStatus(ctx context.Context, value entity.PayrollRunStatus, limit, offset int) (operation.Page[entity.PayrollRun], error) {
	if s == nil {
		return operation.Page[entity.PayrollRun]{}, hrerrors.ErrUnavailable
	}
	return s.payrollRun.List(ctx, []domainrepo.Filter{{Column: "status", Values: []any{value}}}, limit, offset)
}

// FindPayrollRunByJurisdictionCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPayrollRunByJurisdictionCode(ctx context.Context, value string, limit, offset int) (operation.Page[entity.PayrollRun], error) {
	if s == nil {
		return operation.Page[entity.PayrollRun]{}, hrerrors.ErrUnavailable
	}
	return s.payrollRun.List(ctx, []domainrepo.Filter{{Column: "jurisdiction_code", Values: []any{value}}}, limit, offset)
}

// FindPayrollRunByInputSnapshotReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPayrollRunByInputSnapshotReferenceID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.PayrollRun], error) {
	if s == nil {
		return operation.Page[entity.PayrollRun]{}, hrerrors.ErrUnavailable
	}
	return s.payrollRun.List(ctx, []domainrepo.Filter{{Column: "input_snapshot_reference_id", Values: []any{value}}}, limit, offset)
}

// FindPayrollRunByApprovedByPrincipalID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPayrollRunByApprovedByPrincipalID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.PayrollRun], error) {
	if s == nil {
		return operation.Page[entity.PayrollRun]{}, hrerrors.ErrUnavailable
	}
	return s.payrollRun.List(ctx, []domainrepo.Filter{{Column: "approved_by_principal_id", Values: []any{value}}}, limit, offset)
}

// FindPayrollRunByFinanceJournalReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPayrollRunByFinanceJournalReferenceID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.PayrollRun], error) {
	if s == nil {
		return operation.Page[entity.PayrollRun]{}, hrerrors.ErrUnavailable
	}
	return s.payrollRun.List(ctx, []domainrepo.Filter{{Column: "finance_journal_reference_id", Values: []any{value}}}, limit, offset)
}

// FindPayrollRunByPaymentBatchReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPayrollRunByPaymentBatchReferenceID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.PayrollRun], error) {
	if s == nil {
		return operation.Page[entity.PayrollRun]{}, hrerrors.ErrUnavailable
	}
	return s.payrollRun.List(ctx, []domainrepo.Filter{{Column: "payment_batch_reference_id", Values: []any{value}}}, limit, offset)
}

// FindPayrollRunByFinanceSettlementEventReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPayrollRunByFinanceSettlementEventReferenceID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.PayrollRun], error) {
	if s == nil {
		return operation.Page[entity.PayrollRun]{}, hrerrors.ErrUnavailable
	}
	return s.payrollRun.List(ctx, []domainrepo.Filter{{Column: "finance_settlement_event_reference_id", Values: []any{value}}}, limit, offset)
}

// FindPayrollRunByStartedByPrincipalID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPayrollRunByStartedByPrincipalID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.PayrollRun], error) {
	if s == nil {
		return operation.Page[entity.PayrollRun]{}, hrerrors.ErrUnavailable
	}
	return s.payrollRun.List(ctx, []domainrepo.Filter{{Column: "started_by_principal_id", Values: []any{value}}}, limit, offset)
}

// FindPayrollRunByIdempotencyKey menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPayrollRunByIdempotencyKey(ctx context.Context, value string, limit, offset int) (operation.Page[entity.PayrollRun], error) {
	if s == nil {
		return operation.Page[entity.PayrollRun]{}, hrerrors.ErrUnavailable
	}
	return s.payrollRun.List(ctx, []domainrepo.Filter{{Column: "idempotency_key", Values: []any{value}}}, limit, offset)
}

// CreatePayrollResult menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreatePayrollResult(ctx context.Context, in CreatePayrollResultInput) (*entity.PayrollResult, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.payrollResult.Create(ctx, func() (*entity.PayrollResult, error) { return createPayrollResult(in) })
}

// UpdatePayrollResult menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdatePayrollResult(ctx context.Context, key string, in UpdatePayrollResultInput) (*entity.PayrollResult, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.payrollResult.Update(ctx, key, func(value *entity.PayrollResult) error { return patchPayrollResult(value, in) })
}

// DeletePayrollResult menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeletePayrollResult(ctx context.Context, key string) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.payrollResult.Delete(ctx, key)
}

// GetPayrollResultByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetPayrollResultByID(ctx context.Context, key string) (*entity.PayrollResult, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.payrollResult.Get(ctx, key)
}

// GetAllPayrollResults menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllPayrollResults(ctx context.Context, limit, offset int) (operation.Page[entity.PayrollResult], error) {
	if s == nil {
		return operation.Page[entity.PayrollResult]{}, hrerrors.ErrUnavailable
	}
	return s.payrollResult.List(ctx, nil, limit, offset)
}

// FindPayrollResultByPayrollRunID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPayrollResultByPayrollRunID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.PayrollResult], error) {
	if s == nil {
		return operation.Page[entity.PayrollResult]{}, hrerrors.ErrUnavailable
	}
	return s.payrollResult.List(ctx, []domainrepo.Filter{{Column: "payroll_run_id", Values: []any{value}}}, limit, offset)
}

// FindPayrollResultByPayrollGroupID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPayrollResultByPayrollGroupID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.PayrollResult], error) {
	if s == nil {
		return operation.Page[entity.PayrollResult]{}, hrerrors.ErrUnavailable
	}
	return s.payrollResult.List(ctx, []domainrepo.Filter{{Column: "payroll_group_id", Values: []any{value}}}, limit, offset)
}

// FindPayrollResultByWorkerID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPayrollResultByWorkerID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.PayrollResult], error) {
	if s == nil {
		return operation.Page[entity.PayrollResult]{}, hrerrors.ErrUnavailable
	}
	return s.payrollResult.List(ctx, []domainrepo.Filter{{Column: "worker_id", Values: []any{value}}}, limit, offset)
}

// FindPayrollResultByPayrollProfileReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPayrollResultByPayrollProfileReferenceID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.PayrollResult], error) {
	if s == nil {
		return operation.Page[entity.PayrollResult]{}, hrerrors.ErrUnavailable
	}
	return s.payrollResult.List(ctx, []domainrepo.Filter{{Column: "payroll_profile_reference_id", Values: []any{value}}}, limit, offset)
}

// FindPayrollResultByCurrencyCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPayrollResultByCurrencyCode(ctx context.Context, value string, limit, offset int) (operation.Page[entity.PayrollResult], error) {
	if s == nil {
		return operation.Page[entity.PayrollResult]{}, hrerrors.ErrUnavailable
	}
	return s.payrollResult.List(ctx, []domainrepo.Filter{{Column: "currency_code", Values: []any{value}}}, limit, offset)
}

// FindPayrollResultByResultStatus menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPayrollResultByResultStatus(ctx context.Context, value string, limit, offset int) (operation.Page[entity.PayrollResult], error) {
	if s == nil {
		return operation.Page[entity.PayrollResult]{}, hrerrors.ErrUnavailable
	}
	return s.payrollResult.List(ctx, []domainrepo.Filter{{Column: "result_status", Values: []any{value}}}, limit, offset)
}

// FindPayrollResultByPaymentReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPayrollResultByPaymentReferenceID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.PayrollResult], error) {
	if s == nil {
		return operation.Page[entity.PayrollResult]{}, hrerrors.ErrUnavailable
	}
	return s.payrollResult.List(ctx, []domainrepo.Filter{{Column: "payment_reference_id", Values: []any{value}}}, limit, offset)
}

// FindPayrollResultByPayslipDocumentReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPayrollResultByPayslipDocumentReferenceID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.PayrollResult], error) {
	if s == nil {
		return operation.Page[entity.PayrollResult]{}, hrerrors.ErrUnavailable
	}
	return s.payrollResult.List(ctx, []domainrepo.Filter{{Column: "payslip_document_reference_id", Values: []any{value}}}, limit, offset)
}

// CreatePayrollResultLine menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreatePayrollResultLine(ctx context.Context, in CreatePayrollResultLineInput) (*entity.PayrollResultLine, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.payrollResultLine.Create(ctx, func() (*entity.PayrollResultLine, error) { return createPayrollResultLine(in) })
}

// UpdatePayrollResultLine menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdatePayrollResultLine(ctx context.Context, key string, in UpdatePayrollResultLineInput) (*entity.PayrollResultLine, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.payrollResultLine.Update(ctx, key, func(value *entity.PayrollResultLine) error { return patchPayrollResultLine(value, in) })
}

// DeletePayrollResultLine menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeletePayrollResultLine(ctx context.Context, key string) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.payrollResultLine.Delete(ctx, key)
}

// GetPayrollResultLineByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetPayrollResultLineByID(ctx context.Context, key string) (*entity.PayrollResultLine, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.payrollResultLine.Get(ctx, key)
}

// GetAllPayrollResultLines menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllPayrollResultLines(ctx context.Context, limit, offset int) (operation.Page[entity.PayrollResultLine], error) {
	if s == nil {
		return operation.Page[entity.PayrollResultLine]{}, hrerrors.ErrUnavailable
	}
	return s.payrollResultLine.List(ctx, nil, limit, offset)
}

// FindPayrollResultLineByPayrollResultID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPayrollResultLineByPayrollResultID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.PayrollResultLine], error) {
	if s == nil {
		return operation.Page[entity.PayrollResultLine]{}, hrerrors.ErrUnavailable
	}
	return s.payrollResultLine.List(ctx, []domainrepo.Filter{{Column: "payroll_result_id", Values: []any{value}}}, limit, offset)
}

// FindPayrollResultLineByComponentID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPayrollResultLineByComponentID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.PayrollResultLine], error) {
	if s == nil {
		return operation.Page[entity.PayrollResultLine]{}, hrerrors.ErrUnavailable
	}
	return s.payrollResultLine.List(ctx, []domainrepo.Filter{{Column: "component_id", Values: []any{value}}}, limit, offset)
}

// CreatePayrollAdjustment menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreatePayrollAdjustment(ctx context.Context, in CreatePayrollAdjustmentInput) (*entity.PayrollAdjustment, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.payrollAdjustment.Create(ctx, func() (*entity.PayrollAdjustment, error) { return createPayrollAdjustment(in) })
}

// UpdatePayrollAdjustment menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdatePayrollAdjustment(ctx context.Context, key string, in UpdatePayrollAdjustmentInput) (*entity.PayrollAdjustment, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.payrollAdjustment.Update(ctx, key, func(value *entity.PayrollAdjustment) error { return patchPayrollAdjustment(value, in) })
}

// DeletePayrollAdjustment menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeletePayrollAdjustment(ctx context.Context, key string) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.payrollAdjustment.Delete(ctx, key)
}

// GetPayrollAdjustmentByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetPayrollAdjustmentByID(ctx context.Context, key string) (*entity.PayrollAdjustment, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.payrollAdjustment.Get(ctx, key)
}

// GetAllPayrollAdjustments menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllPayrollAdjustments(ctx context.Context, limit, offset int) (operation.Page[entity.PayrollAdjustment], error) {
	if s == nil {
		return operation.Page[entity.PayrollAdjustment]{}, hrerrors.ErrUnavailable
	}
	return s.payrollAdjustment.List(ctx, nil, limit, offset)
}

// FindPayrollAdjustmentByWorkerID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPayrollAdjustmentByWorkerID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.PayrollAdjustment], error) {
	if s == nil {
		return operation.Page[entity.PayrollAdjustment]{}, hrerrors.ErrUnavailable
	}
	return s.payrollAdjustment.List(ctx, []domainrepo.Filter{{Column: "worker_id", Values: []any{value}}}, limit, offset)
}

// FindPayrollAdjustmentByPayrollPeriodID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPayrollAdjustmentByPayrollPeriodID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.PayrollAdjustment], error) {
	if s == nil {
		return operation.Page[entity.PayrollAdjustment]{}, hrerrors.ErrUnavailable
	}
	return s.payrollAdjustment.List(ctx, []domainrepo.Filter{{Column: "payroll_period_id", Values: []any{value}}}, limit, offset)
}

// FindPayrollAdjustmentByComponentID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPayrollAdjustmentByComponentID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.PayrollAdjustment], error) {
	if s == nil {
		return operation.Page[entity.PayrollAdjustment]{}, hrerrors.ErrUnavailable
	}
	return s.payrollAdjustment.List(ctx, []domainrepo.Filter{{Column: "component_id", Values: []any{value}}}, limit, offset)
}

// FindPayrollAdjustmentByAdjustmentType menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPayrollAdjustmentByAdjustmentType(ctx context.Context, value string, limit, offset int) (operation.Page[entity.PayrollAdjustment], error) {
	if s == nil {
		return operation.Page[entity.PayrollAdjustment]{}, hrerrors.ErrUnavailable
	}
	return s.payrollAdjustment.List(ctx, []domainrepo.Filter{{Column: "adjustment_type", Values: []any{value}}}, limit, offset)
}

// FindPayrollAdjustmentByCurrencyCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPayrollAdjustmentByCurrencyCode(ctx context.Context, value string, limit, offset int) (operation.Page[entity.PayrollAdjustment], error) {
	if s == nil {
		return operation.Page[entity.PayrollAdjustment]{}, hrerrors.ErrUnavailable
	}
	return s.payrollAdjustment.List(ctx, []domainrepo.Filter{{Column: "currency_code", Values: []any{value}}}, limit, offset)
}

// FindPayrollAdjustmentByRequestedByPrincipalID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPayrollAdjustmentByRequestedByPrincipalID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.PayrollAdjustment], error) {
	if s == nil {
		return operation.Page[entity.PayrollAdjustment]{}, hrerrors.ErrUnavailable
	}
	return s.payrollAdjustment.List(ctx, []domainrepo.Filter{{Column: "requested_by_principal_id", Values: []any{value}}}, limit, offset)
}

// FindPayrollAdjustmentByApprovedByPrincipalID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPayrollAdjustmentByApprovedByPrincipalID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.PayrollAdjustment], error) {
	if s == nil {
		return operation.Page[entity.PayrollAdjustment]{}, hrerrors.ErrUnavailable
	}
	return s.payrollAdjustment.List(ctx, []domainrepo.Filter{{Column: "approved_by_principal_id", Values: []any{value}}}, limit, offset)
}

// FindPayrollAdjustmentByIdempotencyKey menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPayrollAdjustmentByIdempotencyKey(ctx context.Context, value string, limit, offset int) (operation.Page[entity.PayrollAdjustment], error) {
	if s == nil {
		return operation.Page[entity.PayrollAdjustment]{}, hrerrors.ErrUnavailable
	}
	return s.payrollAdjustment.List(ctx, []domainrepo.Filter{{Column: "idempotency_key", Values: []any{value}}}, limit, offset)
}

// CreatePayrollGroup menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreatePayrollGroup(ctx context.Context, in CreatePayrollGroupInput) (*entity.PayrollGroup, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.payrollGroup.Create(ctx, func() (*entity.PayrollGroup, error) { return createPayrollGroup(in) })
}

// UpdatePayrollGroup menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdatePayrollGroup(ctx context.Context, key string, in UpdatePayrollGroupInput) (*entity.PayrollGroup, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.payrollGroup.Update(ctx, key, func(value *entity.PayrollGroup) error { return patchPayrollGroup(value, in) })
}

// DeletePayrollGroup menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeletePayrollGroup(ctx context.Context, key string) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.payrollGroup.Delete(ctx, key)
}

// GetPayrollGroupByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetPayrollGroupByID(ctx context.Context, key string) (*entity.PayrollGroup, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.payrollGroup.Get(ctx, key)
}

// GetAllPayrollGroups menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllPayrollGroups(ctx context.Context, limit, offset int) (operation.Page[entity.PayrollGroup], error) {
	if s == nil {
		return operation.Page[entity.PayrollGroup]{}, hrerrors.ErrUnavailable
	}
	return s.payrollGroup.List(ctx, nil, limit, offset)
}

// FindPayrollGroupByGroupCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPayrollGroupByGroupCode(ctx context.Context, value string, limit, offset int) (operation.Page[entity.PayrollGroup], error) {
	if s == nil {
		return operation.Page[entity.PayrollGroup]{}, hrerrors.ErrUnavailable
	}
	return s.payrollGroup.List(ctx, []domainrepo.Filter{{Column: "group_code", Values: []any{value}}}, limit, offset)
}

// FindPayrollGroupByEmployingLegalEntityReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPayrollGroupByEmployingLegalEntityReferenceID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.PayrollGroup], error) {
	if s == nil {
		return operation.Page[entity.PayrollGroup]{}, hrerrors.ErrUnavailable
	}
	return s.payrollGroup.List(ctx, []domainrepo.Filter{{Column: "employing_legal_entity_reference_id", Values: []any{value}}}, limit, offset)
}

// FindPayrollGroupByJurisdictionCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPayrollGroupByJurisdictionCode(ctx context.Context, value string, limit, offset int) (operation.Page[entity.PayrollGroup], error) {
	if s == nil {
		return operation.Page[entity.PayrollGroup]{}, hrerrors.ErrUnavailable
	}
	return s.payrollGroup.List(ctx, []domainrepo.Filter{{Column: "jurisdiction_code", Values: []any{value}}}, limit, offset)
}

// FindPayrollGroupByCurrencyCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPayrollGroupByCurrencyCode(ctx context.Context, value string, limit, offset int) (operation.Page[entity.PayrollGroup], error) {
	if s == nil {
		return operation.Page[entity.PayrollGroup]{}, hrerrors.ErrUnavailable
	}
	return s.payrollGroup.List(ctx, []domainrepo.Filter{{Column: "currency_code", Values: []any{value}}}, limit, offset)
}

// CreatePayrollRuleSet menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreatePayrollRuleSet(ctx context.Context, in CreatePayrollRuleSetInput) (*entity.PayrollRuleSet, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.payrollRuleSet.Create(ctx, func() (*entity.PayrollRuleSet, error) { return createPayrollRuleSet(in) })
}

// UpdatePayrollRuleSet menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdatePayrollRuleSet(ctx context.Context, key string, in UpdatePayrollRuleSetInput) (*entity.PayrollRuleSet, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.payrollRuleSet.Update(ctx, key, func(value *entity.PayrollRuleSet) error { return patchPayrollRuleSet(value, in) })
}

// DeletePayrollRuleSet menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeletePayrollRuleSet(ctx context.Context, key string) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.payrollRuleSet.Delete(ctx, key)
}

// GetPayrollRuleSetByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetPayrollRuleSetByID(ctx context.Context, key string) (*entity.PayrollRuleSet, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.payrollRuleSet.Get(ctx, key)
}

// GetAllPayrollRuleSets menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllPayrollRuleSets(ctx context.Context, limit, offset int) (operation.Page[entity.PayrollRuleSet], error) {
	if s == nil {
		return operation.Page[entity.PayrollRuleSet]{}, hrerrors.ErrUnavailable
	}
	return s.payrollRuleSet.List(ctx, nil, limit, offset)
}

// FindPayrollRuleSetByRuleCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPayrollRuleSetByRuleCode(ctx context.Context, value string, limit, offset int) (operation.Page[entity.PayrollRuleSet], error) {
	if s == nil {
		return operation.Page[entity.PayrollRuleSet]{}, hrerrors.ErrUnavailable
	}
	return s.payrollRuleSet.List(ctx, []domainrepo.Filter{{Column: "rule_code", Values: []any{value}}}, limit, offset)
}

// FindPayrollRuleSetByJurisdictionCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPayrollRuleSetByJurisdictionCode(ctx context.Context, value string, limit, offset int) (operation.Page[entity.PayrollRuleSet], error) {
	if s == nil {
		return operation.Page[entity.PayrollRuleSet]{}, hrerrors.ErrUnavailable
	}
	return s.payrollRuleSet.List(ctx, []domainrepo.Filter{{Column: "jurisdiction_code", Values: []any{value}}}, limit, offset)
}

// FindPayrollRuleSetByArtifactDocumentVersionReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPayrollRuleSetByArtifactDocumentVersionReferenceID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.PayrollRuleSet], error) {
	if s == nil {
		return operation.Page[entity.PayrollRuleSet]{}, hrerrors.ErrUnavailable
	}
	return s.payrollRuleSet.List(ctx, []domainrepo.Filter{{Column: "artifact_document_version_reference_id", Values: []any{value}}}, limit, offset)
}

// FindPayrollRuleSetByLegalBasisDocumentVersionReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPayrollRuleSetByLegalBasisDocumentVersionReferenceID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.PayrollRuleSet], error) {
	if s == nil {
		return operation.Page[entity.PayrollRuleSet]{}, hrerrors.ErrUnavailable
	}
	return s.payrollRuleSet.List(ctx, []domainrepo.Filter{{Column: "legal_basis_document_version_reference_id", Values: []any{value}}}, limit, offset)
}

// FindPayrollRuleSetByApprovedByPrincipalID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPayrollRuleSetByApprovedByPrincipalID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.PayrollRuleSet], error) {
	if s == nil {
		return operation.Page[entity.PayrollRuleSet]{}, hrerrors.ErrUnavailable
	}
	return s.payrollRuleSet.List(ctx, []domainrepo.Filter{{Column: "approved_by_principal_id", Values: []any{value}}}, limit, offset)
}
