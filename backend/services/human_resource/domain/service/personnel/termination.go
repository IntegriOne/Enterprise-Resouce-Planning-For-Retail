package service

import (
	"strings"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
)

// BuildTerminatedWorker membentuk perubahan pekerja dari keputusan offboarding.
func BuildTerminatedWorker(worker *entity.Worker, plan *entity.OffboardingPlan, now time.Time) (*entity.Worker, error) {
	if worker == nil || plan == nil || !strings.EqualFold(plan.WorkerID, worker.ID) {
		return nil, hrerrors.ErrReferenceConflict
	}
	if (worker.Status != entity.WorkerStatusActive && worker.Status != entity.WorkerStatusNoticePeriod) ||
		worker.TerminationDate != nil || worker.TerminationReasonCode != nil ||
		plan.Status != "in_progress" || plan.ApprovedAt == nil || plan.ApprovedByPrincipalID == nil ||
		plan.CompletedAt != nil || strings.EqualFold(plan.CreatedByPrincipalID, *plan.ApprovedByPrincipalID) {
		return nil, hrerrors.ErrWorkflowRequired
	}
	if err := plan.Validate(); err != nil {
		return nil, err
	}
	if now.IsZero() || now.Before(worker.UpdatedAt) || now.Before(*plan.ApprovedAt) ||
		plan.ApprovedAt.Before(plan.CreatedAt) ||
		entity.CompareDates(now.UTC(), plan.TerminationEffectiveDate) < 0 ||
		(worker.HireDate != nil && entity.CompareDates(plan.LastWorkingDate, *worker.HireDate) < 0) ||
		(worker.ServiceStartDate != nil && entity.CompareDates(plan.LastWorkingDate, *worker.ServiceStartDate) < 0) {
		return nil, hrerrors.ErrReferenceConflict
	}
	if worker.IAMPrincipalReferenceID != nil && strings.EqualFold(*worker.IAMPrincipalReferenceID, *plan.ApprovedByPrincipalID) {
		return nil, hrerrors.ErrForbidden
	}
	terminated := *worker
	effectiveDate, reasonCode := plan.TerminationEffectiveDate, plan.ReasonCode
	terminated.Status = entity.WorkerStatusTerminated
	terminated.TerminationDate = &effectiveDate
	terminated.TerminationReasonCode = &reasonCode
	terminated.PrimaryAssignmentID = nil
	terminated.UpdatedAt = now
	if err := terminated.Validate(); err != nil {
		return nil, err
	}
	return &terminated, nil
}
