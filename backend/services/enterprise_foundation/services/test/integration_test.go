package services_test

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	efdb "github.com/integrione/erp-retail/backend/services/enterprise_foundation/db"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/entity"
	eferrors "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/repository"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/repository"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/services"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/services/address"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/services/legalentity"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/services/location"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/services/organizationunit"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/services/party"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/transport/grpc/handler"
	"gorm.io/gorm"
)

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
	t.Skip("skipping live PostgreSQL services integration test: set EF_TEST_DATABASE_DSN or EF_DATABASE_DSN")
	return ""
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
		addresses
		RESTART IDENTITY CASCADE`).Error; err != nil {
		t.Fatalf("truncate foundation tables failed: %v", err)
	}
}

func TestLivePostgreSQLServicesAndWiringEndToEnd(t *testing.T) {
	dsn := resolveLiveTestDSN(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	db, err := efdb.Init(ctx, dsn, true)
	if err != nil {
		t.Fatalf("efdb.Init failed: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("db.DB failed: %v", err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	truncateFoundationTables(t, db)
	t.Cleanup(func() {
		truncateFoundationTables(t, db)
	})

	repos, err := repository.New(db)
	if err != nil {
		t.Fatalf("repository.New failed: %v", err)
	}
	reg, err := services.NewRegistry(repos, services.Options{})
	if err != nil {
		t.Fatalf("services.NewRegistry failed: %v", err)
	}
	handlerDeps := handler.Dependencies{Services: reg}
	if handlerDeps.Services == nil {
		t.Fatalf("expected non-nil handler.Dependencies.Services")
	}

	// 1. Address FindOrCreate & dedup + GetAll.
	hqAddr, created, err := reg.Addresses.FindOrCreate(ctx, address.CreateInput{
		AddressLine1: "Jl. Kebon Sirih No. 12",
		City:         "Jakarta Pusat",
		CountryCode:  "ID",
	})
	if err != nil || !created {
		t.Fatalf("Addresses.FindOrCreate first failed: created=%v err=%v", created, err)
	}
	reusedAddr, createdAgain, err := reg.Addresses.FindOrCreate(ctx, address.CreateInput{
		AddressLine1: "jl. kebon sirih no. 12",
		City:         "jakarta pusat",
		CountryCode:  "ID",
	})
	if err != nil || createdAgain || reusedAddr.ID != hqAddr.ID {
		t.Fatalf("Addresses.FindOrCreate dedup failed: reused=%+v created=%v err=%v", reusedAddr, createdAgain, err)
	}

	// 2. LegalEntity Create, duplicate prevention, GetAll, dan Scope.
	le, err := reg.LegalEntities.Create(ctx, legalentity.CreateInput{
		EntityCode:               "PT-INTEGRI",
		LegalName:                "PT IntegriOne Ritel",
		IsPrimaryOperatingEntity: true,
		Status:                   entity.LegalEntityStatusActive,
		JurisdictionCountryCode:  "ID",
		TaxIDNumber:              stringPtr("01.100.200.3-400.000"),
		IsPKP:                    true,
		RegisteredAddressID:      &hqAddr.ID,
		FiscalYearStartMonth:     1,
	})
	if err != nil {
		t.Fatalf("LegalEntities.Create failed: %v", err)
	}
	if _, err := reg.LegalEntities.Create(ctx, legalentity.CreateInput{
		EntityCode:              "pt-integri",
		LegalName:               "Duplikat",
		Status:                  entity.LegalEntityStatusActive,
		JurisdictionCountryCode: "ID",
		FiscalYearStartMonth:    1,
	}); !errors.Is(err, eferrors.ErrLegalEntityCodeExists) {
		t.Fatalf("expected ErrLegalEntityCodeExists on duplicate create, got %v", err)
	}
	if scopeLE, err := reg.LegalEntities.ValidateScopeTarget(ctx, le.ID); err != nil || !scopeLE.Eligible {
		t.Fatalf("ValidateScopeTarget LE failed: %+v err=%v", scopeLE, err)
	}

	// 3. Location Create (HQ & Warehouse) & Scope.
	hqLoc, err := reg.Locations.Create(ctx, location.CreateInput{
		LegalEntityID: le.ID,
		AddressID:     &hqAddr.ID,
		LocationCode:  "HQ-01",
		Name:          "Kantor Pusat IntegriOne",
		Kind:          entity.LocationKindHeadquarters,
		Timezone:      "Asia/Jakarta",
	})
	if err != nil {
		t.Fatalf("Locations.Create HQ failed: %v", err)
	}
	whLoc, err := reg.Locations.Create(ctx, location.CreateInput{
		LegalEntityID:    le.ID,
		ParentLocationID: &hqLoc.ID,
		AddressID:        &hqAddr.ID,
		LocationCode:     "WH-01",
		Name:             "Gudang Utama",
		Kind:             entity.LocationKindWarehouse,
		Timezone:         "Asia/Jakarta",
	})
	if err != nil {
		t.Fatalf("Locations.Create WH failed: %v", err)
	}
	if whScope, err := reg.Locations.ValidateWarehouseScope(ctx, whLoc.ID); err != nil || !whScope.Eligible {
		t.Fatalf("ValidateWarehouseScope WH failed: %+v err=%v", whScope, err)
	}

	// 4. OrganizationUnit Create & Scope.
	orgUnit, err := reg.OrganizationUnits.Create(ctx, organizationunit.CreateInput{
		LegalEntityID:     le.ID,
		PrimaryLocationID: &hqLoc.ID,
		UnitCode:          "OPS-WH",
		Name:              "Operasional Gudang",
		Kind:              entity.OrgUnitKindDepartment,
	})
	if err != nil {
		t.Fatalf("OrganizationUnits.Create failed: %v", err)
	}
	if ouScope, err := reg.OrganizationUnits.ValidateScopeTarget(ctx, orgUnit.ID); err != nil || !ouScope.Eligible {
		t.Fatalf("OrganizationUnits.ValidateScopeTarget failed: %+v err=%v", ouScope, err)
	}

	// 5. Party, Role, Contact + IAM Portal, dan Address Link + GetAll + Outbox.
	vendor, err := reg.Parties.Create(ctx, party.CreateInput{
		PartnerCode:       "VND-001",
		Kind:              entity.PartyKindOrganization,
		DisplayName:       "PT Mitra Kain Jaya",
		PreferredLanguage: "id",
		Status:            entity.PartyStatusPendingVerification,
	})
	if err != nil {
		t.Fatalf("Parties.Create failed: %v", err)
	}
	if _, err := reg.Parties.VerifyAndActivate(ctx, vendor.ID); err != nil {
		t.Fatalf("Parties.VerifyAndActivate failed: %v", err)
	}
	if _, err := reg.Parties.AssignRole(ctx, party.AssignRoleInput{
		PartyID: vendor.ID,
		Role:    entity.PartyRoleKindVendor,
	}); err != nil {
		t.Fatalf("Parties.AssignRole failed: %v", err)
	}
	iamUserID := "99999999-1111-4111-8111-111111111111"
	pic, err := reg.Parties.AddContact(ctx, party.AddContactInput{
		PartyID:            vendor.ID,
		FullName:           "Rina Kusuma",
		ContactRole:        "vendor_pic",
		Email:              stringPtr("rina@kainjaya.id"),
		PortalAccessStatus: entity.PortalAccessStatusNone,
		IsPrimary:          true,
	})
	if err != nil {
		t.Fatalf("Parties.AddContact failed: %v", err)
	}
	if _, err := reg.Parties.RequestPortalLink(ctx, pic.ID, iamUserID); err != nil {
		t.Fatalf("Parties.RequestPortalLink failed: %v", err)
	}
	if _, err := reg.Parties.ActivatePortalAccess(ctx, pic.ID, iamUserID); err != nil {
		t.Fatalf("Parties.ActivatePortalAccess failed: %v", err)
	}
	eval, err := reg.Parties.EvaluateRepresentativePortalAccess(ctx, vendor.ID, iamUserID, entity.PartyRoleKindVendor)
	if err != nil || !eval.Allowed || eval.PartyID != vendor.ID {
		t.Fatalf("EvaluateRepresentativePortalAccess failed: %+v err=%v", eval, err)
	}
	if _, err := reg.Parties.LinkAddress(ctx, party.LinkAddressInput{
		PartyID:   vendor.ID,
		AddressID: hqAddr.ID,
		Usage:     entity.AddressUsageKindRegisteredOffice,
		IsPrimary: true,
	}); err != nil {
		t.Fatalf("Parties.LinkAddress failed: %v", err)
	}

	// GetAll & paginasi across all 5 services.
	allAddrs, err := reg.Addresses.GetAll(ctx, domainrepo.PageRequest{})
	if err != nil || len(allAddrs.Items) != 1 || allAddrs.Total != 1 || allAddrs.HasMore {
		t.Fatalf("Addresses.GetAll failed: page=%+v err=%v", allAddrs, err)
	}
	allLEs, err := reg.LegalEntities.GetAll(ctx, domainrepo.PageRequest{})
	if err != nil || len(allLEs.Items) != 1 || allLEs.Total != 1 || allLEs.HasMore {
		t.Fatalf("LegalEntities.GetAll failed: page=%+v err=%v", allLEs, err)
	}
	allLocs, err := reg.Locations.GetAll(ctx, domainrepo.PageRequest{})
	if err != nil || len(allLocs.Items) != 2 || allLocs.Total != 2 || allLocs.HasMore {
		t.Fatalf("Locations.GetAll failed: page=%+v err=%v", allLocs, err)
	}
	if locsP1, err := reg.Locations.GetAll(ctx, domainrepo.PageRequest{Limit: 1, Offset: 0}); err != nil || len(locsP1.Items) != 1 || locsP1.Total != 2 || !locsP1.HasMore {
		t.Fatalf("Locations.GetAll page 1 failed: page=%+v err=%v", locsP1, err)
	} else if locsP2, err := reg.Locations.GetAll(ctx, domainrepo.PageRequest{Limit: 1, Offset: 1}); err != nil || len(locsP2.Items) != 1 || locsP2.Total != 2 || locsP2.HasMore || locsP2.Items[0].ID == locsP1.Items[0].ID {
		t.Fatalf("Locations.GetAll page 2 failed: page=%+v p1=%+v err=%v", locsP2, locsP1, err)
	}
	allOUs, err := reg.OrganizationUnits.GetAll(ctx, domainrepo.PageRequest{})
	if err != nil || len(allOUs.Items) != 1 || allOUs.Total != 1 || allOUs.HasMore {
		t.Fatalf("OrganizationUnits.GetAll failed: page=%+v err=%v", allOUs, err)
	}
	allParties, err := reg.Parties.GetAll(ctx, domainrepo.PageRequest{})
	if err != nil || len(allParties.Items) != 1 || allParties.Total != 1 || allParties.HasMore {
		t.Fatalf("Parties.GetAll failed: page=%+v err=%v", allParties, err)
	}
	allRoles, err := reg.Parties.GetAllRoles(ctx, domainrepo.PageRequest{})
	if err != nil || len(allRoles.Items) != 1 || allRoles.Total != 1 || allRoles.HasMore {
		t.Fatalf("Parties.GetAllRoles failed: page=%+v err=%v", allRoles, err)
	}
	allContacts, err := reg.Parties.GetAllContacts(ctx, domainrepo.PageRequest{})
	if err != nil || len(allContacts.Items) != 1 || allContacts.Total != 1 || allContacts.HasMore {
		t.Fatalf("Parties.GetAllContacts failed: page=%+v err=%v", allContacts, err)
	}
	allAddrLinks, err := reg.Parties.GetAllAddressLinks(ctx, domainrepo.PageRequest{})
	if err != nil || len(allAddrLinks.Items) != 1 || allAddrLinks.Total != 1 || allAddrLinks.HasMore {
		t.Fatalf("Parties.GetAllAddressLinks failed: page=%+v err=%v", allAddrLinks, err)
	}

	// Verifikasi event outbox tertulis di PostgreSQL.
	outboxPage, err := repos.Outbox.SelectByAggregate(ctx, entity.AggregateTypeParty, vendor.ID, domainrepo.PageRequest{Limit: 10})
	if err != nil || len(outboxPage.Items) < 2 {
		t.Fatalf("expected party outbox events in PostgreSQL, got items=%d err=%v", len(outboxPage.Items), err)
	}
}

func stringPtr(v string) *string { return &v }
