package security

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	"github.com/integrione/erp-retail/backend/shared/authcontext"
)

type performanceReviewReader interface {
	SelectWhere(context.Context, []domainrepo.Filter, int, int) ([]entity.PerformanceReview, error)
}

type performanceReviewValidator struct {
	workers     lockedWorkerReader
	cycles      performanceCycleReader
	reviews     performanceReviewReader
	scores      performanceReferenceCounter
	assessments performanceReferenceCounter
}

// ValidateMutation membatasi CRUD review pada draf mandiri milik pekerja yang terhubung ke IAM.
func (v *performanceReviewValidator) ValidateMutation(ctx context.Context, table string, before, after entity.Validatable) error {
	if v == nil || table != "hr_performance_reviews" || operation.RequireDependency(v.workers) != nil ||
		operation.RequireDependency(v.cycles) != nil || operation.RequireDependency(v.reviews) != nil ||
		operation.RequireDependency(v.scores) != nil || operation.RequireDependency(v.assessments) != nil {
		return hrerrors.ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	caller, ok := authcontext.FromContext(ctx)
	if !ok || !caller.Authenticated {
		return hrerrors.ErrUnauthenticated
	}
	var previous *entity.PerformanceReview
	if before != nil {
		var ok bool
		previous, ok = before.(*entity.PerformanceReview)
		if !ok || previous == nil || !selfReviewDraft(previous) {
			return hrerrors.ErrWorkflowRequired
		}
	}
	if after == nil && previous == nil {
		return hrerrors.ErrInvalidInput
	}
	review := previous
	if after != nil {
		var ok bool
		review, ok = after.(*entity.PerformanceReview)
		if !ok || review == nil {
			return hrerrors.ErrInvalidInput
		}
		if !selfReviewDraft(review) {
			return hrerrors.ErrWorkflowRequired
		}
		if previous != nil && (previous.ID != review.ID || previous.WorkerID != review.WorkerID ||
			previous.PerformanceCycleID != review.PerformanceCycleID || previous.ReviewerPrincipalID != review.ReviewerPrincipalID ||
			previous.ReviewerRole != review.ReviewerRole || previous.ReviewVersion != review.ReviewVersion) {
			return hrerrors.ErrWorkflowRequired
		}
	}
	if !strings.EqualFold(caller.PrincipalID, review.ReviewerPrincipalID) {
		return hrerrors.ErrForbidden
	}
	worker, err := v.workers.FindByIDForUpdate(ctx, review.WorkerID)
	if err != nil {
		return reviewReferenceError(err)
	}
	if worker == nil || worker.Status != entity.WorkerStatusActive || worker.IAMPrincipalReferenceID == nil ||
		!strings.EqualFold(*worker.IAMPrincipalReferenceID, review.ReviewerPrincipalID) {
		return hrerrors.ErrReferenceConflict
	}
	cycle, err := v.cycles.FindByIDForUpdate(ctx, review.PerformanceCycleID)
	if err != nil {
		return reviewReferenceError(err)
	}
	if cycle == nil || cycle.Status != entity.PerformanceCycleStatusOpen {
		return hrerrors.ErrReferenceConflict
	}
	if previous != nil {
		if err := v.rejectIfReferenced(ctx, review.ID); err != nil {
			return err
		}
	}
	if after == nil {
		return nil
	}
	matches, err := v.reviews.SelectWhere(ctx, []domainrepo.Filter{
		{Column: "worker_id", Values: []any{review.WorkerID}},
		{Column: "performance_cycle_id", Values: []any{review.PerformanceCycleID}},
		{Column: "reviewer_principal_id", Values: []any{review.ReviewerPrincipalID}},
		{Column: "reviewer_role", Values: []any{review.ReviewerRole}},
		{Column: "review_version", Values: []any{review.ReviewVersion}},
	}, 2, 0)
	if err != nil {
		return err
	}
	for _, existing := range matches {
		if existing.ID != review.ID {
			return hrerrors.ErrAlreadyExists
		}
	}
	return nil
}

func selfReviewDraft(review *entity.PerformanceReview) bool {
	if review == nil || review.ReviewerRole != "self" || review.ReviewVersion != 1 ||
		review.VisibilityClass != "restricted" || review.OverallRatingCode != nil || review.DevelopmentPlanID != nil ||
		review.SubmittedAt != nil || review.FinalizedAt != nil || len(review.ContentDigest) != 0 {
		return false
	}
	var assessment map[string]json.RawMessage
	return json.Unmarshal(review.CompetencyAssessment, &assessment) == nil && assessment != nil && len(assessment) == 0
}

func (v *performanceReviewValidator) rejectIfReferenced(ctx context.Context, id string) error {
	for _, counter := range []performanceReferenceCounter{v.scores, v.assessments} {
		count, err := counter.CountWhere(ctx, []domainrepo.Filter{{Column: "performance_review_id", Values: []any{id}}})
		if err != nil {
			return err
		}
		if count > 0 {
			return hrerrors.ErrReferenceConflict
		}
	}
	return nil
}

func reviewReferenceError(err error) error {
	if errors.Is(err, hrerrors.ErrNotFound) {
		return hrerrors.ErrReferenceConflict
	}
	return err
}
