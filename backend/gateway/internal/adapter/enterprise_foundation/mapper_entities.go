package enterprise_foundation

import (
	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	efv1 "github.com/integrione/erp-retail/backend/services/enterprise_foundation/contract/gen/enterprise_foundation/v1"
)

func toModelScopeTargetDecision(d *efv1.ScopeTargetDecision) *model.ScopeTargetDecision {
	if d == nil {
		return nil
	}
	kind := mapEnumValue(d.GetScopeKind(), scopeTargetKindToGQL, model.ScopeTargetKindLegalEntity)
	targetID := EncodeGID(scopeTargetResourceForKind(d.GetScopeKind()), d.GetTargetId())
	if targetID == "" {
		targetID = d.GetTargetId()
	}
	legalEntityID := d.GetLegalEntityId()
	if legalEntityID != "" {
		legalEntityID = EncodeGID(ResourceLegalEntity, legalEntityID)
	}
	return &model.ScopeTargetDecision{
		Eligible:      d.GetEligible(),
		ScopeKind:     kind,
		TargetID:      targetID,
		Code:          d.GetCode(),
		LegalEntityID: legalEntityID,
	}
}

func toModelAddress(a *efv1.Address) *model.Address {
	if a == nil {
		return nil
	}
	return &model.Address{
		ID:               EncodeGID(ResourceAddress, a.GetId()),
		RawID:            a.GetId(),
		Label:            cloneOptionalString(a.Label),
		AddressLine1:     a.GetAddressLine_1(),
		AddressLine2:     cloneOptionalString(a.AddressLine_2),
		City:             a.GetCity(),
		Province:         cloneOptionalString(a.Province),
		PostalCode:       cloneOptionalString(a.PostalCode),
		CountryCode:      a.GetCountryCode(),
		FormattedAddress: cloneOptionalString(a.FormattedAddress),
		Latitude:         a.Latitude,
		Longitude:        a.Longitude,
		CreatedAt:        toLocalTime(a.GetCreatedAt()),
		UpdatedAt:        toLocalTime(a.GetUpdatedAt()),
	}
}

func toModelAddresses(items []*efv1.Address) []*model.Address {
	return mapProtoSlice(items, toModelAddress)
}

func toModelLegalEntity(le *efv1.LegalEntity) *model.LegalEntity {
	if le == nil {
		return nil
	}
	return &model.LegalEntity{
		ID:                       EncodeGID(ResourceLegalEntity, le.GetId()),
		RawID:                    le.GetId(),
		ParentLegalEntityID:      encodeOptionalStringPtr(ResourceLegalEntity, le.ParentLegalEntityId),
		EntityCode:               le.GetEntityCode(),
		LegalName:                le.GetLegalName(),
		TradingName:              cloneOptionalString(le.TradingName),
		LegalForm:                cloneOptionalString(le.LegalForm),
		IsPrimaryOperatingEntity: le.GetIsPrimaryOperatingEntity(),
		Status:                   mapEnumValue(le.GetStatus(), legalEntityStatusToGQL, model.LegalEntityStatusPending),
		JurisdictionCountryCode:  le.GetJurisdictionCountryCode(),
		Nib:                      cloneOptionalString(le.Nib),
		TaxIDNumber:              cloneOptionalString(le.TaxIdNumber),
		IsPkp:                    le.GetIsPkp(),
		RegisteredAddressID:      encodeOptionalStringPtr(ResourceAddress, le.RegisteredAddressId),
		FiscalYearStartMonth:     int(le.GetFiscalYearStartMonth()),
		IncorporationDate:        toLocalTimePtr(le.GetIncorporationDate()),
		DissolutionDate:          toLocalTimePtr(le.GetDissolutionDate()),
		CreatedAt:                toLocalTime(le.GetCreatedAt()),
		UpdatedAt:                toLocalTime(le.GetUpdatedAt()),
		DeletedAt:                toLocalTimePtr(le.GetDeletedAt()),
	}
}

func toModelLegalEntities(items []*efv1.LegalEntity) []*model.LegalEntity {
	return mapProtoSlice(items, toModelLegalEntity)
}

func toModelLocation(loc *efv1.Location) *model.Location {
	if loc == nil {
		return nil
	}
	return &model.Location{
		ID:                        EncodeGID(ResourceLocation, loc.GetId()),
		RawID:                     loc.GetId(),
		LegalEntityID:             EncodeGID(ResourceLegalEntity, loc.GetLegalEntityId()),
		ParentLocationID:          encodeOptionalStringPtr(ResourceLocation, loc.ParentLocationId),
		AddressID:                 encodeOptionalStringPtr(ResourceAddress, loc.AddressId),
		LocationCode:              loc.GetLocationCode(),
		Name:                      loc.GetName(),
		Kind:                      mapEnumValue(loc.GetKind(), locationKindToGQL, model.LocationKindWarehouse),
		Phone:                     cloneOptionalString(loc.Phone),
		Timezone:                  loc.GetTimezone(),
		OperationalAttributesJSON: loc.GetOperationalAttributesJson(),
		IsActive:                  loc.GetIsActive(),
		CreatedAt:                 toLocalTime(loc.GetCreatedAt()),
		UpdatedAt:                 toLocalTime(loc.GetUpdatedAt()),
		DeletedAt:                 toLocalTimePtr(loc.GetDeletedAt()),
	}
}

func toModelLocations(items []*efv1.Location) []*model.Location {
	return mapProtoSlice(items, toModelLocation)
}

func toModelOrganizationUnit(unit *efv1.OrganizationUnit) *model.OrganizationUnit {
	if unit == nil {
		return nil
	}
	return &model.OrganizationUnit{
		ID:                EncodeGID(ResourceOrganizationUnit, unit.GetId()),
		RawID:             unit.GetId(),
		LegalEntityID:     EncodeGID(ResourceLegalEntity, unit.GetLegalEntityId()),
		ParentUnitID:      encodeOptionalStringPtr(ResourceOrganizationUnit, unit.ParentUnitId),
		PrimaryLocationID: encodeOptionalStringPtr(ResourceLocation, unit.PrimaryLocationId),
		UnitCode:          unit.GetUnitCode(),
		Name:              unit.GetName(),
		Kind:              mapEnumValue(unit.GetKind(), orgUnitKindToGQL, model.OrgUnitKindDivision),
		CostCenterCode:    cloneOptionalString(unit.CostCenterCode),
		HeadPersonID:      cloneOptionalString(unit.HeadPersonId),
		IsActive:          unit.GetIsActive(),
		CreatedAt:         toLocalTime(unit.GetCreatedAt()),
		UpdatedAt:         toLocalTime(unit.GetUpdatedAt()),
		DeletedAt:         toLocalTimePtr(unit.GetDeletedAt()),
	}
}

func toModelOrganizationUnits(items []*efv1.OrganizationUnit) []*model.OrganizationUnit {
	return mapProtoSlice(items, toModelOrganizationUnit)
}

func toModelParty(p *efv1.Party) *model.Party {
	if p == nil {
		return nil
	}
	return &model.Party{
		ID:                 EncodeGID(ResourceParty, p.GetId()),
		RawID:              p.GetId(),
		PartnerCode:        p.GetPartnerCode(),
		Kind:               mapEnumValue(p.GetKind(), partyKindToGQL, model.PartyKindOrganization),
		DisplayName:        p.GetDisplayName(),
		LegalName:          cloneOptionalString(p.LegalName),
		Email:              cloneOptionalString(p.Email),
		Phone:              cloneOptionalString(p.Phone),
		WebsiteURL:         cloneOptionalString(p.WebsiteUrl),
		TaxIDNumber:        cloneOptionalString(p.TaxIdNumber),
		TaxRegisteredName:  cloneOptionalString(p.TaxRegisteredName),
		IsPkp:              p.GetIsPkp(),
		IsTaxExempt:        p.GetIsTaxExempt(),
		TaxExemptionReason: cloneOptionalString(p.TaxExemptionReason),
		TaxVerifiedAt:      toLocalTimePtr(p.GetTaxVerifiedAt()),
		PreferredLanguage:  p.GetPreferredLanguage(),
		Status:             mapEnumValue(p.GetStatus(), partyStatusToGQL, model.PartyStatusPendingVerification),
		Notes:              cloneOptionalString(p.Notes),
		CreatedAt:          toLocalTime(p.GetCreatedAt()),
		UpdatedAt:          toLocalTime(p.GetUpdatedAt()),
		DeletedAt:          toLocalTimePtr(p.GetDeletedAt()),
	}
}

func toModelParties(items []*efv1.Party) []*model.Party {
	return mapProtoSlice(items, toModelParty)
}

func toModelPartyRole(r *efv1.PartyRole) *model.PartyRole {
	if r == nil {
		return nil
	}
	return &model.PartyRole{
		ID:            EncodeGID(ResourcePartyRole, r.GetId()),
		RawID:         r.GetId(),
		PartyID:       EncodeGID(ResourceParty, r.GetPartyId()),
		Role:          mapEnumValue(r.GetRole(), partyRoleKindToGQL, model.PartyRoleKindOther),
		IsActive:      r.GetIsActive(),
		ReferenceNote: cloneOptionalString(r.ReferenceNote),
		CreatedAt:     toLocalTime(r.GetCreatedAt()),
		UpdatedAt:     toLocalTime(r.GetUpdatedAt()),
	}
}

func toModelPartyContact(c *efv1.PartyContact) *model.PartyContact {
	if c == nil {
		return nil
	}
	return &model.PartyContact{
		ID:                 EncodeGID(ResourcePartyContact, c.GetId()),
		RawID:              c.GetId(),
		PartyID:            EncodeGID(ResourceParty, c.GetPartyId()),
		IamUserID:          encodeOptionalStringPtr(ResourceIAMUser, c.IamUserId),
		FullName:           c.GetFullName(),
		JobTitle:           cloneOptionalString(c.JobTitle),
		ContactRole:        c.GetContactRole(),
		Email:              cloneOptionalString(c.Email),
		Phone:              cloneOptionalString(c.Phone),
		PortalAccessStatus: mapEnumValue(c.GetPortalAccessStatus(), portalAccessStatusToGQL, model.PortalAccessStatusNone),
		IsPrimary:          c.GetIsPrimary(),
		IsActive:           c.GetIsActive(),
		CreatedAt:          toLocalTime(c.GetCreatedAt()),
		UpdatedAt:          toLocalTime(c.GetUpdatedAt()),
		DeletedAt:          toLocalTimePtr(c.GetDeletedAt()),
	}
}

func toModelPartyAddress(pa *efv1.PartyAddress) *model.PartyAddress {
	if pa == nil {
		return nil
	}
	return &model.PartyAddress{
		ID:        EncodeGID(ResourcePartyAddress, pa.GetId()),
		RawID:     pa.GetId(),
		PartyID:   EncodeGID(ResourceParty, pa.GetPartyId()),
		AddressID: EncodeGID(ResourceAddress, pa.GetAddressId()),
		Usage:     mapEnumValue(pa.GetUsage(), addressUsageKindToGQL, model.AddressUsageKindOther),
		IsPrimary: pa.GetIsPrimary(),
		CreatedAt: toLocalTime(pa.GetCreatedAt()),
		UpdatedAt: toLocalTime(pa.GetUpdatedAt()),
	}
}

func toModelPortalAccessEvaluation(ev *efv1.PortalAccessEvaluation) *model.PortalAccessEvaluation {
	if ev == nil {
		return nil
	}
	roles := make([]model.PartyRoleKind, 0, len(ev.GetActiveRoles()))
	for _, r := range ev.GetActiveRoles() {
		roles = append(roles, mapEnumValue(r, partyRoleKindToGQL, model.PartyRoleKindOther))
	}
	return &model.PortalAccessEvaluation{
		Allowed:     ev.GetAllowed(),
		PartyID:     EncodeGID(ResourceParty, ev.GetPartyId()),
		ContactID:   EncodeGID(ResourcePartyContact, ev.GetContactId()),
		IamUserID:   EncodeGID(ResourceIAMUser, ev.GetIamUserId()),
		ActiveRoles: roles,
	}
}

func mapProtoSlice[Proto any, GQL any](items []*Proto, mapper func(*Proto) *GQL) []*GQL {
	out := make([]*GQL, 0, len(items))
	for _, item := range items {
		if mapped := mapper(item); mapped != nil {
			out = append(out, mapped)
		}
	}
	return out
}
