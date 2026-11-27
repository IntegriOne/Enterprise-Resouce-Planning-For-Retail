package mailer

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"runtime/debug"
	"sync"
	"time"

	domainservice "github.com/integrione/erp-retail/backend/services/iam/domain/service"
)

var (
	// ErrEmailQueueFull menandakan kapasitas antrean pengiriman asinkron sedang penuh.
	ErrEmailQueueFull = errors.New("email delivery queue is full")
	// ErrDispatcherClosed menandakan dispatcher sudah ditutup saat shutdown.
	ErrDispatcherClosed = errors.New("email dispatcher is closed")
)

// DispatcherConfig mengonfigurasi eksekusi asinkron/sinkron dan kebijakan retry email transaksional.
type DispatcherConfig struct {
	Mode                DeliveryMode
	Workers             int
	QueueSize           int
	SendTimeout         time.Duration
	MaxAttempts         int
	InitialRetryBackoff time.Duration
}

type emailJob struct {
	msg domainservice.EmailMessage
}

// Dispatcher membungkus EmailSender dengan antrean bounded worker asinkron (mencegah timing leak enumerasi akun)
// serta retry exponential backoff untuk error jaringan sementara.
type Dispatcher struct {
	inner  domainservice.EmailSender
	cfg    DispatcherConfig
	logger *slog.Logger

	mu      sync.RWMutex
	started bool
	closed  bool
	jobs    chan emailJob
	wg      sync.WaitGroup
}

// NewDispatcher membentuk Dispatcher baru untuk backend EmailSender tertentu.
func NewDispatcher(inner domainservice.EmailSender, cfg DispatcherConfig, logger *slog.Logger) *Dispatcher {
	if cfg.Mode == "" {
		cfg.Mode = DeliveryModeAsync
	}
	if cfg.Workers <= 0 {
		cfg.Workers = 2
	}
	if cfg.QueueSize <= 0 {
		cfg.QueueSize = 256
	}
	if cfg.SendTimeout <= 0 {
		cfg.SendTimeout = 8 * time.Second
	}
	if cfg.MaxAttempts <= 0 {
		cfg.MaxAttempts = 3
	}
	if cfg.InitialRetryBackoff <= 0 {
		cfg.InitialRetryBackoff = 250 * time.Millisecond
	}

	return &Dispatcher{
		inner:  inner,
		cfg:    cfg,
		logger: logger,
		jobs:   make(chan emailJob, cfg.QueueSize),
	}
}

// ProviderName mengembalikan pengenal provider yang dibungkus.
func (d *Dispatcher) ProviderName() string {
	if d == nil || d.inner == nil {
		return string(ProviderDisabled)
	}
	return d.inner.ProviderName()
}

// Start menyalakan pool worker untuk mode pengiriman asinkron secara idempoten.
func (d *Dispatcher) Start(_ context.Context) {
	if d == nil || d.cfg.Mode != DeliveryModeAsync {
		return
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.started || d.closed {
		return
	}
	d.started = true
	for i := 0; i < d.cfg.Workers; i++ {
		d.wg.Add(1)
		go d.workerLoop()
	}
}

// Close menghentikan penerimaan pesan baru lalu menguras seluruh pekerjaan tersisa hingga worker selesai atau batas waktu habis.
func (d *Dispatcher) Close(ctx context.Context) error {
	if d == nil {
		return nil
	}
	d.mu.Lock()
	if d.closed {
		d.mu.Unlock()
		return nil
	}
	d.closed = true
	close(d.jobs)
	d.mu.Unlock()

	done := make(chan struct{})
	go func() {
		d.wg.Wait()
		close(done)
	}()

	if ctx == nil {
		ctx = context.Background()
	}
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("email dispatcher shutdown timed out: %w", ctx.Err())
	}
}

// Send mengirim langsung pada mode sync atau memasukkan pekerjaan ke antrean bounded secara constant-time pada mode async.
func (d *Dispatcher) Send(ctx context.Context, msg domainservice.EmailMessage) (*domainservice.EmailSendResult, error) {
	if d == nil || d.inner == nil {
		return &domainservice.EmailSendResult{Provider: string(ProviderDisabled), MessageID: "disabled"}, nil
	}
	if err := validateOutboundMessage(msg); err != nil {
		return nil, &PermanentError{Err: err}
	}

	if d.cfg.Mode == DeliveryModeSync {
		return d.sendWithRetry(ctx, msg)
	}

	d.mu.RLock()
	closed := d.closed
	started := d.started
	if !closed && started {
		select {
		case d.jobs <- emailJob{msg: msg}:
			d.mu.RUnlock()
			return &domainservice.EmailSendResult{
				Provider:  d.inner.ProviderName(),
				MessageID: "queued",
			}, nil
		default:
			d.mu.RUnlock()
			return nil, ErrEmailQueueFull
		}
	}
	d.mu.RUnlock()

	if closed {
		return nil, ErrDispatcherClosed
	}

	// Fallback bila caller belum memanggil Start(ctx) (misalnya pada pengujian terisolasi): jalankan dengan batas waktu terkelola.
	return d.sendWithRetry(ctx, msg)
}

func (d *Dispatcher) workerLoop() {
	defer d.wg.Done()
	for job := range d.jobs {
		d.processJob(job)
	}
}

func (d *Dispatcher) processJob(job emailJob) {
	log := d.logger
	if log == nil {
		log = slog.Default()
	}
	defer func() {
		if rec := recover(); rec != nil {
			log.Error("email worker panic recovered",
				slog.Any("panic", rec),
				slog.String("stack", string(debug.Stack())),
			)
		}
	}()
	// Konteks terisolasi per pengiriman agar pembatalan request RPC klien tidak membatalkan pengiriman email yang sudah masuk antrean.
	jobCtx, cancel := context.WithTimeout(context.Background(), d.totalAttemptBudget())
	defer cancel()
	if _, err := d.sendWithRetry(jobCtx, job.msg); err != nil {
		log.Error("transactional email delivery failed",
			slog.String("provider", d.ProviderName()),
			slog.String("purpose", string(job.msg.Purpose)),
			slog.String("recipient", domainservice.MaskEmailForLog(job.msg.To)),
			slog.String("error", err.Error()),
		)
	}
}

func (d *Dispatcher) totalAttemptBudget() time.Duration {
	budget := time.Duration(d.cfg.MaxAttempts)*d.cfg.SendTimeout + 5*time.Second
	if budget < d.cfg.SendTimeout {
		return d.cfg.SendTimeout
	}
	return budget
}

func (d *Dispatcher) sendWithRetry(ctx context.Context, msg domainservice.EmailMessage) (*domainservice.EmailSendResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	var lastErr error
	backoff := d.cfg.InitialRetryBackoff
	for attempt := 1; attempt <= d.cfg.MaxAttempts; attempt++ {
		attemptCtx, cancel := context.WithTimeout(ctx, d.cfg.SendTimeout)
		res, err := d.inner.Send(attemptCtx, msg)
		cancel()

		if err == nil {
			if d.logger != nil && d.inner.ProviderName() != string(ProviderDisabled) && d.inner.ProviderName() != string(ProviderLog) {
				msgID := ""
				if res != nil {
					msgID = res.MessageID
				}
				d.logger.Info("transactional email delivered",
					slog.String("provider", d.inner.ProviderName()),
					slog.String("purpose", string(msg.Purpose)),
					slog.String("recipient", domainservice.MaskEmailForLog(msg.To)),
					slog.String("message_id", msgID),
					slog.Int("attempt", attempt),
				)
			}
			return res, nil
		}

		lastErr = err
		if IsPermanent(err) || attempt == d.cfg.MaxAttempts {
			break
		}

		timer := time.NewTimer(backoff)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, fmt.Errorf("email delivery aborted during retry wait: %w", ctx.Err())
		case <-timer.C:
		}
		backoff *= 2
	}

	return nil, lastErr
}
