# Repository HRM

Lapisan persistensi mencakup **86 tabel** pada 22 domain bisnis serta inbox/outbox. Modul Go berada di `backend/services/human_resource/`, terisolasi dari workspace bersama selama belum ada runtime HRM. GORM mengikuti versi proyek `v1.31.2`; tidak ada migrasi, koneksi ke management lain, endpoint, seed/grant atau aktivasi Gateway pada constructor.

## Struktur dan kontrak

- `domain/entity/`: model data bertipe, UUID `string`, nullable pointer, date/timestamp `time.Time`, JSON dan numeric presisi eksak. Tag GORM diperbolehkan pada domain data model oleh aturan repository proyek. Model ini tidak berisi keputusan bisnis.
- `domain/repository/`: port tiap tabel dengan assertion implementasi saat kompilasi.
- `repository/<domain>/`: constructor `New<Entity>Repository` dan `FindBy<Field>` bertipe sesuai kolom referensi, kode, nomor, status dan idempotensi. Ada 503 finder kolom, di luar `FindByID`.
- `repository/pgstore/`: CRUD bersama, pagination, pemetaan error dan transaksi lokal.
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
Update(ctx, record) error
DeleteByID(ctx, key) error
Delete(ctx, key) error // alias DeleteByID
```

`FindBy*` dan `GetAll` mengikuti permintaan owner; alias `Select*` mempertahankan konvensi operasi database proyek. Finder non-primary-key selalu mengembalikan halaman, termasuk kode unik, sehingga kontraknya konsisten dan tidak mengasumsikan referensi IAM/worker selalu tunggal.

`key` umumnya UUID string; tiga tabel identity memakai `int64`. `LearningCourseCompetencyKey{CourseID, CompetencyID}` wajib membawa kedua bagian primary key. Tidak ada penghapusan atau update massal tanpa key.

Contoh setelah service menyiapkan koneksi GORM PostgreSQL yang sah:

```go
store, err := repository.New(db)
if err != nil {
    return err
}
workers, err := store.Worker.GetAll(ctx, 25, 50)
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
- `Update` mengganti snapshot semua kolom non-key, termasuk false, nol dan NULL. Baca snapshot lengkap lebih dahulu; hindari struct parsial. Tidak memakai `Save`, upsert, AutoMigrate, perubahan PK atau timestamp otomatis.
- Record hilang pada read/update/delete menjadi `hrm.not_found`. Unique → `hrm.already_exists`; FK → `hrm.reference_conflict`; CHECK/exclusion/immutable → `hrm.constraint_violation`; deadlock/serialization → `hrm.conflict`. Error cancellation/deadline dipertahankan. Pesan wrapper tidak menyalin SQL atau nilai PII; cause tetap tersedia bagi kode internal.
- Delete adalah hard delete yang tunduk FK RESTRICT dan trigger. CRUD tidak menghapus aturan immutable: ledger cuti, rule payroll, framework talent dan keputusan/bukti yang dibekukan akan menolak update/delete sesuai baseline. Koreksi dilakukan oleh workflow service melalui record/revisi yang sah.
- `Decimal` memakai teks, bukan float64, agar numeric payroll dan KPI tidak kehilangan presisi. SQL NULL menggunakan `*Decimal`; kosong bukan nol. JSON nil mewakili SQL NULL, sedangkan `JSON("null")` mewakili literal JSON null.

## Transaksi

```go
err := store.Transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
    worker, err := store.Worker.FindByIDForUpdate(txCtx, workerID)
    if err != nil {
        return err
    }
    // Service memvalidasi perubahan worker dan menyiapkan event yang sah.
    if err := store.Worker.Update(txCtx, worker); err != nil {
        return err
    }
    return store.OutboxEvent.Insert(txCtx, event)
})
```

Seluruh repository mengambil transaksi dari context, memeriksa handle owner yang sama, dan meneruskan cancellation. Callback gagal membatalkan perubahan; nested callback memakai savepoint GORM. `FindByIDForUpdate` menolak pemanggilan tanpa transaksi. Jangan membagikan context transaksi ke goroutine atau menggunakannya sesudah callback selesai. Tidak ada hook setelah commit atau publikasi event otomatis.

Penguncian per-baris, pemeriksaan versi dan retry konflik harus dipilih service sesuai workflow. Repository tidak menghitung payroll/cuti, memberi grant, mengevaluasi SoD, memanggil IAM/Foundation, memutuskan promosi, atau menerbitkan event sendiri. Inbox/outbox di sini baru CRUD; claim `SKIP LOCKED`, consumer dan relay belum dibuat. Scope objek/permission diperiksa sebelum memakai repository; `GetAll` bukan kontrak API publik yang terotorisasi.

## Regenerasi dan pemeriksaan

```sh
python3 backend/services/human_resource/tools/generate_repositories.py
cd backend/services/human_resource
GOWORK=off go fmt ./...
GOWORK=off go vet ./...
GOWORK=off go test -race ./...
GOWORK=off go build ./...
```

Generator memeriksa kesamaan tabel/kolom baseline SQL dengan DBML serta pemetaan README sebelum menghasilkan entity, enum, port, wrapper dan registry. File generasi tidak diedit manual. Implementasi bersama dan value object tetap ditulis manual. Perubahan skema mengikuti generator baseline SQL yang sudah tersedia lebih dahulu.

Verifikasi 16 November 2026 memakai PostgreSQL Docker 18.6 dan dua database sementara `hrm_repository_*_test`, kemudian dibuang. Kedua fixture integritas SQL dijalankan dengan commit pada database sementara agar data valid dapat dibaca Go; seluruh 70 penolakan SQL tetap lulus. Pemeriksaan bukan E2E service/Gateway dan tidak menjadi suite repository wajib tambahan.

| Cakupan | Bukti aktual |
|---|---|
| Model dan registry | 86 constructor/assertion; kesamaan setiap kolom model GORM dengan katalog PostgreSQL; `GetAll` dan `Count` seluruh tabel pada dua fixture. |
| Pencarian | Seluruh 503 finder bertipe dieksekusi pada masing-masing fixture, termasuk nilai nullable. |
| Snapshot fixture | Masing-masing fixture mengisi 27 tabel; read primary key dan 81 operasi write percobaan (update/delete/duplicate insert) per fixture memakai rollback. Penolakan constraint dibedakan dari kegagalan teknis. Ini bukan positive CRUD fixture seluruh 86 tabel. |
| Perilaku persistensi bersama | Create/read/update/delete berhasil; limit-offset dan urutan; halaman kosong; default/cap/negatif; false/NULL; no-upsert; notfound; injeksi nilai/nama kolom; row lock; nested savepoint dan rollback lintas repository; cancellation/nil context. |
| Unit test | Presisi decimal besar/pecahan, penolakan float dan input invalid, SQL NULL versus JSON null, kepemilikan buffer JSON; `go test -race` lulus. |

Uji service dengan mock port dan E2E query/mutation nyata dilakukan ketika service/transport tersedia. Belum ada klaim otorisasi, concurrency workflow payroll/promosi, performa, deployment atau E2E HRM selesai.
