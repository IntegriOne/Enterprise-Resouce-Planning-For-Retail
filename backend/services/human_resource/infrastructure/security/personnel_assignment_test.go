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

type assignmentListStub struct {
	workerRows   []entity.WorkerAssignment
	positionRows []entity.WorkerAssignment
}

func (s *assignmentListStub) CountWhere(_ context.Context, filters []domainrepo.Filter) (int64, error) {
	if len(filters) != 1 {
		return 0, hrerrors.ErrInvalidInput
	}
	if filters[0].Column == "worker_id" {
		return int64(len(s.workerRows)), nil
	}
	return int64(len(s.positionRows)), nil
}

func (s *assignmentListStub) FindByWorkerID(_ context.Context, _ string, limit, offset int) ([]entity.WorkerAssignment, error) {
	return assignmentPage(s.workerRows, limit, offset), nil
}

func (s *assignmentListStub) FindByPositionID(_ context.Context, _ string, limit, offset int) ([]entity.WorkerAssignment, error) {
	return assignmentPage(s.positionRows, limit, offset), nil
}

func assignmentPage(rows []entity.WorkerAssignment, limit, offset int) []entity.WorkerAssignment {
	if offset >= len(rows) {
		return nil
	}
	end := min(offset+limit, len(rows))
	return rows[offset:end]
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
	serviceStart := from.AddDate(0, 0, 1)
	workers.worker.ServiceStartDate = &serviceStart
	if err := v.ValidateMutation(t.Context(), "hr_worker_assignments", nil, draft); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("assignment before service starts: %v", err)
	}
	atServiceStart := *draft
	atServiceStart.ValidFrom = serviceStart
	if err := v.ValidateMutation(t.Context(), "hr_worker_assignments", nil, &atServiceStart); err != nil {
		t.Fatalf("assignment on service start date: %v", err)
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
	workers.manager.Status = entity.WorkerStatusActive
	nextDay := from.AddDate(0, 0, 1)
	workers.manager.HireDate = &nextDay
	if err := v.ValidateMutation(t.Context(), "hr_worker_assignments", nil, draft); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("manager hired after assignment starts: %v", err)
	}
	workers.manager.HireDate = &from
	workers.manager.ServiceStartDate = &nextDay
	if err := v.ValidateMutation(t.Context(), "hr_worker_assignments", nil, draft); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("manager service has not started: %v", err)
	}
	workers.manager.ServiceStartDate = nil
	workers.manager.TerminationDate = &nextDay
	if err := v.ValidateMutation(t.Context(), "hr_worker_assignments", nil, draft); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("open-ended assignment with exiting manager: %v", err)
	}
	draft.ValidTo = &nextDay
	if err := v.ValidateMutation(t.Context(), "hr_worker_assignments", nil, draft); err != nil {
		t.Fatalf("assignment within manager employment period: %v", err)
	}
	workers.manager = workers.worker
	draft.ManagerWorkerID = &workers.worker.ID
	if err := v.ValidateMutation(t.Context(), "hr_worker_assignments", nil, draft); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("worker assigned as their own manager: %v", err)
	}
}

func TestWorkerAssignmentApprovalChecksTemporalCapacity(t *testing.T) {
	t.Parallel()
	from := time.Date(2026, 11, 17, 0, 0, 0, 0, time.UTC)
	requester := "00000000-0000-4000-8000-000000000001"
	approver := "00000000-0000-4000-8000-000000000002"
	legal := "00000000-0000-4000-8000-000000000003"
	workerID := "00000000-0000-4000-8000-000000000004"
	positionID := "00000000-0000-4000-8000-000000000005"
	worker := &entity.Worker{ID: workerID, Status: entity.WorkerStatusActive, PrimaryLegalEntityReferenceID: legal}
	position := &entity.Position{ID: positionID, PositionStatus: "active", LegalEntityReferenceID: legal,
		HeadcountAuthorized: "1", ValidFrom: from}
	rows := &assignmentListStub{}
	v := &workerAssignmentValidator{workers: &assignmentWorkerStub{worker: worker},
		positions: &assignmentPositionStub{position: position}, assignments: rows}
	draft := &entity.WorkerAssignment{ID: "00000000-0000-4000-8000-000000000006", WorkerID: workerID,
		PositionID: positionID, AssignmentType: "secondary", FTEPercent: "60", ValidFrom: from,
		RequestedByPrincipalID: &requester, RequestedAt: &from}
	approved := *draft
	status := string(entity.WorkerStatusActive)
	approved.EmploymentStatusSnapshot = &status
	approved.ApprovedByPrincipalID = &approver
	approvedAt := from.Add(time.Hour)
	approved.ApprovedAt = &approvedAt
	if err := v.ValidateMutation(t.Context(), "hr_worker_assignments", draft, &approved); err != nil {
		t.Fatalf("approve empty capacity: %v", err)
	}
	if err := v.ValidateMutation(t.Context(), "hr_worker_assignments", nil, &approved); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("client approval: %v", err)
	}
	if err := v.ValidateMutation(t.Context(), "hr_worker_assignments", draft, draft); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("generic update: %v", err)
	}
	other := approved
	other.ID = "00000000-0000-4000-8000-000000000007"
	other.FTEPercent = "50"
	rows.workerRows = []entity.WorkerAssignment{other}
	if err := v.ValidateMutation(t.Context(), "hr_worker_assignments", draft, &approved); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("worker FTE overflow: %v", err)
	}
	other.FTEPercent = "40"
	rows.workerRows = []entity.WorkerAssignment{other}
	rows.positionRows = []entity.WorkerAssignment{other}
	if err := v.ValidateMutation(t.Context(), "hr_worker_assignments", draft, &approved); err != nil {
		t.Fatalf("worker and position exact capacity: %v", err)
	}
	other.WorkerID = "00000000-0000-4000-8000-000000000008"
	other.FTEPercent = "50"
	rows.workerRows = nil
	rows.positionRows = []entity.WorkerAssignment{other}
	if err := v.ValidateMutation(t.Context(), "hr_worker_assignments", draft, &approved); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("position FTE overflow: %v", err)
	}
	cancelledAt := approvedAt.Add(time.Hour)
	other.CancelledAt = &cancelledAt
	rows.positionRows = []entity.WorkerAssignment{other}
	if err := v.ValidateMutation(t.Context(), "hr_worker_assignments", draft, &approved); err != nil {
		t.Fatalf("cancelled assignment must not consume capacity: %v", err)
	}
	other.CancelledAt = nil
	other.ValidTo = &from
	approved.ValidFrom = from.AddDate(0, 0, 1)
	rows.positionRows = []entity.WorkerAssignment{other}
	if err := v.ValidateMutation(t.Context(), "hr_worker_assignments", draft, &approved); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("period change during approval: %v", err)
	}
	draft.ValidFrom = approved.ValidFrom
	if err := v.ValidateMutation(t.Context(), "hr_worker_assignments", draft, &approved); err != nil {
		t.Fatalf("nonoverlapping position assignment: %v", err)
	}
	position.PositionStatus = "planned"
	if err := v.ValidateMutation(t.Context(), "hr_worker_assignments", draft, &approved); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("planned position approval: %v", err)
	}
}
