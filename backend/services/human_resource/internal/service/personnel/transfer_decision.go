package service

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"strings"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	personnelpolicy "github.com/integrione/erp-retail/backend/services/human_resource/domain/service/personnel"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

// DecideWorkerAssignmentTransfer mencatat keputusan independen dan menerapkan persetujuan secara atomik.
func (s *Service) DecideWorkerAssignmentTransfer(ctx context.Context, transferID string, approve bool) (*entity.WorkerAssignmentTransfer, error) {
	if s == nil || operation.RequireDependency(s.transferDependencies.WorkerAssignmentTransfer) != nil ||
		operation.RequireDependency(s.transferDependencies.TransferPosition) != nil {
		return nil, hrerrors.ErrUnavailable
	}
	deps := s.transferDependencies
	access := operation.Access{Domain: "personnel", Resource: "hr_worker_assignments", Action: "transfer"}
	actor, err := authorizeTransfer(ctx, deps.Runtime.Policy, access)
	if err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("transfer_id", transferID); err != nil {
		return nil, err
	}
	var result *entity.WorkerAssignmentTransfer
	err = deps.Runtime.Transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		request, err := deps.WorkerAssignmentTransfer.FindByIDForUpdate(txCtx, transferID)
		if err != nil {
			return err
		}
		worker, err := deps.Worker.FindByIDForUpdate(txCtx, request.WorkerID)
		if err != nil {
			return err
		}
		access.Attributes = map[string]string{"worker_id": worker.ID, "legal_entity_reference_id": worker.PrimaryLegalEntityReferenceID}
		verifyActor := func() error {
			checked, err := authorizeTransfer(txCtx, deps.Runtime.Policy, access)
			if err != nil {
				return err
			}
			if checked != actor {
				return hrerrors.ErrForbidden
			}
			return txCtx.Err()
		}
		if err := verifyActor(); err != nil {
			return err
		}
		if strings.EqualFold(request.RequestedByPrincipalID, actor.PrincipalID) ||
			(worker.IAMPrincipalReferenceID != nil && strings.EqualFold(*worker.IAMPrincipalReferenceID, actor.PrincipalID)) {
			return hrerrors.ErrForbidden
		}
		status := "rejected"
		if approve {
			status = "applied"
		}
		if request.Status != "submitted" {
			if request.Status != status || request.DecidedByPrincipalID == nil ||
				!strings.EqualFold(*request.DecidedByPrincipalID, actor.PrincipalID) {
				return hrerrors.ErrWorkflowRequired
			}
			result = request
			return verifyActor()
		}
		now := deps.Runtime.Clock.Now()
		if now.IsZero() || now.Before(request.RequestedAt) {
			return hrerrors.ErrReferenceConflict
		}
		if approve {
			assignmentID, err := s.applyWorkerAssignmentTransfer(txCtx, request, worker, actor, now)
			if err != nil {
				return err
			}
			request.AppliedAssignmentID = &assignmentID
		}
		request.Status, request.DecidedAt, request.DecidedByPrincipalID = status, &now, &actor.PrincipalID
		if err := request.Validate(); err != nil {
			return err
		}
		if err := deps.WorkerAssignmentTransfer.UpdateDecision(txCtx, request); err != nil {
			return err
		}
		if err := s.emitTransferDecision(txCtx, request, actor, now); err != nil {
			return err
		}
		result = request
		return verifyActor()
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *Service) applyWorkerAssignmentTransfer(ctx context.Context, request *entity.WorkerAssignmentTransfer, worker *entity.Worker, actor operation.Actor, now time.Time) (string, error) {
	deps := s.transferDependencies
	if entity.CompareDates(request.EffectiveDate, now) > 0 {
		return "", hrerrors.ErrWorkflowRequired
	}
	source, err := deps.WorkerAssignment.FindByIDForUpdate(ctx, request.SourceAssignmentID)
	if err != nil {
		return "", err
	}
	closed, err := personnelpolicy.ClosePrimaryAssignment(worker, source, request.EffectiveDate)
	if err != nil {
		return "", err
	}
	target, err := deps.TransferPosition.FindByIDForUpdate(ctx, request.TargetPositionID)
	if err != nil {
		return "", err
	}
	if err := validateTransferTarget(worker, source, target, request.EffectiveDate); err != nil {
		return "", err
	}
	if target.ValidTo != nil && entity.CompareDates(*target.ValidTo, now) < 0 {
		return "", hrerrors.ErrReferenceConflict
	}
	id, err := deps.Runtime.IDs.NewID()
	if err != nil {
		return "", err
	}
	draft := entity.WorkerAssignment{
		ID: id, WorkerID: worker.ID, PositionID: target.ID, ManagerWorkerID: source.ManagerWorkerID,
		AssignmentType: "primary", FTEPercent: source.FTEPercent,
		ValidFrom: request.EffectiveDate, ValidTo: target.ValidTo, AssignmentReason: &request.ReasonCode,
		RequestedAt: &request.RequestedAt, RequestedByPrincipalID: &request.RequestedByPrincipalID,
	}
	approved := draft
	workerStatus := string(worker.Status)
	approved.EmploymentStatusSnapshot, approved.ApprovedAt, approved.ApprovedByPrincipalID = &workerStatus, &now, &actor.PrincipalID
	if err := approved.Validate(); err != nil {
		return "", err
	}
	if err := deps.WorkerAssignment.Update(ctx, closed); err != nil {
		return "", err
	}
	if err := s.workerAssignment.CheckIntegrity(ctx, &draft, &approved); err != nil {
		return "", err
	}
	if err := deps.WorkerAssignment.Insert(ctx, &approved); err != nil {
		return "", err
	}
	next := *worker
	next.PrimaryAssignmentID, next.UpdatedAt = &id, now
	if err := next.Validate(); err != nil {
		return "", err
	}
	if err := s.worker.CheckIntegrity(ctx, worker, &next); err != nil {
		return "", err
	}
	if err := deps.Worker.Update(ctx, &next); err != nil {
		return "", err
	}
	payload, err := json.Marshal(map[string]string{"source_assignment_id": source.ID, "applied_assignment_id": id})
	if err != nil {
		return "", err
	}
	transfer := &entity.WorkerLifecycleEvent{
		WorkerID: worker.ID, EventType: "transfer", OccurredAt: now, EffectiveDate: request.EffectiveDate,
		PreviousStatus: &worker.Status, NewStatus: &worker.Status, SourceReference: &request.ID,
		ActorPrincipalID: &actor.PrincipalID, EventPayload: entity.JSON(payload), IdempotencyKey: "worker-assignment-transfer:" + request.ID,
	}
	if err := transfer.Validate(); err != nil {
		return "", err
	}
	if err := s.lifecycleEvents.Insert(ctx, transfer); err != nil {
		return "", err
	}
	return id, nil
}

func (s *Service) emitTransferDecision(ctx context.Context, request *entity.WorkerAssignmentTransfer, actor operation.Actor, now time.Time) error {
	deps := s.transferDependencies.Runtime
	id, err := deps.IDs.NewID()
	if err != nil {
		return err
	}
	payload, err := json.Marshal(map[string]string{
		"aggregate_id": request.ID, "change_kind": request.Status, "actor_principal_id": actor.PrincipalID,
	})
	if err != nil {
		return err
	}
	digest := sha256.Sum256(payload)
	event, err := entity.NewOutboxEvent(entity.OutboxEvent{
		ID: id, AggregateType: "hr_worker_assignment_transfers", AggregateID: request.ID, AggregateVersion: 1,
		EventType: "hr_worker_assignment_transfers." + request.Status, SchemaVersion: 1,
		Payload: entity.JSON(payload), PayloadDigest: digest[:], OccurredAt: now, AvailableAt: now,
	})
	if err != nil {
		return err
	}
	return deps.Outbox.Insert(ctx, event)
}
