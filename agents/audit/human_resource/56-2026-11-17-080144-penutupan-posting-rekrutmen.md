# Audit penutupan posting rekrutmen HRM

## Temuan dan perubahan

Posting `published` dapat ditutup melalui command `CloseJobPosting` dan RPC privat dengan izin sensitif tersendiri. Service mewajibkan actor manusia dan waktu tutup server setelah publikasi. Validator mengunci requisition parent yang masih `open`, hanya menerima perubahan status dan `closes_at`, serta membekukan posting yang sudah `closed`. Command menyimpan posting dan outbox dalam satu transaksi. Penutupan posting memungkinkan `CloseRequisition` berjalan tanpa mengubah fixture posting langsung melalui SQL.

## Bukti pengujian

- `buf lint` dan `buf generate`: lulus.
- `go fmt ./...`, `go vet ./...`, `go test -race ./...`, `go build ./...`: lulus pada modul HRM.
- gRPC `bufconn` → PostgreSQL disposable `_test` dengan `-race`: lulus untuk publish, close posting, retry, outbox, larangan delete posting closed, dan close requisition.
- `docker compose build human-resource` dan `docker compose up -d --wait human-resource`: lulus; container healthy.

## Batasan

Status `closed` menyatakan keputusan HRM menghentikan distribusi. Belum ada acknowledgement dari kanal eksternal atau E2E IAM–Gateway–DRM. Publish posting tetap gagal tertutup pada runtime produksi sampai verifier DRM tersedia. Penarikan posting yang memerlukan alasan dan audit khusus belum diimplementasikan.
