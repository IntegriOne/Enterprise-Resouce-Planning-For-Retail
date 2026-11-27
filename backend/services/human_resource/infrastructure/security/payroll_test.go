package security

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
)

type payrollFoundationStub struct {
	err error
}

func (s *payrollFoundationStub) VerifyPositionReferences(context.Context, PositionReferenceCheck) error {
	return s.err
}

func (s *payrollFoundationStub) VerifyWorkerReferences(context.Context, WorkerReferenceCheck) error {
	return s.err
}

func (s *payrollFoundationStub) VerifyLegalEntity(context.Context, string) error {
	return s.err
}

func (s *payrollFoundationStub) VerifyLegalEntityWindow(context.Context, string, time.Time, time.Time) error {
	return s.err
}

func (s *payrollFoundationStub) VerifyOrganizationUnit(context.Context, string, string) error {
	return s.err
}

func (s *payrollFoundationStub) VerifyLocation(context.Context, string, string) error {
	return s.err
}

func (s *payrollFoundationStub) VerifyParty(context.Context, string, bool) error {
	return s.err
}

func TestPayrollGroupPeriodProfileAdjustmentRuleSetValidators(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	legalEntityID := "55555555-5555-4555-8555-555555555501"
	groupID := "55555555-5555-4555-8555-555555555502"
	periodID := "55555555-5555-4555-8555-555555555503"
	workerID := "55555555-5555-4555-8555-555555555504"
	profileID := "55555555-5555-4555-8555-555555555505"
	componentID := "55555555-5555-4555-8555-555555555506"
	adjustmentID := "55555555-5555-4555-8555-555555555507"
	principalID := "55555555-5555-4555-8555-555555555508"

	groups := newPayrollStoreStub(map[string]entity.PayrollGroup{})
	periods := newPayrollStoreStub(map[string]entity.PayrollPeriod{})
	profiles := newPayrollStoreStub(map[string]entity.PayrollProfile{})
	runs := newPayrollStoreStub(map[string]entity.PayrollRun{})
	results := newPayrollStoreStub(map[string]entity.PayrollResult{})
	adjustments := newPayrollStoreStub(map[string]entity.PayrollAdjustment{})
	hireDate := now.AddDate(-1, 0, 0)
	workers := newPayrollStoreStub(map[string]entity.Worker{
		workerID: {ID: workerID, Status: entity.WorkerStatusActive, PrimaryLegalEntityReferenceID: legalEntityID, HireDate: &hireDate},
	})
	cur := "IDR"
	components := newPayrollStoreStub(map[string]entity.PayComponent{
		componentID: {
			ID: componentID, ComponentCode: "ALLOW-1", ComponentVersion: 1, ComponentType: "earning",
			CalculationMethod: "fixed", JurisdictionCode: "ID", CurrencyCode: &cur,
			EffectiveFrom: hireDate, IsActive: true,
		},
	})
	foundation := &payrollFoundationStub{}

	ruleSetVal := &payrollRuleSetValidator{}
	if err := ruleSetVal.ValidateMutation(t.Context(), "hr_payroll_rule_sets", nil, &entity.PayrollRuleSet{ID: "55555555-5555-4555-8555-555555555599"}); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("immutable rule set = %v, want ErrWorkflowRequired", err)
	}

	groupVal := &payrollGroupValidator{groups: groups, profiles: profiles, periods: periods, runs: runs, foundation: foundation}
	group := &entity.PayrollGroup{
		ID: groupID, GroupCode: "MONTHLY-ID", GroupName: "Monthly Indonesia",
		EmployingLegalEntityReferenceID: legalEntityID, JurisdictionCode: "ID", CurrencyCode: "IDR",
		CurrencyScale: 2, PayFrequency: "monthly", Timezone: "Asia/Jakarta", IsActive: true,
	}
	if err := groupVal.ValidateMutation(t.Context(), "hr_payroll_groups", nil, group); err != nil {
		t.Fatalf("valid group failed: %v", err)
	}
	foundation.err = hrerrors.ErrUnavailable
	if err := groupVal.ValidateMutation(t.Context(), "hr_payroll_groups", nil, group); !errors.Is(err, hrerrors.ErrUnavailable) {
		t.Fatalf("unavailable EF group = %v, want ErrUnavailable", err)
	}
	foundation.err = nil
	groups.rows[groupID] = *group

	periodVal := &payrollPeriodValidator{groups: groups, periods: periods, runs: runs, adjustments: adjustments, foundation: foundation}
	period := &entity.PayrollPeriod{
		ID: periodID, PayrollGroupID: groupID, EmployingLegalEntityReferenceID: legalEntityID,
		CurrencyCode: "IDR", JurisdictionCode: "ID", PeriodCode: "2026-11",
		PeriodStart: now, PeriodEnd: now.AddDate(0, 0, 29), PayDate: now.AddDate(0, 0, 29), PeriodStatus: "open",
	}
	if err := periodVal.ValidateMutation(t.Context(), "hr_payroll_periods", nil, period); err != nil {
		t.Fatalf("valid period failed: %v", err)
	}
	changedScale := *group
	changedScale.CurrencyScale++
	periods.rows[periodID] = *period
	if err := groupVal.ValidateMutation(t.Context(), "hr_payroll_groups", group, &changedScale); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("change currency scale of referenced group = %v, want ErrReferenceConflict", err)
	}
	delete(periods.rows, periodID)

	rescheduled := *period
	rescheduled.PayDate = period.PayDate.AddDate(0, 0, 1)
	if err := periodVal.ValidateMutation(t.Context(), "hr_payroll_periods", period, &rescheduled); err != nil {
		t.Fatalf("reschedule unreferenced payroll period: %v", err)
	}
	runs.rows["scheduled-run"] = entity.PayrollRun{PayrollPeriodID: periodID}
	if err := periodVal.ValidateMutation(t.Context(), "hr_payroll_periods", period, &rescheduled); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("reschedule referenced payroll period = %v, want ErrReferenceConflict", err)
	}
	delete(runs.rows, "scheduled-run")

	calendarRef := "55555555-5555-4555-8555-555555555598"
	periodWithCal := *period
	periodWithCal.CalendarReferenceID = &calendarRef
	if err := periodVal.ValidateMutation(t.Context(), "hr_payroll_periods", nil, &periodWithCal); !errors.Is(err, hrerrors.ErrUnavailable) {
		t.Fatalf("external calendar ref = %v, want ErrUnavailable", err)
	}
	periods.rows[periodID] = *period

	profileVal := &payrollProfileValidator{workers: workers, groups: groups, profiles: profiles, results: results, foundation: foundation}
	profile := &entity.PayrollProfile{
		ID: profileID, WorkerID: workerID, PayrollCountryCode: "ID", PayrollJurisdictionCode: "ID",
		PayrollGroupID: groupID, EmployingLegalEntityReferenceID: legalEntityID, CurrencyCode: "IDR",
		EffectiveFrom: hireDate, IsActive: true,
	}
	if err := profileVal.ValidateMutation(t.Context(), "hr_payroll_profiles", nil, profile); err != nil {
		t.Fatalf("valid profile failed: %v", err)
	}
	taxRef := "55555555-5555-4555-8555-555555555597"
	profileWithTax := *profile
	profileWithTax.TaxProfileReferenceID = &taxRef
	if err := profileVal.ValidateMutation(t.Context(), "hr_payroll_profiles", nil, &profileWithTax); !errors.Is(err, hrerrors.ErrUnavailable) {
		t.Fatalf("unverified tax vault ref = %v, want ErrUnavailable", err)
	}
	overlapping := *profile
	overlapping.ID = "55555555-5555-4555-8555-555555555599"
	pagedProfileVal := *profileVal
	pagedProfileVal.profiles = &payrollProfilePages{current: overlapping}
	if err := pagedProfileVal.ValidateMutation(t.Context(), "hr_payroll_profiles", nil, profile); !errors.Is(err, hrerrors.ErrAlreadyExists) {
		t.Fatalf("overlapping payroll profile after first page = %v, want ErrAlreadyExists", err)
	}

	profiles.rows[profileID] = *profile

	adjVal := &payrollAdjustmentValidator{workers: workers, periods: periods, profiles: profiles, components: components, adjustments: adjustments}
	adj := &entity.PayrollAdjustment{
		ID: adjustmentID, WorkerID: workerID, PayrollPeriodID: periodID, ComponentID: componentID,
		AdjustmentType: "bonus", Amount: "500000.000000", CurrencyCode: "IDR",
		Reason: "Shift allowance", RequestedByPrincipalID: principalID, IdempotencyKey: "adj-1", CreatedAt: now,
	}
	if err := adjVal.ValidateMutation(t.Context(), "hr_payroll_adjustments", nil, adj); err != nil {
		t.Fatalf("valid adjustment failed: %v", err)
	}
	worker := workers.rows[workerID]
	futureHire := period.PeriodEnd.AddDate(0, 0, 1)
	worker.HireDate = &futureHire
	workers.rows[workerID] = worker
	if err := adjVal.ValidateMutation(t.Context(), "hr_payroll_adjustments", nil, adj); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("adjustment before employment = %v, want ErrReferenceConflict", err)
	}
	worker.HireDate = &hireDate
	workers.rows[workerID] = worker
	wrongJurisdiction := *profile
	wrongJurisdiction.PayrollJurisdictionCode = "OTHER"
	profiles.rows[profileID] = wrongJurisdiction
	if err := adjVal.ValidateMutation(t.Context(), "hr_payroll_adjustments", nil, adj); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("adjustment with mismatched profile jurisdiction = %v, want ErrReferenceConflict", err)
	}
}

func TestPayrollRunDraftAndApprovalValidator(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	legalEntityID := "66666666-6666-4666-8666-666666666601"
	groupID := "66666666-6666-4666-8666-666666666602"
	periodID := "66666666-6666-4666-8666-666666666603"
	ruleSetID := "66666666-6666-4666-8666-666666666604"
	runID := "66666666-6666-4666-8666-666666666605"
	starterID := "66666666-6666-4666-8666-666666666606"
	approverID := "66666666-6666-4666-8666-666666666607"

	groups := newPayrollStoreStub(map[string]entity.PayrollGroup{
		groupID: {ID: groupID, EmployingLegalEntityReferenceID: legalEntityID, JurisdictionCode: "ID", CurrencyCode: "IDR", IsActive: true},
	})
	periods := newPayrollStoreStub(map[string]entity.PayrollPeriod{
		periodID: {
			ID: periodID, PayrollGroupID: groupID, EmployingLegalEntityReferenceID: legalEntityID,
			CurrencyCode: "IDR", JurisdictionCode: "ID", PeriodStart: now, PeriodEnd: now.AddDate(0, 0, 29),
			PayDate: now.AddDate(0, 0, 29), PeriodStatus: "open",
		},
	})
	ruleSets := newPayrollStoreStub(map[string]entity.PayrollRuleSet{
		ruleSetID: {ID: ruleSetID, JurisdictionCode: "ID", EffectiveFrom: now.AddDate(-1, 0, 0), ApprovedAt: now},
	})
	runs := newPayrollStoreStub(map[string]entity.PayrollRun{})
	results := newPayrollStoreStub(map[string]entity.PayrollResult{})
	runVal := &payrollRunValidator{
		periods: periods, groups: groups, ruleSets: ruleSets, runs: runs, results: results,
		foundation: &payrollFoundationStub{},
	}

	draftRun := &entity.PayrollRun{
		ID: runID, PayrollPeriodID: periodID, PayrollGroupID: groupID,
		EmployingLegalEntityReferenceID: legalEntityID, CurrencyCode: "IDR", RuleSetID: ruleSetID,
		SettlementDirection: "payable", RunNumber: 1, RunType: "regular", Status: entity.PayrollRunStatusDraft,
		JurisdictionCode: "ID", RuleEngineName: "local", RuleEngineVersion: "1",
		InputSnapshotReferenceID: "66666666-6666-4666-8666-666666666608", InputDigest: []byte{1},
		StartedByPrincipalID: starterID, CreatedAt: now, IdempotencyKey: "run-1",
	}
	if err := runVal.ValidateMutation(t.Context(), "hr_payroll_runs", nil, draftRun); err != nil {
		t.Fatalf("valid draft run failed: %v", err)
	}

	ruleSet := ruleSets.rows[ruleSetID]
	ruleSet.ApprovedAt = now.Add(time.Second)
	ruleSets.rows[ruleSetID] = ruleSet
	if err := runVal.ValidateMutation(t.Context(), "hr_payroll_runs", nil, draftRun); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("create run before rule set approval: %v", err)
	}
	ruleSet.ApprovedAt = now
	ruleSets.rows[ruleSetID] = ruleSet

	calcAt := now.Add(time.Hour)
	calculated := *draftRun
	calculated.Status = entity.PayrollRunStatusCalculated
	calculated.CalculatedAt = &calcAt
	calculated.OutputDigest = []byte{2}

	approveAt := calcAt.Add(time.Minute)
	review := calculated
	review.Status = entity.PayrollRunStatusReview
	approved := review
	approved.Status = entity.PayrollRunStatusApproved
	approved.ApprovedByPrincipalID = &approverID
	approved.ApprovedAt = &approveAt

	ruleSet.ApprovedAt = calcAt.Add(time.Second)
	ruleSets.rows[ruleSetID] = ruleSet
	if err := runVal.ValidateMutation(t.Context(), "hr_payroll_runs", &calculated, &review); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("review calculation from before rule approval: %v", err)
	}
	if err := runVal.ValidateMutation(t.Context(), "hr_payroll_runs", &review, &approved); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("approve calculation from before rule approval: %v", err)
	}
	posting := approved
	posting.PostingRequestedAt = &approveAt
	posting.PostingRequestedByPrincipalID = &approverID
	if err := runVal.ValidateMutation(t.Context(), "hr_payroll_runs", &approved, &posting); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("request posting for a calculation from before rule approval: %v", err)
	}
	ruleSet.ApprovedAt = now
	ruleSets.rows[ruleSetID] = ruleSet
	if err := runVal.ValidateMutation(t.Context(), "hr_payroll_runs", &approved, &posting); err != nil {
		t.Fatalf("request posting for calculation under an approved rule set: %v", err)
	}

	// Approval run membutuhkan minimal satu payroll result dan pemisahan tugas.
	if err := runVal.ValidateMutation(t.Context(), "hr_payroll_runs", &review, &approved); !errors.Is(err, hrerrors.ErrInvalidInput) {
		t.Fatalf("approval without results = %v, want ErrInvalidInput", err)
	}
	results.rows["66666666-6666-4666-8666-666666666609"] = entity.PayrollResult{ID: "66666666-6666-4666-8666-666666666609", PayrollRunID: runID}
	if err := runVal.ValidateMutation(t.Context(), "hr_payroll_runs", &review, &approved); err != nil {
		t.Fatalf("valid approval transition failed: %v", err)
	}

	period := periods.rows[periodID]
	period.PeriodStatus = "closed"
	periods.rows[periodID] = period
	if err := runVal.ValidateMutation(t.Context(), "hr_payroll_runs", &review, &approved); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("approve payroll after period closure: %v", err)
	}
	period.PeriodStatus = "open"
	periods.rows[periodID] = period

	selfApproved := approved
	selfApproved.ApprovedByPrincipalID = &starterID
	if err := runVal.ValidateMutation(t.Context(), "hr_payroll_runs", &review, &selfApproved); !errors.Is(err, hrerrors.ErrForbidden) {
		t.Fatalf("self approval = %v, want ErrForbidden", err)
	}
	if err := runVal.ValidateMutation(t.Context(), "hr_payroll_runs", &approved, nil); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("deleting approved run = %v, want ErrWorkflowRequired", err)
	}
}

func newPayrollStoreStub[T any](initial map[string]T) *payrollStoreStub[T] {
	rows := make(map[string]T, len(initial))
	for k, v := range initial {
		rows[k] = v
	}
	return &payrollStoreStub[T]{rows: rows}
}

type payrollStoreStub[T any] struct {
	rows map[string]T
}

func (s *payrollStoreStub[T]) FindByIDForUpdate(_ context.Context, id string) (*T, error) {
	v, ok := s.rows[id]
	if !ok {
		return nil, hrerrors.ErrNotFound
	}
	copyVal := v
	return &copyVal, nil
}

func (s *payrollStoreStub[T]) SelectWhere(_ context.Context, _ []domainrepo.Filter, _ int, _ int) ([]T, error) {
	out := make([]T, 0, len(s.rows))
	for _, v := range s.rows {
		out = append(out, v)
	}
	return out, nil
}

func (s *payrollStoreStub[T]) CountWhere(_ context.Context, _ []domainrepo.Filter) (int64, error) {
	return int64(len(s.rows)), nil
}

type payrollProfilePages struct {
	payrollStoreStub[entity.PayrollProfile]
	current entity.PayrollProfile
}

func (s *payrollProfilePages) SelectWhere(_ context.Context, _ []domainrepo.Filter, limit, offset int) ([]entity.PayrollProfile, error) {
	if offset == 0 {
		historical := s.current
		historical.EffectiveFrom = historical.EffectiveFrom.AddDate(-1, 0, 0)
		historical.EffectiveTo = &historical.EffectiveFrom
		page := make([]entity.PayrollProfile, limit)
		for i := range page {
			page[i] = historical
		}
		return page, nil
	}
	return []entity.PayrollProfile{s.current}, nil
}

func TestPayrollAdjustmentFindsProfileAfterFirstPage(t *testing.T) {
	t.Parallel()
	start := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	period := &entity.PayrollPeriod{
		PeriodStart: start, PeriodEnd: start.AddDate(0, 1, -1),
		JurisdictionCode: "ID", EmployingLegalEntityReferenceID: "employer",
	}
	profiles := &payrollProfilePages{current: entity.PayrollProfile{
		IsActive: true, CurrencyCode: "IDR", EffectiveFrom: start.AddDate(0, -1, 0),
		PayrollJurisdictionCode: "ID", EmployingLegalEntityReferenceID: "employer",
	}}
	validator := &payrollAdjustmentValidator{profiles: profiles}
	if err := validator.requireActiveProfile(t.Context(), &entity.PayrollAdjustment{CurrencyCode: "IDR"}, period); err != nil {
		t.Fatalf("valid profile on next page was rejected: %v", err)
	}
}
