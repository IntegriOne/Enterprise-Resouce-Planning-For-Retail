# Sinkronisasi fixture pengajuan akun dengan aktivasi pekerja

Regresi PostgreSQL menemukan fixture pengajuan akun yang belum merekam status hasil peristiwa perekrutan. Fixture kini memakai transisi `preboarding` ke `active` dan kunci peristiwa yang sama dengan command aktivasi. Tes mengirim pengajuan tanpa ID peristiwa dan memeriksa bahwa HRM memilih peristiwa yang benar. Regresi terarah dengan race detector lulus. Dokumentasi program kini mengarah ke handoff onboarding, kontrak DRM yang dibutuhkan, dan README HRM aktif.
