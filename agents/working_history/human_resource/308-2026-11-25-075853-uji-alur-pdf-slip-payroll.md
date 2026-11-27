# Uji alur PDF slip payroll

Fixture pengujian gRPC Payroll menjalankan renderer produksi bila `HRM_PAYSLIP_TEST_FONT` tersedia. Stub penyimpanan privat memeriksa PDF, digest, ID pekerja, hasil, penerbitan dan badan hukum, serta kunci idempotensi. Pengujian memakai DejaVu Sans dan PostgreSQL sementara dengan race detector. Pengujian ekstraksi teks Unicode serta kestabilan byte PDF juga lulus. Finance dan DRM masih menggunakan fixture.
