package service_test

import (
	"errors"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/recruitment"
)

func TestCloseJobPostingRequiresPublishedStateAndWritesEvent(t *testing.T) {
	t.Parallel()
	h, svc, _, postings := setupWithVerifiers(t, nil, nil)
	id := "00000000-0000-4000-8000-000000000301"
	publishedAt := h.Now().Add(-time.Hour)
	postings.Rows[id] = entity.JobPosting{
		ID: id, RequisitionID: "00000000-0000-4000-8000-000000000302",
		ChannelCode: "career_site", PostingVersion: 1,
		ContentDocumentVersionReferenceID: "00000000-0000-4000-8000-000000000303",
		ContentDigest:                     make([]byte, 32), Status: "published", PublishedAt: &publishedAt,
		CreatedByPrincipalID: h.Actor.PrincipalID,
	}
	closed, err := svc.CloseJobPosting(t.Context(), id)
	if err != nil || closed.Status != "closed" || closed.ClosesAt == nil || !closed.ClosesAt.Equal(h.Now()) || len(h.Events) != 1 {
		t.Fatalf("close posting = %+v, events = %d, error = %v", closed, len(h.Events), err)
	}
	if _, err := svc.CloseJobPosting(t.Context(), id); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("repeat close = %v", err)
	}
	if len(h.Events) != 1 {
		t.Fatalf("repeat close wrote %d events", len(h.Events))
	}
}

func TestCloseJobPostingRejectsDraftAndInvalidTime(t *testing.T) {
	t.Parallel()
	h, svc, _, postings := setupWithVerifiers(t, nil, nil)
	id := "00000000-0000-4000-8000-000000000304"
	posting := entity.JobPosting{ID: id, RequisitionID: "00000000-0000-4000-8000-000000000305",
		ChannelCode: "career_site", PostingVersion: 1,
		ContentDocumentVersionReferenceID: "00000000-0000-4000-8000-000000000306",
		ContentDigest:                     make([]byte, 32), Status: "draft", CreatedByPrincipalID: h.Actor.PrincipalID}
	postings.Rows[id] = posting
	if _, err := svc.CloseJobPosting(t.Context(), id); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("close draft = %v", err)
	}
	posting.Status = "published"
	now := h.Now()
	posting.PublishedAt = &now
	postings.Rows[id] = posting
	if _, err := svc.CloseJobPosting(t.Context(), id); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("close at publication time = %v", err)
	}
	if len(h.Events) != 0 {
		t.Fatalf("rejected close wrote %d events", len(h.Events))
	}
	var nilService *service.Service
	if _, err := nilService.CloseJobPosting(t.Context(), id); !errors.Is(err, hrerrors.ErrUnavailable) {
		t.Fatalf("nil service = %v", err)
	}
}
