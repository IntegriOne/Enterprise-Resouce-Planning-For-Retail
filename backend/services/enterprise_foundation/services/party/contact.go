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

// GetContactByID mengambil detail satu kontak/PIC mitra berdasarkan UUID.
func (s *Service) GetContactByID(ctx context.Context, contactID string) (*entity.PartyContact, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("contact_id", contactID); err != nil {
		return nil, err
	}
	return s.contacts.SelectByID(ctx, entity.NormalizeUUID(contactID))
}

// ListContactsByParty mengambil daftar kontak/PIC milik satu mitra eksternal secara berhalaman.
func (s *Service) ListContactsByParty(ctx context.Context, partyID string, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.PartyContact], error) {
	if err := s.ensureReady(); err != nil {
		return domainrepo.PageResult[*entity.PartyContact]{}, err
	}
	if err := entity.ValidateUUID("party_id", partyID); err != nil {
		return domainrepo.PageResult[*entity.PartyContact]{}, err
	}
	normPartyID := entity.NormalizeUUID(partyID)
	if _, err := s.repo.SelectByID(ctx, normPartyID); err != nil {
		return domainrepo.PageResult[*entity.PartyContact]{}, err
	}
	return s.contacts.SelectByPartyID(ctx, normPartyID, page.Normalize())
}

// GetPrimaryContact mengambil PIC utama milik satu mitra eksternal.
func (s *Service) GetPrimaryContact(ctx context.Context, partyID string) (*entity.PartyContact, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("party_id", partyID); err != nil {
		return nil, err
	}
	normPartyID := entity.NormalizeUUID(partyID)
	if _, err := s.repo.SelectByID(ctx, normPartyID); err != nil {
		return nil, err
	}
	return s.contacts.SelectPrimaryByPartyID(ctx, normPartyID)
}

// ListContactsByIAMUserID mencari tautan kontak mitra yang terhubung ke satu principal IAM secara berhalaman.
func (s *Service) ListContactsByIAMUserID(ctx context.Context, iamUserID string, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.PartyContact], error) {
	if err := s.ensureReady(); err != nil {
		return domainrepo.PageResult[*entity.PartyContact]{}, err
	}
	if err := entity.ValidateUUID("iam_user_id", iamUserID); err != nil {
		return domainrepo.PageResult[*entity.PartyContact]{}, err
	}
	return s.contacts.SelectByIAMUserID(ctx, entity.NormalizeUUID(iamUserID), page.Normalize())
}

// ListContactsByEmail mencari kontak mitra yang memakai alamat email tertentu secara berhalaman.
func (s *Service) ListContactsByEmail(ctx context.Context, email string, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.PartyContact], error) {
	if err := s.ensureReady(); err != nil {
		return domainrepo.PageResult[*entity.PartyContact]{}, err
	}
	normalized := strings.ToLower(strings.TrimSpace(email))
	if normalized == "" {
		return domainrepo.PageResult[*entity.PartyContact]{}, eferrors.WithField(eferrors.ErrInvalidInput, "email", "email is required")
	}
	return s.contacts.SelectByEmail(ctx, normalized, page.Normalize())
}

// ListContacts mengambil daftar halaman kontak/PIC sesuai filter mitra, peran kontak, dan status portal IAM.
func (s *Service) ListContacts(ctx context.Context, filter domainrepo.PartyContactFilter) (domainrepo.PageResult[*entity.PartyContact], error) {
	if err := s.ensureReady(); err != nil {
		return domainrepo.PageResult[*entity.PartyContact]{}, err
	}
	return s.contacts.SelectList(ctx, filter.Normalize())
}

// GetAllContacts mengambil daftar kontak/PIC non-terhapus secara berhalaman tanpa filter tambahan.
func (s *Service) GetAllContacts(ctx context.Context, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.PartyContact], error) {
	if err := s.ensureReady(); err != nil {
		return domainrepo.PageResult[*entity.PartyContact]{}, err
	}
	return s.contacts.SelectList(ctx, domainrepo.PartyContactFilter{Page: page.Normalize()})
}

// EvaluateRepresentativePortalAccess memvalidasi secara fail-closed hak perwakilan user IAM atas mitra dan peran bisnis yang diminta.
func (s *Service) EvaluateRepresentativePortalAccess(
	ctx context.Context,
	partyID, iamUserID string,
	requiredRole entity.PartyRoleKind,
) (domainservice.PortalAccessEvaluation, error) {
	if err := s.ensureReady(); err != nil {
		return domainservice.PortalAccessEvaluation{}, err
	}
	if err := entity.ValidateUUID("party_id", partyID); err != nil {
		return domainservice.PortalAccessEvaluation{}, err
	}
	if err := entity.ValidateUUID("iam_user_id", iamUserID); err != nil {
		return domainservice.PortalAccessEvaluation{}, err
	}
	if !requiredRole.IsValid() {
		return domainservice.PortalAccessEvaluation{}, eferrors.WithField(eferrors.ErrInvalidPartyRoleKind, "role", "invalid party role kind")
	}

	normPartyID := entity.NormalizeUUID(partyID)
	normIAMUserID := entity.NormalizeUUID(iamUserID)
	party, err := s.repo.SelectByID(ctx, normPartyID)
	if err != nil {
		return domainservice.PortalAccessEvaluation{}, err
	}
	roles, err := s.collectAllPartyRoles(ctx, normPartyID)
	if err != nil {
		return domainservice.PortalAccessEvaluation{}, err
	}
	contact, err := s.contacts.SelectByPartyAndIAMUserID(ctx, normPartyID, normIAMUserID)
	if err != nil {
		if eferrors.KindOf(err) == eferrors.KindNotFound {
			return s.portal.EvaluateRepresentativeAccess(party, roles, nil, normIAMUserID, requiredRole)
		}
		return domainservice.PortalAccessEvaluation{}, err
	}
	return s.portal.EvaluateRepresentativeAccess(party, roles, contact, normIAMUserID, requiredRole)
}

// AddContact menambahkan kontak/PIC mitra setelah memastikan principal IAM belum tertaut pada kontak lain di mitra yang sama.
func (s *Service) AddContact(ctx context.Context, in AddContactInput) (*entity.PartyContact, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	id, err := s.idGen.NewUUID()
	if err != nil {
		return nil, eferrors.Wrap(err, eferrors.KindUnavailable, eferrors.Code("enterprise_foundation.party_contact.id_generation_failed"), "failed to generate party contact id")
	}

	now := s.clock.Now()
	contact, err := entity.NewPartyContact(id, entity.PartyContactParams{
		PartyID:            in.PartyID,
		IAMUserID:          in.IAMUserID,
		FullName:           in.FullName,
		JobTitle:           in.JobTitle,
		ContactRole:        in.ContactRole,
		Email:              in.Email,
		Phone:              in.Phone,
		PortalAccessStatus: in.PortalAccessStatus,
		IsPrimary:          in.IsPrimary,
	}, now)
	if err != nil {
		return nil, err
	}

	if err := s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		party, err := s.repo.SelectByID(txCtx, contact.PartyID)
		if err != nil {
			return err
		}
		if party.IsDeleted() {
			return eferrors.ErrPartyDeleted
		}
		if contact.PortalAccessStatus == entity.PortalAccessStatusActive {
			if err := party.IsEligibleForTransactions(); err != nil {
				return err
			}
		}
		if contact.IAMUserID != nil {
			existing, err := s.contacts.SelectByPartyAndIAMUserID(txCtx, contact.PartyID, *contact.IAMUserID)
			if err != nil && eferrors.KindOf(err) != eferrors.KindNotFound {
				return err
			}
			if existing != nil {
				return eferrors.ErrPartyContactIAMUserExists
			}
		}
		if contact.IsPrimary {
			if err := s.contacts.UpdateClearPrimaryByPartyExcept(txCtx, contact.PartyID, contact.ID, now); err != nil {
				return err
			}
		}
		if err := s.contacts.Insert(txCtx, contact); err != nil {
			return err
		}
		if contact.IAMUserID != nil || contact.PortalAccessStatus != entity.PortalAccessStatusNone {
			return s.emitPortalOutbox(txCtx, contact)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return contact, nil
}

// UpdateContact memperbarui detail kontak/PIC dan menjaga konsistensi penanda kontak utama.
func (s *Service) UpdateContact(ctx context.Context, in UpdateContactInput) (*entity.PartyContact, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("contact_id", in.ID); err != nil {
		return nil, err
	}
	normID := entity.NormalizeUUID(in.ID)

	var updated *entity.PartyContact
	if err := s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		contact, err := s.contacts.SelectByID(txCtx, normID)
		if err != nil {
			return err
		}
		now := s.clock.Now()
		if err := contact.UpdateDetails(entity.PartyContactUpdateParams{
			FullName:    in.FullName,
			JobTitle:    in.JobTitle,
			ContactRole: in.ContactRole,
			Email:       in.Email,
			Phone:       in.Phone,
			IsPrimary:   in.IsPrimary,
		}, now); err != nil {
			return err
		}
		if contact.IsPrimary {
			if err := s.contacts.UpdateClearPrimaryByPartyExcept(txCtx, contact.PartyID, contact.ID, now); err != nil {
				return err
			}
		}
		if err := s.contacts.Update(txCtx, contact); err != nil {
			return err
		}
		updated = contact
		return nil
	}); err != nil {
		return nil, err
	}
	return updated, nil
}

// SetPrimaryContact menetapkan satu kontak/PIC aktif sebagai kontak utama mitra secara atomik.
func (s *Service) SetPrimaryContact(ctx context.Context, partyID, contactID string) (*entity.PartyContact, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("party_id", partyID); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("contact_id", contactID); err != nil {
		return nil, err
	}
	normPartyID := entity.NormalizeUUID(partyID)
	normContactID := entity.NormalizeUUID(contactID)

	var updated *entity.PartyContact
	if err := s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		contact, err := s.contacts.SelectByID(txCtx, normContactID)
		if err != nil {
			return err
		}
		if contact.PartyID != normPartyID {
			return eferrors.WithField(eferrors.ErrPartyContactNotFound, "contact_id", "contact does not belong to party")
		}
		now := s.clock.Now()
		if err := contact.UpdateDetails(entity.PartyContactUpdateParams{
			FullName:    contact.FullName,
			JobTitle:    contact.JobTitle,
			ContactRole: contact.ContactRole,
			Email:       contact.Email,
			Phone:       contact.Phone,
			IsPrimary:   true,
		}, now); err != nil {
			return err
		}
		if err := s.contacts.UpdateClearPrimaryByPartyExcept(txCtx, normPartyID, normContactID, now); err != nil {
			return err
		}
		if err := s.contacts.Update(txCtx, contact); err != nil {
			return err
		}
		updated = contact
		return nil
	}); err != nil {
		return nil, err
	}
	return updated, nil
}

// RequestPortalLink memulai permintaan penautan kontak PIC mitra ke principal IAM (status pending_link).
func (s *Service) RequestPortalLink(ctx context.Context, contactID, iamUserID string) (*entity.PartyContact, error) {
	if err := entity.ValidateUUID("iam_user_id", iamUserID); err != nil {
		return nil, err
	}
	normIAMUserID := entity.NormalizeUUID(iamUserID)

	return s.transitionPortal(ctx, contactID, func(txCtx context.Context, party *entity.Party, contact *entity.PartyContact, now time.Time) error {
		if party.IsDeleted() {
			return eferrors.ErrPartyDeleted
		}
		sameUser := contact.IAMUserID == nil || *contact.IAMUserID == normIAMUserID
		if err := s.portal.CanTransitionPortalStatus(contact.PortalAccessStatus, entity.PortalAccessStatusPendingLink, sameUser); err != nil {
			return err
		}
		exists, err := s.contacts.CheckIAMUserExistsInParty(txCtx, contact.PartyID, normIAMUserID, contact.ID)
		if err != nil {
			return err
		}
		if exists {
			return eferrors.ErrPartyContactIAMUserExists
		}
		return contact.RequestPortalLink(normIAMUserID, now)
	})
}

// ActivatePortalAccess mengaktifkan akses portal PIC mitra pada principal IAM setelah memverifikasi kelayakan mitra.
func (s *Service) ActivatePortalAccess(ctx context.Context, contactID, iamUserID string) (*entity.PartyContact, error) {
	if err := entity.ValidateUUID("iam_user_id", iamUserID); err != nil {
		return nil, err
	}
	normIAMUserID := entity.NormalizeUUID(iamUserID)

	return s.transitionPortal(ctx, contactID, func(txCtx context.Context, party *entity.Party, contact *entity.PartyContact, now time.Time) error {
		if err := party.IsEligibleForTransactions(); err != nil {
			return err
		}
		sameUser := contact.IAMUserID == nil || *contact.IAMUserID == normIAMUserID
		if err := s.portal.CanTransitionPortalStatus(contact.PortalAccessStatus, entity.PortalAccessStatusActive, sameUser); err != nil {
			return err
		}
		exists, err := s.contacts.CheckIAMUserExistsInParty(txCtx, contact.PartyID, normIAMUserID, contact.ID)
		if err != nil {
			return err
		}
		if exists {
			return eferrors.ErrPartyContactIAMUserExists
		}
		return contact.ActivatePortalAccess(normIAMUserID, now)
	})
}

// RevokePortalAccess mencabut akses portal PIC mitra sambil mempertahankan jejak referensi kontak.
func (s *Service) RevokePortalAccess(ctx context.Context, contactID string) (*entity.PartyContact, error) {
	return s.transitionPortal(ctx, contactID, func(_ context.Context, _ *entity.Party, contact *entity.PartyContact, now time.Time) error {
		if err := s.portal.CanTransitionPortalStatus(contact.PortalAccessStatus, entity.PortalAccessStatusRevoked, true); err != nil {
			return err
		}
		return contact.RevokePortalAccess(now)
	})
}

// UnlinkPortalUser memutus seluruh tautan akun IAM dari kontak PIC dan mengembalikan status portal ke none.
func (s *Service) UnlinkPortalUser(ctx context.Context, contactID string) (*entity.PartyContact, error) {
	return s.transitionPortal(ctx, contactID, func(_ context.Context, _ *entity.Party, contact *entity.PartyContact, now time.Time) error {
		if err := s.portal.CanTransitionPortalStatus(contact.PortalAccessStatus, entity.PortalAccessStatusNone, true); err != nil {
			return err
		}
		return contact.UnlinkPortalUser(now)
	})
}

// ActivateContact mengaktifkan kembali kontak/PIC mitra dan mengirim sinkronisasi bila tertaut ke portal IAM.
func (s *Service) ActivateContact(ctx context.Context, contactID string) (*entity.PartyContact, error) {
	return s.toggleContactActive(ctx, contactID, func(contact *entity.PartyContact, now time.Time) error {
		return contact.Activate(now)
	})
}

// DeactivateContact menonaktifkan sementara kontak/PIC mitra beserta akses portal efektifnya.
func (s *Service) DeactivateContact(ctx context.Context, contactID string) (*entity.PartyContact, error) {
	return s.toggleContactActive(ctx, contactID, func(contact *entity.PartyContact, now time.Time) error {
		return contact.Deactivate(now)
	})
}

// DeleteContact menghapus lunak kontak/PIC sekaligus mencabut akses portal aktif jika tertaut ke IAM.
func (s *Service) DeleteContact(ctx context.Context, contactID string) error {
	if err := s.ensureReady(); err != nil {
		return err
	}
	if err := entity.ValidateUUID("contact_id", contactID); err != nil {
		return err
	}
	normID := entity.NormalizeUUID(contactID)

	return s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		contact, err := s.contacts.SelectByID(txCtx, normID)
		if err != nil {
			return err
		}
		hadPortal := contact.IAMUserID != nil || contact.PortalAccessStatus != entity.PortalAccessStatusNone
		now := s.clock.Now()
		if err := contact.SoftDelete(now); err != nil {
			return err
		}
		if err := s.contacts.Update(txCtx, contact); err != nil {
			return err
		}
		if err := s.contacts.SoftDeleteByID(txCtx, normID, now); err != nil {
			return err
		}
		if hadPortal {
			return s.emitPortalOutbox(txCtx, contact)
		}
		return nil
	})
}

func (s *Service) transitionPortal(
	ctx context.Context,
	contactID string,
	apply func(txCtx context.Context, party *entity.Party, contact *entity.PartyContact, now time.Time) error,
) (*entity.PartyContact, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("contact_id", contactID); err != nil {
		return nil, err
	}
	normID := entity.NormalizeUUID(contactID)

	var updated *entity.PartyContact
	if err := s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		contact, err := s.contacts.SelectByID(txCtx, normID)
		if err != nil {
			return err
		}
		party, err := s.repo.SelectByID(txCtx, contact.PartyID)
		if err != nil {
			return err
		}
		if err := apply(txCtx, party, contact, s.clock.Now()); err != nil {
			return err
		}
		if err := s.contacts.Update(txCtx, contact); err != nil {
			return err
		}
		if err := s.emitPortalOutbox(txCtx, contact); err != nil {
			return err
		}
		updated = contact
		return nil
	}); err != nil {
		return nil, err
	}
	return updated, nil
}

func (s *Service) toggleContactActive(
	ctx context.Context,
	contactID string,
	apply func(contact *entity.PartyContact, now time.Time) error,
) (*entity.PartyContact, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("contact_id", contactID); err != nil {
		return nil, err
	}
	normID := entity.NormalizeUUID(contactID)

	var updated *entity.PartyContact
	if err := s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		contact, err := s.contacts.SelectByID(txCtx, normID)
		if err != nil {
			return err
		}
		if err := apply(contact, s.clock.Now()); err != nil {
			return err
		}
		if err := s.contacts.Update(txCtx, contact); err != nil {
			return err
		}
		if contact.IAMUserID != nil || contact.PortalAccessStatus != entity.PortalAccessStatusNone {
			if err := s.emitPortalOutbox(txCtx, contact); err != nil {
				return err
			}
		}
		updated = contact
		return nil
	}); err != nil {
		return nil, err
	}
	return updated, nil
}

func (s *Service) emitPortalOutbox(ctx context.Context, contact *entity.PartyContact) error {
	var iamUserCopy *string
	if contact.IAMUserID != nil {
		v := *contact.IAMUserID
		iamUserCopy = &v
	}
	return servicekit.EmitOutboxInTx(
		ctx,
		s.outbox,
		s.outboxBuilder,
		entity.EventTypePartyContactPortalChanged,
		entity.AggregateTypePartyContact,
		contact.ID,
		domainservice.PartyContactPortalPayload{
			ContactID:          contact.ID,
			PartyID:            contact.PartyID,
			IAMUserID:          iamUserCopy,
			PortalAccessStatus: contact.PortalAccessStatus,
			IsActive:           contact.IsActive && !contact.IsDeleted(),
		},
	)
}
