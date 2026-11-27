// Package handler mengimplementasikan registrasi handler gRPC Enterprise Foundation serta mapper dan pemetaan error domain.
package handler

import (
	"fmt"
	"math"
	"time"

	efv1 "github.com/integrione/erp-retail/backend/services/enterprise_foundation/contract/gen/enterprise_foundation/v1"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/entity"
	eferrors "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/repository"
	domainservice "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	codeInvalidInput     = eferrors.Code("enterprise_foundation.invalid_input")
	codeInvalidEnum      = eferrors.Code("enterprise_foundation.invalid_enum")
	codeInvalidTimestamp = eferrors.Code("enterprise_foundation.invalid_timestamp")
)

var (
	protoToDomainLegalEntityStatus = map[efv1.LegalEntityStatus]entity.LegalEntityStatus{
		efv1.LegalEntityStatus_LEGAL_ENTITY_STATUS_PENDING:   entity.LegalEntityStatusPending,
		efv1.LegalEntityStatus_LEGAL_ENTITY_STATUS_ACTIVE:    entity.LegalEntityStatusActive,
		efv1.LegalEntityStatus_LEGAL_ENTITY_STATUS_SUSPENDED: entity.LegalEntityStatusSuspended,
		efv1.LegalEntityStatus_LEGAL_ENTITY_STATUS_DISSOLVED: entity.LegalEntityStatusDissolved,
	}
	domainToProtoLegalEntityStatus = invertEnumMap(protoToDomainLegalEntityStatus)

	protoToDomainOrgUnitKind = map[efv1.OrgUnitKind]entity.OrgUnitKind{
		efv1.OrgUnitKind_ORG_UNIT_KIND_DIVISION:   entity.OrgUnitKindDivision,
		efv1.OrgUnitKind_ORG_UNIT_KIND_DEPARTMENT: entity.OrgUnitKindDepartment,
		efv1.OrgUnitKind_ORG_UNIT_KIND_TEAM:       entity.OrgUnitKindTeam,
		efv1.OrgUnitKind_ORG_UNIT_KIND_BRANCH:     entity.OrgUnitKindBranch,
		efv1.OrgUnitKind_ORG_UNIT_KIND_OTHER:      entity.OrgUnitKindOther,
	}
	domainToProtoOrgUnitKind = invertEnumMap(protoToDomainOrgUnitKind)

	protoToDomainLocationKind = map[efv1.LocationKind]entity.LocationKind{
		efv1.LocationKind_LOCATION_KIND_HEADQUARTERS:  entity.LocationKindHeadquarters,
		efv1.LocationKind_LOCATION_KIND_WAREHOUSE:     entity.LocationKindWarehouse,
		efv1.LocationKind_LOCATION_KIND_RETAIL_STORE:  entity.LocationKindRetailStore,
		efv1.LocationKind_LOCATION_KIND_BRANCH_OFFICE: entity.LocationKindBranchOffice,
	}
	domainToProtoLocationKind = invertEnumMap(protoToDomainLocationKind)

	protoToDomainPartyKind = map[efv1.PartyKind]entity.PartyKind{
		efv1.PartyKind_PARTY_KIND_ORGANIZATION:     entity.PartyKindOrganization,
		efv1.PartyKind_PARTY_KIND_INDIVIDUAL:       entity.PartyKindIndividual,
		efv1.PartyKind_PARTY_KIND_FINANCIAL_ENTITY: entity.PartyKindFinancialEntity,
		efv1.PartyKind_PARTY_KIND_GOVERNMENT_BODY:  entity.PartyKindGovernmentBody,
	}
	domainToProtoPartyKind = invertEnumMap(protoToDomainPartyKind)

	protoToDomainPartyStatus = map[efv1.PartyStatus]entity.PartyStatus{
		efv1.PartyStatus_PARTY_STATUS_PENDING_VERIFICATION: entity.PartyStatusPendingVerification,
		efv1.PartyStatus_PARTY_STATUS_ACTIVE:               entity.PartyStatusActive,
		efv1.PartyStatus_PARTY_STATUS_SUSPENDED:            entity.PartyStatusSuspended,
		efv1.PartyStatus_PARTY_STATUS_INACTIVE:             entity.PartyStatusInactive,
	}
	domainToProtoPartyStatus = invertEnumMap(protoToDomainPartyStatus)

	protoToDomainPartyRoleKind = map[efv1.PartyRoleKind]entity.PartyRoleKind{
		efv1.PartyRoleKind_PARTY_ROLE_KIND_SUPPLIER:              entity.PartyRoleKindSupplier,
		efv1.PartyRoleKind_PARTY_ROLE_KIND_VENDOR:                entity.PartyRoleKindVendor,
		efv1.PartyRoleKind_PARTY_ROLE_KIND_CUSTOMER_B2B:          entity.PartyRoleKindCustomerB2B,
		efv1.PartyRoleKind_PARTY_ROLE_KIND_CARRIER:               entity.PartyRoleKindCarrier,
		efv1.PartyRoleKind_PARTY_ROLE_KIND_FINANCIAL_INSTITUTION: entity.PartyRoleKindFinancialInstitution,
		efv1.PartyRoleKind_PARTY_ROLE_KIND_OTHER:                 entity.PartyRoleKindOther,
	}
	domainToProtoPartyRoleKind = invertEnumMap(protoToDomainPartyRoleKind)

	protoToDomainPortalAccessStatus = map[efv1.PortalAccessStatus]entity.PortalAccessStatus{
		efv1.PortalAccessStatus_PORTAL_ACCESS_STATUS_NONE:         entity.PortalAccessStatusNone,
		efv1.PortalAccessStatus_PORTAL_ACCESS_STATUS_PENDING_LINK: entity.PortalAccessStatusPendingLink,
		efv1.PortalAccessStatus_PORTAL_ACCESS_STATUS_ACTIVE:       entity.PortalAccessStatusActive,
		efv1.PortalAccessStatus_PORTAL_ACCESS_STATUS_REVOKED:      entity.PortalAccessStatusRevoked,
	}
	domainToProtoPortalAccessStatus = invertEnumMap(protoToDomainPortalAccessStatus)

	protoToDomainAddressUsageKind = map[efv1.AddressUsageKind]entity.AddressUsageKind{
		efv1.AddressUsageKind_ADDRESS_USAGE_KIND_REGISTERED_OFFICE: entity.AddressUsageKindRegisteredOffice,
		efv1.AddressUsageKind_ADDRESS_USAGE_KIND_HEADQUARTERS:      entity.AddressUsageKindHeadquarters,
		efv1.AddressUsageKind_ADDRESS_USAGE_KIND_BILLING:           entity.AddressUsageKindBilling,
		efv1.AddressUsageKind_ADDRESS_USAGE_KIND_SHIPPING:          entity.AddressUsageKindShipping,
		efv1.AddressUsageKind_ADDRESS_USAGE_KIND_OTHER:             entity.AddressUsageKindOther,
	}
	domainToProtoAddressUsageKind = invertEnumMap(protoToDomainAddressUsageKind)

	domainToProtoScopeTargetKind = map[domainservice.ScopeTargetKind]efv1.ScopeTargetKind{
		domainservice.ScopeTargetLegalEntity:      efv1.ScopeTargetKind_SCOPE_TARGET_KIND_LEGAL_ENTITY,
		domainservice.ScopeTargetOrganizationUnit: efv1.ScopeTargetKind_SCOPE_TARGET_KIND_ORGANIZATION_UNIT,
		domainservice.ScopeTargetFacility:         efv1.ScopeTargetKind_SCOPE_TARGET_KIND_FACILITY,
		domainservice.ScopeTargetSupplierParty:    efv1.ScopeTargetKind_SCOPE_TARGET_KIND_SUPPLIER_PARTY,
		domainservice.ScopeTargetVendorParty:      efv1.ScopeTargetKind_SCOPE_TARGET_KIND_VENDOR_PARTY,
	}
)

// invertEnumMap membentuk pemetaan balik hanya-baca agar tabel enum Protobuf dan domain selalu simetris.
func invertEnumMap[K comparable, V comparable](src map[K]V) map[V]K {
	out := make(map[V]K, len(src))
	for k, v := range src {
		out[v] = k
	}
	return out
}

// ensureErrorMapper memastikan setiap handler selalu memiliki ErrorMapper aktif meski pemanggil mengirim nil.
func ensureErrorMapper(errMapper *ErrorMapper) *ErrorMapper {
	if errMapper != nil {
		return errMapper
	}
	return NewErrorMapper(nil)
}

// nilRequestError menghasilkan status InvalidArgument ketika payload request gRPC bernilai nil.
func nilRequestError(errMapper *ErrorMapper) error {
	return validationStatusError(errMapper, "request", codeInvalidInput, "request payload is required")
}

// validationStatusError memetakan kesalahan validasi pada layer handler menjadi gRPC InvalidArgument terstruktur.
func validationStatusError(errMapper *ErrorMapper, field string, code eferrors.Code, message string) error {
	if errMapper != nil {
		return errMapper.ToStatus(eferrors.NewValidation(field, code, message))
	}
	return status.Error(codes.InvalidArgument, message)
}

// toStatus mengubah error domain/sistem melalui ErrorMapper atau mengembalikan Internal error fail-closed bila mapper nil.
func toStatus(errMapper *ErrorMapper, err error) error {
	if errMapper != nil {
		return errMapper.ToStatus(err)
	}
	return status.Error(codes.Internal, "internal service error")
}

// mapProtoEnum mengonversi enum Protobuf ke enum domain, memetakan UNSPECIFIED ke string kosong dan menolak nilai tak dikenal.
func mapProtoEnum[P comparable, D ~string](field string, raw P, unspecified P, mapping map[P]D) (D, error) {
	if raw == unspecified {
		return "", nil
	}
	mapped, ok := mapping[raw]
	if !ok {
		return "", eferrors.NewValidation(field, codeInvalidEnum, fmt.Sprintf("invalid %s value", field))
	}
	return mapped, nil
}

// mapOptionalFilterEnum mengonversi enum filter opsional Protobuf: nil atau UNSPECIFIED dianggap unset, sedangkan nilai tak dikenal ditolak.
func mapOptionalFilterEnum[P comparable, D ~string](field string, raw *P, unspecified P, mapping map[P]D) (*D, error) {
	if raw == nil || *raw == unspecified {
		return nil, nil
	}
	mapped, err := mapProtoEnum(field, *raw, unspecified, mapping)
	if err != nil {
		return nil, err
	}
	return &mapped, nil
}

// mapDomainEnum mengonversi enum domain menjadi enum Protobuf atau nilai UNSPECIFIED bila belum dikenali.
func mapDomainEnum[D comparable, P any](raw D, mapping map[D]P, unspecified P) P {
	if mapped, ok := mapping[raw]; ok {
		return mapped
	}
	return unspecified
}

// fromProtoPageRequest mengonversi PageRequest Protobuf menjadi parameter paginasi repository.
func fromProtoPageRequest(page *efv1.PageRequest) domainrepo.PageRequest {
	if page == nil {
		return domainrepo.PageRequest{}
	}
	return domainrepo.PageRequest{
		Limit:  int(page.GetLimit()),
		Offset: int(page.GetOffset()),
	}
}

// toProtoPageInfo mengonversi PageResult repository menjadi PageInfo Protobuf dengan penjagaan batas int32.
func toProtoPageInfo[T any](page domainrepo.PageResult[T]) *efv1.PageInfo {
	return &efv1.PageInfo{
		Total:   page.Total,
		Limit:   clampInt32(page.Limit),
		Offset:  clampInt32(page.Offset),
		HasMore: page.HasMore,
	}
}

// clampInt32 membatasi nilai int agar konversi ke int32 tidak mengalami overflow.
func clampInt32(value int) int32 {
	if value > math.MaxInt32 {
		return math.MaxInt32
	}
	if value < math.MinInt32 {
		return math.MinInt32
	}
	return int32(value)
}

// parseFiscalYearStartMonth memvalidasi fiscal_year_start_month int32 dari Protobuf agar tidak terjadi integer wrap saat konversi ke int16.
func parseFiscalYearStartMonth(value int32) (int16, error) {
	if value < math.MinInt16 || value > math.MaxInt16 {
		return 0, eferrors.NewValidation("fiscal_year_start_month", codeInvalidInput, "fiscal_year_start_month is out of valid integer range")
	}
	return int16(value), nil
}

// toProtoTimestamp mengonversi time.Time menjadi Timestamp Protobuf UTC bila tidak bernilai nol.
func toProtoTimestamp(t time.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return nil
	}
	return timestamppb.New(t.UTC())
}

// toProtoOptionalTimestamp mengonversi pointer time.Time menjadi Timestamp Protobuf UTC.
func toProtoOptionalTimestamp(t *time.Time) *timestamppb.Timestamp {
	if t == nil || t.IsZero() {
		return nil
	}
	return timestamppb.New(t.UTC())
}

// fromProtoOptionalTimestamp memvalidasi dan mengonversi Timestamp Protobuf opsional ke *time.Time UTC.
func fromProtoOptionalTimestamp(field string, ts *timestamppb.Timestamp) (*time.Time, error) {
	if ts == nil {
		return nil, nil
	}
	if err := ts.CheckValid(); err != nil {
		return nil, eferrors.NewValidation(field, codeInvalidTimestamp, fmt.Sprintf("invalid %s timestamp", field))
	}
	t := ts.AsTime().UTC()
	return &t, nil
}

// clonePtr menyalin nilai pointer generik agar request dan response Protobuf terisolasi dari mutasi pointer.
func clonePtr[T any](v *T) *T {
	if v == nil {
		return nil
	}
	copyVal := *v
	return &copyVal
}

// toProtoScopeDecision memetakan ScopeTargetDecision domain ke pesan ScopeTargetDecision Protobuf.
func toProtoScopeDecision(decision domainservice.ScopeTargetDecision) *efv1.ScopeTargetDecision {
	return &efv1.ScopeTargetDecision{
		Eligible:      decision.Eligible,
		ScopeKind:     mapDomainEnum(decision.ScopeKind, domainToProtoScopeTargetKind, efv1.ScopeTargetKind_SCOPE_TARGET_KIND_UNSPECIFIED),
		TargetId:      decision.TargetID,
		Code:          decision.Code,
		LegalEntityId: decision.LegalEntityID,
	}
}
