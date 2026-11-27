package handler_test

import (
	"testing"
	"time"

	efv1 "github.com/integrione/erp-retail/backend/services/enterprise_foundation/contract/gen/enterprise_foundation/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestLegalEntityService_FullRPCCoverage(t *testing.T) {
	t.Parallel()

	h := newTestHarness(t)
	ctx := h.signedCtx(t)

	// Guard: fiscal_year_start_month overflow int16, invalid Timestamp, unknown optional enum, dan nil request.
	if _, err := h.legalEntities.CreateLegalEntity(ctx, &efv1.CreateLegalEntityRequest{
		EntityCode:              "PT-OVF",
		LegalName:               "PT Overflow",
		JurisdictionCountryCode: "ID",
		FiscalYearStartMonth:    65537,
	}); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument on int16 overflow month, got %v", status.Code(err))
	}
	if _, err := h.legalEntities.DissolveLegalEntity(ctx, &efv1.DissolveLegalEntityRequest{
		Id:              "10000000-0000-4000-8000-000000000001",
		DissolutionDate: &timestamppb.Timestamp{Seconds: 1, Nanos: 1_500_000_000},
	}); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument on invalid timestamp, got %v", status.Code(err))
	}
	unknownStatus := efv1.LegalEntityStatus(99)
	if _, err := h.legalEntities.ListLegalEntities(ctx, &efv1.ListLegalEntitiesRequest{
		Status: &unknownStatus,
	}); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument on unknown optional filter enum, got %v", status.Code(err))
	}
	if _, err := h.legalEntities.GetAllLegalEntities(ctx, &efv1.GetAllLegalEntitiesRequest{Page: &efv1.PageRequest{Limit: 10}}); err != nil {
		t.Fatalf("expected empty GetAllLegalEntities to succeed, got %v", err)
	}

	addrResp, err := h.addresses.CreateAddress(ctx, &efv1.CreateAddressRequest{
		AddressLine_1: "Jl. Thamrin No. 10",
		City:          "Jakarta Pusat",
		CountryCode:   "ID",
	})
	if err != nil {
		t.Fatalf("CreateAddress failed: %v", err)
	}

	incorp := timestamppb.New(time.Date(2020, 1, 15, 0, 0, 0, 0, time.UTC))

	// 1. CreateLegalEntity (root utama) + child (pending)
	rootResp, err := h.legalEntities.CreateLegalEntity(ctx, &efv1.CreateLegalEntityRequest{
		EntityCode:               "PT-INT",
		LegalName:                "PT IntegriOne Retail Tbk",
		TradingName:              strPtr("IntegriOne"),
		LegalForm:                strPtr("PT"),
		IsPrimaryOperatingEntity: true,
		Status:                   efv1.LegalEntityStatus_LEGAL_ENTITY_STATUS_ACTIVE,
		JurisdictionCountryCode:  "ID",
		Nib:                      strPtr("1234567890123"),
		TaxIdNumber:              strPtr("01.234.567.8-901.000"),
		IsPkp:                    true,
		RegisteredAddressId:      strPtr(addrResp.GetAddress().GetId()),
		FiscalYearStartMonth:     1,
		IncorporationDate:        incorp,
	})
	if err != nil {
		t.Fatalf("CreateLegalEntity root failed: %v", err)
	}
	root := rootResp.GetLegalEntity()

	// CreateLegalEntity child dalam status PENDING untuk verifikasi ValidateScopeTarget & ActivateLegalEntity.
	childResp, err := h.legalEntities.CreateLegalEntity(ctx, &efv1.CreateLegalEntityRequest{
		ParentLegalEntityId:     strPtr(root.GetId()),
		EntityCode:              "PT-LOG",
		LegalName:               "PT IntegriOne Logistik",
		Status:                  efv1.LegalEntityStatus_LEGAL_ENTITY_STATUS_PENDING,
		JurisdictionCountryCode: "ID",
		Nib:                     strPtr("1234567890123"),
		TaxIdNumber:             strPtr("02.234.567.8-901.000"),
		FiscalYearStartMonth:    4,
	})
	if err != nil {
		t.Fatalf("CreateLegalEntity child failed: %v", err)
	}
	child := childResp.GetLegalEntity()

	// Scope validation fail-closed ketika status masih PENDING, lalu lolos setelah Activate.
	valPending, err := h.legalEntities.ValidateLegalEntityScopeTarget(ctx, &efv1.ValidateLegalEntityScopeTargetRequest{Id: child.GetId()})
	if err == nil && valPending.GetDecision().GetEligible() {
		t.Fatalf("expected non-eligible scope while PENDING, got %+v", valPending)
	}

	// ActivateLegalEntity
	actResp, err := h.legalEntities.ActivateLegalEntity(ctx, &efv1.ActivateLegalEntityRequest{Id: child.GetId()})
	if err != nil || actResp.GetLegalEntity().GetStatus() != efv1.LegalEntityStatus_LEGAL_ENTITY_STATUS_ACTIVE {
		t.Fatalf("ActivateLegalEntity failed: %v / %+v", err, actResp)
	}
	valActive, err := h.legalEntities.ValidateLegalEntityScopeTarget(ctx, &efv1.ValidateLegalEntityScopeTargetRequest{Id: child.GetId()})
	if err != nil || !valActive.GetDecision().GetEligible() || valActive.GetDecision().GetScopeKind() != efv1.ScopeTargetKind_SCOPE_TARGET_KIND_LEGAL_ENTITY {
		t.Fatalf("ValidateLegalEntityScopeTarget failed: err=%v decision=%+v", err, valActive.GetDecision())
	}

	// Lookup RPCs: GetById, GetByIds, GetByCode, GetByTaxIdNumber, ListByNib, GetPrimaryOperating, List, GetAll, ListChildren.
	if r, err := h.legalEntities.GetLegalEntityById(ctx, &efv1.GetLegalEntityByIdRequest{Id: root.GetId()}); err != nil || r.GetLegalEntity().GetEntityCode() != "pt-int" {
		t.Fatalf("GetLegalEntityById failed: %v / %+v", err, r.GetLegalEntity())
	}
	if r, err := h.legalEntities.GetLegalEntitiesByIds(ctx, &efv1.GetLegalEntitiesByIdsRequest{Ids: []string{root.GetId(), child.GetId()}}); err != nil || len(r.GetItems()) != 2 {
		t.Fatalf("GetLegalEntitiesByIds failed: %v len=%d", err, len(r.GetItems()))
	}
	if r, err := h.legalEntities.GetLegalEntityByCode(ctx, &efv1.GetLegalEntityByCodeRequest{EntityCode: "PT-INT"}); err != nil || r.GetLegalEntity().GetId() != root.GetId() {
		t.Fatalf("GetLegalEntityByCode failed: %v", err)
	}
	if r, err := h.legalEntities.GetLegalEntityByTaxIdNumber(ctx, &efv1.GetLegalEntityByTaxIdNumberRequest{TaxIdNumber: "01.234.567.8-901.000"}); err != nil || r.GetLegalEntity().GetId() != root.GetId() {
		t.Fatalf("GetLegalEntityByTaxIdNumber failed: %v", err)
	}
	if r, err := h.legalEntities.ListLegalEntitiesByNib(ctx, &efv1.ListLegalEntitiesByNibRequest{Nib: "1234567890123", Page: &efv1.PageRequest{Limit: 1, Offset: 0}}); err != nil || len(r.GetItems()) != 1 || r.GetPage().GetTotal() != 2 || !r.GetPage().GetHasMore() {
		t.Fatalf("ListLegalEntitiesByNib pagination failed: err=%v page=%+v", err, r.GetPage())
	}
	if r, err := h.legalEntities.GetPrimaryOperatingLegalEntity(ctx, &efv1.GetPrimaryOperatingLegalEntityRequest{}); err != nil || r.GetLegalEntity().GetId() != root.GetId() {
		t.Fatalf("GetPrimaryOperatingLegalEntity failed: %v", err)
	}

	activeStatus := efv1.LegalEntityStatus_LEGAL_ENTITY_STATUS_ACTIVE
	primaryTrue := true
	if r, err := h.legalEntities.ListLegalEntities(ctx, &efv1.ListLegalEntitiesRequest{
		Status:      &activeStatus,
		PrimaryOnly: &primaryTrue,
	}); err != nil || len(r.GetItems()) != 1 {
		t.Fatalf("ListLegalEntities filtered failed: %v len=%d", err, len(r.GetItems()))
	}
	if r, err := h.legalEntities.GetAllLegalEntities(ctx, &efv1.GetAllLegalEntitiesRequest{Page: &efv1.PageRequest{Limit: 1, Offset: 1}}); err != nil || len(r.GetItems()) != 1 || r.GetPage().GetTotal() != 2 {
		t.Fatalf("GetAllLegalEntities failed: %v len=%d", err, len(r.GetItems()))
	}
	if r, err := h.legalEntities.ListLegalEntityChildren(ctx, &efv1.ListLegalEntityChildrenRequest{ParentLegalEntityId: root.GetId()}); err != nil || len(r.GetItems()) != 1 || r.GetItems()[0].GetId() != child.GetId() {
		t.Fatalf("ListLegalEntityChildren failed: %v len=%d", err, len(r.GetItems()))
	}

	// UpdateLegalEntity & SetPrimaryOperatingLegalEntity
	updResp, err := h.legalEntities.UpdateLegalEntity(ctx, &efv1.UpdateLegalEntityRequest{
		Id:                      child.GetId(),
		ParentLegalEntityId:     strPtr(root.GetId()),
		LegalName:               "PT IntegriOne Logistik Nusantara",
		JurisdictionCountryCode: "ID",
		Nib:                     strPtr("1234567890123"),
		TaxIdNumber:             strPtr("02.234.567.8-901.000"),
		IsPkp:                   true,
		FiscalYearStartMonth:    4,
	})
	if err != nil || updResp.GetLegalEntity().GetLegalName() != "PT IntegriOne Logistik Nusantara" {
		t.Fatalf("UpdateLegalEntity failed: %v", err)
	}
	setPrimaryResp, err := h.legalEntities.SetPrimaryOperatingLegalEntity(ctx, &efv1.SetPrimaryOperatingLegalEntityRequest{Id: child.GetId()})
	if err != nil || !setPrimaryResp.GetLegalEntity().GetIsPrimaryOperatingEntity() {
		t.Fatalf("SetPrimaryOperatingLegalEntity failed: %v", err)
	}
	if _, err := h.legalEntities.SetPrimaryOperatingLegalEntity(ctx, &efv1.SetPrimaryOperatingLegalEntityRequest{Id: root.GetId()}); err != nil {
		t.Fatalf("restore primary failed: %v", err)
	}

	// Suspend -> Dissolve -> Delete on child.
	suspResp, err := h.legalEntities.SuspendLegalEntity(ctx, &efv1.SuspendLegalEntityRequest{Id: child.GetId()})
	if err != nil || suspResp.GetLegalEntity().GetStatus() != efv1.LegalEntityStatus_LEGAL_ENTITY_STATUS_SUSPENDED {
		t.Fatalf("SuspendLegalEntity failed: %v", err)
	}
	dissDate := timestamppb.New(time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC))
	dissResp, err := h.legalEntities.DissolveLegalEntity(ctx, &efv1.DissolveLegalEntityRequest{
		Id:              child.GetId(),
		DissolutionDate: dissDate,
	})
	if err != nil || dissResp.GetLegalEntity().GetStatus() != efv1.LegalEntityStatus_LEGAL_ENTITY_STATUS_DISSOLVED {
		t.Fatalf("DissolveLegalEntity failed: %v", err)
	}
	delResp, err := h.legalEntities.DeleteLegalEntity(ctx, &efv1.DeleteLegalEntityRequest{Id: child.GetId()})
	if err != nil || !delResp.GetDeleted() {
		t.Fatalf("DeleteLegalEntity failed: %v", err)
	}
}

func TestLocationAndOrganizationUnitServices_FullRPCCoverage(t *testing.T) {
	t.Parallel()

	h := newTestHarness(t)
	ctx := h.signedCtx(t)

	addrResp, err := h.addresses.CreateAddress(ctx, &efv1.CreateAddressRequest{
		AddressLine_1: "Jl. Raya Bekasi Km 22",
		City:          "Jakarta Timur",
		CountryCode:   "ID",
	})
	if err != nil {
		t.Fatalf("CreateAddress failed: %v", err)
	}
	leResp, err := h.legalEntities.CreateLegalEntity(ctx, &efv1.CreateLegalEntityRequest{
		EntityCode:               "PT-OPS",
		LegalName:                "PT IntegriOne Ops",
		IsPrimaryOperatingEntity: true,
		Status:                   efv1.LegalEntityStatus_LEGAL_ENTITY_STATUS_ACTIVE,
		JurisdictionCountryCode:  "ID",
		RegisteredAddressId:      strPtr(addrResp.GetAddress().GetId()),
		FiscalYearStartMonth:     1,
	})
	if err != nil {
		t.Fatalf("CreateLegalEntity failed: %v", err)
	}
	leID := leResp.GetLegalEntity().GetId()
	addrID := addrResp.GetAddress().GetId()

	// Invalid JSON operational_attributes harus ditolak.
	if _, err := h.locations.CreateLocation(ctx, &efv1.CreateLocationRequest{
		LegalEntityId:             leID,
		LocationCode:              "BAD-JSON",
		Name:                      "Bad JSON Loc",
		Kind:                      efv1.LocationKind_LOCATION_KIND_WAREHOUSE,
		Timezone:                  "Asia/Jakarta",
		OperationalAttributesJson: `["array-not-object"]`,
	}); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument on non-object JSON, got %v", status.Code(err))
	}

	// CreateLocation HQ parent + Warehouse child.
	hqResp, err := h.locations.CreateLocation(ctx, &efv1.CreateLocationRequest{
		LegalEntityId:             leID,
		AddressId:                 &addrID,
		LocationCode:              "HQ-JKT",
		Name:                      "Kantor Pusat Jakarta",
		Kind:                      efv1.LocationKind_LOCATION_KIND_HEADQUARTERS,
		Timezone:                  "Asia/Jakarta",
		OperationalAttributesJson: `{"dock_count":0,"is_hq":true}`,
	})
	if err != nil {
		t.Fatalf("CreateLocation HQ failed: %v", err)
	}
	hq := hqResp.GetLocation()

	whResp, err := h.locations.CreateLocation(ctx, &efv1.CreateLocationRequest{
		LegalEntityId:             leID,
		ParentLocationId:          strPtr(hq.GetId()),
		AddressId:                 &addrID,
		LocationCode:              "WH-JKT-01",
		Name:                      "Gudang Pusat Cakung",
		Kind:                      efv1.LocationKind_LOCATION_KIND_WAREHOUSE,
		Phone:                     strPtr("+62215551234"),
		Timezone:                  "Asia/Jakarta",
		OperationalAttributesJson: `{"cold_storage":false,"bays":8}`,
	})
	if err != nil {
		t.Fatalf("CreateLocation Warehouse failed: %v", err)
	}
	wh := whResp.GetLocation()

	// Seluruh 15 RPC LocationService:
	if r, err := h.locations.GetLocationById(ctx, &efv1.GetLocationByIdRequest{Id: wh.GetId()}); err != nil || r.GetLocation().GetLocationCode() != "wh-jkt-01" {
		t.Fatalf("GetLocationById failed: %v / %+v", err, r.GetLocation())
	}
	if r, err := h.locations.GetLocationsByIds(ctx, &efv1.GetLocationsByIdsRequest{Ids: []string{hq.GetId(), wh.GetId()}}); err != nil || len(r.GetItems()) != 2 {
		t.Fatalf("GetLocationsByIds failed: %v", err)
	}
	if r, err := h.locations.GetLocationByCode(ctx, &efv1.GetLocationByCodeRequest{LocationCode: "WH-JKT-01"}); err != nil || r.GetLocation().GetId() != wh.GetId() {
		t.Fatalf("GetLocationByCode failed: %v", err)
	}
	if r, err := h.locations.ListLocationsByLegalEntityAndKind(ctx, &efv1.ListLocationsByLegalEntityAndKindRequest{
		LegalEntityId: leID,
		Kind:          efv1.LocationKind_LOCATION_KIND_WAREHOUSE,
		ActiveOnly:    true,
	}); err != nil || len(r.GetItems()) != 1 {
		t.Fatalf("ListLocationsByLegalEntityAndKind failed: %v", err)
	}
	if r, err := h.locations.ListLocationsByAddressId(ctx, &efv1.ListLocationsByAddressIdRequest{AddressId: addrID}); err != nil || len(r.GetItems()) != 2 {
		t.Fatalf("ListLocationsByAddressId failed: %v", err)
	}
	if r, err := h.locations.ListLocationChildren(ctx, &efv1.ListLocationChildrenRequest{ParentLocationId: hq.GetId()}); err != nil || len(r.GetItems()) != 1 {
		t.Fatalf("ListLocationChildren failed: %v", err)
	}
	whKind := efv1.LocationKind_LOCATION_KIND_WAREHOUSE
	if r, err := h.locations.ListLocations(ctx, &efv1.ListLocationsRequest{
		LegalEntityId: leID,
		Kind:          &whKind,
		ActiveOnly:    true,
	}); err != nil || len(r.GetItems()) != 1 {
		t.Fatalf("ListLocations failed: %v", err)
	}
	if r, err := h.locations.GetAllLocations(ctx, &efv1.GetAllLocationsRequest{Page: &efv1.PageRequest{Limit: 1, Offset: 0}}); err != nil || !r.GetPage().GetHasMore() {
		t.Fatalf("GetAllLocations pagination failed: %v / %+v", err, r.GetPage())
	}
	if r, err := h.locations.ValidateFacilityScope(ctx, &efv1.ValidateFacilityScopeRequest{Id: hq.GetId()}); err != nil || !r.GetDecision().GetEligible() {
		t.Fatalf("ValidateFacilityScope failed: %v", err)
	}
	if r, err := h.locations.ValidateWarehouseScope(ctx, &efv1.ValidateWarehouseScopeRequest{Id: wh.GetId()}); err != nil || !r.GetDecision().GetEligible() {
		t.Fatalf("ValidateWarehouseScope for warehouse failed: %v", err)
	}
	// HQ bukanlah warehouse sehingga ValidateWarehouseScope harus menolak fail-closed.
	if _, err := h.locations.ValidateWarehouseScope(ctx, &efv1.ValidateWarehouseScopeRequest{Id: hq.GetId()}); err == nil {
		t.Fatal("expected ValidateWarehouseScope to reject non-warehouse HQ")
	}
	if r, err := h.locations.UpdateLocation(ctx, &efv1.UpdateLocationRequest{
		Id:                        wh.GetId(),
		ParentLocationId:          strPtr(hq.GetId()),
		AddressId:                 &addrID,
		Name:                      "Gudang Pusat Cakung Utama",
		Kind:                      efv1.LocationKind_LOCATION_KIND_WAREHOUSE,
		Timezone:                  "Asia/Jakarta",
		OperationalAttributesJson: `{"cold_storage":true,"bays":10}`,
	}); err != nil || r.GetLocation().GetName() != "Gudang Pusat Cakung Utama" {
		t.Fatalf("UpdateLocation failed: %v", err)
	}
	if r, err := h.locations.DeactivateLocation(ctx, &efv1.DeactivateLocationRequest{Id: wh.GetId()}); err != nil || r.GetLocation().GetIsActive() {
		t.Fatalf("DeactivateLocation failed: %v", err)
	}
	if r, err := h.locations.ActivateLocation(ctx, &efv1.ActivateLocationRequest{Id: wh.GetId()}); err != nil || !r.GetLocation().GetIsActive() {
		t.Fatalf("ActivateLocation failed: %v", err)
	}

	// Seluruh 15 RPC OrganizationUnitService:
	divResp, err := h.orgUnits.CreateOrganizationUnit(ctx, &efv1.CreateOrganizationUnitRequest{
		LegalEntityId:     leID,
		PrimaryLocationId: strPtr(hq.GetId()),
		UnitCode:          "DIV-RETAIL",
		Name:              "Divisi Operasi Ritel",
		Kind:              efv1.OrgUnitKind_ORG_UNIT_KIND_DIVISION,
		CostCenterCode:    strPtr("CC-100"),
	})
	if err != nil {
		t.Fatalf("CreateOrganizationUnit division failed: %v", err)
	}
	div := divResp.GetOrganizationUnit()

	headUUID := "20000000-0000-4000-8000-000000000001"
	deptResp, err := h.orgUnits.CreateOrganizationUnit(ctx, &efv1.CreateOrganizationUnitRequest{
		LegalEntityId:     leID,
		ParentUnitId:      strPtr(div.GetId()),
		PrimaryLocationId: strPtr(wh.GetId()),
		UnitCode:          "DEPT-WMS",
		Name:              "Departemen Pergudangan",
		Kind:              efv1.OrgUnitKind_ORG_UNIT_KIND_DEPARTMENT,
		CostCenterCode:    strPtr("CC-110"),
		HeadPersonId:      &headUUID,
	})
	if err != nil {
		t.Fatalf("CreateOrganizationUnit dept failed: %v", err)
	}
	dept := deptResp.GetOrganizationUnit()

	if r, err := h.orgUnits.GetOrganizationUnitById(ctx, &efv1.GetOrganizationUnitByIdRequest{Id: dept.GetId()}); err != nil || r.GetOrganizationUnit().GetUnitCode() != "dept-wms" {
		t.Fatalf("GetOrganizationUnitById failed: %v / %+v", err, r.GetOrganizationUnit())
	}
	if r, err := h.orgUnits.GetOrganizationUnitsByIds(ctx, &efv1.GetOrganizationUnitsByIdsRequest{Ids: []string{div.GetId(), dept.GetId()}}); err != nil || len(r.GetItems()) != 2 {
		t.Fatalf("GetOrganizationUnitsByIds failed: %v", err)
	}
	if r, err := h.orgUnits.GetOrganizationUnitByEntityAndCode(ctx, &efv1.GetOrganizationUnitByEntityAndCodeRequest{LegalEntityId: leID, UnitCode: "DEPT-WMS"}); err != nil || r.GetOrganizationUnit().GetId() != dept.GetId() {
		t.Fatalf("GetOrganizationUnitByEntityAndCode failed: %v", err)
	}
	if r, err := h.orgUnits.ListOrganizationUnitsByHeadPersonId(ctx, &efv1.ListOrganizationUnitsByHeadPersonIdRequest{HeadPersonId: headUUID}); err != nil || len(r.GetItems()) != 1 {
		t.Fatalf("ListOrganizationUnitsByHeadPersonId failed: %v", err)
	}
	if r, err := h.orgUnits.ListOrganizationUnitsByPrimaryLocationId(ctx, &efv1.ListOrganizationUnitsByPrimaryLocationIdRequest{PrimaryLocationId: wh.GetId()}); err != nil || len(r.GetItems()) != 1 {
		t.Fatalf("ListOrganizationUnitsByPrimaryLocationId failed: %v", err)
	}
	if r, err := h.orgUnits.ListOrganizationUnitChildren(ctx, &efv1.ListOrganizationUnitChildrenRequest{ParentUnitId: div.GetId()}); err != nil || len(r.GetItems()) != 1 {
		t.Fatalf("ListOrganizationUnitChildren failed: %v", err)
	}
	deptKind := efv1.OrgUnitKind_ORG_UNIT_KIND_DEPARTMENT
	if r, err := h.orgUnits.ListOrganizationUnits(ctx, &efv1.ListOrganizationUnitsRequest{LegalEntityId: leID, Kind: &deptKind, ActiveOnly: true}); err != nil || len(r.GetItems()) != 1 {
		t.Fatalf("ListOrganizationUnits failed: %v", err)
	}
	if r, err := h.orgUnits.GetAllOrganizationUnits(ctx, &efv1.GetAllOrganizationUnitsRequest{Page: &efv1.PageRequest{Limit: 1, Offset: 0}}); err != nil || !r.GetPage().GetHasMore() {
		t.Fatalf("GetAllOrganizationUnits pagination failed: %v", err)
	}
	if r, err := h.orgUnits.ValidateOrganizationUnitScopeTarget(ctx, &efv1.ValidateOrganizationUnitScopeTargetRequest{Id: dept.GetId()}); err != nil || !r.GetDecision().GetEligible() {
		t.Fatalf("ValidateOrganizationUnitScopeTarget failed: %v", err)
	}
	if r, err := h.orgUnits.UpdateOrganizationUnit(ctx, &efv1.UpdateOrganizationUnitRequest{
		Id:                dept.GetId(),
		ParentUnitId:      strPtr(div.GetId()),
		PrimaryLocationId: strPtr(wh.GetId()),
		Name:              "Departemen Gudang & Logistik",
		Kind:              efv1.OrgUnitKind_ORG_UNIT_KIND_DEPARTMENT,
		CostCenterCode:    strPtr("CC-111"),
		HeadPersonId:      &headUUID,
	}); err != nil || r.GetOrganizationUnit().GetName() != "Departemen Gudang & Logistik" {
		t.Fatalf("UpdateOrganizationUnit failed: %v", err)
	}
	if r, err := h.orgUnits.AssignOrganizationUnitHeadPerson(ctx, &efv1.AssignOrganizationUnitHeadPersonRequest{Id: dept.GetId(), HeadPersonId: nil}); err != nil || r.GetOrganizationUnit().HeadPersonId != nil {
		t.Fatalf("AssignOrganizationUnitHeadPerson unassign failed: %v", err)
	}
	if r, err := h.orgUnits.DeactivateOrganizationUnit(ctx, &efv1.DeactivateOrganizationUnitRequest{Id: dept.GetId()}); err != nil || r.GetOrganizationUnit().GetIsActive() {
		t.Fatalf("DeactivateOrganizationUnit failed: %v", err)
	}
	if r, err := h.orgUnits.ActivateOrganizationUnit(ctx, &efv1.ActivateOrganizationUnitRequest{Id: dept.GetId()}); err != nil || !r.GetOrganizationUnit().GetIsActive() {
		t.Fatalf("ActivateOrganizationUnit failed: %v", err)
	}
	if r, err := h.orgUnits.DeleteOrganizationUnit(ctx, &efv1.DeleteOrganizationUnitRequest{Id: dept.GetId()}); err != nil || !r.GetDeleted() {
		t.Fatalf("DeleteOrganizationUnit failed: %v", err)
	}
	if r, err := h.locations.DeleteLocation(ctx, &efv1.DeleteLocationRequest{Id: wh.GetId()}); err != nil || !r.GetDeleted() {
		t.Fatalf("DeleteLocation failed: %v", err)
	}
}
