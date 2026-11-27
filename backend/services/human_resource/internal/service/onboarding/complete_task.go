package service

import (
	"context"
	"strings"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

// EvidenceVerifier meminta pemilik dokumen membuktikan versi bukti yang tepat dan immutable.
type EvidenceVerifier interface {
	VerifyOnboardingEvidence(context.Context, EvidenceRequest) (EvidenceResult, error)
}

// EvidenceRequest membawa referensi minimum untuk pembuktian versi dokumen kepada DRM.
type EvidenceRequest struct {
	VersionID   string
	TaskID      string
	PlanID      string
	PrincipalID string
}

// EvidenceResult adalah keputusan DRM untuk versi dokumen yang diminta.
type EvidenceResult struct {
	VersionID string
	TaskID    string
	PlanID    string
	Verified  bool
	Immutable bool
}

// CompleteOnboardingTask menyelesaikan tugas worker yang ditugaskan dengan bukti DRM terverifikasi.
func (s *Service) CompleteOnboardingTask(ctx context.Context, taskID, evidenceVersionID string) (*entity.OnboardingTask, error) {
	if s == nil || operation.RequireDependency(s.evidence) != nil || s.onboardingTask == nil {
		return nil, hrerrors.ErrUnavailable
	}
	if err := entity.ValidateUUID("evidence_document_version_reference_id", evidenceVersionID); err != nil {
		return nil, err
	}
	return s.onboardingTask.Command(ctx, taskID, "complete", func(txCtx context.Context, task *entity.OnboardingTask, actor operation.Actor, now time.Time) (*entity.OnboardingTask, error) {
		if actor.Kind != "human" || task.AssignedPrincipalID == nil ||
			!strings.EqualFold(strings.TrimSpace(*task.AssignedPrincipalID), actor.PrincipalID) {
			return nil, hrerrors.ErrForbidden
		}
		if task.CompletedAt != nil || task.CompletedByPrincipalID != nil || task.CompletionEvidenceReferenceID != nil {
			return nil, hrerrors.ErrWorkflowRequired
		}
		proof, err := s.evidence.VerifyOnboardingEvidence(txCtx, EvidenceRequest{
			VersionID: evidenceVersionID, TaskID: task.ID, PlanID: task.OnboardingPlanID, PrincipalID: actor.PrincipalID,
		})
		if err != nil {
			return nil, operation.OwnerVerificationError(txCtx, err)
		}
		if !proof.Verified || !proof.Immutable || !strings.EqualFold(proof.VersionID, evidenceVersionID) ||
			!strings.EqualFold(proof.TaskID, task.ID) || !strings.EqualFold(proof.PlanID, task.OnboardingPlanID) {
			return nil, hrerrors.ErrReferenceConflict
		}
		task.CompletedAt = &now
		task.CompletedByPrincipalID = &actor.PrincipalID
		task.CompletionEvidenceReferenceID = &evidenceVersionID
		return task, nil
	})
}
