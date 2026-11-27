// Package servicekit menyediakan helper internal bersama bagi paket application service Enterprise Foundation.
package servicekit

import (
	"context"

	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/entity"
	eferrors "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/repository"
	domainservice "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/service"
)

// MaxCollectAllItems membatasi jumlah maksimum baris pada agregasi halaman internal (seperti dedup alamat dan evaluasi role mitra) agar terhindar dari lonjakan memori tak terbatas.
const MaxCollectAllItems = 10000

// ValidateOptionalUUIDList memvalidasi daftar ID UUID dan mengembalikan irisan ternormalisasi unik secara berurutan.
func ValidateOptionalUUIDList(field string, ids []string) ([]string, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	seen := make(map[string]struct{}, len(ids))
	out := make([]string, 0, len(ids))
	for _, raw := range ids {
		if err := entity.ValidateUUID(field, raw); err != nil {
			return nil, err
		}
		norm := entity.NormalizeUUID(raw)
		if _, ok := seen[norm]; ok {
			continue
		}
		seen[norm] = struct{}{}
		out = append(out, norm)
	}
	return out, nil
}

// CollectAllPages menarik seluruh halaman hasil kueri berhalaman sampai tuntas dengan pengaman batas memori.
func CollectAllPages[T any](
	ctx context.Context,
	fetch func(ctx context.Context, page domainrepo.PageRequest) (domainrepo.PageResult[T], error),
) ([]T, error) {
	if fetch == nil {
		return nil, eferrors.Wrap(eferrors.ErrUnavailable, eferrors.KindUnavailable, eferrors.Code("enterprise_foundation.service.nil_page_fetcher"), "page fetcher is required")
	}

	out := make([]T, 0, domainrepo.MaxPageLimit)
	offset := 0
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		page, err := fetch(ctx, domainrepo.PageRequest{
			Limit:  domainrepo.MaxPageLimit,
			Offset: offset,
		})
		if err != nil {
			return nil, err
		}
		if len(page.Items) == 0 {
			break
		}
		if len(out)+len(page.Items) > MaxCollectAllItems {
			return nil, eferrors.New(eferrors.KindConflict, eferrors.Code("enterprise_foundation.service.get_all_limit_exceeded"), "result set exceeds get_all maximum bound")
		}
		out = append(out, page.Items...)
		if !page.HasMore {
			break
		}
		offset += len(page.Items)
	}
	return out, nil
}

// EnsureAddressExists memverifikasi secara fail-closed bahwa alamat referensi tersedia pada master addresses.
func EnsureAddressExists(ctx context.Context, repo domainrepo.AddressRepository, field string, addressID *string) error {
	if addressID == nil {
		return nil
	}
	if repo == nil {
		return eferrors.Wrap(eferrors.ErrUnavailable, eferrors.KindUnavailable, eferrors.Code("enterprise_foundation.service.nil_address_repo"), "address repository is not initialized")
	}
	if _, err := repo.SelectByID(ctx, *addressID); err != nil {
		return eferrors.WithField(err, field, "referenced address not found")
	}
	return nil
}

// EmitOutboxInTx membentuk dan menyimpan event outbox di dalam transaksi aktif yang sama dengan mutasi agregat.
func EmitOutboxInTx(
	ctx context.Context,
	outboxRepo domainrepo.OutboxRepository,
	builder *domainservice.OutboxBuilder,
	eventType, aggregateType, aggregateID string,
	payload any,
) error {
	if outboxRepo == nil {
		return eferrors.Wrap(eferrors.ErrUnavailable, eferrors.KindUnavailable, eferrors.Code("enterprise_foundation.service.nil_outbox_repo"), "outbox repository is not initialized")
	}
	if builder == nil {
		return eferrors.Wrap(eferrors.ErrUnavailable, eferrors.KindUnavailable, eferrors.Code("enterprise_foundation.service.nil_outbox_builder"), "outbox builder is not initialized")
	}
	evt, err := builder.Build(eventType, aggregateType, aggregateID, payload)
	if err != nil {
		return err
	}
	return outboxRepo.Insert(ctx, evt)
}

// ResolveOutboxDependencies memeriksa ketersediaan OutboxRepository dan menyiapkan OutboxBuilder bawaan bila belum disuntikkan.
func ResolveOutboxDependencies(
	outboxRepo domainrepo.OutboxRepository,
	builder *domainservice.OutboxBuilder,
	idGen domainservice.IDGenerator,
	clock domainservice.Clock,
) (*domainservice.OutboxBuilder, error) {
	if outboxRepo == nil {
		return nil, eferrors.NewValidation("outbox", eferrors.Code("enterprise_foundation.service.nil_outbox_repo"), "outbox repository is required")
	}
	if builder != nil {
		return builder, nil
	}
	return domainservice.NewOutboxBuilder(idGen, clock)
}
