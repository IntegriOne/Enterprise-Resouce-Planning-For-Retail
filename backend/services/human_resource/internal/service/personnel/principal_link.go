package service

import (
	"context"
	"strings"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

type PrincipalLinkCheck struct {
	EvidenceID, WorkerID, PersonPartyID, PrincipalID string
	RequestedByPrincipalID                           string
}

// PrincipalLinkVerifier memeriksa bukti provisioning yang mengikat akun employee dan pekerja.
type PrincipalLinkVerifier interface {
	VerifyPrincipalLink(context.Context, PrincipalLinkCheck) error
}

// LinkWorkerPrincipal mencatat kaitan ke akun employee setelah bukti diverifikasi.
func (s *Service) LinkWorkerPrincipal(ctx context.Context, workerID, principalID, evidenceID string) (*entity.Worker, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	if err := entity.ValidateUUID("principal_id", principalID); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("evidence_id", evidenceID); err != nil {
		return nil, err
	}
	return s.worker.Command(ctx, workerID, "link_principal", func(txCtx context.Context, worker *entity.Worker, actor operation.Actor, now time.Time) (*entity.Worker, error) {
		if actor.Kind != "human" || strings.EqualFold(actor.PrincipalID, principalID) {
			return nil, hrerrors.ErrForbidden
		}
		if worker.Status != entity.WorkerStatusActive || worker.IAMPrincipalReferenceID != nil ||
			worker.TerminationDate != nil || now.Before(worker.UpdatedAt) {
			return nil, hrerrors.ErrWorkflowRequired
		}
		historyFilters := []domainrepo.Filter{
			{Column: "iam_principal_reference_id", Values: []any{principalID}},
			{Column: "status", Values: []any{entity.WorkerStatusTerminated, entity.WorkerStatusDeceased}},
		}
		workers := s.transferDependencies.Worker
		historyCount, err := workers.CountWhere(txCtx, historyFilters)
		if err != nil {
			return nil, err
		}
		if historyCount > 0 {
			samePersonCount, err := workers.CountWhere(txCtx, append(historyFilters,
				domainrepo.Filter{Column: "person_party_reference_id", Values: []any{worker.PersonPartyReferenceID}}))
			if err != nil {
				return nil, err
			}
			if samePersonCount != historyCount {
				return nil, hrerrors.ErrReferenceConflict
			}
		}
		verifier := s.transferDependencies.PrincipalLinkVerifier
		if err := operation.RequireDependency(verifier); err != nil {
			return nil, err
		}
		if err := verifier.VerifyPrincipalLink(txCtx, PrincipalLinkCheck{
			EvidenceID: evidenceID, WorkerID: worker.ID, PersonPartyID: worker.PersonPartyReferenceID,
			PrincipalID: principalID, RequestedByPrincipalID: actor.PrincipalID,
		}); err != nil {
			return nil, operation.OwnerVerificationError(txCtx, err)
		}
		event, err := entity.NewWorkerLifecycleEvent(entity.WorkerLifecycleEvent{
			WorkerID: worker.ID, EventType: "principal_linked", OccurredAt: now, EffectiveDate: now.UTC(),
			SourceReference: &evidenceID, ActorPrincipalID: &actor.PrincipalID,
			EventPayload: entity.JSON("{}"), IdempotencyKey: "worker-principal-link:" + worker.ID,
		})
		if err != nil {
			return nil, err
		}
		if err := s.lifecycleEvents.Insert(txCtx, event); err != nil {
			return nil, err
		}
		worker.IAMPrincipalReferenceID = &principalID
		worker.UpdatedAt = now
		return worker, nil
	})
}
