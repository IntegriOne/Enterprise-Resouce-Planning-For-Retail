package competency

import (
	"context"

	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/competency/v1"
	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/competency"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/response"
)

func competency(value *entity.Competency) (*pb.Competency, error) {
	scale, err := response.JSONValue(value.ProficiencyScale)
	if err != nil {
		return nil, err
	}
	return &pb.Competency{
		Id: value.ID, CompetencyCode: value.CompetencyCode, CompetencyVersion: value.CompetencyVersion,
		CompetencyName: value.CompetencyName, CategoryCode: value.CategoryCode,
		ProficiencyScale: scale, IsActive: value.IsActive,
	}, nil
}
func competencyResponse(value *entity.Competency, err error) (*pb.CompetencyResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	item, err := competency(value)
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.CompetencyResponse{Item: item}, nil
}
func competencies(page operation.Page[entity.Competency], err error) (*pb.CompetenciesResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	items := make([]*pb.Competency, 0, len(page.Items))
	for index := range page.Items {
		item, err := competency(&page.Items[index])
		if err != nil {
			return nil, response.Error(err)
		}
		items = append(items, item)
	}
	return &pb.CompetenciesResponse{Items: items, Page: &common.PageInfo{Total: page.Total, Limit: int32(page.Limit), Offset: int32(page.Offset)}}, nil
}
func (h *Handler) CreateCompetency(ctx context.Context, request *pb.CreateCompetencyRequest) (*pb.CompetencyResponse, error) {
	if request == nil || request.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	in := request.Input
	scale, err := response.JSON(in.ProficiencyScale)
	if err != nil {
		return nil, response.Error(err)
	}
	return competencyResponse(h.service.CreateCompetency(ctx, service.CreateCompetencyInput{
		CompetencyCode: in.GetCompetencyCode(), CompetencyVersion: in.GetCompetencyVersion(),
		CompetencyName: in.GetCompetencyName(), CategoryCode: in.GetCategoryCode(),
		ProficiencyScale: scale, IsActive: in.IsActive,
	}))
}
func (h *Handler) UpdateCompetency(ctx context.Context, request *pb.UpdateCompetencyRequest) (*pb.CompetencyResponse, error) {
	if request == nil || request.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	mask, err := response.ValidateMask(request.UpdateMask, "competency_code", "competency_version", "competency_name", "category_code", "proficiency_scale", "is_active")
	if err != nil {
		return nil, response.Error(err)
	}
	in := request.Input
	scale := operation.Field[entity.JSON]{Set: mask["proficiency_scale"]}
	if scale.Set && in.ProficiencyScale != nil {
		value, err := response.JSON(in.ProficiencyScale)
		if err != nil {
			return nil, response.Error(err)
		}
		scale.Value = &value
	}
	return competencyResponse(h.service.UpdateCompetency(ctx, request.Id, service.UpdateCompetencyInput{
		CompetencyCode:    field(mask["competency_code"], in.CompetencyCode),
		CompetencyVersion: field(mask["competency_version"], in.CompetencyVersion),
		CompetencyName:    field(mask["competency_name"], in.CompetencyName),
		CategoryCode:      field(mask["category_code"], in.CategoryCode),
		ProficiencyScale:  scale, IsActive: field(mask["is_active"], in.IsActive),
	}))
}
func (h *Handler) DeleteCompetency(ctx context.Context, request *common.IDRequest) (*common.EmptyResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	if err := h.service.DeleteCompetency(ctx, request.Id); err != nil {
		return nil, response.Error(err)
	}
	return &common.EmptyResponse{}, nil
}
func (h *Handler) GetCompetencyByID(ctx context.Context, request *common.IDRequest) (*pb.CompetencyResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return competencyResponse(h.service.GetCompetencyByID(ctx, request.Id))
}
func (h *Handler) GetAllCompetencies(ctx context.Context, request *common.PageRequest) (*pb.CompetenciesResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return competencies(h.service.GetAllCompetencies(ctx, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindCompetencyByCompetencyCode(ctx context.Context, request *common.StringFinderRequest) (*pb.CompetenciesResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return competencies(h.service.FindCompetencyByCompetencyCode(ctx, *request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindCompetencyByCategoryCode(ctx context.Context, request *common.StringFinderRequest) (*pb.CompetenciesResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return competencies(h.service.FindCompetencyByCategoryCode(ctx, *request.Value, int(request.Limit), int(request.Offset)))
}
