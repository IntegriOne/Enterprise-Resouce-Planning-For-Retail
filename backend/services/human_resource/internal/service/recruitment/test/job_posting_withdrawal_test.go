package service_test

import (
	"errors"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/recruitment"
)

func TestWithdrawJobPostingRequiresReasonAndPersistsActor(t *testing.T) {
	t.Parallel()
	h, svc, _, postings := setupWithVerifiers(t, nil, nil)
	id := "00000000-0000-4000-8000-000000000401"
	publishedAt := h.Now().Add(-time.Hour)
	postings.Rows[id] = entity.JobPosting{
		ID: id, RequisitionID: "00000000-0000-4000-8000-000000000402",
		ChannelCode: "career_site", PostingVersion: 1,
		ContentDocumentVersionReferenceID: "00000000-0000-4000-8000-000000000403",
		ContentDigest:                     make([]byte, 32), Status: "published", PublishedAt: &publishedAt,
		CreatedByPrincipalID: h.Actor.PrincipalID,
	}
	if _, err := svc.WithdrawJobPosting(t.Context(), id, "   "); !errors.Is(err, hrerrors.ErrInvalidInput) {
		t.Fatalf("blank reason = %v", err)
	}
	if _, err := svc.WithdrawJobPosting(t.Context(), id, "DOCUMENT_REVOKED"); err != nil {
		t.Fatalf("withdraw posting: %v", err)
	}
	withdrawn := postings.Rows[id]
	if withdrawn.Status != "withdrawn" || withdrawn.WithdrawnAt == nil || !withdrawn.WithdrawnAt.Equal(h.Now()) ||
		withdrawn.WithdrawalReasonCode == nil || *withdrawn.WithdrawalReasonCode != "DOCUMENT_REVOKED" ||
		withdrawn.WithdrawnByPrincipalID == nil || *withdrawn.WithdrawnByPrincipalID != h.Actor.PrincipalID || len(h.Events) != 1 {
		t.Fatalf("withdrawn posting = %+v, events = %d", withdrawn, len(h.Events))
	}
	if _, err := svc.WithdrawJobPosting(t.Context(), id, "DOCUMENT_REVOKED"); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("repeat withdrawal = %v", err)
	}
	if len(h.Events) != 1 {
		t.Fatalf("repeat withdrawal wrote %d events", len(h.Events))
	}
}

func TestWithdrawJobPostingRejectsDraftAndNilService(t *testing.T) {
	t.Parallel()
	h, svc, _, postings := setupWithVerifiers(t, nil, nil)
	id := "00000000-0000-4000-8000-000000000404"
	postings.Rows[id] = entity.JobPosting{
		ID: id, RequisitionID: "00000000-0000-4000-8000-000000000405",
		ChannelCode: "career_site", PostingVersion: 1,
		ContentDocumentVersionReferenceID: "00000000-0000-4000-8000-000000000406",
		ContentDigest:                     make([]byte, 32), Status: "draft", CreatedByPrincipalID: h.Actor.PrincipalID,
	}
	if _, err := svc.WithdrawJobPosting(t.Context(), id, "CANCELLED"); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("withdraw draft = %v", err)
	}
	var nilService *service.Service
	if _, err := nilService.WithdrawJobPosting(t.Context(), id, "CANCELLED"); !errors.Is(err, hrerrors.ErrUnavailable) {
		t.Fatalf("nil service = %v", err)
	}
}
