//go:build e2e

package e2e

import (
	"fmt"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/shared/rbac"
)

func loadEFPermissionCatalog(t *testing.T) []rbac.PermissionSpec {
	t.Helper()
	catalogs, err := rbac.LoadEmbeddedCatalogs()
	if err != nil {
		t.Fatalf("rbac.LoadEmbeddedCatalogs() error = %v", err)
	}
	specs := catalogs.DomainPermissions("enterprise_foundation")
	if len(specs) != 10 {
		t.Fatalf("expected 10 enterprise_foundation permissions in manifest, got %d", len(specs))
	}
	return specs
}

const (
	gqlEFAddressFields          = `id _id label address_line_1 address_line_2 city province postal_code country_code formatted_address latitude longitude created_at updated_at`
	gqlEFLegalEntityFields      = `id _id parent_legal_entity_id entity_code legal_name trading_name legal_form is_primary_operating_entity status jurisdiction_country_code nib tax_id_number is_pkp registered_address_id fiscal_year_start_month created_at updated_at`
	gqlEFLocationFields         = `id _id legal_entity_id parent_location_id address_id location_code name kind phone timezone operational_attributes_json is_active created_at updated_at`
	gqlEFOrgUnitFields          = `id _id legal_entity_id parent_unit_id primary_location_id unit_code name kind cost_center_code head_person_id is_active created_at updated_at`
	gqlEFPartyFields            = `id _id partner_code kind display_name legal_name email phone website_url tax_id_number tax_registered_name is_pkp is_tax_exempt tax_exemption_reason preferred_language status notes created_at updated_at`
	gqlEFPartyRoleFields        = `id _id party_id role is_active reference_note created_at updated_at`
	gqlEFPartyContactFields     = `id _id party_id iam_user_id full_name job_title contact_role email phone portal_access_status is_primary is_active created_at updated_at`
	gqlEFPartyAddressFields     = `id _id party_id address_id usage is_primary created_at updated_at`
	gqlEFScopeDecisionFields    = `eligible scope_kind target_id code legal_entity_id`
	gqlEFPortalEvaluationFields = `allowed party_id contact_id iam_user_id active_roles`
)

func psqlEFExec(t *testing.T, cfg e2eRuntimeConfig, sqlStmt string) {
	t.Helper()
	cmd := exec.Command(
		"docker", "exec", cfg.postgresService,
		"psql", "-v", "ON_ERROR_STOP=1", "-U", "postgres", "-d", "enterprise_foundation", "-tAc", sqlStmt,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("ef psql exec failed: %v\nSQL: %s\nOutput: %s", err, sqlStmt, strings.TrimSpace(string(out)))
	}
}

func cleanupEFEntityRows(t *testing.T, cfg e2eRuntimeConfig, table, rawID string) {
	t.Helper()
	if strings.TrimSpace(rawID) == "" {
		return
	}
	cmd := exec.Command(
		"docker", "exec", cfg.postgresService,
		"psql", "-U", "postgres", "-d", "enterprise_foundation", "-tAc",
		fmt.Sprintf("DELETE FROM %s WHERE id = %s::uuid;", table, sqlQuote(rawID)),
	)
	_, _ = cmd.CombinedOutput()
}

// seedEFPermissionCatalog memastikan 10 permission Enterprise Foundation tersedia pada katalog IAM lokal dan membersihkan baris uji yang disisipkan sendiri.
func seedEFPermissionCatalog(t *testing.T, cfg e2eRuntimeConfig) {
	t.Helper()
	existingRaw := psqlScalar(t, cfg, "SELECT COALESCE(string_agg(code, ','), '') FROM permissions WHERE code LIKE 'enterprise_foundation.%';")
	existing := make(map[string]bool)
	for _, code := range strings.Split(existingRaw, ",") {
		if trimmed := strings.TrimSpace(code); trimmed != "" {
			existing[trimmed] = true
		}
	}

	efPerms := loadEFPermissionCatalog(t)
	var valueRows []string
	var insertedCodes []string
	for _, p := range efPerms {
		if existing[p.Code] {
			continue
		}
		insertedCodes = append(insertedCodes, p.Code)
		valueRows = append(valueRows, fmt.Sprintf("(%s, %s, %s, %s, %t)",
			sqlQuote(p.Code), sqlQuote(p.Resource), sqlQuote(p.Action), sqlQuote(p.Description), p.IsSensitive))
	}
	if len(valueRows) == 0 {
		return
	}

	sqlStmt := "INSERT INTO permissions (code, resource, action, description, is_sensitive) VALUES\n" +
		strings.Join(valueRows, ",\n") +
		"\nON CONFLICT (code) DO NOTHING;"
	psqlExec(t, cfg, sqlStmt)

	t.Cleanup(func() {
		quoted := make([]string, 0, len(insertedCodes))
		for _, c := range insertedCodes {
			quoted = append(quoted, sqlQuote(c))
		}
		_ = exec.Command(
			"docker", "exec", cfg.postgresService,
			"psql", "-U", "postgres", "-d", "iam", "-tAc",
			fmt.Sprintf("DELETE FROM permissions WHERE code IN (%s);", strings.Join(quoted, ",")),
		).Run()
	})
}

// provisionEFUserWithPerms membuat pengguna staff IAM baru dengan peran yang berisi daftar izin serta cakupan tertentu.
func provisionEFUserWithPerms(
	t *testing.T,
	admin *browserClient,
	cfg e2eRuntimeConfig,
	prefix string,
	permCodes []string,
	scopeType string,
	scopeID *string,
) (fixtureUserAccount, *browserClient) {
	t.Helper()
	roleID, _ := createFixtureRole(t, admin, cfg, prefix)
	for _, code := range permCodes {
		permID := findPermissionIDByCode(t, admin, code)
		_ = requireMutationChildOK(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($r: ID!, $p: ID!) {
			assignRolePermission(role_id: $r, permission_id: $p) { %s }
		}`, gqlRolePermPayloadFields), map[string]any{
			"r": roleID,
			"p": permID,
		}), "assignRolePermission", "role_permission_edge")
	}

	staff := createFixtureUser(t, admin, cfg, "EMPLOYEE", true)
	assignInput := map[string]any{
		"user_id":    staff.ID,
		"role_id":    roleID,
		"scope_type": scopeType,
	}
	if scopeID != nil && strings.TrimSpace(*scopeID) != "" {
		assignInput["scope_id"] = strings.TrimSpace(*scopeID)
	}
	_ = requireMutationChildOK(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($in: AssignUserRoleInput!) {
		assignUserRole(input: $in) { %s }
	}`, gqlUserRolePayloadFields), map[string]any{"in": assignInput}), "assignUserRole", "user_role_edge")

	resetGatewayRateLimits(t, cfg)
	client, payload := loginFixtureUser(t, cfg, staff.Username, staff.Password)
	assertUserLoginPayload(t, payload, "AUTHENTICATED")
	return staff, client
}

func assertContainerHealthy(t *testing.T, containerName string) {
	t.Helper()
	out, err := exec.Command("docker", "inspect", "--format", "{{.State.Status}}:{{if .State.Health}}{{.State.Health.Status}}{{else}}none{{end}}", containerName).CombinedOutput()
	if err != nil {
		t.Fatalf("inspect container %s failed: %v (%s)", containerName, err, string(out))
	}
	got := strings.TrimSpace(string(out))
	if got != "running:healthy" {
		t.Fatalf("container %s state = %q, want running:healthy", containerName, got)
	}
}

func assertNoSensitiveUpstreamDetails(t *testing.T, msg string) {
	t.Helper()
	lower := strings.ToLower(msg)
	for _, forbidden := range []string{"enterprise-foundation:9002", "127.0.0.1", "dial tcp", "postgres", "pq:"} {
		if strings.Contains(lower, forbidden) {
			t.Fatalf("client error message leaked internal detail (%q): %q", forbidden, msg)
		}
	}
}

func TestEnterpriseFoundation_FullLifecycleAndNestedGraphQL(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	seedEFPermissionCatalog(t, cfg)

	admin := newSuperAdminBrowser(t, cfg)
	// Verifikasi langsung human superadmin memiliki akses parties (enterprise_foundation.parties.read) sejak bootstrap.
	adminParties := doGQLOK(t, admin, `query { parties(first: 5) { total_count } }`, nil)
	_ = requireOpMap(t, adminParties, "parties")

	efPerms := loadEFPermissionCatalog(t)
	allPerms := make([]string, 0, len(efPerms))
	for _, p := range efPerms {
		allPerms = append(allPerms, p.Code)
	}
	efStaff, efClient := provisionEFUserWithPerms(t, admin, cfg, "ef_full", allPerms, "GLOBAL", nil)
	_, custClient := createFixtureCustomerAndLogin(t, admin, cfg)
	anonClient := newAnonBrowser(t, cfg)

	// Catat ID entitas untuk cleanup berurutan terlepas dari status sukses/gagal.
	var (
		createdPartyAddrID string
		createdContactID   string
		createdRoleID      string
		createdPartyID     string
		createdUnitChildID string
		createdUnitRootID  string
		createdLocWHID     string
		createdLocHQID     string
		createdLEChildID   string
		createdLERootID    string
		createdAddr1ID     string
		createdAddr2ID     string
		createdAddr3ID     string
	)
	t.Cleanup(func() {
		cleanupEFEntityRows(t, cfg, "party_addresses", rawUUID(createdPartyAddrID))
		cleanupEFEntityRows(t, cfg, "party_contacts", rawUUID(createdContactID))
		cleanupEFEntityRows(t, cfg, "party_roles", rawUUID(createdRoleID))
		cleanupEFEntityRows(t, cfg, "parties", rawUUID(createdPartyID))
		cleanupEFEntityRows(t, cfg, "organization_units", rawUUID(createdUnitChildID))
		cleanupEFEntityRows(t, cfg, "organization_units", rawUUID(createdUnitRootID))
		cleanupEFEntityRows(t, cfg, "locations", rawUUID(createdLocWHID))
		cleanupEFEntityRows(t, cfg, "locations", rawUUID(createdLocHQID))
		cleanupEFEntityRows(t, cfg, "legal_entities", rawUUID(createdLEChildID))
		cleanupEFEntityRows(t, cfg, "legal_entities", rawUUID(createdLERootID))
		cleanupEFEntityRows(t, cfg, "addresses", rawUUID(createdAddr3ID))
		cleanupEFEntityRows(t, cfg, "addresses", rawUUID(createdAddr2ID))
		cleanupEFEntityRows(t, cfg, "addresses", rawUUID(createdAddr1ID))
	})

	// 1. Verifikasi fail-closed: anonim (UNAUTHENTICATED) dan customer (FORBIDDEN) ditolak Gateway sebelum RPC.
	doGQLError(t, anonClient, `query { allAddresses(first: 5) { total_count } }`, nil, "UNAUTHENTICATED")
	_, _ = doMutationFail(t, anonClient, fmt.Sprintf(`mutation ($in: CreateAddressInput!) { createAddress(input: $in) { %s } }`, gqlMutationResultFields), map[string]any{
		"in": map[string]any{"address_line_1": "Anon Street 1", "city": "Jakarta", "country_code": "ID"},
	}, "createAddress", "AUTHENTICATION")

	doGQLError(t, custClient, `query { addresses(first: 5) { total_count } }`, nil, "FORBIDDEN")
	_, _ = doMutationFail(t, custClient, fmt.Sprintf(`mutation ($in: CreateAddressInput!) { createAddress(input: $in) { %s } }`, gqlMutationResultFields), map[string]any{
		"in": map[string]any{"address_line_1": "Cust Street 1", "city": "Jakarta", "country_code": "ID"},
	}, "createAddress", "AUTHORIZATION")

	// 2. Uji AddressService: create, findOrCreate (create -> find), update, pagination Relay, dan isAddressReferenced.
	tag := uniqueTag("ef")
	addr1Data := doGQLOK(t, efClient, fmt.Sprintf(`mutation ($in: CreateAddressInput!) {
		createAddress(input: $in) { address { %s } %s }
	}`, gqlEFAddressFields, gqlMutationResultFields), map[string]any{
		"in": map[string]any{
			"label":          "Kantor Pusat " + tag,
			"address_line_1": "Jl. Jend. Sudirman Kav. 50 " + tag,
			"city":           "Jakarta Selatan",
			"province":       "DKI Jakarta",
			"postal_code":    "12190",
			"country_code":   "ID",
		},
	})
	addr1 := requireMutationChildOK(t, addr1Data, "createAddress", "address")
	createdAddr1ID = requireNonEmptyString(t, addr1, "id")
	addr1RawID := requireNonEmptyString(t, addr1, "_id")

	addr2Data := doGQLOK(t, efClient, fmt.Sprintf(`mutation ($in: CreateAddressInput!) {
		createAddress(input: $in) { address { %s } %s }
	}`, gqlEFAddressFields, gqlMutationResultFields), map[string]any{
		"in": map[string]any{
			"label":          "Gudang Cikarang " + tag,
			"address_line_1": "Kawasan Industri Jababeka Blok A1 " + tag,
			"city":           "Bekasi",
			"country_code":   "ID",
		},
	})
	addr2 := requireMutationChildOK(t, addr2Data, "createAddress", "address")
	createdAddr2ID = requireNonEmptyString(t, addr2, "id")

	// findOrCreateAddress pertama -> created=true, kedua identik -> created=false.
	focQuery := fmt.Sprintf(`mutation ($in: FindOrCreateAddressInput!) {
		findOrCreateAddress(input: $in) { address { %s } created %s }
	}`, gqlEFAddressFields, gqlMutationResultFields)
	focIn := map[string]any{
		"address_line_1": "Jl. Merdeka Bandung 99 " + tag,
		"city":           "Bandung",
		"country_code":   "ID",
	}
	foc1 := requireMutationSuccess(t, doGQLOK(t, efClient, focQuery, map[string]any{"in": focIn}), "findOrCreateAddress")
	createdAddr3ID = requireNonEmptyString(t, requireChildMap(t, foc1, "address"), "id")
	if !requireBoolField(t, foc1, "created") {
		t.Fatalf("first findOrCreateAddress expected created=true, got false")
	}
	foc2 := requireMutationSuccess(t, doGQLOK(t, efClient, focQuery, map[string]any{"in": focIn}), "findOrCreateAddress")
	if requireBoolField(t, foc2, "created") {
		t.Fatalf("second identical findOrCreateAddress expected created=false, got true")
	}

	// Update + query single/list/allAddresses/addressesByIds + delete unreferenced address3.
	updAddr := requireMutationChildOK(t, doGQLOK(t, efClient, fmt.Sprintf(`mutation ($in: UpdateAddressInput!) {
		updateAddress(input: $in) { address { %s } %s }
	}`, gqlEFAddressFields, gqlMutationResultFields), map[string]any{
		"in": map[string]any{
			"id":             createdAddr1ID,
			"label":          "HQ Updated " + tag,
			"address_line_1": "Jl. Jend. Sudirman Kav. 52 " + tag,
			"city":           "Jakarta Selatan",
			"country_code":   "ID",
		},
	}), "updateAddress", "address")
	if updAddr["city"] != "Jakarta Selatan" {
		t.Fatalf("unexpected updated city: %+v", updAddr)
	}

	addrPage1 := requireOpMap(t, doGQLOK(t, efClient, fmt.Sprintf(`query {
		allAddresses(first: 1) { edges { cursor node { %s } } pageInfo { %s } total_count }
	}`, gqlEFAddressFields, gqlPageInfoFields), nil), "allAddresses")
	pageInfo1 := assertPageInfoNonNull(t, addrPage1)
	if requireNumberField(t, addrPage1, "total_count") < 3 || !requireBoolField(t, pageInfo1, "hasNextPage") {
		t.Fatalf("expected >= 3 addresses and hasNextPage=true, got %+v", addrPage1)
	}
	afterCur := requireNonEmptyString(t, pageInfo1, "endCursor")
	addrPage2 := requireOpMap(t, doGQLOK(t, efClient, fmt.Sprintf(`query ($after: String) {
		allAddresses(first: 1, after: $after) { nodes { id } pageInfo { %s } total_count }
	}`, gqlPageInfoFields), map[string]any{"after": afterCur}), "allAddresses")
	if len(requireConnectionNodes(t, addrPage2)) != 1 {
		t.Fatalf("expected 1 node on second cursor page")
	}

	refBefore := doGQLOK(t, efClient, `query ($id: ID!) { isAddressReferenced(id: $id) }`, map[string]any{"id": addr1RawID})
	if refBefore["isAddressReferenced"] != false {
		t.Fatalf("expected unlinked addr1RawID isAddressReferenced=false, got %v", refBefore["isAddressReferenced"])
	}
	refCheck := doGQLOK(t, efClient, `query ($id: ID!, $ids: [ID!]!) {
		isAddressReferenced(id: $id)
		addressesByIds(ids: $ids) { id _id city }
	}`, map[string]any{"id": createdAddr3ID, "ids": []string{createdAddr1ID, createdAddr2ID}})
	if refCheck["isAddressReferenced"] != false {
		t.Fatalf("expected unlinked address3 isAddressReferenced=false, got %v", refCheck["isAddressReferenced"])
	}
	requireMutationBoolOK(t, doGQLOK(t, efClient, fmt.Sprintf(`mutation ($id: ID!) {
		deleteAddress(id: $id) { deleted %s }
	}`, gqlMutationResultFields), map[string]any{"id": createdAddr3ID}), "deleteAddress", "deleted")
	createdAddr3ID = ""

	// 3. Uji LegalEntityService + nested registered_address + children + validateLegalEntityScopeTarget.
	leCodeRoot := "LE-" + strings.ToUpper(strings.ReplaceAll(tag, "_", "-"))
	leCodeChild := leCodeRoot + "-SUB"
	leRoot := requireMutationChildOK(t, doGQLOK(t, efClient, fmt.Sprintf(`mutation ($in: CreateLegalEntityInput!) {
		createLegalEntity(input: $in) { legal_entity { %s } %s }
	}`, gqlEFLegalEntityFields, gqlMutationResultFields), map[string]any{
		"in": map[string]any{
			"entity_code":               leCodeRoot,
			"legal_name":                "PT IntegriOne E2E Holding " + tag,
			"trading_name":              "IntegriOne",
			"jurisdiction_country_code": "ID",
			"nib":                       "NIB-" + tag,
			"tax_id_number":             "NPWP-" + tag,
			"is_pkp":                    true,
			"registered_address_id":     createdAddr1ID,
			"fiscal_year_start_month":   1,
		},
	}), "createLegalEntity", "legal_entity")
	createdLERootID = requireNonEmptyString(t, leRoot, "id")

	leChild := requireMutationChildOK(t, doGQLOK(t, efClient, fmt.Sprintf(`mutation ($in: CreateLegalEntityInput!) {
		createLegalEntity(input: $in) { legal_entity { %s } %s }
	}`, gqlEFLegalEntityFields, gqlMutationResultFields), map[string]any{
		"in": map[string]any{
			"parent_legal_entity_id":    createdLERootID,
			"entity_code":               leCodeChild,
			"legal_name":                "PT IntegriOne Retail Cabang " + tag,
			"jurisdiction_country_code": "ID",
		},
	}), "createLegalEntity", "legal_entity")
	createdLEChildID = requireNonEmptyString(t, leChild, "id")

	_ = requireMutationChildOK(t, doGQLOK(t, efClient, fmt.Sprintf(`mutation ($id: ID!) {
		activateLegalEntity(id: $id) { legal_entity { id status } %s }
	}`, gqlMutationResultFields), map[string]any{"id": createdLERootID}), "activateLegalEntity", "legal_entity")
	_ = requireMutationChildOK(t, doGQLOK(t, efClient, fmt.Sprintf(`mutation ($id: ID!) {
		suspendLegalEntity(id: $id) { legal_entity { id status } %s }
	}`, gqlMutationResultFields), map[string]any{"id": createdLERootID}), "suspendLegalEntity", "legal_entity")
	_ = requireMutationChildOK(t, doGQLOK(t, efClient, fmt.Sprintf(`mutation ($id: ID!) {
		activateLegalEntity(id: $id) { legal_entity { id status } %s }
	}`, gqlMutationResultFields), map[string]any{"id": createdLERootID}), "activateLegalEntity", "legal_entity")

	leNested := doGQLOK(t, efClient, fmt.Sprintf(`query ($id: ID!, $childID: ID!, $code: String!, $nib: String!) {
		legalEntity(id: $id) {
			%s
			registered_address { id city }
			children(first: 10) { nodes { id entity_code } total_count }
		}
		child: legalEntity(id: $childID) {
			id
			parent_legal_entity { id entity_code }
		}
		legalEntityByCode(entity_code: $code) { id entity_code }
		legalEntitiesByNib(nib: $nib, first: 5) { nodes { id } total_count }
		validateLegalEntityScopeTarget(id: $id) { %s }
		isAddressReferenced(id: %q)
	}`, gqlEFLegalEntityFields, gqlEFScopeDecisionFields, createdAddr1ID), map[string]any{
		"id":      createdLERootID,
		"childID": createdLEChildID,
		"code":    leCodeRoot,
		"nib":     "NIB-" + tag,
	})
	leObj := requireOpMap(t, leNested, "legalEntity")
	if requireChildMap(t, leObj, "registered_address")["id"] != createdAddr1ID {
		t.Fatalf("expected nested registered_address=%s, got %+v", createdAddr1ID, leObj["registered_address"])
	}
	if requireNumberField(t, requireChildMap(t, leObj, "children"), "total_count") != 1 {
		t.Fatalf("expected 1 child legal entity, got %+v", leObj["children"])
	}
	if leNested["isAddressReferenced"] != true {
		t.Fatalf("expected linked address1 to report isAddressReferenced=true")
	}

	// 4. Uji LocationService + nested legal_entity/address/parent_location + facility/warehouse scope targets.
	locHQCode := "HQ-" + strings.ToUpper(tag)
	locWHCode := "WH-" + strings.ToUpper(tag)
	locHQ := requireMutationChildOK(t, doGQLOK(t, efClient, fmt.Sprintf(`mutation ($in: CreateLocationInput!) {
		createLocation(input: $in) { location { %s } %s }
	}`, gqlEFLocationFields, gqlMutationResultFields), map[string]any{
		"in": map[string]any{
			"legal_entity_id": createdLERootID,
			"address_id":      createdAddr1ID,
			"location_code":   locHQCode,
			"name":            "Flagship HQ " + tag,
			"kind":            "HEADQUARTERS",
			"timezone":        "Asia/Jakarta",
		},
	}), "createLocation", "location")
	createdLocHQID = requireNonEmptyString(t, locHQ, "id")

	locWH := requireMutationChildOK(t, doGQLOK(t, efClient, fmt.Sprintf(`mutation ($in: CreateLocationInput!) {
		createLocation(input: $in) { location { %s } %s }
	}`, gqlEFLocationFields, gqlMutationResultFields), map[string]any{
		"in": map[string]any{
			"legal_entity_id":    createdLERootID,
			"parent_location_id": createdLocHQID,
			"address_id":         createdAddr2ID,
			"location_code":      locWHCode,
			"name":               "Warehouse Utama " + tag,
			"kind":               "WAREHOUSE",
			"timezone":           "Asia/Jakarta",
		},
	}), "createLocation", "location")
	createdLocWHID = requireNonEmptyString(t, locWH, "id")

	locQueries := doGQLOK(t, efClient, fmt.Sprintf(`query ($whID: ID!, $leID: ID!, $addrID: ID!, $code: String!) {
		location(id: $whID) {
			%s
			legal_entity { id entity_code }
			parent_location { id location_code }
			address { id city }
		}
		locationByCode(location_code: $code) { id }
		locationsByLegalEntityAndKind(legal_entity_id: $leID, kind: WAREHOUSE, first: 5) { nodes { id } total_count }
		locationsByAddressId(address_id: $addrID, first: 5) { nodes { id } total_count }
		validateFacilityScope(id: $whID) { %s }
		validateWarehouseScope(id: $whID) { %s }
	}`, gqlEFLocationFields, gqlEFScopeDecisionFields, gqlEFScopeDecisionFields), map[string]any{
		"whID":   createdLocWHID,
		"leID":   createdLERootID,
		"addrID": createdAddr2ID,
		"code":   locWHCode,
	})
	whObj := requireOpMap(t, locQueries, "location")
	if requireChildMap(t, whObj, "legal_entity")["id"] != createdLERootID ||
		requireChildMap(t, whObj, "parent_location")["id"] != createdLocHQID ||
		requireChildMap(t, whObj, "address")["id"] != createdAddr2ID {
		t.Fatalf("unexpected nested Location links: %+v", whObj)
	}
	if !requireBoolField(t, requireOpMap(t, locQueries, "validateWarehouseScope"), "eligible") {
		t.Fatalf("expected warehouse to pass validateWarehouseScope")
	}

	// 5. Uji OrganizationUnitService + AssignHeadPerson + nested primary_location.
	unitRootCode := "DIV-" + strings.ToUpper(tag)
	unitChildCode := "DEPT-" + strings.ToUpper(tag)
	unitRoot := requireMutationChildOK(t, doGQLOK(t, efClient, fmt.Sprintf(`mutation ($in: CreateOrganizationUnitInput!) {
		createOrganizationUnit(input: $in) { organization_unit { %s } %s }
	}`, gqlEFOrgUnitFields, gqlMutationResultFields), map[string]any{
		"in": map[string]any{
			"legal_entity_id":     createdLERootID,
			"primary_location_id": createdLocHQID,
			"unit_code":           unitRootCode,
			"name":                "Operations Division " + tag,
			"kind":                "DIVISION",
		},
	}), "createOrganizationUnit", "organization_unit")
	createdUnitRootID = requireNonEmptyString(t, unitRoot, "id")

	unitChild := requireMutationChildOK(t, doGQLOK(t, efClient, fmt.Sprintf(`mutation ($in: CreateOrganizationUnitInput!) {
		createOrganizationUnit(input: $in) { organization_unit { %s } %s }
	}`, gqlEFOrgUnitFields, gqlMutationResultFields), map[string]any{
		"in": map[string]any{
			"legal_entity_id": createdLERootID,
			"parent_unit_id":  createdUnitRootID,
			"unit_code":       unitChildCode,
			"name":            "Fulfillment Dept " + tag,
			"kind":            "DEPARTMENT",
		},
	}), "createOrganizationUnit", "organization_unit")
	createdUnitChildID = requireNonEmptyString(t, unitChild, "id")

	_ = requireMutationChildOK(t, doGQLOK(t, efClient, fmt.Sprintf(`mutation ($id: ID!, $head: ID) {
		assignOrganizationUnitHeadPerson(id: $id, head_person_id: $head) { organization_unit { id head_person_id } %s }
	}`, gqlMutationResultFields), map[string]any{
		"id":   createdUnitRootID,
		"head": efStaff.ID,
	}), "assignOrganizationUnitHeadPerson", "organization_unit")

	orgQueries := doGQLOK(t, efClient, fmt.Sprintf(`query ($id: ID!, $head: ID!) {
		organizationUnit(id: $id) {
			%s
			legal_entity { id }
			primary_location { id location_code }
			children(first: 5) { nodes { id unit_code } total_count }
		}
		organizationUnitsByHeadPersonId(head_person_id: $head, first: 5) { nodes { id } total_count }
		validateOrganizationUnitScopeTarget(id: $id) { %s }
	}`, gqlEFOrgUnitFields, gqlEFScopeDecisionFields), map[string]any{
		"id":   createdUnitRootID,
		"head": efStaff.ID,
	})
	if requireChildMap(t, requireOpMap(t, orgQueries, "organizationUnit"), "primary_location")["id"] != createdLocHQID {
		t.Fatalf("expected nested primary_location on OrganizationUnit")
	}

	// 6. Uji PartyService: Party core + Role + Contact + Portal Access + PartyAddress + nested Party fields.
	partnerCode := "SUP-" + strings.ToUpper(tag)
	party := requireMutationChildOK(t, doGQLOK(t, efClient, fmt.Sprintf(`mutation ($in: CreatePartyInput!) {
		createParty(input: $in) { party { %s } %s }
	}`, gqlEFPartyFields, gqlMutationResultFields), map[string]any{
		"in": map[string]any{
			"partner_code":       partnerCode,
			"kind":               "ORGANIZATION",
			"display_name":       "PT Mitra E2E " + tag,
			"email":              tag + "@vendor.test",
			"preferred_language": "id",
		},
	}), "createParty", "party")
	createdPartyID = requireNonEmptyString(t, party, "id")

	_ = requireMutationChildOK(t, doGQLOK(t, efClient, fmt.Sprintf(`mutation ($id: ID!) {
		verifyAndActivateParty(id: $id) { party { id status } %s }
	}`, gqlMutationResultFields), map[string]any{"id": createdPartyID}), "verifyAndActivateParty", "party")

	roleObj := requireMutationChildOK(t, doGQLOK(t, efClient, fmt.Sprintf(`mutation ($in: AssignPartyRoleInput!) {
		assignPartyRole(input: $in) { role { %s } %s }
	}`, gqlEFPartyRoleFields, gqlMutationResultFields), map[string]any{
		"in": map[string]any{"party_id": createdPartyID, "role": "SUPPLIER", "reference_note": "E2E supplier"},
	}), "assignPartyRole", "role")
	createdRoleID = requireNonEmptyString(t, roleObj, "id")

	contactObj := requireMutationChildOK(t, doGQLOK(t, efClient, fmt.Sprintf(`mutation ($in: AddPartyContactInput!) {
		addPartyContact(input: $in) { contact { %s } %s }
	}`, gqlEFPartyContactFields, gqlMutationResultFields), map[string]any{
		"in": map[string]any{
			"party_id":     createdPartyID,
			"full_name":    "Siti Rahayu " + tag,
			"contact_role": "account_lead",
			"email":        "pic_" + tag + "@vendor.test",
			"is_primary":   true,
		},
	}), "addPartyContact", "contact")
	createdContactID = requireNonEmptyString(t, contactObj, "id")

	_ = requireMutationChildOK(t, doGQLOK(t, efClient, fmt.Sprintf(`mutation ($c: ID!, $u: ID!) {
		activatePartyPortalAccess(contact_id: $c, iam_user_id: $u) { contact { id portal_access_status } %s }
	}`, gqlMutationResultFields), map[string]any{"c": createdContactID, "u": efStaff.ID}), "activatePartyPortalAccess", "contact")

	paObj := requireMutationChildOK(t, doGQLOK(t, efClient, fmt.Sprintf(`mutation ($in: LinkPartyAddressInput!) {
		linkPartyAddress(input: $in) { link { %s } %s }
	}`, gqlEFPartyAddressFields, gqlMutationResultFields), map[string]any{
		"in": map[string]any{"party_id": createdPartyID, "address_id": createdAddr1ID, "usage": "REGISTERED_OFFICE", "is_primary": true},
	}), "linkPartyAddress", "link")
	createdPartyAddrID = requireNonEmptyString(t, paObj, "id")

	partyQueries := doGQLOK(t, efClient, fmt.Sprintf(`query ($pID: ID!, $uID: ID!) {
		party(id: $pID) {
			%s
			roles(first: 5) { nodes { id role is_active } total_count }
			contacts(first: 5) { nodes { id full_name portal_access_status } total_count }
			addresses(first: 5) { nodes { id usage address { id city } } total_count }
		}
		evaluateRepresentativePortalAccess(party_id: $pID, iam_user_id: $uID, required_role: SUPPLIER) { %s }
		validatePartyRoleScope(party_id: $pID, required_role: SUPPLIER) { %s }
	}`, gqlEFPartyFields, gqlEFPortalEvaluationFields, gqlEFScopeDecisionFields), map[string]any{
		"pID": createdPartyID,
		"uID": efStaff.ID,
	})
	portalEval := requireOpMap(t, partyQueries, "evaluateRepresentativePortalAccess")
	if !requireBoolField(t, portalEval, "allowed") {
		t.Fatalf("expected portal evaluation allowed=true, got %+v", portalEval)
	}
	pNested := requireOpMap(t, partyQueries, "party")
	if requireNumberField(t, requireChildMap(t, pNested, "roles"), "total_count") != 1 ||
		requireNumberField(t, requireChildMap(t, pNested, "contacts"), "total_count") != 1 ||
		requireNumberField(t, requireChildMap(t, pNested, "addresses"), "total_count") != 1 {
		t.Fatalf("unexpected nested Party connections: %+v", pNested)
	}

	// 7. Uji Relay Query.node polymorphic dispatch lintas seluruh resource EF dan IAM User.
	nodeData := doGQLOK(t, efClient, `query ($a: ID!, $le: ID!, $loc: ID!, $ou: ID!, $p: ID!, $pr: ID!, $pc: ID!, $pa: ID!, $u: ID!) {
		a: node(id: $a) { ... on Address { id city } }
		le: node(id: $le) { ... on LegalEntity { id entity_code } }
		loc: node(id: $loc) { ... on Location { id location_code } }
		ou: node(id: $ou) { ... on OrganizationUnit { id unit_code } }
		p: node(id: $p) { ... on Party { id partner_code } }
		pr: node(id: $pr) { ... on PartyRole { id role } }
		pc: node(id: $pc) { ... on PartyContact { id full_name } }
		pa: node(id: $pa) { ... on PartyAddress { id usage } }
		u: node(id: $u) { ... on User { id kind } }
	}`, map[string]any{
		"a":   createdAddr1ID,
		"le":  createdLERootID,
		"loc": createdLocHQID,
		"ou":  createdUnitRootID,
		"p":   createdPartyID,
		"pr":  createdRoleID,
		"pc":  createdContactID,
		"pa":  createdPartyAddrID,
		"u":   efStaff.ID,
	})
	for _, key := range []string{"a", "le", "loc", "ou", "p", "pr", "pc", "pa", "u"} {
		_ = requireNonEmptyString(t, requireOpMap(t, nodeData, key), "id")
	}
}

func TestEnterpriseFoundation_ScopedRBACEnforcement(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	seedEFPermissionCatalog(t, cfg)
	admin := newSuperAdminBrowser(t, cfg)

	_, creator := provisionEFUserWithPerms(t, admin, cfg, "ef_scope_setup", []string{
		"enterprise_foundation.legal_entities.read",
		"enterprise_foundation.legal_entities.write",
	}, "GLOBAL", nil)

	tag := uniqueTag("efsc")
	leA := requireMutationChildOK(t, doGQLOK(t, creator, fmt.Sprintf(`mutation ($in: CreateLegalEntityInput!) {
		createLegalEntity(input: $in) { legal_entity { id _id entity_code } %s }
	}`, gqlMutationResultFields), map[string]any{
		"in": map[string]any{"entity_code": "SCA-" + tag, "legal_name": "Scope A " + tag, "jurisdiction_country_code": "ID"},
	}), "createLegalEntity", "legal_entity")
	leAID := requireNonEmptyString(t, leA, "id")
	leARawID := requireNonEmptyString(t, leA, "_id")
	_ = requireMutationChildOK(t, doGQLOK(t, creator, fmt.Sprintf(`mutation ($id: ID!) {
		activateLegalEntity(id: $id) { legal_entity { id status } %s }
	}`, gqlMutationResultFields), map[string]any{"id": leAID}), "activateLegalEntity", "legal_entity")

	leB := requireMutationChildOK(t, doGQLOK(t, creator, fmt.Sprintf(`mutation ($in: CreateLegalEntityInput!) {
		createLegalEntity(input: $in) { legal_entity { id _id entity_code } %s }
	}`, gqlMutationResultFields), map[string]any{
		"in": map[string]any{"entity_code": "SCB-" + tag, "legal_name": "Scope B " + tag, "jurisdiction_country_code": "ID"},
	}), "createLegalEntity", "legal_entity")
	leBID := requireNonEmptyString(t, leB, "id")
	leBRawID := requireNonEmptyString(t, leB, "_id")
	_ = requireMutationChildOK(t, doGQLOK(t, creator, fmt.Sprintf(`mutation ($id: ID!) {
		activateLegalEntity(id: $id) { legal_entity { id status } %s }
	}`, gqlMutationResultFields), map[string]any{"id": leBID}), "activateLegalEntity", "legal_entity")

	t.Cleanup(func() {
		cleanupEFEntityRows(t, cfg, "legal_entities", leBRawID)
		cleanupEFEntityRows(t, cfg, "legal_entities", leARawID)
	})

	_, scopedClient := provisionEFUserWithPerms(t, admin, cfg, "ef_scoped_reader", []string{
		"enterprise_foundation.legal_entities.read",
	}, "LEGAL_ENTITY", &leARawID)

	allowedData := doGQLOK(t, scopedClient, `query ($id: ID!) { legalEntity(id: $id) { id entity_code } }`, map[string]any{"id": leAID})
	if requireOpMap(t, allowedData, "legalEntity")["id"] != leAID {
		t.Fatalf("scoped user expected access to LegalEntity A, got %+v", allowedData)
	}
	doGQLError(t, scopedClient, `query ($id: ID!) { legalEntity(id: $id) { id entity_code } }`, map[string]any{"id": leBID}, "FORBIDDEN")
}

func TestEnterpriseFoundation_OutageFaultIsolationAndAutoRecovery(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	seedEFPermissionCatalog(t, cfg)
	admin := newSuperAdminBrowser(t, cfg)
	_, efClient := provisionEFUserWithPerms(t, admin, cfg, "ef_fault", []string{
		"enterprise_foundation.addresses.read",
		"enterprise_foundation.addresses.write",
	}, "GLOBAL", nil)
	_, custClient := createFixtureCustomerAndLogin(t, admin, cfg)

	// Pastikan enterprise-foundation hidup kembali apapun hasil pengujian.
	t.Cleanup(func() {
		_ = exec.Command("docker", "compose", "-f", filepath.Join(cfg.repoRoot, "docker-compose.yml"), "up", "-d", "--wait", "--no-build", "enterprise-foundation").Run()
	})

	// Hentikan container enterprise-foundation -> Gateway dan IAM tetap sehat 100%.
	composeCmd(t, cfg, "stop", "enterprise-foundation")
	assertContainerHealthy(t, "integrione-retail-gateway-1-1")
	assertContainerHealthy(t, "integrione-retail-gateway-2-1")
	assertContainerHealthy(t, "integrione-retail-iam-1")

	// Uji beberapa permintaan beruntun melewati dua replika Gateway saat EF mati total:
	// 1. Gateway health + IAM queries (`me`, `users`) harus tetap sukses 100% dan sesi tidak terputus!
	for i := 0; i < 4; i++ {
		adminRoot := doGQLOK(t, admin, `query { health me { id } users(first: 2) { total_count } }`, nil)
		if adminRoot["health"] != "ok" {
			t.Fatalf("Gateway health != ok during EF outage: %+v", adminRoot)
		}
		_ = requireOpMap(t, adminRoot, "me")

		// 2. Query EF gagal dengan UNAVAILABLE yang disanitasi.
		qErr := doGQLErrorLabeled(t, "EF_OUTAGE_QUERY", efClient, `query { allAddresses(first: 5) { total_count } }`, nil, "UNAVAILABLE")
		assertNoSensitiveUpstreamDetails(t, qErr.Errors[0].Message)

		// 3. Mutation EF mengembalikan HTTP 200 MutationResult { was_successful: false, errors: [{ type: UNAVAILABLE }] }.
		_, msg := doMutationFailLabeled(t, "EF_OUTAGE_MUTATION", efClient, `mutation ($in: CreateAddressInput!) {
			createAddress(input: $in) { errors { type message } was_successful }
		}`, map[string]any{
			"in": map[string]any{"address_line_1": "Outage St 1", "city": "Jakarta", "country_code": "ID"},
		}, "createAddress", "UNAVAILABLE")
		assertNoSensitiveUpstreamDetails(t, msg)

		// 4. Sesi efClient harus tetap valid (bukan logged-out oleh outage EF).
		efMe := doGQLOK(t, efClient, `query { me { id } }`, nil)
		_ = requireOpMap(t, efMe, "me")
	}

	// 5. Customer tanpa izin EF tetap ditolak FORBIDDEN di lapisan RBAC Gateway bahkan saat EF mati.
	doGQLErrorLabeled(t, "EF_OUTAGE_CUSTOMER_RBAC", custClient, `query { allAddresses(first: 5) { total_count } }`, nil, "FORBIDDEN")

	// 6. Nyalakan ulang enterprise-foundation dan verifikasi Gateway pulih otomatis pada kedua replika tanpa restart Gateway/IAM!
	composeCmd(t, cfg, "up", "-d", "--wait", "--no-build", "enterprise-foundation")
	assertContainerHealthy(t, "integrione-retail-enterprise-foundation-1")
	assertContainerHealthy(t, "integrione-retail-gateway-1-1")
	assertContainerHealthy(t, "integrione-retail-gateway-2-1")

	const requiredRecoverySuccesses = 4
	consecutive := 0
	var lastErrs []graphQLError
	for i := 0; i < 24; i++ {
		resp, gql := efClient.doGraphQL(`query { allAddresses(first: 2) { total_count } }`, nil, nil)
		if resp.StatusCode == http.StatusOK && len(gql.Errors) == 0 {
			consecutive++
			if consecutive >= requiredRecoverySuccesses {
				break
			}
			continue
		}
		consecutive = 0
		lastErrs = gql.Errors
		time.Sleep(500 * time.Millisecond)
	}
	if consecutive < requiredRecoverySuccesses {
		t.Fatalf("expected EF queries to auto-recover across Gateway replicas (%d consecutive successes), lastErrs=%+v", requiredRecoverySuccesses, lastErrs)
	}

	// Uji juga Mutation createAddress sukses kembali lalu hapus alamat uji tersebut.
	recTag := uniqueTag("efrec")
	recAddr := requireMutationChildOK(t, doGQLOK(t, efClient, fmt.Sprintf(`mutation ($in: CreateAddressInput!) {
		createAddress(input: $in) { address { id _id city } %s }
	}`, gqlMutationResultFields), map[string]any{
		"in": map[string]any{"address_line_1": "Jl. Pemulihan " + recTag, "city": "Surabaya", "country_code": "ID"},
	}), "createAddress", "address")
	cleanupEFEntityRows(t, cfg, "addresses", requireNonEmptyString(t, recAddr, "_id"))

	// Jalankan waitForGraphQLRecovery untuk memastikan IAM + Gateway juga sehat pasca skenario outage EF.
	waitForGraphQLRecovery(t, admin, "post-EF-outage verification")
}

// TestEnterpriseFoundation_CrossServicePartnerUserProvisioningAndFKE2E menguji provisioning akun VENDOR/SUPPLIER
// yang menautkan FK logis `party_contacts.iam_user_id` di Enterprise Foundation secara end-to-end beserta RBAC dan isolasi outage.
func TestEnterpriseFoundation_CrossServicePartnerUserProvisioningAndFKE2E(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	seedEFPermissionCatalog(t, cfg)
	resetGatewayRateLimits(t, cfg)
	admin := newSuperAdminBrowser(t, cfg)

	mCreateVendor := fmt.Sprintf(`mutation ($in: CreateVendorUserInput!) {
		createVendorUser(input: $in) {
			user { %s }
			party_contact { %s }
			%s
		}
	}`, gqlUserFields, gqlEFPartyContactFields, gqlMutationResultFields)

	mCreateSupplier := fmt.Sprintf(`mutation ($in: CreateSupplierUserInput!) {
		createSupplierUser(input: $in) {
			user { %s }
			party_contact { %s }
			%s
		}
	}`, gqlUserFields, gqlEFPartyContactFields, gqlMutationResultFields)

	// 1. createVendorUser tanpa contact_id otomatis membuat PartyContact baru dengan iam_user_id tertaut (`pending_link`).
	vendorPartyID := createFixturePartyForRole(t, admin, cfg, "VENDOR")
	vendTag := uniqueTag("efv_auto")
	vendEmail := vendTag + "@vendor.integrione.test"
	vendData := doGQLOK(t, admin, mCreateVendor, map[string]any{
		"in": map[string]any{
			"party_id":           vendorPartyID,
			"contact_role":       "billing_pic",
			"job_title":          "Finance Lead",
			"phone":              "+6281200011122",
			"is_primary_contact": true,
			"full_name":          "Rina Vendor " + vendTag,
			"email":              vendEmail,
			"username":           vendTag,
			"password":           "StrongPass#2026!",
		},
	})
	vendPayload := requireMutationSuccess(t, vendData, "createVendorUser")
	vendUser := requireChildMap(t, vendPayload, "user")
	vendContact := requireChildMap(t, vendPayload, "party_contact")
	vendUserRawID := requireNonEmptyString(t, vendUser, "_id")
	vendUserGID := requireNonEmptyString(t, vendUser, "id")
	t.Cleanup(func() { cleanupUserByID(t, cfg, vendUserRawID) })

	if vendContact["iam_user_id"] != vendUserGID || vendContact["party_id"] != vendorPartyID || vendContact["portal_access_status"] != "PENDING_LINK" {
		t.Fatalf("createVendorUser unexpected party_contact payload: %+v", vendContact)
	}
	if dbContactRow := psqlEFScalar(t, cfg, fmt.Sprintf(`SELECT party_id::text || '|' || iam_user_id::text || '|' || portal_access_status::text FROM party_contacts WHERE id = %s::uuid AND deleted_at IS NULL;`, sqlQuote(requireNonEmptyString(t, vendContact, "_id")))); !strings.Contains(dbContactRow, vendUserRawID+"|pending_link") {
		t.Fatalf("postgres party_contacts mismatch for vendor user: %q", dbContactRow)
	}
	if vendKind := psqlScalar(t, cfg, fmt.Sprintf(`SELECT user_kind::text FROM users WHERE id = %s::uuid AND deleted_at IS NULL;`, sqlQuote(vendUserRawID))); vendKind != "vendor" {
		t.Fatalf("iam.users user_kind = %q, want vendor", vendKind)
	}
	vendContactsByIAM := requireOpMap(t, doGQLOK(t, admin, fmt.Sprintf(`query ($u: ID!) {
		partyContactsByIamUserId(iam_user_id: $u, first: 10) { nodes { %s } total_count }
	}`, gqlEFPartyContactFields), map[string]any{"u": vendUserGID}), "partyContactsByIamUserId")
	if len(requireConnectionNodes(t, vendContactsByIAM)) != 1 {
		t.Fatalf("expected 1 partyContactsByIamUserId for vendor user %s, got %+v", vendUserGID, vendContactsByIAM)
	}

	// 2. createSupplierUser menautkan PartyContact yang sudah ada via RequestPartyPortalLink -> verifyEmail -> ActivatePartyPortalAccess -> evaluateRepresentativePortalAccess.
	supplierPartyID := createFixturePartyForRole(t, admin, cfg, "SUPPLIER")
	existingContact := requireMutationChildOK(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($in: AddPartyContactInput!) {
		addPartyContact(input: $in) { contact { %s } %s }
	}`, gqlEFPartyContactFields, gqlMutationResultFields), map[string]any{
		"in": map[string]any{
			"party_id":     supplierPartyID,
			"full_name":    "Hendra Pemasok",
			"contact_role": "account_lead",
			"email":        "hendra_" + vendTag + "@supplier.test",
			"is_primary":   true,
		},
	}), "addPartyContact", "contact")
	existingContactID := requireNonEmptyString(t, existingContact, "id")
	existingContactRawID := requireNonEmptyString(t, existingContact, "_id")

	supTag := uniqueTag("efs_link")
	supEmail := supTag + "@supplier.integrione.test"
	supData := doGQLOK(t, admin, mCreateSupplier, map[string]any{
		"in": map[string]any{
			"party_id":   supplierPartyID,
			"contact_id": existingContactID,
			"full_name":  "Hendra Pemasok",
			"email":      supEmail,
			"username":   supTag,
			"password":   "StrongPass#2026!",
		},
	})
	supPayload := requireMutationSuccess(t, supData, "createSupplierUser")
	supUser := requireChildMap(t, supPayload, "user")
	supContact := requireChildMap(t, supPayload, "party_contact")
	supUserRawID := requireNonEmptyString(t, supUser, "_id")
	supUserGID := requireNonEmptyString(t, supUser, "id")
	t.Cleanup(func() { cleanupUserByID(t, cfg, supUserRawID) })

	if supContact["_id"] != existingContactRawID || supContact["iam_user_id"] != supUserGID || supContact["portal_access_status"] != "PENDING_LINK" {
		t.Fatalf("createSupplierUser did not link existing contact cleanly: %+v", supContact)
	}
	if dbSupRow := psqlEFScalar(t, cfg, fmt.Sprintf(`SELECT id::text || '|' || party_id::text || '|' || iam_user_id::text || '|' || portal_access_status::text FROM party_contacts WHERE id = %s::uuid AND deleted_at IS NULL;`, sqlQuote(existingContactRawID))); !strings.Contains(dbSupRow, existingContactRawID+"|") || !strings.Contains(dbSupRow, "|"+supUserRawID+"|pending_link") {
		t.Fatalf("postgres party_contacts mismatch for supplier user before activation: %q", dbSupRow)
	}
	if supKind := psqlScalar(t, cfg, fmt.Sprintf(`SELECT user_kind::text FROM users WHERE id = %s::uuid AND deleted_at IS NULL;`, sqlQuote(supUserRawID))); supKind != "supplier" {
		t.Fatalf("iam.users user_kind = %q, want supplier", supKind)
	}

	supAnon := newAnonBrowser(t, cfg)
	supVerTok := uniqueTag("ef_sup_ver")
	psqlExec(t, cfg, fmt.Sprintf(
		`UPDATE user_tokens SET token_digest = decode(%s, 'hex') WHERE user_id = %s::uuid AND purpose = 'email_verification' AND status = 'active';`,
		sqlQuote(sha256Hex(supVerTok)), sqlQuote(supUserRawID),
	))
	verRes := requireMutationSuccess(t, doGQLOK(t, supAnon, fmt.Sprintf(`mutation ($in: VerifyEmailInput!) {
		verifyEmail(input: $in) { status verified %s }
	}`, gqlMutationResultFields), map[string]any{"in": map[string]any{"token": supVerTok}}), "verifyEmail")
	if verRes["status"] != "VERIFIED" || !requireBoolField(t, verRes, "verified") {
		t.Fatalf("unexpected verifyEmail result for supplier user: %+v", verRes)
	}
	_ = requireMutationChildOK(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($c: ID!, $u: ID!) {
		activatePartyPortalAccess(contact_id: $c, iam_user_id: $u) { contact { id portal_access_status iam_user_id } %s }
	}`, gqlMutationResultFields), map[string]any{"c": existingContactID, "u": supUserGID}), "activatePartyPortalAccess", "contact")
	if dbSupActiveRow := psqlEFScalar(t, cfg, fmt.Sprintf(`SELECT iam_user_id::text || '|' || portal_access_status::text FROM party_contacts WHERE id = %s::uuid AND deleted_at IS NULL;`, sqlQuote(existingContactRawID))); dbSupActiveRow != supUserRawID+"|active" {
		t.Fatalf("postgres party_contacts mismatch after activation: got %q, want %q", dbSupActiveRow, supUserRawID+"|active")
	}
	supContactsByIAM := requireOpMap(t, doGQLOK(t, admin, fmt.Sprintf(`query ($u: ID!) {
		partyContactsByIamUserId(iam_user_id: $u, first: 10) { nodes { %s } total_count }
	}`, gqlEFPartyContactFields), map[string]any{"u": supUserGID}), "partyContactsByIamUserId")
	if len(requireConnectionNodes(t, supContactsByIAM)) != 1 {
		t.Fatalf("expected 1 partyContactsByIamUserId for supplier user %s, got %+v", supUserGID, supContactsByIAM)
	}

	evalData := doGQLOK(t, admin, fmt.Sprintf(`query ($p: ID!, $u: ID!) {
		evaluateRepresentativePortalAccess(party_id: $p, iam_user_id: $u, required_role: SUPPLIER) { %s }
	}`, gqlEFPortalEvaluationFields), map[string]any{"p": supplierPartyID, "u": supUserGID})
	if !requireBoolField(t, requireOpMap(t, evalData, "evaluateRepresentativePortalAccess"), "allowed") {
		t.Fatalf("expected evaluateRepresentativePortalAccess allowed=true, got %+v", evalData)
	}

	// 3. Domain guards (role mismatch & already linked contact & unlinked worker_id) -> gagal tanpa membuat akun IAM yatim.
	mismatchEmail := uniqueTag("ef_mm") + "@vendor.test"
	_, _ = doMutationFail(t, admin, mCreateVendor, map[string]any{
		"in": map[string]any{
			"party_id":  supplierPartyID,
			"full_name": "Wrong Role",
			"email":     mismatchEmail,
			"username":  uniqueTag("mm"),
			"password":  "StrongPass#2026!",
		},
	}, "createVendorUser", "CONFLICT")
	if cnt := psqlScalar(t, cfg, fmt.Sprintf(`SELECT count(*) FROM user_identities WHERE value = %s;`, sqlQuote(mismatchEmail))); cnt != "0" {
		t.Fatalf("expected 0 IAM rows created on party role mismatch, got %s", cnt)
	}

	dupContactEmail := uniqueTag("ef_dc") + "@supplier.test"
	_, _ = doMutationFail(t, admin, mCreateSupplier, map[string]any{
		"in": map[string]any{
			"party_id":   supplierPartyID,
			"contact_id": existingContactID,
			"full_name":  "Already Linked",
			"email":      dupContactEmail,
			"username":   uniqueTag("dc"),
			"password":   "StrongPass#2026!",
		},
	}, "createSupplierUser", "CONFLICT")
	if cnt := psqlScalar(t, cfg, fmt.Sprintf(`SELECT count(*) FROM user_identities WHERE value = %s;`, sqlQuote(dupContactEmail))); cnt != "0" {
		t.Fatalf("expected 0 IAM rows created for already-linked contact, got %s", cnt)
	}

	empWorkerEmail := uniqueTag("ef_wk") + "@integrione.test"
	mCreateEmp := fmt.Sprintf(`mutation ($in: CreateEmployeeUserInput!) { createEmployeeUser(input: $in) { %s } }`, gqlUserPayloadFields)
	_, _ = doMutationFail(t, admin, mCreateEmp, map[string]any{
		"in": map[string]any{
			"worker_id":               "55555555-5555-4555-8555-555555555555",
			"provisioning_request_id": "66666666-6666-4666-8666-666666666666",
			"full_name":               "HRM Standby",
			"email":                   empWorkerEmail,
			"username":                uniqueTag("wk"),
			"password":                "StrongPass#2026!",
		},
	}, "createEmployeeUser", "AUTHORIZATION", "NOT_FOUND", "UNAVAILABLE")
	if cnt := psqlScalar(t, cfg, fmt.Sprintf(`SELECT count(*) FROM user_identities WHERE value = %s;`, sqlQuote(empWorkerEmail))); cnt != "0" {
		t.Fatalf("expected 0 IAM rows created when HRM worker linking unavailable, got %s", cnt)
	}

	// 4. Batas RBAC TestAB: operator IAM tanpa enterprise_foundation.parties.write ditolak AUTHORIZATION dengan 0 insert.
	_, iamOnlyClient := provisionEFUserWithPerms(t, admin, cfg, "iam_no_ef", []string{"iam.users.write"}, "GLOBAL", nil)
	noEFEmail := uniqueTag("noef") + "@vendor.test"
	_, _ = doMutationFail(t, iamOnlyClient, mCreateVendor, map[string]any{
		"in": map[string]any{
			"party_id":  vendorPartyID,
			"full_name": "No EF Perm",
			"email":     noEFEmail,
			"username":  uniqueTag("noef"),
			"password":  "StrongPass#2026!",
		},
	}, "createVendorUser", "AUTHORIZATION")
	if cnt := psqlScalar(t, cfg, fmt.Sprintf(`SELECT count(*) FROM user_identities WHERE value = %s;`, sqlQuote(noEFEmail))); cnt != "0" {
		t.Fatalf("expected 0 IAM rows when operator lacks EF permission, got %s", cnt)
	}

	// 5. Outage EF saat provisioning: createVendorUser -> UNAVAILABLE + 0 IAM user yatim, IAM/createEmployeeUser tetap jalan.
	t.Cleanup(func() {
		_ = exec.Command("docker", "compose", "-f", filepath.Join(cfg.repoRoot, "docker-compose.yml"), "up", "-d", "--wait", "--no-build", "enterprise-foundation").Run()
	})
	composeCmd(t, cfg, "stop", "enterprise-foundation")

	outageEmail := uniqueTag("ef_down") + "@vendor.test"
	_, outageMsg := doMutationFail(t, admin, mCreateVendor, map[string]any{
		"in": map[string]any{
			"party_id":  vendorPartyID,
			"full_name": "During Outage",
			"email":     outageEmail,
			"username":  uniqueTag("efdown"),
			"password":  "StrongPass#2026!",
		},
	}, "createVendorUser", "UNAVAILABLE")
	assertNoSensitiveUpstreamDetails(t, outageMsg)
	if cnt := psqlScalar(t, cfg, fmt.Sprintf(`SELECT count(*) FROM user_identities WHERE value = %s;`, sqlQuote(outageEmail))); cnt != "0" {
		t.Fatalf("expected 0 orphan IAM rows during EF outage, got %s", cnt)
	}

	// Employee user tanpa worker_id tetap dapat dibuat di IAM saat EF mati.
	outageEmp := createFixtureUser(t, admin, cfg, "EMPLOYEE", true)
	if outageEmp.RawID == "" {
		t.Fatalf("expected EMPLOYEE creation without worker_id to stay operational during EF outage")
	}

	composeCmd(t, cfg, "up", "-d", "--wait", "--no-build", "enterprise-foundation")
	assertContainerHealthy(t, "integrione-retail-enterprise-foundation-1")
	waitForGraphQLRecovery(t, admin, "post-partner-outage IAM recovery")
	waitForGraphQLQueryRecovery(t, admin, `query { allAddresses(first: 1) { total_count } }`, "post-partner-outage EF recovery")
}
