# Pemilihan hiring manager berdasarkan pekerja

CreateRequisition mendukung referensi pekerja tanpa meminta principal IAM pada caller baru. HRM mengunci pekerja aktif dan memakai tautan akun yang tersimpan dalam transaksi. Validasi gRPC dan PostgreSQL membuktikan resolusi serta penolakan dua referensi; kontrak lama ikut lulus regresi.
