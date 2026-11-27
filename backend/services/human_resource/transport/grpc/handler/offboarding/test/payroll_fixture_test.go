//go:build integration

package test

import (
	"context"
	"strings"
	"testing"

	"gorm.io/gorm"
)

func seedTerminationPayroll(t *testing.T, ctx context.Context, db *gorm.DB, id string) {
	t.Helper()
	fixtures := []string{
		`INSERT INTO hr_payroll_groups (id, group_code, group_name, employing_legal_entity_reference_id, jurisdiction_code, currency_code, currency_scale, pay_frequency, timezone)
		VALUES (@id, 'EXIT', 'Exit payroll', @id, 'ID', 'IDR', 0, 'monthly', 'UTC')`,
		`INSERT INTO hr_payroll_periods (id, payroll_group_id, employing_legal_entity_reference_id, currency_code, jurisdiction_code, period_code, period_start, period_end, pay_date)
		VALUES (@id, @id, @id, 'IDR', 'ID', 'EXIT', CURRENT_DATE, CURRENT_DATE, CURRENT_DATE)`,
		`INSERT INTO hr_payroll_profiles (id, worker_id, payroll_country_code, payroll_jurisdiction_code, payroll_group_id, employing_legal_entity_reference_id, currency_code, effective_from)
		VALUES (@id, @id, 'ID', 'ID', @id, @id, 'IDR', CURRENT_DATE)`,
		`INSERT INTO hr_payroll_rule_sets (id, rule_code, rule_version, jurisdiction_code, effective_from, artifact_document_version_reference_id, artifact_digest, legal_basis_document_version_reference_id, approved_by_principal_id, approved_at)
		VALUES (@id, CAST(@id AS text), 'v1', 'ID', CURRENT_DATE, @id, decode(repeat('12', 32), 'hex'), @id, @id, NOW())`,
		`INSERT INTO hr_pay_components (id, component_code, component_version, component_name, component_type, calculation_method, jurisdiction_code, effective_from)
		VALUES (@id, CAST(@id AS text), 1, 'Final pay', 'earning', 'fixed', 'ID', CURRENT_DATE)`,
		`INSERT INTO hr_payroll_runs (id, payroll_period_id, payroll_group_id, employing_legal_entity_reference_id, currency_code, rule_set_id, run_number, run_type, jurisdiction_code,
		 rule_engine_name, rule_engine_version, input_snapshot_reference_id, input_digest, started_by_principal_id, idempotency_key)
		VALUES (@id, @id, @id, @id, 'IDR', @id, 1, 'termination', 'ID', 'test', '1', @id, decode(repeat('12', 32), 'hex'), @id, CAST(@id AS text))`,
		`INSERT INTO hr_payroll_results (id, payroll_run_id, payroll_group_id, worker_id, payroll_profile_reference_id, currency_code, gross_amount, total_deductions, employer_cost_amount, net_pay_amount, calculation_digest)
		VALUES (@id, @id, @id, @id, @id, 'IDR', 100, 0, 100, 100, decode(repeat('12', 32), 'hex'))`,
		`INSERT INTO hr_payroll_result_lines (id, payroll_result_id, component_id, component_code_snapshot, component_name_snapshot, component_type_snapshot, amount, line_sequence)
		VALUES (@id, @id, @id, CAST(@id AS text), 'Final pay', 'earning', 100, 1)`,
		`UPDATE hr_payroll_runs SET status = 'calculated', calculated_at = NOW(),
		 output_digest = decode(repeat('12', 32), 'hex') WHERE id = @id`,
		`UPDATE hr_payroll_runs SET status = 'review' WHERE id = @id`,
		`UPDATE hr_payroll_runs SET status = 'approved', approved_by_principal_id = @id, approved_at = NOW() WHERE id = @id`,
	}
	if err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, statement := range fixtures {
			if err := tx.Exec(statement, map[string]any{"id": id}).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		t.Fatalf("seed termination payroll: %v", err)
	}
}

func cleanupTerminationPayroll(t *testing.T, db *gorm.DB, id string) {
	t.Helper()
	if err := db.Transaction(func(tx *gorm.DB) error {
		guards := []string{
			"hr_payroll_result_lines DISABLE TRIGGER tr_hr_payroll_line_guard",
			"hr_payroll_results DISABLE TRIGGER tr_hr_payroll_result_guard",
			"hr_payroll_runs DISABLE TRIGGER tr_hr_payroll_run_guard",
			"hr_payroll_rule_sets DISABLE TRIGGER tr_hr_payroll_rule_immutable",
		}
		for _, guard := range guards {
			if err := tx.Exec("ALTER TABLE " + guard).Error; err != nil {
				return err
			}
		}
		for _, table := range []string{
			"hr_payroll_result_lines", "hr_payroll_results", "hr_payroll_runs", "hr_pay_components",
			"hr_payroll_rule_sets", "hr_payroll_profiles", "hr_payroll_periods", "hr_payroll_groups",
		} {
			if err := tx.Exec("DELETE FROM "+table+" WHERE id = ?", id).Error; err != nil {
				return err
			}
		}
		for _, guard := range guards {
			if err := tx.Exec("ALTER TABLE " + strings.Replace(guard, "DISABLE", "ENABLE", 1)).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		t.Errorf("clean termination payroll: %v", err)
	}
}

// setTerminationResultFixtureStatus hanya dipakai pada database disposable yang memiliki fixture payroll.
func setTerminationResultFixtureStatus(t *testing.T, ctx context.Context, db *gorm.DB, id, status string) {
	t.Helper()
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("ALTER TABLE hr_payroll_results DISABLE TRIGGER tr_hr_payroll_result_guard").Error; err != nil {
			return err
		}
		if err := tx.Exec("UPDATE hr_payroll_results SET result_status = ? WHERE id = ?", status, id).Error; err != nil {
			return err
		}
		return tx.Exec("ALTER TABLE hr_payroll_results ENABLE TRIGGER tr_hr_payroll_result_guard").Error
	})
	if err != nil {
		t.Fatalf("set termination result fixture status: %v", err)
	}
}
