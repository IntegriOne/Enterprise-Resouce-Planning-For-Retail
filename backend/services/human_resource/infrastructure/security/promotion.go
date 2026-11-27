package security

import (
	"context"
	"reflect"
	"strings"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	personnelpolicy "github.com/integrione/erp-retail/backend/services/human_resource/domain/service/personnel"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository"
	"github.com/integrione/erp-retail/backend/shared/authcontext"
)

type promotionAssignmentReader interface {
	FindByIDForUpdate(context.Context, string) (*entity.WorkerAssignment, error)
}

type promotionPositionReader interface {
	FindByIDForUpdate(context.Context, string) (*entity.Position, error)
}

type promotionAssessmentReader interface {
	FindByIDForUpdate(context.Context, string) (*entity.TalentAssessment, error)
}

type promotionRecommendationReader interface {
	FindByIDForUpdate(context.Context, string) (*entity.PromotionRecommendation, error)
	SelectWhere(context.Context, []domainrepo.Filter, int, int) ([]entity.PromotionRecommendation, error)
	CountWhere(context.Context, []domainrepo.Filter) (int64, error)
}

type promotionDecisionReader interface {
	SelectWhere(context.Context, []domainrepo.Filter, int, int) ([]entity.PromotionDecision, error)
	CountWhere(context.Context, []domainrepo.Filter) (int64, error)
}

func loadPromotionDecisions(ctx context.Context, reader promotionDecisionReader, recommendationID string) ([]entity.PromotionDecision, error) {
	filters := []domainrepo.Filter{{Column: "recommendation_id", Values: []any{recommendationID}}}
	count, err := reader.CountWhere(ctx, filters)
	if err != nil {
		return nil, err
	}
	return operation.CollectComplete(ctx, count, func(limit, offset int) (operation.Page[entity.PromotionDecision], error) {
		rows, err := reader.SelectWhere(ctx, filters, limit, offset)
		return operation.Page[entity.PromotionDecision]{Items: rows}, err
	})
}

type promotionLifecycleEventReader interface {
	FindByIDForUpdate(context.Context, int64) (*entity.WorkerLifecycleEvent, error)
}

type promotionRecommendationValidator struct {
	workers         lockedWorkerReader
	assignments     promotionAssignmentReader
	positions       promotionPositionReader
	assessments     promotionAssessmentReader
	recommendations promotionRecommendationReader
	decisions       promotionDecisionReader
	events          promotionLifecycleEventReader
	positionOwner   PositionReferenceVerifier
}

type promotionDecisionValidator struct {
	recommendations promotionRecommendationReader
	workers         lockedWorkerReader
	decisions       promotionDecisionReader
}

// PromotionValidators mendaftarkan validator integritas rekomendasi dan keputusan promosi.
func PromotionValidators(repos *repository.Registry, owner PositionReferenceVerifier) (map[string]MutationValidator, error) {
	if repos == nil ||
		operation.RequireDependency(repos.Worker) != nil ||
		operation.RequireDependency(repos.WorkerAssignment) != nil ||
		operation.RequireDependency(repos.Position) != nil ||
		operation.RequireDependency(repos.TalentAssessment) != nil ||
		operation.RequireDependency(repos.PromotionRecommendation) != nil ||
		operation.RequireDependency(repos.PromotionDecision) != nil ||
		operation.RequireDependency(repos.WorkerLifecycleEvent) != nil {
		return nil, hrerrors.ErrUnavailable
	}
	return map[string]MutationValidator{
		"hr_promotion_recommendations": &promotionRecommendationValidator{
			workers:         repos.Worker,
			assignments:     repos.WorkerAssignment,
			positions:       repos.Position,
			assessments:     repos.TalentAssessment,
			recommendations: repos.PromotionRecommendation,
			decisions:       repos.PromotionDecision,
			events:          repos.WorkerLifecycleEvent,
			positionOwner:   owner,
		},
		"hr_promotion_decisions": &promotionDecisionValidator{
			recommendations: repos.PromotionRecommendation,
			workers:         repos.Worker,
			decisions:       repos.PromotionDecision,
		},
	}, nil
}

// ValidateMutation menjaga draf, alur pengajuan/persetujuan, dan bukti penerapan promosi tetap konsisten.
func (v *promotionRecommendationValidator) ValidateMutation(ctx context.Context, table string, before, after entity.Validatable) error {
	if v == nil || table != "hr_promotion_recommendations" ||
		operation.RequireDependency(v.workers) != nil ||
		operation.RequireDependency(v.assignments) != nil ||
		operation.RequireDependency(v.positions) != nil ||
		operation.RequireDependency(v.assessments) != nil ||
		operation.RequireDependency(v.recommendations) != nil ||
		operation.RequireDependency(v.decisions) != nil ||
		operation.RequireDependency(v.events) != nil {
		return hrerrors.ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	var previous *entity.PromotionRecommendation
	if before != nil {
		var ok bool
		previous, ok = before.(*entity.PromotionRecommendation)
		if !ok || previous == nil {
			return hrerrors.ErrInvalidInput
		}
	}
	if after == nil {
		if previous == nil {
			return hrerrors.ErrInvalidInput
		}
		if previous.Status != "draft" {
			return hrerrors.ErrWorkflowRequired
		}
		return v.rejectIfReferenced(ctx, previous.ID)
	}
	rec, ok := after.(*entity.PromotionRecommendation)
	if !ok || rec == nil {
		return hrerrors.ErrInvalidInput
	}
	if previous != nil && previous.ID != rec.ID {
		return hrerrors.ErrInvalidInput
	}
	if previous == nil {
		caller, ok := authcontext.FromContext(ctx)
		if !ok || !caller.Authenticated {
			return hrerrors.ErrUnauthenticated
		}
		if !strings.EqualFold(caller.PrincipalID, rec.RecommendedByPrincipalID) {
			return hrerrors.ErrForbidden
		}
		if rec.Status != "draft" || rec.ApprovalDecisionID != nil ||
			rec.AppliedAssignmentID != nil || rec.AppliedLifecycleEventID != nil || rec.AppliedAt != nil {
			return hrerrors.ErrWorkflowRequired
		}
		return v.validateDraftReferencesAndUniqueness(ctx, rec)
	}
	switch {
	case previous.Status == "draft" && rec.Status == "draft":
		if rec.ApprovalDecisionID != nil || rec.AppliedAssignmentID != nil ||
			rec.AppliedLifecycleEventID != nil || rec.AppliedAt != nil {
			return hrerrors.ErrWorkflowRequired
		}
		if previous.RecommendedByPrincipalID != rec.RecommendedByPrincipalID ||
			!previous.RecommendedAt.Equal(rec.RecommendedAt) {
			return hrerrors.ErrWorkflowRequired
		}
		return v.validateDraftReferencesAndUniqueness(ctx, rec)
	case previous.Status == "draft" && rec.Status == "submitted":
		unchanged := *rec
		unchanged.Status = "draft"
		if !reflect.DeepEqual(previous, &unchanged) {
			return hrerrors.ErrWorkflowRequired
		}
		return v.validateSubmittedParents(ctx, rec)
	case previous.Status == "submitted" && (rec.Status == "approved" || rec.Status == "rejected" || rec.Status == "returned"):
		return v.validateDecisionTransition(ctx, previous, rec)
	case previous.Status == "approved" && rec.Status == "applied":
		return v.validateApplicationTransition(ctx, previous, rec)
	default:
		return hrerrors.ErrWorkflowRequired
	}
}

func (v *promotionRecommendationValidator) validateDraftReferencesAndUniqueness(ctx context.Context, rec *entity.PromotionRecommendation) error {
	if rec.SupersedesRecommendationID != nil {
		if *rec.SupersedesRecommendationID == rec.ID {
			return hrerrors.ErrInvalidInput
		}
		prior, err := v.recommendations.FindByIDForUpdate(ctx, *rec.SupersedesRecommendationID)
		if err != nil {
			return promotionReferenceError(err)
		}
		if prior == nil || prior.WorkerID != rec.WorkerID {
			return hrerrors.ErrReferenceConflict
		}
	}
	if err := v.lockBaseParents(ctx, rec, false); err != nil {
		return err
	}
	byNumber, err := v.recommendations.SelectWhere(ctx, []domainrepo.Filter{
		{Column: "recommendation_number", Values: []any{rec.RecommendationNumber}},
	}, 2, 0)
	if err != nil {
		return err
	}
	for _, existing := range byNumber {
		if existing.ID != rec.ID {
			return hrerrors.ErrAlreadyExists
		}
	}
	byIdem, err := v.recommendations.SelectWhere(ctx, []domainrepo.Filter{
		{Column: "idempotency_key", Values: []any{rec.IdempotencyKey}},
	}, 2, 0)
	if err != nil {
		return err
	}
	for _, existing := range byIdem {
		if existing.ID != rec.ID {
			return hrerrors.ErrAlreadyExists
		}
	}
	return nil
}

func (v *promotionRecommendationValidator) validateSubmittedParents(ctx context.Context, rec *entity.PromotionRecommendation) error {
	return v.lockBaseParents(ctx, rec, true)
}

func (v *promotionRecommendationValidator) lockBaseParents(ctx context.Context, rec *entity.PromotionRecommendation, requireFinalizedAssessment bool) error {
	worker, err := v.workers.FindByIDForUpdate(ctx, rec.WorkerID)
	if err != nil {
		return promotionReferenceError(err)
	}
	if worker == nil || worker.Status != entity.WorkerStatusActive {
		return hrerrors.ErrReferenceConflict
	}
	source, err := v.assignments.FindByIDForUpdate(ctx, rec.SourceAssignmentID)
	if err != nil {
		return promotionReferenceError(err)
	}
	if source == nil || source.WorkerID != rec.WorkerID || source.AssignmentType != "primary" ||
		source.PositionID == rec.TargetPositionID {
		return hrerrors.ErrReferenceConflict
	}
	if requireFinalizedAssessment {
		if _, err := personnelpolicy.ClosePrimaryAssignment(worker, source, rec.RecommendedEffectiveDate); err != nil {
			return err
		}
	}
	target, err := v.positions.FindByIDForUpdate(ctx, rec.TargetPositionID)
	if err != nil {
		return promotionReferenceError(err)
	}
	if target == nil || target.PositionStatus != "active" ||
		target.LegalEntityReferenceID != worker.PrimaryLegalEntityReferenceID ||
		entity.CompareDates(rec.RecommendedEffectiveDate, target.ValidFrom) < 0 ||
		(target.ValidTo != nil && entity.CompareDates(rec.RecommendedEffectiveDate, *target.ValidTo) > 0) {
		return hrerrors.ErrReferenceConflict
	}
	assessment, err := v.assessments.FindByIDForUpdate(ctx, rec.AssessmentID)
	if err != nil {
		return promotionReferenceError(err)
	}
	if assessment == nil || assessment.WorkerID != rec.WorkerID {
		return hrerrors.ErrReferenceConflict
	}
	if requireFinalizedAssessment && assessment.Status != "finalized" {
		return hrerrors.ErrWorkflowRequired
	}
	return nil
}

func (v *promotionRecommendationValidator) validateDecisionTransition(ctx context.Context, previous, rec *entity.PromotionRecommendation) error {
	if rec.AppliedAssignmentID != nil || rec.AppliedLifecycleEventID != nil || rec.AppliedAt != nil {
		return hrerrors.ErrWorkflowRequired
	}
	unchanged := *rec
	unchanged.Status = previous.Status
	unchanged.ApprovalDecisionID = previous.ApprovalDecisionID
	if !reflect.DeepEqual(previous, &unchanged) {
		return hrerrors.ErrWorkflowRequired
	}
	if rec.Status == "approved" && rec.ApprovalDecisionID == nil {
		return hrerrors.ErrWorkflowRequired
	}
	if rec.Status != "approved" && rec.ApprovalDecisionID != nil {
		return hrerrors.ErrWorkflowRequired
	}
	decisions, err := loadPromotionDecisions(ctx, v.decisions, rec.ID)
	if err != nil {
		return err
	}
	if len(decisions) == 0 {
		return hrerrors.ErrWorkflowRequired
	}
	latest := decisions[0]
	for _, candidate := range decisions[1:] {
		if candidate.StepNumber > latest.StepNumber {
			latest = candidate
		}
	}
	switch rec.Status {
	case "approved":
		if latest.Decision != entity.ApprovalDecisionApprove || latest.ID != *rec.ApprovalDecisionID {
			return hrerrors.ErrWorkflowRequired
		}
		return v.validateSubmittedParents(ctx, rec)
	case "rejected":
		if latest.Decision != entity.ApprovalDecisionReject {
			return hrerrors.ErrWorkflowRequired
		}
	case "returned":
		if latest.Decision != entity.ApprovalDecisionReturn {
			return hrerrors.ErrWorkflowRequired
		}
	}
	return nil
}

func (v *promotionRecommendationValidator) validateApplicationTransition(ctx context.Context, previous, rec *entity.PromotionRecommendation) error {
	if rec.AppliedAssignmentID == nil || rec.AppliedLifecycleEventID == nil || rec.AppliedAt == nil ||
		*rec.AppliedAssignmentID == rec.SourceAssignmentID || previous.ApprovalDecisionID == nil {
		return hrerrors.ErrWorkflowRequired
	}
	unchanged := *rec
	unchanged.Status = previous.Status
	unchanged.AppliedAssignmentID = nil
	unchanged.AppliedLifecycleEventID = nil
	unchanged.AppliedAt = nil
	if !reflect.DeepEqual(previous, &unchanged) {
		return hrerrors.ErrWorkflowRequired
	}
	worker, err := v.workers.FindByIDForUpdate(ctx, rec.WorkerID)
	if err != nil {
		return promotionReferenceError(err)
	}
	if worker == nil || worker.Status != entity.WorkerStatusActive ||
		worker.PrimaryAssignmentID == nil || *worker.PrimaryAssignmentID != *rec.AppliedAssignmentID {
		return hrerrors.ErrReferenceConflict
	}
	source, err := v.assignments.FindByIDForUpdate(ctx, rec.SourceAssignmentID)
	if err != nil {
		return promotionReferenceError(err)
	}
	previousDay := rec.RecommendedEffectiveDate.AddDate(0, 0, -1)
	if source == nil || source.ValidTo == nil || entity.CompareDates(*source.ValidTo, previousDay) != 0 {
		return hrerrors.ErrReferenceConflict
	}
	nextAssignment, err := v.assignments.FindByIDForUpdate(ctx, *rec.AppliedAssignmentID)
	if err != nil {
		return promotionReferenceError(err)
	}
	if nextAssignment == nil || nextAssignment.WorkerID != rec.WorkerID ||
		nextAssignment.PositionID != rec.TargetPositionID ||
		nextAssignment.AssignmentType != "primary" ||
		entity.CompareDates(nextAssignment.ValidFrom, rec.RecommendedEffectiveDate) != 0 {
		return hrerrors.ErrReferenceConflict
	}
	event, err := v.events.FindByIDForUpdate(ctx, *rec.AppliedLifecycleEventID)
	if err != nil {
		return promotionReferenceError(err)
	}
	if event == nil || event.WorkerID != rec.WorkerID || event.EventType != "promotion" ||
		entity.CompareDates(event.EffectiveDate, rec.RecommendedEffectiveDate) != 0 ||
		event.SourceReference == nil || *event.SourceReference != rec.ID {
		return hrerrors.ErrReferenceConflict
	}
	target, err := v.positions.FindByIDForUpdate(ctx, rec.TargetPositionID)
	if err != nil {
		return promotionReferenceError(err)
	}
	if target == nil || target.PositionStatus != "active" ||
		target.LegalEntityReferenceID != worker.PrimaryLegalEntityReferenceID ||
		!assignmentPeriodWithin(nextAssignment.ValidFrom, nextAssignment.ValidTo, target.ValidFrom, target.ValidTo) {
		return hrerrors.ErrReferenceConflict
	}
	return verifyPlacementPosition(ctx, v.positionOwner, target, nextAssignment)
}

func (v *promotionRecommendationValidator) rejectIfReferenced(ctx context.Context, id string) error {
	revCount, err := v.recommendations.CountWhere(ctx, []domainrepo.Filter{
		{Column: "supersedes_recommendation_id", Values: []any{id}},
	})
	if err != nil {
		return err
	}
	if revCount > 0 {
		return hrerrors.ErrReferenceConflict
	}
	decCount, err := v.decisions.CountWhere(ctx, []domainrepo.Filter{
		{Column: "recommendation_id", Values: []any{id}},
	})
	if err != nil {
		return err
	}
	if decCount > 0 {
		return hrerrors.ErrReferenceConflict
	}
	return nil
}

// ValidateMutation menjaga keputusan promosi tetap append-only, berurutan, dan bebas dari persetujuan mandiri.
func (v *promotionDecisionValidator) ValidateMutation(ctx context.Context, table string, before, after entity.Validatable) error {
	if v == nil || table != "hr_promotion_decisions" ||
		operation.RequireDependency(v.recommendations) != nil ||
		operation.RequireDependency(v.workers) != nil ||
		operation.RequireDependency(v.decisions) != nil {
		return hrerrors.ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if before == nil && after == nil {
		return hrerrors.ErrInvalidInput
	}
	// Trigger tr_hr_promotion_decision_immutable menolak ubah/hapus riwayat keputusan.
	if before != nil || after == nil {
		return hrerrors.ErrWorkflowRequired
	}
	decision, ok := after.(*entity.PromotionDecision)
	if !ok || decision == nil || decision.StepNumber <= 0 {
		return hrerrors.ErrInvalidInput
	}
	caller, ok := authcontext.FromContext(ctx)
	if !ok || !caller.Authenticated {
		return hrerrors.ErrUnauthenticated
	}
	if !strings.EqualFold(caller.PrincipalID, decision.DecidedByPrincipalID) {
		return hrerrors.ErrForbidden
	}
	rec, err := v.recommendations.FindByIDForUpdate(ctx, decision.RecommendationID)
	if err != nil {
		return promotionReferenceError(err)
	}
	if rec == nil || rec.Status != "submitted" {
		return hrerrors.ErrWorkflowRequired
	}
	if strings.EqualFold(rec.RecommendedByPrincipalID, decision.DecidedByPrincipalID) {
		return hrerrors.ErrForbidden
	}
	worker, err := v.workers.FindByIDForUpdate(ctx, rec.WorkerID)
	if err != nil {
		return promotionReferenceError(err)
	}
	if worker == nil || worker.Status != entity.WorkerStatusActive {
		return hrerrors.ErrReferenceConflict
	}
	if worker.IAMPrincipalReferenceID != nil && strings.EqualFold(*worker.IAMPrincipalReferenceID, decision.DecidedByPrincipalID) {
		return hrerrors.ErrForbidden
	}
	existingDecisions, err := loadPromotionDecisions(ctx, v.decisions, decision.RecommendationID)
	if err != nil {
		return err
	}
	for _, existing := range existingDecisions {
		if existing.ID == decision.ID {
			continue
		}
		if existing.StepNumber == decision.StepNumber {
			return hrerrors.ErrAlreadyExists
		}
		if existing.StepNumber > decision.StepNumber {
			return hrerrors.ErrWorkflowRequired
		}
	}
	byIdem, err := v.decisions.SelectWhere(ctx, []domainrepo.Filter{
		{Column: "idempotency_key", Values: []any{decision.IdempotencyKey}},
	}, 2, 0)
	if err != nil {
		return err
	}
	for _, existing := range byIdem {
		if existing.ID != decision.ID {
			return hrerrors.ErrAlreadyExists
		}
	}
	return nil
}

func promotionReferenceError(err error) error {
	return reviewReferenceError(err)
}
