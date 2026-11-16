package service_test

import (
	"testing"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	service "github.com/integrione/erp-retail/backend/services/human_resource/domain/service/payroll"
	f "github.com/integrione/erp-retail/backend/services/human_resource/domain/service/test/fixture"
)

func TestPayrollTotals(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		mutate func([]entity.PayrollResultLine)
		valid  bool
	}{
		{"exact fractions", func(_ []entity.PayrollResultLine) {}, true},
		{"wrong amount", func(lines []entity.PayrollResultLine) { lines[0].Amount = "100.123455" }, false},
		{"foreign result", func(lines []entity.PayrollResultLine) { lines[0].PayrollResultID = f.ID(999) }, false},
		{"duplicate sequence", func(lines []entity.PayrollResultLine) { lines[1].LineSequence = lines[0].LineSequence }, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := f.PayrollResult()
			lines := []entity.PayrollResultLine{f.Line(1, "earning", "100.123456"), f.Line(2, "tax", "10.000001"), f.Line(3, "employer_contribution", "5"), f.Line(4, "reimbursement", "2.000001")}
			tt.mutate(lines)
			err := service.ValidateResultTotals(&result, lines)
			if (err == nil) != tt.valid {
				t.Fatalf("totals: %v", err)
			}
		})
	}
}

func TestPayrollRunScope(t *testing.T) {
	t.Parallel()
	run := entity.PayrollRun{ID: f.ID(6), PayrollPeriodID: f.ID(10), PayrollGroupID: f.ID(7), EmployingLegalEntityReferenceID: f.ID(91), CurrencyCode: "IDR", RuleSetID: f.ID(11), SettlementDirection: "payable", RunNumber: 1, RunType: "regular", Status: entity.PayrollRunStatusDraft, JurisdictionCode: "ID", RuleEngineName: "engine", RuleEngineVersion: "v1", InputSnapshotReferenceID: f.ID(12), InputDigest: []byte{1}, StartedByPrincipalID: f.ID(99), CreatedAt: f.Date(), IdempotencyKey: "RUN"}
	result := f.PayrollResult()
	lines := []entity.PayrollResultLine{f.Line(1, "earning", "100.123456"), f.Line(2, "tax", "10.000001"), f.Line(3, "employer_contribution", "5"), f.Line(4, "reimbursement", "2.000001")}
	if err := service.ValidateRunResults(&run, []entity.PayrollResult{result}, lines); err != nil {
		t.Fatal(err)
	}
	result.CurrencyCode = "USD"
	if err := service.ValidateRunResults(&run, []entity.PayrollResult{result}, lines); err == nil {
		t.Fatal("cross-currency result accepted")
	}
	result = f.PayrollResult()
	orphan := f.Line(5, "earning", "1")
	orphan.PayrollResultID = f.ID(999)
	lines = append(lines, orphan)
	if err := service.ValidateRunResults(&run, []entity.PayrollResult{result}, lines); err == nil {
		t.Fatal("orphan line accepted")
	}
}
