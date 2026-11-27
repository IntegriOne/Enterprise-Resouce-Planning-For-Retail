package handler

import (
	"context"

	efv1 "github.com/integrione/erp-retail/backend/services/enterprise_foundation/contract/gen/enterprise_foundation/v1"
	eferrors "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/repository"
	partysvc "github.com/integrione/erp-retail/backend/services/enterprise_foundation/services/party"
)

// PartyHandler menghubungkan kontrak gRPC PartyService ke layanan mitra bisnis eksternal, peran, PIC portal, dan alamat.
type PartyHandler struct {
	efv1.UnimplementedPartyServiceServer
	svc       *partysvc.Service
	errMapper *ErrorMapper
}

// NewPartyHandler membentuk PartyHandler dengan pemeriksaan fail-closed atas dependensi service.
func NewPartyHandler(svc *partysvc.Service, errMapper *ErrorMapper) (*PartyHandler, error) {
	if svc == nil {
		return nil, eferrors.NewValidation("party_service", eferrors.Code("enterprise_foundation.handler.nil_party_service"), "party service is required")
	}
	return &PartyHandler{
		svc:       svc,
		errMapper: ensureErrorMapper(errMapper),
	}, nil
}

// GetPartyById mengambil satu mitra eksternal berdasarkan UUID.
func (h *PartyHandler) GetPartyById(ctx context.Context, req *efv1.GetPartyByIdRequest) (*efv1.GetPartyByIdResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	p, err := h.svc.GetByID(ctx, req.GetId())
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.GetPartyByIdResponse{Party: toProtoParty(p)}, nil
}

// GetPartiesByIds mengambil daftar mitra eksternal berdasarkan kumpulan UUID.
func (h *PartyHandler) GetPartiesByIds(ctx context.Context, req *efv1.GetPartiesByIdsRequest) (*efv1.GetPartiesByIdsResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	items, err := h.svc.GetByIDs(ctx, req.GetIds())
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.GetPartiesByIdsResponse{Items: toProtoParties(items)}, nil
}

// GetPartyByCode mengambil satu mitra eksternal berdasarkan kode unik partner.
func (h *PartyHandler) GetPartyByCode(ctx context.Context, req *efv1.GetPartyByCodeRequest) (*efv1.GetPartyByCodeResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	p, err := h.svc.GetByCode(ctx, req.GetPartnerCode())
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.GetPartyByCodeResponse{Party: toProtoParty(p)}, nil
}

// GetPartyByTaxIdNumber mengambil satu mitra eksternal berdasarkan NPWP.
func (h *PartyHandler) GetPartyByTaxIdNumber(ctx context.Context, req *efv1.GetPartyByTaxIdNumberRequest) (*efv1.GetPartyByTaxIdNumberResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	p, err := h.svc.GetByTaxIDNumber(ctx, req.GetTaxIdNumber())
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.GetPartyByTaxIdNumberResponse{Party: toProtoParty(p)}, nil
}

// ListPartiesByEmail mengambil daftar mitra berdasarkan email secara berhalaman.
func (h *PartyHandler) ListPartiesByEmail(ctx context.Context, req *efv1.ListPartiesByEmailRequest) (*efv1.ListPartiesByEmailResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	page, err := h.svc.ListByEmail(ctx, req.GetEmail(), fromProtoPageRequest(req.GetPage()))
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.ListPartiesByEmailResponse{
		Items: toProtoParties(page.Items),
		Page:  toProtoPageInfo(page),
	}, nil
}

// ListParties mencari mitra eksternal berdasarkan filter dan parameter paginasi.
func (h *PartyHandler) ListParties(ctx context.Context, req *efv1.ListPartiesRequest) (*efv1.ListPartiesResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	kindFilter, err := mapOptionalFilterEnum("kind", req.Kind, efv1.PartyKind_PARTY_KIND_UNSPECIFIED, protoToDomainPartyKind)
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	statusFilter, err := mapOptionalFilterEnum("status", req.Status, efv1.PartyStatus_PARTY_STATUS_UNSPECIFIED, protoToDomainPartyStatus)
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	roleFilter, err := mapOptionalFilterEnum("role", req.Role, efv1.PartyRoleKind_PARTY_ROLE_KIND_UNSPECIFIED, protoToDomainPartyRoleKind)
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}

	page, err := h.svc.List(ctx, domainrepo.PartyFilter{
		Query:          req.GetQuery(),
		Kind:           kindFilter,
		Status:         statusFilter,
		Role:           roleFilter,
		IncludeDeleted: req.GetIncludeDeleted(),
		Page:           fromProtoPageRequest(req.GetPage()),
	})
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.ListPartiesResponse{
		Items: toProtoParties(page.Items),
		Page:  toProtoPageInfo(page),
	}, nil
}

// GetAllParties mengambil seluruh mitra aktif secara berhalaman.
func (h *PartyHandler) GetAllParties(ctx context.Context, req *efv1.GetAllPartiesRequest) (*efv1.GetAllPartiesResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	page, err := h.svc.GetAll(ctx, fromProtoPageRequest(req.GetPage()))
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.GetAllPartiesResponse{
		Items: toProtoParties(page.Items),
		Page:  toProtoPageInfo(page),
	}, nil
}

// CreateParty mendaftarkan mitra bisnis eksternal baru setelah validasi keunikan kode dan NPWP.
func (h *PartyHandler) CreateParty(ctx context.Context, req *efv1.CreatePartyRequest) (*efv1.CreatePartyResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	kind, err := mapProtoEnum("kind", req.GetKind(), efv1.PartyKind_PARTY_KIND_UNSPECIFIED, protoToDomainPartyKind)
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	statusVal, err := mapProtoEnum("status", req.GetStatus(), efv1.PartyStatus_PARTY_STATUS_UNSPECIFIED, protoToDomainPartyStatus)
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	taxVerifiedAt, err := fromProtoOptionalTimestamp("tax_verified_at", req.GetTaxVerifiedAt())
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}

	created, err := h.svc.Create(ctx, partysvc.CreateInput{
		PartnerCode:        req.GetPartnerCode(),
		Kind:               kind,
		DisplayName:        req.GetDisplayName(),
		LegalName:          clonePtr(req.LegalName),
		Email:              clonePtr(req.Email),
		Phone:              clonePtr(req.Phone),
		WebsiteURL:         clonePtr(req.WebsiteUrl),
		TaxIDNumber:        clonePtr(req.TaxIdNumber),
		TaxRegisteredName:  clonePtr(req.TaxRegisteredName),
		IsPKP:              req.GetIsPkp(),
		IsTaxExempt:        req.GetIsTaxExempt(),
		TaxExemptionReason: clonePtr(req.TaxExemptionReason),
		TaxVerifiedAt:      taxVerifiedAt,
		PreferredLanguage:  req.GetPreferredLanguage(),
		Status:             statusVal,
		Notes:              clonePtr(req.Notes),
	})
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.CreatePartyResponse{Party: toProtoParty(created)}, nil
}

// UpdatePartyProfile memperbarui profil umum mitra bisnis eksternal.
func (h *PartyHandler) UpdatePartyProfile(ctx context.Context, req *efv1.UpdatePartyProfileRequest) (*efv1.UpdatePartyProfileResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	kind, err := mapProtoEnum("kind", req.GetKind(), efv1.PartyKind_PARTY_KIND_UNSPECIFIED, protoToDomainPartyKind)
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	updated, err := h.svc.UpdateProfile(ctx, partysvc.UpdateProfileInput{
		ID:                req.GetId(),
		Kind:              kind,
		DisplayName:       req.GetDisplayName(),
		LegalName:         clonePtr(req.LegalName),
		Email:             clonePtr(req.Email),
		Phone:             clonePtr(req.Phone),
		WebsiteURL:        clonePtr(req.WebsiteUrl),
		PreferredLanguage: req.GetPreferredLanguage(),
		Notes:             clonePtr(req.Notes),
	})
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.UpdatePartyProfileResponse{Party: toProtoParty(updated)}, nil
}

// UpdatePartyTaxProfile memperbarui atribut perpajakan dan status PKP mitra.
func (h *PartyHandler) UpdatePartyTaxProfile(ctx context.Context, req *efv1.UpdatePartyTaxProfileRequest) (*efv1.UpdatePartyTaxProfileResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	taxVerifiedAt, err := fromProtoOptionalTimestamp("tax_verified_at", req.GetTaxVerifiedAt())
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	updated, err := h.svc.UpdateTaxProfile(ctx, partysvc.UpdateTaxProfileInput{
		ID:                 req.GetId(),
		TaxIDNumber:        clonePtr(req.TaxIdNumber),
		TaxRegisteredName:  clonePtr(req.TaxRegisteredName),
		IsPKP:              req.GetIsPkp(),
		IsTaxExempt:        req.GetIsTaxExempt(),
		TaxExemptionReason: clonePtr(req.TaxExemptionReason),
		TaxVerifiedAt:      taxVerifiedAt,
	})
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.UpdatePartyTaxProfileResponse{Party: toProtoParty(updated)}, nil
}

// VerifyAndActivateParty memverifikasi dan mengaktifkan mitra agar layak bertransaksi dan dijadikan scope.
func (h *PartyHandler) VerifyAndActivateParty(ctx context.Context, req *efv1.VerifyAndActivatePartyRequest) (*efv1.VerifyAndActivatePartyResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	updated, err := h.svc.VerifyAndActivate(ctx, req.GetId())
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.VerifyAndActivatePartyResponse{Party: toProtoParty(updated)}, nil
}

// SuspendParty menangguhkan sementara mitra dari transaksi baru.
func (h *PartyHandler) SuspendParty(ctx context.Context, req *efv1.SuspendPartyRequest) (*efv1.SuspendPartyResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	updated, err := h.svc.Suspend(ctx, req.GetId())
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.SuspendPartyResponse{Party: toProtoParty(updated)}, nil
}

// DeactivateParty menonaktifkan mitra secara terkendali.
func (h *PartyHandler) DeactivateParty(ctx context.Context, req *efv1.DeactivatePartyRequest) (*efv1.DeactivatePartyResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	updated, err := h.svc.Deactivate(ctx, req.GetId())
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.DeactivatePartyResponse{Party: toProtoParty(updated)}, nil
}

// DeleteParty melakukan soft delete mitra eksternal.
func (h *PartyHandler) DeleteParty(ctx context.Context, req *efv1.DeletePartyRequest) (*efv1.DeletePartyResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	if err := h.svc.Delete(ctx, req.GetId()); err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.DeletePartyResponse{Deleted: true}, nil
}
