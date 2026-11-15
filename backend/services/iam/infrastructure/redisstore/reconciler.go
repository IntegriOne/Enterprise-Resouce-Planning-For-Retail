package redisstore

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/iam/domain/repository"
	domainservice "github.com/integrione/erp-retail/backend/services/iam/domain/service"
)

const (
	defaultReconcileInterval = 2 * time.Second
	maxOutboxAttempts        = 20
	reconcileBatchSize       = 100
	recoverySessionBatchCap  = 2000
)

// Reconciler menjaga sinkronisasi PG -> Redis, memulihkan sesi aktif saat Redis kosong/restart, dan membuka gate hanya setelah konsisten.
type Reconciler struct {
	runtimeRepo domainrepo.SecurityRuntimeRepository
	users       domainrepo.UserRepository
	sessions    domainrepo.SessionRepository
	projector   *Projector
	coord       *Coordinator
	clock       domainservice.Clock
	logger      *slog.Logger
	interval    time.Duration

	mu sync.Mutex
}

// NewReconciler membentuk worker pemulihan gate keamanan dan antrean outbox IAM.
func NewReconciler(
	runtimeRepo domainrepo.SecurityRuntimeRepository,
	users domainrepo.UserRepository,
	sessions domainrepo.SessionRepository,
	projector *Projector,
	coord *Coordinator,
	clock domainservice.Clock,
	logger *slog.Logger,
) *Reconciler {
	if clock == nil {
		clock = domainservice.SystemClock{}
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &Reconciler{
		runtimeRepo: runtimeRepo,
		users:       users,
		sessions:    sessions,
		projector:   projector,
		coord:       coord,
		clock:       clock,
		logger:      logger,
		interval:    defaultReconcileInterval,
	}
}

// BootstrapGateAndDrain menutup gate selama pemulihan, menguras antrean pending & sesi aktif ke Redis, lalu membuka gate.
func (r *Reconciler) BootstrapGateAndDrain(ctx context.Context) error {
	if r == nil || r.projector == nil || !r.projector.IsEnabled() || r.runtimeRepo == nil {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.bootstrapGateAndDrainLocked(ctx)
}

func (r *Reconciler) bootstrapGateAndDrainLocked(ctx context.Context) error {
	recEpoch, err := r.runtimeRepo.IncrementRecoveryEpoch(ctx)
	if err != nil {
		return err
	}

	if _, err := r.projector.SyncGate(ctx, false, recEpoch, "iam_reconcile_start"); err != nil {
		return err
	}

	if err := r.drainPendingLoop(ctx, recEpoch, true); err != nil {
		return err
	}
	if err := r.hydrateActiveSessions(ctx, recEpoch); err != nil {
		return err
	}

	pendingCount, err := r.runtimeRepo.CountPending(ctx)
	if err != nil {
		return err
	}
	if pendingCount > 0 {
		return errors.New("cannot open security gate while outbox still has pending events")
	}

	failedCount, err := r.runtimeRepo.CountFailed(ctx)
	if err != nil {
		return err
	}
	if failedCount > 0 {
		return errors.New("cannot open security gate while outbox has failed security events")
	}

	if _, err := r.projector.SyncGate(ctx, true, recEpoch, "iam_ready"); err != nil {
		return err
	}
	r.logger.Info("iam redis security gate opened", slog.Int64("recovery_epoch", recEpoch))
	return nil
}

// DrainPending memproses satu siklus event outbox pending yang jatuh tempo.
func (r *Reconciler) DrainPending(ctx context.Context) error {
	if r == nil || r.projector == nil || !r.projector.IsEnabled() || r.runtimeRepo == nil {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	gate, gateErr := r.projector.GetGate(ctx)
	recEpoch, err := r.runtimeRepo.GetRecoveryEpoch(ctx)
	if err != nil {
		return err
	}
	// Jika gate di Redis hilang atau belum terbuka (mis. Redis baru restart), lakukan pemulihan penuh dengan kunci tetap terjaga.
	if gateErr != nil || gate == nil || !gate.Open || gate.RecoveryEpoch < recEpoch {
		return r.bootstrapGateAndDrainLocked(ctx)
	}

	return r.drainPendingLoop(ctx, recEpoch, false)
}

// Run menjalankan loop rekonsiliasi outbox dan pengawasan gate Redis hingga ctx dibatalkan.
func (r *Reconciler) Run(ctx context.Context) {
	if r == nil || r.projector == nil || !r.projector.IsEnabled() {
		return
	}

	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	var notifyCh <-chan struct{}
	if r.coord != nil {
		notifyCh = r.coord.NotifyChannel()
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-notifyCh:
			r.stepOnce(ctx)
		case <-ticker.C:
			r.stepOnce(ctx)
		}
	}
}

func (r *Reconciler) stepOnce(ctx context.Context) {
	defer func() {
		if rec := recover(); rec != nil {
			r.logger.Error("iam outbox reconciler panic recovered", slog.Any("panic", rec))
		}
	}()

	stepCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := r.DrainPending(stepCtx); err != nil {
		r.logger.Warn("iam outbox reconcile cycle failed", slog.String("error", err.Error()))
	}
}

func (r *Reconciler) drainPendingLoop(ctx context.Context, recoveryEpoch int64, includeScheduled bool) error {
	for {
		now := r.clock.Now()
		dueCutoff := now
		if includeScheduled {
			// Saat bootstrap/pemulihan penuh, kuras seluruh antrean pending tanpa menunggu sisa backoff.
			dueCutoff = now.Add(3650 * 24 * time.Hour)
		}
		events, err := r.runtimeRepo.SelectDuePending(ctx, dueCutoff, reconcileBatchSize)
		if err != nil {
			return err
		}
		if len(events) == 0 {
			return nil
		}

		var batchErr error
		for _, ev := range events {
			if err := r.applyEvent(ctx, ev, recoveryEpoch); err != nil {
				nextAttempt := ev.AttemptCount + 1
				if nextAttempt >= maxOutboxAttempts {
					_ = r.runtimeRepo.MarkFailed(ctx, ev.ID, err.Error())
				} else {
					backoff := computeBackoff(nextAttempt)
					_ = r.runtimeRepo.MarkRetry(ctx, ev.ID, nextAttempt, now.Add(backoff), err.Error())
				}
				if batchErr == nil {
					batchErr = err
				}
				continue
			}
			if err := r.runtimeRepo.MarkEffective(ctx, ev.ID, r.clock.Now()); err != nil && batchErr == nil {
				batchErr = err
			}
		}
		if batchErr != nil {
			return batchErr
		}
		if len(events) < reconcileBatchSize {
			return nil
		}
	}
}

func (r *Reconciler) hydrateActiveSessions(ctx context.Context, recoveryEpoch int64) error {
	if r.sessions == nil || r.users == nil {
		return nil
	}
	now := r.clock.Now()
	sessions, err := r.runtimeRepo.SelectActiveSessionsForRecovery(ctx, now, recoverySessionBatchCap)
	if err != nil {
		return err
	}
	if len(sessions) >= recoverySessionBatchCap {
		r.logger.Warn("active session recovery reached batch cap; older sessions truncated",
			slog.Int("cap", recoverySessionBatchCap),
			slog.Int("loaded", len(sessions)),
		)
	}
	userCache := make(map[string]*entity.User, len(sessions))
	for _, sess := range sessions {
		u, ok := userCache[sess.UserID]
		if !ok {
			loaded, loadErr := r.users.SelectByID(ctx, sess.UserID)
			if loadErr != nil {
				if errors.Is(loadErr, iamerrors.ErrUserNotFound) {
					continue
				}
				return loadErr
			}
			u = loaded
			userCache[sess.UserID] = u
		}
		if u.CanAuthenticate() != nil {
			if err := r.projector.ProjectSessionRevoked(ctx, u, sess.TokenDigest(), recoveryEpoch); err != nil {
				return err
			}
			continue
		}
		if err := r.projector.ProjectSessionIssued(ctx, u, sess, recoveryEpoch); err != nil {
			return err
		}
	}
	return nil
}

func (r *Reconciler) applyEvent(ctx context.Context, ev *entity.OutboxEvent, recoveryEpoch int64) error {
	var payload OutboxPayload
	if len(ev.Payload) > 0 {
		_ = json.Unmarshal(ev.Payload, &payload)
	}

	switch ev.EventType {
	case entity.OutboxEventSessionIssued, entity.OutboxEventSessionUpdated:
		sessID := ev.AggregateID
		if payload.SessionID != "" {
			sessID = payload.SessionID
		}
		sess, err := r.sessions.SelectByID(ctx, sessID)
		if err != nil {
			if errors.Is(err, iamerrors.ErrSessionNotFound) {
				return nil
			}
			return err
		}
		u, err := r.users.SelectByID(ctx, sess.UserID)
		if err != nil {
			return err
		}
		if sess.EndedAt != nil {
			return r.projector.ProjectSessionRevoked(ctx, u, sess.TokenDigest(), recoveryEpoch)
		}
		return r.projector.ProjectSessionUpdated(ctx, u, sess, recoveryEpoch)

	case entity.OutboxEventSessionRevoked:
		digests := decodeDigestsHex(payload.TokenDigestsHex)
		var u *entity.User
		if payload.UserID != "" {
			if loaded, err := r.users.SelectByID(ctx, payload.UserID); err == nil {
				u = loaded
			}
		}
		if len(digests) == 0 && payload.SessionID != "" {
			if sess, err := r.sessions.SelectByID(ctx, payload.SessionID); err == nil {
				digests = [][]byte{sess.TokenDigest()}
				if u == nil {
					u, _ = r.users.SelectByID(ctx, sess.UserID)
				}
			}
		}
		return r.projector.ProjectUserSessionsRevoked(ctx, u, digests, recoveryEpoch)

	case entity.OutboxEventUserSessionsRevoked:
		u, err := r.users.SelectByID(ctx, ev.AggregateID)
		if err != nil && !errors.Is(err, iamerrors.ErrUserNotFound) {
			return err
		}
		digests := decodeDigestsHex(payload.TokenDigestsHex)
		return r.projector.ProjectUserSessionsRevoked(ctx, u, digests, recoveryEpoch)

	case entity.OutboxEventUserSecurityChanged, entity.OutboxEventUserPermissionsChanged:
		u, err := r.users.SelectByID(ctx, ev.AggregateID)
		if err != nil {
			if errors.Is(err, iamerrors.ErrUserNotFound) {
				return nil
			}
			return err
		}
		return r.projector.ProjectUserSecurityChanged(ctx, u, recoveryEpoch, payload.InvalidateAuthz || ev.EventType == entity.OutboxEventUserPermissionsChanged)

	case entity.OutboxEventRolePermissionsChanged, entity.OutboxEventGroupPermissionsChanged:
		for _, uid := range payload.AffectedUserIDs {
			u, err := r.users.SelectByID(ctx, uid)
			if err != nil {
				if errors.Is(err, iamerrors.ErrUserNotFound) {
					continue
				}
				return err
			}
			if err := r.projector.ProjectUserSecurityChanged(ctx, u, recoveryEpoch, true); err != nil {
				return err
			}
		}
		return nil

	default:
		return nil
	}
}

func decodeDigestsHex(items []string) [][]byte {
	if len(items) == 0 {
		return nil
	}
	out := make([][]byte, 0, len(items))
	for _, s := range items {
		b, err := hex.DecodeString(s)
		if err == nil && len(b) > 0 {
			out = append(out, b)
		}
	}
	return out
}

func computeBackoff(attempt int) time.Duration {
	if attempt < 1 {
		return time.Second
	}
	sec := 1 << min(attempt-1, 5)
	if sec > 30 {
		sec = 30
	}
	return time.Duration(sec) * time.Second
}
