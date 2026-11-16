package security

import (
	"context"
	"errors"
	"testing"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
)

func TestWorkerQualificationValidator(t *testing.T) {
	t.Parallel()
	workers := &assessmentWorkerStub{worker: &entity.Worker{ID: "worker-1", Status: entity.WorkerStatusActive}}
	validator := &workerQualificationValidator{workers: workers}
	draft := &entity.WorkerQualification{
		ID: "qualification-1", WorkerID: "worker-1", QualificationType: "certification",
		Title: "Safety training", VerificationStatus: "unverified",
	}
	if err := validator.ValidateMutation(t.Context(), "hr_worker_qualifications", nil, draft); err != nil {
		t.Fatalf("create unverified qualification: %v", err)
	}
	updated := *draft
	updated.Title = "Updated safety training"
	if err := validator.ValidateMutation(t.Context(), "hr_worker_qualifications", draft, &updated); err != nil {
		t.Fatalf("update draft qualification: %v", err)
	}
	if err := validator.ValidateMutation(t.Context(), "hr_worker_qualifications", draft, nil); err != nil {
		t.Fatalf("delete draft qualification: %v", err)
	}

	verified := *draft
	verified.VerificationStatus = "verified"
	if err := validator.ValidateMutation(t.Context(), "hr_worker_qualifications", draft, &verified); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("verification through CRUD: %v", err)
	}
	if err := validator.ValidateMutation(t.Context(), "hr_worker_qualifications", &verified, nil); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("delete verified qualification: %v", err)
	}
	actor := "reviewer-1"
	withActor := *draft
	withActor.VerifiedByPrincipalID = &actor
	if err := validator.ValidateMutation(t.Context(), "hr_worker_qualifications", nil, &withActor); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("client-supplied verifier: %v", err)
	}
	changedWorker := *draft
	changedWorker.WorkerID = "worker-2"
	if err := validator.ValidateMutation(t.Context(), "hr_worker_qualifications", draft, &changedWorker); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("changed worker: %v", err)
	}
	externalID := "issuer-1"
	withIssuer := *draft
	withIssuer.IssuerPartyReferenceID = &externalID
	if err := validator.ValidateMutation(t.Context(), "hr_worker_qualifications", nil, &withIssuer); !errors.Is(err, hrerrors.ErrUnavailable) {
		t.Fatalf("unverified issuer: %v", err)
	}
	workers.worker.Status = entity.WorkerStatusTerminated
	if err := validator.ValidateMutation(t.Context(), "hr_worker_qualifications", nil, draft); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("inactive worker: %v", err)
	}
	workers.err = hrerrors.ErrNotFound
	if err := validator.ValidateMutation(t.Context(), "hr_worker_qualifications", nil, draft); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("missing worker: %v", err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if err := validator.ValidateMutation(ctx, "hr_worker_qualifications", nil, draft); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled context: %v", err)
	}
}
