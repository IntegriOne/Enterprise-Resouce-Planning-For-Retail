# Audit ulang seluruh IAM gRPC

Tanggal: 17 November 2026 WIB. Scope aktual: `backend/services/iam`, 16 tabel pada migrasi `000001`, dan 81 RPC pada enam service. Pemeriksaan ini meneruskan audit nomor 4–11, membaca ulang kontrak, wiring, domain, repository, service, handler, serta menguji modul IAM. Seluruh RPC memiliki handler terdaftar; itu belum berarti setiap cabang bisnis lulus E2E.

## Matriks lintas lapisan

| Kelompok | RPC | Jalur yang diperiksa | Status |
| --- | ---: | --- | --- |
| User dan identity | 20 | `user.proto` → `user_handler.go` → `services/user` → repository user/identity/credential/token → tabel `users`, `user_identities`, `user_credentials`, `user_tokens` | Terhubung; pengiriman email masih tidak tahan crash. |
| Auth dan MFA | 12 | `auth.proto` → `auth_handler.go` → `services/user/auth.go`, `mfa.go` → credential, factor, recovery code, token dan session | Terhubung; race MFA pada PostgreSQL nyata belum dibuktikan lengkap. |
| Session | 13 | `session.proto` → `session_handler.go` → `services/session` → session repository, security runtime dan outbox | Terhubung; guard transaktor/outbox diperbaiki dalam audit ini. |
| Role | 15 | `role.proto` → `role_handler.go` → `services/role` → role, permission relation, user assignment, revision/outbox | Terhubung; scope eksternal gagal tertutup sampai verifikasi owner tersedia. |
| Group | 15 | `group.proto` → `group_handler.go` → `services/group` → group, membership, group role, revision/outbox | Terhubung; scope eksternal gagal tertutup. |
| Permission | 6 | `permission.proto` → `permission_handler.go` → `services/permission` → katalog permission | Terhubung; konflik katalog belum seluruhnya diuji pada PostgreSQL. |

`cmd/wiring.go` membangun repository dan registry service, sedangkan `transport/grpc/handler/register.go` mendaftarkan keenam handler dengan validasi registry. Jumlah deklarasi RPC cocok dengan jumlah method handler per kelompok: 20/12/13/15/15/6. `ErrorMapper` memetakan error domain ke status gRPC dan menyembunyikan error internal; pemeriksaan ini tidak membuktikan semua error path dari database nyata.

## Temuan dan tindakan

1. **Tinggi, diperbaiki — sesi dan outbox dapat tidak atomik bila service salah dirangkai.** `session.NewService` menerima `OutboxCoord` aktif dengan `Transactor` nil. `runInTx` kemudian menjalankan callback tanpa transaksi. Konstruktor sekarang menolak kombinasi itu; tes baru memeriksa kode error dan jalur dengan transaktor mock. Wiring produksi sudah memasok transaktor. Tes mock ini belum membuktikan rollback PostgreSQL untuk setiap mutasi sesi.
2. **Tinggi, terbuka — model identitas desain dan implementasi berbeda.** `agents/principals.md` pada HEAD menyatakan satu tabel `principals` dengan `attributes` JSONB dan alur access request/inbox, sedangkan DBML/migrasi IAM yang berjalan memakai tabel `users` tanpa `attributes`, 16 tabel, dan tidak memiliki tabel access request/inbox. Dokumen itu juga merujuk path `backend/services/iam_management` yang tidak ada pada checkout. Penyelarasan memerlukan keputusan desain/migrasi terencana; audit tidak mengubah skema IAM atau membuat tabel paralel.
3. **Tinggi, terbuka — email transaksional tidak persisten.** `services/user/email_dispatch.go` mengirim setelah transaksi dan hanya mencatat kegagalan; proses dapat berhenti setelah token tersimpan sebelum email terkirim. Perlu outbox email atomik, retry, dan deduplikasi tanpa membocorkan status akun.
4. **Sedang, terbuka — enam list relasi tidak berhalaman.** `ListUserRoles`, `ListEffectiveUserRoles`, `ListGroupRoles`, `ListUserGroups`, `ListRolePermissions`, dan `ListUserIdentities` tidak memiliki `PageRequest/PageInfo`, berbeda dari list utama. Kontrak, service, dan query repository perlu batas default 20 dan maksimum 100 sesuai `agents/how_to_code_grpc_managements.md`.
5. **Sedang, terbuka — bukti integrasi belum merata.** Tes bufconn memanggil 81 RPC dengan repository mock. Tes PostgreSQL sebelumnya telah membuktikan migrasi, expiry sesi, dan rollback revoke role/grup; belum ada matriks gRPC→PostgreSQL/Redis untuk seluruh jalur auth/MFA, session, role/group, permission, gangguan dan dua instance. `go test` yang lulus bukan bukti semua kasus D01–D12 atau E2E Gateway lulus.

## Verifikasi pada audit ini

- `GOWORK=off go fmt ./...`, `go vet ./...`, `go build ./...`, dan `go test -race -count=1 ./...` lulus dalam modul IAM.
- Coverage statement gabungan `domain`, `repository`, `services`, dan `transport/grpc/handler` dari `go test -coverpkg=...` adalah **51,8%**. Ini metrik baris eksekusi, bukan persentase alur bisnis selesai.
- `docker compose ps` menunjukkan IAM, PostgreSQL, Redis, Gateway dan komponen lain sehat. Audit ini tidak menjalankan ulang E2E HTTPS, PostgreSQL disposable, atau simulasi outage Redis; laporan audit nomor 5 dan 10 memuat bukti integration terdahulu.
- Codebase-memory digunakan untuk simbol/relasi dan coverage scope IAM; status graph `ready` tanpa gap tercatat pada scope IAM. Perbandingan kode tetap memeriksa source langsung. Berkas `agents/principals.md` sedang terhapus dalam working tree oleh perubahan lain; isi yang dibandingkan diambil dari HEAD dan tidak ditimpa.
