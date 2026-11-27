# Nominasi kandidat suksesi melalui gRPC

Tes memakai PostgreSQL sementara, race detector, dan policy tiruan. Masa berlaku posisi disiapkan langsung pada fixture.

| Kasus | Hasil |
| --- | --- |
| Posisi berlaku pada tanggal asesmen dan berakhir sebelum review rencana | FailedPrecondition |
| Hubungan kerja kandidat berakhir pada tanggal asesmen, sebelum review | FailedPrecondition |
| Posisi dan hubungan kerja kandidat berakhir tepat pada tanggal review | Nominasi draf tersimpan |
| Dua nominasi untuk kandidat dan rencana yang sama diajukan serentak | Satu berhasil, satu AlreadyExists, dan hanya satu nominasi tersimpan |
| Status diubah menjadi ready_now tanpa asesmen final | FailedPrecondition |
| Alasan nominasi draf yang memenuhi syarat diubah | Berhasil |
| Review rencana diubah saat kandidat sudah dinominasikan | FailedPrecondition |
| Kandidat berpindah badan hukum | Perubahan nominasi ditolak, penghapusan draf berhasil |
