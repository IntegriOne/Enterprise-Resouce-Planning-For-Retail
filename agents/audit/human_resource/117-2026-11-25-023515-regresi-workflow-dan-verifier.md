# Regresi workflow dan verifier HRM

Suite `GOWORK=off go test -p 1 -tags=integration -race -count=1 ./...` lulus pada 67 paket dengan PostgreSQL sementara yang dibuat khusus untuk tes dan dihapus setelahnya. Run ini mencakup perubahan klasifikasi verifier lintas domain, pembatasan komentar tiket dengan lampiran yang belum terverifikasi, dan pemeriksaan waktu persetujuan payroll terminasi saat finalisasi offboarding.

Hasil ini membuktikan regresi pada modul HRM. Jalur yang membutuhkan verifier produksi DRM/Finance dan pengujian dengan sesi serta grant IAM nyata masih terbuka.
