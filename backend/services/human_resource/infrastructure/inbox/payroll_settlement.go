package inbox

import (
	"context"

	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

type PayrollSettlementDelivery struct {
	SourceManagement, EventType         string
	SchemaVersion                       int32
	PayrollRunID, BatchID, SettlementID string
}

type PayrollSettlementConfirmer interface {
	ConfirmPayrollSettlement(context.Context, string, string, string) error
}

// HandlePayrollSettlement meneruskan bukti settlement run dari transport Finance.
func HandlePayrollSettlement(ctx context.Context, confirmer PayrollSettlementConfirmer, delivery PayrollSettlementDelivery) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if delivery.SourceManagement != "finance_accounting" {
		return hrerrors.ErrForbidden
	}
	if delivery.EventType != "payroll.settled" || delivery.SchemaVersion != 1 {
		return hrerrors.Invalid("event", "unsupported payroll settlement event")
	}
	if operation.RequireDependency(confirmer) != nil {
		return hrerrors.ErrUnavailable
	}
	return confirmer.ConfirmPayrollSettlement(ctx, delivery.PayrollRunID, delivery.BatchID, delivery.SettlementID)
}
