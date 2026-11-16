# Audit seluruh alur IAM gRPC: domain sampai handler

Tanggal: 17 November 2026 WIB. Source aktif: `backend/services/iam`. Audit ini melanjutkan matriks 81 RPC pada audit nomor 4–7. Inventaris skema berisi 16 tabel; kontrak gRPC berisi 81 RPC. Pemeriksaan dilakukan pada domain, port repository, implementasi PostgreSQL, service, handler, wiring, dan tes yang tersedia. `codebase-memory` dipakai untuk penelusuran struktur; cakupan source IAM lengkap menurut metadata indeks, sedangkan migrasi SQL yang `parse_partial` dibaca langsung saat audit sebelumnya.

| Kelompok | RPC | Pemeriksaan alur dan hasil |
| --- | ---: | --- |
| User dan identity | 20 | Create, lookup/list, profil, status, identitas, verifikasi email, dan snapshot terhubung ke `services/user` serta repository `users`, `user_identities`, `user_credentials`, `user_tokens`. Validasi status/identitas dan rollback pembuatan akun ada. Pengiriman email setelah commit masih belum memakai antrean tahan gangguan. |
| Auth dan MFA | 12 | Password, lockout, token reset, TOTP, recovery code, dan step-up sesi terhubung ke service user dan repository credential/faktor/token/sesi. Konsumsi token/kode bersyarat ada; balapan MFA melalui PostgreSQL nyata belum dibuktikan lengkap. |
| Session | 13 | Issue, authenticate, CSRF, revoke, list, authorize, redirect, snapshot terhubung ke service sesi dan repository. UPDATE aktivitas/MFA memeriksa expiry pada SQL menurut audit nomor 6. Tes repository PostgreSQL terisolasi ada. |
| Role | 15 | CRUD, ikatan izin, grant/revoke user, dan peran efektif memakai transaksi dan revisi izin/outbox. Scope `self` divalidasi terhadap user. Scope non-global milik Foundation baru diperiksa bentuk UUID, belum eksistensi/status target. |
| Group | 15 | CRUD, keanggotaan, grant/revoke role grup memakai transaksi dan revisi izin/outbox. Scope `self` untuk role grup ditolak. Target scope Foundation belum dibuktikan. |
| Permission | 6 | Lookup/list/create/delete dan izin per role terhubung; delete izin yang dipakai role ditolak. List utama memakai pagination dan urutan deterministik. |

## Perbaikan yang dibuat

`services/bootstrap.go` sebelumnya mengabaikan error penyegaran snapshot akun bootstrap, error aktivasi akun yang sudah ada, dan error lookup UUID akun bootstrap. Sekarang ketiganya diteruskan sehingga startup tidak melaporkan bootstrap berhasil bila status keamanan belum dapat dipastikan. Tiga tes injeksi kegagalan ditambahkan pada suite handler/bufconn untuk memeriksa setiap jalur.

## Bukti dan batas

- Seluruh 81 method handler dieksekusi oleh tes bufconn yang sudah ada; ini membuktikan jalur transport dan mock, bukan seluruh cabang transaksi PostgreSQL/Redis nyata.
- `go fmt ./...`, `go vet ./...`, `go build ./...`, `go test -race -count=1 ./...`, tiga tes injeksi bootstrap, dan `docker compose build iam` lulus. Container IAM yang sedang berjalan tidak direstart; perubahan bootstrap ini belum diuji melalui E2E live.
- Tes PostgreSQL sesi terisolasi dan tes PostgreSQL+Redis bootstrap/outbox pernah lulus pada audit nomor 5–6; tidak dijalankan ulang untuk audit ini.
- Belum ada matriks gRPC→PostgreSQL nyata untuk seluruh 81 RPC, khususnya rollback grant/revoke, race MFA, dan propagasi revisi antar instance. Karena itu status seluruh business flow **belum terverifikasi end-to-end**.

## Temuan terbuka

1. Tinggi: pemberian `user_roles` dan `group_roles` berscope `legal_entity`, `organization_unit`, `facility`, `supplier_party`, atau `vendor_party` belum memverifikasi eksistensi, status, dan relasi target lewat kontrak owner Foundation. UUID sintaksis saja dapat tersimpan. Perlu adapter baca privat/kontrak scope owner, tanpa query lintas database.
2. Tinggi: email verifikasi/reset dikirim setelah commit tanpa outbox pengiriman yang tahan crash. Token dapat tersimpan ketika proses berhenti sebelum email terkirim.
3. Sedang: bukti PostgreSQL/Redis nyata untuk race MFA, rollback RBAC, dan invalidasi antar instance belum lengkap; suite mock tidak menggantikannya.
4. Sedang: fixture Gateway `TestQueryPermissions` memakai resource lama `users`, sedangkan katalog sekarang `iam.users`; pemilik Gateway perlu memperbarui fixture dan menjalankan E2E penuh.

Kesimpulan: semua enam kelompok domain dan 81 RPC masuk inventaris serta audit wiring. Satu kelompok error bootstrap diperbaiki. IAM belum boleh dinyatakan selesai sepenuhnya sampai empat temuan terbuka dibuktikan atau diselesaikan.
