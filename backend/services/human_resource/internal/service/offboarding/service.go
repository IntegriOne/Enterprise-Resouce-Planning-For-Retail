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
	Workers              domainrepo.RecordRepository[entity.Worker, string]
	PayrollRuns          domainrepo.RecordRepository[entity.PayrollRun, string]
	ClearanceVerifier    ClearanceVerifier
	PayrollResults       domainrepo.RecordRepository[entity.PayrollResult, string]
	ConfirmationVerifier ConfirmationVerifier
	Runtime              operation.Dependencies
	OffboardingPlan      domainrepo.RecordRepository[entity.OffboardingPlan, string]
	OffboardingTask      domainrepo.RecordRepository[entity.OffboardingTask, string]
}

// Service mengoordinasikan operasi domain offboarding.
type Service struct {
	ids             operation.IDGenerator
	workers         domainrepo.RecordRepository[entity.Worker, string]
	payrollRuns     domainrepo.RecordRepository[entity.PayrollRun, string]
	clearance       ClearanceVerifier
	payrollResults  domainrepo.RecordRepository[entity.PayrollResult, string]
	confirmation    ConfirmationVerifier
	offboardingPlan *operation.Resource[entity.OffboardingPlan, string, *entity.OffboardingPlan]
	offboardingTask *operation.Resource[entity.OffboardingTask, string, *entity.OffboardingTask]
}

// NewService memeriksa seluruh port dan dependency sebelum service dapat dipakai.
func NewService(deps Dependencies) (*Service, error) {
	if operation.RequireDependency(deps.Runtime.IDs) != nil {
		deps.Runtime.IDs = operation.UUIDGenerator{}
	}
	s := &Service{ids: deps.Runtime.IDs, workers: deps.Workers, confirmation: deps.ConfirmationVerifier, clearance: deps.ClearanceVerifier, payrollResults: deps.PayrollResults, payrollRuns: deps.PayrollRuns}
	var err error
	s.offboardingPlan, err = operation.NewResource[entity.OffboardingPlan, string, *entity.OffboardingPlan](
		deps.OffboardingPlan,
		deps.Runtime,
		specOffboardingPlan(),
	)
	if err != nil {
		return nil, err
	}
	s.offboardingTask, err = operation.NewResource[entity.OffboardingTask, string, *entity.OffboardingTask](
		deps.OffboardingTask,
		deps.Runtime,
		specOffboardingTask(),
	)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// CreateOffboardingPlan menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreateOffboardingPlan(ctx context.Context, in CreateOffboardingPlanInput) (*entity.OffboardingPlan, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.offboardingPlan.Create(ctx, func() (*entity.OffboardingPlan, error) { return createOffboardingPlan(in) })
}

// UpdateOffboardingPlan menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdateOffboardingPlan(ctx context.Context, key string, in UpdateOffboardingPlanInput) (*entity.OffboardingPlan, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.offboardingPlan.Update(ctx, key, func(value *entity.OffboardingPlan) error { return patchOffboardingPlan(value, in) })
}

// DeleteOffboardingPlan menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeleteOffboardingPlan(ctx context.Context, key string) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.offboardingPlan.Delete(ctx, key)
}

// GetOffboardingPlanByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetOffboardingPlanByID(ctx context.Context, key string) (*entity.OffboardingPlan, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.offboardingPlan.Get(ctx, key)
}

// GetAllOffboardingPlans menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllOffboardingPlans(ctx context.Context, limit, offset int) (operation.Page[entity.OffboardingPlan], error) {
	if s == nil {
		return operation.Page[entity.OffboardingPlan]{}, hrerrors.ErrUnavailable
	}
	return s.offboardingPlan.List(ctx, nil, limit, offset)
}

// FindOffboardingPlanByWorkerID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindOffboardingPlanByWorkerID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.OffboardingPlan], error) {
	if s == nil {
		return operation.Page[entity.OffboardingPlan]{}, hrerrors.ErrUnavailable
	}
	return s.offboardingPlan.List(ctx, []domainrepo.Filter{{Column: "worker_id", Values: []any{value}}}, limit, offset)
}

// FindOffboardingPlanByPlanNumber menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindOffboardingPlanByPlanNumber(ctx context.Context, value string, limit, offset int) (operation.Page[entity.OffboardingPlan], error) {
	if s == nil {
		return operation.Page[entity.OffboardingPlan]{}, hrerrors.ErrUnavailable
	}
	return s.offboardingPlan.List(ctx, []domainrepo.Filter{{Column: "plan_number", Values: []any{value}}}, limit, offset)
}

// FindOffboardingPlanByReasonCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindOffboardingPlanByReasonCode(ctx context.Context, value string, limit, offset int) (operation.Page[entity.OffboardingPlan], error) {
	if s == nil {
		return operation.Page[entity.OffboardingPlan]{}, hrerrors.ErrUnavailable
	}
	return s.offboardingPlan.List(ctx, []domainrepo.Filter{{Column: "reason_code", Values: []any{value}}}, limit, offset)
}

// FindOffboardingPlanByStatus menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindOffboardingPlanByStatus(ctx context.Context, value string, limit, offset int) (operation.Page[entity.OffboardingPlan], error) {
	if s == nil {
		return operation.Page[entity.OffboardingPlan]{}, hrerrors.ErrUnavailable
	}
	return s.offboardingPlan.List(ctx, []domainrepo.Filter{{Column: "status", Values: []any{value}}}, limit, offset)
}

// FindOffboardingPlanByExitInterviewDocumentVersionReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindOffboardingPlanByExitInterviewDocumentVersionReferenceID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.OffboardingPlan], error) {
	if s == nil {
		return operation.Page[entity.OffboardingPlan]{}, hrerrors.ErrUnavailable
	}
	return s.offboardingPlan.List(ctx, []domainrepo.Filter{{Column: "exit_interview_document_version_reference_id", Values: []any{value}}}, limit, offset)
}

// FindOffboardingPlanByFinalPayrollRunID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindOffboardingPlanByFinalPayrollRunID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.OffboardingPlan], error) {
	if s == nil {
		return operation.Page[entity.OffboardingPlan]{}, hrerrors.ErrUnavailable
	}
	return s.offboardingPlan.List(ctx, []domainrepo.Filter{{Column: "final_payroll_run_id", Values: []any{value}}}, limit, offset)
}

// FindOffboardingPlanByFinanceClearanceReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindOffboardingPlanByFinanceClearanceReferenceID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.OffboardingPlan], error) {
	if s == nil {
		return operation.Page[entity.OffboardingPlan]{}, hrerrors.ErrUnavailable
	}
	return s.offboardingPlan.List(ctx, []domainrepo.Filter{{Column: "finance_clearance_reference_id", Values: []any{value}}}, limit, offset)
}

// FindOffboardingPlanByApprovedByPrincipalID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindOffboardingPlanByApprovedByPrincipalID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.OffboardingPlan], error) {
	if s == nil {
		return operation.Page[entity.OffboardingPlan]{}, hrerrors.ErrUnavailable
	}
	return s.offboardingPlan.List(ctx, []domainrepo.Filter{{Column: "approved_by_principal_id", Values: []any{value}}}, limit, offset)
}

// FindOffboardingPlanByCreatedByPrincipalID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindOffboardingPlanByCreatedByPrincipalID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.OffboardingPlan], error) {
	if s == nil {
		return operation.Page[entity.OffboardingPlan]{}, hrerrors.ErrUnavailable
	}
	return s.offboardingPlan.List(ctx, []domainrepo.Filter{{Column: "created_by_principal_id", Values: []any{value}}}, limit, offset)
}

// CreateOffboardingTask menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreateOffboardingTask(ctx context.Context, in CreateOffboardingTaskInput) (*entity.OffboardingTask, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.offboardingTask.Create(ctx, func() (*entity.OffboardingTask, error) { return createOffboardingTask(in) })
}

// UpdateOffboardingTask menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdateOffboardingTask(ctx context.Context, key string, in UpdateOffboardingTaskInput) (*entity.OffboardingTask, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.offboardingTask.Update(ctx, key, func(value *entity.OffboardingTask) error { return patchOffboardingTask(value, in) })
}

// DeleteOffboardingTask menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeleteOffboardingTask(ctx context.Context, key string) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.offboardingTask.Delete(ctx, key)
}

// GetOffboardingTaskByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetOffboardingTaskByID(ctx context.Context, key string) (*entity.OffboardingTask, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.offboardingTask.Get(ctx, key)
}

// GetAllOffboardingTasks menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllOffboardingTasks(ctx context.Context, limit, offset int) (operation.Page[entity.OffboardingTask], error) {
	if s == nil {
		return operation.Page[entity.OffboardingTask]{}, hrerrors.ErrUnavailable
	}
	return s.offboardingTask.List(ctx, nil, limit, offset)
}

// FindOffboardingTaskByOffboardingPlanID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindOffboardingTaskByOffboardingPlanID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.OffboardingTask], error) {
	if s == nil {
		return operation.Page[entity.OffboardingTask]{}, hrerrors.ErrUnavailable
	}
	return s.offboardingTask.List(ctx, []domainrepo.Filter{{Column: "offboarding_plan_id", Values: []any{value}}}, limit, offset)
}

// FindOffboardingTaskByTaskCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindOffboardingTaskByTaskCode(ctx context.Context, value string, limit, offset int) (operation.Page[entity.OffboardingTask], error) {
	if s == nil {
		return operation.Page[entity.OffboardingTask]{}, hrerrors.ErrUnavailable
	}
	return s.offboardingTask.List(ctx, []domainrepo.Filter{{Column: "task_code", Values: []any{value}}}, limit, offset)
}

// FindOffboardingTaskByAssignedPrincipalID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindOffboardingTaskByAssignedPrincipalID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.OffboardingTask], error) {
	if s == nil {
		return operation.Page[entity.OffboardingTask]{}, hrerrors.ErrUnavailable
	}
	return s.offboardingTask.List(ctx, []domainrepo.Filter{{Column: "assigned_principal_id", Values: []any{value}}}, limit, offset)
}

// FindOffboardingTaskByStatus menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindOffboardingTaskByStatus(ctx context.Context, value string, limit, offset int) (operation.Page[entity.OffboardingTask], error) {
	if s == nil {
		return operation.Page[entity.OffboardingTask]{}, hrerrors.ErrUnavailable
	}
	return s.offboardingTask.List(ctx, []domainrepo.Filter{{Column: "status", Values: []any{value}}}, limit, offset)
}

// FindOffboardingTaskByOwnerConfirmationReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindOffboardingTaskByOwnerConfirmationReferenceID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.OffboardingTask], error) {
	if s == nil {
		return operation.Page[entity.OffboardingTask]{}, hrerrors.ErrUnavailable
	}
	return s.offboardingTask.List(ctx, []domainrepo.Filter{{Column: "owner_confirmation_reference_id", Values: []any{value}}}, limit, offset)
}

// FindOffboardingTaskByWaivedByPrincipalID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindOffboardingTaskByWaivedByPrincipalID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.OffboardingTask], error) {
	if s == nil {
		return operation.Page[entity.OffboardingTask]{}, hrerrors.ErrUnavailable
	}
	return s.offboardingTask.List(ctx, []domainrepo.Filter{{Column: "waived_by_principal_id", Values: []any{value}}}, limit, offset)
}
