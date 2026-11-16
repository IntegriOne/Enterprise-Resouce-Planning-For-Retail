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

func course(value *entity.LearningCourse) *pb.LearningCourse {
	return &pb.LearningCourse{
		Id: value.ID, CourseCode: value.CourseCode, CourseVersion: value.CourseVersion,
		CourseName: value.CourseName, CourseType: value.CourseType,
		ProviderPartyReferenceId: value.ProviderPartyReferenceID, DeliveryMode: value.DeliveryMode,
		DurationHours: decimalString(value.DurationHours), CredentialAwarded: value.CredentialAwarded,
		ExpiryMonths: value.ExpiryMonths, ValidFrom: dateString(value.ValidFrom),
		ValidTo: dateString(value.ValidTo), IsActive: value.IsActive,
	}
}
func courseResponse(value *entity.LearningCourse, err error) (*pb.LearningCourseResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.LearningCourseResponse{Item: course(value)}, nil
}
func courses(page operation.Page[entity.LearningCourse], err error) (*pb.LearningCoursesResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	items := make([]*pb.LearningCourse, 0, len(page.Items))
	for index := range page.Items {
		items = append(items, course(&page.Items[index]))
	}
	return &pb.LearningCoursesResponse{Items: items, Page: &common.PageInfo{Total: page.Total, Limit: int32(page.Limit), Offset: int32(page.Offset)}}, nil
}
func (h *Handler) CreateLearningCourse(ctx context.Context, request *pb.CreateLearningCourseRequest) (*pb.LearningCourseResponse, error) {
	if request == nil || request.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	in := request.Input
	from, err := optionalDate(in.ValidFrom)
	if err != nil {
		return nil, response.Error(err)
	}
	to, err := optionalDate(in.ValidTo)
	if err != nil {
		return nil, response.Error(err)
	}
	return courseResponse(h.service.CreateLearningCourse(ctx, service.CreateLearningCourseInput{
		CourseCode: in.GetCourseCode(), CourseVersion: in.GetCourseVersion(),
		CourseName: in.GetCourseName(), CourseType: in.GetCourseType(),
		ProviderPartyReferenceID: in.ProviderPartyReferenceId, DeliveryMode: in.GetDeliveryMode(),
		DurationHours: decimalPointer(in.DurationHours), CredentialAwarded: in.CredentialAwarded,
		ExpiryMonths: in.ExpiryMonths, ValidFrom: from, ValidTo: to, IsActive: in.IsActive,
	}))
}
func (h *Handler) UpdateLearningCourse(ctx context.Context, request *pb.UpdateLearningCourseRequest) (*pb.LearningCourseResponse, error) {
	if request == nil || request.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	mask, err := response.ValidateMask(request.UpdateMask, "course_code", "course_version", "course_name", "course_type", "provider_party_reference_id", "delivery_mode", "duration_hours", "credential_awarded", "expiry_months", "valid_from", "valid_to", "is_active")
	if err != nil {
		return nil, response.Error(err)
	}
	in := request.Input
	from, err := dateField(mask["valid_from"], in.ValidFrom)
	if err != nil {
		return nil, response.Error(err)
	}
	to, err := dateField(mask["valid_to"], in.ValidTo)
	if err != nil {
		return nil, response.Error(err)
	}
	return courseResponse(h.service.UpdateLearningCourse(ctx, request.Id, service.UpdateLearningCourseInput{
		CourseCode:    field(mask["course_code"], in.CourseCode),
		CourseVersion: field(mask["course_version"], in.CourseVersion),
		CourseName:    field(mask["course_name"], in.CourseName), CourseType: field(mask["course_type"], in.CourseType),
		ProviderPartyReferenceID: field(mask["provider_party_reference_id"], in.ProviderPartyReferenceId),
		DeliveryMode:             field(mask["delivery_mode"], in.DeliveryMode),
		DurationHours:            field(mask["duration_hours"], decimalPointer(in.DurationHours)),
		CredentialAwarded:        field(mask["credential_awarded"], in.CredentialAwarded),
		ExpiryMonths:             field(mask["expiry_months"], in.ExpiryMonths), ValidFrom: from, ValidTo: to,
		IsActive: field(mask["is_active"], in.IsActive),
	}))
}
func (h *Handler) DeleteLearningCourse(ctx context.Context, request *common.IDRequest) (*common.EmptyResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	if err := h.service.DeleteLearningCourse(ctx, request.Id); err != nil {
		return nil, response.Error(err)
	}
	return &common.EmptyResponse{}, nil
}
func (h *Handler) GetLearningCourseByID(ctx context.Context, request *common.IDRequest) (*pb.LearningCourseResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return courseResponse(h.service.GetLearningCourseByID(ctx, request.Id))
}
func (h *Handler) GetAllLearningCourses(ctx context.Context, request *common.PageRequest) (*pb.LearningCoursesResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return courses(h.service.GetAllLearningCourses(ctx, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindLearningCourseByCourseCode(ctx context.Context, request *common.StringFinderRequest) (*pb.LearningCoursesResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return courses(h.service.FindLearningCourseByCourseCode(ctx, *request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindLearningCourseByCourseType(ctx context.Context, request *common.StringFinderRequest) (*pb.LearningCoursesResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return courses(h.service.FindLearningCourseByCourseType(ctx, *request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindLearningCourseByProviderPartyReferenceID(ctx context.Context, request *common.StringFinderRequest) (*pb.LearningCoursesResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return courses(h.service.FindLearningCourseByProviderPartyReferenceID(ctx, request.Value, int(request.Limit), int(request.Offset)))
}
