# Audit penerimaan lamaran rekrutmen HRM

## Temuan dan perbaikan

`CreateApplication` selalu gagal sebelum menyimpan data: `Initialize` mengisi `applied_at` dari server, tetapi guard create menolak timestamp yang tidak kosong. Guard itu dihapus karena `CreateApplicationInput` dan proto tidak menerima `applied_at`; timestamp tetap diisi server.

Validator `hr_applications` sekarang hanya mengizinkan create `received`. Ia mengunci kandidat dan requisition, memastikan kandidat aktif dengan persetujuan notice sebelum aplikasi dan retensi belum berakhir, serta requisition `open` yang sudah disetujui dan dibuka sebelum aplikasi. Field recruiter, screening, tahap, keputusan, hiring dan sumber posting ditolak. Update/delete masih memerlukan command workflow. Keunikan kandidat–requisition, nomor dan idempotency key dijaga constraint database.

## Bukti dan batas

- Tes unit/race validator lulus. Tes gRPC `bufconn`→PostgreSQL disposable `*_test` lulus untuk create kandidat dan aplikasi, duplikasi, pagination, penolakan transisi status, delete aplikasi/kandidat terhubung, serta outbox.
- `go fmt ./...`, `go vet ./...`, `go test -race -count=1 ./...`, dan `go build ./...` lulus pada modul HRM.
- `docker compose build human-resource` lulus; `docker compose up -d --wait human-resource` mengembalikan status `healthy`.
- Requisition `open` disiapkan sebagai fixture SQL; RPC persetujuan/pembukaannya belum ada. Policy tes tiruan, sehingga grant IAM, scope nyata, Gateway dan transport mTLS/HMAC belum dibuktikan.
- Validator produksi mencakup 31/78 resource mutasi; 47 tetap fail-closed.
