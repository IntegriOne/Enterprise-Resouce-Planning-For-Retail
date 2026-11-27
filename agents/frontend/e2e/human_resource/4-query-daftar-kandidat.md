# Daftar kandidat di ruang kerja SDM

Chrome menguji stack HTTPS terisolasi dengan frontend, Gateway, IAM, HRM, Enterprise Foundation, dan PostgreSQL. Superadmin memasang izin melalui GraphQL. Data kandidat memakai kontak Enterprise Foundation yang sah.

| Kasus | Hasil |
| --- | --- |
| Izin baca GLOBAL aktif | Navigasi Kandidat dan data bisnis tampil |
| Pencarian nomor yang tidak ada | Pesan hasil kosong tampil |
| Pencarian nomor yang cocok | Baris kandidat kembali tampil |
| Izin dicabut lalu URL dibuka langsung | Gateway menolak. UI memberi pesan izin dan menyembunyikan tautan kandidat. |
| Pemeriksaan teks area utama | Tidak memuat `gid://` |

Kedua replika frontend development memakai build terbaru dan sehat.

## Koreksi nama pilihan

Dengan izin `human_resource.candidates.update` melalui GraphQL, operator membuka **Ubah nama** pada baris kandidat aktif, menyimpan nama baru, lalu mengosongkannya. Chrome melihat konfirmasi dan tabel berubah menjadi **Belum tercatat**. PostgreSQL mengonfirmasi nama kosong serta referensi individu dan kontak yang tetap sama. Setelah role dicabut, URL langsung menampilkan penolakan akses. Tes melewati Gateway HTTPS pada stack terisolasi.

## Tambah kandidat dengan individu dan kontak

Chrome mengisi data registrasi, mencari individu aktif dengan kode bisnis EF, dan memilih kontak melalui nama. Permintaan HTTPS berhasil dan daftar menampilkan kandidat baru. PostgreSQL mengonfirmasi kecocokan referensi individu dan kontak. Area utama tidak memuat `gid://`. Superadmin mengatur izin melalui GraphQL, lalu stack dan database uji dibersihkan.
