# 16. Blokir Self-Service IAM Saat Status PENDING dan Belum Terverifikasi

**Tanggal:** 15 November 2026 (17:29 WIB)

Memperketat pertahanan berlapis pada domain & service IAM agar peran self-service (`customer` / `iam_selfservice`) yang otomatis terpasang saat akun dibuat tetap tidak dapat digunakan selama status masih `PENDING` atau identitas belum terverifikasi (`is_verified = false`):

- [assignment.go](file:///home/muriarfad/Documents/projects/erp-retail/backend/services/iam/domain/entity/assignment.go): `AccessSnapshot.HasPermission` kini mewajibkan `UserStatus == UserStatusActive`.
- [service.go](file:///home/muriarfad/Documents/projects/erp-retail/backend/services/iam/services/session/service.go): `Session.Issue` memvalidasi `u.CanAuthenticate()` sebelum menerbitkan sesi baru.
- [service.go](file:///home/muriarfad/Documents/projects/erp-retail/backend/services/iam/services/user/service.go), [identity.go](file:///home/muriarfad/Documents/projects/erp-retail/backend/services/iam/services/user/identity.go), [auth.go](file:///home/muriarfad/Documents/projects/erp-retail/backend/services/iam/services/user/auth.go), [mfa.go](file:///home/muriarfad/Documents/projects/erp-retail/backend/services/iam/services/user/mfa.go): menambah helper `loadAuthenticatableUser` pada `UpdateProfile`, `DeletePendingEmail`, `VerifySessionMFA`, `ChangePassword`, `EnrollTOTP`, `ConfirmTOTP`, `DisableTOTP`, dan `RegenerateRecoveryCodes`.
- Menambah unit test `-race` di `shared/authz`, `domain/entity/test`, `services/session/test`, dan `services/user/test`, serta rebuild container `iam`.
