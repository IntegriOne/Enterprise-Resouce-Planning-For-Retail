package service

import (
	"errors"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
)

func TestValidateFinalPayrollRunApprovalTime(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, time.November, 25, 10, 0, 0, 0, time.UTC)
	approver := "00000000-0000-4000-8000-000000000001"
	future := now.Add(time.Minute)
	run := &entity.PayrollRun{
		RunType: "termination", Status: entity.PayrollRunStatusApproved,
		ApprovedAt: &future, ApprovedByPrincipalID: &approver,
	}
	if err := validateFinalPayrollRun(run, now); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("approval after completion time: %v", err)
	}
	if err := validateFinalPayrollRun(run, future); err != nil {
		t.Fatalf("approval at completion time: %v", err)
	}
}

func TestValidateFinalPayrollResults(t *testing.T) {
	t.Parallel()
	for _, status := range []string{"calculated", "approved", "posted"} {
		if err := validateFinalPayrollResults(entity.PayrollRunStatusPaid, []entity.PayrollResult{{ResultStatus: status}}); err != nil {
			t.Errorf("valid payroll result %s: %v", status, err)
		}
	}
	paymentID := "00000000-0000-4000-8000-000000000001"
	paidResult := []entity.PayrollResult{{ResultStatus: "paid", PaymentReferenceID: &paymentID}}
	for _, status := range []entity.PayrollRunStatus{entity.PayrollRunStatusPosted, entity.PayrollRunStatusPaid} {
		if err := validateFinalPayrollResults(status, paidResult); err != nil {
			t.Errorf("valid paid result under run status %s: %v", status, err)
		}
	}
	if err := validateFinalPayrollResults(entity.PayrollRunStatusApproved, paidResult); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Errorf("paid result under approved run: %v", err)
	}
	for _, results := range [][]entity.PayrollResult{
		nil,
		{{ResultStatus: "cancelled"}},
		{{ResultStatus: "paid"}},
		{{ResultStatus: "superseded"}},
		{{ResultStatus: "calculated"}, {ResultStatus: "calculated"}},
	} {
		if err := validateFinalPayrollResults(entity.PayrollRunStatusPaid, results); !errors.Is(err, hrerrors.ErrReferenceConflict) {
			t.Errorf("invalid payroll results %+v: %v", results, err)
		}
	}
	if err := validateFinalPayrollResults(entity.PayrollRunStatusApproved, []entity.PayrollResult{{ResultStatus: "posted"}}); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Errorf("posted result under approved run: %v", err)
	}
}
