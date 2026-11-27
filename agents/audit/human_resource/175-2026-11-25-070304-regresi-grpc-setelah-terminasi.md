# Regresi gRPC HRM setelah terminasi

Regresi checkout terkini menjalankan 24 paket tes handler yang mencakup 22 domain dan paket pemetaan respons. Sebanyak 46 tes tingkat atas lulus tanpa skip atau gagal. Setiap paket memakai database PostgreSQL sementara tersendiri dan race detector. Tes offboarding mencakup terminasi, penutupan kontrak serta penempatan, pembatalan jadwal, dan penolakan transfer tertunda.

`go vet ./...` dan `go test -count=1 ./...` lulus pada modul HRM. Pencocokan nama RPC dari proto ke tabel handoff menemukan 1.001 operasi pada 22 service, terdiri dari 667 Query dan 334 Mutation. Seluruh nama cocok tanpa duplikat.

Hasil ini mencakup skenario integrasi lokal. Wiring Gateway, grant IAM produksi, kalkulator payroll produksi, serta integrasi produksi Finance dan DRM masih terbuka. [Inventaris operasi](../../grpc/hrm/1-inventaris-operasi-grpc-hrm.md)
