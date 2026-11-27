package middleware

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"

	"github.com/integrione/erp-retail/backend/shared/authcontext"
	"github.com/integrione/erp-retail/backend/shared/authsession"
	"github.com/integrione/erp-retail/backend/shared/authz"
)

// SnapshotRefresher memuat AccessSnapshot terbaru secara otoritatif dari layanan IAM gRPC.
type SnapshotRefresher interface {
	RefreshUserAccessSnapshot(ctx context.Context, userID string, caller authcontext.CallerContext) (*authz.AccessSnapshot, error)
}

// SnapshotLoader menggabungkan validasi kesegaran snapshot Redis dengan deduplikasi panggilan singleflight ke IAM.
type SnapshotLoader struct {
	refresher   SnapshotRefresher
	iamTimeout  time.Duration
	maxFreshAge time.Duration
	sfGroup     singleflight.Group

	mu   sync.Mutex
	nowF func() time.Time
}

// NewSnapshotLoader membentuk loader AccessSnapshot dengan deduplikasi singleflight per user_id.
func NewSnapshotLoader(refresher SnapshotRefresher, iamTimeout, maxFreshAge time.Duration) *SnapshotLoader {
	if iamTimeout <= 0 {
		iamTimeout = 5 * time.Second
	}
	if maxFreshAge <= 0 || maxFreshAge > authz.DefaultSnapshotMaxAge {
		maxFreshAge = authz.DefaultSnapshotMaxAge
	}
	return &SnapshotLoader{
		refresher:   refresher,
		iamTimeout:  iamTimeout,
		maxFreshAge: maxFreshAge,
		nowF:        time.Now,
	}
}

// EnsureFresh memeriksa kesegaran & keselarasan epoch snapshot dari Redis; bila basi atau hilang, menyegarkan via singleflight ke IAM.
func (l *SnapshotLoader) EnsureFresh(
	ctx context.Context,
	sessCtx *authsession.SessionValidationContext,
	caller authcontext.CallerContext,
) (*authz.AccessSnapshot, error) {
	if l == nil {
		return nil, ErrUnavailable("snapshot loader unavailable", errors.New("nil snapshot loader"))
	}
	if sessCtx == nil || sessCtx.Token == nil {
		return nil, ErrUnauthenticated("authentication required", nil)
	}

	now := l.now()
	if isSnapshotUsable(sessCtx.Snapshot, sessCtx.Token, sessCtx.Security, l.maxFreshAge, now) {
		return sessCtx.Snapshot, nil
	}

	if l.refresher == nil {
		return nil, ErrUnavailable("authorization snapshot refresher unavailable", errors.New("nil refresher"))
	}

	userID := sessCtx.Token.UserID
	// Lepaskan pembatalan konteks pemanggil pertama agar waiter singleflight lain tidak ikut gagal bila satu klien memutus koneksi.
	detachedParent := context.WithoutCancel(ctx)
	resAny, err, _ := l.sfGroup.Do("refresh:"+userID, func() (res any, callErr error) {
		defer func() {
			if recovered := recover(); recovered != nil {
				res = nil
				callErr = ErrUnavailable("identity service unavailable", fmt.Errorf("recovered snapshot refresh panic: %v", recovered))
			}
		}()
		rpcCtx, cancel := context.WithTimeout(detachedParent, l.iamTimeout)
		defer cancel()
		return l.refresher.RefreshUserAccessSnapshot(rpcCtx, userID, caller)
	})
	if err != nil {
		if _, ok := AsGatewayError(err); ok {
			return nil, err
		}
		return nil, ErrUnavailable("failed to refresh authorization snapshot", err)
	}

	snap, ok := resAny.(*authz.AccessSnapshot)
	if !ok || snap == nil {
		return nil, ErrUnavailable("empty authorization snapshot from IAM", nil)
	}
	if !isSnapshotUsable(snap, sessCtx.Token, sessCtx.Security, l.maxFreshAge, l.now()) {
		if snap.Status != authz.StatusActive {
			return nil, ErrForbidden("user account is not active", nil)
		}
		return nil, ErrUnavailable("refreshed authorization snapshot is stale or out of sync", nil)
	}
	return snap, nil
}

func (l *SnapshotLoader) now() time.Time {
	l.mu.Lock()
	f := l.nowF
	l.mu.Unlock()
	if f == nil {
		return time.Now()
	}
	return f()
}

func isSnapshotUsable(
	snap *authz.AccessSnapshot,
	tok *authsession.SessionTokenRecord,
	sec *authsession.UserSecurityRecord,
	maxFresh time.Duration,
	now time.Time,
) bool {
	if snap == nil || tok == nil {
		return false
	}
	if snap.PrincipalID != tok.UserID || !snap.IsFresh(now) {
		return false
	}
	if maxFresh > 0 && now.Unix()-snap.SyncedAtUnix > int64(maxFresh.Seconds()) {
		return false
	}
	if snap.AuthEpoch < tok.AuthEpoch {
		return false
	}
	if sec != nil {
		if snap.PermRevision < sec.PermRevision || snap.AuthEpoch < sec.AuthEpoch || snap.RecoveryEpoch < sec.RecoveryEpoch {
			return false
		}
	}
	return true
}
