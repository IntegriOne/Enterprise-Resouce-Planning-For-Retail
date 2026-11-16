package service_test

import (
	"errors"
	"testing"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	"github.com/integrione/erp-retail/backend/services/human_resource/domain/service/test/fixture"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

func calculatedRun() entity.PayrollRun {
	now := fixture.Date()
	return entity.PayrollRun{
		ID: fixture.ID(6), PayrollPeriodID: fixture.ID(60), PayrollGroupID: fixture.ID(7),
		EmployingLegalEntityReferenceID: fixture.ID(91), CurrencyCode: "IDR", RuleSetID: fixture.ID(61),
		SettlementDirection: "payable", RunNumber: 1, RunType: "regular",
		Status: entity.PayrollRunStatusCalculated, JurisdictionCode: "ID", RuleEngineName: "fixture",
		RuleEngineVersion: "1", InputSnapshotReferenceID: fixture.ID(62), InputDigest: []byte{1},
		OutputDigest: []byte{2}, CalculatedAt: &now, StartedByPrincipalID: fixture.ID(99),
		CreatedAt: now, IdempotencyKey: "approval-fixture",
	}
}

func calculatedLines() []entity.PayrollResultLine {
	return []entity.PayrollResultLine{
		fixture.Line(1, "earning", "100.123456"),
		fixture.Line(2, "reimbursement", "2.000001"),
		fixture.Line(3, "deduction", "10.000001"),
		fixture.Line(4, "employer_contribution", "5"),
	}
}

func TestApprovePayrollRun(t *testing.T) {
	sentinel := errors.New("outbox unavailable")
	for _, name := range []string{"approved", "self_approval", "invalid_totals", "hidden_results", "outbox_rollback"} {
		t.Run(name, func(t *testing.T) {
			h, svc, _, runs, results, lines := setupPayroll(t)
			run := calculatedRun()
			result := fixture.PayrollResult()
			var want error
			switch name {
			case "self_approval":
				run.StartedByPrincipalID = h.Actor.PrincipalID
				want = hrerrors.ErrForbidden
			case "invalid_totals":
				result.GrossAmount = "101.123456"
				result.EmployerCostAmount = "106.123456"
				result.NetPayAmount = "93.123456"
				want = hrerrors.ErrInvalidInput
			case "hidden_results":
				h.Scope = operation.Scope{Filters: []domainrepo.Filter{{Column: "id", Values: []any{fixture.ID(999)}}}}
				want = hrerrors.ErrForbidden
			case "outbox_rollback":
				h.Failures["outbox"] = sentinel
				want = sentinel
			}
			runs.Rows[run.ID] = run
			results.Rows[result.ID] = result
			for _, line := range calculatedLines() {
				lines.Rows[line.ID] = line
			}
			got, err := svc.ApprovePayrollRun(t.Context(), run.ID)
			if !errors.Is(err, want) {
				t.Fatalf("approval error = %v, want %v", err, want)
			}
			stored := runs.Rows[run.ID]
			if want != nil {
				if got != nil || stored.Status != run.Status || stored.ApprovedAt != nil || len(h.Events) != 0 {
					t.Fatal("rejected approval persisted state or event")
				}
				return
			}
			if got == nil || stored.Status != entity.PayrollRunStatusApproved || stored.ApprovedAt == nil ||
				stored.ApprovedByPrincipalID == nil || *stored.ApprovedByPrincipalID != h.Actor.PrincipalID || len(h.Events) != 1 {
				t.Fatal("approval evidence and outbox must commit together")
			}
			if _, err := svc.ApprovePayrollRun(t.Context(), run.ID); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
				t.Fatalf("repeated approval = %v", err)
			}
		})
	}
}
