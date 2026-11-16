package security

import (
	"context"
	"errors"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

type workerCompetencyAssessmentValidator struct {
	workers      lockedWorkerReader
	competencies lockedCompetencyReader
	now          func() time.Time
}

// ValidateMutation menjaga hasil asesmen sebagai riwayat yang tidak dapat ditimpa.
func (v *workerCompetencyAssessmentValidator) ValidateMutation(ctx context.Context, table string, before, after entity.Validatable) error {
	if v == nil || table != "hr_worker_competency_assessments" ||
		operation.RequireDependency(v.workers) != nil || operation.RequireDependency(v.competencies) != nil || v.now == nil {
		return hrerrors.ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if before != nil || after == nil {
		return hrerrors.ErrWorkflowRequired
	}
	assessment, ok := after.(*entity.WorkerCompetencyAssessment)
	if !ok || assessment == nil {
		return hrerrors.ErrInvalidInput
	}
	if assessment.EvidenceDocumentVersionReferenceID != nil || len(assessment.EvidenceDigest) != 0 {
		return hrerrors.ErrUnavailable
	}
	if assessment.AssessmentMethod != "manual_observation" || assessment.AssessedAt.After(v.now().Add(5*time.Minute)) {
		return hrerrors.ErrWorkflowRequired
	}
	if assessment.ExpiresOn != nil && entity.CompareDates(*assessment.ExpiresOn, assessment.AssessedAt) < 0 {
		return hrerrors.ErrInvalidInput
	}
	worker, err := v.workers.FindByIDForUpdate(ctx, assessment.WorkerID)
	if err != nil {
		return workerAssessmentReferenceError(err)
	}
	if worker == nil || worker.Status != entity.WorkerStatusActive ||
		(worker.HireDate != nil && entity.CompareDates(assessment.AssessedAt, *worker.HireDate) < 0) ||
		(worker.TerminationDate != nil && entity.CompareDates(assessment.AssessedAt, *worker.TerminationDate) > 0) {
		return hrerrors.ErrReferenceConflict
	}
	competency, err := v.competencies.FindByIDForUpdate(ctx, assessment.CompetencyID)
	if err != nil {
		return workerAssessmentReferenceError(err)
	}
	if competency == nil || !competency.IsActive {
		return hrerrors.ErrReferenceConflict
	}
	return nil
}

func workerAssessmentReferenceError(err error) error {
	if errors.Is(err, hrerrors.ErrNotFound) {
		return hrerrors.ErrReferenceConflict
	}
	return err
}
