# Penghapusan kandidat dari ruang kerja SDM

Chrome memakai stack HTTPS terisolasi. Izin dipasang melalui GraphQL oleh superadmin. Dua kandidat uji disiapkan dan salah satunya ditautkan ke lamaran.

| Skenario | Hasil |
| --- | --- |
| Konfirmasi penghapusan tanpa MFA | Formulir menampilkan nomor kandidat. Gateway menolak dan modal verifikasi terbuka. |
| Penghapusan kandidat berlamaran setelah MFA | UI menampilkan penolakan. Kandidat dan lamaran tetap tersimpan. |
| Penghapusan kandidat tanpa lamaran setelah MFA | UI menampilkan konfirmasi berhasil dan memuat ulang daftar. |
| Pemeriksaan database dan teks | Hanya kandidat uji tanpa lamaran terhapus. Area utama tidak memuat `gid://`. |

Verifikasi MFA dilakukan melalui GraphQL sebelum fase browser kedua. Runner selesai dengan status 0 dan membersihkan stack uji.
