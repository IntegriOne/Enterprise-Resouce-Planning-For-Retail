# Audit penolakan requisition HRM

## Temuan dan perbaikan

Requisition `submitted` sebelumnya hanya mempunyai jalur approval. Pengajuan yang ditolak tidak bisa ditutup tanpa mengubah status lewat jalur tidak resmi. `RejectRequisition` menambah command/RPC sensitif dengan alasan wajib, actor manusia berbeda dari pembuat dan pengusul, serta waktu server. Command mengunci requisition, memeriksa izin dan snapshot sebelum commit, lalu menulis keputusan bersama outbox dalam satu transaksi. Requisition `rejected` menjadi terminal. Finder alasan dan actor memakai filter serta paginasi server.

DBML dan migrasi aditif `000006` menambah `rejected_at`, `rejection_reason_code`, `rejected_by_principal_id`. Constraint memakai `IS NOT NULL` eksplisit, memaksa urutan waktu dan pemisahan actor, serta melarang metadata penolakan pada status lain. Baseline dan migrasi lama tetap immutable. Katalog permission aktif bertambah menjadi 338 tanpa menulis grant IAM.

## Bukti

- `buf lint` dan `buf generate`: lulus. DBML diterjemahkan oleh `@dbml/cli@12.0.0`; 86 tabel dan 9 enum.
- `go fmt ./...`, `go vet ./...`, `go test -race ./...`, `go build ./...`: lulus pada modul HRM.
- Tes `bufconn`→PostgreSQL `_test` dengan race: lulus untuk self rejection, alasan kosong, keputusan sah, pencarian, retry, approval setelah reject, outbox, dan constraint SQL actor/alasan.
- Tes migrasi pada schema terisolasi: versi 6 turun/naik, versi asing dan checksum salah ditolak.
- Docker HRM dibangun ulang dan healthy; database development menampilkan migrasi 1–6 serta dua constraint baru.

## Batasan

Policy tes memakai actor tiruan. Grant IAM dan E2E Gateway belum tersedia. Verifier Foundation/Finance untuk approval/open requisition produksi masih fail-closed. Penolakan ini hanya keputusan HRM dan tidak membuat grant atau memodifikasi data management lain.
