# Presisi waktu jadwal wawancara

Tes gRPC mereproduksi interval satu nanodetik yang lolos validasi aplikasi lalu menjadi interval kosong di PostgreSQL. Create dan update sekarang menyamakan presisi timestamp hingga mikrodetik sebelum validasi. Interval yang kosong pada presisi penyimpanan mengembalikan InvalidArgument. Integrasi Recruitment juga memeriksa waktu yang dikembalikan saat create sesuai waktu tersimpan.
