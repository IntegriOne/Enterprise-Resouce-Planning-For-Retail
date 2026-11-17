// Kerangka awal dibuat dengan alat bantu; kode ini dipelihara langsung per domain.
// Aturan bisnis lintas record tetap dijalankan melalui policy/integrity dan command domain.
package service

import (
	"context"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

type Dependencies struct {
	Runtime           operation.Dependencies
	JobFamily         domainrepo.RecordRepository[entity.JobFamily, string]
	Job               domainrepo.RecordRepository[entity.Job, string]
	Position          domainrepo.RecordRepository[entity.Position, string]
	WorkforcePlan     domainrepo.RecordRepository[entity.WorkforcePlan, string]
	WorkforcePlanLine domainrepo.RecordRepository[entity.WorkforcePlanLine, string]
}

// Service mengoordinasikan operasi domain workforceplanning.
type Service struct {
	jobFamily         *operation.Resource[entity.JobFamily, string, *entity.JobFamily]
	job               *operation.Resource[entity.Job, string, *entity.Job]
	position          *operation.Resource[entity.Position, string, *entity.Position]
	workforcePlan     *operation.Resource[entity.WorkforcePlan, string, *entity.WorkforcePlan]
	workforcePlanLine *operation.Resource[entity.WorkforcePlanLine, string, *entity.WorkforcePlanLine]
}

// NewService memeriksa seluruh port dan dependency sebelum service dapat dipakai.
func NewService(deps Dependencies) (*Service, error) {
	s := &Service{}
	var err error
	s.jobFamily, err = operation.NewResource[entity.JobFamily, string, *entity.JobFamily](
		deps.JobFamily,
		deps.Runtime,
		specJobFamily(),
	)
	if err != nil {
		return nil, err
	}
	s.job, err = operation.NewResource[entity.Job, string, *entity.Job](
		deps.Job,
		deps.Runtime,
		specJob(),
	)
	if err != nil {
		return nil, err
	}
	s.position, err = operation.NewResource[entity.Position, string, *entity.Position](
		deps.Position,
		deps.Runtime,
		specPosition(),
	)
	if err != nil {
		return nil, err
	}
	s.workforcePlan, err = operation.NewResource[entity.WorkforcePlan, string, *entity.WorkforcePlan](
		deps.WorkforcePlan,
		deps.Runtime,
		specWorkforcePlan(),
	)
	if err != nil {
		return nil, err
	}
	s.workforcePlanLine, err = operation.NewResource[entity.WorkforcePlanLine, string, *entity.WorkforcePlanLine](
		deps.WorkforcePlanLine,
		deps.Runtime,
		specWorkforcePlanLine(),
	)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// CreateJobFamily menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreateJobFamily(ctx context.Context, in CreateJobFamilyInput) (*entity.JobFamily, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.jobFamily.Create(ctx, func() (*entity.JobFamily, error) { return createJobFamily(in) })
}

// UpdateJobFamily menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdateJobFamily(ctx context.Context, key string, in UpdateJobFamilyInput) (*entity.JobFamily, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.jobFamily.Update(ctx, key, func(value *entity.JobFamily) error { return patchJobFamily(value, in) })
}

// DeleteJobFamily menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeleteJobFamily(ctx context.Context, key string) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.jobFamily.Delete(ctx, key)
}

// GetJobFamilyByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetJobFamilyByID(ctx context.Context, key string) (*entity.JobFamily, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.jobFamily.Get(ctx, key)
}

// GetAllJobFamilies menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllJobFamilies(ctx context.Context, limit, offset int) (operation.Page[entity.JobFamily], error) {
	if s == nil {
		return operation.Page[entity.JobFamily]{}, hrerrors.ErrUnavailable
	}
	return s.jobFamily.List(ctx, nil, limit, offset)
}

// FindJobFamilyByFamilyCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindJobFamilyByFamilyCode(ctx context.Context, value string, limit, offset int) (operation.Page[entity.JobFamily], error) {
	if s == nil {
		return operation.Page[entity.JobFamily]{}, hrerrors.ErrUnavailable
	}
	return s.jobFamily.List(ctx, []domainrepo.Filter{{Column: "family_code", Values: []any{value}}}, limit, offset)
}

// CreateJob menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreateJob(ctx context.Context, in CreateJobInput) (*entity.Job, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.job.Create(ctx, func() (*entity.Job, error) { return createJob(in) })
}

// UpdateJob menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdateJob(ctx context.Context, key string, in UpdateJobInput) (*entity.Job, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.job.Update(ctx, key, func(value *entity.Job) error { return patchJob(value, in) })
}

// DeleteJob menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeleteJob(ctx context.Context, key string) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.job.Delete(ctx, key)
}

// GetJobByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetJobByID(ctx context.Context, key string) (*entity.Job, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.job.Get(ctx, key)
}

// GetAllJobs menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllJobs(ctx context.Context, limit, offset int) (operation.Page[entity.Job], error) {
	if s == nil {
		return operation.Page[entity.Job]{}, hrerrors.ErrUnavailable
	}
	return s.job.List(ctx, nil, limit, offset)
}

// FindJobByJobFamilyID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindJobByJobFamilyID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.Job], error) {
	if s == nil {
		return operation.Page[entity.Job]{}, hrerrors.ErrUnavailable
	}
	return s.job.List(ctx, []domainrepo.Filter{{Column: "job_family_id", Values: []any{value}}}, limit, offset)
}

// FindJobByJobCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindJobByJobCode(ctx context.Context, value string, limit, offset int) (operation.Page[entity.Job], error) {
	if s == nil {
		return operation.Page[entity.Job]{}, hrerrors.ErrUnavailable
	}
	return s.job.List(ctx, []domainrepo.Filter{{Column: "job_code", Values: []any{value}}}, limit, offset)
}

// FindJobByJobLevelCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindJobByJobLevelCode(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.Job], error) {
	if s == nil {
		return operation.Page[entity.Job]{}, hrerrors.ErrUnavailable
	}
	return s.job.List(ctx, []domainrepo.Filter{{Column: "job_level_code", Values: []any{value}}}, limit, offset)
}

// CreatePosition menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreatePosition(ctx context.Context, in CreatePositionInput) (*entity.Position, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.position.Create(ctx, func() (*entity.Position, error) { return createPosition(in) })
}

// UpdatePosition menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdatePosition(ctx context.Context, key string, in UpdatePositionInput) (*entity.Position, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	if in.PositionStatus.Set && (in.PositionStatus.Value == nil || *in.PositionStatus.Value != "planned") {
		return nil, hrerrors.ErrWorkflowRequired
	}
	return s.position.Update(ctx, key, func(value *entity.Position) error { return patchPosition(value, in) })
}

// DeletePosition menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeletePosition(ctx context.Context, key string) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.position.Delete(ctx, key)
}

// GetPositionByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetPositionByID(ctx context.Context, key string) (*entity.Position, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.position.Get(ctx, key)
}

// GetAllPositions menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllPositions(ctx context.Context, limit, offset int) (operation.Page[entity.Position], error) {
	if s == nil {
		return operation.Page[entity.Position]{}, hrerrors.ErrUnavailable
	}
	return s.position.List(ctx, nil, limit, offset)
}

// FindPositionByPositionCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPositionByPositionCode(ctx context.Context, value string, limit, offset int) (operation.Page[entity.Position], error) {
	if s == nil {
		return operation.Page[entity.Position]{}, hrerrors.ErrUnavailable
	}
	return s.position.List(ctx, []domainrepo.Filter{{Column: "position_code", Values: []any{value}}}, limit, offset)
}

// FindPositionByJobID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPositionByJobID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.Position], error) {
	if s == nil {
		return operation.Page[entity.Position]{}, hrerrors.ErrUnavailable
	}
	return s.position.List(ctx, []domainrepo.Filter{{Column: "job_id", Values: []any{value}}}, limit, offset)
}

// FindPositionByOrganizationUnitReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPositionByOrganizationUnitReferenceID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.Position], error) {
	if s == nil {
		return operation.Page[entity.Position]{}, hrerrors.ErrUnavailable
	}
	return s.position.List(ctx, []domainrepo.Filter{{Column: "organization_unit_reference_id", Values: []any{value}}}, limit, offset)
}

// FindPositionByLegalEntityReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPositionByLegalEntityReferenceID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.Position], error) {
	if s == nil {
		return operation.Page[entity.Position]{}, hrerrors.ErrUnavailable
	}
	return s.position.List(ctx, []domainrepo.Filter{{Column: "legal_entity_reference_id", Values: []any{value}}}, limit, offset)
}

// FindPositionByReportsToPositionID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPositionByReportsToPositionID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.Position], error) {
	if s == nil {
		return operation.Page[entity.Position]{}, hrerrors.ErrUnavailable
	}
	return s.position.List(ctx, []domainrepo.Filter{{Column: "reports_to_position_id", Values: []any{value}}}, limit, offset)
}

// FindPositionByPositionStatus menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPositionByPositionStatus(ctx context.Context, value string, limit, offset int) (operation.Page[entity.Position], error) {
	if s == nil {
		return operation.Page[entity.Position]{}, hrerrors.ErrUnavailable
	}
	return s.position.List(ctx, []domainrepo.Filter{{Column: "position_status", Values: []any{value}}}, limit, offset)
}

// FindPositionByWorkLocationReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPositionByWorkLocationReferenceID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.Position], error) {
	if s == nil {
		return operation.Page[entity.Position]{}, hrerrors.ErrUnavailable
	}
	return s.position.List(ctx, []domainrepo.Filter{{Column: "work_location_reference_id", Values: []any{value}}}, limit, offset)
}

// FindPositionByCostCenterReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPositionByCostCenterReferenceID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.Position], error) {
	if s == nil {
		return operation.Page[entity.Position]{}, hrerrors.ErrUnavailable
	}
	return s.position.List(ctx, []domainrepo.Filter{{Column: "cost_center_reference_id", Values: []any{value}}}, limit, offset)
}

// CreateWorkforcePlan menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreateWorkforcePlan(ctx context.Context, in CreateWorkforcePlanInput) (*entity.WorkforcePlan, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.workforcePlan.Create(ctx, func() (*entity.WorkforcePlan, error) { return createWorkforcePlan(in) })
}

// UpdateWorkforcePlan menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdateWorkforcePlan(ctx context.Context, key string, in UpdateWorkforcePlanInput) (*entity.WorkforcePlan, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.workforcePlan.Update(ctx, key, func(value *entity.WorkforcePlan) error { return patchWorkforcePlan(value, in) })
}

// DeleteWorkforcePlan menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeleteWorkforcePlan(ctx context.Context, key string) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.workforcePlan.Delete(ctx, key)
}

// GetWorkforcePlanByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetWorkforcePlanByID(ctx context.Context, key string) (*entity.WorkforcePlan, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.workforcePlan.Get(ctx, key)
}

// GetAllWorkforcePlans menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllWorkforcePlans(ctx context.Context, limit, offset int) (operation.Page[entity.WorkforcePlan], error) {
	if s == nil {
		return operation.Page[entity.WorkforcePlan]{}, hrerrors.ErrUnavailable
	}
	return s.workforcePlan.List(ctx, nil, limit, offset)
}

// FindWorkforcePlanByPlanCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindWorkforcePlanByPlanCode(ctx context.Context, value string, limit, offset int) (operation.Page[entity.WorkforcePlan], error) {
	if s == nil {
		return operation.Page[entity.WorkforcePlan]{}, hrerrors.ErrUnavailable
	}
	return s.workforcePlan.List(ctx, []domainrepo.Filter{{Column: "plan_code", Values: []any{value}}}, limit, offset)
}

// FindWorkforcePlanByLegalEntityReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindWorkforcePlanByLegalEntityReferenceID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.WorkforcePlan], error) {
	if s == nil {
		return operation.Page[entity.WorkforcePlan]{}, hrerrors.ErrUnavailable
	}
	return s.workforcePlan.List(ctx, []domainrepo.Filter{{Column: "legal_entity_reference_id", Values: []any{value}}}, limit, offset)
}

// FindWorkforcePlanByScenarioCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindWorkforcePlanByScenarioCode(ctx context.Context, value string, limit, offset int) (operation.Page[entity.WorkforcePlan], error) {
	if s == nil {
		return operation.Page[entity.WorkforcePlan]{}, hrerrors.ErrUnavailable
	}
	return s.workforcePlan.List(ctx, []domainrepo.Filter{{Column: "scenario_code", Values: []any{value}}}, limit, offset)
}

// FindWorkforcePlanByStatus menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindWorkforcePlanByStatus(ctx context.Context, value string, limit, offset int) (operation.Page[entity.WorkforcePlan], error) {
	if s == nil {
		return operation.Page[entity.WorkforcePlan]{}, hrerrors.ErrUnavailable
	}
	return s.workforcePlan.List(ctx, []domainrepo.Filter{{Column: "status", Values: []any{value}}}, limit, offset)
}

// FindWorkforcePlanByFinanceBudgetReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindWorkforcePlanByFinanceBudgetReferenceID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.WorkforcePlan], error) {
	if s == nil {
		return operation.Page[entity.WorkforcePlan]{}, hrerrors.ErrUnavailable
	}
	return s.workforcePlan.List(ctx, []domainrepo.Filter{{Column: "finance_budget_reference_id", Values: []any{value}}}, limit, offset)
}

// FindWorkforcePlanByCreatedByPrincipalID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindWorkforcePlanByCreatedByPrincipalID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.WorkforcePlan], error) {
	if s == nil {
		return operation.Page[entity.WorkforcePlan]{}, hrerrors.ErrUnavailable
	}
	return s.workforcePlan.List(ctx, []domainrepo.Filter{{Column: "created_by_principal_id", Values: []any{value}}}, limit, offset)
}

// FindWorkforcePlanByApprovedByPrincipalID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindWorkforcePlanByApprovedByPrincipalID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.WorkforcePlan], error) {
	if s == nil {
		return operation.Page[entity.WorkforcePlan]{}, hrerrors.ErrUnavailable
	}
	return s.workforcePlan.List(ctx, []domainrepo.Filter{{Column: "approved_by_principal_id", Values: []any{value}}}, limit, offset)
}

// CreateWorkforcePlanLine menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreateWorkforcePlanLine(ctx context.Context, in CreateWorkforcePlanLineInput) (*entity.WorkforcePlanLine, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.workforcePlanLine.Create(ctx, func() (*entity.WorkforcePlanLine, error) { return createWorkforcePlanLine(in) })
}

// UpdateWorkforcePlanLine menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdateWorkforcePlanLine(ctx context.Context, key string, in UpdateWorkforcePlanLineInput) (*entity.WorkforcePlanLine, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.workforcePlanLine.Update(ctx, key, func(value *entity.WorkforcePlanLine) error { return patchWorkforcePlanLine(value, in) })
}

// DeleteWorkforcePlanLine menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeleteWorkforcePlanLine(ctx context.Context, key string) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.workforcePlanLine.Delete(ctx, key)
}

// GetWorkforcePlanLineByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetWorkforcePlanLineByID(ctx context.Context, key string) (*entity.WorkforcePlanLine, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.workforcePlanLine.Get(ctx, key)
}

// GetAllWorkforcePlanLines menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllWorkforcePlanLines(ctx context.Context, limit, offset int) (operation.Page[entity.WorkforcePlanLine], error) {
	if s == nil {
		return operation.Page[entity.WorkforcePlanLine]{}, hrerrors.ErrUnavailable
	}
	return s.workforcePlanLine.List(ctx, nil, limit, offset)
}

// FindWorkforcePlanLineByWorkforcePlanID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindWorkforcePlanLineByWorkforcePlanID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.WorkforcePlanLine], error) {
	if s == nil {
		return operation.Page[entity.WorkforcePlanLine]{}, hrerrors.ErrUnavailable
	}
	return s.workforcePlanLine.List(ctx, []domainrepo.Filter{{Column: "workforce_plan_id", Values: []any{value}}}, limit, offset)
}

// FindWorkforcePlanLineByLineNumber menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindWorkforcePlanLineByLineNumber(ctx context.Context, value int32, limit, offset int) (operation.Page[entity.WorkforcePlanLine], error) {
	if s == nil {
		return operation.Page[entity.WorkforcePlanLine]{}, hrerrors.ErrUnavailable
	}
	return s.workforcePlanLine.List(ctx, []domainrepo.Filter{{Column: "line_number", Values: []any{value}}}, limit, offset)
}

// FindWorkforcePlanLineByOrganizationUnitReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindWorkforcePlanLineByOrganizationUnitReferenceID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.WorkforcePlanLine], error) {
	if s == nil {
		return operation.Page[entity.WorkforcePlanLine]{}, hrerrors.ErrUnavailable
	}
	return s.workforcePlanLine.List(ctx, []domainrepo.Filter{{Column: "organization_unit_reference_id", Values: []any{value}}}, limit, offset)
}

// FindWorkforcePlanLineByJobID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindWorkforcePlanLineByJobID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.WorkforcePlanLine], error) {
	if s == nil {
		return operation.Page[entity.WorkforcePlanLine]{}, hrerrors.ErrUnavailable
	}
	return s.workforcePlanLine.List(ctx, []domainrepo.Filter{{Column: "job_id", Values: []any{value}}}, limit, offset)
}

// FindWorkforcePlanLineByWorkLocationReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindWorkforcePlanLineByWorkLocationReferenceID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.WorkforcePlanLine], error) {
	if s == nil {
		return operation.Page[entity.WorkforcePlanLine]{}, hrerrors.ErrUnavailable
	}
	return s.workforcePlanLine.List(ctx, []domainrepo.Filter{{Column: "work_location_reference_id", Values: []any{value}}}, limit, offset)
}

// FindWorkforcePlanLineByCurrencyCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindWorkforcePlanLineByCurrencyCode(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.WorkforcePlanLine], error) {
	if s == nil {
		return operation.Page[entity.WorkforcePlanLine]{}, hrerrors.ErrUnavailable
	}
	return s.workforcePlanLine.List(ctx, []domainrepo.Filter{{Column: "currency_code", Values: []any{value}}}, limit, offset)
}

// FindWorkforcePlanLineByActionCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindWorkforcePlanLineByActionCode(ctx context.Context, value string, limit, offset int) (operation.Page[entity.WorkforcePlanLine], error) {
	if s == nil {
		return operation.Page[entity.WorkforcePlanLine]{}, hrerrors.ErrUnavailable
	}
	return s.workforcePlanLine.List(ctx, []domainrepo.Filter{{Column: "action_code", Values: []any{value}}}, limit, offset)
}
