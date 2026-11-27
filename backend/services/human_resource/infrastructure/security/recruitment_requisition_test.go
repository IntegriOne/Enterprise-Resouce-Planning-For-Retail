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
		CreatedByPrincipalID: &principal,
		Headcount:            1, TargetStartDate: &target, EmploymentCategory: entity.WorkerCategoryEmployee,
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
	submitted := *draft
	submitted.Status = "submitted"
	submittedAt := time.Now()
	submitted.SubmittedAt = &submittedAt
	submitted.SubmittedByPrincipalID = &principal
	if err := v.ValidateMutation(t.Context(), "hr_requisitions", draft, &submitted); err != nil {
		t.Fatalf("submit valid draft: %v", err)
	}
	tamperedSubmission := submitted
	tamperedSubmission.Headcount++
	if err := v.ValidateMutation(t.Context(), "hr_requisitions", draft, &tamperedSubmission); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("changed headcount during submit: %v", err)
	}
	otherSubmitter := submitted
	otherPrincipal := "00000000-0000-4000-8000-000000000010"
	otherSubmitter.SubmittedByPrincipalID = &otherPrincipal
	if err := v.ValidateMutation(t.Context(), "hr_requisitions", draft, &otherSubmitter); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("submit by another principal: %v", err)
	}
	approved := submitted
	approved.Status = "approved"
	budgetProof := "00000000-0000-4000-8000-000000000020"
	approver := "00000000-0000-4000-8000-000000000021"
	approved.BudgetReferenceID = &budgetProof
	approved.ApprovedAt = &submittedAt
	approved.ApprovedByPrincipalID = &approver
	if err := v.ValidateMutation(t.Context(), "hr_requisitions", &submitted, &approved); err != nil {
		t.Fatalf("approve submitted requisition: %v", err)
	}
	selfApproved := approved
	selfApproved.ApprovedByPrincipalID = &principal
	if err := v.ValidateMutation(t.Context(), "hr_requisitions", &submitted, &selfApproved); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("self approval: %v", err)
	}
	changedDuringApproval := approved
	changedDuringApproval.Headcount++
	if err := v.ValidateMutation(t.Context(), "hr_requisitions", &submitted, &changedDuringApproval); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("changed headcount during approval: %v", err)
	}
	rejected := submitted
	rejected.Status = "rejected"
	reason := "BUDGET_DENIED"
	rejected.RejectedAt = &submittedAt
	rejected.RejectionReasonCode = &reason
	rejected.RejectedByPrincipalID = &approver
	if err := v.ValidateMutation(t.Context(), "hr_requisitions", &submitted, &rejected); err != nil {
		t.Fatalf("reject submitted requisition: %v", err)
	}
	selfRejected := rejected
	selfRejected.RejectedByPrincipalID = &principal
	if err := v.ValidateMutation(t.Context(), "hr_requisitions", &submitted, &selfRejected); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("self rejection: %v", err)
	}
	tamperedRejection := rejected
	tamperedRejection.Headcount++
	if err := v.ValidateMutation(t.Context(), "hr_requisitions", &submitted, &tamperedRejection); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("changed headcount during rejection: %v", err)
	}
	if err := v.ValidateMutation(t.Context(), "hr_requisitions", &rejected, &approved); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("reopen rejected requisition: %v", err)
	}
	if err := v.ValidateMutation(t.Context(), "hr_requisitions", &submitted, nil); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("delete submitted requisition: %v", err)
	}
	opened := approved
	opened.Status = "open"
	openedAt := submittedAt.Add(time.Second)
	opened.OpenedAt = &openedAt
	if err := v.ValidateMutation(t.Context(), "hr_requisitions", &approved, &opened); err != nil {
		t.Fatalf("open approved requisition: %v", err)
	}
	backdatedOpening := opened
	beforeApproval := submittedAt.Add(-time.Second)
	backdatedOpening.OpenedAt = &beforeApproval
	if err := v.ValidateMutation(t.Context(), "hr_requisitions", &approved, &backdatedOpening); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("opening before approval: %v", err)
	}
	changedDuringOpening := opened
	changedDuringOpening.Headcount++
	if err := v.ValidateMutation(t.Context(), "hr_requisitions", &approved, &changedDuringOpening); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("changed headcount during open: %v", err)
	}
	if err := v.ValidateMutation(t.Context(), "hr_requisitions", &approved, nil); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("delete approved requisition: %v", err)
	}
	closed := opened
	closed.Status = "closed"
	closedAt := openedAt.Add(time.Second)
	closed.ClosedAt = &closedAt
	if err := v.ValidateMutation(t.Context(), "hr_requisitions", &opened, &closed); err != nil {
		t.Fatalf("close open requisition: %v", err)
	}
	applications.count = 1
	if err := v.ValidateMutation(t.Context(), "hr_requisitions", &opened, &closed); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("close with active application: %v", err)
	}
	applications.count = 0
	postings.count = 1
	if err := v.ValidateMutation(t.Context(), "hr_requisitions", &opened, &closed); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("close with published posting: %v", err)
	}
	postings.count = 0
	changedDuringClose := closed
	changedDuringClose.Headcount++
	if err := v.ValidateMutation(t.Context(), "hr_requisitions", &opened, &changedDuringClose); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("changed headcount during close: %v", err)
	}
	if err := v.ValidateMutation(t.Context(), "hr_requisitions", &opened, nil); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("delete open requisition: %v", err)
	}
	changedCreator := updated
	changedCreator.CreatedByPrincipalID = &otherPrincipal
	if err := v.ValidateMutation(t.Context(), "hr_requisitions", draft, &changedCreator); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("changed creator: %v", err)
	}
	withSubmission := updated
	withSubmission.SubmittedAt = &target
	withSubmission.SubmittedByPrincipalID = &principal
	if err := v.ValidateMutation(t.Context(), "hr_requisitions", draft, &withSubmission); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("submission through CRUD: %v", err)
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
	if err := v.ValidateMutation(t.Context(), "hr_requisitions", draft, &submitted); !errors.Is(err, hrerrors.ErrUnavailable) {
		t.Fatalf("submission without owner verifier: %v", err)
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

type requisitionPlanLineStub struct{ line *entity.WorkforcePlanLine }

func (s requisitionPlanLineStub) FindByID(context.Context, string) (*entity.WorkforcePlanLine, error) {
	return s.line, nil
}

func TestRequisitionWorkforceLineRequiresCurrentApprovedPlan(t *testing.T) {
	t.Parallel()
	today := time.Date(2026, 11, 25, 0, 0, 0, 0, time.UTC)
	budgetID := "budget"
	approver := "approver"
	line := &entity.WorkforcePlanLine{
		ID: "line", WorkforcePlanID: "plan", JobID: "job", OrganizationUnitReferenceID: "unit",
		WorkerCategory: entity.WorkerCategoryEmployee, ActionCode: "recruit", NeededFrom: today,
	}
	plan := &entity.WorkforcePlan{
		ID: "plan", LegalEntityReferenceID: "legal", Status: "approved", PeriodStart: today, PeriodEnd: today.AddDate(0, 1, 0),
		ApprovedAt: &today, ApprovedByPrincipalID: &approver, FinanceBudgetReferenceID: &budgetID,
	}
	value := &entity.Requisition{
		WorkforcePlanLineID: &line.ID, JobID: line.JobID, OrganizationUnitReferenceID: line.OrganizationUnitReferenceID,
		EmploymentCategory: line.WorkerCategory,
	}
	v := &requisitionValidator{plans: &workforcePlanStub{plan: plan}, planLines: requisitionPlanLineStub{line}}
	if err := v.verifyWorkforceLine(t.Context(), value, "legal", today); err != nil {
		t.Fatalf("approved workforce line: %v", err)
	}
	if err := v.verifyWorkforceLine(t.Context(), value, "other", today); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("workforce line from another legal entity: %v", err)
	}
	if err := v.verifyWorkforceLine(t.Context(), value, "legal", plan.PeriodEnd.AddDate(0, 0, 1)); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("workforce line after plan period: %v", err)
	}
	plan.Status = "superseded"
	if err := v.verifyWorkforceLine(t.Context(), value, "legal", today); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("superseded workforce line: %v", err)
	}
}
