package service

import (
	"bytes"
	"context"
	"errors"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

// PostingDocumentRequest meminta DRM membuktikan versi dan digest konten posting.
type PostingDocumentRequest struct {
	VersionID     string
	Digest        []byte
	PostingID     string
	RequisitionID string
	PrincipalID   string
}

// PostingDocumentProof mengikat keputusan DRM pada versi, pemilik, dan tujuan dokumen.
type PostingDocumentProof struct {
	VersionID       string
	Digest          []byte
	Verified        bool
	Immutable       bool
	OwnerManagement string
	Purpose         string
}

// PostingDocumentVerifier adalah port pembuktian versi konten milik DRM.
type PostingDocumentVerifier interface {
	VerifyJobPostingDocument(context.Context, PostingDocumentRequest) (PostingDocumentProof, error)
}

// PublishJobPosting melepas draf untuk distribusi setelah bukti konten diverifikasi.
func (s *Service) PublishJobPosting(ctx context.Context, postingID string) (*entity.JobPosting, error) {
	if s == nil || operation.RequireDependency(s.postingVerifier) != nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.jobPosting.Command(ctx, postingID, "publish", func(txCtx context.Context, posting *entity.JobPosting, actor operation.Actor, now time.Time) (*entity.JobPosting, error) {
		if actor.Kind != "human" {
			return nil, hrerrors.ErrForbidden
		}
		if posting.Status != "draft" || posting.PublishedAt != nil || posting.ClosesAt != nil ||
			posting.ExternalPostingReference != nil || len(posting.ContentDigest) != 32 {
			return nil, hrerrors.ErrWorkflowRequired
		}
		proof, err := s.postingVerifier.VerifyJobPostingDocument(txCtx, PostingDocumentRequest{
			VersionID: posting.ContentDocumentVersionReferenceID, Digest: append([]byte(nil), posting.ContentDigest...),
			PostingID: posting.ID, RequisitionID: posting.RequisitionID, PrincipalID: actor.PrincipalID,
		})
		if err != nil {
			if txCtx.Err() != nil {
				return nil, txCtx.Err()
			}
			if errors.Is(err, hrerrors.ErrReferenceConflict) {
				return nil, hrerrors.ErrReferenceConflict
			}
			return nil, hrerrors.ErrUnavailable
		}
		if !proof.Verified || !proof.Immutable || proof.VersionID != posting.ContentDocumentVersionReferenceID ||
			!bytes.Equal(proof.Digest, posting.ContentDigest) || proof.OwnerManagement != "human_resource" ||
			proof.Purpose != "recruitment_job_posting" {
			return nil, hrerrors.ErrReferenceConflict
		}
		posting.Status = "published"
		posting.PublishedAt = &now
		return posting, nil
	})
}
