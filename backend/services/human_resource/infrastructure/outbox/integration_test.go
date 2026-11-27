//go:build integration

package outbox

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"slices"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	hrdb "github.com/integrione/erp-retail/backend/services/human_resource/db"
	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository/pgstore"
)

func TestDeliveryStoreWithPostgreSQL(t *testing.T) {
	dsn := os.Getenv("HRM_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("HRM_TEST_DATABASE_URL is not configured")
	}
	parsed, err := url.Parse(dsn)
	if err != nil || parsed.Scheme != "postgres" || !strings.HasSuffix(strings.TrimPrefix(parsed.Path, "/"), "_test") {
		t.Fatal("integration test requires a dedicated *_test database")
	}
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal("open HRM test database failed")
	}
	pool, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	if err := hrdb.RunMigrationsUp(ctx, pool); err != nil {
		t.Fatal(err)
	}
	event := validEvent()
	if err := db.WithContext(ctx).Exec("DELETE FROM hr_outbox_events WHERE id = ?", event.ID).Error; err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Exec("DELETE FROM hr_outbox_events WHERE id = ?", event.ID).Error }()
	if err := db.WithContext(ctx).Create(event).Error; err != nil {
		t.Fatal(err)
	}
	store, err := NewStore(db)
	if err != nil {
		t.Fatal(err)
	}
	tx, err := pgstore.NewTransactor(db)
	if err != nil {
		t.Fatal(err)
	}
	locked := make(chan struct{})
	release := make(chan struct{})
	firstDone := make(chan error, 1)
	go func() {
		firstDone <- tx.WithinTransaction(ctx, func(tc context.Context) error {
			items, err := store.SelectDue(tc, time.Now().UTC().Add(time.Second), 1, 20)
			if err != nil {
				return err
			}
			if len(items) != 1 || items[0].ID != event.ID {
				return fmt.Errorf("first claimant got %d events", len(items))
			}
			if err := validatePayload(&items[0]); err != nil {
				return fmt.Errorf("verify event digest after JSONB round trip: %w", err)
			}
			close(locked)
			select {
			case <-release:
			case <-tc.Done():
				return tc.Err()
			}
			return store.MarkPublished(tc, event.ID, time.Now().UTC(), 1)
		})
	}()
	select {
	case <-locked:
	case err := <-firstDone:
		t.Fatalf("first claimant failed: %v", err)
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
	secondErr := tx.WithinTransaction(ctx, func(tc context.Context) error {
		items, err := store.SelectDue(tc, time.Now().UTC().Add(time.Second), 1, 20)
		if err != nil {
			return err
		}
		if len(items) != 0 {
			return fmt.Errorf("second claimant received %d locked events", len(items))
		}
		return nil
	})
	close(release)
	firstErr := <-firstDone
	if secondErr != nil {
		t.Fatal(secondErr)
	}
	if firstErr != nil {
		t.Fatal(firstErr)
	}
	var published time.Time
	if err := db.WithContext(ctx).Raw("SELECT published_at FROM hr_outbox_events WHERE id = ?", event.ID).Scan(&published).Error; err != nil {
		t.Fatal(err)
	}
	if published.IsZero() {
		t.Fatal("published event was not persisted")
	}

	retryEvent := validEvent()
	retryEvent.ID = "cf0d6da1-7f1d-4613-846f-e3dc77ba895c"
	retryEvent.AggregateVersion = 2
	if err := db.WithContext(ctx).Create(retryEvent).Error; err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Exec("DELETE FROM hr_outbox_events WHERE id = ?", retryEvent.ID).Error }()

	secret := []byte(strings.Repeat("s", 32))
	var attempts atomic.Int32
	bodies := make(chan []byte, 2)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
		if err != nil {
			t.Errorf("read webhook body: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		mac := hmac.New(sha256.New, secret)
		_, _ = mac.Write(body)
		if r.Header.Get("X-IG-Signature-SHA256") != hex.EncodeToString(mac.Sum(nil)) {
			t.Error("invalid webhook signature")
		}
		var envelope Envelope
		if err := json.Unmarshal(body, &envelope); err != nil ||
			envelope.EventID != retryEvent.ID || envelope.IdempotencyKey != retryEvent.ID {
			t.Errorf("unexpected webhook envelope: %+v, decode error: %v", envelope, err)
		}
		select {
		case bodies <- body:
		default:
			t.Error("unexpected extra webhook delivery")
		}
		if attempts.Add(1) == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	publisher, err := NewWebhookPublisher(server.URL, secret, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	relay, err := NewRelay(tx, store, publisher, Config{MaxAttempts: 3}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := relay.Drain(ctx); err != nil {
		t.Fatal(err)
	}
	var pending entity.OutboxEvent
	if err := db.WithContext(ctx).First(&pending, "id = ?", retryEvent.ID).Error; err != nil {
		t.Fatal(err)
	}
	if pending.AttemptCount != 1 || pending.PublishedAt != nil ||
		pending.LastErrorCode == nil || *pending.LastErrorCode != "delivery_failed" ||
		!pending.AvailableAt.After(retryEvent.AvailableAt) {
		t.Fatalf("unexpected retry state: %+v", pending)
	}
	if err := db.WithContext(ctx).Model(&entity.OutboxEvent{}).Where("id = ?", retryEvent.ID).
		Update("available_at", time.Now().UTC().Add(-time.Second)).Error; err != nil {
		t.Fatal(err)
	}
	if err := relay.Drain(ctx); err != nil {
		t.Fatal(err)
	}
	var delivered entity.OutboxEvent
	if err := db.WithContext(ctx).First(&delivered, "id = ?", retryEvent.ID).Error; err != nil {
		t.Fatal(err)
	}
	if delivered.AttemptCount != 2 || delivered.PublishedAt == nil || delivered.LastErrorCode != nil {
		t.Fatalf("unexpected delivery state: %+v", delivered)
	}
	if attempts.Load() != 2 || len(bodies) != 2 {
		t.Fatalf("webhook attempts = %d, want 2", attempts.Load())
	}
	if !bytes.Equal(<-bodies, <-bodies) {
		t.Fatal("webhook body changed between delivery attempts")
	}
	checkTerminationDeliveryRouting(t, ctx, db, tx)
}

func checkTerminationDeliveryRouting(t *testing.T, ctx context.Context, db *gorm.DB, tx *pgstore.Transactor) {
	t.Helper()
	general := validEvent()
	general.ID = "cf0d6da1-7f1d-4613-846f-e3dc77ba895d"
	general.AggregateVersion = 3
	termination := validEvent()
	termination.ID = "cf0d6da1-7f1d-4613-846f-e3dc77ba895e"
	termination.EventType = entity.WorkerTerminationRecorded
	termination.AggregateType = "hr_worker_terminations"
	for _, event := range []*entity.OutboxEvent{general, termination} {
		if err := db.WithContext(ctx).Create(event).Error; err != nil {
			t.Fatal(err)
		}
		defer func() { _ = db.Exec("DELETE FROM hr_outbox_events WHERE id = ?", event.ID).Error }()
	}
	for _, tc := range []struct {
		name                 string
		general, termination bool
		want                 []string
	}{
		{"umum", true, false, []string{general.ID}},
		{"terminasi saja", false, true, []string{termination.ID}},
		{"keduanya", true, true, []string{general.ID, termination.ID}},
		{"standby", false, false, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			store, err := NewStoreForRoutes(db, tc.general, tc.termination)
			if err != nil {
				t.Fatal(err)
			}
			err = tx.WithinTransaction(ctx, func(tcCtx context.Context) error {
				rows, err := store.SelectDue(tcCtx, time.Now().UTC(), 10, 20)
				if err != nil {
					return err
				}
				got := make([]string, 0, len(rows))
				for _, row := range rows {
					got = append(got, row.ID)
				}
				if !slices.Equal(got, tc.want) {
					return fmt.Errorf("selected %v, want %v", got, tc.want)
				}
				return nil
			})
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}
