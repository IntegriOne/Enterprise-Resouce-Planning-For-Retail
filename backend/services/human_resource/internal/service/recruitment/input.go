// Kerangka awal dibuat dengan alat bantu; kode ini dipelihara langsung per domain.
// Aturan bisnis lintas record tetap dijalankan melalui policy/integrity dan command domain.
package service

import (
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

// CreateRequisitionInput adalah object input aplikasi, bukan entity ORM.
type CreateRequisitionInput struct {
	RequisitionNumber           string
	WorkforcePlanLineID         *string
	PositionID                  *string
	JobID                       string
	OrganizationUnitReferenceID string
	HiringManagerPrincipalID    string
	HiringManagerWorkerID       *string
	RecruiterPrincipalID        *string
	RequisitionType             string
	Status                      *string
	Headcount                   *int32
	TargetStartDate             *time.Time
	BudgetReferenceID           *string
	WorkLocationReferenceID     *string
	EmploymentCategory          entity.WorkerCategory
	MinimumQualifications       entity.JSON
	DesiredQualifications       entity.JSON
	ApprovedAt                  *time.Time
	ApprovedByPrincipalID       *string
	OpenedAt                    *time.Time
	ClosedAt                    *time.Time
	IdempotencyKey              string
}

// UpdateRequisitionInput membedakan field absen, null dan nilai eksplisit.
type UpdateRequisitionInput struct {
	RequisitionNumber           operation.Field[string]
	WorkforcePlanLineID         operation.Field[string]
	PositionID                  operation.Field[string]
	JobID                       operation.Field[string]
	OrganizationUnitReferenceID operation.Field[string]
	HiringManagerPrincipalID    operation.Field[string]
	HiringManagerWorkerID       *string
	RecruiterPrincipalID        operation.Field[string]
	RequisitionType             operation.Field[string]
	Status                      operation.Field[string]
	Headcount                   operation.Field[int32]
	TargetStartDate             operation.Field[time.Time]
	BudgetReferenceID           operation.Field[string]
	WorkLocationReferenceID     operation.Field[string]
	EmploymentCategory          operation.Field[entity.WorkerCategory]
	MinimumQualifications       operation.Field[entity.JSON]
	DesiredQualifications       operation.Field[entity.JSON]
	ApprovedAt                  operation.Field[time.Time]
	ApprovedByPrincipalID       operation.Field[string]
	OpenedAt                    operation.Field[time.Time]
	ClosedAt                    operation.Field[time.Time]
	IdempotencyKey              operation.Field[string]
}

// CreateCandidateInput adalah object input aplikasi, bukan entity ORM.
type CreateCandidateInput struct {
	CandidateNumber        string
	PersonPartyReferenceID *string
	PreferredName          *string
	ContactReferenceID     *string
	SourceType             *string
	SourceReference        *string
	PrivacyNoticeVersion   string
	NoticeAcknowledgedAt   *time.Time
	DataRetentionUntil     *time.Time
	CandidateStatus        *string
}

// UpdateCandidateInput membedakan field absen, null dan nilai eksplisit.
type UpdateCandidateInput struct {
	CandidateNumber        operation.Field[string]
	PersonPartyReferenceID operation.Field[string]
	PreferredName          operation.Field[string]
	ContactReferenceID     operation.Field[string]
	SourceType             operation.Field[string]
	SourceReference        operation.Field[string]
	PrivacyNoticeVersion   operation.Field[string]
	NoticeAcknowledgedAt   operation.Field[time.Time]
	DataRetentionUntil     operation.Field[time.Time]
	CandidateStatus        operation.Field[string]
}

// CreateApplicationInput adalah object input aplikasi, bukan entity ORM.
type CreateApplicationInput struct {
	CandidateID                  string
	RequisitionID                string
	ApplicationNumber            string
	Status                       *entity.ApplicationStatus
	CurrentStageCode             *string
	AssignedRecruiterPrincipalID *string
	ScreeningScore               *entity.Decimal
	ScorecardVersion             *string
	DispositionReasonCode        *string
	DecisionAt                   *time.Time
	DecisionByPrincipalID        *string
	HiredWorkerID                *string
	SourcePostingID              *string
	IdempotencyKey               string
}

// UpdateApplicationInput membedakan field absen, null dan nilai eksplisit.
type UpdateApplicationInput struct {
	CandidateID                  operation.Field[string]
	RequisitionID                operation.Field[string]
	ApplicationNumber            operation.Field[string]
	Status                       operation.Field[entity.ApplicationStatus]
	CurrentStageCode             operation.Field[string]
	AssignedRecruiterPrincipalID operation.Field[string]
	ScreeningScore               operation.Field[entity.Decimal]
	ScorecardVersion             operation.Field[string]
	DispositionReasonCode        operation.Field[string]
	DecisionAt                   operation.Field[time.Time]
	DecisionByPrincipalID        operation.Field[string]
	HiredWorkerID                operation.Field[string]
	SourcePostingID              operation.Field[string]
	IdempotencyKey               operation.Field[string]
}

// CreateApplicationStageHistoryInput adalah object input aplikasi, bukan entity ORM.
type CreateApplicationStageHistoryInput struct {
	ApplicationID        string
	PreviousStatus       *entity.ApplicationStatus
	NewStatus            entity.ApplicationStatus
	StageCode            *string
	ChangedByPrincipalID *string
	ReasonCode           *string
}

// UpdateApplicationStageHistoryInput membedakan field absen, null dan nilai eksplisit.
type UpdateApplicationStageHistoryInput struct {
	ApplicationID        operation.Field[string]
	PreviousStatus       operation.Field[entity.ApplicationStatus]
	NewStatus            operation.Field[entity.ApplicationStatus]
	StageCode            operation.Field[string]
	ChangedByPrincipalID operation.Field[string]
	ReasonCode           operation.Field[string]
}

// CreateInterviewInput adalah object input aplikasi, bukan entity ORM.
type CreateInterviewInput struct {
	ApplicationID       string
	InterviewRound      int32
	InterviewType       string
	ScheduledStart      *time.Time
	ScheduledEnd        *time.Time
	LocationReferenceID *string
	MeetingReferenceID  *string
	Status              *string
	CandidateTimezone   *string
}

// UpdateInterviewInput membedakan field absen, null dan nilai eksplisit.
type UpdateInterviewInput struct {
	ApplicationID       operation.Field[string]
	InterviewRound      operation.Field[int32]
	InterviewType       operation.Field[string]
	ScheduledStart      operation.Field[time.Time]
	ScheduledEnd        operation.Field[time.Time]
	LocationReferenceID operation.Field[string]
	MeetingReferenceID  operation.Field[string]
	Status              operation.Field[string]
	CandidateTimezone   operation.Field[string]
}

// CreateInterviewPanelistInput adalah object input aplikasi, bukan entity ORM.
type CreateInterviewPanelistInput struct {
	InterviewID                  string
	InterviewerPrincipalID       string
	InterviewerRole              *string
	ScorecardTemplateReferenceID *string
	FeedbackDueAt                *time.Time
}

// UpdateInterviewPanelistInput membedakan field absen, null dan nilai eksplisit.
type UpdateInterviewPanelistInput struct {
	InterviewID                  operation.Field[string]
	InterviewerPrincipalID       operation.Field[string]
	InterviewerRole              operation.Field[string]
	ScorecardTemplateReferenceID operation.Field[string]
	FeedbackDueAt                operation.Field[time.Time]
}

// CreateInterviewFeedbackInput adalah object input aplikasi, bukan entity ORM.
type CreateInterviewFeedbackInput struct {
	InterviewID            string
	InterviewerPrincipalID string
	FeedbackVersion        *int32
	OverallRecommendation  string
	CompetencyScores       entity.JSON
	EvidenceNotes          *string
	ContentDigest          []byte
	VisibilityClass        *string
}

// UpdateInterviewFeedbackInput membedakan field absen, null dan nilai eksplisit.
type UpdateInterviewFeedbackInput struct {
	InterviewID            operation.Field[string]
	InterviewerPrincipalID operation.Field[string]
	FeedbackVersion        operation.Field[int32]
	OverallRecommendation  operation.Field[string]
	CompetencyScores       operation.Field[entity.JSON]
	EvidenceNotes          operation.Field[string]
	ContentDigest          operation.Field[[]byte]
	VisibilityClass        operation.Field[string]
}

// CreateOfferInput adalah object input aplikasi, bukan entity ORM.
type CreateOfferInput struct {
	ApplicationID                                  string
	OfferNumber                                    string
	OfferVersion                                   *int32
	OfferStatus                                    *string
	ProposedJobID                                  string
	ProposedLegalEntityReferenceID                 string
	ProposedStartDate                              *time.Time
	ProposedContractType                           string
	ProposedCompensationDocumentVersionReferenceID *string
	CurrencyCode                                   *string
	OfferDocumentReferenceID                       *string
	OfferDocumentVersionReferenceID                *string
	OfferContentDigest                             []byte
	IssuedAt                                       *time.Time
	ExpiresAt                                      *time.Time
	CandidateResponse                              *string
	RespondedAt                                    *time.Time
	ApprovedByPrincipalID                          *string
	ApprovedAt                                     *time.Time
}

// UpdateOfferInput membedakan field absen, null dan nilai eksplisit.
type UpdateOfferInput struct {
	ApplicationID                                  operation.Field[string]
	OfferNumber                                    operation.Field[string]
	OfferVersion                                   operation.Field[int32]
	OfferStatus                                    operation.Field[string]
	ProposedJobID                                  operation.Field[string]
	ProposedLegalEntityReferenceID                 operation.Field[string]
	ProposedStartDate                              operation.Field[time.Time]
	ProposedContractType                           operation.Field[string]
	ProposedCompensationDocumentVersionReferenceID operation.Field[string]
	CurrencyCode                                   operation.Field[string]
	OfferDocumentReferenceID                       operation.Field[string]
	OfferDocumentVersionReferenceID                operation.Field[string]
	OfferContentDigest                             operation.Field[[]byte]
	IssuedAt                                       operation.Field[time.Time]
	ExpiresAt                                      operation.Field[time.Time]
	CandidateResponse                              operation.Field[string]
	RespondedAt                                    operation.Field[time.Time]
	ApprovedByPrincipalID                          operation.Field[string]
	ApprovedAt                                     operation.Field[time.Time]
}

// CreateJobPostingInput adalah object input aplikasi, bukan entity ORM.
type CreateJobPostingInput struct {
	RequisitionID                     string
	ChannelCode                       string
	PostingVersion                    int32
	ExternalPostingReference          *string
	ContentDocumentVersionReferenceID string
	ContentDigest                     []byte
	Status                            *string
	PublishedAt                       *time.Time
	ClosesAt                          *time.Time
}

// UpdateJobPostingInput membedakan field absen, null dan nilai eksplisit.
type UpdateJobPostingInput struct {
	RequisitionID                     operation.Field[string]
	ChannelCode                       operation.Field[string]
	PostingVersion                    operation.Field[int32]
	ExternalPostingReference          operation.Field[string]
	ContentDocumentVersionReferenceID operation.Field[string]
	ContentDigest                     operation.Field[[]byte]
	Status                            operation.Field[string]
	PublishedAt                       operation.Field[time.Time]
	ClosesAt                          operation.Field[time.Time]
}

// CreateApplicationAssessmentInput adalah object input aplikasi, bukan entity ORM.
type CreateApplicationAssessmentInput struct {
	ApplicationID                      string
	AssessmentCode                     string
	AttemptNumber                      *int32
	RubricDocumentVersionReferenceID   string
	RubricDigest                       []byte
	Status                             *string
	AssessorPrincipalID                *string
	ScheduledAt                        *time.Time
	CompletedAt                        *time.Time
	NormalizedScore                    *entity.Decimal
	RecommendationCode                 *string
	EvidenceDocumentVersionReferenceID *string
	EvidenceDigest                     []byte
}

// UpdateApplicationAssessmentInput membedakan field absen, null dan nilai eksplisit.
type UpdateApplicationAssessmentInput struct {
	ApplicationID                      operation.Field[string]
	AssessmentCode                     operation.Field[string]
	AttemptNumber                      operation.Field[int32]
	RubricDocumentVersionReferenceID   operation.Field[string]
	RubricDigest                       operation.Field[[]byte]
	Status                             operation.Field[string]
	AssessorPrincipalID                operation.Field[string]
	ScheduledAt                        operation.Field[time.Time]
	CompletedAt                        operation.Field[time.Time]
	NormalizedScore                    operation.Field[entity.Decimal]
	RecommendationCode                 operation.Field[string]
	EvidenceDocumentVersionReferenceID operation.Field[string]
	EvidenceDigest                     operation.Field[[]byte]
}
