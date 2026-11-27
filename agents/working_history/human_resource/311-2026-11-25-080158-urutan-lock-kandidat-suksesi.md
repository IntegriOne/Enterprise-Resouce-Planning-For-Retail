# Urutan lock kandidat suksesi

Persetujuan rencana mengunci kandidat berdasarkan ID pekerja. Dua rencana berbeda dengan pekerja yang sama memperoleh lock pekerja dalam urutan konsisten. Integrasi gRPC Succession dengan PostgreSQL sementara dan race detector serta go vet lulus. Skenario konkurensi dua rencana belum diuji secara khusus.
