# Implementasi server gRPC CRM dan wiring GraphQL Gateway

- **id** — 1
- **owner** — Full Backend Developer Customer Relationship & Sales. Reviewer Leader Backend.
- **priority** — Internal sales dan supervisor, dilanjutkan koordinasi B2B, kemudian fitur komunikasi yang dependensinya tersedia.
- **status** — SIAP DIKERJAKAN
- **baseline** — Checkout bersama `main`, basis `3459732f459439513d9f0c9fc68c5b63959d5a43`, 29 November 2026 WIB.

## Task

Bangun satu service privat CRM mulai dari data PostgreSQL sampai handler gRPC, lalu hubungkan operasi petugas sales ke HTTPS GraphQL. Pada checkout yang diperiksa, folder service hanya berisi [DBML 35 tabel, 23 enum, 67 Ref lokal](../../../backend/services/customer_relationship_sales/db/schema/customer_relationship_management.dbml). Modul Go, migrasi, kontrak CRM dan SDL CRM adalah pekerjaan baru. Commerce sudah memiliki referensi opportunity opsional pada quotation. [Katalog per operasi](../../../agents/grpc/customer_relationship_sales/1-katalog-operasi-relasi-pelanggan-dan-penjualan.md) menetapkan target 35 Query, 54 Mutation dan 10 RPC khusus workload. Tuntaskan setiap irisan melalui jalur pemakai dan catat status implementasi di katalog.

**Penamaan.** Gunakan direktori yang diberikan owner `backend/services/customer_relationship_sales/` sebagai root source dan suffix modul Go. Header DBML menetapkan `customer_relationship_management` sebagai nama service/database. Gunakan nama ini untuk alias DB dan namespace proto `customer_relationship_management.v1`. Selaraskan nama DNS Compose dan env client di satu tempat. Katalog izin usulan mengikuti folder source `customer_relationship_sales`. `crm_sales_management` pada boundary/konteks historis menunjuk management yang sama. Gunakan satu instance server dan satu database, periksa registry yang sudah ada sebelum menambah mapping. Catat mapping di context saat implementasi.

**Keputusan domain.** Relasi/lead/pipeline dan izin kontak menjadi pekerjaan CRM. Commerce memiliki quotation/submission, Pricing menentukan harga, OMS pemilik keputusan order, WMS pemilik stok, Finance memiliki actual revenue/credit, HRM pemilik insentif. Untuk tiga tabel desain kasus, ikuti boundary otoritatif CS sebagai pemilik tiket dan SLA. CRM mengajukan handoff dan membaca status terotorisasi. Simpan kampanye/segmentasi/broadcast untuk tahap Marketing yang mendapat mandat owner. [Boundary](../../../agents/boundary.md) · [CRM dan Customer Service](../../../docs/03-managements/crm-and-customer-service.md) · [OMS dan Commerce](../../../docs/03-managements/oms-and-commerce.md)

## Request

### Peta file dari DB ke handler dan Gateway

Path relatif di tabel ini adalah **CREATE** di dalam `backend/services/customer_relationship_sales/`, kecuali DBML yang sudah disediakan. `<subdomain>` diisi nama paket pada katalog. Tulis handler dan service per domain.

| Layer | File yang dibuat | Function / simbol | Hasil |
| --- | --- | --- | --- |
| Modul dan bootstrap | `go.mod, go.sum, Makefile, Dockerfile, config/config.go, cmd/main.go, cmd/wiring.go` | `NewConfig, BuildApplication, Run` | Periksa DSN owner, TLS, keyring aktif, shutdown, migrasi dan readiness privat. Daftarkan modul di backend/go.work dan dependensi lokal Gateway sesuai modul lain. |
| Database | `db/database.go, db/migrate.go, db/init.go, db/migrations/000001_create_customer_relationship_management_schema.up.sql dan .down.sql` | `Init, RunMigrationsUp` | Baseline SQL eksplisit, advisory lock, up/down berurutan dan role khusus. Tambahkan init DB dev/test mengikuti backend/deploy/postgres/init/. Hindari AutoMigrate. |
| Domain | `domain/entity/<subdomain>.go, domain/errors/errors.go, domain/repository/<subdomain>.go, domain/service/ports.go` | `Validate, Transactor.WithinTransaction, OwnerEvidencePort` | Enum dan decimal domain, ID opsional, clock/ID factory, evidence dengan producer, versi, tanggal dan digest. Port owner diinjeksikan. |
| Repository | `repository/pgstore/tx.go, repository/pgstore/errors.go, repository/<subdomain>/{models.go,mappers.go,repository.go}, repository/registry.go` | `SelectByIDForUpdate, SelectPage, Insert, UpdateVersion` | GORM mengambil transaksi dari context. Scope menjadi predikat SQL sebelum count/page. Gunakan placeholder, urutan deterministik dan row lock terurut. |
| Service | `services/<subdomain>/{service.go,input.go}, services/registry.go, services/<subdomain>/test/*_test.go` | `Method per RPC pada katalog` | Subdomain mengikuti sembilan paket publik dan integration. Business rule, scope, expected version, audit, outbox dan transaksi di sini. Mock repository/clock/port hanya untuk unit test services. |
| Kontrak | `contract/{buf.yaml,buf.gen.yaml,generate.go}, contract/proto/customer_relationship_management/v1/{common,<subdomain>,integration}.proto, contract/gen/` | `Register…Server, ValidateRequest` | Tipe terpisah per command, oneof untuk jenis evidence, decimal sebagai string kanonik, timestamp UTC dan tipe enum UNSPECIFIED yang ditolak bila wajib. Pertahankan field number. |
| Transport | `transport/grpc/{server.go,interceptor/interceptor.go}, transport/grpc/handler/{register.go,errors.go,<subdomain>_handler.go,mapper_<subdomain>.go}` | `RegisterHandlers, UnaryInterceptor, ToStatus` | Parse typed ID/UUID, cek panjang dan field wajib, ambil aktor tepercaya, petakan ke service. Listener privat, TLS, timeout, limits dan graceful stop. |
| Integrasi dan worker | `infrastructure/integration/<owner>/client.go, infrastructure/{inbox/processor.go,outbox/relay.go}, services/communication/dispatch.go` | `Claim, Dispatch, Reconcile, Acknowledge` | Lease fencing, bounded retry, verifikasi issuer, event version/digest, DLQ dan replay teraudit. Audit hanya metadata minimum. |

File bersama untuk **MODIFY** adalah `backend/go.work`, `backend/docker-compose.yml` (di-include dari root), `backend/gateway/go.mod`, `backend/gateway/gqlgen.yml`, `backend/gateway/internal/source_connections/registry.go`, `backend/gateway/graph/resolvers/resolver.go`, `backend/gateway/graph/resolvers/schema.resolvers.go` dan `backend/gateway/cmd/main.go`. **CREATE** katalog `backend/shared/rbac/catalogs/customer_relationship_sales.json`, spesifikasi `agents/roles-permissions/customer_relationship_sales/1-roles-permissions.md`, `backend/gateway/internal/source_connections/customer_relationship_sales_client.go`, `backend/gateway/internal/adapter/customer_relationship_sales/{adapter.go,permissions.go,gid.go,<subdomain>.go}` dan `backend/gateway/graphql/customer_relationship_sales/crs_<subdomain>.graphql`. `crs_` hanya prefix nama file agar output gqlgen unik. Gunakan tipe dan field bernama bisnis pada API.

### SQL, authorization dan transaksi

- file: `db/schema/customer_relationship_management.dbml, db/migrations/000001_*.sql, domain/repository/, repository/pgstore/, services/registry.go`
- function: `RunMigrationsUp, RequireScope, ExecuteIdempotentCommand`
- logic: Petakan 35 tabel dan 67 deklarasi Ref lokal. Relasi organisasi/kontak, preferensi, pipeline, opportunity, attribution dan audit dipelihara CRM. Tiga tabel customer_cases, customer_case_history, customer_case_actions dipetakan sebagai desain tertahan untuk CS. Bila dipertahankan pada baseline, nonaktifkan writer/endpoint CS. Buat pemetaan eksplisit untuk FK customer_case_id dari follow-up dan activity agar referensi eksternal tidak dimasukkan ke FK lokal. Tambah version bigint pada aggregate yang memakai CAS. Jalankan satu transaksi untuk write + history + audit + outbox + hasil command. Key terikat aktor/realm/scope/operasi dan hash request kanonik. Replay memeriksa ulang hak baca hasil.
- error_handling: SQLSTATE constraint → domain error stabil. ALREADY_EXISTS untuk key/identity bentrok, ABORTED untuk versi usang, FAILED_PRECONDITION untuk state keliru. Kesalahan insert audit/outbox me-rollback command.
- security: Session opaque + IAM snapshot dari jalur tepercaya, deny by default. Pisahkan GLOBAL, LEGAL_ENTITY, ORGANIZATION_UNIT, FACILITY dan SELF sesuai operasi. Resolusi SELF mengikuti principal/assignee, bukan ID dari request. Recheck revisi sebelum commit sensitif. Per halaman, agregat, count dan nested field menegakkan pasangan grant yang sama.
- expected_result: Dua submit sama menghasilkan status dan ID yang sama. Pelaku dari unit lain tidak dapat menyimpulkan jumlah atau keberadaan objek tersembunyi.
- testing: Uji service dengan grant unit A/B, pencabutan izin, key sama payload beda, dua writer dan rollback. Uji race serta constraint melalui HTTPS/gRPC pada lingkungan dev/test.

### Identitas dan privasi

- file: `services/relationship/, services/permission/, repository/relationship/, repository/permission/`
- function: `LinkRelationshipIdentity, MergeRelationshipContacts, RecordContactPreference, LiftContactSuppression`
- logic: Profil B2C boleh provisional. B2B boleh menjadi prospek tanpa party EF. Cocokkan identity dengan bukti principal/party/PIC/customer yang sah. Normalisasikan endpoint per namespace, HMAC dengan key version dan encrypt nilai asli. Partial unique satu primary aktif per contact/channel. Shared number menghasilkan kandidat. Saat merge lock ID terurut, tolak siklus/official ID berbeda, pertahankan canonical redirect, alias, riwayat dan suppression. Tetapkan retensi, erasure/hold dan rotasi key dengan dual lookup sebelum key lama dipensiunkan.
- error_handling: Konflik identitas → FAILED_PRECONDITION. Izin komunikasi unknown/expired/withdrawn → deny. Periksa lintas alias dan suppression berlaku. Grant lama tidak hidup kembali saat keputusan terbaru kedaluwarsa.
- security: Batasi pencarian endpoint dan redaction. Alasan, evidence ref, notice version dan waktu ada pada histori append-only. Lift independen dan teraudit. Log hanya hint tersamarkan.
- expected_result: Dua kontak yang memakai satu nomor tetap terpisah. Opt-out bertahan setelah merge, reimport dan rotasi key. Profil tidak memicu pembuatan akun IAM.
- testing: Nomor bersama, ID EF milik party lain, merge berlawanan secara paralel, key rotation, keputusan terlambat/seri dan retensi dengan legal hold.

### Lead, sales pipeline dan forecast

- file: `services/lead/, services/pipeline/, repository/lead/, repository/pipeline/`
- function: `CreateSalesLead, QualifySalesLead, ConvertSalesLead, MoveSalesOpportunityStage, GetSalesPipelineForecast`
- logic: Pisahkan assignment dari status. Kunci lead saat konversi dan pin satu bukti primer dengan subjek sama. Jika membuat opportunity dalam command, kedua row dan histori commit bersama. Validasi stage aktif dari pipeline yang sama, unique display order, minimal satu open/won/lost saat aktivasi. Lost membutuhkan alasan dan 0%, won 100% hanya dari bukti owner. Probabilitas open 0–100 sesuai kebijakan. Item/versi/UOM PLM bila diketahui, kebutuhan bebas deskriptif bila belum ada SKU. Forecast = nilai terbuka residual × probabilitas/100, berdasarkan expected_close_on. Untuk konversi parsial simpan bukti qty/net per baris sehingga confirmed line tidak dihitung dua kali.
- error_handling: Tolak stage lintas pipeline, currency berbeda, inactive assignee dan konversi lintas subjek. Race → ABORTED atau replay. Order pembatalan memicu review/reopen beraudit.
- security: Batasi sales pada self/mandat unit. Cost dan harga resmi memerlukan izin owner tersendiri. Perubahan assignee tidak mengubah historical credit.
- expected_result: Konversi lead ke opportunity dapat berlangsung tanpa penjualan. Accepted quote belum memenangkan opportunity. Konversi parsial, sisa forecast dan booking dapat dijelaskan per baris.
- testing: Dua konversi bersamaan, salah subjek, lintas stage, quote-only, order sebagian atau dua quote untuk satu pesanan.

### Koordinasi quotation, atribusi dan target

- file: `services/coordination/, services/performance/, infrastructure/integration/{web_commerce,finance_accounting,human_resource}/`
- function: `RequestAssistedQuotation, ApplyVerifiedSalesDocument, ApproveSalesCreditCorrection, ApproveSalesTarget, GetSalesAchievement`
- logic: B2B quote di Commerce, B2C lewat checkout Commerce yang sah. CRM menulis command sebelum gRPC, mencatat ACK setelah owner commit. Link unik per sumber/kind/document. Pada order terkonfirmasi bekukan credited salesperson dan unit. Untuk koreksi siapkan record permintaan dan approval beda aktor dengan effective_at, old/new value serta evidence. Target effective nonoverlap per subjek, unit, metric dan mata uang (tangani NULL unit eksplisit). Supersede atomik. Bookings net = confirmed net sebelum tax/shipping dikurangi pembatalan bagian belum fulfilled menurut kohor konfirmasi. Recognized net = revenue posted dikurangi commercial credit terkait menurut tanggal akuntansi dan credit asli. Achievement denominator nol → N/A. Target/performance bukan perintah payroll.
- error_handling: Timeout → processing/reconcile dengan key yang sama. Delta lama/duplikat ditahan atau diabaikan secara idempoten. Provenance Finance tertinggal → partial/unavailable beserta watermark.
- security: SoD target dan correction. CRM tidak dapat menyetujui harga/kredit Finance. HRM mendapat hasil minimal untuk performance dengan mandat tersendiri.
- expected_result: Perpindahan sales tidak memindahkan credit lama. Angka bookings dan recognized terpisah dan dapat ditelusuri. Credit reversal merevisi periode yang benar.
- testing: Overlap dua approval, target nol, pengembalian setelah fulfilment, dua mata uang, late credit, attribution correction as-of dan timeout quote.

### Pesan, template, lampiran dan follow-up

- file: `services/communication/, services/activity/, infrastructure/integration/document_record/, infrastructure/outbox/`
- function: `QueueRelationshipMessage, ApplyMessageDeliveryEvidence, GetRelationshipMessageAttachment, CompleteSalesFollowUp`
- logic: Bedakan channel account dari endpoint dan participant. Thread unik per account. Pakai satu penerima, template version approved, window/policy provider dan izin tujuan pada enqueue serta tepat sebelum send. Pilih mode manual untuk mencatat panggilan sebagai activity. Withdrawal membatalkan queued, in-flight direkonsiliasi. Provider timeout → unknown. Delivery event monoton sesuai bukti. Jangan fabrikasi waktu. Sesuaikan attachment saat akses/kirim pada exact DRM ID/version/digest, hasil scan clean dan izin audience. Reminder jatuh tempo didedup per occurrence dan notifikasi internal dipakai bila transport tersedia.
- error_handling: Dependency DRM/scan gagal → lampiran pending/terkunci. Unknown outcome direkonsiliasi provider sebelum mencoba kirim. Batasi retry, simpan dead letter berakses terbatas.
- security: Jangan simpan secret atau file biner di CRM. Credential reference hanya untuk workload. Auth provider pada boundary integration yang disetujui. Cek ACL conversation, descendant dan role sebelum rendering konten.
- expected_result: Internal note atau ACK provider tidak menghasilkan status delivered. Dua completion task menghasilkan satu aktivitas, reschedule membatalkan reminder lama.
- testing: Opt-out antara queue/dispatch, duplicate inbound, read sebelum delivered, provider timeout, lampiran dicabut, participant belum teridentifikasi, scheduler paralel.

### Customer 360 dan permintaan dukungan

- file: `services/relationship/overview.go, services/service_handoff/, infrastructure/integration/customer_service/`
- function: `GetCustomerRelationshipOverview, RequestRelationshipSupport, ApplySupportHandoffResult`
- logic: Komposisi opsional per bagian. Tampilkan ringkasan lokal, referensi resmi dan watermark, batasi fan-out dengan deadline. Untuk dukungan persist request dan correlation. Penerima CS memiliki tiket/SLA, first response, reopen/resolution dan eskalasi Commerce/OMS/Finance/WMS. Kesiapan kontrak CS menjadi syarat membuka dua root handoff. Rencana receipt berisi ref tiket/versi/status yang aman.
- error_handling: Kembalikan availability per bagian. Sembunyikan bagian terlarang sebelum memanggil owner. Owner belum tersedia → pending/unavailable, tanpa mencetak nomor tiket keberhasilan.
- security: Cek izin subject, unit dan field pada CRM dan owner. Handoff menyertakan minimum PII serta bukti versi yang boleh dibaca penerima.
- expected_result: Owner timeout tidak menghilangkan lead/profil lokal. Nomor tiket dan status hanya terlihat setelah ACK sah dan boleh dilihat aktor.
- testing: Partial outage Finance, subjek berbeda, repeated handoff, owner offline lalu pulih, revoke akses attachment serta penolakan perubahan refund lokal.

### Hak akses dan alur pengguna

| Aktor atau grant usulan | Perilaku yang diverifikasi |
| --- | --- |
| Salesperson SELF | Melihat dan mengelola kontak/lead/opportunity yang ditugaskan. Resolve objek dan unit dari DB, termasuk pencarian, Node dan aggregate forecast. |
| Supervisor ORGANIZATION_UNIT | Menugaskan salesperson yang memenuhi syarat dan melihat sales team dalam unit yang diizinkan. Assignment lintas unit memerlukan izin pada kedua sisi. |
| Data steward / privacy | Memeriksa kandidat duplikat, link identitas, merge, redaction dan lift suppression berbukti. Izin PII, consent dan lookup tidak otomatis diberikan pada semua sales. |
| Agent inbox | Membaca conversation di inbox/assignment yang diotorisasi dan mengirim satu pesan setelah verifikasi purpose. Pemilik akses ke pesan dan lampiran dapat berbeda. |
| Approver | MFA sesuai sensitivitas operasi, recheck revisi dan subjek pada commit. Approver target, credit correction dan template berbeda dari pembuat/pemohon. |
| Workload | Principal sistem non-interaktif, identitas mTLS, allowlist method dan issuer per event. Jalur internal hanya menyerahkan fakta dalam lingkup yang didelegasikan. |

Petakan `INVALID_ARGUMENT`, `UNAUTHENTICATED`, `PERMISSION_DENIED`, `NOT_FOUND` untuk objek di luar visibilitas sesuai policy, `ABORTED`, `ALREADY_EXISTS`, `FAILED_PRECONDITION`, `RESOURCE_EXHAUSTED`, `UNAVAILABLE` dan `DEADLINE_EXCEEDED` melalui satu error mapper. Error message stabil dalam Bahasa Inggris. Detail SQL, nomor kontak, isi pesan dan path storage disimpan di luar error publik.

### Gateway dan UX kontrak

- file: [schema.graphql](../../../backend/gateway/graphql/schema.graphql), [ef_party.graphql](../../../backend/gateway/graphql/enterprise_foundation/ef_party.graphql), [iam_user.graphql](../../../backend/gateway/graphql/iam/iam_user.graphql), [registry.go](../../../backend/gateway/internal/source_connections/registry.go), [schema.resolvers.go](../../../backend/gateway/graph/resolvers/schema.resolvers.go) dan file CREATE di atas.
- function: Adapter method per field, mapper ID, Relay connection, Query.node dan resolver tipis ke adapter. Pakai SDL `extend type Query` / `extend type Mutation`, camelCase untuk root dan snake_case untuk field/argumen sesuai contoh aktual. Input object untuk perintah kompleks, argumen langsung bila sederhana, payload mengimplementasi MutationResult.
- logic: Terapkan session, CSRF, allowlist dan permission. Parse ID bertipe sebelum gRPC, bangun label UI dari nama/nomor, mask endpoint. Kembalikan edges, nodes, pageInfo, total_count dan cursor stabil sesuai filter/scope. Node detail, nested dan lookup menerapkan kebijakan yang sama dengan root.
- error_handling: Query memakai GraphQL error terpetakan, mutation memakai payload errors/was_successful sesuai pola Gateway. Field owner lambat memakai deadline dan status parsial yang eksplisit. Pertahankan akses ke domain lain ketika CRM unavailable.
- security: Jangan menambahkan route HTTP publik atau RPC workload ke SDL. Seluruh API aplikasi melalui `/api/graphql`. Hindari IDOR, N+1 serta PII dalam error, log, cache lintas user atau audit metadata.
- expected_result: Dua replika Gateway dapat melayani operasi terscope. UI menampilkan nomor/nama/status, ID opaque dipakai hanya sebagai referensi. Halaman fitur yang menunggu bukti tetap berstatus belum tersedia.
- testing: HTTPS pada detail, list, Node, nested field, cursor, count, multi-unit, objek salah tipe, payload error dan respons ketika service privat offline.

## Action

1. **Mulai dari kontak.** Siapkan bootstrap, katalog minimal, TLS, migration dev/test dan irisan create/get/list kontak. Buktikan login, hak lintas unit dan pagination hingga ke SQL. Pertahankan data dan mapping jika instalasi lama ditemukan.
2. **Kerjakan use case internal.** Tambah consent, lead, pipeline, follow-up dan template. Gunakan decimal serta waktu UTC, tanggal bisnis menurut timezone setting. Catat versi aggregate atau histori untuk CAS/as_of. Tambah tabel minimum untuk bukti baris sumber yang sudah teralokasi, approval koreksi kredit dan receipt handoff bila 35 tabel belum mewakili fakta tersebut. Tambahan DBML dan SQL wajib menjelaskan natural key, constraint, retensi dan reprocessing.
3. **Sambungkan owner dan delivery.** Ikuti [instruksi wiring lintas management](../../../agents/instructions/customer_relationship_sales/2-leader-wiring-lintas-management-instructions.md). Prioritaskan quote B2B, confirmed order dan target. Lanjutkan provider/DRM, Customer 360 dan CS sesuai ketersediaan kontrak. Pisahkan saldo otoritatif owner dari proyeksi CRM. Tandai status tiap operasi sebagai dirancang, gRPC teruji, Gateway terhubung dan HTTPS terverifikasi.
4. **Verifikasi dan serahkan.** Dari folder contract service jalankan `go generate ./...`. Dari root module jalankan `go fmt ./...`, `go vet ./...`, `go test -race -count=1 ./services/...` dan `go build -o /tmp/customer-relationship-sales-verify ./cmd`. Unit test gRPC hanya di layer services. Dari `backend/gateway/` jalankan `go generate ./...`, `go vet ./...` dan tes adapter/resolver terdampak. Dari root jalankan `docker compose config --quiet`, build/start service & kedua Gateway, lalu HTTPS `https://erp.localhost:8443/api/graphql`. Login superadmin, buat role/permission/scope dan assignment ke superadmin melalui GraphQL IAM. Gunakan juga salesperson A/B dan approver terpisah. Kumpulkan bukti E2E tanpa secret/PII.
5. **Kriteria selesai.** Laporan per kelompok memuat tabel DB, file, RPC, field SDL, izin/scope, aktor, positive/negative/replay/konkurensi, hasil HTTP dan efek persisten. Sertakan jumlah aktual/target, kasus gagal, status owner, SHA dan lokasi E2E di `agents/e2e/customer_relationship_sales/`. Sebut operasi yang menunggu owner secara eksplisit. Perbarui context, history dan katalog. Akhiri instruksi dengan `-done.md` saat seluruh acceptance terpenuhi.

## Context

**Urutan file yang wajib dibaca agent.** Path ini merujuk checkout bersama. Dua sumber utama, logic dan DBML, belum tercatat di Git pada baseline. Untuk clone lain, ambil versi kedua sumber dari owner sebelum menghasilkan SQL. Referensi implementasi bertanda contoh menunjukkan pola yang perlu diperiksa kembali pada saat mulai.

- **01 · Keputusan dan batas owner.** [AGENTS.md](../../../AGENTS.md) · [agents/arsitektur.md](../../../agents/arsitektur.md) · [agents/boundary.md](../../../agents/boundary.md) · [agents/users.md](../../../agents/users.md) · [docs/README.md](../../../docs/README.md) · [docs/03-managements/crm-and-customer-service.md](../../../docs/03-managements/crm-and-customer-service.md) · [docs/03-managements/oms-and-commerce.md](../../../docs/03-managements/oms-and-commerce.md) · [agents/context/crm_sales_management/1-context.md](../../../agents/context/crm_sales_management/1-context.md) · [agents/context/customer_service_management/1-context.md](../../../agents/context/customer_service_management/1-context.md).
- **02 · Usulan produk dan data.** [agents/logic/customer_relationship_sales/logic.md](../../../agents/logic/customer_relationship_sales/logic.md) · [backend/services/customer_relationship_sales/db/schema/customer_relationship_management.dbml](../../../backend/services/customer_relationship_sales/db/schema/customer_relationship_management.dbml).
- **03 · Cara coding dan format handoff.** [agents/how_to_code_grpc_managements.md](../../../agents/how_to_code_grpc_managements.md) · [agents/how_to_code_wiring_grpc_managements_to_gateway.md](../../../agents/how_to_code_wiring_grpc_managements_to_gateway.md) · [agents/instruction_template.md](../../../agents/instruction_template.md).
- **04 · Implementasi vertikal EF sebagai contoh.** [backend/services/enterprise_foundation/db/migrate.go](../../../backend/services/enterprise_foundation/db/migrate.go) · [backend/services/enterprise_foundation/repository/pgstore/tx.go](../../../backend/services/enterprise_foundation/repository/pgstore/tx.go) · [backend/services/enterprise_foundation/services/party/service.go](../../../backend/services/enterprise_foundation/services/party/service.go) · [backend/services/enterprise_foundation/transport/grpc/handler/register.go](../../../backend/services/enterprise_foundation/transport/grpc/handler/register.go) · [backend/services/enterprise_foundation/transport/grpc/server.go](../../../backend/services/enterprise_foundation/transport/grpc/server.go) · [backend/services/enterprise_foundation/cmd/wiring.go](../../../backend/services/enterprise_foundation/cmd/wiring.go) · [backend/services/enterprise_foundation/contract/generate.go](../../../backend/services/enterprise_foundation/contract/generate.go).
- **05 · SDL dan wiring Gateway.** [backend/gateway/graphql/schema.graphql](../../../backend/gateway/graphql/schema.graphql) · [backend/gateway/graphql/iam/iam_user.graphql](../../../backend/gateway/graphql/iam/iam_user.graphql) · [backend/gateway/graphql/enterprise_foundation/ef_party.graphql](../../../backend/gateway/graphql/enterprise_foundation/ef_party.graphql) · [backend/gateway/gqlgen.yml](../../../backend/gateway/gqlgen.yml) · [backend/gateway/generate.go](../../../backend/gateway/generate.go) · [backend/gateway/graph/resolvers/schema.resolvers.go](../../../backend/gateway/graph/resolvers/schema.resolvers.go) · [backend/gateway/internal/source_connections/registry.go](../../../backend/gateway/internal/source_connections/registry.go) · [backend/gateway/internal/source_connections/web_commerce_client.go](../../../backend/gateway/internal/source_connections/web_commerce_client.go) · [backend/gateway/cmd/main.go](../../../backend/gateway/cmd/main.go).
- **06 · Workspace, deployment dan IAM.** [backend/go.work](../../../backend/go.work) · [docker-compose.yml](../../../docker-compose.yml) · [backend/shared/rbac/catalogs/web_commerce.json](../../../backend/shared/rbac/catalogs/web_commerce.json) · [backend/services/iam/contract/proto/iam/v1/session.proto](../../../backend/services/iam/contract/proto/iam/v1/session.proto).

Pekerjaan lintas domain dirinci pada [instruksi kedua](../../../agents/instructions/customer_relationship_sales/2-leader-wiring-lintas-management-instructions.md). Semua nama operasi, kebutuhan data UI dan RBAC per field ada pada [katalog gRPC/GraphQL](../../../agents/grpc/customer_relationship_sales/1-katalog-operasi-relasi-pelanggan-dan-penjualan.md).
