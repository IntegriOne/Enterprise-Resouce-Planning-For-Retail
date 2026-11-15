# Laporan Audit & Remediasi Keamanan IAM, Gateway, Proxy, dan Edge Middleware (`/golang-security`)

Tanggal: 2026-11-15 08:38 WIB (Pembaruan Verifikasi Toolchain & Runtime: 2026-11-15 08:51 WIB)  
Auditor: Agen Audit Keamanan ERP Retail  
Cakupan: `backend/shared/{authcontext,authsession,authz}`, `backend/services/iam`, `backend/gateway`, `proxy/Caddyfile`, dan `frontend/src/middleware.ts`

---

## 1. Ringkasan Eksekutif

Audit keamanan menyeluruh dilakukan terhadap rantai autentikasi, otorisasi RBAC/ABAC, proteksi CSRF/cookie, sinkronisasi sesi Redis/PostgreSQL, lapisan transport gRPC & GraphQL Gateway, proxy TLS Caddy, serta middleware Edge Next.js. Fokus utama mencakup penutupan celah *user/state enumeration* (*auth clues*), penghapusan jalur *fail-open*, pencegahan eskalasi hak istimewa dan IDOR, batas kedalaman/alias/paginasi GraphQL, penanganan mutex/panik pada reconciler/mailer, serta penambahan header keamanan HTTP.

Seluruh temuan aplikasi berisiko tinggi dan menengah telah diremediasi, dilengkapi test regresi `-race`, dibangun ulang ke stack Docker Compose, dan diverifikasi ulang melalui pengujian unit, build frontend, `curl` HTTPS langsung, serta suite E2E HTTPS lintas stack.

---

## 2. Metodologi Penilaian (DREAD)

Setiap temuan dinilai menggunakan 5 dimensi DREAD berskala `0–10` (**D**amage, **R**eproducibility, **E**xploitability, **A**ffected users, **D**iscoverability) dengan skor rata-rata `Total = (D+R+E+A+D)/5`:
- **Critical (`>= 8.0`)**
- **High (`6.5 – 7.9`)**
- **Medium (`4.0 – 6.4`)**
- **Low / Advisory (`< 4.0`)**

---

## 3. Matriks Temuan, Skor DREAD & Status Remediasi

### F-01 — Kebocoran Clue Autentikasi, Status Akun, Lockout & Replay MFA (*Auth/Oracle Leaks*)
- **Lokasi Sebelum Remediasi:** `services/iam/services/user/{service.go,auth.go,mfa.go,registration_runtime.go}`, `gateway/internal/adapter/iam/auth_adapter.go`
- **Skor DREAD:** D=8, R=9, E=8, A=9, D=7 → **8.2 (Critical)**
- **Masalah:** Kegagalan login pada akun tidak ditemukan/terkunci/dihapus memiliki jalur waktu berbeda dibanding verifikasi Argon2id, dan error spesifik (eksistensi akun, lockout, status *suspended/disabled/deleted*, maupun *TOTP replay* `CONFLICT`) berpotensi menjadi oracle bagi penyerang. Konfirmasi reset sandi & verifikasi email juga perlu memasking akun non-aktif/dihapus.
- **Remediasi:**
  - `AuthenticatePassword` mengeksekusi verifikasi Argon2id *dummy* waktu konstan bagi akun tidak ditemukan/terkunci/terhapus serta mengembalikan error ambigu (`ErrInvalidCredentials`).
  - Gateway `UserLogin` & `ValidateUserMfa` menggunakan `normalizeAuthError` agar `NOT_FOUND`, `PERMISSION_DENIED`, `RESOURCE_EXHAUSTED`, dan `ALREADY_EXISTS` (*replay* TOTP) dipetakan secara seragam ke `AUTHENTICATION` dengan pesan netral, disertai *rate limit* `auth_subject`, `preauth_ip`, serta batas 5 percobaan per challenge MFA (`mfa_challenge` → `THROTTLE`).
  - `ConfirmPasswordReset` memetakan akun non-aktif/terhapus ke `ErrTokenNotFound`, dan `ConfirmEmailVerificationByToken` memetakan akun terhapus ke `VerifyEmailStatusInvalid`.
- **Status:** **Tutup & Terverifikasi** (`TestUserService_UnknownLoginAndDisabledAccountResetMasking`, `TestUserService_ConfirmEmailVerificationByTokenOnDeletedUser`, `TestUserLoginAndValidateMFAAmbiguousErrorsAndThrottle`).

### F-02 — *Fail-Open* Keyring CSRF Produksi & Validasi Proof gRPC IAM
- **Lokasi Sebelum Remediasi:** `services/iam/services/registry.go`, `shared/authcontext/context.go`, `services/iam/transport/grpc/interceptor/interceptor.go`, `gateway/internal/middleware/session.go`
- **Skor DREAD:** D=9, R=6, E=5, A=10, D=5 → **7.0 (High)**
- **Masalah:**
  1. `registry.go` pernah mengizinkan `AllowEphemeralCSRFKey` aktif bila `CSRFKeyring` kosong di mode non-dev.
  2. Validasi proof HMAC gRPC pada `authcontext` & interceptor IAM berpotensi mengalami desinkronisasi `firstClientIP` serta perlu penjagaan *nil info*.
  3. Penanganan kesalahan CSRF vs kehilangan/kedaluwarsa sesi pada `RequireMutationCSRF` perlu dibedakan agar *mismatch* CSRF tidak langsung menghapus cookie sesi sah.
  4. `ChangeUserPassword` harus membersihkan `X-CSRF-Token`, sementara Edge `frontend/src/middleware.ts` harus memeriksa keberadaan nilai cookie non-kosong (`!sessionCookie || !sessionCookie.value`).
- **Remediasi:** Penghapusan bypass `!sec.CSRFKeyringConfigured` dan `len(opts.CSRFKeyring) == 0` di `registry.go` (*fail-closed* di `"production"` tanpa `IAM_CSRF_HMAC_KEYRING`); konsistensi `firstClientIP` HMAC & *nil guard* interceptor; `RequireMutationCSRF` hanya menghapus cookie jika handle hilang/kedaluwarsa; `ChangeUserPassword` membersihkan header CSRF; `frontend/src/middleware.ts` memeriksa `.value` cookie.
- **Status:** **Tutup & Terverifikasi** (`TestWireIAMServices_WithGormRegistry`, `shared/authcontext`, `curl -skSI -H 'Cookie: foo=bar' https://erp.localhost:8443/identity` → `HTTP/2 307` `location: /login`).

### F-03 — Eskalasi Hak Istimewa RBAC, Mutasi Role Sistem (`IsSystem`) & IDOR Admin
- **Lokasi Sebelum Remediasi:** `services/iam/services/role/service.go`, `gateway/internal/adapter/iam/{rbac_adapter.go,user_admin_adapter.go}`
- **Skor DREAD:** D=9, R=8, E=7, A=8, D=6 → **7.6 (High)**
- **Masalah:** Perlu proteksi berlapis agar API publik tidak dapat membuat role sistem (`is_system: true`), role sistem `IsSystem` bersifat *immutable* terhadap `UpdateDetails`/`AssignPermission`/`RevokePermission`, operator non-superadmin tidak bisa menyentuh akun/role superadmin maupun menghapus/menonaktifkan dirinya sendiri, serta mencegah IDOR pencabutan *assignment ID* langsung (`RevokeUserRole`, `RevokeGroupRole`).
- **Remediasi:**
  - Domain & service `role` menegakkan `ErrSystemRoleImmutable` pada `UpdateDetails`, `AssignPermission`, `RevokePermission`, `Deactivate`, dan `Delete` sambil mempertahankan `BootstrapAssignPermission` khusus inisialisasi awal.
  - Gateway `CreateRole` menolak `is_system: true` (`AUTHORIZATION` / `FORBIDDEN`), `ensureCanTouchTargetUser` menolak `SoftDeleteUser`/`Suspend`/`Disable` terhadap akun sendiri maupun target superadmin oleh non-superadmin, dan `RevokeUserRole`/`RevokeGroupRole` mewajibkan `callerIsSuperAdmin`.
- **Status:** **Tutup & Terverifikasi** (`TestRoleService_SystemRoleImmutableMutationsAndBootstrap`, `TestRBACSystemRoleCreationForbiddenAndIDORAndSelfDeleteGuards`).

### F-04 — DoS GraphQL (Kedalaman/Alias/Cycle), Bounds Paginasi Relay & Kebocoran Detail Error Parser JSON
- **Lokasi Sebelum Remediasi:** `gateway/internal/server/{guard.go,error_presenter.go,server.go}`, `gateway/internal/adapter/iam/relay.go`
- **Skor DREAD:** D=7, R=9, E=9, A=9, D=8 → **8.4 (Critical)**
- **Masalah:**
  1. Ketiadaan batas kedalaman AST, alias masif, dan siklus fragment GraphQL berisiko DoS CPU/memori.
  2. `relay.go` rawan panik slice saat paginasi mundur bila `upperBound < offset` serta butuh batas keras `maxPaginationOffset = 1_000_000_000`.
  3. Saat pengujian regresi dijalankan, ditemukan bahwa `transport.POST` gqlgen memancarkan prefix `"json request body could not be decoded: ... body:..."` sehingga pola `"json body could not be decoded"` di `error_presenter.go` terlewat dan membocorkan cuplikan payload + detail parser Go `encoding/json`.
- **Remediasi:** Penerapan `maxGraphQLDepth = 10`, `maxGraphQLAliases = 25`, deteksi *fragment cycle*, default `mutation_ip` `60/1m`; penambahan pengecekan `remaining <= 0` dan `maxPaginationOffset = 1_000_000_000` di `relay.go`; pembaruan `buildErrorPresenter` di `error_presenter.go` untuk memasking `"json request body could not be decoded"` menjadi `"invalid JSON request body"`.
- **Status:** **Tutup & Terverifikasi** (`TestRelayBoundsDecodeCursorAndBackwardUpperBoundNoPanic`, `TestGraphQLDepthAliasLimitsJSONMaskingAndSecurityHeaders`).

### F-05 — Cacat Konkurensi Reconciler, Shutdown, Keepalive gRPC & Cookie Deduplication
- **Lokasi Sebelum Remediasi:** `services/iam/infrastructure/redisstore/reconciler.go`, `services/iam/cmd/main.go`, `services/iam/infrastructure/mailer/dispatcher.go`, `services/iam/transport/grpc/server.go`, `gateway/internal/browserauth/effects.go`
- **Skor DREAD:** D=7, R=5, E=4, A=8, D=4 → **5.6 (Medium)**
- **Masalah:** Unlock manual di dalam fungsi bermutex di `reconciler.go`, tiadanya batas batch sesi pemulihan, pembukaan gerbang sebelum pemeriksaan outbox gagal (`CountFailed`), ketiadaan `recover()` pada worker reconciler/mailer, pengabaian WaitGroup saat shutdown `main.go`, ketiadaan gRPC `KeepaliveEnforcementPolicy` eksplisit, serta duplikasi header `Set-Cookie` untuk nama cookie sama pada `ResponseEffects`.
- **Remediasi:** Refaktor ke `bootstrapGateAndDrainLocked`, pembatasan `recoverySessionBatchCap = 2000`, pemeriksaan `CountFailed`, `recover()` + timeout `30s` di `stepOnce` dan `processJob`, `sync.WaitGroup` di `cmd/main.go`, `KeepaliveEnforcementPolicy` (`MinTime: 20s`, `PermitWithoutStream: true`) di `transport/grpc/server.go`, serta deduplikasi nama cookie pada `ResponseEffects.AddCookie`.
- **Status:** **Tutup & Terverifikasi** (`infrastructure/redisstore`, `infrastructure/mailer`, `cmd`, `browserauth` `-race` PASS).

### F-06 — Header Keamanan HTTP & Validasi IP Proxy Gateway/Caddy
- **Lokasi Sebelum Remediasi:** `gateway/internal/server/server.go`, `proxy/Caddyfile`
- **Skor DREAD:** D=5, R=9, E=7, A=9, D=8 → **7.6 (High)**
- **Masalah:** Penguatan *clickjacking*, MIME sniffing, CSP API, `Referrer-Policy`, `Permissions-Policy`, HSTS, stripping header `Server`, serta sanitasi `clientIP` dan `UserAgent`.
- **Remediasi:** Penambahan header `Strict-Transport-Security`, `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`, `Referrer-Policy: same-origin`, `Permissions-Policy: camera=(), microphone=(), geolocation=()`, `-Server` pada `proxy/Caddyfile`; serta `Cache-Control: no-store...`, `X-Frame-Options: DENY`, `Content-Security-Policy: default-src 'none'; frame-ancestors 'none'`, `Cross-Origin-Resource-Policy: same-origin`, `EnableIPValidation: true`, `normalizeClientIP`, dan `sanitizeUserAgent` di `gateway/internal/server/server.go`.
- **Status:** **Tutup & Terverifikasi** (`TestGraphQLDepthAliasLimitsJSONMaskingAndSecurityHeaders` & `curl -skSI https://erp.localhost:8443/login`).

---

## 4. Temuan Toolchain (`govulncheck` pada `go1.26.8` → Ditutup di `go1.27.4`) & Risiko Tersisa

### A. Remediasi `govulncheck` via Peningkatan Toolchain ke `go1.27.4`
Pada audit awal (`go1.26.8`), `govulncheck` mencatat **2** kerentanan standard library di `shared/authsession`, **13** di `services/iam`, dan **12** di `gateway`. Toolchain backend kini telah diperbarui ke **`go1.27.4`** di [`backend/go.work`](file:///home/muriarfad/Documents/projects/erp-retail/backend/go.work), kelima berkas `go.mod` (`shared/{authcontext,authsession,authz}`, `services/iam`, `gateway`), serta [`backend/services/iam/Dockerfile`](file:///home/muriarfad/Documents/projects/erp-retail/backend/services/iam/Dockerfile) dan [`backend/gateway/Dockerfile`](file:///home/muriarfad/Documents/projects/erp-retail/backend/gateway/Dockerfile) (terverifikasi SHA256 resmi `amd64`/`arm64`).

Hasil pemindaian ulang `govulncheck` menggunakan **`go1.27.4`**:
- `backend/shared/{authcontext,authsession,authz}`: **`No vulnerabilities found`** (**0** temuan stdlib maupun modul).
- `backend/services/iam` & `backend/gateway`: **`0 vulnerabilities` yang memengaruhi kode ataupun package terimpor** (`Your code is affected by 0 vulnerabilities`; hanya tersisa **1** peringatan level modul non-impor `GO-2026-5932` untuk `golang.org/x/crypto/openpgp` yang tidak diimpor/dipanggil oleh kode).

### B. Risiko Tersisa / Tradeoff Operasional (Advisory)
1. **Tradeoff `registerCustomer` (`CONFLICT` vs Strict OWASP Neutral Response) [Skor DREAD: 4.8 Medium]:** Saat ini `registerCustomer` sengaja tetap mengembalikan `CONFLICT` atas duplikasi email/username sesuai kontrak uji [`iam_operations_e2e_test.go`](file:///home/muriarfad/Documents/projects/erp-retail/backend/gateway/internal/e2e/iam_operations_e2e_test.go). Jika kebijakan produk kelak mewajibkan *anti-enumeration* mutlak pada registrasi publik, pertimbangkan pola respons netral sukses + email notifikasi internal.
2. **Konfigurasi Pengembangan Docker Compose Lokal [Skor DREAD: 5.2 Medium — Hanya Lingkungan Dev]:** `docker-compose.yml` menjalankan IAM pada mode default `development` dan Gateway berbicara dengan IAM di dalam jaringan Compose di belakang TLS Caddy. Untuk *staging/production*, pastikan `APP_ENV=production`, rotasi rahasia penuh, keyring CSRF/TOTP wajib terisi, dan gRPC mTLS antar layanan diaktifkan.
3. **Header `x-powered-by: Next.js` — Status: Ditutup & Terverifikasi:** Menambahkan `poweredByHeader: false` di [`frontend/next.config.ts`](file:///home/muriarfad/Documents/projects/erp-retail/frontend/next.config.ts) serta menyalin `next.config.ts` ke stage `runtime` pada [`frontend/Dockerfile`](file:///home/muriarfad/Documents/projects/erp-retail/frontend/Dockerfile) agar `next start` di container memuat konfigurasi tersebut. Verifikasi langsung terhadap container `:3333`/`:3334` dan `https://erp.localhost:8443/{,login,register,identity}` mengonfirmasi header `X-Powered-By` tidak lagi dipancarkan.

---

## 5. Ringkasan Verifikasi Pasca-Perbaikan Terakhir

Seluruh pengujian berikut dijalankan setelah seluruh perbaikan kode, upgrade `go1.27.4`, dan rebuild Docker Compose:
- `backend/shared/{authcontext,authsession,authz}` (`go1.27.4`): `GOWORK=off go fmt ./... && GOWORK=off go vet ./... && GOWORK=off go test -race -count=1 ./...` → **PASS** (`1.005s–1.007s` per modul, exit code `0`).
- `backend/services/iam` (`go1.27.4`): `GOWORK=off go fmt ./... && GOWORK=off go vet ./... && GOWORK=off go test -race -count=1 ./...` → seluruh **20** package bertest **PASS** (exit code `0`).
- `backend/gateway` (`go1.27.4`): `GOWORK=off go fmt ./... && GOWORK=off go vet ./... && GOWORK=off go test -race -count=1 ./...` → seluruh **6** package bertest (`adapter/iam`, `browserauth`, `config`, `gqltools/test`, `middleware`, `server`) **PASS** (exit code `0`).
- `govulncheck` (`go1.27.4`): **`0`** kerentanan terpanggil/terimpor di kelima modul backend.
- `frontend`: `npx next typegen && npm run lint && npx tsc --noEmit && npm run build` (`Next.js 16.4.6`) dan `node --test src/lib/graphql/iam-browser-auth.test.ts` (`4/4`) → **PASS** (exit code `0`).
- Docker Compose, Biner Container & Probe HTTP/HTTPS: image `integrione-retail-iam:dev`, `integrione-retail-gateway:dev`, dan `integrione-retail-frontend:dev` selesai dibangun (`go version -m` pada `/usr/local/bin/iam-server` & `/usr/local/bin/gateway-server` terverifikasi **`go1.27.4`**); seluruh 8 container (`postgres`, `redis`, `iam`, `gateway-1`, `gateway-2`, `frontend-1`, `frontend-2`, `caddy`) berstatus **`healthy`**; `curl` ke `:3333/login`, `:3334/login`, serta `https://erp.localhost:8443/{,login,register,identity}` tidak lagi menyertakan `X-Powered-By` dan log `frontend-1` mencatat `✓ Running next.config.ts`; `/api/graphql?query=%7Bhealth%7D` mengembalikan `HTTP/2 200` `{"data":{"health":"ok"}}`.
- Live HTTPS E2E (`backend/gateway`): `GOWORK=off go test -tags=e2e -v -count=1 -timeout 10m ./internal/e2e` → **PASS** (**`71`** test top-level, **`140`** subtest, **`0`** fail, **`0`** skip, **`57.033s`**).
- Indeks Pengetahuan Kode (`codebase-memory-mcp` `index_repository` mode `full`): berstatus **`indexed`** (`26939` nodes, `73483` edges).
