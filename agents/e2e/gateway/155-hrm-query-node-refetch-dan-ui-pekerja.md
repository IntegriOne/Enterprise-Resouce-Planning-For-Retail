# Uji HTTPS Relay Node dan UI pekerja

Superadmin menyiapkan grant melalui GraphQL IAM pada stack terisolasi. Fixture memakai pekerja tertaut serta data Enterprise Foundation.

| Kasus | Hasil |
| --- | --- |
| Refetch requisition melalui node | Tipe, status, nomor, dan UUID cocok dengan data requisition |
| Refetch jabatan melalui node | Tipe dan judul jabatan cocok |
| Global ID jabatan dikirim pada pembuatan requisition | Draf berhasil dibuat |
| Global ID jabatan dikirim pada detail requisition | Ditolak dengan `INVALID_ARGUMENT` |
| Halaman pekerja di Chrome headless melalui HTTPS | Nama, nomor, badan hukum, tanggal, status, dan akun tertaut tampil. Teks halaman utama bebas dari `gid://` |
| Smoke dev melalui Caddy setelah rollout | Empat permintaan daftar pekerja berhasil |

Uji Gateway dengan race detector membuktikan refetch workforce plan membutuhkan izin baca resource itu. Go test dan Go vet Gateway lulus.
