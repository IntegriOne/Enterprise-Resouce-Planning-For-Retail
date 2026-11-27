package human_resource

import (
	"context"
	"strings"

	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	benefits "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/benefits/v1"
	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
)

// ApproveBenefitEnrollment menyetujui pendaftaran benefit pekerja.
func (a *Adapter) ApproveBenefitEnrollment(ctx context.Context, rawID string) (*model.BenefitEnrollment, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_benefit_enrollments", "approve", "human_resource.benefit_enrollments.approve")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "benefit-enrollment", "benefit-enrollment")
	if err != nil {
		return nil, err
	}
	if a.benefits == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.benefits.ApproveBenefitEnrollment(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapHRMWriteError(err, "benefit-enrollment")
	}
	return toModelBenefitEnrollment(resp.GetItem())
}

// EndBenefitEnrollment mengakhiri kepesertaan benefit pada tanggal efektif tertentu.
func (a *Adapter) EndBenefitEnrollment(ctx context.Context, rawID, effectiveTo string) (*model.BenefitEnrollment, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_benefit_enrollments", "end", "human_resource.benefit_enrollments.end")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "benefit-enrollment", "benefit-enrollment")
	if err != nil {
		return nil, err
	}
	eff := strings.TrimSpace(effectiveTo)
	if eff == "" {
		return nil, middleware.ErrInvalidArgument("missing effective_to", nil)
	}
	if a.benefits == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.benefits.EndBenefitEnrollment(rpcCtx, &benefits.EndBenefitEnrollmentRequest{
		Id:          id,
		EffectiveTo: eff,
	})
	if err != nil {
		return nil, mapHRMWriteError(err, "benefit-enrollment")
	}
	return toModelBenefitEnrollment(resp.GetItem())
}
