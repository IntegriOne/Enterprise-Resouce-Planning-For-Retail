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

func assessmentScore(v *entity.TalentAssessmentScore) *pb.TalentAssessmentScore {
	if v == nil {
		return nil
	}
	return &pb.TalentAssessmentScore{Id: v.ID, AssessmentId: v.AssessmentID,
		FrameworkId: v.FrameworkID, DimensionId: v.DimensionID, Score: int32(v.Score),
		EvidenceDocumentVersionReferenceId: v.EvidenceDocumentVersionReferenceID, EvidenceDigest: v.EvidenceDigest}
}
func scoreResult(v *entity.TalentAssessmentScore, err error) (*pb.TalentAssessmentScoreResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.TalentAssessmentScoreResponse{Value: assessmentScore(v)}, nil
}
func scores(p operation.Page[entity.TalentAssessmentScore], err error) (*pb.TalentAssessmentScoresResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	items := make([]*pb.TalentAssessmentScore, 0, len(p.Items))
	for i := range p.Items {
		items = append(items, assessmentScore(&p.Items[i]))
	}
	return &pb.TalentAssessmentScoresResponse{Items: items, Page: &commonv1.PageInfo{Total: p.Total, Limit: int32(p.Limit), Offset: int32(p.Offset)}}, nil
}
func (s *server) CreateTalentAssessmentScore(ctx context.Context, r *pb.CreateTalentAssessmentScoreRequest) (*pb.TalentAssessmentScoreResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	in := r.Input
	value, err := score(in.Score)
	if err != nil {
		return nil, response.Error(err)
	}
	if value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return scoreResult(s.service.CreateTalentAssessmentScore(ctx, service.CreateTalentAssessmentScoreInput{
		AssessmentID: in.GetAssessmentId(), FrameworkID: in.GetFrameworkId(), DimensionID: in.GetDimensionId(), Score: *value,
		EvidenceDocumentVersionReferenceID: in.GetEvidenceDocumentVersionReferenceId(), EvidenceDigest: in.EvidenceDigest}))
}
func (s *server) UpdateTalentAssessmentScore(ctx context.Context, r *pb.UpdateTalentAssessmentScoreRequest) (*pb.TalentAssessmentScoreResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	f, err := response.ValidateMask(r.UpdateMask, "assessment_id", "framework_id", "dimension_id", "score", "evidence_document_version_reference_id", "evidence_digest")
	if err != nil {
		return nil, response.Error(err)
	}
	in := r.Input
	var value *int16
	if f["score"] {
		value, err = score(in.Score)
		if err != nil {
			return nil, response.Error(err)
		}
	}
	return scoreResult(s.service.UpdateTalentAssessmentScore(ctx, r.Id, service.UpdateTalentAssessmentScoreInput{
		AssessmentID: operation.Field[string]{Set: f["assessment_id"], Value: in.AssessmentId}, FrameworkID: operation.Field[string]{Set: f["framework_id"], Value: in.FrameworkId},
		DimensionID: operation.Field[string]{Set: f["dimension_id"], Value: in.DimensionId}, Score: operation.Field[int16]{Set: f["score"], Value: value},
		EvidenceDocumentVersionReferenceID: operation.Field[string]{Set: f["evidence_document_version_reference_id"], Value: in.EvidenceDocumentVersionReferenceId},
		EvidenceDigest:                     bytesField(f["evidence_digest"], in.EvidenceDigest)}))
}
func (s *server) DeleteTalentAssessmentScore(ctx context.Context, r *commonv1.IDRequest) (*commonv1.EmptyResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	if err := s.service.DeleteTalentAssessmentScore(ctx, r.Id); err != nil {
		return nil, response.Error(err)
	}
	return &commonv1.EmptyResponse{}, nil
}
func (s *server) GetTalentAssessmentScoreByID(ctx context.Context, r *commonv1.IDRequest) (*pb.TalentAssessmentScoreResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return scoreResult(s.service.GetTalentAssessmentScoreByID(ctx, r.Id))
}
func (s *server) GetAllTalentAssessmentScores(ctx context.Context, r *commonv1.PageRequest) (*pb.TalentAssessmentScoresResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return scores(s.service.GetAllTalentAssessmentScores(ctx, int(r.Limit), int(r.Offset)))
}
func (s *server) FindTalentAssessmentScoreByAssessmentID(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.TalentAssessmentScoresResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return scores(s.service.FindTalentAssessmentScoreByAssessmentID(ctx, r.GetValue(), int(r.Limit), int(r.Offset)))
}
func (s *server) FindTalentAssessmentScoreByFrameworkID(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.TalentAssessmentScoresResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return scores(s.service.FindTalentAssessmentScoreByFrameworkID(ctx, r.GetValue(), int(r.Limit), int(r.Offset)))
}
func (s *server) FindTalentAssessmentScoreByDimensionID(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.TalentAssessmentScoresResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return scores(s.service.FindTalentAssessmentScoreByDimensionID(ctx, r.GetValue(), int(r.Limit), int(r.Offset)))
}
