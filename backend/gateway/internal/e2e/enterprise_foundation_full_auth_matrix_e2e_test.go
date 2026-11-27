//go:build e2e

package e2e

import (
	"fmt"
	"strings"
	"testing"
)

// TestEnterpriseFoundation_FullAuthMatrix_All64Queries48MutationsAnd8Nodes menguji matriks penuh
// authorized / unauthorized (CUSTOMER 0-izin, ef_readonly, dan cross-scope) / anonymous (UNAUTHENTICATED / AUTHENTICATION)
// pada seluruh 64 Query, 48 Mutation, dan 8 tipe Relay Node Enterprise Foundation.
func TestEnterpriseFoundation_FullAuthMatrix_All64Queries48MutationsAnd8Nodes(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	seedEFPermissionCatalog(t, cfg)
	resetGatewayRateLimits(t, cfg)
	t.Cleanup(func() { resetGatewayRateLimits(t, cfg) })

	admin := newSuperAdminBrowser(t, cfg)
	anon := newAnonBrowser(t, cfg)
	_, custClient := createFixtureCustomerAndLogin(t, admin, cfg)

	roleIDByCode := func(code string) string {
		r := requireOpMap(t, doGQLOK(t, admin, `query ($c: String!) { roleByCode(code: $c) { id } }`, map[string]any{"c": code}), "roleByCode")
		return requireNonEmptyString(t, r, "id")
	}
	assignGlobalRoleAndLogin := func(prefix, roleCode string) (fixtureUserAccount, *browserClient) {
		acct := createFixtureUser(t, admin, cfg, "EMPLOYEE", true)
		_ = requireMutationChildOK(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($in: AssignUserRoleInput!) {
			assignUserRole(input: $in) { %s }
		}`, gqlUserRolePayloadFields), map[string]any{
			"in": map[string]any{
				"user_id":    acct.ID,
				"role_id":    roleIDByCode(roleCode),
				"scope_type": "GLOBAL",
			},
		}), "assignUserRole", "user_role_edge")
		resetGatewayRateLimits(t, cfg)
		cli, payload := loginFixtureUser(t, cfg, acct.Email, acct.Password)
		assertUserLoginPayload(t, payload, "AUTHENTICATED")
		_ = prefix
		return acct, cli
	}

	efAdminAcct, efAdminClient := assignGlobalRoleAndLogin("efadmin", "ef_admin")
	_, efReadonlyClient := assignGlobalRoleAndLogin("efro", "ef_readonly")

	tag := uniqueTag("efmat")
	tagUpper := strings.ToUpper(strings.ReplaceAll(tag, "_", "-"))

	// =========================================================================
	// 1. Siapkan 1 himpunan entitas EF lengkap yang valid (Address, 2 LegalEntity,
	//    Location, OrganizationUnit, Party, PartyRole, PartyContact, PartyAddress).
	// =========================================================================
	addrObj := requireMutationChildOK(t, doGQLOK(t, efAdminClient, fmt.Sprintf(`mutation ($in: CreateAddressInput!) {
		createAddress(input: $in) { address { %s } %s }
	}`, gqlEFAddressFields, gqlMutationResultFields), map[string]any{
		"in": map[string]any{
			"label":          "Matrix Addr " + tag,
			"address_line_1": "Jl. Matrix Utama No. 1 " + tag,
			"city":           "Jakarta Selatan",
			"country_code":   "ID",
		},
	}), "createAddress", "address")
	addrID := requireNonEmptyString(t, addrObj, "id")
	t.Cleanup(func() { cleanupEFEntityRows(t, cfg, "addresses", requireNonEmptyString(t, addrObj, "_id")) })

	leACode := "LE-A-" + tagUpper
	leATax := "91." + tagUpper
	leANIB := "NIB-A-" + tagUpper
	leAObj := requireMutationChildOK(t, doGQLOK(t, efAdminClient, fmt.Sprintf(`mutation ($in: CreateLegalEntityInput!) {
		createLegalEntity(input: $in) { legal_entity { %s } %s }
	}`, gqlEFLegalEntityFields, gqlMutationResultFields), map[string]any{
		"in": map[string]any{
			"entity_code":               leACode,
			"legal_name":                "PT Matrix Entitas A " + tag,
			"jurisdiction_country_code": "ID",
			"nib":                       leANIB,
			"tax_id_number":             leATax,
			"is_pkp":                    true,
			"registered_address_id":     addrID,
		},
	}), "createLegalEntity", "legal_entity")
	leAID := requireNonEmptyString(t, leAObj, "id")
	leARawID := requireNonEmptyString(t, leAObj, "_id")
	t.Cleanup(func() { cleanupEFEntityRows(t, cfg, "legal_entities", leARawID) })
	_ = requireMutationChildOK(t, doGQLOK(t, efAdminClient, fmt.Sprintf(`mutation ($id: ID!) { activateLegalEntity(id: $id) { legal_entity { id status } %s } }`, gqlMutationResultFields), map[string]any{"id": leAID}), "activateLegalEntity", "legal_entity")

	leBCode := "LE-B-" + tagUpper
	leBObj := requireMutationChildOK(t, doGQLOK(t, efAdminClient, fmt.Sprintf(`mutation ($in: CreateLegalEntityInput!) {
		createLegalEntity(input: $in) { legal_entity { id _id } %s }
	}`, gqlMutationResultFields), map[string]any{
		"in": map[string]any{
			"parent_legal_entity_id":    leAID,
			"entity_code":               leBCode,
			"legal_name":                "PT Matrix Entitas B " + tag,
			"jurisdiction_country_code": "ID",
		},
	}), "createLegalEntity", "legal_entity")
	leBID := requireNonEmptyString(t, leBObj, "id")
	t.Cleanup(func() { cleanupEFEntityRows(t, cfg, "legal_entities", requireNonEmptyString(t, leBObj, "_id")) })
	_ = requireMutationChildOK(t, doGQLOK(t, efAdminClient, fmt.Sprintf(`mutation ($id: ID!) { activateLegalEntity(id: $id) { legal_entity { id } %s } }`, gqlMutationResultFields), map[string]any{"id": leBID}), "activateLegalEntity", "legal_entity")

	locCode := "LOC-" + tagUpper
	locObj := requireMutationChildOK(t, doGQLOK(t, efAdminClient, fmt.Sprintf(`mutation ($in: CreateLocationInput!) {
		createLocation(input: $in) { location { %s } %s }
	}`, gqlEFLocationFields, gqlMutationResultFields), map[string]any{
		"in": map[string]any{
			"legal_entity_id": leAID,
			"address_id":      addrID,
			"location_code":   locCode,
			"name":            "Gudang Matrix " + tag,
			"kind":            "WAREHOUSE",
			"timezone":        "Asia/Jakarta",
		},
	}), "createLocation", "location")
	locID := requireNonEmptyString(t, locObj, "id")
	t.Cleanup(func() { cleanupEFEntityRows(t, cfg, "locations", requireNonEmptyString(t, locObj, "_id")) })

	ouCode := "OU-" + tagUpper
	ouObj := requireMutationChildOK(t, doGQLOK(t, efAdminClient, fmt.Sprintf(`mutation ($in: CreateOrganizationUnitInput!) {
		createOrganizationUnit(input: $in) { organization_unit { %s } %s }
	}`, gqlEFOrgUnitFields, gqlMutationResultFields), map[string]any{
		"in": map[string]any{
			"legal_entity_id":     leAID,
			"primary_location_id": locID,
			"unit_code":           ouCode,
			"name":                "Divisi Matrix " + tag,
			"kind":                "DIVISION",
			"head_person_id":      efAdminAcct.ID,
		},
	}), "createOrganizationUnit", "organization_unit")
	ouID := requireNonEmptyString(t, ouObj, "id")
	t.Cleanup(func() { cleanupEFEntityRows(t, cfg, "organization_units", requireNonEmptyString(t, ouObj, "_id")) })

	pCode := "PRT-" + tagUpper
	pTax := "82." + tagUpper
	pEmail := "matrix_" + tag + "@vendor.test"
	pObj := requireMutationChildOK(t, doGQLOK(t, efAdminClient, fmt.Sprintf(`mutation ($in: CreatePartyInput!) {
		createParty(input: $in) { party { %s } %s }
	}`, gqlEFPartyFields, gqlMutationResultFields), map[string]any{
		"in": map[string]any{
			"partner_code":       pCode,
			"kind":               "ORGANIZATION",
			"display_name":       "PT Matrix Mitra " + tag,
			"email":              pEmail,
			"tax_id_number":      pTax,
			"preferred_language": "id",
		},
	}), "createParty", "party")
	partyID := requireNonEmptyString(t, pObj, "id")
	t.Cleanup(func() { cleanupEFEntityRows(t, cfg, "parties", requireNonEmptyString(t, pObj, "_id")) })
	_ = requireMutationChildOK(t, doGQLOK(t, efAdminClient, fmt.Sprintf(`mutation ($id: ID!) { verifyAndActivateParty(id: $id) { party { id } %s } }`, gqlMutationResultFields), map[string]any{"id": partyID}), "verifyAndActivateParty", "party")

	prObj := requireMutationChildOK(t, doGQLOK(t, efAdminClient, fmt.Sprintf(`mutation ($in: AssignPartyRoleInput!) {
		assignPartyRole(input: $in) { role { %s } %s }
	}`, gqlEFPartyRoleFields, gqlMutationResultFields), map[string]any{
		"in": map[string]any{"party_id": partyID, "role": "SUPPLIER", "reference_note": "Matrix supplier"},
	}), "assignPartyRole", "role")
	partyRoleID := requireNonEmptyString(t, prObj, "id")
	t.Cleanup(func() { cleanupEFEntityRows(t, cfg, "party_roles", requireNonEmptyString(t, prObj, "_id")) })

	pcEmail := "pic_mat_" + tag + "@vendor.test"
	pcObj := requireMutationChildOK(t, doGQLOK(t, efAdminClient, fmt.Sprintf(`mutation ($in: AddPartyContactInput!) {
		addPartyContact(input: $in) { contact { %s } %s }
	}`, gqlEFPartyContactFields, gqlMutationResultFields), map[string]any{
		"in": map[string]any{
			"party_id":     partyID,
			"full_name":    "PIC Matrix " + tag,
			"contact_role": "account_manager",
			"email":        pcEmail,
			"is_primary":   true,
		},
	}), "addPartyContact", "contact")
	contactID := requireNonEmptyString(t, pcObj, "id")
	t.Cleanup(func() { cleanupEFEntityRows(t, cfg, "party_contacts", requireNonEmptyString(t, pcObj, "_id")) })
	_ = requireMutationChildOK(t, doGQLOK(t, efAdminClient, fmt.Sprintf(`mutation ($c: ID!, $u: ID!) { activatePartyPortalAccess(contact_id: $c, iam_user_id: $u) { contact { id } %s } }`, gqlMutationResultFields), map[string]any{"c": contactID, "u": efAdminAcct.ID}), "activatePartyPortalAccess", "contact")

	paObj := requireMutationChildOK(t, doGQLOK(t, efAdminClient, fmt.Sprintf(`mutation ($in: LinkPartyAddressInput!) {
		linkPartyAddress(input: $in) { link { %s } %s }
	}`, gqlEFPartyAddressFields, gqlMutationResultFields), map[string]any{
		"in": map[string]any{"party_id": partyID, "address_id": addrID, "usage": "REGISTERED_OFFICE", "is_primary": true},
	}), "linkPartyAddress", "link")
	paID := requireNonEmptyString(t, paObj, "id")
	t.Cleanup(func() { cleanupEFEntityRows(t, cfg, "party_addresses", requireNonEmptyString(t, paObj, "_id")) })

	// Pastikan ada minimal satu primaryOperatingLegalEntity agar Query.primaryOperatingLegalEntity mengembalikan objek non-null saat authorized.
	primCheck := requireOpMap(t, doGQLOK(t, efAdminClient, `query { legalEntities(primary_only: true, first: 1) { total_count } }`, nil), "legalEntities")
	if requireNumberField(t, primCheck, "total_count") == 0 {
		_ = requireMutationChildOK(t, doGQLOK(t, efAdminClient, fmt.Sprintf(`mutation ($id: ID!) {
			setPrimaryOperatingLegalEntity(id: $id) { legal_entity { id is_primary_operating_entity } %s }
		}`, gqlMutationResultFields), map[string]any{"id": leAID}), "setPrimaryOperatingLegalEntity", "legal_entity")
	}

	// =========================================================================
	// 2. Matriks Penuh 64 Query EF + 8 Relay Node (authorized, unauthorized customer, anonymous)
	// =========================================================================
	efQueries := []gqlOpSpec{
		// Address (5):
		{name: "address", query: `query ($id: ID!) { address(id: $id) { id } }`, vars: map[string]any{"id": addrID}},
		{name: "addressesByIds", query: `query ($ids: [ID!]!) { addressesByIds(ids: $ids) { id } }`, vars: map[string]any{"ids": []string{addrID}}},
		{name: "addresses", query: `query { addresses(first: 2) { total_count } }`},
		{name: "allAddresses", query: `query { allAddresses(first: 2) { total_count } }`},
		{name: "isAddressReferenced", query: `query ($id: ID!) { isAddressReferenced(id: $id) }`, vars: map[string]any{"id": addrID}},
		// LegalEntity (10):
		{name: "legalEntity", query: `query ($id: ID!) { legalEntity(id: $id) { id } }`, vars: map[string]any{"id": leAID}},
		{name: "legalEntitiesByIds", query: `query ($ids: [ID!]!) { legalEntitiesByIds(ids: $ids) { id } }`, vars: map[string]any{"ids": []string{leAID}}},
		{name: "legalEntityByCode", query: `query ($c: String!) { legalEntityByCode(entity_code: $c) { id } }`, vars: map[string]any{"c": leACode}},
		{name: "legalEntityByTaxIdNumber", query: `query ($t: String!) { legalEntityByTaxIdNumber(tax_id_number: $t) { id } }`, vars: map[string]any{"t": leATax}},
		{name: "legalEntitiesByNib", query: `query ($n: String!) { legalEntitiesByNib(nib: $n, first: 2) { total_count } }`, vars: map[string]any{"n": leANIB}},
		{name: "primaryOperatingLegalEntity", query: `query { primaryOperatingLegalEntity { id } }`},
		{name: "legalEntities", query: `query { legalEntities(first: 2) { total_count } }`},
		{name: "allLegalEntities", query: `query { allLegalEntities(first: 2) { total_count } }`},
		{name: "legalEntityChildren", query: `query ($p: ID!) { legalEntityChildren(parent_legal_entity_id: $p, first: 2) { total_count } }`, vars: map[string]any{"p": leAID}},
		{name: "validateLegalEntityScopeTarget", query: `query ($id: ID!) { validateLegalEntityScopeTarget(id: $id) { eligible } }`, vars: map[string]any{"id": leAID}},
		// Location (10):
		{name: "location", query: `query ($id: ID!) { location(id: $id) { id } }`, vars: map[string]any{"id": locID}},
		{name: "locationsByIds", query: `query ($ids: [ID!]!) { locationsByIds(ids: $ids) { id } }`, vars: map[string]any{"ids": []string{locID}}},
		{name: "locationByCode", query: `query ($c: String!) { locationByCode(location_code: $c) { id } }`, vars: map[string]any{"c": locCode}},
		{name: "locationsByLegalEntityAndKind", query: `query ($le: ID!) { locationsByLegalEntityAndKind(legal_entity_id: $le, kind: WAREHOUSE, first: 2) { total_count } }`, vars: map[string]any{"le": leAID}},
		{name: "locationsByAddressId", query: `query ($a: ID!) { locationsByAddressId(address_id: $a, first: 2) { total_count } }`, vars: map[string]any{"a": addrID}},
		{name: "locationChildren", query: `query ($p: ID!) { locationChildren(parent_location_id: $p, first: 2) { total_count } }`, vars: map[string]any{"p": locID}},
		{name: "locations", query: `query { locations(first: 2) { total_count } }`},
		{name: "allLocations", query: `query { allLocations(first: 2) { total_count } }`},
		{name: "validateFacilityScope", query: `query ($id: ID!) { validateFacilityScope(id: $id) { eligible } }`, vars: map[string]any{"id": locID}},
		{name: "validateWarehouseScope", query: `query ($id: ID!) { validateWarehouseScope(id: $id) { eligible } }`, vars: map[string]any{"id": locID}},
		// OrganizationUnit (9):
		{name: "organizationUnit", query: `query ($id: ID!) { organizationUnit(id: $id) { id } }`, vars: map[string]any{"id": ouID}},
		{name: "organizationUnitsByIds", query: `query ($ids: [ID!]!) { organizationUnitsByIds(ids: $ids) { id } }`, vars: map[string]any{"ids": []string{ouID}}},
		{name: "organizationUnitByEntityAndCode", query: `query ($le: ID!, $c: String!) { organizationUnitByEntityAndCode(legal_entity_id: $le, unit_code: $c) { id } }`, vars: map[string]any{"le": leAID, "c": ouCode}},
		{name: "organizationUnitsByHeadPersonId", query: `query ($h: ID!) { organizationUnitsByHeadPersonId(head_person_id: $h, first: 2) { total_count } }`, vars: map[string]any{"h": efAdminAcct.ID}},
		{name: "organizationUnitsByPrimaryLocationId", query: `query ($l: ID!) { organizationUnitsByPrimaryLocationId(primary_location_id: $l, first: 2) { total_count } }`, vars: map[string]any{"l": locID}},
		{name: "organizationUnitChildren", query: `query ($p: ID!) { organizationUnitChildren(parent_unit_id: $p, first: 2) { total_count } }`, vars: map[string]any{"p": ouID}},
		{name: "organizationUnits", query: `query { organizationUnits(first: 2) { total_count } }`},
		{name: "allOrganizationUnits", query: `query { allOrganizationUnits(first: 2) { total_count } }`},
		{name: "validateOrganizationUnitScopeTarget", query: `query ($id: ID!) { validateOrganizationUnitScopeTarget(id: $id) { eligible } }`, vars: map[string]any{"id": ouID}},
		// Party Core (7):
		{name: "party", query: `query ($id: ID!) { party(id: $id) { id } }`, vars: map[string]any{"id": partyID}},
		{name: "partiesByIds", query: `query ($ids: [ID!]!) { partiesByIds(ids: $ids) { id } }`, vars: map[string]any{"ids": []string{partyID}}},
		{name: "partyByCode", query: `query ($c: String!) { partyByCode(partner_code: $c) { id } }`, vars: map[string]any{"c": pCode}},
		{name: "partyByTaxIdNumber", query: `query ($t: String!) { partyByTaxIdNumber(tax_id_number: $t) { id } }`, vars: map[string]any{"t": pTax}},
		{name: "partiesByEmail", query: `query ($e: String!) { partiesByEmail(email: $e, first: 2) { total_count } }`, vars: map[string]any{"e": pEmail}},
		{name: "parties", query: `query { parties(first: 2) { total_count } }`},
		{name: "allParties", query: `query { allParties(first: 2) { total_count } }`},
		// PartyRole (8):
		{name: "partyRole", query: `query ($id: ID!) { partyRole(id: $id) { id } }`, vars: map[string]any{"id": partyRoleID}},
		{name: "partyRolesByParty", query: `query ($p: ID!) { partyRolesByParty(party_id: $p, first: 2) { total_count } }`, vars: map[string]any{"p": partyID}},
		{name: "partyRoleByPartyAndKind", query: `query ($p: ID!) { partyRoleByPartyAndKind(party_id: $p, kind: SUPPLIER) { id } }`, vars: map[string]any{"p": partyID}},
		{name: "partyRoles", query: `query { partyRoles(first: 2) { total_count } }`},
		{name: "allPartyRoles", query: `query { allPartyRoles(first: 2) { total_count } }`},
		{name: "partyRolesByKind", query: `query { partyRolesByKind(kind: SUPPLIER, first: 2) { total_count } }`},
		{name: "allPartyRolesByKind", query: `query { allPartyRolesByKind(kind: SUPPLIER, first: 2) { total_count } }`},
		{name: "validatePartyRoleScope", query: `query ($p: ID!) { validatePartyRoleScope(party_id: $p, required_role: SUPPLIER) { eligible } }`, vars: map[string]any{"p": partyID}},
		// PartyContact & Portal Evaluation (8):
		{name: "partyContact", query: `query ($id: ID!) { partyContact(id: $id) { id } }`, vars: map[string]any{"id": contactID}},
		{name: "partyContactsByParty", query: `query ($p: ID!) { partyContactsByParty(party_id: $p, first: 2) { total_count } }`, vars: map[string]any{"p": partyID}},
		{name: "primaryPartyContact", query: `query ($p: ID!) { primaryPartyContact(party_id: $p) { id } }`, vars: map[string]any{"p": partyID}},
		{name: "partyContactsByIamUserId", query: `query ($u: ID!) { partyContactsByIamUserId(iam_user_id: $u, first: 2) { total_count } }`, vars: map[string]any{"u": efAdminAcct.ID}},
		{name: "partyContactsByEmail", query: `query ($e: String!) { partyContactsByEmail(email: $e, first: 2) { total_count } }`, vars: map[string]any{"e": pcEmail}},
		{name: "partyContacts", query: `query { partyContacts(first: 2) { total_count } }`},
		{name: "allPartyContacts", query: `query { allPartyContacts(first: 2) { total_count } }`},
		{name: "evaluateRepresentativePortalAccess", query: `query ($p: ID!, $u: ID!) { evaluateRepresentativePortalAccess(party_id: $p, iam_user_id: $u, required_role: SUPPLIER) { allowed } }`, vars: map[string]any{"p": partyID, "u": efAdminAcct.ID}},
		// PartyAddressLink (7):
		{name: "partyAddressLink", query: `query ($id: ID!) { partyAddressLink(id: $id) { id } }`, vars: map[string]any{"id": paID}},
		{name: "partyAddressLinksByParty", query: `query ($p: ID!) { partyAddressLinksByParty(party_id: $p, first: 2) { total_count } }`, vars: map[string]any{"p": partyID}},
		{name: "partyAddressLinksByPartyAndUsage", query: `query ($p: ID!) { partyAddressLinksByPartyAndUsage(party_id: $p, usage: REGISTERED_OFFICE, first: 2) { total_count } }`, vars: map[string]any{"p": partyID}},
		{name: "primaryPartyAddressLink", query: `query ($p: ID!) { primaryPartyAddressLink(party_id: $p, usage: REGISTERED_OFFICE) { id } }`, vars: map[string]any{"p": partyID}},
		{name: "partyAddressLinksByAddressId", query: `query ($a: ID!) { partyAddressLinksByAddressId(address_id: $a, first: 2) { total_count } }`, vars: map[string]any{"a": addrID}},
		{name: "partyAddressLinks", query: `query { partyAddressLinks(first: 2) { total_count } }`},
		{name: "allPartyAddressLinks", query: `query { allPartyAddressLinks(first: 2) { total_count } }`},
	}
	if len(efQueries) != 64 {
		t.Fatalf("expected 64 EF queries in matrix, got %d", len(efQueries))
	}

	efNodeQueries := []gqlOpSpec{
		{name: "node(Address)", query: `query ($id: ID!) { node(id: $id) { id ... on Address { city } } }`, vars: map[string]any{"id": addrID}},
		{name: "node(LegalEntity)", query: `query ($id: ID!) { node(id: $id) { id ... on LegalEntity { entity_code } } }`, vars: map[string]any{"id": leAID}},
		{name: "node(Location)", query: `query ($id: ID!) { node(id: $id) { id ... on Location { location_code } } }`, vars: map[string]any{"id": locID}},
		{name: "node(OrganizationUnit)", query: `query ($id: ID!) { node(id: $id) { id ... on OrganizationUnit { unit_code } } }`, vars: map[string]any{"id": ouID}},
		{name: "node(Party)", query: `query ($id: ID!) { node(id: $id) { id ... on Party { partner_code } } }`, vars: map[string]any{"id": partyID}},
		{name: "node(PartyRole)", query: `query ($id: ID!) { node(id: $id) { id ... on PartyRole { role } } }`, vars: map[string]any{"id": partyRoleID}},
		{name: "node(PartyContact)", query: `query ($id: ID!) { node(id: $id) { id ... on PartyContact { full_name } } }`, vars: map[string]any{"id": contactID}},
		{name: "node(PartyAddress)", query: `query ($id: ID!) { node(id: $id) { id ... on PartyAddress { usage } } }`, vars: map[string]any{"id": paID}},
	}

	for _, qSpec := range append(efQueries, efNodeQueries...) {
		doGQLOK(t, efAdminClient, qSpec.query, qSpec.vars)
		doGQLOK(t, efReadonlyClient, qSpec.query, qSpec.vars)
		requireCustomerForbidden(t, custClient, qSpec.query, qSpec.vars)
		requireAnonUnauthenticated(t, anon, qSpec.query, qSpec.vars)
	}

	// =========================================================================
	// 3. Matriks Penuh 48 Mutasi EF (CUSTOMER, ef_readonly -> AUTHORIZATION;
	//    Anonymous -> AUTHENTICATION) tanpa meninggalkan mutasi liar di DB.
	// =========================================================================
	efMutations := []struct {
		op    string
		query string
		vars  map[string]any
	}{
		// Address (4):
		{"createAddress", fmt.Sprintf(`mutation ($in: CreateAddressInput!) { createAddress(input: $in) { %s } }`, gqlMutationResultFields), map[string]any{"in": map[string]any{"address_line_1": "Blocked St", "city": "Jakarta", "country_code": "ID"}}},
		{"findOrCreateAddress", fmt.Sprintf(`mutation ($in: FindOrCreateAddressInput!) { findOrCreateAddress(input: $in) { %s } }`, gqlMutationResultFields), map[string]any{"in": map[string]any{"address_line_1": "Blocked St", "city": "Jakarta", "country_code": "ID"}}},
		{"updateAddress", fmt.Sprintf(`mutation ($in: UpdateAddressInput!) { updateAddress(input: $in) { %s } }`, gqlMutationResultFields), map[string]any{"in": map[string]any{"id": addrID, "address_line_1": "Blocked", "city": "Jakarta", "country_code": "ID"}}},
		{"deleteAddress", fmt.Sprintf(`mutation ($id: ID!) { deleteAddress(id: $id) { %s } }`, gqlMutationResultFields), map[string]any{"id": addrID}},
		// LegalEntity (7):
		{"createLegalEntity", fmt.Sprintf(`mutation ($in: CreateLegalEntityInput!) { createLegalEntity(input: $in) { %s } }`, gqlMutationResultFields), map[string]any{"in": map[string]any{"entity_code": "BLK-LE", "legal_name": "PT Blocked", "jurisdiction_country_code": "ID"}}},
		{"updateLegalEntity", fmt.Sprintf(`mutation ($in: UpdateLegalEntityInput!) { updateLegalEntity(input: $in) { %s } }`, gqlMutationResultFields), map[string]any{"in": map[string]any{"id": leAID, "legal_name": "PT Blocked", "jurisdiction_country_code": "ID"}}},
		{"setPrimaryOperatingLegalEntity", fmt.Sprintf(`mutation ($id: ID!) { setPrimaryOperatingLegalEntity(id: $id) { %s } }`, gqlMutationResultFields), map[string]any{"id": leAID}},
		{"activateLegalEntity", fmt.Sprintf(`mutation ($id: ID!) { activateLegalEntity(id: $id) { %s } }`, gqlMutationResultFields), map[string]any{"id": leAID}},
		{"suspendLegalEntity", fmt.Sprintf(`mutation ($id: ID!) { suspendLegalEntity(id: $id) { %s } }`, gqlMutationResultFields), map[string]any{"id": leAID}},
		{"dissolveLegalEntity", fmt.Sprintf(`mutation ($id: ID!) { dissolveLegalEntity(id: $id) { %s } }`, gqlMutationResultFields), map[string]any{"id": leAID}},
		{"deleteLegalEntity", fmt.Sprintf(`mutation ($id: ID!) { deleteLegalEntity(id: $id) { %s } }`, gqlMutationResultFields), map[string]any{"id": leAID}},
		// Location (5):
		{"createLocation", fmt.Sprintf(`mutation ($in: CreateLocationInput!) { createLocation(input: $in) { %s } }`, gqlMutationResultFields), map[string]any{"in": map[string]any{"legal_entity_id": leAID, "location_code": "BLK-LOC", "name": "Blocked Loc", "kind": "WAREHOUSE", "timezone": "Asia/Jakarta"}}},
		{"updateLocation", fmt.Sprintf(`mutation ($in: UpdateLocationInput!) { updateLocation(input: $in) { %s } }`, gqlMutationResultFields), map[string]any{"in": map[string]any{"id": locID, "name": "Blocked Loc", "kind": "WAREHOUSE", "timezone": "Asia/Jakarta"}}},
		{"activateLocation", fmt.Sprintf(`mutation ($id: ID!) { activateLocation(id: $id) { %s } }`, gqlMutationResultFields), map[string]any{"id": locID}},
		{"deactivateLocation", fmt.Sprintf(`mutation ($id: ID!) { deactivateLocation(id: $id) { %s } }`, gqlMutationResultFields), map[string]any{"id": locID}},
		{"deleteLocation", fmt.Sprintf(`mutation ($id: ID!) { deleteLocation(id: $id) { %s } }`, gqlMutationResultFields), map[string]any{"id": locID}},
		// OrganizationUnit (6):
		{"createOrganizationUnit", fmt.Sprintf(`mutation ($in: CreateOrganizationUnitInput!) { createOrganizationUnit(input: $in) { %s } }`, gqlMutationResultFields), map[string]any{"in": map[string]any{"legal_entity_id": leAID, "unit_code": "BLK-OU", "name": "Blocked OU", "kind": "DIVISION"}}},
		{"updateOrganizationUnit", fmt.Sprintf(`mutation ($in: UpdateOrganizationUnitInput!) { updateOrganizationUnit(input: $in) { %s } }`, gqlMutationResultFields), map[string]any{"in": map[string]any{"id": ouID, "name": "Blocked OU", "kind": "DIVISION"}}},
		{"assignOrganizationUnitHeadPerson", fmt.Sprintf(`mutation ($id: ID!, $h: ID) { assignOrganizationUnitHeadPerson(id: $id, head_person_id: $h) { %s } }`, gqlMutationResultFields), map[string]any{"id": ouID, "h": efAdminAcct.ID}},
		{"activateOrganizationUnit", fmt.Sprintf(`mutation ($id: ID!) { activateOrganizationUnit(id: $id) { %s } }`, gqlMutationResultFields), map[string]any{"id": ouID}},
		{"deactivateOrganizationUnit", fmt.Sprintf(`mutation ($id: ID!) { deactivateOrganizationUnit(id: $id) { %s } }`, gqlMutationResultFields), map[string]any{"id": ouID}},
		{"deleteOrganizationUnit", fmt.Sprintf(`mutation ($id: ID!) { deleteOrganizationUnit(id: $id) { %s } }`, gqlMutationResultFields), map[string]any{"id": ouID}},
		// Party Core (7):
		{"createParty", fmt.Sprintf(`mutation ($in: CreatePartyInput!) { createParty(input: $in) { %s } }`, gqlMutationResultFields), map[string]any{"in": map[string]any{"partner_code": "BLK-PRT", "kind": "ORGANIZATION", "display_name": "PT Blocked"}}},
		{"updatePartyProfile", fmt.Sprintf(`mutation ($in: UpdatePartyProfileInput!) { updatePartyProfile(input: $in) { %s } }`, gqlMutationResultFields), map[string]any{"in": map[string]any{"id": partyID, "kind": "ORGANIZATION", "display_name": "PT Blocked", "preferred_language": "id"}}},
		{"updatePartyTaxProfile", fmt.Sprintf(`mutation ($in: UpdatePartyTaxProfileInput!) { updatePartyTaxProfile(input: $in) { %s } }`, gqlMutationResultFields), map[string]any{"in": map[string]any{"id": partyID}}},
		{"verifyAndActivateParty", fmt.Sprintf(`mutation ($id: ID!) { verifyAndActivateParty(id: $id) { %s } }`, gqlMutationResultFields), map[string]any{"id": partyID}},
		{"suspendParty", fmt.Sprintf(`mutation ($id: ID!) { suspendParty(id: $id) { %s } }`, gqlMutationResultFields), map[string]any{"id": partyID}},
		{"deactivateParty", fmt.Sprintf(`mutation ($id: ID!) { deactivateParty(id: $id) { %s } }`, gqlMutationResultFields), map[string]any{"id": partyID}},
		{"deleteParty", fmt.Sprintf(`mutation ($id: ID!) { deleteParty(id: $id) { %s } }`, gqlMutationResultFields), map[string]any{"id": partyID}},
		// PartyRole (5):
		{"assignPartyRole", fmt.Sprintf(`mutation ($in: AssignPartyRoleInput!) { assignPartyRole(input: $in) { %s } }`, gqlMutationResultFields), map[string]any{"in": map[string]any{"party_id": partyID, "role": "VENDOR"}}},
		{"updatePartyRoleNote", fmt.Sprintf(`mutation ($in: UpdatePartyRoleNoteInput!) { updatePartyRoleNote(input: $in) { %s } }`, gqlMutationResultFields), map[string]any{"in": map[string]any{"id": partyRoleID}}},
		{"activatePartyRole", fmt.Sprintf(`mutation ($r: ID!) { activatePartyRole(role_id: $r) { %s } }`, gqlMutationResultFields), map[string]any{"r": partyRoleID}},
		{"deactivatePartyRole", fmt.Sprintf(`mutation ($r: ID!) { deactivatePartyRole(role_id: $r) { %s } }`, gqlMutationResultFields), map[string]any{"r": partyRoleID}},
		{"removePartyRole", fmt.Sprintf(`mutation ($r: ID!) { removePartyRole(role_id: $r) { %s } }`, gqlMutationResultFields), map[string]any{"r": partyRoleID}},
		// PartyContact & Portal (10):
		{"addPartyContact", fmt.Sprintf(`mutation ($in: AddPartyContactInput!) { addPartyContact(input: $in) { %s } }`, gqlMutationResultFields), map[string]any{"in": map[string]any{"party_id": partyID, "full_name": "Blocked PIC", "contact_role": "pic"}}},
		{"updatePartyContact", fmt.Sprintf(`mutation ($in: UpdatePartyContactInput!) { updatePartyContact(input: $in) { %s } }`, gqlMutationResultFields), map[string]any{"in": map[string]any{"id": contactID, "full_name": "Blocked PIC", "contact_role": "pic"}}},
		{"setPrimaryPartyContact", fmt.Sprintf(`mutation ($p: ID!, $c: ID!) { setPrimaryPartyContact(party_id: $p, contact_id: $c) { %s } }`, gqlMutationResultFields), map[string]any{"p": partyID, "c": contactID}},
		{"requestPartyPortalLink", fmt.Sprintf(`mutation ($c: ID!, $u: ID!) { requestPartyPortalLink(contact_id: $c, iam_user_id: $u) { %s } }`, gqlMutationResultFields), map[string]any{"c": contactID, "u": efAdminAcct.ID}},
		{"activatePartyPortalAccess", fmt.Sprintf(`mutation ($c: ID!, $u: ID!) { activatePartyPortalAccess(contact_id: $c, iam_user_id: $u) { %s } }`, gqlMutationResultFields), map[string]any{"c": contactID, "u": efAdminAcct.ID}},
		{"revokePartyPortalAccess", fmt.Sprintf(`mutation ($c: ID!) { revokePartyPortalAccess(contact_id: $c) { %s } }`, gqlMutationResultFields), map[string]any{"c": contactID}},
		{"unlinkPartyPortalUser", fmt.Sprintf(`mutation ($c: ID!) { unlinkPartyPortalUser(contact_id: $c) { %s } }`, gqlMutationResultFields), map[string]any{"c": contactID}},
		{"activatePartyContact", fmt.Sprintf(`mutation ($c: ID!) { activatePartyContact(contact_id: $c) { %s } }`, gqlMutationResultFields), map[string]any{"c": contactID}},
		{"deactivatePartyContact", fmt.Sprintf(`mutation ($c: ID!) { deactivatePartyContact(contact_id: $c) { %s } }`, gqlMutationResultFields), map[string]any{"c": contactID}},
		{"deletePartyContact", fmt.Sprintf(`mutation ($c: ID!) { deletePartyContact(contact_id: $c) { %s } }`, gqlMutationResultFields), map[string]any{"c": contactID}},
		// PartyAddressLink (4):
		{"linkPartyAddress", fmt.Sprintf(`mutation ($in: LinkPartyAddressInput!) { linkPartyAddress(input: $in) { %s } }`, gqlMutationResultFields), map[string]any{"in": map[string]any{"party_id": partyID, "address_id": addrID, "usage": "BILLING"}}},
		{"updatePartyAddressLink", fmt.Sprintf(`mutation ($in: UpdatePartyAddressLinkInput!) { updatePartyAddressLink(input: $in) { %s } }`, gqlMutationResultFields), map[string]any{"in": map[string]any{"id": paID, "usage": "BILLING"}}},
		{"setPrimaryPartyAddressLink", fmt.Sprintf(`mutation ($id: ID!) { setPrimaryPartyAddressLink(link_id: $id) { %s } }`, gqlMutationResultFields), map[string]any{"id": paID}},
		{"unlinkPartyAddress", fmt.Sprintf(`mutation ($id: ID!) { unlinkPartyAddress(link_id: $id) { %s } }`, gqlMutationResultFields), map[string]any{"id": paID}},
	}
	if len(efMutations) != 48 {
		t.Fatalf("expected 48 EF mutations in matrix, got %d", len(efMutations))
	}

	for i, tc := range efMutations {
		if i%8 == 0 {
			resetGatewayRateLimits(t, cfg)
		}
		requireCustomerMutationForbidden(t, custClient, tc.query, tc.vars, tc.op)
		requireCustomerMutationForbidden(t, efReadonlyClient, tc.query, tc.vars, tc.op)
		requireAnonMutationUnauthenticated(t, anon, tc.query, tc.vars, tc.op)
	}
	resetGatewayRateLimits(t, cfg)

	// =========================================================================
	// 4. Scoped EF RBAC Cross-Scope Matrix (ef_operator scoped ke leA vs leB,
	//    ditambah batas scope facility/location & organization_unit)
	// =========================================================================
	_, scopedAClient := provisionEFUserWithPerms(t, admin, cfg, "ef_scope_leA", []string{
		"enterprise_foundation.legal_entities.read",
		"enterprise_foundation.legal_entities.write",
		"enterprise_foundation.locations.read",
		"enterprise_foundation.locations.write",
		"enterprise_foundation.organization_units.read",
		"enterprise_foundation.organization_units.write",
	}, "LEGAL_ENTITY", &leARawID)

	// Authorized pada leA + daftar lokasi di leA + membuat organizationUnit pada leA:
	_ = doGQLOK(t, scopedAClient, `query ($id: ID!) {
		legalEntity(id: $id) { id entity_code }
		validateLegalEntityScopeTarget(id: $id) { eligible }
		locationsByLegalEntityAndKind(legal_entity_id: $id, kind: WAREHOUSE, first: 5) { total_count }
	}`, map[string]any{"id": leAID})
	createdScopedOU := requireMutationChildOK(t, doGQLOK(t, scopedAClient, fmt.Sprintf(`mutation ($in: CreateOrganizationUnitInput!) {
		createOrganizationUnit(input: $in) { organization_unit { id _id } %s }
	}`, gqlMutationResultFields), map[string]any{
		"in": map[string]any{"legal_entity_id": leAID, "unit_code": "SCP-OU-" + tagUpper, "name": "Scoped OU", "kind": "TEAM"},
	}), "createOrganizationUnit", "organization_unit")
	scopedOURawID := requireNonEmptyString(t, createdScopedOU, "_id")
	t.Cleanup(func() { cleanupEFEntityRows(t, cfg, "organization_units", scopedOURawID) })

	// Unauthorized / Cross-Scope pada leB, scope facility spesifik tanpa grant facility, dan listing global:
	requireCustomerForbidden(t, scopedAClient, `query ($id: ID!) { legalEntity(id: $id) { id } }`, map[string]any{"id": leBID})
	requireCustomerForbidden(t, scopedAClient, `query ($loc: ID!) { location(id: $loc) { id } }`, map[string]any{"loc": locID})
	requireCustomerForbidden(t, scopedAClient, `query { legalEntities(first: 5) { total_count } }`, nil)
	requireCustomerMutationForbidden(t, scopedAClient, fmt.Sprintf(`mutation ($in: CreateLocationInput!) {
		createLocation(input: $in) { %s }
	}`, gqlMutationResultFields), map[string]any{
		"in": map[string]any{"legal_entity_id": leBID, "location_code": "CROSS-LOC", "name": "Cross", "kind": "WAREHOUSE", "timezone": "Asia/Jakarta"},
	}, "createLocation")
	requireCustomerMutationForbidden(t, scopedAClient, fmt.Sprintf(`mutation ($in: CreateOrganizationUnitInput!) {
		createOrganizationUnit(input: $in) { %s }
	}`, gqlMutationResultFields), map[string]any{
		"in": map[string]any{"legal_entity_id": leBID, "unit_code": "CROSS-OU", "name": "Cross OU", "kind": "DIVISION"},
	}, "createOrganizationUnit")
}
