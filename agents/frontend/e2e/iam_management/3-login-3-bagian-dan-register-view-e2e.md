# Laporan E2E Frontend #3: Validasi `/login` 3 Bagian & Penghapusan Rute `/register`

- **Tanggal:** 2026-11-25 01:40 WIB
- **Lingkungan Uji:** Reverse proxy Caddy (`https://erp.localhost:8443`), `frontend-1` (`127.0.0.1:3333`) & `frontend-2` (`127.0.0.1:3334`) (`Next.js 16.4.6` / `React 19.3.0`), `gateway-1` & `gateway-2`, `iam`, `enterprise-foundation`, `redis`, `postgres`.
- **Rute & Endpoint Teruji:** `GET /login`, `GET /login?mode=register`, `GET /register`, `GET /verify-email`, `GET /identity` (anonim), `GET /api/graphql?query={health}`.

## Matriks Kasus Uji

| No | Kasus Uji | Kondisi / Viewport | Hasil Ekspektasi | Hasil Aktual | Status |
|---|---|---|---|---|---|
| 1 | Unit tests, typegen, ESLint, TSC, & `next build` | Host `frontend/` | Exit `0`, `17/17` unit test PASS, `/register` hilang dari route tree (`12/12` rute) | `17/17` PASS, `12/12` static/partial routes terkompilasi tanpa `/register` | PASS |
| 2 | Docker single-build & stack `Healthy` | `frontend-1`, `frontend-2`, `caddy` | `Healthy`, `/api/graphql?query={health}` -> `ok`, `GET /register` -> `404`, `GET /login` -> `200` | Seluruh container `Healthy`, `health: ok`, `/register` `404`, `/login` `200` | PASS |
| 3 | Struktur 3 bagian `/login` desktop & fit fold | Chrome Headless `1440x900` di `/login` | `enterpriseShell` memiliki tepat anak `["HEADER","MAIN","FOOTER"]`, `1` `<section>` di `<main>`, nol section tambahan/link `/register`, `scrollWidth <= 1440`, `bodyHeight <= 900`, nol kebocoran tipe user | `directTags: ["HEADER","MAIN","FOOTER"]`, `topHeaderCount: 1`, `sectionInMainCount: 1`, `footerCount: 1`, `registerLinks: 0`, `forbiddenNodes: []`, `scrollWidth: 1440`, `bodyHeight: 900`, `loginLeak: null` | PASS |
| 4 | Pergantian view pendaftaran di `/login` (Navbar, Tab Card, Footbar) | Chrome Headless `1440x900` di `/login` | Klik tombol/tab Daftar Akun membuka tampilan `"Buat akun baru."` tanpa meninggalkan `/login`; klik kembali membuka `"Selamat datang kembali."` | `navbarToggle`, `toggledToRegister`, `toggledBackToLogin` seluruhnya tetap `pathname: "/login"` (`regLeak: null`) | PASS |
| 5 | Parameter `/login?mode=register` | Chrome Headless langsung ke `/login?mode=register` | Membuka view pendaftaran dengan struktur 3 bagian yang sama | `href: "/login?mode=register"`, heading `"Buat akun baru."`, `#reg-username` tampil, `3` anak & `1` `<section>` | PASS |
| 6 | Audit `/verify-email` & akses anonim `/identity` | Chrome Headless desktop `1440x900` | `/verify-email` bebas link `/register` & bebas bocor tipe user; anonim `/identity` redirect ke `/login` | `/verify-email` `registerLinks: 0`, `verifyLeak: null`; `/identity` -> `/login` | PASS |
| 7 | Responsif Mobile & `prefers-reduced-motion` + Runtime hygiene | Chrome Headless `320x740` + `prefers-reduced-motion: reduce` | `scrollWidth <= 320`, struktur 3 bagian utuh, `0` JS error, `0` console warning/error, `0` HTTP `>= 400` | `scrollWidth: 320`, `directTags: ["HEADER","MAIN","FOOTER"]`, `sectionCount: 1`, `jsExceptionsCount: 0`, `consoleErrorsCount: 0`, `badHttpResponsesCount: 0` | PASS |

## Batas Uji
- Alur login `SystemIAM` -> `/identity/*` & live check username terverifikasi lengkap pada [`2-dashboard-identity-dan-check-username-e2e.md`](file:///home/muriarfad/Documents/projects/erp-retail/agents/frontend/e2e/iam_management/2-dashboard-identity-dan-check-username-e2e.md); pada run verifikasi kali ini fokus skrip browser otomatis diarahkan ke validasi anonim `/login` & `/register` (`404`) karena variabel env kredensial seed admin tidak diekspor pada shell runner.
