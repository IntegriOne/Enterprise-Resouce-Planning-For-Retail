# Uji HTTPS perubahan referensi requisition

Stack terisolasi memakai Caddy dan grant yang dipasang superadmin melalui GraphQL IAM. Unit dan lokasi berasal dari EF, job berasal dari fixture HRM, dan hiring manager memakai pekerja yang akunnya tertaut IAM.

| Kasus | Hasil |
| --- | --- |
| Ubah jabatan dan unit menggunakan GID yang sesuai | Draf menyimpan kedua referensi baru |
| Kirim null untuk lokasi | Lokasi dikosongkan |
| Kirim null untuk jabatan atau GID unit sebagai lokasi | Ditolak dengan `ARGUMENT` |
| Kirim UUID unit yang tidak ditemukan | HRM menolak dengan `CONFLICT` |
| Edit sesudah pengajuan | Ditolak dengan `CONFLICT` |

Go test dan Go vet Gateway lulus.

## Uji lanjutan posisi dan lini workforce plan

Stack terisolasi menggunakan posisi dan workforce plan berstatus disetujui dari fixture PostgreSQL tes. Superadmin memperoleh grant melalui GraphQL IAM.

| Kasus | Hasil |
| --- | --- |
| Ubah draf dengan GID posisi aktif dan lini aksi recruit | Kedua referensi tersimpan dan terbaca dari detail |
| Ganti posisi dengan posisi yang masih planned | Ditolak dengan `CONFLICT` |
| Ganti lini dengan lini aksi retain | Ditolak dengan `CONFLICT` |
| Kirim null untuk kedua referensi opsional | Keduanya dikosongkan |
| Ajukan draf setelah perubahan | Pengajuan berhasil dan edit sesudahnya ditolak |
