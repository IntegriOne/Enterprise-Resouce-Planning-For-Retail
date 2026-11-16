package repository

import (
	"context"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
)

// TalentFrameworkRepository adalah port persistensi hr_talent_frameworks.
type TalentFrameworkRepository interface {
	Repository[entity.TalentFramework, string]
	InsertTalentFrameworks(context.Context, *entity.TalentFramework) error
	UpdateTalentFrameworks(context.Context, *entity.TalentFramework) error
	FindTalentFrameworksByUUID(context.Context, string) (*entity.TalentFramework, error)
	FindTalentFrameworksByUUIDForUpdate(context.Context, string) (*entity.TalentFramework, error)
	DeleteTalentFrameworksByUUID(context.Context, string) error
	GetAllTalentFrameworks(context.Context, int, int) ([]entity.TalentFramework, error)
	CountTalentFrameworks(context.Context) (int64, error)
	SelectTalentFrameworksWhere(context.Context, []Filter, int, int) ([]entity.TalentFramework, error)
	CountTalentFrameworksWhere(context.Context, []Filter) (int64, error)
	FindTalentFrameworksByFrameworkCode(context.Context, string, int, int) ([]entity.TalentFramework, error)
	FindTalentFrameworksByRubricDocumentVersionReferenceID(context.Context, string, int, int) ([]entity.TalentFramework, error)
	FindTalentFrameworksByApprovedByPrincipalID(context.Context, string, int, int) ([]entity.TalentFramework, error)
	FindByFrameworkCode(ctx context.Context, value string, limit, offset int) ([]entity.TalentFramework, error)
	FindByRubricDocumentVersionReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.TalentFramework, error)
	FindByApprovedByPrincipalID(ctx context.Context, value string, limit, offset int) ([]entity.TalentFramework, error)
}

// TalentDimensionRepository adalah port persistensi hr_talent_dimensions.
type TalentDimensionRepository interface {
	Repository[entity.TalentDimension, string]
	InsertTalentDimensions(context.Context, *entity.TalentDimension) error
	UpdateTalentDimensions(context.Context, *entity.TalentDimension) error
	FindTalentDimensionsByUUID(context.Context, string) (*entity.TalentDimension, error)
	FindTalentDimensionsByUUIDForUpdate(context.Context, string) (*entity.TalentDimension, error)
	DeleteTalentDimensionsByUUID(context.Context, string) error
	GetAllTalentDimensions(context.Context, int, int) ([]entity.TalentDimension, error)
	CountTalentDimensions(context.Context) (int64, error)
	SelectTalentDimensionsWhere(context.Context, []Filter, int, int) ([]entity.TalentDimension, error)
	CountTalentDimensionsWhere(context.Context, []Filter) (int64, error)
	FindTalentDimensionsByFrameworkID(context.Context, string, int, int) ([]entity.TalentDimension, error)
	FindTalentDimensionsByDimensionCode(context.Context, string, int, int) ([]entity.TalentDimension, error)
	FindByFrameworkID(ctx context.Context, value string, limit, offset int) ([]entity.TalentDimension, error)
	FindByDimensionCode(ctx context.Context, value string, limit, offset int) ([]entity.TalentDimension, error)
}

// TalentReviewCycleRepository adalah port persistensi hr_talent_review_cycles.
type TalentReviewCycleRepository interface {
	Repository[entity.TalentReviewCycle, string]
	InsertTalentReviewCycles(context.Context, *entity.TalentReviewCycle) error
	UpdateTalentReviewCycles(context.Context, *entity.TalentReviewCycle) error
	FindTalentReviewCyclesByUUID(context.Context, string) (*entity.TalentReviewCycle, error)
	FindTalentReviewCyclesByUUIDForUpdate(context.Context, string) (*entity.TalentReviewCycle, error)
	DeleteTalentReviewCyclesByUUID(context.Context, string) error
	GetAllTalentReviewCycles(context.Context, int, int) ([]entity.TalentReviewCycle, error)
	CountTalentReviewCycles(context.Context) (int64, error)
	SelectTalentReviewCyclesWhere(context.Context, []Filter, int, int) ([]entity.TalentReviewCycle, error)
	CountTalentReviewCyclesWhere(context.Context, []Filter) (int64, error)
	FindTalentReviewCyclesByCycleCode(context.Context, string, int, int) ([]entity.TalentReviewCycle, error)
	FindTalentReviewCyclesByFrameworkID(context.Context, string, int, int) ([]entity.TalentReviewCycle, error)
	FindTalentReviewCyclesByLegalEntityReferenceID(context.Context, string, int, int) ([]entity.TalentReviewCycle, error)
	FindTalentReviewCyclesByOrganizationUnitReferenceID(context.Context, *string, int, int) ([]entity.TalentReviewCycle, error)
	FindTalentReviewCyclesByStatus(context.Context, string, int, int) ([]entity.TalentReviewCycle, error)
	FindTalentReviewCyclesByOwnerPrincipalID(context.Context, string, int, int) ([]entity.TalentReviewCycle, error)
	FindByCycleCode(ctx context.Context, value string, limit, offset int) ([]entity.TalentReviewCycle, error)
	FindByFrameworkID(ctx context.Context, value string, limit, offset int) ([]entity.TalentReviewCycle, error)
	FindByLegalEntityReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.TalentReviewCycle, error)
	FindByOrganizationUnitReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.TalentReviewCycle, error)
	FindByStatus(ctx context.Context, value string, limit, offset int) ([]entity.TalentReviewCycle, error)
	FindByOwnerPrincipalID(ctx context.Context, value string, limit, offset int) ([]entity.TalentReviewCycle, error)
}

// TalentAssessmentRepository adalah port persistensi hr_talent_assessments.
type TalentAssessmentRepository interface {
	Repository[entity.TalentAssessment, string]
	InsertTalentAssessments(context.Context, *entity.TalentAssessment) error
	UpdateTalentAssessments(context.Context, *entity.TalentAssessment) error
	FindTalentAssessmentsByUUID(context.Context, string) (*entity.TalentAssessment, error)
	FindTalentAssessmentsByUUIDForUpdate(context.Context, string) (*entity.TalentAssessment, error)
	DeleteTalentAssessmentsByUUID(context.Context, string) error
	GetAllTalentAssessments(context.Context, int, int) ([]entity.TalentAssessment, error)
	CountTalentAssessments(context.Context) (int64, error)
	SelectTalentAssessmentsWhere(context.Context, []Filter, int, int) ([]entity.TalentAssessment, error)
	CountTalentAssessmentsWhere(context.Context, []Filter) (int64, error)
	FindTalentAssessmentsByWorkerID(context.Context, string, int, int) ([]entity.TalentAssessment, error)
	FindTalentAssessmentsByTalentReviewCycleID(context.Context, string, int, int) ([]entity.TalentAssessment, error)
	FindTalentAssessmentsByFrameworkID(context.Context, string, int, int) ([]entity.TalentAssessment, error)
	FindTalentAssessmentsByPerformanceReviewID(context.Context, *string, int, int) ([]entity.TalentAssessment, error)
	FindTalentAssessmentsByStatus(context.Context, string, int, int) ([]entity.TalentAssessment, error)
	FindTalentAssessmentsByAssessedByPrincipalID(context.Context, string, int, int) ([]entity.TalentAssessment, error)
	FindTalentAssessmentsByEvidenceDocumentVersionReferenceID(context.Context, string, int, int) ([]entity.TalentAssessment, error)
	FindTalentAssessmentsBySegmentCode(context.Context, *string, int, int) ([]entity.TalentAssessment, error)
	FindTalentAssessmentsByFinalizedByPrincipalID(context.Context, *string, int, int) ([]entity.TalentAssessment, error)
	FindByWorkerID(ctx context.Context, value string, limit, offset int) ([]entity.TalentAssessment, error)
	FindByTalentReviewCycleID(ctx context.Context, value string, limit, offset int) ([]entity.TalentAssessment, error)
	FindByFrameworkID(ctx context.Context, value string, limit, offset int) ([]entity.TalentAssessment, error)
	FindByPerformanceReviewID(ctx context.Context, value *string, limit, offset int) ([]entity.TalentAssessment, error)
	FindByStatus(ctx context.Context, value string, limit, offset int) ([]entity.TalentAssessment, error)
	FindByAssessedByPrincipalID(ctx context.Context, value string, limit, offset int) ([]entity.TalentAssessment, error)
	FindByEvidenceDocumentVersionReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.TalentAssessment, error)
	FindBySegmentCode(ctx context.Context, value *string, limit, offset int) ([]entity.TalentAssessment, error)
	FindByFinalizedByPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.TalentAssessment, error)
}

// TalentAssessmentScoreRepository adalah port persistensi hr_talent_assessment_scores.
type TalentAssessmentScoreRepository interface {
	Repository[entity.TalentAssessmentScore, string]
	InsertTalentAssessmentScores(context.Context, *entity.TalentAssessmentScore) error
	UpdateTalentAssessmentScores(context.Context, *entity.TalentAssessmentScore) error
	FindTalentAssessmentScoresByUUID(context.Context, string) (*entity.TalentAssessmentScore, error)
	FindTalentAssessmentScoresByUUIDForUpdate(context.Context, string) (*entity.TalentAssessmentScore, error)
	DeleteTalentAssessmentScoresByUUID(context.Context, string) error
	GetAllTalentAssessmentScores(context.Context, int, int) ([]entity.TalentAssessmentScore, error)
	CountTalentAssessmentScores(context.Context) (int64, error)
	SelectTalentAssessmentScoresWhere(context.Context, []Filter, int, int) ([]entity.TalentAssessmentScore, error)
	CountTalentAssessmentScoresWhere(context.Context, []Filter) (int64, error)
	FindTalentAssessmentScoresByAssessmentID(context.Context, string, int, int) ([]entity.TalentAssessmentScore, error)
	FindTalentAssessmentScoresByFrameworkID(context.Context, string, int, int) ([]entity.TalentAssessmentScore, error)
	FindTalentAssessmentScoresByDimensionID(context.Context, string, int, int) ([]entity.TalentAssessmentScore, error)
	FindTalentAssessmentScoresByEvidenceDocumentVersionReferenceID(context.Context, string, int, int) ([]entity.TalentAssessmentScore, error)
	FindByAssessmentID(ctx context.Context, value string, limit, offset int) ([]entity.TalentAssessmentScore, error)
	FindByFrameworkID(ctx context.Context, value string, limit, offset int) ([]entity.TalentAssessmentScore, error)
	FindByDimensionID(ctx context.Context, value string, limit, offset int) ([]entity.TalentAssessmentScore, error)
	FindByEvidenceDocumentVersionReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.TalentAssessmentScore, error)
}

// TalentCalibrationSessionRepository adalah port persistensi hr_talent_calibration_sessions.
type TalentCalibrationSessionRepository interface {
	Repository[entity.TalentCalibrationSession, string]
	InsertTalentCalibrationSessions(context.Context, *entity.TalentCalibrationSession) error
	UpdateTalentCalibrationSessions(context.Context, *entity.TalentCalibrationSession) error
	FindTalentCalibrationSessionsByUUID(context.Context, string) (*entity.TalentCalibrationSession, error)
	FindTalentCalibrationSessionsByUUIDForUpdate(context.Context, string) (*entity.TalentCalibrationSession, error)
	DeleteTalentCalibrationSessionsByUUID(context.Context, string) error
	GetAllTalentCalibrationSessions(context.Context, int, int) ([]entity.TalentCalibrationSession, error)
	CountTalentCalibrationSessions(context.Context) (int64, error)
	SelectTalentCalibrationSessionsWhere(context.Context, []Filter, int, int) ([]entity.TalentCalibrationSession, error)
	CountTalentCalibrationSessionsWhere(context.Context, []Filter) (int64, error)
	FindTalentCalibrationSessionsByTalentReviewCycleID(context.Context, string, int, int) ([]entity.TalentCalibrationSession, error)
	FindTalentCalibrationSessionsBySessionCode(context.Context, string, int, int) ([]entity.TalentCalibrationSession, error)
	FindTalentCalibrationSessionsByChairPrincipalID(context.Context, string, int, int) ([]entity.TalentCalibrationSession, error)
	FindTalentCalibrationSessionsByStatus(context.Context, string, int, int) ([]entity.TalentCalibrationSession, error)
	FindTalentCalibrationSessionsByMinutesDocumentVersionReferenceID(context.Context, *string, int, int) ([]entity.TalentCalibrationSession, error)
	FindByTalentReviewCycleID(ctx context.Context, value string, limit, offset int) ([]entity.TalentCalibrationSession, error)
	FindBySessionCode(ctx context.Context, value string, limit, offset int) ([]entity.TalentCalibrationSession, error)
	FindByChairPrincipalID(ctx context.Context, value string, limit, offset int) ([]entity.TalentCalibrationSession, error)
	FindByStatus(ctx context.Context, value string, limit, offset int) ([]entity.TalentCalibrationSession, error)
	FindByMinutesDocumentVersionReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.TalentCalibrationSession, error)
}

// TalentCalibrationDecisionRepository adalah port persistensi hr_talent_calibration_decisions.
type TalentCalibrationDecisionRepository interface {
	Repository[entity.TalentCalibrationDecision, string]
	InsertTalentCalibrationDecisions(context.Context, *entity.TalentCalibrationDecision) error
	UpdateTalentCalibrationDecisions(context.Context, *entity.TalentCalibrationDecision) error
	FindTalentCalibrationDecisionsByUUID(context.Context, string) (*entity.TalentCalibrationDecision, error)
	FindTalentCalibrationDecisionsByUUIDForUpdate(context.Context, string) (*entity.TalentCalibrationDecision, error)
	DeleteTalentCalibrationDecisionsByUUID(context.Context, string) error
	GetAllTalentCalibrationDecisions(context.Context, int, int) ([]entity.TalentCalibrationDecision, error)
	CountTalentCalibrationDecisions(context.Context) (int64, error)
	SelectTalentCalibrationDecisionsWhere(context.Context, []Filter, int, int) ([]entity.TalentCalibrationDecision, error)
	CountTalentCalibrationDecisionsWhere(context.Context, []Filter) (int64, error)
	FindTalentCalibrationDecisionsBySessionID(context.Context, string, int, int) ([]entity.TalentCalibrationDecision, error)
	FindTalentCalibrationDecisionsByTalentReviewCycleID(context.Context, string, int, int) ([]entity.TalentCalibrationDecision, error)
	FindTalentCalibrationDecisionsByAssessmentID(context.Context, string, int, int) ([]entity.TalentCalibrationDecision, error)
	FindTalentCalibrationDecisionsByDimensionID(context.Context, string, int, int) ([]entity.TalentCalibrationDecision, error)
	FindTalentCalibrationDecisionsByRationaleDocumentVersionReferenceID(context.Context, string, int, int) ([]entity.TalentCalibrationDecision, error)
	FindTalentCalibrationDecisionsByDecidedByPrincipalID(context.Context, string, int, int) ([]entity.TalentCalibrationDecision, error)
	FindBySessionID(ctx context.Context, value string, limit, offset int) ([]entity.TalentCalibrationDecision, error)
	FindByTalentReviewCycleID(ctx context.Context, value string, limit, offset int) ([]entity.TalentCalibrationDecision, error)
	FindByAssessmentID(ctx context.Context, value string, limit, offset int) ([]entity.TalentCalibrationDecision, error)
	FindByDimensionID(ctx context.Context, value string, limit, offset int) ([]entity.TalentCalibrationDecision, error)
	FindByRationaleDocumentVersionReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.TalentCalibrationDecision, error)
	FindByDecidedByPrincipalID(ctx context.Context, value string, limit, offset int) ([]entity.TalentCalibrationDecision, error)
}

// TalentPoolRepository adalah port persistensi hr_talent_pools.
type TalentPoolRepository interface {
	Repository[entity.TalentPool, string]
	InsertTalentPools(context.Context, *entity.TalentPool) error
	UpdateTalentPools(context.Context, *entity.TalentPool) error
	FindTalentPoolsByUUID(context.Context, string) (*entity.TalentPool, error)
	FindTalentPoolsByUUIDForUpdate(context.Context, string) (*entity.TalentPool, error)
	DeleteTalentPoolsByUUID(context.Context, string) error
	GetAllTalentPools(context.Context, int, int) ([]entity.TalentPool, error)
	CountTalentPools(context.Context) (int64, error)
	SelectTalentPoolsWhere(context.Context, []Filter, int, int) ([]entity.TalentPool, error)
	CountTalentPoolsWhere(context.Context, []Filter) (int64, error)
	FindTalentPoolsByPoolCode(context.Context, string, int, int) ([]entity.TalentPool, error)
	FindTalentPoolsByLegalEntityReferenceID(context.Context, string, int, int) ([]entity.TalentPool, error)
	FindTalentPoolsByTargetJobID(context.Context, *string, int, int) ([]entity.TalentPool, error)
	FindTalentPoolsByEligibilityDocumentVersionReferenceID(context.Context, string, int, int) ([]entity.TalentPool, error)
	FindTalentPoolsByOwnerPrincipalID(context.Context, string, int, int) ([]entity.TalentPool, error)
	FindByPoolCode(ctx context.Context, value string, limit, offset int) ([]entity.TalentPool, error)
	FindByLegalEntityReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.TalentPool, error)
	FindByTargetJobID(ctx context.Context, value *string, limit, offset int) ([]entity.TalentPool, error)
	FindByEligibilityDocumentVersionReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.TalentPool, error)
	FindByOwnerPrincipalID(ctx context.Context, value string, limit, offset int) ([]entity.TalentPool, error)
}

// TalentPoolMembershipRepository adalah port persistensi hr_talent_pool_memberships.
type TalentPoolMembershipRepository interface {
	Repository[entity.TalentPoolMembership, string]
	InsertTalentPoolMemberships(context.Context, *entity.TalentPoolMembership) error
	UpdateTalentPoolMemberships(context.Context, *entity.TalentPoolMembership) error
	FindTalentPoolMembershipsByUUID(context.Context, string) (*entity.TalentPoolMembership, error)
	FindTalentPoolMembershipsByUUIDForUpdate(context.Context, string) (*entity.TalentPoolMembership, error)
	DeleteTalentPoolMembershipsByUUID(context.Context, string) error
	GetAllTalentPoolMemberships(context.Context, int, int) ([]entity.TalentPoolMembership, error)
	CountTalentPoolMemberships(context.Context) (int64, error)
	SelectTalentPoolMembershipsWhere(context.Context, []Filter, int, int) ([]entity.TalentPoolMembership, error)
	CountTalentPoolMembershipsWhere(context.Context, []Filter) (int64, error)
	FindTalentPoolMembershipsByTalentPoolID(context.Context, string, int, int) ([]entity.TalentPoolMembership, error)
	FindTalentPoolMembershipsByWorkerID(context.Context, string, int, int) ([]entity.TalentPoolMembership, error)
	FindTalentPoolMembershipsByAssessmentID(context.Context, string, int, int) ([]entity.TalentPoolMembership, error)
	FindTalentPoolMembershipsByNominatedByPrincipalID(context.Context, string, int, int) ([]entity.TalentPoolMembership, error)
	FindTalentPoolMembershipsByStatus(context.Context, string, int, int) ([]entity.TalentPoolMembership, error)
	FindTalentPoolMembershipsByApprovedByPrincipalID(context.Context, *string, int, int) ([]entity.TalentPoolMembership, error)
	FindTalentPoolMembershipsByExitReasonCode(context.Context, *string, int, int) ([]entity.TalentPoolMembership, error)
	FindByTalentPoolID(ctx context.Context, value string, limit, offset int) ([]entity.TalentPoolMembership, error)
	FindByWorkerID(ctx context.Context, value string, limit, offset int) ([]entity.TalentPoolMembership, error)
	FindByAssessmentID(ctx context.Context, value string, limit, offset int) ([]entity.TalentPoolMembership, error)
	FindByNominatedByPrincipalID(ctx context.Context, value string, limit, offset int) ([]entity.TalentPoolMembership, error)
	FindByStatus(ctx context.Context, value string, limit, offset int) ([]entity.TalentPoolMembership, error)
	FindByApprovedByPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.TalentPoolMembership, error)
	FindByExitReasonCode(ctx context.Context, value *string, limit, offset int) ([]entity.TalentPoolMembership, error)
}
