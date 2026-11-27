# Validasi lokal sebelum konfirmasi tugas offboarding

`CompleteOffboardingTask` memeriksa integritas transisi sebelum menghubungi verifier. Regresi gRPC membuktikan tugas pada rencana draf ditolak tanpa memanggil verifier dan jalur sukses tetap berjalan. Tes PostgreSQL sementara dengan race detector dan `go vet ./...` lulus. [Laporan gRPC](../../e2e/human_resource/40-offboarding-mutation-complete-offboarding-task.md)
