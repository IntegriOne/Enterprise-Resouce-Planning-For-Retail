package service

import (
	"context"
	"strings"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

// OfferCompensationCheck mengikat permintaan verifikasi DRM pada penawaran dan ruang lingkupnya.
type OfferCompensationCheck struct {
	VersionID, OfferID, ApplicationID, JobID, LegalEntityID, CurrencyCode, PrincipalID string
}

// OfferCompensationProof membawa metadata otoritatif versi yang telah disetujui.
type OfferCompensationProof struct {
	VersionID, OfferID, ApplicationID, JobID, LegalEntityID, CurrencyCode string
	OwnerManagement, Purpose                                              string
	Verified, Immutable, Approved                                         bool
}

// OfferCompensationVerifier memeriksa versi usulan kompensasi yang dikelola DRM.
type OfferCompensationVerifier interface {
	VerifyOfferCompensation(context.Context, OfferCompensationCheck) (OfferCompensationProof, error)
}

// ApproveOffer mengesahkan draf melalui penyetuju lain setelah kompensasi diverifikasi.
func (s *Service) ApproveOffer(ctx context.Context, id, versionID string) (*entity.Offer, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	if err := entity.ValidateUUID("compensation_document_version_id", versionID); err != nil {
		return nil, err
	}
	return s.offer.Command(ctx, id, "approve", func(txCtx context.Context, offer *entity.Offer, actor operation.Actor, now time.Time) (*entity.Offer, error) {
		if actor.Kind != "human" {
			return nil, hrerrors.ErrForbidden
		}
		if offer.OfferStatus != "draft" || offer.CreatedAt == nil || offer.CreatedByPrincipalID == nil ||
			offer.ApprovedAt != nil || offer.ApprovedByPrincipalID != nil || offer.IssuedAt != nil || offer.CurrencyCode == nil ||
			now.Before(*offer.CreatedAt) {
			return nil, hrerrors.ErrWorkflowRequired
		}
		if strings.EqualFold(*offer.CreatedByPrincipalID, actor.PrincipalID) {
			return nil, hrerrors.ErrForbidden
		}
		if operation.RequireDependency(s.offerCompensationVerifier) != nil {
			return nil, hrerrors.ErrUnavailable
		}
		check := OfferCompensationCheck{
			VersionID: versionID, OfferID: offer.ID, ApplicationID: offer.ApplicationID, JobID: offer.ProposedJobID,
			LegalEntityID: offer.ProposedLegalEntityReferenceID, CurrencyCode: *offer.CurrencyCode, PrincipalID: actor.PrincipalID,
		}
		proof, err := s.offerCompensationVerifier.VerifyOfferCompensation(txCtx, check)
		if err != nil {
			return nil, operation.OwnerVerificationError(txCtx, err)
		}
		if !proof.Verified || !proof.Immutable || !proof.Approved || proof.OwnerManagement != "human_resource" ||
			proof.Purpose != "recruitment_offer_compensation" || proof.CurrencyCode != check.CurrencyCode ||
			!strings.EqualFold(proof.VersionID, check.VersionID) || !strings.EqualFold(proof.OfferID, check.OfferID) ||
			!strings.EqualFold(proof.ApplicationID, check.ApplicationID) || !strings.EqualFold(proof.JobID, check.JobID) ||
			!strings.EqualFold(proof.LegalEntityID, check.LegalEntityID) {
			return nil, hrerrors.ErrReferenceConflict
		}
		offer.OfferStatus, offer.ApprovedAt, offer.ApprovedByPrincipalID = "approved", &now, &actor.PrincipalID
		offer.ProposedCompensationDocumentVersionReferenceID = &versionID
		return offer, nil
	})
}
