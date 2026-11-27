# RPC persetujuan kompensasi independen

Menambahkan ApproveCompensationRecord dengan izin sensitif. Riwayat created dan updated pada outbox menjadi bukti pihak yang membuat atau mengedit proposal. Actor tersebut ditolak saat menyetujui. Tes gRPC PostgreSQL mencakup pemisahan pihak, pencegahan approval CRUD, transisi dan outbox atomik. Inventaris Gateway cocok dengan 1.011 RPC.
