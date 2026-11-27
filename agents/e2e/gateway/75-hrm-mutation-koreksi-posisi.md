# Pengujian Mutation koreksi posisi

| Kasus fixture | Hasil |
| --- | --- |
| Ubah kapasitas dan kosongkan tanggal akhir | Field mask memuat dua field tersebut |
| HRM menolak perubahan posisi active | CONFLICT |
| Kapasitas null | INVALID_ARGUMENT |
| Patch kosong | INVALID_ARGUMENT |
| Grant create dipakai untuk koreksi | FORBIDDEN |

Regresi pembuatan posisi juga lulus. Uji HTTPS dengan grant resmi masih tertunda.
