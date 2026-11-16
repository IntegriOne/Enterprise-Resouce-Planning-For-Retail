package entity_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
)

func TestEntityValidation(t *testing.T) {
	t.Parallel()
	good := entity.JobFamily{ID: "00000000-0000-0000-0000-000000000001", FamilyCode: "RETAIL", FamilyName: "Retail", JobDomain: "retail"}
	tests := []struct {
		name  string
		value entity.JobFamily
		valid bool
	}{
		{"valid", good, true},
		{"missing id", func() entity.JobFamily { x := good; x.ID = ""; return x }(), false},
		{"nil uuid", func() entity.JobFamily { x := good; x.ID = "00000000-0000-0000-0000-000000000000"; return x }(), false},
		{"whitespace code", func() entity.JobFamily { x := good; x.FamilyCode = " "; return x }(), false},
		{"unicode length", func() entity.JobFamily { x := good; x.FamilyCode = strings.Repeat("界", 80); return x }(), true},
		{"overflow length", func() entity.JobFamily { x := good; x.FamilyCode = strings.Repeat("界", 81); return x }(), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.value.Validate()
			if (err == nil) != tt.valid {
				t.Fatalf("Validate()=%v", err)
			}
			if err != nil && !errors.Is(err, hrerrors.ErrInvalidInput) {
				t.Fatalf("unclassified: %v", err)
			}
		})
	}
	var absent *entity.Worker
	if !errors.Is(absent.Validate(), hrerrors.ErrInvalidInput) {
		t.Fatal("nil receiver must fail closed")
	}
}

func TestDomainConstraints(t *testing.T) {
	t.Parallel()
	id := "00000000-0000-0000-0000-000000000001"
	start := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 0, -1)
	assignment := entity.WorkerAssignment{ID: id, WorkerID: id, PositionID: id, AssignmentType: "primary", FTEPercent: "100", ValidFrom: start}
	if err := assignment.Validate(); err != nil {
		t.Fatal(err)
	}
	assignment.ValidTo = &end
	if err := assignment.Validate(); err == nil {
		t.Fatal("backward assignment range accepted")
	}
	assignment.ValidTo = nil
	assignment.FTEPercent = "100.001"
	if err := assignment.Validate(); err == nil {
		t.Fatal("FTE overflow accepted")
	}
	assignment.FTEPercent = "0.0001"
	if err := assignment.Validate(); err == nil {
		t.Fatal("numeric scale would be rounded silently")
	}
	assignment.FTEPercent = "50"
	assignment.ManagerWorkerID = &id
	if err := assignment.Validate(); err == nil {
		t.Fatal("self manager accepted")
	}
}

func TestEvidenceAndNumericConstraints(t *testing.T) {
	t.Parallel()
	at := time.Date(2026, 11, 16, 0, 0, 0, 0, time.UTC)
	id := "00000000-0000-0000-0000-000000000001"
	run := entity.PayrollRun{ID: id, PayrollPeriodID: id, PayrollGroupID: id, EmployingLegalEntityReferenceID: id, CurrencyCode: "IDR", RuleSetID: id, SettlementDirection: "payable", RunNumber: 1, RunType: "regular", Status: entity.PayrollRunStatusDraft, JurisdictionCode: "ID", RuleEngineName: "engine", RuleEngineVersion: "v1", InputSnapshotReferenceID: id, InputDigest: []byte{1}, StartedByPrincipalID: id, CreatedAt: at, IdempotencyKey: "run"}
	cases := []struct {
		name   string
		mutate func(*entity.PayrollRun)
		rule   string
	}{
		{"correction requires original", func(r *entity.PayrollRun) { r.RunType = "correction" }, "ck_hr_payroll_correction_origin"},
		{"original forbidden on regular", func(r *entity.PayrollRun) { other := "00000000-0000-0000-0000-000000000002"; r.OriginalRunID = &other }, "ck_hr_payroll_correction_origin"},
		{"calculation evidence missing", func(r *entity.PayrollRun) { r.Status = entity.PayrollRunStatusCalculated }, "ck_hr_payroll_run_calculation_state"},
	}
	if err := run.Validate(); err != nil {
		t.Fatal(err)
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			r := run
			tt.mutate(&r)
			err := r.Validate()
			var invalid *hrerrors.ValidationError
			if !errors.As(err, &invalid) || invalid.Rule != tt.rule {
				t.Fatalf("wanted rule %s: %v", tt.rule, err)
			}
		})
	}
	invalidStatus := run
	invalidStatus.Status = "unknown"
	if err := invalidStatus.Validate(); err == nil {
		t.Fatal("invalid enum accepted")
	}
	for _, tt := range []struct {
		raw   entity.Decimal
		valid bool
	}{{"99999999999999.123456", true}, {"100000000000000.000000", false}, {"0.1234567", false}, {"0.123456000", true}} {
		err := tt.raw.ValidatePrecision(20, 6)
		if (err == nil) != tt.valid {
			t.Fatalf("precision %s: %v", tt.raw, err)
		}
	}
}
