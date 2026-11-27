//go:build e2e

// Package e2e berisi pengujian end-to-end HTTPS nyata terhadap seluruh 22 Query dan 45 Mutation IAM pada Gateway GraphQL IntegriOne.
package e2e

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os/exec"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/shared/authsession"
)

const (
	gqlMutationResultFields     = `errors { type message } was_successful`
	gqlBrowserSessionFields     = `authenticated principalId csrfToken`
	gqlWorkspaceAccessFields    = `canAccessIdentity defaultLocation`
	gqlUserLoginFields          = `status mfaChallenge mfaExpiresAt csrfToken redirectLocation ` + gqlMutationResultFields
	gqlValidateUserMfaFields    = `csrfToken redirectLocation ` + gqlMutationResultFields
	gqlCreateUserPayloadFields  = `user { ` + gqlUserFields + ` } redirectLocation ` + gqlMutationResultFields
	gqlVerifyEmailFields        = `status verified ` + gqlMutationResultFields
	gqlUserResetPasswordFields  = `message ` + gqlMutationResultFields
	gqlResetPasswordFields      = `reset ` + gqlMutationResultFields
	gqlUserLogoutFields         = `logged_out ` + gqlMutationResultFields
	gqlUserNamePayloadFields    = `username ` + gqlMutationResultFields
	gqlChangeUserPasswordFields = `changed sessions_revoked password_changed_at ` + gqlMutationResultFields
	gqlUserAvatarPayloadFields  = `avatar_url ` + gqlMutationResultFields
	gqlUserEmailPayloadFields   = `pending_email is_pending expires_at ` + gqlMutationResultFields
	gqlDeletePendingEmailFields = `deleted ` + gqlMutationResultFields
	gqlResendValidationFields   = `is_resend ` + gqlMutationResultFields
	gqlVerifySessionMfaFields   = `verified ` + gqlMutationResultFields
	gqlEnrollTotpFields         = `factor_id otp_auth_uri ` + gqlMutationResultFields
	gqlConfirmTotpFields        = `enabled recovery_codes ` + gqlMutationResultFields
	gqlDisableTotpFields        = `disabled ` + gqlMutationResultFields
	gqlRegenRecoveryCodesFields = `recovery_codes ` + gqlMutationResultFields
	gqlUserFields               = `id _id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision`
	gqlUserPayloadFields        = `user { ` + gqlUserFields + ` } ` + gqlMutationResultFields
	gqlSoftDeleteUserFields     = `deleted ` + gqlMutationResultFields
	gqlIdentityFields           = `id _id user_id kind value is_verified is_primary verified_at created_at updated_at`
	gqlUserIdentityPayload      = `user_identity { ` + gqlIdentityFields + ` } ` + gqlMutationResultFields
	gqlRemoveIdentityFields     = `removed ` + gqlMutationResultFields
	gqlRoleFields               = `id _id code name description management_domain default_redirect_path redirect_priority is_system is_active created_at updated_at`
	gqlRolePayloadFields        = `role { ` + gqlRoleFields + ` } ` + gqlMutationResultFields
	gqlDeleteRoleFields         = `deleted ` + gqlMutationResultFields
	gqlPermissionFields         = `id _id code resource action description is_sensitive created_at`
	gqlRolePermissionEdgeFields = `cursor role_id created_at node { ` + gqlPermissionFields + ` }`
	gqlRolePermPayloadFields    = `role_permission_edge { ` + gqlRolePermissionEdgeFields + ` } ` + gqlMutationResultFields
	gqlRevokeRolePermFields     = `revoked ` + gqlMutationResultFields
	gqlUserRoleEdgeFields       = `cursor assignment_id user_id scope { type id } granted_by expires_at created_at node { ` + gqlRoleFields + ` }`
	gqlUserRolePayloadFields    = `user_role_edge { ` + gqlUserRoleEdgeFields + ` } ` + gqlMutationResultFields
	gqlRevokeUserRoleFields     = `revoked ` + gqlMutationResultFields
	gqlGroupFields              = `id _id code name description is_active created_at updated_at`
	gqlGroupPayloadFields       = `group { ` + gqlGroupFields + ` } ` + gqlMutationResultFields
	gqlDeleteGroupFields        = `deleted ` + gqlMutationResultFields
	gqlGroupMemberEdgeFields    = `cursor group_id created_at node { ` + gqlUserFields + ` }`
	gqlGroupMemberPayloadFields = `group_member_edge { ` + gqlGroupMemberEdgeFields + ` } ` + gqlMutationResultFields
	gqlRemoveGroupMemberFields  = `removed ` + gqlMutationResultFields
	gqlGroupRoleEdgeFields      = `cursor assignment_id group_id scope { type id } granted_by expires_at created_at node { ` + gqlRoleFields + ` }`
	gqlGroupRolePayloadFields   = `group_role_edge { ` + gqlGroupRoleEdgeFields + ` } ` + gqlMutationResultFields
	gqlRevokeGroupRoleFields    = `revoked ` + gqlMutationResultFields
	gqlSessionFields            = `id _id user_id mfa_verified mfa_verified_at ip_address user_agent authenticated_at last_active_at idle_expires_at absolute_expires_at ended_at end_reason created_at generation`
	gqlRevokeSessionFields      = `revoked ` + gqlMutationResultFields
	gqlRevokeAllSessionsFields  = `revoked_count ` + gqlMutationResultFields
	gqlPageInfoFields           = `startCursor endCursor hasPreviousPage hasNextPage`
	gqlAccessSnapshotFields     = `principal_id user_kind status mfa_enabled auth_epoch perm_revision recovery_epoch can_access_identity default_redirect_path allowed_redirect_paths roles { id code name priority default_redirect_path } permissions { permission_code is_sensitive scope { type id } } synced_at_unix fresh_until_unix`
)

var e2eOpSeq atomic.Uint64

type fixtureUserAccount struct {
	ID       string
	RawID    string
	Kind     string
	Email    string
	Username string
	Password string
}

// rawUUID mengekstrak UUID mentah dari Global ID (`gid://integrione/<resource>/<uuid>`) untuk SQL dan kunci Redis.
func rawUUID(id string) string {
	trimmed := strings.TrimSpace(id)
	if idx := strings.LastIndex(trimmed, "/"); idx >= 0 && strings.HasPrefix(trimmed, "gid://") {
		return trimmed[idx+1:]
	}
	return trimmed
}

// requireConnectionNodes mengambil irisan nodes dari sebuah objek Connection Relay.
func requireConnectionNodes(t *testing.T, conn map[string]any) []any {
	t.Helper()
	return requireChildSlice(t, conn, "nodes")
}

// requireConnectionEdges mengambil irisan edges dari sebuah objek Connection Relay.
func requireConnectionEdges(t *testing.T, conn map[string]any) []any {
	t.Helper()
	return requireChildSlice(t, conn, "edges")
}

// uniqueTag menghasilkan pengidentifikasi unik lowercase untuk isolasi data antar pengujian.
// Prefiks "z" pada nomor urut mencegah token numerik murni (mis. "202") bentrok dengan tahun pada kata sandi fixture.
func uniqueTag(prefix string) string {
	seq := e2eOpSeq.Add(1)
	return fmt.Sprintf("%s_%d_z%d", strings.ToLower(strings.TrimSpace(prefix)), time.Now().UnixNano(), seq)
}

// sha256Hex menghitung digest SHA-256 hex lowercase untuk verifikasi atau injeksi token uji di database.
func sha256Hex(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// sqlQuote mengamankan string literal SQL tunggal untuk utilitas fixture pengujian lokal.
func sqlQuote(v string) string {
	return "'" + strings.ReplaceAll(v, "'", "''") + "'"
}

// psqlExec mengeksekusi perintah SQL pembersihan/fixture pada container PostgreSQL IAM.
func psqlExec(t *testing.T, cfg e2eRuntimeConfig, sqlStmt string) {
	t.Helper()
	cmd := exec.Command(
		"docker", "exec", cfg.postgresService,
		"psql", "-v", "ON_ERROR_STOP=1", "-U", "postgres", "-d", "iam", "-tAc", sqlStmt,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("psql exec failed: %v\nSQL: %s\nOutput: %s", err, sqlStmt, strings.TrimSpace(string(out)))
	}
}

// resetGatewayRateLimits membersihkan kunci pembatas laju Gateway di Redis agar ke-67 test berjalan independen tanpa kebocoran jendela 60 detik.
func resetGatewayRateLimits(t *testing.T, cfg e2eRuntimeConfig) {
	t.Helper()
	client := newE2ERedisClient(t, cfg, "gw_runtime", cfg.gwRedisPass)
	defer func() { _ = client.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var cursor uint64
	for {
		keys, nextCursor, err := client.Scan(ctx, cursor, "ig:development:gw:rl:*", 200).Result()
		if err != nil {
			t.Fatalf("scan gateway rate-limit keys: %v", err)
		}
		if len(keys) > 0 {
			if err := client.Del(ctx, keys...).Err(); err != nil {
				t.Fatalf("delete gateway rate-limit keys: %v", err)
			}
		}
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
}

// newAnonBrowser menyiapkan browser HTTPS baru dengan cookie pre-auth dan token CSRF dari Caddy -> Gateway.
func newAnonBrowser(t *testing.T, cfg e2eRuntimeConfig) *browserClient {
	t.Helper()
	client := newHTTPSBrowserClient(t, cfg)
	client.bootstrapSession()
	return client
}

// newSuperAdminBrowser melakukan reset rate limit dan login sebagai human superadmin serta logout otomatis saat cleanup.
func newSuperAdminBrowser(t *testing.T, cfg e2eRuntimeConfig) *browserClient {
	t.Helper()
	resetGatewayRateLimits(t, cfg)
	client := newAnonBrowser(t, cfg)
	status, _, redirectLoc := client.login(cfg.superadminUsername, cfg.superadminPassword)
	if status != "AUTHENTICATED" {
		t.Fatalf("superadmin login status = %q, want AUTHENTICATED", status)
	}
	if redirectLoc != "/identity" {
		t.Fatalf("superadmin redirectLocation = %q, want /identity", redirectLoc)
	}
	t.Cleanup(func() {
		_, _, _ = logoutBrowserBestEffort(client)
	})
	return client
}

// logoutBrowserBestEffort menutup sesi browser saat cleanup agar tidak menyisakan sesi terbuka di database.
func logoutBrowserBestEffort(c *browserClient) (int, graphQLResponse, error) {
	if c == nil || strings.TrimSpace(c.csrfToken) == "" {
		return 0, graphQLResponse{}, nil
	}
	resp, gql := c.doGraphQL(fmt.Sprintf(`mutation { userLogout { %s } }`, gqlUserLogoutFields), nil, nil)
	return resp.StatusCode, gql, nil
}

// compactGQLQuery menormalkan spasi berulang query/mutation GraphQL agar log request tetap ringkas dan terbaca.
func compactGQLQuery(query string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(query)), " ")
}

// formatVarsForLog menyerialisasi variables input GraphQL ke JSON rapi.
func formatVarsForLog(vars map[string]any) string {
	if len(vars) == 0 {
		return "{}"
	}
	b, err := json.MarshalIndent(vars, "", "  ")
	if err != nil {
		return fmt.Sprintf("%+v", vars)
	}
	return string(b)
}

// formatGQLResponseForLog menyerialisasi payload response JSON (data dan/atau errors) secara utuh dan rapi.
func formatGQLResponseForLog(gql graphQLResponse) string {
	out := make(map[string]any)
	if len(gql.Data) > 0 && string(gql.Data) != "null" {
		var parsed any
		if err := json.Unmarshal(gql.Data, &parsed); err == nil {
			out["data"] = parsed
		} else {
			out["data"] = string(gql.Data)
		}
	} else {
		out["data"] = nil
	}
	if len(gql.Errors) > 0 {
		out["errors"] = gql.Errors
	}
	b, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return fmt.Sprintf("data=%s errors=%+v", string(gql.Data), gql.Errors)
	}
	return string(b)
}

// logGraphQLExchange mencetak detail penuh Request (Query/Mutation & Variables) serta Response nyata (HTTP status & Result JSON) pada mode go test -v.
func logGraphQLExchange(t *testing.T, label, query string, vars map[string]any, resp *http.Response, gql graphQLResponse) {
	t.Helper()
	statusCode := 0
	if resp != nil {
		statusCode = resp.StatusCode
	}
	t.Logf("\n[%s] POST /api/graphql (HTTP %d)\nQuery/Mutation:\n  %s\nInput Variables:\n%s\nResult JSON:\n%s\n",
		label,
		statusCode,
		compactGQLQuery(query),
		formatVarsForLog(vars),
		formatGQLResponseForLog(gql),
	)
}

// doGQLOK mengirim query/mutation GraphQL lewat HTTPS nyata, mencatat input/result ke log uji, dan memastikan HTTP 200 tanpa errors.
func doGQLOK(t *testing.T, c *browserClient, query string, vars map[string]any) map[string]any {
	t.Helper()
	resp, gql := c.doGraphQL(query, vars, nil)
	logGraphQLExchange(t, "GQL_OK", query, vars, resp, gql)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("HTTP status = %d, want 200; errors=%+v", resp.StatusCode, gql.Errors)
	}
	if len(gql.Errors) > 0 {
		t.Fatalf("unexpected GraphQL errors: %+v", gql.Errors)
	}
	var root map[string]any
	if err := json.Unmarshal(gql.Data, &root); err != nil {
		t.Fatalf("unmarshal GraphQL data %q: %v", string(gql.Data), err)
	}
	return root
}

// doGQLErrorLabeled mengirim query/mutation kasus negatif, mencatat request/response penuh, dan memverifikasi error GraphQL yang diharapkan.
func doGQLErrorLabeled(t *testing.T, label string, c *browserClient, query string, vars map[string]any, wantCode string) graphQLResponse {
	t.Helper()
	resp, gql := c.doGraphQL(query, vars, nil)
	logGraphQLExchange(t, label, query, vars, resp, gql)
	if wantCode != "" {
		requireErrorCode(t, gql, wantCode)
	} else if len(gql.Errors) == 0 {
		t.Fatalf("expected non-empty GraphQL errors, got none; data=%s", string(gql.Data))
	}
	return gql
}

// doGQLError mengirim query/mutation kasus negatif standar dengan label GQL_EXPECTED_ERROR.
func doGQLError(t *testing.T, c *browserClient, query string, vars map[string]any, wantCode string) graphQLResponse {
	t.Helper()
	return doGQLErrorLabeled(t, "GQL_EXPECTED_ERROR", c, query, vars, wantCode)
}

// doMutationFailLabeled mengirim mutation kasus negatif dan memverifikasi payload MutationResult mengembalikan was_successful=false serta tipe error yang diharapkan.
func doMutationFailLabeled(t *testing.T, label string, c *browserClient, query string, vars map[string]any, opName string, wantTypes ...string) (string, string) {
	t.Helper()
	resp, gql := c.doGraphQL(query, vars, nil)
	logGraphQLExchange(t, label, query, vars, resp, gql)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("%s HTTP status = %d, want 200; errors=%+v", opName, resp.StatusCode, gql.Errors)
	}
	return requireMutationPayloadError(t, gql, opName, wantTypes...)
}

// doMutationFail mengirim mutation kasus negatif standar dan memverifikasi ErrorTypeEnum pada payload MutationError.
func doMutationFail(t *testing.T, c *browserClient, query string, vars map[string]any, opName string, wantTypes ...string) (string, string) {
	t.Helper()
	return doMutationFailLabeled(t, "MUTATION_EXPECTED_ERROR", c, query, vars, opName, wantTypes...)
}

// assertMutationResultSuccess memastikan objek payload MutationResult menandakan keberhasilan (was_successful=true dan errors=[]).
func assertMutationResultSuccess(t *testing.T, payload map[string]any, op string) {
	t.Helper()
	wasOK := requireBoolField(t, payload, "was_successful")
	errs := requireChildSlice(t, payload, "errors")
	if !wasOK {
		t.Fatalf("%s.was_successful = false, want true (errors=%+v)", op, errs)
	}
	if len(errs) != 0 {
		t.Fatalf("%s.errors = %+v, want empty array on success", op, errs)
	}
}

// requireMutationSuccess mengambil objek payload mutation dan memverifikasi kontrak sukses MutationResult.
func requireMutationSuccess(t *testing.T, root map[string]any, op string) map[string]any {
	t.Helper()
	payload := requireOpMap(t, root, op)
	assertMutationResultSuccess(t, payload, op)
	return payload
}

// requireMutationChildOK memverifikasi mutation sukses lalu mengekstrak field objek anak non-null (mis. user, role, group).
func requireMutationChildOK(t *testing.T, root map[string]any, op, childField string) map[string]any {
	t.Helper()
	payload := requireMutationSuccess(t, root, op)
	return requireChildMap(t, payload, childField)
}

// requireMutationBoolOK memverifikasi mutation sukses serta field boolean penanda pada payload bernilai true.
func requireMutationBoolOK(t *testing.T, root map[string]any, op, boolField string) map[string]any {
	t.Helper()
	payload := requireMutationSuccess(t, root, op)
	if !requireBoolField(t, payload, boolField) {
		t.Fatalf("%s.%s = false, want true", op, boolField)
	}
	return payload
}

// requireOpValue memastikan kunci operasi GraphQL ada dan tidak bernilai JSON null.
func requireOpValue(t *testing.T, root map[string]any, op string) any {
	t.Helper()
	val, ok := root[op]
	if !ok {
		t.Fatalf("GraphQL response missing operation field %q in %+v", op, root)
	}
	if val == nil {
		t.Fatalf("GraphQL operation %q returned unexpected null", op)
	}
	return val
}

// requireOpMap memastikan operasi GraphQL mengembalikan objek JSON non-null.
func requireOpMap(t *testing.T, root map[string]any, op string) map[string]any {
	t.Helper()
	raw := requireOpValue(t, root, op)
	m, ok := raw.(map[string]any)
	if !ok || m == nil {
		t.Fatalf("GraphQL operation %q returned %T, want non-null JSON object", op, raw)
	}
	return m
}

// requireOpSlice memastikan operasi GraphQL mengembalikan array JSON non-null.
func requireOpSlice(t *testing.T, root map[string]any, op string) []any {
	t.Helper()
	raw := requireOpValue(t, root, op)
	arr, ok := raw.([]any)
	if !ok || arr == nil {
		t.Fatalf("GraphQL operation %q returned %T, want non-null JSON array", op, raw)
	}
	return arr
}

// requireOpBoolTrue memastikan operasi Boolean! mengembalikan nilai true (bukan null/false).
func requireOpBoolTrue(t *testing.T, root map[string]any, op string) {
	t.Helper()
	raw := requireOpValue(t, root, op)
	b, ok := raw.(bool)
	if !ok || !b {
		t.Fatalf("GraphQL operation %q = %#v, want true", op, raw)
	}
}

// requireOpInt memastikan operasi Int! mengembalikan bilangan bulat non-null.
func requireOpInt(t *testing.T, root map[string]any, op string) int {
	t.Helper()
	raw := requireOpValue(t, root, op)
	n, ok := raw.(float64)
	if !ok {
		t.Fatalf("GraphQL operation %q returned %T (%v), want JSON number", op, raw, raw)
	}
	return int(n)
}

// requireFieldNonNull memastikan setiap field yang diminta ada di objek dan tidak bernilai null.
func requireFieldsNonNull(t *testing.T, obj map[string]any, fields ...string) {
	t.Helper()
	for _, f := range fields {
		v, ok := obj[f]
		if !ok {
			t.Fatalf("object missing field %q: %+v", f, obj)
		}
		if v == nil {
			t.Fatalf("object field %q is unexpected null in %+v", f, obj)
		}
	}
}

// requireNonEmptyString memastikan field bertipe string non-null dan tidak kosong.
func requireNonEmptyString(t *testing.T, obj map[string]any, field string) string {
	t.Helper()
	v, ok := obj[field]
	if !ok || v == nil {
		t.Fatalf("field %q missing or null in %+v", field, obj)
	}
	s, ok := v.(string)
	if !ok || strings.TrimSpace(s) == "" {
		t.Fatalf("field %q = %#v, want non-empty string", field, v)
	}
	return s
}

// requireBoolField membaca field boolean non-null dan memvalidasi nilainya bila diharapkan.
func requireBoolField(t *testing.T, obj map[string]any, field string) bool {
	t.Helper()
	v, ok := obj[field]
	if !ok || v == nil {
		t.Fatalf("field %q missing or null in %+v", field, obj)
	}
	b, ok := v.(bool)
	if !ok {
		t.Fatalf("field %q = %#v, want bool", field, v)
	}
	return b
}

// requireNumberField membaca field numerik non-null sebagai int.
func requireNumberField(t *testing.T, obj map[string]any, field string) int {
	t.Helper()
	v, ok := obj[field]
	if !ok || v == nil {
		t.Fatalf("field %q missing or null in %+v", field, obj)
	}
	n, ok := v.(float64)
	if !ok {
		t.Fatalf("field %q = %#v, want number", field, v)
	}
	return int(n)
}

// requireChildMap memastikan field anak berjenis objek JSON non-null.
func requireChildMap(t *testing.T, obj map[string]any, field string) map[string]any {
	t.Helper()
	v, ok := obj[field]
	if !ok || v == nil {
		t.Fatalf("child object %q missing or null in %+v", field, obj)
	}
	m, ok := v.(map[string]any)
	if !ok || m == nil {
		t.Fatalf("child object %q = %T, want JSON object", field, v)
	}
	return m
}

// requireChildSlice memastikan field anak berjenis array JSON non-null.
func requireChildSlice(t *testing.T, obj map[string]any, field string) []any {
	t.Helper()
	v, ok := obj[field]
	if !ok || v == nil {
		t.Fatalf("child slice %q missing or null in %+v", field, obj)
	}
	s, ok := v.([]any)
	if !ok || s == nil {
		t.Fatalf("child slice %q = %T, want JSON array", field, v)
	}
	return s
}

// assertPageInfoNonNull memastikan struct pageInfo Relay bebas nilai null dan Connection memiliki total_count bertipe numerik.
func assertPageInfoNonNull(t *testing.T, conn map[string]any) map[string]any {
	t.Helper()
	_ = requireNumberField(t, conn, "total_count")
	page := requireChildMap(t, conn, "pageInfo")
	requireKeysPresent(t, page, "startCursor", "endCursor", "hasPreviousPage", "hasNextPage")
	requireNullableStringField(t, page, "startCursor")
	requireNullableStringField(t, page, "endCursor")
	_ = requireBoolField(t, page, "hasPreviousPage")
	_ = requireBoolField(t, page, "hasNextPage")
	return page
}

// requireKeysPresent memastikan setiap nama field ada pada payload JSON (meskipun nilainya nullable).
func requireKeysPresent(t *testing.T, obj map[string]any, fields ...string) {
	t.Helper()
	for _, f := range fields {
		if _, ok := obj[f]; !ok {
			t.Fatalf("object missing key %q: %+v", f, obj)
		}
	}
}

// requireStringField memastikan field non-null bertipe string sesuai skema GraphQL String!.
func requireStringField(t *testing.T, obj map[string]any, field string) string {
	t.Helper()
	v, ok := obj[field]
	if !ok || v == nil {
		t.Fatalf("field %q missing or null in %+v", field, obj)
	}
	s, ok := v.(string)
	if !ok {
		t.Fatalf("field %q = %#v, want string", field, v)
	}
	return s
}

// requireNullableStringField memastikan key ada di JSON dan bertipe string apabila non-null.
func requireNullableStringField(t *testing.T, obj map[string]any, field string) {
	t.Helper()
	v, ok := obj[field]
	if !ok {
		t.Fatalf("nullable field %q key missing from %+v", field, obj)
	}
	if v == nil {
		return
	}
	if _, ok := v.(string); !ok {
		t.Fatalf("nullable field %q = %#v, want string or null", field, v)
	}
}

// assertBrowserSessionPayload memvalidasi seluruh field payload BrowserSessionPayload sesuai status autentikasi.
func assertBrowserSessionPayload(t *testing.T, obj map[string]any, wantAuth bool) {
	t.Helper()
	requireKeysPresent(t, obj, "authenticated", "principalId", "csrfToken")
	gotAuth := requireBoolField(t, obj, "authenticated")
	if gotAuth != wantAuth {
		t.Fatalf("BrowserSessionPayload.authenticated = %v, want %v", gotAuth, wantAuth)
	}
	requireNullableStringField(t, obj, "principalId")
	requireNullableStringField(t, obj, "csrfToken")
	if wantAuth {
		_ = requireNonEmptyString(t, obj, "principalId")
		_ = requireNonEmptyString(t, obj, "csrfToken")
	}
}

// assertWorkspaceAccessPayload memvalidasi seluruh field WorkspaceAccessPayload.
func assertWorkspaceAccessPayload(t *testing.T, obj map[string]any) {
	t.Helper()
	_ = requireBoolField(t, obj, "canAccessIdentity")
	_ = requireNonEmptyString(t, obj, "defaultLocation")
}

// assertUserLoginPayload memvalidasi seluruh field UserLoginPayload sesuai status login (AUTHENTICATED/MFA_REQUIRED).
func assertUserLoginPayload(t *testing.T, obj map[string]any, wantStatus string) {
	t.Helper()
	assertMutationResultSuccess(t, obj, "userLogin")
	requireKeysPresent(t, obj, "status", "mfaChallenge", "mfaExpiresAt", "csrfToken", "redirectLocation")
	status := requireNonEmptyString(t, obj, "status")
	if wantStatus != "" && status != wantStatus {
		t.Fatalf("UserLoginPayload.status = %q, want %q", status, wantStatus)
	}
	requireNullableStringField(t, obj, "mfaChallenge")
	requireNullableStringField(t, obj, "mfaExpiresAt")
	requireNullableStringField(t, obj, "csrfToken")
	requireNullableStringField(t, obj, "redirectLocation")
	switch status {
	case "AUTHENTICATED":
		_ = requireNonEmptyString(t, obj, "csrfToken")
		_ = requireNonEmptyString(t, obj, "redirectLocation")
	case "MFA_REQUIRED":
		_ = requireNonEmptyString(t, obj, "mfaChallenge")
		_ = requireNonEmptyString(t, obj, "mfaExpiresAt")
	default:
		t.Fatalf("unexpected UserLoginPayload.status = %q", status)
	}
}

// assertValidateUserMfaPayload memvalidasi seluruh field ValidateUserMfaPayload.
func assertValidateUserMfaPayload(t *testing.T, obj map[string]any) {
	t.Helper()
	assertMutationResultSuccess(t, obj, "validateUserMfa")
	_ = requireNonEmptyString(t, obj, "csrfToken")
	_ = requireNonEmptyString(t, obj, "redirectLocation")
}

// assertVerifyEmailPayload memvalidasi seluruh field VerifyEmailPayload.
func assertVerifyEmailPayload(t *testing.T, obj map[string]any, wantStatus string, wantVerified bool) {
	t.Helper()
	assertMutationResultSuccess(t, obj, "verifyEmail")
	status := requireNonEmptyString(t, obj, "status")
	verified := requireBoolField(t, obj, "verified")
	if status != wantStatus || verified != wantVerified {
		t.Fatalf("VerifyEmailPayload = (%q, %v), want (%q, %v)", status, verified, wantStatus, wantVerified)
	}
}

// assertEnrollTotpPayload memvalidasi seluruh field EnrollTotpPayload.
func assertEnrollTotpPayload(t *testing.T, obj map[string]any) {
	t.Helper()
	assertMutationResultSuccess(t, obj, "enrollMyTotp")
	_ = requireNonEmptyString(t, obj, "factor_id")
	_ = requireNonEmptyString(t, obj, "otp_auth_uri")
}

// assertConfirmTotpPayload memvalidasi seluruh field ConfirmTotpPayload.
func assertConfirmTotpPayload(t *testing.T, obj map[string]any) []string {
	t.Helper()
	assertMutationResultSuccess(t, obj, "confirmMyTotp")
	if !requireBoolField(t, obj, "enabled") {
		t.Fatalf("ConfirmTotpPayload.enabled = false, want true")
	}
	rawCodes := requireChildSlice(t, obj, "recovery_codes")
	if len(rawCodes) == 0 {
		t.Fatalf("ConfirmTotpPayload.recovery_codes is empty")
	}
	codes := make([]string, 0, len(rawCodes))
	for _, raw := range rawCodes {
		s, ok := raw.(string)
		if !ok || strings.TrimSpace(s) == "" {
			t.Fatalf("ConfirmTotpPayload.recovery_codes element = %#v, want non-empty string", raw)
		}
		codes = append(codes, s)
	}
	return codes
}

// assertScopeObject memvalidasi objek Scope (type non-null, id nullable).
func assertScopeObject(t *testing.T, scope map[string]any) {
	t.Helper()
	_ = requireNonEmptyString(t, scope, "type")
	requireNullableStringField(t, scope, "id")
}

// assertUserObject memvalidasi lengkap seluruh 13 field objek GraphQL User.
func assertUserObject(t *testing.T, u map[string]any) {
	t.Helper()
	_ = requireNonEmptyString(t, u, "id")
	_ = requireNonEmptyString(t, u, "kind")
	_ = requireNonEmptyString(t, u, "status")
	_ = requireNonEmptyString(t, u, "display_name")
	_ = requireStringField(t, u, "avatar_url")
	_ = requireNonEmptyString(t, u, "locale")
	_ = requireNonEmptyString(t, u, "timezone")
	requireNullableStringField(t, u, "last_login_at")
	_ = requireNonEmptyString(t, u, "created_at")
	_ = requireNonEmptyString(t, u, "updated_at")
	requireNullableStringField(t, u, "deleted_at")
	_ = requireNumberField(t, u, "auth_epoch")
	_ = requireNumberField(t, u, "perm_revision")
}

// assertUserIdentityObject memvalidasi lengkap seluruh 9 field objek GraphQL UserIdentity.
func assertUserIdentityObject(t *testing.T, ident map[string]any) {
	t.Helper()
	_ = requireNonEmptyString(t, ident, "id")
	_ = requireNonEmptyString(t, ident, "user_id")
	_ = requireNonEmptyString(t, ident, "kind")
	_ = requireNonEmptyString(t, ident, "value")
	_ = requireBoolField(t, ident, "is_verified")
	_ = requireBoolField(t, ident, "is_primary")
	requireNullableStringField(t, ident, "verified_at")
	_ = requireNonEmptyString(t, ident, "created_at")
	_ = requireNonEmptyString(t, ident, "updated_at")
}

// assertRoleObject memvalidasi lengkap seluruh 11 field objek GraphQL Role.
func assertRoleObject(t *testing.T, role map[string]any) {
	t.Helper()
	_ = requireNonEmptyString(t, role, "id")
	_ = requireNonEmptyString(t, role, "code")
	_ = requireNonEmptyString(t, role, "name")
	_ = requireStringField(t, role, "description")
	_ = requireStringField(t, role, "management_domain")
	_ = requireStringField(t, role, "default_redirect_path")
	_ = requireNumberField(t, role, "redirect_priority")
	_ = requireBoolField(t, role, "is_system")
	_ = requireBoolField(t, role, "is_active")
	_ = requireNonEmptyString(t, role, "created_at")
	_ = requireNonEmptyString(t, role, "updated_at")
}

// assertPermissionObject memvalidasi lengkap seluruh 7 field objek GraphQL Permission.
func assertPermissionObject(t *testing.T, perm map[string]any) {
	t.Helper()
	_ = requireNonEmptyString(t, perm, "id")
	_ = requireNonEmptyString(t, perm, "code")
	_ = requireNonEmptyString(t, perm, "resource")
	_ = requireNonEmptyString(t, perm, "action")
	_ = requireStringField(t, perm, "description")
	_ = requireBoolField(t, perm, "is_sensitive")
	_ = requireNonEmptyString(t, perm, "created_at")
}

// assertRolePermissionObject memvalidasi lengkap seluruh field Edge GraphQL RolePermissionEdge beserta node Permission.
func assertRolePermissionObject(t *testing.T, rp map[string]any) {
	t.Helper()
	_ = requireNonEmptyString(t, rp, "cursor")
	_ = requireNonEmptyString(t, rp, "role_id")
	_ = requireNonEmptyString(t, rp, "created_at")
	assertPermissionObject(t, requireChildMap(t, rp, "node"))
}

// assertUserRoleObject memvalidasi lengkap seluruh field Edge GraphQL UserRoleEdge beserta node Role.
func assertUserRoleObject(t *testing.T, ur map[string]any) {
	t.Helper()
	_ = requireNonEmptyString(t, ur, "cursor")
	_ = requireNonEmptyString(t, ur, "assignment_id")
	_ = requireNonEmptyString(t, ur, "user_id")
	assertScopeObject(t, requireChildMap(t, ur, "scope"))
	requireNullableStringField(t, ur, "granted_by")
	requireNullableStringField(t, ur, "expires_at")
	_ = requireNonEmptyString(t, ur, "created_at")
	assertRoleObject(t, requireChildMap(t, ur, "node"))
}

// assertGroupObject memvalidasi lengkap seluruh 7 field objek GraphQL Group.
func assertGroupObject(t *testing.T, grp map[string]any) {
	t.Helper()
	_ = requireNonEmptyString(t, grp, "id")
	_ = requireNonEmptyString(t, grp, "code")
	_ = requireNonEmptyString(t, grp, "name")
	_ = requireStringField(t, grp, "description")
	_ = requireBoolField(t, grp, "is_active")
	_ = requireNonEmptyString(t, grp, "created_at")
	_ = requireNonEmptyString(t, grp, "updated_at")
}

// assertGroupMemberObject memvalidasi lengkap seluruh field Edge GraphQL GroupMemberEdge beserta node User.
func assertGroupMemberObject(t *testing.T, gm map[string]any) {
	t.Helper()
	_ = requireNonEmptyString(t, gm, "cursor")
	_ = requireNonEmptyString(t, gm, "group_id")
	_ = requireNonEmptyString(t, gm, "created_at")
	assertUserObject(t, requireChildMap(t, gm, "node"))
}

// assertGroupRoleObject memvalidasi lengkap seluruh field Edge GraphQL GroupRoleEdge beserta node Role.
func assertGroupRoleObject(t *testing.T, gr map[string]any) {
	t.Helper()
	_ = requireNonEmptyString(t, gr, "cursor")
	_ = requireNonEmptyString(t, gr, "assignment_id")
	_ = requireNonEmptyString(t, gr, "group_id")
	assertScopeObject(t, requireChildMap(t, gr, "scope"))
	requireNullableStringField(t, gr, "granted_by")
	requireNullableStringField(t, gr, "expires_at")
	_ = requireNonEmptyString(t, gr, "created_at")
	assertRoleObject(t, requireChildMap(t, gr, "node"))
}

// assertSessionObject memvalidasi lengkap seluruh 14 field objek GraphQL Session.
func assertSessionObject(t *testing.T, s map[string]any) {
	t.Helper()
	_ = requireNonEmptyString(t, s, "id")
	_ = requireNonEmptyString(t, s, "user_id")
	_ = requireBoolField(t, s, "mfa_verified")
	requireNullableStringField(t, s, "mfa_verified_at")
	_ = requireStringField(t, s, "ip_address")
	_ = requireStringField(t, s, "user_agent")
	_ = requireNonEmptyString(t, s, "authenticated_at")
	_ = requireNonEmptyString(t, s, "last_active_at")
	_ = requireNonEmptyString(t, s, "idle_expires_at")
	_ = requireNonEmptyString(t, s, "absolute_expires_at")
	requireNullableStringField(t, s, "ended_at")
	requireNullableStringField(t, s, "end_reason")
	_ = requireNonEmptyString(t, s, "created_at")
	_ = requireNumberField(t, s, "generation")
}

// assertAccessSnapshotObject memvalidasi lengkap seluruh field AccessSnapshot beserta nested RoleGrant dan ScopedPermission.
func assertAccessSnapshotObject(t *testing.T, snap map[string]any) {
	t.Helper()
	_ = requireNonEmptyString(t, snap, "principal_id")
	_ = requireNonEmptyString(t, snap, "user_kind")
	_ = requireNonEmptyString(t, snap, "status")
	_ = requireBoolField(t, snap, "mfa_enabled")
	_ = requireNumberField(t, snap, "auth_epoch")
	_ = requireNumberField(t, snap, "perm_revision")
	_ = requireNumberField(t, snap, "recovery_epoch")
	_ = requireBoolField(t, snap, "can_access_identity")
	_ = requireStringField(t, snap, "default_redirect_path")
	_ = requireNumberField(t, snap, "synced_at_unix")
	_ = requireNumberField(t, snap, "fresh_until_unix")

	for _, rawPath := range requireChildSlice(t, snap, "allowed_redirect_paths") {
		p, ok := rawPath.(string)
		if !ok || strings.TrimSpace(p) == "" {
			t.Fatalf("AccessSnapshot.allowed_redirect_paths element = %#v, want non-empty string", rawPath)
		}
	}
	for _, rawRole := range requireChildSlice(t, snap, "roles") {
		rg, ok := rawRole.(map[string]any)
		if !ok || rg == nil {
			t.Fatalf("AccessSnapshot.roles element = %#v, want object", rawRole)
		}
		_ = requireNonEmptyString(t, rg, "id")
		_ = requireNonEmptyString(t, rg, "code")
		_ = requireNonEmptyString(t, rg, "name")
		_ = requireNumberField(t, rg, "priority")
		_ = requireStringField(t, rg, "default_redirect_path")
	}
	for _, rawPerm := range requireChildSlice(t, snap, "permissions") {
		sp, ok := rawPerm.(map[string]any)
		if !ok || sp == nil {
			t.Fatalf("AccessSnapshot.permissions element = %#v, want object", rawPerm)
		}
		_ = requireNonEmptyString(t, sp, "permission_code")
		_ = requireBoolField(t, sp, "is_sensitive")
		assertScopeObject(t, requireChildMap(t, sp, "scope"))
	}
}

// cleanupUserByID menghapus user uji beserta jejak outbox dan kunci Redis-nya agar DB tetap bersih.
func cleanupUserByID(t *testing.T, cfg e2eRuntimeConfig, userID string) {
	t.Helper()
	userID = rawUUID(userID)
	if userID == "" {
		return
	}
	// Pengaman keras agar SystemIAM utama tidak pernah terhapus oleh cleanup test.
	if userID == "853aac77-32e7-4a88-b2d6-7d68db8d13d7" {
		return
	}
	psqlExec(t, cfg, fmt.Sprintf(`DELETE FROM iam_outbox_events WHERE aggregate_id = %s::uuid;`, sqlQuote(userID)))
	psqlExec(t, cfg, fmt.Sprintf(`DELETE FROM users WHERE id = %s::uuid;`, sqlQuote(userID)))

	redisClient := newE2ERedisClient(t, cfg, "iam_writer", cfg.iamRedisPass)
	defer func() { _ = redisClient.Close() }()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_ = redisClient.Del(
		ctx,
		authsession.NewKeyspace("development").AuthzSnapshotKey(userID),
		authsession.NewKeyspace("development").UserSecurityKey(userID),
	).Err()
}

// cleanupRoleByID menghapus peran kustom uji termasuk relasi RESTRICT apabila test gagal di tengah jalan.
func cleanupRoleByID(t *testing.T, cfg e2eRuntimeConfig, roleID string) {
	t.Helper()
	roleID = rawUUID(roleID)
	if roleID == "" {
		return
	}
	psqlExec(t, cfg, fmt.Sprintf(`DELETE FROM user_roles WHERE role_id = %s::uuid;`, sqlQuote(roleID)))
	psqlExec(t, cfg, fmt.Sprintf(`DELETE FROM group_roles WHERE role_id = %s::uuid;`, sqlQuote(roleID)))
	psqlExec(t, cfg, fmt.Sprintf(`DELETE FROM roles WHERE id = %s::uuid AND is_system = false;`, sqlQuote(roleID)))
}

// cleanupGroupByID menghapus grup uji beserta relasinya.
func cleanupGroupByID(t *testing.T, cfg e2eRuntimeConfig, groupID string) {
	t.Helper()
	groupID = rawUUID(groupID)
	if groupID == "" {
		return
	}
	psqlExec(t, cfg, fmt.Sprintf(`DELETE FROM groups WHERE id = %s::uuid;`, sqlQuote(groupID)))
}

// psqlEFScalar menjalankan SQL query tunggal pada database `enterprise_foundation` dan mengembalikan keluaran scalar.
func psqlEFScalar(t *testing.T, cfg e2eRuntimeConfig, sqlStmt string) string {
	t.Helper()
	cmd := exec.Command(
		"docker", "exec", cfg.postgresService,
		"psql", "-v", "ON_ERROR_STOP=1", "-U", "postgres", "-d", "enterprise_foundation", "-tAc", sqlStmt,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("ef psql failed: %v\nSQL: %s\nOutput: %s", err, sqlStmt, strings.TrimSpace(string(out)))
	}
	return strings.TrimSpace(string(out))
}

// createFixturePartyForRole membuat mitra aktif pada Enterprise Foundation beserta peran VENDOR/SUPPLIER untuk provisioning user E2E.
func createFixturePartyForRole(t *testing.T, admin *browserClient, cfg e2eRuntimeConfig, roleKind string) string {
	t.Helper()
	tag := uniqueTag("efp_" + strings.ToLower(roleKind))
	party := requireMutationChildOK(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($in: CreatePartyInput!) {
		createParty(input: $in) { party { id _id status } %s }
	}`, gqlMutationResultFields), map[string]any{
		"in": map[string]any{
			"partner_code":       strings.ToUpper(roleKind[:3]) + "-" + strings.ToUpper(tag),
			"kind":               "ORGANIZATION",
			"display_name":       "PT Mitra " + tag,
			"email":              tag + "@partner.integrione.test",
			"preferred_language": "id",
		},
	}), "createParty", "party")
	partyGID := requireNonEmptyString(t, party, "id")
	partyRawID := requireNonEmptyString(t, party, "_id")
	t.Cleanup(func() {
		cmd := exec.Command(
			"docker", "exec", cfg.postgresService,
			"psql", "-U", "postgres", "-d", "enterprise_foundation", "-tAc",
			fmt.Sprintf("DELETE FROM parties WHERE id = %s::uuid;", sqlQuote(partyRawID)),
		)
		_, _ = cmd.CombinedOutput()
	})

	_ = requireMutationChildOK(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($id: ID!) {
		verifyAndActivateParty(id: $id) { party { id status } %s }
	}`, gqlMutationResultFields), map[string]any{"id": partyGID}), "verifyAndActivateParty", "party")

	_ = requireMutationChildOK(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($in: AssignPartyRoleInput!) {
		assignPartyRole(input: $in) { role { id role is_active } %s }
	}`, gqlMutationResultFields), map[string]any{
		"in": map[string]any{
			"party_id": partyGID,
			"role":     roleKind,
		},
	}), "assignPartyRole", "role")

	return partyGID
}

// createFixtureEFFacility membuat badan hukum dan lokasi aktif di Enterprise Foundation sebagai target scope FACILITY yang valid.
func createFixtureEFFacility(t *testing.T, admin *browserClient, cfg e2eRuntimeConfig) string {
	t.Helper()
	tag := uniqueTag("effac")
	normCode := strings.ToUpper(strings.ReplaceAll(tag, "_", "-"))

	le := requireMutationChildOK(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($in: CreateLegalEntityInput!) {
		createLegalEntity(input: $in) { legal_entity { id _id status } %s }
	}`, gqlMutationResultFields), map[string]any{
		"in": map[string]any{
			"entity_code":               "LE-" + normCode,
			"legal_name":                "PT Scope Facility " + tag,
			"jurisdiction_country_code": "ID",
		},
	}), "createLegalEntity", "legal_entity")
	leID := requireNonEmptyString(t, le, "id")
	leRawID := requireNonEmptyString(t, le, "_id")

	_ = requireMutationChildOK(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($id: ID!) {
		activateLegalEntity(id: $id) { legal_entity { id status } %s }
	}`, gqlMutationResultFields), map[string]any{"id": leID}), "activateLegalEntity", "legal_entity")

	loc := requireMutationChildOK(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($in: CreateLocationInput!) {
		createLocation(input: $in) { location { id _id is_active } %s }
	}`, gqlMutationResultFields), map[string]any{
		"in": map[string]any{
			"legal_entity_id": leID,
			"location_code":   "LOC-" + normCode,
			"name":            "Gudang Fixture " + tag,
			"kind":            "WAREHOUSE",
			"timezone":        "Asia/Jakarta",
		},
	}), "createLocation", "location")
	locRawID := requireNonEmptyString(t, loc, "_id")

	t.Cleanup(func() {
		cleanupEFEntityRows(t, cfg, "locations", locRawID)
		cleanupEFEntityRows(t, cfg, "legal_entities", leRawID)
	})
	return locRawID
}

// createFixtureUser membuat akun uji melalui operasi createUser* (status awal PENDING) lalu mengaktifkannya untuk kebutuhan fixture login.
func createFixtureUser(t *testing.T, admin *browserClient, cfg e2eRuntimeConfig, kind string, withUsername bool) fixtureUserAccount {
	t.Helper()
	_ = withUsername
	tag := uniqueTag("u_" + strings.ToLower(kind))
	email := tag + "@integrione.test"
	username := tag
	password := "StrongPass#2026!"

	var opName, inputType string
	var payloadFields = gqlUserPayloadFields
	input := map[string]any{
		"full_name": "Fixture " + tag,
		"email":     email,
		"username":  username,
		"password":  password,
	}
	switch kind {
	case "CUSTOMER":
		opName, inputType = "createUser", "CreateUserInput"
	case "EMPLOYEE":
		opName, inputType = "createEmployeeUser", "CreateEmployeeUserInput"
	case "VENDOR":
		opName, inputType = "createVendorUser", "CreateVendorUserInput"
		input["party_id"] = createFixturePartyForRole(t, admin, cfg, "VENDOR")
		payloadFields = `user { ` + gqlUserFields + ` } party_contact { id _id party_id iam_user_id portal_access_status } ` + gqlMutationResultFields
	case "SUPPLIER":
		opName, inputType = "createSupplierUser", "CreateSupplierUserInput"
		input["party_id"] = createFixturePartyForRole(t, admin, cfg, "SUPPLIER")
		payloadFields = `user { ` + gqlUserFields + ` } party_contact { id _id party_id iam_user_id portal_access_status } ` + gqlMutationResultFields
	default:
		t.Fatalf("unsupported fixture user kind %q", kind)
	}

	query := fmt.Sprintf(`mutation ($input: %s!) { %s(input: $input) { %s } }`, inputType, opName, payloadFields)
	data := doGQLOK(t, admin, query, map[string]any{"input": input})
	u := requireMutationChildOK(t, data, opName, "user")
	assertUserObject(t, u)
	if u["status"] != "PENDING" {
		t.Fatalf("%s initial status = %v, want PENDING", opName, u["status"])
	}
	id := requireNonEmptyString(t, u, "id")
	rawID := requireNonEmptyString(t, u, "_id")
	if kind == "VENDOR" || kind == "SUPPLIER" {
		pc := requireMutationChildOK(t, data, opName, "party_contact")
		if pc["iam_user_id"] != id || pc["portal_access_status"] != "PENDING_LINK" {
			t.Fatalf("%s party_contact FK mismatch: %+v", opName, pc)
		}
		if fkCnt := psqlEFScalar(t, cfg, fmt.Sprintf(`SELECT count(*) FROM party_contacts WHERE iam_user_id = %s::uuid AND portal_access_status = 'pending_link' AND deleted_at IS NULL;`, sqlQuote(rawID))); fkCnt != "1" {
			t.Fatalf("%s did not persist EF party_contacts.iam_user_id FK: count=%s", opName, fkCnt)
		}
	}
	t.Cleanup(func() {
		cleanupUserByID(t, cfg, rawID)
	})

	// Aktivasi user + tandai identitas email/username terverifikasi agar fixture siap diuji login.
	actData := doGQLOK(t, admin, fmt.Sprintf(`mutation ($user_id: ID!) { activateUser(user_id: $user_id) { %s } }`, gqlUserPayloadFields), map[string]any{"user_id": id})
	actUser := requireMutationChildOK(t, actData, "activateUser", "user")
	assertUserObject(t, actUser)
	psqlExec(t, cfg, fmt.Sprintf(`UPDATE user_identities SET is_verified = true, verified_at = now() WHERE user_id = %s::uuid;`, sqlQuote(rawID)))

	return fixtureUserAccount{
		ID:       id,
		RawID:    rawID,
		Kind:     kind,
		Email:    email,
		Username: username,
		Password: password,
	}
}

// loginFixtureUser membuat browser terpisah lalu login menggunakan akun fixture yang diberikan.
func loginFixtureUser(t *testing.T, cfg e2eRuntimeConfig, identifier, password string) (*browserClient, map[string]any) {
	t.Helper()
	client := newAnonBrowser(t, cfg)
	query := fmt.Sprintf(`mutation ($input: UserLoginInput!) {
		userLogin(input: $input) { %s }
	}`, gqlUserLoginFields)
	data := doGQLOK(t, client, query, map[string]any{
		"input": map[string]any{
			"identifier": identifier,
			"password":   password,
		},
	})
	payload := requireOpMap(t, data, "userLogin")
	assertUserLoginPayload(t, payload, "")
	t.Cleanup(func() {
		_, _, _ = logoutBrowserBestEffort(client)
	})
	return client, payload
}

// createFixtureCustomerAndLogin membuat akun customer aktif non-privileged untuk pengujian negatif RBAC TestAB.
func createFixtureCustomerAndLogin(t *testing.T, admin *browserClient, cfg e2eRuntimeConfig) (fixtureUserAccount, *browserClient) {
	t.Helper()
	cust := createFixtureUser(t, admin, cfg, "CUSTOMER", false)
	client, payload := loginFixtureUser(t, cfg, cust.Email, cust.Password)
	assertUserLoginPayload(t, payload, "AUTHENTICATED")
	return cust, client
}

// createFixtureRole membuat role kustom uji beserta cleanup otomatis.
func createFixtureRole(t *testing.T, admin *browserClient, cfg e2eRuntimeConfig, prefix string) (string, string) {
	t.Helper()
	code := uniqueTag("role_" + prefix)
	data := doGQLOK(t, admin, fmt.Sprintf(`mutation ($input: CreateRoleInput!) {
		createRole(input: $input) { %s }
	}`, gqlRolePayloadFields), map[string]any{
		"input": map[string]any{
			"code":                  code,
			"name":                  "Role " + code,
			"description":           "Custom E2E role " + code,
			"management_domain":     "iam",
			"default_redirect_path": "/identity",
			"redirect_priority":     50,
			"is_system":             false,
		},
	})
	role := requireMutationChildOK(t, data, "createRole", "role")
	assertRoleObject(t, role)
	roleID := requireNonEmptyString(t, role, "id")
	t.Cleanup(func() {
		cleanupRoleByID(t, cfg, roleID)
	})
	return roleID, code
}

// createFixtureGroup membuat group uji beserta cleanup otomatis.
func createFixtureGroup(t *testing.T, admin *browserClient, cfg e2eRuntimeConfig, prefix string) (string, string) {
	t.Helper()
	code := uniqueTag("grp_" + prefix)
	data := doGQLOK(t, admin, fmt.Sprintf(`mutation ($input: CreateGroupInput!) {
		createGroup(input: $input) { %s }
	}`, gqlGroupPayloadFields), map[string]any{
		"input": map[string]any{
			"code":        code,
			"name":        "Group " + code,
			"description": "Group E2E " + code,
		},
	})
	grp := requireMutationChildOK(t, data, "createGroup", "group")
	assertGroupObject(t, grp)
	groupID := requireNonEmptyString(t, grp, "id")
	t.Cleanup(func() {
		cleanupGroupByID(t, cfg, groupID)
	})
	return groupID, code
}

// findPermissionIDByCode mencari ID permission bawaan katalog berdasarkan kode exact.
func findPermissionIDByCode(t *testing.T, admin *browserClient, permCode string) string {
	t.Helper()
	data := doGQLOK(t, admin, fmt.Sprintf(`query ($query: String) {
		permissions(query: $query, first: 25) {
			nodes { %s }
			pageInfo { %s }
			total_count
		}
	}`, gqlPermissionFields, gqlPageInfoFields), map[string]any{"query": permCode})
	conn := requireOpMap(t, data, "permissions")
	_ = assertPageInfoNonNull(t, conn)
	items := requireConnectionNodes(t, conn)
	for _, raw := range items {
		item, ok := raw.(map[string]any)
		if ok {
			assertPermissionObject(t, item)
			if item["code"] == permCode {
				return requireNonEmptyString(t, item, "id")
			}
		}
	}
	t.Fatalf("permission code %q not found in catalog", permCode)
	return ""
}

// enrollAndConfirmTOTP melakukan aktivasi TOTP penuh pada sesi pengguna dan mengembalikan otp_auth_uri + recovery_codes.
func enrollAndConfirmTOTP(t *testing.T, client *browserClient) (string, string, []string) {
	t.Helper()
	enrollData := doGQLOK(t, client, fmt.Sprintf(`mutation {
		enrollMyTotp { %s }
	}`, gqlEnrollTotpFields), nil)
	enrollPayload := requireOpMap(t, enrollData, "enrollMyTotp")
	assertEnrollTotpPayload(t, enrollPayload)
	factorID := requireNonEmptyString(t, enrollPayload, "factor_id")
	otpAuthURI := requireNonEmptyString(t, enrollPayload, "otp_auth_uri")

	// Menggunakan step T-1 (dalam batas skew +-30s) agar langkah verifikasi MFA berikutnya pada T tidak terkena konflik anti-replay.
	code := generateTOTPCode(t, otpAuthURI, time.Now().Add(-30*time.Second))
	confirmData := doGQLOK(t, client, fmt.Sprintf(`mutation ($input: ConfirmTotpInput!) {
		confirmMyTotp(input: $input) { %s }
	}`, gqlConfirmTotpFields), map[string]any{"input": map[string]any{"code": code}})
	confirmPayload := requireOpMap(t, confirmData, "confirmMyTotp")
	codes := assertConfirmTotpPayload(t, confirmPayload)
	return factorID, otpAuthURI, codes
}

// backdateSessionForStaleSensitiveTest memundurkan issued_at sesi di PostgreSQL dan Redis (>10 menit) untuk menguji penolakan aksi sensitif.
func backdateSessionForStaleSensitiveTest(t *testing.T, cfg e2eRuntimeConfig, client *browserClient, userID string) {
	t.Helper()
	userID = rawUUID(userID)
	parsedURL, err := url.Parse(cfg.baseURL)
	if err != nil {
		t.Fatalf("parse baseURL: %v", err)
	}
	var sessionCookieValue string
	for _, ck := range client.httpClient.Jar.Cookies(parsedURL) {
		if ck.Name == "__Host-integrione_session" {
			sessionCookieValue = ck.Value
			break
		}
	}
	if strings.TrimSpace(sessionCookieValue) == "" {
		t.Fatalf("missing session cookie on browser client")
	}

	staleUnix := time.Now().Add(-15 * time.Minute).Unix()
	psqlExec(t, cfg, fmt.Sprintf(`UPDATE sessions SET authenticated_at = now() - interval '15 minutes' WHERE user_id = %s::uuid AND ended_at IS NULL;`, sqlQuote(userID)))

	redisClient := newE2ERedisClient(t, cfg, "iam_writer", cfg.iamRedisPass)
	defer func() { _ = redisClient.Close() }()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := authsession.NewKeyspace("development").SessionTokenKey(sha256Hex(sessionCookieValue))
	rawJSON, err := redisClient.Get(ctx, key).Result()
	if err != nil {
		t.Fatalf("read Redis session token record %s: %v", key, err)
	}
	var rec authsession.SessionTokenRecord
	if err := json.Unmarshal([]byte(rawJSON), &rec); err != nil {
		t.Fatalf("unmarshal SessionTokenRecord: %v", err)
	}
	rec.IssuedAtUnix = staleUnix
	rec.MFAAuthenticated = false
	updatedBytes, err := json.Marshal(rec)
	if err != nil {
		t.Fatalf("marshal SessionTokenRecord: %v", err)
	}
	ttl, err := redisClient.TTL(ctx, key).Result()
	if err != nil || ttl <= 0 {
		ttl = 20 * time.Minute
	}
	if err := redisClient.Set(ctx, key, string(updatedBytes), ttl).Err(); err != nil {
		t.Fatalf("write backdated SessionTokenRecord to Redis: %v", err)
	}
}

// requireCustomerForbidden memastikan endpoint admin menolak pengguna non-privileged dengan kode FORBIDDEN (TestAB).
func requireCustomerForbidden(t *testing.T, custClient *browserClient, query string, vars map[string]any) {
	t.Helper()
	_ = doGQLErrorLabeled(t, "RBAC_TEST_AB_FORBIDDEN", custClient, query, vars, "FORBIDDEN")
}

// requireCustomerMutationForbidden memastikan mutation admin menolak Customer non-privileged melalui payload MutationResult (AUTHORIZATION).
func requireCustomerMutationForbidden(t *testing.T, custClient *browserClient, query string, vars map[string]any, opName string) {
	t.Helper()
	_, _ = doMutationFailLabeled(t, "RBAC_TEST_AB_FORBIDDEN", custClient, query, vars, opName, "AUTHORIZATION")
}

// requireAnonUnauthenticated memastikan endpoint terproteksi menolak sesi anonim dengan UNAUTHENTICATED.
func requireAnonUnauthenticated(t *testing.T, anon *browserClient, query string, vars map[string]any) {
	t.Helper()
	_ = doGQLErrorLabeled(t, "UNAUTHENTICATED_GUARD", anon, query, vars, "UNAUTHENTICATED")
}

// requireAnonMutationUnauthenticated memastikan mutation terproteksi menolak sesi anonim melalui payload MutationResult (AUTHENTICATION).
func requireAnonMutationUnauthenticated(t *testing.T, anon *browserClient, query string, vars map[string]any, opName string) {
	t.Helper()
	_, _ = doMutationFailLabeled(t, "UNAUTHENTICATED_GUARD", anon, query, vars, opName, "AUTHENTICATION")
}

// ============================================================================
// BAGIAN A: 22 QUERY IAM GRAPHQL (TestQuery*)
// ============================================================================

// TestQueryBrowserSession menguji Query browserSession pada kondisi pre-auth anonim, terautentikasi penuh, dan pasca-logout.
func TestQueryBrowserSession(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	q := fmt.Sprintf(`query { browserSession { %s } }`, gqlBrowserSessionFields)

	t.Run("PreAuthAnonymous", func(t *testing.T) {
		resetGatewayRateLimits(t, cfg)
		anon := newAnonBrowser(t, cfg)
		data := doGQLOK(t, anon, q, nil)
		res := requireOpMap(t, data, "browserSession")
		assertBrowserSessionPayload(t, res, false)
		_ = requireNonEmptyString(t, res, "csrfToken")
	})

	t.Run("AuthenticatedAllFieldsNonNull", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		data := doGQLOK(t, admin, q, nil)
		res := requireOpMap(t, data, "browserSession")
		assertBrowserSessionPayload(t, res, true)
	})

	t.Run("AfterLogoutReturnsUnauthenticated", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		emp := createFixtureUser(t, admin, cfg, "EMPLOYEE", false)
		empClient, _ := loginFixtureUser(t, cfg, emp.Email, emp.Password)
		requireMutationBoolOK(t, doGQLOK(t, empClient, fmt.Sprintf(`mutation { userLogout { %s } }`, gqlUserLogoutFields), nil), "userLogout", "logged_out")

		data := doGQLOK(t, empClient, q, nil)
		res := requireOpMap(t, data, "browserSession")
		assertBrowserSessionPayload(t, res, false)
	})
}

// TestQueryMyWorkspaceAccess menguji Query myWorkspaceAccess untuk SuperAdmin (/identity), Customer (/), dan anonim.
func TestQueryMyWorkspaceAccess(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	q := fmt.Sprintf(`query { myWorkspaceAccess { %s } }`, gqlWorkspaceAccessFields)

	t.Run("SuperAdminIdentityWorkspace", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		data := doGQLOK(t, admin, q, nil)
		res := requireOpMap(t, data, "myWorkspaceAccess")
		assertWorkspaceAccessPayload(t, res)
		if !requireBoolField(t, res, "canAccessIdentity") {
			t.Fatalf("expected superadmin canAccessIdentity=true")
		}
		if requireNonEmptyString(t, res, "defaultLocation") != "/identity" {
			t.Fatalf("defaultLocation = %v, want /identity", res["defaultLocation"])
		}
	})

	t.Run("CustomerStorefrontAccess", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		_, custClient := createFixtureCustomerAndLogin(t, admin, cfg)
		data := doGQLOK(t, custClient, q, nil)
		res := requireOpMap(t, data, "myWorkspaceAccess")
		assertWorkspaceAccessPayload(t, res)
		if requireBoolField(t, res, "canAccessIdentity") {
			t.Fatalf("expected customer canAccessIdentity=false")
		}
		if requireNonEmptyString(t, res, "defaultLocation") != "/" {
			t.Fatalf("defaultLocation = %v, want /", res["defaultLocation"])
		}
	})

	t.Run("AnonymousUnauthenticated", func(t *testing.T) {
		resetGatewayRateLimits(t, cfg)
		requireAnonUnauthenticated(t, newAnonBrowser(t, cfg), q, nil)
	})
}

// TestQueryMe menguji Query me memastikan seluruh field User (termasuk last_login_at setelah login) tidak null.
func TestQueryMe(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	q := fmt.Sprintf(`query { me { %s } }`, gqlUserFields)

	t.Run("EmployeeProfileCompleteNonNull", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		emp := createFixtureUser(t, admin, cfg, "EMPLOYEE", false)
		empClient, _ := loginFixtureUser(t, cfg, emp.Email, emp.Password)

		data := doGQLOK(t, empClient, q, nil)
		user := requireOpMap(t, data, "me")
		assertUserObject(t, user)
		_ = requireNonEmptyString(t, user, "last_login_at")
		if user["id"] != emp.ID || user["_id"] != emp.RawID || user["kind"] != "EMPLOYEE" {
			t.Fatalf("unexpected me mismatch: %+v", user)
		}
	})

	t.Run("AnonymousUnauthenticated", func(t *testing.T) {
		resetGatewayRateLimits(t, cfg)
		requireAnonUnauthenticated(t, newAnonBrowser(t, cfg), q, nil)
	})
}

// TestQueryNodeAndNestedResolvers menguji Query.node (Global ID GID + RBAC) serta field relasi Relay bersarang pada User/Group/Role.
func TestQueryNodeAndNestedResolvers(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)

	t.Run("NodeLookupNestedFieldsAndRBAC", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		emp := createFixtureUser(t, admin, cfg, "EMPLOYEE", true)
		roleID, roleCode := createFixtureRole(t, admin, cfg, "node_role")
		groupID, groupCode := createFixtureGroup(t, admin, cfg, "node_grp")
		permID := findPermissionIDByCode(t, admin, "iam.users.read")

		_ = requireMutationChildOK(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($r: ID!, $p: ID!) { assignRolePermission(role_id: $r, permission_id: $p) { %s } }`, gqlRolePermPayloadFields), map[string]any{"r": roleID, "p": permID}), "assignRolePermission", "role_permission_edge")
		_ = requireMutationChildOK(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($g: ID!, $u: ID!) { addGroupMember(group_id: $g, user_id: $u) { %s } }`, gqlGroupMemberPayloadFields), map[string]any{"g": groupID, "u": emp.ID}), "addGroupMember", "group_member_edge")
		_ = requireMutationChildOK(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($input: AssignGroupRoleInput!) { assignGroupRole(input: $input) { %s } }`, gqlGroupRolePayloadFields), map[string]any{"input": map[string]any{"group_id": groupID, "role_id": roleID, "scope_type": "GLOBAL"}}), "assignGroupRole", "group_role_edge")

		empClient, _ := loginFixtureUser(t, cfg, emp.Email, emp.Password)

		// 1. Nested resolvers pada Query.me
		meNestedQ := fmt.Sprintf(`query {
			me {
				id
				identities(first: 10) { nodes { %s } pageInfo { %s } total_count }
				sessions(first: 10, active_only: true) { nodes { %s } pageInfo { %s } total_count }
				roles(first: 10) { pageInfo { %s } total_count }
				effective_roles(first: 10) { nodes { %s } pageInfo { %s } total_count }
				groups(first: 10) { nodes { %s } pageInfo { %s } total_count }
			}
		}`, gqlIdentityFields, gqlPageInfoFields, gqlSessionFields, gqlPageInfoFields, gqlPageInfoFields, gqlRoleFields, gqlPageInfoFields, gqlGroupFields, gqlPageInfoFields)
		meObj := requireOpMap(t, doGQLOK(t, empClient, meNestedQ, nil), "me")
		if requireNumberField(t, requireChildMap(t, meObj, "identities"), "total_count") != 2 {
			t.Fatalf("expected 2 nested identities on me: %+v", meObj["identities"])
		}
		if requireNumberField(t, requireChildMap(t, meObj, "sessions"), "total_count") < 1 {
			t.Fatalf("expected >= 1 active session on me: %+v", meObj["sessions"])
		}
		effNodes := requireConnectionNodes(t, requireChildMap(t, meObj, "effective_roles"))
		var hasInheritedRole bool
		for _, raw := range effNodes {
			if r, ok := raw.(map[string]any); ok && r["code"] == roleCode {
				hasInheritedRole = true
				break
			}
		}
		if !hasInheritedRole {
			t.Fatalf("expected inherited role %q in me.effective_roles, got %+v", roleCode, effNodes)
		}

		// 2. Query.node dengan Global Object ID (`gid://integrione/<resource>/<uuid>`)
		nodeQ := fmt.Sprintf(`query ($u: ID!, $r: ID!, $g: ID!) {
			uNode: node(id: $u) { ... on User { %s } }
			rNode: node(id: $r) { ... on Role { %s permissions(first: 10) { nodes { code } total_count } } }
			gNode: node(id: $g) { ... on Group { %s members(first: 10) { nodes { id } total_count } roles(first: 10) { edges { node { code } } nodes { code } total_count } } }
		}`, gqlUserFields, gqlRoleFields, gqlGroupFields)
		nodeData := doGQLOK(t, admin, nodeQ, map[string]any{"u": emp.ID, "r": roleID, "g": groupID})
		assertUserObject(t, requireOpMap(t, nodeData, "uNode"))
		rNode := requireOpMap(t, nodeData, "rNode")
		assertRoleObject(t, rNode)
		if requireNumberField(t, requireChildMap(t, rNode, "permissions"), "total_count") != 1 {
			t.Fatalf("expected 1 permission on role node, got %+v", rNode["permissions"])
		}
		gNode := requireOpMap(t, nodeData, "gNode")
		assertGroupObject(t, gNode)
		if gNode["code"] != groupCode || requireNumberField(t, requireChildMap(t, gNode, "members"), "total_count") != 1 {
			t.Fatalf("unexpected group node: %+v", gNode)
		}

		// 3. RBAC: customer/non-admin boleh node(self) tetapi wajib ditolak pada node role/admin lain.
		_, custClient := createFixtureCustomerAndLogin(t, admin, cfg)
		requireCustomerForbidden(t, custClient, `query ($id: ID!) { node(id: $id) { ... on Role { id } } }`, map[string]any{"id": roleID})
		requireCustomerForbidden(t, custClient, `query ($id: ID!) { node(id: $id) { ... on User { id } } }`, map[string]any{"id": emp.ID})
	})
}

// TestQueryMyAccessSnapshot menguji Query myAccessSnapshot untuk SuperAdmin, Customer, dan sesi anonim.
func TestQueryMyAccessSnapshot(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	q := fmt.Sprintf(`query { myAccessSnapshot { %s } }`, gqlAccessSnapshotFields)

	t.Run("SuperAdminFullSnapshot", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		data := doGQLOK(t, admin, q, nil)
		snap := requireOpMap(t, data, "myAccessSnapshot")
		assertAccessSnapshotObject(t, snap)
		if !requireBoolField(t, snap, "can_access_identity") {
			t.Fatalf("expected can_access_identity=true")
		}
		if len(requireChildSlice(t, snap, "roles")) == 0 || len(requireChildSlice(t, snap, "permissions")) == 0 {
			t.Fatalf("expected superadmin non-empty roles and permissions")
		}
	})

	t.Run("CustomerSnapshotDeniedIdentity", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		cust, custClient := createFixtureCustomerAndLogin(t, admin, cfg)
		data := doGQLOK(t, custClient, q, nil)
		snap := requireOpMap(t, data, "myAccessSnapshot")
		assertAccessSnapshotObject(t, snap)
		if snap["principal_id"] != cust.RawID || requireBoolField(t, snap, "can_access_identity") {
			t.Fatalf("unexpected customer snapshot: %+v", snap)
		}
	})

	t.Run("AnonymousUnauthenticated", func(t *testing.T) {
		resetGatewayRateLimits(t, cfg)
		requireAnonUnauthenticated(t, newAnonBrowser(t, cfg), q, nil)
	})
}

// TestQueryMySessions menguji Query mySessions dalam format Relay Connection untuk sesi aktif dan riwayat sesi berakhir.
func TestQueryMySessions(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	q := fmt.Sprintf(`query ($active_only: Boolean, $first: Int, $after: String) {
		mySessions(active_only: $active_only, first: $first, after: $after) {
			edges { cursor node { id } }
			nodes { %s }
			pageInfo { %s }
			total_count
		}
	}`, gqlSessionFields, gqlPageInfoFields)

	t.Run("ActiveAndHistoricalSessionsComplete", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		emp := createFixtureUser(t, admin, cfg, "EMPLOYEE", false)

		clientA, _ := loginFixtureUser(t, cfg, emp.Email, emp.Password)
		clientB, _ := loginFixtureUser(t, cfg, emp.Email, emp.Password)
		requireMutationBoolOK(t, doGQLOK(t, clientB, fmt.Sprintf(`mutation { userLogout { %s } }`, gqlUserLogoutFields), nil), "userLogout", "logged_out")

		_, otpURI, _ := enrollAndConfirmTOTP(t, clientA)
		requireMutationBoolOK(t, doGQLOK(t, clientA, fmt.Sprintf(`mutation ($code: String!) { verifyMySessionMfa(totp_code: $code) { %s } }`, gqlVerifySessionMfaFields), map[string]any{
			"code": generateTOTPCode(t, otpURI, time.Now()),
		}), "verifyMySessionMfa", "verified")

		activeData := doGQLOK(t, clientA, q, map[string]any{
			"active_only": true,
			"first":       10,
		})
		activeConn := requireOpMap(t, activeData, "mySessions")
		_ = assertPageInfoNonNull(t, activeConn)
		activeItems := requireConnectionNodes(t, activeConn)
		if len(activeItems) != 1 {
			t.Fatalf("active mySessions count = %d, want 1", len(activeItems))
		}
		activeSess := activeItems[0].(map[string]any)
		assertSessionObject(t, activeSess)
		_ = requireNonEmptyString(t, activeSess, "mfa_verified_at")

		allData := doGQLOK(t, clientA, q, map[string]any{
			"active_only": false,
			"first":       10,
		})
		allConn := requireOpMap(t, allData, "mySessions")
		_ = assertPageInfoNonNull(t, allConn)
		allItems := requireConnectionNodes(t, allConn)
		if len(allItems) < 2 {
			t.Fatalf("historical mySessions count = %d, want >= 2", len(allItems))
		}
		var foundEnded bool
		for _, raw := range allItems {
			s := raw.(map[string]any)
			assertSessionObject(t, s)
			if s["ended_at"] != nil && s["end_reason"] == "LOGOUT" {
				foundEnded = true
			}
		}
		if !foundEnded {
			t.Fatalf("expected ended session with ended_at and end_reason=LOGOUT in %+v", allItems)
		}
	})

	t.Run("AnonymousUnauthenticated", func(t *testing.T) {
		resetGatewayRateLimits(t, cfg)
		requireAnonUnauthenticated(t, newAnonBrowser(t, cfg), q, map[string]any{"active_only": true})
	})
}

// TestQueryUser menguji Query user(id) pada pengguna valid, penolakan RBAC Customer, dan ID tidak ditemukan.
func TestQueryUser(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	q := fmt.Sprintf(`query ($id: ID!) { user(id: $id) { %s } }`, gqlUserFields)

	t.Run("ValidUserWithLoginNonNull", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		emp := createFixtureUser(t, admin, cfg, "EMPLOYEE", false)
		_, _ = loginFixtureUser(t, cfg, emp.Email, emp.Password)

		data := doGQLOK(t, admin, q, map[string]any{"id": emp.ID})
		u := requireOpMap(t, data, "user")
		assertUserObject(t, u)
		_ = requireNonEmptyString(t, u, "last_login_at")
		if u["id"] != emp.ID || u["kind"] != "EMPLOYEE" {
			t.Fatalf("unexpected user object: %+v", u)
		}
	})

	t.Run("RBACCustomerForbidden", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		emp := createFixtureUser(t, admin, cfg, "EMPLOYEE", false)
		_, custClient := createFixtureCustomerAndLogin(t, admin, cfg)
		requireCustomerForbidden(t, custClient, q, map[string]any{"id": emp.ID})
	})

	t.Run("NotFoundError", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		_ = doGQLError(t, admin, q, map[string]any{"id": "00000000-0000-4000-8000-000000000999"}, "NOT_FOUND")
	})
}

// TestQueryUsers menguji Query users dengan filter query, kind, status, include_deleted, pagination, dan RBAC.
func TestQueryUsers(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	q := fmt.Sprintf(`query ($query: String, $kind: UserKind, $status: UserStatus, $include_deleted: Boolean, $first: Int, $after: String) {
		users(query: $query, kind: $kind, status: $status, include_deleted: $include_deleted, first: $first, after: $after) {
			nodes { %s }
			pageInfo { %s }
			total_count
		}
	}`, gqlUserFields, gqlPageInfoFields)

	t.Run("FilterSearchKindStatusPaginationAndDeleted", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		vendor := createFixtureUser(t, admin, cfg, "VENDOR", false)
		emp := createFixtureUser(t, admin, cfg, "EMPLOYEE", false)

		data := doGQLOK(t, admin, q, map[string]any{
			"query":           vendor.Email,
			"kind":            "VENDOR",
			"status":          "ACTIVE",
			"include_deleted": false,
			"first":           10,
		})
		conn := requireOpMap(t, data, "users")
		_ = assertPageInfoNonNull(t, conn)
		if requireNumberField(t, conn, "total_count") != 1 {
			t.Fatalf("filtered users total_count = %v, want 1", conn["total_count"])
		}
		items := requireConnectionNodes(t, conn)
		if len(items) != 1 || items[0].(map[string]any)["id"] != vendor.ID {
			t.Fatalf("unexpected users items: %+v", items)
		}
		assertUserObject(t, items[0].(map[string]any))

		// Uji include_deleted setelah softDeleteUser.
		requireMutationBoolOK(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($id: ID!) { softDeleteUser(user_id: $id) { %s } }`, gqlSoftDeleteUserFields), map[string]any{"id": emp.ID}), "softDeleteUser", "deleted")
		delData := doGQLOK(t, admin, q, map[string]any{
			"query":           emp.Email,
			"include_deleted": true,
			"first":           10,
		})
		delConn := requireOpMap(t, delData, "users")
		_ = assertPageInfoNonNull(t, delConn)
		delItems := requireConnectionNodes(t, delConn)
		if len(delItems) != 1 {
			t.Fatalf("expected 1 deleted user in include_deleted=true, got %d", len(delItems))
		}
		delUser := delItems[0].(map[string]any)
		assertUserObject(t, delUser)
		_ = requireNonEmptyString(t, delUser, "deleted_at")
	})

	t.Run("RBACCustomerForbidden", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		_, custClient := createFixtureCustomerAndLogin(t, admin, cfg)
		requireCustomerForbidden(t, custClient, q, map[string]any{"first": 5})
	})
}

// TestQueryUserIdentities menguji Query userIdentities pada user dengan identitas EMAIL dan USERNAME terverifikasi.
func TestQueryUserIdentities(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	q := fmt.Sprintf(`query ($user_id: ID!, $first: Int) {
		userIdentities(user_id: $user_id, first: $first) {
			nodes { %s }
			pageInfo { %s }
			total_count
		}
	}`, gqlIdentityFields, gqlPageInfoFields)

	t.Run("MultipleVerifiedIdentities", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		emp := createFixtureUser(t, admin, cfg, "EMPLOYEE", true)

		data := doGQLOK(t, admin, q, map[string]any{"user_id": emp.ID, "first": 10})
		conn := requireOpMap(t, data, "userIdentities")
		_ = assertPageInfoNonNull(t, conn)
		items := requireConnectionNodes(t, conn)
		if len(items) != 2 {
			t.Fatalf("userIdentities len = %d, want 2", len(items))
		}
		for _, raw := range items {
			ident := raw.(map[string]any)
			assertUserIdentityObject(t, ident)
			_ = requireNonEmptyString(t, ident, "verified_at")
		}
	})

	t.Run("RBACCustomerForbidden", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		cust, custClient := createFixtureCustomerAndLogin(t, admin, cfg)
		requireCustomerForbidden(t, custClient, q, map[string]any{"user_id": cust.ID, "first": 10})
	})
}

// TestQueryUserRoles menguji Query userRoles termasuk assignment scoped FACILITY + expires_at dan filter include_expired.
func TestQueryUserRoles(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	q := fmt.Sprintf(`query ($user_id: ID!, $include_expired: Boolean, $first: Int) {
		userRoles(user_id: $user_id, include_expired: $include_expired, first: $first) {
			edges { %s }
			pageInfo { %s }
			total_count
		}
	}`, gqlUserRoleEdgeFields, gqlPageInfoFields)

	t.Run("ScopedActiveAndExpiredFilter", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		emp := createFixtureUser(t, admin, cfg, "EMPLOYEE", false)
		roleID, _ := createFixtureRole(t, admin, cfg, "ur")

		scopeID := createFixtureEFFacility(t, admin, cfg)
		futureExp := time.Now().Add(2 * time.Hour).UTC().Format(time.RFC3339)
		assignData := doGQLOK(t, admin, fmt.Sprintf(`mutation ($input: AssignUserRoleInput!) {
			assignUserRole(input: $input) { %s }
		}`, gqlUserRolePayloadFields), map[string]any{
			"input": map[string]any{
				"user_id":    emp.ID,
				"role_id":    roleID,
				"scope_type": "FACILITY",
				"scope_id":   scopeID,
				"expires_at": futureExp,
			},
		})
		assignObj := requireMutationChildOK(t, assignData, "assignUserRole", "user_role_edge")
		assertUserRoleObject(t, assignObj)
		assignID := requireNonEmptyString(t, assignObj, "assignment_id")

		findAssignmentEdge := func(edges []any, targetID string) (map[string]any, bool) {
			for _, raw := range edges {
				edge, ok := raw.(map[string]any)
				if ok && edge["assignment_id"] == targetID {
					return edge, true
				}
			}
			return nil, false
		}

		activeData := doGQLOK(t, admin, q, map[string]any{"user_id": emp.ID, "include_expired": false, "first": 10})
		activeConn := requireOpMap(t, activeData, "userRoles")
		_ = assertPageInfoNonNull(t, activeConn)
		items := requireConnectionEdges(t, activeConn)
		ur, found := findAssignmentEdge(items, assignID)
		if !found {
			t.Fatalf("expected scoped assignment %q in active userRoles, got %+v", assignID, items)
		}
		assertUserRoleObject(t, ur)
		_ = requireNonEmptyString(t, ur, "granted_by")
		_ = requireNonEmptyString(t, ur, "expires_at")
		scopeMap := requireChildMap(t, ur, "scope")
		_ = requireNonEmptyString(t, scopeMap, "id")

		// Mundurkan expires_at di DB lalu cek include_expired: false vs true (peran baseline non-expiring iam_selfservice tetap aktif).
		psqlExec(t, cfg, fmt.Sprintf(`UPDATE user_roles SET expires_at = now() - interval '1 hour' WHERE id = %s::uuid;`, sqlQuote(rawUUID(assignID))))

		noExpConn := requireOpMap(t, doGQLOK(t, admin, q, map[string]any{"user_id": emp.ID, "include_expired": false, "first": 10}), "userRoles")
		_ = assertPageInfoNonNull(t, noExpConn)
		if _, stillActive := findAssignmentEdge(requireConnectionEdges(t, noExpConn), assignID); stillActive {
			t.Fatalf("expected expired assignment %q to be excluded when include_expired=false", assignID)
		}

		withExpConn := requireOpMap(t, doGQLOK(t, admin, q, map[string]any{"user_id": emp.ID, "include_expired": true, "first": 10}), "userRoles")
		_ = assertPageInfoNonNull(t, withExpConn)
		expItems := requireConnectionEdges(t, withExpConn)
		expEdge, reappeared := findAssignmentEdge(expItems, assignID)
		if !reappeared {
			t.Fatalf("expected expired assignment %q with include_expired=true, got %+v", assignID, expItems)
		}
		assertUserRoleObject(t, expEdge)
	})

	t.Run("RBACCustomerForbidden", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		cust, custClient := createFixtureCustomerAndLogin(t, admin, cfg)
		requireCustomerForbidden(t, custClient, q, map[string]any{"user_id": cust.ID, "include_expired": false, "first": 10})
	})
}

// TestQueryUserEffectiveRoles menguji Query userEffectiveRoles untuk gabungan peran langsung dan warisan grup.
func TestQueryUserEffectiveRoles(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	q := fmt.Sprintf(`query ($user_id: ID!, $first: Int) {
		userEffectiveRoles(user_id: $user_id, first: $first) {
			nodes { %s }
			pageInfo { %s }
			total_count
		}
	}`, gqlRoleFields, gqlPageInfoFields)

	t.Run("DirectPlusGroupInheritedRoles", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		emp := createFixtureUser(t, admin, cfg, "EMPLOYEE", false)
		directRoleID, directCode := createFixtureRole(t, admin, cfg, "eff_direct")
		groupRoleID, groupCode := createFixtureRole(t, admin, cfg, "eff_grp")
		groupID, _ := createFixtureGroup(t, admin, cfg, "eff")

		urData := doGQLOK(t, admin, fmt.Sprintf(`mutation ($input: AssignUserRoleInput!) { assignUserRole(input: $input) { %s } }`, gqlUserRolePayloadFields), map[string]any{
			"input": map[string]any{"user_id": emp.ID, "role_id": directRoleID, "scope_type": "GLOBAL"},
		})
		assertUserRoleObject(t, requireMutationChildOK(t, urData, "assignUserRole", "user_role_edge"))

		gmData := doGQLOK(t, admin, fmt.Sprintf(`mutation ($g: ID!, $u: ID!) { addGroupMember(group_id: $g, user_id: $u) { %s } }`, gqlGroupMemberPayloadFields), map[string]any{
			"g": groupID, "u": emp.ID,
		})
		assertGroupMemberObject(t, requireMutationChildOK(t, gmData, "addGroupMember", "group_member_edge"))

		grData := doGQLOK(t, admin, fmt.Sprintf(`mutation ($input: AssignGroupRoleInput!) { assignGroupRole(input: $input) { %s } }`, gqlGroupRolePayloadFields), map[string]any{
			"input": map[string]any{"group_id": groupID, "role_id": groupRoleID, "scope_type": "GLOBAL"},
		})
		assertGroupRoleObject(t, requireMutationChildOK(t, grData, "assignGroupRole", "group_role_edge"))

		data := doGQLOK(t, admin, q, map[string]any{"user_id": emp.ID, "first": 20})
		conn := requireOpMap(t, data, "userEffectiveRoles")
		_ = assertPageInfoNonNull(t, conn)
		roles := requireConnectionNodes(t, conn)
		seen := map[string]bool{}
		for _, raw := range roles {
			r := raw.(map[string]any)
			assertRoleObject(t, r)
			seen[requireNonEmptyString(t, r, "code")] = true
		}
		if !seen[directCode] || !seen[groupCode] {
			t.Fatalf("missing effective roles; seen=%+v", seen)
		}
	})

	t.Run("RBACCustomerForbidden", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		cust, custClient := createFixtureCustomerAndLogin(t, admin, cfg)
		requireCustomerForbidden(t, custClient, q, map[string]any{"user_id": cust.ID, "first": 10})
	})
}

// TestQueryUserGroups menguji Query userGroups untuk user dengan keanggotaan grup maupun user tanpa grup.
func TestQueryUserGroups(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	q := fmt.Sprintf(`query ($user_id: ID!, $first: Int) {
		userGroups(user_id: $user_id, first: $first) {
			nodes { %s }
			pageInfo { %s }
			total_count
		}
	}`, gqlGroupFields, gqlPageInfoFields)

	t.Run("MemberAndEmptySliceNonNull", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		emp := createFixtureUser(t, admin, cfg, "EMPLOYEE", false)
		groupID, groupCode := createFixtureGroup(t, admin, cfg, "ug")

		emptyData := doGQLOK(t, admin, q, map[string]any{"user_id": emp.ID, "first": 10})
		emptyConn := requireOpMap(t, emptyData, "userGroups")
		_ = assertPageInfoNonNull(t, emptyConn)
		emptySlice := requireConnectionNodes(t, emptyConn)
		if len(emptySlice) != 0 {
			t.Fatalf("expected empty non-null array, got %+v", emptySlice)
		}

		gmData := doGQLOK(t, admin, fmt.Sprintf(`mutation ($g: ID!, $u: ID!) { addGroupMember(group_id: $g, user_id: $u) { %s } }`, gqlGroupMemberPayloadFields), map[string]any{
			"g": groupID, "u": emp.ID,
		})
		assertGroupMemberObject(t, requireMutationChildOK(t, gmData, "addGroupMember", "group_member_edge"))

		data := doGQLOK(t, admin, q, map[string]any{"user_id": emp.ID, "first": 10})
		conn := requireOpMap(t, data, "userGroups")
		_ = assertPageInfoNonNull(t, conn)
		groups := requireConnectionNodes(t, conn)
		if len(groups) != 1 {
			t.Fatalf("userGroups len = %d, want 1", len(groups))
		}
		g := groups[0].(map[string]any)
		assertGroupObject(t, g)
		if g["code"] != groupCode {
			t.Fatalf("group code = %v, want %q", g["code"], groupCode)
		}
	})

	t.Run("RBACCustomerForbidden", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		cust, custClient := createFixtureCustomerAndLogin(t, admin, cfg)
		requireCustomerForbidden(t, custClient, q, map[string]any{"user_id": cust.ID, "first": 10})
	})
}

// TestQueryUserSessions menguji Query userSessions oleh admin untuk sesi aktif maupun yang sudah berakhir.
func TestQueryUserSessions(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	q := fmt.Sprintf(`query ($user_id: ID!, $active_only: Boolean, $first: Int) {
		userSessions(user_id: $user_id, active_only: $active_only, first: $first) {
			nodes { %s }
			pageInfo { %s }
			total_count
		}
	}`, gqlSessionFields, gqlPageInfoFields)

	t.Run("AdminInspectsUserActiveAndEndedSessions", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		emp := createFixtureUser(t, admin, cfg, "EMPLOYEE", false)
		c1, _ := loginFixtureUser(t, cfg, emp.Email, emp.Password)
		c2, _ := loginFixtureUser(t, cfg, emp.Email, emp.Password)
		requireMutationBoolOK(t, doGQLOK(t, c2, fmt.Sprintf(`mutation { userLogout { %s } }`, gqlUserLogoutFields), nil), "userLogout", "logged_out")
		_ = c1

		activeData := doGQLOK(t, admin, q, map[string]any{
			"user_id":     emp.ID,
			"active_only": true,
			"first":       10,
		})
		activeConn := requireOpMap(t, activeData, "userSessions")
		_ = assertPageInfoNonNull(t, activeConn)
		activeItems := requireConnectionNodes(t, activeConn)
		if len(activeItems) != 1 {
			t.Fatalf("active userSessions len = %d, want 1", len(activeItems))
		}
		assertSessionObject(t, activeItems[0].(map[string]any))

		allData := doGQLOK(t, admin, q, map[string]any{
			"user_id":     emp.ID,
			"active_only": false,
			"first":       10,
		})
		allConn := requireOpMap(t, allData, "userSessions")
		_ = assertPageInfoNonNull(t, allConn)
		allItems := requireConnectionNodes(t, allConn)
		if len(allItems) < 2 {
			t.Fatalf("expected >= 2 total sessions")
		}
		for _, raw := range allItems {
			assertSessionObject(t, raw.(map[string]any))
		}
	})

	t.Run("RBACCustomerForbidden", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		cust, custClient := createFixtureCustomerAndLogin(t, admin, cfg)
		requireCustomerForbidden(t, custClient, q, map[string]any{"user_id": cust.ID, "active_only": true, "first": 10})
	})
}

// TestQueryRoles menguji Query roles dengan filter query, management_domain, active_only, pagination Relay, dan RBAC.
func TestQueryRoles(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	q := fmt.Sprintf(`query ($query: String, $management_domain: String, $active_only: Boolean, $first: Int) {
		roles(query: $query, management_domain: $management_domain, active_only: $active_only, first: $first) {
			nodes { %s }
			pageInfo { %s }
			total_count
		}
	}`, gqlRoleFields, gqlPageInfoFields)

	t.Run("FilterActiveAndInactiveCustomRoles", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		_, activeCode := createFixtureRole(t, admin, cfg, "rq_act")
		inactID, inactCode := createFixtureRole(t, admin, cfg, "rq_inact")
		deactData := doGQLOK(t, admin, fmt.Sprintf(`mutation ($id: ID!) { deactivateRole(id: $id) { %s } }`, gqlRolePayloadFields), map[string]any{"id": inactID})
		assertRoleObject(t, requireMutationChildOK(t, deactData, "deactivateRole", "role"))

		data := doGQLOK(t, admin, q, map[string]any{
			"query":             activeCode,
			"management_domain": "iam",
			"active_only":       true,
			"first":             10,
		})
		conn := requireOpMap(t, data, "roles")
		_ = assertPageInfoNonNull(t, conn)
		items := requireConnectionNodes(t, conn)
		if len(items) != 1 || items[0].(map[string]any)["code"] != activeCode {
			t.Fatalf("unexpected roles filter items: %+v", items)
		}
		assertRoleObject(t, items[0].(map[string]any))

		hiddenData := doGQLOK(t, admin, q, map[string]any{
			"query":       inactCode,
			"active_only": true,
			"first":       10,
		})
		hiddenConn := requireOpMap(t, hiddenData, "roles")
		_ = assertPageInfoNonNull(t, hiddenConn)
		if len(requireConnectionNodes(t, hiddenConn)) != 0 {
			t.Fatalf("expected inactive role excluded when active_only=true")
		}
	})

	t.Run("RBACCustomerForbidden", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		_, custClient := createFixtureCustomerAndLogin(t, admin, cfg)
		requireCustomerForbidden(t, custClient, q, map[string]any{"first": 5})
	})
}

// TestQueryRole menguji Query role(id) pada role kustom valid serta RBAC Customer.
func TestQueryRole(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	q := fmt.Sprintf(`query ($id: ID!) { role(id: $id) { %s } }`, gqlRoleFields)

	t.Run("ValidRoleNonNull", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		roleID, roleCode := createFixtureRole(t, admin, cfg, "single")
		data := doGQLOK(t, admin, q, map[string]any{"id": roleID})
		role := requireOpMap(t, data, "role")
		assertRoleObject(t, role)
		if role["code"] != roleCode {
			t.Fatalf("role code = %v, want %q", role["code"], roleCode)
		}
	})

	t.Run("RBACCustomerForbidden", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		roleID, _ := createFixtureRole(t, admin, cfg, "rbac")
		_, custClient := createFixtureCustomerAndLogin(t, admin, cfg)
		requireCustomerForbidden(t, custClient, q, map[string]any{"id": roleID})
	})
}

// TestQueryRoleByCode menguji Query roleByCode untuk role sistem (superadmin) dan role kustom.
func TestQueryRoleByCode(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	q := fmt.Sprintf(`query ($code: String!) { roleByCode(code: $code) { %s } }`, gqlRoleFields)

	t.Run("SystemAndCustomRoleByCode", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		sysData := doGQLOK(t, admin, q, map[string]any{"code": "superadmin"})
		sysRole := requireOpMap(t, sysData, "roleByCode")
		assertRoleObject(t, sysRole)
		if !requireBoolField(t, sysRole, "is_system") {
			t.Fatalf("expected superadmin is_system=true")
		}

		_, customCode := createFixtureRole(t, admin, cfg, "bycode")
		custData := doGQLOK(t, admin, q, map[string]any{"code": customCode})
		custRole := requireOpMap(t, custData, "roleByCode")
		assertRoleObject(t, custRole)
		if custRole["code"] != customCode {
			t.Fatalf("custom roleByCode = %+v", custRole)
		}
	})

	t.Run("RBACCustomerForbidden", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		_, custClient := createFixtureCustomerAndLogin(t, admin, cfg)
		requireCustomerForbidden(t, custClient, q, map[string]any{"code": "superadmin"})
	})
}

// TestQueryRolePermissions menguji Query rolePermissions untuk role dengan izin terikat dan role kosong.
func TestQueryRolePermissions(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	q := fmt.Sprintf(`query ($role_id: ID!, $first: Int) {
		rolePermissions(role_id: $role_id, first: $first) {
			nodes { %s }
			edges { %s }
			pageInfo { %s }
			total_count
		}
	}`, gqlPermissionFields, gqlRolePermissionEdgeFields, gqlPageInfoFields)

	t.Run("AssignedAndUnassignedPermissionsNonNull", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		roleID, _ := createFixtureRole(t, admin, cfg, "rperms")
		permID := findPermissionIDByCode(t, admin, "iam.users.read")

		emptyData := doGQLOK(t, admin, q, map[string]any{"role_id": roleID, "first": 10})
		emptyConn := requireOpMap(t, emptyData, "rolePermissions")
		_ = assertPageInfoNonNull(t, emptyConn)
		if len(requireConnectionNodes(t, emptyConn)) != 0 {
			t.Fatalf("expected 0 initial permissions")
		}

		rpData := doGQLOK(t, admin, fmt.Sprintf(`mutation ($r: ID!, $p: ID!) { assignRolePermission(role_id: $r, permission_id: $p) { %s } }`, gqlRolePermPayloadFields), map[string]any{
			"r": roleID, "p": permID,
		})
		assertRolePermissionObject(t, requireMutationChildOK(t, rpData, "assignRolePermission", "role_permission_edge"))

		data := doGQLOK(t, admin, q, map[string]any{"role_id": roleID, "first": 10})
		conn := requireOpMap(t, data, "rolePermissions")
		_ = assertPageInfoNonNull(t, conn)
		perms := requireConnectionNodes(t, conn)
		if len(perms) != 1 {
			t.Fatalf("rolePermissions len = %d, want 1", len(perms))
		}
		p := perms[0].(map[string]any)
		assertPermissionObject(t, p)
		if p["code"] != "iam.users.read" {
			t.Fatalf("unexpected permission: %+v", p)
		}
		edges := requireConnectionEdges(t, conn)
		if len(edges) != 1 {
			t.Fatalf("rolePermissions edges len = %d, want 1", len(edges))
		}
		assertRolePermissionObject(t, edges[0].(map[string]any))
	})

	t.Run("RBACCustomerForbidden", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		roleID, _ := createFixtureRole(t, admin, cfg, "rperms_rbac")
		_, custClient := createFixtureCustomerAndLogin(t, admin, cfg)
		requireCustomerForbidden(t, custClient, q, map[string]any{"role_id": roleID, "first": 10})
	})
}

// TestQueryPermissions menguji Query permissions dengan filter query, resource, is_sensitive, pagination Relay, dan RBAC.
func TestQueryPermissions(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	q := fmt.Sprintf(`query ($query: String, $resource: String, $is_sensitive: Boolean, $first: Int) {
		permissions(query: $query, resource: $resource, is_sensitive: $is_sensitive, first: $first) {
			nodes { %s }
			pageInfo { %s }
			total_count
		}
	}`, gqlPermissionFields, gqlPageInfoFields)

	t.Run("FilterByResourceSensitiveAndPagination", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		data := doGQLOK(t, admin, q, map[string]any{
			"query":        "iam.",
			"resource":     "iam.users",
			"is_sensitive": true,
			"first":        10,
		})
		conn := requireOpMap(t, data, "permissions")
		_ = assertPageInfoNonNull(t, conn)
		if requireNumberField(t, conn, "total_count") == 0 {
			t.Fatalf("expected >= 1 sensitive iam permission")
		}
		items := requireConnectionNodes(t, conn)
		for _, raw := range items {
			p := raw.(map[string]any)
			assertPermissionObject(t, p)
			if !requireBoolField(t, p, "is_sensitive") {
				t.Fatalf("expected is_sensitive=true for item: %+v", p)
			}
		}
	})

	t.Run("RBACCustomerForbidden", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		_, custClient := createFixtureCustomerAndLogin(t, admin, cfg)
		requireCustomerForbidden(t, custClient, q, map[string]any{"first": 5})
	})
}

// TestQueryPermission menguji Query permission(id) pada izin valid dan RBAC Customer.
func TestQueryPermission(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	q := fmt.Sprintf(`query ($id: ID!) { permission(id: $id) { %s } }`, gqlPermissionFields)

	t.Run("ValidPermissionByID", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		permID := findPermissionIDByCode(t, admin, "iam.users.read")
		data := doGQLOK(t, admin, q, map[string]any{"id": permID})
		p := requireOpMap(t, data, "permission")
		assertPermissionObject(t, p)
		if p["code"] != "iam.users.read" {
			t.Fatalf("permission code = %v, want iam.users.read", p["code"])
		}
	})

	t.Run("RBACCustomerForbidden", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		permID := findPermissionIDByCode(t, admin, "iam.users.read")
		_, custClient := createFixtureCustomerAndLogin(t, admin, cfg)
		requireCustomerForbidden(t, custClient, q, map[string]any{"id": permID})
	})
}

// TestQueryGroups menguji Query groups dengan filter query, active_only, pagination, dan RBAC.
func TestQueryGroups(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	q := fmt.Sprintf(`query ($query: String, $active_only: Boolean, $first: Int) {
		groups(query: $query, active_only: $active_only, first: $first) {
			nodes { %s }
			pageInfo { %s }
			total_count
		}
	}`, gqlGroupFields, gqlPageInfoFields)

	t.Run("FilterActiveAndInactiveGroups", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		_, activeCode := createFixtureGroup(t, admin, cfg, "act")
		inactID, inactCode := createFixtureGroup(t, admin, cfg, "inact")
		deactData := doGQLOK(t, admin, fmt.Sprintf(`mutation ($id: ID!) { deactivateGroup(id: $id) { %s } }`, gqlGroupPayloadFields), map[string]any{"id": inactID})
		assertGroupObject(t, requireMutationChildOK(t, deactData, "deactivateGroup", "group"))

		data := doGQLOK(t, admin, q, map[string]any{
			"query":       activeCode,
			"active_only": true,
			"first":       10,
		})
		conn := requireOpMap(t, data, "groups")
		_ = assertPageInfoNonNull(t, conn)
		items := requireConnectionNodes(t, conn)
		if len(items) != 1 || items[0].(map[string]any)["code"] != activeCode {
			t.Fatalf("unexpected groups items: %+v", items)
		}
		assertGroupObject(t, items[0].(map[string]any))

		allData := doGQLOK(t, admin, q, map[string]any{
			"query":       inactCode,
			"active_only": false,
			"first":       10,
		})
		allConn := requireOpMap(t, allData, "groups")
		_ = assertPageInfoNonNull(t, allConn)
		allItems := requireConnectionNodes(t, allConn)
		if len(allItems) != 1 {
			t.Fatalf("expected inactive group when active_only=false")
		}
		assertGroupObject(t, allItems[0].(map[string]any))
	})

	t.Run("RBACCustomerForbidden", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		_, custClient := createFixtureCustomerAndLogin(t, admin, cfg)
		requireCustomerForbidden(t, custClient, q, map[string]any{"first": 5})
	})
}

// TestQueryGroup menguji Query group(id) pada grup valid dan RBAC Customer.
func TestQueryGroup(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	q := fmt.Sprintf(`query ($id: ID!) { group(id: $id) { %s } }`, gqlGroupFields)

	t.Run("ValidGroupNonNull", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		groupID, groupCode := createFixtureGroup(t, admin, cfg, "single")
		data := doGQLOK(t, admin, q, map[string]any{"id": groupID})
		grp := requireOpMap(t, data, "group")
		assertGroupObject(t, grp)
		if grp["code"] != groupCode {
			t.Fatalf("group code = %v, want %q", grp["code"], groupCode)
		}
	})

	t.Run("RBACCustomerForbidden", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		groupID, _ := createFixtureGroup(t, admin, cfg, "rbac")
		_, custClient := createFixtureCustomerAndLogin(t, admin, cfg)
		requireCustomerForbidden(t, custClient, q, map[string]any{"id": groupID})
	})
}

// TestQueryGroupMembers menguji Query groupMembers beserta pagination dan RBAC Customer.
func TestQueryGroupMembers(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	q := fmt.Sprintf(`query ($group_id: ID!, $first: Int, $after: String) {
		groupMembers(group_id: $group_id, first: $first, after: $after) {
			edges { %s }
			nodes { %s }
			pageInfo { %s }
			total_count
		}
	}`, gqlGroupMemberEdgeFields, gqlUserFields, gqlPageInfoFields)

	t.Run("PaginatedMembersNonNull", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		groupID, _ := createFixtureGroup(t, admin, cfg, "members")
		u1 := createFixtureUser(t, admin, cfg, "EMPLOYEE", false)
		u2 := createFixtureUser(t, admin, cfg, "EMPLOYEE", false)

		for _, uid := range []string{u1.ID, u2.ID} {
			gmData := doGQLOK(t, admin, fmt.Sprintf(`mutation ($g: ID!, $u: ID!) { addGroupMember(group_id: $g, user_id: $u) { %s } }`, gqlGroupMemberPayloadFields), map[string]any{
				"g": groupID, "u": uid,
			})
			assertGroupMemberObject(t, requireMutationChildOK(t, gmData, "addGroupMember", "group_member_edge"))
		}

		p1 := doGQLOK(t, admin, q, map[string]any{"group_id": groupID, "first": 1})
		conn1 := requireOpMap(t, p1, "groupMembers")
		page1 := assertPageInfoNonNull(t, conn1)
		if requireNumberField(t, conn1, "total_count") != 2 || !requireBoolField(t, page1, "hasNextPage") {
			t.Fatalf("unexpected page1 info: total=%v pageInfo=%+v", conn1["total_count"], page1)
		}
		items := requireConnectionNodes(t, conn1)
		if len(items) != 1 {
			t.Fatalf("expected 1 member on page 1")
		}
		assertUserObject(t, items[0].(map[string]any))

		// Uji paginasi forward cursor dengan after=endCursor.
		endCursor := requireNonEmptyString(t, page1, "endCursor")
		p2 := doGQLOK(t, admin, q, map[string]any{"group_id": groupID, "first": 1, "after": endCursor})
		conn2 := requireOpMap(t, p2, "groupMembers")
		page2 := assertPageInfoNonNull(t, conn2)
		if !requireBoolField(t, page2, "hasPreviousPage") || len(requireConnectionNodes(t, conn2)) != 1 {
			t.Fatalf("unexpected page2 Relay state: %+v", conn2)
		}
	})

	t.Run("RBACCustomerForbidden", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		groupID, _ := createFixtureGroup(t, admin, cfg, "members_rbac")
		_, custClient := createFixtureCustomerAndLogin(t, admin, cfg)
		requireCustomerForbidden(t, custClient, q, map[string]any{"group_id": groupID, "first": 5})
	})
}

// TestQueryGroupRoles menguji Query groupRoles termasuk scope FACILITY + expires_at dan filter include_expired.
func TestQueryGroupRoles(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	q := fmt.Sprintf(`query ($group_id: ID!, $include_expired: Boolean, $first: Int) {
		groupRoles(group_id: $group_id, include_expired: $include_expired, first: $first) {
			edges { %s }
			pageInfo { %s }
			total_count
		}
	}`, gqlGroupRoleEdgeFields, gqlPageInfoFields)

	t.Run("ScopedActiveAndExpiredGroupRoles", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		groupID, _ := createFixtureGroup(t, admin, cfg, "gr")
		roleID, _ := createFixtureRole(t, admin, cfg, "gr")
		scopeID := createFixtureEFFacility(t, admin, cfg)
		futureExp := time.Now().Add(2 * time.Hour).UTC().Format(time.RFC3339)

		assignData := doGQLOK(t, admin, fmt.Sprintf(`mutation ($input: AssignGroupRoleInput!) {
			assignGroupRole(input: $input) { %s }
		}`, gqlGroupRolePayloadFields), map[string]any{
			"input": map[string]any{
				"group_id":   groupID,
				"role_id":    roleID,
				"scope_type": "FACILITY",
				"scope_id":   scopeID,
				"expires_at": futureExp,
			},
		})
		assignObj := requireMutationChildOK(t, assignData, "assignGroupRole", "group_role_edge")
		assertGroupRoleObject(t, assignObj)
		assignID := requireNonEmptyString(t, assignObj, "assignment_id")

		conn := requireOpMap(t, doGQLOK(t, admin, q, map[string]any{"group_id": groupID, "include_expired": false, "first": 10}), "groupRoles")
		_ = assertPageInfoNonNull(t, conn)
		items := requireConnectionEdges(t, conn)
		if len(items) != 1 {
			t.Fatalf("groupRoles len = %d, want 1", len(items))
		}
		gr := items[0].(map[string]any)
		assertGroupRoleObject(t, gr)
		_ = requireNonEmptyString(t, gr, "granted_by")
		_ = requireNonEmptyString(t, gr, "expires_at")
		_ = requireNonEmptyString(t, requireChildMap(t, gr, "scope"), "id")

		psqlExec(t, cfg, fmt.Sprintf(`UPDATE group_roles SET expires_at = now() - interval '1 hour' WHERE id = %s::uuid;`, sqlQuote(rawUUID(assignID))))

		noExpConn := requireOpMap(t, doGQLOK(t, admin, q, map[string]any{"group_id": groupID, "include_expired": false, "first": 10}), "groupRoles")
		_ = assertPageInfoNonNull(t, noExpConn)
		if len(requireConnectionEdges(t, noExpConn)) != 0 {
			t.Fatalf("expected 0 non-expired groupRoles")
		}
		withExpConn := requireOpMap(t, doGQLOK(t, admin, q, map[string]any{"group_id": groupID, "include_expired": true, "first": 10}), "groupRoles")
		_ = assertPageInfoNonNull(t, withExpConn)
		expItems := requireConnectionEdges(t, withExpConn)
		if len(expItems) != 1 {
			t.Fatalf("expected 1 groupRole with include_expired=true")
		}
		assertGroupRoleObject(t, expItems[0].(map[string]any))
	})

	t.Run("RBACCustomerForbidden", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		groupID, _ := createFixtureGroup(t, admin, cfg, "gr_rbac")
		_, custClient := createFixtureCustomerAndLogin(t, admin, cfg)
		requireCustomerForbidden(t, custClient, q, map[string]any{"group_id": groupID, "include_expired": false, "first": 10})
	})
}

// ============================================================================
// BAGIAN B: 45 MUTATION IAM GRAPHQL (TestMutation*)
// ============================================================================

// TestMutationUserLogin menguji Mutation userLogin untuk login email, username, MFA_REQUIRED, dan INVALID_CREDENTIALS.
func TestMutationUserLogin(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	m := fmt.Sprintf(`mutation ($input: UserLoginInput!) {
		userLogin(input: $input) { %s }
	}`, gqlUserLoginFields)

	t.Run("EmailAndUsernameLoginSuccess", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		emp := createFixtureUser(t, admin, cfg, "EMPLOYEE", true)

		c1 := newAnonBrowser(t, cfg)
		d1 := doGQLOK(t, c1, m, map[string]any{"input": map[string]any{"identifier": emp.Email, "password": emp.Password}})
		p1 := requireOpMap(t, d1, "userLogin")
		assertUserLoginPayload(t, p1, "AUTHENTICATED")
		t.Cleanup(func() { _, _, _ = logoutBrowserBestEffort(c1) })

		c2 := newAnonBrowser(t, cfg)
		d2 := doGQLOK(t, c2, m, map[string]any{"input": map[string]any{"identifier": emp.Username, "password": emp.Password}})
		p2 := requireOpMap(t, d2, "userLogin")
		assertUserLoginPayload(t, p2, "AUTHENTICATED")
		t.Cleanup(func() { _, _, _ = logoutBrowserBestEffort(c2) })
	})

	t.Run("MfaRequiredAndInvalidCredentials", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		emp := createFixtureUser(t, admin, cfg, "EMPLOYEE", false)
		cEnroll, _ := loginFixtureUser(t, cfg, emp.Email, emp.Password)
		_, _, _ = enrollAndConfirmTOTP(t, cEnroll)

		cBad := newAnonBrowser(t, cfg)
		_, _ = doMutationFail(t, cBad, m, map[string]any{"input": map[string]any{"identifier": emp.Email, "password": "WrongPassword#1"}}, "userLogin", "AUTHENTICATION")

		cMFA := newAnonBrowser(t, cfg)
		dMFA := doGQLOK(t, cMFA, m, map[string]any{"input": map[string]any{"identifier": emp.Email, "password": emp.Password}})
		pMFA := requireOpMap(t, dMFA, "userLogin")
		assertUserLoginPayload(t, pMFA, "MFA_REQUIRED")
	})
}

// TestMutationValidateUserMfa menguji Mutation validateUserMfa dengan TOTP, recovery code, dan penolakan replay recovery code.
func TestMutationValidateUserMfa(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	mLogin := fmt.Sprintf(`mutation ($input: UserLoginInput!) {
		userLogin(input: $input) { %s }
	}`, gqlUserLoginFields)
	mMFA := fmt.Sprintf(`mutation ($input: ValidateUserMfaInput!) {
		validateUserMfa(input: $input) { %s }
	}`, gqlValidateUserMfaFields)

	t.Run("TotpAndRecoveryCodeWithReplayProtection", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		emp := createFixtureUser(t, admin, cfg, "EMPLOYEE", false)
		cEnroll, _ := loginFixtureUser(t, cfg, emp.Email, emp.Password)
		_, otpURI, recCodes := enrollAndConfirmTOTP(t, cEnroll)

		// Jalur 1: Verifikasi MFA menggunakan kode TOTP 6 digit.
		cTOTP := newAnonBrowser(t, cfg)
		login1 := requireOpMap(t, doGQLOK(t, cTOTP, mLogin, map[string]any{"input": map[string]any{"identifier": emp.Email, "password": emp.Password}}), "userLogin")
		assertUserLoginPayload(t, login1, "MFA_REQUIRED")
		ch1 := requireNonEmptyString(t, login1, "mfaChallenge")
		dTOTP := doGQLOK(t, cTOTP, mMFA, map[string]any{"input": map[string]any{
			"challenge": ch1,
			"totpCode":  generateTOTPCode(t, otpURI, time.Now()),
		}})
		pTOTP := requireOpMap(t, dTOTP, "validateUserMfa")
		assertValidateUserMfaPayload(t, pTOTP)
		t.Cleanup(func() { _, _, _ = logoutBrowserBestEffort(cTOTP) })

		// Jalur 2: Verifikasi MFA menggunakan recovery code satu kali pakai.
		cRec := newAnonBrowser(t, cfg)
		login2 := requireOpMap(t, doGQLOK(t, cRec, mLogin, map[string]any{"input": map[string]any{"identifier": emp.Email, "password": emp.Password}}), "userLogin")
		assertUserLoginPayload(t, login2, "MFA_REQUIRED")
		ch2 := requireNonEmptyString(t, login2, "mfaChallenge")
		dRec := doGQLOK(t, cRec, mMFA, map[string]any{"input": map[string]any{
			"challenge":    ch2,
			"recoveryCode": recCodes[0],
		}})
		pRec := requireOpMap(t, dRec, "validateUserMfa")
		assertValidateUserMfaPayload(t, pRec)
		t.Cleanup(func() { _, _, _ = logoutBrowserBestEffort(cRec) })

		// Jalur 3: Pemakaian ulang recovery code yang sama wajib ditolak.
		cReplay := newAnonBrowser(t, cfg)
		login3 := requireOpMap(t, doGQLOK(t, cReplay, mLogin, map[string]any{"input": map[string]any{"identifier": emp.Email, "password": emp.Password}}), "userLogin")
		assertUserLoginPayload(t, login3, "MFA_REQUIRED")
		ch3 := requireNonEmptyString(t, login3, "mfaChallenge")
		_, _ = doMutationFail(t, cReplay, mMFA, map[string]any{"input": map[string]any{
			"challenge":    ch3,
			"recoveryCode": recCodes[0],
		}}, "validateUserMfa", "AUTHENTICATION")
	})
}

// TestMutationVerifyEmail menguji Mutation verifyEmail untuk kasus nyata VERIFIED, ALREADY_USED, EXPIRED, dan INVALID.
func TestMutationVerifyEmail(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	mReg := fmt.Sprintf(`mutation ($input: CreateUserInput!) { createUser(input: $input) { %s } }`, gqlCreateUserPayloadFields)
	mVerify := fmt.Sprintf(`mutation ($input: VerifyEmailInput!) { verifyEmail(input: $input) { %s } }`, gqlVerifyEmailFields)

	t.Run("VerifiedAlreadyUsedExpiredAndInvalid", func(t *testing.T) {
		resetGatewayRateLimits(t, cfg)
		anon := newAnonBrowser(t, cfg)

		tag1 := uniqueTag("ver1")
		email1 := tag1 + "@integrione.test"
		reg1Data := doGQLOK(t, anon, mReg, map[string]any{"input": map[string]any{
			"full_name": "Verify One", "email": email1, "username": tag1, "password": "Zx9#mQp2!vLk8@Rt",
		}})
		reg1Res := requireMutationSuccess(t, reg1Data, "createUser")
		if reg1Res["user"] != nil {
			t.Fatalf("public createUser returned non-nil user: %+v", reg1Res["user"])
		}
		user1ID := psqlScalar(t, cfg, fmt.Sprintf(`SELECT user_id::text FROM user_identities WHERE kind = 'email' AND value = %s LIMIT 1;`, sqlQuote(email1)))
		t.Cleanup(func() { cleanupUserByID(t, cfg, user1ID) })

		rawToken1 := "e2e-verify-token-" + tag1
		psqlExec(t, cfg, fmt.Sprintf(
			`UPDATE user_tokens SET token_digest = decode(%s, 'hex') WHERE user_id = %s::uuid AND purpose = 'email_verification' AND status = 'active';`,
			sqlQuote(sha256Hex(rawToken1)), sqlQuote(user1ID),
		))

		// 1. Token valid -> VERIFIED + verified: true, lalu user bisa login!
		verData := doGQLOK(t, anon, mVerify, map[string]any{"input": map[string]any{"token": rawToken1}})
		verRes := requireOpMap(t, verData, "verifyEmail")
		assertVerifyEmailPayload(t, verRes, "VERIFIED", true)
		_, loginPayload := loginFixtureUser(t, cfg, email1, "Zx9#mQp2!vLk8@Rt")
		assertUserLoginPayload(t, loginPayload, "AUTHENTICATED")

		// 2. Token terpakai -> ALREADY_USED + verified: false.
		usedRes := requireOpMap(t, doGQLOK(t, anon, mVerify, map[string]any{"input": map[string]any{"token": rawToken1}}), "verifyEmail")
		assertVerifyEmailPayload(t, usedRes, "ALREADY_USED", false)

		// 3. Token kedaluwarsa -> EXPIRED + verified: false.
		tag2 := uniqueTag("ver2")
		email2 := tag2 + "@integrione.test"
		reg2Data := doGQLOK(t, anon, mReg, map[string]any{"input": map[string]any{
			"full_name": "Verify Expired", "email": email2, "username": tag2, "password": "Zx9#mQp2!vLk8@Rt",
		}})
		reg2Res := requireMutationSuccess(t, reg2Data, "createUser")
		if reg2Res["user"] != nil {
			t.Fatalf("public createUser returned non-nil user: %+v", reg2Res["user"])
		}
		user2ID := psqlScalar(t, cfg, fmt.Sprintf(`SELECT user_id::text FROM user_identities WHERE kind = 'email' AND value = %s LIMIT 1;`, sqlQuote(email2)))
		t.Cleanup(func() { cleanupUserByID(t, cfg, user2ID) })

		rawToken2 := "e2e-expired-token-" + tag2
		psqlExec(t, cfg, fmt.Sprintf(
			`UPDATE user_tokens SET token_digest = decode(%s, 'hex'), created_at = now() - interval '48 hours', expires_at = now() - interval '24 hours' WHERE user_id = %s::uuid AND purpose = 'email_verification' AND status = 'active';`,
			sqlQuote(sha256Hex(rawToken2)), sqlQuote(user2ID),
		))
		expRes := requireOpMap(t, doGQLOK(t, anon, mVerify, map[string]any{"input": map[string]any{"token": rawToken2}}), "verifyEmail")
		assertVerifyEmailPayload(t, expRes, "EXPIRED", false)

		// 4. Token acak -> INVALID + verified: false.
		invRes := requireOpMap(t, doGQLOK(t, anon, mVerify, map[string]any{"input": map[string]any{"token": "invalid-nonexistent-token-xyz"}}), "verifyEmail")
		assertVerifyEmailPayload(t, invRes, "INVALID", false)
	})
}

// TestMutationUserResetPassword menguji Mutation userResetPassword berbasis email terverifikasi dan anti-enumerasi akun.
func TestMutationUserResetPassword(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	m := fmt.Sprintf(`mutation ($input: UserResetPasswordInput!) { userResetPassword(input: $input) { %s } }`, gqlUserResetPasswordFields)
	mReg := fmt.Sprintf(`mutation ($input: CreateUserInput!) { createUser(input: $input) { %s } }`, gqlCreateUserPayloadFields)

	t.Run("VerifiedEmailIssuesSingleActiveTokenAndUnverifiedOrGhostReturnsNeutralMessage", func(t *testing.T) {
		resetGatewayRateLimits(t, cfg)
		admin := newSuperAdminBrowser(t, cfg)
		emp := createFixtureUser(t, admin, cfg, "EMPLOYEE", false)
		anon := newAnonBrowser(t, cfg)

		// 1. Akun PENDING (email belum terverifikasi) -> was_successful=true dengan pesan netral, tetapi TIDAK menerbitkan token reset.
		pendTag := uniqueTag("rst_pend")
		pendEmail := pendTag + "@integrione.test"
		pendReg := requireMutationSuccess(t, doGQLOK(t, anon, mReg, map[string]any{"input": map[string]any{
			"full_name": "Pending Reset", "email": pendEmail, "username": pendTag, "password": "Zx9#mQp2!vLk8@Rt",
		}}), "createUser")
		if pendReg["user"] != nil {
			t.Fatalf("public createUser returned non-nil user: %+v", pendReg["user"])
		}
		pendUserID := psqlScalar(t, cfg, fmt.Sprintf(`SELECT user_id::text FROM user_identities WHERE kind = 'email' AND value = %s LIMIT 1;`, sqlQuote(pendEmail)))
		t.Cleanup(func() { cleanupUserByID(t, cfg, pendUserID) })

		pendPayload := requireMutationSuccess(t, doGQLOK(t, anon, m, map[string]any{"input": map[string]any{"email": pendEmail}}), "userResetPassword")
		_ = requireNonEmptyString(t, pendPayload, "message")
		pendCnt := psqlScalar(t, cfg, fmt.Sprintf(`SELECT count(*) FROM user_tokens WHERE user_id = %s::uuid AND purpose = 'password_reset';`, sqlQuote(pendUserID)))
		if pendCnt != "0" {
			t.Fatalf("expected 0 password_reset tokens for unverified email, got %s", pendCnt)
		}

		// 2. Email terverifikasi -> menerbitkan 1 token aktif, dan pemanggilan ulang menandai token lama expired.
		verPayload1 := requireMutationSuccess(t, doGQLOK(t, anon, m, map[string]any{"input": map[string]any{"email": emp.Email}}), "userResetPassword")
		_ = requireNonEmptyString(t, verPayload1, "message")
		activeCnt1 := psqlScalar(t, cfg, fmt.Sprintf(`SELECT count(*) FROM user_tokens WHERE user_id = %s::uuid AND purpose = 'password_reset' AND status = 'active';`, sqlQuote(emp.RawID)))
		if activeCnt1 != "1" {
			t.Fatalf("active password_reset token count = %s, want 1", activeCnt1)
		}

		verPayload2 := requireMutationSuccess(t, doGQLOK(t, anon, m, map[string]any{"input": map[string]any{"email": emp.Email}}), "userResetPassword")
		_ = requireNonEmptyString(t, verPayload2, "message")
		activeCnt2 := psqlScalar(t, cfg, fmt.Sprintf(`SELECT count(*) FROM user_tokens WHERE user_id = %s::uuid AND purpose = 'password_reset' AND status = 'active';`, sqlQuote(emp.RawID)))
		expiredCnt := psqlScalar(t, cfg, fmt.Sprintf(`SELECT count(*) FROM user_tokens WHERE user_id = %s::uuid AND purpose = 'password_reset' AND status = 'expired';`, sqlQuote(emp.RawID)))
		if activeCnt2 != "1" || expiredCnt != "1" {
			t.Fatalf("expected 1 active and 1 expired password_reset token, got active=%s expired=%s", activeCnt2, expiredCnt)
		}

		// 3. Email tidak terdaftar -> tetap sukses dengan pesan netral (anti-enumerasi).
		ghostPayload := requireMutationSuccess(t, doGQLOK(t, anon, m, map[string]any{"input": map[string]any{"email": "ghost-" + uniqueTag("none") + "@integrione.test"}}), "userResetPassword")
		_ = requireNonEmptyString(t, ghostPayload, "message")

		// 4. Email kosong wajib gagal dengan MutationError tipe ARGUMENT.
		_, _ = doMutationFail(t, anon, m, map[string]any{"input": map[string]any{"email": "   "}}, "userResetPassword", "ARGUMENT")
	})
}

// TestMutationResetPassword menguji Mutation resetPassword dengan token reset nyata, verifikasi password baru dapat login, dan token tidak sah.
func TestMutationResetPassword(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	mReq := fmt.Sprintf(`mutation ($input: UserResetPasswordInput!) { userResetPassword(input: $input) { %s } }`, gqlUserResetPasswordFields)
	mReset := fmt.Sprintf(`mutation ($input: IAMResetPasswordInput!) { resetPassword(input: $input) { %s } }`, gqlResetPasswordFields)

	t.Run("ValidTokenUpdatesPasswordAndInvalidRejected", func(t *testing.T) {
		resetGatewayRateLimits(t, cfg)
		admin := newSuperAdminBrowser(t, cfg)
		emp := createFixtureUser(t, admin, cfg, "EMPLOYEE", false)
		anon := newAnonBrowser(t, cfg)

		_ = requireMutationSuccess(t, doGQLOK(t, anon, mReq, map[string]any{"input": map[string]any{"email": emp.Email}}), "userResetPassword")
		rawReset := "e2e-reset-token-" + uniqueTag("rst")
		psqlExec(t, cfg, fmt.Sprintf(
			`UPDATE user_tokens SET token_digest = decode(%s, 'hex') WHERE user_id = %s::uuid AND purpose = 'password_reset' AND status = 'active';`,
			sqlQuote(sha256Hex(rawReset)), sqlQuote(emp.RawID),
		))

		newPass := "NewResetPass#2026!"
		requireMutationBoolOK(t, doGQLOK(t, anon, mReset, map[string]any{"input": map[string]any{
			"token": rawReset, "newPassword": newPass,
		}}), "resetPassword", "reset")

		_, newLogin := loginFixtureUser(t, cfg, emp.Email, newPass)
		assertUserLoginPayload(t, newLogin, "AUTHENTICATED")

		_, _ = doMutationFail(t, anon, mReset, map[string]any{"input": map[string]any{
			"token": "invalid-reset-token-xyz", "newPassword": "AnotherPass#2026!",
		}}, "resetPassword", "NOT_FOUND")
	})
}

// TestMutationUserLogout menguji Mutation userLogout memastikan logged_out=true, pencabutan sesi, serta penolakan sesi anonim.
func TestMutationUserLogout(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	m := fmt.Sprintf(`mutation { userLogout { %s } }`, gqlUserLogoutFields)

	t.Run("AuthenticatedLogoutRevokesSession", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		emp := createFixtureUser(t, admin, cfg, "EMPLOYEE", false)
		empClient, _ := loginFixtureUser(t, cfg, emp.Email, emp.Password)

		requireMutationBoolOK(t, doGQLOK(t, empClient, m, nil), "userLogout", "logged_out")
		_ = doGQLError(t, empClient, fmt.Sprintf(`query { me { %s } }`, gqlUserFields), nil, "UNAUTHENTICATED")
	})

	t.Run("AnonymousRejected", func(t *testing.T) {
		resetGatewayRateLimits(t, cfg)
		requireAnonMutationUnauthenticated(t, newAnonBrowser(t, cfg), m, nil, "userLogout")
	})
}

// TestMutationUpdateMyProfile menguji Mutation updateMyProfile untuk pembaruan data profil non-null, PreserveAvatarUrl, serta penolakan anonim.
func TestMutationUpdateMyProfile(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	m := fmt.Sprintf(`mutation ($input: UpdateMyProfileInput!) { updateMyProfile(input: $input) { %s } }`, gqlUserPayloadFields)

	t.Run("UpdateAllProfileFields", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		emp := createFixtureUser(t, admin, cfg, "EMPLOYEE", false)
		empClient, _ := loginFixtureUser(t, cfg, emp.Email, emp.Password)

		updatedName := "Updated Name " + uniqueTag("prof")
		data := doGQLOK(t, empClient, m, map[string]any{
			"input": map[string]any{
				"display_name": updatedName,
				"locale":       "en-US",
				"timezone":     "UTC",
			},
		})
		user := requireMutationChildOK(t, data, "updateMyProfile", "user")
		assertUserObject(t, user)
		_ = requireNonEmptyString(t, user, "last_login_at")
		if user["display_name"] != updatedName || user["locale"] != "en-US" || user["timezone"] != "UTC" {
			t.Fatalf("unexpected updated user: %+v", user)
		}
	})

	t.Run("AnonymousRejected", func(t *testing.T) {
		resetGatewayRateLimits(t, cfg)
		requireAnonMutationUnauthenticated(t, newAnonBrowser(t, cfg), m, map[string]any{
			"input": map[string]any{"display_name": "Anon"},
		}, "updateMyProfile")
	})
}

// TestMutationUploadAvatarUser menguji Mutation uploadAvatarUser untuk unggah AVIF base64, persistensi URL, PreserveAvatarUrl pada updateMyProfile, dan penolakan non-AVIF.
func TestMutationUploadAvatarUser(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	m := fmt.Sprintf(`mutation ($input: UserAvatarUploadInput!) { uploadAvatarUser(input: $input) { %s } }`, gqlUserAvatarPayloadFields)

	t.Run("ValidAvifStoresUrlAndUpdateProfilePreservesAvatar", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		emp := createFixtureUser(t, admin, cfg, "EMPLOYEE", false)
		empClient, _ := loginFixtureUser(t, cfg, emp.Email, emp.Password)

		// Box ISOBMFF ftyp minimal valid AVIF (16 byte).
		avifBytes := []byte{0, 0, 0, 16, 'f', 't', 'y', 'p', 'a', 'v', 'i', 'f', 0, 0, 0, 0}
		encoded := "data:image/avif;base64," + base64.StdEncoding.EncodeToString(avifBytes)

		res := requireMutationSuccess(t, doGQLOK(t, empClient, m, map[string]any{
			"input": map[string]any{"image": encoded},
		}), "uploadAvatarUser")
		avatarURL := requireNonEmptyString(t, res, "avatar_url")
		if !strings.HasPrefix(avatarURL, "https://") || !strings.HasSuffix(avatarURL, ".avif") {
			t.Fatalf("unexpected avatar_url %q", avatarURL)
		}

		// Pastikan me() dan updateMyProfile() mempertahankan avatar_url yang sudah terunggah.
		meObj := requireOpMap(t, doGQLOK(t, empClient, fmt.Sprintf(`query { me { %s } }`, gqlUserFields), nil), "me")
		if meObj["avatar_url"] != avatarURL {
			t.Fatalf("me.avatar_url = %v, want %q", meObj["avatar_url"], avatarURL)
		}
		updObj := requireMutationChildOK(t, doGQLOK(t, empClient, fmt.Sprintf(`mutation ($in: UpdateMyProfileInput!) { updateMyProfile(input: $in) { %s } }`, gqlUserPayloadFields), map[string]any{
			"in": map[string]any{"display_name": "Avatar Keeper"},
		}), "updateMyProfile", "user")
		if updObj["avatar_url"] != avatarURL {
			t.Fatalf("updateMyProfile did not preserve avatar_url: got %v want %q", updObj["avatar_url"], avatarURL)
		}

		// Tolak payload non-AVIF / data URI bukan image/avif.
		_, _ = doMutationFail(t, empClient, m, map[string]any{
			"input": map[string]any{"image": "data:image/png;base64,iVBORw0KGgo="},
		}, "uploadAvatarUser", "ARGUMENT")
	})

	t.Run("AnonymousRejected", func(t *testing.T) {
		resetGatewayRateLimits(t, cfg)
		avifB64 := base64.StdEncoding.EncodeToString([]byte{0, 0, 0, 16, 'f', 't', 'y', 'p', 'a', 'v', 'i', 'f', 0, 0, 0, 0})
		requireAnonMutationUnauthenticated(t, newAnonBrowser(t, cfg), m, map[string]any{
			"input": map[string]any{"image": avifB64},
		}, "uploadAvatarUser")
	})
}

// TestMutationChangeUserPassword menguji Mutation changeUserPassword untuk ubah kata sandi, pencabutan sesi, wrong current, dan sesi basi >10 menit.
func TestMutationChangeUserPassword(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	m := fmt.Sprintf(`mutation ($input: ChangeUserPasswordInput!) { changeUserPassword(input: $input) { %s } }`, gqlChangeUserPasswordFields)

	t.Run("FreshSuccessWrongCurrentAndStaleDenied", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		emp := createFixtureUser(t, admin, cfg, "EMPLOYEE", false)
		empClient, _ := loginFixtureUser(t, cfg, emp.Email, emp.Password)

		_, _ = doMutationFail(t, empClient, m, map[string]any{
			"input": map[string]any{"current_password": "WrongCurrent#1", "new_password": "NextPass#2026!"},
		}, "changeUserPassword", "AUTHENTICATION")

		newPass := "ChangedPass#2026!"
		chRes := requireMutationBoolOK(t, doGQLOK(t, empClient, m, map[string]any{
			"input": map[string]any{"current_password": emp.Password, "new_password": newPass},
		}), "changeUserPassword", "changed")
		if requireNumberField(t, chRes, "sessions_revoked") < 1 {
			t.Fatalf("expected sessions_revoked >= 1, got %+v", chRes)
		}
		_ = requireNonEmptyString(t, chRes, "password_changed_at")

		// Sesi lama wajib tercabut setelah changeUserPassword berhasil.
		_ = doGQLError(t, empClient, fmt.Sprintf(`query { me { %s } }`, gqlUserFields), nil, "UNAUTHENTICATED")

		// Sesi kedua harus bisa login dengan password baru, dan jika sesi dibuat basi (>10m) maka ditolak AUTHORIZATION.
		staleClient, _ := loginFixtureUser(t, cfg, emp.Email, newPass)
		backdateSessionForStaleSensitiveTest(t, cfg, staleClient, emp.ID)
		_, _ = doMutationFail(t, staleClient, m, map[string]any{
			"input": map[string]any{"current_password": newPass, "new_password": "ThirdPass#2026!"},
		}, "changeUserPassword", "AUTHORIZATION")
	})
}

// TestMutationVerifyMySessionMfa menguji Mutation verifyMySessionMfa untuk rotasi CSRF, set mfa_verified_at, dan kode TOTP salah.
func TestMutationVerifyMySessionMfa(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	m := fmt.Sprintf(`mutation ($code: String!) { verifyMySessionMfa(totp_code: $code) { %s } }`, gqlVerifySessionMfaFields)

	t.Run("ValidTotpRotatesCsrfAndRejectsInvalidCode", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		emp := createFixtureUser(t, admin, cfg, "EMPLOYEE", false)
		empClient, _ := loginFixtureUser(t, cfg, emp.Email, emp.Password)
		_, otpURI, _ := enrollAndConfirmTOTP(t, empClient)

		oldCSRF := empClient.csrfToken
		_, _ = doMutationFail(t, empClient, m, map[string]any{"code": "000000"}, "verifyMySessionMfa", "AUTHENTICATION")

		requireMutationBoolOK(t, doGQLOK(t, empClient, m, map[string]any{"code": generateTOTPCode(t, otpURI, time.Now())}), "verifyMySessionMfa", "verified")
		if empClient.csrfToken == oldCSRF {
			t.Fatalf("expected CSRF token rotation after verifyMySessionMfa")
		}

		// Verifikasi sesi mencatat mfa_verified=true dan mfa_verified_at non-null.
		sessData := doGQLOK(t, empClient, fmt.Sprintf(`query { mySessions(active_only: true, first: 10) { nodes { %s } pageInfo { %s } total_count } }`, gqlSessionFields, gqlPageInfoFields), nil)
		sessConn := requireOpMap(t, sessData, "mySessions")
		_ = assertPageInfoNonNull(t, sessConn)
		items := requireConnectionNodes(t, sessConn)
		sessObj := items[0].(map[string]any)
		assertSessionObject(t, sessObj)
		_ = requireNonEmptyString(t, sessObj, "mfa_verified_at")
		if !requireBoolField(t, sessObj, "mfa_verified") {
			t.Fatalf("expected mfa_verified=true")
		}
	})
}

// TestMutationEnrollMyTotp menguji Mutation enrollMyTotp pada sesi segar dan penolakan pada sesi basi >10 menit.
func TestMutationEnrollMyTotp(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	m := fmt.Sprintf(`mutation { enrollMyTotp { %s } }`, gqlEnrollTotpFields)

	t.Run("FreshSessionSuccessAndStaleForbidden", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		emp := createFixtureUser(t, admin, cfg, "EMPLOYEE", false)
		empClient, _ := loginFixtureUser(t, cfg, emp.Email, emp.Password)

		res := requireMutationSuccess(t, doGQLOK(t, empClient, m, nil), "enrollMyTotp")
		assertEnrollTotpPayload(t, res)
		uri := requireNonEmptyString(t, res, "otp_auth_uri")
		if !strings.HasPrefix(uri, "otpauth://totp/") {
			t.Fatalf("unexpected otp_auth_uri = %q", uri)
		}

		staleUser := createFixtureUser(t, admin, cfg, "EMPLOYEE", false)
		staleClient, _ := loginFixtureUser(t, cfg, staleUser.Email, staleUser.Password)
		backdateSessionForStaleSensitiveTest(t, cfg, staleClient, staleUser.ID)
		_, _ = doMutationFail(t, staleClient, m, nil, "enrollMyTotp", "AUTHORIZATION")
	})
}

// TestMutationConfirmMyTotp menguji Mutation confirmMyTotp untuk aktivasi TOTP + recovery_codes non-null serta penolakan kode salah.
func TestMutationConfirmMyTotp(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	m := fmt.Sprintf(`mutation ($input: ConfirmTotpInput!) { confirmMyTotp(input: $input) { %s } }`, gqlConfirmTotpFields)

	t.Run("ValidCodeEnablesAndInvalidRejected", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		emp := createFixtureUser(t, admin, cfg, "EMPLOYEE", false)
		empClient, _ := loginFixtureUser(t, cfg, emp.Email, emp.Password)

		enroll := requireMutationSuccess(t, doGQLOK(t, empClient, fmt.Sprintf(`mutation { enrollMyTotp { %s } }`, gqlEnrollTotpFields), nil), "enrollMyTotp")
		assertEnrollTotpPayload(t, enroll)
		otpURI := requireNonEmptyString(t, enroll, "otp_auth_uri")

		_, _ = doMutationFail(t, empClient, m, map[string]any{"input": map[string]any{"code": "999999"}}, "confirmMyTotp", "AUTHENTICATION")

		res := requireMutationSuccess(t, doGQLOK(t, empClient, m, map[string]any{
			"input": map[string]any{"code": generateTOTPCode(t, otpURI, time.Now())},
		}), "confirmMyTotp")
		_ = assertConfirmTotpPayload(t, res)
	})
}

// TestMutationDisableMyTotp menguji Mutation disableMyTotp pada sesi step-up terverifikasi dan penolakan pada sesi tanpa MFA step-up.
func TestMutationDisableMyTotp(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	m := fmt.Sprintf(`mutation { disableMyTotp { %s } }`, gqlDisableTotpFields)

	t.Run("RequiresVerifiedSessionAndDisablesCleanly", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		emp := createFixtureUser(t, admin, cfg, "EMPLOYEE", false)
		empClient, _ := loginFixtureUser(t, cfg, emp.Email, emp.Password)
		_, otpURI, _ := enrollAndConfirmTOTP(t, empClient)

		// Tanpa verifyMySessionMfa, sesi belum mfa_authenticated sehingga disableMyTotp wajib AUTHORIZATION.
		_, _ = doMutationFail(t, empClient, m, nil, "disableMyTotp", "AUTHORIZATION")

		requireMutationBoolOK(t, doGQLOK(t, empClient, fmt.Sprintf(`mutation ($code: String!) { verifyMySessionMfa(totp_code: $code) { %s } }`, gqlVerifySessionMfaFields), map[string]any{
			"code": generateTOTPCode(t, otpURI, time.Now()),
		}), "verifyMySessionMfa", "verified")
		requireMutationBoolOK(t, doGQLOK(t, empClient, m, nil), "disableMyTotp", "disabled")

		snap := requireOpMap(t, doGQLOK(t, empClient, fmt.Sprintf(`query { myAccessSnapshot { %s } }`, gqlAccessSnapshotFields), nil), "myAccessSnapshot")
		assertAccessSnapshotObject(t, snap)
		if requireBoolField(t, snap, "mfa_enabled") {
			t.Fatalf("expected mfa_enabled=false after disableMyTotp")
		}
	})
}

// TestMutationRegenerateMyRecoveryCodes menguji Mutation regenerateMyRecoveryCodes dan invalidasi recovery code lama.
func TestMutationRegenerateMyRecoveryCodes(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	m := fmt.Sprintf(`mutation { regenerateMyRecoveryCodes { %s } }`, gqlRegenRecoveryCodesFields)

	t.Run("GeneratesFreshCodesAndInvalidatesPrevious", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		emp := createFixtureUser(t, admin, cfg, "EMPLOYEE", false)
		empClient, _ := loginFixtureUser(t, cfg, emp.Email, emp.Password)
		_, otpURI, oldCodes := enrollAndConfirmTOTP(t, empClient)

		requireMutationBoolOK(t, doGQLOK(t, empClient, fmt.Sprintf(`mutation ($code: String!) { verifyMySessionMfa(totp_code: $code) { %s } }`, gqlVerifySessionMfaFields), map[string]any{
			"code": generateTOTPCode(t, otpURI, time.Now()),
		}), "verifyMySessionMfa", "verified")

		regenPayload := requireMutationSuccess(t, doGQLOK(t, empClient, m, nil), "regenerateMyRecoveryCodes")
		freshRaw := requireChildSlice(t, regenPayload, "recovery_codes")
		if len(freshRaw) == 0 {
			t.Fatalf("expected non-empty fresh recovery codes")
		}
		newCode := freshRaw[0].(string)
		if strings.TrimSpace(newCode) == "" || newCode == oldCodes[0] {
			t.Fatalf("unexpected fresh recovery code %q", newCode)
		}

		// Recovery code lama harus ditolak saat login MFA baru, sedangkan kode baru diterima.
		cTest := newAnonBrowser(t, cfg)
		loginObj := requireMutationSuccess(t, doGQLOK(t, cTest, fmt.Sprintf(`mutation ($in: UserLoginInput!) {
			userLogin(input: $in) { %s }
		}`, gqlUserLoginFields), map[string]any{"in": map[string]any{"identifier": emp.Email, "password": emp.Password}}), "userLogin")
		assertUserLoginPayload(t, loginObj, "MFA_REQUIRED")
		ch := requireNonEmptyString(t, loginObj, "mfaChallenge")

		_, _ = doMutationFail(t, cTest, fmt.Sprintf(`mutation ($in: ValidateUserMfaInput!) {
			validateUserMfa(input: $in) { %s }
		}`, gqlValidateUserMfaFields), map[string]any{"in": map[string]any{"challenge": ch, "recoveryCode": oldCodes[0]}}, "validateUserMfa", "AUTHENTICATION")

		mfaData := doGQLOK(t, cTest, fmt.Sprintf(`mutation ($in: ValidateUserMfaInput!) {
			validateUserMfa(input: $in) { %s }
		}`, gqlValidateUserMfaFields), map[string]any{"in": map[string]any{"challenge": ch, "recoveryCode": newCode}})
		assertValidateUserMfaPayload(t, requireMutationSuccess(t, mfaData, "validateUserMfa"))
		t.Cleanup(func() { _, _, _ = logoutBrowserBestEffort(cTest) })
	})
}

// TestMutationCreateUser menguji Mutation createUser untuk pendaftaran publik anonim (user=nil, anti-enumerasi email redirect /login, konflik username), pelanggan login, dan admin RBAC.
func TestMutationCreateUser(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	m := fmt.Sprintf(`mutation ($input: CreateUserInput!) { createUser(input: $input) { %s } }`, gqlCreateUserPayloadFields)

	t.Run("PublicCreateUserAndAntiOracleBehavior", func(t *testing.T) {
		resetGatewayRateLimits(t, cfg)
		anon := newAnonBrowser(t, cfg)
		tag := uniqueTag("pub_cu")
		email := tag + "@integrione.test"

		t.Cleanup(func() {
			userID := psqlScalar(t, cfg, fmt.Sprintf(`SELECT user_id::text FROM user_identities WHERE kind = 'email' AND value = %s LIMIT 1;`, sqlQuote(email)))
			cleanupUserByID(t, cfg, userID)
		})

		// (a) Pendaftaran publik anonim sukses -> user=nil, redirectLocation=nil, serta menerbitkan 1 token email_verification aktif.
		res := requireMutationSuccess(t, doGQLOK(t, anon, m, map[string]any{"input": map[string]any{
			"full_name": "Public User " + tag,
			"email":     email,
			"username":  tag,
			"password":  "StrongPass#2026!",
		}}), "createUser")
		if res["user"] != nil {
			t.Fatalf("public createUser user = %+v, want nil", res["user"])
		}
		if res["redirectLocation"] != nil {
			t.Fatalf("public createUser redirectLocation = %+v, want nil", res["redirectLocation"])
		}
		createdUID := psqlScalar(t, cfg, fmt.Sprintf(`SELECT user_id::text FROM user_identities WHERE kind = 'email' AND value = %s LIMIT 1;`, sqlQuote(email)))
		if createdUID == "" {
			t.Fatalf("expected user row for public email %s", email)
		}
		tokCnt := psqlScalar(t, cfg, fmt.Sprintf(`SELECT count(*) FROM user_tokens WHERE user_id = %s::uuid AND purpose = 'email_verification' AND status = 'active';`, sqlQuote(createdUID)))
		if tokCnt != "1" {
			t.Fatalf("expected 1 active email_verification token, got %s", tokCnt)
		}

		// (b) Duplikat username dengan email baru -> CONFLICT "username has been taken" dan user=nil.
		freshEmail := uniqueTag("pub_fresh") + "@integrione.test"
		_, takenMsg := doMutationFail(t, anon, m, map[string]any{"input": map[string]any{
			"full_name": "Dup Username",
			"email":     freshEmail,
			"username":  tag,
			"password":  "StrongPass#2026!",
		}}, "createUser", "CONFLICT")
		if takenMsg != "username has been taken" {
			t.Fatalf("duplicate username message = %q, want %q", takenMsg, "username has been taken")
		}

		// (c) Duplikat email dengan username baru -> was_successful=true, errors=[], user=nil, redirectLocation="/login", dan HTTP Location="/login".
		freshUsername := uniqueTag("pub_uname")
		t.Cleanup(func() {
			orphanID := psqlScalar(t, cfg, fmt.Sprintf(`SELECT user_id::text FROM user_identities WHERE kind = 'username' AND value = %s LIMIT 1;`, sqlQuote(freshUsername)))
			cleanupUserByID(t, cfg, orphanID)
		})
		dupVars := map[string]any{"input": map[string]any{
			"full_name": "Dup Email Public",
			"email":     email,
			"username":  freshUsername,
			"password":  "StrongPass#2026!",
		}}
		dupResp, dupGql := anon.doGraphQL(m, dupVars, nil)
		logGraphQLExchange(t, "GQL_OK_PUBLIC_DUP_EMAIL", m, dupVars, dupResp, dupGql)
		if dupResp.StatusCode != http.StatusOK || len(dupGql.Errors) > 0 {
			t.Fatalf("public duplicate email status=%d errors=%+v", dupResp.StatusCode, dupGql.Errors)
		}
		if gotLoc := dupResp.Header.Get("Location"); gotLoc != "/login" {
			t.Fatalf("HTTP Location header = %q, want /login", gotLoc)
		}
		var dupRoot map[string]any
		if err := json.Unmarshal(dupGql.Data, &dupRoot); err != nil {
			t.Fatalf("unmarshal duplicate email response: %v", err)
		}
		dupRes := requireMutationSuccess(t, dupRoot, "createUser")
		if dupRes["user"] != nil {
			t.Fatalf("public duplicate email user = %+v, want nil", dupRes["user"])
		}
		if dupRes["redirectLocation"] != "/login" {
			t.Fatalf("public duplicate email redirectLocation = %v, want /login", dupRes["redirectLocation"])
		}
		if orphanCnt := psqlScalar(t, cfg, fmt.Sprintf(`SELECT count(*) FROM user_identities WHERE kind = 'username' AND value = %s;`, sqlQuote(freshUsername))); orphanCnt != "0" {
			t.Fatalf("unexpected username %s persisted on duplicate email rejection (count=%s)", freshUsername, orphanCnt)
		}

		// (d) Duplikat email + duplikat username sekaligus -> wajib CONFLICT "username has been taken" (tanpa oracle email).
		_, bothMsg := doMutationFail(t, anon, m, map[string]any{"input": map[string]any{
			"full_name": "Dup Both",
			"email":     email,
			"username":  tag,
			"password":  "StrongPass#2026!",
		}}, "createUser", "CONFLICT")
		if bothMsg != "username has been taken" {
			t.Fatalf("duplicate email+username message = %q, want %q", bothMsg, "username has been taken")
		}
	})

	t.Run("AuthenticatedCustomerAndAdminCreateUser", func(t *testing.T) {
		resetGatewayRateLimits(t, cfg)
		admin := newSuperAdminBrowser(t, cfg)

		// (e1) Admin login -> user non-nil, namun duplikat email tetap CONFLICT bagi admin.
		cust := createFixtureUser(t, admin, cfg, "CUSTOMER", true)
		u := requireOpMap(t, doGQLOK(t, admin, fmt.Sprintf(`query ($id: ID!) { user(id: $id) { %s } }`, gqlUserFields), map[string]any{"id": cust.ID}), "user")
		assertUserObject(t, u)
		if u["kind"] != "CUSTOMER" {
			t.Fatalf("kind = %v, want CUSTOMER", u["kind"])
		}

		custClient, _ := loginFixtureUser(t, cfg, cust.Email, cust.Password)

		// (e2) Pemanggil pelanggan login -> sukses mengembalikan user non-nil, namun duplikat email tetap dialihkan ke /login tanpa pesan enumerasi.
		custCreatedTag := uniqueTag("auth_cu")
		custCreatedEmail := custCreatedTag + "@integrione.test"
		t.Cleanup(func() {
			id := psqlScalar(t, cfg, fmt.Sprintf(`SELECT user_id::text FROM user_identities WHERE kind = 'email' AND value = %s LIMIT 1;`, sqlQuote(custCreatedEmail)))
			cleanupUserByID(t, cfg, id)
		})
		custCreateData := doGQLOK(t, custClient, m, map[string]any{"input": map[string]any{
			"full_name": "Created By LoggedIn Customer",
			"email":     custCreatedEmail,
			"username":  custCreatedTag,
			"password":  "StrongPass#2026!",
		}})
		createdByCust := requireMutationChildOK(t, custCreateData, "createUser", "user")
		assertUserObject(t, createdByCust)
		if createdByCust["kind"] != "CUSTOMER" || createdByCust["status"] != "PENDING" {
			t.Fatalf("unexpected user from logged-in createUser: %+v", createdByCust)
		}

		dupByCustVars := map[string]any{"input": map[string]any{
			"full_name": "Cust Dup Email",
			"email":     custCreatedEmail,
			"username":  uniqueTag("cust_dup_u"),
			"password":  "StrongPass#2026!",
		}}
		dupByCustResp, dupByCustGql := custClient.doGraphQL(m, dupByCustVars, nil)
		logGraphQLExchange(t, "GQL_OK_CUSTOMER_DUP_EMAIL", m, dupByCustVars, dupByCustResp, dupByCustGql)
		if dupByCustResp.StatusCode != http.StatusOK || len(dupByCustGql.Errors) > 0 {
			t.Fatalf("logged-in non-admin duplicate email status=%d errors=%+v", dupByCustResp.StatusCode, dupByCustGql.Errors)
		}
		if gotLoc := dupByCustResp.Header.Get("Location"); gotLoc != "/login" {
			t.Fatalf("HTTP Location header for non-admin = %q, want /login", gotLoc)
		}
		var dupByCustRoot map[string]any
		if err := json.Unmarshal(dupByCustGql.Data, &dupByCustRoot); err != nil {
			t.Fatalf("unmarshal logged-in duplicate email response: %v", err)
		}
		dupByCustPayload := requireMutationSuccess(t, dupByCustRoot, "createUser")
		if dupByCustPayload["user"] != nil || dupByCustPayload["redirectLocation"] != "/login" {
			t.Fatalf("unexpected non-admin duplicate email payload: %+v", dupByCustPayload)
		}

		// Admin dengan izin iam.users.write tetap menerima CONFLICT saat email sudah terdaftar.
		_, adminDupMsg := doMutationFail(t, admin, m, map[string]any{"input": map[string]any{
			"full_name": "Admin Dup Email",
			"email":     custCreatedEmail,
			"username":  uniqueTag("admin_dup_u"),
			"password":  "StrongPass#2026!",
		}}, "createUser", "CONFLICT")
		if strings.TrimSpace(adminDupMsg) == "" {
			t.Fatalf("expected non-empty conflict message for admin duplicate email")
		}
	})
}

// TestMutationCreateEmployeeUser menguji Mutation createEmployeeUser (Employee) beserta login dan RBAC TestAB.
func TestMutationCreateEmployeeUser(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	m := fmt.Sprintf(`mutation ($input: CreateEmployeeUserInput!) { createEmployeeUser(input: $input) { %s } }`, gqlUserPayloadFields)

	t.Run("ProvisioningRequestRequired", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		tag := uniqueTag("e2eemp")
		doGQLError(t, admin, m, map[string]any{"input": map[string]any{
			"worker_id": "55555555-5555-4555-8555-555555555555",
			"full_name": "Employee Request Fixture",
			"email":     tag + "@integrione.test",
			"username":  tag,
			"password":  "StrongPass#2026!",
		}}, "GRAPHQL_VALIDATION_FAILED")
	})

	t.Run("CreateEmployeeNonNullAndLoginWorks", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		emp := createFixtureUser(t, admin, cfg, "EMPLOYEE", true)
		_, loginRes := loginFixtureUser(t, cfg, emp.Email, emp.Password)
		assertUserLoginPayload(t, loginRes, "AUTHENTICATED")
	})

	t.Run("RBACCustomerForbidden", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		_, custClient := createFixtureCustomerAndLogin(t, admin, cfg)
		tag := uniqueTag("denied_e")
		requireCustomerMutationForbidden(t, custClient, m, map[string]any{"input": map[string]any{
			"full_name": "Denied", "email": tag + "@integrione.test", "username": tag, "password": "StrongPass#2026!",
			"worker_id": "55555555-5555-4555-8555-555555555555", "provisioning_request_id": "66666666-6666-4666-8666-666666666666",
		}}, "createEmployeeUser")
	})
}

// TestMutationCreateVendorUser menguji Mutation createVendorUser (Vendor) beserta login dan RBAC TestAB.
func TestMutationCreateVendorUser(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	m := fmt.Sprintf(`mutation ($input: CreateVendorUserInput!) { createVendorUser(input: $input) { %s } }`, gqlUserPayloadFields)

	t.Run("CreateVendorNonNullAndLoginWorks", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		vnd := createFixtureUser(t, admin, cfg, "VENDOR", true)
		_, loginRes := loginFixtureUser(t, cfg, vnd.Email, vnd.Password)
		assertUserLoginPayload(t, loginRes, "AUTHENTICATED")
	})

	t.Run("RBACCustomerForbidden", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		_, custClient := createFixtureCustomerAndLogin(t, admin, cfg)
		tag := uniqueTag("denied_v")
		requireCustomerMutationForbidden(t, custClient, m, map[string]any{"input": map[string]any{
			"party_id": "11111111-1111-4111-8111-111111111111", "full_name": "Denied", "email": tag + "@integrione.test", "username": tag, "password": "StrongPass#2026!",
		}}, "createVendorUser")
	})
}

// TestMutationCreateSupplierUser menguji Mutation createSupplierUser (Supplier) beserta login dan RBAC TestAB.
func TestMutationCreateSupplierUser(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	m := fmt.Sprintf(`mutation ($input: CreateSupplierUserInput!) { createSupplierUser(input: $input) { %s } }`, gqlUserPayloadFields)

	t.Run("CreateSupplierNonNullAndLoginWorks", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		sup := createFixtureUser(t, admin, cfg, "SUPPLIER", true)
		_, loginRes := loginFixtureUser(t, cfg, sup.Email, sup.Password)
		assertUserLoginPayload(t, loginRes, "AUTHENTICATED")
	})

	t.Run("RBACCustomerForbidden", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		_, custClient := createFixtureCustomerAndLogin(t, admin, cfg)
		tag := uniqueTag("denied_s")
		requireCustomerMutationForbidden(t, custClient, m, map[string]any{"input": map[string]any{
			"party_id": "11111111-1111-4111-8111-111111111111", "full_name": "Denied", "email": tag + "@integrione.test", "username": tag, "password": "StrongPass#2026!",
		}}, "createSupplierUser")
	})
}

// TestMutationActivateUser menguji Mutation activateUser untuk memulihkan akun SUSPENDED/DISABLED menjadi ACTIVE.
func TestMutationActivateUser(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	m := fmt.Sprintf(`mutation ($user_id: ID!) { activateUser(user_id: $user_id) { %s } }`, gqlUserPayloadFields)

	t.Run("ActivateSuspendedUser", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		emp := createFixtureUser(t, admin, cfg, "EMPLOYEE", false)
		suspData := doGQLOK(t, admin, fmt.Sprintf(`mutation ($id: ID!) { suspendUser(user_id: $id) { %s } }`, gqlUserPayloadFields), map[string]any{"id": emp.ID})
		assertUserObject(t, requireMutationChildOK(t, suspData, "suspendUser", "user"))

		u := requireMutationChildOK(t, doGQLOK(t, admin, m, map[string]any{"user_id": emp.ID}), "activateUser", "user")
		assertUserObject(t, u)
		if u["status"] != "ACTIVE" {
			t.Fatalf("status = %v, want ACTIVE", u["status"])
		}
		_, loginRes := loginFixtureUser(t, cfg, emp.Email, emp.Password)
		assertUserLoginPayload(t, loginRes, "AUTHENTICATED")
	})

	t.Run("RBACCustomerForbidden", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		cust, custClient := createFixtureCustomerAndLogin(t, admin, cfg)
		requireCustomerMutationForbidden(t, custClient, m, map[string]any{"user_id": cust.ID}, "activateUser")
	})
}

// TestMutationSuspendUser menguji Mutation suspendUser dan pemblokiran sesi/login pengguna tersuspensi.
func TestMutationSuspendUser(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	m := fmt.Sprintf(`mutation ($user_id: ID!) { suspendUser(user_id: $user_id) { %s } }`, gqlUserPayloadFields)

	t.Run("SuspendActiveUserBlocksSession", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		emp := createFixtureUser(t, admin, cfg, "EMPLOYEE", false)
		empClient, _ := loginFixtureUser(t, cfg, emp.Email, emp.Password)

		u := requireMutationChildOK(t, doGQLOK(t, admin, m, map[string]any{"user_id": emp.ID}), "suspendUser", "user")
		assertUserObject(t, u)
		if u["status"] != "SUSPENDED" {
			t.Fatalf("status = %v, want SUSPENDED", u["status"])
		}
		_ = doGQLError(t, empClient, fmt.Sprintf(`query { me { %s } }`, gqlUserFields), nil, "UNAUTHENTICATED")
	})

	t.Run("RBACCustomerForbidden", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		cust, custClient := createFixtureCustomerAndLogin(t, admin, cfg)
		requireCustomerMutationForbidden(t, custClient, m, map[string]any{"user_id": cust.ID}, "suspendUser")
	})
}

// TestMutationDisableUser menguji Mutation disableUser dan pemblokiran sesi/login pengguna nonaktif.
func TestMutationDisableUser(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	m := fmt.Sprintf(`mutation ($user_id: ID!) { disableUser(user_id: $user_id) { %s } }`, gqlUserPayloadFields)

	t.Run("DisableActiveUserBlocksSession", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		emp := createFixtureUser(t, admin, cfg, "EMPLOYEE", false)
		empClient, _ := loginFixtureUser(t, cfg, emp.Email, emp.Password)

		u := requireMutationChildOK(t, doGQLOK(t, admin, m, map[string]any{"user_id": emp.ID}), "disableUser", "user")
		assertUserObject(t, u)
		if u["status"] != "DISABLED" {
			t.Fatalf("status = %v, want DISABLED", u["status"])
		}
		_ = doGQLError(t, empClient, fmt.Sprintf(`query { me { %s } }`, gqlUserFields), nil, "UNAUTHENTICATED")
	})

	t.Run("RBACCustomerForbidden", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		cust, custClient := createFixtureCustomerAndLogin(t, admin, cfg)
		requireCustomerMutationForbidden(t, custClient, m, map[string]any{"user_id": cust.ID}, "disableUser")
	})
}

// TestMutationSoftDeleteUser menguji Mutation softDeleteUser beserta visibilitas deleted_at pada query users(include_deleted: true).
func TestMutationSoftDeleteUser(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	m := fmt.Sprintf(`mutation ($user_id: ID!) { softDeleteUser(user_id: $user_id) { %s } }`, gqlSoftDeleteUserFields)

	t.Run("SoftDeleteSetsDeletedAt", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		emp := createFixtureUser(t, admin, cfg, "EMPLOYEE", false)

		requireMutationBoolOK(t, doGQLOK(t, admin, m, map[string]any{"user_id": emp.ID}), "softDeleteUser", "deleted")

		usersData := doGQLOK(t, admin, fmt.Sprintf(`query ($q: String) {
			users(query: $q, include_deleted: true, first: 10) { nodes { %s } pageInfo { %s } total_count }
		}`, gqlUserFields, gqlPageInfoFields), map[string]any{"q": emp.Email})
		usersConn := requireOpMap(t, usersData, "users")
		_ = assertPageInfoNonNull(t, usersConn)
		items := requireConnectionNodes(t, usersConn)
		if len(items) != 1 || items[0].(map[string]any)["deleted_at"] == nil {
			t.Fatalf("expected non-null deleted_at after softDeleteUser: %+v", items)
		}
		assertUserObject(t, items[0].(map[string]any))
	})

	t.Run("RBACCustomerForbidden", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		cust, custClient := createFixtureCustomerAndLogin(t, admin, cfg)
		requireCustomerMutationForbidden(t, custClient, m, map[string]any{"user_id": cust.ID}, "softDeleteUser")
	})
}

// TestMutationChangeUsername menguji Mutation changeUsername: pembaruan in-place langsung aktif, pelepasan username lama untuk user lain, konflik duplikasi/sama, dan penambahan username awal customer.
func TestMutationChangeUsername(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	m := fmt.Sprintf(`mutation ($input: UsernameInput!) { changeUsername(input: $input) { %s } }`, gqlUserNamePayloadFields)
	mLogin := fmt.Sprintf(`mutation ($input: UserLoginInput!) { userLogin(input: $input) { %s } }`, gqlUserLoginFields)

	t.Run("RenameReleasesOldUsernameAndAllowsReuseAndRejectsConflicts", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		user1 := createFixtureUser(t, admin, cfg, "EMPLOYEE", true)
		user2 := createFixtureUser(t, admin, cfg, "EMPLOYEE", true)
		oldUsername := user1.Username

		u1Client, _ := loginFixtureUser(t, cfg, user1.Email, user1.Password)
		newUsername := uniqueTag("kibo")
		chRes := requireMutationSuccess(t, doGQLOK(t, u1Client, m, map[string]any{
			"input": map[string]any{"username": newUsername},
		}), "changeUsername")
		if chRes["username"] != newUsername {
			t.Fatalf("changeUsername = %v, want %q", chRes["username"], newUsername)
		}

		// Mengajukan username yang sama persis harus menghasilkan CONFLICT.
		_, _ = doMutationFail(t, u1Client, m, map[string]any{
			"input": map[string]any{"username": newUsername},
		}, "changeUsername", "CONFLICT")

		// Username baru langsung bisa login; username lama tidak lagi menunjuk ke user1.
		_, newLogin := loginFixtureUser(t, cfg, newUsername, user1.Password)
		assertUserLoginPayload(t, newLogin, "AUTHENTICATED")
		_, _ = doMutationFail(t, newAnonBrowser(t, cfg), mLogin, map[string]any{
			"input": map[string]any{"identifier": oldUsername, "password": user1.Password},
		}, "userLogin", "AUTHENTICATION")

		// User2 tidak boleh memakai username baru user1 (CONFLICT), tetapi BISA memakai username lama yang sudah dilepas.
		u2Client, _ := loginFixtureUser(t, cfg, user2.Email, user2.Password)
		_, _ = doMutationFail(t, u2Client, m, map[string]any{
			"input": map[string]any{"username": newUsername},
		}, "changeUsername", "CONFLICT")

		reclaimRes := requireMutationSuccess(t, doGQLOK(t, u2Client, m, map[string]any{
			"input": map[string]any{"username": oldUsername},
		}), "changeUsername")
		if reclaimRes["username"] != oldUsername {
			t.Fatalf("reclaimed username = %v, want %q", reclaimRes["username"], oldUsername)
		}
		_, u2Login := loginFixtureUser(t, cfg, oldUsername, user2.Password)
		assertUserLoginPayload(t, u2Login, "AUTHENTICATED")

		// Customer tanpa username awal dapat menambahkan username melalui changeUsername.
		custNoUsername := createFixtureUser(t, admin, cfg, "CUSTOMER", false)
		cClient, _ := loginFixtureUser(t, cfg, custNoUsername.Email, custNoUsername.Password)
		addedUsername := uniqueTag("cust_uname")
		addRes := requireMutationSuccess(t, doGQLOK(t, cClient, m, map[string]any{
			"input": map[string]any{"username": addedUsername},
		}), "changeUsername")
		if addRes["username"] != addedUsername {
			t.Fatalf("added username = %v, want %q", addRes["username"], addedUsername)
		}
		_, custLogin := loginFixtureUser(t, cfg, addedUsername, custNoUsername.Password)
		assertUserLoginPayload(t, custLogin, "AUTHENTICATED")
	})

	t.Run("AnonymousRejected", func(t *testing.T) {
		resetGatewayRateLimits(t, cfg)
		requireAnonMutationUnauthenticated(t, newAnonBrowser(t, cfg), m, map[string]any{
			"input": map[string]any{"username": uniqueTag("anon_u")},
		}, "changeUsername")
	})
}

// TestMutationUpdateUserEmailDeletePendingAndResend menguji alur penuh updateUserEmail (pending email tanpa mengganti email aktif), deletePendingUserEmail, resendValidationEmailToken (Kasus 1 & 2, 1 token aktif), serta verifikasi akhir.
func TestMutationUpdateUserEmailDeletePendingAndResend(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	mReg := fmt.Sprintf(`mutation ($input: CreateUserInput!) { createUser(input: $input) { %s } }`, gqlCreateUserPayloadFields)
	mUpd := fmt.Sprintf(`mutation ($input: UserEmailInput!) { updateUserEmail(input: $input) { %s } }`, gqlUserEmailPayloadFields)
	mResend := fmt.Sprintf(`mutation ($input: ResendValidationEmailInput) { resendValidationEmailToken(input: $input) { %s } }`, gqlResendValidationFields)
	mVerify := fmt.Sprintf(`mutation ($input: VerifyEmailInput!) { verifyEmail(input: $input) { %s } }`, gqlVerifyEmailFields)
	mDel := fmt.Sprintf(`mutation { deletePendingUserEmail { %s } }`, gqlDeletePendingEmailFields)

	t.Run("Case2PendingRegistrationResendSingleActiveToken", func(t *testing.T) {
		resetGatewayRateLimits(t, cfg)
		anon := newAnonBrowser(t, cfg)
		tag := uniqueTag("res_p")
		regEmail := tag + "@integrione.test"

		regRes := requireMutationSuccess(t, doGQLOK(t, anon, mReg, map[string]any{"input": map[string]any{
			"full_name": "Resend Pending", "email": regEmail, "username": tag, "password": "Zx9#mQp2!vLk8@Rt",
		}}), "createUser")
		if regRes["user"] != nil {
			t.Fatalf("public createUser returned non-nil user: %+v", regRes["user"])
		}
		userID := psqlScalar(t, cfg, fmt.Sprintf(`SELECT user_id::text FROM user_identities WHERE kind = 'email' AND value = %s LIMIT 1;`, sqlQuote(regEmail)))
		t.Cleanup(func() { cleanupUserByID(t, cfg, userID) })

		firstTokID := psqlScalar(t, cfg, fmt.Sprintf(`SELECT id::text FROM user_tokens WHERE user_id = %s::uuid AND purpose = 'email_verification' AND status = 'active' LIMIT 1;`, sqlQuote(userID)))
		if firstTokID == "" {
			t.Fatalf("expected active initial email verification token")
		}

		// Tanpa input email pada pemanggil anonim harus ARGUMENT.
		_, _ = doMutationFail(t, anon, mResend, map[string]any{"input": nil}, "resendValidationEmailToken", "ARGUMENT")

		// Kasus 2 anonim: resend menerbitkan token aktif baru dan menandai token lama expired.
		resPayload := requireMutationBoolOK(t, doGQLOK(t, anon, mResend, map[string]any{
			"input": map[string]any{"email": regEmail},
		}), "resendValidationEmailToken", "is_resend")
		_ = resPayload
		activeCnt := psqlScalar(t, cfg, fmt.Sprintf(`SELECT count(*) FROM user_tokens WHERE user_id = %s::uuid AND purpose = 'email_verification' AND status = 'active';`, sqlQuote(userID)))
		expiredCnt := psqlScalar(t, cfg, fmt.Sprintf(`SELECT count(*) FROM user_tokens WHERE user_id = %s::uuid AND purpose = 'email_verification' AND status = 'expired';`, sqlQuote(userID)))
		if activeCnt != "1" || expiredCnt != "1" {
			t.Fatalf("expected 1 active and 1 expired token after resend, got active=%s expired=%s", activeCnt, expiredCnt)
		}

		// Email ghost tetap is_resend=true (anti-enumerasi).
		_ = requireMutationBoolOK(t, doGQLOK(t, anon, mResend, map[string]any{
			"input": map[string]any{"email": "ghost-" + tag + "@integrione.test"},
		}), "resendValidationEmailToken", "is_resend")
	})

	t.Run("Case1UpdateUserEmailPreservesOldResendsDeletesAndVerifies", func(t *testing.T) {
		resetGatewayRateLimits(t, cfg)
		admin := newSuperAdminBrowser(t, cfg)
		emp := createFixtureUser(t, admin, cfg, "EMPLOYEE", false)
		empClient, _ := loginFixtureUser(t, cfg, emp.Email, emp.Password)

		pending1 := uniqueTag("pend1") + "@integrione.test"
		upd1 := requireMutationSuccess(t, doGQLOK(t, empClient, mUpd, map[string]any{
			"input": map[string]any{"email": pending1, "current_password": emp.Password},
		}), "updateUserEmail")
		if upd1["pending_email"] != pending1 || !requireBoolField(t, upd1, "is_pending") {
			t.Fatalf("unexpected updateUserEmail payload: %+v", upd1)
		}
		_ = requireNonEmptyString(t, upd1, "expires_at")
		expiredBefore := psqlScalar(t, cfg, fmt.Sprintf(`SELECT count(*) FROM user_tokens WHERE user_id = %s::uuid AND purpose = 'email_verification' AND status = 'expired';`, sqlQuote(emp.RawID)))

		// Sebelum diverifikasi, login dengan email lama tetap berhasil dan email baru belum terverifikasi.
		_, oldLogin := loginFixtureUser(t, cfg, emp.Email, emp.Password)
		assertUserLoginPayload(t, oldLogin, "AUTHENTICATED")

		// Kasus 1 autentikasi: resendValidationEmailToken tanpa mengirim email ulang sukses & menjaga 1 active token.
		_ = requireMutationBoolOK(t, doGQLOK(t, empClient, mResend, map[string]any{"input": nil}), "resendValidationEmailToken", "is_resend")
		activeAfterResend := psqlScalar(t, cfg, fmt.Sprintf(`SELECT count(*) FROM user_tokens WHERE user_id = %s::uuid AND purpose = 'email_verification' AND status = 'active';`, sqlQuote(emp.RawID)))
		expiredAfterResend := psqlScalar(t, cfg, fmt.Sprintf(`SELECT count(*) FROM user_tokens WHERE user_id = %s::uuid AND purpose = 'email_verification' AND status = 'expired';`, sqlQuote(emp.RawID)))
		if activeAfterResend != "1" || expiredBefore != "1" || expiredAfterResend != "2" {
			t.Fatalf("expected 1 active & 1->2 expired tokens across updateUserEmail->resendValidationEmailToken, got active=%s expiredBefore=%s expiredAfter=%s", activeAfterResend, expiredBefore, expiredAfterResend)
		}

		// Mengganti pending email via updateUserEmail menjaga tetap 1 pending email & 1 active token.
		pending2 := uniqueTag("pend2") + "@integrione.test"
		upd2 := requireMutationSuccess(t, doGQLOK(t, empClient, mUpd, map[string]any{
			"input": map[string]any{"email": pending2, "current_password": emp.Password},
		}), "updateUserEmail")
		if upd2["pending_email"] != pending2 || !requireBoolField(t, upd2, "is_pending") {
			t.Fatalf("unexpected replaced pending email payload: %+v", upd2)
		}

		// Batalkan pending email via deletePendingUserEmail -> identitas pending & token aktif tercabut.
		requireMutationBoolOK(t, doGQLOK(t, empClient, mDel, nil), "deletePendingUserEmail", "deleted")
		pendIdentCnt := psqlScalar(t, cfg, fmt.Sprintf(`SELECT count(*) FROM user_identities WHERE user_id = %s::uuid AND kind = 'email' AND is_verified = false;`, sqlQuote(emp.RawID)))
		activeAfterDel := psqlScalar(t, cfg, fmt.Sprintf(`SELECT count(*) FROM user_tokens WHERE user_id = %s::uuid AND purpose = 'email_verification' AND status = 'active';`, sqlQuote(emp.RawID)))
		if pendIdentCnt != "0" || activeAfterDel != "0" {
			t.Fatalf("expected 0 unverified emails and 0 active tokens after deletePendingUserEmail, got pend=%s active=%s", pendIdentCnt, activeAfterDel)
		}

		// Alur lengkap: updateUserEmail -> verifyEmail -> email baru menjadi aktif & email lama digantikan.
		pending3 := uniqueTag("pend3") + "@integrione.test"
		_ = requireMutationSuccess(t, doGQLOK(t, empClient, mUpd, map[string]any{
			"input": map[string]any{"email": pending3, "current_password": emp.Password},
		}), "updateUserEmail")
		rawTok := "e2e-email-upd-" + uniqueTag("tok")
		psqlExec(t, cfg, fmt.Sprintf(
			`UPDATE user_tokens SET token_digest = decode(%s, 'hex') WHERE user_id = %s::uuid AND purpose = 'email_verification' AND status = 'active';`,
			sqlQuote(sha256Hex(rawTok)), sqlQuote(emp.RawID),
		))
		anon := newAnonBrowser(t, cfg)
		verRes := requireMutationSuccess(t, doGQLOK(t, anon, mVerify, map[string]any{"input": map[string]any{"token": rawTok}}), "verifyEmail")
		assertVerifyEmailPayload(t, verRes, "VERIFIED", true)

		_, newEmailLogin := loginFixtureUser(t, cfg, pending3, emp.Password)
		assertUserLoginPayload(t, newEmailLogin, "AUTHENTICATED")
		_, _ = doMutationFail(t, newAnonBrowser(t, cfg), fmt.Sprintf(`mutation ($in: UserLoginInput!) { userLogin(input: $in) { %s } }`, gqlUserLoginFields), map[string]any{
			"in": map[string]any{"identifier": emp.Email, "password": emp.Password},
		}, "userLogin", "AUTHENTICATION")
	})

	t.Run("AnonymousUpdateOrDeleteRejected", func(t *testing.T) {
		resetGatewayRateLimits(t, cfg)
		anon := newAnonBrowser(t, cfg)
		requireAnonMutationUnauthenticated(t, anon, mUpd, map[string]any{
			"input": map[string]any{"email": "anon-upd@integrione.test", "current_password": "StrongPass#2026!"},
		}, "updateUserEmail")
		requireAnonMutationUnauthenticated(t, anon, mDel, nil, "deletePendingUserEmail")
	})
}

// TestMutationSetPrimaryUserIdentity menguji Mutation setPrimaryUserIdentity antar identitas terverifikasi dan penolakan identitas belum terverifikasi.
func TestMutationSetPrimaryUserIdentity(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	m := fmt.Sprintf(`mutation ($user_id: ID!, $identity_id: ID!) { setPrimaryUserIdentity(user_id: $user_id, identity_id: $identity_id) { %s } }`, gqlUserIdentityPayload)

	t.Run("SwitchPrimaryToVerifiedUsernameAndRejectUnverified", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		emp := createFixtureUser(t, admin, cfg, "EMPLOYEE", true)

		identsConn := requireOpMap(t, doGQLOK(t, admin, fmt.Sprintf(`query ($u: ID!) { userIdentities(user_id: $u, first: 10) { nodes { %s } pageInfo { %s } total_count } }`, gqlIdentityFields, gqlPageInfoFields), map[string]any{"u": emp.ID}), "userIdentities")
		_ = assertPageInfoNonNull(t, identsConn)
		var usernameIdentID string
		for _, raw := range requireConnectionNodes(t, identsConn) {
			it := raw.(map[string]any)
			assertUserIdentityObject(t, it)
			if it["kind"] == "USERNAME" {
				usernameIdentID = requireNonEmptyString(t, it, "id")
			}
		}
		if usernameIdentID == "" {
			t.Fatalf("expected USERNAME identity on employee")
		}

		promoted := requireMutationChildOK(t, doGQLOK(t, admin, m, map[string]any{
			"user_id":     emp.ID,
			"identity_id": usernameIdentID,
		}), "setPrimaryUserIdentity", "user_identity")
		assertUserIdentityObject(t, promoted)
		_ = requireNonEmptyString(t, promoted, "verified_at")
		if !requireBoolField(t, promoted, "is_primary") || promoted["id"] != usernameIdentID {
			t.Fatalf("expected is_primary=true on promoted identity: %+v", promoted)
		}

		// Identitas pending email yang belum terverifikasi tidak boleh dijadikan primary.
		empClient, _ := loginFixtureUser(t, cfg, emp.Email, emp.Password)
		unverEmail := uniqueTag("unver") + "@integrione.test"
		_ = requireMutationSuccess(t, doGQLOK(t, empClient, fmt.Sprintf(`mutation ($in: UserEmailInput!) { updateUserEmail(input: $in) { %s } }`, gqlUserEmailPayloadFields), map[string]any{
			"in": map[string]any{"email": unverEmail, "current_password": emp.Password},
		}), "updateUserEmail")

		afterConn := requireOpMap(t, doGQLOK(t, admin, fmt.Sprintf(`query ($u: ID!) { userIdentities(user_id: $u, first: 10) { nodes { %s } pageInfo { %s } total_count } }`, gqlIdentityFields, gqlPageInfoFields), map[string]any{"u": emp.ID}), "userIdentities")
		var unverID string
		for _, raw := range requireConnectionNodes(t, afterConn) {
			it := raw.(map[string]any)
			assertUserIdentityObject(t, it)
			if it["value"] == unverEmail {
				unverID = requireNonEmptyString(t, it, "id")
			}
		}
		if unverID == "" {
			t.Fatalf("expected unverified pending email identity to exist")
		}
		_, _ = doMutationFail(t, admin, m, map[string]any{"user_id": emp.ID, "identity_id": unverID}, "setPrimaryUserIdentity", "AUTHORIZATION")
	})

	t.Run("RBACCustomerForbidden", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		cust, custClient := createFixtureCustomerAndLogin(t, admin, cfg)
		requireCustomerMutationForbidden(t, custClient, m, map[string]any{"user_id": cust.ID, "identity_id": "00000000-0000-4000-8000-000000000001"}, "setPrimaryUserIdentity")
	})
}

// TestMutationRemoveUserIdentity menguji Mutation removeUserIdentity pada identitas sekunder dan proteksi identitas primary.
func TestMutationRemoveUserIdentity(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	m := fmt.Sprintf(`mutation ($user_id: ID!, $identity_id: ID!) { removeUserIdentity(user_id: $user_id, identity_id: $identity_id) { %s } }`, gqlRemoveIdentityFields)

	t.Run("RemoveSecondaryAndProtectPrimary", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		emp := createFixtureUser(t, admin, cfg, "EMPLOYEE", true)

		identsConn := requireOpMap(t, doGQLOK(t, admin, fmt.Sprintf(`query ($u: ID!) { userIdentities(user_id: $u, first: 10) { nodes { %s } pageInfo { %s } total_count } }`, gqlIdentityFields, gqlPageInfoFields), map[string]any{"u": emp.ID}), "userIdentities")
		_ = assertPageInfoNonNull(t, identsConn)
		var primaryID, secondaryID string
		for _, raw := range requireConnectionNodes(t, identsConn) {
			it := raw.(map[string]any)
			assertUserIdentityObject(t, it)
			if requireBoolField(t, it, "is_primary") {
				primaryID = requireNonEmptyString(t, it, "id")
			} else {
				secondaryID = requireNonEmptyString(t, it, "id")
			}
		}

		requireMutationBoolOK(t, doGQLOK(t, admin, m, map[string]any{"user_id": emp.ID, "identity_id": secondaryID}), "removeUserIdentity", "removed")

		_, _ = doMutationFail(t, admin, m, map[string]any{"user_id": emp.ID, "identity_id": primaryID}, "removeUserIdentity", "UNPROCESSABLE_ENTITY")
	})

	t.Run("RBACCustomerForbidden", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		cust, custClient := createFixtureCustomerAndLogin(t, admin, cfg)
		requireCustomerMutationForbidden(t, custClient, m, map[string]any{"user_id": cust.ID, "identity_id": "00000000-0000-4000-8000-000000000001"}, "removeUserIdentity")
	})
}

// TestMutationCreateRole menguji Mutation createRole untuk role valid, validasi redirect path, duplikasi kode, dan RBAC.
func TestMutationCreateRole(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	m := fmt.Sprintf(`mutation ($input: CreateRoleInput!) { createRole(input: $input) { %s } }`, gqlRolePayloadFields)

	t.Run("CreateValidAndRejectInvalidRedirectOrDuplicate", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		roleID, roleCode := createFixtureRole(t, admin, cfg, "cr")
		_ = roleID

		_, _ = doMutationFail(t, admin, m, map[string]any{"input": map[string]any{
			"code": roleCode, "name": "Dup", "management_domain": "iam", "default_redirect_path": "/identity",
		}}, "createRole", "CONFLICT")

		_, _ = doMutationFail(t, admin, m, map[string]any{"input": map[string]any{
			"code": uniqueTag("bad_red"), "name": "BadRedirect", "management_domain": "iam", "default_redirect_path": "https://evil.example.com",
		}}, "createRole", "ARGUMENT")
	})

	t.Run("RBACCustomerForbidden", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		_, custClient := createFixtureCustomerAndLogin(t, admin, cfg)
		requireCustomerMutationForbidden(t, custClient, m, map[string]any{"input": map[string]any{
			"code": uniqueTag("no_role"), "name": "No", "management_domain": "iam", "default_redirect_path": "/identity",
		}}, "createRole")
	})
}

// TestMutationUpdateRoleDetails menguji Mutation updateRoleDetails dan verifikasi persistensi perubahan field Role.
func TestMutationUpdateRoleDetails(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	m := fmt.Sprintf(`mutation ($input: UpdateRoleDetailsInput!) { updateRoleDetails(input: $input) { %s } }`, gqlRolePayloadFields)

	t.Run("UpdateAllMutableFields", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		roleID, _ := createFixtureRole(t, admin, cfg, "upd")
		newName := "Updated Role " + uniqueTag("r")

		role := requireMutationChildOK(t, doGQLOK(t, admin, m, map[string]any{
			"input": map[string]any{
				"id":                    roleID,
				"name":                  newName,
				"description":           "Updated E2E role description",
				"management_domain":     "finance",
				"default_redirect_path": "/finance",
				"redirect_priority":     85,
			},
		}), "updateRoleDetails", "role")
		assertRoleObject(t, role)
		if role["name"] != newName || role["management_domain"] != "finance" || role["default_redirect_path"] != "/finance" || requireNumberField(t, role, "redirect_priority") != 85 {
			t.Fatalf("unexpected updated role: %+v", role)
		}
	})

	t.Run("RBACCustomerForbidden", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		roleID, _ := createFixtureRole(t, admin, cfg, "upd_rbac")
		_, custClient := createFixtureCustomerAndLogin(t, admin, cfg)
		requireCustomerMutationForbidden(t, custClient, m, map[string]any{
			"input": map[string]any{"id": roleID, "name": "No", "management_domain": "iam", "default_redirect_path": "/identity"},
		}, "updateRoleDetails")
	})
}

// TestMutationActivateRole menguji Mutation activateRole pada role nonaktif.
func TestMutationActivateRole(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	m := fmt.Sprintf(`mutation ($id: ID!) { activateRole(id: $id) { %s } }`, gqlRolePayloadFields)

	t.Run("ActivateDeactivatedRole", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		roleID, _ := createFixtureRole(t, admin, cfg, "act_r")
		deactData := doGQLOK(t, admin, fmt.Sprintf(`mutation ($id: ID!) { deactivateRole(id: $id) { %s } }`, gqlRolePayloadFields), map[string]any{"id": roleID})
		assertRoleObject(t, requireMutationChildOK(t, deactData, "deactivateRole", "role"))

		role := requireMutationChildOK(t, doGQLOK(t, admin, m, map[string]any{"id": roleID}), "activateRole", "role")
		assertRoleObject(t, role)
		if !requireBoolField(t, role, "is_active") {
			t.Fatalf("expected is_active=true after activateRole")
		}
	})

	t.Run("RBACCustomerForbidden", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		roleID, _ := createFixtureRole(t, admin, cfg, "act_rbac")
		_, custClient := createFixtureCustomerAndLogin(t, admin, cfg)
		requireCustomerMutationForbidden(t, custClient, m, map[string]any{"id": roleID}, "activateRole")
	})
}

// TestMutationDeactivateRole menguji Mutation deactivateRole untuk role kustom serta proteksi role sistem.
func TestMutationDeactivateRole(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	m := fmt.Sprintf(`mutation ($id: ID!) { deactivateRole(id: $id) { %s } }`, gqlRolePayloadFields)

	t.Run("DeactivateCustomAndProtectSystemRole", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		roleID, _ := createFixtureRole(t, admin, cfg, "deact_r")

		role := requireMutationChildOK(t, doGQLOK(t, admin, m, map[string]any{"id": roleID}), "deactivateRole", "role")
		assertRoleObject(t, role)
		if requireBoolField(t, role, "is_active") {
			t.Fatalf("expected is_active=false after deactivateRole")
		}

		sysRole := requireOpMap(t, doGQLOK(t, admin, fmt.Sprintf(`query { roleByCode(code: "superadmin") { %s } }`, gqlRoleFields), nil), "roleByCode")
		assertRoleObject(t, sysRole)
		_, _ = doMutationFail(t, admin, m, map[string]any{"id": sysRole["id"]}, "deactivateRole", "UNPROCESSABLE_ENTITY")
	})

	t.Run("RBACCustomerForbidden", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		roleID, _ := createFixtureRole(t, admin, cfg, "deact_rbac")
		_, custClient := createFixtureCustomerAndLogin(t, admin, cfg)
		requireCustomerMutationForbidden(t, custClient, m, map[string]any{"id": roleID}, "deactivateRole")
	})
}

// TestMutationDeleteRole menguji Mutation deleteRole pada role kustom serta proteksi role sistem.
func TestMutationDeleteRole(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	m := fmt.Sprintf(`mutation ($id: ID!) { deleteRole(id: $id) { %s } }`, gqlDeleteRoleFields)

	t.Run("DeleteCustomAndProtectSystemRole", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		roleID, _ := createFixtureRole(t, admin, cfg, "del_r")

		requireMutationBoolOK(t, doGQLOK(t, admin, m, map[string]any{"id": roleID}), "deleteRole", "deleted")
		_ = doGQLError(t, admin, fmt.Sprintf(`query ($id: ID!) { role(id: $id) { %s } }`, gqlRoleFields), map[string]any{"id": roleID}, "NOT_FOUND")

		sysRole := requireOpMap(t, doGQLOK(t, admin, fmt.Sprintf(`query { roleByCode(code: "superadmin") { %s } }`, gqlRoleFields), nil), "roleByCode")
		assertRoleObject(t, sysRole)
		_, _ = doMutationFail(t, admin, m, map[string]any{"id": sysRole["id"]}, "deleteRole", "UNPROCESSABLE_ENTITY")
	})

	t.Run("RBACCustomerForbidden", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		roleID, _ := createFixtureRole(t, admin, cfg, "del_rbac")
		_, custClient := createFixtureCustomerAndLogin(t, admin, cfg)
		requireCustomerMutationForbidden(t, custClient, m, map[string]any{"id": roleID}, "deleteRole")
	})
}

// TestMutationAssignRolePermission menguji Mutation assignRolePermission serta verifikasi izin pada rolePermissions.
func TestMutationAssignRolePermission(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	m := fmt.Sprintf(`mutation ($role_id: ID!, $permission_id: ID!) {
		assignRolePermission(role_id: $role_id, permission_id: $permission_id) { %s }
	}`, gqlRolePermPayloadFields)

	t.Run("AssignPermissionReturnsNonNull", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		roleID, _ := createFixtureRole(t, admin, cfg, "arp")
		permID := findPermissionIDByCode(t, admin, "iam.users.read")

		rp := requireMutationChildOK(t, doGQLOK(t, admin, m, map[string]any{"role_id": roleID, "permission_id": permID}), "assignRolePermission", "role_permission_edge")
		assertRolePermissionObject(t, rp)
		node := requireChildMap(t, rp, "node")
		if rp["role_id"] != roleID || node["id"] != permID {
			t.Fatalf("unexpected RolePermissionEdge: %+v", rp)
		}
	})

	t.Run("FlexibleGIDAndRawUUID", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		roleGID, _ := createFixtureRole(t, admin, cfg, "arp_flex")
		perm1GID := findPermissionIDByCode(t, admin, "iam.users.read")
		perm2GID := findPermissionIDByCode(t, admin, "iam.roles.read")

		roleRawLower := rawUUID(roleGID)
		roleRawUpper := strings.ToUpper(roleRawLower)
		perm1RawLower := rawUUID(perm1GID)
		perm1RawUpper := strings.ToUpper(perm1RawLower)
		perm2RawLower := rawUUID(perm2GID)
		perm2RawUpper := strings.ToUpper(perm2RawLower)

		mRevoke := fmt.Sprintf(`mutation ($role_id: ID!, $permission_id: ID!) { revokeRolePermission(role_id: $role_id, permission_id: $permission_id) { %s } }`, gqlRevokeRolePermFields)
		qRolePerms := fmt.Sprintf(`query ($r: ID!) { rolePermissions(role_id: $r, first: 10) { edges { %s } nodes { %s } pageInfo { %s } total_count } }`, gqlRolePermissionEdgeFields, gqlPermissionFields, gqlPageInfoFields)

		// 1. role_id berbentuk GID + permission_id berbentuk UUID mentah lowercase, lalu dicabut dengan raw UUID + GID.
		rp1 := requireMutationChildOK(t, doGQLOK(t, admin, m, map[string]any{"role_id": roleGID, "permission_id": perm1RawLower}), "assignRolePermission", "role_permission_edge")
		assertRolePermissionObject(t, rp1)
		n1 := requireChildMap(t, rp1, "node")
		if rp1["role_id"] != roleGID || n1["id"] != perm1GID || n1["_id"] != perm1RawLower {
			t.Fatalf("GID role + raw perm unexpected edge: %+v", rp1)
		}
		requireMutationBoolOK(t, doGQLOK(t, admin, mRevoke, map[string]any{"role_id": roleRawLower, "permission_id": perm1GID}), "revokeRolePermission", "revoked")
		rpEmpty := requireOpMap(t, doGQLOK(t, admin, qRolePerms, map[string]any{"r": roleRawUpper}), "rolePermissions")
		if total := requireNumberField(t, rpEmpty, "total_count"); total != 0 {
			t.Fatalf("expected total_count=0 after flexible revoke, got %d", total)
		}

		// 2. Kombinasi raw lowercase + GID serta raw uppercase + raw uppercase harus menormalkan output ke GID lowercase kanonik.
		rp2 := requireMutationChildOK(t, doGQLOK(t, admin, m, map[string]any{"role_id": roleRawLower, "permission_id": perm1GID}), "assignRolePermission", "role_permission_edge")
		assertRolePermissionObject(t, rp2)
		n2 := requireChildMap(t, rp2, "node")
		if rp2["role_id"] != roleGID || n2["id"] != perm1GID || n2["_id"] != perm1RawLower {
			t.Fatalf("raw lower role + GID perm unexpected edge: %+v", rp2)
		}

		rp3 := requireMutationChildOK(t, doGQLOK(t, admin, m, map[string]any{"role_id": roleRawUpper, "permission_id": perm2RawUpper}), "assignRolePermission", "role_permission_edge")
		assertRolePermissionObject(t, rp3)
		n3 := requireChildMap(t, rp3, "node")
		if rp3["role_id"] != roleGID || n3["id"] != perm2GID || n3["_id"] != perm2RawLower {
			t.Fatalf("raw upper role + raw upper perm unexpected edge: %+v", rp3)
		}

		// 3. Query rolePermissions menerima raw UUID dan menampilkan kedua izin secara konsisten, lalu bersih setelah revoke fleksibel.
		rpConn := requireOpMap(t, doGQLOK(t, admin, qRolePerms, map[string]any{"r": roleRawLower}), "rolePermissions")
		if total := requireNumberField(t, rpConn, "total_count"); total != 2 {
			t.Fatalf("rolePermissions(rawLower) total_count=%d, want 2", total)
		}
		for _, rawEdge := range requireConnectionEdges(t, rpConn) {
			edge := rawEdge.(map[string]any)
			if edge["role_id"] != roleGID {
				t.Fatalf("edge.role_id=%v, want %q", edge["role_id"], roleGID)
			}
		}

		requireMutationBoolOK(t, doGQLOK(t, admin, mRevoke, map[string]any{"role_id": roleRawUpper, "permission_id": perm1RawUpper}), "revokeRolePermission", "revoked")
		requireMutationBoolOK(t, doGQLOK(t, admin, mRevoke, map[string]any{"role_id": roleRawLower, "permission_id": perm2RawLower}), "revokeRolePermission", "revoked")
		rpAfter := requireOpMap(t, doGQLOK(t, admin, qRolePerms, map[string]any{"r": roleGID}), "rolePermissions")
		if total := requireNumberField(t, rpAfter, "total_count"); total != 0 {
			t.Fatalf("expected total_count=0 after revoking both permissions, got %d", total)
		}

		// 4. Kasus negatif: GID tertukar resource atau format bukan UUID wajib ditolak payload validation/argument.
		negCases := []struct {
			name   string
			opName string
			query  string
			roleID string
			permID string
		}{
			{name: "AssignSwappedPermissionGIDAsRole", opName: "assignRolePermission", query: m, roleID: perm1GID, permID: perm2GID},
			{name: "AssignSwappedRoleGIDAsPermission", opName: "assignRolePermission", query: m, roleID: roleGID, permID: roleGID},
			{name: "AssignWrongUserGIDAsPermission", opName: "assignRolePermission", query: m, roleID: roleGID, permID: "gid://integrione/user/" + perm1RawLower},
			{name: "AssignMalformedUUID", opName: "assignRolePermission", query: m, roleID: roleRawLower, permID: "not-a-valid-uuid"},
			{name: "RevokeSwappedResourceGID", opName: "revokeRolePermission", query: mRevoke, roleID: perm1GID, permID: roleGID},
		}
		for _, nc := range negCases {
			_, _ = doMutationFail(t, admin, nc.query, map[string]any{"role_id": nc.roleID, "permission_id": nc.permID}, nc.opName, "ARGUMENT", "VALIDATION")
		}
	})

	t.Run("RBACCustomerForbidden", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		roleID, _ := createFixtureRole(t, admin, cfg, "arp_rbac")
		permID := findPermissionIDByCode(t, admin, "iam.users.read")
		_, custClient := createFixtureCustomerAndLogin(t, admin, cfg)
		requireCustomerMutationForbidden(t, custClient, m, map[string]any{"role_id": roleID, "permission_id": permID}, "assignRolePermission")
	})
}

// TestMutationRevokeRolePermission menguji Mutation revokeRolePermission serta pengosongan izin pada rolePermissions.
func TestMutationRevokeRolePermission(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	m := fmt.Sprintf(`mutation ($role_id: ID!, $permission_id: ID!) { revokeRolePermission(role_id: $role_id, permission_id: $permission_id) { %s } }`, gqlRevokeRolePermFields)

	t.Run("RevokeAssignedPermissionReturnsTrue", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		roleID, _ := createFixtureRole(t, admin, cfg, "rrp")
		permID := findPermissionIDByCode(t, admin, "iam.users.read")
		rpData := doGQLOK(t, admin, fmt.Sprintf(`mutation ($r: ID!, $p: ID!) { assignRolePermission(role_id: $r, permission_id: $p) { %s } }`, gqlRolePermPayloadFields), map[string]any{
			"r": roleID, "p": permID,
		})
		assertRolePermissionObject(t, requireMutationChildOK(t, rpData, "assignRolePermission", "role_permission_edge"))

		requireMutationBoolOK(t, doGQLOK(t, admin, m, map[string]any{"role_id": roleID, "permission_id": permID}), "revokeRolePermission", "revoked")
		rpConn := requireOpMap(t, doGQLOK(t, admin, fmt.Sprintf(`query ($r: ID!) { rolePermissions(role_id: $r, first: 10) { nodes { %s } pageInfo { %s } total_count } }`, gqlPermissionFields, gqlPageInfoFields), map[string]any{"r": roleID}), "rolePermissions")
		_ = assertPageInfoNonNull(t, rpConn)
		rem := requireConnectionNodes(t, rpConn)
		if len(rem) != 0 {
			t.Fatalf("expected 0 permissions remaining, got %+v", rem)
		}
	})

	t.Run("RBACCustomerForbidden", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		roleID, _ := createFixtureRole(t, admin, cfg, "rrp_rbac")
		permID := findPermissionIDByCode(t, admin, "iam.users.read")
		_, custClient := createFixtureCustomerAndLogin(t, admin, cfg)
		requireCustomerMutationForbidden(t, custClient, m, map[string]any{"role_id": roleID, "permission_id": permID}, "revokeRolePermission")
	})
}

// TestMutationAssignUserRole menguji Mutation assignUserRole baik scope GLOBAL maupun scope FACILITY + expires_at.
func TestMutationAssignUserRole(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	m := fmt.Sprintf(`mutation ($input: AssignUserRoleInput!) {
		assignUserRole(input: $input) { %s }
	}`, gqlUserRolePayloadFields)

	t.Run("AssignGlobalAndScopedRole", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		emp := createFixtureUser(t, admin, cfg, "EMPLOYEE", false)
		role1ID, _ := createFixtureRole(t, admin, cfg, "aur_g")
		role2ID, _ := createFixtureRole(t, admin, cfg, "aur_f")

		urGlobal := requireMutationChildOK(t, doGQLOK(t, admin, m, map[string]any{
			"input": map[string]any{"user_id": emp.ID, "role_id": role1ID, "scope_type": "GLOBAL"},
		}), "assignUserRole", "user_role_edge")
		assertUserRoleObject(t, urGlobal)
		_ = requireNonEmptyString(t, urGlobal, "granted_by")

		scopeID := createFixtureEFFacility(t, admin, cfg)
		urScoped := requireMutationChildOK(t, doGQLOK(t, admin, m, map[string]any{
			"input": map[string]any{
				"user_id":    emp.ID,
				"role_id":    role2ID,
				"scope_type": "FACILITY",
				"scope_id":   scopeID,
				"expires_at": time.Now().Add(4 * time.Hour).UTC().Format(time.RFC3339),
			},
		}), "assignUserRole", "user_role_edge")
		assertUserRoleObject(t, urScoped)
		_ = requireNonEmptyString(t, urScoped, "granted_by")
		_ = requireNonEmptyString(t, urScoped, "expires_at")
		scopeMap := requireChildMap(t, urScoped, "scope")
		if scopeMap["type"] != "FACILITY" || scopeMap["id"] != scopeID {
			t.Fatalf("unexpected scope on UserRole: %+v", scopeMap)
		}
	})

	t.Run("RBACCustomerForbidden", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		roleID, _ := createFixtureRole(t, admin, cfg, "aur_rbac")
		cust, custClient := createFixtureCustomerAndLogin(t, admin, cfg)
		requireCustomerMutationForbidden(t, custClient, m, map[string]any{
			"input": map[string]any{"user_id": cust.ID, "role_id": roleID, "scope_type": "GLOBAL"},
		}, "assignUserRole")
	})
}

// TestMutationRevokeUserRole menguji Mutation revokeUserRole dan verifikasi terhapus dari userRoles.
func TestMutationRevokeUserRole(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	m := fmt.Sprintf(`mutation ($assignment_id: ID!) { revokeUserRole(assignment_id: $assignment_id) { %s } }`, gqlRevokeUserRoleFields)

	t.Run("RevokeReturnsTrue", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		emp := createFixtureUser(t, admin, cfg, "EMPLOYEE", false)
		roleID, _ := createFixtureRole(t, admin, cfg, "rur")
		assignObj := requireMutationChildOK(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($in: AssignUserRoleInput!) {
			assignUserRole(input: $in) { %s }
		}`, gqlUserRolePayloadFields), map[string]any{"in": map[string]any{"user_id": emp.ID, "role_id": roleID, "scope_type": "GLOBAL"}}), "assignUserRole", "user_role_edge")
		assertUserRoleObject(t, assignObj)
		assignID := requireNonEmptyString(t, assignObj, "assignment_id")

		requireMutationBoolOK(t, doGQLOK(t, admin, m, map[string]any{"assignment_id": assignID}), "revokeUserRole", "revoked")
		urConn := requireOpMap(t, doGQLOK(t, admin, fmt.Sprintf(`query ($u: ID!) { userRoles(user_id: $u, first: 10) { edges { %s } pageInfo { %s } total_count } }`, gqlUserRoleEdgeFields, gqlPageInfoFields), map[string]any{"u": emp.ID}), "userRoles")
		_ = assertPageInfoNonNull(t, urConn)
		rem := requireConnectionEdges(t, urConn)
		for _, raw := range rem {
			if edge, ok := raw.(map[string]any); ok && edge["assignment_id"] == assignID {
				t.Fatalf("expected revoked assignment %q to be absent from userRoles, got %+v", assignID, rem)
			}
		}
	})

	t.Run("RBACCustomerForbidden", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		_, custClient := createFixtureCustomerAndLogin(t, admin, cfg)
		requireCustomerMutationForbidden(t, custClient, m, map[string]any{"assignment_id": "00000000-0000-4000-8000-000000000001"}, "revokeUserRole")
	})
}

// TestMutationCreateGroup menguji Mutation createGroup untuk pembuatan grup valid, konflik duplikasi kode, dan RBAC.
func TestMutationCreateGroup(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	m := fmt.Sprintf(`mutation ($input: CreateGroupInput!) { createGroup(input: $input) { %s } }`, gqlGroupPayloadFields)

	t.Run("CreateValidAndDuplicateConflict", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		_, code := createFixtureGroup(t, admin, cfg, "cg")

		_, _ = doMutationFail(t, admin, m, map[string]any{
			"input": map[string]any{"code": code, "name": "Duplicate", "description": "dup"},
		}, "createGroup", "CONFLICT")
	})

	t.Run("RBACCustomerForbidden", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		_, custClient := createFixtureCustomerAndLogin(t, admin, cfg)
		requireCustomerMutationForbidden(t, custClient, m, map[string]any{
			"input": map[string]any{"code": uniqueTag("no_g"), "name": "No"},
		}, "createGroup")
	})
}

// TestMutationUpdateGroupDetails menguji Mutation updateGroupDetails beserta verifikasi perubahan nama dan deskripsi.
func TestMutationUpdateGroupDetails(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	m := fmt.Sprintf(`mutation ($input: UpdateGroupDetailsInput!) { updateGroupDetails(input: $input) { %s } }`, gqlGroupPayloadFields)

	t.Run("UpdateNameAndDescription", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		groupID, _ := createFixtureGroup(t, admin, cfg, "ugd")
		newName := "Updated Group " + uniqueTag("g")

		grp := requireMutationChildOK(t, doGQLOK(t, admin, m, map[string]any{
			"input": map[string]any{"id": groupID, "name": newName, "description": "Updated E2E desc"},
		}), "updateGroupDetails", "group")
		assertGroupObject(t, grp)
		if grp["name"] != newName || grp["description"] != "Updated E2E desc" {
			t.Fatalf("unexpected updated group: %+v", grp)
		}
	})

	t.Run("RBACCustomerForbidden", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		groupID, _ := createFixtureGroup(t, admin, cfg, "ugd_rbac")
		_, custClient := createFixtureCustomerAndLogin(t, admin, cfg)
		requireCustomerMutationForbidden(t, custClient, m, map[string]any{"input": map[string]any{"id": groupID, "name": "No"}}, "updateGroupDetails")
	})
}

// TestMutationActivateGroup menguji Mutation activateGroup pada grup yang sebelumnya dinonaktifkan.
func TestMutationActivateGroup(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	m := fmt.Sprintf(`mutation ($id: ID!) { activateGroup(id: $id) { %s } }`, gqlGroupPayloadFields)

	t.Run("ActivateDeactivatedGroup", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		groupID, _ := createFixtureGroup(t, admin, cfg, "act_g")
		deactData := doGQLOK(t, admin, fmt.Sprintf(`mutation ($id: ID!) { deactivateGroup(id: $id) { %s } }`, gqlGroupPayloadFields), map[string]any{"id": groupID})
		assertGroupObject(t, requireMutationChildOK(t, deactData, "deactivateGroup", "group"))

		grp := requireMutationChildOK(t, doGQLOK(t, admin, m, map[string]any{"id": groupID}), "activateGroup", "group")
		assertGroupObject(t, grp)
		if !requireBoolField(t, grp, "is_active") {
			t.Fatalf("expected is_active=true after activateGroup")
		}
	})

	t.Run("RBACCustomerForbidden", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		groupID, _ := createFixtureGroup(t, admin, cfg, "act_g_rbac")
		_, custClient := createFixtureCustomerAndLogin(t, admin, cfg)
		requireCustomerMutationForbidden(t, custClient, m, map[string]any{"id": groupID}, "activateGroup")
	})
}

// TestMutationDeactivateGroup menguji Mutation deactivateGroup pada grup aktif dan RBAC Customer.
func TestMutationDeactivateGroup(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	m := fmt.Sprintf(`mutation ($id: ID!) { deactivateGroup(id: $id) { %s } }`, gqlGroupPayloadFields)

	t.Run("DeactivateActiveGroup", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		groupID, _ := createFixtureGroup(t, admin, cfg, "deact_g")

		grp := requireMutationChildOK(t, doGQLOK(t, admin, m, map[string]any{"id": groupID}), "deactivateGroup", "group")
		assertGroupObject(t, grp)
		if requireBoolField(t, grp, "is_active") {
			t.Fatalf("expected is_active=false after deactivateGroup")
		}
	})

	t.Run("RBACCustomerForbidden", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		groupID, _ := createFixtureGroup(t, admin, cfg, "deact_g_rbac")
		_, custClient := createFixtureCustomerAndLogin(t, admin, cfg)
		requireCustomerMutationForbidden(t, custClient, m, map[string]any{"id": groupID}, "deactivateGroup")
	})
}

// TestMutationDeleteGroup menguji Mutation deleteGroup dan verifikasi NOT_FOUND sesudah penghapusan.
func TestMutationDeleteGroup(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	m := fmt.Sprintf(`mutation ($id: ID!) { deleteGroup(id: $id) { %s } }`, gqlDeleteGroupFields)

	t.Run("DeleteGroupReturnsTrue", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		groupID, _ := createFixtureGroup(t, admin, cfg, "del_g")

		requireMutationBoolOK(t, doGQLOK(t, admin, m, map[string]any{"id": groupID}), "deleteGroup", "deleted")
		_ = doGQLError(t, admin, fmt.Sprintf(`query ($id: ID!) { group(id: $id) { %s } }`, gqlGroupFields), map[string]any{"id": groupID}, "NOT_FOUND")
	})

	t.Run("RBACCustomerForbidden", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		groupID, _ := createFixtureGroup(t, admin, cfg, "del_g_rbac")
		_, custClient := createFixtureCustomerAndLogin(t, admin, cfg)
		requireCustomerMutationForbidden(t, custClient, m, map[string]any{"id": groupID}, "deleteGroup")
	})
}

// TestMutationAddGroupMember menguji Mutation addGroupMember dan deteksi RBAC Customer.
func TestMutationAddGroupMember(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	m := fmt.Sprintf(`mutation ($group_id: ID!, $user_id: ID!) {
		addGroupMember(group_id: $group_id, user_id: $user_id) { %s }
	}`, gqlGroupMemberPayloadFields)

	t.Run("AddMemberReturnsNonNull", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		groupID, _ := createFixtureGroup(t, admin, cfg, "agm")
		emp := createFixtureUser(t, admin, cfg, "EMPLOYEE", false)

		gm := requireMutationChildOK(t, doGQLOK(t, admin, m, map[string]any{"group_id": groupID, "user_id": emp.ID}), "addGroupMember", "group_member_edge")
		assertGroupMemberObject(t, gm)
		node := requireChildMap(t, gm, "node")
		if gm["group_id"] != groupID || node["id"] != emp.ID {
			t.Fatalf("unexpected GroupMemberEdge: %+v", gm)
		}
	})

	t.Run("RBACCustomerForbidden", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		groupID, _ := createFixtureGroup(t, admin, cfg, "agm_rbac")
		cust, custClient := createFixtureCustomerAndLogin(t, admin, cfg)
		requireCustomerMutationForbidden(t, custClient, m, map[string]any{"group_id": groupID, "user_id": cust.ID}, "addGroupMember")
	})
}

// TestMutationRemoveGroupMember menguji Mutation removeGroupMember serta kekosongan daftar groupMembers.
func TestMutationRemoveGroupMember(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	m := fmt.Sprintf(`mutation ($group_id: ID!, $user_id: ID!) { removeGroupMember(group_id: $group_id, user_id: $user_id) { %s } }`, gqlRemoveGroupMemberFields)

	t.Run("RemoveMemberReturnsTrue", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		groupID, _ := createFixtureGroup(t, admin, cfg, "rgm")
		emp := createFixtureUser(t, admin, cfg, "EMPLOYEE", false)
		gmData := doGQLOK(t, admin, fmt.Sprintf(`mutation ($g: ID!, $u: ID!) { addGroupMember(group_id: $g, user_id: $u) { %s } }`, gqlGroupMemberPayloadFields), map[string]any{
			"g": groupID, "u": emp.ID,
		})
		assertGroupMemberObject(t, requireMutationChildOK(t, gmData, "addGroupMember", "group_member_edge"))

		requireMutationBoolOK(t, doGQLOK(t, admin, m, map[string]any{"group_id": groupID, "user_id": emp.ID}), "removeGroupMember", "removed")
		conn := requireOpMap(t, doGQLOK(t, admin, fmt.Sprintf(`query ($g: ID!) { groupMembers(group_id: $g, first: 10) { nodes { %s } pageInfo { %s } total_count } }`, gqlUserFields, gqlPageInfoFields), map[string]any{"g": groupID}), "groupMembers")
		_ = assertPageInfoNonNull(t, conn)
		if len(requireConnectionNodes(t, conn)) != 0 {
			t.Fatalf("expected 0 remaining groupMembers")
		}
	})

	t.Run("RBACCustomerForbidden", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		groupID, _ := createFixtureGroup(t, admin, cfg, "rgm_rbac")
		cust, custClient := createFixtureCustomerAndLogin(t, admin, cfg)
		requireCustomerMutationForbidden(t, custClient, m, map[string]any{"group_id": groupID, "user_id": cust.ID}, "removeGroupMember")
	})
}

// TestMutationAssignGroupRole menguji Mutation assignGroupRole (GLOBAL dan FACILITY + expires_at) serta pewarisan anggota.
func TestMutationAssignGroupRole(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	m := fmt.Sprintf(`mutation ($input: AssignGroupRoleInput!) {
		assignGroupRole(input: $input) { %s }
	}`, gqlGroupRolePayloadFields)

	t.Run("AssignScopedGroupRoleNonNullAndInherited", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		groupID, _ := createFixtureGroup(t, admin, cfg, "agr")
		roleID, roleCode := createFixtureRole(t, admin, cfg, "agr")
		emp := createFixtureUser(t, admin, cfg, "EMPLOYEE", false)
		gmData := doGQLOK(t, admin, fmt.Sprintf(`mutation ($g: ID!, $u: ID!) { addGroupMember(group_id: $g, user_id: $u) { %s } }`, gqlGroupMemberPayloadFields), map[string]any{
			"g": groupID, "u": emp.ID,
		})
		assertGroupMemberObject(t, requireMutationChildOK(t, gmData, "addGroupMember", "group_member_edge"))

		scopeID := createFixtureEFFacility(t, admin, cfg)
		gr := requireMutationChildOK(t, doGQLOK(t, admin, m, map[string]any{
			"input": map[string]any{
				"group_id":   groupID,
				"role_id":    roleID,
				"scope_type": "FACILITY",
				"scope_id":   scopeID,
				"expires_at": time.Now().Add(4 * time.Hour).UTC().Format(time.RFC3339),
			},
		}), "assignGroupRole", "group_role_edge")
		assertGroupRoleObject(t, gr)
		_ = requireNonEmptyString(t, gr, "granted_by")
		_ = requireNonEmptyString(t, gr, "expires_at")
		scopeMap := requireChildMap(t, gr, "scope")
		if scopeMap["type"] != "FACILITY" || scopeMap["id"] != scopeID {
			t.Fatalf("unexpected group role scope: %+v", scopeMap)
		}

		effConn := requireOpMap(t, doGQLOK(t, admin, fmt.Sprintf(`query ($u: ID!) { userEffectiveRoles(user_id: $u, first: 20) { nodes { %s } pageInfo { %s } total_count } }`, gqlRoleFields, gqlPageInfoFields), map[string]any{"u": emp.ID}), "userEffectiveRoles")
		_ = assertPageInfoNonNull(t, effConn)
		eff := requireConnectionNodes(t, effConn)
		var inherited bool
		for _, raw := range eff {
			roleObj := raw.(map[string]any)
			assertRoleObject(t, roleObj)
			if roleObj["code"] == roleCode {
				inherited = true
			}
		}
		if !inherited {
			t.Fatalf("expected member to inherit group role %q", roleCode)
		}
	})

	t.Run("RBACCustomerForbidden", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		groupID, _ := createFixtureGroup(t, admin, cfg, "agr_rbac")
		roleID, _ := createFixtureRole(t, admin, cfg, "agr_rbac")
		_, custClient := createFixtureCustomerAndLogin(t, admin, cfg)
		requireCustomerMutationForbidden(t, custClient, m, map[string]any{
			"input": map[string]any{"group_id": groupID, "role_id": roleID, "scope_type": "GLOBAL"},
		}, "assignGroupRole")
	})
}

// TestMutationRevokeGroupRole menguji Mutation revokeGroupRole serta penghapusan dari groupRoles.
func TestMutationRevokeGroupRole(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	m := fmt.Sprintf(`mutation ($assignment_id: ID!) { revokeGroupRole(assignment_id: $assignment_id) { %s } }`, gqlRevokeGroupRoleFields)

	t.Run("RevokeGroupRoleReturnsTrue", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		groupID, _ := createFixtureGroup(t, admin, cfg, "rgr")
		roleID, _ := createFixtureRole(t, admin, cfg, "rgr")
		assignObj := requireMutationChildOK(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($in: AssignGroupRoleInput!) {
			assignGroupRole(input: $in) { %s }
		}`, gqlGroupRolePayloadFields), map[string]any{"in": map[string]any{"group_id": groupID, "role_id": roleID, "scope_type": "GLOBAL"}}), "assignGroupRole", "group_role_edge")
		assertGroupRoleObject(t, assignObj)
		assignID := requireNonEmptyString(t, assignObj, "assignment_id")

		requireMutationBoolOK(t, doGQLOK(t, admin, m, map[string]any{"assignment_id": assignID}), "revokeGroupRole", "revoked")
		grConn := requireOpMap(t, doGQLOK(t, admin, fmt.Sprintf(`query ($g: ID!) { groupRoles(group_id: $g, first: 10) { edges { %s } pageInfo { %s } total_count } }`, gqlGroupRoleEdgeFields, gqlPageInfoFields), map[string]any{"g": groupID}), "groupRoles")
		_ = assertPageInfoNonNull(t, grConn)
		rem := requireConnectionEdges(t, grConn)
		if len(rem) != 0 {
			t.Fatalf("expected 0 groupRoles remaining, got %+v", rem)
		}
	})

	t.Run("RBACCustomerForbidden", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		_, custClient := createFixtureCustomerAndLogin(t, admin, cfg)
		requireCustomerMutationForbidden(t, custClient, m, map[string]any{"assignment_id": "00000000-0000-4000-8000-000000000001"}, "revokeGroupRole")
	})
}

// TestMutationRevokeSession menguji Mutation revokeSession memastikan return revoked=true, pencabutan sesi aktif, serta ended_at/end_reason.
func TestMutationRevokeSession(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	m := fmt.Sprintf(`mutation ($session_id: ID!, $reason: SessionEndReason) { revokeSession(session_id: $session_id, reason: $reason) { %s } }`, gqlRevokeSessionFields)

	t.Run("RevokeTargetSessionInvalidatesClient", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		emp := createFixtureUser(t, admin, cfg, "EMPLOYEE", false)
		empClient, _ := loginFixtureUser(t, cfg, emp.Email, emp.Password)

		sessConn := requireOpMap(t, doGQLOK(t, admin, fmt.Sprintf(`query ($u: ID!) { userSessions(user_id: $u, active_only: true, first: 10) { nodes { %s } pageInfo { %s } total_count } }`, gqlSessionFields, gqlPageInfoFields), map[string]any{"u": emp.ID}), "userSessions")
		_ = assertPageInfoNonNull(t, sessConn)
		items := requireConnectionNodes(t, sessConn)
		activeSess := items[0].(map[string]any)
		assertSessionObject(t, activeSess)
		sessID := requireNonEmptyString(t, activeSess, "id")

		requireMutationBoolOK(t, doGQLOK(t, admin, m, map[string]any{"session_id": sessID, "reason": "ADMIN_REVOKED"}), "revokeSession", "revoked")

		_ = doGQLError(t, empClient, fmt.Sprintf(`query { me { %s } }`, gqlUserFields), nil, "UNAUTHENTICATED")

		histConn := requireOpMap(t, doGQLOK(t, admin, fmt.Sprintf(`query ($u: ID!) { userSessions(user_id: $u, active_only: false, first: 10) { nodes { %s } pageInfo { %s } total_count } }`, gqlSessionFields, gqlPageInfoFields), map[string]any{"u": emp.ID}), "userSessions")
		_ = assertPageInfoNonNull(t, histConn)
		hist := requireConnectionNodes(t, histConn)[0].(map[string]any)
		assertSessionObject(t, hist)
		_ = requireNonEmptyString(t, hist, "ended_at")
		_ = requireNonEmptyString(t, hist, "end_reason")
		if hist["end_reason"] != "ADMIN_REVOKED" {
			t.Fatalf("end_reason = %v, want ADMIN_REVOKED", hist["end_reason"])
		}
	})

	t.Run("RBACCustomerForbidden", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		_, custClient := createFixtureCustomerAndLogin(t, admin, cfg)
		requireCustomerMutationForbidden(t, custClient, m, map[string]any{
			"session_id": "00000000-0000-4000-8000-000000000001", "reason": "ADMIN_REVOKED",
		}, "revokeSession")
	})
}

// TestMutationRevokeAllUserSessions menguji Mutation revokeAllUserSessions pada multi-sesi bersamaan dan verifikasi jumlah sesi tercabut.
func TestMutationRevokeAllUserSessions(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	m := fmt.Sprintf(`mutation ($user_id: ID!, $reason: SessionEndReason) { revokeAllUserSessions(user_id: $user_id, reason: $reason) { %s } }`, gqlRevokeAllSessionsFields)

	t.Run("RevokeAllActiveSessionsReturnsCountAndInvalidatesAll", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		emp := createFixtureUser(t, admin, cfg, "EMPLOYEE", false)
		c1, _ := loginFixtureUser(t, cfg, emp.Email, emp.Password)
		c2, _ := loginFixtureUser(t, cfg, emp.Email, emp.Password)

		revPayload := requireMutationSuccess(t, doGQLOK(t, admin, m, map[string]any{
			"user_id": emp.ID, "reason": "ADMIN_REVOKED",
		}), "revokeAllUserSessions")
		count := requireNumberField(t, revPayload, "revoked_count")
		if count < 2 {
			t.Fatalf("revokeAllUserSessions count = %d, want >= 2", count)
		}

		for _, cl := range []*browserClient{c1, c2} {
			_ = doGQLError(t, cl, fmt.Sprintf(`query { me { %s } }`, gqlUserFields), nil, "UNAUTHENTICATED")
		}
	})

	t.Run("RBACCustomerForbidden", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		cust, custClient := createFixtureCustomerAndLogin(t, admin, cfg)
		requireCustomerMutationForbidden(t, custClient, m, map[string]any{
			"user_id": cust.ID, "reason": "ADMIN_REVOKED",
		}, "revokeAllUserSessions")
	})
}

// TestMutationResendValidationEmailTokenAlreadyVerifiedAndNeutral menguji kasus (1) pengguna login tanpa pending email yang emailnya sudah terverifikasi mendapat error CONFLICT "your email has been verified", sedangkan kasus (2) anonim tetap netral.
func TestMutationResendValidationEmailTokenAlreadyVerifiedAndNeutral(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	mResend := fmt.Sprintf(`mutation ($input: ResendValidationEmailInput) { resendValidationEmailToken(input: $input) { %s } }`, gqlResendValidationFields)
	mUpdEmail := fmt.Sprintf(`mutation ($input: UserEmailInput!) { updateUserEmail(input: $input) { %s } }`, gqlUserEmailPayloadFields)
	mDelPending := fmt.Sprintf(`mutation { deletePendingUserEmail { %s } }`, gqlDeletePendingEmailFields)

	t.Run("LoggedInVerifiedWithoutPendingReturnsConflictAndCreatesNoToken", func(t *testing.T) {
		resetGatewayRateLimits(t, cfg)
		admin := newSuperAdminBrowser(t, cfg)
		emp := createFixtureUser(t, admin, cfg, "EMPLOYEE", true)
		empClient, _ := loginFixtureUser(t, cfg, emp.Email, emp.Password)
		empRawID := rawUUID(emp.ID)
		tokBefore := psqlScalar(t, cfg, fmt.Sprintf(`SELECT count(*) FROM user_tokens WHERE user_id = %s::uuid AND purpose = 'email_verification';`, sqlQuote(empRawID)))

		// Sebelum ada pending email, kirim ulang dengan atau tanpa input.email harus menolak dengan CONFLICT "your email has been verified".
		for _, vars := range []map[string]any{
			{},
			{"input": map[string]any{"email": emp.Email}},
		} {
			_, errMsg := doMutationFail(t, empClient, mResend, vars, "resendValidationEmailToken", "CONFLICT")
			if errMsg != "your email has been verified" {
				t.Fatalf("resendValidationEmailToken(%+v) error message = %q, want %q", vars, errMsg, "your email has been verified")
			}
		}

		tokAfter := psqlScalar(t, cfg, fmt.Sprintf(`SELECT count(*) FROM user_tokens WHERE user_id = %s::uuid AND purpose = 'email_verification';`, sqlQuote(empRawID)))
		if tokAfter != tokBefore {
			t.Fatalf("expected 0 new email_verification tokens created when already verified (before=%s after=%s)", tokBefore, tokAfter)
		}

		// Setelah menambahkan pending email, kasus (1) berhasil menerbitkan token baru, dan sesudah deletePendingUserEmail kembali mengembalikan CONFLICT.
		pendingEmail := uniqueTag("pending-resend") + "@integrione.test"
		_ = requireMutationSuccess(t, doGQLOK(t, empClient, mUpdEmail, map[string]any{
			"input": map[string]any{"email": pendingEmail, "current_password": emp.Password},
		}), "updateUserEmail")

		_ = requireMutationBoolOK(t, doGQLOK(t, empClient, mResend, map[string]any{}), "resendValidationEmailToken", "is_resend")

		_ = requireMutationBoolOK(t, doGQLOK(t, empClient, mDelPending, nil), "deletePendingUserEmail", "deleted")

		_, afterDelMsg := doMutationFail(t, empClient, mResend, map[string]any{}, "resendValidationEmailToken", "CONFLICT")
		if afterDelMsg != "your email has been verified" {
			t.Fatalf("after deletePendingUserEmail resend message = %q, want %q", afterDelMsg, "your email has been verified")
		}
	})

	t.Run("UnauthenticatedAlreadyVerifiedEmailRemainsNeutral", func(t *testing.T) {
		resetGatewayRateLimits(t, cfg)
		admin := newSuperAdminBrowser(t, cfg)
		emp := createFixtureUser(t, admin, cfg, "EMPLOYEE", true)
		anon := newAnonBrowser(t, cfg)

		_ = requireMutationBoolOK(t, doGQLOK(t, anon, mResend, map[string]any{
			"input": map[string]any{"email": emp.Email},
		}), "resendValidationEmailToken", "is_resend")
	})
}

type businessUserMutationSpec struct {
	opName   string
	wantKind string
	query    string
}

func businessUserMutationSpecs() []businessUserMutationSpec {
	return []businessUserMutationSpec{
		{
			opName:   "createEmployeeUser",
			wantKind: "EMPLOYEE",
			query:    fmt.Sprintf(`mutation ($input: CreateEmployeeUserInput!) { createEmployeeUser(input: $input) { %s } }`, gqlUserPayloadFields),
		},
		{
			opName:   "createVendorUser",
			wantKind: "VENDOR",
			query:    fmt.Sprintf(`mutation ($input: CreateVendorUserInput!) { createVendorUser(input: $input) { %s } }`, gqlUserPayloadFields),
		},
		{
			opName:   "createSupplierUser",
			wantKind: "SUPPLIER",
			query:    fmt.Sprintf(`mutation ($input: CreateSupplierUserInput!) { createSupplierUser(input: $input) { %s } }`, gqlUserPayloadFields),
		},
	}
}

func assertBusinessUserCreationDeniedAndZeroDBInsert(
	t *testing.T,
	cfg e2eRuntimeConfig,
	client *browserClient,
	spec businessUserMutationSpec,
	wantErrType string,
	probeEmail string,
) {
	t.Helper()
	tag := uniqueTag("deny_" + strings.ToLower(spec.wantKind))
	targetEmail := probeEmail
	expectZeroAfter := strings.TrimSpace(probeEmail) == ""
	if expectZeroAfter {
		targetEmail = tag + "@integrione.test"
	}
	inVars := map[string]any{
		"full_name": "Denied " + tag,
		"email":     targetEmail,
		"username":  tag,
		"password":  "StrongPass#2026!",
	}
	if spec.wantKind == "VENDOR" || spec.wantKind == "SUPPLIER" {
		inVars["party_id"] = "11111111-1111-4111-8111-111111111111"
	}
	if spec.wantKind == "EMPLOYEE" {
		inVars["worker_id"] = "55555555-5555-4555-8555-555555555555"
		inVars["provisioning_request_id"] = "66666666-6666-4666-8666-666666666666"
	}
	vars := map[string]any{"input": inVars}

	_, _ = doMutationFail(t, client, spec.query, vars, spec.opName, wantErrType)
	if expectZeroAfter {
		cnt := psqlScalar(t, cfg, fmt.Sprintf(`SELECT count(*) FROM user_identities WHERE value = %s;`, sqlQuote(targetEmail)))
		if cnt != "0" {
			t.Fatalf("%s(%s) inserted %s identity rows for %q unexpectedly", spec.opName, wantErrType, cnt, targetEmail)
		}
	}
}

func assertBusinessUserCreationAllowedWithCleanup(
	t *testing.T,
	cfg e2eRuntimeConfig,
	admin *browserClient,
	client *browserClient,
	spec businessUserMutationSpec,
) {
	t.Helper()
	tag := uniqueTag("allow_" + strings.ToLower(spec.wantKind))
	inVars := map[string]any{
		"full_name": "Allowed " + tag,
		"email":     tag + "@integrione.test",
		"username":  tag,
		"password":  "StrongPass#2026!",
	}
	if spec.wantKind == "VENDOR" || spec.wantKind == "SUPPLIER" {
		inVars["party_id"] = createFixturePartyForRole(t, admin, cfg, spec.wantKind)
	}
	data := doGQLOK(t, client, spec.query, map[string]any{"input": inVars})
	u := requireMutationChildOK(t, data, spec.opName, "user")
	assertUserObject(t, u)
	if u["kind"] != spec.wantKind || u["status"] != "PENDING" {
		t.Fatalf("%s returned kind=%v status=%v, want %s/PENDING", spec.opName, u["kind"], u["status"], spec.wantKind)
	}
	rawID := requireNonEmptyString(t, u, "_id")
	t.Cleanup(func() {
		cleanupUserByID(t, cfg, rawID)
	})
}

// TestRBACCreateBusinessUsersStrictEnforcement memverifikasi end-to-end via HTTPS bahwa createEmployeeUser, createVendorUser, dan createSupplierUser hanya diizinkan bagi admin atau pemegang peran/izin iam.users.write berlingkup GLOBAL + sesi sensitif segar.
func TestRBACCreateBusinessUsersStrictEnforcement(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	specs := businessUserMutationSpecs()
	mAssignUserRole := fmt.Sprintf(`mutation ($input: AssignUserRoleInput!) { assignUserRole(input: $input) { %s } }`, gqlUserRolePayloadFields)
	mAssignRolePerm := fmt.Sprintf(`mutation ($role_id: ID!, $permission_id: ID!) { assignRolePermission(role_id: $role_id, permission_id: $permission_id) { %s } }`, gqlRolePermPayloadFields)
	mRevokeRolePerm := fmt.Sprintf(`mutation ($role_id: ID!, $permission_id: ID!) { revokeRolePermission(role_id: $role_id, permission_id: $permission_id) { %s } }`, gqlRevokeRolePermFields)
	mAddGroupMember := fmt.Sprintf(`mutation ($g: ID!, $u: ID!) { addGroupMember(group_id: $g, user_id: $u) { %s } }`, gqlGroupMemberPayloadFields)
	mAssignGroupRole := fmt.Sprintf(`mutation ($input: AssignGroupRoleInput!) { assignGroupRole(input: $input) { %s } }`, gqlGroupRolePayloadFields)
	mRemoveGroupMember := fmt.Sprintf(`mutation ($g: ID!, $u: ID!) { removeGroupMember(group_id: $g, user_id: $u) { %s } }`, gqlRemoveGroupMemberFields)

	t.Run("AnonymousRejectedAuthenticationAndZeroAntiOracle", func(t *testing.T) {
		resetGatewayRateLimits(t, cfg)
		admin := newSuperAdminBrowser(t, cfg)
		existing := createFixtureUser(t, admin, cfg, "EMPLOYEE", true)
		anon := newAnonBrowser(t, cfg)

		for _, sp := range specs {
			assertBusinessUserCreationDeniedAndZeroDBInsert(t, cfg, anon, sp, "AUTHENTICATION", "")
			// Percobaan menggunakan email yang sudah terdaftar tetap ditolak AUTHENTICATION, bukan CONFLICT oracle.
			assertBusinessUserCreationDeniedAndZeroDBInsert(t, cfg, anon, sp, "AUTHENTICATION", existing.Email)
		}
	})

	t.Run("CustomerEmployeeWithoutRoleAndReadOnlyRoleForbidden", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		_, custClient := createFixtureCustomerAndLogin(t, admin, cfg)

		plainEmp := createFixtureUser(t, admin, cfg, "EMPLOYEE", true)
		plainEmpClient, _ := loginFixtureUser(t, cfg, plainEmp.Email, plainEmp.Password)

		roEmp := createFixtureUser(t, admin, cfg, "EMPLOYEE", true)
		roRole := requireOpMap(t, doGQLOK(t, admin, fmt.Sprintf(`query { roleByCode(code: "iam_readonly") { %s } }`, gqlRoleFields), nil), "roleByCode")
		roRoleID := requireNonEmptyString(t, roRole, "id")
		_ = requireMutationChildOK(t, doGQLOK(t, admin, mAssignUserRole, map[string]any{
			"input": map[string]any{"user_id": roEmp.ID, "role_id": roRoleID, "scope_type": "GLOBAL"},
		}), "assignUserRole", "user_role_edge")
		roEmpClient, _ := loginFixtureUser(t, cfg, roEmp.Email, roEmp.Password)

		for _, cl := range []*browserClient{custClient, plainEmpClient, roEmpClient} {
			for _, sp := range specs {
				assertBusinessUserCreationDeniedAndZeroDBInsert(t, cfg, cl, sp, "AUTHORIZATION", "")
				assertBusinessUserCreationDeniedAndZeroDBInsert(t, cfg, cl, sp, "AUTHORIZATION", plainEmp.Email)
			}
		}
	})

	t.Run("NonGlobalScopedUsersWriteForbidden", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		permWriteID := findPermissionIDByCode(t, admin, "iam.users.write")
		scopedRoleID, _ := createFixtureRole(t, admin, cfg, "scoped_user_w")
		_ = requireMutationChildOK(t, doGQLOK(t, admin, mAssignRolePerm, map[string]any{
			"role_id": scopedRoleID, "permission_id": permWriteID,
		}), "assignRolePermission", "role_permission_edge")

		facilityID := createFixtureEFFacility(t, admin, cfg)
		scopedEmp := createFixtureUser(t, admin, cfg, "EMPLOYEE", true)
		_ = requireMutationChildOK(t, doGQLOK(t, admin, mAssignUserRole, map[string]any{
			"input": map[string]any{
				"user_id":    scopedEmp.ID,
				"role_id":    scopedRoleID,
				"scope_type": "FACILITY",
				"scope_id":   facilityID,
			},
		}), "assignUserRole", "user_role_edge")

		scopedClient, _ := loginFixtureUser(t, cfg, scopedEmp.Email, scopedEmp.Password)
		for _, sp := range specs {
			assertBusinessUserCreationDeniedAndZeroDBInsert(t, cfg, scopedClient, sp, "AUTHORIZATION", "")
		}
	})

	t.Run("DirectCustomRoleGlobalAllowedThenRevokedAndStaleBlocked", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		permWriteID := findPermissionIDByCode(t, admin, "iam.users.write")
		efPartyWriteID := findPermissionIDByCode(t, admin, "enterprise_foundation.parties.write")
		customRoleID, _ := createFixtureRole(t, admin, cfg, "global_user_w")
		_ = requireMutationChildOK(t, doGQLOK(t, admin, mAssignRolePerm, map[string]any{
			"role_id": customRoleID, "permission_id": permWriteID,
		}), "assignRolePermission", "role_permission_edge")

		emp := createFixtureUser(t, admin, cfg, "EMPLOYEE", true)
		_ = requireMutationChildOK(t, doGQLOK(t, admin, mAssignUserRole, map[string]any{
			"input": map[string]any{"user_id": emp.ID, "role_id": customRoleID, "scope_type": "GLOBAL"},
		}), "assignUserRole", "user_role_edge")

		empClient, loginRes := loginFixtureUser(t, cfg, emp.Email, emp.Password)
		assertUserLoginPayload(t, loginRes, "AUTHENTICATED")

		// Pemegang iam.users.write tanpa enterprise_foundation.parties.write ditolak AUTHORIZATION untuk createVendorUser & createSupplierUser.
		for _, sp := range specs[1:] {
			assertBusinessUserCreationDeniedAndZeroDBInsert(t, cfg, empClient, sp, "AUTHORIZATION", "")
		}

		_ = requireMutationChildOK(t, doGQLOK(t, admin, mAssignRolePerm, map[string]any{
			"role_id": customRoleID, "permission_id": efPartyWriteID,
		}), "assignRolePermission", "role_permission_edge")

		for _, sp := range specs {
			assertBusinessUserCreationAllowedWithCleanup(t, cfg, admin, empClient, sp)
		}

		// Bila izin dicabut dari role tanpa logout, permintaan berikutnya wajib langsung ditolak AUTHORIZATION.
		requireMutationBoolOK(t, doGQLOK(t, admin, mRevokeRolePerm, map[string]any{
			"role_id": customRoleID, "permission_id": permWriteID,
		}), "revokeRolePermission", "revoked")
		for _, sp := range specs {
			assertBusinessUserCreationDeniedAndZeroDBInsert(t, cfg, empClient, sp, "AUTHORIZATION", "")
		}

		// Beri izin kembali dan uji penolakan ketika sesi melewati batas jendela sensitif 10 menit.
		_ = requireMutationChildOK(t, doGQLOK(t, admin, mAssignRolePerm, map[string]any{
			"role_id": customRoleID, "permission_id": permWriteID,
		}), "assignRolePermission", "role_permission_edge")
		backdateSessionForStaleSensitiveTest(t, cfg, empClient, emp.ID)
		for _, sp := range specs {
			assertBusinessUserCreationDeniedAndZeroDBInsert(t, cfg, empClient, sp, "AUTHORIZATION", "")
		}
	})

	t.Run("GroupInheritedGlobalPermissionAllowedThenMemberRemovedBlocks", func(t *testing.T) {
		admin := newSuperAdminBrowser(t, cfg)
		permWriteID := findPermissionIDByCode(t, admin, "iam.users.write")
		efPartyWriteID := findPermissionIDByCode(t, admin, "enterprise_foundation.parties.write")
		grpRoleID, _ := createFixtureRole(t, admin, cfg, "grp_user_w")
		_ = requireMutationChildOK(t, doGQLOK(t, admin, mAssignRolePerm, map[string]any{
			"role_id": grpRoleID, "permission_id": permWriteID,
		}), "assignRolePermission", "role_permission_edge")
		_ = requireMutationChildOK(t, doGQLOK(t, admin, mAssignRolePerm, map[string]any{
			"role_id": grpRoleID, "permission_id": efPartyWriteID,
		}), "assignRolePermission", "role_permission_edge")

		groupID, _ := createFixtureGroup(t, admin, cfg, "grp_user_w")
		_ = requireMutationChildOK(t, doGQLOK(t, admin, mAssignGroupRole, map[string]any{
			"input": map[string]any{"group_id": groupID, "role_id": grpRoleID, "scope_type": "GLOBAL"},
		}), "assignGroupRole", "group_role_edge")

		emp := createFixtureUser(t, admin, cfg, "EMPLOYEE", true)
		_ = requireMutationChildOK(t, doGQLOK(t, admin, mAddGroupMember, map[string]any{
			"g": groupID, "u": emp.ID,
		}), "addGroupMember", "group_member_edge")

		empClient, loginRes := loginFixtureUser(t, cfg, emp.Email, emp.Password)
		assertUserLoginPayload(t, loginRes, "AUTHENTICATED")
		for _, sp := range specs {
			assertBusinessUserCreationAllowedWithCleanup(t, cfg, admin, empClient, sp)
		}

		// Setelah dikeluarkan dari grup tanpa logout, snapshot langsung tersinkron dan akses ditolak AUTHORIZATION.
		requireMutationBoolOK(t, doGQLOK(t, admin, mRemoveGroupMember, map[string]any{
			"g": groupID, "u": emp.ID,
		}), "removeGroupMember", "removed")
		for _, sp := range specs {
			assertBusinessUserCreationDeniedAndZeroDBInsert(t, cfg, empClient, sp, "AUTHORIZATION", "")
		}
	})
}

// TestUserIDAcceptsGIDAndRawUUID memverifikasi seluruh input user_id (query top-level, nested User.*, dan mutation) menerima format GID "gid://integrione/user/{uuid}" maupun UUID mentah "{uuid}" (lowercase & uppercase).
func TestUserIDAcceptsGIDAndRawUUID(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	admin := newSuperAdminBrowser(t, cfg)
	emp := createFixtureUser(t, admin, cfg, "EMPLOYEE", true)
	empClient, _ := loginFixtureUser(t, cfg, emp.Email, emp.Password)

	roleID, _ := createFixtureRole(t, admin, cfg, "uuid_norm_role")
	groupID, _ := createFixtureGroup(t, admin, cfg, "uuid_norm_grp")

	rawLower := emp.RawID
	rawUpper := strings.ToUpper(emp.RawID)

	// 1. Mutation assignment & membership menggunakan UUID mentah (uppercase & lowercase) serta GID.
	mAssignRole := fmt.Sprintf(`mutation ($in: AssignUserRoleInput!) { assignUserRole(input: $in) { %s } }`, gqlUserRolePayloadFields)
	urEdge := requireMutationChildOK(t, doGQLOK(t, admin, mAssignRole, map[string]any{
		"in": map[string]any{"user_id": rawUpper, "role_id": roleID, "scope_type": "GLOBAL"},
	}), "assignUserRole", "user_role_edge")
	if urEdge["user_id"] != emp.ID {
		t.Fatalf("assignUserRole(rawUpper) user_id=%v, want %q", urEdge["user_id"], emp.ID)
	}

	mAddMember := fmt.Sprintf(`mutation ($g: ID!, $u: ID!) { addGroupMember(group_id: $g, user_id: $u) { %s } }`, gqlGroupMemberPayloadFields)
	gmEdge := requireMutationChildOK(t, doGQLOK(t, admin, mAddMember, map[string]any{
		"g": groupID, "u": rawLower,
	}), "addGroupMember", "group_member_edge")
	if requireChildMap(t, gmEdge, "node")["id"] != emp.ID {
		t.Fatalf("addGroupMember(rawLower) node.id=%v, want %q", requireChildMap(t, gmEdge, "node")["id"], emp.ID)
	}

	// 2. Seluruh Query top-level dan nested resolvers pada User menerima GID maupun UUID mentah (lowercase & uppercase).
	nestedUserQ := fmt.Sprintf(`query ($id: ID!) {
		user(id: $id) {
			%s
			identities(first: 10) { total_count nodes { %s } }
			roles(first: 10) { total_count edges { user_id } }
			effective_roles(first: 10) { total_count nodes { code } }
			groups(first: 10) { total_count nodes { id } }
			sessions(first: 10, active_only: true) { total_count nodes { user_id } }
		}
	}`, gqlUserFields, gqlIdentityFields)

	for _, idInput := range []string{emp.ID, rawLower, rawUpper} {
		uData := requireOpMap(t, doGQLOK(t, admin, nestedUserQ, map[string]any{"id": idInput}), "user")
		assertUserObject(t, uData)
		if uData["id"] != emp.ID || uData["_id"] != rawLower {
			t.Fatalf("user(%q) returned id=%v _id=%v, want %q/%q", idInput, uData["id"], uData["_id"], emp.ID, rawLower)
		}
		if requireNumberField(t, requireChildMap(t, uData, "identities"), "total_count") < 2 {
			t.Fatalf("user(%q).identities expected >=2, got %+v", idInput, uData["identities"])
		}
		if requireNumberField(t, requireChildMap(t, uData, "roles"), "total_count") < 1 {
			t.Fatalf("user(%q).roles expected >=1, got %+v", idInput, uData["roles"])
		}
		if requireNumberField(t, requireChildMap(t, uData, "effective_roles"), "total_count") < 1 {
			t.Fatalf("user(%q).effective_roles expected >=1, got %+v", idInput, uData["effective_roles"])
		}
		if requireNumberField(t, requireChildMap(t, uData, "groups"), "total_count") < 1 {
			t.Fatalf("user(%q).groups expected >=1, got %+v", idInput, uData["groups"])
		}
		if requireNumberField(t, requireChildMap(t, uData, "sessions"), "total_count") < 1 {
			t.Fatalf("user(%q).sessions expected >=1, got %+v", idInput, uData["sessions"])
		}
	}

	topQueries := fmt.Sprintf(`query ($gid: ID!, $rawL: ID!, $rawU: ID!) {
		uGid: userIdentities(user_id: $gid, first: 10) { total_count }
		uRawL: userIdentities(user_id: $rawL, first: 10) { total_count nodes { %s } }
		rRawU: userRoles(user_id: $rawU, first: 10) { total_count }
		erRawL: userEffectiveRoles(user_id: $rawL, first: 10) { total_count }
		gRawU: userGroups(user_id: $rawU, first: 10) { total_count }
		sRawL: userSessions(user_id: $rawL, first: 10, active_only: true) { total_count }
	}`, gqlIdentityFields)
	topRes := doGQLOK(t, admin, topQueries, map[string]any{"gid": emp.ID, "rawL": rawLower, "rawU": rawUpper})
	for _, field := range []string{"uGid", "uRawL"} {
		if requireNumberField(t, requireOpMap(t, topRes, field), "total_count") != 2 {
			t.Fatalf("%s total_count unexpected: %+v", field, topRes[field])
		}
	}
	for _, field := range []string{"rRawU", "erRawL", "gRawU", "sRawL"} {
		if requireNumberField(t, requireOpMap(t, topRes, field), "total_count") < 1 {
			t.Fatalf("%s expected >=1: %+v", field, topRes[field])
		}
	}

	// 3. Akses diri sendiri (self) lewat node(id: emp.ID) termasuk nested *ForUser berhasil tanpa hak admin.
	selfNodeQ := `query ($id: ID!) {
		node(id: $id) {
			... on User {
				id
				identities(first: 10) { total_count }
				roles(first: 10) { total_count }
				effective_roles(first: 10) { total_count }
				groups(first: 10) { total_count }
				sessions(first: 10, active_only: true) { total_count }
			}
		}
	}`
	selfNode := requireOpMap(t, doGQLOK(t, empClient, selfNodeQ, map[string]any{"id": emp.ID}), "node")
	if selfNode["id"] != emp.ID || requireNumberField(t, requireChildMap(t, selfNode, "identities"), "total_count") != 2 {
		t.Fatalf("self node nested lookup failed: %+v", selfNode)
	}

	// 4. Mutation identitas, grup, sesi, dan status user menerima UUID mentah maupun GID.
	var emailIdentID, usernameIdentID string
	for _, raw := range requireConnectionNodes(t, requireOpMap(t, topRes, "uRawL")) {
		it := raw.(map[string]any)
		switch it["kind"] {
		case "EMAIL":
			emailIdentID = requireNonEmptyString(t, it, "_id")
		case "USERNAME":
			usernameIdentID = requireNonEmptyString(t, it, "_id")
		}
	}
	if emailIdentID == "" || usernameIdentID == "" {
		t.Fatalf("expected both EMAIL and USERNAME identities, got email=%q username=%q", emailIdentID, usernameIdentID)
	}

	mSetPrim := fmt.Sprintf(`mutation ($u: ID!, $i: ID!) { setPrimaryUserIdentity(user_id: $u, identity_id: $i) { %s } }`, gqlUserIdentityPayload)
	primUser := requireMutationChildOK(t, doGQLOK(t, admin, mSetPrim, map[string]any{"u": rawUpper, "i": usernameIdentID}), "setPrimaryUserIdentity", "user_identity")
	if !requireBoolField(t, primUser, "is_primary") {
		t.Fatalf("expected USERNAME identity promoted with rawUpper user_id: %+v", primUser)
	}
	_ = requireMutationChildOK(t, doGQLOK(t, admin, mSetPrim, map[string]any{"u": emp.ID, "i": emailIdentID}), "setPrimaryUserIdentity", "user_identity")

	mRemoveIdent := fmt.Sprintf(`mutation ($u: ID!, $i: ID!) { removeUserIdentity(user_id: $u, identity_id: $i) { %s } }`, gqlRemoveIdentityFields)
	requireMutationBoolOK(t, doGQLOK(t, admin, mRemoveIdent, map[string]any{"u": rawLower, "i": usernameIdentID}), "removeUserIdentity", "removed")

	mRemoveMember := fmt.Sprintf(`mutation ($g: ID!, $u: ID!) { removeGroupMember(group_id: $g, user_id: $u) { %s } }`, gqlRemoveGroupMemberFields)
	requireMutationBoolOK(t, doGQLOK(t, admin, mRemoveMember, map[string]any{"g": groupID, "u": rawUpper}), "removeGroupMember", "removed")

	mRevokeAll := fmt.Sprintf(`mutation ($u: ID!, $r: SessionEndReason) { revokeAllUserSessions(user_id: $u, reason: $r) { %s } }`, gqlRevokeAllSessionsFields)
	revRes := requireOpMap(t, doGQLOK(t, admin, mRevokeAll, map[string]any{"u": rawLower, "r": "ADMIN_REVOKED"}), "revokeAllUserSessions")
	assertMutationResultSuccess(t, revRes, "revokeAllUserSessions")

	mSuspend := fmt.Sprintf(`mutation ($u: ID!) { suspendUser(user_id: $u) { %s } }`, gqlUserPayloadFields)
	suspUser := requireMutationChildOK(t, doGQLOK(t, admin, mSuspend, map[string]any{"u": rawUpper}), "suspendUser", "user")
	if suspUser["status"] != "SUSPENDED" {
		t.Fatalf("suspendUser(rawUpper) status=%v, want SUSPENDED", suspUser["status"])
	}

	mActivate := fmt.Sprintf(`mutation ($u: ID!) { activateUser(user_id: $u) { %s } }`, gqlUserPayloadFields)
	actUser := requireMutationChildOK(t, doGQLOK(t, admin, mActivate, map[string]any{"u": rawLower}), "activateUser", "user")
	if actUser["status"] != "ACTIVE" {
		t.Fatalf("activateUser(rawLower) status=%v, want ACTIVE", actUser["status"])
	}

	mDisable := fmt.Sprintf(`mutation ($u: ID!) { disableUser(user_id: $u) { %s } }`, gqlUserPayloadFields)
	disUser := requireMutationChildOK(t, doGQLOK(t, admin, mDisable, map[string]any{"u": rawUpper}), "disableUser", "user")
	if disUser["status"] != "DISABLED" {
		t.Fatalf("disableUser(rawUpper) status=%v, want DISABLED", disUser["status"])
	}

	mSoftDel := fmt.Sprintf(`mutation ($u: ID!) { softDeleteUser(user_id: $u) { %s } }`, gqlSoftDeleteUserFields)
	requireMutationBoolOK(t, doGQLOK(t, admin, mSoftDel, map[string]any{"u": rawLower}), "softDeleteUser", "deleted")

	// 5. Kasus negatif: GID dengan resource salah ditolak INVALID_ARGUMENT / ARGUMENT.
	wrongGID := "gid://integrione/role/" + rawLower
	_ = doGQLError(t, admin, `query ($u: ID!) { userIdentities(user_id: $u, first: 1) { total_count } }`, map[string]any{"u": wrongGID}, "INVALID_ARGUMENT")
	_, _ = doMutationFail(t, admin, mActivate, map[string]any{"u": wrongGID}, "activateUser", "ARGUMENT")
}

// TestSelfServiceModuleRBACAndAntiEscalation memverifikasi modul izin *self (readself/updateself/deleteself),
// peran baseline ScopeSelf per jenis akun, isolasi IDOR horizontal, serta pencegahan eskalasi vertikal secara E2E HTTPS.
func TestSelfServiceModuleRBACAndAntiEscalation(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	resetGatewayRateLimits(t, cfg)
	admin := newSuperAdminBrowser(t, cfg)

	// 1. Seluruh 5 kode izin *self harus terdaftar resmi di katalog IAM.
	selfPermCodes := []string{
		"iam.users.readself",
		"iam.users.updateself",
		"iam.users.deleteself",
		"iam.sessions.readself",
		"iam.sessions.deleteself",
	}
	for _, code := range selfPermCodes {
		_ = findPermissionIDByCode(t, admin, code)
	}
	usersReadPermID := findPermissionIDByCode(t, admin, "iam.users.read")

	roleIDByCode := func(code string) string {
		data := doGQLOK(t, admin, fmt.Sprintf(`query ($c: String!) { roleByCode(code: $c) { %s } }`, gqlRoleFields), map[string]any{"c": code})
		r := requireOpMap(t, data, "roleByCode")
		return requireNonEmptyString(t, r, "id")
	}
	customerRoleID := roleIDByCode("customer")
	superadminRoleID := roleIDByCode("superadmin")

	// 2. Uji peran baseline customer & iam_selfservice (ScopeSelf), can_access_identity=false, dan redirect per UserKind.
	baselineCases := []struct {
		kind         string
		wantRole     string
		wantRedirect string
	}{
		{kind: "CUSTOMER", wantRole: "customer", wantRedirect: "/"},
		{kind: "EMPLOYEE", wantRole: "iam_selfservice", wantRedirect: "/account"},
		{kind: "VENDOR", wantRole: "iam_selfservice", wantRedirect: "/vendor"},
		{kind: "SUPPLIER", wantRole: "iam_selfservice", wantRedirect: "/supplier"},
	}

	wsQ := fmt.Sprintf(`query { myWorkspaceAccess { %s } }`, gqlWorkspaceAccessFields)
	snapQ := fmt.Sprintf(`query { myAccessSnapshot { %s } }`, gqlAccessSnapshotFields)

	var custA, custB fixtureUserAccount
	var custClient, custBClient, empClient *browserClient

	for _, bc := range baselineCases {
		resetGatewayRateLimits(t, cfg)
		acct := createFixtureUser(t, admin, cfg, bc.kind, true)
		client, loginRes := loginFixtureUser(t, cfg, acct.Email, acct.Password)
		assertUserLoginPayload(t, loginRes, "AUTHENTICATED")
		if gotLoc := loginRes["redirectLocation"]; gotLoc != bc.wantRedirect {
			t.Fatalf("%s login redirectLocation = %v, want %q", bc.kind, gotLoc, bc.wantRedirect)
		}

		ws := requireOpMap(t, doGQLOK(t, client, wsQ, nil), "myWorkspaceAccess")
		assertWorkspaceAccessPayload(t, ws)
		if requireBoolField(t, ws, "canAccessIdentity") {
			t.Fatalf("%s must have canAccessIdentity=false for baseline *self only", bc.kind)
		}
		if ws["defaultLocation"] != bc.wantRedirect {
			t.Fatalf("%s myWorkspaceAccess defaultLocation = %v, want %q", bc.kind, ws["defaultLocation"], bc.wantRedirect)
		}

		snap := requireOpMap(t, doGQLOK(t, client, snapQ, nil), "myAccessSnapshot")
		assertAccessSnapshotObject(t, snap)
		if requireBoolField(t, snap, "can_access_identity") {
			t.Fatalf("%s snapshot can_access_identity = true, want false", bc.kind)
		}
		if snap["default_redirect_path"] != bc.wantRedirect {
			t.Fatalf("%s snapshot default_redirect_path = %v, want %q", bc.kind, snap["default_redirect_path"], bc.wantRedirect)
		}

		var foundRole bool
		for _, rawR := range requireChildSlice(t, snap, "roles") {
			if r, ok := rawR.(map[string]any); ok && r["code"] == bc.wantRole {
				foundRole = true
				break
			}
		}
		if !foundRole {
			t.Fatalf("%s expected baseline role %q in snapshot, got %+v", bc.kind, bc.wantRole, snap["roles"])
		}

		permScopes := map[string]map[string]any{}
		for _, rawP := range requireChildSlice(t, snap, "permissions") {
			if p, ok := rawP.(map[string]any); ok {
				code, _ := p["permission_code"].(string)
				sc, _ := p["scope"].(map[string]any)
				permScopes[code] = sc
			}
		}
		for _, reqCode := range selfPermCodes {
			sc, ok := permScopes[reqCode]
			if !ok || sc == nil || sc["type"] != "SELF" || sc["id"] != acct.RawID {
				t.Fatalf("%s permission %s expected ScopeSelf(%s), got %+v", bc.kind, reqCode, acct.RawID, sc)
			}
		}

		switch bc.kind {
		case "CUSTOMER":
			custA = acct
			custClient = client
		case "EMPLOYEE":
			empClient = client
		}
	}

	// Buat Customer B untuk skenario isolasi IDOR horizontal antar akun.
	resetGatewayRateLimits(t, cfg)
	custB, custBClient = createFixtureCustomerAndLogin(t, admin, cfg)

	// 3. Alur positif *self pada Customer A (me, mySessions, updateMyProfile, node user/identity/session sendiri, revokeSession sendiri).
	meData := doGQLOK(t, custClient, fmt.Sprintf(`query { me { %s } }`, gqlUserFields), nil)
	if meObj := requireOpMap(t, meData, "me"); meObj["id"] != custA.ID {
		t.Fatalf("me.id = %v, want %q", meObj["id"], custA.ID)
	}

	mySessQ := fmt.Sprintf(`query { mySessions(active_only: true, first: 10) { nodes { %s } total_count } }`, gqlSessionFields)
	custASessNodes := requireConnectionNodes(t, requireOpMap(t, doGQLOK(t, custClient, mySessQ, nil), "mySessions"))
	if len(custASessNodes) == 0 {
		t.Fatalf("expected >=1 active session in custA mySessions")
	}
	custASessGID := requireNonEmptyString(t, custASessNodes[0].(map[string]any), "id")

	updProfQ := fmt.Sprintf(`mutation ($in: UpdateMyProfileInput!) { updateMyProfile(input: $in) { %s } }`, gqlUserPayloadFields)
	updUser := requireMutationChildOK(t, doGQLOK(t, custClient, updProfQ, map[string]any{
		"in": map[string]any{"display_name": "Customer Self Updated"},
	}), "updateMyProfile", "user")
	if updUser["display_name"] != "Customer Self Updated" {
		t.Fatalf("unexpected display_name after updateMyProfile: %+v", updUser)
	}

	selfNodeQ := fmt.Sprintf(`query ($id: ID!) {
		node(id: $id) {
			... on User {
				id
				identities(first: 10) { nodes { %s } total_count }
				roles(first: 10) { total_count }
				effective_roles(first: 10) { total_count }
				groups(first: 10) { total_count }
				sessions(first: 10, active_only: true) { total_count }
			}
		}
	}`, gqlIdentityFields)
	custANode := requireOpMap(t, doGQLOK(t, custClient, selfNodeQ, map[string]any{"id": custA.ID}), "node")
	custAIdents := requireConnectionNodes(t, requireChildMap(t, custANode, "identities"))
	if custANode["id"] != custA.ID || len(custAIdents) == 0 {
		t.Fatalf("unexpected self node response: %+v", custANode)
	}
	custAIdentGID := requireNonEmptyString(t, custAIdents[0].(map[string]any), "id")

	// Uji node langsung pada UserIdentity & Session milik sendiri.
	ownIdentNode := requireOpMap(t, doGQLOK(t, custClient, `query ($id: ID!) { node(id: $id) { ... on UserIdentity { id user_id } } }`, map[string]any{"id": custAIdentGID}), "node")
	if ownIdentNode["id"] != custAIdentGID {
		t.Fatalf("own UserIdentity node mismatch: %+v", ownIdentNode)
	}
	ownSessNode := requireOpMap(t, doGQLOK(t, custClient, `query ($id: ID!) { node(id: $id) { ... on Session { id user_id } } }`, map[string]any{"id": custASessGID}), "node")
	if ownSessNode["id"] != custASessGID {
		t.Fatalf("own Session node mismatch: %+v", ownSessNode)
	}

	// Buat sesi kedua milik custA lalu cabut menggunakan izin iam.sessions.deleteself.
	resetGatewayRateLimits(t, cfg)
	custASecClient, _ := loginFixtureUser(t, cfg, custA.Email, custA.Password)
	secSessNodes := requireConnectionNodes(t, requireOpMap(t, doGQLOK(t, custASecClient, mySessQ, nil), "mySessions"))
	var secondSessGID string
	for _, raw := range secSessNodes {
		if sid := requireNonEmptyString(t, raw.(map[string]any), "id"); sid != custASessGID {
			secondSessGID = sid
			break
		}
	}
	if secondSessGID == "" {
		t.Fatalf("expected secondary session for custA, got %+v", secSessNodes)
	}
	revSessM := fmt.Sprintf(`mutation ($sid: ID!) { revokeSession(session_id: $sid) { %s } }`, gqlRevokeSessionFields)
	requireMutationBoolOK(t, doGQLOK(t, custClient, revSessM, map[string]any{"sid": secondSessGID}), "revokeSession", "revoked")

	// Ambil ID node Session & UserIdentity milik custB untuk verifikasi IDOR horizontal.
	custBSessNodes := requireConnectionNodes(t, requireOpMap(t, doGQLOK(t, custBClient, mySessQ, nil), "mySessions"))
	custBSessGID := requireNonEmptyString(t, custBSessNodes[0].(map[string]any), "id")
	custBNode := requireOpMap(t, doGQLOK(t, custBClient, selfNodeQ, map[string]any{"id": custB.ID}), "node")
	custBIdents := requireConnectionNodes(t, requireChildMap(t, custBNode, "identities"))
	custBIdentGID := requireNonEmptyString(t, custBIdents[0].(map[string]any), "id")

	// 4. Penolakan saat peran *self dicabut dari akun uji.
	resetGatewayRateLimits(t, cfg)
	bareEmp := createFixtureUser(t, admin, cfg, "EMPLOYEE", true)
	bareRolesConn := requireOpMap(t, doGQLOK(t, admin, fmt.Sprintf(`query ($u: ID!) { userRoles(user_id: $u, first: 10) { edges { %s } total_count } }`, gqlUserRoleEdgeFields), map[string]any{"u": bareEmp.ID}), "userRoles")
	for _, raw := range requireConnectionEdges(t, bareRolesConn) {
		edge := raw.(map[string]any)
		roleNode := requireChildMap(t, edge, "node")
		if roleNode["code"] == "iam_selfservice" {
			assignID := requireNonEmptyString(t, edge, "assignment_id")
			requireMutationBoolOK(t, doGQLOK(t, admin, fmt.Sprintf(`mutation ($id: ID!) { revokeUserRole(assignment_id: $id) { %s } }`, gqlRevokeUserRoleFields), map[string]any{"id": assignID}), "revokeUserRole", "revoked")
		}
	}
	bareClient, _ := loginFixtureUser(t, cfg, bareEmp.Email, bareEmp.Password)
	requireCustomerForbidden(t, bareClient, `query { me { id } }`, nil)
	requireCustomerForbidden(t, bareClient, `query { mySessions(first: 5) { total_count } }`, nil)
	requireCustomerForbidden(t, bareClient, `query ($id: ID!) { node(id: $id) { ... on User { id } } }`, map[string]any{"id": bareEmp.ID})
	requireCustomerMutationForbidden(t, bareClient, updProfQ, map[string]any{"in": map[string]any{"display_name": "Blocked"}}, "updateMyProfile")
	requireCustomerMutationForbidden(t, bareClient, fmt.Sprintf(`mutation { deletePendingUserEmail { %s } }`, gqlDeletePendingEmailFields), nil, "deletePendingUserEmail")

	// 5. Penolakan IDOR horizontal oleh custClient & empClient terhadap sumber daya custB.
	for _, client := range []*browserClient{custClient, empClient} {
		requireCustomerForbidden(t, client, `query ($id: ID!) { node(id: $id) { ... on User { id } } }`, map[string]any{"id": custB.ID})
		requireCustomerForbidden(t, client, `query ($id: ID!) { node(id: $id) { ... on Session { id } } }`, map[string]any{"id": custBSessGID})
		requireCustomerForbidden(t, client, `query ($id: ID!) { node(id: $id) { ... on UserIdentity { id } } }`, map[string]any{"id": custBIdentGID})
		requireCustomerForbidden(t, client, `query ($id: ID!) { user(id: $id) { id } }`, map[string]any{"id": custB.ID})
		requireCustomerForbidden(t, client, `query ($u: ID!) { userIdentities(user_id: $u, first: 5) { total_count } }`, map[string]any{"u": custB.ID})
		requireCustomerForbidden(t, client, `query ($u: ID!) { userSessions(user_id: $u, first: 5) { total_count } }`, map[string]any{"u": custB.ID})
		requireCustomerMutationForbidden(t, client, revSessM, map[string]any{"sid": custBSessGID}, "revokeSession")
	}
	// Sesi milik Customer B wajib tetap hidup setelah percobaan IDOR revokeSession dari pihak lain.
	if bMe := requireOpMap(t, doGQLOK(t, custBClient, `query { me { id } }`, nil), "me"); bMe["id"] != custB.ID {
		t.Fatalf("custB session unexpectedly disrupted after denied IDOR revokeSession: %+v", bMe)
	}

	// 6. Penolakan eskalasi vertikal & mismatch scope SELF.
	createEmpM := fmt.Sprintf(`mutation ($in: CreateEmployeeUserInput!) { createEmployeeUser(input: $in) { %s } }`, gqlUserPayloadFields)
	assignUserRoleM := fmt.Sprintf(`mutation ($in: AssignUserRoleInput!) { assignUserRole(input: $in) { %s } }`, gqlUserRolePayloadFields)
	assignRolePermM := fmt.Sprintf(`mutation ($r: ID!, $p: ID!) { assignRolePermission(role_id: $r, permission_id: $p) { %s } }`, gqlRolePermPayloadFields)

	for _, client := range []*browserClient{custClient, empClient} {
		requireCustomerForbidden(t, client, `query { users(first: 5) { total_count } }`, nil)
		requireCustomerMutationForbidden(t, client, createEmpM, map[string]any{
			"in": map[string]any{"worker_id": "55555555-5555-4555-8555-555555555555", "provisioning_request_id": "66666666-6666-4666-8666-666666666666", "full_name": "Escalate", "email": uniqueTag("esc") + "@integrione.test", "username": uniqueTag("esc_u"), "password": "StrongPass#2026!"},
		}, "createEmployeeUser")
		requireCustomerMutationForbidden(t, client, assignUserRoleM, map[string]any{
			"in": map[string]any{"user_id": custA.ID, "role_id": superadminRoleID, "scope_type": "GLOBAL"},
		}, "assignUserRole")
		requireCustomerMutationForbidden(t, client, assignRolePermM, map[string]any{
			"r": customerRoleID, "p": usersReadPermID,
		}, "assignRolePermission")
	}

	// Admin mencoba memasang scope_type SELF dengan scope_id user lain wajib ditolak ARGUMENT.
	_, _ = doMutationFail(t, admin, assignUserRoleM, map[string]any{
		"in": map[string]any{
			"user_id":    custA.ID,
			"role_id":    customerRoleID,
			"scope_type": "SELF",
			"scope_id":   custB.RawID,
		},
	}, "assignUserRole", "ARGUMENT")
}

// TestCreateUserVariantsAutoAssignSelfRoleAndPermissions memastikan registrasi createUser (publik & admin),
// createEmployeeUser, createVendorUser, dan createSupplierUser langsung mewarisi peran + 5 permission *self sejak status PENDING.
func TestCreateUserVariantsAutoAssignSelfRoleAndPermissions(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	resetGatewayRateLimits(t, cfg)
	admin := newSuperAdminBrowser(t, cfg)

	wantSelfPerms := map[string]bool{
		"iam.users.readself":      true,
		"iam.users.updateself":    true,
		"iam.users.deleteself":    true,
		"iam.sessions.readself":   true,
		"iam.sessions.deleteself": true,
	}

	// 1. Pastikan peran bawaan customer & iam_selfservice tepat memegang seluruh 5 permission *self tanpa izin lintas-user.
	rolePermsQ := fmt.Sprintf(`query ($r: ID!) { rolePermissions(role_id: $r, first: 25) { nodes { %s } total_count } }`, gqlPermissionFields)
	for _, roleCode := range []string{"customer", "iam_selfservice"} {
		rData := doGQLOK(t, admin, fmt.Sprintf(`query ($c: String!) { roleByCode(code: $c) { %s } }`, gqlRoleFields), map[string]any{"c": roleCode})
		rObj := requireOpMap(t, rData, "roleByCode")
		roleID := requireNonEmptyString(t, rObj, "id")
		rpConn := requireOpMap(t, doGQLOK(t, admin, rolePermsQ, map[string]any{"r": roleID}), "rolePermissions")
		rpNodes := requireConnectionNodes(t, rpConn)
		gotCodes := make(map[string]bool, len(rpNodes))
		for _, rawP := range rpNodes {
			pObj := rawP.(map[string]any)
			code := requireNonEmptyString(t, pObj, "code")
			gotCodes[code] = true
			if !wantSelfPerms[code] {
				t.Fatalf("role %q contains unexpected non-*self permission %q", roleCode, code)
			}
		}
		for reqCode := range wantSelfPerms {
			if !gotCodes[reqCode] {
				t.Fatalf("role %q missing required *self permission %q, got %+v", roleCode, reqCode, gotCodes)
			}
		}
	}

	assertUserHasBaselineSelfRole := func(t *testing.T, label, userGID, rawUID, wantRoleCode string) {
		t.Helper()
		urConn := requireOpMap(t, doGQLOK(t, admin, fmt.Sprintf(`query ($u: ID!) {
			userRoles(user_id: $u, first: 10) { edges { %s } total_count }
		}`, gqlUserRoleEdgeFields), map[string]any{"u": userGID}), "userRoles")
		edges := requireConnectionEdges(t, urConn)
		if len(edges) != 1 {
			t.Fatalf("%s userRoles count = %d, want 1 (baseline %q)", label, len(edges), wantRoleCode)
		}
		edge := edges[0].(map[string]any)
		scope := requireChildMap(t, edge, "scope")
		if scope["type"] != "SELF" || scope["id"] != rawUID {
			t.Fatalf("%s userRoles scope = %+v, want SELF(%s)", label, scope, rawUID)
		}
		rNode := requireChildMap(t, edge, "node")
		if rNode["code"] != wantRoleCode {
			t.Fatalf("%s userRoles role code = %v, want %q", label, rNode["code"], wantRoleCode)
		}

		effConn := requireOpMap(t, doGQLOK(t, admin, fmt.Sprintf(`query ($u: ID!) {
			userEffectiveRoles(user_id: $u, first: 10) { nodes { %s } total_count }
		}`, gqlRoleFields), map[string]any{"u": userGID}), "userEffectiveRoles")
		effNodes := requireConnectionNodes(t, effConn)
		if len(effNodes) != 1 || effNodes[0].(map[string]any)["code"] != wantRoleCode {
			t.Fatalf("%s userEffectiveRoles = %+v, want single %q", label, effNodes, wantRoleCode)
		}
	}

	verifyActiveLoginAndSelfPerms := func(t *testing.T, label, userGID, rawUID, email, password, wantRole, wantRedirect string) {
		t.Helper()
		actData := doGQLOK(t, admin, fmt.Sprintf(`mutation ($u: ID!) { activateUser(user_id: $u) { %s } }`, gqlUserPayloadFields), map[string]any{"u": userGID})
		_ = requireMutationChildOK(t, actData, "activateUser", "user")
		psqlExec(t, cfg, fmt.Sprintf(`UPDATE user_identities SET is_verified = true, verified_at = now() WHERE user_id = %s::uuid;`, sqlQuote(rawUID)))

		resetGatewayRateLimits(t, cfg)
		client, loginRes := loginFixtureUser(t, cfg, email, password)
		assertUserLoginPayload(t, loginRes, "AUTHENTICATED")
		if loginRes["redirectLocation"] != wantRedirect {
			t.Fatalf("%s login redirectLocation = %v, want %q", label, loginRes["redirectLocation"], wantRedirect)
		}

		snap := requireOpMap(t, doGQLOK(t, client, fmt.Sprintf(`query { myAccessSnapshot { %s } }`, gqlAccessSnapshotFields), nil), "myAccessSnapshot")
		assertAccessSnapshotObject(t, snap)
		if requireBoolField(t, snap, "can_access_identity") {
			t.Fatalf("%s snapshot can_access_identity = true, want false", label)
		}
		roles := requireChildSlice(t, snap, "roles")
		if len(roles) != 1 || roles[0].(map[string]any)["code"] != wantRole {
			t.Fatalf("%s snapshot roles = %+v, want [%q]", label, roles, wantRole)
		}
		gotPerms := map[string]bool{}
		for _, rawP := range requireChildSlice(t, snap, "permissions") {
			p := rawP.(map[string]any)
			code := requireNonEmptyString(t, p, "permission_code")
			sc := requireChildMap(t, p, "scope")
			if sc["type"] != "SELF" || sc["id"] != rawUID {
				t.Fatalf("%s permission %q unexpected scope = %+v, want SELF(%s)", label, code, sc, rawUID)
			}
			gotPerms[code] = true
		}
		for reqCode := range wantSelfPerms {
			if !gotPerms[reqCode] {
				t.Fatalf("%s snapshot missing auto-assigned permission %q, got %+v", label, reqCode, gotPerms)
			}
		}

		// Pastikan operasi mandiri (me & mySessions) langsung bisa diakses dan akses lintas-user ditolak FORBIDDEN.
		if meObj := requireOpMap(t, doGQLOK(t, client, `query { me { id _id } }`, nil), "me"); meObj["id"] != userGID {
			t.Fatalf("%s me.id = %v, want %q", label, meObj["id"], userGID)
		}
		sessConn := requireOpMap(t, doGQLOK(t, client, `query { mySessions(active_only: true, first: 5) { total_count } }`, nil), "mySessions")
		if requireNumberField(t, sessConn, "total_count") < 1 {
			t.Fatalf("%s expected >= 1 session in mySessions", label)
		}
		requireCustomerForbidden(t, client, `query { users(first: 5) { total_count } }`, nil)
	}

	// 2. Uji registrasi anonim publik createUser (CUSTOMER) - PENDING langsung memiliki peran customer + 5 izin *self.
	resetGatewayRateLimits(t, cfg)
	pubClient := newAnonBrowser(t, cfg)
	pubTag := uniqueTag("auto_pub_cust")
	pubEmail := pubTag + "@integrione.test"
	pubPass := "V7$K9q#R2w!X8mZp"
	pubData := doGQLOK(t, pubClient, fmt.Sprintf(`mutation ($in: CreateUserInput!) { createUser(input: $in) { %s } }`, gqlCreateUserPayloadFields), map[string]any{
		"in": map[string]any{
			"full_name": "Pub " + pubTag,
			"email":     pubEmail,
			"username":  pubTag,
			"password":  pubPass,
		},
	})
	pubPayload := requireOpMap(t, pubData, "createUser")
	if !requireBoolField(t, pubPayload, "was_successful") || pubPayload["user"] != nil {
		t.Fatalf("expected anonymous createUser masking user=nil, got %+v", pubPayload)
	}
	pubUsersConn := requireOpMap(t, doGQLOK(t, admin, fmt.Sprintf(`query ($q: String) { users(query: $q, first: 5) { nodes { %s } total_count } }`, gqlUserFields), map[string]any{"q": pubEmail}), "users")
	pubUsers := requireConnectionNodes(t, pubUsersConn)
	if len(pubUsers) != 1 {
		t.Fatalf("expected 1 user matched by %s, got %d", pubEmail, len(pubUsers))
	}
	pubU := pubUsers[0].(map[string]any)
	pubGID := requireNonEmptyString(t, pubU, "id")
	pubRawID := requireNonEmptyString(t, pubU, "_id")
	t.Cleanup(func() { cleanupUserByID(t, cfg, pubRawID) })
	if pubU["status"] != "PENDING" || pubU["kind"] != "CUSTOMER" {
		t.Fatalf("unexpected public created user: %+v", pubU)
	}
	assertUserHasBaselineSelfRole(t, "anonymous createUser", pubGID, pubRawID, "customer")
	verifyActiveLoginAndSelfPerms(t, "anonymous createUser", pubGID, pubRawID, pubEmail, pubPass, "customer", "/")

	// 3. Uji seluruh mutasi admin (createUser, createEmployeeUser, createVendorUser, createSupplierUser) beserta nested roles di payload PENDING.
	adminMutations := []struct {
		opName       string
		inputType    string
		wantKind     string
		wantRole     string
		wantRedirect string
	}{
		{opName: "createUser", inputType: "CreateUserInput", wantKind: "CUSTOMER", wantRole: "customer", wantRedirect: "/"},
		{opName: "createEmployeeUser", inputType: "CreateEmployeeUserInput", wantKind: "EMPLOYEE", wantRole: "iam_selfservice", wantRedirect: "/account"},
		{opName: "createVendorUser", inputType: "CreateVendorUserInput", wantKind: "VENDOR", wantRole: "iam_selfservice", wantRedirect: "/vendor"},
		{opName: "createSupplierUser", inputType: "CreateSupplierUserInput", wantKind: "SUPPLIER", wantRole: "iam_selfservice", wantRedirect: "/supplier"},
	}

	for _, tc := range adminMutations {
		resetGatewayRateLimits(t, cfg)
		tag := uniqueTag("auto_" + strings.ToLower(tc.wantKind))
		email := tag + "@integrione.test"
		pass := "V7$K9q#R2w!X8mZp"

		mQ := fmt.Sprintf(`mutation ($in: %s!) {
			%s(input: $in) {
				user {
					%s
					roles(first: 10) {
						edges { %s }
						total_count
					}
					effective_roles(first: 10) {
						nodes { %s }
						total_count
					}
				}
				errors { type message }
				was_successful
			}
		}`, tc.inputType, tc.opName, gqlUserFields, gqlUserRoleEdgeFields, gqlRoleFields)

		inVars := map[string]any{
			"full_name": "Auto " + tag,
			"email":     email,
			"username":  tag,
			"password":  pass,
		}
		if tc.wantKind == "VENDOR" || tc.wantKind == "SUPPLIER" {
			inVars["party_id"] = createFixturePartyForRole(t, admin, cfg, tc.wantKind)
		}

		mData := doGQLOK(t, admin, mQ, map[string]any{
			"in": inVars,
		})
		uObj := requireMutationChildOK(t, mData, tc.opName, "user")
		assertUserObject(t, uObj)
		if uObj["kind"] != tc.wantKind || uObj["status"] != "PENDING" {
			t.Fatalf("%s returned user kind=%v status=%v, want kind=%s status=PENDING", tc.opName, uObj["kind"], uObj["status"], tc.wantKind)
		}
		uGID := requireNonEmptyString(t, uObj, "id")
		uRawID := requireNonEmptyString(t, uObj, "_id")
		t.Cleanup(func() { cleanupUserByID(t, cfg, uRawID) })

		// Verifikasi roles & effective_roles langsung tersedia di payload mutasi saat user baru masih berstatus PENDING.
		nestedRoles := requireChildMap(t, uObj, "roles")
		nestedEdges := requireConnectionEdges(t, nestedRoles)
		if len(nestedEdges) != 1 {
			t.Fatalf("%s nested PENDING user.roles len = %d, want 1", tc.opName, len(nestedEdges))
		}
		edge := nestedEdges[0].(map[string]any)
		sc := requireChildMap(t, edge, "scope")
		if sc["type"] != "SELF" || sc["id"] != uRawID || requireChildMap(t, edge, "node")["code"] != tc.wantRole {
			t.Fatalf("%s nested PENDING role mismatch: %+v", tc.opName, edge)
		}
		nestedEff := requireConnectionNodes(t, requireChildMap(t, uObj, "effective_roles"))
		if len(nestedEff) != 1 || nestedEff[0].(map[string]any)["code"] != tc.wantRole {
			t.Fatalf("%s nested PENDING effective_roles = %+v, want [%q]", tc.opName, nestedEff, tc.wantRole)
		}

		assertUserHasBaselineSelfRole(t, tc.opName, uGID, uRawID, tc.wantRole)
		verifyActiveLoginAndSelfPerms(t, tc.opName, uGID, uRawID, email, pass, tc.wantRole, tc.wantRedirect)
	}
}
