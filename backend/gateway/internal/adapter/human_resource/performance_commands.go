package human_resource

import (
	"context"

	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
)

// LaunchPerformanceCycle meluncurkan siklus penilaian kinerja.
func (a *Adapter) LaunchPerformanceCycle(ctx context.Context, rawID string) (*model.PerformanceCycle, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_performance_cycles", "launch", "human_resource.performance_cycles.launch")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "performance-cycle", "performance-cycle")
	if err != nil {
		return nil, err
	}
	if a.performance == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.performance.LaunchPerformanceCycle(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapHRMWriteError(err, "performance-cycle")
	}
	return toModelPerformanceCycle(resp.GetItem())
}

// ClosePerformanceCycle menutup siklus penilaian kinerja.
func (a *Adapter) ClosePerformanceCycle(ctx context.Context, rawID string) (*model.PerformanceCycle, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_performance_cycles", "close", "human_resource.performance_cycles.close")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "performance-cycle", "performance-cycle")
	if err != nil {
		return nil, err
	}
	if a.performance == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.performance.ClosePerformanceCycle(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapHRMWriteError(err, "performance-cycle")
	}
	return toModelPerformanceCycle(resp.GetItem())
}

// SubmitPerformanceGoal mengajukan sasaran kinerja untuk persetujuan.
func (a *Adapter) SubmitPerformanceGoal(ctx context.Context, rawID string) (*model.PerformanceGoal, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_performance_goals", "submit", "human_resource.performance_goals.submit")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "performance-goal", "performance-goal")
	if err != nil {
		return nil, err
	}
	if a.performance == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.performance.SubmitPerformanceGoal(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapHRMWriteError(err, "performance-goal")
	}
	return toModelPerformanceGoal(resp.GetItem())
}

// ApprovePerformanceGoal menyetujui sasaran kinerja pekerja.
func (a *Adapter) ApprovePerformanceGoal(ctx context.Context, rawID string) (*model.PerformanceGoal, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_performance_goals", "approve", "human_resource.performance_goals.approve")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "performance-goal", "performance-goal")
	if err != nil {
		return nil, err
	}
	if a.performance == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.performance.ApprovePerformanceGoal(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapHRMWriteError(err, "performance-goal")
	}
	return toModelPerformanceGoal(resp.GetItem())
}

// RejectPerformanceGoal menolak sasaran kinerja pekerja.
func (a *Adapter) RejectPerformanceGoal(ctx context.Context, rawID string) (*model.PerformanceGoal, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_performance_goals", "reject", "human_resource.performance_goals.reject")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "performance-goal", "performance-goal")
	if err != nil {
		return nil, err
	}
	if a.performance == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.performance.RejectPerformanceGoal(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapHRMWriteError(err, "performance-goal")
	}
	return toModelPerformanceGoal(resp.GetItem())
}

// SubmitPerformanceReview mengirimkan hasil penilaian kinerja.
func (a *Adapter) SubmitPerformanceReview(ctx context.Context, rawID string) (*model.PerformanceReview, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_performance_reviews", "submit", "human_resource.performance_reviews.submit")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "performance-review", "performance-review")
	if err != nil {
		return nil, err
	}
	if a.performance == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.performance.SubmitPerformanceReview(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapHRMWriteError(err, "performance-review")
	}
	return toModelPerformanceReview(resp.GetItem())
}

// FinalizePerformanceReview memfinalisasi penilaian kinerja pekerja.
func (a *Adapter) FinalizePerformanceReview(ctx context.Context, rawID string) (*model.PerformanceReview, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_performance_reviews", "finalize", "human_resource.performance_reviews.finalize")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "performance-review", "performance-review")
	if err != nil {
		return nil, err
	}
	if a.performance == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.performance.FinalizePerformanceReview(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapHRMWriteError(err, "performance-review")
	}
	return toModelPerformanceReview(resp.GetItem())
}
