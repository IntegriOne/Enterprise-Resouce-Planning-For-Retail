package redisstore

import (
	"context"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
)

// VerifyDisabledUserProjection memastikan Redis menolak sesi pada epoch autentikasi terkini.
func (p *Projector) VerifyDisabledUserProjection(ctx context.Context, user *entity.User) error {
	if !p.IsEnabled() || user == nil {
		return iamerrors.ErrUnavailable
	}
	record, err := p.store.ReadUserSecurity(ctx, user.ID)
	if err != nil {
		return iamerrors.Wrap(err, iamerrors.KindUnavailable, iamerrors.ErrUnavailable.Code, "read disabled user projection failed")
	}
	if record.Status != string(entity.UserStatusDisabled) || record.AuthEpoch < user.AuthEpoch {
		return iamerrors.ErrUnavailable
	}
	return nil
}
