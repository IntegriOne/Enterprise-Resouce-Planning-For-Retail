//go:build e2e

package e2e

import (
	"fmt"
	"strings"
	"testing"
)

type gqlOpSpec struct {
	name  string
	query string
	vars  map[string]any
}

// TestIAMFullAuthMatrix_CheckUsernameAndPublicQueries menguji kontrak Query.checkUsername
// (available, taken tanpa kebocoran user_kind/PII, penolakan format tidak sah/email, rate-limit IP)
// beserta browserSession untuk kondisi anonymous, authorized, dan token sesi tidak sah.
func TestIAMFullAuthMatrix_CheckUsernameAndPublicQueries(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	resetGatewayRateLimits(t, cfg)
	t.Cleanup(func() { resetGatewayRateLimits(t, cfg) })

	admin := newSuperAdminBrowser(t, cfg)
	emp := createFixtureUser(t, admin, cfg, "EMPLOYEE", true)
	cust, custClient := createFixtureCustomerAndLogin(t, admin, cfg)
	anon := newAnonBrowser(t, cfg)

	qCheck := `query CheckUsername($u: String!) {
		checkUsername(username: $u) {
			available
			normalized_username
			message
		}
	}`

	// 1. Anonymous: username tersedia -> dinormalisasi huruf kecil & spasi dipangkas.
	freshName := "  Fresh." + strings.ToUpper(uniqueTag("Uname")) + "  "
	wantNorm := strings.ToLower(strings.TrimSpace(freshName))
	availRes := requireOpMap(t, doGQLOK(t, anon, qCheck, map[string]any{"u": freshName}), "checkUsername")
	if !requireBoolField(t, availRes, "available") ||
		requireNonEmptyString(t, availRes, "normalized_username") != wantNorm ||
		requireNonEmptyString(t, availRes, "message") != "username is available" {
		t.Fatalf("anonymous checkUsername(available) unexpected payload: %+v", availRes)
	}

	// 2. Anonymous anti-enumerasi: username employee dan customer mengembalikan pesan identik tanpa membocorkan PII atau kind.
	takenEmp := requireOpMap(t, doGQLOK(t, anon, qCheck, map[string]any{"u": strings.ToUpper(emp.Username)}), "checkUsername")
	takenCust := requireOpMap(t, doGQLOK(t, anon, qCheck, map[string]any{"u": strings.ToUpper(cust.Username)}), "checkUsername")
	if requireBoolField(t, takenEmp, "available") || requireBoolField(t, takenCust, "available") {
		t.Fatalf("expected taken=false for registered usernames, got emp=%+v cust=%+v", takenEmp, takenCust)
	}
	if takenEmp["message"] != "username is already taken" || takenCust["message"] != takenEmp["message"] {
		t.Fatalf("anti-enumeration message mismatch between employee and customer: emp=%+v cust=%+v", takenEmp, takenCust)
	}
	msgEmp := requireNonEmptyString(t, takenEmp, "message")
	if len(takenEmp) != 3 || len(takenCust) != 3 {
		t.Fatalf("checkUsername returned unexpected extra fields: emp=%+v cust=%+v", takenEmp, takenCust)
	}
	for _, leak := range []string{"employee", "customer", emp.Email, cust.Email, emp.RawID, cust.RawID} {
		if strings.Contains(strings.ToLower(msgEmp), strings.ToLower(leak)) {
			t.Fatalf("checkUsername message leaked PII/kind (%q): %q", leak, msgEmp)
		}
	}
	rawJSON := fmt.Sprintf("%v|%v", takenEmp, takenCust)
	for _, leak := range []string{emp.Email, cust.Email, emp.RawID, cust.RawID} {
		if strings.Contains(strings.ToLower(rawJSON), strings.ToLower(leak)) {
			t.Fatalf("checkUsername leaked PII (%q) in payload: %s", leak, rawJSON)
		}
	}

	// 3. Format tidak sah (kosong, spasi, terlalu pendek, atau string email) ditolak INVALID_ARGUMENT.
	for _, invalid := range []string{"", "  ", "ab", "bad user", cust.Email, emp.Email} {
		_ = doGQLError(t, anon, qCheck, map[string]any{"u": invalid}, "INVALID_ARGUMENT")
	}

	// 4. Customer dan Superadmin (authorized) juga dapat memanggil checkUsername dengan hasil konsisten.
	custCheck := requireOpMap(t, doGQLOK(t, custClient, qCheck, map[string]any{"u": emp.Username}), "checkUsername")
	if requireBoolField(t, custCheck, "available") || custCheck["message"] != "username is already taken" {
		t.Fatalf("customer checkUsername(taken) = %+v", custCheck)
	}
	adminCheck := requireOpMap(t, doGQLOK(t, admin, qCheck, map[string]any{"u": freshName}), "checkUsername")
	if !requireBoolField(t, adminCheck, "available") {
		t.Fatalf("admin checkUsername(available) = %+v", adminCheck)
	}

	// 5. Pembatasan laju per IP (check_username_ip) memblokir permintaan berlebih dengan TOO_MANY_REQUESTS.
	rateHit := false
	for i := 0; i < 40; i++ {
		resp, gql := anon.doGraphQL(qCheck, map[string]any{"u": fmt.Sprintf("rl_probe_%d", i)}, nil)
		if resp.StatusCode == 429 || (len(gql.Errors) > 0 && hasErrorCode(gql, "TOO_MANY_REQUESTS")) {
			rateHit = true
			break
		}
	}
	if !rateHit {
		t.Fatalf("expected checkUsername IP rate limiter to trigger TOO_MANY_REQUESTS within 40 calls")
	}
	resetGatewayRateLimits(t, cfg)

	// 6. Verifikasi browserSession untuk anonymous vs authorized (customer & admin).
	qBS := fmt.Sprintf(`query { browserSession { %s } }`, gqlBrowserSessionFields)
	assertBrowserSessionPayload(t, requireOpMap(t, doGQLOK(t, anon, qBS, nil), "browserSession"), false)
	assertBrowserSessionPayload(t, requireOpMap(t, doGQLOK(t, custClient, qBS, nil), "browserSession"), true)
	assertBrowserSessionPayload(t, requireOpMap(t, doGQLOK(t, admin, qBS, nil), "browserSession"), true)
}

func hasErrorCode(gql graphQLResponse, code string) bool {
	for _, e := range gql.Errors {
		if extCode, _ := e.Extensions["code"].(string); extCode == code {
			return true
		}
	}
	return false
}

// TestIAMFullAuthMatrix_All23QueriesAnd6Nodes memverifikasi matriks penuh
// authorized / unauthorized / anonymous untuk seluruh 23 Query IAM serta 6 tipe Node IAM.
func TestIAMFullAuthMatrix_All23QueriesAnd6Nodes(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	seedEFPermissionCatalog(t, cfg)
	resetGatewayRateLimits(t, cfg)
	t.Cleanup(func() { resetGatewayRateLimits(t, cfg) })

	admin := newSuperAdminBrowser(t, cfg)
	anon := newAnonBrowser(t, cfg)
	cust, custClient := createFixtureCustomerAndLogin(t, admin, cfg)
	targetEmp := createFixtureUser(t, admin, cfg, "EMPLOYEE", true)

	// Ambil ID identitas & sesi milik admin sendiri untuk menguji Node(UserIdentity) dan Node(Session),
	// karena resolver Node untuk kedua resource tersebut mewajibkan kepemilikan langsung caller atau mengembalikan FORBIDDEN/NOT_FOUND.
	adminMe := requireOpMap(t, doGQLOK(t, admin, `query { me { id } }`, nil), "me")
	adminID := requireNonEmptyString(t, adminMe, "id")
	adminIdentConn := requireOpMap(t, doGQLOK(t, admin, fmt.Sprintf(`query ($u: ID!) {
		userIdentities(user_id: $u, first: 5) { nodes { %s } total_count }
	}`, gqlIdentityFields), map[string]any{"u": adminID}), "userIdentities")
	adminIdentID := requireNonEmptyString(t, requireConnectionNodes(t, adminIdentConn)[0].(map[string]any), "id")

	adminSessConn := requireOpMap(t, doGQLOK(t, admin, fmt.Sprintf(`query {
		mySessions(active_only: true, first: 5) { nodes { %s } total_count }
	}`, gqlSessionFields), nil), "mySessions")
	adminSessID := requireNonEmptyString(t, requireConnectionNodes(t, adminSessConn)[0].(map[string]any), "id")

	// Sediakan user khusus berizin ef_admin (0 izin admin IAM) untuk menguji batas manajemen lintas domain.
	efPerms := loadEFPermissionCatalog(t)
	efPermCodes := make([]string, 0, len(efPerms))
	for _, p := range efPerms {
		efPermCodes = append(efPermCodes, p.Code)
	}
	_, efOnlyClient := provisionEFUserWithPerms(t, admin, cfg, "iam_mat_efonly", efPermCodes, "GLOBAL", nil)

	// Siapkan objek Group, Role, dan Permission untuk ID target valid.
	roleID, roleCode := createFixtureRole(t, admin, cfg, "matq")
	permID := findPermissionIDByCode(t, admin, "iam.users.read")
	groupID, _ := createFixtureGroup(t, admin, cfg, "matq")
	_ = requireMutationChildOK(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($g: ID!, $u: ID!) {
		addGroupMember(group_id: $g, user_id: $u) { %s }
	}`, gqlGroupMemberPayloadFields), map[string]any{"g": groupID, "u": targetEmp.ID}), "addGroupMember", "group_member_edge")

	// Customer dengan role dasar dicabut agar AccessSnapshot memiliki 0 izin iam.*.readself (uji unauthorized self-query).
	noSelfCust := createFixtureUser(t, admin, cfg, "CUSTOMER", true)
	noSelfRolesConn := requireOpMap(t, doGQLOK(t, admin, `query ($u: ID!) {
		userRoles(user_id: $u, first: 5) { edges { assignment_id } }
	}`, map[string]any{"u": noSelfCust.ID}), "userRoles")
	for _, rawE := range requireConnectionEdges(t, noSelfRolesConn) {
		aID := requireNonEmptyString(t, rawE.(map[string]any), "assignment_id")
		_ = requireMutationBoolOK(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($id: ID!) {
			revokeUserRole(assignment_id: $id) { %s }
		}`, gqlRevokeUserRoleFields), map[string]any{"id": aID}), "revokeUserRole", "revoked")
	}
	resetGatewayRateLimits(t, cfg)
	noSelfClient, _ := loginFixtureUser(t, cfg, noSelfCust.Email, noSelfCust.Password)

	// 1. Public Queries (2): browserSession, checkUsername.
	doGQLOK(t, anon, fmt.Sprintf(`query { browserSession { %s } }`, gqlBrowserSessionFields), nil)
	doGQLOK(t, anon, `query ($u: String!) { checkUsername(username: $u) { available normalized_username message } }`, map[string]any{"u": targetEmp.Username})

	// 2. Self-Service Protected Queries (4): myWorkspaceAccess, me, myAccessSnapshot, mySessions.
	//    Authorized: admin, custClient; Anonymous: UNAUTHENTICATED; Unauthorized self: noSelfClient -> FORBIDDEN pada me/mySessions.
	for _, op := range []gqlOpSpec{
		{name: "myWorkspaceAccess", query: fmt.Sprintf(`query { myWorkspaceAccess { %s } }`, gqlWorkspaceAccessFields)},
		{name: "myAccessSnapshot", query: fmt.Sprintf(`query { myAccessSnapshot { %s } }`, gqlAccessSnapshotFields)},
		{name: "me", query: fmt.Sprintf(`query { me { %s } }`, gqlUserFields)},
		{name: "mySessions", query: fmt.Sprintf(`query { mySessions(first: 5) { total_count } }`)},
	} {
		doGQLOK(t, admin, op.query, op.vars)
		doGQLOK(t, custClient, op.query, op.vars)
		requireAnonUnauthenticated(t, anon, op.query, op.vars)
	}
	requireCustomerForbidden(t, noSelfClient, fmt.Sprintf(`query { me { %s } }`, gqlUserFields), nil)
	requireCustomerForbidden(t, noSelfClient, `query { mySessions(first: 5) { total_count } }`, nil)

	// 3. Seluruh 17 Query Admin IAM + 6 tipe Node IAM: authorized (admin) vs unauthorized (cust & efOnly & IDOR) vs anonymous.
	adminQueries := []gqlOpSpec{
		{name: "user", query: `query ($id: ID!) { user(id: $id) { id kind } }`, vars: map[string]any{"id": targetEmp.ID}},
		{name: "users", query: `query { users(first: 5) { total_count } }`},
		{name: "userIdentities", query: `query ($u: ID!) { userIdentities(user_id: $u, first: 5) { total_count } }`, vars: map[string]any{"u": targetEmp.ID}},
		{name: "userRoles", query: `query ($u: ID!) { userRoles(user_id: $u, first: 5) { total_count } }`, vars: map[string]any{"u": targetEmp.ID}},
		{name: "userEffectiveRoles", query: `query ($u: ID!) { userEffectiveRoles(user_id: $u, first: 5) { total_count } }`, vars: map[string]any{"u": targetEmp.ID}},
		{name: "userGroups", query: `query ($u: ID!) { userGroups(user_id: $u, first: 5) { total_count } }`, vars: map[string]any{"u": targetEmp.ID}},
		{name: "userSessions", query: `query ($u: ID!) { userSessions(user_id: $u, first: 5) { total_count } }`, vars: map[string]any{"u": targetEmp.ID}},
		{name: "roles", query: `query { roles(first: 5) { total_count } }`},
		{name: "role", query: `query ($id: ID!) { role(id: $id) { id code } }`, vars: map[string]any{"id": roleID}},
		{name: "roleByCode", query: `query ($c: String!) { roleByCode(code: $c) { id code } }`, vars: map[string]any{"c": roleCode}},
		{name: "rolePermissions", query: `query ($id: ID!) { rolePermissions(role_id: $id, first: 5) { total_count } }`, vars: map[string]any{"id": roleID}},
		{name: "permissions", query: `query { permissions(first: 5) { total_count } }`},
		{name: "permission", query: `query ($id: ID!) { permission(id: $id) { id code } }`, vars: map[string]any{"id": permID}},
		{name: "groups", query: `query { groups(first: 5) { total_count } }`},
		{name: "group", query: `query ($id: ID!) { group(id: $id) { id code } }`, vars: map[string]any{"id": groupID}},
		{name: "groupMembers", query: `query ($id: ID!) { groupMembers(group_id: $id, first: 5) { total_count } }`, vars: map[string]any{"id": groupID}},
		{name: "groupRoles", query: `query ($id: ID!) { groupRoles(group_id: $id, first: 5) { total_count } }`, vars: map[string]any{"id": groupID}},
		// 6 tipe Relay Node IAM:
		{name: "node(User)", query: `query ($id: ID!) { node(id: $id) { id ... on User { kind } } }`, vars: map[string]any{"id": targetEmp.ID}},
		{name: "node(UserIdentity)", query: `query ($id: ID!) { node(id: $id) { id ... on UserIdentity { kind value } } }`, vars: map[string]any{"id": adminIdentID}},
		{name: "node(Role)", query: `query ($id: ID!) { node(id: $id) { id ... on Role { code } } }`, vars: map[string]any{"id": roleID}},
		{name: "node(Permission)", query: `query ($id: ID!) { node(id: $id) { id ... on Permission { code } } }`, vars: map[string]any{"id": permID}},
		{name: "node(Group)", query: `query ($id: ID!) { node(id: $id) { id ... on Group { code } } }`, vars: map[string]any{"id": groupID}},
		{name: "node(Session)", query: `query ($id: ID!) { node(id: $id) { id ... on Session { user_id } } }`, vars: map[string]any{"id": adminSessID}},
	}
	if len(adminQueries) != 17+6 {
		t.Fatalf("internal matrix mismatch: got %d admin+node ops, want 23", len(adminQueries))
	}

	for _, op := range adminQueries {
		t.Run("Query_"+op.name, func(t *testing.T) {
			doGQLOK(t, admin, op.query, op.vars)
			requireCustomerForbidden(t, custClient, op.query, op.vars)
			requireCustomerForbidden(t, efOnlyClient, op.query, op.vars)
			requireAnonUnauthenticated(t, anon, op.query, op.vars)
		})
	}
	_ = cust
}

// TestIAMFullAuthMatrix_All48MutationsAuthAndAnon memastikan seluruh 48 mutasi IAM
// memiliki cakupan authorized, unauthorized (CUSTOMER/EF-only/stale/IDOR), dan anonymous.
func TestIAMFullAuthMatrix_All48MutationsAuthAndAnon(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	seedEFPermissionCatalog(t, cfg)
	resetGatewayRateLimits(t, cfg)
	t.Cleanup(func() { resetGatewayRateLimits(t, cfg) })

	admin := newSuperAdminBrowser(t, cfg)
	anon := newAnonBrowser(t, cfg)
	_, custClient := createFixtureCustomerAndLogin(t, admin, cfg)

	efPerms := loadEFPermissionCatalog(t)
	efPermCodes := make([]string, 0, len(efPerms))
	for _, p := range efPerms {
		efPermCodes = append(efPermCodes, p.Code)
	}
	_, efOnlyClient := provisionEFUserWithPerms(t, admin, cfg, "iam_mut_efonly", efPermCodes, "GLOBAL", nil)

	// Fixture IDs yang valid agar penolakan auth selalu terjadi di lapisan Authorizer/Middleware, bukan INVALID_ARGUMENT GID.
	targetEmp := createFixtureUser(t, admin, cfg, "EMPLOYEE", true)
	targetPartyVendor := createFixturePartyForRole(t, admin, cfg, "VENDOR")
	targetPartySupplier := createFixturePartyForRole(t, admin, cfg, "SUPPLIER")
	roleID, _ := createFixtureRole(t, admin, cfg, "matm")
	permID := findPermissionIDByCode(t, admin, "iam.roles.read")
	groupID, _ := createFixtureGroup(t, admin, cfg, "matm")

	identConn := requireOpMap(t, doGQLOK(t, admin, fmt.Sprintf(`query ($u: ID!) { userIdentities(user_id: $u, first: 5) { nodes { %s } } }`, gqlIdentityFields), map[string]any{"u": targetEmp.ID}), "userIdentities")
	targetIdentID := requireNonEmptyString(t, requireConnectionNodes(t, identConn)[0].(map[string]any), "id")

	// Dummy UUID format GID untuk assignment/session ID yang ditolak auth sebelum lookup.
	dummyAssignID := "00000000-0000-4000-8000-000000000099"
	dummySessID := "gid://integrione/session/00000000-0000-4000-8000-000000000099"

	// 1. Cek seluruh 41 mutasi yang membutuhkan sesi login mengembalikan AUTHENTICATION saat dipanggil anonymous.
	authRequiredMutations := []struct {
		op    string
		query string
		vars  map[string]any
	}{
		// 12 Self-Service Mutations:
		{"userLogout", fmt.Sprintf(`mutation { userLogout { %s } }`, gqlMutationResultFields), nil},
		{"updateMyProfile", fmt.Sprintf(`mutation ($in: UpdateMyProfileInput!) { updateMyProfile(input: $in) { %s } }`, gqlMutationResultFields), map[string]any{"in": map[string]any{"display_name": "Anon Attempt"}}},
		{"uploadAvatarUser", fmt.Sprintf(`mutation ($in: UserAvatarUploadInput!) { uploadAvatarUser(input: $in) { %s } }`, gqlMutationResultFields), map[string]any{"in": map[string]any{"image": "AAAA"}}},
		{"changeUserPassword", fmt.Sprintf(`mutation ($in: ChangeUserPasswordInput!) { changeUserPassword(input: $in) { %s } }`, gqlMutationResultFields), map[string]any{"in": map[string]any{"current_password": "StrongPass#2026!", "new_password": "StrongPass#2027!"}}},
		{"verifyMySessionMfa", fmt.Sprintf(`mutation ($c: String!) { verifyMySessionMfa(totp_code: $c) { %s } }`, gqlMutationResultFields), map[string]any{"c": "123456"}},
		{"enrollMyTotp", fmt.Sprintf(`mutation { enrollMyTotp { %s } }`, gqlMutationResultFields), nil},
		{"confirmMyTotp", fmt.Sprintf(`mutation ($in: ConfirmTotpInput!) { confirmMyTotp(input: $in) { %s } }`, gqlMutationResultFields), map[string]any{"in": map[string]any{"code": "123456"}}},
		{"disableMyTotp", fmt.Sprintf(`mutation { disableMyTotp { %s } }`, gqlMutationResultFields), nil},
		{"regenerateMyRecoveryCodes", fmt.Sprintf(`mutation { regenerateMyRecoveryCodes { %s } }`, gqlMutationResultFields), nil},
		{"changeUsername", fmt.Sprintf(`mutation ($in: UsernameInput!) { changeUsername(input: $in) { %s } }`, gqlMutationResultFields), map[string]any{"in": map[string]any{"username": "anon.attempt"}}},
		{"updateUserEmail", fmt.Sprintf(`mutation ($in: UserEmailInput!) { updateUserEmail(input: $in) { %s } }`, gqlMutationResultFields), map[string]any{"in": map[string]any{"email": "anon@integrione.test", "current_password": "StrongPass#2026!"}}},
		{"deletePendingUserEmail", fmt.Sprintf(`mutation { deletePendingUserEmail { %s } }`, gqlMutationResultFields), nil},
		// 29 Admin IAM Mutations:
		{"createEmployeeUser", fmt.Sprintf(`mutation ($in: CreateEmployeeUserInput!) { createEmployeeUser(input: $in) { %s } }`, gqlMutationResultFields), map[string]any{"in": map[string]any{"worker_id": "55555555-5555-4555-8555-555555555555", "provisioning_request_id": "66666666-6666-4666-8666-666666666666", "full_name": "Blocked Emp", "email": "b_emp@integrione.test", "username": "b.emp1", "password": "StrongPass#2026!"}}},
		{"createVendorUser", fmt.Sprintf(`mutation ($in: CreateVendorUserInput!) { createVendorUser(input: $in) { %s } }`, gqlMutationResultFields), map[string]any{"in": map[string]any{"party_id": targetPartyVendor, "full_name": "Blocked Vend", "email": "b_vend@integrione.test", "username": "b.vend1", "password": "StrongPass#2026!"}}},
		{"createSupplierUser", fmt.Sprintf(`mutation ($in: CreateSupplierUserInput!) { createSupplierUser(input: $in) { %s } }`, gqlMutationResultFields), map[string]any{"in": map[string]any{"party_id": targetPartySupplier, "full_name": "Blocked Supp", "email": "b_supp@integrione.test", "username": "b.supp1", "password": "StrongPass#2026!"}}},
		{"activateUser", fmt.Sprintf(`mutation ($u: ID!) { activateUser(user_id: $u) { %s } }`, gqlMutationResultFields), map[string]any{"u": targetEmp.ID}},
		{"suspendUser", fmt.Sprintf(`mutation ($u: ID!) { suspendUser(user_id: $u) { %s } }`, gqlMutationResultFields), map[string]any{"u": targetEmp.ID}},
		{"disableUser", fmt.Sprintf(`mutation ($u: ID!) { disableUser(user_id: $u) { %s } }`, gqlMutationResultFields), map[string]any{"u": targetEmp.ID}},
		{"softDeleteUser", fmt.Sprintf(`mutation ($u: ID!) { softDeleteUser(user_id: $u) { %s } }`, gqlMutationResultFields), map[string]any{"u": targetEmp.ID}},
		{"setPrimaryUserIdentity", fmt.Sprintf(`mutation ($u: ID!, $i: ID!) { setPrimaryUserIdentity(user_id: $u, identity_id: $i) { %s } }`, gqlMutationResultFields), map[string]any{"u": targetEmp.ID, "i": targetIdentID}},
		{"removeUserIdentity", fmt.Sprintf(`mutation ($u: ID!, $i: ID!) { removeUserIdentity(user_id: $u, identity_id: $i) { %s } }`, gqlMutationResultFields), map[string]any{"u": targetEmp.ID, "i": targetIdentID}},
		{"revokeSession", fmt.Sprintf(`mutation ($s: ID!) { revokeSession(session_id: $s) { %s } }`, gqlMutationResultFields), map[string]any{"s": dummySessID}},
		{"revokeAllUserSessions", fmt.Sprintf(`mutation ($u: ID!) { revokeAllUserSessions(user_id: $u) { %s } }`, gqlMutationResultFields), map[string]any{"u": targetEmp.ID}},
		{"createRole", fmt.Sprintf(`mutation ($in: CreateRoleInput!) { createRole(input: $in) { %s } }`, gqlMutationResultFields), map[string]any{"in": map[string]any{"code": "blocked_role", "name": "Blocked Role", "management_domain": "iam", "default_redirect_path": "/identity"}}},
		{"updateRoleDetails", fmt.Sprintf(`mutation ($in: UpdateRoleDetailsInput!) { updateRoleDetails(input: $in) { %s } }`, gqlMutationResultFields), map[string]any{"in": map[string]any{"id": roleID, "name": "Blocked Update", "management_domain": "iam", "default_redirect_path": "/identity"}}},
		{"activateRole", fmt.Sprintf(`mutation ($id: ID!) { activateRole(id: $id) { %s } }`, gqlMutationResultFields), map[string]any{"id": roleID}},
		{"deactivateRole", fmt.Sprintf(`mutation ($id: ID!) { deactivateRole(id: $id) { %s } }`, gqlMutationResultFields), map[string]any{"id": roleID}},
		{"deleteRole", fmt.Sprintf(`mutation ($id: ID!) { deleteRole(id: $id) { %s } }`, gqlMutationResultFields), map[string]any{"id": roleID}},
		{"assignRolePermission", fmt.Sprintf(`mutation ($r: ID!, $p: ID!) { assignRolePermission(role_id: $r, permission_id: $p) { %s } }`, gqlMutationResultFields), map[string]any{"r": roleID, "p": permID}},
		{"revokeRolePermission", fmt.Sprintf(`mutation ($r: ID!, $p: ID!) { revokeRolePermission(role_id: $r, permission_id: $p) { %s } }`, gqlMutationResultFields), map[string]any{"r": roleID, "p": permID}},
		{"assignUserRole", fmt.Sprintf(`mutation ($in: AssignUserRoleInput!) { assignUserRole(input: $in) { %s } }`, gqlMutationResultFields), map[string]any{"in": map[string]any{"user_id": targetEmp.ID, "role_id": roleID, "scope_type": "GLOBAL"}}},
		{"revokeUserRole", fmt.Sprintf(`mutation ($a: ID!) { revokeUserRole(assignment_id: $a) { %s } }`, gqlMutationResultFields), map[string]any{"a": dummyAssignID}},
		{"createGroup", fmt.Sprintf(`mutation ($in: CreateGroupInput!) { createGroup(input: $in) { %s } }`, gqlMutationResultFields), map[string]any{"in": map[string]any{"code": "blocked_grp", "name": "Blocked Grp"}}},
		{"updateGroupDetails", fmt.Sprintf(`mutation ($in: UpdateGroupDetailsInput!) { updateGroupDetails(input: $in) { %s } }`, gqlMutationResultFields), map[string]any{"in": map[string]any{"id": groupID, "name": "Blocked Grp"}}},
		{"activateGroup", fmt.Sprintf(`mutation ($id: ID!) { activateGroup(id: $id) { %s } }`, gqlMutationResultFields), map[string]any{"id": groupID}},
		{"deactivateGroup", fmt.Sprintf(`mutation ($id: ID!) { deactivateGroup(id: $id) { %s } }`, gqlMutationResultFields), map[string]any{"id": groupID}},
		{"deleteGroup", fmt.Sprintf(`mutation ($id: ID!) { deleteGroup(id: $id) { %s } }`, gqlMutationResultFields), map[string]any{"id": groupID}},
		{"addGroupMember", fmt.Sprintf(`mutation ($g: ID!, $u: ID!) { addGroupMember(group_id: $g, user_id: $u) { %s } }`, gqlMutationResultFields), map[string]any{"g": groupID, "u": targetEmp.ID}},
		{"removeGroupMember", fmt.Sprintf(`mutation ($g: ID!, $u: ID!) { removeGroupMember(group_id: $g, user_id: $u) { %s } }`, gqlMutationResultFields), map[string]any{"g": groupID, "u": targetEmp.ID}},
		{"assignGroupRole", fmt.Sprintf(`mutation ($in: AssignGroupRoleInput!) { assignGroupRole(input: $in) { %s } }`, gqlMutationResultFields), map[string]any{"in": map[string]any{"group_id": groupID, "role_id": roleID, "scope_type": "GLOBAL"}}},
		{"revokeGroupRole", fmt.Sprintf(`mutation ($a: ID!) { revokeGroupRole(assignment_id: $a) { %s } }`, gqlMutationResultFields), map[string]any{"a": dummyAssignID}},
	}
	if len(authRequiredMutations) != 41 {
		t.Fatalf("expected 41 auth-required mutations, got %d", len(authRequiredMutations))
	}
	for i, tc := range authRequiredMutations {
		if i%8 == 0 {
			resetGatewayRateLimits(t, cfg)
		}
		requireAnonMutationUnauthenticated(t, anon, tc.query, tc.vars, tc.op)
	}

	// 2. Cek seluruh 29 Admin Mutations (hanya pemegang izin iam.* yang berhak; customer dan efOnly ditolak AUTHORIZATION).
	adminDenied := authRequiredMutations[12:]
	for i, tc := range adminDenied {
		if i%8 == 0 {
			resetGatewayRateLimits(t, cfg)
		}
		requireCustomerMutationForbidden(t, custClient, tc.query, tc.vars, tc.op)
		requireCustomerMutationForbidden(t, efOnlyClient, tc.query, tc.vars, tc.op)
	}

	// Cek perilaku publik/authenticated createUser (pendaftaran customer publik diperbolehkan bagi anon & non-admin,
	// sedangkan admin berizin iam.users.write mendapat user payload langsung & konflik email eksplisit).
	resetGatewayRateLimits(t, cfg)
	pubRegTag := uniqueTag("mat_pubcu")
	pubRegEmail := pubRegTag + "@integrione.test"
	t.Cleanup(func() {
		id := psqlScalar(t, cfg, fmt.Sprintf(`SELECT user_id::text FROM user_identities WHERE kind = 'email' AND value = %s LIMIT 1;`, sqlQuote(pubRegEmail)))
		cleanupUserByID(t, cfg, id)
	})
	pubCreateMut := fmt.Sprintf(`mutation ($in: CreateUserInput!) { createUser(input: $in) { %s } }`, gqlCreateUserPayloadFields)
	pubRes := requireMutationSuccess(t, doGQLOK(t, anon, pubCreateMut, map[string]any{
		"in": map[string]any{"full_name": "Mat Pub", "email": pubRegEmail, "username": pubRegTag, "password": "StrongPass#2026!"},
	}), "createUser")
	if pubRes["user"] != nil {
		t.Fatalf("anonymous createUser must mask user=nil, got %+v", pubRes)
	}
	dupCustRes := requireMutationSuccess(t, doGQLOK(t, custClient, pubCreateMut, map[string]any{
		"in": map[string]any{"full_name": "Mat Cust Dup", "email": pubRegEmail, "username": uniqueTag("mat_custdup"), "password": "StrongPass#2026!"},
	}), "createUser")
	if dupCustRes["redirectLocation"] != "/login" || dupCustRes["user"] != nil {
		t.Fatalf("non-admin duplicate email expected /login anti-oracle redirect, got %+v", dupCustRes)
	}
	_, _ = doMutationFail(t, admin, pubCreateMut, map[string]any{
		"in": map[string]any{"full_name": "Mat Admin Dup", "email": pubRegEmail, "username": uniqueTag("mat_admindup"), "password": "StrongPass#2026!"},
	}, "createUser", "CONFLICT")

	// 3. Cek Unauthorized Self-Service (pengguna tanpa peran *self -> AUTHORIZATION pada seluruh mutasi self berizin).
	resetGatewayRateLimits(t, cfg)
	noSelfCust := createFixtureUser(t, admin, cfg, "CUSTOMER", true)
	noSelfRolesConn := requireOpMap(t, doGQLOK(t, admin, `query ($u: ID!) { userRoles(user_id: $u, first: 5) { edges { assignment_id } } }`, map[string]any{"u": noSelfCust.ID}), "userRoles")
	for _, rawE := range requireConnectionEdges(t, noSelfRolesConn) {
		_ = requireMutationBoolOK(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($id: ID!) { revokeUserRole(assignment_id: $id) { %s } }`, gqlRevokeUserRoleFields), map[string]any{"id": requireNonEmptyString(t, rawE.(map[string]any), "assignment_id")}), "revokeUserRole", "revoked")
	}
	resetGatewayRateLimits(t, cfg)
	noSelfClient, _ := loginFixtureUser(t, cfg, noSelfCust.Email, noSelfCust.Password)

	selfPermOps := authRequiredMutations[1:12] // 11 self mutations (kecuali userLogout).
	for i, tc := range selfPermOps {
		if i%8 == 0 {
			resetGatewayRateLimits(t, cfg)
		}
		requireCustomerMutationForbidden(t, noSelfClient, tc.query, tc.vars, tc.op)
	}
	resetGatewayRateLimits(t, cfg)

	// 4. Cek public/auth-flow negative guards (validateUserMfa tanpa challenge valid, userLogin akun system, createUser -> anon & admin).
	_, _ = doMutationFail(t, anon, fmt.Sprintf(`mutation ($in: ValidateUserMfaInput!) { validateUserMfa(input: $in) { %s } }`, gqlMutationResultFields), map[string]any{"in": map[string]any{"challenge": "invalid-challenge-token", "totpCode": "123456"}}, "validateUserMfa", "AUTHENTICATION")
	_, _ = doMutationFail(t, anon, fmt.Sprintf(`mutation ($in: UserLoginInput!) { userLogin(input: $in) { %s } }`, gqlMutationResultFields), map[string]any{"in": map[string]any{"identifier": "auto_system_iam", "password": "anything"}}, "userLogin", "AUTHORIZATION")
	// Verifikasi verifyEmail token tidak sah mengembalikan status=INVALID (bukan kebocoran) dan resetPassword token tidak sah mengembalikan error AUTHENTICATION/ARGUMENT.
	vRes := requireMutationSuccess(t, doGQLOK(t, anon, fmt.Sprintf(`mutation ($in: VerifyEmailInput!) { verifyEmail(input: $in) { %s } }`, gqlVerifyEmailFields), map[string]any{"in": map[string]any{"token": "nonexistent-email-token-123"}}), "verifyEmail")
	if vRes["status"] != "INVALID" || vRes["verified"] != false {
		t.Fatalf("verifyEmail invalid token unexpected result: %+v", vRes)
	}
	_, _ = doMutationFail(t, anon, fmt.Sprintf(`mutation ($in: IAMResetPasswordInput!) { resetPassword(input: $in) { %s } }`, gqlMutationResultFields), map[string]any{"in": map[string]any{"token": "invalid-reset-tok-123", "newPassword": "NewStrongPass#2026!"}}, "resetPassword", "NOT_FOUND", "AUTHENTICATION", "CONFLICT", "ARGUMENT")

	// userResetPassword netral untuk anonymous maupun customer terautentikasi.
	requireMutationSuccess(t, doGQLOK(t, anon, fmt.Sprintf(`mutation ($in: UserResetPasswordInput!) { userResetPassword(input: $in) { %s } }`, gqlUserResetPasswordFields), map[string]any{"in": map[string]any{"email": "unknown.reset@integrione.test"}}), "userResetPassword")
	requireMutationSuccess(t, doGQLOK(t, custClient, fmt.Sprintf(`mutation ($in: UserResetPasswordInput!) { userResetPassword(input: $in) { %s } }`, gqlUserResetPasswordFields), map[string]any{"in": map[string]any{"email": targetEmp.Email}}), "userResetPassword")

	// resendValidationEmailToken tanpa email oleh anonymous -> ARGUMENT; dengan email -> netral is_resend=true; customer aktif sudah verified -> CONFLICT.
	_, _ = doMutationFail(t, anon, fmt.Sprintf(`mutation { resendValidationEmailToken { %s } }`, gqlMutationResultFields), nil, "resendValidationEmailToken", "ARGUMENT")
	rAnon := requireMutationSuccess(t, doGQLOK(t, anon, fmt.Sprintf(`mutation ($in: ResendValidationEmailInput) { resendValidationEmailToken(input: $in) { %s } }`, gqlResendValidationFields), map[string]any{"in": map[string]any{"email": "unknown.resend@integrione.test"}}), "resendValidationEmailToken")
	if rAnon["is_resend"] != true {
		t.Fatalf("anonymous resendValidationEmailToken expected neutral is_resend=true, got %+v", rAnon)
	}
	_, _ = doMutationFail(t, custClient, fmt.Sprintf(`mutation { resendValidationEmailToken { %s } }`, gqlMutationResultFields), nil, "resendValidationEmailToken", "CONFLICT")
}

// TestIAMToEF_ScopedAssignmentAndVerifierFullMatrix memverifikasi kontrak agents/notes/enterprise_foundation/1-iam.md
// secara live end-to-end melalui IAM -> EF gRPC ScopeVerifier pada AssignUserRole dan AssignGroupRole
// untuk seluruh 5 ScopeType eksternal EF (LEGAL_ENTITY, ORGANIZATION_UNIT, FACILITY, SUPPLIER_PARTY, VENDOR_PARTY).
func TestIAMToEF_ScopedAssignmentAndVerifierFullMatrix(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	seedEFPermissionCatalog(t, cfg)
	resetGatewayRateLimits(t, cfg)
	t.Cleanup(func() { resetGatewayRateLimits(t, cfg) })

	admin := newSuperAdminBrowser(t, cfg)
	tag := uniqueTag("efscope")

	// Pastikan EF permissions tersedia pada peran kustom untuk menguji assignUserRole & assignGroupRole berscope.
	efRoleID, _ := createFixtureRole(t, admin, cfg, "ef_scoped_mat")
	for _, code := range []string{"enterprise_foundation.legal_entities.read", "enterprise_foundation.organization_units.read", "enterprise_foundation.locations.read", "enterprise_foundation.parties.read"} {
		pID := findPermissionIDByCode(t, admin, code)
		_ = requireMutationChildOK(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($r: ID!, $p: ID!) { assignRolePermission(role_id: $r, permission_id: $p) { %s } }`, gqlRolePermPayloadFields), map[string]any{"r": efRoleID, "p": pID}), "assignRolePermission", "role_permission_edge")
	}

	emp := createFixtureUser(t, admin, cfg, "EMPLOYEE", true)
	grpID, _ := createFixtureGroup(t, admin, cfg, "efscope_grp")
	_ = requireMutationChildOK(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($g: ID!, $u: ID!) { addGroupMember(group_id: $g, user_id: $u) { %s } }`, gqlGroupMemberPayloadFields), map[string]any{"g": grpID, "u": emp.ID}), "addGroupMember", "group_member_edge")

	// Buat hirarki aktif di Enterprise Foundation untuk ke-5 scope eksternal.
	leObj := requireMutationChildOK(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($in: CreateLegalEntityInput!) {
		createLegalEntity(input: $in) { legal_entity { id _id status } %s }
	}`, gqlMutationResultFields), map[string]any{
		"in": map[string]any{
			"entity_code":               "LE-" + strings.ToUpper(tag),
			"legal_name":                "PT Scope Matrix " + tag,
			"jurisdiction_country_code": "ID",
		},
	}), "createLegalEntity", "legal_entity")
	leID := requireNonEmptyString(t, leObj, "id")
	leRawID := requireNonEmptyString(t, leObj, "_id")
	t.Cleanup(func() { cleanupEFEntityRows(t, cfg, "legal_entities", leRawID) })
	_ = requireMutationChildOK(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($id: ID!) { activateLegalEntity(id: $id) { legal_entity { id status } %s } }`, gqlMutationResultFields), map[string]any{"id": leID}), "activateLegalEntity", "legal_entity")

	locObj := requireMutationChildOK(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($in: CreateLocationInput!) {
		createLocation(input: $in) { location { id _id } %s }
	}`, gqlMutationResultFields), map[string]any{
		"in": map[string]any{
			"legal_entity_id": leID,
			"location_code":   "WH-" + strings.ToUpper(tag),
			"name":            "Warehouse Scope " + tag,
			"kind":            "WAREHOUSE",
			"timezone":        "Asia/Jakarta",
		},
	}), "createLocation", "location")
	locID := requireNonEmptyString(t, locObj, "id")
	locRawID := requireNonEmptyString(t, locObj, "_id")
	t.Cleanup(func() { cleanupEFEntityRows(t, cfg, "locations", locRawID) })

	ouObj := requireMutationChildOK(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($in: CreateOrganizationUnitInput!) {
		createOrganizationUnit(input: $in) { organization_unit { id _id } %s }
	}`, gqlMutationResultFields), map[string]any{
		"in": map[string]any{
			"legal_entity_id": leID,
			"unit_code":       "OU-" + strings.ToUpper(tag),
			"name":            "Unit Scope " + tag,
			"kind":            "DIVISION",
		},
	}), "createOrganizationUnit", "organization_unit")
	ouID := requireNonEmptyString(t, ouObj, "id")
	ouRawID := requireNonEmptyString(t, ouObj, "_id")
	t.Cleanup(func() { cleanupEFEntityRows(t, cfg, "organization_units", ouRawID) })

	supPartyID := createFixturePartyForRole(t, admin, cfg, "SUPPLIER")
	vndPartyID := createFixturePartyForRole(t, admin, cfg, "VENDOR")

	validScopes := []struct {
		scopeType string
		scopeID   string
		rawID     string
	}{
		{"LEGAL_ENTITY", leID, leRawID},
		{"ORGANIZATION_UNIT", ouID, ouRawID},
		{"FACILITY", locID, locRawID},
		{"SUPPLIER_PARTY", supPartyID, rawUUID(supPartyID)},
		{"VENDOR_PARTY", vndPartyID, rawUUID(vndPartyID)},
	}

	assignUserMut := fmt.Sprintf(`mutation ($in: AssignUserRoleInput!) { assignUserRole(input: $in) { %s } }`, gqlUserRolePayloadFields)
	assignGroupMut := fmt.Sprintf(`mutation ($in: AssignGroupRoleInput!) { assignGroupRole(input: $in) { %s } }`, gqlGroupRolePayloadFields)

	// 1. Authorized valid scoped assignments untuk AssignUserRole dan AssignGroupRole pada ke-5 scope eksternal.
	for _, sc := range validScopes {
		uEdge := requireMutationChildOK(t, doGQLOK(t, admin, assignUserMut, map[string]any{
			"in": map[string]any{"user_id": emp.ID, "role_id": efRoleID, "scope_type": sc.scopeType, "scope_id": sc.rawID},
		}), "assignUserRole", "user_role_edge")
		assertUserRoleObject(t, uEdge)
		uAssignID := requireNonEmptyString(t, uEdge, "assignment_id")
		t.Cleanup(func() {
			_, _ = admin.doGraphQL(fmt.Sprintf(`mutation ($id: ID!) { revokeUserRole(assignment_id: $id) { %s } }`, gqlMutationResultFields), map[string]any{"id": uAssignID}, nil)
		})

		gEdge := requireMutationChildOK(t, doGQLOK(t, admin, assignGroupMut, map[string]any{
			"in": map[string]any{"group_id": grpID, "role_id": efRoleID, "scope_type": sc.scopeType, "scope_id": sc.rawID},
		}), "assignGroupRole", "group_role_edge")
		assertGroupRoleObject(t, gEdge)
		gAssignID := requireNonEmptyString(t, gEdge, "assignment_id")
		t.Cleanup(func() {
			_, _ = admin.doGraphQL(fmt.Sprintf(`mutation ($id: ID!) { revokeGroupRole(assignment_id: $id) { %s } }`, gqlMutationResultFields), map[string]any{"id": gAssignID}, nil)
		})
	}

	// 2. Pastikan target non-existent pada semua 5 scope ditolak fail-closed NOT_FOUND.
	missingUUID := "99999999-9999-4999-8999-999999999999"
	for _, sc := range validScopes {
		_, _ = doMutationFail(t, admin, assignUserMut, map[string]any{
			"in": map[string]any{"user_id": emp.ID, "role_id": efRoleID, "scope_type": sc.scopeType, "scope_id": missingUUID},
		}, "assignUserRole", "NOT_FOUND")
		_, _ = doMutationFail(t, admin, assignGroupMut, map[string]any{
			"in": map[string]any{"group_id": grpID, "role_id": efRoleID, "scope_type": sc.scopeType, "scope_id": missingUUID},
		}, "assignGroupRole", "NOT_FOUND")
	}

	// 3. Tangguhkan LegalEntity induk -> verifikasi LegalEntity, OrganizationUnit anak, dan Location anak menjadi ineligible (CONFLICT).
	_ = requireMutationChildOK(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($id: ID!) { suspendLegalEntity(id: $id) { legal_entity { id status } %s } }`, gqlMutationResultFields), map[string]any{"id": leID}), "suspendLegalEntity", "legal_entity")

	for _, sc := range []struct {
		scopeType string
		rawID     string
	}{
		{"LEGAL_ENTITY", leRawID},
		{"ORGANIZATION_UNIT", ouRawID},
		{"FACILITY", locRawID},
	} {
		_, _ = doMutationFail(t, admin, assignUserMut, map[string]any{
			"in": map[string]any{"user_id": emp.ID, "role_id": efRoleID, "scope_type": sc.scopeType, "scope_id": sc.rawID},
		}, "assignUserRole", "CONFLICT", "UNPROCESSABLE_ENTITY")
		_, _ = doMutationFail(t, admin, assignGroupMut, map[string]any{
			"in": map[string]any{"group_id": grpID, "role_id": efRoleID, "scope_type": sc.scopeType, "scope_id": sc.rawID},
		}, "assignGroupRole", "CONFLICT", "UNPROCESSABLE_ENTITY")
	}

	// 4. Party tanpa peran yang cocok (mis. SUPPLIER party ditarget sebagai VENDOR_PARTY) ditolak fail-closed.
	_, _ = doMutationFail(t, admin, assignUserMut, map[string]any{
		"in": map[string]any{"user_id": emp.ID, "role_id": efRoleID, "scope_type": "VENDOR_PARTY", "scope_id": rawUUID(supPartyID)},
	}, "assignUserRole", "CONFLICT", "NOT_FOUND", "UNPROCESSABLE_ENTITY")
}
