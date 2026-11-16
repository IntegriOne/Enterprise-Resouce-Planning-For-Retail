// Package service memvalidasi rekonsiliasi hasil payroll tanpa tarif atau pembayaran eksternal.
package service

import (
	"strings"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
)

// ValidateResultTotals memeriksa satu hasil terhadap seluruh baris komponennya.
func ValidateResultTotals(result *entity.PayrollResult, lines []entity.PayrollResultLine) error {
	if err := result.Validate(); err != nil {
		return err
	}
	if len(lines) == 0 {
		return hrerrors.Invalid("lines", "payroll result requires component lines")
	}
	totals := map[string][]entity.Decimal{}
	ids := map[string]bool{}
	sequences := map[int32]bool{}
	for i := range lines {
		line := &lines[i]
		if err := line.Validate(); err != nil {
			return err
		}
		id := strings.ToLower(line.ID)
		if !strings.EqualFold(line.PayrollResultID, result.ID) || ids[id] || sequences[line.LineSequence] {
			return hrerrors.Invalid("lines", "result reference and line identities must be consistent")
		}
		ids[id] = true
		sequences[line.LineSequence] = true
		totals[line.ComponentTypeSnapshot] = append(totals[line.ComponentTypeSnapshot], line.Amount)
	}
	sum := func(kind string) (entity.Decimal, error) { return entity.DecimalSum(6, totals[kind]...) }
	gross, err := sum("earning")
	if err != nil {
		return err
	}
	deduction, err := sum("deduction")
	if err != nil {
		return err
	}
	tax, err := sum("tax")
	if err != nil {
		return err
	}
	deductions, err := entity.DecimalSum(6, deduction, tax)
	if err != nil {
		return err
	}
	reimbursement, err := sum("reimbursement")
	if err != nil {
		return err
	}
	employer, err := sum("employer_contribution")
	if err != nil {
		return err
	}
	cost, err := entity.DecimalSum(6, gross, employer)
	if err != nil {
		return err
	}
	for _, pair := range [][2]entity.Decimal{{gross, result.GrossAmount}, {deductions, result.TotalDeductions}, {reimbursement, result.ReimbursementAmount}, {cost, result.EmployerCostAmount}} {
		left, err := pair[0].Rat()
		if err != nil {
			return err
		}
		right, err := pair[1].Rat()
		if err != nil {
			return err
		}
		if left.Cmp(right) != 0 {
			return hrerrors.Invalid("totals", "payroll totals do not match component lines")
		}
	}
	return nil
}

// ValidateRunResults mengikat seluruh hasil ke run/group/currency yang sama, sebelum approval.
// Pemeriksaan actor, tarif lokal, snapshot input dan recheck Finance/IAM tetap pada use case.
func ValidateRunResults(run *entity.PayrollRun, results []entity.PayrollResult, lines []entity.PayrollResultLine) error {
	if err := run.Validate(); err != nil {
		return err
	}
	if len(results) == 0 {
		return hrerrors.Invalid("results", "payroll run requires results")
	}
	indexed := map[string][]entity.PayrollResultLine{}
	for _, line := range lines {
		key := strings.ToLower(line.PayrollResultID)
		indexed[key] = append(indexed[key], line)
	}
	ids := map[string]bool{}
	workers := map[string]bool{}
	for i := range results {
		result := &results[i]
		key := strings.ToLower(result.ID)
		worker := strings.ToLower(result.WorkerID)
		if ids[key] || workers[worker] || !strings.EqualFold(result.PayrollRunID, run.ID) || !strings.EqualFold(result.PayrollGroupID, run.PayrollGroupID) || result.CurrencyCode != run.CurrencyCode {
			return hrerrors.Invalid("results", "payroll run references must match without duplicate workers")
		}
		if err := ValidateResultTotals(result, indexed[key]); err != nil {
			return err
		}
		ids[key] = true
		workers[worker] = true
		delete(indexed, key)
	}
	if len(indexed) != 0 {
		return hrerrors.Invalid("lines", "orphan result lines are not allowed")
	}
	return nil
}
