//go:build e2e

package e2e

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// TestEnterpriseFoundation_SuperadminAllRolesPermissionsAndOperations memverifikasi bahwa
// akun human superadmin memegang semua izin enterprise_foundation.*, dapat menetapkan dan memegang
// seluruh peran operasional manusia Enterprise Foundation (GLOBAL dan 5 scope objek eksternal EF),
// serta mampu mengeksekusi seluruh 64 query dan 48 mutation Enterprise Foundation di Gateway /api/graphql.
func TestEnterpriseFoundation_SuperadminAllRolesPermissionsAndOperations(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	seedEFPermissionCatalog(t, cfg)
	admin := newSuperAdminBrowser(t, cfg)

	tag := uniqueTag("efall")

	// =========================================================================
	// 1. Verifikasi Peran & Izin Enterprise Foundation pada Superadmin
	// =========================================================================
	ctxData := doGQLOK(t, admin, `query SuperadminEFContext {
		me { id _id kind status display_name }
		myAccessSnapshot {
			principal_id user_kind status
			roles { id code name priority default_redirect_path }
			permissions { permission_code is_sensitive scope { type id } }
		}
	}`, nil)
	meMap := requireOpMap(t, ctxData, "me")
	superadminUserID := requireNonEmptyString(t, meMap, "id")
	snapshotMap := requireOpMap(t, ctxData, "myAccessSnapshot")

	wantEFPermissions := map[string]bool{
		"enterprise_foundation.addresses.read":           true,
		"enterprise_foundation.addresses.write":          true,
		"enterprise_foundation.legal_entities.read":      true,
		"enterprise_foundation.legal_entities.write":     true,
		"enterprise_foundation.locations.read":           true,
		"enterprise_foundation.locations.write":          true,
		"enterprise_foundation.organization_units.read":  true,
		"enterprise_foundation.organization_units.write": true,
		"enterprise_foundation.parties.read":             true,
		"enterprise_foundation.parties.write":            true,
	}
	foundEFPermissions := map[string]bool{}
	rawPerms, _ := snapshotMap["permissions"].([]any)
	for _, rawP := range rawPerms {
		pm, _ := rawP.(map[string]any)
		code, _ := pm["permission_code"].(string)
		if wantEFPermissions[code] {
			foundEFPermissions[code] = true
		}
	}
	if len(foundEFPermissions) != len(wantEFPermissions) {
		t.Fatalf("superadmin EF permissions = %v (%d), want all %d", foundEFPermissions, len(foundEFPermissions), len(wantEFPermissions))
	}

	rolesData := doGQLOK(t, admin, `query ListEFRoles {
		roles(management_domain: "enterprise_foundation", active_only: false, first: 20) {
			total_count
			nodes { id _id code name management_domain default_redirect_path redirect_priority is_system is_active }
		}
	}`, nil)
	rolesConn := requireOpMap(t, rolesData, "roles")
	rawRoles, _ := rolesConn["nodes"].([]any)
	efRoleByCode := map[string]map[string]any{}
	for _, rawR := range rawRoles {
		rm, _ := rawR.(map[string]any)
		code, _ := rm["code"].(string)
		efRoleByCode[code] = rm
	}

	expectedRoleCounts := map[string]int{
		"ef_admin":             10,
		"ef_operator":          10,
		"ef_structure_manager": 8,
		"ef_partner_manager":   4,
		"ef_readonly":          5,
		"ef_automation":        10,
	}
	for roleCode, wantPermCount := range expectedRoleCounts {
		rm, ok := efRoleByCode[roleCode]
		if !ok {
			t.Fatalf("missing EF role %q in roles(management_domain: enterprise_foundation); got keys=%v", roleCode, efRoleByCode)
		}
		roleID := requireNonEmptyString(t, rm, "id")
		rpData := doGQLOK(t, admin, `query RolePerms($id: ID!) {
			rolePermissions(role_id: $id, first: 50) { total_count nodes { code } }
		}`, map[string]any{"id": roleID})
		rpConn := requireOpMap(t, rpData, "rolePermissions")
		if gotCount := int(rpConn["total_count"].(float64)); gotCount != wantPermCount {
			t.Fatalf("role %q permission count = %d, want %d", roleCode, gotCount, wantPermCount)
		}
	}

	// Tolak penugasan role otomasi (SYSTEM-only) ke akun human superadmin secara fail-closed.
	efAutoRoleID := requireNonEmptyString(t, efRoleByCode["ef_automation"], "id")
	respBadAuto, gqlBadAuto := admin.doGraphQL(fmt.Sprintf(`mutation ($in: AssignUserRoleInput!) {
		assignUserRole(input: $in) { user_role_edge { assignment_id } %s }
	}`, gqlMutationResultFields), map[string]any{
		"in": map[string]any{
			"user_id":    superadminUserID,
			"role_id":    efAutoRoleID,
			"scope_type": "GLOBAL",
		},
	}, nil)
	if respBadAuto.StatusCode != 200 {
		t.Fatalf("unexpected HTTP status %d", respBadAuto.StatusCode)
	}
	requireMutationPayloadError(t, gqlBadAuto, "assignUserRole", "AUTHORIZATION")

	// Pastikan superadmin dapat memegang seluruh role operasional EF (GLOBAL) lalu bersihkan setelah uji.
	existingRolesData := doGQLOK(t, admin, `query ExistingSuperadminRoles($uid: ID!) {
		userRoles(user_id: $uid, include_expired: true, first: 50) {
			nodes { id code }
			edges { assignment_id scope { type id } node { code } }
		}
	}`, map[string]any{"uid": superadminUserID})
	existingEdgeMap := map[string]string{}
	if urConn, ok := existingRolesData["userRoles"].(map[string]any); ok {
		if edges, ok := urConn["edges"].([]any); ok {
			for _, rawE := range edges {
				em, _ := rawE.(map[string]any)
				sc, _ := em["scope"].(map[string]any)
				node, _ := em["node"].(map[string]any)
				rCode, _ := node["code"].(string)
				aID, _ := em["assignment_id"].(string)
				if scType, _ := sc["type"].(string); scType == "GLOBAL" && aID != "" {
					existingEdgeMap[rCode] = aID
				}
			}
		}
	}

	var cleanupAssignments []string
	t.Cleanup(func() {
		resetGatewayRateLimits(t, cfg)
		for _, assignID := range cleanupAssignments {
			_, _ = admin.doGraphQL(fmt.Sprintf(`mutation ($id: ID!) {
				revokeUserRole(assignment_id: $id) { revoked %s }
			}`, gqlMutationResultFields), map[string]any{"id": assignID}, nil)
		}
	})
	humanEFRoleCodes := []string{"ef_admin", "ef_operator", "ef_structure_manager", "ef_partner_manager", "ef_readonly"}
	for _, roleCode := range humanEFRoleCodes {
		if staleAssignID, exists := existingEdgeMap[roleCode]; exists {
			_ = requireMutationSuccess(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($id: ID!) {
				revokeUserRole(assignment_id: $id) { revoked %s }
			}`, gqlMutationResultFields), map[string]any{"id": staleAssignID}), "revokeUserRole")
		}
		roleID := requireNonEmptyString(t, efRoleByCode[roleCode], "id")
		edge := requireMutationChildOK(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($in: AssignUserRoleInput!) {
			assignUserRole(input: $in) { user_role_edge { assignment_id user_id scope { type id } node { code } } %s }
		}`, gqlMutationResultFields), map[string]any{
			"in": map[string]any{
				"user_id":    superadminUserID,
				"role_id":    roleID,
				"scope_type": "GLOBAL",
			},
		}), "assignUserRole", "user_role_edge")
		cleanupAssignments = append(cleanupAssignments, requireNonEmptyString(t, edge, "assignment_id"))
	}

	// Sesuai aturan verifikasi HTTPS pasca-assign role: wajib login ulang pada sesi baru agar myAccessSnapshot dan operasi EF memakai otorisasi segar.
	reloggedAdmin := newSuperAdminBrowser(t, cfg)
	relogCtx := doGQLOK(t, reloggedAdmin, `query ReloggedSuperadminSnapshot {
		myAccessSnapshot {
			principal_id user_kind status can_access_identity
			roles { code }
			permissions { permission_code scope { type id } }
		}
		allAddresses(first: 1) { total_count }
	}`, nil)
	relogSnap := requireOpMap(t, relogCtx, "myAccessSnapshot")
	if !requireBoolField(t, relogSnap, "can_access_identity") {
		t.Fatalf("relogged superadmin expected can_access_identity=true, got %+v", relogSnap)
	}
	relogRoles, _ := relogSnap["roles"].([]any)
	hasEFAdminAfterRelogin := false
	for _, r := range relogRoles {
		if rm, _ := r.(map[string]any); rm != nil && rm["code"] == "ef_admin" {
			hasEFAdminAfterRelogin = true
		}
	}
	if !hasEFAdminAfterRelogin {
		t.Fatalf("relogged superadmin snapshot missing ef_admin after role assignments: %+v", relogRoles)
	}

	// =========================================================================
	// 2. Domain Address: Seluruh 5 Query + 4 Mutation
	// =========================================================================
	addr1 := requireMutationChildOK(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($in: CreateAddressInput!) {
		createAddress(input: $in) { address { id _id label address_line_1 city province postal_code country_code } %s }
	}`, gqlMutationResultFields), map[string]any{
		"in": map[string]any{
			"label":             "HQ-" + tag,
			"address_line_1":    "Jl. Sudirman Kav 10 " + tag,
			"address_line_2":    "Lantai 12",
			"city":              "Jakarta Pusat",
			"province":          "DKI Jakarta",
			"postal_code":       "10220",
			"country_code":      "ID",
			"formatted_address": "Jl. Sudirman Kav 10, Jakarta Pusat " + tag,
			"latitude":          -6.2115,
			"longitude":         106.8219,
		},
	}), "createAddress", "address")
	addr1ID := requireNonEmptyString(t, addr1, "id")
	addr1RawID := requireNonEmptyString(t, addr1, "_id")

	addrTemp := requireMutationChildOK(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($in: FindOrCreateAddressInput!) {
		findOrCreateAddress(input: $in) { address { id _id address_line_1 } created %s }
	}`, gqlMutationResultFields), map[string]any{
		"in": map[string]any{
			"label":          "TEMP-" + tag,
			"address_line_1": "Jl. Gatot Subroto " + tag,
			"city":           "Jakarta Selatan",
			"province":       "DKI Jakarta",
			"postal_code":    "12930",
			"country_code":   "ID",
		},
	}), "findOrCreateAddress", "address")
	addrTempID := requireNonEmptyString(t, addrTemp, "id")
	addrTempRawID := requireNonEmptyString(t, addrTemp, "_id")

	t.Cleanup(func() {
		cleanupEFEntityRows(t, cfg, "addresses", addrTempRawID)
		cleanupEFEntityRows(t, cfg, "addresses", addr1RawID)
	})

	_ = requireMutationChildOK(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($in: UpdateAddressInput!) {
		updateAddress(input: $in) { address { id label city } %s }
	}`, gqlMutationResultFields), map[string]any{
		"in": map[string]any{
			"id":             addr1ID,
			"label":          "HQ-UPD-" + tag,
			"address_line_1": "Jl. Sudirman Kav 10 Updated " + tag,
			"city":           "Jakarta Pusat",
			"province":       "DKI Jakarta",
			"postal_code":    "10220",
			"country_code":   "ID",
		},
	}), "updateAddress", "address")

	_ = doGQLOK(t, admin, `query AddressQueries($id: ID!, $ids: [ID!]!, $q: String) {
		address(id: $id) { id _id label city country_code }
		addressesByIds(ids: $ids) { id label }
		addresses(query: $q, country_code: "ID", city: "Jakarta Pusat", first: 5) { total_count nodes { id label } }
		allAddresses(first: 5) { total_count nodes { id } }
		isAddressReferenced(id: $id)
		node(id: $id) { id ... on Address { label city } }
	}`, map[string]any{"id": addrTempID, "ids": []string{addr1ID, addrTempID}, "q": tag})

	delAddrPayload := requireMutationSuccess(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($id: ID!) {
		deleteAddress(id: $id) { deleted %s }
	}`, gqlMutationResultFields), map[string]any{"id": addrTempID}), "deleteAddress")
	if deleted, _ := delAddrPayload["deleted"].(bool); !deleted {
		t.Fatalf("deleteAddress expected deleted=true, got %+v", delAddrPayload)
	}

	// =========================================================================
	// 3. Domain LegalEntity: Seluruh 10 Query + 7 Mutation
	// =========================================================================
	leParent := requireMutationChildOK(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($in: CreateLegalEntityInput!) {
		createLegalEntity(input: $in) { legal_entity { id _id entity_code status } %s }
	}`, gqlMutationResultFields), map[string]any{
		"in": map[string]any{
			"entity_code":               "LEP-" + tag,
			"legal_name":                "PT IntegriOne Holding " + tag,
			"trading_name":              "IntegriOne " + tag,
			"legal_form":                "PT",
			"jurisdiction_country_code": "ID",
			"nib":                       "912000" + tag[len(tag)-4:],
			"tax_id_number":             "01999888" + tag[len(tag)-4:],
			"is_pkp":                    true,
			"registered_address_id":     addr1ID,
			"fiscal_year_start_month":   1,
		},
	}), "createLegalEntity", "legal_entity")
	leParentID := requireNonEmptyString(t, leParent, "id")
	leParentRawID := requireNonEmptyString(t, leParent, "_id")

	leChild := requireMutationChildOK(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($in: CreateLegalEntityInput!) {
		createLegalEntity(input: $in) { legal_entity { id _id entity_code status } %s }
	}`, gqlMutationResultFields), map[string]any{
		"in": map[string]any{
			"parent_legal_entity_id":    leParentID,
			"entity_code":               "LEC-" + tag,
			"legal_name":                "PT IntegriOne Subsidiary " + tag,
			"jurisdiction_country_code": "ID",
		},
	}), "createLegalEntity", "legal_entity")
	leChildID := requireNonEmptyString(t, leChild, "id")
	leChildRawID := requireNonEmptyString(t, leChild, "_id")

	t.Cleanup(func() {
		cleanupEFEntityRows(t, cfg, "legal_entities", leChildRawID)
		cleanupEFEntityRows(t, cfg, "legal_entities", leParentRawID)
	})

	// Uji negatif ScopeOwnerVerifier IAM: PENDING LegalEntity wajib ditolak CONFLICT.
	efAdminRoleID := requireNonEmptyString(t, efRoleByCode["ef_admin"], "id")
	_, gqlPendingScope := admin.doGraphQL(fmt.Sprintf(`mutation ($in: AssignUserRoleInput!) {
		assignUserRole(input: $in) { user_role_edge { assignment_id } %s }
	}`, gqlMutationResultFields), map[string]any{
		"in": map[string]any{
			"user_id":    superadminUserID,
			"role_id":    efAdminRoleID,
			"scope_type": "LEGAL_ENTITY",
			"scope_id":   leParentRawID,
		},
	}, nil)
	requireMutationPayloadError(t, gqlPendingScope, "assignUserRole", "CONFLICT")

	// Uji negatif ScopeOwnerVerifier IAM: target UUID fiktif wajib ditolak NOT_FOUND.
	_, gqlMissingScope := admin.doGraphQL(fmt.Sprintf(`mutation ($in: AssignUserRoleInput!) {
		assignUserRole(input: $in) { user_role_edge { assignment_id } %s }
	}`, gqlMutationResultFields), map[string]any{
		"in": map[string]any{
			"user_id":    superadminUserID,
			"role_id":    efAdminRoleID,
			"scope_type": "LEGAL_ENTITY",
			"scope_id":   "00000000-0000-4000-8000-000000000999",
		},
	}, nil)
	requireMutationPayloadError(t, gqlMissingScope, "assignUserRole", "NOT_FOUND")

	// Aktifkan kedua badan hukum + update + primary + suspend + reactivate + dissolve + delete child.
	_ = requireMutationChildOK(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($id: ID!) {
		activateLegalEntity(id: $id) { legal_entity { id status } %s }
	}`, gqlMutationResultFields), map[string]any{"id": leParentID}), "activateLegalEntity", "legal_entity")

	_ = requireMutationChildOK(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($in: UpdateLegalEntityInput!) {
		updateLegalEntity(input: $in) { legal_entity { id legal_name } %s }
	}`, gqlMutationResultFields), map[string]any{
		"in": map[string]any{
			"id":                        leParentID,
			"legal_name":                "PT IntegriOne Holding Tbk " + tag,
			"trading_name":              "IntegriOne Retail " + tag,
			"legal_form":                "PT",
			"jurisdiction_country_code": "ID",
			"nib":                       "912000" + tag[len(tag)-4:],
			"tax_id_number":             "01999888" + tag[len(tag)-4:],
			"is_pkp":                    true,
			"registered_address_id":     addr1ID,
			"fiscal_year_start_month":   1,
		},
	}), "updateLegalEntity", "legal_entity")

	_ = requireMutationChildOK(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($id: ID!) {
		setPrimaryOperatingLegalEntity(id: $id) { legal_entity { id is_primary_operating_entity } %s }
	}`, gqlMutationResultFields), map[string]any{"id": leParentID}), "setPrimaryOperatingLegalEntity", "legal_entity")

	_ = requireMutationChildOK(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($id: ID!) {
		activateLegalEntity(id: $id) { legal_entity { id status } %s }
	}`, gqlMutationResultFields), map[string]any{"id": leChildID}), "activateLegalEntity", "legal_entity")

	_ = requireMutationChildOK(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($id: ID!) {
		suspendLegalEntity(id: $id) { legal_entity { id status } %s }
	}`, gqlMutationResultFields), map[string]any{"id": leChildID}), "suspendLegalEntity", "legal_entity")

	dissolveDate := time.Now().UTC().Format(time.RFC3339)
	_ = requireMutationChildOK(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($id: ID!, $d: Time) {
		dissolveLegalEntity(id: $id, dissolution_date: $d) { legal_entity { id status } %s }
	}`, gqlMutationResultFields), map[string]any{"id": leChildID, "d": dissolveDate}), "dissolveLegalEntity", "legal_entity")

	_ = doGQLOK(t, admin, `query LegalEntityAllQueries($pid: ID!, $ids: [ID!]!, $code: String!, $npwp: String!, $nib: String!, $q: String) {
		legalEntity(id: $pid) { id entity_code registered_address { id city } children(first: 5) { total_count } }
		legalEntitiesByIds(ids: $ids) { id entity_code }
		legalEntityByCode(entity_code: $code) { id entity_code }
		legalEntityByTaxIdNumber(tax_id_number: $npwp) { id tax_id_number }
		legalEntitiesByNib(nib: $nib, first: 5) { total_count nodes { id } }
		primaryOperatingLegalEntity { id is_primary_operating_entity }
		legalEntities(query: $q, status: ACTIVE, first: 5) { total_count nodes { id } }
		allLegalEntities(first: 5) { total_count nodes { id } }
		legalEntityChildren(parent_legal_entity_id: $pid, first: 5) { total_count nodes { id } }
		validateLegalEntityScopeTarget(id: $pid) { eligible scope_kind target_id code legal_entity_id }
		node(id: $pid) { id ... on LegalEntity { entity_code legal_name } }
	}`, map[string]any{
		"pid":  leParentID,
		"ids":  []string{leParentID, leChildID},
		"code": "LEP-" + tag,
		"npwp": "01999888" + tag[len(tag)-4:],
		"nib":  "912000" + tag[len(tag)-4:],
		"q":    tag,
	})

	delLEPayload := requireMutationSuccess(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($id: ID!) {
		deleteLegalEntity(id: $id) { deleted %s }
	}`, gqlMutationResultFields), map[string]any{"id": leChildID}), "deleteLegalEntity")
	if deleted, _ := delLEPayload["deleted"].(bool); !deleted {
		t.Fatalf("deleteLegalEntity expected deleted=true, got %+v", delLEPayload)
	}

	// =========================================================================
	// 4. Domain Location: Seluruh 10 Query + 5 Mutation
	// =========================================================================
	resetGatewayRateLimits(t, cfg)
	locParent := requireMutationChildOK(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($in: CreateLocationInput!) {
		createLocation(input: $in) { location { id _id location_code kind is_active } %s }
	}`, gqlMutationResultFields), map[string]any{
		"in": map[string]any{
			"legal_entity_id":             leParentID,
			"address_id":                  addr1ID,
			"location_code":               "LOCP-" + tag,
			"name":                        "Gudang Pusat " + tag,
			"kind":                        "WAREHOUSE",
			"phone":                       "+62215550101",
			"timezone":                    "Asia/Jakarta",
			"operational_attributes_json": `{"dock_count":6}`,
		},
	}), "createLocation", "location")
	locParentID := requireNonEmptyString(t, locParent, "id")
	locParentRawID := requireNonEmptyString(t, locParent, "_id")

	locChild := requireMutationChildOK(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($in: CreateLocationInput!) {
		createLocation(input: $in) { location { id _id location_code } %s }
	}`, gqlMutationResultFields), map[string]any{
		"in": map[string]any{
			"legal_entity_id":    leParentID,
			"parent_location_id": locParentID,
			"address_id":         addr1ID,
			"location_code":      "LOCC-" + tag,
			"name":               "Toko Cabang " + tag,
			"kind":               "RETAIL_STORE",
			"timezone":           "Asia/Jakarta",
		},
	}), "createLocation", "location")
	locChildID := requireNonEmptyString(t, locChild, "id")
	locChildRawID := requireNonEmptyString(t, locChild, "_id")

	t.Cleanup(func() {
		cleanupEFEntityRows(t, cfg, "locations", locChildRawID)
		cleanupEFEntityRows(t, cfg, "locations", locParentRawID)
	})

	_ = requireMutationChildOK(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($in: UpdateLocationInput!) {
		updateLocation(input: $in) { location { id name } %s }
	}`, gqlMutationResultFields), map[string]any{
		"in": map[string]any{
			"id":       locParentID,
			"name":     "Gudang Pusat Distribusi " + tag,
			"kind":     "WAREHOUSE",
			"timezone": "Asia/Jakarta",
		},
	}), "updateLocation", "location")

	_ = requireMutationChildOK(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($id: ID!) {
		deactivateLocation(id: $id) { location { id is_active } %s }
	}`, gqlMutationResultFields), map[string]any{"id": locChildID}), "deactivateLocation", "location")

	_ = requireMutationChildOK(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($id: ID!) {
		activateLocation(id: $id) { location { id is_active } %s }
	}`, gqlMutationResultFields), map[string]any{"id": locChildID}), "activateLocation", "location")

	_ = doGQLOK(t, admin, `query LocationAllQueries($id: ID!, $ids: [ID!]!, $code: String!, $le: ID!, $addr: ID!, $q: String) {
		location(id: $id) { id location_code legal_entity { id } address { id } children(first: 5) { total_count } }
		locationsByIds(ids: $ids) { id location_code }
		locationByCode(location_code: $code) { id location_code }
		locationsByLegalEntityAndKind(legal_entity_id: $le, kind: WAREHOUSE, active_only: true, first: 5) { total_count nodes { id } }
		locationsByAddressId(address_id: $addr, first: 5) { total_count nodes { id } }
		locationChildren(parent_location_id: $id, first: 5) { total_count nodes { id } }
		locations(legal_entity_id: $le, query: $q, active_only: true, first: 5) { total_count nodes { id } }
		allLocations(first: 5) { total_count nodes { id } }
		validateFacilityScope(id: $id) { eligible scope_kind target_id }
		validateWarehouseScope(id: $id) { eligible scope_kind target_id }
		node(id: $id) { id ... on Location { location_code name } }
	}`, map[string]any{
		"id":   locParentID,
		"ids":  []string{locParentID, locChildID},
		"code": "LOCP-" + tag,
		"le":   leParentID,
		"addr": addr1ID,
		"q":    tag,
	})

	delLocPayload := requireMutationSuccess(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($id: ID!) {
		deleteLocation(id: $id) { deleted %s }
	}`, gqlMutationResultFields), map[string]any{"id": locChildID}), "deleteLocation")
	if deleted, _ := delLocPayload["deleted"].(bool); !deleted {
		t.Fatalf("deleteLocation expected deleted=true, got %+v", delLocPayload)
	}

	// =========================================================================
	// 5. Domain OrganizationUnit: Seluruh 9 Query + 6 Mutation
	// =========================================================================
	ouParent := requireMutationChildOK(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($in: CreateOrganizationUnitInput!) {
		createOrganizationUnit(input: $in) { organization_unit { id _id unit_code kind is_active } %s }
	}`, gqlMutationResultFields), map[string]any{
		"in": map[string]any{
			"legal_entity_id":     leParentID,
			"primary_location_id": locParentID,
			"unit_code":           "OUP-" + tag,
			"name":                "Divisi Operasional " + tag,
			"kind":                "DIVISION",
			"cost_center_code":    "CC-OPS-" + tag[len(tag)-4:],
			"head_person_id":      superadminUserID,
		},
	}), "createOrganizationUnit", "organization_unit")
	ouParentID := requireNonEmptyString(t, ouParent, "id")
	ouParentRawID := requireNonEmptyString(t, ouParent, "_id")

	ouChild := requireMutationChildOK(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($in: CreateOrganizationUnitInput!) {
		createOrganizationUnit(input: $in) { organization_unit { id _id unit_code } %s }
	}`, gqlMutationResultFields), map[string]any{
		"in": map[string]any{
			"legal_entity_id": leParentID,
			"parent_unit_id":  ouParentID,
			"unit_code":       "OUC-" + tag,
			"name":            "Departemen Gudang " + tag,
			"kind":            "DEPARTMENT",
		},
	}), "createOrganizationUnit", "organization_unit")
	ouChildID := requireNonEmptyString(t, ouChild, "id")
	ouChildRawID := requireNonEmptyString(t, ouChild, "_id")

	t.Cleanup(func() {
		cleanupEFEntityRows(t, cfg, "organization_units", ouChildRawID)
		cleanupEFEntityRows(t, cfg, "organization_units", ouParentRawID)
	})

	_ = requireMutationChildOK(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($in: UpdateOrganizationUnitInput!) {
		updateOrganizationUnit(input: $in) { organization_unit { id name } %s }
	}`, gqlMutationResultFields), map[string]any{
		"in": map[string]any{
			"id":                  ouParentID,
			"primary_location_id": locParentID,
			"name":                "Divisi Operasional Ritel " + tag,
			"kind":                "DIVISION",
			"cost_center_code":    "CC-RTL-" + tag[len(tag)-4:],
		},
	}), "updateOrganizationUnit", "organization_unit")

	_ = requireMutationChildOK(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($id: ID!, $hp: ID) {
		assignOrganizationUnitHeadPerson(id: $id, head_person_id: $hp) { organization_unit { id head_person_id } %s }
	}`, gqlMutationResultFields), map[string]any{"id": ouParentID, "hp": superadminUserID}), "assignOrganizationUnitHeadPerson", "organization_unit")

	_ = requireMutationChildOK(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($id: ID!) {
		deactivateOrganizationUnit(id: $id) { organization_unit { id is_active } %s }
	}`, gqlMutationResultFields), map[string]any{"id": ouChildID}), "deactivateOrganizationUnit", "organization_unit")

	_ = requireMutationChildOK(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($id: ID!) {
		activateOrganizationUnit(id: $id) { organization_unit { id is_active } %s }
	}`, gqlMutationResultFields), map[string]any{"id": ouChildID}), "activateOrganizationUnit", "organization_unit")

	_ = doGQLOK(t, admin, `query OrgUnitAllQueries($id: ID!, $ids: [ID!]!, $le: ID!, $uc: String!, $hp: ID!, $loc: ID!, $q: String) {
		organizationUnit(id: $id) { id unit_code legal_entity { id } primary_location { id } children(first: 5) { total_count } }
		organizationUnitsByIds(ids: $ids) { id unit_code }
		organizationUnitByEntityAndCode(legal_entity_id: $le, unit_code: $uc) { id unit_code }
		organizationUnitsByHeadPersonId(head_person_id: $hp, first: 5) { total_count nodes { id } }
		organizationUnitsByPrimaryLocationId(primary_location_id: $loc, first: 5) { total_count nodes { id } }
		organizationUnitChildren(parent_unit_id: $id, first: 5) { total_count nodes { id } }
		organizationUnits(legal_entity_id: $le, query: $q, active_only: true, first: 5) { total_count nodes { id } }
		allOrganizationUnits(first: 5) { total_count nodes { id } }
		validateOrganizationUnitScopeTarget(id: $id) { eligible scope_kind target_id legal_entity_id }
		node(id: $id) { id ... on OrganizationUnit { unit_code name } }
	}`, map[string]any{
		"id":  ouParentID,
		"ids": []string{ouParentID, ouChildID},
		"le":  leParentID,
		"uc":  "OUP-" + tag,
		"hp":  superadminUserID,
		"loc": locParentID,
		"q":   tag,
	})

	delOUPayload := requireMutationSuccess(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($id: ID!) {
		deleteOrganizationUnit(id: $id) { deleted %s }
	}`, gqlMutationResultFields), map[string]any{"id": ouChildID}), "deleteOrganizationUnit")
	if deleted, _ := delOUPayload["deleted"].(bool); !deleted {
		t.Fatalf("deleteOrganizationUnit expected deleted=true, got %+v", delOUPayload)
	}

	// =========================================================================
	// 6. Domain Party, PartyRole, PartyContact+Portal, PartyAddressLink (30Q + 26M)
	// =========================================================================
	resetGatewayRateLimits(t, cfg)
	partyActive := requireMutationChildOK(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($in: CreatePartyInput!) {
		createParty(input: $in) { party { id _id partner_code status } %s }
	}`, gqlMutationResultFields), map[string]any{
		"in": map[string]any{
			"partner_code":       "PRT-" + tag,
			"kind":               "ORGANIZATION",
			"display_name":       "CV Tekstil Nusantara " + tag,
			"legal_name":         "CV Tekstil Nusantara Jaya " + tag,
			"email":              "vendor-" + tag + "@partner.test",
			"phone":              "+628123450001",
			"tax_id_number":      "88776655" + tag[len(tag)-4:],
			"is_pkp":             true,
			"preferred_language": "id",
		},
	}), "createParty", "party")
	partyID := requireNonEmptyString(t, partyActive, "id")
	partyRawID := requireNonEmptyString(t, partyActive, "_id")

	partyTemp := requireMutationChildOK(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($in: CreatePartyInput!) {
		createParty(input: $in) { party { id _id partner_code status } %s }
	}`, gqlMutationResultFields), map[string]any{
		"in": map[string]any{
			"partner_code": "PRTD-" + tag,
			"kind":         "INDIVIDUAL",
			"display_name": "Disposable Partner " + tag,
		},
	}), "createParty", "party")
	partyTempID := requireNonEmptyString(t, partyTemp, "id")
	partyTempRawID := requireNonEmptyString(t, partyTemp, "_id")

	t.Cleanup(func() {
		cleanupEFEntityRows(t, cfg, "parties", partyTempRawID)
		cleanupEFEntityRows(t, cfg, "parties", partyRawID)
	})

	_ = requireMutationChildOK(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($in: UpdatePartyProfileInput!) {
		updatePartyProfile(input: $in) { party { id display_name } %s }
	}`, gqlMutationResultFields), map[string]any{
		"in": map[string]any{
			"id":                 partyID,
			"kind":               "ORGANIZATION",
			"display_name":       "CV Tekstil Nusantara Prima " + tag,
			"legal_name":         "CV Tekstil Nusantara Jaya " + tag,
			"email":              "vendor-" + tag + "@partner.test",
			"preferred_language": "id",
		},
	}), "updatePartyProfile", "party")

	_ = requireMutationChildOK(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($in: UpdatePartyTaxProfileInput!) {
		updatePartyTaxProfile(input: $in) { party { id is_pkp } %s }
	}`, gqlMutationResultFields), map[string]any{
		"in": map[string]any{
			"id":                  partyID,
			"tax_id_number":       "88776655" + tag[len(tag)-4:],
			"tax_registered_name": "CV Tekstil Nusantara Jaya " + tag,
			"is_pkp":              true,
			"is_tax_exempt":       false,
			"tax_verified_at":     time.Now().UTC().Format(time.RFC3339),
		},
	}), "updatePartyTaxProfile", "party")

	_ = requireMutationChildOK(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($id: ID!) {
		verifyAndActivateParty(id: $id) { party { id status } %s }
	}`, gqlMutationResultFields), map[string]any{"id": partyID}), "verifyAndActivateParty", "party")

	_ = requireMutationChildOK(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($id: ID!) {
		suspendParty(id: $id) { party { id status } %s }
	}`, gqlMutationResultFields), map[string]any{"id": partyTempID}), "suspendParty", "party")

	_ = requireMutationChildOK(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($id: ID!) {
		deactivateParty(id: $id) { party { id status } %s }
	}`, gqlMutationResultFields), map[string]any{"id": partyTempID}), "deactivateParty", "party")

	delPartyPayload := requireMutationSuccess(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($id: ID!) {
		deleteParty(id: $id) { deleted %s }
	}`, gqlMutationResultFields), map[string]any{"id": partyTempID}), "deleteParty")
	if deleted, _ := delPartyPayload["deleted"].(bool); !deleted {
		t.Fatalf("deleteParty expected deleted=true, got %+v", delPartyPayload)
	}

	// PartyRole: SUPPLIER + VENDOR + CARRIER.
	prSupplier := requireMutationChildOK(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($in: AssignPartyRoleInput!) {
		assignPartyRole(input: $in) { role { id _id role is_active } %s }
	}`, gqlMutationResultFields), map[string]any{
		"in": map[string]any{"party_id": partyID, "role": "SUPPLIER", "reference_note": "Bahan Kain " + tag},
	}), "assignPartyRole", "role")
	prSupplierID := requireNonEmptyString(t, prSupplier, "id")

	_ = requireMutationChildOK(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($in: AssignPartyRoleInput!) {
		assignPartyRole(input: $in) { role { id _id role is_active } %s }
	}`, gqlMutationResultFields), map[string]any{
		"in": map[string]any{"party_id": partyID, "role": "VENDOR", "reference_note": "CMT Jahit " + tag},
	}), "assignPartyRole", "role")

	prCarrier := requireMutationChildOK(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($in: AssignPartyRoleInput!) {
		assignPartyRole(input: $in) { role { id _id role is_active } %s }
	}`, gqlMutationResultFields), map[string]any{
		"in": map[string]any{"party_id": partyID, "role": "CARRIER", "reference_note": "Temp carrier " + tag},
	}), "assignPartyRole", "role")
	prCarrierID := requireNonEmptyString(t, prCarrier, "id")

	_ = requireMutationChildOK(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($in: UpdatePartyRoleNoteInput!) {
		updatePartyRoleNote(input: $in) { role { id reference_note } %s }
	}`, gqlMutationResultFields), map[string]any{
		"in": map[string]any{"id": prSupplierID, "reference_note": "Updated note " + tag},
	}), "updatePartyRoleNote", "role")

	_ = requireMutationChildOK(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($id: ID!) {
		deactivatePartyRole(role_id: $id) { role { id is_active } %s }
	}`, gqlMutationResultFields), map[string]any{"id": prSupplierID}), "deactivatePartyRole", "role")

	_ = requireMutationChildOK(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($id: ID!) {
		activatePartyRole(role_id: $id) { role { id is_active } %s }
	}`, gqlMutationResultFields), map[string]any{"id": prSupplierID}), "activatePartyRole", "role")

	remRolePayload := requireMutationSuccess(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($id: ID!) {
		removePartyRole(role_id: $id) { removed %s }
	}`, gqlMutationResultFields), map[string]any{"id": prCarrierID}), "removePartyRole")
	if removed, _ := remRolePayload["removed"].(bool); !removed {
		t.Fatalf("removePartyRole expected removed=true, got %+v", remRolePayload)
	}

	// PartyContact & Portal Access lifecycle.
	resetGatewayRateLimits(t, cfg)
	picEmail := "pic-" + tag + "@partner.test"
	contactPrimary := requireMutationChildOK(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($in: AddPartyContactInput!) {
		addPartyContact(input: $in) { contact { id _id full_name portal_access_status } %s }
	}`, gqlMutationResultFields), map[string]any{
		"in": map[string]any{
			"party_id":     partyID,
			"full_name":    "Siti Rahmawati " + tag,
			"job_title":    "Account Lead",
			"contact_role": "ACCOUNT_MANAGER",
			"email":        picEmail,
			"phone":        "+62811999888",
			"is_primary":   false,
		},
	}), "addPartyContact", "contact")
	contact1ID := requireNonEmptyString(t, contactPrimary, "id")

	contactDisposable := requireMutationChildOK(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($in: AddPartyContactInput!) {
		addPartyContact(input: $in) { contact { id _id full_name } %s }
	}`, gqlMutationResultFields), map[string]any{
		"in": map[string]any{
			"party_id":     partyID,
			"full_name":    "Temp PIC " + tag,
			"contact_role": "BACKUP",
			"email":        "temp-" + tag + "@partner.test",
		},
	}), "addPartyContact", "contact")
	contact2ID := requireNonEmptyString(t, contactDisposable, "id")

	_ = requireMutationChildOK(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($in: UpdatePartyContactInput!) {
		updatePartyContact(input: $in) { contact { id full_name } %s }
	}`, gqlMutationResultFields), map[string]any{
		"in": map[string]any{
			"id":           contact1ID,
			"full_name":    "Siti Rahmawati Updated " + tag,
			"job_title":    "Director",
			"contact_role": "OWNER",
			"email":        picEmail,
			"phone":        "+62811999888",
			"is_primary":   true,
		},
	}), "updatePartyContact", "contact")

	_ = requireMutationChildOK(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($p: ID!, $c: ID!) {
		setPrimaryPartyContact(party_id: $p, contact_id: $c) { contact { id is_primary } %s }
	}`, gqlMutationResultFields), map[string]any{"p": partyID, "c": contact1ID}), "setPrimaryPartyContact", "contact")

	_ = requireMutationChildOK(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($c: ID!, $u: ID!) {
		requestPartyPortalLink(contact_id: $c, iam_user_id: $u) { contact { id portal_access_status } %s }
	}`, gqlMutationResultFields), map[string]any{"c": contact1ID, "u": superadminUserID}), "requestPartyPortalLink", "contact")

	_ = requireMutationChildOK(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($c: ID!, $u: ID!) {
		activatePartyPortalAccess(contact_id: $c, iam_user_id: $u) { contact { id portal_access_status } %s }
	}`, gqlMutationResultFields), map[string]any{"c": contact1ID, "u": superadminUserID}), "activatePartyPortalAccess", "contact")

	// Evaluate + query while ACTIVE, then revoke & unlink + deactivate/activate/delete disposable contact.
	_ = doGQLOK(t, admin, `query PartyContactAndPortalQueries(
		$pid: ID!, $pids: [ID!]!, $pcode: String!, $npwp: String!, $pemail: String!, $pq: String,
		$rid: ID!, $cid: ID!, $uid: ID!, $cemail: String!
	) {
		party(id: $pid) { id partner_code roles(first: 5) { total_count } contacts(first: 5) { total_count } addresses(first: 5) { total_count } }
		partiesByIds(ids: $pids) { id partner_code }
		partyByCode(partner_code: $pcode) { id partner_code }
		partyByTaxIdNumber(tax_id_number: $npwp) { id tax_id_number }
		partiesByEmail(email: $pemail, first: 5) { total_count nodes { id } }
		parties(query: $pq, status: ACTIVE, role: SUPPLIER, first: 5) { total_count nodes { id } }
		allParties(first: 5) { total_count nodes { id } }

		partyRole(id: $rid) { id role is_active }
		partyRolesByParty(party_id: $pid, first: 5) { total_count nodes { id } }
		partyRoleByPartyAndKind(party_id: $pid, kind: SUPPLIER) { id role }
		partyRoles(party_id: $pid, active_only: true, first: 5) { total_count nodes { id } }
		allPartyRoles(first: 5) { total_count nodes { id } }
		partyRolesByKind(kind: VENDOR, active_only: true, first: 5) { total_count nodes { id } }
		allPartyRolesByKind(kind: SUPPLIER, active_only: true, first: 5) { total_count nodes { id } }
		validatePartyRoleScope(party_id: $pid, required_role: SUPPLIER) { eligible scope_kind target_id code }

		partyContact(id: $cid) { id full_name portal_access_status }
		partyContactsByParty(party_id: $pid, first: 5) { total_count nodes { id } }
		primaryPartyContact(party_id: $pid) { id is_primary }
		partyContactsByIamUserId(iam_user_id: $uid, first: 5) { total_count nodes { id } }
		partyContactsByEmail(email: $cemail, first: 5) { total_count nodes { id } }
		partyContacts(party_id: $pid, portal_access_status: ACTIVE, first: 5) { total_count nodes { id } }
		allPartyContacts(first: 5) { total_count nodes { id } }
		evaluateRepresentativePortalAccess(party_id: $pid, iam_user_id: $uid, required_role: SUPPLIER) { allowed party_id contact_id iam_user_id active_roles }
		node(id: $pid) { id ... on Party { partner_code display_name } }
	}`, map[string]any{
		"pid":    partyID,
		"pids":   []string{partyID},
		"pcode":  "PRT-" + tag,
		"npwp":   "88776655" + tag[len(tag)-4:],
		"pemail": "vendor-" + tag + "@partner.test",
		"pq":     tag,
		"rid":    prSupplierID,
		"cid":    contact1ID,
		"uid":    superadminUserID,
		"cemail": picEmail,
	})

	_ = requireMutationChildOK(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($c: ID!) {
		revokePartyPortalAccess(contact_id: $c) { contact { id portal_access_status } %s }
	}`, gqlMutationResultFields), map[string]any{"c": contact1ID}), "revokePartyPortalAccess", "contact")

	_ = requireMutationChildOK(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($c: ID!) {
		unlinkPartyPortalUser(contact_id: $c) { contact { id iam_user_id portal_access_status } %s }
	}`, gqlMutationResultFields), map[string]any{"c": contact1ID}), "unlinkPartyPortalUser", "contact")

	_ = requireMutationChildOK(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($c: ID!) {
		deactivatePartyContact(contact_id: $c) { contact { id is_active } %s }
	}`, gqlMutationResultFields), map[string]any{"c": contact2ID}), "deactivatePartyContact", "contact")

	_ = requireMutationChildOK(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($c: ID!) {
		activatePartyContact(contact_id: $c) { contact { id is_active } %s }
	}`, gqlMutationResultFields), map[string]any{"c": contact2ID}), "activatePartyContact", "contact")

	delContactPayload := requireMutationSuccess(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($c: ID!) {
		deletePartyContact(contact_id: $c) { deleted %s }
	}`, gqlMutationResultFields), map[string]any{"c": contact2ID}), "deletePartyContact")
	if deleted, _ := delContactPayload["deleted"].(bool); !deleted {
		t.Fatalf("deletePartyContact expected deleted=true, got %+v", delContactPayload)
	}

	// PartyAddressLink: 7 Queries + 4 Mutations.
	paLink1 := requireMutationChildOK(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($in: LinkPartyAddressInput!) {
		linkPartyAddress(input: $in) { link { id _id usage is_primary } %s }
	}`, gqlMutationResultFields), map[string]any{
		"in": map[string]any{"party_id": partyID, "address_id": addr1ID, "usage": "BILLING", "is_primary": false, "replace_primary": true},
	}), "linkPartyAddress", "link")
	pa1ID := requireNonEmptyString(t, paLink1, "id")

	_ = requireMutationChildOK(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($in: UpdatePartyAddressLinkInput!) {
		updatePartyAddressLink(input: $in) { link { id usage is_primary } %s }
	}`, gqlMutationResultFields), map[string]any{
		"in": map[string]any{"id": pa1ID, "usage": "REGISTERED_OFFICE", "is_primary": false, "replace_primary": true},
	}), "updatePartyAddressLink", "link")

	_ = requireMutationChildOK(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($id: ID!) {
		setPrimaryPartyAddressLink(link_id: $id) { link { id is_primary } %s }
	}`, gqlMutationResultFields), map[string]any{"id": pa1ID}), "setPrimaryPartyAddressLink", "link")

	_ = doGQLOK(t, admin, `query PartyAddressQueries($lid: ID!, $pid: ID!, $aid: ID!) {
		partyAddressLink(id: $lid) { id usage is_primary address { id city } }
		partyAddressLinksByParty(party_id: $pid, first: 5) { total_count nodes { id } }
		partyAddressLinksByPartyAndUsage(party_id: $pid, usage: REGISTERED_OFFICE, first: 5) { total_count nodes { id } }
		primaryPartyAddressLink(party_id: $pid, usage: REGISTERED_OFFICE) { id is_primary }
		partyAddressLinksByAddressId(address_id: $aid, first: 5) { total_count nodes { id } }
		partyAddressLinks(party_id: $pid, primary_only: true, first: 5) { total_count nodes { id } }
		allPartyAddressLinks(first: 5) { total_count nodes { id } }
		node(id: $lid) { id ... on PartyAddress { usage is_primary } }
	}`, map[string]any{"lid": pa1ID, "pid": partyID, "aid": addr1ID})

	unlinkPayload := requireMutationSuccess(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($id: ID!) {
		unlinkPartyAddress(link_id: $id) { unlinked %s }
	}`, gqlMutationResultFields), map[string]any{"id": pa1ID}), "unlinkPartyAddress")
	if unlinked, _ := unlinkPayload["unlinked"].(bool); !unlinked {
		t.Fatalf("unlinkPartyAddress expected unlinked=true, got %+v", unlinkPayload)
	}

	// =========================================================================
	// 7. Penugasan Seluruh 5 Scope Eksternal EF Aktif ke Superadmin via IAM <-> EF gRPC
	// =========================================================================
	resetGatewayRateLimits(t, cfg)
	efScopeAssignments := []struct {
		scopeType string
		scopeID   string
		roleCode  string
	}{
		{"LEGAL_ENTITY", leParentRawID, "ef_admin"},
		{"ORGANIZATION_UNIT", ouParentRawID, "ef_structure_manager"},
		{"FACILITY", locParentRawID, "ef_operator"},
		{"SUPPLIER_PARTY", partyRawID, "ef_partner_manager"},
		{"VENDOR_PARTY", partyRawID, "ef_readonly"},
	}
	for _, sc := range efScopeAssignments {
		roleID := requireNonEmptyString(t, efRoleByCode[sc.roleCode], "id")
		edge := requireMutationChildOK(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($in: AssignUserRoleInput!) {
			assignUserRole(input: $in) { user_role_edge { assignment_id scope { type id } node { code } } %s }
		}`, gqlMutationResultFields), map[string]any{
			"in": map[string]any{
				"user_id":    superadminUserID,
				"role_id":    roleID,
				"scope_type": sc.scopeType,
				"scope_id":   sc.scopeID,
			},
		}), "assignUserRole", "user_role_edge")
		assignID := requireNonEmptyString(t, edge, "assignment_id")
		cleanupAssignments = append(cleanupAssignments, assignID)
		gotScope := requireChildMap(t, edge, "scope")
		if gotScope["type"] != sc.scopeType || !strings.EqualFold(fmt.Sprint(gotScope["id"]), sc.scopeID) {
			t.Fatalf("assigned scope mismatch = %+v, want %s/%s", gotScope, sc.scopeType, sc.scopeID)
		}
	}
}
