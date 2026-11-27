# Persetujuan koreksi cuti serentak

`TestLeaveLedgerRPCWithPostgreSQL` lulus tiga kali dengan race detector dan database PostgreSQL sementara yang baru pada setiap eksekusi.

| Kasus | Hasil |
| --- | --- |
| Dua RPC serentak menyetujui koreksi masing-masing −6 hari dari saldo 11 hari | Satu sukses dan satu `FailedPrecondition` |
| Dua RPC serentak menyetujui koreksi masing-masing +0,75 hari dari saldo 11 hari dengan batas 12 hari | Satu sukses dan satu `FailedPrecondition` |
| Hitung koreksi yang benar-benar tersimpan sebagai disetujui | Tepat satu record |

Policy akses pada tes memakai fixture. Cakupan ini memeriksa serialisasi perubahan terhadap saldo pada PostgreSQL.

Setelah kasus penambahan saldo ditambahkan, kedua subtes konkurensi kembali lulus dengan PostgreSQL dan race detector.
