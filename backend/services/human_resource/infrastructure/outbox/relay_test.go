package outbox

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
)

type transactionStub struct{}

func (transactionStub) WithinTransaction(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

type deliveryStoreStub struct {
	event       *entity.OutboxEvent
	selectCalls int
	published   bool
	retryCode   string
	attempts    int32
	selectErr   error
}

func (s *deliveryStoreStub) SelectDue(context.Context, time.Time, int, int) ([]entity.OutboxEvent, error) {
	s.selectCalls++
	if s.selectErr != nil {
		return nil, s.selectErr
	}
	if s.event == nil || s.published || s.retryCode != "" {
		return nil, nil
	}
	return []entity.OutboxEvent{*s.event}, nil
}

func (s *deliveryStoreStub) MarkPublished(_ context.Context, _ string, _ time.Time, attempts int32) error {
	s.published, s.attempts = true, attempts
	return nil
}

func (s *deliveryStoreStub) MarkRetry(_ context.Context, _ string, _ time.Time, attempts int32, code string) error {
	s.retryCode, s.attempts = code, attempts
	return nil
}

type publishStub struct {
	err      error
	calls    int
	eventIDs []string
}

func (p *publishStub) Publish(_ context.Context, event *entity.OutboxEvent) error {
	p.calls++
	p.eventIDs = append(p.eventIDs, event.ID)
	return p.err
}

func validEvent() *entity.OutboxEvent {
	payload := entity.JSON(`{"aggregate_id":"7bb89524-2fd5-46cd-a815-fced5c7beab0"}`)
	digest := sha256.Sum256(payload)
	now := time.Now().UTC()
	return &entity.OutboxEvent{
		ID: "cf0d6da1-7f1d-4613-846f-e3dc77ba895b", AggregateType: "hr_jobs",
		AggregateID: "7bb89524-2fd5-46cd-a815-fced5c7beab0", AggregateVersion: 1,
		EventType: "hr_jobs.created", SchemaVersion: 1, Payload: payload, PayloadDigest: digest[:],
		OccurredAt: now, AvailableAt: now,
	}
}

func TestRelayDeliveryRetryAndInvalidDigest(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name         string
		publishError error
		corrupt      bool
		published    bool
		retryCode    string
		attempts     int32
		publishCalls int
	}{
		{"delivered", nil, false, true, "", 1, 1},
		{"retry", errors.New("temporary"), false, false, "delivery_failed", 1, 1},
		{"invalid digest", nil, true, false, "invalid_payload", 3, 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			event := validEvent()
			if test.corrupt {
				event.PayloadDigest[0] ^= 0xff
			}
			store := &deliveryStoreStub{event: event}
			publisher := &publishStub{err: test.publishError}
			relay, err := NewRelay(transactionStub{}, store, publisher, Config{MaxAttempts: 3}, nil)
			if err != nil {
				t.Fatal(err)
			}
			if err := relay.Drain(t.Context()); err != nil {
				t.Fatal(err)
			}
			if store.published != test.published || store.retryCode != test.retryCode ||
				store.attempts != test.attempts || publisher.calls != test.publishCalls {
				t.Fatalf("unexpected delivery state: %+v, calls=%d", store, publisher.calls)
			}
		})
	}
}

func TestStandbyPreservesPendingEvent(t *testing.T) {
	t.Parallel()
	store := &deliveryStoreStub{event: validEvent()}
	relay, err := NewRelay(transactionStub{}, store, nil, Config{Standby: true}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := relay.Drain(t.Context()); err != nil {
		t.Fatal(err)
	}
	if store.selectCalls != 0 || store.published {
		t.Fatal("standby touched pending event")
	}
}

func TestActiveRelayReturnsDatabaseFailure(t *testing.T) {
	t.Parallel()
	want := errors.New("database unavailable")
	store := &deliveryStoreStub{selectErr: want}
	relay, err := NewRelay(transactionStub{}, store, &publishStub{}, Config{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := relay.Run(t.Context()); !errors.Is(err, want) {
		t.Fatalf("relay failure = %v, want database error", err)
	}
}

func TestWorkerTerminationRequiresIAMAcknowledgement(t *testing.T) {
	t.Parallel()
	event := validEvent()
	event.EventType = entity.WorkerTerminationRecorded
	store := &deliveryStoreStub{event: event}
	general := &publishStub{}
	iam := &publishStub{err: errors.New("iam unavailable")}
	relay, err := NewRelay(transactionStub{}, store, RouteWorkerTermination(general, iam), Config{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := relay.Drain(t.Context()); err != nil {
		t.Fatal(err)
	}
	if general.calls != 0 || iam.calls != 1 || store.published || store.retryCode != "delivery_failed" {
		t.Fatal("termination must remain pending after failed IAM delivery")
	}
	store.retryCode = ""
	store.event.AttemptCount = store.attempts
	iam.err = nil
	if err := relay.Drain(t.Context()); err != nil {
		t.Fatal(err)
	}
	if general.calls != 0 || iam.calls != 2 || !store.published || store.attempts != 2 {
		t.Fatal("termination must be published after IAM acknowledges retry")
	}
}

func TestRelayReportsExhaustedTerminationWithoutPayload(t *testing.T) {
	t.Parallel()
	event := validEvent()
	event.EventType = entity.WorkerTerminationRecorded
	event.AttemptCount = 2
	store := &deliveryStoreStub{event: event}
	publisher := &publishStub{err: errors.New("private upstream detail")}
	var logs bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logs, nil))
	relay, err := NewRelay(transactionStub{}, store, publisher, Config{MaxAttempts: 3}, logger)
	if err != nil {
		t.Fatal(err)
	}
	if err := relay.Drain(t.Context()); err != nil {
		t.Fatal(err)
	}
	if store.published || store.attempts != 3 || store.retryCode != "delivery_failed" {
		t.Fatalf("exhausted event state: %+v", store)
	}
	log := logs.String()
	for _, want := range []string{"HRM outbox delivery exhausted", event.ID, entity.WorkerTerminationRecorded, "delivery_failed"} {
		if !strings.Contains(log, want) {
			t.Fatalf("missing recovery detail %q in %q", want, log)
		}
	}
	if strings.Contains(log, event.AggregateID) || strings.Contains(log, "private upstream detail") {
		t.Fatal("outbox warning disclosed event payload or upstream details")
	}
}
