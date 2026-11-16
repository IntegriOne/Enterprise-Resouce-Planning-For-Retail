// Kerangka awal dibuat dengan alat bantu; kode ini dipelihara langsung per domain.
// Aturan bisnis lintas record tetap dijalankan melalui policy/integrity dan command domain.
package service

import (
	"fmt"
	"reflect"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

func createPayrollProfile(in CreatePayrollProfileInput) (*entity.PayrollProfile, error) {
	v := &entity.PayrollProfile{}
	v.WorkerID = in.WorkerID
	v.PayrollCountryCode = in.PayrollCountryCode
	v.PayrollJurisdictionCode = in.PayrollJurisdictionCode
	v.PayrollGroupID = in.PayrollGroupID
	v.EmployingLegalEntityReferenceID = in.EmployingLegalEntityReferenceID
	v.CurrencyCode = in.CurrencyCode
	v.TaxProfileReferenceID = in.TaxProfileReferenceID
	v.BankAccountTokenReferenceID = in.BankAccountTokenReferenceID
	v.TaxStatusCode = in.TaxStatusCode
	v.WithholdingElectionReferenceID = in.WithholdingElectionReferenceID
	v.EffectiveFrom = in.EffectiveFrom
	v.EffectiveTo = in.EffectiveTo
	if in.IsActive != nil {
		v.IsActive = *in.IsActive
	} else {
		v.IsActive = true
	}
	return v, nil
}
func patchPayrollProfile(v *entity.PayrollProfile, in UpdatePayrollProfileInput) error {
	if in.WorkerID.Set {
		if in.WorkerID.Value == nil {
			return hrerrors.Invalid("worker_id", "must not be null")
		}
		v.WorkerID = *in.WorkerID.Value
	}
	if in.PayrollCountryCode.Set {
		if in.PayrollCountryCode.Value == nil {
			return hrerrors.Invalid("payroll_country_code", "must not be null")
		}
		v.PayrollCountryCode = *in.PayrollCountryCode.Value
	}
	if in.PayrollJurisdictionCode.Set {
		if in.PayrollJurisdictionCode.Value == nil {
			return hrerrors.Invalid("payroll_jurisdiction_code", "must not be null")
		}
		v.PayrollJurisdictionCode = *in.PayrollJurisdictionCode.Value
	}
	if in.PayrollGroupID.Set {
		if in.PayrollGroupID.Value == nil {
			return hrerrors.Invalid("payroll_group_id", "must not be null")
		}
		v.PayrollGroupID = *in.PayrollGroupID.Value
	}
	if in.EmployingLegalEntityReferenceID.Set {
		if in.EmployingLegalEntityReferenceID.Value == nil {
			return hrerrors.Invalid("employing_legal_entity_reference_id", "must not be null")
		}
		v.EmployingLegalEntityReferenceID = *in.EmployingLegalEntityReferenceID.Value
	}
	if in.CurrencyCode.Set {
		if in.CurrencyCode.Value == nil {
			return hrerrors.Invalid("currency_code", "must not be null")
		}
		v.CurrencyCode = *in.CurrencyCode.Value
	}
	if in.TaxProfileReferenceID.Set {
		v.TaxProfileReferenceID = in.TaxProfileReferenceID.Value
	}
	if in.BankAccountTokenReferenceID.Set {
		v.BankAccountTokenReferenceID = in.BankAccountTokenReferenceID.Value
	}
	if in.TaxStatusCode.Set {
		v.TaxStatusCode = in.TaxStatusCode.Value
	}
	if in.WithholdingElectionReferenceID.Set {
		v.WithholdingElectionReferenceID = in.WithholdingElectionReferenceID.Value
	}
	if in.EffectiveFrom.Set {
		if in.EffectiveFrom.Value == nil {
			return hrerrors.Invalid("effective_from", "must not be null")
		}
		v.EffectiveFrom = *in.EffectiveFrom.Value
	}
	if in.EffectiveTo.Set {
		v.EffectiveTo = in.EffectiveTo.Value
	}
	if in.IsActive.Set {
		if in.IsActive.Value == nil {
			return hrerrors.Invalid("is_active", "must not be null")
		}
		v.IsActive = *in.IsActive.Value
	}
	return nil
}
func specPayrollProfile() operation.Specification[entity.PayrollProfile, string] {
	return operation.Specification[entity.PayrollProfile, string]{
		Domain: "payroll", Table: "hr_payroll_profiles",
		Key:         func(v *entity.PayrollProfile) string { return v.ID },
		KeyString:   func(key string) string { return fmt.Sprint(key) },
		ValidateKey: func(key string) error { return entity.ValidateUUID("id", key) },
		AggregateID: func(v *entity.PayrollProfile) string { return v.ID },
		Initialize: func(v *entity.PayrollProfile, actor operation.Actor, now time.Time, id string) {
			v.ID = id
		},
		Touch: func(v *entity.PayrollProfile, actor operation.Actor, now time.Time) {},
		Attributes: func(v *entity.PayrollProfile) map[string]string {
			a := map[string]string{}
			a["id"] = fmt.Sprint(v.ID)
			a["worker_id"] = fmt.Sprint(v.WorkerID)
			a["payroll_group_id"] = fmt.Sprint(v.PayrollGroupID)
			a["employing_legal_entity_reference_id"] = fmt.Sprint(v.EmployingLegalEntityReferenceID)
			if v.TaxProfileReferenceID != nil {
				a["tax_profile_reference_id"] = fmt.Sprint(*v.TaxProfileReferenceID)
			}
			if v.BankAccountTokenReferenceID != nil {
				a["bank_account_token_reference_id"] = fmt.Sprint(*v.BankAccountTokenReferenceID)
			}
			if v.WithholdingElectionReferenceID != nil {
				a["withholding_election_reference_id"] = fmt.Sprint(*v.WithholdingElectionReferenceID)
			}
			return a
		},
		Guard: func(action string, old, next *entity.PayrollProfile) error {
			return nil
		},
	}
}
func createPayComponent(in CreatePayComponentInput) (*entity.PayComponent, error) {
	v := &entity.PayComponent{}
	v.ComponentCode = in.ComponentCode
	v.ComponentVersion = in.ComponentVersion
	v.ComponentName = in.ComponentName
	v.ComponentType = in.ComponentType
	v.CalculationMethod = in.CalculationMethod
	v.TaxTreatmentCode = in.TaxTreatmentCode
	v.JurisdictionCode = in.JurisdictionCode
	v.CurrencyCode = in.CurrencyCode
	v.RuleSetID = in.RuleSetID
	v.FormulaReferenceID = in.FormulaReferenceID
	v.Taxable = in.Taxable
	v.Pensionable = in.Pensionable
	v.EffectiveFrom = in.EffectiveFrom
	v.EffectiveTo = in.EffectiveTo
	if in.IsActive != nil {
		v.IsActive = *in.IsActive
	} else {
		v.IsActive = true
	}
	return v, nil
}
func patchPayComponent(v *entity.PayComponent, in UpdatePayComponentInput) error {
	if in.ComponentCode.Set {
		if in.ComponentCode.Value == nil {
			return hrerrors.Invalid("component_code", "must not be null")
		}
		v.ComponentCode = *in.ComponentCode.Value
	}
	if in.ComponentVersion.Set {
		if in.ComponentVersion.Value == nil {
			return hrerrors.Invalid("component_version", "must not be null")
		}
		v.ComponentVersion = *in.ComponentVersion.Value
	}
	if in.ComponentName.Set {
		if in.ComponentName.Value == nil {
			return hrerrors.Invalid("component_name", "must not be null")
		}
		v.ComponentName = *in.ComponentName.Value
	}
	if in.ComponentType.Set {
		if in.ComponentType.Value == nil {
			return hrerrors.Invalid("component_type", "must not be null")
		}
		v.ComponentType = *in.ComponentType.Value
	}
	if in.CalculationMethod.Set {
		if in.CalculationMethod.Value == nil {
			return hrerrors.Invalid("calculation_method", "must not be null")
		}
		v.CalculationMethod = *in.CalculationMethod.Value
	}
	if in.TaxTreatmentCode.Set {
		v.TaxTreatmentCode = in.TaxTreatmentCode.Value
	}
	if in.JurisdictionCode.Set {
		if in.JurisdictionCode.Value == nil {
			return hrerrors.Invalid("jurisdiction_code", "must not be null")
		}
		v.JurisdictionCode = *in.JurisdictionCode.Value
	}
	if in.CurrencyCode.Set {
		v.CurrencyCode = in.CurrencyCode.Value
	}
	if in.RuleSetID.Set {
		v.RuleSetID = in.RuleSetID.Value
	}
	if in.FormulaReferenceID.Set {
		v.FormulaReferenceID = in.FormulaReferenceID.Value
	}
	if in.Taxable.Set {
		v.Taxable = in.Taxable.Value
	}
	if in.Pensionable.Set {
		v.Pensionable = in.Pensionable.Value
	}
	if in.EffectiveFrom.Set {
		if in.EffectiveFrom.Value == nil {
			return hrerrors.Invalid("effective_from", "must not be null")
		}
		v.EffectiveFrom = *in.EffectiveFrom.Value
	}
	if in.EffectiveTo.Set {
		v.EffectiveTo = in.EffectiveTo.Value
	}
	if in.IsActive.Set {
		if in.IsActive.Value == nil {
			return hrerrors.Invalid("is_active", "must not be null")
		}
		v.IsActive = *in.IsActive.Value
	}
	return nil
}
func specPayComponent() operation.Specification[entity.PayComponent, string] {
	return operation.Specification[entity.PayComponent, string]{
		Domain: "payroll", Table: "hr_pay_components",
		Key:         func(v *entity.PayComponent) string { return v.ID },
		KeyString:   func(key string) string { return fmt.Sprint(key) },
		ValidateKey: func(key string) error { return entity.ValidateUUID("id", key) },
		AggregateID: func(v *entity.PayComponent) string { return v.ID },
		Initialize: func(v *entity.PayComponent, actor operation.Actor, now time.Time, id string) {
			v.ID = id
		},
		Touch: func(v *entity.PayComponent, actor operation.Actor, now time.Time) {},
		Attributes: func(v *entity.PayComponent) map[string]string {
			a := map[string]string{}
			a["id"] = fmt.Sprint(v.ID)
			if v.RuleSetID != nil {
				a["rule_set_id"] = fmt.Sprint(*v.RuleSetID)
			}
			if v.FormulaReferenceID != nil {
				a["formula_reference_id"] = fmt.Sprint(*v.FormulaReferenceID)
			}
			return a
		},
		Guard: func(action string, old, next *entity.PayComponent) error {
			return nil
		},
	}
}
func createPayrollPeriod(in CreatePayrollPeriodInput) (*entity.PayrollPeriod, error) {
	v := &entity.PayrollPeriod{}
	v.PayrollGroupID = in.PayrollGroupID
	v.EmployingLegalEntityReferenceID = in.EmployingLegalEntityReferenceID
	v.CurrencyCode = in.CurrencyCode
	v.JurisdictionCode = in.JurisdictionCode
	v.PeriodCode = in.PeriodCode
	v.PeriodStart = in.PeriodStart
	v.PeriodEnd = in.PeriodEnd
	v.PayDate = in.PayDate
	if in.PeriodStatus != nil {
		v.PeriodStatus = *in.PeriodStatus
	} else {
		v.PeriodStatus = "open"
	}
	v.CalendarReferenceID = in.CalendarReferenceID
	return v, nil
}
func patchPayrollPeriod(v *entity.PayrollPeriod, in UpdatePayrollPeriodInput) error {
	if in.PayrollGroupID.Set {
		if in.PayrollGroupID.Value == nil {
			return hrerrors.Invalid("payroll_group_id", "must not be null")
		}
		v.PayrollGroupID = *in.PayrollGroupID.Value
	}
	if in.EmployingLegalEntityReferenceID.Set {
		if in.EmployingLegalEntityReferenceID.Value == nil {
			return hrerrors.Invalid("employing_legal_entity_reference_id", "must not be null")
		}
		v.EmployingLegalEntityReferenceID = *in.EmployingLegalEntityReferenceID.Value
	}
	if in.CurrencyCode.Set {
		if in.CurrencyCode.Value == nil {
			return hrerrors.Invalid("currency_code", "must not be null")
		}
		v.CurrencyCode = *in.CurrencyCode.Value
	}
	if in.JurisdictionCode.Set {
		if in.JurisdictionCode.Value == nil {
			return hrerrors.Invalid("jurisdiction_code", "must not be null")
		}
		v.JurisdictionCode = *in.JurisdictionCode.Value
	}
	if in.PeriodCode.Set {
		if in.PeriodCode.Value == nil {
			return hrerrors.Invalid("period_code", "must not be null")
		}
		v.PeriodCode = *in.PeriodCode.Value
	}
	if in.PeriodStart.Set {
		if in.PeriodStart.Value == nil {
			return hrerrors.Invalid("period_start", "must not be null")
		}
		v.PeriodStart = *in.PeriodStart.Value
	}
	if in.PeriodEnd.Set {
		if in.PeriodEnd.Value == nil {
			return hrerrors.Invalid("period_end", "must not be null")
		}
		v.PeriodEnd = *in.PeriodEnd.Value
	}
	if in.PayDate.Set {
		if in.PayDate.Value == nil {
			return hrerrors.Invalid("pay_date", "must not be null")
		}
		v.PayDate = *in.PayDate.Value
	}
	if in.PeriodStatus.Set {
		if in.PeriodStatus.Value == nil {
			return hrerrors.Invalid("period_status", "must not be null")
		}
		v.PeriodStatus = *in.PeriodStatus.Value
	}
	if in.CalendarReferenceID.Set {
		v.CalendarReferenceID = in.CalendarReferenceID.Value
	}
	return nil
}
func specPayrollPeriod() operation.Specification[entity.PayrollPeriod, string] {
	return operation.Specification[entity.PayrollPeriod, string]{
		Domain: "payroll", Table: "hr_payroll_periods",
		Key:         func(v *entity.PayrollPeriod) string { return v.ID },
		KeyString:   func(key string) string { return fmt.Sprint(key) },
		ValidateKey: func(key string) error { return entity.ValidateUUID("id", key) },
		AggregateID: func(v *entity.PayrollPeriod) string { return v.ID },
		Initialize: func(v *entity.PayrollPeriod, actor operation.Actor, now time.Time, id string) {
			v.ID = id
		},
		Touch: func(v *entity.PayrollPeriod, actor operation.Actor, now time.Time) {},
		Attributes: func(v *entity.PayrollPeriod) map[string]string {
			a := map[string]string{}
			a["id"] = fmt.Sprint(v.ID)
			a["payroll_group_id"] = fmt.Sprint(v.PayrollGroupID)
			a["employing_legal_entity_reference_id"] = fmt.Sprint(v.EmployingLegalEntityReferenceID)
			a["period_status"] = fmt.Sprint(v.PeriodStatus)
			if v.CalendarReferenceID != nil {
				a["calendar_reference_id"] = fmt.Sprint(*v.CalendarReferenceID)
			}
			return a
		},
		Guard: func(action string, old, next *entity.PayrollPeriod) error {
			if action == "update" && !reflect.DeepEqual(old.PeriodStatus, next.PeriodStatus) {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "create" && string(next.PeriodStatus) != "open" {
				return hrerrors.ErrWorkflowRequired
			}
			if action != "create" {
				switch string(old.PeriodStatus) {
				case "approved", "posted", "paid", "finalized", "applied", "closed", "archived", "terminated", "deceased", "completed", "accepted", "issued", "signed", "hired", "processed", "rejected", "cancelled", "withdrawn", "superseded":
					return hrerrors.ErrWorkflowRequired
				}
			}
			return nil
		},
	}
}
func createPayrollRun(in CreatePayrollRunInput) (*entity.PayrollRun, error) {
	v := &entity.PayrollRun{}
	v.PayrollPeriodID = in.PayrollPeriodID
	v.PayrollGroupID = in.PayrollGroupID
	v.EmployingLegalEntityReferenceID = in.EmployingLegalEntityReferenceID
	v.CurrencyCode = in.CurrencyCode
	v.RuleSetID = in.RuleSetID
	v.OriginalRunID = in.OriginalRunID
	if in.SettlementDirection != nil {
		v.SettlementDirection = *in.SettlementDirection
	} else {
		v.SettlementDirection = "payable"
	}
	v.RunNumber = in.RunNumber
	v.RunType = in.RunType
	if in.Status != nil {
		v.Status = *in.Status
	} else {
		v.Status = "draft"
	}
	v.JurisdictionCode = in.JurisdictionCode
	v.RuleEngineName = in.RuleEngineName
	v.RuleEngineVersion = in.RuleEngineVersion
	v.InputSnapshotReferenceID = in.InputSnapshotReferenceID
	v.InputDigest = in.InputDigest
	v.OutputDigest = in.OutputDigest
	v.CalculatedAt = in.CalculatedAt
	v.ApprovedByPrincipalID = in.ApprovedByPrincipalID
	v.ApprovedAt = in.ApprovedAt
	v.FinanceJournalReferenceID = in.FinanceJournalReferenceID
	v.PaymentBatchReferenceID = in.PaymentBatchReferenceID
	v.FinanceSettlementEventReferenceID = in.FinanceSettlementEventReferenceID
	v.FinanceSettledAt = in.FinanceSettledAt
	v.IdempotencyKey = in.IdempotencyKey
	return v, nil
}
func patchPayrollRun(v *entity.PayrollRun, in UpdatePayrollRunInput) error {
	if in.PayrollPeriodID.Set {
		if in.PayrollPeriodID.Value == nil {
			return hrerrors.Invalid("payroll_period_id", "must not be null")
		}
		v.PayrollPeriodID = *in.PayrollPeriodID.Value
	}
	if in.PayrollGroupID.Set {
		if in.PayrollGroupID.Value == nil {
			return hrerrors.Invalid("payroll_group_id", "must not be null")
		}
		v.PayrollGroupID = *in.PayrollGroupID.Value
	}
	if in.EmployingLegalEntityReferenceID.Set {
		if in.EmployingLegalEntityReferenceID.Value == nil {
			return hrerrors.Invalid("employing_legal_entity_reference_id", "must not be null")
		}
		v.EmployingLegalEntityReferenceID = *in.EmployingLegalEntityReferenceID.Value
	}
	if in.CurrencyCode.Set {
		if in.CurrencyCode.Value == nil {
			return hrerrors.Invalid("currency_code", "must not be null")
		}
		v.CurrencyCode = *in.CurrencyCode.Value
	}
	if in.RuleSetID.Set {
		if in.RuleSetID.Value == nil {
			return hrerrors.Invalid("rule_set_id", "must not be null")
		}
		v.RuleSetID = *in.RuleSetID.Value
	}
	if in.OriginalRunID.Set {
		v.OriginalRunID = in.OriginalRunID.Value
	}
	if in.SettlementDirection.Set {
		if in.SettlementDirection.Value == nil {
			return hrerrors.Invalid("settlement_direction", "must not be null")
		}
		v.SettlementDirection = *in.SettlementDirection.Value
	}
	if in.RunNumber.Set {
		if in.RunNumber.Value == nil {
			return hrerrors.Invalid("run_number", "must not be null")
		}
		v.RunNumber = *in.RunNumber.Value
	}
	if in.RunType.Set {
		if in.RunType.Value == nil {
			return hrerrors.Invalid("run_type", "must not be null")
		}
		v.RunType = *in.RunType.Value
	}
	if in.Status.Set {
		if in.Status.Value == nil {
			return hrerrors.Invalid("status", "must not be null")
		}
		v.Status = *in.Status.Value
	}
	if in.JurisdictionCode.Set {
		if in.JurisdictionCode.Value == nil {
			return hrerrors.Invalid("jurisdiction_code", "must not be null")
		}
		v.JurisdictionCode = *in.JurisdictionCode.Value
	}
	if in.RuleEngineName.Set {
		if in.RuleEngineName.Value == nil {
			return hrerrors.Invalid("rule_engine_name", "must not be null")
		}
		v.RuleEngineName = *in.RuleEngineName.Value
	}
	if in.RuleEngineVersion.Set {
		if in.RuleEngineVersion.Value == nil {
			return hrerrors.Invalid("rule_engine_version", "must not be null")
		}
		v.RuleEngineVersion = *in.RuleEngineVersion.Value
	}
	if in.InputSnapshotReferenceID.Set {
		if in.InputSnapshotReferenceID.Value == nil {
			return hrerrors.Invalid("input_snapshot_reference_id", "must not be null")
		}
		v.InputSnapshotReferenceID = *in.InputSnapshotReferenceID.Value
	}
	if in.InputDigest.Set {
		if in.InputDigest.Value == nil {
			v.InputDigest = nil
		} else {
			v.InputDigest = *in.InputDigest.Value
		}
	}
	if in.OutputDigest.Set {
		if in.OutputDigest.Value == nil {
			v.OutputDigest = nil
		} else {
			v.OutputDigest = *in.OutputDigest.Value
		}
	}
	if in.CalculatedAt.Set {
		v.CalculatedAt = in.CalculatedAt.Value
	}
	if in.ApprovedByPrincipalID.Set {
		v.ApprovedByPrincipalID = in.ApprovedByPrincipalID.Value
	}
	if in.ApprovedAt.Set {
		v.ApprovedAt = in.ApprovedAt.Value
	}
	if in.FinanceJournalReferenceID.Set {
		v.FinanceJournalReferenceID = in.FinanceJournalReferenceID.Value
	}
	if in.PaymentBatchReferenceID.Set {
		v.PaymentBatchReferenceID = in.PaymentBatchReferenceID.Value
	}
	if in.FinanceSettlementEventReferenceID.Set {
		v.FinanceSettlementEventReferenceID = in.FinanceSettlementEventReferenceID.Value
	}
	if in.FinanceSettledAt.Set {
		v.FinanceSettledAt = in.FinanceSettledAt.Value
	}
	if in.IdempotencyKey.Set {
		if in.IdempotencyKey.Value == nil {
			return hrerrors.Invalid("idempotency_key", "must not be null")
		}
		v.IdempotencyKey = *in.IdempotencyKey.Value
	}
	return nil
}
func specPayrollRun() operation.Specification[entity.PayrollRun, string] {
	return operation.Specification[entity.PayrollRun, string]{
		Domain: "payroll", Table: "hr_payroll_runs",
		Key:         func(v *entity.PayrollRun) string { return v.ID },
		KeyString:   func(key string) string { return fmt.Sprint(key) },
		ValidateKey: func(key string) error { return entity.ValidateUUID("id", key) },
		AggregateID: func(v *entity.PayrollRun) string { return v.ID },
		Initialize: func(v *entity.PayrollRun, actor operation.Actor, now time.Time, id string) {
			v.ID = id
			v.StartedByPrincipalID = actor.PrincipalID
			v.CreatedAt = now
		},
		Touch: func(v *entity.PayrollRun, actor operation.Actor, now time.Time) {},
		Attributes: func(v *entity.PayrollRun) map[string]string {
			a := map[string]string{}
			a["id"] = fmt.Sprint(v.ID)
			a["payroll_period_id"] = fmt.Sprint(v.PayrollPeriodID)
			a["payroll_group_id"] = fmt.Sprint(v.PayrollGroupID)
			a["employing_legal_entity_reference_id"] = fmt.Sprint(v.EmployingLegalEntityReferenceID)
			a["rule_set_id"] = fmt.Sprint(v.RuleSetID)
			if v.OriginalRunID != nil {
				a["original_run_id"] = fmt.Sprint(*v.OriginalRunID)
			}
			a["status"] = fmt.Sprint(v.Status)
			a["input_snapshot_reference_id"] = fmt.Sprint(v.InputSnapshotReferenceID)
			if v.ApprovedByPrincipalID != nil {
				a["approved_by_principal_id"] = fmt.Sprint(*v.ApprovedByPrincipalID)
			}
			if v.FinanceJournalReferenceID != nil {
				a["finance_journal_reference_id"] = fmt.Sprint(*v.FinanceJournalReferenceID)
			}
			if v.PaymentBatchReferenceID != nil {
				a["payment_batch_reference_id"] = fmt.Sprint(*v.PaymentBatchReferenceID)
			}
			if v.FinanceSettlementEventReferenceID != nil {
				a["finance_settlement_event_reference_id"] = fmt.Sprint(*v.FinanceSettlementEventReferenceID)
			}
			a["started_by_principal_id"] = fmt.Sprint(v.StartedByPrincipalID)
			return a
		},
		Guard: func(action string, old, next *entity.PayrollRun) error {
			if action == "update" && !reflect.DeepEqual(old.Status, next.Status) {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "create" && string(next.Status) != "draft" {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "create" && next.ApprovedByPrincipalID != nil {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "update" && !reflect.DeepEqual(old.ApprovedByPrincipalID, next.ApprovedByPrincipalID) {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "create" && next.ApprovedAt != nil {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "update" && !reflect.DeepEqual(old.ApprovedAt, next.ApprovedAt) {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "create" && next.FinanceJournalReferenceID != nil {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "update" && !reflect.DeepEqual(old.FinanceJournalReferenceID, next.FinanceJournalReferenceID) {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "create" && next.PaymentBatchReferenceID != nil {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "update" && !reflect.DeepEqual(old.PaymentBatchReferenceID, next.PaymentBatchReferenceID) {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "create" && next.FinanceSettlementEventReferenceID != nil {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "update" && !reflect.DeepEqual(old.FinanceSettlementEventReferenceID, next.FinanceSettlementEventReferenceID) {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "create" && next.FinanceSettledAt != nil {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "update" && !reflect.DeepEqual(old.FinanceSettledAt, next.FinanceSettledAt) {
				return hrerrors.ErrWorkflowRequired
			}
			if action != "create" {
				switch string(old.Status) {
				case "approved", "posted", "paid", "finalized", "applied", "closed", "archived", "terminated", "deceased", "completed", "accepted", "issued", "signed", "hired", "processed", "rejected", "cancelled", "withdrawn", "superseded":
					return hrerrors.ErrWorkflowRequired
				}
			}
			return nil
		},
	}
}
func createPayrollResult(in CreatePayrollResultInput) (*entity.PayrollResult, error) {
	v := &entity.PayrollResult{}
	v.PayrollRunID = in.PayrollRunID
	v.PayrollGroupID = in.PayrollGroupID
	v.WorkerID = in.WorkerID
	v.PayrollProfileReferenceID = in.PayrollProfileReferenceID
	v.CurrencyCode = in.CurrencyCode
	v.GrossAmount = in.GrossAmount
	if in.ReimbursementAmount != nil {
		v.ReimbursementAmount = *in.ReimbursementAmount
	} else {
		v.ReimbursementAmount = entity.Decimal("0")
	}
	v.TotalDeductions = in.TotalDeductions
	v.EmployerCostAmount = in.EmployerCostAmount
	v.NetPayAmount = in.NetPayAmount
	if in.ResultStatus != nil {
		v.ResultStatus = *in.ResultStatus
	} else {
		v.ResultStatus = "calculated"
	}
	v.PaymentReferenceID = in.PaymentReferenceID
	v.PayslipDocumentReferenceID = in.PayslipDocumentReferenceID
	v.CalculationDigest = in.CalculationDigest
	return v, nil
}
func patchPayrollResult(v *entity.PayrollResult, in UpdatePayrollResultInput) error {
	if in.PayrollRunID.Set {
		if in.PayrollRunID.Value == nil {
			return hrerrors.Invalid("payroll_run_id", "must not be null")
		}
		v.PayrollRunID = *in.PayrollRunID.Value
	}
	if in.PayrollGroupID.Set {
		if in.PayrollGroupID.Value == nil {
			return hrerrors.Invalid("payroll_group_id", "must not be null")
		}
		v.PayrollGroupID = *in.PayrollGroupID.Value
	}
	if in.WorkerID.Set {
		if in.WorkerID.Value == nil {
			return hrerrors.Invalid("worker_id", "must not be null")
		}
		v.WorkerID = *in.WorkerID.Value
	}
	if in.PayrollProfileReferenceID.Set {
		if in.PayrollProfileReferenceID.Value == nil {
			return hrerrors.Invalid("payroll_profile_reference_id", "must not be null")
		}
		v.PayrollProfileReferenceID = *in.PayrollProfileReferenceID.Value
	}
	if in.CurrencyCode.Set {
		if in.CurrencyCode.Value == nil {
			return hrerrors.Invalid("currency_code", "must not be null")
		}
		v.CurrencyCode = *in.CurrencyCode.Value
	}
	if in.GrossAmount.Set {
		if in.GrossAmount.Value == nil {
			return hrerrors.Invalid("gross_amount", "must not be null")
		}
		v.GrossAmount = *in.GrossAmount.Value
	}
	if in.ReimbursementAmount.Set {
		if in.ReimbursementAmount.Value == nil {
			return hrerrors.Invalid("reimbursement_amount", "must not be null")
		}
		v.ReimbursementAmount = *in.ReimbursementAmount.Value
	}
	if in.TotalDeductions.Set {
		if in.TotalDeductions.Value == nil {
			return hrerrors.Invalid("total_deductions", "must not be null")
		}
		v.TotalDeductions = *in.TotalDeductions.Value
	}
	if in.EmployerCostAmount.Set {
		if in.EmployerCostAmount.Value == nil {
			return hrerrors.Invalid("employer_cost_amount", "must not be null")
		}
		v.EmployerCostAmount = *in.EmployerCostAmount.Value
	}
	if in.NetPayAmount.Set {
		if in.NetPayAmount.Value == nil {
			return hrerrors.Invalid("net_pay_amount", "must not be null")
		}
		v.NetPayAmount = *in.NetPayAmount.Value
	}
	if in.ResultStatus.Set {
		if in.ResultStatus.Value == nil {
			return hrerrors.Invalid("result_status", "must not be null")
		}
		v.ResultStatus = *in.ResultStatus.Value
	}
	if in.PaymentReferenceID.Set {
		v.PaymentReferenceID = in.PaymentReferenceID.Value
	}
	if in.PayslipDocumentReferenceID.Set {
		v.PayslipDocumentReferenceID = in.PayslipDocumentReferenceID.Value
	}
	if in.CalculationDigest.Set {
		if in.CalculationDigest.Value == nil {
			v.CalculationDigest = nil
		} else {
			v.CalculationDigest = *in.CalculationDigest.Value
		}
	}
	return nil
}
func specPayrollResult() operation.Specification[entity.PayrollResult, string] {
	return operation.Specification[entity.PayrollResult, string]{
		Domain: "payroll", Table: "hr_payroll_results",
		Key:         func(v *entity.PayrollResult) string { return v.ID },
		KeyString:   func(key string) string { return fmt.Sprint(key) },
		ValidateKey: func(key string) error { return entity.ValidateUUID("id", key) },
		AggregateID: func(v *entity.PayrollResult) string { return v.ID },
		Initialize: func(v *entity.PayrollResult, actor operation.Actor, now time.Time, id string) {
			v.ID = id
			v.CreatedAt = now
		},
		Touch: func(v *entity.PayrollResult, actor operation.Actor, now time.Time) {},
		Attributes: func(v *entity.PayrollResult) map[string]string {
			a := map[string]string{}
			a["id"] = fmt.Sprint(v.ID)
			a["payroll_run_id"] = fmt.Sprint(v.PayrollRunID)
			a["payroll_group_id"] = fmt.Sprint(v.PayrollGroupID)
			a["worker_id"] = fmt.Sprint(v.WorkerID)
			a["payroll_profile_reference_id"] = fmt.Sprint(v.PayrollProfileReferenceID)
			a["result_status"] = fmt.Sprint(v.ResultStatus)
			if v.PaymentReferenceID != nil {
				a["payment_reference_id"] = fmt.Sprint(*v.PaymentReferenceID)
			}
			if v.PayslipDocumentReferenceID != nil {
				a["payslip_document_reference_id"] = fmt.Sprint(*v.PayslipDocumentReferenceID)
			}
			return a
		},
		Guard: func(action string, old, next *entity.PayrollResult) error {
			if action == "create" {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "update" && !reflect.DeepEqual(old.ResultStatus, next.ResultStatus) {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "create" && string(next.ResultStatus) != "calculated" {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "create" && next.PaymentReferenceID != nil {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "update" && !reflect.DeepEqual(old.PaymentReferenceID, next.PaymentReferenceID) {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "create" && next.PayslipDocumentReferenceID != nil {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "update" && !reflect.DeepEqual(old.PayslipDocumentReferenceID, next.PayslipDocumentReferenceID) {
				return hrerrors.ErrWorkflowRequired
			}
			if action != "create" {
				switch string(old.ResultStatus) {
				case "approved", "posted", "paid", "finalized", "applied", "closed", "archived", "terminated", "deceased", "completed", "accepted", "issued", "signed", "hired", "processed", "rejected", "cancelled", "withdrawn", "superseded":
					return hrerrors.ErrWorkflowRequired
				}
			}
			return nil
		},
	}
}
func createPayrollResultLine(in CreatePayrollResultLineInput) (*entity.PayrollResultLine, error) {
	v := &entity.PayrollResultLine{}
	v.PayrollResultID = in.PayrollResultID
	v.ComponentID = in.ComponentID
	v.ComponentCodeSnapshot = in.ComponentCodeSnapshot
	v.ComponentNameSnapshot = in.ComponentNameSnapshot
	v.ComponentTypeSnapshot = in.ComponentTypeSnapshot
	v.Amount = in.Amount
	v.Quantity = in.Quantity
	v.Rate = in.Rate
	v.TaxableBase = in.TaxableBase
	v.CalculationReference = in.CalculationReference
	v.LineSequence = in.LineSequence
	return v, nil
}
func patchPayrollResultLine(v *entity.PayrollResultLine, in UpdatePayrollResultLineInput) error {
	if in.PayrollResultID.Set {
		if in.PayrollResultID.Value == nil {
			return hrerrors.Invalid("payroll_result_id", "must not be null")
		}
		v.PayrollResultID = *in.PayrollResultID.Value
	}
	if in.ComponentID.Set {
		if in.ComponentID.Value == nil {
			return hrerrors.Invalid("component_id", "must not be null")
		}
		v.ComponentID = *in.ComponentID.Value
	}
	if in.ComponentCodeSnapshot.Set {
		if in.ComponentCodeSnapshot.Value == nil {
			return hrerrors.Invalid("component_code_snapshot", "must not be null")
		}
		v.ComponentCodeSnapshot = *in.ComponentCodeSnapshot.Value
	}
	if in.ComponentNameSnapshot.Set {
		if in.ComponentNameSnapshot.Value == nil {
			return hrerrors.Invalid("component_name_snapshot", "must not be null")
		}
		v.ComponentNameSnapshot = *in.ComponentNameSnapshot.Value
	}
	if in.ComponentTypeSnapshot.Set {
		if in.ComponentTypeSnapshot.Value == nil {
			return hrerrors.Invalid("component_type_snapshot", "must not be null")
		}
		v.ComponentTypeSnapshot = *in.ComponentTypeSnapshot.Value
	}
	if in.Amount.Set {
		if in.Amount.Value == nil {
			return hrerrors.Invalid("amount", "must not be null")
		}
		v.Amount = *in.Amount.Value
	}
	if in.Quantity.Set {
		v.Quantity = in.Quantity.Value
	}
	if in.Rate.Set {
		v.Rate = in.Rate.Value
	}
	if in.TaxableBase.Set {
		v.TaxableBase = in.TaxableBase.Value
	}
	if in.CalculationReference.Set {
		v.CalculationReference = in.CalculationReference.Value
	}
	if in.LineSequence.Set {
		if in.LineSequence.Value == nil {
			return hrerrors.Invalid("line_sequence", "must not be null")
		}
		v.LineSequence = *in.LineSequence.Value
	}
	return nil
}
func specPayrollResultLine() operation.Specification[entity.PayrollResultLine, string] {
	return operation.Specification[entity.PayrollResultLine, string]{
		Domain: "payroll", Table: "hr_payroll_result_lines",
		Key:         func(v *entity.PayrollResultLine) string { return v.ID },
		KeyString:   func(key string) string { return fmt.Sprint(key) },
		ValidateKey: func(key string) error { return entity.ValidateUUID("id", key) },
		AggregateID: func(v *entity.PayrollResultLine) string { return v.ID },
		Initialize: func(v *entity.PayrollResultLine, actor operation.Actor, now time.Time, id string) {
			v.ID = id
		},
		Touch: func(v *entity.PayrollResultLine, actor operation.Actor, now time.Time) {},
		Attributes: func(v *entity.PayrollResultLine) map[string]string {
			a := map[string]string{}
			a["id"] = fmt.Sprint(v.ID)
			a["payroll_result_id"] = fmt.Sprint(v.PayrollResultID)
			a["component_id"] = fmt.Sprint(v.ComponentID)
			return a
		},
		Guard: func(action string, old, next *entity.PayrollResultLine) error {
			if action == "create" {
				return hrerrors.ErrWorkflowRequired
			}
			return nil
		},
	}
}
func createPayrollAdjustment(in CreatePayrollAdjustmentInput) (*entity.PayrollAdjustment, error) {
	v := &entity.PayrollAdjustment{}
	v.WorkerID = in.WorkerID
	v.PayrollPeriodID = in.PayrollPeriodID
	v.ComponentID = in.ComponentID
	v.AdjustmentType = in.AdjustmentType
	v.Amount = in.Amount
	v.CurrencyCode = in.CurrencyCode
	v.Reason = in.Reason
	v.SourceReference = in.SourceReference
	v.ApprovedByPrincipalID = in.ApprovedByPrincipalID
	v.ApprovedAt = in.ApprovedAt
	v.IdempotencyKey = in.IdempotencyKey
	return v, nil
}
func patchPayrollAdjustment(v *entity.PayrollAdjustment, in UpdatePayrollAdjustmentInput) error {
	if in.WorkerID.Set {
		if in.WorkerID.Value == nil {
			return hrerrors.Invalid("worker_id", "must not be null")
		}
		v.WorkerID = *in.WorkerID.Value
	}
	if in.PayrollPeriodID.Set {
		if in.PayrollPeriodID.Value == nil {
			return hrerrors.Invalid("payroll_period_id", "must not be null")
		}
		v.PayrollPeriodID = *in.PayrollPeriodID.Value
	}
	if in.ComponentID.Set {
		if in.ComponentID.Value == nil {
			return hrerrors.Invalid("component_id", "must not be null")
		}
		v.ComponentID = *in.ComponentID.Value
	}
	if in.AdjustmentType.Set {
		if in.AdjustmentType.Value == nil {
			return hrerrors.Invalid("adjustment_type", "must not be null")
		}
		v.AdjustmentType = *in.AdjustmentType.Value
	}
	if in.Amount.Set {
		if in.Amount.Value == nil {
			return hrerrors.Invalid("amount", "must not be null")
		}
		v.Amount = *in.Amount.Value
	}
	if in.CurrencyCode.Set {
		if in.CurrencyCode.Value == nil {
			return hrerrors.Invalid("currency_code", "must not be null")
		}
		v.CurrencyCode = *in.CurrencyCode.Value
	}
	if in.Reason.Set {
		if in.Reason.Value == nil {
			return hrerrors.Invalid("reason", "must not be null")
		}
		v.Reason = *in.Reason.Value
	}
	if in.SourceReference.Set {
		v.SourceReference = in.SourceReference.Value
	}
	if in.ApprovedByPrincipalID.Set {
		v.ApprovedByPrincipalID = in.ApprovedByPrincipalID.Value
	}
	if in.ApprovedAt.Set {
		v.ApprovedAt = in.ApprovedAt.Value
	}
	if in.IdempotencyKey.Set {
		if in.IdempotencyKey.Value == nil {
			return hrerrors.Invalid("idempotency_key", "must not be null")
		}
		v.IdempotencyKey = *in.IdempotencyKey.Value
	}
	return nil
}
func specPayrollAdjustment() operation.Specification[entity.PayrollAdjustment, string] {
	return operation.Specification[entity.PayrollAdjustment, string]{
		Domain: "payroll", Table: "hr_payroll_adjustments",
		Key:         func(v *entity.PayrollAdjustment) string { return v.ID },
		KeyString:   func(key string) string { return fmt.Sprint(key) },
		ValidateKey: func(key string) error { return entity.ValidateUUID("id", key) },
		AggregateID: func(v *entity.PayrollAdjustment) string { return v.ID },
		Initialize: func(v *entity.PayrollAdjustment, actor operation.Actor, now time.Time, id string) {
			v.ID = id
			v.RequestedByPrincipalID = actor.PrincipalID
			v.CreatedAt = now
		},
		Touch: func(v *entity.PayrollAdjustment, actor operation.Actor, now time.Time) {},
		Attributes: func(v *entity.PayrollAdjustment) map[string]string {
			a := map[string]string{}
			a["id"] = fmt.Sprint(v.ID)
			a["worker_id"] = fmt.Sprint(v.WorkerID)
			a["payroll_period_id"] = fmt.Sprint(v.PayrollPeriodID)
			a["component_id"] = fmt.Sprint(v.ComponentID)
			a["requested_by_principal_id"] = fmt.Sprint(v.RequestedByPrincipalID)
			if v.ApprovedByPrincipalID != nil {
				a["approved_by_principal_id"] = fmt.Sprint(*v.ApprovedByPrincipalID)
			}
			return a
		},
		Guard: func(action string, old, next *entity.PayrollAdjustment) error {
			if action == "create" && next.ApprovedByPrincipalID != nil {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "update" && !reflect.DeepEqual(old.ApprovedByPrincipalID, next.ApprovedByPrincipalID) {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "create" && next.ApprovedAt != nil {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "update" && !reflect.DeepEqual(old.ApprovedAt, next.ApprovedAt) {
				return hrerrors.ErrWorkflowRequired
			}
			return nil
		},
	}
}
func createPayrollGroup(in CreatePayrollGroupInput) (*entity.PayrollGroup, error) {
	v := &entity.PayrollGroup{}
	v.GroupCode = in.GroupCode
	v.GroupName = in.GroupName
	v.EmployingLegalEntityReferenceID = in.EmployingLegalEntityReferenceID
	v.JurisdictionCode = in.JurisdictionCode
	v.CurrencyCode = in.CurrencyCode
	v.CurrencyScale = in.CurrencyScale
	v.PayFrequency = in.PayFrequency
	v.Timezone = in.Timezone
	if in.IsActive != nil {
		v.IsActive = *in.IsActive
	} else {
		v.IsActive = true
	}
	return v, nil
}
func patchPayrollGroup(v *entity.PayrollGroup, in UpdatePayrollGroupInput) error {
	if in.GroupCode.Set {
		if in.GroupCode.Value == nil {
			return hrerrors.Invalid("group_code", "must not be null")
		}
		v.GroupCode = *in.GroupCode.Value
	}
	if in.GroupName.Set {
		if in.GroupName.Value == nil {
			return hrerrors.Invalid("group_name", "must not be null")
		}
		v.GroupName = *in.GroupName.Value
	}
	if in.EmployingLegalEntityReferenceID.Set {
		if in.EmployingLegalEntityReferenceID.Value == nil {
			return hrerrors.Invalid("employing_legal_entity_reference_id", "must not be null")
		}
		v.EmployingLegalEntityReferenceID = *in.EmployingLegalEntityReferenceID.Value
	}
	if in.JurisdictionCode.Set {
		if in.JurisdictionCode.Value == nil {
			return hrerrors.Invalid("jurisdiction_code", "must not be null")
		}
		v.JurisdictionCode = *in.JurisdictionCode.Value
	}
	if in.CurrencyCode.Set {
		if in.CurrencyCode.Value == nil {
			return hrerrors.Invalid("currency_code", "must not be null")
		}
		v.CurrencyCode = *in.CurrencyCode.Value
	}
	if in.CurrencyScale.Set {
		if in.CurrencyScale.Value == nil {
			return hrerrors.Invalid("currency_scale", "must not be null")
		}
		v.CurrencyScale = *in.CurrencyScale.Value
	}
	if in.PayFrequency.Set {
		if in.PayFrequency.Value == nil {
			return hrerrors.Invalid("pay_frequency", "must not be null")
		}
		v.PayFrequency = *in.PayFrequency.Value
	}
	if in.Timezone.Set {
		if in.Timezone.Value == nil {
			return hrerrors.Invalid("timezone", "must not be null")
		}
		v.Timezone = *in.Timezone.Value
	}
	if in.IsActive.Set {
		if in.IsActive.Value == nil {
			return hrerrors.Invalid("is_active", "must not be null")
		}
		v.IsActive = *in.IsActive.Value
	}
	return nil
}
func specPayrollGroup() operation.Specification[entity.PayrollGroup, string] {
	return operation.Specification[entity.PayrollGroup, string]{
		Domain: "payroll", Table: "hr_payroll_groups",
		Key:         func(v *entity.PayrollGroup) string { return v.ID },
		KeyString:   func(key string) string { return fmt.Sprint(key) },
		ValidateKey: func(key string) error { return entity.ValidateUUID("id", key) },
		AggregateID: func(v *entity.PayrollGroup) string { return v.ID },
		Initialize: func(v *entity.PayrollGroup, actor operation.Actor, now time.Time, id string) {
			v.ID = id
		},
		Touch: func(v *entity.PayrollGroup, actor operation.Actor, now time.Time) {},
		Attributes: func(v *entity.PayrollGroup) map[string]string {
			a := map[string]string{}
			a["id"] = fmt.Sprint(v.ID)
			a["employing_legal_entity_reference_id"] = fmt.Sprint(v.EmployingLegalEntityReferenceID)
			return a
		},
		Guard: func(action string, old, next *entity.PayrollGroup) error {
			return nil
		},
	}
}
func createPayrollRuleSet(in CreatePayrollRuleSetInput) (*entity.PayrollRuleSet, error) {
	v := &entity.PayrollRuleSet{}
	v.RuleCode = in.RuleCode
	v.RuleVersion = in.RuleVersion
	v.JurisdictionCode = in.JurisdictionCode
	v.EffectiveFrom = in.EffectiveFrom
	v.EffectiveTo = in.EffectiveTo
	v.ArtifactDocumentVersionReferenceID = in.ArtifactDocumentVersionReferenceID
	v.ArtifactDigest = in.ArtifactDigest
	v.LegalBasisDocumentVersionReferenceID = in.LegalBasisDocumentVersionReferenceID
	v.ApprovedByPrincipalID = in.ApprovedByPrincipalID
	v.ApprovedAt = in.ApprovedAt
	return v, nil
}
func patchPayrollRuleSet(v *entity.PayrollRuleSet, in UpdatePayrollRuleSetInput) error {
	if in.RuleCode.Set {
		if in.RuleCode.Value == nil {
			return hrerrors.Invalid("rule_code", "must not be null")
		}
		v.RuleCode = *in.RuleCode.Value
	}
	if in.RuleVersion.Set {
		if in.RuleVersion.Value == nil {
			return hrerrors.Invalid("rule_version", "must not be null")
		}
		v.RuleVersion = *in.RuleVersion.Value
	}
	if in.JurisdictionCode.Set {
		if in.JurisdictionCode.Value == nil {
			return hrerrors.Invalid("jurisdiction_code", "must not be null")
		}
		v.JurisdictionCode = *in.JurisdictionCode.Value
	}
	if in.EffectiveFrom.Set {
		if in.EffectiveFrom.Value == nil {
			return hrerrors.Invalid("effective_from", "must not be null")
		}
		v.EffectiveFrom = *in.EffectiveFrom.Value
	}
	if in.EffectiveTo.Set {
		v.EffectiveTo = in.EffectiveTo.Value
	}
	if in.ArtifactDocumentVersionReferenceID.Set {
		if in.ArtifactDocumentVersionReferenceID.Value == nil {
			return hrerrors.Invalid("artifact_document_version_reference_id", "must not be null")
		}
		v.ArtifactDocumentVersionReferenceID = *in.ArtifactDocumentVersionReferenceID.Value
	}
	if in.ArtifactDigest.Set {
		if in.ArtifactDigest.Value == nil {
			v.ArtifactDigest = nil
		} else {
			v.ArtifactDigest = *in.ArtifactDigest.Value
		}
	}
	if in.LegalBasisDocumentVersionReferenceID.Set {
		if in.LegalBasisDocumentVersionReferenceID.Value == nil {
			return hrerrors.Invalid("legal_basis_document_version_reference_id", "must not be null")
		}
		v.LegalBasisDocumentVersionReferenceID = *in.LegalBasisDocumentVersionReferenceID.Value
	}
	if in.ApprovedByPrincipalID.Set {
		if in.ApprovedByPrincipalID.Value == nil {
			return hrerrors.Invalid("approved_by_principal_id", "must not be null")
		}
		v.ApprovedByPrincipalID = *in.ApprovedByPrincipalID.Value
	}
	if in.ApprovedAt.Set {
		if in.ApprovedAt.Value == nil {
			return hrerrors.Invalid("approved_at", "must not be null")
		}
		v.ApprovedAt = *in.ApprovedAt.Value
	}
	return nil
}
func specPayrollRuleSet() operation.Specification[entity.PayrollRuleSet, string] {
	return operation.Specification[entity.PayrollRuleSet, string]{
		Domain: "payroll", Table: "hr_payroll_rule_sets",
		Key:         func(v *entity.PayrollRuleSet) string { return v.ID },
		KeyString:   func(key string) string { return fmt.Sprint(key) },
		ValidateKey: func(key string) error { return entity.ValidateUUID("id", key) },
		AggregateID: func(v *entity.PayrollRuleSet) string { return v.ID },
		Initialize: func(v *entity.PayrollRuleSet, actor operation.Actor, now time.Time, id string) {
			v.ID = id
		},
		Touch: func(v *entity.PayrollRuleSet, actor operation.Actor, now time.Time) {},
		Attributes: func(v *entity.PayrollRuleSet) map[string]string {
			a := map[string]string{}
			a["id"] = fmt.Sprint(v.ID)
			a["artifact_document_version_reference_id"] = fmt.Sprint(v.ArtifactDocumentVersionReferenceID)
			a["legal_basis_document_version_reference_id"] = fmt.Sprint(v.LegalBasisDocumentVersionReferenceID)
			a["approved_by_principal_id"] = fmt.Sprint(v.ApprovedByPrincipalID)
			return a
		},
		Guard: func(action string, old, next *entity.PayrollRuleSet) error {
			if action != "create" {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "create" {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "create" && !reflect.ValueOf(next.ApprovedByPrincipalID).IsZero() {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "update" && !reflect.DeepEqual(old.ApprovedByPrincipalID, next.ApprovedByPrincipalID) {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "create" && !reflect.ValueOf(next.ApprovedAt).IsZero() {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "update" && !reflect.DeepEqual(old.ApprovedAt, next.ApprovedAt) {
				return hrerrors.ErrWorkflowRequired
			}
			if action != "create" {
				return hrerrors.ErrWorkflowRequired
			}
			return nil
		},
	}
}
