# Validasi periode sebelum menyimpan slip

HRM kini memeriksa identitas periode, grup payroll, badan hukum, mata uang, dan yurisdiksi sebelum memanggil penyimpan PDF. PostgreSQL juga membatasi relasi ini melalui foreign key gabungan. Regresi gRPC dengan race detector lulus, termasuk pemrosesan ulang setelah sebagian slip berhasil disimpan. Finance dan DRM masih diperlukan untuk membuktikan workflow produksi.
