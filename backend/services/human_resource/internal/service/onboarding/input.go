// Kerangka awal dibuat dengan alat bantu; kode ini dipelihara langsung per domain.
// Aturan bisnis lintas record tetap dijalankan melalui policy/integrity dan command domain.
package service

import (
	"time"

	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

// CreateOnboardingPlanInput adalah object input aplikasi, bukan entity ORM.
type CreateOnboardingPlanInput struct {
	WorkerID             string
	PlanCode             string
	PlanVersion          *int32
	StartDate            time.Time
	TargetCompletionDate *time.Time
	Status               *string
	TemplateReferenceID  *string
}

// UpdateOnboardingPlanInput membedakan field absen, null dan nilai eksplisit.
type UpdateOnboardingPlanInput struct {
	WorkerID             operation.Field[string]
	PlanCode             operation.Field[string]
	PlanVersion          operation.Field[int32]
	StartDate            operation.Field[time.Time]
	TargetCompletionDate operation.Field[time.Time]
	Status               operation.Field[string]
	TemplateReferenceID  operation.Field[string]
}

// CreateOnboardingTaskInput adalah object input aplikasi, bukan entity ORM.
type CreateOnboardingTaskInput struct {
	OnboardingPlanID              string
	TaskCode                      string
	TaskName                      string
	TaskOwnerType                 string
	AssignedPrincipalID           *string
	DueAt                         *time.Time
	CompletedAt                   *time.Time
	CompletionEvidenceReferenceID *string
	IsMandatory                   *bool
}

// UpdateOnboardingTaskInput membedakan field absen, null dan nilai eksplisit.
type UpdateOnboardingTaskInput struct {
	OnboardingPlanID              operation.Field[string]
	TaskCode                      operation.Field[string]
	TaskName                      operation.Field[string]
	TaskOwnerType                 operation.Field[string]
	AssignedPrincipalID           operation.Field[string]
	DueAt                         operation.Field[time.Time]
	CompletedAt                   operation.Field[time.Time]
	CompletionEvidenceReferenceID operation.Field[string]
	IsMandatory                   operation.Field[bool]
}
