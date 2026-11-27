# Pengujian Mutation pengajuan assignment

| Kasus | Hasil |
| --- | --- |
| Grant create, Relay ID pekerja dan manajer, FTE desimal | Pengajuan tercatat dengan waktu dari HRM dan persetujuan kosong |
| Posisi gagal memenuhi prasyarat HRM | CONFLICT |
| ID posisi tidak valid | INVALID_ARGUMENT |
| Grant approve dipakai untuk pengajuan | FORBIDDEN |

Fixture memakai sesi employee dan CSRF. Uji HTTPS dengan grant resmi masih tertunda.
