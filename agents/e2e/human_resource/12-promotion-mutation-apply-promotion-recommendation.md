# Penerapan promosi dengan reservasi posisi di masa depan

Tes memakai PostgreSQL sementara, race detector, dan policy tiruan.

| Kasus | Hasil |
| --- | --- |
| Ajukan rekomendasi setelah referensi penempatan primer dikosongkan pada fixture | FailedPrecondition, lalu pengajuan berhasil setelah referensi dipulihkan |
| Ajukan rekomendasi dengan penempatan asal yang belum disetujui | FailedPrecondition, lalu berhasil setelah bukti persetujuan dilengkapi pada fixture |
| Setujui rekomendasi setelah referensi primer dikosongkan | FailedPrecondition tanpa keputusan tersimpan. Permintaan yang sama berhasil setelah referensi dipulihkan |
| Terapkan saat manajer memiliki tanggal terminasi dan penempatan baru terbuka | FailedPrecondition |
| Terapkan setelah referensi penempatan primer dikosongkan pada fixture | FailedPrecondition |
| Terapkan saat pekerja memiliki penempatan lain yang membuat total FTE melebihi 100% | FailedPrecondition |
| Terapkan promosi saat posisi tersedia hari ini tetapi sudah penuh mulai besok | FailedPrecondition |
| Terapkan pada posisi yang sudah berakhir | FailedPrecondition |
| Terapkan dengan reservasi yang mulai tepat pada tanggal akhir posisi | FailedPrecondition |
| Terapkan dengan reservasi lama yang mulai sehari setelah masa berlaku posisi | Promosi diterapkan |
| Terapkan dengan posisi tujuan yang memiliki tanggal akhir | Penempatan baru mewarisi tanggal akhir posisi |
| Verifier Foundation menolak posisi tujuan | FailedPrecondition. Penempatan asal tetap terbuka dan primer. Penempatan tujuan, lifecycle dan outbox baru tidak tersimpan. Rekomendasi tetap approved. Promosi berhasil setelah verifier menerima |
| Periksa penempatan dan catatan hasil promosi | Penempatan lama berakhir pada H−1, penempatan primer baru tercatat, dan event lifecycle tersimpan |
