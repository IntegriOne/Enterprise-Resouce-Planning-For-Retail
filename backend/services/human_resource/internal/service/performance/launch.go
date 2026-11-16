package service

import (
	"context"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

// LaunchPerformanceCycle membuka siklus draft dengan pelaku dan waktu dari server.
func (s *Service) LaunchPerformanceCycle(ctx context.Context, cycleID string) (*entity.PerformanceCycle, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.performanceCycle.Command(ctx, cycleID, "launch", func(_ context.Context, cycle *entity.PerformanceCycle, actor operation.Actor, now time.Time) (*entity.PerformanceCycle, error) {
		if actor.Kind != "human" {
			return nil, hrerrors.ErrForbidden
		}
		if cycle.Status != entity.PerformanceCycleStatusDraft || cycle.LaunchedAt != nil || cycle.LaunchedByPrincipalID != nil {
			return nil, hrerrors.ErrWorkflowRequired
		}
		if entity.CompareDates(now, cycle.PeriodEnd) > 0 {
			return nil, hrerrors.ErrReferenceConflict
		}
		cycle.Status = entity.PerformanceCycleStatusOpen
		cycle.LaunchedAt = &now
		cycle.LaunchedByPrincipalID = &actor.PrincipalID
		return cycle, nil
	})
}
