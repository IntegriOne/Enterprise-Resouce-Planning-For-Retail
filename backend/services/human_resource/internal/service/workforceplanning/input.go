// Kerangka awal dibuat dengan alat bantu; kode ini dipelihara langsung per domain.
// Aturan bisnis lintas record tetap dijalankan melalui policy/integrity dan command domain.
package service

import (
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

// CreateJobFamilyInput adalah object input aplikasi, bukan entity ORM.
type CreateJobFamilyInput struct {
	FamilyCode  string
	FamilyName  string
	Description *string
	JobDomain   string
	IsActive    *bool
}

// UpdateJobFamilyInput membedakan field absen, null dan nilai eksplisit.
type UpdateJobFamilyInput struct {
	FamilyCode  operation.Field[string]
	FamilyName  operation.Field[string]
	Description operation.Field[string]
	JobDomain   operation.Field[string]
	IsActive    operation.Field[bool]
}

// CreateJobInput adalah object input aplikasi, bukan entity ORM.
type CreateJobInput struct {
	JobFamilyID               string
	JobCode                   string
	JobTitle                  string
	JobLevelCode              *string
	JobDescription            *string
	Responsibilities          entity.JSON
	QualificationRequirements entity.JSON
	ValidFrom                 time.Time
	ValidTo                   *time.Time
	IsActive                  *bool
}

// UpdateJobInput membedakan field absen, null dan nilai eksplisit.
type UpdateJobInput struct {
	JobFamilyID               operation.Field[string]
	JobCode                   operation.Field[string]
	JobTitle                  operation.Field[string]
	JobLevelCode              operation.Field[string]
	JobDescription            operation.Field[string]
	Responsibilities          operation.Field[entity.JSON]
	QualificationRequirements operation.Field[entity.JSON]
	ValidFrom                 operation.Field[time.Time]
	ValidTo                   operation.Field[time.Time]
	IsActive                  operation.Field[bool]
}

// CreatePositionInput adalah object input aplikasi, bukan entity ORM.
type CreatePositionInput struct {
	PositionCode                string
	PositionTitle               string
	JobID                       string
	OrganizationUnitReferenceID string
	LegalEntityReferenceID      string
	ReportsToPositionID         *string
	PositionStatus              *string
	HeadcountAuthorized         *entity.Decimal
	WorkLocationReferenceID     *string
	CostCenterReferenceID       *string
	ValidFrom                   time.Time
	ValidTo                     *time.Time
}

// UpdatePositionInput membedakan field absen, null dan nilai eksplisit.
type UpdatePositionInput struct {
	PositionCode                operation.Field[string]
	PositionTitle               operation.Field[string]
	JobID                       operation.Field[string]
	OrganizationUnitReferenceID operation.Field[string]
	LegalEntityReferenceID      operation.Field[string]
	ReportsToPositionID         operation.Field[string]
	PositionStatus              operation.Field[string]
	HeadcountAuthorized         operation.Field[entity.Decimal]
	WorkLocationReferenceID     operation.Field[string]
	CostCenterReferenceID       operation.Field[string]
	ValidFrom                   operation.Field[time.Time]
	ValidTo                     operation.Field[time.Time]
}

// CreateWorkforcePlanInput adalah object input aplikasi, bukan entity ORM.
type CreateWorkforcePlanInput struct {
	PlanCode                 string
	PlanVersion              int32
	LegalEntityReferenceID   string
	PeriodStart              time.Time
	PeriodEnd                time.Time
	ScenarioCode             *string
	Status                   *string
	FinanceBudgetReferenceID *string
	ApprovedByPrincipalID    *string
	ApprovedAt               *time.Time
}

// UpdateWorkforcePlanInput membedakan field absen, null dan nilai eksplisit.
type UpdateWorkforcePlanInput struct {
	PlanCode                 operation.Field[string]
	PlanVersion              operation.Field[int32]
	LegalEntityReferenceID   operation.Field[string]
	PeriodStart              operation.Field[time.Time]
	PeriodEnd                operation.Field[time.Time]
	ScenarioCode             operation.Field[string]
	Status                   operation.Field[string]
	FinanceBudgetReferenceID operation.Field[string]
	ApprovedByPrincipalID    operation.Field[string]
	ApprovedAt               operation.Field[time.Time]
}

// CreateWorkforcePlanLineInput adalah object input aplikasi, bukan entity ORM.
type CreateWorkforcePlanLineInput struct {
	WorkforcePlanID             string
	LineNumber                  int32
	OrganizationUnitReferenceID string
	JobID                       string
	WorkLocationReferenceID     *string
	WorkerCategory              entity.WorkerCategory
	CurrentFTE                  entity.Decimal
	RequiredFTE                 entity.Decimal
	EstimatedCost               *entity.Decimal
	CurrencyCode                *string
	NeededFrom                  time.Time
	NeededUntil                 *time.Time
	ActionCode                  string
	Rationale                   string
}

// UpdateWorkforcePlanLineInput membedakan field absen, null dan nilai eksplisit.
type UpdateWorkforcePlanLineInput struct {
	WorkforcePlanID             operation.Field[string]
	LineNumber                  operation.Field[int32]
	OrganizationUnitReferenceID operation.Field[string]
	JobID                       operation.Field[string]
	WorkLocationReferenceID     operation.Field[string]
	WorkerCategory              operation.Field[entity.WorkerCategory]
	CurrentFTE                  operation.Field[entity.Decimal]
	RequiredFTE                 operation.Field[entity.Decimal]
	EstimatedCost               operation.Field[entity.Decimal]
	CurrencyCode                operation.Field[string]
	NeededFrom                  operation.Field[time.Time]
	NeededUntil                 operation.Field[time.Time]
	ActionCode                  operation.Field[string]
	Rationale                   operation.Field[string]
}
