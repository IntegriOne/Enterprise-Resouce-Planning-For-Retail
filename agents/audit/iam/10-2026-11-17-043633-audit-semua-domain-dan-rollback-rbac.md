# Audit seluruh alur IAM dan rollback RBAC

Tanggal: 17 November 2026 WIB. Cakupan: 16 tabel IAM, 81 RPC pada enam service gRPC, serta jalur domain → repository → service → handler. Audit nomor 4–9 menjadi baseline inventaris; pemeriksaan ini menelusuri ulang wiring keamanan, kontrak error, transaksi mutasi RBAC, tes, dan bukti runtime. Audit seluruh source tidak berarti seluruh cabang sudah terbukti oleh E2E.

## Matriks cakupan

| Area | RPC | Hasil pemeriksaan | Batas bukti |
| --- | ---: | --- | --- |
| User dan identity | 20 | Handler terdaftar; validasi entitas, repository dan service tersambung. | Pengiriman email setelah commit belum tahan crash; semua status/identitas belum diuji sampai PostgreSQL. |
| Auth dan MFA | 12 | Handler dan service terhubung ke credential, TOTP, token dan sesi. | Persaingan MFA dan rollback lintas transaksi belum lengkap pada database nyata. |
| Session | 13 | Expiry pada UPDATE SQL dan guard sesi diuji sebelumnya; pemetaan error tersedia. | Outage/recovery Redis dua instance belum lengkap. |
| Role | 15 | Mutasi kini wajib memiliki runtime revisi dan koordinator outbox; pencabutan langsung diuji pada PostgreSQL. | Scope eksternal belum dibuktikan ke owner. |
| Group | 15 | Mutasi kini wajib memiliki runtime revisi dan koordinator outbox; pencabutan role grup diuji pada PostgreSQL. | Scope eksternal belum dibuktikan ke owner. |
| Permission | 6 | Katalog, lookup, konflik unik dan pemakaian role terhubung. | Konflik katalog belum seluruhnya diuji pada database nyata. |

## Temuan dan perbaikan

1. **Tinggi, diperbaiki:** constructor service role/group menerima `SecurityRuntime` atau `OutboxCoord` nil; helper perubahan RBAC bisa sukses tanpa revision/outbox. Kedua dependency kini wajib; `Coordinator.RecordInTx` juga mengembalikan error bila dependency internalnya tidak siap. `RevokeUserRole` dan `RevokeRole` selalu menyelesaikan ID subjek sebelum DELETE, kemudian menaikkan revisi dan mencatat outbox dalam transaksi yang sama.
2. **Bukti PostgreSQL baru:** tes integration dengan database unik `iam_disposable_*_test` menyuntik kegagalan insert outbox. Pencabutan role langsung dan role grup mengembalikan error; assignment tetap ada dan `perm_revision` tetap 1. Pada jalur sukses, kedua assignment hilang, revisi menjadi 3, dan dua event outbox tersimpan. Database dibuang setelah tes.
3. **Tinggi, terbuka:** scope non-global pada penugasan role/grup belum memvalidasi keberadaan dan status objek pada Enterprise Foundation. Belum ada kontrak owner yang siap dipanggil; UUID valid saja tidak cukup. Jangan klaim grant berscope tersebut terverifikasi.
4. **Tinggi, terbuka:** email verifikasi/reset/perubahan email masih dikirim setelah commit tanpa antrean tahan crash. Token bisa tersimpan saat pengiriman belum terjadi. Perlu outbox email atomik dan worker dengan retry/deduplikasi.
5. **Sedang, terbuka:** enam RPC daftar relasi (`ListUserRoles`, `ListEffectiveUserRoles`, `ListGroupRoles`, `ListUserGroups`, `ListRolePermissions`, `ListUserIdentities`) belum memiliki pagination kontrak dan query repository yang dibatasi.
6. **Sedang, terbuka:** tes bufconn menyentuh 81 RPC tetapi bukan bukti gRPC→PostgreSQL/Redis untuk semua cabang. Uji autentikasi/MFA bersaing, revoke dua instance, dan matriks E2E Gateway masih diperlukan.

## Verifikasi

- `go fmt ./...`, `go vet ./...`, `go build ./...`, dan `go test -race -count=1 ./...` lulus pada modul IAM.
- `go test -tags integration ./services/role/test -run TestRoleRevocationRollsBackWhenOutboxFails -count=1 -v` lulus pada PostgreSQL Docker dengan database sementara unik.
- `docker compose build iam` dan `docker compose up -d --no-deps --wait iam` lulus; container IAM berstatus healthy pada port privat lokal `127.0.0.1:9001`.
- Graph codebase-memory diperiksa untuk scope role/group/repository/handler dan path perubahan inti; tidak ada gap tercatat, tetapi hasil graph tetap sinyal best-effort.
