package service

import (
	"context"
	"strings"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

// RejectApplication mencatat keputusan manusia dan riwayat tahap secara atomik.
func (s *Service) RejectApplication(ctx context.Context, applicationID, reasonCode string) (*entity.Application, error) {
	if s == nil || operation.RequireDependency(s.stageHistoryWriter) != nil {
		return nil, hrerrors.ErrUnavailable
	}
	reasonCode = strings.TrimSpace(reasonCode)
	if reasonCode == "" || len(reasonCode) > 80 {
		return nil, hrerrors.Invalid("reason_code", "has invalid value")
	}
	return s.application.Command(ctx, applicationID, "reject", func(txCtx context.Context, application *entity.Application, actor operation.Actor, now time.Time) (*entity.Application, error) {
		if actor.Kind != "human" {
			return nil, hrerrors.ErrForbidden
		}
		if application.Status != entity.ApplicationStatusReceived && application.Status != entity.ApplicationStatusScreening {
			return nil, hrerrors.ErrWorkflowRequired
		}
		if application.Status == entity.ApplicationStatusScreening &&
			(application.AssignedRecruiterPrincipalID == nil || *application.AssignedRecruiterPrincipalID != actor.PrincipalID) {
			return nil, hrerrors.ErrForbidden
		}
		if application.Status == entity.ApplicationStatusReceived &&
			(application.CurrentStageCode != nil || application.AssignedRecruiterPrincipalID != nil) {
			return nil, hrerrors.ErrWorkflowRequired
		}
		if application.DispositionReasonCode != nil || application.DecisionAt != nil ||
			application.DecisionByPrincipalID != nil || application.HiredWorkerID != nil || now.Before(application.AppliedAt) {
			return nil, hrerrors.ErrWorkflowRequired
		}
		previousStatus := application.Status
		stageCode := "rejected"
		application.Status = entity.ApplicationStatusRejected
		application.CurrentStageCode = &stageCode
		application.DispositionReasonCode = &reasonCode
		application.DecisionAt = &now
		application.DecisionByPrincipalID = &actor.PrincipalID
		history := &entity.ApplicationStageHistory{
			ApplicationID: application.ID, PreviousStatus: &previousStatus,
			NewStatus: application.Status, StageCode: &stageCode, ChangedAt: now,
			ChangedByPrincipalID: &actor.PrincipalID, ReasonCode: &reasonCode,
		}
		if err := history.Validate(); err != nil {
			return nil, err
		}
		if err := s.stageHistoryWriter.Insert(txCtx, history); err != nil {
			return nil, err
		}
		return application, nil
	})
}
