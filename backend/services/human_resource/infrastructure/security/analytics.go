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

type metricDefinitionReader interface {
	SelectWhere(context.Context, []domainrepo.Filter, int, int) ([]entity.HumanCapitalMetricDefinition, error)
}

type metricSnapshotCounter interface {
	CountWhere(context.Context, []domainrepo.Filter) (int64, error)
}

type metricDefinitionValidator struct {
	definitions metricDefinitionReader
	snapshots   metricSnapshotCounter
}

// AnalyticsValidators memasang pemeriksaan definisi metrik yang dimiliki HRM.
func AnalyticsValidators(repos *repository.Registry) (map[string]MutationValidator, error) {
	if repos == nil || operation.RequireDependency(repos.HumanCapitalMetricDefinition) != nil ||
		operation.RequireDependency(repos.HumanCapitalMetricSnapshot) != nil {
		return nil, hrerrors.ErrUnavailable
	}
	return map[string]MutationValidator{
		"hr_human_capital_metric_definitions": &metricDefinitionValidator{
			definitions: repos.HumanCapitalMetricDefinition, snapshots: repos.HumanCapitalMetricSnapshot,
		},
	}, nil
}

// ValidateMutation menjaga versi metrik dan membekukan definisi yang sudah menghasilkan snapshot.
func (v *metricDefinitionValidator) ValidateMutation(ctx context.Context, table string, before, after entity.Validatable) error {
	if v == nil || table != "hr_human_capital_metric_definitions" ||
		operation.RequireDependency(v.definitions) != nil || operation.RequireDependency(v.snapshots) != nil {
		return hrerrors.ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	var previous *entity.HumanCapitalMetricDefinition
	if before != nil {
		var ok bool
		previous, ok = before.(*entity.HumanCapitalMetricDefinition)
		if !ok || previous == nil {
			return hrerrors.ErrInvalidInput
		}
	}
	if after == nil {
		if previous == nil {
			return hrerrors.ErrInvalidInput
		}
		return v.rejectIfReferenced(ctx, previous.ID)
	}
	definition, ok := after.(*entity.HumanCapitalMetricDefinition)
	if !ok || definition == nil || (previous != nil && previous.ID != definition.ID) {
		return hrerrors.ErrInvalidInput
	}
	matches, err := v.definitions.SelectWhere(ctx, []domainrepo.Filter{
		{Column: "metric_code", Values: []any{definition.MetricCode}},
		{Column: "metric_version", Values: []any{definition.MetricVersion}},
	}, 2, 0)
	if err != nil {
		return err
	}
	for _, existing := range matches {
		if existing.ID != definition.ID {
			return hrerrors.ErrAlreadyExists
		}
	}
	if previous != nil && !reflect.DeepEqual(previous, definition) {
		return v.rejectIfReferenced(ctx, definition.ID)
	}
	return nil
}

func (v *metricDefinitionValidator) rejectIfReferenced(ctx context.Context, id string) error {
	count, err := v.snapshots.CountWhere(ctx, []domainrepo.Filter{{Column: "metric_definition_id", Values: []any{id}}})
	if err != nil {
		return err
	}
	if count > 0 {
		return hrerrors.ErrReferenceConflict
	}
	return nil
}
