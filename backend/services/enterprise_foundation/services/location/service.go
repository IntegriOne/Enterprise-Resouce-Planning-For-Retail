// Package location mengimplementasikan logika bisnis pengelolaan lokasi fisik internal (kantor, gudang, toko, cabang).
package location

import (
	"context"
	"time"

	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/entity"
	eferrors "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/repository"
	domainservice "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/service"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/services/internal/servicekit"
)

// Service mengelola CRUD, validasi hierarki lintas badan hukum, kelayakan scope fasilitas/gudang, serta outbox lokasi.
type Service struct {
	tx            domainrepo.Transactor
	repo          domainrepo.LocationRepository
	legalEntities domainrepo.LegalEntityRepository
	addresses     domainrepo.AddressRepository
	outbox        domainrepo.OutboxRepository
	outboxBuilder *domainservice.OutboxBuilder
	hierarchy     domainservice.HierarchyPolicy
	scope         domainservice.ScopePolicy
	clock         domainservice.Clock
	idGen         domainservice.IDGenerator
}

// NewService membentuk layanan lokasi fisik dengan validasi fail-closed atas seluruh dependensi wajib.
func NewService(deps Dependencies) (*Service, error) {
	if deps.Transactor == nil {
		return nil, eferrors.NewValidation("transactor", eferrors.Code("enterprise_foundation.service.nil_transactor"), "transactor is required")
	}
	if deps.Locations == nil {
		return nil, eferrors.NewValidation("locations", eferrors.Code("enterprise_foundation.service.nil_location_repo"), "location repository is required")
	}
	if deps.LegalEntities == nil {
		return nil, eferrors.NewValidation("legal_entities", eferrors.Code("enterprise_foundation.service.nil_legal_entity_repo"), "legal entity repository is required")
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
	hierarchy := deps.Hierarchy
	if hierarchy.MaxDepth() <= 0 {
		hierarchy = domainservice.NewDefaultHierarchyPolicy()
	}

	return &Service{
		tx:            deps.Transactor,
		repo:          deps.Locations,
		legalEntities: deps.LegalEntities,
		addresses:     deps.Addresses,
		outbox:        deps.Outbox,
		outboxBuilder: outboxBuilder,
		hierarchy:     hierarchy,
		scope:         domainservice.NewScopePolicy(),
		clock:         clock,
		idGen:         deps.IDGen,
	}, nil
}

// GetByID mengambil detail satu lokasi fisik berdasarkan UUID.
func (s *Service) GetByID(ctx context.Context, id string) (*entity.Location, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("location_id", id); err != nil {
		return nil, err
	}
	return s.repo.SelectByID(ctx, entity.NormalizeUUID(id))
}

// GetByIDs mengambil beberapa lokasi fisik sekaligus berdasarkan daftar UUID.
func (s *Service) GetByIDs(ctx context.Context, ids []string) ([]*entity.Location, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	normIDs, err := servicekit.ValidateOptionalUUIDList("location_ids", ids)
	if err != nil {
		return nil, err
	}
	if len(normIDs) == 0 {
		return []*entity.Location{}, nil
	}
	return s.repo.SelectByIDs(ctx, normIDs)
}

// GetByCode mencari lokasi fisik berdasarkan kode lokasi ternormalisasi.
func (s *Service) GetByCode(ctx context.Context, locationCode string) (*entity.Location, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	code := entity.NormalizeCode(locationCode)
	if code == "" {
		return nil, eferrors.WithField(eferrors.ErrInvalidLocationCode, "location_code", "location code is required")
	}
	return s.repo.SelectByCode(ctx, code)
}

// ListByLegalEntityAndKind mengambil daftar lokasi pada satu badan hukum berdasarkan jenisnya secara berhalaman.
func (s *Service) ListByLegalEntityAndKind(ctx context.Context, legalEntityID string, kind entity.LocationKind, activeOnly bool, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.Location], error) {
	if err := s.ensureReady(); err != nil {
		return domainrepo.PageResult[*entity.Location]{}, err
	}
	if err := entity.ValidateUUID("legal_entity_id", legalEntityID); err != nil {
		return domainrepo.PageResult[*entity.Location]{}, err
	}
	if !kind.IsValid() {
		return domainrepo.PageResult[*entity.Location]{}, eferrors.WithField(eferrors.ErrInvalidLocationKind, "kind", "invalid location kind")
	}
	normEntityID := entity.NormalizeUUID(legalEntityID)
	if _, err := s.legalEntities.SelectByID(ctx, normEntityID); err != nil {
		return domainrepo.PageResult[*entity.Location]{}, eferrors.WithField(err, "legal_entity_id", "legal entity not found")
	}
	return s.repo.SelectByLegalEntityAndKind(ctx, normEntityID, kind, activeOnly, page.Normalize())
}

// ListByAddressID mengambil daftar lokasi yang merujuk satu alamat master tertentu secara berhalaman.
func (s *Service) ListByAddressID(ctx context.Context, addressID string, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.Location], error) {
	if err := s.ensureReady(); err != nil {
		return domainrepo.PageResult[*entity.Location]{}, err
	}
	if err := entity.ValidateUUID("address_id", addressID); err != nil {
		return domainrepo.PageResult[*entity.Location]{}, err
	}
	normAddrID := entity.NormalizeUUID(addressID)
	if _, err := s.addresses.SelectByID(ctx, normAddrID); err != nil {
		return domainrepo.PageResult[*entity.Location]{}, eferrors.WithField(err, "address_id", "address not found")
	}
	return s.repo.SelectByAddressID(ctx, normAddrID, page.Normalize())
}

// ListChildren mengambil daftar sub-lokasi langsung di bawah satu lokasi induk secara berhalaman.
func (s *Service) ListChildren(ctx context.Context, parentID string, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.Location], error) {
	if err := s.ensureReady(); err != nil {
		return domainrepo.PageResult[*entity.Location]{}, err
	}
	if err := entity.ValidateUUID("parent_location_id", parentID); err != nil {
		return domainrepo.PageResult[*entity.Location]{}, err
	}
	normID := entity.NormalizeUUID(parentID)
	if _, err := s.repo.SelectByID(ctx, normID); err != nil {
		return domainrepo.PageResult[*entity.Location]{}, eferrors.WithField(err, "parent_location_id", "parent location not found")
	}
	return s.repo.SelectChildrenByParentID(ctx, normID, page.Normalize())
}

// List mengambil halaman daftar lokasi fisik sesuai filter badan hukum, induk, alamat, jenis, dan status.
func (s *Service) List(ctx context.Context, filter domainrepo.LocationFilter) (domainrepo.PageResult[*entity.Location], error) {
	if err := s.ensureReady(); err != nil {
		return domainrepo.PageResult[*entity.Location]{}, err
	}
	return s.repo.SelectList(ctx, filter.Normalize())
}

// GetAll mengambil daftar lokasi fisik non-terhapus secara berhalaman.
func (s *Service) GetAll(ctx context.Context, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.Location], error) {
	if err := s.ensureReady(); err != nil {
		return domainrepo.PageResult[*entity.Location]{}, err
	}
	return s.repo.SelectList(ctx, domainrepo.LocationFilter{Page: page.Normalize()})
}

// ValidateFacilityScope mengevaluasi kelayakan lokasi fisik dan badan hukum pemiliknya sebagai target scope facility IAM.
func (s *Service) ValidateFacilityScope(ctx context.Context, id string) (domainservice.ScopeTargetDecision, error) {
	return s.evaluateScope(ctx, id, false)
}

// ValidateWarehouseScope mengevaluasi kelayakan lokasi berjenis warehouse untuk target scope fulfillment WMS.
func (s *Service) ValidateWarehouseScope(ctx context.Context, id string) (domainservice.ScopeTargetDecision, error) {
	return s.evaluateScope(ctx, id, true)
}

// Create membentuk lokasi fisik baru setelah memeriksa badan hukum aktif, hierarki induk, alamat, dan keunikan kode lokasi.
func (s *Service) Create(ctx context.Context, in CreateInput) (*entity.Location, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}

	id, err := s.idGen.NewUUID()
	if err != nil {
		return nil, eferrors.Wrap(err, eferrors.KindUnavailable, eferrors.Code("enterprise_foundation.location.id_generation_failed"), "failed to generate location id")
	}

	loc, err := entity.NewLocation(id, entity.LocationParams{
		LegalEntityID:         in.LegalEntityID,
		ParentLocationID:      in.ParentLocationID,
		AddressID:             in.AddressID,
		LocationCode:          in.LocationCode,
		Name:                  in.Name,
		Kind:                  in.Kind,
		Phone:                 in.Phone,
		Timezone:              in.Timezone,
		OperationalAttributes: in.OperationalAttributes,
	}, s.clock.Now())
	if err != nil {
		return nil, err
	}

	if err := s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		owner, err := s.legalEntities.SelectByID(txCtx, loc.LegalEntityID)
		if err != nil {
			return eferrors.WithField(err, "legal_entity_id", "owner legal entity not found")
		}
		if err := owner.CanBeUsedInTransactions(); err != nil {
			return eferrors.WithField(err, "legal_entity_id", "owner legal entity is not active")
		}
		if err := s.validateParentAndAddress(txCtx, loc); err != nil {
			return err
		}
		existing, err := s.repo.SelectByCode(txCtx, loc.LocationCode)
		if err != nil && eferrors.KindOf(err) != eferrors.KindNotFound {
			return err
		}
		if existing != nil {
			return eferrors.ErrLocationCodeExists
		}
		if err := s.repo.Insert(txCtx, loc); err != nil {
			return err
		}
		return servicekit.EmitOutboxInTx(
			txCtx,
			s.outbox,
			s.outboxBuilder,
			entity.EventTypeLocationCreated,
			entity.AggregateTypeLocation,
			loc.ID,
			statusPayload(loc),
		)
	}); err != nil {
		return nil, err
	}
	return loc, nil
}

// Update memperbarui atribut lokasi fisik beserta validasi induk lintas badan hukum dan keberadaan alamat.
func (s *Service) Update(ctx context.Context, in UpdateInput) (*entity.Location, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("location_id", in.ID); err != nil {
		return nil, err
	}
	normID := entity.NormalizeUUID(in.ID)

	var updated *entity.Location
	if err := s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		loc, err := s.repo.SelectByID(txCtx, normID)
		if err != nil {
			return err
		}
		if err := loc.UpdateDetails(entity.LocationUpdateParams{
			ParentLocationID:      in.ParentLocationID,
			AddressID:             in.AddressID,
			Name:                  in.Name,
			Kind:                  in.Kind,
			Phone:                 in.Phone,
			Timezone:              in.Timezone,
			OperationalAttributes: in.OperationalAttributes,
		}, s.clock.Now()); err != nil {
			return err
		}
		if err := s.validateParentAndAddress(txCtx, loc); err != nil {
			return err
		}
		if err := s.repo.Update(txCtx, loc); err != nil {
			return err
		}
		if err := servicekit.EmitOutboxInTx(
			txCtx,
			s.outbox,
			s.outboxBuilder,
			entity.EventTypeLocationUpdated,
			entity.AggregateTypeLocation,
			loc.ID,
			statusPayload(loc),
		); err != nil {
			return err
		}
		updated = loc
		return nil
	}); err != nil {
		return nil, err
	}
	return updated, nil
}

// Activate mengaktifkan kembali lokasi fisik setelah memverifikasi badan hukum pemilik masih aktif.
func (s *Service) Activate(ctx context.Context, id string) (*entity.Location, error) {
	return s.changeStatus(ctx, id, true, func(loc *entity.Location, now time.Time) error {
		return loc.Activate(now)
	})
}

// Deactivate menonaktifkan sementara lokasi fisik sambil mencatat event perubahan status ke outbox.
func (s *Service) Deactivate(ctx context.Context, id string) (*entity.Location, error) {
	return s.changeStatus(ctx, id, false, func(loc *entity.Location, now time.Time) error {
		return loc.Deactivate(now)
	})
}

// Delete menghapus lunak lokasi fisik setelah memastikan tidak ada anak lokasi maupun unit kerja aktif yang bergantung.
func (s *Service) Delete(ctx context.Context, id string) error {
	if err := s.ensureReady(); err != nil {
		return err
	}
	if err := entity.ValidateUUID("location_id", id); err != nil {
		return err
	}
	normID := entity.NormalizeUUID(id)

	return s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		loc, err := s.repo.SelectByID(txCtx, normID)
		if err != nil {
			return err
		}
		dependents, err := s.repo.CountActiveDependents(txCtx, normID)
		if err != nil {
			return err
		}
		if dependents > 0 {
			return eferrors.WithField(eferrors.ErrConflict, "location_id", "location still has active dependent records")
		}
		now := s.clock.Now()
		if err := loc.SoftDelete(now); err != nil {
			return err
		}
		if err := s.repo.SoftDeleteByID(txCtx, normID, now); err != nil {
			return err
		}
		return servicekit.EmitOutboxInTx(
			txCtx,
			s.outbox,
			s.outboxBuilder,
			entity.EventTypeLocationStatusChanged,
			entity.AggregateTypeLocation,
			loc.ID,
			statusPayload(loc),
		)
	})
}

func (s *Service) evaluateScope(ctx context.Context, id string, warehouseOnly bool) (domainservice.ScopeTargetDecision, error) {
	if err := s.ensureReady(); err != nil {
		return domainservice.ScopeTargetDecision{}, err
	}
	if err := entity.ValidateUUID("location_id", id); err != nil {
		return domainservice.ScopeTargetDecision{}, err
	}
	loc, err := s.repo.SelectByID(ctx, entity.NormalizeUUID(id))
	if err != nil {
		return domainservice.ScopeTargetDecision{}, err
	}
	owner, err := s.legalEntities.SelectByID(ctx, loc.LegalEntityID)
	if err != nil {
		return domainservice.ScopeTargetDecision{}, err
	}
	if warehouseOnly {
		return s.scope.EvaluateWarehouse(loc, owner)
	}
	return s.scope.EvaluateFacility(loc, owner)
}

func (s *Service) changeStatus(
	ctx context.Context,
	id string,
	requireActiveOwner bool,
	apply func(loc *entity.Location, now time.Time) error,
) (*entity.Location, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("location_id", id); err != nil {
		return nil, err
	}
	normID := entity.NormalizeUUID(id)

	var updated *entity.Location
	if err := s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		loc, err := s.repo.SelectByID(txCtx, normID)
		if err != nil {
			return err
		}
		if requireActiveOwner {
			owner, err := s.legalEntities.SelectByID(txCtx, loc.LegalEntityID)
			if err != nil {
				return err
			}
			if err := owner.CanBeUsedInTransactions(); err != nil {
				return eferrors.WithField(err, "legal_entity_id", "owner legal entity is not active")
			}
		}
		if err := apply(loc, s.clock.Now()); err != nil {
			return err
		}
		if err := s.repo.Update(txCtx, loc); err != nil {
			return err
		}
		if err := servicekit.EmitOutboxInTx(
			txCtx,
			s.outbox,
			s.outboxBuilder,
			entity.EventTypeLocationStatusChanged,
			entity.AggregateTypeLocation,
			loc.ID,
			statusPayload(loc),
		); err != nil {
			return err
		}
		updated = loc
		return nil
	}); err != nil {
		return nil, err
	}
	return updated, nil
}

func (s *Service) validateParentAndAddress(ctx context.Context, loc *entity.Location) error {
	if loc.ParentLocationID != nil {
		parent, err := s.repo.SelectByID(ctx, *loc.ParentLocationID)
		if err != nil {
			return eferrors.WithField(err, "parent_location_id", "parent location not found")
		}
		if err := s.hierarchy.EnsureSameLegalEntityForLocation(loc.LegalEntityID, parent); err != nil {
			return err
		}
		if err := s.hierarchy.ValidateParentChain(
			ctx,
			loc.ID,
			loc.ParentLocationID,
			s.repo.SelectParentID,
			eferrors.ErrLocationSelfParent,
			eferrors.ErrLocationCycleDetected,
		); err != nil {
			return err
		}
	}
	return servicekit.EnsureAddressExists(ctx, s.addresses, "address_id", loc.AddressID)
}

func (s *Service) ensureReady() error {
	if s == nil || s.tx == nil || s.repo == nil || s.legalEntities == nil || s.addresses == nil || s.outbox == nil || s.outboxBuilder == nil || s.idGen == nil {
		return eferrors.Wrap(eferrors.ErrUnavailable, eferrors.KindUnavailable, eferrors.Code("enterprise_foundation.service.not_initialized"), "location service is not initialized")
	}
	return nil
}

func statusPayload(loc *entity.Location) domainservice.EntityStatusPayload {
	active := loc.IsActive && !loc.IsDeleted()
	status := "inactive"
	if active {
		status = "active"
	}
	return domainservice.EntityStatusPayload{
		EntityID:      loc.ID,
		LegalEntityID: loc.LegalEntityID,
		Code:          loc.LocationCode,
		Status:        status,
		IsActive:      active,
	}
}
