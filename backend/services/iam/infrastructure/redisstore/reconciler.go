package redisstore

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"runtime/debug"
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
	r.effectiveLogger().Info("iam redis security gate opened", slog.Int64("recovery_epoch", recEpoch))
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
	if ctx == nil {
		ctx = context.Background()
	}
	log := r.effectiveLogger()
	defer func() {
		if rec := recover(); rec != nil {
			log.Error("iam outbox reconciler loop panic recovered",
				slog.Any("panic", rec),
				slog.String("stack", string(debug.Stack())),
			)
		}
	}()

	interval := r.interval
	if interval <= 0 {
		interval = defaultReconcileInterval
	}
	ticker := time.NewTicker(interval)
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
	log := r.effectiveLogger()
	defer func() {
		if rec := recover(); rec != nil {
			log.Error("iam outbox reconciler panic recovered",
				slog.Any("panic", rec),
				slog.String("stack", string(debug.Stack())),
			)
		}
	}()

	stepCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := r.DrainPending(stepCtx); err != nil {
		log.Warn("iam outbox reconcile cycle failed", slog.String("error", err.Error()))
	}
}

func (r *Reconciler) drainPendingLoop(ctx context.Context, recoveryEpoch int64, includeScheduled bool) error {
	clock := r.effectiveClock()
	log := r.effectiveLogger()
	for {
		now := clock.Now()
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
			if ev == nil {
				continue
			}
			if err := r.applyEvent(ctx, ev, recoveryEpoch); err != nil {
				nextAttempt := ev.AttemptCount + 1
				var markErr error
				if nextAttempt >= maxOutboxAttempts {
					if mErr := r.runtimeRepo.MarkFailed(ctx, ev.ID, err.Error()); mErr != nil {
						markErr = fmt.Errorf("mark outbox event %s failed: %w", ev.ID, mErr)
						log.Warn("failed to mark outbox event failed", slog.String("event_id", ev.ID), slog.String("error", mErr.Error()))
					}
				} else {
					backoff := computeBackoff(nextAttempt)
					if mErr := r.runtimeRepo.MarkRetry(ctx, ev.ID, nextAttempt, now.Add(backoff), err.Error()); mErr != nil {
						markErr = fmt.Errorf("mark outbox event %s retry: %w", ev.ID, mErr)
						log.Warn("failed to schedule outbox event retry", slog.String("event_id", ev.ID), slog.String("error", mErr.Error()))
					}
				}
				batchErr = errors.Join(batchErr, err, markErr)
				continue
			}
			if err := r.runtimeRepo.MarkEffective(ctx, ev.ID, clock.Now()); err != nil {
				batchErr = errors.Join(batchErr, err)
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
	now := r.effectiveClock().Now()
	sessions, err := r.runtimeRepo.SelectActiveSessionsForRecovery(ctx, now, recoverySessionBatchCap)
	if err != nil {
		return err
	}
	if len(sessions) >= recoverySessionBatchCap {
		r.effectiveLogger().Warn("active session recovery reached batch cap; older sessions truncated",
			slog.Int("cap", recoverySessionBatchCap),
			slog.Int("loaded", len(sessions)),
		)
	}
	userCache := make(map[string]*entity.User, len(sessions))
	for _, sess := range sessions {
		if sess == nil {
			continue
		}
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
		if err := json.Unmarshal(ev.Payload, &payload); err != nil {
			return fmt.Errorf("decode iam outbox payload (%s): %w", ev.EventType, err)
		}
	}

	switch ev.EventType {
	case entity.OutboxEventSessionIssued, entity.OutboxEventSessionUpdated:
		if r.sessions == nil || r.users == nil {
			return iamerrors.New(iamerrors.KindUnavailable, iamerrors.Code("iam.outbox.repo_unavailable"), "session or user repository is unavailable")
		}
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
		digests, err := decodeDigestsHex(payload.TokenDigestsHex)
		if err != nil {
			return fmt.Errorf("decode session revoked token digests: %w", err)
		}
		var u *entity.User
		if payload.UserID != "" {
			if r.users == nil {
				return iamerrors.New(iamerrors.KindUnavailable, iamerrors.Code("iam.outbox.repo_unavailable"), "user repository is unavailable")
			}
			loaded, err := r.users.SelectByID(ctx, payload.UserID)
			if err != nil && !errors.Is(err, iamerrors.ErrUserNotFound) {
				return err
			}
			u = loaded
		}
		if len(digests) == 0 && payload.SessionID != "" {
			if r.sessions == nil {
				return iamerrors.New(iamerrors.KindUnavailable, iamerrors.Code("iam.outbox.repo_unavailable"), "session repository is unavailable")
			}
			sess, err := r.sessions.SelectByID(ctx, payload.SessionID)
			if err != nil && !errors.Is(err, iamerrors.ErrSessionNotFound) {
				return err
			}
			if sess != nil {
				digests = [][]byte{sess.TokenDigest()}
				if u == nil && r.users != nil {
					loaded, userErr := r.users.SelectByID(ctx, sess.UserID)
					if userErr != nil && !errors.Is(userErr, iamerrors.ErrUserNotFound) {
						return userErr
					}
					u = loaded
				}
			}
		}
		return r.projector.ProjectUserSessionsRevoked(ctx, u, digests, recoveryEpoch)

	case entity.OutboxEventUserSessionsRevoked:
		if r.users == nil {
			return iamerrors.New(iamerrors.KindUnavailable, iamerrors.Code("iam.outbox.repo_unavailable"), "user repository is unavailable")
		}
		u, err := r.users.SelectByID(ctx, ev.AggregateID)
		if err != nil && !errors.Is(err, iamerrors.ErrUserNotFound) {
			return err
		}
		digests, err := decodeDigestsHex(payload.TokenDigestsHex)
		if err != nil {
			return fmt.Errorf("decode user sessions revoked digests: %w", err)
		}
		return r.projector.ProjectUserSessionsRevoked(ctx, u, digests, recoveryEpoch)

	case entity.OutboxEventUserSecurityChanged, entity.OutboxEventUserPermissionsChanged:
		if r.users == nil {
			return iamerrors.New(iamerrors.KindUnavailable, iamerrors.Code("iam.outbox.repo_unavailable"), "user repository is unavailable")
		}
		u, err := r.users.SelectByID(ctx, ev.AggregateID)
		if err != nil {
			if errors.Is(err, iamerrors.ErrUserNotFound) {
				return nil
			}
			return err
		}
		return r.projector.ProjectUserSecurityChanged(ctx, u, recoveryEpoch, payload.InvalidateAuthz || ev.EventType == entity.OutboxEventUserPermissionsChanged)

	case entity.OutboxEventRolePermissionsChanged, entity.OutboxEventGroupPermissionsChanged:
		if r.users == nil {
			return iamerrors.New(iamerrors.KindUnavailable, iamerrors.Code("iam.outbox.repo_unavailable"), "user repository is unavailable")
		}
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
		return fmt.Errorf("unsupported iam outbox event type %q", ev.EventType)
	}
}

func (r *Reconciler) effectiveLogger() *slog.Logger {
	if r != nil && r.logger != nil {
		return r.logger
	}
	return slog.Default()
}

func (r *Reconciler) effectiveClock() domainservice.Clock {
	if r != nil && r.clock != nil {
		return r.clock
	}
	return domainservice.SystemClock{}
}

func decodeDigestsHex(items []string) ([][]byte, error) {
	if len(items) == 0 {
		return nil, nil
	}
	out := make([][]byte, 0, len(items))
	for _, s := range items {
		b, err := hex.DecodeString(s)
		if err != nil {
			return nil, fmt.Errorf("invalid hex digest: %w", err)
		}
		if len(b) == 0 {
			return nil, errors.New("empty hex digest entry")
		}
		out = append(out, b)
	}
	return out, nil
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
