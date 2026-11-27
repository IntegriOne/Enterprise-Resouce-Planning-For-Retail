# Konsep Objek `users`, Identitas, dan Akses Lintas Management (IAM IntegriOne)

> **Pedoman Normatif Identitas & Akses — Menggantikan `agents/principals.md`.**
> Dokumen ini menetapkan model tunggal objek **`users`** milik [service IAM](../backend/services/iam/db/schema/identity_access_management.dbml), batas kepemilikan data bisnis lintas management, mekanisme autentikasi `/api/graphql`, serta evaluasi otorisasi RBAC–ABAC berbasis `AccessSnapshot` secara *fail-closed*.

Rujukan pendamping: [AGENTS.md](../AGENTS.md), [agents/boundary.md](boundary.md), [agents/arsitektur.md](arsitektur.md), [perspektif produk](../docs/development-product-perspektif.md), dan [kebijakan sesi & keamanan](../docs/02-architecture/security-and-sessions.md).

---

## 1. Satu Sumber Identitas: Objek `users` di IAM

Seluruh akun manusia maupun akun sistem non-interaktif pada ekosistem **IntegriOne (Web Commerce + Retail ERP)** direpresentasikan dalam **satu tabel dan entitas `users`** milik database `iam`. Tidak boleh ada tabel akun login terpisah per management, tidak boleh ada duplikasi password atau role di modul bisnis, dan tabel `iam.users` **tidak memiliki** kolom JSONB `attributes`.

| Tabel / Komponen IAM | Peran dan Aturan Inti (Sesuai Skema & Source Aktif) |
|---|---|
| `users` | Sumber identitas utama (`id` UUID, `user_kind`, `status`, `display_name`, `avatar_url`, `locale` bawaan `id-ID`, `timezone` bawaan `Asia/Jakarta`, `auth_epoch`, `perm_revision`, `last_login_at`, `created_at`, `updated_at`, `deleted_at`). |
| `user_identities` | Pengenal login terpisah dari akun (`kind`: `email` \| `username`, `value` ter-normalisasi lowercase, `is_primary`, `is_verified`, `verified_at`). Maksimal 1 pending email + 1 verified email + 1 username per `user_id`. Identitas login wajib terverifikasi (`is_verified = true`) agar dapat masuk. |
| `user_credentials` | Kredensial sandi satu-ke-satu (`password_hash` Argon2id, `password_changed_at`, `must_change_password`, `failed_login_attempts`, `locked_until`). Batas gagal bawaan: `5` kali (`DefaultMaxFailedLoginAttempts`), kunci sementara `15` menit (`DefaultLoginLockDuration`). |
| `user_totp_factors` & `user_recovery_codes` | Faktor 2FA TOTP terenkripsi (`encrypted_secret`, `key_version`, `is_enabled`, `confirmed_at`, `last_used_step` anti-replay) dan kode pemulihan ter-hash sekali pakai (`code_hash`, `used_at`). |
| `user_tokens` | Token sekali pakai SHA-256 (`purpose`: `email_verification` \| `password_reset` \| `account_invitation` \| `mfa_challenge`; `status`: `active` \| `consumed` \| `expired` \| `revoked`). Maksimal 1 token `active` per `(user_id, purpose)`. |
| `sessions` | Sesi stateful server-side (`token_digest`, `csrf_token_digest`, `generation`, `csrf_key_version`, `ip_address`, `user_agent`, `authenticated_at`, `last_seen_at`, `idle_expires_at`, `absolute_expires_at`, `mfa_verified`, `ended_at`, `end_reason`). |
| `roles`, `permissions`, `role_permissions` | Katalog hak akses atomik (`code`: `<domain>.<resource>.<action>`, `is_sensitive`) serta peran (`management_domain`, `default_redirect_path`, `redirect_priority`, `is_system`, `is_active`). |
| `groups`, `group_members`, `user_roles`, `group_roles` | Penugasan peran langsung maupun berbasis grup dengan pembatasan `scope_type`, `scope_id`, `assigned_by`, dan masa berlaku `expires_at`. |
| `iam_security_control` & `iam_outbox_events` | Kontrol global (`recovery_epoch` di baris `id = 1`) dan outbox transaksional (`pending` \| `effective` \| `failed`) untuk proyeksi keamanan ke Redis (`auth_epoch`, `perm_revision`, `recovery_epoch`). |

### Nilai Enum `user_kind` dan `user_status`

1. **`user_kind`** ([`enums.go`](../backend/services/iam/domain/entity/enums.go)):
   - `customer` — pembeli pada Web Commerce.
   - `employee` — pegawai/operator internal ERP IntegriOne (termasuk human administrator awal `superadmin`).
   - `vendor` — perwakilan mitra vendor jasa eksternal (mis. CMT).
   - `supplier` — perwakilan mitra pemasok bahan/barang eksternal.
   - `system` — akun layanan otomasi non-interaktif per-management (`auto_system_iam`, `auto_system_ef`). **Dilarang keras** memiliki email, kata sandi, TOTP/recovery code, sesi browser interaktif, maupun rute redirect dashboard (`ErrSystemUserInteractiveForbidden`).
2. **`user_status`** ([`enums.go`](../backend/services/iam/domain/entity/enums.go) & [`user.go`](../backend/services/iam/domain/entity/user.go)):
   - `pending` — akun baru menunggu verifikasi email / aktivasi awal.
   - `active` — akun aktif dan memenuhi syarat autentikasi.
   - `suspended` — akun ditangguhkan sementara oleh admin.
   - `disabled` — akun dinonaktifkan (termasuk setelah `SoftDelete` yang mengisi `deleted_at` dan mengubah status ke `disabled`).

> [!IMPORTANT]
> Hanya akun dengan `status = active` dan `deleted_at IS NULL` yang lolos `User.CanAuthenticate()` maupun `AccessSnapshot.HasPermission(...)`. Akun `pending`, `suspended`, `disabled`, atau yang dihapus lunak ditolak secara *fail-closed* pada login, penerbitan sesi, operasi profil mandiri (`*self`), maupun seluruh aksi RBAC.

---

## 2. Batas Kepemilikan Data & Referensi Lintas Management

IAM hanya menyimpan **identitas login, keamanan akun, sesi, serta otorisasi RBAC/scope**. Atribut bisnis tetap berada di database pemilik domain masing-masing ([agents/boundary.md](boundary.md)), tanpa **Foreign Key SQL lintas database** dan tanpa **JSONB `attributes`** di tabel `users`.

| Management Pemilik | Data yang Dimiliki Domain | Tautan Logis ke `iam.users.id` | Hal yang Dilarang |
|---|---|---|---|
| **IAM** | `users`, `user_identities`, `user_credentials`, MFA, `sessions`, `roles`, `permissions`, `groups`, `user_roles`, `group_roles` | `iam.users.id` (UUID sumber kebenaran; Global ID GraphQL: `gid://integrione/user/<uuid>`) | Menyimpan data HR, party B2B, kontrak, gaji, atau alamat commerce di `users`. |
| **Enterprise Foundation** | `legal_entities`, `organization_units`, `locations`, `parties`, `party_roles`, `party_contacts` | `party_contacts.iam_user_id` (`uuid` opsional) + `portal_access_status`; UUID entitas/unit/lokasi/party menjadi `scope_id` di IAM | Memberikan akses hanya karena record `party_contacts` ada tanpa assignment role aktif di IAM; FK SQL ke `iam.users`. |
| **HRM** | `hr_workers`, kontrak kerja, penempatan, talent, kompensasi, cuti, payroll | `hr_workers.iam_principal_reference_id` (`uuid` opsional mengarah ke `iam.users.id`) | Menyamakan `worker_id` dengan `user_id`; menjadikan mutasi/status HRM otomatis membuat/mengubah grant IAM. |
| **Procurement / Finance / WMS / OMS / Commerce / DRM / CRM / CS** | Dokumen PO/RFQ/kontrak, invoice/AP/AR, stok/lokasi gudang, order/cart, arsip dokumen | Menyimpan `user_id` (atau mengecek `CallerContext.UserID` dan `party_id` dari owner) untuk kepemilikan objek & audit | Mengambil daftar izin dari klien; menganggap akun `customer` = staf; melewati cek kepemilikan objek (`node(id:)`, list, export). |

**Kaidah wajib lintas service:**
1. Management non-IAM **dilarang keras** melakukan `SELECT`/`INSERT`/`UPDATE`/`DELETE` langsung ke tabel database `iam` (`users`, `user_roles`, dll.).
2. Pemetaan `user_id` ↔ entitas bisnis (misal PIC mitra pada `party_contacts.iam_user_id` atau pekerja pada `hr_workers.iam_principal_reference_id`) diverifikasi di service pemilik domain saat menjalankan ABAC atau lifecycle domain.
3. Menonaktifkan/suspend/revoke di IAM langsung menghentikan akses (via `auth_epoch`/`perm_revision` dan `sessions`), sedangkan nonaktifnya hubungan bisnis di Foundation/HRM tetap harus memblokir operasi domain meski role IAM belum dicabut.

---

## 3. Autentikasi Terpadu (`/api/graphql`), Sesi Opaque & Post-Login Redirect

Sesuai [keputusan owner API tunggal](../docs/00-governance/decisions.md) dan [perspektif produk](../docs/development-product-perspektif.md):

1. **Satu Halaman Masuk Umum (`/login`):**
   - Halaman `/login` melayani `customer`, `employee`, `vendor`, dan `supplier` lewat mutasi `userLogin` (dan `validateUserMfa` bila TOTP aktif) pada satu endpoint HTTP resmi **`/api/graphql`** (`Content-Type: application/json`, origin tervalidasi, serta guard AST depth `<= 10` & aliases `<= 25`).
   - Dilarang menambah endpoint REST `/api/auth/*` atau memisah halaman login per peran.
   - **Kontrak Non-Login Akun `system` (`ErrSystemCannotLogin`):** segala percobaan `AuthenticatePassword` / `userLogin` terhadap akun `user_kind = system` (mis. `auto_system_iam`, `auto_system_ef`), termasuk dengan **`password` kosong (`""`)** maupun terisi, wajib ditolak eksplisit dengan **`system cannot login`** (`iam.user.system_cannot_login`, gRPC `PermissionDenied` / GraphQL `AUTHORIZATION`), yang ditegakkan mulai dari [`User.CanAuthenticate`](../backend/services/iam/domain/entity/user.go) hingga [`auth.proto`](../backend/services/iam/contract/proto/iam/v1/auth.proto) dan Gateway [`normalizeAuthError`](../backend/gateway/internal/adapter/iam/adapter.go).
2. **Mekanisme Login & Perlindungan Anti-Enumerasi ([`auth.go`](../backend/services/iam/services/user/auth.go)):**
   - Mendukung login dengan `email` maupun `username` (`user_identities`), memverifikasi Argon2id, dan menjalankan dummy hash verification *constant-time* jika identitas tidak ditemukan atau akun terkunci sementara (`locked_until`) agar tidak membocorkan eksistensi/status akun.
   - Pembuatan akun (`createUser`) mengaktifkan status awal `pending` sampai email diverifikasi lewat `verifyEmail` (mengaktifkan seluruh kategori akun non-`system` yang berstatus `pending`). Untuk pendaftar anonim yang mengulang email terdaftar pada `createUser` atau `resendValidationEmailToken`, sistem mengembalikan respons netral/redirect aman tanpa membocorkan data pengguna.
3. **Sesi Opaque & CSRF Terikat Sesi ([`session/service.go`](../backend/services/iam/services/session/service.go), [`middleware/session.go`](../backend/gateway/internal/middleware/session.go)):**
   - Cookie browser memakai **`__Host-integrione_session`** (*HttpOnly*, *Secure*, `Path=/`, *SameSite=Lax*) berisi *handle opaque*, **bukan JWT** yang menyimpan role atau permission.
   - Mutasi mewajibkan header **`X-CSRF-Token`** yang cocok dengan token terikat sesi dari query `browserSession` / pre-auth session (`IssuePreAuthSession`) serta validasi `Origin`/`Referer`.
   - Batas waktu sesi (D01–D12): **idle timeout `30` menit** (`SessionIdleTimeout`) dan **absolute timeout `8` jam** (`SessionAbsoluteTimeout`, tidak diperpanjang oleh aktivitas).
4. **Resolusi Tujuan Pasca-Login (`RedirectResolver`):**
   - Server IAM menentukan `redirectLocation` menggunakan `roles.default_redirect_path` dari peran aktif milik user, diurutkan berdasarkan **`redirect_priority` angka terkecil menang** (`RedirectResolver.ResolvePostLoginPath` di [`redirect_policy.go`](../backend/services/iam/domain/service/redirect_policy.go)).
   - Jalur redirect harus lolos validasi path internal relatif terhadap allowlist resmi: `/`, `/identity`, `/human-resource`, `/procurement`, `/warehouse`, `/finance`, `/accounting`, `/product`, `/supplier`, `/vendor`, `/commerce`, `/account`, `/profile`.
   - Jika seluruh peran aktif user memiliki `default_redirect_path` kosong (misal akun baru dengan role `iam_selfservice`), fallback aman mengikuti `user_kind`: `vendor` → `/vendor`, `supplier` → `/supplier`, `employee` → `/account`, `customer` / `system` → `/`.

---

## 4. Evaluasi Otorisasi RBAC–ABAC, `ScopeType`, dan Layanan Mandiri (`*self`)

### 4.1 Model Evaluasi Gateway + Service Pemilik

1. **Kesegaran `AccessSnapshot` Redis (`<= 30` detik):**
   - Gateway menyegarkan snapshot otorisasi via `SnapshotLoader.EnsureFresh` (menggunakan `singleflight`) dengan batas kesegaran maksimum **`30` detik** (`DefaultSnapshotMaxAge` di [`snapshot.go`](../backend/shared/authz/snapshot.go)).
   - Snapshot memverifikasi kecocokan `AuthEpoch`, `PermRevision`, dan `RecoveryEpoch` terhadap runtime state. Bila IAM atau Redis unavailable atau snapshot usang/gagal disegarkan, permintaan ditolak secara **fail-closed** (`UNAVAILABLE` / `UNAUTHENTICATED` / `PERMISSION_DENIED`) — **tidak pernah ada fallback ALLOW**.
2. **Penegakan di `Authorizer` & `AccessEvaluator` ([`access_policy.go`](../backend/services/iam/domain/service/access_policy.go), [`authorizer.go`](../backend/gateway/internal/middleware/authorizer.go)):**
   - Memastikan `user.CanAuthenticate()` sah (`active`, tidak soft-delete), `session.CheckStatus(now)` sah, `session.UserID == user.ID`, snapshot segar, dan izin `<management>.<resource>.<action>` cocok pada `RequiredScope`.
   - Jika permission berflag **`is_sensitive = true`** (atau mutasi menandai sensitif), sistem mewajibkan re-autentikasi sensitif dalam jendela **`10` menit** (`DefaultSensitiveReauthWindow` / `sensitiveReauthAge`) serta sesi telah terverifikasi MFA bila TOTP akun aktif (`VerifyMySessionMfa`).
3. **Penegakan Domain (ABAC) di Service Pemilik gRPC:**
   - Gateway hanya meneruskan identitas & konteks pemanggil terverifikasi (`CallerContext`) dan menjalankan guard perimeter/izin.
   - **Service pemilik domain** tetap wajib memvalidasi scope objek aktual terhadap database miliknya (misal `legal_entity_id`, `facility_id`, `party_id`, atau kepemilikan `user_id` untuk pelanggan), memverifikasi SoD, melakukan recheck sebelum commit untuk write sensitif, dan **tidak boleh menulis row atau outbox saat ditolak**. Penolakan harus konsisten di query `node(id:)`, nested field, list, count, filter, maupun export.

### 4.2 Tujuh `ScopeType` yang Didukung Skema IAM

| `scope_type` | `scope_id` | Sumber UUID Target | Aturan Khusus |
|---|---|---|---|
| `global` | **Wajib `NULL`** | — | Berlaku di seluruh ruang lingkup perusahaan; `ScopeGrant.Matches` mengizinkan target scope apa pun bila `ID == ""`. |
| `legal_entity` | Wajib UUID valid | `legal_entities.id` (Enterprise Foundation) | Tidak memenuhi permintaan ber-scope `global`; hanya cocok dengan `legal_entity` ber-UUID sama. |
| `organization_unit` | Wajib UUID valid | `organization_units.id` (Enterprise Foundation) | Membatasi operasi ke divisi/unit organisasi tertentu. |
| `facility` | Wajib UUID valid | `locations.id` (Enterprise Foundation) | Membatasi operasi ke kantor/gudang/toko tertentu. |
| `supplier_party` | Wajib UUID valid | `parties.id` (Enterprise Foundation) | Untuk akun PIC mitra pemasok (`user_kind = supplier`) atau delegasi terkait supplier tertentu. |
| `vendor_party` | Wajib UUID valid | `parties.id` (Enterprise Foundation) | Untuk akun PIC mitra vendor (`user_kind = vendor`) atau delegasi terkait vendor tertentu. |
| `self` | Wajib UUID = `users.id` pemilik | `users.id` (IAM sendiri) | Ditolak keras pada `group_roles` (`NewGroupRole`); hanya sah untuk izin `*self` milik user bersangkutan. |

### 4.3 Baseline Auto-Assignment & Isolasi Izin `*self`

- Pada transaksi `CreateUser` ([`user/service.go`](../backend/services/iam/services/user/service.go)), IAM otomatis menetapkan peran baseline berscope **`ScopeSelf{ID: user.ID}`**:
  - `customer` → peran **`customer`** (`default_redirect_path: "/"`, `redirect_priority: 100`).
  - `employee`, `vendor`, `supplier` → peran **`iam_selfservice`** (`default_redirect_path: ""`, `redirect_priority: 999`).
- Kode izin mandiri (`.readself`, `.updateself`, `.deleteself`) dievaluasi secara khusus pada `AccessSnapshot.HasPermission` dan `AccessEvaluator.Authorize`: izin `*self` **tidak pernah berlaku** pada `ScopeGlobal` maupun untuk `user_id` milik pihak lain (`PERMISSION_DENIED`), dan **tetap diblokir total** selama status user masih `pending` / belum terverifikasi email.

### 4.4 Katalog Bootstrap IAM & Sinkronisasi Manifest RBAC Lintas Management ([`bootstrap.go`](../backend/services/iam/services/bootstrap.go))

Saat startup, `BootstrapPrincipalsAndRBAC` menjamin kesiapan baseline inti IAM dan katalog RBAC di database `iam` secara idempoten (menggunakan *PostgreSQL advisory lock* saat sinkronisasi katalog aktif):
- **Human `superadmin` vs Akun Otomasi `system`:**
  - Human `superadmin` (`IAM_SUPERADMIN_*` pada `.env`, `user_kind = employee`) adalah akun manusia pertama dengan akses tata kelola penuh saat rilis awal (`superadmin` global + `iam_selfservice` self).
  - Setiap management yang memerlukan identitas layanan memiliki akun otomasi non-interaktif terpisah (`user_kind = system`, tanpa email/kata sandi/sesi browser): `auto_system_iam` (`iam_automation`) dan `auto_system_ef` (`ef_automation`). Isolasi peran ↔ `user_kind` (`Entity.IsRoleCompatibleWithUserKind`) menolak penetapan `superadmin` ke akun `customer`/`system` maupun peran `*_automation` ke akun manusia (`ErrRoleUserKindMismatch`), baik lewat `user_roles` langsung maupun `group_members`/`group_roles`.
- **Arsitektur Multi-Management (*Declarative Catalog Manifest* — Aktif untuk `iam` & `enterprise_foundation`):**
  - Seluruh baseline inti (`15` permission `iam.*`, `10` permission `enterprise_foundation.*`, dan `7` system role) kini didefinisikan secara deklaratif di `backend/shared/rbac/catalogs/{iam,enterprise_foundation}.json`, disematkan (`//go:embed`), divalidasi ketat oleh `backend/shared/rbac`, dan disinkronkan eksklusif oleh `iam` saat startup melalui `BootstrapPrincipalsAndRBAC` (`syncEmbeddedRBACCatalogs`). Management baru mendaftarkan katalognya di `backend/shared/rbac/catalogs/{management}.json` agar otomatis tersimpan di DB `iam` (local maupun production).
  - *Invarian sinkronisasi produksi:* isolasi namespace `{management}.*`, tanpa pewarisan otomatis ke role `superadmin` untuk domain bisnis selain baseline IAM/EF, tanpa *auto-delete* permission yang sudah pernah terdaftar di DB, validasi `default_redirect_path` terhadap allowlist `RedirectResolver`, pengetatan sensitivitas (`false` → `true`) memicu pembaruan revisi keamanan serta invalidasi outbox/Redis, penurunan sensitivitas (`true` → `false`) ditolak *fail-closed*, *PostgreSQL transaction advisory lock* (`pg_advisory_xact_lock`), serta template role non-sistem (`is_system = false`) bersifat *seed-once* (*create-if-missing*) agar perubahan admin perusahaan lewat `/identity` di production tidak tertimpa saat deploy.
- **7 System Role Bawaan Aktif Saat Ini:**
  - `superadmin` (`/identity`, priority `1`) — seluruh 15 izin `iam.*` + 10 izin `enterprise_foundation.*` (hanya untuk `employee`).
  - `iam_operator` (`/identity`, priority `10`) — operasional user, group, & revoke session (tanpa `iam.roles.write`).
  - `iam_readonly` (`/identity`, priority `20`) — baca users, roles, groups, permissions, sessions, plus `iam.users.readself` & `iam.sessions.readself`.
  - `iam_automation` (`""`, priority `999`) — akun otomasi internal IAM (`auto_system_iam`, `user_kind = system`).
  - `ef_automation` (`""`, priority `999`) — 10 permission `enterprise_foundation.*` untuk akun otomasi Enterprise Foundation (`auto_system_ef`, `user_kind = system`).
  - `customer` (`/`, priority `100`) & `iam_selfservice` (`""`, priority `999`) — hanya 5 permission `*self` (`iam.users.readself`, `iam.users.updateself`, `iam.users.deleteself`, `iam.sessions.readself`, `iam.sessions.deleteself`).

---

## 5. Tata Kelola Permission, Manifest Kode, & Pemberian Akses Lintas Management

1. **Merancang & Mendaftarkan Katalog Permission Management Baru:**
   - **Langkah 1 — Dokumen Rancangan:** Agen/developer pemilik management menuliskan rancangan permission (`<management>.<resource>.<action>`), penanda `is_sensitive`, daftar `scope_type` yang valid, serta kandidat *template role* domain pada **`agents/roles-permissions/{management}/1-roles-permissions.md`** sesuai panduan di [how_to_code_grpc_managements.md](how_to_code_grpc_managements.md) dan [how_to_code_wiring_grpc_managements_to_gateway.md](how_to_code_wiring_grpc_managements_to_gateway.md).
   - **Langkah 2 — Manifest Deklaratif Resmi (`backend/shared/rbac/catalogs/{management}.json`):** Agar permission & template role terverifikasi otomatis tersimpan ke database `iam` saat deploy production (tanpa intervensi manual maupun SQL lintas DB), pemilik management menuangkan spesifikasi final (khusus operasi yang kontrak gRPC/Gateway-nya aktif) ke manifest domain masing-masing serta menjaga test konsistensi 1:1 terhadap kode permission di service/Gateway.
   - **Langkah 3 — Eskalasi Inti IAM (Jika Perlu):** Penambahan `default_redirect_path` baru di luar allowlist `defaultAllowedRedirectRoutes` ([`redirect_policy.go`](../backend/services/iam/domain/service/redirect_policy.go)), konfigurasi principal `auto_system_{mgmt}`, atau wiring `ScopeOwnerVerifier` tetap diajukan melalui `agents/notes/iam_management/` kepada pemilik IAM.
2. **Pemberian & Pencabutan Grant Nyata:**
   - Assignment `user_roles`, `group_roles`, maupun keanggotaan `group_members` hanya dilakukan melalui layanan/mutasi resmi IAM oleh pemegang izin `iam.roles.write` / `iam.groups.write` ([`iam_rbac.graphql`](../backend/gateway/graphql/iam/iam_rbac.graphql)).
   - Dilarang melakukan *seed* spekulatif *grant* ke akun user maupun mendaftarkan permission fitur yang kontraknya belum tersedia; dilarang menulis ke tabel RBAC IAM secara langsung lewat SQL lintas modul.
3. **Catatan Status *Workflow Access Request*:**
   - Alur persetujuan permintaan akses formal lintas management (`access_requests` / *inbox approval* terpisah) adalah **konsep tata kelola** yang belum memiliki tabel tersendiri pada migrasi IAM `000001`. Pada runtime aktif saat ini, onboarding akun pegawai/mitra dilakukan oleh operator berizin melalui mutasi IAM (`createEmployeeUser`, `createVendorUser`, `createSupplierUser`, lalu binding role/group berscope setelah verifikasi domain di HRM/Foundation selesai).

---

## 6. Contoh Penerapan pada Tiga Konteks Utama IntegriOne

### Konteks A — Staf Internal (`user_kind = employee`)
1. HRM menyimpan pekerja di `hr_workers` beserta jabatan/kontrak; tautan ke IAM bersifat logis lewat `hr_workers.iam_principal_reference_id` (opsional, bukan FK SQL).
2. Operator IAM membuat akun lewat `createEmployeeUser` → terbentuk `users` (`employee`, `pending`), `user_identities`, `user_credentials`, dan role baseline `iam_selfservice` (`ScopeSelf`). Setelah verifikasi email (`verifyEmail`), status menjadi `active`.
3. Operator/Super Admin IAM menugaskan role operasional bisnis dengan `scope_type` sesuai kebutuhan (`global`, `legal_entity`, `organization_unit`, atau `facility`) dan `default_redirect_path` sesuai workspace (mis. `/identity`, `/human-resource`, `/warehouse`).

### Konteks B — Mitra Supplier / Vendor (`user_kind = supplier` / `vendor`)
1. Enterprise Foundation mengelola badan usaha mitra (`parties`, `party_roles`) dan daftar narahubung resmi (`party_contacts`) dengan kolom referensi **`party_contacts.iam_user_id`** dan `portal_access_status`.
2. IAM membuat akun manusia PIC (`createSupplierUser` atau `createVendorUser`) lengkap dengan baseline `iam_selfservice` (`ScopeSelf`).
3. Admin IAM menugaskan role portal mitra ber-scope `supplier_party` atau `vendor_party` dengan **`scope_id = parties.id`**. Saat PIC mengakses PO, RFQ, atau dokumen, service domain wajib memeriksa izin IAM + kecocokan `scope_id == party_id` + status aktif `party`/`party_contacts` di Foundation agar PIC vendor A tidak pernah bisa membaca data vendor B.

### Konteks C — Pelanggan Web Commerce (`user_kind = customer`)
1. Pengunjung mendaftar melalui `createUser` → terbentuk `users` (`customer`, `pending`) + baseline role `customer` (`ScopeSelf`, `default_redirect_path: "/"`, `redirect_priority: 100`).
2. Sebelum token verifikasi email dikonfirmasi (`verifyEmail`), `users.status = pending` dan identitas belum terverifikasi, sehingga login serta semua operasi `*self` ditolak.
3. Setelah aktif, pelanggan login melalui `/login` (`userLogin`) dan mendapat cookie opaque `__Host-integrione_session`. Modul Commerce/OMS/DRM memakai `CallerContext.UserID` untuk memastikan pelanggan hanya dapat mengakses keranjang, order, dan invoice miliknya sendiri (anti-IDOR).

---

## 7. Checklist Wajib Developer & Agen Sebelum Menulis Kode Domain

- [ ] Pakai **`users.id`** (`UUID` IAM / `CallerContext.UserID`) sebagai kunci identitas pemanggil — jangan memakai `worker_id` HRM atau `party_id` Foundation sebagai substitusi autentikasi.
- [ ] Pastikan mutasi browser hanya lewat **`/api/graphql`** dengan **`__Host-integrione_session`** + **`X-CSRF-Token`** + validasi `Origin`.
- [ ] Bedakan kode error **`UNAUTHENTICATED`** (belum login / sesi kedaluwarsa), **`PERMISSION_DENIED`** (login sah tanpa grant/scope cocok), dan **`UNAVAILABLE`** (IAM/Redis gagal; wajib *fail-closed*, tanpa data/outbox baru).
- [ ] Untuk akses pribadi (`me`, `mySessions`, `updateMyProfile`, dll.), wajib memakai izin `*self` pada **`ScopeSelf{ID: currentUser.ID}`**, bukan `ScopeGlobal`.
- [ ] Cek otorisasi & scope **ABAC pemilik objek** pada **semua titik masuk**: `node(id:)`, query detail, list/filter/count, nested field, mutation, dan export.
- [ ] Tidak ada `SELECT`/`JOIN`/FK lintas database ke skema PostgreSQL `iam`, dan tidak ada asumsi kolom `attributes` JSONB pada tabel `users`.

---

## 8. Referensi Sumber Terverifikasi

- Skema & Migrasi IAM: [`identity_access_management.dbml`](../backend/services/iam/db/schema/identity_access_management.dbml) · [`000001_create_iam_schema.up.sql`](../backend/services/iam/db/migrations/000001_create_iam_schema.up.sql)
- Entitas & Kebijakan Domain IAM: [`user.go`](../backend/services/iam/domain/entity/user.go) · [`enums.go`](../backend/services/iam/domain/entity/enums.go) · [`role.go`](../backend/services/iam/domain/entity/role.go) · [`assignment.go`](../backend/services/iam/domain/entity/assignment.go) · [`session.go`](../backend/services/iam/domain/entity/session.go) · [`validation.go`](../backend/services/iam/domain/entity/validation.go) · [`access_policy.go`](../backend/services/iam/domain/service/access_policy.go) · [`redirect_policy.go`](../backend/services/iam/domain/service/redirect_policy.go)
- Service Aplikasi & Bootstrap IAM: [`bootstrap.go`](../backend/services/iam/services/bootstrap.go) · [`user/service.go`](../backend/services/iam/services/user/service.go) · [`user/auth.go`](../backend/services/iam/services/user/auth.go) · [`user/registration_runtime.go`](../backend/services/iam/services/user/registration_runtime.go) · [`session/service.go`](../backend/services/iam/services/session/service.go)
- Shared Authz & Gateway Middleware: [`shared/authz/snapshot.go`](../backend/shared/authz/snapshot.go) · [`middleware/session.go`](../backend/gateway/internal/middleware/session.go) · [`middleware/snapshot.go`](../backend/gateway/internal/middleware/snapshot.go) · [`middleware/authorizer.go`](../backend/gateway/internal/middleware/authorizer.go)
- Kontrak GraphQL IAM Gateway: [`iam_auth.graphql`](../backend/gateway/graphql/iam/iam_auth.graphql) · [`iam_user.graphql`](../backend/gateway/graphql/iam/iam_user.graphql) · [`iam_rbac.graphql`](../backend/gateway/graphql/iam/iam_rbac.graphql)
- Tautan Domain Terkait: [`enterprise_foundation.dbml`](../backend/services/enterprise_foundation/db/schema/enterprise_foundation.dbml) · [`human_resource_management.dbml`](../backend/services/human_resource/db/schema/human_resource_management.dbml) · [`iam/1-context.md`](context/iam/1-context.md)
