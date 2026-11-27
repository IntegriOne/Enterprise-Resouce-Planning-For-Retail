package party

import (
	"context"
	"time"

	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/entity"
	eferrors "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/repository"
	domainservice "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/service"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/services/internal/servicekit"
)

// GetRoleByID mengambil detail satu peran bisnis mitra berdasarkan UUID baris party_roles.
func (s *Service) GetRoleByID(ctx context.Context, roleID string) (*entity.PartyRole, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("party_role_id", roleID); err != nil {
		return nil, err
	}
	return s.roles.SelectByID(ctx, entity.NormalizeUUID(roleID))
}

// ListRolesByParty mengambil peran bisnis yang terdaftar pada satu mitra eksternal secara berhalaman.
func (s *Service) ListRolesByParty(ctx context.Context, partyID string, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.PartyRole], error) {
	if err := s.ensureReady(); err != nil {
		return domainrepo.PageResult[*entity.PartyRole]{}, err
	}
	if err := entity.ValidateUUID("party_id", partyID); err != nil {
		return domainrepo.PageResult[*entity.PartyRole]{}, err
	}
	normPartyID := entity.NormalizeUUID(partyID)
	if _, err := s.repo.SelectByID(ctx, normPartyID); err != nil {
		return domainrepo.PageResult[*entity.PartyRole]{}, err
	}
	return s.roles.SelectByPartyID(ctx, normPartyID, page.Normalize())
}

// GetRoleByPartyAndKind mengambil satu peran spesifik yang dimiliki mitra tertentu.
func (s *Service) GetRoleByPartyAndKind(ctx context.Context, partyID string, kind entity.PartyRoleKind) (*entity.PartyRole, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("party_id", partyID); err != nil {
		return nil, err
	}
	if !kind.IsValid() {
		return nil, eferrors.WithField(eferrors.ErrInvalidPartyRoleKind, "role", "invalid party role kind")
	}
	normPartyID := entity.NormalizeUUID(partyID)
	if _, err := s.repo.SelectByID(ctx, normPartyID); err != nil {
		return nil, err
	}
	return s.roles.SelectByPartyAndRole(ctx, normPartyID, kind)
}

// ListRoles mengambil daftar halaman peran mitra berdasarkan filter mitra, jenis peran, keaktifan, dan kata kunci catatan.
func (s *Service) ListRoles(ctx context.Context, filter domainrepo.PartyRoleFilter) (domainrepo.PageResult[*entity.PartyRole], error) {
	if err := s.ensureReady(); err != nil {
		return domainrepo.PageResult[*entity.PartyRole]{}, err
	}
	return s.roles.SelectList(ctx, filter.Normalize())
}

// GetAllRoles mengambil seluruh daftar peran mitra secara berhalaman tanpa filter tambahan.
func (s *Service) GetAllRoles(ctx context.Context, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.PartyRole], error) {
	if err := s.ensureReady(); err != nil {
		return domainrepo.PageResult[*entity.PartyRole]{}, err
	}
	return s.roles.SelectList(ctx, domainrepo.PartyRoleFilter{Page: page.Normalize()})
}

// ListRolesByKind mengambil daftar halaman peran mitra berdasarkan jenis peran bisnis dan filter keaktifan.
func (s *Service) ListRolesByKind(ctx context.Context, kind entity.PartyRoleKind, activeOnly bool, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.PartyRole], error) {
	if err := s.ensureReady(); err != nil {
		return domainrepo.PageResult[*entity.PartyRole]{}, err
	}
	if !kind.IsValid() {
		return domainrepo.PageResult[*entity.PartyRole]{}, eferrors.WithField(eferrors.ErrInvalidPartyRoleKind, "role", "invalid party role kind")
	}
	return s.roles.SelectByRoleKind(ctx, kind, activeOnly, page.Normalize())
}

// GetAllRolesByKind mengambil daftar halaman peran mitra untuk kategori peran tertentu.
func (s *Service) GetAllRolesByKind(ctx context.Context, kind entity.PartyRoleKind, activeOnly bool, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.PartyRole], error) {
	if err := s.ensureReady(); err != nil {
		return domainrepo.PageResult[*entity.PartyRole]{}, err
	}
	if !kind.IsValid() {
		return domainrepo.PageResult[*entity.PartyRole]{}, eferrors.WithField(eferrors.ErrInvalidPartyRoleKind, "role", "invalid party role kind")
	}
	return s.roles.SelectByRoleKind(ctx, kind, activeOnly, page.Normalize())
}

// collectAllPartyRoles menelusuri seluruh peran milik satu mitra secara bertahap dan terikat untuk evaluasi otorisasi domain.
func (s *Service) collectAllPartyRoles(ctx context.Context, normPartyID string) ([]*entity.PartyRole, error) {
	return servicekit.CollectAllPages(ctx, func(pageCtx context.Context, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.PartyRole], error) {
		return s.roles.SelectByPartyID(pageCtx, normPartyID, page)
	})
}

// ValidatePartyRoleScope memeriksa apakah mitra aktif memiliki peran bisnis tertentu yang aktif untuk transaksi atau scope IAM.
func (s *Service) ValidatePartyRoleScope(ctx context.Context, partyID string, requiredRole entity.PartyRoleKind) (domainservice.ScopeTargetDecision, error) {
	if err := s.ensureReady(); err != nil {
		return domainservice.ScopeTargetDecision{}, err
	}
	if err := entity.ValidateUUID("party_id", partyID); err != nil {
		return domainservice.ScopeTargetDecision{}, err
	}
	if !requiredRole.IsValid() {
		return domainservice.ScopeTargetDecision{}, eferrors.WithField(eferrors.ErrInvalidPartyRoleKind, "role", "invalid party role kind")
	}
	normPartyID := entity.NormalizeUUID(partyID)
	party, err := s.repo.SelectByID(ctx, normPartyID)
	if err != nil {
		return domainservice.ScopeTargetDecision{}, err
	}
	roles, err := s.collectAllPartyRoles(ctx, normPartyID)
	if err != nil {
		return domainservice.ScopeTargetDecision{}, err
	}
	return s.scope.EvaluatePartyRole(party, roles, requiredRole)
}

// AssignRole menambahkan peran bisnis baru pada mitra setelah memastikan tidak ada duplikasi peran pada mitra yang sama.
func (s *Service) AssignRole(ctx context.Context, in AssignRoleInput) (*entity.PartyRole, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	id, err := s.idGen.NewUUID()
	if err != nil {
		return nil, eferrors.Wrap(err, eferrors.KindUnavailable, eferrors.Code("enterprise_foundation.party_role.id_generation_failed"), "failed to generate party role id")
	}

	role, err := entity.NewPartyRole(id, in.PartyID, in.Role, in.ReferenceNote, s.clock.Now())
	if err != nil {
		return nil, err
	}

	if err := s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		party, err := s.repo.SelectByID(txCtx, role.PartyID)
		if err != nil {
			return err
		}
		if party.IsDeleted() {
			return eferrors.ErrPartyDeleted
		}
		existing, err := s.roles.SelectByPartyAndRole(txCtx, role.PartyID, role.Role)
		if err != nil && eferrors.KindOf(err) != eferrors.KindNotFound {
			return err
		}
		if existing != nil {
			return eferrors.ErrPartyRoleExists
		}
		if err := s.roles.Insert(txCtx, role); err != nil {
			return err
		}
		return s.emitPartyRoleOutbox(txCtx, party, role, role.IsActive)
	}); err != nil {
		return nil, err
	}
	return role, nil
}

// UpdateRoleNote memperbarui catatan referensi peran bisnis mitra.
func (s *Service) UpdateRoleNote(ctx context.Context, in UpdateRoleInput) (*entity.PartyRole, error) {
	return s.mutateRole(ctx, in.ID, func(role *entity.PartyRole, now time.Time) error {
		return role.UpdateReferenceNote(in.ReferenceNote, now)
	})
}

// ActivateRole mengaktifkan kembali peran bisnis mitra dan mencatat event sinkronisasi outbox.
func (s *Service) ActivateRole(ctx context.Context, roleID string) (*entity.PartyRole, error) {
	return s.mutateRole(ctx, roleID, func(role *entity.PartyRole, now time.Time) error {
		return role.Activate(now)
	})
}

// DeactivateRole menonaktifkan peran bisnis mitra dan mencatat event sinkronisasi outbox.
func (s *Service) DeactivateRole(ctx context.Context, roleID string) (*entity.PartyRole, error) {
	return s.mutateRole(ctx, roleID, func(role *entity.PartyRole, now time.Time) error {
		return role.Deactivate(now)
	})
}

// RemoveRole menghapus penugasan peran bisnis dari mitra beserta event penonaktifannya ke outbox.
func (s *Service) RemoveRole(ctx context.Context, roleID string) error {
	if err := s.ensureReady(); err != nil {
		return err
	}
	if err := entity.ValidateUUID("party_role_id", roleID); err != nil {
		return err
	}
	normID := entity.NormalizeUUID(roleID)

	return s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		role, err := s.roles.SelectByID(txCtx, normID)
		if err != nil {
			return err
		}
		party, err := s.repo.SelectByID(txCtx, role.PartyID)
		if err != nil {
			return err
		}
		if err := s.roles.DeleteByID(txCtx, normID); err != nil {
			return err
		}
		return s.emitPartyRoleOutbox(txCtx, party, role, false)
	})
}

func (s *Service) mutateRole(
	ctx context.Context,
	roleID string,
	apply func(role *entity.PartyRole, now time.Time) error,
) (*entity.PartyRole, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("party_role_id", roleID); err != nil {
		return nil, err
	}
	normID := entity.NormalizeUUID(roleID)

	var updated *entity.PartyRole
	if err := s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		role, err := s.roles.SelectByID(txCtx, normID)
		if err != nil {
			return err
		}
		party, err := s.repo.SelectByID(txCtx, role.PartyID)
		if err != nil {
			return err
		}
		if err := apply(role, s.clock.Now()); err != nil {
			return err
		}
		if err := s.roles.Update(txCtx, role); err != nil {
			return err
		}
		if err := s.emitPartyRoleOutbox(txCtx, party, role, role.IsActive); err != nil {
			return err
		}
		updated = role
		return nil
	}); err != nil {
		return nil, err
	}
	return updated, nil
}

func (s *Service) emitPartyRoleOutbox(ctx context.Context, party *entity.Party, role *entity.PartyRole, active bool) error {
	return servicekit.EmitOutboxInTx(
		ctx,
		s.outbox,
		s.outboxBuilder,
		entity.EventTypePartyRoleChanged,
		entity.AggregateTypePartyRole,
		role.ID,
		domainservice.PartyRolePayload{
			PartyID:     party.ID,
			PartnerCode: party.PartnerCode,
			Role:        role.Role,
			IsActive:    active,
		},
	)
}
