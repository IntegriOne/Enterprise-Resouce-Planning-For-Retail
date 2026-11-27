package handler

import (
	"context"

	efv1 "github.com/integrione/erp-retail/backend/services/enterprise_foundation/contract/gen/enterprise_foundation/v1"
	eferrors "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/repository"
	addresssvc "github.com/integrione/erp-retail/backend/services/enterprise_foundation/services/address"
)

// AddressHandler menghubungkan kontrak gRPC AddressService ke layanan master alamat domain.
type AddressHandler struct {
	efv1.UnimplementedAddressServiceServer
	svc       *addresssvc.Service
	errMapper *ErrorMapper
}

// NewAddressHandler membentuk AddressHandler dengan pemeriksaan fail-closed atas dependensi service.
func NewAddressHandler(svc *addresssvc.Service, errMapper *ErrorMapper) (*AddressHandler, error) {
	if svc == nil {
		return nil, eferrors.NewValidation("address_service", eferrors.Code("enterprise_foundation.handler.nil_address_service"), "address service is required")
	}
	return &AddressHandler{
		svc:       svc,
		errMapper: ensureErrorMapper(errMapper),
	}, nil
}

// GetAddressById mengambil satu alamat master berdasarkan UUID.
func (h *AddressHandler) GetAddressById(ctx context.Context, req *efv1.GetAddressByIdRequest) (*efv1.GetAddressByIdResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	addr, err := h.svc.GetByID(ctx, req.GetId())
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.GetAddressByIdResponse{Address: toProtoAddress(addr)}, nil
}

// GetAddressesByIds mengambil daftar alamat master berdasarkan kumpulan UUID.
func (h *AddressHandler) GetAddressesByIds(ctx context.Context, req *efv1.GetAddressesByIdsRequest) (*efv1.GetAddressesByIdsResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	items, err := h.svc.GetByIDs(ctx, req.GetIds())
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.GetAddressesByIdsResponse{Items: toProtoAddresses(items)}, nil
}

// ListAddresses mencari master alamat berdasarkan filter pencarian dan parameter paginasi.
func (h *AddressHandler) ListAddresses(ctx context.Context, req *efv1.ListAddressesRequest) (*efv1.ListAddressesResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	page, err := h.svc.List(ctx, domainrepo.AddressFilter{
		Query:       req.GetQuery(),
		CountryCode: req.GetCountryCode(),
		City:        req.GetCity(),
		Province:    req.GetProvince(),
		PostalCode:  req.GetPostalCode(),
		Page:        fromProtoPageRequest(req.GetPage()),
	})
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.ListAddressesResponse{
		Items: toProtoAddresses(page.Items),
		Page:  toProtoPageInfo(page),
	}, nil
}

// GetAllAddresses mengambil seluruh master alamat secara berhalaman.
func (h *AddressHandler) GetAllAddresses(ctx context.Context, req *efv1.GetAllAddressesRequest) (*efv1.GetAllAddressesResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	page, err := h.svc.GetAll(ctx, fromProtoPageRequest(req.GetPage()))
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.GetAllAddressesResponse{
		Items: toProtoAddresses(page.Items),
		Page:  toProtoPageInfo(page),
	}, nil
}

// IsAddressReferenced memeriksa apakah alamat masih dipakai oleh badan hukum, lokasi, atau mitra eksternal.
func (h *AddressHandler) IsAddressReferenced(ctx context.Context, req *efv1.IsAddressReferencedRequest) (*efv1.IsAddressReferencedResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	referenced, err := h.svc.IsReferenced(ctx, req.GetId())
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.IsAddressReferencedResponse{Referenced: referenced}, nil
}

// CreateAddress membentuk master alamat baru setelah validasi duplikasi terkanonisasi.
func (h *AddressHandler) CreateAddress(ctx context.Context, req *efv1.CreateAddressRequest) (*efv1.CreateAddressResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	created, err := h.svc.Create(ctx, addresssvc.CreateInput{
		Label:            clonePtr(req.Label),
		AddressLine1:     req.GetAddressLine_1(),
		AddressLine2:     clonePtr(req.AddressLine_2),
		City:             req.GetCity(),
		Province:         clonePtr(req.Province),
		PostalCode:       clonePtr(req.PostalCode),
		CountryCode:      req.GetCountryCode(),
		FormattedAddress: clonePtr(req.FormattedAddress),
		Latitude:         clonePtr(req.Latitude),
		Longitude:        clonePtr(req.Longitude),
	})
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.CreateAddressResponse{Address: toProtoAddress(created)}, nil
}

// FindOrCreateAddress mengembalikan alamat identik yang sudah ada atau membentuk baris alamat baru secara atomik.
func (h *AddressHandler) FindOrCreateAddress(ctx context.Context, req *efv1.FindOrCreateAddressRequest) (*efv1.FindOrCreateAddressResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	addr, created, err := h.svc.FindOrCreate(ctx, addresssvc.CreateInput{
		Label:            clonePtr(req.Label),
		AddressLine1:     req.GetAddressLine_1(),
		AddressLine2:     clonePtr(req.AddressLine_2),
		City:             req.GetCity(),
		Province:         clonePtr(req.Province),
		PostalCode:       clonePtr(req.PostalCode),
		CountryCode:      req.GetCountryCode(),
		FormattedAddress: clonePtr(req.FormattedAddress),
		Latitude:         clonePtr(req.Latitude),
		Longitude:        clonePtr(req.Longitude),
	})
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.FindOrCreateAddressResponse{
		Address: toProtoAddress(addr),
		Created: created,
	}, nil
}

// UpdateAddress memperbarui data master alamat yang sudah terdaftar.
func (h *AddressHandler) UpdateAddress(ctx context.Context, req *efv1.UpdateAddressRequest) (*efv1.UpdateAddressResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	updated, err := h.svc.Update(ctx, addresssvc.UpdateInput{
		ID:               req.GetId(),
		Label:            clonePtr(req.Label),
		AddressLine1:     req.GetAddressLine_1(),
		AddressLine2:     clonePtr(req.AddressLine_2),
		City:             req.GetCity(),
		Province:         clonePtr(req.Province),
		PostalCode:       clonePtr(req.PostalCode),
		CountryCode:      req.GetCountryCode(),
		FormattedAddress: clonePtr(req.FormattedAddress),
		Latitude:         clonePtr(req.Latitude),
		Longitude:        clonePtr(req.Longitude),
	})
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.UpdateAddressResponse{Address: toProtoAddress(updated)}, nil
}

// DeleteAddress menghapus master alamat apabila sudah tidak lagi direferensikan oleh entitas lain.
func (h *AddressHandler) DeleteAddress(ctx context.Context, req *efv1.DeleteAddressRequest) (*efv1.DeleteAddressResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	if err := h.svc.Delete(ctx, req.GetId()); err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.DeleteAddressResponse{Deleted: true}, nil
}
