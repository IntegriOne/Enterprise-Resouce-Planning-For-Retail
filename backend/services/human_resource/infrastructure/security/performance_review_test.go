package security

import (
	"context"
	"errors"
	"testing"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	"github.com/integrione/erp-retail/backend/shared/authcontext"
)

type reviewStub struct {
	rows []entity.PerformanceReview
}

func (s *reviewStub) SelectWhere(_ context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.PerformanceReview, error) {
	if len(filters) != 5 || filters[0].Column != "worker_id" || filters[1].Column != "performance_cycle_id" ||
		filters[2].Column != "reviewer_principal_id" || filters[3].Column != "reviewer_role" ||
		filters[4].Column != "review_version" || limit != 2 || offset != 0 {
		return nil, hrerrors.ErrInvalidInput
	}
	return s.rows, nil
}

type reviewReferenceStub struct {
	count int64
}

func (s *reviewReferenceStub) CountWhere(_ context.Context, filters []domainrepo.Filter) (int64, error) {
	if len(filters) != 1 || filters[0].Column != "performance_review_id" {
		return 0, hrerrors.ErrInvalidInput
	}
	return s.count, nil
}

func TestPerformanceReviewValidatorRestrictsSelfDraftAndReferences(t *testing.T) {
	t.Parallel()
	order := []string{}
	principal := "00000000-0000-4000-8000-000000000001"
	ctx := authcontext.WithCaller(t.Context(), authcontext.CallerContext{PrincipalID: principal, Authenticated: true})
	worker := &lockedWorkerStub{worker: &entity.Worker{
		Status: entity.WorkerStatusActive, IAMPrincipalReferenceID: &principal,
	}, order: &order}
	cycle := &performanceCycleStub{cycle: &entity.PerformanceCycle{Status: entity.PerformanceCycleStatusOpen}}
	reviews := &reviewStub{}
	scores := &reviewReferenceStub{}
	assessments := &reviewReferenceStub{}
	validator := &performanceReviewValidator{workers: worker, cycles: cycle, reviews: reviews, scores: scores, assessments: assessments}
	review := &entity.PerformanceReview{
		ID: "review-1", WorkerID: "worker-1", PerformanceCycleID: "cycle-1", ReviewerPrincipalID: principal,
		ReviewerRole: "self", ReviewVersion: 1, CompetencyAssessment: entity.JSON(`{}`), VisibilityClass: "restricted",
	}
	if err := validator.ValidateMutation(ctx, "hr_performance_reviews", nil, review); err != nil || len(order) != 1 {
		t.Fatalf("create self review with active worker and open cycle: %v", err)
	}
	if err := validator.ValidateMutation(t.Context(), "hr_performance_reviews", nil, review); !errors.Is(err, hrerrors.ErrUnauthenticated) {
		t.Fatalf("review without verified caller: %v", err)
	}
	otherCaller := authcontext.WithCaller(t.Context(), authcontext.CallerContext{
		PrincipalID: "00000000-0000-4000-8000-000000000009", Authenticated: true,
	})
	if err := validator.ValidateMutation(otherCaller, "hr_performance_reviews", review, nil); !errors.Is(err, hrerrors.ErrForbidden) {
		t.Fatalf("other principal deleting self review: %v", err)
	}
	reviews.rows = []entity.PerformanceReview{{ID: "review-2"}}
	if err := validator.ValidateMutation(ctx, "hr_performance_reviews", nil, review); !errors.Is(err, hrerrors.ErrAlreadyExists) {
		t.Fatalf("duplicate review version: %v", err)
	}
	reviews.rows = nil
	spoofed := *review
	spoofed.ReviewerPrincipalID = "00000000-0000-4000-8000-000000000009"
	if err := validator.ValidateMutation(ctx, "hr_performance_reviews", nil, &spoofed); !errors.Is(err, hrerrors.ErrForbidden) {
		t.Fatalf("reviewer does not own worker identity: %v", err)
	}
	worker.worker.IAMPrincipalReferenceID = &spoofed.ReviewerPrincipalID
	if err := validator.ValidateMutation(ctx, "hr_performance_reviews", nil, review); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("worker IAM link does not match reviewer: %v", err)
	}
	worker.worker.IAMPrincipalReferenceID = &principal
	manager := *review
	manager.ReviewerRole = "manager"
	if err := validator.ValidateMutation(ctx, "hr_performance_reviews", nil, &manager); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("manager review without relationship validation: %v", err)
	}
	nullAssessment := *review
	nullAssessment.CompetencyAssessment = entity.JSON(`null`)
	if err := validator.ValidateMutation(ctx, "hr_performance_reviews", nil, &nullAssessment); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("null assessment payload: %v", err)
	}
	cycle.cycle.Status = entity.PerformanceCycleStatusDraft
	if err := validator.ValidateMutation(ctx, "hr_performance_reviews", nil, review); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("review in draft cycle: %v", err)
	}
	cycle.cycle.Status = entity.PerformanceCycleStatusOpen
	changed := *review
	narrative := "Progress documented"
	changed.Narrative = &narrative
	if err := validator.ValidateMutation(ctx, "hr_performance_reviews", review, &changed); err != nil {
		t.Fatalf("edit self draft: %v", err)
	}
	changed.ReviewVersion = 2
	if err := validator.ValidateMutation(ctx, "hr_performance_reviews", review, &changed); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("change review version through CRUD: %v", err)
	}
	scores.count = 1
	if err := validator.ValidateMutation(ctx, "hr_performance_reviews", review, nil); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("delete scored review: %v", err)
	}
	scores.count = 0
	assessments.count = 1
	if err := validator.ValidateMutation(ctx, "hr_performance_reviews", review, nil); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("delete assessed review: %v", err)
	}
	assessments.count = 0
	if err := validator.ValidateMutation(ctx, "hr_performance_reviews", review, nil); err != nil {
		t.Fatalf("delete unused self draft: %v", err)
	}
}
