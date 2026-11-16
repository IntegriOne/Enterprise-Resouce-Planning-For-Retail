package security

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
)

type workerNumberStub struct{ matches []entity.Worker }

func (s *workerNumberStub) FindByWorkerNumber(context.Context, string, int, int) ([]entity.Worker, error) {
	return s.matches, nil
}

type workerChildStub struct{ count int64 }

func (s *workerChildStub) CountWhere(context.Context, []domainrepo.Filter) (int64, error) {
	return s.count, nil
}

type workerOwnerStub struct {
	check WorkerReferenceCheck
	err   error
}

func (s *workerOwnerStub) VerifyWorkerReferences(_ context.Context, check WorkerReferenceCheck) error {
	s.check = check
	return s.err
}

func TestWorkerValidatorKeepsPreboardingDraftSeparateFromIdentityAndEmployment(t *testing.T) {
	t.Parallel()
	rows := &workerNumberStub{}
	contracts, assignments, onboarding := &workerChildStub{}, &workerChildStub{}, &workerChildStub{}
	owner := &workerOwnerStub{}
	v := &workerValidator{workers: rows, contracts: contracts, assignments: assignments, onboarding: onboarding, owner: owner}
	now := time.Now()
	hire := now.AddDate(0, 0, 5)
	serviceStart := now.AddDate(0, 0, 6)
	creator := "00000000-0000-4000-8000-000000000001"
	worker := &entity.Worker{
		ID: "00000000-0000-4000-8000-000000000002", WorkerNumber: "WORKER-DRAFT-1",
		PersonPartyReferenceID:        "00000000-0000-4000-8000-000000000003",
		PrimaryLegalEntityReferenceID: "00000000-0000-4000-8000-000000000004",
		WorkerCategory:                entity.WorkerCategoryEmployee, Status: entity.WorkerStatusPreboarding,
		HireDate: &hire, ServiceStartDate: &serviceStart, CreatedAt: now, CreatedByPrincipalID: &creator,
	}
	if err := v.ValidateMutation(t.Context(), "hr_workers", nil, worker); err != nil {
		t.Fatalf("create preboarding worker: %v", err)
	}
	if owner.check.PersonPartyID != worker.PersonPartyReferenceID || owner.check.LegalEntityID != worker.PrimaryLegalEntityReferenceID {
		t.Fatalf("wrong Foundation reference check: %+v", owner.check)
	}
	updated := *worker
	later := now.AddDate(0, 0, 7)
	updated.ServiceStartDate = &later
	if err := v.ValidateMutation(t.Context(), "hr_workers", worker, &updated); err != nil {
		t.Fatalf("correct preboarding date: %v", err)
	}
	early := *worker
	early.ServiceStartDate = &now
	if err := v.ValidateMutation(t.Context(), "hr_workers", worker, &early); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("service date before hire: %v", err)
	}
	contracts.count = 1
	if err := v.ValidateMutation(t.Context(), "hr_workers", worker, &updated); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("date correction after contract: %v", err)
	}
	contracts.count = 0
	linked := *worker
	linked.IAMPrincipalReferenceID = &creator
	if err := v.ValidateMutation(t.Context(), "hr_workers", nil, &linked); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("IAM link through CRUD: %v", err)
	}
	active := *worker
	active.Status = entity.WorkerStatusActive
	if err := v.ValidateMutation(t.Context(), "hr_workers", worker, &active); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("activation through CRUD: %v", err)
	}
	if err := v.ValidateMutation(t.Context(), "hr_workers", worker, nil); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("delete employment history: %v", err)
	}
	rows.matches = []entity.Worker{{ID: "00000000-0000-4000-8000-000000000005"}}
	if err := v.ValidateMutation(t.Context(), "hr_workers", nil, worker); !errors.Is(err, hrerrors.ErrAlreadyExists) {
		t.Fatalf("duplicate worker number: %v", err)
	}
	rows.matches = nil
	owner.err = errors.New("foundation unavailable")
	if err := v.ValidateMutation(t.Context(), "hr_workers", nil, worker); !errors.Is(err, hrerrors.ErrUnavailable) {
		t.Fatalf("Foundation outage: %v", err)
	}
	owner.err = nil
	v.owner = nil
	if err := v.ValidateMutation(t.Context(), "hr_workers", nil, worker); !errors.Is(err, hrerrors.ErrUnavailable) {
		t.Fatalf("missing Foundation verifier: %v", err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if err := v.ValidateMutation(ctx, "hr_workers", nil, worker); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled validation: %v", err)
	}
}
