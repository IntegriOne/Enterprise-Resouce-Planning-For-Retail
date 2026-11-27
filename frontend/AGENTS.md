# Panduan Wajib Frontend Engineers (`frontend/`) — IntegriOne Retail ERP & Web Commerce

Dokumen ini adalah pedoman kerja (*workrule*) dan standar teknis resmi untuk seluruh agen serta developer Frontend yang bekerja di direktori `frontend/`. Dokumen ini menyelaraskan dan menerjemahkan seluruh aturan operasional [AGENTS.md](../AGENTS.md) ke dalam konteks aplikasi Next.js frontend.

---

## 1. Protokol Masuk Sesi (*Session Onboarding*) & Pemulihan Compaction

Sebelum merancang, mengaudit, mengubah kode/UI, maupun menjalankan pengujian — dan wajib diulang saat sesi baru atau setelah *context compaction* — seluruh agen Frontend **wajib** membaca dan mematuhi rujukan berikut:

1. **[AGENTS.md](../AGENTS.md):** pedoman arsitektur repositori, boundary, *CI Gates*, aturan Git `main`-only, dan `## Aturan Khusus Frontend Engineers`.
2. **[docs/development-product-perspektif.md](../docs/development-product-perspektif.md):** perspektif produk **single-tenant brand IntegriOne** pada fase development, aturan *unified `/login`*, peta route resmi (`1 Management ≠ 1 Route`), serta status backend (8 service aktif vs 5 modul tertunda).
3. **[agents/users.md](../agents/users.md):** model satu objek `users` IAM (`customer`, `employee`, `vendor`, `supplier`, `system`), `user_identities`, `sessions` opaque, RBAC–ABAC (`ScopeType` termasuk `self`), prioritas `default_redirect_path`, serta larangan menyimpan data bisnis/grant di tabel `users` (wajib sebelum membangun UI `/identity`, `/profile`, maupun *post-login workspace redirect*).
4. **[docs/README.md](../docs/README.md) & [docs/00-governance/decisions.md](../docs/00-governance/decisions.md):** cakupan **13 management** aktif (**Manufacturing Operations PAUSED**), urutan fase (**F1 internal → F2 vendor & supplier → F3 marketplace/webhook/analitik → F4 customer & calon pembeli terakhir**), serta **Keputusan Owner 13 November 2026 — satu HTTP API `/api/graphql`**.
5. **[agents/arsitektur.md](../agents/arsitektur.md) & [agents/boundary.md](../agents/boundary.md):** keputusan keamanan **D01–D12**, trust boundary IAM/Gateway, serta kepemilikan data antar-management. Jika instruksi/task UI bertentangan dengan D01–D12 atau boundary, catat blocker dan minta keputusan owner — dilarang membuat *bypass* atau UI palsu.
6. **Konteks & Kontrak API Aktif:** baca berkas konteks terbaru Frontend di `agents/frontend/context/core/{id}-context.md` serta konteks management domain yang dikerjakan di `agents/frontend/context/{management}/{id}-context.md` (ID numerik tertinggi), bersama kontrak keamanan/API di [security-and-sessions.md](../docs/02-architecture/security-and-sessions.md) dan [api-contracts.md](../docs/04-engineering/api-contracts.md). Folder `agents/` di luar `agents/frontend/` khusus untuk Backend (`agents/context/`, `agents/working_history/`, `agents/instructions/`, dll.).
7. **Wajib `codebase-memory-mcp` Sebelum Mengubah Kode:** sebelum mengedit file apa pun di `frontend/`, wajib memverifikasi struktur simbol, komponen, hooks, dan kontrak GraphQL Gateway melalui `codebase-memory-mcp` (`search_graph`, `trace_path`, `get_code_snippet`, `check_index_coverage`); lihat Seksi 8.
8. **Protokol Otomatis Pergantian Device (*Device Migration SOP*):** jika pengguna menginformasikan berganti/berpindah device, agen **dilarang bertanya bingung** dan **wajib langsung secara otomatis** menjalankan:
   - Deteksi root repositori aktif (`pwd`) dan pastikan direktori cache `~/.cache/codebase-memory-mcp` (Linux/macOS) atau `%LOCALAPPDATA%\codebase-memory-mcp\cache` (Windows) tersedia dengan format nama `<project-path-normalized>.db` (contoh Linux: `home-muriarfad-Documents-projects-erp-retail.db`).
   - Restore atau jalankan `index_repository` pada `repo_path` aktif, lalu verifikasi kesiapan lewat `list_projects` dan `index_status`.
   - Periksa kesiapan toolchain & container lokal (`node -v`, `npm -v`, `docker compose ps`, healthcheck Caddy `https://erp.localhost:8443/api/graphql?query=%7Bhealth%7D`).
   - Berikan laporan ringkas (1–2 paragraf) dalam Bahasa Indonesia bahwa codebase memory dan environment frontend siap dipakai, lalu langsung lanjutkan tugas.

---

## 2. Bahasa Komunikasi, Tautan & Larangan Umum

- **Bahasa Wajib:** seluruh interaksi pengguna, ringkasan, catatan konteks, *working history*, audit, benchmark, instruksi, dan laporan E2E wajib menggunakan **Bahasa Indonesia**.
- **Standar Tautan:** gunakan tautan Markdown relatif repositori yang valid pada dokumen kanonik. Di percakapan/checkout lokal, tautan `file://` boleh digunakan untuk menunjuk path lokal aktual; jangan menaruh path absolut mesin tertentu ke dalam dokumen kanonik Git.
- **Larangan Umum:**
  - Dilarang keras menyimpan password, token, webhook secret, atau API key (termasuk Xendit/Midtrans) di kode, Git, maupun variabel `NEXT_PUBLIC_*`.
  - Dilarang menginstal atau menyalakan PostgreSQL/Redis langsung di OS host Windows (wajib container Docker).
  - Dilarang menduplikasi master data atau logika otorisasi bisnis dari backend ke state/storage client.

---

## 3. Stack Teknis, Referensi Package Aktual & Konfigurasi Terverifikasi

Sesuai [`package.json`](package.json), [`package-lock.json`](package-lock.json), [`next.config.ts`](next.config.ts), [`tsconfig.json`](tsconfig.json), [`eslint.config.mjs`](eslint.config.mjs), [`docker-compose.yml`](docker-compose.yml), dan [`Dockerfile`](Dockerfile):

- **Runtime Container:** Ubuntu `26.10` (`ubuntu:26.10`), **Node.js `26.x`** + `libatomic1` & `libstdc++6`, berjalan sebagai `USER 10001:10001` dengan dua replika `frontend-1` (`127.0.0.1:3333:3000`) & `frontend-2` (`127.0.0.1:3334:3000`) di belakang reverse proxy `caddy` (`https://erp.localhost:8443`).
- **Framework:** **Next.js `16.4.6`** (App Router di `src/app/`), **React `19.3.0`**, **TypeScript `^5`** (`strict: true`, `noEmit: true`, alias `@/* -> ./src/*`).
- **Compiler & Fitur Next 16:** `reactCompiler: true` (`babel-plugin-react-compiler` `1.0.0`), `cacheComponents: true`, `partialPrefetching: true`.
- **Styling:** **Tailwind CSS `4.3.3`** + `@tailwindcss/turbopack` `4.3.3` (dikonfigurasi lewat `turbopack.rules["*.css"]`) serta **CSS Modules (`*.module.css`)** untuk isolasi gaya kompleks.
- **Animasi & Motion:** **`gsap` `^3.15.1`** + **`@gsap/react` `^2.1.2`** (`useGSAP`), berlisensi GSAP Standard (*no-charge*, mencakup core + `ScrollTrigger`, `SplitText`, dll.).
- **Dependensi & Lockfile:** [`package-lock.json`](package-lock.json) terkomit di Git sehingga [`Dockerfile`](Dockerfile) memakai `npm ci --no-audit --no-fund` secara deterministik; `node_modules/` dan `.next/` tetap diabaikan di [`.gitignore`](.gitignore) serta [`.dockerignore`](.dockerignore).
- **Linting:** ESLint `^9` + `eslint-config-next` `16.4.6` (`core-web-vitals` + `typescript`).
- **Tipografi & Bahasa (`src/app/layout.tsx`):** `<html lang="id">`; font Google resmi via `next/font/google`: `Fraunces` (`--font-fraunces`), `Source_Sans_3` (`--font-source-sans-3`), `Manrope` (`--font-manrope`), `Geist` (`--font-geist-sans`), dan `Geist_Mono` (`--font-geist-mono`).

### Kewajiban Membaca Definisi API Package Aktual (Setara `go doc`)
- Sebelum memakai API, opsi konfigurasi, atau tipe baru dari `next`, `react`, `@gsap/react`, atau plugin `gsap` (`ScrollTrigger`, `SplitText`, dll.), **wajib memeriksa deklarasi tipe `.d.ts` dan implementasi aktual di `frontend/node_modules/`** (contoh: `node_modules/@gsap/react/types/index.d.ts`, `node_modules/gsap/types/`, `node_modules/next/dist/`).
- Periksa signature, parameter opsional/wajib, tipe ref/SSR safety, dan perilaku cleanup sebelum menulis kode; jangan menebak dari ingatan atau versi lama.

---

## 4. Perspektif Produk, Peta Route Resmi (`1 Management ≠ 1 Route`) & Larangan Fake UI

1. **Identitas Brand & Halaman `/login`:**
   - Aplikasi adalah website Web Commerce + sistem operasional internal milik **1 perusahaan ritel** (saat dev diwakili brand **IntegriOne**), **bukan** platform pemasaran SaaS multitenant yang menjual software ERP.
   - `/login` bersifat **umum (satu pintu masuk)** untuk Customer Web Commerce, Internal Employee, Vendor CMT, dan Supplier.
   - **Dilarang** membuat tab/pilihan jenis user (*Internal / Vendor / Supplier / Customer*) sebelum login, dilarang memakai istilah eksklusif internal backoffice (*"Portal IAM"*, *"Enterprise Resource Planning SaaS"*), dan dilarang menawarkan checkbox *"Ingat saya 30 hari"* (kebijakan IAM: opaque session idle 30 menit, absolut 8 jam).
2. **Peta Dashboard Resmi (`1 Management ≠ 1 Route`):**
   - **Prioritas aktif saat ini:** **Dashboard IAM (`/identity`)** dan **alur autentikasi `/login`**.
   - **Rute umum terautentikasi:** `/profile` (akun dasar, keamanan password/MFA/TOTP/sesi aktif, dan ringkasan peran terisolasi).
   - **Rute internal F1 (saat backend/IAM siap):** `/human-resource` (menyerap struktur perusahaan Foundation), `/procurement` (menyerap master rekanan Foundation), `/warehouse`, `/finance`, `/accounting`, `/product`.
   - **Portal eksternal & Web Commerce:** `/supplier` (pemasok bahan mentah, F2), `/vendor` (perusahaan CMT pengolahan produk jadi, F2), dan Web Commerce (`/`, `/account`, `/commerce`, F3–F4).
   - **Management tanpa route dashboard mandiri:** `enterprise_foundation` dan `document_records_management` (DRM) diserap sebagai sub-menu operasional; modul tertunda (`pricing_revenue_management`, `order_management`, `crm_sales_management`, `customer_service_management`, `document_records_management`) **dilarang** dibuatkan UI/navigasi atau API palsu sebelum backend & kontrak GraphQL-nya benar-benar aktif.

---

## 5. Kontrak API Tunggal (`/api/graphql`), Opaque Auth, CSRF & Keamanan Browser

1. **Hanya Same-Origin `/api/graphql`:**
   - Seluruh operasi fetch data, bootstrap CSRF, login/MFA/logout, reset password, dan mutasi domain wajib mengarah ke same-origin **`/api/graphql`** dengan `credentials: "include"`.
   - [`../proxy/Caddyfile`](../proxy/Caddyfile) memblokir seluruh `/api/*` selain `/api/graphql` dengan HTTP 404 (pengecualian terbatas hanya endpoint inbound webhook server-to-server Payment Gateway pada fase F3/F4). Jangan pernah memanggil path REST legacy `/api/auth/*`, port gRPC backend, ataupun API Xendit/Midtrans langsung dari kode client browser.
2. **Alur Opaque Session & CSRF Header Binding:**
   - Sesi browser disimpan di cookie `HttpOnly` `__Host-integrione_session` yang dikelola oleh Gateway/IAM melalui header `Set-Cookie`.
   - Bootstrap pre-auth dilakukan lewat `GET /api/graphql?query=...` (mis. `browserSession { authenticated csrfToken }`), dan setiap `POST /api/graphql` untuk mutasi wajib mengirim header `X-CSRF-Token` terbaru.
   - **Dilarang keras** menyimpan token sesi, JWT, status admin, role, atau daftar permission di `localStorage`, `sessionStorage`, `document.cookie`, atau URL query string.
3. **Fail-Closed & Zero Client Authorization Trust:**
   - Penyembunyian elemen UI di browser hanya untuk kenyamanan UX; keputusan akses selalu berasal dari server (`__Host-integrione_session` + RBAC–ABAC).
   - Redirect pasca-login hanya boleh mengikuti nilai otoritatif server yang telah divalidasi terhadap **allowlist rute internal aplikasi** (mencegah *open redirect* dan skema `javascript:`/`data:`).
   - Saat sesi invalid/kedaluwarsa (`UNAUTHENTICATED`), bersihkan state sensitif di memori client dan arahkan ke `/login`; saat `FORBIDDEN`, tampilkan state akses ditolak yang jelas tanpa memicu retry loop tanpa batas.
4. **Anti-XSS, Anti-IDOR DRM, & Kredensial Dev Lokal:**
   - Dilarang memakai `dangerouslySetInnerHTML`, `eval`, `new Function`, atau menyuntikkan HTML mentah dari server/user input.
   - Untuk dokumen DRM (kelak: slip gaji karyawan, invoice customer, RFQ/PO/CMT): jangan membangun tautan statis/tebakan berbasis UUID tanpa otorisasi server, dan pastikan halaman customer/mitra tidak pernah memiliki akses maupun tautan ke dokumen internal perusahaan.
   - Untuk pengujian login admin lokal terhadap live IAM, baca kredensial dev (`IAM_ROOT_USERNAME`/`IAM_ROOT_EMAIL` dan `IAM_ROOT_PASSWORD`) secara lokal dari berkas gitignored `backend/services/iam_management/.env`; **dilarang keras** men-hardcode atau meng-commit kredensial tersebut di dalam kode `frontend/`.

---

## 6. Arsitektur Next.js App Router, React 19 Compiler, GSAP & Prinsip Anti God-File

1. **Batas Server Component vs Client Component:**
   - Gunakan **Server Component** secara default untuk `page.tsx`, `layout.tsx`, metadata, dan komposisi shell.
   - Tambahkan `"use client"` **hanya** pada komponen daun interaktif (mis. `*-experience.tsx`, formulir, kontrol filter/modal, komponen beranimasi GSAP, hooks) agar ukuran JS client tetap minimal.
   - **Dilarang** mengimpor atau memanggil `gsap` / `@gsap/react` di dalam Server Component (GSAP membutuhkan DOM browser dan tidak aman saat server render murni).
2. **Aturan React Compiler (`reactCompiler: true`), Caching (`cacheComponents: true`) & `useGSAP`:**
   - Tulis komponen dan custom hooks secara murni (*pure render*): jangan memutasi props, array/object state, `ref.current`, atau memicu tween GSAP di dalam render body; jangan melanggar *Rules of Hooks*.
   - **Standar Animasi `@gsap/react`:** Wajib gunakan **`useGSAP(() => { ... }, { scope: containerRef, dependencies: [...] })`** (beserta **`contextSafe(...)`** dari `useGSAP({ scope: containerRef })` untuk animasi yang dipicu event *click/hover/submit*), bukan `useEffect` polos tanpa `gsap.context().revert()`.
   - `useGSAP()` otomatis menjalankan *revert* tween, timeline, dan `ScrollTrigger` saat komponen unmount atau navigasi rute Next.js App Router, serta memakai isomorphic layout effect aman-SSR.
   - Daftarkan plugin GSAP (mis. `gsap.registerPlugin(useGSAP, ScrollTrigger)`) di tingkat modul `"use client"` (atau helper client terpusat), bukan berulang-ulang di dalam render/callback.
   - Karena `cacheComponents: true` dan `partialPrefetching: true` aktif, data yang membaca sesi pengguna, `cookies()`, `headers()`, data finansial/AR/AP, atau data per-aktor **wajib bersifat dinamis per-request** (gunakan `cache: "no-store"` pada fetch sensitif / client fetch) agar data satu pengguna tidak pernah bocor ke cache halaman pengguna lain.
3. **Prinsip Modularitas & Anti God-File (*Single Responsibility*):**
   - **Dilarang** menumpuk UI, GraphQL query strings, mappers, state logic, validasi, dan helper panjang di satu file raksasa (**god component / god file**).
   - Jadikan **`300–500` baris** sebagai batas tinjauan proaktif: bila file mulai melewati batas ini atau mengelola >1 tanggung jawab utama, pecah lebih dulu sebelum menambah fitur baru.
   - Partisi berdasarkan domain route (`src/app/{route}/`) dan modul bersama terfokus (`src/lib/`, `src/components/`):
     - Komponen presentasi vs *client experience container*
     - Typed GraphQL client, query/mutation operations, dan mapper respons (`src/lib/graphql/...`)
     - Validasi formulir & pemetaan error kode gateway terpisah
     - Helper motion GSAP reusable terpisah
   - **Larangan File "Junk-Drawer":** dilarang membuat `utils.ts`, `helpers.ts`, atau `common.ts` raksasa tanpa batas tanggung jawab; beri nama spesifik sesuai tugas (mis. `csrf-client.ts`, `redirect-allowlist.ts`, `currency-format.ts`).
   - **Pemisahan File Tes:** letakkan unit/component tests di folder `test/` atau `__tests__/` domain terkait, jangan dicampur di samping definisi route/runtime.
   - **Jujur terhadap State Asinkron & Analitik:**
     - Pada checkout / Payment Gateway (`PENDING` vs `PAID` vs `EXPIRED` vs `FAILED`), UI dilarang menampilkan status *"Berhasil"* atau *"Lunas"* sebelum fakta owner dari server mengonfirmasi `PAID`/`CONFIRMED`.
     - Pada tabel/grafik analitik, tampilkan `N/A` bila metrik belum tersedia, jangan menggantinya dengan angka `0` palsu.

---

## 7. Design System IntegriOne, Standar Motion (GSAP) & Aksesibilitas (WCAG 2.2 AA)

1. **Identitas Visual Enterprise Editorial IntegriOne:**
   - Gunakan palet resmi: Primary Navy `#133554`, Deep Navy `#0B2439`, Soft Navy `#E9EEF2`, Accent Coral `#DA414A`, Off White `#F9FAFB`, Dark Text `#17202A`, Gray `#66717C`, Light Gray `#DDE3E8`, Navbar `#F4F7F8` (ref token: [design-system-and-usability.md](../docs/05-experience/design-system-and-usability.md)).
   - Gunakan tipografi resmi dari `src/app/layout.tsx`: **`Fraunces`** (`--font-fraunces`) untuk editorial/display headline, **`Source Sans 3`** (`--font-source-sans-3`) untuk teks UI, label, input, tombol, dan navigasi, serta **`Geist Mono`** (`--font-geist-mono`) untuk kode/ID/angka monospasi bila diperlukan.
   - Gunakan CSS Modules (`*.module.css`) dengan root class ter-scope atau utility Tailwind v4 secara bersih; pastikan aturan CSS global di [`globals.css`](src/app/globals.css) tidak merusak halaman lain.
2. **Standar Motion & Animasi GSAP Sesuai Karakter Aplikasi:**
   - **Dashboard Operasional Internal & Portal Mitra** (`/identity`, `/profile`, `/human-resource`, `/procurement`, `/warehouse`, `/finance`, `/accounting`, `/product`, `/supplier`, `/vendor`): pertahankan gerak **fungsional, cepat (`~150–250ms`), dan subtil** (transisi drawer/panel, tab, state feedback); dilarang memakai *scroll-jacking*, *intro splash*, atau efek dekoratif berlebih yang memperlambat tugas operator.
   - **Web Commerce Storefront & `/login`** (`/`, `/login`, `/commerce`, F4): boleh memakai koreografi editorial lebih kaya (`ScrollTrigger`, `SplitText`, transisi hero/galeri) selama tetap menjaga CLS `<= 0.1`, LCP cepat, dan interaksi/form tidak pernah terblokir animasi masuk.
   - **GPU-Only & Anti-Layout Thrashing:** hanya animasikan properti ramah compositor (`transform` / `x`, `y`, `scale`, `opacity`, atau `clipPath` terukur); **dilarang** menganimasikan `top`, `left`, `width`, `height`, `margin`, atau `padding` di tabel/grid besar.
   - **Wajib `prefers-reduced-motion` via `gsap.matchMedia()`:** bungkus animasi GSAP dengan `gsap.matchMedia()` pada kondisi `(prefers-reduced-motion: no-preference)` dan pastikan skenario `reduce` langsung merender kondisi akhir yang jelas terbaca tanpa gerakan bermasalah.
3. **Standar Responsivitas & Aksesibilitas Wajib:**
   - Layout wajib bebas *horizontal overflow* mulai viewport mobile `320px` hingga desktop `1440px+` serta nyaman saat browser di-zoom hingga `200%`.
   - Wajib memiliki HTML semantik (`<header>`, `<nav>`, `<main>`, satu `<h1>` per halaman), `htmlFor` terikat ke `id` input, `aria-describedby` / `aria-live="polite"` untuk validasi & error form, `focus-visible` ring jelas, target sentuh interaktif minimal `44x44px`, label status berteks (tidak hanya mengandalkan warna), dan hormati `@media (prefers-reduced-motion: reduce)` pada CSS maupun `gsap.matchMedia()` di JS.
   - Sediakan seluruh state wajib UI: *loading*, *empty*, *validation error*, *permission denied (403)*, *service unavailable*, dan *idempotent/retry* agar tidak memicu double-submit saat tombol diklik berkali-kali.

---

## 8. Penggunaan Wajib `codebase-memory-mcp` & Interkomunikasi Lintas Management

### A. Alur Wajib `codebase-memory-mcp` (Sebelum Mengubah Kode & Pasca-Selesai)
Seluruh repositori terindeks sebagai **satu project** `<project-path-normalized>` (contoh Linux: `home-muriarfad-Documents-projects-erp-retail`, cek via `list_projects` / `index_status`).
1. **Konfirmasi & Eksplorasi Graph:** `list_projects` & `index_status` -> `search_graph` (komponen, hooks, fungsi, operasi GraphQL Gateway) -> `trace_path` (`inbound` / `outbound` / `both`) -> `get_code_snippet` / `get_file_outline` -> `check_index_coverage`.
2. **Wajib Verifikasi Coverage & Fallback Presisi:** bila `check_index_coverage` melaporkan file `.tsx`/`.ts`/`.css` bersifat `parse_partial`, `skipped`, `stale`, atau non-AST (seperti `*.module.css`, `.json`, `Dockerfile`, `.graphqls`), baca langsung lewat `read_file` atau `rg` sebelum menarik kesimpulan.
3. **Evidence Tiers & Subagent:** patuhi tier `Scout` (provisional, dilarang klaim negatif), `Verify` (default), dan `Auditor` (cakupan penuh + `scopes`). Jika memanggil subagent, jalankan query graph dan coverage di parent terlebih dahulu lalu sertakan hasilnya di prompt subagent.
4. **CLI Fallback:** jika koneksi MCP bermasalah, gunakan terminal CLI `codebase-memory-mcp cli ...` dengan flag `--<nama-parameter-kebab-case>` (contoh `--project`, `--name-pattern`, `--repo-path`, `--mode full`).
5. **Re-Index Pasca-Perubahan:** wajib jalankan `index_repository` (`mode: "full"`, `repo_path: "<abs-path-repositori>"`) setelah menyelesaikan penambahan/refaktor fitur frontend agar graf kode selalu sinkron.

### B. Hak Baca *Read-Only* Backend & Permintaan Kontrak (`agents/notes/`)
- **Izin Baca (*Read-Only*):** Agen Frontend diizinkan membaca berkas backend secara *read-only* — khususnya `backend/gateway/graph/schema/**/*.graphqls`, `backend/gateway/graph/model/`, serta `backend/services/{management}/README.md` — untuk memastikan nama Query, Mutation, tipe input/output, enum, serta error code 100% akurat.
- **Batas Tegas Ownership:** Agen Frontend **dilarang keras** mengubah file Go (`backend/`), `.graphqls`, `.proto`, skema DBML/SQL, `proxy/Caddyfile`, atau menyentuh database management lain.
- **Permintaan Perubahan ke Management Lain:** Bila UI membutuhkan Query/Mutation baru, field tambahan, atau pendaftaran permission/role baru di Gateway atau IAM/management pemilik, **wajib** tulis dokumen permintaan terstruktur di:
  `agents/notes/{to_management}/{id}-frontend.md`
  menggunakan struktur lengkap [`agents/instructions/instruction_template.md`](../agents/instructions/instruction_template.md) (`## Task`, `## Request` berisi path/operation aktual, `## Action`, `## Context`), dengan `{id}` numerik berurutan per folder tujuan.

---

## 9. Strategi Pengujian Frontend (2 Lapisan), E2E RBAC TestAB & Pelaporan

Mengikuti strategi pengujian dua lapisan repositori pada [AGENTS.md](../AGENTS.md):

1. **Lapisan 1 — Static Check, Unit/Helper Tests & Performance Benchmark:**
   - Jalankan typegen, lint, dan compiler check pada `frontend/`:
     ```bash
     cd frontend && npx next typegen && npm run lint && npx tsc --noEmit
     ```
     (tambahkan `&& npm run build` saat verifikasi bundle penuh di host).
   - Unit/component test untuk helper domain murni (validasi form, allowlist redirect, formatter mata uang/waktu, mapper state) disimpan di folder `test/` atau `__tests__/` domain terkait.
   - **Benchmark Frontend (`agents/frontend/benchmark/{management}/`):** untuk alur render/animasi GSAP berat atau optimasi bundle kritis yang diukur kinerjanya, dokumentasikan hasil pengukuran nyata pada `agents/frontend/benchmark/{management}/{id}-{benchmark-title}.md` (lowercase kebab-case; cantumkan skenario, skrip/alat ukur, metrik nyata sebelum/sesudah, dan batas kondisi pengukuran).

2. **Lapisan 2 — E2E Browser Nyata terhadap Live Gateway + IAM + Redis + Postgres `*_test` (`TestAB` 3 Skenario):**
   - Setiap fitur/halaman yang terintegrasi ke `/api/graphql` wajib diuji nyata lewat origin HTTPS Caddy (`https://erp.localhost:8443`) terhadap backend & database `*_test` aktif mengikuti protokol **TestAB**:
     - **Skenario A (Authorized — User 1):** login akun berizin -> halaman dan setiap operasi `Query`/`Mutation` yang dipakai UI berhasil (HTTP 200, `data` valid, state DB/UI terbarui sesuai ekspektasi, redirect mengikuti allowlist).
     - **Skenario B (Unauthorized / Wrong Scope / Anti-IDOR — User 2):** login akun tanpa izin / lintas party -> UI menampilkan status akses ditolak (`403` / empty fail-closed), tidak membocorkan data/dokumen DRM aktor lain, tidak memicu loop retry, dan **zero write** ke backend.
     - **Skenario C (Unauthenticated / Anonymous / Invalid CSRF):** akses rute terproteksi tanpa cookie `__Host-integrione_session` atau kirim mutasi dengan `X-CSRF-Token` invalid -> ditolak fail-closed / diarahkan ke `/login`, state memori client bersih.
     - **Skenario UI & Aksesibilitas Wajib:** validasi field & `aria-live`, pencegahan *double-submit* saat loading, layout responsif `320px–1440px+` + zoom `200%`, serta verifikasi visual `prefers-reduced-motion` (on/off) untuk GSAP & CSS.
   - **Wajib Simpan Laporan E2E (`agents/frontend/e2e/{management}/`):** dokumentasikan hasil pengujian ke `agents/frontend/e2e/{management}/{id}-{e2e-title-{Query***/Mutation***}}.md` (lowercase kebab-case pada berkas aktual yang mencantumkan operasi `Query`/`Mutation` terkait, mis. `agents/frontend/e2e/iam_management/1-login-mutation-user-login-and-mfa.md`). Wajib memuat daftar operasi GraphQL, prasyarat runtime aktif, commit SHA, perintah/skenario uji, serta tabel matriks seluruh cases (`PASS`/`FAIL`/`BLOCKED`).

---

## 10. Docker Single-Build, *Build After Development*, 5 CI Gates & Konkurensi Multi-Agen

### A. Disiplin Satu Build (*Single-Build Discipline*) & Kebersihan `buildx`
- Pada [`frontend/docker-compose.yml`](docker-compose.yml), blok `build:` **hanya boleh didefinisikan pada `frontend-1`**, sedangkan `frontend-2` wajib memakai tag `image:` yang sama dengan `pull_policy: never` — dilarang memicu kompilasi ganda.
- Jika selama debugging terdapat record build gagal (`Error`) atau duplikat, bersihkan riwayat (`docker buildx history rm <id>` / `docker image prune -f`) sehingga kondisi akhir `docker buildx history ls` bersih (hanya 1 record `Completed` per komponen yang dibangun).

### B. Kewajiban *Build After Development*
Setiap kali selesai mengubah kode aplikasi, konfigurasi, atau `Dockerfile` di `frontend/`, agen **dilarang menyatakan tugas selesai** sebelum menjalankan:
```bash
cd frontend && npx next typegen && npm run lint && npx tsc --noEmit && npm run build
```
Lalu dari root repositori:
```bash
docker compose config --quiet
docker compose build frontend-1
docker compose up -d --wait frontend-1 frontend-2 caddy
docker compose ps
docker buildx history ls
```
Pastikan container `frontend-1`, `frontend-2`, dan `caddy` (serta dependencies backend terkait jika aktif) berstatus `Up` dan `(healthy)`, serta uji respons HTTPS melalui Caddy (`https://erp.localhost:8443/api/graphql?query=%7Bhealth%7D` -> `{"data":{"health":"ok"}}` dan muat halaman UI target tanpa console/network 500 error).

### C. 5 Gerbang CI (*Pre-Push CI Gates*) Frontend
Sebelum commit dan push ke `origin/main`, pastikan 5 gerbang berikut berurutan terpenuhi:
1. **Gate 1 — Static, Lint, Typegen & Build:** `npx next typegen && npm run lint && npx tsc --noEmit && npm run build`.
2. **Gate 2 — Helper Tests & Contract Match:** jalankan tes helper/komponen (bila ada) dan cocokkan payload GraphQL dengan `.graphqls` aktual Gateway.
3. **Gate 3 — Browser E2E RBAC TestAB:** verifikasi Skenario A/B/C + UI/a11y/`prefers-reduced-motion` lewat Caddy terhadap live Gateway/IAM (`*_test`) bila menyentuh alur auth/domain aktif.
4. **Gate 4 — *Build After Development* & Clean Build History:** `docker compose build frontend-1`, `up -d --wait frontend-1 frontend-2 caddy`, `docker compose ps` healthy, `docker buildx history ls` bersih.
5. **Gate 5 — `codebase-memory-mcp` Sync & Git Hygiene:** `index_repository` (`mode: "full"`), `git diff --check`, `git add <path-spesifik>`, commit detail, dan `git push origin main`.

### D. Protokol Konkurensi Multi-Agen (Frontend ↔ Backend)
- Sebelum menjalankan `docker compose build` atau menguji integrasi ke `/api/graphql`, periksa `git status --short backend/gateway/`.
- Jika ada agen backend/Gateway sedang mengerjakan perubahan uncommitted di `backend/gateway/`:
  - **Dilarang keras** menjalankan `git checkout --`, `git restore`, `git reset`, atau mengubah file backend mereka.
  - Tetap buktikan kebersihan scope Frontend secara terisolasi (`cd frontend && npx next typegen && npm run lint && npx tsc --noEmit && npm run build`).
  - Jika operasi GraphQL target di Gateway belum tersedia atau masih dikerjakan backend, jangan memalsukan data dummy sukses di UI produksi; koordinasikan atau laporkan blocker secara faktual hingga kontrak Gateway siap kompilasi.

---

## 11. Aturan Git (`main`-only), Prosedur Commit/Push & Konvensi Handover (`agents/`)

### A. Aturan Git & 7 Langkah Wajib (`main` Tunggal)
Seluruh developer dan agen bekerja **langsung di branch `main`** (`origin/main`). **Dilarang keras** membuat branch baru (lokal/remote), PR baru, `git push --force`, `--force-with-lease`, `git reset --hard`, atau `git clean -fd`.
1. **Sebelum mengubah file:** cek `git status --short --branch`, `git remote -v`, `git fetch origin`. Jalankan `git pull --ff-only origin main` hanya jika working tree bersih dan tidak ada commit lokal tertunda (jangan pernah overwrite/stash pekerjaan agen lain).
2. **Selama bekerja:** ubah hanya file dalam scope `frontend/` dan berkas tracking Frontend di `agents/frontend/`. Dilarang menaruh berkas kerja Frontend ke folder Backend (`agents/context/`, `agents/working_history/`, `agents/instructions/`, `agents/audit/`, `agents/benchmark/`, `agents/e2e/`). Koordinasikan terlebih dahulu jika perlu menyentuh file bersama (`AGENTS.md`, `docker-compose.yml`, `proxy/Caddyfile`).
3. **Sebelum commit:** periksa `git diff --check`, `git status --short`, dan `git diff`. Gunakan **`git add <path-spesifik>`** — **dilarang** `git add -A` atau `git add .` bila ada perubahan agen lain. Jangan pernah commit `node_modules/`, `.next/`, `.env*`, token, atau berkas sementara.
4. **Format Pesan Commit:** gunakan `type(frontend/{route-atau-domain}): ringkasan`, dengan body commit menjelaskan logika perubahan, cakupan file, dan bukti verifikasi nyata (`next typegen`, `lint`, `tsc`, `build`, `docker`, E2E).
5. **Sebelum push:** `git fetch origin`; bila `origin/main` maju dan hanya tertinggal tanpa commit lokal, `git pull --ff-only origin main`; bila ada commit **lokal yang belum pernah dipush**, lakukan `git rebase origin/main` di atas working tree bersih, selesaikan konflik secara cermat tanpa mengubah history publik, lalu verifikasi ulang.
6. **Push & Bukti Handoff:** jalankan `git push origin main`, konfirmasi SHA commit aktual pada `origin/main`, dan laporkan daftar file berubah, hasil tes jujur, blocker, serta path handover. Jangan menyatakan sudah dipush hanya dari commit lokal.
7. **Saat gagal / perlu koreksi:** buat commit koreksi baru setelah analisis dampak dan re-test; jangan menulis ulang riwayat `main` atau menyembunyikan tes gagal.

### B. Pemisahan Tegas FE vs BE & Standar Direktori `agents/frontend/`
Seluruh artefak kerja Frontend wajib **terpisah dari Backend** di bawah root **`agents/frontend/`** per `{management}` (`snake_case` nama domain modul seperti `iam_management`, `ecommerce_commerce`, `finance_accounting`, `human_resource_management`, `procurement_sourcing`, atau **`core`** untuk fondasi/workrule lintas aplikasi Frontend). Penomoran `{id}` dimulai dari `1` berurutan **per folder `{management}`**:

- **Instruksi Tugas Frontend (`agents/frontend/instructions/{management}/{id}-{title}-{undone/done}.md`):**
  - Gunakan akhiran **`-undone.md`** untuk instruksi baru, draft, atau yang masih dikerjakan (contoh: `agents/frontend/instructions/iam_management/1-dashboard-identity-undone.md`).
  - Wajib mandiri memuat 4 heading (`## Task`, `## Request`, `## Action`, `## Context`) mengikuti [`agents/instructions/instruction_template.md`](../agents/instructions/instruction_template.md).
  - Begitu 100% kriteria penerimaan, build, lint, typecheck, Docker, dan pengujian E2E (`PASS`) tuntas serta ter-commit, pelaksana **wajib me-rename ke `-done.md`** (`git mv ...-undone.md ...-done.md`), memperbarui `**Status:** DONE`, mencantumkan commit SHA, serta ringkasan bukti pengujian aktual.
- **Konteks Sesi (`agents/frontend/context/{management}/{id}-context.md`):**
  - Maksimal **1000 baris** per file (jika melewati 1000 baris, lanjut ke `{next-id}-context.md`).
  - Wajib memuat tepat 3 heading Markdown tingkat dua secara berurutan: `## Apa yang sudah dikerjakan`, `## Apa yang belum dikerjakan`, dan `## Apa yang harus dikerjakan selanjutanya` (tulis `Belum ada.` bila kosong tanpa menghapus heading).
  - Simpan hanya arsitektur inti, batasan aktif, keputusan desain, dan status handoff berdampak tinggi; diperbarui **hanya** bila ada perubahan arsitektur/keputusan/milestone penting (bukan typo/formatting minor).
- **Working History (`agents/frontend/working_history/{management}/{id}-{YYYY-MM-DD-HHmmss}-{title}.md`):**
  - Format waktu WIB, **maksimal 150 kata**, lowercase kebab-case pada judul. Tidak perlu dibuat untuk perbaikan typo/formatting minor.
- **Laporan Audit (`agents/frontend/audit/{management}/{id}-{YYYY-MM-DD-HHmmss}-{title}.md`):** lowercase kebab-case, `{id}` berurutan per `{management}`.
- **Laporan Benchmark (`agents/frontend/benchmark/{management}/{id}-{benchmark-title}.md`):** `{id}` berurutan per `{management}`, lowercase kebab-case.
- **Laporan E2E (`agents/frontend/e2e/{management}/{id}-{e2e-title-{Query***/Mutation***}}.md`):** `{id}` berurutan per `{management}`, lowercase kebab-case menyertakan query/mutation yang diuji.
- *(Catatan: permintaan perubahan API/permission ke Backend tetap ditulis ke inbox Backend di `agents/notes/{to_management}/{id}-frontend.md` sesuai Seksi 8B).*
