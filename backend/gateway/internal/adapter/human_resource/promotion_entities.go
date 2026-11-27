package human_resource

import (
	"context"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	promotion "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/promotion/v1"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func toModelPromotionRecommendation(row *promotion.PromotionRecommendation) (*model.PromotionRecommendation, error) {
	if row == nil {
		return nil, middleware.ErrUnavailable("invalid human resource promotion-recommendation response", nil)
	}
	if id, err := uuid.Parse(row.GetId()); err != nil || id == uuid.Nil {
		return nil, middleware.ErrUnavailable("invalid human resource promotion-recommendation ID", nil)
	}
	out := &model.PromotionRecommendation{ID: nodeID("promotion-recommendation", row.GetId()), RawID: row.GetId()}
	out.RecommendationNumber = row.GetRecommendationNumber()
	out.WorkerID = row.GetWorkerId()
	out.SourceAssignmentID = row.GetSourceAssignmentId()
	out.TargetPositionID = row.GetTargetPositionId()
	out.AssessmentID = row.GetAssessmentId()
	out.SupersedesRecommendationID = row.SupersedesRecommendationId
	out.RecommendedEffectiveDate = pbTimePtr(row.GetRecommendedEffectiveDate())
	out.RationaleDocumentVersionReferenceID = row.GetRationaleDocumentVersionReferenceId()
	out.RationaleDigest = pbBase64StringValue(row.GetRationaleDigest())
	out.RecommendedByPrincipalID = row.GetRecommendedByPrincipalId()
	out.RecommendedAt = pbTimePtr(row.GetRecommendedAt())
	out.Status = row.GetStatus()
	out.ApprovalDecisionID = row.ApprovalDecisionId
	out.AppliedAssignmentID = row.AppliedAssignmentId
	out.AppliedLifecycleEventID = pbInt64StrPtr(row.AppliedLifecycleEventId)
	out.AppliedAt = pbTimePtr(row.GetAppliedAt())
	out.IdempotencyKey = row.GetIdempotencyKey()
	return out, nil
}

// PromotionRecommendation membaca detail promotion-recommendation berdasarkan ID.
func (a *Adapter) PromotionRecommendation(ctx context.Context, rawID string) (*model.PromotionRecommendation, error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_promotion_recommendations", "human_resource.promotion_recommendations.read")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "promotion-recommendation", "promotion-recommendation")
	if err != nil {
		return nil, err
	}
	if a.promotion == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.promotion.GetPromotionRecommendationByID(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapReadError(err, "promotion-recommendation")
	}
	return toModelPromotionRecommendation(resp.GetValue())
}

// PromotionRecommendations mendaftar dan memfilter promotion-recommendation dengan paginasi.
func (a *Adapter) PromotionRecommendations(ctx context.Context, limit, offset int, f_recommendation_number *string, f_worker_id *string, f_source_assignment_id *string, f_target_position_id *string, f_assessment_id *string, f_supersedes_recommendation_id *string, f_rationale_document_version_reference_id *string, f_recommended_by_principal_id *string, f_status *string, f_approval_decision_id *string, f_applied_assignment_id *string, f_applied_lifecycle_event_id *string, f_idempotency_key *string) (*Page[model.PromotionRecommendation], error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_promotion_recommendations", "human_resource.promotion_recommendations.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid promotion-recommendation pagination", nil)
	}
	if a.promotion == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	idx, err := pickSingleFilter(f_recommendation_number != nil, f_worker_id != nil, f_source_assignment_id != nil, f_target_position_id != nil, f_assessment_id != nil, f_supersedes_recommendation_id != nil, f_rationale_document_version_reference_id != nil, f_recommended_by_principal_id != nil, f_status != nil, f_approval_decision_id != nil, f_applied_assignment_id != nil, f_applied_lifecycle_event_id != nil, f_idempotency_key != nil)
	if err != nil {
		return nil, err
	}
	var resp *promotion.PromotionRecommendationsResponse
	switch idx {
	case 0:
		v, fErr := validateFilterString(f_recommendation_number, "recommendation_number")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.promotion.FindPromotionRecommendationByRecommendationNumber(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 1:
		v, fErr := validateFilterString(f_worker_id, "worker_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.promotion.FindPromotionRecommendationByWorkerID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 2:
		v, fErr := validateFilterString(f_source_assignment_id, "source_assignment_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.promotion.FindPromotionRecommendationBySourceAssignmentID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 3:
		v, fErr := validateFilterString(f_target_position_id, "target_position_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.promotion.FindPromotionRecommendationByTargetPositionID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 4:
		v, fErr := validateFilterString(f_assessment_id, "assessment_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.promotion.FindPromotionRecommendationByAssessmentID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 5:
		v, fErr := validateFilterString(f_supersedes_recommendation_id, "supersedes_recommendation_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.promotion.FindPromotionRecommendationBySupersedesRecommendationID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 6:
		v, fErr := validateFilterString(f_rationale_document_version_reference_id, "rationale_document_version_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.promotion.FindPromotionRecommendationByRationaleDocumentVersionReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 7:
		v, fErr := validateFilterString(f_recommended_by_principal_id, "recommended_by_principal_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.promotion.FindPromotionRecommendationByRecommendedByPrincipalID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 8:
		v, fErr := validateFilterString(f_status, "status")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.promotion.FindPromotionRecommendationByStatus(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 9:
		v, fErr := validateFilterString(f_approval_decision_id, "approval_decision_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.promotion.FindPromotionRecommendationByApprovalDecisionID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 10:
		v, fErr := validateFilterString(f_applied_assignment_id, "applied_assignment_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.promotion.FindPromotionRecommendationByAppliedAssignmentID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 11:
		v, parseErr := strconv.ParseInt(strings.TrimSpace(*f_applied_lifecycle_event_id), 10, 64)
		if parseErr != nil {
			return nil, middleware.ErrInvalidArgument("invalid applied_lifecycle_event_id", nil)
		}
		resp, err = a.promotion.FindPromotionRecommendationByAppliedLifecycleEventID(rpcCtx, &common.Int64FinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 12:
		v, fErr := validateFilterString(f_idempotency_key, "idempotency_key")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.promotion.FindPromotionRecommendationByIdempotencyKey(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	default:
		resp, err = a.promotion.GetAllPromotionRecommendations(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "promotion-recommendation")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "promotion-recommendation", toModelPromotionRecommendation)
}

// CreatePromotionRecommendation membuat promotion-recommendation baru melalui HRM.
func (a *Adapter) CreatePromotionRecommendation(ctx context.Context, input model.CreatePromotionRecommendationInput) (*model.PromotionRecommendation, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_promotion_recommendations", "create", "human_resource.promotion_recommendations.create")
	if err != nil {
		return nil, err
	}
	if a.promotion == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &promotion.PromotionRecommendationInput{}
	patch.RecommendationNumber = input.RecommendationNumber
	patch.WorkerId = input.WorkerID
	patch.SourceAssignmentId = input.SourceAssignmentID
	patch.TargetPositionId = input.TargetPositionID
	patch.AssessmentId = input.AssessmentID
	patch.SupersedesRecommendationId = input.SupersedesRecommendationID
	patch.RecommendedEffectiveDate = optionalTimestamp(input.RecommendedEffectiveDate)
	patch.RationaleDocumentVersionReferenceId = input.RationaleDocumentVersionReferenceID
	if b, bErr := optionalBase64Bytes(input.RationaleDigest, "rationale_digest"); bErr != nil {
		return nil, bErr
	} else if b != nil {
		patch.RationaleDigest = b
	}
	patch.Status = input.Status
	patch.ApprovalDecisionId = input.ApprovalDecisionID
	patch.AppliedAssignmentId = input.AppliedAssignmentID
	if v, iErr := optionalInt64FromStr(input.AppliedLifecycleEventID, "applied_lifecycle_event_id"); iErr != nil {
		return nil, iErr
	} else {
		patch.AppliedLifecycleEventId = v
	}
	patch.AppliedAt = optionalTimestamp(input.AppliedAt)
	patch.IdempotencyKey = input.IdempotencyKey
	resp, err := a.promotion.CreatePromotionRecommendation(rpcCtx, &promotion.CreatePromotionRecommendationRequest{Input: patch})
	if err != nil {
		return nil, mapHRMWriteError(err, "promotion-recommendation")
	}
	return toModelPromotionRecommendation(resp.GetValue())
}

// UpdatePromotionRecommendation memperbarui promotion-recommendation berdasarkan FieldMask.
func (a *Adapter) UpdatePromotionRecommendation(ctx context.Context, rawID string, input model.UpdatePromotionRecommendationInput) (*model.PromotionRecommendation, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_promotion_recommendations", "update", "human_resource.promotion_recommendations.update")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "promotion-recommendation", "promotion-recommendation")
	if err != nil {
		return nil, err
	}
	if a.promotion == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &promotion.PromotionRecommendationInput{}
	mask := &fieldmaskpb.FieldMask{}
	if err := applyStringMask(mask, "recommendation_number", input.RecommendationNumber, &patch.RecommendationNumber, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "worker_id", input.WorkerID, &patch.WorkerId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "source_assignment_id", input.SourceAssignmentID, &patch.SourceAssignmentId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "target_position_id", input.TargetPositionID, &patch.TargetPositionId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "assessment_id", input.AssessmentID, &patch.AssessmentId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "supersedes_recommendation_id", input.SupersedesRecommendationID, &patch.SupersedesRecommendationId, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "recommended_effective_date", input.RecommendedEffectiveDate, &patch.RecommendedEffectiveDate, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "rationale_document_version_reference_id", input.RationaleDocumentVersionReferenceID, &patch.RationaleDocumentVersionReferenceId, true); err != nil {
		return nil, err
	}
	if err := applyBytesMask(mask, "rationale_digest", input.RationaleDigest, &patch.RationaleDigest, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "status", input.Status, &patch.Status, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "approval_decision_id", input.ApprovalDecisionID, &patch.ApprovalDecisionId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "applied_assignment_id", input.AppliedAssignmentID, &patch.AppliedAssignmentId, true); err != nil {
		return nil, err
	}
	if err := applyInt64Mask(mask, "applied_lifecycle_event_id", input.AppliedLifecycleEventID, &patch.AppliedLifecycleEventId, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "applied_at", input.AppliedAt, &patch.AppliedAt, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "idempotency_key", input.IdempotencyKey, &patch.IdempotencyKey, true); err != nil {
		return nil, err
	}
	if len(mask.Paths) == 0 {
		return nil, middleware.ErrInvalidArgument("promotion-recommendation update is empty", nil)
	}
	resp, err := a.promotion.UpdatePromotionRecommendation(rpcCtx, &promotion.UpdatePromotionRecommendationRequest{Id: id, Input: patch, UpdateMask: mask})
	if err != nil {
		return nil, mapHRMWriteError(err, "promotion-recommendation")
	}
	return toModelPromotionRecommendation(resp.GetValue())
}

// DeletePromotionRecommendation menghapus promotion-recommendation berdasarkan ID.
func (a *Adapter) DeletePromotionRecommendation(ctx context.Context, rawID string) (*bool, error) {
	if a == nil || a.promotion == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	return deleteByUUID(ctx, a, rawID, "hr_promotion_recommendations", "human_resource.promotion_recommendations.delete", "promotion-recommendation", func(c context.Context, req *common.IDRequest) (*common.EmptyResponse, error) {
		return a.promotion.DeletePromotionRecommendation(c, req)
	})
}

func toModelPromotionDecision(row *promotion.PromotionDecision) (*model.PromotionDecision, error) {
	if row == nil {
		return nil, middleware.ErrUnavailable("invalid human resource promotion-decision response", nil)
	}
	if id, err := uuid.Parse(row.GetId()); err != nil || id == uuid.Nil {
		return nil, middleware.ErrUnavailable("invalid human resource promotion-decision ID", nil)
	}
	out := &model.PromotionDecision{ID: nodeID("promotion-decision", row.GetId()), RawID: row.GetId()}
	out.RecommendationID = row.GetRecommendationId()
	out.StepNumber = int(row.GetStepNumber())
	out.Decision = row.GetDecision()
	out.DecidedByPrincipalID = row.GetDecidedByPrincipalId()
	out.DecidedAt = pbTimePtr(row.GetDecidedAt())
	out.RationaleDocumentVersionReferenceID = row.GetRationaleDocumentVersionReferenceId()
	out.RationaleDigest = pbBase64StringValue(row.GetRationaleDigest())
	out.IdempotencyKey = row.GetIdempotencyKey()
	return out, nil
}

// PromotionDecision membaca detail promotion-decision berdasarkan ID.
func (a *Adapter) PromotionDecision(ctx context.Context, rawID string) (*model.PromotionDecision, error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_promotion_decisions", "human_resource.promotion_decisions.read")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "promotion-decision", "promotion-decision")
	if err != nil {
		return nil, err
	}
	if a.promotion == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.promotion.GetPromotionDecisionByID(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapReadError(err, "promotion-decision")
	}
	return toModelPromotionDecision(resp.GetValue())
}

// PromotionDecisions mendaftar dan memfilter promotion-decision dengan paginasi.
func (a *Adapter) PromotionDecisions(ctx context.Context, limit, offset int, f_recommendation_id *string, f_step_number *int, f_decided_by_principal_id *string, f_rationale_document_version_reference_id *string, f_idempotency_key *string) (*Page[model.PromotionDecision], error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_promotion_decisions", "human_resource.promotion_decisions.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid promotion-decision pagination", nil)
	}
	if a.promotion == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	idx, err := pickSingleFilter(f_recommendation_id != nil, f_step_number != nil, f_decided_by_principal_id != nil, f_rationale_document_version_reference_id != nil, f_idempotency_key != nil)
	if err != nil {
		return nil, err
	}
	var resp *promotion.PromotionDecisionsResponse
	switch idx {
	case 0:
		v, fErr := validateFilterString(f_recommendation_id, "recommendation_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.promotion.FindPromotionDecisionByRecommendationID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 1:
		v := int32(*f_step_number)
		resp, err = a.promotion.FindPromotionDecisionByStepNumber(rpcCtx, &common.Int32FinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 2:
		v, fErr := validateFilterString(f_decided_by_principal_id, "decided_by_principal_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.promotion.FindPromotionDecisionByDecidedByPrincipalID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 3:
		v, fErr := validateFilterString(f_rationale_document_version_reference_id, "rationale_document_version_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.promotion.FindPromotionDecisionByRationaleDocumentVersionReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 4:
		v, fErr := validateFilterString(f_idempotency_key, "idempotency_key")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.promotion.FindPromotionDecisionByIdempotencyKey(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	default:
		resp, err = a.promotion.GetAllPromotionDecisions(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "promotion-decision")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "promotion-decision", toModelPromotionDecision)
}

// CreatePromotionDecision membuat promotion-decision baru melalui HRM.
func (a *Adapter) CreatePromotionDecision(ctx context.Context, input model.CreatePromotionDecisionInput) (*model.PromotionDecision, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_promotion_decisions", "create", "human_resource.promotion_decisions.create")
	if err != nil {
		return nil, err
	}
	if a.promotion == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &promotion.PromotionDecisionInput{}
	patch.RecommendationId = input.RecommendationID
	patch.StepNumber = optionalInt32Ptr(input.StepNumber)
	patch.Decision = input.Decision
	patch.RationaleDocumentVersionReferenceId = input.RationaleDocumentVersionReferenceID
	if b, bErr := optionalBase64Bytes(input.RationaleDigest, "rationale_digest"); bErr != nil {
		return nil, bErr
	} else if b != nil {
		patch.RationaleDigest = b
	}
	patch.IdempotencyKey = input.IdempotencyKey
	resp, err := a.promotion.CreatePromotionDecision(rpcCtx, &promotion.CreatePromotionDecisionRequest{Input: patch})
	if err != nil {
		return nil, mapHRMWriteError(err, "promotion-decision")
	}
	return toModelPromotionDecision(resp.GetValue())
}

// UpdatePromotionDecision memperbarui promotion-decision berdasarkan FieldMask.
func (a *Adapter) UpdatePromotionDecision(ctx context.Context, rawID string, input model.UpdatePromotionDecisionInput) (*model.PromotionDecision, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_promotion_decisions", "update", "human_resource.promotion_decisions.update")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "promotion-decision", "promotion-decision")
	if err != nil {
		return nil, err
	}
	if a.promotion == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &promotion.PromotionDecisionInput{}
	mask := &fieldmaskpb.FieldMask{}
	if err := applyStringMask(mask, "recommendation_id", input.RecommendationID, &patch.RecommendationId, true); err != nil {
		return nil, err
	}
	if err := applyInt32Mask(mask, "step_number", input.StepNumber, &patch.StepNumber, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "decision", input.Decision, &patch.Decision, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "rationale_document_version_reference_id", input.RationaleDocumentVersionReferenceID, &patch.RationaleDocumentVersionReferenceId, true); err != nil {
		return nil, err
	}
	if err := applyBytesMask(mask, "rationale_digest", input.RationaleDigest, &patch.RationaleDigest, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "idempotency_key", input.IdempotencyKey, &patch.IdempotencyKey, true); err != nil {
		return nil, err
	}
	if len(mask.Paths) == 0 {
		return nil, middleware.ErrInvalidArgument("promotion-decision update is empty", nil)
	}
	resp, err := a.promotion.UpdatePromotionDecision(rpcCtx, &promotion.UpdatePromotionDecisionRequest{Id: id, Input: patch, UpdateMask: mask})
	if err != nil {
		return nil, mapHRMWriteError(err, "promotion-decision")
	}
	return toModelPromotionDecision(resp.GetValue())
}

// DeletePromotionDecision menghapus promotion-decision berdasarkan ID.
func (a *Adapter) DeletePromotionDecision(ctx context.Context, rawID string) (*bool, error) {
	if a == nil || a.promotion == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	return deleteByUUID(ctx, a, rawID, "hr_promotion_decisions", "human_resource.promotion_decisions.delete", "promotion-decision", func(c context.Context, req *common.IDRequest) (*common.EmptyResponse, error) {
		return a.promotion.DeletePromotionDecision(c, req)
	})
}
