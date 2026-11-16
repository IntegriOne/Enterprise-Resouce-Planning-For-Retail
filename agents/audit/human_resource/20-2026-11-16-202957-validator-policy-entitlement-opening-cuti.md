# Audit validator policy, entitlement, dan opening cuti

## Temuan dan perbaikan

`PostLedgerEntry` sudah memakai lock entitlement, domain `RebuildBalance`, transaksi, dan outbox, tetapi runtime Integrity belum memiliki validator untuk entitlement/ledger. Seluruh mutasi ini sebelumnya fail-closed. Validator produksi sekarang mengunci worker, policy, dan entitlement; memeriksa tipe cuti, periode, status worker/policy, masa kerja minimum, saldo awal nol, revisi, serta satu entry `opening` HRM. Jumlah opening dibatasi hak tahunan dan `max_balance` bila ada. Proyeksi setelah insert harus identik dengan hasil rekonstruksi ledger. Policy `annual_grant` diperiksa versi uniknya, create/update memerlukan step-up, dan policy tidak dapat diubah atau dihapus setelah dipakai entitlement.

## Bukti

- `go fmt ./...`, `go vet ./...`, `go test -race ./...`, dan `go build ./...` lulus pada modul HRM.
- gRPC `bufconn` → service → PostgreSQL Docker `_test`: create policy/entitlement, finder berhalaman, posting opening, revisi saldo, dan outbox lulus. Saldo langsung, tipe cuti salah, grant melebihi policy, policy terpakai, sumber duplikat, dan usage tanpa integrasi ditolak. Database sementara dihapus setelah tes.
- Docker `human-resource` dibangun ulang dan berstatus healthy.

## Batas

Policy tiruan pada tes bukan grant IAM nyata. Sumber Workforce, akrual, usage, adjustment, reversal, verifikasi yurisdiksi, serta approval lintas owner belum aktif. Postgres test tidak membuktikan mTLS/HMAC runtime atau Gateway GraphQL. Validator CRUD produksi kini 17/78; 61 resource masih fail-closed.
