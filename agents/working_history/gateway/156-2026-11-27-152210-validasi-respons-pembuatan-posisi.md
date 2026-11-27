# Pencocokan referensi respons pembuatan posisi

Gateway memeriksa lokasi kerja dan posisi atasan yang dikembalikan HRM terhadap permintaan createPosition. Regresi dengan fixture gRPC membuktikan dua respons tidak konsisten ditolak. Suite Gateway dengan race detector dan go vet lulus.
