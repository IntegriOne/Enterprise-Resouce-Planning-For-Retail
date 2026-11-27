package human_resource

import (
	"context"

	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	onboarding "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/onboarding/v1"
)

// StartOnboardingPlan memulai rencana onboarding pekerja.
func (a *Adapter) StartOnboardingPlan(ctx context.Context, rawID string) (*model.OnboardingPlan, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_onboarding_plans", "start", "human_resource.onboarding_plans.start")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "onboarding-plan", "onboarding-plan")
	if err != nil {
		return nil, err
	}
	if a.onboarding == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.onboarding.StartOnboardingPlan(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapHRMWriteError(err, "onboarding-plan")
	}
	return toModelOnboardingPlan(resp.GetItem())
}

// CompleteOnboardingPlan menyelesaikan rencana onboarding pekerja.
func (a *Adapter) CompleteOnboardingPlan(ctx context.Context, rawID string) (*model.OnboardingPlan, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_onboarding_plans", "complete", "human_resource.onboarding_plans.complete")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "onboarding-plan", "onboarding-plan")
	if err != nil {
		return nil, err
	}
	if a.onboarding == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.onboarding.CompleteOnboardingPlan(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapHRMWriteError(err, "onboarding-plan")
	}
	return toModelOnboardingPlan(resp.GetItem())
}

// CompleteOnboardingTask menyelesaikan satu tugas onboarding beserta bukti dokumennya.
func (a *Adapter) CompleteOnboardingTask(ctx context.Context, rawID, evidenceDocVerRefID string) (*model.OnboardingTask, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_onboarding_tasks", "complete", "human_resource.onboarding_tasks.complete")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "onboarding-task", "onboarding-task")
	if err != nil {
		return nil, err
	}
	docVerID, err := normalizeRequiredUUID(evidenceDocVerRefID, "document-version", "evidence_document_version_reference_id")
	if err != nil {
		return nil, err
	}
	if a.onboarding == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.onboarding.CompleteOnboardingTask(rpcCtx, &onboarding.CompleteOnboardingTaskRequest{
		Id:                                 id,
		EvidenceDocumentVersionReferenceId: docVerID,
	})
	if err != nil {
		return nil, mapHRMWriteError(err, "onboarding-task")
	}
	return toModelOnboardingTask(resp.GetItem())
}
