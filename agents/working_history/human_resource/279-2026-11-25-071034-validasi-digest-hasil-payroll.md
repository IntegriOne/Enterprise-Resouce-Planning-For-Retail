# Validasi digest hasil payroll

Review, persetujuan, dan pengajuan posting memakai validasi bersama yang kini menolak hasil pekerja jika panjang digest kalkulasinya bukan 32 byte. Tes service membuktikan penolakan approval. Regresi gRPC PostgreSQL dengan race detector dan go vet lulus.
