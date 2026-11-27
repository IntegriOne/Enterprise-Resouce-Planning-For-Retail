package service

import (
	"context"
	"strings"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

// CancelInterview membebaskan jadwal tanpa menghapus riwayat wawancara dan panelis.
func (s *Service) CancelInterview(ctx context.Context, id string) (*entity.Interview, error) {
	if s == nil || operation.RequireDependency(s.applicationRecords) != nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.interview.Command(ctx, id, "cancel", func(txCtx context.Context, interview *entity.Interview, actor operation.Actor, _ time.Time) (*entity.Interview, error) {
		if actor.Kind != "human" {
			return nil, hrerrors.ErrForbidden
		}
		if interview.Status != "planned" && interview.Status != "scheduled" {
			return nil, hrerrors.ErrWorkflowRequired
		}
		application, err := s.applicationRecords.FindByIDForUpdate(txCtx, interview.ApplicationID)
		if err != nil {
			return nil, err
		}
		if application == nil {
			return nil, hrerrors.ErrReferenceConflict
		}
		responsible := application.AssignedRecruiterPrincipalID
		if responsible == nil {
			responsible = interview.CreatedByPrincipalID
		}
		if responsible == nil || !strings.EqualFold(*responsible, actor.PrincipalID) {
			return nil, hrerrors.ErrForbidden
		}
		interview.Status = "cancelled"
		return interview, nil
	})
}
