# Panduan Standar: Coding Microservice gRPC Management (`backend/services/{management}`)

Dokumen ini adalah panduan operasional langkah demi langkah (*step-by-step*) untuk membangun microservice gRPC management di repositori ini agar rapi, modular (*anti-motherfile*), aman, dan mudah dikerjakan oleh agen developer mana pun. Pola referensi utama yang sudah berjalan di produksi lokal adalah layanan [Enterprise Foundation](../backend/services/enterprise_foundation/) dan [IAM](../backend/services/iam/).

---

## 0. Aturan Fondasi & Bahasa (Wajib Sebelum Mulai)

1. **Baca Dokumen Acuan:**
   - [AGENTS.md](../AGENTS.md) — konvensi kode, migrasi `000001`, penamaan repository, *anti-motherfile*.
   - [agents/boundary.md](boundary.md) — kepemilikan tabel/data & batas 13 management aktif.
   - [agents/arsitektur.md](arsitektur.md) — keputusan D01–D12, *trust boundary*, DDD, dan syarat E2E.
   - [agents/users.md](users.md) — model objek `users`, RBAC–ABAC (`ScopeType`), batas kepemilikan lintas modul, serta larangan menyimpan data bisnis/grant di `users` maupun akses langsung DB IAM.
   - File konteks terbaru domain target pada `agents/context/{management}/{id}-context.md`.
2. **Aturan Bahasa Ketat (Tanpa Pengecualian):**
   - **Kode implementasi (package, file, struct, method, var, konstanta, proto, SQL, log teknis):** **Bahasa Inggris**.
   - **Error code & pesan error:** **Bahasa Inggris** (contoh: `"address repository is required"`, `enterprise_foundation.invalid_input`).
   - **Komentar kode (`// ...`, kecuali direktif `//go:embed`, `//go:generate`, `//go:build`):** **Bahasa Indonesia**.
   - **Dokumentasi di `agents/` & komunikasi agen:** **Bahasa Indonesia**.
3. **Catatan Tooling Aktual & Aturan Kerja Repositori:**
   - Workspace Go berada di [backend/go.work](../backend/go.work) (`go 1.27.4`). Tidak ada `Makefile` di root/backend maupun `backend/contracts/`; setiap service memiliki direktori kontrak sendiri di `backend/services/{management}/contract/`.
   - **Cek Status Modul:** pastikan management target termasuk yang aktif dan tidak berstatus `PAUSED` di [agents/boundary.md](boundary.md) sebelum membangun service atau mengaktifkan Compose/Gateway.
   - **Riset Kode & API Resmi:** selalu periksa struktur lewat `codebase-memory-mcp` (`search_graph`/`trace_path`/`get_code_snippet`) sebelum mengedit kode, dan gunakan `go doc <pkg> <Symbol>` agar tidak menebak signature library eksternal.
   - **Git `main`-only & Handover:** dilarang membuat branch baru; stage spesifik per file (`git add <path>`). Catat handover hanya pada `agents/context/{management}/{id}-context.md` dan `agents/working_history/{management}/...` — **jangan** pernah membuat `context.md` di root atau langsung di `agents/`.
   - **Anti-Motherfile & Anti-Junk-Drawer:** Pecah file per domain & satu tanggung jawab kohesif. Dilarang membuat file `utils.go` / `common.go` campuran; beri nama spesifik (`servicekit/helpers.go`, `time_helper.go`).

---

## 1. Struktur Direktori Kanonik

Gunakan struktur persis seperti [backend/services/enterprise_foundation/](../backend/services/enterprise_foundation/):

```text
backend/services/{management}/
├── go.mod                                  # replace shared/authcontext => ../../shared/authcontext
├── Dockerfile                              # multi-stage ubuntu:26.10, Go 1.27.4, non-root user 10001
├── .dockerignore
├── cmd/
│   ├── main.go                             # entrypoint + signal handling + graceful shutdown
│   ├── wiring.go                           # BuildDependencies (DB, repos, outbox relay, services)
│   ├── wiring_test.go
│   └── healthcheck/
│       ├── main.go                         # CLI probe grpc.health.v1.Health untuk Docker HEALTHCHECK
│       └── main_test.go
├── config/
│   ├── config.go                           # env loader + Validate() fail-closed di production
│   └── test/config_test.go
├── contract/
│   ├── buf.yaml                            # STANDARD lint + FILE breaking
│   ├── buf.gen.yaml                        # paths=source_relative -> out: gen
│   ├── generate.go                         # //go:generate buf lint & buf generate
│   ├── proto/{management}/v1/*.proto
│   └── gen/{management}/v1/*.pb.go         # hasil generate (DILARANG edit manual)
├── db/
│   ├── schema/{management}.dbml            # DBML otoritatif owner
│   ├── migrations/
│   │   ├── 000001_create_{management}_schema.up.sql
│   │   └── 000001_create_{management}_schema.down.sql
│   ├── database.go                         # EnsureDatabase + DeriveTestDatabaseDSN
│   ├── migrate.go                          # //go:embed migrations/*.sql + RunMigrationsUp + advisory lock
│   ├── init.go                             # Init(ctx, dsn, isTesting) + InitDevAndTest
│   └── test/init_test.go
├── domain/
│   ├── entity/                             # pure Go structs, enum, Validate*(), tanpa tag GORM/proto
│   ├── errors/                             # Kind, Code, DomainError, sentinel errors per subdomain
│   ├── repository/                         # interface Port + PageRequest/PageResult + Transactor
│   └── service/                            # policy murni (hierarchy/scope/outbox_builder) + Clock/IDGen
├── infrastructure/
│   └── outbox/                             # Relay, Notifier (after-commit), Backoff, Publisher
├── repository/
│   ├── pgstore/                            # tx.go (WithinTransaction + after-commit), query.go, errors.go
│   ├── registry.go                         # repository.New(gormDB)
│   └── {entity}/                           # {entity}_repository.go, models.go (GORM tag), mappers.go
├── services/
│   ├── registry.go                         # services.NewRegistry(...) + Validate()
│   ├── internal/servicekit/                # helper spesifik lintas subdomain (validasi batch UUID, outbox)
│   └── {subdomain}/
│       ├── input.go                        # struct CreateInput/UpdateInput terpisah
│       ├── service.go                      # logika bisnis, check-before-create, pagination, outbox tx
│       └── test/service_test.go            # unit test memakai mock repo (tanpa DB asli)
└── transport/
    └── grpc/
        ├── server.go                       # KeepaliveParams, EnforcementPolicy, ChainUnaryInterceptor, Health
        ├── interceptor/interceptor.go      # Recovery, InternalContext HMAC, Logging
        └── handler/
            ├── register.go                 # RegisterAll(...) fail-closed
            ├── errors.go                   # ErrorMapper: DomainError -> gRPC status + ErrorInfo/BadRequest
            ├── mapper_*.go                 # konversi protobuf <-> domain entity
            └── {subdomain}_handler.go      # validasi req != nil, panggil service, petakan error via toStatus
```

Jangan lupa daftarkan `./services/{management}` di [backend/go.work](../backend/go.work).

---

## 2. Posisi Roles & Permissions di Microservice gRPC Management

Sebelum menulis service dan kontrak, pahami pembagian tugas otorisasi agar tidak melanggar boundary sekaligus memastikan role & permission otomatis tersimpan ke database production:

1. **Siapa yang Menentukan Roles–Permissions?**
   - **Agen developer pemilik management** merancang dan menentukan katalog permission (`<management>.<resource>.<action>`), penanda operasi sensitif (`is_sensitive`), scope ABAC yang didukung (`global`, `legal_entity`, `facility`, `organization_unit`, `vendor_party`, `supplier_party`, `self`, dll.), dan kandidat *template role* domain dengan berpedoman pada [agents/users.md](users.md) serta [agents/boundary.md](boundary.md).
2. **Kapan Menulis Dokumen Keputusan & Manifest Deklaratif?**
   - Tulis **`agents/roles-permissions/{management}/1-roles-permissions.md`** dan manifest kode **`backend/shared/rbac/catalogs/{management}.json`** **sebelum atau bersamaan dengan pembuatan kontrak `.proto` dan sebelum wiring ke Gateway** (hanya daftarkan operasi yang kontraknya benar-benar aktif). Setiap penambahan fitur/RPC baru wajib memperbarui kedua dokumen tersebut serta diuji konsistensinya lewat unit test anti-*drift* (`catalog_test.go`) sebelum di-expose di Gateway.
   - Manifest domain **dilarang** menyentuh prefix modul lain, mengubah role `superadmin` (tidak ada pewarisan otomatis), atau menulis langsung ke DB `iam`. Bila modul membutuhkan `default_redirect_path` baru di luar allowlist `RedirectResolver` IAM atau akun `auto_system_{mgmt}` baru, eskalasikan melalui `agents/notes/iam_management/`. *(Selama engine loader manifest di IAM belum selesai diimplementasikan, catat kebutuhan seed aktif melalui handoff IAM yang relevan.)*
3. **Di Mana Pengecekan Roles–Permissions Dilakukan?**
   - **BUKAN lewat query DB IAM dari dalam service bisnis!** Microservice gRPC domain **dilarang** mengakses database `iam`, **dilarang** menduplikasi tabel role/permission, dan **dilarang** membaca hak dari JSON.
   - **Tanggung jawab Microservice gRPC Domain:**
     1. Memverifikasi bukti HMAC pemanggil internal dari Gateway di `InternalContextUnaryInterceptor` (`principal_id`, `request_id` lewat `authcontext.ExtractFromIncomingGRPCContext`). Bila bukti tidak sah/kosong -> kembalikan `codes.Unauthenticated`.
     2. Mengecek aturan integritas/kepemilikan bisnis internal serta menyediakan RPC **pembuktian relasi objek scope** (contoh di EF: `ValidateLegalEntityScopeTarget`, `ValidateFacilityScope`, `ValidatePartyRoleScope`) bila objek tersebut dipakai sebagai `scope_id` ABAC.
   - **Tanggung jawab Gateway Adapter:** seluruh penegakan `RequirePermission`, `is_sensitive` step-up reauth, dan pencocokan `scope` dilakukan di Gateway **sebelum** request masuk ke gRPC domain (lihat [how_to_code_wiring_grpc_managements_to_gateway.md](how_to_code_wiring_grpc_managements_to_gateway.md)).

---

## 3. Urutan Pengerjaan Langkah Demi Langkah (Domain hingga gRPC Server Jadi)

Ikuti urutan **Tahap 1 -> Tahap 9** secara berurutan agar setiap lapisan langsung bisa diuji tanpa berputar-putar:

### Tahap 1 — Analisis DBML, Provisioning DB Role, & Migrasi `000001` (`db/`)
1. Baca [db/schema/{management}.dbml](../backend/services/enterprise_foundation/db/schema/enterprise_foundation.dbml) milik service. Pastikan tidak ada FK lintas database (simpan ID eksternal seperti `iam_user_id` sebagai UUID logis biasa).
2. **Provisioning Role & DB Terisolasi ([backend/deploy/postgres/init/](../backend/deploy/postgres/init/)):**
   - Ikuti pola [02-init-enterprise-foundation.sql](../backend/deploy/postgres/init/02-init-enterprise-foundation.sql): buat file `0N-init-{management}.sql` untuk membuat role khusus `{management}_svc` (membaca env `SERVICES_DB_PASSWORD`, jangan hardcode kredensial produksi di SQL), database `{alias}` serta `{alias}_test` milik role tersebut, dan ekstensi `pgcrypto` / `pg_trgm`.
   - **Gotcha volume `pgdata18`:** `/docker-entrypoint-initdb.d/` hanya berjalan otomatis saat volume Postgres masih kosong. Jika Postgres lokal sudah berjalan, jalankan script init baru satu kali secara manual via `docker compose exec postgres psql ...`.
3. Tulis `db/migrations/000001_create_{management}_schema.up.sql` & `.down.sql`.
   - **Aturan keras pra-rilis:** **hanya boleh ada versi `000001`**. Perubahan skema pra-rilis dikonsolidasi ke `000001`, jangan buat `000002`+.

### Tahap 2 — Auto-Init & Migrasi Database (`db/database.go`, `migrate.go`, `init.go`)
1. Embed SQL via `//go:embed migrations/*.up.sql migrations/*.down.sql`.
2. Implementasikan `EnsureDatabase` (auto-create DB target via database `postgres`), `RunMigrationsUp` (dengan **PostgreSQL advisory lock ID `int64` unik** per service, contoh EF `718291824`), dan `Init(ctx, dsn, isTesting)`.
3. Konfigurasikan pool GORM standar: `MaxIdleConns=10`, `MaxOpenConns=50`, `ConnMaxLifetime=30m`, `ConnMaxIdleTime=5m`.
4. Di `Init(..., isTesting=true)`, wajib fail-fast jika nama database pada DSN tidak berakhiran `_test`. Test DB nyata **wajib** memakai container Postgres (`127.0.0.1:6543`); dilarang menginstal PostgreSQL langsung di host OS.

### Tahap 3 — Lapisan Domain Murni (`domain/`)
1. `domain/entity/`: definisikan struct bisnis murni, enum, serta method/helper `Validate*` tanpa struct tag GORM ataupun protobuf.
2. `domain/errors/`:
   - Definisikan `Kind`: `KindInvalidInput`, `KindNotFound`, `KindAlreadyExists`, `KindUnauthorized`, `KindForbidden`, `KindConflict`, `KindRateLimited`, `KindUnavailable`.
   - Definisikan `DomainError{Kind, Code, Message, Field, Err}` dan sentinel error per entitas (contoh `ErrAddressNotFound`, `ErrLegalEntityCodeExists`).
3. `domain/repository/`: definisikan interface port (`{Entity}Repository`), `PageRequest`, `PageResult`, serta interface `Transactor { WithinTransaction(ctx, func(ctx) error) error }`.
4. `domain/service/`: letakkan policy murni (misal validasi siklus hierarki, pembangunan payload outbox, serta port `Clock` dan `IDGenerator`).

### Tahap 4 — Lapisan Repository GORM & PGStore (`repository/`)
1. Implementasikan [pgstore/tx.go](../backend/services/enterprise_foundation/repository/pgstore/tx.go):
   - `WithinTransaction` menyimpan `*gorm.DB` transaksi ke `ctx` (`txContextKey`) dan mengumpulkan `afterCommitCollector`. Hanya setelah transaksi terluar sukses commit, callback dijalankan lewat `RegisterAfterCommit` memakai `context.WithoutCancel(ctx)`.
2. Implementasikan [pgstore/errors.go](../backend/services/enterprise_foundation/repository/pgstore/errors.go) (`MapError` & `MapDeleteError`):
   - `gorm.ErrRecordNotFound` -> `notFoundErr` / `ErrNotFound`
   - SQLSTATE `23505` (*unique violation*) -> lookup nama constraint ke `Err*AlreadyExists` (`KindAlreadyExists`)
   - SQLSTATE `23503`, `23001` (*FK / restrict*) -> `inUseErr` (`KindConflict`) pada delete, atau `Err*NotFound` pada insert/update
   - SQLSTATE `23514` (*check constraint*), `22P02`, `23502`, `22001` -> `KindInvalidInput`
   - SQLSTATE `40001`, `40P01` (*serialization / deadlock*) -> `KindConflict`
   - SQLSTATE `08*`, `53*`, `57*` / canceled / deadline -> `KindUnavailable`
   - **Wajib:** error database tak terduga dibungkus dengan pesan aman tanpa membocorkan raw SQL/nama tabel!
3. Implementasikan `repository/{entity}/`:
   - Penamaan method **wajib** memakai operasi SQL murni: `Select*`, `Insert*`/`Create*`, `Update*`, `Delete*`/`SoftDelete*`, `Check*`. Dilarang memakai nama use-case bisnis.
   - `SelectList` **wajib** mendukung filter + `Count(&total)` + `Limit`/`Offset` + `Order` deterministik (`PageResult{Items, Total, Limit, Offset, HasMore}`).

### Tahap 5 — Lapisan Application Services & Unit Test Mock (`services/`)
1. Pisahkan `input.go` (`CreateInput`, `UpdateInput`), `service.go`, dan `test/service_test.go` per sub-domain. Pastikan struct `Service` stateless per request (aman `-race`).
2. **Aturan CRUD, Pagination, & ServiceKit Wajib:**
   - **Check-before-create (anti-duplikasi):** selalu periksa kunci unik (`SelectByCode` atau filter atribut kunci) sebelum `Insert` sehingga duplikasi langsung mengembalikan `ErrAlreadyExists`.
   - **Pagination wajib:** ikuti [filter.go](../backend/services/enterprise_foundation/domain/repository/filter.go) (`PageRequest.Normalize()` dengan `DefaultPageLimit = 20`, `MaxPageLimit = 100`, `NewPageResult`). Jangan pernah mengekspos `GetAll` tanpa paginasi! Jika agregasi internal benar-benar perlu menarik multi-halaman, gunakan pola [servicekit/helpers.go](../backend/services/enterprise_foundation/services/internal/servicekit/helpers.go) (`CollectAllPages` dengan batas aman `MaxCollectAllItems = 10000`) dan `eferrors.WithField(...)` saat validasi FK gagal.
   - **Identitas Pemanggil & Outbox Transaksional:** baca audit caller via `caller, ok := authcontext.FromContext(ctx)` (`PrincipalID`, `RequestID`) tanpa mempercayai ID pembuat dari input payload. Bila operasi memancarkan domain event, eksekusi mutasi + `outboxRepo.Insert` bersama di dalam `s.tx.WithinTransaction(...)`.
3. Daftarkan seluruh service di `services/registry.go` lengkap dengan `Validate()`.
4. Jalankan unit test `services/{subdomain}/test/` murni dengan mock repository (tanpa menyentuh database asli).

### Tahap 6 — Infrastruktur Outbox Asinkron (Jika Ada Event Lintas Service)
1. Ikuti pola [entity/outbox.go](../backend/services/enterprise_foundation/domain/entity/outbox.go) (`event_type`, `aggregate_type`, `aggregate_id`, `payload` JSON object tervalidasi, `status = PENDING`) dan [infrastructure/outbox/](../backend/services/enterprise_foundation/infrastructure/outbox/): `Relay` (mode `standby` / `webhook` / `log`, dengan `log` dilarang di `production`), exponential backoff, pembersihan berkala, dan wrapper `outbox.WrapNotifyingRepository(repos.Outbox, relay, pgstore.RegisterAfterCommit)` pada `cmd/wiring.go` khusus untuk lapisan `services/`.

### Tahap 7 — Kontrak Protobuf (`contract/`) & Katalog/Manifest Permission
1. Pastikan `agents/roles-permissions/{management}/1-roles-permissions.md` dan manifest `backend/shared/rbac/catalogs/{management}.json` sudah ditulis untuk memetakan operasi yang akan dibuka (khusus operasi aktif berstatus implementasi, bukan stub), serta terverifikasi oleh unit test konsistensi katalog (`catalog_test.go`).
2. Tulis `.proto` mengikuti [common.proto](../backend/services/enterprise_foundation/contract/proto/enterprise_foundation/v1/common.proto) dan [address.proto](../backend/services/enterprise_foundation/contract/proto/enterprise_foundation/v1/address.proto):
   - Enum dimulai `{ENUM}_UNSPECIFIED = 0;`.
   - ID menggunakan `string`, timestamp menggunakan `google.protobuf.Timestamp`, field opsional memakai `optional`, pagination memakai `PageRequest` & `PageInfo`.
   - Jangan pernah mengubah tag/nama field yang sudah dirilis (gunakan `reserved`).
3. Jalankan generator resmi:
   ```bash
   cd backend/services/{management}/contract && go generate ./...
   ```
   Pastikan `buf lint` dan `buf generate` selesai tanpa warning/error.

### Tahap 8 — Transport gRPC, Interceptor, & ErrorMapper (`transport/grpc/`)
1. **Handler tipis (`handler/{subdomain}_handler.go`):**
   - Cek `if req == nil { return nil, nilRequestError(h.errMapper) }`.
   - Petakan request proto -> `service.Input`, panggil `h.svc.*`, konversi response, dan bila `err != nil`, kembalikan `nil, toStatus(h.errMapper, err)`. Dilarang menaruh logika bisnis atau transaksi SQL di handler.
2. **ErrorMapper (`handler/errors.go`) — Peta Error gRPC Baku:**

   | `DomainError.Kind` | Kode gRPC (`codes.*`) | Catatan Keamanan & Detail |
   | --- | --- | --- |
   | `KindInvalidInput` | `codes.InvalidArgument` | Attach `ErrorInfo{Reason, Domain}` + `BadRequest` jika `Field != ""` |
   | `KindNotFound` | `codes.NotFound` | Attach `ErrorInfo` |
   | `KindAlreadyExists` | `codes.AlreadyExists` | Attach `ErrorInfo` |
   | `KindConflict` | `codes.FailedPrecondition` | Attach `ErrorInfo` |
   | `KindUnauthorized` | `codes.Unauthenticated` | Attach `ErrorInfo` |
   | `KindForbidden` | `codes.PermissionDenied` | Attach `ErrorInfo` |
   | `KindRateLimited` | `codes.ResourceExhausted` | Attach `ErrorInfo` |
   | `KindUnavailable` | `codes.Unavailable` | Log `Error` di server; attach `ErrorInfo` |
   | Non-domain / unknown | `codes.Internal` | Log detail asli di server; pesan ke gRPC **wajib `"internal error"`** |

   *(Selain itu: `context.Canceled` -> `codes.Canceled`, `context.DeadlineExceeded` -> `codes.DeadlineExceeded`.)*
3. **Registrasi & Server (`handler/register.go` & `transport/grpc/server.go`):**
   - `RegisterAll` mengecek fail-closed seluruh dependency handler.
   - `ChainUnaryInterceptor` berurutan: `RecoveryUnaryInterceptor` -> `InternalContextUnaryInterceptor(hmacSecret, logger)` -> `LoggingUnaryInterceptor`.
   - Pasang `KeepaliveParams{MaxConnectionIdle: 5m, Time: 2m, Timeout: 20s}` dan `EnforcementPolicy{MinTime: 20s, PermitWithoutStream: true}` (aman di bawah interval ping `60s` Gateway).
   - Daftarkan `grpc.health.v1.Health` (`SERVING` untuk `""` dan `cfg.ServiceName`).
   - **Pengujian Transport gRPC (`transport/grpc/test/`):** uji interceptor HMAC, *panic recovery*, `ErrorMapper`, dan health probe memakai `google.golang.org/grpc/test/bufconn` tanpa membuka port host.

### Tahap 9 — Config Strict, Sinkronisasi HMAC/TZ, Container Hardening, & Port
1. **`config/config.go`:** Di `APP_ENV=production`, wajib tolak jika `DATABASE_DSN` kosong atau `{MGMT}_INTERNAL_GATEWAY_HMAC_SECRET < 32 byte` atau mode outbox `log`.
2. **Sinkronisasi Secret & `TZ` dengan Gateway:** nilai `{MGMT}_INTERNAL_GATEWAY_HMAC_SECRET` wajib **identik byte-per-byte** dengan `GATEWAY_INTERNAL_HMAC_SECRET` / `IAM_INTERNAL_GATEWAY_HMAC_SECRET`, dan wajib memasang `TZ: Asia/Jakarta` + paket `tzdata` agar jendela timestamp verifikasi HMAC `authcontext` selaras.
3. **`cmd/wiring.go` & `cmd/main.go`:** tangkap `SIGINT`/`SIGTERM`, jalankan gRPC server & outbox relay secara konkuren, shutdown anggun dalam `ShutdownTimeout`, dan `os.Exit(1)` bila worker/server exit dengan error (agar Docker `restart: unless-stopped` otomatis me-restart container).
4. **`cmd/healthcheck/main.go` & [Dockerfile](../backend/services/enterprise_foundation/Dockerfile):** multi-stage `ubuntu:26.10` + `go1.27.4`, build context di Compose **wajib `.`** (`backend/`) supaya `COPY shared/authcontext` berhasil, jalankan dengan non-root `10001:10001` dan `HEALTHCHECK CMD ["/usr/local/bin/grpc-healthcheck"]`.
5. **[backend/docker-compose.yml](../backend/docker-compose.yml) & Port:**
   - Port terpakai saat ini: `9001` (`iam`), `9002` (`enterprise-foundation`); gunakan port berurutan berikutnya (`127.0.0.1:9003:9003`, dst.).
   - Pasang `depends_on: postgres (service_healthy)`, `restart: unless-stopped`, `read_only: true` *(dilarang menulis file/log di luar `/tmp`; semua log wajib ke `stdout`)*, `tmpfs: [/tmp]`, `cap_drop: [ALL]`, `security_opt: [no-new-privileges:true]`.
   - **Jangan** tambahkan service baru ke `depends_on` milik `gateway-1`, `gateway-2`, atau `iam`.

---

## 4. Anti-Pola yang Dilarang

- Membuat Git branch baru atau melakukan force-push / `git add -A` saat ada pekerjaan agen lain.
- Mengaktifkan/meng-wire service yang berstatus `PAUSED` di [boundary.md](boundary.md) tanpa keputusan owner.
- Memberi nama method bisnis (`Register*`, `Approve*`, `Checkout*`) pada lapisan `repository/`.
- Mengekspos query/RPC list atau `GetAll` tanpa paginasi terikat (`DefaultPageLimit = 20`, `MaxPageLimit = 100`).
- Mengakses database `iam` / membuat FK lintas database, menyimpan role/grant di JSON, menyentuh manifest `backend/shared/rbac/catalogs/` milik management lain, atau mengedit `bootstrap.go` IAM untuk permission rutin domain.
- Menggabungkan banyak domain/entitas dalam satu file (*motherfile*) atau membuat file `utils.go`/`common.go`.
- Membocorkan detail SQL/GORM (`sqlstate`, nama tabel/index, IP) pada pesan gRPC `status.Error`.
- Membuat file migrasi `000002`+ selama masa pra-rilis, atau menjalankan test terhadap database non-`*_test` / PostgreSQL host.
- Lupa memasang `InternalContextUnaryInterceptor`, `grpc-healthcheck`, `read_only: true`, atau menambahkan service domain ke `depends_on` Gateway/IAM.

---

## 5. Checklist Verifikasi Wajib Sebelum Selesai

```bash
# 1. Lint & generate Protobuf
cd backend/services/{management}/contract && go generate ./...

# 2. Format, vet, dan seluruh unit test (termasuk -race)
cd backend/services/{management}
gofmt -s -w .
go vet ./...
go test -race -count=1 ./...

# 3. Build & uji container berjalan sehat
cd ../../..
docker compose build {service_name}
docker compose up -d --wait {service_name}
docker compose ps
docker logs --tail 30 integrione-retail-{service_name}-1
```

Pastikan status container `running (healthy)`, tidak ada error di logs, catat konteks di `agents/context/{management}/{id}-context.md` & working history di `agents/working_history/{management}/...`, dan jalankan re-index `codebase-memory-mcp`.
