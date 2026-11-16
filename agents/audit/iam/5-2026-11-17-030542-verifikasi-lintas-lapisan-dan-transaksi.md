# Audit IAM lintas lapisan dan transaksi

Tanggal: 17 November 2026 WIB. Cakupan: DBML/migrasi, entity dan policy domain, port dan implementasi repository, lima service, enam handler gRPC, wiring, serta tes IAM. Inventaris 16 tabel DBML dan 81 RPC: User/Identity 20, Auth/MFA 12, Session 13, Role 15, Group 15, Permission 6. Matriks fungsi dan perubahan business rule sebelumnya ada pada audit IAM nomor 4. Audit ini menilai **bukti alur nyata**, bukan hanya keberadaan handler dan hasil tes mock.

| Domain | Domain → repository → service → handler | Bukti dan batas |
| --- | --- | --- |
| User/identity | Kontrak identitas, credential, token, dan status tersambung ke `services/user` dan `UserHandler`. | Tes mock/bufconn lulus; rollback token/sesi bersama PostgreSQL dan pengiriman email tahan gangguan belum dibuktikan. |
| Auth/MFA | Password, TOTP, recovery, serta penerbitan sesi tersambung ke repository atomik. | Tes serial lulus; konsumsi satu challenge/kode pada request serentak dan rollback jika penerbitan sesi gagal belum diuji di PostgreSQL. |
| Session | Opaque token, CSRF, snapshot, revoke, dan redirect tersambung. | Login, outbox, Redis outage/recovery lulus pada PostgreSQL+Redis nyata dengan DB sementara; TTL idle/absolut dan revoke lintas instance belum diuji end-to-end. |
| Role | CRUD, permission binding, assignment, dan revision/outbox tersambung. | Tes mock lulus; fault injection kenaikan `perm_revision`, rollback grant, dan pembuktian scope non-global ke Foundation belum lengkap. |
| Group | CRUD, membership, assignment role, dan revision/outbox tersambung. | Tes mock lulus; rollback/invalidation dari database nyata dan pembuktian scope target belum lengkap. |
| Permission | Katalog, lookup/list, create/delete, dan referensi role tersambung. | Tes service/handler lulus; rekonsiliasi data lama dan uji constraint PostgreSQL pada operasi terlarang belum lengkap. |

## Temuan yang diperbaiki

1. **Callback pascacommit dari savepoint gagal masih dapat dieksekusi** saat transaksi luar commit. `repository/pgstore/tx.go` kini menghapus callback yang ditambahkan oleh transaksi bersarang yang rollback. Tes SQLite `repository/pgstore/test/tx_after_commit_test.go` membuktikan callback rollback tidak terpanggil, sementara callback transaksi luar dan savepoint sukses tetap berjalan.
2. **Tes live berpotensi mereset database `_test` bersama.** Tes migrasi kini hanya melakukan rollback bila `IAM_TEST_DATABASE_DISPOSABLE=1` dan nama database `iam_disposable_*_test`. Tes PostgreSQL+Redis memakai gerbang yang sama sebelum `TRUNCATE`.

## Bukti dan risiko tersisa

- `go fmt ./...`, `go vet ./...`, `go test -race -count=1 ./...` lulus pada modul IAM.
- Tes migrasi rollback/up dan tes outbox/bootstrap/outage/recovery PostgreSQL+Redis TLS lulus pada database `iam_disposable_*_test` yang dibuat dan dihapus untuk tiap eksekusi; namespace Redis tes terpisah.
- Semua 81 RPC memiliki tes bufconn, tetapi transactor bufconn adalah mock tanpa commit/rollback. Ini **belum** membuktikan seluruh business flow dan error handling PostgreSQL. Prioritas berikutnya: tes RPC→PostgreSQL untuk rollback outbox/session; concurrent MFA; grant/revoke role dan group dengan revisi Redis; scope owner Foundation; E2E dua aktor dan gangguan dependency.
- **Insiden tes audit:** eksekusi pertama keliru memakai konfigurasi `.env` karena DSN sementara kosong, lalu menjalankan rollback/up pada `iam_test` bersama. Skema 17 tabel sudah kembali dan `users` kini 0 baris; isi sebelumnya tidak diketahui. Tidak ada klaim bahwa data uji lama pulih. Gerbang disposable di atas mencegah pengulangan lewat suite ini.
