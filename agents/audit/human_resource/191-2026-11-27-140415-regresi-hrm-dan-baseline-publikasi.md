# Regresi HRM dan baseline publikasi

| Pemeriksaan | Hasil |
| --- | --- |
| HRM `go test -race -p 1 -count=1 -tags=integration ./...` | 69 paket menjalankan tes. PostgreSQL memakai database baru berakhiran `_test` yang dibersihkan setelah suite. Exit code 0. |
| HRM `go vet ./...` | Lulus. |
| IAM, Enterprise Foundation, Gateway `go test -race -count=1 ./...` | 24, 21, dan 9 paket menjalankan tes. Masing-masing exit code 0. |
| Inventaris kontrak privat | 1.014 RPC pada 22 service, nol selisih. |
| Gateway development lewat HTTPS | Bentuk SDL 24 Node dan 24 Connection. Guard lulus untuk 52 Query dan 55 Mutation. |
| Baseline `origin/main` | Commit `e28491c` memuat modul serta empat RPC referensi Enterprise Foundation. SDL domain HRM masih di worktree lokal. |

Hasil tes berasal dari worktree lokal. Jalur yang memerlukan Finance atau DRM masih membutuhkan adapter pemilik dan uji lintas layanan. [Status publikasi](../../grpc/hrm/32-handoff-integrasi-main-gateway.md) · [Matriks Query](../../grpc/hrm/31-bukti-https-query-gateway.md) · [Matriks Mutation](../../grpc/hrm/30-bukti-https-mutation-gateway.md).
