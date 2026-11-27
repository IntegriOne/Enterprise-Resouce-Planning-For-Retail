package handler

import (
	efv1 "github.com/integrione/erp-retail/backend/services/enterprise_foundation/contract/gen/enterprise_foundation/v1"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/entity"
)

// toProtoAddress memetakan entitas Address domain ke pesan Address Protobuf.
func toProtoAddress(addr *entity.Address) *efv1.Address {
	if addr == nil {
		return nil
	}
	return &efv1.Address{
		Id:               addr.ID,
		Label:            clonePtr(addr.Label),
		AddressLine_1:    addr.AddressLine1,
		AddressLine_2:    clonePtr(addr.AddressLine2),
		City:             addr.City,
		Province:         clonePtr(addr.Province),
		PostalCode:       clonePtr(addr.PostalCode),
		CountryCode:      addr.CountryCode,
		FormattedAddress: clonePtr(addr.FormattedAddress),
		Latitude:         clonePtr(addr.Latitude),
		Longitude:        clonePtr(addr.Longitude),
		CreatedAt:        toProtoTimestamp(addr.CreatedAt),
		UpdatedAt:        toProtoTimestamp(addr.UpdatedAt),
	}
}

// toProtoAddresses memetakan irisan entitas Address domain ke irisan pesan Address Protobuf.
func toProtoAddresses(items []*entity.Address) []*efv1.Address {
	out := make([]*efv1.Address, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		out = append(out, toProtoAddress(item))
	}
	return out
}

// toProtoLegalEntity memetakan entitas LegalEntity domain ke pesan LegalEntity Protobuf.
func toProtoLegalEntity(le *entity.LegalEntity) *efv1.LegalEntity {
	if le == nil {
		return nil
	}
	return &efv1.LegalEntity{
		Id:                       le.ID,
		ParentLegalEntityId:      clonePtr(le.ParentLegalEntityID),
		EntityCode:               le.EntityCode,
		LegalName:                le.LegalName,
		TradingName:              clonePtr(le.TradingName),
		LegalForm:                clonePtr(le.LegalForm),
		IsPrimaryOperatingEntity: le.IsPrimaryOperatingEntity,
		Status:                   mapDomainEnum(le.Status, domainToProtoLegalEntityStatus, efv1.LegalEntityStatus_LEGAL_ENTITY_STATUS_UNSPECIFIED),
		JurisdictionCountryCode:  le.JurisdictionCountryCode,
		Nib:                      clonePtr(le.NIB),
		TaxIdNumber:              clonePtr(le.TaxIDNumber),
		IsPkp:                    le.IsPKP,
		RegisteredAddressId:      clonePtr(le.RegisteredAddressID),
		FiscalYearStartMonth:     int32(le.FiscalYearStartMonth),
		IncorporationDate:        toProtoOptionalTimestamp(le.IncorporationDate),
		DissolutionDate:          toProtoOptionalTimestamp(le.DissolutionDate),
		CreatedAt:                toProtoTimestamp(le.CreatedAt),
		UpdatedAt:                toProtoTimestamp(le.UpdatedAt),
		DeletedAt:                toProtoOptionalTimestamp(le.DeletedAt),
	}
}

// toProtoLegalEntities memetakan irisan entitas LegalEntity domain ke irisan pesan LegalEntity Protobuf.
func toProtoLegalEntities(items []*entity.LegalEntity) []*efv1.LegalEntity {
	out := make([]*efv1.LegalEntity, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		out = append(out, toProtoLegalEntity(item))
	}
	return out
}

// toProtoLocation memetakan entitas Location domain ke pesan Location Protobuf.
func toProtoLocation(loc *entity.Location) *efv1.Location {
	if loc == nil {
		return nil
	}
	return &efv1.Location{
		Id:                        loc.ID,
		LegalEntityId:             loc.LegalEntityID,
		ParentLocationId:          clonePtr(loc.ParentLocationID),
		AddressId:                 clonePtr(loc.AddressID),
		LocationCode:              loc.LocationCode,
		Name:                      loc.Name,
		Kind:                      mapDomainEnum(loc.Kind, domainToProtoLocationKind, efv1.LocationKind_LOCATION_KIND_UNSPECIFIED),
		Phone:                     clonePtr(loc.Phone),
		Timezone:                  loc.Timezone,
		OperationalAttributesJson: string(loc.OperationalAttributesCopy()),
		IsActive:                  loc.IsActive,
		CreatedAt:                 toProtoTimestamp(loc.CreatedAt),
		UpdatedAt:                 toProtoTimestamp(loc.UpdatedAt),
		DeletedAt:                 toProtoOptionalTimestamp(loc.DeletedAt),
	}
}

// toProtoLocations memetakan irisan entitas Location domain ke irisan pesan Location Protobuf.
func toProtoLocations(items []*entity.Location) []*efv1.Location {
	out := make([]*efv1.Location, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		out = append(out, toProtoLocation(item))
	}
	return out
}

// toProtoOrganizationUnit memetakan entitas OrganizationUnit domain ke pesan OrganizationUnit Protobuf.
func toProtoOrganizationUnit(unit *entity.OrganizationUnit) *efv1.OrganizationUnit {
	if unit == nil {
		return nil
	}
	return &efv1.OrganizationUnit{
		Id:                unit.ID,
		LegalEntityId:     unit.LegalEntityID,
		ParentUnitId:      clonePtr(unit.ParentUnitID),
		PrimaryLocationId: clonePtr(unit.PrimaryLocationID),
		UnitCode:          unit.UnitCode,
		Name:              unit.Name,
		Kind:              mapDomainEnum(unit.Kind, domainToProtoOrgUnitKind, efv1.OrgUnitKind_ORG_UNIT_KIND_UNSPECIFIED),
		CostCenterCode:    clonePtr(unit.CostCenterCode),
		HeadPersonId:      clonePtr(unit.HeadPersonID),
		IsActive:          unit.IsActive,
		CreatedAt:         toProtoTimestamp(unit.CreatedAt),
		UpdatedAt:         toProtoTimestamp(unit.UpdatedAt),
		DeletedAt:         toProtoOptionalTimestamp(unit.DeletedAt),
	}
}

// toProtoOrganizationUnits memetakan irisan entitas OrganizationUnit domain ke irisan pesan OrganizationUnit Protobuf.
func toProtoOrganizationUnits(items []*entity.OrganizationUnit) []*efv1.OrganizationUnit {
	out := make([]*efv1.OrganizationUnit, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		out = append(out, toProtoOrganizationUnit(item))
	}
	return out
}
