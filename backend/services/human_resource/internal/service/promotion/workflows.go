package service

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	personnelpolicy "github.com/integrione/erp-retail/backend/services/human_resource/domain/service/personnel"
	promotionpolicy "github.com/integrione/erp-retail/backend/services/human_resource/domain/service/promotion"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

// DecidePromotionRecommendationInput membawa dokumen rasional dan idempotency keputusan promosi.
type DecidePromotionRecommendationInput struct {
	RecommendationID                    string
	RationaleDocumentVersionReferenceID string
	RationaleDigest                     []byte
	IdempotencyKey                      string
}

// SubmitPromotionRecommendation mengajukan rekomendasi promosi draf setelah asesmen talenta selesai.
func (s *Service) SubmitPromotionRecommendation(ctx context.Context, recommendationID string) (*entity.PromotionRecommendation, error) {
	if s == nil ||
		operation.RequireDependency(s.dependencies.Worker) != nil ||
		operation.RequireDependency(s.dependencies.WorkerAssignment) != nil ||
		operation.RequireDependency(s.dependencies.Position) != nil ||
		operation.RequireDependency(s.dependencies.TalentAssessment) != nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.promotionRecommendation.Command(ctx, recommendationID, "submit", func(txCtx context.Context, rec *entity.PromotionRecommendation, actor operation.Actor, _ time.Time) (*entity.PromotionRecommendation, error) {
		if actor.Kind != "human" {
			return nil, hrerrors.ErrForbidden
		}
		if rec.Status != "draft" {
			return nil, hrerrors.ErrWorkflowRequired
		}
		worker, err := s.dependencies.Worker.FindByIDForUpdate(txCtx, rec.WorkerID)
		if err != nil {
			return nil, referenceConflictOnMissing(err)
		}
		if worker == nil || worker.Status != entity.WorkerStatusActive {
			return nil, hrerrors.ErrReferenceConflict
		}
		source, err := s.dependencies.WorkerAssignment.FindByIDForUpdate(txCtx, rec.SourceAssignmentID)
		if err != nil {
			return nil, referenceConflictOnMissing(err)
		}
		if source == nil || source.WorkerID != rec.WorkerID || source.AssignmentType != "primary" ||
			source.PositionID == rec.TargetPositionID {
			return nil, hrerrors.ErrReferenceConflict
		}
		target, err := s.dependencies.Position.FindByIDForUpdate(txCtx, rec.TargetPositionID)
		if err != nil {
			return nil, referenceConflictOnMissing(err)
		}
		if target == nil || target.PositionStatus != "active" ||
			target.LegalEntityReferenceID != worker.PrimaryLegalEntityReferenceID {
			return nil, hrerrors.ErrReferenceConflict
		}
		assessment, err := s.dependencies.TalentAssessment.FindByIDForUpdate(txCtx, rec.AssessmentID)
		if err != nil {
			return nil, referenceConflictOnMissing(err)
		}
		if assessment == nil || assessment.WorkerID != rec.WorkerID {
			return nil, hrerrors.ErrReferenceConflict
		}
		if assessment.Status != "finalized" {
			return nil, hrerrors.ErrWorkflowRequired
		}
		rec.Status = "submitted"
		return rec, nil
	})
}

// ApprovePromotionRecommendation menyetujui rekomendasi promosi yang berstatus submitted secara independen.
func (s *Service) ApprovePromotionRecommendation(ctx context.Context, in DecidePromotionRecommendationInput) (*entity.PromotionRecommendation, error) {
	return s.decideRecommendation(ctx, in, "approve", entity.ApprovalDecisionApprove, "approved")
}

// RejectPromotionRecommendation menolak rekomendasi promosi yang berstatus submitted.
func (s *Service) RejectPromotionRecommendation(ctx context.Context, in DecidePromotionRecommendationInput) (*entity.PromotionRecommendation, error) {
	return s.decideRecommendation(ctx, in, "reject", entity.ApprovalDecisionReject, "rejected")
}

// ReturnPromotionRecommendation mengembalikan rekomendasi promosi yang berstatus submitted agar direvisi.
func (s *Service) ReturnPromotionRecommendation(ctx context.Context, in DecidePromotionRecommendationInput) (*entity.PromotionRecommendation, error) {
	return s.decideRecommendation(ctx, in, "return", entity.ApprovalDecisionReturn, "returned")
}

func (s *Service) decideRecommendation(
	ctx context.Context,
	in DecidePromotionRecommendationInput,
	action string,
	choice entity.ApprovalDecision,
	nextStatus string,
) (*entity.PromotionRecommendation, error) {
	if s == nil ||
		operation.RequireDependency(s.dependencies.Worker) != nil ||
		operation.RequireDependency(s.dependencies.PromotionDecision) != nil {
		return nil, hrerrors.ErrUnavailable
	}
	if err := entity.ValidateUUID("rationale_document_version_reference_id", in.RationaleDocumentVersionReferenceID); err != nil {
		return nil, hrerrors.ErrInvalidInput
	}
	if len(in.RationaleDigest) == 0 || strings.TrimSpace(in.IdempotencyKey) == "" {
		return nil, hrerrors.ErrInvalidInput
	}
	return s.promotionRecommendation.Command(ctx, in.RecommendationID, action, func(txCtx context.Context, rec *entity.PromotionRecommendation, actor operation.Actor, now time.Time) (*entity.PromotionRecommendation, error) {
		if actor.Kind != "human" {
			return nil, hrerrors.ErrForbidden
		}
		if rec.Status != "submitted" {
			return nil, hrerrors.ErrWorkflowRequired
		}
		worker, err := s.dependencies.Worker.FindByIDForUpdate(txCtx, rec.WorkerID)
		if err != nil {
			return nil, referenceConflictOnMissing(err)
		}
		if worker == nil || worker.Status != entity.WorkerStatusActive {
			return nil, hrerrors.ErrReferenceConflict
		}
		// Penyetuju tidak boleh sama dengan pengusul ataupun pekerja yang dipromosikan (SoD).
		if strings.EqualFold(actor.PrincipalID, rec.RecommendedByPrincipalID) ||
			(worker.IAMPrincipalReferenceID != nil && strings.EqualFold(actor.PrincipalID, *worker.IAMPrincipalReferenceID)) {
			return nil, hrerrors.ErrForbidden
		}
		prior, err := collectPromotionRecords(txCtx, s.dependencies.PromotionDecision, domainrepo.Filter{
			Column: "recommendation_id", Values: []any{rec.ID},
		})
		if err != nil {
			return nil, err
		}
		var maxStep int32
		for _, prev := range prior {
			if prev.StepNumber > maxStep {
				maxStep = prev.StepNumber
			}
		}
		decisionID, err := s.promotionDecision.NewID()
		if err != nil {
			return nil, err
		}
		decision := &entity.PromotionDecision{
			ID:                                  decisionID,
			RecommendationID:                    rec.ID,
			StepNumber:                          maxStep + 1,
			Decision:                            choice,
			DecidedByPrincipalID:                actor.PrincipalID,
			DecidedAt:                           now,
			RationaleDocumentVersionReferenceID: in.RationaleDocumentVersionReferenceID,
			RationaleDigest:                     append([]byte(nil), in.RationaleDigest...),
			IdempotencyKey:                      strings.TrimSpace(in.IdempotencyKey),
		}
		if err := promotionpolicy.ValidateDecision(rec, worker, decision, prior); err != nil {
			return nil, err
		}
		if err := s.promotionDecision.CheckIntegrity(txCtx, nil, decision); err != nil {
			return nil, err
		}
		// Trigger DB mensyaratkan recommendation masih submitted saat row decision dimasukkan.
		if err := s.dependencies.PromotionDecision.Insert(txCtx, decision); err != nil {
			return nil, err
		}
		rec.Status = nextStatus
		if nextStatus == "approved" {
			id := decision.ID
			rec.ApprovalDecisionID = &id
		}
		return rec, nil
	})
}

// ApplyPromotionRecommendation menerapkan promosi yang telah disetujui secara atomik ke penempatan dan riwayat pekerja.
func (s *Service) ApplyPromotionRecommendation(ctx context.Context, recommendationID string) (*entity.PromotionRecommendation, error) {
	if s == nil ||
		operation.RequireDependency(s.dependencies.Worker) != nil ||
		operation.RequireDependency(s.dependencies.WorkerAssignment) != nil ||
		operation.RequireDependency(s.dependencies.WorkerLifecycleEvent) != nil ||
		operation.RequireDependency(s.dependencies.Position) != nil ||
		operation.RequireDependency(s.dependencies.TalentAssessment) != nil ||
		operation.RequireDependency(s.dependencies.PromotionDecision) != nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.promotionRecommendation.Command(ctx, recommendationID, "apply", func(txCtx context.Context, rec *entity.PromotionRecommendation, actor operation.Actor, now time.Time) (*entity.PromotionRecommendation, error) {
		if actor.Kind != "human" {
			return nil, hrerrors.ErrForbidden
		}
		if rec.Status != "approved" || rec.ApprovalDecisionID == nil {
			return nil, hrerrors.ErrWorkflowRequired
		}
		if entity.CompareDates(rec.RecommendedEffectiveDate, now) > 0 {
			return nil, hrerrors.ErrWorkflowRequired
		}
		worker, err := s.dependencies.Worker.FindByIDForUpdate(txCtx, rec.WorkerID)
		if err != nil {
			return nil, referenceConflictOnMissing(err)
		}
		if worker == nil || now.Before(worker.UpdatedAt) {
			return nil, hrerrors.ErrReferenceConflict
		}
		if worker.IAMPrincipalReferenceID != nil && strings.EqualFold(*worker.IAMPrincipalReferenceID, actor.PrincipalID) {
			return nil, hrerrors.ErrForbidden
		}
		sourceAssignment, err := s.dependencies.WorkerAssignment.FindByIDForUpdate(txCtx, rec.SourceAssignmentID)
		if err != nil {
			return nil, referenceConflictOnMissing(err)
		}
		closedSource, err := personnelpolicy.ClosePrimaryAssignment(worker, sourceAssignment, rec.RecommendedEffectiveDate)
		if err != nil {
			return nil, err
		}
		targetPosition, err := s.dependencies.Position.FindByIDForUpdate(txCtx, rec.TargetPositionID)
		if err != nil {
			return nil, referenceConflictOnMissing(err)
		}
		assessment, err := s.dependencies.TalentAssessment.FindByIDForUpdate(txCtx, rec.AssessmentID)
		if err != nil {
			return nil, referenceConflictOnMissing(err)
		}
		decision, err := s.dependencies.PromotionDecision.FindByIDForUpdate(txCtx, *rec.ApprovalDecisionID)
		if err != nil {
			return nil, referenceConflictOnMissing(err)
		}
		if targetPosition == nil || assessment == nil || decision == nil {
			return nil, hrerrors.ErrReferenceConflict
		}
		if targetPosition.PositionStatus != "active" ||
			!strings.EqualFold(targetPosition.LegalEntityReferenceID, worker.PrimaryLegalEntityReferenceID) ||
			entity.CompareDates(targetPosition.ValidFrom, rec.RecommendedEffectiveDate) > 0 ||
			(targetPosition.ValidTo != nil && entity.CompareDates(*targetPosition.ValidTo, now) < 0) {
			return nil, hrerrors.ErrReferenceConflict
		}
		if sourceAssignment.ManagerWorkerID != nil {
			managers, err := s.dependencies.Worker.SelectWhere(txCtx, []domainrepo.Filter{
				{Column: "id", Values: []any{*sourceAssignment.ManagerWorkerID}},
			}, 1, 0)
			if err != nil {
				return nil, referenceConflictOnMissing(err)
			}
			if len(managers) != 1 {
				return nil, hrerrors.ErrReferenceConflict
			}
			if err := personnelpolicy.ValidateAssignmentManager(worker, &managers[0], *sourceAssignment.ManagerWorkerID,
				rec.RecommendedEffectiveDate, targetPosition.ValidTo); err != nil {
				return nil, err
			}
		}
		if err := ensurePromotionCapacity(txCtx, s.dependencies.WorkerAssignment, targetPosition, sourceAssignment, rec.RecommendedEffectiveDate); err != nil {
			return nil, err
		}

		if err := s.dependencies.WorkerAssignment.Update(txCtx, closedSource); err != nil {
			return nil, err
		}

		newAssignmentID, err := s.promotionRecommendation.NewID()
		if err != nil {
			return nil, err
		}
		activeSnapshot := string(entity.WorkerStatusActive)
		reason := "promotion"
		requestedAt := rec.RecommendedAt
		requestedBy := rec.RecommendedByPrincipalID
		approvedBy := decision.DecidedByPrincipalID
		approvedAt := decision.DecidedAt
		newAssignment := &entity.WorkerAssignment{
			ID:                       newAssignmentID,
			WorkerID:                 rec.WorkerID,
			PositionID:               rec.TargetPositionID,
			ManagerWorkerID:          sourceAssignment.ManagerWorkerID,
			AssignmentType:           "primary",
			EmploymentStatusSnapshot: &activeSnapshot,
			FTEPercent:               sourceAssignment.FTEPercent,
			ValidFrom:                rec.RecommendedEffectiveDate,
			ValidTo:                  targetPosition.ValidTo,
			AssignmentReason:         &reason,
			RequestedByPrincipalID:   &requestedBy,
			RequestedAt:              &requestedAt,
			ApprovedByPrincipalID:    &approvedBy,
			ApprovedAt:               &approvedAt,
		}
		if err := newAssignment.Validate(); err != nil {
			return nil, err
		}
		if err := s.dependencies.WorkerAssignment.Insert(txCtx, newAssignment); err != nil {
			return nil, err
		}

		nextWorker := *worker
		nextWorker.PrimaryAssignmentID = &newAssignment.ID
		nextWorker.UpdatedAt = now
		if err := nextWorker.Validate(); err != nil {
			return nil, err
		}
		if err := s.dependencies.Worker.Update(txCtx, &nextWorker); err != nil {
			return nil, err
		}

		sourceRef := rec.ID
		actorID := actor.PrincipalID
		prevStatus := worker.Status
		newStatus := worker.Status
		event := &entity.WorkerLifecycleEvent{
			WorkerID:         rec.WorkerID,
			EventType:        "promotion",
			OccurredAt:       now,
			EffectiveDate:    rec.RecommendedEffectiveDate,
			PreviousStatus:   &prevStatus,
			NewStatus:        &newStatus,
			SourceReference:  &sourceRef,
			ActorPrincipalID: &actorID,
			EventPayload: entity.JSON(fmt.Sprintf(
				`{"recommendation_id":%q,"source_assignment_id":%q,"applied_assignment_id":%q,"target_position_id":%q}`,
				rec.ID, sourceAssignment.ID, newAssignment.ID, targetPosition.ID,
			)),
			IdempotencyKey: "promotion-apply:" + rec.ID,
		}
		if err := event.Validate(); err != nil {
			return nil, err
		}
		if err := s.dependencies.WorkerLifecycleEvent.Insert(txCtx, event); err != nil {
			return nil, err
		}

		return promotionpolicy.BuildAppliedRecommendation(promotionpolicy.ApplicationInput{
			Recommendation:   rec,
			Decision:         decision,
			Assessment:       assessment,
			Worker:           &nextWorker,
			SourceAssignment: closedSource,
			NewAssignment:    newAssignment,
			TargetPosition:   targetPosition,
			Event:            event,
			BusinessDate:     now,
			AppliedAt:        now,
		})
	})
}

func ensurePromotionCapacity(
	ctx context.Context,
	assignments domainrepo.RecordRepository[entity.WorkerAssignment, string],
	position *entity.Position,
	candidate *entity.WorkerAssignment,
	effectiveDate time.Time,
) error {
	headcount, err := position.HeadcountAuthorized.Rat()
	if err != nil {
		return hrerrors.ErrReferenceConflict
	}
	existing, err := collectPromotionRecords(ctx, assignments, domainrepo.Filter{
		Column: "position_id", Values: []any{position.ID},
	})
	if err != nil {
		return err
	}
	workerLimit := big.NewRat(100, 1)
	positionLimit := new(big.Rat).Mul(headcount, workerLimit)
	if err := personnelpolicy.CheckAssignmentCapacity(candidate, existing, effectiveDate, position.ValidTo, positionLimit); err != nil {
		return err
	}
	existing, err = collectPromotionRecords(ctx, assignments, domainrepo.Filter{
		Column: "worker_id", Values: []any{candidate.WorkerID},
	})
	if err != nil {
		return err
	}
	return personnelpolicy.CheckAssignmentCapacity(candidate, existing, effectiveDate, position.ValidTo, workerLimit)
}

func collectPromotionRecords[T any](ctx context.Context, repo domainrepo.RecordRepository[T, string], filter domainrepo.Filter) ([]T, error) {
	filters := []domainrepo.Filter{filter}
	count, err := repo.CountWhere(ctx, filters)
	if err != nil {
		return nil, err
	}
	return operation.CollectComplete(ctx, count, func(limit, offset int) (operation.Page[T], error) {
		items, err := repo.SelectWhere(ctx, filters, limit, offset)
		return operation.Page[T]{Items: items}, err
	})
}

func referenceConflictOnMissing(err error) error {
	if errors.Is(err, hrerrors.ErrNotFound) {
		return hrerrors.ErrReferenceConflict
	}
	return err
}
