package service_test

import (
	"bytes"
	"context"
	"errors"
	"testing"

	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/recruitment"
)

type postingDocumentVerifierStub struct {
	request service.PostingDocumentRequest
	proof   service.PostingDocumentProof
	err     error
	calls   int
}

func (v *postingDocumentVerifierStub) VerifyJobPostingDocument(_ context.Context, request service.PostingDocumentRequest) (service.PostingDocumentProof, error) {
	v.calls++
	v.request = request
	return v.proof, v.err
}

func TestPublishJobPostingRequiresImmutableOwnerProof(t *testing.T) {
	t.Parallel()
	versionID := "00000000-0000-4000-8000-000000000201"
	digest := bytes.Repeat([]byte{0x42}, 32)
	verifier := &postingDocumentVerifierStub{proof: service.PostingDocumentProof{
		VersionID: versionID, Digest: append([]byte(nil), digest...), Verified: true,
		Immutable: true, OwnerManagement: "human_resource", Purpose: "recruitment_job_posting",
	}}
	h, svc, _, postings := setupWithVerifiers(t, nil, verifier)
	created, err := svc.CreateJobPosting(t.Context(), service.CreateJobPostingInput{
		RequisitionID: "00000000-0000-4000-8000-000000000202", ChannelCode: "career_site",
		PostingVersion: 1, ContentDocumentVersionReferenceID: versionID, ContentDigest: digest,
	})
	if err != nil {
		t.Fatal(err)
	}
	verifier.err = errors.New("drm unavailable")
	if _, err := svc.PublishJobPosting(t.Context(), created.ID); !errors.Is(err, hrerrors.ErrUnavailable) {
		t.Fatalf("DRM outage: %v", err)
	}
	verifier.err = nil
	verifier.proof.Immutable = false
	if _, err := svc.PublishJobPosting(t.Context(), created.ID); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("mutable version: %v", err)
	}
	verifier.proof.Immutable = true
	verifier.proof.OwnerManagement = "commerce"
	if _, err := svc.PublishJobPosting(t.Context(), created.ID); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("wrong owner: %v", err)
	}
	verifier.proof.OwnerManagement = "human_resource"
	verifier.proof.Digest[0]++
	if _, err := svc.PublishJobPosting(t.Context(), created.ID); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("mismatched digest: %v", err)
	}
	if postings.Rows[created.ID].Status != "draft" || len(h.Events) != 1 {
		t.Fatal("rejected publication changed posting or outbox")
	}
	verifier.proof.Digest[0]--
	published, err := svc.PublishJobPosting(t.Context(), created.ID)
	if err != nil || published.Status != "published" || published.PublishedAt == nil || len(h.Events) != 2 {
		t.Fatalf("publication = %+v, error = %v", published, err)
	}
	if verifier.request.PostingID != created.ID || verifier.request.VersionID != versionID ||
		verifier.request.PrincipalID != h.Actor.PrincipalID || !bytes.Equal(verifier.request.Digest, digest) {
		t.Fatalf("DRM request = %+v", verifier.request)
	}
	if _, err := svc.PublishJobPosting(t.Context(), created.ID); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("repeat publication: %v", err)
	}
}

func TestPublishJobPostingWithoutVerifierFailsClosed(t *testing.T) {
	t.Parallel()
	_, svc, _, _ := setupWithVerifiers(t, nil, nil)
	if _, err := svc.PublishJobPosting(t.Context(), "00000000-0000-4000-8000-000000000203"); !errors.Is(err, hrerrors.ErrUnavailable) {
		t.Fatalf("missing DRM verifier: %v", err)
	}
}
