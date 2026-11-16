# Repository HRM

Lapisan persistensi mencakup **86 tabel** pada 22 domain bisnis serta inbox/outbox. Modul Go berada di `backend/services/human_resource/`, terisolasi dari workspace bersama selama belum ada runtime HRM. GORM mengikuti versi proyek `v1.31.2`; tidak ada migrasi, koneksi ke management lain, endpoint, seed/grant atau aktivasi Gateway pada constructor.

## Struktur dan kontrak

- `domain/entity/`: model data bertipe, UUID `string`, nullable pointer, date/timestamp `time.Time`, JSON dan numeric presisi eksak. Tag GORM diperbolehkan pada domain data model oleh aturan repository proyek. Model ini tidak berisi keputusan bisnis.
- `domain/repository/`: port tiap tabel dengan assertion implementasi saat kompilasi.
- `repository/<domain>/`: struct dengan field `db *gorm.DB`, constructor `New<Entity>Repository`, implementasi CRUD langsung dan `FindBy<Field>` bertipe. Ada 505 finder kolom, di luar `FindByID`; dua finder terbaru membaca status dan pembuat career path. Tidak memakai embedding repository generik atau pewarisan method CRUD.
- `repository/pgstore/`: helper teknis untuk context transaksi, validasi key, normalisasi pagination, allowlist filter dan pemetaan error. Eksekusi `Create`, `First`, `Find`, `Count`, `Updates` dan `Delete` berada pada repository entity masing-masing.
- `repository/registry.go`: `New(db)` membangun seluruh repository serta transactor memakai handle PostgreSQL yang sama.

Setiap repository menyediakan:

```go
Insert(ctx, record) error
Create(ctx, record) error // alias Insert
FindByID(ctx, key) (*Entity, error)
SelectByID(ctx, key) (*Entity, error) // alias FindByID
FindByIDForUpdate(ctx, key) (*Entity, error)
GetAll(ctx, limit, offset) ([]Entity, error)
SelectList(ctx, limit, offset) ([]Entity, error) // alias GetAll
Count(ctx) (int64, error)
SelectWhere(ctx, filters, limit, offset) ([]Entity, error)
CountWhere(ctx, filters) (int64, error)
Update(ctx, record) error
DeleteByID(ctx, key) error
Delete(ctx, key) error // alias DeleteByID
```

`FindBy*` dan `GetAll` mengikuti permintaan owner; alias `Select*` mempertahankan konvensi operasi database proyek. Finder non-primary-key selalu mengembalikan halaman, termasuk kode unik, sehingga kontraknya konsisten dan tidak mengasumsikan referensi IAM/worker selalu tunggal.

**Keputusan maintenance owner 16 November 2026:** seluruh repository konkret memakai pola eksplisit seperti IAM dan Enterprise Foundation. Untuk meninjau worker, buka `repository/personnel/workers_repository.go`: field DB, constructor, SELECT, INSERT, UPDATE dan DELETE ada pada tipe tersebut. Method alias hanya memanggil method lokal bertipe yang sama. Interface port boleh memakai generic untuk mendefinisikan kontrak bersama; ini tidak menyediakan atau menyembunyikan implementasi CRUD. Service dan registry menggunakan kontrak yang sama seperti sebelum refactor.

**Susunan berkas:** satu objek memakai satu `*_repository.go` dalam folder domainnya, mengikuti pola `iam/repository/user/user_repository.go` dan repository objek lain pada folder IAM yang sama. Constructor, CRUD, finder, method bernama objek dan alias kompatibilitas berada bersama. Contoh seluruh operasi kebijakan cuti ada di `repository/leave/leave_policies_repository.go`. Jangan memisahkan method ke `operations.go` atau `*_operations.go`; seluruh 26 berkas operasi terpisah sudah digabung dan dihapus. Helper teknis bersama tetap berada di `pgstore`.

**Penamaan objek dan kolom sesuai arahan lanjutan owner:** API repository juga menyebut nama objek/tabel, misalnya:

```go
store.LeavePolicy.InsertLeavePolicies(ctx, policy)
store.LeavePolicy.UpdateLeavePolicies(ctx, policy)
store.LeavePolicy.FindLeavePoliciesByUUID(ctx, policyID)
store.LeavePolicy.DeleteLeavePoliciesByUUID(ctx, policyID)
store.LeavePolicy.FindLeavePoliciesByPolicyCode(ctx, "ANNUAL-ID", 20, 0)
store.LeavePolicy.FindLeavePoliciesByLeaveTypeCode(ctx, "ANNUAL", 20, 0)
store.LeavePolicy.GetAllLeavePolicies(ctx, 20, 0)
```

`ByUUID` menunjuk kolom primary key `id` bertipe UUID, bukan kolom baru bernama `uuid`. Tiga tabel identity memakai `ByID` dengan parameter `int64`. Tabel hubungan kursus/kompetensi memakai `ByCourseIDAndCompetencyID` dengan kedua field pada `LearningCourseCompetencyKey`. Finder lain menyebut kolom nyata pada model dan mempertahankan tipe nullable/enum serta pagination. Nama objek mengikuti nama tabel tanpa awalan `hr_`, termasuk `ApplicationStageHistory` dan `InterviewFeedback` yang memang tidak berakhiran `s`.

Method bernama objek tersedia pada interface `domain/repository` dan registry, bukan hanya tipe concrete. Tiap tabel memiliki operasi bernama objek untuk insert, update, find/delete berdasarkan key, row lock, get-all, count, select/count dengan filter; seluruh 505 finder kolom juga memiliki nama objek. Alias lama dipertahankan agar service yang sudah ada tetap berjalan. Satu operasi bernama plural tetap menulis **satu** record; nama tidak berarti bulk write. Seluruh caller tetap wajib memeriksa error.

`key` umumnya UUID string; tiga tabel identity memakai `int64`. `LearningCourseCompetencyKey{CourseID, CompetencyID}` wajib membawa kedua bagian primary key. Tidak ada penghapusan atau update massal tanpa key.

Contoh setelah service menyiapkan koneksi GORM PostgreSQL yang sah:

```go
store, err := repository.New(db)
if err != nil {
    return err
}
workers, err := store.Worker.GetAllWorkers(ctx, 25, 50)
if err != nil {
    return err
}
// Gunakan workers setelah service memverifikasi permission dan scope pemanggil.
```

Finder nullable menerima pointer; nilai nil mencari `IS NULL`, bukan mengabaikan filter. Contoh `Worker.FindByIAMPrincipalReferenceID(ctx, &iamUserID, 25, 0)` hanya lookup referensi logis. Hasilnya tidak memberikan izin login atau akses. Pemilihan hubungan kerja aktif, badan hukum, waktu efektif dan otorisasi tetap tanggung jawab service.

## Semantik pagination dan write

- Default limit `20` saat limit `0`; maksimum `100`. Limit/offset negatif ditolak sebagai `hrm.invalid_input`; tidak ada jalur limit `-1` tanpa batas.
- Urutan seluruh primary key ascending, termasuk kedua kolom composite key. Halaman kosong adalah slice `[]`, bukan nil. Offset besar dapat mahal; keyset pagination belum ditambahkan.
- `Count` terpisah dari `GetAll`. Keduanya tidak menjanjikan snapshot yang sama saat ada write bersamaan; transaksi default tetap PostgreSQL READ COMMITTED. Pagination offset juga dapat bergeser saat dataset berubah.
- `Insert/Create` menyimpan seluruh field secara eksplisit. Caller mengisi UUID, status awal, default bisnis, JSON wajib, tanggal dan timestamp. Default database non-identity tidak diterapkan otomatis ke zero value model, sehingga nilai false/nol tidak tertimpa default true. Identity bernilai nol diperoleh dari PostgreSQL melalui RETURNING.
- `Update` mengganti snapshot semua kolom non-key, termasuk false, nol dan NULL. Setiap repository menulis predicate key secara eksplisit serta memakai `Select("*").Omit(<kolom primary key>).Updates(record)` agar zero value ikut tersimpan. Baca snapshot lengkap lebih dahulu; hindari struct parsial. Tidak memakai `Save`, upsert, AutoMigrate, perubahan PK atau timestamp otomatis.
- Record hilang pada read/update/delete menjadi `hrm.not_found`. Unique → `hrm.already_exists`; FK → `hrm.reference_conflict`; CHECK/exclusion/immutable → `hrm.constraint_violation`; deadlock/serialization → `hrm.conflict`. Error cancellation/deadline dipertahankan. Pesan wrapper tidak menyalin SQL atau nilai PII; cause tetap tersedia bagi kode internal.
- Delete adalah hard delete yang tunduk FK RESTRICT dan trigger. CRUD tidak menghapus aturan immutable: ledger cuti, rule payroll, framework talent dan keputusan/bukti yang dibekukan akan menolak update/delete sesuai baseline. Koreksi dilakukan oleh workflow service melalui record/revisi yang sah.
- `Decimal` memakai teks, bukan float64, agar numeric payroll dan KPI tidak kehilangan presisi. SQL NULL menggunakan `*Decimal`; kosong bukan nol. JSON nil mewakili SQL NULL, sedangkan `JSON("null")` mewakili literal JSON null.

## Transaksi

```go
err := store.Transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
    worker, err := store.Worker.FindWorkersByUUIDForUpdate(txCtx, workerID)
    if err != nil {
        return err
    }
    // Service memvalidasi perubahan worker dan menyiapkan event yang sah.
    if err := store.Worker.UpdateWorkers(txCtx, worker); err != nil {
        return err
    }
    return store.OutboxEvent.InsertOutboxEvents(txCtx, event)
})
```

Seluruh repository mengambil transaksi dari context, memeriksa handle owner yang sama, dan meneruskan cancellation. Callback gagal membatalkan perubahan; nested callback memakai savepoint GORM. `FindByIDForUpdate` menolak pemanggilan tanpa transaksi. Jangan membagikan context transaksi ke goroutine atau menggunakannya sesudah callback selesai. Tidak ada hook setelah commit atau publikasi event otomatis.

Penguncian per-baris, pemeriksaan versi dan retry konflik harus dipilih service sesuai workflow. Repository tidak menghitung payroll/cuti, memberi grant, mengevaluasi SoD, memanggil IAM/Foundation, memutuskan promosi, atau menerbitkan event sendiri. Inbox/outbox di sini baru CRUD; claim `SKIP LOCKED`, consumer dan relay belum dibuat. Scope objek/permission diperiksa sebelum memakai repository; `GetAll` bukan kontrak API publik yang terotorisasi.

## Maintenance dan pemeriksaan

Implementasi `repository/<domain>/*_repository.go` dipelihara langsung dengan patch per berkas. Tidak ada marker `DO NOT EDIT` pada implementasi konkret. Perubahan skema memerlukan penyesuaian model, port, mapper/service dan repository terkait secara eksplisit agar mudah direview.

Tool lama `tools/generate_repositories.py` tetap untuk model, enum, validasi domain dan registry; penulisan port maupun implementasi repository konkret sudah dihapus dari tool tersebut. Port dan implementasi dipelihara manual bersama nama operasi objeknya. Tool tidak boleh mengembalikan embedding atau menimpa CRUD/port manual. Menjalankan tool bukan bagian refactor ini.

```sh
python3 backend/services/human_resource/tools/generate_repositories.py
cd backend/services/human_resource
GOWORK=off go fmt ./...
GOWORK=off go vet ./...
GOWORK=off go test -race ./...
GOWORK=off go build ./...
```

Generator memeriksa kesamaan tabel/kolom baseline SQL dengan DBML serta pemetaan README sebelum menghasilkan entity, enum dan registry. Hanya hasil generasi tersebut yang mengikuti aturan file generated; port/repository konkret tetap manual. Perubahan skema mengikuti generator baseline SQL yang sudah tersedia lebih dahulu.

Verifikasi 16 November 2026 memakai PostgreSQL Docker 18.6 dan dua database sementara `hrm_repository_*_test`, kemudian dibuang. Kedua fixture integritas SQL dijalankan dengan commit pada database sementara agar data valid dapat dibaca Go; seluruh 70 penolakan SQL tetap lulus. Pemeriksaan bukan E2E service/Gateway dan tidak menjadi suite repository wajib tambahan.

| Cakupan | Bukti aktual |
|---|---|
| Model dan registry | 86 constructor/assertion; kesamaan setiap kolom model GORM dengan katalog PostgreSQL; `GetAll` dan `Count` seluruh tabel pada dua fixture. |
| Pencarian | Seluruh 503 finder bertipe dieksekusi pada masing-masing fixture, termasuk nilai nullable. |
| Snapshot fixture | Masing-masing fixture mengisi 27 tabel; read primary key dan 81 operasi write percobaan (update/delete/duplicate insert) per fixture memakai rollback. Penolakan constraint dibedakan dari kegagalan teknis. Ini bukan positive CRUD fixture seluruh 86 tabel. |
| Perilaku persistensi bersama | Create/read/update/delete berhasil; limit-offset dan urutan; halaman kosong; default/cap/negatif; false/NULL; no-upsert; notfound; injeksi nilai/nama kolom; row lock; nested savepoint dan rollback lintas repository; cancellation/nil context. |
| Unit test | Presisi decimal besar/pecahan, penolakan float dan input invalid, SQL NULL versus JSON null, kepemilikan buffer JSON; `go test -race` lulus. |

Uji service dengan mock port dan E2E query/mutation nyata dilakukan ketika service/transport tersedia. Belum ada klaim otorisasi, concurrency workflow payroll/promosi, performa, deployment atau E2E HRM selesai.

Validasi entity dan aturan agregat tersedia pada [kontrak domain](domain.md); jalur mutasi terkoordinasi berada pada [application service](service.md). Adapter `Registry.Writers()` lama telah dihapus karena tidak dipakai service aktif. Repository mentah tetap menangani persistensi; use case menjalankan validasi dan aturan agregat sebelum write.

## Verifikasi refactor dan cleanup 16 November 2026

Seluruh 86 implementasi konkret kini memiliki CRUD lokal tanpa embedding. Penamaan objek menambah sembilan operasi inti per tabel beserta 503 finder kolom bernama objek pada port dan implementasi. Method lama tetap kompatibel; tidak ada perubahan entity, DBML, migrasi atau service.

`GOWORK=off go fmt ./...`, `go vet ./...`, `go test -race ./...`, `go build ./...` lulus. Harness manual PostgreSQL Docker 18.6 pada dua database disposable mengulang seluruh pembacaan 86 tabel, 503 finder lama dan 503 finder bernama objek per fixture, serta filter list/count scoped dan penolakan nama kolom tak dikenal. Setiap fixture mengisi 27 tabel dan menjalankan 162 percobaan write lama/bernama objek dengan rollback; penolakan constraint dibedakan dari error teknis. Positive CRUD representatif, false/NULL, no-upsert, not-found, pagination, lock, nested savepoint, rollback lintas repository dan cancellation lulus. Ini belum positive CRUD seluruh tabel, uji concurrency ORM menyeluruh atau E2E Gateway.

Cleanup menghapus **28 berkas**: 25 `.gitkeep` pada folder yang sudah berisi kode, dua helper CRUD generik `pgstore/repository.go`/`filter.go`, dan adapter `domain_writers.go` yang tidak mempunyai caller pada backend aktif. Helper context/transaksi, filter, pagination dan error tetap dipakai. Tool model tidak lagi menulis port/repository konkret atau membuat adapter yang telah dihapus. [Audit refactor dan cleanup](../../../../agents/audit/human_resource/6-2026-11-16-110220-repository-eksplisit-dan-cleanup.md).
