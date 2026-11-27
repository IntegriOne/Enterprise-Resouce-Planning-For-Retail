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

type ContractSignatureCheck struct {
	Contract                       entity.EmploymentContract
	DocumentVersionID, PrincipalID string
}

type ContractSignatureProof struct {
	DocumentID, DocumentVersionID string
	SignedAt                      time.Time
	ContentDigest                 []byte
}

// ContractSignatureVerifier membuktikan kesesuaian kontrak dan tanda tangan para pihak pada versi DRM.
type ContractSignatureVerifier interface {
	VerifyContractSignature(context.Context, ContractSignatureCheck) (ContractSignatureProof, error)
}

// RecordSignedEmploymentContract mencatat bukti tanda tangan yang telah diverifikasi DRM.
func (s *Service) RecordSignedEmploymentContract(ctx context.Context, contractID, versionID string) (*entity.EmploymentContract, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	if err := entity.ValidateUUID("document_version_id", versionID); err != nil {
		return nil, err
	}
	return s.employmentContract.Command(ctx, contractID, "record_signature", func(txCtx context.Context, contract *entity.EmploymentContract, actor operation.Actor, now time.Time) (*entity.EmploymentContract, error) {
		if actor.Kind != "human" {
			return nil, hrerrors.ErrForbidden
		}
		if contract.ContractStatus != "draft" || contract.SignedAt != nil || contract.SignedDocumentReferenceID != nil ||
			contract.SignedDocumentVersionReferenceID != nil || contract.SignedContentDigest != nil {
			return nil, hrerrors.ErrWorkflowRequired
		}
		verifier := s.transferDependencies.ContractSignatureVerifier
		if err := operation.RequireDependency(verifier); err != nil {
			return nil, err
		}
		proof, err := verifier.VerifyContractSignature(txCtx, ContractSignatureCheck{Contract: *contract, DocumentVersionID: versionID, PrincipalID: actor.PrincipalID})
		if err != nil {
			return nil, operation.OwnerVerificationError(txCtx, err)
		}
		if !strings.EqualFold(versionID, proof.DocumentVersionID) || entity.ValidateUUID("document_id", proof.DocumentID) != nil ||
			proof.SignedAt.IsZero() || proof.SignedAt.Before(contract.CreatedAt) || proof.SignedAt.After(now) ||
			len(proof.ContentDigest) != sha256.Size {
			return nil, hrerrors.ErrReferenceConflict
		}
		contract.ContractStatus = "signed"
		contract.SignedAt, contract.SignedDocumentReferenceID, contract.SignedDocumentVersionReferenceID = &proof.SignedAt, &proof.DocumentID, &versionID
		contract.SignedContentDigest = bytes.Clone(proof.ContentDigest)
		return contract, nil
	})
}
