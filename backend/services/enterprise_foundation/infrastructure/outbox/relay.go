package outbox

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	domainrepo "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/repository"
	domainservice "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/service"
)

const (
	// DefaultPollInterval adalah interval sweep periodik antrean outbox di PostgreSQL.
	DefaultPollInterval = 5 * time.Second
	// DefaultBatchSize adalah ukuran batch maksimum event yang diklaim dalam satu transaksi SKIP LOCKED.
	DefaultBatchSize = 25
	// DefaultMaxAttempts adalah jumlah total percobaan pengiriman sebelum event ditandai failed (dead-letter).
	DefaultMaxAttempts = 20
	// DefaultPublishTimeout adalah batas waktu per panggilan Publisher eksternal untuk satu event.
	DefaultPublishTimeout = 5 * time.Second
	// DefaultCleanupInterval adalah interval pembersihan berkala untuk arsip event berstatus effective.
	DefaultCleanupInterval = 1 * time.Hour
	// DefaultRetentionPeriod adalah lama simpan arsip event effective sebelum dihapus oleh cleanup.
	DefaultRetentionPeriod = 7 * 24 * time.Hour

	maxBatchesPerDrain = 20
)

// Mode menunjukkan perilaku eksekusi relay outbox Enterprise Foundation.
type Mode string

const (
	// ModeActive menjalankan pengurasan antrean pending, pengiriman ke Publisher, retry backoff, dan dead-letter.
	ModeActive Mode = "active"
	// ModeStandby memelihara siklus hidup dan retensi tanpa memandai event pending sebagai effective saat konsumen belum aktif.
	ModeStandby Mode = "standby"
)

// Config mengatur parameter eksekusi relay asinkron outbox.
type Config struct {
	Standby         bool
	PollInterval    time.Duration
	BatchSize       int
	MaxAttempts     int
	PublishTimeout  time.Duration
	BatchTimeout    time.Duration
	CleanupInterval time.Duration
	RetentionPeriod time.Duration
	Backoff         BackoffPolicy
}

// Dependencies menampung seluruh dependensi yang dibutuhkan oleh Relay.
type Dependencies struct {
	Transactor domainrepo.Transactor
	Repo       domainrepo.OutboxRepository
	Publisher  Publisher
	Clock      domainservice.Clock
	Logger     *slog.Logger
}

// DrainStats merangkum hasil satu putaran pengurasan batch atau siklus drain antrean outbox.
type DrainStats struct {
	Claimed   int
	Published int
	Retried   int
	Failed    int
}

// Stats memuat ringkasan kondisi antrean outbox saat ini untuk observabilitas runtime.
type Stats struct {
	Mode         Mode
	PendingCount int64
	FailedCount  int64
}

// Relay memproses event foundation_outbox_events secara asinkron dengan wakeup pasca-commit, backoff retry, dead-letter, dan retensi.
type Relay struct {
	tx        domainrepo.Transactor
	repo      domainrepo.OutboxRepository
	publisher Publisher
	clock     domainservice.Clock
	logger    *slog.Logger
	cfg       Config
	notifyCh  chan struct{}
	drainMu   sync.Mutex
	runMu     sync.Mutex
}

var _ Notifier = (*Relay)(nil)

// DefaultConfig mengembalikan konfigurasi standar yang aman untuk production dan development.
func DefaultConfig() Config {
	return Config{
		PollInterval:    DefaultPollInterval,
		BatchSize:       DefaultBatchSize,
		MaxAttempts:     DefaultMaxAttempts,
		PublishTimeout:  DefaultPublishTimeout,
		CleanupInterval: DefaultCleanupInterval,
		RetentionPeriod: DefaultRetentionPeriod,
		Backoff:         DefaultBackoffPolicy(),
	}
}

// NewRelay membangun Relay outbox tervalidasi secara fail-closed atas transactor, repositori, dan Publisher (bila bukan mode standby).
func NewRelay(deps Dependencies, cfg Config) (*Relay, error) {
	if deps.Transactor == nil || deps.Repo == nil {
		return nil, ErrRelayNotInitialized
	}
	if !cfg.Standby && deps.Publisher == nil {
		return nil, ErrPublisherUnavailable
	}

	normalized := normalizeConfig(cfg)
	clock := deps.Clock
	if clock == nil {
		clock = domainservice.SystemClock{}
	}
	logger := deps.Logger
	if logger == nil {
		logger = slog.Default()
	}

	return &Relay{
		tx:        deps.Transactor,
		repo:      deps.Repo,
		publisher: deps.Publisher,
		clock:     clock,
		logger:    logger,
		cfg:       normalized,
		notifyCh:  make(chan struct{}, 1),
	}, nil
}

func normalizeConfig(cfg Config) Config {
	if cfg.PollInterval <= 0 {
		cfg.PollInterval = DefaultPollInterval
	}
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = DefaultBatchSize
	} else if cfg.BatchSize > domainrepo.MaxPageLimit {
		cfg.BatchSize = domainrepo.MaxPageLimit
	}
	if cfg.MaxAttempts <= 0 {
		cfg.MaxAttempts = DefaultMaxAttempts
	}
	if cfg.PublishTimeout <= 0 {
		cfg.PublishTimeout = DefaultPublishTimeout
	}
	minBatchTimeout := max(30*time.Second, cfg.PublishTimeout*time.Duration(cfg.BatchSize)+5*time.Second)
	if cfg.BatchTimeout < minBatchTimeout {
		cfg.BatchTimeout = minBatchTimeout
	}
	if cfg.CleanupInterval < 0 {
		cfg.CleanupInterval = 0
	}
	if cfg.RetentionPeriod < 0 {
		cfg.RetentionPeriod = 0
	}
	cfg.Backoff = cfg.Backoff.normalized()
	return cfg
}

// Mode mengembalikan mode operasi relay terkini (active atau standby).
func (r *Relay) Mode() Mode {
	if r == nil || r.cfg.Standby || r.publisher == nil {
		return ModeStandby
	}
	return ModeActive
}

// Notify mengirim sinyal non-blocking agar loop relay segera menguras antrean tanpa menunggu ticker berikutnya.
func (r *Relay) Notify() {
	if r == nil || r.notifyCh == nil {
		return
	}
	select {
	case r.notifyCh <- struct{}{}:
	default:
	}
}

// Run menjalankan worker background relay outbox hingga ctx dibatalkan, lalu berhenti secara bersih tanpa bocor goroutine.
func (r *Relay) Run(ctx context.Context) error {
	if r == nil {
		return ErrRelayNotInitialized
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if !r.runMu.TryLock() {
		return errors.New("outbox relay is already running")
	}
	defer r.runMu.Unlock()

	if err := ctx.Err(); err != nil {
		return nil
	}

	if r.Mode() == ModeStandby {
		return r.runStandby(ctx)
	}

	// Lakukan pemulihan antrean tertunda segera saat startup proses sebelum menunggu timer pertama.
	if _, err := r.DrainPending(ctx); err != nil && ctx.Err() == nil {
		r.logger.Warn("initial outbox recovery drain failed", slog.String("error", err.Error()))
	}

	pollTimer := time.NewTimer(r.cfg.PollInterval)
	defer pollTimer.Stop()

	cleanupEnabled := r.cfg.CleanupInterval > 0 && r.cfg.RetentionPeriod > 0
	var cleanupTimer *time.Timer
	var cleanupCh <-chan time.Time
	if cleanupEnabled {
		cleanupTimer = time.NewTimer(r.cfg.CleanupInterval)
		defer cleanupTimer.Stop()
		cleanupCh = cleanupTimer.C
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-r.notifyCh:
			if !pollTimer.Stop() {
				select {
				case <-pollTimer.C:
				default:
				}
			}
			if _, err := r.DrainPending(ctx); err != nil && ctx.Err() == nil {
				r.logger.Warn("outbox wakeup drain failed", slog.String("error", err.Error()))
			}
			if ctx.Err() != nil {
				return nil
			}
			pollTimer.Reset(r.cfg.PollInterval)
		case <-pollTimer.C:
			if _, err := r.DrainPending(ctx); err != nil && ctx.Err() == nil {
				r.logger.Warn("outbox periodic drain failed", slog.String("error", err.Error()))
			}
			if ctx.Err() != nil {
				return nil
			}
			pollTimer.Reset(r.cfg.PollInterval)
		case <-cleanupCh:
			if deleted, err := r.CleanupEffective(ctx); err != nil && ctx.Err() == nil {
				r.logger.Warn("outbox retention cleanup failed", slog.String("error", err.Error()))
			} else if deleted > 0 {
				r.logger.Info("cleaned expired effective outbox events", slog.Int64("deleted_count", deleted))
			}
			if ctx.Err() != nil {
				return nil
			}
			cleanupTimer.Reset(r.cfg.CleanupInterval)
		}
	}
}

func (r *Relay) runStandby(ctx context.Context) error {
	r.logger.Info("outbox relay running in standby mode; pending events will be preserved until a downstream publisher is configured")

	cleanupEnabled := r.cfg.CleanupInterval > 0 && r.cfg.RetentionPeriod > 0
	var cleanupTimer *time.Timer
	var cleanupCh <-chan time.Time
	if cleanupEnabled {
		cleanupTimer = time.NewTimer(r.cfg.CleanupInterval)
		defer cleanupTimer.Stop()
		cleanupCh = cleanupTimer.C
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-r.notifyCh:
			// Kosongkan sinyal wakeup tanpa mengubah status event pending saat mode standby aktif.
		case <-cleanupCh:
			if deleted, err := r.CleanupEffective(ctx); err != nil && ctx.Err() == nil {
				r.logger.Warn("standby outbox retention cleanup failed", slog.String("error", err.Error()))
			} else if deleted > 0 {
				r.logger.Info("cleaned expired effective outbox events in standby mode", slog.Int64("deleted_count", deleted))
			}
			if ctx.Err() != nil {
				return nil
			}
			cleanupTimer.Reset(r.cfg.CleanupInterval)
		}
	}
}

// DrainPending menguras antrean event outbox yang sudah jatuh tempo secara bertahap hingga habis atau dibatalkan.
func (r *Relay) DrainPending(ctx context.Context) (DrainStats, error) {
	if r == nil {
		return DrainStats{}, ErrRelayNotInitialized
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if r.Mode() == ModeStandby {
		return DrainStats{}, nil
	}

	var total DrainStats
	for i := 0; i < maxBatchesPerDrain; i++ {
		if ctx.Err() != nil {
			return total, ctx.Err()
		}
		batchStats, err := r.DrainBatch(ctx)
		total.Claimed += batchStats.Claimed
		total.Published += batchStats.Published
		total.Retried += batchStats.Retried
		total.Failed += batchStats.Failed
		if err != nil {
			return total, err
		}
		if batchStats.Claimed < r.cfg.BatchSize {
			return total, nil
		}
	}

	// Jadwalkan lanjutkan segera pada iterasi loop berikutnya jika backlog masih penuh setelah batas batch.
	r.Notify()
	return total, nil
}

// DrainBatch mengklaim satu batch event outbox di dalam transaksi FOR UPDATE SKIP LOCKED dan memproses setiap itemnya.
func (r *Relay) DrainBatch(ctx context.Context) (DrainStats, error) {
	if r == nil || r.tx == nil || r.repo == nil {
		return DrainStats{}, ErrRelayNotInitialized
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if r.Mode() == ModeStandby {
		return DrainStats{}, nil
	}
	if r.publisher == nil {
		return DrainStats{}, ErrPublisherUnavailable
	}
	if err := ctx.Err(); err != nil {
		return DrainStats{}, err
	}

	r.drainMu.Lock()
	defer r.drainMu.Unlock()

	var stats DrainStats
	// Pisahkan batas waktu transaksi dari pembatalan shutdown agar update DB untuk event yang sudah terkirim tetap dapat di-commit.
	txRootCtx, cancelTx := context.WithTimeout(context.WithoutCancel(ctx), r.cfg.BatchTimeout)
	defer cancelTx()

	txErr := r.tx.WithinTransaction(txRootCtx, func(txCtx context.Context) error {
		now := r.clock.Now()
		events, err := r.repo.SelectDuePending(txCtx, now, r.cfg.BatchSize)
		if err != nil {
			return err
		}
		stats.Claimed = len(events)
		if len(events) == 0 {
			return nil
		}

		for _, ev := range events {
			if ctx.Err() != nil {
				// Hentikan pemrosesan item tersisa saat shutdown agar status pending dan next_retry_at tetap utuh.
				break
			}
			if ev == nil {
				continue
			}

			pubCtx, cancelPub := context.WithTimeout(ctx, r.cfg.PublishTimeout)
			pubErr := r.publisher.Publish(pubCtx, ev.Clone())
			cancelPub()

			attemptNow := r.clock.Now()
			if pubErr == nil {
				if err := r.repo.UpdateEffective(txCtx, ev.ID, attemptNow); err != nil {
					return err
				}
				stats.Published++
				continue
			}

			// Jangan hukum attempt_count jika pengiriman dibatalkan karena context induk selesai (shutdown atau batas pemanggil).
			if ctx.Err() != nil {
				break
			}

			errMsg := normalizeErrorMessage(pubErr)
			nextAttempt := ev.AttemptCount + 1
			permanent := IsPermanent(pubErr)
			if permanent || nextAttempt >= r.cfg.MaxAttempts {
				if err := r.repo.UpdateFailed(txCtx, ev.ID, errMsg); err != nil {
					return err
				}
				stats.Failed++
				r.logger.Error(
					"outbox event marked failed (dead-letter)",
					slog.String("event_id", ev.ID),
					slog.String("event_type", ev.EventType),
					slog.String("aggregate_type", ev.AggregateType),
					slog.String("aggregate_id", ev.AggregateID),
					slog.Int("attempt_count", nextAttempt),
					slog.Bool("permanent_error", permanent),
					slog.String("error", errMsg),
				)
				continue
			}

			delay := r.cfg.Backoff.NextDelay(nextAttempt)
			nextRetryAt := attemptNow.Add(delay)
			if err := r.repo.UpdateRetry(txCtx, ev.ID, nextAttempt, nextRetryAt, errMsg); err != nil {
				return err
			}
			stats.Retried++
			r.logger.Warn(
				"outbox event delivery failed; scheduled retry",
				slog.String("event_id", ev.ID),
				slog.String("event_type", ev.EventType),
				slog.Int("attempt_count", nextAttempt),
				slog.Duration("retry_delay", delay),
				slog.String("error", errMsg),
			)
		}

		return nil
	})
	if txErr != nil {
		return DrainStats{}, txErr
	}
	return stats, nil
}

// CleanupEffective menghapus arsip event berstatus effective yang lebih tua dari periode retensi.
func (r *Relay) CleanupEffective(ctx context.Context) (int64, error) {
	if r == nil || r.repo == nil {
		return 0, ErrRelayNotInitialized
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if r.cfg.RetentionPeriod <= 0 {
		return 0, nil
	}
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	cutoff := r.clock.Now().Add(-r.cfg.RetentionPeriod)
	return r.repo.DeleteEffectiveBefore(ctx, cutoff)
}

// Stats mengembalikan snapshot jumlah antrean pending dan failed beserta mode relay.
func (r *Relay) Stats(ctx context.Context) (Stats, error) {
	if r == nil || r.repo == nil {
		return Stats{}, ErrRelayNotInitialized
	}
	if ctx == nil {
		ctx = context.Background()
	}

	pending, err := r.repo.CountPending(ctx)
	if err != nil {
		return Stats{}, err
	}
	failed, err := r.repo.CountFailed(ctx)
	if err != nil {
		return Stats{}, err
	}
	return Stats{
		Mode:         r.Mode(),
		PendingCount: pending,
		FailedCount:  failed,
	}, nil
}
