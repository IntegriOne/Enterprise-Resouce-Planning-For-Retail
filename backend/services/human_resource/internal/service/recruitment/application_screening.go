package service

import (
	"context"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

// StartApplicationScreening mencatat perpindahan lamaran ke tahap screening secara atomik.
func (s *Service) StartApplicationScreening(ctx context.Context, applicationID string) (*entity.Application, error) {
	if s == nil || operation.RequireDependency(s.stageHistoryWriter) != nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.application.Command(ctx, applicationID, "start_screening", func(txCtx context.Context, application *entity.Application, actor operation.Actor, now time.Time) (*entity.Application, error) {
		if actor.Kind != "human" {
			return nil, hrerrors.ErrForbidden
		}
		if application.Status != entity.ApplicationStatusReceived || application.CurrentStageCode != nil ||
			application.AssignedRecruiterPrincipalID != nil || application.ScreeningScore != nil ||
			application.ScorecardVersion != nil || application.DispositionReasonCode != nil ||
			application.DecisionAt != nil || application.DecisionByPrincipalID != nil || application.HiredWorkerID != nil {
			return nil, hrerrors.ErrWorkflowRequired
		}
		previousStatus := application.Status
		stageCode := "screening"
		reasonCode := "SCREENING_STARTED"
		application.Status = entity.ApplicationStatusScreening
		application.CurrentStageCode = &stageCode
		application.AssignedRecruiterPrincipalID = &actor.PrincipalID
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
