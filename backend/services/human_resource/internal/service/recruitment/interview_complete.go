package service

import (
	"context"
	"strings"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

// CompleteInterview menutup wawancara setelah bukti feedback panel lengkap.
func (s *Service) CompleteInterview(ctx context.Context, id string) (*entity.Interview, error) {
	if s == nil || operation.RequireDependency(s.applicationRecords) != nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.interview.Command(ctx, id, "complete", func(txCtx context.Context, interview *entity.Interview, actor operation.Actor, now time.Time) (*entity.Interview, error) {
		if actor.Kind != "human" {
			return nil, hrerrors.ErrForbidden
		}
		if interview.Status != "in_progress" || interview.ScheduledStart == nil || now.Before(*interview.ScheduledStart) {
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
		if application.Status != entity.ApplicationStatusInterview {
			return nil, hrerrors.ErrWorkflowRequired
		}
		interview.Status = "completed"
		return interview, nil
	})
}
