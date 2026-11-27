package inbox

import (
	"context"

	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

type PayrollPostingDelivery struct {
	SourceManagement, EventType string
	SchemaVersion               int32
	PayrollRunID, JournalID     string
}

type PayrollPostingConfirmer interface {
	ConfirmPayrollPosting(context.Context, string, string) error
}

// HandlePayrollPosting menerima konfirmasi jurnal dari identitas transport Finance.
func HandlePayrollPosting(ctx context.Context, confirmer PayrollPostingConfirmer, delivery PayrollPostingDelivery) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if delivery.SourceManagement != "finance_accounting" {
		return hrerrors.ErrForbidden
	}
	if delivery.EventType != "payroll.posted" || delivery.SchemaVersion != 1 {
		return hrerrors.Invalid("event", "unsupported payroll posting event")
	}
	if operation.RequireDependency(confirmer) != nil {
		return hrerrors.ErrUnavailable
	}
	return confirmer.ConfirmPayrollPosting(ctx, delivery.PayrollRunID, delivery.JournalID)
}
