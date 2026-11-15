# Konteks Aktif IAM Management (`backend/services/iam` & `backend/shared/authsession`)

## Apa yang sudah dikerjakan

- **Pemisahan Modul Self-Service (`*self`), Auto-Assignment Saat Pembuatan Akun, & Fail-Closed `UserRoles`:**
  - Mendefinisikan `5` kode permission eksplisit di `backend/shared/authz` & [bootstrap.go](file:///home/muriarfad/Documents/projects/erp-retail/backend/services/iam/services/bootstrap.go): `iam.users.readself`, `iam.users.updateself`, `iam.users.deleteself`, `iam.sessions.readself`, dan `iam.sessions.deleteself`.
  - `CreateUser` ([service.go](file:///home/muriarfad/Documents/projects/erp-retail/backend/services/iam/services/user/service.go)) otomatis menugaskan peran `customer` (untuk `CUSTOMER`, `ScopeSelf{ID: user.ID}`, redirect `"/"`) atau `iam_selfservice` (untuk `EMPLOYEE`, `VENDOR`, `SUPPLIER`, `ScopeSelf{ID: user.ID}`, redirect netral `""`) secara atomik di transaksi DB yang sama sejak status awal `PENDING`.
- **Blokir Layanan `*self` Saat Status Masih `PENDING` atau Identitas Belum Terverifikasi (*Defense-in-Depth*):**
  - Meski peran baseline sudah terpasang saat akun `PENDING`, seluruh operasi `*self`, penerbitan sesi, dan login tetap diblokir hingga verifikasi identitas (`is_verified = true`) dan status `ACTIVE` terpenuhi:
    - [assignment.go](file:///home/muriarfad/Documents/projects/erp-retail/backend/services/iam/domain/entity/assignment.go) (`AccessSnapshot.HasPermission`): menolak bila `s.UserStatus != UserStatusActive`.
    - [service.go](file:///home/muriarfad/Documents/projects/erp-retail/backend/services/iam/services/session/service.go) (`Issue`): memuat user dan menegakkan `u.CanAuthenticate()` sebelum menerbitkan sesi.
    - [service.go](file:///home/muriarfad/Documents/projects/erp-retail/backend/services/iam/services/user/service.go) (`loadAuthenticatableUser` & `UpdateProfile`), [identity.go](file:///home/muriarfad/Documents/projects/erp-retail/backend/services/iam/services/user/identity.go) (`DeletePendingEmail`), [auth.go](file:///home/muriarfad/Documents/projects/erp-retail/backend/services/iam/services/user/auth.go) (`VerifySessionMFA`, `ChangePassword`), dan [mfa.go](file:///home/muriarfad/Documents/projects/erp-retail/backend/services/iam/services/user/mfa.go) (`EnrollTOTP`, `ConfirmTOTP`, `DisableTOTP`, `RegenerateRecoveryCodes`): menegakkan `u.CanAuthenticate()`.
    - Menambah unit test guard di `backend/shared/authz`, `domain/entity/test`, `services/session/test`, dan `services/user/test`.
- **Verifikasi Penuh:**
  - `go fmt`, `go vet`, dan `GOWORK=off go test -race -count=1 ./...` lulus pada `backend/shared/authz` (`1` package) dan seluruh `20` package bertest `backend/services/iam`.
  - Container `iam` terbangun sehat bersama 8 container Compose, dan live HTTPS E2E (`76` top-level test / `155` subtest / `0` fail / `0` skip / `68.048s`) lulus seluruhnya.

## Apa yang belum dikerjakan

- Konfigurasi deployment produksi (`APP_ENV=production` dan mTLS internal gRPC) pada lingkungan staging/production nyata.
- Integrasi layanan domain hilir lain (Enterprise Foundation, HRM, Procurement, Commerce) terhadap event/lookup IAM mengikuti fase berikutnya.

## Apa yang harus dikerjakan selanjutanya

1. Lanjutkan integrasi lintas management sesuai prioritas program di [docs/README.md](file:///home/muriarfad/Documents/projects/erp-retail/docs/README.md) menggunakan principal tunggal IAM, batas RBAC `*self` vs internal staf, dan kontrak gRPC privat yang terverifikasi.
