package enterprise_foundation

import (
	"context"
	"strings"

	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	"github.com/integrione/erp-retail/backend/gateway/internal/source_connections"
	efv1 "github.com/integrione/erp-retail/backend/services/enterprise_foundation/contract/gen/enterprise_foundation/v1"
)

func buildPartyRoleConn(
	after, before *string,
	first, last *int,
	fetch func(limit, offset int32) ([]*efv1.PartyRole, *efv1.PageInfo, error),
) (*model.PartyRoleConnection, error) {
	return buildRelayConnection(
		after, before, first, last,
		fetch,
		toModelPartyRole,
		func(node *model.PartyRole, cursor string) *model.PartyRoleEdge {
			return &model.PartyRoleEdge{Node: node, Cursor: cursor}
		},
		func(edges []*model.PartyRoleEdge, nodes []*model.PartyRole, pi *model.PageInfo, total int) *model.PartyRoleConnection {
			return &model.PartyRoleConnection{Edges: edges, Nodes: nodes, PageInfo: pi, TotalCount: total}
		},
	)
}

func buildPartyContactConn(
	after, before *string,
	first, last *int,
	fetch func(limit, offset int32) ([]*efv1.PartyContact, *efv1.PageInfo, error),
) (*model.PartyContactConnection, error) {
	return buildRelayConnection(
		after, before, first, last,
		fetch,
		toModelPartyContact,
		func(node *model.PartyContact, cursor string) *model.PartyContactEdge {
			return &model.PartyContactEdge{Node: node, Cursor: cursor}
		},
		func(edges []*model.PartyContactEdge, nodes []*model.PartyContact, pi *model.PageInfo, total int) *model.PartyContactConnection {
			return &model.PartyContactConnection{Edges: edges, Nodes: nodes, PageInfo: pi, TotalCount: total}
		},
	)
}

func buildPartyAddressConn(
	after, before *string,
	first, last *int,
	fetch func(limit, offset int32) ([]*efv1.PartyAddress, *efv1.PageInfo, error),
) (*model.PartyAddressConnection, error) {
	return buildRelayConnection(
		after, before, first, last,
		fetch,
		toModelPartyAddress,
		func(node *model.PartyAddress, cursor string) *model.PartyAddressEdge {
			return &model.PartyAddressEdge{Node: node, Cursor: cursor}
		},
		func(edges []*model.PartyAddressEdge, nodes []*model.PartyAddress, pi *model.PageInfo, total int) *model.PartyAddressConnection {
			return &model.PartyAddressConnection{Edges: edges, Nodes: nodes, PageInfo: pi, TotalCount: total}
		},
	)
}

// ==================== PartyRole Queries & Mutations ====================

// PartyRole mengambil satu penugasan peran bisnis mitra berdasarkan ID/GID.
func (a *Adapter) PartyRole(ctx context.Context, id string) (*model.PartyRole, error) {
	sess, err := a.requirePerm(ctx, PermPartiesRead, false)
	if err != nil {
		return nil, err
	}
	cleanID, err := ParseTypedID(id, ResourcePartyRole)
	if err != nil {
		return nil, err
	}
	resp, err := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.GetPartyRoleByIdResponse, error) {
		return c.Parties.GetPartyRoleById(rpcCtx, &efv1.GetPartyRoleByIdRequest{RoleId: cleanID})
	})
	if err != nil {
		return nil, err
	}
	return toModelPartyRole(resp.GetRole()), nil
}

// PartyRolesByParty mengambil daftar peran bisnis milik satu mitra.
func (a *Adapter) PartyRolesByParty(ctx context.Context, partyID string, after, before *string, first, last *int) (*model.PartyRoleConnection, error) {
	cleanPartyID, parseErr := ParseTypedID(partyID, ResourceParty)
	if parseErr != nil {
		if _, authErr := a.requirePerm(ctx, PermPartiesRead, false); authErr != nil {
			return nil, authErr
		}
		return nil, parseErr
	}
	sess, err := a.requireAnyScope(ctx, PermPartiesRead, false, partyScopes(cleanPartyID)...)
	if err != nil {
		return nil, err
	}
	return buildPartyRoleConn(after, before, first, last, func(limit, offset int32) ([]*efv1.PartyRole, *efv1.PageInfo, error) {
		resp, rpcErr := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.ListPartyRolesByPartyResponse, error) {
			return c.Parties.ListPartyRolesByParty(rpcCtx, &efv1.ListPartyRolesByPartyRequest{
				PartyId: cleanPartyID,
				Page:    &efv1.PageRequest{Limit: limit, Offset: offset},
			})
		})
		if rpcErr != nil {
			return nil, nil, rpcErr
		}
		return resp.GetItems(), resp.GetPage(), nil
	})
}

// PartyRoleByPartyAndKind mengambil satu peran spesifik milik mitra.
func (a *Adapter) PartyRoleByPartyAndKind(ctx context.Context, partyID string, kind model.PartyRoleKind) (*model.PartyRole, error) {
	cleanPartyID, parseErr := ParseTypedID(partyID, ResourceParty)
	if parseErr != nil {
		if _, authErr := a.requirePerm(ctx, PermPartiesRead, false); authErr != nil {
			return nil, authErr
		}
		return nil, parseErr
	}
	sess, err := a.requireAnyScope(ctx, PermPartiesRead, false, partyScopes(cleanPartyID)...)
	if err != nil {
		return nil, err
	}
	resp, err := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.GetPartyRoleByPartyAndKindResponse, error) {
		return c.Parties.GetPartyRoleByPartyAndKind(rpcCtx, &efv1.GetPartyRoleByPartyAndKindRequest{
			PartyId: cleanPartyID,
			Kind:    mapEnumValue(kind, partyRoleKindToProto, efv1.PartyRoleKind_PARTY_ROLE_KIND_UNSPECIFIED),
		})
	})
	if err != nil {
		return nil, err
	}
	return toModelPartyRole(resp.GetRole()), nil
}

// PartyRoles mengambil daftar peran mitra terpaginasi dengan filter opsional.
func (a *Adapter) PartyRoles(
	ctx context.Context,
	after, before *string,
	first, last *int,
	partyID *string,
	role *model.PartyRoleKind,
	activeOnly *bool,
	query *string,
) (*model.PartyRoleConnection, error) {
	cleanPartyID, parseErr := ParseOptionalTypedID(partyID, ResourceParty)
	var sess *middleware.ResolvedSession
	var authErr error
	if cleanPartyID != nil && parseErr == nil {
		sess, authErr = a.requireAnyScope(ctx, PermPartiesRead, false, partyScopes(*cleanPartyID)...)
	} else {
		sess, authErr = a.requirePerm(ctx, PermPartiesRead, false)
	}
	if authErr != nil {
		return nil, authErr
	}
	if parseErr != nil {
		return nil, parseErr
	}
	return buildPartyRoleConn(after, before, first, last, func(limit, offset int32) ([]*efv1.PartyRole, *efv1.PageInfo, error) {
		resp, rpcErr := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.ListPartyRolesResponse, error) {
			return c.Parties.ListPartyRoles(rpcCtx, &efv1.ListPartyRolesRequest{
				PartyId:    derefStr(cleanPartyID),
				Role:       mapOptionalEnumPtr(role, partyRoleKindToProto),
				ActiveOnly: derefBool(activeOnly, false),
				Query:      strings.TrimSpace(derefStr(query)),
				Page:       &efv1.PageRequest{Limit: limit, Offset: offset},
			})
		})
		if rpcErr != nil {
			return nil, nil, rpcErr
		}
		return resp.GetItems(), resp.GetPage(), nil
	})
}

// AllPartyRoles mengambil seluruh daftar peran bisnis mitra secara terpaginasi.
func (a *Adapter) AllPartyRoles(ctx context.Context, after, before *string, first, last *int) (*model.PartyRoleConnection, error) {
	sess, err := a.requirePerm(ctx, PermPartiesRead, false)
	if err != nil {
		return nil, err
	}
	return buildPartyRoleConn(after, before, first, last, func(limit, offset int32) ([]*efv1.PartyRole, *efv1.PageInfo, error) {
		resp, rpcErr := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.GetAllPartyRolesResponse, error) {
			return c.Parties.GetAllPartyRoles(rpcCtx, &efv1.GetAllPartyRolesRequest{
				Page: &efv1.PageRequest{Limit: limit, Offset: offset},
			})
		})
		if rpcErr != nil {
			return nil, nil, rpcErr
		}
		return resp.GetItems(), resp.GetPage(), nil
	})
}

// PartyRolesByKind mengambil daftar peran bisnis mitra menurut jenis peran tertentu.
func (a *Adapter) PartyRolesByKind(ctx context.Context, kind model.PartyRoleKind, activeOnly *bool, after, before *string, first, last *int) (*model.PartyRoleConnection, error) {
	sess, err := a.requirePerm(ctx, PermPartiesRead, false)
	if err != nil {
		return nil, err
	}
	protoKind := mapEnumValue(kind, partyRoleKindToProto, efv1.PartyRoleKind_PARTY_ROLE_KIND_UNSPECIFIED)
	return buildPartyRoleConn(after, before, first, last, func(limit, offset int32) ([]*efv1.PartyRole, *efv1.PageInfo, error) {
		resp, rpcErr := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.ListPartyRolesByKindResponse, error) {
			return c.Parties.ListPartyRolesByKind(rpcCtx, &efv1.ListPartyRolesByKindRequest{
				Kind:       protoKind,
				ActiveOnly: derefBool(activeOnly, false),
				Page:       &efv1.PageRequest{Limit: limit, Offset: offset},
			})
		})
		if rpcErr != nil {
			return nil, nil, rpcErr
		}
		return resp.GetItems(), resp.GetPage(), nil
	})
}

// AllPartyRolesByKind mengambil seluruh peran bisnis mitra menurut jenis peran tanpa filter tambahan.
func (a *Adapter) AllPartyRolesByKind(ctx context.Context, kind model.PartyRoleKind, activeOnly *bool, after, before *string, first, last *int) (*model.PartyRoleConnection, error) {
	sess, err := a.requirePerm(ctx, PermPartiesRead, false)
	if err != nil {
		return nil, err
	}
	protoKind := mapEnumValue(kind, partyRoleKindToProto, efv1.PartyRoleKind_PARTY_ROLE_KIND_UNSPECIFIED)
	return buildPartyRoleConn(after, before, first, last, func(limit, offset int32) ([]*efv1.PartyRole, *efv1.PageInfo, error) {
		resp, rpcErr := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.GetAllPartyRolesByKindResponse, error) {
			return c.Parties.GetAllPartyRolesByKind(rpcCtx, &efv1.GetAllPartyRolesByKindRequest{
				Kind:       protoKind,
				ActiveOnly: derefBool(activeOnly, false),
				Page:       &efv1.PageRequest{Limit: limit, Offset: offset},
			})
		})
		if rpcErr != nil {
			return nil, nil, rpcErr
		}
		return resp.GetItems(), resp.GetPage(), nil
	})
}

// ValidatePartyRoleScope memvalidasi kelayakan peran mitra sebagai target scope IAM.
func (a *Adapter) ValidatePartyRoleScope(ctx context.Context, partyID string, requiredRole model.PartyRoleKind) (*model.ScopeTargetDecision, error) {
	cleanPartyID, parseErr := ParseTypedID(partyID, ResourceParty)
	if parseErr != nil {
		if _, authErr := a.requirePerm(ctx, PermPartiesRead, false); authErr != nil {
			return nil, authErr
		}
		return nil, parseErr
	}
	sess, err := a.requireAnyScope(ctx, PermPartiesRead, false, partyScopes(cleanPartyID)...)
	if err != nil {
		return nil, err
	}
	resp, err := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.ValidatePartyRoleScopeResponse, error) {
		return c.Parties.ValidatePartyRoleScope(rpcCtx, &efv1.ValidatePartyRoleScopeRequest{
			PartyId:      cleanPartyID,
			RequiredRole: mapEnumValue(requiredRole, partyRoleKindToProto, efv1.PartyRoleKind_PARTY_ROLE_KIND_UNSPECIFIED),
		})
	})
	if err != nil {
		return nil, err
	}
	return toModelScopeTargetDecision(resp.GetDecision()), nil
}

// AssignPartyRole menugaskan peran bisnis kepada suatu mitra.
func (a *Adapter) AssignPartyRole(ctx context.Context, input model.AssignPartyRoleInput) *model.AssignPartyRolePayload {
	cleanPartyID, idErr := ParseTypedID(input.PartyID, ResourceParty)
	if idErr != nil {
		if _, authErr := a.requirePerm(ctx, PermPartiesWrite, false); authErr != nil {
			return &model.AssignPartyRolePayload{Errors: ToMutationErrors(authErr), WasSuccessful: false}
		}
		return &model.AssignPartyRolePayload{Errors: ToMutationErrors(idErr), WasSuccessful: false}
	}
	sess, err := a.requireAnyScope(ctx, PermPartiesWrite, false, partyScopes(cleanPartyID)...)
	if err != nil {
		return &model.AssignPartyRolePayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	resp, err := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.AssignPartyRoleResponse, error) {
		return c.Parties.AssignPartyRole(rpcCtx, &efv1.AssignPartyRoleRequest{
			PartyId:       cleanPartyID,
			Role:          mapEnumValue(input.Role, partyRoleKindToProto, efv1.PartyRoleKind_PARTY_ROLE_KIND_UNSPECIFIED),
			ReferenceNote: input.ReferenceNote,
		})
	})
	if err != nil {
		return &model.AssignPartyRolePayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	return &model.AssignPartyRolePayload{
		Role:          toModelPartyRole(resp.GetRole()),
		Errors:        NoMutationErrors(),
		WasSuccessful: true,
	}
}

// UpdatePartyRoleNote memperbarui catatan referensi peran bisnis mitra.
func (a *Adapter) UpdatePartyRoleNote(ctx context.Context, input model.UpdatePartyRoleNoteInput) *model.UpdatePartyRoleNotePayload {
	sess, err := a.requirePerm(ctx, PermPartiesWrite, false)
	if err != nil {
		return &model.UpdatePartyRoleNotePayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	cleanID, err := ParseTypedID(input.ID, ResourcePartyRole)
	if err != nil {
		return &model.UpdatePartyRoleNotePayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	resp, err := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.UpdatePartyRoleNoteResponse, error) {
		return c.Parties.UpdatePartyRoleNote(rpcCtx, &efv1.UpdatePartyRoleNoteRequest{
			Id:            cleanID,
			ReferenceNote: input.ReferenceNote,
		})
	})
	if err != nil {
		return &model.UpdatePartyRoleNotePayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	return &model.UpdatePartyRoleNotePayload{
		Role:          toModelPartyRole(resp.GetRole()),
		Errors:        NoMutationErrors(),
		WasSuccessful: true,
	}
}

// ActivatePartyRole mengaktifkan peran bisnis mitra.
func (a *Adapter) ActivatePartyRole(ctx context.Context, roleID string) *model.ActivatePartyRolePayload {
	sess, err := a.requirePerm(ctx, PermPartiesWrite, false)
	if err != nil {
		return &model.ActivatePartyRolePayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	cleanID, err := ParseTypedID(roleID, ResourcePartyRole)
	if err != nil {
		return &model.ActivatePartyRolePayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	resp, err := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.ActivatePartyRoleResponse, error) {
		return c.Parties.ActivatePartyRole(rpcCtx, &efv1.ActivatePartyRoleRequest{RoleId: cleanID})
	})
	if err != nil {
		return &model.ActivatePartyRolePayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	return &model.ActivatePartyRolePayload{
		Role:          toModelPartyRole(resp.GetRole()),
		Errors:        NoMutationErrors(),
		WasSuccessful: true,
	}
}

// DeactivatePartyRole menonaktifkan peran bisnis mitra.
func (a *Adapter) DeactivatePartyRole(ctx context.Context, roleID string) *model.DeactivatePartyRolePayload {
	sess, err := a.requirePerm(ctx, PermPartiesWrite, false)
	if err != nil {
		return &model.DeactivatePartyRolePayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	cleanID, err := ParseTypedID(roleID, ResourcePartyRole)
	if err != nil {
		return &model.DeactivatePartyRolePayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	resp, err := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.DeactivatePartyRoleResponse, error) {
		return c.Parties.DeactivatePartyRole(rpcCtx, &efv1.DeactivatePartyRoleRequest{RoleId: cleanID})
	})
	if err != nil {
		return &model.DeactivatePartyRolePayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	return &model.DeactivatePartyRolePayload{
		Role:          toModelPartyRole(resp.GetRole()),
		Errors:        NoMutationErrors(),
		WasSuccessful: true,
	}
}

// RemovePartyRole menghapus penugasan peran bisnis dari mitra dengan syarat re-autentikasi sensitif.
func (a *Adapter) RemovePartyRole(ctx context.Context, roleID string) *model.RemovePartyRolePayload {
	sess, err := a.requirePerm(ctx, PermPartiesWrite, true)
	if err != nil {
		return &model.RemovePartyRolePayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	cleanID, err := ParseTypedID(roleID, ResourcePartyRole)
	if err != nil {
		return &model.RemovePartyRolePayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	resp, err := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.RemovePartyRoleResponse, error) {
		return c.Parties.RemovePartyRole(rpcCtx, &efv1.RemovePartyRoleRequest{RoleId: cleanID})
	})
	if err != nil {
		return &model.RemovePartyRolePayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	removed := resp.GetRemoved()
	if !removed {
		return &model.RemovePartyRolePayload{
			Errors:        ToMutationErrors(middleware.ErrConflict("party role could not be removed", nil)),
			WasSuccessful: false,
		}
	}
	return &model.RemovePartyRolePayload{
		Removed:       &removed,
		Errors:        NoMutationErrors(),
		WasSuccessful: true,
	}
}

// ==================== PartyContact & Portal Access Queries & Mutations ====================

// PartyContact mengambil satu data PIC mitra berdasarkan ID/GID.
func (a *Adapter) PartyContact(ctx context.Context, id string) (*model.PartyContact, error) {
	sess, err := a.requirePerm(ctx, PermPartiesRead, false)
	if err != nil {
		return nil, err
	}
	cleanID, err := ParseTypedID(id, ResourcePartyContact)
	if err != nil {
		return nil, err
	}
	resp, err := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.GetPartyContactByIdResponse, error) {
		return c.Parties.GetPartyContactById(rpcCtx, &efv1.GetPartyContactByIdRequest{ContactId: cleanID})
	})
	if err != nil {
		return nil, err
	}
	return toModelPartyContact(resp.GetContact()), nil
}

// PartyContactsByParty mengambil daftar PIC milik satu mitra.
func (a *Adapter) PartyContactsByParty(ctx context.Context, partyID string, after, before *string, first, last *int) (*model.PartyContactConnection, error) {
	cleanPartyID, parseErr := ParseTypedID(partyID, ResourceParty)
	if parseErr != nil {
		if _, authErr := a.requirePerm(ctx, PermPartiesRead, false); authErr != nil {
			return nil, authErr
		}
		return nil, parseErr
	}
	sess, err := a.requireAnyScope(ctx, PermPartiesRead, false, partyScopes(cleanPartyID)...)
	if err != nil {
		return nil, err
	}
	return buildPartyContactConn(after, before, first, last, func(limit, offset int32) ([]*efv1.PartyContact, *efv1.PageInfo, error) {
		resp, rpcErr := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.ListPartyContactsByPartyResponse, error) {
			return c.Parties.ListPartyContactsByParty(rpcCtx, &efv1.ListPartyContactsByPartyRequest{
				PartyId: cleanPartyID,
				Page:    &efv1.PageRequest{Limit: limit, Offset: offset},
			})
		})
		if rpcErr != nil {
			return nil, nil, rpcErr
		}
		return resp.GetItems(), resp.GetPage(), nil
	})
}

// PrimaryPartyContact mengambil PIC utama milik suatu mitra.
func (a *Adapter) PrimaryPartyContact(ctx context.Context, partyID string) (*model.PartyContact, error) {
	cleanPartyID, parseErr := ParseTypedID(partyID, ResourceParty)
	if parseErr != nil {
		if _, authErr := a.requirePerm(ctx, PermPartiesRead, false); authErr != nil {
			return nil, authErr
		}
		return nil, parseErr
	}
	sess, err := a.requireAnyScope(ctx, PermPartiesRead, false, partyScopes(cleanPartyID)...)
	if err != nil {
		return nil, err
	}
	resp, err := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.GetPrimaryPartyContactResponse, error) {
		return c.Parties.GetPrimaryPartyContact(rpcCtx, &efv1.GetPrimaryPartyContactRequest{PartyId: cleanPartyID})
	})
	if err != nil {
		return nil, err
	}
	return toModelPartyContact(resp.GetContact()), nil
}

// PartyContactsByIamUserId mencari daftar PIC yang tertaut pada principal user IAM tertentu.
func (a *Adapter) PartyContactsByIamUserId(ctx context.Context, iamUserID string, after, before *string, first, last *int) (*model.PartyContactConnection, error) {
	sess, err := a.requirePerm(ctx, PermPartiesRead, false)
	if err != nil {
		return nil, err
	}
	cleanIAMUserID, err := ParseUUIDOrGID(iamUserID, ResourceIAMUser)
	if err != nil {
		return nil, err
	}
	return buildPartyContactConn(after, before, first, last, func(limit, offset int32) ([]*efv1.PartyContact, *efv1.PageInfo, error) {
		resp, rpcErr := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.ListPartyContactsByIamUserIdResponse, error) {
			return c.Parties.ListPartyContactsByIamUserId(rpcCtx, &efv1.ListPartyContactsByIamUserIdRequest{
				IamUserId: cleanIAMUserID,
				Page:      &efv1.PageRequest{Limit: limit, Offset: offset},
			})
		})
		if rpcErr != nil {
			return nil, nil, rpcErr
		}
		return resp.GetItems(), resp.GetPage(), nil
	})
}

// PartyContactsByEmail mencari daftar PIC berdasarkan alamat email.
func (a *Adapter) PartyContactsByEmail(ctx context.Context, email string, after, before *string, first, last *int) (*model.PartyContactConnection, error) {
	sess, err := a.requirePerm(ctx, PermPartiesRead, false)
	if err != nil {
		return nil, err
	}
	trimmed := strings.TrimSpace(email)
	if trimmed == "" {
		return nil, middleware.ErrInvalidArgument("email is required", nil)
	}
	return buildPartyContactConn(after, before, first, last, func(limit, offset int32) ([]*efv1.PartyContact, *efv1.PageInfo, error) {
		resp, rpcErr := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.ListPartyContactsByEmailResponse, error) {
			return c.Parties.ListPartyContactsByEmail(rpcCtx, &efv1.ListPartyContactsByEmailRequest{
				Email: trimmed,
				Page:  &efv1.PageRequest{Limit: limit, Offset: offset},
			})
		})
		if rpcErr != nil {
			return nil, nil, rpcErr
		}
		return resp.GetItems(), resp.GetPage(), nil
	})
}

// PartyContacts mengambil daftar PIC mitra secara terpaginasi beserta filter.
func (a *Adapter) PartyContacts(
	ctx context.Context,
	after, before *string,
	first, last *int,
	partyID, query, contactRole *string,
	portalStatus *model.PortalAccessStatus,
	activeOnly, includeDeleted *bool,
) (*model.PartyContactConnection, error) {
	cleanPartyID, parseErr := ParseOptionalTypedID(partyID, ResourceParty)
	var sess *middleware.ResolvedSession
	var authErr error
	if cleanPartyID != nil && parseErr == nil {
		sess, authErr = a.requireAnyScope(ctx, PermPartiesRead, false, partyScopes(*cleanPartyID)...)
	} else {
		sess, authErr = a.requirePerm(ctx, PermPartiesRead, false)
	}
	if authErr != nil {
		return nil, authErr
	}
	if parseErr != nil {
		return nil, parseErr
	}
	return buildPartyContactConn(after, before, first, last, func(limit, offset int32) ([]*efv1.PartyContact, *efv1.PageInfo, error) {
		resp, rpcErr := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.ListPartyContactsResponse, error) {
			return c.Parties.ListPartyContacts(rpcCtx, &efv1.ListPartyContactsRequest{
				PartyId:            derefStr(cleanPartyID),
				Query:              strings.TrimSpace(derefStr(query)),
				ContactRole:        strings.TrimSpace(derefStr(contactRole)),
				PortalAccessStatus: mapOptionalEnumPtr(portalStatus, portalAccessStatusToProto),
				ActiveOnly:         derefBool(activeOnly, false),
				IncludeDeleted:     derefBool(includeDeleted, false),
				Page:               &efv1.PageRequest{Limit: limit, Offset: offset},
			})
		})
		if rpcErr != nil {
			return nil, nil, rpcErr
		}
		return resp.GetItems(), resp.GetPage(), nil
	})
}

// AllPartyContacts mengambil seluruh PIC mitra terpaginasi.
func (a *Adapter) AllPartyContacts(ctx context.Context, after, before *string, first, last *int) (*model.PartyContactConnection, error) {
	sess, err := a.requirePerm(ctx, PermPartiesRead, false)
	if err != nil {
		return nil, err
	}
	return buildPartyContactConn(after, before, first, last, func(limit, offset int32) ([]*efv1.PartyContact, *efv1.PageInfo, error) {
		resp, rpcErr := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.GetAllPartyContactsResponse, error) {
			return c.Parties.GetAllPartyContacts(rpcCtx, &efv1.GetAllPartyContactsRequest{
				Page: &efv1.PageRequest{Limit: limit, Offset: offset},
			})
		})
		if rpcErr != nil {
			return nil, nil, rpcErr
		}
		return resp.GetItems(), resp.GetPage(), nil
	})
}

// EvaluateRepresentativePortalAccess mengevaluasi apakah principal IAM merupakan perwakilan aktif mitra untuk peran portal tertentu.
func (a *Adapter) EvaluateRepresentativePortalAccess(ctx context.Context, partyID, iamUserID string, requiredRole model.PartyRoleKind) (*model.PortalAccessEvaluation, error) {
	cleanPartyID, parseErr := ParseTypedID(partyID, ResourceParty)
	if parseErr != nil {
		if _, authErr := a.requirePerm(ctx, PermPartiesRead, false); authErr != nil {
			return nil, authErr
		}
		return nil, parseErr
	}
	sess, err := a.requireAnyScope(ctx, PermPartiesRead, false, partyScopes(cleanPartyID)...)
	if err != nil {
		return nil, err
	}
	cleanIAMUserID, err := ParseUUIDOrGID(iamUserID, ResourceIAMUser)
	if err != nil {
		return nil, err
	}
	resp, err := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.EvaluateRepresentativePortalAccessResponse, error) {
		return c.Parties.EvaluateRepresentativePortalAccess(rpcCtx, &efv1.EvaluateRepresentativePortalAccessRequest{
			PartyId:      cleanPartyID,
			IamUserId:    cleanIAMUserID,
			RequiredRole: mapEnumValue(requiredRole, partyRoleKindToProto, efv1.PartyRoleKind_PARTY_ROLE_KIND_UNSPECIFIED),
		})
	})
	if err != nil {
		return nil, err
	}
	return toModelPortalAccessEvaluation(resp.GetEvaluation()), nil
}

// AddPartyContact menambahkan PIC baru milik suatu mitra.
func (a *Adapter) AddPartyContact(ctx context.Context, input model.AddPartyContactInput) *model.AddPartyContactPayload {
	cleanPartyID, idErr := ParseTypedID(input.PartyID, ResourceParty)
	if idErr != nil {
		if _, authErr := a.requirePerm(ctx, PermPartiesWrite, false); authErr != nil {
			return &model.AddPartyContactPayload{Errors: ToMutationErrors(authErr), WasSuccessful: false}
		}
		return &model.AddPartyContactPayload{Errors: ToMutationErrors(idErr), WasSuccessful: false}
	}
	sess, err := a.requireAnyScope(ctx, PermPartiesWrite, false, partyScopes(cleanPartyID)...)
	if err != nil {
		return &model.AddPartyContactPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	iamUserID, err := ParseOptionalUUIDOrGID(input.IamUserID, ResourceIAMUser)
	if err != nil {
		return &model.AddPartyContactPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	portalStatus := efv1.PortalAccessStatus_PORTAL_ACCESS_STATUS_NONE
	if input.PortalAccessStatus != nil {
		portalStatus = mapEnumValue(*input.PortalAccessStatus, portalAccessStatusToProto, portalStatus)
	}
	resp, err := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.AddPartyContactResponse, error) {
		return c.Parties.AddPartyContact(rpcCtx, &efv1.AddPartyContactRequest{
			PartyId:            cleanPartyID,
			IamUserId:          iamUserID,
			FullName:           strings.TrimSpace(input.FullName),
			JobTitle:           input.JobTitle,
			ContactRole:        strings.TrimSpace(input.ContactRole),
			Email:              input.Email,
			Phone:              input.Phone,
			PortalAccessStatus: portalStatus,
			IsPrimary:          derefBool(input.IsPrimary, false),
		})
	})
	if err != nil {
		return &model.AddPartyContactPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	return &model.AddPartyContactPayload{
		Contact:       toModelPartyContact(resp.GetContact()),
		Errors:        NoMutationErrors(),
		WasSuccessful: true,
	}
}

// UpdatePartyContact memperbarui informasi profil PIC mitra.
func (a *Adapter) UpdatePartyContact(ctx context.Context, input model.UpdatePartyContactInput) *model.UpdatePartyContactPayload {
	sess, err := a.requirePerm(ctx, PermPartiesWrite, false)
	if err != nil {
		return &model.UpdatePartyContactPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	cleanID, err := ParseTypedID(input.ID, ResourcePartyContact)
	if err != nil {
		return &model.UpdatePartyContactPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	resp, err := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.UpdatePartyContactResponse, error) {
		return c.Parties.UpdatePartyContact(rpcCtx, &efv1.UpdatePartyContactRequest{
			Id:          cleanID,
			FullName:    strings.TrimSpace(input.FullName),
			JobTitle:    input.JobTitle,
			ContactRole: strings.TrimSpace(input.ContactRole),
			Email:       input.Email,
			Phone:       input.Phone,
			IsPrimary:   derefBool(input.IsPrimary, false),
		})
	})
	if err != nil {
		return &model.UpdatePartyContactPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	return &model.UpdatePartyContactPayload{
		Contact:       toModelPartyContact(resp.GetContact()),
		Errors:        NoMutationErrors(),
		WasSuccessful: true,
	}
}

// SetPrimaryPartyContact menetapkan PIC utama suatu mitra.
func (a *Adapter) SetPrimaryPartyContact(ctx context.Context, partyID, contactID string) *model.SetPrimaryPartyContactPayload {
	cleanPartyID, idErr := ParseTypedID(partyID, ResourceParty)
	if idErr != nil {
		if _, authErr := a.requirePerm(ctx, PermPartiesWrite, false); authErr != nil {
			return &model.SetPrimaryPartyContactPayload{Errors: ToMutationErrors(authErr), WasSuccessful: false}
		}
		return &model.SetPrimaryPartyContactPayload{Errors: ToMutationErrors(idErr), WasSuccessful: false}
	}
	sess, err := a.requireAnyScope(ctx, PermPartiesWrite, false, partyScopes(cleanPartyID)...)
	if err != nil {
		return &model.SetPrimaryPartyContactPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	cleanContactID, err := ParseTypedID(contactID, ResourcePartyContact)
	if err != nil {
		return &model.SetPrimaryPartyContactPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	resp, err := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.SetPrimaryPartyContactResponse, error) {
		return c.Parties.SetPrimaryPartyContact(rpcCtx, &efv1.SetPrimaryPartyContactRequest{
			PartyId:   cleanPartyID,
			ContactId: cleanContactID,
		})
	})
	if err != nil {
		return &model.SetPrimaryPartyContactPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	return &model.SetPrimaryPartyContactPayload{
		Contact:       toModelPartyContact(resp.GetContact()),
		Errors:        NoMutationErrors(),
		WasSuccessful: true,
	}
}

// RequestPartyPortalLink mengajukan penautan PIC mitra terhadap akun pengguna IAM.
func (a *Adapter) RequestPartyPortalLink(ctx context.Context, contactID, iamUserID string) *model.RequestPartyPortalLinkPayload {
	sess, err := a.requirePerm(ctx, PermPartiesWrite, false)
	if err != nil {
		return &model.RequestPartyPortalLinkPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	cleanContactID, err := ParseTypedID(contactID, ResourcePartyContact)
	if err != nil {
		return &model.RequestPartyPortalLinkPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	cleanIAMUserID, err := ParseUUIDOrGID(iamUserID, ResourceIAMUser)
	if err != nil {
		return &model.RequestPartyPortalLinkPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	resp, err := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.RequestPartyPortalLinkResponse, error) {
		return c.Parties.RequestPartyPortalLink(rpcCtx, &efv1.RequestPartyPortalLinkRequest{
			ContactId: cleanContactID,
			IamUserId: cleanIAMUserID,
		})
	})
	if err != nil {
		return &model.RequestPartyPortalLinkPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	return &model.RequestPartyPortalLinkPayload{
		Contact:       toModelPartyContact(resp.GetContact()),
		Errors:        NoMutationErrors(),
		WasSuccessful: true,
	}
}

// ActivatePartyPortalAccess mengaktifkan akses portal PIC mitra dengan syarat re-autentikasi sensitif.
func (a *Adapter) ActivatePartyPortalAccess(ctx context.Context, contactID, iamUserID string) *model.ActivatePartyPortalAccessPayload {
	sess, err := a.requirePerm(ctx, PermPartiesWrite, true)
	if err != nil {
		return &model.ActivatePartyPortalAccessPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	cleanContactID, err := ParseTypedID(contactID, ResourcePartyContact)
	if err != nil {
		return &model.ActivatePartyPortalAccessPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	cleanIAMUserID, err := ParseUUIDOrGID(iamUserID, ResourceIAMUser)
	if err != nil {
		return &model.ActivatePartyPortalAccessPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	resp, err := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.ActivatePartyPortalAccessResponse, error) {
		return c.Parties.ActivatePartyPortalAccess(rpcCtx, &efv1.ActivatePartyPortalAccessRequest{
			ContactId: cleanContactID,
			IamUserId: cleanIAMUserID,
		})
	})
	if err != nil {
		return &model.ActivatePartyPortalAccessPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	return &model.ActivatePartyPortalAccessPayload{
		Contact:       toModelPartyContact(resp.GetContact()),
		Errors:        NoMutationErrors(),
		WasSuccessful: true,
	}
}

// RevokePartyPortalAccess mencabut akses portal PIC mitra dengan syarat re-autentikasi sensitif.
func (a *Adapter) RevokePartyPortalAccess(ctx context.Context, contactID string) *model.RevokePartyPortalAccessPayload {
	sess, err := a.requirePerm(ctx, PermPartiesWrite, true)
	if err != nil {
		return &model.RevokePartyPortalAccessPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	cleanContactID, err := ParseTypedID(contactID, ResourcePartyContact)
	if err != nil {
		return &model.RevokePartyPortalAccessPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	resp, err := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.RevokePartyPortalAccessResponse, error) {
		return c.Parties.RevokePartyPortalAccess(rpcCtx, &efv1.RevokePartyPortalAccessRequest{ContactId: cleanContactID})
	})
	if err != nil {
		return &model.RevokePartyPortalAccessPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	return &model.RevokePartyPortalAccessPayload{
		Contact:       toModelPartyContact(resp.GetContact()),
		Errors:        NoMutationErrors(),
		WasSuccessful: true,
	}
}

// UnlinkPartyPortalUser memutus tautan akun IAM dari PIC mitra dengan syarat re-autentikasi sensitif.
func (a *Adapter) UnlinkPartyPortalUser(ctx context.Context, contactID string) *model.UnlinkPartyPortalUserPayload {
	sess, err := a.requirePerm(ctx, PermPartiesWrite, true)
	if err != nil {
		return &model.UnlinkPartyPortalUserPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	cleanContactID, err := ParseTypedID(contactID, ResourcePartyContact)
	if err != nil {
		return &model.UnlinkPartyPortalUserPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	resp, err := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.UnlinkPartyPortalUserResponse, error) {
		return c.Parties.UnlinkPartyPortalUser(rpcCtx, &efv1.UnlinkPartyPortalUserRequest{ContactId: cleanContactID})
	})
	if err != nil {
		return &model.UnlinkPartyPortalUserPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	return &model.UnlinkPartyPortalUserPayload{
		Contact:       toModelPartyContact(resp.GetContact()),
		Errors:        NoMutationErrors(),
		WasSuccessful: true,
	}
}

// ActivatePartyContact mengaktifkan PIC mitra.
func (a *Adapter) ActivatePartyContact(ctx context.Context, contactID string) *model.ActivatePartyContactPayload {
	sess, err := a.requirePerm(ctx, PermPartiesWrite, false)
	if err != nil {
		return &model.ActivatePartyContactPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	cleanContactID, err := ParseTypedID(contactID, ResourcePartyContact)
	if err != nil {
		return &model.ActivatePartyContactPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	resp, err := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.ActivatePartyContactResponse, error) {
		return c.Parties.ActivatePartyContact(rpcCtx, &efv1.ActivatePartyContactRequest{ContactId: cleanContactID})
	})
	if err != nil {
		return &model.ActivatePartyContactPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	return &model.ActivatePartyContactPayload{
		Contact:       toModelPartyContact(resp.GetContact()),
		Errors:        NoMutationErrors(),
		WasSuccessful: true,
	}
}

// DeactivatePartyContact menonaktifkan PIC mitra.
func (a *Adapter) DeactivatePartyContact(ctx context.Context, contactID string) *model.DeactivatePartyContactPayload {
	sess, err := a.requirePerm(ctx, PermPartiesWrite, false)
	if err != nil {
		return &model.DeactivatePartyContactPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	cleanContactID, err := ParseTypedID(contactID, ResourcePartyContact)
	if err != nil {
		return &model.DeactivatePartyContactPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	resp, err := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.DeactivatePartyContactResponse, error) {
		return c.Parties.DeactivatePartyContact(rpcCtx, &efv1.DeactivatePartyContactRequest{ContactId: cleanContactID})
	})
	if err != nil {
		return &model.DeactivatePartyContactPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	return &model.DeactivatePartyContactPayload{
		Contact:       toModelPartyContact(resp.GetContact()),
		Errors:        NoMutationErrors(),
		WasSuccessful: true,
	}
}

// DeletePartyContact melakukan soft delete PIC mitra dengan wajib re-autentikasi sensitif.
func (a *Adapter) DeletePartyContact(ctx context.Context, contactID string) *model.DeletePartyContactPayload {
	sess, err := a.requirePerm(ctx, PermPartiesWrite, true)
	if err != nil {
		return &model.DeletePartyContactPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	cleanContactID, err := ParseTypedID(contactID, ResourcePartyContact)
	if err != nil {
		return &model.DeletePartyContactPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	resp, err := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.DeletePartyContactResponse, error) {
		return c.Parties.DeletePartyContact(rpcCtx, &efv1.DeletePartyContactRequest{ContactId: cleanContactID})
	})
	if err != nil {
		return &model.DeletePartyContactPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	deleted := resp.GetDeleted()
	if !deleted {
		return &model.DeletePartyContactPayload{
			Errors:        ToMutationErrors(middleware.ErrConflict("party contact could not be deleted", nil)),
			WasSuccessful: false,
		}
	}
	return &model.DeletePartyContactPayload{
		Deleted:       &deleted,
		Errors:        NoMutationErrors(),
		WasSuccessful: true,
	}
}

// ==================== PartyAddressLink Queries & Mutations ====================

// PartyAddressLink mengambil satu relasi alamat mitra berdasarkan ID/GID.
func (a *Adapter) PartyAddressLink(ctx context.Context, id string) (*model.PartyAddress, error) {
	sess, err := a.requirePerm(ctx, PermPartiesRead, false)
	if err != nil {
		return nil, err
	}
	cleanID, err := ParseTypedID(id, ResourcePartyAddress)
	if err != nil {
		return nil, err
	}
	resp, err := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.GetPartyAddressLinkByIdResponse, error) {
		return c.Parties.GetPartyAddressLinkById(rpcCtx, &efv1.GetPartyAddressLinkByIdRequest{LinkId: cleanID})
	})
	if err != nil {
		return nil, err
	}
	return toModelPartyAddress(resp.GetLink()), nil
}

// PartyAddressLinksByParty mengambil daftar tautan alamat milik satu mitra.
func (a *Adapter) PartyAddressLinksByParty(ctx context.Context, partyID string, after, before *string, first, last *int) (*model.PartyAddressConnection, error) {
	cleanPartyID, parseErr := ParseTypedID(partyID, ResourceParty)
	if parseErr != nil {
		if _, authErr := a.requirePerm(ctx, PermPartiesRead, false); authErr != nil {
			return nil, authErr
		}
		return nil, parseErr
	}
	sess, err := a.requireAnyScope(ctx, PermPartiesRead, false, partyScopes(cleanPartyID)...)
	if err != nil {
		return nil, err
	}
	return buildPartyAddressConn(after, before, first, last, func(limit, offset int32) ([]*efv1.PartyAddress, *efv1.PageInfo, error) {
		resp, rpcErr := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.ListPartyAddressLinksByPartyResponse, error) {
			return c.Parties.ListPartyAddressLinksByParty(rpcCtx, &efv1.ListPartyAddressLinksByPartyRequest{
				PartyId: cleanPartyID,
				Page:    &efv1.PageRequest{Limit: limit, Offset: offset},
			})
		})
		if rpcErr != nil {
			return nil, nil, rpcErr
		}
		return resp.GetItems(), resp.GetPage(), nil
	})
}

// PartyAddressLinksByPartyAndUsage mengambil tautan alamat milik satu mitra berdasarkan peruntukannya.
func (a *Adapter) PartyAddressLinksByPartyAndUsage(
	ctx context.Context,
	partyID string,
	usage model.AddressUsageKind,
	after, before *string,
	first, last *int,
) (*model.PartyAddressConnection, error) {
	cleanPartyID, parseErr := ParseTypedID(partyID, ResourceParty)
	if parseErr != nil {
		if _, authErr := a.requirePerm(ctx, PermPartiesRead, false); authErr != nil {
			return nil, authErr
		}
		return nil, parseErr
	}
	sess, err := a.requireAnyScope(ctx, PermPartiesRead, false, partyScopes(cleanPartyID)...)
	if err != nil {
		return nil, err
	}
	protoUsage := mapEnumValue(usage, addressUsageKindToProto, efv1.AddressUsageKind_ADDRESS_USAGE_KIND_UNSPECIFIED)
	return buildPartyAddressConn(after, before, first, last, func(limit, offset int32) ([]*efv1.PartyAddress, *efv1.PageInfo, error) {
		resp, rpcErr := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.ListPartyAddressLinksByPartyAndUsageResponse, error) {
			return c.Parties.ListPartyAddressLinksByPartyAndUsage(rpcCtx, &efv1.ListPartyAddressLinksByPartyAndUsageRequest{
				PartyId: cleanPartyID,
				Usage:   protoUsage,
				Page:    &efv1.PageRequest{Limit: limit, Offset: offset},
			})
		})
		if rpcErr != nil {
			return nil, nil, rpcErr
		}
		return resp.GetItems(), resp.GetPage(), nil
	})
}

// PrimaryPartyAddressLink mengambil alamat utama milik satu mitra untuk peruntukan tertentu.
func (a *Adapter) PrimaryPartyAddressLink(ctx context.Context, partyID string, usage model.AddressUsageKind) (*model.PartyAddress, error) {
	cleanPartyID, parseErr := ParseTypedID(partyID, ResourceParty)
	if parseErr != nil {
		if _, authErr := a.requirePerm(ctx, PermPartiesRead, false); authErr != nil {
			return nil, authErr
		}
		return nil, parseErr
	}
	sess, err := a.requireAnyScope(ctx, PermPartiesRead, false, partyScopes(cleanPartyID)...)
	if err != nil {
		return nil, err
	}
	resp, err := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.GetPrimaryPartyAddressLinkResponse, error) {
		return c.Parties.GetPrimaryPartyAddressLink(rpcCtx, &efv1.GetPrimaryPartyAddressLinkRequest{
			PartyId: cleanPartyID,
			Usage:   mapEnumValue(usage, addressUsageKindToProto, efv1.AddressUsageKind_ADDRESS_USAGE_KIND_UNSPECIFIED),
		})
	})
	if err != nil {
		return nil, err
	}
	return toModelPartyAddress(resp.GetLink()), nil
}

// PartyAddressLinksByAddressId mengambil daftar tautan mitra yang merujuk pada satu alamat.
func (a *Adapter) PartyAddressLinksByAddressId(ctx context.Context, addressID string, after, before *string, first, last *int) (*model.PartyAddressConnection, error) {
	sess, err := a.requirePerm(ctx, PermPartiesRead, false)
	if err != nil {
		return nil, err
	}
	cleanAddrID, err := ParseTypedID(addressID, ResourceAddress)
	if err != nil {
		return nil, err
	}
	return buildPartyAddressConn(after, before, first, last, func(limit, offset int32) ([]*efv1.PartyAddress, *efv1.PageInfo, error) {
		resp, rpcErr := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.ListPartyAddressLinksByAddressIdResponse, error) {
			return c.Parties.ListPartyAddressLinksByAddressId(rpcCtx, &efv1.ListPartyAddressLinksByAddressIdRequest{
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

// PartyAddressLinks mengambil daftar tautan alamat mitra terpaginasi dengan filter opsional.
func (a *Adapter) PartyAddressLinks(
	ctx context.Context,
	after, before *string,
	first, last *int,
	partyID, addressID *string,
	usage *model.AddressUsageKind,
	primaryOnly *bool,
) (*model.PartyAddressConnection, error) {
	cleanPartyID, partyErr := ParseOptionalTypedID(partyID, ResourceParty)
	var sess *middleware.ResolvedSession
	var authErr error
	if cleanPartyID != nil && partyErr == nil {
		sess, authErr = a.requireAnyScope(ctx, PermPartiesRead, false, partyScopes(*cleanPartyID)...)
	} else {
		sess, authErr = a.requirePerm(ctx, PermPartiesRead, false)
	}
	if authErr != nil {
		return nil, authErr
	}
	if partyErr != nil {
		return nil, partyErr
	}
	cleanAddrID, err := ParseOptionalTypedID(addressID, ResourceAddress)
	if err != nil {
		return nil, err
	}
	return buildPartyAddressConn(after, before, first, last, func(limit, offset int32) ([]*efv1.PartyAddress, *efv1.PageInfo, error) {
		resp, rpcErr := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.ListPartyAddressLinksResponse, error) {
			return c.Parties.ListPartyAddressLinks(rpcCtx, &efv1.ListPartyAddressLinksRequest{
				PartyId:     derefStr(cleanPartyID),
				AddressId:   derefStr(cleanAddrID),
				Usage:       mapOptionalEnumPtr(usage, addressUsageKindToProto),
				PrimaryOnly: primaryOnly,
				Page:        &efv1.PageRequest{Limit: limit, Offset: offset},
			})
		})
		if rpcErr != nil {
			return nil, nil, rpcErr
		}
		return resp.GetItems(), resp.GetPage(), nil
	})
}

// AllPartyAddressLinks mengambil seluruh relasi alamat mitra terpaginasi.
func (a *Adapter) AllPartyAddressLinks(ctx context.Context, after, before *string, first, last *int) (*model.PartyAddressConnection, error) {
	sess, err := a.requirePerm(ctx, PermPartiesRead, false)
	if err != nil {
		return nil, err
	}
	return buildPartyAddressConn(after, before, first, last, func(limit, offset int32) ([]*efv1.PartyAddress, *efv1.PageInfo, error) {
		resp, rpcErr := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.GetAllPartyAddressLinksResponse, error) {
			return c.Parties.GetAllPartyAddressLinks(rpcCtx, &efv1.GetAllPartyAddressLinksRequest{
				Page: &efv1.PageRequest{Limit: limit, Offset: offset},
			})
		})
		if rpcErr != nil {
			return nil, nil, rpcErr
		}
		return resp.GetItems(), resp.GetPage(), nil
	})
}

// LinkPartyAddress menautkan suatu alamat ke mitra sesuai peruntukan.
func (a *Adapter) LinkPartyAddress(ctx context.Context, input model.LinkPartyAddressInput) *model.LinkPartyAddressPayload {
	cleanPartyID, idErr := ParseTypedID(input.PartyID, ResourceParty)
	if idErr != nil {
		if _, authErr := a.requirePerm(ctx, PermPartiesWrite, false); authErr != nil {
			return &model.LinkPartyAddressPayload{Errors: ToMutationErrors(authErr), WasSuccessful: false}
		}
		return &model.LinkPartyAddressPayload{Errors: ToMutationErrors(idErr), WasSuccessful: false}
	}
	sess, err := a.requireAnyScope(ctx, PermPartiesWrite, false, partyScopes(cleanPartyID)...)
	if err != nil {
		return &model.LinkPartyAddressPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	cleanAddrID, err := ParseTypedID(input.AddressID, ResourceAddress)
	if err != nil {
		return &model.LinkPartyAddressPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	resp, err := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.LinkPartyAddressResponse, error) {
		return c.Parties.LinkPartyAddress(rpcCtx, &efv1.LinkPartyAddressRequest{
			PartyId:        cleanPartyID,
			AddressId:      cleanAddrID,
			Usage:          mapEnumValue(input.Usage, addressUsageKindToProto, efv1.AddressUsageKind_ADDRESS_USAGE_KIND_UNSPECIFIED),
			IsPrimary:      derefBool(input.IsPrimary, false),
			ReplacePrimary: derefBool(input.ReplacePrimary, false),
		})
	})
	if err != nil {
		return &model.LinkPartyAddressPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	return &model.LinkPartyAddressPayload{
		Link:          toModelPartyAddress(resp.GetLink()),
		Errors:        NoMutationErrors(),
		WasSuccessful: true,
	}
}

// UpdatePartyAddressLink memperbarui jenis penggunaan atau status utama relasi alamat mitra.
func (a *Adapter) UpdatePartyAddressLink(ctx context.Context, input model.UpdatePartyAddressLinkInput) *model.UpdatePartyAddressLinkPayload {
	sess, err := a.requirePerm(ctx, PermPartiesWrite, false)
	if err != nil {
		return &model.UpdatePartyAddressLinkPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	cleanID, err := ParseTypedID(input.ID, ResourcePartyAddress)
	if err != nil {
		return &model.UpdatePartyAddressLinkPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	resp, err := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.UpdatePartyAddressLinkResponse, error) {
		return c.Parties.UpdatePartyAddressLink(rpcCtx, &efv1.UpdatePartyAddressLinkRequest{
			Id:             cleanID,
			Usage:          mapEnumValue(input.Usage, addressUsageKindToProto, efv1.AddressUsageKind_ADDRESS_USAGE_KIND_UNSPECIFIED),
			IsPrimary:      derefBool(input.IsPrimary, false),
			ReplacePrimary: derefBool(input.ReplacePrimary, false),
		})
	})
	if err != nil {
		return &model.UpdatePartyAddressLinkPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	return &model.UpdatePartyAddressLinkPayload{
		Link:          toModelPartyAddress(resp.GetLink()),
		Errors:        NoMutationErrors(),
		WasSuccessful: true,
	}
}

// SetPrimaryPartyAddressLink menetapkan satu tautan alamat sebagai alamat utama untuk peruntukannya.
func (a *Adapter) SetPrimaryPartyAddressLink(ctx context.Context, linkID string) *model.SetPrimaryPartyAddressLinkPayload {
	sess, err := a.requirePerm(ctx, PermPartiesWrite, false)
	if err != nil {
		return &model.SetPrimaryPartyAddressLinkPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	cleanID, err := ParseTypedID(linkID, ResourcePartyAddress)
	if err != nil {
		return &model.SetPrimaryPartyAddressLinkPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	resp, err := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.SetPrimaryPartyAddressLinkResponse, error) {
		return c.Parties.SetPrimaryPartyAddressLink(rpcCtx, &efv1.SetPrimaryPartyAddressLinkRequest{LinkId: cleanID})
	})
	if err != nil {
		return &model.SetPrimaryPartyAddressLinkPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	return &model.SetPrimaryPartyAddressLinkPayload{
		Link:          toModelPartyAddress(resp.GetLink()),
		Errors:        NoMutationErrors(),
		WasSuccessful: true,
	}
}

// UnlinkPartyAddress menghapus relasi alamat dari mitra.
func (a *Adapter) UnlinkPartyAddress(ctx context.Context, linkID string) *model.UnlinkPartyAddressPayload {
	sess, err := a.requirePerm(ctx, PermPartiesWrite, false)
	if err != nil {
		return &model.UnlinkPartyAddressPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	cleanID, err := ParseTypedID(linkID, ResourcePartyAddress)
	if err != nil {
		return &model.UnlinkPartyAddressPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	resp, err := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.UnlinkPartyAddressResponse, error) {
		return c.Parties.UnlinkPartyAddress(rpcCtx, &efv1.UnlinkPartyAddressRequest{LinkId: cleanID})
	})
	if err != nil {
		return &model.UnlinkPartyAddressPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	unlinked := resp.GetUnlinked()
	if !unlinked {
		return &model.UnlinkPartyAddressPayload{
			Errors:        ToMutationErrors(middleware.ErrConflict("party address could not be unlinked", nil)),
			WasSuccessful: false,
		}
	}
	return &model.UnlinkPartyAddressPayload{
		Unlinked:      &unlinked,
		Errors:        NoMutationErrors(),
		WasSuccessful: true,
	}
}

// ResolvePartyAddressDetail mengambil detail master Address milik relasi PartyAddress secara aman.
func (a *Adapter) ResolvePartyAddressDetail(ctx context.Context, obj *model.PartyAddress) (*model.Address, error) {
	if obj == nil || strings.TrimSpace(obj.AddressID) == "" {
		return nil, nil
	}
	return a.Address(ctx, obj.AddressID)
}
