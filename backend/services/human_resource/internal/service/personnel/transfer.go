package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	personnelpolicy "github.com/integrione/erp-retail/backend/services/human_resource/domain/service/personnel"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

// SubmitWorkerAssignmentTransferInput membawa usulan transfer internal pada badan hukum yang sama.
type SubmitWorkerAssignmentTransferInput struct {
	WorkerID, SourceAssignmentID, TargetPositionID string
	EffectiveDate                                  time.Time
	ReasonCode, IdempotencyKey                     string
}

// SubmitWorkerAssignmentTransfer mencatat usulan tanpa menutup penempatan sumber.
func (s *Service) SubmitWorkerAssignmentTransfer(ctx context.Context, in SubmitWorkerAssignmentTransferInput) (*entity.WorkerAssignmentTransfer, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	deps := s.transferDependencies
	if operation.RequireDependency(deps.WorkerAssignmentTransfer) != nil ||
		operation.RequireDependency(deps.TransferPosition) != nil {
		return nil, hrerrors.ErrUnavailable
	}
	in.ReasonCode, in.IdempotencyKey = strings.TrimSpace(in.ReasonCode), strings.TrimSpace(in.IdempotencyKey)
	access := operation.Access{Domain: "personnel", Resource: "hr_worker_assignments", Action: "create"}
	actor, err := authorizeTransfer(ctx, deps.Runtime.Policy, access)
	if err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("worker_id", in.WorkerID); err != nil {
		return nil, err
	}
	var result *entity.WorkerAssignmentTransfer
	err = deps.Runtime.Transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		worker, err := deps.Worker.FindByIDForUpdate(txCtx, in.WorkerID)
		if err != nil {
			return err
		}
		if worker == nil {
			return hrerrors.ErrReferenceConflict
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
		existing, err := deps.WorkerAssignmentTransfer.FindByIdempotencyKey(txCtx, in.IdempotencyKey)
		if err == nil {
			if !matchesTransferSubmission(existing, in, actor.PrincipalID) {
				return hrerrors.ErrConflict
			}
			result = existing
			return verifyActor()
		}
		if !errors.Is(err, hrerrors.ErrNotFound) {
			return err
		}
		id, err := deps.Runtime.IDs.NewID()
		if err != nil {
			return err
		}
		request, err := entity.NewWorkerAssignmentTransfer(entity.WorkerAssignmentTransfer{
			ID: id, WorkerID: worker.ID, SourceAssignmentID: in.SourceAssignmentID,
			TargetPositionID: in.TargetPositionID, EffectiveDate: in.EffectiveDate,
			ReasonCode: in.ReasonCode, IdempotencyKey: in.IdempotencyKey, Status: "submitted",
			RequestedByPrincipalID: actor.PrincipalID, RequestedAt: deps.Runtime.Clock.Now(),
		})
		if err != nil {
			return err
		}
		source, err := deps.WorkerAssignment.SelectByID(txCtx, request.SourceAssignmentID)
		if err != nil {
			return err
		}
		if _, err := personnelpolicy.ClosePrimaryAssignment(worker, source, request.EffectiveDate); err != nil {
			return err
		}
		target, err := deps.TransferPosition.FindByIDForUpdate(txCtx, request.TargetPositionID)
		if err != nil {
			return err
		}
		if err := validateTransferTarget(worker, source, target, request.EffectiveDate); err != nil {
			return err
		}
		if err := deps.WorkerAssignmentTransfer.Insert(txCtx, request); err != nil {
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

func validateTransferTarget(worker *entity.Worker, source *entity.WorkerAssignment, target *entity.Position, effectiveDate time.Time) error {
	if target == nil || target.PositionStatus != "active" || strings.EqualFold(source.PositionID, target.ID) ||
		!strings.EqualFold(target.LegalEntityReferenceID, worker.PrimaryLegalEntityReferenceID) ||
		entity.CompareDates(effectiveDate, target.ValidFrom) < 0 ||
		(target.ValidTo != nil && entity.CompareDates(effectiveDate, *target.ValidTo) > 0) {
		return hrerrors.ErrReferenceConflict
	}
	return nil
}

func matchesTransferSubmission(existing *entity.WorkerAssignmentTransfer, in SubmitWorkerAssignmentTransferInput, principalID string) bool {
	return existing != nil && strings.EqualFold(existing.WorkerID, in.WorkerID) &&
		strings.EqualFold(existing.SourceAssignmentID, in.SourceAssignmentID) &&
		strings.EqualFold(existing.TargetPositionID, in.TargetPositionID) &&
		strings.EqualFold(existing.RequestedByPrincipalID, principalID) &&
		existing.ReasonCode == in.ReasonCode && entity.CompareDates(existing.EffectiveDate, in.EffectiveDate) == 0
}

// GetWorkerAssignmentTransfer membaca permintaan transfer dengan otorisasi atas pekerja terkait.
func (s *Service) GetWorkerAssignmentTransfer(ctx context.Context, id string) (*entity.WorkerAssignmentTransfer, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	deps := s.transferDependencies
	if operation.RequireDependency(deps.WorkerAssignmentTransfer) != nil ||
		operation.RequireDependency(deps.Worker) != nil || operation.RequireDependency(deps.Runtime.Policy) != nil {
		return nil, hrerrors.ErrUnavailable
	}
	access := operation.Access{Domain: "personnel", Resource: "hr_worker_assignment_transfers", Action: "read"}
	actor, err := authorizeTransfer(ctx, deps.Runtime.Policy, access)
	if err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("transfer_id", id); err != nil {
		return nil, err
	}
	transfer, err := deps.WorkerAssignmentTransfer.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	worker, err := deps.Worker.SelectByID(ctx, transfer.WorkerID)
	if err != nil {
		return nil, err
	}
	access.Attributes = map[string]string{"worker_id": worker.ID, "legal_entity_reference_id": worker.PrimaryLegalEntityReferenceID}
	checked, err := authorizeTransfer(ctx, deps.Runtime.Policy, access)
	if err != nil {
		return nil, err
	}
	if checked != actor {
		return nil, hrerrors.ErrForbidden
	}
	return transfer, nil
}

// ListWorkerAssignmentTransfers membaca riwayat transfer pekerja yang dapat diakses.
func (s *Service) ListWorkerAssignmentTransfers(ctx context.Context, workerID string, limit, offset int) (operation.Page[entity.WorkerAssignmentTransfer], error) {
	var page operation.Page[entity.WorkerAssignmentTransfer]
	if s == nil {
		return page, hrerrors.ErrUnavailable
	}
	deps := s.transferDependencies
	if operation.RequireDependency(deps.WorkerAssignmentTransfer) != nil ||
		operation.RequireDependency(deps.Worker) != nil || operation.RequireDependency(deps.Runtime.Policy) != nil {
		return page, hrerrors.ErrUnavailable
	}
	access := operation.Access{Domain: "personnel", Resource: "hr_worker_assignment_transfers", Action: "read"}
	actor, err := authorizeTransfer(ctx, deps.Runtime.Policy, access)
	if err != nil {
		return page, err
	}
	if err := entity.ValidateUUID("worker_id", workerID); err != nil {
		return page, err
	}
	if limit < 1 || limit > 100 || offset < 0 {
		return page, hrerrors.Invalid("pagination", "has invalid limits")
	}
	worker, err := deps.Worker.SelectByID(ctx, workerID)
	if err != nil {
		return page, err
	}
	access.Attributes = map[string]string{"worker_id": worker.ID, "legal_entity_reference_id": worker.PrimaryLegalEntityReferenceID}
	checked, err := authorizeTransfer(ctx, deps.Runtime.Policy, access)
	if err != nil {
		return page, err
	}
	if checked != actor {
		return page, hrerrors.ErrForbidden
	}
	items, total, err := deps.WorkerAssignmentTransfer.ListByWorkerID(ctx, worker.ID, limit, offset)
	if err != nil {
		return page, err
	}
	return operation.Page[entity.WorkerAssignmentTransfer]{Items: items, Total: total, Limit: limit, Offset: offset}, nil
}
