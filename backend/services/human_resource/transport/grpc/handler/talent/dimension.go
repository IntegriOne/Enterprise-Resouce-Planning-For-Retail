package talent

import (
	"context"
	commonv1 "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/talent/v1"
	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/talent"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/response"
)

func dimension(v *entity.TalentDimension) (*pb.TalentDimension, error) {
	if v == nil {
		return nil, nil
	}
	anchors, err := response.JSONValue(v.RatingAnchors)
	if err != nil {
		return nil, err
	}
	return &pb.TalentDimension{Id: v.ID, FrameworkId: v.FrameworkID, DimensionCode: v.DimensionCode,
		DimensionName: v.DimensionName, Definition: v.Definition, RatingAnchors: anchors, WeightPercent: string(v.WeightPercent), IsRequired: v.IsRequired}, nil
}
func dimensionResult(v *entity.TalentDimension, err error) (*pb.TalentDimensionResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	value, err := dimension(v)
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.TalentDimensionResponse{Value: value}, nil
}
func dimensions(p operation.Page[entity.TalentDimension], err error) (*pb.TalentDimensionsResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	items := make([]*pb.TalentDimension, 0, len(p.Items))
	for i := range p.Items {
		value, err := dimension(&p.Items[i])
		if err != nil {
			return nil, response.Error(err)
		}
		items = append(items, value)
	}
	return &pb.TalentDimensionsResponse{Items: items, Page: &commonv1.PageInfo{Total: p.Total, Limit: int32(p.Limit), Offset: int32(p.Offset)}}, nil
}
func (s *server) CreateTalentDimension(ctx context.Context, r *pb.CreateTalentDimensionRequest) (*pb.TalentDimensionResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	in := r.Input
	anchors, err := response.JSON(in.RatingAnchors)
	if err != nil {
		return nil, response.Error(err)
	}
	return dimensionResult(s.service.CreateTalentDimension(ctx, service.CreateTalentDimensionInput{
		FrameworkID: in.GetFrameworkId(), DimensionCode: in.GetDimensionCode(), DimensionName: in.GetDimensionName(),
		Definition: in.GetDefinition(), RatingAnchors: anchors, WeightPercent: entity.Decimal(in.GetWeightPercent()), IsRequired: in.IsRequired}))
}
func (s *server) UpdateTalentDimension(ctx context.Context, r *pb.UpdateTalentDimensionRequest) (*pb.TalentDimensionResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	f, err := response.ValidateMask(r.UpdateMask, "framework_id", "dimension_code", "dimension_name", "definition", "rating_anchors", "weight_percent", "is_required")
	if err != nil {
		return nil, response.Error(err)
	}
	in := r.Input
	anchors := operation.Field[entity.JSON]{Set: f["rating_anchors"]}
	if f["rating_anchors"] && in.RatingAnchors != nil {
		value, err := response.JSON(in.RatingAnchors)
		if err != nil {
			return nil, response.Error(err)
		}
		anchors.Value = &value
	}
	return dimensionResult(s.service.UpdateTalentDimension(ctx, r.Id, service.UpdateTalentDimensionInput{
		FrameworkID: operation.Field[string]{Set: f["framework_id"], Value: in.FrameworkId}, DimensionCode: operation.Field[string]{Set: f["dimension_code"], Value: in.DimensionCode},
		DimensionName: operation.Field[string]{Set: f["dimension_name"], Value: in.DimensionName}, Definition: operation.Field[string]{Set: f["definition"], Value: in.Definition},
		RatingAnchors: anchors, WeightPercent: operation.Field[entity.Decimal]{Set: f["weight_percent"], Value: decimal(in.WeightPercent)},
		IsRequired: operation.Field[bool]{Set: f["is_required"], Value: in.IsRequired}}))
}
func (s *server) DeleteTalentDimension(ctx context.Context, r *commonv1.IDRequest) (*commonv1.EmptyResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	if err := s.service.DeleteTalentDimension(ctx, r.Id); err != nil {
		return nil, response.Error(err)
	}
	return &commonv1.EmptyResponse{}, nil
}
func (s *server) GetTalentDimensionByID(ctx context.Context, r *commonv1.IDRequest) (*pb.TalentDimensionResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return dimensionResult(s.service.GetTalentDimensionByID(ctx, r.Id))
}
func (s *server) GetAllTalentDimensions(ctx context.Context, r *commonv1.PageRequest) (*pb.TalentDimensionsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return dimensions(s.service.GetAllTalentDimensions(ctx, int(r.Limit), int(r.Offset)))
}
func (s *server) FindTalentDimensionByFrameworkID(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.TalentDimensionsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return dimensions(s.service.FindTalentDimensionByFrameworkID(ctx, r.GetValue(), int(r.Limit), int(r.Offset)))
}
func (s *server) FindTalentDimensionByDimensionCode(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.TalentDimensionsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return dimensions(s.service.FindTalentDimensionByDimensionCode(ctx, r.GetValue(), int(r.Limit), int(r.Offset)))
}
