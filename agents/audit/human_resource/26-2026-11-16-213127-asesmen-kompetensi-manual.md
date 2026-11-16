# Audit asesmen kompetensi manual HRM

## Perubahan

Validator produksi `hr_worker_competency_assessments` mengizinkan create observasi manual pada worker dan kompetensi aktif. Penilai diisi oleh application service dari actor terautentikasi. Waktu asesmen, kedaluwarsa, dan referensi lokal diperiksa dalam transaksi. Bukti dokumen ditolak sampai DRM menyediakan verifikasi. Update dan delete hasil ditolak agar riwayat asesmen tidak tertimpa. Read/create/update ditandai sensitif dan memerlukan step-up pada policy produksi.

## Bukti

- Unit test memeriksa worker/kompetensi tidak aktif, metode atau bukti belum didukung, asesmen di masa depan, kedaluwarsa sebelum asesmen, referensi hilang, cancellation, serta larangan update/delete.
- Tes `bufconn` → PostgreSQL Docker `_test` lima pengulangan `-race`: create, actor server, duplikasi idempotency key, get/finder berhalaman, larangan update/delete, outbox tunggal. Database sementara telah dihapus.
- `go fmt`, `go vet`, `go test -race ./...`, `go build ./...`, dan Docker build/health lulus pada modul HRM.

## Batas

Validator kini terpasang pada 24/78 resource mutasi; asesmen hanya mendukung create, 54 resource lainnya tetap fail-closed. Tes memakai policy tiruan, sehingga grant/MFA IAM, scope parsial, Gateway GraphQL, dan bukti DRM nyata belum teruji.
