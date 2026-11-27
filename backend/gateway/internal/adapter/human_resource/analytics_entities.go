package human_resource

import (
	"context"

	"github.com/google/uuid"
	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	analytics "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/analytics/v1"
	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func toModelHumanCapitalMetricDefinition(row *analytics.HumanCapitalMetricDefinition) (*model.HumanCapitalMetricDefinition, error) {
	if row == nil {
		return nil, middleware.ErrUnavailable("invalid human resource human-capital-metric-definition response", nil)
	}
	if id, err := uuid.Parse(row.GetId()); err != nil || id == uuid.Nil {
		return nil, middleware.ErrUnavailable("invalid human resource human-capital-metric-definition ID", nil)
	}
	out := &model.HumanCapitalMetricDefinition{ID: nodeID("human-capital-metric-definition", row.GetId()), RawID: row.GetId()}
	out.MetricCode = row.GetMetricCode()
	out.MetricVersion = row.GetMetricVersion()
	out.MetricName = row.GetMetricName()
	out.MetricDomain = row.GetMetricDomain()
	out.NumeratorDefinition = row.NumeratorDefinition
	out.DenominatorDefinition = row.DenominatorDefinition
	if doc := row.GetInclusionExclusionRules(); doc != nil {
		out.InclusionExclusionRules = pbJSONString(doc.GetUtf8())
	}
	out.UnitCode = row.UnitCode
	out.StandardReference = row.StandardReference
	return out, nil
}

// HumanCapitalMetricDefinition membaca detail human-capital-metric-definition berdasarkan ID.
func (a *Adapter) HumanCapitalMetricDefinition(ctx context.Context, rawID string) (*model.HumanCapitalMetricDefinition, error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_human_capital_metric_definitions", "human_resource.human_capital_metric_definitions.read")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "human-capital-metric-definition", "human-capital-metric-definition")
	if err != nil {
		return nil, err
	}
	if a.analytics == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.analytics.GetHumanCapitalMetricDefinitionByID(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapReadError(err, "human-capital-metric-definition")
	}
	return toModelHumanCapitalMetricDefinition(resp.GetValue())
}

// HumanCapitalMetricDefinitions mendaftar dan memfilter human-capital-metric-definition dengan paginasi.
func (a *Adapter) HumanCapitalMetricDefinitions(ctx context.Context, limit, offset int, f_metric_code *string, f_unit_code *string) (*Page[model.HumanCapitalMetricDefinition], error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_human_capital_metric_definitions", "human_resource.human_capital_metric_definitions.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid human-capital-metric-definition pagination", nil)
	}
	if a.analytics == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	idx, err := pickSingleFilter(f_metric_code != nil, f_unit_code != nil)
	if err != nil {
		return nil, err
	}
	var resp *analytics.HumanCapitalMetricDefinitionsResponse
	switch idx {
	case 0:
		v, fErr := validateFilterString(f_metric_code, "metric_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.analytics.FindHumanCapitalMetricDefinitionByMetricCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 1:
		v, fErr := validateFilterString(f_unit_code, "unit_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.analytics.FindHumanCapitalMetricDefinitionByUnitCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	default:
		resp, err = a.analytics.GetAllHumanCapitalMetricDefinitions(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "human-capital-metric-definition")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "human-capital-metric-definition", toModelHumanCapitalMetricDefinition)
}

// CreateHumanCapitalMetricDefinition membuat human-capital-metric-definition baru melalui HRM.
func (a *Adapter) CreateHumanCapitalMetricDefinition(ctx context.Context, input model.CreateHumanCapitalMetricDefinitionInput) (*model.HumanCapitalMetricDefinition, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_human_capital_metric_definitions", "create", "human_resource.human_capital_metric_definitions.create")
	if err != nil {
		return nil, err
	}
	if a.analytics == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &analytics.HumanCapitalMetricDefinitionInput{}
	patch.MetricCode = input.MetricCode
	patch.MetricVersion = input.MetricVersion
	patch.MetricName = input.MetricName
	patch.MetricDomain = input.MetricDomain
	patch.NumeratorDefinition = input.NumeratorDefinition
	patch.DenominatorDefinition = input.DenominatorDefinition
	if b, jErr := optionalJSONBytes(input.InclusionExclusionRules, "inclusion_exclusion_rules"); jErr != nil {
		return nil, jErr
	} else if b != nil {
		patch.InclusionExclusionRules = &common.JsonDocument{Utf8: b}
	}
	patch.UnitCode = input.UnitCode
	patch.StandardReference = input.StandardReference
	resp, err := a.analytics.CreateHumanCapitalMetricDefinition(rpcCtx, &analytics.CreateHumanCapitalMetricDefinitionRequest{Input: patch})
	if err != nil {
		return nil, mapHRMWriteError(err, "human-capital-metric-definition")
	}
	return toModelHumanCapitalMetricDefinition(resp.GetValue())
}

// UpdateHumanCapitalMetricDefinition memperbarui human-capital-metric-definition berdasarkan FieldMask.
func (a *Adapter) UpdateHumanCapitalMetricDefinition(ctx context.Context, rawID string, input model.UpdateHumanCapitalMetricDefinitionInput) (*model.HumanCapitalMetricDefinition, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_human_capital_metric_definitions", "update", "human_resource.human_capital_metric_definitions.update")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "human-capital-metric-definition", "human-capital-metric-definition")
	if err != nil {
		return nil, err
	}
	if a.analytics == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &analytics.HumanCapitalMetricDefinitionInput{}
	mask := &fieldmaskpb.FieldMask{}
	if err := applyStringMask(mask, "metric_code", input.MetricCode, &patch.MetricCode, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "metric_version", input.MetricVersion, &patch.MetricVersion, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "metric_name", input.MetricName, &patch.MetricName, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "metric_domain", input.MetricDomain, &patch.MetricDomain, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "numerator_definition", input.NumeratorDefinition, &patch.NumeratorDefinition, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "denominator_definition", input.DenominatorDefinition, &patch.DenominatorDefinition, true); err != nil {
		return nil, err
	}
	if val, set := input.InclusionExclusionRules.ValueOK(); set {
		if val == nil {
			patch.InclusionExclusionRules = nil
		} else if b, jErr := requiredJSONBytes(*val, "inclusion_exclusion_rules"); jErr != nil {
			return nil, jErr
		} else {
			patch.InclusionExclusionRules = &common.JsonDocument{Utf8: b}
		}
		mask.Paths = append(mask.Paths, "inclusion_exclusion_rules")
	}
	if err := applyStringMask(mask, "unit_code", input.UnitCode, &patch.UnitCode, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "standard_reference", input.StandardReference, &patch.StandardReference, true); err != nil {
		return nil, err
	}
	if len(mask.Paths) == 0 {
		return nil, middleware.ErrInvalidArgument("human-capital-metric-definition update is empty", nil)
	}
	resp, err := a.analytics.UpdateHumanCapitalMetricDefinition(rpcCtx, &analytics.UpdateHumanCapitalMetricDefinitionRequest{Id: id, Input: patch, UpdateMask: mask})
	if err != nil {
		return nil, mapHRMWriteError(err, "human-capital-metric-definition")
	}
	return toModelHumanCapitalMetricDefinition(resp.GetValue())
}

// DeleteHumanCapitalMetricDefinition menghapus human-capital-metric-definition berdasarkan ID.
func (a *Adapter) DeleteHumanCapitalMetricDefinition(ctx context.Context, rawID string) (*bool, error) {
	if a == nil || a.analytics == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	return deleteByUUID(ctx, a, rawID, "hr_human_capital_metric_definitions", "human_resource.human_capital_metric_definitions.delete", "human-capital-metric-definition", func(c context.Context, req *common.IDRequest) (*common.EmptyResponse, error) {
		return a.analytics.DeleteHumanCapitalMetricDefinition(c, req)
	})
}

func toModelHumanCapitalMetricSnapshot(row *analytics.HumanCapitalMetricSnapshot) (*model.HumanCapitalMetricSnapshot, error) {
	if row == nil {
		return nil, middleware.ErrUnavailable("invalid human resource human-capital-metric-snapshot response", nil)
	}
	if id, err := uuid.Parse(row.GetId()); err != nil || id == uuid.Nil {
		return nil, middleware.ErrUnavailable("invalid human resource human-capital-metric-snapshot ID", nil)
	}
	out := &model.HumanCapitalMetricSnapshot{ID: nodeID("human-capital-metric-snapshot", row.GetId()), RawID: row.GetId()}
	out.MetricDefinitionID = row.GetMetricDefinitionId()
	out.PeriodStart = pbTimePtr(row.GetPeriodStart())
	out.PeriodEnd = pbTimePtr(row.GetPeriodEnd())
	out.OrganizationUnitReferenceID = row.OrganizationUnitReferenceId
	out.JurisdictionCode = row.JurisdictionCode
	out.ValueNumeric = row.ValueNumeric
	out.ValueText = row.ValueText
	out.PopulationCount = pbInt64StrPtr(row.PopulationCount)
	out.DataQualityStatus = row.GetDataQualityStatus()
	out.SnapshotVersion = int(row.GetSnapshotVersion())
	out.MethodologyVersion = row.GetMethodologyVersion()
	out.ConfidentialityClass = row.GetConfidentialityClass()
	out.CalculatedAt = pbTimePtr(row.GetCalculatedAt())
	out.SourceDigest = pbBase64StringValue(row.GetSourceDigest())
	return out, nil
}

// HumanCapitalMetricSnapshot membaca detail human-capital-metric-snapshot berdasarkan ID.
func (a *Adapter) HumanCapitalMetricSnapshot(ctx context.Context, rawID string) (*model.HumanCapitalMetricSnapshot, error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_human_capital_metric_snapshots", "human_resource.human_capital_metric_snapshots.read")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "human-capital-metric-snapshot", "human-capital-metric-snapshot")
	if err != nil {
		return nil, err
	}
	if a.analytics == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.analytics.GetHumanCapitalMetricSnapshotByID(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapReadError(err, "human-capital-metric-snapshot")
	}
	return toModelHumanCapitalMetricSnapshot(resp.GetValue())
}

// HumanCapitalMetricSnapshots mendaftar dan memfilter human-capital-metric-snapshot dengan paginasi.
func (a *Adapter) HumanCapitalMetricSnapshots(ctx context.Context, limit, offset int, f_metric_definition_id *string, f_organization_unit_reference_id *string, f_jurisdiction_code *string, f_data_quality_status *string) (*Page[model.HumanCapitalMetricSnapshot], error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_human_capital_metric_snapshots", "human_resource.human_capital_metric_snapshots.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid human-capital-metric-snapshot pagination", nil)
	}
	if a.analytics == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	idx, err := pickSingleFilter(f_metric_definition_id != nil, f_organization_unit_reference_id != nil, f_jurisdiction_code != nil, f_data_quality_status != nil)
	if err != nil {
		return nil, err
	}
	var resp *analytics.HumanCapitalMetricSnapshotsResponse
	switch idx {
	case 0:
		v, fErr := validateFilterString(f_metric_definition_id, "metric_definition_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.analytics.FindHumanCapitalMetricSnapshotByMetricDefinitionID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 1:
		v, fErr := validateFilterString(f_organization_unit_reference_id, "organization_unit_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.analytics.FindHumanCapitalMetricSnapshotByOrganizationUnitReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 2:
		v, fErr := validateFilterString(f_jurisdiction_code, "jurisdiction_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.analytics.FindHumanCapitalMetricSnapshotByJurisdictionCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 3:
		v, fErr := validateFilterString(f_data_quality_status, "data_quality_status")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.analytics.FindHumanCapitalMetricSnapshotByDataQualityStatus(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	default:
		resp, err = a.analytics.GetAllHumanCapitalMetricSnapshots(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "human-capital-metric-snapshot")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "human-capital-metric-snapshot", toModelHumanCapitalMetricSnapshot)
}

// CreateHumanCapitalMetricSnapshot membuat human-capital-metric-snapshot baru melalui HRM.
func (a *Adapter) CreateHumanCapitalMetricSnapshot(ctx context.Context, input model.CreateHumanCapitalMetricSnapshotInput) (*model.HumanCapitalMetricSnapshot, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_human_capital_metric_snapshots", "create", "human_resource.human_capital_metric_snapshots.create")
	if err != nil {
		return nil, err
	}
	if a.analytics == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &analytics.HumanCapitalMetricSnapshotInput{}
	patch.MetricDefinitionId = input.MetricDefinitionID
	patch.PeriodStart = optionalTimestamp(input.PeriodStart)
	patch.PeriodEnd = optionalTimestamp(input.PeriodEnd)
	patch.OrganizationUnitReferenceId = input.OrganizationUnitReferenceID
	patch.JurisdictionCode = input.JurisdictionCode
	patch.ValueNumeric = input.ValueNumeric
	patch.ValueText = input.ValueText
	if v, iErr := optionalInt64FromStr(input.PopulationCount, "population_count"); iErr != nil {
		return nil, iErr
	} else {
		patch.PopulationCount = v
	}
	patch.DataQualityStatus = input.DataQualityStatus
	patch.SnapshotVersion = optionalInt32Ptr(input.SnapshotVersion)
	patch.MethodologyVersion = input.MethodologyVersion
	patch.ConfidentialityClass = input.ConfidentialityClass
	if b, bErr := optionalBase64Bytes(input.SourceDigest, "source_digest"); bErr != nil {
		return nil, bErr
	} else if b != nil {
		patch.SourceDigest = b
	}
	resp, err := a.analytics.CreateHumanCapitalMetricSnapshot(rpcCtx, &analytics.CreateHumanCapitalMetricSnapshotRequest{Input: patch})
	if err != nil {
		return nil, mapHRMWriteError(err, "human-capital-metric-snapshot")
	}
	return toModelHumanCapitalMetricSnapshot(resp.GetValue())
}

// UpdateHumanCapitalMetricSnapshot memperbarui human-capital-metric-snapshot berdasarkan FieldMask.
func (a *Adapter) UpdateHumanCapitalMetricSnapshot(ctx context.Context, rawID string, input model.UpdateHumanCapitalMetricSnapshotInput) (*model.HumanCapitalMetricSnapshot, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_human_capital_metric_snapshots", "update", "human_resource.human_capital_metric_snapshots.update")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "human-capital-metric-snapshot", "human-capital-metric-snapshot")
	if err != nil {
		return nil, err
	}
	if a.analytics == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &analytics.HumanCapitalMetricSnapshotInput{}
	mask := &fieldmaskpb.FieldMask{}
	if err := applyStringMask(mask, "metric_definition_id", input.MetricDefinitionID, &patch.MetricDefinitionId, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "period_start", input.PeriodStart, &patch.PeriodStart, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "period_end", input.PeriodEnd, &patch.PeriodEnd, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "organization_unit_reference_id", input.OrganizationUnitReferenceID, &patch.OrganizationUnitReferenceId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "jurisdiction_code", input.JurisdictionCode, &patch.JurisdictionCode, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "value_numeric", input.ValueNumeric, &patch.ValueNumeric, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "value_text", input.ValueText, &patch.ValueText, true); err != nil {
		return nil, err
	}
	if err := applyInt64Mask(mask, "population_count", input.PopulationCount, &patch.PopulationCount, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "data_quality_status", input.DataQualityStatus, &patch.DataQualityStatus, true); err != nil {
		return nil, err
	}
	if err := applyInt32Mask(mask, "snapshot_version", input.SnapshotVersion, &patch.SnapshotVersion, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "methodology_version", input.MethodologyVersion, &patch.MethodologyVersion, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "confidentiality_class", input.ConfidentialityClass, &patch.ConfidentialityClass, true); err != nil {
		return nil, err
	}
	if err := applyBytesMask(mask, "source_digest", input.SourceDigest, &patch.SourceDigest, true); err != nil {
		return nil, err
	}
	if len(mask.Paths) == 0 {
		return nil, middleware.ErrInvalidArgument("human-capital-metric-snapshot update is empty", nil)
	}
	resp, err := a.analytics.UpdateHumanCapitalMetricSnapshot(rpcCtx, &analytics.UpdateHumanCapitalMetricSnapshotRequest{Id: id, Input: patch, UpdateMask: mask})
	if err != nil {
		return nil, mapHRMWriteError(err, "human-capital-metric-snapshot")
	}
	return toModelHumanCapitalMetricSnapshot(resp.GetValue())
}

// DeleteHumanCapitalMetricSnapshot menghapus human-capital-metric-snapshot berdasarkan ID.
func (a *Adapter) DeleteHumanCapitalMetricSnapshot(ctx context.Context, rawID string) (*bool, error) {
	if a == nil || a.analytics == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	return deleteByUUID(ctx, a, rawID, "hr_human_capital_metric_snapshots", "human_resource.human_capital_metric_snapshots.delete", "human-capital-metric-snapshot", func(c context.Context, req *common.IDRequest) (*common.EmptyResponse, error) {
		return a.analytics.DeleteHumanCapitalMetricSnapshot(c, req)
	})
}
