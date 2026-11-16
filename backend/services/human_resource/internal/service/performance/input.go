// Kerangka awal dibuat dengan alat bantu; kode ini dipelihara langsung per domain.
// Aturan bisnis lintas record tetap dijalankan melalui policy/integrity dan command domain.
package service

import (
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

// CreatePerformanceCycleInput adalah object input aplikasi, bukan entity ORM.
type CreatePerformanceCycleInput struct {
	CycleCode             string
	CycleName             string
	CycleType             string
	Status                *entity.PerformanceCycleStatus
	PeriodStart           time.Time
	PeriodEnd             time.Time
	SelfReviewDueAt       *time.Time
	ManagerReviewDueAt    *time.Time
	CalibrationDueAt      *time.Time
	LaunchedAt            *time.Time
	LaunchedByPrincipalID *string
	Configuration         entity.JSON
}

// UpdatePerformanceCycleInput membedakan field absen, null dan nilai eksplisit.
type UpdatePerformanceCycleInput struct {
	CycleCode             operation.Field[string]
	CycleName             operation.Field[string]
	CycleType             operation.Field[string]
	Status                operation.Field[entity.PerformanceCycleStatus]
	PeriodStart           operation.Field[time.Time]
	PeriodEnd             operation.Field[time.Time]
	SelfReviewDueAt       operation.Field[time.Time]
	ManagerReviewDueAt    operation.Field[time.Time]
	CalibrationDueAt      operation.Field[time.Time]
	LaunchedAt            operation.Field[time.Time]
	LaunchedByPrincipalID operation.Field[string]
	Configuration         operation.Field[entity.JSON]
}

// CreatePerformanceGoalInput adalah object input aplikasi, bukan entity ORM.
type CreatePerformanceGoalInput struct {
	WorkerID              string
	PerformanceCycleID    string
	ParentGoalID          *string
	GoalCode              *string
	GoalTitle             string
	GoalDescription       *string
	GoalType              string
	MeasureDefinition     entity.JSON
	TargetValue           *entity.Decimal
	ActualValue           *entity.Decimal
	UnitCode              *string
	WeightPercent         *entity.Decimal
	DueDate               *time.Time
	GoalStatus            *string
	ApprovedByPrincipalID *string
	ApprovedAt            *time.Time
}

// UpdatePerformanceGoalInput membedakan field absen, null dan nilai eksplisit.
type UpdatePerformanceGoalInput struct {
	WorkerID              operation.Field[string]
	PerformanceCycleID    operation.Field[string]
	ParentGoalID          operation.Field[string]
	GoalCode              operation.Field[string]
	GoalTitle             operation.Field[string]
	GoalDescription       operation.Field[string]
	GoalType              operation.Field[string]
	MeasureDefinition     operation.Field[entity.JSON]
	TargetValue           operation.Field[entity.Decimal]
	ActualValue           operation.Field[entity.Decimal]
	UnitCode              operation.Field[string]
	WeightPercent         operation.Field[entity.Decimal]
	DueDate               operation.Field[time.Time]
	GoalStatus            operation.Field[string]
	ApprovedByPrincipalID operation.Field[string]
	ApprovedAt            operation.Field[time.Time]
}

// CreatePerformanceReviewInput adalah object input aplikasi, bukan entity ORM.
type CreatePerformanceReviewInput struct {
	WorkerID             string
	PerformanceCycleID   string
	ReviewerPrincipalID  string
	ReviewerRole         string
	ReviewVersion        *int32
	OverallRatingCode    *string
	CompetencyAssessment entity.JSON
	Narrative            *string
	DevelopmentPlanID    *string
	SubmittedAt          *time.Time
	FinalizedAt          *time.Time
	VisibilityClass      *string
	ContentDigest        []byte
}

// UpdatePerformanceReviewInput membedakan field absen, null dan nilai eksplisit.
type UpdatePerformanceReviewInput struct {
	WorkerID             operation.Field[string]
	PerformanceCycleID   operation.Field[string]
	ReviewerPrincipalID  operation.Field[string]
	ReviewerRole         operation.Field[string]
	ReviewVersion        operation.Field[int32]
	OverallRatingCode    operation.Field[string]
	CompetencyAssessment operation.Field[entity.JSON]
	Narrative            operation.Field[string]
	DevelopmentPlanID    operation.Field[string]
	SubmittedAt          operation.Field[time.Time]
	FinalizedAt          operation.Field[time.Time]
	VisibilityClass      operation.Field[string]
	ContentDigest        operation.Field[[]byte]
}

// CreatePerformanceGoalUpdateInput adalah object input aplikasi, bukan entity ORM.
type CreatePerformanceGoalUpdateInput struct {
	GoalID                             string
	ObservedAt                         time.Time
	ActualValue                        entity.Decimal
	EvidenceDocumentVersionReferenceID string
	EvidenceDigest                     []byte
	SupersedesUpdateID                 *string
	IdempotencyKey                     string
}

// UpdatePerformanceGoalUpdateInput membedakan field absen, null dan nilai eksplisit.
type UpdatePerformanceGoalUpdateInput struct {
	GoalID                             operation.Field[string]
	ObservedAt                         operation.Field[time.Time]
	ActualValue                        operation.Field[entity.Decimal]
	EvidenceDocumentVersionReferenceID operation.Field[string]
	EvidenceDigest                     operation.Field[[]byte]
	SupersedesUpdateID                 operation.Field[string]
	IdempotencyKey                     operation.Field[string]
}

// CreatePerformanceReviewGoalScoreInput adalah object input aplikasi, bukan entity ORM.
type CreatePerformanceReviewGoalScoreInput struct {
	PerformanceReviewID                string
	GoalID                             string
	WorkerID                           string
	PerformanceCycleID                 string
	Rating                             int16
	WeightPercentSnapshot              entity.Decimal
	RubricDocumentVersionReferenceID   string
	EvidenceDocumentVersionReferenceID string
}

// UpdatePerformanceReviewGoalScoreInput membedakan field absen, null dan nilai eksplisit.
type UpdatePerformanceReviewGoalScoreInput struct {
	PerformanceReviewID                operation.Field[string]
	GoalID                             operation.Field[string]
	WorkerID                           operation.Field[string]
	PerformanceCycleID                 operation.Field[string]
	Rating                             operation.Field[int16]
	WeightPercentSnapshot              operation.Field[entity.Decimal]
	RubricDocumentVersionReferenceID   operation.Field[string]
	EvidenceDocumentVersionReferenceID operation.Field[string]
}
