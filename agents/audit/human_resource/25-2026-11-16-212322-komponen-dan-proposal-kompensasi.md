# Audit komponen dan proposal kompensasi HRM

## Perubahan

Validator `hr_pay_components` mengizinkan katalog earning fixed dengan mata uang eksplisit dan tanpa formula, rule set, atau klaim pajak/pensiun. Kombinasi kode/versi/yurisdiksi unik. Komponen dibekukan saat sudah dipakai catatan kompensasi, adjustment, atau result line payroll.

Validator `hr_compensation_records` mengizinkan proposal base salary dan allowance. Referensi worker, assignment, position, job, component, dan band dikunci dalam transaksi. Periode, mata uang, kode komponen, level job, dan nominal band diperiksa; update identitas relasi dan approval langsung ditolak. Referensi hilang dikembalikan sebagai konflik prasyarat, tanpa membedakan ID yang tidak ada pada respons gRPC.

## Bukti

- Unit test menolak komponen berformula, versi duplikat, perubahan katalog terpakai, worker tidak aktif, assignment tidak cocok, nominal di luar band, dan approval lewat CRUD.
- Tes `bufconn` melintasi RPC Payroll dan Compensation sampai PostgreSQL Docker `_test` lima pengulangan dengan `-race`: create komponen/band/proposal, finder offset, update/delete proposal, pembekuan referensi, serta event outbox. Database sementara telah dihapus.
- `go fmt`, `go vet`, `go test -race ./...`, `go build ./...`, dan Docker build/health lulus pada modul HRM.

## Batas

Validator produksi 23/78 resource mutasi; 55 lainnya fail-closed. Tes memakai policy tiruan dan fixture HRM langsung, bukan grant IAM atau Gateway GraphQL. Proposal belum mengubah kompensasi efektif; approval, perhitungan payroll, aturan pajak/pensiun dan bukti Finance tetap belum tersedia.
