// Package service memeriksa bukti promosi tanpa memberi grant IAM atau mengubah jabatan otomatis.
package service

import (
	"strings"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
)

// ValidateDecision menjaga independensi pengusul dan urutan tahap; mandat IAM tetap diperiksa use case.
func ValidateDecision(recommendation *entity.PromotionRecommendation, worker *entity.Worker, decision *entity.PromotionDecision, previous []entity.PromotionDecision) error {
	for _, record := range []entity.Validatable{recommendation, worker, decision} {
		if err := record.Validate(); err != nil {
			return err
		}
	}
	if recommendation.Status != "submitted" || !strings.EqualFold(recommendation.WorkerID, worker.ID) || !strings.EqualFold(decision.RecommendationID, recommendation.ID) || strings.EqualFold(decision.DecidedByPrincipalID, recommendation.RecommendedByPrincipalID) || (worker.IAMPrincipalReferenceID != nil && strings.EqualFold(decision.DecidedByPrincipalID, *worker.IAMPrincipalReferenceID)) {
		return hrerrors.Invalid("decision", "submitted recommendation requires an independent decision maker")
	}
	for _, prior := range previous {
		if err := prior.Validate(); err != nil {
			return err
		}
		if !strings.EqualFold(prior.RecommendationID, recommendation.ID) || decision.StepNumber <= prior.StepNumber || strings.EqualFold(prior.ID, decision.ID) {
			return hrerrors.Invalid("step_number", "decision steps must increase within one recommendation")
		}
	}
	return nil
}

// BuildAppliedRecommendation menghasilkan salinan applied setelah semua bukti assignment/event tersedia.
// Pemanggil wajib menyimpan worker, assignment, event, recommendation dan outbox secara atomik.
// Fungsi ini tidak menyimpan record, tidak mengeksekusi approval, dan tidak mengaktifkan akses.
func BuildAppliedRecommendation(input ApplicationInput) (*entity.PromotionRecommendation, error) {
	for _, record := range []entity.Validatable{input.Recommendation, input.Decision, input.Assessment, input.Worker, input.SourceAssignment, input.NewAssignment, input.TargetPosition, input.Event} {
		if err := record.Validate(); err != nil {
			return nil, err
		}
	}
	r, d, a, w, source, next, target, event := input.Recommendation, input.Decision, input.Assessment, input.Worker, input.SourceAssignment, input.NewAssignment, input.TargetPosition, input.Event
	if input.BusinessDate.IsZero() || input.AppliedAt.IsZero() || entity.CompareDates(r.RecommendedEffectiveDate, input.BusinessDate) > 0 || entity.CompareDates(r.RecommendedEffectiveDate, input.AppliedAt) > 0 {
		return nil, hrerrors.Invalid("recommended_effective_date", "future promotion cannot be applied")
	}
	if entity.CompareDates(input.AppliedAt, input.BusinessDate) != 0 || d.DecidedAt.After(input.AppliedAt) || d.DecidedAt.Before(r.RecommendedAt) || event.OccurredAt.After(input.AppliedAt) {
		return nil, hrerrors.Invalid("applied_at", "approval, lifecycle evidence and server business date must be consistent")
	}
	if r.Status != "approved" || r.ApprovalDecisionID == nil || !strings.EqualFold(*r.ApprovalDecisionID, d.ID) || !strings.EqualFold(d.RecommendationID, r.ID) || d.Decision != entity.ApprovalDecisionApprove || strings.EqualFold(d.DecidedByPrincipalID, r.RecommendedByPrincipalID) || (w.IAMPrincipalReferenceID != nil && strings.EqualFold(d.DecidedByPrincipalID, *w.IAMPrincipalReferenceID)) {
		return nil, hrerrors.Invalid("approval_decision_id", "matching independent approval is required")
	}
	if a.Status != "finalized" || !strings.EqualFold(a.ID, r.AssessmentID) || !strings.EqualFold(a.WorkerID, w.ID) || !strings.EqualFold(r.WorkerID, w.ID) || !strings.EqualFold(source.ID, r.SourceAssignmentID) || !strings.EqualFold(source.WorkerID, w.ID) || !strings.EqualFold(next.WorkerID, w.ID) || source.AssignmentType != "primary" || next.AssignmentType != "primary" || strings.EqualFold(source.PositionID, target.ID) || !strings.EqualFold(next.PositionID, target.ID) || !strings.EqualFold(target.ID, r.TargetPositionID) || strings.EqualFold(source.ID, next.ID) {
		return nil, hrerrors.Invalid("assignment", "promotion worker, assessment and primary positions must match")
	}
	if w.Status != entity.WorkerStatusActive || target.PositionStatus != "active" || !strings.EqualFold(target.LegalEntityReferenceID, w.PrimaryLegalEntityReferenceID) || w.PrimaryAssignmentID == nil || !strings.EqualFold(*w.PrimaryAssignmentID, next.ID) {
		return nil, hrerrors.Invalid("primary_assignment_id", "active employment and target assignment evidence required")
	}
	previousDay := r.RecommendedEffectiveDate.AddDate(0, 0, -1)
	if source.ValidTo == nil || entity.CompareDates(*source.ValidTo, previousDay) != 0 || entity.CompareDates(next.ValidFrom, r.RecommendedEffectiveDate) != 0 || entity.CompareDates(target.ValidFrom, r.RecommendedEffectiveDate) > 0 || (target.ValidTo != nil && entity.CompareDates(*target.ValidTo, r.RecommendedEffectiveDate) < 0) {
		return nil, hrerrors.Invalid("valid_from", "assignment dates must match the approved effective date")
	}
	if event.ID <= 0 || !strings.EqualFold(event.WorkerID, w.ID) || event.EventType != "promotion" || event.SourceReference == nil || !strings.EqualFold(*event.SourceReference, r.ID) || entity.CompareDates(event.EffectiveDate, r.RecommendedEffectiveDate) != 0 {
		return nil, hrerrors.Invalid("event", "persisted promotion lifecycle evidence required")
	}
	result := *r
	result.Status = "applied"
	assignmentID := next.ID
	result.AppliedAssignmentID = &assignmentID
	eventID := event.ID
	result.AppliedLifecycleEventID = &eventID
	at := input.AppliedAt
	result.AppliedAt = &at
	if err := result.Validate(); err != nil {
		return nil, err
	}
	return &result, nil
}
