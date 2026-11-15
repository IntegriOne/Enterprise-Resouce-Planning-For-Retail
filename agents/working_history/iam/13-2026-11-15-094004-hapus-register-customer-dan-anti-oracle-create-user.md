# Working History IAM #13: Penghapusan `RegisterCustomer`, Anti-Oracle `CreateUser`, dan `ResendValidationEmailToken`

**Tanggal:** 2026-11-15 09:40 WIB

## Ringkasan Perubahan
- **`ResendValidationEmailToken` Kasus 1:** Menambahkan error `ErrEmailAlreadyVerified` (`iam.identity.email_already_verified` -> `"your email has been verified"`) ketika pengguna terautentikasi tidak memiliki *pending email* dan email utamanya sudah terverifikasi; kasus tidak login tetap netral anti-enumerasi.
- **Urutan Konflik & Kebijakan Anti-Oracle `CreateUser`:** Menyelaraskan prioritas pengecekan `username` sebelum `email` pada `ensureEmailAndUsernameAvailable` (`services/user/identity.go`), `pickIdentityConflict` (`repository/user/user_repository.go`), dan `mapIdentityUniqueConstraint` (`repository/pgstore/errors.go`) dengan pesan `"username has been taken"`, serta memastikan `Service.CreateUser` mencatat outbox keamanan transaksi dan mengirim email verifikasi pada user `PENDING`.
- **Penghapusan `RegisterCustomer`:** Menghapus RPC `RegisterCustomer`, pesan proto di `user.proto`, `Service.RegisterCustomer`, serta handler gRPC; seluruh tes diselaraskan ke `CreateUser` + `TestMapError_IdentityUniqueField`.

## Verifikasi
- `cd backend/services/iam && GOWORK=off go test -race -count=1 ./...` -> `0` (`20` package bertest `PASS`).
