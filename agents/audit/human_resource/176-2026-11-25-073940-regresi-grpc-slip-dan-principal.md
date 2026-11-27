# Regresi gRPC HRM setelah slip pribadi dan pengaitan principal

Seluruh 24 paket handler bertes dijalankan dengan tag integration, race detector, dan database PostgreSQL sementara per paket. Seluruh 46 tes tingkat atas lulus tanpa skip, mencakup 22 domain serta pemetaan respons. Tes unit modul dengan race detector lulus pada 49 paket yang memiliki tes, termasuk renderer slip dengan DejaVu Sans dan pdftotext. Go vet dan verifikasi skema PostgreSQL up–integritas–down–up ulang lulus.

Pencocokan nama RPC dengan handoff Gateway lulus untuk 1.004 operasi pada 22 service. Tabel terdiri dari 669 Query dan 335 Mutation. Handoff mencatat verifier pengaitan akun dan slip pribadi pada bagian kesiapan runtime. [Inventaris operasi](../../grpc/hrm/1-inventaris-operasi-grpc-hrm.md)

Pengujian lokal memakai fixture untuk kontrak owner yang masih terbuka. Startup produksi belum memasang verifier pengaitan akun, penyimpanan serta pembacaan slip DRM, transport Finance dan kalkulator payroll. Uji ini belum membuktikan akses dari Gateway dengan grant IAM produksi.
