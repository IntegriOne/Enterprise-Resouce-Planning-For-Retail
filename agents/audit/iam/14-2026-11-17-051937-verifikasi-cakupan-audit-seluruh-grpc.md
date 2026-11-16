# Verifikasi cakupan audit seluruh IAM gRPC

Tanggal: 17 November 2026 WIB. Scope: `backend/services/iam`, enam service dan 81 RPC. Laporan ini menguji ulang klaim audit nomor 13; tidak menyatakan 81 alur bisnis sudah lulus E2E.

## Cakupan yang terbukti

| Kelompok | RPC | Bukti saat ini | Batas bukti |
| --- | ---: | --- | --- |
| User dan identity | 20 | Kontrak, handler, service, repository dan tes bufconn terhubung. | Email verifikasi/reset dikirim setelah commit tanpa outbox tahan crash; `ListUserIdentities` tanpa pagination. |
| Auth dan MFA | 12 | Handler dan service diuji dengan mock serta race detector. | Race MFA pada PostgreSQL nyata dan gangguan Redis belum dibuktikan untuk seluruh cabang. |
| Session | 13 | Handler dan service diuji; tes PostgreSQL terpisah membuktikan expiry update. | Seluruh operasi sesi belum diuji gRPC sampai PostgreSQL/Redis dua instance. |
| Role | 15 | Handler dan service diuji; tes PostgreSQL terdahulu membuktikan rollback revoke jika outbox gagal. | Scope objek eksternal masih fail closed; tiga list relasi tanpa pagination. |
| Group | 15 | Handler dan service diuji; rollback revoke sebelumnya dibuktikan pada PostgreSQL. | Scope objek eksternal masih fail closed; dua list relasi tanpa pagination. |
| Permission | 6 | Handler dan service diuji; `ListPermissionsByRole` ditelusuri sampai `SelectByRoleID`. | List tersebut tanpa pagination; konflik katalog belum seluruhnya diuji pada PostgreSQL. |

Jumlah RPC kontrak kembali dihitung 81. Tes bufconn menyentuh method, tetapi banyak cabang menggunakan mock dan tidak membuktikan penyimpanan atau kegagalan dependency. Tidak ada perubahan source IAM pada verifikasi ini.

## Temuan yang masih harus diperbaiki

1. **Tinggi:** email token yang dibuat dalam transaksi dapat tidak terkirim bila proses berhenti sebelum `emailSender.Send`; `email_dispatch.go` hanya mencatat kegagalan. Perlu outbox persisten atomik, retry, dan deduplikasi.
2. **Sedang:** tujuh RPC relasi tanpa `PageRequest`/`PageInfo`: `ListUserIdentities`, `ListUserGroups`, `ListGroupRoles`, `ListRolePermissions`, `ListUserRoles`, `ListEffectiveUserRoles`, `ListPermissionsByRole`. Perubahan kontrak, repository dan Gateway harus dikoordinasikan agar pemanggil lama tidak menerima hasil terpotong diam-diam.
3. **Tinggi:** model `principals` dan access request pada keputusan arsitektur belum selaras dengan 16 tabel migrasi IAM berbasis `users`. Penyelarasan perlu migrasi dan keputusan desain; perubahan working tree pada `agents/principals.md` milik pekerjaan lain tidak disentuh.
4. **Tinggi:** tes integrasi lintas seluruh 81 RPC belum tersedia. Prioritas berikutnya: gRPC→PostgreSQL untuk auth/MFA, session, role/group, permission; fault injection Redis dan dua instance; kemudian E2E Gateway pada checkout yang stabil.
5. **Tinggi:** assignment scope non-global/non-self sengaja ditolak sampai owner objek menyediakan verifier. Fail closed ini benar, tetapi use case belum aktif.

## Verifikasi ulang

- `GOWORK=off go test -race -count=1 ./...` lulus pada modul IAM.
- `GOWORK=off go vet ./...` dan `go build ./...` lulus.
- `docker compose ps iam postgres redis`: ketiganya `healthy` pada saat pemeriksaan.
- Dua tes dengan build tag `integration` dan database PostgreSQL disposable unik lulus: `TestSessionUpdatesRejectExpiredRows` (aktif, idle expired, absolute expired) dan `TestRoleRevocationRollsBackWhenOutboxFails` (direct role, group role). Database sementara dibersihkan setelah tes.
- E2E HTTPS, outage Redis dan dua instance tidak dijalankan ulang. Karena itu status akhir audit adalah **seluruh RPC terinventarisasi dan tes unit/race lulus; seluruh business flow belum terverifikasi atau selesai diperbaiki**.
