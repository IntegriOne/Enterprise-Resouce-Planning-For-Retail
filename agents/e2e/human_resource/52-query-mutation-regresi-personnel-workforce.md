# Regresi gRPC personnel dan workforce planning

Tes dijalankan dari checkout lokal dengan race detector dan database PostgreSQL sekali pakai per paket.

| Tes | Hasil |
| --- | --- |
| `TestWorkerScopedQueryRPCWithPostgreSQL` | Lulus, mencakup baca worker, aktivasi, persetujuan assignment, pengajuan transfer dan pengulangan pengajuan |
| `TestPlannedPositionRPCWithPostgreSQL` | Lulus, mencakup aturan posisi terencana dan aktivasi |
| `TestWorkforcePlanAndLineRPCWithPostgreSQL` | Lulus, mencakup rencana, rincian, workflow dan pembatasan scope |
| `go test -race ./...` dan `go vet ./...` modul HRM | Lulus |

Tes gRPC memakai fixture policy dan verifier. Uji HTTPS dengan grant IAM resmi tetap memerlukan akun uji dan fixture lintas badan hukum.
