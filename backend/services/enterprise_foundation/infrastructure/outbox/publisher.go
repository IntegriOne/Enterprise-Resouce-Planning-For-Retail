package outbox

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/entity"
	eferrors "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/errors"
	domainservice "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/service"
)

const (
	defaultWebhookTimeout = 5 * time.Second
	maxResponseBodyBytes  = 4096
)

// Publisher memublikasikan satu event outbox Enterprise Foundation ke transport atau konsumen eksternal.
type Publisher interface {
	Publish(ctx context.Context, event *entity.OutboxEvent) error
}

// PublishFunc mengadaptasi fungsi biasa menjadi implementasi interface Publisher.
type PublishFunc func(ctx context.Context, event *entity.OutboxEvent) error

// Publish mengeksekusi fungsi penerbitan untuk satu event outbox.
func (f PublishFunc) Publish(ctx context.Context, event *entity.OutboxEvent) error {
	if f == nil {
		return MarkPermanent(ErrPublisherUnavailable)
	}
	return f(ctx, event)
}

// Envelope merepresentasikan amplop JSON terstruktur yang dikirim ke konsumen event Enterprise Foundation.
type Envelope struct {
	EventID        string          `json:"event_id"`
	IdempotencyKey string          `json:"idempotency_key"`
	EventType      string          `json:"event_type"`
	AggregateType  string          `json:"aggregate_type"`
	AggregateID    string          `json:"aggregate_id"`
	Payload        json.RawMessage `json:"payload"`
	AttemptCount   int             `json:"attempt_count"`
	OccurredAt     time.Time       `json:"occurred_at"`
	PublishedAt    time.Time       `json:"published_at"`
}

// BuildEnvelope membentuk dan memvalidasi amplop JSON event sebelum dikirim keluar proses.
func BuildEnvelope(event *entity.OutboxEvent, publishedAt time.Time) (Envelope, error) {
	if event == nil {
		return Envelope{}, MarkPermanent(eferrors.ErrOutboxEventNotFound)
	}
	if err := entity.ValidateUUID("event_id", event.ID); err != nil {
		return Envelope{}, MarkPermanent(err)
	}
	if err := entity.ValidateUUID("aggregate_id", event.AggregateID); err != nil {
		return Envelope{}, MarkPermanent(err)
	}
	eventType := strings.TrimSpace(event.EventType)
	if eventType == "" {
		return Envelope{}, MarkPermanent(eferrors.NewValidation("event_type", eferrors.Code("enterprise_foundation.outbox.empty_event_type"), "outbox event type is required"))
	}
	aggregateType := strings.TrimSpace(event.AggregateType)
	if aggregateType == "" {
		return Envelope{}, MarkPermanent(eferrors.NewValidation("aggregate_type", eferrors.Code("enterprise_foundation.outbox.empty_aggregate_type"), "outbox aggregate type is required"))
	}

	payloadBytes := event.PayloadCopy()
	validatedPayload, err := entity.ValidateJSONObject("payload", payloadBytes)
	if err != nil {
		return Envelope{}, MarkPermanent(err)
	}

	if publishedAt.IsZero() {
		publishedAt = time.Now()
	}
	occurredAt := event.CreatedAt
	if occurredAt.IsZero() {
		occurredAt = publishedAt
	}

	eventID := entity.NormalizeUUID(event.ID)
	return Envelope{
		EventID:        eventID,
		IdempotencyKey: eventID,
		EventType:      eventType,
		AggregateType:  aggregateType,
		AggregateID:    entity.NormalizeUUID(event.AggregateID),
		Payload:        json.RawMessage(validatedPayload),
		AttemptCount:   event.AttemptCount,
		OccurredAt:     occurredAt.UTC(),
		PublishedAt:    publishedAt.UTC(),
	}, nil
}

// RouterPublisher merutekan event outbox ke Publisher berbeda berdasarkan event_type persis atau pola prefix "<prefix>.*".
type RouterPublisher struct {
	mu              sync.RWMutex
	exactRoutes     map[string]Publisher
	prefixRoutes    []prefixRoute
	fallback        Publisher
	ignoreUnmatched bool
}

type prefixRoute struct {
	prefix    string
	publisher Publisher
}

// NewRouterPublisher membentuk router penerbit event dengan perilaku default fail-closed bila rute tidak ditemukan.
func NewRouterPublisher(fallback Publisher, ignoreUnmatched bool) *RouterPublisher {
	return &RouterPublisher{
		exactRoutes:     make(map[string]Publisher),
		fallback:        fallback,
		ignoreUnmatched: ignoreUnmatched,
	}
}

// Register mendaftarkan Publisher untuk pola eventType persis atau prefix berakhiran ".*" secara aman dari race condition.
func (r *RouterPublisher) Register(pattern string, pub Publisher) error {
	if r == nil {
		return ErrPublisherUnavailable
	}
	if pub == nil {
		return eferrors.NewValidation("publisher", eferrors.Code("enterprise_foundation.outbox.nil_route_publisher"), "route publisher must not be nil")
	}
	cleaned := strings.TrimSpace(pattern)
	if cleaned == "" {
		return eferrors.NewValidation("pattern", eferrors.Code("enterprise_foundation.outbox.empty_route_pattern"), "route pattern must not be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if strings.HasSuffix(cleaned, ".*") {
		prefix := strings.TrimSuffix(cleaned, "*")
		replaced := false
		for i := range r.prefixRoutes {
			if r.prefixRoutes[i].prefix == prefix {
				r.prefixRoutes[i].publisher = pub
				replaced = true
				break
			}
		}
		if !replaced {
			r.prefixRoutes = append(r.prefixRoutes, prefixRoute{prefix: prefix, publisher: pub})
			// Urutkan prefix terpanjang lebih dulu agar pencocokan rute spesifik lebih diprioritaskan.
			sort.Slice(r.prefixRoutes, func(i, j int) bool {
				return len(r.prefixRoutes[i].prefix) > len(r.prefixRoutes[j].prefix)
			})
		}
		return nil
	}

	if r.exactRoutes == nil {
		r.exactRoutes = make(map[string]Publisher)
	}
	r.exactRoutes[cleaned] = pub
	return nil
}

// Publish meneruskan event ke rute yang paling spesifik, atau ke fallback jika tersedia.
func (r *RouterPublisher) Publish(ctx context.Context, event *entity.OutboxEvent) error {
	if r == nil {
		return MarkPermanent(ErrPublisherUnavailable)
	}
	if event == nil {
		return MarkPermanent(eferrors.ErrOutboxEventNotFound)
	}

	r.mu.RLock()
	pub := r.resolveLocked(strings.TrimSpace(event.EventType))
	ignoreUnmatched := r.ignoreUnmatched
	r.mu.RUnlock()

	if pub == nil {
		if ignoreUnmatched {
			return nil
		}
		return MarkPermanent(fmt.Errorf("%w: %s", ErrRouteNotFound, event.EventType))
	}
	return pub.Publish(ctx, event)
}

func (r *RouterPublisher) resolveLocked(eventType string) Publisher {
	if pub, ok := r.exactRoutes[eventType]; ok && pub != nil {
		return pub
	}
	for _, route := range r.prefixRoutes {
		if strings.HasPrefix(eventType, route.prefix) && route.publisher != nil {
			return route.publisher
		}
	}
	return r.fallback
}

// MultiPublisher mengirimkan satu event ke beberapa Publisher secara berurutan (fan-out fail-fast).
type MultiPublisher struct {
	publishers []Publisher
}

// NewMultiPublisher membangun fan-out publisher dengan validasi minimal satu publisher aktif.
func NewMultiPublisher(publishers ...Publisher) (*MultiPublisher, error) {
	filtered := make([]Publisher, 0, len(publishers))
	for _, p := range publishers {
		if p != nil {
			filtered = append(filtered, p)
		}
	}
	if len(filtered) == 0 {
		return nil, ErrPublisherUnavailable
	}
	return &MultiPublisher{publishers: filtered}, nil
}

// Publish mengirimkan salinan event ke seluruh Publisher terdaftar secara berurutan.
func (m *MultiPublisher) Publish(ctx context.Context, event *entity.OutboxEvent) error {
	if m == nil || len(m.publishers) == 0 {
		return MarkPermanent(ErrPublisherUnavailable)
	}
	if event == nil {
		return MarkPermanent(eferrors.ErrOutboxEventNotFound)
	}
	for _, pub := range m.publishers {
		if err := pub.Publish(ctx, event.Clone()); err != nil {
			return err
		}
	}
	return nil
}

// LogPublisher mencatat metadata event outbox terstruktur ke slog untuk mode diagnostik lokal non-produksi.
type LogPublisher struct {
	logger *slog.Logger
	clock  domainservice.Clock
}

// NewLogPublisher membuat Publisher berbasis slog untuk audit/diagnostik lokal.
func NewLogPublisher(logger *slog.Logger, clock domainservice.Clock) *LogPublisher {
	if logger == nil {
		logger = slog.Default()
	}
	if clock == nil {
		clock = domainservice.SystemClock{}
	}
	return &LogPublisher{
		logger: logger,
		clock:  clock,
	}
}

// Publish memvalidasi amplop event dan mencatat pengiriman metadata secara terstruktur tanpa membocorkan PII mentah.
func (p *LogPublisher) Publish(_ context.Context, event *entity.OutboxEvent) error {
	if p == nil {
		return MarkPermanent(ErrPublisherUnavailable)
	}
	env, err := BuildEnvelope(event, p.clock.Now())
	if err != nil {
		return err
	}
	p.logger.Info(
		"published outbox event to log sink",
		slog.String("event_id", env.EventID),
		slog.String("event_type", env.EventType),
		slog.String("aggregate_type", env.AggregateType),
		slog.String("aggregate_id", env.AggregateID),
		slog.Int("attempt_count", env.AttemptCount),
	)
	return nil
}

// WebhookConfig menyimpan parameter konfigurasi HTTP push webhook untuk konsumen eksternal.
type WebhookConfig struct {
	Endpoint     string
	Timeout      time.Duration
	HMACSecret   []byte
	RequireHTTPS bool
	Headers      map[string]string
	HTTPClient   *http.Client
	Clock        domainservice.Clock
}

// WebhookPublisher memublikasikan amplop event outbox ke endpoint HTTP(S) dengan kunci idempotensi dan tanda tangan HMAC opsional.
type WebhookPublisher struct {
	endpoint   string
	hmacSecret []byte
	headers    map[string]string
	client     *http.Client
	clock      domainservice.Clock
}

// NewWebhookPublisher memvalidasi konfigurasi URL/timeout dan membentuk HTTP webhook publisher yang aman.
func NewWebhookPublisher(cfg WebhookConfig) (*WebhookPublisher, error) {
	rawURL := strings.TrimSpace(cfg.Endpoint)
	if rawURL == "" || strings.ContainsAny(rawURL, "\r\n\t ") {
		return nil, ErrInvalidWebhookURL
	}
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed == nil || parsed.Host == "" || parsed.User != nil {
		return nil, fmt.Errorf("%w: invalid url syntax or embedded credentials", ErrInvalidWebhookURL)
	}
	scheme := strings.ToLower(parsed.Scheme)
	if cfg.RequireHTTPS {
		if scheme != "https" {
			return nil, fmt.Errorf("%w: https scheme is required", ErrInvalidWebhookURL)
		}
	} else if scheme != "http" && scheme != "https" {
		return nil, fmt.Errorf("%w: only http and https schemes are supported", ErrInvalidWebhookURL)
	}

	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = defaultWebhookTimeout
	}
	client := cfg.HTTPClient
	if client == nil {
		client = &http.Client{
			Timeout: timeout,
			// Tolak pengalihan redirect HTTP otomatis agar pengiriman outbox hanya menuju endpoint terkonfigurasi.
			CheckRedirect: func(*http.Request, []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
	}
	clock := cfg.Clock
	if clock == nil {
		clock = domainservice.SystemClock{}
	}

	var copiedSecret []byte
	if len(cfg.HMACSecret) > 0 {
		copiedSecret = bytes.Clone(cfg.HMACSecret)
	}

	var copiedHeaders map[string]string
	if len(cfg.Headers) > 0 {
		copiedHeaders = make(map[string]string, len(cfg.Headers))
		for k, v := range cfg.Headers {
			key := strings.TrimSpace(k)
			if key == "" {
				continue
			}
			copiedHeaders[key] = v
		}
	}

	return &WebhookPublisher{
		endpoint:   parsed.String(),
		hmacSecret: copiedSecret,
		headers:    copiedHeaders,
		client:     client,
		clock:      clock,
	}, nil
}

// Publish menyerialisasi amplop event dan mengirimkannya melalui HTTP POST dengan klasifikasi error permanen vs sementara.
func (p *WebhookPublisher) Publish(ctx context.Context, event *entity.OutboxEvent) error {
	if p == nil || p.client == nil {
		return MarkPermanent(ErrPublisherUnavailable)
	}
	if ctx == nil {
		ctx = context.Background()
	}

	now := p.clock.Now()
	env, err := BuildEnvelope(event, now)
	if err != nil {
		return err
	}

	body, err := json.Marshal(env)
	if err != nil {
		return MarkPermanent(fmt.Errorf("marshal outbox envelope: %w", err))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.endpoint, bytes.NewReader(body))
	if err != nil {
		return MarkPermanent(fmt.Errorf("build webhook request: %w", err))
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "integrione-enterprise-foundation-outbox/1.0")
	req.Header.Set("Idempotency-Key", env.IdempotencyKey)
	req.Header.Set("X-Integrione-Event-ID", env.EventID)
	req.Header.Set("X-Integrione-Event-Type", env.EventType)
	req.Header.Set("X-Integrione-Aggregate-Type", env.AggregateType)
	req.Header.Set("X-Integrione-Aggregate-ID", env.AggregateID)
	for k, v := range p.headers {
		req.Header.Set(k, v)
	}

	if len(p.hmacSecret) > 0 {
		tsStr := strconv.FormatInt(now.Unix(), 10)
		mac := hmac.New(sha256.New, p.hmacSecret)
		mac.Write([]byte(tsStr))
		mac.Write([]byte("."))
		mac.Write(body)
		req.Header.Set("X-Integrione-Timestamp", tsStr)
		req.Header.Set("X-Integrione-Signature-256", "sha256="+hex.EncodeToString(mac.Sum(nil)))
	}

	resp, err := p.client.Do(req)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return err
		}
		return fmt.Errorf("webhook request failed: %w", err)
	}
	defer resp.Body.Close()

	respSnippet, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, maxResponseBodyBytes))

	if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
		return nil
	}

	msg := strings.TrimSpace(string(respSnippet))
	if msg == "" {
		msg = http.StatusText(resp.StatusCode)
	}
	statusErr := fmt.Errorf("webhook returned status %d: %s", resp.StatusCode, msg)

	// Status 408, 429, dan 5xx dianggap sementara agar dapat dijadwalkan retry; 4xx lain bersifat permanen.
	if resp.StatusCode >= http.StatusBadRequest &&
		resp.StatusCode < http.StatusInternalServerError &&
		resp.StatusCode != http.StatusRequestTimeout &&
		resp.StatusCode != http.StatusTooManyRequests {
		return MarkPermanent(statusErr)
	}
	return statusErr
}
