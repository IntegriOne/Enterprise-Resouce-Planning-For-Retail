# Regresi kontrak Relay dan server gRPC HRM

| Cakupan | Bukti pada worktree |
| --- | --- |
| Server HRM | `go test -race -p 1 -count=1 -tags=integration ./...` dengan PostgreSQL disposable berakhiran `_test` berakhir dengan exit code 0. Ada 68 paket yang menjalankan tes. |
| IAM | `go test -race ./...` berakhir dengan exit code 0. Ada 24 paket yang menjalankan tes. |
| Enterprise Foundation | `go test -race ./...` berakhir dengan exit code 0. Ada 21 paket yang menjalankan tes. |
| Gateway | `go test -race ./...` berakhir dengan exit code 0. Ada 9 paket yang menjalankan tes. |
| Inventaris RPC | Counter nama dari kontrak proto dan tabel panduan operasi cocok untuk seluruh 1.013 deklarasi. |
| Dev | Kontainer HRM, IAM, Enterprise Foundation, Gateway dan frontend sehat. Guard HTTPS mencakup 51 Query dan 55 Mutation. [Bukti](../../e2e/gateway/158-hrm-query-mutation-guard-kontrak-relay.md). |

Scope uji lokal mencakup perubahan terkait Relay dan penempatan. Inventaris membuktikan cakupan dokumentasi, sementara uji guard membuktikan penolakan pada batas akses. Jalur sukses HTTPS dibuktikan per skenario.

Startup HRM memasang integrasi EF dan verifier penautan principal. Jalur produksi penerbitan PDF tetap membutuhkan event Finance serta penyimpanan dan akses versi privat DRM. Persetujuan requisition juga memerlukan mandat peninjau yang ditetapkan serta otorisasi anggaran Finance. [Finance](../../grpc/hrm/25-kebutuhan-integrasi-finance-payroll.md) · [DRM](../../grpc/hrm/24-kebutuhan-integrasi-drm.md) · [Requisition](../../grpc/hrm/35-kebutuhan-persetujuan-requisition.md).

Hasil audit mengikuti worktree bersama. Perubahan yang masih lokal belum masuk Git HEAD.
