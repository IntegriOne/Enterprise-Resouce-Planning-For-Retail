# Audit seluruh flow IAM gRPC: domain sampai handler

Tanggal: 17 November 2026 WIB. Cakupan: 81 RPC pada enam handler IAM, method publik service terkait, kontrak domain dan repository, pemetaan error, transaksi, outbox, serta jalur Gateway HTTPS untuk Query/Mutation. Inventaris berasal dari seluruh method handler dan service; setiap RPC handler memiliki eksekusi tes (tidak ada method RPC dengan cakupan 0%). Cakupan statement handler 69,4% dan service 62,3%, sehingga cabang kegagalan yang belum tercakup tetap dicatat sebagai keterbatasan.

| Domain | RPC | Flow yang diperiksa | Hasil |
| --- | ---: | --- | --- |
| User & identity | 20 | Create, GetByID, snapshot, list/pagination, profil/email/username/status, soft delete, identitas, token verifikasi | Token email sekarang terikat tepat ke identitas email; token tanpa ikatan ditolak, email terverifikasi tidak dapat diterbitkan token baru, pemilik deleted/suspended/disabled ditolak, dan permintaan token mengirim email setelah transaksi. RPC self menuntut caller terautentikasi. |
| Auth & MFA | 12 | Login sandi, tantangan TOTP/recovery, step-up, reset/change password, enrol/confirm/disable TOTP, recovery code | Jalur self menuntut caller terautentikasi; verifikasi faktor dan pencegahan replay tetap diuji. |
| Session | 13 | Issue/authenticate/CSRF/logout/revoke/list/cleanup/authorize/redirect/snapshot | `GetSessionCsrf` sekarang hanya memberi token ke caller terautentikasi dengan session ID yang sama. Recovery epoch snapshot repository gagal tertutup pada DB error atau nilai tidak valid. |
| Role | 15 | CRUD, ikatan permission, assignment user, daftar efektif | RPC CreateRole menolak `is_system=true`; pembuatan role sistem tetap hanya melalui bootstrap internal. |
| Group | 15 | CRUD, membership, role grup, daftar | Wiring dan transaksi/outbox diperiksa; tidak ada perubahan business rule tambahan. |
| Permission | 6 | Lookup/list, create/delete, daftar per role | Wiring dan validasi katalog dari audit sebelumnya diperiksa ulang. |

## Bukti

- `go fmt ./...`, `go vet ./...`, `go test -race -count=1 ./...` pada modul IAM: lulus.
- `go test -coverpkg=.../transport/grpc/handler ./transport/grpc/handler/test`: 69,4% statement; semua 81 method RPC mempunyai cakupan bukan nol.
- `go test -coverpkg=.../services/... ./...`: 62,3% statement; method service publik mempunyai cakupan bukan nol, beberapa helper bootstrap/email pada cabang tertentu belum dieksekusi.
- `docker compose build --quiet iam` dan `docker compose up -d --wait --no-deps iam`: lulus, IAM healthy.
- E2E Gateway/IAM `Test(Query|Mutation)` dengan `TestQueryPermissions` dikecualikan: lulus. Tes yang dikecualikan masih memakai filter resource lama `users`, sementara katalog sekarang `iam.users`; file Gateway sedang dikerjakan owner lain dan tidak disentuh.

## Batas dan tindak lanjut

1. IAM masih hanya memvalidasi tipe/UUID scope non-global. Eksistensi serta kelayakan legal entity, facility, organization unit, supplier, dan vendor harus dibuktikan lewat owner Enterprise Foundation sebelum assignment; ini memerlukan kontrak lintas management.
2. Pengiriman email dilakukan setelah commit. Kegagalan penyedia email dicatat tetapi belum punya outbox pengiriman yang tahan gangguan; user dapat memakai resend, tetapi delivery guarantee belum setara outbox security.
3. Tes Gateway `TestQueryPermissions` harus diubah oleh pemilik Gateway menjadi filter `iam.users`, lalu matriks E2E penuh diulang tanpa pengecualian.
4. Angka cakupan tidak membuktikan semua kombinasi kegagalan transaksi/DB; repository PostgreSQL dan Redis perlu uji fault injection lebih luas sebelum klaim produksi penuh.
