package human_resource

import (
	"context"

	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
)

// StartLearningEnrollment memulai progres pembelajaran untuk enrollment terkait.
func (a *Adapter) StartLearningEnrollment(ctx context.Context, rawID string) (*model.LearningEnrollment, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_learning_enrollments", "start", "human_resource.learning_enrollments.start")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "learning-enrollment", "learning-enrollment")
	if err != nil {
		return nil, err
	}
	if a.learning == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.learning.StartLearningEnrollment(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapHRMWriteError(err, "learning-enrollment")
	}
	return toModelLearningEnrollment(resp.GetItem())
}

// CompleteLearningEnrollment menandai pembelajaran pekerja telah selesai.
func (a *Adapter) CompleteLearningEnrollment(ctx context.Context, rawID string) (*model.LearningEnrollment, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_learning_enrollments", "complete", "human_resource.learning_enrollments.complete")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "learning-enrollment", "learning-enrollment")
	if err != nil {
		return nil, err
	}
	if a.learning == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.learning.CompleteLearningEnrollment(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapHRMWriteError(err, "learning-enrollment")
	}
	return toModelLearningEnrollment(resp.GetItem())
}

// WithdrawLearningEnrollment menarik kembali kepesertaan pembelajaran.
func (a *Adapter) WithdrawLearningEnrollment(ctx context.Context, rawID string) (*model.LearningEnrollment, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_learning_enrollments", "withdraw", "human_resource.learning_enrollments.withdraw")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "learning-enrollment", "learning-enrollment")
	if err != nil {
		return nil, err
	}
	if a.learning == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.learning.WithdrawLearningEnrollment(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapHRMWriteError(err, "learning-enrollment")
	}
	return toModelLearningEnrollment(resp.GetItem())
}
