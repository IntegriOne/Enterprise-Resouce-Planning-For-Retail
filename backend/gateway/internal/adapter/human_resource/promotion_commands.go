package human_resource

import (
	"context"
	"encoding/base64"
	"strings"

	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	promotion "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/promotion/v1"
)

func buildDecidePromotionReq(rawID string, input model.DecidePromotionRecommendationInput) (*promotion.DecidePromotionRecommendationRequest, error) {
	id, err := normalizeRequiredUUID(rawID, "promotion-recommendation", "promotion-recommendation")
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
	idem := strings.TrimSpace(input.IdempotencyKey)
	if idem == "" {
		return nil, middleware.ErrInvalidArgument("missing idempotency_key", nil)
	}
	return &promotion.DecidePromotionRecommendationRequest{
		RecommendationId:                    id,
		RationaleDocumentVersionReferenceId: docVerID,
		RationaleDigest:                     digest,
		IdempotencyKey:                      idem,
	}, nil
}

// SubmitPromotionRecommendation mengajukan rekomendasi promosi draft untuk diproses.
func (a *Adapter) SubmitPromotionRecommendation(ctx context.Context, rawID string) (*model.PromotionRecommendation, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_promotion_recommendations", "submit", "human_resource.promotion_recommendations.submit")
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
	resp, err := a.promotion.SubmitPromotionRecommendation(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapHRMWriteError(err, "promotion-recommendation")
	}
	return toModelPromotionRecommendation(resp.GetValue())
}

// ApprovePromotionRecommendation menyetujui rekomendasi promosi pada tahap keputusan.
func (a *Adapter) ApprovePromotionRecommendation(ctx context.Context, rawID string, input model.DecidePromotionRecommendationInput) (*model.PromotionRecommendation, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_promotion_recommendations", "approve", "human_resource.promotion_decisions.approve")
	if err != nil {
		return nil, err
	}
	req, err := buildDecidePromotionReq(rawID, input)
	if err != nil {
		return nil, err
	}
	if a.promotion == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.promotion.ApprovePromotionRecommendation(rpcCtx, req)
	if err != nil {
		return nil, mapHRMWriteError(err, "promotion-recommendation")
	}
	return toModelPromotionRecommendation(resp.GetValue())
}

// RejectPromotionRecommendation menolak rekomendasi promosi pada tahap keputusan.
func (a *Adapter) RejectPromotionRecommendation(ctx context.Context, rawID string, input model.DecidePromotionRecommendationInput) (*model.PromotionRecommendation, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_promotion_recommendations", "reject", "human_resource.promotion_decisions.reject")
	if err != nil {
		return nil, err
	}
	req, err := buildDecidePromotionReq(rawID, input)
	if err != nil {
		return nil, err
	}
	if a.promotion == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.promotion.RejectPromotionRecommendation(rpcCtx, req)
	if err != nil {
		return nil, mapHRMWriteError(err, "promotion-recommendation")
	}
	return toModelPromotionRecommendation(resp.GetValue())
}

// ReturnPromotionRecommendation mengembalikan rekomendasi promosi untuk revisi.
func (a *Adapter) ReturnPromotionRecommendation(ctx context.Context, rawID string, input model.DecidePromotionRecommendationInput) (*model.PromotionRecommendation, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_promotion_recommendations", "return", "human_resource.promotion_decisions.return")
	if err != nil {
		return nil, err
	}
	req, err := buildDecidePromotionReq(rawID, input)
	if err != nil {
		return nil, err
	}
	if a.promotion == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.promotion.ReturnPromotionRecommendation(rpcCtx, req)
	if err != nil {
		return nil, mapHRMWriteError(err, "promotion-recommendation")
	}
	return toModelPromotionRecommendation(resp.GetValue())
}

// ApplyPromotionRecommendation menerapkan promosi yang sudah disetujui ke penugasan pekerja.
func (a *Adapter) ApplyPromotionRecommendation(ctx context.Context, rawID string) (*model.PromotionRecommendation, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_promotion_recommendations", "apply", "human_resource.promotion_decisions.apply")
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
	resp, err := a.promotion.ApplyPromotionRecommendation(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapHRMWriteError(err, "promotion-recommendation")
	}
	return toModelPromotionRecommendation(resp.GetValue())
}
