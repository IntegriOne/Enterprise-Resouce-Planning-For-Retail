package outbox_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/entity"
	eferrors "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/repository"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/infrastructure/outbox"
)

type fakeRelayClock struct {
	mu  sync.Mutex
	now time.Time
}

func newFakeRelayClock(start time.Time) *fakeRelayClock {
	return &fakeRelayClock{now: start}
}

func (c *fakeRelayClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *fakeRelayClock) Advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(d)
}

type memoryOutboxStore struct {
	mu     sync.Mutex
	events map[string]*entity.OutboxEvent
}

func newMemoryOutboxStore() *memoryOutboxStore {
	return &memoryOutboxStore{events: make(map[string]*entity.OutboxEvent)}
}

func (m *memoryOutboxStore) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

func (m *memoryOutboxStore) Insert(_ context.Context, ev *entity.OutboxEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events[ev.ID] = ev.Clone()
	return nil
}

func (m *memoryOutboxStore) SelectByID(_ context.Context, id string) (*entity.OutboxEvent, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	ev, ok := m.events[id]
	if !ok {
		return nil, eferrors.ErrOutboxEventNotFound
	}
	return ev.Clone(), nil
}

func (m *memoryOutboxStore) SelectByAggregate(context.Context, string, string, domainrepo.PageRequest) (domainrepo.PageResult[*entity.OutboxEvent], error) {
	return domainrepo.PageResult[*entity.OutboxEvent]{}, nil
}

func (m *memoryOutboxStore) SelectDuePending(_ context.Context, now time.Time, limit int) ([]*entity.OutboxEvent, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	due := make([]*entity.OutboxEvent, 0, len(m.events))
	for _, ev := range m.events {
		if ev.Status == entity.OutboxStatusPending && !ev.NextRetryAt.After(now) {
			due = append(due, ev.Clone())
		}
	}
	sort.Slice(due, func(i, j int) bool {
		if !due[i].NextRetryAt.Equal(due[j].NextRetryAt) {
			return due[i].NextRetryAt.Before(due[j].NextRetryAt)
		}
		if !due[i].CreatedAt.Equal(due[j].CreatedAt) {
			return due[i].CreatedAt.Before(due[j].CreatedAt)
		}
		return due[i].ID < due[j].ID
	})
	if limit > 0 && len(due) > limit {
		due = due[:limit]
	}
	return due, nil
}

func (m *memoryOutboxStore) UpdateEffective(_ context.Context, id string, processedAt time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	ev, ok := m.events[id]
	if !ok {
		return eferrors.ErrOutboxEventNotFound
	}
	return ev.MarkEffective(processedAt)
}

func (m *memoryOutboxStore) UpdateRetry(_ context.Context, id string, attemptCount int, nextRetryAt time.Time, errMsg string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	ev, ok := m.events[id]
	if !ok {
		return eferrors.ErrOutboxEventNotFound
	}
	if err := ev.ScheduleRetry(nextRetryAt, errMsg, nextRetryAt); err != nil {
		return err
	}
	ev.AttemptCount = attemptCount
	return nil
}

func (m *memoryOutboxStore) UpdateFailed(_ context.Context, id string, errMsg string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	ev, ok := m.events[id]
	if !ok {
		return eferrors.ErrOutboxEventNotFound
	}
	ev.Status = entity.OutboxStatusFailed
	ev.LastError = &errMsg
	return nil
}

func (m *memoryOutboxStore) CountPending(context.Context) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var n int64
	for _, ev := range m.events {
		if ev.Status == entity.OutboxStatusPending {
			n++
		}
	}
	return n, nil
}

func (m *memoryOutboxStore) CountFailed(context.Context) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var n int64
	for _, ev := range m.events {
		if ev.Status == entity.OutboxStatusFailed {
			n++
		}
	}
	return n, nil
}

func (m *memoryOutboxStore) DeleteEffectiveBefore(_ context.Context, cutoff time.Time) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var deleted int64
	for id, ev := range m.events {
		if ev.Status == entity.OutboxStatusEffective && ev.ProcessedAt != nil && ev.ProcessedAt.Before(cutoff) {
			delete(m.events, id)
			deleted++
		}
	}
	return deleted, nil
}

func makeTestOutboxEvent(t *testing.T, id string, now time.Time) *entity.OutboxEvent {
	t.Helper()
	ev, err := entity.NewOutboxEvent(
		id,
		entity.EventTypeLegalEntityCreated,
		entity.AggregateTypeLegalEntity,
		"99999999-2222-4333-8444-555555555555",
		map[string]any{"entity_code": "ENT-1"},
		now,
	)
	if err != nil {
		t.Fatalf("NewOutboxEvent failed: %v", err)
	}
	return ev
}

func TestNewRelayValidation(t *testing.T) {
	t.Parallel()

	store := newMemoryOutboxStore()
	if _, err := outbox.NewRelay(outbox.Dependencies{}, outbox.DefaultConfig()); !errors.Is(err, outbox.ErrRelayNotInitialized) {
		t.Fatalf("expected ErrRelayNotInitialized, got %v", err)
	}
	if _, err := outbox.NewRelay(outbox.Dependencies{Transactor: store, Repo: store}, outbox.DefaultConfig()); !errors.Is(err, outbox.ErrPublisherUnavailable) {
		t.Fatalf("expected ErrPublisherUnavailable without publisher, got %v", err)
	}
	relay, err := outbox.NewRelay(outbox.Dependencies{Transactor: store, Repo: store}, outbox.Config{Standby: true})
	if err != nil {
		t.Fatalf("expected standby relay without publisher to succeed, got %v", err)
	}
	if relay.Mode() != outbox.ModeStandby {
		t.Fatalf("got mode %q, want %q", relay.Mode(), outbox.ModeStandby)
	}
}

func TestRelayRetryBackoffDeadLetterAndRecovery(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, 11, 15, 21, 0, 0, 0, time.UTC)
	clk := newFakeRelayClock(start)
	store := newMemoryOutboxStore()
	discardLogger := slog.New(slog.NewTextHandler(io.Discard, nil))

	evTransientID := "11111111-1111-4111-8111-111111111111"
	evPermanentID := "22222222-2222-4222-8222-222222222222"
	_ = store.Insert(context.Background(), makeTestOutboxEvent(t, evTransientID, start))
	_ = store.Insert(context.Background(), makeTestOutboxEvent(t, evPermanentID, start))

	var attemptTransient int
	pub := outbox.PublishFunc(func(_ context.Context, ev *entity.OutboxEvent) error {
		if ev.ID == evPermanentID {
			return outbox.MarkPermanent(errors.New("schema contract mismatch"))
		}
		attemptTransient++
		if attemptTransient < 3 {
			return errors.New("temporary downstream timeout")
		}
		return nil
	})

	relay, err := outbox.NewRelay(
		outbox.Dependencies{
			Transactor: store,
			Repo:       store,
			Publisher:  pub,
			Clock:      clk,
			Logger:     discardLogger,
		},
		outbox.Config{
			MaxAttempts: 5,
			BatchSize:   10,
			Backoff: outbox.BackoffPolicy{
				Base:        2 * time.Second,
				Max:         16 * time.Second,
				JitterRatio: 0,
			},
		},
	)
	if err != nil {
		t.Fatalf("NewRelay failed: %v", err)
	}

	// Putaran 1: evPermanent langsung failed; evTransient retry (attempt 1, +2s).
	stats, err := relay.DrainPending(context.Background())
	if err != nil {
		t.Fatalf("DrainPending #1 failed: %v", err)
	}
	if stats.Claimed != 2 || stats.Retried != 1 || stats.Failed != 1 || stats.Published != 0 {
		t.Fatalf("unexpected stats #1: %+v", stats)
	}

	permRow, _ := store.SelectByID(context.Background(), evPermanentID)
	if permRow.Status != entity.OutboxStatusFailed {
		t.Fatalf("expected permanent event failed, got %s", permRow.Status)
	}

	// Sebelum backoff 2s habis, drain berikutnya tidak boleh mengklaim ulang event retry.
	clk.Advance(1 * time.Second)
	statsEarly, err := relay.DrainPending(context.Background())
	if err != nil || statsEarly.Claimed != 0 {
		t.Fatalf("expected 0 claimed before backoff deadline, got stats=%+v err=%v", statsEarly, err)
	}

	// Setelah +2s total, percobaan 2 masih gagal -> retry backoff kedua = 4s.
	clk.Advance(1 * time.Second)
	stats2, err := relay.DrainPending(context.Background())
	if err != nil || stats2.Retried != 1 {
		t.Fatalf("expected second retry, got stats=%+v err=%v", stats2, err)
	}

	// Setelah +4s lagi, percobaan 3 sukses -> effective.
	clk.Advance(4 * time.Second)
	stats3, err := relay.DrainPending(context.Background())
	if err != nil || stats3.Published != 1 {
		t.Fatalf("expected third attempt to succeed, got stats=%+v err=%v", stats3, err)
	}

	finalRow, _ := store.SelectByID(context.Background(), evTransientID)
	if finalRow.Status != entity.OutboxStatusEffective || finalRow.LastError != nil || finalRow.ProcessedAt == nil {
		t.Fatalf("expected transient event to become effective cleanly, got %+v", finalRow)
	}
}

func TestRelayMaxAttemptsAndStandbyPreservation(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, 11, 15, 21, 0, 0, 0, time.UTC)
	clk := newFakeRelayClock(start)
	store := newMemoryOutboxStore()
	discardLogger := slog.New(slog.NewTextHandler(io.Discard, nil))

	evID := "33333333-3333-4333-8333-333333333333"
	_ = store.Insert(context.Background(), makeTestOutboxEvent(t, evID, start))

	// Mode standby harus membiarkan event pending tetap utuh.
	standbyRelay, err := outbox.NewRelay(
		outbox.Dependencies{Transactor: store, Repo: store, Clock: clk, Logger: discardLogger},
		outbox.Config{Standby: true},
	)
	if err != nil {
		t.Fatalf("NewRelay standby failed: %v", err)
	}
	if _, err := standbyRelay.DrainPending(context.Background()); err != nil {
		t.Fatalf("standby DrainPending failed: %v", err)
	}
	row, _ := store.SelectByID(context.Background(), evID)
	if row.Status != entity.OutboxStatusPending || row.AttemptCount != 0 {
		t.Fatalf("standby mode modified pending event: %+v", row)
	}

	// Mode aktif dengan MaxAttempts=2 menandai failed pada percobaan gagal ke-2.
	activeRelay, err := outbox.NewRelay(
		outbox.Dependencies{
			Transactor: store,
			Repo:       store,
			Clock:      clk,
			Logger:     discardLogger,
			Publisher: outbox.PublishFunc(func(context.Context, *entity.OutboxEvent) error {
				return errors.New("broker down")
			}),
		},
		outbox.Config{
			MaxAttempts: 2,
			Backoff:     outbox.BackoffPolicy{Base: time.Second, Max: 10 * time.Second},
		},
	)
	if err != nil {
		t.Fatalf("NewRelay active failed: %v", err)
	}
	if _, err := activeRelay.DrainPending(context.Background()); err != nil {
		t.Fatalf("attempt 1 failed: %v", err)
	}
	clk.Advance(2 * time.Second)
	stats2, err := activeRelay.DrainPending(context.Background())
	if err != nil || stats2.Failed != 1 {
		t.Fatalf("expected failed dead-letter on max attempts, got stats=%+v err=%v", stats2, err)
	}

	summary, err := activeRelay.Stats(context.Background())
	if err != nil {
		t.Fatalf("Stats failed: %v", err)
	}
	if summary.PendingCount != 0 || summary.FailedCount != 1 {
		t.Fatalf("unexpected Stats summary: %+v", summary)
	}
}

func TestRelayGracefulShutdownMidBatchDoesNotBurnAttempts(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, 11, 15, 21, 0, 0, 0, time.UTC)
	clk := newFakeRelayClock(start)
	store := newMemoryOutboxStore()
	discardLogger := slog.New(slog.NewTextHandler(io.Discard, nil))

	ev1 := makeTestOutboxEvent(t, "44444444-4444-4444-8444-444444444441", start)
	ev2 := makeTestOutboxEvent(t, "44444444-4444-4444-8444-444444444442", start.Add(time.Millisecond))
	_ = store.Insert(context.Background(), ev1)
	_ = store.Insert(context.Background(), ev2)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var calls atomic.Int32
	relay, err := outbox.NewRelay(
		outbox.Dependencies{
			Transactor: store,
			Repo:       store,
			Clock:      clk,
			Logger:     discardLogger,
			Publisher: outbox.PublishFunc(func(pCtx context.Context, ev *entity.OutboxEvent) error {
				if calls.Add(1) == 1 {
					return nil
				}
				// Batalkan parent context saat event kedua sedang dikirim.
				cancel()
				<-pCtx.Done()
				return pCtx.Err()
			}),
		},
		outbox.Config{BatchSize: 10, PublishTimeout: 5 * time.Second},
	)
	if err != nil {
		t.Fatalf("NewRelay failed: %v", err)
	}

	stats, err := relay.DrainBatch(ctx)
	if err != nil {
		t.Fatalf("DrainBatch should commit already published event cleanly on shutdown: %v", err)
	}
	if stats.Published != 1 {
		t.Fatalf("expected 1 published before shutdown cancellation, got %+v", stats)
	}

	r1, _ := store.SelectByID(context.Background(), ev1.ID)
	r2, _ := store.SelectByID(context.Background(), ev2.ID)
	if r1.Status != entity.OutboxStatusEffective {
		t.Fatalf("expected first event effective, got %s", r1.Status)
	}
	if r2.Status != entity.OutboxStatusPending || r2.AttemptCount != 0 {
		t.Fatalf("expected aborted second event untouched (pending, attempt 0), got %+v", r2)
	}
}

func TestRelayRunImmediateWakeupAndCleanup(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, 11, 15, 21, 0, 0, 0, time.UTC)
	clk := newFakeRelayClock(start)
	store := newMemoryOutboxStore()
	discardLogger := slog.New(slog.NewTextHandler(io.Discard, nil))

	delivered := make(chan string, 4)
	relay, err := outbox.NewRelay(
		outbox.Dependencies{
			Transactor: store,
			Repo:       store,
			Clock:      clk,
			Logger:     discardLogger,
			Publisher: outbox.PublishFunc(func(_ context.Context, ev *entity.OutboxEvent) error {
				delivered <- ev.ID
				return nil
			}),
		},
		outbox.Config{
			PollInterval:    10 * time.Second,
			CleanupInterval: 1 * time.Hour,
			RetentionPeriod: 24 * time.Hour,
		},
	)
	if err != nil {
		t.Fatalf("NewRelay failed: %v", err)
	}

	runCtx, cancelRun := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() {
		errCh <- relay.Run(runCtx)
	}()

	// Pastikan loop kedua yang berjalan bersamaan ditolak.
	time.Sleep(20 * time.Millisecond)
	if err := relay.Run(runCtx); err == nil {
		t.Fatal("expected concurrent Run() call to fail")
	}

	evID := "55555555-5555-4555-8555-555555555555"
	_ = store.Insert(context.Background(), makeTestOutboxEvent(t, evID, start))
	relay.Notify()

	select {
	case got := <-delivered:
		if got != evID {
			t.Fatalf("got delivered %q, want %q", got, evID)
		}
	case <-time.After(1 * time.Second):
		cancelRun()
		t.Fatal("timed out waiting for immediate post-notify delivery")
	}

	// Uji pembersihan retensi untuk event effective yang melewati batas RetentionPeriod.
	clk.Advance(48 * time.Hour)
	deleted, err := relay.CleanupEffective(context.Background())
	if err != nil {
		cancelRun()
		t.Fatalf("CleanupEffective failed: %v", err)
	}
	if deleted != 1 {
		cancelRun()
		t.Fatalf("expected 1 expired event deleted, got %d", deleted)
	}

	cancelRun()
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("Run exited with unexpected error: %v", err)
		}
	case <-time.After(1 * time.Second):
		t.Fatal(fmt.Sprintf("timed out waiting for Run() graceful exit"))
	}
}
