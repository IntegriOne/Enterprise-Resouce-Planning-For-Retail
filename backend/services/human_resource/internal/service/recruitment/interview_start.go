package service

import (
	"context"
	"strings"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

// StartInterview memulai wawancara dan mencatat perpindahan tahap lamaran dalam transaksi yang sama.
func (s *Service) StartInterview(ctx context.Context, id string) (*entity.Interview, error) {
	if s == nil || operation.RequireDependency(s.applicationRecords) != nil ||
		operation.RequireDependency(s.stageHistoryWriter) != nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.interview.Command(ctx, id, "start", func(txCtx context.Context, interview *entity.Interview, actor operation.Actor, now time.Time) (*entity.Interview, error) {
		if actor.Kind != "human" {
			return nil, hrerrors.ErrForbidden
		}
		if (interview.Status != "planned" && interview.Status != "scheduled") ||
			interview.ScheduledStart == nil || interview.ScheduledEnd == nil ||
			now.Before(*interview.ScheduledStart) || !now.Before(*interview.ScheduledEnd) {
			return nil, hrerrors.ErrWorkflowRequired
		}
		application, err := s.applicationRecords.FindByIDForUpdate(txCtx, interview.ApplicationID)
		if err != nil {
			return nil, err
		}
		if application == nil {
			return nil, hrerrors.ErrReferenceConflict
		}
		if application.AssignedRecruiterPrincipalID == nil || !strings.EqualFold(*application.AssignedRecruiterPrincipalID, actor.PrincipalID) {
			return nil, hrerrors.ErrForbidden
		}
		if application.Status != entity.ApplicationStatusShortlisted && application.Status != entity.ApplicationStatusInterview {
			return nil, hrerrors.ErrWorkflowRequired
		}
		if application.Status == entity.ApplicationStatusShortlisted {
			next := *application
			stage := "interview"
			next.Status, next.CurrentStageCode = entity.ApplicationStatusInterview, &stage
			if err := next.Validate(); err != nil {
				return nil, err
			}
			if err := s.application.CheckIntegrity(txCtx, application, &next); err != nil {
				return nil, err
			}
			if err := s.applicationRecords.Update(txCtx, &next); err != nil {
				return nil, err
			}
			history, err := entity.NewApplicationStageHistory(entity.ApplicationStageHistory{
				ApplicationID: application.ID, PreviousStatus: &application.Status, NewStatus: next.Status,
				StageCode: &stage, ChangedAt: now, ChangedByPrincipalID: &actor.PrincipalID,
			})
			if err != nil {
				return nil, err
			}
			if err := s.stageHistoryWriter.Insert(txCtx, history); err != nil {
				return nil, err
			}
		}
		interview.Status = "in_progress"
		return interview, nil
	})
}
