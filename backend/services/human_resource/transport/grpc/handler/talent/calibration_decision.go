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

func calibrationDecision(v *entity.TalentCalibrationDecision) *pb.TalentCalibrationDecision {
	if v == nil {
		return nil
	}
	return &pb.TalentCalibrationDecision{Id: v.ID, SessionId: v.SessionID,
		TalentReviewCycleId: v.TalentReviewCycleID, AssessmentId: v.AssessmentID, DimensionId: v.DimensionID,
		OriginalScore: int32(v.OriginalScore), CalibratedScore: int32(v.CalibratedScore), RationaleDocumentVersionReferenceId: v.RationaleDocumentVersionReferenceID,
		RationaleDigest: v.RationaleDigest, DecidedByPrincipalId: v.DecidedByPrincipalID, DecidedAt: timestamppb.New(v.DecidedAt)}
}
func decisionResult(v *entity.TalentCalibrationDecision, err error) (*pb.TalentCalibrationDecisionResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.TalentCalibrationDecisionResponse{Value: calibrationDecision(v)}, nil
}
func decisions(p operation.Page[entity.TalentCalibrationDecision], err error) (*pb.TalentCalibrationDecisionsResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	items := make([]*pb.TalentCalibrationDecision, 0, len(p.Items))
	for i := range p.Items {
		items = append(items, calibrationDecision(&p.Items[i]))
	}
	return &pb.TalentCalibrationDecisionsResponse{Items: items, Page: &commonv1.PageInfo{Total: p.Total, Limit: int32(p.Limit), Offset: int32(p.Offset)}}, nil
}
func (s *server) CreateTalentCalibrationDecision(ctx context.Context, r *pb.CreateTalentCalibrationDecisionRequest) (*pb.TalentCalibrationDecisionResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	in := r.Input
	original, err := score(in.OriginalScore)
	if err != nil {
		return nil, response.Error(err)
	}
	calibrated, err := score(in.CalibratedScore)
	if err != nil {
		return nil, response.Error(err)
	}
	if original == nil || calibrated == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return decisionResult(s.service.CreateTalentCalibrationDecision(ctx, service.CreateTalentCalibrationDecisionInput{
		SessionID: in.GetSessionId(), TalentReviewCycleID: in.GetTalentReviewCycleId(), AssessmentID: in.GetAssessmentId(),
		DimensionID: in.GetDimensionId(), OriginalScore: *original, CalibratedScore: *calibrated,
		RationaleDocumentVersionReferenceID: in.GetRationaleDocumentVersionReferenceId(), RationaleDigest: in.RationaleDigest}))
}
func (s *server) UpdateTalentCalibrationDecision(ctx context.Context, r *pb.UpdateTalentCalibrationDecisionRequest) (*pb.TalentCalibrationDecisionResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	f, err := response.ValidateMask(r.UpdateMask, "session_id", "talent_review_cycle_id", "assessment_id", "dimension_id", "original_score", "calibrated_score", "rationale_document_version_reference_id", "rationale_digest")
	if err != nil {
		return nil, response.Error(err)
	}
	in := r.Input
	var original, calibrated *int16
	if f["original_score"] {
		original, err = score(in.OriginalScore)
		if err != nil {
			return nil, response.Error(err)
		}
	}
	if f["calibrated_score"] {
		calibrated, err = score(in.CalibratedScore)
		if err != nil {
			return nil, response.Error(err)
		}
	}
	return decisionResult(s.service.UpdateTalentCalibrationDecision(ctx, r.Id, service.UpdateTalentCalibrationDecisionInput{
		SessionID: operation.Field[string]{Set: f["session_id"], Value: in.SessionId}, TalentReviewCycleID: operation.Field[string]{Set: f["talent_review_cycle_id"], Value: in.TalentReviewCycleId},
		AssessmentID: operation.Field[string]{Set: f["assessment_id"], Value: in.AssessmentId}, DimensionID: operation.Field[string]{Set: f["dimension_id"], Value: in.DimensionId},
		OriginalScore: operation.Field[int16]{Set: f["original_score"], Value: original}, CalibratedScore: operation.Field[int16]{Set: f["calibrated_score"], Value: calibrated},
		RationaleDocumentVersionReferenceID: operation.Field[string]{Set: f["rationale_document_version_reference_id"], Value: in.RationaleDocumentVersionReferenceId},
		RationaleDigest:                     bytesField(f["rationale_digest"], in.RationaleDigest)}))
}
func (s *server) DeleteTalentCalibrationDecision(ctx context.Context, r *commonv1.IDRequest) (*commonv1.EmptyResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	if err := s.service.DeleteTalentCalibrationDecision(ctx, r.Id); err != nil {
		return nil, response.Error(err)
	}
	return &commonv1.EmptyResponse{}, nil
}
func (s *server) GetTalentCalibrationDecisionByID(ctx context.Context, r *commonv1.IDRequest) (*pb.TalentCalibrationDecisionResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return decisionResult(s.service.GetTalentCalibrationDecisionByID(ctx, r.Id))
}
func (s *server) RecordTalentCalibrationDecision(ctx context.Context, r *pb.RecordTalentCalibrationDecisionRequest) (*pb.TalentCalibrationDecisionResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	calibrated, err := score(&r.CalibratedScore)
	if err != nil {
		return nil, response.Error(err)
	}
	return decisionResult(s.service.RecordTalentCalibrationDecision(ctx, service.RecordTalentCalibrationDecisionInput{
		SessionID:                           r.SessionId,
		AssessmentID:                        r.AssessmentId,
		DimensionID:                         r.DimensionId,
		CalibratedScore:                     *calibrated,
		RationaleDocumentVersionReferenceID: r.RationaleDocumentVersionReferenceId,
		RationaleDigest:                     append([]byte(nil), r.RationaleDigest...),
	}))
}
func (s *server) GetAllTalentCalibrationDecisions(ctx context.Context, r *commonv1.PageRequest) (*pb.TalentCalibrationDecisionsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return decisions(s.service.GetAllTalentCalibrationDecisions(ctx, int(r.Limit), int(r.Offset)))
}
func (s *server) FindTalentCalibrationDecisionBySessionID(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.TalentCalibrationDecisionsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return decisions(s.service.FindTalentCalibrationDecisionBySessionID(ctx, r.GetValue(), int(r.Limit), int(r.Offset)))
}
func (s *server) FindTalentCalibrationDecisionByTalentReviewCycleID(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.TalentCalibrationDecisionsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return decisions(s.service.FindTalentCalibrationDecisionByTalentReviewCycleID(ctx, r.GetValue(), int(r.Limit), int(r.Offset)))
}
func (s *server) FindTalentCalibrationDecisionByAssessmentID(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.TalentCalibrationDecisionsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return decisions(s.service.FindTalentCalibrationDecisionByAssessmentID(ctx, r.GetValue(), int(r.Limit), int(r.Offset)))
}
func (s *server) FindTalentCalibrationDecisionByDimensionID(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.TalentCalibrationDecisionsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return decisions(s.service.FindTalentCalibrationDecisionByDimensionID(ctx, r.GetValue(), int(r.Limit), int(r.Offset)))
}
func (s *server) FindTalentCalibrationDecisionByRationaleDocumentVersionReferenceID(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.TalentCalibrationDecisionsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return decisions(s.service.FindTalentCalibrationDecisionByRationaleDocumentVersionReferenceID(ctx, r.GetValue(), int(r.Limit), int(r.Offset)))
}
func (s *server) FindTalentCalibrationDecisionByDecidedByPrincipalID(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.TalentCalibrationDecisionsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return decisions(s.service.FindTalentCalibrationDecisionByDecidedByPrincipalID(ctx, r.GetValue(), int(r.Limit), int(r.Offset)))
}
