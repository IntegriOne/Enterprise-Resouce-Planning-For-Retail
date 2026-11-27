package entity

import "time"

// EmployeeProvisioningRequest mencatat penerimaan dan hasil pengajuan akun dari HRM di IAM.
type EmployeeProvisioningRequest struct {
	ID                     string
	HRMRequestID           string
	WorkerID               string
	WorkerNumber           string
	LegalEntityID          string
	CatalogRoleID          string
	RequestedByPrincipalID string
	RequestedAt            time.Time
	ReceivedAt             time.Time
	Status                 string
	ProcessedByUserID      *string
	ClaimedAt              *time.Time
	AccountUserID          *string
	ResolvedAt             *time.Time
}

const (
	EmployeeProvisioningPending    = "pending"
	EmployeeProvisioningProcessing = "processing"
	EmployeeProvisioningLinked     = "linked"
	EmployeeProvisioningRejected   = "rejected"
)
