package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"strings"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

type OfferResponseCheck struct {
	EvidenceID, OfferID, CandidateID, VersionID, PrincipalID string
	Digest                                                   []byte
}

type OfferResponseProof struct {
	EvidenceID, OfferID, CandidateID, VersionID, Decision string
	Digest                                                []byte
	RespondedAt                                           time.Time
	CandidateVerified                                     bool
}

// OfferResponseVerifier mengambil keputusan yang dibuat kandidat melalui kanal terverifikasi.
type OfferResponseVerifier interface {
	VerifyOfferResponse(context.Context, OfferResponseCheck) (OfferResponseProof, error)
}

func (s *Service) AcceptOffer(ctx context.Context, id, evidenceID string) (*entity.Offer, error) {
	return s.recordOfferResponse(ctx, id, evidenceID, "accept", "accepted")
}

func (s *Service) DeclineOffer(ctx context.Context, id, evidenceID string) (*entity.Offer, error) {
	return s.recordOfferResponse(ctx, id, evidenceID, "decline", "declined")
}

func (s *Service) recordOfferResponse(ctx context.Context, id, evidenceID, action, decision string) (*entity.Offer, error) {
	if s == nil || operation.RequireDependency(s.applicationRecords) != nil {
		return nil, hrerrors.ErrUnavailable
	}
	if err := entity.ValidateUUID("response_evidence_id", evidenceID); err != nil {
		return nil, err
	}
	return s.offer.Command(ctx, id, action, func(txCtx context.Context, offer *entity.Offer, actor operation.Actor, now time.Time) (*entity.Offer, error) {
		if actor.Kind != "human" {
			return nil, hrerrors.ErrForbidden
		}
		if offer.OfferStatus != "issued" || offer.IssuedAt == nil || offer.ExpiresAt == nil ||
			offer.OfferDocumentVersionReferenceID == nil || len(offer.OfferContentDigest) != sha256.Size ||
			offer.CandidateResponse != nil || offer.RespondedAt != nil || offer.CandidateResponseEvidenceID != nil {
			return nil, hrerrors.ErrWorkflowRequired
		}
		application, err := s.applicationRecords.FindByIDForUpdate(txCtx, offer.ApplicationID)
		if err != nil {
			return nil, err
		}
		if application == nil || application.Status != entity.ApplicationStatusOffer {
			return nil, hrerrors.ErrWorkflowRequired
		}
		if application.AssignedRecruiterPrincipalID == nil || !strings.EqualFold(*application.AssignedRecruiterPrincipalID, actor.PrincipalID) {
			return nil, hrerrors.ErrForbidden
		}
		if operation.RequireDependency(s.offerResponseVerifier) != nil {
			return nil, hrerrors.ErrUnavailable
		}
		proof, err := s.offerResponseVerifier.VerifyOfferResponse(txCtx, OfferResponseCheck{
			EvidenceID: evidenceID, OfferID: offer.ID, CandidateID: application.CandidateID,
			VersionID: *offer.OfferDocumentVersionReferenceID, PrincipalID: actor.PrincipalID,
			Digest: bytes.Clone(offer.OfferContentDigest),
		})
		if err != nil {
			return nil, operation.OwnerVerificationError(txCtx, err)
		}
		if !proof.CandidateVerified || proof.Decision != decision || proof.RespondedAt.IsZero() ||
			proof.RespondedAt.Before(*offer.IssuedAt) || !proof.RespondedAt.Before(*offer.ExpiresAt) || proof.RespondedAt.After(now) ||
			!strings.EqualFold(proof.EvidenceID, evidenceID) || !strings.EqualFold(proof.OfferID, offer.ID) ||
			!strings.EqualFold(proof.CandidateID, application.CandidateID) ||
			!strings.EqualFold(proof.VersionID, *offer.OfferDocumentVersionReferenceID) || !bytes.Equal(proof.Digest, offer.OfferContentDigest) {
			return nil, hrerrors.ErrReferenceConflict
		}
		offer.OfferStatus, offer.CandidateResponse = proof.Decision, &proof.Decision
		offer.RespondedAt, offer.CandidateResponseEvidenceID = &proof.RespondedAt, &evidenceID
		return offer, nil
	})
}
