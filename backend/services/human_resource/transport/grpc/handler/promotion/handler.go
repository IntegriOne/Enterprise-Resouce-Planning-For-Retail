// Package promotion menerjemahkan kontrak RPC bertipe ke service promosi HRM.
package promotion

import (
	"context"
	"time"

	commonv1 "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/promotion/v1"
	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/promotion"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/response"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type server struct {
	pb.UnimplementedPromotionServiceServer
	service *service.Service
}

// Register memasang service promosi pada registrar milik server HRM.
func Register(registrar grpc.ServiceRegistrar, svc *service.Service) error {
	if err := operation.RequireDependency(registrar); err != nil {
		return err
	}
	if svc == nil {
		return hrerrors.ErrUnavailable
	}
	pb.RegisterPromotionServiceServer(registrar, &server{service: svc})
	return nil
}

func timestamp(value *time.Time) *timestamppb.Timestamp {
	if value == nil {
		return nil
	}
	return timestamppb.New(*value)
}
func timeField(selected bool, value *timestamppb.Timestamp) (operation.Field[time.Time], error) {
	if !selected {
		return operation.Field[time.Time]{}, nil
	}
	parsed, err := response.OptionalTimestamp(value)
	return operation.Field[time.Time]{Set: true, Value: parsed}, err
}
func bytesField(selected bool, value []byte) operation.Field[[]byte] {
	if value == nil {
		return operation.Field[[]byte]{Set: selected}
	}
	return operation.Field[[]byte]{Set: selected, Value: &value}
}
func recommendation(v *entity.PromotionRecommendation) *pb.PromotionRecommendation {
	if v == nil {
		return nil
	}
	return &pb.PromotionRecommendation{Id: v.ID, RecommendationNumber: v.RecommendationNumber,
		WorkerId: v.WorkerID, SourceAssignmentId: v.SourceAssignmentID, TargetPositionId: v.TargetPositionID,
		AssessmentId: v.AssessmentID, SupersedesRecommendationId: v.SupersedesRecommendationID,
		RecommendedEffectiveDate:            timestamppb.New(v.RecommendedEffectiveDate),
		RationaleDocumentVersionReferenceId: v.RationaleDocumentVersionReferenceID, RationaleDigest: v.RationaleDigest,
		RecommendedByPrincipalId: v.RecommendedByPrincipalID, RecommendedAt: timestamppb.New(v.RecommendedAt),
		Status: v.Status, ApprovalDecisionId: v.ApprovalDecisionID, AppliedAssignmentId: v.AppliedAssignmentID,
		AppliedLifecycleEventId: v.AppliedLifecycleEventID, AppliedAt: timestamp(v.AppliedAt), IdempotencyKey: v.IdempotencyKey}
}
func recommendationResult(v *entity.PromotionRecommendation, err error) (*pb.PromotionRecommendationResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.PromotionRecommendationResponse{Value: recommendation(v)}, nil
}
func recommendations(p operation.Page[entity.PromotionRecommendation], err error) (*pb.PromotionRecommendationsResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	items := make([]*pb.PromotionRecommendation, 0, len(p.Items))
	for i := range p.Items {
		items = append(items, recommendation(&p.Items[i]))
	}
	return &pb.PromotionRecommendationsResponse{Items: items, Page: &commonv1.PageInfo{Total: p.Total, Limit: int32(p.Limit), Offset: int32(p.Offset)}}, nil
}
func decision(v *entity.PromotionDecision) *pb.PromotionDecision {
	if v == nil {
		return nil
	}
	return &pb.PromotionDecision{Id: v.ID, RecommendationId: v.RecommendationID, StepNumber: v.StepNumber,
		Decision: string(v.Decision), DecidedByPrincipalId: v.DecidedByPrincipalID, DecidedAt: timestamppb.New(v.DecidedAt),
		RationaleDocumentVersionReferenceId: v.RationaleDocumentVersionReferenceID, RationaleDigest: v.RationaleDigest, IdempotencyKey: v.IdempotencyKey}
}
func decisionResult(v *entity.PromotionDecision, err error) (*pb.PromotionDecisionResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.PromotionDecisionResponse{Value: decision(v)}, nil
}
func decisions(p operation.Page[entity.PromotionDecision], err error) (*pb.PromotionDecisionsResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	items := make([]*pb.PromotionDecision, 0, len(p.Items))
	for i := range p.Items {
		items = append(items, decision(&p.Items[i]))
	}
	return &pb.PromotionDecisionsResponse{Items: items, Page: &commonv1.PageInfo{Total: p.Total, Limit: int32(p.Limit), Offset: int32(p.Offset)}}, nil
}
func (s *server) CreatePromotionRecommendation(ctx context.Context, r *pb.CreatePromotionRecommendationRequest) (*pb.PromotionRecommendationResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	in := r.Input
	effective, err := response.Timestamp(in.RecommendedEffectiveDate)
	if err != nil {
		return nil, response.Error(err)
	}
	applied, err := response.OptionalTimestamp(in.AppliedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	return recommendationResult(s.service.CreatePromotionRecommendation(ctx, service.CreatePromotionRecommendationInput{
		RecommendationNumber: in.GetRecommendationNumber(), WorkerID: in.GetWorkerId(), SourceAssignmentID: in.GetSourceAssignmentId(),
		TargetPositionID: in.GetTargetPositionId(), AssessmentID: in.GetAssessmentId(), SupersedesRecommendationID: in.SupersedesRecommendationId,
		RecommendedEffectiveDate: effective, RationaleDocumentVersionReferenceID: in.GetRationaleDocumentVersionReferenceId(),
		RationaleDigest: in.RationaleDigest, Status: in.Status, ApprovalDecisionID: in.ApprovalDecisionId,
		AppliedAssignmentID: in.AppliedAssignmentId, AppliedLifecycleEventID: in.AppliedLifecycleEventId, AppliedAt: applied,
		IdempotencyKey: in.GetIdempotencyKey()}))
}
func (s *server) UpdatePromotionRecommendation(ctx context.Context, r *pb.UpdatePromotionRecommendationRequest) (*pb.PromotionRecommendationResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	fields, err := response.ValidateMask(r.UpdateMask, "recommendation_number", "worker_id", "source_assignment_id", "target_position_id",
		"assessment_id", "supersedes_recommendation_id", "recommended_effective_date", "rationale_document_version_reference_id",
		"rationale_digest", "status", "approval_decision_id", "applied_assignment_id", "applied_lifecycle_event_id", "applied_at", "idempotency_key")
	if err != nil {
		return nil, response.Error(err)
	}
	in := r.Input
	effective, err := timeField(fields["recommended_effective_date"], in.RecommendedEffectiveDate)
	if err != nil {
		return nil, response.Error(err)
	}
	applied, err := timeField(fields["applied_at"], in.AppliedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	return recommendationResult(s.service.UpdatePromotionRecommendation(ctx, r.Id, service.UpdatePromotionRecommendationInput{
		RecommendationNumber:       operation.Field[string]{Set: fields["recommendation_number"], Value: in.RecommendationNumber},
		WorkerID:                   operation.Field[string]{Set: fields["worker_id"], Value: in.WorkerId},
		SourceAssignmentID:         operation.Field[string]{Set: fields["source_assignment_id"], Value: in.SourceAssignmentId},
		TargetPositionID:           operation.Field[string]{Set: fields["target_position_id"], Value: in.TargetPositionId},
		AssessmentID:               operation.Field[string]{Set: fields["assessment_id"], Value: in.AssessmentId},
		SupersedesRecommendationID: operation.Field[string]{Set: fields["supersedes_recommendation_id"], Value: in.SupersedesRecommendationId},
		RecommendedEffectiveDate:   effective, RationaleDocumentVersionReferenceID: operation.Field[string]{Set: fields["rationale_document_version_reference_id"], Value: in.RationaleDocumentVersionReferenceId},
		RationaleDigest: bytesField(fields["rationale_digest"], in.RationaleDigest), Status: operation.Field[string]{Set: fields["status"], Value: in.Status},
		ApprovalDecisionID:      operation.Field[string]{Set: fields["approval_decision_id"], Value: in.ApprovalDecisionId},
		AppliedAssignmentID:     operation.Field[string]{Set: fields["applied_assignment_id"], Value: in.AppliedAssignmentId},
		AppliedLifecycleEventID: operation.Field[int64]{Set: fields["applied_lifecycle_event_id"], Value: in.AppliedLifecycleEventId},
		AppliedAt:               applied, IdempotencyKey: operation.Field[string]{Set: fields["idempotency_key"], Value: in.IdempotencyKey}}))
}
func (s *server) DeletePromotionRecommendation(ctx context.Context, r *commonv1.IDRequest) (*commonv1.EmptyResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	if err := s.service.DeletePromotionRecommendation(ctx, r.Id); err != nil {
		return nil, response.Error(err)
	}
	return &commonv1.EmptyResponse{}, nil
}
func (s *server) GetPromotionRecommendationByID(ctx context.Context, r *commonv1.IDRequest) (*pb.PromotionRecommendationResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return recommendationResult(s.service.GetPromotionRecommendationByID(ctx, r.Id))
}
func (s *server) GetAllPromotionRecommendations(ctx context.Context, r *commonv1.PageRequest) (*pb.PromotionRecommendationsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return recommendations(s.service.GetAllPromotionRecommendations(ctx, int(r.Limit), int(r.Offset)))
}
func (s *server) FindPromotionRecommendationByRecommendationNumber(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.PromotionRecommendationsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return recommendations(s.service.FindPromotionRecommendationByRecommendationNumber(ctx, r.GetValue(), int(r.Limit), int(r.Offset)))
}
func (s *server) FindPromotionRecommendationByWorkerID(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.PromotionRecommendationsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return recommendations(s.service.FindPromotionRecommendationByWorkerID(ctx, r.GetValue(), int(r.Limit), int(r.Offset)))
}
func (s *server) FindPromotionRecommendationBySourceAssignmentID(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.PromotionRecommendationsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return recommendations(s.service.FindPromotionRecommendationBySourceAssignmentID(ctx, r.GetValue(), int(r.Limit), int(r.Offset)))
}
func (s *server) FindPromotionRecommendationByTargetPositionID(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.PromotionRecommendationsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return recommendations(s.service.FindPromotionRecommendationByTargetPositionID(ctx, r.GetValue(), int(r.Limit), int(r.Offset)))
}
func (s *server) FindPromotionRecommendationByAssessmentID(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.PromotionRecommendationsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return recommendations(s.service.FindPromotionRecommendationByAssessmentID(ctx, r.GetValue(), int(r.Limit), int(r.Offset)))
}
func (s *server) FindPromotionRecommendationBySupersedesRecommendationID(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.PromotionRecommendationsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return recommendations(s.service.FindPromotionRecommendationBySupersedesRecommendationID(ctx, r.Value, int(r.Limit), int(r.Offset)))
}
func (s *server) FindPromotionRecommendationByRationaleDocumentVersionReferenceID(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.PromotionRecommendationsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return recommendations(s.service.FindPromotionRecommendationByRationaleDocumentVersionReferenceID(ctx, r.GetValue(), int(r.Limit), int(r.Offset)))
}
func (s *server) FindPromotionRecommendationByRecommendedByPrincipalID(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.PromotionRecommendationsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return recommendations(s.service.FindPromotionRecommendationByRecommendedByPrincipalID(ctx, r.GetValue(), int(r.Limit), int(r.Offset)))
}
func (s *server) FindPromotionRecommendationByStatus(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.PromotionRecommendationsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return recommendations(s.service.FindPromotionRecommendationByStatus(ctx, r.GetValue(), int(r.Limit), int(r.Offset)))
}
func (s *server) FindPromotionRecommendationByApprovalDecisionID(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.PromotionRecommendationsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return recommendations(s.service.FindPromotionRecommendationByApprovalDecisionID(ctx, r.Value, int(r.Limit), int(r.Offset)))
}
func (s *server) FindPromotionRecommendationByAppliedAssignmentID(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.PromotionRecommendationsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return recommendations(s.service.FindPromotionRecommendationByAppliedAssignmentID(ctx, r.Value, int(r.Limit), int(r.Offset)))
}
func (s *server) FindPromotionRecommendationByAppliedLifecycleEventID(ctx context.Context, r *commonv1.Int64FinderRequest) (*pb.PromotionRecommendationsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return recommendations(s.service.FindPromotionRecommendationByAppliedLifecycleEventID(ctx, r.Value, int(r.Limit), int(r.Offset)))
}
func (s *server) FindPromotionRecommendationByIdempotencyKey(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.PromotionRecommendationsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return recommendations(s.service.FindPromotionRecommendationByIdempotencyKey(ctx, r.GetValue(), int(r.Limit), int(r.Offset)))
}
func (s *server) CreatePromotionDecision(ctx context.Context, r *pb.CreatePromotionDecisionRequest) (*pb.PromotionDecisionResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	in := r.Input
	return decisionResult(s.service.CreatePromotionDecision(ctx, service.CreatePromotionDecisionInput{
		RecommendationID: in.GetRecommendationId(), StepNumber: in.GetStepNumber(), Decision: entity.ApprovalDecision(in.GetDecision()),
		RationaleDocumentVersionReferenceID: in.GetRationaleDocumentVersionReferenceId(), RationaleDigest: in.RationaleDigest, IdempotencyKey: in.GetIdempotencyKey()}))
}
func (s *server) UpdatePromotionDecision(ctx context.Context, r *pb.UpdatePromotionDecisionRequest) (*pb.PromotionDecisionResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	fields, err := response.ValidateMask(r.UpdateMask, "recommendation_id", "step_number", "decision", "rationale_document_version_reference_id", "rationale_digest", "idempotency_key")
	if err != nil {
		return nil, response.Error(err)
	}
	in := r.Input
	var choice *entity.ApprovalDecision
	if in.Decision != nil {
		value := entity.ApprovalDecision(*in.Decision)
		choice = &value
	}
	return decisionResult(s.service.UpdatePromotionDecision(ctx, r.Id, service.UpdatePromotionDecisionInput{
		RecommendationID: operation.Field[string]{Set: fields["recommendation_id"], Value: in.RecommendationId},
		StepNumber:       operation.Field[int32]{Set: fields["step_number"], Value: in.StepNumber}, Decision: operation.Field[entity.ApprovalDecision]{Set: fields["decision"], Value: choice},
		RationaleDocumentVersionReferenceID: operation.Field[string]{Set: fields["rationale_document_version_reference_id"], Value: in.RationaleDocumentVersionReferenceId},
		RationaleDigest:                     bytesField(fields["rationale_digest"], in.RationaleDigest), IdempotencyKey: operation.Field[string]{Set: fields["idempotency_key"], Value: in.IdempotencyKey}}))
}
func (s *server) DeletePromotionDecision(ctx context.Context, r *commonv1.IDRequest) (*commonv1.EmptyResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	if err := s.service.DeletePromotionDecision(ctx, r.Id); err != nil {
		return nil, response.Error(err)
	}
	return &commonv1.EmptyResponse{}, nil
}
func (s *server) GetPromotionDecisionByID(ctx context.Context, r *commonv1.IDRequest) (*pb.PromotionDecisionResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return decisionResult(s.service.GetPromotionDecisionByID(ctx, r.Id))
}
func (s *server) GetAllPromotionDecisions(ctx context.Context, r *commonv1.PageRequest) (*pb.PromotionDecisionsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return decisions(s.service.GetAllPromotionDecisions(ctx, int(r.Limit), int(r.Offset)))
}
func (s *server) FindPromotionDecisionByRecommendationID(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.PromotionDecisionsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return decisions(s.service.FindPromotionDecisionByRecommendationID(ctx, r.GetValue(), int(r.Limit), int(r.Offset)))
}
func (s *server) FindPromotionDecisionByStepNumber(ctx context.Context, r *commonv1.Int32FinderRequest) (*pb.PromotionDecisionsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return decisions(s.service.FindPromotionDecisionByStepNumber(ctx, r.GetValue(), int(r.Limit), int(r.Offset)))
}
func (s *server) FindPromotionDecisionByDecidedByPrincipalID(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.PromotionDecisionsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return decisions(s.service.FindPromotionDecisionByDecidedByPrincipalID(ctx, r.GetValue(), int(r.Limit), int(r.Offset)))
}
func (s *server) FindPromotionDecisionByRationaleDocumentVersionReferenceID(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.PromotionDecisionsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return decisions(s.service.FindPromotionDecisionByRationaleDocumentVersionReferenceID(ctx, r.GetValue(), int(r.Limit), int(r.Offset)))
}
func (s *server) FindPromotionDecisionByIdempotencyKey(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.PromotionDecisionsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return decisions(s.service.FindPromotionDecisionByIdempotencyKey(ctx, r.GetValue(), int(r.Limit), int(r.Offset)))
}
