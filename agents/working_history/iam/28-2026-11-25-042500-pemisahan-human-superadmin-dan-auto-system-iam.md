# Working History IAM — Pemisahan Human Superadmin dan Auto System IAM

- **Tanggal:** 2026-11-25 04:25:00 WIB
- **Cakupan:** Konfigurasi `backend/services/iam/.env`, `.env.example`, `docker-compose.yml`, IAM gRPC service (`config`, `domain`, `repository`, `services`, `cmd`), serta proteksi & E2E `backend/gateway`.

## Ringkasan Perubahan
1. Memisahkan akun human administrator awal `superadmin` (`user_kind = employee`, kredensial baru pada variabel `IAM_SUPERADMIN_*` di `.env`) dari akun otomasi non-interaktif `auto_system_iam` (`user_kind = system`, UUID lama `853aac77-...`, role `iam_automation`) dan `auto_system_ef` (`ef_automation`).
2. Menambahkan rekonsiliasi bootstrap idempoten (`BootstrapPrincipalsAndRBAC`) yang mengonversi `systemiam` historis ke `auto_system_iam` sambil menghapus email, kata sandi (`DeleteByUserID`), faktor TOTP/recovery code, sesi aktif, serta role/grup non-kompatibel.
3. Menolak login, sesi browser, reset kata sandi, TOTP, redirect UI, dan kredensial interaktif untuk `UserKindSystem` (`ErrSystemUserInteractiveForbidden`), serta memblokir assignment role/grup lintas `user_kind` (`ErrRoleUserKindMismatch`) dan menambah 10 permission `enterprise_foundation.*` untuk `superadmin`.
4. Memperbarui mapping error dan guard `*_automation` di Gateway serta menyelaraskan E2E suite.
