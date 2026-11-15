# 10 - Kardinalitas Identitas 1 Email + 1 Username, `UpdateUsername`, dan Lookup Per-Kind

Tanggal: 2026-11-15 07:29 WIB

## Ringkasan Perubahan
- Memastikan setiap akun maksimal memiliki `1` `username` dan `1` `email` aktif (plus maksimal `1` *pending email* sementara saat proses ganti email) lewat indeks unik parsial PostgreSQL dan pemetaan constraint di [`pgstore/errors.go`](file:///home/muriarfad/Documents/projects/erp-retail/backend/services/iam/repository/pgstore/errors.go).
- Menambahkan lookup eksplisit [`FindUserIdentityByEmailKind`](file:///home/muriarfad/Documents/projects/erp-retail/backend/services/iam/repository/user/identity_repository.go) dan `FindUserIdentityByUsernameKind` pada repository serta memakainya di `CreateUser`, `RegisterCustomer`, `AddIdentity`, `UpdateEmail`, `UpdateUsername`, dan `AuthenticatePassword`.
- Mengimplementasikan [`UpdateUsername`](file:///home/muriarfad/Documents/projects/erp-retail/backend/services/iam/services/user/identity.go) beserta RPC `UpdateUserUsername` sehingga perubahan username berlaku instan tanpa verifikasi email, segera membebaskan username lama bagi pengguna lain, dan menolak username identik (`ErrSameUsername`).

## Verifikasi
- `go fmt ./...`, `go vet ./...`, dan `go test -race -count=1 ./...` pada `backend/services/iam` seluruhnya `PASS`.
- Indeks parsial disinkronkan pada DB `iam` & `iam_test`, image `iam` direbuild (`healthy`), dan live E2E HTTPS (`71` test / `140` subtest) seluruhnya `PASS`.
