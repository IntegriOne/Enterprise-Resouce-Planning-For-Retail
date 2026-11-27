# Ikatan bukti DRM ke asesmen dan lamaran

CompleteApplicationAssessment mensyaratkan verifier DRM mengembalikan ID asesmen dan lamaran yang cocok. Versi, digest, pemilik dan tujuan tetap diperiksa. Tes service menolak bukti dari asesmen lain dan lamaran lain. Regresi gRPC Recruitment memakai verifier tiruan, PostgreSQL sementara, race detector, dan go vet. Adapter DRM produksi masih terbuka.
