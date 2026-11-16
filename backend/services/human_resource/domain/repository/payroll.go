// Sumber model adalah baseline SQL HRM; aturan bisnis tetap milik service.
package repository

import (
	"context"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
)

// PayrollProfileRepository adalah port persistensi hr_payroll_profiles.
type PayrollProfileRepository interface {
	Repository[entity.PayrollProfile, string]
	InsertPayrollProfiles(context.Context, *entity.PayrollProfile) error
	UpdatePayrollProfiles(context.Context, *entity.PayrollProfile) error
	DeletePayrollProfilesByUUID(context.Context, string) error
	FindPayrollProfilesByUUID(context.Context, string) (*entity.PayrollProfile, error)
	FindPayrollProfilesByUUIDForUpdate(context.Context, string) (*entity.PayrollProfile, error)
	GetAllPayrollProfiles(context.Context, int, int) ([]entity.PayrollProfile, error)
	CountPayrollProfiles(context.Context) (int64, error)
	SelectPayrollProfilesWhere(context.Context, []Filter, int, int) ([]entity.PayrollProfile, error)
	CountPayrollProfilesWhere(context.Context, []Filter) (int64, error)
	FindPayrollProfilesByWorkerID(context.Context, string, int, int) ([]entity.PayrollProfile, error)
	FindPayrollProfilesByPayrollCountryCode(context.Context, string, int, int) ([]entity.PayrollProfile, error)
	FindPayrollProfilesByPayrollJurisdictionCode(context.Context, string, int, int) ([]entity.PayrollProfile, error)
	FindPayrollProfilesByPayrollGroupID(context.Context, string, int, int) ([]entity.PayrollProfile, error)
	FindPayrollProfilesByEmployingLegalEntityReferenceID(context.Context, string, int, int) ([]entity.PayrollProfile, error)
	FindPayrollProfilesByCurrencyCode(context.Context, string, int, int) ([]entity.PayrollProfile, error)
	FindPayrollProfilesByTaxProfileReferenceID(context.Context, *string, int, int) ([]entity.PayrollProfile, error)
	FindPayrollProfilesByBankAccountTokenReferenceID(context.Context, *string, int, int) ([]entity.PayrollProfile, error)
	FindPayrollProfilesByTaxStatusCode(context.Context, *string, int, int) ([]entity.PayrollProfile, error)
	FindPayrollProfilesByWithholdingElectionReferenceID(context.Context, *string, int, int) ([]entity.PayrollProfile, error)
	FindByWorkerID(ctx context.Context, value string, limit, offset int) ([]entity.PayrollProfile, error)
	FindByPayrollCountryCode(ctx context.Context, value string, limit, offset int) ([]entity.PayrollProfile, error)
	FindByPayrollJurisdictionCode(ctx context.Context, value string, limit, offset int) ([]entity.PayrollProfile, error)
	FindByPayrollGroupID(ctx context.Context, value string, limit, offset int) ([]entity.PayrollProfile, error)
	FindByEmployingLegalEntityReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.PayrollProfile, error)
	FindByCurrencyCode(ctx context.Context, value string, limit, offset int) ([]entity.PayrollProfile, error)
	FindByTaxProfileReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.PayrollProfile, error)
	FindByBankAccountTokenReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.PayrollProfile, error)
	FindByTaxStatusCode(ctx context.Context, value *string, limit, offset int) ([]entity.PayrollProfile, error)
	FindByWithholdingElectionReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.PayrollProfile, error)
}

// PayComponentRepository adalah port persistensi hr_pay_components.
type PayComponentRepository interface {
	Repository[entity.PayComponent, string]
	InsertPayComponents(context.Context, *entity.PayComponent) error
	UpdatePayComponents(context.Context, *entity.PayComponent) error
	DeletePayComponentsByUUID(context.Context, string) error
	FindPayComponentsByUUID(context.Context, string) (*entity.PayComponent, error)
	FindPayComponentsByUUIDForUpdate(context.Context, string) (*entity.PayComponent, error)
	GetAllPayComponents(context.Context, int, int) ([]entity.PayComponent, error)
	CountPayComponents(context.Context) (int64, error)
	SelectPayComponentsWhere(context.Context, []Filter, int, int) ([]entity.PayComponent, error)
	CountPayComponentsWhere(context.Context, []Filter) (int64, error)
	FindPayComponentsByComponentCode(context.Context, string, int, int) ([]entity.PayComponent, error)
	FindPayComponentsByComponentType(context.Context, string, int, int) ([]entity.PayComponent, error)
	FindPayComponentsByTaxTreatmentCode(context.Context, *string, int, int) ([]entity.PayComponent, error)
	FindPayComponentsByJurisdictionCode(context.Context, string, int, int) ([]entity.PayComponent, error)
	FindPayComponentsByCurrencyCode(context.Context, *string, int, int) ([]entity.PayComponent, error)
	FindPayComponentsByRuleSetID(context.Context, *string, int, int) ([]entity.PayComponent, error)
	FindPayComponentsByFormulaReferenceID(context.Context, *string, int, int) ([]entity.PayComponent, error)
	FindByComponentCode(ctx context.Context, value string, limit, offset int) ([]entity.PayComponent, error)
	FindByComponentType(ctx context.Context, value string, limit, offset int) ([]entity.PayComponent, error)
	FindByTaxTreatmentCode(ctx context.Context, value *string, limit, offset int) ([]entity.PayComponent, error)
	FindByJurisdictionCode(ctx context.Context, value string, limit, offset int) ([]entity.PayComponent, error)
	FindByCurrencyCode(ctx context.Context, value *string, limit, offset int) ([]entity.PayComponent, error)
	FindByRuleSetID(ctx context.Context, value *string, limit, offset int) ([]entity.PayComponent, error)
	FindByFormulaReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.PayComponent, error)
}

// PayrollPeriodRepository adalah port persistensi hr_payroll_periods.
type PayrollPeriodRepository interface {
	Repository[entity.PayrollPeriod, string]
	InsertPayrollPeriods(context.Context, *entity.PayrollPeriod) error
	UpdatePayrollPeriods(context.Context, *entity.PayrollPeriod) error
	DeletePayrollPeriodsByUUID(context.Context, string) error
	FindPayrollPeriodsByUUID(context.Context, string) (*entity.PayrollPeriod, error)
	FindPayrollPeriodsByUUIDForUpdate(context.Context, string) (*entity.PayrollPeriod, error)
	GetAllPayrollPeriods(context.Context, int, int) ([]entity.PayrollPeriod, error)
	CountPayrollPeriods(context.Context) (int64, error)
	SelectPayrollPeriodsWhere(context.Context, []Filter, int, int) ([]entity.PayrollPeriod, error)
	CountPayrollPeriodsWhere(context.Context, []Filter) (int64, error)
	FindPayrollPeriodsByPayrollGroupID(context.Context, string, int, int) ([]entity.PayrollPeriod, error)
	FindPayrollPeriodsByEmployingLegalEntityReferenceID(context.Context, string, int, int) ([]entity.PayrollPeriod, error)
	FindPayrollPeriodsByCurrencyCode(context.Context, string, int, int) ([]entity.PayrollPeriod, error)
	FindPayrollPeriodsByJurisdictionCode(context.Context, string, int, int) ([]entity.PayrollPeriod, error)
	FindPayrollPeriodsByPeriodCode(context.Context, string, int, int) ([]entity.PayrollPeriod, error)
	FindPayrollPeriodsByPeriodStatus(context.Context, string, int, int) ([]entity.PayrollPeriod, error)
	FindPayrollPeriodsByCalendarReferenceID(context.Context, *string, int, int) ([]entity.PayrollPeriod, error)
	FindByPayrollGroupID(ctx context.Context, value string, limit, offset int) ([]entity.PayrollPeriod, error)
	FindByEmployingLegalEntityReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.PayrollPeriod, error)
	FindByCurrencyCode(ctx context.Context, value string, limit, offset int) ([]entity.PayrollPeriod, error)
	FindByJurisdictionCode(ctx context.Context, value string, limit, offset int) ([]entity.PayrollPeriod, error)
	FindByPeriodCode(ctx context.Context, value string, limit, offset int) ([]entity.PayrollPeriod, error)
	FindByPeriodStatus(ctx context.Context, value string, limit, offset int) ([]entity.PayrollPeriod, error)
	FindByCalendarReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.PayrollPeriod, error)
}

// PayrollRunRepository adalah port persistensi hr_payroll_runs.
type PayrollRunRepository interface {
	Repository[entity.PayrollRun, string]
	InsertPayrollRuns(context.Context, *entity.PayrollRun) error
	UpdatePayrollRuns(context.Context, *entity.PayrollRun) error
	DeletePayrollRunsByUUID(context.Context, string) error
	FindPayrollRunsByUUID(context.Context, string) (*entity.PayrollRun, error)
	FindPayrollRunsByUUIDForUpdate(context.Context, string) (*entity.PayrollRun, error)
	GetAllPayrollRuns(context.Context, int, int) ([]entity.PayrollRun, error)
	CountPayrollRuns(context.Context) (int64, error)
	SelectPayrollRunsWhere(context.Context, []Filter, int, int) ([]entity.PayrollRun, error)
	CountPayrollRunsWhere(context.Context, []Filter) (int64, error)
	FindPayrollRunsByPayrollPeriodID(context.Context, string, int, int) ([]entity.PayrollRun, error)
	FindPayrollRunsByPayrollGroupID(context.Context, string, int, int) ([]entity.PayrollRun, error)
	FindPayrollRunsByEmployingLegalEntityReferenceID(context.Context, string, int, int) ([]entity.PayrollRun, error)
	FindPayrollRunsByCurrencyCode(context.Context, string, int, int) ([]entity.PayrollRun, error)
	FindPayrollRunsByRuleSetID(context.Context, string, int, int) ([]entity.PayrollRun, error)
	FindPayrollRunsByOriginalRunID(context.Context, *string, int, int) ([]entity.PayrollRun, error)
	FindPayrollRunsByRunNumber(context.Context, int32, int, int) ([]entity.PayrollRun, error)
	FindPayrollRunsByRunType(context.Context, string, int, int) ([]entity.PayrollRun, error)
	FindPayrollRunsByStatus(context.Context, entity.PayrollRunStatus, int, int) ([]entity.PayrollRun, error)
	FindPayrollRunsByJurisdictionCode(context.Context, string, int, int) ([]entity.PayrollRun, error)
	FindPayrollRunsByInputSnapshotReferenceID(context.Context, string, int, int) ([]entity.PayrollRun, error)
	FindPayrollRunsByApprovedByPrincipalID(context.Context, *string, int, int) ([]entity.PayrollRun, error)
	FindPayrollRunsByFinanceJournalReferenceID(context.Context, *string, int, int) ([]entity.PayrollRun, error)
	FindPayrollRunsByPaymentBatchReferenceID(context.Context, *string, int, int) ([]entity.PayrollRun, error)
	FindPayrollRunsByFinanceSettlementEventReferenceID(context.Context, *string, int, int) ([]entity.PayrollRun, error)
	FindPayrollRunsByStartedByPrincipalID(context.Context, string, int, int) ([]entity.PayrollRun, error)
	FindPayrollRunsByIdempotencyKey(context.Context, string, int, int) ([]entity.PayrollRun, error)
	FindByPayrollPeriodID(ctx context.Context, value string, limit, offset int) ([]entity.PayrollRun, error)
	FindByPayrollGroupID(ctx context.Context, value string, limit, offset int) ([]entity.PayrollRun, error)
	FindByEmployingLegalEntityReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.PayrollRun, error)
	FindByCurrencyCode(ctx context.Context, value string, limit, offset int) ([]entity.PayrollRun, error)
	FindByRuleSetID(ctx context.Context, value string, limit, offset int) ([]entity.PayrollRun, error)
	FindByOriginalRunID(ctx context.Context, value *string, limit, offset int) ([]entity.PayrollRun, error)
	FindByRunNumber(ctx context.Context, value int32, limit, offset int) ([]entity.PayrollRun, error)
	FindByRunType(ctx context.Context, value string, limit, offset int) ([]entity.PayrollRun, error)
	FindByStatus(ctx context.Context, value entity.PayrollRunStatus, limit, offset int) ([]entity.PayrollRun, error)
	FindByJurisdictionCode(ctx context.Context, value string, limit, offset int) ([]entity.PayrollRun, error)
	FindByInputSnapshotReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.PayrollRun, error)
	FindByApprovedByPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.PayrollRun, error)
	FindByFinanceJournalReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.PayrollRun, error)
	FindByPaymentBatchReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.PayrollRun, error)
	FindByFinanceSettlementEventReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.PayrollRun, error)
	FindByStartedByPrincipalID(ctx context.Context, value string, limit, offset int) ([]entity.PayrollRun, error)
	FindByIdempotencyKey(ctx context.Context, value string, limit, offset int) ([]entity.PayrollRun, error)
}

// PayrollResultRepository adalah port persistensi hr_payroll_results.
type PayrollResultRepository interface {
	Repository[entity.PayrollResult, string]
	InsertPayrollResults(context.Context, *entity.PayrollResult) error
	UpdatePayrollResults(context.Context, *entity.PayrollResult) error
	DeletePayrollResultsByUUID(context.Context, string) error
	FindPayrollResultsByUUID(context.Context, string) (*entity.PayrollResult, error)
	FindPayrollResultsByUUIDForUpdate(context.Context, string) (*entity.PayrollResult, error)
	GetAllPayrollResults(context.Context, int, int) ([]entity.PayrollResult, error)
	CountPayrollResults(context.Context) (int64, error)
	SelectPayrollResultsWhere(context.Context, []Filter, int, int) ([]entity.PayrollResult, error)
	CountPayrollResultsWhere(context.Context, []Filter) (int64, error)
	FindPayrollResultsByPayrollRunID(context.Context, string, int, int) ([]entity.PayrollResult, error)
	FindPayrollResultsByPayrollGroupID(context.Context, string, int, int) ([]entity.PayrollResult, error)
	FindPayrollResultsByWorkerID(context.Context, string, int, int) ([]entity.PayrollResult, error)
	FindPayrollResultsByPayrollProfileReferenceID(context.Context, string, int, int) ([]entity.PayrollResult, error)
	FindPayrollResultsByCurrencyCode(context.Context, string, int, int) ([]entity.PayrollResult, error)
	FindPayrollResultsByResultStatus(context.Context, string, int, int) ([]entity.PayrollResult, error)
	FindPayrollResultsByPaymentReferenceID(context.Context, *string, int, int) ([]entity.PayrollResult, error)
	FindPayrollResultsByPayslipDocumentReferenceID(context.Context, *string, int, int) ([]entity.PayrollResult, error)
	FindByPayrollRunID(ctx context.Context, value string, limit, offset int) ([]entity.PayrollResult, error)
	FindByPayrollGroupID(ctx context.Context, value string, limit, offset int) ([]entity.PayrollResult, error)
	FindByWorkerID(ctx context.Context, value string, limit, offset int) ([]entity.PayrollResult, error)
	FindByPayrollProfileReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.PayrollResult, error)
	FindByCurrencyCode(ctx context.Context, value string, limit, offset int) ([]entity.PayrollResult, error)
	FindByResultStatus(ctx context.Context, value string, limit, offset int) ([]entity.PayrollResult, error)
	FindByPaymentReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.PayrollResult, error)
	FindByPayslipDocumentReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.PayrollResult, error)
}

// PayrollResultLineRepository adalah port persistensi hr_payroll_result_lines.
type PayrollResultLineRepository interface {
	Repository[entity.PayrollResultLine, string]
	InsertPayrollResultLines(context.Context, *entity.PayrollResultLine) error
	UpdatePayrollResultLines(context.Context, *entity.PayrollResultLine) error
	DeletePayrollResultLinesByUUID(context.Context, string) error
	FindPayrollResultLinesByUUID(context.Context, string) (*entity.PayrollResultLine, error)
	FindPayrollResultLinesByUUIDForUpdate(context.Context, string) (*entity.PayrollResultLine, error)
	GetAllPayrollResultLines(context.Context, int, int) ([]entity.PayrollResultLine, error)
	CountPayrollResultLines(context.Context) (int64, error)
	SelectPayrollResultLinesWhere(context.Context, []Filter, int, int) ([]entity.PayrollResultLine, error)
	CountPayrollResultLinesWhere(context.Context, []Filter) (int64, error)
	FindPayrollResultLinesByPayrollResultID(context.Context, string, int, int) ([]entity.PayrollResultLine, error)
	FindPayrollResultLinesByComponentID(context.Context, string, int, int) ([]entity.PayrollResultLine, error)
	FindByPayrollResultID(ctx context.Context, value string, limit, offset int) ([]entity.PayrollResultLine, error)
	FindByComponentID(ctx context.Context, value string, limit, offset int) ([]entity.PayrollResultLine, error)
}

// PayrollAdjustmentRepository adalah port persistensi hr_payroll_adjustments.
type PayrollAdjustmentRepository interface {
	Repository[entity.PayrollAdjustment, string]
	InsertPayrollAdjustments(context.Context, *entity.PayrollAdjustment) error
	UpdatePayrollAdjustments(context.Context, *entity.PayrollAdjustment) error
	DeletePayrollAdjustmentsByUUID(context.Context, string) error
	FindPayrollAdjustmentsByUUID(context.Context, string) (*entity.PayrollAdjustment, error)
	FindPayrollAdjustmentsByUUIDForUpdate(context.Context, string) (*entity.PayrollAdjustment, error)
	GetAllPayrollAdjustments(context.Context, int, int) ([]entity.PayrollAdjustment, error)
	CountPayrollAdjustments(context.Context) (int64, error)
	SelectPayrollAdjustmentsWhere(context.Context, []Filter, int, int) ([]entity.PayrollAdjustment, error)
	CountPayrollAdjustmentsWhere(context.Context, []Filter) (int64, error)
	FindPayrollAdjustmentsByWorkerID(context.Context, string, int, int) ([]entity.PayrollAdjustment, error)
	FindPayrollAdjustmentsByPayrollPeriodID(context.Context, string, int, int) ([]entity.PayrollAdjustment, error)
	FindPayrollAdjustmentsByComponentID(context.Context, string, int, int) ([]entity.PayrollAdjustment, error)
	FindPayrollAdjustmentsByAdjustmentType(context.Context, string, int, int) ([]entity.PayrollAdjustment, error)
	FindPayrollAdjustmentsByCurrencyCode(context.Context, string, int, int) ([]entity.PayrollAdjustment, error)
	FindPayrollAdjustmentsByRequestedByPrincipalID(context.Context, string, int, int) ([]entity.PayrollAdjustment, error)
	FindPayrollAdjustmentsByApprovedByPrincipalID(context.Context, *string, int, int) ([]entity.PayrollAdjustment, error)
	FindPayrollAdjustmentsByIdempotencyKey(context.Context, string, int, int) ([]entity.PayrollAdjustment, error)
	FindByWorkerID(ctx context.Context, value string, limit, offset int) ([]entity.PayrollAdjustment, error)
	FindByPayrollPeriodID(ctx context.Context, value string, limit, offset int) ([]entity.PayrollAdjustment, error)
	FindByComponentID(ctx context.Context, value string, limit, offset int) ([]entity.PayrollAdjustment, error)
	FindByAdjustmentType(ctx context.Context, value string, limit, offset int) ([]entity.PayrollAdjustment, error)
	FindByCurrencyCode(ctx context.Context, value string, limit, offset int) ([]entity.PayrollAdjustment, error)
	FindByRequestedByPrincipalID(ctx context.Context, value string, limit, offset int) ([]entity.PayrollAdjustment, error)
	FindByApprovedByPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.PayrollAdjustment, error)
	FindByIdempotencyKey(ctx context.Context, value string, limit, offset int) ([]entity.PayrollAdjustment, error)
}

// PayrollGroupRepository adalah port persistensi hr_payroll_groups.
type PayrollGroupRepository interface {
	Repository[entity.PayrollGroup, string]
	InsertPayrollGroups(context.Context, *entity.PayrollGroup) error
	UpdatePayrollGroups(context.Context, *entity.PayrollGroup) error
	DeletePayrollGroupsByUUID(context.Context, string) error
	FindPayrollGroupsByUUID(context.Context, string) (*entity.PayrollGroup, error)
	FindPayrollGroupsByUUIDForUpdate(context.Context, string) (*entity.PayrollGroup, error)
	GetAllPayrollGroups(context.Context, int, int) ([]entity.PayrollGroup, error)
	CountPayrollGroups(context.Context) (int64, error)
	SelectPayrollGroupsWhere(context.Context, []Filter, int, int) ([]entity.PayrollGroup, error)
	CountPayrollGroupsWhere(context.Context, []Filter) (int64, error)
	FindPayrollGroupsByGroupCode(context.Context, string, int, int) ([]entity.PayrollGroup, error)
	FindPayrollGroupsByEmployingLegalEntityReferenceID(context.Context, string, int, int) ([]entity.PayrollGroup, error)
	FindPayrollGroupsByJurisdictionCode(context.Context, string, int, int) ([]entity.PayrollGroup, error)
	FindPayrollGroupsByCurrencyCode(context.Context, string, int, int) ([]entity.PayrollGroup, error)
	FindByGroupCode(ctx context.Context, value string, limit, offset int) ([]entity.PayrollGroup, error)
	FindByEmployingLegalEntityReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.PayrollGroup, error)
	FindByJurisdictionCode(ctx context.Context, value string, limit, offset int) ([]entity.PayrollGroup, error)
	FindByCurrencyCode(ctx context.Context, value string, limit, offset int) ([]entity.PayrollGroup, error)
}

// PayrollRuleSetRepository adalah port persistensi hr_payroll_rule_sets.
type PayrollRuleSetRepository interface {
	Repository[entity.PayrollRuleSet, string]
	InsertPayrollRuleSets(context.Context, *entity.PayrollRuleSet) error
	UpdatePayrollRuleSets(context.Context, *entity.PayrollRuleSet) error
	DeletePayrollRuleSetsByUUID(context.Context, string) error
	FindPayrollRuleSetsByUUID(context.Context, string) (*entity.PayrollRuleSet, error)
	FindPayrollRuleSetsByUUIDForUpdate(context.Context, string) (*entity.PayrollRuleSet, error)
	GetAllPayrollRuleSets(context.Context, int, int) ([]entity.PayrollRuleSet, error)
	CountPayrollRuleSets(context.Context) (int64, error)
	SelectPayrollRuleSetsWhere(context.Context, []Filter, int, int) ([]entity.PayrollRuleSet, error)
	CountPayrollRuleSetsWhere(context.Context, []Filter) (int64, error)
	FindPayrollRuleSetsByRuleCode(context.Context, string, int, int) ([]entity.PayrollRuleSet, error)
	FindPayrollRuleSetsByJurisdictionCode(context.Context, string, int, int) ([]entity.PayrollRuleSet, error)
	FindPayrollRuleSetsByArtifactDocumentVersionReferenceID(context.Context, string, int, int) ([]entity.PayrollRuleSet, error)
	FindPayrollRuleSetsByLegalBasisDocumentVersionReferenceID(context.Context, string, int, int) ([]entity.PayrollRuleSet, error)
	FindPayrollRuleSetsByApprovedByPrincipalID(context.Context, string, int, int) ([]entity.PayrollRuleSet, error)
	FindByRuleCode(ctx context.Context, value string, limit, offset int) ([]entity.PayrollRuleSet, error)
	FindByJurisdictionCode(ctx context.Context, value string, limit, offset int) ([]entity.PayrollRuleSet, error)
	FindByArtifactDocumentVersionReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.PayrollRuleSet, error)
	FindByLegalBasisDocumentVersionReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.PayrollRuleSet, error)
	FindByApprovedByPrincipalID(ctx context.Context, value string, limit, offset int) ([]entity.PayrollRuleSet, error)
}
