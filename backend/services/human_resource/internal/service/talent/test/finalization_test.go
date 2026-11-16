package service_test

import (
	"errors"
	"testing"

	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	"github.com/integrione/erp-retail/backend/services/human_resource/domain/service/test/fixture"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

func TestFinalizeTalentAssessment(t *testing.T) {
	sentinel := errors.New("outbox unavailable")
	for _, name := range []string{"finalized", "missing_score", "closed_cycle", "hidden_dimensions", "missing_review_reader", "outbox_rollback"} {
		t.Run(name, func(t *testing.T) {
			h, svc, _, frameworks, dimensions, cycles, assessments, scores := setupTalent(t)
			assessment, cycle := fixture.Assessment(), fixture.Cycle()
			framework, dimension, score := fixture.Framework(), fixture.Dimension(), fixture.Score()
			var want error
			switch name {
			case "missing_score":
				want = hrerrors.ErrInvalidInput
			case "closed_cycle":
				cycle.Status = "closed"
				want = hrerrors.ErrInvalidInput
			case "hidden_dimensions":
				h.Scope = operation.Scope{Filters: []domainrepo.Filter{{Column: "id", Values: []any{fixture.ID(999)}}}}
				want = hrerrors.ErrForbidden
			case "missing_review_reader":
				id := fixture.ID(80)
				assessment.PerformanceReviewID = &id
				want = hrerrors.ErrUnavailable
			case "outbox_rollback":
				h.Failures["outbox"] = sentinel
				want = sentinel
			}
			frameworks.Rows[framework.ID] = framework
			dimensions.Rows[dimension.ID] = dimension
			cycles.Rows[cycle.ID] = cycle
			assessments.Rows[assessment.ID] = assessment
			if name != "missing_score" {
				scores.Rows[score.ID] = score
			}
			got, err := svc.FinalizeTalentAssessment(t.Context(), assessment.ID)
			if !errors.Is(err, want) {
				t.Fatalf("finalization error = %v, want %v", err, want)
			}
			stored := assessments.Rows[assessment.ID]
			if want != nil {
				if got != nil || stored.Status != "draft" || stored.FinalizedAt != nil || len(h.Events) != 0 {
					t.Fatal("rejected finalization persisted state or event")
				}
				return
			}
			if got == nil || stored.Status != "finalized" || stored.FinalizedAt == nil ||
				stored.FinalizedByPrincipalID == nil || *stored.FinalizedByPrincipalID != h.Actor.PrincipalID || len(h.Events) != 1 {
				t.Fatal("finalization evidence and outbox must commit together")
			}
		})
	}
}
