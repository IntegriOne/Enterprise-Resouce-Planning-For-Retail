package outbox

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
)

// DeliveryStore mengubah status event hanya dalam transaksi pemrosesan yang menguncinya.
type DeliveryStore interface {
	SelectDue(context.Context, time.Time, int, int) ([]entity.OutboxEvent, error)
	MarkPublished(context.Context, string, time.Time, int32) error
	MarkRetry(context.Context, string, time.Time, int32, string) error
}

// Transactor adalah batas transaksi lokal untuk klaim dan pembaruan outbox.
type Transactor interface {
	WithinTransaction(context.Context, func(context.Context) error) error
}

// Publisher mengirim satu event dan wajib idempoten terhadap event ID.
type Publisher interface {
	Publish(context.Context, *entity.OutboxEvent) error
}

// Config membatasi polling, durasi jaringan dan percobaan pengiriman.
type Config struct {
	Standby        bool
	PollInterval   time.Duration
	PublishTimeout time.Duration
	MaxAttempts    int
}

// Relay mengirim event secara at-least-once; event published tetap tersimpan untuk versi agregat.
type Relay struct {
	tx        Transactor
	store     DeliveryStore
	publisher Publisher
	cfg       Config
	logger    *slog.Logger
	notify    chan struct{}
}

// NewRelay menolak mode aktif tanpa publisher dan dependency database.
func NewRelay(tx Transactor, store DeliveryStore, publisher Publisher, cfg Config, logger *slog.Logger) (*Relay, error) {
	if tx == nil || store == nil || (!cfg.Standby && publisher == nil) {
		return nil, hrerrors.ErrUnavailable
	}
	if cfg.PollInterval <= 0 {
		cfg.PollInterval = 5 * time.Second
	}
	if cfg.PublishTimeout <= 0 {
		cfg.PublishTimeout = 5 * time.Second
	}
	if cfg.MaxAttempts <= 0 {
		cfg.MaxAttempts = 20
	}
	if cfg.MaxAttempts > 1000 || cfg.PublishTimeout > time.Minute || cfg.PollInterval > time.Hour {
		return nil, hrerrors.ErrInvalidInput
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &Relay{tx: tx, store: store, publisher: publisher, cfg: cfg, logger: logger, notify: make(chan struct{}, 1)}, nil
}

// Notify membangunkan worker tanpa memblokir transaksi penulis event.
func (r *Relay) Notify() {
	if r == nil {
		return
	}
	select {
	case r.notify <- struct{}{}:
	default:
	}
}

// Run memproses event jatuh tempo dan berhenti bersih saat context dibatalkan.
func (r *Relay) Run(ctx context.Context) error {
	if r == nil || ctx == nil {
		return hrerrors.ErrInvalidInput
	}
	if r.cfg.Standby {
		r.logger.Info("HRM outbox relay standby; pending events are preserved")
		<-ctx.Done()
		return nil
	}
	if err := r.Drain(ctx); err != nil {
		if ctx.Err() != nil {
			return nil
		}
		return err
	}
	ticker := time.NewTicker(r.cfg.PollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		case <-r.notify:
		}
		if err := r.Drain(ctx); err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return err
		}
	}
}

// Drain membatasi satu putaran agar backlog besar tidak memonopoli worker.
func (r *Relay) Drain(ctx context.Context) error {
	if r == nil || ctx == nil {
		return hrerrors.ErrInvalidInput
	}
	if r.cfg.Standby {
		return nil
	}
	for range 25 {
		if err := ctx.Err(); err != nil {
			return err
		}
		processed, err := r.drainOne(ctx)
		if err != nil {
			return err
		}
		if !processed {
			return nil
		}
	}
	r.Notify()
	return nil
}

func (r *Relay) drainOne(ctx context.Context) (bool, error) {
	processed := false
	var terminalID, terminalType, terminalCode string
	txCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), r.cfg.PublishTimeout+5*time.Second)
	defer cancel()
	err := r.tx.WithinTransaction(txCtx, func(tc context.Context) error {
		events, err := r.store.SelectDue(tc, time.Now().UTC(), 1, r.cfg.MaxAttempts)
		if err != nil {
			return err
		}
		if len(events) == 0 {
			return nil
		}
		processed = true
		event := events[0]
		attempts := event.AttemptCount + 1
		if attempts <= 0 {
			return hrerrors.ErrConflict
		}
		if err := validatePayload(&event); err != nil {
			terminalID, terminalType, terminalCode = event.ID, event.EventType, "invalid_payload"
			return r.store.MarkRetry(tc, event.ID, time.Now().UTC(), int32(r.cfg.MaxAttempts), "invalid_payload")
		}
		pubCtx, stop := context.WithTimeout(ctx, r.cfg.PublishTimeout)
		defer stop()
		if err := r.publisher.Publish(pubCtx, &event); err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			if int(attempts) == r.cfg.MaxAttempts {
				terminalID, terminalType, terminalCode = event.ID, event.EventType, "delivery_failed"
			}
			return r.store.MarkRetry(tc, event.ID, time.Now().UTC().Add(retryDelay(int(attempts))), attempts, "delivery_failed")
		}
		return r.store.MarkPublished(tc, event.ID, time.Now().UTC(), attempts)
	})
	if err == nil && terminalCode != "" {
		r.logger.WarnContext(ctx, "HRM outbox delivery exhausted",
			"event_id", terminalID, "event_type", terminalType, "error_code", terminalCode)
	}
	return processed, err
}

func validatePayload(event *entity.OutboxEvent) error {
	if event == nil || event.ID == "" || event.AggregateID == "" || len(event.PayloadDigest) != sha256.Size {
		return hrerrors.ErrInvalidInput
	}
	if err := event.Validate(); err != nil {
		return err
	}
	if !json.Valid(event.Payload) {
		return hrerrors.ErrInvalidInput
	}
	// Marshal ulang mengikuti format map JSON penulis sebelum PostgreSQL menormalisasi JSONB.
	var object map[string]any
	decoder := json.NewDecoder(bytes.NewReader(event.Payload))
	decoder.UseNumber()
	if err := decoder.Decode(&object); err != nil || object == nil {
		return hrerrors.ErrInvalidInput
	}
	canonical, err := json.Marshal(object)
	if err != nil {
		return hrerrors.ErrInvalidInput
	}
	digest := sha256.Sum256(canonical)
	if subtle.ConstantTimeCompare(digest[:], event.PayloadDigest) != 1 {
		return hrerrors.ErrConstraintViolation
	}
	return nil
}

func retryDelay(attempt int) time.Duration {
	delay := time.Second
	for i := 1; i < attempt && delay < time.Minute; i++ {
		if delay > 30*time.Second {
			return time.Minute
		}
		delay *= 2
	}
	if delay > time.Minute {
		return time.Minute
	}
	return delay
}
