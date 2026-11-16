# Konkurensi draf suksesi

Tes gRPC→PostgreSQL suksesi diperluas dengan dua create bersamaan pada posisi, kode, dan versi yang sama. Sepuluh pengulangan dengan race detector membuktikan satu draf berhasil, satu `AlreadyExists`, dan tidak ada event outbox duplikat. Seluruh fmt/vet/build/race test HRM lulus; database disposable dihapus. Nominasi kandidat tetap fail-closed karena command panel dan pembuktian mandat belum tersedia. Lihat audit HRM nomor 38.
