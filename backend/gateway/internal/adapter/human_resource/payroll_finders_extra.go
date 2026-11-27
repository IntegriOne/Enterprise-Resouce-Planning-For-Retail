package human_resource

import (
	"context"

	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	payroll "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/payroll/v1"
)

// PayrollRunsFiltered mendaftar proses penggajian dengan filter konsolidasi.
func (a *Adapter) PayrollRunsFiltered(ctx context.Context, limit, offset int, idempotencyKey, payrollPeriodID, payrollGroupID, employingLegalEntityRefID, currencyCode, ruleSetID, originalRunID *string, runNumber *int, runType, status, jurisdictionCode, inputSnapshotRefID, startedByPrincipalID, approvedByPrincipalID, financeJournalRefID, paymentBatchRefID, financeSettlementEventRefID *string) (*Page[model.PayrollRun], error) {
	if payrollGroupID == nil && employingLegalEntityRefID == nil && currencyCode == nil && ruleSetID == nil && originalRunID == nil && runNumber == nil && runType == nil && status == nil && jurisdictionCode == nil && inputSnapshotRefID == nil && startedByPrincipalID == nil && approvedByPrincipalID == nil && financeJournalRefID == nil && paymentBatchRefID == nil && financeSettlementEventRefID == nil {
		return a.PayrollRuns(ctx, limit, offset, idempotencyKey, payrollPeriodID)
	}
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_payroll_runs", "human_resource.payroll_runs.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid payroll-run pagination", nil)
	}
	if a.payroll == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	idx, err := pickSingleFilter(
		idempotencyKey != nil,
		payrollPeriodID != nil,
		payrollGroupID != nil,
		employingLegalEntityRefID != nil,
		currencyCode != nil,
		ruleSetID != nil,
		originalRunID != nil,
		runNumber != nil,
		runType != nil,
		status != nil,
		jurisdictionCode != nil,
		inputSnapshotRefID != nil,
		startedByPrincipalID != nil,
		approvedByPrincipalID != nil,
		financeJournalRefID != nil,
		paymentBatchRefID != nil,
		financeSettlementEventRefID != nil,
	)
	if err != nil {
		return nil, err
	}
	var resp *payroll.PayrollRunsResponse
	switch idx {
	case 0:
		v, fErr := validateFilterString(idempotencyKey, "idempotency_key")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.payroll.FindPayrollRunByIdempotencyKey(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 1:
		v, fErr := validateUUIDFilter(payrollPeriodID, "payroll-period", "payroll_period_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.payroll.FindPayrollRunByPayrollPeriodID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 2:
		v, fErr := validateUUIDFilter(payrollGroupID, "payroll-group", "payroll_group_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.payroll.FindPayrollRunByPayrollGroupID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 3:
		v, fErr := validateUUIDFilter(employingLegalEntityRefID, "legal-entity", "employing_legal_entity_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.payroll.FindPayrollRunByEmployingLegalEntityReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 4:
		v, fErr := validateFilterString(currencyCode, "currency_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.payroll.FindPayrollRunByCurrencyCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 5:
		v, fErr := validateUUIDFilter(ruleSetID, "payroll-rule-set", "rule_set_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.payroll.FindPayrollRunByRuleSetID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 6:
		v, fErr := validateUUIDFilter(originalRunID, "payroll-run", "original_run_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.payroll.FindPayrollRunByOriginalRunID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 7:
		v := int32(*runNumber)
		resp, err = a.payroll.FindPayrollRunByRunNumber(rpcCtx, &common.Int32FinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 8:
		v, fErr := validateFilterString(runType, "run_type")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.payroll.FindPayrollRunByRunType(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 9:
		v, fErr := validateFilterString(status, "status")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.payroll.FindPayrollRunByStatus(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 10:
		v, fErr := validateFilterString(jurisdictionCode, "jurisdiction_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.payroll.FindPayrollRunByJurisdictionCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 11:
		v, fErr := validateUUIDFilter(inputSnapshotRefID, "input-snapshot", "input_snapshot_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.payroll.FindPayrollRunByInputSnapshotReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 12:
		v, fErr := validateUUIDFilter(startedByPrincipalID, "principal", "started_by_principal_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.payroll.FindPayrollRunByStartedByPrincipalID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 13:
		v, fErr := validateUUIDFilter(approvedByPrincipalID, "principal", "approved_by_principal_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.payroll.FindPayrollRunByApprovedByPrincipalID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 14:
		v, fErr := validateUUIDFilter(financeJournalRefID, "finance-journal", "finance_journal_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.payroll.FindPayrollRunByFinanceJournalReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 15:
		v, fErr := validateUUIDFilter(paymentBatchRefID, "payment-batch", "payment_batch_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.payroll.FindPayrollRunByPaymentBatchReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 16:
		v, fErr := validateUUIDFilter(financeSettlementEventRefID, "finance-settlement-event", "finance_settlement_event_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.payroll.FindPayrollRunByFinanceSettlementEventReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	default:
		resp, err = a.payroll.GetAllPayrollRuns(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "payroll-run")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "payroll-run", toModelPayrollRun)
}

// PayrollPeriodsFiltered mendaftar periode penggajian dengan filter konsolidasi.
func (a *Adapter) PayrollPeriodsFiltered(ctx context.Context, limit, offset int, payrollGroupID, employingLegalEntityRefID, currencyCode, jurisdictionCode, periodCode, periodStatus, calendarRefID *string) (*Page[model.PayrollPeriod], error) {
	if employingLegalEntityRefID == nil && currencyCode == nil && jurisdictionCode == nil && periodCode == nil && periodStatus == nil && calendarRefID == nil {
		return a.PayrollPeriods(ctx, limit, offset, payrollGroupID)
	}
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_payroll_periods", "human_resource.payroll_periods.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid payroll-period pagination", nil)
	}
	if a.payroll == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	idx, err := pickSingleFilter(
		payrollGroupID != nil,
		employingLegalEntityRefID != nil,
		currencyCode != nil,
		jurisdictionCode != nil,
		periodCode != nil,
		periodStatus != nil,
		calendarRefID != nil,
	)
	if err != nil {
		return nil, err
	}
	var resp *payroll.PayrollPeriodsResponse
	switch idx {
	case 0:
		v, fErr := validateUUIDFilter(payrollGroupID, "payroll-group", "payroll_group_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.payroll.FindPayrollPeriodByPayrollGroupID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 1:
		v, fErr := validateUUIDFilter(employingLegalEntityRefID, "legal-entity", "employing_legal_entity_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.payroll.FindPayrollPeriodByEmployingLegalEntityReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 2:
		v, fErr := validateFilterString(currencyCode, "currency_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.payroll.FindPayrollPeriodByCurrencyCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 3:
		v, fErr := validateFilterString(jurisdictionCode, "jurisdiction_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.payroll.FindPayrollPeriodByJurisdictionCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 4:
		v, fErr := validateFilterString(periodCode, "period_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.payroll.FindPayrollPeriodByPeriodCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 5:
		v, fErr := validateFilterString(periodStatus, "period_status")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.payroll.FindPayrollPeriodByPeriodStatus(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 6:
		v, fErr := validateUUIDFilter(calendarRefID, "calendar", "calendar_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.payroll.FindPayrollPeriodByCalendarReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	default:
		resp, err = a.payroll.GetAllPayrollPeriods(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "payroll-period")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "payroll-period", toModelPayrollPeriod)
}

// PayrollGroupsFiltered mendaftar kelompok penggajian dengan filter konsolidasi.
func (a *Adapter) PayrollGroupsFiltered(ctx context.Context, limit, offset int, groupCode, employingLegalEntityRefID, jurisdictionCode, currencyCode *string) (*Page[model.PayrollGroup], error) {
	if groupCode == nil && employingLegalEntityRefID == nil && jurisdictionCode == nil && currencyCode == nil {
		return a.PayrollGroups(ctx, limit, offset)
	}
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_payroll_groups", "human_resource.payroll_groups.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid payroll-group pagination", nil)
	}
	if a.payroll == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	idx, err := pickSingleFilter(groupCode != nil, employingLegalEntityRefID != nil, jurisdictionCode != nil, currencyCode != nil)
	if err != nil {
		return nil, err
	}
	var resp *payroll.PayrollGroupsResponse
	switch idx {
	case 0:
		v, fErr := validateFilterString(groupCode, "group_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.payroll.FindPayrollGroupByGroupCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 1:
		v, fErr := validateUUIDFilter(employingLegalEntityRefID, "legal-entity", "employing_legal_entity_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.payroll.FindPayrollGroupByEmployingLegalEntityReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 2:
		v, fErr := validateFilterString(jurisdictionCode, "jurisdiction_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.payroll.FindPayrollGroupByJurisdictionCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 3:
		v, fErr := validateFilterString(currencyCode, "currency_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.payroll.FindPayrollGroupByCurrencyCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	default:
		resp, err = a.payroll.GetAllPayrollGroups(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "payroll-group")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "payroll-group", toModelPayrollGroup)
}

// PayrollRuleSetsFiltered mendaftar set aturan penggajian dengan filter konsolidasi.
func (a *Adapter) PayrollRuleSetsFiltered(ctx context.Context, limit, offset int, ruleCode, jurisdictionCode, artifactDocVerRefID, legalBasisDocVerRefID, approvedByPrincipalID *string) (*Page[model.PayrollRuleSet], error) {
	if ruleCode == nil && jurisdictionCode == nil && artifactDocVerRefID == nil && legalBasisDocVerRefID == nil && approvedByPrincipalID == nil {
		return a.PayrollRuleSets(ctx, limit, offset)
	}
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_payroll_rule_sets", "human_resource.payroll_rule_sets.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid payroll-rule-set pagination", nil)
	}
	if a.payroll == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	idx, err := pickSingleFilter(ruleCode != nil, jurisdictionCode != nil, artifactDocVerRefID != nil, legalBasisDocVerRefID != nil, approvedByPrincipalID != nil)
	if err != nil {
		return nil, err
	}
	var resp *payroll.PayrollRuleSetsResponse
	switch idx {
	case 0:
		v, fErr := validateFilterString(ruleCode, "rule_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.payroll.FindPayrollRuleSetByRuleCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 1:
		v, fErr := validateFilterString(jurisdictionCode, "jurisdiction_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.payroll.FindPayrollRuleSetByJurisdictionCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 2:
		v, fErr := validateUUIDFilter(artifactDocVerRefID, "document-version", "artifact_document_version_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.payroll.FindPayrollRuleSetByArtifactDocumentVersionReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 3:
		v, fErr := validateUUIDFilter(legalBasisDocVerRefID, "document-version", "legal_basis_document_version_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.payroll.FindPayrollRuleSetByLegalBasisDocumentVersionReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 4:
		v, fErr := validateUUIDFilter(approvedByPrincipalID, "principal", "approved_by_principal_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.payroll.FindPayrollRuleSetByApprovedByPrincipalID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	default:
		resp, err = a.payroll.GetAllPayrollRuleSets(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "payroll-rule-set")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "payroll-rule-set", toModelPayrollRuleSet)
}

// PayrollResultsFiltered mendaftar hasil penggajian dengan filter konsolidasi.
func (a *Adapter) PayrollResultsFiltered(ctx context.Context, limit, offset int, payrollRunID, workerID, payrollGroupID, payrollProfileRefID, currencyCode, resultStatus, paymentRefID, payslipDocRefID *string) (*Page[model.PayrollResult], error) {
	if payrollRunID != nil && workerID == nil && payrollGroupID == nil && payrollProfileRefID == nil && currencyCode == nil && resultStatus == nil && paymentRefID == nil && payslipDocRefID == nil {
		return a.PayrollResults(ctx, *payrollRunID, limit, offset)
	}
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_payroll_results", "human_resource.payroll_results.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid payroll-result pagination", nil)
	}
	if a.payroll == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	idx, err := pickSingleFilter(
		payrollRunID != nil,
		workerID != nil,
		payrollGroupID != nil,
		payrollProfileRefID != nil,
		currencyCode != nil,
		resultStatus != nil,
		paymentRefID != nil,
		payslipDocRefID != nil,
	)
	if err != nil {
		return nil, err
	}
	var resp *payroll.PayrollResultsResponse
	switch idx {
	case 0:
		v, fErr := validateUUIDFilter(payrollRunID, "payroll-run", "payroll_run_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.payroll.FindPayrollResultByPayrollRunID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 1:
		v, fErr := validateUUIDFilter(workerID, "worker", "worker_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.payroll.FindPayrollResultByWorkerID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 2:
		v, fErr := validateUUIDFilter(payrollGroupID, "payroll-group", "payroll_group_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.payroll.FindPayrollResultByPayrollGroupID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 3:
		v, fErr := validateUUIDFilter(payrollProfileRefID, "payroll-profile", "payroll_profile_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.payroll.FindPayrollResultByPayrollProfileReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 4:
		v, fErr := validateFilterString(currencyCode, "currency_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.payroll.FindPayrollResultByCurrencyCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 5:
		v, fErr := validateFilterString(resultStatus, "result_status")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.payroll.FindPayrollResultByResultStatus(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 6:
		v, fErr := validateUUIDFilter(paymentRefID, "payment", "payment_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.payroll.FindPayrollResultByPaymentReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 7:
		v, fErr := validateUUIDFilter(payslipDocRefID, "document", "payslip_document_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.payroll.FindPayrollResultByPayslipDocumentReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	default:
		resp, err = a.payroll.GetAllPayrollResults(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "payroll-result")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "payroll-result", toModelPayrollResult)
}

// PayrollResultLinesFiltered mendaftar baris hasil penggajian dengan filter konsolidasi.
func (a *Adapter) PayrollResultLinesFiltered(ctx context.Context, limit, offset int, payrollResultID, componentID *string) (*Page[model.PayrollResultLine], error) {
	if payrollResultID != nil && componentID == nil {
		return a.PayrollResultLines(ctx, *payrollResultID, limit, offset)
	}
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_payroll_result_lines", "human_resource.payroll_result_lines.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid payroll-result-line pagination", nil)
	}
	if a.payroll == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	idx, err := pickSingleFilter(payrollResultID != nil, componentID != nil)
	if err != nil {
		return nil, err
	}
	var resp *payroll.PayrollResultLinesResponse
	switch idx {
	case 0:
		v, fErr := validateUUIDFilter(payrollResultID, "payroll-result", "payroll_result_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.payroll.FindPayrollResultLineByPayrollResultID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 1:
		v, fErr := validateUUIDFilter(componentID, "pay-component", "component_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.payroll.FindPayrollResultLineByComponentID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	default:
		resp, err = a.payroll.GetAllPayrollResultLines(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "payroll-result-line")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "payroll-result-line", toModelPayrollResultLine)
}
