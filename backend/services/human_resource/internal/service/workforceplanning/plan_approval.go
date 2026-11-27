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

type WorkforceBudgetCheck struct {
	PlanID, LegalEntityID, ScenarioCode, ApproverPrincipalID string
	PlanVersion                                              int32
	PeriodStart, PeriodEnd                                   time.Time
	Lines                                                    []WorkforceBudgetLine
}

type WorkforceBudgetLine struct {
	ID, OrganizationUnitID, JobID string
	WorkLocationID                *string
	WorkerCategory                entity.WorkerCategory
	ActionCode                    string
	CurrentFTE, RequiredFTE       entity.Decimal
	EstimatedCost                 *entity.Decimal
	CurrencyCode                  *string
	NeededFrom                    time.Time
	NeededUntil                   *time.Time
}

// WorkforceBudgetVerifier membuktikan otorisasi Finance untuk snapshot rencana.
type WorkforceBudgetVerifier interface {
	VerifyWorkforceBudget(context.Context, WorkforceBudgetCheck) (string, error)
}

// ApproveWorkforcePlan menyetujui rencana review setelah bukti Finance diterima.
func (s *Service) ApproveWorkforcePlan(ctx context.Context, id string) (*entity.WorkforcePlan, error) {
	if s == nil || operation.RequireDependency(s.budget) != nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.workforcePlan.Command(ctx, id, "approve", func(txCtx context.Context, plan *entity.WorkforcePlan, actor operation.Actor, now time.Time) (*entity.WorkforcePlan, error) {
		if actor.Kind != "human" || strings.EqualFold(actor.PrincipalID, plan.CreatedByPrincipalID) {
			return nil, hrerrors.ErrForbidden
		}
		if plan.Status != "review" || plan.ApprovedAt != nil || plan.ApprovedByPrincipalID != nil || plan.FinanceBudgetReferenceID != nil {
			return nil, hrerrors.ErrWorkflowRequired
		}
		if entity.CompareDates(now.UTC(), plan.PeriodEnd) > 0 {
			return nil, hrerrors.ErrReferenceConflict
		}
		lines, err := s.workforceBudgetLines(txCtx, plan.ID)
		if err != nil {
			return nil, err
		}
		budgetID, err := s.budget.VerifyWorkforceBudget(txCtx, WorkforceBudgetCheck{
			PlanID: plan.ID, LegalEntityID: plan.LegalEntityReferenceID, ScenarioCode: plan.ScenarioCode, PlanVersion: plan.PlanVersion,
			PeriodStart: plan.PeriodStart, PeriodEnd: plan.PeriodEnd, ApproverPrincipalID: actor.PrincipalID, Lines: lines,
		})
		if err != nil {
			return nil, operation.OwnerVerificationError(txCtx, err)
		}
		if err := entity.ValidateUUID("finance_budget_reference_id", budgetID); err != nil {
			return nil, hrerrors.ErrReferenceConflict
		}
		plan.Status = "approved"
		plan.ApprovedAt = &now
		plan.ApprovedByPrincipalID = &actor.PrincipalID
		plan.FinanceBudgetReferenceID = &budgetID
		return plan, nil
	})
}

func (s *Service) workforceBudgetLines(ctx context.Context, planID string) ([]WorkforceBudgetLine, error) {
	filter := []domainrepo.Filter{{Column: "workforce_plan_id", Values: []any{planID}}}
	var lines []WorkforceBudgetLine
	for offset := 0; ; offset += domainrepo.MaxLimit {
		rows, err := s.planLines.SelectWhere(ctx, filter, domainrepo.MaxLimit, offset)
		if err != nil {
			return nil, err
		}
		for _, row := range rows {
			lines = append(lines, WorkforceBudgetLine{
				ID: row.ID, OrganizationUnitID: row.OrganizationUnitReferenceID, JobID: row.JobID,
				WorkLocationID: row.WorkLocationReferenceID, WorkerCategory: row.WorkerCategory, ActionCode: row.ActionCode,
				CurrentFTE: row.CurrentFTE, RequiredFTE: row.RequiredFTE, EstimatedCost: row.EstimatedCost, CurrencyCode: row.CurrencyCode,
				NeededFrom: row.NeededFrom, NeededUntil: row.NeededUntil,
			})
		}
		if len(rows) < domainrepo.MaxLimit {
			if len(lines) == 0 {
				return nil, hrerrors.ErrReferenceConflict
			}
			return lines, nil
		}
	}
}
