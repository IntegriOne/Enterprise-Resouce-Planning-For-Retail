package outbox

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestWebhookPublishesSignedIdempotentEnvelope(t *testing.T) {
	t.Parallel()
	secret := []byte(strings.Repeat("s", 32))
	called := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called++
		body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
		if err != nil {
			t.Errorf("read body: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		mac := hmac.New(sha256.New, secret)
		_, _ = mac.Write(body)
		want := hex.EncodeToString(mac.Sum(nil))
		if r.Header.Get("X-IG-Signature-SHA256") != want {
			t.Error("invalid HMAC header")
		}
		var envelope Envelope
		if err := json.Unmarshal(body, &envelope); err != nil || envelope.EventID == "" ||
			envelope.IdempotencyKey != envelope.EventID || envelope.AggregateVersion != 1 {
			t.Errorf("invalid envelope: %v", err)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	publisher, err := NewWebhookPublisher(server.URL, secret, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if err := publisher.Publish(t.Context(), validEvent()); err != nil {
		t.Fatalf("publish: %v", err)
	}
	if called != 1 {
		t.Fatalf("HTTP calls = %d, want 1", called)
	}
}

func TestWebhookRejectsUnsuccessfulResponse(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()
	publisher, err := NewWebhookPublisher(server.URL, []byte(strings.Repeat("s", 32)), time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if err := publisher.Publish(t.Context(), validEvent()); err == nil {
		t.Fatal("non-2xx response was accepted")
	}
}
