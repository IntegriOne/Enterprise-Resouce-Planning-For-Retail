package service

import (
	"context"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

// AssignServiceRequest menugaskan permintaan kepada petugas terverifikasi yang mengambilnya.
func (s *Service) AssignServiceRequest(ctx context.Context, id string) (*entity.ServiceRequest, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.serviceRequest.Command(ctx, id, "assign", func(_ context.Context, request *entity.ServiceRequest, actor operation.Actor, _ time.Time) (*entity.ServiceRequest, error) {
		if actor.Kind != "human" {
			return nil, hrerrors.ErrForbidden
		}
		if request.Status != "open" || request.AssignedPrincipalID != nil {
			return nil, hrerrors.ErrWorkflowRequired
		}
		request.Status = "in_progress"
		request.AssignedPrincipalID = &actor.PrincipalID
		return request, nil
	})
}
