package service_test

import (
	"crypto/sha256"
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
		RuleEngineVersion: "1", InputSnapshotReferenceID: fixture.ID(62), InputDigest: make([]byte, sha256.Size),
		OutputDigest: make([]byte, sha256.Size), CalculatedAt: &now, StartedByPrincipalID: fixture.ID(99),
		CreatedAt: now, IdempotencyKey: "approval-fixture",
	}
}

func calculatedResult() entity.PayrollResult {
	result := fixture.PayrollResult()
	result.CalculationDigest = make([]byte, sha256.Size)
	return result
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
	for _, name := range []string{"approved", "approved_multiple_batches", "workload_approval", "self_approval", "invalid_totals", "excess_settlement_precision", "invalid_output_digest", "invalid_result_digest", "cancelled_result", "paid_result", "hidden_results", "outbox_rollback"} {
		t.Run(name, func(t *testing.T) {
			scale := int16(6)
			if name == "excess_settlement_precision" {
				scale = 2
			}
			h, svc, _, runs, results, lines := setupPayroll(t, scale)
			run := calculatedRun()
			run.Status = entity.PayrollRunStatusReview
			result := calculatedResult()
			var want error
			switch name {
			case "workload_approval":
				h.Actor.Kind = "workload"
				want = hrerrors.ErrForbidden
			case "self_approval":
				run.StartedByPrincipalID = h.Actor.PrincipalID
				want = hrerrors.ErrForbidden
			case "excess_settlement_precision":
				want = hrerrors.ErrReferenceConflict
			case "invalid_output_digest":
				run.OutputDigest = []byte{2}
				want = hrerrors.ErrReferenceConflict
			case "cancelled_result":
				result.ResultStatus = "cancelled"
				want = hrerrors.ErrReferenceConflict
			case "paid_result":
				result.ResultStatus = "paid"
				want = hrerrors.ErrReferenceConflict
			case "invalid_result_digest":
				result.CalculationDigest = []byte{1}
				want = hrerrors.ErrReferenceConflict
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
			if name == "approved_multiple_batches" {
				for batchIndex := range domainrepo.MaxLimit {
					additional := calculatedResult()
					additional.ID = fixture.ID(1000 + batchIndex)
					additional.WorkerID = fixture.ID(2000 + batchIndex)
					results.Rows[additional.ID] = additional
					for lineIndex, line := range calculatedLines() {
						line.ID = fixture.ID(3000 + 10*batchIndex + lineIndex)
						line.PayrollResultID = additional.ID
						lines.Rows[line.ID] = line
					}
				}
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

func TestRequestPayrollPostingRejectsLegacySettlementPrecision(t *testing.T) {
	h, svc, _, runs, results, lines := setupPayroll(t, 2)
	run := calculatedRun()
	run.Status = entity.PayrollRunStatusApproved
	run.ApprovedAt = run.CalculatedAt
	run.ApprovedByPrincipalID = new(string(h.Actor.PrincipalID))
	runs.Rows[run.ID] = run
	result := calculatedResult()
	results.Rows[result.ID] = result
	for _, line := range calculatedLines() {
		lines.Rows[line.ID] = line
	}

	got, err := svc.RequestPayrollPosting(t.Context(), run.ID)
	if got != nil || !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("posting legacy result with excess settlement precision: %v, %v", got, err)
	}
	stored := runs.Rows[run.ID]
	if stored.PostingRequestedAt != nil || stored.PostingRequestedByPrincipalID != nil || len(h.Events) != 0 {
		t.Fatal("rejected posting persisted state or an outbox event")
	}
}
