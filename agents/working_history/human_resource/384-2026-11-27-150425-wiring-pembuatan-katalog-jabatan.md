# Wiring pembuatan katalog jabatan

Gateway membuka mutation `createJobFamily` dan `createJob` melalui RPC HRM yang sudah tersedia. Dua izin create dimasukkan ke katalog RBAC. Pengujian disposable lewat HTTPS memastikan superadmin dengan grant GraphQL IAM dapat membuat keluarga jabatan dan jabatan, membaca ulang jabatan melalui Relay, lalu membuat requisition dengan jabatan tersebut dan unit Enterprise Foundation. Kode duplikat, GID induk bertipe salah, dan tanggal invalid ditolak.
