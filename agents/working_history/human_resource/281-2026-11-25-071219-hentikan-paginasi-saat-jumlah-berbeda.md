# Hentikan paginasi saat jumlah baris melebihi snapshot

Pengumpul data HRM menghentikan pengambilan halaman begitu jumlah baris melampaui hitungan awal dan mempertahankan kode penolakan Forbidden. Batas 10.000 dan pemeriksaan snapshot kurang lengkap tetap ditegakkan. Tes regresi membuktikan tidak ada permintaan halaman lanjutan. Seluruh paket unit service HRM dengan race detector dan go vet lulus.
