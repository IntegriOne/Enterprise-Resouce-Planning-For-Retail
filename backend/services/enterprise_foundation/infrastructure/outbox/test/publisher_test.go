package outbox_test

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/entity"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/infrastructure/outbox"
)

type fixedClock struct {
	now time.Time
}

func (c fixedClock) Now() time.Time {
	return c.now
}

func sampleOutboxEvent(t *testing.T, eventType string, payload any) *entity.OutboxEvent {
	t.Helper()
	now := time.Date(2026, 11, 15, 21, 0, 0, 0, time.UTC)
	evt, err := entity.NewOutboxEvent(
		"11111111-2222-4333-8444-555555555555",
		eventType,
		entity.AggregateTypeLegalEntity,
		"aaaaaaaa-bbbb-4ccc-8ddd-eeeeeeeeeeee",
		payload,
		now,
	)
	if err != nil {
		t.Fatalf("NewOutboxEvent failed: %v", err)
	}
	return evt
}

func TestBuildEnvelope(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 11, 15, 21, 5, 0, 0, time.UTC)
	evt := sampleOutboxEvent(t, entity.EventTypeLegalEntityCreated, map[string]any{"entity_code": "PT-ONE"})

	env, err := outbox.BuildEnvelope(evt, now)
	if err != nil {
		t.Fatalf("BuildEnvelope failed: %v", err)
	}
	if env.EventID != evt.ID || env.IdempotencyKey != evt.ID {
		t.Fatalf("unexpected envelope IDs: %+v", env)
	}
	if env.EventType != entity.EventTypeLegalEntityCreated || env.AggregateType != entity.AggregateTypeLegalEntity {
		t.Fatalf("unexpected envelope metadata: %+v", env)
	}

	t.Run("corrupted payload triggers permanent error", func(t *testing.T) {
		t.Parallel()
		bad := sampleOutboxEvent(t, entity.EventTypeLegalEntityCreated, map[string]any{"ok": true})
		bad.Payload = []byte("{not-json")
		_, err := outbox.BuildEnvelope(bad, now)
		if err == nil || !outbox.IsPermanent(err) {
			t.Fatalf("expected permanent error for invalid payload, got %v", err)
		}
	})
}

func TestRouterPublisher(t *testing.T) {
	t.Parallel()

	router := outbox.NewRouterPublisher(nil, false)
	var called []string
	var mu sync.Mutex
	record := func(name string) outbox.Publisher {
		return outbox.PublishFunc(func(_ context.Context, ev *entity.OutboxEvent) error {
			mu.Lock()
			called = append(called, name+":"+ev.EventType)
			mu.Unlock()
			return nil
		})
	}

	if err := router.Register("party.*", record("party_wildcard")); err != nil {
		t.Fatalf("Register wildcard failed: %v", err)
	}
	if err := router.Register("party.role.*", record("party_role_wildcard")); err != nil {
		t.Fatalf("Register more specific wildcard failed: %v", err)
	}
	if err := router.Register(entity.EventTypePartyCreated, record("party_created_exact")); err != nil {
		t.Fatalf("Register exact failed: %v", err)
	}

	// Uji prioritas rute: exact > longest prefix wildcard > unmatched permanent fail-closed.
	if err := router.Publish(context.Background(), sampleOutboxEvent(t, entity.EventTypePartyCreated, map[string]any{"k": 1})); err != nil {
		t.Fatalf("Publish exact failed: %v", err)
	}
	if err := router.Publish(context.Background(), sampleOutboxEvent(t, "party.role.assigned", map[string]any{"k": 2})); err != nil {
		t.Fatalf("Publish longer wildcard failed: %v", err)
	}
	if err := router.Publish(context.Background(), sampleOutboxEvent(t, entity.EventTypePartyUpdated, map[string]any{"k": 3})); err != nil {
		t.Fatalf("Publish shorter wildcard failed: %v", err)
	}

	err := router.Publish(context.Background(), sampleOutboxEvent(t, entity.EventTypeLocationCreated, map[string]any{"k": 4}))
	if err == nil || !outbox.IsPermanent(err) || !errors.Is(err, outbox.ErrRouteNotFound) {
		t.Fatalf("expected permanent ErrRouteNotFound for unmatched event, got %v", err)
	}

	want := []string{
		"party_created_exact:" + entity.EventTypePartyCreated,
		"party_role_wildcard:party.role.assigned",
		"party_wildcard:" + entity.EventTypePartyUpdated,
	}
	mu.Lock()
	defer mu.Unlock()
	if strings.Join(called, ",") != strings.Join(want, ",") {
		t.Fatalf("got calls %v, want %v", called, want)
	}
}

func TestMultiAndLogPublisher(t *testing.T) {
	t.Parallel()

	if _, err := outbox.NewMultiPublisher(nil); !errors.Is(err, outbox.ErrPublisherUnavailable) {
		t.Fatalf("expected ErrPublisherUnavailable for empty MultiPublisher, got %v", err)
	}

	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	logPub := outbox.NewLogPublisher(logger, fixedClock{now: time.Date(2026, 11, 15, 21, 0, 0, 0, time.UTC)})

	var seenID string
	customPub := outbox.PublishFunc(func(_ context.Context, ev *entity.OutboxEvent) error {
		seenID = ev.ID
		return nil
	})
	multi, err := outbox.NewMultiPublisher(logPub, customPub)
	if err != nil {
		t.Fatalf("NewMultiPublisher failed: %v", err)
	}
	evt := sampleOutboxEvent(t, entity.EventTypeOrgUnitCreated, map[string]any{"code": "OPS"})
	if err := multi.Publish(context.Background(), evt); err != nil {
		t.Fatalf("MultiPublisher.Publish failed: %v", err)
	}
	if seenID != evt.ID {
		t.Fatalf("got seenID %q, want %q", seenID, evt.ID)
	}
	if !strings.Contains(buf.String(), `"event_type":"organization_unit.created"`) {
		t.Fatalf("expected structured slog entry, got %s", buf.String())
	}
}

func TestWebhookPublisherValidationAndDelivery(t *testing.T) {
	t.Parallel()

	invalidURLs := []outbox.WebhookConfig{
		{Endpoint: ""},
		{Endpoint: "ftp://example.com/hook"},
		{Endpoint: "http://user:pass@example.com/hook"},
		{Endpoint: "http://example.com/hook", RequireHTTPS: true},
		{Endpoint: "https://example.com/bad url"},
	}
	for _, cfg := range invalidURLs {
		if _, err := outbox.NewWebhookPublisher(cfg); !errors.Is(err, outbox.ErrInvalidWebhookURL) {
			t.Fatalf("expected ErrInvalidWebhookURL for %+v, got %v", cfg, err)
		}
	}

	t.Run("signed delivery succeeds on 2xx", func(t *testing.T) {
		t.Parallel()

		secret := []byte("0123456789abcdef0123456789abcdef")
		now := time.Date(2026, 11, 15, 21, 10, 0, 0, time.UTC)
		var (
			gotMethod    string
			gotIdemKey   string
			gotTimestamp string
			gotSig       string
			gotCustomHdr string
			gotBody      []byte
		)
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotMethod = r.Method
			gotIdemKey = r.Header.Get("Idempotency-Key")
			gotTimestamp = r.Header.Get("X-Integrione-Timestamp")
			gotSig = r.Header.Get("X-Integrione-Signature-256")
			gotCustomHdr = r.Header.Get("X-Custom-Header")
			gotBody, _ = io.ReadAll(r.Body)
			w.WriteHeader(http.StatusAccepted)
		}))
		defer srv.Close()

		pub, err := outbox.NewWebhookPublisher(outbox.WebhookConfig{
			Endpoint:   srv.URL,
			Timeout:    2 * time.Second,
			HMACSecret: secret,
			Headers:    map[string]string{"X-Custom-Header": "ef-relay"},
			Clock:      fixedClock{now: now},
		})
		if err != nil {
			t.Fatalf("NewWebhookPublisher failed: %v", err)
		}

		evt := sampleOutboxEvent(t, entity.EventTypeLegalEntityCreated, map[string]any{"code": "INT-01"})
		if err := pub.Publish(context.Background(), evt); err != nil {
			t.Fatalf("WebhookPublisher.Publish failed: %v", err)
		}
		if gotMethod != http.MethodPost || gotIdemKey != evt.ID || gotCustomHdr != "ef-relay" {
			t.Fatalf("unexpected request metadata: method=%s idem=%s hdr=%s", gotMethod, gotIdemKey, gotCustomHdr)
		}

		mac := hmac.New(sha256.New, secret)
		mac.Write([]byte(gotTimestamp + "."))
		mac.Write(gotBody)
		wantSig := "sha256=" + hex.EncodeToString(mac.Sum(nil))
		if gotSig != wantSig {
			t.Fatalf("signature mismatch: got %q want %q", gotSig, wantSig)
		}

		var env outbox.Envelope
		if err := json.Unmarshal(gotBody, &env); err != nil {
			t.Fatalf("invalid envelope json: %v", err)
		}
		if env.EventID != evt.ID || env.EventType != entity.EventTypeLegalEntityCreated {
			t.Fatalf("unexpected envelope payload: %+v", env)
		}
	})

	t.Run("classifies 4xx permanent vs 429/5xx/3xx retryable", func(t *testing.T) {
		t.Parallel()

		cases := []struct {
			name          string
			status        int
			wantPermanent bool
		}{
			{name: "400 bad request is permanent", status: http.StatusBadRequest, wantPermanent: true},
			{name: "422 unprocessable is permanent", status: http.StatusUnprocessableEntity, wantPermanent: true},
			{name: "429 too many requests is retryable", status: http.StatusTooManyRequests, wantPermanent: false},
			{name: "408 request timeout is retryable", status: http.StatusRequestTimeout, wantPermanent: false},
			{name: "503 unavailable is retryable", status: http.StatusServiceUnavailable, wantPermanent: false},
			{name: "302 redirect blocked and retryable", status: http.StatusFound, wantPermanent: false},
		}

		for _, tc := range cases {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if tc.status == http.StatusFound {
						w.Header().Set("Location", "http://127.0.0.1:1/should-not-follow")
					}
					w.WriteHeader(tc.status)
					_, _ = w.Write([]byte("status response"))
				}))
				defer srv.Close()

				pub, err := outbox.NewWebhookPublisher(outbox.WebhookConfig{Endpoint: srv.URL})
				if err != nil {
					t.Fatalf("NewWebhookPublisher failed: %v", err)
				}
				pubErr := pub.Publish(context.Background(), sampleOutboxEvent(t, entity.EventTypePartyCreated, map[string]any{"x": 1}))
				if pubErr == nil {
					t.Fatalf("expected error for HTTP %d", tc.status)
				}
				if got := outbox.IsPermanent(pubErr); got != tc.wantPermanent {
					t.Fatalf("status %d IsPermanent=%v, want %v (err=%v)", tc.status, got, tc.wantPermanent, pubErr)
				}
			})
		}
	})
}
