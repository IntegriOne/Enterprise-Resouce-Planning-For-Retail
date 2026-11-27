# Permintaan Keputusan Owner & Kontrak Gateway: Redirect Login Belum Terverifikasi (`/verify-email`)

## Task
Mengekspos hasil `userLogin` ke Frontend apabila pengguna mencoba masuk sebelum memverifikasi alamat emailnya agar Frontend dapat melakukan `window.location.replace("/verify-email")`.

## Temuan Saat Ini
- Pada `backend/gateway/graphql/iam/iam_auth.graphql`, enum `UserLoginStatus` saat ini hanya memuat `AUTHENTICATED` dan `MFA_REQUIRED`.
- Pada `backend/gateway/internal/adapter/iam/auth_adapter.go` (baris 106–108) dan `adapter.go` (baris 254–264), `normalizeAuthError` menormalkan seluruh kegagalan login (termasuk identitas belum terverifikasi) menjadi `AUTHENTICATION: "invalid login credentials"`, yang dijaga ketat oleh `TestUserLoginAndValidateMFAAmbiguousErrorsAndThrottle` pada `relay_test.go`.

## Action (Menunggu Persetujuan Owner & Arsitektur)
1. Bila Owner menyetujui pengecualian kebijakan anti-enumerasi khusus akun `PENDING` / `!IsVerified` **setelah** kecocokan kata sandi tervalidasi di IAM (`agents/notes/iam_management/4-frontend.md`), tambahkan status enum `EMAIL_VERIFICATION_REQUIRED` pada `UserLoginStatus` di `iam_auth.graphql` dengan `redirectLocation: "/verify-email"` (tanpa menerbitkan cookie sesi terautentikasi).
2. Jika kebijakan anti-enumerasi penuh tetap dipertahankan Owner, konfirmasikan hal ini ke tim Frontend agar alur verifikasi tetap aman via pesan netral + tombol *Kirim ulang email verifikasi*.

## Context
- Menu navigasi ke `/verify-email` telah dihapus dari Navbar & Footbar Frontend sesuai arahan Owner; halaman `/verify-email` sudah mendukung mode tanpa `#token=` untuk pengalihan pasca-login, sementara [`login-experience.tsx`](file:///home/muriarfad/Documents/projects/erp-retail/frontend/src/app/login/login-experience.tsx) siap mengeksekusi `window.location.replace("/verify-email")` segera setelah sinyal tersebut tersedia dari `/api/graphql`.
