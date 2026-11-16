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

type requisitionWorkerStub struct{ worker *entity.Worker }

func (s *requisitionWorkerStub) SelectWhere(context.Context, []domainrepo.Filter, int, int) ([]entity.Worker, error) {
	if s.worker == nil {
		return nil, nil
	}
	return []entity.Worker{*s.worker}, nil
}
func (s *requisitionWorkerStub) FindByIDForUpdate(context.Context, string) (*entity.Worker, error) {
	return s.worker, nil
}

type requisitionJobStub struct{ job *entity.Job }

func (s *requisitionJobStub) FindByIDForUpdate(context.Context, string) (*entity.Job, error) {
	return s.job, nil
}

type requisitionPositionStub struct{ position *entity.Position }

func (s *requisitionPositionStub) FindByIDForUpdate(context.Context, string) (*entity.Position, error) {
	return s.position, nil
}

type requisitionRowsStub struct{ rows []entity.Requisition }

func (s *requisitionRowsStub) SelectWhere(context.Context, []domainrepo.Filter, int, int) ([]entity.Requisition, error) {
	return s.rows, nil
}

type requisitionCountStub struct{ count int64 }

func (s *requisitionCountStub) CountWhere(context.Context, []domainrepo.Filter) (int64, error) {
	return s.count, nil
}

func TestRequisitionValidatorKeepsDraftReferencesAndChildren(t *testing.T) {
	t.Parallel()
	now := time.Now()
	start := now.AddDate(0, -1, 0)
	target := now.AddDate(0, 0, 10)
	principal := "00000000-0000-4000-8000-000000000001"
	worker := &requisitionWorkerStub{worker: &entity.Worker{
		ID: "00000000-0000-4000-8000-000000000002", Status: entity.WorkerStatusActive,
		IAMPrincipalReferenceID: &principal, PrimaryLegalEntityReferenceID: "00000000-0000-4000-8000-000000000003",
	}}
	job := &requisitionJobStub{job: &entity.Job{ID: "00000000-0000-4000-8000-000000000004", IsActive: true, ValidFrom: start}}
	positionID := "00000000-0000-4000-8000-000000000005"
	unitID := "00000000-0000-4000-8000-000000000006"
	position := &requisitionPositionStub{position: &entity.Position{ID: positionID, PositionStatus: "active",
		JobID: job.job.ID, OrganizationUnitReferenceID: unitID,
		LegalEntityReferenceID: worker.worker.PrimaryLegalEntityReferenceID, ValidFrom: start}}
	rows := &requisitionRowsStub{}
	applications, postings := &requisitionCountStub{}, &requisitionCountStub{}
	owner := &positionOwnerStub{}
	v := &requisitionValidator{workers: worker, jobs: job, positions: position, requisitions: rows,
		applications: applications, postings: postings, owner: owner}
	draft := &entity.Requisition{ID: "00000000-0000-4000-8000-000000000007", RequisitionNumber: "REQ-1",
		JobID: job.job.ID, PositionID: &positionID, OrganizationUnitReferenceID: unitID,
		HiringManagerPrincipalID: principal, RequisitionType: "replacement", Status: "draft",
		Headcount: 1, TargetStartDate: &target, EmploymentCategory: entity.WorkerCategoryEmployee,
		MinimumQualifications: entity.JSON(`{}`), DesiredQualifications: entity.JSON(`{}`), IdempotencyKey: "REQ-1"}
	if err := v.ValidateMutation(t.Context(), "hr_requisitions", nil, draft); err != nil {
		t.Fatalf("create valid draft: %v", err)
	}
	if owner.check.OrganizationUnitID != unitID || owner.check.LegalEntityID != worker.worker.PrimaryLegalEntityReferenceID {
		t.Fatalf("wrong Foundation verification input: %+v", owner.check)
	}
	updated := *draft
	updated.Headcount = 2
	if err := v.ValidateMutation(t.Context(), "hr_requisitions", draft, &updated); err != nil {
		t.Fatalf("update valid draft: %v", err)
	}
	if err := v.ValidateMutation(t.Context(), "hr_requisitions", draft, nil); err != nil {
		t.Fatalf("delete unreferenced draft: %v", err)
	}
	rows.rows = []entity.Requisition{*draft}
	if err := v.ValidateMutation(t.Context(), "hr_requisitions", nil, draft); !errors.Is(err, hrerrors.ErrAlreadyExists) {
		t.Fatalf("duplicate number: %v", err)
	}
	rows.rows = nil
	applications.count = 1
	if err := v.ValidateMutation(t.Context(), "hr_requisitions", draft, nil); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("delete referenced draft: %v", err)
	}
	applications.count = 0
	owner.err = errors.New("foundation unavailable")
	if err := v.ValidateMutation(t.Context(), "hr_requisitions", nil, draft); !errors.Is(err, hrerrors.ErrUnavailable) {
		t.Fatalf("Foundation outage: %v", err)
	}
	owner.err = nil
	v.owner = nil
	if err := v.ValidateMutation(t.Context(), "hr_requisitions", nil, draft); !errors.Is(err, hrerrors.ErrUnavailable) {
		t.Fatalf("missing owner verifier: %v", err)
	}
	if err := v.ValidateMutation(t.Context(), "hr_requisitions", draft, nil); err != nil {
		t.Fatalf("delete unreferenced draft without owner verifier: %v", err)
	}
	v.owner = owner
	withBudget := *draft
	budget := "00000000-0000-4000-8000-000000000008"
	withBudget.BudgetReferenceID = &budget
	if err := v.ValidateMutation(t.Context(), "hr_requisitions", nil, &withBudget); !errors.Is(err, hrerrors.ErrUnavailable) {
		t.Fatalf("unverified Finance budget: %v", err)
	}
	worker.worker.Status = entity.WorkerStatusTerminated
	if err := v.ValidateMutation(t.Context(), "hr_requisitions", nil, draft); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("inactive hiring manager: %v", err)
	}
	worker.worker.Status = entity.WorkerStatusActive
	position.position.OrganizationUnitReferenceID = "00000000-0000-4000-8000-000000000009"
	if err := v.ValidateMutation(t.Context(), "hr_requisitions", nil, draft); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("position from another unit: %v", err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if err := v.ValidateMutation(ctx, "hr_requisitions", nil, draft); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled validation: %v", err)
	}
}
