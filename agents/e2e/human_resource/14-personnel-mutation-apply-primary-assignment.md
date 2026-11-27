# Penerapan penempatan primer setelah periode posisi berubah

Tes menggunakan PostgreSQL sementara, Redis TLS, race detector, dan verifier Foundation tiruan. Tanggal akhir posisi dan penempatan disiapkan langsung di fixture.

| Kasus | Hasil |
| --- | --- |
| Terapkan ID yang bukan penempatan | FailedPrecondition |
| Terapkan penempatan terbuka ketika transaksi perubahan tanggal akhir posisi masih berjalan | RPC menunggu commit, lalu mengembalikan FailedPrecondition. Referensi primer tetap kosong |
| Terapkan setelah tanggal masuk manajer dikoreksi menjadi setelah awal penempatan | FailedPrecondition |
| Akhiri penempatan pada tanggal akhir posisi, lalu terapkan dua kali secara paralel | Satu berhasil dan satu FailedPrecondition |
| Baca pekerja dan outbox setelah percobaan | Penempatan yang benar menjadi primer dan mutasi percobaan tidak menduplikasi outbox |
| Ajukan, setujui, dan terapkan penempatan pengganti setelah periode lama disiapkan sebagai berakhir | Referensi primer berpindah ke penempatan baru |
| Ulangi penerapan pengganti dan baca riwayat | FailedPrecondition, tepat satu event transfer, dan satu tambahan outbox pekerja |
