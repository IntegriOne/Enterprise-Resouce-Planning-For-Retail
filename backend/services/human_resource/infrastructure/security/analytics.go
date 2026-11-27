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

type metricDefinitionLockReader interface {
	FindByIDForUpdate(context.Context, string) (*entity.HumanCapitalMetricDefinition, error)
}

type metricSnapshotReader interface {
	SelectWhere(context.Context, []domainrepo.Filter, int, int) ([]entity.HumanCapitalMetricSnapshot, error)
	CountWhere(context.Context, []domainrepo.Filter) (int64, error)
}

type metricSnapshotCounter interface {
	CountWhere(context.Context, []domainrepo.Filter) (int64, error)
}

type metricDefinitionValidator struct {
	definitions metricDefinitionReader
	snapshots   metricSnapshotCounter
}

type metricSnapshotValidator struct {
	definitions metricDefinitionLockReader
	snapshots   metricSnapshotReader
	foundation  FoundationVerifier
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

// AnalyticsSnapshotValidators mendaftarkan validator integritas snapshot metrik analitik modal manusia.
func AnalyticsSnapshotValidators(repos *repository.Registry, foundation FoundationVerifier) (map[string]MutationValidator, error) {
	if repos == nil ||
		operation.RequireDependency(repos.HumanCapitalMetricDefinition) != nil ||
		operation.RequireDependency(repos.HumanCapitalMetricSnapshot) != nil ||
		operation.RequireDependency(foundation) != nil {
		return nil, hrerrors.ErrUnavailable
	}
	return map[string]MutationValidator{
		"hr_human_capital_metric_snapshots": &metricSnapshotValidator{
			definitions: repos.HumanCapitalMetricDefinition,
			snapshots:   repos.HumanCapitalMetricSnapshot,
			foundation:  foundation,
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

// ValidateMutation menjaga snapshot metrik tetap terikat pada definisi dan lingkup unit unik.
func (v *metricSnapshotValidator) ValidateMutation(ctx context.Context, table string, before, after entity.Validatable) error {
	if v == nil || table != "hr_human_capital_metric_snapshots" ||
		operation.RequireDependency(v.definitions) != nil ||
		operation.RequireDependency(v.snapshots) != nil ||
		operation.RequireDependency(v.foundation) != nil {
		return hrerrors.ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if before == nil && after == nil {
		return hrerrors.ErrInvalidInput
	}
	if before != nil || after == nil {
		return hrerrors.ErrWorkflowRequired
	}
	snap, ok := after.(*entity.HumanCapitalMetricSnapshot)
	if !ok || snap == nil {
		return hrerrors.ErrInvalidInput
	}
	if snap.DataQualityStatus != "provisional" {
		return hrerrors.ErrWorkflowRequired
	}
	if !supportedAnalyticsConfidentiality(snap.ConfidentialityClass) {
		return hrerrors.ErrInvalidInput
	}
	def, err := v.definitions.FindByIDForUpdate(ctx, snap.MetricDefinitionID)
	if err != nil {
		return reviewReferenceError(err)
	}
	if def == nil {
		return hrerrors.ErrReferenceConflict
	}
	if snap.OrganizationUnitReferenceID != nil {
		if err := verifyFoundationOrganizationUnit(ctx, v.foundation, *snap.OrganizationUnitReferenceID, ""); err != nil {
			return err
		}
	}
	candidates, err := v.snapshots.SelectWhere(ctx, []domainrepo.Filter{
		{Column: "metric_definition_id", Values: []any{snap.MetricDefinitionID}},
		{Column: "methodology_version", Values: []any{snap.MethodologyVersion}},
		{Column: "snapshot_version", Values: []any{snap.SnapshotVersion}},
	}, domainrepo.MaxLimit, 0)
	if err != nil {
		return err
	}
	for _, existing := range candidates {
		if existing.ID == snap.ID {
			continue
		}
		if entity.CompareDates(existing.PeriodStart, snap.PeriodStart) == 0 &&
			entity.CompareDates(existing.PeriodEnd, snap.PeriodEnd) == 0 &&
			optionalStringValue(existing.OrganizationUnitReferenceID) == optionalStringValue(snap.OrganizationUnitReferenceID) &&
			optionalStringValue(existing.JurisdictionCode) == optionalStringValue(snap.JurisdictionCode) {
			return hrerrors.ErrAlreadyExists
		}
	}
	return nil
}

func optionalStringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func supportedAnalyticsConfidentiality(value string) bool {
	switch value {
	case "public", "internal", "restricted", "confidential", "highly_restricted":
		return true
	default:
		return false
	}
}
