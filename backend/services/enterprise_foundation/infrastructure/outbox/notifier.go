package outbox

import (
	"context"
	"time"

	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/entity"
	domainrepo "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/repository"
)

// Notifier menerima sinyal ringan untuk segera membangunkan loop relay outbox setelah transaksi database commit.
type Notifier interface {
	Notify()
}

// AfterCommitFunc mendaftarkan fungsi callback yang dijalankan setelah transaksi database paling luar berhasil commit.
type AfterCommitFunc func(ctx context.Context, fn func(context.Context))

// NotifyingRepository membungkus OutboxRepository untuk mengirim sinyal bangun pasca-commit pada setiap Insert sukses.
type NotifyingRepository struct {
	base        domainrepo.OutboxRepository
	notifier    Notifier
	afterCommit AfterCommitFunc
}

var _ domainrepo.OutboxRepository = (*NotifyingRepository)(nil)

// WrapNotifyingRepository membungkus OutboxRepository dengan notifikasi relay otomatis pasca-commit bila notifier tersedia.
func WrapNotifyingRepository(
	base domainrepo.OutboxRepository,
	notifier Notifier,
	afterCommit AfterCommitFunc,
) domainrepo.OutboxRepository {
	if base == nil || notifier == nil {
		return base
	}
	if afterCommit == nil {
		afterCommit = func(ctx context.Context, fn func(context.Context)) {
			if fn != nil {
				fn(ctx)
			}
		}
	}
	return &NotifyingRepository{
		base:        base,
		notifier:    notifier,
		afterCommit: afterCommit,
	}
}

// Insert menyimpan event outbox dan mendaftarkan wakeup relay non-blocking yang hanya terpicu setelah commit berhasil.
func (r *NotifyingRepository) Insert(ctx context.Context, event *entity.OutboxEvent) error {
	if err := r.base.Insert(ctx, event); err != nil {
		return err
	}
	if r.notifier != nil && r.afterCommit != nil {
		r.afterCommit(ctx, func(_ context.Context) {
			r.notifier.Notify()
		})
	}
	return nil
}

// SelectByID meneruskan kueri pembacaan berdasarkan ID ke repositori dasar.
func (r *NotifyingRepository) SelectByID(ctx context.Context, id string) (*entity.OutboxEvent, error) {
	return r.base.SelectByID(ctx, id)
}

// SelectByAggregate meneruskan kueri daftar event agregat ke repositori dasar.
func (r *NotifyingRepository) SelectByAggregate(ctx context.Context, aggregateType, aggregateID string, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.OutboxEvent], error) {
	return r.base.SelectByAggregate(ctx, aggregateType, aggregateID, page)
}

// SelectDuePending meneruskan pengambilan antrean jatuh tempo ke repositori dasar.
func (r *NotifyingRepository) SelectDuePending(ctx context.Context, now time.Time, limit int) ([]*entity.OutboxEvent, error) {
	return r.base.SelectDuePending(ctx, now, limit)
}

// UpdateEffective meneruskan penandaan status effective ke repositori dasar.
func (r *NotifyingRepository) UpdateEffective(ctx context.Context, id string, processedAt time.Time) error {
	return r.base.UpdateEffective(ctx, id, processedAt)
}

// UpdateRetry meneruskan penjadwalan retry ke repositori dasar.
func (r *NotifyingRepository) UpdateRetry(ctx context.Context, id string, attemptCount int, nextRetryAt time.Time, errMsg string) error {
	return r.base.UpdateRetry(ctx, id, attemptCount, nextRetryAt, errMsg)
}

// UpdateFailed meneruskan penandaan status failed ke repositori dasar.
func (r *NotifyingRepository) UpdateFailed(ctx context.Context, id string, errMsg string) error {
	return r.base.UpdateFailed(ctx, id, errMsg)
}

// CountPending meneruskan penghitungan antrean pending ke repositori dasar.
func (r *NotifyingRepository) CountPending(ctx context.Context) (int64, error) {
	return r.base.CountPending(ctx)
}

// CountFailed meneruskan penghitungan event failed ke repositori dasar.
func (r *NotifyingRepository) CountFailed(ctx context.Context) (int64, error) {
	return r.base.CountFailed(ctx)
}

// DeleteEffectiveBefore meneruskan pembersihan retensi event effective ke repositori dasar.
func (r *NotifyingRepository) DeleteEffectiveBefore(ctx context.Context, cutoff time.Time) (int64, error) {
	return r.base.DeleteEffectiveBefore(ctx, cutoff)
}
