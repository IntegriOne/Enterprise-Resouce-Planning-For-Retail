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

type positionJobStub struct {
	job *entity.Job
	err error
}

func (s *positionJobStub) FindByIDForUpdate(context.Context, string) (*entity.Job, error) {
	return s.job, s.err
}

type positionReaderStub struct {
	parent  *entity.Position
	matches []entity.Position
	count   int64
}

func (s *positionReaderStub) FindByIDForUpdate(context.Context, string) (*entity.Position, error) {
	return s.parent, nil
}

func (s *positionReaderStub) FindByPositionCode(context.Context, string, int, int) ([]entity.Position, error) {
	return s.matches, nil
}

func (s *positionReaderStub) CountWhere(context.Context, []domainrepo.Filter) (int64, error) {
	return s.count, nil
}

type positionCountStub struct{ count int64 }

func (s *positionCountStub) CountWhere(context.Context, []domainrepo.Filter) (int64, error) {
	return s.count, nil
}

type positionOwnerStub struct {
	check PositionReferenceCheck
	err   error
}

func (s *positionOwnerStub) VerifyPositionReferences(_ context.Context, check PositionReferenceCheck) error {
	s.check = check
	return s.err
}

func TestPositionValidatorPlannedDraft(t *testing.T) {
	t.Parallel()
	start := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	jobs := &positionJobStub{job: &entity.Job{ID: "job", IsActive: true, ValidFrom: start}}
	positions := &positionReaderStub{}
	assignments := &positionCountStub{}
	requisitions := &positionCountStub{}
	succession := &positionCountStub{}
	owner := &positionOwnerStub{}
	validator := &positionValidator{
		jobs: jobs, positions: positions, assignments: assignments,
		requisitions: requisitions, succession: succession, owner: owner,
	}
	draft := &entity.Position{
		ID: "position", PositionCode: "STORE", PositionTitle: "Store associate",
		JobID: "job", OrganizationUnitReferenceID: "unit", LegalEntityReferenceID: "legal",
		PositionStatus: "planned", ValidFrom: start,
	}
	if err := validator.ValidateMutation(t.Context(), "hr_positions", nil, draft); err != nil {
		t.Fatalf("create planned position: %v", err)
	}
	if owner.check.OrganizationUnitID != "unit" || owner.check.LegalEntityID != "legal" {
		t.Fatalf("Foundation verifier received wrong scope: %+v", owner.check)
	}
	updated := *draft
	updated.PositionTitle = "Senior store associate"
	if err := validator.ValidateMutation(t.Context(), "hr_positions", draft, &updated); err != nil {
		t.Fatalf("update planned position: %v", err)
	}
	if err := validator.ValidateMutation(t.Context(), "hr_positions", draft, nil); err != nil {
		t.Fatalf("delete unreferenced planned position: %v", err)
	}
	positions.matches = []entity.Position{{ID: "other"}}
	if err := validator.ValidateMutation(t.Context(), "hr_positions", nil, draft); !errors.Is(err, hrerrors.ErrAlreadyExists) {
		t.Fatalf("duplicate position code = %v", err)
	}
	positions.matches = nil
	assignments.count = 1
	if err := validator.ValidateMutation(t.Context(), "hr_positions", draft, nil); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("delete assigned position = %v", err)
	}
	assignments.count = 0
	owner.err = errors.New("foundation unavailable")
	if err := validator.ValidateMutation(t.Context(), "hr_positions", nil, draft); !errors.Is(err, hrerrors.ErrUnavailable) {
		t.Fatalf("Foundation outage = %v", err)
	}
	owner.err = hrerrors.ErrReferenceConflict
	if err := validator.ValidateMutation(t.Context(), "hr_positions", nil, draft); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("invalid Foundation relation = %v", err)
	}
}

func TestPositionValidatorRejectsUnverifiedAndWorkflowMutation(t *testing.T) {
	t.Parallel()
	start := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	jobs := &positionJobStub{job: &entity.Job{ID: "job", IsActive: true, ValidFrom: start}}
	positions := &positionReaderStub{}
	validator := &positionValidator{
		jobs: jobs, positions: positions, assignments: &positionCountStub{},
		requisitions: &positionCountStub{}, succession: &positionCountStub{},
	}
	draft := &entity.Position{
		ID: "position", PositionCode: "STORE", JobID: "job", PositionStatus: "planned",
		OrganizationUnitReferenceID: "unit", LegalEntityReferenceID: "legal", ValidFrom: start,
	}
	if err := validator.ValidateMutation(t.Context(), "hr_positions", nil, draft); !errors.Is(err, hrerrors.ErrUnavailable) {
		t.Fatalf("missing owner verifier = %v", err)
	}
	active := *draft
	active.PositionStatus = "active"
	if err := validator.ValidateMutation(t.Context(), "hr_positions", draft, &active); !errors.Is(err, hrerrors.ErrUnavailable) {
		t.Fatalf("activation without owner verifier = %v", err)
	}
	changedParent := *draft
	parentID := "parent"
	changedParent.ReportsToPositionID = &parentID
	if err := validator.ValidateMutation(t.Context(), "hr_positions", draft, &changedParent); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("change reporting line through CRUD = %v", err)
	}
	withCostCenter := *draft
	withCostCenter.CostCenterReferenceID = &parentID
	if err := validator.ValidateMutation(t.Context(), "hr_positions", nil, &withCostCenter); !errors.Is(err, hrerrors.ErrUnavailable) {
		t.Fatalf("unverified Finance cost center = %v", err)
	}
	validator.owner = &positionOwnerStub{}
	jobs.job.IsActive = false
	if err := validator.ValidateMutation(t.Context(), "hr_positions", nil, draft); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("inactive job = %v", err)
	}
	jobs.job.IsActive = true
	positions.parent = &entity.Position{
		ID: parentID, LegalEntityReferenceID: "other", PositionStatus: "active", ValidFrom: start,
	}
	if err := validator.ValidateMutation(t.Context(), "hr_positions", nil, &changedParent); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("manager position in another employer = %v", err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if err := validator.ValidateMutation(ctx, "hr_positions", nil, draft); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled context = %v", err)
	}
}

func TestPositionValidatorActivationRechecksReferences(t *testing.T) {
	t.Parallel()
	start := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	parentID := "parent"
	positions := &positionReaderStub{parent: &entity.Position{ID: parentID, LegalEntityReferenceID: "legal",
		PositionStatus: "planned", ValidFrom: start}}
	owner := &positionOwnerStub{}
	validator := &positionValidator{
		jobs:      &positionJobStub{job: &entity.Job{ID: "job", IsActive: true, ValidFrom: start}},
		positions: positions, assignments: &positionCountStub{count: 1}, requisitions: &positionCountStub{},
		succession: &positionCountStub{}, owner: owner,
	}
	draft := &entity.Position{ID: "position", PositionCode: "SEAT", PositionTitle: "Store associate",
		JobID: "job", OrganizationUnitReferenceID: "unit", LegalEntityReferenceID: "legal",
		ReportsToPositionID: &parentID, PositionStatus: "planned", HeadcountAuthorized: "1", ValidFrom: start}
	active := *draft
	active.PositionStatus = "active"
	if err := validator.ValidateMutation(t.Context(), "hr_positions", draft, &active); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("planned parent activation: %v", err)
	}
	positions.parent.PositionStatus = "active"
	if err := validator.ValidateMutation(t.Context(), "hr_positions", draft, &active); err != nil {
		t.Fatalf("activation with verified references: %v", err)
	}
	if owner.check.LegalEntityID != "legal" || owner.check.OrganizationUnitID != "unit" {
		t.Fatalf("activation skipped owner verification: %+v", owner.check)
	}
	changed := active
	changed.HeadcountAuthorized = "2"
	if err := validator.ValidateMutation(t.Context(), "hr_positions", draft, &changed); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("activation changed seat capacity: %v", err)
	}
	if err := validator.ValidateMutation(t.Context(), "hr_positions", &active, nil); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("delete active position: %v", err)
	}
}
