# Audit approval dan posting koreksi cuti

## Perubahan

Command `ApproveLeaveAdjustment` ditambahkan ke application service dan gRPC privat. Kode `human_resource.leave_adjustments.approve` berbeda dari CRUD dan posting ledger, memakai step-up, mengunci adjustment dan entitlement, menolak self approval, serta mencatat actor/waktu server melalui transaksi dan outbox. Validator baru membatasi koreksi draft, approval, idempotency, relasi entitlement, dan pembekuan sesudah disetujui atau diposting. Validator ledger mengizinkan `adjustment` HRM hanya setelah approval independen; proyeksi entitlement dibangun ulang dari seluruh ledger dan dibatasi `max_balance` policy. `opening` sebelumnya tetap tersedia; usage/accrual/reversal dan sumber luar tetap ditolak.

## Bukti

- `go generate ./contract` menjalankan `buf lint` dan `buf generate`; diff generated hanya pada kontrak Leave.
- `go fmt ./...`, `go vet ./...`, `go test -race ./...`, dan `go build ./...` lulus pada modul HRM.
- Tes `-race -tags=integration` memakai gRPC `bufconn` → service → PostgreSQL Docker `_test`: create policy/entitlement/adjustment, self approval ditolak, approval independen berhasil, posting koreksi mengubah revisi 2→3 dan saldo 12→11, koreksi melewati `max_balance` ditolak tanpa ledger tambahan.
- Docker `human-resource` dibangun ulang dan healthy.

## Batas

Tes memakai policy tiruan; metadata aktor khusus tes bukan mekanisme autentikasi runtime. Belum ada grant IAM aktif, uji mTLS/HMAC nyata, scope parsial, atau Gateway GraphQL HRM. Approval tidak otomatis memposting saldo. Validator produksi mencakup 18/78 resource mutasi; 60 lainnya fail-closed.
