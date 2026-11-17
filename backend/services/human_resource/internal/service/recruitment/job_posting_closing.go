package service

import (
	"context"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

// CloseJobPosting menutup posting yang sudah dirilis dan mencatat waktu dari server.
func (s *Service) CloseJobPosting(ctx context.Context, postingID string) (*entity.JobPosting, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.jobPosting.Command(ctx, postingID, "close", func(_ context.Context, posting *entity.JobPosting, actor operation.Actor, now time.Time) (*entity.JobPosting, error) {
		if actor.Kind != "human" {
			return nil, hrerrors.ErrForbidden
		}
		if posting.Status != "published" || posting.PublishedAt == nil || posting.ClosesAt != nil ||
			!now.After(*posting.PublishedAt) {
			return nil, hrerrors.ErrWorkflowRequired
		}
		posting.Status = "closed"
		posting.ClosesAt = &now
		return posting, nil
	})
}
