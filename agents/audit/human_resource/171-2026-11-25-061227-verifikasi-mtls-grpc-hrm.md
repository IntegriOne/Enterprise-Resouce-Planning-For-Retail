# Verifikasi mTLS gRPC HRM

Pengujian dijalankan pada kontainer HRM lokal yang sedang berjalan. Binary health check berhasil membuka koneksi TLS 1.3 dengan sertifikat klien dan menerima status SERVING. Koneksi TLS 1.3 dengan CA tepercaya dan tanpa sertifikat klien menerima alert `certificate required`.

Hasil mencakup transport pada image aktif. Pengujian ini tidak menjalankan resolver Gateway atau provisioning sesi dan grant IAM.

Panggilan `PersonnelService/GetAllWorkers` melalui koneksi mTLS dengan sertifikat klien lokal sah dan tanpa metadata bukti Gateway ditolak dengan status gRPC `Unauthenticated` dan pesan `valid session required`.
