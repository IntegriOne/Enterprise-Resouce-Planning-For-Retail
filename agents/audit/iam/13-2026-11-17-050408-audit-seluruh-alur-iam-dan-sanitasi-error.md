# Audit seluruh alur IAM gRPC dan sanitasi error

Tanggal: 17 November 2026 WIB. Scope source aktual: `backend/services/iam`. Audit ini meneruskan audit 4–12 dan memeriksa ulang keenam kontrak proto, 81 method handler, registry/wiring, service, port domain, repository, migrasi 16 tabel, dan aturan D01–D12. Istilah **terhubung** di bawah berarti jalur kode tersedia; belum berarti setiap cabang bisnis lulus E2E PostgreSQL dan Redis.

## Matriks domain sampai handler

| Domain | RPC | Pemeriksaan alur | Hasil |
| --- | ---: | --- | --- |
| User dan identity | 20 | `user.proto` → `user_handler.go` → `services/user/{service,identity,registration_runtime}.go` → repository user/identity/credential/token → `users`, `user_identities`, `user_credentials`, `user_tokens` | Handler terdaftar; pembuatan akun, identitas, credential, baseline self role dan token verifikasi berjalan dalam transaksi. Pengiriman email setelah commit masih tidak persisten. `ListUserIdentities` belum berhalaman. |
| Auth dan MFA | 12 | `auth.proto` → `auth_handler.go` → `services/user/{auth,mfa}.go` → credential, TOTP, recovery code, token, sesi | Login memeriksa status/identitas, password, lockout; MFA memakai pembaruan langkah TOTP atomik. Bukti race MFA pada PostgreSQL nyata belum lengkap. |
| Session | 13 | `session.proto` → `session_handler.go` → `services/session/{service,runtime_sync}.go` → repository sesi, revision/outbox dan Redis | Penerbitan, touch, revoke dan evaluasi izin terhubung; konstruktor menolak outbox tanpa transaktor. Bukti rollback dan outage untuk **semua** operasi sesi belum lengkap. |
| Role | 15 | `role.proto` → `role_handler.go` → `services/role` → role/permission/user role/revision/outbox | Assignment memeriksa user, role aktif, scope dan duplikasi; mutasi grant mencatat revisi/outbox dalam transaksi. Scope yang memerlukan fakta owner tetap gagal tertutup. Tiga list relasi belum berhalaman. |
| Group | 15 | `group.proto` → `group_handler.go` → `services/group` → group/member/group role/revision/outbox | Membership dan assignment terhubung; revision/outbox satu transaksi. Dua list relasi belum berhalaman. |
| Permission | 6 | `permission.proto` → `permission_handler.go` → `services/permission` → katalog permission | Kode permission divalidasi dan duplikasi dicek sebelum insert; constraint SQL menjaga race. `ListPermissionsByRole` belum berhalaman. |

Jumlah handler `func (h *...Handler)` adalah 81 dan cocok dengan deklarasi RPC 20/12/13/15/15/6. `RegisterAll` memvalidasi registry lalu mendaftarkan keenam service. Interceptor menuntut bukti HMAC untuk RPC bisnis di luar `APP_ENV=test`; health dan reflection dikecualikan. Pemeriksaan statis ini tidak membuktikan bahwa setiap keputusan otorisasi Gateway sesuai scope objek bisnis.

## Temuan dan perbaikan

1. **Tinggi, diperbaiki — status gRPC berisi detail internal dapat diteruskan mentah.** `ErrorMapper.ToStatus` semula mengembalikan semua error yang dikenali `status.FromError`. Untuk error status yang dibungkus, API gRPC memasukkan teks wrapper ke pesan status. Tes regresi mereproduksi bocornya `secret_token_idx` dan `password=topsecret`. Cabang passthrough dihapus; error tak dikenal kini menjadi `Internal` dengan pesan `internal error` dan `ErrorInfo` aman. Tes terarah gagal sebelum perbaikan dan lulus sesudahnya.
2. **Tinggi, terbuka — skema identitas runtime dan rancangan principal berbeda.** Migrasi IAM berjalan memakai `users` dan tidak memuat `principals.attributes` atau access request/inbox. Dokumen `agents/principals.md` sedang dihapus oleh perubahan working tree lain; tidak disentuh. Penyelarasan membutuhkan keputusan skema dan migrasi terencana, bukan tabel identitas paralel.
3. **Tinggi, terbuka — email verifikasi/reset tidak tahan crash.** Token disimpan dalam transaksi lalu email dikirim setelah commit melalui `email_dispatch.go`. Crash di sela keduanya dapat meninggalkan token valid tanpa email. Perlu outbox email atomik, retry dan deduplikasi.
4. **Sedang, terbuka — tujuh RPC list relasi belum berhalaman.** `ListUserIdentities`, `ListUserGroups`, `ListGroupRoles`, `ListRolePermissions`, `ListUserRoles`, `ListEffectiveUserRoles`, **dan `ListPermissionsByRole`** tidak menyediakan `PageRequest`/`PageInfo`; repository terkait mengembalikan seluruh hasil. Audit 12 hanya mencatat enam dan melewatkan `ListPermissionsByRole`. Kontrak dan seluruh pemanggil Gateway harus diselaraskan sebelum memberlakukan batas 20/100 agar hasil tidak terpotong diam-diam.
5. **Tinggi, terbuka — bukti lintas runtime belum merata.** Tes mock/bufconn menyentuh seluruh 81 RPC dan tes PostgreSQL terdahulu membuktikan beberapa transaksi/expiry. Belum ada matriks gRPC→PostgreSQL→Redis untuk setiap cabang auth/MFA, sesi, grant/revoke, outage, dan dua instance. Keputusan D01–D12 belum dapat dinyatakan sepenuhnya VERIFIED.
6. **Tinggi, terbuka — target scope eksternal belum dapat dibuktikan.** Assignment role pengguna dan grup pada scope non-global/non-self ditolak `Unavailable` sebelum tulis. Ini benar sebagai fail-closed; aktivasi memerlukan kontrak verifikasi objek dari service pemilik seperti dicatat pada `agents/notes/enterprise_foundation/1-iam.md`.

## Verifikasi tugas ini

- Tes regresi `TestErrorMapper_ToStatus/wrapped_grpc_status...` gagal sebelum perbaikan dengan status `Unavailable` dan pesan berisi detail internal, lalu lulus setelahnya.
- `GOWORK=off go test -race -count=1 ./...`, `go vet ./...`, dan `go build ./...` lulus pada modul IAM; `git diff --check` lulus.
- Docker, E2E HTTPS, PostgreSQL disposable, dan outage Redis **tidak** dijalankan ulang pada audit ini. Perubahan source hanya pada pemetaan error dan tesnya. Working tree Gateway, docs dan service lain sedang berubah, sehingga hasil runtime lintas service tidak boleh disimpulkan dari unit test IAM.
