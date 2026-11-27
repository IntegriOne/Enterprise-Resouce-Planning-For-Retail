package human_resource

import (
	"context"

	"github.com/google/uuid"
	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	succession "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/succession/v1"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func toModelSuccessionPlan(row *succession.SuccessionPlan) (*model.SuccessionPlan, error) {
	if row == nil {
		return nil, middleware.ErrUnavailable("invalid human resource succession-plan response", nil)
	}
	if id, err := uuid.Parse(row.GetId()); err != nil || id == uuid.Nil {
		return nil, middleware.ErrUnavailable("invalid human resource succession-plan ID", nil)
	}
	out := &model.SuccessionPlan{ID: nodeID("succession-plan", row.GetId()), RawID: row.GetId()}
	out.PositionID = row.GetPositionId()
	out.PlanCode = row.GetPlanCode()
	out.PlanVersion = int(row.GetPlanVersion())
	out.ReviewDate = pbTimePtr(row.GetReviewDate())
	out.ConfidentialityClass = row.GetConfidentialityClass()
	out.ApprovedByPrincipalID = row.ApprovedByPrincipalId
	out.ApprovedAt = pbTimePtr(row.GetApprovedAt())
	out.SuccessionRisk = row.SuccessionRisk
	out.ApprovalMandateReferenceID = row.ApprovalMandateReferenceId
	out.CreatedByPrincipalID = row.CreatedByPrincipalId
	out.CreatedAt = pbTimePtr(row.GetCreatedAt())
	return out, nil
}

// SuccessionPlan membaca detail succession-plan berdasarkan ID.
func (a *Adapter) SuccessionPlan(ctx context.Context, rawID string) (*model.SuccessionPlan, error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_succession_plans", "human_resource.succession_plans.read")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "succession-plan", "succession-plan")
	if err != nil {
		return nil, err
	}
	if a.succession == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.succession.GetSuccessionPlanByID(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapReadError(err, "succession-plan")
	}
	return toModelSuccessionPlan(resp.GetValue())
}

// SuccessionPlans mendaftar dan memfilter succession-plan dengan paginasi.
func (a *Adapter) SuccessionPlans(ctx context.Context, limit, offset int, f_position_id *string, f_plan_code *string, f_approved_by_principal_id *string) (*Page[model.SuccessionPlan], error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_succession_plans", "human_resource.succession_plans.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid succession-plan pagination", nil)
	}
	if a.succession == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	idx, err := pickSingleFilter(f_position_id != nil, f_plan_code != nil, f_approved_by_principal_id != nil)
	if err != nil {
		return nil, err
	}
	var resp *succession.SuccessionPlansResponse
	switch idx {
	case 0:
		v, fErr := validateFilterString(f_position_id, "position_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.succession.FindSuccessionPlanByPositionID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 1:
		v, fErr := validateFilterString(f_plan_code, "plan_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.succession.FindSuccessionPlanByPlanCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 2:
		v, fErr := validateFilterString(f_approved_by_principal_id, "approved_by_principal_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.succession.FindSuccessionPlanByApprovedByPrincipalID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	default:
		resp, err = a.succession.GetAllSuccessionPlans(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "succession-plan")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "succession-plan", toModelSuccessionPlan)
}

// CreateSuccessionPlan membuat succession-plan baru melalui HRM.
func (a *Adapter) CreateSuccessionPlan(ctx context.Context, input model.CreateSuccessionPlanInput) (*model.SuccessionPlan, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_succession_plans", "create", "human_resource.succession_plans.create")
	if err != nil {
		return nil, err
	}
	if a.succession == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &succession.SuccessionPlanInput{}
	patch.PositionId = input.PositionID
	patch.PlanCode = input.PlanCode
	patch.PlanVersion = optionalInt32Ptr(input.PlanVersion)
	patch.ReviewDate = optionalTimestamp(input.ReviewDate)
	patch.ConfidentialityClass = input.ConfidentialityClass
	patch.ApprovedByPrincipalId = input.ApprovedByPrincipalID
	patch.ApprovedAt = optionalTimestamp(input.ApprovedAt)
	patch.SuccessionRisk = input.SuccessionRisk
	resp, err := a.succession.CreateSuccessionPlan(rpcCtx, &succession.CreateSuccessionPlanRequest{Input: patch})
	if err != nil {
		return nil, mapHRMWriteError(err, "succession-plan")
	}
	return toModelSuccessionPlan(resp.GetValue())
}

// UpdateSuccessionPlan memperbarui succession-plan berdasarkan FieldMask.
func (a *Adapter) UpdateSuccessionPlan(ctx context.Context, rawID string, input model.UpdateSuccessionPlanInput) (*model.SuccessionPlan, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_succession_plans", "update", "human_resource.succession_plans.update")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "succession-plan", "succession-plan")
	if err != nil {
		return nil, err
	}
	if a.succession == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &succession.SuccessionPlanInput{}
	mask := &fieldmaskpb.FieldMask{}
	if err := applyStringMask(mask, "position_id", input.PositionID, &patch.PositionId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "plan_code", input.PlanCode, &patch.PlanCode, true); err != nil {
		return nil, err
	}
	if err := applyInt32Mask(mask, "plan_version", input.PlanVersion, &patch.PlanVersion, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "review_date", input.ReviewDate, &patch.ReviewDate, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "confidentiality_class", input.ConfidentialityClass, &patch.ConfidentialityClass, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "approved_by_principal_id", input.ApprovedByPrincipalID, &patch.ApprovedByPrincipalId, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "approved_at", input.ApprovedAt, &patch.ApprovedAt, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "succession_risk", input.SuccessionRisk, &patch.SuccessionRisk, true); err != nil {
		return nil, err
	}
	if len(mask.Paths) == 0 {
		return nil, middleware.ErrInvalidArgument("succession-plan update is empty", nil)
	}
	resp, err := a.succession.UpdateSuccessionPlan(rpcCtx, &succession.UpdateSuccessionPlanRequest{Id: id, Input: patch, UpdateMask: mask})
	if err != nil {
		return nil, mapHRMWriteError(err, "succession-plan")
	}
	return toModelSuccessionPlan(resp.GetValue())
}

// DeleteSuccessionPlan menghapus succession-plan berdasarkan ID.
func (a *Adapter) DeleteSuccessionPlan(ctx context.Context, rawID string) (*bool, error) {
	if a == nil || a.succession == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	return deleteByUUID(ctx, a, rawID, "hr_succession_plans", "human_resource.succession_plans.delete", "succession-plan", func(c context.Context, req *common.IDRequest) (*common.EmptyResponse, error) {
		return a.succession.DeleteSuccessionPlan(c, req)
	})
}

func toModelSuccessionCandidate(row *succession.SuccessionCandidate) (*model.SuccessionCandidate, error) {
	if row == nil {
		return nil, middleware.ErrUnavailable("invalid human resource succession-candidate response", nil)
	}
	if id, err := uuid.Parse(row.GetId()); err != nil || id == uuid.Nil {
		return nil, middleware.ErrUnavailable("invalid human resource succession-candidate ID", nil)
	}
	out := &model.SuccessionCandidate{ID: nodeID("succession-candidate", row.GetId()), RawID: row.GetId()}
	out.SuccessionPlanID = row.GetSuccessionPlanId()
	out.CandidateWorkerID = row.GetCandidateWorkerId()
	out.ReadinessCode = row.GetReadinessCode()
	out.ReadinessAssessedAt = pbTimePtr(row.GetReadinessAssessedAt())
	out.ReadinessRationale = row.ReadinessRationale
	out.TalentAssessmentID = row.TalentAssessmentId
	out.DevelopmentPlanID = row.DevelopmentPlanId
	out.RiskOfLoss = row.RiskOfLoss
	return out, nil
}

// SuccessionCandidate membaca detail succession-candidate berdasarkan ID.
func (a *Adapter) SuccessionCandidate(ctx context.Context, rawID string) (*model.SuccessionCandidate, error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_succession_candidates", "human_resource.succession_candidates.read")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "succession-candidate", "succession-candidate")
	if err != nil {
		return nil, err
	}
	if a.succession == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.succession.GetSuccessionCandidateByID(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapReadError(err, "succession-candidate")
	}
	return toModelSuccessionCandidate(resp.GetValue())
}

// SuccessionCandidates mendaftar dan memfilter succession-candidate dengan paginasi.
func (a *Adapter) SuccessionCandidates(ctx context.Context, limit, offset int, f_succession_plan_id *string, f_candidate_worker_id *string, f_readiness_code *string, f_talent_assessment_id *string, f_development_plan_id *string) (*Page[model.SuccessionCandidate], error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_succession_candidates", "human_resource.succession_candidates.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid succession-candidate pagination", nil)
	}
	if a.succession == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	idx, err := pickSingleFilter(f_succession_plan_id != nil, f_candidate_worker_id != nil, f_readiness_code != nil, f_talent_assessment_id != nil, f_development_plan_id != nil)
	if err != nil {
		return nil, err
	}
	var resp *succession.SuccessionCandidatesResponse
	switch idx {
	case 0:
		v, fErr := validateFilterString(f_succession_plan_id, "succession_plan_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.succession.FindSuccessionCandidateBySuccessionPlanID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 1:
		v, fErr := validateFilterString(f_candidate_worker_id, "candidate_worker_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.succession.FindSuccessionCandidateByCandidateWorkerID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 2:
		v, fErr := validateFilterString(f_readiness_code, "readiness_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.succession.FindSuccessionCandidateByReadinessCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 3:
		v, fErr := validateFilterString(f_talent_assessment_id, "talent_assessment_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.succession.FindSuccessionCandidateByTalentAssessmentID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 4:
		v, fErr := validateFilterString(f_development_plan_id, "development_plan_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.succession.FindSuccessionCandidateByDevelopmentPlanID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	default:
		resp, err = a.succession.GetAllSuccessionCandidates(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "succession-candidate")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "succession-candidate", toModelSuccessionCandidate)
}

// CreateSuccessionCandidate membuat succession-candidate baru melalui HRM.
func (a *Adapter) CreateSuccessionCandidate(ctx context.Context, input model.CreateSuccessionCandidateInput) (*model.SuccessionCandidate, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_succession_candidates", "create", "human_resource.succession_candidates.create")
	if err != nil {
		return nil, err
	}
	if a.succession == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &succession.SuccessionCandidateInput{}
	patch.SuccessionPlanId = input.SuccessionPlanID
	patch.CandidateWorkerId = input.CandidateWorkerID
	patch.ReadinessCode = input.ReadinessCode
	patch.ReadinessAssessedAt = optionalTimestamp(input.ReadinessAssessedAt)
	patch.ReadinessRationale = input.ReadinessRationale
	patch.TalentAssessmentId = input.TalentAssessmentID
	patch.DevelopmentPlanId = input.DevelopmentPlanID
	patch.RiskOfLoss = input.RiskOfLoss
	resp, err := a.succession.CreateSuccessionCandidate(rpcCtx, &succession.CreateSuccessionCandidateRequest{Input: patch})
	if err != nil {
		return nil, mapHRMWriteError(err, "succession-candidate")
	}
	return toModelSuccessionCandidate(resp.GetValue())
}

// UpdateSuccessionCandidate memperbarui succession-candidate berdasarkan FieldMask.
func (a *Adapter) UpdateSuccessionCandidate(ctx context.Context, rawID string, input model.UpdateSuccessionCandidateInput) (*model.SuccessionCandidate, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_succession_candidates", "update", "human_resource.succession_candidates.update")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "succession-candidate", "succession-candidate")
	if err != nil {
		return nil, err
	}
	if a.succession == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &succession.SuccessionCandidateInput{}
	mask := &fieldmaskpb.FieldMask{}
	if err := applyStringMask(mask, "succession_plan_id", input.SuccessionPlanID, &patch.SuccessionPlanId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "candidate_worker_id", input.CandidateWorkerID, &patch.CandidateWorkerId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "readiness_code", input.ReadinessCode, &patch.ReadinessCode, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "readiness_assessed_at", input.ReadinessAssessedAt, &patch.ReadinessAssessedAt, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "readiness_rationale", input.ReadinessRationale, &patch.ReadinessRationale, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "talent_assessment_id", input.TalentAssessmentID, &patch.TalentAssessmentId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "development_plan_id", input.DevelopmentPlanID, &patch.DevelopmentPlanId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "risk_of_loss", input.RiskOfLoss, &patch.RiskOfLoss, true); err != nil {
		return nil, err
	}
	if len(mask.Paths) == 0 {
		return nil, middleware.ErrInvalidArgument("succession-candidate update is empty", nil)
	}
	resp, err := a.succession.UpdateSuccessionCandidate(rpcCtx, &succession.UpdateSuccessionCandidateRequest{Id: id, Input: patch, UpdateMask: mask})
	if err != nil {
		return nil, mapHRMWriteError(err, "succession-candidate")
	}
	return toModelSuccessionCandidate(resp.GetValue())
}

// DeleteSuccessionCandidate menghapus succession-candidate berdasarkan ID.
func (a *Adapter) DeleteSuccessionCandidate(ctx context.Context, rawID string) (*bool, error) {
	if a == nil || a.succession == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	return deleteByUUID(ctx, a, rawID, "hr_succession_candidates", "human_resource.succession_candidates.delete", "succession-candidate", func(c context.Context, req *common.IDRequest) (*common.EmptyResponse, error) {
		return a.succession.DeleteSuccessionCandidate(c, req)
	})
}
