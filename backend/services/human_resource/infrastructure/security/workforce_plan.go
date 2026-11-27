package security

import (
	"context"
	"errors"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository"
)

type workforcePlanReader interface {
	FindByIDForUpdate(context.Context, string) (*entity.WorkforcePlan, error)
	SelectWhere(context.Context, []domainrepo.Filter, int, int) ([]entity.WorkforcePlan, error)
}

type workforcePlanLineReader interface {
	CountWhere(context.Context, []domainrepo.Filter) (int64, error)
	SelectWhere(context.Context, []domainrepo.Filter, int, int) ([]entity.WorkforcePlanLine, error)
}

type workforcePlanValidator struct {
	jobs       positionJobReader
	plans      workforcePlanReader
	lines      workforcePlanLineReader
	foundation FoundationVerifier
}

type workforcePlanLineValidator struct {
	plans        workforcePlanReader
	jobs         positionJobReader
	lines        workforcePlanLineReader
	requisitions positionReferenceCounter
	foundation   FoundationVerifier
}

// WorkforcePlanValidators memasang aturan draf perencanaan tenaga kerja serta baris kebutuhannya.
func WorkforcePlanValidators(repos *repository.Registry, foundation FoundationVerifier) (map[string]MutationValidator, error) {
	if repos == nil || operation.RequireDependency(repos.WorkforcePlan) != nil ||
		operation.RequireDependency(repos.WorkforcePlanLine) != nil ||
		operation.RequireDependency(repos.Job) != nil ||
		operation.RequireDependency(repos.Requisition) != nil {
		return nil, hrerrors.ErrUnavailable
	}
	return map[string]MutationValidator{
		"hr_workforce_plans": &workforcePlanValidator{
			plans: repos.WorkforcePlan, lines: repos.WorkforcePlanLine, jobs: repos.Job, foundation: foundation,
		},
		"hr_workforce_plan_lines": &workforcePlanLineValidator{
			plans: repos.WorkforcePlan, jobs: repos.Job, lines: repos.WorkforcePlanLine,
			requisitions: repos.Requisition, foundation: foundation,
		},
	}, nil
}

func (v *workforcePlanValidator) ValidateMutation(ctx context.Context, table string, before, after entity.Validatable) error {
	if v == nil || table != "hr_workforce_plans" || operation.RequireDependency(v.plans) != nil ||
		operation.RequireDependency(v.lines) != nil {
		return hrerrors.ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	var previous *entity.WorkforcePlan
	if before != nil {
		var ok bool
		previous, ok = before.(*entity.WorkforcePlan)
		if !ok || previous == nil {
			return hrerrors.ErrInvalidInput
		}
		if previous.Status != "draft" && previous.Status != "review" && previous.Status != "approved" {
			return hrerrors.ErrWorkflowRequired
		}
	}
	if after == nil {
		if previous == nil {
			return hrerrors.ErrInvalidInput
		}
		if previous.Status != "draft" {
			return hrerrors.ErrWorkflowRequired
		}
		count, err := v.lines.CountWhere(ctx, []domainrepo.Filter{{Column: "workforce_plan_id", Values: []any{previous.ID}}})
		if err != nil {
			return err
		}
		if count > 0 {
			return hrerrors.ErrReferenceConflict
		}
		return nil
	}
	plan, ok := after.(*entity.WorkforcePlan)
	if !ok || plan == nil {
		return hrerrors.ErrInvalidInput
	}
	if previous != nil && previous.Status == "approved" && plan.Status == "superseded" {
		return v.verifySupersession(ctx, previous, plan)
	}
	if plan.SupersededByPlanID != nil || plan.SupersessionReasonCode != nil {
		return hrerrors.ErrWorkflowRequired
	}
	if previous != nil && plan.Status == "cancelled" {
		expected := *previous
		expected.Status = "cancelled"
		if previous.ApprovedAt != nil || previous.ApprovedByPrincipalID != nil ||
			previous.FinanceBudgetReferenceID != nil || !reflect.DeepEqual(&expected, plan) {
			return hrerrors.ErrWorkflowRequired
		}
		return nil
	}
	approving := previous != nil && previous.Status == "review" && plan.Status == "approved"
	if approving {
		expected := *previous
		expected.Status = "approved"
		expected.ApprovedAt = plan.ApprovedAt
		expected.ApprovedByPrincipalID = plan.ApprovedByPrincipalID
		expected.FinanceBudgetReferenceID = plan.FinanceBudgetReferenceID
		if previous.ApprovedAt != nil || previous.ApprovedByPrincipalID != nil || previous.FinanceBudgetReferenceID != nil ||
			plan.ApprovedAt == nil || plan.ApprovedByPrincipalID == nil || plan.FinanceBudgetReferenceID == nil ||
			strings.EqualFold(plan.CreatedByPrincipalID, *plan.ApprovedByPrincipalID) ||
			plan.ApprovedAt.Before(plan.CreatedAt) || entity.CompareDates(plan.ApprovedAt.UTC(), plan.PeriodEnd) > 0 ||
			!reflect.DeepEqual(&expected, plan) {
			return hrerrors.ErrWorkflowRequired
		}
	} else if previous != nil && previous.Status != "draft" {
		return hrerrors.ErrWorkflowRequired
	}
	submitting := previous != nil && plan.Status == "review"
	if submitting {
		expected := *previous
		expected.Status = "review"
		if !reflect.DeepEqual(&expected, plan) {
			return hrerrors.ErrWorkflowRequired
		}
	}
	if !approving && ((!submitting && plan.Status != "draft") || plan.ApprovedByPrincipalID != nil || plan.ApprovedAt != nil) {
		return hrerrors.ErrWorkflowRequired
	}
	if !approving && plan.FinanceBudgetReferenceID != nil {
		return hrerrors.ErrUnavailable
	}
	if previous != nil && (previous.ID != plan.ID || previous.PlanCode != plan.PlanCode ||
		previous.PlanVersion != plan.PlanVersion || previous.LegalEntityReferenceID != plan.LegalEntityReferenceID ||
		previous.CreatedByPrincipalID != plan.CreatedByPrincipalID || !previous.CreatedAt.Equal(plan.CreatedAt)) {
		return hrerrors.ErrWorkflowRequired
	}
	if err := verifyFoundationLegalEntity(ctx, v.foundation, plan.LegalEntityReferenceID); err != nil {
		return err
	}
	matches, err := v.plans.SelectWhere(ctx, []domainrepo.Filter{
		{Column: "legal_entity_reference_id", Values: []any{plan.LegalEntityReferenceID}},
		{Column: "plan_code", Values: []any{plan.PlanCode}},
		{Column: "plan_version", Values: []any{plan.PlanVersion}},
	}, 2, 0)
	if err != nil {
		return err
	}
	for _, existing := range matches {
		if existing.ID != plan.ID {
			return hrerrors.ErrAlreadyExists
		}
	}
	checkingReferences := submitting || approving
	if checkingReferences || (previous != nil && (!previous.PeriodStart.Equal(plan.PeriodStart) || !previous.PeriodEnd.Equal(plan.PeriodEnd))) {
		count, err := v.lines.CountWhere(ctx, []domainrepo.Filter{{Column: "workforce_plan_id", Values: []any{plan.ID}}})
		if err != nil {
			return err
		}
		if checkingReferences && count == 0 {
			return hrerrors.ErrReferenceConflict
		}
		jobWindows := make(map[string]workforceJobWindow)
		if count > 0 {
			for offset := 0; int64(offset) < count; offset += domainrepo.MaxLimit {
				lines, err := v.lines.SelectWhere(ctx, []domainrepo.Filter{{Column: "workforce_plan_id", Values: []any{plan.ID}}}, domainrepo.MaxLimit, offset)
				if err != nil {
					return err
				}
				if len(lines) == 0 {
					return hrerrors.ErrReferenceConflict
				}
				for i := range lines {
					if !planPeriodContainsLine(plan, &lines[i]) {
						return hrerrors.ErrReferenceConflict
					}
					if checkingReferences {
						line := &lines[i]
						jobWindows[line.JobID] = jobWindows[line.JobID].include(line)
						if err := verifyWorkforceLineFoundation(ctx, v.foundation, plan, &lines[i]); err != nil {
							return err
						}
					}
				}
			}
		}
		if checkingReferences {
			if err := verifyWorkforcePlanJobs(ctx, v.jobs, jobWindows); err != nil {
				return err
			}
		}
	}
	return nil
}

func (v *workforcePlanValidator) verifySupersession(ctx context.Context, previous, next *entity.WorkforcePlan) error {
	if previous.ApprovedAt == nil || previous.ApprovedByPrincipalID == nil || previous.FinanceBudgetReferenceID == nil ||
		previous.SupersededByPlanID != nil || previous.SupersessionReasonCode != nil ||
		next.SupersededByPlanID == nil || next.SupersessionReasonCode == nil ||
		strings.TrimSpace(*next.SupersessionReasonCode) == "" {
		return hrerrors.ErrWorkflowRequired
	}
	expected := *previous
	expected.Status = "superseded"
	expected.SupersededByPlanID = next.SupersededByPlanID
	expected.SupersessionReasonCode = next.SupersessionReasonCode
	if !reflect.DeepEqual(&expected, next) {
		return hrerrors.ErrWorkflowRequired
	}
	id := *next.SupersededByPlanID
	// Periksa arah versi sebelum mengambil lock kedua agar urutan lock tetap naik.
	candidates, err := v.plans.SelectWhere(ctx, []domainrepo.Filter{{Column: "id", Values: []any{id}}}, 1, 0)
	if err != nil {
		return err
	}
	if len(candidates) != 1 || !validWorkforceReplacement(previous, &candidates[0]) {
		return hrerrors.ErrReferenceConflict
	}
	replacement, err := v.plans.FindByIDForUpdate(ctx, id)
	if err != nil {
		return workerAssessmentReferenceError(err)
	}
	if !validWorkforceReplacement(previous, replacement) {
		return hrerrors.ErrReferenceConflict
	}
	return nil
}

func validWorkforceReplacement(old, replacement *entity.WorkforcePlan) bool {
	return replacement != nil && replacement.Status == "approved" &&
		replacement.ApprovedAt != nil && replacement.ApprovedByPrincipalID != nil && replacement.FinanceBudgetReferenceID != nil &&
		replacement.PlanVersion > old.PlanVersion && replacement.PlanCode == old.PlanCode &&
		replacement.LegalEntityReferenceID == old.LegalEntityReferenceID && replacement.ScenarioCode == old.ScenarioCode &&
		entity.CompareDates(replacement.PeriodStart, old.PeriodStart) == 0 &&
		entity.CompareDates(replacement.PeriodEnd, old.PeriodEnd) == 0
}

func (v *workforcePlanLineValidator) ValidateMutation(ctx context.Context, table string, before, after entity.Validatable) error {
	if v == nil || table != "hr_workforce_plan_lines" || operation.RequireDependency(v.plans) != nil ||
		operation.RequireDependency(v.jobs) != nil || operation.RequireDependency(v.lines) != nil ||
		operation.RequireDependency(v.requisitions) != nil {
		return hrerrors.ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	var previous *entity.WorkforcePlanLine
	if before != nil {
		var ok bool
		previous, ok = before.(*entity.WorkforcePlanLine)
		if !ok || previous == nil {
			return hrerrors.ErrInvalidInput
		}
	}
	if previous == nil && after == nil {
		return hrerrors.ErrInvalidInput
	}
	target := previous
	var line *entity.WorkforcePlanLine
	if after != nil {
		var ok bool
		line, ok = after.(*entity.WorkforcePlanLine)
		if !ok || line == nil {
			return hrerrors.ErrInvalidInput
		}
		if previous != nil && (previous.ID != line.ID || previous.WorkforcePlanID != line.WorkforcePlanID) {
			return hrerrors.ErrWorkflowRequired
		}
		target = line
	}
	plan, err := v.plans.FindByIDForUpdate(ctx, target.WorkforcePlanID)
	if err != nil {
		return workerAssessmentReferenceError(err)
	}
	if plan == nil || plan.Status != "draft" {
		return hrerrors.ErrWorkflowRequired
	}
	if previous != nil && (after == nil || !reflect.DeepEqual(previous, line)) {
		count, err := v.requisitions.CountWhere(ctx, []domainrepo.Filter{{Column: "workforce_plan_line_id", Values: []any{target.ID}}})
		if err != nil {
			return err
		}
		if count > 0 {
			return hrerrors.ErrReferenceConflict
		}
	}
	if after == nil {
		return nil
	}
	if !planPeriodContainsLine(plan, line) {
		return hrerrors.ErrReferenceConflict
	}
	if err := verifyWorkforceJob(ctx, v.jobs, line.JobID, workforceJobWindow{from: line.NeededFrom, to: line.NeededUntil}); err != nil {
		return err
	}
	if err := verifyWorkforceLineFoundation(ctx, v.foundation, plan, line); err != nil {
		return err
	}
	matches, err := v.lines.SelectWhere(ctx, []domainrepo.Filter{
		{Column: "workforce_plan_id", Values: []any{line.WorkforcePlanID}},
		{Column: "line_number", Values: []any{line.LineNumber}},
	}, 2, 0)
	if err != nil {
		return err
	}
	for _, existing := range matches {
		if existing.ID != line.ID {
			return hrerrors.ErrAlreadyExists
		}
	}
	return nil
}

type workforceJobWindow struct {
	from time.Time
	to   *time.Time
}

func (w workforceJobWindow) include(line *entity.WorkforcePlanLine) workforceJobWindow {
	if w.from.IsZero() {
		return workforceJobWindow{from: line.NeededFrom, to: line.NeededUntil}
	}
	if entity.CompareDates(line.NeededFrom, w.from) < 0 {
		w.from = line.NeededFrom
	}
	if w.to != nil && (line.NeededUntil == nil || entity.CompareDates(*line.NeededUntil, *w.to) > 0) {
		w.to = line.NeededUntil
	}
	return w
}

// verifyWorkforcePlanJobs mengunci jabatan dalam urutan konsisten antar rencana.
func verifyWorkforcePlanJobs(ctx context.Context, jobs positionJobReader, windows map[string]workforceJobWindow) error {
	if operation.RequireDependency(jobs) != nil {
		return hrerrors.ErrUnavailable
	}
	ids := make([]string, 0, len(windows))
	for id := range windows {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		if err := verifyWorkforceJob(ctx, jobs, id, windows[id]); err != nil {
			return err
		}
	}
	return nil
}

func verifyWorkforceJob(ctx context.Context, jobs positionJobReader, id string, window workforceJobWindow) error {
	job, err := jobs.FindByIDForUpdate(ctx, id)
	if err != nil {
		return workerAssessmentReferenceError(err)
	}
	if job == nil || !job.IsActive || entity.CompareDates(window.from, job.ValidFrom) < 0 ||
		(job.ValidTo != nil && (window.to == nil || entity.CompareDates(*window.to, *job.ValidTo) > 0)) {
		return hrerrors.ErrReferenceConflict
	}
	return nil
}

func verifyWorkforceLineFoundation(ctx context.Context, foundation FoundationVerifier, plan *entity.WorkforcePlan, line *entity.WorkforcePlanLine) error {
	if operation.RequireDependency(foundation) != nil {
		return hrerrors.ErrUnavailable
	}
	if err := foundation.VerifyPositionReferences(ctx, PositionReferenceCheck{
		OrganizationUnitID: line.OrganizationUnitReferenceID,
		LegalEntityID:      plan.LegalEntityReferenceID,
		WorkLocationID:     line.WorkLocationReferenceID,
		ValidFrom:          line.NeededFrom,
		ValidTo:            line.NeededUntil,
	}); err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if errors.Is(err, hrerrors.ErrReferenceConflict) {
			return hrerrors.ErrReferenceConflict
		}
		return hrerrors.ErrUnavailable
	}
	return nil
}

func planPeriodContainsLine(plan *entity.WorkforcePlan, line *entity.WorkforcePlanLine) bool {
	if plan == nil || line == nil {
		return false
	}
	if entity.CompareDates(line.NeededFrom, plan.PeriodStart) < 0 ||
		entity.CompareDates(line.NeededFrom, plan.PeriodEnd) > 0 {
		return false
	}
	if line.NeededUntil != nil && entity.CompareDates(*line.NeededUntil, plan.PeriodEnd) > 0 {
		return false
	}
	return true
}
