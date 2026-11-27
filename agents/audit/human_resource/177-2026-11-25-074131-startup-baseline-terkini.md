# Startup HRM pada baseline terkini

Binary `./cmd` dari checkout terbaru lulus smoke test. Kontainer sementara memakai database PostgreSQL terpisah, konfigurasi Redis TLS lokal, dan listener gRPC yang diisolasi dari port server aktif. Health check menerima `SERVING` melalui mTLS. SHA-256 baseline versi 1 yang dicatat database sama dengan baseline checkout. Kontainer dan database sementara dibersihkan setelah pengujian.

Bukti mencakup startup dan health check binary terbaru. Operasi bisnis lintas domain dibuktikan pada regresi handler terpisah. Integrasi produksi IAM, Finance, DRM dan Gateway tetap bergantung pada kesiapan kontrak masing-masing.
