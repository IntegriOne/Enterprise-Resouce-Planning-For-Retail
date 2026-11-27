# E2E Verification #6 — Workbench Penugasan Peran & Izin Pengguna IAM (`UserDetailDrawer`, `RoleDetailDrawer`, `ScopePicker`)

**Tanggal:** 2026-11-27 WIB  
**Lingkungan Uji:** Docker Compose lokal (`https://erp.localhost:8443`), `frontend-1` & `frontend-2` (`Next.js 16.4.6`), `gateway-1` & `gateway-2`, `iam`, `enterprise-foundation`, `human-resource`, `postgres`, `redis`, `caddy` (`docker compose ps`: seluruh 10 layanan `Up` / `healthy`).

---

## 1. Verifikasi Unit Test, Typegen, Lint, & Typecheck (`frontend/`)

Perintah:
```bash
cd frontend && node --test src/lib/auth/*.test.ts src/lib/graphql/*.test.ts && npx next typegen && npm run lint && npx tsc --noEmit
```

Hasil teramati:
- `node --test`: **23/23 PASS** (`fail: 0`, `skipped: 0`), termasuk [`iam-role-access-helpers.test.ts`](file:///home/muriarfad/Documents/projects/erp-retail/frontend/src/lib/graphql/iam-role-access-helpers.test.ts) (uji kompatibilitas `SYSTEM` vs akun manusia untuk `*_automation` & `superadmin`, pengelompokan & deduplikasi izin efektif lintas peran, serta hint kesiapan verifier `GLOBAL`/`SELF` vs scope eksternal EF).
- `npx next typegen`: **PASS** (`✓ Types generated successfully`).
- `npm run lint`: **PASS** (`eslint` keluar dengan kode `0`).
- `npx tsc --noEmit`: **PASS** (keluar dengan kode `0`).

---

## 2. Matriks Skenario UI Live Chrome Headless CDP (`https://erp.localhost:8443`)

Skrip uji: [`verify_users_roles_ui.mjs`](/home/muriarfad/.gemini/antigravity-cli/brain/73a9c5fb-560c-4093-aa59-de3689eabcb1/scratch/verify_users_roles_ui.mjs) (`exit 0`, JS runtime `exceptions: []`, `badHttp: []`).

| Kasus | Rute / Query / Mutasi / Komponen | Hasil Teramati | Status |
| :--- | :--- | :--- | :---: |
| 1. Autentikasi Operator | `/login` (`userLogin`, `me`, `myAccessSnapshot`) | Login sebagai SuperAdmin berhasil dan redirect otomatis ke `/identity`. | PASS |
| 2. Akses 1-Klik Tab Peran & Workbench | `/identity/users/internals` -> klik tombol `Peran` pada baris `Super Administrator` (`f300be11-ef17-4b88-8a4c-9c938f1bacd0`) | Drawer membuka langsung tab `Peran & Akses` (`UserRoleWorkbench`); menampilkan strip alur `1. Izin Atomik` & `2. Peran RBAC (Role)` serta badge kompatibilitas `KOMPATIBEL` dan `KHUSUS SYSTEM`. | PASS |
| 3. Pratinjau Izin Peran & Matriks Efektif | Query `rolePermissions` & `userEffectiveRoles` (paginasi aman `first: 40`) | Kartu **PRATINJAU IZIN PERAN** menampilkan `enterprise_foundation (10)` dan `iam (15)`; bagian **Izin Atomik Efektif Akun (25)** terisi penuh tanpa melebihi batas kompleksitas Gateway. | PASS |
| 4. Inline Quick Role Builder Domain EF | Klik `Buat Peran Kustom` (`QuickRoleBuilder`) | Form terisi default `quick-role-code = "ef_operator"`, `quick-role-domain = "enterprise_foundation"`, serta memuat katalog izin dan mencentang otomatis `10` izin `enterprise_foundation.*` (`Pilih Izin Atomik (10 dipilih)`). | PASS |
| 5. ScopePicker `SELF` & Peringatan `LEGAL_ENTITY` | `ScopePicker` pada `UserRoleWorkbench` | Memilih `SELF` mengisi otomatis `user-role-scope-id = "f300be11-ef17-4b88-8a4c-9c938f1bacd0"` dengan input terkunci (`disabled: true`) dan hint `Otomatis mengikat UUID akun target ini`. Memilih `LEGAL_ENTITY` menampilkan banner peringatan `MENUNGGU VERIFIER EF`. | PASS |
| 6. Responsivitas Mobile `320px` | Viewport `320x740` dengan drawer terbuka | `documentElement.scrollWidth === 320` dan `window.innerWidth === 320` (0 horizontal overflow). | PASS |
| 7. Deep-Link Pembuatan Peran Domain EF | `/identity/roles?action=create&domain=enterprise_foundation` (`RoleFormModal`) | Modal terbuka otomatis dengan `code = "ef_operator"`, `management_domain = "enterprise_foundation"`, dan `default_redirect_path = "/human-resource"` (valid di `iam-role-access-helpers.ts` / allowlist Gateway). | PASS |
| 8. Proteksi Peran Sistem Immutable | `/identity/roles` -> klik `Izin & Detail` pada `superadmin` (`RoleDetailDrawer`) | Menampilkan banner **Peran Bawaan Sistem (Immutable)** dan menyembunyikan aksi mutasi izin pada peran `is_system: true`. | PASS |
| 9. Drawer Grup (`*_automation` Filter) | `/identity/groups` (`GroupDetailDrawer`) | Halaman `/identity/groups` dimuat normal tanpa error, namun belum ada data grup seeded pada DB lokal (`groupCheck.tested: false`), sehingga pengujian visual pemilih peran di dalam drawer grup belum dapat diverifikasi terhadap data live. | N/A (Belum Ada Data Grup) |

---

## 3. Catatan Cakupan Mutasi

- Penugasan peran & pembuatan peran kustom pada level kontrak mutasi (`submitCreateRole`, `submitAssignRolePermission`, `submitAssignUserRole`, `submitAssignGroupRole`, `X-CSRF-Token`, serta validasi invarian scope client-side) terverifikasi melalui unit test [`iam-dashboard.test.ts`](file:///home/muriarfad/Documents/projects/erp-retail/frontend/src/lib/graphql/iam-dashboard.test.ts) dan pratinjau/persiapan form UI live tanpa menuliskan peran dummy permanen ke database lokal.
