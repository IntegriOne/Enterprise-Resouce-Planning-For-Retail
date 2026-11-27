package handler

import (
	"context"

	efv1 "github.com/integrione/erp-retail/backend/services/enterprise_foundation/contract/gen/enterprise_foundation/v1"
	eferrors "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/repository"
	legalentitysvc "github.com/integrione/erp-retail/backend/services/enterprise_foundation/services/legalentity"
)

// LegalEntityHandler menghubungkan kontrak gRPC LegalEntityService ke layanan badan hukum domain.
type LegalEntityHandler struct {
	efv1.UnimplementedLegalEntityServiceServer
	svc       *legalentitysvc.Service
	errMapper *ErrorMapper
}

// NewLegalEntityHandler membentuk LegalEntityHandler dengan validasi fail-closed atas dependensi service.
func NewLegalEntityHandler(svc *legalentitysvc.Service, errMapper *ErrorMapper) (*LegalEntityHandler, error) {
	if svc == nil {
		return nil, eferrors.NewValidation("legal_entity_service", eferrors.Code("enterprise_foundation.handler.nil_legal_entity_service"), "legal entity service is required")
	}
	return &LegalEntityHandler{
		svc:       svc,
		errMapper: ensureErrorMapper(errMapper),
	}, nil
}

// GetLegalEntityById mengambil satu badan hukum berdasarkan UUID.
func (h *LegalEntityHandler) GetLegalEntityById(ctx context.Context, req *efv1.GetLegalEntityByIdRequest) (*efv1.GetLegalEntityByIdResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	le, err := h.svc.GetByID(ctx, req.GetId())
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.GetLegalEntityByIdResponse{LegalEntity: toProtoLegalEntity(le)}, nil
}

// GetLegalEntitiesByIds mengambil daftar badan hukum berdasarkan kumpulan UUID.
func (h *LegalEntityHandler) GetLegalEntitiesByIds(ctx context.Context, req *efv1.GetLegalEntitiesByIdsRequest) (*efv1.GetLegalEntitiesByIdsResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	items, err := h.svc.GetByIDs(ctx, req.GetIds())
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.GetLegalEntitiesByIdsResponse{Items: toProtoLegalEntities(items)}, nil
}

// GetLegalEntityByCode mengambil satu badan hukum berdasarkan kode unik entitas.
func (h *LegalEntityHandler) GetLegalEntityByCode(ctx context.Context, req *efv1.GetLegalEntityByCodeRequest) (*efv1.GetLegalEntityByCodeResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	le, err := h.svc.GetByCode(ctx, req.GetEntityCode())
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.GetLegalEntityByCodeResponse{LegalEntity: toProtoLegalEntity(le)}, nil
}

// GetLegalEntityByTaxIdNumber mengambil satu badan hukum berdasarkan NPWP.
func (h *LegalEntityHandler) GetLegalEntityByTaxIdNumber(ctx context.Context, req *efv1.GetLegalEntityByTaxIdNumberRequest) (*efv1.GetLegalEntityByTaxIdNumberResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	le, err := h.svc.GetByTaxIDNumber(ctx, req.GetTaxIdNumber())
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.GetLegalEntityByTaxIdNumberResponse{LegalEntity: toProtoLegalEntity(le)}, nil
}

// ListLegalEntitiesByNib mengambil daftar badan hukum berdasarkan NIB secara berhalaman.
func (h *LegalEntityHandler) ListLegalEntitiesByNib(ctx context.Context, req *efv1.ListLegalEntitiesByNibRequest) (*efv1.ListLegalEntitiesByNibResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	page, err := h.svc.ListByNIB(ctx, req.GetNib(), fromProtoPageRequest(req.GetPage()))
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.ListLegalEntitiesByNibResponse{
		Items: toProtoLegalEntities(page.Items),
		Page:  toProtoPageInfo(page),
	}, nil
}

// GetPrimaryOperatingLegalEntity mengambil badan hukum operasi utama perusahaan.
func (h *LegalEntityHandler) GetPrimaryOperatingLegalEntity(ctx context.Context, req *efv1.GetPrimaryOperatingLegalEntityRequest) (*efv1.GetPrimaryOperatingLegalEntityResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	le, err := h.svc.GetPrimaryOperating(ctx)
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.GetPrimaryOperatingLegalEntityResponse{LegalEntity: toProtoLegalEntity(le)}, nil
}

// ListLegalEntities mencari daftar badan hukum berdasarkan filter dan paginasi.
func (h *LegalEntityHandler) ListLegalEntities(ctx context.Context, req *efv1.ListLegalEntitiesRequest) (*efv1.ListLegalEntitiesResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	statusFilter, err := mapOptionalFilterEnum("status", req.Status, efv1.LegalEntityStatus_LEGAL_ENTITY_STATUS_UNSPECIFIED, protoToDomainLegalEntityStatus)
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	page, err := h.svc.List(ctx, domainrepo.LegalEntityFilter{
		Query:               req.GetQuery(),
		ParentLegalEntityID: clonePtr(req.ParentLegalEntityId),
		Status:              statusFilter,
		PrimaryOnly:         clonePtr(req.PrimaryOnly),
		IncludeDeleted:      req.GetIncludeDeleted(),
		Page:                fromProtoPageRequest(req.GetPage()),
	})
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.ListLegalEntitiesResponse{
		Items: toProtoLegalEntities(page.Items),
		Page:  toProtoPageInfo(page),
	}, nil
}

// GetAllLegalEntities mengambil seluruh badan hukum aktif secara berhalaman.
func (h *LegalEntityHandler) GetAllLegalEntities(ctx context.Context, req *efv1.GetAllLegalEntitiesRequest) (*efv1.GetAllLegalEntitiesResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	page, err := h.svc.GetAll(ctx, fromProtoPageRequest(req.GetPage()))
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.GetAllLegalEntitiesResponse{
		Items: toProtoLegalEntities(page.Items),
		Page:  toProtoPageInfo(page),
	}, nil
}

// ListLegalEntityChildren mengambil anak badan hukum langsung dari satu induk.
func (h *LegalEntityHandler) ListLegalEntityChildren(ctx context.Context, req *efv1.ListLegalEntityChildrenRequest) (*efv1.ListLegalEntityChildrenResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	page, err := h.svc.ListChildren(ctx, req.GetParentLegalEntityId(), fromProtoPageRequest(req.GetPage()))
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.ListLegalEntityChildrenResponse{
		Items: toProtoLegalEntities(page.Items),
		Page:  toProtoPageInfo(page),
	}, nil
}

// ValidateLegalEntityScopeTarget memeriksa kelayakan badan hukum sebagai target scope IAM secara fail-closed.
func (h *LegalEntityHandler) ValidateLegalEntityScopeTarget(ctx context.Context, req *efv1.ValidateLegalEntityScopeTargetRequest) (*efv1.ValidateLegalEntityScopeTargetResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	decision, err := h.svc.ValidateScopeTarget(ctx, req.GetId())
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.ValidateLegalEntityScopeTargetResponse{Decision: toProtoScopeDecision(decision)}, nil
}

// CreateLegalEntity membentuk badan hukum baru beserta validasi hierarki, keunikan kode/NPWP, dan outbox.
func (h *LegalEntityHandler) CreateLegalEntity(ctx context.Context, req *efv1.CreateLegalEntityRequest) (*efv1.CreateLegalEntityResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	statusVal, err := mapProtoEnum("status", req.GetStatus(), efv1.LegalEntityStatus_LEGAL_ENTITY_STATUS_UNSPECIFIED, protoToDomainLegalEntityStatus)
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	fiscalMonth, err := parseFiscalYearStartMonth(req.GetFiscalYearStartMonth())
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	incorpDate, err := fromProtoOptionalTimestamp("incorporation_date", req.GetIncorporationDate())
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	dissolDate, err := fromProtoOptionalTimestamp("dissolution_date", req.GetDissolutionDate())
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}

	created, err := h.svc.Create(ctx, legalentitysvc.CreateInput{
		ParentLegalEntityID:      clonePtr(req.ParentLegalEntityId),
		EntityCode:               req.GetEntityCode(),
		LegalName:                req.GetLegalName(),
		TradingName:              clonePtr(req.TradingName),
		LegalForm:                clonePtr(req.LegalForm),
		IsPrimaryOperatingEntity: req.GetIsPrimaryOperatingEntity(),
		Status:                   statusVal,
		JurisdictionCountryCode:  req.GetJurisdictionCountryCode(),
		NIB:                      clonePtr(req.Nib),
		TaxIDNumber:              clonePtr(req.TaxIdNumber),
		IsPKP:                    req.GetIsPkp(),
		RegisteredAddressID:      clonePtr(req.RegisteredAddressId),
		FiscalYearStartMonth:     fiscalMonth,
		IncorporationDate:        incorpDate,
		DissolutionDate:          dissolDate,
	})
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.CreateLegalEntityResponse{LegalEntity: toProtoLegalEntity(created)}, nil
}

// UpdateLegalEntity memperbarui profil dan atribut kepatuhan badan hukum.
func (h *LegalEntityHandler) UpdateLegalEntity(ctx context.Context, req *efv1.UpdateLegalEntityRequest) (*efv1.UpdateLegalEntityResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	fiscalMonth, err := parseFiscalYearStartMonth(req.GetFiscalYearStartMonth())
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	incorpDate, err := fromProtoOptionalTimestamp("incorporation_date", req.GetIncorporationDate())
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	dissolDate, err := fromProtoOptionalTimestamp("dissolution_date", req.GetDissolutionDate())
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}

	updated, err := h.svc.Update(ctx, legalentitysvc.UpdateInput{
		ID:                      req.GetId(),
		ParentLegalEntityID:     clonePtr(req.ParentLegalEntityId),
		LegalName:               req.GetLegalName(),
		TradingName:             clonePtr(req.TradingName),
		LegalForm:               clonePtr(req.LegalForm),
		JurisdictionCountryCode: req.GetJurisdictionCountryCode(),
		NIB:                     clonePtr(req.Nib),
		TaxIDNumber:             clonePtr(req.TaxIdNumber),
		IsPKP:                   req.GetIsPkp(),
		RegisteredAddressID:     clonePtr(req.RegisteredAddressId),
		FiscalYearStartMonth:    fiscalMonth,
		IncorporationDate:       incorpDate,
		DissolutionDate:         dissolDate,
	})
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.UpdateLegalEntityResponse{LegalEntity: toProtoLegalEntity(updated)}, nil
}

// SetPrimaryOperatingLegalEntity menetapkan badan hukum sebagai entitas operasi utama secara atomik.
func (h *LegalEntityHandler) SetPrimaryOperatingLegalEntity(ctx context.Context, req *efv1.SetPrimaryOperatingLegalEntityRequest) (*efv1.SetPrimaryOperatingLegalEntityResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	updated, err := h.svc.SetPrimaryOperating(ctx, req.GetId())
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.SetPrimaryOperatingLegalEntityResponse{LegalEntity: toProtoLegalEntity(updated)}, nil
}

// ActivateLegalEntity mengaktifkan badan hukum agar layak dipakai untuk transaksi dan scope IAM.
func (h *LegalEntityHandler) ActivateLegalEntity(ctx context.Context, req *efv1.ActivateLegalEntityRequest) (*efv1.ActivateLegalEntityResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	updated, err := h.svc.Activate(ctx, req.GetId())
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.ActivateLegalEntityResponse{LegalEntity: toProtoLegalEntity(updated)}, nil
}

// SuspendLegalEntity membekukan sementara badan hukum yang sebelumnya aktif.
func (h *LegalEntityHandler) SuspendLegalEntity(ctx context.Context, req *efv1.SuspendLegalEntityRequest) (*efv1.SuspendLegalEntityResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	updated, err := h.svc.Suspend(ctx, req.GetId())
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.SuspendLegalEntityResponse{LegalEntity: toProtoLegalEntity(updated)}, nil
}

// DissolveLegalEntity membubarkan badan hukum secara permanen beserta tanggal pembubaran opsional.
func (h *LegalEntityHandler) DissolveLegalEntity(ctx context.Context, req *efv1.DissolveLegalEntityRequest) (*efv1.DissolveLegalEntityResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	dissolDate, err := fromProtoOptionalTimestamp("dissolution_date", req.GetDissolutionDate())
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	updated, err := h.svc.Dissolve(ctx, req.GetId(), dissolDate)
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.DissolveLegalEntityResponse{LegalEntity: toProtoLegalEntity(updated)}, nil
}

// DeleteLegalEntity melakukan soft delete badan hukum apabila tidak lagi memiliki anak entitas.
func (h *LegalEntityHandler) DeleteLegalEntity(ctx context.Context, req *efv1.DeleteLegalEntityRequest) (*efv1.DeleteLegalEntityResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	if err := h.svc.Delete(ctx, req.GetId()); err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.DeleteLegalEntityResponse{Deleted: true}, nil
}
