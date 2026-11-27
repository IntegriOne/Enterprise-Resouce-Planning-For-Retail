package outbox

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
)

// Envelope mengirim identitas dan versi event agar penerima dapat melakukan deduplikasi.
type Envelope struct {
	EventID          string          `json:"event_id"`
	IdempotencyKey   string          `json:"idempotency_key"`
	AggregateType    string          `json:"aggregate_type"`
	AggregateID      string          `json:"aggregate_id"`
	AggregateVersion int64           `json:"aggregate_version"`
	EventType        string          `json:"event_type"`
	SchemaVersion    int32           `json:"schema_version"`
	OccurredAt       time.Time       `json:"occurred_at"`
	Payload          json.RawMessage `json:"payload"`
}

// WebhookPublisher mengirim event ke endpoint yang dikonfigurasi dengan tanda tangan HMAC.
type WebhookPublisher struct {
	endpoint string
	secret   []byte
	client   *http.Client
}

// NewWebhookPublisher membangun client tanpa redirect agar tanda tangan tidak bocor ke host lain.
func NewWebhookPublisher(endpoint string, secret []byte, timeout time.Duration) (*WebhookPublisher, error) {
	if endpoint == "" || len(secret) < 32 || timeout <= 0 || timeout > time.Minute {
		return nil, hrerrors.ErrInvalidInput
	}
	parsed, err := url.Parse(endpoint)
	if err != nil || parsed.Host == "" || parsed.User != nil || parsed.Fragment != "" ||
		(parsed.Scheme != "http" && parsed.Scheme != "https") {
		return nil, hrerrors.ErrInvalidInput
	}
	return &WebhookPublisher{
		endpoint: endpoint,
		secret:   append([]byte(nil), secret...),
		client: &http.Client{
			Timeout:       timeout,
			CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
		},
	}, nil
}

// Publish menandai sukses hanya pada respons HTTP 2xx; penerima wajib dedupe dengan event_id.
func (p *WebhookPublisher) Publish(ctx context.Context, event *entity.OutboxEvent) error {
	if p == nil || p.client == nil || event == nil {
		return hrerrors.ErrUnavailable
	}
	if err := validatePayload(event); err != nil {
		return err
	}
	envelope := Envelope{
		EventID: event.ID, IdempotencyKey: event.ID,
		AggregateType: event.AggregateType, AggregateID: event.AggregateID,
		AggregateVersion: event.AggregateVersion, EventType: event.EventType,
		SchemaVersion: event.SchemaVersion, OccurredAt: event.OccurredAt,
		Payload: json.RawMessage(event.Payload),
	}
	body, err := json.Marshal(envelope)
	if err != nil {
		return hrerrors.ErrInvalidInput
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.endpoint, bytes.NewReader(body))
	if err != nil {
		return hrerrors.ErrInvalidInput
	}
	signature := hmac.New(sha256.New, p.secret)
	if _, err := signature.Write(body); err != nil {
		return hrerrors.ErrUnavailable
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-IG-Event-ID", event.ID)
	req.Header.Set("X-IG-Signature-SHA256", hex.EncodeToString(signature.Sum(nil)))
	response, err := p.client.Do(req)
	if err != nil {
		return hrerrors.ErrUnavailable
	}
	defer response.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 4096))
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return hrerrors.ErrUnavailable
	}
	return nil
}

// LogPublisher hanya untuk development; payload sensitif tidak dicetak.
type LogPublisher struct{ logger *slog.Logger }

// NewLogPublisher menyiapkan sink event lokal untuk development eksplisit.
func NewLogPublisher(logger *slog.Logger) *LogPublisher {
	if logger == nil {
		logger = slog.Default()
	}
	return &LogPublisher{logger: logger}
}

// Publish menulis identitas event tanpa payload dan menganggapnya terkirim.
func (p *LogPublisher) Publish(ctx context.Context, event *entity.OutboxEvent) error {
	if p == nil || event == nil {
		return hrerrors.ErrInvalidInput
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	p.logger.Info("HRM outbox event", "event_id", event.ID, "event_type", event.EventType,
		"aggregate_id", event.AggregateID, "aggregate_version", event.AggregateVersion)
	return nil
}

type workerTerminationRoute struct {
	general     Publisher
	termination Publisher
}

// RouteWorkerTermination mengarahkan event terminasi hanya ke receiver IAM.
func RouteWorkerTermination(general, termination Publisher) Publisher {
	return workerTerminationRoute{general: general, termination: termination}
}

func (r workerTerminationRoute) Publish(ctx context.Context, event *entity.OutboxEvent) error {
	if event == nil {
		return hrerrors.ErrInvalidInput
	}
	if event.EventType == entity.WorkerTerminationRecorded {
		if r.termination == nil {
			return hrerrors.ErrUnavailable
		}
		return r.termination.Publish(ctx, event)
	}
	if r.general == nil {
		return hrerrors.ErrUnavailable
	}
	return r.general.Publish(ctx, event)
}
