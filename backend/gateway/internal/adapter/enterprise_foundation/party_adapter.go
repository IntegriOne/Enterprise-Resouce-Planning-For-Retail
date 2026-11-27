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

func buildPartyConn(
	after, before *string,
	first, last *int,
	fetch func(limit, offset int32) ([]*efv1.Party, *efv1.PageInfo, error),
) (*model.PartyConnection, error) {
	return buildRelayConnection(
		after, before, first, last,
		fetch,
		toModelParty,
		func(node *model.Party, cursor string) *model.PartyEdge {
			return &model.PartyEdge{Node: node, Cursor: cursor}
		},
		func(edges []*model.PartyEdge, nodes []*model.Party, pi *model.PageInfo, total int) *model.PartyConnection {
			return &model.PartyConnection{Edges: edges, Nodes: nodes, PageInfo: pi, TotalCount: total}
		},
	)
}

// Party mengambil satu entitas mitra eksternal berdasarkan ID/GID dengan pemeriksaan scope vendor/supplier.
func (a *Adapter) Party(ctx context.Context, id string) (*model.Party, error) {
	cleanID, err := ParseTypedID(id, ResourceParty)
	if err != nil {
		if _, authErr := a.requirePerm(ctx, PermPartiesRead, false); authErr != nil {
			return nil, authErr
		}
		return nil, err
	}
	sess, err := a.requireAnyScope(ctx, PermPartiesRead, false, partyScopes(cleanID)...)
	if err != nil {
		return nil, err
	}
	resp, err := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.GetPartyByIdResponse, error) {
		return c.Parties.GetPartyById(rpcCtx, &efv1.GetPartyByIdRequest{Id: cleanID})
	})
	if err != nil {
		return nil, err
	}
	return toModelParty(resp.GetParty()), nil
}

// PartiesByIds mengambil daftar mitra eksternal secara batch dan memeriksa scope per ID target.
func (a *Adapter) PartiesByIds(ctx context.Context, ids []string) ([]*model.Party, error) {
	sess, err := a.authorizer.RequirePermission(ctx, PermPartiesRead, authz.ScopeGrant{Type: authz.ScopeGlobal}, false)
	hasGlobal := err == nil
	if err != nil {
		gwErr, ok := middleware.AsGatewayError(err)
		if !ok || gwErr.Code != middleware.CodeForbidden {
			return nil, err
		}
	}

	cleanIDs, parseErr := ParseTypedIDList(ids, ResourceParty)
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
		return []*model.Party{}, nil
	}

	for _, id := range cleanIDs {
		s, scopeErr := a.requireAnyScope(ctx, PermPartiesRead, false, partyScopes(id)...)
		if scopeErr != nil {
			return nil, scopeErr
		}
		sess = s
	}

	resp, err := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.GetPartiesByIdsResponse, error) {
		return c.Parties.GetPartiesByIds(rpcCtx, &efv1.GetPartiesByIdsRequest{Ids: cleanIDs})
	})
	if err != nil {
		return nil, err
	}
	return toModelParties(resp.GetItems()), nil
}

// PartyByCode mencari satu mitra eksternal berdasarkan partner_code.
func (a *Adapter) PartyByCode(ctx context.Context, partnerCode string) (*model.Party, error) {
	sess, err := a.requirePerm(ctx, PermPartiesRead, false)
	if err != nil {
		return nil, err
	}
	trimmed := strings.TrimSpace(partnerCode)
	if trimmed == "" {
		return nil, middleware.ErrInvalidArgument("partner_code is required", nil)
	}
	resp, err := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.GetPartyByCodeResponse, error) {
		return c.Parties.GetPartyByCode(rpcCtx, &efv1.GetPartyByCodeRequest{PartnerCode: trimmed})
	})
	if err != nil {
		return nil, err
	}
	return toModelParty(resp.GetParty()), nil
}

// PartyByTaxIdNumber mencari mitra eksternal berdasarkan NPWP.
func (a *Adapter) PartyByTaxIdNumber(ctx context.Context, taxIDNumber string) (*model.Party, error) {
	sess, err := a.requirePerm(ctx, PermPartiesRead, false)
	if err != nil {
		return nil, err
	}
	trimmed := strings.TrimSpace(taxIDNumber)
	if trimmed == "" {
		return nil, middleware.ErrInvalidArgument("tax_id_number is required", nil)
	}
	resp, err := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.GetPartyByTaxIdNumberResponse, error) {
		return c.Parties.GetPartyByTaxIdNumber(rpcCtx, &efv1.GetPartyByTaxIdNumberRequest{TaxIdNumber: trimmed})
	})
	if err != nil {
		return nil, err
	}
	return toModelParty(resp.GetParty()), nil
}

// PartiesByEmail mencari daftar mitra eksternal berdasarkan email secara terpaginasi.
func (a *Adapter) PartiesByEmail(ctx context.Context, email string, after, before *string, first, last *int) (*model.PartyConnection, error) {
	sess, err := a.requirePerm(ctx, PermPartiesRead, false)
	if err != nil {
		return nil, err
	}
	trimmed := strings.TrimSpace(email)
	if trimmed == "" {
		return nil, middleware.ErrInvalidArgument("email is required", nil)
	}
	return buildPartyConn(after, before, first, last, func(limit, offset int32) ([]*efv1.Party, *efv1.PageInfo, error) {
		resp, rpcErr := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.ListPartiesByEmailResponse, error) {
			return c.Parties.ListPartiesByEmail(rpcCtx, &efv1.ListPartiesByEmailRequest{
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

// Parties mengambil daftar mitra terpaginasi beserta filter teks, jenis, status, dan peran bisnis.
func (a *Adapter) Parties(
	ctx context.Context,
	after, before *string,
	first, last *int,
	query *string,
	kind *model.PartyKind,
	status *model.PartyStatus,
	role *model.PartyRoleKind,
	includeDeleted *bool,
) (*model.PartyConnection, error) {
	sess, err := a.requirePerm(ctx, PermPartiesRead, false)
	if err != nil {
		return nil, err
	}
	return buildPartyConn(after, before, first, last, func(limit, offset int32) ([]*efv1.Party, *efv1.PageInfo, error) {
		resp, rpcErr := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.ListPartiesResponse, error) {
			return c.Parties.ListParties(rpcCtx, &efv1.ListPartiesRequest{
				Query:          strings.TrimSpace(derefStr(query)),
				Kind:           mapOptionalEnumPtr(kind, partyKindToProto),
				Status:         mapOptionalEnumPtr(status, partyStatusToProto),
				Role:           mapOptionalEnumPtr(role, partyRoleKindToProto),
				IncludeDeleted: derefBool(includeDeleted, false),
				Page:           &efv1.PageRequest{Limit: limit, Offset: offset},
			})
		})
		if rpcErr != nil {
			return nil, nil, rpcErr
		}
		return resp.GetItems(), resp.GetPage(), nil
	})
}

// AllParties mengambil seluruh daftar mitra terpaginasi.
func (a *Adapter) AllParties(ctx context.Context, after, before *string, first, last *int) (*model.PartyConnection, error) {
	sess, err := a.requirePerm(ctx, PermPartiesRead, false)
	if err != nil {
		return nil, err
	}
	return buildPartyConn(after, before, first, last, func(limit, offset int32) ([]*efv1.Party, *efv1.PageInfo, error) {
		resp, rpcErr := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.GetAllPartiesResponse, error) {
			return c.Parties.GetAllParties(rpcCtx, &efv1.GetAllPartiesRequest{
				Page: &efv1.PageRequest{Limit: limit, Offset: offset},
			})
		})
		if rpcErr != nil {
			return nil, nil, rpcErr
		}
		return resp.GetItems(), resp.GetPage(), nil
	})
}

// CreateParty membuat entitas mitra eksternal baru.
func (a *Adapter) CreateParty(ctx context.Context, input model.CreatePartyInput) *model.CreatePartyPayload {
	sess, err := a.requirePerm(ctx, PermPartiesWrite, false)
	if err != nil {
		return &model.CreatePartyPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	statusVal := efv1.PartyStatus_PARTY_STATUS_PENDING_VERIFICATION
	if input.Status != nil {
		statusVal = mapEnumValue(*input.Status, partyStatusToProto, statusVal)
	}
	resp, err := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.CreatePartyResponse, error) {
		return c.Parties.CreateParty(rpcCtx, &efv1.CreatePartyRequest{
			PartnerCode:        strings.TrimSpace(input.PartnerCode),
			Kind:               mapEnumValue(input.Kind, partyKindToProto, efv1.PartyKind_PARTY_KIND_UNSPECIFIED),
			DisplayName:        strings.TrimSpace(input.DisplayName),
			LegalName:          input.LegalName,
			Email:              input.Email,
			Phone:              input.Phone,
			WebsiteUrl:         input.WebsiteURL,
			TaxIdNumber:        input.TaxIDNumber,
			TaxRegisteredName:  input.TaxRegisteredName,
			IsPkp:              derefBool(input.IsPkp, false),
			IsTaxExempt:        derefBool(input.IsTaxExempt, false),
			TaxExemptionReason: input.TaxExemptionReason,
			TaxVerifiedAt:      toProtoTimePtr(input.TaxVerifiedAt),
			PreferredLanguage:  normalizeStringDefault(input.PreferredLanguage, "id"),
			Status:             statusVal,
			Notes:              input.Notes,
		})
	})
	if err != nil {
		return &model.CreatePartyPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	return &model.CreatePartyPayload{
		Party:         toModelParty(resp.GetParty()),
		Errors:        NoMutationErrors(),
		WasSuccessful: true,
	}
}

// UpdatePartyProfile memperbarui profil umum mitra eksternal.
func (a *Adapter) UpdatePartyProfile(ctx context.Context, input model.UpdatePartyProfileInput) *model.UpdatePartyProfilePayload {
	cleanID, idErr := ParseTypedID(input.ID, ResourceParty)
	if idErr != nil {
		if _, authErr := a.requirePerm(ctx, PermPartiesWrite, false); authErr != nil {
			return &model.UpdatePartyProfilePayload{Errors: ToMutationErrors(authErr), WasSuccessful: false}
		}
		return &model.UpdatePartyProfilePayload{Errors: ToMutationErrors(idErr), WasSuccessful: false}
	}
	sess, err := a.requireAnyScope(ctx, PermPartiesWrite, false, partyScopes(cleanID)...)
	if err != nil {
		return &model.UpdatePartyProfilePayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	resp, err := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.UpdatePartyProfileResponse, error) {
		return c.Parties.UpdatePartyProfile(rpcCtx, &efv1.UpdatePartyProfileRequest{
			Id:                cleanID,
			Kind:              mapEnumValue(input.Kind, partyKindToProto, efv1.PartyKind_PARTY_KIND_UNSPECIFIED),
			DisplayName:       strings.TrimSpace(input.DisplayName),
			LegalName:         input.LegalName,
			Email:             input.Email,
			Phone:             input.Phone,
			WebsiteUrl:        input.WebsiteURL,
			PreferredLanguage: strings.TrimSpace(input.PreferredLanguage),
			Notes:             input.Notes,
		})
	})
	if err != nil {
		return &model.UpdatePartyProfilePayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	return &model.UpdatePartyProfilePayload{
		Party:         toModelParty(resp.GetParty()),
		Errors:        NoMutationErrors(),
		WasSuccessful: true,
	}
}

// UpdatePartyTaxProfile memperbarui profil perpajakan mitra eksternal.
func (a *Adapter) UpdatePartyTaxProfile(ctx context.Context, input model.UpdatePartyTaxProfileInput) *model.UpdatePartyTaxProfilePayload {
	cleanID, idErr := ParseTypedID(input.ID, ResourceParty)
	if idErr != nil {
		if _, authErr := a.requirePerm(ctx, PermPartiesWrite, false); authErr != nil {
			return &model.UpdatePartyTaxProfilePayload{Errors: ToMutationErrors(authErr), WasSuccessful: false}
		}
		return &model.UpdatePartyTaxProfilePayload{Errors: ToMutationErrors(idErr), WasSuccessful: false}
	}
	sess, err := a.requireAnyScope(ctx, PermPartiesWrite, false, partyScopes(cleanID)...)
	if err != nil {
		return &model.UpdatePartyTaxProfilePayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	resp, err := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.UpdatePartyTaxProfileResponse, error) {
		return c.Parties.UpdatePartyTaxProfile(rpcCtx, &efv1.UpdatePartyTaxProfileRequest{
			Id:                 cleanID,
			TaxIdNumber:        input.TaxIDNumber,
			TaxRegisteredName:  input.TaxRegisteredName,
			IsPkp:              derefBool(input.IsPkp, false),
			IsTaxExempt:        derefBool(input.IsTaxExempt, false),
			TaxExemptionReason: input.TaxExemptionReason,
			TaxVerifiedAt:      toProtoTimePtr(input.TaxVerifiedAt),
		})
	})
	if err != nil {
		return &model.UpdatePartyTaxProfilePayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	return &model.UpdatePartyTaxProfilePayload{
		Party:         toModelParty(resp.GetParty()),
		Errors:        NoMutationErrors(),
		WasSuccessful: true,
	}
}

// VerifyAndActivateParty memverifikasi sekaligus mengaktifkan mitra eksternal.
func (a *Adapter) VerifyAndActivateParty(ctx context.Context, id string) *model.VerifyAndActivatePartyPayload {
	cleanID, idErr := ParseTypedID(id, ResourceParty)
	if idErr != nil {
		if _, authErr := a.requirePerm(ctx, PermPartiesWrite, false); authErr != nil {
			return &model.VerifyAndActivatePartyPayload{Errors: ToMutationErrors(authErr), WasSuccessful: false}
		}
		return &model.VerifyAndActivatePartyPayload{Errors: ToMutationErrors(idErr), WasSuccessful: false}
	}
	sess, err := a.requireAnyScope(ctx, PermPartiesWrite, false, partyScopes(cleanID)...)
	if err != nil {
		return &model.VerifyAndActivatePartyPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	resp, err := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.VerifyAndActivatePartyResponse, error) {
		return c.Parties.VerifyAndActivateParty(rpcCtx, &efv1.VerifyAndActivatePartyRequest{Id: cleanID})
	})
	if err != nil {
		return &model.VerifyAndActivatePartyPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	return &model.VerifyAndActivatePartyPayload{
		Party:         toModelParty(resp.GetParty()),
		Errors:        NoMutationErrors(),
		WasSuccessful: true,
	}
}

// SuspendParty membekukan sementara aktivitas mitra eksternal.
func (a *Adapter) SuspendParty(ctx context.Context, id string) *model.SuspendPartyPayload {
	cleanID, idErr := ParseTypedID(id, ResourceParty)
	if idErr != nil {
		if _, authErr := a.requirePerm(ctx, PermPartiesWrite, false); authErr != nil {
			return &model.SuspendPartyPayload{Errors: ToMutationErrors(authErr), WasSuccessful: false}
		}
		return &model.SuspendPartyPayload{Errors: ToMutationErrors(idErr), WasSuccessful: false}
	}
	sess, err := a.requireAnyScope(ctx, PermPartiesWrite, false, partyScopes(cleanID)...)
	if err != nil {
		return &model.SuspendPartyPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	resp, err := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.SuspendPartyResponse, error) {
		return c.Parties.SuspendParty(rpcCtx, &efv1.SuspendPartyRequest{Id: cleanID})
	})
	if err != nil {
		return &model.SuspendPartyPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	return &model.SuspendPartyPayload{
		Party:         toModelParty(resp.GetParty()),
		Errors:        NoMutationErrors(),
		WasSuccessful: true,
	}
}

// DeactivateParty menonaktifkan mitra eksternal.
func (a *Adapter) DeactivateParty(ctx context.Context, id string) *model.DeactivatePartyPayload {
	cleanID, idErr := ParseTypedID(id, ResourceParty)
	if idErr != nil {
		if _, authErr := a.requirePerm(ctx, PermPartiesWrite, false); authErr != nil {
			return &model.DeactivatePartyPayload{Errors: ToMutationErrors(authErr), WasSuccessful: false}
		}
		return &model.DeactivatePartyPayload{Errors: ToMutationErrors(idErr), WasSuccessful: false}
	}
	sess, err := a.requireAnyScope(ctx, PermPartiesWrite, false, partyScopes(cleanID)...)
	if err != nil {
		return &model.DeactivatePartyPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	resp, err := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.DeactivatePartyResponse, error) {
		return c.Parties.DeactivateParty(rpcCtx, &efv1.DeactivatePartyRequest{Id: cleanID})
	})
	if err != nil {
		return &model.DeactivatePartyPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	return &model.DeactivatePartyPayload{
		Party:         toModelParty(resp.GetParty()),
		Errors:        NoMutationErrors(),
		WasSuccessful: true,
	}
}

// DeleteParty melakukan soft delete entitas mitra dengan kewajiban re-autentikasi sensitif.
func (a *Adapter) DeleteParty(ctx context.Context, id string) *model.DeletePartyPayload {
	cleanID, idErr := ParseTypedID(id, ResourceParty)
	if idErr != nil {
		if _, authErr := a.requirePerm(ctx, PermPartiesWrite, true); authErr != nil {
			return &model.DeletePartyPayload{Errors: ToMutationErrors(authErr), WasSuccessful: false}
		}
		return &model.DeletePartyPayload{Errors: ToMutationErrors(idErr), WasSuccessful: false}
	}
	sess, err := a.requireAnyScope(ctx, PermPartiesWrite, true, partyScopes(cleanID)...)
	if err != nil {
		return &model.DeletePartyPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	resp, err := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.DeletePartyResponse, error) {
		return c.Parties.DeleteParty(rpcCtx, &efv1.DeletePartyRequest{Id: cleanID})
	})
	if err != nil {
		return &model.DeletePartyPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	deleted := resp.GetDeleted()
	if !deleted {
		return &model.DeletePartyPayload{
			Errors:        ToMutationErrors(middleware.ErrConflict("party could not be deleted", nil)),
			WasSuccessful: false,
		}
	}
	return &model.DeletePartyPayload{
		Deleted:       &deleted,
		Errors:        NoMutationErrors(),
		WasSuccessful: true,
	}
}
