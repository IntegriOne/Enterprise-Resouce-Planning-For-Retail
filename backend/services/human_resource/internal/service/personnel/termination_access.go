package service

import (
	"context"
	"strings"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
)

type TerminationAccessDecision struct {
	PrincipalID   string
	EffectiveDate time.Time
	MayDisable    bool
}

// VerifyTerminationAccess mencocokkan lifecycle terminasi dengan hubungan kerja terkini untuk integrasi privat IAM.
func (s *Service) VerifyTerminationAccess(ctx context.Context, workerID, planID string, effectiveDate *time.Time) (TerminationAccessDecision, error) {
	var decision TerminationAccessDecision
	if s == nil {
		return decision, hrerrors.ErrUnavailable
	}
	if err := entity.ValidateUUID("worker_id", workerID); err != nil {
		return decision, err
	}
	if err := entity.ValidateUUID("offboarding_plan_id", planID); err != nil {
		return decision, err
	}
	events, err := s.lifecycleEvents.SelectWhere(ctx, []domainrepo.Filter{
		{Column: "worker_id", Values: []any{workerID}},
		{Column: "idempotency_key", Values: []any{"worker-termination:" + strings.ToLower(planID)}},
	}, 2, 0)
	if err != nil {
		return decision, err
	}
	if len(events) != 1 {
		return decision, hrerrors.ErrReferenceConflict
	}
	event := events[0]
	if !strings.EqualFold(event.WorkerID, workerID) || event.EventType != "termination" ||
		event.SourceReference == nil || !strings.EqualFold(*event.SourceReference, planID) ||
		event.NewStatus == nil || *event.NewStatus != entity.WorkerStatusTerminated || event.EffectiveDate.IsZero() ||
		(effectiveDate != nil && entity.CompareDates(*effectiveDate, event.EffectiveDate) != 0) ||
		entity.CompareDates(event.EffectiveDate, s.transferDependencies.Runtime.Clock.Now().UTC()) > 0 {
		return decision, hrerrors.ErrReferenceConflict
	}
	decision.EffectiveDate = event.EffectiveDate
	repository := s.transferDependencies.Worker
	worker, err := repository.SelectByID(ctx, workerID)
	if err != nil {
		return TerminationAccessDecision{}, err
	}
	if worker == nil || !strings.EqualFold(worker.ID, workerID) {
		return TerminationAccessDecision{}, hrerrors.ErrReferenceConflict
	}
	if worker.IAMPrincipalReferenceID == nil {
		return decision, nil
	}
	decision.PrincipalID = *worker.IAMPrincipalReferenceID
	if entity.ValidateUUID("principal_id", decision.PrincipalID) != nil ||
		entity.ValidateUUID("person_party_reference_id", worker.PersonPartyReferenceID) != nil {
		return TerminationAccessDecision{}, hrerrors.ErrReferenceConflict
	}
	if worker.Status != entity.WorkerStatusTerminated || worker.TerminationDate == nil ||
		entity.CompareDates(*worker.TerminationDate, event.EffectiveDate) != 0 {
		return decision, nil
	}
	matchingWorker := false
	for offset := 0; ; offset += domainrepo.MaxLimit {
		workers, err := repository.SelectWhere(ctx, []domainrepo.Filter{
			{Column: "iam_principal_reference_id", Values: []any{decision.PrincipalID}},
		}, domainrepo.MaxLimit, offset)
		if err != nil {
			return TerminationAccessDecision{}, err
		}
		for _, linked := range workers {
			if linked.IAMPrincipalReferenceID == nil || !strings.EqualFold(*linked.IAMPrincipalReferenceID, decision.PrincipalID) ||
				!strings.EqualFold(linked.PersonPartyReferenceID, worker.PersonPartyReferenceID) {
				return TerminationAccessDecision{}, hrerrors.ErrReferenceConflict
			}
			if linked.Status != entity.WorkerStatusTerminated && linked.Status != entity.WorkerStatusDeceased {
				return decision, nil
			}
			if strings.EqualFold(linked.ID, workerID) {
				matchingWorker = linked.Status == entity.WorkerStatusTerminated && linked.TerminationDate != nil &&
					entity.CompareDates(*linked.TerminationDate, event.EffectiveDate) == 0
			}
		}
		if len(workers) < domainrepo.MaxLimit {
			break
		}
	}
	decision.MayDisable = matchingWorker
	return decision, nil
}

// VerifyWorkerAccountLink memeriksa bahwa akun provisioning menunjuk pada pekerja aktif dalam badan hukum yang sama.
func (s *Service) VerifyWorkerAccountLink(ctx context.Context, workerID, principalID, legalEntityID string) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	for _, attribute := range []struct{ name, id string }{
		{"worker_id", workerID}, {"principal_id", principalID}, {"legal_entity_id", legalEntityID},
	} {
		if err := entity.ValidateUUID(attribute.name, attribute.id); err != nil {
			return err
		}
	}
	worker, err := s.transferDependencies.Worker.SelectByID(ctx, workerID)
	if err != nil {
		return err
	}
	if worker == nil || worker.Status != entity.WorkerStatusActive || worker.TerminationDate != nil ||
		worker.IAMPrincipalReferenceID == nil || !strings.EqualFold(*worker.IAMPrincipalReferenceID, principalID) ||
		!strings.EqualFold(worker.ID, workerID) || !strings.EqualFold(worker.PrimaryLegalEntityReferenceID, legalEntityID) {
		return hrerrors.ErrReferenceConflict
	}
	return nil
}
