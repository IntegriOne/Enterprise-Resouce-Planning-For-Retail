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
	Worker                 domainrepo.RecordRepository[entity.Worker, string]
	Runtime                operation.Dependencies
	DevelopmentPlan        domainrepo.RecordRepository[entity.DevelopmentPlan, string]
	DevelopmentAction      domainrepo.RecordRepository[entity.DevelopmentAction, string]
	CareerPath             domainrepo.RecordRepository[entity.CareerPath, string]
	CareerPathTransition   domainrepo.RecordRepository[entity.CareerPathTransition, string]
	WorkerCareerAspiration domainrepo.RecordRepository[entity.WorkerCareerAspiration, string]
}

// Service mengoordinasikan operasi domain career.
type Service struct {
	workers                domainrepo.RecordRepository[entity.Worker, string]
	plans                  domainrepo.RecordRepository[entity.DevelopmentPlan, string]
	actions                domainrepo.RecordRepository[entity.DevelopmentAction, string]
	developmentPlan        *operation.Resource[entity.DevelopmentPlan, string, *entity.DevelopmentPlan]
	developmentAction      *operation.Resource[entity.DevelopmentAction, string, *entity.DevelopmentAction]
	careerPath             *operation.Resource[entity.CareerPath, string, *entity.CareerPath]
	careerPathTransition   *operation.Resource[entity.CareerPathTransition, string, *entity.CareerPathTransition]
	workerCareerAspiration *operation.Resource[entity.WorkerCareerAspiration, string, *entity.WorkerCareerAspiration]
}

// NewService memeriksa seluruh port dan dependency sebelum service dapat dipakai.
func NewService(deps Dependencies) (*Service, error) {
	s := &Service{workers: deps.Worker, plans: deps.DevelopmentPlan, actions: deps.DevelopmentAction}
	var err error
	s.developmentPlan, err = operation.NewResource[entity.DevelopmentPlan, string, *entity.DevelopmentPlan](
		deps.DevelopmentPlan,
		deps.Runtime,
		specDevelopmentPlan(),
	)
	if err != nil {
		return nil, err
	}
	s.developmentAction, err = operation.NewResource[entity.DevelopmentAction, string, *entity.DevelopmentAction](
		deps.DevelopmentAction,
		deps.Runtime,
		specDevelopmentAction(),
	)
	if err != nil {
		return nil, err
	}
	s.careerPath, err = operation.NewResource[entity.CareerPath, string, *entity.CareerPath](
		deps.CareerPath,
		deps.Runtime,
		specCareerPath(),
	)
	if err != nil {
		return nil, err
	}
	s.careerPathTransition, err = operation.NewResource[entity.CareerPathTransition, string, *entity.CareerPathTransition](
		deps.CareerPathTransition,
		deps.Runtime,
		specCareerPathTransition(),
	)
	if err != nil {
		return nil, err
	}
	s.workerCareerAspiration, err = operation.NewResource[
		entity.WorkerCareerAspiration,
		string,
		*entity.WorkerCareerAspiration,
	](
		deps.WorkerCareerAspiration,
		deps.Runtime,
		specWorkerCareerAspiration(),
	)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// CreateDevelopmentPlan menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreateDevelopmentPlan(ctx context.Context, in CreateDevelopmentPlanInput) (*entity.DevelopmentPlan, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.developmentPlan.Create(ctx, func() (*entity.DevelopmentPlan, error) { return createDevelopmentPlan(in) })
}

// UpdateDevelopmentPlan menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdateDevelopmentPlan(ctx context.Context, key string, in UpdateDevelopmentPlanInput) (*entity.DevelopmentPlan, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.developmentPlan.Update(ctx, key, func(value *entity.DevelopmentPlan) error { return patchDevelopmentPlan(value, in) })
}

// DeleteDevelopmentPlan menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeleteDevelopmentPlan(ctx context.Context, key string) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.developmentPlan.Delete(ctx, key)
}

// GetDevelopmentPlanByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetDevelopmentPlanByID(ctx context.Context, key string) (*entity.DevelopmentPlan, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.developmentPlan.Get(ctx, key)
}

// GetAllDevelopmentPlans menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllDevelopmentPlans(ctx context.Context, limit, offset int) (operation.Page[entity.DevelopmentPlan], error) {
	if s == nil {
		return operation.Page[entity.DevelopmentPlan]{}, hrerrors.ErrUnavailable
	}
	return s.developmentPlan.List(ctx, nil, limit, offset)
}

// FindDevelopmentPlanByWorkerID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindDevelopmentPlanByWorkerID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.DevelopmentPlan], error) {
	if s == nil {
		return operation.Page[entity.DevelopmentPlan]{}, hrerrors.ErrUnavailable
	}
	return s.developmentPlan.List(ctx, []domainrepo.Filter{{Column: "worker_id", Values: []any{value}}}, limit, offset)
}

// FindDevelopmentPlanByPlanCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindDevelopmentPlanByPlanCode(ctx context.Context, value string, limit, offset int) (operation.Page[entity.DevelopmentPlan], error) {
	if s == nil {
		return operation.Page[entity.DevelopmentPlan]{}, hrerrors.ErrUnavailable
	}
	return s.developmentPlan.List(ctx, []domainrepo.Filter{{Column: "plan_code", Values: []any{value}}}, limit, offset)
}

// FindDevelopmentPlanByTargetJobID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindDevelopmentPlanByTargetJobID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.DevelopmentPlan], error) {
	if s == nil {
		return operation.Page[entity.DevelopmentPlan]{}, hrerrors.ErrUnavailable
	}
	return s.developmentPlan.List(ctx, []domainrepo.Filter{{Column: "target_job_id", Values: []any{value}}}, limit, offset)
}

// FindDevelopmentPlanByStatus menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindDevelopmentPlanByStatus(ctx context.Context, value string, limit, offset int) (operation.Page[entity.DevelopmentPlan], error) {
	if s == nil {
		return operation.Page[entity.DevelopmentPlan]{}, hrerrors.ErrUnavailable
	}
	return s.developmentPlan.List(ctx, []domainrepo.Filter{{Column: "status", Values: []any{value}}}, limit, offset)
}

// FindDevelopmentPlanByMentorWorkerID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindDevelopmentPlanByMentorWorkerID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.DevelopmentPlan], error) {
	if s == nil {
		return operation.Page[entity.DevelopmentPlan]{}, hrerrors.ErrUnavailable
	}
	return s.developmentPlan.List(ctx, []domainrepo.Filter{{Column: "mentor_worker_id", Values: []any{value}}}, limit, offset)
}

// FindDevelopmentPlanByApprovedByPrincipalID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindDevelopmentPlanByApprovedByPrincipalID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.DevelopmentPlan], error) {
	if s == nil {
		return operation.Page[entity.DevelopmentPlan]{}, hrerrors.ErrUnavailable
	}
	return s.developmentPlan.List(ctx, []domainrepo.Filter{{Column: "approved_by_principal_id", Values: []any{value}}}, limit, offset)
}

// CreateDevelopmentAction menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreateDevelopmentAction(ctx context.Context, in CreateDevelopmentActionInput) (*entity.DevelopmentAction, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.developmentAction.Create(ctx, func() (*entity.DevelopmentAction, error) { return createDevelopmentAction(in) })
}

// UpdateDevelopmentAction menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdateDevelopmentAction(ctx context.Context, key string, in UpdateDevelopmentActionInput) (*entity.DevelopmentAction, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.developmentAction.Update(ctx, key, func(value *entity.DevelopmentAction) error { return patchDevelopmentAction(value, in) })
}

// DeleteDevelopmentAction menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeleteDevelopmentAction(ctx context.Context, key string) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.developmentAction.Delete(ctx, key)
}

// GetDevelopmentActionByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetDevelopmentActionByID(ctx context.Context, key string) (*entity.DevelopmentAction, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.developmentAction.Get(ctx, key)
}

// GetAllDevelopmentActions menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllDevelopmentActions(ctx context.Context, limit, offset int) (operation.Page[entity.DevelopmentAction], error) {
	if s == nil {
		return operation.Page[entity.DevelopmentAction]{}, hrerrors.ErrUnavailable
	}
	return s.developmentAction.List(ctx, nil, limit, offset)
}

// FindDevelopmentActionByDevelopmentPlanID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindDevelopmentActionByDevelopmentPlanID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.DevelopmentAction], error) {
	if s == nil {
		return operation.Page[entity.DevelopmentAction]{}, hrerrors.ErrUnavailable
	}
	return s.developmentAction.List(ctx, []domainrepo.Filter{{Column: "development_plan_id", Values: []any{value}}}, limit, offset)
}

// FindDevelopmentActionByActionNumber menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindDevelopmentActionByActionNumber(ctx context.Context, value int32, limit, offset int) (operation.Page[entity.DevelopmentAction], error) {
	if s == nil {
		return operation.Page[entity.DevelopmentAction]{}, hrerrors.ErrUnavailable
	}
	return s.developmentAction.List(ctx, []domainrepo.Filter{{Column: "action_number", Values: []any{value}}}, limit, offset)
}

// FindDevelopmentActionByCompetencyID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindDevelopmentActionByCompetencyID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.DevelopmentAction], error) {
	if s == nil {
		return operation.Page[entity.DevelopmentAction]{}, hrerrors.ErrUnavailable
	}
	return s.developmentAction.List(ctx, []domainrepo.Filter{{Column: "competency_id", Values: []any{value}}}, limit, offset)
}

// FindDevelopmentActionByCourseID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindDevelopmentActionByCourseID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.DevelopmentAction], error) {
	if s == nil {
		return operation.Page[entity.DevelopmentAction]{}, hrerrors.ErrUnavailable
	}
	return s.developmentAction.List(ctx, []domainrepo.Filter{{Column: "course_id", Values: []any{value}}}, limit, offset)
}

// FindDevelopmentActionByActionType menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindDevelopmentActionByActionType(ctx context.Context, value string, limit, offset int) (operation.Page[entity.DevelopmentAction], error) {
	if s == nil {
		return operation.Page[entity.DevelopmentAction]{}, hrerrors.ErrUnavailable
	}
	return s.developmentAction.List(ctx, []domainrepo.Filter{{Column: "action_type", Values: []any{value}}}, limit, offset)
}

// FindDevelopmentActionByEvidenceDocumentVersionReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindDevelopmentActionByEvidenceDocumentVersionReferenceID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.DevelopmentAction], error) {
	if s == nil {
		return operation.Page[entity.DevelopmentAction]{}, hrerrors.ErrUnavailable
	}
	return s.developmentAction.List(ctx, []domainrepo.Filter{{Column: "evidence_document_version_reference_id", Values: []any{value}}}, limit, offset)
}

// CreateCareerPath menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreateCareerPath(ctx context.Context, in CreateCareerPathInput) (*entity.CareerPath, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.careerPath.Create(ctx, func() (*entity.CareerPath, error) { return createCareerPath(in) })
}

// UpdateCareerPath menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdateCareerPath(ctx context.Context, key string, in UpdateCareerPathInput) (*entity.CareerPath, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.careerPath.Update(ctx, key, func(value *entity.CareerPath) error { return patchCareerPath(value, in) })
}

// DeleteCareerPath menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeleteCareerPath(ctx context.Context, key string) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.careerPath.Delete(ctx, key)
}

// GetCareerPathByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetCareerPathByID(ctx context.Context, key string) (*entity.CareerPath, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.careerPath.Get(ctx, key)
}

// GetAllCareerPaths menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllCareerPaths(ctx context.Context, limit, offset int) (operation.Page[entity.CareerPath], error) {
	if s == nil {
		return operation.Page[entity.CareerPath]{}, hrerrors.ErrUnavailable
	}
	return s.careerPath.List(ctx, nil, limit, offset)
}

// FindCareerPathByPathCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindCareerPathByPathCode(ctx context.Context, value string, limit, offset int) (operation.Page[entity.CareerPath], error) {
	if s == nil {
		return operation.Page[entity.CareerPath]{}, hrerrors.ErrUnavailable
	}
	return s.careerPath.List(ctx, []domainrepo.Filter{{Column: "path_code", Values: []any{value}}}, limit, offset)
}

// FindCareerPathByJobFamilyID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindCareerPathByJobFamilyID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.CareerPath], error) {
	if s == nil {
		return operation.Page[entity.CareerPath]{}, hrerrors.ErrUnavailable
	}
	return s.careerPath.List(ctx, []domainrepo.Filter{{Column: "job_family_id", Values: []any{value}}}, limit, offset)
}

// FindCareerPathByStatus membaca jalur berdasarkan status dengan paginasi.
func (s *Service) FindCareerPathByStatus(ctx context.Context, value string, limit, offset int) (operation.Page[entity.CareerPath], error) {
	if s == nil {
		return operation.Page[entity.CareerPath]{}, hrerrors.ErrUnavailable
	}
	return s.careerPath.List(ctx, []domainrepo.Filter{{Column: "status", Values: []any{value}}}, limit, offset)
}

// FindCareerPathByCreatedByPrincipalID membaca jalur berdasarkan pembuat dengan paginasi.
func (s *Service) FindCareerPathByCreatedByPrincipalID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.CareerPath], error) {
	if s == nil {
		return operation.Page[entity.CareerPath]{}, hrerrors.ErrUnavailable
	}
	return s.careerPath.List(ctx, []domainrepo.Filter{{Column: "created_by_principal_id", Values: []any{value}}}, limit, offset)
}

// FindCareerPathByApprovedByPrincipalID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindCareerPathByApprovedByPrincipalID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.CareerPath], error) {
	if s == nil {
		return operation.Page[entity.CareerPath]{}, hrerrors.ErrUnavailable
	}
	return s.careerPath.List(ctx, []domainrepo.Filter{{Column: "approved_by_principal_id", Values: []any{value}}}, limit, offset)
}

// CreateCareerPathTransition menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreateCareerPathTransition(ctx context.Context, in CreateCareerPathTransitionInput) (*entity.CareerPathTransition, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.careerPathTransition.Create(ctx, func() (*entity.CareerPathTransition, error) { return createCareerPathTransition(in) })
}

// UpdateCareerPathTransition menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdateCareerPathTransition(ctx context.Context, key string, in UpdateCareerPathTransitionInput) (*entity.CareerPathTransition, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.careerPathTransition.Update(ctx, key, func(value *entity.CareerPathTransition) error { return patchCareerPathTransition(value, in) })
}

// DeleteCareerPathTransition menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeleteCareerPathTransition(ctx context.Context, key string) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.careerPathTransition.Delete(ctx, key)
}

// GetCareerPathTransitionByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetCareerPathTransitionByID(ctx context.Context, key string) (*entity.CareerPathTransition, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.careerPathTransition.Get(ctx, key)
}

// GetAllCareerPathTransitions menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllCareerPathTransitions(ctx context.Context, limit, offset int) (operation.Page[entity.CareerPathTransition], error) {
	if s == nil {
		return operation.Page[entity.CareerPathTransition]{}, hrerrors.ErrUnavailable
	}
	return s.careerPathTransition.List(ctx, nil, limit, offset)
}

// FindCareerPathTransitionByCareerPathID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindCareerPathTransitionByCareerPathID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.CareerPathTransition], error) {
	if s == nil {
		return operation.Page[entity.CareerPathTransition]{}, hrerrors.ErrUnavailable
	}
	return s.careerPathTransition.List(ctx, []domainrepo.Filter{{Column: "career_path_id", Values: []any{value}}}, limit, offset)
}

// FindCareerPathTransitionByFromJobID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindCareerPathTransitionByFromJobID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.CareerPathTransition], error) {
	if s == nil {
		return operation.Page[entity.CareerPathTransition]{}, hrerrors.ErrUnavailable
	}
	return s.careerPathTransition.List(ctx, []domainrepo.Filter{{Column: "from_job_id", Values: []any{value}}}, limit, offset)
}

// FindCareerPathTransitionByToJobID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindCareerPathTransitionByToJobID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.CareerPathTransition], error) {
	if s == nil {
		return operation.Page[entity.CareerPathTransition]{}, hrerrors.ErrUnavailable
	}
	return s.careerPathTransition.List(ctx, []domainrepo.Filter{{Column: "to_job_id", Values: []any{value}}}, limit, offset)
}

// FindCareerPathTransitionByMovementType menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindCareerPathTransitionByMovementType(ctx context.Context, value string, limit, offset int) (operation.Page[entity.CareerPathTransition], error) {
	if s == nil {
		return operation.Page[entity.CareerPathTransition]{}, hrerrors.ErrUnavailable
	}
	return s.careerPathTransition.List(ctx, []domainrepo.Filter{{Column: "movement_type", Values: []any{value}}}, limit, offset)
}

// FindCareerPathTransitionByRequirementsDocumentVersionReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindCareerPathTransitionByRequirementsDocumentVersionReferenceID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.CareerPathTransition], error) {
	if s == nil {
		return operation.Page[entity.CareerPathTransition]{}, hrerrors.ErrUnavailable
	}
	return s.careerPathTransition.List(ctx, []domainrepo.Filter{{Column: "requirements_document_version_reference_id", Values: []any{value}}}, limit, offset)
}

// CreateWorkerCareerAspiration menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreateWorkerCareerAspiration(ctx context.Context, in CreateWorkerCareerAspirationInput) (*entity.WorkerCareerAspiration, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.workerCareerAspiration.Create(ctx, func() (*entity.WorkerCareerAspiration, error) { return createWorkerCareerAspiration(in) })
}

// UpdateWorkerCareerAspiration menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdateWorkerCareerAspiration(ctx context.Context, key string, in UpdateWorkerCareerAspirationInput) (*entity.WorkerCareerAspiration, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.workerCareerAspiration.Update(ctx, key, func(value *entity.WorkerCareerAspiration) error { return patchWorkerCareerAspiration(value, in) })
}

// DeleteWorkerCareerAspiration menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeleteWorkerCareerAspiration(ctx context.Context, key string) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.workerCareerAspiration.Delete(ctx, key)
}

// GetWorkerCareerAspirationByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetWorkerCareerAspirationByID(ctx context.Context, key string) (*entity.WorkerCareerAspiration, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.workerCareerAspiration.Get(ctx, key)
}

// GetAllWorkerCareerAspirations menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllWorkerCareerAspirations(ctx context.Context, limit, offset int) (operation.Page[entity.WorkerCareerAspiration], error) {
	if s == nil {
		return operation.Page[entity.WorkerCareerAspiration]{}, hrerrors.ErrUnavailable
	}
	return s.workerCareerAspiration.List(ctx, nil, limit, offset)
}

// FindWorkerCareerAspirationByWorkerID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindWorkerCareerAspirationByWorkerID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.WorkerCareerAspiration], error) {
	if s == nil {
		return operation.Page[entity.WorkerCareerAspiration]{}, hrerrors.ErrUnavailable
	}
	return s.workerCareerAspiration.List(ctx, []domainrepo.Filter{{Column: "worker_id", Values: []any{value}}}, limit, offset)
}

// FindWorkerCareerAspirationByTargetJobID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindWorkerCareerAspirationByTargetJobID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.WorkerCareerAspiration], error) {
	if s == nil {
		return operation.Page[entity.WorkerCareerAspiration]{}, hrerrors.ErrUnavailable
	}
	return s.workerCareerAspiration.List(ctx, []domainrepo.Filter{{Column: "target_job_id", Values: []any{value}}}, limit, offset)
}

// FindWorkerCareerAspirationByCareerPathTransitionID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindWorkerCareerAspirationByCareerPathTransitionID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.WorkerCareerAspiration], error) {
	if s == nil {
		return operation.Page[entity.WorkerCareerAspiration]{}, hrerrors.ErrUnavailable
	}
	return s.workerCareerAspiration.List(ctx, []domainrepo.Filter{{Column: "career_path_transition_id", Values: []any{value}}}, limit, offset)
}

// FindWorkerCareerAspirationByRecordedByPrincipalID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindWorkerCareerAspirationByRecordedByPrincipalID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.WorkerCareerAspiration], error) {
	if s == nil {
		return operation.Page[entity.WorkerCareerAspiration]{}, hrerrors.ErrUnavailable
	}
	return s.workerCareerAspiration.List(ctx, []domainrepo.Filter{{Column: "recorded_by_principal_id", Values: []any{value}}}, limit, offset)
}

// FindWorkerCareerAspirationByPreferredLocationReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindWorkerCareerAspirationByPreferredLocationReferenceID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.WorkerCareerAspiration], error) {
	if s == nil {
		return operation.Page[entity.WorkerCareerAspiration]{}, hrerrors.ErrUnavailable
	}
	return s.workerCareerAspiration.List(ctx, []domainrepo.Filter{{Column: "preferred_location_reference_id", Values: []any{value}}}, limit, offset)
}
