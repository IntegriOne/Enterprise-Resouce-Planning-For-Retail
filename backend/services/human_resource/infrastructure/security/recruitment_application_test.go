package security

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
)

type applicationCandidateStub struct {
	candidate *entity.Candidate
}

func (s *applicationCandidateStub) FindByIDForUpdate(_ context.Context, id string) (*entity.Candidate, error) {
	if s.candidate == nil || s.candidate.ID != id {
		return nil, hrerrors.ErrNotFound
	}
	return s.candidate, nil
}

type applicationRequisitionStub struct {
	requisition *entity.Requisition
}

func (s *applicationRequisitionStub) FindByIDForUpdate(_ context.Context, id string) (*entity.Requisition, error) {
	if s.requisition == nil || s.requisition.ID != id {
		return nil, hrerrors.ErrNotFound
	}
	return s.requisition, nil
}

func TestApplicationValidatorReceivedOnly(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 11, 17, 8, 0, 0, 0, time.UTC)
	ack := now.Add(-24 * time.Hour)
	retention := now.AddDate(0, 1, 0)
	opened := now.Add(-time.Hour)
	approved := opened.Add(-time.Minute)
	approver := "00000000-0000-4000-8000-000000000001"
	candidate := &entity.Candidate{ID: "00000000-0000-4000-8000-000000000002", CandidateStatus: "active",
		PrivacyNoticeVersion: "v1", NoticeAcknowledgedAt: &ack, DataRetentionUntil: &retention}
	requisition := &entity.Requisition{ID: "00000000-0000-4000-8000-000000000003", Status: "open",
		ApprovedAt: &approved, ApprovedByPrincipalID: &approver, OpenedAt: &opened}
	application := &entity.Application{CandidateID: candidate.ID, RequisitionID: requisition.ID,
		Status: entity.ApplicationStatusReceived, AppliedAt: now}
	v := &applicationValidator{candidates: &applicationCandidateStub{candidate},
		requisitions: &applicationRequisitionStub{requisition}}
	if err := v.ValidateMutation(t.Context(), "hr_applications", nil, application); err != nil {
		t.Fatalf("valid received application: %v", err)
	}
	changed := *application
	changed.Status = entity.ApplicationStatusScreening
	if err := v.ValidateMutation(t.Context(), "hr_applications", nil, &changed); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("screening through CRUD: %v", err)
	}
	changed = *application
	stage := "screening"
	changed.CurrentStageCode = &stage
	if err := v.ValidateMutation(t.Context(), "hr_applications", nil, &changed); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("stage through CRUD: %v", err)
	}
	changed = *application
	changed.SourcePostingID = &approver
	if err := v.ValidateMutation(t.Context(), "hr_applications", nil, &changed); !errors.Is(err, hrerrors.ErrUnavailable) {
		t.Fatalf("posting without verification: %v", err)
	}
	if err := v.ValidateMutation(t.Context(), "hr_applications", application, &changed); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("update through CRUD: %v", err)
	}
	if err := v.ValidateMutation(t.Context(), "hr_applications", application, nil); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("delete through CRUD: %v", err)
	}
	rejectionStage := "rejected"
	rejectionReason := "POSITION_CANCELLED"
	rejected := *application
	rejected.Status = entity.ApplicationStatusRejected
	rejected.CurrentStageCode = &rejectionStage
	rejected.DispositionReasonCode = &rejectionReason
	rejected.DecisionAt = &now
	rejected.DecisionByPrincipalID = &approver
	if err := v.ValidateMutation(t.Context(), "hr_applications", application, &rejected); err != nil {
		t.Fatalf("reject received application: %v", err)
	}
	invalidRejection := rejected
	invalidRejection.DispositionReasonCode = nil
	if err := v.ValidateMutation(t.Context(), "hr_applications", application, &invalidRejection); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("reject without reason: %v", err)
	}
	invalidRejection = rejected
	invalidRejection.CandidateID = "00000000-0000-4000-8000-000000000009"
	if err := v.ValidateMutation(t.Context(), "hr_applications", application, &invalidRejection); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("change candidate during rejection: %v", err)
	}
	if err := v.ValidateMutation(t.Context(), "hr_applications", &rejected, &rejected); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("edit rejected application: %v", err)
	}
	candidate.DataRetentionUntil = &now
	if err := v.ValidateMutation(t.Context(), "hr_applications", nil, application); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("expired retention: %v", err)
	}
	candidate.DataRetentionUntil = &retention
	requisition.Status = "closed"
	if err := v.ValidateMutation(t.Context(), "hr_applications", nil, application); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("closed requisition: %v", err)
	}
	requisition.Status = "open"
	requisition.OpenedAt = nil
	if err := v.ValidateMutation(t.Context(), "hr_applications", nil, application); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("unopened requisition: %v", err)
	}
}
