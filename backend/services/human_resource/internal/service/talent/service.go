// Kerangka awal dibuat oleh alat bantu; selanjutnya dipelihara langsung per domain.
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
	PerformanceReviewReader   ReviewReader
	Runtime                   operation.Dependencies
	TalentFramework           operation.Repository[entity.TalentFramework, string]
	TalentDimension           operation.Repository[entity.TalentDimension, string]
	TalentReviewCycle         operation.Repository[entity.TalentReviewCycle, string]
	TalentAssessment          operation.Repository[entity.TalentAssessment, string]
	TalentAssessmentScore     operation.Repository[entity.TalentAssessmentScore, string]
	TalentCalibrationSession  operation.Repository[entity.TalentCalibrationSession, string]
	TalentCalibrationDecision operation.Repository[entity.TalentCalibrationDecision, string]
	TalentPool                operation.Repository[entity.TalentPool, string]
	TalentPoolMembership      operation.Repository[entity.TalentPoolMembership, string]
}

// Service mengoordinasikan operasi domain talent.
type Service struct {
	dependencies              Dependencies
	talentFramework           *operation.Resource[entity.TalentFramework, string, *entity.TalentFramework]
	talentDimension           *operation.Resource[entity.TalentDimension, string, *entity.TalentDimension]
	talentReviewCycle         *operation.Resource[entity.TalentReviewCycle, string, *entity.TalentReviewCycle]
	talentAssessment          *operation.Resource[entity.TalentAssessment, string, *entity.TalentAssessment]
	talentAssessmentScore     *operation.Resource[entity.TalentAssessmentScore, string, *entity.TalentAssessmentScore]
	talentCalibrationSession  *operation.Resource[entity.TalentCalibrationSession, string, *entity.TalentCalibrationSession]
	talentCalibrationDecision *operation.Resource[entity.TalentCalibrationDecision, string, *entity.TalentCalibrationDecision]
	talentPool                *operation.Resource[entity.TalentPool, string, *entity.TalentPool]
	talentPoolMembership      *operation.Resource[entity.TalentPoolMembership, string, *entity.TalentPoolMembership]
}

// NewService memeriksa seluruh port dan dependency sebelum service dapat dipakai.
func NewService(deps Dependencies) (*Service, error) {
	s := &Service{dependencies: deps}
	var err error
	s.talentFramework, err = operation.NewResource[entity.TalentFramework, string, *entity.TalentFramework](deps.TalentFramework, deps.Runtime, specTalentFramework())
	if err != nil {
		return nil, err
	}
	s.talentDimension, err = operation.NewResource[entity.TalentDimension, string, *entity.TalentDimension](deps.TalentDimension, deps.Runtime, specTalentDimension())
	if err != nil {
		return nil, err
	}
	s.talentReviewCycle, err = operation.NewResource[entity.TalentReviewCycle, string, *entity.TalentReviewCycle](deps.TalentReviewCycle, deps.Runtime, specTalentReviewCycle())
	if err != nil {
		return nil, err
	}
	s.talentAssessment, err = operation.NewResource[entity.TalentAssessment, string, *entity.TalentAssessment](deps.TalentAssessment, deps.Runtime, specTalentAssessment())
	if err != nil {
		return nil, err
	}
	s.talentAssessmentScore, err = operation.NewResource[entity.TalentAssessmentScore, string, *entity.TalentAssessmentScore](deps.TalentAssessmentScore, deps.Runtime, specTalentAssessmentScore())
	if err != nil {
		return nil, err
	}
	s.talentCalibrationSession, err = operation.NewResource[entity.TalentCalibrationSession, string, *entity.TalentCalibrationSession](deps.TalentCalibrationSession, deps.Runtime, specTalentCalibrationSession())
	if err != nil {
		return nil, err
	}
	s.talentCalibrationDecision, err = operation.NewResource[entity.TalentCalibrationDecision, string, *entity.TalentCalibrationDecision](deps.TalentCalibrationDecision, deps.Runtime, specTalentCalibrationDecision())
	if err != nil {
		return nil, err
	}
	s.talentPool, err = operation.NewResource[entity.TalentPool, string, *entity.TalentPool](deps.TalentPool, deps.Runtime, specTalentPool())
	if err != nil {
		return nil, err
	}
	s.talentPoolMembership, err = operation.NewResource[entity.TalentPoolMembership, string, *entity.TalentPoolMembership](deps.TalentPoolMembership, deps.Runtime, specTalentPoolMembership())
	if err != nil {
		return nil, err
	}
	return s, nil
}

// CreateTalentFramework menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreateTalentFramework(ctx context.Context, in CreateTalentFrameworkInput) (*entity.TalentFramework, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.talentFramework.Create(ctx, func() (*entity.TalentFramework, error) { return createTalentFramework(in) })
}

// UpdateTalentFramework menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdateTalentFramework(ctx context.Context, key string, in UpdateTalentFrameworkInput) (*entity.TalentFramework, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.talentFramework.Update(ctx, key, func(value *entity.TalentFramework) error { return patchTalentFramework(value, in) })
}

// DeleteTalentFramework menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeleteTalentFramework(ctx context.Context, key string) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.talentFramework.Delete(ctx, key)
}

// GetTalentFrameworkByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetTalentFrameworkByID(ctx context.Context, key string) (*entity.TalentFramework, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.talentFramework.Get(ctx, key)
}

// GetAllTalentFrameworks menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllTalentFrameworks(ctx context.Context, limit, offset int) (operation.Page[entity.TalentFramework], error) {
	if s == nil {
		return operation.Page[entity.TalentFramework]{}, hrerrors.ErrUnavailable
	}
	return s.talentFramework.List(ctx, nil, limit, offset)
}

// FindTalentFrameworkByFrameworkCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindTalentFrameworkByFrameworkCode(ctx context.Context, value string, limit, offset int) (operation.Page[entity.TalentFramework], error) {
	if s == nil {
		return operation.Page[entity.TalentFramework]{}, hrerrors.ErrUnavailable
	}
	return s.talentFramework.List(ctx, []domainrepo.Filter{{Column: "framework_code", Values: []any{value}}}, limit, offset)
}

// FindTalentFrameworkByRubricDocumentVersionReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindTalentFrameworkByRubricDocumentVersionReferenceID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.TalentFramework], error) {
	if s == nil {
		return operation.Page[entity.TalentFramework]{}, hrerrors.ErrUnavailable
	}
	return s.talentFramework.List(ctx, []domainrepo.Filter{{Column: "rubric_document_version_reference_id", Values: []any{value}}}, limit, offset)
}

// FindTalentFrameworkByApprovedByPrincipalID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindTalentFrameworkByApprovedByPrincipalID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.TalentFramework], error) {
	if s == nil {
		return operation.Page[entity.TalentFramework]{}, hrerrors.ErrUnavailable
	}
	return s.talentFramework.List(ctx, []domainrepo.Filter{{Column: "approved_by_principal_id", Values: []any{value}}}, limit, offset)
}

// CreateTalentDimension menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreateTalentDimension(ctx context.Context, in CreateTalentDimensionInput) (*entity.TalentDimension, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.talentDimension.Create(ctx, func() (*entity.TalentDimension, error) { return createTalentDimension(in) })
}

// UpdateTalentDimension menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdateTalentDimension(ctx context.Context, key string, in UpdateTalentDimensionInput) (*entity.TalentDimension, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.talentDimension.Update(ctx, key, func(value *entity.TalentDimension) error { return patchTalentDimension(value, in) })
}

// DeleteTalentDimension menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeleteTalentDimension(ctx context.Context, key string) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.talentDimension.Delete(ctx, key)
}

// GetTalentDimensionByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetTalentDimensionByID(ctx context.Context, key string) (*entity.TalentDimension, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.talentDimension.Get(ctx, key)
}

// GetAllTalentDimensions menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllTalentDimensions(ctx context.Context, limit, offset int) (operation.Page[entity.TalentDimension], error) {
	if s == nil {
		return operation.Page[entity.TalentDimension]{}, hrerrors.ErrUnavailable
	}
	return s.talentDimension.List(ctx, nil, limit, offset)
}

// FindTalentDimensionByFrameworkID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindTalentDimensionByFrameworkID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.TalentDimension], error) {
	if s == nil {
		return operation.Page[entity.TalentDimension]{}, hrerrors.ErrUnavailable
	}
	return s.talentDimension.List(ctx, []domainrepo.Filter{{Column: "framework_id", Values: []any{value}}}, limit, offset)
}

// FindTalentDimensionByDimensionCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindTalentDimensionByDimensionCode(ctx context.Context, value string, limit, offset int) (operation.Page[entity.TalentDimension], error) {
	if s == nil {
		return operation.Page[entity.TalentDimension]{}, hrerrors.ErrUnavailable
	}
	return s.talentDimension.List(ctx, []domainrepo.Filter{{Column: "dimension_code", Values: []any{value}}}, limit, offset)
}

// CreateTalentReviewCycle menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreateTalentReviewCycle(ctx context.Context, in CreateTalentReviewCycleInput) (*entity.TalentReviewCycle, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.talentReviewCycle.Create(ctx, func() (*entity.TalentReviewCycle, error) { return createTalentReviewCycle(in) })
}

// UpdateTalentReviewCycle menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdateTalentReviewCycle(ctx context.Context, key string, in UpdateTalentReviewCycleInput) (*entity.TalentReviewCycle, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.talentReviewCycle.Update(ctx, key, func(value *entity.TalentReviewCycle) error { return patchTalentReviewCycle(value, in) })
}

// DeleteTalentReviewCycle menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeleteTalentReviewCycle(ctx context.Context, key string) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.talentReviewCycle.Delete(ctx, key)
}

// GetTalentReviewCycleByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetTalentReviewCycleByID(ctx context.Context, key string) (*entity.TalentReviewCycle, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.talentReviewCycle.Get(ctx, key)
}

// GetAllTalentReviewCycles menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllTalentReviewCycles(ctx context.Context, limit, offset int) (operation.Page[entity.TalentReviewCycle], error) {
	if s == nil {
		return operation.Page[entity.TalentReviewCycle]{}, hrerrors.ErrUnavailable
	}
	return s.talentReviewCycle.List(ctx, nil, limit, offset)
}

// FindTalentReviewCycleByCycleCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindTalentReviewCycleByCycleCode(ctx context.Context, value string, limit, offset int) (operation.Page[entity.TalentReviewCycle], error) {
	if s == nil {
		return operation.Page[entity.TalentReviewCycle]{}, hrerrors.ErrUnavailable
	}
	return s.talentReviewCycle.List(ctx, []domainrepo.Filter{{Column: "cycle_code", Values: []any{value}}}, limit, offset)
}

// FindTalentReviewCycleByFrameworkID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindTalentReviewCycleByFrameworkID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.TalentReviewCycle], error) {
	if s == nil {
		return operation.Page[entity.TalentReviewCycle]{}, hrerrors.ErrUnavailable
	}
	return s.talentReviewCycle.List(ctx, []domainrepo.Filter{{Column: "framework_id", Values: []any{value}}}, limit, offset)
}

// FindTalentReviewCycleByLegalEntityReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindTalentReviewCycleByLegalEntityReferenceID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.TalentReviewCycle], error) {
	if s == nil {
		return operation.Page[entity.TalentReviewCycle]{}, hrerrors.ErrUnavailable
	}
	return s.talentReviewCycle.List(ctx, []domainrepo.Filter{{Column: "legal_entity_reference_id", Values: []any{value}}}, limit, offset)
}

// FindTalentReviewCycleByOrganizationUnitReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindTalentReviewCycleByOrganizationUnitReferenceID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.TalentReviewCycle], error) {
	if s == nil {
		return operation.Page[entity.TalentReviewCycle]{}, hrerrors.ErrUnavailable
	}
	return s.talentReviewCycle.List(ctx, []domainrepo.Filter{{Column: "organization_unit_reference_id", Values: []any{value}}}, limit, offset)
}

// FindTalentReviewCycleByStatus menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindTalentReviewCycleByStatus(ctx context.Context, value string, limit, offset int) (operation.Page[entity.TalentReviewCycle], error) {
	if s == nil {
		return operation.Page[entity.TalentReviewCycle]{}, hrerrors.ErrUnavailable
	}
	return s.talentReviewCycle.List(ctx, []domainrepo.Filter{{Column: "status", Values: []any{value}}}, limit, offset)
}

// FindTalentReviewCycleByOwnerPrincipalID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindTalentReviewCycleByOwnerPrincipalID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.TalentReviewCycle], error) {
	if s == nil {
		return operation.Page[entity.TalentReviewCycle]{}, hrerrors.ErrUnavailable
	}
	return s.talentReviewCycle.List(ctx, []domainrepo.Filter{{Column: "owner_principal_id", Values: []any{value}}}, limit, offset)
}

// CreateTalentAssessment menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreateTalentAssessment(ctx context.Context, in CreateTalentAssessmentInput) (*entity.TalentAssessment, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.talentAssessment.Create(ctx, func() (*entity.TalentAssessment, error) { return createTalentAssessment(in) })
}

// UpdateTalentAssessment menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdateTalentAssessment(ctx context.Context, key string, in UpdateTalentAssessmentInput) (*entity.TalentAssessment, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.talentAssessment.Update(ctx, key, func(value *entity.TalentAssessment) error { return patchTalentAssessment(value, in) })
}

// DeleteTalentAssessment menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeleteTalentAssessment(ctx context.Context, key string) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.talentAssessment.Delete(ctx, key)
}

// GetTalentAssessmentByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetTalentAssessmentByID(ctx context.Context, key string) (*entity.TalentAssessment, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.talentAssessment.Get(ctx, key)
}

// GetAllTalentAssessments menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllTalentAssessments(ctx context.Context, limit, offset int) (operation.Page[entity.TalentAssessment], error) {
	if s == nil {
		return operation.Page[entity.TalentAssessment]{}, hrerrors.ErrUnavailable
	}
	return s.talentAssessment.List(ctx, nil, limit, offset)
}

// FindTalentAssessmentByWorkerID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindTalentAssessmentByWorkerID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.TalentAssessment], error) {
	if s == nil {
		return operation.Page[entity.TalentAssessment]{}, hrerrors.ErrUnavailable
	}
	return s.talentAssessment.List(ctx, []domainrepo.Filter{{Column: "worker_id", Values: []any{value}}}, limit, offset)
}

// FindTalentAssessmentByTalentReviewCycleID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindTalentAssessmentByTalentReviewCycleID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.TalentAssessment], error) {
	if s == nil {
		return operation.Page[entity.TalentAssessment]{}, hrerrors.ErrUnavailable
	}
	return s.talentAssessment.List(ctx, []domainrepo.Filter{{Column: "talent_review_cycle_id", Values: []any{value}}}, limit, offset)
}

// FindTalentAssessmentByFrameworkID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindTalentAssessmentByFrameworkID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.TalentAssessment], error) {
	if s == nil {
		return operation.Page[entity.TalentAssessment]{}, hrerrors.ErrUnavailable
	}
	return s.talentAssessment.List(ctx, []domainrepo.Filter{{Column: "framework_id", Values: []any{value}}}, limit, offset)
}

// FindTalentAssessmentByPerformanceReviewID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindTalentAssessmentByPerformanceReviewID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.TalentAssessment], error) {
	if s == nil {
		return operation.Page[entity.TalentAssessment]{}, hrerrors.ErrUnavailable
	}
	return s.talentAssessment.List(ctx, []domainrepo.Filter{{Column: "performance_review_id", Values: []any{value}}}, limit, offset)
}

// FindTalentAssessmentByStatus menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindTalentAssessmentByStatus(ctx context.Context, value string, limit, offset int) (operation.Page[entity.TalentAssessment], error) {
	if s == nil {
		return operation.Page[entity.TalentAssessment]{}, hrerrors.ErrUnavailable
	}
	return s.talentAssessment.List(ctx, []domainrepo.Filter{{Column: "status", Values: []any{value}}}, limit, offset)
}

// FindTalentAssessmentByAssessedByPrincipalID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindTalentAssessmentByAssessedByPrincipalID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.TalentAssessment], error) {
	if s == nil {
		return operation.Page[entity.TalentAssessment]{}, hrerrors.ErrUnavailable
	}
	return s.talentAssessment.List(ctx, []domainrepo.Filter{{Column: "assessed_by_principal_id", Values: []any{value}}}, limit, offset)
}

// FindTalentAssessmentByEvidenceDocumentVersionReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindTalentAssessmentByEvidenceDocumentVersionReferenceID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.TalentAssessment], error) {
	if s == nil {
		return operation.Page[entity.TalentAssessment]{}, hrerrors.ErrUnavailable
	}
	return s.talentAssessment.List(ctx, []domainrepo.Filter{{Column: "evidence_document_version_reference_id", Values: []any{value}}}, limit, offset)
}

// FindTalentAssessmentBySegmentCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindTalentAssessmentBySegmentCode(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.TalentAssessment], error) {
	if s == nil {
		return operation.Page[entity.TalentAssessment]{}, hrerrors.ErrUnavailable
	}
	return s.talentAssessment.List(ctx, []domainrepo.Filter{{Column: "segment_code", Values: []any{value}}}, limit, offset)
}

// FindTalentAssessmentByFinalizedByPrincipalID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindTalentAssessmentByFinalizedByPrincipalID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.TalentAssessment], error) {
	if s == nil {
		return operation.Page[entity.TalentAssessment]{}, hrerrors.ErrUnavailable
	}
	return s.talentAssessment.List(ctx, []domainrepo.Filter{{Column: "finalized_by_principal_id", Values: []any{value}}}, limit, offset)
}

// CreateTalentAssessmentScore menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreateTalentAssessmentScore(ctx context.Context, in CreateTalentAssessmentScoreInput) (*entity.TalentAssessmentScore, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.talentAssessmentScore.Create(ctx, func() (*entity.TalentAssessmentScore, error) { return createTalentAssessmentScore(in) })
}

// UpdateTalentAssessmentScore menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdateTalentAssessmentScore(ctx context.Context, key string, in UpdateTalentAssessmentScoreInput) (*entity.TalentAssessmentScore, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.talentAssessmentScore.Update(ctx, key, func(value *entity.TalentAssessmentScore) error { return patchTalentAssessmentScore(value, in) })
}

// DeleteTalentAssessmentScore menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeleteTalentAssessmentScore(ctx context.Context, key string) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.talentAssessmentScore.Delete(ctx, key)
}

// GetTalentAssessmentScoreByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetTalentAssessmentScoreByID(ctx context.Context, key string) (*entity.TalentAssessmentScore, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.talentAssessmentScore.Get(ctx, key)
}

// GetAllTalentAssessmentScores menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllTalentAssessmentScores(ctx context.Context, limit, offset int) (operation.Page[entity.TalentAssessmentScore], error) {
	if s == nil {
		return operation.Page[entity.TalentAssessmentScore]{}, hrerrors.ErrUnavailable
	}
	return s.talentAssessmentScore.List(ctx, nil, limit, offset)
}

// FindTalentAssessmentScoreByAssessmentID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindTalentAssessmentScoreByAssessmentID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.TalentAssessmentScore], error) {
	if s == nil {
		return operation.Page[entity.TalentAssessmentScore]{}, hrerrors.ErrUnavailable
	}
	return s.talentAssessmentScore.List(ctx, []domainrepo.Filter{{Column: "assessment_id", Values: []any{value}}}, limit, offset)
}

// FindTalentAssessmentScoreByFrameworkID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindTalentAssessmentScoreByFrameworkID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.TalentAssessmentScore], error) {
	if s == nil {
		return operation.Page[entity.TalentAssessmentScore]{}, hrerrors.ErrUnavailable
	}
	return s.talentAssessmentScore.List(ctx, []domainrepo.Filter{{Column: "framework_id", Values: []any{value}}}, limit, offset)
}

// FindTalentAssessmentScoreByDimensionID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindTalentAssessmentScoreByDimensionID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.TalentAssessmentScore], error) {
	if s == nil {
		return operation.Page[entity.TalentAssessmentScore]{}, hrerrors.ErrUnavailable
	}
	return s.talentAssessmentScore.List(ctx, []domainrepo.Filter{{Column: "dimension_id", Values: []any{value}}}, limit, offset)
}

// FindTalentAssessmentScoreByEvidenceDocumentVersionReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindTalentAssessmentScoreByEvidenceDocumentVersionReferenceID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.TalentAssessmentScore], error) {
	if s == nil {
		return operation.Page[entity.TalentAssessmentScore]{}, hrerrors.ErrUnavailable
	}
	return s.talentAssessmentScore.List(ctx, []domainrepo.Filter{{Column: "evidence_document_version_reference_id", Values: []any{value}}}, limit, offset)
}

// CreateTalentCalibrationSession menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreateTalentCalibrationSession(ctx context.Context, in CreateTalentCalibrationSessionInput) (*entity.TalentCalibrationSession, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.talentCalibrationSession.Create(ctx, func() (*entity.TalentCalibrationSession, error) { return createTalentCalibrationSession(in) })
}

// UpdateTalentCalibrationSession menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdateTalentCalibrationSession(ctx context.Context, key string, in UpdateTalentCalibrationSessionInput) (*entity.TalentCalibrationSession, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.talentCalibrationSession.Update(ctx, key, func(value *entity.TalentCalibrationSession) error { return patchTalentCalibrationSession(value, in) })
}

// DeleteTalentCalibrationSession menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeleteTalentCalibrationSession(ctx context.Context, key string) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.talentCalibrationSession.Delete(ctx, key)
}

// GetTalentCalibrationSessionByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetTalentCalibrationSessionByID(ctx context.Context, key string) (*entity.TalentCalibrationSession, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.talentCalibrationSession.Get(ctx, key)
}

// GetAllTalentCalibrationSessions menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllTalentCalibrationSessions(ctx context.Context, limit, offset int) (operation.Page[entity.TalentCalibrationSession], error) {
	if s == nil {
		return operation.Page[entity.TalentCalibrationSession]{}, hrerrors.ErrUnavailable
	}
	return s.talentCalibrationSession.List(ctx, nil, limit, offset)
}

// FindTalentCalibrationSessionByTalentReviewCycleID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindTalentCalibrationSessionByTalentReviewCycleID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.TalentCalibrationSession], error) {
	if s == nil {
		return operation.Page[entity.TalentCalibrationSession]{}, hrerrors.ErrUnavailable
	}
	return s.talentCalibrationSession.List(ctx, []domainrepo.Filter{{Column: "talent_review_cycle_id", Values: []any{value}}}, limit, offset)
}

// FindTalentCalibrationSessionBySessionCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindTalentCalibrationSessionBySessionCode(ctx context.Context, value string, limit, offset int) (operation.Page[entity.TalentCalibrationSession], error) {
	if s == nil {
		return operation.Page[entity.TalentCalibrationSession]{}, hrerrors.ErrUnavailable
	}
	return s.talentCalibrationSession.List(ctx, []domainrepo.Filter{{Column: "session_code", Values: []any{value}}}, limit, offset)
}

// FindTalentCalibrationSessionByChairPrincipalID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindTalentCalibrationSessionByChairPrincipalID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.TalentCalibrationSession], error) {
	if s == nil {
		return operation.Page[entity.TalentCalibrationSession]{}, hrerrors.ErrUnavailable
	}
	return s.talentCalibrationSession.List(ctx, []domainrepo.Filter{{Column: "chair_principal_id", Values: []any{value}}}, limit, offset)
}

// FindTalentCalibrationSessionByStatus menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindTalentCalibrationSessionByStatus(ctx context.Context, value string, limit, offset int) (operation.Page[entity.TalentCalibrationSession], error) {
	if s == nil {
		return operation.Page[entity.TalentCalibrationSession]{}, hrerrors.ErrUnavailable
	}
	return s.talentCalibrationSession.List(ctx, []domainrepo.Filter{{Column: "status", Values: []any{value}}}, limit, offset)
}

// FindTalentCalibrationSessionByMinutesDocumentVersionReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindTalentCalibrationSessionByMinutesDocumentVersionReferenceID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.TalentCalibrationSession], error) {
	if s == nil {
		return operation.Page[entity.TalentCalibrationSession]{}, hrerrors.ErrUnavailable
	}
	return s.talentCalibrationSession.List(ctx, []domainrepo.Filter{{Column: "minutes_document_version_reference_id", Values: []any{value}}}, limit, offset)
}

// CreateTalentCalibrationDecision menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreateTalentCalibrationDecision(ctx context.Context, in CreateTalentCalibrationDecisionInput) (*entity.TalentCalibrationDecision, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.talentCalibrationDecision.Create(ctx, func() (*entity.TalentCalibrationDecision, error) { return createTalentCalibrationDecision(in) })
}

// UpdateTalentCalibrationDecision menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdateTalentCalibrationDecision(ctx context.Context, key string, in UpdateTalentCalibrationDecisionInput) (*entity.TalentCalibrationDecision, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.talentCalibrationDecision.Update(ctx, key, func(value *entity.TalentCalibrationDecision) error { return patchTalentCalibrationDecision(value, in) })
}

// DeleteTalentCalibrationDecision menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeleteTalentCalibrationDecision(ctx context.Context, key string) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.talentCalibrationDecision.Delete(ctx, key)
}

// GetTalentCalibrationDecisionByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetTalentCalibrationDecisionByID(ctx context.Context, key string) (*entity.TalentCalibrationDecision, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.talentCalibrationDecision.Get(ctx, key)
}

// GetAllTalentCalibrationDecisions menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllTalentCalibrationDecisions(ctx context.Context, limit, offset int) (operation.Page[entity.TalentCalibrationDecision], error) {
	if s == nil {
		return operation.Page[entity.TalentCalibrationDecision]{}, hrerrors.ErrUnavailable
	}
	return s.talentCalibrationDecision.List(ctx, nil, limit, offset)
}

// FindTalentCalibrationDecisionBySessionID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindTalentCalibrationDecisionBySessionID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.TalentCalibrationDecision], error) {
	if s == nil {
		return operation.Page[entity.TalentCalibrationDecision]{}, hrerrors.ErrUnavailable
	}
	return s.talentCalibrationDecision.List(ctx, []domainrepo.Filter{{Column: "session_id", Values: []any{value}}}, limit, offset)
}

// FindTalentCalibrationDecisionByTalentReviewCycleID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindTalentCalibrationDecisionByTalentReviewCycleID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.TalentCalibrationDecision], error) {
	if s == nil {
		return operation.Page[entity.TalentCalibrationDecision]{}, hrerrors.ErrUnavailable
	}
	return s.talentCalibrationDecision.List(ctx, []domainrepo.Filter{{Column: "talent_review_cycle_id", Values: []any{value}}}, limit, offset)
}

// FindTalentCalibrationDecisionByAssessmentID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindTalentCalibrationDecisionByAssessmentID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.TalentCalibrationDecision], error) {
	if s == nil {
		return operation.Page[entity.TalentCalibrationDecision]{}, hrerrors.ErrUnavailable
	}
	return s.talentCalibrationDecision.List(ctx, []domainrepo.Filter{{Column: "assessment_id", Values: []any{value}}}, limit, offset)
}

// FindTalentCalibrationDecisionByDimensionID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindTalentCalibrationDecisionByDimensionID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.TalentCalibrationDecision], error) {
	if s == nil {
		return operation.Page[entity.TalentCalibrationDecision]{}, hrerrors.ErrUnavailable
	}
	return s.talentCalibrationDecision.List(ctx, []domainrepo.Filter{{Column: "dimension_id", Values: []any{value}}}, limit, offset)
}

// FindTalentCalibrationDecisionByRationaleDocumentVersionReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindTalentCalibrationDecisionByRationaleDocumentVersionReferenceID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.TalentCalibrationDecision], error) {
	if s == nil {
		return operation.Page[entity.TalentCalibrationDecision]{}, hrerrors.ErrUnavailable
	}
	return s.talentCalibrationDecision.List(ctx, []domainrepo.Filter{{Column: "rationale_document_version_reference_id", Values: []any{value}}}, limit, offset)
}

// FindTalentCalibrationDecisionByDecidedByPrincipalID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindTalentCalibrationDecisionByDecidedByPrincipalID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.TalentCalibrationDecision], error) {
	if s == nil {
		return operation.Page[entity.TalentCalibrationDecision]{}, hrerrors.ErrUnavailable
	}
	return s.talentCalibrationDecision.List(ctx, []domainrepo.Filter{{Column: "decided_by_principal_id", Values: []any{value}}}, limit, offset)
}

// CreateTalentPool menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreateTalentPool(ctx context.Context, in CreateTalentPoolInput) (*entity.TalentPool, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.talentPool.Create(ctx, func() (*entity.TalentPool, error) { return createTalentPool(in) })
}

// UpdateTalentPool menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdateTalentPool(ctx context.Context, key string, in UpdateTalentPoolInput) (*entity.TalentPool, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.talentPool.Update(ctx, key, func(value *entity.TalentPool) error { return patchTalentPool(value, in) })
}

// DeleteTalentPool menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeleteTalentPool(ctx context.Context, key string) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.talentPool.Delete(ctx, key)
}

// GetTalentPoolByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetTalentPoolByID(ctx context.Context, key string) (*entity.TalentPool, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.talentPool.Get(ctx, key)
}

// GetAllTalentPools menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllTalentPools(ctx context.Context, limit, offset int) (operation.Page[entity.TalentPool], error) {
	if s == nil {
		return operation.Page[entity.TalentPool]{}, hrerrors.ErrUnavailable
	}
	return s.talentPool.List(ctx, nil, limit, offset)
}

// FindTalentPoolByPoolCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindTalentPoolByPoolCode(ctx context.Context, value string, limit, offset int) (operation.Page[entity.TalentPool], error) {
	if s == nil {
		return operation.Page[entity.TalentPool]{}, hrerrors.ErrUnavailable
	}
	return s.talentPool.List(ctx, []domainrepo.Filter{{Column: "pool_code", Values: []any{value}}}, limit, offset)
}

// FindTalentPoolByLegalEntityReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindTalentPoolByLegalEntityReferenceID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.TalentPool], error) {
	if s == nil {
		return operation.Page[entity.TalentPool]{}, hrerrors.ErrUnavailable
	}
	return s.talentPool.List(ctx, []domainrepo.Filter{{Column: "legal_entity_reference_id", Values: []any{value}}}, limit, offset)
}

// FindTalentPoolByTargetJobID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindTalentPoolByTargetJobID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.TalentPool], error) {
	if s == nil {
		return operation.Page[entity.TalentPool]{}, hrerrors.ErrUnavailable
	}
	return s.talentPool.List(ctx, []domainrepo.Filter{{Column: "target_job_id", Values: []any{value}}}, limit, offset)
}

// FindTalentPoolByEligibilityDocumentVersionReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindTalentPoolByEligibilityDocumentVersionReferenceID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.TalentPool], error) {
	if s == nil {
		return operation.Page[entity.TalentPool]{}, hrerrors.ErrUnavailable
	}
	return s.talentPool.List(ctx, []domainrepo.Filter{{Column: "eligibility_document_version_reference_id", Values: []any{value}}}, limit, offset)
}

// FindTalentPoolByOwnerPrincipalID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindTalentPoolByOwnerPrincipalID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.TalentPool], error) {
	if s == nil {
		return operation.Page[entity.TalentPool]{}, hrerrors.ErrUnavailable
	}
	return s.talentPool.List(ctx, []domainrepo.Filter{{Column: "owner_principal_id", Values: []any{value}}}, limit, offset)
}

// CreateTalentPoolMembership menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreateTalentPoolMembership(ctx context.Context, in CreateTalentPoolMembershipInput) (*entity.TalentPoolMembership, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.talentPoolMembership.Create(ctx, func() (*entity.TalentPoolMembership, error) { return createTalentPoolMembership(in) })
}

// UpdateTalentPoolMembership menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdateTalentPoolMembership(ctx context.Context, key string, in UpdateTalentPoolMembershipInput) (*entity.TalentPoolMembership, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.talentPoolMembership.Update(ctx, key, func(value *entity.TalentPoolMembership) error { return patchTalentPoolMembership(value, in) })
}

// DeleteTalentPoolMembership menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeleteTalentPoolMembership(ctx context.Context, key string) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.talentPoolMembership.Delete(ctx, key)
}

// GetTalentPoolMembershipByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetTalentPoolMembershipByID(ctx context.Context, key string) (*entity.TalentPoolMembership, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.talentPoolMembership.Get(ctx, key)
}

// GetAllTalentPoolMemberships menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllTalentPoolMemberships(ctx context.Context, limit, offset int) (operation.Page[entity.TalentPoolMembership], error) {
	if s == nil {
		return operation.Page[entity.TalentPoolMembership]{}, hrerrors.ErrUnavailable
	}
	return s.talentPoolMembership.List(ctx, nil, limit, offset)
}

// FindTalentPoolMembershipByTalentPoolID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindTalentPoolMembershipByTalentPoolID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.TalentPoolMembership], error) {
	if s == nil {
		return operation.Page[entity.TalentPoolMembership]{}, hrerrors.ErrUnavailable
	}
	return s.talentPoolMembership.List(ctx, []domainrepo.Filter{{Column: "talent_pool_id", Values: []any{value}}}, limit, offset)
}

// FindTalentPoolMembershipByWorkerID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindTalentPoolMembershipByWorkerID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.TalentPoolMembership], error) {
	if s == nil {
		return operation.Page[entity.TalentPoolMembership]{}, hrerrors.ErrUnavailable
	}
	return s.talentPoolMembership.List(ctx, []domainrepo.Filter{{Column: "worker_id", Values: []any{value}}}, limit, offset)
}

// FindTalentPoolMembershipByAssessmentID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindTalentPoolMembershipByAssessmentID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.TalentPoolMembership], error) {
	if s == nil {
		return operation.Page[entity.TalentPoolMembership]{}, hrerrors.ErrUnavailable
	}
	return s.talentPoolMembership.List(ctx, []domainrepo.Filter{{Column: "assessment_id", Values: []any{value}}}, limit, offset)
}

// FindTalentPoolMembershipByNominatedByPrincipalID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindTalentPoolMembershipByNominatedByPrincipalID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.TalentPoolMembership], error) {
	if s == nil {
		return operation.Page[entity.TalentPoolMembership]{}, hrerrors.ErrUnavailable
	}
	return s.talentPoolMembership.List(ctx, []domainrepo.Filter{{Column: "nominated_by_principal_id", Values: []any{value}}}, limit, offset)
}

// FindTalentPoolMembershipByStatus menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindTalentPoolMembershipByStatus(ctx context.Context, value string, limit, offset int) (operation.Page[entity.TalentPoolMembership], error) {
	if s == nil {
		return operation.Page[entity.TalentPoolMembership]{}, hrerrors.ErrUnavailable
	}
	return s.talentPoolMembership.List(ctx, []domainrepo.Filter{{Column: "status", Values: []any{value}}}, limit, offset)
}

// FindTalentPoolMembershipByApprovedByPrincipalID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindTalentPoolMembershipByApprovedByPrincipalID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.TalentPoolMembership], error) {
	if s == nil {
		return operation.Page[entity.TalentPoolMembership]{}, hrerrors.ErrUnavailable
	}
	return s.talentPoolMembership.List(ctx, []domainrepo.Filter{{Column: "approved_by_principal_id", Values: []any{value}}}, limit, offset)
}

// FindTalentPoolMembershipByExitReasonCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindTalentPoolMembershipByExitReasonCode(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.TalentPoolMembership], error) {
	if s == nil {
		return operation.Page[entity.TalentPoolMembership]{}, hrerrors.ErrUnavailable
	}
	return s.talentPoolMembership.List(ctx, []domainrepo.Filter{{Column: "exit_reason_code", Values: []any{value}}}, limit, offset)
}
