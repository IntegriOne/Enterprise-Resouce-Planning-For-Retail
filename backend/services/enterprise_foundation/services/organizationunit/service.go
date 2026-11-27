// Package organizationunit mengimplementasikan logika bisnis pengelolaan unit organisasi internal Enterprise Foundation.
package organizationunit

import (
	"context"
	"time"

	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/entity"
	eferrors "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/repository"
	domainservice "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/service"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/services/internal/servicekit"
)

// Service mengelola CRUD, hierarki pohon unit kerja, lokasi utama lintas badan hukum, penugasan kepala unit, dan scope IAM.
type Service struct {
	tx            domainrepo.Transactor
	repo          domainrepo.OrganizationUnitRepository
	legalEntities domainrepo.LegalEntityRepository
	locations     domainrepo.LocationRepository
	outbox        domainrepo.OutboxRepository
	outboxBuilder *domainservice.OutboxBuilder
	hierarchy     domainservice.HierarchyPolicy
	scope         domainservice.ScopePolicy
	clock         domainservice.Clock
	idGen         domainservice.IDGenerator
}

// NewService membentuk layanan unit organisasi dengan validasi fail-closed atas seluruh dependensi wajib.
func NewService(deps Dependencies) (*Service, error) {
	if deps.Transactor == nil {
		return nil, eferrors.NewValidation("transactor", eferrors.Code("enterprise_foundation.service.nil_transactor"), "transactor is required")
	}
	if deps.OrganizationUnits == nil {
		return nil, eferrors.NewValidation("organization_units", eferrors.Code("enterprise_foundation.service.nil_org_unit_repo"), "organization unit repository is required")
	}
	if deps.LegalEntities == nil {
		return nil, eferrors.NewValidation("legal_entities", eferrors.Code("enterprise_foundation.service.nil_legal_entity_repo"), "legal entity repository is required")
	}
	if deps.Locations == nil {
		return nil, eferrors.NewValidation("locations", eferrors.Code("enterprise_foundation.service.nil_location_repo"), "location repository is required")
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
		repo:          deps.OrganizationUnits,
		legalEntities: deps.LegalEntities,
		locations:     deps.Locations,
		outbox:        deps.Outbox,
		outboxBuilder: outboxBuilder,
		hierarchy:     hierarchy,
		scope:         domainservice.NewScopePolicy(),
		clock:         clock,
		idGen:         deps.IDGen,
	}, nil
}

// GetByID mengambil detail satu unit organisasi berdasarkan UUID.
func (s *Service) GetByID(ctx context.Context, id string) (*entity.OrganizationUnit, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("organization_unit_id", id); err != nil {
		return nil, err
	}
	return s.repo.SelectByID(ctx, entity.NormalizeUUID(id))
}

// GetByIDs mengambil beberapa unit organisasi sekaligus berdasarkan daftar UUID.
func (s *Service) GetByIDs(ctx context.Context, ids []string) ([]*entity.OrganizationUnit, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	normIDs, err := servicekit.ValidateOptionalUUIDList("organization_unit_ids", ids)
	if err != nil {
		return nil, err
	}
	if len(normIDs) == 0 {
		return []*entity.OrganizationUnit{}, nil
	}
	return s.repo.SelectByIDs(ctx, normIDs)
}

// GetByEntityAndCode mencari unit organisasi berdasarkan badan hukum pemilik dan kode unit ternormalisasi.
func (s *Service) GetByEntityAndCode(ctx context.Context, legalEntityID, unitCode string) (*entity.OrganizationUnit, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("legal_entity_id", legalEntityID); err != nil {
		return nil, err
	}
	code := entity.NormalizeCode(unitCode)
	if code == "" {
		return nil, eferrors.WithField(eferrors.ErrInvalidOrgUnitCode, "unit_code", "organization unit code is required")
	}
	return s.repo.SelectByEntityAndCode(ctx, entity.NormalizeUUID(legalEntityID), code)
}

// ListByHeadPersonID mengambil daftar unit organisasi yang dipimpin oleh satu ID kepala unit secara berhalaman.
func (s *Service) ListByHeadPersonID(ctx context.Context, headPersonID string, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.OrganizationUnit], error) {
	if err := s.ensureReady(); err != nil {
		return domainrepo.PageResult[*entity.OrganizationUnit]{}, err
	}
	if err := entity.ValidateUUID("head_person_id", headPersonID); err != nil {
		return domainrepo.PageResult[*entity.OrganizationUnit]{}, err
	}
	return s.repo.SelectByHeadPersonID(ctx, entity.NormalizeUUID(headPersonID), page.Normalize())
}

// ListByPrimaryLocationID mengambil daftar unit organisasi yang berpusat pada satu lokasi fisik secara berhalaman.
func (s *Service) ListByPrimaryLocationID(ctx context.Context, primaryLocationID string, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.OrganizationUnit], error) {
	if err := s.ensureReady(); err != nil {
		return domainrepo.PageResult[*entity.OrganizationUnit]{}, err
	}
	if err := entity.ValidateUUID("primary_location_id", primaryLocationID); err != nil {
		return domainrepo.PageResult[*entity.OrganizationUnit]{}, err
	}
	normLocID := entity.NormalizeUUID(primaryLocationID)
	if _, err := s.locations.SelectByID(ctx, normLocID); err != nil {
		return domainrepo.PageResult[*entity.OrganizationUnit]{}, eferrors.WithField(err, "primary_location_id", "primary location not found")
	}
	return s.repo.SelectByPrimaryLocationID(ctx, normLocID, page.Normalize())
}

// ListChildren mengambil daftar sub-unit langsung di bawah satu unit organisasi induk secara berhalaman.
func (s *Service) ListChildren(ctx context.Context, parentID string, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.OrganizationUnit], error) {
	if err := s.ensureReady(); err != nil {
		return domainrepo.PageResult[*entity.OrganizationUnit]{}, err
	}
	if err := entity.ValidateUUID("parent_unit_id", parentID); err != nil {
		return domainrepo.PageResult[*entity.OrganizationUnit]{}, err
	}
	normID := entity.NormalizeUUID(parentID)
	if _, err := s.repo.SelectByID(ctx, normID); err != nil {
		return domainrepo.PageResult[*entity.OrganizationUnit]{}, eferrors.WithField(err, "parent_unit_id", "parent organization unit not found")
	}
	return s.repo.SelectChildrenByParentID(ctx, normID, page.Normalize())
}

// List mengambil halaman daftar unit organisasi sesuai filter pencarian, badan hukum, induk, lokasi, jenis, dan status.
func (s *Service) List(ctx context.Context, filter domainrepo.OrganizationUnitFilter) (domainrepo.PageResult[*entity.OrganizationUnit], error) {
	if err := s.ensureReady(); err != nil {
		return domainrepo.PageResult[*entity.OrganizationUnit]{}, err
	}
	return s.repo.SelectList(ctx, filter.Normalize())
}

// GetAll mengambil daftar unit organisasi aktif/non-terhapus secara berhalaman.
func (s *Service) GetAll(ctx context.Context, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.OrganizationUnit], error) {
	if err := s.ensureReady(); err != nil {
		return domainrepo.PageResult[*entity.OrganizationUnit]{}, err
	}
	return s.repo.SelectList(ctx, domainrepo.OrganizationUnitFilter{Page: page.Normalize()})
}

// ValidateScopeTarget mengevaluasi kelayakan unit organisasi dan badan hukum pemiliknya sebagai target scope IAM.
func (s *Service) ValidateScopeTarget(ctx context.Context, id string) (domainservice.ScopeTargetDecision, error) {
	if err := s.ensureReady(); err != nil {
		return domainservice.ScopeTargetDecision{}, err
	}
	if err := entity.ValidateUUID("organization_unit_id", id); err != nil {
		return domainservice.ScopeTargetDecision{}, err
	}
	unit, err := s.repo.SelectByID(ctx, entity.NormalizeUUID(id))
	if err != nil {
		return domainservice.ScopeTargetDecision{}, err
	}
	owner, err := s.legalEntities.SelectByID(ctx, unit.LegalEntityID)
	if err != nil {
		return domainservice.ScopeTargetDecision{}, err
	}
	return s.scope.EvaluateOrganizationUnit(unit, owner)
}

// Create membentuk unit organisasi baru setelah memverifikasi badan hukum aktif, induk/lokasi satu entitas, dan keunikan kode unit.
func (s *Service) Create(ctx context.Context, in CreateInput) (*entity.OrganizationUnit, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}

	id, err := s.idGen.NewUUID()
	if err != nil {
		return nil, eferrors.Wrap(err, eferrors.KindUnavailable, eferrors.Code("enterprise_foundation.organization_unit.id_generation_failed"), "failed to generate organization unit id")
	}

	unit, err := entity.NewOrganizationUnit(id, entity.OrganizationUnitParams{
		LegalEntityID:     in.LegalEntityID,
		ParentUnitID:      in.ParentUnitID,
		PrimaryLocationID: in.PrimaryLocationID,
		UnitCode:          in.UnitCode,
		Name:              in.Name,
		Kind:              in.Kind,
		CostCenterCode:    in.CostCenterCode,
		HeadPersonID:      in.HeadPersonID,
	}, s.clock.Now())
	if err != nil {
		return nil, err
	}

	if err := s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		owner, err := s.legalEntities.SelectByID(txCtx, unit.LegalEntityID)
		if err != nil {
			return eferrors.WithField(err, "legal_entity_id", "owner legal entity not found")
		}
		if err := owner.CanBeUsedInTransactions(); err != nil {
			return eferrors.WithField(err, "legal_entity_id", "owner legal entity is not active")
		}
		if err := s.validateParentAndPrimaryLocation(txCtx, unit); err != nil {
			return err
		}
		existing, err := s.repo.SelectByEntityAndCode(txCtx, unit.LegalEntityID, unit.UnitCode)
		if err != nil && eferrors.KindOf(err) != eferrors.KindNotFound {
			return err
		}
		if existing != nil {
			return eferrors.ErrOrgUnitCodeExists
		}
		if err := s.repo.Insert(txCtx, unit); err != nil {
			return err
		}
		return servicekit.EmitOutboxInTx(
			txCtx,
			s.outbox,
			s.outboxBuilder,
			entity.EventTypeOrgUnitCreated,
			entity.AggregateTypeOrganizationUnit,
			unit.ID,
			statusPayload(unit),
		)
	}); err != nil {
		return nil, err
	}
	return unit, nil
}

// Update memperbarui atribut unit organisasi beserta validasi induk dan lokasi utama lintas badan hukum.
func (s *Service) Update(ctx context.Context, in UpdateInput) (*entity.OrganizationUnit, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("organization_unit_id", in.ID); err != nil {
		return nil, err
	}
	if in.PreserveHeadPerson && in.HeadPersonID != nil {
		return nil, eferrors.NewValidation("head_person_id", eferrors.Code("enterprise_foundation.organization_unit.conflicting_head_update"), "head person cannot be supplied when preserving the current assignment")
	}
	normID := entity.NormalizeUUID(in.ID)

	var updated *entity.OrganizationUnit
	if err := s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		unit, err := s.repo.SelectByIDForUpdate(txCtx, normID)
		if err != nil {
			return err
		}
		headPersonID := in.HeadPersonID
		if in.PreserveHeadPerson {
			headPersonID = unit.HeadPersonID
		}
		if err := unit.UpdateDetails(entity.OrganizationUnitUpdateParams{
			ParentUnitID:      in.ParentUnitID,
			PrimaryLocationID: in.PrimaryLocationID,
			Name:              in.Name,
			Kind:              in.Kind,
			CostCenterCode:    in.CostCenterCode,
			HeadPersonID:      headPersonID,
		}, s.clock.Now()); err != nil {
			return err
		}
		if err := s.validateParentAndPrimaryLocation(txCtx, unit); err != nil {
			return err
		}
		if err := s.repo.Update(txCtx, unit); err != nil {
			return err
		}
		if err := servicekit.EmitOutboxInTx(
			txCtx,
			s.outbox,
			s.outboxBuilder,
			entity.EventTypeOrgUnitUpdated,
			entity.AggregateTypeOrganizationUnit,
			unit.ID,
			statusPayload(unit),
		); err != nil {
			return err
		}
		updated = unit
		return nil
	}); err != nil {
		return nil, err
	}
	return updated, nil
}

// AssignHeadPerson menetapkan atau melepas ID penanggung jawab unit organisasi secara transaksional.
func (s *Service) AssignHeadPerson(ctx context.Context, id string, headPersonID *string) (*entity.OrganizationUnit, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("organization_unit_id", id); err != nil {
		return nil, err
	}
	normID := entity.NormalizeUUID(id)

	var updated *entity.OrganizationUnit
	if err := s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		unit, err := s.repo.SelectByIDForUpdate(txCtx, normID)
		if err != nil {
			return err
		}
		if err := unit.AssignHeadPerson(headPersonID, s.clock.Now()); err != nil {
			return err
		}
		if err := s.repo.Update(txCtx, unit); err != nil {
			return err
		}
		if err := servicekit.EmitOutboxInTx(
			txCtx,
			s.outbox,
			s.outboxBuilder,
			entity.EventTypeOrgUnitUpdated,
			entity.AggregateTypeOrganizationUnit,
			unit.ID,
			statusPayload(unit),
		); err != nil {
			return err
		}
		updated = unit
		return nil
	}); err != nil {
		return nil, err
	}
	return updated, nil
}

// Activate mengaktifkan kembali unit organisasi setelah memastikan badan hukum pemilik tetap aktif.
func (s *Service) Activate(ctx context.Context, id string) (*entity.OrganizationUnit, error) {
	return s.changeStatus(ctx, id, true, func(unit *entity.OrganizationUnit, now time.Time) error {
		return unit.Activate(now)
	})
}

// Deactivate menonaktifkan sementara unit organisasi dan menerbitkan event perubahan status.
func (s *Service) Deactivate(ctx context.Context, id string) (*entity.OrganizationUnit, error) {
	return s.changeStatus(ctx, id, false, func(unit *entity.OrganizationUnit, now time.Time) error {
		return unit.Deactivate(now)
	})
}

// Delete menghapus lunak unit organisasi setelah memastikan tidak memiliki sub-unit aktif yang bergantung.
func (s *Service) Delete(ctx context.Context, id string) error {
	if err := s.ensureReady(); err != nil {
		return err
	}
	if err := entity.ValidateUUID("organization_unit_id", id); err != nil {
		return err
	}
	normID := entity.NormalizeUUID(id)

	return s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		unit, err := s.repo.SelectByIDForUpdate(txCtx, normID)
		if err != nil {
			return err
		}
		dependents, err := s.repo.CountActiveDependents(txCtx, normID)
		if err != nil {
			return err
		}
		if dependents > 0 {
			return eferrors.WithField(eferrors.ErrConflict, "organization_unit_id", "organization unit still has active dependent records")
		}
		now := s.clock.Now()
		if err := unit.SoftDelete(now); err != nil {
			return err
		}
		if err := s.repo.SoftDeleteByID(txCtx, normID, now); err != nil {
			return err
		}
		return servicekit.EmitOutboxInTx(
			txCtx,
			s.outbox,
			s.outboxBuilder,
			entity.EventTypeOrgUnitStatusChanged,
			entity.AggregateTypeOrganizationUnit,
			unit.ID,
			statusPayload(unit),
		)
	})
}

func (s *Service) changeStatus(
	ctx context.Context,
	id string,
	requireActiveOwner bool,
	apply func(unit *entity.OrganizationUnit, now time.Time) error,
) (*entity.OrganizationUnit, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("organization_unit_id", id); err != nil {
		return nil, err
	}
	normID := entity.NormalizeUUID(id)

	var updated *entity.OrganizationUnit
	if err := s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		unit, err := s.repo.SelectByIDForUpdate(txCtx, normID)
		if err != nil {
			return err
		}
		if requireActiveOwner {
			owner, err := s.legalEntities.SelectByID(txCtx, unit.LegalEntityID)
			if err != nil {
				return err
			}
			if err := owner.CanBeUsedInTransactions(); err != nil {
				return eferrors.WithField(err, "legal_entity_id", "owner legal entity is not active")
			}
		}
		if err := apply(unit, s.clock.Now()); err != nil {
			return err
		}
		if err := s.repo.Update(txCtx, unit); err != nil {
			return err
		}
		if err := servicekit.EmitOutboxInTx(
			txCtx,
			s.outbox,
			s.outboxBuilder,
			entity.EventTypeOrgUnitStatusChanged,
			entity.AggregateTypeOrganizationUnit,
			unit.ID,
			statusPayload(unit),
		); err != nil {
			return err
		}
		updated = unit
		return nil
	}); err != nil {
		return nil, err
	}
	return updated, nil
}

func (s *Service) validateParentAndPrimaryLocation(ctx context.Context, unit *entity.OrganizationUnit) error {
	var parentUnit *entity.OrganizationUnit
	if unit.ParentUnitID != nil {
		var err error
		parentUnit, err = s.repo.SelectByID(ctx, *unit.ParentUnitID)
		if err != nil {
			return eferrors.WithField(err, "parent_unit_id", "parent organization unit not found")
		}
	}

	var primaryLocation *entity.Location
	if unit.PrimaryLocationID != nil {
		var err error
		primaryLocation, err = s.locations.SelectByID(ctx, *unit.PrimaryLocationID)
		if err != nil {
			return eferrors.WithField(err, "primary_location_id", "primary location not found")
		}
	}

	if err := s.hierarchy.EnsureSameLegalEntityForOrgUnit(unit.LegalEntityID, parentUnit, primaryLocation); err != nil {
		return err
	}
	if unit.ParentUnitID != nil {
		return s.hierarchy.ValidateParentChain(
			ctx,
			unit.ID,
			unit.ParentUnitID,
			s.repo.SelectParentID,
			eferrors.ErrOrgUnitSelfParent,
			eferrors.ErrOrgUnitCycleDetected,
		)
	}
	return nil
}

func (s *Service) ensureReady() error {
	if s == nil || s.tx == nil || s.repo == nil || s.legalEntities == nil || s.locations == nil || s.outbox == nil || s.outboxBuilder == nil || s.idGen == nil {
		return eferrors.Wrap(eferrors.ErrUnavailable, eferrors.KindUnavailable, eferrors.Code("enterprise_foundation.service.not_initialized"), "organization unit service is not initialized")
	}
	return nil
}

func statusPayload(unit *entity.OrganizationUnit) domainservice.EntityStatusPayload {
	active := unit.IsActive && !unit.IsDeleted()
	status := "inactive"
	if active {
		status = "active"
	}
	return domainservice.EntityStatusPayload{
		EntityID:      unit.ID,
		LegalEntityID: unit.LegalEntityID,
		Code:          unit.UnitCode,
		Status:        status,
		IsActive:      active,
	}
}
