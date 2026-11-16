// Kerangka awal dibuat dengan alat bantu; kode ini dipelihara langsung per domain.
// Aturan bisnis lintas record tetap dijalankan melalui policy/integrity dan command domain.
package service

import (
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

// CreateHumanCapitalMetricDefinitionInput adalah object input aplikasi, bukan entity ORM.
type CreateHumanCapitalMetricDefinitionInput struct {
	MetricCode              string
	MetricVersion           string
	MetricName              string
	MetricDomain            string
	NumeratorDefinition     *string
	DenominatorDefinition   *string
	InclusionExclusionRules entity.JSON
	UnitCode                *string
	StandardReference       *string
}

// UpdateHumanCapitalMetricDefinitionInput membedakan field absen, null dan nilai eksplisit.
type UpdateHumanCapitalMetricDefinitionInput struct {
	MetricCode              operation.Field[string]
	MetricVersion           operation.Field[string]
	MetricName              operation.Field[string]
	MetricDomain            operation.Field[string]
	NumeratorDefinition     operation.Field[string]
	DenominatorDefinition   operation.Field[string]
	InclusionExclusionRules operation.Field[entity.JSON]
	UnitCode                operation.Field[string]
	StandardReference       operation.Field[string]
}

// CreateHumanCapitalMetricSnapshotInput adalah object input aplikasi, bukan entity ORM.
type CreateHumanCapitalMetricSnapshotInput struct {
	MetricDefinitionID          string
	PeriodStart                 time.Time
	PeriodEnd                   time.Time
	OrganizationUnitReferenceID *string
	JurisdictionCode            *string
	ValueNumeric                *entity.Decimal
	ValueText                   *string
	PopulationCount             *int64
	DataQualityStatus           *string
	SnapshotVersion             *int32
	MethodologyVersion          string
	ConfidentialityClass        *string
	SourceDigest                []byte
}

// UpdateHumanCapitalMetricSnapshotInput membedakan field absen, null dan nilai eksplisit.
type UpdateHumanCapitalMetricSnapshotInput struct {
	MetricDefinitionID          operation.Field[string]
	PeriodStart                 operation.Field[time.Time]
	PeriodEnd                   operation.Field[time.Time]
	OrganizationUnitReferenceID operation.Field[string]
	JurisdictionCode            operation.Field[string]
	ValueNumeric                operation.Field[entity.Decimal]
	ValueText                   operation.Field[string]
	PopulationCount             operation.Field[int64]
	DataQualityStatus           operation.Field[string]
	SnapshotVersion             operation.Field[int32]
	MethodologyVersion          operation.Field[string]
	ConfidentialityClass        operation.Field[string]
	SourceDigest                operation.Field[[]byte]
}
