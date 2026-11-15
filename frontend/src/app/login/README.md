# IntegriOne login — App Router

Route: `/login`. Menggunakan Next.js **16.4.6** App Router, React 19, dan CSS Modules yang di-scope di bawah `.enterpriseShell`.

## Authentication contract (only this page)

- `GET /api/graphql?query=...` menjalankan query `browserSession { authenticated csrfToken }` untuk membentuk sesi pre-auth dan cookie HttpOnly.
- `POST /api/graphql` menjalankan mutasi `userLogin(input: {identifier,password}) {status mfaChallenge mfaExpiresAt csrfToken}` dengan header `X-CSRF-Token` dan `credentials: include`.
- Status hasil login sepenuhnya ditentukan oleh respons GraphQL: hanya `MFA_REQUIRED` yang masuk ke langkah verifikasi MFA, sedangkan `AUTHENTICATED` menyelesaikan autentikasi. Navigasi pasca-login dipandu secara ketat oleh header HTTP 200 `Location` (`/identity` untuk Super Admin, `/` untuk customer), divalidasi via allowlist `validatePostLoginLocation`.
- Menyediakan tautan pendaftaran mandiri akun menuju `/register`.
- `POST /api/graphql` menjalankan `validateUserMfa(input: {challenge,totpCode})` atau `{challenge,recoveryCode}` saat status `MFA_REQUIRED`.
- `POST /api/graphql` menjalankan `userResetPassword(input: {email})` dan `resetPassword(input: {token, newPassword})` pada alur pemulihan kata sandi (khusus akun dengan email terverifikasi).
- Login sosial (`beginCustomerOAuth`) tidak ditampilkan karena provider OAuth belum terkonfigurasi untuk produksi dan fokus portal ini adalah autentikasi enterprise email/username + kata sandi.
- Tidak menyediakan opsi persisten "30 days" karena kebijakan sesi IAM IntegriOne memakai opaque session dengan idle 30 menit dan absolut 8 jam.
- `/api/graphql` di-proxy oleh Caddy pada origin yang sama dengan frontend Next.js (`proxy/Caddyfile`).

## Design & Accessibility

- **Enterprise Editorial Aesthetics**: Sticky translucent header (`68px`), asymmetric two-column split (`1.05fr 0.95fr` pada desktop `>= 1024px`, `min-height: calc(100svh - 68px)`, radius `24-28px`). Pada mobile/tablet (`< 1024px`), form autentikasi diprioritaskan di atas (`order: 1`) dan panel editorial ringkas di bawah (`order: 2`).
- **Palet Warna Resmi IntegriOne**: Primary Navy `#133554`, Deep Navy `#0B2439`, Soft Navy `#E9EEF2`, Accent Coral `#DA414A`, Off White `#F9FAFB`, Dark Text `#17202A`, Gray `#66717C`, Light Gray `#DDE3E8`, Navbar `#F4F7F8`.
- **Tipografi**: `Fraunces` (`--font-fraunces`, optical serif) untuk *editorial display & major headings* dan `Source Sans 3` (`--font-source-sans-3`) untuk label, input, tombol, dan navigasi, dimuat melalui `next/font/google` di `frontend/src/app/layout.tsx`.
- **Komponen Khusus**: Coral brand accent rule, subtle grid + grain texture pada panel Deep Navy, oversized editorial index numeral (`01`/`02`), animated 1px link underline (`Lupa kata sandi?`), focus ring `4px rgba(19, 53, 84, 0.10)`, toggle visibilitas kata sandi (`>= 44px`), dan dukungan `prefers-reduced-motion`.
