# Audit dan perbaikan domain, repository, service HRM

Tanggal: 16 November 2026. Scope: `backend/services/human_resource/{domain,repository,internal/service}`. Acuan wiring dibaca dari IAM; aturan lokasi aktif tetap `internal/service/<domain>` sesuai AGENTS. IAM, Foundation, Gateway, DBML dan migrasi tidak diubah.

## Temuan dan perbaikan

| Lapisan | Temuan | Perbaikan dan bukti |
| --- | --- | --- |
| Domain cuti | Rebuild saldo menerima ledger di luar periode entitlement. | Tolak tanggal sebelum awal/sesudah akhir; batas inklusif. Dua tes negatif gagal sebelum fix dan lulus sesudahnya; dua tes batas lulus. Dasar: catatan skema bahwa masa ledger mengikuti entitlement. |
| Domain talent | Kalibrasi belum mencocokkan framework skor sumber dengan framework assessment. | Tambah kecocokan framework, sesuai FK komposit skema. Tes framework berbeda gagal sebelum fix dan lulus sesudahnya. |
| Repository | Tidak ditemukan kesalahan alias/CRUD/query pada inventaris dan pemeriksaan pola yang dijelaskan di bawah. | Tambah `RecordRepository[T,K]` tujuh operasi minimum; port lengkap tetap kompatibel. Tidak mengubah semantik SQL. |
| Service/list | Revoke saat CountWhere masih dapat mengembalikan halaman sukses. | Recheck actor/revisi dan context sesudah count; tes reproduksi gagal tanpa fix, lulus dengan fix dan memeriksa halaman kosong. |
| Dependency | Draft callback dan reflection registry menambah lapisan pemetaan yang berbeda dari IAM. | Draft dibuang. Registry menginjeksi repository langsung; 84 dependency memakai port minimum milik domain. Alias lama operation.Repository dipertahankan untuk kompatibilitas. |
| Code style | Constructor panjang, pengelompokan import tes, literal outbox panjang dan dua diagnostik staticcheck. | Constructor multiline, goimports/gofmt, pemisahan literal, tagged switch ledger dan selector GORM yang disederhanakan. |

## Cakupan pemeriksaan

- Sembilan package aturan agregat dan helper validasi record/value/field diperiksa terarah. Bukan pembuktian exhaustif seluruh CHECK hasil translasi DBML.
- Inventaris 86 repository konkret dan port: field `db *gorm.DB`, satu objek satu berkas, tanpa embedding konkret; 503 finder kolom lama dan 675 alias finder (503 kolom + 86 key + 86 lock) sesuai target. Read/update/delete composite key diperiksa. Immutable/FK tetap dipagari database dan guard service; hard delete mengikuti skema HRM yang tidak memiliki deleted_at.
- Helper teknis: allowlist kolom dan parameter SQL, AND antar filter/IN per filter, scope sebelum pagination/count, limit default 20/maksimum 100, guard key/RowsAffected, row lock dalam transaksi serta penggunaan context transaksi. Count/list READ COMMITTED tidak menjanjikan snapshot atomik.
- Application service 22 domain tetap memakai helper operasi bersama. Injection repository langsung tidak berarti seluruh method mengakses nama finder khusus tabel; pencarian service tetap memakai query berscope.
- Penolakan typed nil repository/transactor/policy/integrity/outbox diuji tanpa menjalankan persistensi. Mock tidak perlu mengimplementasikan finder yang tidak dikonsumsi.
- Codebase-memory digunakan sebelum edit; coverage scope domain/repository/service tidak melaporkan gap tercatat (sinyal best-effort), diikuti pembacaan source terkait.

## Verifikasi

Verifikasi akhir modul mandiri (`GOWORK=off`) lulus: `go fmt ./...`, `go vet ./...`, `go test -race ./...`, `go build ./...`, serta `golangci-lint run --no-config ./domain/... ./repository/... ./internal/service/...` (0 issues, pemeriksaan default). `goimports -l domain repository internal/service` kosong dan `git diff --check` lulus. Lint default tidak mengukur seluruh pertimbangan keterbacaan manual. Benchmark tidak diulang karena tugas ini audit/perbaikan correctness dan wiring, bukan pengukuran performa; angka lama tidak dianggap hasil baru.

Full reindex codebase-memory selesai. Gap parser SQL yang sudah ada berada di luar scope Go yang diperbaiki; bukti skema terkait invariant dibaca langsung, bukan disimpulkan dari graph parsial.

Audit repository sempat menjalankan pemeriksaan PostgreSQL ad-hoc lama sebelum dihentikan arahannya: dua database disposable `_test`, 86 port dan 503 finder lama + 503 finder bernama objek, 27 tabel berisi/162 percobaan write-rollback per fixture lulus, kedua database dibersihkan. Ini bukan suite wajib baru atau bukti E2E; tidak ada source harness/SQL yang diubah. Validasi database berikutnya harus melalui E2E sesuai AGENTS saat ini.

## Batas dan tindak lanjut

- Adapter Policy/Integrity produksi, runtime/config, gRPC dan Gateway HRM serta E2E belum tersedia; tidak ada klaim seluruh HRM selesai atau aman diaktifkan.
- Pengujian CRUD service tetap representatif satu resource per 22 domain, bukan seluruh 84 tabel bisnis.
- Recheck keamanan hanya menjamin penolakan apabila perubahan terlihat pada pemeriksaan terakhir; tidak menjamin revoke setelah check tersebut.
- Ada perbedaan label bucket ledger pada komentar SQL dan dokumentasi domain: komentar mengelompokkan adjustment/expiry/reversal ke adjusted, sedangkan kode/dokumen domain memakai expiry pada used dan reversal membalik bucket asal. Closing balance tetap konsisten. Klasifikasi bisnis ini tidak diubah diam-diam; perlu penyelarasan sebelum workflow produksi.
