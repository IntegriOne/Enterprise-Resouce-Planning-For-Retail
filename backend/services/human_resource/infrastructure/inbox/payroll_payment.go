package inbox

import (
	"context"

	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

type PayrollPaymentDelivery struct {
	SourceManagement, EventType       string
	SchemaVersion                     int32
	PayrollRunID, ResultID, PaymentID string
}

type PayrollPaymentConfirmer interface {
	ConfirmPayrollPayment(context.Context, string, string, string) error
}

// HandlePayrollPayment meneruskan konfirmasi hasil individual dari transport Finance tepercaya.
func HandlePayrollPayment(ctx context.Context, confirmer PayrollPaymentConfirmer, delivery PayrollPaymentDelivery) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if delivery.SourceManagement != "finance_accounting" {
		return hrerrors.ErrForbidden
	}
	if delivery.EventType != "payroll.result_paid" || delivery.SchemaVersion != 1 {
		return hrerrors.Invalid("event", "unsupported payroll payment event")
	}
	if operation.RequireDependency(confirmer) != nil {
		return hrerrors.ErrUnavailable
	}
	return confirmer.ConfirmPayrollPayment(ctx, delivery.PayrollRunID, delivery.ResultID, delivery.PaymentID)
}
