package handler

import (
	"context"

	efv1 "github.com/integrione/erp-retail/backend/services/enterprise_foundation/contract/gen/enterprise_foundation/v1"
	domainrepo "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/repository"
	partysvc "github.com/integrione/erp-retail/backend/services/enterprise_foundation/services/party"
)

// GetPartyContactById mengambil satu PIC/kontak mitra berdasarkan UUID kontak.
func (h *PartyHandler) GetPartyContactById(ctx context.Context, req *efv1.GetPartyContactByIdRequest) (*efv1.GetPartyContactByIdResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	contact, err := h.svc.GetContactByID(ctx, req.GetContactId())
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.GetPartyContactByIdResponse{Contact: toProtoPartyContact(contact)}, nil
}

// ListPartyContactsByParty mengambil daftar PIC/kontak milik satu mitra secara berhalaman.
func (h *PartyHandler) ListPartyContactsByParty(ctx context.Context, req *efv1.ListPartyContactsByPartyRequest) (*efv1.ListPartyContactsByPartyResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	page, err := h.svc.ListContactsByParty(ctx, req.GetPartyId(), fromProtoPageRequest(req.GetPage()))
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.ListPartyContactsByPartyResponse{
		Items: toProtoPartyContacts(page.Items),
		Page:  toProtoPageInfo(page),
	}, nil
}

// GetPrimaryPartyContact mengambil PIC utama milik satu mitra.
func (h *PartyHandler) GetPrimaryPartyContact(ctx context.Context, req *efv1.GetPrimaryPartyContactRequest) (*efv1.GetPrimaryPartyContactResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	contact, err := h.svc.GetPrimaryContact(ctx, req.GetPartyId())
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.GetPrimaryPartyContactResponse{Contact: toProtoPartyContact(contact)}, nil
}

// ListPartyContactsByIamUserId mengambil daftar kontak mitra yang tertaut ke satu akun IAM.
func (h *PartyHandler) ListPartyContactsByIamUserId(ctx context.Context, req *efv1.ListPartyContactsByIamUserIdRequest) (*efv1.ListPartyContactsByIamUserIdResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	page, err := h.svc.ListContactsByIAMUserID(ctx, req.GetIamUserId(), fromProtoPageRequest(req.GetPage()))
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.ListPartyContactsByIamUserIdResponse{
		Items: toProtoPartyContacts(page.Items),
		Page:  toProtoPageInfo(page),
	}, nil
}

// ListPartyContactsByEmail mengambil daftar kontak mitra berdasarkan email secara berhalaman.
func (h *PartyHandler) ListPartyContactsByEmail(ctx context.Context, req *efv1.ListPartyContactsByEmailRequest) (*efv1.ListPartyContactsByEmailResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	page, err := h.svc.ListContactsByEmail(ctx, req.GetEmail(), fromProtoPageRequest(req.GetPage()))
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.ListPartyContactsByEmailResponse{
		Items: toProtoPartyContacts(page.Items),
		Page:  toProtoPageInfo(page),
	}, nil
}

// ListPartyContacts mencari daftar PIC/kontak mitra berdasarkan filter dan parameter paginasi.
func (h *PartyHandler) ListPartyContacts(ctx context.Context, req *efv1.ListPartyContactsRequest) (*efv1.ListPartyContactsResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	portalStatusFilter, err := mapOptionalFilterEnum("portal_access_status", req.PortalAccessStatus, efv1.PortalAccessStatus_PORTAL_ACCESS_STATUS_UNSPECIFIED, protoToDomainPortalAccessStatus)
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	page, err := h.svc.ListContacts(ctx, domainrepo.PartyContactFilter{
		PartyID:            req.GetPartyId(),
		Query:              req.GetQuery(),
		ContactRole:        req.GetContactRole(),
		PortalAccessStatus: portalStatusFilter,
		ActiveOnly:         req.GetActiveOnly(),
		IncludeDeleted:     req.GetIncludeDeleted(),
		Page:               fromProtoPageRequest(req.GetPage()),
	})
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.ListPartyContactsResponse{
		Items: toProtoPartyContacts(page.Items),
		Page:  toProtoPageInfo(page),
	}, nil
}

// GetAllPartyContacts mengambil seluruh kontak mitra aktif secara berhalaman.
func (h *PartyHandler) GetAllPartyContacts(ctx context.Context, req *efv1.GetAllPartyContactsRequest) (*efv1.GetAllPartyContactsResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	page, err := h.svc.GetAllContacts(ctx, fromProtoPageRequest(req.GetPage()))
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.GetAllPartyContactsResponse{
		Items: toProtoPartyContacts(page.Items),
		Page:  toProtoPageInfo(page),
	}, nil
}

// EvaluateRepresentativePortalAccess memeriksa kelayakan PIC mitra mewakili mitra pada portal IAM secara fail-closed.
func (h *PartyHandler) EvaluateRepresentativePortalAccess(ctx context.Context, req *efv1.EvaluateRepresentativePortalAccessRequest) (*efv1.EvaluateRepresentativePortalAccessResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	requiredRole, err := mapProtoEnum("required_role", req.GetRequiredRole(), efv1.PartyRoleKind_PARTY_ROLE_KIND_UNSPECIFIED, protoToDomainPartyRoleKind)
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	eval, err := h.svc.EvaluateRepresentativePortalAccess(ctx, req.GetPartyId(), req.GetIamUserId(), requiredRole)
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.EvaluateRepresentativePortalAccessResponse{Evaluation: toProtoPortalAccessEvaluation(eval)}, nil
}

// AddPartyContact menambahkan PIC/kontak mitra baru dengan validasi keunikan tautan portal IAM dan kanal komunikasi.
func (h *PartyHandler) AddPartyContact(ctx context.Context, req *efv1.AddPartyContactRequest) (*efv1.AddPartyContactResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	portalStatus, err := mapProtoEnum("portal_access_status", req.GetPortalAccessStatus(), efv1.PortalAccessStatus_PORTAL_ACCESS_STATUS_UNSPECIFIED, protoToDomainPortalAccessStatus)
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	created, err := h.svc.AddContact(ctx, partysvc.AddContactInput{
		PartyID:            req.GetPartyId(),
		IAMUserID:          clonePtr(req.IamUserId),
		FullName:           req.GetFullName(),
		JobTitle:           clonePtr(req.JobTitle),
		ContactRole:        req.GetContactRole(),
		Email:              clonePtr(req.Email),
		Phone:              clonePtr(req.Phone),
		PortalAccessStatus: portalStatus,
		IsPrimary:          req.GetIsPrimary(),
	})
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.AddPartyContactResponse{Contact: toProtoPartyContact(created)}, nil
}

// UpdatePartyContact memperbarui profil kontak/PIC mitra.
func (h *PartyHandler) UpdatePartyContact(ctx context.Context, req *efv1.UpdatePartyContactRequest) (*efv1.UpdatePartyContactResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	updated, err := h.svc.UpdateContact(ctx, partysvc.UpdateContactInput{
		ID:          req.GetId(),
		FullName:    req.GetFullName(),
		JobTitle:    clonePtr(req.JobTitle),
		ContactRole: req.GetContactRole(),
		Email:       clonePtr(req.Email),
		Phone:       clonePtr(req.Phone),
		IsPrimary:   req.GetIsPrimary(),
	})
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.UpdatePartyContactResponse{Contact: toProtoPartyContact(updated)}, nil
}

// SetPrimaryPartyContact menetapkan satu PIC aktif sebagai kontak utama mitra secara atomik.
func (h *PartyHandler) SetPrimaryPartyContact(ctx context.Context, req *efv1.SetPrimaryPartyContactRequest) (*efv1.SetPrimaryPartyContactResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	updated, err := h.svc.SetPrimaryContact(ctx, req.GetPartyId(), req.GetContactId())
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.SetPrimaryPartyContactResponse{Contact: toProtoPartyContact(updated)}, nil
}

// RequestPartyPortalLink menautkan kontak dengan principal IAM dalam status pending_link.
func (h *PartyHandler) RequestPartyPortalLink(ctx context.Context, req *efv1.RequestPartyPortalLinkRequest) (*efv1.RequestPartyPortalLinkResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	updated, err := h.svc.RequestPortalLink(ctx, req.GetContactId(), req.GetIamUserId())
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.RequestPartyPortalLinkResponse{Contact: toProtoPartyContact(updated)}, nil
}

// ActivatePartyPortalAccess mengaktifkan akses portal PIC terhadap principal IAM secara atomik.
func (h *PartyHandler) ActivatePartyPortalAccess(ctx context.Context, req *efv1.ActivatePartyPortalAccessRequest) (*efv1.ActivatePartyPortalAccessResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	updated, err := h.svc.ActivatePortalAccess(ctx, req.GetContactId(), req.GetIamUserId())
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.ActivatePartyPortalAccessResponse{Contact: toProtoPartyContact(updated)}, nil
}

// RevokePartyPortalAccess mencabut akses portal PIC dan mencatat outbox ke IAM.
func (h *PartyHandler) RevokePartyPortalAccess(ctx context.Context, req *efv1.RevokePartyPortalAccessRequest) (*efv1.RevokePartyPortalAccessResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	updated, err := h.svc.RevokePortalAccess(ctx, req.GetContactId())
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.RevokePartyPortalAccessResponse{Contact: toProtoPartyContact(updated)}, nil
}

// UnlinkPartyPortalUser melepas keterkaitan akun IAM pada kontak dan mengembalikan status portal ke none.
func (h *PartyHandler) UnlinkPartyPortalUser(ctx context.Context, req *efv1.UnlinkPartyPortalUserRequest) (*efv1.UnlinkPartyPortalUserResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	updated, err := h.svc.UnlinkPortalUser(ctx, req.GetContactId())
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.UnlinkPartyPortalUserResponse{Contact: toProtoPartyContact(updated)}, nil
}

// ActivatePartyContact mengaktifkan kembali baris kontak mitra.
func (h *PartyHandler) ActivatePartyContact(ctx context.Context, req *efv1.ActivatePartyContactRequest) (*efv1.ActivatePartyContactResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	updated, err := h.svc.ActivateContact(ctx, req.GetContactId())
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.ActivatePartyContactResponse{Contact: toProtoPartyContact(updated)}, nil
}

// DeactivatePartyContact menonaktifkan kontak mitra dan akses portal terkait bila ada.
func (h *PartyHandler) DeactivatePartyContact(ctx context.Context, req *efv1.DeactivatePartyContactRequest) (*efv1.DeactivatePartyContactResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	updated, err := h.svc.DeactivateContact(ctx, req.GetContactId())
	if err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.DeactivatePartyContactResponse{Contact: toProtoPartyContact(updated)}, nil
}

// DeletePartyContact melakukan soft delete kontak mitra.
func (h *PartyHandler) DeletePartyContact(ctx context.Context, req *efv1.DeletePartyContactRequest) (*efv1.DeletePartyContactResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	if err := h.svc.DeleteContact(ctx, req.GetContactId()); err != nil {
		return nil, toStatus(h.errMapper, err)
	}
	return &efv1.DeletePartyContactResponse{Deleted: true}, nil
}
