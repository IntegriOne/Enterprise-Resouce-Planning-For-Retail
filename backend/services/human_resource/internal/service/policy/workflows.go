package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

// AcknowledgePolicy mencatat pengakuan kebijakan oleh pekerja yang bersangkutan dengan identitas IAM terverifikasi.
func (s *Service) AcknowledgePolicy(
	ctx context.Context,
	acknowledgementID string,
	evidenceDocVersionRefID *string,
) (*entity.PolicyAcknowledgement, error) {
	if s == nil || operation.RequireDependency(s.dependencies.Worker) != nil {
		return nil, hrerrors.ErrUnavailable
	}
	if evidenceDocVersionRefID != nil {
		if err := entity.ValidateUUID("evidence_document_version_reference_id", *evidenceDocVersionRefID); err != nil {
			return nil, hrerrors.ErrInvalidInput
		}
	}
	return s.policyAcknowledgement.Command(ctx, acknowledgementID, "acknowledge", func(txCtx context.Context, ack *entity.PolicyAcknowledgement, actor operation.Actor, now time.Time) (*entity.PolicyAcknowledgement, error) {
		if actor.Kind != "human" {
			return nil, hrerrors.ErrForbidden
		}
		if ack.AcknowledgedAt != nil || ack.AcknowledgedByPrincipalID != nil {
			return nil, hrerrors.ErrWorkflowRequired
		}
		worker, err := s.dependencies.Worker.FindByIDForUpdate(txCtx, ack.WorkerID)
		if err != nil {
			if errors.Is(err, hrerrors.ErrNotFound) {
				return nil, hrerrors.ErrReferenceConflict
			}
			return nil, err
		}
		if worker == nil || worker.Status != entity.WorkerStatusActive || worker.IAMPrincipalReferenceID == nil {
			return nil, hrerrors.ErrReferenceConflict
		}
		if !strings.EqualFold(*worker.IAMPrincipalReferenceID, actor.PrincipalID) {
			return nil, hrerrors.ErrForbidden
		}
		if now.Before(ack.AssignedAt) {
			return nil, hrerrors.ErrReferenceConflict
		}
		acknowledgedBy := actor.PrincipalID
		acknowledgedAt := now
		ack.AcknowledgedByPrincipalID = &acknowledgedBy
		ack.AcknowledgedAt = &acknowledgedAt
		if evidenceDocVersionRefID != nil {
			evidence := *evidenceDocVersionRefID
			ack.EvidenceDocumentVersionReferenceID = &evidence
		}
		return ack, nil
	})
}
