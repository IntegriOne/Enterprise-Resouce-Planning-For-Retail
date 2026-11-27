package enterprise_foundation

import (
	"context"
	"strings"

	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	"github.com/integrione/erp-retail/backend/gateway/internal/source_connections"
	efv1 "github.com/integrione/erp-retail/backend/services/enterprise_foundation/contract/gen/enterprise_foundation/v1"
	"github.com/integrione/erp-retail/backend/shared/authz"
)

func buildLocationConn(
	after, before *string,
	first, last *int,
	fetch func(limit, offset int32) ([]*efv1.Location, *efv1.PageInfo, error),
) (*model.LocationConnection, error) {
	return buildRelayConnection(
		after, before, first, last,
		fetch,
		toModelLocation,
		func(node *model.Location, cursor string) *model.LocationEdge {
			return &model.LocationEdge{Node: node, Cursor: cursor}
		},
		func(edges []*model.LocationEdge, nodes []*model.Location, pi *model.PageInfo, total int) *model.LocationConnection {
			return &model.LocationConnection{Edges: edges, Nodes: nodes, PageInfo: pi, TotalCount: total}
		},
	)
}

func facilityScope(locationID string) authz.ScopeGrant {
	if strings.TrimSpace(locationID) == "" {
		return authz.ScopeGrant{Type: authz.ScopeGlobal}
	}
	return authz.ScopeGrant{Type: scopeTypeFacility, ID: locationID}
}

// Location mengambil satu lokasi fisik berdasarkan ID/GID dengan pemeriksaan scope facility.
func (a *Adapter) Location(ctx context.Context, id string) (*model.Location, error) {
	cleanID, err := ParseTypedID(id, ResourceLocation)
	if err != nil {
		if _, authErr := a.requirePerm(ctx, PermLocationsRead, false); authErr != nil {
			return nil, authErr
		}
		return nil, err
	}
	sess, err := a.requireAnyScope(ctx, PermLocationsRead, false, facilityScope(cleanID))
	if err != nil {
		return nil, err
	}
	resp, err := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.GetLocationByIdResponse, error) {
		return c.Locations.GetLocationById(rpcCtx, &efv1.GetLocationByIdRequest{Id: cleanID})
	})
	if err != nil {
		return nil, err
	}
	return toModelLocation(resp.GetLocation()), nil
}

// LocationsByIds mengambil banyak lokasi fisik sekaligus secara batch dengan verifikasi scope per ID.
func (a *Adapter) LocationsByIds(ctx context.Context, ids []string) ([]*model.Location, error) {
	sess, err := a.authorizer.RequirePermission(ctx, PermLocationsRead, authz.ScopeGrant{Type: authz.ScopeGlobal}, false)
	hasGlobal := err == nil
	if err != nil {
		gwErr, ok := middleware.AsGatewayError(err)
		if !ok || gwErr.Code != middleware.CodeForbidden {
			return nil, err
		}
	}

	cleanIDs, parseErr := ParseTypedIDList(ids, ResourceLocation)
	if parseErr != nil {
		if !hasGlobal {
			return nil, err
		}
		return nil, parseErr
	}
	if len(cleanIDs) == 0 {
		if !hasGlobal {
			return nil, err
		}
		return []*model.Location{}, nil
	}

	for _, id := range cleanIDs {
		s, scopeErr := a.requireAnyScope(ctx, PermLocationsRead, false, facilityScope(id))
		if scopeErr != nil {
			return nil, scopeErr
		}
		sess = s
	}

	resp, err := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.GetLocationsByIdsResponse, error) {
		return c.Locations.GetLocationsByIds(rpcCtx, &efv1.GetLocationsByIdsRequest{Ids: cleanIDs})
	})
	if err != nil {
		return nil, err
	}
	return toModelLocations(resp.GetItems()), nil
}

// LocationByCode mengambil satu lokasi fisik berdasarkan kode lokasi.
func (a *Adapter) LocationByCode(ctx context.Context, code string) (*model.Location, error) {
	sess, err := a.requirePerm(ctx, PermLocationsRead, false)
	if err != nil {
		return nil, err
	}
	trimmed := strings.TrimSpace(code)
	if trimmed == "" {
		return nil, middleware.ErrInvalidArgument("location_code is required", nil)
	}
	resp, err := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.GetLocationByCodeResponse, error) {
		return c.Locations.GetLocationByCode(rpcCtx, &efv1.GetLocationByCodeRequest{LocationCode: trimmed})
	})
	if err != nil {
		return nil, err
	}
	return toModelLocation(resp.GetLocation()), nil
}

// LocationsByLegalEntityAndKind mengambil daftar lokasi milik suatu badan hukum dan jenis tertentu.
func (a *Adapter) LocationsByLegalEntityAndKind(
	ctx context.Context,
	legalEntityID string,
	kind model.LocationKind,
	activeOnly *bool,
	after, before *string,
	first, last *int,
) (*model.LocationConnection, error) {
	cleanEntityID, parseErr := ParseTypedID(legalEntityID, ResourceLegalEntity)
	if parseErr != nil {
		if _, authErr := a.requirePerm(ctx, PermLocationsRead, false); authErr != nil {
			return nil, authErr
		}
		return nil, parseErr
	}
	sess, err := a.requireAnyScope(ctx, PermLocationsRead, false, legalEntityScope(cleanEntityID))
	if err != nil {
		return nil, err
	}
	protoKind := mapEnumValue(kind, locationKindToProto, efv1.LocationKind_LOCATION_KIND_UNSPECIFIED)
	return buildLocationConn(after, before, first, last, func(limit, offset int32) ([]*efv1.Location, *efv1.PageInfo, error) {
		resp, rpcErr := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.ListLocationsByLegalEntityAndKindResponse, error) {
			return c.Locations.ListLocationsByLegalEntityAndKind(rpcCtx, &efv1.ListLocationsByLegalEntityAndKindRequest{
				LegalEntityId: cleanEntityID,
				Kind:          protoKind,
				ActiveOnly:    derefBool(activeOnly, false),
				Page:          &efv1.PageRequest{Limit: limit, Offset: offset},
			})
		})
		if rpcErr != nil {
			return nil, nil, rpcErr
		}
		return resp.GetItems(), resp.GetPage(), nil
	})
}

// LocationsByAddressId mengambil daftar lokasi yang merujuk pada suatu master alamat.
func (a *Adapter) LocationsByAddressId(ctx context.Context, addressID string, after, before *string, first, last *int) (*model.LocationConnection, error) {
	sess, err := a.requirePerm(ctx, PermLocationsRead, false)
	if err != nil {
		return nil, err
	}
	cleanAddrID, err := ParseTypedID(addressID, ResourceAddress)
	if err != nil {
		return nil, err
	}
	return buildLocationConn(after, before, first, last, func(limit, offset int32) ([]*efv1.Location, *efv1.PageInfo, error) {
		resp, rpcErr := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.ListLocationsByAddressIdResponse, error) {
			return c.Locations.ListLocationsByAddressId(rpcCtx, &efv1.ListLocationsByAddressIdRequest{
				AddressId: cleanAddrID,
				Page:      &efv1.PageRequest{Limit: limit, Offset: offset},
			})
		})
		if rpcErr != nil {
			return nil, nil, rpcErr
		}
		return resp.GetItems(), resp.GetPage(), nil
	})
}

// LocationChildren mengambil anak lokasi langsung dari lokasi induk.
func (a *Adapter) LocationChildren(ctx context.Context, parentLocationID string, after, before *string, first, last *int) (*model.LocationConnection, error) {
	cleanParentID, parseErr := ParseTypedID(parentLocationID, ResourceLocation)
	if parseErr != nil {
		if _, authErr := a.requirePerm(ctx, PermLocationsRead, false); authErr != nil {
			return nil, authErr
		}
		return nil, parseErr
	}
	sess, err := a.requireAnyScope(ctx, PermLocationsRead, false, facilityScope(cleanParentID))
	if err != nil {
		return nil, err
	}
	return buildLocationConn(after, before, first, last, func(limit, offset int32) ([]*efv1.Location, *efv1.PageInfo, error) {
		resp, rpcErr := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.ListLocationChildrenResponse, error) {
			return c.Locations.ListLocationChildren(rpcCtx, &efv1.ListLocationChildrenRequest{
				ParentLocationId: cleanParentID,
				Page:             &efv1.PageRequest{Limit: limit, Offset: offset},
			})
		})
		if rpcErr != nil {
			return nil, nil, rpcErr
		}
		return resp.GetItems(), resp.GetPage(), nil
	})
}

// Locations mengambil daftar lokasi terpaginasi dengan filter badan hukum, induk, alamat, jenis, dan status.
func (a *Adapter) Locations(
	ctx context.Context,
	after, before *string,
	first, last *int,
	query, legalEntityID, parentLocationID, addressID *string,
	kind *model.LocationKind,
	activeOnly, includeDeleted *bool,
) (*model.LocationConnection, error) {
	cleanEntityID, entityParseErr := ParseOptionalTypedID(legalEntityID, ResourceLegalEntity)
	var sess *middleware.ResolvedSession
	var authErr error
	if cleanEntityID != nil && entityParseErr == nil {
		sess, authErr = a.requireAnyScope(ctx, PermLocationsRead, false, legalEntityScope(*cleanEntityID))
	} else {
		sess, authErr = a.requirePerm(ctx, PermLocationsRead, false)
	}
	if authErr != nil {
		return nil, authErr
	}
	if entityParseErr != nil {
		return nil, entityParseErr
	}
	cleanParentID, err := ParseOptionalTypedID(parentLocationID, ResourceLocation)
	if err != nil {
		return nil, err
	}
	cleanAddressID, err := ParseOptionalTypedID(addressID, ResourceAddress)
	if err != nil {
		return nil, err
	}

	return buildLocationConn(after, before, first, last, func(limit, offset int32) ([]*efv1.Location, *efv1.PageInfo, error) {
		resp, rpcErr := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.ListLocationsResponse, error) {
			return c.Locations.ListLocations(rpcCtx, &efv1.ListLocationsRequest{
				Query:            strings.TrimSpace(derefStr(query)),
				LegalEntityId:    derefStr(cleanEntityID),
				ParentLocationId: cleanParentID,
				AddressId:        cleanAddressID,
				Kind:             mapOptionalEnumPtr(kind, locationKindToProto),
				ActiveOnly:       derefBool(activeOnly, false),
				IncludeDeleted:   derefBool(includeDeleted, false),
				Page:             &efv1.PageRequest{Limit: limit, Offset: offset},
			})
		})
		if rpcErr != nil {
			return nil, nil, rpcErr
		}
		return resp.GetItems(), resp.GetPage(), nil
	})
}

// AllLocations mengambil seluruh lokasi fisik secara terpaginasi.
func (a *Adapter) AllLocations(ctx context.Context, after, before *string, first, last *int) (*model.LocationConnection, error) {
	sess, err := a.requirePerm(ctx, PermLocationsRead, false)
	if err != nil {
		return nil, err
	}
	return buildLocationConn(after, before, first, last, func(limit, offset int32) ([]*efv1.Location, *efv1.PageInfo, error) {
		resp, rpcErr := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.GetAllLocationsResponse, error) {
			return c.Locations.GetAllLocations(rpcCtx, &efv1.GetAllLocationsRequest{
				Page: &efv1.PageRequest{Limit: limit, Offset: offset},
			})
		})
		if rpcErr != nil {
			return nil, nil, rpcErr
		}
		return resp.GetItems(), resp.GetPage(), nil
	})
}

// ValidateFacilityScope memvalidasi kelayakan lokasi fisik sebagai target scope facility IAM.
func (a *Adapter) ValidateFacilityScope(ctx context.Context, id string) (*model.ScopeTargetDecision, error) {
	cleanID, parseErr := ParseTypedID(id, ResourceLocation)
	if parseErr != nil {
		if _, authErr := a.requirePerm(ctx, PermLocationsRead, false); authErr != nil {
			return nil, authErr
		}
		return nil, parseErr
	}
	sess, err := a.requireAnyScope(ctx, PermLocationsRead, false, facilityScope(cleanID))
	if err != nil {
		return nil, err
	}
	resp, err := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.ValidateFacilityScopeResponse, error) {
		return c.Locations.ValidateFacilityScope(rpcCtx, &efv1.ValidateFacilityScopeRequest{Id: cleanID})
	})
	if err != nil {
		return nil, err
	}
	return toModelScopeTargetDecision(resp.GetDecision()), nil
}

// ValidateWarehouseScope memvalidasi kelayakan lokasi gudang sebagai target scope warehouse.
func (a *Adapter) ValidateWarehouseScope(ctx context.Context, id string) (*model.ScopeTargetDecision, error) {
	cleanID, parseErr := ParseTypedID(id, ResourceLocation)
	if parseErr != nil {
		if _, authErr := a.requirePerm(ctx, PermLocationsRead, false); authErr != nil {
			return nil, authErr
		}
		return nil, parseErr
	}
	sess, err := a.requireAnyScope(ctx, PermLocationsRead, false, facilityScope(cleanID))
	if err != nil {
		return nil, err
	}
	resp, err := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.ValidateWarehouseScopeResponse, error) {
		return c.Locations.ValidateWarehouseScope(rpcCtx, &efv1.ValidateWarehouseScopeRequest{Id: cleanID})
	})
	if err != nil {
		return nil, err
	}
	return toModelScopeTargetDecision(resp.GetDecision()), nil
}

// CreateLocation membuat lokasi fisik operasional baru.
func (a *Adapter) CreateLocation(ctx context.Context, input model.CreateLocationInput) *model.CreateLocationPayload {
	cleanEntityID, idErr := ParseTypedID(input.LegalEntityID, ResourceLegalEntity)
	if idErr != nil {
		if _, authErr := a.requirePerm(ctx, PermLocationsWrite, false); authErr != nil {
			return &model.CreateLocationPayload{Errors: ToMutationErrors(authErr), WasSuccessful: false}
		}
		return &model.CreateLocationPayload{Errors: ToMutationErrors(idErr), WasSuccessful: false}
	}
	sess, err := a.requireAnyScope(ctx, PermLocationsWrite, false, legalEntityScope(cleanEntityID))
	if err != nil {
		return &model.CreateLocationPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	parentID, err := ParseOptionalTypedID(input.ParentLocationID, ResourceLocation)
	if err != nil {
		return &model.CreateLocationPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	addrID, err := ParseOptionalTypedID(input.AddressID, ResourceAddress)
	if err != nil {
		return &model.CreateLocationPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}

	resp, err := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.CreateLocationResponse, error) {
		return c.Locations.CreateLocation(rpcCtx, &efv1.CreateLocationRequest{
			LegalEntityId:             cleanEntityID,
			ParentLocationId:          parentID,
			AddressId:                 addrID,
			LocationCode:              strings.TrimSpace(input.LocationCode),
			Name:                      strings.TrimSpace(input.Name),
			Kind:                      mapEnumValue(input.Kind, locationKindToProto, efv1.LocationKind_LOCATION_KIND_UNSPECIFIED),
			Phone:                     input.Phone,
			Timezone:                  strings.TrimSpace(input.Timezone),
			OperationalAttributesJson: normalizeStringDefault(input.OperationalAttributesJSON, "{}"),
		})
	})
	if err != nil {
		return &model.CreateLocationPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	return &model.CreateLocationPayload{
		Location:      toModelLocation(resp.GetLocation()),
		Errors:        NoMutationErrors(),
		WasSuccessful: true,
	}
}

// UpdateLocation memperbarui data lokasi fisik operasional.
func (a *Adapter) UpdateLocation(ctx context.Context, input model.UpdateLocationInput) *model.UpdateLocationPayload {
	cleanID, idErr := ParseTypedID(input.ID, ResourceLocation)
	if idErr != nil {
		if _, authErr := a.requirePerm(ctx, PermLocationsWrite, false); authErr != nil {
			return &model.UpdateLocationPayload{Errors: ToMutationErrors(authErr), WasSuccessful: false}
		}
		return &model.UpdateLocationPayload{Errors: ToMutationErrors(idErr), WasSuccessful: false}
	}
	sess, err := a.requireAnyScope(ctx, PermLocationsWrite, false, facilityScope(cleanID))
	if err != nil {
		return &model.UpdateLocationPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	parentID, err := ParseOptionalTypedID(input.ParentLocationID, ResourceLocation)
	if err != nil {
		return &model.UpdateLocationPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	addrID, err := ParseOptionalTypedID(input.AddressID, ResourceAddress)
	if err != nil {
		return &model.UpdateLocationPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}

	resp, err := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.UpdateLocationResponse, error) {
		return c.Locations.UpdateLocation(rpcCtx, &efv1.UpdateLocationRequest{
			Id:                        cleanID,
			ParentLocationId:          parentID,
			AddressId:                 addrID,
			Name:                      strings.TrimSpace(input.Name),
			Kind:                      mapEnumValue(input.Kind, locationKindToProto, efv1.LocationKind_LOCATION_KIND_UNSPECIFIED),
			Phone:                     input.Phone,
			Timezone:                  strings.TrimSpace(input.Timezone),
			OperationalAttributesJson: normalizeStringDefault(input.OperationalAttributesJSON, "{}"),
		})
	})
	if err != nil {
		return &model.UpdateLocationPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	return &model.UpdateLocationPayload{
		Location:      toModelLocation(resp.GetLocation()),
		Errors:        NoMutationErrors(),
		WasSuccessful: true,
	}
}

// ActivateLocation mengaktifkan kembali lokasi fisik.
func (a *Adapter) ActivateLocation(ctx context.Context, id string) *model.ActivateLocationPayload {
	cleanID, idErr := ParseTypedID(id, ResourceLocation)
	if idErr != nil {
		if _, authErr := a.requirePerm(ctx, PermLocationsWrite, false); authErr != nil {
			return &model.ActivateLocationPayload{Errors: ToMutationErrors(authErr), WasSuccessful: false}
		}
		return &model.ActivateLocationPayload{Errors: ToMutationErrors(idErr), WasSuccessful: false}
	}
	sess, err := a.requireAnyScope(ctx, PermLocationsWrite, false, facilityScope(cleanID))
	if err != nil {
		return &model.ActivateLocationPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	resp, err := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.ActivateLocationResponse, error) {
		return c.Locations.ActivateLocation(rpcCtx, &efv1.ActivateLocationRequest{Id: cleanID})
	})
	if err != nil {
		return &model.ActivateLocationPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	return &model.ActivateLocationPayload{
		Location:      toModelLocation(resp.GetLocation()),
		Errors:        NoMutationErrors(),
		WasSuccessful: true,
	}
}

// DeactivateLocation menonaktifkan lokasi fisik tanpa menghapus historinya.
func (a *Adapter) DeactivateLocation(ctx context.Context, id string) *model.DeactivateLocationPayload {
	cleanID, idErr := ParseTypedID(id, ResourceLocation)
	if idErr != nil {
		if _, authErr := a.requirePerm(ctx, PermLocationsWrite, false); authErr != nil {
			return &model.DeactivateLocationPayload{Errors: ToMutationErrors(authErr), WasSuccessful: false}
		}
		return &model.DeactivateLocationPayload{Errors: ToMutationErrors(idErr), WasSuccessful: false}
	}
	sess, err := a.requireAnyScope(ctx, PermLocationsWrite, false, facilityScope(cleanID))
	if err != nil {
		return &model.DeactivateLocationPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	resp, err := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.DeactivateLocationResponse, error) {
		return c.Locations.DeactivateLocation(rpcCtx, &efv1.DeactivateLocationRequest{Id: cleanID})
	})
	if err != nil {
		return &model.DeactivateLocationPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	return &model.DeactivateLocationPayload{
		Location:      toModelLocation(resp.GetLocation()),
		Errors:        NoMutationErrors(),
		WasSuccessful: true,
	}
}

// DeleteLocation melakukan soft delete lokasi fisik dengan wajib re-autentikasi sensitif.
func (a *Adapter) DeleteLocation(ctx context.Context, id string) *model.DeleteLocationPayload {
	cleanID, idErr := ParseTypedID(id, ResourceLocation)
	if idErr != nil {
		if _, authErr := a.requirePerm(ctx, PermLocationsWrite, true); authErr != nil {
			return &model.DeleteLocationPayload{Errors: ToMutationErrors(authErr), WasSuccessful: false}
		}
		return &model.DeleteLocationPayload{Errors: ToMutationErrors(idErr), WasSuccessful: false}
	}
	sess, err := a.requireAnyScope(ctx, PermLocationsWrite, true, facilityScope(cleanID))
	if err != nil {
		return &model.DeleteLocationPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	resp, err := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.DeleteLocationResponse, error) {
		return c.Locations.DeleteLocation(rpcCtx, &efv1.DeleteLocationRequest{Id: cleanID})
	})
	if err != nil {
		return &model.DeleteLocationPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	deleted := resp.GetDeleted()
	if !deleted {
		return &model.DeleteLocationPayload{
			Errors:        ToMutationErrors(middleware.ErrConflict("location could not be deleted", nil)),
			WasSuccessful: false,
		}
	}
	return &model.DeleteLocationPayload{
		Deleted:       &deleted,
		Errors:        NoMutationErrors(),
		WasSuccessful: true,
	}
}

// ResolveLocationParent mengambil lokasi induk milik Location bila ada.
func (a *Adapter) ResolveLocationParent(ctx context.Context, obj *model.Location) (*model.Location, error) {
	if obj == nil || obj.ParentLocationID == nil || strings.TrimSpace(*obj.ParentLocationID) == "" {
		return nil, nil
	}
	return a.Location(ctx, *obj.ParentLocationID)
}

// ResolveLocationAddress mengambil alamat fisik milik Location bila ada.
func (a *Adapter) ResolveLocationAddress(ctx context.Context, obj *model.Location) (*model.Address, error) {
	if obj == nil || obj.AddressID == nil || strings.TrimSpace(*obj.AddressID) == "" {
		return nil, nil
	}
	return a.Address(ctx, *obj.AddressID)
}
