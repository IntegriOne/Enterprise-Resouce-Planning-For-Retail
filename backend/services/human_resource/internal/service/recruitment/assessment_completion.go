package service

import (
	"bytes"
	"context"
	"strings"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

// AssessmentEvidenceRequest meminta DRM membuktikan versi bukti hasil asesmen kandidat.
type AssessmentEvidenceRequest struct {
	VersionID     string
	Digest        []byte
	AssessmentID  string
	ApplicationID string
	PrincipalID   string
}

// AssessmentEvidenceProof mengikat keputusan DRM pada versi, digest, pemilik, tujuan, dan objek asesmen.
type AssessmentEvidenceProof struct {
	VersionID       string
	Digest          []byte
	Verified        bool
	Immutable       bool
	OwnerManagement string
	Purpose         string
	AssessmentID    string
	ApplicationID   string
}

// AssessmentEvidenceVerifier adalah port pembuktian bukti asesmen kandidat milik DRM.
type AssessmentEvidenceVerifier interface {
	VerifyAssessmentEvidence(context.Context, AssessmentEvidenceRequest) (AssessmentEvidenceProof, error)
}

// CompleteApplicationAssessmentInput membawa hasil penyelesaian asesmen lamaran kandidat.
type CompleteApplicationAssessmentInput struct {
	NormalizedScore                    *entity.Decimal
	RecommendationCode                 *string
	EvidenceDocumentVersionReferenceID *string
	EvidenceDigest                     []byte
}

// CompleteApplicationAssessment menyelesaikan asesmen terencana serta mencatat assessor dan waktu server.
func (s *Service) CompleteApplicationAssessment(ctx context.Context, assessmentID string, in CompleteApplicationAssessmentInput) (*entity.ApplicationAssessment, error) {
	if s == nil || s.applicationAssessment == nil {
		return nil, hrerrors.ErrUnavailable
	}
	var recommendationCode *string
	if in.RecommendationCode != nil {
		trimmed := strings.TrimSpace(*in.RecommendationCode)
		if trimmed == "" || len(trimmed) > 40 {
			return nil, hrerrors.Invalid("recommendation_code", "has invalid value")
		}
		recommendationCode = &trimmed
	}
	var evidenceVersionID *string
	if in.EvidenceDocumentVersionReferenceID != nil {
		trimmed := strings.TrimSpace(*in.EvidenceDocumentVersionReferenceID)
		if err := entity.ValidateUUID("evidence_document_version_reference_id", trimmed); err != nil {
			return nil, err
		}
		evidenceVersionID = &trimmed
	}
	hasEvidenceRef := evidenceVersionID != nil
	hasEvidenceDigest := len(in.EvidenceDigest) > 0
	if hasEvidenceRef != hasEvidenceDigest {
		return nil, hrerrors.Invalid("evidence_digest", "evidence document version and digest must be provided together")
	}
	if hasEvidenceDigest && len(in.EvidenceDigest) != 32 {
		return nil, hrerrors.Invalid("evidence_digest", "must be a 32-byte digest")
	}
	if hasEvidenceRef && operation.RequireDependency(s.assessmentEvidence) != nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.applicationAssessment.Command(ctx, assessmentID, "complete", func(txCtx context.Context, assessment *entity.ApplicationAssessment, actor operation.Actor, now time.Time) (*entity.ApplicationAssessment, error) {
		if actor.Kind != "human" {
			return nil, hrerrors.ErrForbidden
		}
		if (assessment.Status != "planned" && assessment.Status != "in_progress") || assessment.CompletedAt != nil {
			return nil, hrerrors.ErrWorkflowRequired
		}
		if assessment.AssessorPrincipalID != nil &&
			!strings.EqualFold(strings.TrimSpace(*assessment.AssessorPrincipalID), actor.PrincipalID) {
			return nil, hrerrors.ErrForbidden
		}
		if assessment.ScheduledAt != nil && now.Before(*assessment.ScheduledAt) {
			return nil, hrerrors.ErrWorkflowRequired
		}
		if hasEvidenceRef {
			proof, err := s.assessmentEvidence.VerifyAssessmentEvidence(txCtx, AssessmentEvidenceRequest{
				VersionID:     *evidenceVersionID,
				Digest:        append([]byte(nil), in.EvidenceDigest...),
				AssessmentID:  assessment.ID,
				ApplicationID: assessment.ApplicationID,
				PrincipalID:   actor.PrincipalID,
			})
			if err != nil {
				return nil, operation.OwnerVerificationError(txCtx, err)
			}
			if !proof.Verified || !proof.Immutable || proof.VersionID != *evidenceVersionID ||
				!bytes.Equal(proof.Digest, in.EvidenceDigest) ||
				!strings.EqualFold(proof.AssessmentID, assessment.ID) || !strings.EqualFold(proof.ApplicationID, assessment.ApplicationID) ||
				proof.OwnerManagement != "human_resource" || proof.Purpose != "recruitment_assessment_evidence" {
				return nil, hrerrors.ErrReferenceConflict
			}
			assessment.EvidenceDocumentVersionReferenceID = evidenceVersionID
			assessment.EvidenceDigest = append([]byte(nil), in.EvidenceDigest...)
		}
		assessment.Status = "completed"
		assessment.AssessorPrincipalID = &actor.PrincipalID
		assessment.CompletedAt = &now
		assessment.NormalizedScore = in.NormalizedScore
		assessment.RecommendationCode = recommendationCode
		return assessment, nil
	})
}
