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
	Runtime                 operation.Dependencies
	Requisition             domainrepo.RecordRepository[entity.Requisition, string]
	Candidate               domainrepo.RecordRepository[entity.Candidate, string]
	Application             domainrepo.RecordRepository[entity.Application, string]
	ApplicationStageHistory domainrepo.RecordRepository[entity.ApplicationStageHistory, int64]
	Interview               domainrepo.RecordRepository[entity.Interview, string]
	InterviewPanelist       domainrepo.RecordRepository[entity.InterviewPanelist, string]
	InterviewFeedback       domainrepo.RecordRepository[entity.InterviewFeedback, string]
	Offer                   domainrepo.RecordRepository[entity.Offer, string]
	JobPosting              domainrepo.RecordRepository[entity.JobPosting, string]
	ApplicationAssessment   domainrepo.RecordRepository[entity.ApplicationAssessment, string]
}

// Service mengoordinasikan operasi domain recruitment.
type Service struct {
	requisition             *operation.Resource[entity.Requisition, string, *entity.Requisition]
	candidate               *operation.Resource[entity.Candidate, string, *entity.Candidate]
	application             *operation.Resource[entity.Application, string, *entity.Application]
	applicationStageHistory *operation.Resource[entity.ApplicationStageHistory, int64, *entity.ApplicationStageHistory]
	stageHistoryWriter      domainrepo.RecordRepository[entity.ApplicationStageHistory, int64]
	interview               *operation.Resource[entity.Interview, string, *entity.Interview]
	interviewPanelist       *operation.Resource[entity.InterviewPanelist, string, *entity.InterviewPanelist]
	interviewFeedback       *operation.Resource[entity.InterviewFeedback, string, *entity.InterviewFeedback]
	offer                   *operation.Resource[entity.Offer, string, *entity.Offer]
	jobPosting              *operation.Resource[entity.JobPosting, string, *entity.JobPosting]
	applicationAssessment   *operation.Resource[entity.ApplicationAssessment, string, *entity.ApplicationAssessment]
}

// NewService memeriksa seluruh port dan dependency sebelum service dapat dipakai.
func NewService(deps Dependencies) (*Service, error) {
	s := &Service{}
	var err error
	s.requisition, err = operation.NewResource[entity.Requisition, string, *entity.Requisition](
		deps.Requisition,
		deps.Runtime,
		specRequisition(),
	)
	if err != nil {
		return nil, err
	}
	s.candidate, err = operation.NewResource[entity.Candidate, string, *entity.Candidate](
		deps.Candidate,
		deps.Runtime,
		specCandidate(),
	)
	if err != nil {
		return nil, err
	}
	s.application, err = operation.NewResource[entity.Application, string, *entity.Application](
		deps.Application,
		deps.Runtime,
		specApplication(),
	)
	if err != nil {
		return nil, err
	}
	s.applicationStageHistory, err = operation.NewResource[
		entity.ApplicationStageHistory,
		int64,
		*entity.ApplicationStageHistory,
	](
		deps.ApplicationStageHistory,
		deps.Runtime,
		specApplicationStageHistory(),
	)
	if err != nil {
		return nil, err
	}
	s.stageHistoryWriter = deps.ApplicationStageHistory
	s.interview, err = operation.NewResource[entity.Interview, string, *entity.Interview](
		deps.Interview,
		deps.Runtime,
		specInterview(),
	)
	if err != nil {
		return nil, err
	}
	s.interviewPanelist, err = operation.NewResource[entity.InterviewPanelist, string, *entity.InterviewPanelist](
		deps.InterviewPanelist,
		deps.Runtime,
		specInterviewPanelist(),
	)
	if err != nil {
		return nil, err
	}
	s.interviewFeedback, err = operation.NewResource[entity.InterviewFeedback, string, *entity.InterviewFeedback](
		deps.InterviewFeedback,
		deps.Runtime,
		specInterviewFeedback(),
	)
	if err != nil {
		return nil, err
	}
	s.offer, err = operation.NewResource[entity.Offer, string, *entity.Offer](
		deps.Offer,
		deps.Runtime,
		specOffer(),
	)
	if err != nil {
		return nil, err
	}
	s.jobPosting, err = operation.NewResource[entity.JobPosting, string, *entity.JobPosting](
		deps.JobPosting,
		deps.Runtime,
		specJobPosting(),
	)
	if err != nil {
		return nil, err
	}
	s.applicationAssessment, err = operation.NewResource[
		entity.ApplicationAssessment,
		string,
		*entity.ApplicationAssessment,
	](
		deps.ApplicationAssessment,
		deps.Runtime,
		specApplicationAssessment(),
	)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// CreateRequisition menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreateRequisition(ctx context.Context, in CreateRequisitionInput) (*entity.Requisition, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.requisition.Create(ctx, func() (*entity.Requisition, error) { return createRequisition(in) })
}

// UpdateRequisition menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdateRequisition(ctx context.Context, key string, in UpdateRequisitionInput) (*entity.Requisition, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.requisition.Update(ctx, key, func(value *entity.Requisition) error { return patchRequisition(value, in) })
}

// DeleteRequisition menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeleteRequisition(ctx context.Context, key string) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.requisition.Delete(ctx, key)
}

// GetRequisitionByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetRequisitionByID(ctx context.Context, key string) (*entity.Requisition, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.requisition.Get(ctx, key)
}

// GetAllRequisitions menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllRequisitions(ctx context.Context, limit, offset int) (operation.Page[entity.Requisition], error) {
	if s == nil {
		return operation.Page[entity.Requisition]{}, hrerrors.ErrUnavailable
	}
	return s.requisition.List(ctx, nil, limit, offset)
}

// FindRequisitionByRequisitionNumber menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindRequisitionByRequisitionNumber(ctx context.Context, value string, limit, offset int) (operation.Page[entity.Requisition], error) {
	if s == nil {
		return operation.Page[entity.Requisition]{}, hrerrors.ErrUnavailable
	}
	return s.requisition.List(ctx, []domainrepo.Filter{{Column: "requisition_number", Values: []any{value}}}, limit, offset)
}

// FindRequisitionByWorkforcePlanLineID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindRequisitionByWorkforcePlanLineID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.Requisition], error) {
	if s == nil {
		return operation.Page[entity.Requisition]{}, hrerrors.ErrUnavailable
	}
	return s.requisition.List(ctx, []domainrepo.Filter{{Column: "workforce_plan_line_id", Values: []any{value}}}, limit, offset)
}

// FindRequisitionByPositionID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindRequisitionByPositionID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.Requisition], error) {
	if s == nil {
		return operation.Page[entity.Requisition]{}, hrerrors.ErrUnavailable
	}
	return s.requisition.List(ctx, []domainrepo.Filter{{Column: "position_id", Values: []any{value}}}, limit, offset)
}

// FindRequisitionByJobID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindRequisitionByJobID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.Requisition], error) {
	if s == nil {
		return operation.Page[entity.Requisition]{}, hrerrors.ErrUnavailable
	}
	return s.requisition.List(ctx, []domainrepo.Filter{{Column: "job_id", Values: []any{value}}}, limit, offset)
}

// FindRequisitionByOrganizationUnitReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindRequisitionByOrganizationUnitReferenceID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.Requisition], error) {
	if s == nil {
		return operation.Page[entity.Requisition]{}, hrerrors.ErrUnavailable
	}
	return s.requisition.List(ctx, []domainrepo.Filter{{Column: "organization_unit_reference_id", Values: []any{value}}}, limit, offset)
}

// FindRequisitionByHiringManagerPrincipalID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindRequisitionByHiringManagerPrincipalID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.Requisition], error) {
	if s == nil {
		return operation.Page[entity.Requisition]{}, hrerrors.ErrUnavailable
	}
	return s.requisition.List(ctx, []domainrepo.Filter{{Column: "hiring_manager_principal_id", Values: []any{value}}}, limit, offset)
}

// FindRequisitionByRecruiterPrincipalID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindRequisitionByRecruiterPrincipalID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.Requisition], error) {
	if s == nil {
		return operation.Page[entity.Requisition]{}, hrerrors.ErrUnavailable
	}
	return s.requisition.List(ctx, []domainrepo.Filter{{Column: "recruiter_principal_id", Values: []any{value}}}, limit, offset)
}

// FindRequisitionByRequisitionType menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindRequisitionByRequisitionType(ctx context.Context, value string, limit, offset int) (operation.Page[entity.Requisition], error) {
	if s == nil {
		return operation.Page[entity.Requisition]{}, hrerrors.ErrUnavailable
	}
	return s.requisition.List(ctx, []domainrepo.Filter{{Column: "requisition_type", Values: []any{value}}}, limit, offset)
}

// FindRequisitionByStatus menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindRequisitionByStatus(ctx context.Context, value string, limit, offset int) (operation.Page[entity.Requisition], error) {
	if s == nil {
		return operation.Page[entity.Requisition]{}, hrerrors.ErrUnavailable
	}
	return s.requisition.List(ctx, []domainrepo.Filter{{Column: "status", Values: []any{value}}}, limit, offset)
}

// FindRequisitionByBudgetReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindRequisitionByBudgetReferenceID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.Requisition], error) {
	if s == nil {
		return operation.Page[entity.Requisition]{}, hrerrors.ErrUnavailable
	}
	return s.requisition.List(ctx, []domainrepo.Filter{{Column: "budget_reference_id", Values: []any{value}}}, limit, offset)
}

// FindRequisitionByWorkLocationReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindRequisitionByWorkLocationReferenceID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.Requisition], error) {
	if s == nil {
		return operation.Page[entity.Requisition]{}, hrerrors.ErrUnavailable
	}
	return s.requisition.List(ctx, []domainrepo.Filter{{Column: "work_location_reference_id", Values: []any{value}}}, limit, offset)
}

// FindRequisitionByApprovedByPrincipalID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindRequisitionByApprovedByPrincipalID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.Requisition], error) {
	if s == nil {
		return operation.Page[entity.Requisition]{}, hrerrors.ErrUnavailable
	}
	return s.requisition.List(ctx, []domainrepo.Filter{{Column: "approved_by_principal_id", Values: []any{value}}}, limit, offset)
}

// FindRequisitionByIdempotencyKey menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindRequisitionByIdempotencyKey(ctx context.Context, value string, limit, offset int) (operation.Page[entity.Requisition], error) {
	if s == nil {
		return operation.Page[entity.Requisition]{}, hrerrors.ErrUnavailable
	}
	return s.requisition.List(ctx, []domainrepo.Filter{{Column: "idempotency_key", Values: []any{value}}}, limit, offset)
}

// CreateCandidate menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreateCandidate(ctx context.Context, in CreateCandidateInput) (*entity.Candidate, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.candidate.Create(ctx, func() (*entity.Candidate, error) { return createCandidate(in) })
}

// UpdateCandidate menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdateCandidate(ctx context.Context, key string, in UpdateCandidateInput) (*entity.Candidate, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.candidate.Update(ctx, key, func(value *entity.Candidate) error { return patchCandidate(value, in) })
}

// DeleteCandidate menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeleteCandidate(ctx context.Context, key string) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.candidate.Delete(ctx, key)
}

// GetCandidateByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetCandidateByID(ctx context.Context, key string) (*entity.Candidate, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.candidate.Get(ctx, key)
}

// GetAllCandidates menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllCandidates(ctx context.Context, limit, offset int) (operation.Page[entity.Candidate], error) {
	if s == nil {
		return operation.Page[entity.Candidate]{}, hrerrors.ErrUnavailable
	}
	return s.candidate.List(ctx, nil, limit, offset)
}

// FindCandidateByCandidateNumber menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindCandidateByCandidateNumber(ctx context.Context, value string, limit, offset int) (operation.Page[entity.Candidate], error) {
	if s == nil {
		return operation.Page[entity.Candidate]{}, hrerrors.ErrUnavailable
	}
	return s.candidate.List(ctx, []domainrepo.Filter{{Column: "candidate_number", Values: []any{value}}}, limit, offset)
}

// FindCandidateByPersonPartyReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindCandidateByPersonPartyReferenceID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.Candidate], error) {
	if s == nil {
		return operation.Page[entity.Candidate]{}, hrerrors.ErrUnavailable
	}
	return s.candidate.List(ctx, []domainrepo.Filter{{Column: "person_party_reference_id", Values: []any{value}}}, limit, offset)
}

// FindCandidateByContactReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindCandidateByContactReferenceID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.Candidate], error) {
	if s == nil {
		return operation.Page[entity.Candidate]{}, hrerrors.ErrUnavailable
	}
	return s.candidate.List(ctx, []domainrepo.Filter{{Column: "contact_reference_id", Values: []any{value}}}, limit, offset)
}

// FindCandidateBySourceType menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindCandidateBySourceType(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.Candidate], error) {
	if s == nil {
		return operation.Page[entity.Candidate]{}, hrerrors.ErrUnavailable
	}
	return s.candidate.List(ctx, []domainrepo.Filter{{Column: "source_type", Values: []any{value}}}, limit, offset)
}

// FindCandidateByCandidateStatus menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindCandidateByCandidateStatus(ctx context.Context, value string, limit, offset int) (operation.Page[entity.Candidate], error) {
	if s == nil {
		return operation.Page[entity.Candidate]{}, hrerrors.ErrUnavailable
	}
	return s.candidate.List(ctx, []domainrepo.Filter{{Column: "candidate_status", Values: []any{value}}}, limit, offset)
}

// FindCandidateByCreatedByPrincipalID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindCandidateByCreatedByPrincipalID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.Candidate], error) {
	if s == nil {
		return operation.Page[entity.Candidate]{}, hrerrors.ErrUnavailable
	}
	return s.candidate.List(ctx, []domainrepo.Filter{{Column: "created_by_principal_id", Values: []any{value}}}, limit, offset)
}

// CreateApplication menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreateApplication(ctx context.Context, in CreateApplicationInput) (*entity.Application, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.application.Create(ctx, func() (*entity.Application, error) { return createApplication(in) })
}

// UpdateApplication menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdateApplication(ctx context.Context, key string, in UpdateApplicationInput) (*entity.Application, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.application.Update(ctx, key, func(value *entity.Application) error { return patchApplication(value, in) })
}

// DeleteApplication menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeleteApplication(ctx context.Context, key string) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.application.Delete(ctx, key)
}

// GetApplicationByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetApplicationByID(ctx context.Context, key string) (*entity.Application, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.application.Get(ctx, key)
}

// GetAllApplications menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllApplications(ctx context.Context, limit, offset int) (operation.Page[entity.Application], error) {
	if s == nil {
		return operation.Page[entity.Application]{}, hrerrors.ErrUnavailable
	}
	return s.application.List(ctx, nil, limit, offset)
}

// FindApplicationByCandidateID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindApplicationByCandidateID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.Application], error) {
	if s == nil {
		return operation.Page[entity.Application]{}, hrerrors.ErrUnavailable
	}
	return s.application.List(ctx, []domainrepo.Filter{{Column: "candidate_id", Values: []any{value}}}, limit, offset)
}

// FindApplicationByRequisitionID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindApplicationByRequisitionID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.Application], error) {
	if s == nil {
		return operation.Page[entity.Application]{}, hrerrors.ErrUnavailable
	}
	return s.application.List(ctx, []domainrepo.Filter{{Column: "requisition_id", Values: []any{value}}}, limit, offset)
}

// FindApplicationByApplicationNumber menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindApplicationByApplicationNumber(ctx context.Context, value string, limit, offset int) (operation.Page[entity.Application], error) {
	if s == nil {
		return operation.Page[entity.Application]{}, hrerrors.ErrUnavailable
	}
	return s.application.List(ctx, []domainrepo.Filter{{Column: "application_number", Values: []any{value}}}, limit, offset)
}

// FindApplicationByStatus menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindApplicationByStatus(ctx context.Context, value entity.ApplicationStatus, limit, offset int) (operation.Page[entity.Application], error) {
	if s == nil {
		return operation.Page[entity.Application]{}, hrerrors.ErrUnavailable
	}
	return s.application.List(ctx, []domainrepo.Filter{{Column: "status", Values: []any{value}}}, limit, offset)
}

// FindApplicationByCurrentStageCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindApplicationByCurrentStageCode(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.Application], error) {
	if s == nil {
		return operation.Page[entity.Application]{}, hrerrors.ErrUnavailable
	}
	return s.application.List(ctx, []domainrepo.Filter{{Column: "current_stage_code", Values: []any{value}}}, limit, offset)
}

// FindApplicationByAssignedRecruiterPrincipalID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindApplicationByAssignedRecruiterPrincipalID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.Application], error) {
	if s == nil {
		return operation.Page[entity.Application]{}, hrerrors.ErrUnavailable
	}
	return s.application.List(ctx, []domainrepo.Filter{{Column: "assigned_recruiter_principal_id", Values: []any{value}}}, limit, offset)
}

// FindApplicationByDispositionReasonCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindApplicationByDispositionReasonCode(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.Application], error) {
	if s == nil {
		return operation.Page[entity.Application]{}, hrerrors.ErrUnavailable
	}
	return s.application.List(ctx, []domainrepo.Filter{{Column: "disposition_reason_code", Values: []any{value}}}, limit, offset)
}

// FindApplicationByDecisionByPrincipalID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindApplicationByDecisionByPrincipalID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.Application], error) {
	if s == nil {
		return operation.Page[entity.Application]{}, hrerrors.ErrUnavailable
	}
	return s.application.List(ctx, []domainrepo.Filter{{Column: "decision_by_principal_id", Values: []any{value}}}, limit, offset)
}

// FindApplicationByHiredWorkerID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindApplicationByHiredWorkerID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.Application], error) {
	if s == nil {
		return operation.Page[entity.Application]{}, hrerrors.ErrUnavailable
	}
	return s.application.List(ctx, []domainrepo.Filter{{Column: "hired_worker_id", Values: []any{value}}}, limit, offset)
}

// FindApplicationBySourcePostingID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindApplicationBySourcePostingID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.Application], error) {
	if s == nil {
		return operation.Page[entity.Application]{}, hrerrors.ErrUnavailable
	}
	return s.application.List(ctx, []domainrepo.Filter{{Column: "source_posting_id", Values: []any{value}}}, limit, offset)
}

// FindApplicationByIdempotencyKey menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindApplicationByIdempotencyKey(ctx context.Context, value string, limit, offset int) (operation.Page[entity.Application], error) {
	if s == nil {
		return operation.Page[entity.Application]{}, hrerrors.ErrUnavailable
	}
	return s.application.List(ctx, []domainrepo.Filter{{Column: "idempotency_key", Values: []any{value}}}, limit, offset)
}

// CreateApplicationStageHistory menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreateApplicationStageHistory(ctx context.Context, in CreateApplicationStageHistoryInput) (*entity.ApplicationStageHistory, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.applicationStageHistory.Create(ctx, func() (*entity.ApplicationStageHistory, error) { return createApplicationStageHistory(in) })
}

// UpdateApplicationStageHistory menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdateApplicationStageHistory(ctx context.Context, key int64, in UpdateApplicationStageHistoryInput) (*entity.ApplicationStageHistory, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.applicationStageHistory.Update(ctx, key, func(value *entity.ApplicationStageHistory) error { return patchApplicationStageHistory(value, in) })
}

// DeleteApplicationStageHistory menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeleteApplicationStageHistory(ctx context.Context, key int64) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.applicationStageHistory.Delete(ctx, key)
}

// GetApplicationStageHistoryByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetApplicationStageHistoryByID(ctx context.Context, key int64) (*entity.ApplicationStageHistory, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.applicationStageHistory.Get(ctx, key)
}

// GetAllApplicationStageHistory menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllApplicationStageHistory(ctx context.Context, limit, offset int) (operation.Page[entity.ApplicationStageHistory], error) {
	if s == nil {
		return operation.Page[entity.ApplicationStageHistory]{}, hrerrors.ErrUnavailable
	}
	return s.applicationStageHistory.List(ctx, nil, limit, offset)
}

// FindApplicationStageHistoryByApplicationID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindApplicationStageHistoryByApplicationID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.ApplicationStageHistory], error) {
	if s == nil {
		return operation.Page[entity.ApplicationStageHistory]{}, hrerrors.ErrUnavailable
	}
	return s.applicationStageHistory.List(ctx, []domainrepo.Filter{{Column: "application_id", Values: []any{value}}}, limit, offset)
}

// FindApplicationStageHistoryByPreviousStatus menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindApplicationStageHistoryByPreviousStatus(ctx context.Context, value *entity.ApplicationStatus, limit, offset int) (operation.Page[entity.ApplicationStageHistory], error) {
	if s == nil {
		return operation.Page[entity.ApplicationStageHistory]{}, hrerrors.ErrUnavailable
	}
	return s.applicationStageHistory.List(ctx, []domainrepo.Filter{{Column: "previous_status", Values: []any{value}}}, limit, offset)
}

// FindApplicationStageHistoryByNewStatus menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindApplicationStageHistoryByNewStatus(ctx context.Context, value entity.ApplicationStatus, limit, offset int) (operation.Page[entity.ApplicationStageHistory], error) {
	if s == nil {
		return operation.Page[entity.ApplicationStageHistory]{}, hrerrors.ErrUnavailable
	}
	return s.applicationStageHistory.List(ctx, []domainrepo.Filter{{Column: "new_status", Values: []any{value}}}, limit, offset)
}

// FindApplicationStageHistoryByStageCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindApplicationStageHistoryByStageCode(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.ApplicationStageHistory], error) {
	if s == nil {
		return operation.Page[entity.ApplicationStageHistory]{}, hrerrors.ErrUnavailable
	}
	return s.applicationStageHistory.List(ctx, []domainrepo.Filter{{Column: "stage_code", Values: []any{value}}}, limit, offset)
}

// FindApplicationStageHistoryByChangedByPrincipalID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindApplicationStageHistoryByChangedByPrincipalID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.ApplicationStageHistory], error) {
	if s == nil {
		return operation.Page[entity.ApplicationStageHistory]{}, hrerrors.ErrUnavailable
	}
	return s.applicationStageHistory.List(ctx, []domainrepo.Filter{{Column: "changed_by_principal_id", Values: []any{value}}}, limit, offset)
}

// FindApplicationStageHistoryByReasonCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindApplicationStageHistoryByReasonCode(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.ApplicationStageHistory], error) {
	if s == nil {
		return operation.Page[entity.ApplicationStageHistory]{}, hrerrors.ErrUnavailable
	}
	return s.applicationStageHistory.List(ctx, []domainrepo.Filter{{Column: "reason_code", Values: []any{value}}}, limit, offset)
}

// CreateInterview menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreateInterview(ctx context.Context, in CreateInterviewInput) (*entity.Interview, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.interview.Create(ctx, func() (*entity.Interview, error) { return createInterview(in) })
}

// UpdateInterview menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdateInterview(ctx context.Context, key string, in UpdateInterviewInput) (*entity.Interview, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.interview.Update(ctx, key, func(value *entity.Interview) error { return patchInterview(value, in) })
}

// DeleteInterview menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeleteInterview(ctx context.Context, key string) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.interview.Delete(ctx, key)
}

// GetInterviewByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetInterviewByID(ctx context.Context, key string) (*entity.Interview, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.interview.Get(ctx, key)
}

// GetAllInterviews menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllInterviews(ctx context.Context, limit, offset int) (operation.Page[entity.Interview], error) {
	if s == nil {
		return operation.Page[entity.Interview]{}, hrerrors.ErrUnavailable
	}
	return s.interview.List(ctx, nil, limit, offset)
}

// FindInterviewByApplicationID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindInterviewByApplicationID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.Interview], error) {
	if s == nil {
		return operation.Page[entity.Interview]{}, hrerrors.ErrUnavailable
	}
	return s.interview.List(ctx, []domainrepo.Filter{{Column: "application_id", Values: []any{value}}}, limit, offset)
}

// FindInterviewByInterviewType menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindInterviewByInterviewType(ctx context.Context, value string, limit, offset int) (operation.Page[entity.Interview], error) {
	if s == nil {
		return operation.Page[entity.Interview]{}, hrerrors.ErrUnavailable
	}
	return s.interview.List(ctx, []domainrepo.Filter{{Column: "interview_type", Values: []any{value}}}, limit, offset)
}

// FindInterviewByLocationReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindInterviewByLocationReferenceID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.Interview], error) {
	if s == nil {
		return operation.Page[entity.Interview]{}, hrerrors.ErrUnavailable
	}
	return s.interview.List(ctx, []domainrepo.Filter{{Column: "location_reference_id", Values: []any{value}}}, limit, offset)
}

// FindInterviewByMeetingReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindInterviewByMeetingReferenceID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.Interview], error) {
	if s == nil {
		return operation.Page[entity.Interview]{}, hrerrors.ErrUnavailable
	}
	return s.interview.List(ctx, []domainrepo.Filter{{Column: "meeting_reference_id", Values: []any{value}}}, limit, offset)
}

// FindInterviewByStatus menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindInterviewByStatus(ctx context.Context, value string, limit, offset int) (operation.Page[entity.Interview], error) {
	if s == nil {
		return operation.Page[entity.Interview]{}, hrerrors.ErrUnavailable
	}
	return s.interview.List(ctx, []domainrepo.Filter{{Column: "status", Values: []any{value}}}, limit, offset)
}

// FindInterviewByCreatedByPrincipalID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindInterviewByCreatedByPrincipalID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.Interview], error) {
	if s == nil {
		return operation.Page[entity.Interview]{}, hrerrors.ErrUnavailable
	}
	return s.interview.List(ctx, []domainrepo.Filter{{Column: "created_by_principal_id", Values: []any{value}}}, limit, offset)
}

// CreateInterviewPanelist menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreateInterviewPanelist(ctx context.Context, in CreateInterviewPanelistInput) (*entity.InterviewPanelist, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.interviewPanelist.Create(ctx, func() (*entity.InterviewPanelist, error) { return createInterviewPanelist(in) })
}

// UpdateInterviewPanelist menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdateInterviewPanelist(ctx context.Context, key string, in UpdateInterviewPanelistInput) (*entity.InterviewPanelist, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.interviewPanelist.Update(ctx, key, func(value *entity.InterviewPanelist) error { return patchInterviewPanelist(value, in) })
}

// DeleteInterviewPanelist menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeleteInterviewPanelist(ctx context.Context, key string) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.interviewPanelist.Delete(ctx, key)
}

// GetInterviewPanelistByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetInterviewPanelistByID(ctx context.Context, key string) (*entity.InterviewPanelist, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.interviewPanelist.Get(ctx, key)
}

// GetAllInterviewPanelists menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllInterviewPanelists(ctx context.Context, limit, offset int) (operation.Page[entity.InterviewPanelist], error) {
	if s == nil {
		return operation.Page[entity.InterviewPanelist]{}, hrerrors.ErrUnavailable
	}
	return s.interviewPanelist.List(ctx, nil, limit, offset)
}

// FindInterviewPanelistByInterviewID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindInterviewPanelistByInterviewID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.InterviewPanelist], error) {
	if s == nil {
		return operation.Page[entity.InterviewPanelist]{}, hrerrors.ErrUnavailable
	}
	return s.interviewPanelist.List(ctx, []domainrepo.Filter{{Column: "interview_id", Values: []any{value}}}, limit, offset)
}

// FindInterviewPanelistByInterviewerPrincipalID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindInterviewPanelistByInterviewerPrincipalID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.InterviewPanelist], error) {
	if s == nil {
		return operation.Page[entity.InterviewPanelist]{}, hrerrors.ErrUnavailable
	}
	return s.interviewPanelist.List(ctx, []domainrepo.Filter{{Column: "interviewer_principal_id", Values: []any{value}}}, limit, offset)
}

// FindInterviewPanelistByScorecardTemplateReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindInterviewPanelistByScorecardTemplateReferenceID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.InterviewPanelist], error) {
	if s == nil {
		return operation.Page[entity.InterviewPanelist]{}, hrerrors.ErrUnavailable
	}
	return s.interviewPanelist.List(ctx, []domainrepo.Filter{{Column: "scorecard_template_reference_id", Values: []any{value}}}, limit, offset)
}

// CreateInterviewFeedback menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreateInterviewFeedback(ctx context.Context, in CreateInterviewFeedbackInput) (*entity.InterviewFeedback, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.interviewFeedback.Create(ctx, func() (*entity.InterviewFeedback, error) { return createInterviewFeedback(in) })
}

// UpdateInterviewFeedback menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdateInterviewFeedback(ctx context.Context, key string, in UpdateInterviewFeedbackInput) (*entity.InterviewFeedback, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.interviewFeedback.Update(ctx, key, func(value *entity.InterviewFeedback) error { return patchInterviewFeedback(value, in) })
}

// DeleteInterviewFeedback menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeleteInterviewFeedback(ctx context.Context, key string) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.interviewFeedback.Delete(ctx, key)
}

// GetInterviewFeedbackByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetInterviewFeedbackByID(ctx context.Context, key string) (*entity.InterviewFeedback, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.interviewFeedback.Get(ctx, key)
}

// GetAllInterviewFeedback menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllInterviewFeedback(ctx context.Context, limit, offset int) (operation.Page[entity.InterviewFeedback], error) {
	if s == nil {
		return operation.Page[entity.InterviewFeedback]{}, hrerrors.ErrUnavailable
	}
	return s.interviewFeedback.List(ctx, nil, limit, offset)
}

// FindInterviewFeedbackByInterviewID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindInterviewFeedbackByInterviewID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.InterviewFeedback], error) {
	if s == nil {
		return operation.Page[entity.InterviewFeedback]{}, hrerrors.ErrUnavailable
	}
	return s.interviewFeedback.List(ctx, []domainrepo.Filter{{Column: "interview_id", Values: []any{value}}}, limit, offset)
}

// FindInterviewFeedbackByInterviewerPrincipalID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindInterviewFeedbackByInterviewerPrincipalID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.InterviewFeedback], error) {
	if s == nil {
		return operation.Page[entity.InterviewFeedback]{}, hrerrors.ErrUnavailable
	}
	return s.interviewFeedback.List(ctx, []domainrepo.Filter{{Column: "interviewer_principal_id", Values: []any{value}}}, limit, offset)
}

// CreateOffer menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreateOffer(ctx context.Context, in CreateOfferInput) (*entity.Offer, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.offer.Create(ctx, func() (*entity.Offer, error) { return createOffer(in) })
}

// UpdateOffer menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdateOffer(ctx context.Context, key string, in UpdateOfferInput) (*entity.Offer, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.offer.Update(ctx, key, func(value *entity.Offer) error { return patchOffer(value, in) })
}

// DeleteOffer menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeleteOffer(ctx context.Context, key string) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.offer.Delete(ctx, key)
}

// GetOfferByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetOfferByID(ctx context.Context, key string) (*entity.Offer, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.offer.Get(ctx, key)
}

// GetAllOffers menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllOffers(ctx context.Context, limit, offset int) (operation.Page[entity.Offer], error) {
	if s == nil {
		return operation.Page[entity.Offer]{}, hrerrors.ErrUnavailable
	}
	return s.offer.List(ctx, nil, limit, offset)
}

// FindOfferByApplicationID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindOfferByApplicationID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.Offer], error) {
	if s == nil {
		return operation.Page[entity.Offer]{}, hrerrors.ErrUnavailable
	}
	return s.offer.List(ctx, []domainrepo.Filter{{Column: "application_id", Values: []any{value}}}, limit, offset)
}

// FindOfferByOfferNumber menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindOfferByOfferNumber(ctx context.Context, value string, limit, offset int) (operation.Page[entity.Offer], error) {
	if s == nil {
		return operation.Page[entity.Offer]{}, hrerrors.ErrUnavailable
	}
	return s.offer.List(ctx, []domainrepo.Filter{{Column: "offer_number", Values: []any{value}}}, limit, offset)
}

// FindOfferByOfferStatus menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindOfferByOfferStatus(ctx context.Context, value string, limit, offset int) (operation.Page[entity.Offer], error) {
	if s == nil {
		return operation.Page[entity.Offer]{}, hrerrors.ErrUnavailable
	}
	return s.offer.List(ctx, []domainrepo.Filter{{Column: "offer_status", Values: []any{value}}}, limit, offset)
}

// FindOfferByProposedJobID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindOfferByProposedJobID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.Offer], error) {
	if s == nil {
		return operation.Page[entity.Offer]{}, hrerrors.ErrUnavailable
	}
	return s.offer.List(ctx, []domainrepo.Filter{{Column: "proposed_job_id", Values: []any{value}}}, limit, offset)
}

// FindOfferByProposedLegalEntityReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindOfferByProposedLegalEntityReferenceID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.Offer], error) {
	if s == nil {
		return operation.Page[entity.Offer]{}, hrerrors.ErrUnavailable
	}
	return s.offer.List(ctx, []domainrepo.Filter{{Column: "proposed_legal_entity_reference_id", Values: []any{value}}}, limit, offset)
}

// FindOfferByProposedContractType menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindOfferByProposedContractType(ctx context.Context, value string, limit, offset int) (operation.Page[entity.Offer], error) {
	if s == nil {
		return operation.Page[entity.Offer]{}, hrerrors.ErrUnavailable
	}
	return s.offer.List(ctx, []domainrepo.Filter{{Column: "proposed_contract_type", Values: []any{value}}}, limit, offset)
}

// FindOfferByProposedCompensationDocumentVersionReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindOfferByProposedCompensationDocumentVersionReferenceID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.Offer], error) {
	if s == nil {
		return operation.Page[entity.Offer]{}, hrerrors.ErrUnavailable
	}
	return s.offer.List(ctx, []domainrepo.Filter{{Column: "proposed_compensation_document_version_reference_id", Values: []any{value}}}, limit, offset)
}

// FindOfferByCurrencyCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindOfferByCurrencyCode(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.Offer], error) {
	if s == nil {
		return operation.Page[entity.Offer]{}, hrerrors.ErrUnavailable
	}
	return s.offer.List(ctx, []domainrepo.Filter{{Column: "currency_code", Values: []any{value}}}, limit, offset)
}

// FindOfferByOfferDocumentReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindOfferByOfferDocumentReferenceID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.Offer], error) {
	if s == nil {
		return operation.Page[entity.Offer]{}, hrerrors.ErrUnavailable
	}
	return s.offer.List(ctx, []domainrepo.Filter{{Column: "offer_document_reference_id", Values: []any{value}}}, limit, offset)
}

// FindOfferByOfferDocumentVersionReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindOfferByOfferDocumentVersionReferenceID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.Offer], error) {
	if s == nil {
		return operation.Page[entity.Offer]{}, hrerrors.ErrUnavailable
	}
	return s.offer.List(ctx, []domainrepo.Filter{{Column: "offer_document_version_reference_id", Values: []any{value}}}, limit, offset)
}

// FindOfferByApprovedByPrincipalID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindOfferByApprovedByPrincipalID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.Offer], error) {
	if s == nil {
		return operation.Page[entity.Offer]{}, hrerrors.ErrUnavailable
	}
	return s.offer.List(ctx, []domainrepo.Filter{{Column: "approved_by_principal_id", Values: []any{value}}}, limit, offset)
}

// CreateJobPosting menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreateJobPosting(ctx context.Context, in CreateJobPostingInput) (*entity.JobPosting, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.jobPosting.Create(ctx, func() (*entity.JobPosting, error) { return createJobPosting(in) })
}

// UpdateJobPosting menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdateJobPosting(ctx context.Context, key string, in UpdateJobPostingInput) (*entity.JobPosting, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.jobPosting.Update(ctx, key, func(value *entity.JobPosting) error { return patchJobPosting(value, in) })
}

// DeleteJobPosting menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeleteJobPosting(ctx context.Context, key string) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.jobPosting.Delete(ctx, key)
}

// GetJobPostingByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetJobPostingByID(ctx context.Context, key string) (*entity.JobPosting, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.jobPosting.Get(ctx, key)
}

// GetAllJobPostings menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllJobPostings(ctx context.Context, limit, offset int) (operation.Page[entity.JobPosting], error) {
	if s == nil {
		return operation.Page[entity.JobPosting]{}, hrerrors.ErrUnavailable
	}
	return s.jobPosting.List(ctx, nil, limit, offset)
}

// FindJobPostingByRequisitionID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindJobPostingByRequisitionID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.JobPosting], error) {
	if s == nil {
		return operation.Page[entity.JobPosting]{}, hrerrors.ErrUnavailable
	}
	return s.jobPosting.List(ctx, []domainrepo.Filter{{Column: "requisition_id", Values: []any{value}}}, limit, offset)
}

// FindJobPostingByChannelCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindJobPostingByChannelCode(ctx context.Context, value string, limit, offset int) (operation.Page[entity.JobPosting], error) {
	if s == nil {
		return operation.Page[entity.JobPosting]{}, hrerrors.ErrUnavailable
	}
	return s.jobPosting.List(ctx, []domainrepo.Filter{{Column: "channel_code", Values: []any{value}}}, limit, offset)
}

// FindJobPostingByContentDocumentVersionReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindJobPostingByContentDocumentVersionReferenceID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.JobPosting], error) {
	if s == nil {
		return operation.Page[entity.JobPosting]{}, hrerrors.ErrUnavailable
	}
	return s.jobPosting.List(ctx, []domainrepo.Filter{{Column: "content_document_version_reference_id", Values: []any{value}}}, limit, offset)
}

// FindJobPostingByStatus menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindJobPostingByStatus(ctx context.Context, value string, limit, offset int) (operation.Page[entity.JobPosting], error) {
	if s == nil {
		return operation.Page[entity.JobPosting]{}, hrerrors.ErrUnavailable
	}
	return s.jobPosting.List(ctx, []domainrepo.Filter{{Column: "status", Values: []any{value}}}, limit, offset)
}

// FindJobPostingByCreatedByPrincipalID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindJobPostingByCreatedByPrincipalID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.JobPosting], error) {
	if s == nil {
		return operation.Page[entity.JobPosting]{}, hrerrors.ErrUnavailable
	}
	return s.jobPosting.List(ctx, []domainrepo.Filter{{Column: "created_by_principal_id", Values: []any{value}}}, limit, offset)
}

// CreateApplicationAssessment menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreateApplicationAssessment(ctx context.Context, in CreateApplicationAssessmentInput) (*entity.ApplicationAssessment, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.applicationAssessment.Create(ctx, func() (*entity.ApplicationAssessment, error) { return createApplicationAssessment(in) })
}

// UpdateApplicationAssessment menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdateApplicationAssessment(ctx context.Context, key string, in UpdateApplicationAssessmentInput) (*entity.ApplicationAssessment, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.applicationAssessment.Update(ctx, key, func(value *entity.ApplicationAssessment) error { return patchApplicationAssessment(value, in) })
}

// DeleteApplicationAssessment menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeleteApplicationAssessment(ctx context.Context, key string) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.applicationAssessment.Delete(ctx, key)
}

// GetApplicationAssessmentByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetApplicationAssessmentByID(ctx context.Context, key string) (*entity.ApplicationAssessment, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.applicationAssessment.Get(ctx, key)
}

// GetAllApplicationAssessments menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllApplicationAssessments(ctx context.Context, limit, offset int) (operation.Page[entity.ApplicationAssessment], error) {
	if s == nil {
		return operation.Page[entity.ApplicationAssessment]{}, hrerrors.ErrUnavailable
	}
	return s.applicationAssessment.List(ctx, nil, limit, offset)
}

// FindApplicationAssessmentByApplicationID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindApplicationAssessmentByApplicationID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.ApplicationAssessment], error) {
	if s == nil {
		return operation.Page[entity.ApplicationAssessment]{}, hrerrors.ErrUnavailable
	}
	return s.applicationAssessment.List(ctx, []domainrepo.Filter{{Column: "application_id", Values: []any{value}}}, limit, offset)
}

// FindApplicationAssessmentByAssessmentCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindApplicationAssessmentByAssessmentCode(ctx context.Context, value string, limit, offset int) (operation.Page[entity.ApplicationAssessment], error) {
	if s == nil {
		return operation.Page[entity.ApplicationAssessment]{}, hrerrors.ErrUnavailable
	}
	return s.applicationAssessment.List(ctx, []domainrepo.Filter{{Column: "assessment_code", Values: []any{value}}}, limit, offset)
}

// FindApplicationAssessmentByAttemptNumber menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindApplicationAssessmentByAttemptNumber(ctx context.Context, value int32, limit, offset int) (operation.Page[entity.ApplicationAssessment], error) {
	if s == nil {
		return operation.Page[entity.ApplicationAssessment]{}, hrerrors.ErrUnavailable
	}
	return s.applicationAssessment.List(ctx, []domainrepo.Filter{{Column: "attempt_number", Values: []any{value}}}, limit, offset)
}

// FindApplicationAssessmentByRubricDocumentVersionReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindApplicationAssessmentByRubricDocumentVersionReferenceID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.ApplicationAssessment], error) {
	if s == nil {
		return operation.Page[entity.ApplicationAssessment]{}, hrerrors.ErrUnavailable
	}
	return s.applicationAssessment.List(ctx, []domainrepo.Filter{{Column: "rubric_document_version_reference_id", Values: []any{value}}}, limit, offset)
}

// FindApplicationAssessmentByStatus menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindApplicationAssessmentByStatus(ctx context.Context, value string, limit, offset int) (operation.Page[entity.ApplicationAssessment], error) {
	if s == nil {
		return operation.Page[entity.ApplicationAssessment]{}, hrerrors.ErrUnavailable
	}
	return s.applicationAssessment.List(ctx, []domainrepo.Filter{{Column: "status", Values: []any{value}}}, limit, offset)
}

// FindApplicationAssessmentByAssessorPrincipalID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindApplicationAssessmentByAssessorPrincipalID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.ApplicationAssessment], error) {
	if s == nil {
		return operation.Page[entity.ApplicationAssessment]{}, hrerrors.ErrUnavailable
	}
	return s.applicationAssessment.List(ctx, []domainrepo.Filter{{Column: "assessor_principal_id", Values: []any{value}}}, limit, offset)
}

// FindApplicationAssessmentByRecommendationCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindApplicationAssessmentByRecommendationCode(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.ApplicationAssessment], error) {
	if s == nil {
		return operation.Page[entity.ApplicationAssessment]{}, hrerrors.ErrUnavailable
	}
	return s.applicationAssessment.List(ctx, []domainrepo.Filter{{Column: "recommendation_code", Values: []any{value}}}, limit, offset)
}

// FindApplicationAssessmentByEvidenceDocumentVersionReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindApplicationAssessmentByEvidenceDocumentVersionReferenceID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.ApplicationAssessment], error) {
	if s == nil {
		return operation.Page[entity.ApplicationAssessment]{}, hrerrors.ErrUnavailable
	}
	return s.applicationAssessment.List(ctx, []domainrepo.Filter{{Column: "evidence_document_version_reference_id", Values: []any{value}}}, limit, offset)
}
