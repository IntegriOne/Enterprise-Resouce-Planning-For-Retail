package recruitment

import (
	"context"

	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/recruitment/v1"
	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/recruitment"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/response"
)

func applicationAssessment(v *entity.ApplicationAssessment) *pb.ApplicationAssessment {
	return &pb.ApplicationAssessment{
		Id: v.ID, ApplicationId: v.ApplicationID, AssessmentCode: v.AssessmentCode, AttemptNumber: v.AttemptNumber,
		RubricDocumentVersionReferenceId: v.RubricDocumentVersionReferenceID, RubricDigest: append([]byte(nil), v.RubricDigest...), Status: v.Status,
		AssessorPrincipalId: v.AssessorPrincipalID, ScheduledAt: timestamp(v.ScheduledAt), CompletedAt: timestamp(v.CompletedAt), NormalizedScore: stringPointer(v.NormalizedScore),
		RecommendationCode: v.RecommendationCode, EvidenceDocumentVersionReferenceId: v.EvidenceDocumentVersionReferenceID, EvidenceDigest: append([]byte(nil), v.EvidenceDigest...),
	}
}
func applicationAssessmentResponse(v *entity.ApplicationAssessment, err error) (*pb.ApplicationAssessmentResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.ApplicationAssessmentResponse{Item: applicationAssessment(v)}, nil
}
func applicationAssessments(page operation.Page[entity.ApplicationAssessment], err error) (*pb.ApplicationAssessmentsResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	items := make([]*pb.ApplicationAssessment, 0, len(page.Items))
	for i := range page.Items {
		items = append(items, applicationAssessment(&page.Items[i]))
	}
	return &pb.ApplicationAssessmentsResponse{Items: items, Page: pageInfo(page)}, nil
}
func (h *Handler) CreateApplicationAssessment(ctx context.Context, r *pb.CreateApplicationAssessmentRequest) (*pb.ApplicationAssessmentResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	in := r.Input
	scheduled, err := response.OptionalTimestamp(in.ScheduledAt)
	if err != nil {
		return nil, response.Error(err)
	}
	completed, err := response.OptionalTimestamp(in.CompletedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	return applicationAssessmentResponse(h.service.CreateApplicationAssessment(ctx, service.CreateApplicationAssessmentInput{
		ApplicationID: in.GetApplicationId(), AssessmentCode: in.GetAssessmentCode(), AttemptNumber: in.AttemptNumber,
		RubricDocumentVersionReferenceID: in.GetRubricDocumentVersionReferenceId(), RubricDigest: append([]byte(nil), in.RubricDigest...), Status: in.Status,
		AssessorPrincipalID: in.AssessorPrincipalId, ScheduledAt: scheduled, CompletedAt: completed, NormalizedScore: decimalPointer(in.NormalizedScore),
		RecommendationCode: in.RecommendationCode, EvidenceDocumentVersionReferenceID: in.EvidenceDocumentVersionReferenceId, EvidenceDigest: append([]byte(nil), in.EvidenceDigest...),
	}))
}
func (h *Handler) UpdateApplicationAssessment(ctx context.Context, r *pb.UpdateApplicationAssessmentRequest) (*pb.ApplicationAssessmentResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	m, err := response.ValidateMask(r.UpdateMask, "application_id", "assessment_code", "attempt_number", "rubric_document_version_reference_id", "rubric_digest", "status", "assessor_principal_id", "scheduled_at", "completed_at", "normalized_score", "recommendation_code", "evidence_document_version_reference_id", "evidence_digest")
	if err != nil {
		return nil, response.Error(err)
	}
	in := r.Input
	scheduled, err := timeField(m["scheduled_at"], in.ScheduledAt)
	if err != nil {
		return nil, response.Error(err)
	}
	completed, err := timeField(m["completed_at"], in.CompletedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	return applicationAssessmentResponse(h.service.UpdateApplicationAssessment(ctx, r.Id, service.UpdateApplicationAssessmentInput{
		ApplicationID: field(m["application_id"], in.ApplicationId), AssessmentCode: field(m["assessment_code"], in.AssessmentCode), AttemptNumber: field(m["attempt_number"], in.AttemptNumber),
		RubricDocumentVersionReferenceID: field(m["rubric_document_version_reference_id"], in.RubricDocumentVersionReferenceId), RubricDigest: bytesField(m["rubric_digest"], in.RubricDigest),
		Status: field(m["status"], in.Status), AssessorPrincipalID: field(m["assessor_principal_id"], in.AssessorPrincipalId), ScheduledAt: scheduled, CompletedAt: completed,
		NormalizedScore: field(m["normalized_score"], decimalPointer(in.NormalizedScore)), RecommendationCode: field(m["recommendation_code"], in.RecommendationCode),
		EvidenceDocumentVersionReferenceID: field(m["evidence_document_version_reference_id"], in.EvidenceDocumentVersionReferenceId), EvidenceDigest: bytesField(m["evidence_digest"], in.EvidenceDigest),
	}))
}
func (h *Handler) DeleteApplicationAssessment(ctx context.Context, r *common.IDRequest) (*common.EmptyResponse, error) {
	return deleteRecord(ctx, r, h.service.DeleteApplicationAssessment)
}
func (h *Handler) GetApplicationAssessmentByID(ctx context.Context, r *common.IDRequest) (*pb.ApplicationAssessmentResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return applicationAssessmentResponse(h.service.GetApplicationAssessmentByID(ctx, r.Id))
}
func (h *Handler) GetAllApplicationAssessments(ctx context.Context, r *common.PageRequest) (*pb.ApplicationAssessmentsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return applicationAssessments(h.service.GetAllApplicationAssessments(ctx, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindApplicationAssessmentByApplicationID(ctx context.Context, r *common.StringFinderRequest) (*pb.ApplicationAssessmentsResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return applicationAssessments(h.service.FindApplicationAssessmentByApplicationID(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindApplicationAssessmentByAssessmentCode(ctx context.Context, r *common.StringFinderRequest) (*pb.ApplicationAssessmentsResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return applicationAssessments(h.service.FindApplicationAssessmentByAssessmentCode(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindApplicationAssessmentByAttemptNumber(ctx context.Context, r *common.Int32FinderRequest) (*pb.ApplicationAssessmentsResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return applicationAssessments(h.service.FindApplicationAssessmentByAttemptNumber(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindApplicationAssessmentByRubricDocumentVersionReferenceID(ctx context.Context, r *common.StringFinderRequest) (*pb.ApplicationAssessmentsResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return applicationAssessments(h.service.FindApplicationAssessmentByRubricDocumentVersionReferenceID(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindApplicationAssessmentByStatus(ctx context.Context, r *common.StringFinderRequest) (*pb.ApplicationAssessmentsResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return applicationAssessments(h.service.FindApplicationAssessmentByStatus(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindApplicationAssessmentByAssessorPrincipalID(ctx context.Context, r *common.StringFinderRequest) (*pb.ApplicationAssessmentsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return applicationAssessments(h.service.FindApplicationAssessmentByAssessorPrincipalID(ctx, r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindApplicationAssessmentByRecommendationCode(ctx context.Context, r *common.StringFinderRequest) (*pb.ApplicationAssessmentsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return applicationAssessments(h.service.FindApplicationAssessmentByRecommendationCode(ctx, r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindApplicationAssessmentByEvidenceDocumentVersionReferenceID(ctx context.Context, r *common.StringFinderRequest) (*pb.ApplicationAssessmentsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return applicationAssessments(h.service.FindApplicationAssessmentByEvidenceDocumentVersionReferenceID(ctx, r.Value, int(r.Limit), int(r.Offset)))
}
