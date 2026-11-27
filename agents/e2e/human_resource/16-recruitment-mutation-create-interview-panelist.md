# Penugasan panelis pada slot wawancara

Tes memakai PostgreSQL sementara, policy tiruan, dan race detector.

| Kasus | Hasil |
| --- | --- |
| Tugaskan panelis yang masa kerjanya berakhir sebelum hari wawancara | FailedPrecondition |
| Setel tenggat feedback saat wawancara masih berlangsung | InvalidArgument |
| Tugaskan panelis aktif ke slot yang tersedia pada hari terakhir kerja dengan tenggat feedback tepat saat wawancara berakhir | Panelis tersimpan |
| Tugaskan panelis yang sama pada wawancara dengan interval bertumpang tindih | FailedPrecondition |
| Tugaskan panelis yang sama pada wawancara yang mulai tepat ketika wawancara pertama selesai | Panelis kedua tersimpan |
| Kirim dua penugasan serentak untuk panelis dan slot yang sama pada lamaran berbeda | Satu berhasil, satu FailedPrecondition, dan tepat satu reservasi tersimpan |

Jadwal kedua menggunakan kandidat dan lamaran berbeda pada database tes. Kasus ini membuktikan benturan berasal dari jadwal panelis lintas lamaran.
