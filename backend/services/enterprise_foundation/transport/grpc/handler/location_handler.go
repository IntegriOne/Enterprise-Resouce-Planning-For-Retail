package handler

import (
	"context"
	"encoding/json"
	"strings"

	efv1 "github.com/integrione/erp-retail/backend/services/enterprise_foundation/contract/gen/enterprise_foundation/v1"
	eferrors "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/repository"
	locationsvc "github.com/integrione/erp-retail/backend/services/enterprise_foundation/services/location"
)

// LocationHandler menghubungkan kontrak gRPC LocationService ke layanan lokasi fisik domain.
type LocationHandler struct {
	efv1.UnimplementedLocationServiceServer
	svc       *locationsvc.Service
	errMapper *ErrorMapper
}

// NewLocationHandler membentuk LocationHandler dengan validasi fail-closed atas dependensi service.
func NewLocationHandler(svc *locationsvc.Service, errMapper *ErrorMapper) (*LocationHandler, error) {
	if svc == nil {
		return nil, eferrors.NewValidation("location_service", eferrors.Code("enterprise_foundation.handler.nil_location_service"), "location service is required")
	}
	return &LocationHandler{
		svc:       svc,
		errMapper: ensureErrorMapper(errMapper),
	}, nil
}

// parseOperationalAttributesJSON mengonversi atribut operasional JSON teks menjadi json.RawMessage atau nil bila kosong.
func parseOperationalAttributesJSON(raw string) json.RawMessage {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil
	}
	return json.RawMessage(trimmed)
}

// GetLocationById mengambil satu lokasi fisik berdasarkan UUID.
func (h *LocationHandler) GetLocationById(ctx context.Context, req *efv1.GetLocationByIdRequest) (*efv1.GetLocationByIdResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	loc, err := h.svc.GetByID(ctx, req.GetId())
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.GetLocationByIdResponse{Location: toProtoLocation(loc)}, nil
}

// GetLocationsByIds mengambil daftar lokasi fisik berdasarkan kumpulan UUID.
func (h *LocationHandler) GetLocationsByIds(ctx context.Context, req *efv1.GetLocationsByIdsRequest) (*efv1.GetLocationsByIdsResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	items, err := h.svc.GetByIDs(ctx, req.GetIds())
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.GetLocationsByIdsResponse{Items: toProtoLocations(items)}, nil
}

// GetLocationByCode mengambil satu lokasi fisik berdasarkan kode unik lokasi.
func (h *LocationHandler) GetLocationByCode(ctx context.Context, req *efv1.GetLocationByCodeRequest) (*efv1.GetLocationByCodeResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	loc, err := h.svc.GetByCode(ctx, req.GetLocationCode())
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.GetLocationByCodeResponse{Location: toProtoLocation(loc)}, nil
}

// ListLocationsByLegalEntityAndKind mengambil daftar lokasi pada satu badan hukum berdasarkan jenisnya.
func (h *LocationHandler) ListLocationsByLegalEntityAndKind(ctx context.Context, req *efv1.ListLocationsByLegalEntityAndKindRequest) (*efv1.ListLocationsByLegalEntityAndKindResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	kind, err := mapProtoEnum("kind", req.GetKind(), efv1.LocationKind_LOCATION_KIND_UNSPECIFIED, protoToDomainLocationKind)
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	page, err := h.svc.ListByLegalEntityAndKind(ctx, req.GetLegalEntityId(), kind, req.GetActiveOnly(), fromProtoPageRequest(req.GetPage()))
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.ListLocationsByLegalEntityAndKindResponse{
		Items: toProtoLocations(page.Items),
		Page:  toProtoPageInfo(page),
	}, nil
}

// ListLocationsByAddressId mengambil daftar lokasi fisik yang terhubung ke satu master alamat.
func (h *LocationHandler) ListLocationsByAddressId(ctx context.Context, req *efv1.ListLocationsByAddressIdRequest) (*efv1.ListLocationsByAddressIdResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	page, err := h.svc.ListByAddressID(ctx, req.GetAddressId(), fromProtoPageRequest(req.GetPage()))
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.ListLocationsByAddressIdResponse{
		Items: toProtoLocations(page.Items),
		Page:  toProtoPageInfo(page),
	}, nil
}

// ListLocationChildren mengambil daftar sub-lokasi anak langsung dari satu lokasi induk.
func (h *LocationHandler) ListLocationChildren(ctx context.Context, req *efv1.ListLocationChildrenRequest) (*efv1.ListLocationChildrenResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	page, err := h.svc.ListChildren(ctx, req.GetParentLocationId(), fromProtoPageRequest(req.GetPage()))
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.ListLocationChildrenResponse{
		Items: toProtoLocations(page.Items),
		Page:  toProtoPageInfo(page),
	}, nil
}

// ListLocations mencari daftar lokasi fisik berdasarkan filter dan paginasi.
func (h *LocationHandler) ListLocations(ctx context.Context, req *efv1.ListLocationsRequest) (*efv1.ListLocationsResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	kindFilter, err := mapOptionalFilterEnum("kind", req.Kind, efv1.LocationKind_LOCATION_KIND_UNSPECIFIED, protoToDomainLocationKind)
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	page, err := h.svc.List(ctx, domainrepo.LocationFilter{
		Query:            req.GetQuery(),
		LegalEntityID:    req.GetLegalEntityId(),
		ParentLocationID: clonePtr(req.ParentLocationId),
		AddressID:        clonePtr(req.AddressId),
		Kind:             kindFilter,
		ActiveOnly:       req.GetActiveOnly(),
		IncludeDeleted:   req.GetIncludeDeleted(),
		Page:             fromProtoPageRequest(req.GetPage()),
	})
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.ListLocationsResponse{
		Items: toProtoLocations(page.Items),
		Page:  toProtoPageInfo(page),
	}, nil
}

// GetAllLocations mengambil seluruh lokasi aktif secara berhalaman.
func (h *LocationHandler) GetAllLocations(ctx context.Context, req *efv1.GetAllLocationsRequest) (*efv1.GetAllLocationsResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	page, err := h.svc.GetAll(ctx, fromProtoPageRequest(req.GetPage()))
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.GetAllLocationsResponse{
		Items: toProtoLocations(page.Items),
		Page:  toProtoPageInfo(page),
	}, nil
}

// ValidateFacilityScope memvalidasi kelayakan lokasi sebagai target scope facility IAM secara fail-closed.
func (h *LocationHandler) ValidateFacilityScope(ctx context.Context, req *efv1.ValidateFacilityScopeRequest) (*efv1.ValidateFacilityScopeResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	decision, err := h.svc.ValidateFacilityScope(ctx, req.GetId())
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.ValidateFacilityScopeResponse{Decision: toProtoScopeDecision(decision)}, nil
}

// ValidateWarehouseScope memvalidasi kelayakan lokasi bertipe gudang untuk WMS secara fail-closed.
func (h *LocationHandler) ValidateWarehouseScope(ctx context.Context, req *efv1.ValidateWarehouseScopeRequest) (*efv1.ValidateWarehouseScopeResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	decision, err := h.svc.ValidateWarehouseScope(ctx, req.GetId())
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.ValidateWarehouseScopeResponse{Decision: toProtoScopeDecision(decision)}, nil
}

// CreateLocation membentuk lokasi fisik baru dengan validasi badan hukum, alamat, hierarki, dan atribut operasional JSON.
func (h *LocationHandler) CreateLocation(ctx context.Context, req *efv1.CreateLocationRequest) (*efv1.CreateLocationResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	kind, err := mapProtoEnum("kind", req.GetKind(), efv1.LocationKind_LOCATION_KIND_UNSPECIFIED, protoToDomainLocationKind)
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	created, err := h.svc.Create(ctx, locationsvc.CreateInput{
		LegalEntityID:         req.GetLegalEntityId(),
		ParentLocationID:      clonePtr(req.ParentLocationId),
		AddressID:             clonePtr(req.AddressId),
		LocationCode:          req.GetLocationCode(),
		Name:                  req.GetName(),
		Kind:                  kind,
		Phone:                 clonePtr(req.Phone),
		Timezone:              req.GetTimezone(),
		OperationalAttributes: parseOperationalAttributesJSON(req.GetOperationalAttributesJson()),
	})
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.CreateLocationResponse{Location: toProtoLocation(created)}, nil
}

// UpdateLocation memperbarui profil, relasi induk/alamat, jenis, dan atribut operasional JSON lokasi fisik.
func (h *LocationHandler) UpdateLocation(ctx context.Context, req *efv1.UpdateLocationRequest) (*efv1.UpdateLocationResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	kind, err := mapProtoEnum("kind", req.GetKind(), efv1.LocationKind_LOCATION_KIND_UNSPECIFIED, protoToDomainLocationKind)
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	updated, err := h.svc.Update(ctx, locationsvc.UpdateInput{
		ID:                    req.GetId(),
		ParentLocationID:      clonePtr(req.ParentLocationId),
		AddressID:             clonePtr(req.AddressId),
		Name:                  req.GetName(),
		Kind:                  kind,
		Phone:                 clonePtr(req.Phone),
		Timezone:              req.GetTimezone(),
		OperationalAttributes: parseOperationalAttributesJSON(req.GetOperationalAttributesJson()),
	})
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.UpdateLocationResponse{Location: toProtoLocation(updated)}, nil
}

// ActivateLocation mengaktifkan kembali lokasi fisik agar dapat beroperasi.
func (h *LocationHandler) ActivateLocation(ctx context.Context, req *efv1.ActivateLocationRequest) (*efv1.ActivateLocationResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	updated, err := h.svc.Activate(ctx, req.GetId())
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.ActivateLocationResponse{Location: toProtoLocation(updated)}, nil
}

// DeactivateLocation menonaktifkan sementara lokasi fisik dari operasional dan scope baru.
func (h *LocationHandler) DeactivateLocation(ctx context.Context, req *efv1.DeactivateLocationRequest) (*efv1.DeactivateLocationResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	updated, err := h.svc.Deactivate(ctx, req.GetId())
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.DeactivateLocationResponse{Location: toProtoLocation(updated)}, nil
}

// DeleteLocation melakukan soft delete lokasi fisik bila sudah tidak memiliki sub-lokasi anak aktif.
func (h *LocationHandler) DeleteLocation(ctx context.Context, req *efv1.DeleteLocationRequest) (*efv1.DeleteLocationResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	if err := h.svc.Delete(ctx, req.GetId()); err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.DeleteLocationResponse{Deleted: true}, nil
}
