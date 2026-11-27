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
	"google.golang.org/protobuf/types/known/timestamppb"
)

func assessment(v *entity.TalentAssessment) *pb.TalentAssessment {
	if v == nil {
		return nil
	}
	return &pb.TalentAssessment{Id: v.ID, WorkerId: v.WorkerID, TalentReviewCycleId: v.TalentReviewCycleID,
		FrameworkId: v.FrameworkID, AssessmentVersion: v.AssessmentVersion, PerformanceReviewId: v.PerformanceReviewID, Status: v.Status,
		AssessedByPrincipalId: v.AssessedByPrincipalID, AssessedAt: timestamppb.New(v.AssessedAt), EvidenceDocumentVersionReferenceId: v.EvidenceDocumentVersionReferenceID,
		EvidenceDigest: v.EvidenceDigest, SegmentCode: v.SegmentCode, FinalizedByPrincipalId: v.FinalizedByPrincipalID, FinalizedAt: timestamp(v.FinalizedAt)}
}
func assessmentResult(v *entity.TalentAssessment, err error) (*pb.TalentAssessmentResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.TalentAssessmentResponse{Value: assessment(v)}, nil
}
func assessments(p operation.Page[entity.TalentAssessment], err error) (*pb.TalentAssessmentsResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	items := make([]*pb.TalentAssessment, 0, len(p.Items))
	for i := range p.Items {
		items = append(items, assessment(&p.Items[i]))
	}
	return &pb.TalentAssessmentsResponse{Items: items, Page: &commonv1.PageInfo{Total: p.Total, Limit: int32(p.Limit), Offset: int32(p.Offset)}}, nil
}
func (s *server) CreateTalentAssessment(ctx context.Context, r *pb.CreateTalentAssessmentRequest) (*pb.TalentAssessmentResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	in := r.Input
	finalized, err := response.OptionalTimestamp(in.FinalizedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	return assessmentResult(s.service.CreateTalentAssessment(ctx, service.CreateTalentAssessmentInput{
		WorkerID: in.GetWorkerId(), TalentReviewCycleID: in.GetTalentReviewCycleId(), FrameworkID: in.GetFrameworkId(),
		AssessmentVersion: in.GetAssessmentVersion(), PerformanceReviewID: in.PerformanceReviewId, Status: in.Status,
		EvidenceDocumentVersionReferenceID: in.GetEvidenceDocumentVersionReferenceId(), EvidenceDigest: in.EvidenceDigest,
		SegmentCode: in.SegmentCode, FinalizedByPrincipalID: in.FinalizedByPrincipalId, FinalizedAt: finalized}))
}
func (s *server) UpdateTalentAssessment(ctx context.Context, r *pb.UpdateTalentAssessmentRequest) (*pb.TalentAssessmentResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	f, err := response.ValidateMask(r.UpdateMask, "worker_id", "talent_review_cycle_id", "framework_id", "assessment_version", "performance_review_id", "status",
		"evidence_document_version_reference_id", "evidence_digest", "segment_code", "finalized_by_principal_id", "finalized_at")
	if err != nil {
		return nil, response.Error(err)
	}
	in := r.Input
	finalized, err := timeField(f["finalized_at"], in.FinalizedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	return assessmentResult(s.service.UpdateTalentAssessment(ctx, r.Id, service.UpdateTalentAssessmentInput{
		WorkerID: operation.Field[string]{Set: f["worker_id"], Value: in.WorkerId}, TalentReviewCycleID: operation.Field[string]{Set: f["talent_review_cycle_id"], Value: in.TalentReviewCycleId},
		FrameworkID: operation.Field[string]{Set: f["framework_id"], Value: in.FrameworkId}, AssessmentVersion: operation.Field[int32]{Set: f["assessment_version"], Value: in.AssessmentVersion},
		PerformanceReviewID: operation.Field[string]{Set: f["performance_review_id"], Value: in.PerformanceReviewId}, Status: operation.Field[string]{Set: f["status"], Value: in.Status},
		EvidenceDocumentVersionReferenceID: operation.Field[string]{Set: f["evidence_document_version_reference_id"], Value: in.EvidenceDocumentVersionReferenceId},
		EvidenceDigest:                     bytesField(f["evidence_digest"], in.EvidenceDigest), SegmentCode: operation.Field[string]{Set: f["segment_code"], Value: in.SegmentCode},
		FinalizedByPrincipalID: operation.Field[string]{Set: f["finalized_by_principal_id"], Value: in.FinalizedByPrincipalId}, FinalizedAt: finalized}))
}
func (s *server) DeleteTalentAssessment(ctx context.Context, r *commonv1.IDRequest) (*commonv1.EmptyResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	if err := s.service.DeleteTalentAssessment(ctx, r.Id); err != nil {
		return nil, response.Error(err)
	}
	return &commonv1.EmptyResponse{}, nil
}
func (s *server) GetTalentAssessmentByID(ctx context.Context, r *commonv1.IDRequest) (*pb.TalentAssessmentResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return assessmentResult(s.service.GetTalentAssessmentByID(ctx, r.Id))
}
func (s *server) FinalizeTalentAssessment(ctx context.Context, r *commonv1.IDRequest) (*pb.TalentAssessmentResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return assessmentResult(s.service.FinalizeTalentAssessment(ctx, r.Id))
}
func (s *server) SubmitTalentAssessment(ctx context.Context, r *commonv1.IDRequest) (*pb.TalentAssessmentResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return assessmentResult(s.service.SubmitTalentAssessment(ctx, r.Id))
}
func (s *server) WithdrawTalentAssessment(ctx context.Context, r *commonv1.IDRequest) (*pb.TalentAssessmentResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return assessmentResult(s.service.WithdrawTalentAssessment(ctx, r.Id))
}
func (s *server) GetAllTalentAssessments(ctx context.Context, r *commonv1.PageRequest) (*pb.TalentAssessmentsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return assessments(s.service.GetAllTalentAssessments(ctx, int(r.Limit), int(r.Offset)))
}
func (s *server) FindTalentAssessmentByWorkerID(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.TalentAssessmentsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return assessments(s.service.FindTalentAssessmentByWorkerID(ctx, r.GetValue(), int(r.Limit), int(r.Offset)))
}
func (s *server) FindTalentAssessmentByTalentReviewCycleID(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.TalentAssessmentsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return assessments(s.service.FindTalentAssessmentByTalentReviewCycleID(ctx, r.GetValue(), int(r.Limit), int(r.Offset)))
}
func (s *server) FindTalentAssessmentByFrameworkID(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.TalentAssessmentsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return assessments(s.service.FindTalentAssessmentByFrameworkID(ctx, r.GetValue(), int(r.Limit), int(r.Offset)))
}
func (s *server) FindTalentAssessmentByPerformanceReviewID(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.TalentAssessmentsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return assessments(s.service.FindTalentAssessmentByPerformanceReviewID(ctx, r.Value, int(r.Limit), int(r.Offset)))
}
func (s *server) FindTalentAssessmentByStatus(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.TalentAssessmentsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return assessments(s.service.FindTalentAssessmentByStatus(ctx, r.GetValue(), int(r.Limit), int(r.Offset)))
}
func (s *server) FindTalentAssessmentByAssessedByPrincipalID(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.TalentAssessmentsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return assessments(s.service.FindTalentAssessmentByAssessedByPrincipalID(ctx, r.GetValue(), int(r.Limit), int(r.Offset)))
}
func (s *server) FindTalentAssessmentByEvidenceDocumentVersionReferenceID(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.TalentAssessmentsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return assessments(s.service.FindTalentAssessmentByEvidenceDocumentVersionReferenceID(ctx, r.GetValue(), int(r.Limit), int(r.Offset)))
}
func (s *server) FindTalentAssessmentBySegmentCode(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.TalentAssessmentsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return assessments(s.service.FindTalentAssessmentBySegmentCode(ctx, r.Value, int(r.Limit), int(r.Offset)))
}
func (s *server) FindTalentAssessmentByFinalizedByPrincipalID(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.TalentAssessmentsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return assessments(s.service.FindTalentAssessmentByFinalizedByPrincipalID(ctx, r.Value, int(r.Limit), int(r.Offset)))
}
