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

func buildOrgUnitConn(
	after, before *string,
	first, last *int,
	fetch func(limit, offset int32) ([]*efv1.OrganizationUnit, *efv1.PageInfo, error),
) (*model.OrganizationUnitConnection, error) {
	return buildRelayConnection(
		after, before, first, last,
		fetch,
		toModelOrganizationUnit,
		func(node *model.OrganizationUnit, cursor string) *model.OrganizationUnitEdge {
			return &model.OrganizationUnitEdge{Node: node, Cursor: cursor}
		},
		func(edges []*model.OrganizationUnitEdge, nodes []*model.OrganizationUnit, pi *model.PageInfo, total int) *model.OrganizationUnitConnection {
			return &model.OrganizationUnitConnection{Edges: edges, Nodes: nodes, PageInfo: pi, TotalCount: total}
		},
	)
}

func orgUnitScope(unitID string) authz.ScopeGrant {
	if strings.TrimSpace(unitID) == "" {
		return authz.ScopeGrant{Type: authz.ScopeGlobal}
	}
	return authz.ScopeGrant{Type: scopeTypeOrganizationUnit, ID: unitID}
}

// OrganizationUnit mengambil satu unit organisasi berdasarkan ID/GID.
func (a *Adapter) OrganizationUnit(ctx context.Context, id string) (*model.OrganizationUnit, error) {
	cleanID, err := ParseTypedID(id, ResourceOrganizationUnit)
	if err != nil {
		if _, authErr := a.requirePerm(ctx, PermOrganizationUnitsRead, false); authErr != nil {
			return nil, authErr
		}
		return nil, err
	}
	sess, err := a.requireAnyScope(ctx, PermOrganizationUnitsRead, false, orgUnitScope(cleanID))
	if err != nil {
		return nil, err
	}
	resp, err := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.GetOrganizationUnitByIdResponse, error) {
		return c.OrganizationUnits.GetOrganizationUnitById(rpcCtx, &efv1.GetOrganizationUnitByIdRequest{Id: cleanID})
	})
	if err != nil {
		return nil, err
	}
	return toModelOrganizationUnit(resp.GetOrganizationUnit()), nil
}

// OrganizationUnitsByIds mengambil daftar unit organisasi secara batch dengan verifikasi scope per unit.
func (a *Adapter) OrganizationUnitsByIds(ctx context.Context, ids []string) ([]*model.OrganizationUnit, error) {
	sess, err := a.authorizer.RequirePermission(ctx, PermOrganizationUnitsRead, authz.ScopeGrant{Type: authz.ScopeGlobal}, false)
	hasGlobal := err == nil
	if err != nil {
		gwErr, ok := middleware.AsGatewayError(err)
		if !ok || gwErr.Code != middleware.CodeForbidden {
			return nil, err
		}
	}

	cleanIDs, parseErr := ParseTypedIDList(ids, ResourceOrganizationUnit)
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
		return []*model.OrganizationUnit{}, nil
	}

	for _, id := range cleanIDs {
		s, scopeErr := a.requireAnyScope(ctx, PermOrganizationUnitsRead, false, orgUnitScope(id))
		if scopeErr != nil {
			return nil, scopeErr
		}
		sess = s
	}

	resp, err := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.GetOrganizationUnitsByIdsResponse, error) {
		return c.OrganizationUnits.GetOrganizationUnitsByIds(rpcCtx, &efv1.GetOrganizationUnitsByIdsRequest{Ids: cleanIDs})
	})
	if err != nil {
		return nil, err
	}
	return toModelOrganizationUnits(resp.GetItems()), nil
}

// OrganizationUnitByEntityAndCode mencari unit organisasi berdasarkan badan hukum dan unit_code.
func (a *Adapter) OrganizationUnitByEntityAndCode(ctx context.Context, legalEntityID, unitCode string) (*model.OrganizationUnit, error) {
	cleanEntityID, parseErr := ParseTypedID(legalEntityID, ResourceLegalEntity)
	if parseErr != nil {
		if _, authErr := a.requirePerm(ctx, PermOrganizationUnitsRead, false); authErr != nil {
			return nil, authErr
		}
		return nil, parseErr
	}
	sess, err := a.requireAnyScope(ctx, PermOrganizationUnitsRead, false, legalEntityScope(cleanEntityID))
	if err != nil {
		return nil, err
	}
	trimmedCode := strings.TrimSpace(unitCode)
	if trimmedCode == "" {
		return nil, middleware.ErrInvalidArgument("unit_code is required", nil)
	}
	resp, err := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.GetOrganizationUnitByEntityAndCodeResponse, error) {
		return c.OrganizationUnits.GetOrganizationUnitByEntityAndCode(rpcCtx, &efv1.GetOrganizationUnitByEntityAndCodeRequest{
			LegalEntityId: cleanEntityID,
			UnitCode:      trimmedCode,
		})
	})
	if err != nil {
		return nil, err
	}
	return toModelOrganizationUnit(resp.GetOrganizationUnit()), nil
}

// OrganizationUnitsByHeadPersonId mengambil unit organisasi yang dipimpin personil tertentu secara terpaginasi.
func (a *Adapter) OrganizationUnitsByHeadPersonId(
	ctx context.Context,
	headPersonID string,
	after, before *string,
	first, last *int,
) (*model.OrganizationUnitConnection, error) {
	sess, err := a.requirePerm(ctx, PermOrganizationUnitsRead, false)
	if err != nil {
		return nil, err
	}
	cleanHeadID, err := ParseTypedID(headPersonID, ResourceWorker)
	if err != nil {
		return nil, err
	}
	return buildOrgUnitConn(after, before, first, last, func(limit, offset int32) ([]*efv1.OrganizationUnit, *efv1.PageInfo, error) {
		resp, rpcErr := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.ListOrganizationUnitsByHeadPersonIdResponse, error) {
			return c.OrganizationUnits.ListOrganizationUnitsByHeadPersonId(rpcCtx, &efv1.ListOrganizationUnitsByHeadPersonIdRequest{
				HeadPersonId: cleanHeadID,
				Page:         &efv1.PageRequest{Limit: limit, Offset: offset},
			})
		})
		if rpcErr != nil {
			return nil, nil, rpcErr
		}
		return resp.GetItems(), resp.GetPage(), nil
	})
}

// OrganizationUnitsByPrimaryLocationId mengambil daftar unit organisasi berdasarkan lokasi utamanya.
func (a *Adapter) OrganizationUnitsByPrimaryLocationId(
	ctx context.Context,
	primaryLocationID string,
	after, before *string,
	first, last *int,
) (*model.OrganizationUnitConnection, error) {
	cleanLocID, parseErr := ParseTypedID(primaryLocationID, ResourceLocation)
	if parseErr != nil {
		if _, authErr := a.requirePerm(ctx, PermOrganizationUnitsRead, false); authErr != nil {
			return nil, authErr
		}
		return nil, parseErr
	}
	sess, err := a.requireAnyScope(ctx, PermOrganizationUnitsRead, false, facilityScope(cleanLocID))
	if err != nil {
		return nil, err
	}
	return buildOrgUnitConn(after, before, first, last, func(limit, offset int32) ([]*efv1.OrganizationUnit, *efv1.PageInfo, error) {
		resp, rpcErr := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.ListOrganizationUnitsByPrimaryLocationIdResponse, error) {
			return c.OrganizationUnits.ListOrganizationUnitsByPrimaryLocationId(rpcCtx, &efv1.ListOrganizationUnitsByPrimaryLocationIdRequest{
				PrimaryLocationId: cleanLocID,
				Page:              &efv1.PageRequest{Limit: limit, Offset: offset},
			})
		})
		if rpcErr != nil {
			return nil, nil, rpcErr
		}
		return resp.GetItems(), resp.GetPage(), nil
	})
}

// OrganizationUnitChildren mengambil sub-unit organisasi di bawah parent tertentu.
func (a *Adapter) OrganizationUnitChildren(
	ctx context.Context,
	parentUnitID string,
	after, before *string,
	first, last *int,
) (*model.OrganizationUnitConnection, error) {
	cleanParentID, parseErr := ParseTypedID(parentUnitID, ResourceOrganizationUnit)
	if parseErr != nil {
		if _, authErr := a.requirePerm(ctx, PermOrganizationUnitsRead, false); authErr != nil {
			return nil, authErr
		}
		return nil, parseErr
	}
	sess, err := a.requireAnyScope(ctx, PermOrganizationUnitsRead, false, orgUnitScope(cleanParentID))
	if err != nil {
		return nil, err
	}
	return buildOrgUnitConn(after, before, first, last, func(limit, offset int32) ([]*efv1.OrganizationUnit, *efv1.PageInfo, error) {
		resp, rpcErr := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.ListOrganizationUnitChildrenResponse, error) {
			return c.OrganizationUnits.ListOrganizationUnitChildren(rpcCtx, &efv1.ListOrganizationUnitChildrenRequest{
				ParentUnitId: cleanParentID,
				Page:         &efv1.PageRequest{Limit: limit, Offset: offset},
			})
		})
		if rpcErr != nil {
			return nil, nil, rpcErr
		}
		return resp.GetItems(), resp.GetPage(), nil
	})
}

// OrganizationUnits mengambil daftar unit organisasi terpaginasi dengan filter multi-kriteria.
func (a *Adapter) OrganizationUnits(
	ctx context.Context,
	after, before *string,
	first, last *int,
	query, legalEntityID, parentUnitID, primaryLocationID *string,
	kind *model.OrgUnitKind,
	activeOnly, includeDeleted *bool,
) (*model.OrganizationUnitConnection, error) {
	cleanEntityID, entityErr := ParseOptionalTypedID(legalEntityID, ResourceLegalEntity)
	var sess *middleware.ResolvedSession
	var authErr error
	if cleanEntityID != nil && entityErr == nil {
		sess, authErr = a.requireAnyScope(ctx, PermOrganizationUnitsRead, false, legalEntityScope(*cleanEntityID))
	} else {
		sess, authErr = a.requirePerm(ctx, PermOrganizationUnitsRead, false)
	}
	if authErr != nil {
		return nil, authErr
	}
	if entityErr != nil {
		return nil, entityErr
	}
	cleanParentID, err := ParseOptionalTypedID(parentUnitID, ResourceOrganizationUnit)
	if err != nil {
		return nil, err
	}
	cleanLocID, err := ParseOptionalTypedID(primaryLocationID, ResourceLocation)
	if err != nil {
		return nil, err
	}
	return buildOrgUnitConn(after, before, first, last, func(limit, offset int32) ([]*efv1.OrganizationUnit, *efv1.PageInfo, error) {
		resp, rpcErr := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.ListOrganizationUnitsResponse, error) {
			return c.OrganizationUnits.ListOrganizationUnits(rpcCtx, &efv1.ListOrganizationUnitsRequest{
				Query:             strings.TrimSpace(derefStr(query)),
				LegalEntityId:     derefStr(cleanEntityID),
				ParentUnitId:      cleanParentID,
				PrimaryLocationId: cleanLocID,
				Kind:              mapOptionalEnumPtr(kind, orgUnitKindToProto),
				ActiveOnly:        derefBool(activeOnly, false),
				IncludeDeleted:    derefBool(includeDeleted, false),
				Page:              &efv1.PageRequest{Limit: limit, Offset: offset},
			})
		})
		if rpcErr != nil {
			return nil, nil, rpcErr
		}
		return resp.GetItems(), resp.GetPage(), nil
	})
}

// AllOrganizationUnits mengambil seluruh unit organisasi terpaginasi.
func (a *Adapter) AllOrganizationUnits(ctx context.Context, after, before *string, first, last *int) (*model.OrganizationUnitConnection, error) {
	sess, err := a.requirePerm(ctx, PermOrganizationUnitsRead, false)
	if err != nil {
		return nil, err
	}
	return buildOrgUnitConn(after, before, first, last, func(limit, offset int32) ([]*efv1.OrganizationUnit, *efv1.PageInfo, error) {
		resp, rpcErr := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.GetAllOrganizationUnitsResponse, error) {
			return c.OrganizationUnits.GetAllOrganizationUnits(rpcCtx, &efv1.GetAllOrganizationUnitsRequest{
				Page: &efv1.PageRequest{Limit: limit, Offset: offset},
			})
		})
		if rpcErr != nil {
			return nil, nil, rpcErr
		}
		return resp.GetItems(), resp.GetPage(), nil
	})
}

// ValidateOrganizationUnitScopeTarget memvalidasi kelayakan unit organisasi sebagai target scope IAM.
func (a *Adapter) ValidateOrganizationUnitScopeTarget(ctx context.Context, id string) (*model.ScopeTargetDecision, error) {
	cleanID, parseErr := ParseTypedID(id, ResourceOrganizationUnit)
	if parseErr != nil {
		if _, authErr := a.requirePerm(ctx, PermOrganizationUnitsRead, false); authErr != nil {
			return nil, authErr
		}
		return nil, parseErr
	}
	sess, err := a.requireAnyScope(ctx, PermOrganizationUnitsRead, false, orgUnitScope(cleanID))
	if err != nil {
		return nil, err
	}
	resp, err := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.ValidateOrganizationUnitScopeTargetResponse, error) {
		return c.OrganizationUnits.ValidateOrganizationUnitScopeTarget(rpcCtx, &efv1.ValidateOrganizationUnitScopeTargetRequest{Id: cleanID})
	})
	if err != nil {
		return nil, err
	}
	return toModelScopeTargetDecision(resp.GetDecision()), nil
}

// CreateOrganizationUnit membuat unit organisasi internal baru.
func (a *Adapter) CreateOrganizationUnit(ctx context.Context, input model.CreateOrganizationUnitInput) *model.CreateOrganizationUnitPayload {
	cleanEntityID, idErr := ParseTypedID(input.LegalEntityID, ResourceLegalEntity)
	if idErr != nil {
		if _, authErr := a.requirePerm(ctx, PermOrganizationUnitsWrite, false); authErr != nil {
			return &model.CreateOrganizationUnitPayload{Errors: ToMutationErrors(authErr), WasSuccessful: false}
		}
		return &model.CreateOrganizationUnitPayload{Errors: ToMutationErrors(idErr), WasSuccessful: false}
	}
	sess, err := a.requireAnyScope(ctx, PermOrganizationUnitsWrite, false, legalEntityScope(cleanEntityID))
	if err != nil {
		return &model.CreateOrganizationUnitPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	parentID, err := ParseOptionalTypedID(input.ParentUnitID, ResourceOrganizationUnit)
	if err != nil {
		return &model.CreateOrganizationUnitPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	locID, err := ParseOptionalTypedID(input.PrimaryLocationID, ResourceLocation)
	if err != nil {
		return &model.CreateOrganizationUnitPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	headID, err := ParseOptionalTypedID(input.HeadPersonID, ResourceWorker)
	if err != nil {
		return &model.CreateOrganizationUnitPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}

	resp, err := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.CreateOrganizationUnitResponse, error) {
		return c.OrganizationUnits.CreateOrganizationUnit(rpcCtx, &efv1.CreateOrganizationUnitRequest{
			LegalEntityId:     cleanEntityID,
			ParentUnitId:      parentID,
			PrimaryLocationId: locID,
			UnitCode:          strings.TrimSpace(input.UnitCode),
			Name:              strings.TrimSpace(input.Name),
			Kind:              mapEnumValue(input.Kind, orgUnitKindToProto, efv1.OrgUnitKind_ORG_UNIT_KIND_UNSPECIFIED),
			CostCenterCode:    input.CostCenterCode,
			HeadPersonId:      headID,
		})
	})
	if err != nil {
		return &model.CreateOrganizationUnitPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	return &model.CreateOrganizationUnitPayload{
		OrganizationUnit: toModelOrganizationUnit(resp.GetOrganizationUnit()),
		Errors:           NoMutationErrors(),
		WasSuccessful:    true,
	}
}

// UpdateOrganizationUnit memperbarui struktur maupun atribut unit organisasi.
func (a *Adapter) UpdateOrganizationUnit(ctx context.Context, input model.UpdateOrganizationUnitInput) *model.UpdateOrganizationUnitPayload {
	cleanID, idErr := ParseTypedID(input.ID, ResourceOrganizationUnit)
	if idErr != nil {
		if _, authErr := a.requirePerm(ctx, PermOrganizationUnitsWrite, false); authErr != nil {
			return &model.UpdateOrganizationUnitPayload{Errors: ToMutationErrors(authErr), WasSuccessful: false}
		}
		return &model.UpdateOrganizationUnitPayload{Errors: ToMutationErrors(idErr), WasSuccessful: false}
	}
	sess, err := a.requireAnyScope(ctx, PermOrganizationUnitsWrite, false, orgUnitScope(cleanID))
	if err != nil {
		return &model.UpdateOrganizationUnitPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	parentID, err := ParseOptionalTypedID(input.ParentUnitID, ResourceOrganizationUnit)
	if err != nil {
		return &model.UpdateOrganizationUnitPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	locID, err := ParseOptionalTypedID(input.PrimaryLocationID, ResourceLocation)
	if err != nil {
		return &model.UpdateOrganizationUnitPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	headID, err := ParseOptionalTypedID(input.HeadPersonID, ResourceWorker)
	if err != nil {
		return &model.UpdateOrganizationUnitPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}

	resp, err := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.UpdateOrganizationUnitResponse, error) {
		return c.OrganizationUnits.UpdateOrganizationUnit(rpcCtx, &efv1.UpdateOrganizationUnitRequest{
			Id:                 cleanID,
			ParentUnitId:       parentID,
			PrimaryLocationId:  locID,
			Name:               strings.TrimSpace(input.Name),
			Kind:               mapEnumValue(input.Kind, orgUnitKindToProto, efv1.OrgUnitKind_ORG_UNIT_KIND_UNSPECIFIED),
			CostCenterCode:     input.CostCenterCode,
			HeadPersonId:       headID,
			PreserveHeadPerson: input.PreserveHeadPerson,
		})
	})
	if err != nil {
		return &model.UpdateOrganizationUnitPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	return &model.UpdateOrganizationUnitPayload{
		OrganizationUnit: toModelOrganizationUnit(resp.GetOrganizationUnit()),
		Errors:           NoMutationErrors(),
		WasSuccessful:    true,
	}
}

// AssignOrganizationUnitHeadPerson menetapkan atau mengosongkan kepala unit organisasi.
func (a *Adapter) AssignOrganizationUnitHeadPerson(ctx context.Context, id string, headPersonID *string) *model.AssignOrganizationUnitHeadPersonPayload {
	cleanID, idErr := ParseTypedID(id, ResourceOrganizationUnit)
	if idErr != nil {
		if _, authErr := a.requirePerm(ctx, PermOrganizationUnitsWrite, false); authErr != nil {
			return &model.AssignOrganizationUnitHeadPersonPayload{Errors: ToMutationErrors(authErr), WasSuccessful: false}
		}
		return &model.AssignOrganizationUnitHeadPersonPayload{Errors: ToMutationErrors(idErr), WasSuccessful: false}
	}
	sess, err := a.requireAnyScope(ctx, PermOrganizationUnitsWrite, false, orgUnitScope(cleanID))
	if err != nil {
		return &model.AssignOrganizationUnitHeadPersonPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	headID, err := ParseOptionalTypedID(headPersonID, ResourceWorker)
	if err != nil {
		return &model.AssignOrganizationUnitHeadPersonPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}

	resp, err := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.AssignOrganizationUnitHeadPersonResponse, error) {
		return c.OrganizationUnits.AssignOrganizationUnitHeadPerson(rpcCtx, &efv1.AssignOrganizationUnitHeadPersonRequest{
			Id:           cleanID,
			HeadPersonId: headID,
		})
	})
	if err != nil {
		return &model.AssignOrganizationUnitHeadPersonPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	return &model.AssignOrganizationUnitHeadPersonPayload{
		OrganizationUnit: toModelOrganizationUnit(resp.GetOrganizationUnit()),
		Errors:           NoMutationErrors(),
		WasSuccessful:    true,
	}
}

// ActivateOrganizationUnit mengaktifkan unit organisasi.
func (a *Adapter) ActivateOrganizationUnit(ctx context.Context, id string) *model.ActivateOrganizationUnitPayload {
	cleanID, idErr := ParseTypedID(id, ResourceOrganizationUnit)
	if idErr != nil {
		if _, authErr := a.requirePerm(ctx, PermOrganizationUnitsWrite, false); authErr != nil {
			return &model.ActivateOrganizationUnitPayload{Errors: ToMutationErrors(authErr), WasSuccessful: false}
		}
		return &model.ActivateOrganizationUnitPayload{Errors: ToMutationErrors(idErr), WasSuccessful: false}
	}
	sess, err := a.requireAnyScope(ctx, PermOrganizationUnitsWrite, false, orgUnitScope(cleanID))
	if err != nil {
		return &model.ActivateOrganizationUnitPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	resp, err := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.ActivateOrganizationUnitResponse, error) {
		return c.OrganizationUnits.ActivateOrganizationUnit(rpcCtx, &efv1.ActivateOrganizationUnitRequest{Id: cleanID})
	})
	if err != nil {
		return &model.ActivateOrganizationUnitPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	return &model.ActivateOrganizationUnitPayload{
		OrganizationUnit: toModelOrganizationUnit(resp.GetOrganizationUnit()),
		Errors:           NoMutationErrors(),
		WasSuccessful:    true,
	}
}

// DeactivateOrganizationUnit menonaktifkan unit organisasi.
func (a *Adapter) DeactivateOrganizationUnit(ctx context.Context, id string) *model.DeactivateOrganizationUnitPayload {
	cleanID, idErr := ParseTypedID(id, ResourceOrganizationUnit)
	if idErr != nil {
		if _, authErr := a.requirePerm(ctx, PermOrganizationUnitsWrite, false); authErr != nil {
			return &model.DeactivateOrganizationUnitPayload{Errors: ToMutationErrors(authErr), WasSuccessful: false}
		}
		return &model.DeactivateOrganizationUnitPayload{Errors: ToMutationErrors(idErr), WasSuccessful: false}
	}
	sess, err := a.requireAnyScope(ctx, PermOrganizationUnitsWrite, false, orgUnitScope(cleanID))
	if err != nil {
		return &model.DeactivateOrganizationUnitPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	resp, err := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.DeactivateOrganizationUnitResponse, error) {
		return c.OrganizationUnits.DeactivateOrganizationUnit(rpcCtx, &efv1.DeactivateOrganizationUnitRequest{Id: cleanID})
	})
	if err != nil {
		return &model.DeactivateOrganizationUnitPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	return &model.DeactivateOrganizationUnitPayload{
		OrganizationUnit: toModelOrganizationUnit(resp.GetOrganizationUnit()),
		Errors:           NoMutationErrors(),
		WasSuccessful:    true,
	}
}

// DeleteOrganizationUnit melakukan soft delete unit organisasi dengan wajib re-autentikasi sensitif.
func (a *Adapter) DeleteOrganizationUnit(ctx context.Context, id string) *model.DeleteOrganizationUnitPayload {
	cleanID, idErr := ParseTypedID(id, ResourceOrganizationUnit)
	if idErr != nil {
		if _, authErr := a.requirePerm(ctx, PermOrganizationUnitsWrite, true); authErr != nil {
			return &model.DeleteOrganizationUnitPayload{Errors: ToMutationErrors(authErr), WasSuccessful: false}
		}
		return &model.DeleteOrganizationUnitPayload{Errors: ToMutationErrors(idErr), WasSuccessful: false}
	}
	sess, err := a.requireAnyScope(ctx, PermOrganizationUnitsWrite, true, orgUnitScope(cleanID))
	if err != nil {
		return &model.DeleteOrganizationUnitPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	resp, err := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.DeleteOrganizationUnitResponse, error) {
		return c.OrganizationUnits.DeleteOrganizationUnit(rpcCtx, &efv1.DeleteOrganizationUnitRequest{Id: cleanID})
	})
	if err != nil {
		return &model.DeleteOrganizationUnitPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
	}
	deleted := resp.GetDeleted()
	if !deleted {
		return &model.DeleteOrganizationUnitPayload{
			Errors:        ToMutationErrors(middleware.ErrConflict("organization unit could not be deleted", nil)),
			WasSuccessful: false,
		}
	}
	return &model.DeleteOrganizationUnitPayload{
		Deleted:       &deleted,
		Errors:        NoMutationErrors(),
		WasSuccessful: true,
	}
}

// ResolveOrgUnitParent mengambil induk OrganizationUnit bila ada.
func (a *Adapter) ResolveOrgUnitParent(ctx context.Context, obj *model.OrganizationUnit) (*model.OrganizationUnit, error) {
	if obj == nil || obj.ParentUnitID == nil || strings.TrimSpace(*obj.ParentUnitID) == "" {
		return nil, nil
	}
	return a.OrganizationUnit(ctx, *obj.ParentUnitID)
}

// ResolveOrgUnitPrimaryLocation mengambil lokasi utama milik OrganizationUnit bila ada.
func (a *Adapter) ResolveOrgUnitPrimaryLocation(ctx context.Context, obj *model.OrganizationUnit) (*model.Location, error) {
	if obj == nil || obj.PrimaryLocationID == nil || strings.TrimSpace(*obj.PrimaryLocationID) == "" {
		return nil, nil
	}
	return a.Location(ctx, *obj.PrimaryLocationID)
}
