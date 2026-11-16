// Sumber model adalah baseline SQL HRM; aturan bisnis tetap milik service.
package repository

import (
	"context"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
)

// OnboardingPlanRepository adalah port persistensi hr_onboarding_plans.
type OnboardingPlanRepository interface {
	Repository[entity.OnboardingPlan, string]
	InsertOnboardingPlans(context.Context, *entity.OnboardingPlan) error
	UpdateOnboardingPlans(context.Context, *entity.OnboardingPlan) error
	DeleteOnboardingPlansByUUID(context.Context, string) error
	FindOnboardingPlansByUUID(context.Context, string) (*entity.OnboardingPlan, error)
	FindOnboardingPlansByUUIDForUpdate(context.Context, string) (*entity.OnboardingPlan, error)
	GetAllOnboardingPlans(context.Context, int, int) ([]entity.OnboardingPlan, error)
	CountOnboardingPlans(context.Context) (int64, error)
	SelectOnboardingPlansWhere(context.Context, []Filter, int, int) ([]entity.OnboardingPlan, error)
	CountOnboardingPlansWhere(context.Context, []Filter) (int64, error)
	FindOnboardingPlansByWorkerID(context.Context, string, int, int) ([]entity.OnboardingPlan, error)
	FindOnboardingPlansByPlanCode(context.Context, string, int, int) ([]entity.OnboardingPlan, error)
	FindOnboardingPlansByStatus(context.Context, string, int, int) ([]entity.OnboardingPlan, error)
	FindOnboardingPlansByTemplateReferenceID(context.Context, *string, int, int) ([]entity.OnboardingPlan, error)
	FindOnboardingPlansByCreatedByPrincipalID(context.Context, *string, int, int) ([]entity.OnboardingPlan, error)
	FindByWorkerID(ctx context.Context, value string, limit, offset int) ([]entity.OnboardingPlan, error)
	FindByPlanCode(ctx context.Context, value string, limit, offset int) ([]entity.OnboardingPlan, error)
	FindByStatus(ctx context.Context, value string, limit, offset int) ([]entity.OnboardingPlan, error)
	FindByTemplateReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.OnboardingPlan, error)
	FindByCreatedByPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.OnboardingPlan, error)
}

// OnboardingTaskRepository adalah port persistensi hr_onboarding_tasks.
type OnboardingTaskRepository interface {
	Repository[entity.OnboardingTask, string]
	InsertOnboardingTasks(context.Context, *entity.OnboardingTask) error
	UpdateOnboardingTasks(context.Context, *entity.OnboardingTask) error
	DeleteOnboardingTasksByUUID(context.Context, string) error
	FindOnboardingTasksByUUID(context.Context, string) (*entity.OnboardingTask, error)
	FindOnboardingTasksByUUIDForUpdate(context.Context, string) (*entity.OnboardingTask, error)
	GetAllOnboardingTasks(context.Context, int, int) ([]entity.OnboardingTask, error)
	CountOnboardingTasks(context.Context) (int64, error)
	SelectOnboardingTasksWhere(context.Context, []Filter, int, int) ([]entity.OnboardingTask, error)
	CountOnboardingTasksWhere(context.Context, []Filter) (int64, error)
	FindOnboardingTasksByOnboardingPlanID(context.Context, string, int, int) ([]entity.OnboardingTask, error)
	FindOnboardingTasksByTaskCode(context.Context, string, int, int) ([]entity.OnboardingTask, error)
	FindOnboardingTasksByTaskOwnerType(context.Context, string, int, int) ([]entity.OnboardingTask, error)
	FindOnboardingTasksByAssignedPrincipalID(context.Context, *string, int, int) ([]entity.OnboardingTask, error)
	FindOnboardingTasksByCompletionEvidenceReferenceID(context.Context, *string, int, int) ([]entity.OnboardingTask, error)
	FindByOnboardingPlanID(ctx context.Context, value string, limit, offset int) ([]entity.OnboardingTask, error)
	FindByTaskCode(ctx context.Context, value string, limit, offset int) ([]entity.OnboardingTask, error)
	FindByTaskOwnerType(ctx context.Context, value string, limit, offset int) ([]entity.OnboardingTask, error)
	FindByAssignedPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.OnboardingTask, error)
	FindByCompletionEvidenceReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.OnboardingTask, error)
}
