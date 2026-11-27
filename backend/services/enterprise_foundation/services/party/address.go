package party

import (
	"context"
	"time"

	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/entity"
	eferrors "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/repository"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/services/internal/servicekit"
)

// GetAddressLinkByID mengambil satu tautan alamat milik mitra berdasarkan UUID baris party_addresses.
func (s *Service) GetAddressLinkByID(ctx context.Context, linkID string) (*entity.PartyAddress, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("party_address_id", linkID); err != nil {
		return nil, err
	}
	return s.partyAddresses.SelectByID(ctx, entity.NormalizeUUID(linkID))
}

// ListAddressLinksByParty mengambil tautan alamat yang dimiliki satu mitra eksternal secara berhalaman.
func (s *Service) ListAddressLinksByParty(ctx context.Context, partyID string, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.PartyAddress], error) {
	if err := s.ensureReady(); err != nil {
		return domainrepo.PageResult[*entity.PartyAddress]{}, err
	}
	if err := entity.ValidateUUID("party_id", partyID); err != nil {
		return domainrepo.PageResult[*entity.PartyAddress]{}, err
	}
	normPartyID := entity.NormalizeUUID(partyID)
	if _, err := s.repo.SelectByID(ctx, normPartyID); err != nil {
		return domainrepo.PageResult[*entity.PartyAddress]{}, err
	}
	return s.partyAddresses.SelectByPartyID(ctx, normPartyID, page.Normalize())
}

// ListAddressLinksByPartyAndUsage mengambil daftar tautan alamat mitra untuk kategori peruntukan tertentu secara berhalaman.
func (s *Service) ListAddressLinksByPartyAndUsage(ctx context.Context, partyID string, usage entity.AddressUsageKind, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.PartyAddress], error) {
	if err := s.ensureReady(); err != nil {
		return domainrepo.PageResult[*entity.PartyAddress]{}, err
	}
	if err := entity.ValidateUUID("party_id", partyID); err != nil {
		return domainrepo.PageResult[*entity.PartyAddress]{}, err
	}
	if !usage.IsValid() {
		return domainrepo.PageResult[*entity.PartyAddress]{}, eferrors.WithField(eferrors.ErrInvalidAddressUsage, "usage", "invalid address usage kind")
	}
	normPartyID := entity.NormalizeUUID(partyID)
	if _, err := s.repo.SelectByID(ctx, normPartyID); err != nil {
		return domainrepo.PageResult[*entity.PartyAddress]{}, err
	}
	return s.partyAddresses.SelectByPartyAndUsage(ctx, normPartyID, usage, page.Normalize())
}

// GetPrimaryAddressLink mengambil tautan alamat utama mitra untuk peruntukan tertentu.
func (s *Service) GetPrimaryAddressLink(ctx context.Context, partyID string, usage entity.AddressUsageKind) (*entity.PartyAddress, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("party_id", partyID); err != nil {
		return nil, err
	}
	if !usage.IsValid() {
		return nil, eferrors.WithField(eferrors.ErrInvalidAddressUsage, "usage", "invalid address usage kind")
	}
	normPartyID := entity.NormalizeUUID(partyID)
	if _, err := s.repo.SelectByID(ctx, normPartyID); err != nil {
		return nil, err
	}
	return s.partyAddresses.SelectPrimaryByPartyAndUsage(ctx, normPartyID, usage)
}

// ListAddressLinksByAddressID mengambil daftar tautan mitra yang merujuk satu baris alamat master secara berhalaman.
func (s *Service) ListAddressLinksByAddressID(ctx context.Context, addressID string, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.PartyAddress], error) {
	if err := s.ensureReady(); err != nil {
		return domainrepo.PageResult[*entity.PartyAddress]{}, err
	}
	if err := entity.ValidateUUID("address_id", addressID); err != nil {
		return domainrepo.PageResult[*entity.PartyAddress]{}, err
	}
	normAddrID := entity.NormalizeUUID(addressID)
	if _, err := s.addresses.SelectByID(ctx, normAddrID); err != nil {
		return domainrepo.PageResult[*entity.PartyAddress]{}, eferrors.WithField(err, "address_id", "referenced address not found")
	}
	return s.partyAddresses.SelectByAddressID(ctx, normAddrID, page.Normalize())
}

// ListAddressLinks mengambil daftar halaman relasi alamat mitra sesuai filter mitra, alamat master, peruntukan, dan penanda utama.
func (s *Service) ListAddressLinks(ctx context.Context, filter domainrepo.PartyAddressFilter) (domainrepo.PageResult[*entity.PartyAddress], error) {
	if err := s.ensureReady(); err != nil {
		return domainrepo.PageResult[*entity.PartyAddress]{}, err
	}
	return s.partyAddresses.SelectList(ctx, filter.Normalize())
}

// GetAllAddressLinks mengambil daftar relasi alamat mitra secara berhalaman tanpa filter tambahan.
func (s *Service) GetAllAddressLinks(ctx context.Context, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.PartyAddress], error) {
	if err := s.ensureReady(); err != nil {
		return domainrepo.PageResult[*entity.PartyAddress]{}, err
	}
	return s.partyAddresses.SelectList(ctx, domainrepo.PartyAddressFilter{Page: page.Normalize()})
}

// LinkAddress menautkan alamat master ke mitra sambil mencegah duplikasi (party_id, address_id, usage) dan konflik alamat utama.
func (s *Service) LinkAddress(ctx context.Context, in LinkAddressInput) (*entity.PartyAddress, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	id, err := s.idGen.NewUUID()
	if err != nil {
		return nil, eferrors.Wrap(err, eferrors.KindUnavailable, eferrors.Code("enterprise_foundation.party_address.id_generation_failed"), "failed to generate party address id")
	}

	now := s.clock.Now()
	link, err := entity.NewPartyAddress(id, in.PartyID, in.AddressID, in.Usage, in.IsPrimary, now)
	if err != nil {
		return nil, err
	}

	if err := s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		party, err := s.repo.SelectByID(txCtx, link.PartyID)
		if err != nil {
			return err
		}
		if party.IsDeleted() {
			return eferrors.ErrPartyDeleted
		}
		if err := servicekit.EnsureAddressExists(txCtx, s.addresses, "address_id", &link.AddressID); err != nil {
			return err
		}
		existing, err := s.partyAddresses.SelectByPartyAddressUsage(txCtx, link.PartyID, link.AddressID, link.Usage)
		if err != nil && eferrors.KindOf(err) != eferrors.KindNotFound {
			return err
		}
		if existing != nil {
			return eferrors.ErrPartyAddressExists
		}
		if link.IsPrimary {
			if err := s.resolvePrimaryConflict(txCtx, link.PartyID, link.Usage, link.ID, in.ReplacePrimary, now); err != nil {
				return err
			}
		}
		return s.partyAddresses.Insert(txCtx, link)
	}); err != nil {
		return nil, err
	}
	return link, nil
}

// UpdateAddressLink memperbarui peruntukan dan penanda utama tautan alamat mitra secara aman tanpa duplikasi.
func (s *Service) UpdateAddressLink(ctx context.Context, in UpdateAddressLinkInput) (*entity.PartyAddress, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("party_address_id", in.ID); err != nil {
		return nil, err
	}
	normID := entity.NormalizeUUID(in.ID)

	var updated *entity.PartyAddress
	if err := s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		link, err := s.partyAddresses.SelectByID(txCtx, normID)
		if err != nil {
			return err
		}
		now := s.clock.Now()
		if err := link.UpdateUsage(in.Usage, in.IsPrimary, now); err != nil {
			return err
		}
		existing, err := s.partyAddresses.SelectByPartyAddressUsage(txCtx, link.PartyID, link.AddressID, link.Usage)
		if err != nil && eferrors.KindOf(err) != eferrors.KindNotFound {
			return err
		}
		if existing != nil && existing.ID != link.ID {
			return eferrors.ErrPartyAddressExists
		}
		if link.IsPrimary {
			if err := s.resolvePrimaryConflict(txCtx, link.PartyID, link.Usage, link.ID, in.ReplacePrimary, now); err != nil {
				return err
			}
		}
		if err := s.partyAddresses.Update(txCtx, link); err != nil {
			return err
		}
		updated = link
		return nil
	}); err != nil {
		return nil, err
	}
	return updated, nil
}

// SetPrimaryAddressLink mempromosikan satu tautan alamat menjadi alamat utama untuk kategori peruntukannya secara atomik.
func (s *Service) SetPrimaryAddressLink(ctx context.Context, linkID string) (*entity.PartyAddress, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("party_address_id", linkID); err != nil {
		return nil, err
	}
	normID := entity.NormalizeUUID(linkID)

	var updated *entity.PartyAddress
	if err := s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		link, err := s.partyAddresses.SelectByID(txCtx, normID)
		if err != nil {
			return err
		}
		now := s.clock.Now()
		if err := link.SetPrimary(true, now); err != nil {
			return err
		}
		if err := s.partyAddresses.UpdateClearPrimaryByPartyUsageExcept(txCtx, link.PartyID, link.Usage, link.ID, now); err != nil {
			return err
		}
		if err := s.partyAddresses.Update(txCtx, link); err != nil {
			return err
		}
		updated = link
		return nil
	}); err != nil {
		return nil, err
	}
	return updated, nil
}

// UnlinkAddress menghapus tautan relasi alamat milik mitra tanpa menghapus alamat pada master addresses.
func (s *Service) UnlinkAddress(ctx context.Context, linkID string) error {
	if err := s.ensureReady(); err != nil {
		return err
	}
	if err := entity.ValidateUUID("party_address_id", linkID); err != nil {
		return err
	}
	normID := entity.NormalizeUUID(linkID)
	if _, err := s.partyAddresses.SelectByID(ctx, normID); err != nil {
		return err
	}
	return s.partyAddresses.DeleteByID(ctx, normID)
}

func (s *Service) resolvePrimaryConflict(
	ctx context.Context,
	partyID string,
	usage entity.AddressUsageKind,
	keepID string,
	replacePrimary bool,
	now time.Time,
) error {
	if replacePrimary {
		return s.partyAddresses.UpdateClearPrimaryByPartyUsageExcept(ctx, partyID, usage, keepID, now)
	}
	existingPrimary, err := s.partyAddresses.SelectPrimaryByPartyAndUsage(ctx, partyID, usage)
	if err != nil && eferrors.KindOf(err) != eferrors.KindNotFound {
		return err
	}
	if existingPrimary != nil && existingPrimary.ID != keepID {
		return eferrors.ErrPrimaryPartyAddressConflict
	}
	return nil
}
