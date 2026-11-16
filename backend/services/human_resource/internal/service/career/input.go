// Kerangka awal dibuat dengan alat bantu; kode ini dipelihara langsung per domain.
// Aturan bisnis lintas record tetap dijalankan melalui policy/integrity dan command domain.
package service

import (
	"time"

	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

// CreateDevelopmentPlanInput adalah object input aplikasi, bukan entity ORM.
type CreateDevelopmentPlanInput struct {
	WorkerID              string
	PlanCode              string
	PlanVersion           int32
	TargetJobID           *string
	StartDate             time.Time
	TargetDate            time.Time
	Status                *string
	MentorWorkerID        *string
	ApprovedByPrincipalID *string
	ApprovedAt            *time.Time
}

// UpdateDevelopmentPlanInput membedakan field absen, null dan nilai eksplisit.
type UpdateDevelopmentPlanInput struct {
	WorkerID              operation.Field[string]
	PlanCode              operation.Field[string]
	PlanVersion           operation.Field[int32]
	TargetJobID           operation.Field[string]
	StartDate             operation.Field[time.Time]
	TargetDate            operation.Field[time.Time]
	Status                operation.Field[string]
	MentorWorkerID        operation.Field[string]
	ApprovedByPrincipalID operation.Field[string]
	ApprovedAt            operation.Field[time.Time]
}

// CreateDevelopmentActionInput adalah object input aplikasi, bukan entity ORM.
type CreateDevelopmentActionInput struct {
	DevelopmentPlanID                  string
	ActionNumber                       int32
	CompetencyID                       *string
	CourseID                           *string
	ActionType                         string
	Objective                          string
	TargetLevel                        *int16
	DueDate                            time.Time
	CompletedAt                        *time.Time
	EvidenceDocumentVersionReferenceID *string
}

// UpdateDevelopmentActionInput membedakan field absen, null dan nilai eksplisit.
type UpdateDevelopmentActionInput struct {
	DevelopmentPlanID                  operation.Field[string]
	ActionNumber                       operation.Field[int32]
	CompetencyID                       operation.Field[string]
	CourseID                           operation.Field[string]
	ActionType                         operation.Field[string]
	Objective                          operation.Field[string]
	TargetLevel                        operation.Field[int16]
	DueDate                            operation.Field[time.Time]
	CompletedAt                        operation.Field[time.Time]
	EvidenceDocumentVersionReferenceID operation.Field[string]
}

// CreateCareerPathInput adalah object input aplikasi, bukan entity ORM.
type CreateCareerPathInput struct {
	PathCode              string
	PathVersion           int32
	PathName              string
	JobFamilyID           string
	ValidFrom             time.Time
	ValidTo               *time.Time
	ApprovedByPrincipalID string
	ApprovedAt            time.Time
}

// UpdateCareerPathInput membedakan field absen, null dan nilai eksplisit.
type UpdateCareerPathInput struct {
	PathCode              operation.Field[string]
	PathVersion           operation.Field[int32]
	PathName              operation.Field[string]
	JobFamilyID           operation.Field[string]
	ValidFrom             operation.Field[time.Time]
	ValidTo               operation.Field[time.Time]
	ApprovedByPrincipalID operation.Field[string]
	ApprovedAt            operation.Field[time.Time]
}

// CreateCareerPathTransitionInput adalah object input aplikasi, bukan entity ORM.
type CreateCareerPathTransitionInput struct {
	CareerPathID                           string
	FromJobID                              string
	ToJobID                                string
	MovementType                           string
	MinimumTenureMonths                    *int32
	RequirementsDocumentVersionReferenceID string
	RequirementsDigest                     []byte
}

// UpdateCareerPathTransitionInput membedakan field absen, null dan nilai eksplisit.
type UpdateCareerPathTransitionInput struct {
	CareerPathID                           operation.Field[string]
	FromJobID                              operation.Field[string]
	ToJobID                                operation.Field[string]
	MovementType                           operation.Field[string]
	MinimumTenureMonths                    operation.Field[int32]
	RequirementsDocumentVersionReferenceID operation.Field[string]
	RequirementsDigest                     operation.Field[[]byte]
}

// CreateWorkerCareerAspirationInput adalah object input aplikasi, bukan entity ORM.
type CreateWorkerCareerAspirationInput struct {
	WorkerID                     string
	TargetJobID                  string
	CareerPathTransitionID       *string
	PreferenceRank               int32
	TargetDate                   *time.Time
	MobilityPreference           *string
	PreferredLocationReferenceID *string
	WithdrawnAt                  *time.Time
}

// UpdateWorkerCareerAspirationInput membedakan field absen, null dan nilai eksplisit.
type UpdateWorkerCareerAspirationInput struct {
	WorkerID                     operation.Field[string]
	TargetJobID                  operation.Field[string]
	CareerPathTransitionID       operation.Field[string]
	PreferenceRank               operation.Field[int32]
	TargetDate                   operation.Field[time.Time]
	MobilityPreference           operation.Field[string]
	PreferredLocationReferenceID operation.Field[string]
	WithdrawnAt                  operation.Field[time.Time]
}
