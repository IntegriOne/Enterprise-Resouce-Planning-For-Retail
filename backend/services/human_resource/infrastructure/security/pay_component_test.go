package security

import (
	"context"
	"errors"
	"testing"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
)

type payComponentStub struct {
	rows []entity.PayComponent
	err  error
}

func (s *payComponentStub) SelectWhere(context.Context, []domainrepo.Filter, int, int) ([]entity.PayComponent, error) {
	return s.rows, s.err
}

type payComponentReferenceStub struct {
	count int64
	err   error
}

func (s *payComponentReferenceStub) CountWhere(context.Context, []domainrepo.Filter) (int64, error) {
	return s.count, s.err
}

func TestPayComponentValidator(t *testing.T) {
	t.Parallel()
	components := &payComponentStub{}
	compensations := &payComponentReferenceStub{}
	adjustments := &payComponentReferenceStub{}
	results := &payComponentReferenceStub{}
	validator := &payComponentValidator{components: components, compensations: compensations,
		adjustments: adjustments, resultLines: results}
	currency := "IDR"
	component := &entity.PayComponent{ID: "component-1", ComponentCode: "BASE", ComponentVersion: 1,
		ComponentType: "earning", CalculationMethod: "fixed", CurrencyCode: &currency, JurisdictionCode: "ID"}
	if err := validator.ValidateMutation(t.Context(), "hr_pay_components", nil, component); err != nil {
		t.Fatalf("create fixed earning: %v", err)
	}
	formulaID := "formula-1"
	formula := *component
	formula.FormulaReferenceID = &formulaID
	if err := validator.ValidateMutation(t.Context(), "hr_pay_components", nil, &formula); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("unverified formula: %v", err)
	}
	statutory := *component
	statutory.CalculationMethod = "statutory_engine"
	if err := validator.ValidateMutation(t.Context(), "hr_pay_components", nil, &statutory); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("unverified statutory calculation: %v", err)
	}
	taxable := true
	taxClaim := *component
	taxClaim.Taxable = &taxable
	if err := validator.ValidateMutation(t.Context(), "hr_pay_components", nil, &taxClaim); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("unverified tax treatment: %v", err)
	}
	components.rows = []entity.PayComponent{{ID: "component-2"}}
	if err := validator.ValidateMutation(t.Context(), "hr_pay_components", nil, component); !errors.Is(err, hrerrors.ErrAlreadyExists) {
		t.Fatalf("duplicate version: %v", err)
	}
	components.rows = nil
	changed := *component
	changed.ComponentName = "Basic salary"
	for _, reference := range []*payComponentReferenceStub{compensations, adjustments, results} {
		reference.count = 1
		if err := validator.ValidateMutation(t.Context(), "hr_pay_components", component, &changed); !errors.Is(err, hrerrors.ErrReferenceConflict) {
			t.Fatalf("change referenced component: %v", err)
		}
		if err := validator.ValidateMutation(t.Context(), "hr_pay_components", component, nil); !errors.Is(err, hrerrors.ErrReferenceConflict) {
			t.Fatalf("delete referenced component: %v", err)
		}
		reference.count = 0
	}
	if err := validator.ValidateMutation(t.Context(), "hr_pay_components", component, component); err != nil {
		t.Fatalf("no-op update: %v", err)
	}
	results.err = hrerrors.ErrUnavailable
	if err := validator.ValidateMutation(t.Context(), "hr_pay_components", component, nil); !errors.Is(err, hrerrors.ErrUnavailable) {
		t.Fatalf("counter error: %v", err)
	}
	results.err = nil
	components.err = hrerrors.ErrUnavailable
	if err := validator.ValidateMutation(t.Context(), "hr_pay_components", nil, component); !errors.Is(err, hrerrors.ErrUnavailable) {
		t.Fatalf("lookup error: %v", err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if err := validator.ValidateMutation(ctx, "hr_pay_components", nil, component); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled context: %v", err)
	}
}
