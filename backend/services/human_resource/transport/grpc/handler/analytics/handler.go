// Package analytics memetakan metrik HRM dengan decimal eksak dan JSON hanya pada kolom bisnis yang ditentukan.
package analytics

import (
	"context"
	"time"

	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/analytics/v1"
	commonv1 "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/analytics"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/response"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type server struct {
	pb.UnimplementedAnalyticsServiceServer
	service *service.Service
}

// Register memasang application service analytics pada gRPC privat.
func Register(registrar grpc.ServiceRegistrar, svc *service.Service) error {
	if err := operation.RequireDependency(registrar); err != nil {
		return err
	}
	if svc == nil {
		return hrerrors.ErrUnavailable
	}
	pb.RegisterAnalyticsServiceServer(registrar, &server{service: svc})
	return nil
}
func timeField(selected bool, value *timestamppb.Timestamp) (operation.Field[time.Time], error) {
	if !selected {
		return operation.Field[time.Time]{}, nil
	}
	parsed, err := response.OptionalTimestamp(value)
	return operation.Field[time.Time]{Set: true, Value: parsed}, err
}
func bytesField(selected bool, value []byte) operation.Field[[]byte] {
	if value == nil {
		return operation.Field[[]byte]{Set: selected}
	}
	return operation.Field[[]byte]{Set: selected, Value: &value}
}
func decimal(value *string) *entity.Decimal {
	if value == nil {
		return nil
	}
	parsed := entity.Decimal(*value)
	return &parsed
}
func decimalString(value *entity.Decimal) *string {
	if value == nil {
		return nil
	}
	parsed := string(*value)
	return &parsed
}
func jsonField(selected bool, value *commonv1.JsonDocument) (operation.Field[entity.JSON], error) {
	if !selected || value == nil {
		return operation.Field[entity.JSON]{Set: selected}, nil
	}
	parsed, err := response.JSON(value)
	return operation.Field[entity.JSON]{Set: selected, Value: &parsed}, err
}
func definition(value *entity.HumanCapitalMetricDefinition) (*pb.HumanCapitalMetricDefinition, error) {
	if value == nil {
		return nil, nil
	}
	rules, err := response.JSONValue(value.InclusionExclusionRules)
	if err != nil {
		return nil, err
	}
	return &pb.HumanCapitalMetricDefinition{
		Id: value.ID, MetricCode: value.MetricCode, MetricVersion: value.MetricVersion, MetricName: value.MetricName,
		MetricDomain: value.MetricDomain, NumeratorDefinition: value.NumeratorDefinition, DenominatorDefinition: value.DenominatorDefinition,
		InclusionExclusionRules: rules, UnitCode: value.UnitCode, StandardReference: value.StandardReference,
	}, nil
}
func snapshot(value *entity.HumanCapitalMetricSnapshot) *pb.HumanCapitalMetricSnapshot {
	if value == nil {
		return nil
	}
	return &pb.HumanCapitalMetricSnapshot{
		Id: value.ID, MetricDefinitionId: value.MetricDefinitionID, PeriodStart: timestamppb.New(value.PeriodStart), PeriodEnd: timestamppb.New(value.PeriodEnd),
		OrganizationUnitReferenceId: value.OrganizationUnitReferenceID, JurisdictionCode: value.JurisdictionCode,
		ValueNumeric: decimalString(value.ValueNumeric), ValueText: value.ValueText, PopulationCount: value.PopulationCount,
		DataQualityStatus: value.DataQualityStatus, SnapshotVersion: value.SnapshotVersion, MethodologyVersion: value.MethodologyVersion,
		ConfidentialityClass: value.ConfidentialityClass, CalculatedAt: timestamppb.New(value.CalculatedAt), SourceDigest: value.SourceDigest,
	}
}
func definitionResponse(value *entity.HumanCapitalMetricDefinition, err error) (*pb.HumanCapitalMetricDefinitionResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	record, err := definition(value)
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.HumanCapitalMetricDefinitionResponse{Value: record}, nil
}
func definitions(page operation.Page[entity.HumanCapitalMetricDefinition], err error) (*pb.HumanCapitalMetricDefinitionsResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	items := make([]*pb.HumanCapitalMetricDefinition, 0, len(page.Items))
	for index := range page.Items {
		record, err := definition(&page.Items[index])
		if err != nil {
			return nil, response.Error(err)
		}
		items = append(items, record)
	}
	return &pb.HumanCapitalMetricDefinitionsResponse{Items: items, Page: &commonv1.PageInfo{Total: page.Total, Limit: int32(page.Limit), Offset: int32(page.Offset)}}, nil
}
func snapshots(page operation.Page[entity.HumanCapitalMetricSnapshot], err error) (*pb.HumanCapitalMetricSnapshotsResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	items := make([]*pb.HumanCapitalMetricSnapshot, 0, len(page.Items))
	for index := range page.Items {
		items = append(items, snapshot(&page.Items[index]))
	}
	return &pb.HumanCapitalMetricSnapshotsResponse{Items: items, Page: &commonv1.PageInfo{Total: page.Total, Limit: int32(page.Limit), Offset: int32(page.Offset)}}, nil
}

func (s *server) CreateHumanCapitalMetricDefinition(ctx context.Context, request *pb.CreateHumanCapitalMetricDefinitionRequest) (*pb.HumanCapitalMetricDefinitionResponse, error) {
	if request == nil || request.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	in := request.Input
	rules, err := response.JSON(in.InclusionExclusionRules)
	if err != nil {
		return nil, response.Error(err)
	}
	return definitionResponse(s.service.CreateHumanCapitalMetricDefinition(ctx, service.CreateHumanCapitalMetricDefinitionInput{
		MetricCode: in.GetMetricCode(), MetricVersion: in.GetMetricVersion(), MetricName: in.GetMetricName(), MetricDomain: in.GetMetricDomain(),
		NumeratorDefinition: in.NumeratorDefinition, DenominatorDefinition: in.DenominatorDefinition, InclusionExclusionRules: rules, UnitCode: in.UnitCode, StandardReference: in.StandardReference,
	}))
}
func (s *server) UpdateHumanCapitalMetricDefinition(ctx context.Context, request *pb.UpdateHumanCapitalMetricDefinitionRequest) (*pb.HumanCapitalMetricDefinitionResponse, error) {
	if request == nil || request.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	selected, err := response.ValidateMask(request.UpdateMask,
		"metric_code", "metric_version", "metric_name", "metric_domain", "numerator_definition", "denominator_definition", "inclusion_exclusion_rules", "unit_code", "standard_reference",
	)
	if err != nil {
		return nil, response.Error(err)
	}
	in := request.Input
	rules, err := jsonField(selected["inclusion_exclusion_rules"], in.InclusionExclusionRules)
	if err != nil {
		return nil, response.Error(err)
	}
	return definitionResponse(s.service.UpdateHumanCapitalMetricDefinition(ctx, request.Id, service.UpdateHumanCapitalMetricDefinitionInput{
		MetricCode: operation.Field[string]{Set: selected["metric_code"], Value: in.MetricCode}, MetricVersion: operation.Field[string]{Set: selected["metric_version"], Value: in.MetricVersion},
		MetricName: operation.Field[string]{Set: selected["metric_name"], Value: in.MetricName}, MetricDomain: operation.Field[string]{Set: selected["metric_domain"], Value: in.MetricDomain},
		NumeratorDefinition: operation.Field[string]{Set: selected["numerator_definition"], Value: in.NumeratorDefinition}, DenominatorDefinition: operation.Field[string]{Set: selected["denominator_definition"], Value: in.DenominatorDefinition},
		InclusionExclusionRules: rules, UnitCode: operation.Field[string]{Set: selected["unit_code"], Value: in.UnitCode}, StandardReference: operation.Field[string]{Set: selected["standard_reference"], Value: in.StandardReference},
	}))
}
func (s *server) DeleteHumanCapitalMetricDefinition(ctx context.Context, request *commonv1.IDRequest) (*commonv1.EmptyResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	if err := s.service.DeleteHumanCapitalMetricDefinition(ctx, request.Id); err != nil {
		return nil, response.Error(err)
	}
	return &commonv1.EmptyResponse{}, nil
}
func (s *server) GetHumanCapitalMetricDefinitionByID(ctx context.Context, request *commonv1.IDRequest) (*pb.HumanCapitalMetricDefinitionResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return definitionResponse(s.service.GetHumanCapitalMetricDefinitionByID(ctx, request.Id))
}
func (s *server) GetAllHumanCapitalMetricDefinitions(ctx context.Context, request *commonv1.PageRequest) (*pb.HumanCapitalMetricDefinitionsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return definitions(s.service.GetAllHumanCapitalMetricDefinitions(ctx, int(request.Limit), int(request.Offset)))
}
func (s *server) FindHumanCapitalMetricDefinitionByMetricCode(ctx context.Context, request *commonv1.StringFinderRequest) (*pb.HumanCapitalMetricDefinitionsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return definitions(s.service.FindHumanCapitalMetricDefinitionByMetricCode(ctx, request.GetValue(), int(request.Limit), int(request.Offset)))
}
func (s *server) FindHumanCapitalMetricDefinitionByUnitCode(ctx context.Context, request *commonv1.StringFinderRequest) (*pb.HumanCapitalMetricDefinitionsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return definitions(s.service.FindHumanCapitalMetricDefinitionByUnitCode(ctx, request.Value, int(request.Limit), int(request.Offset)))
}

func (s *server) CreateHumanCapitalMetricSnapshot(ctx context.Context, request *pb.CreateHumanCapitalMetricSnapshotRequest) (*pb.HumanCapitalMetricSnapshotResponse, error) {
	if request == nil || request.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	in := request.Input
	start, err := response.Timestamp(in.PeriodStart)
	if err != nil {
		return nil, response.Error(err)
	}
	end, err := response.Timestamp(in.PeriodEnd)
	if err != nil {
		return nil, response.Error(err)
	}
	value, err := s.service.CreateHumanCapitalMetricSnapshot(ctx, service.CreateHumanCapitalMetricSnapshotInput{
		MetricDefinitionID: in.GetMetricDefinitionId(), PeriodStart: start, PeriodEnd: end, OrganizationUnitReferenceID: in.OrganizationUnitReferenceId,
		JurisdictionCode: in.JurisdictionCode, ValueNumeric: decimal(in.ValueNumeric), ValueText: in.ValueText, PopulationCount: in.PopulationCount,
		DataQualityStatus: in.DataQualityStatus, SnapshotVersion: in.SnapshotVersion, MethodologyVersion: in.GetMethodologyVersion(), ConfidentialityClass: in.ConfidentialityClass, SourceDigest: in.SourceDigest,
	})
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.HumanCapitalMetricSnapshotResponse{Value: snapshot(value)}, nil
}
func (s *server) UpdateHumanCapitalMetricSnapshot(ctx context.Context, request *pb.UpdateHumanCapitalMetricSnapshotRequest) (*pb.HumanCapitalMetricSnapshotResponse, error) {
	if request == nil || request.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	selected, err := response.ValidateMask(request.UpdateMask,
		"metric_definition_id", "period_start", "period_end", "organization_unit_reference_id", "jurisdiction_code", "value_numeric", "value_text", "population_count",
		"data_quality_status", "snapshot_version", "methodology_version", "confidentiality_class", "source_digest",
	)
	if err != nil {
		return nil, response.Error(err)
	}
	in := request.Input
	start, err := timeField(selected["period_start"], in.PeriodStart)
	if err != nil {
		return nil, response.Error(err)
	}
	end, err := timeField(selected["period_end"], in.PeriodEnd)
	if err != nil {
		return nil, response.Error(err)
	}
	value, err := s.service.UpdateHumanCapitalMetricSnapshot(ctx, request.Id, service.UpdateHumanCapitalMetricSnapshotInput{
		MetricDefinitionID: operation.Field[string]{Set: selected["metric_definition_id"], Value: in.MetricDefinitionId}, PeriodStart: start, PeriodEnd: end,
		OrganizationUnitReferenceID: operation.Field[string]{Set: selected["organization_unit_reference_id"], Value: in.OrganizationUnitReferenceId}, JurisdictionCode: operation.Field[string]{Set: selected["jurisdiction_code"], Value: in.JurisdictionCode},
		ValueNumeric: operation.Field[entity.Decimal]{Set: selected["value_numeric"], Value: decimal(in.ValueNumeric)}, ValueText: operation.Field[string]{Set: selected["value_text"], Value: in.ValueText}, PopulationCount: operation.Field[int64]{Set: selected["population_count"], Value: in.PopulationCount},
		DataQualityStatus: operation.Field[string]{Set: selected["data_quality_status"], Value: in.DataQualityStatus}, SnapshotVersion: operation.Field[int32]{Set: selected["snapshot_version"], Value: in.SnapshotVersion},
		MethodologyVersion: operation.Field[string]{Set: selected["methodology_version"], Value: in.MethodologyVersion}, ConfidentialityClass: operation.Field[string]{Set: selected["confidentiality_class"], Value: in.ConfidentialityClass}, SourceDigest: bytesField(selected["source_digest"], in.SourceDigest),
	})
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.HumanCapitalMetricSnapshotResponse{Value: snapshot(value)}, nil
}
func (s *server) DeleteHumanCapitalMetricSnapshot(ctx context.Context, request *commonv1.IDRequest) (*commonv1.EmptyResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	if err := s.service.DeleteHumanCapitalMetricSnapshot(ctx, request.Id); err != nil {
		return nil, response.Error(err)
	}
	return &commonv1.EmptyResponse{}, nil
}
func (s *server) GetHumanCapitalMetricSnapshotByID(ctx context.Context, request *commonv1.IDRequest) (*pb.HumanCapitalMetricSnapshotResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	value, err := s.service.GetHumanCapitalMetricSnapshotByID(ctx, request.Id)
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.HumanCapitalMetricSnapshotResponse{Value: snapshot(value)}, nil
}
func (s *server) GetAllHumanCapitalMetricSnapshots(ctx context.Context, request *commonv1.PageRequest) (*pb.HumanCapitalMetricSnapshotsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return snapshots(s.service.GetAllHumanCapitalMetricSnapshots(ctx, int(request.Limit), int(request.Offset)))
}
func (s *server) FindHumanCapitalMetricSnapshotByMetricDefinitionID(ctx context.Context, request *commonv1.StringFinderRequest) (*pb.HumanCapitalMetricSnapshotsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return snapshots(s.service.FindHumanCapitalMetricSnapshotByMetricDefinitionID(ctx, request.GetValue(), int(request.Limit), int(request.Offset)))
}
func (s *server) FindHumanCapitalMetricSnapshotByOrganizationUnitReferenceID(ctx context.Context, request *commonv1.StringFinderRequest) (*pb.HumanCapitalMetricSnapshotsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return snapshots(s.service.FindHumanCapitalMetricSnapshotByOrganizationUnitReferenceID(ctx, request.Value, int(request.Limit), int(request.Offset)))
}
func (s *server) FindHumanCapitalMetricSnapshotByJurisdictionCode(ctx context.Context, request *commonv1.StringFinderRequest) (*pb.HumanCapitalMetricSnapshotsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return snapshots(s.service.FindHumanCapitalMetricSnapshotByJurisdictionCode(ctx, request.Value, int(request.Limit), int(request.Offset)))
}
func (s *server) FindHumanCapitalMetricSnapshotByDataQualityStatus(ctx context.Context, request *commonv1.StringFinderRequest) (*pb.HumanCapitalMetricSnapshotsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return snapshots(s.service.FindHumanCapitalMetricSnapshotByDataQualityStatus(ctx, request.GetValue(), int(request.Limit), int(request.Offset)))
}
