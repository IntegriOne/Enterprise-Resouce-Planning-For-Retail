# Regresi integrasi dan status HRM

| Cakupan | Bukti terkini |
| --- | --- |
| HRM, IAM, Enterprise Foundation, Gateway | `go test -count=1 ./...` dan `go vet ./...` berakhir dengan exit code 0 di masing-masing modul. |
| HRM dengan PostgreSQL | 68 paket lulus dengan race detector, tag integration, database disposable, dan font DejaVu Sans untuk tes PDF. |
| Onboarding | Runner repo membangun empat service, menjalankan aktivasi dan pengajuan dari HRM, memproses penautan melalui IAM, dan mengonfirmasi satu akun serta satu penerimaan. Kontrak bertanda tangan memakai fixture di database disposable. |
| Scope Workforce Planning | Dalam uji HTTPS, grant badan hukum memberi akses ke satu rencana dan rinciannya. Detail di badan hukum lain ditolak dan total daftar terfilter. [Bukti](../../e2e/gateway/99-hrm-query-workforce-plan-scope-badan-hukum.md). |
| Katalog dan schema | IAM development memiliki 63 izin HRM resmi yang cocok dengan sensitivitas gRPC. Ada 44 Query dan 47 Mutation khusus HRM dengan guard sesi atau CSRF teruji. [Katalog](../../e2e/gateway/141-iam-query-katalog-resmi-hrm.md) · [Schema](../../e2e/gateway/84-hrm-query-mutation-guard-seluruh-schema.md). |

Startup HRM memasang verifier penautan principal. Kalkulator payroll, verifier anggaran dan clearance, serta penerbitan dan akses slip masih memerlukan adapter produksi. Finance memiliki penerbitan, transfer, dan rekonsiliasi. DRM memiliki versi PDF privat dan bukti dokumen. [Finance](../../grpc/hrm/25-kebutuhan-integrasi-finance-payroll.md) · [DRM](../../grpc/hrm/24-kebutuhan-integrasi-drm.md).

Hasil ini berlaku pada checkout bersama yang berisi perubahan lokal. Git HEAD belum merepresentasikan seluruh implementasi yang diuji.
