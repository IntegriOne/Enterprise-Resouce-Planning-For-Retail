package human_resource

import (
	"context"

	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
)

// ApproveCompensationRecord menyetujui catatan kompensasi pekerja.
func (a *Adapter) ApproveCompensationRecord(ctx context.Context, rawID string) (*model.CompensationRecord, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_compensation_records", "approve", "human_resource.compensation_records.approve")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "compensation-record", "compensation-record")
	if err != nil {
		return nil, err
	}
	if a.compensation == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.compensation.ApproveCompensationRecord(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapHRMWriteError(err, "compensation-record")
	}
	return toModelCompensationRecord(resp.GetItem())
}
