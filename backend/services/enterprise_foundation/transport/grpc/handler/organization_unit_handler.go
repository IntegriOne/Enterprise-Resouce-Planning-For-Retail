package handler

import (
	"context"

	efv1 "github.com/integrione/erp-retail/backend/services/enterprise_foundation/contract/gen/enterprise_foundation/v1"
	eferrors "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/repository"
	orgunitsvc "github.com/integrione/erp-retail/backend/services/enterprise_foundation/services/organizationunit"
)

// OrganizationUnitHandler menghubungkan kontrak gRPC OrganizationUnitService ke layanan unit organisasi domain.
type OrganizationUnitHandler struct {
	efv1.UnimplementedOrganizationUnitServiceServer
	svc       *orgunitsvc.Service
	errMapper *ErrorMapper
}

// NewOrganizationUnitHandler membentuk OrganizationUnitHandler dengan validasi fail-closed atas dependensi service.
func NewOrganizationUnitHandler(svc *orgunitsvc.Service, errMapper *ErrorMapper) (*OrganizationUnitHandler, error) {
	if svc == nil {
		return nil, eferrors.NewValidation("organization_unit_service", eferrors.Code("enterprise_foundation.handler.nil_organization_unit_service"), "organization unit service is required")
	}
	return &OrganizationUnitHandler{
		svc:       svc,
		errMapper: ensureErrorMapper(errMapper),
	}, nil
}

// GetOrganizationUnitById mengambil satu unit organisasi berdasarkan UUID.
func (h *OrganizationUnitHandler) GetOrganizationUnitById(ctx context.Context, req *efv1.GetOrganizationUnitByIdRequest) (*efv1.GetOrganizationUnitByIdResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	unit, err := h.svc.GetByID(ctx, req.GetId())
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.GetOrganizationUnitByIdResponse{OrganizationUnit: toProtoOrganizationUnit(unit)}, nil
}

// GetOrganizationUnitsByIds mengambil daftar unit organisasi berdasarkan kumpulan UUID.
func (h *OrganizationUnitHandler) GetOrganizationUnitsByIds(ctx context.Context, req *efv1.GetOrganizationUnitsByIdsRequest) (*efv1.GetOrganizationUnitsByIdsResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	items, err := h.svc.GetByIDs(ctx, req.GetIds())
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.GetOrganizationUnitsByIdsResponse{Items: toProtoOrganizationUnits(items)}, nil
}

// GetOrganizationUnitByEntityAndCode mengambil satu unit organisasi berdasarkan badan hukum dan kode unik unit.
func (h *OrganizationUnitHandler) GetOrganizationUnitByEntityAndCode(ctx context.Context, req *efv1.GetOrganizationUnitByEntityAndCodeRequest) (*efv1.GetOrganizationUnitByEntityAndCodeResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	unit, err := h.svc.GetByEntityAndCode(ctx, req.GetLegalEntityId(), req.GetUnitCode())
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.GetOrganizationUnitByEntityAndCodeResponse{OrganizationUnit: toProtoOrganizationUnit(unit)}, nil
}

// ListOrganizationUnitsByHeadPersonId mengambil daftar unit organisasi yang dipimpin oleh satu kepala unit HRM.
func (h *OrganizationUnitHandler) ListOrganizationUnitsByHeadPersonId(ctx context.Context, req *efv1.ListOrganizationUnitsByHeadPersonIdRequest) (*efv1.ListOrganizationUnitsByHeadPersonIdResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	page, err := h.svc.ListByHeadPersonID(ctx, req.GetHeadPersonId(), fromProtoPageRequest(req.GetPage()))
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.ListOrganizationUnitsByHeadPersonIdResponse{
		Items: toProtoOrganizationUnits(page.Items),
		Page:  toProtoPageInfo(page),
	}, nil
}

// ListOrganizationUnitsByPrimaryLocationId mengambil daftar unit organisasi yang berpusat pada satu lokasi fisik.
func (h *OrganizationUnitHandler) ListOrganizationUnitsByPrimaryLocationId(ctx context.Context, req *efv1.ListOrganizationUnitsByPrimaryLocationIdRequest) (*efv1.ListOrganizationUnitsByPrimaryLocationIdResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	page, err := h.svc.ListByPrimaryLocationID(ctx, req.GetPrimaryLocationId(), fromProtoPageRequest(req.GetPage()))
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.ListOrganizationUnitsByPrimaryLocationIdResponse{
		Items: toProtoOrganizationUnits(page.Items),
		Page:  toProtoPageInfo(page),
	}, nil
}

// ListOrganizationUnitChildren mengambil daftar sub-unit organisasi langsung di bawah unit induk.
func (h *OrganizationUnitHandler) ListOrganizationUnitChildren(ctx context.Context, req *efv1.ListOrganizationUnitChildrenRequest) (*efv1.ListOrganizationUnitChildrenResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	page, err := h.svc.ListChildren(ctx, req.GetParentUnitId(), fromProtoPageRequest(req.GetPage()))
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.ListOrganizationUnitChildrenResponse{
		Items: toProtoOrganizationUnits(page.Items),
		Page:  toProtoPageInfo(page),
	}, nil
}

// ListOrganizationUnits mencari daftar unit organisasi berdasarkan filter dan parameter paginasi.
func (h *OrganizationUnitHandler) ListOrganizationUnits(ctx context.Context, req *efv1.ListOrganizationUnitsRequest) (*efv1.ListOrganizationUnitsResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	kindFilter, err := mapOptionalFilterEnum("kind", req.Kind, efv1.OrgUnitKind_ORG_UNIT_KIND_UNSPECIFIED, protoToDomainOrgUnitKind)
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	page, err := h.svc.List(ctx, domainrepo.OrganizationUnitFilter{
		Query:             req.GetQuery(),
		LegalEntityID:     req.GetLegalEntityId(),
		ParentUnitID:      clonePtr(req.ParentUnitId),
		PrimaryLocationID: clonePtr(req.PrimaryLocationId),
		Kind:              kindFilter,
		ActiveOnly:        req.GetActiveOnly(),
		IncludeDeleted:    req.GetIncludeDeleted(),
		Page:              fromProtoPageRequest(req.GetPage()),
	})
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.ListOrganizationUnitsResponse{
		Items: toProtoOrganizationUnits(page.Items),
		Page:  toProtoPageInfo(page),
	}, nil
}

// GetAllOrganizationUnits mengambil seluruh unit organisasi aktif secara berhalaman.
func (h *OrganizationUnitHandler) GetAllOrganizationUnits(ctx context.Context, req *efv1.GetAllOrganizationUnitsRequest) (*efv1.GetAllOrganizationUnitsResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	page, err := h.svc.GetAll(ctx, fromProtoPageRequest(req.GetPage()))
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.GetAllOrganizationUnitsResponse{
		Items: toProtoOrganizationUnits(page.Items),
		Page:  toProtoPageInfo(page),
	}, nil
}

// ValidateOrganizationUnitScopeTarget memvalidasi kelayakan unit organisasi sebagai target scope IAM secara fail-closed.
func (h *OrganizationUnitHandler) ValidateOrganizationUnitScopeTarget(ctx context.Context, req *efv1.ValidateOrganizationUnitScopeTargetRequest) (*efv1.ValidateOrganizationUnitScopeTargetResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	decision, err := h.svc.ValidateScopeTarget(ctx, req.GetId())
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.ValidateOrganizationUnitScopeTargetResponse{Decision: toProtoScopeDecision(decision)}, nil
}

// CreateOrganizationUnit membentuk unit organisasi baru dengan validasi badan hukum, hierarki, lokasi utama, dan duplikasi kode.
func (h *OrganizationUnitHandler) CreateOrganizationUnit(ctx context.Context, req *efv1.CreateOrganizationUnitRequest) (*efv1.CreateOrganizationUnitResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	kind, err := mapProtoEnum("kind", req.GetKind(), efv1.OrgUnitKind_ORG_UNIT_KIND_UNSPECIFIED, protoToDomainOrgUnitKind)
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	created, err := h.svc.Create(ctx, orgunitsvc.CreateInput{
		LegalEntityID:     req.GetLegalEntityId(),
		ParentUnitID:      clonePtr(req.ParentUnitId),
		PrimaryLocationID: clonePtr(req.PrimaryLocationId),
		UnitCode:          req.GetUnitCode(),
		Name:              req.GetName(),
		Kind:              kind,
		CostCenterCode:    clonePtr(req.CostCenterCode),
		HeadPersonID:      clonePtr(req.HeadPersonId),
	})
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.CreateOrganizationUnitResponse{OrganizationUnit: toProtoOrganizationUnit(created)}, nil
}

// UpdateOrganizationUnit memperbarui struktur hierarki, lokasi utama, dan atribut unit organisasi.
func (h *OrganizationUnitHandler) UpdateOrganizationUnit(ctx context.Context, req *efv1.UpdateOrganizationUnitRequest) (*efv1.UpdateOrganizationUnitResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	kind, err := mapProtoEnum("kind", req.GetKind(), efv1.OrgUnitKind_ORG_UNIT_KIND_UNSPECIFIED, protoToDomainOrgUnitKind)
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	updated, err := h.svc.Update(ctx, orgunitsvc.UpdateInput{
		ID:                 req.GetId(),
		ParentUnitID:       clonePtr(req.ParentUnitId),
		PrimaryLocationID:  clonePtr(req.PrimaryLocationId),
		Name:               req.GetName(),
		Kind:               kind,
		CostCenterCode:     clonePtr(req.CostCenterCode),
		HeadPersonID:       clonePtr(req.HeadPersonId),
		PreserveHeadPerson: req.GetPreserveHeadPerson(),
	})
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.UpdateOrganizationUnitResponse{OrganizationUnit: toProtoOrganizationUnit(updated)}, nil
}

// AssignOrganizationUnitHeadPerson menetapkan atau melepas referensi kepala unit dari domain HRM.
func (h *OrganizationUnitHandler) AssignOrganizationUnitHeadPerson(ctx context.Context, req *efv1.AssignOrganizationUnitHeadPersonRequest) (*efv1.AssignOrganizationUnitHeadPersonResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	updated, err := h.svc.AssignHeadPerson(ctx, req.GetId(), clonePtr(req.HeadPersonId))
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.AssignOrganizationUnitHeadPersonResponse{OrganizationUnit: toProtoOrganizationUnit(updated)}, nil
}

// ActivateOrganizationUnit mengaktifkan kembali unit organisasi.
func (h *OrganizationUnitHandler) ActivateOrganizationUnit(ctx context.Context, req *efv1.ActivateOrganizationUnitRequest) (*efv1.ActivateOrganizationUnitResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	updated, err := h.svc.Activate(ctx, req.GetId())
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.ActivateOrganizationUnitResponse{OrganizationUnit: toProtoOrganizationUnit(updated)}, nil
}

// DeactivateOrganizationUnit menonaktifkan sementara unit organisasi dari operasional baru.
func (h *OrganizationUnitHandler) DeactivateOrganizationUnit(ctx context.Context, req *efv1.DeactivateOrganizationUnitRequest) (*efv1.DeactivateOrganizationUnitResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	updated, err := h.svc.Deactivate(ctx, req.GetId())
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.DeactivateOrganizationUnitResponse{OrganizationUnit: toProtoOrganizationUnit(updated)}, nil
}

// DeleteOrganizationUnit melakukan soft delete unit organisasi apabila tidak memiliki sub-unit anak.
func (h *OrganizationUnitHandler) DeleteOrganizationUnit(ctx context.Context, req *efv1.DeleteOrganizationUnitRequest) (*efv1.DeleteOrganizationUnitResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	if err := h.svc.Delete(ctx, req.GetId()); err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.DeleteOrganizationUnitResponse{Deleted: true}, nil
}
