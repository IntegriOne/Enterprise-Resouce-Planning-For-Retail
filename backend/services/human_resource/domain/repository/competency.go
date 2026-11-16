package repository

import (
	"context"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
)

// WorkerQualificationRepository adalah port persistensi hr_worker_qualifications.
type WorkerQualificationRepository interface {
	Repository[entity.WorkerQualification, string]
	InsertWorkerQualifications(context.Context, *entity.WorkerQualification) error
	UpdateWorkerQualifications(context.Context, *entity.WorkerQualification) error
	FindWorkerQualificationsByUUID(context.Context, string) (*entity.WorkerQualification, error)
	FindWorkerQualificationsByUUIDForUpdate(context.Context, string) (*entity.WorkerQualification, error)
	DeleteWorkerQualificationsByUUID(context.Context, string) error
	GetAllWorkerQualifications(context.Context, int, int) ([]entity.WorkerQualification, error)
	CountWorkerQualifications(context.Context) (int64, error)
	SelectWorkerQualificationsWhere(context.Context, []Filter, int, int) ([]entity.WorkerQualification, error)
	CountWorkerQualificationsWhere(context.Context, []Filter) (int64, error)
	FindWorkerQualificationsByWorkerID(context.Context, string, int, int) ([]entity.WorkerQualification, error)
	FindWorkerQualificationsByQualificationType(context.Context, string, int, int) ([]entity.WorkerQualification, error)
	FindWorkerQualificationsByQualificationCode(context.Context, *string, int, int) ([]entity.WorkerQualification, error)
	FindWorkerQualificationsByIssuerPartyReferenceID(context.Context, *string, int, int) ([]entity.WorkerQualification, error)
	FindWorkerQualificationsByCredentialReferenceID(context.Context, *string, int, int) ([]entity.WorkerQualification, error)
	FindWorkerQualificationsByVerificationStatus(context.Context, string, int, int) ([]entity.WorkerQualification, error)
	FindWorkerQualificationsByVerifiedByPrincipalID(context.Context, *string, int, int) ([]entity.WorkerQualification, error)
	FindByWorkerID(ctx context.Context, value string, limit, offset int) ([]entity.WorkerQualification, error)
	FindByQualificationType(ctx context.Context, value string, limit, offset int) ([]entity.WorkerQualification, error)
	FindByQualificationCode(ctx context.Context, value *string, limit, offset int) ([]entity.WorkerQualification, error)
	FindByIssuerPartyReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.WorkerQualification, error)
	FindByCredentialReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.WorkerQualification, error)
	FindByVerificationStatus(ctx context.Context, value string, limit, offset int) ([]entity.WorkerQualification, error)
	FindByVerifiedByPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.WorkerQualification, error)
}

// CompetencyRepository adalah port persistensi hr_competencies.
type CompetencyRepository interface {
	Repository[entity.Competency, string]
	InsertCompetencies(context.Context, *entity.Competency) error
	UpdateCompetencies(context.Context, *entity.Competency) error
	FindCompetenciesByUUID(context.Context, string) (*entity.Competency, error)
	FindCompetenciesByUUIDForUpdate(context.Context, string) (*entity.Competency, error)
	DeleteCompetenciesByUUID(context.Context, string) error
	GetAllCompetencies(context.Context, int, int) ([]entity.Competency, error)
	CountCompetencies(context.Context) (int64, error)
	SelectCompetenciesWhere(context.Context, []Filter, int, int) ([]entity.Competency, error)
	CountCompetenciesWhere(context.Context, []Filter) (int64, error)
	FindCompetenciesByCompetencyCode(context.Context, string, int, int) ([]entity.Competency, error)
	FindCompetenciesByCategoryCode(context.Context, string, int, int) ([]entity.Competency, error)
	FindByCompetencyCode(ctx context.Context, value string, limit, offset int) ([]entity.Competency, error)
	FindByCategoryCode(ctx context.Context, value string, limit, offset int) ([]entity.Competency, error)
}

// JobCompetencyRepository adalah port persistensi hr_job_competencies.
type JobCompetencyRepository interface {
	Repository[entity.JobCompetency, string]
	InsertJobCompetencies(context.Context, *entity.JobCompetency) error
	UpdateJobCompetencies(context.Context, *entity.JobCompetency) error
	FindJobCompetenciesByUUID(context.Context, string) (*entity.JobCompetency, error)
	FindJobCompetenciesByUUIDForUpdate(context.Context, string) (*entity.JobCompetency, error)
	DeleteJobCompetenciesByUUID(context.Context, string) error
	GetAllJobCompetencies(context.Context, int, int) ([]entity.JobCompetency, error)
	CountJobCompetencies(context.Context) (int64, error)
	SelectJobCompetenciesWhere(context.Context, []Filter, int, int) ([]entity.JobCompetency, error)
	CountJobCompetenciesWhere(context.Context, []Filter) (int64, error)
	FindJobCompetenciesByJobID(context.Context, string, int, int) ([]entity.JobCompetency, error)
	FindJobCompetenciesByCompetencyID(context.Context, string, int, int) ([]entity.JobCompetency, error)
	FindByJobID(ctx context.Context, value string, limit, offset int) ([]entity.JobCompetency, error)
	FindByCompetencyID(ctx context.Context, value string, limit, offset int) ([]entity.JobCompetency, error)
}

// WorkerCompetencyAssessmentRepository adalah port persistensi hr_worker_competency_assessments.
type WorkerCompetencyAssessmentRepository interface {
	Repository[entity.WorkerCompetencyAssessment, string]
	InsertWorkerCompetencyAssessments(context.Context, *entity.WorkerCompetencyAssessment) error
	UpdateWorkerCompetencyAssessments(context.Context, *entity.WorkerCompetencyAssessment) error
	FindWorkerCompetencyAssessmentsByUUID(context.Context, string) (*entity.WorkerCompetencyAssessment, error)
	FindWorkerCompetencyAssessmentsByUUIDForUpdate(context.Context, string) (*entity.WorkerCompetencyAssessment, error)
	DeleteWorkerCompetencyAssessmentsByUUID(context.Context, string) error
	GetAllWorkerCompetencyAssessments(context.Context, int, int) ([]entity.WorkerCompetencyAssessment, error)
	CountWorkerCompetencyAssessments(context.Context) (int64, error)
	SelectWorkerCompetencyAssessmentsWhere(context.Context, []Filter, int, int) ([]entity.WorkerCompetencyAssessment, error)
	CountWorkerCompetencyAssessmentsWhere(context.Context, []Filter) (int64, error)
	FindWorkerCompetencyAssessmentsByWorkerID(context.Context, string, int, int) ([]entity.WorkerCompetencyAssessment, error)
	FindWorkerCompetencyAssessmentsByCompetencyID(context.Context, string, int, int) ([]entity.WorkerCompetencyAssessment, error)
	FindWorkerCompetencyAssessmentsByAssessedByPrincipalID(context.Context, string, int, int) ([]entity.WorkerCompetencyAssessment, error)
	FindWorkerCompetencyAssessmentsByEvidenceDocumentVersionReferenceID(context.Context, *string, int, int) ([]entity.WorkerCompetencyAssessment, error)
	FindWorkerCompetencyAssessmentsByIdempotencyKey(context.Context, string, int, int) ([]entity.WorkerCompetencyAssessment, error)
	FindByWorkerID(ctx context.Context, value string, limit, offset int) ([]entity.WorkerCompetencyAssessment, error)
	FindByCompetencyID(ctx context.Context, value string, limit, offset int) ([]entity.WorkerCompetencyAssessment, error)
	FindByAssessedByPrincipalID(ctx context.Context, value string, limit, offset int) ([]entity.WorkerCompetencyAssessment, error)
	FindByEvidenceDocumentVersionReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.WorkerCompetencyAssessment, error)
	FindByIdempotencyKey(ctx context.Context, value string, limit, offset int) ([]entity.WorkerCompetencyAssessment, error)
}
