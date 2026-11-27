# Laporan E2E Frontend #4: `/login` Bebas Menu `/verify-email` & Alur `/verify-email` Pasca-Login / Tautan Email

- **Tanggal:** 2026-11-25 01:55 WIB
- **Lingkungan Uji:** Reverse proxy Caddy (`https://erp.localhost:8443`), `frontend-1` & `frontend-2` (`Next.js 16.4.6` / `React 19.3.0`), `gateway-1` & `gateway-2`, `iam`, `postgres`, `redis`.
- **Operasi & Rute Teruji:** `GET /login`, `GET /login?mode=register`, `GET /register`, `GET /verify-email`, `GET /verify-email#token=demo.email.token`, `GET /identity` (anonim), `Mutation resendValidationEmailToken`.

## Matriks Kasus Uji

| No | Kasus Uji | Kondisi / Viewport | Hasil Ekspektasi | Hasil Aktual | Status |
|---|---|---|---|---|---|
| 1 | Unit tests, typegen, ESLint, TSC, & `next build` | Host `frontend/` | Exit `0`, `18/18` test termasuk `isUnverifiedLoginResponse` & `isUnverifiedLoginError` | `18/18` PASS, `next build` `12/12` rute sukses | PASS |
| 2 | Docker single-build & stack status | `frontend-1`, `frontend-2`, `caddy` | `Healthy`, `/register` `404`, `/verify-email` `200`, `/login` `200` | Seluruh container `Healthy`, `/register` `404`, `/verify-email` `200`, `/login` `200` | PASS |
| 3 | Penghapusan menu `/verify-email` & `/register` di `/login` | Chrome Headless `1440x900` di `/login` | `a[href*="/verify-email"]` = `0`, `a[href*="/register"]` = `0`, struktur 3 bagian tepat `["HEADER","MAIN","FOOTER"]`, fit `900px` | `verifyEmailLinks: 0`, `registerLinks: 0`, `directTags: ["HEADER","MAIN","FOOTER"]`, `scrollWidth: 1440`, `bodyHeight: 900`, `loginLeak: null` | PASS |
| 4 | Halaman `/verify-email` tanpa fragment `#token=` (mode redirect belum terverifikasi) | Chrome Headless langsung ke `/verify-email` lalu submit `#verify-resend-email` | Heading `"Verifikasi Email Diperlukan"`, input email ada, mutasi `/api/graphql` `resendValidationEmailToken` mengembalikan pesan netral sukses | `noTokenHeading: "Verifikasi Email Diperlukan"`, `hasResendInput: true`, `verifyResendStatus` sukses netral, `verifyLeak: null` | PASS |
| 5 | Halaman `/verify-email#token=...` (tautan dari email) | Chrome Headless ke `/verify-email#token=demo.email.token` | Fragment URL dibersihkan (`hash === ""`) dan tampil tombol konfirmasi `"Verifikasi Email Sekarang"` | `withTokenHeading: "Konfirmasi Verifikasi Email"`, `withTokenHashStripped: true` | PASS |
| 6 | Responsif Mobile & `prefers-reduced-motion` + runtime hygiene | Chrome Headless `320x740` + `reduced-motion` | `scrollWidth <= 320`, `0` JS error, `0` console error/warning, `0` HTTP `>= 400` | `scrollWidth: 320`, `jsExceptionsCount: 0`, `consoleErrorsCount: 0`, `badHttpResponsesCount: 0` | PASS |

## Batas Kontrak Backend (Tercatat ke IAM & Gateway)
- Pada uji end-to-end live `userLogin` untuk akun yang baru mendaftar dan belum memverifikasi email, backend IAM (`AuthenticatePassword` pada `backend/services/iam/services/user/auth.go`) dan Gateway (`normalizeAuthError` pada `backend/gateway/internal/adapter/iam/adapter.go`) saat ini masih mengembalikan error seragam `AUTHENTICATION: "invalid login credentials"` untuk menjaga anti-enumerasi, sehingga redirect otomatis `window.location.replace("/verify-email")` saat gagal login belum terpicu oleh server sebelum Owner memutuskan pengecualian kebijakan pada [`4-frontend.md`](file:///home/muriarfad/Documents/projects/erp-retail/agents/notes/iam_management/4-frontend.md) dan [`2-frontend.md`](file:///home/muriarfad/Documents/projects/erp-retail/agents/notes/gateway/2-frontend.md).
