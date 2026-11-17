package service

import (
	"context"
	"strings"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

// ShortlistApplication mencatat keputusan manual recruiter setelah screening.
func (s *Service) ShortlistApplication(ctx context.Context, applicationID, reasonCode string) (*entity.Application, error) {
	if s == nil || operation.RequireDependency(s.stageHistoryWriter) != nil {
		return nil, hrerrors.ErrUnavailable
	}
	reasonCode = strings.TrimSpace(reasonCode)
	if reasonCode == "" || len(reasonCode) > 80 {
		return nil, hrerrors.Invalid("reason_code", "has invalid value")
	}
	return s.application.Command(ctx, applicationID, "advance", func(txCtx context.Context, application *entity.Application, actor operation.Actor, now time.Time) (*entity.Application, error) {
		if actor.Kind != "human" {
			return nil, hrerrors.ErrForbidden
		}
		if application.Status != entity.ApplicationStatusScreening || application.CurrentStageCode == nil ||
			*application.CurrentStageCode != "screening" || application.AssignedRecruiterPrincipalID == nil {
			return nil, hrerrors.ErrWorkflowRequired
		}
		if *application.AssignedRecruiterPrincipalID != actor.PrincipalID {
			return nil, hrerrors.ErrForbidden
		}
		if application.ScreeningScore != nil || application.ScorecardVersion != nil ||
			application.DispositionReasonCode != nil || application.DecisionAt != nil ||
			application.DecisionByPrincipalID != nil || application.HiredWorkerID != nil || now.Before(application.AppliedAt) {
			return nil, hrerrors.ErrWorkflowRequired
		}
		previousStatus := application.Status
		stageCode := "shortlisted"
		application.Status = entity.ApplicationStatusShortlisted
		application.CurrentStageCode = &stageCode
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
