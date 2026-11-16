package service

import (
	"context"
	"strings"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

// ApproveCareerPath mengesahkan draft oleh pelaku lain dalam satu transaksi dan outbox.
func (s *Service) ApproveCareerPath(ctx context.Context, pathID string) (*entity.CareerPath, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.careerPath.Command(ctx, pathID, "approve", func(_ context.Context, path *entity.CareerPath, actor operation.Actor, now time.Time) (*entity.CareerPath, error) {
		if actor.Kind != "human" || strings.EqualFold(path.CreatedByPrincipalID, actor.PrincipalID) {
			return nil, hrerrors.ErrForbidden
		}
		if path.Status != "draft" || path.ApprovedByPrincipalID != nil || path.ApprovedAt != nil {
			return nil, hrerrors.ErrWorkflowRequired
		}
		if path.ValidTo != nil && entity.CompareDates(*path.ValidTo, now) < 0 {
			return nil, hrerrors.ErrReferenceConflict
		}
		path.Status = "approved"
		path.ApprovedByPrincipalID = &actor.PrincipalID
		path.ApprovedAt = &now
		return path, nil
	})
}
