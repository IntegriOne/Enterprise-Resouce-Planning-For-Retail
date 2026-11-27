# IntegriOne Login — App Router

Route: `/login`. Menggunakan Next.js **16.4.6** App Router, React 19 (`reactCompiler: true`), GSAP (`@gsap/react`), Tailwind CSS v4, dan CSS Modules yang di-scope di bawah `.enterpriseShell`.

Layout `/login` dibatasi ketat menjadi **3 bagian**: Navbar (`<header>`), Login Section (`<main>` -> 1 `<section>`), dan Footbar (`<footer>`).

## Struktur Modul

- `page.tsx`: Server Component metadata dan shell route `/login`.
- `login-experience.tsx`: State container autentikasi & pendaftaran, sinkronisasi CSRF/sesi, validasi redirect, serta animasi transisi & entrance hero via GSAP (`useGSAP` + `gsap.matchMedia()`).
- `auth-forms.tsx`: Komponen formulir modular (`CredentialsForm`, `RegisterForm`, `RegisterPendingView`, `MfaForm`, `ResetRequestForm`, `ResetConfirmForm`, `ResendVerificationForm`, `StatusStepView`).
- `login-footbar.tsx`: Footer editorial bertumpu di bawah halaman (`<footer>`, latar Deep Navy `#0B2439`) dengan grid asimetris 3 kolom (identitas brand & jaminan keamanan sesi, navigasi akses akun ke `/`, pergantian view daftar/masuk & lupa kata sandi di `/login`, serta ringkasan kebijakan sesi) — tetap tanpa menu `/verify-email` ataupun `/register`.
- `icons.tsx`: Ikon SVG geometris dan mark logo IntegriOne.

Mendukung parameter opsional `/login?mode=register` untuk membuka tampilan pendaftaran langsung pada rute `/login` tanpa rute `/register` terpisah, serta pengalihan `window.location.replace("/verify-email")` bila server mengembalikan sinyal akun belum terverifikasi saat login.

## Kontrak `/api/graphql` (Same-Origin)

- `GET /api/graphql?query=...` -> `browserSession { authenticated principalId csrfToken }` untuk bootstrap cookie pre-auth dan `X-CSRF-Token`.
- `GET /api/graphql?query=...` -> `myWorkspaceAccess { canAccessIdentity defaultLocation }` apabila pengunjung membuka `/login` dengan sesi yang sudah terautentikasi (`authenticated: true`), menggantikan redirect statis lama.
- `POST /api/graphql` -> `userLogin(input: { identifier, password })`: mendukung email maupun username (mis. `adminiam`), menangani `AUTHENTICATED` maupun `MFA_REQUIRED`, memvalidasi `Location` / `redirectLocation` lewat `validatePostLoginLocation`.
- `POST /api/graphql` -> `createUser(input: { full_name, email, username, password })`: mendaftarkan akun pelanggan baru (`CUSTOMER`, status awal `PENDING`) dengan paritas aturan `PasswordPolicy` backend (**12–128 karakter**, huruf besar, huruf kecil, angka, simbol, tanpa bagian identitas) serta penanganan anti-enumerasi ketika email sudah terdaftar (`redirectLocation === "/login"`).
- `POST /api/graphql` -> `resendValidationEmailToken`, `validateUserMfa`, `userResetPassword`, `resetPassword`.
