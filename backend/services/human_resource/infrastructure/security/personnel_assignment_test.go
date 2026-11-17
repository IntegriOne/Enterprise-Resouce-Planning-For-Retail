package security

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
)

type assignmentWorkerStub struct {
	worker  *entity.Worker
	manager *entity.Worker
	err     error
}

func (s *assignmentWorkerStub) FindByIDForUpdate(context.Context, string) (*entity.Worker, error) {
	return s.worker, s.err
}

func (s *assignmentWorkerStub) FindByID(_ context.Context, id string) (*entity.Worker, error) {
	if s.manager != nil && id == s.manager.ID {
		return s.manager, s.err
	}
	return nil, hrerrors.ErrNotFound
}

type assignmentPositionStub struct {
	position *entity.Position
	err      error
}

func (s *assignmentPositionStub) FindByIDForUpdate(context.Context, string) (*entity.Position, error) {
	return s.position, s.err
}

func TestWorkerAssignmentValidatorAcceptsOnlyConsistentDraft(t *testing.T) {
	t.Parallel()
	from := time.Date(2026, 11, 17, 0, 0, 0, 0, time.UTC)
	until := from.AddDate(0, 3, 0)
	requester := "00000000-0000-4000-8000-000000000001"
	legal := "00000000-0000-4000-8000-000000000002"
	workers := &assignmentWorkerStub{worker: &entity.Worker{
		ID: "00000000-0000-4000-8000-000000000003", Status: entity.WorkerStatusActive,
		PrimaryLegalEntityReferenceID: legal, HireDate: &from,
	}}
	positions := &assignmentPositionStub{position: &entity.Position{
		ID: "00000000-0000-4000-8000-000000000004", PositionStatus: "planned",
		LegalEntityReferenceID: legal, ValidFrom: from, ValidTo: &until,
	}}
	v := &workerAssignmentValidator{workers: workers, positions: positions}
	draft := &entity.WorkerAssignment{
		ID: "00000000-0000-4000-8000-000000000005", WorkerID: workers.worker.ID,
		PositionID: positions.position.ID, AssignmentType: "primary", FTEPercent: "100",
		ValidFrom: from, ValidTo: &until, RequestedByPrincipalID: &requester, RequestedAt: &from,
	}
	if err := draft.Validate(); err != nil {
		t.Fatalf("valid draft entity: %v", err)
	}
	unpaired := *draft
	unpaired.RequestedAt = nil
	if err := unpaired.Validate(); err == nil {
		t.Fatal("requester without timestamp passed entity validation")
	}
	selfApproved := *draft
	selfApproved.ApprovedByPrincipalID = &requester
	selfApproved.ApprovedAt = &from
	if err := selfApproved.Validate(); err == nil {
		t.Fatal("self approval passed entity validation")
	}
	approver := "00000000-0000-4000-8000-000000000007"
	approved := *draft
	approved.ApprovedByPrincipalID = &approver
	approved.ApprovedAt = &from
	if err := approved.Validate(); err != nil {
		t.Fatalf("independent approval metadata: %v", err)
	}
	if err := v.ValidateMutation(t.Context(), "hr_worker_assignments", nil, draft); err != nil {
		t.Fatalf("valid draft references: %v", err)
	}
	for _, tc := range []struct {
		name   string
		change func(*entity.WorkerAssignment)
		want   error
	}{
		{"missing requester", func(a *entity.WorkerAssignment) { a.RequestedByPrincipalID = nil }, hrerrors.ErrWorkflowRequired},
		{"client approval", func(a *entity.WorkerAssignment) { a.ApprovedByPrincipalID = &requester }, hrerrors.ErrWorkflowRequired},
		{"status snapshot", func(a *entity.WorkerAssignment) { status := "active"; a.EmploymentStatusSnapshot = &status }, hrerrors.ErrWorkflowRequired},
		{"unknown assignment type", func(a *entity.WorkerAssignment) { a.AssignmentType = "permanent" }, hrerrors.ErrWorkflowRequired},
		{"outside position", func(a *entity.WorkerAssignment) { a.ValidTo = nil }, hrerrors.ErrReferenceConflict},
	} {
		t.Run(tc.name, func(t *testing.T) {
			candidate := *draft
			tc.change(&candidate)
			if err := v.ValidateMutation(t.Context(), "hr_worker_assignments", nil, &candidate); !errors.Is(err, tc.want) {
				t.Fatalf("validation = %v, want %v", err, tc.want)
			}
		})
	}
	if err := v.ValidateMutation(t.Context(), "hr_worker_assignments", draft, draft); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("generic update: %v", err)
	}
	if err := v.ValidateMutation(t.Context(), "hr_worker_assignments", draft, nil); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("generic delete: %v", err)
	}
	workers.worker.Status = entity.WorkerStatusTerminated
	if err := v.ValidateMutation(t.Context(), "hr_worker_assignments", nil, draft); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("terminated worker: %v", err)
	}
	workers.worker.Status = entity.WorkerStatusActive
	positions.position.LegalEntityReferenceID = "00000000-0000-4000-8000-000000000006"
	if err := v.ValidateMutation(t.Context(), "hr_worker_assignments", nil, draft); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("wrong legal entity: %v", err)
	}
	positions.position.LegalEntityReferenceID = legal
	workers.err = hrerrors.ErrNotFound
	if err := v.ValidateMutation(t.Context(), "hr_worker_assignments", nil, draft); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("missing worker: %v", err)
	}
}

func TestWorkerAssignmentValidatorChecksManager(t *testing.T) {
	t.Parallel()
	from := time.Date(2026, 11, 17, 0, 0, 0, 0, time.UTC)
	legal := "00000000-0000-4000-8000-000000000002"
	managerID := "00000000-0000-4000-8000-000000000006"
	requester := "00000000-0000-4000-8000-000000000001"
	workers := &assignmentWorkerStub{worker: &entity.Worker{ID: "00000000-0000-4000-8000-000000000003",
		Status: entity.WorkerStatusActive, PrimaryLegalEntityReferenceID: legal},
		manager: &entity.Worker{ID: managerID, Status: entity.WorkerStatusActive, PrimaryLegalEntityReferenceID: legal}}
	v := &workerAssignmentValidator{workers: workers, positions: &assignmentPositionStub{position: &entity.Position{
		ID: "00000000-0000-4000-8000-000000000004", PositionStatus: "active",
		LegalEntityReferenceID: legal, ValidFrom: from,
	}}}
	draft := &entity.WorkerAssignment{WorkerID: workers.worker.ID, PositionID: "00000000-0000-4000-8000-000000000004",
		ManagerWorkerID: &managerID, AssignmentType: "secondary", ValidFrom: from,
		RequestedByPrincipalID: &requester, RequestedAt: &from}
	if err := v.ValidateMutation(t.Context(), "hr_worker_assignments", nil, draft); err != nil {
		t.Fatalf("active manager: %v", err)
	}
	workers.manager.Status = entity.WorkerStatusTerminated
	if err := v.ValidateMutation(t.Context(), "hr_worker_assignments", nil, draft); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("inactive manager: %v", err)
	}
}
