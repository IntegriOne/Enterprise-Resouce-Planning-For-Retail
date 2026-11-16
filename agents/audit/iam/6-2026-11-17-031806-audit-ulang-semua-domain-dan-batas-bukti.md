# Audit ulang seluruh domain IAM gRPC

Tanggal: 17 November 2026 WIB. Source aktif: `backend/services/iam`. Pemeriksaan ini melanjutkan audit nomor 4 dan 5, dengan inventaris ulang DBML, 81 RPC, service, repository, handler, wiring, serta tes. Kata "seluruh" berarti semua **enam kelompok domain** masuk cakupan; hasil mock tidak disamakan dengan pembuktian transaksi nyata.

## Matriks alur bisnis

| Domain | RPC | Alur domain → repository → service → handler yang ditelusuri | Hasil |
| --- | ---: | --- | --- |
| User dan identity | 20 | `users`, `user_identities`, `user_credentials`, `user_tokens`; create, status, profil, identitas, verifikasi, list, snapshot | Terhubung pada `services/user`, `UserHandler`, dan repository terpisah. Status nonaktif serta identitas belum terverifikasi ditolak saat login. Token email terikat ke identity. Pengiriman email pascacommit masih tidak memiliki antrean tahan gangguan. |
| Auth dan MFA | 12 | Password, lockout, token challenge, TOTP, recovery code, reset password, step-up | Terhubung pada `services/user`, `AuthHandler`, dan repository credential/faktor/token/sesi. Konsumsi token/kode memakai UPDATE bersyarat atomik; uji konkurensi PostgreSQL end-to-end belum ada. |
| Session | 13 | Terbit, authenticate, CSRF, revoke, list, authorize, redirect, snapshot | Terhubung pada `services/session`, `SessionHandler`, repository sesi dan runtime security. Ditemukan UPDATE aktivitas/MFA yang hanya mengecek `ended_at`; sudah diperbaiki dengan syarat idle dan absolute expiry pada SQL UPDATE. Tes PostgreSQL terisolasi membuktikan sesi aktif berubah, sesi idle/absolute expired tidak berubah. |
| Role | 15 | CRUD role, ikatan permission, penugasan user, role efektif | Terhubung pada `services/role`, `RoleHandler`, repository role/user-role/role-permission dan outbox. Mutasi grant/revoke menaikkan revisi izin. Validitas objek scope non-global di owner Enterprise Foundation belum dibuktikan. |
| Group | 15 | CRUD grup, anggota, penugasan role grup | Terhubung pada `services/group`, `GroupHandler`, repository grup/member/group-role dan outbox. Penambahan/pencabutan mengubah revisi izin pengguna. Validitas objek scope non-global dan rollback PostgreSQL seluruh variasi belum dibuktikan. |
| Permission | 6 | Lookup, list, create, delete, permission role | Terhubung pada `services/permission`, `PermissionHandler`, repository permission. Delete izin yang masih dipakai role ditolak. Rekonsiliasi katalog lama pada database development belum diperiksa ulang. |

## Temuan yang diperbaiki pada audit ini

`repository/session/session_repository.go`: `UpdateActivity` dan `UpdateMFAVerified` kini menolak sesi yang berakhir secara alami **pada saat UPDATE**, sehingga hasil pembacaan lama tidak dapat memperpanjang sesi kedaluwarsa atau menandainya lolos MFA. Cap `last_seen_at` dan `idle_expires_at` juga hanya dapat maju saat dua permintaan datang tidak berurutan. Error dikembalikan sebagai `iam.session.idle_expired` atau `iam.session.absolute_expired` untuk baris yang masih ada; baris yang sudah dicabut tetap `already_ended`.

Tes baru `repository/session/session_repository_integration_test.go` memakai gerbang `IAM_TEST_DATABASE_DISPOSABLE=1` dan nama `iam_disposable_*_test`. Kasus aktif, idle expired, absolute expired, dan permintaan aktivitas lama lulus pada PostgreSQL 18 container dengan migrasi `000001`; database dibuat dan dihapus khusus eksekusi ini.

## Bukti verifikasi dan batas klaim

- `go fmt ./...`, `go vet ./...`, `go test -race -count=1 ./...` pada modul IAM: lulus.
- Tes repository sesi PostgreSQL `-tags integration -race -count=1`: lulus tiga kasus; pengujian ini tidak memakai `iam_test` bersama.
- Audit nomor 4 menunjukkan semua 81 method handler dieksekusi oleh tes bufconn, namun cakupan statement handler 69,4% dan service 62,3%. Hal tersebut **belum** membuktikan seluruh cabang kesalahan, rollback, atau kontrak Redis/Gateway.
- Audit nomor 5 membuktikan migrasi serta bootstrap/login/outbox/outage/recovery pada PostgreSQL dan Redis nyata, tetapi belum semua RPC melewati database nyata.
- `agents/boundary.md` dan `docs/03-managements/iam.md` masih memuat path `iam_management` lama, sedangkan source aktif ada di `backend/services/iam`; file dokumen itu sedang memiliki perubahan kerja lain dan tidak diubah dalam audit ini.

## Sisa temuan yang perlu diselesaikan

1. **Tinggi:** validasi eksistensi/status target `legal_entity`, `organization_unit`, `facility`, `supplier_party`, dan `vendor_party` saat assignment role belum terhubung ke kontrak pembuktian Enterprise Foundation. IAM saat ini menegakkan tipe/format UUID serta `self`, tetapi UUID objek bisnis palsu masih bisa disimpan sebagai scope.
2. **Tinggi:** email verifikasi/reset dikirim setelah commit tanpa outbox pengiriman; kegagalan provider dicatat, tetapi pekerjaan kirim tidak bertahan bila proses berhenti setelah commit.
3. **Sedang:** transaksi PostgreSQL dan Redis lintas instance untuk race MFA, rollback grant/revoke, serta propagasi revisi belum dibuktikan melalui seluruh jalur gRPC.
4. **Sedang:** fixture Gateway `TestQueryPermissions` masih memakai resource `users`, sedangkan katalog aktif `iam.users`; pemilik Gateway perlu menyelaraskan fixture dan mengulang matriks E2E tanpa pengecualian.

Kesimpulan: seluruh domain telah **diaudit pada level alur dan wiring**, satu celah sesi diperbaiki dan diuji nyata. Status IAM **belum layak dinyatakan seluruh business flow tervalidasi end-to-end** sampai butir sisa di atas selesai.
