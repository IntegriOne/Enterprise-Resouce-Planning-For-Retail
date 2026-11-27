# Audit Wiring Gateway–IAM & Go Error Handling (`golang-error-handling`)

**Tanggal:** 2026-11-25 05:38:18 WIB  
**Ruang Lingkup:** Wiring Gateway–IAM dan ketahanan outage IAM (`backend/gateway/cmd`, `internal/config`, `internal/server`, `internal/source_connections`, `internal/middleware`, `internal/adapter/iam`, `internal/browserauth`, `graph/resolvers`, `Dockerfile`, `docker-compose.yml`, `proxy/Caddyfile`)  
**Standar Acuan:** Skill `/golang-error-handling` (5 kategori audit), `AGENTS.md`, `agents/boundary.md`, dan `agents/arsitektur.md`

---

## 1. Ringkasan Eksekutif & Status Ketahanan Outage IAM

Ketika layanan `iam` dimatikan saat startup maupun di tengah runtime, **Gateway tetap hidup (`healthy`), tidak panic/crash, tetap melayani `query { health }` dan domain independen seperti `Enterprise Foundation` selama snapshot Redis masih segar, serta melakukan *auto-recovery* tanpa restart kontainer ketika `iam` pulih**.

Dari audit menyeluruh lintas 5 kategori `/golang-error-handling`, ditemukan **10 temuan** teknis (2 *linter* `errcheck`, 4 celah ketahanan *panic*/nil-*safety*/S2S *classification*/half-open *probe*, 2 celah *error wrapping* & *single-handling*, serta 2 celah *structured logging* & *cleanup* resource startup). Seluruh temuan yang relevan telah diperbaiki langsung pada `backend/gateway` dan diverifikasi ulang menggunakan `go fmt`, `go vet`, `golangci-lint` (`0 issues`), unit test `-race`, `govulncheck` (`0` kerentanan aktif), pembangunan ulang kontainer `gateway-1` & `gateway-2`, serta pengujian E2E outage langsung (`TestGateway_*` + `TestLiveHTTPSIAMGatewayE2E`).

---

## 2. Hasil Audit per Kategori (`/golang-error-handling`)

### Kategori 1: Error Creation, Sentinels, Custom Types & Klasifikasi Outage IAM
- **Temuan K1.1 [HIGH] — Status `codes.Unauthenticated` S2S dari interceptor IAM menghapus cookie browser pengguna**  
  - *Lokasi:* `backend/gateway/internal/source_connections/iam_client.go` (`MapIAMGRPCError`) & `backend/gateway/internal/adapter/iam/auth_adapter.go` (`BrowserSession`).  
  - *Akar Masalah:* Interceptor IAM (`backend/services/iam/transport/grpc/interceptor/interceptor.go`) membalas kegagalan verifikasi HMAC/konteks internal dengan `codes.Unauthenticated` tanpa detail `ErrorInfo` (`"internal caller verification is unavailable"`, `"invalid internal caller context"`, `"missing internal caller context"`). `MapIAMGRPCError` sebelumnya memetakan seluruh `codes.Unauthenticated` ke `CodeUnauthenticated`, sehingga `BrowserSession` menghapus cookie sesi sah pengguna (`ClearSessionCookie`).  
  - *Perbaikan:* Menambahkan `isInternalCallerUnauthenticatedMessage` pada `iam_client.go` agar kegagalan otentikasi konteks internal S2S dipetakan menjadi `middleware.ErrUnavailable("identity service unavailable", err)` (sejalan dengan `ef_client.go`) tanpa menghapus sesi browser pengguna.
- **Temuan K1.2 [MEDIUM] — `codes.Aborted` jatuh ke `CodeInternalError` & kapitalisasi sentinel startup**  
  - *Lokasi:* `backend/gateway/internal/source_connections/iam_client.go` (`MapIAMGRPCError`, `initIAMConnection`).  
  - *Perbaikan:* Memasukkan `codes.Aborted` bersama `codes.AlreadyExists` dan `codes.FailedPrecondition` ke `CodeConflict`, serta menormalkan string error target kosong ke huruf kecil (`"identity grpc target is empty"`).
- **Status Lulus:** Sentinel `ErrIAMCircuitOpen`, `ErrIAMUnavailable`, `GatewayError`, serta pemetaan `UNAVAILABLE` pada query, mutasi (`MutationResult`), dan `VerifySensitiveSession` (tanpa dimask menjadi `FORBIDDEN`) berjalan konsisten dan deterministik.

### Kategori 2: Error Wrapping (`%w` vs `%v`), Chain Inspection & `errors.Join`
- **Temuan K2.1 [MEDIUM] — Double-wrapping `GatewayError` dan kehilangan cause pada `normalizeAuthError`**  
  - *Lokasi:* `backend/gateway/internal/adapter/iam/adapter.go` (`normalizeAuthError`).  
  - *Akar Masalah:* Error dari `callIAMAuth` sudah terbungkus `*middleware.GatewayError`; pemetaan ulang `iam.user.system_cannot_login` membungkus `*GatewayError` di dalam `*GatewayError` kedua, sementara jalur fallback `ErrUnauthenticated` melempar `nil` sebagai cause.  
  - *Perbaikan:* Menambahkan helper `unwrapGatewayCause` sehingga pembungkusan tidak berlapis ganda dan `cause` internal tetap terjaga bagi diagnostik server tanpa mengubah pesan anti-enumerasi ke klien.
- **Temuan K2.2 [LOW] — Pemakaian `errors.As` lama pada codebase Go 1.27.4 & sanitasi `isLocalCallerCancellation`**  
  - *Lokasi:* `backend/gateway/internal/middleware/errors.go` (`AsGatewayError`), `backend/gateway/internal/server/error_presenter.go` (`buildErrorPresenter`), `backend/gateway/internal/source_connections/resilience.go` (`isUpstreamUnavailableFailure`).  
  - *Perbaikan:* Migrasi ke `errors.AsType[*GatewayError](err)` dan `errors.AsType[*gqlerror.Error](err)` (Go 1.26+), serta memastikan `isUpstreamUnavailableFailure` memeriksa gRPC `codes.Canceled` terbungkus via `status.FromError(err)` saat `parentCtx.Err() != nil`.
- **Status Lulus:** `Registry.Close()` memakai `errors.Join(errs...)`, dan pembatasan batas publik GraphQL menggunakan `GatewayError` terklasifikasi sehingga rincian internal gRPC/net tidak bocor ke klien.

### Kategori 3: Single-Handling Rule, Swallowed Errors & Ketahanan Runtime Outage IAM
- **Temuan K3.1 [HIGH] — Pembatalan klien saat *half-open probe* menutup *circuit breaker* secara prematur**  
  - *Lokasi:* `backend/gateway/internal/source_connections/resilience.go` (`afterCall`), `iam_client.go` (`InvokeIAM`), `ef_client.go` (`InvokeEF`).  
  - *Akar Masalah:* Ketika breaker berstatus `circuitHalfOpen`, pembatalan lokal oleh klien HTTP (`context.Canceled`) menghasilkan `upstreamDown == false`, sehingga `afterCall(true, false)` langsung menutup sirkuit (`circuitClosed`) meski IAM belum terbukti sehat.  
  - *Perbaikan:* Memperluas `afterCall(isProbe, upstreamDown, callerCanceled)` dengan `isLocalCallerCancellation(ctx, rpcErr)` sehingga pembatalan lokal saat probe tetap mempertahankan breaker di status `circuitOpen` dan mereset flag `halfOpenInFlight`.
- **Temuan K3.2 [MEDIUM] — Kegagalan `errcheck` dan error terbuang `_ =`**  
  - *Lokasi:* `backend/gateway/cmd/healthcheck/main.go` (`resp.Body.Close`), `backend/gateway/internal/server/server_test.go` (`readBody`), `backend/gateway/internal/middleware/session.go` (`PromoteAuthenticatedCookie`).  
  - *Perbaikan:* Menangani hasil error `resp.Body.Close()` secara eksplisit pada `cmd/healthcheck/main.go` dan `server_test.go`, serta mengganti `_ = m.store.DeletePreAuth(...)` menjadi pengecekan error dengan `slog.Warn` (kecuali `ErrPreAuthNotFound`).
- **Temuan K3.3 [MEDIUM] — `VerifySensitiveSession` menyamarkan `CodeInternalError` menjadi `FORBIDDEN`**  
  - *Lokasi:* `backend/gateway/internal/middleware/authorizer.go` (`enforceSensitiveReauth`).  
  - *Perbaikan:* Mengekstrak helper bersama `verifySensitiveWithChecker`, meloloskan `CodeUnavailable` **dan** `CodeInternalError` tanpa masking `FORBIDDEN`, serta mempertahankan cause pada `ErrForbidden`.
- **Temuan K3.4 [MEDIUM] — `os.Exit` di `main()` melewati `defer Close()`**  
  - *Lokasi:* `backend/gateway/cmd/main.go`.  
  - *Perbaikan:* Mengekstrak fungsi `run(logger *slog.Logger) error` sehingga `defer registry.Close()` dan `defer redisClient.Close()` selalu dieksekusi sebelum proses berhenti.
- **Status Ketahanan Startup & Infrastruktur:** `backend/docker-compose.yml` tidak menyematkan `depends_on` antara `gateway-1`/`gateway-2` terhadap `iam` maupun `redis`; `cmd/healthcheck` dan `proxy/Caddyfile` memeriksa `{health}` GraphQL tanpa memanggil IAM; `grpc.NewClient` non-blocking dengan `WaitForReady(false)` memastikan Gateway tetap siap saat IAM mati.

### Kategori 4: Panic/Recover, Concurrency & Nil-Safety
- **Temuan K4.1 [HIGH] — Tidak adanya middleware panic recovery pada lapis HTTP Fiber di luar `gqlSrv.ServeHTTP`**  
  - *Lokasi:* `backend/gateway/internal/server/server.go` (`New`).  
  - *Perbaikan:* Menambahkan `requestPanicRecoverMiddleware(deps.Logger)` sebelum middleware logging Fiber agar setiap panic tak terduga di router/adaptor HTTP terpulihkan bersih dengan HTTP 500 JSON `INTERNAL_ERROR` dan log `slog.Error` ber-`stack` tanpa menjatuhkan proses Gateway.
- **Temuan K4.2 [HIGH] — Nil-dereference guard pada `ErrorPresenter`, `SnapshotLoader` (`singleflight`), dan method publik `iam.Adapter`**  
  - *Lokasi:* `backend/gateway/internal/server/error_presenter.go` (saat `err == nil`), `backend/gateway/internal/middleware/snapshot.go` (`sfGroup.Do`), dan `backend/gateway/internal/adapter/iam/{adapter,auth_adapter,self_adapter,user_admin_adapter,rbac_adapter}.go`.  
  - *Perbaikan:* Menambahkan guard `err == nil` sebelum `err.Error()`, `recover()` di dalam `sfGroup.Do` menjadi `ErrUnavailable`, serta `ensureAuthDeps` / `ensureAuthorizerDeps` / `ensureAuthFlowDeps` pada method publik `Adapter` termasuk cek `sess.Token == nil` di `Node`.

### Kategori 5: Structured Logging (`slog`), Severity Level & Keamanan PII/Internal
- **Temuan K5.1 [MEDIUM] — Ketidaktepatan level log `CodeInternalError`, hilangnya `request_id` di `ErrorPresenter`/`guard`, serta daftar blokir pesan upstream**  
  - *Lokasi:* `backend/gateway/internal/server/error_presenter.go`, `backend/gateway/internal/server/guard.go` (`singleErrorResponse`), `backend/gateway/internal/source_connections/resilience.go` (`commonBlockedErrorFragments`).  
  - *Perbaikan:* Memisahkan log `CodeInternalError` ke `logger.Error` dan `CodeUnavailable` ke `logger.Warn` dengan atribut `request_id`; meneruskan `ctx` (bukan `context.Background()`) pada `singleErrorResponse` di `guard.go`; serta memperluas `commonBlockedErrorFragments` dengan `"lookup "`, `"grpc:"`, `"tls:"`, `"x509:"`, `"broken pipe"`, `"unexpected eof"`, dan `"context deadline exceeded"`.
- **Status Lulus:** Tidak ada pemakaian `fmt.Print*` atau `log.Print*` di kode produksi Gateway, serta pesan client bebas dari rincian host/IP/SQL/token internal.

---

## 3. Verifikasi Pasca-Perbaikan

Seluruh pemeriksaan dijalankan setelah edit terakhir dan lulus penuh:
1. `cd backend/gateway && go fmt ./... && go vet ./... && golangci-lint run ./...`: **Lulus (0 issues)**.
2. `cd backend/gateway && go test -race -count=1 ./...`: **Lulus** seluruh `8` paket ber-test (`internal/adapter/enterprise_foundation`, `internal/adapter/iam`, `internal/browserauth`, `internal/config`, `internal/gqltools/test`, `internal/middleware`, `internal/server`, `internal/source_connections`).
3. `docker compose build gateway-1 && docker compose up -d --wait gateway-1 gateway-2`: **Lulus** (`integrione-retail-gateway-1-1` & `integrione-retail-gateway-2-1` berstatus `healthy`).
4. `cd backend/gateway && go test -v -tags=e2e -count=1 ./internal/e2e -run 'TestGateway_|TestLiveHTTPSIAMGatewayE2E'`: **Lulus seluruhnya (`71.882s`)**:
   - `TestGateway_BootsWithoutIAMAndRedisThenAutoRecovers` (`22.67s`)
   - `TestGateway_ReplicaPeerAndCaddyFailoverIndependence` (`19.65s`)
   - `TestGateway_RuntimeIAMOutageFaultIsolationAndAutoRecovery` (`12.22s`)
   - `TestLiveHTTPSIAMGatewayE2E` (`7` subtest, `17.35s`)
5. `govulncheck ./...` di `backend/gateway`: **0 kerentanan aktif** (*0 vulnerabilities in packages you import*).
