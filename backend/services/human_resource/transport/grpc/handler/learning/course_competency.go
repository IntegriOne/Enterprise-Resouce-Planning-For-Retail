package learning

import (
	"context"

	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/learning/v1"
	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/learning"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/response"
)

func courseCompetency(value *entity.LearningCourseCompetency) *pb.LearningCourseCompetency {
	return &pb.LearningCourseCompetency{CourseId: value.CourseID, CompetencyId: value.CompetencyID, TargetLevel: int32(value.TargetLevel)}
}
func courseCompetencyResponse(value *entity.LearningCourseCompetency, err error) (*pb.LearningCourseCompetencyResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.LearningCourseCompetencyResponse{Item: courseCompetency(value)}, nil
}
func courseCompetencies(page operation.Page[entity.LearningCourseCompetency], err error) (*pb.LearningCourseCompetenciesResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	items := make([]*pb.LearningCourseCompetency, 0, len(page.Items))
	for index := range page.Items {
		items = append(items, courseCompetency(&page.Items[index]))
	}
	return &pb.LearningCourseCompetenciesResponse{Items: items, Page: &common.PageInfo{Total: page.Total, Limit: int32(page.Limit), Offset: int32(page.Offset)}}, nil
}
func (h *Handler) CreateLearningCourseCompetency(ctx context.Context, request *pb.CreateLearningCourseCompetencyRequest) (*pb.LearningCourseCompetencyResponse, error) {
	if request == nil || request.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	in := request.Input
	target, err := level(in.GetTargetLevel())
	if err != nil {
		return nil, response.Error(err)
	}
	return courseCompetencyResponse(h.service.CreateLearningCourseCompetency(ctx, service.CreateLearningCourseCompetencyInput{
		CourseID: in.GetCourseId(), CompetencyID: in.GetCompetencyId(), TargetLevel: target,
	}))
}
func (h *Handler) UpdateLearningCourseCompetency(ctx context.Context, request *pb.UpdateLearningCourseCompetencyRequest) (*pb.LearningCourseCompetencyResponse, error) {
	if request == nil || request.Key == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	mask, err := response.ValidateMask(request.UpdateMask, "target_level")
	if err != nil {
		return nil, response.Error(err)
	}
	target, err := levelField(mask["target_level"], request.TargetLevel)
	if err != nil {
		return nil, response.Error(err)
	}
	key := entity.LearningCourseCompetencyKey{
		CourseID: request.Key.CourseId, CompetencyID: request.Key.CompetencyId,
	}
	return courseCompetencyResponse(h.service.UpdateLearningCourseCompetency(ctx, key,
		service.UpdateLearningCourseCompetencyInput{TargetLevel: target},
	))
}
func (h *Handler) DeleteLearningCourseCompetency(ctx context.Context, request *pb.LearningCourseCompetencyKey) (*common.EmptyResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	if err := h.service.DeleteLearningCourseCompetency(ctx, entity.LearningCourseCompetencyKey{CourseID: request.CourseId, CompetencyID: request.CompetencyId}); err != nil {
		return nil, response.Error(err)
	}
	return &common.EmptyResponse{}, nil
}
func (h *Handler) GetLearningCourseCompetencyByID(ctx context.Context, request *pb.LearningCourseCompetencyKey) (*pb.LearningCourseCompetencyResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return courseCompetencyResponse(h.service.GetLearningCourseCompetencyByID(ctx, entity.LearningCourseCompetencyKey{CourseID: request.CourseId, CompetencyID: request.CompetencyId}))
}
func (h *Handler) GetAllLearningCourseCompetencies(ctx context.Context, request *common.PageRequest) (*pb.LearningCourseCompetenciesResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return courseCompetencies(h.service.GetAllLearningCourseCompetencies(ctx, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindLearningCourseCompetencyByCourseID(ctx context.Context, request *common.StringFinderRequest) (*pb.LearningCourseCompetenciesResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return courseCompetencies(h.service.FindLearningCourseCompetencyByCourseID(ctx, *request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindLearningCourseCompetencyByCompetencyID(ctx context.Context, request *common.StringFinderRequest) (*pb.LearningCourseCompetenciesResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return courseCompetencies(h.service.FindLearningCourseCompetencyByCompetencyID(ctx, *request.Value, int(request.Limit), int(request.Offset)))
}
