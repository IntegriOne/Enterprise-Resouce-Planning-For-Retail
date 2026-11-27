# Validasi masa berlaku rule set payroll

Kalkulasi memeriksa yurisdiksi dan cakupan seluruh periode sebelum memanggil adapter. Regresi gRPC membuktikan rule set yang berakhir sebelum periode selesai ditolak tanpa panggilan adapter. Tes PostgreSQL dengan race detector dan go vet lulus. [Laporan gRPC](../../e2e/human_resource/35-payroll-mutation-calculate-payroll-run.md)
