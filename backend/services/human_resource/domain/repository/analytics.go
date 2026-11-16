package repository

import (
	"context"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
)

// HumanCapitalMetricDefinitionRepository adalah port persistensi hr_human_capital_metric_definitions.
type HumanCapitalMetricDefinitionRepository interface {
	Repository[entity.HumanCapitalMetricDefinition, string]
	InsertHumanCapitalMetricDefinitions(context.Context, *entity.HumanCapitalMetricDefinition) error
	UpdateHumanCapitalMetricDefinitions(context.Context, *entity.HumanCapitalMetricDefinition) error
	FindHumanCapitalMetricDefinitionsByUUID(context.Context, string) (*entity.HumanCapitalMetricDefinition, error)
	FindHumanCapitalMetricDefinitionsByUUIDForUpdate(context.Context, string) (*entity.HumanCapitalMetricDefinition, error)
	DeleteHumanCapitalMetricDefinitionsByUUID(context.Context, string) error
	GetAllHumanCapitalMetricDefinitions(context.Context, int, int) ([]entity.HumanCapitalMetricDefinition, error)
	CountHumanCapitalMetricDefinitions(context.Context) (int64, error)
	SelectHumanCapitalMetricDefinitionsWhere(context.Context, []Filter, int, int) ([]entity.HumanCapitalMetricDefinition, error)
	CountHumanCapitalMetricDefinitionsWhere(context.Context, []Filter) (int64, error)
	FindHumanCapitalMetricDefinitionsByMetricCode(context.Context, string, int, int) ([]entity.HumanCapitalMetricDefinition, error)
	FindHumanCapitalMetricDefinitionsByUnitCode(context.Context, *string, int, int) ([]entity.HumanCapitalMetricDefinition, error)
	FindByMetricCode(ctx context.Context, value string, limit, offset int) ([]entity.HumanCapitalMetricDefinition, error)
	FindByUnitCode(ctx context.Context, value *string, limit, offset int) ([]entity.HumanCapitalMetricDefinition, error)
}

// HumanCapitalMetricSnapshotRepository adalah port persistensi hr_human_capital_metric_snapshots.
type HumanCapitalMetricSnapshotRepository interface {
	Repository[entity.HumanCapitalMetricSnapshot, string]
	InsertHumanCapitalMetricSnapshots(context.Context, *entity.HumanCapitalMetricSnapshot) error
	UpdateHumanCapitalMetricSnapshots(context.Context, *entity.HumanCapitalMetricSnapshot) error
	FindHumanCapitalMetricSnapshotsByUUID(context.Context, string) (*entity.HumanCapitalMetricSnapshot, error)
	FindHumanCapitalMetricSnapshotsByUUIDForUpdate(context.Context, string) (*entity.HumanCapitalMetricSnapshot, error)
	DeleteHumanCapitalMetricSnapshotsByUUID(context.Context, string) error
	GetAllHumanCapitalMetricSnapshots(context.Context, int, int) ([]entity.HumanCapitalMetricSnapshot, error)
	CountHumanCapitalMetricSnapshots(context.Context) (int64, error)
	SelectHumanCapitalMetricSnapshotsWhere(context.Context, []Filter, int, int) ([]entity.HumanCapitalMetricSnapshot, error)
	CountHumanCapitalMetricSnapshotsWhere(context.Context, []Filter) (int64, error)
	FindHumanCapitalMetricSnapshotsByMetricDefinitionID(context.Context, string, int, int) ([]entity.HumanCapitalMetricSnapshot, error)
	FindHumanCapitalMetricSnapshotsByOrganizationUnitReferenceID(context.Context, *string, int, int) ([]entity.HumanCapitalMetricSnapshot, error)
	FindHumanCapitalMetricSnapshotsByJurisdictionCode(context.Context, *string, int, int) ([]entity.HumanCapitalMetricSnapshot, error)
	FindHumanCapitalMetricSnapshotsByDataQualityStatus(context.Context, string, int, int) ([]entity.HumanCapitalMetricSnapshot, error)
	FindByMetricDefinitionID(ctx context.Context, value string, limit, offset int) ([]entity.HumanCapitalMetricSnapshot, error)
	FindByOrganizationUnitReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.HumanCapitalMetricSnapshot, error)
	FindByJurisdictionCode(ctx context.Context, value *string, limit, offset int) ([]entity.HumanCapitalMetricSnapshot, error)
	FindByDataQualityStatus(ctx context.Context, value string, limit, offset int) ([]entity.HumanCapitalMetricSnapshot, error)
}
