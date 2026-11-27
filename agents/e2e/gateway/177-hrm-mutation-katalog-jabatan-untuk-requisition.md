# Uji HTTPS pembuatan katalog jabatan

Uji memakai Gateway, IAM, Enterprise Foundation dan HRM pada lingkungan disposable. Superadmin masuk melalui GraphQL. Role uji, grant permission dan assignment superadmin dipasang dan dicabut melalui GraphQL IAM.

| Kasus | Hasil |
| --- | --- |
| Membuat job family | Berhasil dan aktif secara default |
| Mengulang kode job family | Ditolak dengan `CONFLICT` |
| Membuat jabatan dengan GID job family | Berhasil dengan referensi induk dan tanggal yang benar, aktif secara default |
| Mengulang kode jabatan | Ditolak dengan `CONFLICT` |
| Memakai GID jabatan sebagai job family | Ditolak dengan `ARGUMENT` |
| Mengirim tanggal tidak valid | Ditolak dengan `ARGUMENT` |
| Membaca jabatan melalui Relay `node` | ID dan judul sesuai |
| Membuat unit Enterprise Foundation dan requisition dengan jabatan baru | Requisition tersimpan sebagai draf dengan referensi jabatan yang benar |
