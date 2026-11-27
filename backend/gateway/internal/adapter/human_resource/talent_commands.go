package human_resource

import (
	"context"
	"encoding/base64"
	"strings"

	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	talent "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/talent/v1"
)

// LaunchTalentReviewCycle meluncurkan siklus review talenta.
func (a *Adapter) LaunchTalentReviewCycle(ctx context.Context, rawID string) (*model.TalentReviewCycle, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_talent_review_cycles", "launch", "human_resource.talent_review_cycles.launch")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "talent-review-cycle", "talent-review-cycle")
	if err != nil {
		return nil, err
	}
	if a.talent == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.talent.LaunchTalentReviewCycle(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapHRMWriteError(err, "talent-review-cycle")
	}
	return toModelTalentReviewCycle(resp.GetValue())
}

// StartTalentReviewCycleCalibration memulai fase kalibrasi pada siklus review talenta.
func (a *Adapter) StartTalentReviewCycleCalibration(ctx context.Context, rawID string) (*model.TalentReviewCycle, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_talent_review_cycles", "start_calibration", "human_resource.talent_review_cycles.start_calibration")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "talent-review-cycle", "talent-review-cycle")
	if err != nil {
		return nil, err
	}
	if a.talent == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.talent.StartTalentReviewCycleCalibration(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapHRMWriteError(err, "talent-review-cycle")
	}
	return toModelTalentReviewCycle(resp.GetValue())
}

// CloseTalentReviewCycle menutup siklus review talenta.
func (a *Adapter) CloseTalentReviewCycle(ctx context.Context, rawID string) (*model.TalentReviewCycle, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_talent_review_cycles", "close", "human_resource.talent_review_cycles.close")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "talent-review-cycle", "talent-review-cycle")
	if err != nil {
		return nil, err
	}
	if a.talent == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.talent.CloseTalentReviewCycle(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapHRMWriteError(err, "talent-review-cycle")
	}
	return toModelTalentReviewCycle(resp.GetValue())
}

// SubmitTalentAssessment mengirimkan asesmen talenta pekerja.
func (a *Adapter) SubmitTalentAssessment(ctx context.Context, rawID string) (*model.TalentAssessment, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_talent_assessments", "submit", "human_resource.talent_assessments.submit")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "talent-assessment", "talent-assessment")
	if err != nil {
		return nil, err
	}
	if a.talent == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.talent.SubmitTalentAssessment(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapHRMWriteError(err, "talent-assessment")
	}
	return toModelTalentAssessment(resp.GetValue())
}

// WithdrawTalentAssessment menarik kembali asesmen talenta pekerja.
func (a *Adapter) WithdrawTalentAssessment(ctx context.Context, rawID string) (*model.TalentAssessment, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_talent_assessments", "withdraw", "human_resource.talent_assessments.withdraw")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "talent-assessment", "talent-assessment")
	if err != nil {
		return nil, err
	}
	if a.talent == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.talent.WithdrawTalentAssessment(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapHRMWriteError(err, "talent-assessment")
	}
	return toModelTalentAssessment(resp.GetValue())
}

// FinalizeTalentAssessment memfinalisasi asesmen talenta pekerja.
func (a *Adapter) FinalizeTalentAssessment(ctx context.Context, rawID string) (*model.TalentAssessment, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_talent_assessments", "finalize", "human_resource.talent_assessments.finalize")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "talent-assessment", "talent-assessment")
	if err != nil {
		return nil, err
	}
	if a.talent == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.talent.FinalizeTalentAssessment(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapHRMWriteError(err, "talent-assessment")
	}
	return toModelTalentAssessment(resp.GetValue())
}

// OpenTalentCalibrationSession membuka sesi kalibrasi talenta.
func (a *Adapter) OpenTalentCalibrationSession(ctx context.Context, rawID string) (*model.TalentCalibrationSession, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_talent_calibration_sessions", "open", "human_resource.talent_calibration_sessions.open")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "talent-calibration-session", "talent-calibration-session")
	if err != nil {
		return nil, err
	}
	if a.talent == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.talent.OpenTalentCalibrationSession(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapHRMWriteError(err, "talent-calibration-session")
	}
	return toModelTalentCalibrationSession(resp.GetValue())
}

// RecordTalentCalibrationDecision mencatat keputusan hasil kalibrasi talenta.
func (a *Adapter) RecordTalentCalibrationDecision(ctx context.Context, input model.RecordTalentCalibrationDecisionInput) (*model.TalentCalibrationDecision, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_talent_calibration_sessions", "approve", "human_resource.talent_calibration_sessions.approve")
	if err != nil {
		return nil, err
	}
	sessionID, err := normalizeRequiredUUID(input.SessionID, "talent-calibration-session", "session_id")
	if err != nil {
		return nil, err
	}
	assessmentID, err := normalizeRequiredUUID(input.AssessmentID, "talent-assessment", "assessment_id")
	if err != nil {
		return nil, err
	}
	dimensionID, err := normalizeRequiredUUID(input.DimensionID, "talent-dimension", "dimension_id")
	if err != nil {
		return nil, err
	}
	docVerID, err := normalizeRequiredUUID(input.RationaleDocumentVersionReferenceID, "document-version", "rationale_document_version_reference_id")
	if err != nil {
		return nil, err
	}
	digest, err := base64.StdEncoding.DecodeString(strings.TrimSpace(input.RationaleDigest))
	if err != nil || len(digest) == 0 {
		return nil, middleware.ErrInvalidArgument("invalid base64 for rationale_digest", nil)
	}
	if a.talent == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.talent.RecordTalentCalibrationDecision(rpcCtx, &talent.RecordTalentCalibrationDecisionRequest{
		SessionId:                           sessionID,
		AssessmentId:                        assessmentID,
		DimensionId:                         dimensionID,
		CalibratedScore:                     int32(input.CalibratedScore),
		RationaleDocumentVersionReferenceId: docVerID,
		RationaleDigest:                     digest,
	})
	if err != nil {
		return nil, mapHRMWriteError(err, "talent-calibration-decision")
	}
	return toModelTalentCalibrationDecision(resp.GetValue())
}

// ApproveTalentPoolMembership menyetujui nominasi keanggotaan talent pool.
func (a *Adapter) ApproveTalentPoolMembership(ctx context.Context, rawID string) (*model.TalentPoolMembership, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_talent_pool_memberships", "approve", "human_resource.talent_pool_memberships.approve")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "talent-pool-membership", "talent-pool-membership")
	if err != nil {
		return nil, err
	}
	if a.talent == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.talent.ApproveTalentPoolMembership(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapHRMWriteError(err, "talent-pool-membership")
	}
	return toModelTalentPoolMembership(resp.GetValue())
}

// RejectTalentPoolMembership menolak nominasi keanggotaan talent pool.
func (a *Adapter) RejectTalentPoolMembership(ctx context.Context, rawID, reasonCode string) (*model.TalentPoolMembership, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_talent_pool_memberships", "reject", "human_resource.talent_pool_memberships.reject")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "talent-pool-membership", "talent-pool-membership")
	if err != nil {
		return nil, err
	}
	reason := strings.TrimSpace(reasonCode)
	if reason == "" {
		return nil, middleware.ErrInvalidArgument("missing reason_code", nil)
	}
	if a.talent == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.talent.RejectTalentPoolMembership(rpcCtx, &talent.TalentReasonActionRequest{
		Id:         id,
		ReasonCode: reason,
	})
	if err != nil {
		return nil, mapHRMWriteError(err, "talent-pool-membership")
	}
	return toModelTalentPoolMembership(resp.GetValue())
}

// ExitTalentPoolMembership mengakhiri keanggotaan pekerja di talent pool.
func (a *Adapter) ExitTalentPoolMembership(ctx context.Context, rawID, reasonCode string) (*model.TalentPoolMembership, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_talent_pool_memberships", "exit", "human_resource.talent_pool_memberships.exit")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "talent-pool-membership", "talent-pool-membership")
	if err != nil {
		return nil, err
	}
	reason := strings.TrimSpace(reasonCode)
	if reason == "" {
		return nil, middleware.ErrInvalidArgument("missing reason_code", nil)
	}
	if a.talent == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.talent.ExitTalentPoolMembership(rpcCtx, &talent.TalentReasonActionRequest{
		Id:         id,
		ReasonCode: reason,
	})
	if err != nil {
		return nil, mapHRMWriteError(err, "talent-pool-membership")
	}
	return toModelTalentPoolMembership(resp.GetValue())
}
