package security

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
)

type assessmentWorkerStub struct {
	worker *entity.Worker
	err    error
}

func (s *assessmentWorkerStub) FindByIDForUpdate(context.Context, string) (*entity.Worker, error) {
	return s.worker, s.err
}

type assessmentCompetencyStub struct {
	competency *entity.Competency
	err        error
}

func (s *assessmentCompetencyStub) FindByIDForUpdate(context.Context, string) (*entity.Competency, error) {
	return s.competency, s.err
}

func TestWorkerCompetencyAssessmentValidator(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 11, 16, 14, 0, 0, 0, time.UTC)
	hired := now.AddDate(-1, 0, 0)
	workers := &assessmentWorkerStub{worker: &entity.Worker{ID: "worker-1", Status: entity.WorkerStatusActive, HireDate: &hired}}
	competencies := &assessmentCompetencyStub{competency: &entity.Competency{ID: "competency-1", IsActive: true}}
	validator := &workerCompetencyAssessmentValidator{workers: workers, competencies: competencies, now: func() time.Time { return now }}
	assessment := &entity.WorkerCompetencyAssessment{ID: "assessment-1", WorkerID: "worker-1", CompetencyID: "competency-1",
		AssessmentMethod: "manual_observation", AssessedAt: now.Add(-time.Hour), AssessedLevel: 3}
	if err := validator.ValidateMutation(t.Context(), "hr_worker_competency_assessments", nil, assessment); err != nil {
		t.Fatalf("manual observation: %v", err)
	}
	if err := validator.ValidateMutation(t.Context(), "hr_worker_competency_assessments", assessment, assessment); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("overwrite assessment: %v", err)
	}
	if err := validator.ValidateMutation(t.Context(), "hr_worker_competency_assessments", assessment, nil); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("delete assessment: %v", err)
	}
	withEvidence := *assessment
	evidenceID := "document-1"
	withEvidence.EvidenceDocumentVersionReferenceID = &evidenceID
	if err := validator.ValidateMutation(t.Context(), "hr_worker_competency_assessments", nil, &withEvidence); !errors.Is(err, hrerrors.ErrUnavailable) {
		t.Fatalf("unverified document: %v", err)
	}
	future := *assessment
	future.AssessedAt = now.Add(time.Hour)
	if err := validator.ValidateMutation(t.Context(), "hr_worker_competency_assessments", nil, &future); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("future assessment: %v", err)
	}
	badMethod := *assessment
	badMethod.AssessmentMethod = "certification"
	if err := validator.ValidateMutation(t.Context(), "hr_worker_competency_assessments", nil, &badMethod); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("unverified method: %v", err)
	}
	expired := *assessment
	expires := hired.AddDate(0, -1, 0)
	expired.ExpiresOn = &expires
	if err := validator.ValidateMutation(t.Context(), "hr_worker_competency_assessments", nil, &expired); !errors.Is(err, hrerrors.ErrInvalidInput) {
		t.Fatalf("expiry before assessment: %v", err)
	}
	workers.worker.Status = entity.WorkerStatusTerminated
	if err := validator.ValidateMutation(t.Context(), "hr_worker_competency_assessments", nil, assessment); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("inactive worker: %v", err)
	}
	workers.worker.Status = entity.WorkerStatusActive
	competencies.competency.IsActive = false
	if err := validator.ValidateMutation(t.Context(), "hr_worker_competency_assessments", nil, assessment); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("inactive competency: %v", err)
	}
	competencies.competency.IsActive = true
	workers.err = hrerrors.ErrNotFound
	if err := validator.ValidateMutation(t.Context(), "hr_worker_competency_assessments", nil, assessment); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("missing worker: %v", err)
	}
	workers.err = nil
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if err := validator.ValidateMutation(ctx, "hr_worker_competency_assessments", nil, assessment); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled context: %v", err)
	}
}
