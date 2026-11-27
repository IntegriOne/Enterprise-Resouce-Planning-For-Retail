// Package legalentity mengimplementasikan logika bisnis pengelolaan badan hukum perusahaan Enterprise Foundation.
package legalentity

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

// Service mengelola CRUD, hierarki induk-anak, penentuan entitas utama, validasi scope, dan outbox badan hukum.
type Service struct {
	tx            domainrepo.Transactor
	repo          domainrepo.LegalEntityRepository
	addresses     domainrepo.AddressRepository
	outbox        domainrepo.OutboxRepository
	outboxBuilder *domainservice.OutboxBuilder
	hierarchy     domainservice.HierarchyPolicy
	scope         domainservice.ScopePolicy
	clock         domainservice.Clock
	idGen         domainservice.IDGenerator
}

// NewService membentuk layanan badan hukum dengan pemeriksaan fail-closed atas seluruh dependensi wajib.
func NewService(deps Dependencies) (*Service, error) {
	if deps.Transactor == nil {
		return nil, eferrors.NewValidation("transactor", eferrors.Code("enterprise_foundation.service.nil_transactor"), "transactor is required")
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
		repo:          deps.LegalEntities,
		addresses:     deps.Addresses,
		outbox:        deps.Outbox,
		outboxBuilder: outboxBuilder,
		hierarchy:     hierarchy,
		scope:         domainservice.NewScopePolicy(),
		clock:         clock,
		idGen:         deps.IDGen,
	}, nil
}

// GetByID mengambil detail badan hukum berdasarkan UUID.
func (s *Service) GetByID(ctx context.Context, id string) (*entity.LegalEntity, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("legal_entity_id", id); err != nil {
		return nil, err
	}
	return s.repo.SelectByID(ctx, entity.NormalizeUUID(id))
}

// GetByIDs mengambil beberapa badan hukum sekaligus berdasarkan daftar UUID.
func (s *Service) GetByIDs(ctx context.Context, ids []string) ([]*entity.LegalEntity, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	normIDs, err := servicekit.ValidateOptionalUUIDList("legal_entity_ids", ids)
	if err != nil {
		return nil, err
	}
	if len(normIDs) == 0 {
		return []*entity.LegalEntity{}, nil
	}
	return s.repo.SelectByIDs(ctx, normIDs)
}

// GetByCode mencari badan hukum berdasarkan kode unik internal yang dinormalisasi.
func (s *Service) GetByCode(ctx context.Context, entityCode string) (*entity.LegalEntity, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	code := entity.NormalizeCode(entityCode)
	if code == "" {
		return nil, eferrors.WithField(eferrors.ErrInvalidLegalEntityCode, "entity_code", "legal entity code is required")
	}
	return s.repo.SelectByCode(ctx, code)
}

// GetByTaxIDNumber mencari badan hukum berdasarkan nomor NPWP/Tax ID.
func (s *Service) GetByTaxIDNumber(ctx context.Context, taxIDNumber string) (*entity.LegalEntity, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	trimmed := strings.TrimSpace(taxIDNumber)
	if trimmed == "" {
		return nil, eferrors.WithField(eferrors.ErrInvalidInput, "tax_id_number", "tax id number is required")
	}
	return s.repo.SelectByTaxIDNumber(ctx, trimmed)
}

// ListByNIB mengambil daftar badan hukum yang memiliki Nomor Induk Berusaha tertentu secara berhalaman.
func (s *Service) ListByNIB(ctx context.Context, nib string, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.LegalEntity], error) {
	if err := s.ensureReady(); err != nil {
		return domainrepo.PageResult[*entity.LegalEntity]{}, err
	}
	trimmed := strings.TrimSpace(nib)
	if trimmed == "" {
		return domainrepo.PageResult[*entity.LegalEntity]{}, eferrors.WithField(eferrors.ErrInvalidInput, "nib", "nib is required")
	}
	return s.repo.SelectByNIB(ctx, trimmed, page.Normalize())
}

// GetPrimaryOperating mengembalikan badan hukum operasional utama yang sedang aktif.
func (s *Service) GetPrimaryOperating(ctx context.Context) (*entity.LegalEntity, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	return s.repo.SelectPrimaryOperating(ctx)
}

// List mengambil halaman daftar badan hukum berdasarkan filter pencarian dan status.
func (s *Service) List(ctx context.Context, filter domainrepo.LegalEntityFilter) (domainrepo.PageResult[*entity.LegalEntity], error) {
	if err := s.ensureReady(); err != nil {
		return domainrepo.PageResult[*entity.LegalEntity]{}, err
	}
	return s.repo.SelectList(ctx, filter.Normalize())
}

// GetAll mengambil daftar badan hukum aktif secara berhalaman.
func (s *Service) GetAll(ctx context.Context, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.LegalEntity], error) {
	if err := s.ensureReady(); err != nil {
		return domainrepo.PageResult[*entity.LegalEntity]{}, err
	}
	return s.repo.SelectList(ctx, domainrepo.LegalEntityFilter{Page: page.Normalize()})
}

// ListChildren mengambil daftar badan hukum anak langsung di bawah satu induk tertentu secara berhalaman.
func (s *Service) ListChildren(ctx context.Context, parentID string, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.LegalEntity], error) {
	if err := s.ensureReady(); err != nil {
		return domainrepo.PageResult[*entity.LegalEntity]{}, err
	}
	if err := entity.ValidateUUID("parent_legal_entity_id", parentID); err != nil {
		return domainrepo.PageResult[*entity.LegalEntity]{}, err
	}
	normID := entity.NormalizeUUID(parentID)
	if _, err := s.repo.SelectByID(ctx, normID); err != nil {
		return domainrepo.PageResult[*entity.LegalEntity]{}, eferrors.WithField(err, "parent_legal_entity_id", "parent legal entity not found")
	}
	return s.repo.SelectChildrenByParentID(ctx, normID, page.Normalize())
}

// ValidateScopeTarget mengevaluasi apakah badan hukum layak digunakan sebagai target scope IAM.
func (s *Service) ValidateScopeTarget(ctx context.Context, id string) (domainservice.ScopeTargetDecision, error) {
	if err := s.ensureReady(); err != nil {
		return domainservice.ScopeTargetDecision{}, err
	}
	if err := entity.ValidateUUID("legal_entity_id", id); err != nil {
		return domainservice.ScopeTargetDecision{}, err
	}
	le, err := s.repo.SelectByID(ctx, entity.NormalizeUUID(id))
	if err != nil {
		return domainservice.ScopeTargetDecision{}, err
	}
	return s.scope.EvaluateLegalEntity(le)
}

// Create membentuk badan hukum baru setelah memeriksa keunikan kode, NPWP, status utama, induk, dan alamat terdaftar.
func (s *Service) Create(ctx context.Context, in CreateInput) (*entity.LegalEntity, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}

	id, err := s.idGen.NewUUID()
	if err != nil {
		return nil, eferrors.Wrap(err, eferrors.KindUnavailable, eferrors.Code("enterprise_foundation.legal_entity.id_generation_failed"), "failed to generate legal entity id")
	}

	le, err := entity.NewLegalEntity(id, entity.LegalEntityParams{
		ParentLegalEntityID:      in.ParentLegalEntityID,
		EntityCode:               in.EntityCode,
		LegalName:                in.LegalName,
		TradingName:              in.TradingName,
		LegalForm:                in.LegalForm,
		IsPrimaryOperatingEntity: in.IsPrimaryOperatingEntity,
		Status:                   in.Status,
		JurisdictionCountryCode:  in.JurisdictionCountryCode,
		NIB:                      in.NIB,
		TaxIDNumber:              in.TaxIDNumber,
		IsPKP:                    in.IsPKP,
		RegisteredAddressID:      in.RegisteredAddressID,
		FiscalYearStartMonth:     in.FiscalYearStartMonth,
		IncorporationDate:        in.IncorporationDate,
		DissolutionDate:          in.DissolutionDate,
	}, s.clock.Now())
	if err != nil {
		return nil, err
	}

	if err := s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := s.validateParentAndAddress(txCtx, le); err != nil {
			return err
		}
		if err := s.ensureUniqueOnCreate(txCtx, le); err != nil {
			return err
		}
		if err := s.repo.Insert(txCtx, le); err != nil {
			return err
		}
		return servicekit.EmitOutboxInTx(
			txCtx,
			s.outbox,
			s.outboxBuilder,
			entity.EventTypeLegalEntityCreated,
			entity.AggregateTypeLegalEntity,
			le.ID,
			statusPayload(le),
		)
	}); err != nil {
		return nil, err
	}
	return le, nil
}

// Update memperbarui detail badan hukum beserta validasi hierarki, alamat terdaftar, dan keunikan NPWP.
func (s *Service) Update(ctx context.Context, in UpdateInput) (*entity.LegalEntity, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("legal_entity_id", in.ID); err != nil {
		return nil, err
	}
	normID := entity.NormalizeUUID(in.ID)

	var updated *entity.LegalEntity
	if err := s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		le, err := s.repo.SelectByID(txCtx, normID)
		if err != nil {
			return err
		}
		if err := le.UpdateDetails(entity.LegalEntityUpdateParams{
			ParentLegalEntityID:     in.ParentLegalEntityID,
			LegalName:               in.LegalName,
			TradingName:             in.TradingName,
			LegalForm:               in.LegalForm,
			JurisdictionCountryCode: in.JurisdictionCountryCode,
			NIB:                     in.NIB,
			TaxIDNumber:             in.TaxIDNumber,
			IsPKP:                   in.IsPKP,
			RegisteredAddressID:     in.RegisteredAddressID,
			FiscalYearStartMonth:    in.FiscalYearStartMonth,
			IncorporationDate:       in.IncorporationDate,
			DissolutionDate:         in.DissolutionDate,
		}, s.clock.Now()); err != nil {
			return err
		}
		if err := s.validateParentAndAddress(txCtx, le); err != nil {
			return err
		}
		if le.TaxIDNumber != nil {
			exists, err := s.repo.CheckTaxIDExists(txCtx, *le.TaxIDNumber, le.ID)
			if err != nil {
				return err
			}
			if exists {
				return eferrors.ErrLegalEntityTaxIDExists
			}
		}
		if err := s.repo.Update(txCtx, le); err != nil {
			return err
		}
		if err := servicekit.EmitOutboxInTx(
			txCtx,
			s.outbox,
			s.outboxBuilder,
			entity.EventTypeLegalEntityUpdated,
			entity.AggregateTypeLegalEntity,
			le.ID,
			statusPayload(le),
		); err != nil {
			return err
		}
		updated = le
		return nil
	}); err != nil {
		return nil, err
	}
	return updated, nil
}

// SetPrimaryOperating memindahkan penanda entitas operasional utama ke badan hukum aktif yang dipilih secara atomik.
func (s *Service) SetPrimaryOperating(ctx context.Context, id string) (*entity.LegalEntity, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("legal_entity_id", id); err != nil {
		return nil, err
	}
	normID := entity.NormalizeUUID(id)

	var updated *entity.LegalEntity
	if err := s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		le, err := s.repo.SelectByID(txCtx, normID)
		if err != nil {
			return err
		}
		if err := le.CanBeUsedInTransactions(); err != nil {
			return err
		}
		now := s.clock.Now()
		if err := s.repo.UpdateClearPrimaryOperatingExcept(txCtx, le.ID, now); err != nil {
			return err
		}
		if err := le.SetPrimaryOperatingEntity(true, now); err != nil {
			return err
		}
		if err := s.repo.Update(txCtx, le); err != nil {
			return err
		}
		if err := servicekit.EmitOutboxInTx(
			txCtx,
			s.outbox,
			s.outboxBuilder,
			entity.EventTypeLegalEntityUpdated,
			entity.AggregateTypeLegalEntity,
			le.ID,
			statusPayload(le),
		); err != nil {
			return err
		}
		updated = le
		return nil
	}); err != nil {
		return nil, err
	}
	return updated, nil
}

// Activate mengaktifkan badan hukum agar dapat dipakai transaksi dan target scope.
func (s *Service) Activate(ctx context.Context, id string) (*entity.LegalEntity, error) {
	return s.changeStatus(ctx, id, func(le *entity.LegalEntity, now time.Time) error {
		return le.Activate(now)
	})
}

// Suspend membekukan sementara badan hukum sambil menjaga aturan entitas utama operasional.
func (s *Service) Suspend(ctx context.Context, id string) (*entity.LegalEntity, error) {
	return s.changeStatus(ctx, id, func(le *entity.LegalEntity, now time.Time) error {
		return le.Suspend(now)
	})
}

// Dissolve membubarkan badan hukum secara permanen setelah memastikan tidak ada tanggungan aktif.
func (s *Service) Dissolve(ctx context.Context, id string, dissolutionDate *time.Time) (*entity.LegalEntity, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("legal_entity_id", id); err != nil {
		return nil, err
	}
	normID := entity.NormalizeUUID(id)

	var updated *entity.LegalEntity
	if err := s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		le, err := s.repo.SelectByID(txCtx, normID)
		if err != nil {
			return err
		}
		if err := le.Dissolve(dissolutionDate, s.clock.Now()); err != nil {
			return err
		}
		dependents, err := s.repo.CountActiveDependents(txCtx, normID)
		if err != nil {
			return err
		}
		if dependents > 0 {
			return eferrors.ErrLegalEntityHasDependents
		}
		if err := s.repo.Update(txCtx, le); err != nil {
			return err
		}
		if err := servicekit.EmitOutboxInTx(
			txCtx,
			s.outbox,
			s.outboxBuilder,
			entity.EventTypeLegalEntityStatusChanged,
			entity.AggregateTypeLegalEntity,
			le.ID,
			statusPayload(le),
		); err != nil {
			return err
		}
		updated = le
		return nil
	}); err != nil {
		return nil, err
	}
	return updated, nil
}

// Delete menghapus lunak badan hukum bila bukan entitas operasional utama dan bebas dari tanggungan aktif.
func (s *Service) Delete(ctx context.Context, id string) error {
	if err := s.ensureReady(); err != nil {
		return err
	}
	if err := entity.ValidateUUID("legal_entity_id", id); err != nil {
		return err
	}
	normID := entity.NormalizeUUID(id)

	return s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		le, err := s.repo.SelectByID(txCtx, normID)
		if err != nil {
			return err
		}
		dependents, err := s.repo.CountActiveDependents(txCtx, normID)
		if err != nil {
			return err
		}
		if err := s.hierarchy.ValidateLegalEntityRemoval(le, dependents); err != nil {
			return err
		}
		now := s.clock.Now()
		if err := le.SoftDelete(now); err != nil {
			return err
		}
		if err := s.repo.SoftDeleteByID(txCtx, normID, now); err != nil {
			return err
		}
		return servicekit.EmitOutboxInTx(
			txCtx,
			s.outbox,
			s.outboxBuilder,
			entity.EventTypeLegalEntityStatusChanged,
			entity.AggregateTypeLegalEntity,
			le.ID,
			statusPayload(le),
		)
	})
}

func (s *Service) changeStatus(
	ctx context.Context,
	id string,
	apply func(le *entity.LegalEntity, now time.Time) error,
) (*entity.LegalEntity, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("legal_entity_id", id); err != nil {
		return nil, err
	}
	normID := entity.NormalizeUUID(id)

	var updated *entity.LegalEntity
	if err := s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		le, err := s.repo.SelectByID(txCtx, normID)
		if err != nil {
			return err
		}
		if err := apply(le, s.clock.Now()); err != nil {
			return err
		}
		if err := s.repo.Update(txCtx, le); err != nil {
			return err
		}
		if err := servicekit.EmitOutboxInTx(
			txCtx,
			s.outbox,
			s.outboxBuilder,
			entity.EventTypeLegalEntityStatusChanged,
			entity.AggregateTypeLegalEntity,
			le.ID,
			statusPayload(le),
		); err != nil {
			return err
		}
		updated = le
		return nil
	}); err != nil {
		return nil, err
	}
	return updated, nil
}

func (s *Service) validateParentAndAddress(ctx context.Context, le *entity.LegalEntity) error {
	if le.ParentLegalEntityID != nil {
		if _, err := s.repo.SelectByID(ctx, *le.ParentLegalEntityID); err != nil {
			return eferrors.WithField(err, "parent_legal_entity_id", "parent legal entity not found")
		}
		if err := s.hierarchy.ValidateParentChain(
			ctx,
			le.ID,
			le.ParentLegalEntityID,
			s.repo.SelectParentID,
			eferrors.ErrLegalEntitySelfParent,
			eferrors.ErrLegalEntityCycleDetected,
		); err != nil {
			return err
		}
	}
	return servicekit.EnsureAddressExists(ctx, s.addresses, "registered_address_id", le.RegisteredAddressID)
}

func (s *Service) ensureUniqueOnCreate(ctx context.Context, le *entity.LegalEntity) error {
	existingByCode, err := s.repo.SelectByCode(ctx, le.EntityCode)
	if err != nil && eferrors.KindOf(err) != eferrors.KindNotFound {
		return err
	}
	if existingByCode != nil {
		return eferrors.ErrLegalEntityCodeExists
	}

	if le.TaxIDNumber != nil {
		existingByTax, err := s.repo.SelectByTaxIDNumber(ctx, *le.TaxIDNumber)
		if err != nil && eferrors.KindOf(err) != eferrors.KindNotFound {
			return err
		}
		if existingByTax != nil {
			return eferrors.ErrLegalEntityTaxIDExists
		}
	}

	if le.IsPrimaryOperatingEntity {
		existingPrimary, err := s.repo.SelectPrimaryOperating(ctx)
		if err != nil && eferrors.KindOf(err) != eferrors.KindNotFound {
			return err
		}
		if existingPrimary != nil {
			return eferrors.ErrPrimaryLegalEntityExists
		}
	}
	return nil
}

func (s *Service) ensureReady() error {
	if s == nil || s.tx == nil || s.repo == nil || s.addresses == nil || s.outbox == nil || s.outboxBuilder == nil || s.idGen == nil {
		return eferrors.Wrap(eferrors.ErrUnavailable, eferrors.KindUnavailable, eferrors.Code("enterprise_foundation.service.not_initialized"), "legal entity service is not initialized")
	}
	return nil
}

func statusPayload(le *entity.LegalEntity) domainservice.EntityStatusPayload {
	return domainservice.EntityStatusPayload{
		EntityID:      le.ID,
		LegalEntityID: le.ID,
		Code:          le.EntityCode,
		Status:        string(le.Status),
		IsActive:      le.Status == entity.LegalEntityStatusActive && !le.IsDeleted(),
	}
}
