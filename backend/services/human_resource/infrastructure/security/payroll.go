package security

import (
	"context"
	"reflect"
	"strings"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository"
)

type payrollGroupStore interface {
	FindByIDForUpdate(context.Context, string) (*entity.PayrollGroup, error)
	SelectWhere(context.Context, []domainrepo.Filter, int, int) ([]entity.PayrollGroup, error)
}

type payrollPeriodStore interface {
	FindByIDForUpdate(context.Context, string) (*entity.PayrollPeriod, error)
	SelectWhere(context.Context, []domainrepo.Filter, int, int) ([]entity.PayrollPeriod, error)
	CountWhere(context.Context, []domainrepo.Filter) (int64, error)
}

type payrollProfileStore interface {
	SelectWhere(context.Context, []domainrepo.Filter, int, int) ([]entity.PayrollProfile, error)
	CountWhere(context.Context, []domainrepo.Filter) (int64, error)
}

type payrollRuleSetReader interface {
	FindByIDForUpdate(context.Context, string) (*entity.PayrollRuleSet, error)
}

type payrollRunStore interface {
	FindByIDForUpdate(context.Context, string) (*entity.PayrollRun, error)
	SelectWhere(context.Context, []domainrepo.Filter, int, int) ([]entity.PayrollRun, error)
	CountWhere(context.Context, []domainrepo.Filter) (int64, error)
}

type payrollAdjustmentStore interface {
	SelectWhere(context.Context, []domainrepo.Filter, int, int) ([]entity.PayrollAdjustment, error)
	CountWhere(context.Context, []domainrepo.Filter) (int64, error)
}

type payrollWorkerReader interface {
	FindByIDForUpdate(context.Context, string) (*entity.Worker, error)
}

type payrollComponentReader interface {
	FindByIDForUpdate(context.Context, string) (*entity.PayComponent, error)
}

type payrollResultCounter interface {
	CountWhere(context.Context, []domainrepo.Filter) (int64, error)
}

type payrollRuleSetValidator struct{}

type payrollGroupValidator struct {
	groups     payrollGroupStore
	profiles   payrollProfileStore
	periods    payrollPeriodStore
	runs       payrollRunStore
	foundation FoundationVerifier
}

type PayrollFoundationVerifier interface {
	FoundationVerifier
	VerifyLegalEntityWindow(ctx context.Context, legalEntityID string, start, end time.Time) error
}

type payrollPeriodValidator struct {
	groups      payrollGroupStore
	periods     payrollPeriodStore
	runs        payrollRunStore
	adjustments payrollAdjustmentStore
	foundation  PayrollFoundationVerifier
}

type payrollProfileValidator struct {
	workers    payrollWorkerReader
	groups     payrollGroupStore
	profiles   payrollProfileStore
	results    payrollResultCounter
	foundation FoundationVerifier
}

type payrollRunValidator struct {
	periods    payrollPeriodStore
	groups     payrollGroupStore
	ruleSets   payrollRuleSetReader
	runs       payrollRunStore
	results    payrollResultCounter
	foundation PayrollFoundationVerifier
}

type payrollAdjustmentValidator struct {
	workers     payrollWorkerReader
	periods     payrollPeriodStore
	profiles    payrollProfileStore
	components  payrollComponentReader
	adjustments payrollAdjustmentStore
}

// PayrollValidators melindungi integritas grup, periode, profil, penyesuaian, rule set, dan approval payroll run.
func PayrollValidators(repos *repository.Registry, foundation PayrollFoundationVerifier) (map[string]MutationValidator, error) {
	if repos == nil || operation.RequireDependency(repos.Worker) != nil ||
		operation.RequireDependency(repos.PayrollGroup) != nil || operation.RequireDependency(repos.PayrollPeriod) != nil ||
		operation.RequireDependency(repos.PayrollProfile) != nil || operation.RequireDependency(repos.PayrollRuleSet) != nil ||
		operation.RequireDependency(repos.PayrollRun) != nil || operation.RequireDependency(repos.PayrollResult) != nil ||
		operation.RequireDependency(repos.PayComponent) != nil || operation.RequireDependency(repos.PayrollAdjustment) != nil {
		return nil, hrerrors.ErrUnavailable
	}
	return map[string]MutationValidator{
		"hr_payroll_rule_sets": &payrollRuleSetValidator{},
		"hr_payroll_groups": &payrollGroupValidator{
			groups: repos.PayrollGroup, profiles: repos.PayrollProfile, periods: repos.PayrollPeriod,
			runs: repos.PayrollRun, foundation: foundation,
		},
		"hr_payroll_periods": &payrollPeriodValidator{
			groups: repos.PayrollGroup, periods: repos.PayrollPeriod, runs: repos.PayrollRun,
			adjustments: repos.PayrollAdjustment, foundation: foundation,
		},
		"hr_payroll_profiles": &payrollProfileValidator{
			workers: repos.Worker, groups: repos.PayrollGroup, profiles: repos.PayrollProfile,
			results: repos.PayrollResult, foundation: foundation,
		},
		"hr_payroll_runs": &payrollRunValidator{
			periods: repos.PayrollPeriod, groups: repos.PayrollGroup, ruleSets: repos.PayrollRuleSet,
			runs: repos.PayrollRun, results: repos.PayrollResult, foundation: foundation,
		},
		"hr_payroll_adjustments": &payrollAdjustmentValidator{
			workers: repos.Worker, periods: repos.PayrollPeriod, profiles: repos.PayrollProfile,
			components: repos.PayComponent, adjustments: repos.PayrollAdjustment,
		},
	}, nil
}

func (v *payrollRuleSetValidator) ValidateMutation(ctx context.Context, table string, _, _ entity.Validatable) error {
	if v == nil || table != "hr_payroll_rule_sets" {
		return hrerrors.ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	// Rule set pajak/payroll immutable serta memerlukan sertifikasi DRM & otorisasi hukum terpisah.
	return hrerrors.ErrWorkflowRequired
}

func (v *payrollGroupValidator) ValidateMutation(ctx context.Context, table string, before, after entity.Validatable) error {
	if v == nil || table != "hr_payroll_groups" || operation.RequireDependency(v.groups) != nil ||
		operation.RequireDependency(v.profiles) != nil || operation.RequireDependency(v.periods) != nil ||
		operation.RequireDependency(v.runs) != nil {
		return hrerrors.ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	var previous *entity.PayrollGroup
	if before != nil {
		var ok bool
		previous, ok = before.(*entity.PayrollGroup)
		if !ok || previous == nil {
			return hrerrors.ErrInvalidInput
		}
	}
	if previous == nil && after == nil {
		return hrerrors.ErrInvalidInput
	}
	if after == nil {
		return v.rejectIfReferenced(ctx, previous.ID)
	}
	group, ok := after.(*entity.PayrollGroup)
	if !ok || group == nil {
		return hrerrors.ErrInvalidInput
	}
	if !validISO3UpperCode(group.CurrencyCode) {
		return hrerrors.Invalid("currency_code", "must be a 3-letter uppercase code")
	}
	if !supportedPayFrequency(group.PayFrequency) {
		return hrerrors.Invalid("pay_frequency", "unsupported pay frequency")
	}
	if strings.TrimSpace(group.Timezone) == "" || group.Timezone == "Local" {
		return hrerrors.Invalid("timezone", "an IANA timezone is required")
	}
	if _, err := time.LoadLocation(group.Timezone); err != nil {
		return hrerrors.Invalid("timezone", "must be a valid IANA timezone")
	}
	if previous != nil {
		if previous.ID != group.ID {
			return hrerrors.ErrInvalidInput
		}
		if previous.EmployingLegalEntityReferenceID != group.EmployingLegalEntityReferenceID ||
			previous.JurisdictionCode != group.JurisdictionCode || previous.CurrencyCode != group.CurrencyCode ||
			previous.CurrencyScale != group.CurrencyScale {
			if err := v.rejectIfReferenced(ctx, previous.ID); err != nil {
				return err
			}
		}
	}
	if err := verifyFoundationLegalEntity(ctx, v.foundation, group.EmployingLegalEntityReferenceID); err != nil {
		return err
	}
	matches, err := v.groups.SelectWhere(ctx, []domainrepo.Filter{
		{Column: "employing_legal_entity_reference_id", Values: []any{group.EmployingLegalEntityReferenceID}},
		{Column: "group_code", Values: []any{group.GroupCode}},
	}, 2, 0)
	if err != nil {
		return err
	}
	for _, existing := range matches {
		if existing.ID != group.ID {
			return hrerrors.ErrAlreadyExists
		}
	}
	return nil
}

func (v *payrollGroupValidator) rejectIfReferenced(ctx context.Context, groupID string) error {
	for _, counter := range []payrollResultCounter{v.profiles, v.periods, v.runs} {
		count, err := counter.CountWhere(ctx, []domainrepo.Filter{{Column: "payroll_group_id", Values: []any{groupID}}})
		if err != nil {
			return err
		}
		if count != 0 {
			return hrerrors.ErrReferenceConflict
		}
	}
	return nil
}

func (v *payrollPeriodValidator) ValidateMutation(ctx context.Context, table string, before, after entity.Validatable) error {
	if v == nil || table != "hr_payroll_periods" || operation.RequireDependency(v.groups) != nil ||
		operation.RequireDependency(v.periods) != nil || operation.RequireDependency(v.runs) != nil ||
		operation.RequireDependency(v.adjustments) != nil {
		return hrerrors.ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	var previous *entity.PayrollPeriod
	if before != nil {
		var ok bool
		previous, ok = before.(*entity.PayrollPeriod)
		if !ok || previous == nil || previous.PeriodStatus != "open" {
			return hrerrors.ErrWorkflowRequired
		}
	}
	if previous == nil && after == nil {
		return hrerrors.ErrInvalidInput
	}
	if after == nil {
		return v.rejectIfReferenced(ctx, previous.ID)
	}
	period, ok := after.(*entity.PayrollPeriod)
	if !ok || period == nil {
		return hrerrors.ErrInvalidInput
	}
	if period.CalendarReferenceID != nil {
		return hrerrors.ErrUnavailable
	}
	if period.PeriodStatus != "open" {
		return hrerrors.ErrWorkflowRequired
	}
	if !validISO3UpperCode(period.CurrencyCode) {
		return hrerrors.Invalid("currency_code", "must be a 3-letter uppercase code")
	}
	if entity.CompareDates(period.PayDate, period.PeriodStart) < 0 {
		return hrerrors.Invalid("pay_date", "must not precede period_start")
	}
	if previous != nil {
		if previous.ID != period.ID || previous.PayrollGroupID != period.PayrollGroupID ||
			previous.EmployingLegalEntityReferenceID != period.EmployingLegalEntityReferenceID ||
			previous.CurrencyCode != period.CurrencyCode || previous.JurisdictionCode != period.JurisdictionCode ||
			previous.PeriodCode != period.PeriodCode || entity.CompareDates(previous.PeriodStart, period.PeriodStart) != 0 ||
			entity.CompareDates(previous.PeriodEnd, period.PeriodEnd) != 0 ||
			entity.CompareDates(previous.PayDate, period.PayDate) != 0 {
			if err := v.rejectIfReferenced(ctx, previous.ID); err != nil {
				return err
			}
			if previous.ID != period.ID {
				return hrerrors.ErrInvalidInput
			}
		}
	}
	group, err := v.groups.FindByIDForUpdate(ctx, period.PayrollGroupID)
	if err != nil {
		return workerAssessmentReferenceError(err)
	}
	if group == nil || (previous == nil && !group.IsActive) ||
		!strings.EqualFold(group.EmployingLegalEntityReferenceID, period.EmployingLegalEntityReferenceID) ||
		group.JurisdictionCode != period.JurisdictionCode || group.CurrencyCode != period.CurrencyCode {
		return hrerrors.ErrReferenceConflict
	}
	if operation.RequireDependency(v.foundation) != nil {
		return hrerrors.ErrUnavailable
	}
	if err := v.foundation.VerifyLegalEntityWindow(ctx, period.EmployingLegalEntityReferenceID, period.PeriodStart, period.PeriodEnd); err != nil {
		return err
	}
	matches, err := v.periods.SelectWhere(ctx, []domainrepo.Filter{
		{Column: "payroll_group_id", Values: []any{period.PayrollGroupID}},
		{Column: "period_code", Values: []any{period.PeriodCode}},
	}, 2, 0)
	if err != nil {
		return err
	}
	for _, existing := range matches {
		if existing.ID != period.ID {
			return hrerrors.ErrAlreadyExists
		}
	}
	return nil
}

func (v *payrollPeriodValidator) rejectIfReferenced(ctx context.Context, periodID string) error {
	for _, counter := range []payrollResultCounter{v.runs, v.adjustments} {
		count, err := counter.CountWhere(ctx, []domainrepo.Filter{{Column: "payroll_period_id", Values: []any{periodID}}})
		if err != nil {
			return err
		}
		if count != 0 {
			return hrerrors.ErrReferenceConflict
		}
	}
	return nil
}

func (v *payrollProfileValidator) ValidateMutation(ctx context.Context, table string, before, after entity.Validatable) error {
	if v == nil || table != "hr_payroll_profiles" || operation.RequireDependency(v.workers) != nil ||
		operation.RequireDependency(v.groups) != nil || operation.RequireDependency(v.profiles) != nil ||
		operation.RequireDependency(v.results) != nil {
		return hrerrors.ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	var previous *entity.PayrollProfile
	if before != nil {
		var ok bool
		previous, ok = before.(*entity.PayrollProfile)
		if !ok || previous == nil {
			return hrerrors.ErrInvalidInput
		}
	}
	if previous == nil && after == nil {
		return hrerrors.ErrInvalidInput
	}
	if after == nil {
		count, err := v.results.CountWhere(ctx, []domainrepo.Filter{{Column: "payroll_profile_reference_id", Values: []any{previous.ID}}})
		if err != nil {
			return err
		}
		if count != 0 {
			return hrerrors.ErrReferenceConflict
		}
		return nil
	}
	profile, ok := after.(*entity.PayrollProfile)
	if !ok || profile == nil {
		return hrerrors.ErrInvalidInput
	}
	if profile.TaxProfileReferenceID != nil || profile.BankAccountTokenReferenceID != nil || profile.WithholdingElectionReferenceID != nil {
		return hrerrors.ErrUnavailable
	}
	if !validISO2UpperCode(profile.PayrollCountryCode) {
		return hrerrors.Invalid("payroll_country_code", "must be a 2-letter uppercase code")
	}
	if !validISO3UpperCode(profile.CurrencyCode) {
		return hrerrors.Invalid("currency_code", "must be a 3-letter uppercase code")
	}
	if previous != nil {
		if previous.ID != profile.ID || previous.WorkerID != profile.WorkerID ||
			previous.PayrollGroupID != profile.PayrollGroupID ||
			previous.PayrollJurisdictionCode != profile.PayrollJurisdictionCode ||
			previous.CurrencyCode != profile.CurrencyCode ||
			!strings.EqualFold(previous.EmployingLegalEntityReferenceID, profile.EmployingLegalEntityReferenceID) {
			return hrerrors.ErrWorkflowRequired
		}
		if entity.CompareDates(previous.EffectiveFrom, profile.EffectiveFrom) != 0 {
			count, err := v.results.CountWhere(ctx, []domainrepo.Filter{{Column: "payroll_profile_reference_id", Values: []any{previous.ID}}})
			if err != nil {
				return err
			}
			if count != 0 {
				return hrerrors.ErrReferenceConflict
			}
		}
	}
	worker, err := v.workers.FindByIDForUpdate(ctx, profile.WorkerID)
	if err != nil {
		return workerAssessmentReferenceError(err)
	}
	if worker == nil || (worker.Status != entity.WorkerStatusPreboarding && worker.Status != entity.WorkerStatusActive) ||
		!strings.EqualFold(worker.PrimaryLegalEntityReferenceID, profile.EmployingLegalEntityReferenceID) ||
		(worker.HireDate != nil && entity.CompareDates(profile.EffectiveFrom, *worker.HireDate) < 0) ||
		(worker.TerminationDate != nil && (profile.EffectiveTo == nil || entity.CompareDates(*profile.EffectiveTo, *worker.TerminationDate) > 0)) {
		return hrerrors.ErrReferenceConflict
	}
	group, err := v.groups.FindByIDForUpdate(ctx, profile.PayrollGroupID)
	if err != nil {
		return workerAssessmentReferenceError(err)
	}
	if group == nil || (previous == nil && !group.IsActive) ||
		group.JurisdictionCode != profile.PayrollJurisdictionCode ||
		group.CurrencyCode != profile.CurrencyCode ||
		!strings.EqualFold(group.EmployingLegalEntityReferenceID, profile.EmployingLegalEntityReferenceID) {
		return hrerrors.ErrReferenceConflict
	}
	if err := verifyFoundationLegalEntity(ctx, v.foundation, profile.EmployingLegalEntityReferenceID); err != nil {
		return err
	}
	filters := []domainrepo.Filter{
		{Column: "worker_id", Values: []any{profile.WorkerID}},
		{Column: "payroll_group_id", Values: []any{profile.PayrollGroupID}},
	}
	for offset := 0; ; {
		existing, err := v.profiles.SelectWhere(ctx, filters, domainrepo.MaxLimit, offset)
		if err != nil {
			return err
		}
		for _, row := range existing {
			if !strings.EqualFold(row.ID, profile.ID) && datesOverlap(profile.EffectiveFrom, profile.EffectiveTo, row.EffectiveFrom, row.EffectiveTo) {
				return hrerrors.ErrAlreadyExists
			}
		}
		if len(existing) < domainrepo.MaxLimit {
			return nil
		}
		offset += len(existing)
	}
}

func (v *payrollRunValidator) ValidateMutation(ctx context.Context, table string, before, after entity.Validatable) error {
	if v == nil || table != "hr_payroll_runs" || operation.RequireDependency(v.periods) != nil ||
		operation.RequireDependency(v.groups) != nil || operation.RequireDependency(v.ruleSets) != nil ||
		operation.RequireDependency(v.runs) != nil || operation.RequireDependency(v.results) != nil {
		return hrerrors.ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	var previous *entity.PayrollRun
	if before != nil {
		var ok bool
		previous, ok = before.(*entity.PayrollRun)
		if !ok || previous == nil {
			return hrerrors.ErrInvalidInput
		}
		if previous.Status != entity.PayrollRunStatusDraft &&
			previous.Status != entity.PayrollRunStatusCalculated &&
			previous.Status != entity.PayrollRunStatusReview &&
			previous.Status != entity.PayrollRunStatusApproved {
			return hrerrors.ErrWorkflowRequired
		}
	}
	if previous == nil && after == nil {
		return hrerrors.ErrInvalidInput
	}
	if after == nil {
		if previous.Status != entity.PayrollRunStatusDraft {
			return hrerrors.ErrWorkflowRequired
		}
		resultCount, err := v.results.CountWhere(ctx, []domainrepo.Filter{{Column: "payroll_run_id", Values: []any{previous.ID}}})
		if err != nil {
			return err
		}
		if resultCount != 0 {
			return hrerrors.ErrReferenceConflict
		}
		correctionCount, err := v.runs.CountWhere(ctx, []domainrepo.Filter{{Column: "original_run_id", Values: []any{previous.ID}}})
		if err != nil {
			return err
		}
		if correctionCount != 0 {
			return hrerrors.ErrReferenceConflict
		}
		return nil
	}
	run, ok := after.(*entity.PayrollRun)
	if !ok || run == nil {
		return hrerrors.ErrInvalidInput
	}
	if previous != nil && previous.Status == entity.PayrollRunStatusDraft && run.Status == entity.PayrollRunStatusCalculated {
		unchanged := *run
		unchanged.Status, unchanged.CalculatedAt, unchanged.OutputDigest = previous.Status, nil, nil
		if !reflect.DeepEqual(*previous, unchanged) || run.CalculatedAt == nil ||
			run.CalculatedAt.Before(run.CreatedAt) || len(run.OutputDigest) != 32 {
			return hrerrors.ErrWorkflowRequired
		}
		period, _, _, err := v.lockRunScopeReferences(ctx, run, true)
		if err != nil {
			return err
		}
		if period.PeriodStatus != "open" {
			return hrerrors.ErrWorkflowRequired
		}
		return nil
	}
	if previous != nil && previous.Status == entity.PayrollRunStatusReview && run.Status == entity.PayrollRunStatusCancelled {
		unchanged := *run
		unchanged.Status = previous.Status
		unchanged.CancelledAt, unchanged.CancelledByPrincipalID, unchanged.CancellationReason = nil, nil, nil
		if !reflect.DeepEqual(*previous, unchanged) {
			return hrerrors.ErrWorkflowRequired
		}
		return run.Validate()
	}
	approving := previous != nil && previous.Status == entity.PayrollRunStatusReview &&
		run.Status == entity.PayrollRunStatusApproved
	if approving {
		return v.validateApprovalTransition(ctx, previous, run)
	}
	if previous != nil && previous.Status == entity.PayrollRunStatusCalculated && run.Status == entity.PayrollRunStatusReview {
		unchanged := *run
		unchanged.Status = previous.Status
		if !reflect.DeepEqual(*previous, unchanged) || run.CalculatedAt == nil ||
			run.CalculatedAt.Before(run.CreatedAt) || len(run.OutputDigest) == 0 ||
			run.ApprovedAt != nil || run.ApprovedByPrincipalID != nil || run.FinanceJournalReferenceID != nil ||
			run.PaymentBatchReferenceID != nil || run.FinanceSettlementEventReferenceID != nil || run.FinanceSettledAt != nil {
			return hrerrors.ErrWorkflowRequired
		}
		period, _, ruleSet, err := v.lockRunScopeReferences(ctx, run, false)
		if err != nil {
			return err
		}
		if ruleSet.ApprovedAt.IsZero() || ruleSet.ApprovedAt.After(*run.CalculatedAt) {
			return hrerrors.ErrReferenceConflict
		}
		if period.PeriodStatus != "open" {
			return hrerrors.ErrWorkflowRequired
		}
		return nil
	}
	if previous != nil && previous.Status == entity.PayrollRunStatusApproved && run.Status == entity.PayrollRunStatusApproved &&
		previous.PostingRequestedAt == nil && previous.PostingRequestedByPrincipalID == nil &&
		run.PostingRequestedAt != nil && run.PostingRequestedByPrincipalID != nil {
		unchanged := *run
		unchanged.PostingRequestedAt, unchanged.PostingRequestedByPrincipalID = nil, nil
		if !reflect.DeepEqual(*previous, unchanged) {
			return hrerrors.ErrWorkflowRequired
		}
		_, _, ruleSet, err := v.lockRunScopeReferences(ctx, run, false)
		if err != nil {
			return err
		}
		if run.CalculatedAt == nil || ruleSet.ApprovedAt.IsZero() || ruleSet.ApprovedAt.After(*run.CalculatedAt) {
			return hrerrors.ErrReferenceConflict
		}
		return nil
	}
	if previous != nil && previous.Status != entity.PayrollRunStatusDraft {
		return hrerrors.ErrWorkflowRequired
	}
	if run.Status != entity.PayrollRunStatusDraft || len(run.OutputDigest) != 0 ||
		run.CalculatedAt != nil || run.ApprovedByPrincipalID != nil || run.ApprovedAt != nil ||
		run.FinanceJournalReferenceID != nil || run.PaymentBatchReferenceID != nil ||
		run.FinanceSettlementEventReferenceID != nil || run.FinanceSettledAt != nil {
		return hrerrors.ErrWorkflowRequired
	}
	if !validISO3UpperCode(run.CurrencyCode) {
		return hrerrors.Invalid("currency_code", "must be a 3-letter uppercase code")
	}
	if previous != nil {
		if previous.ID != run.ID || previous.PayrollPeriodID != run.PayrollPeriodID ||
			previous.PayrollGroupID != run.PayrollGroupID ||
			!strings.EqualFold(previous.EmployingLegalEntityReferenceID, run.EmployingLegalEntityReferenceID) ||
			previous.CurrencyCode != run.CurrencyCode || previous.RuleSetID != run.RuleSetID ||
			!reflect.DeepEqual(previous.OriginalRunID, run.OriginalRunID) ||
			previous.RunNumber != run.RunNumber || previous.RunType != run.RunType ||
			previous.JurisdictionCode != run.JurisdictionCode ||
			!strings.EqualFold(previous.StartedByPrincipalID, run.StartedByPrincipalID) ||
			!previous.CreatedAt.Equal(run.CreatedAt) || previous.IdempotencyKey != run.IdempotencyKey {
			return hrerrors.ErrWorkflowRequired
		}
	}
	period, _, ruleSet, err := v.lockRunScopeReferences(ctx, run, previous == nil)
	if err != nil {
		return err
	}
	if period.PeriodStatus != "open" {
		return hrerrors.ErrWorkflowRequired
	}
	if previous == nil && (ruleSet.ApprovedAt.IsZero() || ruleSet.ApprovedAt.After(run.CreatedAt)) {
		return hrerrors.ErrReferenceConflict
	}
	if run.OriginalRunID != nil {
		origin, err := v.runs.FindByIDForUpdate(ctx, *run.OriginalRunID)
		if err != nil {
			return workerAssessmentReferenceError(err)
		}
		if origin == nil || origin.PayrollGroupID != run.PayrollGroupID || origin.CurrencyCode != run.CurrencyCode ||
			(origin.Status != entity.PayrollRunStatusApproved && origin.Status != entity.PayrollRunStatusPosted && origin.Status != entity.PayrollRunStatusPaid) {
			return hrerrors.ErrReferenceConflict
		}
	}
	numMatches, err := v.runs.SelectWhere(ctx, []domainrepo.Filter{
		{Column: "payroll_period_id", Values: []any{run.PayrollPeriodID}},
		{Column: "run_number", Values: []any{run.RunNumber}},
	}, 2, 0)
	if err != nil {
		return err
	}
	for _, existing := range numMatches {
		if existing.ID != run.ID {
			return hrerrors.ErrAlreadyExists
		}
	}
	idemMatches, err := v.runs.SelectWhere(ctx, []domainrepo.Filter{
		{Column: "idempotency_key", Values: []any{run.IdempotencyKey}},
	}, 2, 0)
	if err != nil {
		return err
	}
	for _, existing := range idemMatches {
		if existing.ID != run.ID {
			return hrerrors.ErrAlreadyExists
		}
	}
	return nil
}

func (v *payrollRunValidator) validateApprovalTransition(ctx context.Context, previous, run *entity.PayrollRun) error {
	if run.CalculatedAt == nil || len(run.OutputDigest) == 0 ||
		run.ApprovedByPrincipalID == nil || run.ApprovedAt == nil || run.ApprovedAt.Before(*run.CalculatedAt) ||
		run.FinanceJournalReferenceID != nil || run.PaymentBatchReferenceID != nil ||
		run.FinanceSettlementEventReferenceID != nil || run.FinanceSettledAt != nil {
		return hrerrors.ErrWorkflowRequired
	}
	if strings.EqualFold(run.StartedByPrincipalID, *run.ApprovedByPrincipalID) {
		return hrerrors.ErrForbidden
	}
	unchanged := *run
	unchanged.Status = previous.Status
	unchanged.ApprovedByPrincipalID = nil
	unchanged.ApprovedAt = nil
	if !reflect.DeepEqual(*previous, unchanged) {
		return hrerrors.ErrWorkflowRequired
	}
	period, _, ruleSet, err := v.lockRunScopeReferences(ctx, run, false)
	if err != nil {
		return err
	}
	if ruleSet.ApprovedAt.IsZero() || ruleSet.ApprovedAt.After(*run.CalculatedAt) {
		return hrerrors.ErrReferenceConflict
	}
	if period.PeriodStatus != "open" {
		return hrerrors.ErrWorkflowRequired
	}
	resultCount, err := v.results.CountWhere(ctx, []domainrepo.Filter{{Column: "payroll_run_id", Values: []any{run.ID}}})
	if err != nil {
		return err
	}
	if resultCount == 0 {
		return hrerrors.Invalid("results", "payroll run requires results")
	}
	return nil
}

func (v *payrollRunValidator) lockRunScopeReferences(ctx context.Context, run *entity.PayrollRun, requireActiveGroup bool) (*entity.PayrollPeriod, *entity.PayrollGroup, *entity.PayrollRuleSet, error) {
	period, err := v.periods.FindByIDForUpdate(ctx, run.PayrollPeriodID)
	if err != nil {
		return nil, nil, nil, workerAssessmentReferenceError(err)
	}
	if period == nil || !strings.EqualFold(period.PayrollGroupID, run.PayrollGroupID) ||
		!strings.EqualFold(period.EmployingLegalEntityReferenceID, run.EmployingLegalEntityReferenceID) ||
		period.CurrencyCode != run.CurrencyCode || period.JurisdictionCode != run.JurisdictionCode {
		return nil, nil, nil, hrerrors.ErrReferenceConflict
	}
	group, err := v.groups.FindByIDForUpdate(ctx, run.PayrollGroupID)
	if err != nil {
		return nil, nil, nil, workerAssessmentReferenceError(err)
	}
	if group == nil || (requireActiveGroup && !group.IsActive) ||
		!strings.EqualFold(group.EmployingLegalEntityReferenceID, run.EmployingLegalEntityReferenceID) ||
		group.JurisdictionCode != run.JurisdictionCode || group.CurrencyCode != run.CurrencyCode {
		return nil, nil, nil, hrerrors.ErrReferenceConflict
	}
	ruleSet, err := v.ruleSets.FindByIDForUpdate(ctx, run.RuleSetID)
	if err != nil {
		return nil, nil, nil, workerAssessmentReferenceError(err)
	}
	if ruleSet == nil || ruleSet.JurisdictionCode != run.JurisdictionCode ||
		!compensationPeriodWithin(period.PeriodStart, &period.PeriodEnd, ruleSet.EffectiveFrom, ruleSet.EffectiveTo) {
		return nil, nil, nil, hrerrors.ErrReferenceConflict
	}
	if operation.RequireDependency(v.foundation) != nil {
		return nil, nil, nil, hrerrors.ErrUnavailable
	}
	if err := v.foundation.VerifyLegalEntityWindow(ctx, run.EmployingLegalEntityReferenceID, period.PeriodStart, period.PeriodEnd); err != nil {
		return nil, nil, nil, err
	}
	return period, group, ruleSet, nil
}

func (v *payrollAdjustmentValidator) ValidateMutation(ctx context.Context, table string, before, after entity.Validatable) error {
	if v == nil || table != "hr_payroll_adjustments" || operation.RequireDependency(v.workers) != nil ||
		operation.RequireDependency(v.periods) != nil || operation.RequireDependency(v.profiles) != nil ||
		operation.RequireDependency(v.components) != nil || operation.RequireDependency(v.adjustments) != nil {
		return hrerrors.ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	var previous *entity.PayrollAdjustment
	if before != nil {
		var ok bool
		previous, ok = before.(*entity.PayrollAdjustment)
		if !ok || previous == nil || previous.ApprovedByPrincipalID != nil || previous.ApprovedAt != nil {
			return hrerrors.ErrWorkflowRequired
		}
	}
	if previous == nil && after == nil {
		return hrerrors.ErrInvalidInput
	}
	adj := previous
	if after != nil {
		var ok bool
		adj, ok = after.(*entity.PayrollAdjustment)
		if !ok || adj == nil {
			return hrerrors.ErrInvalidInput
		}
		if adj.ApprovedByPrincipalID != nil || adj.ApprovedAt != nil {
			return hrerrors.ErrWorkflowRequired
		}
		if !validISO3UpperCode(adj.CurrencyCode) {
			return hrerrors.Invalid("currency_code", "must be a 3-letter uppercase code")
		}
		amountRat, err := adj.Amount.Rat()
		if err != nil || amountRat.Sign() == 0 {
			return hrerrors.Invalid("amount", "must be non-zero")
		}
		if previous != nil && (previous.ID != adj.ID || previous.WorkerID != adj.WorkerID ||
			previous.PayrollPeriodID != adj.PayrollPeriodID || previous.ComponentID != adj.ComponentID ||
			!strings.EqualFold(previous.RequestedByPrincipalID, adj.RequestedByPrincipalID) ||
			previous.IdempotencyKey != adj.IdempotencyKey || !previous.CreatedAt.Equal(adj.CreatedAt)) {
			return hrerrors.ErrWorkflowRequired
		}
	}
	period, err := v.periods.FindByIDForUpdate(ctx, adj.PayrollPeriodID)
	if err != nil {
		return workerAssessmentReferenceError(err)
	}
	if period == nil || period.PeriodStatus != "open" {
		return hrerrors.ErrWorkflowRequired
	}
	if after == nil {
		return nil
	}
	if adj.CurrencyCode != period.CurrencyCode {
		return hrerrors.ErrReferenceConflict
	}
	worker, err := v.workers.FindByIDForUpdate(ctx, adj.WorkerID)
	if err != nil {
		return workerAssessmentReferenceError(err)
	}
	if worker == nil || worker.Status != entity.WorkerStatusActive ||
		!strings.EqualFold(worker.PrimaryLegalEntityReferenceID, period.EmployingLegalEntityReferenceID) ||
		(worker.HireDate != nil && entity.CompareDates(*worker.HireDate, period.PeriodEnd) > 0) ||
		(worker.TerminationDate != nil && entity.CompareDates(*worker.TerminationDate, period.PeriodStart) < 0) {
		return hrerrors.ErrReferenceConflict
	}
	if err := v.requireActiveProfile(ctx, adj, period); err != nil {
		return err
	}
	component, err := v.components.FindByIDForUpdate(ctx, adj.ComponentID)
	if err != nil {
		return workerAssessmentReferenceError(err)
	}
	if component == nil || !component.IsActive || component.JurisdictionCode != period.JurisdictionCode ||
		(component.CurrencyCode != nil && *component.CurrencyCode != adj.CurrencyCode) ||
		!compensationPeriodWithin(period.PeriodStart, &period.PeriodEnd, component.EffectiveFrom, component.EffectiveTo) {
		return hrerrors.ErrReferenceConflict
	}
	idemMatches, err := v.adjustments.SelectWhere(ctx, []domainrepo.Filter{
		{Column: "idempotency_key", Values: []any{adj.IdempotencyKey}},
	}, 2, 0)
	if err != nil {
		return err
	}
	for _, existing := range idemMatches {
		if existing.ID != adj.ID {
			return hrerrors.ErrAlreadyExists
		}
	}
	return nil
}

func supportedPayFrequency(value string) bool {
	switch value {
	case "monthly", "biweekly", "semi_monthly", "weekly", "daily":
		return true
	default:
		return false
	}
}

func validISO2UpperCode(value string) bool {
	return len(value) == 2 && value[0] >= 'A' && value[0] <= 'Z' && value[1] >= 'A' && value[1] <= 'Z'
}

func datesOverlap(startA time.Time, endA *time.Time, startB time.Time, endB *time.Time) bool {
	if endA != nil && entity.CompareDates(*endA, startB) < 0 {
		return false
	}
	if endB != nil && entity.CompareDates(*endB, startA) < 0 {
		return false
	}
	return true
}

// requireActiveProfile memeriksa halaman profil sampai ditemukan profil yang mencakup periode.
func (v *payrollAdjustmentValidator) requireActiveProfile(ctx context.Context, adj *entity.PayrollAdjustment, period *entity.PayrollPeriod) error {
	filters := []domainrepo.Filter{
		{Column: "worker_id", Values: []any{adj.WorkerID}},
		{Column: "payroll_group_id", Values: []any{period.PayrollGroupID}},
		{Column: "is_active", Values: []any{true}},
	}
	for offset := 0; ; {
		profiles, err := v.profiles.SelectWhere(ctx, filters, domainrepo.MaxLimit, offset)
		if err != nil {
			return err
		}
		for _, profile := range profiles {
			if profile.IsActive && profile.CurrencyCode == adj.CurrencyCode &&
				profile.PayrollJurisdictionCode == period.JurisdictionCode &&
				strings.EqualFold(profile.EmployingLegalEntityReferenceID, period.EmployingLegalEntityReferenceID) &&
				compensationPeriodWithin(period.PeriodStart, &period.PeriodEnd, profile.EffectiveFrom, profile.EffectiveTo) {
				return nil
			}
		}
		if len(profiles) < domainrepo.MaxLimit {
			return hrerrors.ErrReferenceConflict
		}
		offset += len(profiles)
	}
}
