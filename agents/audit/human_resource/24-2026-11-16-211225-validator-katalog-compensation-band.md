# Audit validator katalog compensation band HRM

## Perubahan

`hr_compensation_bands` kini mempunyai validator produksi yang memeriksa kombinasi `band_code`, `band_version`, dan `jurisdiction_code` dalam transaksi. Referensi approval dari klien ditolak karena command persetujuan belum tersedia. Band yang sudah direferensikan `hr_compensation_records` tidak dapat diubah atau dihapus. Band lama yang sudah memiliki referensi approval juga tidak dapat dimutasi melalui CRUD biasa.

## Bukti

- `GOWORK=off go fmt ./...`, `go vet ./...`, `go test -race ./...`, dan `go build ./...` lulus pada modul HRM.
- Tes gRPC `bufconn` → PostgreSQL Docker database `human_resource_compensation_test` dengan `-race -count=5` lulus: create/get/finder berhalaman/update/delete, duplikasi versi, penolakan approval tanpa workflow, dan dua event outbox setelah create/update. Database sementara telah dihapus.
- `docker compose build human-resource` dan `docker compose up -d --wait --no-deps human-resource` lulus; container healthy.

## Batas

Validator produksi kini 21/78 resource mutasi; 57 lainnya tetap fail-closed. Tes memakai policy tiruan sehingga belum membuktikan grant IAM, mTLS/HMAC runtime, Gateway GraphQL, scope parsial, atau approval kompensasi. Penolakan band terpakai dibuktikan unit test dengan repository tiruan, belum fixture PostgreSQL catatan kompensasi.
