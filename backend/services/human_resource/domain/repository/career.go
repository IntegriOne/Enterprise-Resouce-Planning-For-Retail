package repository

import (
	"context"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
)

// DevelopmentPlanRepository adalah port persistensi hr_development_plans.
type DevelopmentPlanRepository interface {
	Repository[entity.DevelopmentPlan, string]
	InsertDevelopmentPlans(context.Context, *entity.DevelopmentPlan) error
	UpdateDevelopmentPlans(context.Context, *entity.DevelopmentPlan) error
	FindDevelopmentPlansByUUID(context.Context, string) (*entity.DevelopmentPlan, error)
	FindDevelopmentPlansByUUIDForUpdate(context.Context, string) (*entity.DevelopmentPlan, error)
	DeleteDevelopmentPlansByUUID(context.Context, string) error
	GetAllDevelopmentPlans(context.Context, int, int) ([]entity.DevelopmentPlan, error)
	CountDevelopmentPlans(context.Context) (int64, error)
	SelectDevelopmentPlansWhere(context.Context, []Filter, int, int) ([]entity.DevelopmentPlan, error)
	CountDevelopmentPlansWhere(context.Context, []Filter) (int64, error)
	FindDevelopmentPlansByWorkerID(context.Context, string, int, int) ([]entity.DevelopmentPlan, error)
	FindDevelopmentPlansByPlanCode(context.Context, string, int, int) ([]entity.DevelopmentPlan, error)
	FindDevelopmentPlansByTargetJobID(context.Context, *string, int, int) ([]entity.DevelopmentPlan, error)
	FindDevelopmentPlansByStatus(context.Context, string, int, int) ([]entity.DevelopmentPlan, error)
	FindDevelopmentPlansByMentorWorkerID(context.Context, *string, int, int) ([]entity.DevelopmentPlan, error)
	FindDevelopmentPlansByApprovedByPrincipalID(context.Context, *string, int, int) ([]entity.DevelopmentPlan, error)
	FindByWorkerID(ctx context.Context, value string, limit, offset int) ([]entity.DevelopmentPlan, error)
	FindByPlanCode(ctx context.Context, value string, limit, offset int) ([]entity.DevelopmentPlan, error)
	FindByTargetJobID(ctx context.Context, value *string, limit, offset int) ([]entity.DevelopmentPlan, error)
	FindByStatus(ctx context.Context, value string, limit, offset int) ([]entity.DevelopmentPlan, error)
	FindByMentorWorkerID(ctx context.Context, value *string, limit, offset int) ([]entity.DevelopmentPlan, error)
	FindByApprovedByPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.DevelopmentPlan, error)
}

// DevelopmentActionRepository adalah port persistensi hr_development_actions.
type DevelopmentActionRepository interface {
	Repository[entity.DevelopmentAction, string]
	InsertDevelopmentActions(context.Context, *entity.DevelopmentAction) error
	UpdateDevelopmentActions(context.Context, *entity.DevelopmentAction) error
	FindDevelopmentActionsByUUID(context.Context, string) (*entity.DevelopmentAction, error)
	FindDevelopmentActionsByUUIDForUpdate(context.Context, string) (*entity.DevelopmentAction, error)
	DeleteDevelopmentActionsByUUID(context.Context, string) error
	GetAllDevelopmentActions(context.Context, int, int) ([]entity.DevelopmentAction, error)
	CountDevelopmentActions(context.Context) (int64, error)
	SelectDevelopmentActionsWhere(context.Context, []Filter, int, int) ([]entity.DevelopmentAction, error)
	CountDevelopmentActionsWhere(context.Context, []Filter) (int64, error)
	FindDevelopmentActionsByDevelopmentPlanID(context.Context, string, int, int) ([]entity.DevelopmentAction, error)
	FindDevelopmentActionsByActionNumber(context.Context, int32, int, int) ([]entity.DevelopmentAction, error)
	FindDevelopmentActionsByCompetencyID(context.Context, *string, int, int) ([]entity.DevelopmentAction, error)
	FindDevelopmentActionsByCourseID(context.Context, *string, int, int) ([]entity.DevelopmentAction, error)
	FindDevelopmentActionsByActionType(context.Context, string, int, int) ([]entity.DevelopmentAction, error)
	FindDevelopmentActionsByEvidenceDocumentVersionReferenceID(context.Context, *string, int, int) ([]entity.DevelopmentAction, error)
	FindByDevelopmentPlanID(ctx context.Context, value string, limit, offset int) ([]entity.DevelopmentAction, error)
	FindByActionNumber(ctx context.Context, value int32, limit, offset int) ([]entity.DevelopmentAction, error)
	FindByCompetencyID(ctx context.Context, value *string, limit, offset int) ([]entity.DevelopmentAction, error)
	FindByCourseID(ctx context.Context, value *string, limit, offset int) ([]entity.DevelopmentAction, error)
	FindByActionType(ctx context.Context, value string, limit, offset int) ([]entity.DevelopmentAction, error)
	FindByEvidenceDocumentVersionReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.DevelopmentAction, error)
}

// CareerPathRepository adalah port persistensi hr_career_paths.
type CareerPathRepository interface {
	Repository[entity.CareerPath, string]
	InsertCareerPaths(context.Context, *entity.CareerPath) error
	UpdateCareerPaths(context.Context, *entity.CareerPath) error
	FindCareerPathsByUUID(context.Context, string) (*entity.CareerPath, error)
	FindCareerPathsByUUIDForUpdate(context.Context, string) (*entity.CareerPath, error)
	DeleteCareerPathsByUUID(context.Context, string) error
	GetAllCareerPaths(context.Context, int, int) ([]entity.CareerPath, error)
	CountCareerPaths(context.Context) (int64, error)
	SelectCareerPathsWhere(context.Context, []Filter, int, int) ([]entity.CareerPath, error)
	CountCareerPathsWhere(context.Context, []Filter) (int64, error)
	FindCareerPathsByPathCode(context.Context, string, int, int) ([]entity.CareerPath, error)
	FindCareerPathsByJobFamilyID(context.Context, string, int, int) ([]entity.CareerPath, error)
	FindCareerPathsByApprovedByPrincipalID(context.Context, string, int, int) ([]entity.CareerPath, error)
	FindByPathCode(ctx context.Context, value string, limit, offset int) ([]entity.CareerPath, error)
	FindByJobFamilyID(ctx context.Context, value string, limit, offset int) ([]entity.CareerPath, error)
	FindByApprovedByPrincipalID(ctx context.Context, value string, limit, offset int) ([]entity.CareerPath, error)
}

// CareerPathTransitionRepository adalah port persistensi hr_career_path_transitions.
type CareerPathTransitionRepository interface {
	Repository[entity.CareerPathTransition, string]
	InsertCareerPathTransitions(context.Context, *entity.CareerPathTransition) error
	UpdateCareerPathTransitions(context.Context, *entity.CareerPathTransition) error
	FindCareerPathTransitionsByUUID(context.Context, string) (*entity.CareerPathTransition, error)
	FindCareerPathTransitionsByUUIDForUpdate(context.Context, string) (*entity.CareerPathTransition, error)
	DeleteCareerPathTransitionsByUUID(context.Context, string) error
	GetAllCareerPathTransitions(context.Context, int, int) ([]entity.CareerPathTransition, error)
	CountCareerPathTransitions(context.Context) (int64, error)
	SelectCareerPathTransitionsWhere(context.Context, []Filter, int, int) ([]entity.CareerPathTransition, error)
	CountCareerPathTransitionsWhere(context.Context, []Filter) (int64, error)
	FindCareerPathTransitionsByCareerPathID(context.Context, string, int, int) ([]entity.CareerPathTransition, error)
	FindCareerPathTransitionsByFromJobID(context.Context, string, int, int) ([]entity.CareerPathTransition, error)
	FindCareerPathTransitionsByToJobID(context.Context, string, int, int) ([]entity.CareerPathTransition, error)
	FindCareerPathTransitionsByMovementType(context.Context, string, int, int) ([]entity.CareerPathTransition, error)
	FindCareerPathTransitionsByRequirementsDocumentVersionReferenceID(context.Context, string, int, int) ([]entity.CareerPathTransition, error)
	FindByCareerPathID(ctx context.Context, value string, limit, offset int) ([]entity.CareerPathTransition, error)
	FindByFromJobID(ctx context.Context, value string, limit, offset int) ([]entity.CareerPathTransition, error)
	FindByToJobID(ctx context.Context, value string, limit, offset int) ([]entity.CareerPathTransition, error)
	FindByMovementType(ctx context.Context, value string, limit, offset int) ([]entity.CareerPathTransition, error)
	FindByRequirementsDocumentVersionReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.CareerPathTransition, error)
}

// WorkerCareerAspirationRepository adalah port persistensi hr_worker_career_aspirations.
type WorkerCareerAspirationRepository interface {
	Repository[entity.WorkerCareerAspiration, string]
	InsertWorkerCareerAspirations(context.Context, *entity.WorkerCareerAspiration) error
	UpdateWorkerCareerAspirations(context.Context, *entity.WorkerCareerAspiration) error
	FindWorkerCareerAspirationsByUUID(context.Context, string) (*entity.WorkerCareerAspiration, error)
	FindWorkerCareerAspirationsByUUIDForUpdate(context.Context, string) (*entity.WorkerCareerAspiration, error)
	DeleteWorkerCareerAspirationsByUUID(context.Context, string) error
	GetAllWorkerCareerAspirations(context.Context, int, int) ([]entity.WorkerCareerAspiration, error)
	CountWorkerCareerAspirations(context.Context) (int64, error)
	SelectWorkerCareerAspirationsWhere(context.Context, []Filter, int, int) ([]entity.WorkerCareerAspiration, error)
	CountWorkerCareerAspirationsWhere(context.Context, []Filter) (int64, error)
	FindWorkerCareerAspirationsByWorkerID(context.Context, string, int, int) ([]entity.WorkerCareerAspiration, error)
	FindWorkerCareerAspirationsByTargetJobID(context.Context, string, int, int) ([]entity.WorkerCareerAspiration, error)
	FindWorkerCareerAspirationsByCareerPathTransitionID(context.Context, *string, int, int) ([]entity.WorkerCareerAspiration, error)
	FindWorkerCareerAspirationsByRecordedByPrincipalID(context.Context, string, int, int) ([]entity.WorkerCareerAspiration, error)
	FindWorkerCareerAspirationsByPreferredLocationReferenceID(context.Context, *string, int, int) ([]entity.WorkerCareerAspiration, error)
	FindByWorkerID(ctx context.Context, value string, limit, offset int) ([]entity.WorkerCareerAspiration, error)
	FindByTargetJobID(ctx context.Context, value string, limit, offset int) ([]entity.WorkerCareerAspiration, error)
	FindByCareerPathTransitionID(ctx context.Context, value *string, limit, offset int) ([]entity.WorkerCareerAspiration, error)
	FindByRecordedByPrincipalID(ctx context.Context, value string, limit, offset int) ([]entity.WorkerCareerAspiration, error)
	FindByPreferredLocationReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.WorkerCareerAspiration, error)
}
