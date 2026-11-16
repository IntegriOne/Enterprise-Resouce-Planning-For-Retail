# E2E Gateway/IAM setelah hardening gRPC

Tanggal: 17 November 2026 WIB. Jalur: HTTPS `/api/graphql` → Gateway → IAM gRPC yang direstart.

| Case | Hasil | Bukti |
| --- | --- | --- |
| `TestMutationUserLogin`, email dan username berhasil | Lulus | `go test -tags=e2e ./internal/e2e -run '^TestMutationUserLogin$' -count=1` |
| `TestMutationUserLogin`, MFA diperlukan dan kredensial salah ditolak | Lulus | Perintah yang sama; kedua subtest selesai tanpa gagal |
| Seluruh `TestQuery*` dan `TestMutation*`, termasuk sesi/RBAC/MFA | Gagal 1 case | `go test -tags=e2e ./internal/e2e -run '^Test(Query|Mutation)' -count=1`: hanya `TestQueryPermissions/FilterByResourceSensitiveAndPagination` gagal |
| `TestQueryPermissions`, filter `resource="users"` dan `is_sensitive=true` | Gagal karena fixture usang | Respons `total_count=0`; katalog IAM memakai resource `iam.users` setelah koreksi namespace |
| Seluruh `TestQuery*` dan `TestMutation*` kecuali `TestQueryPermissions` | Lulus | Perintah matriks dengan `-skip '^TestQueryPermissions$'` selesai `ok` dalam 36,511 detik |
| Login, assignment user role, penambahan anggota grup, dan assignment group role setelah rebuild terakhir | Lulus | `go test -tags=e2e ./internal/e2e -run '^TestMutation(UserLogin|AssignUserRole|AddGroupMember|AssignGroupRole)$' -count=1` setelah IAM direstart |

Test Gateway pada case gagal sedang berada pada working tree management lain sehingga belum diubah dalam task IAM. Ulangi matriks penuh setelah fixture filter resource diselaraskan oleh pemilik Gateway. Tidak ada bukti di sini untuk kelayakan target scope lintas Enterprise Foundation.
