// Kerangka awal dibuat dengan alat bantu; kode ini dipelihara langsung per domain.
// Aturan bisnis lintas record tetap dijalankan melalui policy/integrity dan command domain.
package service

import (
	"time"

	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

// CreateOffboardingPlanInput adalah object input aplikasi, bukan entity ORM.
type CreateOffboardingPlanInput struct {
	WorkerID                                string
	PlanNumber                              string
	ReasonCode                              string
	NoticeDate                              time.Time
	LastWorkingDate                         time.Time
	TerminationEffectiveDate                time.Time
	Status                                  *string
	ExitInterviewDocumentVersionReferenceID *string
	FinalPayrollRunID                       *string
	FinanceClearanceReferenceID             *string
	ApprovedByPrincipalID                   *string
	ApprovedAt                              *time.Time
	CompletedAt                             *time.Time
}

// UpdateOffboardingPlanInput membedakan field absen, null dan nilai eksplisit.
type UpdateOffboardingPlanInput struct {
	WorkerID                                operation.Field[string]
	PlanNumber                              operation.Field[string]
	ReasonCode                              operation.Field[string]
	NoticeDate                              operation.Field[time.Time]
	LastWorkingDate                         operation.Field[time.Time]
	TerminationEffectiveDate                operation.Field[time.Time]
	Status                                  operation.Field[string]
	ExitInterviewDocumentVersionReferenceID operation.Field[string]
	FinalPayrollRunID                       operation.Field[string]
	FinanceClearanceReferenceID             operation.Field[string]
	ApprovedByPrincipalID                   operation.Field[string]
	ApprovedAt                              operation.Field[time.Time]
	CompletedAt                             operation.Field[time.Time]
}

// CreateOffboardingTaskInput adalah object input aplikasi, bukan entity ORM.
type CreateOffboardingTaskInput struct {
	OffboardingPlanID            string
	TaskCode                     string
	TaskName                     string
	OwnerManagement              string
	AssignedPrincipalID          *string
	Status                       *string
	IsMandatory                  *bool
	DueAt                        *time.Time
	CompletedAt                  *time.Time
	OwnerConfirmationReferenceID *string
	WaiverReason                 *string
	WaivedByPrincipalID          *string
}

// UpdateOffboardingTaskInput membedakan field absen, null dan nilai eksplisit.
type UpdateOffboardingTaskInput struct {
	OffboardingPlanID            operation.Field[string]
	TaskCode                     operation.Field[string]
	TaskName                     operation.Field[string]
	OwnerManagement              operation.Field[string]
	AssignedPrincipalID          operation.Field[string]
	Status                       operation.Field[string]
	IsMandatory                  operation.Field[bool]
	DueAt                        operation.Field[time.Time]
	CompletedAt                  operation.Field[time.Time]
	OwnerConfirmationReferenceID operation.Field[string]
	WaiverReason                 operation.Field[string]
	WaivedByPrincipalID          operation.Field[string]
}
