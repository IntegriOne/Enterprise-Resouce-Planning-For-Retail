package human_resource

import (
	"context"

	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
)

// ApproveSuccessionPlan menyetujui rencana suksesi jabatan.
func (a *Adapter) ApproveSuccessionPlan(ctx context.Context, rawID string) (*model.SuccessionPlan, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_succession_plans", "approve", "human_resource.succession_plans.approve")
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
	resp, err := a.succession.ApproveSuccessionPlan(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapHRMWriteError(err, "succession-plan")
	}
	return toModelSuccessionPlan(resp.GetValue())
}
