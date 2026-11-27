package role

import (
	"context"
	"errors"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
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
					if errors.Is(err, iamerrors.ErrUserNotFound) {
						continue
					}
					return err
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
		s.outboxCoord.CompleteAfterCommit(postCtx, evID, func(recEpoch int64) error {
			if s.projector == nil || !s.projector.IsEnabled() {
				return nil
			}
			u, err := s.users.SelectByID(postCtx, userID)
			if err != nil {
				if errors.Is(err, iamerrors.ErrUserNotFound) {
					return nil
				}
				return err
			}
			return s.projector.ProjectUserSecurityChanged(postCtx, u, recEpoch, true)
		})
	})
	return nil
}
