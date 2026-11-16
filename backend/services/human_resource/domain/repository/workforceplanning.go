// Sumber model adalah baseline SQL HRM; aturan bisnis tetap milik service.
package repository

import (
	"context"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
)

// JobFamilyRepository adalah port persistensi hr_job_families.
type JobFamilyRepository interface {
	Repository[entity.JobFamily, string]
	InsertJobFamilies(context.Context, *entity.JobFamily) error
	UpdateJobFamilies(context.Context, *entity.JobFamily) error
	DeleteJobFamiliesByUUID(context.Context, string) error
	FindJobFamiliesByUUID(context.Context, string) (*entity.JobFamily, error)
	FindJobFamiliesByUUIDForUpdate(context.Context, string) (*entity.JobFamily, error)
	GetAllJobFamilies(context.Context, int, int) ([]entity.JobFamily, error)
	CountJobFamilies(context.Context) (int64, error)
	SelectJobFamiliesWhere(context.Context, []Filter, int, int) ([]entity.JobFamily, error)
	CountJobFamiliesWhere(context.Context, []Filter) (int64, error)
	FindJobFamiliesByFamilyCode(context.Context, string, int, int) ([]entity.JobFamily, error)
	FindByFamilyCode(ctx context.Context, value string, limit, offset int) ([]entity.JobFamily, error)
}

// JobRepository adalah port persistensi hr_jobs.
type JobRepository interface {
	Repository[entity.Job, string]
	InsertJobs(context.Context, *entity.Job) error
	UpdateJobs(context.Context, *entity.Job) error
	DeleteJobsByUUID(context.Context, string) error
	FindJobsByUUID(context.Context, string) (*entity.Job, error)
	FindJobsByUUIDForUpdate(context.Context, string) (*entity.Job, error)
	GetAllJobs(context.Context, int, int) ([]entity.Job, error)
	CountJobs(context.Context) (int64, error)
	SelectJobsWhere(context.Context, []Filter, int, int) ([]entity.Job, error)
	CountJobsWhere(context.Context, []Filter) (int64, error)
	FindJobsByJobFamilyID(context.Context, string, int, int) ([]entity.Job, error)
	FindJobsByJobCode(context.Context, string, int, int) ([]entity.Job, error)
	FindJobsByJobLevelCode(context.Context, *string, int, int) ([]entity.Job, error)
	FindByJobFamilyID(ctx context.Context, value string, limit, offset int) ([]entity.Job, error)
	FindByJobCode(ctx context.Context, value string, limit, offset int) ([]entity.Job, error)
	FindByJobLevelCode(ctx context.Context, value *string, limit, offset int) ([]entity.Job, error)
}

// PositionRepository adalah port persistensi hr_positions.
type PositionRepository interface {
	Repository[entity.Position, string]
	InsertPositions(context.Context, *entity.Position) error
	UpdatePositions(context.Context, *entity.Position) error
	DeletePositionsByUUID(context.Context, string) error
	FindPositionsByUUID(context.Context, string) (*entity.Position, error)
	FindPositionsByUUIDForUpdate(context.Context, string) (*entity.Position, error)
	GetAllPositions(context.Context, int, int) ([]entity.Position, error)
	CountPositions(context.Context) (int64, error)
	SelectPositionsWhere(context.Context, []Filter, int, int) ([]entity.Position, error)
	CountPositionsWhere(context.Context, []Filter) (int64, error)
	FindPositionsByPositionCode(context.Context, string, int, int) ([]entity.Position, error)
	FindPositionsByJobID(context.Context, string, int, int) ([]entity.Position, error)
	FindPositionsByOrganizationUnitReferenceID(context.Context, string, int, int) ([]entity.Position, error)
	FindPositionsByLegalEntityReferenceID(context.Context, string, int, int) ([]entity.Position, error)
	FindPositionsByReportsToPositionID(context.Context, *string, int, int) ([]entity.Position, error)
	FindPositionsByPositionStatus(context.Context, string, int, int) ([]entity.Position, error)
	FindPositionsByWorkLocationReferenceID(context.Context, *string, int, int) ([]entity.Position, error)
	FindPositionsByCostCenterReferenceID(context.Context, *string, int, int) ([]entity.Position, error)
	FindByPositionCode(ctx context.Context, value string, limit, offset int) ([]entity.Position, error)
	FindByJobID(ctx context.Context, value string, limit, offset int) ([]entity.Position, error)
	FindByOrganizationUnitReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.Position, error)
	FindByLegalEntityReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.Position, error)
	FindByReportsToPositionID(ctx context.Context, value *string, limit, offset int) ([]entity.Position, error)
	FindByPositionStatus(ctx context.Context, value string, limit, offset int) ([]entity.Position, error)
	FindByWorkLocationReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.Position, error)
	FindByCostCenterReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.Position, error)
}

// WorkforcePlanRepository adalah port persistensi hr_workforce_plans.
type WorkforcePlanRepository interface {
	Repository[entity.WorkforcePlan, string]
	InsertWorkforcePlans(context.Context, *entity.WorkforcePlan) error
	UpdateWorkforcePlans(context.Context, *entity.WorkforcePlan) error
	DeleteWorkforcePlansByUUID(context.Context, string) error
	FindWorkforcePlansByUUID(context.Context, string) (*entity.WorkforcePlan, error)
	FindWorkforcePlansByUUIDForUpdate(context.Context, string) (*entity.WorkforcePlan, error)
	GetAllWorkforcePlans(context.Context, int, int) ([]entity.WorkforcePlan, error)
	CountWorkforcePlans(context.Context) (int64, error)
	SelectWorkforcePlansWhere(context.Context, []Filter, int, int) ([]entity.WorkforcePlan, error)
	CountWorkforcePlansWhere(context.Context, []Filter) (int64, error)
	FindWorkforcePlansByPlanCode(context.Context, string, int, int) ([]entity.WorkforcePlan, error)
	FindWorkforcePlansByLegalEntityReferenceID(context.Context, string, int, int) ([]entity.WorkforcePlan, error)
	FindWorkforcePlansByScenarioCode(context.Context, string, int, int) ([]entity.WorkforcePlan, error)
	FindWorkforcePlansByStatus(context.Context, string, int, int) ([]entity.WorkforcePlan, error)
	FindWorkforcePlansByFinanceBudgetReferenceID(context.Context, *string, int, int) ([]entity.WorkforcePlan, error)
	FindWorkforcePlansByCreatedByPrincipalID(context.Context, string, int, int) ([]entity.WorkforcePlan, error)
	FindWorkforcePlansByApprovedByPrincipalID(context.Context, *string, int, int) ([]entity.WorkforcePlan, error)
	FindByPlanCode(ctx context.Context, value string, limit, offset int) ([]entity.WorkforcePlan, error)
	FindByLegalEntityReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.WorkforcePlan, error)
	FindByScenarioCode(ctx context.Context, value string, limit, offset int) ([]entity.WorkforcePlan, error)
	FindByStatus(ctx context.Context, value string, limit, offset int) ([]entity.WorkforcePlan, error)
	FindByFinanceBudgetReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.WorkforcePlan, error)
	FindByCreatedByPrincipalID(ctx context.Context, value string, limit, offset int) ([]entity.WorkforcePlan, error)
	FindByApprovedByPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.WorkforcePlan, error)
}

// WorkforcePlanLineRepository adalah port persistensi hr_workforce_plan_lines.
type WorkforcePlanLineRepository interface {
	Repository[entity.WorkforcePlanLine, string]
	InsertWorkforcePlanLines(context.Context, *entity.WorkforcePlanLine) error
	UpdateWorkforcePlanLines(context.Context, *entity.WorkforcePlanLine) error
	DeleteWorkforcePlanLinesByUUID(context.Context, string) error
	FindWorkforcePlanLinesByUUID(context.Context, string) (*entity.WorkforcePlanLine, error)
	FindWorkforcePlanLinesByUUIDForUpdate(context.Context, string) (*entity.WorkforcePlanLine, error)
	GetAllWorkforcePlanLines(context.Context, int, int) ([]entity.WorkforcePlanLine, error)
	CountWorkforcePlanLines(context.Context) (int64, error)
	SelectWorkforcePlanLinesWhere(context.Context, []Filter, int, int) ([]entity.WorkforcePlanLine, error)
	CountWorkforcePlanLinesWhere(context.Context, []Filter) (int64, error)
	FindWorkforcePlanLinesByWorkforcePlanID(context.Context, string, int, int) ([]entity.WorkforcePlanLine, error)
	FindWorkforcePlanLinesByLineNumber(context.Context, int32, int, int) ([]entity.WorkforcePlanLine, error)
	FindWorkforcePlanLinesByOrganizationUnitReferenceID(context.Context, string, int, int) ([]entity.WorkforcePlanLine, error)
	FindWorkforcePlanLinesByJobID(context.Context, string, int, int) ([]entity.WorkforcePlanLine, error)
	FindWorkforcePlanLinesByWorkLocationReferenceID(context.Context, *string, int, int) ([]entity.WorkforcePlanLine, error)
	FindWorkforcePlanLinesByCurrencyCode(context.Context, *string, int, int) ([]entity.WorkforcePlanLine, error)
	FindWorkforcePlanLinesByActionCode(context.Context, string, int, int) ([]entity.WorkforcePlanLine, error)
	FindByWorkforcePlanID(ctx context.Context, value string, limit, offset int) ([]entity.WorkforcePlanLine, error)
	FindByLineNumber(ctx context.Context, value int32, limit, offset int) ([]entity.WorkforcePlanLine, error)
	FindByOrganizationUnitReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.WorkforcePlanLine, error)
	FindByJobID(ctx context.Context, value string, limit, offset int) ([]entity.WorkforcePlanLine, error)
	FindByWorkLocationReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.WorkforcePlanLine, error)
	FindByCurrencyCode(ctx context.Context, value *string, limit, offset int) ([]entity.WorkforcePlanLine, error)
	FindByActionCode(ctx context.Context, value string, limit, offset int) ([]entity.WorkforcePlanLine, error)
}
