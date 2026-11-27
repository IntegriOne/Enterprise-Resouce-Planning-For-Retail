package service

import (
	"context"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

func authorizeTransfer(ctx context.Context, policy operation.Policy, access operation.Access) (operation.Actor, error) {
	actor, err := policy.Authorize(ctx, access)
	if err != nil {
		return operation.Actor{}, err
	}
	if actor.Kind != "human" || entity.ValidateUUID("principal_id", actor.PrincipalID) != nil ||
		entity.ValidateUUID("session_id", actor.SessionID) != nil || actor.AuthEpoch < 0 || actor.PermissionRevision < 0 {
		return operation.Actor{}, hrerrors.ErrForbidden
	}
	return actor, nil
}
