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
	"google.golang.org/protobuf/types/known/timestamppb"
)

func assessment(value *entity.WorkerCompetencyAssessment) *pb.WorkerCompetencyAssessment {
	return &pb.WorkerCompetencyAssessment{
		Id: value.ID, WorkerId: value.WorkerID, CompetencyId: value.CompetencyID,
		AssessedLevel: int32(value.AssessedLevel), AssessmentMethod: value.AssessmentMethod,
		AssessedAt: timestamppb.New(value.AssessedAt), AssessedByPrincipalId: value.AssessedByPrincipalID,
		ExpiresOn:                          dateString(value.ExpiresOn),
		EvidenceDocumentVersionReferenceId: value.EvidenceDocumentVersionReferenceID,
		EvidenceDigest:                     value.EvidenceDigest, IdempotencyKey: value.IdempotencyKey,
	}
}
func assessmentResponse(value *entity.WorkerCompetencyAssessment, err error) (*pb.WorkerCompetencyAssessmentResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.WorkerCompetencyAssessmentResponse{Item: assessment(value)}, nil
}
func assessments(page operation.Page[entity.WorkerCompetencyAssessment], err error) (*pb.WorkerCompetencyAssessmentsResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	items := make([]*pb.WorkerCompetencyAssessment, 0, len(page.Items))
	for index := range page.Items {
		items = append(items, assessment(&page.Items[index]))
	}
	return &pb.WorkerCompetencyAssessmentsResponse{Items: items, Page: &common.PageInfo{Total: page.Total, Limit: int32(page.Limit), Offset: int32(page.Offset)}}, nil
}
func (h *Handler) CreateWorkerCompetencyAssessment(ctx context.Context, request *pb.CreateWorkerCompetencyAssessmentRequest) (*pb.WorkerCompetencyAssessmentResponse, error) {
	if request == nil || request.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	in := request.Input
	assessed, err := level(in.GetAssessedLevel())
	if err != nil {
		return nil, response.Error(err)
	}
	at, err := response.Timestamp(in.AssessedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	expires, err := optionalDate(in.ExpiresOn)
	if err != nil {
		return nil, response.Error(err)
	}
	return assessmentResponse(h.service.CreateWorkerCompetencyAssessment(ctx, service.CreateWorkerCompetencyAssessmentInput{
		WorkerID: in.GetWorkerId(), CompetencyID: in.GetCompetencyId(), AssessedLevel: assessed,
		AssessmentMethod: in.GetAssessmentMethod(), AssessedAt: at, ExpiresOn: expires,
		EvidenceDocumentVersionReferenceID: in.EvidenceDocumentVersionReferenceId,
		EvidenceDigest:                     in.EvidenceDigest, IdempotencyKey: in.GetIdempotencyKey(),
	}))
}
func (h *Handler) UpdateWorkerCompetencyAssessment(ctx context.Context, request *pb.UpdateWorkerCompetencyAssessmentRequest) (*pb.WorkerCompetencyAssessmentResponse, error) {
	if request == nil || request.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	mask, err := response.ValidateMask(request.UpdateMask, "worker_id", "competency_id", "assessed_level", "assessment_method", "assessed_at", "expires_on", "evidence_document_version_reference_id", "evidence_digest", "idempotency_key")
	if err != nil {
		return nil, response.Error(err)
	}
	in := request.Input
	assessed, err := levelField(mask["assessed_level"], in.AssessedLevel)
	if err != nil {
		return nil, response.Error(err)
	}
	at, err := timestampField(mask["assessed_at"], in.AssessedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	expires, err := dateField(mask["expires_on"], in.ExpiresOn)
	if err != nil {
		return nil, response.Error(err)
	}
	return assessmentResponse(h.service.UpdateWorkerCompetencyAssessment(ctx, request.Id, service.UpdateWorkerCompetencyAssessmentInput{
		WorkerID: field(mask["worker_id"], in.WorkerId), CompetencyID: field(mask["competency_id"], in.CompetencyId),
		AssessedLevel: assessed, AssessmentMethod: field(mask["assessment_method"], in.AssessmentMethod),
		AssessedAt: at, ExpiresOn: expires,
		EvidenceDocumentVersionReferenceID: field(mask["evidence_document_version_reference_id"], in.EvidenceDocumentVersionReferenceId),
		EvidenceDigest:                     bytesField(mask["evidence_digest"], in.EvidenceDigest),
		IdempotencyKey:                     field(mask["idempotency_key"], in.IdempotencyKey),
	}))
}
func (h *Handler) DeleteWorkerCompetencyAssessment(ctx context.Context, request *common.IDRequest) (*common.EmptyResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	if err := h.service.DeleteWorkerCompetencyAssessment(ctx, request.Id); err != nil {
		return nil, response.Error(err)
	}
	return &common.EmptyResponse{}, nil
}
func (h *Handler) GetWorkerCompetencyAssessmentByID(ctx context.Context, request *common.IDRequest) (*pb.WorkerCompetencyAssessmentResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return assessmentResponse(h.service.GetWorkerCompetencyAssessmentByID(ctx, request.Id))
}
func (h *Handler) GetAllWorkerCompetencyAssessments(ctx context.Context, request *common.PageRequest) (*pb.WorkerCompetencyAssessmentsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return assessments(h.service.GetAllWorkerCompetencyAssessments(ctx, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindWorkerCompetencyAssessmentByWorkerID(ctx context.Context, request *common.StringFinderRequest) (*pb.WorkerCompetencyAssessmentsResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return assessments(h.service.FindWorkerCompetencyAssessmentByWorkerID(ctx, *request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindWorkerCompetencyAssessmentByCompetencyID(ctx context.Context, request *common.StringFinderRequest) (*pb.WorkerCompetencyAssessmentsResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return assessments(h.service.FindWorkerCompetencyAssessmentByCompetencyID(ctx, *request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindWorkerCompetencyAssessmentByAssessedByPrincipalID(ctx context.Context, request *common.StringFinderRequest) (*pb.WorkerCompetencyAssessmentsResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return assessments(h.service.FindWorkerCompetencyAssessmentByAssessedByPrincipalID(ctx, *request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindWorkerCompetencyAssessmentByEvidenceDocumentVersionReferenceID(ctx context.Context, request *common.StringFinderRequest) (*pb.WorkerCompetencyAssessmentsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return assessments(h.service.FindWorkerCompetencyAssessmentByEvidenceDocumentVersionReferenceID(ctx, request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindWorkerCompetencyAssessmentByIdempotencyKey(ctx context.Context, request *common.StringFinderRequest) (*pb.WorkerCompetencyAssessmentsResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return assessments(h.service.FindWorkerCompetencyAssessmentByIdempotencyKey(ctx, *request.Value, int(request.Limit), int(request.Offset)))
}
