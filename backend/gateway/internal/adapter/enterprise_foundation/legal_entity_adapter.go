package enterprise_foundation

import (
	"context"
	"strings"
	"time"

	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	"github.com/integrione/erp-retail/backend/gateway/internal/source_connections"
	efv1 "github.com/integrione/erp-retail/backend/services/enterprise_foundation/contract/gen/enterprise_foundation/v1"
	"github.com/integrione/erp-retail/backend/shared/authz"
)

func buildLegalEntityConn(
	after, before *string,
	first, last *int,
	fetch func(limit, offset int32) ([]*efv1.LegalEntity, *efv1.PageInfo, error),
) (*model.LegalEntityConnection, error) {
	return buildRelayConnection(
		after, before, first, last,
		fetch,
		toModelLegalEntity,
		func(node *model.LegalEntity, cursor string) *model.LegalEntityEdge {
			return &model.LegalEntityEdge{Node: node, Cursor: cursor}
		},
		func(edges []*model.LegalEntityEdge, nodes []*model.LegalEntity, pi *model.PageInfo, total int) *model.LegalEntityConnection {
			return &model.LegalEntityConnection{Edges: edges, Nodes: nodes, PageInfo: pi, TotalCount: total}
		},
	)
}

func legalEntityScope(entityID string) authz.ScopeGrant {
	if strings.TrimSpace(entityID) == "" {
		return authz.ScopeGrant{Type: authz.ScopeGlobal}
	}
	return authz.ScopeGrant{Type: scopeTypeLegalEntity, ID: entityID}
}

// LegalEntity mengambil satu badan hukum berdasarkan ID/GID dengan verifikasi scope legal_entity.
func (a *Adapter) LegalEntity(ctx context.Context, id string) (*model.LegalEntity, error) {
	cleanID, err := ParseTypedID(id, ResourceLegalEntity)
	if err != nil {
		if _, authErr := a.requirePerm(ctx, PermLegalEntitiesRead, false); authErr != nil {
			return nil, authErr
		}
		return nil, err
	}
	sess, err := a.requireAnyScope(ctx, PermLegalEntitiesRead, false, legalEntityScope(cleanID))
	if err != nil {
		return nil, err
	}
	resp, err := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.GetLegalEntityByIdResponse, error) {
		return c.LegalEntities.GetLegalEntityById(rpcCtx, &efv1.GetLegalEntityByIdRequest{Id: cleanID})
	})
	if err != nil {
		return nil, err
	}
	return toModelLegalEntity(resp.GetLegalEntity()), nil
}

// LegalEntitiesByIds mengambil daftar badan hukum secara batch dan memvalidasi scope setiap ID target.
func (a *Adapter) LegalEntitiesByIds(ctx context.Context, ids []string) ([]*model.LegalEntity, error) {
	sess, err := a.authorizer.RequirePermission(ctx, PermLegalEntitiesRead, authz.ScopeGrant{Type: authz.ScopeGlobal}, false)
	hasGlobal := err == nil
	if err != nil {
		gwErr, ok := middleware.AsGatewayError(err)
		if !ok || gwErr.Code == middleware.CodeUnauthenticated || gwErr.Code != middleware.CodeForbidden {
			return nil, err
		}
	}

	cleanIDs, parseErr := ParseTypedIDList(ids, ResourceLegalEntity)
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
		return []*model.LegalEntity{}, nil
	}

	for _, id := range cleanIDs {
		s, scopeErr := a.requireAnyScope(ctx, PermLegalEntitiesRead, false, legalEntityScope(id))
		if scopeErr != nil {
			return nil, scopeErr
		}
		sess = s
	}

	resp, err := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.GetLegalEntitiesByIdsResponse, error) {
		return c.LegalEntities.GetLegalEntitiesByIds(rpcCtx, &efv1.GetLegalEntitiesByIdsRequest{Ids: cleanIDs})
	})
	if err != nil {
		return nil, err
	}
	return toModelLegalEntities(resp.GetItems()), nil
}

// LegalEntityByCode mengambil badan hukum berdasarkan entity_code unik.
func (a *Adapter) LegalEntityByCode(ctx context.Context, entityCode string) (*model.LegalEntity, error) {
	sess, err := a.requirePerm(ctx, PermLegalEntitiesRead, false)
	if err != nil {
		return nil, err
	}
	trimmed := strings.TrimSpace(entityCode)
	if trimmed == "" {
		return nil, middleware.ErrInvalidArgument("entity_code is required", nil)
	}
	resp, err := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.GetLegalEntityByCodeResponse, error) {
		return c.LegalEntities.GetLegalEntityByCode(rpcCtx, &efv1.GetLegalEntityByCodeRequest{EntityCode: trimmed})
	})
	if err != nil {
		return nil, err
	}
	return toModelLegalEntity(resp.GetLegalEntity()), nil
}

// LegalEntityByTaxIdNumber mengambil badan hukum berdasarkan nomor NPWP/pajak.
func (a *Adapter) LegalEntityByTaxIdNumber(ctx context.Context, taxIDNumber string) (*model.LegalEntity, error) {
	sess, err := a.requirePerm(ctx, PermLegalEntitiesRead, false)
	if err != nil {
		return nil, err
	}
	trimmed := strings.TrimSpace(taxIDNumber)
	if trimmed == "" {
		return nil, middleware.ErrInvalidArgument("tax_id_number is required", nil)
	}
	resp, err := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.GetLegalEntityByTaxIdNumberResponse, error) {
		return c.LegalEntities.GetLegalEntityByTaxIdNumber(rpcCtx, &efv1.GetLegalEntityByTaxIdNumberRequest{TaxIdNumber: trimmed})
	})
	if err != nil {
		return nil, err
	}
	return toModelLegalEntity(resp.GetLegalEntity()), nil
}

// LegalEntitiesByNib mengambil daftar badan hukum berdasarkan NIB secara terpaginasi.
func (a *Adapter) LegalEntitiesByNib(ctx context.Context, nib string, after, before *string, first, last *int) (*model.LegalEntityConnection, error) {
	sess, err := a.requirePerm(ctx, PermLegalEntitiesRead, false)
	if err != nil {
		return nil, err
	}
	trimmed := strings.TrimSpace(nib)
	if trimmed == "" {
		return nil, middleware.ErrInvalidArgument("nib is required", nil)
	}
	return buildLegalEntityConn(after, before, first, last, func(limit, offset int32) ([]*efv1.LegalEntity, *efv1.PageInfo, error) {
		resp, rpcErr := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.ListLegalEntitiesByNibResponse, error) {
			return c.LegalEntities.ListLegalEntitiesByNib(rpcCtx, &efv1.ListLegalEntitiesByNibRequest{
				Nib:  trimmed,
				Page: &efv1.PageRequest{Limit: limit, Offset: offset},
			})
		})
		if rpcErr != nil {
			return nil, nil, rpcErr
		}
		return resp.GetItems(), resp.GetPage(), nil
	})
}

// PrimaryOperatingLegalEntity mengambil badan hukum operasi utama perusahaan.
func (a *Adapter) PrimaryOperatingLegalEntity(ctx context.Context) (*model.LegalEntity, error) {
	sess, err := a.requirePerm(ctx, PermLegalEntitiesRead, false)
	if err != nil {
		return nil, err
	}
	resp, err := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.GetPrimaryOperatingLegalEntityResponse, error) {
		return c.LegalEntities.GetPrimaryOperatingLegalEntity(rpcCtx, &efv1.GetPrimaryOperatingLegalEntityRequest{})
	})
	if err != nil {
		return nil, err
	}
	return toModelLegalEntity(resp.GetLegalEntity()), nil
}

// LegalEntities mengambil daftar badan hukum terpaginasi beserta filter hierarki dan status.
func (a *Adapter) LegalEntities(
	ctx context.Context,
	after, before *string,
	first, last *int,
	query *string,
	parentID *string,
	status *model.LegalEntityStatus,
	primaryOnly, includeDeleted *bool,
) (*model.LegalEntityConnection, error) {
	cleanParentID, parseErr := ParseOptionalTypedID(parentID, ResourceLegalEntity)
	var sess *middleware.ResolvedSession
	var authErr error
	if cleanParentID != nil && parseErr == nil {
		sess, authErr = a.requireAnyScope(ctx, PermLegalEntitiesRead, false, legalEntityScope(*cleanParentID))
	} else {
		sess, authErr = a.requirePerm(ctx, PermLegalEntitiesRead, false)
	}
	if authErr != nil {
		return nil, authErr
	}
	if parseErr != nil {
		return nil, parseErr
	}
	return buildLegalEntityConn(after, before, first, last, func(limit, offset int32) ([]*efv1.LegalEntity, *efv1.PageInfo, error) {
		resp, rpcErr := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.ListLegalEntitiesResponse, error) {
			return c.LegalEntities.ListLegalEntities(rpcCtx, &efv1.ListLegalEntitiesRequest{
				Query:               strings.TrimSpace(derefStr(query)),
				ParentLegalEntityId: cleanParentID,
				Status:              mapOptionalEnumPtr(status, legalEntityStatusToProto),
				PrimaryOnly:         primaryOnly,
				IncludeDeleted:      derefBool(includeDeleted, false),
				Page:                &efv1.PageRequest{Limit: limit, Offset: offset},
			})
		})
		if rpcErr != nil {
			return nil, nil, rpcErr
		}
		return resp.GetItems(), resp.GetPage(), nil
	})
}

// AllLegalEntities mengambil seluruh badan hukum terpaginasi.
func (a *Adapter) AllLegalEntities(ctx context.Context, after, before *string, first, last *int) (*model.LegalEntityConnection, error) {
	sess, err := a.requirePerm(ctx, PermLegalEntitiesRead, false)
	if err != nil {
		return nil, err
	}
	return buildLegalEntityConn(after, before, first, last, func(limit, offset int32) ([]*efv1.LegalEntity, *efv1.PageInfo, error) {
		resp, rpcErr := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.GetAllLegalEntitiesResponse, error) {
			return c.LegalEntities.GetAllLegalEntities(rpcCtx, &efv1.GetAllLegalEntitiesRequest{
				Page: &efv1.PageRequest{Limit: limit, Offset: offset},
			})
		})
		if rpcErr != nil {
			return nil, nil, rpcErr
		}
		return resp.GetItems(), resp.GetPage(), nil
	})
}

// LegalEntityChildren mengambil anak badan hukum di bawah parent tertentu.
func (a *Adapter) LegalEntityChildren(ctx context.Context, parentID string, after, before *string, first, last *int) (*model.LegalEntityConnection, error) {
	cleanParentID, parseErr := ParseTypedID(parentID, ResourceLegalEntity)
	if parseErr != nil {
		if _, authErr := a.requirePerm(ctx, PermLegalEntitiesRead, false); authErr != nil {
			return nil, authErr
		}
		return nil, parseErr
	}
	sess, err := a.requireAnyScope(ctx, PermLegalEntitiesRead, false, legalEntityScope(cleanParentID))
	if err != nil {
		return nil, err
	}
	return buildLegalEntityConn(after, before, first, last, func(limit, offset int32) ([]*efv1.LegalEntity, *efv1.PageInfo, error) {
		resp, rpcErr := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.ListLegalEntityChildrenResponse, error) {
			return c.LegalEntities.ListLegalEntityChildren(rpcCtx, &efv1.ListLegalEntityChildrenRequest{
				ParentLegalEntityId: cleanParentID,
				Page:                &efv1.PageRequest{Limit: limit, Offset: offset},
			})
		})
		if rpcErr != nil {
			return nil, nil, rpcErr
		}
		return resp.GetItems(), resp.GetPage(), nil
	})
}

// ValidateLegalEntityScopeTarget memvalidasi kelayakan badan hukum sebagai target scope IAM.
func (a *Adapter) ValidateLegalEntityScopeTarget(ctx context.Context, id string) (*model.ScopeTargetDecision, error) {
	cleanID, parseErr := ParseTypedID(id, ResourceLegalEntity)
	if parseErr != nil {
		if _, authErr := a.requirePerm(ctx, PermLegalEntitiesRead, false); authErr != nil {
			return nil, authErr
		}
		return nil, parseErr
	}
	sess, err := a.requireAnyScope(ctx, PermLegalEntitiesRead, false, legalEntityScope(cleanID))
	if err != nil {
		return nil, err
	}
	resp, err := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.ValidateLegalEntityScopeTargetResponse, error) {
		return c.LegalEntities.ValidateLegalEntityScopeTarget(rpcCtx, &efv1.ValidateLegalEntityScopeTargetRequest{Id: cleanID})
	})
	if err != nil {
		return nil, err
	}
	return toModelScopeTargetDecision(resp.GetDecision()), nil
}

// CreateLegalEntity membuat entitas badan hukum baru.
func (a *Adapter) CreateLegalEntity(ctx context.Context, input model.CreateLegalEntityInput) *model.CreateLegalEntityPayload {
	parentID, err := ParseOptionalTypedID(input.ParentLegalEntityID, ResourceLegalEntity)
	var sess *middleware.ResolvedSession
	var authErr error
	if parentID != nil && err == nil {
		sess, authErr = a.requireAnyScope(ctx, PermLegalEntitiesWrite, false, legalEntityScope(*parentID))
	} else {
		sess, authErr = a.requirePerm(ctx, PermLegalEntitiesWrite, false)
	}
	if authErr != nil {
		return &model.CreateLegalEntityPayload{Errors: ToMutationErrors(authErr), WasSuccessful: false}
	}
	if err != nil {
		return &model.CreateLegalEntityPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	regAddrID, err := ParseOptionalTypedID(input.RegisteredAddressID, ResourceAddress)
	if err != nil {
		return &model.CreateLegalEntityPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}

	statusVal := efv1.LegalEntityStatus_LEGAL_ENTITY_STATUS_PENDING
	if input.Status != nil {
		statusVal = mapEnumValue(*input.Status, legalEntityStatusToProto, statusVal)
	}

	resp, err := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.CreateLegalEntityResponse, error) {
		return c.LegalEntities.CreateLegalEntity(rpcCtx, &efv1.CreateLegalEntityRequest{
			ParentLegalEntityId:      parentID,
			EntityCode:               strings.TrimSpace(input.EntityCode),
			LegalName:                strings.TrimSpace(input.LegalName),
			TradingName:              input.TradingName,
			LegalForm:                input.LegalForm,
			IsPrimaryOperatingEntity: derefBool(input.IsPrimaryOperatingEntity, false),
			Status:                   statusVal,
			JurisdictionCountryCode:  strings.TrimSpace(input.JurisdictionCountryCode),
			Nib:                      input.Nib,
			TaxIdNumber:              input.TaxIDNumber,
			IsPkp:                    derefBool(input.IsPkp, false),
			RegisteredAddressId:      regAddrID,
			FiscalYearStartMonth:     derefInt32(input.FiscalYearStartMonth, 1),
			IncorporationDate:        toProtoTimePtr(input.IncorporationDate),
			DissolutionDate:          toProtoTimePtr(input.DissolutionDate),
		})
	})
	if err != nil {
		return &model.CreateLegalEntityPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	return &model.CreateLegalEntityPayload{
		LegalEntity:   toModelLegalEntity(resp.GetLegalEntity()),
		Errors:        NoMutationErrors(),
		WasSuccessful: true,
	}
}

// UpdateLegalEntity memperbarui profil, alamat terdaftar, maupun data pajak badan hukum.
func (a *Adapter) UpdateLegalEntity(ctx context.Context, input model.UpdateLegalEntityInput) *model.UpdateLegalEntityPayload {
	cleanID, idErr := ParseTypedID(input.ID, ResourceLegalEntity)
	if idErr != nil {
		if _, authErr := a.requirePerm(ctx, PermLegalEntitiesWrite, false); authErr != nil {
			return &model.UpdateLegalEntityPayload{Errors: ToMutationErrors(authErr), WasSuccessful: false}
		}
		return &model.UpdateLegalEntityPayload{Errors: ToMutationErrors(idErr), WasSuccessful: false}
	}
	sess, err := a.requireAnyScope(ctx, PermLegalEntitiesWrite, false, legalEntityScope(cleanID))
	if err != nil {
		return &model.UpdateLegalEntityPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	parentID, err := ParseOptionalTypedID(input.ParentLegalEntityID, ResourceLegalEntity)
	if err != nil {
		return &model.UpdateLegalEntityPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	regAddrID, err := ParseOptionalTypedID(input.RegisteredAddressID, ResourceAddress)
	if err != nil {
		return &model.UpdateLegalEntityPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}

	resp, err := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.UpdateLegalEntityResponse, error) {
		return c.LegalEntities.UpdateLegalEntity(rpcCtx, &efv1.UpdateLegalEntityRequest{
			Id:                      cleanID,
			ParentLegalEntityId:     parentID,
			LegalName:               strings.TrimSpace(input.LegalName),
			TradingName:             input.TradingName,
			LegalForm:               input.LegalForm,
			JurisdictionCountryCode: strings.TrimSpace(input.JurisdictionCountryCode),
			Nib:                     input.Nib,
			TaxIdNumber:             input.TaxIDNumber,
			IsPkp:                   derefBool(input.IsPkp, false),
			RegisteredAddressId:     regAddrID,
			FiscalYearStartMonth:    derefInt32(input.FiscalYearStartMonth, 1),
			IncorporationDate:       toProtoTimePtr(input.IncorporationDate),
			DissolutionDate:         toProtoTimePtr(input.DissolutionDate),
		})
	})
	if err != nil {
		return &model.UpdateLegalEntityPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	return &model.UpdateLegalEntityPayload{
		LegalEntity:   toModelLegalEntity(resp.GetLegalEntity()),
		Errors:        NoMutationErrors(),
		WasSuccessful: true,
	}
}

// SetPrimaryOperatingLegalEntity menetapkan satu badan hukum sebagai entitas operasi utama.
func (a *Adapter) SetPrimaryOperatingLegalEntity(ctx context.Context, id string) *model.SetPrimaryOperatingLegalEntityPayload {
	cleanID, idErr := ParseTypedID(id, ResourceLegalEntity)
	if idErr != nil {
		if _, authErr := a.requirePerm(ctx, PermLegalEntitiesWrite, true); authErr != nil {
			return &model.SetPrimaryOperatingLegalEntityPayload{Errors: ToMutationErrors(authErr), WasSuccessful: false}
		}
		return &model.SetPrimaryOperatingLegalEntityPayload{Errors: ToMutationErrors(idErr), WasSuccessful: false}
	}
	sess, err := a.requireAnyScope(ctx, PermLegalEntitiesWrite, true, legalEntityScope(cleanID))
	if err != nil {
		return &model.SetPrimaryOperatingLegalEntityPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	resp, err := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.SetPrimaryOperatingLegalEntityResponse, error) {
		return c.LegalEntities.SetPrimaryOperatingLegalEntity(rpcCtx, &efv1.SetPrimaryOperatingLegalEntityRequest{Id: cleanID})
	})
	if err != nil {
		return &model.SetPrimaryOperatingLegalEntityPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	return &model.SetPrimaryOperatingLegalEntityPayload{
		LegalEntity:   toModelLegalEntity(resp.GetLegalEntity()),
		Errors:        NoMutationErrors(),
		WasSuccessful: true,
	}
}

// ActivateLegalEntity mengaktifkan badan hukum.
func (a *Adapter) ActivateLegalEntity(ctx context.Context, id string) *model.ActivateLegalEntityPayload {
	cleanID, idErr := ParseTypedID(id, ResourceLegalEntity)
	if idErr != nil {
		if _, authErr := a.requirePerm(ctx, PermLegalEntitiesWrite, false); authErr != nil {
			return &model.ActivateLegalEntityPayload{Errors: ToMutationErrors(authErr), WasSuccessful: false}
		}
		return &model.ActivateLegalEntityPayload{Errors: ToMutationErrors(idErr), WasSuccessful: false}
	}
	sess, err := a.requireAnyScope(ctx, PermLegalEntitiesWrite, false, legalEntityScope(cleanID))
	if err != nil {
		return &model.ActivateLegalEntityPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	resp, err := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.ActivateLegalEntityResponse, error) {
		return c.LegalEntities.ActivateLegalEntity(rpcCtx, &efv1.ActivateLegalEntityRequest{Id: cleanID})
	})
	if err != nil {
		return &model.ActivateLegalEntityPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	return &model.ActivateLegalEntityPayload{
		LegalEntity:   toModelLegalEntity(resp.GetLegalEntity()),
		Errors:        NoMutationErrors(),
		WasSuccessful: true,
	}
}

// SuspendLegalEntity membekukan sementara badan hukum.
func (a *Adapter) SuspendLegalEntity(ctx context.Context, id string) *model.SuspendLegalEntityPayload {
	cleanID, idErr := ParseTypedID(id, ResourceLegalEntity)
	if idErr != nil {
		if _, authErr := a.requirePerm(ctx, PermLegalEntitiesWrite, false); authErr != nil {
			return &model.SuspendLegalEntityPayload{Errors: ToMutationErrors(authErr), WasSuccessful: false}
		}
		return &model.SuspendLegalEntityPayload{Errors: ToMutationErrors(idErr), WasSuccessful: false}
	}
	sess, err := a.requireAnyScope(ctx, PermLegalEntitiesWrite, false, legalEntityScope(cleanID))
	if err != nil {
		return &model.SuspendLegalEntityPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	resp, err := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.SuspendLegalEntityResponse, error) {
		return c.LegalEntities.SuspendLegalEntity(rpcCtx, &efv1.SuspendLegalEntityRequest{Id: cleanID})
	})
	if err != nil {
		return &model.SuspendLegalEntityPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	return &model.SuspendLegalEntityPayload{
		LegalEntity:   toModelLegalEntity(resp.GetLegalEntity()),
		Errors:        NoMutationErrors(),
		WasSuccessful: true,
	}
}

// DissolveLegalEntity membubarkan badan hukum secara permanen beserta tanggal pembubaran.
func (a *Adapter) DissolveLegalEntity(ctx context.Context, id string, dissolutionDate *time.Time) *model.DissolveLegalEntityPayload {
	cleanID, idErr := ParseTypedID(id, ResourceLegalEntity)
	if idErr != nil {
		if _, authErr := a.requirePerm(ctx, PermLegalEntitiesWrite, true); authErr != nil {
			return &model.DissolveLegalEntityPayload{Errors: ToMutationErrors(authErr), WasSuccessful: false}
		}
		return &model.DissolveLegalEntityPayload{Errors: ToMutationErrors(idErr), WasSuccessful: false}
	}
	sess, err := a.requireAnyScope(ctx, PermLegalEntitiesWrite, true, legalEntityScope(cleanID))
	if err != nil {
		return &model.DissolveLegalEntityPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	resp, err := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.DissolveLegalEntityResponse, error) {
		return c.LegalEntities.DissolveLegalEntity(rpcCtx, &efv1.DissolveLegalEntityRequest{
			Id:              cleanID,
			DissolutionDate: toProtoTimePtr(dissolutionDate),
		})
	})
	if err != nil {
		return &model.DissolveLegalEntityPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	return &model.DissolveLegalEntityPayload{
		LegalEntity:   toModelLegalEntity(resp.GetLegalEntity()),
		Errors:        NoMutationErrors(),
		WasSuccessful: true,
	}
}

// DeleteLegalEntity melakukan soft delete badan hukum dengan syarat re-autentikasi sensitif.
func (a *Adapter) DeleteLegalEntity(ctx context.Context, id string) *model.DeleteLegalEntityPayload {
	cleanID, idErr := ParseTypedID(id, ResourceLegalEntity)
	if idErr != nil {
		if _, authErr := a.requirePerm(ctx, PermLegalEntitiesWrite, true); authErr != nil {
			return &model.DeleteLegalEntityPayload{Errors: ToMutationErrors(authErr), WasSuccessful: false}
		}
		return &model.DeleteLegalEntityPayload{Errors: ToMutationErrors(idErr), WasSuccessful: false}
	}
	sess, err := a.requireAnyScope(ctx, PermLegalEntitiesWrite, true, legalEntityScope(cleanID))
	if err != nil {
		return &model.DeleteLegalEntityPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	resp, err := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.DeleteLegalEntityResponse, error) {
		return c.LegalEntities.DeleteLegalEntity(rpcCtx, &efv1.DeleteLegalEntityRequest{Id: cleanID})
	})
	if err != nil {
		return &model.DeleteLegalEntityPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	deleted := resp.GetDeleted()
	if !deleted {
		return &model.DeleteLegalEntityPayload{
			Errors:        ToMutationErrors(middleware.ErrConflict("legal entity could not be deleted", nil)),
			WasSuccessful: false,
		}
	}
	return &model.DeleteLegalEntityPayload{
		Deleted:       &deleted,
		Errors:        NoMutationErrors(),
		WasSuccessful: true,
	}
}

// ResolveLegalEntityParent mengambil entitas induk milik LegalEntity secara aman.
func (a *Adapter) ResolveLegalEntityParent(ctx context.Context, obj *model.LegalEntity) (*model.LegalEntity, error) {
	if obj == nil || obj.ParentLegalEntityID == nil || strings.TrimSpace(*obj.ParentLegalEntityID) == "" {
		return nil, nil
	}
	return a.LegalEntity(ctx, *obj.ParentLegalEntityID)
}

// ResolveLegalEntityRegisteredAddress mengambil alamat terdaftar milik LegalEntity secara aman.
func (a *Adapter) ResolveLegalEntityRegisteredAddress(ctx context.Context, obj *model.LegalEntity) (*model.Address, error) {
	if obj == nil || obj.RegisteredAddressID == nil || strings.TrimSpace(*obj.RegisteredAddressID) == "" {
		return nil, nil
	}
	return a.Address(ctx, *obj.RegisteredAddressID)
}
