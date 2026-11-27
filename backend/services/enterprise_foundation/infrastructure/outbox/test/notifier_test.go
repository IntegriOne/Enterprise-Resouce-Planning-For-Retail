package outbox_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/entity"
	domainrepo "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/repository"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/infrastructure/outbox"
)

type spyNotifier struct {
	count atomic.Int32
}

func (n *spyNotifier) Notify() {
	n.count.Add(1)
}

type stubOutboxRepo struct {
	insertErr error
	inserted  []*entity.OutboxEvent
}

func (s *stubOutboxRepo) Insert(_ context.Context, event *entity.OutboxEvent) error {
	if s.insertErr != nil {
		return s.insertErr
	}
	s.inserted = append(s.inserted, event.Clone())
	return nil
}

func (s *stubOutboxRepo) SelectByID(context.Context, string) (*entity.OutboxEvent, error) {
	return nil, nil
}
func (s *stubOutboxRepo) SelectByAggregate(context.Context, string, string, domainrepo.PageRequest) (domainrepo.PageResult[*entity.OutboxEvent], error) {
	return domainrepo.PageResult[*entity.OutboxEvent]{}, nil
}
func (s *stubOutboxRepo) SelectDuePending(context.Context, time.Time, int) ([]*entity.OutboxEvent, error) {
	return nil, nil
}
func (s *stubOutboxRepo) UpdateEffective(context.Context, string, time.Time) error { return nil }
func (s *stubOutboxRepo) UpdateRetry(context.Context, string, int, time.Time, string) error {
	return nil
}
func (s *stubOutboxRepo) UpdateFailed(context.Context, string, string) error { return nil }
func (s *stubOutboxRepo) CountPending(context.Context) (int64, error)        { return 0, nil }
func (s *stubOutboxRepo) CountFailed(context.Context) (int64, error)         { return 0, nil }
func (s *stubOutboxRepo) DeleteEffectiveBefore(context.Context, time.Time) (int64, error) {
	return 0, nil
}

func TestNotifyingRepository(t *testing.T) {
	t.Parallel()

	t.Run("notifies only when after-commit callback runs", func(t *testing.T) {
		t.Parallel()

		base := &stubOutboxRepo{}
		notifier := &spyNotifier{}
		var pendingHooks []func(context.Context)

		wrapped := outbox.WrapNotifyingRepository(base, notifier, func(_ context.Context, fn func(context.Context)) {
			pendingHooks = append(pendingHooks, fn)
		})

		evt := sampleOutboxEvent(t, entity.EventTypeLegalEntityCreated, map[string]any{"code": "ONE"})
		if err := wrapped.Insert(context.Background(), evt); err != nil {
			t.Fatalf("Insert failed: %v", err)
		}
		if notifier.count.Load() != 0 {
			t.Fatalf("expected 0 notifications before commit, got %d", notifier.count.Load())
		}
		if len(pendingHooks) != 1 {
			t.Fatalf("expected 1 registered after-commit hook, got %d", len(pendingHooks))
		}

		// Simulasikan transaksi DB benar-benar commit.
		pendingHooks[0](context.Background())
		if notifier.count.Load() != 1 {
			t.Fatalf("expected 1 notification after commit, got %d", notifier.count.Load())
		}
	})

	t.Run("does not register hook when insert fails", func(t *testing.T) {
		t.Parallel()

		wantErr := errors.New("insert failed")
		base := &stubOutboxRepo{insertErr: wantErr}
		notifier := &spyNotifier{}
		hookCount := 0

		wrapped := outbox.WrapNotifyingRepository(base, notifier, func(_ context.Context, fn func(context.Context)) {
			hookCount++
		})
		evt := sampleOutboxEvent(t, entity.EventTypePartyCreated, map[string]any{"code": "P-1"})
		if err := wrapped.Insert(context.Background(), evt); !errors.Is(err, wantErr) {
			t.Fatalf("got err %v, want %v", err, wantErr)
		}
		if hookCount != 0 || notifier.count.Load() != 0 {
			t.Fatalf("expected zero hooks and zero notifications on insert failure")
		}
	})
}
