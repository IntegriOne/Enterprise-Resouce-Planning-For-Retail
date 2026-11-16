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
	Runtime        operation.Dependencies
	OnboardingPlan domainrepo.RecordRepository[entity.OnboardingPlan, string]
	OnboardingTask domainrepo.RecordRepository[entity.OnboardingTask, string]
	Evidence       EvidenceVerifier
}

// Service mengoordinasikan operasi domain onboarding.
type Service struct {
	onboardingPlan *operation.Resource[entity.OnboardingPlan, string, *entity.OnboardingPlan]
	onboardingTask *operation.Resource[entity.OnboardingTask, string, *entity.OnboardingTask]
	tasks          domainrepo.RecordRepository[entity.OnboardingTask, string]
	evidence       EvidenceVerifier
}

// NewService memeriksa seluruh port dan dependency sebelum service dapat dipakai.
func NewService(deps Dependencies) (*Service, error) {
	s := &Service{}
	var err error
	s.onboardingPlan, err = operation.NewResource[entity.OnboardingPlan, string, *entity.OnboardingPlan](
		deps.OnboardingPlan,
		deps.Runtime,
		specOnboardingPlan(),
	)
	if err != nil {
		return nil, err
	}
	s.onboardingTask, err = operation.NewResource[entity.OnboardingTask, string, *entity.OnboardingTask](
		deps.OnboardingTask,
		deps.Runtime,
		specOnboardingTask(),
	)
	if err != nil {
		return nil, err
	}
	s.tasks = deps.OnboardingTask
	s.evidence = deps.Evidence
	return s, nil
}

// CreateOnboardingPlan menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreateOnboardingPlan(ctx context.Context, in CreateOnboardingPlanInput) (*entity.OnboardingPlan, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.onboardingPlan.Create(ctx, func() (*entity.OnboardingPlan, error) { return createOnboardingPlan(in) })
}

// UpdateOnboardingPlan menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdateOnboardingPlan(ctx context.Context, key string, in UpdateOnboardingPlanInput) (*entity.OnboardingPlan, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.onboardingPlan.Update(ctx, key, func(value *entity.OnboardingPlan) error { return patchOnboardingPlan(value, in) })
}

// DeleteOnboardingPlan menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeleteOnboardingPlan(ctx context.Context, key string) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.onboardingPlan.Delete(ctx, key)
}

// GetOnboardingPlanByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetOnboardingPlanByID(ctx context.Context, key string) (*entity.OnboardingPlan, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.onboardingPlan.Get(ctx, key)
}

// GetAllOnboardingPlans menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllOnboardingPlans(ctx context.Context, limit, offset int) (operation.Page[entity.OnboardingPlan], error) {
	if s == nil {
		return operation.Page[entity.OnboardingPlan]{}, hrerrors.ErrUnavailable
	}
	return s.onboardingPlan.List(ctx, nil, limit, offset)
}

// FindOnboardingPlanByWorkerID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindOnboardingPlanByWorkerID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.OnboardingPlan], error) {
	if s == nil {
		return operation.Page[entity.OnboardingPlan]{}, hrerrors.ErrUnavailable
	}
	return s.onboardingPlan.List(ctx, []domainrepo.Filter{{Column: "worker_id", Values: []any{value}}}, limit, offset)
}

// FindOnboardingPlanByPlanCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindOnboardingPlanByPlanCode(ctx context.Context, value string, limit, offset int) (operation.Page[entity.OnboardingPlan], error) {
	if s == nil {
		return operation.Page[entity.OnboardingPlan]{}, hrerrors.ErrUnavailable
	}
	return s.onboardingPlan.List(ctx, []domainrepo.Filter{{Column: "plan_code", Values: []any{value}}}, limit, offset)
}

// FindOnboardingPlanByStatus menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindOnboardingPlanByStatus(ctx context.Context, value string, limit, offset int) (operation.Page[entity.OnboardingPlan], error) {
	if s == nil {
		return operation.Page[entity.OnboardingPlan]{}, hrerrors.ErrUnavailable
	}
	return s.onboardingPlan.List(ctx, []domainrepo.Filter{{Column: "status", Values: []any{value}}}, limit, offset)
}

// FindOnboardingPlanByTemplateReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindOnboardingPlanByTemplateReferenceID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.OnboardingPlan], error) {
	if s == nil {
		return operation.Page[entity.OnboardingPlan]{}, hrerrors.ErrUnavailable
	}
	return s.onboardingPlan.List(ctx, []domainrepo.Filter{{Column: "template_reference_id", Values: []any{value}}}, limit, offset)
}

// FindOnboardingPlanByCreatedByPrincipalID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindOnboardingPlanByCreatedByPrincipalID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.OnboardingPlan], error) {
	if s == nil {
		return operation.Page[entity.OnboardingPlan]{}, hrerrors.ErrUnavailable
	}
	return s.onboardingPlan.List(ctx, []domainrepo.Filter{{Column: "created_by_principal_id", Values: []any{value}}}, limit, offset)
}

// CreateOnboardingTask menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreateOnboardingTask(ctx context.Context, in CreateOnboardingTaskInput) (*entity.OnboardingTask, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.onboardingTask.Create(ctx, func() (*entity.OnboardingTask, error) { return createOnboardingTask(in) })
}

// UpdateOnboardingTask menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdateOnboardingTask(ctx context.Context, key string, in UpdateOnboardingTaskInput) (*entity.OnboardingTask, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.onboardingTask.Update(ctx, key, func(value *entity.OnboardingTask) error { return patchOnboardingTask(value, in) })
}

// DeleteOnboardingTask menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeleteOnboardingTask(ctx context.Context, key string) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.onboardingTask.Delete(ctx, key)
}

// GetOnboardingTaskByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetOnboardingTaskByID(ctx context.Context, key string) (*entity.OnboardingTask, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.onboardingTask.Get(ctx, key)
}

// GetAllOnboardingTasks menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllOnboardingTasks(ctx context.Context, limit, offset int) (operation.Page[entity.OnboardingTask], error) {
	if s == nil {
		return operation.Page[entity.OnboardingTask]{}, hrerrors.ErrUnavailable
	}
	return s.onboardingTask.List(ctx, nil, limit, offset)
}

// FindOnboardingTaskByOnboardingPlanID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindOnboardingTaskByOnboardingPlanID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.OnboardingTask], error) {
	if s == nil {
		return operation.Page[entity.OnboardingTask]{}, hrerrors.ErrUnavailable
	}
	return s.onboardingTask.List(ctx, []domainrepo.Filter{{Column: "onboarding_plan_id", Values: []any{value}}}, limit, offset)
}

// FindOnboardingTaskByTaskCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindOnboardingTaskByTaskCode(ctx context.Context, value string, limit, offset int) (operation.Page[entity.OnboardingTask], error) {
	if s == nil {
		return operation.Page[entity.OnboardingTask]{}, hrerrors.ErrUnavailable
	}
	return s.onboardingTask.List(ctx, []domainrepo.Filter{{Column: "task_code", Values: []any{value}}}, limit, offset)
}

// FindOnboardingTaskByTaskOwnerType menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindOnboardingTaskByTaskOwnerType(ctx context.Context, value string, limit, offset int) (operation.Page[entity.OnboardingTask], error) {
	if s == nil {
		return operation.Page[entity.OnboardingTask]{}, hrerrors.ErrUnavailable
	}
	return s.onboardingTask.List(ctx, []domainrepo.Filter{{Column: "task_owner_type", Values: []any{value}}}, limit, offset)
}

// FindOnboardingTaskByAssignedPrincipalID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindOnboardingTaskByAssignedPrincipalID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.OnboardingTask], error) {
	if s == nil {
		return operation.Page[entity.OnboardingTask]{}, hrerrors.ErrUnavailable
	}
	return s.onboardingTask.List(ctx, []domainrepo.Filter{{Column: "assigned_principal_id", Values: []any{value}}}, limit, offset)
}

// FindOnboardingTaskByCompletionEvidenceReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindOnboardingTaskByCompletionEvidenceReferenceID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.OnboardingTask], error) {
	if s == nil {
		return operation.Page[entity.OnboardingTask]{}, hrerrors.ErrUnavailable
	}
	return s.onboardingTask.List(ctx, []domainrepo.Filter{{Column: "completion_evidence_reference_id", Values: []any{value}}}, limit, offset)
}
