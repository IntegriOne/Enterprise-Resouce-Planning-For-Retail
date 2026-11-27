# Ikat event jurnal payroll ke snapshot

Event pengajuan jurnal menyimpan digest SHA-256 hasil payroll, periode, grup, badan hukum, mata uang, dan arah settlement. Permintaan dengan digest bukan SHA-256 ditolak. Integrasi gRPC–PostgreSQL memeriksa metadata pada payload tersimpan. Jalur Finance belum mengonsumsi event ini.
