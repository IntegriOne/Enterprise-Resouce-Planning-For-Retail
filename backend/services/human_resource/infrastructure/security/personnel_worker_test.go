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

type workerAssignmentLookupStub struct {
	rows map[string]*entity.WorkerAssignment
}

func (s *workerAssignmentLookupStub) FindByID(_ context.Context, id string) (*entity.WorkerAssignment, error) {
	if row := s.rows[id]; row != nil {
		return row, nil
	}
	return nil, hrerrors.ErrNotFound
}

type workerPositionLookupStub struct{ row *entity.Position }

func (s *workerPositionLookupStub) FindByID(context.Context, string) (*entity.Position, error) {
	return s.row, nil
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

func TestWorkerValidatorAppliesOnlyApprovedCurrentPrimaryAssignment(t *testing.T) {
	t.Parallel()
	day := time.Date(2026, 11, 17, 10, 0, 0, 0, time.UTC)
	requester := "00000000-0000-4000-8000-000000000001"
	approver := "00000000-0000-4000-8000-000000000002"
	workerID := "00000000-0000-4000-8000-000000000003"
	assignmentID := "00000000-0000-4000-8000-000000000004"
	positionID := "00000000-0000-4000-8000-000000000005"
	legalID := "00000000-0000-4000-8000-000000000006"
	status := string(entity.WorkerStatusActive)
	start := day.AddDate(0, 0, -1)
	worker := &entity.Worker{ID: workerID, WorkerNumber: "WORKER", PersonPartyReferenceID: requester,
		WorkerCategory: entity.WorkerCategoryEmployee, Status: entity.WorkerStatusActive,
		PrimaryLegalEntityReferenceID: legalID, HireDate: &start, CreatedAt: start, UpdatedAt: start}
	assignment := &entity.WorkerAssignment{ID: assignmentID, WorkerID: workerID, PositionID: positionID,
		AssignmentType: "primary", EmploymentStatusSnapshot: &status, FTEPercent: "100",
		ValidFrom: start, RequestedByPrincipalID: &requester, RequestedAt: &start,
		ApprovedByPrincipalID: &approver, ApprovedAt: &day}
	lookup := &workerAssignmentLookupStub{rows: map[string]*entity.WorkerAssignment{assignmentID: assignment}}
	position := &workerPositionLookupStub{row: &entity.Position{ID: positionID, PositionStatus: "active",
		LegalEntityReferenceID: legalID, ValidFrom: start}}
	owner := &workerOwnerStub{}
	v := &workerValidator{workers: &workerNumberStub{}, contracts: &workerChildStub{}, assignments: &workerChildStub{},
		assignmentLookup: lookup, positions: position, onboarding: &workerChildStub{}, owner: owner}
	applied := *worker
	applied.PrimaryAssignmentID = &assignmentID
	applied.UpdatedAt = day
	if err := v.ValidateMutation(t.Context(), "hr_workers", worker, &applied); err != nil {
		t.Fatalf("approved primary assignment: %v", err)
	}
	if owner.check.LegalEntityID != legalID {
		t.Fatalf("owner proof missing: %+v", owner.check)
	}
	assignment.ApprovedAt = nil
	if err := v.ValidateMutation(t.Context(), "hr_workers", worker, &applied); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("draft assignment: %v", err)
	}
	assignment.ApprovedAt = &day
	lateApproval := day.Add(time.Hour)
	assignment.ApprovedAt = &lateApproval
	if err := v.ValidateMutation(t.Context(), "hr_workers", worker, &applied); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("approval after application: %v", err)
	}
	assignment.ApprovedAt = &day
	assignment.ValidFrom = day.AddDate(0, 0, 1)
	if err := v.ValidateMutation(t.Context(), "hr_workers", worker, &applied); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("future assignment: %v", err)
	}
	assignment.ValidFrom = start
	position.row.PositionStatus = "planned"
	if err := v.ValidateMutation(t.Context(), "hr_workers", worker, &applied); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("planned position: %v", err)
	}
	position.row.PositionStatus = "active"
	changed := applied
	changed.WorkerNumber = "OTHER"
	if err := v.ValidateMutation(t.Context(), "hr_workers", worker, &changed); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("unrelated worker mutation: %v", err)
	}
	prior := *assignment
	prior.ID = "00000000-0000-4000-8000-000000000007"
	prior.ValidTo = nil
	lookup.rows[prior.ID] = &prior
	worker.PrimaryAssignmentID = &prior.ID
	if err := v.ValidateMutation(t.Context(), "hr_workers", worker, &applied); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("replace ongoing primary assignment: %v", err)
	}
	end := start
	prior.ValidTo = &end
	assignment.ValidFrom = day
	if err := v.ValidateMutation(t.Context(), "hr_workers", worker, &applied); err != nil {
		t.Fatalf("replace ended primary assignment: %v", err)
	}
}
