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
	Runtime                  operation.Dependencies
	LearningCourse           operation.Repository[entity.LearningCourse, string]
	LearningEnrollment       operation.Repository[entity.LearningEnrollment, string]
	LearningCourseCompetency operation.Repository[entity.LearningCourseCompetency, entity.LearningCourseCompetencyKey]
}

// Service mengoordinasikan operasi domain learning.
type Service struct {
	learningCourse           *operation.Resource[entity.LearningCourse, string, *entity.LearningCourse]
	learningEnrollment       *operation.Resource[entity.LearningEnrollment, string, *entity.LearningEnrollment]
	learningCourseCompetency *operation.Resource[entity.LearningCourseCompetency, entity.LearningCourseCompetencyKey, *entity.LearningCourseCompetency]
}

// NewService memeriksa seluruh port dan dependency sebelum service dapat dipakai.
func NewService(deps Dependencies) (*Service, error) {
	s := &Service{}
	var err error
	s.learningCourse, err = operation.NewResource[entity.LearningCourse, string, *entity.LearningCourse](deps.LearningCourse, deps.Runtime, specLearningCourse())
	if err != nil {
		return nil, err
	}
	s.learningEnrollment, err = operation.NewResource[entity.LearningEnrollment, string, *entity.LearningEnrollment](deps.LearningEnrollment, deps.Runtime, specLearningEnrollment())
	if err != nil {
		return nil, err
	}
	s.learningCourseCompetency, err = operation.NewResource[entity.LearningCourseCompetency, entity.LearningCourseCompetencyKey, *entity.LearningCourseCompetency](deps.LearningCourseCompetency, deps.Runtime, specLearningCourseCompetency())
	if err != nil {
		return nil, err
	}
	return s, nil
}

// CreateLearningCourse menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreateLearningCourse(ctx context.Context, in CreateLearningCourseInput) (*entity.LearningCourse, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.learningCourse.Create(ctx, func() (*entity.LearningCourse, error) { return createLearningCourse(in) })
}

// UpdateLearningCourse menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdateLearningCourse(ctx context.Context, key string, in UpdateLearningCourseInput) (*entity.LearningCourse, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.learningCourse.Update(ctx, key, func(value *entity.LearningCourse) error { return patchLearningCourse(value, in) })
}

// DeleteLearningCourse menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeleteLearningCourse(ctx context.Context, key string) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.learningCourse.Delete(ctx, key)
}

// GetLearningCourseByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetLearningCourseByID(ctx context.Context, key string) (*entity.LearningCourse, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.learningCourse.Get(ctx, key)
}

// GetAllLearningCourses menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllLearningCourses(ctx context.Context, limit, offset int) (operation.Page[entity.LearningCourse], error) {
	if s == nil {
		return operation.Page[entity.LearningCourse]{}, hrerrors.ErrUnavailable
	}
	return s.learningCourse.List(ctx, nil, limit, offset)
}

// FindLearningCourseByCourseCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindLearningCourseByCourseCode(ctx context.Context, value string, limit, offset int) (operation.Page[entity.LearningCourse], error) {
	if s == nil {
		return operation.Page[entity.LearningCourse]{}, hrerrors.ErrUnavailable
	}
	return s.learningCourse.List(ctx, []domainrepo.Filter{{Column: "course_code", Values: []any{value}}}, limit, offset)
}

// FindLearningCourseByCourseType menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindLearningCourseByCourseType(ctx context.Context, value string, limit, offset int) (operation.Page[entity.LearningCourse], error) {
	if s == nil {
		return operation.Page[entity.LearningCourse]{}, hrerrors.ErrUnavailable
	}
	return s.learningCourse.List(ctx, []domainrepo.Filter{{Column: "course_type", Values: []any{value}}}, limit, offset)
}

// FindLearningCourseByProviderPartyReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindLearningCourseByProviderPartyReferenceID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.LearningCourse], error) {
	if s == nil {
		return operation.Page[entity.LearningCourse]{}, hrerrors.ErrUnavailable
	}
	return s.learningCourse.List(ctx, []domainrepo.Filter{{Column: "provider_party_reference_id", Values: []any{value}}}, limit, offset)
}

// CreateLearningEnrollment menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreateLearningEnrollment(ctx context.Context, in CreateLearningEnrollmentInput) (*entity.LearningEnrollment, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.learningEnrollment.Create(ctx, func() (*entity.LearningEnrollment, error) { return createLearningEnrollment(in) })
}

// UpdateLearningEnrollment menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdateLearningEnrollment(ctx context.Context, key string, in UpdateLearningEnrollmentInput) (*entity.LearningEnrollment, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.learningEnrollment.Update(ctx, key, func(value *entity.LearningEnrollment) error { return patchLearningEnrollment(value, in) })
}

// DeleteLearningEnrollment menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeleteLearningEnrollment(ctx context.Context, key string) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.learningEnrollment.Delete(ctx, key)
}

// GetLearningEnrollmentByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetLearningEnrollmentByID(ctx context.Context, key string) (*entity.LearningEnrollment, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.learningEnrollment.Get(ctx, key)
}

// GetAllLearningEnrollments menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllLearningEnrollments(ctx context.Context, limit, offset int) (operation.Page[entity.LearningEnrollment], error) {
	if s == nil {
		return operation.Page[entity.LearningEnrollment]{}, hrerrors.ErrUnavailable
	}
	return s.learningEnrollment.List(ctx, nil, limit, offset)
}

// FindLearningEnrollmentByWorkerID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindLearningEnrollmentByWorkerID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.LearningEnrollment], error) {
	if s == nil {
		return operation.Page[entity.LearningEnrollment]{}, hrerrors.ErrUnavailable
	}
	return s.learningEnrollment.List(ctx, []domainrepo.Filter{{Column: "worker_id", Values: []any{value}}}, limit, offset)
}

// FindLearningEnrollmentByCourseID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindLearningEnrollmentByCourseID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.LearningEnrollment], error) {
	if s == nil {
		return operation.Page[entity.LearningEnrollment]{}, hrerrors.ErrUnavailable
	}
	return s.learningEnrollment.List(ctx, []domainrepo.Filter{{Column: "course_id", Values: []any{value}}}, limit, offset)
}

// FindLearningEnrollmentByEnrollmentStatus menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindLearningEnrollmentByEnrollmentStatus(ctx context.Context, value entity.LearningStatus, limit, offset int) (operation.Page[entity.LearningEnrollment], error) {
	if s == nil {
		return operation.Page[entity.LearningEnrollment]{}, hrerrors.ErrUnavailable
	}
	return s.learningEnrollment.List(ctx, []domainrepo.Filter{{Column: "enrollment_status", Values: []any{value}}}, limit, offset)
}

// FindLearningEnrollmentByAssignedByPrincipalID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindLearningEnrollmentByAssignedByPrincipalID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.LearningEnrollment], error) {
	if s == nil {
		return operation.Page[entity.LearningEnrollment]{}, hrerrors.ErrUnavailable
	}
	return s.learningEnrollment.List(ctx, []domainrepo.Filter{{Column: "assigned_by_principal_id", Values: []any{value}}}, limit, offset)
}

// FindLearningEnrollmentByCredentialReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindLearningEnrollmentByCredentialReferenceID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.LearningEnrollment], error) {
	if s == nil {
		return operation.Page[entity.LearningEnrollment]{}, hrerrors.ErrUnavailable
	}
	return s.learningEnrollment.List(ctx, []domainrepo.Filter{{Column: "credential_reference_id", Values: []any{value}}}, limit, offset)
}

// FindLearningEnrollmentByFundingReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindLearningEnrollmentByFundingReferenceID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.LearningEnrollment], error) {
	if s == nil {
		return operation.Page[entity.LearningEnrollment]{}, hrerrors.ErrUnavailable
	}
	return s.learningEnrollment.List(ctx, []domainrepo.Filter{{Column: "funding_reference_id", Values: []any{value}}}, limit, offset)
}

// FindLearningEnrollmentByIdempotencyKey menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindLearningEnrollmentByIdempotencyKey(ctx context.Context, value string, limit, offset int) (operation.Page[entity.LearningEnrollment], error) {
	if s == nil {
		return operation.Page[entity.LearningEnrollment]{}, hrerrors.ErrUnavailable
	}
	return s.learningEnrollment.List(ctx, []domainrepo.Filter{{Column: "idempotency_key", Values: []any{value}}}, limit, offset)
}

// CreateLearningCourseCompetency menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreateLearningCourseCompetency(ctx context.Context, in CreateLearningCourseCompetencyInput) (*entity.LearningCourseCompetency, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.learningCourseCompetency.Create(ctx, func() (*entity.LearningCourseCompetency, error) { return createLearningCourseCompetency(in) })
}

// UpdateLearningCourseCompetency menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdateLearningCourseCompetency(ctx context.Context, key entity.LearningCourseCompetencyKey, in UpdateLearningCourseCompetencyInput) (*entity.LearningCourseCompetency, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.learningCourseCompetency.Update(ctx, key, func(value *entity.LearningCourseCompetency) error { return patchLearningCourseCompetency(value, in) })
}

// DeleteLearningCourseCompetency menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeleteLearningCourseCompetency(ctx context.Context, key entity.LearningCourseCompetencyKey) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.learningCourseCompetency.Delete(ctx, key)
}

// GetLearningCourseCompetencyByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetLearningCourseCompetencyByID(ctx context.Context, key entity.LearningCourseCompetencyKey) (*entity.LearningCourseCompetency, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.learningCourseCompetency.Get(ctx, key)
}

// GetAllLearningCourseCompetencies menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllLearningCourseCompetencies(ctx context.Context, limit, offset int) (operation.Page[entity.LearningCourseCompetency], error) {
	if s == nil {
		return operation.Page[entity.LearningCourseCompetency]{}, hrerrors.ErrUnavailable
	}
	return s.learningCourseCompetency.List(ctx, nil, limit, offset)
}

// FindLearningCourseCompetencyByCourseID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindLearningCourseCompetencyByCourseID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.LearningCourseCompetency], error) {
	if s == nil {
		return operation.Page[entity.LearningCourseCompetency]{}, hrerrors.ErrUnavailable
	}
	return s.learningCourseCompetency.List(ctx, []domainrepo.Filter{{Column: "course_id", Values: []any{value}}}, limit, offset)
}

// FindLearningCourseCompetencyByCompetencyID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindLearningCourseCompetencyByCompetencyID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.LearningCourseCompetency], error) {
	if s == nil {
		return operation.Page[entity.LearningCourseCompetency]{}, hrerrors.ErrUnavailable
	}
	return s.learningCourseCompetency.List(ctx, []domainrepo.Filter{{Column: "competency_id", Values: []any{value}}}, limit, offset)
}
