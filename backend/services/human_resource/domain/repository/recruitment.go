// Sumber model adalah baseline SQL HRM; aturan bisnis tetap milik service.
package repository

import (
	"context"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
)

// RequisitionRepository adalah port persistensi hr_requisitions.
type RequisitionRepository interface {
	Repository[entity.Requisition, string]
	InsertRequisitions(context.Context, *entity.Requisition) error
	UpdateRequisitions(context.Context, *entity.Requisition) error
	DeleteRequisitionsByUUID(context.Context, string) error
	FindRequisitionsByUUID(context.Context, string) (*entity.Requisition, error)
	FindRequisitionsByUUIDForUpdate(context.Context, string) (*entity.Requisition, error)
	GetAllRequisitions(context.Context, int, int) ([]entity.Requisition, error)
	CountRequisitions(context.Context) (int64, error)
	SelectRequisitionsWhere(context.Context, []Filter, int, int) ([]entity.Requisition, error)
	CountRequisitionsWhere(context.Context, []Filter) (int64, error)
	FindRequisitionsByRequisitionNumber(context.Context, string, int, int) ([]entity.Requisition, error)
	FindRequisitionsByWorkforcePlanLineID(context.Context, *string, int, int) ([]entity.Requisition, error)
	FindRequisitionsByPositionID(context.Context, *string, int, int) ([]entity.Requisition, error)
	FindRequisitionsByJobID(context.Context, string, int, int) ([]entity.Requisition, error)
	FindRequisitionsByOrganizationUnitReferenceID(context.Context, string, int, int) ([]entity.Requisition, error)
	FindRequisitionsByHiringManagerPrincipalID(context.Context, string, int, int) ([]entity.Requisition, error)
	FindRequisitionsByRecruiterPrincipalID(context.Context, *string, int, int) ([]entity.Requisition, error)
	FindRequisitionsByRequisitionType(context.Context, string, int, int) ([]entity.Requisition, error)
	FindRequisitionsByStatus(context.Context, string, int, int) ([]entity.Requisition, error)
	FindRequisitionsByBudgetReferenceID(context.Context, *string, int, int) ([]entity.Requisition, error)
	FindRequisitionsByWorkLocationReferenceID(context.Context, *string, int, int) ([]entity.Requisition, error)
	FindRequisitionsByApprovedByPrincipalID(context.Context, *string, int, int) ([]entity.Requisition, error)
	FindRequisitionsByRejectionReasonCode(context.Context, *string, int, int) ([]entity.Requisition, error)
	FindRequisitionsByRejectedByPrincipalID(context.Context, *string, int, int) ([]entity.Requisition, error)
	FindRequisitionsByIdempotencyKey(context.Context, string, int, int) ([]entity.Requisition, error)
	FindByRequisitionNumber(ctx context.Context, value string, limit, offset int) ([]entity.Requisition, error)
	FindByWorkforcePlanLineID(ctx context.Context, value *string, limit, offset int) ([]entity.Requisition, error)
	FindByPositionID(ctx context.Context, value *string, limit, offset int) ([]entity.Requisition, error)
	FindByJobID(ctx context.Context, value string, limit, offset int) ([]entity.Requisition, error)
	FindByOrganizationUnitReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.Requisition, error)
	FindByHiringManagerPrincipalID(ctx context.Context, value string, limit, offset int) ([]entity.Requisition, error)
	FindByRecruiterPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.Requisition, error)
	FindByRequisitionType(ctx context.Context, value string, limit, offset int) ([]entity.Requisition, error)
	FindByStatus(ctx context.Context, value string, limit, offset int) ([]entity.Requisition, error)
	FindByBudgetReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.Requisition, error)
	FindByWorkLocationReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.Requisition, error)
	FindByApprovedByPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.Requisition, error)
	FindByRejectionReasonCode(ctx context.Context, value *string, limit, offset int) ([]entity.Requisition, error)
	FindByRejectedByPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.Requisition, error)
	FindByIdempotencyKey(ctx context.Context, value string, limit, offset int) ([]entity.Requisition, error)
}

// CandidateRepository adalah port persistensi hr_candidates.
type CandidateRepository interface {
	Repository[entity.Candidate, string]
	InsertCandidates(context.Context, *entity.Candidate) error
	UpdateCandidates(context.Context, *entity.Candidate) error
	DeleteCandidatesByUUID(context.Context, string) error
	FindCandidatesByUUID(context.Context, string) (*entity.Candidate, error)
	FindCandidatesByUUIDForUpdate(context.Context, string) (*entity.Candidate, error)
	GetAllCandidates(context.Context, int, int) ([]entity.Candidate, error)
	CountCandidates(context.Context) (int64, error)
	SelectCandidatesWhere(context.Context, []Filter, int, int) ([]entity.Candidate, error)
	CountCandidatesWhere(context.Context, []Filter) (int64, error)
	FindCandidatesByCandidateNumber(context.Context, string, int, int) ([]entity.Candidate, error)
	FindCandidatesByPersonPartyReferenceID(context.Context, *string, int, int) ([]entity.Candidate, error)
	FindCandidatesByContactReferenceID(context.Context, *string, int, int) ([]entity.Candidate, error)
	FindCandidatesBySourceType(context.Context, *string, int, int) ([]entity.Candidate, error)
	FindCandidatesByCandidateStatus(context.Context, string, int, int) ([]entity.Candidate, error)
	FindCandidatesByCreatedByPrincipalID(context.Context, *string, int, int) ([]entity.Candidate, error)
	FindByCandidateNumber(ctx context.Context, value string, limit, offset int) ([]entity.Candidate, error)
	FindByPersonPartyReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.Candidate, error)
	FindByContactReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.Candidate, error)
	FindBySourceType(ctx context.Context, value *string, limit, offset int) ([]entity.Candidate, error)
	FindByCandidateStatus(ctx context.Context, value string, limit, offset int) ([]entity.Candidate, error)
	FindByCreatedByPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.Candidate, error)
}

// ApplicationRepository adalah port persistensi hr_applications.
type ApplicationRepository interface {
	Repository[entity.Application, string]
	InsertApplications(context.Context, *entity.Application) error
	UpdateApplications(context.Context, *entity.Application) error
	DeleteApplicationsByUUID(context.Context, string) error
	FindApplicationsByUUID(context.Context, string) (*entity.Application, error)
	FindApplicationsByUUIDForUpdate(context.Context, string) (*entity.Application, error)
	GetAllApplications(context.Context, int, int) ([]entity.Application, error)
	CountApplications(context.Context) (int64, error)
	SelectApplicationsWhere(context.Context, []Filter, int, int) ([]entity.Application, error)
	CountApplicationsWhere(context.Context, []Filter) (int64, error)
	FindApplicationsByCandidateID(context.Context, string, int, int) ([]entity.Application, error)
	FindApplicationsByRequisitionID(context.Context, string, int, int) ([]entity.Application, error)
	FindApplicationsByApplicationNumber(context.Context, string, int, int) ([]entity.Application, error)
	FindApplicationsByStatus(context.Context, entity.ApplicationStatus, int, int) ([]entity.Application, error)
	FindApplicationsByCurrentStageCode(context.Context, *string, int, int) ([]entity.Application, error)
	FindApplicationsByAssignedRecruiterPrincipalID(context.Context, *string, int, int) ([]entity.Application, error)
	FindApplicationsByDispositionReasonCode(context.Context, *string, int, int) ([]entity.Application, error)
	FindApplicationsByDecisionByPrincipalID(context.Context, *string, int, int) ([]entity.Application, error)
	FindApplicationsByHiredWorkerID(context.Context, *string, int, int) ([]entity.Application, error)
	FindApplicationsBySourcePostingID(context.Context, *string, int, int) ([]entity.Application, error)
	FindApplicationsByIdempotencyKey(context.Context, string, int, int) ([]entity.Application, error)
	FindByCandidateID(ctx context.Context, value string, limit, offset int) ([]entity.Application, error)
	FindByRequisitionID(ctx context.Context, value string, limit, offset int) ([]entity.Application, error)
	FindByApplicationNumber(ctx context.Context, value string, limit, offset int) ([]entity.Application, error)
	FindByStatus(ctx context.Context, value entity.ApplicationStatus, limit, offset int) ([]entity.Application, error)
	FindByCurrentStageCode(ctx context.Context, value *string, limit, offset int) ([]entity.Application, error)
	FindByAssignedRecruiterPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.Application, error)
	FindByDispositionReasonCode(ctx context.Context, value *string, limit, offset int) ([]entity.Application, error)
	FindByDecisionByPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.Application, error)
	FindByHiredWorkerID(ctx context.Context, value *string, limit, offset int) ([]entity.Application, error)
	FindBySourcePostingID(ctx context.Context, value *string, limit, offset int) ([]entity.Application, error)
	FindByIdempotencyKey(ctx context.Context, value string, limit, offset int) ([]entity.Application, error)
}

// ApplicationStageHistoryRepository adalah port persistensi hr_application_stage_history.
type ApplicationStageHistoryRepository interface {
	Repository[entity.ApplicationStageHistory, int64]
	InsertApplicationStageHistory(context.Context, *entity.ApplicationStageHistory) error
	UpdateApplicationStageHistory(context.Context, *entity.ApplicationStageHistory) error
	DeleteApplicationStageHistoryByID(context.Context, int64) error
	FindApplicationStageHistoryByID(context.Context, int64) (*entity.ApplicationStageHistory, error)
	FindApplicationStageHistoryByIDForUpdate(context.Context, int64) (*entity.ApplicationStageHistory, error)
	GetAllApplicationStageHistory(context.Context, int, int) ([]entity.ApplicationStageHistory, error)
	CountApplicationStageHistory(context.Context) (int64, error)
	SelectApplicationStageHistoryWhere(context.Context, []Filter, int, int) ([]entity.ApplicationStageHistory, error)
	CountApplicationStageHistoryWhere(context.Context, []Filter) (int64, error)
	FindApplicationStageHistoryByApplicationID(context.Context, string, int, int) ([]entity.ApplicationStageHistory, error)
	FindApplicationStageHistoryByPreviousStatus(context.Context, *entity.ApplicationStatus, int, int) ([]entity.ApplicationStageHistory, error)
	FindApplicationStageHistoryByNewStatus(context.Context, entity.ApplicationStatus, int, int) ([]entity.ApplicationStageHistory, error)
	FindApplicationStageHistoryByStageCode(context.Context, *string, int, int) ([]entity.ApplicationStageHistory, error)
	FindApplicationStageHistoryByChangedByPrincipalID(context.Context, *string, int, int) ([]entity.ApplicationStageHistory, error)
	FindApplicationStageHistoryByReasonCode(context.Context, *string, int, int) ([]entity.ApplicationStageHistory, error)
	FindByApplicationID(ctx context.Context, value string, limit, offset int) ([]entity.ApplicationStageHistory, error)
	FindByPreviousStatus(ctx context.Context, value *entity.ApplicationStatus, limit, offset int) ([]entity.ApplicationStageHistory, error)
	FindByNewStatus(ctx context.Context, value entity.ApplicationStatus, limit, offset int) ([]entity.ApplicationStageHistory, error)
	FindByStageCode(ctx context.Context, value *string, limit, offset int) ([]entity.ApplicationStageHistory, error)
	FindByChangedByPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.ApplicationStageHistory, error)
	FindByReasonCode(ctx context.Context, value *string, limit, offset int) ([]entity.ApplicationStageHistory, error)
}

// InterviewRepository adalah port persistensi hr_interviews.
type InterviewRepository interface {
	Repository[entity.Interview, string]
	InsertInterviews(context.Context, *entity.Interview) error
	UpdateInterviews(context.Context, *entity.Interview) error
	DeleteInterviewsByUUID(context.Context, string) error
	FindInterviewsByUUID(context.Context, string) (*entity.Interview, error)
	FindInterviewsByUUIDForUpdate(context.Context, string) (*entity.Interview, error)
	GetAllInterviews(context.Context, int, int) ([]entity.Interview, error)
	CountInterviews(context.Context) (int64, error)
	SelectInterviewsWhere(context.Context, []Filter, int, int) ([]entity.Interview, error)
	CountInterviewsWhere(context.Context, []Filter) (int64, error)
	FindInterviewsByApplicationID(context.Context, string, int, int) ([]entity.Interview, error)
	FindInterviewsByInterviewType(context.Context, string, int, int) ([]entity.Interview, error)
	FindInterviewsByLocationReferenceID(context.Context, *string, int, int) ([]entity.Interview, error)
	FindInterviewsByMeetingReferenceID(context.Context, *string, int, int) ([]entity.Interview, error)
	FindInterviewsByStatus(context.Context, string, int, int) ([]entity.Interview, error)
	FindInterviewsByCreatedByPrincipalID(context.Context, *string, int, int) ([]entity.Interview, error)
	FindByApplicationID(ctx context.Context, value string, limit, offset int) ([]entity.Interview, error)
	FindByInterviewType(ctx context.Context, value string, limit, offset int) ([]entity.Interview, error)
	FindByLocationReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.Interview, error)
	FindByMeetingReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.Interview, error)
	FindByStatus(ctx context.Context, value string, limit, offset int) ([]entity.Interview, error)
	FindByCreatedByPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.Interview, error)
}

// InterviewPanelistRepository adalah port persistensi hr_interview_panelists.
type InterviewPanelistRepository interface {
	Repository[entity.InterviewPanelist, string]
	InsertInterviewPanelists(context.Context, *entity.InterviewPanelist) error
	UpdateInterviewPanelists(context.Context, *entity.InterviewPanelist) error
	DeleteInterviewPanelistsByUUID(context.Context, string) error
	FindInterviewPanelistsByUUID(context.Context, string) (*entity.InterviewPanelist, error)
	FindInterviewPanelistsByUUIDForUpdate(context.Context, string) (*entity.InterviewPanelist, error)
	GetAllInterviewPanelists(context.Context, int, int) ([]entity.InterviewPanelist, error)
	CountInterviewPanelists(context.Context) (int64, error)
	SelectInterviewPanelistsWhere(context.Context, []Filter, int, int) ([]entity.InterviewPanelist, error)
	CountInterviewPanelistsWhere(context.Context, []Filter) (int64, error)
	FindInterviewPanelistsByInterviewID(context.Context, string, int, int) ([]entity.InterviewPanelist, error)
	FindInterviewPanelistsByInterviewerPrincipalID(context.Context, string, int, int) ([]entity.InterviewPanelist, error)
	FindInterviewPanelistsByScorecardTemplateReferenceID(context.Context, *string, int, int) ([]entity.InterviewPanelist, error)
	FindByInterviewID(ctx context.Context, value string, limit, offset int) ([]entity.InterviewPanelist, error)
	FindByInterviewerPrincipalID(ctx context.Context, value string, limit, offset int) ([]entity.InterviewPanelist, error)
	FindByScorecardTemplateReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.InterviewPanelist, error)
}

// InterviewFeedbackRepository adalah port persistensi hr_interview_feedback.
type InterviewFeedbackRepository interface {
	Repository[entity.InterviewFeedback, string]
	InsertInterviewFeedback(context.Context, *entity.InterviewFeedback) error
	UpdateInterviewFeedback(context.Context, *entity.InterviewFeedback) error
	DeleteInterviewFeedbackByUUID(context.Context, string) error
	FindInterviewFeedbackByUUID(context.Context, string) (*entity.InterviewFeedback, error)
	FindInterviewFeedbackByUUIDForUpdate(context.Context, string) (*entity.InterviewFeedback, error)
	GetAllInterviewFeedback(context.Context, int, int) ([]entity.InterviewFeedback, error)
	CountInterviewFeedback(context.Context) (int64, error)
	SelectInterviewFeedbackWhere(context.Context, []Filter, int, int) ([]entity.InterviewFeedback, error)
	CountInterviewFeedbackWhere(context.Context, []Filter) (int64, error)
	FindInterviewFeedbackByInterviewID(context.Context, string, int, int) ([]entity.InterviewFeedback, error)
	FindInterviewFeedbackByInterviewerPrincipalID(context.Context, string, int, int) ([]entity.InterviewFeedback, error)
	FindByInterviewID(ctx context.Context, value string, limit, offset int) ([]entity.InterviewFeedback, error)
	FindByInterviewerPrincipalID(ctx context.Context, value string, limit, offset int) ([]entity.InterviewFeedback, error)
}

// OfferRepository adalah port persistensi hr_offers.
type OfferRepository interface {
	Repository[entity.Offer, string]
	InsertOffers(context.Context, *entity.Offer) error
	UpdateOffers(context.Context, *entity.Offer) error
	DeleteOffersByUUID(context.Context, string) error
	FindOffersByUUID(context.Context, string) (*entity.Offer, error)
	FindOffersByUUIDForUpdate(context.Context, string) (*entity.Offer, error)
	GetAllOffers(context.Context, int, int) ([]entity.Offer, error)
	CountOffers(context.Context) (int64, error)
	SelectOffersWhere(context.Context, []Filter, int, int) ([]entity.Offer, error)
	CountOffersWhere(context.Context, []Filter) (int64, error)
	FindOffersByApplicationID(context.Context, string, int, int) ([]entity.Offer, error)
	FindOffersByOfferNumber(context.Context, string, int, int) ([]entity.Offer, error)
	FindOffersByOfferStatus(context.Context, string, int, int) ([]entity.Offer, error)
	FindOffersByProposedJobID(context.Context, string, int, int) ([]entity.Offer, error)
	FindOffersByProposedLegalEntityReferenceID(context.Context, string, int, int) ([]entity.Offer, error)
	FindOffersByProposedContractType(context.Context, string, int, int) ([]entity.Offer, error)
	FindOffersByProposedCompensationDocumentVersionReferenceID(context.Context, *string, int, int) ([]entity.Offer, error)
	FindOffersByCurrencyCode(context.Context, *string, int, int) ([]entity.Offer, error)
	FindOffersByOfferDocumentReferenceID(context.Context, *string, int, int) ([]entity.Offer, error)
	FindOffersByOfferDocumentVersionReferenceID(context.Context, *string, int, int) ([]entity.Offer, error)
	FindOffersByApprovedByPrincipalID(context.Context, *string, int, int) ([]entity.Offer, error)
	FindByApplicationID(ctx context.Context, value string, limit, offset int) ([]entity.Offer, error)
	FindByOfferNumber(ctx context.Context, value string, limit, offset int) ([]entity.Offer, error)
	FindByOfferStatus(ctx context.Context, value string, limit, offset int) ([]entity.Offer, error)
	FindByProposedJobID(ctx context.Context, value string, limit, offset int) ([]entity.Offer, error)
	FindByProposedLegalEntityReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.Offer, error)
	FindByProposedContractType(ctx context.Context, value string, limit, offset int) ([]entity.Offer, error)
	FindByProposedCompensationDocumentVersionReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.Offer, error)
	FindByCurrencyCode(ctx context.Context, value *string, limit, offset int) ([]entity.Offer, error)
	FindByOfferDocumentReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.Offer, error)
	FindByOfferDocumentVersionReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.Offer, error)
	FindByApprovedByPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.Offer, error)
}

// JobPostingRepository adalah port persistensi hr_job_postings.
type JobPostingRepository interface {
	Repository[entity.JobPosting, string]
	InsertJobPostings(context.Context, *entity.JobPosting) error
	UpdateJobPostings(context.Context, *entity.JobPosting) error
	DeleteJobPostingsByUUID(context.Context, string) error
	FindJobPostingsByUUID(context.Context, string) (*entity.JobPosting, error)
	FindJobPostingsByUUIDForUpdate(context.Context, string) (*entity.JobPosting, error)
	GetAllJobPostings(context.Context, int, int) ([]entity.JobPosting, error)
	CountJobPostings(context.Context) (int64, error)
	SelectJobPostingsWhere(context.Context, []Filter, int, int) ([]entity.JobPosting, error)
	CountJobPostingsWhere(context.Context, []Filter) (int64, error)
	FindJobPostingsByRequisitionID(context.Context, string, int, int) ([]entity.JobPosting, error)
	FindJobPostingsByChannelCode(context.Context, string, int, int) ([]entity.JobPosting, error)
	FindJobPostingsByContentDocumentVersionReferenceID(context.Context, string, int, int) ([]entity.JobPosting, error)
	FindJobPostingsByStatus(context.Context, string, int, int) ([]entity.JobPosting, error)
	FindJobPostingsByCreatedByPrincipalID(context.Context, string, int, int) ([]entity.JobPosting, error)
	FindJobPostingsByWithdrawalReasonCode(context.Context, string, int, int) ([]entity.JobPosting, error)
	FindJobPostingsByWithdrawnByPrincipalID(context.Context, string, int, int) ([]entity.JobPosting, error)
	FindByRequisitionID(ctx context.Context, value string, limit, offset int) ([]entity.JobPosting, error)
	FindByChannelCode(ctx context.Context, value string, limit, offset int) ([]entity.JobPosting, error)
	FindByContentDocumentVersionReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.JobPosting, error)
	FindByStatus(ctx context.Context, value string, limit, offset int) ([]entity.JobPosting, error)
	FindByCreatedByPrincipalID(ctx context.Context, value string, limit, offset int) ([]entity.JobPosting, error)
	FindByWithdrawalReasonCode(ctx context.Context, value string, limit, offset int) ([]entity.JobPosting, error)
	FindByWithdrawnByPrincipalID(ctx context.Context, value string, limit, offset int) ([]entity.JobPosting, error)
}

// ApplicationAssessmentRepository adalah port persistensi hr_application_assessments.
type ApplicationAssessmentRepository interface {
	Repository[entity.ApplicationAssessment, string]
	InsertApplicationAssessments(context.Context, *entity.ApplicationAssessment) error
	UpdateApplicationAssessments(context.Context, *entity.ApplicationAssessment) error
	DeleteApplicationAssessmentsByUUID(context.Context, string) error
	FindApplicationAssessmentsByUUID(context.Context, string) (*entity.ApplicationAssessment, error)
	FindApplicationAssessmentsByUUIDForUpdate(context.Context, string) (*entity.ApplicationAssessment, error)
	GetAllApplicationAssessments(context.Context, int, int) ([]entity.ApplicationAssessment, error)
	CountApplicationAssessments(context.Context) (int64, error)
	SelectApplicationAssessmentsWhere(context.Context, []Filter, int, int) ([]entity.ApplicationAssessment, error)
	CountApplicationAssessmentsWhere(context.Context, []Filter) (int64, error)
	FindApplicationAssessmentsByApplicationID(context.Context, string, int, int) ([]entity.ApplicationAssessment, error)
	FindApplicationAssessmentsByAssessmentCode(context.Context, string, int, int) ([]entity.ApplicationAssessment, error)
	FindApplicationAssessmentsByAttemptNumber(context.Context, int32, int, int) ([]entity.ApplicationAssessment, error)
	FindApplicationAssessmentsByRubricDocumentVersionReferenceID(context.Context, string, int, int) ([]entity.ApplicationAssessment, error)
	FindApplicationAssessmentsByStatus(context.Context, string, int, int) ([]entity.ApplicationAssessment, error)
	FindApplicationAssessmentsByAssessorPrincipalID(context.Context, *string, int, int) ([]entity.ApplicationAssessment, error)
	FindApplicationAssessmentsByRecommendationCode(context.Context, *string, int, int) ([]entity.ApplicationAssessment, error)
	FindApplicationAssessmentsByEvidenceDocumentVersionReferenceID(context.Context, *string, int, int) ([]entity.ApplicationAssessment, error)
	FindByApplicationID(ctx context.Context, value string, limit, offset int) ([]entity.ApplicationAssessment, error)
	FindByAssessmentCode(ctx context.Context, value string, limit, offset int) ([]entity.ApplicationAssessment, error)
	FindByAttemptNumber(ctx context.Context, value int32, limit, offset int) ([]entity.ApplicationAssessment, error)
	FindByRubricDocumentVersionReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.ApplicationAssessment, error)
	FindByStatus(ctx context.Context, value string, limit, offset int) ([]entity.ApplicationAssessment, error)
	FindByAssessorPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.ApplicationAssessment, error)
	FindByRecommendationCode(ctx context.Context, value *string, limit, offset int) ([]entity.ApplicationAssessment, error)
	FindByEvidenceDocumentVersionReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.ApplicationAssessment, error)
}
