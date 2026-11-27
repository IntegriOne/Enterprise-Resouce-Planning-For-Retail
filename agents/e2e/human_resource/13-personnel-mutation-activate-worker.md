# Aktivasi pekerja berdasarkan masa berlaku kontrak

Tes menggunakan gRPC Personnel, PostgreSQL sementara, Redis TLS terisolasi, dan race detector. Bukti tanda tangan dan perubahan tanggal akhir kontrak disiapkan langsung di database tes.

| Kasus | Hasil |
| --- | --- |
| Aktivasi pekerja preboarding tanpa kontrak bertanda tangan | FailedPrecondition |
| Aktivasi dengan kontrak yang baru mulai setelah tanggal masuk | FailedPrecondition |
| Aktivasi sehari setelah kontrak bertanda tangan berakhir | FailedPrecondition |
| Aktivasi tepat pada tanggal akhir kontrak bertanda tangan | Status pekerja active |
| Ulangi aktivasi | FailedPrecondition |
| Baca riwayat setelah aktivasi gagal | Riwayat kosong dan jumlah outbox tetap |
| Baca riwayat setelah sukses dan percobaan ulang | Tepat satu event hire dengan actor, tanggal efektif, waktu server, dan transisi preboarding ke active yang sesuai |

Jalur verifikasi dokumen melalui DRM masih diperlukan untuk membuktikan kontrak produksi.
