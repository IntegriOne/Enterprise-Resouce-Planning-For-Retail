// Kerangka awal dibuat dengan alat bantu; kode ini dipelihara langsung per domain.
// Aturan bisnis lintas record tetap dijalankan melalui policy/integrity dan command domain.
package service

import (
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

// CreateTalentFrameworkInput adalah object input aplikasi, bukan entity ORM.
type CreateTalentFrameworkInput struct {
	FrameworkCode                    string
	FrameworkVersion                 int32
	FrameworkName                    string
	RubricDocumentVersionReferenceID string
	RubricDigest                     []byte
	EffectiveFrom                    time.Time
	EffectiveTo                      *time.Time
	ApprovedByPrincipalID            string
	ApprovedAt                       time.Time
}

// UpdateTalentFrameworkInput membedakan field absen, null dan nilai eksplisit.
type UpdateTalentFrameworkInput struct {
	FrameworkCode                    operation.Field[string]
	FrameworkVersion                 operation.Field[int32]
	FrameworkName                    operation.Field[string]
	RubricDocumentVersionReferenceID operation.Field[string]
	RubricDigest                     operation.Field[[]byte]
	EffectiveFrom                    operation.Field[time.Time]
	EffectiveTo                      operation.Field[time.Time]
	ApprovedByPrincipalID            operation.Field[string]
	ApprovedAt                       operation.Field[time.Time]
}

// CreateTalentDimensionInput adalah object input aplikasi, bukan entity ORM.
type CreateTalentDimensionInput struct {
	FrameworkID   string
	DimensionCode string
	DimensionName string
	Definition    string
	RatingAnchors entity.JSON
	WeightPercent entity.Decimal
	IsRequired    *bool
}

// UpdateTalentDimensionInput membedakan field absen, null dan nilai eksplisit.
type UpdateTalentDimensionInput struct {
	FrameworkID   operation.Field[string]
	DimensionCode operation.Field[string]
	DimensionName operation.Field[string]
	Definition    operation.Field[string]
	RatingAnchors operation.Field[entity.JSON]
	WeightPercent operation.Field[entity.Decimal]
	IsRequired    operation.Field[bool]
}

// CreateTalentReviewCycleInput adalah object input aplikasi, bukan entity ORM.
type CreateTalentReviewCycleInput struct {
	CycleCode                   string
	FrameworkID                 string
	LegalEntityReferenceID      string
	OrganizationUnitReferenceID *string
	PeriodStart                 time.Time
	PeriodEnd                   time.Time
	Status                      *string
	OwnerPrincipalID            string
	ClosedAt                    *time.Time
}

// UpdateTalentReviewCycleInput membedakan field absen, null dan nilai eksplisit.
type UpdateTalentReviewCycleInput struct {
	CycleCode                   operation.Field[string]
	FrameworkID                 operation.Field[string]
	LegalEntityReferenceID      operation.Field[string]
	OrganizationUnitReferenceID operation.Field[string]
	PeriodStart                 operation.Field[time.Time]
	PeriodEnd                   operation.Field[time.Time]
	Status                      operation.Field[string]
	OwnerPrincipalID            operation.Field[string]
	ClosedAt                    operation.Field[time.Time]
}

// CreateTalentAssessmentInput adalah object input aplikasi, bukan entity ORM.
type CreateTalentAssessmentInput struct {
	WorkerID                           string
	TalentReviewCycleID                string
	FrameworkID                        string
	AssessmentVersion                  int32
	PerformanceReviewID                *string
	Status                             *string
	EvidenceDocumentVersionReferenceID string
	EvidenceDigest                     []byte
	SegmentCode                        *string
	FinalizedByPrincipalID             *string
	FinalizedAt                        *time.Time
}

// UpdateTalentAssessmentInput membedakan field absen, null dan nilai eksplisit.
type UpdateTalentAssessmentInput struct {
	WorkerID                           operation.Field[string]
	TalentReviewCycleID                operation.Field[string]
	FrameworkID                        operation.Field[string]
	AssessmentVersion                  operation.Field[int32]
	PerformanceReviewID                operation.Field[string]
	Status                             operation.Field[string]
	EvidenceDocumentVersionReferenceID operation.Field[string]
	EvidenceDigest                     operation.Field[[]byte]
	SegmentCode                        operation.Field[string]
	FinalizedByPrincipalID             operation.Field[string]
	FinalizedAt                        operation.Field[time.Time]
}

// CreateTalentAssessmentScoreInput adalah object input aplikasi, bukan entity ORM.
type CreateTalentAssessmentScoreInput struct {
	AssessmentID                       string
	FrameworkID                        string
	DimensionID                        string
	Score                              int16
	EvidenceDocumentVersionReferenceID string
	EvidenceDigest                     []byte
}

// UpdateTalentAssessmentScoreInput membedakan field absen, null dan nilai eksplisit.
type UpdateTalentAssessmentScoreInput struct {
	AssessmentID                       operation.Field[string]
	FrameworkID                        operation.Field[string]
	DimensionID                        operation.Field[string]
	Score                              operation.Field[int16]
	EvidenceDocumentVersionReferenceID operation.Field[string]
	EvidenceDigest                     operation.Field[[]byte]
}

// CreateTalentCalibrationSessionInput adalah object input aplikasi, bukan entity ORM.
type CreateTalentCalibrationSessionInput struct {
	TalentReviewCycleID               string
	SessionCode                       string
	ChairPrincipalID                  string
	ScheduledAt                       time.Time
	Status                            *string
	ClosedAt                          *time.Time
	MinutesDocumentVersionReferenceID *string
}

// UpdateTalentCalibrationSessionInput membedakan field absen, null dan nilai eksplisit.
type UpdateTalentCalibrationSessionInput struct {
	TalentReviewCycleID               operation.Field[string]
	SessionCode                       operation.Field[string]
	ChairPrincipalID                  operation.Field[string]
	ScheduledAt                       operation.Field[time.Time]
	Status                            operation.Field[string]
	ClosedAt                          operation.Field[time.Time]
	MinutesDocumentVersionReferenceID operation.Field[string]
}

// CreateTalentCalibrationDecisionInput adalah object input aplikasi, bukan entity ORM.
type CreateTalentCalibrationDecisionInput struct {
	SessionID                           string
	TalentReviewCycleID                 string
	AssessmentID                        string
	DimensionID                         string
	OriginalScore                       int16
	CalibratedScore                     int16
	RationaleDocumentVersionReferenceID string
	RationaleDigest                     []byte
}

// UpdateTalentCalibrationDecisionInput membedakan field absen, null dan nilai eksplisit.
type UpdateTalentCalibrationDecisionInput struct {
	SessionID                           operation.Field[string]
	TalentReviewCycleID                 operation.Field[string]
	AssessmentID                        operation.Field[string]
	DimensionID                         operation.Field[string]
	OriginalScore                       operation.Field[int16]
	CalibratedScore                     operation.Field[int16]
	RationaleDocumentVersionReferenceID operation.Field[string]
	RationaleDigest                     operation.Field[[]byte]
}

// CreateTalentPoolInput adalah object input aplikasi, bukan entity ORM.
type CreateTalentPoolInput struct {
	PoolCode                              string
	PoolVersion                           int32
	PoolName                              string
	LegalEntityReferenceID                string
	TargetJobID                           *string
	EligibilityDocumentVersionReferenceID string
	EligibilityDigest                     []byte
	ValidFrom                             time.Time
	ValidTo                               *time.Time
	OwnerPrincipalID                      string
}

// UpdateTalentPoolInput membedakan field absen, null dan nilai eksplisit.
type UpdateTalentPoolInput struct {
	PoolCode                              operation.Field[string]
	PoolVersion                           operation.Field[int32]
	PoolName                              operation.Field[string]
	LegalEntityReferenceID                operation.Field[string]
	TargetJobID                           operation.Field[string]
	EligibilityDocumentVersionReferenceID operation.Field[string]
	EligibilityDigest                     operation.Field[[]byte]
	ValidFrom                             operation.Field[time.Time]
	ValidTo                               operation.Field[time.Time]
	OwnerPrincipalID                      operation.Field[string]
}

// CreateTalentPoolMembershipInput adalah object input aplikasi, bukan entity ORM.
type CreateTalentPoolMembershipInput struct {
	TalentPoolID           string
	WorkerID               string
	AssessmentID           string
	NominatedByPrincipalID string
	Status                 *string
	ValidFrom              time.Time
	ValidTo                *time.Time
	ApprovedByPrincipalID  *string
	ApprovedAt             *time.Time
	ExitReasonCode         *string
}

// UpdateTalentPoolMembershipInput membedakan field absen, null dan nilai eksplisit.
type UpdateTalentPoolMembershipInput struct {
	TalentPoolID           operation.Field[string]
	WorkerID               operation.Field[string]
	AssessmentID           operation.Field[string]
	NominatedByPrincipalID operation.Field[string]
	Status                 operation.Field[string]
	ValidFrom              operation.Field[time.Time]
	ValidTo                operation.Field[time.Time]
	ApprovedByPrincipalID  operation.Field[string]
	ApprovedAt             operation.Field[time.Time]
	ExitReasonCode         operation.Field[string]
}
