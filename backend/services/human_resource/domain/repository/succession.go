package repository

import (
	"context"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
)

// SuccessionPlanRepository adalah port persistensi hr_succession_plans.
type SuccessionPlanRepository interface {
	Repository[entity.SuccessionPlan, string]
	InsertSuccessionPlans(context.Context, *entity.SuccessionPlan) error
	UpdateSuccessionPlans(context.Context, *entity.SuccessionPlan) error
	FindSuccessionPlansByUUID(context.Context, string) (*entity.SuccessionPlan, error)
	FindSuccessionPlansByUUIDForUpdate(context.Context, string) (*entity.SuccessionPlan, error)
	DeleteSuccessionPlansByUUID(context.Context, string) error
	GetAllSuccessionPlans(context.Context, int, int) ([]entity.SuccessionPlan, error)
	CountSuccessionPlans(context.Context) (int64, error)
	SelectSuccessionPlansWhere(context.Context, []Filter, int, int) ([]entity.SuccessionPlan, error)
	CountSuccessionPlansWhere(context.Context, []Filter) (int64, error)
	FindSuccessionPlansByPositionID(context.Context, string, int, int) ([]entity.SuccessionPlan, error)
	FindSuccessionPlansByPlanCode(context.Context, string, int, int) ([]entity.SuccessionPlan, error)
	FindSuccessionPlansByApprovedByPrincipalID(context.Context, *string, int, int) ([]entity.SuccessionPlan, error)
	FindByPositionID(ctx context.Context, value string, limit, offset int) ([]entity.SuccessionPlan, error)
	FindByPlanCode(ctx context.Context, value string, limit, offset int) ([]entity.SuccessionPlan, error)
	FindByApprovedByPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.SuccessionPlan, error)
}

// SuccessionCandidateRepository adalah port persistensi hr_succession_candidates.
type SuccessionCandidateRepository interface {
	Repository[entity.SuccessionCandidate, string]
	InsertSuccessionCandidates(context.Context, *entity.SuccessionCandidate) error
	UpdateSuccessionCandidates(context.Context, *entity.SuccessionCandidate) error
	FindSuccessionCandidatesByUUID(context.Context, string) (*entity.SuccessionCandidate, error)
	FindSuccessionCandidatesByUUIDForUpdate(context.Context, string) (*entity.SuccessionCandidate, error)
	DeleteSuccessionCandidatesByUUID(context.Context, string) error
	GetAllSuccessionCandidates(context.Context, int, int) ([]entity.SuccessionCandidate, error)
	CountSuccessionCandidates(context.Context) (int64, error)
	SelectSuccessionCandidatesWhere(context.Context, []Filter, int, int) ([]entity.SuccessionCandidate, error)
	CountSuccessionCandidatesWhere(context.Context, []Filter) (int64, error)
	FindSuccessionCandidatesBySuccessionPlanID(context.Context, string, int, int) ([]entity.SuccessionCandidate, error)
	FindSuccessionCandidatesByCandidateWorkerID(context.Context, string, int, int) ([]entity.SuccessionCandidate, error)
	FindSuccessionCandidatesByReadinessCode(context.Context, string, int, int) ([]entity.SuccessionCandidate, error)
	FindSuccessionCandidatesByTalentAssessmentID(context.Context, *string, int, int) ([]entity.SuccessionCandidate, error)
	FindSuccessionCandidatesByDevelopmentPlanID(context.Context, *string, int, int) ([]entity.SuccessionCandidate, error)
	FindBySuccessionPlanID(ctx context.Context, value string, limit, offset int) ([]entity.SuccessionCandidate, error)
	FindByCandidateWorkerID(ctx context.Context, value string, limit, offset int) ([]entity.SuccessionCandidate, error)
	FindByReadinessCode(ctx context.Context, value string, limit, offset int) ([]entity.SuccessionCandidate, error)
	FindByTalentAssessmentID(ctx context.Context, value *string, limit, offset int) ([]entity.SuccessionCandidate, error)
	FindByDevelopmentPlanID(ctx context.Context, value *string, limit, offset int) ([]entity.SuccessionCandidate, error)
}
