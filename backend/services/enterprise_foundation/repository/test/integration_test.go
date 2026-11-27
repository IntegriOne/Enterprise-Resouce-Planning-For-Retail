package repository_test

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"gorm.io/gorm"

	efdb "github.com/integrione/erp-retail/backend/services/enterprise_foundation/db"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/entity"
	eferrors "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/repository"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/repository"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/repository/pgstore"
)

func mustUUID(t *testing.T) string {
	t.Helper()
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		t.Fatalf("failed to generate uuid bytes: %v", err)
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

func resolveLiveTestDSN(t *testing.T) string {
	t.Helper()
	if testDSN := strings.TrimSpace(os.Getenv("EF_TEST_DATABASE_DSN")); testDSN != "" {
		return testDSN
	}
	if devDSN := strings.TrimSpace(os.Getenv("EF_DATABASE_DSN")); devDSN != "" {
		derived, err := efdb.DeriveTestDatabaseDSN(devDSN)
		if err != nil {
			t.Fatalf("failed to derive test dsn: %v", err)
		}
		return derived
	}
	t.Skip("skipping live PostgreSQL repository integration test: set EF_TEST_DATABASE_DSN or EF_DATABASE_DSN")
	return ""
}

func openLiveTestRegistry(t *testing.T) (*gorm.DB, *repository.Registry) {
	t.Helper()
	dsn := resolveLiveTestDSN(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	db, err := efdb.Init(ctx, dsn, true)
	if err != nil {
		t.Fatalf("efdb.Init failed: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("db.DB() failed: %v", err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	reg, err := repository.New(db)
	if err != nil {
		t.Fatalf("repository.New failed: %v", err)
	}
	return db, reg
}

func truncateFoundationTables(t *testing.T, db *gorm.DB) {
	t.Helper()
	if err := db.Exec(`TRUNCATE TABLE
		foundation_outbox_events,
		party_addresses,
		party_contacts,
		party_roles,
		parties,
		organization_units,
		locations,
		legal_entities,
		addresses CASCADE`).Error; err != nil {
		t.Fatalf("failed to truncate test tables: %v", err)
	}
}

func TestLivePostgresRepositories(t *testing.T) {
	db, reg := openLiveTestRegistry(t)
	truncateFoundationTables(t, db)
	t.Cleanup(func() {
		truncateFoundationTables(t, db)
	})

	ctx := context.Background()

	t.Run("AddressRepository CRUD, batch SelectByIDs, filtering, and restrict guard", func(t *testing.T) {
		truncateFoundationTables(t, db)
		now := time.Now().Truncate(time.Microsecond)

		lat, lng := -6.9174640, 107.6191230
		addr1, err := entity.NewAddress(mustUUID(t), entity.AddressParams{
			Label:        strPtr("Kantor Pusat"),
			AddressLine1: "Jl. Asia Afrika No. 100",
			City:         "Bandung",
			Province:     strPtr("Jawa Barat"),
			PostalCode:   strPtr("40111"),
			CountryCode:  "id",
			Latitude:     &lat,
			Longitude:    &lng,
		}, now)
		if err != nil {
			t.Fatalf("NewAddress failed: %v", err)
		}
		addr2, err := entity.NewAddress(mustUUID(t), entity.AddressParams{
			Label:        strPtr("Gudang_100%"),
			AddressLine1: "Jl. Industri Raya No. 8",
			City:         "Bekasi",
			Province:     strPtr("Jawa Barat"),
			PostalCode:   strPtr("17530"),
			CountryCode:  "ID",
		}, now)
		if err != nil {
			t.Fatalf("NewAddress 2 failed: %v", err)
		}

		if err := reg.Addresses.Insert(ctx, addr1); err != nil {
			t.Fatalf("Insert addr1 failed: %v", err)
		}
		if err := reg.Addresses.Insert(ctx, addr2); err != nil {
			t.Fatalf("Insert addr2 failed: %v", err)
		}

		got, err := reg.Addresses.SelectByID(ctx, strings.ToUpper(addr1.ID))
		if err != nil {
			t.Fatalf("SelectByID failed: %v", err)
		}
		if got.CountryCode != "ID" || got.Latitude == nil || *got.Latitude != lat {
			t.Fatalf("unexpected address round-trip: %+v", got)
		}

		batch, err := reg.Addresses.SelectByIDs(ctx, []string{addr2.ID, addr1.ID, addr1.ID})
		if err != nil || len(batch) != 2 {
			t.Fatalf("SelectByIDs failed: len=%d err=%v", len(batch), err)
		}

		page, err := reg.Addresses.SelectList(ctx, domainrepo.AddressFilter{
			Query:      "100%",
			Province:   "jawa barat",
			PostalCode: "17530",
			Page:       domainrepo.PageRequest{Limit: 10, Offset: 0},
		})
		if err != nil || page.Total != 1 || page.Items[0].ID != addr2.ID {
			t.Fatalf("SelectList escaped wildcard/postal filter failed: page=%+v err=%v", page, err)
		}

		if err := addr1.Update(entity.AddressParams{
			AddressLine1: "Jl. Asia Afrika No. 102",
			City:         "Bandung",
			CountryCode:  "ID",
		}, now.Add(time.Minute)); err != nil {
			t.Fatalf("addr1.Update failed: %v", err)
		}
		if err := reg.Addresses.Update(ctx, addr1); err != nil {
			t.Fatalf("Addresses.Update failed: %v", err)
		}
		updated, err := reg.Addresses.SelectByID(ctx, addr1.ID)
		if err != nil || updated.Label != nil || updated.Latitude != nil || updated.AddressLine1 != "Jl. Asia Afrika No. 102" {
			t.Fatalf("expected nullable columns cleared cleanly on update, got %+v err=%v", updated, err)
		}

		ref, err := reg.Addresses.CheckReferenced(ctx, addr1.ID)
		if err != nil || ref {
			t.Fatalf("expected unreferenced address, ref=%v err=%v", ref, err)
		}

		if err := reg.Addresses.DeleteByID(ctx, addr1.ID); err != nil {
			t.Fatalf("DeleteByID failed: %v", err)
		}
		if _, err := reg.Addresses.SelectByID(ctx, addr1.ID); !errors.Is(err, eferrors.ErrAddressNotFound) {
			t.Fatalf("expected ErrAddressNotFound after delete, got %v", err)
		}
	})

	t.Run("LegalEntity, Location, and OrganizationUnit CRUD + hierarchy + constraints", func(t *testing.T) {
		truncateFoundationTables(t, db)
		now := time.Now().Truncate(time.Microsecond)
		incDate := time.Date(2022, 3, 15, 0, 0, 0, 0, time.UTC)

		regAddr, _ := entity.NewAddress(mustUUID(t), entity.AddressParams{
			AddressLine1: "Jl. Sudirman Kav 1",
			City:         "Jakarta Selatan",
			CountryCode:  "ID",
		}, now)
		if err := reg.Addresses.Insert(ctx, regAddr); err != nil {
			t.Fatalf("Insert regAddr: %v", err)
		}

		holding, err := entity.NewLegalEntity(mustUUID(t), entity.LegalEntityParams{
			EntityCode:               "ENT-HOLDING",
			LegalName:                "PT IntegriOne Holding",
			IsPrimaryOperatingEntity: false,
			Status:                   entity.LegalEntityStatusActive,
			JurisdictionCountryCode:  "ID",
			TaxIDNumber:              strPtr("01.111.111.1-011.000"),
			NIB:                      strPtr("1234567890001"),
			RegisteredAddressID:      &regAddr.ID,
			FiscalYearStartMonth:     1,
			IncorporationDate:        &incDate,
		}, now)
		if err != nil {
			t.Fatalf("NewLegalEntity holding: %v", err)
		}
		primary, err := entity.NewLegalEntity(mustUUID(t), entity.LegalEntityParams{
			ParentLegalEntityID:      &holding.ID,
			EntityCode:               "ENT-RETAIL",
			LegalName:                "PT IntegriOne Ritel Nusantara",
			TradingName:              strPtr("IntegriOne"),
			IsPrimaryOperatingEntity: true,
			Status:                   entity.LegalEntityStatusActive,
			JurisdictionCountryCode:  "ID",
			TaxIDNumber:              strPtr("01.222.222.2-012.000"),
			NIB:                      strPtr("1234567890001"),
			IsPKP:                    true,
			FiscalYearStartMonth:     1,
		}, now)
		if err != nil {
			t.Fatalf("NewLegalEntity primary: %v", err)
		}
		if err := reg.LegalEntities.Insert(ctx, holding); err != nil {
			t.Fatalf("Insert holding: %v", err)
		}
		if err := reg.LegalEntities.Insert(ctx, primary); err != nil {
			t.Fatalf("Insert primary: %v", err)
		}

		// Verifikasi deteksi pelanggaran constraint unik SQLSTATE 23505.
		dupPrimary, _ := entity.NewLegalEntity(mustUUID(t), entity.LegalEntityParams{
			EntityCode:               "ENT-DUP-PRIM",
			LegalName:                "PT Konflik Utama",
			IsPrimaryOperatingEntity: true,
			Status:                   entity.LegalEntityStatusActive,
			JurisdictionCountryCode:  "ID",
			FiscalYearStartMonth:     1,
		}, now)
		if err := reg.LegalEntities.Insert(ctx, dupPrimary); !errors.Is(err, eferrors.ErrPrimaryLegalEntityExists) {
			t.Fatalf("expected ErrPrimaryLegalEntityExists, got %v", err)
		}

		// Verifikasi FindBy/SelectBy bisnis pada LegalEntity.
		if gotID, err := reg.LegalEntities.SelectByID(ctx, holding.ID); err != nil || gotID.ID != holding.ID {
			t.Fatalf("SelectByID holding failed: got=%v err=%v", gotID, err)
		}
		if batchLE, err := reg.LegalEntities.SelectByIDs(ctx, []string{holding.ID, primary.ID}); err != nil || len(batchLE) != 2 {
			t.Fatalf("LegalEntities.SelectByIDs failed: len=%d err=%v", len(batchLE), err)
		}
		primOnly := true
		if listLE, err := reg.LegalEntities.SelectList(ctx, domainrepo.LegalEntityFilter{
			PrimaryOnly: &primOnly,
			Query:       "Integrione",
		}); err != nil || listLE.Total != 1 || listLE.Items[0].ID != primary.ID {
			t.Fatalf("LegalEntities.SelectList failed: page=%+v err=%v", listLE, err)
		}
		if gotPrim, err := reg.LegalEntities.SelectPrimaryOperating(ctx); err != nil || gotPrim.ID != primary.ID {
			t.Fatalf("SelectPrimaryOperating failed: got=%v err=%v", gotPrim, err)
		}
		if gotTax, err := reg.LegalEntities.SelectByTaxIDNumber(ctx, "01.222.222.2-012.000"); err != nil || gotTax.ID != primary.ID {
			t.Fatalf("SelectByTaxIDNumber failed: got=%v err=%v", gotTax, err)
		}
		if byNIB, err := reg.LegalEntities.SelectByNIB(ctx, "1234567890001", domainrepo.PageRequest{}); err != nil || len(byNIB.Items) != 2 || byNIB.Total != 2 || byNIB.HasMore {
			t.Fatalf("SelectByNIB failed: page=%+v err=%v", byNIB, err)
		}
		if byNIBP1, err := reg.LegalEntities.SelectByNIB(ctx, "1234567890001", domainrepo.PageRequest{Limit: 1, Offset: 0}); err != nil || len(byNIBP1.Items) != 1 || byNIBP1.Total != 2 || !byNIBP1.HasMore {
			t.Fatalf("SelectByNIB page 1 failed: page=%+v err=%v", byNIBP1, err)
		} else if byNIBP2, err := reg.LegalEntities.SelectByNIB(ctx, "1234567890001", domainrepo.PageRequest{Limit: 1, Offset: 1}); err != nil || len(byNIBP2.Items) != 1 || byNIBP2.Total != 2 || byNIBP2.HasMore || byNIBP2.Items[0].ID == byNIBP1.Items[0].ID {
			t.Fatalf("SelectByNIB page 2 failed: page=%+v p1=%+v err=%v", byNIBP2, byNIBP1, err)
		}
		if gotCode, err := reg.LegalEntities.SelectByCode(ctx, "  ent-retail "); err != nil || gotCode.ID != primary.ID {
			t.Fatalf("SelectByCode failed: got=%v err=%v", gotCode, err)
		}
		if parentID, err := reg.LegalEntities.SelectParentID(ctx, primary.ID); err != nil || parentID == nil || *parentID != holding.ID {
			t.Fatalf("SelectParentID failed: parentID=%v err=%v", parentID, err)
		}
		if children, err := reg.LegalEntities.SelectChildrenByParentID(ctx, holding.ID, domainrepo.PageRequest{}); err != nil || len(children.Items) != 1 || children.Total != 1 || children.HasMore {
			t.Fatalf("SelectChildrenByParentID failed: page=%+v err=%v", children, err)
		}
		if exists, err := reg.LegalEntities.CheckCodeExists(ctx, "ent-retail", primary.ID); err != nil || exists {
			t.Fatalf("CheckCodeExists with self exclude should be false, got (%v, %v)", exists, err)
		}
		if exists, err := reg.LegalEntities.CheckTaxIDExists(ctx, "01.222.222.2-012.000", ""); err != nil || !exists {
			t.Fatalf("CheckTaxIDExists should be true, got (%v, %v)", exists, err)
		}

		// Lokasi internal + JSONB operational_attributes + FK RESTRICT ke Address.
		locAddr, _ := entity.NewAddress(mustUUID(t), entity.AddressParams{
			AddressLine1: "Kawasan Pergudangan Cakung Blok A2",
			City:         "Jakarta Timur",
			CountryCode:  "ID",
		}, now)
		if err := reg.Addresses.Insert(ctx, locAddr); err != nil {
			t.Fatalf("Insert locAddr: %v", err)
		}
		hqLoc, _ := entity.NewLocation(mustUUID(t), entity.LocationParams{
			LegalEntityID:         primary.ID,
			LocationCode:          "HQ-JKT",
			Name:                  "Kantor Pusat IntegriOne",
			Kind:                  entity.LocationKindHeadquarters,
			Timezone:              "Asia/Jakarta",
			OperationalAttributes: json.RawMessage(`{"floors":3}`),
		}, now)
		whLoc, _ := entity.NewLocation(mustUUID(t), entity.LocationParams{
			LegalEntityID:         primary.ID,
			ParentLocationID:      &hqLoc.ID,
			AddressID:             &locAddr.ID,
			LocationCode:          "WH-CKG",
			Name:                  "Gudang Utama Cakung",
			Kind:                  entity.LocationKindWarehouse,
			Phone:                 strPtr("+6221-889900"),
			Timezone:              "Asia/Jakarta",
			OperationalAttributes: json.RawMessage(`{"dock_bays":4,"cold_storage":false}`),
		}, now)
		if err := reg.Locations.Insert(ctx, hqLoc); err != nil {
			t.Fatalf("Insert hqLoc: %v", err)
		}
		if err := reg.Locations.Insert(ctx, whLoc); err != nil {
			t.Fatalf("Insert whLoc: %v", err)
		}

		// Pastikan Address tidak dapat dihapus saat dirujuk location (RESTRICT -> ErrAddressInUse).
		if ref, err := reg.Addresses.CheckReferenced(ctx, locAddr.ID); err != nil || !ref {
			t.Fatalf("expected CheckReferenced=true for locAddr, got (%v, %v)", ref, err)
		}
		if err := reg.Addresses.DeleteByID(ctx, locAddr.ID); !errors.Is(err, eferrors.ErrAddressInUse) {
			t.Fatalf("expected ErrAddressInUse when referenced by location, got %v", err)
		}

		if byKind, err := reg.Locations.SelectByLegalEntityAndKind(ctx, primary.ID, entity.LocationKindWarehouse, true, domainrepo.PageRequest{}); err != nil || len(byKind.Items) != 1 || byKind.Total != 1 || byKind.HasMore || byKind.Items[0].ID != whLoc.ID {
			t.Fatalf("SelectByLegalEntityAndKind failed: page=%+v err=%v", byKind, err)
		}
		if byAddr, err := reg.Locations.SelectByAddressID(ctx, locAddr.ID, domainrepo.PageRequest{}); err != nil || len(byAddr.Items) != 1 || byAddr.Total != 1 || byAddr.HasMore || byAddr.Items[0].ID != whLoc.ID {
			t.Fatalf("SelectByAddressID failed: page=%+v err=%v", byAddr, err)
		}
		if gotLoc, err := reg.Locations.SelectByID(ctx, whLoc.ID); err != nil || gotLoc.ID != whLoc.ID {
			t.Fatalf("Locations.SelectByID failed: got=%v err=%v", gotLoc, err)
		}
		if batchLoc, err := reg.Locations.SelectByIDs(ctx, []string{hqLoc.ID, whLoc.ID}); err != nil || len(batchLoc) != 2 {
			t.Fatalf("Locations.SelectByIDs failed: len=%d err=%v", len(batchLoc), err)
		}
		if byCodeLoc, err := reg.Locations.SelectByCode(ctx, "wh-ckg"); err != nil || byCodeLoc.ID != whLoc.ID {
			t.Fatalf("Locations.SelectByCode failed: got=%v err=%v", byCodeLoc, err)
		}
		if locParent, err := reg.Locations.SelectParentID(ctx, whLoc.ID); err != nil || locParent == nil || *locParent != hqLoc.ID {
			t.Fatalf("Locations.SelectParentID failed: got=%v err=%v", locParent, err)
		}
		if locChildren, err := reg.Locations.SelectChildrenByParentID(ctx, hqLoc.ID, domainrepo.PageRequest{}); err != nil || len(locChildren.Items) != 1 || locChildren.Total != 1 || locChildren.HasMore || locChildren.Items[0].ID != whLoc.ID {
			t.Fatalf("Locations.SelectChildrenByParentID failed: page=%+v err=%v", locChildren, err)
		}
		if locExists, err := reg.Locations.CheckCodeExists(ctx, "wh-ckg", whLoc.ID); err != nil || locExists {
			t.Fatalf("Locations.CheckCodeExists self-exclude failed: got=(%v, %v)", locExists, err)
		}
		if err := whLoc.UpdateDetails(entity.LocationUpdateParams{
			ParentLocationID:      &hqLoc.ID,
			AddressID:             &locAddr.ID,
			Name:                  "Gudang Utama Cakung Baru",
			Kind:                  entity.LocationKindWarehouse,
			Timezone:              "Asia/Jakarta",
			OperationalAttributes: json.RawMessage(`{"dock_bays":6}`),
		}, now.Add(30*time.Second)); err != nil {
			t.Fatalf("whLoc.UpdateDetails failed: %v", err)
		}
		if err := reg.Locations.Update(ctx, whLoc); err != nil {
			t.Fatalf("Locations.Update failed: %v", err)
		}

		// Unit Organisasi + HeadPersonID HRM + PrimaryLocationID.
		headWorkerID := mustUUID(t)
		divOps, _ := entity.NewOrganizationUnit(mustUUID(t), entity.OrganizationUnitParams{
			LegalEntityID:     primary.ID,
			PrimaryLocationID: &hqLoc.ID,
			UnitCode:          "DIV-OPS",
			Name:              "Divisi Operasional",
			Kind:              entity.OrgUnitKindDivision,
			CostCenterCode:    strPtr("CC-OPS"),
			HeadPersonID:      &headWorkerID,
		}, now)
		deptWh, _ := entity.NewOrganizationUnit(mustUUID(t), entity.OrganizationUnitParams{
			LegalEntityID:     primary.ID,
			ParentUnitID:      &divOps.ID,
			PrimaryLocationID: &whLoc.ID,
			UnitCode:          "DEPT-WH",
			Name:              "Departemen Gudang & Logistik",
			Kind:              entity.OrgUnitKindDepartment,
			HeadPersonID:      &headWorkerID,
		}, now)
		if err := reg.OrganizationUnits.Insert(ctx, divOps); err != nil {
			t.Fatalf("Insert divOps: %v", err)
		}
		if err := reg.OrganizationUnits.Insert(ctx, deptWh); err != nil {
			t.Fatalf("Insert deptWh: %v", err)
		}

		if byHead, err := reg.OrganizationUnits.SelectByHeadPersonID(ctx, headWorkerID, domainrepo.PageRequest{}); err != nil || len(byHead.Items) != 2 || byHead.Total != 2 || byHead.HasMore {
			t.Fatalf("SelectByHeadPersonID failed: page=%+v err=%v", byHead, err)
		}
		if byHeadP1, err := reg.OrganizationUnits.SelectByHeadPersonID(ctx, headWorkerID, domainrepo.PageRequest{Limit: 1, Offset: 0}); err != nil || len(byHeadP1.Items) != 1 || byHeadP1.Total != 2 || !byHeadP1.HasMore {
			t.Fatalf("SelectByHeadPersonID page 1 failed: page=%+v err=%v", byHeadP1, err)
		} else if byHeadP2, err := reg.OrganizationUnits.SelectByHeadPersonID(ctx, headWorkerID, domainrepo.PageRequest{Limit: 1, Offset: 1}); err != nil || len(byHeadP2.Items) != 1 || byHeadP2.Total != 2 || byHeadP2.HasMore || byHeadP2.Items[0].ID == byHeadP1.Items[0].ID {
			t.Fatalf("SelectByHeadPersonID page 2 failed: page=%+v p1=%+v err=%v", byHeadP2, byHeadP1, err)
		}
		if byLoc, err := reg.OrganizationUnits.SelectByPrimaryLocationID(ctx, whLoc.ID, domainrepo.PageRequest{}); err != nil || len(byLoc.Items) != 1 || byLoc.Total != 1 || byLoc.HasMore || byLoc.Items[0].ID != deptWh.ID {
			t.Fatalf("SelectByPrimaryLocationID failed: page=%+v err=%v", byLoc, err)
		}
		if gotUnit, err := reg.OrganizationUnits.SelectByID(ctx, deptWh.ID); err != nil || gotUnit.ID != deptWh.ID {
			t.Fatalf("OrganizationUnits.SelectByID failed: got=%v err=%v", gotUnit, err)
		}
		if batchUnits, err := reg.OrganizationUnits.SelectByIDs(ctx, []string{divOps.ID, deptWh.ID}); err != nil || len(batchUnits) != 2 {
			t.Fatalf("OrganizationUnits.SelectByIDs failed: len=%d err=%v", len(batchUnits), err)
		}
		if byCodeUnit, err := reg.OrganizationUnits.SelectByEntityAndCode(ctx, primary.ID, "dept-wh"); err != nil || byCodeUnit.ID != deptWh.ID {
			t.Fatalf("OrganizationUnits.SelectByEntityAndCode failed: got=%v err=%v", byCodeUnit, err)
		}
		if unitParent, err := reg.OrganizationUnits.SelectParentID(ctx, deptWh.ID); err != nil || unitParent == nil || *unitParent != divOps.ID {
			t.Fatalf("OrganizationUnits.SelectParentID failed: got=%v err=%v", unitParent, err)
		}
		if unitChildren, err := reg.OrganizationUnits.SelectChildrenByParentID(ctx, divOps.ID, domainrepo.PageRequest{}); err != nil || len(unitChildren.Items) != 1 || unitChildren.Total != 1 || unitChildren.HasMore || unitChildren.Items[0].ID != deptWh.ID {
			t.Fatalf("OrganizationUnits.SelectChildrenByParentID failed: page=%+v err=%v", unitChildren, err)
		}
		if unitExists, err := reg.OrganizationUnits.CheckCodeExistsInEntity(ctx, primary.ID, "dept-wh", ""); err != nil || !unitExists {
			t.Fatalf("OrganizationUnits.CheckCodeExistsInEntity failed: got=(%v, %v)", unitExists, err)
		}
		if divDeps, err := reg.OrganizationUnits.CountActiveDependents(ctx, divOps.ID); err != nil || divDeps != 1 {
			t.Fatalf("OrganizationUnits.CountActiveDependents want=1, got=%d err=%v", divDeps, err)
		}
		if err := deptWh.UpdateDetails(entity.OrganizationUnitUpdateParams{
			ParentUnitID:      &divOps.ID,
			PrimaryLocationID: &whLoc.ID,
			Name:              "Departemen Gudang & Distribusi",
			Kind:              entity.OrgUnitKindDepartment,
			HeadPersonID:      &headWorkerID,
		}, now.Add(30*time.Second)); err != nil {
			t.Fatalf("deptWh.UpdateDetails failed: %v", err)
		}
		if err := reg.OrganizationUnits.Update(ctx, deptWh); err != nil {
			t.Fatalf("OrganizationUnits.Update failed: %v", err)
		}
		if deps, err := reg.Locations.CountActiveDependents(ctx, whLoc.ID); err != nil || deps != 1 {
			t.Fatalf("Locations.CountActiveDependents want=1, got=%d err=%v", deps, err)
		}
		if deps, err := reg.LegalEntities.CountActiveDependents(ctx, primary.ID); err != nil || deps != 4 {
			t.Fatalf("LegalEntities.CountActiveDependents want=4 (2 loc + 2 orgUnit), got=%d err=%v", deps, err)
		}

		// Switch primary operating entity lalu soft delete entitas lama.
		if err := reg.LegalEntities.UpdateClearPrimaryOperatingExcept(ctx, holding.ID, now.Add(time.Minute)); err != nil {
			t.Fatalf("UpdateClearPrimaryOperatingExcept failed: %v", err)
		}
		if err := holding.SetPrimaryOperatingEntity(true, now.Add(time.Minute)); err != nil {
			t.Fatalf("holding.SetPrimaryOperatingEntity failed: %v", err)
		}
		if err := reg.LegalEntities.Update(ctx, holding); err != nil {
			t.Fatalf("Update holding primary: %v", err)
		}
		if gotPrim, err := reg.LegalEntities.SelectPrimaryOperating(ctx); err != nil || gotPrim.ID != holding.ID {
			t.Fatalf("expected holding as new primary, got=%v err=%v", gotPrim, err)
		}

		if err := reg.OrganizationUnits.SoftDeleteByID(ctx, deptWh.ID, now.Add(2*time.Minute)); err != nil {
			t.Fatalf("OrgUnit SoftDeleteByID failed: %v", err)
		}
		if _, err := reg.OrganizationUnits.SelectByID(ctx, deptWh.ID); !errors.Is(err, eferrors.ErrOrgUnitNotFound) {
			t.Fatalf("expected ErrOrgUnitNotFound after soft delete, got %v", err)
		}
		if page, err := reg.OrganizationUnits.SelectList(ctx, domainrepo.OrganizationUnitFilter{
			LegalEntityID:  primary.ID,
			IncludeDeleted: true,
		}); err != nil || page.Total != 2 {
			t.Fatalf("SelectList IncludeDeleted failed: page=%+v err=%v", page, err)
		}

		if err := reg.Locations.SoftDeleteByID(ctx, whLoc.ID, now.Add(2*time.Minute)); err != nil {
			t.Fatalf("Location SoftDeleteByID failed: %v", err)
		}
		if locRows, err := reg.Locations.SelectList(ctx, domainrepo.LocationFilter{
			LegalEntityID:  primary.ID,
			IncludeDeleted: true,
		}); err != nil || locRows.Total != 2 {
			t.Fatalf("Locations.SelectList IncludeDeleted failed: page=%+v err=%v", locRows, err)
		}

		// Soft-delete entitas non-utama setelah status primary dipindahkan ke holding.
		if err := reg.LegalEntities.SoftDeleteByID(ctx, primary.ID, now.Add(3*time.Minute)); err != nil {
			t.Fatalf("LegalEntities.SoftDeleteByID failed: %v", err)
		}
		if _, err := reg.LegalEntities.SelectByID(ctx, primary.ID); !errors.Is(err, eferrors.ErrLegalEntityNotFound) {
			t.Fatalf("expected ErrLegalEntityNotFound after soft delete, got %v", err)
		}
	})

	t.Run("Party, PartyRole, PartyContact (IAM link + SoftDelete revoke), and PartyAddress", func(t *testing.T) {
		truncateFoundationTables(t, db)
		now := time.Now().Truncate(time.Microsecond)

		supplier, err := entity.NewParty(mustUUID(t), entity.PartyParams{
			PartnerCode:       "SUP-KAIN-01",
			Kind:              entity.PartyKindOrganization,
			DisplayName:       "CV Tekstil Nusantara",
			LegalName:         strPtr("CV Tekstil Nusantara Jaya"),
			Email:             strPtr("sales@tekstil.co.id"),
			Phone:             strPtr("+6222-778899"),
			TaxIDNumber:       strPtr("02.333.444.5-422.000"),
			IsPKP:             true,
			PreferredLanguage: "id",
			Status:            entity.PartyStatusActive,
		}, now)
		if err != nil {
			t.Fatalf("NewParty failed: %v", err)
		}
		if err := reg.Parties.Insert(ctx, supplier); err != nil {
			t.Fatalf("Insert supplier: %v", err)
		}

		roleSupp, _ := entity.NewPartyRole(mustUUID(t), supplier.ID, entity.PartyRoleKindSupplier, strPtr("Main fabric mill"), now)
		roleVendor, _ := entity.NewPartyRole(mustUUID(t), supplier.ID, entity.PartyRoleKindVendor, nil, now)
		if err := reg.PartyRoles.Insert(ctx, roleSupp); err != nil {
			t.Fatalf("Insert roleSupp: %v", err)
		}
		if err := reg.PartyRoles.Insert(ctx, roleVendor); err != nil {
			t.Fatalf("Insert roleVendor: %v", err)
		}
		dupRoleSupp, _ := entity.NewPartyRole(mustUUID(t), supplier.ID, entity.PartyRoleKindSupplier, nil, now)
		if err := reg.PartyRoles.Insert(ctx, dupRoleSupp); !errors.Is(err, eferrors.ErrPartyRoleExists) {
			t.Fatalf("expected ErrPartyRoleExists on duplicate role, got %v", err)
		}
		if err := reg.PartyRoles.DeleteByID(ctx, roleVendor.ID); err != nil {
			t.Fatalf("PartyRoles.DeleteByID failed: %v", err)
		}
		if exists, err := reg.PartyRoles.CheckExists(ctx, supplier.ID, entity.PartyRoleKindVendor); err != nil || exists {
			t.Fatalf("expected deleted role CheckExists=false, got (%v, %v)", exists, err)
		}

		// Verifikasi SelectList Party dengan filter Role berbasis EXISTS dan SelectBy* bisnis.
		supRole := entity.PartyRoleKindSupplier
		if page, err := reg.Parties.SelectList(ctx, domainrepo.PartyFilter{
			Role:  &supRole,
			Query: "tekstil",
		}); err != nil || page.Total != 1 || page.Items[0].ID != supplier.ID {
			t.Fatalf("Parties.SelectList by Role failed: page=%+v err=%v", page, err)
		}
		if byTax, err := reg.Parties.SelectByTaxIDNumber(ctx, "02.333.444.5-422.000"); err != nil || byTax.ID != supplier.ID {
			t.Fatalf("Parties.SelectByTaxIDNumber failed: got=%v err=%v", byTax, err)
		}
		if byEmail, err := reg.Parties.SelectByEmail(ctx, "SALES@tekstil.co.id", domainrepo.PageRequest{}); err != nil || len(byEmail.Items) != 1 || byEmail.Total != 1 || byEmail.HasMore || byEmail.Items[0].ID != supplier.ID {
			t.Fatalf("Parties.SelectByEmail failed: page=%+v err=%v", byEmail, err)
		}
		if rolesPage, err := reg.PartyRoles.SelectByRoleKind(ctx, entity.PartyRoleKindSupplier, true, domainrepo.PageRequest{}); err != nil || rolesPage.Total != 1 {
			t.Fatalf("PartyRoles.SelectByRoleKind failed: page=%+v err=%v", rolesPage, err)
		}
		if batchParties, err := reg.Parties.SelectByIDs(ctx, []string{supplier.ID}); err != nil || len(batchParties) != 1 {
			t.Fatalf("Parties.SelectByIDs failed: len=%d err=%v", len(batchParties), err)
		}
		if byCodeParty, err := reg.Parties.SelectByCode(ctx, "sup-kain-01"); err != nil || byCodeParty.ID != supplier.ID {
			t.Fatalf("Parties.SelectByCode failed: got=%v err=%v", byCodeParty, err)
		}
		if codeExists, err := reg.Parties.CheckCodeExists(ctx, "sup-kain-01", ""); err != nil || !codeExists {
			t.Fatalf("Parties.CheckCodeExists failed: got=(%v, %v)", codeExists, err)
		}
		if taxExists, err := reg.Parties.CheckTaxIDExists(ctx, "02.333.444.5-422.000", supplier.ID); err != nil || taxExists {
			t.Fatalf("Parties.CheckTaxIDExists self-exclude failed: got=(%v, %v)", taxExists, err)
		}
		if err := supplier.UpdateProfile(entity.PartyProfileParams{
			Kind:              entity.PartyKindOrganization,
			DisplayName:       "CV Tekstil Nusantara (Updated)",
			LegalName:         strPtr("CV Tekstil Nusantara Jaya"),
			Email:             strPtr("sales@tekstil.co.id"),
			Phone:             strPtr("+6222-778899"),
			PreferredLanguage: "id",
		}, now.Add(20*time.Second)); err != nil {
			t.Fatalf("supplier.UpdateProfile failed: %v", err)
		}
		if err := reg.Parties.Update(ctx, supplier); err != nil {
			t.Fatalf("Parties.Update failed: %v", err)
		}
		if gotRole, err := reg.PartyRoles.SelectByID(ctx, roleSupp.ID); err != nil || gotRole.ID != roleSupp.ID {
			t.Fatalf("PartyRoles.SelectByID failed: got=%v err=%v", gotRole, err)
		}
		if partyRoles, err := reg.PartyRoles.SelectByPartyID(ctx, supplier.ID, domainrepo.PageRequest{}); err != nil || len(partyRoles.Items) != 1 || partyRoles.Total != 1 || partyRoles.HasMore {
			t.Fatalf("PartyRoles.SelectByPartyID failed: page=%+v err=%v", partyRoles, err)
		}
		roleKindSupp := entity.PartyRoleKindSupplier
		if rolesList, err := reg.PartyRoles.SelectList(ctx, domainrepo.PartyRoleFilter{
			PartyID: supplier.ID,
			Role:    &roleKindSupp,
			Query:   "Mill",
			Page:    domainrepo.PageRequest{Limit: 10, Offset: 0},
		}); err != nil || rolesList.Total != 1 || len(rolesList.Items) != 1 || rolesList.Items[0].ID != roleSupp.ID {
			t.Fatalf("PartyRoles.SelectList failed: page=%+v err=%v", rolesList, err)
		}
		if byPairRole, err := reg.PartyRoles.SelectByPartyAndRole(ctx, supplier.ID, entity.PartyRoleKindSupplier); err != nil || byPairRole.ID != roleSupp.ID {
			t.Fatalf("PartyRoles.SelectByPartyAndRole failed: got=%v err=%v", byPairRole, err)
		}
		_ = roleSupp.UpdateReferenceNote(strPtr("Updated fabric mill"), now.Add(25*time.Second))
		if err := reg.PartyRoles.Update(ctx, roleSupp); err != nil {
			t.Fatalf("PartyRoles.Update failed: %v", err)
		}

		// PartyContacts: PIC dengan tautan portal IAM + kontak direktori non-portal.
		iamUserID := mustUUID(t)
		picPortal, _ := entity.NewPartyContact(mustUUID(t), entity.PartyContactParams{
			PartyID:            supplier.ID,
			IAMUserID:          &iamUserID,
			FullName:           "Budi Santoso",
			JobTitle:           strPtr("Account Manager"),
			ContactRole:        "portal_admin",
			Email:              strPtr("budi@tekstil.co.id"),
			PortalAccessStatus: entity.PortalAccessStatusActive,
			IsPrimary:          true,
		}, now)
		picFinance, _ := entity.NewPartyContact(mustUUID(t), entity.PartyContactParams{
			PartyID:            supplier.ID,
			FullName:           "Sari Finance",
			ContactRole:        "billing",
			Phone:              strPtr("+628123456789"),
			PortalAccessStatus: entity.PortalAccessStatusNone,
			IsPrimary:          false,
		}, now)
		if err := reg.PartyContacts.Insert(ctx, picPortal); err != nil {
			t.Fatalf("Insert picPortal: %v", err)
		}
		if err := reg.PartyContacts.Insert(ctx, picFinance); err != nil {
			t.Fatalf("Insert picFinance: %v", err)
		}

		if byIAM, err := reg.PartyContacts.SelectByIAMUserID(ctx, iamUserID, domainrepo.PageRequest{}); err != nil || len(byIAM.Items) != 1 || byIAM.Total != 1 || byIAM.Items[0].ID != picPortal.ID {
			t.Fatalf("SelectByIAMUserID failed: page=%+v err=%v", byIAM, err)
		}
		if byPair, err := reg.PartyContacts.SelectByPartyAndIAMUserID(ctx, supplier.ID, iamUserID); err != nil || byPair.ID != picPortal.ID {
			t.Fatalf("SelectByPartyAndIAMUserID failed: got=%v err=%v", byPair, err)
		}
		if primPIC, err := reg.PartyContacts.SelectPrimaryByPartyID(ctx, supplier.ID); err != nil || primPIC.ID != picPortal.ID {
			t.Fatalf("SelectPrimaryByPartyID failed: got=%v err=%v", primPIC, err)
		}
		if byEmail, err := reg.PartyContacts.SelectByEmail(ctx, "BUDI@TEKSTIL.CO.ID", domainrepo.PageRequest{}); err != nil || len(byEmail.Items) != 1 || byEmail.Total != 1 {
			t.Fatalf("PartyContacts.SelectByEmail failed: page=%+v err=%v", byEmail, err)
		}
		if gotPIC, err := reg.PartyContacts.SelectByID(ctx, picPortal.ID); err != nil || gotPIC.ID != picPortal.ID {
			t.Fatalf("PartyContacts.SelectByID failed: got=%v err=%v", gotPIC, err)
		}
		if byPartyPICs, err := reg.PartyContacts.SelectByPartyID(ctx, supplier.ID, domainrepo.PageRequest{}); err != nil || len(byPartyPICs.Items) != 2 || byPartyPICs.Total != 2 || byPartyPICs.HasMore {
			t.Fatalf("PartyContacts.SelectByPartyID failed: page=%+v err=%v", byPartyPICs, err)
		}
		if picP1, err := reg.PartyContacts.SelectByPartyID(ctx, supplier.ID, domainrepo.PageRequest{Limit: 1, Offset: 0}); err != nil || len(picP1.Items) != 1 || picP1.Total != 2 || !picP1.HasMore {
			t.Fatalf("PartyContacts.SelectByPartyID page 1 failed: page=%+v err=%v", picP1, err)
		} else if picP2, err := reg.PartyContacts.SelectByPartyID(ctx, supplier.ID, domainrepo.PageRequest{Limit: 1, Offset: 1}); err != nil || len(picP2.Items) != 1 || picP2.Total != 2 || picP2.HasMore || picP2.Items[0].ID == picP1.Items[0].ID {
			t.Fatalf("PartyContacts.SelectByPartyID page 2 failed: page=%+v p1=%+v err=%v", picP2, picP1, err)
		}
		if iamDup, err := reg.PartyContacts.CheckIAMUserExistsInParty(ctx, supplier.ID, iamUserID, ""); err != nil || !iamDup {
			t.Fatalf("PartyContacts.CheckIAMUserExistsInParty failed: got=(%v, %v)", iamDup, err)
		}

		// Tukar kontak utama ke picFinance.
		if err := reg.PartyContacts.UpdateClearPrimaryByPartyExcept(ctx, supplier.ID, picFinance.ID, now.Add(time.Minute)); err != nil {
			t.Fatalf("UpdateClearPrimaryByPartyExcept failed: %v", err)
		}
		picFinance.IsPrimary = true
		picFinance.UpdatedAt = now.Add(time.Minute)
		if err := reg.PartyContacts.Update(ctx, picFinance); err != nil {
			t.Fatalf("PartyContacts.Update failed: %v", err)
		}
		if newPrim, err := reg.PartyContacts.SelectPrimaryByPartyID(ctx, supplier.ID); err != nil || newPrim.ID != picFinance.ID {
			t.Fatalf("expected picFinance as primary contact, got=%v err=%v", newPrim, err)
		}

		// SoftDeleteByID pada picPortal wajib mencabut portal_access_status menjadi revoked dan is_active=false.
		if err := reg.PartyContacts.SoftDeleteByID(ctx, picPortal.ID, now.Add(2*time.Minute)); err != nil {
			t.Fatalf("PartyContacts.SoftDeleteByID failed: %v", err)
		}
		deletedContacts, err := reg.PartyContacts.SelectList(ctx, domainrepo.PartyContactFilter{
			PartyID:        supplier.ID,
			IncludeDeleted: true,
		})
		if err != nil || deletedContacts.Total != 2 {
			t.Fatalf("SelectList IncludeDeleted failed: page=%+v err=%v", deletedContacts, err)
		}
		for _, c := range deletedContacts.Items {
			if c.ID == picPortal.ID {
				if c.IsActive || c.DeletedAt == nil || c.PortalAccessStatus != entity.PortalAccessStatusRevoked {
					t.Fatalf("expected soft-deleted portal contact revoked & inactive, got %+v", c)
				}
			}
		}

		// PartyAddresses: billing primary + shipping conflict + clear primary.
		addrBill, _ := entity.NewAddress(mustUUID(t), entity.AddressParams{
			AddressLine1: "Jl. Gatot Subroto No. 12",
			City:         "Bandung",
			CountryCode:  "ID",
		}, now)
		addrShip, _ := entity.NewAddress(mustUUID(t), entity.AddressParams{
			AddressLine1: "Jl. Soekarno Hatta No. 88",
			City:         "Bandung",
			CountryCode:  "ID",
		}, now)
		_ = reg.Addresses.Insert(ctx, addrBill)
		_ = reg.Addresses.Insert(ctx, addrShip)

		pa1, _ := entity.NewPartyAddress(mustUUID(t), supplier.ID, addrBill.ID, entity.AddressUsageKindBilling, true, now)
		pa2, _ := entity.NewPartyAddress(mustUUID(t), supplier.ID, addrShip.ID, entity.AddressUsageKindBilling, true, now)
		if err := reg.PartyAddresses.Insert(ctx, pa1); err != nil {
			t.Fatalf("Insert pa1: %v", err)
		}
		if err := reg.PartyAddresses.Insert(ctx, pa2); !errors.Is(err, eferrors.ErrPrimaryPartyAddressConflict) {
			t.Fatalf("expected ErrPrimaryPartyAddressConflict, got %v", err)
		}

		pa2.IsPrimary = false
		if err := reg.PartyAddresses.Insert(ctx, pa2); err != nil {
			t.Fatalf("Insert secondary billing pa2: %v", err)
		}
		if primBill, err := reg.PartyAddresses.SelectPrimaryByPartyAndUsage(ctx, supplier.ID, entity.AddressUsageKindBilling); err != nil || primBill.ID != pa1.ID {
			t.Fatalf("SelectPrimaryByPartyAndUsage failed: got=%v err=%v", primBill, err)
		}
		if byAddr, err := reg.PartyAddresses.SelectByAddressID(ctx, addrBill.ID, domainrepo.PageRequest{}); err != nil || len(byAddr.Items) != 1 || byAddr.Total != 1 || byAddr.HasMore {
			t.Fatalf("PartyAddresses.SelectByAddressID failed: page=%+v err=%v", byAddr, err)
		}

		if err := reg.PartyAddresses.UpdateClearPrimaryByPartyUsageExcept(ctx, supplier.ID, entity.AddressUsageKindBilling, pa2.ID, now.Add(time.Minute)); err != nil {
			t.Fatalf("UpdateClearPrimaryByPartyUsageExcept failed: %v", err)
		}
		pa2.IsPrimary = true
		if err := reg.PartyAddresses.Update(ctx, pa2); err != nil {
			t.Fatalf("PartyAddresses.Update failed: %v", err)
		}
		if newPrimBill, err := reg.PartyAddresses.SelectPrimaryByPartyAndUsage(ctx, supplier.ID, entity.AddressUsageKindBilling); err != nil || newPrimBill.ID != pa2.ID {
			t.Fatalf("expected pa2 as new primary billing, got=%v err=%v", newPrimBill, err)
		}
		if gotPA, err := reg.PartyAddresses.SelectByID(ctx, pa1.ID); err != nil || gotPA.ID != pa1.ID {
			t.Fatalf("PartyAddresses.SelectByID failed: got=%v err=%v", gotPA, err)
		}
		if listPA, err := reg.PartyAddresses.SelectByPartyID(ctx, supplier.ID, domainrepo.PageRequest{}); err != nil || len(listPA.Items) != 2 || listPA.Total != 2 || listPA.HasMore {
			t.Fatalf("PartyAddresses.SelectByPartyID failed: page=%+v err=%v", listPA, err)
		}
		if paP1, err := reg.PartyAddresses.SelectByPartyID(ctx, supplier.ID, domainrepo.PageRequest{Limit: 1, Offset: 0}); err != nil || len(paP1.Items) != 1 || paP1.Total != 2 || !paP1.HasMore {
			t.Fatalf("PartyAddresses.SelectByPartyID page 1 failed: page=%+v err=%v", paP1, err)
		} else if paP2, err := reg.PartyAddresses.SelectByPartyID(ctx, supplier.ID, domainrepo.PageRequest{Limit: 1, Offset: 1}); err != nil || len(paP2.Items) != 1 || paP2.Total != 2 || paP2.HasMore || paP2.Items[0].ID == paP1.Items[0].ID {
			t.Fatalf("PartyAddresses.SelectByPartyID page 2 failed: page=%+v p1=%+v err=%v", paP2, paP1, err)
		}
		if usagePAs, err := reg.PartyAddresses.SelectByPartyAndUsage(ctx, supplier.ID, entity.AddressUsageKindBilling, domainrepo.PageRequest{}); err != nil || len(usagePAs.Items) != 2 || usagePAs.Total != 2 || usagePAs.HasMore {
			t.Fatalf("PartyAddresses.SelectByPartyAndUsage failed: page=%+v err=%v", usagePAs, err)
		}
		if addrList, err := reg.PartyAddresses.SelectList(ctx, domainrepo.PartyAddressFilter{
			PartyID: supplier.ID,
			Page:    domainrepo.PageRequest{Limit: 1, Offset: 1},
		}); err != nil || len(addrList.Items) != 1 || addrList.Total != 2 || addrList.HasMore {
			t.Fatalf("PartyAddresses.SelectList failed: page=%+v err=%v", addrList, err)
		}
		if triplePA, err := reg.PartyAddresses.SelectByPartyAddressUsage(ctx, supplier.ID, addrBill.ID, entity.AddressUsageKindBilling); err != nil || triplePA.ID != pa1.ID {
			t.Fatalf("PartyAddresses.SelectByPartyAddressUsage failed: got=%v err=%v", triplePA, err)
		}
		if err := reg.PartyAddresses.DeleteByID(ctx, pa1.ID); err != nil {
			t.Fatalf("PartyAddresses.DeleteByID failed: %v", err)
		}
		if _, err := reg.PartyAddresses.SelectByID(ctx, pa1.ID); !errors.Is(err, eferrors.ErrPartyAddressNotFound) {
			t.Fatalf("expected ErrPartyAddressNotFound after delete, got %v", err)
		}

		// Soft-delete Party dan pastikan status berubah ke inactive serta tersembunyi dari SelectByID.
		if err := reg.Parties.SoftDeleteByID(ctx, supplier.ID, now.Add(3*time.Minute)); err != nil {
			t.Fatalf("Parties.SoftDeleteByID failed: %v", err)
		}
		if _, err := reg.Parties.SelectByID(ctx, supplier.ID); !errors.Is(err, eferrors.ErrPartyNotFound) {
			t.Fatalf("expected ErrPartyNotFound after soft delete, got %v", err)
		}
	})

	t.Run("Transactor rollback/commit + RegisterAfterCommit + OutboxRepository", func(t *testing.T) {
		truncateFoundationTables(t, db)
		now := time.Now().Truncate(time.Microsecond)

		// 1. Rollback membatalkan insert dan tidak menjalankan after-commit hook.
		rolledBackAddr, _ := entity.NewAddress(mustUUID(t), entity.AddressParams{
			AddressLine1: "Jl. Rollback",
			City:         "Surabaya",
			CountryCode:  "ID",
		}, now)
		afterHookRan := false
		sentinelErr := errors.New("force rollback")
		txErr := reg.Transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
			if err := reg.Addresses.Insert(txCtx, rolledBackAddr); err != nil {
				return err
			}
			pgstore.RegisterAfterCommit(txCtx, func(context.Context) {
				afterHookRan = true
			})
			return sentinelErr
		})
		if !errors.Is(txErr, eferrors.ErrUnavailable) {
			t.Fatalf("expected wrapped unavailable error on generic txErr, got %v", txErr)
		}
		if afterHookRan {
			t.Fatal("after-commit hook must not execute when transaction rolls back")
		}
		if _, err := reg.Addresses.SelectByID(ctx, rolledBackAddr.ID); !errors.Is(err, eferrors.ErrAddressNotFound) {
			t.Fatalf("expected rolled back address not persisted, got %v", err)
		}

		// 2. Commit menyimpan entitas + outbox dan menjalankan after-commit hook di luar konteks transaksi.
		partyID := mustUUID(t)
		committedParty, _ := entity.NewParty(partyID, entity.PartyParams{
			PartnerCode:       "VND-TX-01",
			Kind:              entity.PartyKindOrganization,
			DisplayName:       "PT Vendor Transaksional",
			PreferredLanguage: "id",
			Status:            entity.PartyStatusPendingVerification,
		}, now)
		outboxEvt, err := entity.NewOutboxEvent(
			mustUUID(t),
			entity.EventTypePartyCreated,
			entity.AggregateTypeParty,
			partyID,
			map[string]any{"partner_code": "vnd-tx-01"},
			now.Add(-time.Minute),
		)
		if err != nil {
			t.Fatalf("NewOutboxEvent failed: %v", err)
		}

		afterHookRan = false
		if err := reg.Transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
			if err := reg.Parties.Insert(txCtx, committedParty); err != nil {
				return err
			}
			if err := reg.Outbox.Insert(txCtx, outboxEvt); err != nil {
				return err
			}
			pgstore.RegisterAfterCommit(txCtx, func(postCtx context.Context) {
				afterHookRan = true
				// Pastikan postCtx dapat melakukan pembacaan pasca-commit.
				if p, err := reg.Parties.SelectByID(postCtx, partyID); err != nil || p.ID != partyID {
					t.Errorf("post-commit lookup failed: got=%v err=%v", p, err)
				}
			})
			return nil
		}); err != nil {
			t.Fatalf("WithinTransaction commit failed: %v", err)
		}
		if !afterHookRan {
			t.Fatal("expected after-commit hook to run after commit")
		}

		if pending, err := reg.Outbox.CountPending(ctx); err != nil || pending != 1 {
			t.Fatalf("CountPending want=1 got=%d err=%v", pending, err)
		}

		// Relay worker pickup -> retry -> failed -> new effective cleanup.
		due, err := reg.Outbox.SelectDuePending(ctx, now, 10)
		if err != nil || len(due) != 1 || due[0].ID != outboxEvt.ID {
			t.Fatalf("SelectDuePending failed: got=%v err=%v", due, err)
		}

		if err := reg.Outbox.UpdateRetry(ctx, outboxEvt.ID, 1, now.Add(5*time.Minute), "temporary relay timeout"); err != nil {
			t.Fatalf("UpdateRetry failed: %v", err)
		}
		dueLater, err := reg.Outbox.SelectDuePending(ctx, now, 10)
		if err != nil || len(dueLater) != 0 {
			t.Fatalf("expected zero due events before retry timestamp, got len=%d err=%v", len(dueLater), err)
		}

		if err := reg.Outbox.UpdateFailed(ctx, outboxEvt.ID, "permanent dead letter"); err != nil {
			t.Fatalf("UpdateFailed failed: %v", err)
		}
		if failed, err := reg.Outbox.CountFailed(ctx); err != nil || failed != 1 {
			t.Fatalf("CountFailed want=1 got=%d err=%v", failed, err)
		}

		if err := reg.Outbox.UpdateEffective(ctx, outboxEvt.ID, now.Add(-2*time.Hour)); err != nil {
			t.Fatalf("UpdateEffective failed: %v", err)
		}
		history, err := reg.Outbox.SelectByAggregate(ctx, entity.AggregateTypeParty, partyID, domainrepo.PageRequest{})
		if err != nil || history.Total != 1 || history.Items[0].Status != entity.OutboxStatusEffective {
			t.Fatalf("SelectByAggregate failed: history=%+v err=%v", history, err)
		}

		deletedCount, err := reg.Outbox.DeleteEffectiveBefore(ctx, now.Add(-time.Hour))
		if err != nil || deletedCount != 1 {
			t.Fatalf("DeleteEffectiveBefore want=1 got=%d err=%v", deletedCount, err)
		}
		if _, err := reg.Outbox.SelectByID(ctx, outboxEvt.ID); !errors.Is(err, eferrors.ErrOutboxEventNotFound) {
			t.Fatalf("expected ErrOutboxEventNotFound after retention purge, got %v", err)
		}
	})
}
