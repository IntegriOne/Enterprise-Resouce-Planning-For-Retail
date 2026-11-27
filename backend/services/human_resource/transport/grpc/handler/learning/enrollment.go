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
	"google.golang.org/protobuf/types/known/timestamppb"
)

func enrollment(value *entity.LearningEnrollment) *pb.LearningEnrollment {
	return &pb.LearningEnrollment{
		Id: value.ID, WorkerId: value.WorkerID, CourseId: value.CourseID,
		EnrollmentStatus: string(value.EnrollmentStatus), AssignedByPrincipalId: value.AssignedByPrincipalID,
		AssignedAt: timestamppb.New(value.AssignedAt), DueAt: timestamp(value.DueAt),
		StartedAt: timestamp(value.StartedAt), CompletedAt: timestamp(value.CompletedAt),
		Score: decimalString(value.Score), CredentialReferenceId: value.CredentialReferenceID,
		FundingReferenceId: value.FundingReferenceID, IdempotencyKey: value.IdempotencyKey,
	}
}
func enrollmentResponse(value *entity.LearningEnrollment, err error) (*pb.LearningEnrollmentResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.LearningEnrollmentResponse{Item: enrollment(value)}, nil
}
func enrollments(page operation.Page[entity.LearningEnrollment], err error) (*pb.LearningEnrollmentsResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	items := make([]*pb.LearningEnrollment, 0, len(page.Items))
	for index := range page.Items {
		items = append(items, enrollment(&page.Items[index]))
	}
	return &pb.LearningEnrollmentsResponse{Items: items, Page: &common.PageInfo{Total: page.Total, Limit: int32(page.Limit), Offset: int32(page.Offset)}}, nil
}
func (h *Handler) CreateLearningEnrollment(ctx context.Context, request *pb.CreateLearningEnrollmentRequest) (*pb.LearningEnrollmentResponse, error) {
	if request == nil || request.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	in := request.Input
	due, err := response.OptionalTimestamp(in.DueAt)
	if err != nil {
		return nil, response.Error(err)
	}
	started, err := response.OptionalTimestamp(in.StartedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	completed, err := response.OptionalTimestamp(in.CompletedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	return enrollmentResponse(h.service.CreateLearningEnrollment(ctx, service.CreateLearningEnrollmentInput{
		WorkerID: in.GetWorkerId(), CourseID: in.GetCourseId(),
		EnrollmentStatus: statusPointer(in.EnrollmentStatus), AssignedByPrincipalID: in.AssignedByPrincipalId,
		DueAt: due, StartedAt: started, CompletedAt: completed, Score: decimalPointer(in.Score),
		CredentialReferenceID: in.CredentialReferenceId, FundingReferenceID: in.FundingReferenceId,
		IdempotencyKey: in.GetIdempotencyKey(),
	}))
}
func (h *Handler) UpdateLearningEnrollment(ctx context.Context, request *pb.UpdateLearningEnrollmentRequest) (*pb.LearningEnrollmentResponse, error) {
	if request == nil || request.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	mask, err := response.ValidateMask(request.UpdateMask, "worker_id", "course_id", "enrollment_status", "assigned_by_principal_id", "due_at", "started_at", "completed_at", "score", "credential_reference_id", "funding_reference_id", "idempotency_key")
	if err != nil {
		return nil, response.Error(err)
	}
	in := request.Input
	due, err := timestampField(mask["due_at"], in.DueAt)
	if err != nil {
		return nil, response.Error(err)
	}
	started, err := timestampField(mask["started_at"], in.StartedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	completed, err := timestampField(mask["completed_at"], in.CompletedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	return enrollmentResponse(h.service.UpdateLearningEnrollment(ctx, request.Id, service.UpdateLearningEnrollmentInput{
		WorkerID: field(mask["worker_id"], in.WorkerId), CourseID: field(mask["course_id"], in.CourseId),
		EnrollmentStatus:      field(mask["enrollment_status"], statusPointer(in.EnrollmentStatus)),
		AssignedByPrincipalID: field(mask["assigned_by_principal_id"], in.AssignedByPrincipalId),
		DueAt:                 due, StartedAt: started, CompletedAt: completed,
		Score:                 field(mask["score"], decimalPointer(in.Score)),
		CredentialReferenceID: field(mask["credential_reference_id"], in.CredentialReferenceId),
		FundingReferenceID:    field(mask["funding_reference_id"], in.FundingReferenceId),
		IdempotencyKey:        field(mask["idempotency_key"], in.IdempotencyKey),
	}))
}
func (h *Handler) DeleteLearningEnrollment(ctx context.Context, request *common.IDRequest) (*common.EmptyResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	if err := h.service.DeleteLearningEnrollment(ctx, request.Id); err != nil {
		return nil, response.Error(err)
	}
	return &common.EmptyResponse{}, nil
}
func (h *Handler) GetLearningEnrollmentByID(ctx context.Context, request *common.IDRequest) (*pb.LearningEnrollmentResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return enrollmentResponse(h.service.GetLearningEnrollmentByID(ctx, request.Id))
}
func (h *Handler) GetAllLearningEnrollments(ctx context.Context, request *common.PageRequest) (*pb.LearningEnrollmentsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return enrollments(h.service.GetAllLearningEnrollments(ctx, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindLearningEnrollmentByWorkerID(ctx context.Context, request *common.StringFinderRequest) (*pb.LearningEnrollmentsResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return enrollments(h.service.FindLearningEnrollmentByWorkerID(ctx, *request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindLearningEnrollmentByCourseID(ctx context.Context, request *common.StringFinderRequest) (*pb.LearningEnrollmentsResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return enrollments(h.service.FindLearningEnrollmentByCourseID(ctx, *request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindLearningEnrollmentByEnrollmentStatus(ctx context.Context, request *common.StringFinderRequest) (*pb.LearningEnrollmentsResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return enrollments(h.service.FindLearningEnrollmentByEnrollmentStatus(ctx, entity.LearningStatus(*request.Value), int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindLearningEnrollmentByAssignedByPrincipalID(ctx context.Context, request *common.StringFinderRequest) (*pb.LearningEnrollmentsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return enrollments(h.service.FindLearningEnrollmentByAssignedByPrincipalID(ctx, request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindLearningEnrollmentByCredentialReferenceID(ctx context.Context, request *common.StringFinderRequest) (*pb.LearningEnrollmentsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return enrollments(h.service.FindLearningEnrollmentByCredentialReferenceID(ctx, request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindLearningEnrollmentByFundingReferenceID(ctx context.Context, request *common.StringFinderRequest) (*pb.LearningEnrollmentsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return enrollments(h.service.FindLearningEnrollmentByFundingReferenceID(ctx, request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindLearningEnrollmentByIdempotencyKey(ctx context.Context, request *common.StringFinderRequest) (*pb.LearningEnrollmentsResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return enrollments(h.service.FindLearningEnrollmentByIdempotencyKey(ctx, *request.Value, int(request.Limit), int(request.Offset)))
}

func (h *Handler) StartLearningEnrollment(ctx context.Context, request *common.IDRequest) (*pb.LearningEnrollmentResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return enrollmentResponse(h.service.StartLearningEnrollment(ctx, request.Id))
}

func (h *Handler) CompleteLearningEnrollment(ctx context.Context, request *common.IDRequest) (*pb.LearningEnrollmentResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return enrollmentResponse(h.service.CompleteLearningEnrollment(ctx, request.Id))
}

func (h *Handler) WithdrawLearningEnrollment(ctx context.Context, request *common.IDRequest) (*pb.LearningEnrollmentResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return enrollmentResponse(h.service.WithdrawLearningEnrollment(ctx, request.Id))
}
