// Package inbox menerima event lintas management pada batas internal HRM.
package inbox

import (
	"context"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

// PayrollPublicationDelivery membawa identitas event dari transport Finance yang telah diautentikasi.
type PayrollPublicationDelivery struct {
	SourceManagement, EventType string
	SchemaVersion               int32
	PublicationID, PayrollRunID string
}

type PayrollPayslipPublisher interface {
	PublishPayslipsForRun(context.Context, string, string) error
}

// HandlePayrollPublication meneruskan event penerbitan ke pemroses slip.
// Transport hanya mengakui event setelah handler sukses dan harus mengirim ulang saat gagal sementara.
func HandlePayrollPublication(ctx context.Context, publisher PayrollPayslipPublisher, delivery PayrollPublicationDelivery) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if operation.RequireDependency(publisher) != nil {
		return hrerrors.ErrUnavailable
	}
	if delivery.SourceManagement != "finance_accounting" {
		return hrerrors.ErrForbidden
	}
	if delivery.EventType != "payroll.published" || delivery.SchemaVersion != 1 {
		return hrerrors.Invalid("event", "unsupported payroll publication event")
	}
	if err := entity.ValidateUUID("publication_id", delivery.PublicationID); err != nil {
		return err
	}
	if err := entity.ValidateUUID("payroll_run_id", delivery.PayrollRunID); err != nil {
		return err
	}
	return publisher.PublishPayslipsForRun(ctx, delivery.PublicationID, delivery.PayrollRunID)
}
