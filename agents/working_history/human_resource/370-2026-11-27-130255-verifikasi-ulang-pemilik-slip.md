# Verifikasi ulang pemilik slip

HRM memeriksa kembali sesi dan tautan principal ke pekerja sebelum mengembalikan detail atau PDF. Daftar juga memeriksa bahwa himpunan pekerja yang boleh diakses tetap sama sebelum mengembalikan halaman dan total. Tes gRPC dengan PostgreSQL terisolasi membuktikan bahwa konten PDF ditahan ketika tautan dilepas saat pembaca dokumen berjalan. Go vet lulus.
