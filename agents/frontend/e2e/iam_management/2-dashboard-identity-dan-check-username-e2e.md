# Laporan E2E Frontend #2: Login Umum Netral, Dashboard IAM (`/identity/*`), Sidebar, Navbar, & Check Username

- **Tanggal:** 2026-11-25 01:30 WIB
- **Lingkungan Uji:** Reverse proxy Caddy (`https://erp.localhost:8443`), `frontend-1` (`127.0.0.1:3333`) & `frontend-2` (`127.0.0.1:3334`) (Next.js `16.4.6` / React `19.3.0` / Node `26.x`), `gateway-1` & `gateway-2`, `iam`, `enterprise-foundation`, `redis`, `postgres`.
- **Operasi `/api/graphql` & Rute Teruji:**
  - `Query`: `browserSession`, `myWorkspaceAccess`, `me`, `myAccessSnapshot`, `users`, `roles`, `groups`, `permissions`, `mySessions`, serta lookup username `users(query) { nodes { ... identities { nodes { kind value is_verified } } } }`
  - `Mutation`: `userLogin`, `userLogout`
  - Route: `/login`, `/register`, `/verify-email`, `/identity`, `/identity/users`, `/identity/roles`, `/identity/groups`, `/identity/permissions`, `/identity/sessions`

## Matriks Kasus Uji

| No | Kasus Uji | Kondisi / Aktor | Hasil Ekspektasi | Hasil Aktual | Status |
|---|---|---|---|---|---|
| 1 | Unit test auth & dashboard IAM client | `node --test` (`redirect-allowlist.test.ts`, `iam-browser-auth.test.ts`, `iam-dashboard.test.ts`) | Seluruh `17` tes lulus termasuk `checkUsernameAvailabilityViaGraphQL` | `17/17` PASS (`0` fail) | PASS |
| 2 | Typegen, ESLint React Compiler, TypeScript, & Production Build | Host `frontend/` | Exit `0`, zero lint errors/warnings, 13 rute terkompilasi | `next typegen`, `lint`, `tsc`, & `next build` (13/13) exit `0` | PASS |
| 3 | Docker single-build `frontend-1` & runtime multi-replika + Caddy | `docker compose build frontend-1` & `up -d --wait frontend-1 frontend-2 caddy` | `frontend-1`, `frontend-2`, dan `caddy` berstatus `Healthy` | Seluruhnya `Healthy` | PASS |
| 4 | Akses anonim `/identity` (Skenario C) & health GraphQL | Anonim (`curl -sk`) via `https://erp.localhost:8443` | `/identity` -> `307` ke `/login`; `/api/graphql?query={health}` -> `ok` | HTTP `307` redirect `/login`; `health` ok | PASS |
| 5 | Audit zero kebocoran tipe user / arsitektur internal di publik | Chrome Headless `1440x900` pada `/login` (Masuk + Daftar Akun), `/register`, `/verify-email` | Nol kemunculan kata `vendor`, `supplier`, `customer`, `pelanggan`, `internal`, `mitra`, `portal`, `karyawan`, `staf`, `3 jenis`, `gateway`, `ruang kerja` | `loginLeak: null`, `regTabLeak: null`, `registerLeak: null`, `verifyLeak: null` | PASS |
| 6 | Login Admin `SystemIAM` -> seluruh rute `/identity/*`, Sidebar, Navbar, Command Palette, Step-Up Modal & Logout (Skenario A) | Chrome Headless `https://erp.localhost:8443` (`1440x900` & `320x740` `prefers-reduced-motion`) | Semua 6 rute `HTTP 200`, font `Geist` aktif di Dashboard, Command Palette & Step-Up dialog terbuka, `scrollWidth <= innerWidth`, `0` JS error, `0` HTTP >= 400, logout bersih ke `/login` | `/identity/*` `200`, `h1Font: Geist...`, `cmdOpen: true`, `stepUpOpen: true`, `mobileIdentity` & `mobileLogin` `320<=320`, `jsErrors: []`, `badResponses: []` | PASS |
| 7 | Cek ketersediaan username via GraphQL pada Provision Modal (`/identity/users`) | Chrome Headless `systemiam` (`blur`) & `e2e.unik.2026` (`blur`) | Konflik menampilkan user `SystemIAM (SYSTEM · ACTIVE)`; username unik menampilkan status tersedia | `"Username 'systemiam' sudah digunakan oleh SystemIAM (SYSTEM · ACTIVE)."` & `"Username \`e2e.unik.2026\` tersedia."` | PASS |

## Batas Terbuka
- Skenario B (`403 Forbidden` pada user non-admin di UI `/identity`) terjaga secara kontrak oleh `identity/layout.tsx` (`!canAccessIdentity`), sementara pengujian browser live belum membuat akun tambahan pada database dev agar data tetap bersih.
- Pengecekan ketersediaan username pra-submit untuk pemanggil publik anonim (`/login` & `/register`) menunggu implementasi query `checkUsername` di IAM & Gateway sesuai catatan [`2-frontend.md`](file:///home/muriarfad/Documents/projects/erp-retail/agents/notes/iam_management/2-frontend.md) dan [`1-frontend.md`](file:///home/muriarfad/Documents/projects/erp-retail/agents/notes/gateway/1-frontend.md).
