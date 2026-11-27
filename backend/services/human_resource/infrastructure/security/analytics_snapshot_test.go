package security

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
)

type metricDefLookupStub struct {
	byID map[string]entity.HumanCapitalMetricDefinition
}

func (s *metricDefLookupStub) FindByIDForUpdate(_ context.Context, id string) (*entity.HumanCapitalMetricDefinition, error) {
	v, ok := s.byID[id]
	if !ok {
		return nil, hrerrors.ErrNotFound
	}
	copyVal := v
	return &copyVal, nil
}

func (s *metricDefLookupStub) SelectWhere(context.Context, []domainrepo.Filter, int, int) ([]entity.HumanCapitalMetricDefinition, error) {
	return nil, nil
}

type metricSnapStub struct {
	rows []entity.HumanCapitalMetricSnapshot
}

func (s *metricSnapStub) SelectWhere(context.Context, []domainrepo.Filter, int, int) ([]entity.HumanCapitalMetricSnapshot, error) {
	return s.rows, nil
}

func (s *metricSnapStub) CountWhere(context.Context, []domainrepo.Filter) (int64, error) {
	return int64(len(s.rows)), nil
}

func TestAnalyticsSnapshotValidator(t *testing.T) {
	t.Parallel()
	orgID := "00000000-0000-4000-8000-000000000701"
	defs := &metricDefLookupStub{byID: map[string]entity.HumanCapitalMetricDefinition{
		"def-1": {ID: "def-1", MetricCode: "HEADCOUNT", MetricVersion: "1"},
	}}
	snaps := &metricSnapStub{}
	validator := &metricSnapshotValidator{definitions: defs, snapshots: snaps, foundation: &erFoundationStub{}}

	start := time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2027, 1, 31, 0, 0, 0, 0, time.UTC)
	snap := &entity.HumanCapitalMetricSnapshot{
		ID: "snap-1", MetricDefinitionID: "def-1", PeriodStart: start, PeriodEnd: end,
		OrganizationUnitReferenceID: &orgID, DataQualityStatus: "provisional", SnapshotVersion: 1,
		MethodologyVersion: "m1", ConfidentialityClass: "restricted", CalculatedAt: time.Now().UTC(), SourceDigest: []byte("h"),
	}
	if err := validator.ValidateMutation(t.Context(), "hr_human_capital_metric_snapshots", nil, snap); err != nil {
		t.Fatalf("create valid metric snapshot: %v", err)
	}
	if err := validator.ValidateMutation(t.Context(), "hr_human_capital_metric_snapshots", snap, snap); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("update metric snapshot must fail: %v", err)
	}
	snaps.rows = []entity.HumanCapitalMetricSnapshot{
		{
			ID: "snap-existing", MetricDefinitionID: "def-1", PeriodStart: start, PeriodEnd: end,
			OrganizationUnitReferenceID: &orgID, MethodologyVersion: "m1", SnapshotVersion: 1,
		},
	}
	if err := validator.ValidateMutation(t.Context(), "hr_human_capital_metric_snapshots", nil, snap); !errors.Is(err, hrerrors.ErrAlreadyExists) {
		t.Fatalf("duplicate metric snapshot scope must fail: %v", err)
	}
}
