# Audit semua alur bisnis IAM gRPC

Tanggal: 17 November 2026 WIB. Cakupan: `backend/services/iam`, skema 16 tabel, enam service gRPC dan 81 RPC. Audit ini memeriksa keterhubungan domain → repository → service → handler, invariant penting, penanganan error, serta bukti tes. Audit nomor 4–8 dipakai sebagai riwayat; kelulusan tes mock tidak dianggap bukti transaksi nyata.

## Matriks seluruh domain

| Domain/RPC | Domain dan repository | Service dan handler | Status bukti |
| --- | --- | --- | --- |
| User/identity (20) | Entitas user, credential, identity, token, dan akses snapshot; repository melakukan operasi pilih/tulis dengan pemetaan error. | Create, profil, status, identitas dan verifikasi tersambung ke handler. Pembuatan akun, credential, baseline role, dan token berada dalam transaksi. | RPC tersentuh bufconn; jalur PostgreSQL seluruh status/identitas belum terbukti. |
| Auth/MFA (12) | Credential, TOTP, recovery code dan token sekali pakai tersimpan melalui port terpisah. | Login memeriksa status aktif, identitas terverifikasi, lockout, lalu menerbitkan challenge/sesi. Handler menerjemahkan error domain. | Unit/mock ada; balapan MFA dan rollback token/sesi pada PostgreSQL belum lengkap. |
| Session (13) | Sesi opaque, expiry, activity, CSRF dan outbox; operasi activity/MFA memakai predicate expiry SQL. | Issue/authenticate/revoke/authorize menghubungkan sesi, snapshot dan runtime. | Tes repository sesi PostgreSQL tersedia; E2E seluruh cabang Redis/outage/revoke antar instance belum ada. |
| Role (15) | Role, role-permission, user-role, scope dan security revision. | CRUD, assign/revoke dan effective role terhubung. Scope `self` mengikat ID pengguna. | Mutasi tercakup mock; scope non-global belum diverifikasi ke Foundation. |
| Group (15) | Group, member, group-role dan security revision. | CRUD, anggota, role grup dan revoke terhubung. Scope `self` pada grup ditolak. | Mutasi tercakup mock; scope non-global dan invalidasi lintas instance belum terbukti. |
| Permission (6) | Katalog izin, lookup dan FK pemakaian role; delete izin terpasang ditolak. | Create/read/list/delete dan list per role terhubung. | Bufconn ada; pengujian database untuk semua konflik katalog belum lengkap. |

## Temuan prioritas

1. **Tinggi — grant scope eksternal belum divalidasi pemilik objek.** `role.Service.AssignToUser` dan `group.Service.AssignRole` memanggil `entity.NewScope`, yang memeriksa tipe/UUID tetapi tidak eksistensi, status atau hubungan legal entity/unit/facility/party pada Enterprise Foundation. Grant dengan UUID valid namun tidak sah dapat tersimpan. Perlu lookup privat Foundation sebelum commit dan kebijakan kegagalan dependency yang fail-closed.
2. **Tinggi — email transaksional tidak tahan crash.** `CreateUser` menyimpan token dalam transaksi, kemudian `dispatchVerificationEmail` dipanggil setelah commit; `sendTransactionalMessage` hanya mencatat kegagalan pengiriman. Proses berhenti pada celah ini meninggalkan token tanpa pekerjaan pengiriman yang persisten. Pola sama perlu diaudit pada reset password dan perubahan email. Solusi: outbox email atomik, worker dengan retry dan deduplikasi.
3. **Tinggi — port runtime keamanan masih opsional pada mutasi RBAC.** `RevokeUserRole` dan `group.RevokeRole` menghapus assignment walau `SecurityRuntime == nil`, lalu melewati kenaikan revision karena ID subjek tidak diketahui. Helper `bumpAndRecord*` juga bisa sukses tanpa revision/outbox. Wiring produksi saat ini memasok port tersebut, tetapi service dapat dipakai tanpa jaminan yang sama. Jadikan dependency wajib untuk operasi grant/revoke atau sediakan port resolusi yang selalu tersedia, lalu uji bahwa kegagalannya me-rollback transaksi.
4. **Sedang — beberapa list relasi tidak dibatasi.** `ListUserRoles`, `ListEffectiveRoles`, `ListGroupRoles`, `ListUserGroups`, `ListRolePermissions`, dan `ListIdentities` mengembalikan slice penuh tanpa `limit/offset`; repository memakai query tanpa limit. RPC list utama (user/role/group/permission/session) sudah berhalaman. Tambahkan pagination kontrak pada list relasi agar ukuran respons dan biaya query terkendali.
5. **Sedang — klaim kelengkapan tes harus dipersempit.** 81 RPC telah dipanggil pada bufconn, tetapi tidak semuanya diuji sampai PostgreSQL/Redis. Cakupan statement gabungan domain/repository/service/handler dari `go test -coverpkg=...` adalah **51,1%** pada eksekusi ini; angka itu bukan cakupan skenario keamanan. Kasus yang perlu dibuktikan: rollback grant/revoke, concurrent MFA, expiry/CSRF, Redis down/recovery, dan dua instance yang memakai revisi akses berbeda.

## Pemeriksaan aktual dan batas

- `go vet ./...` lulus; `go test -race -count=1 ./...` lulus pada modul IAM.
- `go test -count=1 -coverpkg=./domain/...,./repository/...,./services/...,./transport/grpc/handler/... -coverprofile=/tmp/iam-audit-coverpkg.out ./domain/... ./repository/... ./services/... ./transport/grpc/handler/...` lulus; `go tool cover -func` menghasilkan total 51,1%.
- Audit ini membaca source untuk alur user/auth, session, role, group, permission, mapper error, repository transaksi dan pemetaan error. Sebagian pembuktian detail 81 RPC mengacu pada audit nomor 4–8; **belum ada bukti E2E nyata untuk semua business flow**. Tidak ada perubahan source IAM atau tes live pada audit ini.
- Metadata codebase-memory menunjukkan scope IAM tanpa gap tercatat; ini sinyal best-effort, bukan jaminan isi setiap cabang terindeks.

## Urutan perbaikan

1. Kunci mutasi RBAC agar revision/outbox wajib dan tambahkan tes rollback pada repository nyata.
2. Implementasikan verifikasi scope target melalui kontrak Foundation; gagal tertutup saat owner tidak tersedia.
3. Tambahkan outbox email dan pagination list relasi.
4. Jalankan matriks gRPC→PostgreSQL/Redis dan E2E Gateway tanpa fixture lama, lalu perbarui status hanya berdasarkan hasil nyata.
