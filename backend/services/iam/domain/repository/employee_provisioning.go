package repository

import (
	"context"
	"time"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
)

// EmployeeProvisioningRepository menyimpan penerimaan dan transisi pemrosesan pengajuan.
type EmployeeProvisioningRepository interface {
	InsertReceipt(context.Context, entity.EmployeeProvisioningRequest) (*entity.EmployeeProvisioningRequest, error)
	SelectOpen(ctx context.Context, limit, offset int) ([]entity.EmployeeProvisioningRequest, int64, error)
	SelectByHRMRequest(ctx context.Context, hrmRequestID, legalEntityID string) (*entity.EmployeeProvisioningRequest, error)
	SelectByID(context.Context, string) (*entity.EmployeeProvisioningRequest, error)
	SelectByIDForUpdate(context.Context, string) (*entity.EmployeeProvisioningRequest, error)
	HasUnlinkedAccount(context.Context, string) (bool, error)
	UpdateLinked(ctx context.Context, id, operatorID, accountID string, at time.Time) error
	UpdateAccount(ctx context.Context, id, operatorID, accountID string) error
	UpdateClaim(ctx context.Context, id, operatorID string, at time.Time) (*entity.EmployeeProvisioningRequest, error)
}
