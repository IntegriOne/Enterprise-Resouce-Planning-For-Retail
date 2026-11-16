// Kerangka awal dibuat dengan alat bantu; kode ini dipelihara langsung per domain.
// Aturan bisnis lintas record tetap dijalankan melalui policy/integrity dan command domain.
package service

import (
	"fmt"
	"reflect"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

func createHumanCapitalMetricDefinition(in CreateHumanCapitalMetricDefinitionInput) (*entity.HumanCapitalMetricDefinition, error) {
	v := &entity.HumanCapitalMetricDefinition{}
	v.MetricCode = in.MetricCode
	v.MetricVersion = in.MetricVersion
	v.MetricName = in.MetricName
	v.MetricDomain = in.MetricDomain
	v.NumeratorDefinition = in.NumeratorDefinition
	v.DenominatorDefinition = in.DenominatorDefinition
	v.InclusionExclusionRules = in.InclusionExclusionRules
	v.UnitCode = in.UnitCode
	v.StandardReference = in.StandardReference
	return v, nil
}
func patchHumanCapitalMetricDefinition(v *entity.HumanCapitalMetricDefinition, in UpdateHumanCapitalMetricDefinitionInput) error {
	if in.MetricCode.Set {
		if in.MetricCode.Value == nil {
			return hrerrors.Invalid("metric_code", "must not be null")
		}
		v.MetricCode = *in.MetricCode.Value
	}
	if in.MetricVersion.Set {
		if in.MetricVersion.Value == nil {
			return hrerrors.Invalid("metric_version", "must not be null")
		}
		v.MetricVersion = *in.MetricVersion.Value
	}
	if in.MetricName.Set {
		if in.MetricName.Value == nil {
			return hrerrors.Invalid("metric_name", "must not be null")
		}
		v.MetricName = *in.MetricName.Value
	}
	if in.MetricDomain.Set {
		if in.MetricDomain.Value == nil {
			return hrerrors.Invalid("metric_domain", "must not be null")
		}
		v.MetricDomain = *in.MetricDomain.Value
	}
	if in.NumeratorDefinition.Set {
		v.NumeratorDefinition = in.NumeratorDefinition.Value
	}
	if in.DenominatorDefinition.Set {
		v.DenominatorDefinition = in.DenominatorDefinition.Value
	}
	if in.InclusionExclusionRules.Set {
		if in.InclusionExclusionRules.Value == nil {
			v.InclusionExclusionRules = nil
		} else {
			v.InclusionExclusionRules = *in.InclusionExclusionRules.Value
		}
	}
	if in.UnitCode.Set {
		v.UnitCode = in.UnitCode.Value
	}
	if in.StandardReference.Set {
		v.StandardReference = in.StandardReference.Value
	}
	return nil
}
func specHumanCapitalMetricDefinition() operation.Specification[entity.HumanCapitalMetricDefinition, string] {
	return operation.Specification[entity.HumanCapitalMetricDefinition, string]{
		Domain: "analytics", Table: "hr_human_capital_metric_definitions",
		Key:         func(v *entity.HumanCapitalMetricDefinition) string { return v.ID },
		KeyString:   func(key string) string { return fmt.Sprint(key) },
		ValidateKey: func(key string) error { return entity.ValidateUUID("id", key) },
		AggregateID: func(v *entity.HumanCapitalMetricDefinition) string { return v.ID },
		Initialize: func(v *entity.HumanCapitalMetricDefinition, actor operation.Actor, now time.Time, id string) {
			v.ID = id
		},
		Touch: func(v *entity.HumanCapitalMetricDefinition, actor operation.Actor, now time.Time) {},
		Attributes: func(v *entity.HumanCapitalMetricDefinition) map[string]string {
			a := map[string]string{}
			a["id"] = fmt.Sprint(v.ID)
			return a
		},
		Guard: func(action string, old, next *entity.HumanCapitalMetricDefinition) error {
			return nil
		},
	}
}
func createHumanCapitalMetricSnapshot(in CreateHumanCapitalMetricSnapshotInput) (*entity.HumanCapitalMetricSnapshot, error) {
	v := &entity.HumanCapitalMetricSnapshot{}
	v.MetricDefinitionID = in.MetricDefinitionID
	v.PeriodStart = in.PeriodStart
	v.PeriodEnd = in.PeriodEnd
	v.OrganizationUnitReferenceID = in.OrganizationUnitReferenceID
	v.JurisdictionCode = in.JurisdictionCode
	v.ValueNumeric = in.ValueNumeric
	v.ValueText = in.ValueText
	v.PopulationCount = in.PopulationCount
	if in.DataQualityStatus != nil {
		v.DataQualityStatus = *in.DataQualityStatus
	} else {
		v.DataQualityStatus = "provisional"
	}
	if in.SnapshotVersion != nil {
		v.SnapshotVersion = *in.SnapshotVersion
	} else {
		v.SnapshotVersion = 1
	}
	v.MethodologyVersion = in.MethodologyVersion
	if in.ConfidentialityClass != nil {
		v.ConfidentialityClass = *in.ConfidentialityClass
	} else {
		v.ConfidentialityClass = "restricted"
	}
	v.SourceDigest = in.SourceDigest
	return v, nil
}
func patchHumanCapitalMetricSnapshot(v *entity.HumanCapitalMetricSnapshot, in UpdateHumanCapitalMetricSnapshotInput) error {
	if in.MetricDefinitionID.Set {
		if in.MetricDefinitionID.Value == nil {
			return hrerrors.Invalid("metric_definition_id", "must not be null")
		}
		v.MetricDefinitionID = *in.MetricDefinitionID.Value
	}
	if in.PeriodStart.Set {
		if in.PeriodStart.Value == nil {
			return hrerrors.Invalid("period_start", "must not be null")
		}
		v.PeriodStart = *in.PeriodStart.Value
	}
	if in.PeriodEnd.Set {
		if in.PeriodEnd.Value == nil {
			return hrerrors.Invalid("period_end", "must not be null")
		}
		v.PeriodEnd = *in.PeriodEnd.Value
	}
	if in.OrganizationUnitReferenceID.Set {
		v.OrganizationUnitReferenceID = in.OrganizationUnitReferenceID.Value
	}
	if in.JurisdictionCode.Set {
		v.JurisdictionCode = in.JurisdictionCode.Value
	}
	if in.ValueNumeric.Set {
		v.ValueNumeric = in.ValueNumeric.Value
	}
	if in.ValueText.Set {
		v.ValueText = in.ValueText.Value
	}
	if in.PopulationCount.Set {
		v.PopulationCount = in.PopulationCount.Value
	}
	if in.DataQualityStatus.Set {
		if in.DataQualityStatus.Value == nil {
			return hrerrors.Invalid("data_quality_status", "must not be null")
		}
		v.DataQualityStatus = *in.DataQualityStatus.Value
	}
	if in.SnapshotVersion.Set {
		if in.SnapshotVersion.Value == nil {
			return hrerrors.Invalid("snapshot_version", "must not be null")
		}
		v.SnapshotVersion = *in.SnapshotVersion.Value
	}
	if in.MethodologyVersion.Set {
		if in.MethodologyVersion.Value == nil {
			return hrerrors.Invalid("methodology_version", "must not be null")
		}
		v.MethodologyVersion = *in.MethodologyVersion.Value
	}
	if in.ConfidentialityClass.Set {
		if in.ConfidentialityClass.Value == nil {
			return hrerrors.Invalid("confidentiality_class", "must not be null")
		}
		v.ConfidentialityClass = *in.ConfidentialityClass.Value
	}
	if in.SourceDigest.Set {
		if in.SourceDigest.Value == nil {
			v.SourceDigest = nil
		} else {
			v.SourceDigest = *in.SourceDigest.Value
		}
	}
	return nil
}
func specHumanCapitalMetricSnapshot() operation.Specification[entity.HumanCapitalMetricSnapshot, string] {
	return operation.Specification[entity.HumanCapitalMetricSnapshot, string]{
		Domain: "analytics", Table: "hr_human_capital_metric_snapshots",
		Key:         func(v *entity.HumanCapitalMetricSnapshot) string { return v.ID },
		KeyString:   func(key string) string { return fmt.Sprint(key) },
		ValidateKey: func(key string) error { return entity.ValidateUUID("id", key) },
		AggregateID: func(v *entity.HumanCapitalMetricSnapshot) string { return v.ID },
		Initialize: func(v *entity.HumanCapitalMetricSnapshot, actor operation.Actor, now time.Time, id string) {
			v.ID = id
			v.CalculatedAt = now
		},
		Touch: func(v *entity.HumanCapitalMetricSnapshot, actor operation.Actor, now time.Time) {},
		Attributes: func(v *entity.HumanCapitalMetricSnapshot) map[string]string {
			a := map[string]string{}
			a["id"] = fmt.Sprint(v.ID)
			a["metric_definition_id"] = fmt.Sprint(v.MetricDefinitionID)
			if v.OrganizationUnitReferenceID != nil {
				a["organization_unit_reference_id"] = fmt.Sprint(*v.OrganizationUnitReferenceID)
			}
			a["data_quality_status"] = fmt.Sprint(v.DataQualityStatus)
			return a
		},
		Guard: func(action string, old, next *entity.HumanCapitalMetricSnapshot) error {
			if action != "create" {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "update" && !reflect.DeepEqual(old.DataQualityStatus, next.DataQualityStatus) {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "create" && string(next.DataQualityStatus) != "provisional" {
				return hrerrors.ErrWorkflowRequired
			}
			if action != "create" {
				switch string(old.DataQualityStatus) {
				case "approved", "posted", "paid", "finalized", "applied", "closed", "archived", "terminated", "deceased", "completed", "accepted", "issued", "signed", "hired", "processed", "rejected", "cancelled", "withdrawn", "superseded":
					return hrerrors.ErrWorkflowRequired
				}
			}
			return nil
		},
	}
}
