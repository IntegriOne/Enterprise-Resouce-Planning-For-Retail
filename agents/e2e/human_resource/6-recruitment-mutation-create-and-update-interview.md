# Regresi jadwal wawancara melalui gRPC

Tes berjalan dengan race detector, PostgreSQL sementara, serta policy tiruan.

| Kasus | Hasil |
| --- | --- |
| Buat jadwal putaran pertama | Berhasil |
| Ulangi kombinasi lamaran, putaran, dan tipe yang sama | AlreadyExists |
| Buat jadwal dengan interval satu nanodetik yang kosong pada presisi PostgreSQL | InvalidArgument |
| Buat putaran lain dengan irisan 30 menit | FailedPrecondition |
| Buat wawancara pada lamaran berbeda milik kandidat yang sama yang beririsan dengan jadwal aktif | FailedPrecondition |
| Dua RPC serentak dari lamaran berbeda milik kandidat yang sama meminta slot yang dimulai tepat saat putaran pertama selesai | Tepat satu berhasil dan satu FailedPrecondition |
| Perpanjang putaran pertama hingga beririsan dengan jadwal yang diterima | FailedPrecondition |
| Pakai slot draf wawancara dari lamaran yang ditarik | Berhasil |
| Hapus fixture bersebelahan dan ubah jadwal putaran pertama | Berhasil |

Suite Recruitment lulus tiga kali. Setiap eksekusi memakai database sementara yang berbeda.

Kasus tambahan memverifikasi FailedPrecondition ketika waktu buka requisition masih di masa depan. Pembuatan berhasil setelah waktu buka pada fixture dipulihkan.

Wawancara fixture berstatus `in_progress` tetap memblokir nominasi slot yang bertumpang tindih untuk kandidat yang sama. RPC mengembalikan `FailedPrecondition`.
