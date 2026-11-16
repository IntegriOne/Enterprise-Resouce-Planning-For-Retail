package repository

import (
	"context"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
)

// LearningCourseRepository adalah port persistensi hr_learning_courses.
type LearningCourseRepository interface {
	Repository[entity.LearningCourse, string]
	InsertLearningCourses(context.Context, *entity.LearningCourse) error
	UpdateLearningCourses(context.Context, *entity.LearningCourse) error
	FindLearningCoursesByUUID(context.Context, string) (*entity.LearningCourse, error)
	FindLearningCoursesByUUIDForUpdate(context.Context, string) (*entity.LearningCourse, error)
	DeleteLearningCoursesByUUID(context.Context, string) error
	GetAllLearningCourses(context.Context, int, int) ([]entity.LearningCourse, error)
	CountLearningCourses(context.Context) (int64, error)
	SelectLearningCoursesWhere(context.Context, []Filter, int, int) ([]entity.LearningCourse, error)
	CountLearningCoursesWhere(context.Context, []Filter) (int64, error)
	FindLearningCoursesByCourseCode(context.Context, string, int, int) ([]entity.LearningCourse, error)
	FindLearningCoursesByCourseType(context.Context, string, int, int) ([]entity.LearningCourse, error)
	FindLearningCoursesByProviderPartyReferenceID(context.Context, *string, int, int) ([]entity.LearningCourse, error)
	FindByCourseCode(ctx context.Context, value string, limit, offset int) ([]entity.LearningCourse, error)
	FindByCourseType(ctx context.Context, value string, limit, offset int) ([]entity.LearningCourse, error)
	FindByProviderPartyReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.LearningCourse, error)
}

// LearningEnrollmentRepository adalah port persistensi hr_learning_enrollments.
type LearningEnrollmentRepository interface {
	Repository[entity.LearningEnrollment, string]
	InsertLearningEnrollments(context.Context, *entity.LearningEnrollment) error
	UpdateLearningEnrollments(context.Context, *entity.LearningEnrollment) error
	FindLearningEnrollmentsByUUID(context.Context, string) (*entity.LearningEnrollment, error)
	FindLearningEnrollmentsByUUIDForUpdate(context.Context, string) (*entity.LearningEnrollment, error)
	DeleteLearningEnrollmentsByUUID(context.Context, string) error
	GetAllLearningEnrollments(context.Context, int, int) ([]entity.LearningEnrollment, error)
	CountLearningEnrollments(context.Context) (int64, error)
	SelectLearningEnrollmentsWhere(context.Context, []Filter, int, int) ([]entity.LearningEnrollment, error)
	CountLearningEnrollmentsWhere(context.Context, []Filter) (int64, error)
	FindLearningEnrollmentsByWorkerID(context.Context, string, int, int) ([]entity.LearningEnrollment, error)
	FindLearningEnrollmentsByCourseID(context.Context, string, int, int) ([]entity.LearningEnrollment, error)
	FindLearningEnrollmentsByEnrollmentStatus(context.Context, entity.LearningStatus, int, int) ([]entity.LearningEnrollment, error)
	FindLearningEnrollmentsByAssignedByPrincipalID(context.Context, *string, int, int) ([]entity.LearningEnrollment, error)
	FindLearningEnrollmentsByCredentialReferenceID(context.Context, *string, int, int) ([]entity.LearningEnrollment, error)
	FindLearningEnrollmentsByFundingReferenceID(context.Context, *string, int, int) ([]entity.LearningEnrollment, error)
	FindLearningEnrollmentsByIdempotencyKey(context.Context, string, int, int) ([]entity.LearningEnrollment, error)
	FindByWorkerID(ctx context.Context, value string, limit, offset int) ([]entity.LearningEnrollment, error)
	FindByCourseID(ctx context.Context, value string, limit, offset int) ([]entity.LearningEnrollment, error)
	FindByEnrollmentStatus(ctx context.Context, value entity.LearningStatus, limit, offset int) ([]entity.LearningEnrollment, error)
	FindByAssignedByPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.LearningEnrollment, error)
	FindByCredentialReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.LearningEnrollment, error)
	FindByFundingReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.LearningEnrollment, error)
	FindByIdempotencyKey(ctx context.Context, value string, limit, offset int) ([]entity.LearningEnrollment, error)
}

// LearningCourseCompetencyRepository adalah port persistensi hr_learning_course_competencies.
type LearningCourseCompetencyRepository interface {
	Repository[entity.LearningCourseCompetency, entity.LearningCourseCompetencyKey]
	InsertLearningCourseCompetencies(context.Context, *entity.LearningCourseCompetency) error
	UpdateLearningCourseCompetencies(context.Context, *entity.LearningCourseCompetency) error
	FindLearningCourseCompetenciesByCourseIDAndCompetencyID(context.Context, entity.LearningCourseCompetencyKey) (*entity.LearningCourseCompetency, error)
	FindLearningCourseCompetenciesByCourseIDAndCompetencyIDForUpdate(context.Context, entity.LearningCourseCompetencyKey) (*entity.LearningCourseCompetency, error)
	DeleteLearningCourseCompetenciesByCourseIDAndCompetencyID(context.Context, entity.LearningCourseCompetencyKey) error
	GetAllLearningCourseCompetencies(context.Context, int, int) ([]entity.LearningCourseCompetency, error)
	CountLearningCourseCompetencies(context.Context) (int64, error)
	SelectLearningCourseCompetenciesWhere(context.Context, []Filter, int, int) ([]entity.LearningCourseCompetency, error)
	CountLearningCourseCompetenciesWhere(context.Context, []Filter) (int64, error)
	FindLearningCourseCompetenciesByCourseID(context.Context, string, int, int) ([]entity.LearningCourseCompetency, error)
	FindLearningCourseCompetenciesByCompetencyID(context.Context, string, int, int) ([]entity.LearningCourseCompetency, error)
	FindByCourseID(ctx context.Context, value string, limit, offset int) ([]entity.LearningCourseCompetency, error)
	FindByCompetencyID(ctx context.Context, value string, limit, offset int) ([]entity.LearningCourseCompetency, error)
}
