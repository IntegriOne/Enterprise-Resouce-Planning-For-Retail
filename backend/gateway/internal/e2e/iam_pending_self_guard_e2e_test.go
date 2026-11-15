//go:build e2e

package e2e

import (
	"fmt"
	"strings"
	"testing"
)

// TestPendingAndUnverifiedAutoAssignedSelfServiceGuardE2E memverifikasi secara langsung lewat HTTPS + Caddy + Gateway + IAM gRPC
// bahwa akun baru dari ke-5 alur pembuatan user (createUser publik & admin, createEmployeeUser, createVendorUser, createSupplierUser)
// langsung menerima role dasar self-service, namun login dan seluruh layanan *self tetap diblokir selama status masih PENDING
// atau identitas belum terverifikasi, lalu langsung terbuka setelah verifikasi email & aktivasi selesai.
func TestPendingAndUnverifiedAutoAssignedSelfServiceGuardE2E(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)

	mUserRoles := fmt.Sprintf(`query ($u: ID!) {
		userRoles(user_id: $u, first: 10) {
			edges { %s }
			total_count
		}
		userEffectiveRoles(user_id: $u, first: 10) {
			nodes { %s }
			total_count
		}
	}`, gqlUserRoleEdgeFields, gqlRoleFields)
	mLogin := fmt.Sprintf(`mutation ($input: UserLoginInput!) { userLogin(input: $input) { %s } }`, gqlUserLoginFields)
	mVerify := fmt.Sprintf(`mutation ($input: VerifyEmailInput!) { verifyEmail(input: $input) { %s } }`, gqlVerifyEmailFields)
	mResend := fmt.Sprintf(`mutation ($input: ResendValidationEmailInput) { resendValidationEmailToken(input: $input) { %s } }`, gqlResendValidationFields)
	mActivate := fmt.Sprintf(`mutation ($u: ID!) { activateUser(user_id: $u) { %s } }`, gqlUserPayloadFields)

	mUpdProfile := fmt.Sprintf(`mutation ($input: UpdateMyProfileInput!) { updateMyProfile(input: $input) { %s } }`, gqlUserPayloadFields)
	mUpdEmail := fmt.Sprintf(`mutation ($input: UserEmailInput!) { updateUserEmail(input: $input) { %s } }`, gqlUserEmailPayloadFields)
	mChgUsername := fmt.Sprintf(`mutation ($input: UsernameInput!) { changeUsername(input: $input) { %s } }`, gqlUserNamePayloadFields)
	mDelPending := fmt.Sprintf(`mutation { deletePendingUserEmail { %s } }`, gqlDeletePendingEmailFields)
	mChgPassword := fmt.Sprintf(`mutation ($input: ChangeUserPasswordInput!) { changeUserPassword(input: $input) { %s } }`, gqlChangeUserPasswordFields)
	mEnrollTOTP := fmt.Sprintf(`mutation { enrollMyTotp { %s } }`, gqlEnrollTotpFields)
	mDisableTOTP := fmt.Sprintf(`mutation { disableMyTotp { %s } }`, gqlDisableTotpFields)
	mRegenCodes := fmt.Sprintf(`mutation { regenerateMyRecoveryCodes { %s } }`, gqlRegenRecoveryCodesFields)

	qMe := fmt.Sprintf(`query { me { %s } }`, gqlUserFields)
	qMySess := fmt.Sprintf(`query { mySessions(active_only: true, first: 5) { nodes { %s } total_count } }`, gqlSessionFields)
	qSnap := fmt.Sprintf(`query { myAccessSnapshot { %s } }`, gqlAccessSnapshotFields)
	qWorkspace := fmt.Sprintf(`query { myWorkspaceAccess { %s } }`, gqlWorkspaceAccessFields)
	qBrowserSess := fmt.Sprintf(`query { browserSession { %s } }`, gqlBrowserSessionFields)
	qUserByID := fmt.Sprintf(`query ($id: ID!) { user(id: $id) { %s } }`, gqlUserFields)

	assertBaselineSelfRoleAssigned := func(t *testing.T, admin *browserClient, userGID, rawID, wantRole string) {
		t.Helper()
		roleData := doGQLOK(t, admin, mUserRoles, map[string]any{"u": userGID})
		urConn := requireOpMap(t, roleData, "userRoles")
		if got := requireNumberField(t, urConn, "total_count"); got != 1 {
			t.Fatalf("userRoles total_count = %d, want 1", got)
		}
		edges := requireConnectionEdges(t, urConn)
		if len(edges) != 1 {
			t.Fatalf("userRoles edges len = %d, want 1", len(edges))
		}
		edge, _ := edges[0].(map[string]any)
		scopeObj := requireChildMap(t, edge, "scope")
		if scopeObj["type"] != "SELF" || scopeObj["id"] != rawID {
			t.Fatalf("baseline role scope = %+v, want SELF(%s)", scopeObj, rawID)
		}
		nodeRole := requireChildMap(t, edge, "node")
		if nodeRole["code"] != wantRole {
			t.Fatalf("baseline userRoles code = %v, want %q", nodeRole["code"], wantRole)
		}

		effConn := requireOpMap(t, roleData, "userEffectiveRoles")
		effNodes := requireConnectionNodes(t, effConn)
		if len(effNodes) != 1 {
			t.Fatalf("userEffectiveRoles len = %d, want 1", len(effNodes))
		}
		effRole, _ := effNodes[0].(map[string]any)
		if effRole["code"] != wantRole {
			t.Fatalf("baseline userEffectiveRoles code = %v, want %q", effRole["code"], wantRole)
		}
	}

	assertLoginAndSelfBlocked := func(t *testing.T, email, username, password string) {
		t.Helper()
		anon := newAnonBrowser(t, cfg)

		// Upaya login baik memakai email maupun username wajib ditolak AUTHENTICATION.
		for _, identifier := range []string{email, username} {
			_, _ = doMutationFail(t, anon, mLogin, map[string]any{
				"input": map[string]any{
					"identifier": identifier,
					"password":   password,
				},
			}, "userLogin", "AUTHENTICATION")
		}

		// Browser tetap berstatus pre-auth non-authenticated.
		bData := doGQLOK(t, anon, qBrowserSess, nil)
		bSess := requireOpMap(t, bData, "browserSession")
		assertBrowserSessionPayload(t, bSess, false)

		// Seluruh query *self wajib ditolak UNAUTHENTICATED.
		for _, q := range []string{qMe, qMySess, qSnap, qWorkspace} {
			requireAnonUnauthenticated(t, anon, q, nil)
		}

		// Seluruh mutation *self wajib ditolak AUTHENTICATION.
		mutCases := []struct {
			op    string
			query string
			vars  map[string]any
		}{
			{op: "updateMyProfile", query: mUpdProfile, vars: map[string]any{"input": map[string]any{"display_name": "Blocked"}}},
			{op: "updateUserEmail", query: mUpdEmail, vars: map[string]any{"input": map[string]any{"email": "new." + email, "current_password": password}}},
			{op: "changeUsername", query: mChgUsername, vars: map[string]any{"input": map[string]any{"username": username + "_blocked"}}},
			{op: "deletePendingUserEmail", query: mDelPending, vars: nil},
			{op: "changeUserPassword", query: mChgPassword, vars: map[string]any{"input": map[string]any{"current_password": password, "new_password": "NewStrongPass#2026!"}}},
			{op: "enrollMyTotp", query: mEnrollTOTP, vars: nil},
			{op: "disableMyTotp", query: mDisableTOTP, vars: nil},
			{op: "regenerateMyRecoveryCodes", query: mRegenCodes, vars: nil},
		}
		for _, mc := range mutCases {
			requireAnonMutationUnauthenticated(t, anon, mc.query, mc.vars, mc.op)
		}
	}

	verifyUserViaRealTokenMutation := func(t *testing.T, rawID string) {
		t.Helper()
		anon := newAnonBrowser(t, cfg)
		rawToken := uniqueTag("e2e_ver_tok")
		psqlExec(t, cfg, fmt.Sprintf(
			`UPDATE user_tokens SET token_digest = decode(%s, 'hex') WHERE user_id = %s::uuid AND purpose = 'email_verification' AND status = 'active';`,
			sqlQuote(sha256Hex(rawToken)), sqlQuote(rawID),
		))
		verData := doGQLOK(t, anon, mVerify, map[string]any{
			"input": map[string]any{"token": rawToken},
		})
		verRes := requireOpMap(t, verData, "verifyEmail")
		assertVerifyEmailPayload(t, verRes, "VERIFIED", true)
	}

	t.Run("All5CreationVariants_PendingBlockedThenVerifyEmailUnlocks", func(t *testing.T) {
		variants := []struct {
			name     string
			opName   string
			inType   string
			asPublic bool
			wantKind string
			wantRole string
		}{
			{name: "PublicCreateUser", opName: "createUser", inType: "CreateUserInput", asPublic: true, wantKind: "CUSTOMER", wantRole: "customer"},
			{name: "AdminCreateUser", opName: "createUser", inType: "CreateUserInput", asPublic: false, wantKind: "CUSTOMER", wantRole: "customer"},
			{name: "AdminCreateEmployeeUser", opName: "createEmployeeUser", inType: "CreateEmployeeUserInput", asPublic: false, wantKind: "EMPLOYEE", wantRole: "iam_selfservice"},
			{name: "AdminCreateVendorUser", opName: "createVendorUser", inType: "CreateVendorUserInput", asPublic: false, wantKind: "VENDOR", wantRole: "iam_selfservice"},
			{name: "AdminCreateSupplierUser", opName: "createSupplierUser", inType: "CreateSupplierUserInput", asPublic: false, wantKind: "SUPPLIER", wantRole: "iam_selfservice"},
		}

		for _, vc := range variants {
			t.Run(vc.name, func(t *testing.T) {
				admin := newSuperAdminBrowser(t, cfg)
				tag := uniqueTag("psg_" + strings.ToLower(vc.wantKind))
				email := tag + "@integrione.test"
				username := tag
				password := "StrongPass#2026!"

				caller := admin
				payloadFields := gqlUserPayloadFields
				if vc.opName == "createUser" {
					payloadFields = gqlCreateUserPayloadFields
				}
				if vc.asPublic {
					caller = newAnonBrowser(t, cfg)
				}

				createMutation := fmt.Sprintf(`mutation ($input: %s!) { %s(input: $input) { %s } }`, vc.inType, vc.opName, payloadFields)
				data := doGQLOK(t, caller, createMutation, map[string]any{
					"input": map[string]any{
						"full_name": "Pending Guard " + tag,
						"email":     email,
						"username":  username,
						"password":  password,
					},
				})
				var userGID, rawID string
				if vc.asPublic {
					payload := requireMutationSuccess(t, data, vc.opName)
					if payload["user"] != nil {
						t.Fatalf("public %s expected masked user=null, got %+v", vc.opName, payload["user"])
					}
					rawID = psqlScalar(t, cfg, fmt.Sprintf(`SELECT user_id::text FROM user_identities WHERE kind = 'email' AND value = %s LIMIT 1;`, sqlQuote(email)))
					if rawID == "" {
						t.Fatalf("public %s did not persist user_identities row for %s", vc.opName, email)
					}
					userGID = "gid://integrione/user/" + rawID
					fetched := requireOpMap(t, doGQLOK(t, admin, qUserByID, map[string]any{"id": userGID}), "user")
					assertUserObject(t, fetched)
					if fetched["kind"] != vc.wantKind || fetched["status"] != "PENDING" {
						t.Fatalf("public %s persisted kind=%v status=%v, want kind=%s status=PENDING", vc.opName, fetched["kind"], fetched["status"], vc.wantKind)
					}
				} else {
					u := requireMutationChildOK(t, data, vc.opName, "user")
					assertUserObject(t, u)
					if u["kind"] != vc.wantKind || u["status"] != "PENDING" {
						t.Fatalf("%s returned kind=%v status=%v, want kind=%s status=PENDING", vc.opName, u["kind"], u["status"], vc.wantKind)
					}
					userGID = requireNonEmptyString(t, u, "id")
					rawID = requireNonEmptyString(t, u, "_id")
				}
				t.Cleanup(func() {
					cleanupUserByID(t, cfg, rawID)
				})

				// Pastikan token verifikasi email aktif sudah terbit sejak pembuatan akun PENDING.
				if tokCnt := psqlScalar(t, cfg, fmt.Sprintf(`SELECT count(*) FROM user_tokens WHERE user_id = %s::uuid AND purpose = 'email_verification' AND status = 'active';`, sqlQuote(rawID))); tokCnt != "1" {
					t.Fatalf("%s active email_verification token count = %s, want 1", vc.name, tokCnt)
				}

				// 1. Role self-service dasar sudah terpasang otomatis dalam lingkup SELF milik user.
				assertBaselineSelfRoleAssigned(t, admin, userGID, rawID, vc.wantRole)

				// 2. Selama masih PENDING & belum terverifikasi, login dan seluruh *self wajib ditolak.
				assertLoginAndSelfBlocked(t, email, username, password)

				// Untuk varian registrasi publik, pastikan alur resend token pending (Kasus 2) juga tetap menjaga 1 token aktif.
				if vc.asPublic {
					resendAnon := newAnonBrowser(t, cfg)
					resendRes := requireMutationBoolOK(t, doGQLOK(t, resendAnon, mResend, map[string]any{
						"input": map[string]any{"email": email},
					}), "resendValidationEmailToken", "is_resend")
					_ = resendRes
					if tokCnt := psqlScalar(t, cfg, fmt.Sprintf(`SELECT count(*) FROM user_tokens WHERE user_id = %s::uuid AND purpose = 'email_verification' AND status = 'active';`, sqlQuote(rawID))); tokCnt != "1" {
						t.Fatalf("after resend active email_verification token count = %s, want 1", tokCnt)
					}
				}

				// 3. Jalankan verifikasi email nyata -> identitas terverifikasi & status otomatis aktif (ACTIVE).
				verifyUserViaRealTokenMutation(t, rawID)
				checkedUser := requireOpMap(t, doGQLOK(t, admin, qUserByID, map[string]any{"id": userGID}), "user")
				if checkedUser["status"] != "ACTIVE" {
					t.Fatalf("after verifyEmail user.status = %v, want ACTIVE", checkedUser["status"])
				}

				// 4. Kini login email & username berhasil, serta 5 izin *self pada snapshot langsung aktif digunakan.
				resetGatewayRateLimits(t, cfg)
				emailClient, loginEmailPayload := loginFixtureUser(t, cfg, email, password)
				assertUserLoginPayload(t, loginEmailPayload, "AUTHENTICATED")

				snap := requireOpMap(t, doGQLOK(t, emailClient, qSnap, nil), "myAccessSnapshot")
				assertAccessSnapshotObject(t, snap)
				if snap["status"] != "active" {
					t.Fatalf("myAccessSnapshot.status = %v, want active", snap["status"])
				}
				seenPerms := map[string]bool{}
				for _, rawP := range requireChildSlice(t, snap, "permissions") {
					pm, _ := rawP.(map[string]any)
					sc := requireChildMap(t, pm, "scope")
					if sc["type"] == "SELF" && sc["id"] == rawID {
						if code, ok := pm["permission_code"].(string); ok {
							seenPerms[code] = true
						}
					}
				}
				for _, wantPerm := range []string{
					"iam.users.readself",
					"iam.users.updateself",
					"iam.users.deleteself",
					"iam.sessions.readself",
					"iam.sessions.deleteself",
				} {
					if !seenPerms[wantPerm] {
						t.Fatalf("%s missing SELF-scoped permission %q after verifyEmail; seen=%+v", vc.name, wantPerm, seenPerms)
					}
				}

				meObj := requireOpMap(t, doGQLOK(t, emailClient, qMe, nil), "me")
				assertUserObject(t, meObj)
				sessConn := requireOpMap(t, doGQLOK(t, emailClient, qMySess, nil), "mySessions")
				if got := requireNumberField(t, sessConn, "total_count"); got < 1 {
					t.Fatalf("%s mySessions total_count = %d, want >= 1", vc.name, got)
				}
				updUser := requireMutationChildOK(t, doGQLOK(t, emailClient, mUpdProfile, map[string]any{
					"input": map[string]any{"display_name": "Verified Self " + tag},
				}), "updateMyProfile", "user")
				if updUser["display_name"] != "Verified Self "+tag {
					t.Fatalf("updateMyProfile display_name = %v, want %q", updUser["display_name"], "Verified Self "+tag)
				}

				// Verifikasi username juga ikut terverifikasi sehingga login memakai username berhasil.
				requireMutationBoolOK(t, doGQLOK(t, emailClient, fmt.Sprintf(`mutation { userLogout { %s } }`, gqlUserLogoutFields), nil), "userLogout", "logged_out")
				unameClient, loginUnamePayload := loginFixtureUser(t, cfg, username, password)
				assertUserLoginPayload(t, loginUnamePayload, "AUTHENTICATED")
				_ = unameClient
			})
		}
	})

	t.Run("ActiveWithoutVerifiedIdentityStillBlockedFromLoginAndSelf", func(t *testing.T) {
		for _, tc := range []struct {
			name     string
			opName   string
			inType   string
			wantRole string
		}{
			{name: "CustomerActivatedBeforeVerify", opName: "createUser", inType: "CreateUserInput", wantRole: "customer"},
			{name: "EmployeeActivatedBeforeVerify", opName: "createEmployeeUser", inType: "CreateEmployeeUserInput", wantRole: "iam_selfservice"},
		} {
			t.Run(tc.name, func(t *testing.T) {
				resetGatewayRateLimits(t, cfg)
				admin := newSuperAdminBrowser(t, cfg)
				tag := uniqueTag("act_unver")
				email := tag + "@integrione.test"
				username := tag
				password := "StrongPass#2026!"

				payloadFields := gqlUserPayloadFields
				if tc.opName == "createUser" {
					payloadFields = gqlCreateUserPayloadFields
				}
				data := doGQLOK(t, admin, fmt.Sprintf(`mutation ($input: %s!) { %s(input: $input) { %s } }`, tc.inType, tc.opName, payloadFields), map[string]any{
					"input": map[string]any{
						"full_name": "Unverified Active " + tag,
						"email":     email,
						"username":  username,
						"password":  password,
					},
				})
				u := requireMutationChildOK(t, data, tc.opName, "user")
				userGID := requireNonEmptyString(t, u, "id")
				rawID := requireNonEmptyString(t, u, "_id")
				t.Cleanup(func() {
					cleanupUserByID(t, cfg, rawID)
				})

				assertBaselineSelfRoleAssigned(t, admin, userGID, rawID, tc.wantRole)

				// Admin mengaktifkan akun ke ACTIVE sementara email & username masih is_verified=false.
				actUser := requireMutationChildOK(t, doGQLOK(t, admin, mActivate, map[string]any{"u": userGID}), "activateUser", "user")
				if actUser["status"] != "ACTIVE" {
					t.Fatalf("activateUser status = %v, want ACTIVE", actUser["status"])
				}
				if verCnt := psqlScalar(t, cfg, fmt.Sprintf(`SELECT count(*) FROM user_identities WHERE user_id = %s::uuid AND is_verified = true;`, sqlQuote(rawID))); verCnt != "0" {
					t.Fatalf("verified identity count before verifyEmail = %s, want 0", verCnt)
				}

				// Walau status ACTIVE & role *self sudah ada, login & *self tetap ditolak karena identitas belum terverifikasi.
				assertLoginAndSelfBlocked(t, email, username, password)

				// Baru setelah verifyEmail dijalankan, login email/username dan akses *self langsung bekerja.
				verifyUserViaRealTokenMutation(t, rawID)
				client, payload := loginFixtureUser(t, cfg, email, password)
				assertUserLoginPayload(t, payload, "AUTHENTICATED")
				_ = requireOpMap(t, doGQLOK(t, client, qMe, nil), "me")
			})
		}
	})
}
