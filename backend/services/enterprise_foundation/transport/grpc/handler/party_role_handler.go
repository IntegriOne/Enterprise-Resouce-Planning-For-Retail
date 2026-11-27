package handler

import (
	"context"

	efv1 "github.com/integrione/erp-retail/backend/services/enterprise_foundation/contract/gen/enterprise_foundation/v1"
	domainrepo "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/repository"
	partysvc "github.com/integrione/erp-retail/backend/services/enterprise_foundation/services/party"
)

// GetPartyRoleById mengambil satu peran bisnis mitra berdasarkan UUID baris peran.
func (h *PartyHandler) GetPartyRoleById(ctx context.Context, req *efv1.GetPartyRoleByIdRequest) (*efv1.GetPartyRoleByIdResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	role, err := h.svc.GetRoleByID(ctx, req.GetRoleId())
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.GetPartyRoleByIdResponse{Role: toProtoPartyRole(role)}, nil
}

// ListPartyRolesByParty mengambil daftar peran bisnis milik satu mitra secara berhalaman.
func (h *PartyHandler) ListPartyRolesByParty(ctx context.Context, req *efv1.ListPartyRolesByPartyRequest) (*efv1.ListPartyRolesByPartyResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	page, err := h.svc.ListRolesByParty(ctx, req.GetPartyId(), fromProtoPageRequest(req.GetPage()))
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.ListPartyRolesByPartyResponse{
		Items: toProtoPartyRoles(page.Items),
		Page:  toProtoPageInfo(page),
	}, nil
}

// GetPartyRoleByPartyAndKind mengambil satu peran spesifik milik mitra berdasarkan jenis peran.
func (h *PartyHandler) GetPartyRoleByPartyAndKind(ctx context.Context, req *efv1.GetPartyRoleByPartyAndKindRequest) (*efv1.GetPartyRoleByPartyAndKindResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	kind, err := mapProtoEnum("kind", req.GetKind(), efv1.PartyRoleKind_PARTY_ROLE_KIND_UNSPECIFIED, protoToDomainPartyRoleKind)
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	role, err := h.svc.GetRoleByPartyAndKind(ctx, req.GetPartyId(), kind)
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.GetPartyRoleByPartyAndKindResponse{Role: toProtoPartyRole(role)}, nil
}

// ListPartyRoles mencari daftar peran bisnis mitra berdasarkan filter dan parameter paginasi.
func (h *PartyHandler) ListPartyRoles(ctx context.Context, req *efv1.ListPartyRolesRequest) (*efv1.ListPartyRolesResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	roleFilter, err := mapOptionalFilterEnum("role", req.Role, efv1.PartyRoleKind_PARTY_ROLE_KIND_UNSPECIFIED, protoToDomainPartyRoleKind)
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	page, err := h.svc.ListRoles(ctx, domainrepo.PartyRoleFilter{
		PartyID:    req.GetPartyId(),
		Role:       roleFilter,
		ActiveOnly: req.GetActiveOnly(),
		Query:      req.GetQuery(),
		Page:       fromProtoPageRequest(req.GetPage()),
	})
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.ListPartyRolesResponse{
		Items: toProtoPartyRoles(page.Items),
		Page:  toProtoPageInfo(page),
	}, nil
}

// GetAllPartyRoles mengambil seluruh peran bisnis mitra secara berhalaman.
func (h *PartyHandler) GetAllPartyRoles(ctx context.Context, req *efv1.GetAllPartyRolesRequest) (*efv1.GetAllPartyRolesResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	page, err := h.svc.GetAllRoles(ctx, fromProtoPageRequest(req.GetPage()))
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.GetAllPartyRolesResponse{
		Items: toProtoPartyRoles(page.Items),
		Page:  toProtoPageInfo(page),
	}, nil
}

// ListPartyRolesByKind mengambil daftar peran bisnis mitra berdasarkan jenis peran.
func (h *PartyHandler) ListPartyRolesByKind(ctx context.Context, req *efv1.ListPartyRolesByKindRequest) (*efv1.ListPartyRolesByKindResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	kind, err := mapProtoEnum("kind", req.GetKind(), efv1.PartyRoleKind_PARTY_ROLE_KIND_UNSPECIFIED, protoToDomainPartyRoleKind)
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	page, err := h.svc.ListRolesByKind(ctx, kind, req.GetActiveOnly(), fromProtoPageRequest(req.GetPage()))
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.ListPartyRolesByKindResponse{
		Items: toProtoPartyRoles(page.Items),
		Page:  toProtoPageInfo(page),
	}, nil
}

// GetAllPartyRolesByKind mengambil seluruh peran bisnis mitra berdasarkan jenis peran secara berhalaman.
func (h *PartyHandler) GetAllPartyRolesByKind(ctx context.Context, req *efv1.GetAllPartyRolesByKindRequest) (*efv1.GetAllPartyRolesByKindResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	kind, err := mapProtoEnum("kind", req.GetKind(), efv1.PartyRoleKind_PARTY_ROLE_KIND_UNSPECIFIED, protoToDomainPartyRoleKind)
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	page, err := h.svc.GetAllRolesByKind(ctx, kind, req.GetActiveOnly(), fromProtoPageRequest(req.GetPage()))
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.GetAllPartyRolesByKindResponse{
		Items: toProtoPartyRoles(page.Items),
		Page:  toProtoPageInfo(page),
	}, nil
}

// ValidatePartyRoleScope memeriksa kelayakan mitra terhadap target scope IAM supplier_party/vendor_party secara fail-closed.
func (h *PartyHandler) ValidatePartyRoleScope(ctx context.Context, req *efv1.ValidatePartyRoleScopeRequest) (*efv1.ValidatePartyRoleScopeResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	requiredRole, err := mapProtoEnum("required_role", req.GetRequiredRole(), efv1.PartyRoleKind_PARTY_ROLE_KIND_UNSPECIFIED, protoToDomainPartyRoleKind)
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	decision, err := h.svc.ValidatePartyRoleScope(ctx, req.GetPartyId(), requiredRole)
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.ValidatePartyRoleScopeResponse{Decision: toProtoScopeDecision(decision)}, nil
}

// AssignPartyRole menetapkan peran bisnis baru kepada mitra eksternal.
func (h *PartyHandler) AssignPartyRole(ctx context.Context, req *efv1.AssignPartyRoleRequest) (*efv1.AssignPartyRoleResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	role, err := mapProtoEnum("role", req.GetRole(), efv1.PartyRoleKind_PARTY_ROLE_KIND_UNSPECIFIED, protoToDomainPartyRoleKind)
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	assigned, err := h.svc.AssignRole(ctx, partysvc.AssignRoleInput{
		PartyID:       req.GetPartyId(),
		Role:          role,
		ReferenceNote: clonePtr(req.ReferenceNote),
	})
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.AssignPartyRoleResponse{Role: toProtoPartyRole(assigned)}, nil
}

// UpdatePartyRoleNote memperbarui catatan referensi pada peran bisnis mitra.
func (h *PartyHandler) UpdatePartyRoleNote(ctx context.Context, req *efv1.UpdatePartyRoleNoteRequest) (*efv1.UpdatePartyRoleNoteResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	updated, err := h.svc.UpdateRoleNote(ctx, partysvc.UpdateRoleInput{
		ID:            req.GetId(),
		ReferenceNote: clonePtr(req.ReferenceNote),
	})
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.UpdatePartyRoleNoteResponse{Role: toProtoPartyRole(updated)}, nil
}

// ActivatePartyRole mengaktifkan peran bisnis mitra yang sebelumnya nonaktif.
func (h *PartyHandler) ActivatePartyRole(ctx context.Context, req *efv1.ActivatePartyRoleRequest) (*efv1.ActivatePartyRoleResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	updated, err := h.svc.ActivateRole(ctx, req.GetRoleId())
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.ActivatePartyRoleResponse{Role: toProtoPartyRole(updated)}, nil
}

// DeactivatePartyRole menonaktifkan peran bisnis mitra dan mencatat outbox sinkronisasi IAM.
func (h *PartyHandler) DeactivatePartyRole(ctx context.Context, req *efv1.DeactivatePartyRoleRequest) (*efv1.DeactivatePartyRoleResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	updated, err := h.svc.DeactivateRole(ctx, req.GetRoleId())
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.DeactivatePartyRoleResponse{Role: toProtoPartyRole(updated)}, nil
}

// RemovePartyRole menghapus penetapan peran bisnis dari mitra eksternal.
func (h *PartyHandler) RemovePartyRole(ctx context.Context, req *efv1.RemovePartyRoleRequest) (*efv1.RemovePartyRoleResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	if err := h.svc.RemoveRole(ctx, req.GetRoleId()); err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.RemovePartyRoleResponse{Removed: true}, nil
}
