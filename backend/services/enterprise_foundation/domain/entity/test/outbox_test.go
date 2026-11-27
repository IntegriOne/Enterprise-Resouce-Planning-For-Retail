package entity_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/entity"
	eferrors "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/errors"
)

func TestNewOutboxEvent(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 11, 15, 13, 0, 0, 0, time.UTC)

	t.Run("nil payload defaults to empty json object", func(t *testing.T) {
		t.Parallel()
		ev, err := entity.NewOutboxEvent(testUUID1, entity.EventTypePartyCreated, entity.AggregateTypeParty, testUUID2, nil, now)
		if err != nil {
			t.Fatalf("NewOutboxEvent(nil) failed: %v", err)
		}
		if string(ev.Payload) != "{}" {
			t.Errorf("Payload = %q, want {}", string(ev.Payload))
		}
		if ev.Status != entity.OutboxStatusPending || ev.AttemptCount != 0 {
			t.Errorf("unexpected initial status/attempts: %+v", ev)
		}
	})

	t.Run("struct payload marshals to valid json object", func(t *testing.T) {
		t.Parallel()
		payload := struct {
			PartyID string `json:"party_id"`
			Status  string `json:"status"`
		}{
			PartyID: testUUID2,
			Status:  "active",
		}
		ev, err := entity.NewOutboxEvent(testUUID1, entity.EventTypePartyStatusChanged, entity.AggregateTypeParty, testUUID2, payload, now)
		if err != nil {
			t.Fatalf("NewOutboxEvent(struct) failed: %v", err)
		}
		copied := ev.PayloadCopy()
		if !bytes.Equal(copied, ev.Payload) {
			t.Errorf("PayloadCopy() mismatch")
		}
		// Mutasi pada salinan tidak boleh memengaruhi slice internal.
		copied[0] = 'X'
		if ev.Payload[0] == 'X' {
			t.Errorf("PayloadCopy() leaked internal slice reference")
		}
	})

	t.Run("raw json bytes and json.RawMessage", func(t *testing.T) {
		t.Parallel()
		if _, err := entity.NewOutboxEvent(testUUID1, entity.EventTypePartyUpdated, entity.AggregateTypeParty, testUUID2, []byte(`{"ok":true}`), now); err != nil {
			t.Errorf("[]byte object payload failed: %v", err)
		}
		if _, err := entity.NewOutboxEvent(testUUID1, entity.EventTypePartyUpdated, entity.AggregateTypeParty, testUUID2, json.RawMessage(`{"ok":true}`), now); err != nil {
			t.Errorf("json.RawMessage object payload failed: %v", err)
		}
		if _, err := entity.NewOutboxEvent(testUUID1, entity.EventTypePartyUpdated, entity.AggregateTypeParty, testUUID2, []byte(`[1,2,3]`), now); !errors.Is(err, eferrors.ErrInvalidOutboxPayload) {
			t.Errorf("array payload should fail with ErrInvalidOutboxPayload, got %v", err)
		}
		if _, err := entity.NewOutboxEvent(testUUID1, entity.EventTypePartyUpdated, entity.AggregateTypeParty, testUUID2, make(chan int), now); eferrors.KindOf(err) != eferrors.KindInvalidInput {
			t.Errorf("unmarshalable struct err Kind = %v, want KindInvalidInput", eferrors.KindOf(err))
		}
	})

	t.Run("lifecycle retry truncation and effective", func(t *testing.T) {
		t.Parallel()
		ev, err := entity.NewOutboxEvent(testUUID1, entity.EventTypeLocationCreated, entity.AggregateTypeLocation, testUUID2, nil, now)
		if err != nil {
			t.Fatalf("setup failed: %v", err)
		}

		// Waktu retry di masa lalu harus dijepit minimal sama dengan now.
		longErr := strings.Repeat("é", 1200)
		retryNow := now.Add(time.Minute)
		if err := ev.ScheduleRetry(now, longErr, retryNow); err != nil {
			t.Fatalf("ScheduleRetry() failed: %v", err)
		}
		if ev.AttemptCount != 1 || !ev.NextRetryAt.Equal(retryNow) {
			t.Errorf("AttemptCount=%d NextRetryAt=%v, want 1 and %v", ev.AttemptCount, ev.NextRetryAt, retryNow)
		}
		if ev.LastError == nil || utf8.RuneCountInString(*ev.LastError) != 1000 {
			t.Errorf("expected LastError clamped to 1000 runes, got %v", ev.LastError)
		}

		processedAt := retryNow.Add(time.Minute)
		if err := ev.MarkEffective(processedAt); err != nil {
			t.Fatalf("MarkEffective() failed: %v", err)
		}
		if ev.Status != entity.OutboxStatusEffective || ev.LastError != nil || ev.ProcessedAt == nil {
			t.Errorf("unexpected state after MarkEffective: %+v", ev)
		}

		if err := ev.MarkFailed("   fatal upstream 502   ", processedAt.Add(time.Minute)); err != nil {
			t.Fatalf("MarkFailed() failed: %v", err)
		}
		if ev.Status != entity.OutboxStatusFailed || ev.AttemptCount != 2 || *ev.LastError != "fatal upstream 502" {
			t.Errorf("unexpected state after MarkFailed: %+v", ev)
		}

		clone := ev.Clone()
		if clone == nil || clone == ev || clone.Status != ev.Status {
			t.Errorf("Clone() failed")
		}

		var nilEv *entity.OutboxEvent
		if nilEv.Clone() != nil {
			t.Errorf("nil Clone() want nil")
		}
		if string(nilEv.PayloadCopy()) != "{}" {
			t.Errorf("nil PayloadCopy() want {}")
		}
		if err := nilEv.MarkEffective(now); !errors.Is(err, eferrors.ErrOutboxEventNotFound) {
			t.Errorf("nil MarkEffective() = %v, want ErrOutboxEventNotFound", err)
		}
		if err := nilEv.ScheduleRetry(now, "x", now); !errors.Is(err, eferrors.ErrOutboxEventNotFound) {
			t.Errorf("nil ScheduleRetry() = %v, want ErrOutboxEventNotFound", err)
		}
		if err := nilEv.MarkFailed("x", now); !errors.Is(err, eferrors.ErrOutboxEventNotFound) {
			t.Errorf("nil MarkFailed() = %v, want ErrOutboxEventNotFound", err)
		}
	})
}
