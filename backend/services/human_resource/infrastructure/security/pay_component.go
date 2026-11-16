package security

import (
	"context"
	"reflect"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository"
)

type payComponentReader interface {
	SelectWhere(context.Context, []domainrepo.Filter, int, int) ([]entity.PayComponent, error)
}

type payComponentReferenceCounter interface {
	CountWhere(context.Context, []domainrepo.Filter) (int64, error)
}

type payComponentValidator struct {
	components    payComponentReader
	compensations payComponentReferenceCounter
	adjustments   payComponentReferenceCounter
	resultLines   payComponentReferenceCounter
}

// PayComponentValidators memasang aturan katalog komponen yang dimiliki HRM.
func PayComponentValidators(repos *repository.Registry) (map[string]MutationValidator, error) {
	if repos == nil || operation.RequireDependency(repos.PayComponent) != nil ||
		operation.RequireDependency(repos.CompensationRecord) != nil ||
		operation.RequireDependency(repos.PayrollAdjustment) != nil ||
		operation.RequireDependency(repos.PayrollResultLine) != nil {
		return nil, hrerrors.ErrUnavailable
	}
	return map[string]MutationValidator{
		"hr_pay_components": &payComponentValidator{
			components: repos.PayComponent, compensations: repos.CompensationRecord,
			adjustments: repos.PayrollAdjustment, resultLines: repos.PayrollResultLine,
		},
	}, nil
}

// ValidateMutation hanya mengizinkan komponen earning tetap tanpa klaim aturan pajak atau formula.
func (v *payComponentValidator) ValidateMutation(ctx context.Context, table string, before, after entity.Validatable) error {
	if v == nil || table != "hr_pay_components" || operation.RequireDependency(v.components) != nil ||
		operation.RequireDependency(v.compensations) != nil || operation.RequireDependency(v.adjustments) != nil ||
		operation.RequireDependency(v.resultLines) != nil {
		return hrerrors.ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	var previous *entity.PayComponent
	if before != nil {
		var ok bool
		previous, ok = before.(*entity.PayComponent)
		if !ok || previous == nil {
			return hrerrors.ErrInvalidInput
		}
		if !isLocalFixedEarning(previous) {
			return hrerrors.ErrWorkflowRequired
		}
	}
	if after == nil {
		if previous == nil {
			return hrerrors.ErrInvalidInput
		}
		return v.rejectIfUsed(ctx, previous.ID)
	}
	component, ok := after.(*entity.PayComponent)
	if !ok || component == nil || (previous != nil && previous.ID != component.ID) {
		return hrerrors.ErrInvalidInput
	}
	if !isLocalFixedEarning(component) {
		return hrerrors.ErrWorkflowRequired
	}
	matches, err := v.components.SelectWhere(ctx, []domainrepo.Filter{
		{Column: "component_code", Values: []any{component.ComponentCode}},
		{Column: "component_version", Values: []any{component.ComponentVersion}},
		{Column: "jurisdiction_code", Values: []any{component.JurisdictionCode}},
	}, 2, 0)
	if err != nil {
		return err
	}
	for _, existing := range matches {
		if existing.ID != component.ID {
			return hrerrors.ErrAlreadyExists
		}
	}
	if previous != nil && !reflect.DeepEqual(previous, component) {
		return v.rejectIfUsed(ctx, component.ID)
	}
	return nil
}

func isLocalFixedEarning(component *entity.PayComponent) bool {
	return component != nil && component.ComponentType == "earning" && component.CalculationMethod == "fixed" &&
		component.CurrencyCode != nil && component.RuleSetID == nil && component.FormulaReferenceID == nil &&
		component.TaxTreatmentCode == nil && component.Taxable == nil && component.Pensionable == nil
}

func (v *payComponentValidator) rejectIfUsed(ctx context.Context, componentID string) error {
	for _, counter := range []payComponentReferenceCounter{v.compensations, v.adjustments, v.resultLines} {
		count, err := counter.CountWhere(ctx, []domainrepo.Filter{{Column: "component_id", Values: []any{componentID}}})
		if err != nil {
			return err
		}
		if count != 0 {
			return hrerrors.ErrReferenceConflict
		}
	}
	return nil
}
