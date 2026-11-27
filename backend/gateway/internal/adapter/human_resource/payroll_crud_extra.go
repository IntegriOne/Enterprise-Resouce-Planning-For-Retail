package human_resource

import (
	"context"

	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	payroll "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/payroll/v1"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// CreatePayrollPeriod membuat periode penggajian baru melalui HRM.
func (a *Adapter) CreatePayrollPeriod(ctx context.Context, input model.CreatePayrollPeriodInput) (*model.PayrollPeriod, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_payroll_periods", "create", "human_resource.payroll_periods.create")
	if err != nil {
		return nil, err
	}
	if a.payroll == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &payroll.PayrollPeriodInput{
		PayrollGroupId:                  input.PayrollGroupID,
		EmployingLegalEntityReferenceId: input.EmployingLegalEntityReferenceID,
		CurrencyCode:                    input.CurrencyCode,
		JurisdictionCode:                input.JurisdictionCode,
		PeriodCode:                      input.PeriodCode,
		PeriodStart:                     input.PeriodStart,
		PeriodEnd:                       input.PeriodEnd,
		PayDate:                         input.PayDate,
		PeriodStatus:                    input.PeriodStatus,
		CalendarReferenceId:             input.CalendarReferenceID,
	}
	resp, err := a.payroll.CreatePayrollPeriod(rpcCtx, &payroll.CreatePayrollPeriodRequest{Input: patch})
	if err != nil {
		return nil, mapHRMWriteError(err, "payroll-period")
	}
	return toModelPayrollPeriod(resp.GetItem())
}

// UpdatePayrollPeriod memperbarui periode penggajian berdasarkan FieldMask.
func (a *Adapter) UpdatePayrollPeriod(ctx context.Context, rawID string, input model.UpdatePayrollPeriodInput) (*model.PayrollPeriod, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_payroll_periods", "update", "human_resource.payroll_periods.update")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "payroll-period", "payroll-period")
	if err != nil {
		return nil, err
	}
	if a.payroll == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &payroll.PayrollPeriodInput{}
	mask := &fieldmaskpb.FieldMask{}
	if err := applyStringMask(mask, "payroll_group_id", input.PayrollGroupID, &patch.PayrollGroupId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "employing_legal_entity_reference_id", input.EmployingLegalEntityReferenceID, &patch.EmployingLegalEntityReferenceId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "currency_code", input.CurrencyCode, &patch.CurrencyCode, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "jurisdiction_code", input.JurisdictionCode, &patch.JurisdictionCode, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "period_code", input.PeriodCode, &patch.PeriodCode, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "period_start", input.PeriodStart, &patch.PeriodStart, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "period_end", input.PeriodEnd, &patch.PeriodEnd, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "pay_date", input.PayDate, &patch.PayDate, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "period_status", input.PeriodStatus, &patch.PeriodStatus, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "calendar_reference_id", input.CalendarReferenceID, &patch.CalendarReferenceId, true); err != nil {
		return nil, err
	}
	if len(mask.Paths) == 0 {
		return nil, middleware.ErrInvalidArgument("payroll-period update is empty", nil)
	}
	resp, err := a.payroll.UpdatePayrollPeriod(rpcCtx, &payroll.UpdatePayrollPeriodRequest{
		Id:         id,
		Input:      patch,
		UpdateMask: mask,
	})
	if err != nil {
		return nil, mapHRMWriteError(err, "payroll-period")
	}
	return toModelPayrollPeriod(resp.GetItem())
}

// DeletePayrollPeriod menghapus periode penggajian berdasarkan ID.
func (a *Adapter) DeletePayrollPeriod(ctx context.Context, rawID string) (*bool, error) {
	if a == nil || a.payroll == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	return deleteByUUID(ctx, a, rawID, "hr_payroll_periods", "human_resource.payroll_periods.delete", "payroll-period", func(c context.Context, req *common.IDRequest) (*common.EmptyResponse, error) {
		return a.payroll.DeletePayrollPeriod(c, req)
	})
}

// CreatePayrollGroup membuat kelompok penggajian baru melalui HRM.
func (a *Adapter) CreatePayrollGroup(ctx context.Context, input model.CreatePayrollGroupInput) (*model.PayrollGroup, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_payroll_groups", "create", "human_resource.payroll_groups.create")
	if err != nil {
		return nil, err
	}
	if a.payroll == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &payroll.PayrollGroupInput{
		GroupCode:                       input.GroupCode,
		GroupName:                       input.GroupName,
		EmployingLegalEntityReferenceId: input.EmployingLegalEntityReferenceID,
		JurisdictionCode:                input.JurisdictionCode,
		CurrencyCode:                    input.CurrencyCode,
		CurrencyScale:                   optionalInt32Ptr(input.CurrencyScale),
		PayFrequency:                    input.PayFrequency,
		Timezone:                        input.Timezone,
		IsActive:                        input.IsActive,
	}
	resp, err := a.payroll.CreatePayrollGroup(rpcCtx, &payroll.CreatePayrollGroupRequest{Input: patch})
	if err != nil {
		return nil, mapHRMWriteError(err, "payroll-group")
	}
	return toModelPayrollGroup(resp.GetItem())
}

// UpdatePayrollGroup memperbarui kelompok penggajian berdasarkan FieldMask.
func (a *Adapter) UpdatePayrollGroup(ctx context.Context, rawID string, input model.UpdatePayrollGroupInput) (*model.PayrollGroup, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_payroll_groups", "update", "human_resource.payroll_groups.update")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "payroll-group", "payroll-group")
	if err != nil {
		return nil, err
	}
	if a.payroll == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &payroll.PayrollGroupInput{}
	mask := &fieldmaskpb.FieldMask{}
	if err := applyStringMask(mask, "group_code", input.GroupCode, &patch.GroupCode, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "group_name", input.GroupName, &patch.GroupName, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "employing_legal_entity_reference_id", input.EmployingLegalEntityReferenceID, &patch.EmployingLegalEntityReferenceId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "jurisdiction_code", input.JurisdictionCode, &patch.JurisdictionCode, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "currency_code", input.CurrencyCode, &patch.CurrencyCode, true); err != nil {
		return nil, err
	}
	if err := applyInt32Mask(mask, "currency_scale", input.CurrencyScale, &patch.CurrencyScale, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "pay_frequency", input.PayFrequency, &patch.PayFrequency, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "timezone", input.Timezone, &patch.Timezone, true); err != nil {
		return nil, err
	}
	if err := applyBoolMask(mask, "is_active", input.IsActive, &patch.IsActive, true); err != nil {
		return nil, err
	}
	if len(mask.Paths) == 0 {
		return nil, middleware.ErrInvalidArgument("payroll-group update is empty", nil)
	}
	resp, err := a.payroll.UpdatePayrollGroup(rpcCtx, &payroll.UpdatePayrollGroupRequest{
		Id:         id,
		Input:      patch,
		UpdateMask: mask,
	})
	if err != nil {
		return nil, mapHRMWriteError(err, "payroll-group")
	}
	return toModelPayrollGroup(resp.GetItem())
}

// DeletePayrollGroup menghapus kelompok penggajian berdasarkan ID.
func (a *Adapter) DeletePayrollGroup(ctx context.Context, rawID string) (*bool, error) {
	if a == nil || a.payroll == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	return deleteByUUID(ctx, a, rawID, "hr_payroll_groups", "human_resource.payroll_groups.delete", "payroll-group", func(c context.Context, req *common.IDRequest) (*common.EmptyResponse, error) {
		return a.payroll.DeletePayrollGroup(c, req)
	})
}

// CreatePayrollRuleSet membuat set aturan penggajian baru melalui HRM.
func (a *Adapter) CreatePayrollRuleSet(ctx context.Context, input model.CreatePayrollRuleSetInput) (*model.PayrollRuleSet, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_payroll_rule_sets", "create", "human_resource.payroll_rule_sets.create")
	if err != nil {
		return nil, err
	}
	if a.payroll == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	digest, err := optionalBase64Bytes(input.ArtifactDigest, "artifact_digest")
	if err != nil {
		return nil, err
	}
	patch := &payroll.PayrollRuleSetInput{
		RuleCode:                             input.RuleCode,
		RuleVersion:                          input.RuleVersion,
		JurisdictionCode:                     input.JurisdictionCode,
		EffectiveFrom:                        input.EffectiveFrom,
		EffectiveTo:                          input.EffectiveTo,
		ArtifactDocumentVersionReferenceId:   input.ArtifactDocumentVersionReferenceID,
		ArtifactDigest:                       digest,
		LegalBasisDocumentVersionReferenceId: input.LegalBasisDocumentVersionReferenceID,
		ApprovedByPrincipalId:                input.ApprovedByPrincipalID,
		ApprovedAt:                           optionalTimestamp(input.ApprovedAt),
	}
	resp, err := a.payroll.CreatePayrollRuleSet(rpcCtx, &payroll.CreatePayrollRuleSetRequest{Input: patch})
	if err != nil {
		return nil, mapHRMWriteError(err, "payroll-rule-set")
	}
	return toModelPayrollRuleSet(resp.GetItem())
}

// UpdatePayrollRuleSet memperbarui set aturan penggajian berdasarkan FieldMask.
func (a *Adapter) UpdatePayrollRuleSet(ctx context.Context, rawID string, input model.UpdatePayrollRuleSetInput) (*model.PayrollRuleSet, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_payroll_rule_sets", "update", "human_resource.payroll_rule_sets.update")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "payroll-rule-set", "payroll-rule-set")
	if err != nil {
		return nil, err
	}
	if a.payroll == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &payroll.PayrollRuleSetInput{}
	mask := &fieldmaskpb.FieldMask{}
	if err := applyStringMask(mask, "rule_code", input.RuleCode, &patch.RuleCode, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "rule_version", input.RuleVersion, &patch.RuleVersion, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "jurisdiction_code", input.JurisdictionCode, &patch.JurisdictionCode, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "effective_from", input.EffectiveFrom, &patch.EffectiveFrom, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "effective_to", input.EffectiveTo, &patch.EffectiveTo, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "artifact_document_version_reference_id", input.ArtifactDocumentVersionReferenceID, &patch.ArtifactDocumentVersionReferenceId, true); err != nil {
		return nil, err
	}
	if err := applyBytesMask(mask, "artifact_digest", input.ArtifactDigest, &patch.ArtifactDigest, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "legal_basis_document_version_reference_id", input.LegalBasisDocumentVersionReferenceID, &patch.LegalBasisDocumentVersionReferenceId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "approved_by_principal_id", input.ApprovedByPrincipalID, &patch.ApprovedByPrincipalId, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "approved_at", input.ApprovedAt, &patch.ApprovedAt, true); err != nil {
		return nil, err
	}
	if len(mask.Paths) == 0 {
		return nil, middleware.ErrInvalidArgument("payroll-rule-set update is empty", nil)
	}
	resp, err := a.payroll.UpdatePayrollRuleSet(rpcCtx, &payroll.UpdatePayrollRuleSetRequest{
		Id:         id,
		Input:      patch,
		UpdateMask: mask,
	})
	if err != nil {
		return nil, mapHRMWriteError(err, "payroll-rule-set")
	}
	return toModelPayrollRuleSet(resp.GetItem())
}

// DeletePayrollRuleSet menghapus set aturan penggajian berdasarkan ID.
func (a *Adapter) DeletePayrollRuleSet(ctx context.Context, rawID string) (*bool, error) {
	if a == nil || a.payroll == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	return deleteByUUID(ctx, a, rawID, "hr_payroll_rule_sets", "human_resource.payroll_rule_sets.delete", "payroll-rule-set", func(c context.Context, req *common.IDRequest) (*common.EmptyResponse, error) {
		return a.payroll.DeletePayrollRuleSet(c, req)
	})
}

// UpdatePayrollRun memperbarui proses penggajian berdasarkan FieldMask.
func (a *Adapter) UpdatePayrollRun(ctx context.Context, rawID string, input model.UpdatePayrollRunInput) (*model.PayrollRun, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_payroll_runs", "update", "human_resource.payroll_runs.update")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "payroll-run", "payroll-run")
	if err != nil {
		return nil, err
	}
	if a.payroll == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &payroll.PayrollRunInput{}
	mask := &fieldmaskpb.FieldMask{}
	if err := applyStringMask(mask, "payroll_period_id", input.PayrollPeriodID, &patch.PayrollPeriodId, true); err != nil {
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
	if err := applyStringMask(mask, "rule_set_id", input.RuleSetID, &patch.RuleSetId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "original_run_id", input.OriginalRunID, &patch.OriginalRunId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "settlement_direction", input.SettlementDirection, &patch.SettlementDirection, true); err != nil {
		return nil, err
	}
	if err := applyInt32Mask(mask, "run_number", input.RunNumber, &patch.RunNumber, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "run_type", input.RunType, &patch.RunType, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "status", input.Status, &patch.Status, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "jurisdiction_code", input.JurisdictionCode, &patch.JurisdictionCode, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "rule_engine_name", input.RuleEngineName, &patch.RuleEngineName, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "rule_engine_version", input.RuleEngineVersion, &patch.RuleEngineVersion, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "input_snapshot_reference_id", input.InputSnapshotReferenceID, &patch.InputSnapshotReferenceId, true); err != nil {
		return nil, err
	}
	if err := applyBytesMask(mask, "input_digest", input.InputDigest, &patch.InputDigest, true); err != nil {
		return nil, err
	}
	if err := applyBytesMask(mask, "output_digest", input.OutputDigest, &patch.OutputDigest, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "calculated_at", input.CalculatedAt, &patch.CalculatedAt, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "approved_by_principal_id", input.ApprovedByPrincipalID, &patch.ApprovedByPrincipalId, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "approved_at", input.ApprovedAt, &patch.ApprovedAt, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "finance_journal_reference_id", input.FinanceJournalReferenceID, &patch.FinanceJournalReferenceId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "payment_batch_reference_id", input.PaymentBatchReferenceID, &patch.PaymentBatchReferenceId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "finance_settlement_event_reference_id", input.FinanceSettlementEventReferenceID, &patch.FinanceSettlementEventReferenceId, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "finance_settled_at", input.FinanceSettledAt, &patch.FinanceSettledAt, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "idempotency_key", input.IdempotencyKey, &patch.IdempotencyKey, true); err != nil {
		return nil, err
	}
	if len(mask.Paths) == 0 {
		return nil, middleware.ErrInvalidArgument("payroll-run update is empty", nil)
	}
	resp, err := a.payroll.UpdatePayrollRun(rpcCtx, &payroll.UpdatePayrollRunRequest{
		Id:         id,
		Input:      patch,
		UpdateMask: mask,
	})
	if err != nil {
		return nil, mapHRMWriteError(err, "payroll-run")
	}
	return toModelPayrollRun(resp.GetItem())
}

// DeletePayrollRun menghapus proses penggajian berdasarkan ID.
func (a *Adapter) DeletePayrollRun(ctx context.Context, rawID string) (*bool, error) {
	if a == nil || a.payroll == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	return deleteByUUID(ctx, a, rawID, "hr_payroll_runs", "human_resource.payroll_runs.delete", "payroll-run", func(c context.Context, req *common.IDRequest) (*common.EmptyResponse, error) {
		return a.payroll.DeletePayrollRun(c, req)
	})
}
