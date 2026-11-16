// Sumber model adalah baseline SQL HRM; aturan bisnis tetap milik service.
package repository

import (
	"context"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
)

// OffboardingPlanRepository adalah port persistensi hr_offboarding_plans.
type OffboardingPlanRepository interface {
	Repository[entity.OffboardingPlan, string]
	InsertOffboardingPlans(context.Context, *entity.OffboardingPlan) error
	UpdateOffboardingPlans(context.Context, *entity.OffboardingPlan) error
	DeleteOffboardingPlansByUUID(context.Context, string) error
	FindOffboardingPlansByUUID(context.Context, string) (*entity.OffboardingPlan, error)
	FindOffboardingPlansByUUIDForUpdate(context.Context, string) (*entity.OffboardingPlan, error)
	GetAllOffboardingPlans(context.Context, int, int) ([]entity.OffboardingPlan, error)
	CountOffboardingPlans(context.Context) (int64, error)
	SelectOffboardingPlansWhere(context.Context, []Filter, int, int) ([]entity.OffboardingPlan, error)
	CountOffboardingPlansWhere(context.Context, []Filter) (int64, error)
	FindOffboardingPlansByWorkerID(context.Context, string, int, int) ([]entity.OffboardingPlan, error)
	FindOffboardingPlansByPlanNumber(context.Context, string, int, int) ([]entity.OffboardingPlan, error)
	FindOffboardingPlansByReasonCode(context.Context, string, int, int) ([]entity.OffboardingPlan, error)
	FindOffboardingPlansByStatus(context.Context, string, int, int) ([]entity.OffboardingPlan, error)
	FindOffboardingPlansByExitInterviewDocumentVersionReferenceID(context.Context, *string, int, int) ([]entity.OffboardingPlan, error)
	FindOffboardingPlansByFinalPayrollRunID(context.Context, *string, int, int) ([]entity.OffboardingPlan, error)
	FindOffboardingPlansByFinanceClearanceReferenceID(context.Context, *string, int, int) ([]entity.OffboardingPlan, error)
	FindOffboardingPlansByApprovedByPrincipalID(context.Context, *string, int, int) ([]entity.OffboardingPlan, error)
	FindOffboardingPlansByCreatedByPrincipalID(context.Context, string, int, int) ([]entity.OffboardingPlan, error)
	FindByWorkerID(ctx context.Context, value string, limit, offset int) ([]entity.OffboardingPlan, error)
	FindByPlanNumber(ctx context.Context, value string, limit, offset int) ([]entity.OffboardingPlan, error)
	FindByReasonCode(ctx context.Context, value string, limit, offset int) ([]entity.OffboardingPlan, error)
	FindByStatus(ctx context.Context, value string, limit, offset int) ([]entity.OffboardingPlan, error)
	FindByExitInterviewDocumentVersionReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.OffboardingPlan, error)
	FindByFinalPayrollRunID(ctx context.Context, value *string, limit, offset int) ([]entity.OffboardingPlan, error)
	FindByFinanceClearanceReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.OffboardingPlan, error)
	FindByApprovedByPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.OffboardingPlan, error)
	FindByCreatedByPrincipalID(ctx context.Context, value string, limit, offset int) ([]entity.OffboardingPlan, error)
}

// OffboardingTaskRepository adalah port persistensi hr_offboarding_tasks.
type OffboardingTaskRepository interface {
	Repository[entity.OffboardingTask, string]
	InsertOffboardingTasks(context.Context, *entity.OffboardingTask) error
	UpdateOffboardingTasks(context.Context, *entity.OffboardingTask) error
	DeleteOffboardingTasksByUUID(context.Context, string) error
	FindOffboardingTasksByUUID(context.Context, string) (*entity.OffboardingTask, error)
	FindOffboardingTasksByUUIDForUpdate(context.Context, string) (*entity.OffboardingTask, error)
	GetAllOffboardingTasks(context.Context, int, int) ([]entity.OffboardingTask, error)
	CountOffboardingTasks(context.Context) (int64, error)
	SelectOffboardingTasksWhere(context.Context, []Filter, int, int) ([]entity.OffboardingTask, error)
	CountOffboardingTasksWhere(context.Context, []Filter) (int64, error)
	FindOffboardingTasksByOffboardingPlanID(context.Context, string, int, int) ([]entity.OffboardingTask, error)
	FindOffboardingTasksByTaskCode(context.Context, string, int, int) ([]entity.OffboardingTask, error)
	FindOffboardingTasksByAssignedPrincipalID(context.Context, *string, int, int) ([]entity.OffboardingTask, error)
	FindOffboardingTasksByStatus(context.Context, string, int, int) ([]entity.OffboardingTask, error)
	FindOffboardingTasksByOwnerConfirmationReferenceID(context.Context, *string, int, int) ([]entity.OffboardingTask, error)
	FindOffboardingTasksByWaivedByPrincipalID(context.Context, *string, int, int) ([]entity.OffboardingTask, error)
	FindByOffboardingPlanID(ctx context.Context, value string, limit, offset int) ([]entity.OffboardingTask, error)
	FindByTaskCode(ctx context.Context, value string, limit, offset int) ([]entity.OffboardingTask, error)
	FindByAssignedPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.OffboardingTask, error)
	FindByStatus(ctx context.Context, value string, limit, offset int) ([]entity.OffboardingTask, error)
	FindByOwnerConfirmationReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.OffboardingTask, error)
	FindByWaivedByPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.OffboardingTask, error)
}
