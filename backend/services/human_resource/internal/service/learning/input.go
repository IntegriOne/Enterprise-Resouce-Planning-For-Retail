// Kerangka awal dibuat dengan alat bantu; kode ini dipelihara langsung per domain.
// Aturan bisnis lintas record tetap dijalankan melalui policy/integrity dan command domain.
package service

import (
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

// CreateLearningCourseInput adalah object input aplikasi, bukan entity ORM.
type CreateLearningCourseInput struct {
	CourseCode               string
	CourseVersion            string
	CourseName               string
	CourseType               string
	ProviderPartyReferenceID *string
	DeliveryMode             string
	DurationHours            *entity.Decimal
	CredentialAwarded        *bool
	ExpiryMonths             *int32
	ValidFrom                *time.Time
	ValidTo                  *time.Time
	IsActive                 *bool
}

// UpdateLearningCourseInput membedakan field absen, null dan nilai eksplisit.
type UpdateLearningCourseInput struct {
	CourseCode               operation.Field[string]
	CourseVersion            operation.Field[string]
	CourseName               operation.Field[string]
	CourseType               operation.Field[string]
	ProviderPartyReferenceID operation.Field[string]
	DeliveryMode             operation.Field[string]
	DurationHours            operation.Field[entity.Decimal]
	CredentialAwarded        operation.Field[bool]
	ExpiryMonths             operation.Field[int32]
	ValidFrom                operation.Field[time.Time]
	ValidTo                  operation.Field[time.Time]
	IsActive                 operation.Field[bool]
}

// CreateLearningEnrollmentInput adalah object input aplikasi, bukan entity ORM.
type CreateLearningEnrollmentInput struct {
	WorkerID              string
	CourseID              string
	EnrollmentStatus      *entity.LearningStatus
	AssignedByPrincipalID *string
	DueAt                 *time.Time
	StartedAt             *time.Time
	CompletedAt           *time.Time
	Score                 *entity.Decimal
	CredentialReferenceID *string
	FundingReferenceID    *string
	IdempotencyKey        string
}

// UpdateLearningEnrollmentInput membedakan field absen, null dan nilai eksplisit.
type UpdateLearningEnrollmentInput struct {
	WorkerID              operation.Field[string]
	CourseID              operation.Field[string]
	EnrollmentStatus      operation.Field[entity.LearningStatus]
	AssignedByPrincipalID operation.Field[string]
	DueAt                 operation.Field[time.Time]
	StartedAt             operation.Field[time.Time]
	CompletedAt           operation.Field[time.Time]
	Score                 operation.Field[entity.Decimal]
	CredentialReferenceID operation.Field[string]
	FundingReferenceID    operation.Field[string]
	IdempotencyKey        operation.Field[string]
}

// CreateLearningCourseCompetencyInput adalah object input aplikasi, bukan entity ORM.
type CreateLearningCourseCompetencyInput struct {
	CourseID     string
	CompetencyID string
	TargetLevel  int16
}

// UpdateLearningCourseCompetencyInput membedakan field absen, null dan nilai eksplisit.
type UpdateLearningCourseCompetencyInput struct {
	TargetLevel operation.Field[int16]
}
