# Panduan Gemini untuk ERP Core & Schema

 codex resume 01a1e0cd-bfc6-7f60-8760-337fe8a9c501

Panduan operasional dan onboarding sesi bagi Gemini dalam repositori ERP (desain skema DBML dan backend Go microservices).

## 1. Protokol Masuk Sesi (Session Onboarding)

**Perspektif Produk pada Fase Development (WAJIB untuk seluruh agen, khususnya Frontend, UI/UX, IAM & Gateway):** Sebelum merancang atau mengubah antarmuka pengguna, halaman `/login`, navigasi, mekanisme *post-login redirect*, maupun *copywriting*, **wajib membaca [docs/development-product-perspektif.md](docs/development-product-perspektif.md)**. Produk ini adalah codebase aplikasi terpadu (**Web Commerce + Retail ERP**) dengan model satu instalasi per perusahaan (*single-tenant*), **bukan** platform SaaS multitenant yang menjual langganan ERP. Selama fase *development*, sistem diposisikan seolah-olah perusahaan/brand ritel bernama **IntegriOne** sedang memakai aplikasi ini. Halaman `/login` adalah **satu pintu masuk umum (*unified login*)** untuk **calon pembeli Web Commerce, internal (employee), serta eksternal mitra (vendor/supplier)** — dilarang menggunakan redaksi eksklusif internal/admin ERP pada halaman login umum. **Fokus pengembangan saat ini adalah IAM Management dan Dashboard IAM (`/identity`)** beserta pemetaan redirect setelah autentikasi berhasil sesuai hak akses dari server IAM.

**Konsep identitas dan akses lintas management (wajib):** Sebelum mengimplementasikan user, login, provisioning, group, role, permission, scope atau pemeriksaan ABAC di management apa pun, baca [agents/users.md](agents/users.md). Dokumen ini menetapkan satu objek `users` IAM (`customer`, `employee`, `vendor`, `supplier`, `system`) tanpa kolom JSONB `attributes`, autentikasi `/api/graphql` bersesi opaque, evaluasi RBAC–ABAC berbasis `AccessSnapshot` & `ScopeType`, sumber atribut bisnis di service pemilik, serta batas kepemilikan HRM, Procurement, Enterprise Foundation dan Commerce. Role/permission bisnis hanya boleh diaktifkan sesuai katalog yang disetujui pemilik management; jangan menyimpan data bisnis/grant di `users` atau menulis tabel IAM langsung dari management lain.

Sebelum menjalankan tugas atau setelah context compaction, wajib membaca file berikut:

**Fokus pengembangan Retail ERP sesuai owner (12 November 2026):** untuk pekerjaan pada **13 management terpilih** (IAM, Enterprise Foundation, PLM, Commerce, Pricing, OMS, WMS, Finance, Procurement, Customer Service, CRM, DRM, HRM) atau integration Shopee/TikTok Shop, webhook, analytics, website, developer **wajib membaca** [indeks program Retail ERP](docs/README.md) dan dokumen kebutuhan/arsitektur/tes **yang relevan dengan domain aktif** pada struktur `docs/` baru (tidak wajib membaca seluruh halaman setiap sesi). Urutan pengembangan: **internal → vendor/supplier → integrasi marketplace/webhook/analytics → customer dan calon pembeli terakhir**. Pengembangan baru Manufacturing Operations **PAUSED**; kode dan implementasi historis tidak dihapus. Semua keputusan keamanan owner D01–D12, boundary serta larangan branch baru tetap berlaku. Draf legal/API/anggaran yang belum disetujui tidak boleh diperlakukan sebagai instruksi implementasi otomatis.

**KEPUTUSAN OWNER 13 November 2026 — API GraphQL tunggal:** Satu-satunya endpoint
HTTP API aplikasi yang boleh dipublikasikan adalah `/api/graphql`, termasuk
login, CSRF, session, logout dan seluruh domain. Jangan membuat route REST
`/api/auth/*`, `/healthz`, atau endpoint API tambahan. Internal gRPC
tetap privat. Kebijakan D01–D12 tidak berubah; migrasi API historis harus
mengutamakan tes aman dan tidak membypass IAM/CSRF. Lihat
[keputusan produk](docs/00-governance/decisions.md) dan
[instruksi migrasi Gateway](agents/instructions/gateway/1-owner-gateway-resolver-graphql-autentikasi-instructions.md).

**WAJIB untuk seluruh agen/developer management (tanpa pengecualian):** Baca [agents/arsitektur.md](agents/arsitektur.md) **sebelum audit, perancangan, penulisan/perubahan kode, database, GraphQL/proto, integrasi, maupun pengujian**; ulangi saat sesi baru, setelah compaction, atau ketika melanjutkan pekerjaan lama. Pahami keputusan owner **D01–D12**, status implementasi nyata versus desain, trust boundary IAM/Gateway, pola DDD lintas management, serta syarat E2E. Dokumen arsitektur melengkapi [agents/boundary.md](agents/boundary.md), bukan mengizinkan akses/modifikasi management lain. Bila task bertentangan dengan keputusan arsitektur, catat dan minta keputusan leader/owner; jangan membuat bypass.

1. [AGENTS.md](AGENTS.md) — pedoman arsitektur repositori, aturan handoff, dan batas operasional.

   - **Wajib baca [agents/boundary.md](agents/boundary.md)** sebelum audit, desain, perubahan kode/skema, atau integrasi. Baca aturan umum dan bagian management yang dikerjakan beserta pemilik dependency yang sudah tercantum. Boundary memuat 13 management dalam scope dan membedakan service bisnis yang telah diperiksa dari scaffold/desain; keberadaan go.mod, DBML, atau bootstrap tidak cukup sebagai bukti implementasi. Patuhi kepemilikan data dan batas perubahan; penambahan management di luar 13 memerlukan keputusan owner.

2. File konteks terbaru untuk setiap management yang dikerjakan di `agents/context/{management}/{id}-context.md` (pilih ID numerik tertinggi **dalam folder management tersebut**). Bila tugas melibatkan beberapa management, baca konteks terbaru masing-masing. Jika belum ada, lanjutkan tugas dan buat `1-context.md` saat perubahan konteks memang diperlukan. Jangan membuat `context.md` di root repositori atau `agents/`.

3. Spesifikasi domain dan DBML di [backend/services/](backend/services/) (per management pada `db/schema/`), serta workspace di [backend/](backend/). Direktori lama `database_module/` dan ejaan `back-end/` bukan path aktif; gunakan struktur `backend/` yang benar.

4. **Wajib Gunakan `codebase-memory-mcp` Sebelum Mengubah Kode**: Sebelum memulai penulisan atau merubah kode apapun, wajib menggunakan `codebase-memory-mcp` terlebih dahulu (seperti `search_graph`, `trace_path`, atau `get_code_snippet`) untuk mengecek struktur, fungsi, dan dependensi terkait agar alur perubahan jelas, tepat sasaran, dan hemat context window / token.

5. **Penanganan Otomatis Saat Pengguna Berganti Device**: Jika pengguna menginformasikan berpindah/berganti device, agen **DILARANG BERPANJANG LEBAR ATAU BERTANYA BINGUNG**, melainkan **WAJIB LANGSUNG SECARA OTOMATIS** menjalankan langkah pemulihan codebase memory & environment sesuai Seksi 6 di bawah.

## 2. Bahasa Komunikasi & Catatan

- **Aturan Bahasa (Wajib):**

  - **Komunikasi & Catatan Agen:** seluruh interaksi dengan pengguna, ringkasan, catatan konteks sesi, *working history*, audit, benchmark, instruksi lintas agen, dan laporan E2E wajib menggunakan **Bahasa Indonesia**.

  - **Komentar Kode (*Code Comments*):** seluruh komentar di dalam kode sumber (doc comment package/fungsi/tipe, inline comment, dan komentar blok) wajib menggunakan **Bahasa Indonesia**, kecuali direktif compiler/tool standar (seperti `//go:embed`, `//go:generate`, atau anotasi linter) yang wajib mengikuti sintaks aslinya.

  - **Error Code & Pesan Error:** seluruh *error code* (domain, gRPC, GraphQL, HTTP, dan konstanta error) serta pesan error di dalam kode wajib menggunakan **Bahasa Inggris**.

  - **Implementasi Kode:** seluruh implementasi kode (nama package, file, struct, interface, fungsi/method, variabel, konstanta, enum, identifier skema/DBML/SQL, kontrak API/proto/GraphQL, dan log teknis) wajib menggunakan **Bahasa Inggris**.

- Gunakan tautan Markdown relatif repository yang valid untuk dokumen/source di GitHub. Saat bekerja pada checkout lokal, tautan `file://` boleh dipakai jika menunjuk path lokal aktual; jangan menaruh path absolut perangkat tertentu sebagai sumber kanonik.

## 3. Struktur Repositori & Perintah Kerja

- **`backend/services/{service_name}/db/schema/`**: DBML per owner; tree 13 November 2026 memuat **9 DBML service** di antara 13 management (termasuk Commerce dan DRM). Tidak ada direktori `database_module/` di root; jangan menganggap skema lama di audit historis aktif.

  - Inspeksi: `rg --files backend/services -g '*.dbml'`, `rg -n '^Table ' backend/services -g '*.dbml'`, `rg -n '^Ref:' backend/services -g '*.dbml'`.

- **`backend/`**: Workspace Go (Go 1.27+, toolchain proyek `go1.27.4`, gRPC, gqlgen, Fiber v3.7.0). Seluruh infrastruktur dan container aplikasi local development dijalankan melalui Docker Compose di root repositori (`docker-compose.yml`) dengan baseline terbaru 2026: base OS aplikasi **Ubuntu 26.10** (`ubuntu:26.10`), **Go 1.27.4** (`gateway-1`/`gateway-2` pada `127.0.0.1:5000-5001`), **Node.js 26.x** (`v26.13.1` + `libatomic1` pada `frontend-1`/`frontend-2` `127.0.0.1:3333-3334`), **Caddy** `caddy:latest` (`v2.11.9` pada `https://erp.localhost:8443`), **PostgreSQL** `postgres:latest` (**PostgreSQL 18.7** pada `127.0.0.1:6543`), dan **Redis** `redis:latest` (**Redis 8.10.3** TLS + ACL pada `127.0.0.1:6379`). Versi aktual wajib diperiksa setelah setiap pull/build.

  - Alur build/generate/run: `make -C backend bootstrap`, `make -C backend proto`, `make -C backend gql`, `make -C backend tidy`, `make -C backend run-gateway`, `make -C backend run-<service_name>` (misal: `run-iam_management`, `run-enterprise_foundation`, `run-human_resource_management`), serta orkestrasi Docker: `docker compose build`, `docker compose up -d --wait`, dan `docker compose ps`.

  - **Panduan wajib Backend Engineers (Skill Golang & Superpowers):** Baca [`## Aturan Khusus Backend Engineers: Skill Golang & Superpowers`](#aturan-khusus-backend-engineers-skill-golang--superpowers) sebelum melakukan coding, perancangan, review, debugging, maupun testing backend Go.

  - Verifikasi Go: jalankan `go fmt ./...`, `go vet ./...`, dan unit test pada modul yang dimodifikasi.

- **`frontend/`**: Aplikasi web **Next.js `16.4.6`** (App Router di [`frontend/src/app/`](frontend/src/app/)), **React `19.3.0`**, **TypeScript `^5`** (`strict: true`), **Tailwind CSS `4.3.3`** (`@tailwindcss/turbopack`), CSS Modules (`*.module.css`), animasi **GSAP (`gsap` `^3.15.1` + `@gsap/react` `^2.1.2`)**, lockfile [`frontend/package-lock.json`](frontend/package-lock.json) (`npm ci`), serta `reactCompiler: true`, `cacheComponents: true`, dan `partialPrefetching: true` di [`frontend/next.config.ts`](frontend/next.config.ts). Berjalan di container Node.js `26.x` (`frontend-1` `127.0.0.1:3333:3000`, `frontend-2` `127.0.0.1:3334:3000`) di belakang Caddy (`https://erp.localhost:8443`).
  - **Panduan wajib Frontend Engineers:** Baca [`frontend/AGENTS.md`](frontend/AGENTS.md) beserta [perspektif produk](docs/development-product-perspektif.md) dan [`## Aturan Khusus Frontend Engineers`](#aturan-khusus-frontend-engineers) sebelum mengubah UI, komponen, CSS, animasi GSAP, atau integrasi `/api/graphql`.
  - Verifikasi Frontend: `npx next typegen && npm run lint && npx tsc --noEmit` (plus `npm run build`), serta `docker compose build frontend-1` & `docker compose up -d --wait frontend-1 frontend-2 caddy` + uji HTTPS nyata pada `https://erp.localhost:8443`.

- **Larangan**: Dilarang menyimpan password/token/kredensial terbuka, dilarang menginstal atau menyalakan service PostgreSQL/Redis langsung di OS host Windows (wajib gunakan container Docker), dilarang menjalankan live database tanpa konfigurasi resmi, dan hindari duplikasi master data lintas modul.

### Membaca Dokumentasi API Package Go

- Jika implementasi atau audit membutuhkan penjelasan API suatu package Go, **wajib gunakan `go doc ...` melalui terminal** dari direktori module yang terkait agar dokumentasi mengikuti dependency yang digunakan proyek.

- Mulai dari simbol yang diperlukan: `go doc <import-path>` untuk ringkasan package, `go doc <import-path> <Symbol>` untuk tipe/function, atau `go doc <import-path> <Type.Method>` untuk method. Contoh: `go doc context WithCancel` dan `go doc gorm.io/gorm DB.Transaction`.

- Periksa signature, parameter, nilai kembali, penanganan error, dan batas perilaku yang didokumentasikan sebelum memakai API; jangan menebak dari nama atau ingatan. Jika dokumentasi belum cukup, gunakan `go doc -src <import-path> <Symbol>` atau baca source package pada versi dependency proyek. Laporkan jika dokumentasi/package tidak tersedia; jangan menganggap API telah terverifikasi.

### Penggunaan Wajib Skill Golang & Superpowers (Backend)

Before any Go coding, review, debugging, troubleshooting, or setup task, load the `samber/cc-skills-golang@golang-how-to` skill first — it routes to whichever other Go skills the task needs.

- **Wajib Skill Golang Sesuai Konteks Tugas:** Saat mengerjakan backend Go (`backend/services/`, `backend/gateway/`, `backend/contracts/`, atau skrip/tooling Go), agen wajib membaca `SKILL.md` dari skill Golang yang sesuai sebelum menulis atau meninjau kode (misal `golang-database` + `golang-error-handling` untuk DB/migrasi, `golang-grpc` + `grpc-golang` untuk kontrak/transport gRPC, `golang-graphql` + `gqlgen-go` untuk GraphQL Gateway, `golang-testing` + `golang-safety` + `golang-security` untuk implementasi/pengujian, dan `go-reviewer`/`go-build-resolver` untuk review/perbaikan build). Daftar lengkap & matriks pemetaan tersedia di [`## Aturan Khusus Backend Engineers: Skill Golang & Superpowers`](#aturan-khusus-backend-engineers-skill-golang--superpowers).
- **Gunakan Superpowers Skills Jika Diperlukan:** Untuk perancangan fitur baru (`brainstorming`), perencanaan bertahap (`writing-plans`, `executing-plans`), implementasi berbasis tes (`test-driven-development`), investigasi bug (`systematic-debugging`), review (`requesting-code-review`, `receiving-code-review`), eksekusi paralel (`dispatching-parallel-agents`, `subagent-driven-development`), maupun verifikasi akhir (`verification-before-completion`), muat dan ikuti skill di `plugins/superpowers/skills/` / `skills/`.
- **Hirarki Aturan Normatif:** Aturan khusus repositori di `AGENTS.md`, `agents/boundary.md`, `agents/arsitektur.md`, dan `agents/users.md` selalu mengungguli saran generik di dalam skill (termasuk aturan `main`-only Git, penamaan repository `Select`/`Insert`/`Create`/`Update`/`Delete`, GORM, `slog`, bahasa komentar/kode/error, dan baseline migrasi `000001`).

## 4. Konvensi Pembaruan Memori & Handover

Setelah setiap tugas **Backend** selesai dan sebelum memberikan respons akhir ke pengguna (untuk **Frontend**, seluruh berkas handover/instruksi dipisahkan ke `agents/frontend/` sesuai [`frontend/AGENTS.md`](frontend/AGENTS.md)):

1. **Perbarui Konteks Sesi**: Simpan konteks aktif menurut management di `agents/context/{management}/{id}-context.md`, misalnya `agents/context/finance_accounting/1-context.md`. Nama management mengikuti direktori modul dan memakai `snake_case`; ID numerik dimulai dari `1` dan berurutan **per folder management**. Jika belum ada konteks untuk management tersebut, buat `1-context.md`.

   - **Batas Panjang**: Maksimal **1000 baris** per berkas konteks, tanpa batas jumlah kata. Jika berkas terbaru sudah 1000 baris atau penambahan akan melampaui 1000 baris, lanjutkan pada `agents/context/{management}/{next-id}-context.md`. Konteks sebelumnya boleh dihapus setelah informasi pentingnya tercakup dalam konteks aktif.

   - **Struktur Wajib**: Setiap berkas konteks harus memuat tiga heading Markdown tingkat dua, berurutan: `## Apa yang sudah dikerjakan`, `## Apa yang belum dikerjakan`, dan `## Apa yang harus dikerjakan selanjutanya`. Isi tiap bagian secara faktual untuk management tersebut; jika belum ada isi, tulis `Belum ada.` tanpa menghapus heading.

   - **Prinsip Konten**: **Hanya menyimpan hal-hal penting dan berdampak tinggi** (arsitektur inti, batasan modul, keputusan desain, dan handoff status aktif) agar mudah dan jelas dibaca oleh LLM tanpa noise.

   - **Kapan Diperbarui**: Diperbarui **HANYA jika pekerjaan memang membutuhkan perubahan konteks** (perubahan arsitektur, penetapan keputusan baru, penyelesaian modul/milestone besar). **Pekerjaan remeh (minor bugfix, typo, formatting kode, sinkronisasi minor) TIDAK PERLU memperbarui context**.

2. **Buat Working History**: Tambahkan file baru `agents/working_history/{management}/{id}-{date-time}-{title}.md`. Nama management mengikuti direktori modul dan memakai `snake_case`; ID numerik dimulai dari `1` dan berurutan dari ID tertinggi **dalam folder management tersebut** (format `date-time`: `YYYY-MM-DD-HHmmss` WIB, maksimal 150 kata).

   - **Pekerjaan Remeh**: Sekiranya pekerjaan tergolong remeh (typo, formatting, sinkronisasi kecil), **TIDAK PERLU menambah working history**.

3. **Simpan Audit per Management**: Setiap laporan audit baru disimpan di `agents/audit/{management}/{id}-{date-time}-{title}.md`. Nama management memakai `snake_case` sesuai direktori modul. ID numerik dimulai dari `1` dan berurutan dari ID tertinggi **dalam folder management tersebut**; `date-time` memakai format `YYYY-MM-DD-HHmmss` WIB dan judul memakai lowercase kebab-case.

4. **Simpan Laporan Benchmark Service**: Setiap eksekusi benchmark service disimpan di `agents/benchmark/{management}/{id}-{benchmark-title}.md`. `{management}` memakai `snake_case` sesuai direktori modul; `{id}` dimulai dari `1` berurutan **dalam folder management tersebut**; `{benchmark-title}` memakai lowercase kebab-case yang spesifik terhadap domain/operasi yang diukur.

5. **Simpan Laporan E2E (Berbagai Cases Query/Mutation)**: Setiap laporan hasil pengujian E2E disimpan di `agents/e2e/{management}/{id}-{e2e-title-{Query***/Mutation***}}.md` (dalam nama berkas aktual gunakan lowercase kebab-case yang menyertakan jenis serta nama operasi, misal `agents/e2e/iam_management/1-authentication-mutation-user-login.md` atau `2-user-query-search-and-list.md`). `{management}` memakai `snake_case`, `{id}` dimulai dari `1` berurutan **dalam folder management tersebut**, dan isi laporan wajib merinci seluruh matriks *cases* yang diuji.

6. **Pembaruan Codebase Memory (`codebase-memory-mcp`)**: Setelah mengubah kode atau menyelesaikan suatu goal/task, seluruh agen wajib menjalankan pembaruan/re-index `codebase-memory-mcp` (misal via `codebase-memory-mcp cli index_repository --mode full`) agar graph pengetahuan kode selalu sinkron dengan kondisi codebase terkini.

7. **Git Releases Pasca-Versioning**: Setelah memperbarui versi pada [`backend/VERSION`](backend/VERSION) dan [`backend/CHANGELOG.md`](backend/CHANGELOG.md), wajib membuat Git Release melalui tagging rilis annotated (`git tag -a vX.Y.Z -m "vX.Y.Z"`) dan push ke remote agar rilis terpublikasi resmi.

## 5. Aturan Git Commit & Skema Versioning Hierarkis

### Semua developer management bekerja langsung di `main`

Setiap developer dan agen di seluruh management ERP wajib mengerjakan perubahan dan membuat commit langsung di branch `main`, lalu melakukan push ke `origin/main`. Dilarang membuat branch baru, baik lokal maupun remote, untuk pekerjaan management. Owner meninjau perubahan seluruh management melalui riwayat commit di `main`.

Sinkronkan `main` dengan `origin/main` sebelum mulai bekerja. Jika push ditolak karena developer lain sudah mengirim commit, ambil perubahan terbaru, lakukan rebase atas commit lokal, selesaikan konflik jika ada, lalu jalankan kembali pemeriksaan yang relevan sebelum push. Dilarang melakukan force-push ke `main`.

### Prosedur GitHub/Git wajib untuk seluruh agen dan developer

**1. Sebelum mengubah file: cek status dan sinkronkan.** Pastikan direktori adalah checkout repository ERP yang benar, remote `origin` mengarah ke repository owner, dan branch aktif adalah `main`. Jalankan `git status --short --branch`, `git remote -v`, dan `git fetch origin`. Baca perubahan terbaru, termasuk diff untuk file bersama yang akan disentuh. Jika working tree bersih dan tidak ada commit lokal yang belum dikirim, jalankan `git pull --ff-only origin main`. Jangan menjalankan `git pull` tanpa mode yang jelas karena dapat membentuk merge commit tak disengaja. Jangan otomatis menghapus, me-reset, men-stash, atau menimpa perubahan lokal developer lain. Jika working tree kotor, identifikasi pemilik perubahan dan selesaikan/koordinasikan lebih dulu sebelum sinkronisasi.

**2. Selama bekerja: isolasi scope dan koordinasikan file bersama.** Ubah hanya file dalam management/domain yang ditugaskan menurut `agents/boundary.md`, termasuk wiring domain management tersebut ke `backend/gateway/` sesuai partisi Zona 1–3 pada bagian [`Aturan Tanggung Jawab Wiring Management ke Gateway`](#aturan-tanggung-jawab-wiring-management-ke-gateway). File bersama seperti `AGENTS.md`, `backend/VERSION`, `backend/CHANGELOG.md`, `backend/go.work`, `backend/Makefile`, `backend/contracts/gen/`, serta file Zona 2 Gateway (`backend/gateway/gqlgen.yml`, `backend/gateway/generate.go`, file inti bersama `backend/gateway/graph/generated/root_.generated.go` & `schema.generated.go`, `backend/gateway/graph/model/schema_models_gen.go`, `backend/gateway/graph/resolvers/resolver.go`, `backend/gateway/internal/source_connections/registry.go`, `backend/gateway/cmd/main.go`, `backend/gateway/configs/gateway.yaml`) harus dibaca ulang dan dicek status Git-nya sebelum diedit serta hanya boleh diubah secara *minimal-diff* (*additive-only*) khusus wiring domain sendiri (atau dikoordinasikan dengan pemilik jika menyangkut struktur inti). Jangan mengambil hasil kerja orang lain, jangan menyentuh Zona 1 milik management lain maupun Zona 3 keamanan Gateway, jangan menyentuh database atau migrasi management lain, dan jangan mengedit file generated secara manual; jalankan generator resmi dan tinjau diff-nya.

**3. Sebelum commit: tinjau semua perubahan dan jalankan verifikasi.** Gunakan `git diff --check`, `git status --short`, `git diff`, dan `git diff --staged` untuk memeriksa file mana yang berubah. Tambahkan file dengan path spesifik (`git add <path>`); jangan memakai `git add -A` atau `git add .` secara membabi buta saat ada pekerjaan agen lain. Jangan commit `.env`, password, token, file privat, berkas sementara, atau hasil build yang tidak diminta. Jalankan formatter, generator, lint, build, unit test, dan E2E yang relevan menurut instruksi task; laporkan pemeriksaan yang belum dijalankan secara jujur. Buat commit kecil per domain/tujuan dengan pesan `type(management/domain): ringkasan` dan isi yang menjelaskan logika, cakupan, serta hasil verifikasi. Jangan mencampur perubahan tak terkait.

**4. Sebelum push: integrasikan commit baru developer lain tanpa menulis ulang history remote.** Jalankan `git fetch origin` dan periksa apakah `origin/main` maju. Bila lokal hanya tertinggal dan tidak ada commit sendiri, gunakan `git pull --ff-only origin main`. Bila ada commit **lokal yang belum pernah dipush**, gunakan `git rebase origin/main` setelah working tree bersih; selesaikan konflik dengan memeriksa kedua perubahan dan menjalankan ulang pengujian yang terdampak. Jangan melakukan rebase/reset terhadap commit yang telah dipublikasikan di `origin/main`. Jika konflik menyangkut business rule atau file bersama milik agen lain, hentikan push dan sepakati perubahannya dengan pemilik, bukan memilih `ours`/`theirs` secara otomatis.

**5. Push dan bukti handoff.** Push hanya dengan `git push origin main` setelah verifikasi; bila remote bergerak lagi, ulangi langkah fetch/rebase commit lokal yang belum dipublikasikan dan verifikasi. Setelah push, konfirmasi SHA commit dan keberadaannya di `origin/main`/`main` GitHub. Laporkan link commit, daftar file/domain yang berubah, hasil tes aktual, blocker, dan catatan `agents/context/` atau `agents/working_history/` bila diwajibkan. Jangan menyatakan pekerjaan sudah dipush hanya berdasarkan commit lokal.

**6. Larangan operasi berisiko dan batas persetujuan owner.** Dilarang `git push --force`, `--force-with-lease`, `git reset --hard`, `git clean -fd`, penghapusan branch/tag/release, overwrite commit, serta perubahan konfigurasi GitHub atau aturan branch tanpa permintaan eksplisit owner. Dilarang membuat branch feature/hotfix dan Pull Request (PR) baru sebagai jalur kerja rutin; review owner dilakukan pada commit `main`. PR lama tidak boleh di-merge, di-close, atau dihapus tanpa instruksi owner. Tag versi dan GitHub Release hanya dibuat sesuai keputusan versioning yang telah disetujui owner. Jika GitHub API/connector dipakai sebagai pengganti CLI, verifikasi HEAD `main` sebelum write, gunakan SHA file/commit yang masih berlaku, lakukan update non-force dan verifikasi ref setelah commit; jangan membuat branch sementara untuk menjalankan pemeriksaan.

**7. Saat ada kesalahan atau dependency gagal.** Jangan berusaha memperbaikinya dengan menghapus riwayat publik atau menyembunyikan test gagal. Untuk perbaikan atas perubahan yang sudah ada di `main`, buat commit koreksi baru yang terpisah setelah mengidentifikasi dampaknya dan menguji ulang. Untuk rollback yang diperlukan, ajukan keputusan kepada owner terlebih dahulu, gunakan mekanisme yang menjaga history (misalnya revert yang ditinjau) dan tidak membatalkan commit management lain. Jika akses GitHub/CI gagal, laporkan status faktual dan SHA terakhir yang diketahui; jangan mengklaim push atau tes berhasil.

### Standar Pesan Commit, Skema Versioning, dan Release

1. **Pesan Commit yang Detail**: Setelah menyelesaikan satu task dan kode `backend` bersih dari error (`go fmt`, `go vet`, pengujian lulus), wajib menyiapkan pesan commit yang detail (mencakup domain/scope, detail fungsional, dan bukti verifikasi).

2. **Skema Versioning Berjenjang**:

   - **Tingkat Patch (`*.*.{version}`)**: Untuk setiap perubahan task pada management/domain apapun sebelum domain selesai penuh (misal `0.0.1`, `0.0.2`).

   - **Tingkat Minor (`*.{version}.0`)**: Ketika **selesai 1 domain penuh** tanpa ada error (misal selesai domain `user` -> `0.1.0`).

   - **Tingkat Major (`{version}.0.0`)**: Ketika **selesai 1 management penuh** (seluruh domain dalam microservice tuntas 100%, misal `iam_management` beres penuh -> `1.0.0`).

   - **Baseline versi saat ini: `0.0.0`**, sesuai `backend/VERSION` dan `backend/CHANGELOG.md` pada 13 November 2026. Rangkaian versi/tag sampai `8.0.0` pada dokumen dan audit lama adalah catatan historis, **bukan** status aplikasi aktif atau klaim pengujian terbaru.

3. **Penerbitan Release**: Setiap pembaruan versi pada [`backend/VERSION`](backend/VERSION) dan [`backend/CHANGELOG.md`](backend/CHANGELOG.md) harus diikuti dengan pembuatan tag annotated (`git tag -a vX.Y.Z -m "vX.Y.Z"`) dan push ke remote agar memicu publikasi GitHub Release.

## 6. Protokol Pergantian Device & Pemulihan Otomatis (Device Migration SOP)

Jika pengguna menyatakan telah berganti device (misal: "aku ganti device", "pindah laptop/pc", "device baru"), seluruh agen **DILARANG berpanjang lebar bertanya atau meminta arahan konfigurasi**. Agen **WAJIB LANGSUNG dan OTOMATIS** mengeksekusi tahapan berikut:

1. **Deteksi Path Direktori & Normalisasi Nama Cache DB**:

   - Dapatkan direktori absolut repositori saat ini (misal via `pwd` atau `Get-Location`).

   - Buat format penamaan cache `codebase-memory-mcp` dengan mengganti pemisah path (`/` atau `:\` / `\`) menjadi tanda minus `-` (tanpa tanda minus di paling depan). Contoh Linux: `/home/user/projects/erp-retail` -> `home-user-projects-erp-retail.db`; contoh Windows: `D:\projects\Enterprise-Resouce-Planning-For-Retail` -> `D-projects-Enterprise-Resouce-Planning-For-Retail.db`.

   - Pastikan direktori cache tujuan sudah ada: `~/.cache/codebase-memory-mcp` (Linux/macOS) atau `$env:LOCALAPPDATA\codebase-memory-mcp\cache` (Windows).

2. **Restore atau Re-index Basis Data Graf Pengetahuan (`codebase-memory-mcp`)**:

   - Jika arsip graf terkompresi tersedia dan cocok dengan path OS, ekstrak dan salin ke direktori cache, atau jalankan langsung `index_repository` pada `repo_path` repositori aktif agar basis data graf langsung terbentuk sesuai path perangkat saat ini.

3. **Verifikasi Kesiapan Graph**:

   - Panggil MCP tool `list_projects` dan `index_status` (atau jalankan `codebase-memory-mcp cli index_status --project <project-path-normalized>`) untuk memastikan graph terbaca sempurna di device baru.

   - Jika terdapat perbedaan berkas, jalankan sinkronisasi / incremental index (`detect_changes` atau `index_repository`).

4. **Pemeriksaan Lingkungan Kerja, Docker & Toolchain**:

   - Pastikan runtime Go (`go version`), Docker Desktop (`docker compose ps`), PostgreSQL Docker (`127.0.0.1:6543`, termasuk akses pembuatan DB `*_test`), dan Redis TLS Docker (`127.0.0.1:6379`) berjalan sehat (`Healthy`).

   - Pastikan berkas secret/cert lokal di `backend/deploy/` (`setup-local-env.sh` dan `generate-dev-certs.sh`) serta `.env` lokal di microservice terkait (misal `backend/services/iam_management/.env`) telah tersedia.

5. **Respons Cepat & Langsung Siap Bekerja**:

   - Berikan laporan singkat (1-2 paragraf) dalam Bahasa Indonesia bahwa codebase memory dan environment telah berhasil dipulihkan secara otomatis pada device baru.

   - Langsung siap menerima atau melanjutkan instruksi pengerjaan berikutnya tanpa jeda onboarding.

## 7. Panduan Lengkap Penggunaan Codebase Memory MCP (`codebase-memory-mcp`)

Seluruh agen wajib menjadikan `codebase-memory-mcp` sebagai alat utama eksplorasi kode sebelum menggunakan `rg`/`grep` atau membaca file secara menyeluruh.

### A. Struktur Satu Project Index di Repositori Ini

Seluruh repositori terindeks sebagai **satu project** `codebase-memory-mcp`; tidak ada pemisahan per folder:

1. **`<project-path-normalized>`** (misal Windows: `D-projects-Enterprise-Resouce-Planning-For-Retail` pada root `D:/projects/Enterprise-Resouce-Planning-For-Retail`, atau Linux: `home-muriarfad-Documents-projects-erp-retail` pada `/home/.../erp-retail`):

   - Berisi seluruh graf pengetahuan kode Go (`backend/services/`, `backend/gateway/`, `backend/contracts/`), frontend, simbol, fungsi, method, struct, interface, resolver GraphQL, dan handler gRPC, serta graf dokumentasi (`agents/`, `docs/`, dan dokumen root: ~85.000+ nodes, ~457.000+ edges).

   - **WAJIB gunakan project `<project-path-normalized>` ini** untuk seluruh pencarian simbol, penelusuran call-chain, dan analisis dampak kode.

   - Watch aktif (`watch: true`, strategi git) pada root repositori sehingga perubahan berkas terpantau otomatis.

### B. Urutan Prioritas & Alur Kerja Wajib (Workflow)

1. **Konfirmasi Project & Status (`list_projects` & `index_status`)**:

   - Di awal sesi atau setelah compaction, panggil `list_projects` dan `index_status` untuk memastikan nama project dan status `ready`.

2. **Cari Simbol (`search_graph`)**:

   - Temukan fungsi, method, struct, interface, route, atau variabel menggunakan filter:

     - `name_pattern`: regex nama simbol (misal `".*CreateUser.*"` atau `"^Select.*"`).

     - `qn_pattern`: regex qualified name (misal `".*service/user\\..*"`).

     - `file_pattern`: filter path file (misal `"*invitation*"`).

     - `label`: tipe node (misal `"Function"`, `"Method"`, `"Class"`, `"Interface"`, `"Route"`).

     - `query` (BM25 kata kunci) atau `semantic_query` (array kata kunci semantik, tidak boleh digabung dengan `query`).

3. **Telusuri Alur Pemanggilan (`trace_path`)**:

   - Lacak relasi fungsi dengan `function_name` dan `direction`:

     - `"inbound"`: siapa saja yang memanggil fungsi ini (callers).

     - `"outbound"`: fungsi/method apa saja yang dipanggil oleh fungsi ini (callees).

     - `"both"`: kedua arah sekaligus.

   - Mode tambahan (`mode`): `"calls"` (default), `"data_flow"` (aliran argumen/parameter), `"cross_service"`, `"topology"`.

   - Gunakan `include_tests: true` bila ingin menyertakan pemanggilan dari file `*_test.go`.

4. **Baca Kode Spesifik (`get_code_snippet` & `get_file_outline`)**:

   - Gunakan `get_code_snippet` dengan `qualified_name` dari hasil `search_graph` dan `source_mode`:

     - `"auto"` (default): menampilkan kode penuh untuk fungsi kecil atau outline untuk container besar (200+ baris).

     - `"full"`: menampilkan kode sumber lengkap (hingga 500 baris per halaman; gunakan `start_line` untuk paginasi).

     - `"outline"`: menampilkan daftar member/method di dalam struct/class.

   - Gunakan `get_file_outline` dengan `file_path` relatif terhadap root project untuk melihat daftar seluruh deklarasi simbol dalam satu file.

5. **Validasi Cakupan Indeks (`check_index_coverage`)**:

   - Setelah kandidat path ditemukan, panggil `check_index_coverage` dengan `paths` (array file relatif) dan/atau `scopes` (array direktori).

   - Jika ada file yang dilaporkan `parse_partial`, `skipped`, `excluded`, `stale`, atau `not_indexed`, **wajib** baca langsung file sumbernya (`read_file` / `rg`) sebelum membuat kesimpulan atau klaim audit.

6. **Analisis Lanjutan (`query_graph`, `detect_changes`, `get_architecture`, `search_code`)**:

   - `query_graph`: jalankan kueri Cypher read-only untuk relasi multi-hop kompleks (cek skema label/edge terlebih dahulu dengan `get_graph_schema` dan `diagnostics: "full"`).

   - `detect_changes`: petakan `git diff` terhadap simbol dan fungsi yang terdampak (`scope: "impact" | "files" | "tests"`).

   - `get_architecture`: ringkasan arsitektur statis (`aspects`: `["packages", "entry_points", "routes", "hotspots", "boundaries", "cycles"]`).

   - `search_code`: pencarian teks/literal yang diperingkat oleh graf (`mode: "compact" | "full" | "files"`).

7. **Sinkronisasi Pasca-Perubahan (`index_repository`)**:

   - Project ini memiliki `watch: true` (strategi git) pada root repositori sehingga perubahan berkas terpantau otomatis; tetap **wajib** menjalankan `index_repository` pada `repo_path: "<abs-path-repositori>"` dengan `mode: "full"` setelah selesai mengubah kode Go/proto/GraphQL dalam jumlah besar atau bila graph dirasa tidak sinkron, lalu menyalin `.db` terbaru ke `codebase_memory/` bila diperlukan untuk pencadangan lintas-device.

### C. Cara Pemanggilan: MCP Tool vs CLI Fallback (Terminal)

Agen dapat memanggil `codebase-memory-mcp` melalui dua jalur:

#### 1. Melalui Native MCP Tool (`call_mcp_tool`)

Gunakan `ServerName: "codebase-memory-mcp"` dan `ToolName` sesuai nama tool:

- **Cek Project**: `ToolName: "list_projects"`, `Arguments: {}`

- **Cari Simbol**: `ToolName: "search_graph"`, `Arguments: {"project": "home-muriarfad-Documents-projects-erp-retail", "name_pattern": ".*CreateAuthenticatedSession.*"}`

- **Lacak Pemanggil**: `ToolName: "trace_path"`, `Arguments: {"project": "home-muriarfad-Documents-projects-erp-retail", "function_name": "CreateAuthenticatedSession", "direction": "both", "depth": 3}`

- **Baca Snippet**: `ToolName: "get_code_snippet"`, `Arguments: {"project": "home-muriarfad-Documents-projects-erp-retail", "qualified_name": "<qn_dari_search_graph>", "source_mode": "full"}`

- **Cek Coverage**: `ToolName: "check_index_coverage"`, `Arguments: {"project": "home-muriarfad-Documents-projects-erp-retail", "paths": ["backend/services/iam_management/internal/service/authentication/service.go"]}`

- **Re-index**: `ToolName: "index_repository"`, `Arguments: {"repo_path": "/home/muriarfad/Documents/projects/erp-retail", "mode": "full"}`

#### 2. Melalui CLI Lokal (`codebase-memory-mcp cli`)

Jika koneksi MCP mengalami kendala (`Transport closed`) atau pada subagent yang tidak mewarisi sesi MCP, **wajib gunakan CLI lokal via terminal** (jangan langsung menyerah atau mengklaim graph tidak bisa diakses).

> **Penting**: Argumen raw JSON pada CLI sudah **deprecated**; gunakan flag `--<nama-parameter-kebab-case>`:

```bash

# 1. Lihat daftar project yang terindeks

codebase-memory-mcp cli list_projects

# 2. Cek status indeks project repositori

codebase-memory-mcp cli index_status --project home-muriarfad-Documents-projects-erp-retail

# 3. Cari simbol berdasarkan regex nama atau file

codebase-memory-mcp cli search_graph \

--project home-muriarfad-Documents-projects-erp-retail \

--name-pattern ".*ChangePassword.*" \

--file-pattern "*password*"

# 4. Lacak caller (inbound) dan callee (outbound)

codebase-memory-mcp cli trace_path \

--project home-muriarfad-Documents-projects-erp-retail \

--function-name "ChangePassword" \

--direction both \

--depth 3

# 5. Baca kode sumber lengkap dari qualified_name

codebase-memory-mcp cli get_code_snippet \

--project home-muriarfad-Documents-projects-erp-retail \

--qualified-name "<qualified_name>" \

--source-mode full

# 6. Lihat outline satu file

codebase-memory-mcp cli get_file_outline \

--project home-muriarfad-Documents-projects-erp-retail \

--file-path "backend/services/iam_management/internal/service/password/service.go"

# 7. Validasi cakupan indeks (coverage)

codebase-memory-mcp cli check_index_coverage \

--project home-muriarfad-Documents-projects-erp-retail \

--paths '["backend/services/iam_management/internal/service/password/service.go"]'

# 8. Jalankan Cypher query pada graph

codebase-memory-mcp cli query_graph \

--project home-muriarfad-Documents-projects-erp-retail \

--query "MATCH (n:Method)-[r:CALLS]->(m:Method) WHERE n.name CONTAINS 'Password' RETURN n.qn, m.qn LIMIT 25"

# 9. Re-index penuh setelah perubahan kode

codebase-memory-mcp cli index_repository \

--repo-path /home/muriarfad/Documents/projects/erp-retail \

--mode full

```

### D. Tingkatan Bukti (Evidence Tiers) & Aturan Subagent

- **Scout (Tier 1)**: Pencarian positif cepat dengan beberapa pemanggilan graph dan pengecekan sumber terarah. Tandai sebagai sementara (**provisional**); dilarang membuat klaim negatif/menyeluruh ("tidak ada fungsi X di seluruh repo").

- **Verify (Tier 2, Default)**: Bukti graph terarah sesuai tugas, `trace_path` pada arah yang relevan, `get_code_snippet` eksak untuk klaim penting, dan `check_index_coverage` pada seluruh path bukti.

- **Auditor (Tier 3)**: Verifikasi penuh pada lingkup terbatas dengan indeks generasi terbaru, paginasi tuntas, penelusuran `inbound` & `outbound`, `check_index_coverage` dengan `scopes`, dan pembacaan langsung seluruh rentang file yang tidak tercakup indeks.

- **Delegasi Subagent**: Sebelum memanggil subagent, jalankan query graph dan coverage di parent agent, lalu sertakan nama project, hasil `qualified_name`, path file, temuan call-chain, dan status coverage di dalam prompt subagent.

### E. Kapan Boleh Fallback ke `rg` / `grep` / `read_file`

- Mencari string literal, pesan error spesifik, query SQL mentah, atau nilai konfigurasi `.env`.

- Memeriksa file non-kode yang tidak membentuk AST simbol (seperti `backend/services/{service_name}/db/schema/*.dbml`, `.gitignore`, `Makefile`, migrasi `.sql`).

- Membaca baris kode yang dilaporkan memiliki gap oleh `check_index_coverage`.

## 8. Aturan Docker Local Development, Single-Build, Build After Development, CI Gates, dan Konkurensi Multi-Agen

### A. Eksekusi Docker & Standar Tech Stack Container (2026 Latest Baseline)

1. **Eksekusi Native Docker CLI (Tanpa Wrapper WSL Manual)**:
   - Di lingkungan Windows dengan Docker Desktop (WSL2 backend `docker-desktop`), seluruh perintah `docker` dan `docker compose` **wajib dieksekusi langsung dari terminal host (`PowerShell`)** di root repositori (`D:\projects\Enterprise-Resouce-Planning-For-Retail`). Dilarang membungkus perintah Docker di dalam `wsl -d Ubuntu`.
2. **Infrastruktur Database & Cache Murni via Docker**:
   - Dilarang menginstal atau menyalakan service PostgreSQL maupun Redis secara langsung di OS host Windows. Seluruh koneksi development dan E2E testing (`*_test`) diarahkan ke container Docker:
     - **PostgreSQL (`postgres:latest` — PostgreSQL 18.7)**: `127.0.0.1:6543 -> 5432/tcp` (volume persisten `deploy_pgdata18`, provisioning role/DB otomatis via `backend/deploy/postgres/init/00-roles.sh`).
     - **Redis (`redis:latest` — Redis 8.10.3)**: `127.0.0.1:6379 -> 6379/tcp` (TLS wajib, AOF + `noeviction`, ACL per komponen di `backend/deploy/secrets/redis-users.acl`).
3. **Standar Versi Runtime Container Aplikasi (2026 Latest)**:
   - Seluruh stage build dan runtime container aplikasi (`gateway`, `frontend`, `caddy`) wajib menggunakan base OS **`ubuntu:26.10`** (bukan Debian/Alpine):
     - **Gateway (`integrione-retail-gateway:dev`)**: **Go `1.27.4`** di atas `ubuntu:26.10` (`gateway-1` di `127.0.0.1:5000`, `gateway-2` di `127.0.0.1:5001`).
     - **Frontend (`integrione-retail-frontend:dev`)**: **Node.js `26.x` (`v26.13.1`)** + pustaka sistem `libatomic1` & `libstdc++6` di atas `ubuntu:26.10` (`frontend-1` di `127.0.0.1:3333`, `frontend-2` di `127.0.0.1:3334`).
     - **Reverse Proxy (`integrione-retail-caddy:dev`)**: **Caddy `caddy:latest` (`v2.11.9`)** di atas `ubuntu:26.10` (`127.0.0.1:8443 -> 443/tcp`, domain lokal `https://erp.localhost:8443`).

### B. Disiplin Satu Build (*Single-Build Discipline*) & Larangan Duplikasi Build

1. **Satu Definisi `build:` per Image Aplikasi**:
   - Pada konfigurasi Docker Compose yang menjalankan beberapa replika container (`gateway-1` & `gateway-2` di `backend/deploy/docker-compose.apps.yml`, serta `frontend-1` & `frontend-2` di `frontend/docker-compose.yml`), blok `build:` **hanya boleh didefinisikan pada instance pertama (`gateway-1` dan `frontend-1`)**.
   - Instance replika berikutnya (`gateway-2`, `frontend-2`, dst.) wajib menggunakan tag `image:` yang sama dengan `pull_policy: never` pada YAML anchor agar Docker Compose Bake **hanya menjalankan tepat 1 build per image** (`backend`, `frontend`, `proxy`) dan tidak pernah memicu kompilasi paralel ganda.
2. **Kebersihan Riwayat Build (`docker buildx history`) & Image**:
   - Jika selama proses debugging atau pengembangan terdapat *build* yang gagal (`Error`) atau duplikat, agen **wajib** menghapus riwayat *build* lama/gagal (`docker buildx history rm <id>` atau `docker buildx history rm --all` sebelum menjalankan *build* bersih terakhir) serta membersihkan image *dangling* (`docker image prune -f`).
   - Kondisi akhir pada `docker buildx history ls` wajib bersih: hanya memuat **1 record `Completed`** per komponen yang dibangun (`backend`, `frontend`, `proxy`) tanpa status `Error` atau duplikat.

### C. Kewajiban *Build After Development*

Setiap kali agen selesai melakukan perubahan kode atau konfigurasi pada `backend/`, `frontend/`, `proxy/`, atau `backend/deploy/`, agen **dilarang menyatakan tugas selesai** sebelum mengeksekusi verifikasi *Build After Development* berikut:

1. Validasi sintaks dan struktur Compose:
   ```powershell
   docker compose config --quiet
   ```
2. Jalankan *single build* dan perbarui container yang berjalan hingga seluruh *healthcheck* lulus:
   ```powershell
   docker compose build
   docker compose up -d --wait
   docker compose ps
   ```
3. Pastikan seluruh 7 container (`postgres`, `redis`, `gateway-1`, `gateway-2`, `frontend-1`, `frontend-2`, `caddy`) berstatus `Up` dan `(healthy)`, serta verifikasi endpoint HTTPS melalui Caddy:
   ```powershell
   curl.exe -sk --resolve erp.localhost:8443:127.0.0.1 "https://erp.localhost:8443/api/graphql?query=%7Bhealth%7D"
   ```
   Respons wajib `{"data":{"health":"ok"}}` (`HTTP 200`).

### D. Protokol *Continuous Integration (CI) & Pre-Push Quality Gates*

Karena seluruh agen bekerja langsung di branch `main`, setiap perubahan kode wajib melewati **5 Gerbang CI (CI Gates)** secara berurutan sebelum melakukan `git commit` dan `git push origin main`:

1. **Gate 1 — Static Analysis, Formatting & Codegen**:
   - **Go**: Jalankan `go fmt ./...` dan `go vet ./...` pada modul yang diubah (wajib 0 error/warning).
   - **Protobuf** (bila menyentuh `backend/contracts/proto/`): Jalankan `cd backend/contracts && buf lint` dan `make -C backend proto` (atau `proto-<management>`).
   - **GraphQL** (bila menyentuh SDL/resolver/model di `backend/gateway/`): Jalankan generator resmi (`cd backend/gateway && go generate ./...` / `make -C backend gql` bila target Makefile tersedia). Dilarang mengedit file di `backend/gateway/graph/generated/` maupun `backend/gateway/graph/model/*_models_gen.go` secara manual.
   - **Frontend** (bila menyentuh `frontend/`): Pastikan `npm run lint`, `npx tsc --noEmit`, dan `npm run build` (serta `docker compose build frontend-1` bila dependencies hanya berada di stage Docker) lulus tanpa error atau warning yang memblokir.
2. **Gate 2 — Service Layer Unit, Mock & Race Tests**:
   - Jalankan `go test -race ./...` pada package domain service yang dimodifikasi menggunakan `mock_db` sesuai aturan pengujian service. Jika menyentuh manifest RBAC (`backend/shared/rbac/catalogs/*.json`) atau izin modul, pastikan test validasi katalog RBAC dan test anti-*drift* kode permission modul tersebut ikut lulus.
3. **Gate 3 — Real E2E & RBAC TestAB (`*_test` Database di Docker)**:
   - Jalankan pengujian E2E nyata ke `/api/graphql` dengan 3 skenario wajib (**Skenario A** Authorized User 1, **Skenario B** Unauthorized User 2, **Skenario C** Unauthenticated/Anonymous) menggunakan database disposable berakhiran `_test` pada container PostgreSQL Docker (`127.0.0.1:6543`) dan Redis TLS Docker (`127.0.0.1:6379`).
4. **Gate 4 — *Build After Development* & Clean Build History**:
   - Jalankan `docker compose build` (sertakan `iam` apabila ada perubahan pada `backend/shared/rbac/catalogs/*.json` agar image/binary IAM memuat katalog RBAC terbaru) dan `docker compose up -d --wait`. Pastikan seluruh container `Healthy` di `docker compose ps` dan riwayat `docker buildx history ls` bersih (hanya 1 build `Completed` per komponen, tanpa `Error` atau duplikat).
5. **Gate 5 — Sinkronisasi `codebase-memory-mcp` & Git Hygiene**:
   - Jalankan `index_repository` pada `codebase-memory-mcp`, periksa `git diff --check` (bebas konflik marker, whitespace error, maupun file `.env`/secret), lakukan `git add <path-spesifik>` hanya pada file dalam scope tugas, buat commit detail, dan jalankan `git push origin main`.

### E. Protokol Konkurensi Multi-Agen (Saat 2+ Agen Management Bekerja Bersamaan)

Dalam monorepo ini, setiap microservice (`backend/services/{management}/`) terisolasi penuh (memiliki `go.mod` dan database sendiri), tetapi berbagi **3 lapisan bersama (*shared layers*)**: `backend/gateway/`, `backend/contracts/`, dan `docker compose build gateway-1`. Agar dua atau lebih agen yang bekerja paralel tidak saling mematahkan *build*:

1. **Urutan Pengerjaan *Bottom-Up Atomic* (Pencegahan Utama)**:
   - **Tahap 1 — Kerjakan Lapisan Terisolasi Terlebih Dahulu (`backend/services/{management}/`)**: Selesaikan DBML, migrasi SQL, repository, domain service, dan unit test (`mock_db`) di dalam folder microservice masing-masing sampai `go vet ./...` dan `go test -race ./...` pada service tersebut **100% lulus**. Selama tahap ini, agen-agen di management berbeda tidak akan pernah saling mengganggu.
   - **Tahap 2 — Update Kontrak Protobuf Secara Atomik (`backend/contracts/`)**: Tulis atau perbarui `.proto` domain, lalu **seketika itu juga** jalankan `buf lint` dan `make -C backend proto-<management>` sampai kompilasi bersih sebelum berpindah langkah.
   - **Tahap 3 — Sentuh `backend/gateway/` Hanya di Langkah Terakhir (*Atomic Window*)**: Setelah `git status --short backend/gateway/` bersih dari pekerjaan agen lain pada Zona 2, lengkapi `internal/adapter/{management}/`, wrapper `internal/source_connections/{management}*.go` (bila perlu) dan registrasi *additive-only* di `internal/source_connections/registry.go` + `cmd/main.go`, edit skema `graphql/{management}/{management}_{domain}.graphql` milik management sendiri, jalankan generator resmi (`cd backend/gateway && go generate ./...`) sehingga `graph/generated/{management}*.generated.go`, `graph/model/{management}*_models_gen.go`, dan `graph/resolvers/{management}*.resolvers.go` terbentuk terpisah per domain, lalu lengkapi delegator tipis `{management}*.resolvers.go` / helper `Node`, adapter/mapper, dan test wiring dalam **satu langkah kerja utuh tanpa jeda** sampai `go vet ./...`, `go build ./cmd/...`, dan test terkait kembali hijau (lihat [`Aturan Tanggung Jawab Wiring Management ke Gateway`](#aturan-tanggung-jawab-wiring-management-ke-gateway)). **Dilarang keras** meninggalkan file di `backend/gateway/` atau `backend/contracts/` dalam kondisi setengah jadi (*uncompilable*).
2. **Pemeriksaan Sebelum Menjalankan Generator Bersama (`gql` / `proto`)**:
   - Sebelum menjalankan generator GraphQL (`cd backend/gateway && go generate ./...` / `make -C backend gql`), `make -C backend proto`, atau `docker compose build gateway-1`, wajib periksa `git status --short backend/gateway/ backend/contracts/`.
   - Jika terdapat file `.graphql`/`.graphqls`, `.proto`, atau resolver milik **management lain (Agen 2)** yang sedang *modified/untracked*, periksa apakah file tersebut valid dan jangan pernah menimpa, menghapus, atau me-*reset* pekerjaan agen lain.
3. **Prosedur Saat Agen 1 Selesai tetapi Build Global Tertahan oleh Error Agen 2 yang Belum Selesai**:
   - **Langkah 1 (Verifikasi Terisolasi Scope Agen 1)**: Buktikan terlebih dahulu bahwa seluruh kode milik Agen 1 bebas error dengan menjalankan `go fmt`, `go vet`, dan `go test -race` secara spesifik pada modul service Agen 1 (`backend/services/{management_1}/...`) dan package resolver spesifik milik Agen 1 (`backend/gateway/graph/resolvers/{management_1}/...`).
   - **Langkah 2 (Identifikasi Faktual Pemilik Error)**: Periksa lokasi file dan baris pada pesan error compiler (`file:line`) dan cocokkan dengan `git status --short` untuk memastikan 100% bahwa error berasal dari file *uncommitted* milik management Agen 2. Agen 1 **dilarang keras** menjalankan `git checkout --`, `git restore`, `git reset`, atau mengubah aturan bisnis milik Agen 2 (`agents/boundary.md`).
   - **Langkah 3 (Jalur Penyelesaian Terkoordinasi)**:
     - **Kondisi A (Agen 2 sedang aktif berjalan secara paralel)**: Tunggu Agen 2 menyelesaikan kompilasi blok kodenya (*safe compilable checkpoint*) atau koordinasikan melalui Leader agar Agen 2 menutup *syntax/compile error* pada file bersama tersebut terlebih dahulu sebelum Agen 1 mengeksekusi Gate 4 (`docker compose build` & `up -d --wait`).
     - **Kondisi B (Sesi Agen 2 sudah berhenti/terputus tetapi meninggalkan kode yang memblokir kompilasi Gateway)**:
       - Jika error hanya disebabkan oleh method stub resolver hasil `gqlgen` yang belum diisi body-nya, tambahkan *fail-closed return* standar (`return nil, errors.New("not implemented")`) agar kompilasi `backend/gateway` tidak patah.
       - Jika error menyangkut struktur kode/kontrak Agen 2 yang lebih dalam, laporkan daftar tepat file dan baris error milik Agen 2 kepada Leader/Owner untuk diisolasi terlebih dahulu sebelum menjalankan `docker compose build`, lalu lakukan staging **hanya** pada path file milik Agen 1 (`git add <path-spesifik-agen-1>`; dilarang keras memakai `git add .` atau `git add -A`).

## Audit Development Berbasis Domain (DDD)

- Audit development memakai pendekatan Domain-Driven Development: tetapkan satu domain aktif beserta batas bisnis dan pemilik datanya. Telusuri domain itu secara vertikal dari DBML/SQL, service, kontrak/API, hingga pengujian sebelum beralih ke domain lain.

- Selesaikan temuan dalam lingkup tugas dan verifikasi domain aktif terlebih dahulu. Temuan atau dependency lintas domain dicatat pada management pemilik; jangan memperluas audit menjadi beberapa domain sekaligus. Domain yang terhalang dependency harus dilaporkan sebagai belum selesai.

- Domain `workforce` pada Inventory & Warehouse Management berfokus pada orkestrasi pekerjaan gudang (kelas, pool, template, standar, dan eksekusi tugas). Workforce Management memiliki roster/jadwal, penugasan shift kanonik, presensi, serta waktu kerja. HRM memiliki master pekerja dan kontrak; IAM memiliki principal, autentikasi, dan otorisasi.

- `warehouse_shifts` dan `warehouse_worker_assignments` dirancang sebagai proyeksi operasional WFM. Jangan menjadikannya sumber roster/presensi/payroll kedua. Sinkronisasi harus memakai referensi logis dan kontrak pemilik data, termasuk pemetaan worker HRM ke principal IAM; jangan menyamakan kedua ID atau menambah FK lintas database.

- CRUD shift/assignment lokal yang sudah ada tidak membuktikan integrasi WFM. Sebelum mengaktifkan sinkronisasi, tetapkan sumber penulisan tunggal, pemetaan waktu/status, idempotensi/versi, penanganan pembatalan, serta E2E untuk seluruh query dan mutation terkait. Laporkan gap sampai mekanisme tersebut benar-benar tersedia.

## Aturan Pengujian Service

Strategi pengujian repositori ini hanya memiliki **dua lapisan**:

1. **Service**: unit, mock, race, dan benchmark untuk aturan bisnis dan operasi service.

2. **E2E**: integration dan full dataflow test yang menjalankan komponen nyata sampai state persisten, dengan skenario bisnis dan kegagalan yang relevan.

Jangan menambah repository test atau integration test terpisah sebagai lapisan wajib ketiga. Repository dan query nyata divalidasi melalui E2E. Kelompokkan kasus yang berkaitan, gunakan table-driven tests untuk variasi input, dan hindari satu file untuk setiap kasus kecil. Tambahkan tes hanya untuk perilaku atau risiko yang belum dicakup kedua lapisan.

- Semua kode tes harus berada di folder `test/` di dalam area management/domain yang diuji; buat folder tersebut jika belum ada. Contoh: `backend/services/{service_name}/internal/service/{domain}/test/` untuk service dan `backend/gateway/graph/resolvers/{management}/{domain}/test/` untuk unit test resolver. Jangan letakkan kode tes berdampingan dengan implementasi atau kode init/runtime seperti `db/init.go`.

- Gunakan package `service_test` untuk tes service domain dan uji melalui API yang diekspor. Simpan fixture `mock_db` di folder tes domain dan injeksikan melalui interface dependency service. Tes service tidak memanggil repository nyata, driver/database, atau service eksternal.

- Cakupan service melalui `mock_db` mencakup aturan bisnis, hasil dan perubahan state, kegagalan penting, serta operasi utama yang tersedia (read/create/update/delete bila relevan). Pastikan race test benar-benar menjalankan operasi bersamaan dengan sinkronisasi deterministik. Hindari menduplikasi skenario yang sudah terwakili.

- Benchmark termasuk lapisan service dan diprioritaskan untuk alur penting, kompleks, atau berisiko kinerja; tidak wajib untuk setiap operasi CRUD. Benchmark yang dibuat memakai `mock_db`, mengisolasi setup dari bagian yang diukur, menjaga state antariterasi, dan memisahkan fungsi operasi jika penggabungan mencemari hasil. Beberapa benchmark terkait boleh disimpan dalam satu file tes domain.

- Jalankan unit/mocking dan race test service yang relevan dengan perubahan. Jalankan benchmark terpilih dengan `-benchmem` bila relevan (serta `-cpu=1,2,4,8` jika mengukur skalabilitas konkurensi). Catat pemeriksaan yang benar-benar dijalankan.

- **Wajib Simpan Laporan Benchmark Services (`agents/benchmark/`)**: Hasil eksekusi benchmark service wajib dilaporkan pada `agents/benchmark/{management}/{id}-{benchmark-title}.md` (`{id}` berurutan mulai dari `1` per folder `{management}` `snake_case`, judul lowercase kebab-case). Setiap laporan minimum memuat: target `domain` & file tes, perintah aktual yang dijalankan, kondisi dataset/`mock_db`, hasil metrik nyata (`N`, `ns/op`, `B/op`, `allocs/op`), analisis bottleneck/alokasi, serta batasan bahwa benchmark mock service tidak merepresentasikan kapasitas DB/Redis produksi.

## Pengujian E2E dan GraphQL

- Simpan kode E2E di folder `test/` area terkait dan kelompokkan per management/domain; buat folder bila belum ada. Contoh Gateway–IAM: `backend/gateway/test/iam/{domain}/{domain}_e2e_test.go`. Jangan menaruh tes di folder init, service, transport, atau runtime.

- Untuk **setiap operasi GraphQL yang tersedia di SDL pada setiap domain**, E2E wajib menjalankan operasi tersebut: semua `query` untuk baca/cari/list dan semua `mutation` untuk create/update/delete atau aksi domain lain. Verifikasi input dan filter, bentuk/error respons, pagination bila ada, serta perubahan atau keberadaan state persisten sesuai perilaku operasi. Jangan mengasumsikan CRUD tersedia jika SDL tidak menyediakannya.

- E2E memvalidasi full dataflow lintas komponen dengan service nyata; jika persistence termasuk alur, gunakan database testing disposable berakhiran `_test`. Untuk GraphQL Gateway, kirim HTTP POST JSON ke `/api/graphql` dan validasi sambungan gRPC ke service terkait. Untuk batas Go `internal`, jalankan service pemilik sebagai proses terpisah atau gunakan harness E2E eksternal yang sah.

### Uji fungsional HTTPS dengan superadmin

Untuk uji fungsional GraphQL di lingkungan development atau testing:

1. Login sebagai `superadmin` melalui endpoint HTTPS `/api/graphql`.
2. Cari kode permission yang dibutuhkan melalui Query GraphQL IAM `permissions`. Jika kode belum terdaftar, gunakan manifest katalog resmi dan bootstrap IAM di lingkungan uji. Buat role uji dengan Mutation `createRole` atau gunakan role uji yang sudah ada, lalu pasang permission melalui Mutation `assignRolePermission`.
3. Assign role uji ke akun `superadmin` melalui Mutation `assignUserRole`. Login ulang agar sesi menggunakan grant terbaru.
4. Jalankan Query dan Mutation melalui HTTPS `/api/graphql` dengan sesi `superadmin`. Sertakan token CSRF yang sah pada Mutation dan catat hasil aktual sesuai aturan laporan E2E.

Gunakan akun dengan grant terbatas untuk menguji penolakan akses di luar scope, termasuk lintas badan hukum. Izin global `superadmin` dapat menutupi pembatasan tersebut. Ikuti [tata kelola IAM](agents/users.md#5-tata-kelola-permission-manifest-kode--pemberian-akses-lintas-management) saat mendaftarkan izin dan memberikan role.

Semua perubahan role, pemasangan atau pencabutan permission pada role, serta pemberian atau pencabutan role pada akun uji wajib melalui GraphQL IAM, termasuk saat cleanup. Penulisan langsung ke tabel RBAC IAM melalui SQL dilarang, termasuk pada database disposable.

### Cakupan dan bukti E2E

- Cakup jalur berhasil dan kegagalan yang berdampak. Kelompokkan variasi sejenis dengan table-driven tests dan hindari kombinasi yang tidak menambah cakupan berarti. Bersihkan data tes setelah pengujian. Jangan membuat suite repository, transport/API, atau integration test terpisah.

- Pastikan perintah test module mengompilasi paket E2E. Catat prasyarat dan hasil aktual.

- **Wajib Simpan Laporan E2E (`agents/e2e/`)**: Setiap eksekusi E2E wajib didokumentasikan di `agents/e2e/{management}/{id}-{e2e-title-{Query***/Mutation***}}.md` (`{id}` berurutan mulai dari `1` per folder `{management}` `snake_case`; dalam nama berkas nyata gunakan lowercase kebab-case yang memuat penanda `query` / `mutation` / `query-mutation` beserta operasi yang diuji, misal `1-authentication-mutation-user-login.md`). Laporan wajib memuat: daftar operasi `Query`/`Mutation` yang diuji, prasyarat runtime aktif (Gateway, gRPC owner, DB `*_test`, Redis test), perintah tes aktual, commit SHA, serta tabel **berbagai *cases*** (Authorized, Unauthorized/Wrong Scope, Unauthenticated/Invalid CSRF, validasi input/error, konkurensi/CAS/idempotensi, dan bukti state DB `*_test` / zero-write) beserta hasil jujur (`PASS`/`FAIL`/`BLOCKED`).

## Aturan Khusus Backend Engineers: Skill Golang & Superpowers

Before any Go coding, review, debugging, troubleshooting, or setup task, load the `samber/cc-skills-golang@golang-how-to` skill first — it routes to whichever other Go skills the task needs.

Seluruh agen dan developer Backend Go wajib membaca serta menerapkan skill Golang dan Superpowers yang relevan dengan tugas yang sedang dikerjakan (lokasi skill lokal: `~/.gemini/config/skills/` dan `~/.gemini/config/plugins/{cc-skills-golang,ecc,superpowers}/skills/`).

### 1. Required Go skills (Core Selalu Relevan untuk Proyek Ini)

Selain `samber/cc-skills-golang@golang-how-to` sebagai orchestrator utama, kelompok skill berikut menjadi baseline wajib sesuai konteks tugas Go di repositori ini:

- `samber/cc-skills-golang@golang-code-style`
- `samber/cc-skills-golang@golang-naming`
- `samber/cc-skills-golang@golang-documentation`
- `samber/cc-skills-golang@golang-error-handling`
- `samber/cc-skills-golang@golang-safety`
- `samber/cc-skills-golang@golang-security`
- `samber/cc-skills-golang@golang-data-structures`
- `samber/cc-skills-golang@golang-database`
- `samber/cc-skills-golang@golang-design-patterns`
- `samber/cc-skills-golang@golang-structs-interfaces`
- `samber/cc-skills-golang@golang-modernize`
- `samber/cc-skills-golang@golang-testing`
- `samber/cc-skills-golang@golang-troubleshooting`

Skill pendamping spesifik stack repositori (muat setiap kali menyentuh area terkait):

- **gRPC & Protobuf (Buf):** `samber/cc-skills-golang@golang-grpc` dan `grpc-golang`
- **GraphQL Gateway (`gqlgen`):** `samber/cc-skills-golang@golang-graphql` dan `gqlgen-go`
- **Concurrency & Context Propagation:** `samber/cc-skills-golang@golang-concurrency`, `samber/cc-skills-golang@golang-context`, dan `go-concurrency-patterns`
- **Review & Resolusi Build Go (`ecc`):** `ecc:go-reviewer` (`plugins/ecc/skills/go-reviewer/SKILL.md`), `ecc:go-build-resolver` (`plugins/ecc/skills/go-build-resolver/SKILL.md`), `ecc:golang-patterns`, `ecc:golang-testing`
- **Layanan Mikro & Pola Go Modern:** `golang-pro`

### 2. Matriks Pemetaan Tugas Backend → Skill Golang

| Jenis Pekerjaan Backend | Skill Utama | Skill Pendamping |
|---|---|---|
| Desain layer `domain/`, struct, interface, functional options | `golang-design-patterns`, `golang-structs-interfaces` | `golang-naming`, `golang-code-style` |
| Implementasi service, error wrapping/sentinel, pencegahan nil panic | `golang-error-handling`, `golang-safety` | `golang-context`, `golang-code-style` |
| Akses PostgreSQL/GORM/pgx, transaksi, migrasi & pooling | `golang-database` | `golang-error-handling`, `golang-security`, `postgresql`, `postgres-best-practices` |
| Kontrak `.proto`, Buf lint, interceptor, TLS/mTLS gRPC | `golang-grpc`, `grpc-golang` | `golang-testing`, `golang-error-handling` |
| Skema GraphQL, `gqlgen.yml`, resolver Gateway, dataloader, custom scalar | `golang-graphql`, `gqlgen-go` | `golang-testing`, `golang-error-handling` |
| Goroutine, channel, worker pool, graceful shutdown, deadline | `golang-concurrency`, `golang-context` | `go-concurrency-patterns`, `golang-safety` |
| Keamanan auth, cookie, crypto, Argon2id/TOTP, cegah SQLi/SSRF/PII di log | `golang-security` | `golang-safety`, `golang-lint` |
| Unit test, table-driven, `t.Context`, mock, race detector | `golang-testing`, `ecc:golang-testing` | `golang-stretchr-testify` (bila memakai `testify`), `golang-modernize` |
| Benchmark service (`-benchmem`), pprof, optimasi alokasi/hot-path | `golang-benchmark`, `golang-performance` | `golang-data-structures`, `golang-troubleshooting` |
| Observability produksi (`slog`, Prometheus, OpenTelemetry) | `golang-observability` | `golang-error-handling`, `golang-samber-slog` |
| Debug bug/race/deadlock/panic atau perbaikan error `go build`/`go vet` | `golang-troubleshooting`, `ecc:go-build-resolver` | `golang-safety`, `golang-lint` |
| Refactoring *mother file* / *god file* (*anti god-file*), rename aman, navigasi simbol | `golang-refactoring`, `golang-gopls` | `golang-project-layout`, `golang-naming`, `golang-modernize` |
| Audit library, cek `go.mod`/`go.work`, dokumen package via pkg.go.dev | `golang-dependency-management`, `golang-pkg-go-dev`, `golang-popular-libraries` | `golang-security` (`govulncheck`) |
| Code review perubahan Go sebelum commit | `ecc:go-reviewer`, `golang-lint` | `golang-code-style`, `golang-safety`, `golang-security` |

> **Aturan Library Opsional:** Skill khusus framework/library eksternal (seperti `golang-google-wire`, `golang-uber-dig`, `golang-uber-fx`, `golang-samber-do`, `golang-samber-hot`, `golang-samber-lo`, `golang-samber-mo`, `golang-samber-oops`, `golang-samber-ro`, `golang-samber-slog`, `golang-spf13-cobra`, `golang-spf13-viper`, `golang-swagger`, `golang-stretchr-testify`, `dbos-golang`, `temporal-golang-pro`) **hanya dimuat jika `go.mod` module yang dikerjakan memang memakai library tersebut atau saat sedang mengevaluasi adopsinya**.

### 3. Katalog Lengkap Skill Golang Terinstal

1. **`samber/cc-skills-golang` (46 skill di `plugins/cc-skills-golang/skills/`, ter-link di `skills/`):**
   - *Orchestrator & Navigasi:* `golang-how-to`, `golang-gopls`, `golang-pkg-go-dev`, `golang-stay-updated`.
   - *Kualitas, Gaya & Keamanan Kode:* `golang-code-style`, `golang-naming`, `golang-documentation`, `golang-lint`, `golang-error-handling`, `golang-safety`, `golang-security`, `golang-modernize`, `golang-refactoring`.
   - *Arsitektur, Struktur Data & Concurrency:* `golang-project-layout`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-data-structures`, `golang-concurrency`, `golang-context`, `golang-database`.
   - *API & Transport:* `golang-grpc`, `golang-graphql`, `golang-swagger`.
   - *Testing, Debugging, Observability & Performa:* `golang-testing`, `golang-stretchr-testify`, `golang-troubleshooting`, `golang-benchmark`, `golang-performance`, `golang-observability`.
   - *Dependency, CLI & CI:* `golang-dependency-management`, `golang-popular-libraries`, `golang-dependency-injection`, `golang-cli`, `golang-continuous-integration`.
   - *Library & Framework Spesifik:* `golang-google-wire`, `golang-uber-dig`, `golang-uber-fx`, `golang-samber-do`, `golang-samber-hot`, `golang-samber-lo`, `golang-samber-mo`, `golang-samber-oops`, `golang-samber-ro`, `golang-samber-slog`, `golang-spf13-cobra`, `golang-spf13-viper`.
2. **Skill Go Standalone (8 skill di `skills/`):**
   - `golang-pro` (pola modern Go 1.21+, concurrency lanjutan, microservices produksi).
   - `grpc-golang` (gRPC Go, kontrak Protobuf/Buf, mTLS, streaming, OpenTelemetry).
   - `gqlgen-go` (GraphQL server `gqlgen` v0.17+, `gqlgen.yml`, resolver `follow-schema`, dataloader, custom scalar `Decimal`, security directives, Fiber v3).
   - `go-concurrency-patterns` (goroutines, channels, `sync` primitives, worker pools, race prevention).
   - `dbos-golang` (workflow durable Go berbasis DBOS).
   - `temporal-golang-pro` (sistem terdistribusi deterministik dengan Temporal Go SDK).
   - `go-playwright` (otomasi browser berbasis Playwright Go — bila diperlukan untuk QA/automation Go).
   - `go-rod-master` (otomasi browser & CDP berbasis `go-rod` — bila diperlukan).
3. **Skill Go dari Plugin ECC (4 skill Go di `plugins/ecc/skills/`):**
   - `go-reviewer` (audit idiomatik Go, concurrency, error handling, performa).
   - `go-build-resolver` (perbaikan error kompilasi, `go vet`, dan warning linter dengan perubahan minimal).
   - `golang-patterns` (pola idiomatis & best practices Go).
   - `golang-testing` (TDD Go, table-driven tests, subtests, benchmark, fuzzing, coverage).

### 4. Penggunaan Wajib Superpowers Skills (`plugins/superpowers/skills/` & `skills/`)

Gunakan Superpowers skills apabila tugas Backend membutuhkan alur proses disiplin:

- **`using-superpowers`**: memastikan pengecekan skill yang relevan dilakukan sebelum merespons atau mengeksekusi tugas.
- **`brainstorming`**: wajib dipakai sebelum membuat desain fitur, domain, kontrak API, atau skema baru yang belum terkunci keputusannya.
- **`writing-plans` & `executing-plans`**: menyusun rencana bertahap yang terverifikasi dan mengeksekusinya secara disiplin per milestone.
- **`test-driven-development`**: menulis tes unit/domain (`*_test.go`) gagal terlebih dahulu sebelum menulis kode implementasi baru atau perbaikan bug.
- **`systematic-debugging`**: investigasi akar masalah secara berurutan saat menemui panic, bug logika, kegagalan tes, atau race condition — dilarang menebak/menambal tanpa diagnosis.
- **`verification-before-completion`**: wajib menjalankan perintah verifikasi nyata (`go fmt ./...`, `go vet ./...`, `go build ./...`, `go test -race ./...`, dan E2E terkait) serta mengamati output penuh sebelum menyatakan tugas selesai.
- **`requesting-code-review` & `receiving-code-review`**: menjalankan review teknis objektif dan memverifikasi masukan review berbasis bukti kode.
- **`dispatching-parallel-agents` & `subagent-driven-development`**: saat mengerjakan 2+ sub-tugas independen atau rencana bertahap lintas modul tanpa konflik state.
- **`writing-skills`**: hanya ketika membuat atau memperbarui skill.
- **Pengecualian Konflik Aturan Repositori (`main`-only):** Skill `using-git-worktrees` dan `finishing-a-development-branch` **tidak boleh** dipakai untuk membuat branch baru atau PR rutin karena melanggar aturan `main`-only di [Seksi 5](#5-aturan-git-commit--skema-versioning-hierarkis) (gunakan hanya bila ada instruksi eksplisit owner).

### 5. Batas Precedence (`AGENTS.md` > Skill Generik)

Jika ada perbedaan antara pedoman di dalam skill dan aturan repositori, **aturan repositori selalu menang**, khususnya:

1. **Bahasa:** komentar kode memakai **Bahasa Indonesia**; nama identifier, log teknis `slog`, *error code*, dan pesan error memakai **Bahasa Inggris**.
2. **Persistensi & Repository:** memakai **GORM** (`gorm.io/gorm`) dengan penamaan method berbasis operasi DB (`Select...`, `Insert...`/`Create...`, `Update...`, `Delete...`), bukan nama use-case bisnis; dilarang menaruh logika bisnis/hashing di repository.
3. **Skema & Migrasi Pra-Rilis:** tetap pada baseline `000001` serta pemisahan `{alias}` dan `{alias}_test` dengan `db/init.go`.
4. **Git & Keamanan:** dilarang membuat branch baru, dilarang menyimpan kredensial terbuka di repositori, dan dilarang mengedit file generated Protobuf/`gqlgen` secara manual.

## Aturan Khusus Frontend Engineers

Panduan lengkap tersedia di [`frontend/AGENTS.md`](frontend/AGENTS.md). Seluruh agen dan developer Frontend wajib mematuhi invariant utama berikut saat mengerjakan `frontend/`:

1. **Perspektif Produk & Peta Route Resmi (`docs/development-product-perspektif.md`)**:
   - Aplikasi adalah sistem terpadu **Web Commerce + Retail ERP single-tenant** (selama development direpresentasikan brand **IntegriOne**), **bukan** platform pemasaran SaaS multitenant.
   - **`/login` adalah satu pintu masuk umum (*unified login*)** untuk Customer Web Commerce, Internal Employee, Vendor CMT, dan Supplier: dilarang menambah tab pemilih tipe pengguna sebelum login atau istilah backoffice eksklusif pada `/login`.
   - **1 Management ≠ 1 Route Dashboard:** Gunakan peta route resmi (prioritas aktif sekarang **IAM `/identity`**, universal **`/profile`**, internal siap backend `/human-resource`, `/procurement`, `/warehouse`, `/finance`, `/accounting`, `/product`, portal `/supplier` & `/vendor` CMT, serta Web Commerce `/`, `/account`, `/commerce`). Dilarang membuat route dashboard mandiri untuk `enterprise_foundation` atau `document_records_management`, maupun UI/API palsu untuk modul yang backend-nya masih tertunda (`pricing`, `oms`, `crm`, `cs`, `drm`).

2. **Satu HTTP API `/api/graphql`, Opaque Session, CSRF & Keamanan Browser**:
   - Semua request client/server Next.js ke backend **wajib** memakai same-origin **`/api/graphql`** (`credentials: "include"`). Dilarang memanggil path REST `/api/auth/*`, port internal gRPC, atau API eksternal pihak ketiga secara langsung dari browser (Caddy di [`proxy/Caddyfile`](proxy/Caddyfile) menolak `/api/*` selain `/api/graphql` dengan 404).
   - Sesi memakai cookie `HttpOnly` `__Host-integrione_session` (idle 30 menit, absolut 8 jam — dilarang menyediakan opsi *"ingat saya 30 hari"*). **Dilarang keras** menyimpan token sesi, JWT, role, atau permission di `localStorage`, `sessionStorage`, atau cookie JS non-HttpOnly.
   - Setiap mutasi GraphQL wajib menyertakan header `X-CSRF-Token` yang diperoleh dari bootstrap `GET /api/graphql` (`browserSession { authenticated csrfToken }`) atau respons mutasi login/rotasi terbaru.
   - **Zero Client Trust & Anti-IDOR:** Penentuan `Location` redirect dan otorisasi RBAC–ABAC berasal dari server/Gateway dengan validasi allowlist rute internal. Jangan merender atau membiarkan unduhan dokumen DRM berdasarkan UUID mentah tanpa verifikasi server, dan **dilarang** menaruh rahasia apa pun (termasuk kunci Xendit/Midtrans) di variabel `NEXT_PUBLIC_*`.
   - **Integritas Status Asinkron:** Jangan menampilkan label *"Berhasil"* atau *"Lunas"* jika backend/Payment Gateway masih mengembalikan status `PENDING`, dan tampilkan `N/A` (bukan `0` palsu) saat data analitik belum tersedia.

3. **Arsitektur Next.js App Router, React 19 Compiler, GSAP & Modularitas**:
   - Gunakan **React Server Component** secara default; letakkan `"use client"` hanya pada batas komponen interaktif (`*-experience.tsx`, form, komponen beranimasi GSAP, hooks). **Dilarang** mengimpor atau menjalankan `gsap` / `@gsap/react` di Server Component.
   - Hormati konfigurasi [`frontend/next.config.ts`](frontend/next.config.ts) (`reactCompiler: true`, `cacheComponents: true`, `partialPrefetching: true`): komponen/hooks harus murni sesuai Rules of React, gunakan **`useGSAP()`** (`{ scope: containerRef, dependencies: [...] }`) + `contextSafe(...)` dari `@gsap/react` untuk cleanup otomatis & kompatibilitas React 19 (bukan `useEffect` polos tanpa `gsap.context().revert()`), dan data yang bergantung pada cookie/sesi pengguna **tidak boleh** masuk ke cache statis publik (`cache: "no-store"` / dinamis per-request).
   - Jaga modularitas (*anti god-component* / larangan *mother file*, tanpa batasan jumlah baris kode): pisahkan typed GraphQL operations, helper domain, form validation, dan sub-komponen UI secara terstruktur per route domain (`frontend/src/app/{route}/`) dan modul bersama di `frontend/src/`.

4. **Design System IntegriOne, Motion (GSAP) & Aksesibilitas (WCAG 2.2 AA)**:
   - Konsisten dengan sistem visual **Enterprise Editorial** IntegriOne (palet Navy `#133554` / `#0B2439`, Coral `#DA414A`, Off White `#F9FAFB`; tipografi `Fraunces` & `Source Sans 3` dari [`frontend/src/app/layout.tsx`](frontend/src/app/layout.tsx); styling Tailwind v4 + CSS Modules ter-scope).
   - Untuk animasi **GSAP (`gsap` `^3.15.1` + `@gsap/react` `^2.1.2`)**, gunakan properti GPU (`transform`, `opacity`) tanpa layout thrashing (`top`/`left`/`width`/`height`), bungkus dengan `gsap.matchMedia()` untuk `(prefers-reduced-motion: no-preference)`, pertahankan gerak **fungsional & cepat (`~150–250ms`)** pada dashboard internal ERP/portal (tanpa *scroll-jacking*/splash blocking) sementara storefront Web Commerce (`/`, `/login`) boleh memakai koreografi editorial lebih kaya.
   - Wajib responsif dari mobile (`320px`) hingga desktop lebar tanpa horizontal overflow, memakai semantic HTML, `htmlFor` pada input, `aria-live` / pesan error terhubung ke field, contrast memadai, keyboard navigation jelas, dan touch target interaktif minimal `44x44px`.

5. **Verifikasi, Build Container & Pelaporan Frontend**:
   - Setiap perubahan di `frontend/` wajib diverifikasi: `cd frontend && npx next typegen && npm run lint && npx tsc --noEmit` (plus `npm run build`), lalu `docker compose build frontend-1`, `docker compose up -d --wait frontend-1 frontend-2 caddy`, serta uji browser nyata lewat Caddy `https://erp.localhost:8443` (termasuk mobile/zoom, `prefers-reduced-motion`, & skenario unauthorized/CSRF). Jangan lupa sinkronkan `frontend/package-lock.json` setiap menambah/mengubah dependency.
   - Seluruh artefak tracking Frontend dipisahkan dari folder Backend `agents/` ke bawah **`agents/frontend/`** per `{management}` (gunakan `core` untuk fondasi lintas aplikasi): instruksi `agents/frontend/instructions/{management}/{id}-{title}-{undone/done}.md` (`-undone.md` → `-done.md` pasca-selesai), konteks `agents/frontend/context/{management}/{id}-context.md`, *working history* `agents/frontend/working_history/{management}/`, audit `agents/frontend/audit/{management}/`, benchmark `agents/frontend/benchmark/{management}/`, dan laporan E2E `agents/frontend/e2e/{management}/`. Workrule otoritatif Frontend tersimpan di [`frontend/AGENTS.md`](frontend/AGENTS.md).

## Struktur Model dan Resolver GraphQL Gateway

Untuk mencegah munculnya *motherfile* raksasa (`generated.go`, `models_gen.go`) maupun konflik lintas tim saat 13 management berkembang, pemecahan skema, kode hasil generate, model, dan resolver Gateway wajib mengikuti pola modular **Hybrid A1** berikut:

- **Pemecahan Skema per Management & Domain (`graphql/**/*.graphql`):**
  - Simpan kernel GraphQL bersama pada `backend/gateway/graphql/schema.graphql` secara minimal (*non-motherfile*: `scalar Time`, `ScopeType`/`Scope`, Relay `PageInfo`, `interface Node`, root `Query`, dan placeholder `Mutation`).
  - Organisasikan skema per management dan sub-domain di `backend/gateway/graphql/{management}/{management}_{domain}.graphql` (atau `graphql/{management}.graphql` untuk modul ringkas). Karena generator `gqlgen` menamai berkas keluaran berdasarkan *basename* (`filepath.Base`), **setiap file `.graphql` wajib memakai nama unik berprefix `{management}_`** (misalnya `iam_auth.graphql`, `iam_user.graphql`, `plm_product.graphql`), dan `gqlgen.yml` memuatnya dengan pola `graphql/**/*.graphql`.

- **Pemecahan Kode Eksekusi (`exec.layout: follow-schema`):**
  - Konfigurasi `exec:` di `backend/gateway/gqlgen.yml` wajib memakai `layout: follow-schema` (`dir: graph/generated`, `package: generated`, `filename_template: "{name}.generated.go"`).
  - Boilerplate global berada pada `root_.generated.go`, dispatcher root `Query`/`Mutation` pada `schema.generated.go`, sedangkan seluruh tipe, input, interface, edge, dan connection setiap domain terpecah 1:1 ke `backend/gateway/graph/generated/{management}_{domain}.generated.go`. Dilarang mengedit berkas `graph/generated/*.go` secara manual.

- **Pemecahan Model Ter-Generate (`package model` Multi-File + `autobind` Opsional):**
  - Dilarang menumpuk seluruh model hasil generate dalam satu motherfile `models_gen.go`. Melalui custom `modelgen` splitter di `backend/gateway/generate.go`, keluaran model dipecah 1:1 mengikuti asal berkas skema dalam satu `package model`: `graph/model/schema_models_gen.go` untuk tipe inti bersama serta `graph/model/{management}_{domain}_models_gen.go` untuk tipe, input, enum, edge, dan connection domain.
  - Pemilihan satu `package model` dengan file terpecah mencegah *Go import cycle* ketika relasi GraphQL lintas management saling merujuk (misalnya OMS ↔ Finance ↔ WMS ↔ IAM) serta menghilangkan beban penulisan boilerplate Relay secara manual. Tipe kustom tertentu tetap dapat diarahkan melalui `autobind:` / `@goModel` ke `graph/model/{management}/{domain}/model.go` apabila benar-benar membutuhkan perilaku Go khusus (dengan tetap menjaga tipe global `Node`/`PageInfo` tidak bergantung ke domain). Dilarang mengedit `*_models_gen.go` secara manual.

- **Pemecahan Resolver (`resolver.layout: follow-schema` + Delegator Tipis):**
  - Konfigurasi `resolver:` wajib memakai `layout: follow-schema` (`dir: graph/resolvers`, `filename_template: "{name}.resolvers.go"`), sehingga setiap file `.graphql` menghasilkan berkas delegator terpisah (`{management}_{domain}.resolvers.go`).
  - Pertahankan method pada `graph/resolvers/*.resolvers.go` sebagai delegator tipis. Implementasi transport adapter, otorisasi per operasi, pemanggilan gRPC, dan pemetaan protobuf ↔ model `package model` ditempatkan terisolasi per management di `backend/gateway/internal/adapter/{management}/` (atau `backend/gateway/graph/resolvers/{management}/{domain}/resolver.go` + `mapper/` bila dipisah ke subpackage resolver). Jangan membuat dependency silang antar adapter/resolver domain maupun membuat package/skema placeholder kosong untuk management yang belum diimplementasikan.

- **Perbarui & Regenerasi Resmi:** Bila ada pemecahan file `.graphql` atau perubahan binding `gqlgen.yml` / plugin `generate.go`, jalankan generator resmi (`cd backend/gateway && go generate ./...`) lalu verifikasi diff berkas per domain. Untuk skala besar, opsi performa resmi `gqlgen` (`fast_validation: true`, `use_buffer_pooling: true`, dan bila diperlukan `omit_resolver_embedding: true` / `skip_import_grouping: true`) diaktifkan secara terpusat pada `gqlgen.yml`.

### Aturan Tanggung Jawab Wiring Management ke Gateway

Setiap developer backend management dalam cakupan aktif (`13 management` pada `agents/boundary.md`) **berhak sekaligus bertanggung jawab penuh** mengerjakan dan memelihara wiring domain miliknya ke `backend/gateway/` (gRPC adapter, koneksi sumber, schema/mapper/resolver GraphQL, hingga test wiring & E2E Gateway). Supaya pengembangan paralel antar-management tetap aman dari konflik Git, bug kompilasi, dan pelanggaran keamanan `D01–D12`, seluruh coding di `backend/gateway/` wajib mematuhi aturan berikut:

1. **Partisi Kepemilikan 3 Zona di `backend/gateway/`**:
   - **Zona 1 — Milik Eksklusif Management (Boleh Dikoding Penuh oleh Owner `{management}`)**:
     - `backend/gateway/graphql/{management}/{management}*.graphql` (atau `backend/gateway/graphql/{management}*.graphql` untuk modul ringkas; tetap harus unik lintas folder)
     - `backend/gateway/graph/generated/{management}*.generated.go` (hasil regenerasi dari skema management sendiri; tidak diedit manual)
     - `backend/gateway/graph/model/{management}*_models_gen.go` (hasil regenerasi dari skema management sendiri; tidak diedit manual), serta `backend/gateway/graph/model/{management}/{domain}/` bila memakai `autobind` kustom
     - `backend/gateway/graph/resolvers/{management}*.resolvers.go` (delegator terpecah per subdomain dan helper dispatch `Node` terpisah per management jika diperlukan, tanpa menumpuk logika di `schema.resolvers.go`)
     - `backend/gateway/graph/resolvers/{management}/...` (jika memakai subpackage resolver/mapper) serta `backend/gateway/internal/adapter/{management}/...` dan wrapper koneksi per management di `backend/gateway/internal/source_connections/{management}*.go`
     - `backend/gateway/test/{management}/...` (khusus IAM: `backend/gateway/test/iam/` atau `internal/e2e/`)
     - Manifest deklaratif RBAC eksklusif modul: `backend/shared/rbac/catalogs/{management}.json` (satu berkas per management; murni metadata katalog izin & template role tanpa logika bisnis/ABAC domain).
     - *Aturan:* Developer Management A **dilarang** mengedit, memindah, atau menghapus file Zona 1 maupun manifest RBAC milik Management B.
   - **Zona 2 — File Wiring Bersama (*Minimal-Diff & Additive-Only*)**:
     - `backend/gateway/gqlgen.yml` (pola glob `graphql/**/*.graphql` bersama; hanya menambah entri `autobind:` atau `models:` untuk `{management}` sendiri bila dibutuhkan) dan `backend/gateway/generate.go` (entrypoint/plugin generator bersama; tidak diubah per domain tanpa koordinasi Gateway Core)
     - Berkas hasil generate inti bersama (`backend/gateway/graph/generated/root_.generated.go`, `backend/gateway/graph/generated/schema.generated.go`, dan `backend/gateway/graph/model/schema_models_gen.go`) — tidak pernah diedit manual; hasilkan ulang lewat generator resmi (`cd backend/gateway && go generate ./...`) lalu verifikasi `git diff` hanya mencerminkan `extend type Query/Mutation` atau field milik management sendiri
     - `backend/gateway/graph/resolvers/resolver.go` (`struct Resolver` dan getter adapter/client management sendiri), `backend/gateway/internal/source_connections/registry.go` (+ unit test terkait), `backend/gateway/cmd/main.go`, `backend/gateway/internal/config/`, dan konfigurasi runtime terkait
     - `backend/gateway/graph/resolvers/schema.resolvers.go` (hanya menambah *case* delegasi tipis di `Query.Node` menuju helper/adapter `{management}`, bukan implementasi bisnis)
     - `backend/gateway/graph/resolvers/shared/mapper/` dan `cross_domain_*_resolver.go` (hanya untuk kontrak lintas domain yang disepakati pemilik terkait sesuai `agents/boundary.md`)
     - *Aturan:* Seluruh edit pada Zona 2 wajib **additive/minimal-diff**, dilarang menyentuh atau mematikan/mengubah flag/binding milik management lain, dan dilarang mengaktifkan entry untuk modul scaffold/*PAUSED*.
   - **Zona 3 — Core Keamanan & Perimeter Gateway (Terlarang untuk Wiring Domain Biasa)**:
     - `backend/gateway/internal/server/` (server HTTP Fiber, penegakan endpoint tunggal `/api/graphql`, CSRF binding, rate limit, cookie opak, CORS, redaksi error)
     - `backend/gateway/internal/middleware/` (ekstraksi konteks autentikasi, validasi sesi/principal)
     - `backend/gateway/internal/browserauth/`, `backend/gateway/internal/store/`, serta package akses khusus seperti `internal/hrmaccess/` dan `internal/commerceaccess/` (kecuali atas mandat khusus IAM/owner terkait)
     - `backend/gateway/graphql/schema.graphql` (`Query.health`, `Query.node`, interface `Node`, `PageInfo`, `Scope`, scalar dasar) serta tipe global bersama pada area `graph/model/`
     - *Aturan:* Developer management **dilarang** mengubah konfigurasi HTTP, middleware autentikasi/CSRF, kebijakan cookie, atau mekanisme keamanan inti demi "meloloskan" wiring maupun pengujian domain. Jika wiring membutuhkan kemampuan baru pada Zona 3, eskalasikan ke pemilik IAM/Gateway Core.

2. **Batas Arsitektur, Keamanan & `boundary.md` Saat Wiring**:
   - **Murni Transport & Mapping Tipis**: Resolver/adapter Gateway hanya memvalidasi bentuk input GraphQL, menerjemahkan tipe protobuf ↔ model GraphQL, meneruskan `context.Context` beserta metadata sesi/principal/trace/deadline, serta mensanitasi pesan error internal agar detail SQL/infra tidak bocor. **Dilarang keras** menaruh logika bisnis, query database langsung (PostgreSQL/Redis domain), atau duplikasi katalog permission/ABAC di Gateway ([`agents/users.md`](agents/users.md)).
   - **Fail-Closed & Kesiapan Layanan**: Setiap getter di `source_connections.Registry` wajib *fail-closed* (mengembalikan gRPC `codes.Unavailable` atau error terstruktur bila koneksi `nil`/disabled), dan seluruh operasi wiring wajib melewati `/api/graphql` tanpa menambah route REST baru (`docs/00-governance/decisions.md`).
   - **Larangan Ekspos Scaffold/PAUSED**: Management yang masih berstatus scaffold atau *PAUSED* (misalnya `order_management`, `pricing_revenue_management`, `crm_sales_management`, `customer_service_management`, atau modul yang belum siap implementasi E2E) wajib tetap nonaktif di konfigurasi Gateway dan **tidak boleh** menaruh berkas `.graphql` aktif di bawah `backend/gateway/graphql/` (karena pola glob `graphql/**/*.graphql` akan otomatis memuatnya) maupun dibuatkan package/resolver placeholder kosong sebelum ada keputusan owner.

3. **Alur Kerja Atomik & Gerbang Verifikasi Wiring**:
   - **Prasyarat Sebelum Menyentuh Gateway**: Pastikan microservice & unit test (`backend/services/{management}/`) serta kontrak proto (`backend/contracts/`) sudah lulus kompilasi (`make -C backend proto` / `buf lint`). Cek `git status --short backend/gateway/` terlebih dahulu; bila terdapat perubahan Zona 2 yang belum tuntas milik agen lain, selaraskan dulu sebelum menjalankan generator GraphQL (`cd backend/gateway && go generate ./...`).
   - **Satu Jendela Atomik sampai Hijau**: Setiap sesi wiring wajib menyelesaikan rangkaian `proto → adapter → source_connections/main.go → graphql/{management}/*.graphql → go generate ./... → mapper/resolver` dalam satu jendela kerja utuh hingga `backend/gateway` dapat dikompilasi bersih.
   - **Verifikasi Wajib & Isolasi Commit**: Sebelum handover/commit, jalankan dari `backend/gateway/`: `go fmt ./...`, `go vet ./...`, `go build ./cmd/...`, serta `go test -race` untuk unit test `adapter`/`resolver`/`source_connections` dan suite E2E milik management tersebut. Gunakan `git diff --check` dan stage file secara spesifik (`git add <path-spesifik>`) agar file milik management lain tidak ikut terbawa.

## Database Lokal, Testing per Management, dan Inisialisasi Otomatis (db/init.go)

### Kebijakan Migrasi Skema Pra-Rilis (Keputusan Owner 11 November 2026)

- Selama software **belum dirilis ke publik**, setiap management hanya boleh memiliki **satu versi migrasi skema: `000001`**. Perubahan `backend/services/{service_name}/db/schema/*.dbml`, salinan DBML di service, tabel, kolom, indeks, constraint, atau seed skema harus dimasukkan ke baseline migrasi `000001_*.up.sql` dan diselaraskan dengan `000001_*.down.sql` milik management terkait.

- **Dilarang membuat migrasi bernomor baru** (`000002`, `000003`, dan seterusnya) untuk perubahan skema pra-rilis. Jangan menaikkan nomor migrasi hanya karena ada task, perubahan DBML, atau perubahan versi aplikasi. Nomor migrasi skema berbeda dari versi produk di `backend/VERSION`, Git tag, dan versi kontrak API/protobuf.

- Beberapa service sudah memiliki migrasi `000002` ke atas dari pekerjaan sebelumnya. Saat mengubah skema service tersebut, **jangan menambah versi lagi**; konsolidasikan semua efeknya ke `000001_*.up.sql` dan pastikan `000001_*.down.sql` membatalkan baseline secara konsisten. Setelah memastikan semua efek sudah terwakili, hapus file migrasi bernomor lama dari folder embed agar migrasi baru benar-benar berhenti di versi `000001`. Selaraskan DBML dan inisialisasi yang terdampak. Database development/testing boleh dibuat ulang sesuai kebijakan development di atas.

- Selama masih dalam tahap development, jika perubahan SQL/DBML membuat skema database tidak sesuai dengan baseline `000001`, agen boleh menghapus dan membuat ulang database development atau testing milik management terkait, lalu menjalankan migrasi dari awal. Pastikan nama database dan management yang dituju benar sebelum penghapusan; jangan menyentuh database management lain.

- Owner akan memberi tahu agen saat sistem mulai digunakan di production. Sebelum pemberitahuan tersebut, perlakukan database sebagai development/testing sesuai namanya. Setelah owner menyatakan sistem production, jangan menghapus atau membuat ulang database production; tata cara migrasi production harus ditetapkan terpisah oleh owner.

- Setelah rilis publik pertama, kebijakan migrasi versi berikutnya harus ditetapkan ulang secara eksplisit oleh owner.

- **Manajemen Siklus Hidup & Konvensi Penamaan Database**:

  - Setiap management mengelola koneksi dan lifecycle database melalui engine miliknya di `backend/services/{service_name}/internal/engine/` serta bootstrapper di `backend/services/{service_name}/db/init.go`.

  - Gunakan alias management yang singkat, lowercase, dan konsisten: database local development bernama `{alias}`, sedangkan database testing wajib bernama `{alias}_test`. Contoh: IAM memakai `iam` dan `iam_test`; HRM memakai `hrm` dan `hrm_test`. Management lain mengikuti pola yang sama dan mendokumentasikan aliasnya pada engine terkait.

- **Aturan Isolasi Database Testing (`*_test`)**:

- Seluruh E2E yang memakai basis data PostgreSQL nyata **wajib menggunakan database testing dengan akhiran `_test`** (`{alias}_test`). Jangan membuat suite repository atau integration test terpisah; validasi repository/query nyata melalui alur E2E. E2E tidak boleh menyentuh database production.

  - Test harness dan helper pengujian (misal `testutil.OpenPostgres(t)`) **wajib menerapkan validasi fail-fast**: jika nama database target yang diparse dari DSN/environment tidak berakhiran `_test`, pengujian wajib langsung digagalkan (`t.Fatalf`), bukan peringatan atau fallback.

  - Tes lapisan service tetap murni menggunakan `mock_db` sesuai aturan pengujian service tanpa memanggil database asli.

- **Inisialisasi Otomatis dan Migrasi Skema (`db/init.go`)**:

  - Setiap service wajib menyediakan `backend/services/{service_name}/db/init.go` (`package db`) yang berfungsi sebagai bootstrapper basis data otomatis ketika aplikasi dijalankan (`cmd/server/main.go`) maupun saat test suite dieksekusi.

  - **Portabilitas dengan `embed.FS`**: Seluruh berkas migrasi wajib di-embed ke dalam binary menggunakan standar Go `//go:embed migrations/*.up.sql` sehingga migrasi independen terhadap direktori kerja host.

  - **Auto-Create Database**: Fungsi `EnsureDatabase(dsn string)` wajib mengurai nama database target, membuka koneksi sementara ke database maintenance PostgreSQL (`postgres`), mengecek apakah database target sudah ada (`SELECT 1 FROM pg_database WHERE datname = $1`), dan mengeksekusi `CREATE DATABASE "<target_db>" ENCODING 'UTF8'` jika belum ada. Mekanisme ini berlaku untuk database local development (`{alias}`) maupun testing (`{alias}_test`).

  - **Auto-Migrate Up**: Fungsi `RunMigrationsUp(db *sql.DB)` wajib memastikan tabel pelacak `schema_migrations` (`version BIGINT PRIMARY KEY, dirty BOOLEAN NOT NULL DEFAULT FALSE, applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`) tersedia, memindai seluruh file `*.up.sql` yang ter-embed secara berurutan, dan mengeksekusi migrasi yang belum terpasang di dalam transaksi transaksional (`BEGIN ... COMMIT`).

  - **Fasad Inisialisasi**: Menyediakan fungsi utama `Init(dsn string, isTesting bool) (*gorm.DB, error)` yang memvalidasi kesesuaian nama database (wajib `*_test` jika `isTesting == true`), memastikan database fisik ada, menjalankan migrasi up, dan mengembalikan instance `*gorm.DB` terkonfigurasi.

## Instruksi Lintas Agen

- Sebelum menugaskan management lain, pengirim wajib membaca `agents/arsitektur.md` dan `agents/boundary.md`, mengacu pada keputusan keamanan/ownership yang relevan, serta mencantumkan kedua referensi dalam `## Context` instruksi. Developer penerima juga wajib membaca arsitektur sebelum implementasi. D01–D12 hanya dapat berubah atas keputusan owner.

### Lokasi, nama file, nomor instruksi, dan status penyelesaian (-done.md) (wajib)

- Seluruh penugasan dari leader, developer, atau agen kepada developer **management penerima** wajib ditulis **sebelum handoff** di `agents/instructions/{management}/{id}-{from}-{title-instructions}.md`. Gunakan folder `agents/` (jamak) yang sudah menjadi standar repository; jangan membuat folder paralel `agent/`. Catatan ini tidak menggantikan `agents/context/`, `agents/notes/`, dan working history.
- `{management}` adalah **management yang menerima dan mengerjakan instruksi**, memakai nama direktori service dalam `snake_case` (misalnya `ecommerce_commerce` atau `finance_accounting`); bukan management pengirim. `{id}` adalah bilangan berurutan mulai dari `1`, dihitung **per folder management penerima** berdasarkan ID terbesar yang sudah ada. Jangan menduplikasi, mengubah nomor, atau menimpa instruksi terdahulu.
- `{from}` mengidentifikasi pengirim secara jelas, misalnya `leader_backend_erp` atau `iam_management`. `{title-instructions}` adalah deskripsi tugas yang spesifik, lowercase kebab-case, **diakhiri `-instructions`**. Contoh: `agents/instructions/ecommerce_commerce/1-leader_backend_erp-implementasi-backend-toko-web-instructions.md`. Nama file harus dapat dibedakan tanpa membuka isinya.
- **Aturan Penandaan Instruksi Selesai (`-done.md`)**:
  - **Tugas Belum Selesai (Pending / In-Progress)**: File instruksi yang baru dibuat, sedang dikerjakan, atau belum tuntas tetap mempertahankan akhiran `-instructions.md` (misal: `1-owner-iam-migrasi-skema-customer-oauth-instructions.md`).
  - **Tugas Sudah Selesai (DONE)**: Begitu seluruh tugas, kriteria penerimaan, dan pengujian dalam instruksi telah selesai dikerjakan 100% dan terverifikasi, developer/agen pelaksana **WAJIB menambahkan akhiran `-done.md`** pada nama file tersebut (misal: `git mv 1-owner-iam-migrasi-skema-customer-oauth-instructions.md 1-owner-iam-migrasi-skema-customer-oauth-instructions-done.md`). Hal ini memastikan seluruh tim, agen, dan owner dapat langsung mengetahui secara sekilas mana tugas yang belum dikerjakan (`*-instructions.md`) dan mana tugas yang sudah selesai (`*-done.md`).
- Seluruh instruksi lama sudah dikelompokkan menurut management di `agents/instructions/{management}/`; folder root `agents/instructions/` hanya untuk indeks navigasi. **Instruksi baru wajib memakai struktur per management**. Jangan mengembalikan dokumen instruksi langsung ke root, menggandakan arsip, atau mengganti nomor instruksi yang sudah terpakai.
- Jika dokumen menjadi pedoman untuk beberapa management sekaligus tanpa penerima tunggal, tempatkan satu sumber otoritatif di folder management **pemilik keputusan** (misalnya katalog dan permintaan akses IAM di `iam_management/`) dan berikan tautan atau catatan pada management penerima. Ini bukan izin mengubah kode management lain.

### Isi minimum instruksi (wajib konkret, bukan arahan umum)

Setiap dokumen wajib mandiri dan memuat **empat heading eksplisit: `## Task`, `## Request`, `## Action`, dan `## Context`**. Tulis dalam Bahasa Indonesia dengan istilah teknis yang tepat. Rinci hal-hal berikut sesuai tugas; bila sebuah bagian tidak relevan, jelaskan alasannya, jangan mengarang kemampuan yang belum ada.

- **Task:** latar tujuan bisnis, masalah yang harus diselesaikan, management/domain penanggung jawab, hasil akhir yang bisa diamati, prioritas dan batas scope. Sebut secara tegas apa yang **tidak boleh** dikerjakan karena milik management lain menurut `agents/boundary.md`.
- **Request:** daftar **path file yang sudah ada untuk dibaca/diubah dan path file baru untuk dibuat** (dengan status masing-masing), paket/lapisan yang bertanggung jawab (DBML, migrasi, entity/domain, service, repository, protobuf/gRPC, adapter, GraphQL SDL/model/resolver, konfigurasi, dan test jika relevan), serta nama **function/method/interface/RPC/Query/Mutation** yang wajib dibuat atau diubah dengan package/receiver. Tentukan kontrak input-output, tipe, field wajib/opsional, batas otorisasi dan deliverable yang harus diserahkan. Untuk pekerjaan dokumentasi saja, nyatakan tidak ada function yang perlu ditulis.
- **Action:** urutan implementasi **langkah per langkah** dari satu domain/vertical slice sampai lulus sebelum pindah domain lain. Jelaskan **logika tiap function/use case**, validasi, invariant dan state transition, data yang dibaca/ditulis, skema atau tabel pemilik, transaksi/locking/idempotency, event/outbox, pemanggilan service lain dan timeout/retry, error handling dan rollback, audit, serta penanganan akses yang ditolak. Bedakan aturan yang ditegakkan database, domain service, dan gateway. Jangan hanya menulis “buat CRUD” atau “implementasikan integrasi”; rinci perilaku setiap operasi dan contoh skenario penting. Instruksi berupa uraian operasional, **bukan kode, pseudocode, atau patch**.
- **Context:** keputusan owner, kondisi kode saat instruksi dibuat (sudah ada, stub, belum ada), referensi path dan simbol yang sudah **diverifikasi di repository**, kontrak/DBML/GraphQL/README terkait, pekerjaan sebelumnya, ketergantungan antar-management, asumsi yang masih perlu keputusan pemilik, dan risiko kompatibilitas. Jangan menyatakan integrasi sudah siap hanya karena proto atau foldernya tersedia.

### Kriteria penerimaan, bukti, dan handoff

- Setiap langkah atau kelompok operasi wajib memiliki **definition of done** yang bisa diuji: skenario berhasil dan gagal, otorisasi IAM/role/permission/scope, validasi field dan state, konkurensi, retry, idempotensi, data persisten, dan perilaku saat service dependency tidak tersedia. Sebut path test dan perintah build/generate/lint/vet/unit/race/E2E yang relevan dengan modul; catat ekspektasi hasil dan bagaimana membuktikannya lewat API/GraphQL sampai database ketika diminta.
- Saat ada dependency lintas management, instruksi penerima hanya boleh meminta kontrak dan mengonsumsi fakta pemilik. Tulis data minimum, referensi ID dan versi, bentuk respons/error, serta penanganan timeout. Catat permintaan pada `agents/notes/{management_pemilik}/{id}-{management_peminta}.md` menggunakan struktur lengkap dari [`agents/instructions/instruction_template.md`](agents/instructions/instruction_template.md). **Dilarang** menugaskan developer penerima menambah tabel, mengubah business logic, atau membuat fallback master data di management pemilik tanpa mandat terpisah.
- Tetapkan milestone berurutan dan keluaran yang harus dilaporkan di tiap milestone: perubahan file, keputusan domain, hasil pengujian aktual, blocker beserta management pemilik, dan referensi commit. Pengirim harus memeriksa kecocokan daftar path dan simbol, menghindari referensi palsu, dan memastikan instruksi bisa dijalankan developer **tanpa perlu membaca percakapan asal**. Aturan Git tetap berlaku: bekerja dan push hanya ke `main`, tanpa membuat branch baru atau force-push.
- **Wajib Rename Menjadi `-done.md` Pasca Selesai**: Begitu seluruh kriteria penerimaan, checklist *Action*, dan pengujian (*PASS*) telah tuntas dan commit dibuat, pelaksana **wajib mengubah nama file instruksi terkait dengan menambahkan akhiran `-done.md`** (contoh: `git mv {id}-{from}-{title}-instructions.md {id}-{from}-{title}-instructions-done.md`). Di dalam file instruksi, ubah `**Status:**` menjadi `DONE`, cantumkan SHA commit verifikasi, serta sertakan ringkasan bukti pengujian. Dilarang menandai `-done.md` jika instruksi masih berupa draf, diblokir, atau belum teruji.

## Interkomunikasi, Inspeksi Read-Only (Code & Database), dan Permintaan Perubahan Lintas Management (`agents/notes/`)

Apabila dalam pengembangan suatu management (`{from_management}`) terdapat **kebutuhan layanan atau integrasi lintas management** dengan management lain (`{to_management}`), berlaku aturan interkomunikasi berikut:

1. **Hak Akses *Read-Only* terhadap Source Code Management Lain (Khusus Kebutuhan Integrasi)**:
   - Agen `{from_management}` **diizinkan melakukan pembacaan (*read-only*) terhadap source code milik `{to_management}`** (`backend/services/{to_management}/`, `backend/contracts/proto/{to_management}/`, dan `backend/gateway/graph/resolvers/{to_management}/`) apabila dibutuhkan untuk memahami kontrak, alur service, signature fungsi, struktur entity/repository, mapper, atau handler gRPC terkait integrasi.
   - Gunakan `codebase-memory-mcp` (`search_graph`, `trace_path`, `get_code_snippet`, `get_file_outline`) dan `read_file` untuk menginspeksi kode `{to_management}` secara presisi.
   - **Batas Tegas Code Ownership**: Hak ini murni **sebatas baca (*read-only*)**. Agen `{from_management}` **dilarang keras** mengedit, menambah, atau menghapus file di dalam `backend/services/{to_management}/`, dan kode Go `{from_management}` **dilarang** mengimpor package `internal` milik `{to_management}`.

2. **Hak Akses *Read-Only* terhadap Database & Skema Management Lain (Khusus Kebutuhan Integrasi)**:
   - Apabila dibutuhkan untuk merancang atau memverifikasi integrasi layanan antar-management, agen `{from_management}` **diizinkan memeriksa skema dan database milik `{to_management}` secara *read-only***:
     - Membaca berkas DBML (`backend/services/{to_management}/db/schema/*.dbml`) dan migrasi (`000001_*.up.sql`).
     - Menjalankan query **READ-ONLY (`SELECT`, `\d`, `\dt`)** pada database PostgreSQL milik `{to_management}` di container Docker (`127.0.0.1:6543`) untuk memeriksa struktur tabel, tipe data kolom, enum, constraint, atau memverifikasi state data aktual.
   - **Batas Tegas Database Ownership**:
     - Agen `{from_management}` **dilarang keras** menjalankan operasi tulis/mutasi (`INSERT`, `UPDATE`, `DELETE`, `ALTER`, `DROP`, `TRUNCATE`) pada database `{to_management}`.
     - Kode aplikasi Go `{from_management}` pada saat runtime **tetap dilarang keras** membuka koneksi SQL/`DB_DSN` langsung atau melakukan `JOIN`/foreign key lintas database ke database `{to_management}`. Seluruh komunikasi runtime antar-service wajib melalui kontrak gRPC internal atau event outbox/inbox resmi milik pemilik data.

3. **Kewajiban Menulis Permintaan Perubahan Kode dengan Format `instruction_template.md` di `agents/notes/`**:
   - Jika berdasarkan hasil inspeksi *read-only* atas kode dan database tersebut `{from_management}` membutuhkan penyesuaian atau penambahan kode/kontrak pada `{to_management}` (misal: penambahan RPC gRPC, field Protobuf/skema, publisher/consumer event, atau penyesuaian validasi), maka agen `{from_management}` **wajib** menuliskan detail keinginannya pada:
     `agents/notes/{to_management}/{id}-{from_management}.md`
   - Gunakan ID numerik berurutan berdasarkan ID tertinggi di folder `agents/notes/{to_management}/`, mulai dari `1`. Buat file baru untuk setiap kebutuhan baru; jangan menimpa catatan lama.
   - **Wajib Menggunakan Template [`agents/instructions/instruction_template.md`](agents/instructions/instruction_template.md)**: Isi dokumen di `agents/notes/{to_management}/{id}-{from_management}.md` dilarang ditulis sebagai catatan bebas/abstrak, melainkan **wajib mengikuti seluruh struktur [`agents/instructions/instruction_template.md`](agents/instructions/instruction_template.md)** secara lengkap dan konkret berdasarkan temuan *read-only* sebelumnya:
     - **Metadata Header**: `# {from_management} - {to_management} - {Title Task}`, `Role`, `Priority`, `Status`, `Repository`, `Baseline commit`, `Target management`.
     - **`## Task`**: `Objective`, `Current problem`, `Expected result`, `Scope`, dan `Out of scope` (batas ownership sesuai `agents/boundary.md`).
     - **`## Request`**:
       1. **`1. Files`**: Tabel path file aktual di `{to_management}` (`READ`, `CREATE`, `MODIFY`, `TEST`) yang telah diverifikasi lewat inspeksi *read-only*.
       2. **`2. Functions / Methods / Operations`**: Rincian `F1`, `F2`, dst. (lokasi file, tanggung jawab, parameter `Input`/`Output`, `Required behavior`, `Wiring`, `Error handling`, `Security / authorization`, dan `Required tests`).
       3. **`3. Integration Requirements`**: `Source` (`{from_management}`), `Destination` (`{to_management}`), `Contract` (RPC/event), perilaku integrasi, dan `Ownership restriction`.
     - **`## Action`**: Langkah implementasi (`Step 1` s.d. `Step 4`), tabel skenario pengujian (`TC-01` s.d. `TC-05`), perintah build/test aktual, `Acceptance Criteria`, dan `Required Final Report`.
     - **`## Context`**: Referensi arsitektur (`AGENTS.md`, `agents/arsitektur.md`, `agents/boundary.md`), kondisi implementasi `{to_management}` saat ini dari hasil inspeksi *read-only*, batasan D01–D12, serta dependensi dan risiko.
   - Catatan kebutuhan di `agents/notes/` ini berfungsi sebagai spesifikasi teknis siap-eksekusi (*executable-ready specification*) bagi agen `{to_management}` ketika ditugaskan oleh Leader/Owner, bukan izin bagi `{from_management}` untuk mengubah sendiri kode `{to_management}`.
   - *Pengecualian pendaftaran permission & template role rutin:* begitu mekanisme manifest RBAC aktif, developer `{from_management}` cukup mendaftarkan izin dan role domainnya melalui `backend/shared/rbac/catalogs/{from_management}.json` tanpa membuat catatan permintaan edit `bootstrap.go` ke `iam`; catatan di `agents/notes/iam_management/` (atau `iam/`) hanya dibuat bila membutuhkan perubahan inti IAM seperti prefiks `default_redirect_path` baru, akun otomasi `auto_system_{mgmt}`, atau integrasi `ScopeOwnerVerifier`.

## Struktur Service Berdasarkan Domain

- Implementasi service wajib dikelompokkan dalam folder domain: `backend/services/{service_name}/internal/service/{domain}/service.go`. Contoh IAM: `internal/service/user/service.go`, `internal/service/authentication/service.go`, `internal/service/password/service.go`, dan `internal/service/role/service.go`.

- Setiap folder domain menjadi package Go terpisah berdasarkan import path dan memakai deklarasi `package service`, sesuai keputusan pengguna. Nama folder domain tetap singkat, lowercase, tanpa underscore, misalnya `user`, `oauth`, `session`, atau `roleassignment`. Saat mengimpor beberapa domain, gunakan alias yang jelas seperti `userservice` dan `passwordservice` agar nama package tidak bertabrakan.

- Jangan menumpuk implementasi dalam pola flat `internal/service/{domain}_service.go`. Pola folder domain menggantikan pola tersebut untuk pekerjaan selanjutnya. Jangan membuat salinan implementasi aktif pada kedua lokasi.

- File utama domain bernama `service.go`. File pendukung seperti `types.go` atau `ports.go` boleh ditambahkan dalam folder domain pemiliknya jika diperlukan; jangan mengumpulkan seluruh logika domain pada satu file atau package pusat.

- Root `internal/service/` hanya boleh berisi komposisi/dependency lintas domain yang memang diperlukan, bukan business logic semua domain. Jaga arah dependency dan hindari import cycle antardomain.

- Tes setiap domain berada di `internal/service/{domain}/test/` dengan package `service_test`, mengimpor API publik package domain yang diuji. `mock_db` dan fixture lain milik domain itu tetap di folder test tersebut.

- Helper tetap terpisah dari service, mengikuti lokasi `internal/helper/{domain}_helper.go`; jangan memakai file helper untuk menyembunyikan seluruh orkestrasi bisnis.

- Migrasi file lama dilakukan hanya ketika tugas migrasi/implementasi diminta; perubahan aturan ini tidak otomatis memindahkan atau menghapus kerangka yang sudah ada. Instruksi lintas agen yang masih memakai pola flat harus diselaraskan sebelum dijalankan.

## Input dan Pemetaan Request Service

- Tentukan bentuk parameter berdasarkan struktur data, bukan jumlah parameter. Nilai sederhana seperti ID, email, status, dan reason diterima langsung sebagai parameter function; jangan membuat struct pembungkus yang tidak diperlukan.

- Jika input berupa object terstruktur, definisikan struct sesuai operasi pada `backend/services/{service_name}/internal/service/{domain}/input.go`, dalam `package service` yang sama dengan `service.go`. Buat file ini hanya ketika ada input object yang memang diperlukan.

- Contoh: `SearchUserByID` menerima ID langsung; `SearchUserByEmail` menerima email langsung; `CreateUser` menerima `CreateUserInput`; `UpdateUser` menerima ID langsung dan `UpdateUserInput` untuk object perubahan.

- Gunakan nama struct dan field yang spesifik, tipe yang tepat, serta pembagian field yang mudah dibaca. Bedakan field wajib, opsional, dan patch update tiga keadaan: tidak dikirim, null, atau nilai. `context.Context` tetap parameter tersendiri.

- Handler/adapter memetakan request gateway/protobuf ke parameter atau struct input service. Jangan mengikat input service langsung ke model GraphQL, message protobuf, atau entity ORM. Pemetaan tidak menggantikan validasi bisnis dalam service/helper.

- `input.go` berisi definisi input object, bukan tempat orkestrasi business logic atau seluruh tipe lintas domain. Validasi dan logika tetap mengikuti tanggung jawab service serta helper domain.

## Aturan Desain dan Struktur Protobuf

- **Struktur Folder Per Management dan Domain**:

  - Simpan kontrak di `backend/contracts/proto/{management}/`; buat satu direktori untuk setiap domain, dengan versi sebagai anak direktori (`{domain}/v1/`), mengikuti pola IAM:

     ```text

     {management}/

     ├── common/v1/common.proto

     ├── {domain_a}/v1/{domain_a}.proto

     ├── {domain_b}/v1/{domain_b}.proto

     └── v1/{management}.proto

     ```

  - Buat satu file domain utama bernama `{domain}.proto`. File hanya boleh berisi pesan, enum, request/response, edge, dan payload milik domain tersebut. Jangan menggabungkan beberapa domain dalam satu file.

  - Tidak ada batas jumlah baris untuk file `.proto` domain; file boleh panjang selama tetap fokus pada satu domain yang sama (bukan *mother file* yang mencampur beberapa domain). Pemecahan menjadi beberapa file dalam satu domain (misalnya `{domain}.proto`, `requests.proto`, atau `events.proto`) didasarkan pada tanggung jawab domain, bukan target batas jumlah baris. Semua file pecahan wajib memakai package dan `go_package` domain yang sama; jangan membuat `types.proto` atau `service.proto` sebagai tempat penampungan semua domain.

  - Tipe lintas domain yang benar-benar digunakan bersama diletakkan di `common/v1/common.proto`. Jangan menyalin tipe umum ke setiap domain. File event tetap berada di domain pemiliknya; `v1/events.proto` hanya untuk event yang memang melintasi beberapa domain, bukan tempat penampungan umum.

- **Fasad Service Management**:

  - `v1/{management}.proto` adalah fasad seperti `iam_management/v1/iam_management.proto`: isinya hanya deklarasi package, opsi `go_package`, import file domain, komentar kelompok domain, dan RPC yang merujuk request/response domain.

  - Jangan menaruh definisi `message`, `enum`, payload, atau logika domain di fasad. Jangan menduplikasi definisi request/response di fasad atau mendeklarasikan ulang RPC yang sama pada file domain.

- **Package, Nama File, dan Versioning**:

  - Gunakan `package {management}.{domain}.v1;` untuk file domain, `package {management}.common.v1;` untuk common, dan `package {management}.v1;` untuk fasad. Direktori, package, dan versi harus cocok dengan aturan Buf (`DIRECTORY_SAME_PACKAGE`, `PACKAGE_DIRECTORY_MATCH`, `PACKAGE_VERSION_SUFFIX`).

  - Tetapkan `option go_package` sesuai lokasi output, misalnya `github.com/hackerenesia/erp/contracts/gen/{management}/{domain}/v1;{domain}v1`; jangan memakai satu package Go bersama untuk semua domain.

- **Kontrak dan Kompatibilitas**:

  - ID entitas berupa `string` global opaque. Waktu gunakan `google.protobuf.Timestamp`, objek dinamis `google.protobuf.Value`, biner `bytes`, dan desimal `string`.

  - Update parsial wajib menyediakan `google.protobuf.FieldMask update_mask`. Pagination memakai pesan edge per resource dan `common.v1.PageInfo`.

  - Jangan mengubah nomor atau nama field protobuf yang sudah dirilis. Saat menghapus field, tandai nomor dan nama dengan `reserved`; jangan pernah menggunakannya kembali.

  - Enum memakai prefix nama enum dalam `UPPER_SNAKE_CASE` dan nilai pertama `{ENUM_NAME}_UNSPECIFIED = 0;`.

- **Validasi dan Generate**:

  - Jalankan `cd backend/contracts && buf lint` tanpa warning/error, lalu `make -C backend proto` atau target management terkait, misalnya `make -C backend proto-iam`.

  - Setelah generate, jalankan `go vet ./...` dan tes yang relevan di contracts, gateway, serta service pemilik domain. Jangan mengedit berkas generated secara manual.

## Prinsip Modularitas Kode, DDD, dan Larangan *Mother File* (*Anti God-File*)

- **Larangan *Mother File* / *God File* (Tanpa Batas Jumlah Baris Kode)**:

  - Tidak ada batasan maksimal jumlah baris per file kode; sebuah file **boleh berukuran panjang** selama masih melayani satu domain dan satu tanggung jawab utama yang kohesif.

  - Meskipun tidak ada batas jumlah baris, **dilarang keras membuat *mother file* / *god file* (*monolithic file*)** yang menumpuk kode, struktur data, entitas, atau logika dari banyak domain maupun tanggung jawab yang berbeda dalam satu file.

  - Agen wajib memecah kode berdasarkan **batas domain (DDD)** dan **single responsibility** supaya kode tetap terisolasi, mudah dilacak (**trackable**), dipelihara (**maintainable**), dan diuji (**testable**).

- **Pemisahan Berdasarkan Domain (Domain Boundaries / DDD)**:

  - Seluruh lapisan (Protobuf, Service, Gateway Adapter, Resolver, Repository) wajib mengikuti partisi domain yang jelas (misal: `user`, `group`, `role`, `session`, `permission`).

  - Dilarang mencampur entitas atau logika domain yang berbeda ke dalam satu *mother file* implementasi tunggal.

- **Pemisahan Berdasarkan Tanggung Jawab Teknis**:

  - **Transport vs Adapter**: Lapisan transport murni (koneksi gRPC, TLS, pooling di `source_connections/`) wajib terpisah dari adapter domain (`adapter/{domain}/`).

  - **Mapper & Transformation**: Konversi antar-layer (seperti gRPC Protobuf ke GraphQL atau DB entity ke Domain model) wajib dipisahkan ke file mapper khusus (misal: `iam_mapper.go`), dilarang mencemari resolver atau service dengan fungsi konversi lintas tanggung jawab.

  - **Input Objects**: Struct input terstruktur diletakkan terpisah pada `input.go` di domain terkait, bukan dicampur di `service.go`.

- **Utilitas yang Terfokus (Focused Utilities)**:

  - Dilarang membuat file utilitas "tempat sampah" (*junk-drawer* / *mother utility file*) seperti `utils.go` atau `common.go` tanpa batas tanggung jawab yang jelas.

  - Setiap utilitas atau helper wajib spesifik dan diberi nama sesuai tugasnya (misal: `password_helper.go`, `token_helper.go`, `time_helper.go`).

- **Refaktorisasi Proaktif Berbasis Domain & Tanggung Jawab**:

  - Jika suatu file mulai berkembang menjadi *mother file* yang mengelola lebih dari satu domain atau lebih dari satu tanggung jawab utama (**single responsibility**), agen wajib memecahnya secara terstruktur sebelum menambahkan fitur baru — pemecahan dipicu oleh pencampuran tanggung jawab/domain, **bukan** karena melewati ambang jumlah baris tertentu.

## Aturan Lapisan Repository (GORM)

- **Framework Persistensi**:

  - Repositori wajib menggunakan GORM sebagai persistence framework/ORM untuk interaksi dengan basis data PostgreSQL.

- **Standar Penamaan Method/Fungsi Repository (Database Operations Mapping)**:

  - Setiap fungsi atau method repository wajib mewakili operasi database nyata (SQL/GORM operations) per entitas atau tabel, bukan nama use-case bisnis.

  - **Operasi SELECT / READ**:

    - Menggunakan prefix `Select...`, misalnya `SelectUser`, `SelectUserByID`, `SelectUserByEmail`, `SelectUsers`, `CountUsers`, `CheckUserExistsByID`.

  - **Operasi INSERT / CREATE**:

    - Menggunakan prefix `Insert...` atau `Create...`, misalnya `InsertUser`, `CreateUser`, `BatchInsertUsers`.

  - **Operasi UPDATE**:

    - Menggunakan prefix `Update...`, misalnya `UpdateUser`, `UpdateUserStatus`, `UpdateUserProfile`.

  - **Operasi DELETE**:

    - Menggunakan prefix `Delete...`, misalnya `DeleteUser`, `DeleteUserByID`, `SoftDeleteUserByID`.

  - **Larangan Penamaan Berorientasi Bisnis**: Dilarang menggunakan nama use-case bisnis pada fungsi repository (seperti `RegisterUser`, `PromoteAdmin`, `VerifyUserEmail`, `CheckoutOrder`). Logika dan orkestrasi bisnis adalah tanggung jawab service layer; repository murni mengeksekusi operasi database `Select`, `Insert`/`Create`, `Update`, dan `Delete`.

- **Pemisahan Tanggung Jawab & Isolasi Error**:

  - Repository bertanggung jawab mengelola kueri SQL/GORM, filtering klausul `WHERE`, pagination limit/offset, penguncian baris (`Clauses(clause.Locking{...})`), transaksi (`*gorm.DB`), dan pemetaan error database/driver (seperti `gorm.ErrRecordNotFound` atau constraint violation) menjadi domain error.

  - Repository dilarang melakukan validasi aturan bisnis, otorisasi, enkripsi/hashing (seperti bcrypt), pemanggilan service eksternal, atau mutasi data lintas domain yang tidak terkait.

  - Entity GORM diletakkan di lapisan internal repository atau domain data model, terisolasi dari protobuf message dan model GraphQL.

## Evolusi Document & Records Management

- Keputusan pengguna 9 November 2026: DRM harus mengikuti kebutuhan management/domain yang benar-benar tersedia. DBML tetap rancangan evolutif, bukan arsitektur final atau bukti implementasi/compliance.

- Setiap penambahan management/domain atau perubahan alur yang memerlukan dokumen wajib meninjau inventaris dan matriks kebutuhan DRM: pemilik sumber, jenis bukti, versi yang dirujuk, metadata, IAM/scope, privasi, retensi/trigger, hold, serta penghapusan/arsip.

- Perbarui `backend/services/document_records_management/db/schema/document_records_management.dbml` bila diperlukan, bersama kontrak service/event, adapter, migrasi dan pengujian sesuai tahap implementasi yang diotorisasi. Bila struktur yang ada cukup, catat alasan tanpa memaksakan tabel baru.

- Hubungan lintas management berupa ID logis yang divalidasi service pemilik; jangan membuat FK lintas database atau menyalin master bisnis. Bukti resmi harus menunjuk versi tetap. Melepas tautan tidak otomatis menghapus bukti/kewajiban simpan.

- Profil draft, stub, kontrak yang direncanakan dan integrasi yang telah lulus tes harus dibedakan. Menambah modul tidak boleh mengaktifkan profil produksi atau jadwal retensi hukum tanpa evaluasi.

- Baseline riset dan matriks: `agents/research/4-2026-11-09-drm-standards-module-alignment.md`; inventaris mesin: `agents/audits/drm-module-inventory-2026-11-09.json`. Hitung ulang dari workspace/source saat berubah; angka baseline bukan konstanta arsitektur.

## Akun development IAM dan pendaftaran roles–permissions lintas management

Selama development maupun rilis awal production, operator menyiapkan satu akun human **Super Admin** perdana per instalasi melalui bootstrap resmi IAM ([`backend/services/iam/services/bootstrap.go`](backend/services/iam/services/bootstrap.go)). Identitas login (`IAM_SUPERADMIN_USERNAME` / `IAM_SUPERADMIN_EMAIL`) dan sandi (`IAM_SUPERADMIN_PASSWORD`) tercatat di berkas lokal `backend/services/iam/.env` (di luar Git / *gitignored*, mode privat `0600`), **bukan** di-hardcode ke dalam Git. **Setiap developer management membaca kredensial akun Super Admin development dari `backend/services/iam/.env` saat menjalankan E2E lokal**, serta menjaga isolasi akun otomasi non-interaktif per management (`user_kind = system`, misalnya `auto_system_iam` dan `auto_system_ef` yang dilarang login interaktif). Lihat [pedoman user & akses IAM](agents/users.md).

**Aturan baku agar role & permission tiap management langsung tersimpan ke database IAM saat deploy production (tanpa edit manual DB atau `bootstrap.go`):**

1. **Desain Katalog Domain (`agents/roles-permissions/{management}/1-roles-permissions.md`):** Pemilik management merancang dan mendokumentasikan daftar permission `<management>.<resource>.<action>`, penanda `is_sensitive`, daftar `scope_type` yang valid, *template role* operasional, *default redirect path*, serta aturan pemisahan tugas (SoD). Gunakan [katalog IAM](agents/roles-permissions/iam/1-roles-permissions.md) dan [katalog Enterprise Foundation](agents/roles-permissions/enterprise_foundation/1-roles-permissions.md) sebagai pola acuan.
2. **Manifest Deklaratif Resmi per Management (`backend/shared/rbac/catalogs/{management}.json`):**
   - Setiap management mendaftarkan katalog `permissions[]` dan template `roles[]` terverifikasi (hanya untuk kontrak/operasi domain yang sudah dibangun) ke manifest deklaratif milik modulnya sendiri di `backend/shared/rbac/catalogs/{management}.json` (satu berkas per management, hak edit eksklusif pemilik `{management}`).
   - **Dilarang** mengakses atau menulis tabel database `iam` secara langsung lewat SQL lintas modul maupun membuat role/permission ad-hoc hanya lewat UI/DB lokal tanpa manifest (karena tidak akan ter-replikasi ke database production baru).
   - **Sinkronisasi Otomatis & Aman oleh IAM:** Service `iam` adalah satu-satunya penulis ke tabel `permissions`, `roles`, dan `role_permissions` pada PostgreSQL `iam`, dijalankan secara idempoten saat startup menggunakan *PostgreSQL advisory lock* serta validasi *fail-closed*:
     - **Isolasi Namespace:** manifest `{management}.json` hanya boleh mendefinisikan permission berawalan `{management}.*` dan role dengan `management_domain = "{management}"`, dan **dilarang** mengubah role `superadmin`, domain `iam.*`, atau role otomasi modul lain.
     - **Tanpa Pewarisan Otomatis ke `superadmin`:** domain bisnis baru (`human_resource`, `finance_accounting`, `procurement_sourcing`, `inventory_warehouse_management`, `product_lifecycle_management`, `ecommerce_commerce`, dll.) tidak otomatis mewariskan izin ke role `superadmin` IAM (prinsip SoD), kecuali disetujui eksplisit oleh owner/IAM (baseline `iam.*` + `enterprise_foundation.*` tetap dipertahankan).
     - **Perlindungan Upgrade Production:** permission baru di-*insert* idempoten; permission yang ditarik dari manifest **tidak di-auto-delete** agar role existing/audit tidak rusak; pengetatan `is_sensitive` (`false` -> `true`) memicu invalidasi revisi keamanan, sedangkan penurunan sensitivitas wajib ditolak/dibatasi ketat; *system role* (`is_system = true`, termasuk `{mgmt}_automation`) selalu disinkronkan otoritatif oleh IAM, sedangkan template role operasional manusia (`is_system = false`) memakai kebijakan **seed-once** (*create-if-missing*) agar kustomisasi izin oleh admin perusahaan di production via `/identity` tidak tertimpa saat rilis baru.
     - **Test Anti-Drift:** tiap microservice & adapter Gateway wajib menjaga kecocokan 1:1 antara kode izin yang dipakai di handler/adapter dengan manifest `backend/shared/rbac/catalogs/{management}.json`.
     - *(Status implementasi aktif 25 November 2026: engine pemuat manifest di [`backend/shared/rbac/`](backend/shared/rbac/catalog.go) & [`rbac_catalog_sync.go`](backend/services/iam/services/rbac_catalog_sync.go) sudah aktif untuk `iam.json` dan `enterprise_foundation.json`, disertai test anti-drift pada adapter Gateway IAM & Enterprise Foundation; daftar statis `defaultBootstrapPermissions`/`defaultBootstrapRoles` sudah dihapus.)*
3. **Pemberian Akses Nyata ke User/Group:** Penugasan `user_roles`, `group_roles`, maupun `group_members` kepada akun manusia tetap dilakukan lewat API resmi IAM (`/api/graphql`), tidak lewat manifest statis, dan grant berscope objek eksternal (`legal_entity`, `organization_unit`, `facility`, `vendor_party`, `supplier_party`) tunduk pada validasi pemilik target (`RequireScopeOwnerVerification`).

### Prosedur Standar Wajib E2E Testing RBAC & Session TestAB Lintas Management

Setiap pengujian E2E (End-to-End) untuk query dan mutation GraphQL di seluruh management wajib menjalankan verifikasi real code dengan server gRPC service dan HTTP GraphQL Gateway (`/api/graphql`) aktif menyala, mengikuti 4 tahapan protokol berikut:

1. **Login dengan Super Admin**:
   - Baca kredensial Super Admin dari `backend/services/iam/.env` (`IAM_SUPERADMIN_USERNAME` / `IAM_SUPERADMIN_EMAIL` dan `IAM_SUPERADMIN_PASSWORD`).
   - Eksekusi mutasi `userLogin` ke GraphQL API (`/api/graphql`) dengan pre-auth CSRF untuk mendapatkan cookie sesi opaque (`__Host-integrione_session`) dan token CSRF (`X-CSRF-Token`) milik Super Admin.

2. **Membuat 2 Principals Uji (TestAB: Authorized vs Unauthorized)**:
   - Menggunakan hak akses Super Admin melalui API resmi IAM (memakai katalog permission & role yang sudah tersinkron di DB `iam`), buat/provision 2 akun pengguna (principals) pada database `iam_test`:
     - **User 1 (Authorized / With Permissions):** Diberikan assignment role/group dan permissions yang sah sesuai domain/operasi yang sedang diuji.
     - **User 2 (Unauthorized / Without Permissions):** Dibuat tanpa permissions (atau tanpa role yang berhak mengakses operasi target).

3. **Login dengan Kedua Akun & Ekstraksi Cookie + CSRF**:
   - Lakukan mutasi `userLogin` terpisah untuk User 1 dan User 2 melalui GraphQL API (`/api/graphql`).
   - Simpan cookie sesi opaque (`__Host-integrione_session`) dan token CSRF (`X-CSRF-Token`) dari masing-masing akun untuk digunakan dalam sesi HTTP client terisolasi.

4. **Eksekusi HTTP POST Request ke GraphQL API (`/api/graphql`) untuk Operasi yang Diuji**:
   - Kirimkan HTTP POST request GraphQL yang memuat query atau mutation target dengan 3 skenario wajib:
     - **Skenario A (Authorized User):** Mengirim request dengan Cookie Sesi + CSRF Token milik User 1. **Hasil wajib berhasil** (HTTP 200, field `data` terisi, tanpa error otorisasi/RBAC).
     - **Skenario B (Unauthorized User):** Mengirim request dengan Cookie Sesi + CSRF Token milik User 2. **Hasil wajib tidak berhasil / ditolak** (mengembalikan error otorisasi/permission denied).
     - **Skenario C (Unauthenticated / Anonymous):** Mengirim request tanpa cookie sesi dan tanpa token CSRF. **Hasil wajib tidak berhasil / ditolak** (mengembalikan error unauthenticated / missing CSRF).
   - **Verifikasi Integritas State:** Pastikan seluruh request yang ditolak (Skenario B dan C) **tidak mengubah data di database bisnis** dan **tidak menghasilkan event outbox/inbox baru**.
   - **Dokumentasi Laporan E2E:** Seluruh hasil Skenario A/B/C beserta *edge cases* bisnis tiap `Query` dan `Mutation` yang diuji wajib dicatat dalam laporan `agents/e2e/{management}/{id}-{e2e-title-{Query***/Mutation***}}.md` (lowercase kebab-case pada file aktual) sesuai Seksi 4 dan `## Pengujian E2E dan GraphQL`.
