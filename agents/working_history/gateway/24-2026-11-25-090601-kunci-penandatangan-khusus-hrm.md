# Kunci penandatangan khusus HRM

Gateway dapat membaca kunci HMAC HRM dari file privat yang dipasang pada kedua replika. File eksplisit yang hilang menggagalkan inisialisasi HRM. Tes TCP mTLS membuktikan verifikasi caller berhasil dengan kunci file ketika kunci umum berbeda. Race test, vet, dan validasi Compose lulus.
