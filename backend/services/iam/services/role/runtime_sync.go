package role

import (
	"context"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	"github.com/integrione/erp-retail/backend/services/iam/infrastructure/redisstore"
	"github.com/integrione/erp-retail/backend/services/iam/repository/pgstore"
)

func (s *Service) bumpAndRecordRoleSecurityInTx(txCtx context.Context, roleID string) error {
	affected, err := s.securityRuntime.BumpPermRevisionForRole(txCtx, roleID)
	if err != nil {
		return err
	}
	payload := redisstore.OutboxPayload{
		AffectedUserIDs: affected,
		InvalidateAuthz: true,
	}
	evID, err := s.outboxCoord.RecordInTx(txCtx, entity.OutboxEventRolePermissionsChanged, "role", roleID, payload)
	if err != nil {
		return err
	}

	pgstore.RegisterAfterCommit(txCtx, func(postCtx context.Context) {
		projectAffected := func(recEpoch int64) error {
			if s.projector == nil || !s.projector.IsEnabled() {
				return nil
			}
			for _, uid := range affected {
				u, err := s.users.SelectByID(postCtx, uid)
				if err != nil {
					continue
				}
				if err := s.projector.ProjectUserSecurityChanged(postCtx, u, recEpoch, true); err != nil {
					return err
				}
			}
			return nil
		}
		s.outboxCoord.CompleteAfterCommit(postCtx, evID, projectAffected)
	})
	return nil
}

func (s *Service) bumpAndRecordUserPermissionsInTx(txCtx context.Context, userID string) error {
	if _, _, err := s.securityRuntime.IncrementUserSecurityCounters(txCtx, userID, false, true); err != nil {
		return err
	}
	payload := redisstore.OutboxPayload{
		UserID:          userID,
		InvalidateAuthz: true,
	}
	evID, err := s.outboxCoord.RecordInTx(txCtx, entity.OutboxEventUserPermissionsChanged, "user", userID, payload)
	if err != nil {
		return err
	}

	pgstore.RegisterAfterCommit(txCtx, func(postCtx context.Context) {
		u, err := s.users.SelectByID(postCtx, userID)
		if err != nil {
			return
		}
		s.outboxCoord.CompleteAfterCommit(postCtx, evID, func(recEpoch int64) error {
			if s.projector == nil {
				return nil
			}
			return s.projector.ProjectUserSecurityChanged(postCtx, u, recEpoch, true)
		})
	})
	return nil
}
