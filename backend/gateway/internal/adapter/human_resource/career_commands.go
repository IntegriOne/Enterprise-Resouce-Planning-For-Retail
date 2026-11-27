package human_resource

import (
	"context"

	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
)

// CompleteDevelopmentPlan menyelesaikan rencana pengembangan pekerja.
func (a *Adapter) CompleteDevelopmentPlan(ctx context.Context, rawID string) (*model.DevelopmentPlan, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_development_plans", "complete", "human_resource.development_plans.complete")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "development-plan", "development-plan")
	if err != nil {
		return nil, err
	}
	if a.career == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.career.CompleteDevelopmentPlan(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapHRMWriteError(err, "development-plan")
	}
	return toModelDevelopmentPlan(resp.GetValue())
}

// ApproveDevelopmentPlan menyetujui rencana pengembangan pekerja.
func (a *Adapter) ApproveDevelopmentPlan(ctx context.Context, rawID string) (*model.DevelopmentPlan, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_development_plans", "approve", "human_resource.development_plans.approve")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "development-plan", "development-plan")
	if err != nil {
		return nil, err
	}
	if a.career == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.career.ApproveDevelopmentPlan(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapHRMWriteError(err, "development-plan")
	}
	return toModelDevelopmentPlan(resp.GetValue())
}

// CompleteDevelopmentAction menyelesaikan tindakan pengembangan pekerja.
func (a *Adapter) CompleteDevelopmentAction(ctx context.Context, rawID string) (*model.DevelopmentAction, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_development_actions", "complete", "human_resource.development_actions.complete")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "development-action", "development-action")
	if err != nil {
		return nil, err
	}
	if a.career == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.career.CompleteDevelopmentAction(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapHRMWriteError(err, "development-action")
	}
	return toModelDevelopmentAction(resp.GetValue())
}

// ApproveCareerPath menyetujui jalur karier yang masih berstatus draft.
func (a *Adapter) ApproveCareerPath(ctx context.Context, rawID string) (*model.CareerPath, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_career_paths", "approve", "human_resource.career_paths.approve")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "career-path", "career-path")
	if err != nil {
		return nil, err
	}
	if a.career == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.career.ApproveCareerPath(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapHRMWriteError(err, "career-path")
	}
	return toModelCareerPath(resp.GetValue())
}
