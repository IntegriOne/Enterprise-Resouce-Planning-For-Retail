package enterprise_foundation

import (
	"context"
	"strings"

	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	"github.com/integrione/erp-retail/backend/gateway/internal/source_connections"
	efv1 "github.com/integrione/erp-retail/backend/services/enterprise_foundation/contract/gen/enterprise_foundation/v1"
)

func buildAddressConn(
	after, before *string,
	first, last *int,
	fetch func(limit, offset int32) ([]*efv1.Address, *efv1.PageInfo, error),
) (*model.AddressConnection, error) {
	return buildRelayConnection(
		after, before, first, last,
		fetch,
		toModelAddress,
		func(node *model.Address, cursor string) *model.AddressEdge {
			return &model.AddressEdge{Node: node, Cursor: cursor}
		},
		func(edges []*model.AddressEdge, nodes []*model.Address, pi *model.PageInfo, total int) *model.AddressConnection {
			return &model.AddressConnection{Edges: edges, Nodes: nodes, PageInfo: pi, TotalCount: total}
		},
	)
}

// Address mengambil satu master alamat berdasarkan ID/GID.
func (a *Adapter) Address(ctx context.Context, id string) (*model.Address, error) {
	sess, err := a.requirePerm(ctx, PermAddressesRead, false)
	if err != nil {
		return nil, err
	}
	cleanID, err := ParseTypedID(id, ResourceAddress)
	if err != nil {
		return nil, err
	}
	resp, err := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.GetAddressByIdResponse, error) {
		return c.Addresses.GetAddressById(rpcCtx, &efv1.GetAddressByIdRequest{Id: cleanID})
	})
	if err != nil {
		return nil, err
	}
	return toModelAddress(resp.GetAddress()), nil
}

// AddressesByIds mengambil banyak master alamat secara batch berdasarkan daftar ID/GID.
func (a *Adapter) AddressesByIds(ctx context.Context, ids []string) ([]*model.Address, error) {
	sess, err := a.requirePerm(ctx, PermAddressesRead, false)
	if err != nil {
		return nil, err
	}
	cleanIDs, err := ParseTypedIDList(ids, ResourceAddress)
	if err != nil {
		return nil, err
	}
	if len(cleanIDs) == 0 {
		return []*model.Address{}, nil
	}
	resp, err := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.GetAddressesByIdsResponse, error) {
		return c.Addresses.GetAddressesByIds(rpcCtx, &efv1.GetAddressesByIdsRequest{Ids: cleanIDs})
	})
	if err != nil {
		return nil, err
	}
	return toModelAddresses(resp.GetItems()), nil
}

// Addresses mengambil daftar alamat terpaginasi dengan filter teks serta wilayah geografis.
func (a *Adapter) Addresses(
	ctx context.Context,
	after, before *string,
	first, last *int,
	query, countryCode, city, province, postalCode *string,
) (*model.AddressConnection, error) {
	sess, err := a.requirePerm(ctx, PermAddressesRead, false)
	if err != nil {
		return nil, err
	}
	return buildAddressConn(after, before, first, last, func(limit, offset int32) ([]*efv1.Address, *efv1.PageInfo, error) {
		resp, rpcErr := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.ListAddressesResponse, error) {
			return c.Addresses.ListAddresses(rpcCtx, &efv1.ListAddressesRequest{
				Query:       strings.TrimSpace(derefStr(query)),
				CountryCode: strings.TrimSpace(derefStr(countryCode)),
				City:        strings.TrimSpace(derefStr(city)),
				Province:    strings.TrimSpace(derefStr(province)),
				PostalCode:  strings.TrimSpace(derefStr(postalCode)),
				Page:        &efv1.PageRequest{Limit: limit, Offset: offset},
			})
		})
		if rpcErr != nil {
			return nil, nil, rpcErr
		}
		return resp.GetItems(), resp.GetPage(), nil
	})
}

// AllAddresses mengambil seluruh master alamat terpaginasi tanpa filter opsional.
func (a *Adapter) AllAddresses(ctx context.Context, after, before *string, first, last *int) (*model.AddressConnection, error) {
	sess, err := a.requirePerm(ctx, PermAddressesRead, false)
	if err != nil {
		return nil, err
	}
	return buildAddressConn(after, before, first, last, func(limit, offset int32) ([]*efv1.Address, *efv1.PageInfo, error) {
		resp, rpcErr := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.GetAllAddressesResponse, error) {
			return c.Addresses.GetAllAddresses(rpcCtx, &efv1.GetAllAddressesRequest{
				Page: &efv1.PageRequest{Limit: limit, Offset: offset},
			})
		})
		if rpcErr != nil {
			return nil, nil, rpcErr
		}
		return resp.GetItems(), resp.GetPage(), nil
	})
}

// IsAddressReferenced memeriksa apakah suatu alamat masih dipakai oleh entitas lain.
func (a *Adapter) IsAddressReferenced(ctx context.Context, id string) (bool, error) {
	sess, err := a.requirePerm(ctx, PermAddressesRead, false)
	if err != nil {
		return false, err
	}
	cleanID, err := ParseTypedID(id, ResourceAddress)
	if err != nil {
		return false, err
	}
	resp, err := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.IsAddressReferencedResponse, error) {
		return c.Addresses.IsAddressReferenced(rpcCtx, &efv1.IsAddressReferencedRequest{Id: cleanID})
	})
	if err != nil {
		return false, err
	}
	return resp.GetReferenced(), nil
}

// CreateAddress membuat entitas alamat baru.
func (a *Adapter) CreateAddress(ctx context.Context, input model.CreateAddressInput) *model.CreateAddressPayload {
	sess, err := a.requirePerm(ctx, PermAddressesWrite, false)
	if err != nil {
		return &model.CreateAddressPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	resp, err := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.CreateAddressResponse, error) {
		return c.Addresses.CreateAddress(rpcCtx, &efv1.CreateAddressRequest{
			Label:            input.Label,
			AddressLine_1:    strings.TrimSpace(input.AddressLine1),
			AddressLine_2:    input.AddressLine2,
			City:             strings.TrimSpace(input.City),
			Province:         input.Province,
			PostalCode:       input.PostalCode,
			CountryCode:      strings.TrimSpace(input.CountryCode),
			FormattedAddress: input.FormattedAddress,
			Latitude:         input.Latitude,
			Longitude:        input.Longitude,
		})
	})
	if err != nil {
		return &model.CreateAddressPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	return &model.CreateAddressPayload{
		Address:       toModelAddress(resp.GetAddress()),
		Errors:        NoMutationErrors(),
		WasSuccessful: true,
	}
}

// FindOrCreateAddress mencari alamat identik yang sudah ada atau membuatnya secara idempoten.
func (a *Adapter) FindOrCreateAddress(ctx context.Context, input model.FindOrCreateAddressInput) *model.FindOrCreateAddressPayload {
	sess, err := a.requirePerm(ctx, PermAddressesWrite, false)
	if err != nil {
		return &model.FindOrCreateAddressPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	resp, err := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.FindOrCreateAddressResponse, error) {
		return c.Addresses.FindOrCreateAddress(rpcCtx, &efv1.FindOrCreateAddressRequest{
			Label:            input.Label,
			AddressLine_1:    strings.TrimSpace(input.AddressLine1),
			AddressLine_2:    input.AddressLine2,
			City:             strings.TrimSpace(input.City),
			Province:         input.Province,
			PostalCode:       input.PostalCode,
			CountryCode:      strings.TrimSpace(input.CountryCode),
			FormattedAddress: input.FormattedAddress,
			Latitude:         input.Latitude,
			Longitude:        input.Longitude,
		})
	})
	if err != nil {
		return &model.FindOrCreateAddressPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	created := resp.GetCreated()
	return &model.FindOrCreateAddressPayload{
		Address:       toModelAddress(resp.GetAddress()),
		Created:       &created,
		Errors:        NoMutationErrors(),
		WasSuccessful: true,
	}
}

// UpdateAddress memperbarui atribut master alamat yang sudah terdaftar.
func (a *Adapter) UpdateAddress(ctx context.Context, input model.UpdateAddressInput) *model.UpdateAddressPayload {
	sess, err := a.requirePerm(ctx, PermAddressesWrite, false)
	if err != nil {
		return &model.UpdateAddressPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	cleanID, err := ParseTypedID(input.ID, ResourceAddress)
	if err != nil {
		return &model.UpdateAddressPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	resp, err := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.UpdateAddressResponse, error) {
		return c.Addresses.UpdateAddress(rpcCtx, &efv1.UpdateAddressRequest{
			Id:               cleanID,
			Label:            input.Label,
			AddressLine_1:    strings.TrimSpace(input.AddressLine1),
			AddressLine_2:    input.AddressLine2,
			City:             strings.TrimSpace(input.City),
			Province:         input.Province,
			PostalCode:       input.PostalCode,
			CountryCode:      strings.TrimSpace(input.CountryCode),
			FormattedAddress: input.FormattedAddress,
			Latitude:         input.Latitude,
			Longitude:        input.Longitude,
		})
	})
	if err != nil {
		return &model.UpdateAddressPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	return &model.UpdateAddressPayload{
		Address:       toModelAddress(resp.GetAddress()),
		Errors:        NoMutationErrors(),
		WasSuccessful: true,
	}
}

// DeleteAddress menghapus master alamat yang tidak lagi direferensikan dengan kewajiban re-autentikasi sensitif.
func (a *Adapter) DeleteAddress(ctx context.Context, id string) *model.DeleteAddressPayload {
	sess, err := a.requirePerm(ctx, PermAddressesWrite, true)
	if err != nil {
		return &model.DeleteAddressPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	cleanID, err := ParseTypedID(id, ResourceAddress)
	if err != nil {
		return &model.DeleteAddressPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	resp, err := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.DeleteAddressResponse, error) {
		return c.Addresses.DeleteAddress(rpcCtx, &efv1.DeleteAddressRequest{Id: cleanID})
	})
	if err != nil {
		return &model.DeleteAddressPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	deleted := resp.GetDeleted()
	if !deleted {
		return &model.DeleteAddressPayload{
			Errors:        ToMutationErrors(middleware.ErrConflict("address could not be deleted", nil)),
			WasSuccessful: false,
		}
	}
	return &model.DeleteAddressPayload{
		Deleted:       &deleted,
		Errors:        NoMutationErrors(),
		WasSuccessful: true,
	}
}
