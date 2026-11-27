# Integrasi scope rencana workforce

Regresi gRPC ke PostgreSQL menemukan bahwa Get rencana dengan grant parsial membutuhkan KeyFilters yang belum terpasang. Spesifikasi rencana kini menggabungkan filter primary key dan badan hukum pada query yang sama. Rencana di luar scope menghasilkan NotFound. List dan finder dengan kode yang sama di dua badan hukum hanya mengembalikan baris dan total yang diizinkan, termasuk saat offset melampaui halaman.

Integrasi memakai policy produksi, caller dari interceptor tes dan snapshot IAM tiruan. Redis dan mTLS runtime belum tercakup. Tes paket Workforce Planning dengan race detector lulus.
