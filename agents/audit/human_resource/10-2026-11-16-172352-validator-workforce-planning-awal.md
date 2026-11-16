# Audit validator workforce planning awal

Tanggal: 16 November 2026. Tujuan: menutup sebagian gap `NewIntegrity(nil)` tanpa membuka resource yang belum memiliki pembuktian referensi.

## Bukti

- `BuildDependencies` memasang validator secara fail-fast setelah repository HRM terbentuk. `hr_job_families` dan `hr_jobs` menjadi dua resource pertama yang memiliki validator produksi; 76 resource berkode mutasi lainnya tetap ditolak `ErrUnavailable`. Enam resource riwayat/hasil hanya baca.
- Kode unik dicek sebelum insert/update. Job mengunci job family melalui `FindByIDForUpdate` dan menolak parent tidak aktif. Delete job family/job memeriksa referensi inti; FK PostgreSQL tetap menolak referensi lain yang belum diprecheck.
- Tes unit meliputi kode duplikat, parent tidak aktif, dan delete/deaktivasi saat masih direferensikan. `go fmt`, `go vet`, `go test -race -count=1 ./...`, serta build dan health Docker lulus.

## Batas

- Belum ada E2E positif dengan grant IAM sah dan database test, sehingga klaim CRUD siap pakai belum terbukti.
- Scope legal entity/unit/self, validator untuk 76 resource mutabel lain, relay outbox, dan wiring Gateway HRM belum selesai. Goal gRPC HRM keseluruhan tetap aktif.
