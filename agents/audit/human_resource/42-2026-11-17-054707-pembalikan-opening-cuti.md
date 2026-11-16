# Audit pembalikan opening cuti HRM

Tanggal: 17 November 2026 (WIB). Scope: hanya `backend/services/human_resource` dan katalog izin HRM.

## Temuan dan perbaikan

`PostLedgerEntry` sudah menerima `reverses_entry_id` dan aturan domain dapat menghitung pembalikan eksak, tetapi validator produksi menolak semua reversal. Pembalikan `opening` dari sumber HRM kini diterima bila pelaku adalah manusia yang berbeda dari pelaku posting awal, referensi entry ada pada entitlement yang sama, tanggal pembalikan tidak mendahului entry awal, jumlah tepat berlawanan, dan belum ada pembalikan lain. Reversal setelah expiry tetap ditolak. Lock entitlement, deduplikasi sumber, proyeksi saldo, dan outbox memakai transaksi command yang sama. Pembacaan ulang validator entitlement menjaga aturan pembalikan pada setiap mutasi saldo berikutnya; lookup berdasarkan ID dan `entry_sequence` mencegah hasil bergantung pada urutan halaman repository.

## Bukti

- `go test -race -count=1 ./...`, `go vet ./...`, dan `go build ./...` pada modul HRM lulus.
- Tes `TestLeaveLedgerRPCWithPostgreSQL` dengan database disposable `_test` diulang 10 kali dengan race detector; membuktikan pembalikan valid, penolakan pelaku yang sama, jumlah parsial, referensi lintas entitlement, sumber duplikat, pembalikan kedua, dua entry ledger, dan tiga event outbox.
- `docker compose build --quiet human-resource` dan `docker compose up -d --wait --no-deps human-resource` lulus; container healthy.

## Batas

Policy tes tiruan; grant IAM, Gateway GraphQL, dan E2E browser belum dibuktikan. Jenis `accrual`, `usage`, reversal `adjustment`/`expiry`, dan sumber management lain tetap fail-closed. Tidak ada permission baru: operasi memakai izin sensitif `human_resource.leave_entitlements.post` yang sudah ada.
