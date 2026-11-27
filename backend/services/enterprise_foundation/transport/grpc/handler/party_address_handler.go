package handler

import (
	"context"

	efv1 "github.com/integrione/erp-retail/backend/services/enterprise_foundation/contract/gen/enterprise_foundation/v1"
	domainrepo "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/repository"
	partysvc "github.com/integrione/erp-retail/backend/services/enterprise_foundation/services/party"
)

// GetPartyAddressLinkById mengambil satu relasi alamat mitra berdasarkan UUID tautan.
func (h *PartyHandler) GetPartyAddressLinkById(ctx context.Context, req *efv1.GetPartyAddressLinkByIdRequest) (*efv1.GetPartyAddressLinkByIdResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	link, err := h.svc.GetAddressLinkByID(ctx, req.GetLinkId())
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.GetPartyAddressLinkByIdResponse{Link: toProtoPartyAddress(link)}, nil
}

// ListPartyAddressLinksByParty mengambil daftar relasi alamat milik satu mitra secara berhalaman.
func (h *PartyHandler) ListPartyAddressLinksByParty(ctx context.Context, req *efv1.ListPartyAddressLinksByPartyRequest) (*efv1.ListPartyAddressLinksByPartyResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	page, err := h.svc.ListAddressLinksByParty(ctx, req.GetPartyId(), fromProtoPageRequest(req.GetPage()))
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.ListPartyAddressLinksByPartyResponse{
		Items: toProtoPartyAddresses(page.Items),
		Page:  toProtoPageInfo(page),
	}, nil
}

// ListPartyAddressLinksByPartyAndUsage mengambil daftar relasi alamat mitra berdasarkan peruntukan alamat.
func (h *PartyHandler) ListPartyAddressLinksByPartyAndUsage(ctx context.Context, req *efv1.ListPartyAddressLinksByPartyAndUsageRequest) (*efv1.ListPartyAddressLinksByPartyAndUsageResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	usage, err := mapProtoEnum("usage", req.GetUsage(), efv1.AddressUsageKind_ADDRESS_USAGE_KIND_UNSPECIFIED, protoToDomainAddressUsageKind)
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	page, err := h.svc.ListAddressLinksByPartyAndUsage(ctx, req.GetPartyId(), usage, fromProtoPageRequest(req.GetPage()))
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.ListPartyAddressLinksByPartyAndUsageResponse{
		Items: toProtoPartyAddresses(page.Items),
		Page:  toProtoPageInfo(page),
	}, nil
}

// GetPrimaryPartyAddressLink mengambil alamat utama mitra untuk satu jenis peruntukan.
func (h *PartyHandler) GetPrimaryPartyAddressLink(ctx context.Context, req *efv1.GetPrimaryPartyAddressLinkRequest) (*efv1.GetPrimaryPartyAddressLinkResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	usage, err := mapProtoEnum("usage", req.GetUsage(), efv1.AddressUsageKind_ADDRESS_USAGE_KIND_UNSPECIFIED, protoToDomainAddressUsageKind)
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	link, err := h.svc.GetPrimaryAddressLink(ctx, req.GetPartyId(), usage)
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.GetPrimaryPartyAddressLinkResponse{Link: toProtoPartyAddress(link)}, nil
}

// ListPartyAddressLinksByAddressId mengambil daftar relasi mitra yang memakai satu alamat master.
func (h *PartyHandler) ListPartyAddressLinksByAddressId(ctx context.Context, req *efv1.ListPartyAddressLinksByAddressIdRequest) (*efv1.ListPartyAddressLinksByAddressIdResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	page, err := h.svc.ListAddressLinksByAddressID(ctx, req.GetAddressId(), fromProtoPageRequest(req.GetPage()))
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.ListPartyAddressLinksByAddressIdResponse{
		Items: toProtoPartyAddresses(page.Items),
		Page:  toProtoPageInfo(page),
	}, nil
}

// ListPartyAddressLinks mencari relasi alamat mitra berdasarkan filter dan parameter paginasi.
func (h *PartyHandler) ListPartyAddressLinks(ctx context.Context, req *efv1.ListPartyAddressLinksRequest) (*efv1.ListPartyAddressLinksResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	usageFilter, err := mapOptionalFilterEnum("usage", req.Usage, efv1.AddressUsageKind_ADDRESS_USAGE_KIND_UNSPECIFIED, protoToDomainAddressUsageKind)
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	page, err := h.svc.ListAddressLinks(ctx, domainrepo.PartyAddressFilter{
		PartyID:     req.GetPartyId(),
		AddressID:   req.GetAddressId(),
		Usage:       usageFilter,
		PrimaryOnly: clonePtr(req.PrimaryOnly),
		Page:        fromProtoPageRequest(req.GetPage()),
	})
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.ListPartyAddressLinksResponse{
		Items: toProtoPartyAddresses(page.Items),
		Page:  toProtoPageInfo(page),
	}, nil
}

// GetAllPartyAddressLinks mengambil seluruh relasi alamat mitra secara berhalaman.
func (h *PartyHandler) GetAllPartyAddressLinks(ctx context.Context, req *efv1.GetAllPartyAddressLinksRequest) (*efv1.GetAllPartyAddressLinksResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	page, err := h.svc.GetAllAddressLinks(ctx, fromProtoPageRequest(req.GetPage()))
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.GetAllPartyAddressLinksResponse{
		Items: toProtoPartyAddresses(page.Items),
		Page:  toProtoPageInfo(page),
	}, nil
}

// LinkPartyAddress menautkan alamat master ke mitra dengan opsi penggantian alamat utama per peruntukan.
func (h *PartyHandler) LinkPartyAddress(ctx context.Context, req *efv1.LinkPartyAddressRequest) (*efv1.LinkPartyAddressResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	usage, err := mapProtoEnum("usage", req.GetUsage(), efv1.AddressUsageKind_ADDRESS_USAGE_KIND_UNSPECIFIED, protoToDomainAddressUsageKind)
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	link, err := h.svc.LinkAddress(ctx, partysvc.LinkAddressInput{
		PartyID:        req.GetPartyId(),
		AddressID:      req.GetAddressId(),
		Usage:          usage,
		IsPrimary:      req.GetIsPrimary(),
		ReplacePrimary: req.GetReplacePrimary(),
	})
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.LinkPartyAddressResponse{Link: toProtoPartyAddress(link)}, nil
}

// UpdatePartyAddressLink memperbarui peruntukan dan penanda utama pada relasi alamat mitra.
func (h *PartyHandler) UpdatePartyAddressLink(ctx context.Context, req *efv1.UpdatePartyAddressLinkRequest) (*efv1.UpdatePartyAddressLinkResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	usage, err := mapProtoEnum("usage", req.GetUsage(), efv1.AddressUsageKind_ADDRESS_USAGE_KIND_UNSPECIFIED, protoToDomainAddressUsageKind)
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	updated, err := h.svc.UpdateAddressLink(ctx, partysvc.UpdateAddressLinkInput{
		ID:             req.GetId(),
		Usage:          usage,
		IsPrimary:      req.GetIsPrimary(),
		ReplacePrimary: req.GetReplacePrimary(),
	})
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.UpdatePartyAddressLinkResponse{Link: toProtoPartyAddress(updated)}, nil
}

// SetPrimaryPartyAddressLink menetapkan tautan alamat sebagai alamat utama untuk peruntukannya.
func (h *PartyHandler) SetPrimaryPartyAddressLink(ctx context.Context, req *efv1.SetPrimaryPartyAddressLinkRequest) (*efv1.SetPrimaryPartyAddressLinkResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	updated, err := h.svc.SetPrimaryAddressLink(ctx, req.GetLinkId())
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.SetPrimaryPartyAddressLinkResponse{Link: toProtoPartyAddress(updated)}, nil
}

// UnlinkPartyAddress melepaskan tautan alamat dari mitra eksternal.
func (h *PartyHandler) UnlinkPartyAddress(ctx context.Context, req *efv1.UnlinkPartyAddressRequest) (*efv1.UnlinkPartyAddressResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	if err := h.svc.UnlinkAddress(ctx, req.GetLinkId()); err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.UnlinkPartyAddressResponse{Unlinked: true}, nil
}
