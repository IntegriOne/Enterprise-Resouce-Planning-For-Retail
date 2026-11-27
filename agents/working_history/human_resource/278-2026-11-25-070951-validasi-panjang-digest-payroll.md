# Validasi panjang digest payroll

Review, persetujuan, dan pengajuan posting memeriksa panjang digest input dan output SHA-256. Tes service membuktikan persetujuan menolak panjang digest keluaran yang salah. Tes service, regresi gRPC PostgreSQL dengan race detector, dan go vet lulus.
