# Verifikasi repository eksplisit HRM

Scope: repository HRM, port bernama objek/kolom, pembatasan tool model, cleanup dan dokumentasi. Dasar perubahan: owner meminta pola IAM/Enterprise Foundation tanpa embedding, nama method seperti `FindLeavePoliciesByUUID`/`DeleteLeavePoliciesByUUID`, kemudian penghapusan berkas repository yang tidak terpakai.

## Implementasi

- 86 struct konkret memakai `db *gorm.DB`. INSERT/SELECT/UPDATE/DELETE dieksekusi langsung pada repository entity, tanpa embedded atau named generic repository.
- Sembilan operasi inti bernama objek per tabel ditambah 503 finder kolom bernama objek diekspos melalui port dan registry. Alias method lama dipertahankan untuk service. UUID primary key memakai `ByUUID`, tiga identity memakai `ByID`, composite learning menyebut kedua kolom.
- `Update` memakai predicate key statis, `Select("*").Omit(PK...).Updates(snapshot)` untuk mempertahankan false/nol/NULL tanpa upsert atau perubahan key. Transaksi context, owner handle, row lock dan pemetaan error tetap dipakai.
- Helper bersama hanya menyiapkan query/model, validasi key, filter allowlist/parameter SQL, pagination, transaksi dan error; tidak mengimplementasikan CRUD generik.
- Constructor sekarang memvalidasi handle PostgreSQL; metadata model diparse saat query. Ini berbeda dari constructor generic lama yang juga memeriksa metadata/prefix/PK saat inisialisasi. Semua model tetap statis dan terikat port saat kompilasi. Kombinasi kolom ilegal + context dibatalkan dapat mengembalikan cancellation lebih dahulu; tidak melonggarkan filter.

## Cleanup

28 berkas repository dihapus: 25 `.gitkeep`, `pgstore/repository.go`, `pgstore/filter.go` dan `domain_writers.go`. Folder placeholder sudah berisi kode; generic CRUD digantikan implementasi langsung. Graph inbound `Writers` termasuk tes menunjukkan nol caller, kemudian pencarian source seluruh backend hanya menemukan deklarasi dan generatornya. Adapter tidak digunakan application service, yang sudah memvalidasi entity dan invariant melalui `operation.Resource`/Integrity. Generator adapter dan penulisan port/concrete repository di tool lama ikut dihapus agar cleanup tidak dibatalkan pada regenerasi model. `record.Writer` beserta tes domainnya tetap helper mandiri di luar scope repository.

## Perintah dan bukti aktual

Dari modul `backend/services/human_resource`:

```bash
GOWORK=off go fmt ./...
GOWORK=off go vet ./...
GOWORK=off go test -race ./...
GOWORK=off go build ./...
python3 -m py_compile tools/generate_repositories.py tools/domain_validation.py
```

Seluruhnya lulus. Tool pembangkit tidak dijalankan; `py_compile` hanya memeriksa sintaks. Semua source refactor/alias ditulis menggunakan patch langsung.

Harness verifikasi manual yang sudah tersedia pada sesi (`/tmp/hrm-repository-verification/run.py`, menjalankan Go `-race` di module sementara) disesuaikan melalui patch langsung untuk API baru, lalu dieksekusi dengan `python3 /tmp/hrm-repository-verification/run.py`. Harness membuat dua database `hrm_repository_*_test`, memasang baseline dan fixture valid, serta menghapus database dalam cleanup. PostgreSQL aktual `18.6 (Debian 18.6-1.pgdg13+2)`. Tidak menambah suite repository wajib atau endpoint API.

| Kasus | Hasil |
|---|---|
| Constructor database nil dan 86 port menolak snapshot nil | PASS |
| Pembacaan GetAll/Count dan kesamaan metadata setiap kolom 86 tabel | PASS pada kedua fixture |
| 503 finder lama dan 503 finder nama objek | PASS per fixture; hasil pasangan identik |
| Named get-all/count, scoped list/count, key read dan row lock | PASS; mencakup UUID/int64/composite pada tabel berisi data |
| Scoped filter PK lengkap sebelum pagination dan count | PASS pada seluruh repository; tabel kosong menghasilkan nol |
| Nama kolom ilegal pada SelectWhere/CountWhere | PASS ditolak pada seluruh repository |
| CRUD representatif JobFamily, false/NULL, notfound/no-upsert | PASS |
| Default/cap/negatif limit, offset, urutan PK dan slice kosong | PASS |
| Nil/canceled context, transaction owner mismatch, lock tanpa transaksi | PASS |
| Rollback lintas repository dan nested savepoint | PASS |
| 27 tabel berisi data per fixture: enam percobaan write lama/nama objek per tabel | 162 percobaan/fixture PASS; seluruh transaksi dibatalkan; constraint refusal diklasifikasikan |

Output ringkas akhir masing-masing fixture: `repositories=86 finders=503 named_finders=503 populated=27 writes=162`. Tidak ditemukan kegagalan pada verifikasi final. Kompilasi awal harness sempat memakai `SelectBy` yang tidak diekspos interface; harness diperbaiki memakai `SelectWhere`, tanpa perubahan kontrak untuk mengakomodasi tes.

## Batas bukti

Ini verifikasi manual persistensi, **bukan E2E** Gateway/gRPC/IAM, positive CRUD semua 86 tabel, pengukuran performa PostgreSQL, atau concurrency ORM menyeluruh. Tes race service memakai mock; tidak membuktikan serialisasi workflow PostgreSQL. SQL/FK/trigger tetap baseline yang sama. Adapter produksi, bootstrap, gRPC dan E2E tidak diaktifkan oleh refactor ini. Benchmark service tidak diulang karena mock tidak mengukur implementasi repository yang berubah.
