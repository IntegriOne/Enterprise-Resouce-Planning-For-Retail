# Audit posting expiry saldo cuti HRM

## Temuan dan perbaikan

`PostLedgerEntry` sebelumnya menolak semua `expiry`, sehingga saldo periode usai tidak dapat ditutup. Validator kini menerima `expiry` dari HRM hanya ketika tanggal efektif tepat akhir periode, waktu posting tidak lebih awal, dan debit persis menghabiskan saldo positif. Entry expiry harus menjadi urutan ledger terakhir; posting berikutnya dan pengajuan koreksi baru ditolak. Revisi, proyeksi, batas saldo, dan outbox tetap berada dalam transaksi entitlement.

Pengujian berulang menemukan asumsi keliru pada validator baru: repository mengurutkan halaman menurut UUID, bukan `entry_sequence`. Pemeriksaan expiry diperbaiki memakai nomor urut ledger; domain replay tetap mengurutkan entry secara eksplisit. Penolakan posting ulang diberi error workflow sebelum perhitungan saldo agar status gRPC stabil.

## Bukti dan batas

- `go fmt ./...`, `go vet ./...`, `go test -race ./...`, `go build ./...` lulus pada modul HRM.
- `go test -race -tags=integration -count=20 -run '^TestLeaveLedgerRPCWithPostgreSQL$' ./transport/grpc/handler/leave` lulus memakai PostgreSQL Docker `_test` dan gRPC `bufconn`.
- Kasus meliputi expiry terlalu dini, debit parsial, expiry benar saldo 11→0 dan revisi 3→4, expiry ulang, posting adjustment serta create adjustment setelah expiry.
- Tes memakai policy tiruan; IAM grant, Gateway GraphQL, mTLS/HMAC runtime, dan scope parsial belum dibuktikan. Usage/accrual/reversal serta sumber luar tetap ditolak.
