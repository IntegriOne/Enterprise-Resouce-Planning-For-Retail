# Permintaan Keputusan Owner & Kontrak IAM: Sinyal Login Akun Belum Terverifikasi (`/verify-email`)

## Task
Menentukan kontrak gRPC `AuthService.AuthenticatePassword` apabila pengguna yang baru mendaftar mencoba login dengan kata sandi yang valid, namun status akun masih `PENDING` atau email identitasnya belum terverifikasi (`!ident.IsVerified`).

## Temuan Saat Ini
- Pada `backend/services/iam/services/user/auth.go` (baris 106–117), `AuthenticatePassword` memverifikasi kata sandi terlebih dahulu (`s.passwordHasher.Verify`), lalu mengembalikan `iamerrors.ErrInvalidCredentials` ketika `u.CanAuthenticate() != nil` (`PENDING`) maupun ketika `!ident.IsVerified`.
- Akibatnya, baik UI `/login` maupun Gateway tidak dapat membedakan kata sandi salah vs akun valid yang emailnya belum terverifikasi untuk mengalihkan browser ke `/verify-email` (`window.location.replace("/verify-email")`).

## Action (Menunggu Persetujuan Owner / Arsitektur)
1. **Validasi Kebijakan Keamanan (Anti-Enumerasi):** Konfirmasikan kepada Owner apakah akun yang lolos pemeriksaan kata sandi (`matched == true`) namun berstatus `PENDING` / `!ident.IsVerified` diizinkan mengembalikan sinyal terpisah (mis. error khusus `ErrEmailVerificationRequired` atau `AuthStatus_AUTH_STATUS_EMAIL_VERIFICATION_REQUIRED`), mengingat respons berbeda pasca-password dapat dimanfaatkan sebagai *password/account-status verification oracle* bila dilakukan *credential stuffing*.
2. Bila disetujui Owner, terbitkan status/error eksplisit tersebut dari `AuthenticatePassword` khusus saat `!u.IsDeleted()`, `!cred.IsLocked(now)`, **dan** `matched == true`.

## Context
- UI Frontend (`/login` & `/verify-email`) sudah siap menangani `status: "EMAIL_VERIFICATION_REQUIRED"` / `redirectLocation: "/verify-email"` maupun kode error `EMAIL_NOT_VERIFIED` / `IDENTITY_NOT_VERIFIED` dengan `window.location.replace("/verify-email")`, serta menyediakan formulir kirim ulang tautan verifikasi di `/verify-email`.
- Sesuai batas ownership, Frontend tidak mengubah `backend/services/iam/`.
