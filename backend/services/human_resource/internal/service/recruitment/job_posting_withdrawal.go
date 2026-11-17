package service

import (
	"context"
	"strings"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

// WithdrawJobPosting menarik posting yang telah dirilis dengan alasan dan actor tercatat.
func (s *Service) WithdrawJobPosting(ctx context.Context, postingID, reasonCode string) (*entity.JobPosting, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	reasonCode = strings.TrimSpace(reasonCode)
	if reasonCode == "" || len(reasonCode) > 80 {
		return nil, hrerrors.Invalid("reason_code", "has invalid value")
	}
	return s.jobPosting.Command(ctx, postingID, "withdraw", func(_ context.Context, posting *entity.JobPosting, actor operation.Actor, now time.Time) (*entity.JobPosting, error) {
		if actor.Kind != "human" {
			return nil, hrerrors.ErrForbidden
		}
		if posting.Status != "published" || posting.PublishedAt == nil || posting.ClosesAt != nil ||
			posting.WithdrawnAt != nil || !now.After(*posting.PublishedAt) {
			return nil, hrerrors.ErrWorkflowRequired
		}
		posting.Status = "withdrawn"
		posting.WithdrawnAt = &now
		posting.WithdrawalReasonCode = &reasonCode
		posting.WithdrawnByPrincipalID = &actor.PrincipalID
		return posting, nil
	})
}
