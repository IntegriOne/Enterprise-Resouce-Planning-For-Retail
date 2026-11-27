# Ambang anonimitas hasil survei

Penyimpanan hasil sebelumnya hanya mensyaratkan lima responden, sedangkan setiap survei dapat menetapkan ambang lebih tinggi. Hasil di bawah ambang survei kini disaring di query PostgreSQL sebelum pembacaan berdasarkan ID, daftar, finder, pagination dan total. Pemeriksaan mutasi survei tetap menghitung semua hasil tersimpan, termasuk hasil yang disembunyikan, sehingga survei yang sudah memiliki hasil tetap terlindungi dari perubahan atau penghapusan.

Integrasi gRPC ke PostgreSQL sementara membuktikan bahwa hasil dari enam responden terlihat dengan ambang enam. Hasil dari lima responden kemudian menghasilkan NotFound, halaman kosong dan total nol. Survei tetap terlindungi dari perubahan dan penghapusan. Publikasi, penutupan dan pemasukan agregat survei masih memerlukan workflow terpisah.
