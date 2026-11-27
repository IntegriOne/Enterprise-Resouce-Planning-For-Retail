# Katalog resmi slip gaji mandiri

Manifest HRM memuat izin payslips.read_self tanpa assignment otomatis. Unit test katalog, sinkronisasi IAM, shared authz, dan adapter Gateway lulus. Regresi HTTPS memakai manifest resmi dan membuktikan grant global ditolak. Grant SELF yang sah mencapai HRM, yang tetap membutuhkan verifier Finance dan DRM untuk membaca slip.
