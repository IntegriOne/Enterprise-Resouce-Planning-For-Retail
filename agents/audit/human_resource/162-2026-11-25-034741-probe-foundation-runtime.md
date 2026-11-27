# Verifikasi koneksi adapter HRM ke Foundation lokal

Adapter HRM dari checkout lokal terhubung dengan Enterprise Foundation pada 127.0.0.1:9002 menggunakan kredensial integrasi development yang sudah dikonfigurasi. Referensi badan hukum yang tidak ada menghasilkan konflik referensi. Query daftar badan hukum aktif berhasil dengan hasil kosong. Pemeriksaan positif tidak dapat dijalankan karena tidak ada badan hukum aktif pada runtime ini.

Pengujian hanya membaca. Tidak ada data fixture yang ditambahkan pada database bersama. Alur Personnel lengkap masih memerlukan fixture badan hukum, unit, posisi dan pekerja yang saling terkait, serta aktor IAM yang sah.
