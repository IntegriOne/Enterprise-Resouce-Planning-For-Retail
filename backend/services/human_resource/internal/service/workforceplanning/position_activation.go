package service

import (
	"context"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

// ActivatePosition mengesahkan posisi terencana melalui command HRM.
func (s *Service) ActivatePosition(ctx context.Context, positionID string) (*entity.Position, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.position.Command(ctx, positionID, "activate", func(_ context.Context, position *entity.Position, actor operation.Actor, _ time.Time) (*entity.Position, error) {
		if actor.Kind != "human" {
			return nil, hrerrors.ErrForbidden
		}
		if position.PositionStatus != "planned" {
			return nil, hrerrors.ErrWorkflowRequired
		}
		position.PositionStatus = "active"
		return position, nil
	})
}
