# Handler event penerbitan payroll

Handler internal inbox memvalidasi sumber, versi, dan ID event penerbitan sebelum memanggil pemrosesan run. Tes PostgreSQL Payroll dengan race detector membuktikan sumber yang berbeda ditolak dan pengiriman ulang melanjutkan slip yang belum berhasil. Go vet lulus. Transport Finance, pencatatan receipt, serta adapter DRM produksi masih terbuka.
