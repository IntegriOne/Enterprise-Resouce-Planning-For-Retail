package human_resource

import (
	"context"

	"github.com/google/uuid"
	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	payroll "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/payroll/v1"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func toModelPayrollProfile(row *payroll.PayrollProfile) (*model.PayrollProfile, error) {
	if row == nil {
		return nil, middleware.ErrUnavailable("invalid human resource payroll-profile response", nil)
	}
	if id, err := uuid.Parse(row.GetId()); err != nil || id == uuid.Nil {
		return nil, middleware.ErrUnavailable("invalid human resource payroll-profile ID", nil)
	}
	out := &model.PayrollProfile{ID: nodeID("payroll-profile", row.GetId()), RawID: row.GetId()}
	out.WorkerID = row.GetWorkerId()
	out.PayrollCountryCode = row.GetPayrollCountryCode()
	out.PayrollJurisdictionCode = row.GetPayrollJurisdictionCode()
	out.PayrollGroupID = row.GetPayrollGroupId()
	out.EmployingLegalEntityReferenceID = row.GetEmployingLegalEntityReferenceId()
	out.CurrencyCode = row.GetCurrencyCode()
	out.TaxProfileReferenceID = row.TaxProfileReferenceId
	out.BankAccountTokenReferenceID = row.BankAccountTokenReferenceId
	out.TaxStatusCode = row.TaxStatusCode
	out.WithholdingElectionReferenceID = row.WithholdingElectionReferenceId
	out.EffectiveFrom = row.GetEffectiveFrom()
	out.EffectiveTo = row.EffectiveTo
	out.IsActive = row.GetIsActive()
	return out, nil
}

// PayrollProfile membaca detail payroll-profile berdasarkan ID.
func (a *Adapter) PayrollProfile(ctx context.Context, rawID string) (*model.PayrollProfile, error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_payroll_profiles", "human_resource.payroll_profiles.read")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "payroll-profile", "payroll-profile")
	if err != nil {
		return nil, err
	}
	if a.payroll == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.payroll.GetPayrollProfileByID(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapReadError(err, "payroll-profile")
	}
	return toModelPayrollProfile(resp.GetItem())
}

// PayrollProfiles mendaftar dan memfilter payroll-profile dengan paginasi.
func (a *Adapter) PayrollProfiles(ctx context.Context, limit, offset int, f_worker_id *string, f_payroll_country_code *string, f_payroll_jurisdiction_code *string, f_payroll_group_id *string, f_employing_legal_entity_reference_id *string, f_currency_code *string, f_tax_profile_reference_id *string, f_bank_account_token_reference_id *string, f_tax_status_code *string, f_withholding_election_reference_id *string) (*Page[model.PayrollProfile], error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_payroll_profiles", "human_resource.payroll_profiles.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid payroll-profile pagination", nil)
	}
	if a.payroll == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	idx, err := pickSingleFilter(f_worker_id != nil, f_payroll_country_code != nil, f_payroll_jurisdiction_code != nil, f_payroll_group_id != nil, f_employing_legal_entity_reference_id != nil, f_currency_code != nil, f_tax_profile_reference_id != nil, f_bank_account_token_reference_id != nil, f_tax_status_code != nil, f_withholding_election_reference_id != nil)
	if err != nil {
		return nil, err
	}
	var resp *payroll.PayrollProfilesResponse
	switch idx {
	case 0:
		v, fErr := validateFilterString(f_worker_id, "worker_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.payroll.FindPayrollProfileByWorkerID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 1:
		v, fErr := validateFilterString(f_payroll_country_code, "payroll_country_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.payroll.FindPayrollProfileByPayrollCountryCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 2:
		v, fErr := validateFilterString(f_payroll_jurisdiction_code, "payroll_jurisdiction_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.payroll.FindPayrollProfileByPayrollJurisdictionCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 3:
		v, fErr := validateFilterString(f_payroll_group_id, "payroll_group_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.payroll.FindPayrollProfileByPayrollGroupID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 4:
		v, fErr := validateFilterString(f_employing_legal_entity_reference_id, "employing_legal_entity_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.payroll.FindPayrollProfileByEmployingLegalEntityReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 5:
		v, fErr := validateFilterString(f_currency_code, "currency_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.payroll.FindPayrollProfileByCurrencyCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 6:
		v, fErr := validateFilterString(f_tax_profile_reference_id, "tax_profile_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.payroll.FindPayrollProfileByTaxProfileReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 7:
		v, fErr := validateFilterString(f_bank_account_token_reference_id, "bank_account_token_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.payroll.FindPayrollProfileByBankAccountTokenReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 8:
		v, fErr := validateFilterString(f_tax_status_code, "tax_status_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.payroll.FindPayrollProfileByTaxStatusCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 9:
		v, fErr := validateFilterString(f_withholding_election_reference_id, "withholding_election_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.payroll.FindPayrollProfileByWithholdingElectionReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	default:
		resp, err = a.payroll.GetAllPayrollProfiles(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "payroll-profile")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "payroll-profile", toModelPayrollProfile)
}

// CreatePayrollProfile membuat payroll-profile baru melalui HRM.
func (a *Adapter) CreatePayrollProfile(ctx context.Context, input model.CreatePayrollProfileInput) (*model.PayrollProfile, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_payroll_profiles", "create", "human_resource.payroll_profiles.create")
	if err != nil {
		return nil, err
	}
	if a.payroll == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &payroll.PayrollProfileInput{}
	patch.WorkerId = input.WorkerID
	patch.PayrollCountryCode = input.PayrollCountryCode
	patch.PayrollJurisdictionCode = input.PayrollJurisdictionCode
	patch.PayrollGroupId = input.PayrollGroupID
	patch.EmployingLegalEntityReferenceId = input.EmployingLegalEntityReferenceID
	patch.CurrencyCode = input.CurrencyCode
	patch.TaxProfileReferenceId = input.TaxProfileReferenceID
	patch.BankAccountTokenReferenceId = input.BankAccountTokenReferenceID
	patch.TaxStatusCode = input.TaxStatusCode
	patch.WithholdingElectionReferenceId = input.WithholdingElectionReferenceID
	patch.EffectiveFrom = input.EffectiveFrom
	patch.EffectiveTo = input.EffectiveTo
	patch.IsActive = input.IsActive
	resp, err := a.payroll.CreatePayrollProfile(rpcCtx, &payroll.CreatePayrollProfileRequest{Input: patch})
	if err != nil {
		return nil, mapHRMWriteError(err, "payroll-profile")
	}
	return toModelPayrollProfile(resp.GetItem())
}

// UpdatePayrollProfile memperbarui payroll-profile berdasarkan FieldMask.
func (a *Adapter) UpdatePayrollProfile(ctx context.Context, rawID string, input model.UpdatePayrollProfileInput) (*model.PayrollProfile, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_payroll_profiles", "update", "human_resource.payroll_profiles.update")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "payroll-profile", "payroll-profile")
	if err != nil {
		return nil, err
	}
	if a.payroll == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &payroll.PayrollProfileInput{}
	mask := &fieldmaskpb.FieldMask{}
	if err := applyStringMask(mask, "worker_id", input.WorkerID, &patch.WorkerId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "payroll_country_code", input.PayrollCountryCode, &patch.PayrollCountryCode, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "payroll_jurisdiction_code", input.PayrollJurisdictionCode, &patch.PayrollJurisdictionCode, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "payroll_group_id", input.PayrollGroupID, &patch.PayrollGroupId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "employing_legal_entity_reference_id", input.EmployingLegalEntityReferenceID, &patch.EmployingLegalEntityReferenceId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "currency_code", input.CurrencyCode, &patch.CurrencyCode, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "tax_profile_reference_id", input.TaxProfileReferenceID, &patch.TaxProfileReferenceId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "bank_account_token_reference_id", input.BankAccountTokenReferenceID, &patch.BankAccountTokenReferenceId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "tax_status_code", input.TaxStatusCode, &patch.TaxStatusCode, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "withholding_election_reference_id", input.WithholdingElectionReferenceID, &patch.WithholdingElectionReferenceId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "effective_from", input.EffectiveFrom, &patch.EffectiveFrom, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "effective_to", input.EffectiveTo, &patch.EffectiveTo, true); err != nil {
		return nil, err
	}
	if err := applyBoolMask(mask, "is_active", input.IsActive, &patch.IsActive, true); err != nil {
		return nil, err
	}
	if len(mask.Paths) == 0 {
		return nil, middleware.ErrInvalidArgument("payroll-profile update is empty", nil)
	}
	resp, err := a.payroll.UpdatePayrollProfile(rpcCtx, &payroll.UpdatePayrollProfileRequest{Id: id, Input: patch, UpdateMask: mask})
	if err != nil {
		return nil, mapHRMWriteError(err, "payroll-profile")
	}
	return toModelPayrollProfile(resp.GetItem())
}

// DeletePayrollProfile menghapus payroll-profile berdasarkan ID.
func (a *Adapter) DeletePayrollProfile(ctx context.Context, rawID string) (*bool, error) {
	if a == nil || a.payroll == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	return deleteByUUID(ctx, a, rawID, "hr_payroll_profiles", "human_resource.payroll_profiles.delete", "payroll-profile", func(c context.Context, req *common.IDRequest) (*common.EmptyResponse, error) {
		return a.payroll.DeletePayrollProfile(c, req)
	})
}

func toModelPayComponent(row *payroll.PayComponent) (*model.PayComponent, error) {
	if row == nil {
		return nil, middleware.ErrUnavailable("invalid human resource pay-component response", nil)
	}
	if id, err := uuid.Parse(row.GetId()); err != nil || id == uuid.Nil {
		return nil, middleware.ErrUnavailable("invalid human resource pay-component ID", nil)
	}
	out := &model.PayComponent{ID: nodeID("pay-component", row.GetId()), RawID: row.GetId()}
	out.ComponentCode = row.GetComponentCode()
	out.ComponentVersion = int(row.GetComponentVersion())
	out.ComponentName = row.GetComponentName()
	out.ComponentType = row.GetComponentType()
	out.CalculationMethod = row.GetCalculationMethod()
	out.TaxTreatmentCode = row.TaxTreatmentCode
	out.JurisdictionCode = row.GetJurisdictionCode()
	out.CurrencyCode = row.CurrencyCode
	out.RuleSetID = row.RuleSetId
	out.FormulaReferenceID = row.FormulaReferenceId
	out.Taxable = row.Taxable
	out.Pensionable = row.Pensionable
	out.EffectiveFrom = row.GetEffectiveFrom()
	out.EffectiveTo = row.EffectiveTo
	out.IsActive = row.GetIsActive()
	return out, nil
}

// PayComponent membaca detail pay-component berdasarkan ID.
func (a *Adapter) PayComponent(ctx context.Context, rawID string) (*model.PayComponent, error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_pay_components", "human_resource.pay_components.read")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "pay-component", "pay-component")
	if err != nil {
		return nil, err
	}
	if a.payroll == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.payroll.GetPayComponentByID(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapReadError(err, "pay-component")
	}
	return toModelPayComponent(resp.GetItem())
}

// PayComponents mendaftar dan memfilter pay-component dengan paginasi.
func (a *Adapter) PayComponents(ctx context.Context, limit, offset int, f_component_code *string, f_component_type *string, f_tax_treatment_code *string, f_jurisdiction_code *string, f_currency_code *string, f_rule_set_id *string, f_formula_reference_id *string) (*Page[model.PayComponent], error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_pay_components", "human_resource.pay_components.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid pay-component pagination", nil)
	}
	if a.payroll == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	idx, err := pickSingleFilter(f_component_code != nil, f_component_type != nil, f_tax_treatment_code != nil, f_jurisdiction_code != nil, f_currency_code != nil, f_rule_set_id != nil, f_formula_reference_id != nil)
	if err != nil {
		return nil, err
	}
	var resp *payroll.PayComponentsResponse
	switch idx {
	case 0:
		v, fErr := validateFilterString(f_component_code, "component_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.payroll.FindPayComponentByComponentCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 1:
		v, fErr := validateFilterString(f_component_type, "component_type")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.payroll.FindPayComponentByComponentType(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 2:
		v, fErr := validateFilterString(f_tax_treatment_code, "tax_treatment_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.payroll.FindPayComponentByTaxTreatmentCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 3:
		v, fErr := validateFilterString(f_jurisdiction_code, "jurisdiction_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.payroll.FindPayComponentByJurisdictionCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 4:
		v, fErr := validateFilterString(f_currency_code, "currency_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.payroll.FindPayComponentByCurrencyCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 5:
		v, fErr := validateFilterString(f_rule_set_id, "rule_set_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.payroll.FindPayComponentByRuleSetID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 6:
		v, fErr := validateFilterString(f_formula_reference_id, "formula_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.payroll.FindPayComponentByFormulaReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	default:
		resp, err = a.payroll.GetAllPayComponents(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "pay-component")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "pay-component", toModelPayComponent)
}

// CreatePayComponent membuat pay-component baru melalui HRM.
func (a *Adapter) CreatePayComponent(ctx context.Context, input model.CreatePayComponentInput) (*model.PayComponent, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_pay_components", "create", "human_resource.pay_components.create")
	if err != nil {
		return nil, err
	}
	if a.payroll == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &payroll.PayComponentInput{}
	patch.ComponentCode = input.ComponentCode
	patch.ComponentVersion = optionalInt32Ptr(input.ComponentVersion)
	patch.ComponentName = input.ComponentName
	patch.ComponentType = input.ComponentType
	patch.CalculationMethod = input.CalculationMethod
	patch.TaxTreatmentCode = input.TaxTreatmentCode
	patch.JurisdictionCode = input.JurisdictionCode
	patch.CurrencyCode = input.CurrencyCode
	patch.RuleSetId = input.RuleSetID
	patch.FormulaReferenceId = input.FormulaReferenceID
	patch.Taxable = input.Taxable
	patch.Pensionable = input.Pensionable
	patch.EffectiveFrom = input.EffectiveFrom
	patch.EffectiveTo = input.EffectiveTo
	patch.IsActive = input.IsActive
	resp, err := a.payroll.CreatePayComponent(rpcCtx, &payroll.CreatePayComponentRequest{Input: patch})
	if err != nil {
		return nil, mapHRMWriteError(err, "pay-component")
	}
	return toModelPayComponent(resp.GetItem())
}

// UpdatePayComponent memperbarui pay-component berdasarkan FieldMask.
func (a *Adapter) UpdatePayComponent(ctx context.Context, rawID string, input model.UpdatePayComponentInput) (*model.PayComponent, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_pay_components", "update", "human_resource.pay_components.update")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "pay-component", "pay-component")
	if err != nil {
		return nil, err
	}
	if a.payroll == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &payroll.PayComponentInput{}
	mask := &fieldmaskpb.FieldMask{}
	if err := applyStringMask(mask, "component_code", input.ComponentCode, &patch.ComponentCode, true); err != nil {
		return nil, err
	}
	if err := applyInt32Mask(mask, "component_version", input.ComponentVersion, &patch.ComponentVersion, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "component_name", input.ComponentName, &patch.ComponentName, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "component_type", input.ComponentType, &patch.ComponentType, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "calculation_method", input.CalculationMethod, &patch.CalculationMethod, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "tax_treatment_code", input.TaxTreatmentCode, &patch.TaxTreatmentCode, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "jurisdiction_code", input.JurisdictionCode, &patch.JurisdictionCode, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "currency_code", input.CurrencyCode, &patch.CurrencyCode, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "rule_set_id", input.RuleSetID, &patch.RuleSetId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "formula_reference_id", input.FormulaReferenceID, &patch.FormulaReferenceId, true); err != nil {
		return nil, err
	}
	if err := applyBoolMask(mask, "taxable", input.Taxable, &patch.Taxable, true); err != nil {
		return nil, err
	}
	if err := applyBoolMask(mask, "pensionable", input.Pensionable, &patch.Pensionable, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "effective_from", input.EffectiveFrom, &patch.EffectiveFrom, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "effective_to", input.EffectiveTo, &patch.EffectiveTo, true); err != nil {
		return nil, err
	}
	if err := applyBoolMask(mask, "is_active", input.IsActive, &patch.IsActive, true); err != nil {
		return nil, err
	}
	if len(mask.Paths) == 0 {
		return nil, middleware.ErrInvalidArgument("pay-component update is empty", nil)
	}
	resp, err := a.payroll.UpdatePayComponent(rpcCtx, &payroll.UpdatePayComponentRequest{Id: id, Input: patch, UpdateMask: mask})
	if err != nil {
		return nil, mapHRMWriteError(err, "pay-component")
	}
	return toModelPayComponent(resp.GetItem())
}

// DeletePayComponent menghapus pay-component berdasarkan ID.
func (a *Adapter) DeletePayComponent(ctx context.Context, rawID string) (*bool, error) {
	if a == nil || a.payroll == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	return deleteByUUID(ctx, a, rawID, "hr_pay_components", "human_resource.pay_components.delete", "pay-component", func(c context.Context, req *common.IDRequest) (*common.EmptyResponse, error) {
		return a.payroll.DeletePayComponent(c, req)
	})
}

func toModelPayrollAdjustment(row *payroll.PayrollAdjustment) (*model.PayrollAdjustment, error) {
	if row == nil {
		return nil, middleware.ErrUnavailable("invalid human resource payroll-adjustment response", nil)
	}
	if id, err := uuid.Parse(row.GetId()); err != nil || id == uuid.Nil {
		return nil, middleware.ErrUnavailable("invalid human resource payroll-adjustment ID", nil)
	}
	out := &model.PayrollAdjustment{ID: nodeID("payroll-adjustment", row.GetId()), RawID: row.GetId()}
	out.WorkerID = row.GetWorkerId()
	out.PayrollPeriodID = row.GetPayrollPeriodId()
	out.ComponentID = row.GetComponentId()
	out.AdjustmentType = row.GetAdjustmentType()
	out.Amount = row.GetAmount()
	out.CurrencyCode = row.GetCurrencyCode()
	out.Reason = row.GetReason()
	out.SourceReference = row.SourceReference
	out.RequestedByPrincipalID = row.GetRequestedByPrincipalId()
	out.ApprovedByPrincipalID = row.ApprovedByPrincipalId
	out.ApprovedAt = pbTimePtr(row.GetApprovedAt())
	out.IdempotencyKey = row.GetIdempotencyKey()
	out.CreatedAt = pbTimePtr(row.GetCreatedAt())
	return out, nil
}

// PayrollAdjustment membaca detail payroll-adjustment berdasarkan ID.
func (a *Adapter) PayrollAdjustment(ctx context.Context, rawID string) (*model.PayrollAdjustment, error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_payroll_adjustments", "human_resource.payroll_adjustments.read")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "payroll-adjustment", "payroll-adjustment")
	if err != nil {
		return nil, err
	}
	if a.payroll == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.payroll.GetPayrollAdjustmentByID(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapReadError(err, "payroll-adjustment")
	}
	return toModelPayrollAdjustment(resp.GetItem())
}

// PayrollAdjustments mendaftar dan memfilter payroll-adjustment dengan paginasi.
func (a *Adapter) PayrollAdjustments(ctx context.Context, limit, offset int, f_worker_id *string, f_payroll_period_id *string, f_component_id *string, f_adjustment_type *string, f_currency_code *string, f_requested_by_principal_id *string, f_approved_by_principal_id *string, f_idempotency_key *string) (*Page[model.PayrollAdjustment], error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_payroll_adjustments", "human_resource.payroll_adjustments.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid payroll-adjustment pagination", nil)
	}
	if a.payroll == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	idx, err := pickSingleFilter(f_worker_id != nil, f_payroll_period_id != nil, f_component_id != nil, f_adjustment_type != nil, f_currency_code != nil, f_requested_by_principal_id != nil, f_approved_by_principal_id != nil, f_idempotency_key != nil)
	if err != nil {
		return nil, err
	}
	var resp *payroll.PayrollAdjustmentsResponse
	switch idx {
	case 0:
		v, fErr := validateFilterString(f_worker_id, "worker_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.payroll.FindPayrollAdjustmentByWorkerID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 1:
		v, fErr := validateFilterString(f_payroll_period_id, "payroll_period_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.payroll.FindPayrollAdjustmentByPayrollPeriodID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 2:
		v, fErr := validateFilterString(f_component_id, "component_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.payroll.FindPayrollAdjustmentByComponentID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 3:
		v, fErr := validateFilterString(f_adjustment_type, "adjustment_type")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.payroll.FindPayrollAdjustmentByAdjustmentType(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 4:
		v, fErr := validateFilterString(f_currency_code, "currency_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.payroll.FindPayrollAdjustmentByCurrencyCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 5:
		v, fErr := validateFilterString(f_requested_by_principal_id, "requested_by_principal_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.payroll.FindPayrollAdjustmentByRequestedByPrincipalID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 6:
		v, fErr := validateFilterString(f_approved_by_principal_id, "approved_by_principal_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.payroll.FindPayrollAdjustmentByApprovedByPrincipalID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 7:
		v, fErr := validateFilterString(f_idempotency_key, "idempotency_key")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.payroll.FindPayrollAdjustmentByIdempotencyKey(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	default:
		resp, err = a.payroll.GetAllPayrollAdjustments(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "payroll-adjustment")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "payroll-adjustment", toModelPayrollAdjustment)
}

// CreatePayrollAdjustment membuat payroll-adjustment baru melalui HRM.
func (a *Adapter) CreatePayrollAdjustment(ctx context.Context, input model.CreatePayrollAdjustmentInput) (*model.PayrollAdjustment, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_payroll_adjustments", "create", "human_resource.payroll_adjustments.create")
	if err != nil {
		return nil, err
	}
	if a.payroll == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &payroll.PayrollAdjustmentInput{}
	patch.WorkerId = input.WorkerID
	patch.PayrollPeriodId = input.PayrollPeriodID
	patch.ComponentId = input.ComponentID
	patch.AdjustmentType = input.AdjustmentType
	patch.Amount = input.Amount
	patch.CurrencyCode = input.CurrencyCode
	patch.Reason = input.Reason
	patch.SourceReference = input.SourceReference
	patch.ApprovedByPrincipalId = input.ApprovedByPrincipalID
	patch.ApprovedAt = optionalTimestamp(input.ApprovedAt)
	patch.IdempotencyKey = input.IdempotencyKey
	resp, err := a.payroll.CreatePayrollAdjustment(rpcCtx, &payroll.CreatePayrollAdjustmentRequest{Input: patch})
	if err != nil {
		return nil, mapHRMWriteError(err, "payroll-adjustment")
	}
	return toModelPayrollAdjustment(resp.GetItem())
}

// UpdatePayrollAdjustment memperbarui payroll-adjustment berdasarkan FieldMask.
func (a *Adapter) UpdatePayrollAdjustment(ctx context.Context, rawID string, input model.UpdatePayrollAdjustmentInput) (*model.PayrollAdjustment, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_payroll_adjustments", "update", "human_resource.payroll_adjustments.update")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "payroll-adjustment", "payroll-adjustment")
	if err != nil {
		return nil, err
	}
	if a.payroll == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &payroll.PayrollAdjustmentInput{}
	mask := &fieldmaskpb.FieldMask{}
	if err := applyStringMask(mask, "worker_id", input.WorkerID, &patch.WorkerId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "payroll_period_id", input.PayrollPeriodID, &patch.PayrollPeriodId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "component_id", input.ComponentID, &patch.ComponentId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "adjustment_type", input.AdjustmentType, &patch.AdjustmentType, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "amount", input.Amount, &patch.Amount, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "currency_code", input.CurrencyCode, &patch.CurrencyCode, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "reason", input.Reason, &patch.Reason, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "source_reference", input.SourceReference, &patch.SourceReference, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "approved_by_principal_id", input.ApprovedByPrincipalID, &patch.ApprovedByPrincipalId, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "approved_at", input.ApprovedAt, &patch.ApprovedAt, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "idempotency_key", input.IdempotencyKey, &patch.IdempotencyKey, true); err != nil {
		return nil, err
	}
	if len(mask.Paths) == 0 {
		return nil, middleware.ErrInvalidArgument("payroll-adjustment update is empty", nil)
	}
	resp, err := a.payroll.UpdatePayrollAdjustment(rpcCtx, &payroll.UpdatePayrollAdjustmentRequest{Id: id, Input: patch, UpdateMask: mask})
	if err != nil {
		return nil, mapHRMWriteError(err, "payroll-adjustment")
	}
	return toModelPayrollAdjustment(resp.GetItem())
}

// DeletePayrollAdjustment menghapus payroll-adjustment berdasarkan ID.
func (a *Adapter) DeletePayrollAdjustment(ctx context.Context, rawID string) (*bool, error) {
	if a == nil || a.payroll == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	return deleteByUUID(ctx, a, rawID, "hr_payroll_adjustments", "human_resource.payroll_adjustments.delete", "payroll-adjustment", func(c context.Context, req *common.IDRequest) (*common.EmptyResponse, error) {
		return a.payroll.DeletePayrollAdjustment(c, req)
	})
}
