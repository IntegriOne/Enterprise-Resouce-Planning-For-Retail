# Persetujuan rencana pengembangan kosong

Tes gRPC mereproduksi persetujuan rencana tanpa tindakan. Rencana yang sudah aktif tidak dapat menambah tindakan melalui CRUD dan tidak bisa diselesaikan bila tindakan kosong. Validator kini mewajibkan minimal satu tindakan pada transisi draft ke active saat rencana terkunci. Tes membuktikan persetujuan ditolak saat kosong dan alur persetujuan hingga penyelesaian tetap berhasil setelah tindakan ditambahkan. Integrasi Career dengan PostgreSQL sementara, race detector, tes validator keamanan, dan vet modul lulus.
