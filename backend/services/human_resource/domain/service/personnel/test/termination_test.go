package service_test

import (
	"errors"
	"testing"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	service "github.com/integrione/erp-retail/backend/services/human_resource/domain/service/personnel"
	f "github.com/integrione/erp-retail/backend/services/human_resource/domain/service/test/fixture"
)

func TestBuildTerminatedWorker(t *testing.T) {
	t.Parallel()
	now := f.Date().AddDate(0, 0, 1)
	worker := f.Worker()
	assignmentID, approver := f.ID(2), f.ID(3)
	worker.PrimaryAssignmentID = &assignmentID
	plan := entity.OffboardingPlan{
		ID: f.ID(4), WorkerID: worker.ID, PlanNumber: "EXIT-1", ReasonCode: "RESIGNATION",
		NoticeDate: f.Date(), LastWorkingDate: f.Date(), TerminationEffectiveDate: now,
		Status: "in_progress", ApprovedByPrincipalID: &approver, ApprovedAt: &worker.CreatedAt,
		CreatedAt: worker.CreatedAt, CreatedByPrincipalID: f.ID(5),
	}
	result, err := service.BuildTerminatedWorker(&worker, &plan, now)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != entity.WorkerStatusTerminated || result.PrimaryAssignmentID != nil ||
		result.TerminationDate == nil || !result.TerminationDate.Equal(now) ||
		result.TerminationReasonCode == nil || *result.TerminationReasonCode != plan.ReasonCode {
		t.Fatalf("unexpected termination snapshot: %+v", result)
	}
	if worker.Status != entity.WorkerStatusActive || worker.TerminationDate != nil || worker.PrimaryAssignmentID == nil {
		t.Fatalf("source worker was changed: %+v", worker)
	}
	if _, err := service.BuildTerminatedWorker(&worker, &plan, f.Date()); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("termination before effective date: %v", err)
	}
	plan.Status = "approved"
	if _, err := service.BuildTerminatedWorker(&worker, &plan, now); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("termination before plan starts: %v", err)
	}
}
