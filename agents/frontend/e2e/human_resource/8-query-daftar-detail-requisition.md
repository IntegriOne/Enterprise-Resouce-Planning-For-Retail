# Uji browser daftar dan detail requisition

Chrome headless mengakses workspace HRM melalui HTTPS dengan sesi superadmin pada stack terisolasi. Superadmin memberikan role uji lewat GraphQL IAM. Draf, unit, dan hiring manager memakai HRM, IAM, dan EF.

| Kasus | Hasil |
| --- | --- |
| Tabel daftar | Nomor requisition, kode dan judul jabatan, kode dan nama unit, serta status tampil |
| Buka detail | Nama dan nomor hiring manager tampil |
| Periksa teks tabel dan panel | Tidak ditemukan `gid://` |
| Cari nomor yang tidak ditemukan | Tampil kondisi kosong |
| API pembacaan label jabatan disimulasikan ditolak | Tampil keterangan jabatan tidak tersedia, label unit tetap tampil dan ID tidak ditampilkan |

TypeScript, ESLint untuk file terkait, dan build frontend lulus. Paginasi Relay maju dan mundur pada API diverifikasi terpisah.
