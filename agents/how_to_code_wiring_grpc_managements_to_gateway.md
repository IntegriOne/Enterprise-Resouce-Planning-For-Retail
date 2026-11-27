# Panduan Standar: Wiring gRPC Management ke GraphQL Gateway (`backend/gateway`)

Dokumen ini adalah panduan operasional langkah demi langkah (*step-by-step*) lengkap dengan pola kode kanonik untuk menyambungkan microservice gRPC management yang sudah berjalan ke GraphQL Gateway tunggal (`/api/graphql`). Referensi implementasi aktif yang menjadi acuan baku adalah wiring [Enterprise Foundation](../backend/gateway/internal/adapter/enterprise_foundation/) dan lapisan transport [source_connections](../backend/gateway/internal/source_connections/).

Prinsip inti: **Gateway adalah perimeter tunggal *fail-closed* dan *fault-isolated***. Mati atau error-nya satu microservice bisnis **tidak boleh** mematikan `gateway-1`, `gateway-2`, `iam`, maupun layanan lain, dan **tidak boleh** mencabut sesi login browser pengguna.

---

## 0. Aturan Zona, Fondasi, & Bahasa (WAJIB)

### A. Aturan Kerja & Bahasa
1. **Bahasa:** Kode Go (package, file, struct, method, var, konstanta, log teknis, error code, error message) wajib **Bahasa Inggris**. Komentar kode (`// ...`, kecuali direktif `//go:generate`, `//go:embed`, `//go:build`) dan dokumentasi di `agents/` wajib **Bahasa Indonesia**.
2. **Cek Status Aktif vs `PAUSED` ([agents/boundary.md](boundary.md)):** Dilarang meng-wire management berstatus `PAUSED` atau membuat skema `.graphql` placeholder kosong karena [gqlgen.yml](../backend/gateway/gqlgen.yml) memuat otomatis pola `graphql/**/*.graphql`.
3. **Riset, Git `main`-only, & Zona Aman:** Selalu cek dengan `codebase-memory-mcp` + `go doc`, cek `git status --short backend/gateway/` sebelum menyentuh file bersama, dilarang membuat Git branch baru, dan stage spesifik (`git add <path>`).

### B. Partisi Kepemilikan 3 Zona di `backend/gateway/`
1. **Zona 1 — Milik Eksklusif Management (bebas dikoding penuh):**
   - `graphql/{management}/{management}_*.graphql`
   - `graph/generated/{management}_*.generated.go` *(hasil generator, dilarang edit manual)*
   - `graph/model/{management}_*_models_gen.go` *(hasil splitter generator, dilarang edit manual)*
   - `graph/resolvers/{management}_*.resolvers.go` *(delegator tipis 4–5 baris ke adapter)*
   - `internal/adapter/{management}/...`
   - `internal/source_connections/{management}_client.go` & `*_test.go`
   - `internal/config/{management}_config_test.go`
   - `internal/e2e/{management}_e2e_test.go`
2. **Zona 2 — File Bersama (*Additive-Only* & *Minimal-Diff*):**
   - `go.mod`, [Dockerfile](../backend/gateway/Dockerfile), `.env`, `.env.example`, [backend/docker-compose.yml](../backend/docker-compose.yml)
   - [gqlgen.yml](../backend/gateway/gqlgen.yml) (hanya tambah entri `models:` nested field domain sendiri)
   - [graph/resolvers/resolver.go](../backend/gateway/graph/resolvers/resolver.go) (field adapter + getter *fail-closed* + `NewComplexityRoot()`)
   - [graph/resolvers/schema.resolvers.go](../backend/gateway/graph/resolvers/schema.resolvers.go) (hanya tambah cabang dispatch GID di `Query.Node`)
   - [internal/source_connections/registry.go](../backend/gateway/internal/source_connections/registry.go) (field `conn` & `*ClientSet`, init non-fatal di `NewRegistry`, tutup di `Close()`)
   - [internal/config/config.go](../backend/gateway/internal/config/config.go) & [cmd/main.go](../backend/gateway/cmd/main.go)
   - File bersama hasil generator (`root_.generated.go`, `schema.generated.go`, `schema_models_gen.go`) — cukup cek diff agar hanya menambah `Query`/`Mutation` domain sendiri.
3. **Zona 3 — Core Keamanan Gateway (DILARANG DIUBAH untuk wiring domain):**
   - `internal/server/`, `internal/middleware/`, `internal/browserauth/`, [graphql/schema.graphql](../backend/gateway/graphql/schema.graphql).
   - **Dilarang** menambah route REST di luar `/api/graphql`, melonggarkan CSRF/Origin/cookie, menaruh logika bisnis di Gateway, atau query langsung PostgreSQL/Redis milik service domain.

---

## 1. Roles, Permissions, Scope, & Peta Error Handling

### A. Penentuan & Penulisan Dokumen Roles–Permissions
1. **Siapa yang menentukan?** **Agen developer pemilik management** merancang daftar permission `<management>.<resource>.<action>`, penanda sensitivitas (`is_sensitive`), dan `scope_type` (`global`, `legal_entity`, `facility`, `organization_unit`, `vendor_party`, `supplier_party`, `self`, dll.) dengan mengacu pada [agents/users.md](users.md).
2. **Kapan ditulis?** Tulis/mutakhirkan **`agents/roles-permissions/{management}/1-roles-permissions.md`** beserta manifest kode **`backend/shared/rbac/catalogs/{management}.json` di Tahap 1 wiring (SEBELUM menulis `.graphql` atau adapter)**. Contoh baku: [enterprise_foundation/1-roles-permissions.md](roles-permissions/enterprise_foundation/1-roles-permissions.md).
3. **Kapan diterapkan di kode & di mana dicek?**
   - Diterapkan langsung pada lapisan **Gateway Adapter (`internal/adapter/{management}/permissions.go` + `{subdomain}_adapter.go`)** menggunakan `Authorizer.RequirePermission` / `RequireSelfPermission` **sebelum** panggilan gRPC ke service domain dieksekusi, disertai unit test konsistensi 1:1 antara konstanta permission adapter dan manifest domain-nya.
   - **Panduan `requireSensitive` (`true` vs `false`):** set `false` untuk query baca dan mutasi create/update rutin; set **`true`** untuk operasi berisiko tinggi seperti `delete*`, perubahan status kritikal/dissolve, atau pemberian akses portal/keuangan sehingga mewajibkan re-autentikasi < 10 menit + MFA jika aktif.
   - **Sinkronisasi ke Database `iam`:** Mekanisme manifest loader di IAM sudah aktif (`backend/shared/rbac` + `rbac_catalog_sync.go`): setiap rebuild & startup container `iam` otomatis menyinkronkan manifest `backend/shared/rbac/catalogs/{management}.json` ke DB `iam` secara idempoten dengan advisory lock PostgreSQL, tanpa perlu mengedit manual `bootstrap.go` atau menaruh seed SQL terpisah. *(Di tes E2E, `seedEFPermissionCatalog` hanya berfungsi sebagai pembaca manifest `shared/rbac` dan fallback safety lokal bila dijalankan di DB kosong.)*

### B. Alur Lengkap Error Handling & Isolasi Gangguan

#### 1. Tabel Klasifikasi Error Gateway ([errors.go](../backend/gateway/internal/middleware/errors.go))

| Sumber / Status gRPC | `GatewayError.Code` | `ErrorTypeEnum` Frontend | Aturan & Keamanan Wajib |
| --- | --- | --- | --- |
| Tidak ada sesi / sesi expired | `CodeUnauthenticated` | `AUTHENTICATION` | Hanya berasal dari `AuthManager` (IAM/Redis), **bukan** dari domain downstream |
| Tidak punya permission / scope / reauth kadaluarsa | `CodeForbidden` | `AUTHORIZATION` | Pesan aman, mis. `"missing required permission: ..."` |
| IAM gagal saat `VerifySensitiveSession` / `EnsureFresh` | **`CodeUnavailable`** | `UNAVAILABLE` | `Authorizer` mem-*bubble-up* `UNAVAILABLE` tanpa memasking ke `FORBIDDEN`; cookie sesi dipertahankan |
| `codes.InvalidArgument`, invalid GID, invalid cursor | `CodeInvalidArgument` | `ARGUMENT` | Pesan disanitasi via `sanitizeUpstreamMessage` |
| `codes.NotFound` | `CodeNotFound` | `NOT_FOUND` | - |
| `codes.AlreadyExists`, `FailedPrecondition`, `Aborted` | `CodeConflict` | `CONFLICT` | Contoh: kode duplikat atau objek masih dirujuk |
| `codes.PermissionDenied` | `CodeForbidden` | `AUTHORIZATION` | - |
| `codes.ResourceExhausted` | `CodeTooManyRequests` | `THROTTLE` | - |
| `codes.Unavailable`, `DeadlineExceeded`, `Canceled`, circuit open, init degraded, atau panic | **`CodeUnavailable`** | `UNAVAILABLE` | Detail `dial tcp`, `connection refused`, host/port, `sqlstate`, `pq:` wajib diblokir |
| **`codes.Unauthenticated` dari microservice domain** | **`CodeUnavailable`** | `UNAVAILABLE` | **WAJIB!** Kegagalan bukti HMAC internal antarlayanan tidak boleh melogout user dari sesi IAM |
| Nil response / error tak terklasifikasi | `CodeInternalError` | `INTERNAL_ERROR` | Pesan ke klien wajib `"internal server error"` |

#### 2. Perbedaan Output GraphQL: Query vs Mutation (`HTTP 200 OK` `/api/graphql`)
- **Operasi Query:** kembalikan `(nil, err)`. `buildErrorPresenter` ([error_presenter.go](../backend/gateway/internal/server/error_presenter.go)) membentuk JSON `errors: [{ message, extensions: { code, type } }]` dan mencatat log `Warn` untuk `INTERNAL_ERROR`/`UNAVAILABLE`.
- **Operasi Mutation:** bungkus **seluruh** error adapter (`requirePerm`, `ParseTypedID`, validasi input, error gRPC, adapter `nil`) ke `ToMutationErrors(err)` lalu kembalikan `&model.{Op}Payload{ Errors: ..., WasSuccessful: false }, nil`.
  *(Catatan: penolakan pra-resolver pada `AroundOperations` di [guard.go](../backend/gateway/internal/server/guard.go) — seperti `X-CSRF-Token` hilang, Origin asing, complexity berlebih, depth > 10, alias > 25, rate-limit Redis, atau `EnsureFresh` snapshot > 30s gagal saat Redis/IAM mati — keluar sebagai top-level GraphQL `errors` dengan `code` terstandar `FORBIDDEN`/`INVALID_ARGUMENT`/`UNAVAILABLE`.)*

---

## 2. Urutan Pengerjaan Wiring: Tahap 0 s.d. Tahap 7

Kerjakan seluruh tahap dalam **satu jendela atomik** sampai `backend/gateway` kompilasi bersih dan teruji:

### Tahap 0 — Verifikasi Prasyarat & Lingkungan
1. Pastikan service domain (`backend/services/{management}/`) lulus `go test -race` dan container-nya berstatus `healthy`.
2. Cek `git status --short backend/gateway/` untuk memastikan file Zona 2 bersih dari perubahan agen lain yang belum selesai.

### Tahap 1 — Dokumen & Manifest Roles–Permissions (`1-roles-permissions.md` + `shared/rbac/catalogs/`)
1. Daftar semua Query & Mutation, kode izin `<management>.<resource>.<action>`, penanda `is_sensitive`, serta daftar `scope_type` target di `agents/roles-permissions/{management}/1-roles-permissions.md` dan manifest `backend/shared/rbac/catalogs/{management}.json`. Pastikan seluruh konstanta permission di Gateway Adapter cocok 1:1 dengan manifest domain tersebut.

### Tahap 2 — Module Go, Dockerfile, Config, & Compose Non-Blocking (`Zona 2`)
> **Gotcha Build Docker yang Sering Terlewat:** saat menambah dependency module baru ke `backend/gateway/go.mod`, **`backend/gateway/Dockerfile` juga wajib diperbarui** karena Dockerfile menyalin masing-masing folder `services/*` secara eksplisit!

1. **[backend/go.work](../backend/go.work):** pastikan `./services/{management}` sudah ada.
2. **[backend/gateway/go.mod](../backend/gateway/go.mod):** tambahkan `require` & `replace github.com/integrione/erp-retail/backend/services/{management} => ../services/{management}`.
3. **[backend/gateway/Dockerfile](../backend/gateway/Dockerfile):**
   - Pada stage `deps`: tambahkan `COPY services/{management}/go.mod services/{management}/go.sum /src/services/{management}/` sebelum `RUN go mod download`.
   - Pada stage `builder`: tambahkan `COPY services/{management} /src/services/{management}`.
4. **[internal/config/config.go](../backend/gateway/internal/config/config.go):** tambahkan `{MGMT}GRPCEnabled` (bool), `{MGMT}GRPCTarget` (mis. `127.0.0.1:9003`), `{MGMT}GRPCTimeout` (default `5s`).
5. **[backend/docker-compose.yml](../backend/docker-compose.yml) & `.env` / `.env.example`:**
   - Tambahkan `{MGMT}_GRPC_ENABLED`, `{MGMT}_GRPC_TARGET` (`{service_name}:9003`), `{MGMT}_GRPC_TIMEOUT_SECONDS` ke **kedua replika** `gateway-1` **dan** `gateway-2`.
   - Pastikan `GATEWAY_INTERNAL_HMAC_SECRET` **identik byte-per-byte** dengan `{MGMT}_INTERNAL_GATEWAY_HMAC_SECRET` (>= 32 karakter).
   - **ATURAN DECOUPLING KERAS:** **DILARANG** menambah `{service_name}` ke `depends_on` milik `gateway-1`, `gateway-2`, `iam`, atau `caddy`!
6. **Unit Test Config (`internal/config/{mgmt}_config_test.go`):** tiru [ef_config_test.go](../backend/gateway/internal/config/ef_config_test.go), verifikasi default dan pastikan saat `{MGMT}_GRPC_ENABLED=false` di `APP_ENV=production`, `ValidateRuntimeStrict()` tetap lulus.

### Tahap 3 — Transport Client Resilien & Circuit Breaker (`internal/source_connections/`)
Ikuti pola [resilience.go](../backend/gateway/internal/source_connections/resilience.go), [ef_client.go](../backend/gateway/internal/source_connections/ef_client.go), dan [registry.go](../backend/gateway/internal/source_connections/registry.go):

1. Buat `internal/source_connections/{management}_client.go`:
   - Gunakan `newResilientGRPCConn(target, secretCopy)` (`grpc.NewClient` non-blocking, `WaitForReady(false)`, backoff `250ms`–`2s`, keepalive client `Time: 60s` agar aman di atas `EnforcementPolicy.MinTime` server `20s`–`30s`, interceptor HMAC `internalContextClientInterceptor`).
   - Buat `{Domain}ClientSet` berisi seluruh stub gRPC + `enabled`, `timeout`, `initErr`, `breaker *circuitBreaker` (`threshold = 5`, `openDuration = 3s`).
   - Sediakan `init{Domain}Connection(cfg, secretCopy)` yang **non-fatal** saat target kosong/degraded.
   - Pola fungsi generik `Invoke{Domain}` & pemetaan error wajib:

```go
// InvokeDomain menjalankan satu panggilan unary domain secara terisolasi lengkap dengan timeout, circuit breaker, pemulihan panic, dan pemeriksaan nil.
func InvokeDomain[Resp any](
	ctx context.Context,
	clients *DomainClientSet,
	call func(context.Context, *DomainClientSet) (*Resp, error),
) (resp *Resp, err error) {
	if clients == nil || !clients.Enabled() {
		return nil, middleware.ErrUnavailable("domain service unavailable", ErrDomainDisabled)
	}
	if !clients.Ready() {
		return nil, middleware.ErrUnavailable("domain service unavailable", clients.InitError())
	}
	if call == nil {
		return nil, middleware.NewGatewayError(middleware.CodeInternalError, "internal server error", errors.New("rpc callback is nil"))
	}
	isProbe, cbErr := clients.breaker.beforeCall()
	if cbErr != nil {
		return nil, middleware.ErrUnavailable("domain service unavailable", cbErr)
	}
	rpcCtx, cancel := withBoundedTimeout(ctx, clients.timeout)
	defer cancel()
	defer func() {
		if recovered := recover(); recovered != nil {
			clients.breaker.afterCall(isProbe, true)
			resp = nil
			err = middleware.ErrUnavailable("domain service unavailable", fmt.Errorf("recovered panic: %v", recovered))
		}
	}()
	res, rpcErr := call(rpcCtx, clients)
	clients.breaker.afterCall(isProbe, isUpstreamUnavailableFailure(ctx, rpcErr))
	if rpcErr != nil {
		return nil, MapDomainGRPCError(rpcErr)
	}
	if res == nil {
		return nil, middleware.NewGatewayError(middleware.CodeInternalError, "internal server error", errors.New("domain returned nil response"))
	}
	return res, nil
}
```
   - Di `MapDomainGRPCError`: petakan `codes.Unauthenticated` -> `middleware.ErrUnavailable("... service unavailable", err)`, gunakan `sanitizeUpstreamMessage(st.Message(), fallback, "{service_name}:", ":900X")`, dan ekstrak `extractErrorReason(st)` bila butuh pemetaan reason spesifik.
2. Registrasikan di [registry.go](../backend/gateway/internal/source_connections/registry.go) (field `conn`, `*ClientSet`, `InitError`, serta tutup di `Close()`).
3. Tambahkan unit test `internal/source_connections/{mgmt}_client_test.go` mengikuti [ef_client_test.go](../backend/gateway/internal/source_connections/ef_client_test.go):
   - Startup `NewRegistry` saat disabled/target kosong/server mati tetap sukses & tidak mengganggu IAM.
   - Circuit breaker membuka setelah 5 kegagalan `Unavailable`/`DeadlineExceeded`, fail-fast saat open, probe half-open, dan menutup saat pulih.
   - Recovery panic & respons nil aman.
   - `codes.Unauthenticated` dipetakan ke `UNAVAILABLE`, pesan `dial tcp` / IP internal ter-sanitasi.

### Tahap 4 — Skema GraphQL Modular Hybrid A1 & Code Generation (`graphql/{management}/`)
Ikuti pola [ef_common.graphql](../backend/gateway/graphql/enterprise_foundation/ef_common.graphql) dan [ef_address.graphql](../backend/gateway/graphql/enterprise_foundation/ef_address.graphql):

1. Buat `backend/gateway/graphql/{management}/{management}_common.graphql` dan `{management}_{subdomain}.graphql` (prefix `{management}_` wajib unik karena `gqlgen` memetakan output berdasarkan nama file):

```graphql
type SampleEntity implements Node {
  id: ID!                # Relay Global ID: gid://integrione/sample-entity/<uuid>
  _id: ID!               # UUID mentah (di struct model Go menjadi field RawID string)
  code: String!
  created_at: Time!
  updated_at: Time!
  parent: SampleEntity   # Daftarkan resolver: true di gqlgen.yml bila relasi nested
}

type SampleEntityEdge {
  node: SampleEntity!
  cursor: String!
}

type SampleEntityConnection {
  edges: [SampleEntityEdge!]!
  nodes: [SampleEntity!]!
  pageInfo: PageInfo!
  total_count: Int!
}

type CreateSampleEntityPayload implements MutationResult {
  sample_entity: SampleEntity
  errors: [MutationError!]!
  was_successful: Boolean!
}

extend type Query {
  sampleEntity(id: ID!): SampleEntity
  sampleEntities(after: String, before: String, first: Int, last: Int, query: String): SampleEntityConnection!
}

extend type Mutation {
  createSampleEntity(input: CreateSampleEntityInput!): CreateSampleEntityPayload!
}
```
2. Jika entitas punya field relasi nested (`parent`, `legal_entity`, `children`, dll.), daftarkan `resolver: true` pada `models:` di [gqlgen.yml](../backend/gateway/gqlgen.yml).
3. Jalankan generator resmi:
   ```bash
   cd backend/gateway && go generate ./...
   ```
   Pastikan `graph/generated/{management}_*.generated.go`, `graph/model/{management}_*_models_gen.go`, dan `graph/resolvers/{management}_*.resolvers.go` terpecah otomatis per berkas skema.

### Tahap 5 — Lapisan Adapter Domain (`internal/adapter/{management}/`)
Pecah teratur per file mengikuti [internal/adapter/enterprise_foundation/](../backend/gateway/internal/adapter/enterprise_foundation/):

1. **`adapter.go`:** `NewAdapter(clients, authorizer, rpcTimeout)` (tolak hanya jika `authorizer == nil`; boleh `clients` degraded), `Node(ctx, rawGID)`, dan helper `call{Domain}` yang membungkus `authcontext.WithCaller(callCtx, sess.Caller)` lalu memanggil `source_connections.Invoke{Domain}`.
2. **`permissions.go`:** konstanta permission & scope + `requirePerm` + `requireAnyScope`:
   - Di `requireAnyScope`, jika `Authorizer.RequirePermission` mengembalikan error selain `CodeForbidden`, atau `CodeForbidden` yang **tidak** berawalan `"missing required permission:"` (mis. gagal step-up reauth sensitif), langsung kembalikan error tersebut — jangan mencoba scope alternatif!
3. **`gid.go`:** `EncodeGID`, `ParseTypedID`, `ParseOptionalTypedID`, `ParseTypedIDList`, `DecodeNodeGID`, dan `Is{Domain}NodeID` (`gid://integrione/<resource>/<uuid>`).
4. **`relay.go`:** `ParseRelayPage`, `EncodeCursor`/`DecodeCursor` (versi `"v1:"` base64 URL), dan `buildRelayConnection` (default limit `20`, max `100`, support `first/after` & `last/before`, clamp integer).
5. **`mutation_result.go`:** `ToMutationErrors(err)` & `NoMutationErrors()`.
6. **`mapper_common.go` & `mapper_*.go`:** konversi Protobuf <-> `model.*` (perhatikan `RawID: pb.GetId()`, `ID: EncodeGID(..., pb.GetId())`, waktu ke `wibZoneOnce`) serta helper resolver nested (`if obj == nil { return nil, nil }`, ambil ID dari `obj.RawID` / FK ter-encode).
7. **`{subdomain}_adapter.go` — 3 Pola Implementasi Kritis:**
   - **Pola 1 — Query by-ID berscope + Anti-ID-Oracle ([legal_entity_adapter.go](../backend/gateway/internal/adapter/enterprise_foundation/legal_entity_adapter.go#L41-L60)):**
     Jika `ParseTypedID` gagal, jalankan `a.requirePerm(ctx, PermRead, false)` terlebih dahulu sebelum mengembalikan `err`, sehingga pemanggil ilegal tidak mendapat bocoran validasi input. Setelah UUID valid, panggil `a.requireAnyScope(ctx, PermRead, false, targetScope(cleanID))`.
   - **Pola 2 — Batch Query `ByIds` berscope:** cek `ScopeGlobal` terlebih dahulu; jika tidak punya `ScopeGlobal`, validasi `requireAnyScope` untuk **setiap** item di `cleanIDs` sebelum memanggil gRPC batch agar tidak membocorkan entitas milik scope lain!
   - **Pola 3 — Mutation Payload Konsisten:** selalu gunakan helper `fail`:
     ```go
     fail := func(err error) *model.CreateSampleEntityPayload {
         return &model.CreateSampleEntityPayload{Errors: ToMutationErrors(err), WasSuccessful: false}
     }
     ```
     Tangkap **semua** error (auth, `ParseTypedID`, validasi input, gRPC, unavailable) lewat `fail(err)`.
8. **Unit Tests Adapter (`adapter_test.go` & `fault_isolation_rbac_test.go`):** uji RBAC *fail-closed* (`UNAUTHENTICATED`, `FORBIDDEN`), kecocokan scope spesifik vs global, *sensitive reauth* (`UNAVAILABLE` dari IAM tidak dimasking jadi `FORBIDDEN`), validasi GID, Relay pagination, serta simulasi gRPC mati/nil.

### Tahap 6 — Wiring Tipis Resolvers, ComplexityRoot, `Query.Node`, & `cmd/main.go` (`Zona 2`)
1. **`graph/resolvers/{management}_{subdomain}.resolvers.go` (delegator 4–5 baris):**
   - **Query:** `adp, err := r.{domain}(); if err != nil { return nil, err }; return adp.{Op}(ctx, ...)`
   - **Mutation:** `adp, err := r.{domain}(); if err != nil { return &model.{Op}Payload{Errors: adapter.ToMutationErrors(err), WasSuccessful: false}, nil }; return adp.{Op}(ctx, input), nil`
   - **Nested field:** cek `if obj == nil { return nil, nil }` bila perlu, lalu delegasikan ke `adp.Resolve...` atau `adp.{Children}(ctx, obj.RawID, ...)`.
2. **[graph/resolvers/resolver.go](../backend/gateway/graph/resolvers/resolver.go):**
   - Tambah field `{Domain}Adapter` dan getter `func (r *Resolver) {domain}() (*adapter.Adapter, error)` yang mengembalikan `middleware.ErrUnavailable` bila nil.
   - Daftarkan **seluruh** Query & nested Connection berpaginasi di `NewComplexityRoot()` memakai `paginated(childComplexity, first, last)`.
3. **[graph/resolvers/schema.resolvers.go](../backend/gateway/graph/resolvers/schema.resolvers.go):** tambah cabang `if {domain}adapter.Is{Domain}NodeID(id)` pada `Query.Node`.
4. **[cmd/main.go](../backend/gateway/cmd/main.go):** buat adapter setelah `iam` & `authorizer`; bila `registry.{Domain}InitError != nil` atau `NewAdapter` error, hanya log `logger.Warn(...)` agar Gateway dan IAM tetap berjalan.

### Tahap 7 — Live HTTPS E2E (`internal/e2e/{management}_e2e_test.go`)
Gunakan build tag `//go:build e2e` dan pakai ulang helper E2E yang sudah tersedia di [internal/e2e/](../backend/gateway/internal/e2e/) (jangan membuat ulang client HTTPS/cookie/CSRF!):
- Helper siap pakai: `loadE2ERuntimeConfig(t)`, `newSuperAdminBrowser(t, cfg)`, `newAnonBrowser(t, cfg)`, `createFixtureCustomerAndLogin(t, admin, cfg)`, `doGQLOK`, `doGQLError`, `doGQLErrorLabeled`, `doMutationFailLabeled`, `requireOpMap`, `requireMutationChildOK`, `requireNonEmptyString`, `requireErrorCode`, `composeCmd`, `assertContainerHealthy`, `waitForGraphQLRecovery`, `uniqueTag`, `psqlExec`.
- **3 Test Wajib (lihat pola [enterprise_foundation_e2e_test.go](../backend/gateway/internal/e2e/enterprise_foundation_e2e_test.go)):**
  1. `Test{Domain}_FullLifecycleAndNestedGraphQL`: fail-closed anonim (`UNAUTHENTICATED`) & customer (`FORBIDDEN`), full CRUD + lookup + transisi status, Relay cursor forward, nested resolvers, serta resolusi polimorfik `Query.node`.
  2. `Test{Domain}_ScopedRBACEnforcement`: user dengan scope entitas A sukses mengakses A tetapi ditolak `FORBIDDEN` saat mengakses entitas B.
  3. `Test{Domain}_OutageFaultIsolationAndAutoRecovery`:
     - Daftarkan `t.Cleanup` **sebelum** menghentikan container agar container `{service_name}` dijamin menyala kembali meski tes gagal.
     - Jalankan `composeCmd(t, cfg, "stop", "{service_name}")` -> verifikasi `gateway-1`, `gateway-2`, `iam`, `redis` tetap `healthy`, `health == "ok"`, `me`/`users` sukses, cookie sesi aman, query & mutasi domain gagal tertutup `UNAVAILABLE` tanpa kebocoran `host:port`/`dial tcp`/`postgres`, dan customer tetap `FORBIDDEN`.
     - Nyalakan kembali `{service_name}` -> verifikasi 4 kali request berturut-turut melintasi dua replika Gateway + 1 mutasi pulih otomatis **tanpa restart Gateway/IAM**.
  - Pastikan seluruh entitas uji di DB domain dan user/role uji TestAB di DB `iam` dibersihkan rapi via `t.Cleanup`. Pastikan pula katalog permission & template role resmi modul sudah tersinkron otomatis di DB `iam` dari manifest `backend/shared/rbac/catalogs/{management}.json`.

---

## 3. Anti-Pola yang Dilarang

- Mengedit file Zona 3 (`internal/server/`, `internal/middleware/`, `browserauth/`, `graphql/schema.graphql`).
- Menambah modul ke `backend/gateway/go.mod` tetapi lupa menyalinnya ke stage `deps` & `builder` di [backend/gateway/Dockerfile](../backend/gateway/Dockerfile).
- Menambah env hanya di `gateway-1` tanpa memperbarui `gateway-2` di [backend/docker-compose.yml](../backend/docker-compose.yml), atau menambahkan `{service_name}` ke `depends_on` milik `gateway-1`, `gateway-2`, atau `caddy`.
- Mengedit manual `*.generated.go` atau `*_models_gen.go` (selalu pakai `cd backend/gateway && go generate ./...`).
- Lupa mendaftarkan Connection query/nested baru ke `NewComplexityRoot()`.
- Memetakan `codes.Unauthenticated` dari downstream bisnis ke `CodeUnauthenticated` (berisiko melogout user IAM).
- Mengembalikan top-level GraphQL error dari resolver Mutation alih-alih `MutationResult { errors, was_successful: false }`.
- Melewatkan cek `requirePerm`/`requireAnyScope` atau menaruh query SQL/logika bisnis di Gateway.

---

## 4. Checklist Verifikasi Wajib Sebelum Handover

> **Catatan `GOWORK=off`:** Saat menjalankan toolchain dari dalam direktori `backend/gateway/`, gunakan `GOWORK=off` agar `go vet` dan `go test` hanya mengompilasi modul Gateway & dependency-nya secara cepat dan terisolasi.

```bash
# 1. Regenerasi GraphQL & cek whitespace/diff
cd backend/gateway
go generate ./...
git diff --check

# 2. Format, vet (dengan tag e2e), dan seluruh unit test + race detector
gofmt -s -w .
GOWORK=off go vet -tags=e2e ./...
GOWORK=off go test -race -count=1 ./...

# 3. Build ulang Gateway & recreates replika + Caddy
cd ../..
docker compose build gateway-1
docker compose up -d --wait --no-build --force-recreate gateway-1 gateway-2 caddy
docker compose exec -T caddy caddy validate --config /etc/caddy/Caddyfile

# 4. Jalankan suite E2E HTTPS penuh (wajib 0 fail, 0 skip)
cd backend/gateway
GOWORK=off go test -tags=e2e -v -count=1 -timeout 12m ./internal/e2e/...
```

Setelah seluruh unit & E2E test lulus (`0` fail, `0` skip) dan database kembali bersih:
1. Perbarui `agents/context/gateway/{id}-context.md`.
2. Simpan laporan E2E lengkap di `agents/e2e/gateway/{id}-{judul}-query-dan-mutation.md` dan *working history* di `agents/working_history/gateway/{id}-YYYY-MM-DD-HHmmss-{judul}.md` (maksimal 150 kata).
3. Jalankan re-index penuh `codebase-memory-mcp`.
