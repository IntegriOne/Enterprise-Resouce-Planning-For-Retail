// Package party mengimplementasikan logika bisnis pengelolaan mitra eksternal, peran bisnis, PIC portal IAM, dan alamat mitra.
package party

import (
	"context"
	"strings"
	"time"

	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/entity"
	eferrors "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/repository"
	domainservice "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/service"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/services/internal/servicekit"
)

// Service mengelola agregat Party beserta sub-entitas PartyRole, PartyContact, dan PartyAddress secara transaksional.
type Service struct {
	tx             domainrepo.Transactor
	repo           domainrepo.PartyRepository
	roles          domainrepo.PartyRoleRepository
	contacts       domainrepo.PartyContactRepository
	partyAddresses domainrepo.PartyAddressRepository
	addresses      domainrepo.AddressRepository
	outbox         domainrepo.OutboxRepository
	outboxBuilder  *domainservice.OutboxBuilder
	scope          domainservice.ScopePolicy
	portal         domainservice.PortalAccessPolicy
	clock          domainservice.Clock
	idGen          domainservice.IDGenerator
}

// NewService membentuk layanan mitra eksternal dengan pemeriksaan fail-closed atas seluruh repositori dan generator ID.
func NewService(deps Dependencies) (*Service, error) {
	if deps.Transactor == nil {
		return nil, eferrors.NewValidation("transactor", eferrors.Code("enterprise_foundation.service.nil_transactor"), "transactor is required")
	}
	if deps.Parties == nil {
		return nil, eferrors.NewValidation("parties", eferrors.Code("enterprise_foundation.service.nil_party_repo"), "party repository is required")
	}
	if deps.PartyRoles == nil {
		return nil, eferrors.NewValidation("party_roles", eferrors.Code("enterprise_foundation.service.nil_party_role_repo"), "party role repository is required")
	}
	if deps.PartyContacts == nil {
		return nil, eferrors.NewValidation("party_contacts", eferrors.Code("enterprise_foundation.service.nil_party_contact_repo"), "party contact repository is required")
	}
	if deps.PartyAddresses == nil {
		return nil, eferrors.NewValidation("party_addresses", eferrors.Code("enterprise_foundation.service.nil_party_address_repo"), "party address repository is required")
	}
	if deps.Addresses == nil {
		return nil, eferrors.NewValidation("addresses", eferrors.Code("enterprise_foundation.service.nil_address_repo"), "address repository is required")
	}
	if deps.IDGen == nil {
		return nil, eferrors.NewValidation("id_gen", eferrors.Code("enterprise_foundation.service.nil_id_generator"), "id generator is required")
	}

	clock := deps.Clock
	if clock == nil {
		clock = domainservice.SystemClock{}
	}
	outboxBuilder, err := servicekit.ResolveOutboxDependencies(deps.Outbox, deps.OutboxBuilder, deps.IDGen, clock)
	if err != nil {
		return nil, err
	}

	return &Service{
		tx:             deps.Transactor,
		repo:           deps.Parties,
		roles:          deps.PartyRoles,
		contacts:       deps.PartyContacts,
		partyAddresses: deps.PartyAddresses,
		addresses:      deps.Addresses,
		outbox:         deps.Outbox,
		outboxBuilder:  outboxBuilder,
		scope:          domainservice.NewScopePolicy(),
		portal:         domainservice.NewPortalAccessPolicy(),
		clock:          clock,
		idGen:          deps.IDGen,
	}, nil
}

// GetByID mengambil detail mitra eksternal berdasarkan UUID.
func (s *Service) GetByID(ctx context.Context, id string) (*entity.Party, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("party_id", id); err != nil {
		return nil, err
	}
	return s.repo.SelectByID(ctx, entity.NormalizeUUID(id))
}

// GetByIDs mengambil beberapa mitra eksternal sekaligus berdasarkan daftar UUID.
func (s *Service) GetByIDs(ctx context.Context, ids []string) ([]*entity.Party, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	normIDs, err := servicekit.ValidateOptionalUUIDList("party_ids", ids)
	if err != nil {
		return nil, err
	}
	if len(normIDs) == 0 {
		return []*entity.Party{}, nil
	}
	return s.repo.SelectByIDs(ctx, normIDs)
}

// GetByCode mencari mitra eksternal berdasarkan partner_code ternormalisasi.
func (s *Service) GetByCode(ctx context.Context, partnerCode string) (*entity.Party, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	code := entity.NormalizeCode(partnerCode)
	if code == "" {
		return nil, eferrors.WithField(eferrors.ErrInvalidPartyCode, "partner_code", "party partner code is required")
	}
	return s.repo.SelectByCode(ctx, code)
}

// GetByTaxIDNumber mencari mitra eksternal berdasarkan NPWP/Tax ID.
func (s *Service) GetByTaxIDNumber(ctx context.Context, taxIDNumber string) (*entity.Party, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	trimmed := strings.TrimSpace(taxIDNumber)
	if trimmed == "" {
		return nil, eferrors.WithField(eferrors.ErrInvalidInput, "tax_id_number", "tax id number is required")
	}
	return s.repo.SelectByTaxIDNumber(ctx, trimmed)
}

// ListByEmail mencari mitra eksternal yang menggunakan alamat email korespondensi tertentu secara berhalaman.
func (s *Service) ListByEmail(ctx context.Context, email string, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.Party], error) {
	if err := s.ensureReady(); err != nil {
		return domainrepo.PageResult[*entity.Party]{}, err
	}
	normalized := strings.ToLower(strings.TrimSpace(email))
	if normalized == "" {
		return domainrepo.PageResult[*entity.Party]{}, eferrors.WithField(eferrors.ErrInvalidInput, "email", "email is required")
	}
	return s.repo.SelectByEmail(ctx, normalized, page.Normalize())
}

// List mengambil halaman daftar mitra eksternal berdasarkan filter pencarian, jenis, status, dan peran aktif.
func (s *Service) List(ctx context.Context, filter domainrepo.PartyFilter) (domainrepo.PageResult[*entity.Party], error) {
	if err := s.ensureReady(); err != nil {
		return domainrepo.PageResult[*entity.Party]{}, err
	}
	return s.repo.SelectList(ctx, filter.Normalize())
}

// GetAll mengambil daftar mitra eksternal non-terhapus secara berhalaman.
func (s *Service) GetAll(ctx context.Context, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.Party], error) {
	if err := s.ensureReady(); err != nil {
		return domainrepo.PageResult[*entity.Party]{}, err
	}
	return s.repo.SelectList(ctx, domainrepo.PartyFilter{Page: page.Normalize()})
}

// Create mendaftarkan mitra eksternal baru setelah memeriksa duplikasi partner_code dan tax_id_number sebelum Insert.
func (s *Service) Create(ctx context.Context, in CreateInput) (*entity.Party, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}

	id, err := s.idGen.NewUUID()
	if err != nil {
		return nil, eferrors.Wrap(err, eferrors.KindUnavailable, eferrors.Code("enterprise_foundation.party.id_generation_failed"), "failed to generate party id")
	}

	party, err := entity.NewParty(id, entity.PartyParams{
		PartnerCode:        in.PartnerCode,
		Kind:               in.Kind,
		DisplayName:        in.DisplayName,
		LegalName:          in.LegalName,
		Email:              in.Email,
		Phone:              in.Phone,
		WebsiteURL:         in.WebsiteURL,
		TaxIDNumber:        in.TaxIDNumber,
		TaxRegisteredName:  in.TaxRegisteredName,
		IsPKP:              in.IsPKP,
		IsTaxExempt:        in.IsTaxExempt,
		TaxExemptionReason: in.TaxExemptionReason,
		TaxVerifiedAt:      in.TaxVerifiedAt,
		PreferredLanguage:  in.PreferredLanguage,
		Status:             in.Status,
		Notes:              in.Notes,
	}, s.clock.Now())
	if err != nil {
		return nil, err
	}

	if err := s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := s.ensureUniqueOnCreate(txCtx, party); err != nil {
			return err
		}
		if err := s.repo.Insert(txCtx, party); err != nil {
			return err
		}
		return servicekit.EmitOutboxInTx(
			txCtx,
			s.outbox,
			s.outboxBuilder,
			entity.EventTypePartyCreated,
			entity.AggregateTypeParty,
			party.ID,
			partyStatusPayload(party),
		)
	}); err != nil {
		return nil, err
	}
	return party, nil
}

// UpdateProfile memperbarui profil umum mitra secara atomik beserta event outbox pembaruan.
func (s *Service) UpdateProfile(ctx context.Context, in UpdateProfileInput) (*entity.Party, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("party_id", in.ID); err != nil {
		return nil, err
	}
	normID := entity.NormalizeUUID(in.ID)

	var updated *entity.Party
	if err := s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		party, err := s.repo.SelectByID(txCtx, normID)
		if err != nil {
			return err
		}
		if err := party.UpdateProfile(entity.PartyProfileParams{
			Kind:              in.Kind,
			DisplayName:       in.DisplayName,
			LegalName:         in.LegalName,
			Email:             in.Email,
			Phone:             in.Phone,
			WebsiteURL:        in.WebsiteURL,
			PreferredLanguage: in.PreferredLanguage,
			Notes:             in.Notes,
		}, s.clock.Now()); err != nil {
			return err
		}
		if err := s.repo.Update(txCtx, party); err != nil {
			return err
		}
		if err := servicekit.EmitOutboxInTx(
			txCtx,
			s.outbox,
			s.outboxBuilder,
			entity.EventTypePartyUpdated,
			entity.AggregateTypeParty,
			party.ID,
			partyStatusPayload(party),
		); err != nil {
			return err
		}
		updated = party
		return nil
	}); err != nil {
		return nil, err
	}
	return updated, nil
}

// UpdateTaxProfile memperbarui data perpajakan mitra sekaligus mencegah duplikasi NPWP terhadap mitra lain.
func (s *Service) UpdateTaxProfile(ctx context.Context, in UpdateTaxProfileInput) (*entity.Party, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("party_id", in.ID); err != nil {
		return nil, err
	}
	normID := entity.NormalizeUUID(in.ID)

	var updated *entity.Party
	if err := s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		party, err := s.repo.SelectByID(txCtx, normID)
		if err != nil {
			return err
		}
		if err := party.UpdateTaxProfile(entity.PartyTaxParams{
			TaxIDNumber:        in.TaxIDNumber,
			TaxRegisteredName:  in.TaxRegisteredName,
			IsPKP:              in.IsPKP,
			IsTaxExempt:        in.IsTaxExempt,
			TaxExemptionReason: in.TaxExemptionReason,
			TaxVerifiedAt:      in.TaxVerifiedAt,
		}, s.clock.Now()); err != nil {
			return err
		}
		if party.TaxIDNumber != nil {
			exists, err := s.repo.CheckTaxIDExists(txCtx, *party.TaxIDNumber, party.ID)
			if err != nil {
				return err
			}
			if exists {
				return eferrors.ErrPartyTaxIDExists
			}
		}
		if err := s.repo.Update(txCtx, party); err != nil {
			return err
		}
		if err := servicekit.EmitOutboxInTx(
			txCtx,
			s.outbox,
			s.outboxBuilder,
			entity.EventTypePartyUpdated,
			entity.AggregateTypeParty,
			party.ID,
			partyStatusPayload(party),
		); err != nil {
			return err
		}
		updated = party
		return nil
	}); err != nil {
		return nil, err
	}
	return updated, nil
}

// VerifyAndActivate memverifikasi mitra agar memenuhi syarat transaksi dan otorisasi portal IAM.
func (s *Service) VerifyAndActivate(ctx context.Context, id string) (*entity.Party, error) {
	return s.changeStatus(ctx, id, func(party *entity.Party, now time.Time) error {
		return party.VerifyAndActivate(now)
	})
}

// Suspend menangguhkan sementara mitra dari transaksi baru dan akses perwakilan portal.
func (s *Service) Suspend(ctx context.Context, id string) (*entity.Party, error) {
	return s.changeStatus(ctx, id, func(party *entity.Party, now time.Time) error {
		return party.Suspend(now)
	})
}

// Deactivate menonaktifkan mitra tanpa menghapus riwayat referensi transaksinya.
func (s *Service) Deactivate(ctx context.Context, id string) (*entity.Party, error) {
	return s.changeStatus(ctx, id, func(party *entity.Party, now time.Time) error {
		return party.Deactivate(now)
	})
}

// Delete menghapus lunak entitas mitra serta merekam event perubahan status ke outbox secara atomik.
func (s *Service) Delete(ctx context.Context, id string) error {
	if err := s.ensureReady(); err != nil {
		return err
	}
	if err := entity.ValidateUUID("party_id", id); err != nil {
		return err
	}
	normID := entity.NormalizeUUID(id)

	return s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		party, err := s.repo.SelectByID(txCtx, normID)
		if err != nil {
			return err
		}
		now := s.clock.Now()
		if err := party.SoftDelete(now); err != nil {
			return err
		}
		if err := s.repo.SoftDeleteByID(txCtx, normID, now); err != nil {
			return err
		}
		return servicekit.EmitOutboxInTx(
			txCtx,
			s.outbox,
			s.outboxBuilder,
			entity.EventTypePartyStatusChanged,
			entity.AggregateTypeParty,
			party.ID,
			partyStatusPayload(party),
		)
	})
}

func (s *Service) changeStatus(
	ctx context.Context,
	id string,
	apply func(party *entity.Party, now time.Time) error,
) (*entity.Party, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("party_id", id); err != nil {
		return nil, err
	}
	normID := entity.NormalizeUUID(id)

	var updated *entity.Party
	if err := s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		party, err := s.repo.SelectByID(txCtx, normID)
		if err != nil {
			return err
		}
		if err := apply(party, s.clock.Now()); err != nil {
			return err
		}
		if err := s.repo.Update(txCtx, party); err != nil {
			return err
		}
		if err := servicekit.EmitOutboxInTx(
			txCtx,
			s.outbox,
			s.outboxBuilder,
			entity.EventTypePartyStatusChanged,
			entity.AggregateTypeParty,
			party.ID,
			partyStatusPayload(party),
		); err != nil {
			return err
		}
		updated = party
		return nil
	}); err != nil {
		return nil, err
	}
	return updated, nil
}

func (s *Service) ensureUniqueOnCreate(ctx context.Context, party *entity.Party) error {
	existingByCode, err := s.repo.SelectByCode(ctx, party.PartnerCode)
	if err != nil && eferrors.KindOf(err) != eferrors.KindNotFound {
		return err
	}
	if existingByCode != nil {
		return eferrors.ErrPartyCodeExists
	}

	if party.TaxIDNumber != nil {
		existingByTax, err := s.repo.SelectByTaxIDNumber(ctx, *party.TaxIDNumber)
		if err != nil && eferrors.KindOf(err) != eferrors.KindNotFound {
			return err
		}
		if existingByTax != nil {
			return eferrors.ErrPartyTaxIDExists
		}
	}
	return nil
}

func (s *Service) ensureReady() error {
	if s == nil || s.tx == nil || s.repo == nil || s.roles == nil || s.contacts == nil || s.partyAddresses == nil || s.addresses == nil || s.outbox == nil || s.outboxBuilder == nil || s.idGen == nil {
		return eferrors.Wrap(eferrors.ErrUnavailable, eferrors.KindUnavailable, eferrors.Code("enterprise_foundation.service.not_initialized"), "party service is not initialized")
	}
	return nil
}

func partyStatusPayload(party *entity.Party) domainservice.EntityStatusPayload {
	return domainservice.EntityStatusPayload{
		EntityID: party.ID,
		Code:     party.PartnerCode,
		Status:   string(party.Status),
		IsActive: party.Status == entity.PartyStatusActive && !party.IsDeleted(),
	}
}
