package security

import (
	"context"
	"errors"
	"testing"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
)

type metricDefinitionStub struct {
	rows []entity.HumanCapitalMetricDefinition
}

func (s *metricDefinitionStub) SelectWhere(_ context.Context, filters []domainrepo.Filter, _, _ int) ([]entity.HumanCapitalMetricDefinition, error) {
	if len(filters) != 2 || filters[0].Column != "metric_code" || filters[1].Column != "metric_version" {
		return nil, hrerrors.ErrInvalidInput
	}
	return s.rows, nil
}

type metricSnapshotStub struct {
	count int64
}

func (s *metricSnapshotStub) CountWhere(_ context.Context, filters []domainrepo.Filter) (int64, error) {
	if len(filters) != 1 || filters[0].Column != "metric_definition_id" {
		return 0, hrerrors.ErrInvalidInput
	}
	return s.count, nil
}

func TestMetricDefinitionValidatorProtectsVersionAndPublishedSnapshots(t *testing.T) {
	t.Parallel()
	definitions := &metricDefinitionStub{}
	snapshots := &metricSnapshotStub{}
	validator := &metricDefinitionValidator{definitions: definitions, snapshots: snapshots}
	value := &entity.HumanCapitalMetricDefinition{ID: "metric-1", MetricCode: "retention", MetricVersion: "1", MetricName: "Retention"}
	if err := validator.ValidateMutation(t.Context(), "hr_human_capital_metric_definitions", nil, value); err != nil {
		t.Fatalf("new definition: %v", err)
	}
	definitions.rows = []entity.HumanCapitalMetricDefinition{{ID: "metric-2"}}
	if err := validator.ValidateMutation(t.Context(), "hr_human_capital_metric_definitions", nil, value); !errors.Is(err, hrerrors.ErrAlreadyExists) {
		t.Fatalf("duplicate code and version: %v", err)
	}
	definitions.rows = nil
	changed := *value
	changed.MetricName = "Retention revised"
	if err := validator.ValidateMutation(t.Context(), "hr_human_capital_metric_definitions", value, &changed); err != nil {
		t.Fatalf("unreferenced definition update: %v", err)
	}
	snapshots.count = 1
	if err := validator.ValidateMutation(t.Context(), "hr_human_capital_metric_definitions", value, &changed); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("definition update after snapshot: %v", err)
	}
	if err := validator.ValidateMutation(t.Context(), "hr_human_capital_metric_definitions", value, nil); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("definition deletion after snapshot: %v", err)
	}
	snapshots.count = 0
	if err := validator.ValidateMutation(t.Context(), "hr_human_capital_metric_definitions", value, nil); err != nil {
		t.Fatalf("unreferenced definition deletion: %v", err)
	}
}
