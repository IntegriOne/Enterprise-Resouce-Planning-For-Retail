# Laporan E2E Frontend #1: Halaman `/login` (`userLogin`, `createUser`, `browserSession`, `myWorkspaceAccess`, & `proxy.ts`)

- **Tanggal:** 2026-11-16 18:44 WIB
- **Basis Commit Git:** `0b8d3ee` (beserta perubahan lokal teruji di `frontend/` sebelum commit akhir)
- **Lingkungan Uji:** Reverse proxy Caddy (`https://erp.localhost:8443`), `frontend-1` & `frontend-2` (Next.js `16.4.6` / Node `26.x`), `gateway-1` & `gateway-2`, `iam`, `redis`, `postgres`.
- **Operasi `/api/graphql` & Rute Teruji:**
  - `Query`: `browserSession`, `myWorkspaceAccess`
  - `Mutation`: `userLogin`, `createUser`, `userLogout`
  - Route & Proxy: `/login`, `/register`, `/verify-email`, `/identity` ([`proxy.ts`](file:///home/muriarfad/Documents/projects/erp-retail/frontend/src/proxy.ts))

## Matriks Kasus Uji

| No | Kasus Uji | Kondisi / Aktor | Hasil Ekspektasi | Hasil Aktual | Status |
|---|---|---|---|---|---|
| 1 | Unit test validasi 12-karakter, strength meter, `fetchMyWorkspaceAccessBrowser`, & `submitCreateUser` | Node `test` runner (`iam-browser-auth.test.ts` + `redirect-allowlist.test.ts`) | Seluruh `13` tes kontrak & keamanan allowlist lulus | `13/13` tests lulus (`0` fail) | PASS |
| 2 | Typegen, Lint, TypeScript, & Production Build Next.js 16 | Host `frontend/` | Zero error, zero warning, zero deprecated `middleware` warning | `next build` & `lint` exit `0`, `ƒ Proxy (Middleware)` aktif | PASS |
| 3 | Akses anonim ke `/identity` lewat `proxy.ts` | Anonim (tanpa cookie `__Host-integrione_session`) | Redirect `HTTP 307` menuju `/login` | `HTTP 307` -> `Location: /login` | PASS |
| 4 | Bootstrap `browserSession` + proteksi CSRF fail-closed | Anonim -> `POST userLogin` tanpa `X-CSRF-Token` | Ditolak `FORBIDDEN` saat `X-CSRF-Token` hilang | `errors[0].extensions.code = FORBIDDEN` | PASS |
| 5 | Login kredensial salah (`AUTHENTICATION`) | Identitas tak terdaftar | `was_successful: false`, `errors[0].type = AUTHENTICATION` | `was_successful: false`, `type = AUTHENTICATION` | PASS |
| 6 | Paritas kontrak `createUser` pada password lemah (< 12 char) | Anonim + valid CSRF | `was_successful: false`, error `ARGUMENT`, tanpa tulis user baru | `was_successful: false`, `type = ARGUMENT` | PASS |
| 7 | Login akun admin IAM (`userLogin`) -> `myWorkspaceAccess` -> `/identity` -> `userLogout` | Akun bootstrap `systemiam` (`superadmin`, kredensial lokal `.env`) | `status=AUTHENTICATED`, `redirectLocation=/identity`, `Location: /identity`, `/identity` HTTP `200`, `userLogout` bersih | Seluruh langkah berhasil (`HTTP 200`, `canAccessIdentity=true`, `defaultLocation=/identity`) | PASS |
| 8 | Rendering Chrome Headless (`1440x900` OS Dark Mode + `320x740` `prefers-reduced-motion`) | `google-chrome --headless=new` via `https://erp.localhost:8443/login` | Warna dasar `#f9fafb` (`rgb(249, 250, 251)`), `H1` `51.84px` (2 baris), tab `Daftar Akun` terender lengkap, `0` JS error, `0` HTTP >= 400, `scrollWidth <= innerWidth` | `bodyBg=rgb(249, 250, 251)`, `scrollW=1425<=1440` & `320<=320`, `badNet=[]`, `jsErrors=[]` | PASS |
