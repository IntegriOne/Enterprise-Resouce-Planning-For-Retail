package enterprise_foundation

import (
	"strings"
	"time"

	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	efv1 "github.com/integrione/erp-retail/backend/services/enterprise_foundation/contract/gen/enterprise_foundation/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	wibZoneOnce = time.FixedZone("WIB", 7*3600)

	legalEntityStatusToGQL = map[efv1.LegalEntityStatus]model.LegalEntityStatus{
		efv1.LegalEntityStatus_LEGAL_ENTITY_STATUS_PENDING:   model.LegalEntityStatusPending,
		efv1.LegalEntityStatus_LEGAL_ENTITY_STATUS_ACTIVE:    model.LegalEntityStatusActive,
		efv1.LegalEntityStatus_LEGAL_ENTITY_STATUS_SUSPENDED: model.LegalEntityStatusSuspended,
		efv1.LegalEntityStatus_LEGAL_ENTITY_STATUS_DISSOLVED: model.LegalEntityStatusDissolved,
	}
	legalEntityStatusToProto = invertLookup(legalEntityStatusToGQL)

	locationKindToGQL = map[efv1.LocationKind]model.LocationKind{
		efv1.LocationKind_LOCATION_KIND_HEADQUARTERS:  model.LocationKindHeadquarters,
		efv1.LocationKind_LOCATION_KIND_WAREHOUSE:     model.LocationKindWarehouse,
		efv1.LocationKind_LOCATION_KIND_RETAIL_STORE:  model.LocationKindRetailStore,
		efv1.LocationKind_LOCATION_KIND_BRANCH_OFFICE: model.LocationKindBranchOffice,
	}
	locationKindToProto = invertLookup(locationKindToGQL)

	orgUnitKindToGQL = map[efv1.OrgUnitKind]model.OrgUnitKind{
		efv1.OrgUnitKind_ORG_UNIT_KIND_DIVISION:   model.OrgUnitKindDivision,
		efv1.OrgUnitKind_ORG_UNIT_KIND_DEPARTMENT: model.OrgUnitKindDepartment,
		efv1.OrgUnitKind_ORG_UNIT_KIND_TEAM:       model.OrgUnitKindTeam,
		efv1.OrgUnitKind_ORG_UNIT_KIND_BRANCH:     model.OrgUnitKindBranch,
		efv1.OrgUnitKind_ORG_UNIT_KIND_OTHER:      model.OrgUnitKindOther,
	}
	orgUnitKindToProto = invertLookup(orgUnitKindToGQL)

	partyKindToGQL = map[efv1.PartyKind]model.PartyKind{
		efv1.PartyKind_PARTY_KIND_ORGANIZATION:     model.PartyKindOrganization,
		efv1.PartyKind_PARTY_KIND_INDIVIDUAL:       model.PartyKindIndividual,
		efv1.PartyKind_PARTY_KIND_FINANCIAL_ENTITY: model.PartyKindFinancialEntity,
		efv1.PartyKind_PARTY_KIND_GOVERNMENT_BODY:  model.PartyKindGovernmentBody,
	}
	partyKindToProto = invertLookup(partyKindToGQL)

	partyStatusToGQL = map[efv1.PartyStatus]model.PartyStatus{
		efv1.PartyStatus_PARTY_STATUS_PENDING_VERIFICATION: model.PartyStatusPendingVerification,
		efv1.PartyStatus_PARTY_STATUS_ACTIVE:               model.PartyStatusActive,
		efv1.PartyStatus_PARTY_STATUS_SUSPENDED:            model.PartyStatusSuspended,
		efv1.PartyStatus_PARTY_STATUS_INACTIVE:             model.PartyStatusInactive,
	}
	partyStatusToProto = invertLookup(partyStatusToGQL)

	partyRoleKindToGQL = map[efv1.PartyRoleKind]model.PartyRoleKind{
		efv1.PartyRoleKind_PARTY_ROLE_KIND_SUPPLIER:              model.PartyRoleKindSupplier,
		efv1.PartyRoleKind_PARTY_ROLE_KIND_VENDOR:                model.PartyRoleKindVendor,
		efv1.PartyRoleKind_PARTY_ROLE_KIND_CUSTOMER_B2B:          model.PartyRoleKindCustomerB2b,
		efv1.PartyRoleKind_PARTY_ROLE_KIND_CARRIER:               model.PartyRoleKindCarrier,
		efv1.PartyRoleKind_PARTY_ROLE_KIND_FINANCIAL_INSTITUTION: model.PartyRoleKindFinancialInstitution,
		efv1.PartyRoleKind_PARTY_ROLE_KIND_OTHER:                 model.PartyRoleKindOther,
	}
	partyRoleKindToProto = invertLookup(partyRoleKindToGQL)

	portalAccessStatusToGQL = map[efv1.PortalAccessStatus]model.PortalAccessStatus{
		efv1.PortalAccessStatus_PORTAL_ACCESS_STATUS_NONE:         model.PortalAccessStatusNone,
		efv1.PortalAccessStatus_PORTAL_ACCESS_STATUS_PENDING_LINK: model.PortalAccessStatusPendingLink,
		efv1.PortalAccessStatus_PORTAL_ACCESS_STATUS_ACTIVE:       model.PortalAccessStatusActive,
		efv1.PortalAccessStatus_PORTAL_ACCESS_STATUS_REVOKED:      model.PortalAccessStatusRevoked,
	}
	portalAccessStatusToProto = invertLookup(portalAccessStatusToGQL)

	addressUsageKindToGQL = map[efv1.AddressUsageKind]model.AddressUsageKind{
		efv1.AddressUsageKind_ADDRESS_USAGE_KIND_REGISTERED_OFFICE: model.AddressUsageKindRegisteredOffice,
		efv1.AddressUsageKind_ADDRESS_USAGE_KIND_HEADQUARTERS:      model.AddressUsageKindHeadquarters,
		efv1.AddressUsageKind_ADDRESS_USAGE_KIND_BILLING:           model.AddressUsageKindBilling,
		efv1.AddressUsageKind_ADDRESS_USAGE_KIND_SHIPPING:          model.AddressUsageKindShipping,
		efv1.AddressUsageKind_ADDRESS_USAGE_KIND_OTHER:             model.AddressUsageKindOther,
	}
	addressUsageKindToProto = invertLookup(addressUsageKindToGQL)

	scopeTargetKindToGQL = map[efv1.ScopeTargetKind]model.ScopeTargetKind{
		efv1.ScopeTargetKind_SCOPE_TARGET_KIND_LEGAL_ENTITY:      model.ScopeTargetKindLegalEntity,
		efv1.ScopeTargetKind_SCOPE_TARGET_KIND_ORGANIZATION_UNIT: model.ScopeTargetKindOrganizationUnit,
		efv1.ScopeTargetKind_SCOPE_TARGET_KIND_FACILITY:          model.ScopeTargetKindFacility,
		efv1.ScopeTargetKind_SCOPE_TARGET_KIND_SUPPLIER_PARTY:    model.ScopeTargetKindSupplierParty,
		efv1.ScopeTargetKind_SCOPE_TARGET_KIND_VENDOR_PARTY:      model.ScopeTargetKindVendorParty,
	}
)

func invertLookup[K comparable, V comparable](src map[K]V) map[V]K {
	out := make(map[V]K, len(src))
	for k, v := range src {
		out[v] = k
	}
	return out
}

func mapEnumValue[K comparable, V any](key K, table map[K]V, fallback V) V {
	if mapped, ok := table[key]; ok {
		return mapped
	}
	return fallback
}

func mapOptionalEnumPtr[G comparable, P any](ptr *G, table map[G]P) *P {
	if ptr == nil {
		return nil
	}
	if mapped, ok := table[*ptr]; ok {
		return &mapped
	}
	return nil
}

func toLocalTime(ts *timestamppb.Timestamp) time.Time {
	if ts == nil {
		return time.Time{}
	}
	return ts.AsTime().In(wibZoneOnce)
}

func toLocalTimePtr(ts *timestamppb.Timestamp) *time.Time {
	if ts == nil {
		return nil
	}
	t := ts.AsTime().In(wibZoneOnce)
	return &t
}

func toProtoTimePtr(t *time.Time) *timestamppb.Timestamp {
	if t == nil {
		return nil
	}
	return timestamppb.New(*t)
}

func derefStr(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

func derefBool(ptr *bool, fallback bool) bool {
	if ptr == nil {
		return fallback
	}
	return *ptr
}

func derefInt32(ptr *int, fallback int32) int32 {
	if ptr == nil {
		return fallback
	}
	return int32(*ptr)
}

func cloneOptionalString(ptr *string) *string {
	if ptr == nil {
		return nil
	}
	val := *ptr
	return &val
}

func encodeOptionalStringPtr(resource GIDResource, raw *string) *string {
	if raw == nil {
		return nil
	}
	return EncodeOptionalGID(resource, *raw)
}

func normalizeStringDefault(ptr *string, fallback string) string {
	if ptr == nil {
		return fallback
	}
	trimmed := strings.TrimSpace(*ptr)
	if trimmed == "" {
		return fallback
	}
	return trimmed
}

func scopeTargetResourceForKind(kind efv1.ScopeTargetKind) GIDResource {
	switch kind {
	case efv1.ScopeTargetKind_SCOPE_TARGET_KIND_LEGAL_ENTITY:
		return ResourceLegalEntity
	case efv1.ScopeTargetKind_SCOPE_TARGET_KIND_ORGANIZATION_UNIT:
		return ResourceOrganizationUnit
	case efv1.ScopeTargetKind_SCOPE_TARGET_KIND_FACILITY:
		return ResourceLocation
	case efv1.ScopeTargetKind_SCOPE_TARGET_KIND_SUPPLIER_PARTY, efv1.ScopeTargetKind_SCOPE_TARGET_KIND_VENDOR_PARTY:
		return ResourceParty
	default:
		return ResourceLegalEntity
	}
}
