package handler

import (
	efv1 "github.com/integrione/erp-retail/backend/services/enterprise_foundation/contract/gen/enterprise_foundation/v1"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/entity"
	domainservice "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/service"
)

// toProtoParty memetakan entitas Party domain ke pesan Party Protobuf.
func toProtoParty(p *entity.Party) *efv1.Party {
	if p == nil {
		return nil
	}
	return &efv1.Party{
		Id:                 p.ID,
		PartnerCode:        p.PartnerCode,
		Kind:               mapDomainEnum(p.Kind, domainToProtoPartyKind, efv1.PartyKind_PARTY_KIND_UNSPECIFIED),
		DisplayName:        p.DisplayName,
		LegalName:          clonePtr(p.LegalName),
		Email:              clonePtr(p.Email),
		Phone:              clonePtr(p.Phone),
		WebsiteUrl:         clonePtr(p.WebsiteURL),
		TaxIdNumber:        clonePtr(p.TaxIDNumber),
		TaxRegisteredName:  clonePtr(p.TaxRegisteredName),
		IsPkp:              p.IsPKP,
		IsTaxExempt:        p.IsTaxExempt,
		TaxExemptionReason: clonePtr(p.TaxExemptionReason),
		TaxVerifiedAt:      toProtoOptionalTimestamp(p.TaxVerifiedAt),
		PreferredLanguage:  p.PreferredLanguage,
		Status:             mapDomainEnum(p.Status, domainToProtoPartyStatus, efv1.PartyStatus_PARTY_STATUS_UNSPECIFIED),
		Notes:              clonePtr(p.Notes),
		CreatedAt:          toProtoTimestamp(p.CreatedAt),
		UpdatedAt:          toProtoTimestamp(p.UpdatedAt),
		DeletedAt:          toProtoOptionalTimestamp(p.DeletedAt),
	}
}

// toProtoParties memetakan irisan entitas Party domain ke irisan pesan Party Protobuf.
func toProtoParties(items []*entity.Party) []*efv1.Party {
	out := make([]*efv1.Party, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		out = append(out, toProtoParty(item))
	}
	return out
}

// toProtoPartyRole memetakan entitas PartyRole domain ke pesan PartyRole Protobuf.
func toProtoPartyRole(role *entity.PartyRole) *efv1.PartyRole {
	if role == nil {
		return nil
	}
	return &efv1.PartyRole{
		Id:            role.ID,
		PartyId:       role.PartyID,
		Role:          mapDomainEnum(role.Role, domainToProtoPartyRoleKind, efv1.PartyRoleKind_PARTY_ROLE_KIND_UNSPECIFIED),
		IsActive:      role.IsActive,
		ReferenceNote: clonePtr(role.ReferenceNote),
		CreatedAt:     toProtoTimestamp(role.CreatedAt),
		UpdatedAt:     toProtoTimestamp(role.UpdatedAt),
	}
}

// toProtoPartyRoles memetakan irisan entitas PartyRole domain ke irisan pesan PartyRole Protobuf.
func toProtoPartyRoles(items []*entity.PartyRole) []*efv1.PartyRole {
	out := make([]*efv1.PartyRole, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		out = append(out, toProtoPartyRole(item))
	}
	return out
}

// toProtoPartyContact memetakan entitas PartyContact domain ke pesan PartyContact Protobuf.
func toProtoPartyContact(contact *entity.PartyContact) *efv1.PartyContact {
	if contact == nil {
		return nil
	}
	return &efv1.PartyContact{
		Id:                 contact.ID,
		PartyId:            contact.PartyID,
		IamUserId:          clonePtr(contact.IAMUserID),
		FullName:           contact.FullName,
		JobTitle:           clonePtr(contact.JobTitle),
		ContactRole:        contact.ContactRole,
		Email:              clonePtr(contact.Email),
		Phone:              clonePtr(contact.Phone),
		PortalAccessStatus: mapDomainEnum(contact.PortalAccessStatus, domainToProtoPortalAccessStatus, efv1.PortalAccessStatus_PORTAL_ACCESS_STATUS_UNSPECIFIED),
		IsPrimary:          contact.IsPrimary,
		IsActive:           contact.IsActive,
		CreatedAt:          toProtoTimestamp(contact.CreatedAt),
		UpdatedAt:          toProtoTimestamp(contact.UpdatedAt),
		DeletedAt:          toProtoOptionalTimestamp(contact.DeletedAt),
	}
}

// toProtoPartyContacts memetakan irisan entitas PartyContact domain ke irisan pesan PartyContact Protobuf.
func toProtoPartyContacts(items []*entity.PartyContact) []*efv1.PartyContact {
	out := make([]*efv1.PartyContact, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		out = append(out, toProtoPartyContact(item))
	}
	return out
}

// toProtoPartyAddress memetakan entitas PartyAddress domain ke pesan PartyAddress Protobuf.
func toProtoPartyAddress(link *entity.PartyAddress) *efv1.PartyAddress {
	if link == nil {
		return nil
	}
	return &efv1.PartyAddress{
		Id:        link.ID,
		PartyId:   link.PartyID,
		AddressId: link.AddressID,
		Usage:     mapDomainEnum(link.Usage, domainToProtoAddressUsageKind, efv1.AddressUsageKind_ADDRESS_USAGE_KIND_UNSPECIFIED),
		IsPrimary: link.IsPrimary,
		CreatedAt: toProtoTimestamp(link.CreatedAt),
		UpdatedAt: toProtoTimestamp(link.UpdatedAt),
	}
}

// toProtoPartyAddresses memetakan irisan entitas PartyAddress domain ke irisan pesan PartyAddress Protobuf.
func toProtoPartyAddresses(items []*entity.PartyAddress) []*efv1.PartyAddress {
	out := make([]*efv1.PartyAddress, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		out = append(out, toProtoPartyAddress(item))
	}
	return out
}

// toProtoPortalAccessEvaluation memetakan hasil evaluasi akses portal PIC mitra ke pesan Protobuf.
func toProtoPortalAccessEvaluation(eval domainservice.PortalAccessEvaluation) *efv1.PortalAccessEvaluation {
	roles := make([]efv1.PartyRoleKind, 0, len(eval.ActiveRoles))
	for _, role := range eval.ActiveRoles {
		roles = append(roles, mapDomainEnum(role, domainToProtoPartyRoleKind, efv1.PartyRoleKind_PARTY_ROLE_KIND_UNSPECIFIED))
	}
	return &efv1.PortalAccessEvaluation{
		Allowed:     eval.Allowed,
		PartyId:     eval.PartyID,
		ContactId:   eval.ContactID,
		IamUserId:   eval.IAMUserID,
		ActiveRoles: roles,
	}
}
