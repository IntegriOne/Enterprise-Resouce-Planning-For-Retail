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

type postingRequisitionStub struct {
	requisition *entity.Requisition
	locks       int
}

func (s *postingRequisitionStub) FindByIDForUpdate(_ context.Context, id string) (*entity.Requisition, error) {
	s.locks++
	if s.requisition == nil || s.requisition.ID != id {
		return nil, hrerrors.ErrNotFound
	}
	return s.requisition, nil
}

type postingApplicationStub struct{ count int64 }

func (s *postingApplicationStub) CountWhere(context.Context, []domainrepo.Filter) (int64, error) {
	return s.count, nil
}

func TestJobPostingValidatorLocksOpenRequisitionAndFreezesWorkflow(t *testing.T) {
	t.Parallel()
	now := time.Now()
	budgetID := "00000000-0000-4000-8000-000000000001"
	actorID := "00000000-0000-4000-8000-000000000002"
	requisitions := &postingRequisitionStub{requisition: &entity.Requisition{
		ID: "00000000-0000-4000-8000-000000000003", Status: "open", ApprovedAt: &now,
		ApprovedByPrincipalID: &actorID, OpenedAt: &now, BudgetReferenceID: &budgetID,
	}}
	applications := &postingApplicationStub{}
	v := &jobPostingValidator{requisitions: requisitions, applications: applications}
	posting := &entity.JobPosting{
		ID: "00000000-0000-4000-8000-000000000004", RequisitionID: requisitions.requisition.ID,
		ChannelCode: "career_site", PostingVersion: 1, Status: "draft",
		ContentDocumentVersionReferenceID: "00000000-0000-4000-8000-000000000005",
		ContentDigest:                     []byte("draft"), CreatedByPrincipalID: actorID,
	}
	if err := v.ValidateMutation(t.Context(), "hr_job_postings", nil, posting); err != nil {
		t.Fatalf("create draft posting: %v", err)
	}
	if requisitions.locks != 1 {
		t.Fatalf("parent lock calls = %d", requisitions.locks)
	}
	updated := *posting
	updated.ContentDigest = []byte("revised draft")
	if err := v.ValidateMutation(t.Context(), "hr_job_postings", posting, &updated); err != nil {
		t.Fatalf("revise draft posting: %v", err)
	}
	published := updated
	published.Status = "published"
	published.PublishedAt = &now
	if err := v.ValidateMutation(t.Context(), "hr_job_postings", posting, &published); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("publish through CRUD: %v", err)
	}
	changedParent := updated
	changedParent.RequisitionID = "00000000-0000-4000-8000-000000000006"
	if err := v.ValidateMutation(t.Context(), "hr_job_postings", posting, &changedParent); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("change requisition: %v", err)
	}
	applications.count = 1
	if err := v.ValidateMutation(t.Context(), "hr_job_postings", posting, nil); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("delete referenced posting: %v", err)
	}
	applications.count = 0
	if err := v.ValidateMutation(t.Context(), "hr_job_postings", posting, nil); err != nil {
		t.Fatalf("delete unreferenced draft: %v", err)
	}
	requisitions.requisition.Status = "closed"
	requisitions.requisition.ClosedAt = &now
	if err := v.ValidateMutation(t.Context(), "hr_job_postings", nil, posting); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("create after close: %v", err)
	}
	if err := v.ValidateMutation(t.Context(), "hr_job_postings", posting, nil); err != nil {
		t.Fatalf("delete closed requisition draft: %v", err)
	}
}
