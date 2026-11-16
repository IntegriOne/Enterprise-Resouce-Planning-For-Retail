# Audit draf review kinerja mandiri HRM

## Perubahan

Validator produksi `hr_performance_reviews` mengizinkan CRUD draf `self` untuk worker aktif dengan link IAM ke principal reviewer dan siklus terbuka. Application service menetapkan reviewer dari actor server; validator mewajibkan caller terverifikasi yang sama pada semua mutasi. Identitas worker, siklus, reviewer, role, dan versi tidak dapat dipindah lewat update; rating, asesmen kompetensi terisi, development plan, submit, finalisasi, dan digest masih ditolak. Update/delete ditolak bila review sudah direferensikan skor goal atau asesmen talent. Read/list/create/update sensitif dan memerlukan step-up.

## Bukti

- Unit test memeriksa worker/siklus, hubungan reviewer, duplikasi versi, pembatasan role dan assessment, serta referensi child.
- Tes `bufconn` → PostgreSQL Docker `_test` lima pengulangan dengan `-race`: create dengan reviewer dari server, duplikasi, get/finder, update narasi, larangan manager review, delete dan outbox. Database sementara dihapus.
- `go fmt`, `go vet`, `go test -race ./...`, `go build ./...`, serta Docker build/health modul HRM lulus.

## Batas

Validator terpasang pada 25/78 resource mutasi; 53 lainnya tetap fail-closed. Tes integrasi memakai policy tiruan, sehingga grant/MFA IAM, scope parsial, Gateway GraphQL, dan identitas worker dari IAM nyata belum terbukti. Review manager/peer, rating, submit, finalisasi dan keberatan masih perlu workflow serta verifikasi relasi.
