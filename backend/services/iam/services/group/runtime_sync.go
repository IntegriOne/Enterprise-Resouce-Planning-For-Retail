package group

import (
	"context"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	"github.com/integrione/erp-retail/backend/services/iam/infrastructure/redisstore"
	"github.com/integrione/erp-retail/backend/services/iam/repository/pgstore"
)

func (s *Service) bumpAndRecordGroupSecurityInTx(txCtx context.Context, groupID string) error {
	if s.securityRuntime == nil && s.outboxCoord == nil && s.projector == nil {
		return nil
	}
	var affected []string
	if s.securityRuntime != nil {
		ids, err := s.securityRuntime.BumpPermRevisionForGroup(txCtx, groupID)
		if err != nil {
			return err
		}
		affected = ids
	}
	if s.outboxCoord == nil && s.projector == nil {
		return nil
	}

	payload := redisstore.OutboxPayload{
		AffectedUserIDs: affected,
		InvalidateAuthz: true,
	}
	var evID string
	if s.outboxCoord != nil {
		var err error
		evID, err = s.outboxCoord.RecordInTx(txCtx, entity.OutboxEventGroupPermissionsChanged, "group", groupID, payload)
		if err != nil {
			return err
		}
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
		if s.outboxCoord != nil {
			s.outboxCoord.CompleteAfterCommit(postCtx, evID, projectAffected)
			return
		}
		_ = projectAffected(1)
	})
	return nil
}

func (s *Service) bumpAndRecordUserPermissionsInTx(txCtx context.Context, userID string) error {
	if s.securityRuntime != nil {
		if _, _, err := s.securityRuntime.IncrementUserSecurityCounters(txCtx, userID, false, true); err != nil {
			return err
		}
	}
	if s.outboxCoord == nil && s.projector == nil {
		return nil
	}
	payload := redisstore.OutboxPayload{
		UserID:          userID,
		InvalidateAuthz: true,
	}
	var evID string
	if s.outboxCoord != nil {
		var err error
		evID, err = s.outboxCoord.RecordInTx(txCtx, entity.OutboxEventUserPermissionsChanged, "user", userID, payload)
		if err != nil {
			return err
		}
	}

	pgstore.RegisterAfterCommit(txCtx, func(postCtx context.Context) {
		u, err := s.users.SelectByID(postCtx, userID)
		if err != nil {
			return
		}
		if s.outboxCoord != nil {
			s.outboxCoord.CompleteAfterCommit(postCtx, evID, func(recEpoch int64) error {
				if s.projector == nil {
					return nil
				}
				return s.projector.ProjectUserSecurityChanged(postCtx, u, recEpoch, true)
			})
			return
		}
		if s.projector != nil && s.projector.IsEnabled() {
			_ = s.projector.ProjectUserSecurityChanged(postCtx, u, 1, true)
		}
	})
	return nil
}
