package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"strings"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

// OfferDocumentCheck membawa identitas dokumen dan kompensasi yang disetujui.
type OfferDocumentCheck struct {
	DocumentID, VersionID, OfferID, ApplicationID, CompensationVersionID, PrincipalID string
	Digest                                                                            []byte
}

// OfferDocumentProof membawa bukti versi dokumen yang dipublikasikan untuk penawaran terkait.
type OfferDocumentProof struct {
	DocumentID, VersionID, OfferID, ApplicationID, CompensationVersionID string
	OwnerManagement, Purpose                                             string
	Digest                                                               []byte
	Verified, Immutable                                                  bool
}

type OfferDocumentVerifier interface {
	VerifyOfferDocument(context.Context, OfferDocumentCheck) (OfferDocumentProof, error)
}

// IssueOffer menerbitkan penawaran yang disetujui setelah DRM memverifikasi dokumen final.
func (s *Service) IssueOffer(ctx context.Context, id, documentID, versionID string, digest []byte) (*entity.Offer, error) {
	if s == nil || operation.RequireDependency(s.applicationRecords) != nil ||
		operation.RequireDependency(s.stageHistoryWriter) != nil {
		return nil, hrerrors.ErrUnavailable
	}
	if err := entity.ValidateUUID("document_id", documentID); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("document_version_id", versionID); err != nil {
		return nil, err
	}
	if len(digest) != sha256.Size {
		return nil, hrerrors.Invalid("content_digest", "must contain a SHA-256 digest")
	}
	digest = bytes.Clone(digest)
	return s.offer.Command(ctx, id, "issue", func(txCtx context.Context, offer *entity.Offer, actor operation.Actor, now time.Time) (*entity.Offer, error) {
		if actor.Kind != "human" {
			return nil, hrerrors.ErrForbidden
		}
		if offer.OfferStatus != "approved" || offer.ApprovedByPrincipalID == nil || offer.ApprovedAt == nil ||
			offer.CreatedByPrincipalID == nil || offer.CreatedAt == nil ||
			strings.EqualFold(*offer.CreatedByPrincipalID, *offer.ApprovedByPrincipalID) ||
			offer.ApprovedAt.Before(*offer.CreatedAt) || now.Before(*offer.ApprovedAt) ||
			offer.ProposedCompensationDocumentVersionReferenceID == nil || offer.IssuedAt != nil ||
			offer.OfferDocumentReferenceID != nil || offer.OfferDocumentVersionReferenceID != nil || len(offer.OfferContentDigest) != 0 ||
			offer.CandidateResponse != nil || offer.RespondedAt != nil || offer.ExpiresAt == nil || !now.Before(*offer.ExpiresAt) ||
			(offer.ProposedStartDate != nil && entity.CompareDates(*offer.ProposedStartDate, now) < 0) {
			return nil, hrerrors.ErrWorkflowRequired
		}
		application, err := s.applicationRecords.FindByIDForUpdate(txCtx, offer.ApplicationID)
		if err != nil {
			return nil, err
		}
		if application == nil || now.Before(application.AppliedAt) {
			return nil, hrerrors.ErrReferenceConflict
		}
		if application.AssignedRecruiterPrincipalID == nil || !strings.EqualFold(*application.AssignedRecruiterPrincipalID, actor.PrincipalID) {
			return nil, hrerrors.ErrForbidden
		}
		switch application.Status {
		case entity.ApplicationStatusInterview, entity.ApplicationStatusAssessment, entity.ApplicationStatusOffer:
		default:
			return nil, hrerrors.ErrWorkflowRequired
		}
		activeOffers, err := s.offerRecords.CountWhere(txCtx, []domainrepo.Filter{
			{Column: "application_id", Values: []any{application.ID}},
			{Column: "offer_status", Values: []any{"issued", "accepted"}},
		})
		if err != nil {
			return nil, err
		}
		if activeOffers != 0 {
			return nil, hrerrors.ErrWorkflowRequired
		}
		if operation.RequireDependency(s.offerDocumentVerifier) != nil {
			return nil, hrerrors.ErrUnavailable
		}
		check := OfferDocumentCheck{
			DocumentID: documentID, VersionID: versionID, OfferID: offer.ID, ApplicationID: application.ID,
			CompensationVersionID: *offer.ProposedCompensationDocumentVersionReferenceID, PrincipalID: actor.PrincipalID,
			Digest: bytes.Clone(digest),
		}
		proof, err := s.offerDocumentVerifier.VerifyOfferDocument(txCtx, check)
		if err != nil {
			return nil, operation.OwnerVerificationError(txCtx, err)
		}
		if !proof.Verified || !proof.Immutable || proof.OwnerManagement != "human_resource" ||
			proof.Purpose != "recruitment_offer" || !bytes.Equal(proof.Digest, digest) ||
			!strings.EqualFold(proof.DocumentID, check.DocumentID) || !strings.EqualFold(proof.VersionID, check.VersionID) ||
			!strings.EqualFold(proof.OfferID, check.OfferID) || !strings.EqualFold(proof.ApplicationID, check.ApplicationID) ||
			!strings.EqualFold(proof.CompensationVersionID, check.CompensationVersionID) {
			return nil, hrerrors.ErrReferenceConflict
		}
		if application.Status != entity.ApplicationStatusOffer {
			next := *application
			stage := "offer"
			next.Status, next.CurrentStageCode = entity.ApplicationStatusOffer, &stage
			if err := next.Validate(); err != nil {
				return nil, err
			}
			if err := s.application.CheckIntegrity(txCtx, application, &next); err != nil {
				return nil, err
			}
			if err := s.applicationRecords.Update(txCtx, &next); err != nil {
				return nil, err
			}
			history, err := entity.NewApplicationStageHistory(entity.ApplicationStageHistory{
				ApplicationID: application.ID, PreviousStatus: &application.Status, NewStatus: next.Status,
				StageCode: &stage, ChangedAt: now, ChangedByPrincipalID: &actor.PrincipalID,
			})
			if err != nil {
				return nil, err
			}
			if err := s.stageHistoryWriter.Insert(txCtx, history); err != nil {
				return nil, err
			}
		}
		offer.OfferStatus, offer.IssuedAt = "issued", &now
		offer.OfferDocumentReferenceID, offer.OfferDocumentVersionReferenceID = &documentID, &versionID
		offer.OfferContentDigest = digest
		return offer, nil
	})
}
